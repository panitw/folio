# `.folio` format changes

Field-level changes for SPEC-table-rules. The base contract is
[`../spec-folio/folio-format.md`](../spec-folio/folio-format.md); everything not restated here is
unchanged.

**Target.** A ruled statement form: a thick outer frame, a vertical rule at every column boundary
running the full height of a fixed ruled area, a header row with a fill and a rule under it, no
horizontal rules between data rows, and a body that is mostly empty because the form is printed
before the data exists. Four capabilities are needed and one of them exists today.

| The form needs | Today |
|---|---|
| Header fill and a rule under it | **Works.** `headerStyle.background` + `headerStyle.border.edges: ["bottom"]` |
| An outer frame that is not a grid | `style.border` on a table is consumed as cell chrome — §1 |
| Rules the author controls, separate from the frame | No field exists — §2 |
| A ruled area taller than its rows | A table has no height at all (AD-13) — §3 |
| Header labels of two lines, Thai over English | A label is shaped single-line; `\n` draws as a missing glyph — §4 |

---

## 1. `style.border` and `style.background` on a table become the table's own box

**Breaking.** Today a table is excluded *by name* from the element box painter
(`folio-go/element_box.go:47`) and its `style.border` is stamped onto every cell rect instead —
header cells, every data cell of every row, and every footer cell. A 1pt border therefore prints a
full grid, and each interior rule is stroked twice because adjacent cells both own that edge.

After this change a table's `style` behaves as every other element type's does:

| Field | Was | Becomes |
|---|---|---|
| `style.border` | stroked around every cell | stroked once around the table's box |
| `style.background` | filled behind every cell | filled once behind the table's box |

The other `style` members are unchanged: `fontFamily`, `fontSize`, `lineSpacing`, `color`, `bold`,
`italic`, `align`, `valign` and `padding` still cascade into data cells, because they describe the
text inside a cell rather than the chrome around it.

**Documents that relied on the old meaning lose their grid.** This is deliberate and was chosen over
a load-time migration: a legacy arm in the loader would outlive everyone who remembers why it is
there, and the two meanings cannot both be the default. An author who wants the grid back declares
`rules: {"between": ["columns", "rows"]}` (§2) — which is not quite what `style.border` used to
mean, because the perimeter is now the box's and each interior line is drawn once.

**Every golden whose document sets a table `style.border` moves**, and that is the change being
witnessed rather than a casualty. A document that sets no table border must hash identically.

## 2. `table.rules` — the lines inside the table

*Optional.* The interior lines: a `width`, a `color`, and **which boundaries are ruled**.

```json
"rules": { "width": 0.5, "color": "#000000", "between": ["columns"] }
```

| Field | Meaning |
|---|---|
| `between` | Which boundaries get a line. `"columns"` rules every boundary between adjacent columns; `"rows"` rules every boundary between adjacent rows. Both, either, or `[]` for none. |
| `width` | Default 0.5pt, as everywhere else. |
| `color` | Default `#000000`, as everywhere else. |

`{"between": ["columns"]}` is the target form: the columns are ruled, the rows are not.

### The box owns its perimeter; the rules own the inside

**A rule is drawn once, at a boundary between two things, and never on the table's own edge.** The
outermost column has no outer vertical rule and the bottom-most row has no bottom rule — those two
lines are the box's left, right and bottom border (§1), and the box is the only thing that draws
them.

This is a ruling against the cell-chrome model it replaces, on two counts that are defects in it
today rather than matters of taste:

- **Today every interior line is stroked twice.** Adjacent cells each own the edge they share, and
  nothing de-duplicates them (`folio-go/table_render.go:672`, `internal/pdf/rectdoc.go:57`). Two
  strokes at one coordinate is not twice as dark in every renderer, but it is twice the content
  stream, and it makes the line's width a function of how many cells touch it.
- **Today the perimeter is drawn by whichever cells happen to be on it**, so the frame is as thick
  as the grid and moves when the row count changes. An author who wants a heavy frame around a fine
  grid cannot have one. Splitting the perimeter (§1) from the rules (here) makes those two weights
  two decisions.

`between: ["columns"]` therefore draws its verticals **from the top of the table's box to the
bottom of it**, through whatever empty ruled area `minHeight` (§3) creates — because the bottom of
a column rule is the box's bottom border, and a rule that stopped at the last row would leave the
form's columns hanging in mid-air. This is the whole reason `minHeight` and `rules` are one change
and not two.

The header row keeps `headerStyle.border` unchanged, which is what draws the rule under the header
in the target form. A `rows` rule between the header and the first data row would be a second line
at the same coordinate; the header's own border wins and the rules skip that boundary.

## 3. `table.minHeight` — a ruled area taller than its rows

