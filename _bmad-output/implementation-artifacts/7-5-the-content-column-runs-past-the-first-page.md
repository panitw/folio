---
title: 'Story 7.5: The content column runs past the first page'
type: 'feature'
created: '2026-08-31'
status: 'blocked'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized'] # oversized: the refusal-split surface, the four-mirror surface and the window-count evidence are three wide surfaces that must be stated, not summarised. NOT multiple-goals: the projection field is AC4 of this same story, not a second shippable.
deferred: []
---

## In plain terms

*Non-normative. This section explains the story in plain language; the contract below governs.*

Today a template author can only place things on the first page. The designer refuses any component dropped below the foot of page one, because the single check that keeps a component inside its band treats one page's worth of content as the whole of the available space. That refusal is what makes a schedule or a signature block impossible to author at all.

This story removes that ceiling for the main content area, and only there. Repeating page furniture — the strip at the top of every page and the strip at the bottom — stays exactly one page tall, because that is what repeating means. Coordinates that are simply nonsense, negative ones or ones too large for a browser to hold safely, are still refused, in the same words as today.

The story also has the engine report two numbers the designer will need: how tall one page's worth of content is, and how many of those page-fulls the current column runs to. The designer must be told these, never work them out itself.

This story makes a tall column legal and reportable. It does not draw it — that is the next story. It changes nothing about how the engine decides where pages break. One test is expected to stay red throughout: the corpus exercise floor recorded as P6g, a standing, deliberate failure that must never be "fixed".

<intent-contract>

## Intent

**Problem:** A template author cannot place a component below the foot of page one, so a schedule and a signature block cannot be authored at all. The engine already renders fifty-page documents from content elements placed windows apart — `fixtures/page-count-50` proves it — but every mutating designer command refuses such a placement, because one shared containment check treats the content band's one-page height as the column's limit. The canvas projection also reports nothing about page windows, so the designer has no engine-supplied basis for drawing them.

**Approach:** Lift the content band's vertical limit in the designer command layer only, preserving the horizontal limit, the two repeating bands' limits and every existing refusal message verbatim; lift the browser-side mirror of the same bound in the same commit; and add a projection field reporting the page-height window and how many windows the content column occupies. No engine layout work: `internal/layout/paginate.go` is an input to this epic, not a target.

## Boundaries & Constraints

**Always:**
- The vertical lift applies to the **content** band alone, keyed on the band's identity.
- Every surviving refusal keeps its **exact message text**, and this story adds assertions on that text — today no test asserts any of these strings, so they are unprotected.
- Whichever bound changes on the Go side changes its TypeScript mirror **in the same commit**, with an executable assertion tying the two (D-7.4.5, as corrected to four mirrors).
- The reported window count is derived from the document's laid-out geometry, **never** from `CanvasTextPaint`; the independence asserted by `TestPaginationIsIndependentOfCanvasPaintTruncation` and by `canvas-authority-contract.test.ts`'s two prohibited regexes must survive and be re-asserted from this side (D-7.4.2 §5).
- Corpus neutrality: every existing template's rendered bytes and every existing projection field value are unchanged. Assert it; do not assume it.
- The top-left → PDF-bottom-left flip lives in exactly one function (AD-24). A window offset must not become a second one.

**Block If:**
- **The window count's derivation is not settled by the intent.** See Design Notes, "The window-count gap". This story is blocked on it.
- A `.folio` format field turns out to be required. It is not expected to be — this lifts a validation bound and adds a projection field — but if one is, that is a version question (D-1.4.9 / D-1.4.12 / D-1.4.13) to **flag, not decide**. `SupportedMajor` is 2 and does not move here.
- The work touches DW-33's path (a text element whose first line already exceeds the per-line guard paints zero lines). Flag it for a ruling; do not decide whether a partial prefix should paint.

**Never:**
- Never modify `internal/layout/paginate.go`. Never reintroduce the closed form `ceil(lowestBottom / H)`, which that file forbids by name at `:69-74`.
- Never widen the `pageHeader` or `pageFooter` band bounds.
- Never draw multiple sheets (Story 7.6), implement keep-together (7.7), or implement DW-29's load rejection (7.8).
- Never close DW-26, DW-27, DW-28, DW-30, DW-31, DW-32, DW-33, DW-34, DW-35.
- Never modify, move, delete or commit the repository-root `README.md`. Never `git add -A` / `git add .`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Content Y past one window | `setComponentBounds` / `moveComponent` / `createComponent` in `content` with `y` far beyond the band height | **Accepted**; element persists at that Y; canonical bytes carry it | No error expected |
| Content box bottom past one window | content element with `y + height > band.Height` | **Accepted** | No error expected |
| Content X past band width | content element with `x > band.Width` | **Refused**, message unchanged | `folio: component geometry must stay within content` |
| pageHeader / pageFooter Y past band | same command in either repeating band | **Refused exactly as today** | `folio: component geometry must stay within pageHeader` / `... pageFooter` |
| Negative coordinate, any band | `x`, `y`, `width` or `height` < 0 | **Refused**, message unchanged | `folio: component geometry must stay within <band>` |
| Beyond JS-safe geometry bound | a coordinate magnitude > `MaxCanvasMillipoints` | **Refused**, message unchanged, separate path | `folio: component.y: y exceeds the JavaScript-safe geometry bound` |
| Projection of a tall column | template whose content column spans N windows | Projection reports the window height and N | No error expected |
| Projection of every shipped template | the existing corpus, none placing past page one | Every pre-existing field value identical; window count reports one window | No error expected |
| Truncated paint, tall column | element long enough to truncate its paint | Reported window count **identical** to the untruncated document | No error expected |
| Browser admission of a tall column | projection carrying a content component past the band height | Snapshot **admitted**, not silently dropped | No error expected |

