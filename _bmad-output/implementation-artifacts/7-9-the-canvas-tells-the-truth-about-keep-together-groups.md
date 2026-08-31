---
title: 'Story 7.9: The canvas tells the truth about keep-together groups'
type: 'bugfix'
created: '2026-08-31'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
warnings: ['oversized']
deferred: []
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

**Execution:**

- `folio-go/page_setup.go` -- build the keep-together index once from the `*Template` already in scope
  and tag **both** column-item literals with it: the non-text arm at `:647-657` and the text-line arm at
  `:948-953` inside `addCanvasTextPaint`. Use the existing `keepTogetherGroup` / `orKeepTogether`
  methods; construct no `layout.ItemGroup` literal in this file (the AST tripwire). -- This is the whole
  defect: untagged items are why count and origins are wrong, and the tags are the only missing input.
- `folio-go/canvas_window_count_test.go` -- **fix the vacuous oracle**: `renderPathWindows` at `:46` must
  pass the real `keepTogetherTags(tpl)` instead of `nil`, and must return the window **origins** as well
  as the count, so an origins oracle exists at all. -- An oracle that disables grouping on the render
  side cannot detect a grouping divergence; this is the absence that let the defect ship.
- `folio-go/canvas_window_count_template.go` -- add a grouped template const in the established idiom
  (body element, two-member group at y 700 / y 740, untagged tail at y 1440) **and its untagged twin**,
  differing only by the tags. -- The shipped fixture is an origins divergence; this const supplies the
  **count** divergence (3 vs 2), and the twin is what makes "differs only by the tags" checkable. This
  precedes the test task below, which consumes both consts.
- `folio-go/canvas_window_count_test.go` -- add the **count-and-origins equality** tests for a grouped
  document, asserting against a real render, plus the `no-new-floor-cause` assertion and the ungrouped-twin
  control. Use the grouped const for the count arm and `fixtures/keep-together/` for the origins arm.
  Red-proof each by reverting the tagging in `page_setup.go` and confirming they fail. -- The AC is an
  equality with the render path, so the test must observe both sides rather than the flag.
- `folio-go/component_commands.go` -- in `duplicateComponent`, clear the cloned tag immediately after
  `:1743` with `clone.KeepTogether = template.Presence[string]{}` (the zero `Presence`, **not** an explicit
  null). -- DW-48/D-7.7.10: a duplicate must not silently join a group the designer offers no way to see
  or clear.
- `folio-go/component_commands_test.go` -- assert the drop directly: duplicate a **content-band** element
  carrying `keepTogether`, assert the copy's `KeepTogether` is absent (`Set == false`) **and the original
  element is unchanged**. Read the template's `Element.KeepTogether` (or the serialized bytes), not the
  `CanvasProjection`, which carries no tag field. -- D-7.7.10 requires the drop be asserted, not
  incidental; nothing today asserts any field-copy behaviour.
- `fixtures/keep-together/README.md` -- rewrite the `## What it does not cover` section at `:60-64`,
  which becomes false when this lands, to state that the canvas now previews grouping and that its window
  count and origins agree with the render path. **Leave the recorded sha256 string untouched.** -- A
  fixture README that contradicts HEAD is the same defect DW-49(a) exists to fix.
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` --
  widen AD-14's over-tall carve-out at `:319` from `Over-tall rows (FR25)` to cover rows **and
  author-declared keep-together groups**, citing FR51 alongside FR25, describing what the engine does
  **today**. Prefer a rewrap that leaves the file's total line count unchanged. Add **no** discriminator
  clause. -- DW-49(a)/D-7.7.14: the sentence has been stale since `ed485eb`, and half (b) rides Story 7.10.

**Acceptance Criteria:**

- Given a template declaring a keep-together group, when the canvas projection is built, then its
  `ContentWindowCount` equals the page count of a **real render** of the same template, and its
  `ContentWindowOrigins` equals the render path's `plan.Pages[i].Shift` sequence computed with the real
  `keepTogetherTags` — asserted directly, not via `ContentWindowCountIsFloor`.
- Given the same template, when `ContentWindowCountIsFloor` is computed, then it is **false**, and
  `page_setup.go:359-375`'s enumeration of exactly three causes is still accurate — no fourth cause was
  added anywhere.
- Given an element of type `table` carrying `keepTogether` in either its tag or its explicit-null form,
  when the document is parsed, then it is refused with a located load error, and
  `TestKeepTogetherIsNotValidOnATable` is green and unmodified.
- Given a content-band element carrying a `keepTogether` tag, when `duplicateComponent` runs, then the
  copy's `KeepTogether` is absent and the original element is byte-for-byte unchanged.
- Given a template declaring no groups, when it is rendered and its canvas projected, then its bytes and
  every projected canvas value are identical to baseline, and all 21 `goldenDigestRecord` digests are
  unchanged.
- Given the tagging is reverted in `page_setup.go`, when the new equality tests run, then they **fail** —
  each new assertion is red-proved rather than assumed to bite.
- Given `folio-go/internal/layout/`, when the diff is inspected, then it contains **zero** paths under it.

## Spec Change Log

_Empty until the first `bad_spec` loopback._

## Review Triage Log

_Empty until the first review pass._

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
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test && npm run test:e2e:compile` --
  expected: typecheck clean; lint with **exactly 4** pre-existing `only-export-components` warnings (a
  fifth is a regression); vitest **284 tests across 34 files** as the arithmetic HEAD baseline — report
  the measured numbers, and note that no in-tree guard asserts them; e2e compile-only (D-000.4).
- `shasum -a 256` over all **21** `expected.pdf` paths named in `goldenDigestRecord`
  (`byte_neutrality_test.go:92-487`) -- expected: **every digest identical to baseline `c3df718`**. Any
  movement is the `golden digest moved` HALT, not a re-record.

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
