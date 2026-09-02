---
title: 'Story 10.1: Text prints in a declared colour'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 1
followup_review_recommended: true
reconstructed: true
reconstructed_from_commit: '304442f97daf94d7e406fe405b3ce1116f4e95ce'
baseline_commit: '791ed002462807b71bb2223e1fe4f0900121f0e5'
baseline_revision: '791ed002462807b71bb2223e1fe4f0900121f0e5'
audit_revision: '0f4d62c1a8109c6f3382cbb8ea885e8063fa9e24'
context:
  - '{project-root}/_bmad-output/planning-artifacts/epics.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
  - '{project-root}/_bmad-output/implementation-artifacts/9-1-the-engine-paints-a-component-s-background-and-border.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-8-15-decision-log.md'
warnings: [reconstructed-after-the-fact, unreviewed-at-ship]
---

## ⚠ This spec was written after the code shipped

**Nothing here was authored before implementation.** Story 10.1 was implemented on 2026-08-30 in a
single commit, `304442f` *("Let a component's text print in a colour", 22 files, +479/−15)*, with
**no story file, no acceptance criteria, no review and no retrospective**. It sat at `review` in
`sprint-status.yaml` until this reconstruction, which exists — rather than the story simply being
accepted as-is — because **Story 15.3 cuts a release tag over this code** (owner decision D-000.5).

**The contract below was derived by reading the shipped diff and probing the shipped binary**, not
by recalling an intent. Where the original intent could not be recovered, this spec says so under
`## What could not be determined` instead of inventing one. The epic-level acceptance criteria in
`_bmad-output/planning-artifacts/epics.md` (`## Epic 10`) were written **at filing time on
2026-08-30, in the same session as the implementation**, so they are contemporaneous evidence of
intent but not an independent pre-implementation contract.

**Commit ownership — every hunk in `304442f` belongs to this story.** This is the one place Epic 10
differs cleanly from Epic 9: `791ed00` carried `text_alignment.go`, the font-family projection and a
stale Playwright manifest that belonged to no story, whereas `304442f` carries nothing but
`style.color`. `304442f` and `791ed00` both touch `page_setup.go` and `render.go`; the two are
separate commits, so each commit's own diff already separates them, and the split was **re-confirmed
at HEAD** rather than assumed:

