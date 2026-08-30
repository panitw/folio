---
title: 'Story 7.6: The canvas draws every page the document will produce'
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: 'c95fa9bc6202142a69f52d571f04d86ab12c8edc'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized'] # oversized: four wide surfaces must be stated rather than summarised — the window-ORIGIN gap (Ruling E, the reason a new projection field exists at all), the honesty obligation and its assertable form (Ruling F), the sheet-stack geometry the drag inverts (Ruling I), and the three-consumer mirror the clamp lift joins (Part 3, DW-36). NOT multiple-goals: DW-36's clamp lift is not independently shippable — 7.5 declined it precisely because there are no later sheets to drag onto until this story draws them — so it is a prerequisite of AC3, not a second goal.
deferred:
  - summary: >-
      createComponentCommand hardcodes width 72 / height 24 for every palette kind, while Go's
      dropComponent path gives an image 96x48 and a line a 1pt height, so the same placement now
      produces a different box on a later sheet than on the first one.
    evidence: |-
      folio-designer/src/component-command.ts:17 emits "width":72,"height":24 unconditionally;
      folio-go/component_commands.go:1353-1395 uses dropWidth/dropHeight 72000/24000, but
      imageDropWidth/imageDropHeight 96000/48000 for images and lineDropHeight 1000 for lines.
      Story 7.6 put createComponentCommand on the user path for the first time (it was
      imported by tests only), so the divergence is newly visible rather than newly created.
      Not patched here: the fix either duplicates Go's per-kind constants into TypeScript — a
      fourth spelling, against the mirror discipline this story's own fence freezes at six
      numeral pairs — or moves the defaults into Go, which the intent does not settle.
    location: >-
      folio-designer/src/component-command.ts:17
    severity: medium
  - summary: >-
      A content component whose home window is past MAX_CANVAS_SHEETS produces no occurrence at
      all, so it is unreachable — not selectable, nameable or deletable — and the disclosure says
      only that sheets are truncated, never that components are hidden.
    evidence: |-
      folio-designer/src/sheet-stack.ts iterates `index < drawn`, and homes are computed over the
      full origins list, so a home index >= drawn is never emitted. The budget test uses a
      component-free projection, so nothing observes it. Requires a document of more than 120
      windows, which Ruling J's derivation argues the paint budget cannot fill.
    location: >-
      folio-designer/src/sheet-stack.ts
    severity: medium
  - summary: >-
      The floor flag's causes cover a bound TABLE but not a bound content TEXT element, so a
      document whose length comes from a bound scalar longer than its placeholder under-counts
      with the flag false — withholding the very sentence that describes it.
    evidence: |-
      folio-go/page_setup.go's canvasContentBandHasBoundTable tests for a table with a non-empty
      binding; the disclosure it drives generalises to "A document whose length comes from data
      prints more pages than are shown here." A bound text element takes neither the table branch
      nor the font-chain branch. Widening the cause is a projection-honesty decision the intent
      scopes to three named causes, so it is recorded rather than taken.
    location: >-
      folio-go/page_setup.go:636
    severity: medium
  - summary: >-
      The origins refusal branch in addCanvasWindowCount — a Shift that is negative, above
      MaxCanvasMillipoints, non-zero at index 0, or non-increasing — silently collapses a genuine
      multi-window document to one sheet with the floor flag set, and no test reaches it.
    evidence: |-
      Every fixture in canvas_window_count_test.go produces a well-formed Shift sequence, so the
      `!ok` path is dead code as far as the suite knows. Its collapse is indistinguishable from
      the three legitimate floor causes, so a regression that made it fire on ordinary documents
      would surface as "this document prints more pages than are shown" rather than as an error.
    location: >-
      folio-go/page_setup.go
    severity: medium
  - summary: >-
      SHEET_STACK_GAP and the .sheet-stack CSS gap are coupled by nothing executable, and the
      cross-seam drag inverse is only correct while the laid-out pixel gap equals the constant.
    evidence: |-
      sheet-stack.ts declares SHEET_STACK_GAP = 24 and sheetPitch adds it; App.css consumes it
      only through the inline --sheet-stack-gap custom property. design-contract.test.ts reads
      App.css as text but asserts nothing about .sheet-stack, and jsdom applies no stylesheet, so
      changing or dropping the CSS gap keeps every test green and drifts every drag across a seam.
      The repo already has the idiom for this tie in engine-bounds-mirror.test.ts.
    location: >-
      folio-designer/src/App.css
    severity: medium
  - summary: >-
      No shipped fixture draws an in-sheet seam, so the seam's rendering is covered only by
      synthetic literals.
    evidence: |-
      The page-count fixtures place elements a round 728000 apart in a 727890 window, so
      next - origin > contentWindowHeight on every sheet and the no-marker branch is always
      taken. The only seam coverage is the hand-authored `prose` fixture in sheet-stack.test.ts
      and the projection literals in App.test.tsx.
    location: >-
      folio-go/canvas_window_count_template.go
    severity: low
  - summary: >-
      MAX_ENGINE_CONTENT_WINDOWS (100000) has no Go counterpart and no test on either side; if a
      projection ever exceeded it the whole snapshot would be discarded and the canvas blanked
      with no attribution.
    evidence: |-
      engine-protocol.ts declares the cap and its own comment concedes the failure mode. It is
      deliberately excluded from the mirror's six numeral pairs on the stated ground that Go
      declares no maximum — which is exactly why nothing bounds the producible value.
    location: >-
      folio-designer/src/engine-protocol.ts
    severity: low
  - summary: >-
      sheetStack is rebuilt on every render, including every pointermove during a drag, at
      O(sheets x components) with up to 120 sheets.
    evidence: |-
      App.tsx computes `const stack = canvas ? sheetStack(canvas) : undefined` with no useMemo,
      and a drag calls setDrag per pointermove; stackYForColumn additionally allocates a fresh
      origins array per call and columnEdgeAfterDrag calls it twice per move. Adjacent to DW-34
      (the canvas is unvirtualised), which this story bounds with a drawing budget and leaves open.
    location: >-
      folio-designer/src/App.tsx
    severity: low
---

## In plain terms

*Non-normative. This section describes what the story must produce; the contract below governs implementation.*

Today the authoring surface is one page tall even when the document is not. This story turns it into a run of sheets — one per page-full of the main column, each repeating the strip that prints at the top of every page and the one that prints at the bottom, because the engine repeats them. The seam between two sheets falls where the engine will really break, not at an even interval: a page ends where the next thing did not fit, so the leftover space at a sheet's foot is shown. A component can be dragged onto a later sheet, and what it commits is a position in the column, not a pin to sheet three.

What the sheets promise is narrower than the title, and the interface must say so out loud, not in a tooltip nobody reads. The sheets show the pages the column **as currently laid out** occupies, not a forecast of the printed document. Wherever a document's length comes from data — a table that grows a row per record — the surface has never been given that data, so it counts the table's heading and none of its rows. Such a document shows **fewer sheets than it prints**: the four sample statements each show one sheet while printing one, five, twenty and fifty pages. A component taller than a page collapses the same way.

One test stays deliberately red throughout — the mandated corpus exercise floor, P6g — and must never be "fixed".

<intent-contract>

## Intent

**Problem:** The canvas draws one sheet however long the document is. Story 7.5 lifted the content column past page one and reported the window height and window count, but nothing draws them; the author still lays out page three by imagining where page two ended. Three things block an honest drawing: the projection reports *how many* windows but not *where each one begins*, so a browser could only guess at a fixed interval — the exact closed form `internal/layout/paginate.go` forbids by name, and measurably wrong; the live drag clamp still bounds Y in all three bands (DW-36), so the lifted column is reachable by command and not by hand; and the reported count is a **floor**, badly wrong for the documents that most need paging — measured at this baseline, `statement-1/-5/-20/-50` all report **1** window while printing **1/5/20/50** pages.

**Approach:** Project the column offset at which each window begins, and a flag saying when the count is known to be a floor; draw one sheet per window from those numbers alone, repeating the page-header and page-footer chrome and marking each seam at the projected offset; lift the drag clamp for the content band only, through the same shared list the Go validator and the browser protocol already read; translate every later-sheet gesture into a column coordinate before it becomes an opaque command; and state what the sheets do and do not claim in accessible text that a test asserts. No engine layout work: `internal/layout/paginate.go` is an input to this epic, not a target.

## Boundaries & Constraints