</intent-contract>

## Code Map

Every anchor verified in the tree at `23a4647`.

### The single refusal — `folio-go/component_commands.go`

- **`:1766-1771` `containComponent(band CanvasBand, x, y, width, height geom.Length) error`** — the **only** band-extent validation in the designer command path. **One `if`, one message:**
  ```go
  if x < 0 || y < 0 || width < 0 || height < 0 || x > geom.Length(band.Width) || y > geom.Length(band.Height) || width > geom.Length(band.Width)-x || height > geom.Length(band.Height)-y {
      return fmt.Errorf("folio: component geometry must stay within %s", band.Name)
  }
  ```
- ⚠ **THE LOAD-BEARING FACT.** The content-band one-page cap (the clauses `y > band.Height` and `height > band.Height-y`), the `pageHeader`/`pageFooter` caps, and the **negative-coordinate** refusal are **the same expression producing the same format string**. Only `%s` differs, filled from `band.Name`. There is no existing seam. "Lift the content Y cap" and "keep the negative refusal with its message unchanged" are contradictory demands on one string until the expression is split.
- **Twelve call sites**, all reaching `:1767`: `addTableColumn` `:131`→`:165`; `updateTableColumn` `:245`→`:300`; `updateComponentPropertiesInPlace` `:788`→`:829` (the property path for `x`/`y`/`width`/`height`); `dropComponent` `:1362`→`:1409` (probe, discarded); `createComponentInBand` `:1417`→`:1472`; `moveComponent` `:1531`→`:1565` (probe), `:1569` (refusal); `resizeComponent` `:1576`→`:1606`; `setComponentBounds` `:1619`→`:1668` (probe), `:1674` (refusal); `duplicateComponent` `:1699`→`:1726` (probe; on failure falls back to the original origin at `:1727`, **no refusal**).
- **Pull-back arithmetic that changes meaning when the cap lifts** — `containEdge` `:1752-1757`, `floorToGrid` `:1759-1764`, and their callers `dropComponent:1410-1411`, `moveComponent:1566-1567`, `setComponentBounds:1669-1672`. Each clamps against `band.Height`.
- **Band identity is a bare `string`**, not an enum: `CanvasBand.Name` (`page_setup.go:157`). The three literals exist only inline at `page_setup.go:354/355/356`; re-spelled in `bandByName` `:1294-1301`, `hitTestBand` `:1504-1511`, `canvasComponents` `page_setup.go:720-727`. **There is no named constant to key off.**
- Band values reach `containComponent` through `bandByName` `:1285-1304` (calls `Canvas(t)` at `:1286`) and `hitTestBand` `:1494-1514` (`Canvas(t)` at `:1495`).
- `applyPropertyChanges` `:894` sets `X/Y/Width/Height` at `:962-969` and enforces only positivity `:954-956` — **no band bound of its own**.

### The separate, surviving refusal — the JS-safe geometry bound

- `MaxCanvasMillipoints int64 = 9007199254740991` — `page_setup.go:25`.
- `lengthField` `page_setup.go:1043-1067`, enforced `:1063-1064`: `"%s exceeds the JavaScript-safe geometry bound"`. Wrapped by `componentLength` `component_commands.go:1258-1267` (`"folio: component.%s: %w"` at `:1261`) and by `propertyLength` `:884-886` → `:952`.
- Projection-time sibling `page_setup.go:730-734`: `"folio: component exceeds the JavaScript-safe geometry bound"`. **Retain verbatim.**
- ✅ **Separate path, separate message, upstream of `containComponent` — untouched by any split.**

### The band rectangle — `folio-go/page_setup.go`

- **`:341`** `innerW, innerH := w-m.Left-m.Right, h-m.Top-m.Bottom`; **`:353-357`** the three bands, open-coded. Content is `Height: int64(innerH - header - footer)` at **`:355`** — **this is the one-page cap's source.**
- ⚠ **`page_setup.go` does NOT import `internal/layout`** (imports `:3-15`). `layout.ContentHeight` (`internal/layout/band.go:75-77`) is a **second, independent spelling** used only by the render path. The designer-side cap is `page_setup.go:355`, not `band.go:75`. **`paginate.go` is genuinely uninvolved in the refusal.**
- Frame difference to state in any new field's comment: `layout.Origins` measures downward from the printable top edge (`PageHeader: 0`); `CanvasBand.Y` is **page-absolute**.
- A4 content-band height **729890 mp** exists only as a comment at `:74`/`:76`; as a literal in `App.test.tsx:17` and `DataPanel.test.tsx:18`. (`727890` at `multi_page_fixture_test.go:67` is a different geometry.)