| shared file | Epic 10's lines at HEAD (`74f10bc`, unchanged at `0f4d62c`) | proof |
|---|---|---|
| `page_setup.go` | `:266` (`CanvasComponent.Color`) and `:1738-1743` (`applyCanvasStyle`'s colour arm) — **and nothing else** | `git blame -L 266,266` and `-L 1738,1743` → `304442f9` on every line |
| `render.go` | `textRunSource.hasColor/color`, the `elementInk` call at `:908`, the per-run stamping, and the two `TextRun` fields in `buildShapedPDFRuns` | `git log -S "hasColor" -- folio-go/render.go` → **`304442f` alone**; `git log -S "elementInk(el.Style" -- folio-go/render.go` → `304442f` alone; `git blame -L 908,912` → `304442f9` |

**No hunk in `304442f` belongs to neither story.**

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Every document this product had printed came out black — not because black was chosen, but because
there was nowhere to choose anything else. The panel offered a typeface, a size, bold, italic and
two kinds of alignment, then stopped: the format could colour the panel behind the words and the
rule around it, but had no place at all for the words themselves. A heading, a total and a warning
all read as more of the same paragraph.

They can be told apart now. A component that says what colour its words are prints them that way,
and every line takes the same ink, because the colour belongs to the component rather than to
whichever fragment happened to be drawn. A table hands the ink to its cells as it hands down
everything else, and its heading row can be given its own. Saying nothing still means black, in the
strictest sense: the file carries no instruction about colour at all, so everything written before
this comes out byte for byte as it did.

Three things the code does not say out loud. A colour that is not a colour is refused, but only
where there are visible words to paint — so the same file passes or fails on the data it is handed.
The design surface will show you a colour the printer refuses. And a heading row told to have no
colour of its own does not fall back to the table's, the way every other setting does.

<intent-contract>

## Intent

**Problem:** `style` carries two colours — `background`, the box behind the glyphs, and
`border.color`, the rule around it — and no colour for the glyphs themselves. `pagemodel.TextRun`
has no ink field and `internal/pdf/textdoc.go` emits no colour operator, so the fill colour of every
run in every document this engine has ever produced is the PDF's own initial fill, black. The
inspector's TYPOGRAPHY section offers family, size, bold, italic and two alignments and cannot offer
a colour, because there is no field to bind one to.

**Approach:** Add **one** optional format field, `style.color`, and carry it end to end with no new
machinery: the format's third colour, behaving like the other two. `#RRGGBB` only; **no
colour-by-data** (the schema-derived placeholder fence covers it by construction, not by being told
to); validated **at render**, through `parseHexColor` — the module's one hex parser, the same one
`buildCellRectWithBackgroundField` and `validPropertyColor` already use. Resolve the ink **once per
element** and stamp it on every run that element produces, for the same reason `clipToBox` is
stamped that way: the colour is a property of the ELEMENT, never of one line or one face segment.
A table cascades it through `resolveHeaderStyle`/`resolveBodyStyle`, the arm every other cell
property already uses. `internal/pdf` brackets a coloured run's fill in `q`/`Q`, the discipline
`rectdoc.go`'s fill and stroke halves already use. **Absent emits nothing at all**, which is what
leaves the corpus byte-identical.

## Boundaries & Constraints

**Always:**
- **One hex parser.** `parseHexColor` is the only reading of "is this a colour"; the field default
  (`#000000` on the box path) is never restated for the ink path — an absent ink is *no operator*,
  not black.
- **Absent means no operator.** A run with `HasColor == false` emits no `rg` and no bracket, so a
  document declaring no colour is byte-identical to one rendered before the field existed (AD-21).
- **`q`/`Q`-bracketed.** A coloured run's fill never leaks into whatever draws next, and the bracket
  nests correctly inside the clip bracket `clipToBox` already opens (LIFO).
- **Resolved once per element**, stamped on every run — every line, every face segment.
- **No colour-by-data.** A `{{ }}` placeholder in `style.color` is a **load** error, and the fence
  that says so is derived from the schema by reflection, never hand-listed.
- **Fixed point throughout** (AD-2 / AD-23): `pagemodel.Color` is three `uint8`s and
  `appendColorChannels` converts through `geom.ScaleRound(…, 1000, 255)`. No `float64`/`float32`.
- **`HasColor bool` + `Color`, never `*Color`** — `pagemodel.go:71-77` states the reason: a pointer
  invites a float and an alias, exactly as `Rect.HasFill`/`HasStroke` already avoid.
- **No format-version thinking was applied.** See Finding 6 — it should have been.

**Block If:**
- A second reading of "is this a `#RRGGBB` colour" would be written anywhere.
- A golden digest moved.
- An ink would be resolved per line or per face segment rather than per element.

**Never:** a browser-side colour model · colour-by-data in any style field · a `*Color` in the page
model · a default black substituted at render time (that would make "unstyled" and "black"
indistinguishable in the emitted bytes).

## I/O & Edge-Case Matrix

*Every row below was measured at `74f10bc` unless marked otherwise; the probe is described under
`## Verification`.*

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Text element declares a colour | `style.color: "#c81e1e"` | Every run of that element carries `HasColor`; content stream emits `q\n0.784 0.118 0.118 rg\n` … `ET\nQ\n`. Other elements untouched | none |
| No colour declared | field absent | **No `rg` and no bracket anywhere** — measured: 8002 bytes vs 8028 with the colour | none |
| Present-null | `style.color: null` | Treated as no declaration; byte-identical to absent (8002 bytes) | none |
| Malformed colour, **visible text** | `"red"`, `""`, `"#GGGGGG"`, `"#c81e1eff"` | Located render error, `DiagCodeStyleColorInvalid`, naming the element and `style.color` | render fails; no bytes |
| Malformed colour, **element produces no visible text** | value `""`, value binds to `null`, or `visibleIf` false | **AS SHIPPED: renders clean, no diagnostic.** Finding 1 — the same file passes or fails on the data it is handed | none — silently accepted |
| `style.color` on a `rect`, `line` or `image` | any value, valid or malformed | **AS SHIPPED: loaded, serialized, round-tripped, never painted, never validated, never projected.** Finding 4. Contrast: `style.background: "red"` on the same element **is** a located render error | none |
| Table declares an ink | `style.color` | Cascades to header labels, every data cell and the footer aggregates — measured, 3 ink operators for a header + 2 rows | none |
| Table header overrides | `headerStyle.color` | Wins for the header row alone — measured, header `#1b2a4a`, rows `#c81e1e` | none |
| **`headerStyle.color: null`** with a base `style.color` | present, null | **AS SHIPPED: the header prints BLACK — the base colour does not fall through.** Every sibling field (`background`, `border`, `fontSize`, `padding`, `valign`, `align`, `fontFamily`, `lineSpacing`) *does* fall through on the same shape. Finding 2 — a defect, not a documented arm | none |
| Malformed table ink, **table has zero columns** | `style.color` or `headerStyle.color` = `"red"` | **AS SHIPPED: renders clean.** Finding 5 | none — silently accepted |
| Malformed table ink, one column | `"red"` | Located render error — but the field path is the literal `"headerStyle.color/style.color"`, naming **both** fields without saying which is wrong | render fails |
| Placeholder in the field | `style.color: "{{customer.brand}}"` | **Load** error naming the element and `style.color` | load fails |
| Over-long value at the projection | `style.color` longer than 512 bytes | `folio: component color exceeds the projection bound` | `Canvas`/`CanvasWithTextPaint` fails |
| **Non-hex value at the projection** | `"red"`, `""`, `"rgba(1,2,3,.5)"`, `"var(--x)"` | **AS SHIPPED: projected verbatim; the canvas paints it; `Render` refuses the same document.** Finding 3 — Epic 9's Finding 2 with the guard on the opposite end | none |
| Inspector commit of a bad colour | typed into the Text colour control | Refused: `color must be a #RRGGBB colour` (`validPropertyColor` → `parseHexColor`) | anchored at the control |
| Colour on a non-typographic selection | rect/line/image selected | The control is not offered, and `applyPropertyChanges` refuses `color` for those kinds | located command error |
| A coloured element also clipped | `clipToBox` true and `HasColor` true | Clip bracket opens first, colour bracket inside it, both closed in reverse order | none |

</intent-contract>

## Code Map

**Go — engine (`folio-go/`), at audit revision `0f4d62c`**
- `internal/template/model.go:250-262` — `Style.Color Presence[string]`, with the doc comment stating
  absent means the PDF's own initial fill and emits nothing.
- `internal/template/parse_bands.go:559-570` — `decodeStyle`'s `color` arm: `consumed["color"]`,
  present-null preserved, a non-string is a located load error. **No colour-shape check at load, by
  design** — the format checks a style colour string for a placeholder and for nothing else.
- `internal/template/serialize.go:348-354` — `writeStyle`'s `color` arm; `writeObject` sorts keys
  (`serialize.go:42`), so the field's position is derived, not authored.
- `folio_expr_validate.go:278-282` — `checkStyleHasNoPlaceholders`'s `color` arm: the no-colour-
  by-data fence.
- `folio_expr_validate_test.go:790-830` — `TestStyleStringFieldPopulationMatchesSchema`: reflects
  over `template.Style`/`Border`/`TableExt`/`Column` for `Presence[string]`/`Presence[[]string]`
  fields and asserts the checked set is exactly what remains after named exclusions, with an
  explicit **vacuity guard** (`len(all) == 0` → fatal). This is the "schema-derived fence" the commit
  message credits; it is real.
- `element_box.go:207-231` — `elementInk` / `styleInk`, this story's only new resolver. `styleInk`
  returns `(pagemodel.Color, bool, error)`; a malformed value becomes a `RenderError` carrying
  `DiagCodeStyleColorInvalid`.
- `render.go:158-168` — `textRunSource.hasColor` / `.color`.
- `render.go:908-911` — **the `elementInk` call site.** See Finding 1: it sits below four `continue`s
  (`:730` non-text, `:733` empty literal value, `:795` empty bound text, `:818` hidden) while
  `fontChain` at `:786` was deliberately placed above two of them.
- `render.go:902-905` — the per-run stamping inside the placed-line loop.
- `render.go:2374-2378` — `buildShapedPDFRuns` copies `HasColor`/`Color` onto `pagemodel.TextRun`.
- `internal/pagemodel/pagemodel.go:68-79` — `TextRun.HasColor` / `.Color`; `:253-255` — `Color` is
  three `uint8`s.
- `internal/pdf/textdoc.go:823-851` — the emitter: `q` + `appendColorChannels` + ` rg` before `BT`,
  `Q` after `ET`, inside the clip bracket.
- `internal/pdf/rectdoc.go:109-125` — `appendColorChannels`, shared with the box paint; fixed point
  via `geom.ScaleRound(geom.Length(c.R), 1000, 255)`.
- `table_render.go:277-283, 330-341` — `resolvedHeaderStyle.inkStyle` and its cascade arm. **The one
  arm in that function missing a `!…Null` guard** (Finding 2).
- `table_render.go:387-393, 414-417` — `resolvedBodyStyle.inkStyle` and its arm.
- `table_render.go:706-721` — the header ink, resolved **inside the per-column loop**.
- `table_render.go:784-793` — `bodyInk`, hoisted once per table; stamped at `:1031-1035` (data rows)
  and `:1206-1210` (footer aggregates).
- `table_render.go:468-495` — `parseHexColor` / `hexDigit`, the module's one hex parser. Strict:
  length 7, leading `#`, six hex digits.
- `component_commands.go:1068-1078` — the `color` property-command arm, gated through
  `validPropertyColor`; `:910-913` — `color` allowed for `text` and `table` only; `:1265-1268` —
  `validPropertyColor` = `parseHexColor`.
- `page_setup.go:266` — `CanvasComponent.Color *string json:"color,omitempty"`; `:1738-1743` —
  `applyCanvasStyle`'s arm: **length bound only, no colour-shape check** (Finding 3).

**Designer (`folio-designer/src/`)**
- `component-property-command.ts:3` — `'color'` added to `PropertyField`.
- `engine-protocol.ts:61` — `color?: string` on the projected component; `:142` — added to
  `hasOnly`'s key list; `:154` — added to the `optionalString` group (≤512 chars, any content).
- `App.tsx:1082` — `colorField`, `label: 'Text colour'`, `swatch: true`, `empty: 'black'`; rendered
  inside TYPOGRAPHY at `:913`, gated by `typographic` (`:1126`: every selected component is `text` or
  `table`).
- `App.tsx:1087` — BOX's border-colour row's affix renamed `'Colour'` → `'Border colour'`.
- `App.tsx:1058` — `swatchColor`: a non-`#RRGGBB` value renders the swatch as `#000000` with class
  `property-swatch-unset`.
- `App.tsx:1170` — `TextPaint` sets `--text-ink` only when `component.color !== undefined`.
- `App.css:76-79` — `.canvas-text-paint { color: var(--text-ink, var(--color-page-ink)); }`, with the
  comment stating the fallback is the engine's own no-declaration behaviour.

**Spec**
- `_bmad-output/specs/spec-folio/folio-format.md:439` (example), `:463` (defaults table), `:469`
  (placeholder-fence field list), `:47` (the `1.1` minor — **retrofitted later, not by this commit**;
  see Finding 6).

## Tasks & Acceptance

*Reconstructed. These describe what the shipped code **does**, verified by reading it and by the
probes recorded under `## Verification`. Where shipped behaviour departs from the epic's own words,
the AC says so rather than restating the epic.*

**AC1 — Every run an element produces takes the element's ink.** A text element declaring
`style.color` renders with `HasColor` on every run — every line and every face segment — resolved
once per element. *Shipped; `element_box.go:218-231` resolves once at `render.go:908`, stamped at
`:902-905`; `TestStyleColorInksEveryRunOfItsElementAndNoOther` asserts inked and plain runs coexist
and that the inked ones carry `#c81e1e`.*

**AC2 — An undeclared colour emits no operator at all.** *Shipped and measured: 8002 bytes with no
colour, 8028 with; zero `rg` and zero `q` in the uncoloured stream.
`TestAnUndeclaredColorIsAbsentRatherThanBlack` and `TestColouredRunBracketsItsInk` both pin it, the
latter by asserting the coloured stream **minus the bracket and operator** is byte-equal to the
plain one.*

**AC3 — The existing corpus is byte-identical.** *Shipped; **23/23 goldens hold** at `74f10bc`.
⚠ This is satisfied trivially: `grep -rln '"color"' fixtures/ folio-go/testdata/` returns **nothing**,
so no fixture declares the field and no golden exercises the new path. See Finding 7.*

**AC4 — A coloured run's fill is `q`/`Q`-bracketed.** *Shipped; `textdoc.go:826-838` and `:848-850`.
`TestColouredRunBracketsItsInk` asserts the prefix `q\n1 0 0 rg\nBT\n` and the suffix `ET\nQ\n`. The
bracket nests correctly inside the clip bracket (LIFO) by reading the emitter.*

**AC5 — A malformed colour is a located render error naming the element and the field.**
*Shipped **for a text element that produces visible text**: `DiagCodeStyleColorInvalid`, element id,
`style.color`, the offending value. **Not shipped** for an element whose text is empty, binds to
null, or is hidden (Finding 1); for a `rect`/`line`/`image` (Finding 4); or for a table with no
columns (Finding 5). On the table path the field name is the literal `"headerStyle.color/style.color"`,
which names both fields and identifies neither.*

**AC6 — A placeholder in `style.color` is a load error.** *Shipped, and shipped **by construction**:
the fence is reflection-derived over the schema with a vacuity guard, so the field was covered the
moment it existed. `TestAPlaceholderInStyleColorIsALoadError` asserts the error names `style.color`.*

**AC7 — A table cascades the ink to its cells; `headerStyle.color` wins for the header row.**
*Shipped for the positive cases — measured: base-only → 3 ink operators; base + header override →
`#1b2a4a` then `#c81e1e` twice. **Defeated for `headerStyle.color: null`**, which suppresses the ink
for the header rather than falling through to the base, unlike every sibling field (Finding 2).*

**AC8 — The inspector offers the colour in TYPOGRAPHY, and the canvas paints the engine's ink.**
*Shipped; `colorField` sits between the size row and the alignment grid, gated on `typographic`.
`App.test.tsx:869-883` asserts both halves — that `--text-ink` on `.canvas-text-paint` is the
projected `#c81e1e`, and that the commit emits exactly
`{"kind":"updateComponentProperties",…,"changes":{"color":{"op":"set","value":"#1b2a4a"}}}`.
⚠ For a **table** the colour projects but the canvas draws no cells at all, so the ink is invisible
there; that is Epic 14.9's gap, not this story's regression.*

**AC9 — Fixed-point discipline (AD-2 / AD-23).** *Shipped; `pagemodel.Color` is three `uint8`s and
`appendColorChannels` converts through `geom.ScaleRound`. Epic 10 introduces **no** `float64` or
`float32` anywhere; `internal/arch_test.go`'s AD-23 scanner is green.*

**AC10 (undeclared, and it should have been declared) — a new format field raises the document's
minimum format version.** ***Not shipped by this commit.** `304442f` added an optional field and
bumped nothing; documents using `style.color` continued to serialize `"version": "1.0"` while
requiring `1.1`. Retrofitted later under D-7.2.1 and pinned at HEAD. See Finding 6.*

## Review Triage Log

### 2026-09-02 — First review, at audit revision `0f4d62c` (source identical to `74f10bc`)

**This code had never been reviewed.** Findings are ranked and evidenced below; **none was fixed** —
this pass had no mandate to modify source. Every measurement was taken from an out-of-tree probe
module (working directory
`<scratch>/probe`, `replace` onto `/Users/panitw/Projects/folio/folio-go`) at `74f10bc`, `main`,
tree clean — source-identical to the recorded audit revision `0f4d62c`. Nothing was written into the repository.

---

**Finding 1 — VALIDATION PLACED BELOW THE SHORT-CIRCUIT A PRIOR REVIEW MOVED IT ABOVE (high).
A malformed `style.color` fails the render only when the element happens to produce visible text.**

`collectBandTextRuns` has four `continue`s before `elementInk`:

| line | condition |
|---|---|
| `render.go:730` | `el.Type != template.ElementText` |
| `render.go:733` | `!el.Value.Set \|\| el.Value.Null \|\| el.Value.Value == ""` |
| `render.go:795` | `boundText == ""` (the bound value resolved to null or empty) |
| `render.go:818` | `!elVisible` |
| **`render.go:908`** | **`elementInk(el.Style, …)`** |

Measured:

```
text, value "Hello",      color "red"   RENDER-ERROR: style.color "red" is not a #RRGGBB colour
text, value "",           color "red"   bytes=530  diags=0     ← accepted
text, value "{{a}}"→null, color "red"   bytes=530  diags=0     ← accepted
text, HIDDEN,             color "red"   bytes=530  diags=0     ← accepted
```

Control, the guard three lines above it in the same file:

```
text, value "",  fontFamily "nope"      bytes=530  diags=0
text, HIDDEN,    fontFamily "nope"      RENDER-ERROR: style.fontFamily "nope" names a chain
                                                     with no entries in the document's fonts map
```

`fontChain` is at `render.go:786` — **above** `:795` and `:818` — and it is there on purpose.
`render.go:778-785` is the comment that put it there:

> *QA Finding 5 (this story's review, Major): the fontFamily chain must be validated BEFORE the AC9
> empty-text short-circuit below, not after it. The previous ordering let `boundText == ""` skip
> font-chain validation entirely … the SAME broken template passing or failing depending on which
> report it was handed.*

`render.go:818-824`'s own comment makes the rule explicit for the visibility arm — *"Every validation
above this line has already run and succeeded for THIS element … so a hidden element with a broken
font chain still fails the render exactly as a visible one would"* — and `render.go:737-739` states
it a third time: *"elVisible decides ONLY whether this element's OWN diagnostics/output are emitted
below — it gates no validation call in this loop (R2/AC7)."*

**Story 10.1 put a new validation call on the wrong side of the line all three comments describe.**
The consequence is precisely the one the first comment names: a template with a broken
`style.color` renders successfully against one report's data and fails against another's. It is not
a content-stream hazard — `parseHexColor` is strict and `pagemodel.Color` is three `uint8`s, so
nothing out-of-range can physically reach the stream (this is where Epic 10 is genuinely better than
Epic 9's `-5 w`) — it is a *diagnosis* hazard, and it is the exact class this repository already paid
to fix once.

---

**Finding 2 — CASCADE ASYMMETRY, AND A FALSE COMMENT (high). `headerStyle.color: null` suppresses
the header's ink instead of falling through to the base, unlike every sibling field.**

`resolveHeaderStyle` (`table_render.go:334-340`):

```go
switch {
case hasHeader && header.Color.Set:
	r.inkStyle.Color = header.Color
case base.Color.Set:
	r.inkStyle.Color = base.Color
}
```

Every one of the other eight arms in that same function reads
`case hasHeader && header.X.Set && !header.X.Null:`. The colour arm omits `&& !header.Color.Null`,
so a present-null header value is *taken* (as a null), `styleInk` returns `hasInk = false`, and the
base colour never gets its turn. Measured on a table with `style.color: "#c81e1e"`, one column and
two rows (ink operators in the content stream):

| document | ink operators | header printed |
|---|---|---|
| `headerStyle` absent | 3 | `#c81e1e` |
| `headerStyle: {}` | 3 | `#c81e1e` |
| **`headerStyle: {"color": null}`** | **2** | **black** |

Controls, same null shape on the same table:

| document | result |
|---|---|
| `style.background`, `headerStyle: {"background": null}` | 3 fills — **falls through** |
| `style.border`, `headerStyle: {"border": null}` | 3 stroke groups — **falls through** |

The comment introducing the field (`table_render.go:279-281`) says the ink is

> *"cascaded exactly as its background is — `headerStyle.color` wins over `style.color`."*

**The first half is false, measured.** The second half is true. `resolveBodyStyle:414-416` has the
same missing `!Null` guard; there it is harmless, because there is no second level to fall through
to, but it is the same inconsistency.

The same false claim has since propagated into the format specification —
`folio-format.md:463` tells authors *"A table cascades it to its cells like every other cell
property"*.

Note this is **not** obviously a wrong *design*: "explicit null means suppress" is a defensible rule.
It is simply not the rule any other field in the file follows, it is nowhere written down, and the
one place that describes it describes it wrongly.

---

**Finding 3 — ENGINE/DESIGNER ASYMMETRY (high). The canvas accepts, projects and paints a colour the
engine refuses to render. This is Epic 9's Finding 2 with the guard on the opposite end.**

`applyCanvasStyle` (`page_setup.go:1738-1743`) bounds `style.color` to `maxCanvasPropertyString`
(512) and does nothing else. `parseHexColor` — the module's one hex parser, called by
`validPropertyColor` **nine hundred lines away in the same package** — is not called. Measured, via
`CanvasWithTextPaint` at `74f10bc`:

```
color "#c81e1e"        CANVAS  …,"color":"#c81e1e",…      RENDER  ok
color "red"            CANVAS  …,"color":"red",…          RENDER  ERROR
color ""               CANVAS  …,"color":"",…             RENDER  ERROR
color "rgba(1,2,3,.5)" CANVAS  …,"color":"rgba(1,2,3,.5)",…  RENDER  ERROR
color "var(--x)"       CANVAS  …,"color":"var(--x)",…     RENDER  ERROR
```

Downstream, `engine-protocol.ts:154` puts `color` in the `optionalString` group — any string ≤512
chars — and `App.tsx:1170` writes it straight into the `--text-ink` custom property, which
`App.css:79` resolves as `color: var(--text-ink, …)`. So the designer **opens, paints and lets you
keep editing** a document `Render` will not produce bytes for.

It is worse than a mismatch, because the inspector then disagrees with itself three ways for the
same value: `swatchColor` (`App.tsx:1058`) coerces a non-hex value to `#000000` and marks the swatch
`property-swatch-unset`, the text field shows `red`, and the canvas paints red.

Epic 9's Finding 2 was the mirror image — `canvasPropertyLength` refused a negative `borderWidth` on
the projection path while the render path emitted `-5 w`. Epic 10 has put the strict guard on the
render path and left the projection path open. **Both stories shipped one guard on one of the two
paths, and neither noticed the other end.** `background` and `borderColor` have the identical hole on
this path, so the pattern predates Epic 10 — but Epic 10 added a third instance of it while its own
sibling arm, four lines further down, was doing the same thing.

⚠ Not verified: whether a real browser paints `rgba(…)` and `var(--x)` as described, or drops them.
Playwright is excluded from this pass by the run's own cadence. The projection half is measured; the
paint half is read from the source.

---

**Finding 4 — FAIL-OPEN (medium). `style.color` on a `rect`, `line` or `image` is loaded, saved,
round-tripped, and then never painted, never validated and never shown.**

Measured:

```
rect, color "#c81e1e"           bytes=530  diags=0   ← accepted, nothing painted
rect, color "red" (malformed)   bytes=530  diags=0   ← accepted, never checked
rect, background "red" (CONTROL) RENDER-ERROR: style.background "red" is not a #RRGGBB colour
line, color "#c81e1e"           bytes=530  diags=0
line, color "red"               bytes=530  diags=0
line, background "red" (CONTROL) RENDER-ERROR: style.background "red" is not a #RRGGBB colour
```

`decodeStyle` is type-agnostic, `writeStyle` is type-agnostic, so the field survives a full
save/load cycle on a shape element. `applyCanvasStyle:1738` gates the projection on
`ElementText || ElementTable`, so the designer never sees it either — the value is invisible in
every surface and silently inert in the output. This is Epic 9's shape: *a value that means
"declared" to one layer and "absent" to another.*

`applyPropertyChanges:910-913` refuses `color` for those kinds, so the inspector cannot create this
state; only a hand-edited file or a third-party writer can. `element_box.go:213-216`'s comment claims
the field is *"validated at RENDER … exactly as `style.background` and `style.border.color` already
are"* — the control above shows that is **false for these three kinds**.

Whether "ignored on a shape" was a decision or an oversight is not recoverable; see
`## What could not be determined`.

---

**Finding 5 — FAIL-OPEN (medium). A table's ink is validated only where there is a column to paint;
and the located error names two fields at once.**

Measured:

```
headerStyle.color "red", table with NO columns    bytes=530  diags=0   ← accepted
style.color       "red", table with NO columns    bytes=530  diags=0   ← accepted
style.color       "red", table with one column    RENDER-ERROR: element e1:
                                                  headerStyle.color/style.color "red" …
```

`styleInk` for the header is called at `table_render.go:766`, **inside the per-column loop**, so a
table with no columns validates nothing. Two consequences:

1. Same class as Finding 1 — a malformed colour is accepted or refused depending on document shape.
2. The commit message's claim that *"the ink is resolved once per element"* is **true for the body**
   (`bodyInk` is hoisted at `:784-793`) and **false for the header** (once per column). The cost is
   trivial; the claim is not accurate.

Separately, the header path passes the literal field path `"headerStyle.color/style.color"`, so the
diagnostic names **both** fields and identifies neither — measured above, where the malformed value
was on `style.color` alone. The text path names `style.color` exactly. AC5's *"a located error
naming the element and the field"* is half-honoured on the table path.

---

**Finding 6 — FORMAT FIELD SHIPPED WITHOUT A VERSION BUMP (medium; already closed elsewhere,
recorded here because Epic 10's record does not contain it).**

`304442f` added an optional format field and raised no minimum version, so a document using
`style.color` continued to serialize `"version": "1.0"` while requiring `1.1`. Retrofitted later
under **D-7.2.1** and pinned at HEAD by `folio-go/line_spacing_test.go:717`, whose own case label
says it in as many words:

> `{"color, retrofitted", "color": "#1B2A4A", "1.0", "1.1", "Epic 10 shipped style.color and bumped nothing; documents using it declared 1.0 while requiring 1.1"}`

Also covered by `internal/template/linespacing_test.go:120-121`, and reflected in
`folio-format.md:47`. **Closed — no action.** It is recorded because it is a real Epic 10 defect that
exists nowhere in Epic 10's own paperwork, and because the epic text's premise — *"the format's
third colour, behaving like the other two"* — is precisely what made it easy to miss: the other two
colours are `1.0` fields and needed no bump.

---

**Finding 7 — COVERAGE (low). Nothing vacuous; several holes, and no fixture exercises the feature.**

What is asserted, and asserted well:
- `element_ink_test.go` (155 lines, 6 tests) asserts real properties — inked vs plain runs, exact
  channel values, absent-not-black, present-null, the located error's **code and element id**, and
  the load error's field name.
- `internal/pdf/textdoc_test.go`'s `TestColouredRunBracketsItsInk` is the strongest test in the
  commit: it asserts the prefix, the suffix, **and** that the coloured stream minus the bracket is
  byte-equal to the plain one — so it would catch a colour change that perturbed anything else.
- `folio_expr_validate_test.go`'s `TestStyleStringFieldPopulationMatchesSchema` is schema-derived by
  reflection and carries an explicit vacuity guard. The commit message's claim that the fence
  "picked it up on its own" is **true**.
- `canvas_body_text_bounds_test.go:368` covers the 512-byte projection bound for `color`.
- `App.test.tsx:869-883` asserts both the painted `--text-ink` and the exact command bytes.

What is not covered:
- **Finding 1** — no case for a malformed colour on an empty-valued, null-bound or hidden element.
- **Finding 2** — no case for `headerStyle.color: null`. `TestATableCascadesItsInkToHeaderAndCells`
  covers only the two positive arms.
- **Finding 3** — no test pairs the canvas projection against `Render` for a non-hex colour.
- **Finding 4** — no case for `style.color` on a `rect`, `line` or `image`.
- **Finding 5** — no case for a table with no columns.
- **No serialize/parse round-trip test for `style.color`** at all.
- **No fixture declares `style.color`**: `grep -rln '"color"' fixtures/ folio-go/testdata/` returns
  nothing. So the 23 goldens, and the cross-target byte-identity gate that runs after this pass,
  **never render a coloured document**. AC3's byte-identity holds trivially rather than being proved,
  and the emitter's bytes are pinned only by the unit test above, never by a golden.

---

**Finding 8 — COMMENT CLAIMS CHECKED AGAINST THE CODE (this run's standing hazard).**

| claim | site | verdict |
|---|---|---|
| "brackets a coloured run's fill in q/Q … a run with no colour emits no operator at all" | `textdoc.go:823-830` | **TRUE**, measured both halves |
| "resolved ONCE per element and stamped on every run" | `render.go:161-164` | **TRUE for a text element**; **FALSE for a table header** (once per column, Finding 5) |
| "cascaded exactly as its background is" | `table_render.go:279-281` | **FALSE**, measured (Finding 2) |
| "validated at RENDER … exactly as style.background and style.border.color already are" | `element_box.go:213-216` | **FALSE for rect/line/image** (Finding 4) and **conditional for text** (Finding 1) |
| "Spelled as a bool plus a value, never a `*Color` … a pointer invites a float and an alias" | `pagemodel.go:73-75` | **TRUE**; `Color` is three `uint8`s and no pointer exists |
| "--text-ink is set only where the engine projects a style.color" | `App.css:74-76` | **TRUE**; `App.tsx:1170` guards on `component.color === undefined` |
| "the format checks a colour string for a placeholder at load and for nothing else, by design" | `element_box.go:211-212` | **TRUE**; `decodeStyle`'s colour arm checks string-ness only |
| the epic's "so the whole existing corpus hashes identically" | `epics.md` Epic 10 | **TRUE but unearned** — no fixture declares the field (Finding 7) |

Two of eight load-bearing comments are false and a third is half-false. This run has now found five
false comments in this repository.

---

**Finding 9 — HOW EPIC 9's RULINGS LAND ON EPIC 10 (not a defect; a cross-reference the next lead
should not have to re-derive).**

While this review was being written, a parallel session committed **D-9.R** — the engineering lead's
nine rulings on Epic 9's review — at `f906e5f`/`0f4d62c`
(`_bmad-output/implementation-artifacts/epic-8-15-decision-log.md`). **Both commits are
documentation-only**: `git diff --stat 74f10bc 0f4d62c -- folio-go lint folio-designer fixtures` is
empty, so every measurement in this spec stands unchanged at `0f4d62c`. Three of those rulings bear
directly on findings above, and are recorded here as *evidence about the settled direction*, not as
rulings on Epic 10 — Epic 10's triage is the next lead's call, not this pass's.

1. **D-9.R.2 chose LOAD, not render, for a malformed template colour value** — grounded on the
   owner's settled split, *strict on templates, lenient on data*, and on "the bound already exists on
   the sibling path, so the designer refuses to open what the engine renders and nobody ever decided
   the engine accepts it." **That is Finding 3 verbatim, with the two paths swapped.** It is also the
   single change that would close **Findings 1, 3, 4 and 5 at once**: a `#RRGGBB` check on
   `style.color` at load makes the render-path placement (F1), the unguarded projection (F3), the
   shape-element hole (F4) and the no-column table hole (F5) all unreachable, because the value never
   enters the document. ⚠ It also **contradicts this story's own stated premise** and
   `folio-format.md:469`, both of which say a style colour's *shape* is deliberately a render
   concern. Whether D-9.R.2's grounding now overrides that premise for `style.color` too is a
   direction question, and it is flagged, not answered, here.
2. **D-9.R.0's discriminator** — a finding that is *the epic's stated goal, wrong* is on-goal;
   findings *along the way* register. Applied to Epic 10's own acceptance sentences: **Finding 2 is
   Epic 10's stated goal wrong** (the epic's second AC is the table cascade and `headerStyle.color`),
   and **Finding 5's error-path half** touches the fifth AC's "naming the element and the field".
   Findings 1, 3 and 4 are along-the-way by that test — though Finding 3 is the same *class* D-9.R.2
   ruled on-goal for Epic 9.
3. **D-9.R.5 forbids removing `omitempty`** as a fix for a projection divergence. Recorded because
   `CanvasComponent.Color` carries `json:"color,omitempty"` — but note it is a `*string`, so
   `omitempty` drops only a nil pointer and an **empty string still projects** (measured: `"color":""`
   appears in the projection). Epic 9's Finding 5 mechanism — `omitempty` silently eating a
   *meaningful empty value* — **does not recur on this field.** That is the check the brief asked for,
   and it comes back negative.

## Design Notes

**Why the ink lives on the element, not the run.** A face segment is an artefact of coverage
resolution and a line is an artefact of packing; neither is anything an author declared. Stamping
the colour once per element from the element's own style keeps the ink independent of how the text
happened to be shaped and broken — the same argument `clipToBox` already makes, and the reason a
multi-script coloured string does not change colour at a script boundary.

**Why absent is "no operator" rather than "black".** Emitting `0 0 0 rg` for an undeclared colour
would be semantically identical output and byte-different output. The corpus-identity guarantee is
what forces the distinction, and it is also what makes "unstyled" and "explicitly black"
distinguishable in the emitted bytes as well as in the file.

**Why `HasColor bool` + `Color` rather than `*Color`.** Stated at `pagemodel.go:73-75`: a pointer
invites a nil-vs-zero ambiguity and an alias, and the page model already answered this question the
same way for `Rect.HasFill`/`HasStroke`. It also keeps `TextRun` copyable and comparable.

**Why validation is at render rather than at load.** The format's position (`folio-format.md:469`)
is that a style colour string is checked at load for a placeholder and for nothing else — colour
*shape* is a render concern, so that `style.background` and `style.color` are refused in the same
place, by the same parser, with the same located-error shape. That reasoning is sound; Finding 1 is
about *where in the render* the check sits, not about render-vs-load.

**The renamed BOX row.** `borderColor`'s affix moved from `'Colour'` to `'Border colour'` because
TYPOGRAPHY's new row is also called Colour and the two sit one section apart. Recorded because it is
a user-visible label change carried in a feature commit.

## What could not be determined

Stated plainly, because "could not look" is not "all clear":

1. **Whether `headerStyle.color: null` was intended to suppress the header's ink** (Finding 2), or
   whether the missing `!Null` guard is a transcription slip from the eight arms around it. The
   comment beside it describes the *opposite* behaviour, which makes the record actively misleading
   rather than merely silent — but a misleading comment is not proof of intent either way.
2. **Whether `style.color` on a `rect`, `line` or `image` was meant to be ignored, refused, or
   painted** (Finding 4). The property command refuses it, which is evidence for "text-only"; the
   loader, the serializer and the placeholder fence all accept it, which is evidence for "carried
   opaquely". Nothing states a rule, and `folio-format.md:463` does not mention the restriction at
   all.
3. **Whether the validation placement in Finding 1 was considered.** The comment explaining why
   `fontChain` sits above the same short-circuit is 122 lines above the `elementInk` call, in the
   same function. That makes silence ambiguous between "not noticed" and "noticed and judged
   different"; the record cannot distinguish them, and no review artefact exists for this commit.
4. **Why no fixture was added** (Finding 7). Adding one would have moved no existing golden — a new
   fixture directory is additive. The commit message argues corpus identity from the four statement
   fixtures' unchanged byte counts, which is evidence the field is inert when absent, not evidence
   the emitter is right when present.
5. **Whether the format-version omission (Finding 6) was noticed at the time.** It was found and
   fixed later, by a different story, under D-7.2.1; nothing records who noticed or when.
6. **What the original acceptance criteria were.** The epic text in
   `_bmad-output/planning-artifacts/epics.md` was written in the same session as the implementation;
   there is no earlier draft in the repository.

## Stale documents tripped over

Recorded, **not corrected** — this reconstruction's mandate is limited to one spec file.

1. `_bmad-output/implementation-artifacts/sprint-status.yaml`, Epic 10's comment block: *"NOT yet
   reviewed or retrospected; the story file is not written."* Stale as of this spec. `epic-10:
   in-progress` and `10-1-text-prints-in-a-declared-colour: review` are likewise now behind the
   record.
2. `_bmad-output/specs/spec-folio/folio-format.md:463` — *"A table cascades it to its cells like
   every other cell property"* is **false as shipped** (Finding 2), and the entry does not say the
   field is ignored on `rect`, `line` and `image` (Finding 4). The false claim was copied from the
   code comment into the specification, which is the direction that does the most damage.
3. `_bmad-output/implementation-artifacts/sprint-status.yaml`, Epic 9's comment block still describes
   DW-24 as open. Already recorded as stale in Story 9.1's spec on 2026-09-02; still uncorrected.

## Verification

**Run at revision `74f10bc` (`main`, tree clean), 2026-09-02.** During the run a parallel session
committed two **documentation-only** commits, moving `main` to `0f4d62c`;
`git diff --stat 74f10bc 0f4d62c -- folio-go lint folio-designer fixtures` is **empty**, so no gate
number or probe below is affected and the audit revision is recorded as `0f4d62c`. See Finding 9. Every invocation's working
directory is recorded, per D-8.4j.8. **Per-epic cadence: the four `FOLIO_MATRIX_TARGET` legs,
`TestCrossTargetByteIdentity` and Playwright were NOT run** — those are Epic 9/10's boundary gate,
which comes after this pass.

| # | Command | Working directory | Result |
|---|---|---|---|
| 1 | `go test -count=1 ./...` | `/Users/panitw/Projects/folio/folio-go` | **1877 pass / 2 fail / 5 skip** — counted from `-json`; the two failures are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest only. Matches the documented permanent red exactly. |
| 2 | `go vet ./...` | `/Users/panitw/Projects/folio/folio-go` | no output, exit 0 |
| 3 | `go test -count=1 ./...` ×5 | `/Users/panitw/Projects/folio/lint` | **green 3 of 5, red 2 of 5** — every red is `internal/manifest`'s `TestCommittedAssetPopulationClassifiesCleanly`, the known cross-package race with `internal/rules`' `TestFontsAssetsNoticeRemovalRedProof` (Story 9.1's Finding 8). Identified, not a regression, not fixed. |
| 4 | `gofmt -l folio-go lint` | **`/Users/panitw/Projects/folio` (repo root)** | exactly `lint/internal/rules/licencegraph_test.go` — standing red. ⚠ From `lint/` this prints `lstat folio-go: no such file or directory` and reads as clean; it was run from the root. |
| 5 | `go run ./cmd/genmanifest` | `/Users/panitw/Projects/folio/lint` | `wrote /Users/panitw/Projects/folio/lint/MANIFEST.md` |
| 6 | `git ls-files --error-unmatch lint/MANIFEST.md && git diff --exit-code --stat -- lint/MANIFEST.md` | **`/Users/panitw/Projects/folio` (repo root)** | pathspec resolves; **empty diff, exit 0**. ⚠ `git diff <a> <b> -- <path>` returns empty patch text in this environment, so `--stat` was used and the pathspec was proved separately. |
| 7 | `npm run typecheck` | `/Users/panitw/Projects/folio/folio-designer` | **exit 0** |
| 8 | `npm run lint` | `/Users/panitw/Projects/folio/folio-designer` | **exactly 4** `react(only-export-components)` warnings — `pdf-viewer.tsx:16,17`, `App.tsx:1403,1410`. Standing red. |
| 9 | `npm test` | `/Users/panitw/Projects/folio/folio-designer` | **42 files passed / 432 tests passed**, exit 0 |
| 10 | `shasum -a 256 fixtures/*/expected.pdf` | **`/Users/panitw/Projects/folio` (repo root)** | **23 digests, all holding** |

**The 23 goldens hold** at `74f10bc`/`0f4d62c` — identical to the digests recorded in Story 9.1's spec at
`6e06cc7`:

```
986400a1…  alignment-rounding      e491d628…  alternating-rows
3283b81c…  component-asset-import  f533b04b…  embedded-font
a69a6653…  font-text               e5778eb8…  image-embed
6da3b12e…  justified-text          58ca4777…  justified-thai
6ed495b4…  keep-together           de212115…  line-spacing
7cf743de…  mandatory-break         0f925e1b…  minimal-rect
66ce0ee4…  multi-page              4699c8d7…  multi-script-fallback
b32fa1c5…  page-count-20           6c040ef7…  shaped-text
114df1d6…  statement-1             56bfbbd9…  statement-20
70dce051…  statement-5             5d090b0f…  statement-50
d5077f33…  thai-stacked-marks      746efcbc…  three-band-page
07c38cf7…  wrapped-text
```

**AD-23 check.** Epic 10 introduces no `float64` or `float32`: `pagemodel.Color` is three `uint8`s
and `appendColorChannels` converts through `geom.ScaleRound(geom.Length(c.R), 1000, 255)`.
`internal/arch_test.go`'s AD-23 scanner is green in run 1.

**Behavioural probe.** All Finding evidence above was measured from an out-of-tree Go module
(working directory `<scratch>/probe`; `go.mod` with
`replace github.com/panitw/folio/folio-go => /Users/panitw/Projects/folio/folio-go`) calling the
public API `ParseTemplate` / `Render` / `CanvasWithTextPaint`, decompressing the emitted content
streams and counting colour operators. **Nothing was written into the repository**, and the probe
carries no build tag or file inside the module.

**Tree state.** `git status --porcelain` from the repo root was **empty** before this spec was
written — including after the `lint` runs, so no interrupted run left `folio-go/fonts/notosans/NOTICE.md`
deleted — and shows only this one new spec file afterwards. **No source file was modified, nothing
was committed, nothing was pushed, no branch was created.**

## Delivery Log

### 2026-08-30 — shipped, unrecorded
Implemented and committed as `304442f` with no story file, no acceptance criteria, no review and no
retrospective. Left at `review` in `sprint-status.yaml`.

### 2026-09-02 — reconstructed and reviewed
Spec written retroactively from the shipped diff at audit revision `0f4d62c`. First review of the
code: **8 findings plus one cross-reference, none fixed** (this pass had no mandate to modify source). Findings 1, 2 and 3 are
fail-open, cascade-asymmetry and engine/designer-divergence defects with reproduced evidence and
should be triaged before Story 15.3 cuts a tag over this code. Finding 6 is already closed under
D-7.2.1. Gates re-run and recorded above; the 23 goldens hold.
