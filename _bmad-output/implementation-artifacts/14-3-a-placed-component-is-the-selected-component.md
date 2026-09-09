---
title: 'Story 14.3: A placed component is the selected component'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: 'af2cfbd5b2245355959e406d5cd25f819896ea9a'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Placing a component from the palette leaves nothing selected — `place`/`placeInBand`
(`App.tsx:901-921`) commit through `commitComponent` and never call `select`, so the inspector stays on
PAGE SETUP showing "Component properties require a selection." The author must then find and click the
thing they just made, and for a Line that target is 2px tall. Placing a component and editing it should be
one gesture, not a hunt.

**Approach:** Select and focus the newly placed component by diffing the component ids across the commit
that created it, and give thin components a hit region that is padded in chrome only — CSS on the existing
`.canvas-component` element, driven by a projection-derived custom property. Nothing about what is drawn,
what is stored, or what is sent to the engine changes.

## Boundaries & Constraints

**Always:**

- **The padding is chrome and is never serialized.** It exists only as CSS on the canvas element. No
  command carries it, no `FieldSpec` authors it, and no document byte changes because of it. *D-12.4.1
  rules "No `FieldSpec` may author a padding field" — that ruling is about `style.padding`, a **document**
  property the engine renders. AC2's "padded" is a hit region in the editor's chrome. These are different
  things that share a word, and this spec says so to pre-empt the collision.*
- **The comfortable hit size is 12px**, on the thin axis, matching `.selection-handle` (`App.css:250`, a
  12×12 transparent box over a 6×6 painted mark). **Owner's ruling.** The other in-repo precedent,
  `.resize-handle` at 24px (`:248`), is a *corner* grab of which there is one per component and which can
  therefore afford to be generous. An edge-to-edge pad of ~11px per side on every thin component would make
  overlap the common case rather than the edge case — and overlap is exactly what AC3 must keep
  deterministic and what the armed-placement suppression must keep from eating placements. 12px is still 6×
  today's reachable 2px while staying proportionate to what it pads.

- **The pre-existing 2px floor at `App.css:197` and `:242` is left alone. RULED at the plan gate (Q1).**
  Be clear about what that floor *is*, because the tempting description of it is wrong: **it is itself hit
  padding, implemented by inflating the box.** It sits on `.canvas-component` — the element carrying
  `cursor: move` and the pointer events — and `App.css:238-241` states its reason as a 1pt rule being too
  small a *target*. The paint only follows because `.canvas-box { inset: 0 }` (`:237`) fills whatever box it
  is handed. So it is not a different mechanism from this story's; it is the same mechanism built the way
  AC2's next clause forbids — *"padded independently of the drawn thickness"*. **On a plain reading, AC2 is
  a description of what is wrong with the shipped floor.** Do not defend leaving it by arguing about which
  padding the clause governs; that argument does not survive contact with the CSS. It is left alone on three
  grounds that do not depend on parsing the clause:
  1. **The canvas is architecturally the approximate representation.** D-13.4.1: *"FR34 is the Design Canvas,
     the approximate representation"*, with the exactness promise living on the PDF and the Preview surface.
     A paint floor there is inside that surface's remit.
  2. **Blast radius, measured.** The floor is declared twice — `.canvas-component` (`:197`) floors **both
     axes for every component type**, and `.canvas-component-line` (`:242`) re-declares the height floor. So
     removing or lowering it changes what every text box, image, table and rule draws at every zoom below 2,
     inside a story about which component is selected.
  3. **It is a considered decision with a stated reason** (`:238-241`), so overturning it is a
     visual-fidelity product call, not something a selection story folds in.
  **What the floor actually costs, and why a 1px floor was rejected on arithmetic rather than taste:** the
  floor does not merely inflate, **it collapses distinct document states into one appearance**. At zoom 1,
  1pt and 2pt both draw 2px. At zoom 0.5, 1pt, 2pt *and* 3pt all draw 2px. The author cannot see a
  difference the PDF will show. A 1px floor would fix zoom 1, leave zoom 0.5 collapsed, and change what
  every user sees — retiring the defect's visibility without retiring the defect.
  **This is registered as DW-345**, which carries the collapse arithmetic above. 14.3 re-owns it rather than
  absorbing it. **D-000.9 item 2's "both halves need reconciling with shipped CSS" is discharged not by this
  paragraph but by an executable assertion** — see the `.canvas-box` byte-identity task under Tasks &
  Acceptance. Prose about the distinction cannot fail; that test can, and it is what makes leaving the floor
  alone honest rather than a deferral wearing a paragraph.
