---
title: 'Story 14.10: A table column is bound from the main window'
type: 'feature'
created: '2026-09-10'
status: 'done'
review_loop_iteration: 0
baseline_commit: 'c209c74'
context: []
---

# Story 14.10: A table column is bound from the main window

**In plain terms.** *(Rewritten at close; the frozen Intent below governs implementation.)*

A table column is now bound the way everything else in a template is bound. You click the column on the
canvas and it becomes the selection; the DATA tab then offers exactly the fields that make sense for it,
picking one binds the column, and it undoes in a single step. Fields that would not work are not offered,
and the panel says why rather than letting you pick something the engine would refuse. The table
editor still *shows* each column's bound field, but no longer lets you type one there. Closing that
second way in was the point of the story, not a side effect of it.

Two things a later reader may find surprising, both deliberate. A column selected this way can be reached
with a mouse and by no other means. The product carries a stated obligation to keep this reachable from
the keyboard, and it is not met; that was the owner's explicit call, and it is written down as unpaid work
rather than quietly absorbed. And the byte-for-byte reproducibility proof, which used to build its document
by typing into the box this story deletes, now opens a prepared document instead — the same proof taken a
different way, and run in a real browser before the epic closed.

Six further shortcomings noticed along the way were recorded as follow-up work instead of fixed here.
Nothing in the engine changed: the browser was already being told which column was which, and nothing was
reading it.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A table column cannot be pointed at. `.canvas-table` carries `pointer-events: none`
(`App.css:415`), so a click inside a painted table falls through to the component, and the paint
carries no addressability — headings and cells are plain `<span>`s with a React `key` and nothing
else. The only way to bind a column today is the table editor's `Row field` input, which puts one
value behind a dialog while every other binding in the product is made in the main window. That is
the shape [D-14.4.Q2(a)] already removed once, and `TableEditor.dc.html` never drew it: the design
draws BOUND FIELD as a dense binding chip with no input border, no chevron and muted ink, visibly
unlike every editable neighbour in the same row. **The shipped input is the thing that departs from
the design.**

**Approach:** Make a column addressable by **element resolution** — restore `pointer-events` on the
column spans, carry the column id as `data-column-id`, and read it off the event target inside the
existing `select` seam. Hold the column selection as **separate UI state keyed on both the owning
table id and the column id**, so `selected` keeps meaning exactly what it has always meant. Teach the
DATA panel to offer that column's row-scope fields, gated by the **same set** the table editor's
candidates already use, and commit through the existing `updateTableColumnBinding` command unchanged.
Then finish [D-14.10.1] by making the editor's BOUND FIELD display-only.

## Boundaries & Constraints

**Always:**
- **Element resolution, never coordinate geometry.** Read the column from the event target
  (`closest('[data-column-id]')`). Mapping a pointer coordinate to a column means computing where
  each column's edge lands on screen — a browser-side model of the columns — which **AD-15 / I-4**
  bars and which is the exact premise Story 14.9 was built on: *the canvas paints tables from the
  engine's projection, never a browser-side model of the columns.* Cite AD-15 and 14.9's premise.
  **Do not cite the AD-17 corpus scan for this** — see the correction under Rulings.
- **Resolve the column from the projection on every render. Never cache the column object.**
  If the id does not resolve, the selection drops.
- **Clear the column selection at every site that clears or replaces `selected`.** The authoritative
  set is the six measured `setSelected(` sites in the Code Map — not a remembered list.
- **Every existing `length === 1` gate keeps its current meaning, and none of them learns about
  columns.**
- **One source of truth for "in row scope".** The panel's pickable set is *exactly*
  `tableSampleCandidates`' output, asserted against a fixture. AC4's refusal message is derived from
  that same set, never written independently.
- No new design token. `design-contract.test.ts` asserts token-name equality against a read-only
  `DESIGN.md`.
- Selection is marked in cyan. `DESIGN.md` states three separate times that the bind accent never
  marks selection.

**Ask First:**
- Any change to a file the Tasks do not name.
- Any pin, floor, census or fixture that must move for a reason **other** than transcription of a
  determined value (see Q4a under Rulings for the test).
- Any second derivation of "in row scope", of the collection key, or of the row-relative field path.
- Any widening of the AD-17 carve-out, or new top-level code between `placementPoint` and
  `pageStyle` in `App.tsx` — that gap is inside the carve-out's lazy match and would silently
  receive a free waiver.

**Never:**
- **No keyboard reach.** [D-14.10.3] rules it out of scope; DW-387 registers it. Do not add
  `tabIndex`, key handlers or roving focus to the canvas table, and do not raise it as an Open
  Question — the answer is already on the record. The unmet requirement is quoted, unsoftened, under
  Rulings.
- No `role`, `tabindex`, `aria-label`, second `data-component-id`, `role="group"` or `.canvas-box`
  on any inner node of the table paint. `canvas-table-paint.test.tsx:185` plant-proves all six.
- No binding section and no column controls in the inspector (Q3).
- No fix for DW-388 (the `[].every()` ghost panel). It is filed; this story neither creates nor
  touches it.
- No change to what the table editor's `candidates` prop receives — that is a silent 14.7 regression.
- No change to `sample-data.ts`.
- No new engine or projection field. The column id is already on the wire.
- Never widen, de-anchor or "fix" `KNOWN_RED_TEST: "^TestCorpusMeetsP6ExerciseFloors$"` (D-000.17).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Click a column heading or cell | Table painted with columns; click lands on a `[data-column-id]` span | That column becomes the column selection; the owning table stays the component selection; the column is marked in cyan | N/A |
| Click the table outside any column | Click lands on the component but resolves no `data-column-id` | Table selected, **no** column selection | N/A |
| Shift-click a column | Column already selected, shift held | Behaves as a plain click; a mixed table+column selection is unrepresentable by construction (Q1(b)) | N/A |
| DATA tab with a column selected | Sample loaded; table bound to `transactions[]` | Row-scope fields under that collection are **pickable**; context bar names the column and what will be bound | N/A |
| Pick a row-scope field | Column selected, field picked | One `updateTableColumnBinding` command; one undo step | Engine refusal shown where the pick was made |
| Pick the collection itself | `transactions[]` node clicked with a column selected | Not pickable; reason stated **before** the engine would refuse (UX-DR24) | N/A |
| Pick a path outside row scope | A document-scope scalar clicked with a column selected | Not pickable; reason derived from the same set that gates pickability | N/A |
| Selected column removed in the editor | Column selected, then removed; `commitTableColumn` re-projects | Id no longer resolves → column selection drops; the panel never offers a column that does not exist | N/A |
| Table has no columns | `columns === undefined` | 14.9's empty-state notice unchanged; nothing addressable | N/A |
| Column with an empty bind | `column.bind === ''` | Still selectable and bindable; paints as unbound per 14.9 | N/A |

</frozen-after-approval>

## Rulings — the plan gate. SETTLED. Carry these; do not re-open them.

All five plan-gate forks were ruled before this spec was written. They are recorded here verbatim in
substance so the implementer meets the ruling rather than the question.

