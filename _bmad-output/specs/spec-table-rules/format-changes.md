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

## Resolved — how far do the column rules run

**To the bottom of the box.** Ruled by the owner: *"the bottom most row will get the bottom border
from the table box border."* If the bottom line belongs to the box, the verticals must reach it, so
they span the box rather than the rows. The three readings this replaces were all consequences of
keeping a per-cell model: spanning edges that behave differently from non-spanning ones, a second
field that also draws vertical lines, or phantom rows that pagination and `altRowBackground` would
each have to be taught to ignore. A boundary-based `rules` block needs none of them.

## Not in this change

**Two-line header labels.** The target's headers are Thai over English in one cell. A column label
is shaped single-line and positioned directly, never packed (`folio-go/table_render.go:842`), so a
`\n` in a label is drawn as a rune no font covers — a `TEXT_MISSING_GLYPH` warning and no second
line. A label wider than its column is clipped **silently**, with no diagnostic, unlike a body cell.
Both are real gaps against the target and neither is a border. They want their own spec.