- **AC2 must be CSS geometry, never a computed hit test.** `canvas-authority-contract.test.ts` bans
  `getBoundingClientRect`, `getClientRects`, `offset[XY]`, `client(Width|Height|Left|Top)`, `getComputedStyle`,
  `ResizeObserver` and `devicePixelRatio` across `src/**`, `src/**.test.*` and `e2e/**`. Deriving the pad
  from the **projection** (`component.width`/`height`, already in hand) is arithmetic on document values and
  is permitted; asking the DOM how big anything is, is not.
- **Do not disturb the `placementPoint` seam.** `canvas-authority-contract.test.ts:908-909` asserts the
  source-text seam `/export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/`
  **is present**, and `App.tsx:4314-4317` matches it byte-for-byte. Do not rename `placementPoint`, do not
  change its parameter type, and insert **no function** between it and `pageStyle`. Breaking this makes the
  contract test **throw**, not merely fail.
- Selection stays local. Selecting or focusing a component sends no command and produces no engine traffic.

**Ask First:**

- Any change to what is painted on the canvas — including touching `max(2px, …)` at `App.css:197` or `:242`.
- Adding a new focusable control, or any nested `button`/`role="button"`, inside `.canvas-component`. It is
  already `role="button" tabIndex={0}` (`App.tsx:4349`) and therefore already inside
  `control-vocabulary-contract.test.tsx`'s population (`:177`); a nested control enters R1–R4.
- Any new numbered carve-out in `canvas-authority-contract.test.ts`. None is needed for this story.
- Any change to `component-command.ts`'s command shape, or to Go.

**Never:**

- Never change document bytes, the wire format, or any command payload.
- Never make the hit padding paint, tint, outline or otherwise become visible.
- Never override the natural DOM stacking order to resolve overlap **between two thin components' hit pads**
  — AC3's determinism is already free there, and imposing an order would replace a working rule with an
  invented one. **This does not extend to a pad overlapping another component's painted box:** a component's
  own paint always outranks a neighbour's invisible pad, and arbitrating that is required, not forbidden. The
  original clause was written about thin-vs-thin and over-reached into a case the matrix did not cover.
  **The rule to remember: a component's own paint outranks a neighbour's invisible pad.**
- Never take DW-337, DW-340 or DW-327. All three are deferred and re-owned off this story at the plan gate.
- Never fix `createComponentCommand`'s hardcoded `width: 72, height: 24` (`component-command.ts:33`).
  Registered by the orchestrator as **DW-344**.
- Never edit `sprint-status.yaml`, `deferred-work.md`, `epics.md` or `DESIGN.md`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Place selects | Palette armed, band clicked, engine returns a snapshot with one new id | That id becomes the sole selection; the inspector renders ComponentProperties, not PAGE SETUP; the string "Component properties require a selection." is absent | N/A |
| Place focuses | As above | The new component's element holds DOM focus and shows its `:focus-visible` outline | N/A |
| Placement refused | Engine rejects the create command | Selection is unchanged; nothing is focused; the existing `componentDiagnostic` error path is untouched | Existing `setCommitError` path |
| Snapshot adds no id | Command resolves but the id set is unchanged | Selection unchanged, no focus move, no throw | Degrade silently; never select a stale id |
| Thin line padded | A 1pt Line, zoom 1 | The pointer region reaches a 12px extent on the thin axis; `--component-height` and the painted `.canvas-box` are unchanged | N/A |
| Padding never paints | Any thin component | No pixel of the padded region is painted, tinted or outlined; the dotted `::after` outline stays on the drawn box | N/A |
| Paint is byte-identical | A 1pt Line, padding present vs. absent | The painted `.canvas-box` geometry is **identical** across both, while the pointer-reachable area **grows** | N/A |
| Thin non-line padded | A 1pt-tall Image or Rectangle, or any element under 12px on either axis | Padded on the same terms as a Line — the rule is per-element-size, never per-kind | N/A |
| Thick component unpadded | A component already at or over 12px on both axes | Pad is zero; its hit region is exactly its box, as today | N/A |
| Overlap is stable | Two thin components whose padded regions overlap, clicked repeatedly in the same spot | The same component — the last in document order — is selected every time; the selection never alternates | N/A |
| Paint outranks a neighbour's pad | A 1pt rule placed after (below in document order) a text box, their regions overlapping by ~5px; press the text box's bottom edge | The **text box** is selected and dragged, not the rule — a component's own paint always outranks a neighbour's invisible pad | N/A |
| The pad still wins over empty canvas | The same rule, pressed where its pad crosses no neighbour's paint | The **rule** is selected — the arbitration must not degrade into "the pad never wins anything" | N/A |
| Placement beats padding | Palette armed, click 4px from an existing 1pt rule, inside that rule's padded region | The component is **placed in the band**; the nearby rule is not selected and steals no pointerup | N/A |
| Padding restored | Placement disarmed after the above | The padded region takes pointer events again; clicking 4px from the rule selects it | N/A |

