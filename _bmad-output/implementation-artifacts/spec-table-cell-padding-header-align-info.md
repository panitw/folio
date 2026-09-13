---
title: 'Table cell padding, per-column header alignment, and info buttons in the Table Editor'
type: 'feature'
created: '2026-09-13'
status: 'done'
route: 'dispatch'
baseline_commit: '5bdfc7991e6e249295362978c48edbda8a393b74'
review_loop_iteration: 1
context:
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** In a printed table, left- or right-aligned cell text sits flush against the column rules. One `align` per column moves the header and the data together, so a column can't have, say, a centred header over right-aligned numbers. The Table Editor also explains proportion sizing and bindings/formulas only in loose sentences.

**Approach:** Put table-level cell padding (left/right) in the Table Editor, written to the table's existing `style.padding`. Add an optional `columns[].headerAlign`, which gives the matrix a HEADER ALIGN control next to CELL ALIGN. Put (i) buttons on the PROPORTION/WIDTH and BINDING column headers that open a short explanation.

## Boundaries & Constraints

**Always:**
- Owner revises D-12.4.1 (2026-09-13): the designer MAY author `style.padding.left/right` on a **table**, from the Table Editor only. Padding stays refused off a table, and the inspector keeps no padding rows.
- Padding goes through the existing element property command (`paddingLeft`/`paddingRight`, points, `set`/`clear`). No new Go command.
- Decision (owner, 2026-09-13): HEADER ALIGN is **always explicit**. Its segments only ever *set* `columns[].headerAlign`, and there is no clear or "follow cell" toggle. While unset, the pressed header segment is **the alignment the header actually prints**: the engine's resolved header cascade (`columns[].align` → `headerStyle.align` → `style.align` → `left`), never a browser-side guess (owner, 2026-09-13, from review). Changing CELL ALIGN or the table-wide header alignment still moves an unset header.
- Header cell alignment cascade: `columns[].headerAlign` → `columns[].align` → `headerStyle.align` → `style.align` → `left`. Data and footer cells are unchanged.
- `headerAlign` is its own closed set (`left`·`center`·`right`). Declaring it raises the document to **3.2**. A document without it serializes byte-identically at its current version.
- Each (i) button has an accessible name and a keyboard toggle. Escape closes an open explanation before it closes the dialog.

**Never:**
- No top/bottom padding control, and no change to `headerStyle.padding` authoring.
- No new default padding for new or existing tables.
- No full formula reference inside the dialog. A link to the docs comes later.
- No new `role="group"` in the Table Editor (keeps `control-vocabulary-contract` group floors meaningful).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Padding set | Left 3, Right 3 on a table | `style.padding` `{left:3,right:3}`; PDF cell text inset 3pt from the column edges, header too unless `headerStyle.padding` exists | Non-number → engine refusal shown, box restored |
| Padding cleared | Emptied box | edge key removed; empty `style.padding` dropped | N/A |
| Header padding overridden | table has `headerStyle.padding` | the controls edit the data rows; a note says the header row uses its own padding | N/A |
| Header align set | column align `right`, headerAlign `center` | header centred, data right; file gains `headerAlign`, version 3.2 | invalid token → located load/command error |
| Header align absent | legacy 3.1 document | renders and saves byte-identically | N/A |
| Info toggle | click/Enter on (i) by PROPORTION | explanation opens; click again, Escape or focus leaving closes it | N/A |

</frozen-after-approval>

## Code Map

