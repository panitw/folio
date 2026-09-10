---
title: 'Story 14.8: The table editor carries the header, cell and border sections'
type: 'feature'
created: '2026-09-10'
status: 'draft'
review_loop_iteration: 0
context: []
---

# Story 14.8: The table editor carries the header, cell and border sections

**In plain terms.** The table editor's settings all sit in one undivided run today, even though the
design draws them as three named sections. This story separates them into HEADER, CELLS and BORDERS,
which is mostly a tidying job — the controls themselves do not change, they just end up under the
heading they belong to, and the headings become real headings so that someone using a screen reader
hears the same three sections a sighted author sees.

The one genuinely new thing is a header border. A table can already be given a border, but the header
row cannot be given a *different* one, even though the engine has always known how to draw that and
has been holding the field in reserve for this story by name. So BORDERS gets a width, a colour and a
choice of which edges to draw, for the header row only.

That comes with a catch worth understanding, because it is the reason this was left until now. The
engine treats the header's border as all-or-nothing: the moment you set any part of it, the header
stops taking anything from the table's border. So setting just a width would quietly throw away the
colour the header had been drawing with. The panel therefore does two things. It shows you what the
parts you have not set will actually draw — half a point, black, all four edges — and it tells you, in
words, that the header border has stopped following the table's. Setting one thing never silently
sets the other two on your behalf.

Two smaller notes. Some things the design drew are not settings at all: the header always repeats onto
continuation pages, and row height always comes from the content, so those are stated as facts rather
than offered as controls. And the design file itself is being corrected, because it draws four
settings the product does not have and should not keep promising them.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The table editor holds everything Story 12.3 made authorable in one undifferentiated
block — `div.table-editor-header`, `role="group" aria-label="Table header and rows"` — so a surface the
design draws as three named sections reads as one list. Separately, `headerStyle.border` is a field the
format carries and the renderer already paints, and no command can author it: the closed set at
`tableHeaderStyleFields` names nine fields and its own comment says `border` *"waits on Story 14.8's
BORDERS section"*. And `TableEditor.dc.html` draws four settings the product does not have.

**Approach:** Regroup the shipped controls under three headings — HEADER, CELLS, BORDERS — inside the
one existing group, adding no control and rewiring none. Then extend the closed set from nine fields to
twelve with `border.width`, `border.color` and `border.edges`, authored one attribute per command,
projected with the resolved twin every sibling field already carries, and disclosed in words because the
cascade takes the block whole. Finally, correct the drawing so it stops promising what cannot be built.

## Boundaries & Constraints

**Always:**
- **The header border cascade is BLOCK-GRANULAR and the panel must say so in words.**
  `table_render.go` resolves `case hasHeader && header.Border.Set && !header.Border.Null: r.hasBorder,
  r.border = true, header.Border.Value`, with the table's own `base.Border` only as a sibling `case` —
  an either/or take, never a field-by-field merge. Authoring any one attribute materialises the block
  (`component_commands.go` does `if !st.Border.Set || st.Border.Null { st.Border = Presence[Border]{Set:
  true} }`) and the table's border stops contributing entirely from that moment.
- **Arm C.** Send only the attribute the author touched. Disclose the other two through the resolved-twin
  `<output>` idiom the nine incumbent header fields already use.
- **A resolved twin shows the RESOLVED value, never the authored one.** A twin echoing what the author
  just typed is the both-sides-move-together shape and asserts nothing.
- **The resolved values are computed in Go and projected.** The engine's paint-time defaults (0.5pt,
  `#000000`, all four edges) must not gain a TypeScript copy — the canvas already mirrors two of them at
  `App.tsx`'s `?? 500` / `?? '#000000'`, and a third copy is a third thing to drift.
- Every control placed in HEADER and CELLS is the control already shipped, sending the command already
  registered. A regroup changes no accessible name.
- Design tokens only; no hard-coded hex, including inside a CSS comment.

**Ask First:**
- Any change to `parse_bands.go`, `serialize.go`, `table_render.go`, or any format version constant. The
  field already exists, already parses, already round-trips and already paints; if the implementation
  reaches for any of these, it has gone wrong.
- Editing `control-vocabulary-contract.test.tsx`'s pinned `r0Violations` list, `GROUP_INSTANCE_FLOOR`, or
  `CHECKED_GROUPS`. The chosen markup is specifically the arm that touches none of them.
- Adding any `role="group"`, or promoting a section heading to a group with `aria-labelledby`.
- Adding `bold` or `italic` to the panel's authoring union (DW-369 — out of fence).

**Never:**
- **Never commit, branch or push.** See the operational constraints in `## Tasks & Acceptance`.
- Never author padding, on a table or anywhere else (D-12.4.1). The struck padding criterion retained at
  the foot of `epics.md` §14.8 is **not live scope** and must not be implemented — nor deleted from
  `epics.md`, which is not this story's file to edit.