**D-14.10.1 (OWNER, 2026-09-07/10) — the table editor's BOUND FIELD is display-only.** Binding is
edited in the main window. The editor becomes structure-only: add, remove, reorder, width, align,
footer aggregate. Its **collection** and **row alias** stay editable — this ruling scopes to the
per-column bound field and nothing else. `epics.md` §14.10's final criterion was amended on
2026-09-10 to say so and to cite the ruling; the "unless the owner rules otherwise" clause is spent.

**D-14.10.3 (OWNER, 2026-09-10) — 14.10 ships mouse-only.** Keyboard reachability is out of scope and
registered as DW-387, with AD-17's `[ASSUMPTION]` clause attached. Verified rather than assumed: every
accessibility or keyboard mention inside Epic 14 (`epics.md` 4821–5497) sits inside a specific story's
criteria — 14.1, 14.3, 14.5, 14.6, 14.7, 14.7b, 14.8 — and none is epic-level prose. §14.9 and §14.10
carry none of any kind.

**The unmet requirement, quoted and reported unmet (D-000.17 is absolute).** UX-DR25 (`epics.md:283`):
> *"Meet the accessibility floor as behavioural obligations, independent of formal conformance being
> out of scope: every interactive element keyboard-reachable and operable (palette, properties,
> binding tree, **table columns**); visible focus on every focusable element using `colors.select`;
> accessible names on every icon-only control; errors and diagnostics announced and distinguished by
> shape before colour; canvas handle hit targets larger than their visual footprint; the Table Editor
> behaving as a data grid under keyboard navigation."*

The requirement names table columns explicitly. **Story 14.10 does not meet it.** A column selected by
this story is reachable by mouse and by no other means. This is not a gap the story closed, narrowed,
or found acceptable; it is a stated obligation left unpaid by owner ruling, registered as DW-387.

**Q1 — (b). The column selection is separate UI state, keyed on BOTH the owning table id and the
column id.** `selected` keeps meaning exactly what it has always meant. Grounding is **AD-15 / I-4**:
transient interaction state lives in the UI, never in the document. Decisive against a compound id
inside `selected`: it **breaks working function** — `openTableEditor` guards on
`selectedRef.current[0] !== id`, so a compound id kills the table's own "Configure columns" button,
and a story that silently repairs four other paths to accommodate its own state shape has left its
fence. The arm that spends nothing was available. Three guardrails, all required, are in Boundaries.

**Q2 — (a). Thread `tableSampleCandidates`' output into `DataPanel`; leave `sample-data.ts` exactly as
it is.** The decisive property is **agreement with the engine**, not "leave the shared parser alone".
AC4 requires the panel to decide row-scope membership *before* the engine would refuse, and
`tableSampleCandidates` already builds the collection key byte-identically to the canvas's
`tableBind`. That agreement is shipped and exercised. A second derivation would owe a proof that the
two agree; there is only one, so it owes nothing.

