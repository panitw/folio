---
title: 'Story 7.9: The canvas tells the truth about the window count'
type: 'bugfix'
created: '2026-08-31'
status: 'done'
baseline_revision: 'ddd6a3ffdcb3f39bc4a91bc4210bd5f5b4e6b94c'
review_loop_iteration: 0
followup_review_recommended: true
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: ['oversized']
deferred:
  - summary: >-
      A content-band TEXT element declaring a background or border contributes a
      full declared-box column item on the render path, while the canvas
      contributes only its shaped lines, so the canvas can UNDER-count.
    evidence: |-
      collectElementBoxRects is eligible for four kinds including text
      (folio-go/element_box.go's own header lists text, image, rect and line),
      and buildPageModel folds its output into the slice contentColumnItems
      turns into content-column items. Measured: a content band of a text at
      y 0 plus a styled text element at y 700 with height 200 and a short
      value projects canvas count 1, origins [0], exact TRUE against a real
      buildPageModel render of 2 pages. Every text element in every canvas
      window-count fixture is unstyled, so nothing reddens.
      NOT fixed here: closing it adds a NEW canvas column-item source rather
      than removing a spurious one, which is outside RULING B's stated subject
      (the unstyled NON-TEXT element) and is the kind of growth the spec's
      `story materially larger than the ruling implies` Block If guards. The
      overstated comment that claimed the invariant unconditionally WAS fixed.
    location: >-
      folio-go/page_setup.go (addCanvasWindowCount's text arm)
    severity: medium
  - summary: >-
      folio-go/diagnostic.go still restates AD-14 as "over-tall rows (FR25, not
      yet built)"; Story 7.7 built them and Story 7.9 widened the spine clause
      to cover author-declared keep-together groups.
    evidence: |-
      The spine sentence this comment paraphrases now reads "Rows and
      author-declared keep-together groups too tall for one content window
      (FR25, FR51)". The "not yet built" half has been false since ed485eb.
    location: >-
      folio-go/diagnostic.go:344
    severity: low
  - summary: >-
      folio-go/internal/diag/diag.go cites ARCHITECTURE-SPINE.md by line number
      and that line falls inside a mermaid ER diagram.
    evidence: |-
      Already stale before this story touched anything; the AD-14 rewrap
      deliberately held the file at 722 lines so the citation is no worse. The
      spec's Design Notes direct that it be carried, not repaired here.
    location: >-
      folio-go/internal/diag/diag.go:10
    severity: low
---

<intent-contract>

## Intent

**Problem:** Story 7.7 taught the render path to keep a declared keep-together group whole and did not
teach the canvas the same thing. `addCanvasWindowCount` builds its `layout.ColumnItem`s with **no
`Group`**, so for a grouped document the canvas reports a window **count** that is too low and window
**origins** at the wrong column positions — and reports the count as **exact**, because grouping is not
among `ContentWindowCountIsFloor`'s three causes. Measured: a body element, a two-member group at
y 700 / y 740 and an untagged tail at y 1440 renders **3** pages while the canvas reports **2**,
`IsFloor: false`, origins `[0, 740000]` where the render's second window begins at **700000**. Untag the
group and the canvas is right. This makes Story 7.6's AC2 false at HEAD and gates `epic-7: done`
(D-7.7.8).

**Approach:** Tag the canvas's column items with the same keep-together groups the render path already
uses, so count and origins become correct **by construction**. `keepTogetherTags(t *Template)` takes the
Template and nothing else — no data, no params, no `FontSet` — so grouping is a pure template property
and the canvas already holds every input it needs. This is a **defect fix**, not a disclosure: it
registers **no** fourth floor cause. One fix closes both halves, because Story 7.6 projected origins from
`pages[page].Shift` rather than computing them, so the real `Paginate` produces true origins as a
by-product. Riding with it: a duplicated component must join no group (D-7.7.10), and
`ARCHITECTURE-SPINE.md`'s over-tall carve-out is widened to describe what the engine already does
(D-7.7.14 half (a)).

## Boundaries & Constraints

**Always:**
- The canvas's window **count** and window **origins** equal the render path's for a grouped document,
  asserted **directly against a real render** rather than against the flag. This assertion is the
  story's spine: its absence is what let the defect ship.
- `ContentWindowCountIsFloor` keeps **exactly its three existing causes**. No fourth cause is added.
- `internal/template/parse_bands.go`'s refusal of `keepTogether` on a table stays **asserted** — that
  refusal is what stops a group inheriting a table's data dependency, and it is what makes the
  no-new-floor-cause claim true.
- Reuse the shipped `ItemGroup` machinery and the existing `keepTogetherTags` / `keepTogetherGroup` /
  `orKeepTogether` authority. One authority gaining a second caller, exactly as Story 7.5's extents did.
- A duplicated component joins **no** group; the drop is **asserted, not incidental** — duplicate a
  tagged element, assert the copy carries no tag **and the original is unchanged**.
- Designer-side group **authoring** is out of Epic 7. This is a **stated scope boundary**, not an
  omission: FR51 asks only that a group can be *declared*, and file-only authoring is in scope.
- A template declaring no groups renders **byte-identically** and projects **identical** canvas values.
  All 21 `goldenDigestRecord` entries stay byte-identical.
- Commits go to `main` only. Never branch, never push. Never `git add -A` or `git add .` — explicit
  paths only. `README.md` at the repository root is the user's file and is never touched.

- **THE HALT IS RESOLVED. Three rulings, 2026-08-31, after this story halted on its own `Block If`
  with blocking condition `grouping case the canvas cannot know`. Implement them; do not re-derive
  them.** The halt was correct and the retained commit `6a31a7f` is correct for its subject — build
  forward on it, do not revert.

- **RULING A — the `visibleIf` case IS a genuine cause, but it is NOT about grouping and it is NOT a
  floor.** `page_setup.go` only **projects** `VisibleIf` as a string; nothing evaluates it, so the
  canvas cannot resolve it without data and it passes the knowability test honestly. Two corrections
  to how it gets registered:
  - **Name it for CONDITIONAL VISIBILITY, not for grouping — it PREDATES this story.** An
    **ungrouped** element carrying `visibleIf` has the identical problem: the canvas places it, the
    render omits it, and AD-24 makes a hidden element **absent with no gap**, so the counts differ.
    That has been true since Story 7.5 shipped the count and 7.6 shipped the flag. **Grouping did not
    create this cause; it is how it was found.** Register it as *an element whose visibility depends
    on data*, and say in the register that it has been **undisclosed since 7.5**.
  - **It is a CEILING, not a floor**, and this is the part that changes a name. Check the
    directions: a **bound table** shows header height only and the render adds rows, so canvas ≤
    render — a **floor**. **Conditional visibility** has the canvas place an element the render
    omits, so canvas ≥ render — a **ceiling**. The measured numbers agree: canvas 3 against a render
    of 2-or-3. **A boolean named `IsFloor` set true on a ceiling case is a second confidently-wrong
    disclosure** — the exact failure D-7.7.8 gated the epic on, one layer down. And a document
    carrying **both** a bound table and a `visibleIf` element can be wrong in **either** direction,
    so no single direction is honest for the general case.
  - **The fix is a RENAME, not a redesign: `ContentWindowCountIsFloor` →
    `ContentWindowCountIsExact`.** One Go field, its TypeScript mirror, one UI string. **Do not
    build an enum.**
    **THE SENSE INVERTS, AND THAT IS THE POINT — do not ship `…IsApproximate`.** Its zero value is
    `false`, which reads *"this count is exact"*, so a projection path that forgets to set it
    **claims exactness** — the precise bug that started this whole thread, rebuilt into the field's
    default. A hazard indicator must not fail toward the quiet variant. `…IsExact`'s zero value
    `false` reads *"do not trust this count"*, so a forgotten set degrades to the **honest** claim.
    Free to decide now, expensive once 7.6's UI is built against it.
    **MUTATION-PROVE THE FLIP IN BOTH DIRECTIONS.** Inverting a boolean flips every call site and
    every assertion, it is the easiest thing in the world to get backwards, and **the corpus will
    not catch it because most documents are exact**. Required: a document that *should* be exact
    **reddens when the field is forced false**, and a document carrying any registered cause
    **reddens when it is forced true**.
  - **Direction is dropped DELIBERATELY, and that must be written down or it will be restored.**
    Direction would not change what the canvas should *do*: a floor means there may be more sheets
    than drawn, a ceiling fewer, and neither is a safe side to act on — so there is no decision it
    informs. **If the projection carries the cause set, direction belongs there** (a cause knows its
    own direction, so a future consumer derives it without the flag re-acquiring a claim). **If the
    projection carries only the boolean, the field's comment must say so:** *direction was
    deliberately dropped; it lives with the causes if anyone ever needs it.* Without that line a
    future reader restores a floor claim because it looks like lost fidelity.

- **RULING B — the unstyled non-text element is a DEFECT, it is IN SCOPE for this story, and it must
  NOT be disclosed.** `element_box.go:69-74`'s rule is `style.Background.Set || style.Border.Set` —
  **both pure template properties** — so the canvas can apply the identical rule with **no data**. It
  fails the knowability test in the other direction, and parking it in the disclosure is precisely
  what was refused when the fourth cause was refused the first time.
  **Why "pre-existing" does not make it out of scope:** this story **propagates** it into a number
  the flag claims is exact. A story whose gate is *the canvas is honest* cannot ship a fresh
  dishonest row. Measured, the four rows are:
  `untagged+styled 2/2` · `untagged+unstyled 2/2` · `tagged+styled 3/3 (fixed here)` ·
  `tagged+unstyled canvas 3 / render 2 (REGRESSED here)`.
  - **REUSE `element_box.go`'s predicate; do not restate it.** One authority, two callers — the same
    shape as `keepTogetherTags`. A second copy of *"styled means background or border"* is the drift
    this project keeps paying for.
  - **Verify against ALL FOUR rows, not just the regressed one.** The two untagged rows are correct
    today and must stay correct; changing what the canvas places can move them.

- **RULING C — this story's subject widens, and the title with it**, to *the canvas tells the truth
  about the window count*. It is still a **repair** of Epic 7 under D-7.7.12, not an extension.

**Block If:**
- The implementation finds a grouping case the canvas **genuinely cannot know**. HALT with status
  `blocked`, blocking condition `grouping case the canvas cannot know`, and return to the engineering
  lead **before any fourth floor cause is added**. Do not add one.
- Tagging the canvas items would require any change under `folio-go/internal/layout/`, or the work turns
  out **materially larger** than D-7.7.6 implies. HALT with blocking condition
  `story materially larger than the ruling implies` — D-7.7.8 routes that trade to the lead, and to the
  owner only if the lead says so.
- Any of the 21 golden digests moves. HALT with blocking condition `golden digest moved`.
  `fixtures/statement-signoff.json` is a **human attestation**: no agent writes its `reader`, `date` or
  `examined` fields.

**Never:**
- Never register a fourth `ContentWindowCountIsFloor` cause.
- Never edit any path under `folio-go/internal/layout/`. `paginate.go`'s prohibition on a closed-form
  page count stands untouched; `TestPaginateNeverProducesAnEmptyPage` and the four-rule tests must be
  green **and unmodified**.
- Never introduce a second grouping model, and never derive a count or origin from a closed form
  (`ceil(lowestBottom / H)` or `index * ContentWindowHeight`).
- Never widen into designer grouping UI — no inspector control to view, set or clear a tag.
- Never add the over-tall **discriminator** clause to the spine (that an individually over-tall element
  is fatal regardless of tagging). That describes behaviour which does not exist yet and rides Story
  7.10. A spine running ahead of the code is the same defect as one lagging it.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Count divergence, the story's subject | Grouped template: body element, two-member group at y 700 / y 740, untagged tail at y 1440 | Canvas `ContentWindowCount` **equals** the real render's page count (3, not 2) | No error expected |
| Origin divergence | `fixtures/keep-together/` — group astride the window ceiling; render's second window begins at the group's first member, not its overflowing tail | Canvas `ContentWindowOrigins` **equals** the render path's `plan.Pages[i].Shift` sequence | No error expected |
| No new floor cause | Any grouped document that is otherwise well-formed | `ContentWindowCountIsFloor` is **false**; grouping never sets it | If grouping forces it true, HALT `grouping case the canvas cannot know` |
| Ungrouped twin, byte identity | `keepTogetherUngroupedTemplateJSON` and every untagged fixture | Rendered bytes and every projected canvas value **unchanged** from baseline | No error expected |
| Table refusal preserved | Element of type `table` carrying `keepTogether`, tag or explicit `null` | Load error on field `keepTogether`, element id reported; stays asserted | Located `*LoadError` |
| Duplicate drops the tag | `duplicateComponent` on a content-band element carrying `keepTogether: "signature"` | The copy carries **no** tag (absent, not `null`); the **original is unchanged** | No error expected |
| Non-monotonic origins under grouping | Grouped column whose shifts are not strictly increasing from 0 | Must not occur; the existing origin-protocol guard would degrade to a floor | HALT `grouping case the canvas cannot know` |

</intent-contract>

## Code Map

**Every anchor below was verified by reading the file at baseline `c3df718`.** Two premises supplied in
the dispatch were re-measured and are corrected at the end of this section.

### The defect site — `folio-go/page_setup.go` (package `folio`)

- **`:627`** — `func addCanvasWindowCount(t *Template, projection *CanvasProjection, column canvasColumnExtents) error`.
  Doc `:604-625`. Builds `[]layout.ColumnItem` in two tranches, then calls `layout.Paginate` at `:670`.
  It **imports `internal/layout` at `:13`** and has `t *Template` in scope as its own parameter.
- ⚠ **THE TWO COLUMN-ITEM LITERALS, AND BOTH ARE UNTAGGED.** These are the whole defect:
  - **`:647-657`** — the non-text arm, inside the loop `:639-658` over
    `t.doc.Bands.Content.Elements`, skipping `template.ElementText` at `:640-645`. Fields set:
    `ElementID`, `Top`, `Bottom`, `Rects: []layout.RectRef{0}`. **No `Group`.** Images, rects and
    tables all pass through this one arm; there is no separate image branch.
  - **`:948-953`** — the text arm, inside `addCanvasTextPaint` (`:823`), one item per **shaped line**,
    accumulated into `column.Items` and copied wholesale at `:638`. Fields set: `ElementID`, `Top`,
    `Bottom`, `Runs`. **No `Group`.** `addCanvasTextPaint` also has `t *Template` in scope (`:823`).
  - **Both must be tagged.** Tagging only one is a partial fix: the shipped fixture's group has **text**
    members (`e2`, `e4`) and a **rect** member (`e3`), so the equality AC cannot pass unless both arms
    carry groups. `addCanvasTextPaint` runs at `:524`, before `addCanvasWindowCount` at `:526`.
- **`:565-584`** — `canvasWindowOrigins(plan)`. Reads `page.Shift` at `:570-571`, appends `int64(shift)`
  at `:581`. ⚠ **It validates a protocol at `:572-580`: `origins[0] == 0` and strictly increasing.**
  Failing it degrades to a floor. Doc `:550-557`: *"Nothing is derived, and in particular nothing is
  multiplied."*
- **`:358`** — `ContentWindowOrigins []int64` (`json:"contentWindowOrigins"`), millipoints,
  band-relative to the content column. Assigned at `:701` from the call at `:690`.
- **`:376`** — `ContentWindowCountIsFloor bool`. ⚠ **Doc `:359-375` enumerates the THREE causes** and is
  the text that must remain true: (a) `:363-365` a content-band table with a non-empty binding;
  (b) `:366-368` Paginate could not place the column at all; (c) `:369-371` a text element contributed
  no extents because its font chain would not resolve. `:372-375`: *"What it must never do is decide for
  itself that, say, a table means more pages: that would be a second authority on a question this flag
  answers exactly."*
- **Write sites of `true`** — all pre-existing, all read-only for this story: `:702` (from
  `floor := column.FontChainDegraded || canvasContentBandHasBoundTable(t)` at `:636`, causes a+c);
  `:687` (Paginate error, guard `:671`, cause b); `:697` (origin protocol failed, guard `:690-691`,
  folded into cause b by the inline comment `:692-694`). Also `:503` — the shapeless `Canvas` entry
  point hardcodes `true` and does **not** call `addCanvasWindowCount`; it is out of scope.
- **`:591-598`** — `canvasContentBandHasBoundTable`, condition `:593`. Read-only.
- **`:307-313`** (`ContentWindowCount` doc) and **`:340-349`** (origins doc): both forbid the closed
  form by name. Read-only, and they stay true.

### The authority being given a second caller — `folio-go/render.go` (same package)

- **`:2017`** — `func keepTogetherTags(t *Template) keepTogetherIndex`. **Takes only `*Template`.** Nil-safe
  at `:2018-2020`; returns a nil map when nothing is tagged. Reads only
  `t.doc.Bands.Content.Elements[i].KeepTogether` (`:2022-2029`).
- **`:2010`** — `type keepTogetherIndex map[string]string` (element id → author's tag).
- **`:2049`** — `func (idx keepTogetherIndex) keepTogetherGroup(elementID string) layout.ItemGroup`.
  Pure map lookup plus a struct literal `:2055-2059`. No data, no FontSet. Safe on a nil receiver.
- **`:2069`** — `func (idx keepTogetherIndex) orKeepTogether(g layout.ItemGroup, elementID string) layout.ItemGroup`.
  ⚠ **The guard is its first two lines, `:2070-2072`: `if g.Present { return g }`** — this is what leaves
  an existing table-row group untouched and preserves byte identity. Doc `:2061-2068`.
- **`:1990`** — `const keepTogetherKeyPrefix = "keepTogether:"`. The `:` cannot appear in a real element
  id (`internal/template/ids.go` admits only `^e[0-9a-z]+$`), which is what makes every table path
  unreachable. **`:2000`** — `keepTogetherGroupIndex = -2`.
- ⚠ **`page_setup.go:1` and `render.go:1` are both `package folio`, same directory.** So this is a
  same-package call: no import, no cycle, no unexported-access problem. **Reusability is confirmed, not
  assumed** — D-7.7.8's flip-to-the-owner condition ("the canvas cannot reuse `keepTogetherTags`") does
  **not** trigger.

### The render path this must equal — read-only reference

- **`folio-go/page_number.go:87`** — `contentColumnItems(contentRuns, imageRuns, tableRects, visible, keepTogether)`.
  Appends **text → images → rects** (`:88`, `:115`, `:145`). Three substitution sites: `:104`
  (`orKeepTogether(lineRowGroup(), …)`), `:132` (`keepTogetherGroup(…)`, images), `:157`
  (`orKeepTogether(chromeRowGroup(), …)`). Documented at `:46-50` as deliberately the reverse order of
  `paginateDocument`. Called from `render.go:1676`, right after `keepTogether := keepTogetherTags(t)` at
  `render.go:1675`.
- **`folio-go/render.go:2120`** — `paginateDocument`. Appends **rects → text → images** (`:2158`,
  `:2196`, `:2256`); tags at `:2183`, `:2237`, `:2279`. Both orders produce the same partition —
  `TestBothPaginationPassesAgreeWithAKeepTogetherGroup` (`keep_together_fixture_test.go:654`) pins it.
- **`folio-go/internal/layout/paginate.go`** — `ColumnItem.Group` `:143`; `type ItemGroup` `:157`
  (`Present` `:161`, `Key` `:167`); `ItemGroupKey` `:175-178`; `PageAssignment.Shift` `:500`, set at
  `:801` and `:1001` as `windowStart - contentTop`. ⚠ **ZERO PATHS UNDER `internal/layout/` MAY CHANGE.**
- **Read-only and must stay green AND unmodified:** `internal/layout/paginate_test.go:275`
  `TestPaginateNeverProducesAnEmptyPage`; the four-rule tests `:80`, `:169`, `:231`, `:311`, `:362`,
  `:394`, `:450`; `paginate_group_test.go` `:59, :134, :197, :242, :298, :340, :397, :482, :556`.

### The absent assertion — `folio-go/canvas_window_count_test.go`

- ⚠ **`:35-51`** — `renderPathWindows(t, tpl) int`, the existing "render path oracle". **Line `:46` is the
  reason this defect shipped**, verbatim:
  ```go
  plan, err := layout.Paginate(mustPageGeometry(t, tpl), contentColumnItems(runs, nil, nil, nil, nil))
  ```
  The **fifth argument of `contentColumnItems` is `keepTogether keepTogetherIndex`**, and the oracle
  passes `nil`. So the render side is deliberately ungrouped too, and
  `TestCanvasWindowCountAgreesWithTheRenderPathOracle` (`:127`, asserted at `:111` and `:148`) is
  **vacuous for grouping** — both sides are wrong in the same way and agree. It also returns only
  `len(plan.Pages)`: **there is no origins oracle anywhere.**
- Existing tests, all read-only unless named in a task: `:64`, `:102`, `:127`, `:161`, `:190`, `:213`,
  `:270` (`TestCanvasProjectsWhereEachWindowBegins`, asserts hardcoded `[0 7280000]` and
  `[0 728000 1456000]`), `:299`, `:352`, `:395`. Helpers `assertWindowOriginsAreWellFormed` `:240`,
  `sheetOf` `:374`.
- **`folio-go/canvas_window_count_template.go`** — the established idiom for canvas-window subjects is a
  **Go template const, not a fixture directory**: `canvasWindowCountGapTemplateJSON` `:43`,
  `…ControlTemplateJSON` `:78`, `…OversizedTemplateJSON` `:111`, `…BoundTableTemplateJSON` `:142`,
  `…UnshapedTextTemplateJSON` `:178`. Window height on this geometry is **727890 mp** (`:70`).

### The duplicate path — DW-48

- **`folio-go/component_commands.go:1723-1757`** — `duplicateComponent`. ⚠ **`:1742` is `clone := *element`**,
  a whole-struct copy that carries **every** field of `template.Element`, `KeepTogether` included. Only
  two fields are overwritten afterwards: `:1743` `clone.ID = template.AllocateElementID(t.doc)` and
  `:1753` `clone.X, clone.Y = x, y`. Appended at `:1754`.
- **`folio-go/internal/template/model.go:219`** — `KeepTogether Presence[string]` (doc `:201-218`).
  `Presence[T]` is `internal/template/presence.go:16-20` (`Set`, `Null`, `Value`). ⚠ **"No tag" is the
  zero value `Presence[string]{}` (`Set:false`), NOT an explicit null**: `Set:true, Null:true`
  serializes back as `"keepTogether": null` (`serialize.go:250-251`) and still raises the required
  format version (`version.go:231`). Clear to the zero `Presence`, not to null.
- **The in-tree idiom for dropping an optional field** (no clone-and-clear precedent exists; this is the
  spelling to copy): `component_commands.go:1029` `element.VisibleIf = template.Presence[string]{}`;
  also `:358`, `:463`, `:1243`.
- **Reachable from the UI today**, not test-only: `folio-designer/src/component-command.ts:38-40`
  issues it; `folio-designer/src/App.tsx:459` `duplicateSelection`, bound to ⌘/Ctrl+D at `App.tsx:681`
  and a visible Duplicate button at `App.tsx:763`. Go dispatch `component_commands.go:70-71`.
- **The only existing duplicate test** is `folio-go/wasm/engine_test.go:507`
  `TestEngineDuplicateIsACommittedGoCommand`, and its **only** field-level claim is that the id changed
  (`:521-523`). **Nothing anywhere asserts which fields are or are not copied.**
- **Where the new test belongs:** `folio-go/component_commands_test.go` (19 `func Test`s, none touching
  duplicate). Neighbours: `:52`, `:76`, `:551`, `:715`. Helpers `componentTemplate(t)`,
  `projectedBands(t, tpl)` `:753`. ⚠ The fixture element must sit in the **content** band — parse
  refuses `keepTogether` elsewhere.
- ⚠ **No designer-side group authoring exists**: zero hits for `keepTogether` across
  `folio-designer/src/`, and `updateComponentProperties` never mentions the field. Confirmed scope
  boundary, not an accident.

### The refusals that must stay asserted — read-only

- **`folio-go/internal/template/parse_bands.go:254-256`** — the **table** refusal, inside
  `if ktRaw, ok := obj["keepTogether"]; ok {` (`:250`) and **before** the null branch (`:257`), so
  `"keepTogether": null` on a table is refused too. Asserted by
  **`internal/template/keeptogether_test.go:111` `TestKeepTogetherIsNotValidOnATable`**, over both key
  forms (`keepTogetherKeyForms` `:64`), asserting `Field`/`ElementID` at `:122-124`, with a positive
  control at `:128-130`. **This test must stay green and unmodified.**
- `parse_bands.go:251-253` — content-band-only refusal, asserted by `keeptogether_test.go:79`.
  `:264-266` — empty-tag refusal, asserted by `keeptogether_test.go:137`.

### The AST tripwire — stays green, but know why

- **`folio-go/table_row_clip_test.go:1076`** (test func `:976`) —
  `TestAPresentItemGroupIsATableRowOrAKeepTogetherGroup`. Scans for `layout.ItemGroup{…}` literals with
  `Present` non-false and `.Present = …` assignments **outside** the allowlist
  `{chromeRowGroup, lineRowGroup, keepTogetherGroup}` (`:1091-1095`). Vacuity floor 3 at `:1183`.
  ⚠ Calling `keepTogether.keepTogetherGroup(id)` from `page_setup.go` **constructs no literal there**, so
  the guard stays green and the floor only rises. Building an `ItemGroup` literal inline in
  `page_setup.go` would trip it — do not.

### The two documents that are false at HEAD

- **`fixtures/keep-together/README.md:60-64`**, section `## What it does not cover`, verbatim:
  > The canvas does **not** preview grouping: it builds its column items ungrouped, and grouping is not
  > among the causes that make its window count a floor. A declared group can therefore move a break the
  > canvas does not draw. That is a stated limit of this story, not something this fixture measures.

  **This becomes false the moment this story lands.** ⚠ The README is a declared `readme` digest site in
  `goldenDigestRecord` (`byte_neutrality_test.go:477`), and the completeness half of
  `TestGoldenDigestAgreesAtEveryDeclaredSite` requires the recorded sha256 string to stay present in it.
  Editing the prose is safe; deleting or altering the digest string is not.
- **`ARCHITECTURE-SPINE.md`** at
  `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`.
  AD-14 spans `:311-324` as three bullets (`Binds:` `:313`, `Prevents:` `:314-315`, `Rule:` `:316-324`);
  no `Why`/`Consequences` sub-headings. **The carve-out is inside the Rule bullet, `:319-320`**, verbatim:
  > Over-tall rows (FR25) and clipped content (FR44) are
  > `Warning`s returned alongside PDF bytes, never silent and never fatal.

  The target substring is `Over-tall rows (FR25)` on `:319`. Confirmed scoped to **rows** only: the file
  has zero occurrences of "keep" in any grouping sense. Stale since **`ed485eb`** ("Keep a signature
  block together across a page break", Story 7.7), whose 31-file diff amended `folio-format.md` and the
  decision log but **not** the spine. The already-widened downstream text to be consistent with is
  `_bmad-output/specs/spec-folio/folio-format.md:215` and `:224`.
- **Guards over the spine:** `lint/internal/rules/stagerank_test.go:226/:234`
  `TestSpineStageLadderMatchesStageRankTable` parses the **stage ladder**, not AD-14 — a prose edit inside
  the Rule bullet should not disturb it, but do not touch headings.

### Anchor corrections — two dispatch premises re-measured

1. ⚠ **The "284 tests / 34 files guard added at `c3df718`" does not exist.** Searched for the literal
   `284` across `folio-designer/**`, the Makefile, `.github/` and repo-wide markdown/yaml: every hit is
   the unrelated `P6e:284` corpus statistic, and `git show --stat c3df718` adds no such assertion (its
   only designer file is `folio-designer/src/property-prose-height.test.ts`, 4 `it`s). 284/34 is the
   **arithmetic HEAD baseline** (280+4 tests, 33+1 files) recorded as prose in the Story 7.7/7.8 records,
   asserted by no in-tree guard. Treat it as an expected count to report, not a gate that will fail.
2. ⚠ **`fixtures/statement-signoff.json` has no `sha256` field.** Its fields are `reader` `:2`, `date`
   `:3`, `examined` `:4` and **`digests`** `:5-10` (a map of fixture slug → sha256). Schema at
   `folio-go/statement_golden_fixture_test.go:161-177`; the reader is
   `statement_signoff_matrix_test.go`, whose line 1 is `//go:build matrix`. The prohibition is unchanged
   and absolute: **no agent writes `reader`, `date` or `examined`.**

## Tasks & Acceptance

⚠ **This section was re-derived on 2026-08-31 from the AMENDED intent contract (rulings A / A2 / B / C)
and from the retained commit `6a31a7f`. See `## Spec Change Log`.** The contract itself is unchanged
and remains the sole authority; this section had been left describing only the pre-halt work.

### Already delivered at `6a31a7f` — RE-VERIFY, do not redo and do not revert

The halt was on the spec's own `Block If`; the commit was retained, and it is correct for its subject.
Build **forward** on it. Confirm each of these still holds, then leave it alone:

- `folio-go/page_setup.go` — both column-item literals carry the document's author-declared groups,
  taken from one `keepTogetherTags(t)` index (the non-text arm and the per-shaped-line text arm inside
  `addCanvasTextPaint`). No `layout.ItemGroup` literal is constructed in `page_setup.go`.
- `folio-go/canvas_window_count_test.go` — `renderPathWindows` passes the real `keepTogetherTags(tpl)`
  as `contentColumnItems`' fifth argument instead of `nil`, and returns **origins as well as the
  count**. The equality, ungrouped-twin and no-fourth-cause tests exist.
- `folio-go/canvas_window_count_template.go` — the grouped count discriminator and its untagged twin.
- `folio-go/component_commands.go` — `duplicateComponent` clears `clone.KeepTogether` to the zero
  `Presence` (DW-48 / D-7.7.10).
- `folio-go/component_commands_test.go` — the drop is asserted, along with the untouched original.
- `fixtures/keep-together/README.md`, `ARCHITECTURE-SPINE.md` — see the two tasks below, which
  **amend** what landed rather than starting it.

### This dispatch's work — the three rulings

**RULING A2 — rename the flag and INVERT its sense.**

- `folio-go/page_setup.go` -- rename the `CanvasProjection` field `ContentWindowCountIsFloor` to
  **`ContentWindowCountIsExact`** (json tag `contentWindowCountIsExact`) and **invert its sense at
  every write and read site**: it is `true` only when the canvas's count is trustworthy, and `false`
  whenever any registered cause applies. There is **one Go field, no enum**. The pre-existing write
  sites of `true` become writes of `false`, and the default/exact path must now set `true`
  explicitly. -- The zero value must be the safe claim: `false` reads *"do not trust this count"*, so
  a projection path that forgets to set it degrades to the honest claim instead of claiming exactness.
- `folio-go/page_setup.go` -- rewrite the field's doc comment: it now enumerates the **causes that make
  the count untrustworthy**, in **both** directions, and it must carry the sentence that **direction
  was deliberately dropped and lives with the causes if anyone ever needs it**. Without that line a
  future reader restores the floor claim mistaking it for lost fidelity. -- D-7.9.6 requires the note
  be written where the field is declared; the projection carries only the boolean, not a cause set.
- `folio-go/page_setup.go` -- register **conditional visibility** as a cause, named for what it is:
  *an element whose visibility depends on data*. The register entry must state that it applies to an
  **ungrouped** `visibleIf` element exactly as much as to a grouped one, that AD-24 makes a hidden
  element absent **with no gap**, and that it has been **undisclosed since Story 7.5** shipped the
  count. **Grouping is how it was found, not what caused it.** -- RULING A: `page_setup.go` only
  projects `VisibleIf` as a string; nothing evaluates it, so the canvas cannot resolve it without data.
- `folio-go/page_setup.go:503` -- the shapeless `Canvas` entry point hardcoded the old flag `true`
  (*"this is a floor"*); under the inverted sense it must hardcode `ContentWindowCountIsExact: false`.
  ⚠ A mechanical rename that leaves this literal as `true` silently converts a documented floor into a
  claim of exactness. -- This is the single highest-risk site of the inversion.
- `folio-designer/src/engine-protocol.ts` -- rename the mirrored field on the canvas projection type,
  in `hasOnly`'s **exact key list** (`:207`) and in the `typeof … !== 'boolean'` guard (`:222`), and
  update the comment at `:144`. ⚠ `hasOnly` is an exact-key check: a Go-side rename without this one
  makes `parseInbound` discard the **whole snapshot** with no attributable error.
- `folio-designer/src/sheet-stack.ts` -- rename the derived `isFloor` field (`:69`, `:133`) to match
  the inverted sense, and `folio-designer/src/App.tsx` -- invert the condition at `:1358`/`:1360` and
  **rewrite the one UI string** `canvasFloorClaim` (`:1352`). Its current text, *"prints more pages
  than are shown here"*, is direction-committed and is now false for a ceiling cause. -- One UI string,
  and nothing else in the designer changes; grouping **authoring** stays out of scope.
- `folio-designer/src/*.test.ts(x)` -- update the fixtures and assertions that name the old key
  (`engine-protocol.test.ts`, `sheet-stack.test.ts`, `App.test.tsx`, `DataPanel.test.tsx`).
- **MUTATION-PROVE THE FLIP IN BOTH DIRECTIONS — mandatory, and this is the acceptance for the
  inversion.** (i) Force the field `false` on a document that *should* be exact and confirm a test
  reddens; (ii) force it `true` on a document carrying a registered cause and confirm a test reddens.
  Record both mutants and their verbatim failure output. -- Inverting a boolean flips every call site
  and assertion, and **the corpus will not catch a backwards one, because most documents are exact**.

**RULING B — fix the unstyled non-text element; do NOT disclose it.**

- `folio-go/element_box.go` -- extract the styled test currently inlined at `:69-74`
  (`style.Background.Set && !style.Background.Null`, `style.Border.Set && !style.Border.Null`) into
  **one named predicate in this file**, and have `collectElementBoxRects` call it. -- **One authority,
  two callers**, the same shape as `keepTogetherTags`. A second copy of *"styled means background or
  border"* is the drift this project keeps paying for; **do not restate the rule in `page_setup.go`**.
- `folio-go/page_setup.go` -- make the canvas's non-text arm contribute a column item only where the
  **render path** would, calling that predicate rather than restating it, and correct the arm's comment
  (which currently asserts *"EVERY CONTENT COMPONENT CONTRIBUTES, STYLED OR NOT"*). ⚠ **Measure before
  gating: the predicate governs the kinds `element_box.go` itself governs.** `collectElementBoxRects`
  skips `template.ElementTable` outright, and an unstyled **image** still reaches the render path's
  column items through the separate image-run arm (`page_number.go:115`). Gating all three kinds on
  the styled test would drop tables and unstyled images the render path does place. The acceptance is
  the measured four-row table below, not any particular placement of the predicate.
- **Verify ALL FOUR rows, canvas against a REAL render** (`buildPageModel`), not only the regressed
  one — the two untagged rows are correct today and changing what the canvas places can move them:

  | document | canvas must be | real render |
  |---|---|---|
  | untagged + styled | 2, origins `[0 740000]` | 2 |
  | untagged + unstyled | 2, origins `[0 740000]` | 2 |
  | tagged + styled | 3, origins `[0 700000 1440000]` | 3 |
  | tagged + unstyled | **2** | 2 |

**RULING C and the two documents.**

- `fixtures/keep-together/README.md` -- the `## What it does not cover` section is **false at
  `6a31a7f`** and must be rewritten to describe HEAD: the canvas now previews grouping, and its window
  count and origins agree with the render path. **Leave the recorded sha256 digest string untouched** —
  it is a declared `readme` digest site in `goldenDigestRecord` and the completeness half of
  `TestGoldenDigestAgreesAtEveryDeclaredSite` requires the string to stay present.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` --
  widen AD-14's over-tall carve-out from **rows** to rows **and author-declared keep-together groups**,
  citing FR51 alongside FR25, **describing what the engine does today**. Add **no** discriminator
  clause — that describes behaviour which does not exist yet and rides Story 7.10. Prefer a rewrap that
  leaves the file's total line count unchanged. -- DW-49(a) / D-7.7.14 half (a).

**Acceptance Criteria:**

- Given a template declaring a keep-together group, when the canvas projection is built, then its
  `ContentWindowCount` equals the page count of a **real render** of the same template, and its
  `ContentWindowOrigins` equals the render path's `plan.Pages[i].Shift` sequence computed with the real
  `keepTogetherTags` — asserted directly against the render, **not** via the honesty flag.
- Given each of the four documents in the table above, when its canvas projection is compared to a real
  `buildPageModel` render, then the counts agree on **all four rows**.
- Given a well-formed grouped document with no registered cause, when the projection is built, then
  `ContentWindowCountIsExact` is **true**; and given a document carrying a registered cause, then it is
  **false**. No `ContentWindowCountIsFloor` identifier survives anywhere in Go or TypeScript.
- Given the honesty flag is forced to the wrong value in either direction, when the suite runs, then it
  **reddens** — both mutations are recorded with their verbatim output.
- Given the `CanvasProjection` field declaration, when it is read, then its comment states that
  direction was **deliberately dropped** and lives with the causes.
- Given an element of type `table` carrying `keepTogether` in either its tag or its explicit-null form,
  when the document is parsed, then it is refused with a located load error, and
  `TestKeepTogetherIsNotValidOnATable` is green and unmodified.
- Given a content-band element carrying a `keepTogether` tag, when `duplicateComponent` runs, then the
  copy's `KeepTogether` is absent (zero `Presence`, **not** an explicit null) and the original element
  is unchanged.
- Given a template declaring no groups, when it is rendered and its canvas projected, then its bytes are
  identical to baseline and all **22** `goldenDigestRecord` digests are unchanged.
- Given the tagging is reverted in `page_setup.go`, or the oracle's fifth argument is reverted to `nil`,
  when the equality tests run, then they **fail** — each is red-proved rather than assumed to bite.
- Given `folio-go/internal/layout/`, when the diff is inspected, then it contains **zero** paths under it.

## Spec Change Log

### 2026-08-31 — `## Tasks & Acceptance` and `## Verification` re-derived from the amended contract

Not a `bad_spec` loopback. The `<intent-contract>` was amended twice after this spec was planned
(commits `b6b21be` and `ddd6a3f`) to carry rulings D-7.9.1 … D-7.9.7, but the **derived** sections
below the contract were left describing only the pre-halt work. Three contradictions resulted, all of
which would have misdirected an implementer reading the spec as its own source of truth:

1. `## Tasks & Acceptance` listed only the tagging / oracle / duplicate / documents work — every item
   of which **already shipped at `6a31a7f`** — and named none of rulings A, A2 or B.
2. An acceptance criterion required `ContentWindowCountIsFloor` to be computed and `false`. RULING A2
   renames that field and **inverts** it, so the criterion had become unsatisfiable by construction.
3. `## Verification` said **21** `goldenDigestRecord` digests. Dispatch 2 measured **22** — Story 8.0
   added `thai-stacked-marks` at `26bfd49`. It also omitted `-count=1` on the `lint` module, whose
   cached green had been hiding two reds (D-7.9.5).

The contract itself was **not** edited. Both sections were re-derived from it, from the retained
commit, and from Dispatch 2's measurements; the four-row acceptance table and the both-directions flip
mutation are transcribed from the rulings rather than re-derived.

## Review Triage Log

### 2026-08-31 — Review pass

- intent_gap: 2: (high 2, medium 0, low 0)
- bad_spec: 0
- patch: 10: (high 0, medium 4, low 6)
- defer: 4: (high 0, medium 1, low 3)
- reject: 8: (high 0, medium 0, low 8)
- addressed_findings:
  - none

The two intent_gap findings are the halt. Everything below them is moot for this
pass by the cascade and was neither applied nor written to `deferred:`; it is
recorded here so nothing is lost when the story is re-dispatched.

**intent_gap 1 — `[high]` a group member carrying `visibleIf` is a grouping case the
canvas genuinely cannot know.** This is the spec's own `Block If`, reached by
measurement rather than by argument. A content-band element may carry both
`keepTogether` and `visibleIf`; membership in the group is then a pure template
property, but whether the member is *placed* is a property of the DATA. Measured at
`6a31a7f` on a group whose rect member is `"visibleIf": "showRule"`:

    CANVAS (no data at all): count=3 origins=[0 700000 1440000] floor=false
    REAL RENDER data={"showRule": true}   pages=3
    REAL RENDER data={"showRule": false}  pages=2

The canvas answers 3 and reports it as EXACT, and it cannot do better, because it has
no data. This falsifies the Approach's central premise as it is written — that
`keepTogetherTags` taking only the `*Template` makes grouping knowable canvas-side.
The tag INDEX is knowable; the grouping's EFFECT on pagination is not, once a member
is conditional. `ContentWindowCountIsFloor`'s three causes do not cover it: (a) is a
content-band table with a non-empty binding, (b) a Paginate failure, (c) a failed font
chain. The honest disclosure would be a fourth cause — which this spec forbids
absolutely, and whose addition the `Block If` reserves to the engineering lead. No
fourth cause was added.

**intent_gap 2 — `[high]` a non-text group member the render path does not place makes
the canvas confidently wrong, and this change is what made the COUNT wrong.** The
render path contributes a column item for a non-text element only when it declares a
background or a border (`folio-go/element_box.go:70-73`); the canvas contributes one for
every non-text content element, styled or not (`folio-go/page_setup.go:613-616`).
Measured at both revisions against the real `buildPageModel`:

    document                    canvas @26bfd49    canvas @6a31a7f    REAL RENDER
    untagged + styled           2  [0 740000]      2  [0 740000]      2 pages
    untagged + unstyled         2  [0 740000]      2  [0 740000]      2 pages
    tagged   + styled           2  [0 740000]      3  [0 700000 1440000]  3 pages   <- fixed here
    tagged   + unstyled         2  [0 740000]      3  [0 700000 1440000]  2 pages   <- regressed here

`ContentWindowCountIsFloor` is `false` in every row. The story fixes its subject at the
real render surface and, for the adjacent unstyled class, turns a correct count into an
incorrect one. The underlying population asymmetry is pre-existing (the untagged rows are
byte-identical at both revisions, and already diverge on origins); what is new is that
grouping propagates it into the count. Three resolutions are defensible and the contract
settles none: gate the tag on render-path placeability; disclose via a floor cause (the
contract's `Never` forbids it); or fix the population asymmetry (the contract's own
`Block If` for size). The AC is universally quantified — "Given a template declaring a
keep-together group" — so the intent's own words include this document rather than
excluding it, and the scope-authority rule forbids routing it out.

**patch (10, moot this pass)** — 1 `[medium]` the widened AD-14 sentence now reads as
"any author-declared keep-together group is a Warning", which is false; only an
over-tall one warns, and the shipped fixture renders with zero diagnostics. 2 `[medium]`
no coverage for a multi-line tagged text element (the canvas emits one item per shaped
line), two independent groups, non-adjacent members, or a group as the column's first
item. 3 `[medium]` `folio-format.md` records no rule that duplicating a tagged component
drops the tag; that user-visible rule lives only in a Go comment and a test name.
4 `[medium]` the table-refusal assertion checks only `err != nil`, not the located
`*LoadError`'s field and element id nor the explicit-null form the matrix names.
5 `[low]` `renderPathWindows`' doc comment does not disclose that it still passes `nil`
image runs and `nil` visibility verdicts, collects no table runs, and skips the
footer-orphan pass the shipping path applies. 6 `[low]` `fixtures/keep-together/README.md`
says "those three causes" with no antecedent in that file, and writes the origins as
"706.000"/"734.000" without a unit where the tests assert 706000/734000 millipoints.
7 `[low]` the non-text arm's comment names `table` among the kinds that join a group
there, which `parse_bands.go` makes unreachable. 8 `[low]` the Code Map states its anchors
were verified at `c3df718` while `baseline_revision` is `26bfd49`; all anchors were
re-verified at `26bfd49` in this dispatch and hold, with `keepTogetherKeyPrefix` and
`keepTogetherIndex` drifted to `:1992` and `:2011`. 9 `[low]`
`TestCanvasWindowCountIsIndependentOfPaintTruncation` discards the oracle's new origins
return. 10 `[low]` the FontSet is selected by string-comparing whole template constants
where the table already carries a name.

**defer (4, moot this pass — NOT written to `deferred:`)** — 1 `[medium]` the canvas
contributes a column item for every non-text content element while the render path places
only styled ones; undisclosed and `IsFloor: false`. Pre-existing: the untagged rows above
are identical at `26bfd49`. 2 `[low]` `folio-go/diagnostic.go:343-346` still restates AD-14
as "over-tall rows (FR25, not yet built)"; Story 7.7 built them. 3 `[low]`
`folio-go/internal/diag/diag.go:10` cites `ARCHITECTURE-SPINE.md:613`, which falls inside a
mermaid ER diagram (already recorded in Design Notes; the rewrap held the file at 722 lines,
so it is no worse). 4 `[low]` `lint/internal/rules` has two RED tests at baseline —
`TestFloatTypedTestScopeInventory` and `TestStageRankProductionScan`
(`text/shape_shipped_face_test.go:8` imports a higher stage rank). Both reproduce
byte-identically at `26bfd49`; they arrived with Story 8.0 and are not caused here.

**reject (8, all `[low]`)** — the claim that
`element.VisibleIf = template.Presence[string]{}` does not exist (it is at
`component_commands.go:1029`; verified); "the two arms share an authority, not an index"
(both call the same deterministic pure function of the same `*Template`, so they cannot
disagree — the invariant is structural); duplicated floor coverage across two tests
(deliberate); `reflect.DeepEqual` versus `slices.Equal` (the nil/empty distinction is
load-bearing here); the duplicate test's cross-file fixture coupling (style);
the absence of a designer-side assertion (the intent's stated scope boundary);
a speculative duplicate-`Shift` origin-protocol failure (not demonstrated — the protocol
held on every document measured, and grouping never set the floor); and DW-46/DW-48 still
reading `OPEN` in `deferred-work.md` (that register is not this workflow's file).

### 2026-08-31 — Review pass (Dispatch 3)

- intent_gap: 0
- bad_spec: 0
- patch: 10: (high 1, medium 4, low 5)
- defer: 3: (high 0, medium 1, low 2)
- reject: 6: (high 0, medium 0, low 6)
- addressed_findings:
  - `[high]` `[patch]` `canvasElementIsPlaced`'s image arm returned `true` unconditionally while
    `collectImageRuns` (`render.go:459-467`) skips a **null-asset** image — and `createComponent`
    (`component_commands.go:1486-1494`) gives EVERY newly dropped image exactly that state. Measured:
    canvas 2, origins `[0 1440000]`, exact **true** against a real render of **1** page. Fixed by
    extracting `imageDrawsItsAsset` in `render.go` and calling it (OR-ed with `elementDeclaresBox`,
    since a null-asset image that declares a box still reaches an element-box rect source). New matrix
    row plus a filled-asset discriminator; red-proved.
  - `[medium]` `[patch]` the same function's table arm ignored `collectBandTableRuns`' zero-column skip
    (`table_render.go:615-621`); `"columns": []` parses. Fixed by extracting `tableDrawsColumns` and
    calling it; new `TestCanvasCountsOnlyTheTablesTheRenderDraws`; red-proved.
  - `[medium]` `[patch]` the extracted "one authority" carried only the style clause of a two-clause
    render condition — `collectElementBoxRects` also requires `declaredBox`, which demands a strictly
    positive rectangle the loader accepts as zero or negative. Measured: a styled rect with
    `"height": 0` gave canvas 2 against a real render of 1. The positivity test was folded into
    `elementDeclaresBox` so the shared predicate is the whole rule; red-proved.
  - `[medium]` `[patch]` nothing tied the renamed wire key across the Go/TypeScript seam: Go tests read
    the struct field and the TS tests hand-author the string, so both sides could agree with themselves
    while disagreeing with each other, and `hasOnly` is an exact-key check whose failure mode is a blank
    canvas with nothing to attribute it to. Added `canvas_projection_wire_test.go`, which asserts the
    marshalled key set against a recorded literal AND extracts `isCanvas`'s own `hasOnly` list from
    `engine-protocol.ts` and asserts it against the same record. Red-proved in BOTH directions.
  - `[medium]` `[patch]` `canvasElementIsPlaced` and `elementDeclaresBox` both carried doc comments
    stating parity claims that were false ("Always placed, whatever the element's style says";
    "exactly where this file contributes a rect source and nowhere else"). Rewritten to the real
    per-kind conditions and to name what the predicate leaves to its caller.
  - `[low]` `[patch]` `addCanvasWindowCount` asserted "A COMPONENT CONTRIBUTES EXACTLY WHERE THE RENDER
    PATH PLACES ONE" unconditionally; it is false for a styled TEXT element in the opposite direction.
    The claim is now scoped to non-text components with the exception written out. Behaviour
    deliberately unchanged — see `deferred:`.
  - `[low]` `[patch]` `canvasContentBandHasConditionalVisibility`'s `VisibleIf.Value != ""` clause was
    unreachable (an empty `visibleIf` is refused at load) and unreachable **in the unsafe direction** —
    it would have claimed exactness. Dropped, with the reason recorded.
  - `[low]` `[patch]` stale "floor" prose survived at the two most-read sites and at two further
    comments, though the rename's whole point is that the word is wrong. Reworded to the inverted flag.
  - `[low]` `[patch]` the conditional fixture escaped the mechanical twin-equality discipline the file
    establishes for its other twins, so it could drift into differing for a second reason. Assertion
    added; red-proved by injecting a second difference.
  - `[low]` `[patch]` `fixtures/keep-together/README.md` called all the causes "things the canvas
    genuinely cannot know without the data", untrue of the Paginate-degradation cause, which the canvas
    observed in its own answer. Corrected; the recorded sha256 string untouched.

**Deferred (3)** — written to `deferred:`: the styled-TEXT box divergence (real, measured canvas 1 vs
render 2, but closing it ADDS a canvas item source and is outside RULING B's subject);
`diagnostic.go:344`'s "FR25, not yet built"; `internal/diag/diag.go:10`'s stale line citation.

**Rejected (6, all `[low]`)**, with the authority each was tested against:
- *AD-14's superseded wording still quoted in `internal/layout/paginate.go` and
  `paginate_group_test.go`* — the contract's `Never` fences **ZERO paths** under `internal/layout/`.
- *Cause (d) over-discloses: any content-band `visibleIf` sets the flag even where the element could
  not move a boundary* — a real observation, but D-7.9.6 settles the direction: the flag must fail
  toward the honest claim, so conservative over-disclosure is the intended behaviour, not a defect.
- *"The defect cause (d) discloses is not registered anywhere"* — it **is** registered, as cause (d) at
  the field's declaration, and RULING A directs that it be registered rather than fixed.
- *`sprint-status.yaml` still reads `ready-for-dev`* — that register is not this workflow's file.
- *"The mandatory mutation evidence is absent"* — it belongs in `## Auto Run Result`, not the diff, and
  it was performed; it is recorded below.
- *Two comments rewrapped raggedly* — cosmetic; `gofmt -l folio-go` is clean.

## Design Notes

**Why one fix closes both halves.** A wrong origin is genuinely a different failure from a wrong count —
a floor flag is a claim about a count and there is no flag on the origins array at all. But Story 7.6
*projected* origins from `pages[page].Shift` rather than computing them, so once the items carry their
groups the real `Paginate` produces true origins as a by-product. There is no second fix to write.

**Why no fourth floor cause.** The flag's three causes are each things the canvas *genuinely cannot know*
— a bound table's row count needs the data, the degraded placement needs the render, a failed font chain
needs the face. `keepTogetherTags` takes the Template and nothing else, so grouping is knowable
canvas-side. Registering it would park a defect inside the mechanism that exists to be honest about
shortfalls, and the author cannot avoid it: they declared a group in their file and the canvas is simply
wrong. A group's aggregate height still varies with bound text length, but that is the
canvas-versus-bound-data divergence the existing causes already name — and `parse_bands.go`'s table
refusal is what stops a group inheriting a table's data dependency, which is why that refusal must stay
asserted.

**The one invariant to re-verify while tagging.** `canvasWindowOrigins` (`page_setup.go:572-580`) refuses
any origin sequence that is not strictly increasing from 0 and degrades to a floor. Tagging changes the
`Shift` values. A keep-together slide should still produce strictly increasing shifts, but if it does not,
that is a floor reached *through grouping* — which is the `Block If`, not a fourth cause to register.

**On `multiple-goals` — the plan gate considered it and declined.** D-7.7.10 provides that if this gate
calls the pair `multiple-goals`, D-7.7.6 is the half that gates the epic and DW-48 splits out, and that
the split must not reverse. The gate declines to call it, for three reasons, recorded so the decision is
auditable rather than silent. (1) The lead already judged the subject unity: D-7.7.7 named this story and
ruled that "it carries D-7.7.6 and D-7.7.10", and D-7.7.14 states that half (a) "is not a second subject,
so it does not trip D-7.7.7's attribution concern". (2) The three parts share one subject — a surface
telling the truth about keep-together groups: the canvas lies about windows, the duplicate command creates
group state the author cannot see or clear, and the spine describes behaviour the engine no longer has.
(3) The route-out test this project uses requires the routed work to carry a decision above a builder's
authority; DW-48 carries none, because D-7.7.10 already ruled it completely. The split clause remains
available to the lead if implementation proves it wrong.

**A finding to carry, not to fix here.** `folio-go/internal/diag/diag.go:10` cites
`ARCHITECTURE-SPINE.md:613` by line number; at HEAD line 613 falls inside a mermaid ER diagram, so that
citation is **already stale before this story touches anything**. Widening AD-14 without changing the
file's line count leaves it no worse. Do not repair it here — it is unrelated to FR51 and belongs in the
deferred register.

## Verification

**Commands** (the D-R7.1 heavy set; run from the repository root unless stated):

- `cd folio-go && go test -count=1 ./...` -- expected: exactly **one** failure, the mandated red
  `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` (got 7, need >=20). Never touch it or its
  drift twin. Report measured pass/fail counts, not "green".
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean; proves matrix-tagged code still compiles.
- `gofmt -l folio-go` -- expected: no output. (`lint/…/licencegraph_test.go` has a known gofmt finding,
  DW-23, outside this path.)
- `cd folio-go && FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash -v .`
  -- and once each for `linux/amd64`, `linux/arm64`, `js/wasm`. Expected: each leg asserts.
  ⚠ **Plus an unset control**: run it with `FOLIO_MATRIX_TARGET` **unset** and confirm it is a deliberate
  no-op — that control is what proves the four legs were not no-ops. Say which legs ran.
  `TestShippedFacesReproduceFromUpstream` fails/skips without `fontTools`; environmental, not a regression.
- `cd folio-go && go test -tags=matrix -count=1 -run TestCrossTargetByteIdentity -v .` -- expected: pass.
- `cd lint && go test -count=1 ./...` -- ⚠ **`-count=1` is MANDATORY here.** The `rules` package scans
  `folio-go` by walking the directory and Go's test cache does not track `ReadDir` results, so a new
  file never invalidates it: this command was returning a **cached** four-`ok` while two tests were RED
  (D-7.9.5). Expected: **4 packages `ok`**. A cached green is no measurement at all.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` --
  expected: typecheck clean; lint with **exactly 4** pre-existing `only-export-components` warnings (a
  fifth is a regression); vitest **284 tests across 34 files** as the arithmetic HEAD baseline — report
  the measured numbers, and note that no in-tree guard asserts them; e2e compile-only (D-000.4).
- `shasum -a 256` over all **22** `expected.pdf` paths named in `goldenDigestRecord`
  (`byte_neutrality_test.go:92-487`) -- expected: **every digest identical to baseline `c3df718`**. Any
  movement is the `golden digest moved` HALT, not a re-record.

**Mandatory proofs for this dispatch** (report each verbatim; a prose "red-proved" line is not evidence):

- **The honesty-flag flip, BOTH directions.** (i) A document that should be exact reddens when
  `ContentWindowCountIsExact` is forced `false`; (ii) a document carrying a registered cause reddens
  when it is forced `true`. Neither is optional, and the corpus will not catch a backwards flip on its
  own because most documents are exact.
- **The four-row table**, canvas against a real `buildPageModel` render — all four rows, not only the
  regressed one.
- **The oracle revert.** Restore `renderPathWindows`' fifth argument to `nil` with production left
  correct; it must redden. This proves the repair is still non-vacuous.
- **Each column-item arm in isolation.** Neutralise the non-text arm's tagging alone, then the text
  arm's alone; each must redden on its own. A combined mutant hides whether the second site is covered.
- **The duplicate drop.** Remove the `clone.KeepTogether` clear; it must redden.
- ⚠ **Assert the mutation landed before concluding anything from it.** Count the occurrences of the
  substring being replaced — a doc comment carrying the same expression has silently absorbed a
  mutation in this repo before, which reads as "the test is vacuous" when it is not.

**Manual checks:**

- `git diff --stat` contains **zero** paths under `folio-go/internal/layout/`, and
  `internal/layout/paginate_test.go` is unmodified.
- `git status --porcelain` never shows `README.md`. Stage explicit paths only — never `git add -A` or
  `git add .`.
- `fixtures/statement-signoff.json` is unmodified; its `reader`, `date` and `examined` fields are
  untouched.
- Red-proof: revert the `page_setup.go` tagging, re-run the new equality tests, confirm they fail, then
  restore. Record the failure output.

## Auto Run Result

### Dispatch 1 — 2026-08-31, plan gate (`Halt after planning.`)

Status: `ready-for-dev`
Blocking condition: none
Baseline revision: `c3df7189a734d3a63bb72500159f845e8dc71a79` (`main`)
Commits created: none — the halt-after-planning directive stops the workflow at step 02, before any
implement or finalize step. No code was written.

**Dispatch mode:** classic intent (this repository has no `stories.yaml`). Spec written to the
story-key path directed by the orchestrator.

**Plan-gate determinations recorded for the orchestrator:**

- **D-7.7.8's flip-to-the-owner condition does NOT trigger.** It reserved a return to the owner if the
  canvas turned out unable to reuse `keepTogetherTags`. Verified at baseline: `page_setup.go:1` and
  `render.go:1` are both `package folio` in one directory, `page_setup.go` already imports
  `internal/layout` (`:13`), `keepTogetherTags` needs only the `*Template` both call sites already hold,
  and `keepTogetherGroup` / `orKeepTogether` are nil-safe pure lookups. No import cycle, no
  unexported-access problem, no data or `FontSet` dependency. The story is the size D-7.7.6 implies.
- **`multiple-goals` was considered and declined**, so DW-48 does **not** split out. Reasoning recorded
  in Design Notes. D-7.7.10's split clause remains available to the lead if implementation proves it
  wrong; the non-reversal condition is untouched.
- **Both DW-48 and DW-49(a) are carried by this spec**, as D-7.7.7 and D-7.7.14 direct.
- **No new fixture directory is required.** The count arm uses a new Go template const in the
  established `canvas_window_count_template.go` idiom; the origins arm reuses the shipped
  `fixtures/keep-together/`. This keeps the eight-site fixture-registration surface and the golden
  digests entirely out of the diff.

**Two dispatch premises re-measured and corrected** (details in Code Map, *Anchor corrections*):

1. The claimed designer guard asserting **284 tests / 34 files added at `c3df718`** does not exist in
   the tree. 284/34 is the arithmetic HEAD baseline recorded as prose in the Story 7.7/7.8 records; no
   in-tree assertion enforces it.
2. `fixtures/statement-signoff.json` has no `sha256` field; its fields are `reader`, `date`, `examined`
   and `digests` (a slug → sha256 map). The prohibition on an agent writing it is unchanged.

**One finding surfaced for the deferred register, not fixed here:** `folio-go/internal/diag/diag.go:10`
cites `ARCHITECTURE-SPINE.md:613` by line number, and at HEAD line 613 falls inside a mermaid ER
diagram — the citation is already stale before this story touches anything. Unrelated to FR51.

### Dispatch 2 — 2026-08-31, implement / review / triage

Status: `blocked`
Blocking condition: `grouping case the canvas cannot know`
Baseline revision: `26bfd49ef68e212ef041c8449f3ad7e2bfc3cc35` (`main`)
Commit created and RETAINED: `6a31a7fd1263177d806ccc95c75ed129bdb75df6` — "Tell the truth on
the canvas about keep-together groups", 8 files, no path under `folio-go/internal/layout/`,
root `README.md` untouched, trailer present.

**Why this is a halt and not a `done`.** The spec's `Block If` names this exact outcome:
"The implementation finds a grouping case the canvas **genuinely cannot know**. HALT ... and
return to the engineering lead **before any fourth floor cause is added**. Do not add one."
A tagged group member carrying `visibleIf` is such a case, and it was reached by measurement,
not by argument — see Review Triage Log, intent_gap 1. **No fourth cause was added.** The
`Block If` prescribes a halt and a return to the lead; it does not prescribe a revert, so the
commit is retained as the artifact the ruling is about. `git reset --hard 26bfd49` reverses it
exactly if the lead prefers a clean baseline (it is local and unpushed).

**What the change does, and it is correct for its subject.** Both canvas column-item literals
now carry the document's author-declared groups, taken from the shipped authority:
`page_setup.go:670` (the non-text arm — rects, lines, images, table boxes) and
`page_setup.go:987` (the per-shaped-line text arm inside `addCanvasTextPaint`). No
`layout.ItemGroup` literal is constructed in `page_setup.go`, so the AST tripwire stays green.
Validated at the SHIPPING surface, not only against the oracle: for the grouped fixture
`buildPageModel` returns 3 pages and the canvas answers 3; for its untagged twin, 2 and 2.

**Files changed**
- `folio-go/page_setup.go` — tag both column-item arms from one `keepTogetherTags(t)` index.
- `folio-go/canvas_window_count_test.go` — repair the vacuous oracle; add the equality,
  twin and no-fourth-cause tests.
- `folio-go/canvas_window_count_template.go` — the grouped count discriminator and its
  untagged twin.
- `folio-go/component_commands.go` — `duplicateComponent` clears `clone.KeepTogether` to the
  zero `Presence` (DW-48 / D-7.7.10).
- `folio-go/component_commands_test.go` — assert the drop, the untouched original, and the
  serialized bytes.
- `folio-go/table_row_clip_test.go` — refresh the verbatim AD-14 quote in a comment.
- `fixtures/keep-together/README.md` — rewrite `## What it does not cover`; digest string
  untouched.
- `_bmad-output/planning-artifacts/.../ARCHITECTURE-SPINE.md` — widen AD-14's carve-out
  (DW-49(a) / D-7.7.14 half (a)); no discriminator clause; file still 722 lines.

**Review findings breakdown:** patches applied 0, items deferred 0, items rejected 8 — the
intent_gap cascade makes the 10 `patch` and 4 `defer` findings moot for this pass, so none was
applied and `deferred:` was left as it was. All of them are enumerated in the Review Triage Log
so the re-dispatch loses nothing.

**Follow-up review recommendation:** `false`. Only findings triaged `patch` count, and none was
patched in this pass (patched high 0, medium 0, low 0; score 0).

**Verification performed** — the full D-R7.1 set ran against `6a31a7f`:
- `cd folio-go && go test -count=1 ./...` — 18 packages `ok`; exactly ONE red, the mandated
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (got 7, need >=20).
- `go vet -tags=matrix ./...` — exit 0, no output. `gofmt -l folio-go` from the repo root — no output.
- `TestTargetRenderHash` — four legs, each with `FOLIO_MATRIX_TARGET` exported and each PASS:
  `darwin/arm64` 0.71s, `linux/amd64` 6.08s, `linux/arm64` 4.63s, `js/wasm` 10.48s; every leg
  logged its witness line. UNSET control PASS in 0.00s and printed "this test asserts NOTHING
  and is a deliberate no-op".
- `TestCrossTargetByteIdentity` — PASS (21.98s).
- `cd lint && go test -count=1 ./...` — 2 RED, both reproduced byte-identically at `26bfd49`:
  `TestFloatTypedTestScopeInventory` and `TestStageRankProductionScan`. Pre-existing, verified
  at the baseline, not caused here.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile`
  — typecheck exit 0; lint exactly 4 pre-existing `only-export-components` warnings (the designer
  is absent from this diff entirely); vitest 284 passed across 34 files; e2e compile exit 0.
- **All 22 `goldenDigestRecord` digests byte-identical to baseline** (`diff` of the before/after
  `shasum -a 256` lists is empty). The spec's "21" is stale: `thai-stacked-marks` was added by
  Story 8.0 at `26bfd49`. All 22 also match the digests recorded in `byte_neutrality_test.go`.
- `fixtures/statement-signoff.json` untouched. Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`,
  unchanged, and last touched by commit `58ee9f7`.

**Mutation proofs (run independently, not accepted from the implementer).** Each file was
restored and `cmp`-verified byte-identical afterwards:
- Non-text arm alone neutralised — RED: "canvas counts 2 windows and the render path 3",
  origins `[0 740000]` against `[0 700000 1440000]`.
- Text arm alone neutralised — RED on BOTH subtests, count and origins.
- **Oracle reverted to its `nil` fifth argument with production left correct — RED**: canvas 3 /
  `[0 700000 1440000]` against the crippled oracle's 2 / `[0 740000]`. The oracle repair is
  load-bearing and non-vacuous.
- `clone.KeepTogether` clear removed — RED: "the duplicate carries keepTogether
  Presence[string]{Set:true, Null:false, Value:\"signature\"}".

**Residual risks**
- The two intent_gap classes above are live at `6a31a7f`: a `visibleIf` group member and an
  unstyled non-text group member both produce a confidently-exact wrong canvas answer.
- `ContentWindowCountIsFloor` still has exactly three write sites (`page_setup.go:701`, `:711`,
  `:716`) and its three-cause doc comment is unmodified — but that comment is now *incomplete*
  rather than wrong, which is the substance of the ruling being requested.

### Dispatch 3 — 2026-08-31, implement / review / triage (the three rulings)

Status: `done`
Blocking condition: none
Baseline revision: `ddd6a3ffdcb3f39bc4a91bc4210bd5f5b4e6b94c` (`main`)
Commits: `36e4718c31207b7bc7f16b5702d01fddf133c5c9` (implementation) and the finalize commit carrying
the review patches and this record. **`6a31a7f` was retained and built forward on, never reverted.**

**Summary.** The halt is resolved. RULING A2 renames `ContentWindowCountIsFloor` to
`ContentWindowCountIsExact` and **inverts its sense**, so the zero value `false` reads *"do not trust
this count"* and a forgotten set degrades to the honest claim; one Go field, its TypeScript mirror and
one UI string, no enum. RULING A registers **conditional visibility** as a cause named for what it is,
recorded as applying to an ungrouped `visibleIf` element exactly as much as a grouped one and as
undisclosed since Story 7.5. RULING B fixes the unstyled non-text element rather than disclosing it, by
giving `element_box.go`'s placement rule a second caller instead of a second copy. RULING C's widened
subject is carried into the fixture README and AD-14.

**Files changed** (14 in `36e4718`, 8 in finalize; no path under `folio-go/internal/layout/`)
- `folio-go/page_setup.go` — the renamed and inverted field with its cause register and the
  direction-dropped note; `canvasElementIsPlaced`; `canvasContentBandHasConditionalVisibility`.
- `folio-go/element_box.go` — `elementBoxDeclaration` / `elementDeclaresBox`, the one authority.
- `folio-go/render.go`, `folio-go/table_render.go` — `imageDrawsItsAsset`, `tableDrawsColumns`
  extracted for the same one-authority-two-callers reason.
- `folio-go/canvas_window_count_template.go`, `canvas_window_count_test.go` — the four-row matrix, the
  three added rows, the table test, the twin disciplines.
- `folio-go/canvas_projection_wire_test.go` (new) — the Go↔TypeScript wire-key tripwire.
- `folio-designer/src/{engine-protocol,sheet-stack,App}.ts(x)` + four test files — the mirror and the
  one UI string.
- `fixtures/keep-together/README.md`, `ARCHITECTURE-SPINE.md`, `folio-go/table_row_clip_test.go`.

**Review findings breakdown:** patches applied **10** (high 1, medium 4, low 5); items deferred **3**;
items rejected **6**.

**Follow-up review recommendation: `true`** — one patched finding was `high` severity (the null-asset
image, which is the default state of every image dropped in the designer). Patched counts: high 1,
medium 4, low 5; the `high` alone sets the flag, and the score `3×4 + 1×5 = 17` also exceeds 5.

**Verification performed** — re-run independently by the parent after the patches, not accepted from a
subagent:
- `cd folio-go && go test -count=1 ./...` — **13 packages `ok`, 4 `[no test files]`, exactly ONE red**:
  the mandated `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, "floor not met: got 7, need >=20".
- `go vet -tags=matrix ./...` exit 0, no output. `gofmt -l folio-go` **from the repo root** — no output.
- `TestTargetRenderHash` four legs, each with `FOLIO_MATRIX_TARGET` exported: `darwin/arm64` ok 1.11s,
  `linux/amd64` ok 6.51s, `linux/arm64` ok 5.03s, `js/wasm` ok 10.99s. **UNSET control** ok 0.39s,
  printing "this test asserts NOTHING and is a deliberate no-op" — which is what proves the four legs
  were not no-ops. `TestCrossTargetByteIdentity` ok 22.71s.
- `cd lint && go test -count=1 ./...` — **4 packages `ok`**. `-count=1` is mandatory here (D-7.9.5); the
  two reds that record names were fixed at `de7b126` and do not reproduce.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` —
  typecheck 0; lint **exactly 4** pre-existing `only-export-components` warnings; vitest **284 passed
  across 34 files**; e2e compile 0.
- **All 22 `goldenDigestRecord` digests re-hashed with `shasum -a 256`: 22 identical, 0 moved.** The
  contract's "21" predates Story 8.0's `thai-stacked-marks`.
- `fixtures/statement-signoff.json` unmodified. Root `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`,
  unchanged and never staged.

**The four rows, measured canvas against a REAL `buildPageModel` render:**

| document | canvas | real render |
|---|---|---|
| untagged + styled | 2, origins `[0 740000]` | 2 pages, origins `[0 740000]` |
| untagged + unstyled | 2, origins `[0 1440000]` | 2 pages, origins `[0 1440000]` |
| tagged + styled | 3, origins `[0 700000 1440000]` | 3 pages, origins `[0 700000 1440000]` |
| tagged + unstyled | 2, origins `[0 1440000]` | 2 pages, origins `[0 1440000]` |

⚠ **The dispatch's expected origins for `untagged + unstyled` were `[0 740000]`; the measured value is
`[0 1440000]` on BOTH sides.** `[0 740000]` was the **pre-fix canvas** value — the wrong answer RULING B
removes — transcribed from D-7.9.2's `canvas @6a31a7f` column rather than from the render. All four
**counts** match the dispatch exactly, and every row's origins now equal the render's.

**Mutation proofs, run by the parent (each landed-check asserted by occurrence count, each file
restored and `cmp`-verified):**

| mutant | result |
|---|---|
| `exact` forced **false** (documents that should be exact) | RED — all 4 matrix rows "…is false for a document carrying no registered cause", plus gap, empty column, windows-down |
| `exact` forced **true** (document carrying a registered cause) | RED — "bound table: count 1, origins [0], exact true; want 1, [0] and false" |
| cause (d) neutralised alone | RED — "conditional visibility: …is true, but whether one content element is placed at all is a property of data the canvas has never been given" |
| placement gate removed | RED — "the canvas draws 3 sheets and the document prints 2 pages" (reproduces D-7.9.2 exactly) |
| non-text tagging arm alone | RED — canvas 2 vs render 3, `[0 740000]` vs `[0 700000 1440000]` |
| text tagging arm alone | RED — both the count and the origins subtests |
| oracle's 5th arg back to `nil` | RED — canvas 3 vs the crippled oracle's 2 |
| `Canvas`'s hardcoded `false` → `true` | RED — "the shapeless entry point cannot count and must not claim it can" |
| `clone.KeepTogether` clear removed | RED — "the duplicate carries keepTogether …Value:\"signature\"" |
| image arm → unconditional `true` | RED — canvas 2 vs render 1, plus "an image with a file and an image without one both occupy 2 windows" |
| table arm → unconditional `true` | RED — canvas 2 vs render 1, plus the discriminator |
| positivity clause dropped | RED — "styled_rect_with_no_height: canvas count 2, want 1" |
| wire key renamed on **Go** only | RED — "CanvasProjection marshals the keys … and the recorded protocol set is …" |
| wire key renamed on **TypeScript** only | RED — "the designer's isCanvas guard accepts the keys … one side of this seam has been renamed and the other has not" |

⚠ Two mutants were rejected as invalid before being counted: forcing the field at its assignment
produced a *compile* error ("declared and not used"), not a test failure, and a three-tab pattern for
the non-text arm matched **both** arms. Both were re-run correctly. An unexpectedly-green or
unexpectedly-broken mutant is a reason to prove the mutation landed, not a conclusion about coverage.

**Residual risks**
- The **styled-TEXT** divergence is live and deliberately unclosed — see `deferred:`. The comment that
  claimed the parity invariant unconditionally has been corrected, so nothing in-tree overstates it.
- `canvas_projection_wire_test.go` reads `folio-designer/src/engine-protocol.ts` by walking to the repo
  root; the `folio-go` package suite now fails rather than skips if the designer tree is absent. That is
  the intended strictness for a cross-language seam, but it is a new coupling.
- Cause (d) is a conservative over-disclosure: any content-band `visibleIf` marks the count
  untrustworthy, including where the element could not move a window boundary. Intended per D-7.9.6.
