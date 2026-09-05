---
title: 'A band boundary is dragged on the canvas'
type: 'feature'
created: '2026-09-05'
status: 'done'
baseline_commit: 'cfa426330c8171313f6cae4160547f665354f146'
review_loop_iteration: 0
context:
  - '{project-root}/_bmad-output/implementation-artifacts/source-text-guard-inventory.md'
---

## In plain terms (read this first if you just want the gist)

*Not normative, and rewritten after the fact: the frozen Intent below governs implementation, while
this section says what actually shipped.*

You can now grab the line between two bands on the canvas and drag it, so fitting a letterhead is a
gesture instead of a number you guess and retype. Nothing is committed while you drag: a proposed line
follows your pointer and the height reads out in points beside it. When you let go, the designer sends
the one command Story 12.1 already built, and the canvas redraws from whatever the engine accepted.
The same boundary is reachable with the Tab key and nudged with the arrow keys, because a gesture must
never be the only way to reach a value.

The drag stops by itself at zero and at the point where no content area would be left. It deliberately
does **not** stop short of a component standing in the way — that proposal is sent, so the engine can
refuse it and name the component. A silent stop there would throw away the one thing the author needs
to know.

Two defects were caught before release that nothing running could see, because the tests apply no
stylesheet and stay at the default zoom: the proposed line would have painted in the wrong place
entirely, and at any other zoom the height sent would have been malformed. A whole-browser test of the
gesture was written, but no workflow runs that suite, so it is not coverage.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A band height is authorable only as a typed number in the Page Setup panel. The canvas
already draws the two band boundaries as dashed cyan rules with a labelled tab beside each, and they
are exactly where an author reaches to resize a letterhead — but they are inert decoration
(`.page-band-content::after` / `.page-band-pageFooter::after`, `pointer-events: none`). Nothing on the
canvas is keyboard-reachable for a band height either.

**Approach:** Put one focusable hit strip on each of the two real band boundaries — the
header/content boundary and the content/footer boundary — co-located with the CONTENT and PAGE FOOTER
tabs that already mark them. Press-and-drag paints a proposed boundary line and a points readout as a
purely local proposal; release sends the one `setBandHeight` command Story 12.1 shipped and the canvas
re-projects from the engine's accepted document. The same strip is a tab stop and takes the arrow-key
nudge. No band rect is re-placed by the browser during the gesture, and no engine rule is
re-implemented: only the proposed boundary LINE moves, at `original ± Δ` over one projected number.

## Boundaries & Constraints

