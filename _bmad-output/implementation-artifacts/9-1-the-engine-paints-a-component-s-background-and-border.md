---
title: 'Story 9.1: The engine paints a component''s background and border'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 1
followup_review_recommended: true
reconstructed: true
reconstructed_from_commit: '791ed002462807b71bb2223e1fe4f0900121f0e5'
baseline_commit: 'be40f248f3f7ff9749a54f4a070e54d8e5326300'
baseline_revision: 'be40f248f3f7ff9749a54f4a070e54d8e5326300'
audit_revision: '6e06cc784417bed50396ff742440a8fa0ef875fc'
context:
  - '{project-root}/_bmad-output/planning-artifacts/epics.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
warnings: [reconstructed-after-the-fact, shared-commit, unreviewed-at-ship]
---

## ⚠ This spec was written after the code shipped

**Nothing here was authored before implementation.** Stories 9.1 and 9.2 were implemented on
2026-08-30 in a single commit, `791ed00` *("Make a component's box print, on the page and on the
canvas", 22 files, +1649/−162)*, with **no story file, no acceptance criteria, no review and no
retrospective**. Both sat at `review` in `sprint-status.yaml` until this reconstruction. It is being
written now — rather than accepted as-is — because **Story 15.3 cuts a release tag over this code**
(owner decision D-000.5).

**The contract below was derived by reading the shipped diff**, not by recalling an intent. Where the
original intent could not be recovered from the code, this spec says so in place of inventing one;
those places are collected under `## What could not be determined`. The epic-level acceptance
criteria in `_bmad-output/planning-artifacts/epics.md` (`## Epic 9`) were written **at filing time on
2026-08-30, in the same session as the implementation**, so they are contemporaneous evidence of
intent but not an independent pre-implementation contract.

**Which parts of `791ed00` this story owns** (Story 9.2 owns the rest; a third group is owned by
neither):

*Owned by 9.1 — the engine box paint on the page:*
- `folio-go/element_box.go` (+137) — the whole collector.
- `folio-go/element_box_test.go` (+372) — its suite.
- `folio-go/render.go` — **only** the `collectElementBoxRects` call and the `tableRects` ordering
  comment inside `predictDocument` (`render.go:2074-2091` at the audit revision).
- `folio-go/internal/pdf/rectdoc.go` (+17/−5) and `rectdoc_test.go` (+29) — the `appendEdge` operand
  order fix, a latent Epic 4 defect surfaced by this story.
- `folio-go/page_setup.go` — **only** the `canvasElementIsPlaced` / `addCanvasWindowCount` arm that
  makes `elementDeclaresBox` the canvas's placement predicate.
- `_bmad-output/planning-artifacts/epics.md`, `_bmad-output/implementation-artifacts/sprint-status.yaml`
  — the epic filing.

*Owned by neither story, carried in the same commit:*
- `folio-go/text_alignment.go` (+89), `text_alignment_test.go` (+210) and their callers in
  `render.go`, `wrap.go`, `page_setup.go`. **This — not the box paint — is what moved the golden
  corpus.** Story 15.1 attributed it (D-15.1.1): all four statement goldens moved **+4 bytes per
  page**, from one `Tm` x-operand on the per-page footer element. The move is intended (D-R7.6) and
  the goldens are re-recorded. The box paint moved nothing.
- `folio-go/page_setup.go`'s `CanvasProjection.FontFamilies` / `DefaultFontSize` fields,
  `canvasFontFamilies`, `maxCanvasFontFamilies`, and the matching `engine-protocol.ts` validator
  additions. **These have no story, no acceptance criterion and no FR in `791ed00`** and are not
  mentioned in its commit message, which accounts for `page_setup.go` as text-alignment callers only.
  See Finding 6, and the attribution group below.

*Owned by neither story, and EXERCISED — the third group (D-9.R.6):*

The gap on these five is **attribution, not coverage**: D-000.5 already ruled the remedy for unstoried
shipped code, and each of them has a live consumer or a test that would fail if it were removed. They
are named here with that exercise beside them, the treatment `text_alignment.go` got from Story 15.1 —
**an attribution with no exercise named is the finding restated, not closed.** Epic 8's chain-editor
stories built on all five, which is why they were already exercised before anyone noticed they were
unstoried.

| Field / symbol | Where it is defined | What exercises it |
|---|---|---|
| `CanvasProjection.FontFamilies` | `folio-go/page_setup.go` | `folio-designer/src/App.tsx:903` passes `canvas.fontFamilies` into `ComponentProperties`; `:1112` is that component's own parameter; `:1145` hands it to `FontFamilyProperty`, the family picker in the TYPOGRAPHY section. Asserted on the wire in `folio-designer/src/engine-protocol.test.ts:9` (the validator's canonical projection fixture) and `folio-designer/src/sheet-stack.test.ts:16`. |
| `CanvasProjection.DefaultFontSize` | `folio-go/page_setup.go` | `App.tsx:903` passes `canvas.defaultFontSize`; `:1112` receives it; `:1145` uses it as the font-size field's `empty` placeholder, so an element with no declared size shows the engine's default rather than a blank. Same two test fixtures: `engine-protocol.test.ts:9`, `sheet-stack.test.ts:16`. |
| `canvasFontFamilies` | `folio-go/page_setup.go` | The producer of the field above — exercised through every `Canvas`/`CanvasWithTextPaint` assertion that reads `FontFamilies`, and on the TS side by `engine-protocol.test.ts:146` ("accepts the projected font chains only when they agree with fontFamilies"). |
| `maxCanvasFontFamilies` | `folio-go/page_setup.go` | Mirrored and pinned across the boundary by `folio-designer/src/engine-bounds-mirror.test.ts:80`, which asserts the Go constant and TS's `MAX_ENGINE_FONT_FAMILIES` name the same bound at the same site. |
| `engine-protocol.ts`'s `fontFamilies` / `defaultFontSize` validator arms | `folio-designer/src/engine-protocol.ts` | `engine-protocol.test.ts:146-190` and `:260` — agreement with `fontChains` in each of its three disagreement shapes, the bound at its edge, and order. |