### The projection structs — `folio-go/page_setup.go`

- **`CanvasProjection` `:265-293`** — `Width`/`Height` `:266-267`, `Orientation`, `Preset`, four margins `:270-273`, `GridIncrement` `:274`, `CommandWidth`/`CommandHeight` `:275-276`, `Bands` `:277`, `Components` `:278`, `FontFamilies` `:286`, `DefaultFontSize` `:292`. lowerCamelCase json tags, **no `omitempty` on any top-level field**. `FontFamilies`/`DefaultFontSize` (`:279-292`) are the newest additions and carry long comments justifying *why the engine, not the browser, owns the number* — **that is the tone a new field must follow.**
- `CanvasBand` `:156-162`; `CanvasComponent` `:163-215`; `CanvasTextPaint` `:134-147` (`Overflow` `:135`, **`Truncated` `:145`**, `Lines` `:146`); `CanvasTextLine` `:124-130`; `CanvasTextFragment` `:116-119`.
- **There is no page-window or window-count field anywhere.** `bands[1].height` is already exactly the page-height window, stated for one page.
- `Canvas` `:323-367` calls only `canvasDimensions` `:904`, the inline band arithmetic, `canvasComponents` `:716`, `canvasFontFamilies` `:302`. `CanvasWithTextPaint` `:372-384` adds `addCanvasTextPaint` `:503` and `addCanvasImagePaint` `:407`. **Neither reaches `Paginate`, `ContentHeight`, `Origins`, `pageGeometryOf`, `documentBands` or `contentColumnItems`.**
- `canvasComponents` `:716-777` does **not** re-check containment — only `> MaxCanvasMillipoints` at `:730-734`. **`Canvas()` already projects out-of-band content elements happily; no Go projection change is needed for admission.**

### The window count — evidence gathered, derivation UNSETTLED

- ⚠ **`internal/layout/paginate.go:69-74` forbids the closed form by name:**
  > "PAGE COUNT IS NOT A CLOSED FORM. It falls out of the advance. In particular it is NOT `ceil(lowestBottom / H)`, and that spelling must not be reintroduced: the window advances to the first UNPLACED item, not by a fixed H, so an element declared far below the text STARTS THE NEXT WINDOW rather than generating blank pages before it."
- Measured consequence: `len(Paginate(...).Pages)` diverges from `ceil(max(Top+Height)/H)` **in both directions**. Larger — slide waste compounds (four 60-tall items at tops 0/60/120/180 with H=100 give **4** pages, closed form says 3). Smaller — a lone item at Top=1000 with H=100 gives **1** page (`pageHasItem` guard, `paginate.go:995`), closed form says 11.
- `Paginate` `:548`; the sweep `:740-1099`; the window slide `:983-1002` (`windowStart = effectiveTop`, `pages[page].Shift = windowStart - contentTop`). `PageAssignment.Shift` `:474-500` is **unconditionally zero on page 0**. `RowDisplacement` `:520-525` and `Pagination.Clipped` `:341-357` are **unreachable without tables/groups**. An **ungrouped** item taller than one window is a hard `*OverflowError` at `:1075-1089`, not a page of its own.
- `contentColumnItems` `page_number.go:87-155`; produced in the render path from `collectBandTextRuns` (`render.go:686`, called `:1656`), `collectImageRuns` (`:437`), `collectElementBoxRects`/`collectBandTableRuns` (`table_render.go:587`), `computeVisibility` (`render_visibility.go:86`); consumed `render.go:1670-1673`.
- ⚠ **The render machinery cannot be reused with empty data.** `collectBandTextRuns` binds at `render.go:727` via `bind.BindTextSpans` → `bind.Resolve` (`internal/bind/text.go:206-283`). An **absent** path is a hard error (`internal/bind/text.go:398-401`), returned as `*RenderError{DiagCodeBindingPathAbsent}` at `render.go:728-735` — **not** a diagnostic, and it aborts the whole collection. Only an explicit JSON `null` renders empty (`text.go:266-270`). **Every shipped fixture with a content-band placeholder errors under `{}`**: `testdata/example/first-pdf.folio:6`, `fixtures/statement-{1,5,20,50}/input.folio` (`e5`/`e6`/`e7`), `fixtures/wrapped-text/input.folio` (`e4`), `fixtures/mandatory-break/input.folio` (`e3`). A value binding to empty is also **dropped from the page model entirely** (`render.go:775-782`).
- The canvas shapes the **raw authored string** (`page_setup.go:544`, `:548-552`) with `nil` substitutions (D-7.4.4), so canvas and render measure **different characters** whenever a placeholder is present.
- `paginate.go:98-104` warns against re-deriving item extents: "Re-deriving it here would be a second derivation of the same number, which is exactly what D-2.4.2's amendment exists to prevent."
- ✅ Same package (`package folio` in both `render.go:1` and `page_setup.go:1`); `internal/layout` imports only `fmt`, `slices`, `strconv`, `internal/geom` — **no import cycle** would block either option. `pageGeometryOf` `render.go:291`; `documentBands` `render.go:316`. `isVisible` (`render_visibility.go:39-45`) returns `true` for a `nil` map, so `contentColumnItems(runs, nil, nil, nil)` is safe.