*Optional.* A length in points. The table's **box** is at least this tall; rows are unaffected.

This **narrows AD-13**, which today rejects `height` on a table at load with "a table declares x and
y only — never height". That ruling stands for `height` and is what `minHeight` is carefully not: a
table's drawn extent is still *derived*, as `max(minHeight, header + Σ rows + footer)`. An author
cannot shorten a table by declaring a small `minHeight`, and cannot pin a row to a size. What they
gain is a floor, which is the whole of what a pre-printed form needs.

| Rule | Behaviour |
|---|---|
| Rows | Content-sized, exactly as now. `minHeight` never stretches, shrinks or pads a row. |
| The box | `style.border` and `style.background` (§1) are drawn to the box's full height. |
| Column rules | Drawn to the box's full height (§2), which is what the ruled area is for. |
| Pagination | The box's height is what a continuation page must accommodate. **A `minHeight` taller than a content window is a load error**, naming the element — the alternative is a table that can never be placed. |
| Absent / `null` | No floor; the box is the rows, which is today's behaviour. |

Declaring this key raises the document's `version`.

---

## 4. A column label may be more than one line

**No new field.** `columns[].label` is already an unbounded Unicode string; what changes is how it
is laid out.

Today a label is shaped with `breaksAreDrawn` and positioned directly, never packed
(`folio-go/table_render.go:842`), against a vertical model that is one line by construction —
`textBlockHeight := vm.FirstBaseline + vm.LastDescent` (`:880`) — with every glyph placed at line
index 0. So a `\n` in a label is handed to the shaper as a rune to draw, no font in the chain covers
it, and the author gets a `TEXT_MISSING_GLYPH` warning and one line.

A label is laid out **through the same packer a data cell already uses** — `breaksAreConsumed`, then
`packLines` against the column's content width (`:1066`, `:1077`) — never a second implementation of
the same rule. Consequences, all of them intended:

| | Was | Becomes |
|---|---|---|
| `"label": "วันที่\nDATE"` | one line, `TEXT_MISSING_GLYPH` | two lines |
| A label wider than its column | clipped to the padded box, **silently** | wrapped, like every other text in the document |
| `headerHeight` | the header row's exact height | its **floor** — see below |

### `headerHeight` becomes a floor

The header row is `max(headerHeight, the packed label's height + padding)`. The field stays
**required**, so no command can clear it, and it is still accounted for on **every** continuation
page — the labels are static, so the packed height is settled at layout, before pagination, and is
the same on every page a header repeats on.

This narrows the field the same way `minHeight` narrows a table's extent (§3), and for the same
reason: an author declaring a floor is declaring the form's proportions, not overriding what the
text needs. The alternative — keeping `headerHeight` exact and clipping a second line — reproduces
in the vertical the defect this section removes in the horizontal.

### The silent clip is retired, not relocated

A body cell that clips emits `DiagCodeTextClippedWidth` (`folio-go/table_render.go:1084`). The
header's clip path appends no diagnostic at all (`:908-921`), so an author whose column heading was
too narrow was told nothing, on screen or in the render. With a label that wraps and a header that
grows to fit it, there is nothing left to clip and nothing left to fail to report.

### What moves, and what must not

A document whose labels hold no `\n` and whose every label fits its column must **hash identically**
— that is the whole corpus standing as a witness. Two kinds of document do move, and both are
repairs: one whose label held a `\n` (was a warning and one line, becomes two lines), and one whose
label was too wide (was clipped in silence, now wraps and may grow the header row, which can change
where the table paginates).

### The designer must be able to type one

The Table Editor's header cell is an `<input>`, which cannot hold a line feed, and the command
bounds a label at 256 **bytes** (`folio-go/component_commands.go:523`) — about 85 Thai characters.
Authoring a two-line Thai/English label needs a control that accepts a break and a bound expressed
in characters rather than bytes. The canvas's `.canvas-table-heading` must paint the second line
too, or the canvas resumes lying about the header the way it currently does about the border.

## Resolved — how far do the column rules run

**To the bottom of the box.** Ruled by the owner: *"the bottom most row will get the bottom border
from the table box border."* If the bottom line belongs to the box, the verticals must reach it, so
they span the box rather than the rows. The three readings this replaces were all consequences of
keeping a per-cell model: spanning edges that behave differently from non-spanning ones, a second
field that also draws vertical lines, or phantom rows that pagination and `altRowBackground` would
each have to be taught to ignore. A boundary-based `rules` block needs none of them.

## Not in this change

**Soft wrapping is not opt-out.** A label now wraps rather than clipping, and there is no field to
ask for the old behaviour. If a form needs a label held to one line whatever its width, that is a
new field and wants its own argument — clipping in silence is not the precedent to preserve.
