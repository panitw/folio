# The `.folio` format

The field-level contract for a Folio template. `ARCHITECTURE-SPINE.md` governs *how* the file
behaves — canonical serialization (AD-9), element identity (AD-10), row scope (AD-11), locale
(AD-12), derived table geometry (AD-13), band-relative coordinates (AD-24). This document fixes
*what the fields are called*.

It exists because the format is a public contract, not an implementation detail: FR12 requires a
human or an AI agent to edit a template without opening the designer, and S9 requires a
hand-written template to render. Only one implementation of this schema will ever exist — the
engine owns the document and the designer never parses `.folio` (AD-15) — so this is a
readability contract, not a synchronization one.

> Supersedes the sketch in `addendum.md` §C, which predates the three-band layout model and shows
> a flat `"body"` array.

## Units

**Coordinates, sizes, margins and font sizes are in PDF points** (1 pt = 1/72 inch), written as
JSON numbers with **at most three decimal places**.

Three decimals is exactly one millipoint, the engine's internal unit (AD-2), so the conversion is
an exact ×1000 with no rounding. Values are parsed through the same exact-decimal path as report
data (AD-23) — never `float64`. A coordinate with more than three decimal places is a load error,
because it cannot be represented exactly.

Points rather than raw millipoints because a hand-editor writes `"x": 36`, not `"x": 36000`.

## Document

```json
{
  "assets": {},
  "bands": {},
  "fonts": {},
  "locale": "th",
  "nextId": 14,
  "page": {},
  "utcOffset": "+07:00",
  "version": "1.0"
}
```

| Field | Meaning |
|---|---|
| `version` | `"MAJOR.MINOR"`. A higher `MAJOR` than the library supports is a load error, never a best-effort render (FR13). |
| `locale` | One tag from the closed set `en`, `th`, `zh-Hans`, `ja`. An unlisted tag is a load error (AD-12). |
| `utcOffset` | Fixed offset, `±HH:MM`. The engine reads no host time zone. |
| `page` | Page setup (below). |
| `fonts` | Named font stacks (below). |
| `bands` | Exactly three bands (below). |
| `assets` | Embedded binary assets, keyed by content hash (below). |
| `nextId` | The next element-id counter value. Persisted so ids survive a save without renumbering (AD-10). |

Top-level keys appear sorted, as does every object in the file — that is the serializer's job
(AD-9), not something an author maintains by hand.

## `page`

```json
"page": {
  "margin": { "bottom": 36, "left": 36, "right": 36, "top": 36 },
  "orientation": "portrait",
  "size": "A4"
}
```

`size` is `"A4"`, `"Letter"`, or an object `{"height": 841.89, "width": 595.28}` for a custom
page. `orientation` is `"portrait"` or `"landscape"`.

## `fonts`

```json
"fonts": {
  "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]
}
```

Each key names an **ordered fallback chain**, tried left to right per glyph. `style.fontFamily`
references a key of this object, never a face name directly — so a chain is declared once and
reused, and the chain is part of the render's identity (AD-8). A glyph covered by no face in the
chain produces a diagnostic naming the element and the rune; it is never silently blank.

## `bands`

```json
"bands": {
  "content":    { "elements": [] },
  "pageFooter": { "elements": [], "height": 40 },
  "pageHeader": { "elements": [], "height": 80 }
}
```