### Paint independence, landed by Story 7.4 — must survive and be re-asserted

- `CanvasTextPaint.Truncated` `page_setup.go:145`, set at `:570` and `:659`.
- `TestPaginationIsIndependentOfCanvasPaintTruncation` — `folio-go/canvas_body_text_bounds_test.go:237-278`; its framing comment `:230-236` names Story 7.5 explicitly. Obtains its count via `documentBands` → `collectBandTextRuns(..., data, data, ...)` → `contentColumnItems(runs, nil, nil, nil)` → `layout.Paginate(mustPageGeometry(t, tpl), ...)` at `:243-256`; `data` is `emptyBindValue(t)` at `:239` (test-only helper, `shaped_fixture_test.go:1446-1453`); `mustPageGeometry` is test-only (`collect_text_runs_composition_test.go:102-109`). Presence precondition `:265-267`; equality `:269-271`; non-coincidence `:275-277`.
- Prohibited regexes — `folio-designer/src/canvas-authority-contract.test.ts`, rationale `:29-37` (names Story 7.6), entries **`:38`** `/\b(?:textPaint|paint)\??\.lines\.length\b/` and **`:39`** `/\blines\.length\s*[*/]|[*/]\s*\blines\.length\b/`. Non-vacuity `:51-56`; red-proofs `:70-71`.
- **Nothing reads `paint.lines.length` today.** Go: only tests. TypeScript: one production read, `engine-protocol.ts:291`, a bound check (`>`), not a derivation — it survives both regexes by construction.

### The browser mirrors — `folio-designer/src/`

- ⚠ **`engine-protocol.ts:193`** independently re-enforces band containment on every projection:
  ```ts
  if (!(box.x + box.width <= band.width && box.y + box.height <= band.height)) return false
  ```
  `parseInbound` returning `undefined` **discards the whole snapshot** — no canvas, no error, silent blank. **This must lift for the content band in the same commit, or Story 7.5 is invisible in the app.**
- Companion band invariants to leave alone: `:171` (`paint.y + paint.height <= page.height`), `:174` (bands contiguous).
- Exact-key `hasOnly` `:139`. Allow-lists: snapshot `:311`; **canvas/page `:156`** (the list a new top-level projection field must join); band `:169`; component `:184`; textPaint `:291`; line `:300`; fragment `:307`; image `:248`; envelope `:333`. Type decls `CanvasProjection` `:95-104`, `EngineSnapshot` `:84-91`. Value predicates alongside `:157-158` (`integer(key, positive)`).
- Test `engine-protocol.test.ts:24-31` — the out-of-band case at `:26` is an **X** overflow (`x: 991` in a 1000-wide band), so it survives a Y-only lift; fixture content band `height: 1800` at `:4`.
- `engine-bounds-mirror.test.ts` — `goSources` `:33-36`, `tsPath` `:37`, `pairs` `:47-54` (**six** pairs, `toHaveLength(6)` at `:81`), extractors `:59-65`, red-proof `:110-128`. It ties **numbers only** — not struct field names, not `hasOnly` key lists.
- **Third gate, and it is Story 7.6's, not this story's:** `resize-anchor.ts:29` `proposedBounds(..., limit?)` clamps live drags to `DragLimit` `:25`, filled from `{ width: band.width, height: band.height }` at `App.tsx:701` and applied at `App.tsx:1176`. Pinned by `resize-anchor.test.ts:43-53`. Dragging onto a later sheet is Story 7.6's own AC.
- Fixtures that must gain any new field: `App.test.tsx:17`, `DataPanel.test.tsx:18`, and `engine-protocol.test.ts`'s fixtures.

### The designer surface — `folio-designer/src/App.tsx`

- Single sheet `:700` styled by `pageStyle(canvas, zoom)` **`:1165`**; bands `:701` via `bandStyle` `:1166`; `canvasDisplay` `:1155-1158`; `placementPoint` **exported, `:1162-1164`**.
- ⚠ **Positional fragility.** `canvas-authority-contract.test.ts:86-88` holds a source-text seam requiring `export function placementPoint(event: Pick<MouseEvent,` to be followed **immediately** by `\n}\nfunction pageStyle`. It is the sole approved carve-out for the `offsetX`/`offsetY` ban (`:22`). **Any new helper must go after `pageStyle`, never between the two** — inserting there is a double red.

### Tests pinning today's behaviour