</frozen-after-approval>

## Code Map

**The placement path — AC1, AC4**
- `folio-designer/src/App.tsx:207` — `const [placing, setPlacing] = useState<PaletteKind>()`. The armed
  palette kind. **Also the gate for the AC3/placement conflict** (see `App.css` below).
- `folio-designer/src/App.tsx:53` — `paletteItems`, the closed five.
- `folio-designer/src/App.tsx:2299` — the palette rail. Buttons only *arm* a kind; they never place.
- `folio-designer/src/App.tsx:2172` — the band's `onPointerUp`/`onKeyDown`. Placement fires here, gated on
  `event.currentTarget === event.target`. **This gate is why an enlarged hit region can break placement.**
- `folio-designer/src/App.tsx:901-921` — `place` / `placeInBand`. Both call `clearInteraction()` then
  `commitComponent(...)`. **Neither selects. This is the gap AC1 names.**
- `folio-designer/src/App.tsx:813-822` — `commitComponent(payload, after?)`. Already takes an `after`
  callback (used by `deleteSelection`), but `after` receives no result, so the id diff must be taken around
  the `await` **inside** `commitComponent`, or the callback re-shaped to receive the new snapshot.
- `folio-designer/src/engine-client.ts:10` — `EngineResult`. **Read-only evidence: there is no created-id
  field.** A `command` response carries only `snapshot`. The new id must be *derived*, by diffing
  `snapshotRef.current?.canvas?.components` against `result.snapshot.canvas.components`. Exactly one id is
  added per create. Prefer the diff over "last in the band": the append position is a Go implementation
  detail, the diff is not.
- `folio-designer/src/App.tsx:922` — `select(id, extend)`. The one selection function. Reuse it; do not
  add a second path to `setSelected`.
- `folio-designer/src/App.tsx:1886-1890` — `returnWithOptionalSelection`, the **existing precedent** for
  select-then-focus: `setSelected([id])` followed by `setTimeout(() => canvasRegionRef.current?.focus(), 0)`.
  Copy the deferred-focus shape; AC4 targets the component element rather than the region.

**The inspector branch — AC1**
- `folio-designer/src/App.tsx:2345` — `selected.length > 0 && canvas ? <ComponentProperties …/> : <PageSetup …/>`.
  The branch AC1 flips.
- `folio-designer/src/App.tsx:2539` — `PageSetup`, first two nodes, carrying the exact string
  `Component properties require a selection.` (trailing full stop). AC1 names this string; assert its absence.

**The canvas element and its geometry — AC2, AC3**
- `folio-designer/src/App.tsx:4349` — `CanvasComponent`'s root `<div>`: `role="button" tabIndex={0}`,
  `style={componentStyle(active, zoom)}`, and the `onClick` that suppresses the duplicate via `selectedByPointer`.
- `folio-designer/src/App.tsx:4337` — `begin()`. **Selection is taken on `pointerdown`, not click**, and calls
  `event.stopPropagation()`. That `stopPropagation` is the mechanism that makes AC3's anti-behaviour impossible.
- `folio-designer/src/App.tsx:4330` — `selectedByPointer` ref. Do not "simplify" this.
- `folio-designer/src/App.tsx:4551` — `componentStyle()`, which emits `--component-x/y/width/height`.
  **This is where `--hit-pad-*` belongs**, computed from the projection.