- `/private/tmp/claude-501/-Users-panitw-Projects-folio/26b98a46-6122-45bd-bc54-8d6bd2154a5e/scratchpad/spec-table-padding-iteration0-KEEP.diff` -- the reverted iteration-0 implementation (verified green), a unified diff against `baseline_commit`. Apply it with `git apply`, then make only the changes the Tasks below add. See the Spec Change Log for what it got wrong.
- `folio-go/internal/template/model.go:527` -- `Column`; add `HeaderAlign Presence[string]`.
- `folio-go/internal/template/parse_bands.go:595-609` -- column `align` decode; copy it for `headerAlign` (mark `consumed`).
- `folio-go/internal/template/serialize.go:472-497` -- `writeColumn`; emit `headerAlign` (keys sorted globally).
- `folio-go/internal/template/closedsets.go:66,121` -- column align sets; document/reuse for headerAlign.
- `folio-go/internal/template/version.go:88,138,326-343,366,384` -- `SupportedVersion`, rank enum, `versionRequiredByContent`; add 3.2 rank for `headerAlign`.
- `folio-go/table_render.go:545,938` -- `columnAlign`; add a header variant used only at the header call site (covers repeated headers, L1015).
- `folio-go/page_setup.go:1960` -- canvas `HeaderAlign`; use the header variant.
- `folio-go/table_columns_projection.go:244-256,346,380` -- per-column projection; add committed `headerAlign` AND `headerAlignResolved` = `columnHeaderAlign(resolved.alignFallback, column)`, where `resolved := resolveHeaderStyle(*element)` already exists at ~L289. Use the renderer's own helper, never a copy of the cascade. (The table-level `headerAlign` at L97 is `headerStyle.align` — a different object. The `header*` pair rule in table_header_style_test.go ~L882 reflects only the TABLE struct, so the column-level name is free.)
- `folio-go/component_commands.go:498-586` -- `updateTableColumn`; add a set-only `headerAlign` field (validated like `align`, failure field `column.headerAlign`).
- `folio-go/component_commands.go:2979,3633` -- stale D-12.4.1 comments.
- `folio-go/internal/template/drift_test.go`, `fixtures_test.go:270-312,472`, `font_chain_variants_test.go:438-480`, `version_test.go:302` -- key/fixture/version pins to extend.
- `folio-designer/src/engine-protocol.ts:270,308,650,727` -- `TableColumn` type + exact-key validators; add column `headerAlign` (`''`|triple), column `headerAlignResolved` (triple, never `''`), and table `paddingLeft/Right` (committed).
- `folio-designer/src/table-column-command.ts:21` -- field union; add `headerAlign`.
- `folio-designer/src/component-property-command.ts:31,48,117` -- existing padding route to reuse.
- `folio-designer/src/TableEditor.tsx:39-42,304-349,721-763,842` -- lattice constants, `tabThroughMatrix`/`moveFocus`, header row & cells, existing table-wide "Header alignment" select (avoid name clash).
- `folio-designer/src/App.tsx:3220,3545,5432-5452` -- TableEditor wiring (`commitTableColumn`), D-12.4.1 comment, canvas table cells.
- `folio-designer/src/App.css:98-103,484-490,891` -- `.tool-hint` pattern, canvas cell padding, matrix grid tracks.
- Tests pinning shape: `TableEditor.test.tsx:540,542,596,622,854-870`; `App.test.tsx:308,806`; `control-vocabulary-contract.test.tsx:488-560,597`; `e2e/table-editor.spec.ts:42-44`; `e2e/table-matrix-layout.spec.ts:60,102-109,164`.

## Tasks & Acceptance

**Execution:**
- [x] Working tree -- `git apply` the iteration-0 KEEP diff (Code Map, first entry), then confirm `go test ./...` and `npx vitest run` are green before amending -- reuse verified work.
- [x] `folio-go/internal/template/{model,parse_bands,serialize,closedsets,version}.go` + tests -- add `columns[].headerAlign` with a 3.2 raise -- format support.
- [x] `folio-go/table_render.go`, `page_setup.go` + tests -- header cascade via header-only helper -- PDF and canvas agree.
- [x] `folio-go/table_columns_projection.go`, `component_commands.go`, `canvas_projection_wire_test.go` + tests -- project column `headerAlign` (committed) **and `headerAlignResolved`** (via `columnHeaderAlign`), plus table `paddingLeft/Right`; accept a set-only `headerAlign` update. The Go test must show a column with no `align` under `headerStyle.align: center` resolving `center` -- the editor presses what prints.
- [x] `folio-designer/src/engine-protocol.ts`, `table-column-command.ts` + tests -- protocol and command, including `headerAlignResolved`.
- [x] `folio-designer/src/TableEditor.tsx`, `App.tsx`, `App.css` -- cell padding L/R boxes near Total width; HEADER ALIGN + CELL ALIGN segmented controls (seven tracks, lattice/Tab updated). HEADER ALIGN presses `headerAlign` when set, else `headerAlignResolved`; its Tab stop is that pressed segment. Clicking the pressed segment of an unset header writes that value. (i) disclosure buttons with brief text: `aria-controls` present only while open, and a mousedown inside the open panel does not close it. Canvas cells use declared padding -- UI.
- [x] Designer unit/contract/e2e tests listed in Code Map -- update pins, and add tests for each I/O row plus: HEADER ALIGN presses `headerAlignResolved` when a table-wide header alignment differs from the column's; an `App.test.tsx` case committing Cell padding left `4` then clearing it, asserting the `updateComponentProperties` bytes target the table id (`set` 4, then `clear`); the (i) panel stays open on a mousedown inside it -- coverage.
- [x] Stale wording -- "six" columns in the `TableEditor.test.tsx` test title/describe, the `TableEditor.tsx` lattice comment and the `App.test.tsx` comment now say seven/fifteen -- review fix.
- [x] `_bmad-output/specs/spec-folio/folio-format.md:148,510-512,553,661`, `epic-11-14-decision-log.md` -- document `headerAlign`, 3.2 and the D-12.4.1 revision; the alignment-sets paragraph must read as whole sentences (no parenthetical splitting "rejected it … So") -- docs.