- `_bmad-output/implementation-artifacts/evidence/story-6.7-roundtrip-manifest.json` (204 lines).
  Epic 8's boundary gate proved that manifest is stale and cannot have come from a real Playwright
  run — Playwright had **never executed** in this repository until 2026-09-02. **Noted, not fixed
  here; it is not Epic 9's to fix.**

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Set a fill and a border on something you have placed, and until this story neither appeared. The
file remembered both settings faithfully — you could save, close, reopen, and they were still there
— but nothing on the way to the printed page ever asked for them. The one exception was a table,
whose fill and border had always been drawn as the shading and ruling of its own cells. Everything
else carried the settings and printed as though it had none. The panel that offered those settings
was, in effect, promising something the engine never delivered.

This story makes the promise good. Anything you have placed that declares a fill or a border now
prints that box, in the position and at the size you declared, behind its own words or picture
rather than over them. A box in a repeated band prints on every page. A box in the flowing middle of
the document travels with everything else in that column, breaking across pages by the same rules
that already move a table. Something you have hidden by condition contributes no box, and leaves no
gap where it would have been. A table is deliberately untouched, so its shading is never drawn twice.

Underneath all this sat an older mistake. When a border was asked for on only some of its sides, the
instructions written into the file put the numbers on the wrong side of the words that consume them.
Four sides came out as diagonals crossing the box. No printed document in the reference set had ever
asked for a partial border, so nothing had ever caught it. It is now written the right way round,
and a test reads the order rather than merely counting the pieces.

<intent-contract>

## Intent

