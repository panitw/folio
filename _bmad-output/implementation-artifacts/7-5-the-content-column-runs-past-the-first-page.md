---
title: 'Story 7.5: The content column runs past the first page'
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: '4bf201abadeafe29f7a9c19efb6f50b302a1b8b0'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized'] # oversized: the refusal-split surface, the Go/TS mirror inventory (R4) and the window-count derivation (R1) are three wide surfaces that must be stated, not summarised. NOT multiple-goals: the projection field is AC4 of this same story, not a second shippable.
deferred:
  - summary: >-
      The designer's live drag/resize clamp still bounds Y by band.height for all
      three bands, so an author dragging in the canvas cannot reach past page one
      even though the engine and the protocol guard now accept it.
    evidence: |-
      folio-designer/src/resize-anchor.ts:29 proposedBounds clamps y to
      limitHeight - originalHeight (:34-35) and a south resize to limitHeight
      (:52); App.tsx:701 fills DragLimit from { width: band.width, height:
      band.height } for every band, and DragLimit carries no band name, so the
      helper cannot tell the content band from a repeating one even in principle.
      resize-anchor.test.ts:46,53 still pins the clamp and is unmodified.
      Pre-existing, and the intent scopes this story to the command layer and the
      projection-admission mirror; Design Notes judgment 3 assigns it to Story 7.6,
      which is also the story that draws the later sheets a drag would target.
    location: >-
      folio-designer/src/resize-anchor.ts:29
    severity: medium
  - summary: >-
      contentWindowCount has no upper bound, and the non-text extent is an
      unguarded sum, so a content component placed at the JavaScript-safe ceiling
      yields a count near 2.5e10 that the browser guard admits.
    evidence: |-
      After the lift the only remaining ceiling on a content Y is
      MaxCanvasMillipoints (Design Notes judgment 1, which the ACs select). Text
      line tops go through canvasLineTop's JS-safe guard, but the non-text branch
      of addCanvasWindowCount builds Bottom: element.Y + height with a raw +, and
      engine-protocol.ts requires only Number.isSafeInteger and > 0 of the count.
      Nothing else on the projection is uncapped. Not reachable through any
      shipped template, and Story 7.6 is the story that would try to draw one
      sheet per window.
    location: >-
      folio-go/page_setup.go:496 (addCanvasWindowCount)
    severity: low
---

## In plain terms

*Non-normative. This section describes what shipped; the contract below governs implementation.*

An author can now place a component below the foot of page one, in the main content area and only there. The two repeating strips — the one at the top of every page and the one at the bottom — stay exactly one page tall, because that is what repeating means. Coordinates that are simply nonsense, negative or too large for a browser to hold safely, are refused in the same words as before.

The engine also reports two numbers the designer must never work out for itself: how tall one page's worth of content is, and how many of those page-fulls the column runs to. The second describes the column **as the canvas draws it today**, not as a prediction of the printed document. Where a table's rows come from data the canvas has never been given, the canvas counts the table's heading and none of its rows, so the number is a **floor** — the finished document may run longer, never shorter. Four shipped examples show this plainly: each reports one page-full while printing one, five, twenty and fifty pages.

One thing to expect and not mistake for a bug: **dragging** a component past page one still stops at the bottom of the first page. The new room is reachable by command, not yet by hand; the next story both draws the later pages and frees the drag. One test stays deliberately red throughout and must never be "fixed".

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
## Rulings applied — the intent gap is closed (2026-08-31)

Five engineering-lead rulings, each verified in code by the orchestrator. **Settled; do not re-open.**

### R1 — Derive the count from the REAL `layout.Paginate`, fed canvas extents
Call `layout.Paginate` — **never a second pagination** — with `ColumnItem`s built from **the extents the
canvas paint plan has already computed** (`canvasLineTop(originY, i, vm.Advance)` and the vertical
model). **Not a fresh shaping pass.** One shaping, two consumers: the paint plan and the window count.

**Why this is the intended usage, not an abuse.** Verified: `Paginate(g PageGeometry, items
[]ColumnItem) (Pagination, error)` — **no data, no bindings, no template**. It is designed to receive
caller-derived extents and contribute only placement. And `ColumnItem`'s own doc comment relocates the
hazard: *"A line's extent is … a quantity the vertical model already computed once. Re-deriving it here
would be a second derivation of the same number."* **The prohibition is against a second derivation of
EXTENTS, not of pagination.** So the thing that makes this safe is where the extents come from.

**Rejected — reuse the render/bind machinery**, on three independent grounds, any one sufficient: it
answers a different question ("how many pages will this render to") that the canvas cannot ask; feeding
it `{}` means changing AD-14's owner-settled rule that an absent path is an `Error` carrying the path;
and **Story 13.4 is that change, in another epic** — 7.5 must not pre-solve it. It is also the
bound-value parity claim D-7.4.4 reserved for 7.6's plan gate.

**Rejected — defer the count to 7.6.** AC4 exists in 7.5 *so that* 7.6 measures nothing — AD-17 operating
across a story boundary. Deferring pushes engine work into 7.6, whose every AC is about drawing. The
intermediate state is the worst of the three: an author could place a component below page one and the
canvas would still show one page, with no signal.

### R2 — AC4 is amended: the count is a claim about the CANVAS, not a prediction of the render
AC4 stays in 7.5 and gains one clause: the window count is reported **for the column as the canvas
currently paints it**, not as a prediction of the rendered document.

**The fact that forces it, verified:** `projectedSize` (`component_commands.go:1735-1744`) returns, for a
table, `(sum of column widths, element.Table.Value.HeaderHeight)` — **header height only, no rows**,
because the canvas has no data. So **for any document with a bound table in the content band the canvas
count is a FLOOR, not a prediction**, by construction and irreparably within 7.5. Pre-existing (Epic 6
shipped it). **Record it explicitly** rather than letting 7.6 discover it by drawing four sheets for a
fifty-page statement.

This also makes the projection agree with the UI: 7.6's AC already requires the interface to say a
component's page *"is a consequence of the content above it and can change when the data does — it is a
column position, not a pin to page three."*

**The fixture must DISCRIMINATE, not merely exercise.** There is no multi-page canvas fixture anywhere,
so 7.5 adds the first, and it must contain **an element declared far below the text**, asserting it
occupies **ONE** extra window, not ten. `ceil(lowestBottom / H)` gets that wrong by nine, so the fixture
red-proves the exact spelling `paginate.go:69-74` bans. Anything less exercises the code without
discriminating it.

### R3 — Split `containComponent` by what the constraint MEANS, and characterize before changing
- **Representational refusals** — `x < 0 || y < 0 || width < 0 || height < 0`, plus the JS-safe bound —
  apply to **every** band and keep the existing message **verbatim**. This is AC3's population.
- **Band-capacity refusals** — `y > band.Height`, `height > band.Height - y` — apply to the **repeating
  bands only**. The content band is a column, not a box, so it has no height cap. This is AC1's
  population.
- `x > band.Width` and `width > band.Width - x` stay universal: the column is unbounded **vertically,
  never horizontally**.

Split by meaning, not by clause count — the same axis D-7.3.1 got wrong (it split by JSON key location)
and the correction got right. A clause-by-clause split leaves the next reader unable to say which half a
new clause belongs in.

**CHARACTERIZE, THEN CHANGE — this step is not optional.** Verified: `grep "must stay within"` returns
**exactly two hits repo-wide and neither is a test** (`component_commands.go:1768`, and an unrelated
`column.footerOf` message at `:460`). AC3's "message unchanged" protects **nothing** today. So 7.5 must
**land the assertions against current behaviour FIRST, as their own step, and red-proof them by
perturbing the message** — then split. A story that splits first and asserts after is asserting whatever
it happened to produce. New assertions must not match `:460`'s footerOf message.

**Audit all twelve call sites** for which band each passes: a site passing the content band that *relied*
on the height cap changes behaviour silently. **Enumerate them in the story record**; do not assume the
split is transparent.

### R4 — The TypeScript gate lifts in the SAME commit, and the standing obligation widens
`engine-protocol.ts:193`'s band-containment gate lifts with the Go side. A Go-only lift ships a story
that is **invisible in the running app**.

This is the **third** occurrence of the shape, and it is a **fifth mirror DW-25 never covered** — DW-25
closed its own predicate (the four size caps); band containment is a different invariant that happens to
live in the same file. **An audit closes only what it measured.**

**The standing obligation widens** from "the size caps move together" to: **any invariant duplicated
across the Go/TS boundary moves in one commit, with a test that reads both sides.** 7.5 must **enumerate
the mirrors that exist in `engine-protocol.ts` today** rather than adding one more to an unlisted set.
That enumeration is the artifact — a rule with no inventory rots the way DW-24's anchors did, three
times, the last inside the commit that closed it.

### R5 — Collapse `page_setup.go`'s second `ContentHeight` spelling, here
In scope, and load-bearing rather than cosmetic. AD-13 is explicit: the content height is derived *"by
one function in `internal/layout`"*, so the inline `innerH - header - footer` is already an AD-13
violation. **AC4 makes it dangerous**: the projection must report the page-height window, that number
*is* `ContentHeight`, and it is what the canvas draws sheet boundaries from — reporting it from a second
spelling ships a divergence into the one place it would be invisible, where the canvas and the engine
draw different pages and agree on the bytes.

**The scope fence is not crossed.** `page_setup.go` is in package `folio`, the shell, which the spine
says *"composes the whole pipeline and imports every stage by design"*; stage ranks govern `internal/`
only. 7.5 imports `internal/layout` for `Paginate` regardless, so the collapse costs one line.


## Code Map

**Every anchor below was RE-VERIFIED at the baseline `83d49b6`.** `git log 23a4647..83d49b6` is two commits, **both `_bmad-output/` only** — no Go, no TypeScript moved — so the previous dispatch's anchors survive, with the corrections marked ⚠ below.

### The single refusal — `folio-go/component_commands.go`

- **`:1766-1771` `containComponent(band CanvasBand, x, y, width, height geom.Length) error`** — the **only** band-extent validation in the designer command path. **One `if`, eight disjuncts, one message:**
  ```go
  if x < 0 || y < 0 || width < 0 || height < 0 || x > geom.Length(band.Width) || y > geom.Length(band.Height) || width > geom.Length(band.Width)-x || height > geom.Length(band.Height)-y {
      return fmt.Errorf("folio: component geometry must stay within %s", band.Name)
  }
  ```
- ⚠ **THE LOAD-BEARING FACT.** The content-band one-page cap (`y > band.Height`, `height > band.Height-y`), the `pageHeader`/`pageFooter` caps, and the **negative-coordinate** refusal are **the same expression producing the same format string**. Only `%s` differs, from `band.Name`. There is no existing seam.
- ⚠ **CORRECTION: there are ELEVEN call sites, not twelve.** The twelfth grep hit at `:1766` is the definition. **The band-by-band audit is in Design Notes, "The eleven-call-site band audit".** Sites: `:165`, `:300`, `:829`, `:1409`, `:1472`, `:1565`, `:1569`, `:1606`, `:1668`, `:1674`, `:1726`.
- ⚠ **CORRECTION: no call site pins a band literal — every one of the eleven can receive ANY of the three bands.** `findComponent` `:1516-1530` loops `[]string{"pageHeader","content","pageFooter"}` at `:1517`; `bandByName` `:1285-1304` and `hitTestBand` `:1494-1515` switch over all three. The split must therefore key on `band.Name` **inside** `containComponent`, never at a call site.
- **Pull-back arithmetic** — `containEdge` `:1752-1758` (doc `:1746-1751`), `floorToGrid` `:1759-1765`, `GridIncrement = 6000` (`page_setup.go:19`). Eight `containEdge` calls in three functions: `dropComponent` `:1410`(x)/`:1411`(y), `moveComponent` `:1566`(x)/`:1567`(y), `setComponentBounds` `:1669`(w)/`:1670`(h)/`:1671`(x)/`:1672`(y).
- ⚠ **CORRECTION to the previous dispatch's reading: the pre-clamps are NEVER unconditional.** Each is gated on `snap && containComponent(unsnapped…) == nil`. So they do **not** make a lifted validator "accept nothing new" — the opposite: lifting the cap **widens the gate** and makes the Y pull-backs fire on inputs that today bypass them entirely (sites `:1409`, `:1565`, `:1668`). See Design Notes.
- **Band identity is a bare `string`** — `CanvasBand.Name` (`page_setup.go:156-162`). **No named constant exists.** Load-bearing literal sites: `page_setup.go:354/355/356` (where the names are minted), `:417-419`, `:514-516`, `:721/723/725`; `component_commands.go:1295/1297/1299`, `:1505/1507/1509`, `:1517`; `internal/template/serialize.go:188-190`; `internal/template/parse_bands.go:24/36/40/44`.
- `applyPropertyChanges` `:894-1190` writes `X/Y/Width/Height` at `:962-968` and enforces **only positivity** at `:953-955` — `x` and `y` have **no bound at all** there. `:829` is the sole band bound for the entire property-edit surface.

### The separate, surviving refusal — the JS-safe geometry bound