**Acceptance Criteria:**
- Given a table with Left/Right padding 4, when previewed, then data, footer and header text sit 4pt inside the column edges and the canvas shows the same inset.
- Given the matrix, when tabbing through a row, then order is header label → binding → proportion → header align → cell align → footer aggregate → next row.
- Given the dialog, when inspecting the PROPORTION and BINDING headers, then each has a named (i) button whose explanation covers proportion sizing / full bindings, formulas, Shift+Enter and Alt+Down.

## Design Notes

Header helper (engine):

```go
func columnHeaderAlign(fallback string, col template.Column) string {
	if col.HeaderAlign.Set && !col.HeaderAlign.Null {
		return col.HeaderAlign.Value
	}
	return columnAlign(fallback, col)
}
```

The loose "Proportion sizing · …" sentence moves into the PROPORTION (i) explanation. The binding sentences in `#table-editor-help` move into the BINDING one, but `#table-editor-help` stays as the bindings' `aria-describedby`.

## Verification

**Commands:**
- `cd folio-go && go test ./...` -- expected: pass
- `cd folio-go && gofmt -l .` -- expected: empty
- `cd folio-designer && npx tsc -b && npx vitest run` -- expected: pass
- `cd folio-designer && npx playwright test e2e/table-*.spec.ts` -- expected: pass

## Implementation Notes

- HEADER ALIGN reuses `SegmentedControl` with a new optional `role="toolbar"`: a `role="group"` per row would clear `GROUP_INSTANCE_FLOOR` (forbidden by Never). Its three segments and the two (i) buttons are therefore recorded in `V2_CENSUS`.
- Projection adds `columns[].headerAlign` (committed, `''` absent) and table `paddingLeft`/`paddingRight` (signed thousandths strings, `''` absent) plus `paddingHeaderOverride` (bool; not `header…`-prefixed because every `header*` key must be a committed/resolved pair).
- ~~The pressed header segment while unset is the projected column `align`.~~ Superseded at iteration 1: while unset, HEADER ALIGN presses the per-column `headerAlignResolved`, which Go projects through `columnHeaderAlign(resolveHeaderStyle(el).alignFallback, col)`, so the editor presses what the PDF prints.
- Canvas applies the table's declared left/right padding to both rows; it does not see `headerStyle.padding`, so an overridden header row paints with the table's padding.
- Moved focus pins (intended): Tab from Total width now reaches Cell padding left; Shift+Tab out of the matrix reaches the PROPORTION/WIDTH (i) button.
- `internal/text` `TestCorpusMeetsP6ExerciseFloors/P6g` fails identically on the baseline commit (pre-existing, unrelated).

## Spec Change Log

- **Iteration 1 (2026-09-13), from review triage #2 (intent_gap, medium).** *Finding:* while `headerAlign` was unset, HEADER ALIGN pressed the projected column `align` (default `left`), not the engine's header cascade. A column with no `align` under a table-wide `headerStyle.align: center` showed LEFT pressed while the PDF centred it, and clicking that "pressed" LEFT silently moved the printed header. *Owner resolution (frozen block):* press what actually prints. *Amended:* the Code Map and Tasks add a per-column `headerAlignResolved` projected through `columnHeaderAlign(resolveHeaderStyle(el).alignFallback, col)`, which the editor presses while unset. The Tasks also carry the five review patches (#9 `aria-controls`, #10 panel mousedown, #13 stale "six", #16 format-doc sentence, #18 App padding wiring test). *Known-bad state avoided:* the editor contradicting the PDF and a no-op-looking click that changes print. Implementation Notes' third bullet ("pressed … is the projected column `align`") is superseded. *KEEP:* everything else in the iteration-0 diff was verified green (Go, 1674 vitest, 23 table e2e):
  - the `columnHeaderAlign` helper at both the renderer and canvas sites
  - the 3.2 rank
  - the set-only `headerAlign` command at `column.headerAlign`
  - `paddingLeft`/`paddingRight`/`paddingHeaderOverride` on the table projection, with the TS guard admitting signed thousandths strings
  - `onCellPadding` routed through `updateComponentPropertiesCommand`
  - the `role="toolbar"` HEADER ALIGN control and its `V2_CENSUS` entries
  - the lattice offsets and `tabThroughMatrix` stops
  - the (i) text and the Escape-first handling
  - the canvas `inset`
  - the format doc and decision-log entries