- `folio-designer/src/App.tsx:4307-4310` — `canvasDisplay.css/documentDelta`. **1pt = 1 CSS px at zoom 1**
  (`millipoints * zoom / 1000`). Not the CSS standard's 1.333px — this codebase does its own arithmetic.
- `folio-designer/src/App.tsx:185` — `zoom`, `useState(1)`, clamped 0.5–2.0 in 0.1 steps.

**The stylesheet — AC2, AC3, and the placement conflict**
- `folio-designer/src/App.css:197` — `.canvas-component { width: max(2px, …); height: max(2px, …); overflow: visible }`.
  **The pre-existing paint floor. Do not touch.** `overflow: visible` is what lets a negative-inset
  pseudo-element escape the box.
- `folio-designer/src/App.css:237` — `.canvas-box { position: absolute; inset: 0; pointer-events: none }` —
  the painted fill. It fills the *floored* element, which is why the floor changes what is drawn.
- `folio-designer/src/App.css:242-243` — `.canvas-component-line`, a redundant repeat of the height floor,
  plus `::after { content: none }` which removes the dotted outline for lines.
- `folio-designer/src/App.css:198` and `:245` — `.canvas-component::after` is **already taken** (the dotted
  outline, `pointer-events: none`) and re-used by `.canvas-component-selected::after`. **`::before` is free
  on `.canvas-component`** — grep confirms no `.canvas-component::before` rule exists.
- `folio-designer/src/App.css:248-251` — `.resize-handle` (24×24 transparent box, 6×6 `::after` mark) and
  `.selection-handle` (12×12 box, `inset: 3px` mark). **This is the established idiom for "hit target larger
  than its visual footprint" in this very file** — `App.css:174` calls it "the shared idiom". Copy its shape.
- `folio-designer/src/App.css:131-133` — `.band-window { overflow: hidden; pointer-events: none }` and
  `.band-window > .canvas-component { pointer-events: auto }`. **The band window clips**, so a padded region
  is cut off at a band boundary. That is acceptable and should be stated, not worked around.
- `folio-designer/src/App.css:140-141` — `.canvas-component-echo { pointer-events: none }`. **The precedent
  for Q3's ruling**: echoes were made inert precisely because they "swallowed the placement pointerup the
  band beneath it was meant to receive". Same defect class, same fix.
- `folio-designer/src/App.css:247` — `.canvas-component:focus-visible { outline: 2px solid var(--color-select) }`.
  The affordance AC4 relies on; it already exists.
- Read-only: `design-contract.test.ts` bans raw `#hex`, `rgb(`, `hsl(` and non-`var()` `border-radius` in
  `App.css`. New CSS must use tokens.

**Guards this story must stay green against**
- `folio-designer/src/canvas-authority-contract.test.ts:18-147` — the 17 prohibitions; `:297` scans
  production + tests + `e2e/`; `:762-766` pins population floors with `toBeGreaterThanOrEqual`.
  Verified read-only: `clientX`, `clientY`, `elementFromPoint`, `pointerdown`, `onPointerDown`,
  `pointer-events` and `inset` return **zero hits** — pointer input and CSS geometry are unpoliced.
- `folio-designer/src/control-vocabulary-contract.test.tsx:177` — sweeps `button, [role="button"]`.
  `.canvas-component` is already in that population.

**Tests**
- `folio-designer/src/App.test.tsx:1170` — the palette/placement test. Its mocked response is `snapshot(2)`
  with **no new component in it**, so it cannot observe AC1 today; it needs a snapshot that actually adds one.
- `folio-designer/src/App.test.tsx:1303-1307` — `pointerSelect`, the shared helper. It resolves its target
  **by accessible name** and passes `clientX: 1, clientY: 1` only for drag deltas. **Read-only evidence that
  jsdom cannot prove AC2/AC3**: jsdom applies no stylesheet and computes no layout.
