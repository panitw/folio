---
title: 'Story 14.9: The canvas draws the table it will print'
type: 'feature'
created: '2026-09-10'
status: 'done'
review_loop_iteration: 1
baseline_commit: '40f3bfce5f831efc43d2f557b6ce1e5bdc5cfdd1'
context: []
---

# Story 14.9: The canvas draws the table it will print

**In plain terms.** Put a table on the canvas today and you get a small box with the word "Table" in it.
You cannot see how many columns it has, what they are called, what data they are bound to, or how wide
they are. To learn any of that you have to open the table editor, and to see it laid out you have to
render a preview. This story makes the canvas draw the table instead: a strip along the top naming the
collection it is bound to and how many columns it has, the real column headings underneath, and one
example row below that.

The example row shows each column's binding rather than a value — `{{date}}`, `{{debit}}` — because a
canvas shows you the shape of the document, not its contents, and because one row is enough to show
structure where thirty-four would just be noise. Columns are drawn at the widths they were given and
aligned the way they will print, so a money column set to the right reads as right-aligned before you
render anything.

Two cases get said out loud rather than drawn as an empty frame. A table with no columns yet says so, so
that it does not look like something failed. And a column whose binding has not been set reads as
unbound rather than as a blank cell, so an unfinished table looks unfinished.

The rule this story had to be careful about is that the canvas is not allowed to work out the shape of
anything for itself — every position and size comes from the engine, so that the picture on screen cannot
drift from the document that prints. This is the first time the canvas will draw a piece of text that
actually appears in the PDF, so the conditions under which that is safe are written down here and in the
test that enforces them, not just followed once.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The canvas paints a table as a small box containing the word `Table` (`App.tsx`, two
sites: `CanvasComponent` and `ComponentEcho`). An author cannot see the columns they are building
without opening the table editor or rendering a preview. The canvas projection carries only
`tableBind` — one string — for a table component; the columns live in the separate `table-columns`
projection the editor requests when it opens.

**Approach:** Widen the **canvas** projection so each table component carries its columns — id,
label, width, alignment and binding — resolved in Go, and repaint the table on the canvas as
`Binding.dc.html` draws it: a chip carrying the table glyph, the bound collection in the bind
accent and the column count; the real header labels beneath it; and one representative row whose
cells show each column's binding placeholder. A table with no columns says so; a column with no
binding reads as unbound rather than blank. No format field, no version increment, no rendered byte
changes.

## Boundaries & Constraints

**Always:**
- **AD-17 / AD-15 govern absolutely.** The canvas paints from the engine's own projection and never
  from a browser-side model of the table. Every value drawn — label, width, alignment, binding,
  column count — comes from the projection. The browser performs **no** style cascade, **no**
  width arithmetic beyond the existing `canvasDisplay.css(millipoints, zoom)` scale, and **no**
  measurement of any kind.
- **The AD-17 display-paint test — three conditions, all required, ruled at the plan gate (R1).**
  Text may be painted by the browser on the canvas **only** where: (1) **the rectangle is the
  engine's** — every coordinate and extent bounding the text comes from the projection, and no
  browser measurement contributes to the box; (2) **the browser makes no break decision** — wrapping
  disabled (`white-space: pre`/`nowrap`), overflow clipped or ellipsised **by CSS**, never re-flowed;
  (3) **nothing flows back** — no quantity derived from the painted text (width, height, line count,
  overflow state, scroll extent) reaches geometry, a page count, a band-fit decision, a command, or a
  projection field. If all three hold the paint is display-only and AD-17 permits it **whether or not
  the string also prints**. If any one fails, the text must arrive as pre-measured, pre-broken runs
  from the measure API. **The permitted residue is that a label may clip where the PDF wraps** — a
  text-only inaccuracy, never a geometry error.
- **Go and TypeScript move in one commit.** The component-level guard is `hasOnly` (a subset check:
  it rejects a surplus key and tolerates a missing one), so a Go key the browser does not list makes
  `isCanvas` false → `parseInbound` returns undefined → `engine-client` raises `PROTOCOL_INVALID`,
  **terminates the worker, and there is no respawn path**. The session is dead until reload.
- **⚠ AMENDED 2026-09-10 BY THE ORCHESTRATOR, AFTER APPROVAL, RESOLVING A CONTRADICTION INSIDE THIS BLOCK.**
  As approved, this Always clause and the two projection-bound rows of the I/O matrix gave **opposite answers
  about the same document**, and step-04 proved it by execution: the golden fixture with one 600-byte column
  label parses, renders a real 64,123-byte PDF, and had its canvas projection **abort** — which by this spec's
  own text terminates the worker with no respawn, so a template that prints correctly blanks the designer's
  canvas until reload. **The matrix rows are amended to clip rather than abort; this clause governs.** The
  contradiction was mine: I approved a frozen block containing both readings. Recorded rather than quietly
  rewritten, because a frozen block edited after approval must show what changed and why.
- **The reading is settled by the matrix's own neighbouring row**, not only by this clause: *"Zero or negative
  column width … the canvas paints what it is given and refuses nothing it accepts today — N/A, must not become
  an error."* Same principle, two rows apart, opposite answer. **A document the engine prints is a document the
  canvas draws.**
- **Clip, never omit.** A column dropped from the projection would make the canvas *systematically wrong about
  the document's structure* — the chip's column count would disagree with the header row it sits above — and
  being systematically wrong about structure is precisely the harm AD-17 exists to prevent. Clipping is a
  display concern on a surface that is display-only paint by [R1], and `.canvas-display-paint` already
  ellipsises visually.
- **Clip on a rune boundary.** A byte-cut mid-rune emits invalid UTF-8 and breaks the JSON envelope, turning a
  display concern back into the fatal one this amendment removes. **The browser guard must accept the clipped
  value** — note the projection bounds bytes while the guard bounds UTF-16 length, so a multibyte label is the
  case where those two disagree, and it needs a test.
- **The new field must not newly refuse a document that ships today.** A column with width `0` or a
  negative width loads, projects, and paints now; the browser-side guard must accept any safe
  integer width, never `> 0`.
- Millipoints stay the wire unit; points stay the displayed unit; the conversion stays in the
  browser, in the existing helpers.
- Every colour is a design token. Square corners. 1px borders. Mono for what a machine reads, sans
  for what a person wrote.
- Accessibility floor: the canvas table adds no new interactive element, keeps
  `.canvas-component` as the single `role="button"`, and does not change any accessible name.

**Ask First:**
- Adding, removing or renaming any **format** field, or touching `internal/template/`.
- Any change to `sprint-status.yaml`, `deferred-work.md`, `epics.md`, `DESIGN.md`, or **anything**
  under `fixtures/declared-variants/` (human-attested — never touch, for any reason).
- Making a column, or anything inside the table, selectable or focusable. That is Story 14.10's
  work and its selection-model change; 14.9 paints only.
- Any work toward DW-142 (asset keys in `CanvasProjection`) or DW-145 (`"border": {}` painting
  nothing on the canvas). Both name "whichever story next widens the canvas projection" as owner.
  They are **out of this story's scope** and must not be quietly absorbed.
- Changing `_bmad-output/planning-artifacts/ux-designs/.../mockups/Binding.dc.html`.

**Never:**
- Never call any browser measurement API: `getComputedStyle`, `getBoundingClientRect`,
  `clientWidth`/`clientHeight`, `scrollWidth`/`scrollHeight`, `scrollTop`/`scrollLeft`,
  `offsetWidth`, `measureText`, `ResizeObserver`, `Range`, `getSelection`, `devicePixelRatio` —
  in production, in unit tests, or in `e2e/`, including inside `page.evaluate`.
  `canvas-authority-contract.test.ts` scans all three trees and will red.
- Never name a new CSS class or JSX `className` containing `canvas-text`, and never group a new rule
  with a `.canvas-text*` selector — rules 15/16/17 of that contract forbid `font-weight`/`font-style`
  on that surface in every spelling, and the header labels need weight.
- **Never truncate by measuring.** CSS `text-overflow` is condition-2 compliant; a JS truncation that
  measures the string or its box is not, and there is no computed ellipsis anywhere in this story.
- Never name the display-paint class after this feature. **The orchestrator pins it: `.canvas-display-paint`** — chosen so a later author grepping that class lands on the three-condition rule, and so the next candidate for this exception recognises itself. Do not substitute a feature name. It is named for the **property** —
  display-only paint — so the next author with a candidate for this exception recognises their own
  case in it. `.canvas-table-header-label` would say nothing about the rule it lives under.