- ⚠ **The message text is asserted NOWHERE.** `grep -rn "must stay within" --include="*_test.go"` → **zero hits**; `"component.geometry"` in tests → **zero hits**. Every existing test asserts only `err == nil` / `err != nil` plus canonical-byte immutability. **AC3's "message unchanged" is currently unprotected, and this story must add the assertions.**
- **`TestSnapDoesNotPushAnEdgeDragOutOfItsBand` — `component_commands_test.go:583` — the one test that must be inverted.** Content band at `:586`; `edgeY` from `band.Height` at `:603`; Y clamp assertions at **`:612-614`** and **`:624`**; the refusal case **`"a grid step past the bottom edge"` at `:632`** is exactly category (a). Its X cases `:631`/`:633` and the grid-step-distance assertion `:615-617` must survive.
- Survive unchanged: `TestComponentCommandsSnapContainAndFailureAreTransactional` `:73` (X case `:101-103`); `TestSetComponentBoundsMovesOriginAndSizeInOneCommand` `:548` (negative-X `:570`, band-width `:571`) — **the natural home for a new "tall content Y is accepted" case**; `TestTableColumnCommandsAreTransactionalAtThePublicSeam` `:350` (width-driven); `TestDropComponentUsesGoHalfOpenBandHitTesting` `:514` (`hitTestBand`, `:1513`).
- **Coverage gap:** `updateComponentProperties` with a large content-band `y` (the path through `:829`) has **no test today**, on either side of the change. `component_properties_test.go` has no band-bounds coverage at all.
- Fixture `componentTemplate` `component_commands_test.go:13-24` → `testdata/template/golden/worked-example.json`: A4, 36pt margins, header 60, footer 30 ⇒ content band **523276 × 679890 mp**. Use these numbers.

### Existing multi-page evidence — the engine needs no work

- **`fixtures/page-count-{1,5,20,50}`** — templates `page_count_matrix_templates.go` (`:141` for 50); the shape comment `:13-19` states N single-line content elements at `y = i*728pt`, one content-band window apart, "so EVERY element lands in its own pagination window". `page-count-50` reaches y ≈ 35672pt, ~49 windows down, and renders to exactly 50 pages (`page_count_matrix_test.go:87-95`). **Hand-authored JSON, never produced through a designer command.**
- **No load-time or render-time refusal of a far-below content element exists.** `parse_bands.go:147-151` → `decodePointsRaw` (`parse.go:428-438`) checks representability only; no band-height comparison anywhere in `internal/template`. The only positional error in `Paginate` is `itemHeight > height` (`:1075`) — the item's **own height**, never its position.
- ⚠ **No canvas/projection fixture is multi-page at all.** `multi_page_fixture_test.go`, `statement_fixture_test.go` and `page_count_matrix_test.go` never call `Canvas`. Story 7.5 adds the first one.

### Verification surface

- `folio-go/byte_neutrality_test.go:92-460` `goldenDigestRecord` — 20 entries; "invalidated IN WHOLE" clause `:225`.
- `matrix_test.go:69-74` `matrixTargets` (4); `TestTargetRenderHash` `:1979`, gate `:1980-1990` (**logs "asserts NOTHING" and returns when `FOLIO_MATRIX_TARGET` is unset**); `TestCrossTargetByteIdentity` `:1802`.
- `internal/text/corpus_test.go:169` `TestCorpusMeetsP6ExerciseFloors`; P6g floor `:185`; drift twin `TestCorpusP6StatsMatchDeclaredBaseline` `:243` **must stay green**.
- `fontgen_matrix_test.go:64` `TestShippedFacesReproduceFromUpstream` — environmental (no `fontTools`). `lint/internal/rules/licencegraph_test.go:112` — the gofmt-dirty file (DW-23).

## Tasks & Acceptance

⚠ **This section is INCOMPLETE BY DESIGN.** The window-count tasks cannot be written until the gap in Design Notes is ruled. Everything below the divider is settled by the ACs and the evidence; everything above the AC list that depends on the count is marked `[BLOCKED]`.

**Execution — Part 1: the refusal split (settled).**

- `folio-go/component_commands.go` — **Split `containComponent`'s single conditional** so the content band's vertical clauses (`y > band.Height`, `height > band.Height-y`) no longer apply, while `x`/`width` clauses, the negative clauses, and both repeating bands keep every clause they have today. Key on the band's identity; introduce named constants for the three band names rather than adding a fourth inline spelling of the literals. **The format string `"folio: component geometry must stay within %s"` must be preserved byte-for-byte for every surviving clause.** -- Rationale: the content cap, the repeating-band caps and the negative refusal are one expression and one message today, so lifting one without a split silently changes the other two.
- `folio-go/component_commands.go` — **Leave `containEdge`/`floorToGrid` (`:1752-1764`) and their five callers arithmetically correct under the lift**: a content-band pull-back must no longer clamp Y against one window while still clamping X against the band width. -- Rationale: `moveComponent:1566` and `setComponentBounds:1669` pre-clamp before validating; a lifted validator behind an unlifted clamp accepts nothing new.
- `folio-go/component_commands.go` — **Do not add any new refusal for a content element taller than one window.** The render path already returns `*OverflowError` (`paginate.go:1075-1089`) for an ungrouped over-tall item; that is the authority. -- Rationale: D-7.4.2's rejected option (c) — a canvas/command bound must never become a document validity rule, and the canvas-approximate/preview-exact asymmetry is the product concept.