- `folio-designer/src/App.test.tsx:1313-1415` — Story 17.2's eight-row selection matrix. Extend, don't rewrite.
- `folio-designer/src/line-rect-vocabulary.test.tsx:89` — the Line/Rect mount helper.
- `folio-designer/e2e/component-manipulation.spec.ts:18` — **the model for AC2/AC3's real-coordinate proof.**
  It already drives `page.mouse.move/down/up` off `boundingBox()`. Playwright's `boundingBox()` does **not**
  match the contract's `getBoundingClientRect|getClientRects` ban, and this file is in the scanned corpus and
  green today.

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/App.tsx` -- in `commitComponent` (`:813`), capture the component id set from
      `snapshotRef.current` before `await engine.request(...)` and diff it against the resolved snapshot;
      expose the newly added id (if exactly one) to the caller. -- The protocol carries no created-id
      (`engine-client.ts:10`), so AC1 must derive it, and the diff is the derivation that does not depend on
      Go's append position.
- [x] `folio-designer/src/App.tsx` -- in `place` (`:901`) and `placeInBand` (`:916`), select the derived new
      id via the existing `select()` (`:922`) and move DOM focus to that component's element, deferred in the
      shape `returnWithOptionalSelection` already uses (`:1890`). If no single new id is derived, change
      nothing. -- AC1 and AC4. Reusing `select()` keeps one selection path.
- [x] `folio-designer/src/App.tsx` -- in `componentStyle` (`:4551`), emit a `--hit-pad-x`/`--hit-pad-y`
      custom property computed from the **projected** width/height and zoom: zero when the axis already
      reaches **12px**, otherwise half the shortfall to 12px. -- AC2. Arithmetic on projection values keeps
      this clear of every measurement ban; no DOM query is permitted.
- [x] `folio-designer/src/App.css` -- add a `.canvas-component::before` rule with `content: ''`,
      `background: transparent`, negative `inset` driven by `--hit-pad-*`, and no paint of any kind —
      **modelled on `.resize-handle` (`:248-249`), the established in-repo idiom: a 24px transparent target
      carrying a 6px painted mark, i.e. pointer surface and painted mark deliberately different sizes on a
      `background: transparent` box with negative insets.** That is already the shape AC2 describes. Add a
      companion rule making the pseudo-element `pointer-events: none` while a placement is armed, modelled on
      `.canvas-component-echo` (`:140-141`). Do **not** modify `:197`, `:237` or `:242`. -- AC2 and the
      placement conflict ruled at Q3.
      **Why `::before` and not `::after`, and do not "simplify" this back:** `::after` on `.canvas-component`
      carries the dotted outline (`:198`), is restyled for the selected state (`:245`), and is freed
      (`content: none`) **only for lines** (`:243`). AC2 covers *"a 1pt Line, **or any element under the
      comfortable hit size**"* — so padding hung on `::after` would work for a Line and **silently fail for
      every other thin element**, while passing any test written only against lines. `::before` is unused on
      `.canvas-component` (verified) and is free for every kind.
- [x] `folio-designer/src/App.test.tsx` -- **assert the painted box is byte-identical with the padding
      present and absent, while the pointer-reachable area grows.** Compare the `.canvas-box` geometry and
      the component's own `--component-width`/`--component-height` across both states and require exact
      equality, then show the reachable extent differs. -- **This is how D-000.9 item 2 is discharged: an
      assertion that can fail, not a paragraph.** A test that only asserts "a click near a line selects it"
      **passes on the wrong implementation** — the one that inflates the box — which is exactly how the 2px
      floor came to exist. Cover the thin-non-line matrix row here too, so the guard is not line-only.
- [x] `folio-designer/src/App.tsx` -- expose the armed-placement state to the stylesheet as a data attribute
      on an ancestor already in the tree (`placing`, `:207`), so the suppression rule above has a hook. --
      Q3's ruling; no new element and no new control.
- [x] `folio-designer/src/App.test.tsx` -- extend the placement test at `:1170` so its mocked snapshot
      actually adds a component, and assert: the new id is selected, ComponentProperties renders, the string
      "Component properties require a selection." is absent, the element holds focus, and **no command is
      sent by the selection or the focus**. Add the refusal and no-new-id rows from the I/O matrix. -- AC1,
      AC4 and the matrix rows jsdom can honestly observe.
- [x] `folio-designer/src/App.test.tsx` -- assert `componentStyle` emits the expected `--hit-pad-*` values for
      a thin component and **zero** for one already over the floor, reading the inline custom property the way
      `:1262` already reads `--component-x`. -- The arithmetic half of AC2, which jsdom *can* see.
- [x] `folio-designer/src/App.test.tsx` -- add the focus/selection discriminating test: placing sets both, and
      the document-level arrow-nudge handler does not also fire. -- D-12.A Fork 3: two handlers that happen to
      agree are one regression from both firing, and a happy-state-only assertion cannot tell the difference.
- [x] `folio-designer/e2e/` -- add a spec, modelled on `component-manipulation.spec.ts:18`, that places a
      1pt Line and clicks a real offset outside its drawn box but inside its padded region (AC2); clicks the
      same spot repeatedly over two overlapping thin components and asserts the *same* one is selected every
      time (AC3); and arms a placement, clicks 4px from an existing rule, and asserts the component is placed
      rather than the rule selected (the Q3 row). -- **The only honest witness for pointer geometry.** jsdom
      computes no layout, so these claims cannot be proved in the unit suite. Per D-000.33 this executes at
      the Epic 14 boundary gate, not in this story.

**Acceptance Criteria:**

- Given a component placed from the palette by pointer or by keyboard, when the engine accepts it, then that
  component is the sole selection and holds focus, and the inspector shows its properties rather than PAGE SETUP.
- Given the hit padding is active, when any document is saved or any command is sent, then the bytes are
  byte-identical to what the same authoring produced before this story — the padding reaches neither.
- Given a thin component, when its padded region is added, then `--component-width`/`--component-height` and
  the painted `.canvas-box` are unchanged, and no pixel of the padded region is painted.
- Given the padding is re-implemented as box inflation instead of a negative-inset pseudo-element, when the
  suite runs, then the `.canvas-box` byte-identity test **fails** — proving the guard separates this story's
  mechanism from the one DW-345 records rather than merely observing that clicking near a line works.
- Given a thin Image or Rectangle rather than a Line, when it is under 12px on either axis, then it is padded
  on the same terms — the padding is keyed to element size, never to component kind.
- Given the repeated-click stability test from AC3, when it is run against a build in which the padded region
  is reverted, then it still passes — proving it measures determinism rather than padding.
- Given `npx vitest run`, `npx tsc -b --force`, `npx oxlint` and `npm run test:e2e:compile`, when all four run,
  then all pass, and `canvas-authority-contract.test.ts` and `control-vocabulary-contract.test.tsx` are green
  without a new carve-out.

## Spec Change Log

### 2026-09-09 — frozen-block amendment, authored by the human coordinator (not by an agent)

**Triggering finding.** All three step-04 review layers converged on a real interaction regression: a thin
component's transparent `::before` pad hit-tests **above a neighbouring component's painted box** and its
resize handles. Verified at the tree, not relayed — `.canvas-component` is `position: absolute` with
`z-index: auto` and no `isolation`/`transform`/`opacity`, so it creates **no stacking context**; and there is
**zero** `z-index` on `.canvas-component`, `.resize-handle` or `.selection-handle`. A later sibling's pad
therefore paints and hit-tests over an earlier sibling's whole box. Ordinary case: a 1pt rule under a
heading steals the heading's bottom edge.

**Root cause — a defect in this spec, not in the implementation.** The frozen **Never** clause read *"Never
override the natural DOM stacking order to resolve overlaps (see AC3 — it is already deterministic)."* That
was written about **thin-vs-thin** overlap, where determinism was already free, and it over-reached into
**pad-vs-a-neighbour's-paint**, a case the I/O matrix never covered — forbidding the one remedy that case
needs. A second lock compounded it: the byte-identity guard's declaration allowlist pinned
`.canvas-component::before` to `['background','content','inset','position']`, so `z-index` reddened the
story's own test.

**What was amended** (frozen block; only a human may change it, and the wording is the coordinator's):
1. The stacking clause was **narrowed to thin-vs-thin** and now states explicitly that a component's own
   paint outranks a neighbour's invisible pad, and that arbitrating that is *required, not forbidden*.
2. **Two I/O matrix rows added** — paint outranks a neighbour's pad, and its complement, the pad still wins
   over empty canvas. The complement is deliberate: it stops the arbitration being implemented as "the pad
   never wins anything".

**Known-bad state avoided.** Shipping a story whose stated purpose is making thin components easier to grab
while silently making every component adjacent to a thin one *harder* to grab — with the regression
unreachable by any test, because the frozen block forbade the fix and the guard's allowlist blocked it.

**No loopback occurred. `review_loop_iteration` stays 0.** The captured intent was right; one clause
over-reached. Reverting verified work to re-derive near-identical code would have burned paid-for
verification and changed no conclusion. The frozen block reserves amendment to a human precisely so this
case has a remedy that is not a revert.

**KEEP — what worked and must survive.** The `::before` (never `::after`) choice and its reason: `::after`
is freed only for lines, so padding hung there would silently fail for every non-line thin element. The
`.canvas-box` byte-identity guard and its **declaration allowlist** — it did its job here, catching an
unauthorized property; it is widened by authorization with the reason recorded, never bypassed. The
gather-from-the-sheet idiom for pad rules. The AC3 stability test's independence from the padding (proved:
it still passes with `COMFORTABLE_HIT_PX = 0`).

## Design Notes

**Why a `::before` and not a wrapper element.** `.canvas-component` is already `role="button" tabIndex={0}`
and already inside `control-vocabulary-contract.test.tsx`'s sweep. A nested focusable element would put a
button inside a button and enter R1–R4 of the vocabulary contract. A pseudo-element adds no node, no role and
no accessible name, and hit-tests as the originating element — which is exactly how `.resize-handle::after`
already works, in reverse. **And it must be `::before`, not `::after`, for a reason that would otherwise be
found the hard way:** `::after` is freed (`content: none`) **only for lines**, at `App.css:243`. Padding hung
there would work for a Line and silently fail for every other thin element — passing a test suite written
only against lines, which is the shape of defect this story is trying not to repeat.

**Why the pad is computed in TS rather than in CSS.** CSS cannot branch on whether a custom property's value
is under a threshold, and the pad must be zero for components already large enough. The projection values are
in hand at `componentStyle`, so the comparison is ordinary arithmetic on document units. Deriving it from the
DOM instead would hit `getBoundingClientRect`/`offsetWidth` and need a numbered carve-out this story has been
told not to open.

**AC3 is preservation, and the test must say so.** Alternation is already structurally impossible: no
`.canvas-component` carries a `z-index` (App.css has six `z-index` declarations, none on it), so the last DOM
sibling paints on top and receives the pointer; and `begin()` calls `stopPropagation()` at `App.tsx:4337`, so
exactly one component handles a given pointerdown. "Topmost in document order" therefore resolves to
last-in-document, which is also the one visually on top. The repeated-click test is a **regression guard over
a named mechanism**, not a discovery — write it so a reader can see which mechanism it protects.

**What jsdom can and cannot prove here, stated plainly.** It can see which element carries which handler, the
inline custom properties, the class names, focus, the absence of the AC1 sentence, and the exact command bytes.
It cannot see any consequence of `max(2px, …)`, of a pseudo-element, of stacking, or of a pointer landing
near-but-not-on a rule — jsdom applies no stylesheet and every rect is zeros. Story 14.2 shipped three guards
that read correctly and could not observe what they named; the split above is drawn to avoid repeating that.

## Verification

**Commands (the per-story cadence, D-000.33):**
- `cd folio-designer && npx vitest run` -- expected: all pass. Record file/test counts, and diff the test-name
  set against the baseline rather than comparing totals.
- `cd folio-designer && npx tsc -b --force` -- expected: clean. **`--force` is mandatory**; `tsc -b` is
  incremental and can exit 0 having checked nothing.
- `cd folio-designer && npx oxlint` -- expected: no errors; exactly 4 pre-existing `only-export-components`
  warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx` — re-measure the two line numbers, they move).