- Never add `role="group"` or `role="tablist"` anywhere on the canvas table.
- Never add a second `data-component-id`, a nested `role="button"`, or an `aria-label` on an inner
  canvas node.
- Never use the `.canvas-box` class on a table cell.
- Never add an `@media` query to `App.css`, never a colour literal, never a gradient, never a
  `border-radius` that is not `var(--radius…)`, and never a second use of `var(--type-display)` or
  `var(--type-numeric-lg)`.
- No new format field, no format version increment, no change under `folio-go/internal/template/`.
- No new dependency.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Bound table with columns | table `bind: "transactions[]"`, 5 columns with labels, widths, aligns, binds | canvas draws chip (glyph · `transactions[]` in bind accent · `5 columns`), header labels at declared widths and resolved header alignment, one row of `{{…}}` placeholders in the bind accent at resolved cell alignment | N/A |
| Unbound table | table `bind: ""` | chip reads `Not set` in place of the collection, in the muted page ink, not the bind accent | N/A |
| No columns | `columns: []` | the projection emits no columns member; the canvas states the table has no columns yet in the design's own terms, on the dashed placeholder outline, rather than drawing an empty frame | N/A |
| Column with no binding | `bind: ""` on one column | that cell reads as unbound in the muted page ink — **not** in the bind accent, because amber means data and an unbound cell has none | N/A |
| Column with empty label | `label: ""` (legal; the key is required, the value may be empty) | the header cell is drawn, no label text is painted — matching the renderer, which builds the rect then skips glyphs on an empty label | N/A |
| Single column | 1 column | chip reads `1 column`, singular | N/A |
| Column label over the projection bound | label longer than 512 bytes (**loads and prints today** — `decodeColumn` imposes no cap) | **CLIPPED at the bound, on a rune boundary, and painted.** Never an error, and the column is never omitted: the chip's count and the header row must keep agreeing with the document | N/A — must not become an error |
| Column binding over the projection bound | bind longer than 512 bytes | as above, clipped on a rune boundary, its own column still drawn | N/A — must not become an error |
| Zero or negative column width | `width: 0` / `width: -5` (both load today) | projected verbatim; the browser guard accepts it; the canvas paints what it is given and refuses nothing it accepts today | N/A — must not become an error |
| Non-table component | any `text`/`image`/`line`/`rect` component | the columns member is absent, and the browser guard **rejects** a non-table component that carries it, mirroring the existing `tableBind` cross-clause | guard returns false |
| Echoed table on a later sheet | multi-sheet document | `ComponentEcho` draws the same table body, `aria-hidden`, with no chip duplication of an accessible name | N/A |

</frozen-after-approval>

## Rulings — the plan gate. SETTLED. Carry these; do not re-open them.

Both Open Questions this spec raised were ruled by the engineering lead on 2026-09-10, at the plan
gate, before any code was derived from this spec. **The ruling text below is theirs, quoted. The
surrounding evidence is this spec's.** Where the two differ, the ruling governs.

### R1 — AD-17 permits this paint. The boundary is the DISPLAY-PAINT TEST, not "chrome versus document text"

Ruled **(a)**, and the reasoning is explicitly **not** the one this spec originally proposed. The
"every string the canvas paints today is chrome that never prints" axis was refused, verbatim:

> *"That axis is a coincidence, not the rule, and the spec must not state it as one. The reason every
> string the canvas paints today has been safe was never that it does not print — it is that
> **nothing depends on its measured extent.** Were the rule 'chrome only', it would forbid something
> harmless and permit something dangerous the moment someone painted a chrome string whose width
> drove a layout decision."*

**THE AD-17 DISPLAY-PAINT TEST — three conditions, all required:**

> 1. **The rectangle is the engine's.** Every coordinate and extent bounding the painted text comes
>    from the engine's projection. No browser measurement contributes to the box.
> 2. **The browser makes no break decision.** Wrapping disabled (`white-space: pre`/`nowrap`);
>    overflow clipped or ellipsised **by CSS**, never re-flowed.
> 3. **Nothing flows back.** No quantity derived from the painted text — width, height, line count,
>    overflow state, scroll extent — reaches geometry, a page count, a band-fit decision, a command,
>    or a projection field.
>
> *"If all three hold the paint is display-only and AD-17 permits it whether or not the string also
> prints. If any one fails, the text must arrive as pre-measured, pre-broken runs from the measure
> API."*

**Why this is a scoped permission and not a waiver — the sentence the spec was told to carry:**

> *"Condition 3 is already enforced. `canvas-authority-contract.test.ts` scans `src/` and `e2e/` for
> every route by which a browser-measured quantity could be **obtained at all** —
> `getBoundingClientRect`, `offset*`, `client*`, `scroll*`, `ResizeObserver`, `getComputedStyle`,
> `Range`. **A value that cannot be obtained cannot flow back.**"*

**Binding guardrails, all carried into Tasks:**

- **The three-condition test goes into `canvas-authority-contract.test.ts`'s own comment block**, not
  only into this spec. *"That file's comments have carried scoping rulings correctly across this
  entire run and are the one place in this repo demonstrated to be read years later. **A spec is read
  once.**"*
- **Name the new class for the PROPERTY, not the feature** — a later author must read it as
  *display-only paint* and recognise their own case in it. *"`.canvas-table-header-label` tells them
  nothing about the rule it lives under."*
- **Bound the precedent in that same comment:** any future use that fails condition 1 or 2
  mechanically, or where condition 3 is *asserted* rather than covered by the scan, is a **new
  ruling** and is not carried by this one.
- **No computed ellipsis.** CSS `text-overflow` is condition-2 compliant; a JS truncation that
  measures is not.
- The permitted residue is exactly what `epics.md`'s Story 5.13 AC already scopes: **a label may clip
  where the PDF wraps.** That is the allowed text-only inaccuracy, not a geometry error.

**(b) — routing header labels through the measure API — was refused on scope, not merit:** *"it
exceeds AC4's frozen field list, and had the story genuinely needed it that would have been an owner
escalation rather than a lead ruling. It does not need it."*

### R2 — Two resolved alignments per column, and that is a FORCED CONSEQUENCE of AC1 + AC3, not an expansion

Ruled **(a)**: `headerAlign` and `cellAlign` per column, each produced by the renderer's **own
cascade functions, called — not re-implemented beside them.**

The ruling line, to be quoted at review so this is not re-litigated:

> *"AC1 has this story draw **two rows**, and the engine resolves alignment differently for each. A
> field list naming 'alignment' for a story that draws both rows means the alignment of **each**; the
> reading under which it means one value is the reading that makes AC3 false, **and a reading that
> falsifies another criterion is the wrong reading.**"*

> *"(b) and (c) are refused because each makes AC3 false in a **reachable** document — (b) on the
> header row whenever `headerStyle.align` differs, which 14.8 made authorable from the shipped UI,
> and (c) whenever the table declares `style.align` and a column does not, which is the ordinary
> authoring shape and AC3's own money column. **(d) is refused on principle:** it puts the cascade in
> the browser, which AD-15 and AD-17 jointly exist to prevent, and it would be wrong in a way no test
> on either side would see."*

**Binding guardrails, all carried into Tasks:**

- **Call `resolveHeaderStyle` / `resolveBodyStyle`.** *"This is 14.8's Part 4 requirement again — one
  source shared with the renderer. A projection that mirrors the cascade's logic will drift, and the
  failure mode is a canvas that lies about print while every test passes."*
- **A test in which the two values DIFFER.** Set `headerStyle.align` to something other than
  `style.align` and assert two different values. *"Without it both keys could be populated from the
  same cascade and every assertion would still pass — the both-sides-move-together shape, and the
  exact case that made (b) wrong."*
- **Assert which row CONSUMES which key**, not merely that both exist. **Swapping `headerAlign` and
  `cellAlign` must red something.** *"A presence assertion cannot see a swap."*

### R3 — The four register entries, confirmed as proposed

- **DW-74 closes as this story's guard.** *"Pinning all 31 existing component keys is **not** a side
  effect to apologise for — that is what a key-set record *is*, and a record pinning only the new key
  would be the denylist shape this run refuses. It is also the **only** artefact that can prove this
  change, since the projection is byte-invisible by construction."*
- **DW-104:** probes for this story's two new bounded sites only. The pre-existing eight-vs-nine gap
  is out of fence and stays open.
- **DW-142 and DW-145:** neither touched. Their non-closure is a recorded decision, not an oversight.

### R4 — Confirmed without change