- `MaxCanvasMillipoints int64 = 9007199254740991` — `page_setup.go:25`.
- `lengthField` `page_setup.go:1043-1068`, enforced `:1063`: `"%s exceeds the JavaScript-safe geometry bound"`. Wrapped by `componentLength` `component_commands.go:1258-1268` (`"folio: component.%s: %w"`) and `propertyLength` `:884-886`.
- Siblings to retain verbatim: `page_setup.go:732` `"folio: component exceeds the JavaScript-safe geometry bound"`; `page_setup.go:338` `"folio: page setup exceeds the JavaScript-safe geometry bound"`.
- ⚠ `lengthField` permits **negative** values down to `-MaxCanvasMillipoints`. **Negativity is caught ONLY by `containComponent`'s `x < 0 || y < 0 || …` terms** — which is why those terms are load-bearing and must stay universal.
- ✅ Separate path, separate message, upstream of `containComponent` — untouched by any split.

### The band rectangle and the two `ContentHeight` spellings — `folio-go/page_setup.go`

- **`:341`** `innerW, innerH := w-m.Left-m.Right, h-m.Top-m.Bottom`; guard **`:342-344`** (`header >= innerH-footer` → `"folio: page setup leaves no positive content region"`); **`:353-357`** the three bands, open-coded. Content is `Height: int64(innerH - header - footer)` at **`:355`** — **R5's target.**
- `layout.ContentHeight(g)` — `internal/layout/band.go:75-77` — is `g.Height - g.MarginTop - g.MarginBottom - g.PageHeaderHeight - g.PageFooterHeight`. ✅ **Verified arithmetically identical** to `:355`: same five terms, same signs, same `geom.Length` int64 arithmetic, no rounding.
- ⚠ **`page_setup.go` does NOT import `internal/layout`** — full import block `:1-15`. ✅ No cycle risk: `internal/layout` imports only `fmt`, `slices`, `strconv`, `internal/geom`, `internal/pagemodel`; `render.go:14` already imports it; both files are `package folio`.
- ⚠ **`pageGeometryOf` (`render.go:291-311`) MUST NOT be the canvas's source of `PageGeometry`.** It delegates to `pageDimensions` (`render.go:75-91`), whose `default` branch **hard-errors on `"Letter"`** — while `canvasDimensions` (`page_setup.go:904-920`) **supports Letter** at `:908-909`. A Letter document projects a canvas today; routing the canvas through `pageGeometryOf` would break it. See Design Notes, "Ruling A".
- Frame fact for the new fields' comments: `layout.Origins` (`band.go:101-109`) measures **downward from the printable top edge** (`Content: g.PageHeaderHeight`); `CanvasBand.Y` is **paper-absolute** (`content.Y = m.Top + header`). Exactly: `CanvasBand{content}.Y == MarginTop + layout.Origins(g).Content`.

### The projection struct — `folio-go/page_setup.go`

- **`CanvasProjection` `:265-294`** — lowerCamelCase json tags, **no `omitempty` on any top-level field**. `FontFamilies` `:286` and `DefaultFontSize` `:293` are the newest additions; their comments (`:279-292`) justify **why the engine, not the browser, owns the number** — **that is the tone the two new fields must match.**
- `CanvasBand` `:156-162`; `CanvasComponent` `:163-215`; `CanvasTextPaint` `:134-147` (`Truncated` `:145`).
- **No page-window or window-count field exists anywhere.** `bands[1].height` is already exactly the page-height window, stated for one page.
- `canvasComponents` `:716-776` does **not** re-check containment — only `> MaxCanvasMillipoints` at `:730-734`. **`Canvas()` already projects out-of-band content elements happily; no Go projection change is needed for admission.**

### The canvas paint plan — the ONLY source of extents R1 permits

- **`addCanvasTextPaint` `page_setup.go:503-669`**, called only from `CanvasWithTextPaint` at `:377`.
- Shaping `:544` (`shapeSegments` on the **raw** `element.Value.Value`), `:548-549` (`atomicSpansFor(..., nil)`, nil substitutions), `:554` `lines := packLines(...)` — **`lines` is the FULL, untruncated list and is never reassigned.**
- ⚠ **Truncation has TWO sites, both of which the count must ignore:** `:567-570` (`painted := lines`, sliced to `maxCanvasBodyTextLines = 1920`, `page_setup.go:82`) **before** the loop, and `:658-661` (`oversized || !budget.admits(...)` → `break`) **inside** it. `paint.Truncated` is set at both.
- ✅ **The code already states the rule the count needs**, `:571-575`: *"Overflow and the vertical origin below are still derived from the FULL line list, so the prefix paints at exactly the coordinates it occupies in the whole block."*
- Vertical model `:581` `vm, err := chainVerticalModel(chain, fontSize, styleLineSpacing(element.Style), fs, cache)`. `verticalMetrics` — `wrap.go:514-534`: **`FirstBaseline` `:516`, `Advance` `:519`, `LastDescent` `:533`**. `LastDescent`'s own doc (`:528-531`) already names its two consumers as *"the per-line item extent the page splitter reads, and textBlockHeight"*.
- Origin `:595` `originY := element.Y + textValignOffset(valign, boxHeight, textBlockHeight(len(lines), vm))` — **computed from `len(lines)`, untruncated by construction.**
- Paint loop `:597` `for i, line := range painted`; `:598` `top, err := canvasLineTop(originY, i, vm.Advance)`. `canvasLineTop` `:709-714` = `elementY + index*advance` with a JS-safe overflow guard.
- ⚠ **The extent formula is already written in the render path and must be copied term for term, not re-derived** — `render.go:937-938`:
  ```go
  placed[j].itemTop    = lineY
  placed[j].itemBottom = lineY + vm.FirstBaseline + vm.LastDescent
  ```
  with the comment at `render.go:918-924`: *"computed here from the vertical model that is already in hand … Same numbers, no re-derivation."* **The canvas's `top` at `:598` is `lineY`'s exact analogue.**
- **DW-33's path is `:658-661`** (comment `:655-657`), reached when `i == 0` already fails `budget.admits` or `oversized` (`:644-648`). ✅ **A count that iterates `lines` and uses `vm` arithmetic never reads `painted`, `budget`, `oversized` or `placed`, so it is PROVABLY independent of DW-33.** That is the flag the contract's Block If asks for; it is not a decision.
- Degradation idioms already in this loop, to be matched not invented: font-chain failure `:526-534` → empty `TextPaint`, `continue`, **whole projection survives**; empty/unset value `:536-539` → same.

### `layout.Paginate` — the one function that decides how many pages a column has

- **`Paginate(g PageGeometry, items []ColumnItem) (Pagination, error)`** — `internal/layout/paginate.go:548`. **No data, no bindings, no template**, exactly as R1 states.
- `PageGeometry` — `band.go:54-59`: `Width, Height, MarginTop, MarginBottom, MarginLeft, MarginRight, PageHeaderHeight, PageFooterHeight`.
- `ColumnItem` — `paginate.go:105-144`: `ElementID` `:108`, `Top, Bottom` `:111`, `Runs` `:125`, `Images` `:126`, `Rects` `:127`, `Group` `:143`.
- **The prohibition R1 relocates, verbatim `paginate.go:98-104`:** *"WHY THE EXTENT IS CARRIED RATHER THAN DERIVED HERE … Re-deriving it here would be a second derivation of the same number."* — **against a second derivation of EXTENTS, not of pagination.**
- **The forbidden closed form, verbatim `paginate.go:69-74`:** *"PAGE COUNT IS NOT A CLOSED FORM … it is NOT `ceil(lowestBottom / H)` … an element declared far below the text STARTS THE NEXT WINDOW rather than generating blank pages before it."*
- **The count is `len(plan.Pages)`** — there is no count field. `render.go:1675` reads it exactly that way.
- ⚠ **`Paginate` has an EXCLUSIVITY PRE-PASS, `:567-584`.** Every item must populate **exactly one** of `Runs`/`Images`/`Rects`; zero or two returns `*MixedItemError` (`:579`, `:582`) **before the sweep**. The refs are index-only types (`TextRunRef`/`ImageRef`/`RectRef` = `int`, `:185-194`).
- ✅ **The dummy-ref idiom is already sanctioned in-tree**, `page_number.go:141-146`: *"the actual RectRef values are never read back … so they need only be non-empty, one per rect, to satisfy Paginate's exclusivity check."*
- ⚠ **`*OverflowError` at `:1075-1083`** fires when an **ungrouped** item's own height exceeds `ContentHeight(g)`. **This story newly makes such a component authorable** (Tasks Part 2 forbids refusing it), so the canvas count can hit it. See Design Notes, "Ruling C".
- **Zero items → exactly ONE page**, `:545-557`: *"A document with NO content items is ONE page, not zero."*
- `contentColumnItems` `page_number.go:87-155` is the render path's builder. ⚠ **It is NOT directly reusable** — it takes `[]textRunSource`, a render type carrying glyphs, faces, CIDs and page slots. Its **composition rule** is what 7.5 mirrors, not its signature.
- Render-path composition, for reference: text per line (`:89-109`), images by **declared box** `Top: r.y, Bottom: r.y + r.boxH` (`:124-125`), table/element-box rects (`:138-153`). `collectElementBoxRects` (`element_box.go:52`) additionally **skips unstyled elements** (`:71-75`) — a render-only filter 7.5 does **not** copy; see Design Notes, "Ruling D".

### `projectedSize` — why a bound table makes the count a FLOOR (R2)

- `component_commands.go:1735-1744`, verbatim:
  ```go
  func projectedSize(element template.Element) (geom.Length, geom.Length) {
      if element.Type != template.ElementTable {
          return element.Width.Value, element.Height.Value
      }
      var width geom.Length
      for _, column := range element.Table.Value.Columns {
          width += column.Width
      }
      return width, element.Table.Value.HeaderHeight
  }
  ```
  ✅ **Header height only, no rows** — the canvas has no data. Consumed at `page_setup.go:729`, written to `CanvasComponent.Width/Height` at `:735`.
- Element kinds — `internal/template/model.go:177-181`: `text`, `image`, `table`, `line`, `rect`.

### Paint independence, landed by Story 7.4 — must survive and be re-asserted

- `TestPaginationIsIndependentOfCanvasPaintTruncation` — `folio-go/canvas_body_text_bounds_test.go:237-278`; framing comment `:231-236` names Story 7.5 explicitly. Its `pages` closure `:242-258` is `documentBands` → `collectBandTextRuns(..., data, data, ...)` → `contentColumnItems(runs, nil, nil, nil)` → `layout.Paginate(mustPageGeometry(t, tpl), ...)` → `len(plan.Pages)`. Presence precondition `:264-266`; equality `:267-270`; non-coincidence `:274-276`.
- Helpers: `emptyBindValue` `shaped_fixture_test.go:1446-1453`; `mustPageGeometry` `collect_text_runs_composition_test.go:102-109`.
- ⚠ **That closure is the RENDER-path oracle and it stays.** 7.5's field uses the canvas path (R1). The story must therefore **assert the two agree on a text-only fixture**, not assume it.
- Prohibited regexes — `folio-designer/src/canvas-authority-contract.test.ts`, rationale `:29-37` (names a **window count** explicitly), entries **`:38`** `/\b(?:textPaint|paint)\??\.lines\.length\b/` and **`:39`** `/\blines\.length\s*[*/]|[*/]\s*\blines\.length\b/`. Non-vacuity `:52-55`; red-proofs `:70-71`.
- ✅ **Nothing in production TS derives from `paint.lines.length`.** The one read is `engine-protocol.ts:291`, spelled `value.lines.length` with a `>` — it survives both regexes. ⚠ **Renaming that parameter from `value` to `paint` would make regex `:38` fire on a legitimate bound check.** Do not rename it.

### The browser mirrors — `folio-designer/src/engine-protocol.ts` (346 lines)

- ⚠ **`:193`** the band-containment gate, inside `isCanvas`'s `components.every`, with `band` bound at `:189` and `box` at `:190`:
  ```ts
  if (!(box.x + box.width <= band.width && box.y + box.height <= band.height)) return false
  ```
  **Traced consequence of `parseInbound` → `undefined`:** `isCanvas` false → `isSnapshot` false (`:311`) → `parseInbound` returns `undefined` (`:335`) → `engine-client.ts:85-87` `#fail('PROTOCOL_INVALID', …)` → `:117-124` **terminates the worker**, rejects every in-flight request → `App.tsx:702` renders `Waiting for Go page geometry.` **One millipoint out of band kills the session, with no element id and no attribution.** A Go-only lift makes 7.5 invisible in the app.
- **THE MIRROR INVENTORY (R4) is in Design Notes, "The `engine-protocol.ts` mirror inventory".** Headline: **6 of ~35 duplicated invariants are tied**; band containment is **untied**, and DW-25 closed only the four size caps.
- Companion band invariants to leave alone: `:171` (paint inside page), `:172-175` (band contiguity), `:165-169` (three bands, that order).
- Exact-key `hasOnly` `:139` (strict sibling `hasExactKeys` `:140`). **The canvas/page allow-list a new top-level field must join is `:156`:**
  ```ts
  hasOnly(value, ['width','height','orientation','preset','marginTop','marginRight','marginBottom','marginLeft','gridIncrement','commandWidth','commandHeight','fontFamilies','defaultFontSize','bands','components'])
  ```
  Other allow-lists: band `:169`, component `:184`, image `:248`, textPaint `:291`, line `:300`, fragment `:307`, snapshot `:311`, envelope `:333`.
