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
| `version` | `"MAJOR.MINOR"`. A higher `MAJOR` than the library supports is a load error, never a best-effort render (FR13). **It describes the document, not the writer**: a file declares the lowest version its own content requires — `2.0` if any style sets `align: "justify"` (which only a **non-table** element's `style` can, see *Alignment is three closed sets* below) **or any chain in `fonts` declares an embedded-face entry** (see *`fonts`* below), else `1.2` if any element sets `keepTogether`, else `1.1` if any style sets `lineSpacing` or `color`, else `1.0` — and saving raises it to the **highest** requirement the document actually carries, never lowers it, and never stamps the library's own ceiling on a document that does not need it. They coexist: a document using none of those keys still declares `1.0` however new the library that wrote it. |
| `locale` | One tag from the closed set `en`, `th`, `zh-Hans`, `ja`. An unlisted tag is a load error (AD-12). |
| `utcOffset` | Fixed offset, `±HH:MM`. The engine reads no host time zone. |
| `page` | Page setup (below). |
| `fonts` | Named font stacks (below). |
| `bands` | Exactly three bands (below). |
| `assets` | Embedded binary assets, keyed by content hash (below). |
| `nextId` | The next element-id counter value. Persisted so ids survive a save without renumbering (AD-10). |
| `unbreakableValues` | *Optional.* A list of **bare root-relative dotted value paths** (e.g. `"customer.name"`) whose bound values must never be split across a line break — the same path convention `columns[].footerOf` uses: no `{{ }}`, no function call, no `[]`. Row-scoped paths are written root-relative under that same convention. The engine **never infers** membership; see *Line breaking* below. Declared once for the document because the property belongs to the data, not to a box. Absent means no value is protected. |

> `locale: "ja"` renders completely — the shipped font set has every Japanese glyph — but in
> Simplified-Chinese kanji SHAPES, because a font holds one drawing per codepoint and Chinese and
> Japanese share codepoints they draw slightly differently. This is legible and correct-content,
> never a missing-glyph error; it is a typography limitation, not a rendering failure. Supply your
> own face for Japanese typography — `FontSet` is a plain `map[string][]byte` (AD-8), and the
> font chain is always declared by the document, never inferred from `locale`.

Top-level keys appear sorted, as does every object in the file — that is the serializer's job
(AD-9), not something an author maintains by hand.

> A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an
> existing key, and it may **not** extend a closed set of legal values (element `type`, `locale`,
> `style.align`, `columns[].align`, `valign`, `columns[].footer`, `border.edges`,
> `page.orientation`, `page.size`).
> Extending a closed set is a **MAJOR** change, because every existing library validates those
> sets as load errors.

> **This has happened once. `style.align` gained `justify` (Story 7.3, FR47), and the format is
> therefore at `2.0`.** The alternative — a separate additive `style.justified` boolean, which a
> MINOR could have carried — was rejected deliberately: an older reader would have ignored the
> unknown key and drawn the paragraph ragged while believing it had rendered the document
> correctly, which is precisely the silently-wrong render the closed-set rule exists to prevent. A
> `2.0` document is *unreadable* to a `1.x` reader, and that is the honest outcome.

> **A SECOND MAJOR-CLASS EXTENSION JOINS THE SAME `2.0` (Story 8.3, FR53/FR56): a `fonts` chain
> entry may now be an object, `{"asset": "<key>"}`, and not only a string.** It is not a closed-set
> extension — it changes the legal SHAPE of an existing value — but it fails the same pre-reader
> test, and harder: a `1.x` reader decodes a chain entry as a string and never coerces, so it
> REFUSES the file outright rather than mis-drawing it. Declaring anything below `2.0` would be a
> version that lies: it would claim a reader sufficient for content that reader cannot load. It
> joins `2.0` rather than opening a `3.0` because `2.0` is not yet a released ceiling any reader has
> been shipped against, so nothing is orphaned twice.
>
> **The trigger is the ENTRY, not the asset.** A document that carries a font asset no chain
> references still declares whatever its other content requires: an unreferenced asset rides through
> a `1.x` reader as ordinary passthrough and renders correctly, so raising it would orphan a
> document from readers that can in fact read it.

### Alignment is three closed sets, partitioned by consumer

`align` is not one closed set, and it is not two. It is **three**, and they are keyed on **the code
that consumes the value**, which in this format means **the element type that owns the style block**:

| Set | Where it applies | Legal values |
|---|---|---|
| Style align | a **non-table** element's own `style.align` | `left` · `center` · `right` · `justify` |
| Table style align | a **table**'s `style.align` and its `headerStyle.align` | `left` · `center` · `right` |
| Column align | `columns[].align` | `left` · `center` · `right` |

A value outside its own set is a **located load error** naming the element and the field, and the
message lists exactly the members of the set that rejected it — so a table's rejection never names
`justify` as legal, and a text element's always does.

**Why the partition is by consumer and not by key path, which is the reusable half.** Story 7.3
split the original single set in two along the JSON key — `style`/`headerStyle` on one side,
`columns[]` on the other — intending to make justified table cells impossible. It did not, because a
table's `style.align`, its `headerStyle.align` and its `columns[].align` are all read into one
fallback and drawn by one set of cell switches: they are **one consumer wearing three key paths**.
So a table declaring `style.align: "justify"` loaded, paid the MAJOR version bump, and rendered
identically to `left` with no diagnostic — the author paid the whole cost of `2.0` and received
nothing (Story 7.8, DW-29).

> **When splitting a closed set, partition it by the code that consumes the value, not by where the
> value is written in the document.**

Justified table cells remain a scope decision that has not been taken. What changed at Story 7.8 is
that the format now *refuses* the declaration instead of accepting it and ignoring it.

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

**A chain entry has exactly two legal shapes**, and they may be mixed in one chain, in any order:

```json
"fonts": {
  "body": [
    "Noto Sans",
    { "asset": "9ab1e6c2f0d34b7a5c8e1f20d4b6a839c7e5024f1b8d63a09e4c7512fb3d8a6e" }
  ]
}
```

1. **A string** — the name of a face the renderer is given at render time (the shipped set, or one
   the integrator supplies). This is the only shape the format had before `2.0`.
2. **A one-key object `{"asset": "<key>"}`** — a face carried *inside the document*, whose value is
   a key of the top-level `assets` object. A document declaring one is a `2.0` document (above).

Anything else — a number, an array, an object with no `asset` key, or an object carrying any key
besides `asset` — is a **load error naming the chain and the entry's index**, e.g.
`fonts.body[1]`. An `{"asset": …}` entry whose key is **not present in `assets`** is likewise a
load error, and it names the chain, the index and the key: `fonts.body[1].asset`. A chain entry is
never silently dropped and never coerced.

**An entry naming an asset that is not a font is ACCEPTED AT LOAD and errors at RENDER.** The load
path checks that the key exists in `assets` and nothing else — it never inspects the asset's
`mediaType` or its bytes to decide whether the entry is legal. That is D-1.8.1 as amended, the same
rule the open `mediaType` set rests on (see *A font asset* below), and it applies to a *wrong-kind*
asset exactly as it does to an unrecognised font container: a chain entry naming an `image/png`
asset is a valid `.folio`. The failure arrives at render, and **only when something must actually
draw with that entry** — when a rune reaches it because no earlier entry in the chain covers that
rune. It is a located error naming **the chain, the entry's index and the asset key**. A document
whose text is covered entirely by the entries ahead of it renders clean and says nothing, because
nothing ever asked what those bytes were.

**Validation returns the identical error.** Validating that document — without rendering it —
returns the *same* error, with the same text and the same coordinates, and no diagnostics alongside
it. A validator that accepted a document the renderer refuses would be a second rule system, and the
one an author would trust is the one that says yes.

**Refusing is NOT what happens when a chain names a face nobody supplied, and the difference is
deliberate.** Given `["Noto Sans", <a non-font asset>, "Noto Sans Thai"]`, the render is refused at
the middle entry even though the entry after it covers every rune. Given
`["Noto Sans", "No Such Face", "Noto Sans Thai"]`, the middle entry is **skipped in silence** and the
third entry draws. The two conditions are not the same kind of thing:

- A chain entry naming a **non-font asset** is a defect *inside the document*. It travels with the
  file, it is wrong on every machine that will ever open it, and no deployment can make it right —
  so the moment something must draw with it, it is refused and located.
- A chain entry naming a **face the renderer was not given** is a property of *this* render, not of
  the document. The same file is correct wherever that face is supplied, and a fallback chain exists
  precisely so a document survives a host that is missing one of its faces (AD-8).

**A face is resolved by ASSET KEY, never by name.** An embedded entry's `font.family` is display
identity — what a chain editor shows a person — and is never used to resolve or substitute a face.
Where a document carries a face whose `font.family` is `"Inter"` and the renderer is also given a
face named `"Inter"`, the two are **different faces** and neither ever stands in for the other; the
chain entry's shape decides which one is meant (AD-8).

The map's keys have **no authored order**: they are sorted on write, like every other object in the
file (AD-9). Only the array *inside* a chain is ordered, and that order is the author's and is
preserved verbatim.

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

### Pagination

A document whose content is taller than one content band becomes **several pages**. The page header
and the page footer are drawn on **every** one of them, identically, at the same band origins: page
thirty-four is as complete as page one.

**The content band is a window onto one tall column.** The elements of `content` form a single
column of unbounded height; each page shows one page-height window onto it. A longer report is
**more windows**, never rearranged furniture. Four rules decide what a reader sees.

| | rule |
|---|---|
| 1 | **Where a page begins.** The first page's window begins at the top of the content band. Each later window begins at the top of the **first item that did not fit** in the window before it. |
| 2 | **What the unit is.** The unit that lands on a page is the **line**, not the element. A paragraph continues from the foot of one page to the head of the next. |
| 3 | **No line is ever split.** A line is drawn on the first page whose window holds it **entirely**, from the top of its tallest possible ascender to the bottom of its deepest possible descender. |
| 4 | **An image is atomic**, and the same rule applies to its **declared box**. |

**Whitespace at the foot of a page is correct.** Rule 3 means a line that would fall half on one
sheet and half on the next is drawn whole on the next sheet, and the space it vacated stays empty. A
statement cannot ship a half-line, and there is no setting that trades this away.

**Nothing moves sideways and nothing reflows to close a gap.** Every element keeps exactly the
position its author gave it within the column, so no element is ever displaced because a neighbour
grew (AD-24). One consequence follows directly and an author should know it before designing a
report rather than finding it in a diff:

> **Across a window boundary, declared vertical gaps collapse.** An element that begins a window is
> drawn at the top of its page, whatever gap was declared above it. That is the price of never
> splitting a line and never reflowing a sibling.

**No page is ever blank.** Because a window begins at the first item that did not fit rather than at
a fixed multiple of the content height, an element declared far below the preceding content starts
the next page instead of generating empty pages in front of it.

**An item that fits in no window is an error, not a surprise.** If a single line is taller than the
content band — a font size larger than the space available — or an image's **declared box** is, then
no window of any position can hold it, and loading or rendering the document **fails with an error
naming the element**. It is never drawn partly, never split, and never spilled past the page edge.

Both cases are decidable from the **template and the font set alone**, with no report data: line
height is a function of the declared font stack and font size (see *Vertical placement*), and the
window is page height minus the declared margins and band heights. So this is a fault in the
document, reported the same way to everyone who renders it, and not something one report's data can
trigger and another's cannot.

**A table row taller than the page is clipped, not refused — and the reason is authorship.** A row's
height is not something the author typed; it is whatever the record made it. One customer's address
runs to nine lines while every other customer's runs to two, and a statement run of a hundred
thousand documents can contain exactly one record that no page could hold. Refusing that document
would take down the run for a fault the author could not have seen when designing the template. So
an over-tall **table row** — a header row, a data row, or the footer row — is placed alone on a
fresh page, drawn as far down as the page has room for, and cut off there. The render **succeeds**
and returns the finished document, with a warning (`TABLE_ROW_CLIPPED_HEIGHT`) naming the table,
which row, the row's height and the height it was measured against. Whole lines are dropped, never
half of one, and the row's own rectangle stops at the page's content bottom.

The line between the two answers is therefore **who is responsible for the height**:

| the thing that is too tall | where its height came from | what happens |
|---|---|---|
| a table row (header, data or footer) | the **data** — the author may never have seen the record | clipped to a page of its own, **warning**, document produced |
| a declared keep-together group whose members each fit, but whose **union** does not | the **author's** own declaration that these elements travel as one | clipped to a page of its own, **warning**, document produced |
| a declared keep-together group holding an element that is by itself too tall | the **author's** own declaration — and removing the tag is the fix | **error**, no document |
| a line of a text element | the **author's** declared font size | **error**, no document |
| an image | the **author's** declared box | **error**, no document |

Folio absorbs what the data made too tall, and refuses what the author typed too tall — with one
deliberate exception, which is a *declared group* rather than a typed box. A set of elements the
author declared inseparable can add up to more than a page even though every member fits, and the
author has already said what should happen to it: keep it whole. Refusing the whole document at that
point would throw away a signature block for the sake of a rule about typed heights, so a group that
is too tall **only in aggregate** takes the row's answer instead — a page of its own, cut off at that
page's content bottom, and the same `TABLE_ROW_CLIPPED_HEIGHT` warning, worded for a group rather
than a row and naming the group the author declared. Whole members' lines and images are dropped,
never half of one: an image inside such a group is **removed, not moved**.

That exception reaches the *aggregate* and nothing else. If a single element of the group is by
itself taller than a content window, the document is **refused** and the error names that element.
What decides is **what** is too tall, never whether it happens to be tagged: a group of one adds
nothing, so tagging an element can never turn a refusal into a warning. A long text element the
author declares inseparable is refused for the same reason a too-tall image is — no page can hold
what was declared atomic. The difference from the untagged case is deliberate: untagged, that same
text element's lines simply split across pages and print in full, and the tag is what makes it
unsatisfiable, so removing the tag is the fix. A table row is the one thing that is never refused,
because its height comes from the data and its author has nothing to remove.

A typo in a template should still be found by the person who can fix it, at the moment they can fix
it; a pathological record should still not be able to stop a print run. Nothing is silent in any of
these directions — the clip always carries its warning, and the refusal always names its element.

There is no page-break key and no widow or orphan control. `keepTogether` (below) is the one thing an
author writes that pagination reads: it says which elements must not be separated, never where a
page ends. Where the pages fall is derived from the four rules above and from nothing else.

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
| `x`, `y`, `width`, `height` | Band-relative position and size, in points. **A `table` declares `x` and `y` only** — see below. For a **text** element, `width` bounds the laid-out content: content wider than the declared `width` is clipped at the box's left/right edges, never reflowed and never dropped, and a diagnostic names the element (FR44, Story 2.8). `height` on a **text** element is **not** a clip bound — content taller than the declared `height` renders in full and no diagnostic is reported, because no layout stage consults a text element's declared height. (`style.lineSpacing` does let an author set the leading, so a vertical bound is now something a template can be tuned towards by hand; it still is not something the engine checks the box against.) For an **image** element, `height` (together with `width`) is honoured: the image is scaled to fit the box and centred, never cropped and never stretched (AD-24), and is reserved for `valign` should a future story add one. |
| `visibleIf` | *Optional.* A bare expression (no `{{ }}` wrapping — see Expressions, below); the element is absent from the page model when it evaluates false, and its siblings do not move (FR20, AD-24; Story 3.5). Evaluated during bind, before pagination — it can never depend on the page an element lands on (AD-4). Condition semantics are `if()`'s own, unchanged: `true`/`false` decide visibility directly; an explicit `null` result is silently `false` (no diagnostic); a path absent from the data is a located Error; a string or a number is a located Error (no truthiness). A **field that is absent, or present with the JSON value `null`** (`"visibleIf": null`) both mean "no condition declared" — the element is visible, and there is nothing to evaluate; this is a *different* null from the condition **resolving** to `null` at evaluation, which is what hides the element. A bare literal (e.g. `"visibleIf": "42"`) can never resolve to a boolean and is rejected at **load**, naming the element (Story 3.5, closing the same-shaped rejection `if()`'s own condition slot already has). **Not valid on a table column — rejected at load, naming the column id** (Story 3.5; row-level visibility would make pagination a function of data, which FR25 does not define). |
| `keepTogether` | *Optional.* A string naming a **keep-together group**, e.g. `"keepTogether": "signature"`. Every content-band element carrying the same tag paginates as **one indivisible unit** (FR51): the whole set stays within the window it started in, or the whole set moves to the next one — each member still at its own declared position, with no sibling moved, no gap invented and no page left empty. The members need not be adjacent in the element list, and a tag is scoped to the document. **Content band only** — rejected at load on a `pageHeader`/`pageFooter` element, which is repeated verbatim on every page and never paginated — and **not valid on a `table`**, whose rows already carry their own grouping, rejected at load naming the element and the field. An absent field and an explicit `null` both mean "no group declared". A group taller than a whole content window **only in aggregate** — every member fitting, the sum not — is *clipped*, not refused; a group holding an element that is by itself taller than a content window is **refused**, naming that element. The tag is what makes such an element unsatisfiable, so removing it is the author's fix. See *Pagination*, above. Declaring this key raises the document's `version` to `1.2`. |
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
| `columns[]` | Ordered. Each carries its own `id` (same counter as elements, so a diagnostic can name a column), `label`, `width`, `align`, and `bind`. `columns[].align` is its **own** closed set — `left` · `center` · `right` — and does **not** admit `justify` (Story 7.3, D-7.3.1). Nor does a table's own `style.align` or its `headerStyle.align`, which feed the same cell alignment and therefore carry the same three values (Story 7.8). The sets are separate declarations so that extending one cannot legalise another by accident. |
| `headerStyle` | *Optional.* A `Style` block governing the header row ONLY — the same vocabulary as an element's own `style` (below) **except for `align`**, which admits `left` · `center` · `right` here and never `justify`, because a header cell is a table cell (see *Alignment is three closed sets*, above). A table's own `style.align` carries the same three values, for the same reason — never a data row. A field the header style leaves absent falls back to the table's own `style` for that field, then to that field's documented default (Story 4.1). `columns[].align` still wins over both for that column's own header cell (see `style`, below). |
| `columns[].footer` | *Optional.* `sum` · `count` · `avg`. **Unchanged — names the operation only** (D-1.4.1); the numeric source is `columns[].footerOf`, below. Computed over the **whole collection**, never per page (AD-11). Omitted means no footer cell for that column. |
| `columns[].footerOf` | *Optional.* A bare root-relative dotted value path (e.g. `"transactions.amount"`) naming the numeric source the footer aggregates — no `{{ }}`, no function call, no `[]`. Legal only alongside `footer`, and never alongside `footer: "count"` (storing it would be a second source of truth against `bind`, AD-13). When `footer` is present and `footerOf` is omitted, it is **derived** from the column's own `bind`, but only when `bind` is one of exactly two syntactic shapes: (1) a bare row-scoped path `{{<alias>.<rest>}}` → `footerOf` = `<collection>.<rest>`; (2) a single `formatNumber(<bare row-scoped path>, <pattern literal>)` call → `footerOf` = `<collection>.<rest>` from the first argument, **and** `footerFormat` defaults to `<pattern>`. `<collection>` is the table's own `bind` with `[]` stripped. Any other `bind` shape is a load error — never a guess. **As of Story 3.2, this derivation runs at load time** (`folio.ParseTemplate`) and the derived value is resolved alongside the document, never written back into it — a document that omits `footerOf` still serializes without it. **As of Story 4.5, the aggregate is computed** (`sum`/`count`/`avg`) and can be formatted (`formatNumber`), then rendered into the footer cell through the same expression evaluator used by ordinary bindings. Story 3.6 supplies the diagnostic codes: `TABLE_FOOTER_SOURCE_UNRESOLVED` (derivation failed) and `TABLE_FOOTER_SOURCE_FORBIDDEN` (an explicit `footerOf` conflicts with `bind`'s own shape). |
| `columns[].footerFormat` | *Optional.* A `formatNumber` pattern applied to the computed footer value. Legal with all three `footer` operations. |
| `altRowBackground` | *Optional.* Colour for alternating rows (FR28). Collection index zero retains the ordinary body treatment; the alternate colour applies to odd zero-based collection indexes (the second, fourth, sixth rows). Alternation follows that collection index, so it does not reset per page. Colour-by-data is out of scope for this field exactly as it is for `style`'s own colour fields (see "Colours are `#RRGGBB`" below): a `{{ }}` placeholder here is a **load error** naming the element, under the same rejection. `altRowBackground` wins over `style.background` on the rows where it applies; intervening data rows retain the ordinary body background or remain unfilled when none is declared. Headers and footers are not alternating data rows. |

As of Story 4.2, a **data cell** (every row the table's `bind` produces) cascades its font, border, background, padding, align and valign from the table's own `style` **only** — there is no `headerStyle` arm for a data row (`headerStyle` governs the header row exclusively, Story 4.1/D-000.76). `columns[].align` still wins over `style.align` for that column's own data cells, exactly as it does for the header. A table declaring `headerStyle.fontFamily` and no `style.fontFamily` therefore renders its header successfully and fails its data cells with the same "no resolvable `fontFamily`" error any other text-bearing element without one produces (there is no font default, above).

### Expressions

A `{{ }}` binding holds one expression (Story 3.2, AD-9): a bare dotted path (`customer.name`), a
function call over comma-separated arguments (`upper(customer.name)`), a double-quoted string
literal, or a number literal — nesting to any depth
(`formatNumber(sum(t.amount), "#,##0.00")`). The parser is hand-written recursive descent; there is
no operator (`+`, `-`, `==`, `&&`, …) anywhere in the grammar, and none will be added without a
direction change (AD-9, D-3.2.2 — every general-purpose expression library, including CEL, is
rejected on its numeric model: no exact decimal type).

There are, and will only ever be, **eight** named functions (FR18): `sum`, `count`, `avg`,
`formatDate`, `formatNumber`, `upper`, `lower`, `if`. This count is mechanically pinned in the
engine (`internal/expr`'s own closed table) — a ninth is a compile-time-enforced diff, not a runtime
surprise. All eight are implemented. Aggregate functions evaluate exact-decimal values over the
whole collection, and `formatDate`/`formatNumber` apply the declared locale rules; unsupported
function names remain located load/evaluation errors rather than silently wrong values.

**`upper(x)` / `lower(x)`** apply Go's Unicode case mapping. `x` must resolve to a string; any other
kind (including an absent or null path) is a located error, never coerced. A script with no case
distinction (Thai, CJK) is unchanged, byte for byte.

**`if(condition, then, else)`** takes exactly three arguments and evaluates only the branch it
selects (the other is never evaluated at all — an unimplemented function or a mistyped path in the
branch NOT taken produces no error). `condition` must resolve to a JSON **boolean** — there is no
truthiness anywhere in this grammar: a JSON `0`, an empty string `""`, and an empty array `[]` are
all the WRONG KIND for a condition, and each is a located error, exactly as any other wrong-kind
value is (AD-14) — never treated as false.

**An absent path as `condition` is a located error naming the path.** This is deliberately
different from the next rule:

**An explicit JSON `null` as `condition` silently selects the `else` branch — no error, no
diagnostic, no warning.** This is the one behaviour in the engine that leaves no signal anywhere in
its output: a reader of the rendered document cannot tell a section hidden by a null condition from
one that was simply never authored. This trade was made deliberately, with that cost stated plainly,
in preference to a warning that most template authors would never see. If a rendered document is
missing a section you expected, and its visibility is driven by `if(row.someFlag, …)`, check whether
`someFlag` can be `null` in your data — that is the one case this format will never flag for you.

### `style`

Every field optional; omitted fields inherit the documented default.

```json
"style": {
  "align": "left",
  "background": "#F1F4F7",
  "bold": false,
  "border": { "color": "#000000", "edges": ["bottom", "top"], "width": 0.5 },
  "color": "#1B2A4A",
  "fontFamily": "body",
  "fontSize": 9,
  "italic": false,
  "lineSpacing": 1.5,
  "padding": { "bottom": 2, "left": 3, "right": 3, "top": 2 },
  "valign": "top"
}
```

| Field | Default |
|---|---|
| `fontFamily` | **none — required on any element carrying text** |
| `fontSize` | `10` |
| `bold`, `italic` | `false` |
| `lineSpacing` | absent — the leading the declared font chain itself rules. A ratio scaling the baseline-to-baseline advance, and **only** that: the ascent above the first baseline and the descent below the last are untouched, so the ratio never re-measures a line, and a component's siblings never move. Under the default `valign` (`top`) the first baseline therefore stays exactly where it was. Note the one place the ratio is still visible in a first line's position: `valign: middle`/`bottom` seat the whole packed block inside the declared `height`, and a ratio makes that block taller, so the block is re-seated and its first baseline moves — measured at 11pt over two lines, `1.5` lifts a `bottom`-aligned first baseline by 7.491pt. That is `valign` doing its job on a taller block, not the ratio touching the first line. An exact decimal of at most three places, between `0.001` and `1000.0` inclusive; anything outside that, or a fourth decimal place, is a located load error naming the component — never a silent clamp. Values below `1` are legal and genuinely tight: one line's letters may reach into the line below, which is what tight leading is and what the page draws. |
| `align` | `left` · also `center`, `right`, `justify` — **`justify` is for a non-table element's own `style` only** (Story 7.3, FR47; narrowed at Story 7.8); `columns[].align`, a table's `style.align` and its `headerStyle.align` all keep the three-value set. `justify` flushes both edges by distributing the line's leftover width across its interior break opportunities, in whole millipoints: every gap receives `slack / gaps` and the first `slack mod gaps` gaps *in reading order* each receive one more, so the distributed amounts sum to the slack exactly and the last piece's right edge meets the declared `width` exactly. Three independent conditions leave a line ragged at the element's own start edge: it is the **last line** of the element; it was ended by a **mandatory break** the author typed; or it has **no interior break opportunity** to place slack in (an atomic unknown Thai run offers none). An element with no declared `width` has no box to justify to, and a line that meets or overflows its width has no slack — FR44's clip-and-warn applies unchanged. Declaring `justify` raises the document to version `2.0`. It is legal **only on a non-table element's own `style`**: a table's `style.align`, its `headerStyle.align` and its `columns[].align` all admit `left` · `center` · `right` alone, and `justify` at any of the three is a located load error naming the element and the field (Story 7.8 — see *Alignment is three closed sets*). So a table can never reach `2.0` through `align`. |
| `valign` | `top` · also `middle`, `bottom` |
| `padding` | `0` on all four edges |
| `border` | absent — no border drawn |
| `border.width` | `0.5` pt |
| `border.color` | `"#000000"` |
| `border.edges` | all four; a subset draws only those edges |
| `background` | absent — transparent |
| `color` | absent — the PDF's own initial fill, black. The INK the element's text prints in, as against `background`, the box behind it. A table cascades it to its cells like every other cell property, and `headerStyle.color` wins for the header row. Declaring none emits no colour operator at all, so a document written before this field renders byte-identically. |

There is no font default. An element with text and no `style.fontFamily` is a located error naming the element. A default was documented here from the format's first draft and never implemented; `fonts` is a mapping with no authored key order, so "the first key" was never well-defined. If a default is added later it will name its rule explicitly.

Colours are `#RRGGBB`. There is no colour-by-data: conditional *visibility* is in scope,
conditional *formatting* is not. As of Story 3.5, a `{{ }}` placeholder found inside any
string-valued style field (`align`, `background`, `color`, `fontFamily`, `valign`, `border.color`,
`border.edges`) — or in a table's own `altRowBackground` above, the format's one colour field
outside `style` — is a **load error** naming the element and the field — a component's condition
turns it on or off (`visibleIf`), it never changes how the component looks. This is not general
style-field validation: a hex colour, an alignment token or a font-family name is still accepted
with no format check at all — the rejection is narrowly "does this string contain a
placeholder", nothing broader. The covered set of fields is derived from the document model at
build time (every string-valued field of this shape on `Style`, `Border`, `TableExt` and
`Column`, minus a short, individually-justified exclusion list for fields that are identifiers or
closed enums rather than appearance properties — e.g. a column's own `align`, already governed by
its own closed `left`/`center`/`right` set, and a style's `align`, governed by the separate
four-value set that adds `justify`) rather than maintained as a second, hand-written list that can
silently drift from it.

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

### A font asset

An asset is not only an image. A **font face** is stored by exactly the same mechanism — same key
rule (the lowercase hex SHA-256 of the decoded bytes), same 76-column `data` wrapping, same
deduplication, same emission order — and differs only in its `mediaType` and in one additive,
optional record:

```json
"assets": {
  "9ab1e6c2f0d34b7a5c8e1f20d4b6a839c7e5024f1b8d63a09e4c7512fb3d8a6e": {
    "data": ["AAEAAAAP…"],
    "font": {
      "family": "Noto Sans Thai",
      "licence": "SIL Open Font License 1.1",
      "source": "https://github.com/notofonts/thai",
      "style": "Regular"
    },
    "mediaType": "font/ttf"
  }
}
```

`font` is **optional**, and every key inside it is optional. It is a record *about* the face for
the people reading and reusing the document — the `family` and `style` a designer shows in a chain
editor, and the `licence` and `source` that say where the bytes came from and on what terms. The
engine derives none of it from the bytes and none of it is required to render.

**`mediaType` remains an OPEN set for fonts exactly as it is for images.** It is not one of the
closed sets listed above and it never will be: under D-1.4.12 a closed set can only be extended by
a MAJOR bump, which would make every new font container a breaking format change. So, per D-1.8.1
(as amended):

- A **recognised** font media type whose bytes are not actually that format is a **load error** —
  the file lies about itself, and that is reader-independent.
- An **unrecognised** font media type (`font/woff2`, say, on a library that cannot decode it)
  **loads clean**. The document is valid and the asset is preserved verbatim. The failure, if any,
  arrives at **render**, and only when something actually needs to draw that face — a library
  capability limit, not a format error.
- A media type that is not a font type **at all** — an `image/png` asset named by a chain entry —
  takes the same path as the unrecognised one, for the same reason: it **loads clean**, and it
  errors at render only when something must draw with it. See *`fonts`* above for the shape of that
  error.

Both render-time refusals report the same thing, and it is a statement about **this build**, never
about the document: *the document is valid — `mediaType` is an open set — and this library cannot
draw with these bytes.*

## Line breaking

Text wraps inside its element's declared `width`. Where a line may end is decided per script, and
what follows is the **whole** rule — this section is deliberately written as a list of what the
engine does **not** do as well as what it does, because every omission below is a deliberate
narrowing rather than an unfinished edge.

Two kinds of break exist, and the difference between them is the difference between the engine
**guessing** and the engine **being told**:

- an **inferred** break is an *opportunity*. The engine proposes it from the text's script, and the
  line packer takes it only if the line needs it. The three script rules below are all of this kind.
- a **mandatory** break is *not* an opportunity. It is a line feed the author, or the data, put in
  the text, and the packer may not decline it (FR46).

### Inferred breaks — where a line *may* end

| Script | Rule |
|---|---|
| Latin, and anything with no rule of its own | A line may end **after** a run of whitespace, and nowhere else. The whitespace run is consumed by the break: it is drawn on neither line. |
| CJK | A line may end between any two adjacent Han or kana characters. |
| Thai | Thai is written without interword spaces, so break positions come from an embedded dictionary. A stretch of Thai the dictionary cannot account for is kept whole. |

### Mandatory breaks — where a line *must* end

A `U+000A` line feed in an element's text, or in a value bound into it, **always** ends the line —
however much width remained. It is the only character with this meaning.

- **Breaks are separators: *k* of them produce *k+1* lines.** `"a\nb"` is two lines. `"a\n\nb"` is
  three, the middle one empty — which is how a paragraph gap is expressed. `"a\n"` is two lines, the
  second empty; `"\na"` is two lines, the first empty; a value that is nothing but a line feed is two
  empty lines.
- **An empty line is a real line.** It draws nothing and occupies one full baseline-to-baseline
  advance, so it adds to the element's height and to a table row's height exactly as a drawn line
  does — and can therefore move a page break.
- **`\r\n` is one break, never two.** A carriage return carries no line feed of its own; a lone `\r`
  is ordinary whitespace and stays an inferred break.
- **The whole whitespace run around the break is consumed**, exactly as an inferred whitespace break
  consumes its run: `"a \n b"` is `a` / `b`, with neither space drawn on either line.
- **A mandatory break is not affected by `unbreakableValues`** — see *Values that must never be
  split*, below.

The engine never *invents* a mandatory break: it does not break at a declared width, at a hyphen, or
anywhere else on its own initiative. Only a line feed in the input produces one.

**This is not UAX #14, and nothing in folio claims conformance to it.** Absent, by name:
hyphenation; a break at `-` or any other punctuation; the contextual pair rules that make up the
bulk of the standard.

**No break falls inside a dictionary headword, including a lexicalised compound a native reader
would accept breaking.** The Thai engine matches against a shipped dictionary and never infers word
membership; where a compound word happens to be a headword, it is kept whole even if some readers
would break it. This is a stated capability limit, not a hidden one, and it is fail-closed: the
compound moves to the next line whole and is never rendered with a break inside it.

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
substituted from a listed path is kept on one line. This binds the break opportunities the engine
**infers** — from whitespace, a script, or the dictionary — not literal control characters present
in the input: a line feed the caller supplied is a break the engine was **told** about rather than
one it proposed, so it is still taken inside a declared value (FR46). Literal text around the
placeholder is unaffected — `"Statement for {{customer.name}}"` still breaks between *Statement* and
*for*.

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
            "fontFamily": "body",
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
            "fontFamily": "body",
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
            "fontFamily": "body",
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

These resolve in the page header and page footer bands. Elsewhere — in the content band — the
document fails to render, naming the element. (D-2.7.3)