Exactly these three keys (FR6). `pageHeader` and `pageFooter` declare a `height`; **`content` does
not** — its height is derived as page height minus margins minus header minus footer, by one
function (AD-13's sibling rule). Storing it would be a second source of truth.

Every element's `x` and `y` are relative to **its band's** top-left corner, never to the page
(AD-24).

## Elements

Common to all five types:

```json
{
  "height": 14,
  "id": "e7",
  "type": "text",
  "width": 200,
  "x": 20,
  "y": 12
}
```

| Field | Meaning |
|---|---|
| `id` | `e` + the counter in lowercase base 36 — `e1`, `ea`, `e1z`. Opaque: never derived from position or content, never reused, never renumbered on save (AD-10). Every diagnostic that concerns an element carries this. |
| `type` | `text` · `image` · `table` · `line` · `rect`. The set is closed (FR4); a sixth type is a load error. |
| `x`, `y`, `width`, `height` | Band-relative position and size, in points. **A `table` declares `x` and `y` only** — see below. |
| `visibleIf` | *Optional.* An expression; the element is absent from the page model when it evaluates false, and its siblings do not move (FR20, AD-24). Not valid on a table column. |
| `style` | *Optional.* See below. |

**`text`** — adds `"value"`, the string, which may contain `{{ }}` bindings.
**`image`** — adds `"asset"`, a key of the top-level `assets` object. The image is scaled to fit
its box preserving aspect ratio and centred (AD-24).
**`line`**, **`rect`** — no extra fields; both are drawn from `style.border` and
`style.background`.

### `table`

```json
{
  "as": "transaction",
  "bind": "transactions[]",
  "columns": [
    { "align": "left",  "bind": "{{transaction.date}}", "id": "e10", "label": "Date",   "width": 70 },
    { "align": "right", "bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}",
      "footer": "sum", "id": "e11", "label": "Amount", "width": 80 }
  ],
  "headerHeight": 16,
  "id": "e9",
  "type": "table",
  "x": 0,
  "y": 0
}
```

A table declares **`x` and `y` but neither `width` nor `height`**. Width is the sum of its column
widths (AD-13); height is however many rows the data produces. Storing either would let the
designer and the engine disagree.

| Field | Meaning |
|---|---|
| `bind` | The collection path, suffixed `[]`. |
| `as` | The row-scope alias. Optional; defaults to `row`. Inside the table, `<alias>.field` is the current row; unqualified paths still resolve from the document root (AD-11). |
| `headerHeight` | Height of the repeated header row, in points. Accounted for on **every** continuation page. |
| `columns[]` | Ordered. Each carries its own `id` (same counter as elements, so a diagnostic can name a column), `label`, `width`, `align`, and `bind`. |
| `columns[].footer` | *Optional.* `sum` · `count` · `avg`. Computed over the **whole collection**, never per page (AD-11). Omitted means no footer cell for that column. |
| `altRowBackground` | *Optional.* Colour for alternating rows (FR28). Alternation follows the row's index in the collection, so it does not reset per page. |

### `style`

Every field optional; omitted fields inherit the documented default.

```json
"style": {
  "align": "left",
  "background": "#F1F4F7",
  "bold": false,
  "border": { "color": "#000000", "edges": ["bottom", "top"], "width": 0.5 },
  "fontFamily": "body",
  "fontSize": 9,
  "italic": false,
  "padding": { "bottom": 2, "left": 3, "right": 3, "top": 2 },
  "valign": "top"
}
```

| Field | Default |
|---|---|
| `fontFamily` | the first key of `fonts` |
| `fontSize` | `10` |
| `bold`, `italic` | `false` |
| `align` | `left` · also `center`, `right` |
| `valign` | `top` · also `middle`, `bottom` |
| `padding` | `0` on all four edges |
| `border` | absent — no border drawn |
| `border.width` | `0.5` pt |
| `border.color` | `"#000000"` |
| `border.edges` | all four; a subset draws only those edges |
| `background` | absent — transparent |

Colours are `#RRGGBB`. There is no colour-by-data: conditional *visibility* is in scope,
conditional *formatting* is not.

## `assets`

```json
"assets": {
  "9f2b…c41d": {
    "data": [
      "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk",
      "YPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
    ],
    "mediaType": "image/png"
  }
}
```

Keyed by the **lowercase hex SHA-256 of the raw bytes**, so identical images stored twice
deduplicate and emission order is stable. `data` is base64 hard-wrapped at 76 columns into an
array of strings — the file stays valid JSON, and the template's non-asset content stays readable
in a text diff (AD-9). Elements reference an asset by its key.

Images are only ever embedded. Folio never fetches by URL and never reads from disk at render
time (FR33).

## What is *not* in the file

- **Report data.** Supplied by the caller at render time (FR14).
- **Sample data.** A separate local file the designer loads for binding discovery and preview
  (FR9). Whether its path persists in the template is open — PRD Q8, UX2.
- **Parameters.** Supplied at render time (FR21). The designer keeps an author-edited parameter
  document as a sibling file, passed as the third preview input (AD-16).
- **Dates.** Nothing in a template is stamped with a wall-clock time.

## Worked example

A minimal but complete template, in canonical form.

```json
{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {
          "as": "transaction",
          "bind": "transactions[]",
          "columns": [
            { "align": "left", "bind": "{{transaction.date}}", "id": "e3", "label": "Date", "width": 80 },
            { "align": "right", "bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}", "footer": "sum", "id": "e4", "label": "Amount", "width": 90 }
          ],
          "headerHeight": 16,
          "id": "e2",
          "style": { "border": { "edges": ["bottom"] }, "fontSize": 8, "padding": { "left": 3, "right": 3 } },
          "type": "table",
          "x": 0,
          "y": 0
        }
      ]
    },
    "pageFooter": {
      "elements": [
        {
          "height": 10,
          "id": "e5",
          "style": { "align": "center", "fontSize": 7 },
          "type": "text",
          "value": "Page {{page}} of {{pages}}",
          "width": 523,
          "x": 0,
          "y": 8
        }
      ],
      "height": 30
    },
    "pageHeader": {
      "elements": [
        {
          "height": 16,
          "id": "e1",
          "style": { "bold": true, "fontSize": 12 },
          "type": "text",
          "value": "Statement for {{customer.name}}",
          "width": 400,
          "x": 0,
          "y": 10
        }
      ],
      "height": 60
    }
  },
  "fonts": { "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"] },
  "locale": "th",
  "nextId": 6,
  "page": {
    "margin": { "bottom": 36, "left": 36, "right": 36, "top": 36 },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+07:00",
  "version": "1.0"
}
```

`{{page}}` and `{{pages}}` are **not** expressions and **not** a data namespace — they are the two
late-bound page-number slots, the only values in the format that depend on pagination (AD-4). No
`page` namespace exists for expressions to reach, and none may be added.
