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
  "unbreakableValues": ["customer.name"],
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
| `unbreakableValues` | *Optional.* A list of **bare root-relative dotted value paths** (e.g. `"customer.name"`) whose bound values must never be split across a line break — the same path convention `columns[].footerOf` uses: no `{{ }}`, no function call, no `[]`. Row-scoped paths are written root-relative under that same convention. The engine **never infers** membership; see *Line breaking* below. Declared once for the document because the property belongs to the data, not to a box. Absent means no value is protected. |

Top-level keys appear sorted, as does every object in the file — that is the serializer's job
(AD-9), not something an author maintains by hand.

> A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an
> existing key, and it may **not** extend a closed set of legal values (element `type`, `locale`,
> `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`).
> Extending a closed set is a **MAJOR** change, because every existing library validates those
> sets as load errors.

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
| `columns[].footer` | *Optional.* `sum` · `count` · `avg`. **Unchanged — names the operation only** (D-1.4.1); the numeric source is `columns[].footerOf`, below. Computed over the **whole collection**, never per page (AD-11). Omitted means no footer cell for that column. |
| `columns[].footerOf` | *Optional.* A bare root-relative dotted value path (e.g. `"transactions.amount"`) naming the numeric source the footer aggregates — no `{{ }}`, no function call, no `[]`. Legal only alongside `footer`, and never alongside `footer: "count"` (storing it would be a second source of truth against `bind`, AD-13). When `footer` is present and `footerOf` is omitted, it is **derived** from the column's own `bind`, but only when `bind` is one of exactly two syntactic shapes: (1) a bare row-scoped path `{{<alias>.<rest>}}` → `footerOf` = `<collection>.<rest>`; (2) a single `formatNumber(<bare row-scoped path>, <pattern literal>)` call → `footerOf` = `<collection>.<rest>` from the first argument, **and** `footerFormat` defaults to `<pattern>`. `<collection>` is the table's own `bind` with `[]` stripped. Any other `bind` shape is a load error — never a guess. **As of Story 1.4, this derivation is not yet implemented** — until Story 3.2 lands it, a `footer` with no `footerOf` simply loads, and the aggregate itself is not computed until Story 4.5. Story 3.6 mints the two diagnostic codes this eventually becomes: `TABLE_FOOTER_SOURCE_UNRESOLVED` (derivation failed) and `TABLE_FOOTER_SOURCE_FORBIDDEN` (an explicit `footerOf` conflicts with `bind`'s own shape) — neither exists yet. |
| `columns[].footerFormat` | *Optional.* A `formatNumber` pattern applied to the computed footer value. Legal with all three `footer` operations. |
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
  "a31f60866b4aa41953176fee9ddb90dc9bc53dce174421f8f567fac364c8bc27": {
    "data": [
      "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAIAAAD91JpzAAAAGklEQVR42mLhEpGTk5NjsbGxkZOT",
      "AwQAAP//CoABrYEc9NQAAAAASUVORK5CYII="
    ],
    "mediaType": "image/png"
  }
}
```

Keyed by the **lowercase hex SHA-256 of the raw bytes**, so identical images stored twice
deduplicate and emission order is stable. `data` is base64 hard-wrapped at 76 columns into an
array of strings — the file stays valid JSON, and the template's non-asset content stays readable
in a text diff (AD-9). Elements reference an asset by its key.

*(This is a real, supported, non-alpha PNG — canonically wrapped and independently verified; a test
keeps it byte-identical to this fence on every run, so it cannot silently drift. See the Delivery
Log for how it was derived.)*

Images are only ever embedded. Folio never fetches by URL and never reads from disk at render
time (FR33).

## Line breaking

Text wraps inside its element's declared `width`. Where a line may end is decided per script, and
what follows is the **whole** rule — this section is deliberately written as a list of what the
engine does **not** do as well as what it does, because every omission below is a deliberate
narrowing rather than an unfinished edge.

| Script | Rule |
|---|---|
| Latin, and anything with no rule of its own | A line may end **after** a run of whitespace, and nowhere else. The whitespace run is consumed by the break: it is drawn on neither line. |
| CJK | A line may end between any two adjacent Han or kana characters. |
| Thai | Thai is written without interword spaces, so break positions come from an embedded dictionary. A stretch of Thai the dictionary cannot account for is kept whole. |

**This is not UAX #14, and nothing in folio claims conformance to it.** Absent, by name:
hyphenation; a break at `-` or any other punctuation; the contextual pair rules that make up the
bulk of the standard.

**Kinsoku is not implemented.** A CJK line may begin with `，` or `。` and may end with an opening
bracket. Fullwidth punctuation and fullwidth digits are not break candidates at all.

### Vertical placement

Where an element's lines sit vertically is a function of the element's **declared font stack** and
its font size, and of nothing else — in particular it does **not** depend on which characters happen
to land on a given line. Adding one Chinese character to a paragraph never reflows it.

> **A stack declares what may appear in an element. Vertical placement must accommodate what may
> appear — not what does appear.**

"What does appear" is content-dependent, and content-dependent placement would make a box negotiate
with its contents. "What may appear" is exactly the declared stack.

**One rule, three spans, three maxima.** Write `A`, `D` and `gap` for a face's `hhea` ascent, the
absolute value of its `hhea` descent, and its `hhea` lineGap, each read from the face's own `hhea`
table and scaled to the font size. Then, over the faces of the declared stack:

| span | distance |
|---|---|
| top of the element → **first** baseline | `max(A)` |
| baseline → next baseline | `max(A) + max(D) + max(gap)` |
| **last** baseline → bottom of the text | `max(D)` |

**Each maximum is taken independently, over its own axis.** That is the whole of the rule and it is
easy to get subtly wrong: the natural-looking `max(A - D + gap)` — the largest single face — is
**not** the same quantity, and it is too small.

The space between two baselines has to hold the **descenders of the line above** and the
**ascenders of the line below**. Those are two different lines, and on a mixed stack they can resolve
to two different faces, so the constraint is the worst **adjacent pair**, not the worst single face.
On the shipped stack the two axes are won by different faces — Noto Sans Thai has the deepest
descender (450/1000 em) and Noto Sans SC the tallest ascender (1160/1000 em):

```
worst pair       = max(A) + max(D) = 1160 + 450          = 1610
largest one face = max(A - D + gap) = max(1362,1511,1448) = 1511
```

The single-face form is **99 units of the em short** — enough for a Thai line's below-vowels to
touch the next line's ideograph ascenders, on the default stack. For a stack resolving to a **single**
face the two forms are identical, since one face cannot fail to supply both axes.

The first baseline is placed by `max(A)` for the same reason and by the same argument, asked about
the ascent axis alone: the tallest thing that may appear on the first line must fit above it. It is
**not** the point size. The two coincide only by accident, and they diverge in **both** directions —
Noto Sans SC's ascent is 1160/em, so its first baseline sits *lower* than the point size implies,
while a face whose `hhea` ascent is below its em sits *higher*.

The cost is bounded by the author's own choice. A Latin-only element in a
`["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]` stack gets taller lines than Noto Sans alone would
need — but the author declared that stack, and an author who wants Latin metrics declares a
Latin-only stack. **No element pays for a face its own stack does not name.**

There is no `lineHeight` key and no first-baseline key. Vertical placement is derived, not authored.

### Values that must never be split

The engine does **not** guess which stretches of text are names. It cannot: Thai surnames are
coined, one per family, out of ordinary everyday words, so a dictionary genuinely cannot tell a
person's name from the words it was built from — `ศรีสุข` as a surname is character-for-character
the two common words `ศรี` and `สุข`.

So a template **declares** it, in the document-level `unbreakableValues` list. Every value
substituted from a listed path is kept on one line. Literal text around the placeholder is
unaffected — `"Statement for {{customer.name}}"` still breaks between *Statement* and *for*.

**The declaration protects bound values only.** A Thai name that appears inside free-form literal
text carries no declaration and remains breakable. That limitation is stated, not fixed.

If a value that must not be split is wider than its box, it **overflows visibly** rather than being
re-broken, squeezed or silently dropped.

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
            {
              "align": "left",
              "bind": "{{transaction.date}}",
              "id": "e3",
              "label": "Date",
              "width": 80
            },
            {
              "align": "right",
              "bind": "{{formatNumber(transaction.amount, \"#,##0.00\")}}",
              "footer": "sum",
              "id": "e4",
              "label": "Amount",
              "width": 90
            }
          ],
          "headerHeight": 16,
          "id": "e2",
          "style": {
            "border": {
              "edges": [
                "bottom"
              ]
            },
            "fontSize": 8,
            "padding": {
              "left": 3,
              "right": 3
            }
          },
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
          "style": {
            "align": "center",
            "fontSize": 7
          },
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
          "style": {
            "bold": true,
            "fontSize": 12
          },
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
  "fonts": {
    "body": [
      "Noto Sans",
      "Noto Sans Thai",
      "Noto Sans SC"
    ]
  },
  "locale": "th",
  "nextId": 6,
  "page": {
    "margin": {
      "bottom": 36,
      "left": 36,
      "right": 36,
      "top": 36
    },
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+07:00",
  "version": "1.0"
}
```

*(This example is generated by the module's own canonical serializer and cross-checked, structure, indentation and key order, against an independent pretty-printer — `json.dumps(..., sort_keys=True, indent=2, ensure_ascii=False)` — which reproduces it byte-identically with scalars normalised out, D-1.4.10, AC16. It is shipped as the golden fixture `folio-go/testdata/template/golden/worked-example.json`, and a test asserts this fenced block stays byte-identical to that file.)*

`{{page}}` and `{{pages}}` are **not** expressions and **not** a data namespace — they are the two
late-bound page-number slots, the only values in the format that depend on pagination (AD-4). No
`page` namespace exists for expressions to reach, and none may be added.