- Never build `Show header row` or the three-way None / Horizontal / All borders preset. Neither has a
  format field.
- Never restate the table's *body* border. That is already authored in the inspector's BOX section for
  every non-line component including a table (D-14.4.Q2(a)). BORDERS adds the header-row override only.
- Never offer `Row height` or `Repeat on continuation pages` as settings; both are engine-derived.
- Never word `Repeat on continuation pages` as an absolute — `DiagCodeTableHeaderRepeatSuppressed` is the
  engine's own record of the case where the repeat is suspended on a single page.
- Never edit `sprint-status.yaml`, `deferred-work.md`, `epics.md`, `DESIGN.md`, or anything under
  `fixtures/declared-variants/`.
- Never edit `.working/TableEditor.dc.html` (DW-375).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| First border attribute authored | Table with `style.border` set, no `headerStyle.border`; author sets header border width | `updateTableHeaderStyle` with `field: "border.width"`, `op: "set"`; block materialises; panel states the takeover in words and the twins show the resolved 0.5pt/`#000000`/all-four for the two untouched attributes | N/A |
| Second attribute authored | `headerStyle.border` already carries a width; author sets the colour | Only `border.color` is sent; the width already stored is untouched | N/A |
| Clearing one attribute | `headerStyle.border` carries width and colour; author clears the colour | `op: "clear"`; colour becomes absent; the width stands; border still drawn | N/A |
| Clearing the last attribute | `headerStyle.border` carries only a width; author clears it | The empty `border` block is collapsed away, and if it was the last member of `headerStyle` that block is collapsed too — no `"border": {}` and no `"headerStyle": {}` in the file | N/A |
| Clearing what is already absent | No `headerStyle.border` at all; author clears a border attribute | Byte-level no-op: no history entry, and the dialog's edit count does not move | N/A |
| Negative width | Author enters a negative header border width | Engine refuses, located at `table.headerStyle.border.width` | Located alert in the dialog's existing error channel; no document change |
| Zero width | Author enters `0` | Accepted — zero is the thinnest device line PDF can draw, not an absent border | N/A |
| Non-`#RRGGBB` colour | Author enters `red` | Refused at the command door, located at `table.headerStyle.border.color` | Located alert; no document change |
| No border anywhere | Neither the table nor the header declares a border | The resolved edges twin is empty, which is exactly the state "nothing is painted" | N/A |
| Hand-authored border preserved | Document hand-edited to carry `headerStyle.border` with an unknown extra key | The panel edits the attributes it knows and preserves the rest untouched on save | N/A |

</frozen-after-approval>

## Code Map

All anchors below were measured by me at `c3e7a6b` on 2026-09-10. **Anchor rot is endemic in this epic —
`TableEditor.tsx` moved twice in 24 hours — so prefer the named identifier over the line and re-measure
before you edit.** Nothing here was carried from the epic context or from a register without checking.