**Q3 — (a). The owning table's panel unchanged, plus an identity strip naming the selected column. No
column controls in the inspector, and no binding section in the inspector.** [D-14.10.1] refuses one
value with two editing sites; a binding section here would put binding in the inspector *and* the
DATA panel, re-committing the exact error [D-14.4.Q2(a)] removed, inside the story written to finish
removing it. AC1's own gloss settles what "addresses it" means: *"selecting a column is how binding
one **begins**"* — beginning is acknowledgement; the binding itself is AC2's, in the DATA panel. The
strip names the column in the vocabulary the canvas and editor already use (the column's label),
offers no control, and disappears with the column selection. It is a forced consequence of AC1 and is
the whole of what is forced. "Configure columns" stays live throughout.

**Q4 — INSIDE THE FENCE. 14.10 removes the editable Row field, as AC5 reads.** The owner ruled this
exact point knowingly: [D-14.10.1]'s *"Why the question was put at all"* paragraph exists precisely
because the criterion *"removes a path that works, which is a different act from declining to build
one."* It is a **restoration**, not merely an authorized subtraction — the `<output
id={binding-display-N}>` already exists, so *"keeps showing"* is satisfied today and only the removal
is outstanding, and `TableEditor.dc.html` draws BOUND FIELD as a chip the shipped input departs from.

**Q4a — the four pinned traversal assertions are transcription, not accommodation. Update all four in
this story's commit, keep them as walks, form unchanged.** The reusable line, on the record: **an edit
is accommodation when the new value is chosen; it is transcription when the new value is determined.**
[D-14.8.2]'s floor move was a judgement about how much margin to keep, and an arm existed that spent
nothing. Neither holds here — AC5 forces the removal, and ArrowRight from `Header for column 1` lands
on `Width for column 1 in points` whether anyone likes it or not. `TableEditor.tsx:12-35` finishes it:
the lattice covers the **maximal** row shape, a smaller row has holes, and `moveFocus`'s `enabled()`
already treats an absent cell exactly like a disabled one. A hole at address 4 is designed behaviour,
so the updated walks assert the same property over a lattice one cell smaller — same guard, smaller
population, no weakening.
- **Delete exactly one hop; change nothing else about the walk.** Same exact-accessible-name
  assertions, same form. **If an update reaches for a regex, a `some`-style match, or "the next
  enabled cell", STOP** — that is the accommodation [D-14.8.2] forbids, and it converts a pin into a
  guard that cannot fail.
- **Keeping them as walks is binding, not preference.** A deleted traversal assertion is how the next
  removal goes unnoticed. Four assertions that survive a lattice change by *failing* are the only
  thing standing between this grid and silent keyboard rot.

**Q4b — OWNER DECISION: split the TEST, not the story.** CAP-13
(`e2e/browser-native-roundtrip.spec.ts`) keeps its byte-identity claim and **authors from a LOADED
document** — it stops typing its document into the UI and opens a prepared `.folio` instead. **A
SEPARATE spec covers "the new binding path produces the expected document."** This is the only arm
that fixes **failure attribution**: a red in CAP-13 then means reproducibility broke; a red in the new
spec means the UI moved. It does not *delete* the "authored through the browser's own commands
round-trips" leg — it **relocates** it into its own test, so nothing is lost and each test says one
thing. Permanently coupling the most trusted number in the project to first-run designer UI was the
arm the lead least wanted to defend later, given what a misread of exactly this signal cost in DW-383.

**Q5 — the browser proof is REQUIRED, narrowed to a targeted run.** Both covers ship: the `App.css`
source-text pin **mutation-proved** by deleting `pointer-events: auto`, **and** the real-browser
click. The browser run is the forced price of AC1's own stated acceptance — *"the author clicks a
column"* has no other instrument. The pin alone is a **proxy**: it proves the rule exists in the
stylesheet, not that the click resolves to the column, and a later overriding rule, a covering
element, or a wrapper that still captures would each leave it green. **The Playwright assertion must
fail if `pointer-events: auto` is removed — prove that red, do not assert it.**

**A correction to a premise this story was dispatched with, carried so it is not repeated.** The claim
*"coordinate-based hit-testing is blocked at the AD-17 contract"* is **withdrawn as to its reason**.
Measured against the `prohibited` array in `canvas-authority-contract.test.ts`: `clientX`/`clientY`
are **not banned anywhere** — pattern 4 is `client(?:Width|Height|Left|Top)` only, and `App.tsx`
already uses `event.clientX/clientY` at eight sites outside the carve-out with the scan green.
`elementFromPoint` is not banned either. What the scan forecloses is **box measurement**:
`getBoundingClientRect`, `offset*` (including `offsetX/offsetY`), `client{Width,Height,Left,Top}`,
`scroll*`, `getComputedStyle`, `ResizeObserver`. **The verdict stands on a different AD** — AD-15 /
I-4, and 14.9's premise — because mapping a coordinate to a column is a browser-side model of the
columns: the AD-17 *harm* reached without spelling a prohibited identifier, which is the case a scan
cannot catch and a ruling must. This is the catalogue's *"a correct finding carrying an incorrect
explanation"*, caught at the gate.

**DW-388 — filed, not fixed.** The inspector opens on an unresolvable selection and renders every
control live: `App.tsx:2812` gates on `selected.length > 0 && canvas` — on the selection being
non-empty, never on its ids resolving — and `all = (predicate) => [...types].every(predicate)` over an
empty Set is **vacuously true for every predicate**, so the panel renders in full under an identity
strip reading `0 selected`. Reachability is **UNKNOWN and stated that way**: `applyHistory`
(`App.tsx:2278`) passes `clearDocumentInteraction = true`, so undo/redo — the obvious route — does
clear the selection. That is an unsuccessful search, not a proof. **This story neither creates nor
touches it**, because Q1(b) leaves the panel seeing exactly what it sees today.

**D-14.0.1 / D-14.10.4 — the slot-cost obligation.** Five consecutive stories skipped it and the
recorded margin turned out wrong by 50%. **14.10's slot cost is 0** — it adds no asset, no font, no
image, no vendored file; every change is TypeScript, CSS and tests. Because Q5 puts a real
`npm run build` in the path anyway, `s1.assetCount` is read off that build and the **actual integer**
reported, not this reasoning.

## Code Map

All anchors measured at `96ef4a6`/`48e4f40` on 2026-09-10. Those two commits differ only in
decision-log documents — `git diff --stat 96ef4a6..48e4f40` is two `.md` files, no toolchain input —
so every measurement holds at either. **Anchor rot is endemic in this epic: find every site by its
named identifier, `role` or `aria-label` and re-measure before editing.**

### The selection model — `folio-designer/src/App.tsx`

- `const [selected, setSelected] = useState<ReadonlyArray<string>>([])` (`:271`). Holds **element ids
  only**, and continues to.
- **`setSelected(` occurs at exactly SIX lines, all in `App.tsx`, and nowhere else in the repo.**
  This is the authoritative clear-site set, measured, and it supersedes any remembered list:
  | line | site | what it does |
  |---|---|---|
  | 1253 | `select` | `setBindingError(undefined); revokeTableEditor(); setSelected(…)` — the one selection primitive |
  | 1331 | `deleteSelection` | clears on success |
  | 2171 | `setCurrentSnapshot`'s `clearDocumentInteraction` branch | clears everything |
  | 2294 | Locate in Design | the only mutation site that already validates against the projection |
  | 2568 | `page-surface` `onClick` | the deselect route |
  | 2779 | canvas-region Escape | the other deselect route |
  Positive control for the repo-wide half: `DataPanel.tsx` and `TableEditor.tsx` both contain
  `selected` and **neither contains `setSelected`**, so the empty result elsewhere is evidence rather
  than a filtered grep.
- `selectedRef` is a `useEffect` mirror on `[selected]` (`:381`, `:447`). **It needs no separate
  clear** — it cannot hold a value `selected` does not.
- `const select = (id: string, extend: boolean)` (`:1253`), bound as `onSelect={select}` at `:2585`.
  Called from `selectPlaced` (`:1277`) and from three places inside `CanvasComponent`: `begin`
  (`:4830`, the pointerdown path and **the one that runs on a real click**), the div's `onClick`
  (`:4842`, a de-duplicating fallback via `selectedByPointer`), and the div's `onKeyDown` Enter/Space.
  `begin` and `onClick` both call `event.stopPropagation()` first, to stop the click reaching the
  `page-surface` clear at `:2568`.
- `deleteSelection` (`:1331`), `duplicateSelection` (`:1332`), the canvas-tools Delete and Duplicate
  buttons (`:2780`, both `disabled={selected.length !== 1}`), the canvas-region Delete/Backspace
  handler (`:2779`), and window Cmd/Ctrl+D (`:2426`). **Under Q1(b) every one of these keeps seeing a
  real table id and behaves exactly as it does today.** They are listed so the required guard is
  written against the measured set, not a remembered one.
- `openTableEditor` (`:1004`) guards `selectedRef.current[0] !== id`. Under Q1(b) the table remains
  the component selection, so **"Configure columns" stays live** — this is the working function the
  compound-id arm would have broken.
- `bindPickedPath` (`:1199`) — the scalar-binding helper. Its own `bindingInFlight` ref, `bindingBusy`
  state and scoped `bindingError`, one round trip. This is the shape to match for a column bind; it is
  **not** `commitComponent` and **not** `commitTableColumn`.
- `const selectedComponent = selected.length === 1 ? canvas?.components.find(…) : undefined` (`:599`).
- The inspector gate (`:2812`): `selected.length > 0 && canvas ? <ComponentProperties … /> : <PageSetup … />`,
  with `components={canvas.components.filter((component) => selected.includes(component.id))}`. See
  DW-388 — **do not fix it here.**

### The paint — `folio-designer/src/App.tsx`, `App.css`

- `function TablePaint({ component, zoom })` (`:4972`). Headings at `:5003`, cells at `:5009`, both
  `columns.map((column) => <span key={`${component.id}-…-${column.id}`} className="… canvas-display-paint" style={{ textAlign: … }}>`.
  **`column.id` is already in the React key — the hook a hit test needs is on the wire.** No engine or
  projection change is required.
- `.canvas-table { … pointer-events: none; … }` — `App.css:415`. **No descendant overrides it**, and
  no test in the corpus pins it. `pointer-events: auto` on a descendant is the only CSS change needed.
- ⚠ **jsdom does not implement `pointer-events` hit-testing.** `fireEvent.click(span)` makes the span
  the target regardless. In a real browser the paint is removed from hit testing and the click
  resolves to the wrapper. **A DOM-targeting implementation goes green in all 1409 unit tests and is
  completely inert in the product if the CSS is omitted or later removed.** This is why Q5 requires
  both covers.
- Do **not** put new top-level code between `placementPoint` and `pageStyle` — the AD-17 carve-out is
  `/export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/`, a lazy
  match that would silently swallow it and grant a free waiver.

### The DATA panel — `folio-designer/src/DataPanel.tsx`

- `const rowFor = (entry: VisibleNode): RowShape` — the whole pickability decision. Today:
  `pickable: node.segments !== undefined && !collection && !runtime && !scoped`, with a per-row
  `reason` string rendered as `{shape.reason && <span className="tree-reason">{shape.reason}</span>}`
  **inside** the treeitem button, plus `data-tree-dim` on the `<li>` and
  `aria-disabled={!branch && !shape.pickable ? true : undefined}`.
- ⚠ **Row-scope leaves carry NO `segments` at all, so flipping `scoped` out of that expression is not
  enough.** Verified: `sample-data.ts` `array()` recurses with a hardcoded `rootScoped: false`, and
  `value()` gates the marker on `rootScoped && segments.length > 0`. The row-relative path the engine
  wants is simply not on the node. **This is why Q2 threads `tableSampleCandidates` in.**
- `const context = …` — the context bar, `<p className="binding-chip data-context…" role="status">`,
  no `aria-label`; tests find it by literal text. Today a selected table reads *"Table selected · a
  table binds its collection in the table editor, under Configure columns."* — still true under
  [D-14.10.1], since the **collection** stays editable there. A selected **column** needs its own
  sentence.
- Refusal: `BindingErrorScope` re-narrowed at `:50` by sample identity, component id and picked path;
  rendered as one `<p className="data-message" role="alert">`.
- ⚠ **`engine-bounds-mirror.test.ts` pins DataPanel's import line byte-exactly**
  (`^import \{ SCALAR_BINDING_COMPONENT_TYPES, type CanvasComponentType \} from '\./engine-protocol'$`),
  pins `bindableKind`'s exact expression, and **forbids the literal `=== 'text'` anywhere in the
  file**. A **second, separate import statement** on its own line leaves the pin byte-identical and is
  legal; nothing enables `no-duplicate-imports`. `App.tsx`'s import line is **not** pinned.

### The row-scope derivation to reuse — `folio-designer/src/App.tsx`

- `tableSampleCandidates` (`:68`) — walks a collection's children, builds `prefix.join('.')` from
  `child.label` gated on `SAMPLE_PATH_SEGMENT = /^[A-Za-z_][A-Za-z0-9_]*$/` (the browser's spelling of
  Go's `rootValuePath`), and builds the collection key as `` `${node.segments.join('.')}[]` `` —
  **byte-identical to the canvas's `tableBind`**, so matching is a plain string compare with no
  parsing. Widening its signature or return is in-fence; **changing what the table editor receives is
  a silent 14.7 regression.**

### The commit path

- `folio-designer/src/table-column-command.ts:22` —
  `export function updateTableColumnBindingCommand(id: string, columnId: string, field: string): ArrayBuffer`.
  **Unchanged by this story.**
- Engine side (`folio-go/component_commands.go:576`) validates `field` against
  `^[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*$` and `len(field) > 192`, resolves the alias
  itself (`"row"` unless `table.As` is set), and writes `Bind = "{{" + alias + "." + field + "}}"`.
  **The command takes the bare row-relative field — never the alias, never the collection.**
- **One undo step is guaranteed by Go, not the browser**: `wasm/engine.go`'s `Apply` pushes exactly
  one undo per accepted byte-changing command and pushes nothing when the bytes are unchanged. AC6 is
  satisfied by *not bypassing the command path* — the risk to watch is a bind that also had to send
  `configureTableBinding` first. **Assert it; build nothing.**

### The Row field removal — `folio-designer/src/TableEditor.tsx`

- `:492` — the `<input {...matrixCell(index, CELL.bound)} aria-label={`Row field for column ${index + 1}`} list=… disabled={busy || !column.rowFieldEditable} aria-describedby=… onBlur={…onBind(column.id, field)} />`,
  its `<datalist>`, and the sibling `<output id={`binding-display-${index}`} aria-label={`Binding for column ${index + 1}`}>{column.binding}</output>` **which already displays the bound field and stays.**
- `onBind` has **one** production call site (`App.tsx:2854`) and two test render sites that pass it as
  a prop without exercising it.
- `CELL = { …, bound: 4, … }` and `ALIGN_CELL = 6` are **constants, not computed from occupancy**, so
  removing the occupant of address 4 moves nothing else. `TableEditor.tsx:12-35` states the lattice
  covers the maximal row shape and a smaller row simply has holes.
- **`rowFieldEditable` must keep being projected.** `engine-protocol.ts:639` uses `hasExactKeys`, so
  dropping it from the wire is fatal. It will then have **zero production readers** — assert that it
  still arrives, and **tie the assertion's message to [D-14.10.1]** so the next reader finds the
  ruling instead of the puzzle.

### The guard census

**Will red if the change is wrong, and is the point:**
- `folio-designer/src/canvas-table-paint.test.tsx:185` — *'keeps the component the single control…'*.
  Over a **five-column** table, asserts `[role="button"], button`, `[data-component-id]`,
  `[aria-label]`, `[role="group"], [role="tablist"]`, `[tabindex]` and `.canvas-table .canvas-box` are
  each length 0, and **plant-proves every one** by adding the forbidden attribute at four positions in
  turn. **`data-column-id` and `pointer-events` are not in that list.** It also asserts a letter or
  digit survives the `aria-hidden` strip — keep the paint reading as a `word`.
- `canvas-table-paint.test.tsx:266` — every element painting an engine-owned string carries
  `canvas-display-paint`. **Do not drop that class while adding `data-column-id`.**
- The four pinned traversal/focus assertions the removal moves (Q4a): `App.test.tsx:303`, `:755`,
  `:1463` (each `header.focus()` → ArrowRight → *Row field* → ArrowRight → *Width*), and
  `e2e/table-editor.spec.ts:46`.

**Measured and NOT tripped — checked so nobody chases them:**
- `TableEditor.test.tsx`'s two roving assertions are **self-deriving**: `:344-346` compares the walk
  against `[data-matrix-cell^="0:"]` filtered by `:disabled`, and `:780` asserts address uniqueness.
  Removing a cell removes it from both sides. `:469` pins `aria-colcount` at `'6'`, which counts
  **drawn** columns — BOUND FIELD stays drawn as a chip, so it holds.
- `control-vocabulary-contract.test.tsx` is **blind** to this change: its `tableComponent` fixture is
  `{ id: 'e7', type: 'table', …, resizable: false }` with **no columns and no `tableBind`**, so
  `TablePaint` takes its empty-state early return in all seven states and no column span ever renders.
  All four floors (`PER_STATE_CONTROL_FLOOR 15`, `CONTROL_FLOOR 220`, `CLASS_FAMILY_FLOOR 18`,
  `GROUP_INSTANCE_FLOOR 33`) are one-sided `<` comparisons — **adding never reds a floor.**
- `canvas-authority-contract.test.ts` — `dataset`, `getAttribute`, `closest`, `matches`, `target` and
  `pointer-events` appear **zero** times in the `prohibited` array (positive control: `getComputedStyle`
  appears once, as rule 9). Its file-count floors are all `toBeGreaterThan(OrEqual)`, so adding never
  reds; production `.css` sits **exactly on its floor of 3**, so do not merge or delete a stylesheet.
  It pins `App.css`'s media queries to exactly `['prefers-reduced-motion: reduce']` — **add no
  `@media`** — and forbids a `.canvas-display-*` class name containing `table|column|header|chip`.
- `design-contract.test.ts` — `App.css` must carry **zero** colour literals (the regex scans comments
  too), no gradient, every `border-radius` a `var(--radius…)`. `cursor:` is unconstrained anywhere for
  `.canvas-table-*`. A new `.canvas-table-heading { pointer-events: auto; cursor: pointer }` trips
  nothing.
- `App.test.tsx:1876`/`:1896` — every non-pseudo rule whose selector contains `.canvas-component` must
  carry no `z-index`. Write the new rule as a bare `.canvas-table-*` selector with no `z-index`.
- `App.test.tsx:1921` pins `[data-component-id]` document order to `['e1','e2']` — another reason the
  column attribute must be `data-column-id`, not a second `data-component-id`.
- `e2e/table-matrix-layout.spec.ts:50` uses `getByRole('button', { name: /table component/ })` under
  Playwright **strict mode** — a second control matching that name would break it.

### CAP-13's replacement fixture (Q4b) — proven to meet the precondition

- **`fixtures/statement-1/input.folio` carries bound table columns.** Measured, not assumed: element
  `e8`, `"type": "table"`, `"bind": "transactions[]"`, `"as": "txn"`, and **five columns, every one
  bound** — `e9`/`Date`/`{{txn.date}}`, `ea`/`Reference`/`{{txn.ref}}`,
  `eb`/`Description`/`{{txn.description}}`, `ec`/`Note`/`{{txn.note}}`, and
  `ed`/`Amount`/`{{formatNumber(txn.amount, "#,##0.00")}}` with a `sum` footer. Its sibling
  `data.json` and `params.json` supply what a render needs. **All READ-ONLY.**
- The loading idiom is shipped and exercised — do not invent one. `preview-navigation.spec.ts:27`:
  `const template = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '…'))`,
  then `await (await templateChooser).setFiles({ name: 'statement.folio', mimeType: 'application/json', buffer: template })`
  after clicking `Open local template`, asserting `.document-name` reads back. Eleven other specs use
  the same shape.

### The e2e gate — the cadence's stated scope is stale

`folio-designer/package.json`: `"test:e2e:compile": "tsc -p tsconfig.e2e.json --noEmit"` and
`"test:e2e": "playwright test"`. **CI runs BOTH.** `.github/workflows/ci.yml` has a second job,
`folio-designer-e2e`, `timeout-minutes: 45`, no `continue-on-error`, which runs `npm run test:e2e` in
full. The suite declares **52 tests** across 24 spec files (counted at column 0; indent-tolerant count
agrees, and there is no `test.describe`/`.only`/`.skip`/`.each` anywhere, so declarations equal tests).
⚠ The CI comment beside that job still says *"36 Playwright tests"* and is **stale by 16** — prose
only, no behavioural effect, and **the orchestrator's to correct, not this story's.**
`playwright.config.ts` has `testDir: './e2e'` and no `projects`, so a single-file or `--grep`
invocation is accepted; the `webServer` cold build (`npm run build && npm run preview`, where `build`
chains `scan:font-hosts && scan:host-fonts && build:wasm && tsc -b && vite build && build:offline &&
verify:offline`) runs regardless, ~141s warm by the config's own comment. **That is the story's cost,
reported rather than buried.**

## Tasks & Acceptance

> ### HARD PROHIBITIONS — reproduced verbatim per D-14.4.1, and they bind every subagent
>
> You must **NEVER commit**. You must never run `git add`, `git commit`, `git push`, `git stash`,
> `git checkout`, `git reset`, `git revert`, `git restore`, or `git clean`. You must never create a
> branch. The orchestrator makes every commit in this run.
> Reading git state (`git log`, `git status`, `git show`, `git diff`, `git ls-files`) is permitted and
> encouraged.
> Never push. Never write to `~/.claude/` or any memory directory.
> You must NEVER touch `fixtures/declared-variants/expected.pdf`, `input.folio`, or `signoff.json` —
> human-attested. Every file under `fixtures/statement-1/` is **READ-ONLY** for this story.
> Do NOT edit `deferred-work.md`, `epics.md`, `DESIGN.md` or `sprint-status.yaml` — send the wording up
> instead.
> Never widen, de-anchor, or "fix" the known-red quarantine
> `KNOWN_RED_TEST: "^TestCorpusMeetsP6ExerciseFloors$"`. An unmet floor is reported unmet, never
> filled (D-000.17).
> Do **NOT** read `scratchpad/rulings.md` or any unprefixed scratchpad file — the scratchpad is SHARED
> across this multi-epic run. Name anything you write there with a `14-10-` prefix and a timestamp.
>
> **Stay inside this section's fence.** If the right change is in a file these tasks do not name — an
> adjacent bug, a refactor that would make this cleaner, a guard for a case the spec never demonstrated
> — **stop and report it**. Do not implement it, and do not implement a smaller version to avoid
> asking. The one exception is work with no alternative that makes a stated acceptance criterion pass;
> name the AC when you report it.

**Execution:**

- [x] `folio-designer/src/App.tsx` — in `TablePaint`, add `data-column-id={column.id}` to the heading
      and cell spans. **Add nothing else**: no `role`, no `tabIndex`, no `aria-label`, no second
      `data-component-id`, and keep `canvas-display-paint` on both. — `canvas-table-paint.test.tsx:185`
      plant-proves all six absences; `:266` pins the display-paint marker.
- [x] `folio-designer/src/App.css` — restore hit-testing on the column spans with
      `pointer-events: auto`, and add the cyan column-selected treatment reusing
      `var(--color-select)` + `var(--tint-select-fill-soft)`. No new token, no `@media`, no colour
      literal (comments are scanned), no `z-index`, and do not group the rule with a `.canvas-text*`
      selector. — the design draws no selected column anywhere, so this matches
      `.canvas-component-selected`, the shipped page-safe idiom.
- [x] `folio-designer/src/App.tsx` — add the column selection as separate UI state keyed on **both**
      the table id and the column id; resolve the column from the projection on every render and drop
      the selection when the id stops resolving; clear it at **all six** measured `setSelected(` sites.
      Read the column id off the event target inside the existing `select` seam via
      `closest('[data-column-id]')`. — Q1(b); `selected` keeps its meaning and every `length === 1`
      gate is untouched.
- [x] `folio-designer/src/App.tsx` — widen `tableSampleCandidates`' signature/return as needed and
      thread its output to `DataPanel`, **without changing what `TableEditor` receives**. — Q2(a); one
      source of truth for row scope, already agreeing with the engine.
- [x] `folio-designer/src/App.tsx` — add the inspector's column identity strip (label only, no
      control, disappears with the selection); add the column-bind commit helper in `bindPickedPath`'s
      shape, issuing `updateTableColumnBindingCommand` unchanged. — Q3(a); no binding section in the
      inspector.
- [x] `folio-designer/src/DataPanel.tsx` — make row-scope fields pickable for a selected column, gated
      by **exactly** the threaded candidate set; derive AC4's refusal reason from that **same** set;
      add the column context bar sentence. Add any new import as a **second, separate import line** —
      the existing one is pinned byte-exactly — and do not write the literal `=== 'text'`.
- [x] `folio-designer/src/TableEditor.tsx` — remove the Row field `<input>` and its `<datalist>`,
      keep the `<output>` that already displays the binding, drop the `onBind` prop, and **update the
      `honest-note` and every comment describing the Row field as editable in this same change**. —
      Q4/D-14.10.1; D-000.28's stale assertion has been found in this file twice already.
- [x] `folio-designer/src/App.tsx` — remove the `onBind={…}` prop from the `<TableEditor>` render.
- [x] `folio-designer/src/App.test.tsx` — update the three traversal walks at `:303`, `:755`, `:1463`:
      **delete exactly one hop, change nothing else.** Same exact-accessible-name assertions, same
      form. **A regex, a `some`-style match, or "the next enabled cell" is forbidden — STOP and report
      instead.** — Q4a.
- [x] `folio-designer/e2e/table-editor.spec.ts` — update the pinned focus assertion at `:46` the same
      way, as transcription.
- [x] Add a test asserting `rowFieldEditable` **still arrives on the wire**, with a failure message
      naming [D-14.10.1]. — it now has zero production readers and is next quarter's silent deletion.
- [x] `folio-designer/src/App.test.tsx` (or the nearest owning unit file) — add the `App.css`
      source-text pin for `pointer-events: auto` on the column spans, in the shape `App.test.tsx:1709`
      already uses for `.canvas-box`. **Mutation-prove it** by deleting the declaration and showing
      the red.
- [x] Add unit tests for every row of the I/O & Edge-Case Matrix, including: the column selection
      dropping when the id stops resolving; the pickable set equalling `tableSampleCandidates`' output
      against a fixture; and the refusal reason coming from that same set.
- [x] **THE REQUIRED GUARD.** Prove positively that the delete path refuses or no-ops on a column
      selection, **with a test that FAILS if the guard is removed**. Demonstrate the failure and report
      the red output — do not merely assert it. Cover the measured consumer set: `deleteSelection`,
      `duplicateSelection`, the canvas-tools Delete and Duplicate buttons, the canvas-region
      Delete/Backspace handler, and Cmd/Ctrl+D. *"A guard that cannot fail"* is the first entry in this
      run's defect catalogue and we have shipped one before.
- [x] `folio-designer/e2e/` — **new spec**: in a real browser, click a table column and prove the
      selection lands on that column, and prove the DATA panel offers its row-scope fields and commits.
      **The assertion must fail if `pointer-events: auto` is removed — prove that red.** AC1's own
      acceptance has no other instrument. Fold Q5's browser proof in here if one spec serves; say so if
      it cannot and take two.
- [x] `folio-designer/e2e/browser-native-roundtrip.spec.ts` — **Q4b**: stop authoring the document
      through the UI. Load `fixtures/statement-1/input.folio` via the shipped `readFileSync` +
      `setFiles` idiom, keeping the byte-identity claim intact. The fixture is **proven** to carry five
      bound table columns. **This rewritten spec must be OBSERVED GREEN IN A REAL RUN before Epic 14
      closes — report the actual run, not the intent to run it.**

**Acceptance Criteria:**
- Given a table drawn on the canvas with visible columns, when the author clicks a column, then that
  column is the column selection, the owning table remains the component selection, and "Configure
  columns" stays live.
- Given a selected column, when the DATA tab is shown, then the row-scope fields under the table's
  bound collection are pickable and the context bar names the column and what will be bound.
- Given a picked row-scope field, when it is connected, then exactly one `updateTableColumnBinding`
  command is sent, unchanged, and any refusal is presented where the pick was made.
- Given the collection itself or a path outside the table's row scope, when a column is selected, then
  it is not pickable and the panel states the reason **before** the engine would refuse.
- Given the table editor, when it is opened, then the matrix's BOUND FIELD shows the bound field and
  offers no control to edit it.
- Given a column bound this way, when it is undone, then it undoes as one step — asserted, with its
  honest limit stated, and nothing built for it.
- Given a column selection, when Delete or Duplicate is invoked by any of its six routes, then no
  command naming a column id is sent — proved by a test that fails when the guard is removed.

## Spec Change Log

## Design Notes

**Why the column selection is not in `selected`, in one sentence the implementer can act on.** Every
one of the 32 read expressions over `selected`/`selectedRef.current` was written on the premise that a
member is an element id in `projection.components`. Q1(b) keeps that premise true, which is why no
guard is needed at 26 of them and why the six delete/duplicate routes keep working unchanged. The
compound-id arm would have required repairing four working paths — including killing "Configure
columns" outright — to accommodate this story's own state shape.

**Why the pickable set must be *identical* to `tableSampleCandidates`' output rather than merely
similar.** AC4 asks the panel to refuse *before* the engine would. If pickability and the refusal
message are computed from two sets, the panel can say "outside row scope" about a path it also offers,
which is D-000.25's shape. One set, both consumers, asserted against a fixture.

**Why `data-column-id` and not a `role`.** `canvas-table-paint.test.tsx:185` plant-proves that no
inner node may carry `role`, `tabindex`, `aria-label`, a second `data-component-id`, `role="group"` or
`.canvas-box` — each absence verified by adding the forbidden attribute at four positions. That guard
was written by 14.9 naming this story: *"Making a column selectable is Story 14.10's work and its
selection-model change; 14.9 paints only."* `data-column-id` and `pointer-events` sit outside its
list, so the permitted shape and the needed shape are the same shape.

**Why the unit suite cannot prove this story's central mechanism.** jsdom does not implement
`pointer-events` hit-testing, so the unit tests would pass with the CSS change entirely absent. The
source-text pin proves the rule is in the stylesheet; only the browser proves the click resolves to
the column. Neither alone is sufficient and the spec requires both.

## Verification

*This block is the orchestrator's. Run every command **separately** — two commands in one block produce
one undifferentiated stream, and a command matching nothing contributes nothing, so surviving lines
re-attribute to whichever command the reader expects ([D-14.8.5]).*

**Baseline is `c209c74`.** Every figure below was measured there.

**The cadence ([D-000.33]), from `folio-designer/` — never from the repo root, which is a known false
mass-failure:**

1. `npx vitest run` — **77 files / 1409 tests / 0 failures** at baseline. Diff test names as a
   **MULTISET**, never as an integer: a suite that gains three and loses three is not unchanged.
2. `npx tsc -b --force` — exit 0 with **zero bytes** of output. `--force` is not optional; without it
   a stale build-info can short-circuit the typecheck.
3. `npx oxlint` — exit 0, and the warning **SET** unchanged, recorded per file, never as a count
   ([D-14.7.2]). Baseline: `src/App.tsx` ×2, `src/preview/pdf-viewer.tsx` ×2,
   `src/segmented-control.tsx` ×3. Note oxlint prints no total line — those seven lines are the whole
   output.
4. `npm run test:e2e:compile` — exit 0.

**This story's own instrument, and it is not optional.** AC1 says *"the author clicks a column"*, and
**jsdom does not implement `pointer-events` hit-testing** — a simulated click targets the span
regardless of any CSS, while a real browser removes the whole `pointer-events: none` subtree from hit
testing. So a DOM-targeting implementation **passes all 1409 unit tests while being completely inert
in the product**:

5. **A targeted Playwright run** of this story's new spec (single file or `--grep`; separability
   confirmed — `testDir: './e2e'`, no `projects`). **Not** the full 52-test suite. It must be green,
   **and it must go RED when `pointer-events: auto` is deleted — prove that red by running it, do not
   assert it.**
6. **The `App.css` source-text pin**, mutation-proved the same way. This is the second cover, not the
   first: the pin proves the rule exists in the stylesheet, not that the click resolves to the column.
   A later overriding rule, a covering element, or a wrapper that still captures would each leave it
   green.

**Report the cost rather than burying it.** The Playwright `webServer` performs a full cold
`npm run build` first — `scan:font-hosts && scan:host-fonts && build:wasm && tsc -b && vite build &&
build:offline && verify:offline`, ~141s warm, longer cold. That is the price of the only instrument
that can observe this story's central mechanism.

**Two obligations that outlive this story's own gates:**

7. **The rewritten `browser-native-roundtrip.spec.ts` must be OBSERVED GREEN IN A REAL RUN before Epic
   14 closes.** Not compiled, not reasoned about — run, with the output reported. A byte-identity test
   that has been rewritten and not executed is [DW-383]'s shape wearing the same disguise a second
   time. **And prove the fixture it loads still carries bound table columns**, or CAP-13 silently
   stops covering the thing the escalation was about ([D-14.8.4]: a fixture that does not match the
   precondition is a guard that cannot see the defect).
8. **Read `s1.assetCount` off the `npm run build` that step 5 performs and report the actual
   integer.** Expected slot cost of this story: **0**. This discharges [D-14.0.1]'s per-story
   obligation, which **none of 14.6, 14.7, 14.7b, 14.8 or 14.9 kept** — and the recorded margin turned
   out to be wrong by 50% ([D-14.10.4]), so a measurement that is free here is worth taking. Report
   the integer, never a remembered figure.

**Go is not expected to change.** If it does, add, from `folio-go/`:
- `go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` — CI's verbatim invocation
  (`.github/workflows/ci.yml:141`). Never widen or de-anchor that filter.

**What did NOT run, named in these words ([D-000.32]):** the browser suite, the matrix legs,
`npm run build` as a gate, the `verify:offline*` chain, and the font-host scans.

*Two notes on that sentence, so it stays true rather than becoming a recitation.* **A targeted
Playwright run is not "the browser suite"** — the suite is 52 tests and this runs one spec; the other
51 remain unrun locally and CI is what covers them. And **if the Go suite is added under the clause
above, "the Go suites" must NOT appear in that list** — reciting an absence that did not happen is a
false absence claim.

## Suggested Review Order

**The selection model — start here**

- The whole design in one binding: transient column state, keyed on both ids, never in the document.
  [`App.tsx:318`](../../folio-designer/src/App.tsx#L318)

- Re-resolves from the projection every render, so a removed column's selection simply drops.
  [`App.tsx:667`](../../folio-designer/src/App.tsx#L667)

- The seam: the column id is read off the event target, never from a coordinate.
  [`App.tsx:1397`](../../folio-designer/src/App.tsx#L1397)

**Addressability — the mechanism the unit suite cannot see**

- `data-column-id` and nothing else: no role, no tabindex, no aria-label, no second component id.
  [`App.tsx:5141`](../../folio-designer/src/App.tsx#L5141)

- The one CSS declaration the whole story rests on; echoes excluded so repeats stay inert.
  [`App.css:474`](../../folio-designer/src/App.css#L474)

**Row scope — one set, two consumers**

- One walk of the sample; `candidates` unchanged for the editor, `byNode` added for the panel.
  [`App.tsx:679`](../../folio-designer/src/App.tsx#L679)

- Pickability and the refusal reason come from the same set, so neither can contradict the other.
  [`DataPanel.tsx:292`](../../folio-designer/src/DataPanel.tsx#L292)

- The refusal narrows on table id, column id and field together — the key it is documented against.
  [`DataPanel.tsx:87`](../../folio-designer/src/DataPanel.tsx#L87)

- Names what is selected and what will be bound, including the bound-to-nothing case.
  [`DataPanel.tsx:110`](../../folio-designer/src/DataPanel.tsx#L110)

**The removal — D-14.10.1's second half**

- BOUND FIELD keeps the `<output>` and loses the input; `CELL.bound` stays a deliberate hole.
  [`TableEditor.tsx:525`](../../folio-designer/src/TableEditor.tsx#L525)

**Guards — read these before believing any of the above**

- The required guard: six routes, each proved to name the table and never the column.
  [`table-column-binding.test.tsx:422`](../../folio-designer/src/table-column-binding.test.tsx#L422)

- AC5 pinned to the cell being uneditable, not to the removed control's name; proved by planting.
  [`TableEditor.test.tsx:507`](../../folio-designer/src/TableEditor.test.tsx#L507)

- The source-text cover for hit testing; the browser cover is the one that can actually see it.
  [`table-column-binding.test.tsx:459`](../../folio-designer/src/table-column-binding.test.tsx#L459)

- A wire member with no production reader, pinned so a future deletion meets D-14.10.1.
  [`engine-protocol.test.ts:841`](../../folio-designer/src/engine-protocol.test.ts#L841)

**Browser proof and CAP-13 — peripherals last, but the only instruments that see the mechanism**

- Clicks a heading, a body cell and an unbound `Not set` cell in a real browser.
  [`table-column-binding.spec.ts:76`](../../folio-designer/e2e/table-column-binding.spec.ts#L76)

- CAP-13 now loads its document instead of typing it, so a red means reproducibility, not UI.
  [`browser-native-roundtrip.spec.ts:29`](../../folio-designer/e2e/browser-native-roundtrip.spec.ts#L29)

## Delivery Log

### 2026-09-10 — done

Baseline `c209c74`. Shipped as `09a41e4` — *"Bind a table column from the main window, and close the second
way in"*, **17 paths**. **It was already pushed when this close began** — `HEAD` and `origin/main` are both
`f175ccf`, one commit further on — so nothing was amended and **no commit was made here**; this closer makes
no commits at all, and its edits to this file are handed to the orchestrator to commit. `git show --stat`
carries only this story's work: this spec, the tracker, the register, the decision log, `epic-14-context.md`,
eleven designer sources/tests/specs, and one evidence artefact — `evidence/story-6.7-roundtrip-manifest.json`,
regenerated by the CAP-13 spec this story rewrote, whose filename misattributes it to Story 6.7 and which is
registered as [DW-395] rather than renamed here. The span `c209c74..HEAD` holds **two** other commits, both
**record-only** and neither this story's churn, named so a later reader diffing the span is not surprised:
`84304d8`, a 46-line register append filing [DW-389]; and `f175ccf`, the **Epic 14 boundary gate**
([D-14.11.1]) — a register append, a decision-log entry and one new Playwright instrument, no product code.
Subject line matches the project's imperative convention and the required `Co-Authored-By:` trailer is
present on both. On `main`; nothing ahead of upstream. This file's frontmatter already read `status: 'done'`
in the commit — the build loop sets it there and the tracker lags, which is this project's normal shape — so
the only status hop left is the tracker's, and that is the orchestrator's to apply. The tree is deliberately
**not** clean at finish: it holds this file's record edits and nothing else.

**What shipped.** A transient column selection keyed on both the table id and the column id, held outside
the document and re-resolved from the projection every render, so a removed column's selection simply drops.
The column id is read off the event target rather than from a coordinate, which is why the paint needed one
data attribute and one CSS declaration and no new role, tabindex or accessible name. One walk of the sample
now serves two consumers — the editor's candidate list unchanged, a by-node view added for the panel — so
pickability and the refusal reason are computed from the *same* set and cannot contradict each other. The
editor's BOUND FIELD cell keeps its read-only display and loses its input. **No engine capability was added:**
the wire already carried what was needed to identify a column, and nothing was reading it.

**Triage — and the record does not support a tally, which is said here rather than invented.** This spec's
`## Spec Change Log` is **empty**, `review_loop_iteration` is **0**, and the file carries **no enumerated
finding-by-finding triage**, no severity split and no rejection count; this closer neither found one elsewhere
nor reconstructed one. What the record does support: **one review finding**, described below; **six deferrals
filed by this story** — [DW-390] through [DW-395]; and two filed against it before it ran, at its plan gate
and dispatch — [DW-387] (UX-DR25 unmet) and [DW-388] (the "0 selected" ghost panel), both **explicitly not
fixed here**. The register was checked for integrity rather than assumed: **399** `### DW-` headings, **four
pre-existing** duplicated numbers (DW-100, DW-162, DW-238, DW-284) and a maximum of **DW-395**, so
399 − 4 = 395 reconciles, and this story's commit minted **no** duplicate. All six new entries are contiguous
and OPEN, and [DW-391] is the builder registering a weakness in its own story's owner ruling: relocating
CAP-13's authoring leg did not fully preserve it, and no test now authors a full golden document end to end.

**What review caught, and it is the same defect class this story was written to end.** AC5's absence claim —
that nothing in the BOUND FIELD cell is editable — was pinned to the **name of the control that was removed**,
so re-adding an editable control in that cell under a different name left the whole suite green. Removing
something cannot falsify an absence claim. It is now pinned to the **cell**, and proved by **planting seven
forbidden controls in turn** and watching each one red. The same shape appears in the story's own accepted
lint cost, below: a hand-copied expectation is a guard that stops guarding the moment the function diverges
from it.

**The measured gates**, run by the orchestrator at the final tree, each command alone, from `folio-designer/`:

- `npx vitest run` — **78 files / 1436 tests / 0 failures**. Baseline 77 / 1409 → **+1 file, net +27 tests**.
  Diffed as a **MULTISET**, and **GONE was not empty**, which is reported rather than flattened into "no tests
  were lost": **one test name disappeared** — a rename in place, plus **one assertion genuinely dropped**
  because AC5 removed the datalist that fed it, with its coverage re-homed to a stronger test.
- `npx tsc -b --force` — exit 0, **zero bytes** of output. Run with `--force`, so it typechecked rather than
  short-circuiting on a build-info cache.
- `npx oxlint` — exit 0, recorded as a per-file **SET** and never an integer ([D-14.7.2]): `src/App.tsx` **×3**,
  `src/preview/pdf-viewer.tsx` ×2, `src/segmented-control.tsx` ×3. **The set MOVED — `App.tsx` 2 → 3 — and it
  is accepted knowingly, not a regression** ([D-14.10.7]): `tableSampleCandidates` had to be exported so the
  pickable-set test compares against the function's real return instead of a hand-copied list. A ninth warning
  in a category already present eight times is cheaper than a guard that silently stops guarding.
- `npm run test:e2e:compile` — exit 0.
- **A targeted Playwright run**, re-run by the orchestrator at the final tree rather than taken second-hand:
  `e2e/table-column-binding.spec.ts` and `e2e/browser-native-roundtrip.spec.ts` — **4 passed (2.6m)**. This is
  the only instrument that can observe this story's central mechanism: jsdom does not implement
  `pointer-events` hit testing, so an implementation could be **completely inert in the product while green
  across all 1436 unit tests**. The red-proof the spec demanded is delivered in a **standing** form rather than
  as a one-off mutation run: a third test reinstates `pointer-events: none` on exactly the spans the stylesheet
  restores it on and asserts that **no** column resolves while the component still selects — so the pair
  disagrees on the same mechanism on **every** run, not once.

**The two obligations that outlived this story's own gates, both discharged.** The rewritten CAP-13 spec was
**observed green in a real run**, not compiled and not reasoned about — that was [DW-383]'s shape wearing a
second disguise and it was not repeated. Its **fixture precondition is asserted per run, not assumed**
([D-14.8.4]): the prepared statement's five bound table columns are re-read through the shipped dialog on
every execution. An unpredicted consequence worth recording: CAP-13's three renders now all produce
**76,744 bytes / `114df1d6…`**, which **is** the attested golden digest for that fixture, so byte identity is
anchored to the golden corpus instead of to an ad-hoc document unrelated to it. And the slot-cost obligation
([D-14.0.1]) that **none of 14.6, 14.7, 14.7b, 14.8 or 14.9 kept** is discharged with a measured integer
printed by the build itself: **62 assets against a maximum of 64, margin 2**, and **this story adds 0**. That
independently confirms [D-14.10.4] and refutes the withdrawn "margin of 3".

**Suites that deliberately did not run** ([D-000.32]): **the browser suite, the matrix legs, `npm run build`
as a gate, the `verify:offline*` chain, and the font-host scans.** Two nuances, without which that sentence
becomes false. **The targeted Playwright run above is NOT "the browser suite"** — **4** tests ran out of a
suite of **55**, so **51 remain unrun locally** and CI is what covered them. *That population was
re-measured here rather than carried, and it had gone stale by this story's own hand:* 52 is the count at
the baseline `c209c74`, and **this story added 3** (`table-column-binding.spec.ts`), so 55 is the suite at
`09a41e4` and 52 − 4 = 48 would be a baseline-era population subtracted from a post-story run. At `HEAD`
`f175ccf` the suite is **58**, the boundary gate's `document-bar-fit.spec.ts` having added 3 more. The 4
that ran are the 3 new binding tests and CAP-13's **single** test. And **"the Go suites" is
deliberately absent from that list**: **no Go file changed in this story** (`git show --stat 09a41e4 --
folio-go/` returns **0 paths**), so Go was **untouched**, which is a different statement from an unrun suite
and reciting it as one would be a false absence claim.

**CI, attributed to CI and never to the local cadence.** `09a41e4` is **fully green — all seven jobs across
both workflows**, which includes the full browser suite and the matrix legs this cadence did not run.

**⚠ Reported unmet rather than softened ([D-000.17]).** **UX-DR25 is UNMET.** It names table columns
explicitly in its keyboard-reachability list, and a column selected by this story is reachable **by mouse and
by no other means**. That is a stated obligation left unpaid by an **owner ruling** ([D-14.10.3]) — which
scoped one story and explicitly did **not** retire the commitment or find the canvas acceptable — registered
as [DW-387] with AD-17's `[ASSUMPTION]` clause attached, alongside [DW-385]'s finding that the canvas table's
drawing is entirely presentational and reaches assistive technology not at all.

**Record checks, run rather than assumed.** All **16** file:line anchors in `## Suggested Review Order` were
re-resolved at `f175ccf` and every one lands on what it claims; the single note is that the CAP-13 anchor
lands inside the fixture's precondition comment block, about eight lines above the read it describes.
`_bmad-output/` carries **no debris from this story** — no orphaned review prompt, no reverted-implementation
patch, no unresolved result file, no slug-collision duplicate of this spec, and nothing untracked anywhere in
the tree. The one patch-shaped file in the output folder, `8-4j-attempted-implementation.patch`, is
**preserved evidence** cited by two specs and a decision log, and was left alone. Every one of Epic 14's
**11** `sprint-status` story keys resolves to a spec file on disk, and the 11 keys match the 11 story headings
in `epics.md` §Epic 14 exactly. `epic-14-context.md` is **not stale**: it was rewritten in this story's own
commit, after every planning-artifact commit in the repository.