**Problem:** `style.background` and `style.border` are parsed, validated for placeholders and
round-tripped by the format, but the **only** consumer on any render path is `table_render.go`,
which resolves a TABLE element's own style into per-cell chrome. On a text, image, rect or line
element both fields are inert. Worse, rect and line reach the page model not at all — their
visibility verdicts are computed and unconsumed (`render_visibility.go`'s "LATENT, not broken").
The designer offers all four BOX controls on all five components, so the panel promises what the
engine does not deliver.

**Approach:** Add **no new machinery**. A box is a rect group with a band, an element id and a
vertical extent — which is exactly `tableRectSource`, the carrier Story 4.1 already built and that
`paginateDocument`, `contentColumnItems` and the page assembler already place, repeat, shift and
clip. Collect one `tableRectSource` per element that declares a box, built through
`buildCellRectWithBackgroundField` — the **same** builder a table cell uses — so "what does
`style.border` mean" keeps exactly one implementation (D-000.38's rule applied to geometry).
Exclude TABLE by name, not by omission. An element declaring no box contributes no source at all,
which is what leaves the corpus byte-identical.

## Boundaries & Constraints

**Always:**
- **One implementation of the border rule.** The width default (0.5pt), colour default (`#000000`)
  and the four-edge default come from `buildCellRectWithBackgroundField` and are never restated.
- **Element boxes lead the rect slice**, then header, content and footer table chrome, so a box
  painted behind a table sits *under* that table's cell chrome. Within each population the order is
  `documentBands`' band order then declaration order — the emitted byte order.
- **A hidden element contributes nothing** — AD-24's Visibility clause, checked before any geometry.
- **A table is excluded by name.** Its style already paints as cell chrome.
- **Fixed point throughout** (AD-2 / AD-23): `geom.Length`, int64 millipoints, `geom.ScaleRound`'s
  round-half-to-even. No `float64` or `float32` anywhere under `internal/`.
- **The 23 golden digests stay byte-identical *for the box paint*.** (The commit as a whole moved
  four of them; that move is `text_alignment.go`'s, attributed by Story 15.1 — see the ownership
  note above.)
- **No new format field, no new property, no format-version move.** Every value consumed is one the
  format already carries, validates and round-trips.

**Block If:**
- A second reading of `style.border` would be written anywhere.
- A golden digest moved and the move traced to the box paint rather than to `text_alignment.go`.
- `pagemodel.Rect` would gain a field, or `appendRectContentStream` a second emitter.

**Never:** a browser-side model of style · inferring a box from measured content (a box drawn from
anything but a declared rectangle makes the box a function of the data) · painting a table's style
twice · `float64`/`float32` under `internal/`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Background on a text element | `style.background: "#112233"`, declared w/h positive | One `pagemodel.Rect` with `HasFill`, at the element's declared box, emitted **before** the element's own text | none |
| Border on any eligible kind | `style.border` present, non-null | One `Rect` with `HasStroke`, width/colour/edges resolved by `buildCellRectWithBackgroundField`'s defaults | none |
| Both at once | background **and** border | One `Rect` carrying fill and stroke; fill and stroke are `q…Q`-bracketed independently in the content stream | none |
| Box in page header / page footer | any eligible kind | Repeats on every page, exactly as that band's text already does | none |
| Box in the content band | any eligible kind | Travels as its **own ungrouped** content item (`isDataRow`/`isHeaderRow`/`isFooterRow` all false), so no row displacement is ever applied to it | none |
| Hidden element | `visibleIf` resolves false | No box at all; absent from the page model, no gap (AD-24) | none |
| Table element | declares background/border | Unchanged — painted once, as cell chrome | none |
| Zero or negative rectangle | `width` or `height` ≤ 0, which the loader accepts | No box drawn; render otherwise unaffected. `elementDeclaresBox` is false, so the canvas places no window either | none |
| No style at all | neither field present | No source; byte-identical output | none |
| `style` present and null | `"style": null` | Treated as no style — a present-null is no style, as every other Presence reader treats it | none |
| Malformed colour | `background`/`border.color` not `#RRGGBB` | Located **render** error, `DiagCodeStyleColorInvalid`, naming the element and the field path. The format checks a colour string for a placeholder at load and for nothing else, by design | render fails; no bytes |
| Stroked edge subset | `border.edges: ["top","left"]` | Two one-segment subpaths, **operands before operator** (`x1 y1 m x2 y2 l`), closed by a single `S` | none |
| `border.edges: []` | present, non-null, empty | **AS SHIPPED: a box is declared and nothing is painted, yet the element still occupies a content-column item and a canvas window.** This is Finding 1 and it is a defect, not an intended arm | none — silently paints nothing |
| Negative `border.width` | `"width": -5` | **AS SHIPPED: `-5 w` is emitted into the content stream.** Finding 2 | none |

</intent-contract>

## Code Map

**Go — engine (`folio-go/`), at audit revision `6e06cc7`**
- `element_box.go:52-111` — `collectElementBoxRects(bands, visible) ([]tableRectSource, error)`.
  Walks bands in `documentBands` order, elements in declaration order. Skips TABLE by name, then
  `isVisible`, then `elementDeclaresBox`. Builds one `pagemodel.Rect` per eligible element through
  `buildCellRectWithBackgroundField` and wraps it in a `tableRectSource` whose `top`/`bottom` are
  `layout.PlaceInBand(b.origin, el.Y)` and `+ h`.
- `element_box.go:123-129` — `elementBoxDeclaration`: the single reading of "declares a box" —
  a present, non-null `style.background` **or** `style.border`.
- `element_box.go:165-172` — `elementDeclaresBox`: that verdict **AND** `declaredBox`'s positive
  rectangle. Second caller is `page_setup.go:909-911`.
- `element_box.go:187-195` — `declaredBox`: width and height both present, non-null and **> 0**.
- `element_box.go:207-231` — `elementInk`/`styleInk`. **Not this story's**: added later for Story
  10.1's text ink, in the same file.
- `table_render.go:525-575` — `buildCellRectWithBackgroundField`, the shared builder. Width default
  `geom.Length(500)`, colour default `#000000`, edge default all four. **The `edges` loop at
  `:555-566` ignores any value it does not recognise and an empty array yields all-false edges while
  `HasStroke` stays true** — see Finding 1.
- `render.go:2074-2091` — the `predictDocument` hook; `tableRects` is built element-boxes-first.
- `internal/pdf/rectdoc.go:56-79` — `appendRectContentStream`'s stroke arm, guarded by
  `r.HasStroke && (any edge)`; `:101-110` — `appendEdge`, now postfix.
- `internal/pdf/rectdoc_test.go` — `TestAppendRectContentStreamEdgeOperandsArePostfix` reads the
  **order**, not the subpath count.
- `page_setup.go:895-912` — `canvasElementIsPlaced`; `:988-1090` — `addCanvasWindowCount`, whose
  `exact` flag at `:1007-1009` has **four** causes and does not include the styled-text divergence
  its own doc comment at `:975-987` documents. See Finding 3.
- `internal/template/parse_bands.go:770-806` — `decodeBorder`. `edges` is checked against
  `closedBorderEdges`; **`width` is decoded by `decodePointsRaw` with no sign or range check**.
- `internal/template/decimal.go:30-68` — `decodePoints`, which accepts negatives by design.

## Tasks & Acceptance

*Reconstructed. These describe what the shipped code does, verified by reading it and by the probes
recorded under `## Verification`.*

**AC1 — Background fills the declared box.** A text, image, rect or line element declaring
`style.background` renders one filled `pagemodel.Rect` at its declared box, in every band, emitted
before that element's own text or picture. *Shipped; `element_box_test.go:74` and `:287`.*

**AC2 — Border strokes through the table-cell builder.** An element declaring `style.border` renders
one stroked `Rect` at the border's width and colour on the edges `border.edges` names, resolved by
`buildCellRectWithBackgroundField` — the same width default, colour default and edge set a table
cell resolves, never a second implementation. *Shipped; `element_box_test.go:93`, `:116`.*

**AC3 — Repeated bands repeat the box.** A box declared in the page header or page footer repeats on
every page. *Shipped; `element_box_test.go:210`.*

**AC4 — The content column carries the box.** A box in the content band travels as an ungrouped
content item and is clipped and shifted by the rules that already govern a table's chrome.
*Shipped; the `tableRectSource` literal at `element_box.go:99-110` leaves all three row flags false.*

**AC5 — A hidden element contributes no box.** *Shipped; `element_box_test.go:123`.*

**AC6 — A table's style is never painted twice.** *Shipped; `element_box.go:56-58` and
`element_box_test.go:248`.*

**AC7 — A rectangle with no area draws nothing.** *Shipped; `element_box_test.go:152`.*

**AC8 — A stroked edge's operands precede its operator.** *Shipped; `rectdoc.go:101-110`, pinned by
`TestAppendRectContentStreamEdgeOperandsArePostfix`, which reads order rather than counting subpaths.*

**AC9 — The corpus is byte-identical under the box paint.** *Shipped and verified: no golden digest
is attributable to the box paint. The four statement goldens that moved in `791ed00` moved by +4
bytes per page from `text_alignment.go`'s `Tm` x-operand (D-15.1.1), which this story does not own.*

> ⚠ **THIS CRITERION PASSES VACUOUSLY, in that word.** **No fixture in the 23-golden corpus declares a
> `style.background`, a `style.border` or a `style.color`** — that is DW-147, and it is a gating input
> to Story 15.3 with the owner named. The corpus is therefore byte-identical under the box paint
> because *its subject is absent from the corpus*, not because the box paint was measured to be
> byte-neutral over documents that use it. **An AC that passes because its subject is absent must never
> be filed as evidence that the subject is byte-identical.** What AC9 does establish is the narrower
> and still-useful claim it was written for: this story did not disturb documents that declare no box.

**AC10 — Fixed-point discipline.** No `float64`/`float32` reaches `internal/`. *Verified at the
audit revision: all 15 textual matches under `folio-go/internal/**/*.go` (non-test) are comments;
the only code sites are `internal/arch_test.go`'s own AD-23 scanner.*

## Review Triage Log

### 2026-09-02 — First review, at audit revision `6e06cc7`

**This code had never been reviewed.** Findings are ranked and evidenced below; **none was fixed** —
this pass had no mandate to modify source.

**Finding 1 — FAIL-OPEN (high). `style.border: {"edges": []}` declares a box, paints nothing, and
still occupies a page-model column item and a canvas window.**
`decodeBorder` accepts an empty `edges` array (the closed-set loop simply does not execute), so
`Border.Edges` is `{Set: true, Null: false, Value: []}`. `elementBoxDeclaration` therefore reports
`hasBorder = true`, `elementDeclaresBox` returns true, and `collectElementBoxRects` emits a
`tableRectSource`. `buildCellRectWithBackgroundField` sets `HasStroke = true` with all four
`RectEdges` false, and `appendRectContentStream`'s guard (`r.HasStroke && (r.Edges.Top || …)`)
correctly emits **no drawing operator at all**. The element is placed and prints nothing.

*This is exactly the defect class `elementDeclaresBox`'s own doc comment claims to have closed* —
"a caller that read only the style clause would place a column item the printed document has nothing
in". The predicate closed the *zero-rectangle* half of that hazard and left the *zero-ink* half open.

Evidence, from an out-of-tree probe against the module at `6e06cc7` (working directory
`<scratch>/probe`, `replace` onto `/Users/panitw/Projects/folio/folio-go`; nothing written into the
repo):

| document | PDF bytes | `/Count` | drawing ops |
|---|---|---|---|
| rect `y=600 h=200`, **no style** | 51657 | 1 | — |
| rect `y=600 h=200`, `border:{"edges":[],"color":"#000000"}` | 51894 | **2** | **0 fills, 0 strokes** |
| rect `y=600 h=200`, `border:{}` | 52017 | 2 | 1 stroke |

**A declaration that paints nothing adds a second, entirely blank page.** The 237-byte difference is
page-object scaffolding; the content stream carries no `re f` and no `S`.

**Finding 2 — UNVALIDATED INPUT REACHES THE PDF CONTENT STREAM (high). A negative
`style.border.width` is emitted verbatim as a line-width operand.**
`decodeBorder` decodes `width` through `decodePointsRaw` → `decodePoints`, which explicitly supports
a negative sign (`decimal.go:60-62`). Nothing between the loader and `appendRectContentStream`
checks the sign. Probe output at `6e06cc7`:

```
border default     w-ops=[0.5 w]   strokes=1
border width -5    w-ops=[-5 w]    strokes=1   diags=0
border width 0     w-ops=[0 w]     strokes=1
```

`-5 w` is not a valid content stream: ISO 32000-1 §8.4.3.2 requires the line-width operand to be
non-negative. No diagnostic is raised. **The bound already exists on the sibling path** —
`Canvas`/`CanvasWithTextPaint` refuses the same document with
`folio: component borderWidth exceeds the projection bound` (`canvasPropertyLength`). So the
designer cannot open a document the engine will happily render into an invalid PDF. `width: 0` is
legal PDF but means "thinnest line the device can render", which is a silent floor rather than an
invisible border; whether that was intended is not recoverable from the code.

**Finding 3 — FAIL-OPEN (high). The canvas reports `ContentWindowCountIsExact: true` for a count
this story made wrong.**
`collectElementBoxRects` accepts **text** among its four eligible kinds, so a content-band text
element declaring a background contributes a **full-declared-box** column item on the render path.
`addCanvasWindowCount` contributes only that element's shaped **lines**. The divergence is written
down honestly in `page_setup.go`'s own doc comment ("Measured: a styled text element at y 700 with a
declared height of 200 and a one-line value gives a canvas count of 1 against a real render of 2 …
That divergence is LEFT AS IT IS"), **but it is not one of the four causes that clear the `exact`
flag** (`column.FontChainDegraded`, a bound table, conditional visibility, and the degradation
branch). The canvas therefore asserts exactness it does not have.

Reproduced at `6e06cc7` — content band with a plain text element at `y=10` and a one-line text
element at `y`/`height` as given, `pageHeader`/`pageFooter` 20pt, A4, 36pt margins:

```
y=700 h=200 nobox   RENDER pages=1 | CANVAS windows=1 exact=true
y=700 h=200 +bg     RENDER pages=2 | CANVAS windows=1 exact=true   ← wrong, and flagged exact
y=650 h=100 +bg     RENDER pages=2 | CANVAS windows=1 exact=true
y=600 h=200 +bg     RENDER pages=2 | CANVAS windows=1 exact=true
```

The comment's own measured example is **true**, and the flag does not report it.

**Finding 4 — UNVALIDATED GEOMETRY (medium). An element box's rectangle is never bounded to the page
or the band.**
Unlike a table cell, whose coordinates are derived by the layout engine from column widths, an
element box takes the author's raw `x`, `y`, `width`, `height` straight into `pagemodel.Rect` and
then into the content stream. `declaredBox` checks only that width and height are **positive**.
FR44's clip-and-warn covers a text element's *widest line* against its declared width, not a box.
Probe at `6e06cc7`, all with `diags=0`:

```
bg, x=-500                -464 735.89 200 30 re f      painted off the left of the sheet
bg, w=100000 (off page)   48 735.89 100000 30 re f      painted ~1389 inches wide
bg, y=-9000               48 9755.89 200 30 re f        painted above the top of the sheet
```

Valid PDF (the renderer clips at the page boundary), but silent: the author is told nothing, and the
canvas — which bounds its own projected lengths — cannot show it.

**Finding 5 — CANVAS/PDF DIVERGENCE (medium), undisclosed. `edges: []` makes the canvas draw a
border the PDF does not print.**
`CanvasComponent.BorderEdges` carries `json:"borderEdges,omitempty"`, and `omitempty` drops an
**empty** slice. `App.tsx:1631` then reads `const edges = component.borderEdges ?? boxEdges` and
defaults to all four. Probe at `6e06cc7`, `border:{"edges":[],"color":"#ff0000"}` on a rect:

```
canvas = [{"id":"e1","type":"rect",…,"borderColor":"#ff0000"}]     ← no borderEdges key
pdf    = 0 fills, 0 strokes
```

`bordered` is true (because `borderColor` is defined), `edges` falls back to all four, and the
canvas paints a full red rectangle the printed document does not have. The neighbouring case —
`"border": {}` projecting nothing while the PDF still strokes a 0.5pt black box — **is** disclosed in
`App.tsx:1624-1626`; this one is not. Both violate AD-17 / Story 5.9's "the canvas shows what
prints"; only one is written down.

**Finding 6 — UNATTRIBUTED SCOPE (medium). `791ed00` also shipped the font-family projection.**
`CanvasProjection.FontFamilies`, `DefaultFontSize`, `canvasFontFamilies`, `maxCanvasFontFamilies`
and the matching `engine-protocol.ts` validator arm were introduced by this commit
(`git log -S "canvasFontFamilies" -- folio-go/page_setup.go` → `c1430b0`, `a67ab9b`, **`791ed00`**;
`git log -S "DefaultFontSize"` → **`791ed00`** alone). The commit message accounts for `page_setup.go`
solely as "callers of the in-flight text-alignment work". **No story, no acceptance criterion, no
FR.** It is a second instance of the pattern Story 15.1 already recorded for `text_alignment.go`.

**Finding 7 — GUARD RELAXED WITHOUT A STORY (low).** `engine-protocol.ts`'s `isTextPaint` fragment
bound moved from `fragment.x <= component.x + paint.width` to
`… + Math.max(paint.width, component.width)`. Widening a browser-side validation bound is
substantive; it belongs to the text-alignment work (a right- or centre-aligned fragment can sit
beyond `paint.width`) and therefore inherits that work's absence of a story.

**Finding 8 — CROSS-PACKAGE TEST RACE IN THE `lint` GATE (medium; Epic 8's code, not Epic 9's).**
`lint/internal/rules/fontsassets_test.go:283-299`'s `TestFontsAssetsNoticeRemovalRedProof`
**deletes the real committed file** `folio-go/fonts/notosans/NOTICE.md` from the working tree and
restores it in `t.Cleanup`. `lint/internal/manifest`'s `TestCommittedAssetPopulationClassifiesCleanly`
reads that same file, and `go test ./...` runs the two packages **concurrently**. Measured at
`6e06cc7`, clean tree, working directory `/Users/panitw/Projects/folio/lint`: `go test -count=1
./internal/manifest/` passed 4/4 in isolation; `go test -count=1 ./...` over the whole module failed
**2 of 5 runs** with
`folio-go/fonts/notosans: contains a committed font binary but no NOTICE* file naming its copyright
line (AC25, AD-26)`. Two consequences: the `lint` gate is nondeterministically red, and a test that
is interrupted mid-run leaves a committed file deleted in the user's tree.

**Finding 9 — COVERAGE GAP (low).** `element_box_test.go`'s 372 lines assert real properties — no
test in it is vacuous. But **no test covers any of Findings 1, 2, 4 or 5**: there is no case for
`edges: []`, none for a negative or zero `border.width`, none for a box whose rectangle leaves the
page, and none pairing the canvas projection against the PDF for a partial edge set.

**Comment claims checked against the code** (this run's standing hazard — three false comments have
been found in this repository already):
- `render.go:2074` "element boxes FIRST … Element boxes lead" — **true**; `tableRects` is
  `append`ed from `elementBoxes` before any table population.
- `element_box.go:85-86` "sized is necessarily true here" — **true**; `elementDeclaresBox` is the
  conjunction and `declaredBox` is pure.
- `rectdoc.go:88-100` "the byte-level tests count `m ` occurrences … rather than reading the operand
  order" and "no golden was ever opened … with a stroked edge in it" — **consistent** with Story
  15.1's operator census, which confirmed the path is never reached rather than assuming it.
- `element_box.go:131-133` "elementDeclaresBox is **THE WHOLE PLACEMENT RULE** for an element box" —
  **overstated**, and Finding 1 is the counterexample: it is the whole *declaration-and-rectangle*
  rule, and says nothing about whether the declaration has any ink.
- `page_setup.go:975-987`'s measured styled-text example — **true and reproduced** (Finding 3).

## Design Notes

**Why `tableRectSource` rather than a new carrier.** Everything downstream of `predictDocument` —
`contentColumnItems`' page-count pass, `paginateDocument`'s placement, the page assembler's
header/footer repetition — already reads "a rect group with a band and a vertical extent". Adding a
second carrier would have duplicated four behaviours to gain nothing. The cost is the one recorded in
Finding 1: the carrier says nothing about whether the group has ink, so a zero-ink group paginates
like any other.

**Why the corpus did not move.** An element declaring neither field produces no source, so
`paginateDocument` receives a byte-identical rect slice for every document in the corpus. The four
statement goldens that did move in this commit moved for a different reason entirely
(`text_alignment.go`; D-15.1.1).

**DW-24 is closed** (Story 7.3, 2026-08-30) with a re-derived enumeration, per-site falsifiers and a
registered `fixtures/alignment-rounding/` fixture, per D-000.9. Coverage of the rounded alignment
branches is **evidence available to cite**, not outstanding work. ⚠ `sprint-status.yaml`'s Epic 9
comment block still describes DW-24 as open ("zero golden coverage: DW-24"); that comment is stale
and was **not** corrected here, because this reconstruction's mandate is limited to spec files.

## What could not be determined

Stated plainly, because "could not look" is not "all clear":

1. **Whether `edges: []` was ever considered.** Nothing in the code, the tests, the commit message or
   the epic text mentions the empty-array arm. It reads as unconsidered rather than as a decision,
   but the record cannot distinguish the two.
2. **Whether `border.width: 0` was intended as "hairline" or as "no border".** The code produces the
   former (`0 w`). No comment, test or epic sentence takes a position.
3. **Whether text was deliberately included among the four eligible kinds** knowing it would diverge
   the canvas window count. `page_setup.go`'s comment documents the divergence and says it is "LEFT
   AS IT IS … out of this arm's subject", but that comment was written during Story 7.9/7.10, *after*
   this commit. What the 9.1 author knew at the time cannot be recovered.
4. **Why `text_alignment.go` and the font-family projection were carried in this commit.** The commit
   message explains the first ("those callers do not compile without it") and is silent on the second.
5. **Whether the `isTextPaint` bound relaxation (Finding 7) was reviewed by anyone.** No review
   artefact exists for this commit at all.
6. **What the original acceptance criteria were.** The epic text was written in the same session as
   the implementation; there is no earlier draft in the repository.

## Verification

**Run at audit revision `6e06cc7` (`main`, tree clean), 2026-09-02.** Every invocation's working
directory is recorded, per D-8.4j.8. **Per-epic cadence: the four `FOLIO_MATRIX_TARGET` legs,
`TestCrossTargetByteIdentity` and Playwright are excluded** — those are Epic 9/10's boundary gate,
after Epic 10.

**Standing reds, by identity — three were expected; a fourth was found and is Finding 8.**

| # | Command | Working directory | Result |
|---|---|---|---|
| 1 | `go test -count=1 ./...` | `/Users/panitw/Projects/folio/folio-go` | **1877 pass / 2 fail / 5 skip** — `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` only (`got 7, need >=20`). Matches the documented permanent red exactly. |
| 2 | `go vet ./...` | `/Users/panitw/Projects/folio/folio-go` | no output, exit 0 |
| 3 | `go test -count=1 ./...` | `/Users/panitw/Projects/folio/lint` | `cmd/genmanifest` ok · `internal/licence` ok · `internal/rules` ok · `internal/manifest` **FAIL 2 of 5 runs** — Finding 8, a cross-package race, not a code regression |
| 4 | `gofmt -l folio-go lint` | **`/Users/panitw/Projects/folio` (repo root)** | exactly `lint/internal/rules/licencegraph_test.go` — standing red 2. ⚠ From `lint/` this prints `lstat folio-go: no such file or directory` and reads as clean; it was re-run from the root. |
| 5 | `go run ./cmd/genmanifest` | `/Users/panitw/Projects/folio/lint` | `wrote …/lint/MANIFEST.md` |
| 6 | `git ls-files --error-unmatch lint/MANIFEST.md && git diff --exit-code --stat -- lint/MANIFEST.md` | **`/Users/panitw/Projects/folio` (repo root)** | pathspec resolves; **empty diff, exit 0** — the committed manifest still matches. ⚠ `git diff <a> <b> -- <path>` returns empty patch text in this environment, so `--stat` was used and the pathspec proved separately. |
| 7 | `npm run typecheck` | `/Users/panitw/Projects/folio/folio-designer` | exit 0, no diagnostics |
| 8 | `npm run lint` | `/Users/panitw/Projects/folio/folio-designer` | **exactly 4** `react(only-export-components)` warnings — `pdf-viewer.tsx:16,17`, `App.tsx:1403,1410`. Standing red 3. |
| 9 | `npm test` | `/Users/panitw/Projects/folio/folio-designer` | **42 files passed / 432 tests passed** |
| 10 | `shasum -a 256 fixtures/*/expected.pdf` | **`/Users/panitw/Projects/folio` (repo root)** | **23 digests, all holding** (see below) |

**The 23 goldens hold** at `6e06cc7`:

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

**AD-23 check.** `grep -rnE '\bfloat(64|32)\b' internal/ --include='*.go'` from
`/Users/panitw/Projects/folio/folio-go` returns 15 lines under non-test files; **all 15 are prose
inside comments**. The only code references are in `internal/arch_test.go`, which is AD-23's own
scanner. **No `float64` or `float32` under `internal/`.**

**Tree state.** `git status --porcelain` from the repo root was empty before this spec was written,
and shows only the two new spec files afterwards. **No source file was modified, nothing was
committed, nothing was pushed, no branch was created.**

## Delivery Log

### 2026-08-30 — shipped, unrecorded
Implemented and committed as part of `791ed00` with no story file, no acceptance criteria, no review
and no retrospective. Left at `review` in `sprint-status.yaml`.

### 2026-09-02 — reconstructed and reviewed
Spec written retroactively from the shipped diff at audit revision `6e06cc7`. First review of the
code: **9 findings, none fixed** (this pass had no mandate to modify source). Findings 1, 2 and 3
are fail-open or unvalidated-input defects with reproduced evidence and should be triaged before
Story 15.3 cuts a tag over this code. Gates re-run and recorded above; the 23 goldens hold.