**Always:**
- **The sheets claim the pages the column AS THE CANVAS CURRENTLY PAINTS IT occupies — never a prediction of the rendered document.** Every disclosure the interface makes says so, and says that a document whose length comes from data will print more pages than are shown.
- Every window origin, the window height, the window count and the floor claim come from the **projection** (AD-17). The browser never derives a window position by multiplying the window height by an index, never from a paint's line count, and never from a DOM measurement.
- The drag clamp lifts for the **content band alone**, keyed on band identity through the **same list** `folio-go/component_commands.go` and `folio-designer/src/engine-protocol.ts` already share — a **third consumer of the tie, not a fourth spelling of it** (DW-36's own wording), moved in one commit with a test that reads every side.
- **A genuinely single-page template — one window, and no content-band element whose length comes from data — renders a DOM, accessible names and command payloads byte-identical to today.** Assert it; do not assume it.
- Corpus and projection neutrality: every existing template's rendered bytes and every pre-existing projection field value are unchanged, and all **twenty** `goldenDigestRecord` digests hold. Assert; do not assume.
- Any new browser-side geometry is a **pure function of projected numbers plus the canvas's own declared display constants**, unit-tested without a DOM.
- Never modify, move, delete or commit the repository-root `README.md`. Never `git add -A` / `git add .`. Stage explicit paths only.

**Block If:**
- **The acceptance criteria cannot be met honestly with what the engine can project.** If drawing the sheets would require the canvas to know something only bound data can supply, HALT with `intent gap` rather than drawing a confident wrong number. (Assessed at this gate and **NOT triggered** — see Design Notes, "Ruling E" and "Ruling F".)
- A `.folio` **format field** turns out to be required. It is not expected to be — this adds projection fields and browser drawing — but if one is, that is a version question (D-1.4.9 / D-1.4.12 / D-1.4.13) to **flag, not decide**. `SupportedMajor` is **2** and does not move here.
- The work needs a **ruling on DW-33** (a text element whose first packed line already exceeds the per-line guard paints zero lines). Flag it; do not decide whether a partial prefix should paint.
- The work would **close** DW-26, DW-27, DW-28, DW-30, DW-31, DW-32, DW-33, DW-34, DW-35 or DW-37, or would need one of them decided. Flag and stop.
- A **human sign-off record** is implicated (D-2.3.5 / D-2.4.3 / D-4.7.1). Never write one.

**Never:**
- Never modify `internal/layout/paginate.go`. It must be **absent from the diff**. Never reintroduce the closed form `ceil(lowestBottom / H)` in Go or in TypeScript, in any spelling.
- Never implement keep-together (Story 7.7) or DW-29's load rejection (Story 7.8).
- Never move `hitTestBand`'s half-open band rectangle (`folio-go/component_commands.go:1501`). It answers "which band is this point in", a different question, and is pinned by `TestDropComponentUsesGoHalfOpenBandHitTesting`.
- Never widen the `pageHeader` or `pageFooter` vertical bound, in Go, in the protocol, or in the drag clamp.
- Never take DW-35 (the canvas hard-codes one font stack) — it is **Epic 8's plan gate's**. Note only whether this story's drawing makes it more visible.
- Never add a second `@media` query to `folio-designer/src/App.css`; `canvas-authority-contract.test.ts` pins the list to exactly one.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Window origins projected | a content column spanning N windows | Projection reports N origins, band-relative millipoints, `origins[0] == 0`, strictly increasing, `len(origins) == contentWindowCount` | No error expected |
| Origins are not a fixed interval | the one-window control fixture (elements at y 0, 728pt, 1456pt; window 727890 mp) | Origins are exactly `[0, 728000, 1456000]` — **not** `[0, 727890, 1455780]` | No error expected |
| A declared gap | the ten-window-gap fixture (text at y 0; element at y 7280pt) | Origins are exactly `[0, 7280000]`; count 2 | No error expected |
| Floor flag — bound table | content band holds a table with a non-empty binding | `contentWindowCountIsFloor` is **true**; count is 1 for the shipped statement shape | No error expected |
| Floor flag — degradation | a content component taller than one window (`layout.Paginate` errors) | Count 1, origins `[0]`, `contentWindowCountIsFloor` **true**; projection **succeeds** | Never fails the projection |
| Floor flag — exact column | the gap and control fixtures, no table, no degradation | `contentWindowCountIsFloor` is **false** | No error expected |
| Sheets drawn | projection reports N windows | N sheets, in document order, each repeating the pageHeader and pageFooter bands and their components | No error expected |
| Seam inside a sheet | `origins[N+1] - origins[N] <= contentWindowHeight` | Sheet N draws the break marker at `origins[N+1] - origins[N]` down its content band; the space below it is visibly the leftover the engine will not use | No error expected |
| Seam past a sheet's foot | `origins[N+1] - origins[N] > contentWindowHeight` (a declared gap) | No in-sheet marker; the sheet's own content foot is the boundary, and the skipped column region is not drawn | No error expected |
| Component spanning windows | a content component whose extent crosses `origins[N+1]` | Drawn on every window it intersects, positioned at `component.y - origins[N]`, clipped to that window; only its **home** window's occurrence is interactive and accessibly named | No error expected |
| Existing component dragged to a later sheet | pointer drag from sheet 1 into sheet 3's content band | Component tracks the hand across the seams; the committed Y is the **column** coordinate, sent as the existing opaque move/bounds command | Go still validates and may refuse |
| New component placed on a later sheet | palette placement released over sheet 3's content band | Translated to a column coordinate and sent as the existing band-aware `createComponent` command | Go's own refusals unchanged |
| New component placed on the first sheet | palette placement released over sheet 1 | The **existing** `dropComponent` payload, byte-identical to today | Unchanged |
| Drag in a repeating band | pointer drag in `pageHeader` or `pageFooter` | Clamped at the band foot exactly as today | Unchanged |
| Single-page template | one window, no bound-length content element | DOM, accessible names and payloads identical to today: one sheet, no seam marker, no page qualifier in any label | No error expected |
| Count beyond the drawing budget | `contentWindowCount` greater than the declared sheet budget | The first budgeted sheets are drawn and the interface **says** it is showing the first N of M | Degrades; never blanks, never aborts |
| Malformed new fields | origins absent, non-array, wrong length, not increasing, or `origins[0] != 0` | The snapshot is rejected by the protocol exactly as any other malformed projection | Existing protocol failure path |

</intent-contract>

## Code Map

**Every anchor below was verified by reading the file at the baseline `2c5cfa1`.** Corrections to the dispatch's and the deferral register's line hints are marked ⚠.

### What Story 7.5 shipped, and the one thing it threw away — `folio-go/page_setup.go`

- **`CanvasProjection` `:266-332`** — lowerCamelCase json tags, **no `omitempty` on any top-level field**. `ContentWindowHeight` `:306` (comment `:294-305`), `ContentWindowCount` `:331` (comment `:307-330`). The count's comment already states the floor doctrine verbatim: *"for any document with a bound table in the content band this count is a FLOOR, never a prediction."* **That is the register the two new fields' comments must match.**
- **`canvasPageGeometry` `:371-387`** (doc `:360-370`) — builds the one `layout.PageGeometry` from `canvasDimensions`. ⚠ **Must NOT be replaced by `render.go`'s `pageGeometryOf`**, which hard-errors on `"Letter"` (Ruling A, 7.5).
- **`addCanvasWindowCount` `:506-559`** (doc `:480-505`) — builds `[]layout.ColumnItem`, translates by `layout.Origins(g).Content` at `:539-543`, calls `layout.Paginate` at `:544`, degrades to `ContentWindowCount = 1` on any error at `:546-556`, and at **`:557`** reads **only** `int64(len(plan.Pages))`.
- ⚠ **THE LOAD-BEARING FACT.** `plan.Pages[i]` carries **`PageAssignment.Shift`** — the column offset of window `i`'s top — and `:557` discards it. **The datum AC2 needs already exists inside the value this function throws away.** No new pagination, no new derivation.
- Non-text extents `:513-532`: text elements are skipped (`:514`) because `addCanvasTextPaint` emits one item per shaped line; every other content element emits one item from `projectedSize`, with the sanctioned dummy `Rects: []layout.RectRef{0}` at `:530`.
- Per-line extents, `addCanvasTextPaint` **`:785-798`** — for `band.name == bandContent`, one item per line of the **full untruncated** `lines`, `Top: canvasLineTop(originY, i, vm.Advance)`, `Bottom: top + vm.FirstBaseline + vm.LastDescent`.
- **`Canvas` `:399-452`** — no `FontSet`, cannot shape; `:437` `window := layout.ContentHeight(g)`; sets `Bands[1].Height` `:440` and returns `ContentWindowHeight: int64(window), ContentWindowCount: 1` at `:451`, the documented floor (`:392-398`).
- **`CanvasWithTextPaint` `:457-478`** — `Canvas` + `addCanvasTextPaint` + `addCanvasImagePaint` + `addCanvasWindowCount`, which **overwrites** the count. ✅ **Verified: `wasm/engine.go:119`, `:255`, `:294` are the only three seams reaching the browser and all three call `CanvasWithTextPaint`** — so `Canvas`'s values must be honest but never reach the app (Ruling B, 7.5).
- **`canvasComponents` `:919-…`** — ⚠ **`CanvasComponent.X/Y` are `element.X/element.Y`, i.e. BAND-RELATIVE.** For the content band that is already the column coordinate. (The dispatch-time assumption that they are page-absolute is wrong; `.canvas-component { left: var(--component-x) }` inside an absolutely-positioned `.page-band` confirms it.)
- `CanvasBand` `:157-163`; the three bands minted `:438-442`, paper-absolute X/Y, fixed order.

### `folio-go/internal/layout/paginate.go` — READ-ONLY. It must be absent from the diff.

- **`PageAssignment.Shift` `:500`** — `windowStart − contentTop`, i.e. the window's band-relative column offset. Window 0's Shift is unconditionally `0` (`:553`).
- **The sweep that sets it, verbatim `:995-1001`:**
  ```go
  if pageHasItem { page++; pages = append(pages, PageAssignment{}); pageHasItem = false }
  windowStart = effectiveTop
  pages[page].Shift = windowStart - contentTop
  ```
  ⚠ **`windowStart = effectiveTop` — the item's own Top. There is no `+= height` anywhere in the sweep.** The Story 4.6 over-tall branch (`:800-801`) does the same.
- **The ban, verbatim `:69-74`:** *"PAGE COUNT IS NOT A CLOSED FORM … it is NOT `ceil(lowestBottom / H)` … the window advances to the first UNPLACED item, not by a fixed H."* ⚠ **The same ban applies to origins**, and origins expose it more sharply: on the control fixture the closed form is wrong by 110 mp per window, and on the gap fixture by nine windows.
- Window **height** is invariant (`:550 height := ContentHeight(g)`), so window `i` spans `[origins[i], origins[i] + contentWindowHeight)`. Only the tops slide.
- ⚠ **A page is appended only when `pageHasItem`** (`:995`), so `len(plan.Pages) <= len(items)` — the count is bounded by the column-item count. Recorded because this story draws one sheet per window; it does **not** close DW-37, whose subject is the unguarded non-text sum and the protocol's admission of the resulting integer.
- `*OverflowError` type `:250-255`, raised `:1075-1083`. ⚠ The canvas's dummy `Rects` ref makes an over-tall canvas component report `Kind: "table"` — irrelevant while `addCanvasWindowCount` swallows the error; do not surface it.
- Zero items → exactly one page, `:548-557`.

### The browser protocol — `folio-designer/src/engine-protocol.ts` (379 lines)

- **`BANDS_CAPPING_VERTICALLY` `:78`**, exported: `['pageHeader', 'pageFooter']`. ⚠ **No `as const`, no type annotation** — `engine-bounds-mirror.test.ts:160` extracts it with `/^export const BANDS_CAPPING_VERTICALLY = \[([^\]]*)\]$/m`; annotating it silently breaks the tie.
- The band-containment gate, **`:225-226`** (not `:193`, which is `bands.length !== 3`):
  ```ts
  if (!(box.x + box.width <= band.width)) return false
  if (BANDS_CAPPING_VERTICALLY.includes(component.band as string) && !(box.y + box.height <= band.height)) return false
  ```
- **`CanvasProjection` type `:116-132`**; the two 7.5 fields at `:125` with their comment `:119-124`.
- ⚠ **`hasOnly` `:167` is a SUBSET check** (`hasExactKeys` `:168` is the strict sibling and is not used for the canvas), so a **missing** new key is caught by its value predicate, not by the allow-list. The canvas allow-list is **`:184`**; the strictly-positive integer list is **`:186`**.
- **The bounded-array precedent a new array field follows is `fontFamilies` at `:190`** — `Array.isArray`, a length cap, a per-entry predicate, and an ordering predicate, all in one guard. Copy that shape.
- ⚠ **`parseInbound` returning `undefined` discards the WHOLE snapshot**: `isCanvas` false → `isSnapshot` false → `engine-client.ts:85-87` `#fail('PROTOCOL_INVALID', …)` → `:117-124` terminates the worker → `App.tsx:702` renders `Waiting for Go page geometry.` **A Go field the TS validator rejects blanks the canvas with no attribution.** A `nil` Go slice marshals to `null` and would do exactly that — the origins slice must always be non-nil.

### The Go/TS mirror tie — `folio-designer/src/engine-bounds-mirror.test.ts`

- Numeral pairs: `goSources` `:43-50`, `tsPath` `:51`, six `pairs` `:47-54`, `toHaveLength(6)` `:81`, source set `:86`. ⚠ **These govern numerals only and need no change.**
- **The predicate tie is the `'band containment mirror'` describe, `:177-219`**, four `it`s: `:181` both sides non-empty, `:188` equality plus `not.toContain('content')`, **`:196` consumption at the validator site on both sides**, `:204` the horizontal cap stays universal, `:211` a one-sided edit turns red.
- Extractors: Go `:151-157` (resolves `bandPageHeader = "pageHeader"` identifiers, then `/^var bandsCappingVertically = \[\]string\{([^}]*)\}$/m`); TS `:159-163`.
- Site regexes `:199-201` pin two Go sites (`containComponent`'s guarded `if`, `containEdgeY`'s guard) and one TS site (`:226`'s `includes(...)`).
- ⚠ **`resize-anchor.ts` is not read by this test at all.** Adding it as the third consumer means adding a second TS source read inside `:178-179` and a fourth site regex at `:196-202`; the describe's title *"on both sides"* becomes stale and must be re-worded.

### The drag clamp (DW-36) — `folio-designer/src/resize-anchor.ts` (54 lines)

- **`DragLimit` `:23-25`** — `Readonly<{ width: number; height: number }>`, **no band name**; its own comment asserts the very thing that is no longer true (*"the band's own size is the limit"*).
- **`proposedBounds` `:29`**, `limit` optional, both axes default to `+Infinity` at `:30-31`.
- ⚠ **Exactly two clamps read `limitHeight`: `:35`** (move) and **`:52`** (south resize). `:50` is the **east/width** clamp and must not move; `:49`/`:51` clamp to `0` (the band origin floor) and must not move — Go keeps `y < 0` universal.
- Header `:9-18` declares the file a UX affordance, not an authority, and states its own governing principle: *"A pointer that leaves the band should leave the component against that edge with its other axis still **tracking the hand**"* — the principle that selects Ruling I.
- `resize-anchor.test.ts` — the only two Y-clamp assertions are **`:46`** (`y: 645_000 = 669_000 − 24_000`) and **`:53`** (`height: 649_000 = 669_000 − 20_000`); both pass a bare `{ width: 523_276, height: 669_000 }` literal declared twice, at `:42` and `:50`. Width assertions in the same `it`s (`:53`'s `width: 513_276`) must survive unchanged. ⚠ **Adding a band field to `DragLimit` is a deliberate compile break at both literals.**
- ⚠ **Only two call sites of `proposedBounds` exist repo-wide**: `App.tsx:1176` and `resize-anchor.test.ts`.

### The canvas drawing surface — `folio-designer/src/App.tsx` (1316 lines)

- **The single sheet, `:700`** — one `<section className="page-surface…" aria-label="Report page with Page Header, Content, and Page Footer" style={pageStyle(canvas, zoom)}>`; the else branch `:702` is `<p className="canvas-awaiting" role="status">Waiting for Go page geometry.</p>`.
- **The three bands, `:701`** — `canvas.bands.map(...)`, `aria-label={bandName(band.name)}`, `style={bandStyle(band, zoom)}`, `tabIndex={0}`, a `<span>` band tab, then `canvas.components.filter((c) => c.band === band.name).map(...)` with `limit={{ width: band.width, height: band.height }}`. Placement is `onPointerUp`: `placementPoint(event.nativeEvent, band, zoom)` → `place(...)`.
- `bandName` **`:1149`** → `'Page Header' | 'Content' | 'Page Footer'`.
- **`canvasDisplay` `:1155-1158`** — the app's ONLY millipoints↔px mapping: `css(mp, zoom)` and `documentDelta(px, zoom)`.
- **`placementPoint` `:1162-1164`**, `pageStyle` **`:1165`** (immediately adjacent, no blank line), `bandStyle` `:1166`, `componentStyle` `:1294`.
- `place` `:432-437` → `dropComponentCommand(kind, x, y, snapEnabled)`; keyboard fallback `place(band.x / 1000, band.y / 1000)`.
- Drag commit `:701` — `moveComponentCommand(id, finished.x, finished.y, snap)` or `setComponentBoundsCommand(...)`; `:1176` applies `proposedBounds(preview.mode, preview, dx, dy, limit)`. ⚠ **Deltas are linear `documentDelta(clientY)` — they know nothing about the chrome and gap between sheets. This is what Ruling I fixes.**
- **`canvasTruncationNotice` `:1260` and `componentAccessibleName` `:1261-1267`** — the shipped idiom for folding a sentence into a component's accessible name, asserted at `App.test.tsx:1178-1180`. **AC4's per-component half copies this exactly.**
- Status idiom, shipped: `<p role="status" aria-live="polite" aria-label="…">` — `:678`, `:702`, `:705`, `:713`. `.sr-only` exists (`App.css:7`) and is **used by nothing**; the house idiom is a visible status node.

### The canvas's stylesheet — `folio-designer/src/App.css`

- **`.canvas-region` `:53`** — `display: grid; align-content: start; justify-items: center; gap: var(--space-5); overflow: auto`. ✅ **A second `.page-surface` sibling stacks vertically with no layout change** — but the gap it inherits is a token the browser cannot read back, which is why Ruling I requires the stack to own its own gap.
- `.page-surface` `:60`, sized entirely from `--page-display-width/height`.
- **`.page-band` `:63`** — `position: absolute; left/top` from `--band-x/-y`. ⚠ **It carries NO `overflow` rule, and `.page-band::before` `:64`, `.page-band-*::after` `:68` and `.page-band > span` `:70` all position themselves OUTSIDE the band with negative offsets.** Clipping a window's content to its band therefore **must not** be `overflow: hidden` on `.page-band` — it would clip the band tab and the boundary rule. A dedicated inner element is required.
- `.canvas-tools` `:58` is `position: sticky; top: 0` inside the scrolling region — it already survives a taller stack.

### The authority contract — `folio-designer/src/canvas-authority-contract.test.ts` (89 lines)

- Corpus `:6-16` (production `src/**` non-test, tests, `e2e/**`). Prohibited patterns `:19-40`; ⚠ **`:21` bans `offset*`/`client*`/`scroll*` metrics, `:20` bans `getBoundingClientRect`, `:22` bans `offsetX|offsetY`, `:27` bans `getComputedStyle`.** No DOM measurement is available to this story at all.
- **`:38-39`** ban a height/count derived from a paint's `lines.length`; **the rationale comment `:29-37` names Story 7.6 by name.** Both regexes, their non-vacuity (`:52-55`) and their red-proofs (`:70-71`) must survive.
- ⚠ **The adjacency seam `:83-88`** — `/export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/` must match, and it runs for **every** file inside `violations()`. **No declaration, comment or blank line may be inserted between `placementPoint`'s closing brace and `function pageStyle`**, and the `offsetX/offsetY` exemption covers **only** that region.
- **`:58-61`** locks `App.css` to exactly one `@media`, `prefers-reduced-motion: reduce`.

### Commands already on the channel — `folio-designer/src/component-command.ts`

- ⚠ **`createComponentCommand(type, band, x, y, snap)` `:17` already exists**, carries the band name and **band-relative** x/y — i.e. a column coordinate — and Go handles `"createComponent"` at `component_commands.go:58` → `createComponent` `:1306` → `createComponentInBand` (`containComponent` at `:1472`, which after 7.5 caps the content band nowhere vertically). ✅ **It is imported by tests only; `App.tsx:13` does not import it.** This is the reuse point for later-sheet placement — no Go change, no `hitTestBand` change.
- `dropComponentCommand` `:20` takes a **page point** and is hit-tested. Its payload is asserted verbatim at `App.test.tsx:530-540`.
- `moveComponentCommand` `:23` / `setComponentBoundsCommand` `:32` take **band-relative millipoints**, so an existing component's committed Y is already a column coordinate.

### Tests that pin today's single sheet

- ⚠ **RTL `getByLabelText` throws on more than one match.** `App.test.tsx:28`, `:36`, `:535`, `:645`, `:737`, `:738` address the page and the bands by exact label; `:1175` asserts `document.querySelectorAll('.canvas-text-line')` has length **2**; ~32 assertions address components by `text component e1`-style names. ✅ **Every fixture has `contentWindowCount: 1`** (`App.test.tsx:17`, `DataPanel.test.tsx:18`, `engine-protocol.test.ts:4`), so all of them stay green **iff** a one-window projection draws exactly one sheet with today's labels.
- ⚠ **`e2e/application-shell.spec.ts:53-57`** uses Playwright **strict-mode** `getByLabel('Report page …')` and `getByRole('region', { name: 'Page Header'|'Content'|'Page Footer', exact: true })`. Strict mode fails on >1 match. Browser e2e is compile-only (D-000.4) so nothing would catch it — this is the one class of change with no executable proof; keep the single-window labels exact.
- `design-contract.test.ts:64` requires `.page-surface` in the shell CSS; `canvas-font-stack.test.ts:38,71` requires a line **starting with** `.canvas-text-fragment {`.
- Go: `canvas_window_count_test.go` — `:64` count/height, `:102` truncation independence, `:127` render-oracle agreement, `:161` Ruling C degradation, `:190` the bound-table floor, `:213` the empty column. Fixtures in `canvas_window_count_template.go`: gap `:43`, control `:78`, oversized `:111`, bound table `:142`.
- `component_commands_test.go` — `TestDropComponentUsesGoHalfOpenBandHitTesting` **`:517-549`** (⚠ not `:514`); `TestTheColumnLiftIsExercisedAtTheCommandSurface` `:911-1041`, whose comment `:969-974` states *"hitTestBand's rectangle is still one page tall, so this is the lowest a drop can land."*

### The measured floor — evidence for AC4, taken at this baseline

| fixture | `contentWindowCount` | rendered pages |
|---|---|---|
| `statement-1` | **1** | 1 |
| `statement-5` | **1** | **5** |
| `statement-20` | **1** | **20** |
| `statement-50` | **1** | **50** |
| `page-count-1/5/20/50` | 1 / 5 / 20 / 50 | 1 / 5 / 20 / 50 |

⚠ **All four `statement-*/input.folio` are byte-identical** (md5 `1399ba14c7f173f89abd18d9cd8a2939`), and `CanvasWithTextPaint(t *Template, fs FontSet)` **takes no data argument**. One number for four documents of 1/5/20/50 pages is forced by the signature, not by an arithmetic accident. Content band: three text elements at y 0/16/32pt, then `e8`, a table bound to `transactions[]` with `headerHeight: 28`, at y 54pt — the whole column measures 82pt in a 681.89pt window. The `page-count-*` fixtures are text elements at `y = i*728pt` with **no data at all**, which is why they agree and say nothing about the statement case (`statement_fixture_test.go:22-27` already says so).

## Tasks & Acceptance

**The ordering below is normative.** Parts 1–2 are the engine's contribution and must land before any drawing consumes them; Part 3's clamp lift is a prerequisite of Part 5's drag.

**Execution — Part 1: project the window origins and the floor claim.**

- `folio-go/page_setup.go` -- **Add `ContentWindowOrigins []int64 \`json:"contentWindowOrigins"\`` and `ContentWindowCountIsFloor bool \`json:"contentWindowCountIsFloor"\`` to `CanvasProjection` (`:266-332`)**, no `omitempty`, with comments in the register of `ContentWindowHeight`/`ContentWindowCount` (`:294-330`): which coordinate frame the origins are in (band-relative to the content column, `origins[0] == 0`), that they come from `layout.Paginate`'s own `PageAssignment.Shift` and are **never** `index * ContentWindowHeight`, and — for the flag — the three causes it reports. -- Rationale: AC2 is unsatisfiable without the origins, and AD-17 forbids the browser deriving them; the flag is what makes AC4's disclosure engine-owned rather than a browser rule about what a table means.
- `folio-go/page_setup.go` -- **In `addCanvasWindowCount` (`:506-559`), read `plan.Pages[i].Shift` into the origins alongside `len(plan.Pages)` at `:557`.** Set `ContentWindowCountIsFloor` when **(a)** any content-band element is a table with a non-empty binding, **(b)** the `layout.Paginate` degradation branch (`:546-556`) fired, or **(c)** a content-band text element contributed no extents because its font chain failed (`addCanvasTextPaint`'s existing degradation site). On the degradation branch, set origins to `[]int64{0}`. -- Rationale: the datum already exists in the value `:557` discards; **no second pagination and no new derivation.** ⚠ The slice must **always be non-nil** — a `nil` marshals to `null`, which the protocol rejects, which blanks the canvas with no attribution.
- `folio-go/page_setup.go` -- **In `Canvas` (`:399-452`), set `ContentWindowOrigins: []int64{0}` and `ContentWindowCountIsFloor: true`, and say in the comment why**: `Canvas` has no `FontSet`, cannot shape, and therefore reports the documented one-window floor. -- Rationale: Ruling B (7.5) — `Canvas` never reaches the browser, but the struct is shared and its values must be honest.
- `folio-go/canvas_window_count_test.go` -- **Assert the origins for every 7.5 fixture, by exact value**: gap ⇒ `[0, 7280000]`; control ⇒ **`[0, 728000, 1456000]`, explicitly not `[0, 727890, 1455780]`**; oversized ⇒ `[0]` with the flag **true**; bound table ⇒ `[0]`, count 1, flag **true**; gap and control ⇒ flag **false**; empty column ⇒ `[0]`, flag **false**. Assert `len(origins) == count`, `origins[0] == 0` and strict increase for each. -- Rationale: ⚠ **the control fixture red-proves the closed form for origins by 110 mp per window** — a sharper discriminator than the count case, which the control only distinguishes by construction.

**Execution — Part 2: admit the two fields in the browser protocol.**

- `folio-designer/src/engine-protocol.ts` -- **Add both keys to the `CanvasProjection` type (`:116-132`) and to the `hasOnly` allow-list (`:184`); validate the origins in the shape of `fontFamilies` (`:190`)** — `Array.isArray`, a declared length cap, every entry a safe non-negative integer, `origins[0] === 0`, strictly increasing, and `origins.length === contentWindowCount`; validate the flag with a `typeof … === 'boolean'`. -- Rationale: ⚠ `hasOnly` is a **subset** check, so a missing key is caught by the value predicate, not the allow-list — the predicate is what actually protects this.
- `folio-designer/src/engine-protocol.test.ts`, `folio-designer/src/App.test.tsx:17`, `folio-designer/src/DataPanel.test.tsx:18` -- **Add both fields to the three base fixture literals** (everything else is a spread), and add rejection cases for each malformed origins shape in the matrix. -- Rationale: verified — these are the only three non-spread projection literals in the designer.

**Execution — Part 3: lift the drag clamp for the content band only (DW-36).**

- `folio-designer/src/resize-anchor.ts` -- **Give `DragLimit` the band's identity and gate the two `limitHeight` clamps (`:35`, `:52`) on `BANDS_CAPPING_VERTICALLY`, imported from `engine-protocol.ts`.** Leave `:49`, `:50` and `:51` untouched, and update the `DragLimit` comment, which currently asserts the thing that is no longer true. -- Rationale: DW-36 — *"a third consumer of the tie, not a fourth spelling of it."* ⚠ **Gate INSIDE `proposedBounds`, not at the App.tsx call site**, on the precedent `component_commands.go:1815-1817` states for the identical shape: every caller can receive any band.
- `folio-designer/src/App.tsx` -- **Pass the band name in the `limit` prop at `:701`.** -- Rationale: `band.name` is already in scope on that line for the key, class, label and component filter.
- `folio-designer/src/resize-anchor.test.ts` -- **Update the two band literals (`:42`, `:50`) to carry a name, keep `:46` and `:53` as `pageHeader` assertions with their existing numbers, and add content-band cases** asserting a move and a south resize past one window are **not** clamped, while every width assertion is unchanged. -- Rationale: `:46`/`:53` are the only Y-clamp assertions in the repo; without a content twin the lift has no red proof.
- `folio-designer/src/engine-bounds-mirror.test.ts` -- **Add `resize-anchor.ts` as a third source in the `'band containment mirror'` describe (`:177-219`), with a fourth site regex pinning its guarded clamp, and re-word the describe's "on both sides" titles.** -- Rationale: R4's widened obligation — *any* invariant duplicated across the boundary moves in one commit with a test that reads every side. ⚠ Do **not** touch the six numeral `pairs` or `toHaveLength(6)` at `:81`.

**Execution — Part 4: the sheet stack, as a pure function.**

- `folio-designer/src/sheet-stack.ts` (new, **`.ts` not `.tsx`**) -- **Build the sheet model from the projection alone**: for each window `i` under the drawing budget, its origin `origins[i]`, the column range it shows, the in-sheet seam position `origins[i+1] - origins[i]` **when that is `<= contentWindowHeight`** and none otherwise, and which components intersect the window (with their in-sheet Y as `component.y - origins[i]` and which window is each component's **home**). Also export the display-space inverse used by the drag: display-y ⇒ `(window index, column Y)`, from `canvas.height`, `zoom`, the window origins and the stack's own declared gap constant. Declare `MAX_CANVAS_SHEETS` here with its **derivation in its comment**. -- Rationale: ⚠ **oxlint's baseline is exactly 4 `only-export-components` warnings; a non-component export added to a `.tsx` makes a fifth, which is a regression.** A pure module is also the only way to unit-test the geometry, since every DOM measurement identifier is banned.
- `folio-designer/src/sheet-stack.test.ts` (new) -- **Cover every matrix row that is geometry**: the control origins, the declared-gap case where no in-sheet seam is drawn, a component spanning two windows appearing on both with one home, the display-space inverse round-tripping a column Y through a sheet index at zoom 1 and at a non-unit zoom, and the budget case. **Red-proof the seam and the inverse by substituting `index * contentWindowHeight` for the projected origin and showing each test fails.** -- Rationale: matrix-audit discipline — a covering test that never calls the function, or that re-derives the rule in its own body, proves nothing.

**Execution — Part 5: draw the sheets, repeat the chrome, and translate the gestures.**

- `folio-designer/src/App.tsx` -- **Render one `.page-surface` per modelled sheet inside a stack element**, each repeating all three bands and, for the two repeating bands, all of their components — because the engine repeats them. Draw each content component on every window it intersects at its in-sheet Y; **mark the non-home occurrences `aria-hidden` and non-interactive**, so one component never presents two identical accessible names. Draw the seam marker where the model puts one. **When the projection reports one window, emit exactly today's DOM**: one sheet, today's `aria-label`s verbatim, no page qualifier, no seam, no stack disclosure. -- Rationale: AC1 and AC5. ⚠ RTL and Playwright both fail on duplicate exact labels; sheets 2+ and their bands need qualified names, and sheet 1 must not gain one when it is alone.
- `folio-designer/src/App.tsx` -- **Translate both later-sheet gestures to column coordinates.** Placement: the band element the author released over identifies its window, so extend `placementPoint` to yield the column point and send **`createComponentCommand(kind, band.name, x, columnY, snap)`** for windows after the first, leaving the **first window on today's `dropComponent` payload byte-identical**. Drag: feed `proposedBounds` a column Y derived through the stack model's display-space inverse rather than a linear pixel delta. ⚠ **Any new pointer-offset arithmetic must live inside `placementPoint` itself** — the authority contract exempts only the text between its signature and `function pageStyle`, and **nothing may be inserted between `:1164` and `:1165`.** -- Rationale: AC3; and `hitTestBand`'s rectangle is one page tall and must not move, which is exactly what the band-aware create command routes around.
- `folio-designer/src/App.css` -- **Style the stack and clip each window's content to its own band using a dedicated inner element.** ⚠ **Do not put `overflow` on `.page-band`** — `::before` (`:64`), `::after` (`:68`) and `> span` (`:70`) deliberately position outside it. Give the stack its gap from the constant `sheet-stack.ts` declares, applied as a custom property, so the browser never reads a token back. **Add no `@media` query.** -- Rationale: `canvas-authority-contract.test.ts:58-61` locks the media list; `getComputedStyle` is banned.

**Execution — Part 6: the honesty obligation, made assertable (AC4).**

- `folio-designer/src/App.tsx` -- **Add a canvas status node in the shipped idiom** (`<p role="status" aria-live="polite" aria-label="…">`, as at `:678`/`:702`/`:705`) that states, whenever the projection reports more than one window **or** reports the count as a floor: how many sheets are shown; that a component's page is **a consequence of the content above it and can change when the data does — a column position, not a pin to page three**; and, when the floor flag is set, that **a document whose length comes from data prints more pages than are shown here.** When the drawing budget truncates the stack, it also says it is showing the first N of M. -- Rationale: AC4 is the one criterion with no test that can fail unless it is given a surface; the epic's own wording is quoted here so the assertion has a literal to hold.
- `folio-designer/src/App.tsx` -- **Fold the same claim into `componentAccessibleName` (`:1261-1267`) for a component whose home window is not the first**, exactly as `canvasTruncationNotice` (`:1260`) is folded in today. -- Rationale: the shipped idiom, already asserted at `App.test.tsx:1178-1180`; a tooltip is not an obligation.
- `folio-designer/src/App.test.tsx` -- **Assert both halves by role and accessible name, in both directions**: present for a multi-window projection and for a floor projection; **absent** for a one-window non-floor projection; and the per-component sentence present on a later-sheet component and absent on a first-sheet one. **Red-proof by deleting the sentence and showing the test fails.** -- Rationale: mutation-proof, not presence-proof.
- `folio-designer/src/canvas-authority-contract.test.ts` -- **Add a prohibited pattern for a window position multiplied out of the window height** (the `contentWindowHeight * index` family, both operand orders), with a **red-proof** beside the existing pair at `:70-71` and the non-vacuity floors at `:52-55` intact. -- Rationale: AC2 says the boundary is taken from the projection **rather than computed in the browser**; this is the same guard the file already applies to a paint's line count, and its rationale comment `:29-37` names this story.

**Execution — Part 7: the single-page proof, and the record.**

- `folio-designer/src/App.test.tsx` -- **Assert the single-window DOM is identical to today**: one `Report page with Page Header, Content, and Page Footer`, three bands with today's exact labels, no seam marker, no stack disclosure, and the `dropComponent` payload at `:530-540` **unchanged byte-for-byte**. -- Rationale: AC5, and ⚠ the compile-only e2e specs address these labels in Playwright strict mode where nothing executable would catch a break.
- This spec's Delivery Log -- **Record: the measured statement/page-count table re-taken at the closing revision; that DW-36 is CLOSED and DW-37, DW-33, DW-34 and DW-35 are OPEN and untouched; whether the multi-sheet drawing makes DW-35 more visible; and the `MAX_CANVAS_SHEETS` derivation.** -- Rationale: DW-36's owner is this plan gate; the others are fenced, and an audit closes only what it measured.

**Acceptance Criteria:**

- Given a content column spanning N windows, when the canvas renders, then **N sheets** are drawn in order, each repeating the pageHeader and pageFooter bands and their components, and the sheet count comes from `contentWindowCount` — never from a paint's line count and never from a DOM measurement.
- Given the boundary between two windows, when it is drawn, then its position is `contentWindowOrigins[N+1] − contentWindowOrigins[N]` taken from the projection; and given the designer sources, when the authority contract scans them, then no window position is derived by multiplying the window height by an index, in either operand order.
- Given the one-window control fixture, when the projection is measured, then the origins are exactly `[0, 728000, 1456000]`; and given the ten-window-gap fixture, then they are exactly `[0, 7280000]` — in both cases the closed form is a measurably different answer.
- Given a component dragged from one sheet onto a later sheet, when the drop is committed, then the component **tracked the pointer across the seam**, the committed Y is a **column** coordinate, and it is sent as one opaque command; and given a placement released over a later sheet, then it is sent as the band-aware `createComponent` command carrying a column coordinate.
- Given a drag in `pageHeader` or `pageFooter`, when it reaches the band foot, then it is clamped exactly as today; and given the Go validator, the browser protocol and the drag clamp, when the commit is inspected, then all three read the **same** shared band list and a test reads every side.
- Given a projection reporting more than one window, or reporting the count as a floor, when the canvas renders, then the interface states in **accessible text** — asserted by role and accessible name, and red-proved by deleting the sentence — that a component's page is a consequence of the content above it and can change when the data does, and that a document whose length comes from data prints more pages than are shown.
- Given a component whose home window is not the first, when its accessible name is read, then it carries that same claim.
- Given a template whose content column occupies one window and whose content band holds no bound-length element, when the canvas renders, then the DOM, the accessible names and the `dropComponent` payload are **identical to `2c5cfa1`**.
- Given a projection whose window count exceeds the declared drawing budget, when the canvas renders, then the first budgeted sheets are drawn, the interface says it is showing the first N of M, and the canvas is never blanked; and given a projection whose origins are absent, mis-ordered, wrong-length or not starting at zero, then the snapshot is rejected by the existing protocol failure path.
- Given every existing template, when it is projected and rendered, then every pre-existing projection field value is unchanged and all **twenty** golden digests are **measured** unchanged.
- Given the diff, when it is inspected, then `internal/layout/paginate.go` is **absent** from it, `hitTestBand`'s band rectangle is unchanged, and the repository-root `README.md` is byte-identical to its committed state and appears in no commit.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 1, medium 3, low 1)
- defer: 8: (high 0, medium 5, low 3)
- reject: 13
- addressed_findings:
  - `[high]` `[patch]` The stack's display-space inverse was not an inverse for a column offset in
    a region a declared gap skips, so a drag that moved the pointer ZERO pixels committed a move of
    more than nine windows. Proven by execution before triage: with origins `[0, 7280000]`, a
    component at column offset `3000000` came back as `9414110`. `stackYForColumn` placed such a
    point past its own sheet's foot and `columnForStackY` then floored it onto a later sheet and
    added that sheet's origin. Reachable through this story's own third floor cause — a text
    element whose font chain will not resolve contributes no column items, so no window ever opens
    at its top. Fixed by giving the drawing and the drag ONE clamp (`offsetWithinWindow`) so a
    point with no drawn position of its own is shown, and dragged, against the foot of the sheet
    that owns it; the shipped `prose` round-trip never saw it because every offset it tries is
    inside its own window. Red-proved: removing the clamp restores `9414110`.
  - `[medium]` `[patch]` Every echo of a component that crosses a seam was pointer-interactive and
    advertised the `move` cursor, defeating Ruling G's "only the home occurrence is interactive":
    `.canvas-component-echo { pointer-events: none }` is specificity (0,1,0) and lost to
    `.band-window > .canvas-component { pointer-events: auto }` at (0,2,0). An echo therefore
    swallowed the placement pointerup the band beneath it was meant to receive. Fixed by matching
    the ancestor in the echo's own selector. jsdom applies no stylesheet, so no unit test could
    have observed this.
  - `[medium]` `[patch]` The later-sheet POINTER placement branch had no test — only the keyboard
    branch did, and the two are separate expressions on the same handler. Demonstrated: replacing
    it with `placeInBand(band.name, point.x, point.y)` left all 94 App tests green while a
    mouse-dropped component silently landed on sheet one with a page-absolute x. Added an assertion
    on the exact command bytes; red-proved against that same mutation.
  - `[medium]` `[patch]` The Go floor flag's `band.name == bandContent` guard had no test:
    deleting it left the entire Go suite green, so any template with an unshapeable HEADER title
    would have told the author its document prints more pages than are drawn — a false claim on a
    count that is exact. Added `canvasWindowCountUnshapedHeaderTemplateJSON` as a negative case
    beside the existing exact fixtures; red-proved against the guard's deletion.
  - `[low]` `[patch]` The drag's clip lift (`band-window-open`) was observed by nothing: pinning
    the class to the constant `band-window` left all 94 App tests green while the dragged component
    would be clipped out of view at the very seam it was crossing. Added in-flight and
    after-settle assertions; red-proved against that mutation.


## Design Notes

### Ruling E — the projection must carry the window ORIGINS; the fixed interval is both banned and measurably wrong

AC2 requires the boundary to be *"marked where the engine will actually break, taken from the projection rather than computed in the browser."* Story 7.5 projected the window **height** and the window **count** and nothing about where a window begins. The only browser-side derivation available is `index × contentWindowHeight` — which is `ceil(lowestBottom / H)`'s twin, forbidden by name at `paginate.go:69-74`, and wrong: the sweep sets `windowStart = effectiveTop` (`:1000`), the item's own Top, and there is no `+= height` anywhere in it.

**Selected, not chosen.** Three settled things point one way and none points the other: AC2's own wording; AD-17 (*"Anything the canvas draws about page windows and break positions must come from the engine's projection, not from a browser computation"*); and the ban itself. And the datum needs no derivation at all — `PageAssignment.Shift` (`paginate.go:500`) **is** the window's band-relative column offset, and `addCanvasWindowCount:557` currently throws it away while keeping `len(plan.Pages)` from the same value.

**Measured, so the story can red-prove it.** On the control fixture the closed form gives `[0, 727890, 1455780]` where the engine gives `[0, 728000, 1456000]` — wrong by 110 mp per window, small enough to survive a casual eye and large enough for an exact assertion. On the gap fixture it gives eleven windows where the engine gives two.

### Ruling F — what the sheets CLAIM, and why that is not an intent gap

**The sheets claim the pages the column, as the canvas currently paints it, occupies.** They do not claim the pages the document will print. That reading is not a builder's softening of the title: it is Story 7.5's **R2**, settled and recorded in `ContentWindowCount`'s own doc comment; it is the epic's own AC4 (*"its page is a consequence of the content above it and can change when the data does — it is a column position, not a pin to page three"*); and it is the only reading the measured evidence permits — the four statement fixtures are **byte-identical templates** and `CanvasWithTextPaint` takes **no data argument**, so one number for documents of 1/5/20/50 pages is forced by the signature.

**Why this is not `intent gap`.** The plan-gate test is whether a governing principle stated in the intent selects between readings. Here the intent supplies the disclosure itself: AC4 already obliges the interface to say the page is a consequence and not a pin, and the dispatch settles that the same honesty must cover the bound-table floor. The ACs are therefore satisfiable *as written*, provided the claim is stated — which is precisely what makes AC4 the load-bearing criterion rather than the decorative one.

**What makes it assertable, since AC4 has no test that can fail on its own.** Two surfaces, both already shipped idioms in this file, both asserted in both directions and red-proved: the canvas status node (`App.tsx:678`/`:702`/`:705`) and `componentAccessibleName`'s folded sentence (`:1260-1267`, asserted at `App.test.tsx:1178-1180`). A tooltip would satisfy neither.

**Why the floor claim is a projected flag and not a browser rule.** AD-17 puts anything the canvas says about page windows on the engine's side of the line, and only the engine knows all three causes: a bound table (`projectedSize` gives header height and no rows), the Ruling C degradation, and a text element whose font chain failed. A browser rule spelled `component.type === 'table'` would be a fourth authority on a question the engine can answer exactly.

### Ruling G — a component spanning windows is drawn once per window, and only its home occurrence is interactive

Body copy running for pages is the epic's whole premise, so a content element will routinely cross a window boundary. Drawing it only on its home sheet would leave later sheets empty and make the drawing a lie in a second way; drawing it identically on both would give one component two identical accessible names, breaking `getByLabelText` in ~32 assertions and Playwright's strict mode, and making selection ambiguous — against the epic's own *"Band identity and drop targeting stay unambiguous as the canvas grows past one page."* So: draw every intersecting window, clip each to its window, and let exactly one occurrence — the home window's — carry the role, the tab stop and the name.

### Ruling H — later-sheet placement uses the band-aware command that already exists; `hitTestBand` does not move

`hitTestBand`'s rectangle is one page tall, and the three bands **exactly tile** the printable column, so a page point one millipoint past the content foot resolves **silently to `pageFooter`** — the component is created in the wrong band, not refused. Story 7.5 recorded this asymmetry as *"Story 7.6's to resolve when it draws the later sheets"*, and the resolution is to route around it, not to move it: `hitTestBand` answers *which band is this point in*, a different question, pinned by `TestDropComponentUsesGoHalfOpenBandHitTesting`.

`createComponentCommand(type, band, x, y, snap)` already exists on the channel, already carries the band name and band-relative coordinates, and Go already handles it. The **first** window keeps today's `dropComponent` payload byte-for-byte, because AC5 says a single-page template behaves exactly as today and `App.test.tsx:530-540` asserts that payload; the second path exists only where no behaviour existed before. Two spellings, each correct for its case, rather than one spelling that moves a shipped payload.

### Ruling I — the drag tracks the hand across a seam, through a pure inverse of the stack

`resize-anchor.ts`'s own header states the principle: *"A pointer that leaves the band should leave the component against that edge with its other axis still tracking the hand."* Today the drag accumulates a linear `documentDelta(clientY)`, which knows nothing about the repeated chrome and the gap between sheets, so across a seam the component would drift from the pointer by the full height of a footer, a gap and a header. Tracking the hand therefore requires the display-space inverse of the stack.

Every DOM measurement identifier is banned — `getBoundingClientRect`, `offset*`, `client*`, `scroll*`, `getComputedStyle` — so the inverse must be arithmetic over numbers the canvas already holds: `canvas.height`, `zoom`, the projected origins and window height, and **the stack's own gap**, which is why the gap becomes a constant the module declares and writes out as a custom property rather than a token the browser would have to read back.

### Ruling J — the sheets drawn are budgeted, the budget degrades honestly, and DW-37 stays open

`len(plan.Pages) <= len(items)` (a page is appended only when `pageHasItem`, `paginate.go:995`), so the count is bounded by the column-item count — but that bound is the shaped-line count, not a small number, and this story turns each window into a sheet's worth of DOM, which DW-34 already records as unvirtualised. D-7.4.2 settles the shape without a new decision: **truncate the drawing, never the value; keep the projection; make the degraded state distinguishable from the empty one; derive the number and record the criterion.**

**The derivation.** Epic 7's narrative target is forty pages. The projection's own body-text budget is 1920 lines (`maxCanvasBodyTextLines`), which at the decision log's corrected forty-to-fifty lines per A4 window spans under fifty windows of solid prose — so beyond that a window can only come from a *declared* placement gap. `MAX_CANVAS_SHEETS = 120` is three times the epic's stated target and more than twice what the paint budget can fill, and its comment must say exactly that rather than present a round number.

**DW-37 is NOT closed by this.** Its subject is the unguarded non-text sum in `addCanvasWindowCount` and the protocol's admission of the resulting integer; both are untouched here. A browser-side drawing budget is an additive guard on the drawing, in the same relation to the count that a truncated paint has to a value.

### Limits to state, not to fix

- **D-7.4.4 stands, and this is the plan gate that was told to raise it.** The canvas calls `atomicSpansFor(..., nil)` — nil substitutions — so it breaks the **raw template string**, `{{...}}` and all, not the bound value. **This story therefore claims nothing about where the engine will break a bound value**, and its disclosures must not imply otherwise. Reuse of the render/bind machinery was rejected by 7.5's R1 on three independent grounds and is Story 13.4, in another epic. Nothing here needs it, because the origins come from the same extents the canvas already paints.
- **DW-33 is provably untouched.** The origins and the count read `plan.Pages`, `lines`, `originY` and the vertical model — never `painted`, `budget`, `oversized` or `placed` — so their answer is identical whether an element paints zero lines or all of them. No ruling on the partial prefix is made or needed.
- **DW-35 becomes more visible and is still Epic 8's.** Drawing N sheets multiplies the text the hard-coded font stack paints; a chain mismatch that shows as one page of colliding glyphs today would show as N. Note it in the Delivery Log; do not take it.
- **DW-34 is multiplied by the sheet count** and is bounded, not closed, by Ruling J's budget.
- **No `.folio` format field is required.** This adds projection fields and browser drawing; the projection is not the file format. `SupportedMajor` stays **2**.

## Verification

This story changes a channel schema on both sides and the canvas's whole drawing surface, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). **Report measured pass/fail counts, never "green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:169`, P6g subtest, floor at `:184`, `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` (`:243`) must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean, exit 0.
- `gofmt -l folio-go` -- run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- run **once per leg** with `FOLIO_MATRIX_TARGET` set: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` (`matrix_test.go:69-74`). ⚠ **Unset, this test logs "asserts NOTHING" and returns — a no-op is not a pass.** Grep each leg's output for `asserts NOTHING` and report the count per leg; name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; all four targets from one process, not gated on the env var.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass. **Baseline is 248 tests / 32 files**, plus whatever this story adds. **oxlint baseline is exactly 4 `only-export-components` warnings and 0 errors** (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1155,1162`); ⚠ **a fifth is a regression** — keep new exports in a `.ts` file. This story is mostly browser-side, so **the designer suite is the load-bearing one.**
- `cd folio-designer && npm run test:e2e:compile` -- expected: pass. ⚠ It is `tsc --noEmit` only. **Browser e2e is deferred by D-000.4 and does NOT execute — say so rather than implying it ran.**

**Nine digests to report byte-identical** (all **twenty** `goldenDigestRecord` entries must hold): statement-1 76,744 `114df1d6…`; statement-5 127,363 `70dce051…`; statement-20 269,884 `56bfbbd9…`; statement-50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`. **Assert; do not assume.** Confirm `git status fixtures/` is empty before quoting any digest.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` (`fontgen_matrix_test.go:64`) fails under `-tags=matrix` without `fontTools`; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, Story 15.2).

**Manual checks:**
- **Re-measure the statement/page-count table at the closing revision** and record it in the Delivery Log — do not copy the numbers above forward unmeasured.
- **Confirm `internal/layout/paginate.go` is absent from `git diff --name-only`**, that `hitTestBand`'s band rectangle is unchanged, and that `README.md` appears in no commit and its md5 is still `078d7d80d518d54af2fc04fb270d46b8`.
- **Confirm the Go split, the protocol and the drag clamp landed in the SAME commit** (`git show --stat`), and that the mirror test reads all three sources.
- **Demonstrate end to end** that a component authored windows down the column is drawn on its own sheet, can be dragged to a further sheet, and survives the round trip through the canonical bytes — not that a conditional changed.
- **Red-proof, and record what failed**: the origin seam against `index * contentWindowHeight`; the AC4 sentences against deletion; the content-band clamp lift against restoring the clamp.

## Auto Run Result

### Dispatch 1 — 2026-08-31, plan only

Status: `ready-for-dev`
Blocking condition: none
Baseline: `2c5cfa1d074e80f24488b1ea7fa1621588c05cdb` on `main`, tree clean
Directive: `Halt after planning.` — **no implementation code, no commits, and none were made.**

**The honesty question the dispatch reserved for this gate was assessed and does NOT halt.** The ACs are
satisfiable as written, because AC4 is itself the disclosure the floor requires: the sheets claim the
pages the column **as the canvas currently paints it** occupies, never a prediction of the render, and
the interface states that in accessible text a test asserts. Ruling F records the selection.

**One thing the projection could not supply, and the story now adds.** AC2 requires the window boundary
to be *"taken from the projection rather than computed in the browser"*, and Story 7.5 projected the
window **height** and **count** but nothing about where a window **begins**. The only browser-side
derivation available is `index × contentWindowHeight` — the closed form `paginate.go:69-74` forbids by
name, and measurably wrong. `PageAssignment.Shift` (`paginate.go:500`) is exactly the datum, and
`addCanvasWindowCount:557` discards it while keeping `len(plan.Pages)` from the same value. Two
projection fields therefore land here: `contentWindowOrigins` and `contentWindowCountIsFloor`. Not an
intent gap — AC2's own wording, AD-17 and the ban all select the same answer (Ruling E).

**Six forks surfaced and all were RULED, none is a new intent gap.** Each was selected by something
already settled: **E** (project the origins — AC2 + AD-17 + the ban); **F** (what the sheets claim —
epic AC4 + 7.5's R2 + measured evidence); **G** (a spanning component is drawn per window, one
interactive home — the epic's own unambiguous-targeting rule, and duplicate accessible names would
break ~32 shipped assertions and Playwright strict mode); **H** (later-sheet placement uses the
band-aware `createComponent` that already exists, and `hitTestBand` does not move — AC5 protects
today's `dropComponent` payload, and 7.5 recorded the asymmetry as 7.6's to route around);
**I** (the drag tracks the hand across a seam through a pure inverse — `resize-anchor.ts`'s own header
states the principle, and every DOM-measurement identifier is banned); **J** (the sheets drawn are
budgeted and degrade honestly — D-7.4.2's settled shape).

**Measured at this baseline, not assumed** — the floor is real and exactly as the dispatch stated:
`statement-1/-5/-20/-50` report `contentWindowCount` **1/1/1/1** while rendering **1/5/20/50** pages;
`page-count-1/5/20/50` report **1/5/20/50** and render **1/5/20/50**. All four statement templates are
byte-identical (md5 `1399ba14c7f173f89abd18d9cd8a2939`) and `CanvasWithTextPaint` takes no data
argument, so one number for four documents is forced by the signature. Measured with a throwaway
program outside the repo; the tree was left clean.

**Anchor corrections found by re-measurement** (the dispatch's and the register's hints were off):
`hitTestBand` is `:1494-1514` with the half-open rectangle at `:1501`, and its test is
`component_commands_test.go:517-549`, not `:514`; `engine-protocol.ts`'s band-containment gate is
`:225-226`, not `:193`; `resize-anchor.ts`'s two vertical clamps are `:35` and `:52` — `:50` is the
**east/width** clamp and must not move; `App.tsx`'s `canvasTruncationNotice` is `:1260` and
`componentAccessibleName` `:1261-1267`. ⚠ **`CanvasComponent.X/Y` are BAND-relative**, so for the
content band they are already the column coordinate.

**DW-36 is this story's and is scoped to close.** DW-33 is flagged and provably untouched (the origins
and the count read `plan.Pages`, `lines`, `originY` and the vertical model, never `painted`, `budget`
or `oversized`). DW-26/27/28/30/31/32/33/34/35/37 stay **open**: DW-34 is multiplied by the sheet count
and is bounded, not closed, by Ruling J's budget; DW-37's subject — the unguarded non-text sum and the
protocol's admission of the integer — is untouched; DW-35 becomes **more visible** (N sheets multiply
the text the hard-coded stack paints) and remains Epic 8's plan gate's.

**No format field is required.** `SupportedMajor` stays **2**.

**D-7.4.4 raised, as reserved.** The canvas breaks the raw template string with nil substitutions, so
this story claims nothing about where the engine will break a **bound** value, and its disclosures must
not imply otherwise. Nothing here needs the render/bind machinery: the origins come from the same
extents the canvas already paints.

Verification was **not** run in this dispatch: no code changed. The `## Verification` section states
what the implementing dispatch must measure.

### Dispatch 2 — 2026-08-31, implemented and reviewed

Status: `done`
Blocking condition: none
Baseline: `c95fa9bc6202142a69f52d571f04d86ab12c8edc` on `main`, tree clean.
Commits: `c834158` (implementation) and the review-patch commit that follows it.

**What shipped.** The engine projects `contentWindowOrigins` — read from `layout.Paginate`'s own
`PageAssignment.Shift`, the datum `addCanvasWindowCount` was discarding while keeping
`len(plan.Pages)` from the same value — and `contentWindowCountIsFloor`, OR-ing three causes only
the engine can see: a content-band table with a non-empty binding, the Ruling C degradation, and a
content text element whose font chain will not resolve. The browser protocol admits both keys in the
`fontFamilies` shape. A new pure `sheet-stack.ts` turns those numbers into sheets, seams, spanning
occurrences with one interactive home, and the display-space inverse the drag uses; `App.tsx` draws
one `.page-surface` per sheet, repeats the two bands the engine repeats, routes later-sheet
placement onto the existing band-aware `createComponent`, and states the claim in the shipped
`role="status"` idiom plus the folded per-component sentence. DW-36's clamp lifts for the content
band alone, keyed on `BANDS_CAPPING_VERTICALLY` imported from the protocol — a third consumer of the
tie, with `engine-bounds-mirror.test.ts` reading all three sources.

**Files changed.** `folio-go/page_setup.go` (two projection fields, the origins read, the three floor
causes); `folio-go/canvas_window_count_test.go` and `canvas_window_count_template.go` (origins by
exact value, the floor's causes and its absence, plus the unshaped-header negative added at review);
`folio-designer/src/engine-protocol.ts` and `.test.ts` (admission and rejection of both fields);
`resize-anchor.ts` and `.test.ts` (the band-keyed clamp); `engine-bounds-mirror.test.ts` (the third
source); `sheet-stack.ts` and `.test.ts` (new, the whole geometry); `App.tsx`, `App.css`,
`App.test.tsx` (the drawing, the gestures, the disclosure); `canvas-authority-contract.test.ts` (the
closed-form ban for origins); `DataPanel.test.tsx` (fixture field).

**Review findings.** 5 patched (1 high, 3 medium, 1 low), 8 deferred, 13 rejected; 0 intent gaps and
0 bad-spec loopbacks. The high one was a real defect proven by execution before triage: the stack's
display-space inverse was not an inverse for a column offset in a region a declared gap skips, so a
drag that moved the pointer zero pixels committed a move of more than nine windows. See the Review
Triage Log for all five and their red proofs.

**Follow-up review recommended: true** — patched counts high 1, medium 3, low 1; a high-severity
patch sets it, and the score `3 x 3 + 1 x 1 = 10` is at or above 5 independently.

**Verification, measured after the patches.**
- `go test -count=1 ./...` — exactly ONE red, `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)`,
  `got 7, need >=20`, the mandated permanent red, untouched. Its drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` is green. Every other package `ok`.
- `go vet -tags=matrix ./...` exit 0. `gofmt -l folio-go` from the repo root: no output.
- `TestTargetRenderHash` ran on all four legs with `FOLIO_MATRIX_TARGET` set — darwin/arm64,
  linux/amd64, linux/arm64, js/wasm — each PASS with an `asserts NOTHING` count of **0**.
  `TestCrossTargetByteIdentity` ok.
- `cd lint && go test ./...` — all four packages ok.
- Designer: typecheck clean; oxlint **4 warnings / 0 errors**, the exact baseline set
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1218,1225` — the same two exports, renumbered by the
  file's growth); **280 tests / 33 files**, up from the 248 / 32 baseline.
- `npm run test:e2e:compile` passes. It is `tsc --noEmit` only: **browser e2e is deferred by
  D-000.4 and did NOT execute.**
- Nine digests re-measured byte-identical with `git status fixtures/` empty: statement-1 76,744
  `114df1d6`; -5 127,363 `70dce051`; -20 269,884 `56bfbbd9`; -50 555,829 `5d090b0f`;
  mandatory-break 56,681 `7cf743de`; line-spacing 57,770 `de212115`; justified-text 59,894
  `6da3b12e`; alignment-rounding 61,346 `986400a1`; justified-thai 15,079 `58ca4777`. The
  `goldenDigestRecord` declaration still holds twenty entries.
- Fences: `internal/layout/paginate.go` and `component_commands.go` are both **absent** from the
  diff, so `hitTestBand`'s half-open rectangle is untouched. `README.md` appears in no commit and
  its md5 is still `078d7d80d518d54af2fc04fb270d46b8`. No `@media` query added. `SupportedMajor`
  stays 2.

**Residual risks.** The three surfaces where this story's correctness actually lives — the clipping,
the seam painting and real pointer travel — are CSS and layout, which jsdom cannot observe and the
e2e suite does not execute; the echo specificity defect patched here is exactly that class of bug,
and the deferred `SHEET_STACK_GAP` coupling is another. DW-35 becomes **more visible** (N sheets
multiply the text the one hard-coded font stack paints) and remains Epic 8's. DW-34 is multiplied by
the sheet count and is bounded, not closed, by Ruling J's budget. DW-26/27/28/30/31/32/33/34/35/37
stay open and untouched; DW-33 needed no ruling, as the Design Notes predicted.

## Delivery Log

### Dispatch 2 — 2026-08-31, implemented

Baseline: `c95fa9bc6202142a69f52d571f04d86ab12c8edc` on `main`. All seven Parts landed, in one commit.

**What shipped.** `CanvasProjection` gained `ContentWindowOrigins []int64` and `ContentWindowCountIsFloor bool`.
`addCanvasWindowCount` now reads `plan.Pages[i].Shift` — the datum `:557` used to discard while keeping
`len(plan.Pages)` from the same value — so there is no second pagination and no derivation. The floor flag
OR-s its three documented causes: a content-band table with a non-empty binding, the `layout.Paginate`
degradation branch, and a content text element whose font chain would not resolve (carried out of
`addCanvasTextPaint` on the new `canvasColumnExtents` struct, because the extents it would have contributed
are indistinguishable from an element with nothing to say once they are gone). The protocol admits both
fields with a `fontFamilies`-shaped guard. `resize-anchor.ts` reads `BANDS_CAPPING_VERTICALLY` and gates the
ONE `limitHeight` both vertical clamps consume — `:49`/`:50`/`:51` untouched. `sheet-stack.ts` (new, `.ts`)
holds the whole geometry as pure arithmetic; `App.tsx` draws one `.page-surface` per modelled sheet, repeats
both repeating bands, echoes a spanning component onto every window it crosses with exactly one home, and
translates later-sheet placement onto the band-aware `createComponent` that already existed.

**Re-measured at the closing revision**, with a throwaway program outside the repo; the tree was left clean.

| fixture | `contentWindowCount` | `contentWindowCountIsFloor` | `contentWindowOrigins` | rendered pages |
|---|---|---|---|---|
| `statement-1` | 1 | **true** | `[0]` | 1 |
| `statement-5` | 1 | **true** | `[0]` | **5** |
| `statement-20` | 1 | **true** | `[0]` | **20** |
| `statement-50` | 1 | **true** | `[0]` | **50** |
| `page-count-1` | 1 | false | `[0]` | 1 |
| `page-count-5` | 5 | false | `[0 728000 1456000 2184000 2912000]` | 5 |
| `page-count-20` | 20 | false | `[0 … 13832000]`, a round 728000 apart | 20 |
| `page-count-50` | 50 | false | `[0 … 35672000]`, a round 728000 apart | 50 |

The floor is unchanged and is now SAID: the four byte-identical statement templates still report one window
each while printing 1/5/20/50 pages, and the flag is what makes the interface say so. The `page-count-*`
origins are a live restatement of Ruling E — they are the elements' own tops, 728000 apart, where the closed
form would answer 727890 apart, adrift by 110 millipoints per window.

**Deferral register.** **DW-36 is CLOSED**: the drag clamp is the third consumer of the band-containment
tie, reading the same list Go's `containComponent` and the protocol's `isCanvas` read, and
`engine-bounds-mirror.test.ts` now reads all three sources with a fourth site regex and its own drift proof.
**DW-37, DW-33, DW-34 and DW-35 are OPEN and untouched.** DW-37's subject — the unguarded non-text sum in
`addCanvasWindowCount` and the protocol's admission of the resulting integer — was not changed; the drawing
budget is an additive guard on the DRAWING, in the same relation to the count that a truncated paint has to
a value. DW-33 is provably untouched: the origins and the count read `plan.Pages`, `lines`, `originY` and
the vertical model, never `painted`, `budget`, `oversized` or `placed`. DW-34 is multiplied by the sheet
count and bounded, not closed, by `MAX_CANVAS_SHEETS`. **DW-35 is now MORE VISIBLE and was not taken**: the
canvas paints text through one hard-coded font stack, and drawing N sheets multiplies every glyph it paints,
so a chain mismatch that reads as one page of wrong-looking text today reads as up to 120. It stays Epic 8's
plan gate's.

**`MAX_CANVAS_SHEETS = 120`, derived.** Epic 7's narrative target is forty pages. The projection's own
body-text paint budget is 1920 lines (`maxCanvasBodyTextLines`), which at forty-to-fifty lines per A4 window
is under fifty windows of solid prose — past that a window can only come from a declared placement gap. 120
is three times the stated target and more than twice what the paint budget can fill. The value is never
truncated, only the drawing, and the interface says it is showing the first N of M.

**No `.folio` format field was required.** `SupportedMajor` stays **2**. `internal/layout/paginate.go` is
absent from the diff; `folio-go/component_commands.go` is absent from the diff entirely, so `hitTestBand`'s
half-open band rectangle is unchanged; `README.md` is untouched and its md5 is still
`078d7d80d518d54af2fc04fb270d46b8`.

**Verification, measured.**

- `cd folio-go && go test -count=1 ./...` — **exactly one** failure, `TestCorpusMeetsP6ExerciseFloors/P6g`
  (`got 7, need >=20`), the mandated permanent red; untouched. Its drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` is **green**. Every other package: `ok`.
- `cd folio-go && go vet -tags=matrix ./...` — clean, exit 0. `gofmt -l folio-go` from the repo root — no output.
- `go test -tags=matrix -run TestTargetRenderHash` with `FOLIO_MATRIX_TARGET` set, once per leg —
  `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` all **ok**, and each leg's verbose output contained
  `asserts NOTHING` **0 times**, so all four legs really ran.
- `go test -tags=matrix -run TestCrossTargetByteIdentity .` — ok.
- `cd lint && go test ./...` — ok (4 packages).
- **All twenty `goldenDigestRecord` entries hold**, measured: `TestGoldenDigestAgreesAtEveryDeclaredSite`
  passes, `TestStatementGoldenFixtures` passes all four sizes, and every other golden fixture test passes.
  `git status fixtures/` was empty before and after.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; oxlint **4
  `only-export-components` warnings, 0 errors**, exactly the baseline set (`preview/pdf-viewer.tsx:16,17`;
  `App.tsx:1216,1223` — the same two declarations, moved by the added lines); **278 tests / 33 files**, all
  passing, up from the 248 / 32 baseline.
- `cd folio-designer && npm run test:e2e:compile` — passes. It is `tsc --noEmit` only: **browser e2e is
  deferred by D-000.4 and did NOT execute.** The single-window labels were kept exact so the compile-only
  Playwright strict-mode selectors still address one node each.

**End to end, demonstrated rather than asserted about a conditional.**
`TestAComponentAuthoredWindowsDownTheColumnLandsOnItsOwnSheet` creates a component two windows down through
the same band-aware `createComponent` the later-sheet placement sends, serializes to canonical bytes, parses
them back, and asks the projection which sheet the component is on — then moves it five windows down and
asks again, checking the sheet index really increased and the bytes carried the column coordinate. On the
browser side, `App.test.tsx` drags a component from sheet one across a seam and asserts the committed
payload is `{"kind":"moveComponent",…,"y":700,…}` — 700pt being `contentWindowOrigins[1]` in points.

**Red proofs, run and recorded.**

1. **The origin seam against the closed form.** Replacing the projected origin with the window height
   multiplied by an index turned **13 tests red** across `sheet-stack.test.ts` and `App.test.tsx` — the
   sheet origins, both seam cases, the spanning component, the budget, all three zoom round trips and both
   drag assertions. Spelled in `sheet-stack.ts`'s own identifiers it additionally turns the authority
   contract's new prohibited pattern red. (`sheetStack`'s local is deliberately named `windowHeight`, not
   `height`, so the text guard can see the plausible mutation.)
2. **The AC4 sentences against deletion.** Emptying `canvasColumnClaim`, `canvasFloorClaim` and the
   per-component notice turned **3 tests red**: the multi-sheet disclosure, the floor disclosure and the
   later-sheet accessible name.
3. **The content-band clamp lift against restoring the clamp.** Restoring
   `limit ? limit.height : Number.POSITIVE_INFINITY` turned **4 tests red**: the content-band move and south
   resize in `resize-anchor.test.ts`, both mirror assertions that read `resize-anchor.ts` as the third
   consumer, and the App drag that runs 800pt down a 729.89pt window (pinned at 705.89px with the clamp back).

**One judgement recorded, not hidden.** `MAX_ENGINE_CONTENT_WINDOWS = 100_000` in `engine-protocol.ts` is a
channel backstop in the shape of `MAX_ENGINE_DIAGNOSTICS`, deliberately **not** added to the mirror's
numeral `pairs` (still six, still `toHaveLength(6)`): Go declares no maximum window count, so there is
nothing on the other side for it to drift against, and the constant is set orders of magnitude above
anything the projection can produce because a field this side rejects discards the whole snapshot.