- **Value predicates:** helper `:157` `const integer = (key, positive = false) => … Number.isSafeInteger(value[key]) && (positive ? value[key] > 0 : value[key] >= 0)`; applied `:158` — **a strictly-positive millipoint field joins the FIRST array on `:158`.**
- Type decls: `CanvasProjection` `:95-104` (mixed tabs/spaces — match the neighbour), `EngineSnapshot` `:84-91`.
- `engine-bounds-mirror.test.ts` — `goSources` `:33-36` (**`page_setup.go` and `internal/template/linespacing.go` only**), `tsPath` `:37`, `pairs` `:47-54` (**6**, `toHaveLength(6)` at `:81`), extractors `:59-65` (`^(?:const[ \t]+|[ \t]+)NAME = (\d+)$` / `^export const NAME = (\d+)$` — **decimal literals only**), site-consumption `:97-99`, red-proof `:110-128`. ⚠ **It ties numerals and names, and one narrow "the constant is actually consumed" check. It ties NO behavioural predicate** — not containment, not contiguity, not vocabularies.
- ⚠ `engine-protocol.test.ts` — the sole out-of-band case is `:26`, `x: 991` in a **1000-wide** band; fixture content band `:4` is `{ name:'content', x:0, y:100, width:1000, height:1800 }`. **Every content component in the file has `y: 0` and `y+height <= 1800`, so nothing exercises the Y conjunct at all.** A Y-only lift leaves this file green and vacuous.
- Fixtures that must gain the new fields — ⚠ **exactly three base literals**, everything else is a spread: `engine-protocol.test.ts:4`, `App.test.tsx:17`, `DataPanel.test.tsx:18`.
- **Third gate, and it is Story 7.6's:** `resize-anchor.ts:29` `proposedBounds(anchor, origin, dx, dy, limit?)`, `DragLimit` `:23-25`, clamps at `:34-35`/`:50`/`:52`; filled from `{ width: band.width, height: band.height }` at `App.tsx:701`, applied at `App.tsx:1176`. Pinned by `resize-anchor.test.ts`. Its own header (`:9-18`) says it is a UX affordance, not a second authority.
- ⚠ **Positional fragility.** `canvas-authority-contract.test.ts:83-88` requires `export function placementPoint(event: Pick<MouseEvent,` … `\n}\nfunction pageStyle` to be **textually adjacent**, and `expect(source).toMatch(seam)` at `:87` makes a break a hard failure. **Any new helper goes AFTER `pageStyle` (`App.tsx:1165`), never between `:1164` and `:1165`.**

### Tests pinning today's behaviour

- ⚠ **The message text is asserted NOWHERE.** `grep -rn "must stay within" --include="*_test.go"` → **zero hits**; `"component.geometry"` in tests → **zero hits**; the JS-bound message in tests → **zero hits**. **AC3's "message unchanged" protects nothing today.**
- ⚠ **The collision R3 warns about is live.** `component_commands.go:460` is `componentFailure(id, "column.footerOf", "footerOf must stay within the table collection")` — unrelated, and **exercised by `TestTableDataBindingAndFooterCommandsAreCanonicalAndTransactional` (`component_commands_test.go:407`)**. **New assertions must compare the FULL string, never `strings.Contains(err, "must stay within")`.**
- **`TestSnapDoesNotPushAnEdgeDragOutOfItsBand` — `component_commands_test.go:583-642` — the one test that must be inverted.** Content band `:591-596`; `edgeX, edgeY := band.Width-created.Width, band.Height-created.Height` at `:603`; grid-multiple `:609-611`; containment `:612-614` and `:624-626` (**X and Y in one conditional each — the Y half must change, the X half must survive**); grid-step distance `:615-617`; refusal map `:631-633` with **`"a grid step past the bottom edge"` at `:632`** (must invert), X cases `:631`/`:633` (must survive); refusal + byte-immutability loop `:635-640`.
- ⚠ **After the lift the Y arm of `:615-617` becomes VACUOUS** (nothing pulls Y back, so `edgeY - component.Y` is 0). A new Y case must replace its discriminating power.
- Survive unchanged (verified): `TestComponentCommandsSnapContainAndFailureAreTransactional` `:73` (its out-of-band probe at `:101` is **X**, `x:999999`, `y:0` — the X-ness is load-bearing); `TestSetComponentBoundsMovesOriginAndSizeInOneCommand` `:548` (refusal map `:568-573` is **all negative-X or width-based; NO `y`/`height` overflow probe exists** — the natural home for a new "tall content Y is accepted" case); `TestTableColumnCommandsAreTransactionalAtThePublicSeam` `:350` (width-driven, reaches `:829`); `TestTableColumnRejectionsDoNotMutate` `:321`; `TestComponentCommandsRejectTableResizeAndPreserveTableGeometry` `:250`.
- ⚠ **TRIPWIRE — `TestDropComponentUsesGoHalfOpenBandHitTesting` `:514`.** Band **hit-testing** at `component_commands.go:1501` (`y >= top+band.Height`) is a **different concern** from `containComponent`: it resolves *which band a point is in*. **It must not move.** Lifting it would make a content Y past the band resolve to `pageFooter` or to nothing.
- **Coverage gap confirmed: `component_properties_test.go` has NO band-bounds coverage at all** — every component it creates sits at a safe interior position (`:41`, `:65`). The path through `component_commands.go:829` with a large content-band `y` is untested on both sides of the change.
- Fixture `componentTemplate` `component_commands_test.go:13-24` → `testdata/template/golden/worked-example.json`: A4, 36pt margins, header 60, footer 30 ⇒ content band **523276 × 679890 mp** ✅ (re-verified). ⇒ `edgeX = 451276`, `edgeY = 655890`.

### Existing multi-page evidence — the engine needs no work

- **`fixtures/page-count-{1,5,20,50}`** — `page_count_matrix_templates.go`, shape comment `:11-21`, templates `:28-55`/`:57-88`/`:90-136`/`:138-214`. Geometry: A4, margins `{30,54,42,36}`, header 18, footer 24 ⇒ **content height 727890 mp**; elements at `y = i*728pt`, **one window apart**. Tests `page_count_matrix_test.go:28-39`, `:44-60`, `:76-113`.
- ⚠ **CRITICAL: these fixtures CANNOT discriminate the closed form.** One-window spacing makes `ceil(lowestBottom/H)` and the true slide count **coincide at every N**. That is why the forbidden spelling survived until now, and why R2's new fixture must use a **GAP**.
- **No load-time or render-time refusal of a far-below content element exists.** `parse_bands.go:147-151` → `decodePointsRaw` (`parse.go:428-438`) check representability only. `Paginate`'s only positional error is `itemHeight > height` (`:1075`) — the item's **own height**, never its position.
- ⚠ **No canvas/projection fixture is multi-page, and none COULD be** — `containComponent`'s `y > band.Height` clause makes such a template unauthorable today. `multi_page_fixture_test.go`, `statement_fixture_test.go` and `page_count_matrix_test.go` **never call `Canvas` or `CanvasWithTextPaint`** (verified: zero hits). **Every existing canvas assertion about pagination is vacuous by construction. Story 7.5 adds the first fixture able to express the defect.**

### The projection seam — which entry point reaches the browser

- ✅ **VERIFIED: every projection that reaches the browser comes from `CanvasWithTextPaint`** — `wasm/engine.go:119`, `:255`, `:294` are the only three seams, and all three call it.
- ✅ **Every mutating command's own `Canvas(t)` is DISCARDED and recomputed by `wasm/engine.go`.** This is already documented at `page_setup.go:388-392`: *"every mutating command's own Canvas(t) is discarded and recomputed by wasm/engine.go, so the paint must be derivable from template state alone, exactly like text paint."* **That is the existing architectural rule that places the window count in the paint plan — see Design Notes, "Ruling B".**
- `Canvas` `:323-367`; `CanvasWithTextPaint` `:372-384` = `Canvas` + `addCanvasTextPaint` `:377` + `addCanvasImagePaint` `:379`.

### Verification surface