**Execution — Part 2: the browser mirror (settled).**

- `folio-designer/src/engine-protocol.ts` — **Lift `:193`'s vertical containment for the content band in this same commit**, leaving the horizontal check and the band-level invariants `:171`/`:174` intact. -- Rationale: `parseInbound` discards the entire snapshot with no attributable error, so a Go-only lift makes the story invisible in the app; and D-7.4.5 requires the mirror to move in the same commit.
- `folio-designer/src/engine-protocol.test.ts` — **Add a case admitting a content component whose `y + height` exceeds the band height, and keep the X-overflow rejection at `:26`.** -- Rationale: the existing out-of-band case is an X overflow and would pass vacuously after a Y-only lift.

**Execution — Part 3: the refusal-message assertions (settled, and currently missing).**

- `folio-go/component_commands_test.go` — **Invert `TestSnapDoesNotPushAnEdgeDragOutOfItsBand` (`:583`) on the Y axis only**: `:612-614` and `:624` must no longer require `Y+Height <= band.Height` for content, and the `"a grid step past the bottom edge"` case at `:632` must become an accepted case. Keep both X cases and the grid-step-distance assertion `:615-617`. -- Rationale: this is the only shipped test that pins the content-band Y cap; leaving it would either fail or be weakened into vacuity.
- `folio-go/component_commands_test.go` — **Assert the surviving refusals BY MESSAGE TEXT**, not merely that an error occurred: negative `x`/`y`/`width`/`height` in each of the three bands; a content `x` past the band width; a `pageHeader` and a `pageFooter` `y` past their band heights; and the JS-safe bound's own distinct message from `componentLength`. -- Rationale: no test asserts any of these strings today, so AC3's "message unchanged" is unprotected and the split could silently reword them.
- `folio-go/component_properties_test.go` — **Cover `updateComponentProperties` with a large content-band `y`** (the path through `component_commands.go:829`) on both sides of the change. -- Rationale: that entry point has no band-bounds coverage at all today.

**Execution — Part 4: the projection field. `[BLOCKED — see Design Notes]`**

- `folio-go/page_setup.go` — Add the window-height and window-count projection fields, following `FontFamilies`/`DefaultFontSize` (`:279-292`) in stating **why the engine, not the browser, owns the number**, and stating which coordinate frame the value is in. **The derivation of the count is the blocking question.**
- `folio-designer/src/engine-protocol.ts` — Add the new keys to the `hasOnly` list at `:156`, the `CanvasProjection` type at `:95-104`, and a value predicate at `:157-158`; update the fixtures at `App.test.tsx:17`, `DataPanel.test.tsx:18` and in `engine-protocol.test.ts`. -- Rationale: `hasOnly` is exact-key, so a Go-only field addition silently drops the entire snapshot.
- `folio-go/` + `folio-designer/src/canvas-authority-contract.test.ts` — Re-assert paint independence **from this side**: the reported window count must be identical for a document whose paint truncates and the same document untruncated, and the two prohibited regexes at `:38-39` must keep their non-vacuity and red-proofs.
- A new **multi-page canvas fixture** — there is none today.

**Acceptance Criteria:**

- Given a placement or bounds command in the **content** band with a Y beyond one page's content height, when the engine validates it, then it is **accepted** and the element persists at that Y in the canonical bytes.
- Given the same command in the **pageHeader** or **pageFooter** band, when it is validated, then it is refused, and the test asserts the **exact message text** `folio: component geometry must stay within pageHeader` / `... pageFooter`.
- Given a negative coordinate in any band, or a coordinate beyond the JavaScript-safe geometry bound, when it is validated, then it is still refused and the test asserts the **exact message text**, unchanged from `23a4647`.
- Given a content-band component whose box extends past one window, when the projection reaches the browser, then the snapshot is **admitted**, not silently discarded.
- Given a template whose content column is longer than one page, when the canvas is projected, then the projection reports the page-height window and how many windows the column occupies. `[BLOCKED]`
- Given the same document projected with and without paint truncation, when the reported window count is compared, then it is identical; and given the designer sources, when the authority contract scans them, then no height or window count is derived from a paint's line count. `[BLOCKED]`
- Given every existing template, none of which places anything past page one, when it is projected and rendered, then every pre-existing projection field value and all twenty golden digests are **measured** unchanged.
- Given the working tree at the end of the story, when it is inspected, then the repository-root `README.md` is byte-identical to its committed state and appears in no commit.

## Spec Change Log

## Review Triage Log

## Design Notes

### The window-count gap — why this story is blocked

**AC4 requires the projection to report "how many windows the column currently occupies". No derivation available to this story satisfies the constraints the architecture already imposes, and nothing in the intent selects between the survivors.**

Three candidate derivations, and what each costs:

**(C) The closed form, `ceil(max(y+height) / windowHeight)` — FORECLOSED, not a choice.** `internal/layout/paginate.go:69-74` forbids this spelling by name. It is wrong in both directions, measurably: slide waste compounds (four 60-tall items at tops 0/60/120/180 with H=100 paginate to **4** pages, closed form says 3), and the `pageHasItem` guard at `:995` collapses a lone far-below element to **1** page where the closed form says 11. This is recorded so the next reader does not re-propose it.

**(A) Reuse the render machinery** — `documentBands` → `collectBandTextRuns` → `contentColumnItems` → `layout.Paginate`, the exact sequence Story 7.4's own `TestPaginationIsIndependentOfCanvasPaintTruncation` uses as its oracle. **It cannot be fed empty data.** `collectBandTextRuns` binds at `render.go:727`; an absent path is a hard `*RenderError{DiagCodeBindingPathAbsent}` (`render.go:728-735`) that aborts the whole collection, and **every shipped fixture with a content-band placeholder errors under `{}`** — `first-pdf.folio`, all four `statement-*`, `wrapped-text`, `mandatory-break`. Making (A) work means threading the designer's **actual** data into `Canvas`/`CanvasWithTextPaint` — a public signature change, a new coupling from the canvas projection to `internal/bind` and the render pipeline, and a count measured over *substituted* text while the canvas *paints* the raw string. It is defensible: the UX line "its page is a consequence of the content above it and **can change with the data**" reads as data-dependence by design, and "**currently** occupies" reads as *with the data now bound*.

**(B) Build `layout.ColumnItem`s from the canvas's own shaping** of the raw authored string (which `addCanvasTextPaint` already performs at `page_setup.go:544-554`) and call `layout.Paginate` on those. This stays coherent with what the canvas actually paints, needs no bind coupling and no signature change, and still routes through the one function that decides how many pages a document has. It costs a **second derivation of the item extent**, which `paginate.go:98-104` warns against by name, and it requires a shaping pass deliberately separate from the (truncatable) paint so D-7.4.2 §5's independence survives.

**Why this is not rulable here.** The two survivors produce **different integers for the same template** — for any document with a content-band placeholder, (A) measures substituted text and (B) measures the literal `{{...}}` characters, and (A) additionally drops an element whose value binds to empty (`render.go:775-782`). Story 7.6 draws one sheet per reported window, so the choice is directly observable. Nothing selects: AD-17 and "the count must come from layout, not from `CanvasTextPaint`" (D-7.4.2 §5) are satisfied by both; `paginate.go:98-104` leans against (B); Story 7.4's landed test precedent leans toward (A); the epic's "the canvas is an approximation, the preview is exact" permits either.

**And it lands one story early.** D-7.4.4 recorded that the canvas passes `nil` substitutions and breaks the raw template string, with the consequence: *"If any story claims the canvas shows where the engine will break a **bound** value, that claim is false today. To be raised at **7.6's** plan gate rather than decided now."* Option (A) is exactly that claim, arriving at 7.5. A builder choosing it would be settling a question the lead explicitly reserved.

**A fourth disposition the lead may prefer:** report only the **page-height window** in this story and defer the **count** to 7.6, where the parity question was already scheduled. That satisfies the first half of AC4 and the whole of AC1–AC3 and AC5, and it is separably shippable — but it does not satisfy AC4 as written, so it needs explicit sanction rather than a builder's judgment.

### Judgments made at this gate, offered as recommendations (not blocking)

1. **What the content band's Y is bounded by after the lift: the JavaScript-safe geometry bound and nothing else.** AC1 says a Y beyond one page's content height is accepted; AC3 names the surviving upper bound. That is the ACs selecting, not a builder choosing.
2. **AC4 vs AC5 is an apparent, not a real, contradiction.** AC5 says the projection of every existing template is "unchanged" while AC4 adds a field to it. Read as "no pre-existing field's value moves, and the window count reports one window", both clauses stand — the same in-epic reading D-7.1.3 applied. Read as "no new key", AC4 and AC5 are literally unsatisfiable together. `fixtures/page-count-*` are outside AC5's premise: they do place content past page one.
3. **`engine-protocol.ts:193` lifts in the same commit.** D-7.4.5's consequence is explicit that a Go bound and its TypeScript mirror move together. Leaving it would accept the command and then blank the canvas with no attributable error.
4. **The live-drag clamp (`resize-anchor.ts:29` / `App.tsx:701`) stays.** Dragging a component onto a later sheet is Story 7.6's own acceptance criterion; there are no later sheets to drag onto until 7.6 draws them.
5. **An over-tall content element is not newly refused.** See Tasks Part 1.

### Limits to state, not to fix

- **D-7.4.4 stands and this story must not contradict it.** The canvas breaks the **raw** template string; it does not know where the engine will break a **bound** value. Whatever the ruling, the projection's comment must say what its number is a number *about*.
- **DW-33 is adjacent and needs a ruling, not a patch.** A text element whose first line already exceeds the per-line guard paints zero lines; whether it should paint a partial prefix is a design question 7.4's contract did not settle. If the window-count work touches that path, flag it.
- **No format field is required.** This lifts a validation bound and adds a projection field; the projection is not the file format. `SupportedMajor` stays 2.