- `cd folio-designer && npm run test:e2e:compile` -- expected: clean. **This is `tsc --noEmit` only. It proves
  the new e2e spec compiles; it does not run it and is not coverage.**

Capture each as `cmd > log 2>&1; echo $?`. The shell is zsh: `${PIPESTATUS[0]}` is empty and `$?` after a pipe
is the last stage's. Quote every glob.

**Suites that DID NOT RUN in this story, named per D-000.32:** the browser suite, the Go suites, the matrix
legs, `npm run build` as a gate, the `verify:offline*` chain, and the font-host scans. These run at the Epic 14
boundary gate, which is the owner's. The new `e2e/` spec is **compiled, not executed** — it is scheduled to run
at that gate, and must not be described as coverage until a run log shows it.

**Mutation proofs required (not optional).** For each of AC1, AC2's arithmetic, AC3's stability and AC4's focus,
revert the implementation and confirm the naming test actually reds.

**One mutation is mandatory and is not a revert — it is the whole point of the `.canvas-box` byte-identity
assertion.** Re-implement the padding **as box inflation** (widen the component box itself, the way
`max(2px, …)` does, instead of hanging a negative-inset `::before` off it) and confirm the byte-identity test
**reds**. If it stays green, the assertion cannot tell this story's mechanism from the defect DW-345 records,
and it is worthless — that is the 14.2 failure mode exactly. Also confirm the reachable-area half of the same
test reds when the padding is removed entirely, so both halves are shown to be load-bearing rather than one
carrying the other. `App.tsx` contains two literal NUL bytes
near lines 3700-3701: plain `diff` prints "Binary files differ" with zero changed lines, so **`diff -a` and
`cmp` are mandatory** to prove each mutation landed and was restored. Use absolute paths for every restore.