- `folio-go/byte_neutrality_test.go:92-460` `goldenDigestRecord` — ✅ **20 entries** (re-counted); "invalidated IN WHOLE" clause `:225`. Fixture dirs with **no** entry: `expected-breaks`, `hidden-image`, `page-count-1`, `page-count-5`, `page-count-50`, `thai-break-corpus`.
- `matrix_test.go:69-74` `matrixTargets` (4), `wantMatrixLegs = 4` `:84`; `TestTargetRenderHash` `:1979`, gate `:1980-1990` (**logs "asserts NOTHING" and returns when `FOLIO_MATRIX_TARGET` is unset**); a **second** such no-op branch at `:2076-2091`; `TestCrossTargetByteIdentity` `:1802`.
- `internal/text/corpus_test.go:169` `TestCorpusMeetsP6ExerciseFloors`; **P6g floor at `:184`** (`want 20`, actual 7); drift twin `TestCorpusP6StatsMatchDeclaredBaseline` `:243` (`baselineP6g = 7` at `:251`) **must stay green**.
- `fontgen_matrix_test.go:64` `TestShippedFacesReproduceFromUpstream` — environmental. `lint/internal/rules/licencegraph_test.go:112` — gofmt-dirty (DW-23).
- Designer: **32 vitest files** (30 `src/**`, 2 `scripts/**`); static `it`/`test` count **231** — a lower bound, since `it.each` expands at runtime (the dispatch's runtime baseline is **239 / 32 files**). ✅ **oxlint at baseline: exactly 4 warnings, 0 errors**, all `only-export-components` (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1155,1162`). ⚠ **Adding a non-component export to any `.tsx` raises that count — keep new helpers in a `.ts` file.**

## Tasks & Acceptance

**The ordering below is normative.** R3 requires characterization to LAND BEFORE the split; Part 1 is therefore its own step and must be committed-ready before Part 2 begins.

**Execution — Part 1: CHARACTERIZE (must be complete before Part 2 changes anything).**

- `folio-go/component_commands_test.go` -- **Add assertions on the CURRENT refusal messages, by FULL-STRING equality**, for: negative `x`, `y`, `width`, `height` in each of the three bands; a content `x` past the band width; a `pageHeader` `y` and a `pageFooter` `y` past their band heights; **and — asserted here while it is still true — a content `y` past the band height**. Use `err.Error() == "folio: component geometry must stay within <band>"`; **never `strings.Contains`**, which also matches `component_commands.go:460`'s unrelated `footerOf` message. -- Rationale: R3. `grep "must stay within" --include="*_test.go"` returns zero hits, so AC3's "message unchanged" protects nothing; a story that splits first and asserts after is asserting whatever it happened to produce.
- `folio-go/component_commands_test.go` -- **Add the same full-string assertion for the JS-safe bound's distinct message** from `componentLength` (`"folio: component.y: y exceeds the JavaScript-safe geometry bound"`), and for `page_setup.go:732`'s projection-time sibling. -- Rationale: it is the surviving upper bound after the lift (Design Notes judgment 1) and is equally unasserted today.
- `folio-go/component_commands.go` + this spec's Delivery Log -- **RED-PROOF the characterization before proceeding**: perturb `component_commands.go:1768`'s format string locally, confirm the new assertions fail, and revert. **Record in the Delivery Log that this was done and what failed.** -- Rationale: R3 requires red-proof, not merely presence; an assertion that cannot fail is not characterization.

**Execution — Part 2: the refusal split.**

- `folio-go/component_commands.go` -- **Split `containComponent` (`:1766-1771`) BY MEANING, keying on `band.Name` inside the function** (no call site can be keyed — all eleven receive any band). **Representational refusals** (`x<0 || y<0 || width<0 || height<0`) stay universal. **Horizontal caps** (`x > band.Width`, `width > band.Width-x`) stay universal. **Band-capacity refusals** (`y > band.Height`, `height > band.Height-y`) apply to `pageHeader` and `pageFooter` **only**. **Preserve `"folio: component geometry must stay within %s"` byte-for-byte for every surviving clause.** -- Rationale: R3. Split by meaning, not clause count — the axis D-7.3.1 got wrong and its correction got right.
- `folio-go/component_commands.go` -- **Introduce named constants for the three band names** and use them in the split, in `bandByName` `:1295-1299`, `hitTestBand` `:1505-1509`, `findComponent` `:1517`, `canvasComponents` `page_setup.go:721-725` and the band mint at `page_setup.go:354-356`. -- Rationale: the split makes band identity load-bearing for the first time; a fifth inline spelling of a bare string is how it silently diverges.
- `folio-go/component_commands.go` -- **Do NOT add any new refusal for a content element taller than one window**, and **do NOT touch `hitTestBand`'s `:1501` half-open band rectangle.** -- Rationale: D-7.4.2's rejected option (c) — a canvas/command bound must never become a document validity rule; and `:1501` answers "which band is this point in", a different question, pinned by `TestDropComponentUsesGoHalfOpenBandHitTesting`.
- `folio-go/component_commands.go` -- **Make the `containEdge` pull-backs correct under the lift**: a content-band pull-back must no longer clamp Y against one window, while still clamping X against the band width (`:1411`, `:1567`, `:1670`, `:1672`). -- Rationale: ⚠ the pre-clamps are gated on the probe succeeding, so lifting the cap **widens** the gate and makes the Y pull-backs fire on drags that today bypass them entirely — sites `:1409`, `:1565`, `:1668` change behaviour silently otherwise. See Design Notes, "The eleven-call-site band audit".

**Execution — Part 3: re-point the tests the split moves.**

- `folio-go/component_commands_test.go` -- **Invert `TestSnapDoesNotPushAnEdgeDragOutOfItsBand` (`:583-642`) on the Y axis only**: `"a grid step past the bottom edge"` (`:632`) becomes an **accepted** case; the Y halves of `:612-614` and `:624-626` stop requiring `Y+Height <= band.Height`. **Keep both X cases (`:631`, `:633`), the grid-multiple assertion (`:609-611`) and the X arm of `:615-617`.** -- Rationale: it is the only shipped test pinning the content-band Y cap.
- `folio-go/component_commands_test.go` -- **Add a Y-axis case that still discriminates**: a content drag far below `band.Height` is **accepted**, **still snaps to the grid**, and persists at that Y in the canonical bytes. -- Rationale: ⚠ after the lift the Y arm of `:615-617` is vacuous (nothing pulls Y back); without a replacement, Y coverage evaporates silently.
- `folio-go/component_commands_test.go` -- **Add a "tall content Y is accepted" case to `TestSetComponentBoundsMovesOriginAndSizeInOneCommand` (`:548`)**, whose refusal map `:568-573` today contains no `y`/`height` overflow probe at all. -- Rationale: the combined move+resize seam has zero coverage of the clause being lifted.
- `folio-go/component_properties_test.go` -- **Cover `updateComponentProperties` with a large content-band `y`** (the path through `component_commands.go:829`), accepted after the change, and a `pageHeader` `y` past its band, still refused by message. -- Rationale: `:829` is the **sole** band bound for the entire property-edit surface (`applyPropertyChanges` bounds neither `x` nor `y`), and the file has no band-bounds coverage today.

**Execution — Part 4: the browser mirror lifts in the SAME commit (R4).**

- `folio-designer/src/engine-protocol.ts` -- **Lift `:193`'s vertical containment for the content band**, leaving the horizontal check and the band-level invariants `:171`/`:172-175` intact. -- Rationale: `parseInbound` returning `undefined` terminates the worker and blanks the canvas with no attribution; a Go-only lift ships a story invisible in the app.
- `folio-designer/src/engine-protocol.test.ts` -- **Add a content component with `y + height > band.height` that must be ADMITTED, and a `pageHeader`/`pageFooter` component with `y + height > band.height` that must still be REJECTED**, keeping the X-overflow rejection at `:26`. -- Rationale: ⚠ every content component in the file has `y: 0`; the Y conjunct is unexercised, so a Y-only lift leaves the file green and vacuous.
- `folio-designer/src/engine-bounds-mirror.test.ts` -- **Widen the standing obligation from "the size caps move together" to "any invariant duplicated across the Go/TS boundary moves in one commit, with a test that reads both sides", and add a test that reads BOTH sides of band containment** — asserting that `component_commands.go`'s split and `engine-protocol.ts:193` agree on which bands cap vertically. Update `toHaveLength(6)` at `:81` if a `pairs` entry is added, and extend `goSources` `:33-36` to reach `component_commands.go`. -- Rationale: R4. This is the **third** occurrence of the shape and a **fifth mirror DW-25 never covered** — DW-25 closed the four size caps; band containment is a different invariant in the same file. **An audit closes only what it measured.**
- This spec's Delivery Log -- **record the mirror inventory** (Design Notes, "The `engine-protocol.ts` mirror inventory") in the Delivery Log as the story's artifact. -- Rationale: R4 — a rule with no inventory rots the way DW-24's anchors did three times.

**Execution — Part 5: collapse the second `ContentHeight` spelling (R5) and build one `PageGeometry`.**

- `folio-go/page_setup.go` -- **Construct a single `layout.PageGeometry` inside `Canvas` from the values already in scope at `:332-341`** (`canvasDimensions`' `w`/`h`, `m`, `header`, `footer`), and **replace `:355`'s `int64(innerH - header - footer)` with `int64(layout.ContentHeight(g))`.** Add `internal/layout` to the import block `:1-15`. Keep the `:342-344` guard as it stands. -- Rationale: R5 — AD-13 requires one function in `internal/layout`, and AC4 makes it load-bearing: the reported window height **is** that number, and a divergence would leave the canvas and the engine drawing different pages while agreeing on the bytes.
- `folio-go/page_setup.go` -- ⚠ **do NOT call `pageGeometryOf` (`render.go:291`).** It routes through `pageDimensions`, which **hard-errors on `"Letter"`** while `canvasDimensions` supports it. -- Rationale: a Letter document projects a canvas today; routing through `pageGeometryOf` would regress it, violating the contract's corpus/projection-neutrality rule. See Design Notes, "Ruling A".

**Execution — Part 6: the window count, from the paint plan's own extents (R1).**

- `folio-go/page_setup.go` -- **Build `[]layout.ColumnItem` for the CONTENT band from extents already computed, and call `layout.Paginate` once.** For a text element, inside `addCanvasTextPaint`'s per-element block **after `originY` (`:595`)**, iterate the **full `lines`** (never `painted`) and emit one item per line with `Top = originY + i*vm.Advance` (the value `canvasLineTop` returns) and `Bottom = Top + vm.FirstBaseline + vm.LastDescent` — **the same terms as `render.go:937-938`, copied, not re-derived.** Translate into the printable frame by adding `layout.Origins(g).Content`; **do not add `MarginTop`.** For every other content-band component, emit one item from the box `canvasComponents` already projects via `projectedSize`. Give each item exactly one non-empty ref slice, following the sanctioned dummy-ref idiom at `page_number.go:141-146`. **The count is `len(plan.Pages)`.** -- Rationale: R1 — one shaping, two consumers. `paginate.go:98-104` forbids a second derivation of **extents**, not of pagination; `Paginate` takes no data, no bindings and no template by design.
- `folio-go/page_setup.go` -- **Compute the count in `CanvasWithTextPaint`, as a third paint producer beside `addCanvasTextPaint` and `addCanvasImagePaint` — never in `Canvas`.** `Canvas` reports the geometric floor of **one** window, and its comment must say so. -- Rationale: the count needs shaping, which needs a `FontSet` that `Canvas` does not have; and `page_setup.go:388-392` already establishes that every command's `Canvas(t)` is discarded and recomputed by `wasm/engine.go` with fonts, so **every projection reaching the browser is a `CanvasWithTextPaint`** (verified at `wasm/engine.go:119`, `:255`, `:294`). See Design Notes, "Ruling B".
- `folio-go/page_setup.go` -- **Never fail the projection because pagination could not be computed.** On any `layout.Paginate` error — including `*OverflowError` for the over-tall content element this story newly permits — report **one** window and keep the projection. -- Rationale: Part 2 forbids turning the render path's overflow into a canvas refusal; R2 already establishes the count as a floor; and the file's existing degradation idiom (`:526-534`) keeps the projection alive on a font-chain failure. See Design Notes, "Ruling C".
- `folio-go/page_setup.go` -- **Propagate `canvasLineTop`'s JS-safe-bound error exactly as the paint loop does at `:598-601`**, for lines the paint loop may never reach. -- Rationale: the count must iterate the full `lines` list to stay independent of truncation; the error is the file's established, consistent refusal, not a new policy.

**Execution — Part 7: the projection fields and their browser admission.**

- `folio-go/page_setup.go` -- **Add `ContentWindowHeight int64 \`json:"contentWindowHeight"\`` and `ContentWindowCount int64 \`json:"contentWindowCount"\`` to `CanvasProjection` (`:265-294`)**, no `omitempty`, with comments in the register of `FontFamilies`/`DefaultFontSize` (`:279-292`): why the **engine**, not the browser, owns each number; **which coordinate frame** it is in; and — for the count — that it describes **the column as the canvas paints it**, that it is a **FLOOR for any bound table** (`projectedSize` gives header height only), and that it is **never derived from `CanvasTextPaint`**. -- Rationale: R2, and D-7.4.4 requires the comment to say what the number is a number *about*.
- `folio-designer/src/engine-protocol.ts` -- **Add both keys to the `hasOnly` list at `:156`, to the `CanvasProjection` type at `:95-104`, and to the strictly-positive array at `:158`.** -- Rationale: `hasOnly` is exact-key, so a Go-only field addition silently discards the entire snapshot and blanks the canvas.
- `folio-designer/src/engine-protocol.test.ts:4`, `folio-designer/src/App.test.tsx:17`, `folio-designer/src/DataPanel.test.tsx:18` -- **Add both fields to the three base fixture literals** (all other fixtures are spreads and inherit them). -- Rationale: verified — these are the only three non-spread projection literals in the designer.
- `folio-designer/src/canvas-authority-contract.test.ts` -- **Keep the two prohibited regexes (`:38-39`), their non-vacuity (`:52-55`) and their red-proofs (`:70-71`) intact**, and confirm the new field does not tempt a browser-side derivation. ⚠ **Do not rename `isTextPaint`'s `value` parameter** — spelling it `paint` would make regex `:38` fire on the legitimate bound check at `:291`. -- Rationale: D-7.4.2 §5; the regexes' own rationale comment names a **window count** explicitly.

**Execution — Part 8: the DISCRIMINATING fixture (R2).**

- `folio-go/canvas_window_count_template.go` (new) -- **Add the first multi-page CANVAS fixture**, as a template constant in the shape of `multi_page_template.go` / `page_count_matrix_templates.go`. Geometry: A4 portrait, margins `{top:30, right:54, bottom:42, left:36}`, header 18, footer 24 ⇒ **content window `H` = 727890 mp** (the `page-count-*` geometry, already documented and cross-checked). Content band: a text element at `y:0`, and **a second element declared far below it at `y:7280pt` — ten windows down, with nothing between.** **Assert the reported count is exactly 2.** -- Rationale: R2. `ceil(lowestBottom / H)` gives **11** here, so the fixture red-proves the exact spelling `paginate.go:69-74` bans. ⚠ **The existing `page-count-*` fixtures CANNOT discriminate it** — their one-window spacing makes the closed form and the true slide count coincide at every N.
- `folio-go/canvas_window_count_template.go` (new) -- **Add a negative control constant in the same file**: elements at `y:0, 728, 1456` ⇒ **3 by both routes**. -- Rationale: without it the discriminating test could pass by always answering 2, and the control records *why* the closed form survived this long.
- `folio-go/canvas_window_count_test.go` (new) -- **Assert the count is identical for a truncating and a non-truncating paint of the same document**, and that it **agrees with the render-path oracle** (`documentBands` → `collectBandTextRuns` → `contentColumnItems` → `layout.Paginate`) on a text-only fixture with no placeholders. -- Rationale: D-7.4.2 §5 re-asserted from this side; and the render oracle at `canvas_body_text_bounds_test.go:242-258` guards a **different** path, so agreement must be measured, not assumed.
- `folio-go/canvas_window_count_test.go` (new) -- **Assert an over-tall content component degrades to one window and does NOT fail the projection**, and that a bound table's count is a floor. -- Rationale: Ruling C, and R2's instruction to record the floor explicitly rather than let 7.6 discover it.
- `folio-go/byte_neutrality_test.go` -- ⚠ **a canvas-only fixture needs NO `goldenDigestRecord` entry.** If it also gains an `expected.pdf`, the entry and **every** site (`expected.json`, second literal, any README quoting the digest) must be declared or `byte_neutrality_test.go`'s completeness half fails. -- Rationale: `:225`'s "invalidated IN WHOLE" clause.

**Acceptance Criteria:**

- Given a placement or bounds command in the **content** band with a Y beyond one page's content height, when the engine validates it, then it is **accepted** and the element persists at that Y in the canonical bytes.
- Given the same command in the **pageHeader** or **pageFooter** band, when it is validated, then it is refused, and the test asserts the **exact message text** `folio: component geometry must stay within pageHeader` / `... pageFooter` by full-string equality.
- Given a negative coordinate in any band, or a coordinate beyond the JavaScript-safe geometry bound, when it is validated, then it is still refused and the test asserts the **exact message text**, unchanged from `83d49b6` — and those assertions were **landed and red-proved before** the split, not after.
- Given a content-band component whose box extends past one window, when the projection reaches the browser, then the snapshot is **admitted**, not silently discarded; and given the same component in a repeating band, then it is still rejected.
- Given a template whose content column is longer than one page, when the canvas is projected, then the projection reports the page-height window and how many windows the column occupies — **as the canvas currently paints it, not as a prediction of the rendered document** (R2). Given a content band containing a bound table, then the reported count is a **floor**, and the projection's own comment says so.
- Given a content element declared ten windows below the text with nothing between, when the count is reported, then it is **2**, not 11 — and given a control template with elements one window apart, then it is **3**.
- Given the same document projected with and without paint truncation, when the reported window count is compared, then it is identical; and given the designer sources, when the authority contract scans them, then no height or window count is derived from a paint's line count.
- Given a content component taller than one window, or any other input for which `layout.Paginate` returns an error, when the canvas is projected, then the projection **succeeds** and reports one window; it never fails and never blanks the canvas.
- Given the Go band-containment split, when the commit is inspected, then `folio-designer/src/engine-protocol.ts`'s mirror moved **in that same commit**, with a test that reads both sides, and the story record **enumerates the mirrors that exist in `engine-protocol.ts` today**.
- Given every existing template, none of which places anything past page one, when it is projected and rendered, then every pre-existing projection field value and all **twenty** golden digests are **measured** unchanged.
- Given the diff, when it is inspected, then `internal/layout/paginate.go` is **absent** from it, and the repository-root `README.md` is byte-identical to its committed state and appears in no commit.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 0, medium 2, low 3)
- defer: 2: (high 0, medium 1, low 1)
- reject: 14: (high 0, medium 3, low 11)
- addressed_findings:
  - `[medium]` `[patch]` **`containEdgeY` was unverified at three of its four call sites.** Reverting `dropComponent`'s Y (`:1411`) and `setComponentBounds`' height and Y (`:1670`, `:1672`) to the unconditional `containEdge` left the ENTIRE Go suite green — measured, not assumed. Added `TestTheColumnLiftIsExercisedAtTheCommandSurface`, which drives a snapped `dropComponent` at the foot of the content band and a snapped `setComponentBounds` three windows down. Red-proved against both mutants separately: the drop mutant pulls the box back to `y=654000` inside the one-window band, and the bounds mutant commits `height = 0`, because `containEdge(height, band.Height-y)` sees a negative limit, `floorToGrid` returns 0, and the positivity guard ran long before.
  - `[medium]` `[patch]` **`containEdgeY`'s CLAMPING branch was unexercised in every band.** No test snapped a component in `pageHeader` or `pageFooter`, so a `containEdgeY` rewritten to `return value` unconditionally stayed green at all four sites. Added a `pageHeader` edge move with an off-grid far edge (height 25pt against a 60pt band, so `edgeY = 35000` is not a grid multiple and snapping rounds it past the band foot). Red-proved: with the clamp removed the command is refused outright rather than rescued.
  - `[low]` `[patch]` **AC1/AC2/AC3 were asserted at an internal proxy, not the surface the matrix names.** `TestBandContainmentRefusalsCarryTheirExactMessages` calls `containComponent` directly, and only the `pageHeader` property path asserted the `*ComponentCommandError` wrapping an author actually sees. Matrix row 1 also names `createComponent`, which no test exercised with a tall content Y. Added command-surface full-string message assertions for `content` and `pageFooter` (with `DataPath` and byte-immutability), and a `createComponent` four windows down that round-trips through the canonical bytes.
  - `[low]` `[patch]` **`addCanvasTextPaint`'s own band table still spelled the three band names inline** (`page_setup.go:679-681`) — inside the very function that now compares `band.name == bandContent`, which is exactly the fifth-inline-spelling risk the constants' own comment names. Converted to `bandPageHeader`/`bandContent`/`bandPageFooter`; `page_setup.go` now carries zero inline band-name literals.
  - `[low]` `[patch]` **Two comments claimed more than the code does.** `addCanvasWindowCount`'s "EVERY CONTENT COMPONENT CONTRIBUTES, STYLED OR NOT" omitted Ruling D's stated exception — a text element that shapes no lines contributes nothing, so an empty-valued text box windows down is one more reason the count is a floor. And the new TS test attributed the omitted-field rejection to `hasOnly`, which is a SUBSET check (`hasExactKeys` is the strict sibling); the rejection actually comes from `integer(key, true)` reading `undefined`. The same test exercised only a missing `contentWindowCount`. Corrected both comments and extended the test to cover a missing `contentWindowHeight` and five malformed values per field.

**Rejected, with the reason each was tested against (not silently dropped):** swallowing `*MixedItemError` alongside `*OverflowError` — Part 6 says "on any `layout.Paginate` error"; `ENGINE_PROTOCOL_VERSION` not bumped for two added keys — `791ed00` added `fontFamilies`/`defaultFontSize` without bumping it and the wasm is rebuilt by `build:wasm` on every typecheck and test run; `BANDS_CAPPING_VERTICALLY` not `as const` — the mirror test's own regex terminates at `]`, so hardening it is churn against the tie; positional `Bands[1]` indexing in the new tests — the Code Map itself identifies `bands[1]` as the content band; the mirror test's source-formatting regexes — that is the in-tree idiom the six numeral pairs already use; `layout.PlaceInBand` instead of the inline `+= layout.Origins(g).Content` — Part 6 prescribed that spelling, and AD-24 governs the top-left → PDF-bottom-left flip, which this translation is not; `visibleIf` making the count an over-estimate — render-time visibility is outside what the canvas knows, the same family as the bound-table floor R2 already records; a corpus-wide projection enumeration — AC10's measurement is the twenty golden digests plus the pre-existing projection suite, and that suite is untouched in this diff (zero deleted lines in any pre-existing projection test file) and green; `Canvas()`'s hard-coded count of 1 — Ruling B; `addCanvasWindowCount` re-deriving `canvasPageGeometry` — a dead error branch, cosmetic; the `*[]layout.ColumnItem` out-parameter seam — cosmetic; no image/line window-count fixtures — the non-text branch is uniform through `projectedSize`; no `testing.Short()` guard on the truncation fixture; the fourth fixture inlined in the empty-column test.


## Design Notes

### The window-count gap — CLOSED by R1 (the contract's `Block If` no longer fires)

The contract's `Block If` points here: *"The window count's derivation is not settled by the intent."*
**It is now settled.** R1 selects a derivation that is neither of the two the first dispatch could not
choose between: **the real `layout.Paginate`, fed `ColumnItem`s built from extents the canvas paint
plan has ALREADY computed.** Option (A)'s bind coupling is rejected on three independent grounds and
is Story 13.4's in another epic; option (C)'s closed form is foreclosed by name at `paginate.go:69-74`;
deferring the count to 7.6 is rejected because AC4 exists in 7.5 *so that* 7.6 measures nothing.

**Why the previously-cited prohibition does not bite.** `paginate.go:98-104` forbids a **second
derivation of an item's EXTENT**, not a second call to `Paginate` — and `Paginate`'s own signature
(`PageGeometry` + `[]ColumnItem`, no data, no bindings, no template) shows it is designed to receive
caller-derived extents. Reusing the per-line tops and advances the paint plan computes once is
therefore the safe spelling: **one shaping, two consumers** — the paint plan and the window count.
A fresh shaping pass would be the violation, and is forbidden.

This subsection exists so the contract's pointer still resolves. **Do not re-open the question.**

### The eleven-call-site band audit (R3)

⚠ **There are eleven, not twelve** — the twelfth grep hit is the definition at `:1766`. ⚠ **No site pins a band**: `findComponent` `:1517`, `bandByName` `:1295-1299` and `hitTestBand` `:1505-1509` all range over the three names, so **every site can receive any band** and the split must key on `band.Name` inside `containComponent`.

| # | Line | Function | Kind | Relies on the vertical cap? |
|---|------|----------|------|------------------------------|
| 1 | `:165` | `addTableColumn` | refusal (`column.width`) | **NO** — mutates `width` only; `Y`/`HeaderHeight` untouched |
| 2 | `:300` | `updateTableColumn` | refusal (`column.width`) | **NO** — editable fields are `header`/`width`/`align` |
| 3 | `:829` | `updateComponentPropertiesInPlace` | refusal (`component.geometry`) | **YES — highest stakes.** `applyPropertyChanges` bounds neither `x` nor `y`; `:829` is the **sole** band bound for the whole property surface |
| 4 | `:1409` | `dropComponent` | **probe**, gates `containEdge` `:1410-1411` | **YES** — `hitTestBand` keeps `y ∈ [0,H)`, but `height > H-y` fires near the bottom. Lifting makes the probe pass and **engages a y snap-back that does not run today** |
| 5 | `:1472` | `createComponentInBand` | refusal | **YES** — primary create/drop gate; only `width<=0 \|\| height<=0` upstream |
| 6 | `:1565` | `moveComponent` | **probe**, gates `containEdge` `:1566-1567` | **YES** — lifting engages the y pull-back and a drag that is refused today silently succeeds at a clamped y |
| 7 | `:1569` | `moveComponent` | refusal | **YES** — the only check stopping a move below the band bottom |
| 8 | `:1606` | `resizeComponent` | refusal | **YES** — `height > H-y` is the sole ceiling on a resize's height |
| 9 | `:1668` | `setComponentBounds` | **probe**, gates four `containEdge` `:1669-1672` | **YES** — two of the four pull-backs are vertical |
| 10 | `:1674` | `setComponentBounds` | refusal | **YES** — only vertical ceiling on the combined rectangle |
| 11 | `:1726` | `duplicateComponent` | **probe, silent fallback** (`x,y = clone.X, clone.Y`; no error ever) | **YES** — the `+6000` offset near the band bottom is exactly what triggers the fallback today; lifting keeps the offset and lands the duplicate outside the band |

**The finding R3 asked for:** sites 4, 6, 9 and 11 are the ones that **change behaviour silently** — they are probes whose result is discarded, so lifting the cap widens a gate rather than removing a refusal. Sites 4/6/9 begin engaging pull-backs that today never run; site 11 stops falling back. **The pre-clamps are not unconditional**, which is the correction to the previous dispatch's reading that they would make a lifted validator "accept nothing new".

### The `engine-protocol.ts` mirror inventory (R4) — the artifact

**6 of ~35 duplicated Go/TS invariants are tied.** DW-25 closed **Group A only**.

**Group A — TIED by `engine-bounds-mirror.test.ts` (the six `pairs`, `:47-54`):**

| TS | Go | Value |
|---|---|---|
| `MAX_CANVAS_BODY_TEXT` `:35` | `page_setup.go:70` | 1048576 |
| `MAX_CANVAS_BODY_TEXT_LINES` `:37` | `page_setup.go:82` | 1920 |
| `MAX_CANVAS_BODY_TEXT_FRAGMENTS` `:42` | `page_setup.go:112` | 65536 |
| `MAX_CANVAS_PROPERTY_STRING` `:45` | `page_setup.go:38` | 512 |
| `MIN_LINE_SPACING_THOUSANDTHS` `:56` | `internal/template/linespacing.go:53` | 1 |
| `MAX_LINE_SPACING_THOUSANDTHS` `:57` | `internal/template/linespacing.go:54` | 1000000 |

**Group B — numeric literals duplicated from Go, UNTIED:** `MAX_ENGINE_FONT_FAMILIES` `:15` ↔ `page_setup.go:298` (256); ⚠ **a bare inline `512` at `:162`** for font-family name length ↔ `page_setup.go:310` — **invisible to the mirror test, whose site regex only matches `boundedString(key, MAX_CANVAS_PROPERTY_STRING)`**; `MAX_ENGINE_BINDING_LENGTH` `:11` ↔ `page_setup.go:263` (256); `MAX_ENGINE_PAYLOAD_BYTES` `:6` ↔ `component_commands.go:622` (8 MiB — **whose Go comment declares itself a mirror**); `MAX_ENGINE_ELEMENT_ID_LENGTH` `:9` (relied on by name at `component_commands.go:626-627`); `MAX_ENGINE_PARAMETER_NAME_LENGTH` `:12-13` ↔ `parameter_references.go:16`; inline `128` table-column cap `:153` ↔ `table_columns_projection.go:10`; asset-key shape `:256` ↔ `asset_bytes.go:41-50`; SHA-256 hex shape `:148`; table field caps `:153`.

**Group C — mirrored PREDICATES, all untied:** the JS-safe bound (`Number.isSafeInteger` at `:157/169/184/203/221/257/300/307` ↔ `MaxCanvasMillipoints` — **structurally untieable as written, and that spelling is the safer one**); paint-inside-page `:171`; band contiguity `:172-175`; band count and order `:165-169`; ⚠ **component band containment `:193` ↔ `containComponent` — THIS STORY'S mirror**; image draw containment `:260`; `resizable ⇔ type≠table` `:192`; component band ordering `:182-188`; component id uniqueness `:181-185`; fontFamilies sortedness `:162`; text-paint line invariants `:302`; fragment x containment `:307`; the `align`/`valign` `:213`, `borderEdges` `:214`, `preset`/`orientation` `:156`, component `type` `:178`, band-name `:179`, `imageUnavailable` `:235`, table-column `align`/`footer` `:153` vocabularies; field/type coupling `:222-229`; `lineSpacing` range `:221` (literals tied, predicate untied); parameter-name shape/sort/uniqueness `:149`.

⚠ **The asymmetry that makes `:193` dangerous:** Go enforces containment **only on the command path**. `canvasComponents` (`page_setup.go:716-776`) does not re-check it when projecting, and `parse_bands.go` has none either. **So Go already projects out-of-band elements happily and TypeScript alone kills them** — which is why the lift is a TS-side necessity, not a courtesy.

⚠ **Highest-risk untied item, for the record:** the bare `512` at `:162`. A DW-25-style raise updates `:45` and leaves `:162` behind, every test stays green, and documents with long family names silently blank the canvas — the same defect class DW-25 was created to close, still present in the file DW-25 hardened. **Not this story's to fix; recorded so the inventory is honest.**

### Ruling A — the canvas builds its own `PageGeometry`; it must not call `pageGeometryOf`

`pageGeometryOf` (`render.go:291`) → `pageDimensions` (`render.go:75-91`) **hard-errors on `"Letter"`**, by deliberate design (its comment: *"failing loudly here is more honest than a silent A4 substitution"*). `canvasDimensions` (`page_setup.go:908-909`) **supports Letter**. A Letter document projects a canvas today. **Selected by the contract's own corpus/projection-neutrality rule** — the count must not be able to break a projection that works today. Constructing `layout.PageGeometry` in `Canvas` from the values already in scope at `:332-341` also serves R5: the same `g` yields the band height **and** the reported window height, so they cannot diverge. This is not a third spelling — it is the one construction both consumers read.

### Ruling B — the count is a paint-plan product, computed in `CanvasWithTextPaint`

R1 sources the extents from the paint plan, which exists only in `addCanvasTextPaint`, which only `CanvasWithTextPaint` calls. **Selected by an existing documented rule, not by builder preference:** `page_setup.go:388-392` already states that every mutating command's `Canvas(t)` is discarded and recomputed by `wasm/engine.go`, *"so the paint must be derivable from template state alone, exactly like text paint."* Verified at the seam: `wasm/engine.go:119`, `:255`, `:294` are the only three producers reaching the browser, and **all three call `CanvasWithTextPaint`**. `Canvas` therefore reports the documented floor of one window and never reaches the app; its comment must say so rather than leave a silent zero.

### Ruling C — a pagination failure degrades the count; it never fails the projection

`Paginate` returns `*OverflowError` (`paginate.go:1075-1083`) for an **ungrouped** item taller than one window — precisely the component this story newly makes authorable. Three settled positions converge, so this is a rule and not a coin flip: **(i)** Tasks Part 2 forbids turning the render path's overflow into a canvas refusal (*"a canvas/command bound must never become a document validity rule"*); **(ii)** R2 already establishes the count as a floor and requires floors to be recorded rather than hidden; **(iii)** the paint loop's existing idiom (`:526-534`) keeps the whole projection alive when a font chain fails. Reporting **one** window matches `Paginate`'s own documented answer for a column it cannot place (`:545-557`). **Consequence for 7.6, flagged deliberately:** such a document draws one sheet. That is a floor, not a lie — and it is strictly better than the alternative the settled positions exclude, which is a blank canvas with no attributable error.

### Ruling D — every content-band component contributes an extent, styled or not

The render path's `collectElementBoxRects` (`element_box.go:71-75`) **skips elements with no background or border** — correct there, because an unstyled rect draws nothing. **R2's own wording selects differently for the canvas:** the count is *"a claim about the column as the canvas currently paints it"*, and the canvas paints every component's box. R2's fixture requirement says *"an element declared far below the text"*, not a styled one. So: **text contributes one item per shaped line** (R1's named mechanism, never its box); **every other content-band component contributes one item from the box `projectedSize` already gives it** — for a table, header height only, which is R2's recorded floor. An element that shapes to zero lines, or whose value is empty or whose font chain fails, contributes nothing, matching the render path's treatment of a value that binds to empty (`render.go:775-782`).

### DW-33 — flagged, and provably not touched

DW-33 is `page_setup.go:658-661`: a text element whose first line already exceeds the per-line guard paints zero lines. ✅ **The window count reads `lines`, `originY` and `vm` — never `painted`, `budget`, `oversized` or `placed` — so its answer is identical whether the element paints zero lines or all of them.** The contract's Block If asks for a flag if the work touches that path; it does not, and this is the flag. **No ruling on whether a partial prefix should paint is made or needed here.**

### Judgments carried forward from the previous gate (unchanged, non-blocking)

1. **After the lift, the content band's Y is bounded by the JavaScript-safe geometry bound and nothing else.** AC1 and AC3 select it; that is the ACs speaking, not a builder choosing.
2. **AC4 vs AC5 is an apparent, not a real, contradiction** — "no pre-existing field's value moves, and the window count reports one window", the same in-epic reading D-7.1.3 applied. `fixtures/page-count-*` are outside AC5's premise: they do place content past page one.
3. **The live-drag clamp (`resize-anchor.ts:29` / `App.tsx:701`) stays.** Dragging onto a later sheet is Story 7.6's own AC; there are no later sheets until 7.6 draws them.
4. **An over-tall content element is not newly refused.** See Tasks Part 2 and Ruling C.
5. **No `.folio` format field is required.** This lifts a validation bound and adds a projection field; the projection is not the file format. `SupportedMajor` stays **2**.

### Limits to state, not to fix

- **D-7.4.4 stands.** The canvas breaks the **raw** template string with `nil` substitutions (`page_setup.go:544`, `:548-549`); it does not know where the engine will break a **bound** value. The new fields' comments must say what their numbers are numbers *about*. Reuse of the render/bind machinery is **rejected** by R1 on three independent grounds, and Story 13.4 in another epic is that change.
- **The bound-table floor is pre-existing (Epic 6) and irreparable within 7.5.** `projectedSize` returns header height only because the canvas has no data.
- ⚠ **An element at `y:7280pt` is storable but not droppable-at**: `hitTestBand`'s `:1501` half-open rectangle is one page tall and does not move in this story. That asymmetry is Story 7.6's to resolve when it draws the later sheets.

## Verification

7.5 changes a command-validation bound and a channel schema on both sides, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). **Report measured pass/fail counts, never "green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:169`, P6g subtest, floor at `:184`, `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` (`:243`) must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean.
- `gofmt -l folio-go` -- run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- run **once per leg** with `FOLIO_MATRIX_TARGET` set: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` (`matrix_test.go:69-74`). **Unset, this test logs "asserts NOTHING" and returns — a no-op is not a pass.** Grep each leg's output for "asserts NOTHING" and report the count. Name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; it exercises all four targets from one process and is not gated on the env var.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass, **239 tests / 32 files at baseline** plus whatever this story adds. **oxlint baseline is exactly 4 `only-export-components` warnings and 0 errors** (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1155,1162`); **a fifth is a regression** — keep new helpers in a `.ts` file.
- `cd folio-designer && npm run test:e2e:compile` -- expected: pass. Browser e2e is deferred by D-000.4 and does not execute; do not claim it ran.