**The group and its nine controls (Q1's subject)**

- `folio-designer/src/TableEditor.tsx:448` — `<div className="table-editor-header" role="group"
  aria-label="Table header and rows">`, heading `<p className="section-label">HEADER AND ROWS</p>` at
  `:449`, closing at `:481`. **Find it by the aria-label, not the line.** Nine flat sibling controls, in
  order: header height (`:450-453`), alternating row background (`:454-461`), header font family
  (`:462-468`), font size (`:469`), line spacing (`:470`), background (`:471`), text colour (`:477`),
  vertical alignment (`:478`), alignment (`:479`); trailing `<p className="honest-note">` at `:480`.
- **Exactly one control is data-row-scoped**: alternating row background. The other eight are header
  scoped. So CELLS receives one control and the regroup moves one block of JSX.
- Helper factories: `styleSelect` `:259`, `styleColour` `:286`, `styleNumber` `:300`; `commitStyleText`
  `:247` holds the per-control no-op comparison (`if (value === committed) return`).
- The resolved-twin idiom BORDERS must copy: `:467` is
  `<output aria-label="Resolved Header font family">{resolvedNote(table.headerFontFamilyResolved, 'Using:
  nothing — this table names no font chain')}</output>`, with the same shape at `:265`, `:293`, `:306`.

**The command surface (Part 3's subject)**

- `folio-go/component_commands.go:2528` — `var tableHeaderStyleFields = []string{...}`, **nine** members.
  **Cite it by name; `:2511-2527` is the ruling comment above it, not the set.** The comment carries the
  `border`-is-deferred paragraph that this story replaces in place.
- `folio-go/component_commands.go:2683` `updateTableHeaderStyle` — the arm. Gate at `:2699`; path built at
  `:2702` as `path := "table.headerStyle." + field`; `tableCommandOp(raw, id, path, 5)` at `:2703` counts
  **top-level** keys, so a value's shape never moves that arity. Clear arm `:2723-2741`; the
  already-absent short-circuit at `:2717-2719`.
- Only three decoders exist on this path: `propertyString` `:1113`, `propertyBool` `:1120`,
  `propertyLength` `:1127`. **The array precedent is `borderEdges` at `:1444-1463`** — plain
  `json.Unmarshal(value, &edges)`, rejecting an empty array, with block materialisation at `:1457-1459`.
- `folio-go/component_commands.go:2829-2838` `cleanupEmptyHeaderStyle` — already counts `style.Border.Set`
  at `:2835`, but **has no empty-`Border` collapse**. Its element-level twin `cleanupEmptyStyle`
  `:1480-1500` **does**, at `:1485-1490` (and the same for `Padding` at `:1491-1496`). That six-line block
  is the positive control for the mirror this story owes.
- **Element-level authoring is the precedent being completed:** `applyPropertyChanges` carries
  `borderWidth`/`borderColor`/`borderEdges` as three flat keys (`:1082`, `:1139`, `:1148`, `:1232`,
  `:1382`, `:1444`). Measured: setting `borderWidth` alone on an element with no style yields exactly
  `{"style":{"border":{"width":2}}}` — no colour and no edges are seeded.

**The projection, and the test that co-determines it**

- `folio-go/table_columns_projection.go:76-96` — 24 keys, **every one a flat scalar**; nine
  committed/`Resolved` pairs plus `headerHeight` (the deliberate singleton) and `altRowBackground`.
  Builder `TableColumns` at `:154`, struct literal at `:184`. The comment at `:70-75` is the standing
  ruling that this projection is engine↔browser, in-repo, moving in one commit, and **not tag-bound**.
- ⚠ `folio-go/table_header_style_test.go:831`
  `TestTheProjectionCarriesAPairForEveryHeaderStyleFieldACommandCanWrite` — **this is why the wire and the
  projection cannot be chosen independently.** It reflects the json tags, strips the `header` prefix,
  lowercases the first letter, and asserts at `:869` that `paired` **equals** `sorted(tableHeaderStyleFields)`,
  and at `:872` that `unpaired` is exactly `["height"]`. Its positive control at `:843` fails the test
  outright if the tag count and `tableColumnsProjectionWireKeys` disagree.
  **Consequence: a wire field named `border.width` forces the json tag `headerBorder.width` and its twin
  `headerBorder.widthResolved`.** The dot propagates into the tag, into `hasExactKeys`, and into quoted
  TypeScript property access. That is a measured consequence of the ruled spelling, not a choice left open.
- `folio-designer/src/engine-protocol.ts:469` `isTableColumns` — `hasExactKeys(value.table, [...24 names])`.
  `hasExactKeys` (`:460`) rejects a **missing** key as well as a surplus one, unlike `isCanvas`'s `hasOnly`.
  **So Go and TS must move in one commit** or `parseInbound` refuses, the worker terminates, and the first
  table-editor open shows nothing at all.
- The array-and-closed-set precedent to copy: `table.columns` is validated in the same expression as
  `Array.isArray && length <= 128 && every(isRecord && hasExactKeys(column, [...10]) &&
  ['left','center','right'].includes(column.align) && ['','sum','avg','count'].includes(column.footer))`.
- `folio-designer/src/engine-client.ts:117-133` — the `#settle` hand-enumeration **is fixed** for the table
  object (`{ ...message.tableColumns.table, columns: [...].map(...) }`), so a new table-level member
  survives to `App.tsx`. ⚠ The **preview** arm one line above is still hand-enumerated — do not read the
  fix as global.

**The renderer's defaults, which the resolved twins must report (read-only)**

- `folio-go/table_render.go:321` `resolveHeaderStyle`; the border arm at `:357-362`.
- `folio-go/table_render.go:595-612` inside `buildCellRectWithBackgroundField` — width absent →
  `geom.Length(500)`; colour absent → `#000000`; edges absent → all four. Emission gate in
  `internal/pdf/rectdoc.go`: `if r.HasStroke && (Edges.Top || Right || Bottom || Left)`.
- `folio-go/internal/template/parse_bands.go:805-823` — every border sub-key independently optional;
  **"ZERO IS VALID and stays accepted: it is the thinnest device line PDF can draw, not an absent
  border. Only a NEGATIVE width is refused."** Unknown edge refused at `:830-834` against
  `closedBorderEdges` (`internal/template/closedsets.go:189`). A bad colour is **not** caught at load
  (AD-1 forbids the import) and surfaces at render.
- `folio-go/internal/template/serialize.go` — `headerStyle` writes through the **same** `writeStyle` as
  element `style` (`:410` vs `:376`), reaching the **same** `writeBorder` (`:508`, defined `:532`), which
  emits only `Set` sub-keys. **Executed:** `{"headerStyle":{"border":{"width":9}}}` round-tripped
  byte-identically and `Parse∘Serialize` was a fixed point. **Nothing on the write path changes in this
  story, which is why AD-21 is untouched.**

**Tests that will red, and why**

- `folio-go/table_header_style_test.go:730` — `border` sits in an explicitly-refused field list. Reds by
  design; the list is what this story retires.
- `folio-go/table_header_style_test.go:753` — the cascade `rows` list must equal `tableHeaderStyleFields`;
  its own comment at `:746` says a tenth field reds it.
- `folio-go/table_header_style_test.go:870` + `:843` — the pair/wire tie described above.
- `folio-go/table_header_style_test.go:895-905` `committedTwin` — a `switch` returning `""` on default.
  ⚠ **A missing arm does not red; it makes the new field's leg-one check measure nothing.** Vacuity, not
  failure.
- `folio-go/canvas_projection_wire_test.go:474-499` `tableColumnsProjectionWireKeys`, asserted `:581`;
  `:605` extracts the `hasExactKeys` list out of `engine-protocol.ts` and compares — the cross-language
  fence. `:521-535` + `:562-568` require the fixture to exercise every committed member non-zero.
- `folio-designer/src/engine-bounds-mirror.test.ts:885` — pins the **exact whole import line** of
  `TableEditor.tsx` (`^import \{ MAX_ENGINE_HISTORY_ENTRIES, type TableColumns \} from './engine-protocol'$`).
  A third import specifier reds it. Sleeper.
- `folio-go/table_header_style_test.go:919`
  `TestClearingTheLastAuthorableHeaderFieldKeepsAHandAuthoredBorder` — stays **green**, but its comment at
  `:908-914` becomes a stale assertion the moment `border` becomes authorable. Update the prose.

**Tests that will NOT red — checked so nobody chases them**

- `control-vocabulary-contract.test.tsx` — floors at `:591-594` (`GROUP_INSTANCE_FLOOR = 33`) and the
  pinned `r0Violations` `toEqual` at `:885-889` whose first member is
  `'R0 the sweep visited 32 group instances, under the floor of 33'`. **The chosen markup adds no
  `role="group"`, so the shrunk count stays 32 and neither moves.** `CHECKED_GROUPS` `:566` keeps
  `'Table header and rows'` because the group is kept. ⚠ The engine mock at `:315-320` **does** answer
  `table-columns` and the table-editor state is the largest swept at 54 controls — so new controls here
  **are** checked, and `V2_CENSUS` `:506-521` will move if BORDERS ships `×` clear glyphs.
- `App.test.tsx:482` — `getByRole('group', { name: 'Table header and rows' })`. Kept, so green.
- The ~50 accessible-name assertions in `App.test.tsx` (`:456-728`) — a regroup does not touch accessible
  names. `:563`'s clear-button loop is a containment loop, not an exact list.
- The focus trap `App.test.tsx:311-381` — it re-derives its ends from the DOM and asserts only
  `Root collection` first and `Done`/`Cancel` last. New controls land mid-list. It reds only if a control
  is placed before `Root collection` or after the footer.
- `folio-designer/e2e/table-editor.spec.ts` and `e2e/browser-native-roundtrip.spec.ts` — neither names the
  group or any header-style control.
- All four format fences: `drift_test.go` (`border` is already emitted and already backticked in
  `folio-format.md`), `numeric_classification_test.go` (`width` already registered), `fixtures_test.go`
  (`maximalFixture` **already carries a full `headerStyle.border`**), `goldenfixture_test.go`. **No format
  change, no version increment.**
- `canvas-authority-contract.test.ts` population floors — all `toBeGreaterThanOrEqual`, all comfortably
  clear except production `.css` sitting exactly on 3. Adding a file never reds; only deleting does.

**The mockup**

- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/TableEditor.dc.html`, 207 lines,
  **100% inline styles and zero `class` attributes** — an edit must match that. Options band `147-186`:
  HEADER `149-162`, CELLS `164-176`, BORDERS `178-185`.
  `Show header row` `152-155`; `Repeat on continuation pages` `156-160` with a `REQUIRED` badge at `:159`;
  `Padding` `167-170`; `Row height` `171-174`, drawn with the **same dropdown chevron** as the real footer
  aggregate pickers, so `auto` currently reads as one selectable option among several; the three-way
  preset `180-184`.
- **Nothing reads this file.** No test or script opens any `.dc.html` (positive control:
  `design-contract.test.ts:132` and three other sites cite `Font Browser.dc.html` by name in prose only).
- **No mockup in this repo has ever been modified** — `git log --diff-filter=M` over `mockups/` is empty.
  14.8 is the first.
- The locked/derived idiom already exists and must be reused rather than invented: `Binding.dc.html`
  `222-238` — a badge naming the constraint, a plain-sentence reason line at `font-size:10px;
  color:#5e666f; line-height:1.45`, and `opacity: 0.42` on the unavailable control. `DESIGN.md:585` is the
  rule it honours: *"State the reason next to anything disabled."* `DESIGN.md:602-604`: *"Don't draw an
  affordance the product cannot honour."* `DESIGN.md:302-304`: where DESIGN.md and any mockup disagree,
  **DESIGN.md wins.**

## Tasks & Acceptance

**Operational constraints for the implementer (`D-14.4.1` — these are binding):**

> **You must never commit.** You may not run `git commit`, `git add`, `git stash`, `git checkout`,
> `git reset`, `git revert`, `git restore`, `git clean`, `git branch`, `git merge`, `git rebase`,
> `git push`, or `git tag`. The orchestrator makes every commit in this run. Reading git state —
> `git log`, `git status`, `git show`, `git diff`, `git ls-files` — is permitted and encouraged. The
> single carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`,
> removed in a `finally`, is permitted for a throwaway test harness.

**Execution:**

*Part 1 — the regroup (adds no control, rewires none)*

- [ ] `folio-designer/src/TableEditor.tsx` -- inside the ONE existing `role="group" aria-label="Table
      header and rows"`, replace the single `<p className="section-label">HEADER AND ROWS</p>` with three
      `<h3 className="section-label">` headings — HEADER, CELLS, BORDERS — and move the alternating-row
      background block under CELLS. -- The class is kept so `App.css`'s `grid-column: 1 / -1` keeps
      matching; `<h3>` is a correct non-skipping descent under the dialog's existing `<h2>Configure
      columns</h2>`. **Add no `role="group"` and do not promote a heading with `aria-labelledby`.**
- [ ] `folio-designer/src/App.css` -- give the second and third headings their section separation. --
      **The `<p>`→`<h3>` risk is already discharged by measurement and needs no CSS change:** `:133`
      `.section-label` sets `margin: 0 0 var(--space-5)` and `font: var(--type-label)`, and
      `--type-label` is `600 9px/1 var(--font-sans)` — the `font` shorthand carries the weight and resets
      line-height, so an `<h3>` renders identically to the `<p>`. `:652` then overrides to `margin: 0`
      with `grid-column: 1 / -1` inside this editor specifically.
      **The real residual is different:** with `margin: 0`, headings two and three sit flush against the
      control row above, separated only by the grid's `gap`, so a new section is not visually marked.
      Add that separation with tokens only — no hex (the whole-file scan reads comments too) and no
      literal `border-radius`.

*Part 2 — the two derived-fact statements*

- [ ] `folio-designer/src/TableEditor.tsx` -- state `Row height` as the derived fact `auto` and
      `Repeat on continuation pages` as a locked statement, using the reason-line idiom rather than a
      disabled control. -- **`Repeat on continuation pages` must not be worded as an absolute:**
      `DiagCodeTableHeaderRepeatSuppressed` is the engine's record of the single-page case where
      reserving the header leaves no room for a row. Neither is a control and neither carries a
      `REQUIRED` badge promising a guarantee the engine can suspend.

*Part 3 — the closed set goes to twelve*

- [ ] `folio-go/component_commands.go` -- add `"border.width"`, `"border.color"`, `"border.edges"` to
      `tableHeaderStyleFields` and give `updateTableHeaderStyle` their set and clear arms. -- **Dotted,
      not camelCase**, and the reason is measured: `propertyPath` returns the **bare** field name and the
      element path reports `component.borderWidth` (`:1069`), a key the document does not have — that is
      DW-333. Dotted names make `path := "table.headerStyle." + field` produce
      `table.headerStyle.border.width`, truthful by construction. Validate through the same predicates the
      loader asks — never a second literal beside them. `edges` is an array on the wire, decoded like
      `borderEdges` at `:1444-1463`, rejecting the empty array.
- [ ] `folio-go/component_commands.go` -- mirror `cleanupEmptyStyle:1485-1490`'s empty-`Border` collapse
      into `cleanupEmptyHeaderStyle`. -- Flat keys + materialise-on-write + **collapse-on-clear** is one
      three-part mechanism; `cleanupEmptyHeaderStyle` has two of three only because nothing could reach
      the third until now. Without it, clearing the last border attribute leaves `border: {}`, whose
      `Border.Set` keeps `:2835` from ever firing and pins `headerStyle` alive forever. **Red-prove by
      deleting the collapse, not by reverting.**
- [ ] `folio-go/component_commands.go` -- replace the `border`-is-deferred paragraph in the
      `tableHeaderStyleFields` comment **in place** — edited, not deleted, as `:2520-2527` did for the
      seven-to-nine change. -- Carry the **reasoning**: why flat dotted keys won, that a composite `value`
      cannot express one-attribute-at-a-time authoring, and that the collapse-on-clear is the third leg of
      the mechanism. The next person adding a nested member to this closed set hits the identical trap,
      and this comment is the one location in this repo with a demonstrated record of being read.

*Part 4 — the projection*

- [ ] `folio-go/table_columns_projection.go` -- add six members: `headerBorder.width` /
      `headerBorder.widthResolved` (millipoints), `headerBorder.color` / `headerBorder.colorResolved`,
      `headerBorder.edges` / `headerBorder.edgesResolved` (canonical comma-joined `top,right,bottom,left`).
      -- The dotted json tags are **forced** by `table_header_style_test.go:869`, which derives the key
      names from the field names. 24 keys become 30.
- [ ] `folio-go/table_columns_projection.go` -- compute the resolved trio from **one** source shared with
      the renderer's defaults, never a second copy of 500 / `#000000` / all-four. -- `table_render.go`
      `:595-612` owns those defaults today. **`edgesResolved` is empty exactly when nothing is painted** —
      both when no border resolves at all and when `edges: []` suppresses the stroke — which is the only
      unambiguous way this flat shape can express "no border", since a resolved width of `0` is
      indistinguishable from the legal zero-width border the loader explicitly accepts.
- [ ] `folio-designer/src/engine-protocol.ts` -- extend `TableHeaderStyle`/`TableColumns` and the
      `isTableColumns` `hasExactKeys` list to the same 30 keys, with typed clauses copying `column.align`'s
      closed-set idiom for the edges string. -- **Go and TS must land in the same commit**: `hasExactKeys`
      refuses a missing key, and the failure mode is a terminated worker and a blank first open.

*Part 5 — the control*

- [ ] `folio-designer/src/table-style-command.ts` -- widen `TableHeaderStyleField` by the three names and
      add the array encoder case. -- The file holds **no rule of its own** by design; keep it that way so
      it incurs no `engine-bounds-mirror` obligation. ⚠ Check whether `TableEditor.tsx`'s import line
      gains a specifier — `engine-bounds-mirror.test.ts:885` pins it exactly.
- [ ] `folio-designer/src/TableEditor.tsx` -- build three BORDERS controls in the shipped idiom: width via
      `styleNumber`, colour via `styleColour`, and an edges control, each carrying the resolved twin. --
      **Arm C:** send only the touched attribute. The twin shows the **resolved** value read from the
      projection, never the authored one and never a TypeScript copy of the defaults.
- [ ] ⚠ `folio-designer/src/TableEditor.tsx` -- build the edges control **without** a `role="group"` and
      **without** a `×` clear button. -- The obvious move is to copy `BorderEdgesProperty`
      (`App.tsx:4735`), and copying it verbatim breaks two guards at once. It is a
      `<div className="property-edges" role="group" aria-label="Border edges">` holding four checkboxes
      **plus** a `.property-inline-action` `×` `Clear Border edges` button. The `role="group"` takes the
      shrunk sweep 32 → 33, **clears** `GROUP_INSTANCE_FLOOR` and reds the pinned `r0Violations` —
      identical to the arm Q1 refused, arriving through the back door. The `×` is a glyph button outside
      any segmented control, so it moves `V2_CENSUS`. Four bare `<input type="checkbox">` with individual
      accessible names cost nothing on either guard: the sweep's population is `button, [role="button"]`
      only, so checkboxes are never swept. If a clear affordance is wanted, unchecking every edge already
      expresses it — `borderEdges` rejects the empty array, so the command is `op: "clear"`.
- [ ] `folio-designer/src/TableEditor.tsx` -- state the takeover in words in the BORDERS section. --
      Three resolved numbers tell the author what is drawn; they do not say that the header border is now
      authored **as a whole** and no longer inherits the table's. This is the story's principal
      correctness risk and it is not visible from the field's name.

*Part 6 — the drawing*

- [ ] `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/TableEditor.dc.html` --
      remove `Show header row` (`152-155`), the three-way preset (`180-184`) and `Padding` (`167-170`);
      demote `Row height` to a statement by dropping its chevron; redraw `Repeat on continuation pages`
      with a reason line instead of a `REQUIRED` badge on a checkbox. -- Inline styles only, matching the
      file's own idiom; reuse `Binding.dc.html:222-238`'s badge-plus-reason-plus-`opacity:0.42` pattern.
      **Do not touch `.working/TableEditor.dc.html`** (DW-375). Note the CELLS grid is `1fr 1fr` and loses
      a child.

*Part 7 — the guards*

- [ ] `folio-designer/src/TableEditor.test.tsx` -- add the three border entries to `HEADER_STYLE_KEYS`
      **and a negative control**: a test asserting that a command naming a field absent from the map
      **throws** rather than being silently ignored. -- `:163`'s `if (key !== undefined)` makes an unknown
      field a no-op, so a border test would stay green proving nothing. The map is seven entries matching
      the panel's union, so this is not a live vacuity today — it becomes one the moment the union grows,
      which is this story. The negative control protects the twelfth field's *successor*, not just the
      tenth through twelfth.
- [ ] `folio-designer/src/control-vocabulary-contract.test.tsx` -- **prove positively that Part 1 moved no
      group**: assert the shrunk sweep's group-instance count is still 32 against the floor of 33. -- A
      measured null-diff is the proof; "the suite stayed green" is not, because a sweep that stopped
      visiting the state would also stay green. Update `V2_CENSUS` **by executing the sweep and reading
      what it reports**, never by hand — `:881` records that rule.
- [ ] `folio-go/table_header_style_test.go` -- extend the cascade `rows` list, the refused-field list and
      `committedTwin` with the three new fields. -- ⚠ A missing `committedTwin` arm returns `""` and makes
      the new field's leg-one check **measure nothing** rather than fail. Also update `:908-914`'s comment,
      which becomes a stale assertion once `border` is authorable.
- [ ] `folio-go/canvas_projection_wire_test.go` -- update `tableColumnsProjectionWireKeys` and the
      non-zero fixture so every committed member is exercised. -- `:605` compares the record against
      `engine-protocol.ts`'s own list, so this is the cross-language tie.
- [ ] `folio-designer/src/TableEditor.test.tsx` and `folio-go/table_header_style_test.go` -- cover every
      row of the I/O matrix, splitting by layer: the four wire-shape rows (one attribute per command, the
      second attribute leaving the first alone, clear-one, clear-the-last-collapses) and the two refusal
      rows in Go; the disclosure, the resolved twins and the already-absent no-op in the designer suite.
      -- Two rows need naming because they are the ones a green suite can fake: **"clearing what is
      already absent"** must assert the edit count **did not move**, not merely that no error appeared
      (`App.tsx`'s `commitTableColumn` only counts when `committed.snapshot.revision !== revision`); and
      **"no border anywhere"** must assert `edgesResolved` is empty, which is the only signal that
      distinguishes nothing-painted from a legal zero-width border. Red-prove each by **deleting** the
      behaviour under test, never by reverting the story — an absence claim passes more easily under a
      revert.

**Acceptance Criteria:**

- Given the table editor, when it is opened, then it shows HEADER, CELLS and BORDERS as three headed
  sections inside a single `role="group" aria-label="Table header and rows"`, with every pre-existing
  control present under an unchanged accessible name.
- Given the `control-vocabulary-contract` sweep, when it runs after Part 1, then the shrunk sweep still
  reports exactly 32 group instances against a floor of 33 — measured and asserted, not inferred from a
  green suite.
- Given a table whose header inherits the table's border, when the author sets any one header border
  attribute, then the panel states in words that the header border is now authored as a whole and no
  longer inherits the table's, and the two untouched attributes show their resolved values.
- Given a header border command, when it is sent, then it carries exactly one attribute — the panel never
  transmits a value the author did not choose.
- Given `npx tsc -b --force` and `npx vitest run`, when they run after the projection change, then both
  pass, demonstrating that the Go wire record and the TypeScript `hasExactKeys` list moved together.
- Given the mockup, when the story is done, then it no longer draws `Show header row`, the three-way
  borders preset, `Padding`, or `Row height` as a dropdown, and `Repeat on continuation pages` is stated
  with its reason rather than badged as an absolute guarantee.
- Given `git diff --stat`, when the story is done, then no file under `folio-go/internal/template/` and no
  version constant appears in it.

## Spec Change Log

## Design Notes

**Why flat dotted keys and not one `border` block, recorded because the epic's own criterion says "ten".**
`epics.md` §14.8 says the closed set goes to ten with `border` as one member carrying an object. That was
written before any of this was measured and is provisional mechanism inside a settled capability — the
owner ruled the header border **authorable**, not a wire shape. Two grounds decided it, and a third that
was withdrawn should not be repeated:

1. **The product already answered this question in the same file.** The element's nested `style.border` is
   authored from three flat keys with materialise-on-write and collapse-on-clear. Block-on-the-wire would
   be a second, contradictory answer to a question already answered.
2. **A composite `value` cannot express Arm C.** The surface is `{id, field, op, value}` — one field, one
   op. A block `set` writes the whole object, so changing a width means re-transmitting colour and edges
   read back from the projection: read-modify-write across an async boundary, and Arm B smuggled in as a
   wire format. **Arm C and block-on-the-wire are incompatible.**
3. ⚠ **Not a reason: "the block has no local example."** It does — `engine-protocol.ts:517` validates
   `column` as a nested record with a closed sub-key list and per-field typed clauses. Do not repeat the
   novelty-cost argument; it was measured false and withdrawn.

**Conceded against the choice, so the record is honest:** the block would have cost 26 keys rather than 30,
would have made the naming question moot rather than answered, and would have avoided a Go object-decode
arm. Flat wins despite all three.

**Why the shape and the projection could not be mixed.** `table_header_style_test.go:869` derives the
projection's `header*` key names from the command's field names by string transformation. Block-on-the-wire
with flat projection keys reds it with its own message — *"a field on one side only is a value an author
can write and cannot read back (DW-240)"* — and could only be satisfied by editing the tie. There is no
independent projection-shape decision.

**Why three `<h3>`s inside one group rather than three groups.** Three `role="group"`s would take the
shrunk sweep from 32 to 34, **clearing** `GROUP_INSTANCE_FLOOR = 33` — so the pinned clause would stop
proving anything about a dropped state and become a guard that cannot fail, arriving as a side effect of
markup. Raising the floor would have been defensible (the counted population genuinely grew) but
"defensible" is the wrong bar when an arm exists that spends nothing. Plain `<p>` headings were refused on
the accessibility axis: the design would draw three sections while a screen-reader user perceived one
undifferentiated group — the epic's own subject failing inside the epic. Nothing pins `section-label` to
`<p>` and no heading-order contract exists, so `<h3>` reds nothing.

**The latent DataPath collision, stated because it is invisible from the field name.** `App.tsx:3969` is
`!isPropertyField(error.dataPath.split('.').pop() ?? '')` — it pops the last segment of a located path.
Under the dotted spelling, `table.headerStyle.border.width` pops to `width`, which is a `PROPERTY_FIELDS`
member. Under camelCase it would pop to `borderWidth`, also a member. **Both spellings collide**; both stay
latent only because `printsDataPath` takes a `PropertyCommitError` and the table editor holds its own error
channel. It would bite the moment a story routes table-editor refusals through the property channel.

**What this story does not touch, stated so a reviewer can check it cheaply.** No parse, no serialize, no
render, no format version. `headerStyle` already writes through the same `writeStyle`→`writeBorder` path as
an element border, which already emits partial blocks; a partial block already round-trips as a fixed
point. A document declaring none of this cannot move a byte, which is why AD-21 is untouched and why the
shape decision never needed the owner.

## Verification

Run every command from `folio-designer/`. ⚠ `npx vitest run` **from the repo root is a false mass-failure**
— it picks up the Playwright specs and runs them in the wrong environment.

**Commands:**
- `npx vitest run` -- expected: exit 0. Baseline measured by me at `c3e7a6b`: **76 files / 1368 tests /
  0 failures**. Report the new counts, and diff test names as a **multiset**, not a set.
- `npx tsc -b --force` -- expected: exit 0 and **0 bytes** of output.
- `npx oxlint` -- expected: exit 0, 0 errors. Baseline warning **SET** (never an integer, D-14.7.2):
  `src/App.tsx` ×2, `src/preview/pdf-viewer.tsx` ×2, `src/segmented-control.tsx` ×3 — all
  `react(only-export-components)`, 7 lines total. **Never change code to restore a count.**
- `npm run test:e2e:compile` -- expected: exit 0.
- **ADDED BY THE ORCHESTRATOR FOR THIS STORY ONLY**, run from `folio-go/`:
  `go test -count=1 -skip "^TestCorpusMeetsP6ExerciseFloors$" ./...` -- expected: exit 0.

  **Why it is here.** You correctly reported that the four designer commands cannot execute `folio-go`,
  so Parts 3, 4 and 7 — the closed set, the collapse mirror, the projection tags and the pair/wire ties —
  would have been verified only by CI, *after* the commit. **This story changes Go, so the story's own
  gate runs Go.** Measured at `ee225ea`: **9.6 seconds**, all packages green. That is too cheap to
  justify deferring the tie test that would catch a field-name/projection-key mismatch.

  **This is not a change to D-000.33 and does not need the owner.** The owner's cadence is the *minimum*
  a story must run; adding a suite strictly increases verification and removes nothing. It applies to
  **this story only**, because this story touches Go — do not carry it into a later dispatch on your own.

  **The invocation is copied from CI verbatim** (`.github/workflows/ci.yml:141`) and must not be
  re-derived. The `-skip` is not a convenience: `TestCorpusMeetsP6ExerciseFloors` is a **deliberate,
  anchored quarantine** (`KNOWN_RED_TEST`) for P6g's unmet opaque-name floor, and it is asserted to
  **FAIL** in its own `folio-go-known-red` job, which reds if it ever starts passing. **Never widen that
  filter, never remove the anchors, and never "fix" that test** — D-000.17 says a floor that is not met
  is reported unmet, never filled.

**Suites that deliberately do NOT run** (D-000.32 — name them in the report in these words): the browser
suite, the matrix legs, `npm run build` as a gate, the `verify:offline*` chain, and the font-host scans.

⚠ **"The Go suites" has been REMOVED from that list, and the report must be precise about why.** The
untagged Go suite **does** run for this story, per the fifth command above. Say that it ran, that it was
CI's exact invocation, and that the known-red stayed quarantined rather than silenced. Do **not** recite
the old list with "the Go suites" still in it — that would be a false absence claim, which is the shape
this project has been burned by most often. The **matrix legs** still do not run: they are `//go:build
matrix`-tagged and `./...` does not reach them.

**Manual checks:**
- `git diff --stat` shows no file under `folio-go/internal/template/` and no version constant.
- `git status --porcelain` shows `.working/TableEditor.dc.html` **unmodified**.