- The token gate is **[K]**, pre-answered by the standing instruction; the spec is not split.
- This spec's `## Design Notes` reconciliation of D-13.4.1 against AC3 was **verified and accepted**:
  D-13.4.1 does not carry the "approximate representation" clause, the text-only scoping lives in
  `epics.md`'s Story 5.13 AC, and *"the Q1 ruling above leans on exactly that scoping, so the Design
  Notes reconciliation and the ruling are consistent by construction."* It stays where it is.
- The claim that **nothing byte-level can prove this change** was independently verified —
  `CanvasProjection` appears **zero** times in `serialize.go` and in `internal/pdf`. **DW-383's
  remedy is for a format field and would have been reflexive here.**

## Code Map

All anchors measured by me at `06c279a` on 2026-09-10, tree clean. **HEAD has since moved to `40f3bfc` (this story's `baseline_commit`) and the anchors still hold: `git diff --name-only 06c279a..40f3bfc` is three paths, all under `_bmad-output/implementation-artifacts/` (this spec, `deferred-work.md`, `sprint-status.yaml`) — no file under `folio-go/` or `folio-designer/` moved, and the five baseline suite measurements below were taken before that and are unaffected.** **Anchor rot is endemic in this
epic — `App.tsx` and `TableEditor.tsx` have both moved repeatedly — so find every site by the named
identifier and re-measure before editing. Nothing here was carried from the epic context or a
register without checking; three claims that arrived from investigation were measured wrong and are
corrected below.**

### Go — the producer (`folio-go/page_setup.go`, package `folio`)

- `type CanvasComponent struct` — 31 json keys. `TableBind *string \`json:"tableBind,omitempty"\`` is
  the only table-aware member. Every optional member is a pointer or a slice with `omitempty`;
  the first eight (`id`…`resizable`) carry none, and that is enforced (see the wire record below).
- `func canvasComponents(t *Template, bands []CanvasBand) ([]CanvasComponent, error)` — the loop to
  edit. The site, verbatim:
  ```go
  if element.Type == template.ElementTable && element.Table.Set && !element.Table.Null {
      if len(element.Table.Value.Bind) > maxCanvasPropertyString {
          return nil, fmt.Errorf("folio: component table bind exceeds the projection bound")
      }
      component.TableBind = stringPointer(element.Table.Value.Bind)
  }
  ```
  **`element.Table.Value.Columns` — the full `[]template.Column` — is in scope at that line**, and is
  already read two calls earlier by `projectedSize`. `element.Style` is in scope too (applied just
  below by `applyCanvasStyle`). Nothing needs threading in.
- `const maxCanvasPropertyString = 512`. Its comment says *"Seven sites, all legitimately short, and
  all of them still ABORT the projection"* — that comment's count moves if this story adds sites.
- `func applyCanvasStyle(component *CanvasComponent, elementType, style)` — sets `component.Align`
  from a **table's** own `style.align` when Set and non-null. So the canvas already carries half the
  body cascade and none of the header one; `headerStyle.align` is not projected anywhere today.
- `func Canvas(t *Template)` → `func CanvasWithTextPaint(t, fs)` is the only path the browser
  reaches (three call sites in `folio-go/wasm/engine.go`). `wasm/engine.go` returns
  `Canvas: e.canvas` **whole** — no hand-enumeration on the Go side.
- Precedents for a nested projection object: `CanvasTextPaint`, `CanvasImagePaint` — both `*T` with
  `omitempty`, each with its own TS guard.

### Go — where to source the four values (read-only; reuse the derivations, **not** the gate)

- `folio-go/internal/template/model.go`, `type Column`: `ID ElementID`, `Label string`,
  `Width geom.Length` (int64 **millipoints**), `Align Presence[string]`, `Bind string`, plus footer
  members and `Extra`. **No json struct tags** — `internal/template` hand-decodes and hand-serializes.
  `Label`, `Width` and `Bind` are **required keys whose values may be empty/zero**, so `""` means
  "declared empty", not "absent"; `Align` is the only three-state member, and `"align": null` is a
  load error, so in practice it is absent or Set-and-non-null.
- `folio-go/table_columns_projection.go`, `func TableColumns(t, tableID)` — the derivations to copy:
  `Header: column.Label`, `Width: int64(column.Width)` (**millipoints, unconverted**),
  `Binding: column.Bind`, each bounded at 256. ⚠ **Do not reuse this function or its gate.** It is a
  *validating* projection that hard-errors on `width <= 0`, on >128 columns, and on a bind that fails
  `rootCollectionPath` — all of which the canvas tolerates today. A canvas projection error blanks
  the whole designer, so reusing the gate would newly kill documents that currently paint.
  Its `align` derivation is `align := "left"; if column.Align.Set && !column.Align.Null { align = … }`
  — a hardcoded fallback, **not** the renderer's cascade. See Open Question Q2.
- `folio-go/table_render.go`: `func resolveHeaderStyle(el template.Element) resolvedHeaderStyle` —
  `alignFallback` starts `"left"`, then `headerStyle.align` wins, else `style.align`.
  `func resolveBodyStyle(el)` — `alignFallback` starts `"left"`, then `style.align` only. Both are
  pure, need no fonts and no data, and are callable from `page_setup.go` (same package). The header
  cell site reads `align := hs.alignFallback; if col.Align.Set && !col.Align.Null { align = col.Align.Value }`.
- **An empty label at render time:** the header loop builds the cell rect, then `if col.Label == "" { continue }`
  — rect drawn, no glyphs. The *"a column label but no style.fontFamily"* render error fires only
  for a non-empty label, after that skip. The canvas should match: paint the cell, paint no text.
- `func projectedSize(element template.Element)` in `folio-go/component_commands.go` — a table's
  canvas `width` **is** `Σ column.Width` and its `height` is `headerHeight`. Nothing stores a table
  width; `internal/template/parse_bands.go` refuses the keys outright (*"a table declares x and y
  only — never width"*). `layout.TableGeometry.Width()` recomputes the sum on every call.
- Closed sets in `folio-go/internal/template/closedsets.go`, **partitioned by consumer**:
  `ColumnAlignTokens` = {left, center, right} for `columns[].align`; `TableStyleAlignTokens` = the same
  three for a table's `style.align` and `headerStyle.align`; `StyleAlignTokens` = those plus `justify`
  for a **non-table** element. So a resolved column alignment is always one of three — `justify`
  cannot reach it, and a justified table is refused at load.
- `columns: []` is legal at load, renders as nothing (`tableDrawsColumns` false), and projects a
  component with `width: 0` today, which the browser guard accepts.

### TypeScript — the guard (`folio-designer/src/engine-protocol.ts`)

- `const hasOnly = (value, keys) => Object.keys(value).every((key) => keys.includes(key))` — a
  **subset** check: rejects a surplus key, tolerates a missing one.
- `const hasExactKeys = (value, keys) => Object.keys(value).length === keys.length && keys.every(…)`
  — rejects both.
- `isCanvas`'s component clause uses **`hasOnly(component, [...31 keys])`**: `id, type, band, x, y,
  width, height, resizable, value, binding, visibleIf, fontFamily, fontSize, lineSpacing, bold,
  italic, align, valign, color, background, borderWidth, borderColor, borderEdges, paddingTop,
  paddingRight, paddingBottom, paddingLeft, tableBind, textPaint, image, imageUnavailable`. Measured:
  31, not 30 and not 32.
- ⚠ **Ordering consequence, and it is asymmetric.** Because the component level is `hasOnly`,
  **TS-then-Go is safe** (an extra name in the list changes nothing) and **Go-then-TS is fatal**.
  Contrast `fontChains` / `isFontChainEntry` / `isTableColumns`'s table object, which are
  `hasExactKeys` and must move together in both directions.
- The cross-clause to mirror: `if (component.type !== 'table' && component.tableBind !== undefined) return false`.
- The nested-record-with-closed-set precedent to copy is `isTableColumns`'s per-column clause:
  `Array.isArray && length <= 128 && every(isRecord && hasExactKeys(column, [...10]) &&
  ['left','center','right'].includes(column.align) && ['','sum','avg','count'].includes(column.footer))`.
- The failure path, from the file's own comment: *"isCanvas false -> parseInbound undefined ->
  PROTOCOL_INVALID -> engine-client terminates the worker and the canvas is permanently blank"*, and
  `#fail` in `engine-client.ts` does `this.#detach(); this.worker.terminate()` with **no respawn**.
- **Optional vs required on the TS type.** `tsconfig.app.json` includes `src`, so test files are
  typechecked, and six test files hand-author `type: 'table'` canvas literals. A **required** new
  member breaks all six; an optional one (`?`, like `tableBind`) breaks none.

### TypeScript — the transport (measured, and this corrects a warning I was given)

- `engine-client.ts` `#settle`: `const snapshot = deepFreeze({ ...message.snapshot }) as EngineSnapshot`
  — **the canvas rides a spread and is not re-enumerated, so no `#settle` change is needed.**
  `deepFreeze` mutates in place and rebuilds nothing. The hand-enumerated arms in the same function
  are `preview` and `parameterReferences`; `tableColumns` was fixed to spread by 12.3. Do not read
  13.3's comment about "the second hand-enumerated hop" as covering the canvas.
- `engine.worker.ts` posts `snapshot: result.snapshot` whole. No worker change.

### TypeScript / CSS — the painter (`folio-designer/src/App.tsx`, `App.css`)

- **Two sites paint a table, and both are one-line ternaries inside long JSX returns.** Find them by
  the string `component.type === 'table' ? 'Table' : ''`, not by line:
  - `function CanvasComponent({ component, carriedFaces, origin, note, limit, zoom, selected, preview, engine, generation, trackColumn, onSelect, onDelete, onDragStart, onDragEnd })` — the interactive one.
    Root: `<div className={\`canvas-component canvas-component-${component.type}…\`} aria-label={componentAccessibleName(component, note)} role="button" tabIndex={0} data-component-id={component.id} …>`
    then `<ComponentBox/>`, an optional `.canvas-text-truncated` notice, the content ternary, then the
    `.canvas-dimension` badge and resize anchors.
  - `function ComponentEcho({ component, carriedFaces, y, zoom, engine, generation })` — the
    `aria-hidden` repeat on later sheets, same ternary.
- ⚠ **`trackColumn` on `CanvasComponent` is NOT about table columns.** It is the sheet-stack's
  content-column drag inverse (`columnEdgeAfterDrag`). Do not conflate.
- `function points(value: number): string` and `export const canvasDisplay` (`.css(millipoints, zoom)`)
  — the only millipoints→display conversion, and the one to use for column track widths.
- `paletteGlyphs.table` is `<><path d="M2.5 3.5h11v9h-11z"/><path d="M2.5 6.5h11"/><path d="M6.5 6.5v6"/><path d="M10 6.5v6"/></>`
  — the same shape the mockup draws for the chip glyph. Reuse it; do not hand-copy a second path.
- `function ImagePlaceholder({ children }: { children: string })` — **the shipped empty-state idiom to
  match**: `<span className="canvas-image-placeholder" aria-hidden="true"><svg className="canvas-placeholder-icon" viewBox="0 0 16 16" …>{paletteGlyphs.image}</svg><span>{children}</span></span>`,
  with `App.css` flipping the component's outline from dotted to dashed for exactly this state:
  `.canvas-component:not(.canvas-component-selected):has(.canvas-image-placeholder)::after { border-style: dashed; border-color: var(--color-page-outline-dash); }`.
  That is DESIGN.md's *"Dashed grey on page | A placeholder with no content yet"* already implemented.
- `function componentAccessibleName(component, note)` returns `` `${type} component ${id}` `` for
  anything non-`text`, delivered as `aria-label`. `aria-label` wins over name-from-content, so the
  ~15 tests that locate a table by `getByRole('button', { name: 'table component e7' })` are
  indifferent to the markup inside — **unless** a nested `role="button"` or a second `aria-label`
  appears.
- `.canvas-text-expression { color: var(--color-bind-on-page) }` is the existing on-page amber for a
  `{{…}}` placeholder. There is **no** `.canvas-component-table` rule in `App.css` today.

### The design source

- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Binding.dc.html`, the table
  drawing (find it by the `5 columns` string; 100% inline styles, no `class` attributes). Chip:
  `display:flex; align-items:center; gap:6px; background:#f1f4f7; border-bottom:1px solid #d5dbe2; padding:3px 5px`,
  a 10×10 `stroke-width="1.4"` `#6b7480` table glyph, `transactions[]` in mono 8px `#a8801f`, a
  `flex-grow:1` spacer, `5 columns` in mono 8px `#6b7480`. Grid:
  `grid-template-columns: 58px 1fr 62px 62px 70px`. Header cells: `padding:4px 5px; font-size:8px;
  font-weight:600; color:#1c2530; border-bottom:1px solid #1c2530`, with `text-align:right` on the
  three money columns. Row cells: same padding, `font-family:'IBM Plex Mono'`, `font-size:8px`,
  `color:#a8801f`, matching `text-align`.
- ⚠ **The mockup's pixel tracks are NOT the declared widths.** `TableEditor.dc.html` declares
  22.0 / 64.0 / 28.0 / 28.0 / 32.0 while `Binding.dc.html` draws 58 / 1fr / 62 / 62 / 70.
  **Take widths from the projection. Never from these numbers.**
- **Every hex in that drawing already has a token** (`folio-designer/src/tokens.css`): `#cdd4db`
  → `--color-page-outline-dot`, `#f1f4f7` → `--color-page-thead`, `#d5dbe2` →
  `--color-page-thead-line`, `#6b7480` → `--color-page-ink-muted`, `#a8801f` →
  `--color-bind-on-page`, `#1c2530` → `--color-page-ink`. **`--color-page-thead` and
  `--color-page-thead-line` are defined and referenced by nothing in `App.css` today** — they were
  minted for this table. **No new token is needed, and none should be added:** `design-contract.test.ts`
  asserts token-**name** equality between `designTokenSets` and `DESIGN.md`, and `DESIGN.md` is
  read-only for this story.
- The 8px/600 sans header label has **no page-ramp token** (`--type-page-body` is 400 weight and
  leads with `--font-page`; `--type-page-eyebrow` carries tracking). Write the weight as a literal
  `font-weight: 600` on a `canvas-table-*` class — legal, since rules 15/16/17 are scoped to the
  `.canvas-text*` surface — and say in a comment that the page ramp has no 600 face.
- `DESIGN.md` (normative; it says *"Where this file and any mockup disagree, this file wins"*) has
  **no canvas-table component at all** — its component list ends at "matrix row … the table editor's
  unit". So the mockup is the only source of shape and DESIGN.md governs only its tokens and
  grammar: *"amber means data, and only data. … It never marks selection and never marks structure"*;
  *"State the reason next to anything disabled"*; *"Dashed grey on page | A placeholder with no
  content yet"*.
- `Main.dc.html` draws the same table as a **superset** — identical colours, sizes and paddings, plus
  `5 columns · header repeats`, a per-cell `1px solid #f0f2f5` hairline, an italic row-scope note and
  a footer aggregate row. **Those four additions are outside 14.9's ACs.** Do not build them.
- Nothing reads `.dc.html` programmatically except `e2e/font-embed-boundary.spec.ts`, which
  regex-extracts one chevron shape out of `Font Browser.dc.html`. Every other citation is prose.
- **The empty-columns wording and the unbound-cell treatment are drawn nowhere.** Match, do not
  invent: `TableEditor.tsx` already says `No columns yet. Add a column to start the matrix.`;
  `App.tsx`'s honest note already says `Table binding: {table.tableBind ?? 'Not set'}`, so **`Not set`
  is the shipped word for an unset binding**; EXPERIENCE.md's voice example is `No binding for
  transactions[].amount`.

### The guard census — what reds, what goes vacuous, what is indifferent

**Will red if the change is wrong, and is the point:**
- `folio-go/canvas_projection_wire_test.go` — four test functions, five recorded key sets:
  `TestCanvasProjectionWireKeysAreTheRecordedSet` (Go side, from marshalled bytes, plus a
  zero-value identity that bans `omitempty` at the top level),
  `TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts` (TS side, by four regexes over
  `engine-protocol.ts`, `t.Fatal` — not skip — if a list cannot be found), and the two
  `TableColumnsProjection` twins. ⚠ **None of them descends to the component level**, and
  `canvasGuardKeyList`'s `(?s)const isCanvas = .*?hasOnly\(value, \[(.*?)\]\)` is non-greedy, so it
  stops at the **top-level** list and never reaches `hasOnly(component, [`. **This is DW-74, and this
  story is the first to walk into it.** The record to copy in shape is the **fragment** one
  (`canvasTextFragmentWireKeys` with per-population exact sets), not the chain one — because the new
  member is `omitempty` and absent on non-tables, a zero-vs-projected identity would not hold.
- ⚠ **A component-level record built on today's canvas fixture would be vacuous.** The four existing
  records project `canvasWindowCountControlTemplateJSON`, which **has no table at all** (measured:
  0 occurrences of `table` in its block; positive control: 3 `"type": "table"` hits in the whole
  file). Table-bearing templates that do exist:
  `canvasWindowCountBoundTableTemplateJSON`, `canvasWindowCountOneColumnTableTemplateJSON`,
  `canvasWindowCountColumnlessTableTemplateJSON`, and
  `folio-go/testdata/template/golden/worked-example.json` (element `e2`, `bind: "transactions[]"`,
  2 columns, the first `{"align":"left","bind":"{{transaction.date}}","id":"e3","label":"Date","width":80}`).
  The record must measure against a table-bearing document, with a `t.Fatal` fixture precondition in
  the shape the fragment record already uses, so it cannot pass vacuously.
- `folio-go/component_properties_test.go` — the `tableBind` behavioural precedent to copy:
  `if component.Type == "table" && component.TableBind == nil { t.Fatalf("table bind missing from display-only projection: %#v", component) }`,
  guarded by a `probed == 0` precondition so it cannot pass vacuously, plus a
  `maxCanvasPropertyString+1` over-bound probe.
- `folio-go/canvas_body_text_bounds_test.go`, `TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve`
  — a **hand-list of probes**, correct *"by EXCEPTION, not by exhaustion"* per its own comment. Two
  new bounded string sites need two new probes here or the new refusals are unguarded. This is DW-104.

**Will red if 14.9 chooses the wrong markup — measured, with the exact mechanism:**
- `folio-designer/src/control-vocabulary-contract.test.tsx`. `groupsIn` is exactly
  `root.querySelectorAll('[role="group"], [role="tablist"]')`; `controlsIn` is exactly
  `root.querySelectorAll('button, [role="button"]')`.
  - `const GROUP_INSTANCE_FLOOR = 33`, and the shrink proof pins
    `expect(shrunkGroups, 'Story 14.8 regrouped the table editor with headings, not groups — …').toBe(32)`
    plus `.toBeLessThan(GROUP_INSTANCE_FLOOR)`. **Six of its seven states mount a table on the
    canvas, so one `role="group"` on the table adds ≥6 and reds both, and then the pinned R0 list
    too.** ⚠ Correcting the brief: the four floors (`PER_STATE_CONTROL_FLOOR 15`, `CONTROL_FLOOR 220`,
    `CLASS_FAMILY_FLOOR 18`, `GROUP_INSTANCE_FLOOR 33`) are one-sided `<` comparisons — **adding
    never reds a floor**; the equality in the shrink proof is what reds.
  - `treatmentOf(control)` returns `'word'` when `visibleText` (which strips `aria-hidden` subtrees)
    matches a letter or digit, else `'glyph'` if the control contains an `<svg>`. The table is
    `'word'` today because `'Table'` is a bare, non-hidden text node. ⚠ **If 14.9 wraps the whole
    table paint in `aria-hidden` and puts the chip's `<svg>` inside it, the treatment flips to
    `'glyph'` and the pinned `V2_CENSUS` moves.** Keep at least one non-`aria-hidden` letter or digit
    inside `.canvas-component` for a table.
  - `CHECKED_GROUPS` and `UNDER_ARITY_GROUPS` are closed sets — any *new named group* reds even at
    arity zero.
  - Its `tableComponent` fixture is `{ id: 'e7', type: 'table', …, resizable: false }` — **no
    columns, no `tableBind`** — so this file exercises 14.9's **empty-columns notice** in six states
    on every run, whether the author intended it or not.
- `folio-designer/src/canvas-authority-contract.test.ts` — 17 prohibited patterns over **production
  `src/**` + unit tests + `e2e/**`**, plain regex over whole file text, so `page.evaluate` bodies are
  scanned identically. All six of `getComputedStyle`, `clientWidth`, `clientHeight`, `scrollWidth`,
  `scrollTop`, `scrollLeft` are named. `boundingBox()` and `mouse.wheel()` are named by nothing.
  Rules 15/16/17 are `/\.canvas-text[^{}]*\{[^}]*\bfont-(?:weight|style)\s*:/` and its shorthand and
  JSX twins — note `[^{}]*` reaches across a **grouped** selector, so never group a new rule with a
  `.canvas-text*` one. The CSS wrap rule bans `text-align: justify` (and `normal|wrap|balance|
  pretty|anywhere|break-word|<digits>` on the wrap properties); `left|center|right` and camelCase
  JSX `textAlign` are outside it, and `white-space: pre` is the sanctioned idiom.
  ⚠ Correcting the brief in both directions: the floors are **both** kinds —
  `toBeGreaterThan(10)/(10)/(3)` early and `toBeGreaterThanOrEqual(58)/(51)/(15)/(8 .tsx)/(3 .css)`
  in the recorded-counts block. All one-sided: **adding files never reds; deleting can.** Production
  `.css` sits exactly on its floor of 3, so do not merge or delete a stylesheet.
  It also pins `App.css`'s media queries to exactly `['prefers-reduced-motion: reduce']`.
- `folio-designer/src/design-contract.test.ts` — `App.css` must contain **zero** colour literals
  (`/#[0-9a-f]{3,8}|\b(?:rgb|hsl)\(/i`, and it scans comments too), no gradient, every
  `border-radius` a `var(--radius…)`, no `--color-page-*` inside a `.canvas-region` selector, and
  **exactly one** use each of `var(--type-display)` and `var(--type-numeric-lg)`. Adding
  `canvas-table-*` rules is otherwise fine. Its first test compares `designTokenSets` group-by-group
  against `DESIGN.md`, so **do not mint a token**.
- `folio-designer/src/engine-bounds-mirror.test.ts` — pins six whole `engine-protocol` import lines
  byte-exactly, including `TableEditor.tsx`'s `^import \{ MAX_ENGINE_HISTORY_ENTRIES, type TableColumns \} from '\./engine-protocol'$`
  and `DataPanel.tsx`'s. **Those two files cannot gain an import specifier.** `App.tsx`'s import line
  is **not** pinned, so a new type may be imported there. It also pins the verbatim clause
  `if (!SCALAR_BINDING_COMPONENT_TYPES.includes(component.type as CanvasComponentType) && component.binding !== undefined) return false`
  inside `isCanvas` — do not restructure the function around it.
- `folio-designer/src/engine-ownership-contract.test.ts` — reds if any type or object literal declares
  ≥2 of `{version, page, bands, elements, assets}`. A column type of `{id,label,width,align,binding}`
  is safe.

**Indifferent — checked so nobody chases them:**
- All four format fences (`drift_test.go`, `numeric_classification_test.go`, `fixtures_test.go`,
  `goldenfixture_test.go`) live in `internal/template` and are about the **format** and
  `folio-format.md`. A projection-only field is outside their universe. `drift_test.go` notes it
  *"does not use encoding/json struct tags at all"*.
- `App.test.tsx` (~15 table sites) and the four e2e specs that name a table — all locate it by
  accessible name, which does not change.
- `e2e/e9-5-border-no-ink.spec.ts` enumerates every `.canvas-box` and pins the result exactly. Its
  fixture has no table, so it is safe — **provided cell borders are not painted with `.canvas-box`
  and no inner node gains an `aria-label`.**
- No screenshot or snapshot test exists anywhere (`toHaveScreenshot`/`toMatchSnapshot`: 0 hits), and
  nothing anywhere asserts the string `'Table'` on the canvas (positive control: the same grep finds
  `paletteItems`, the two paint sites, and `DataPanel.tsx`'s `kindNoun`).

## Tasks & Acceptance

> **You must never commit.** You may not run `git commit`, `git add`, `git stash`, `git checkout`,
> `git reset`, `git revert`, `git restore`, `git clean`, `git branch`, `git merge`, `git rebase`,
> `git push`, or `git tag`. The orchestrator makes every commit in this run. Reading git state —
> `git log`, `git status`, `git show`, `git diff`, `git ls-files` — is permitted and encouraged. The
> single carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`,
> removed in a `finally`, is permitted for a throwaway test harness.

**That prohibition binds every subagent too. Never create a branch. Never push. Work on `main` and
leave the tree dirty for the orchestrator.**

**Do not write to `~/.claude/` or any memory directory. Story knowledge belongs in this file.**

**Execution:**

- [x] `folio-designer/src/engine-protocol.ts` -- add the per-column type and widen the guard
  **first**, before the Go change -- the component guard is `hasOnly`, so TS-then-Go is safe and
  Go-then-TS terminates the worker. Add the member to the `CanvasProjection['components']` type as
  **optional**, add its name to the `hasOnly(component, [...])` list, add a nested per-column clause
  in the shape of `isTableColumns`' (array, length bound, `hasExactKeys` per column, closed set for
  each alignment, safe-integer width with **no `> 0` requirement**), and add the cross-clause twin of
  `component.type !== 'table' && component.tableBind !== undefined`.
- [x] `folio-go/page_setup.go` -- add the per-column projection struct and the `CanvasComponent`
  member (`omitempty`), and populate it inside `canvasComponents`' existing
  `element.Type == template.ElementTable` block, beside the `TableBind` assignment -- the columns are
  already in scope there. ~~Bound the label and the binding at `maxCanvasPropertyString` with their own
  named errors, matching the seven existing identifier sites~~ **— SUPERSEDED by the frozen amendment
  (Spec Change Log 1). CLIP both at `maxCanvasPropertyString`, on a rune boundary; never return an
  error and never omit the column.** Update that constant's site-count comment, which must now
  distinguish the seven sites that ABORT from the two that CLIP. **Per R2, carry TWO resolved alignments per column, `headerAlign` and `cellAlign`,
  each obtained by CALLING `resolveHeaderStyle` / `resolveBodyStyle` from `table_render.go`** --
  never by re-implementing or mirroring their cascades beside them (14.8's Part 4 requirement again:
  one source shared with the renderer; a mirrored cascade drifts, and the failure mode is a canvas
  that lies about print while every test passes). Never reuse `TableColumns`' validating gate.
- [x] `folio-go/canvas_projection_wire_test.go` -- add the missing **component-level** record: a
  `canvasComponentWireKeys` literal asserted against the marshalled bytes of a **table-bearing**
  document, and a `canvasGuardComponentKeyList` regex anchored on `hasOnly(component, [` compared
  against `engine-protocol.ts`. Give it a `t.Fatal` fixture precondition so it cannot pass
  vacuously on a document with no table. This closes DW-74 and is the only guard that would catch a
  Go/TS drift on this field before it kills the worker.
  **The record pins the whole accepted component key set — all 31 existing keys plus the new one —
  and that is its purpose, not a side effect.** A record pinning only the new key would be a denylist,
  which is the shape this project refuses: it would go green on the next field someone adds. Since
  the canvas projection is byte-invisible by construction, this record is also the **only** artefact
  in the repository that can prove this change at the seam.
- [x] `folio-go/canvas_body_text_bounds_test.go` -- ~~add one probe per new bounded string site
  (column label, column binding) to `TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve`,
  each asserting its own message~~ **— SUPERSEDED by the frozen amendment (Spec Change Log 1). Those
  two sites no longer refuse, so a probe asserting refusal there would assert a defect.** The
  coverage MOVES rather than disappearing: assert CLIPPING for both, in that same function, and make
  the function's site-count comment name clipping as a third category on the same constant. The
  probe table returns to its original **eight** entries. Do not touch the pre-existing probe gap
  (DW-104's wider scope).
- [x] `folio-go/` (a Go test file, new or existing beside the canvas projection tests) -- assert the
  projected columns behaviourally, in the shape of `component_properties_test.go`'s `tableBind`
  check: labels, widths in millipoints, both resolved alignments, and bindings for a table-bearing
  document; the member **absent** for a table with `columns: []`; the member absent for every
  non-table component; a `width: 0` column projected rather than refused. Every assertion needs a
  `probed == 0`-style precondition so it cannot pass vacuously.
  **R2 requires two specific tests here, and neither is optional:**
  (i) **a document in which the two alignments DIFFER** -- set `headerStyle.align` to something other
  than `style.align`, on a column that declares no `align` of its own, and assert `headerAlign` and
  `cellAlign` come back **different**. Without it, both keys could be populated from the same cascade
  and every other assertion would still pass -- the both-sides-move-together shape, and the exact
  case that made option (b) wrong.
  (ii) **an assertion of which row CONSUMES which key.** Presence is not enough: **swapping
  `headerAlign` and `cellAlign` must red something.** A presence assertion cannot see a swap.
- [x] `folio-designer/src/canvas-authority-contract.test.ts` -- **write R1's three-condition
  display-paint test into that file's own comment block**, in the shape its existing scoping comments
  use. It must state the three conditions; state that condition 3 is enforced by this very file,
  because it scans `src/` and `e2e/` for every route by which a browser-measured quantity could be
  **obtained at all**, and *a value that cannot be obtained cannot flow back*; name the display-paint
  class as the marker for paint claiming the exception; and **bound the precedent**: any future use
  that fails condition 1 or 2 mechanically, or where condition 3 is *asserted* rather than covered by
  this scan, is a **new ruling** and is not carried by this one. **This is not documentation, it is
  the deliverable** -- the ruling's own words: *"that file's comments have carried scoping rulings
  correctly across this entire run and are the one place in this repo demonstrated to be read years
  later. A spec is read once."* Do not put the test only in this spec.
- [x] `folio-designer/src/App.tsx` -- replace `component.type === 'table' ? 'Table' : ''` at **both**
  sites (`CanvasComponent` and `ComponentEcho`) with one new non-interactive painter that draws the
  chip, the header labels and one representative row from the projection, using `canvasDisplay.css`
  for track widths and `paletteGlyphs.table` for the glyph. **Every element painting text under R1's
  exception carries the display-paint marker class, named for the PROPERTY and not for this feature**
  (and containing no `canvas-text` substring). No `role`, no `tabIndex`, no `data-component-id`, no
  inner `aria-label`. Keep at least one non-`aria-hidden` letter or digit inside `.canvas-component`
  so `treatmentOf` stays `'word'`. Draw the no-columns state in the `ImagePlaceholder` idiom.
  **The header row must consume `headerAlign` and the representative row `cellAlign`** -- not one
  value used twice.
  ⚠ **The no-columns notice must NOT be a `<button>` and must NOT carry `role="group"`.** Meet the
  constraint deliberately rather than discovering it as a failure: `control-vocabulary-contract.test.tsx`'s
  `tableComponent` fixture carries no columns and no `tableBind`, and **six of its seven states mount
  it**, so this notice renders in six swept states on every run of that file. That is free incidental
  coverage and worth keeping — but a `role="group"` there reds three separate assertions (the pinned
  `toBe(32)` shrink proof, its `toBeLessThan(GROUP_INSTANCE_FLOOR)` sibling, and the pinned R0 list),
  and a `<button>` enters the R1/R4 populations.
- [x] `folio-designer/src/App.css` -- add the `canvas-table-*` rules: chip on
  `var(--color-page-thead)` with a `var(--color-page-thead-line)` bottom rule, collection text in
  `var(--color-bind-on-page)`, count in `var(--color-page-ink-muted)`, header labels in
  `var(--color-page-ink)` with `font-weight: 600` and a `var(--color-page-ink)` bottom rule, cells in
  mono, `white-space: pre` and CSS-only clipping (`overflow: hidden`, optionally `text-overflow`) on
  every text cell — **condition 2 of R1: the browser must make no break decision, and there is no
  computed ellipsis** — and the dashed-outline flip for the no-columns state. Tokens only, no
  `@media`, no `border-radius`, and **never** a class containing `canvas-text` nor a rule grouped
  with a `.canvas-text*` selector.
- [x] `folio-designer/src/` (unit tests) -- cover every row of the I/O & Edge-Case Matrix against the
  canvas: bound table, unbound table, no columns, unbound column, empty label, one column (singular),
  non-table component rejected by the guard, zero-width column accepted, and the echo. Assert the
  **unbound** cell is **not** in the bind accent. **Assert the consumption direction (R2): give a
  projection whose `headerAlign` and `cellAlign` DIFFER and assert the header row takes the first and
  the representative row the second — a swap of the two must red this test.** For each "is not
  offered / is not there" claim, the test must **ADD** the forbidden thing rather than revert the
  implementation — reverting cannot falsify an absence claim, and four false greens in this epic were
  produced that way.

**Acceptance Criteria:**
- Given a table component with columns, when the canvas draws it, then a reader can see the bound
  collection, the column count, every header label and one row of binding placeholders without
  opening the table editor or rendering a preview.
- Given the projection, when the canvas draws any part of the table, then every drawn value came
  from the projection and the browser performed no cascade, no measurement and no width arithmetic
  beyond the existing zoom scale.
- Given a table whose `headerStyle.align` differs from its `style.align` and a column that declares
  no alignment of its own, when the canvas draws it, then the header label and the representative
  cell are aligned **differently**, each matching what the PDF prints for that row.
- Given `canvas-authority-contract.test.ts`, when a later author reads its comments, then the
  three-condition display-paint test is stated there, with condition 3's enforcement named and the
  precedent explicitly bounded — so a future candidate for this exception recognises itself and knows
  that failing condition 1 or 2, or asserting condition 3 rather than having it scanned, needs a new
  ruling.
- Given a Go-side change to the component projection's key set, when the Go suite runs, then a
  mismatch with `engine-protocol.ts`'s `hasOnly(component, …)` list fails a test — rather than
  shipping green and terminating the worker at runtime.
- Given the five verification commands, when they run at the tree this story produces, then each
  matches or improves on the recorded baseline, with every new test name accounted for.

## Spec Change Log

### 1 — ORCHESTRATOR AMENDMENT TO FROZEN INTENT AFTER APPROVAL (2026-09-10)

**Triggering finding (step-04, `intent_gap`, review_loop_iteration 0 -> 1).** The frozen `## Boundaries &
Constraints` **Always** clause *"The new field must not newly refuse a document that ships today"* and the two
frozen I/O-matrix rows for an over-long column `label` / `bind` gave **opposite answers about the same
document**. Proved by execution, not argument: the golden fixture `worked-example.json` with column `e3`'s
label at 600 bytes **parses**, **renders a real 64,123-byte PDF**, and had its canvas projection **abort**
(`folio: component table column label exceeds the projection bound`) — which by this spec's own frozen text
terminates the worker with no respawn. Measured on both sides: at `2c1dd33` without the story the same
document projected **OK**; with the story it aborted. The precondition was checked rather than assumed —
`decodeColumn` in `internal/template/parse_bands.go` imposes **no** length cap on `label` or `bind`.

**What was amended, and by whom.** The **orchestrator** amended the frozen block: the two projection-bound
matrix rows now read *clipped, never an error, column never omitted*, and four bullets were added to
**Always** — the amendment notice, the neighbouring-row argument, clip-never-omit, and the rune-boundary
requirement. This is an edit to frozen intent **after approval**, which only a human may make; it is recorded
here rather than quietly rewritten so the change and its reason stay visible.

**Ruling: (B), remedy CLIP — not omit.** The reading was settled by the matrix's **own neighbouring row**,
which neither the builder nor the orchestrator had cited: *"Zero or negative column width … the canvas paints
what it is given and refuses nothing it accepts today — must not become an error."* Same principle, two rows
apart, opposite answer — so the matrix was **not internally consistent**, which removes the consistency
argument for keeping the aborts.

**The builder recommended OMIT and was overruled, and the reason is the substance.** A column dropped from the
projection makes the canvas *systematically wrong about the document's structure* — the chip's `5 columns`
would sit above four drawn headers — and being systematically wrong about structure is precisely the harm
AD-17 exists to prevent, the same sentence [R1] leans on. **Omitting converts a display problem into a
structural lie.** Clipping keeps every column, keeps the projection bounded, and is a display concern on a
surface that is display-only paint by [R1].

**Known-bad state avoided.** A template that prints correctly blanking the designer's canvas until reload —
a fatal, silent-until-reload outcome for a display concern. And the two worse remedies: **omit** (the
structural lie above) and **(C) bound at the loader** (refused — it touches `internal/template/`, which the
frozen **Never** forbids, and would newly refuse the document at *load*, which is worse than what is fixed).

**KEEP instructions — what must survive and was NOT re-derived.** The gap was **two bounded sites**. This was
patched, not looped back: re-deriving would have discarded work already proved. Preserve all of it —
- the **R2 pair**, both asserting **direction** and re-proved by mutation at step-04: the Go
  differ-precondition plus `HeaderAlign == "right"` / `CellAlign == "center"` from their own cascades, and the
  TSX `['right','right']` / `['center','center']` with the explicit `not.toEqual` between rows;
- `columnAlign`'s extraction in `table_render.go` — a behaviour-preserving change forced by this spec's own
  anti-mirroring rule, adopted at **all four** open-coded render sites (the fifth hit,
  `table_columns_projection.go`, is the editor's deliberately different cascade and is correctly left alone);
- the **component-level wire record** closing DW-74, with its `t.Fatal` fixture preconditions;
- the contract-test **comment block** carrying [R1]'s three conditions, and its **three executable planted
  violations**, which pin by regex identity (`prohibited[1]`, `[2]`, `[9]`) rather than by description;
- **Deviation 1** (no `columns` length cap in the browser guard) and **Deviation 2** (the shortened
  `No columns yet.` notice) — both premises independently verified and **ratified**.

**Counted deliberately:** this is the **third** clause in Epic 14 to forbid the thing it required, and it is
to be counted alongside the two from Story 14.8.

**Two Task bullets were left stale by the amendment and are now reconciled (builder, step-04).** The
`page_setup.go` bullet still said *"with their own named errors"* and the `canvas_body_text_bounds_test.go`
bullet still said *"one probe per new bounded string site ... each asserting its own message"* — both written
under the pre-amendment reading. **The implementer followed the amendment, which governs, and flagged the
divergence rather than silently reconciling it or silently obeying stale text; that was the right call and is
recorded as such.** Both bullets are struck through in place rather than rewritten away, so a later reader
sees what the instruction used to say and why it changed. The probe table returns to its original **eight**
entries — verified: no pre-existing subtest was lost, and the clipping assertion took the coverage's place in
the same function.

## Design Notes

**Why AC3 and D-13.4.1 do not conflict, and no Open Question is raised about it.** The dispatch
asked me to reconcile *"the canvas is the approximate representation"* (D-13.4.1) with AC3's
*"exactly as it will print"*, or raise it. Measured, they do not conflict, and the evidence is
one-directional:

- D-13.4.1 is a ruling about **building the no-data stand-in preview in Go**. Its only use of
  "approximate" is a closing citation-hygiene clause — *"FR34 is the Design Canvas, the approximate
  representation, and citing it for the exact-preview surface would mislead every later reader"*. It
  makes no claim about how faithful the canvas may be. The clause the dispatch is thinking of is in
  D-14.3.1/Q1: *"the canvas is architecturally the approximate representation (D-13.4.1 — the
  exactness promise lives on the PDF and the Preview surface)"*.
- The approximation is scoped to **text metrics and pagination, not geometry.** AD-17's stated harm
  is *"a canvas that is not merely approximate but systematically wrong about where content lands"*.
  Story 5.9's own framing: *"approximate in scale, not wrong about where my content lands"*. Story
  5.13's AC is the most direct sentence in the corpus: *"the canvas stays explicitly approximate
  about **text only**"*.
- `TableEditor.dc.html` states the geometry rule the canvas would be violating by approximating it:
  *"Column widths are fixed and never negotiated against content."*

**So honouring a declared width and a resolved alignment is the same class of act as painting an
image's fit rectangle from engine millipoints (Story 5.13): geometry supplied by the engine, drawn
faithfully.** The one wording caution: AC3's *"exactly as it will print"* is honoured as *the
declared width and the resolved alignment are honoured*, not as a claim that the surface is
print-exact — that claim belongs to Preview alone.

**Why the guard goes in before the Go field.** Not a style preference. The component guard is
`hasOnly`, a subset check, so the two sides' failure modes are asymmetric: an unlisted TS key is
inert, an unlisted **Go** key makes `isCanvas` false, `parseInbound` undefined, and
`engine-client.#fail` calls `worker.terminate()` with no respawn. Ordering TS first means no
intermediate state of the working tree can kill the worker.

**Why the wire record must be measured against a table-bearing document.** The new member is
`omitempty` and absent on every non-table component. All four existing canvas records project
`canvasWindowCountControlTemplateJSON`, which contains no table, so a component-level record built
on it would never see the key and would go green proving nothing — *a fixture more complete than the
defect's precondition is a guard that cannot see it*, in this run's own words. The fragment record is
the right template because it already handles two mutually-exclusive keys with per-population exact
sets; the chain record is the wrong one, because its zero-vs-projected identity only holds where
nothing is optional.

**Why a new token is not minted for the 600-weight header label.** No page-ramp token is 600 weight,
and `design-contract.test.ts` asserts token-**name** equality between `designTokenSets` and
`DESIGN.md`, which is read-only for this story. The design's own precedent for a value with no token
is a component literal (`components.band-tab` declares `paddingY: '3px'` as a literal, and
`tokens.css` carries `--palette-item-padding-y: 7px`). A literal `font-weight: 600` on a
`canvas-table-*` class, with a comment saying the page ramp has no 600 face, is the honest arm.

**Why the unbound cell is not amber.** DESIGN.md: *"amber means data, and only data. … It never
marks selection and never marks structure."* An unbound cell has no data. Painting `Not set` in the
bind accent would say the opposite of what the accent means, and AC6's requirement is only that the
cell *"reads as unbound rather than as an empty string"*.

## Verification

Run the first four from `folio-designer/`. ⚠ `npx vitest run` **from the repo root is a false
mass-failure** — it picks up the Playwright specs and runs them in the wrong environment. Run the
fifth from `folio-go/`; `go test ./...` inheriting a `folio-designer` cwd reports a spurious
`exit 1`.

**Commands:**
- `npx vitest run` -- expected: exit 0. **Baseline measured by me at `06c279a`, tree clean: 76 files
  / 1383 tests / 0 failures.** Report the new counts and diff test names as a **multiset**, not a set.
- `npx tsc -b --force` -- expected: exit 0 and **0 bytes** of output. Measured 0 bytes at baseline.
- `npx oxlint` -- expected: exit 0, 0 errors. Baseline warning **SET**, measured at `06c279a`
  (never an integer — D-14.7.2): `src/App.tsx` ×2, `src/preview/pdf-viewer.tsx` ×2,
  `src/segmented-control.tsx` ×3, all `react(only-export-components)`. **Re-measure the set and
  explain any new key. Never change code to restore a count. Do not quote the line numbers — they
  move whenever `App.tsx` moves.**
- `npm run test:e2e:compile` -- expected: exit 0. Measured exit 0 at baseline.
- **ADDED BY THE ORCHESTRATOR FOR THIS STORY ONLY**, run from `folio-go/`:
  `go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: exit 0.
  Baseline measured at `06c279a`: **15 `ok`**, exit 0, three `[no test files]` packages.

  **Why it is here.** This story changes Go — the canvas projection, its wire record and its bound
  probes — so the story's own gate runs Go. The four designer commands cannot execute `folio-go`, so
  without this the projection work would be verified only by CI, *after* the commit. It costs ~10s.

  **This is not a change to D-000.33 and does not need the owner.** The owner's cadence is the
  *minimum* a story must run; adding a suite strictly increases verification and removes nothing. It
  applies to **this story only** — do not carry it into a later dispatch.

  **The invocation is copied from CI verbatim** (`.github/workflows/ci.yml`, the
  `go test -count=1 -skip "$KNOWN_RED_TEST" ./...` step) and must not be re-derived.
  `TestCorpusMeetsP6ExerciseFloors` is a **deliberate, anchored quarantine** for P6g's unmet
  opaque-name floor, asserted to **FAIL** in its own `folio-go-known-red` job, which reds if it ever
  starts passing. **Never widen that filter, never remove the anchors, and never "fix" that test** —
  D-000.17: a floor that is not met is reported unmet, never filled.

**Suites that deliberately do NOT run** (D-000.32 — name them in the report in these words): the
browser suite, the matrix legs, `npm run build` as a gate, the `verify:offline*` chain, and the
font-host scans.

⚠ **"The Go suites" must NOT appear in that list.** The untagged Go suite **does** run for this
story, per the fifth command. Say that it ran, on CI's exact invocation, with the known-red still
quarantined rather than silenced. Reciting the old list with "the Go suites" in it is a false absence
claim. The **matrix legs** still do not run: they are `//go:build matrix`-tagged and `./...` does not
reach them.

**What this change is and is NOT proved by — state this plainly in the report.**
- **NOTHING byte-level proves this change, and adding a fixture would not change that.** The matrix
  legs, `hashmatrix` and the cross-target byte-identity workflow all render **PDF bytes** from
  `fixtures/*/input.folio`. `CanvasProjection` is never serialized into a golden and never enters a
  PDF, so a canvas-projection field is invisible to every one of them **by construction**, not by an
  oversight a fixture could fix. DW-383's remedy (a new matrix-registered fixture) is the right
  remedy for a *format* field and the wrong one here.
- What **does** cover it: the component-level wire record this story adds (Go bytes ↔
  `engine-protocol.ts` text), the behavioural Go tests over `Canvas(...)`, the new bound probes, and
  the designer unit tests over the painted DOM.
- The **browser** never executes any of it in this story's gate: no Playwright spec runs, so the
  claim "the canvas draws this correctly in a real browser" is **unverified** at this tree and must
  be reported as such.

**Manual checks:**
- `git diff --stat` shows no version constant moved and **nothing** under
  `folio-go/internal/template/`.
- `git status --porcelain` shows nothing under `fixtures/`, nothing under
  `_bmad-output/planning-artifacts/`, and no `sprint-status.yaml`, `deferred-work.md`, `epics.md` or
  `DESIGN.md` change.
- `git log --oneline -1` still reports `06c279a` — the implementer makes no commit.

## Suggested Review Order

Stops are grouped by concern. Every line number was measured at the final tree; `App.tsx` and
`TableEditor.tsx` move constantly in this epic, so re-measure before quoting any of these elsewhere.

**The wire seam — read this first; a drift here is a dead worker with every test green**

- The entry point: the projected column, six keys, millipoints, both alignments resolved in Go.
  [`page_setup.go:358`](../../folio-go/page_setup.go#L358)

- The member on the component, `omitempty`, so a non-table carries nothing.
  [`page_setup.go:312`](../../folio-go/page_setup.go#L312)

- The browser's mirror of that shape, and the closed sets it admits.
  [`engine-protocol.ts:799`](../../folio-designer/src/engine-protocol.ts#L799)

- `hasOnly` is a subset check, so Go-then-TS is fatal and TS-then-Go is inert.
  [`engine-protocol.ts:715`](../../folio-designer/src/engine-protocol.ts#L715)

- Closes DW-74: the component key set had no record on either side until now.
  [`canvas_projection_wire_test.go:663`](../../folio-go/canvas_projection_wire_test.go#L663)

- The cross-language fence, anchored on `hasOnly(component, [` rather than the top-level list.
  [`canvas_projection_wire_test.go:820`](../../folio-go/canvas_projection_wire_test.go#L820)

**Clipping — the frozen amendment, and the reason it is not an abort**

- Clips at the bound walking back to a rune boundary; a byte cut would break the JSON envelope.
  [`page_setup.go:1959`](../../folio-go/page_setup.go#L1959)

- Returns no error at all: a document that prints must never blank the canvas.
  [`page_setup.go:1921`](../../folio-go/page_setup.go#L1921)

**Alignment — one source shared with the renderer, never a mirror of it**

- The cascade's last step, extracted so the projection calls it instead of copying it.
  [`table_render.go:521`](../../folio-go/table_render.go#L521)

**The display-paint exception — R1's ruling, and the only place it is enforced**

- The three conditions, where a later author will actually find them.
  [`canvas-authority-contract.test.ts:150`](../../folio-designer/src/canvas-authority-contract.test.ts#L150)

- Condition 2 in CSS: wrapping off, clipping by CSS, no computed ellipsis.
  [`App.css:414`](../../folio-designer/src/App.css#L414)

- The marker pinned per element; removing it from any one site reds.
  [`canvas-authority-contract.test.ts:642`](../../folio-designer/src/canvas-authority-contract.test.ts#L642)

**The paint**

- Replaces the word "Table"; non-interactive, so 14.10 still owns addressability.
  [`App.tsx:4972`](../../folio-designer/src/App.tsx#L4972)

- Tracks come from engine millipoints through the one zoom mapping.
  [`App.tsx:4995`](../../folio-designer/src/App.tsx#L4995)

- A display-only floor: a negative width would void the whole grid declaration.
  [`App.tsx:4958`](../../folio-designer/src/App.tsx#L4958)

- The surface never takes pointer events, keeping one control per component.
  [`App.css:415`](../../folio-designer/src/App.css#L415)

- The no-columns placeholder takes the design's dashed grammar, not an empty frame.
  [`App.css:449`](../../folio-designer/src/App.css#L449)

**Peripherals**

- Every matrix row, plus the marker, zoom, negative-width and long-path cases.
  [`canvas-table-paint.test.tsx:1`](../../folio-designer/src/canvas-table-paint.test.tsx#L1)

- Behavioural cover for both cascades, clipping, the rune boundary and column-id uniqueness.
  [`canvas_table_column_projection_test.go:1`](../../folio-go/canvas_table_column_projection_test.go#L1)

- The probe table returns to eight; clipping is asserted in the same function.
  [`canvas_body_text_bounds_test.go:1`](../../folio-go/canvas_body_text_bounds_test.go#L1)