**Nine digests to report byte-identical** (all **twenty** in `goldenDigestRecord` must hold): statement-1 76,744 `114df1d6…`; statement-5 127,363 `70dce051…`; statement-20 269,884 `56bfbbd9…`; statement-50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`. **Corpus and projection both unchanged for existing templates — assert, do not assume.**

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` (`fontgen_matrix_test.go:64`) fails under `-tags=matrix` without `fontTools`; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, Story 15.2).

**Manual checks:**
- **Re-derive the `containComponent` call-site enumeration by grep at the closing revision** and confirm **eleven** sites still route through the split validator.
- **Confirm `internal/layout/paginate.go` is absent from the diff** (`git diff --stat`), and that `README.md` appears in no commit.
- **Demonstrate end to end** that a component placed windows below the content top survives a command, a projection, and a round trip through the canonical bytes — not that a conditional changed.
- **Confirm the Go split and `engine-protocol.ts:193` landed in the SAME commit** (`git show --stat`).
- **Confirm `git status` is empty for `fixtures/`** before quoting any digest.

## Auto Run Result

### Dispatch 1 — 2026-08-31, plan only

Status: `blocked`
Blocking condition: `intent gap`
Baseline: `23a46470be4078dad62d625a64806a6adf3625a2` on `main`, tree clean
Directive: `Halt after planning.` — **no implementation code, no commits.**