## Verification

7.5 changes a command-validation bound and a channel schema on both sides, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). **Report measured pass/fail counts, never "green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:169`, P6g subtest, `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` (`:243`) must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- run **once per leg** with `FOLIO_MATRIX_TARGET` set: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` (`matrix_test.go:69-74`). **Unset, this test logs "asserts NOTHING" and returns — a no-op is not a pass.** Grep each leg's output for "asserts NOTHING" and report the count. Name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass, **239 tests / 32 files at baseline** plus whatever this story adds. The **4** pre-existing `only-export-components` lint warnings are not a regression; **a fifth would be**.
- `cd folio-designer && npm run test:e2e:compile` -- expected: pass. Browser e2e is deferred by D-000.4 and does not execute; do not claim it ran.

**Nine digests to report byte-identical** (all twenty in `goldenDigestRecord` must hold): statement-1 76,744 `114df1d6…`; statement-5 127,363 `70dce051…`; statement-20 269,884 `56bfbbd9…`; statement-50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` (`fontgen_matrix_test.go:64`) fails under `-tags=matrix` without `fontTools`; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, Story 15.2).

**Manual checks:**
- **Re-derive the `containComponent` call-site enumeration by grep at the closing revision** and confirm all twelve still route through the split validator.
- **Demonstrate end to end** that a component placed windows below the content top survives a command, a projection, and a round trip through the canonical bytes — not that a conditional changed.
- **Confirm `git status` is empty for `fixtures/`** before quoting any digest.

## Auto Run Result

### Dispatch 1 — 2026-08-31, plan only

Status: `blocked`
Blocking condition: `intent gap`
Baseline: `23a46470be4078dad62d625a64806a6adf3625a2` on `main`, tree clean
Directive: `Halt after planning.` — **no implementation code, no commits.**

**⚠ ORCHESTRATOR: when you amend `<intent-contract>` with the lead's ruling, set `status:` back to `draft`.** Step-01 of the next dispatch HALTs on `blocked spec supplied` otherwise.

**The gap, in one sentence.** AC4 requires the projection to report how many page-height windows the content column occupies; the cheap derivation is forbidden by name in `internal/layout/paginate.go:69-74`, and the two remaining derivations produce different integers for the same template with nothing in the intent selecting between them. Full framing in Design Notes, "The window-count gap".

**What the lead must rule:** derivation **(A)** reuse the render/bind machinery with the designer's real data, **(B)** build column items from the canvas's own raw-string shaping, or **(D)** report the window height in 7.5 and defer the count to 7.6 where D-7.4.4 already scheduled the canvas/bound-data parity question. Recommendation: the gap is genuine; (D) is the only option that does not settle D-7.4.4's reserved question a story early, but it does not satisfy AC4 as written and therefore needs explicit sanction.

**Five judgments made at this gate and NOT blocking** — recorded in Design Notes: the surviving content-band upper bound is the JavaScript-safe geometry bound; AC4/AC5's apparent contradiction resolves the same way D-7.1.3 resolved the epic header's; `engine-protocol.ts:193` lifts in the same commit; the live-drag clamp stays for 7.6; an over-tall content element gains no new refusal.

**Findings established by measurement at this baseline, load-bearing for whatever is ruled:**

- **The content cap, both repeating-band caps and the negative-coordinate refusal are ONE conditional producing ONE format string** (`component_commands.go:1767-1768`). There is no existing seam. "Lift the content Y cap" and "keep the negative refusal's message unchanged" are contradictory demands on one string until it is split.
- **No test asserts any of these messages** — `grep "must stay within" --include="*_test.go"` returns zero hits repo-wide. AC3's "message unchanged" is entirely unprotected today.
- **The designer-side cap does not come from `internal/layout`.** It is `page_setup.go:355`, an independent second spelling of `ContentHeight`. `page_setup.go` does not import `internal/layout` at all. **The scope fence holds: `paginate.go` is genuinely uninvolved in the refusal.**
- **The engine needs no pagination work.** No load-time or render-time check refuses a far-below content element; `fixtures/page-count-50` places elements ~49 windows down and renders to exactly 50 pages. The refusal is designer-side only.
- **`engine-protocol.ts:193` is a second, independent band-containment gate** that discards the whole snapshot with no attributable error. A Go-only lift leaves the story invisible in the app.
- **`TestSnapDoesNotPushAnEdgeDragOutOfItsBand` (`component_commands_test.go:583`) is the only shipped test that must be inverted**, at `:612-614`, `:624` and the `"a grid step past the bottom edge"` case at `:632`.
- **There is no multi-page canvas fixture anywhere.** `multi_page_fixture_test.go`, `statement_fixture_test.go` and `page_count_matrix_test.go` never call `Canvas`. This story adds the first.
- **No format field is needed.** `SupportedMajor` stays 2.

Verification was **not** run in this dispatch: no code changed. The `## Verification` section states what the implementing dispatch must measure.

## Delivery Log