## Suggested Review Order

**Start here — the placement-to-selection path (AC1, AC4)**

- The entry point: how a placement learns the id the protocol never returns.
  [`App.tsx:835`](../../folio-designer/src/App.tsx#L835)

- Both placement spellings; `from` is captured at the gesture, not the response.
  [`App.tsx:926`](../../folio-designer/src/App.tsx#L926)

- The second spelling, for sheets that did not exist before.
  [`App.tsx:947`](../../folio-designer/src/App.tsx#L947)

- Selection routes through the existing `select()`; focus becomes a claim, not a timer.
  [`App.tsx:976`](../../folio-designer/src/App.tsx#L976)

- The claim is taken on the render that mounts the element — no macrotask to cancel.
  [`App.tsx:1021`](../../folio-designer/src/App.tsx#L1021)

**The hit pad, and the arbitration that keeps it honest (AC2, AC3)**

- Why the pad sinks below every box: the load-bearing stacking note.
  [`App.css:153`](../../folio-designer/src/App.css#L153)

- `isolation: isolate` — the enabling half; the z-index is inert without it.
  [`App.css:169`](../../folio-designer/src/App.css#L169)

- The pad itself: paint-free, negative inset, `z-index: -1` by authorization.
  [`App.css:269`](../../folio-designer/src/App.css#L269)

- Inert while a placement is armed, keyed on the pre-existing class.
  [`App.css:278`](../../folio-designer/src/App.css#L278)

- The 12px contract, floored at the paint floor so the extent is exactly 12px (DW-345).
  [`App.tsx:4710`](../../folio-designer/src/App.tsx#L4710)

- Pad emitted as a custom property from the projection — never from the DOM.
  [`App.tsx:4718`](../../folio-designer/src/App.tsx#L4718)

**The guards that could actually fail**

- The one that separates this story's mechanism from DW-345's defect.
  [`App.test.tsx:1401`](../../folio-designer/src/App.test.tsx#L1401)

- Pad below every box, plus the ancestor stacking context that makes it work.
  [`App.test.tsx:1555`](../../folio-designer/src/App.test.tsx#L1555)

- The row whose mutation previously left the whole suite green at 1217/1217.
  [`App.test.tsx:5169`](../../folio-designer/src/App.test.tsx#L5169)

- Focus lands on the mounting render: flushes microtasks only, so a timer cannot pass.
  [`App.test.tsx:1237`](../../folio-designer/src/App.test.tsx#L1237)

- Comment-stripping helper — a guard that read commented-out code was blind.
  [`App.test.tsx:1196`](../../folio-designer/src/App.test.tsx#L1196)

- AC3 determinism, proved independent of the padding.
  [`App.test.tsx:1589`](../../folio-designer/src/App.test.tsx#L1589)

**Peripheral — pointer geometry, compiled only until the boundary gate**

- The two arbitration matrix rows, at real coordinates.
  [`placed-component-selection.spec.ts:122`](../../folio-designer/e2e/placed-component-selection.spec.ts#L122)

- Placement beats the pad, and the pad takes the pointer back afterwards.
  [`placed-component-selection.spec.ts:206`](../../folio-designer/e2e/placed-component-selection.spec.ts#L206)