**⚠ ORCHESTRATOR: when you amend `<intent-contract>` with the lead's ruling, set `status:` back to `draft`.** Step-01 of the next dispatch HALTs on `blocked spec supplied` otherwise.

**The gap, in one sentence.** AC4 requires the projection to report how many page-height windows the content column occupies; the cheap derivation is forbidden by name in `internal/layout/paginate.go:69-74`, and the two remaining derivations produce different integers for the same template with nothing in the intent selecting between them.

**What the lead must rule:** derivation **(A)** reuse the render/bind machinery with the designer's real data, **(B)** build column items from the canvas's own raw-string shaping, or **(D)** report the window height in 7.5 and defer the count to 7.6. Recommendation: the gap is genuine; (D) is the only option that does not settle D-7.4.4's reserved question a story early, but it does not satisfy AC4 as written.

**Five judgments made at this gate and NOT blocking** — the surviving content-band upper bound is the JavaScript-safe geometry bound; AC4/AC5's apparent contradiction resolves the way D-7.1.3 resolved the epic header's; `engine-protocol.ts:193` lifts in the same commit; the live-drag clamp stays for 7.6; an over-tall content element gains no new refusal.

Verification was **not** run in this dispatch: no code changed.

### Dispatch 2 — 2026-08-31, plan only

Status: `ready-for-dev`
Blocking condition: none
Baseline: `83d49b689bf3bd8b93d9da0c41f15a3be3e32f3e` on `main`, tree clean
Directive: `Halt after planning.` — **no implementation code, no commits, and none were made.**

**The intent gap is closed.** R1–R5 were supplied as settled rulings and are recorded verbatim above the Code Map. The contract's `Block If` — *"The window count's derivation is not settled by the intent"* — is **discharged by R1** and no longer fires. `<intent-contract>` was carried forward **byte-for-byte unchanged**; every amendment landed in the Code Map, Tasks & Acceptance, and Design Notes.

**Anchor re-verification.** `git log 23a4647..83d49b6` is two commits, **both `_bmad-output/` only** — no Go and no TypeScript moved. The prior Code Map's anchors therefore held, with three corrections found by re-measurement:
1. ⚠ **Eleven `containComponent` call sites, not twelve** — the twelfth grep hit is the definition.
2. ⚠ **The `containEdge` pre-clamps are conditional, not unconditional.** Each is gated on the probe succeeding, so lifting the cap **widens** the gate rather than being neutralised by it — the opposite of the prior reading. Four sites (`:1409`, `:1565`, `:1668`, `:1726`) change behaviour **silently**.
3. ⚠ **No call site pins a band literal** — all eleven can receive any of the three, so the split must key on `band.Name` inside `containComponent`.