## Review Triage Log

| # | Source | Finding | Verdict | Evidence | Route |
|---|--------|---------|---------|----------|-------|
| 1 | blind | version.go says a 3.1 reader silently aligns the header with the data; claimed a 3.1 reader refuses the key | false | `decodeColumn` passes unconsumed keys to `extraFields` (decodehelpers.go:116), which carries them opaquely, so a 3.1 reader keeps `headerAlign` and aligns the header with the data, as the comment says | reject |
| 2 | blind, edge (×2) | While `headerAlign` is unset, HEADER ALIGN presses the projected column `align` (defaults `left`, table_columns_projection.go), not the header cascade. With `headerStyle.align` set, the PDF header uses that alignment; clicking the "pressed" segment silently moves the printed header | medium | New columns get no `align` (addTableColumn), and `headerStyle.align` is authorable in the same dialog, so this is everyday. The frozen text equates "resolved alignment" with "the column's `align`", which diverge here | intent_gap |
| 3 | blind | Pressing an already-pressed header segment writes an explicit `headerAlign` with no visible change | false | This is the owner's "always explicit" decision (segments only ever set; no way back); the title text marks the followed state | reject |
| 4 | blind | Table-wide "Header alignment" select is silently overridden once a column has `headerAlign` | low | Real, but only after an explicit per-column choice; the fix adds new UI copy | reject |
| 5 | blind, edge (×3) | Canvas applies table padding to headings even when `headerStyle.padding` overrides it, and floors negative padding at 0 while the PDF uses it | low | `headerStyle.padding` can't be authored in the designer (hand-edited files only), and typing a negative padding is unlikely; the fix needs a new canvas projection member | reject |
| 6 | blind | No Go test for footer/right padding, for `headerAlign` on a continuation-page header, or for a footer ignoring `headerAlign` | low | Repeated header rows reuse `headerCells` (table_render.go ~L1026), and footer cells share `bs.padding` with data cells, so no untested path diverges | reject |
| 7 | blind | No test that a non-number is refused end-to-end, or that padding is still refused off a table | false | Off-table refusal is pinned by TestPaddingPropertyCommandsAreGrantedOnATableAlone (component_properties_test.go:746). A non-number encodes as `null` (component-property-command.ts:135), which Go refuses with a located message | reject |
| 8 | blind | Binding syntax left `#table-editor-help`, so screen-reader users lose it | low | The help text still points to the BINDING (i), which is keyboard-reachable; the move is as the Design Notes planned | reject |
| 9 | blind | (i) `aria-controls` references an element absent while closed | low | Panel renders only when open (TableEditor.tsx infoButton); direct correction | patch |
| 10 | blind, edge | Clicking or selecting text inside an open explanation blurs the button and closes it | low | The panel isn't focusable, so mousedown moves focus off the button and `onBlur` closes it; a one-line `onMouseDown` preventDefault on the panel fixes it | patch |
| 11 | blind | `role="toolbar"` is chosen to avoid the group-floor contract | low | A toolbar is a legitimate role for a button set; harm is limited to the hand-listed census entries | reject |
| 12 | blind | Closed-set test is weak; the token map and the command message are hand-copied | low | Mirrors the existing `align` pattern exactly (closedColumnAligns, the align command message) | reject |
| 13 | blind | Stale "six columns" wording: test title/describe, TableEditor.tsx lattice comment, App.test comment | low | The text still says six while the code asserts seven; direct correction | patch |
| 14 | blind | drift_test.go/version_test.go untouched; designer-tests task unchecked | false | drift_test passes via maximalFixture (updated); TestColumnHeaderAlignRequires32 covers the 3.2 rank; every task is now [x] | reject |
| 15 | blind | Byte-identity is tested with a 1.0 fixture, not a real 3.1 document | low | `versionRequiredByContent` only adds a rank checked on `HeaderAlign.Set`, so the 3.1 rules path is unchanged | reject |
| 16 | blind | folio-format.md alignment paragraph: the inserted parenthetical splits a sentence | low | Reads "…rejected it. (…) So a table's rejection…"; direct text fix | patch |
| 17 | edge | Padding above 2^53 thousandths makes the TS guard refuse the projection | low | Needs a hand-authored padding above 9e12 pt; the other length members share this bound | reject |
| 18 | verification-gap | App.tsx `onCellPadding` wiring is untested: no test checks that the command targets the table id or that clear sends `op: clear` | low | Pre-verified by the layer: every TableEditor test mocks `onCellPadding`, and no App/e2e test commits padding | patch |
| 19 | blind (iter 1) | folio-format.md heading "Alignment is three closed sets" and its cross-references now contradict "a fourth, separate declaration" | low | The section lists four sets, but the heading and its citations still say three; a direct doc correction | patch |
| 20 | blind (iter 1) | folio-format.md "`SupportedVersion` is `3.2` since `columns[].headerAlign` … renders with the new meaning" credits the 3.1 `rules` meaning change to `headerAlign` | low | That sentence belongs to the rules note, so the 3.2 edit made it misattribute; a direct doc correction | patch |
| 21 | blind (iter 1) | Spec Code Map/Tasks point at a session scratchpad diff; status mismatch | low | The fix would edit this build's spec | reject |
| 22 | blind, edge (iter 1) | Canvas applies table padding to headings under `headerStyle.padding`, and floors negatives | low | carried: same claim as #5, code unchanged there | reject |
| 23 | blind (iter 1) | No Go test for right/footer padding | low | carried: same claim as #6 | reject |
| 24 | blind (iter 1) | Closed-set test weak; token lists hand-copied in Go/TS | low | carried: same claim as #12 | reject |
| 25 | blind (iter 1) | `TestColumnHeaderAlignMovesTheHeaderOnly` accepts any header X in (0, 60000) | low | Still separates centred from right-aligned (the data run asserts > 90000); developer-only looseness | reject |
| 26 | blind (iter 1) | Padding projection reuses `committedBorderWidth` | low | Both are the same thousandths-string length spelling today; the harm is a hypothetical future border-only clamp | reject |
| 27 | blind (iter 1) | No visible set-vs-inherited state; clicking the pressed segment of an unset header writes a value (3.2, one edit) | false | carried: same claim as #3, the owner's "always explicit" decision; the pressed segment is now what prints, so that click never changes print | reject |
| 28 | blind (iter 1) | HEADER ALIGN `toolbar` vs CELL ALIGN `group` | low | carried: same claim as #11 | reject |
| 29 | blind (iter 1) | (i) buttons and open panels sit inside `role="columnheader"`, so the header's accessible name includes them | low | Real, but tolerable ("BINDING About bindings"); the fix restructures the header cells | reject |
| 30 | blind (iter 1) | Binding syntax removed from the grid's `aria-describedby` | low | carried: same claim as #8 | reject |
| 31 | blind, edge (iter 1) | `3.0` over a committed `3` sends a redundant set; a non-number refusal names `null`; a no-op accept leaves the non-canonical text in the box | low | Uncommon typing; the only harm is one extra undo step or a vaguer message; the fix adds number normalisation in the panel, which the file deliberately avoids | reject |
| 32 | blind (iter 1) | No range guidance for padding wider than its column | low | The engine already packs a column narrower than its padding (table_render.go contentW <= 0 path); pre-existing behaviour | reject |
| 33 | edge (iter 1) | WebKit doesn't focus a clicked button, so a mouse-opened (i) panel may not close when focus moves elsewhere | low | It stays closable by a second click or Escape; the fix adds an outside-pointer listener | reject |
| 34 | edge (iter 1) | While an (i) panel is open, Escape is captured with stopPropagation before another focused control can handle it | low | It only happens when the panel outlives its button's focus (#33); Escape then closes the panel and a second Escape works | reject |
| 35 | verification-gap (iter 1) | Canvas padding test would pass with left/right swapped or zoom ignored; no Go test that `Canvas()` carries a table's `paddingLeft/Right` | low | Pre-verified by the layer: assertions only compare relative values at zoom 1, and no `Canvas()` padding assertion exists | patch |
| 36 | verification-gap (iter 1) | e2e/table-matrix-layout.spec.ts still says "six tracks" and gives a stale pixel figure; e2e/table-editor.spec.ts header comment says six | low | Comments contradict the seven-track grid; a direct correction (#13 missed these files) | patch |