**Always:**
- **The engine owns the document (AD-15).** The drag is transient UI state. Exactly one command is
  sent, at release, and only when the proposed height differs from the projected one — the same
  send-only-if-changed rule 12.1 established, which is also what keeps a no-op gesture out of undo
  history (`wasm/engine.go`'s byte-equality short circuit).
- **The command is the one Story 12.1 shipped**: `setBandHeight`, encoded only by
  `bandHeightCommand` in `band-height-command.ts`, routed through `commandBytes`. No second encoder,
  no second arm, no page-setup widening.
- **The browser measures nothing (AD-17).** Pointer geometry is `clientY` arithmetic against the
  press anchor, through `canvasDisplay.documentDelta`, exactly as `CanvasComponent`'s `begin`/`move`
  and `beginProseResize`/`moveProseResize` already do. `getBoundingClientRect`, `offset*`, `client*`
  metrics, `getComputedStyle` and `ResizeObserver` remain unreachable.
- **Bands are not re-placed by the browser during a drag (AD-24).** The three `.page-band` rects keep
  the geometry the engine projected for the whole gesture. What moves is one additional proposed
  boundary line and one readout. Band placement stays `internal/layout`'s.
- **Any browser-side bound this story builds must CONSUME the engine's own declaration, not re-spell
  it**, and `engine-bounds-mirror.test.ts` must read it doing so — the standing condition under which
  the drag clamp in `resize-anchor.ts` was allowed to exist (DW-36).
- **A refusal reaches the author through the existing canvas `role="alert"`** (`commitError`, via
  `componentDiagnostic`), never through `pageSetupDiagnostic`.
- **One interactive boundary per document.** The canvas draws 3N band sections for an N-page stack;
  the document has one page-header height. Only the home sheet's boundaries are interactive and
  accessibly named, following the `occurrence.home` / `ComponentEcho` idiom already shipped.
- **Byte-identity holds.** A document whose band heights are never edited serializes identically, and
  `content` still carries no `height` key.

**Ask First:**
- **Settled at CHECKPOINT 1 as R1 — do not re-open.** Narrowing D-12.1-Q4 to the panel, so the
  gesture may hold a mirrored ceiling. If implementation finds a reason this is wrong, that is an
  intent gap: HALT, do not decide it.
- Building the strand floor in the browser. R1 declined it deliberately. If it starts to look
  necessary, HALT.
- Any change to `bandsCappingVertically`, `containComponent` or `containEdgeY` (12.1's standing
  Ask First).
- Minting a diagnostic code, or adding any vertical clip behaviour to the render path.
- Adding a token to `DESIGN.md`, or any `@media` rule to `App.css` (the latter is forbidden outright
  by `canvas-authority-contract.test.ts`).
- Changing `setBandHeight`'s arity beyond the one field this story's snapping AC requires.

**Never:**
- No browser-side model of the document, no second content-window arithmetic, no re-derivation of
  band origins.
- No streaming of commands during a drag — one command at release, or none.
- No new `@media` rule in `App.css`; no DOM measurement of any kind; no rounded corners; no third
  accent colour.
- **No `#hex`, no `rgb(`, no `hsl(` anywhere in `App.css` — including inside a comment** — and no
  `linear-gradient`/`conic-gradient`; `design-contract.test.ts` bans these over the raw file text. Any
  `border-radius` must be `var(--radius…)`. Colour comes from `var(--color-*)` only.
- No new source-text guard written as a regex where an AST or a structural parse would do (D-000.27).
  Any guard this story does add must **declare its comment policy and assert it in both directions**,
  and must **fail loudly** — a non-vacuity floor or a throw — if its extraction returns empty.
- No vertical clip, no new diagnostic code, no change to FR44's horizontal-only clip axis (D-2.8.1).
- Do not make the PAGE HEADER tab draggable, and do not give it a resize affordance — the edge above
  it is the page margin.
- Do not touch the `placementPoint` … `function pageStyle` adjacency in `App.tsx`.

## I/O & Edge-Case Matrix

| # | Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|----------|--------------|---------------------------|----------------|
| 1 | Drag the header/content boundary down | press on the content boundary strip, move +40pt worth of pixels, release | proposed line tracks the pointer, readout shows the proposed page-header height in points; on release one `setBandHeight` `pageHeader` command; canvas re-projects | N/A |
| 2 | Drag the content/footer boundary up | press on the footer boundary strip, move −25pt worth, release | footer grows: proposed line tracks upward, readout shows the proposed page-footer height; on release one `setBandHeight` `pageFooter` command | N/A |
| 3 | PAGE HEADER tab | press and move on the page-header band's top edge | nothing resizes; no strip exists there; no command | N/A |
| 4 | Drag then release on the value already in force | press, move, return to the start, release | **no command is sent**; no history entry; no dirty mark | N/A |
| 5 | Sub-threshold press | press and release with < 2px travel | treated as a press, not a drag; no command | N/A |
| 6 | Snapping on | snap enabled, drag to a non-grid value, release | the committed height is on the grid increment; snapping is applied by the engine, exactly as every other gesture's is | N/A |
| 7 | Snapping off | snap disabled, drag, release | the committed height is the unsnapped proposal | N/A |
| 8 | Drag toward a negative height | pointer travels past the band's own origin | the proposal stops at 0 — a negative band height is a property of the FIELD (17.4) and is never proposed | N/A |
| 9 | Drag past the point where no content window remains | pointer travels past `innerH − other − 1` | the proposal **stops** at the ceiling; the boundary does not follow further; the released value is legal and is accepted | N/A — the impossible document is never proposed |
| 10 | Drag that would strand a component | band shortened past a component's lowest edge, released | the command is refused by 12.1's strand check; the proposal is discarded; the author reads `a pageHeader height of Npt would leave e1 outside the band: it reaches Mpt` | canvas `role="alert"`; document byte-unchanged |
| 11 | Keyboard nudge | boundary strip focused, ArrowUp / ArrowDown | the boundary resizes by 1pt; Shift by 10pt; one command per press | engine refusal via `role="alert"` |
| 12 | Keyboard nudge with a component also selected | one component selected AND the boundary focused, ArrowDown | the boundary moves; the component does **not** | N/A |
| 13 | Multi-page stack | a document drawing 3 sheets | exactly two interactive boundary strips exist, on the home sheet; the other sheets' boundaries stay inert and unnamed | N/A |
| 14 | Non-primary button | right-button press on the strip | no drag begins | N/A |
| 15 | Pointer lost mid-drag | `pointercancel`, or a move with `buttons === 0` | the gesture ends, the proposal is discarded, no command is sent | N/A |
| 16 | Second pointer during a drag | a second `pointerdown` with a different `pointerId` | ignored; the first pointer owns the gesture | N/A |
| 17 | Unedited document | open and serialize, no band height touched | bytes identical; `content` carries no `height` key | N/A |

</frozen-after-approval>

## Rulings Carried In

All four plan-gate questions were **ruled at CHECKPOINT 1 on 2026-09-05**. Nothing below is open.

**R1 — BUILD the content-window ceiling as a drag limit; do NOT build the strand floor.** The lead
corrected **its own** prior ruling, not this spec's reading. D-12.B's sentence — *"the objection is the
browser computing layout geometry at all"* — was **over-broad and would have blocked correct work.**
17.4 item 9's operative qualifiers are *"in the inspector"* and *"quietly-drifting"*: DW-36 defeats
*quietly* by requiring the copy to read the engine's declaration and be caught doing so, and a canvas
boundary has **no typed counterpart**, so *consistency with typing* has nothing to be consistent with.
**17.4 forbids an unmirrored clamp in a control that has a typed counterpart; it does not forbid a
mirrored clamp on a gesture that has none.** D-12.B's *conclusion* stands unchanged — the band-height
**panel field** is typed and still gets no floor — only its ground was too wide. Verified by the lead:
`resize-anchor.ts`'s `limitHeight` and both clamp sites, and `engine-bounds-mirror.test.ts`'s
`dragClampPath` declaration, which puts the clamp inside the mirror census.

**R2 — AC7 is discharged by 12.1's refusal**, restated as Matrix row 10. The planning text has been
corrected at HEAD `f6416c3`: **both** places `epics.md` claimed a short band clips are struck, with the
measurement beside them. The coordinator's sweep found **exactly two** instances of the false premise,
both on the vertical band axis (12.1's AC and 12.5's AC7), and **no third** waiting in Epic 13 or 14.
**Story 7.3's clip-and-warn survives untouched** — it concerns a justified line exceeding its declared
**width**, the one axis where `TEXT_CLIPPED_WIDTH` exists. The premise was false only where it was
applied vertically; a sweep that struck every mention of FR44 clipping would have destroyed a true
criterion alongside the two wrong ones.

**R3 — Add the fifth field `snap`; the panel passes `false`.** Two conditions, both binding:
1. **The 22 payload literals move in ONE commit with the arity gate.** A strict-arity check and its
   callers are a single change; splitting them leaves a window in which every command is refused.
2. **State "byte-preserved" precisely.** The **command payload** DOES change — it gains
   `"snap":false`. What is preserved is the **document bytes** the panel's typed path produces. 12.1
   carries a byte-identity AC, so an imprecise phrasing invites a reviewer to see the payload move and
   file a defect against a story that did exactly what it said.

**R4 — Matrix and Playwright: both deferred to the Epic 12 boundary gate; the e2e spec is WRITTEN
here.** See Verification → *Cadence*.

**Standing expectation set at this checkpoint:** a story citing a mockup **states what it found there,
including absence, with the positive control.** The design record is now known unreliable in both
directions — over-promising in one place (14.8), silent in another (no affordance drawn for a
capability `EXPERIENCE.md` says exists). Neither is discoverable without a stated measurement.

## Code Map

Every anchor below was measured at HEAD `8b91afe` (tree clean but for the workflow's own recompiled
`epic-12-context.md`). **Cited by symbol; line numbers rot.**

### The command — already built, do not rebuild

- `folio-go/component_commands.go` — `setBandHeight(t *Template, raw map[string]json.RawMessage)`, the
  25th arm of `ApplyComponentCommand`'s switch. Gates: `componentFields(raw, 4)` over
  `kind`/`version`/`band`/`height`; `commandString(raw,"band")` restricted by
  `slices.Contains(bandsCappingVertically, name)`; `lengthField(raw,"height")`; `proposed < 0`;
  `bandsLeaveContentWindow`; then `containComponent` over **every** element of the band against a
  candidate `CanvasBand{Name, Width, Height: proposed}`. `literal := string(raw["height"])` is echoed
  in every refusal. `bandHeightPath(name)` builds DataPath `bands.<band>.height`.
- `folio-go/page_setup.go` — `bandsLeaveContentWindow(header, footer, innerH)` and
  `bandContentWindowCeiling(other, innerH) = innerH - other - 1`. **One implementation, two callers**
  (`Canvas` at load, `setBandHeight` at authoring) — 12.1's Q3 condition. `const GridIncrement int64 =
  6000` (6pt) and `SnapToGrid(proposed geom.Length)` live here too.
- `folio-go/internal/layout/band.go` — `ContentHeight(g PageGeometry)`, `Origins(g)`. Band placement
  lives here alone (AD-24). Do not mirror it.
- `folio-go/internal/geom/snap.go` — `(Length).SnapNearest(increment)`, nearest multiple, exact halves
  away from zero.
- `folio-go/wasm/engine.go` — `(*Engine).Apply`; `if bytes.Equal(canonical, e.bytes) { return
  e.Snapshot(), nil }` is why a no-op gesture costs no history entry. No band-specific branch exists
  here; keep it that way.
- `folio-go/band_height_command_test.go` — 22 tests, the whole 12.1 matrix.
  `TestSetBandHeightResentUnchangedLeavesTheBytesIdentical` and
  `TestUneditedDocumentSerializesByteIdentically` are this story's byte-identity gate.
  `TestSetBandHeightRefusesToStrandAComponent` / `…ThatIsNotTheFirst` / `…InTheFooter` pin the strand
  refusal. **Read-only unless Q3 changes the arity**, in which case
  `TestSetBandHeightIsRefusedByTheDoorsExistingGates` and every payload literal in the file move.

### The encoder

- `folio-designer/src/band-height-command.ts` — `bandHeightCommand(band: CappingBand, height: string)`,
  emitting `['band', jsonString(band)], ['height', jsonNumber(height)]` through `commandBytes`. Its
  header comment carries D-12.1-Q4's ruling verbatim (*"THIS MODULE HOLDS NO BOUND OF ITS OWN"*) — if
  Q1 is answered "build the limit", **amend that comment to scope the ruling to the panel; do not
  delete it.**
- `folio-designer/src/band-height-command.test.ts` — byte-exact payload and exact key order (4 tests).
- ⚠ **`folio-designer/src/engine-bounds-mirror.test.ts` line-anchors this file's import:**
  `expect(factory).toMatch(/^import type \{ CappingBand \} from '\.\/engine-protocol'$/m)` and
  `expect(factory).not.toMatch(/'pageHeader' \| 'pageFooter'/)`. Adding a second type to that import
  line reds it. Import anything new on its own line.
- ⚠ `folio-go/command_json_authority_wire_test.go` anchors
  ``/(?m)^import \{[^}]*\} from '\.\/command-json'$/`` — single-line only. If a formatter wraps
  `band-height-command.ts`'s value import across lines, this reds for a non-defect.
- `folio-designer/src/command-json-soleness.test.ts` — the named-factory list already contains
  `band-height-command.ts`.

### The canvas

- `folio-designer/src/App.tsx` — `sheetSurface(projection, model, sheet)` renders
  `<section className="page-surface">` and maps `projection.bands` to
  `<section className={`page-band page-band-${band.name}`} … tabIndex={0} style={bandStyle(band, zoom)}>`.
  The band tab is that section's **first child `<span>`**, class-less, styled only by `.page-band >
  span`. Its label comes from `bandName(name)` → `'Page Header' | 'Content' | 'Page Footer'`
  (uppercased by CSS). **The tab has no role, no tabIndex, no handler.** All four band handlers
  (`onPointerEnter/Leave/Up/KeyDown`) are gated on `placing`.
- `folio-designer/src/App.tsx` — `canvasDisplay` (`css(mp, zoom)` → px string; `documentDelta(px,
  zoom)` → **points**, callers `* 1000` for millipoints), `points(value)` (millipoint → trimmed
  decimal string), `bandStyle` (writes `--band-x/-y/-width/-height`), `pageStyle` (writes
  `--page-display-width/-height`, `--grid-display-pitch`, `--page-margin-left/-right`; **there is no
  `--page-margin-top`**).
- ⚠ `folio-designer/src/App.tsx` — `export function placementPoint(event: Pick<MouseEvent, 'offsetX' |
  'offsetY'>, …)` must remain **immediately** followed by `function pageStyle` at column 0, with
  nothing between the closing brace and that line. This is the `offsetX/offsetY` exemption's seam and
  `canvas-authority-contract.test.ts` `expect`s it. **12.5 needs neither `offsetX` nor `offsetY` — use
  `clientY` deltas and leave the seam alone.**
- `folio-designer/src/App.tsx` — `CanvasComponent`'s `begin` / `move` / `finish`: the drag idiom.
  `type DragState` records `startClientX/startClientY` and the `changed` gate
  (`Math.abs(rawDX) >= 2 || Math.abs(rawDY) >= 2`). The proposal survives release via
  `.finally(() => setDrag(undefined))`. The readout is
  `<span className="canvas-dimension" aria-hidden="true">{points(active.width)} × {points(active.height)}</span>`.
- `folio-designer/src/App.tsx` — `beginProseResize` / `moveProseResize` / `endProseResize` and the
  `<span className="property-prose-resize" aria-hidden="true">` strip. **The hardened pointer
  bookkeeping to copy**, all four items review-found in 17.5: `if (event.button !== 0) return`;
  pointerId ownership check; `event.preventDefault()` before focus moves; anchor recorded **before**
  `setPointerCapture?.()` (which throws `NotFoundError` on an inactive pointerId); and
  `if (event.buttons === 0) { end…; return }` in `move`, **after** the id check. The canvas drag has
  none of these — take them from here, not from `CanvasComponent`.
- `folio-designer/src/App.tsx` — `nudgeSelection(dx, dy)` and the window `keydown` `shortcut` handler:
  `mode === 'design' && selectedRef.current.length === 1 && ['ArrowLeft','ArrowRight','ArrowUp','ArrowDown'].includes(event.key)`,
  step `event.shiftKey ? 10_000 : 1_000`, early return on `isEditableTarget(event.target) ||
  event.isComposing`. **This is selection-driven, not focus-driven** — it will fire alongside a
  focused boundary handle unless the boundary's own handler stops the native event (Matrix row 12).
- `folio-designer/src/App.tsx` — `commitComponent(payload, after?)`, `componentDiagnostic(error)`,
  `setCommitError`, `clearInteraction()`, and the single canvas alert
  `{commitError && <p role="alert" className="file-message">{commitError}</p>}`, last child of
  `<main className="canvas-region">`. ⚠ **Measured: no test asserts the rendered text of a
  canvas-gesture refusal** (`/usr/bin/grep -n 'commitError\|file-message' src/App.test.tsx` → 0 hits;
  positive control `points(` fires). This story should close that.
- `folio-designer/src/App.tsx` — `PageSetup`'s `Field label="Page header height (pt)"` /
  `"Page footer height (pt)"` rows, `draftFor`, `projectedBandHeight`, `bandDraft`, and the
  `for (const band of CAPPING_BANDS)` send loop inside `applyPageSetup`. 12.1's typed path; its
  behaviour must not change.
- `folio-designer/src/App.css` — `.page-band` (`--band-x/-y/-width/-height`, absolute);
  `.page-band > span` (the tab: `left: calc(-1 * var(--band-x) - 8px); transform: translateX(-100%)`);
  **`.page-band-content::after, .page-band-pageFooter::after`** — the boundary rule:
  `left: calc(-1 * var(--band-x) - 14px); width: calc(var(--page-display-width) + 28px); top: 0;
  border-top: var(--hairline) dashed var(--color-select); opacity: 0.75; pointer-events: none;`.
  `.page-band:focus-visible, .page-band-target { outline: 2px solid var(--color-select); outline-offset: -2px }`.
  `.property-prose-resize { position:absolute; left:0; right:0; bottom:-3px; height:7px; background:transparent; cursor: ns-resize; touch-action: none }` — the strip idiom, and the only
  `ns-resize` in the product. `.band-window { position:absolute; inset:0; overflow:hidden;
  pointer-events:none }` — **a handle must sit outside `.band-window` or it is clipped and inert.**
  ⚠ **Exactly one `@media` may exist in this file** (`prefers-reduced-motion: reduce`), asserted by
  exact-equality on an ordered array.
- `folio-designer/src/sheet-stack.ts` — `sheetStack(canvas)`, `SHEET_STACK_GAP`, and
  `occurrence.home` / `sheet.index === 0`, the existing rule for "one occurrence is interactive, the
  rest are echoes". `ComponentEcho` + `.canvas-echo { pointer-events: none; cursor: default }` is the
  inert-copy precedent.
- `folio-designer/src/engine-protocol.ts` — `CanvasProjection` (`width`, `height`, `marginTop/Right/
  Bottom/Left`, `gridIncrement`, `bands`, `contentWindowHeight/Count/Origins/CountIsExact`,
  `components`), the inline `bands` member type, `BANDS_CAPPING_VERTICALLY`, `CAPPING_BANDS`
  (`toBe`-identical to it), `type CappingBand = Exclude<…,'content'>`. `isCanvas` enforces exactly
  three bands in order and **contiguity** (`paint.y === prior.y + prior.height`) — so
  `bands[0].height + bands[1].height` is the header's ceiling input with no extra arithmetic.

### The guards that will notice

- `folio-designer/src/canvas-authority-contract.test.ts` — 14 prohibited regexes over `src/**`
  (production **and** unit tests) and `e2e/**`, plus a production-only refusal-vocabulary scan.
  **Strips comments** (`withoutComments`, quote-aware). Bans `getBoundingClientRect|getClientRects`,
  `offset(Width|Height|Left|Top|Parent)`, `client(Width|Height|Left|Top)`,
  `scroll(Width|Height|Left|Top)`, `offsetX|offsetY`, `ResizeObserver`, `getComputedStyle(`,
  `devicePixelRatio`, `document.fonts`, `new FontFace`, `measureText`, `Range|getSelection`, a
  `text-align: justify`-shaped CSS pair, and two `lines.length` arithmetic shapes. **`clientX`,
  `clientY`, `buttons`, `button`, `pointerId`, `setPointerCapture` and `touch-action` are NOT banned.**
  Population floors: production ≥ 58 (`.tsx` ≥ 8, `.css` ≥ 3), tests ≥ 51, e2e ≥ 15.
- `folio-designer/src/engine-bounds-mirror.test.ts` — 31 `it(`s. Reads `page_setup.go`,
  `component_commands.go`, `linespacing.go`, `locale.go`, `engine-protocol.ts`, `resize-anchor.ts`,
  `App.tsx`, `band-height-command.ts`. Its `pairs` table ties each Go constant to its TS mirror **and
  to the site that consumes it**. It line-anchors the drag clamp:
  `/^ {2}const limitHeight = limit && BANDS_CAPPING_VERTICALLY\.includes\(limit\.band\) \? limit\.height : Number\.POSITIVE_INFINITY$/m`.
  **This is the file a new mirrored bound joins** (Q1).
- `folio-designer/src/resize-anchor.ts` — `proposedBounds(anchor, origin, dx, dy, limit?)`, `DragLimit`,
  `minimumSize = 1000`. **The precedent that decides Q1** — see Design Notes.
- `folio-designer/src/transient-interaction.test.ts` — a proposal stays UI-local and reaches the engine
  only on explicit commit.
- `folio-designer/src/design-contract.test.ts` — token-name equality with `DESIGN.md`. `band-tab` and
  `band-boundary` are already declared components. **It does not read `App.tsx`.** Adding a token
  requires a `DESIGN.md` edit (Ask First).
- ⚠ `folio-designer/src/App.test.tsx` — *"states its own test count truthfully in its header"* reads
  its own source; its `spelled` map **ends at `'TWENTY-THREE': 23`** and the current count is exactly
  23. Adding an `it(` to that describe without extending the map fails on `toBeDefined()`.
- `folio-designer/src/property-prose-height.test.ts` — 10 tests over `App.css` pinning the prose
  strip's geometry. Read-only, but it is the shape a boundary-strip CSS test should copy.

### Source-text guards over the files this story edits (D-000.27, `source-text-guard-inventory.md`)

Measured against the inventory **and re-measured directly at this plan gate**. Legend: **STRIPS** =
comments blanked before matching; **RAW** = a commented-out occurrence counts as live;
**wrap-fragile** = the pattern needs its phrase to stay on one line.

| File this story edits | Guard that reads it | Comment policy | Wrap-fragile | What it constrains here |
|---|---|---|---|---|
| `App.tsx` | `canvas-authority-contract.test.ts` | **STRIPS** | no | The 14 measurement bans. A drag handler is exactly the code that reaches for them — **this story needs none of them**, see below. |
| `App.tsx` | `engine-bounds-mirror.test.ts` | **RAW** | **yes** — every extraction `^…$` | `expect(app).toContain('CAPPING_BANDS')`, `expect(app).not.toMatch(/\['pageHeader', 'pageFooter'\]/)`. Do not spell the band list in `App.tsx`. |
| `App.tsx` | `property-prose-height.test.ts` | **RAW** | **yes** | Pins `PROSE_MIN_HEIGHT_PX = 72` against the CSS floor, and counts `<span className="property-prose-resize"` at **exactly 1**. **Measured: its "exactly one" assertions are keyed to the `property-value-prose` / `property-prose-resize` CLASS NAMES, not to drag affordances in general — a `.band-boundary-handle` does not trip them.** |
| `App.tsx` | `command-json-soleness.test.ts` | partial (whole-line `//`) | no | No command JSON assembled outside a named factory. |
| `App.tsx` | `engine-ownership-contract.test.ts` | RAW (`.postMessage(` arm) | no | One Worker-owning module. |
| `App.css` | `design-contract.test.ts` | **RAW** for the colour ban | **yes** for its `toContain` literals | **`not.toMatch(/#[0-9a-f]{3,8}|\b(?:rgb\|hsl)\(/i)` over the whole file, comments included**; `not.toMatch(/border-radius:(?!\s*var\(--radius)/)`; no gradients; and the exact one-line `.tree-item:focus-visible { outline: 2px solid var(--color-select); outline-offset: -2px; }` must not be reflowed. |
| `App.css` | `canvas-authority-contract.test.ts` | STRIPS | no | **Exactly one `@media`**, by ordered-array equality. Bans a `text-align: justify`-shaped pair. |
| `App.css` | `property-prose-height.test.ts` | RAW | **yes** | Extracts rules as `^<selector> \{[^}]*\}`. Its allowlist test (`permitted = ['position','left','right','bottom','height','background','cursor','touch-action']`, exact set equality) is **the shape to copy for `.band-boundary-handle`** — review already forced the denylist→allowlist move once, after `border-top`, `box-shadow` and `outline` all slipped through. |
| `App.css` | `canvas-font-stack.test.ts` | STRIPS | yes (`line.startsWith`) | Canvas font-family stack as an exact ordered list. |
| `band-height-command.ts` | `engine-bounds-mirror.test.ts` | RAW | **yes** | `^import type \{ CappingBand \} from '\./engine-protocol'$` — exact, line-anchored. New imports go on their own line. |
| `band-height-command.ts` | `command_json_authority_wire_test.go` | partial | **yes** | `^import \{[^}]*\} from '\./command-json'$` — single-line only; a wrapped value import reds it. |
| `component_commands.go` | `engine-bounds-mirror.test.ts` | RAW | **yes** | The `containComponent` / `containEdgeY` spellings. |
| `engine-protocol.ts` | `engine-bounds-mirror.test.ts`, `canvas_projection_wire_test.go` | RAW | **yes** | The projection's band-name union and the wire key set. |

**The measurement that keeps this story out of the seam:** a boundary drag is a **delta** gesture, so it
needs `clientY` only. `clientX`/`clientY`/`buttons`/`button`/`pointerId`/`setPointerCapture`/
`touch-action` are **not** on `canvas-authority-contract.test.ts`'s prohibited list — only
`client(Width|Height|Left|Top)` is. So this story adds no measurement site, needs no widening of the
`placementPoint` local-pointer-input seam, and must leave that seam byte-untouched.

### Read-only evidence gathered at the plan gate (do not re-derive)

- **The strand is REFUSED, not clipped.** Throwaway module outside the repo, exported API only,
  deleted, tree verified: header 100 → 20 with `e1` at `y=50 h=30` → `folio: a pageHeader height of
  20pt would leave e1 outside the band: it reaches 80pt`. **Partial** overflow (`y=10 h=30`) refuses
  too. `TestSetBandHeightAcceptsTheLowestOccupiedEdge` pins that `y+h == height` is accepted.
- **FR44 has no vertical axis.** `internal/diag/diag.go` declares **17** codes; the only clip codes are
  `TEXT_CLIPPED_WIDTH` and `TABLE_ROW_CLIPPED_HEIGHT`. `render.go` states *"D-2.8.1 fences the vertical
  axis out entirely, so there is no axis field to carry"* and *"the declared WIDTH is FR44's only clip
  bound"*. **AC7's clip does not exist and cannot be built without minting a code.**
- **Zero is a legal band height.** Probe: `{"height":0}` on an empty band is ACCEPTED and serializes
  `"height": 0`. Negative is refused (`a pageFooter height of -5pt is negative`). No test covers zero.
- **A hand-edited strand loads in Go and kills the browser.** `ParseTemplate` nil, `Canvas` nil — but
  `engine-protocol.ts`'s `isCanvas` band clause
  (`BANDS_CAPPING_VERTICALLY.includes(component.band) && !(box.y + box.height <= band.height)`)
  rejects the projection → `PROTOCOL_INVALID` → worker terminated. That, per 12.1's Spec Change Log,
  is the real reason the engine refuses a strand.
- **No focus-order test would break.** Every order assertion in the designer is inside the Table
  Editor dialog, the Font Browser dialog, or the DataPanel tree; both `trapDialog` copies scope their
  `querySelectorAll` to `dialog.current`. The canvas has no focus-order census and no e2e tabs through
  it. **The orchestrator's `trapDialog` hazard is measured NOT to fire.** What *is* pinned is exact
  accessible naming: `e2e/application-shell.spec.ts` and `e2e/browser-native-roundtrip.spec.ts` use
  `getByRole('region', { name: 'Page Header', exact: true })` under Playwright strict mode.
- **Bands are drawn once per sheet.** `App.test.tsx` asserts
  `document.querySelectorAll('.page-band-pageHeader')` has length **3** for a 3-sheet stack.
- **No band-drag machinery exists.** No `onPointerDown` on any band; no `setPointerCapture` on the
  canvas outside `CanvasComponent`; no test queries a band tab element.
- **`gridIncrement` is projected but the browser never snaps.** `/usr/bin/grep -rn 'gridIncrement'
  folio-designer/src` → `pageStyle`'s custom property, the protocol guard, and fixtures. Nothing
  rounds a coordinate in the browser.
- **The mockup gives the tabs zero draggability signal.** `Main.dc.html` draws all three tabs as
  byte-identical static labels at `left: -104px`, and the boundaries as `1px dashed #58a6c4;
  opacity: 0.75` overhanging 14px. `cursor` appears once in that file (`cursor: default`, on a menu
  item); `ns-resize`/`grip` appear **zero** times in any mockup (positive control: 12 `cursor` hits in
  `Font Browser.dc.html`). `DESIGN.md`/`EXPERIENCE.md` contain **no cursor vocabulary at all** (rc=1;
  positive control `select` → 47 hits in DESIGN.md).
- **The design record's only warrant for this story** is `EXPERIENCE.md`'s Band pattern: *"A
  container: always visible, always labelled. Cannot be deleted, reordered, or nested. **Resizable in
  height.** Accepts components by drop."*
- **The snap increment is unwritten in prose.** "grid 2mm" exists only as mockup status-bar text; the
  shipped increment is `GridIncrement = 6000` (6pt).

## Tasks & Acceptance

**Execution:**

> **R3 CONDITION — the next four tasks are ONE commit.** A strict-arity gate and every caller that
> feeds it are a single change. Landing the gate before the callers, or either before the payload
> literals, leaves a window in which every band-height command is refused. Do not split them, and do
> not let a review loop split them either.

- [x] `folio-go/component_commands.go` — extend `setBandHeight` to `componentFields(raw, 5)` with a
      `snap` boolean, read by the same helper the component arms use, and apply `SnapToGrid` to the
      proposed height **before** the negative, content-window and strand checks so every refusal names
      the value that would actually be written. — R3: Go stays the only place a coordinate is snapped.
- [x] `folio-go/band_height_command_test.go` — add: snap on rounds to the grid; snap off does not;
      a refusal after snapping names the snapped value; and a red-proof deleting the `SnapToGrid` call.
      Update every existing payload literal to carry `snap`. — The arity change touches all 22 tests.
- [x] `folio-designer/src/band-height-command.ts` — `bandHeightCommand(band, height, snap: boolean)`
      emitting `band`, `height`, `snap` in that order. **Amend the header comment to NARROW
      D-12.1-Q4's ruling to the PANEL — do not delete it — and state WHY the panel keeps no bound
      while the gesture gets one**: the panel field is typed, so 17.4's asserted property
      (*consistency with typing*) has something to be consistent with; a canvas boundary has no typed
      counterpart, so that property is vacuous there, and DW-36's mirror census answers the
      *quietly-drifting* half. Without that sentence the next reader meets an asymmetry with no reason
      attached. Keep the `CappingBand` import on its own line, unchanged (it is line-anchored by
      `engine-bounds-mirror.test.ts`). — The one encoder.
- [x] `folio-designer/src/band-height-command.test.ts` — extend the byte-exact payload and key-order
      tests for the fifth field; assert `snap: false` and `snap: true` produce distinct bytes. — Pins
      the wire.
- [x] `folio-designer/src/band-boundary.ts` (new) — `proposedBandHeight(band, originalHeight, dy,
      limit)`: pure millipoint arithmetic returning the proposed height, clamped to `[0, limit]`, with
      the sign flip for `pageFooter` (a footer grows as the pointer rises). `limit` is supplied by the
      caller; this module holds no bound of its own. Export `bandBoundaryCeiling(bands, band)` reading
      **only** `bands[i].height` off the projection plus the mirrored strict-positivity margin. —
      Q1: one small pure module, testable and mirror-readable, keeps the arithmetic out of `App.tsx`.
- [x] `folio-designer/src/band-boundary.test.ts` (new) — the pure arithmetic: both bands, both
      directions, the floor at 0, the ceiling, the zero-delta identity (a drag that returns to the
      start yields exactly the original height), and a red-proof deleting each clamp. — D-000.14.
- [x] `folio-designer/src/engine-bounds-mirror.test.ts` — add a pair tying Go's
      `bandContentWindowCeiling`'s `- 1` to the TS margin constant **and** to the site in
      `band-boundary.ts` that consumes it, with a one-sided-edit red proof, in the shape the drag-clamp
      pair already uses. — DW-36's condition: a browser bound must read the engine's declaration.
- [x] `folio-designer/src/App.tsx` — render, inside `.page-band-content` and `.page-band-pageFooter`
      **on the home sheet only**, a `<button type="button" className="band-boundary-handle">` with
      `aria-label` `Resize the page header` / `Resize the page footer`; handlers
      `onPointerDown/Move/Up/Cancel` and `onKeyDown`. Add `boundaryDrag` transient state and, while it
      is active, a `<span className="band-boundary-proposal" aria-hidden="true">` at the proposed
      offset plus a `<span className="band-boundary-readout" aria-hidden="true">{points(proposed)}</span>`.
      Release sends `bandHeightCommand(band, points(proposed), snapEnabled)` through `commitComponent`
      **only when the proposed height differs from the projected one**. Copy 17.5's four pointer
      guards. The `onKeyDown` handles Arrow/Shift+Arrow at 1pt/10pt and must stop the native event so
      the window `shortcut` handler does not also nudge a selected component. — AC1–AC6; one handle
      per document; no band rect is re-placed.
- [x] `folio-designer/src/App.css` — `.band-boundary-handle` as a transparent hit strip on the
      boundary (`position:absolute; top:-3px; height:7px; left: calc(-1 * var(--band-x) - 104px);
      width: calc(var(--page-display-width) + 118px); background:transparent; border:0; padding:0;
      cursor: ns-resize; touch-action: none`), plus `.band-boundary-proposal` (the dashed rule at the
      proposed offset) and `.band-boundary-readout` (mono, `--color-select`, `pointer-events:none`).
      **No `@media`.** The 104px is `DESIGN.md`'s `band-tab.offsetFromPage`, so the strip reaches the
      tab the AC names. — UX-DR25's "hit targets larger than their visual footprint"; the affordance
      distinction AC2 requires (PAGE HEADER gets no strip and no `ns-resize`).
- [x] `folio-designer/src/App.test.tsx` — cover Matrix rows 1–5, 8–16: the two directions and the two
      bands; the PAGE HEADER band has no handle; the zero-delta gesture sends **nothing**; the readout
      shows points; the refusal's rendered text appears in the canvas `role="alert"` (closing the
      measured gap); arrow keys resize the focused boundary and do **not** move a selected component;
      a 3-sheet stack has exactly two handles. **Put these in a new describe, not the one whose
      self-count map ends at TWENTY-THREE.** — The Matrix's behavioural half.
- [x] `folio-designer/e2e/band-boundary-drag.spec.ts` (new) — a real-browser drag of both boundaries
      through the real Go worker: press, move, assert the proposed line and the points readout appear
      while **no** band rect has moved and no command has been sent, release, assert the band rect then
      moves and the Page Setup row shows the accepted height. Locate the handles by
      `getByRole('button', { name: 'Resize the page header', exact: true })` and the page-footer
      equivalent — **the names matter**: `application-shell.spec.ts` and `browser-native-roundtrip.spec.ts`
      already use `getByRole('region', { name: 'Page Header', exact: true })` under Playwright **strict
      mode**, so a handle named `Page Header` would break two shipped specs on a duplicate match.
      **The file must carry a header comment stating that this suite is executed by no workflow today
      and citing DW-193.** — jsdom cannot express pointer capture, `buttons === 0` recovery, or a real
      hit area; this is the only suite that can.
- [x] `folio-designer/src/canvas-authority-contract.test.ts` — no change expected; run it and confirm
      the new files clear every prohibition and the production floor still holds. **Note its `e2e`
      corpus floor is ≥ 15 and the new spec raises the count — the floor is a minimum, so it stays
      green, but the independent directory walk must see the new file.** — Read-only check.

**Acceptance Criteria:**
- Given a 3-sheet document, when the canvas is rendered, then exactly two band-boundary handles exist,
  both on the home sheet, and no accessible name collides with `Page Header`, `Content` or
  `Page Footer`.
- Given a boundary drag in flight, when the pointer moves, then no `.page-band` element's
  `--band-height` or `--band-y` changes and no command has been sent.
- Given the designer at this commit, when `folio-designer/src/` is searched for band-height arithmetic,
  then every bound it holds is read by `engine-bounds-mirror.test.ts` against the Go declaration it
  mirrors — there is no unmirrored copy.
- Given a document whose band heights are never edited, when it is opened and serialized, then the
  bytes are identical and `content` still carries no `height` key.
- Given a boundary released on the height already in force, when the gesture ends, then no command is
  sent and the undo history is unchanged.
- Given a band-height command refused by the engine, when the refusal reaches the canvas, then the
  engine's own located sentence is **rendered** in the canvas `role="alert"` and a test asserts that
  rendered text — the first such assertion in the repository. This story's whole safety argument is
  that a refusal reaches the author legibly, and nothing currently proves any canvas-gesture refusal
  is ever displayed at all.
- Given the new e2e spec, when the commit boundary is reported, then it is reported as
  **written-not-run**, separately from the gates that executed — it is not coverage until DW-193 is
  fixed and something runs it.

## Design Notes

**R1, measured — is a drag limit the thing Story 17.4 forbade? No, and here is the measurement.**

17.4 item 9, verbatim: *"**The floor and the ceiling are not the same kind of bound, which is why one
is mirrored and the other must not be.** Zero and negative are values these fields can NEVER legally
hold — whatever document is open, wherever the component sits. That fact is a property of the FIELD…
A band-edge ceiling is a property of the LAYOUT… For the browser to clamp there, the browser would
have to compute where the content band ends — geometry the engine owns and projects. That is the same
authority boundary **AD-17** draws for text, and a second, quietly-drifting copy of the engine's
containment rule does not belong in the inspector to save an author one error message."*

Three things that summary loses, and all three matter:

1. **It is about the INSPECTOR.** Its closing clause says so, and its asserted property is
   *"consistency with typing"*. There is no typing to be consistent with on a canvas boundary.
2. **The repo already ships a browser-side LAYOUT clamp — on a pointer drag.** `resize-anchor.ts`'s
   `proposedBounds` clamps a component drag to `limit.height` for the capping bands. Its own comment
   states 12.5's AC5 rationale in almost the same words: *"What the edges below do own is where a
   gesture runs out. A pointer that leaves the band should leave the component against that edge…
   not hand Go a rectangle it can only reject — a rejected drag lands the component back where it
   started, which reads as the drag having been thrown away."* And 17.4 itself preserved this split:
   its arrow keys send and are refused, while `proposedBounds` clamps. **The line this codebase
   actually draws is gesture vs. field, not field-property vs. layout-property.**
3. **The "quietly drifting" objection has a named answer here.** DW-36 permitted the drag clamp on one
   condition — that it *reads* the engine's own list rather than re-spelling it, and that
   `engine-bounds-mirror.test.ts` reads it doing so. That test line-anchors the clamp and red-proofs
   an inline restatement. A copy inside that census is not a quiet one.

So the honest decomposition of AC5 is **three bounds, not one**:

| Bound | Kind | RULED (R1) |
|---|---|---|
| Floor at `0` (negative height) | property of the FIELD — a band height is never negative in any document | **CLAMP.** 17.4 authorises it. Uncontroversial. |
| Ceiling at `innerH − other − 1` | property of the LAYOUT | **CLAMP, mirrored.** Every input is projected (`bands[0].height + bands[1].height`, by `isCanvas`'s contiguity invariant); the only thing not projected is the `− 1`, i.e. `bandContentWindowCeiling`'s strict positivity, which joins the mirror census as one new pair. |
| Floor at the lowest occupied edge (the strand) | property of the LAYOUT | **DO NOT BUILD.** 12.1's Q4 named this floor explicitly, said it *"**is** computable from the projection alone — and is deliberately **not** built"*, and recorded the omission as ruled. |

**The discriminator, and it is the reusable part — carry it forward.** Gesture-versus-field is not
quite the rule; this is:

> **Clamp a gesture at bounds that carry no information; send, and let the engine refuse, where the
> refusal names something the author needs.**

That is why one acceptance criterion produces three different answers. A ceiling that says only "no
further" carries nothing a stopped pointer does not already communicate, so clamping loses the author
nothing. The **strand** refusal carries an `ElementID` — it tells the author *which component is in the
way* — and a silent drag limit would destroy exactly that. The information content of the refusal, not
the taxonomy of the bound, is what decides.

**What building the ceiling costs, precisely:** it narrows 12.1's AC4 (*"there is none"*), which is
scoped *"at this commit"* and — measured — is enforced by **no test**, only by
`band-height-command.ts`'s header comment. So the cost is one amended comment plus one new mirrored
pair.

**The amendment is a NARROWING, not a deletion, and it must carry its reason.** Scope the ruling to
**the panel**, and state **why the panel keeps no bound while the gesture gets one**: the panel field
is typed, so 17.4's asserted property — *consistency with typing* — has something to be consistent
with, and an unmirrored clamp there would make the arrow key behave unlike the keystroke beside it. A
canvas boundary has no typed counterpart, so that property is vacuous there, and DW-36's census
answers the *quietly-drifting* half. **Without that sentence the next reader finds an asymmetry with no
reason attached and "fixes" it in whichever direction they happen to prefer.** That sentence is the
whole value of the amendment.

**The mirror tie is confirmed, not assumed.** The lead flagged a risk that the TS clamp's
`BANDS_CAPPING_VERTICALLY` and Go's capping-band set might be two independently maintained lists that
merely agree today. Measured at this plan gate: they are **already tied**.
`engine-bounds-mirror.test.ts` extracts Go's `^var bandsCappingVertically = \[\]string\{…\}$` from
`page_setup.go` and TS's `BANDS_CAPPING_VERTICALLY` from `engine-protocol.ts`, asserts
`toEqual` between them, asserts neither contains `content`, red-proofs a one-sided edit on **each**
side, and separately pins that `resize-anchor.ts` *imports* the list rather than restating it —
red-proofing the inline restatement too. `CAPPING_BANDS` is `toBe`-identical to it, and `CappingBand`
is derived from the projection union by `Exclude`. Go's `setBandHeight` gates on that same
`bandsCappingVertically`. So the chain Go → TS → clamp is closed, and the new ceiling pair **joins an
existing census rather than founding one**. Re-verify before writing the pair; if any link has moved,
tie it in the same commit.

**R2, measured — AC7 was already discharged, by 12.1, in the opposite direction.** AC7 and 12.1's own
epic text both said a shortened band clips its content with FR44's diagnostic. 12.1 shipped the
opposite under D-12.B: the writer refuses, evaluating **every** element. FR44 has no vertical axis
(D-2.8.1) and no vertical clip code exists. So AC7 as written would have required minting a diagnostic
code, adding a render behaviour, and lifting `isCanvas`'s band clause — three things 12.1 fenced. AC7
is recorded as satisfied by the refusal (Matrix row 10) rather than building a second overflow rule
the epic's own context says must not exist. **The planning text now agrees**: both instances are
struck in `epics.md` at HEAD `f6416c3`, with the measurement beside them.

**The instructive half of that sweep is what it did NOT strike.** Story 7.3's clip-and-warn survives
untouched, because it concerns a justified line exceeding its declared **width** — the one axis where
`TEXT_CLIPPED_WIDTH` actually exists. The premise was false only where it was applied *vertically*. A
sweep that struck every mention of FR44 clipping would have destroyed a true criterion alongside the
two wrong ones. **When a false premise is found, sweep for the premise, not for the phrase.**

Note DW-205 is priced LOW *"while it requires a hand-edited file, to be re-priced if any route can
strand a component again"* — a drag cannot strand, because the engine refuses and R1 declined the
browser-side strand floor, so the pricing survives unchanged.

**R3 — why a fifth wire field and not browser-side rounding.** Every other geometry factory
(`create`, `drop`, `move`, `resize`, `setComponentBounds`, `duplicate`) already carries `snap`, and
`SnapNearest`'s exact-halves-away-from-zero rule exists in exactly one place. Rounding in the browser
would be the first grid arithmetic in `folio-designer` and a fourth spelling of a rule already stated
once.

**And say "byte-preserved" precisely, because 12.1 carries a byte-identity AC and the phrase is
ambiguous.** The **command payload DOES change** — it gains `"snap":false`, and
`band-height-command.test.ts`'s byte-exact expectations move with it, deliberately. What is preserved
is the **document bytes** the panel's typed path produces: passing `snap: false` means an author who
types 83 still gets 83, so no shipped control changes behaviour and no golden moves. Written the wrong
way round, a reviewer reads a byte-identity claim, sees the payload move, and files a defect against a
story that did exactly what it said.

**Why the proposal paints a line and not moved bands — RULED 2026-09-05, and the reason generalises.**
AD-24 reserves band placement to `internal/layout`. Re-placing three rects in the browser for the
duration of a drag would be exactly that computation. AC3 asks only that *"the canvas paints the
proposed boundary"* — one line and one readout, at `original ± Δ` over a single projected number.
The coordinator's formulation is the rule the whole story turns on: **the browser may show where a
gesture has got to; it may not compute what the document would become.** Everything else waits for the
engine's re-projection at release, which is also what makes the gesture honest — what you see after
release is what the engine accepted, including the snap.

**"Drags the CONTENT tab" is read deliberately, not literally — RULED 2026-09-05.** Measured, the tab
is a class-less `<span>` hanging 104px outside the page's left edge whose **top** merely happens to be
flush with the boundary, and the boundary rule itself is a `pointer-events: none` pseudo-element. The
handle specified above is a transparent 7px strip **on the boundary**, extended left to `−104px` so it
spans the tab's horizontal position as well as the rule; a press on the line, at or beside the tab,
starts the drag, and the tab stays a label.

**What the literal reading would have meant, and why it is worse.** Making the `<span>` itself the
draggable control gives the author a *label* that resizes the page when grabbed — an affordance
attached to the wrong object, since the label names a band while the gesture moves the edge between
two of them. It would also force the tab to become a `<button>`, whose accessible name would then
either be `CONTENT` (which says nothing about resizing, and collides under Playwright strict mode with
the band region already named `Content`) or an `aria-label` that overrides its own visible text.
Dragging the boundary and labelling it as the boundary avoids all three. The AC was written before
anyone measured where the tab actually sits; this reading serves its intent.

## Verification

**Cadence (Q4) — RULED 2026-09-05.** Epic 12's recorded cadence is end-of-epic, and 12.5 is the last
story, so the boundary gate falls immediately after this commit.

- **Matrix — boundary gate, not this story.** 12.5 adds no render path and changes no laid-out
  geometry. Cross-target byte identity cannot be moved by a browser gesture. `-tags=matrix` also needs
  a Docker daemon with buildx for both Linux legs and Node for `js/wasm`, and **hard-fails rather than
  skipping** when a prerequisite is missing — a second reason not to hang it off a UI story.
- **Playwright — NOT run in this story.** Deferred to the boundary gate, where a change can be
  attributed against a known baseline. `npm run test:e2e` is **not** in `npm test` and **no workflow
  runs it** (DW-193, open); it needs a full production build plus
  `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH`; it last measured **32 pass / 1 fail**, the failure being
  `browser-native-roundtrip.spec.ts` › *fresh authored sessions close exactly through admitted Preview
  and native Folio*, which **hangs** — measured at both 300s and 1500s budgets, ~27 min wall, 9% CPU.
  It waits; it is not slow. Thirty minutes to reach a 32/33 that inherits a standing red is a toll,
  not a gate, and a green would not mean what it appears to.
- **But this story WRITES the e2e spec** (`e2e/band-boundary-drag.spec.ts`). DW-193 is about the
  runner, not the spec: an unrun spec still pins intent, costs nothing to hold, and is nearly free to
  write now versus archaeology in three months. **It must be reported at the commit boundary as
  written-not-run, never counted as coverage**, and its own header must say so and cite DW-193 — an
  unrun test that looks run is precisely the false-green this project keeps finding.

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected: rc 1, **exactly two** failures,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child (mandated permanent red). A
  **third failure is a hard stop** — report before triaging.
- `cd folio-go && gofmt -l .` — expected: empty.
- `cd folio-go && go vet ./...` — expected: rc 0, empty.
- `cd folio-designer && npx vitest run` — expected: rc 0, floor **62 files / 893 tests**, rising by
  this story's additions.
- `cd folio-designer && npx tsc -b --force` — expected: rc 0. (`npx tsc --noEmit` checks **zero** files
  here — DW-207. Never cite it.)
- `cd folio-designer && npx tsc -p tsconfig.e2e.json --noEmit` — expected: rc 0. (`tsc -b` does **not**
  cover `e2e/` — DW-212.) **This is the ONLY gate that will look at the new e2e spec at all**, and it
  checks that it compiles, not that it passes. Say exactly that when reporting.
- `cd folio-designer && npx oxlint` — expected: rc 0, exactly 4 `only-export-components` warnings.
  `App.tsx` changes in this story, so **prove the SET is the same four, not that the count is four**:
  the stable identity is `src/preview/pdf-viewer.tsx` × 2 and `src/App.tsx` × 2, same two exports;
  show byte-identity from the first `App.tsx` warning site to EOF and that every diff hunk sits above
  it.
- `cd folio-designer && npm run scan:font-hosts` — expected: 0 occurrences in **635** source files.
- `cd folio-designer && npm run scan:host-fonts` — expected: 0 occurrences, **149** files.
- `cd lint && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...` — expected:
  rc 0 / rc 0 / empty / rc 0. `-count=1` is mandatory: `internal/rules` walks the tree with `ReadDir`,
  which the test cache does not track.
- `cd hashmatrix && go vet ./... && gofmt -l . && go test -count=1 ./...` — expected: rc 0 / empty /
  rc 0. (A third Go module CI also gates.)

**Manual checks:**
- Mutate each new guard before submitting it: delete the floor clamp, delete the ceiling clamp, delete
  the send-only-if-changed condition, delete the `stopPropagation` that keeps the window nudge out —
  and show each mutation reds a named test, and **where** it reds.
- Confirm `git status --porcelain` shows only files this story names.
</content>

## Suggested Review Order

**The authority boundary — read this first**

- Where the gesture's limits live: floor 0, caller-supplied ceiling, footer sign flip, integer millipoints.
  [`band-boundary.ts:73`](../../folio-designer/src/band-boundary.ts#L73)

- The ceiling read off projected band heights alone — the one bound R1 narrowed 12.1-Q4 to allow.
  [`band-boundary.ts:38`](../../folio-designer/src/band-boundary.ts#L38)

- The only input to Go's ceiling that is not projected; it exists to be mirrored, not to be spelled twice.
  [`engine-protocol.ts:161`](../../folio-designer/src/engine-protocol.ts#L161)

- The mirror census that makes the copy non-quiet: Go's numeral, the TS numeral, the consuming site, red-proofed.
  [`engine-bounds-mirror.test.ts:321`](../../folio-designer/src/engine-bounds-mirror.test.ts#L321)

- The ruling narrowed rather than deleted, with the reason the panel keeps no bound attached.
  [`band-height-command.ts:71`](../../folio-designer/src/band-height-command.ts#L71)

**The engine arm**

- Arity 4 → 5, and `SnapToGrid` applied before every check so a refusal names what would be written.
  [`component_commands.go:2223`](../../folio-go/component_commands.go#L2223)

- A malformed `snap` is now located at its own field, not at the height's (D-000.25).
  [`component_commands.go:2176`](../../folio-go/component_commands.go#L2176)

- The ordering proof: 80pt accepted unsnapped, stranded at 78 snapped, refusal names 78.
  [`band_height_command_test.go:783`](../../folio-go/band_height_command_test.go#L783)

**The gesture**

- One conversion seam shared by move and release, so both spell the proposal identically.
  [`App.tsx:815`](../../folio-designer/src/App.tsx#L815)

- The only place a boundary gesture becomes a command; send-only-if-changed keeps no-ops out of history.
  [`App.tsx:773`](../../folio-designer/src/App.tsx#L773)

- 17.5's four pointer guards, plus the `placing` gate review added.
  [`App.tsx:782`](../../folio-designer/src/App.tsx#L782)

- Release recomputes from the pointer's own coordinate rather than the last move's.
  [`App.tsx:832`](../../folio-designer/src/App.tsx#L832)

- The focused handle owns its keys; modified keys and Tab still reach the window handler.
  [`App.tsx:855`](../../folio-designer/src/App.tsx#L855)

- One clear point for a live drag, reached by Escape and by the canvas-wide clear.
  [`App.tsx:148`](../../folio-designer/src/App.tsx#L148)

- Handle on the home sheet only; the proposal and readout are `div`s — see the CSS note below.
  [`App.tsx:1735`](../../folio-designer/src/App.tsx#L1735)

**The paint, and the defect that nearly shipped**

- The hit strip: transparent, 7px, reaching the tab at DESIGN.md's own 104px offset.
  [`App.css:123`](../../folio-designer/src/App.css#L123)

- The proposed line. As `<span>` this lost to `.page-band > span` (0,1,1) and never moved.
  [`App.css:129`](../../folio-designer/src/App.css#L129)

- The readout shows the PROPOSAL; the engine's accepted value replaces it on re-projection.
  [`App.css:138`](../../folio-designer/src/App.css#L138)

**Tests and the unrun proof**

- The story's 18+ behavioural rows, including the zoom row nothing else pins.
  [`App.test.tsx:5177`](../../folio-designer/src/App.test.tsx#L5177)

- The pure arithmetic, both bands, both directions, each clamp red-proofed.
  [`band-boundary.test.ts:1`](../../folio-designer/src/band-boundary.test.ts#L1)

- WRITTEN, NEVER RUN: no workflow executes this suite (DW-193). Compiles only.
  [`band-boundary-drag.spec.ts:38`](../../folio-designer/e2e/band-boundary-drag.spec.ts#L38)

## Delivery Log

### 2026-09-06 — done

Baseline `cfa4263`. Shipped in **`dff158d`** on `main` — 16 files, +2565 / −89 — which was already
pushed before this close, so nothing here was folded into it. This closer touched only the story file
(opener + this log) and `sprint-status.yaml`, and staged nothing.

**The rule this story settled, and it is the reusable half.** *Clamp a gesture at bounds that carry no
information; send, and let the engine refuse, where the refusal names something the author needs.* That
single sentence is why one acceptance criterion produced three different answers. The floor at zero and
the content-window ceiling say only *no further*, which a stopped pointer already communicates, so
clamping there costs the author nothing. **The strand floor is deliberately NOT clamped**: its refusal
is the only one in this feature that carries an `ElementID` — it names *which component is in the way*,
and at what depth it reaches — and a silent drag limit would destroy exactly that. The information
content of the refusal, not the taxonomy of the bound, is what decides. Gesture-versus-field was the
near-miss formulation; this is the one that survives.

**D-12.B's GROUND was corrected, not its verdict.** The lead corrected its own earlier sentence — *"the
objection is the browser computing layout geometry at all"* — as over-broad, and it would have blocked
correct work. What stands unchanged is D-12.B's **conclusion**: a typed band-height panel field still
gets **no floor**. 17.4 forbids an unmirrored clamp in a control that has a typed counterpart; it does
not forbid a mirrored clamp on a gesture that has none. The narrowing landed as an amended header
comment on the encoder — scoped to the panel, never deleted — carrying the reason attached, so the next
reader meets the asymmetry with its justification rather than "fixing" it in whichever direction they
prefer.

**The fifth `snap` field, and "byte-preserved" said precisely.** `setBandHeight` went from arity 4 to 5
with the 22 payload literals and the strict-arity gate in one commit, per R3's first condition. The
**command payload DOES change** — it gains `"snap":false` — and the byte-exact wire expectations moved
with it, deliberately. What is preserved is the **DOCUMENT bytes the panel's typed path produces**: the
panel passes `false`, so an author who types 83 still gets 83, no shipped control changes behaviour and
no golden moves. Written the other way round, a reviewer sees the payload move against a byte-identity
claim and files a defect against a story that did exactly what it said.

**A file the task list never named, disclosed rather than smuggled.** `engine-protocol.ts` was edited —
one new mirrored constant, the content-window ceiling's strict-positivity margin — and it appears in no
task line. It belongs there and nowhere else: that module is the declared **authority** for every
mirrored engine bound in the designer (the capping-band list, the locale tags, the line-spacing and
payload limits all live there), and the mirror census explicitly exempts it from the "no file hand-copies
a bound" walk that every other source file is held to. Spelling the margin in the consuming module
instead would have put a bare literal outside the census — the precise thing DW-36's standing condition
forbids. The consuming site reads the constant by name and the census red-proofs a one-sided edit on
each side.

**Three review findings worth carrying forward, all invisible to every executing gate.**

- **P1 — the proposal line was visually inert and would have shipped that way.** The proposed boundary
  and its readout were direct `<span>` children of the band, and the band-tab rule `.page-band > span`
  at specificity **(0,1,1)** beats a class rule's **(0,1,0)**. `top: 0` would have won over the proposed
  offset; the elements would have painted at the band's top-left, translated a tab's width off the page,
  wearing the tab's border, tint and `text-transform: uppercase`. The line would never have tracked the
  pointer in a real browser — **with all 938 tests green, because jsdom applies no stylesheet.** They are
  `<div>`s now, and the replacement guard's shape is the durable part: it **gathers every `.page-band > …`
  selector out of the stylesheet** rather than listing them, asserts neither painted element matches any
  of them, and uses **the band tab as a positive control** — the same gathered selectors must reach the
  tab, so "does not match" is a measurement rather than a regex that matched nothing. A child rule added
  later is covered without anyone remembering to come back.
- **P2 — at any zoom but 1 the delta was not integral.** Every other row runs at zoom 1, where the
  pixel-to-millipoint conversion is a whole number by luck. Measured at zoom 1.1: a 36px drag on a 60pt
  header yields `27273.000000000004`, which the points formatter spells `"27.273.00000000000364"` — a
  string with two decimal points. `JSON_NUMBER` rejects it, so the command went out as `"height":null`
  while the author read the malformed string in the readout for the whole gesture. **The zoom-ignoring
  mutation left all 929 tests green**, which is why the row asserts the zoom's own arithmetic (the same
  drag must not read `24`) and not merely that the readout is well-formed.
- **P10 — a self-referential guard anchored on bare literals.** The 12.5 block was first spliced between
  Story 17.1's header line and 17.1's body, so 17.1's documentation read as if it described 12.5, and
  17.1's self-count guard — which counts from its describe to EOF — passed the whole time. Writing the
  new placement guard with plain literals made it worse: its own quoted markers occur **earlier in the
  file than the real ones**, so it sliced its own text and passed on nothing, *and* it dragged 17.1's
  self-count guard onto this block (measured: that guard went red with `expected undefined to be
  defined`). Every marker is now matched line-anchored as a regex, which excludes both, because a marker
  quoted inside an expect is indented and a real one sits at column 0.

**The e2e spec is WRITTEN, NEVER RUN.** `e2e/band-boundary-drag.spec.ts` exists, compiles, and pins the
gesture's intent. **It is not coverage.** No workflow executes that suite, it is not in `npm test`, and
the only gate that looked at it proves it typechecks. Its own header says so and cites **DW-193**, which
is about the runner, not the spec.

**Measured gates at this close, re-run against the shipped tree.**

- `folio-go`: `go test -count=1 ./...` **rc 1**, **exactly two** `--- FAIL:` lines —
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child (mandated permanent red; P6g got
  7, needs ≥ 20). **No third.** `gofmt -l .` empty **rc 0**; `go vet ./...` **rc 0**, silent.
- `folio-designer`: `npx vitest run` **rc 0**, **63 files / 938 tests passed**. `npx tsc -b` **rc 0** and
  `npx tsc -b --force` **rc 0** (`tsc --noEmit` typechecks zero files here — DW-207 — and is not cited).
  `npx tsc -p tsconfig.e2e.json --noEmit` **rc 0**: this is the only gate that looks at the new e2e spec
  at all, and it proves the file **compiles**, not that it passes.
- `npx oxlint` **rc 0**, exactly **4** `only-export-components` warnings. **The SET is proved by symbol,
  not counted**: `App.tsx`'s exported-symbol list is byte-identical between `cfa4263` and `dff158d`
  (five `^export ` declarations, `diff` rc 0), the two warned symbols are the same two non-component
  exports as before, and `pdf-viewer.tsx` is not in the commit at all. *Deviation worth recording:* the
  Verification section's alternative method — *"byte-identity from the first `App.tsx` warning site to
  EOF, every diff hunk above it"* — **would not have held**. A hunk lands at `+3256`, below the first
  warning site at line 3240. The symbol-set diff is the method that actually discriminates here.
- `npm run scan:font-hosts` **rc 0**: 0 occurrences in **638** source files (**638 tracked + 0
  untracked**, floor 400). `npm run scan:host-fonts` **rc 0**: 0 occurrences of 4 spellings across
  **152** files (**152 tracked + 0 untracked**, floor 86). Both rose by exactly 3 from the spec's 635 /
  149 — this story's three new files, now committed, so the whole rise sits on the tracked side of the
  split. The totals moved; the split did not repartition.
- `lint`: `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test -count=1 ./...` — **rc 0 / rc 0 /
  empty / rc 0**, four `ok` lines. `hashmatrix`: `go vet ./...` **rc 0**, `gofmt -l .` empty,
  `go test -count=1 ./...` **rc 0**.

**Suites NOT run, and when they come due.** The **cross-target matrix** and **Playwright** were out of
scope by this epic's heavy-test cadence, which is **end of epic**. Epic 12 is now five stories done, so
they fall due **immediately**, at the **Epic 12 boundary gate**, which the orchestrator runs next — they
are owed for all five stories, not for 12.5 alone. Playwright last measured **32 pass / 1 fail**, the
failure being `browser-native-roundtrip.spec.ts` › *fresh authored sessions close exactly through
admitted Preview and native Folio*, which **hangs** (measured at both 300 s and 1500 s budgets, ~27 min
wall at 9% CPU — it waits, it is not slow). That red is carried in, not this story's.

**Triage: no tally exists in the record, and this closer did not invent one.** There is no
`## Implementation Notes` or triage section in this file, and no count of patched / deferred / rejected
was written anywhere the build loop reached — so a later reader should read that as an **absence**, not
as an omission here. What is measurable: **seven** review findings left a numbered marker in the shipped
code — P1, P2, P3, P5, P8, P10, P12 — and the highest marker is **P12**, so the review population was at
least twelve. That is a floor on the population, not a triage split.

**Deferred, four entries, all filed by the builder inside `dff158d` and all classified PRE-EXISTING.**

- **DW-226** (MEDIUM, open, unassigned) — `.page-seam` is out-specificitied by the same band-tab rule
  P1 hit, and has never rendered where it is placed. Story 7.6's element; 12.5's cure did not transfer
  because `.page-band > span` is byte-unchanged.
- **DW-227** (MEDIUM, open, **needs a ruling**) — an arrow-key nudge is inert whenever snapping is on,
  for components and band boundaries alike, and Shift+Arrow moves 12pt rather than the documented 10.
  Two defensible fixes; picking one quietly inside a story would change a shipped component's behaviour.
  **Owner: engineering lead.**
- **DW-228** (LOW, open, unassigned) — `isCanvas` pins band contiguity but never anchors the stack to
  the printable column, so the ceiling's sum is correct by construction rather than by the check the
  comment cites.
- **DW-229** (LOW, open, **needs a ruling**) — canvas gestures expose no accessible value and their
  feedback is `aria-hidden`. Names the concrete shape `EXPERIENCE.md`'s standing UX1 gap now takes for a
  value-bearing gesture. **Owner: engineering lead**, with the owner behind them.

**Housekeeping at this close.** Frozen block **byte-identical**, proved by `sha256` before and after
(`10db520dadf262f90c953f8e41d85b0c02231f5e3e6e58bad4a37009e2a2ec6e`, with a one-line-short window as
negative control). Tracker: `12-5…` moved `review` → `done`; **`epic-12` deliberately left
`in-progress`** pending its boundary gate. Output tree swept: nothing moved, nothing removed, no
untracked files under `_bmad-output`, and every shape-suspicious artifact checked individually and found
referenced. All six context caches carry the right header and are non-empty; the only planning artifact
newer than any of them is `epics.md`, whose last change (`f6416c3`) edited **only** the Epic 12 section,
and `epic-12-context.md` was recompiled after it inside `dff158d`.