**Four forks surfaced by applying the rulings, all RULED here — none is a new intent gap.** Each was tested against a principle already settled in this spec, and each was selected rather than chosen: **Ruling A** (the canvas must not call `pageGeometryOf`, which hard-errors on `Letter` where `canvasDimensions` succeeds — selected by corpus/projection neutrality); **Ruling B** (the count is a `CanvasWithTextPaint` product — selected by the documented rule at `page_setup.go:388-392`, verified at all three `wasm/engine.go` seams); **Ruling C** (a `Paginate` error degrades to one window and never fails the projection — selected by settled Tasks Part 2, R2's floor doctrine and the paint loop's existing degradation idiom); **Ruling D** (every content-band component contributes an extent, styled or not — selected by R2's own "as the canvas currently paints it" wording).

**DW-33 is flagged and provably untouched**: the count reads `lines`, `originY` and `vm`, never `painted`/`budget`/`oversized`, so its answer is identical whether the element paints zero lines or all of them. No ruling on the partial prefix is made.

**No format field is required.** `SupportedMajor` stays 2.

Verification was **not** run in this dispatch: no code changed. The `## Verification` section states what the implementing dispatch must measure.

### Dispatch 3 — 2026-08-31, implement, review and commit

Status: `done`
Blocking condition: none
Baseline: `4bf201abadeafe29f7a9c19efb6f50b302a1b8b0` on `main`, tree clean. Only `_bmad-output/` moved between the Code Map's `83d49b6` and this baseline, so every anchor held.
Directive: implement — no halt-after-planning.

**What was implemented.** The content band's vertical cap is gone and only there. `containComponent` is split by MEANING inside the function, keyed on `band.Name` through a named `bandsCappingVertically = {pageHeader, pageFooter}`: representational refusals (negative coordinates) and the horizontal caps stay universal, band-capacity refusals apply to the two repeating bands alone, and every surviving refusal keeps `"folio: component geometry must stay within %s"` to the byte. `containEdgeY` makes the four vertical pull-backs no-ops in the content band — necessary, not cosmetic, because the pre-clamps are GATED on the probe succeeding, so lifting the cap widens the gate rather than neutralising it. The browser's copy of the same gate lifted in the same commit under a name both files declare. The projection gained `contentWindowHeight` and `contentWindowCount`, the count coming from the real `layout.Paginate` fed the extents the canvas paint plan already computed — one shaping, two consumers. `page_setup.go`'s second `ContentHeight` spelling collapsed into `layout.ContentHeight` (R5).

**Files changed.**
- `folio-go/component_commands.go` — the split, `containEdgeY`, band-name constants.
- `folio-go/page_setup.go` — `canvasPageGeometry`, the R5 collapse, the two projection fields, the per-line extent emission in `addCanvasTextPaint`, `addCanvasWindowCount` as the third paint producer.
- `folio-go/canvas_window_count_template.go` (new) — the first multi-page CANVAS fixtures: the ten-window gap, the one-window control, an over-tall component, a bound table.
- `folio-go/canvas_window_count_test.go` (new) — the discriminating count, truncation independence, render-oracle agreement, Ruling C degradation, the bound-table floor, the empty column.
- `folio-go/component_commands_test.go` — the Part 1 characterization, the inverted snap test, and (review pass) `TestTheColumnLiftIsExercisedAtTheCommandSurface`.
- `folio-go/component_properties_test.go` — the `:829` property-edit surface, which had no band-bounds coverage at all.
- `folio-designer/src/engine-protocol.ts` — `BANDS_CAPPING_VERTICALLY`, the vertical lift, the two fields in the type, the exact-key allow-list and the strictly-positive array.
- `folio-designer/src/engine-bounds-mirror.test.ts` — the widened obligation and the first PREDICATE tie, reading both `component_commands.go` and `engine-protocol.ts`.
- `folio-designer/src/engine-protocol.test.ts`, `App.test.tsx`, `DataPanel.test.tsx` — the admission cases and the three base fixture literals.

**Review findings breakdown.** 4 layers ran (blind hunter, edge-case hunter, verification-gap, intent-alignment). intent_gap 0, bad_spec 0, **patch 5 applied** (2 medium, 3 low), **defer 2**, reject 14. Every patch is a verification or accuracy fix; no production behaviour changed in the review pass except two comments and one band-literal substitution. Details in the Review Triage Log.

**Follow-up review recommendation: `true`.** Patched this pass: high 0, medium 2, low 3 → `3x2 + 1x3 = 9`, which is >= 5. Nothing high was patched; the score is carried by the two mutation-proved verification gaps.

**Verification performed — measured, and re-run in full after the patches.**
- `go test -count=1 ./...` — **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)`, `got 7, need >=20`, the mandated permanent red. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` ran and **passed**. Every other package `ok`.
- `go vet -tags=matrix ./...` — clean, exit 0. `gofmt -l folio-go` from the repo root — no output.
- `TestTargetRenderHash` once per leg with `FOLIO_MATRIX_TARGET` set: **darwin/arm64, linux/amd64, linux/arm64, js/wasm — all PASS, `grep -c "asserts NOTHING"` = 0 on every leg**, so no leg was a no-op. `TestCrossTargetByteIdentity` — `ok`.
- `cd lint && go test -count=1 ./...` — 4/4 packages `ok`.
- Designer: `typecheck` clean; `oxlint` **exactly 4 warnings, 0 errors**, the baseline `only-export-components` set (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1155,1162`) — no fifth, the new export is in a `.ts` file; `npm test` **248 passed / 32 files** (baseline 239/32); `test:e2e:compile` passes — browser e2e is deferred by D-000.4 and did **not** run.
- **Nine digests re-measured byte-identical after the patches**, with `git status fixtures/` empty: statement-1 76,744 `114df1d6`; -5 127,363 `70dce051`; -20 269,884 `56bfbbd9`; -50 555,829 `5d090b0f`; mandatory-break 56,681 `7cf743de`; line-spacing 57,770 `de212115`; justified-text 59,894 `6da3b12e`; alignment-rounding 61,346 `986400a1`; justified-thai 15,079 `58ca4777`. All twenty `goldenDigestRecord` entries hold via `TestGoldenDigestAgreesAtEveryDeclaredSite` and the per-fixture golden tests.
- **Manual checks.** `containComponent` re-grepped at the closing revision: **twelve hits, eleven call sites** (`:165`, `:300`, `:829`, `:1409`, `:1472`, `:1565`, `:1569`, `:1606`, `:1668`, `:1674`, `:1726`) plus the definition, matching the audit table exactly. `internal/layout/paginate.go` is **absent** from `git diff --name-only`. `README.md` appears in no commit and its md5 is `078d7d80d518d54af2fc04fb270d46b8`, unchanged. `git show --stat` confirms the Go split and `engine-protocol.ts` in the **same** commit.
- Known-environmental, not regressions: `TestShippedFacesReproduceFromUpstream` (no `fontTools`); `lint/internal/rules/licencegraph_test.go` gofmt (DW-23).

**Residual risks.**
- The count is a **FLOOR**, by construction and irreparably within 7.5: `projectedSize` gives a table its header height and none of its rows, and a text element that shapes no lines contributes nothing. Story 7.6 must not read it as a prediction.
- The two deferred items: the designer's live drag clamp (`resize-anchor.ts`) still stops an author at page one, and `contentWindowCount` has no upper bound.
- An element windows down is **storable but not droppable-at**: `hitTestBand`'s rectangle is still one page tall, deliberately, and that asymmetry is Story 7.6's.
- DW-33 flagged and **provably untouched** — the count reads `lines`, `originY` and `vm`, never `painted`, `budget` or `oversized`. No ruling on the partial prefix was made or needed.

## Delivery Log

### Dispatch 3 — 2026-08-31, implemented

Baseline: `4bf201abadeafe29f7a9c19efb6f50b302a1b8b0` on `main`, tree clean apart from this spec's own frontmatter.
Commits: `deccd2f` (Part 1, characterization) and `cd196f0` (Parts 2–8).

**Part 1 landed as its own commit, and was RED-PROVED before Part 2 touched anything.**
`folio-go/component_commands_test.go` gained `TestBandContainmentRefusalsCarryTheirExactMessages`
and `TestJavaScriptSafeGeometryBoundRefusalsCarryTheirExactMessages`, both by full-string
equality and never `strings.Contains`. Three separate perturbations were applied and reverted:

| Perturbation | What failed |
|---|---|
| `must stay within %s` → `must remain within %s` | `negative x in pageHeader = "folio: component geometry must remain within pageHeader", want exactly "folio: component geometry must stay within pageHeader"` |
| `folio: component.%s: %w` → `folio: component %s: %w` | `move past the JavaScript-safe bound = folio: component y: … , want exactly "folio: component.y: y exceeds the JavaScript-safe geometry bound"` |
| `page_setup.go`'s `JavaScript-safe` → `JS-safe` | `projection of an unrepresentable coordinate = folio: component exceeds the JS-safe geometry bound, want exactly "folio: component exceeds the JavaScript-safe geometry bound"` |

The content band's own vertical cap was asserted in that first commit **while it was still
true**, and the second commit changes those two lines into acceptance assertions — which is the
whole point of the ordering.

**The eleven-call-site band audit, RE-DERIVED BY GREP AT THE CLOSING REVISION.** `grep -n
"containComponent(" folio-go/component_commands.go` returns twelve hits; the twelfth is the
definition. The eleven call sites are at `:165`, `:300`, `:829`, `:1409`, `:1472`, `:1565`,
`:1569`, `:1606`, `:1668`, `:1674`, `:1726` — the same line numbers the Code Map recorded,
because everything this story added to that file sits below `:1766`. No site pins a band, so the
split keys on `band.Name` **inside** `containComponent`, exactly as R3 required.

Sites 4 (`:1409`), 6 (`:1565`) and 9 (`:1668`) are probes gating `containEdge` pull-backs, and
lifting the cap **widens** those gates. Their vertical pull-backs now route through a new
`containEdgeY(band, value, limit)`, which is a no-op in the bands that do not cap vertically —
without it a drag that is refused outright today would have started passing the probe and been
silently clamped to the foot of page one. Four sites moved: `:1411`, `:1567`, `:1670`, `:1672`.
Site 11 (`duplicateComponent`, `:1726`) stops falling back near the old band bottom, which is the
intended new behaviour: the duplicate lands one grid step down the column.

**The `engine-protocol.ts` mirror inventory (R4), as the story's artifact.** 6 of ~35 duplicated
Go/TS invariants were tied before this story; **7 are tied now**. DW-25 closed Group A only.

*Group A — numeral pairs, TIED by `engine-bounds-mirror.test.ts` (unchanged, still six):*
`MAX_CANVAS_BODY_TEXT` ↔ `page_setup.go` `maxCanvasBodyText` (1048576); `MAX_CANVAS_BODY_TEXT_LINES` ↔
`maxCanvasBodyTextLines` (1920); `MAX_CANVAS_BODY_TEXT_FRAGMENTS` ↔ `maxCanvasBodyTextFragments`
(65536); `MAX_CANVAS_PROPERTY_STRING` ↔ `maxCanvasPropertyString` (512);
`MIN_LINE_SPACING_THOUSANDTHS` / `MAX_LINE_SPACING_THOUSANDTHS` ↔ `internal/template/linespacing.go`
(1 / 1000000).

*Newly TIED by this story — the first PREDICATE tie:* `BANDS_CAPPING_VERTICALLY` ↔
`component_commands.go`'s `bandsCappingVertically`. The test resolves Go's named constants before
comparing, asserts the list is non-empty on both sides, asserts `content` is on neither, asserts
each side actually CONSUMES its list at the validator it governs, asserts the horizontal cap stays
un-guarded on both sides, and red-proofs a one-sided edit in each direction. `goSources` now reads
a third Go file; `pairs` is still six and `toHaveLength(6)` is unchanged, because this tie is a
predicate and contributes no numeral pair.

*Group B — numeric literals duplicated from Go, still UNTIED:* `MAX_ENGINE_FONT_FAMILIES` ↔
`page_setup.go`'s `maxCanvasFontFamilies` (256); ⚠ **a bare inline `512`** for font-family name
length ↔ `maxCanvasPropertyString`, invisible to the mirror test because its site regex only
matches `boundedString(key, MAX_CANVAS_PROPERTY_STRING)`; `MAX_ENGINE_BINDING_LENGTH` ↔
`maxCanvasBindingString` (256); `MAX_ENGINE_PAYLOAD_BYTES` ↔ `component_commands.go`'s 8 MiB, whose
Go comment declares itself a mirror; `MAX_ENGINE_ELEMENT_ID_LENGTH`; `MAX_ENGINE_PARAMETER_NAME_LENGTH`
↔ `parameter_references.go`; the inline `128` table-column cap ↔ `table_columns_projection.go`;
the asset-key shape ↔ `asset_bytes.go`; the SHA-256 hex shape; the table field caps.

*Group C — mirrored PREDICATES, all still untied except this story's:* the JS-safe bound
(`Number.isSafeInteger` ↔ `MaxCanvasMillipoints`, structurally untieable as written and that
spelling is the safer one); paint-inside-page; band contiguity; band count and order; image draw
containment; `resizable ⇔ type≠table`; component band ordering; component id uniqueness;
`fontFamilies` sortedness; text-paint line invariants; fragment x containment; the `align`/`valign`,
`borderEdges`, `preset`/`orientation`, component `type`, band-name, `imageUnavailable` and
table-column vocabularies; field/type coupling; the `lineSpacing` range predicate; parameter-name
shape, sort and uniqueness.

⚠ **Highest-risk untied item, recorded so the inventory is honest and NOT fixed here:** the bare
inline `512` for font-family name length. A DW-25-style raise updates the named constant and leaves
that literal behind, every test stays green, and documents with long family names silently blank the
canvas — the same defect class DW-25 was created to close, still present in the file DW-25 hardened.

**The fixture DISCRIMINATES, measured rather than argued.** `canvas_window_count_template.go`'s gap
fixture (text at `y:0`, an unstyled rect at `y:7280pt`, one window = 727890 mp) reports **2**.
Replacing the derivation with `ceil(lowestBottom / H)` locally turned the test red with the message
`contentWindowCount = 11, want 2 (the forbidden closed form answers 11 here)`, and the control
fixture (elements one window apart ⇒ **3** by both routes) stayed green under the same
perturbation — which is exactly why the `page-count-*` fixtures could not have caught this.

**Ruling C exercised, not assumed.** An over-tall content component (900pt in a 727.89pt column)
reaches `layout.Paginate`'s `*OverflowError` from the canvas for the first time; the projection
succeeds, reports one window, and keeps the component.

**DW-33: flagged and provably untouched.** The count reads `lines`, `originY` and `vm`, never
`painted`, `budget`, `oversized` or the placed runs. No ruling on the partial prefix is made.
**No `.folio` format field was required; `SupportedMajor` stays 2.** DW-26 … DW-35 remain open.

**Verification, measured.**

- `cd folio-go && go test -count=1 ./...` — **exactly ONE failure**, the mandated permanent red:
  `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)`, `got 7, need >=20`. Its drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` is green. Identical to the baseline run taken before any
  edit. All **20** `goldenDigestRecord` entries hold, so the nine reported digests —
  statement-1 `114df1d6…`, statement-5 `70dce051…`, statement-20 `56bfbbd9…`, statement-50
  `5d090b0f…`, mandatory-break `7cf743de…`, line-spacing `de212115…`, justified-text `6da3b12e…`,
  alignment-rounding `986400a1…`, justified-thai `58ca4777…` — are byte-identical.
  `git status fixtures/` was empty throughout.
- `cd folio-go && go vet -tags=matrix ./...` — clean. `gofmt -l folio-go` from the repo root — no output.
- `go test -tags=matrix -run TestTargetRenderHash -v .` — run once per leg with `FOLIO_MATRIX_TARGET`
  set: **darwin/arm64, linux/amd64, linux/arm64, js/wasm**. `grep -c "asserts NOTHING"` was **0** on
  every leg, so all four legs actually asserted. All four pass.
- `go test -tags=matrix -run TestCrossTargetByteIdentity .` — pass.
- `cd lint && go test ./...` — pass (4 packages).
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; oxlint
  **exactly 4 `only-export-components` warnings, 0 errors**, the baseline set (`preview/pdf-viewer.tsx:16,17`,
  `App.tsx:1155,1162`) — no fifth, because the new export is in a `.ts` file. **248 tests / 32 files,
  all passing**, against a 239/32 baseline: +9.
- `cd folio-designer && npm run test:e2e:compile` — pass. Browser e2e is deferred by D-000.4 and did
  not execute; no claim is made that it ran.
- `git diff HEAD -- folio-go/internal/layout/ README.md` is **empty**: `paginate.go` is absent from
  the diff and `README.md` appears in no commit. Nothing was staged with `git add -A` or `git add .`.
- The Go split (`component_commands.go`) and the browser mirror (`engine-protocol.ts`) are in **one
  commit**, `cd196f0`, confirmed by `git show --stat`.

### Review pass — 2026-08-31, the two mutation red-proofs the review added

The review found that the Part 2 task "make the `containEdge` pull-backs correct under the lift"
had landed correct but **unverified at three of its four sites**, and that `containEdgeY`'s
clamping branch was unverified at all four. Both were established by mutation, not by reading:

| Mutant | Suite before the patch | Suite after the patch |
|---|---|---|
| `dropComponent:1411` + `setComponentBounds:1670,:1672` → unconditional `containEdge` | **green** (only the mandated P6g red) | `dropped box (y 654000 + height 24000) was pulled back inside the one-window band height 679890` |
| `setComponentBounds:1670,:1672` alone → unconditional `containEdge` | **green** | `snapped bounds committed height = 0; the pull-back collapsed the component` |
| `containEdgeY` → `return value` in every band | **green** | `an edge move inside the pageHeader was refused; the pull-back did not rescue the grid's rounding` |

The `height = 0` case is the expensive one and is worth stating plainly: the positivity guard at
`component_commands.go:1665` runs **before** the pull-backs at `:1669-1672`, so
`containEdge(height, band.Height-y)` with a negative limit returns `floorToGrid(negative) = 0` and
the component is committed with no height at all. That is a silent data-loss path that this
story's own lift opens (by widening the probe gate) and that nothing would have caught.

`TestTheColumnLiftIsExercisedAtTheCommandSurface` also moves AC1/AC2/AC3 to the surface the I/O
matrix names. The characterization tests call `containComponent` directly — correct for
characterizing a predicate, but the matrix rows are phrased as commands, and the refusal an author
sees is a `*ComponentCommandError` whose `Message` and `DataPath` the eleven call sites set. The
new test asserts those, by full-string equality, for `content` and `pageFooter`; adds the
`createComponent` case matrix row 1 names by name and nothing exercised; and round-trips each
accepted placement through the canonical bytes.

### 2026-08-31 — done

Baseline `4bf201a`, on `main`, closed at `2b173a1` plus this closing commit. **Planned** across three
dispatches, **built** in four commits, **done** here — with every gate below re-measured at the closing
revision rather than carried forward from the build's report.

**The story halted, and five rulings unblocked it.** The first dispatch stopped on a real intent gap: AC4
asks how many page-height windows the content column occupies, the cheap derivation is forbidden **by
name** in `internal/layout/paginate.go:69-74`, and the two survivors answer different integers for the
same template with nothing in the intent choosing between them. The engineering lead ruled — **R1** the
count comes from the real `layout.Paginate`, fed extents the canvas paint plan has already computed (one
shaping, two consumers; the prohibition is against a second derivation of *extents*, not of pagination);
**R2** the count is a claim about the canvas, not a prediction of the render, and its fixture must
*discriminate* rather than merely exercise; **R3** split the containment check by what its clauses *mean*,
and characterize before changing; **R4** the TypeScript mirror lifts in the same commit, and the standing
obligation widens from "the size caps move together" to "any invariant duplicated across the boundary
moves in one commit, with a test that reads both sides"; **R5** collapse `page_setup.go`'s second
`ContentHeight` spelling here, because AC4 makes it load-bearing rather than cosmetic.

**Three of the four commits were created by the implementation subagent unprompted.** The dispatch was
asked for the story; it produced `deccd2f` (characterization), `cd196f0` (the split and everything after
it) and `2b173a1` (the review patches) on its own initiative, plus `eae8863` for the mirror inventory.
That is more history than a single dispatch normally leaves, and it is recorded because the closing audit
had to verify four commits, not one — including that the characterization genuinely preceded the change.

**The ordering held, and it was checked rather than believed.** `deccd2f` touches
`folio-go/component_commands_test.go` **and nothing else** — no production file — and it asserts, by
full-string equality, that a content Y past the band height *is refused*. `cd196f0` is the commit that
deletes those two assertions and replaces them with acceptance. Had the order been reversed the
assertions would have described whatever the split produced. Full-string equality is load-bearing here
and is used everywhere: an unrelated `column.footerOf` refusal at `component_commands.go:460` also
contains "must stay within" and is exercised by a live test, so `strings.Contains` would have been
satisfied by a message about table collections.

**The fixture discriminates 2 from 11, reproduced at this revision.** The gap fixture declares one line of
text at the top of the column and an unstyled rect **ten windows below it**, with nothing between; the
window is 727890 mp. Replacing the derivation with the banned closed form `ceil(lowestBottom / H)`
locally turned the test red with `contentWindowCount = 11, want 2`, and — measured under the same mutant
— the one-window control still answered **3**. That control is what isolates the discriminating case: it
is exactly why the shipped `page-count-*` fixtures, spaced one window apart, could never have caught
this, because there the two routes coincide at every N. The mutant was reverted and `page_setup.go`
restored byte-identically (md5 `f5911f78…`).

**The mutation-proved gap, and its fix, re-proved here.** The review found the story's own pull-back
correction unverified at three of its four sites: reverting `containEdgeY` at `dropComponent:1411` and
`setComponentBounds:1670,:1672` left the **entire Go suite green**. The reason is worth stating plainly —
the positivity guard at `component_commands.go:1665` runs *before* the pull-backs, so `containEdge` with a
negative limit floors to zero and the mutant commits `height = 0`: silent data loss on a path this
story's own lift opens by widening the probe gate. `TestTheColumnLiftIsExercisedAtTheCommandSurface` is
the fix, and all three mutants were re-run at this revision: the drop-site mutant reddens with `dropped
box (y 654000 + height 24000) was pulled back inside the one-window band height 679890`; the bounds-site
mutant with `snapped bounds committed height = 0; the pull-back collapsed the component`; and a
`containEdgeY` that never clamps in any band reddens with `an edge move inside the pageHeader was
refused`. Restored byte-identically (md5 `6e7f239f…`).

**Projection neutrality was measured corpus-wide, not argued.** The build rejected a reviewer's ask for a
corpus-wide projection enumeration on the grounds that the golden digests plus the untouched projection
suite cover AC10. That reasoning is sound but it is reasoning; the closing pass measured it instead.
Every one of the **22 fixture templates** carrying an `input.folio` was projected through
`CanvasWithTextPaint` at baseline `4bf201a` and at `2b173a1`, against the shipped Noto set, with the two
new keys stripped: **123,433 bytes of JSON, byte-identical, zero projection errors on either side.** The
R5 collapse is therefore neutral in fact as well as in arithmetic — and the arithmetic was re-derived
independently: `canvasPageGeometry` takes its margins from `t.doc.Page.Margin` and its band heights from
the same two fields the old inline expression read, so `layout.ContentHeight(g)` and
`innerH - header - footer` are the same five `geom.Length` terms with the same signs.

**What the counts actually say, end to end.** Measured across the corpus at this revision: `page-count-1`
→ **1**, `page-count-5` → **5**, `page-count-20` → **20**, `page-count-50` → **50** — the canvas agrees
with the render on the fixtures whose page counts are known by name. And **`statement-1`, `-5`, `-20` and
`-50` all report 1**, while printing one, five, twenty and fifty pages. That is R2's bound-table floor,
observed rather than predicted: `projectedSize` gives a table `(sum of column widths, HeaderHeight)` —
header only, no rows — because the canvas has no data. **Story 7.6 must not read this number as a
prediction.** It is recorded in the projection's own comment, in the plain-terms opener, and here.

**The Go/TS tie is real in both directions.** The band-containment mirror is the first *predicate* tie in
`engine-bounds-mirror.test.ts` — DW-25 closed only the four size caps, and band containment was a fifth
mirror it never covered, because an audit closes only what it measured. Audited at this revision: it
asserts each side's list resolves to exactly `['pageHeader','pageFooter']` (so a regex that quietly
stopped matching cannot make the equalities vacuously true), asserts `content` is on neither, asserts
each side *consumes* its list at the validator it governs, asserts the horizontal cap stays unguarded on
both sides, and red-proves a one-sided edit **in each direction** with a `not.toBe` guard confirming the
drift edit actually changed the source. The lift landed in `cd196f0` — the same commit as the Go split,
confirmed by `git show --stat`.

**Truncation independence holds at the source, not just in a test.** The window count's extents are
emitted from `for i := range lines` — the full, untruncated list — placed *before* the `painted` loop and
using an origin derived from `len(lines)`. It reads neither `painted`, nor the fragment budget, nor the
oversized flag. `canvas-authority-contract.test.ts` is **absent from the whole diff**; its two prohibited
regexes, their non-vacuity and their two window-count red-proofs are intact and green.

**Verification, measured at the closing revision.**
- `cd folio-go && go test -count=1 ./...` — **exactly ONE** distinct red, the mandated permanent
  `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)`, `got 7, need >=20`; stats
  `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. Every other package `ok` (14 with tests).
  Untouched, as required.
- `go vet -tags=matrix ./...` — clean, exit 0. `gofmt -l folio-go` **from the repo root** — no output.
- `TestTargetRenderHash` once per leg with `FOLIO_MATRIX_TARGET` exported: **darwin/arm64, linux/amd64,
  linux/arm64, js/wasm — all PASS, `grep -c "asserts NOTHING"` = 0 on every leg**, so no leg was a
  no-op. Re-run clean after the mutation work, since the first run overlapped it.
  `TestCrossTargetByteIdentity` — PASS (20.15s).
- `cd lint && go test -count=1 ./...` — 4 packages `ok`.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; oxlint
  **exactly 4 warnings, 0 errors**, the baseline `only-export-components` set, no fifth; **248 tests /
  32 files passed** against a 239/32 baseline (+9). `npm run test:e2e:compile` — pass. Browser e2e is
  deferred by D-000.4 and **did not run**; no claim is made that it did.
- **Nine digests re-measured byte-identical**, `git status fixtures/` empty: statement-1 76,744
  `114df1d6`; -5 127,363 `70dce051`; -20 269,884 `56bfbbd9`; -50 555,829 `5d090b0f`; mandatory-break
  56,681 `7cf743de`; line-spacing 57,770 `de212115`; justified-text 59,894 `6da3b12e`;
  alignment-rounding 61,346 `986400a1`; justified-thai 15,079 `58ca4777`.
- **Manual checks.** `grep -n "containComponent(" folio-go/component_commands.go` at this revision:
  **twelve hits, eleven call sites** at `:165`, `:300`, `:829`, `:1409`, `:1472`, `:1565`, `:1569`,
  `:1606`, `:1668`, `:1674`, `:1726`, plus the definition (now `:1818`) — matching the audit table
  exactly. `internal/layout/paginate.go` is **absent** from `git diff 4bf201a..2b173a1`.
  `SupportedMajor` is still **2** and no `.folio` format field was added — no `internal/template/` file
  is in the diff at all. Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`, 8,470 bytes,
  unchanged and **absent from all four commits**.
- Known-environmental, not regressions: `TestShippedFacesReproduceFromUpstream` (no `fontTools`);
  `gofmt -l lint` reports `lint/internal/rules/licencegraph_test.go` (DW-23, Story 15.2).

**`followup_review_recommended` cleared to `false`.** The flag was raised by score (2 medium + 3 low
patched, nothing high). The extra scrutiny it asks for was performed and is recorded above: all three
`containEdgeY` mutants re-proved, the 2-vs-11 discrimination reproduced with the control held at 3, the
Go/TS tie audited for non-vacuity and red-proved both ways, truncation independence checked at the source,
projection neutrality measured across the whole corpus, and two rejections spot-checked — the
version-bump rejection verified against `791ed00`, which did add two projection keys with
`ENGINE_PROTOCOL_VERSION` unmoved at 1, and the corpus-enumeration rejection replaced with an actual
measurement. Nothing was re-opened. No production code was changed at close.

**Deferred, filed with owners.** **DW-36** — the designer's live drag/resize clamp still bounds Y by band
height for all three bands, so the lift is **not yet reachable by dragging, only by command**; owner
Story 7.6's plan gate. **DW-37** — `contentWindowCount` has no upper bound and the non-text extent is an
unguarded sum, the same axis as DW-26; owner Epic 7's retrospective or the next story touching canvas
geometry bounds. **Also for 7.6's plan gate:** a content component *taller than one window* — which this
story newly permits — makes the count degrade to **one**, so 7.6 would draw **one sheet** for it. A
floor, not a lie, and deliberate per Ruling C; but 7.6 must know it before it draws.

DW-26, DW-27, DW-28, DW-30, DW-31, DW-32, DW-33, DW-34 and DW-35 remain **OPEN**; DW-29 stays routed to
Story 7.8. **DW-33 needs a ruling, not a patch, and 7.5 left it untouched** — verified in code: the count
reads `lines`, `originY` and the vertical model, never `painted`, the fragment budget or the oversized
flag, so its answer is identical whether that element paints every line or none.
