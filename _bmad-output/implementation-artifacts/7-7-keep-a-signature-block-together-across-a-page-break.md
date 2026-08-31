---
title: 'Keep a signature block together across a page break'
type: 'feature'
created: '2026-08-31'
status: 'done'
baseline_revision: '45b11c5be341fd29fefad37200e805d14cdf0bbb'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: [oversized]
deferred:
  - summary: >-
      The canvas reports an exact window count and window origins that a
      keep-together group can make wrong, while ContentWindowCountIsFloor
      still says the count is exact.
    evidence: |-
      addCanvasWindowCount (folio-go/page_setup.go:627-702) builds its
      layout.ColumnItems with no Group at all, and grouping is not among the
      floor causes. Measured during review: for a document with a body
      element, a two-member "signature" group at y 700 / y 740 and an
      untagged tail at y 1440, Render produces 3 pages while
      CanvasWithTextPaint reports ContentWindowCount = 2,
      ContentWindowCountIsFloor = false and ContentWindowOrigins
      [0, 740000] — the render's second window actually starts at 700000.
      No test reads either value for a tagged document. Not fixed here: the
      intent's Never clause forbids making the canvas group-aware, and the
      spec's "Limits to state" explicitly says do not add a floor cause and
      do not touch the projection. Needs a ruling or its own story.
    location: >-
      folio-go/page_setup.go:627-702
    severity: high
  - summary: >-
      An over-tall SINGLE-member keep-together group is clipped and warned,
      where the same element untagged is a fatal located OverflowError.
    evidence: |-
      Matrix row "Single-member group" says placement is identical to the
      same element untagged, while row "Group taller than one window" says a
      group is clipped with a Warning. Both rows apply to one over-tall
      tagged element, and they disagree: tagging converts D-2.6.1's located
      error into a clip-and-warn. The lead's D-4.6.2 amendment rules
      clip-and-warn for keep-together groups and a single-member group is
      still a group, so the shipped behaviour is defensible — but row 5's
      "identical to untagged" is false in this corner and should be tightened.
    location: >-
      folio-go/internal/layout/paginate.go:790-953 (clip branch), reached via
      folio-go/render.go keepTogetherGroup
    severity: medium
  - summary: >-
      duplicateComponent copies a keepTogether tag, silently adding a member
      to a group the designer offers no way to see or clear.
    evidence: |-
      duplicateComponent's clone := *element copies the whole Element,
      including KeepTogether. The designer has no grouping concept at all
      (component-command.ts has zero hits, and component_commands.go's
      property surface does not accept the key), so a group can only be
      authored by hand-editing the .folio file, yet duplicating a tagged
      signature element joins the copy to that group invisibly. The story
      records the canvas PREVIEW limit; this authoring/duplication limit is
      recorded nowhere.
    location: >-
      folio-go/component_commands.go (duplicateComponent)
    severity: medium
  - summary: >-
      ARCHITECTURE-SPINE.md still scopes the over-tall clip carve-out to
      "rows" although a second population is now clipped.
    evidence: |-
      ARCHITECTURE-SPINE.md:319 reads "Over-tall **rows** (FR25)". D-4.6.2 was
      amended in folio-mvp-decision-log.md for this story, but the spine that
      D-4.6.2 quotes as the authority for that noun was not updated. Amending
      the spine is the engineering lead's, not this story's.
    location: >-
      _bmad-output/planning-artifacts/ARCHITECTURE-SPINE.md:319
    severity: low
  - summary: >-
      A tagged multi-line text element becomes atomic, so the matrix's
      "single-member group changes nothing" holds only for single-line
      elements.
    evidence: |-
      contentColumnItems emits one ColumnItem per shaped line and the
      substitution stamps the same key on all of them, so tagging one
      multi-line element makes its lines unbreakable and its placement can
      change. TestKeepTogetherSingleMemberChangesNothing deliberately uses a
      single-line element and its own doc comment concedes the multi-line
      case "would be a real change". Arguably the feature working as
      intended, but the contract row is stated at the element level.
    location: >-
      folio-go/keep_together_fixture_test.go (TestKeepTogetherSingleMemberChangesNothing)
    severity: low
  - summary: >-
      No test combines non-contiguous membership with a union extent that
      crosses the window ceiling.
    evidence: |-
      TestKeepTogetherMembersNeedNotBeContiguous uses a group whose union
      FITS window one, so it exercises the ride-along but never the slide.
      The shipped fixture crosses the ceiling but is contiguous. The case
      where the window must slide to the group's earliest Top across an
      intervening ungrouped item — the two constraints together — is
      exercised by nothing.
    location: >-
      folio-go/keep_together_fixture_test.go
    severity: low
  - summary: >-
      asLoadError uses a bare type assertion while its comment claims
      errors.As semantics, so it fails on any wrapped LoadError.
    evidence: |-
      The body is err.(*LoadError), not errors.As, yet the same package wraps
      with fmt.Errorf("template: %s: %w", ...) in decodeBand/decodeElement.
      Pre-existing: the symbol does not appear in this story's diff. Either
      use errors.As or correct the comment.
    location: >-
      folio-go/internal/template
    severity: low
---

## In plain terms

*Non-normative. This section describes what shipped; the contract below governs implementation.*

A contract ends with a signature block — a name, a date, a ruled line to sign on — and a page break
could fall straight through it. A template author can now mark a set of content elements as
travelling together, and the engine treats that set as one indivisible unit: it all stays on the page
it started on, or it all moves to the next. The members need not be neighbours. Nothing else moves,
no page is left empty, and a template declaring no such set prints exactly the bytes it printed
before.

One case cannot be honoured: a set too tall to fit even an empty page. Rather than refuse the
document, the engine gives that set a page of its own, prints as much as fits, and returns a warning
beside the finished file — it is **clipped, not refused**. The cost was accepted deliberately:
whatever falls past the bottom is dropped, so a signature **image** inside such a set is **removed
from the document, not moved**.

Two limits are known and unfixed. The authoring canvas does not preview this grouping, so for a
document declaring one it can draw fewer pages than will print while calling that count exact; that
sits with the engineering lead. And a single element tagged alone can paginate differently from the
same element untagged.

One check here is required to stay failing — the text-corpus exercise floor. That red is expected and
must never be repaired.

<intent-contract>

## Intent

**Problem:** A signature block authored as several loose content-band elements is severed by a page boundary, because the paginator's only unit of atomicity is the individual line, image or table row. FR51 requires an author-declared set of content-band elements to paginate as one.

**Approach:** Add one optional, absent-by-default element key naming a keep-together group, and tag the affected column items with the *existing* `layout.ItemGroup` so the paginator's shipped grouping machinery — union extent, ride-along, the Story 4.6 over-tall clip — carries the whole feature. `internal/layout` gains nothing and changes nothing.

## Boundaries & Constraints

**Always:**
- `internal/layout` is READ-ONLY. `internal/layout/paginate.go` must be **absent from `git diff --name-only`**. D-2.6.1's window model and the four rules are inputs, not targets: no element's declared column position changes, no sibling moves, no band reflows, no page-count closed form appears, and the window still advances to the first unplaced item.
- Reuse `layout.ItemGroup` exactly as shipped. `ItemGroupKey`'s three fields are sufficient — two items are one group iff their `Key` compares equal (`paginate.go:167`), and the paginator requires **no contiguity** (`paginate.go:619-635` records the contiguity premise as measured false). No second grouping model, no new layout type, no key extension.
- A keep-together group's `Key.ElementID` must lie in a namespace that **cannot** be a valid element id, and `Key.IsHeader` must be false. This is what provably keeps the group out of every FR26 header-repeat, row-displacement, header-suppression and footer-orphan path.
- Every field this epic adds is optional and absent-by-default; a template declaring no groups renders byte-identically.
- Any invariant duplicated across the Go/TS boundary moves in one commit, with a test reading both sides.

**Block If:**
- Honouring a group would require any edit to `internal/layout` -> HALT, blocking condition `window model change`.
- The nine digests or any of the twenty `goldenDigestRecord` entries move for a template declaring no groups -> HALT, blocking condition `AC4 byte-identity lost`.
- **The engineering lead has not recorded the D-4.6.2 amendment in the decision log at the time of commit** -> HALT, blocking condition `D-4.6.2 amendment not recorded`. This story deliberately widens a ratified carve-out; landing the code while the log still says the carve-out is table-only would amend architecture silently, which is exactly what D-4.6.2's tripwire exists to prevent.
- The lead overturns the diagnostic-code ruling recorded in Design Notes (Flag 2) -> re-plan; do not mint a code unattended.

**Never:**
- Never mint a diagnostic code in this story (see Flag 2 — the dissent is recorded, the decision is the lead's).
- Never make the canvas group-aware, and never claim the canvas previews grouping. Never touch `hitTestBand`, the sheet stack, or `MAX_CANVAS_SHEETS`.
- Never close DW-26/27/28/30/31/32/33/34/35/37/38-45. DW-33 and DW-38 need rulings, not patches. DW-35 is Epic 8's. DW-29 is Story 7.8's.
- Never move `SupportedMajor` (stays 2) or `SupportedVersion` (stays "2.0"); never regress `SupportedVersion`'s doc comment (D-1.4.13).
- Never modify, move, delete or commit `README.md` (md5 `078d7d80d518d54af2fc04fb270d46b8`). Never `git add -A`. Never branch, never push.
- Never repair `TestCorpusMeetsP6ExerciseFloors` or the P6g floor.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Group fits below the window start | Three content-band text/image elements sharing one tag; union extent within the current window | All three placed on the same page, each at its own declared column position; nothing shifts | No error expected |
| Group crosses the window ceiling | Same three, union extent crossing the ceiling | All three move whole to the next window; the window starts at the group's **earliest Top** | No error expected |
| Group taller than one window **in aggregate** | Two or more elements, each fitting a window, union extent > content window height | A page of its own, clipped at that page's content bottom, one `TABLE_ROW_CLIPPED_HEIGHT` **Warning** beside the PDF bytes, located at the group's first member, message naming the keep-together group (not a row) | Warning, never fatal, never silent |
| Group holding an **individually** over-tall element | Any tagged element whose own extent > content window height | Located fatal `OverflowError` naming that element — refused, no bytes (Story 7.10, D-7.10.1) | Error, never a Warning |
| No group declared | Any existing corpus template | Byte-identical bytes; all twenty golden digests unchanged; saved `version` unchanged | No error expected |
| Single-member group **that fits a window** | One element tagged, its own extent within one content window | Placement identical to the same element untagged | No error expected |
| Single-member group **that does not** | One element tagged, its own extent > content window height | Located fatal `OverflowError` naming it (Story 7.10, D-7.10.1) — the "no-op" reading was false for a multi-line text element, whose own lines ARE something to keep together | Error, never a Warning |
| Two distinct tags | Two groups in one content band | Independent; each moves whole on its own; neither constrains the other | No error expected |
| Tag on a non-content-band element | The key on a `pageHeader` or `pageFooter` element | Refused at load, naming the element and the field | Located load error |
| Tag on a table element | The key on `type: "table"` | Refused at load, naming the element and the field | Located load error (a table's items already carry a row key; honouring both would be a second grouping model, which AC2 forbids) |
| Tag present anywhere | Document declaring the key | Saved `version` is `"1.2"`; a document not declaring it is unchanged; a document also using `align: "justify"` still saves `"2.0"` | No error expected |
| Tag absent / JSON null | Key absent, or explicitly `null` | Ungrouped; zero-value `ItemGroup`; bytes unchanged | No error expected |

</intent-contract>
## Gate rulings applied — 2026-08-31

Full text in `epic-7-8-decision-log.md` as **D-7.7.1 … D-7.7.5**. Settled; do not re-open.

**D-4.6.2 IS NOW AMENDED — the `Block If` on `D-4.6.2 amendment not recorded` is DISCHARGED.**
`folio-mvp-decision-log.md`'s D-4.6.2 entry carries the widening: an over-tall **keep-together group**
takes the Story 4.6 exception unchanged — a page of its own, clipped, a **Warning beside the bytes**,
never fatal. The tripwire fired exactly as designed and this is the second of the two arms it named.
It is **applying the criterion, not overriding a reservation**: D-4.6.2's ratio is *leniency follows
authorship*, and a keep-together group is author-declared. The grounding needs no analogy — AD-14 makes
**clipped content (FR44)** a Warning returned alongside bytes independently of its "rows" noun.
**Consciously accepted:** the clip drops whole members' runs and images, so a signature **image** inside
an over-tall group is **removed, not moved**. **No new diagnostic code** — `TABLE_ROW_CLIPPED_HEIGHT`
gains a fourth role arm, which is **not optional** (without it a signature block is announced as "row 0
of the bound collection"). Anything other than clip-and-warn returns to the lead.

**D-7.7.1 — the fence is the invariant, not the filename.** File-absence was always a proxy for "the
model must not change" (D-7.1.3, proven by 7.2's AC5 requiring page breaks to move *with
`internal/layout` unchanged*). At 7.7 the proxy and the purpose come apart. **The fence is: the four
pagination rules, window advance to the first unplaced item, no page is ever empty, one `Shift` per
page, and the column is never mutated — and `TestPaginateNeverProducesAnEmptyPage` plus the four-rule
tests stay green AND UNMODIFIED.**

**The co-extensiveness audit is mandatory and has measured teeth.** `ItemGroupKey` carries `IsHeader` and
an `ElementID` meaning *the table*. Four sites assume it — `paginate.go:833` (`!it.Group.Key.IsHeader`),
`:839` (`tbl := it.Group.Key.ElementID`), `:949-950` (`headerPageOf[...]`), `:960-962` (`headerExtent(...)`)
— and `:783`'s comment ("Keyed on `Group.Present`, and NOT on the item's kind") is the decision this story
puts under load. **Enumerate all of them in the record. A group that silently acquires header semantics is
the failure mode.**

**D-7.7.2 — the version rank is `1.2`, its own new rank.** Confirmed against the plan gate's independent
conclusion. **Not 1.1** (a 1.1 reader predates the key and silently splits the block — a version claiming
a reader sufficient for content it cannot render). **Not 2.0** (additive; a 1.x reader loads and ignores
it, so 2.0 overstates and needlessly orphans). `style.color` is the precedent: an additive key whose
absence renders *wrong* is still MINOR. `SupportedMajor` stays **2**.
**Guardrail:** `versionForRank` is `[...]string` indexed by an `iota` rank, so inserting `1.2` renumbers
`rankMajorFeature`. **Assert `versionForRank` is strictly ascending by parsed version** — the anchor that
table lacks. The hand-enumerated lists at `linespacing_test.go:229` and `:246` go vacuous otherwise.

**Not this story's:** DW-38 → **Epic 14.2** (engine-owned per-kind defaults, projected — *a projection is
not a mirror*). DW-42 → a named **Epic 15** story before the tag (a build-failing inventory witness, not
29 tests). Neither is taken here.


## Code Map

**Every anchor below was verified by reading the file at the baseline `ae82752`.**

### `folio-go/internal/layout/paginate.go` -- READ-ONLY. It must be absent from the diff.

The whole feature is already here. Nothing needs adding.

- **`ItemGroup` `:157-168`** -- exactly two fields: `Present bool` `:161`, `Key ItemGroupKey` `:167`. `Key`'s doc `:167`: *"Two grouped items with equal Key belong to the same group."* **`ItemGroupKey` `:175-179`** -- `{ElementID string; IsHeader bool; Index int}`.
- **`ColumnItem.Group` doc `:129-142`** -- *"ORTHOGONAL to the Runs/Images/Rects exclusivity check ... a SEPARATE statement about which items must land on the same page."* So a group may span a `Runs` item, an `Images` item and a `Rects` item; exclusivity is **per item**, confirmed at `:567-584` where the counter resets per item and `Group` is never read.
- **The three lines that ARE the feature.** Union pre-pass `:643-659` (`groups map[ItemGroupKey]*groupExtent`, one unordered scan); ride-along `:743-751` (`if p, ok := groupPage[it.Group.Key]; ok { pageOf[idx] = p; continue }`); effective extent `:762-768` (`effectiveTop, effectiveBottom = e.top, e.bottom`). The slide at `:986-1002` then reads `effectiveTop`, which is the group's earliest Top.
- ⚠ **THE LOAD-BEARING FACT.** `paginate.go:619-635` records R7's contiguity premise as **measured FALSE** and removed: *"a group's members are contiguous in column order by construction — is FALSE for real templates."* **A keep-together group of loose, non-adjacent elements therefore needs no new mechanism and no key extension.** (Story 4.3's review prose at `4-3-...:1402` says the key "groups exactly one row and never a run of rows"; re-measured at this baseline that is a statement about the *values package folio produces*, not about what the type can express. Do not plan from that sentence.)
- `internal/layout` **attaches no meaning to `Index`** beyond equality -- `:372-374`.
- Story 4.6 clip branch `:790-953`, entered by `if it.Group.Present { if effectiveBottom-effectiveTop > height {`. Its own page advance `:794-800` duplicates `:995-1001` deliberately. Records `TableRowClipped` at `:935-941`. ⚠ The cut at `:900-931` sets `dropped[j] = true`, and emission `:1137-1144` omits **both `Runs` and `Images`** of dropped members -- an image inside an over-tall group is removed, not moved.
- Ungrouped over-tall -> `*OverflowError` `:1075-1090`; comment `:1066-1074`: *"a grouped item took the clip branch above and never reaches here."*
- R6 guard `:1101-1120` -- key-agnostic; fires only if `groupPage[Key] != pageOf[idx]`. No table assumption.
- `Pagination.Clipped []TableRowClipped` `:354`; `TableRowClipped` `:357-381` carries both `ElementID` (the real element) and `Key`.

### The paths a non-table group must provably avoid -- verified gates

- **`headerExtent` `:675-682`** -- exact-struct lookup of `{ElementID: table, IsHeader: true, Index: 0}`. **Gate B `:956-964`**: `table` stays `""` unless `Group.Present && !IsHeader && headerExtent(Key.ElementID)` hits. `table == ""` disables `ceilingFor`'s narrowing `:972-979` and the whole reservation block `:1008-1062`.
- **Clip-branch repeat `:832-889`** -- guarded by `!IsHeader` **and** `hasHeader` **and** `headerPlaced`. `RowDisplacement` produced only at `:866-867` and `:1026-1027`; `Suppressed` only at `:880-886` and `:1045-1051`, and `TableHeaderSuppressed` carries **no Key**, so a spurious entry would be indistinguishable at the caller.
- **`footerOrphanTargetsFrom` `folio-go/table_footer.go:112-149`** -- skips `IsHeader`; `hasFooter[id]` requires `Index == footerGroupIndex` (`:68`, `-1`). `applyFooterMerge` `:346-362` rewrites only keys present in `mergeKey`, built solely from targets' `footerKey`.
- ⚠ **The collision consequence, measured:** a keep-together `Key.ElementID` equal to a real table's id would let the group *decide* FR26 reservation for that table's rows, emit a `TableHeaderRepeat` whose `Shift` is computed from the group's own top, and emit a `TableHeaderSuppressed` quoting an unrelated height. **The namespace rule in Boundaries is what makes all of this unreachable, not a comment.**
- **`validateElementID` `folio-go/internal/template/ids.go:46-69`** -- grammar `^e[0-9a-z]+$`, lowercase base-36, no leading zero, decoded counter >= 1. Any string containing a character outside `[0-9a-z]` (e.g. `:`) is therefore impossible as an element id, and enforced document-wide at `parse.go:34-46`.

### The tripwire this story is required to trip -- `folio-go/table_row_clip_test.go`

- **`TestAPresentItemGroupIsAlwaysATableRow` `:1062-1168`**, doc `:1042-1061`, banner `:1040`. An **AST scan**, not a behavioural test: `os.ReadDir(".")` `:1063-1069`, skipping dirs / non-`.go` / `_test.go` `:1084-1088`. Whitelist `:1076-1079` = `{"chromeRowGroup", "lineRowGroup"}`. Two spellings caught: a `.Present` **assignment** `:1111-1127` and a `layout.ItemGroup{...}` **composite literal** with `Present` set to anything but literal `false` `:1129-1141`. Vacuity floor `:1155-1168`, `const derivationFloor = 2`. Helpers `isLayoutItemGroupType` `:1173-1180` (matches only the spelling `layout.ItemGroup`), `setsPresentTrue` `:1194+`.
- Its own error text names the two lawful arms: *"Either give the new grouping its own placement rule, or take the decision to widen the clip deliberately and update D-4.6.2."*
- ⚠ **Three escape hatches that would leave it green and MUST NOT be used:** the scan is root-directory only (anything under `internal/**` is invisible); it matches only the literal `layout.ItemGroup` selector (a package alias hides it); it is name-based (a helper called *by* a whitelisted function is invisible). Tripping it honestly and widening it deliberately is the story.

### Where the tagging goes -- package folio, root

- **`paginateDocument` `folio-go/render.go:1974`**, appending **rects -> text -> images**. Rects `:2005-2036` (`Group: ts.chromeRowGroup()` `:2033`); text, **one item per shaped line**, `:2043-2097` (`Group: runs[i].lineRowGroup()` `:2086`); images `:2103-2126` -- ⚠ **images set no `Group` at all** (`:2119-2125`).
- **`contentColumnItems` `folio-go/page_number.go:87`** -- the PHASE A page-count pass, appending **text -> images -> rects** (the reverse order, stated at `:44-48`). `Group` at `:99` and `:151`; images `:112-124` ungrouped. **Both passes must agree or the page-count and the render disagree.**
- The two existing derivations: `tableRectSource.chromeRowGroup` `folio-go/table_render.go:243-256` and `textRunSource.lineRowGroup` `folio-go/render.go:233-255`. Both return `layout.ItemGroup{}` in their `default` arm -- **that zero return is the substitution point**: a tagged element's item is ungrouped today, so filling only the not-`Present` case leaves every table row untouched and is what preserves AC4.
- Element boxes travel the **rect** path: `collectElementBoxRects` `folio-go/element_box.go:53-56`, deliberately zero-group at `:103-107`. `declaredBox` `:126-136`.
- Extents: text `render.go:930-938` carried, never re-derived; image `render.go:2121-2122`; box `element_box.go:100-101`. Band translation is `layout.PlaceInBand` `internal/layout/band.go:121-123`; content band index `render.go:1946`, hard guard `render.go:2065-2067`.
- Build order is **element declaration order within a band** (`render.go:715`, `table_render.go:602`, `element_box.go:56`); pagination order is **stable sort by `Top`** (`paginate.go:585-604`). Neither needs changing.

### The diagnostic surface

- `clippedRowDiagnostic` **`folio-go/render.go:2291-2313`**, called unconditionally for every `plan.Clipped` entry at `:2157-2160`. The role switch `:2296-2304` has three arms: `IsHeader` -> `"the header row"`; `Index == footerGroupIndex` -> `"the footer row"`; default -> `fmt.Sprintf("row %d of the bound collection", c.Key.Index)`. Message `:2310` ends *"Reduce this row's height (font size or cell padding)"*. ⚠ **Without a fourth arm a signature block is announced as "row 0 of the bound collection".**
- `diag.CodeTableRowClippedHeight` `internal/diag/diag.go:220`, `allCodes :276`, `dispositions :310` (**Warning**), public alias `folio-go/diagnostic.go:314`. Census witness `diagnostic_registry_census_test.go:174-184`; pin `internal/diag/diag_test.go:44`. Public result type `Result{Bytes, Diagnostics}` `diagnostic.go:384-387`.
- Uncoded load errors: `newLoadError` `internal/template/errors.go:52` -> `TEMPLATE_MALFORMED` at `render_error.go:97-106`, and the designer **destroys the message** at `wasm/cmd/engine/main.go:276-279`.

### Format, version and the fences

- `Element` `internal/template/model.go:191-211`; optional-key slot beside `VisibleIf :198` / `Style :199`, **before** the kind-specific keys and `Extra :210`. `Presence[T]` `presence.go:17-21`, constructors `present :24`, `presentNull :29`, `absent :35`.
- Decode `decodeElement` `parse_bands.go:105`; `consumed` seeded `:124`; copy the `visibleIf` shape `:186-197`; **insertion point `:210`**, before the `switch el.Type` at `:212`. `bandField` (`:105`, values from `decodeBands :36/:40/:44`) is the existing hook for a band-scoped refusal; the column precedent for refusing a key in the wrong place is `:498-500`. Unknown keys -> `extraFields` `decodehelpers.go:116-135`.
- Serialize `writeElement` `serialize.go:224-289`; **insertion point after `:257`**; ⚠ **must use the unkeyed `{"key", …}` literal form** -- `drift_test.go`'s AST reader (`extractGoKeys :58-99`) is blind to a keyed literal, red-proved at `:459-485`. Sorting is `writeObject :38-42`; passthrough merges at `:287`.
- `version.go`: `SupportedMajor = 2` / `SupportedVersion = "2.0"` `:55-58` (doc `:20-54` -- **the discharged D-1.4.13 debt; do not regress**); `baseVersion`/`minorFeatureVersion`/`majorFeatureVersion` `:70-74`; `versionRank` `:230-236`; `versionForRank` `:241-245`; `versionRequiredByContent` `:196-219` (element loop `:201-217`, consults **only** `el.Style` and `el.Table.Value.HeaderStyle` -- an element-level key needs a **new probe**, `styleVersionRank :257-266` is the wrong site); `versionForSave` `:149-166`, sole production caller `serialize.go:118`.
- Version tests: `TestContentVersionNeverExceedsTheLibraryCeiling` `linespacing_test.go:221-257` -- ⚠ **hand-enumerates the constants at `:229` and document shapes at `:246`; it goes vacuous unless both are extended.** `TestVersionForSaveIsRaisedOnlyByContent` `:111-205` (table `:118-142`, ordering pair `:183-198`). `TestNewContentIsNeverStampedWithTheLibraryCeiling` `version_test.go:65-93` builds only through a **style** helper -- a new element-level doc builder is needed. `TestHigherMajorIsLoadError` `version_test.go:22-30` uses `"3.0"` and is **untouched** by a MINOR.
- Fences: `numericFieldRegistry` `numeric_classification_test.go:38-51` + floor `:123` -- ⚠ **a string-valued key is NOT captured** (the extractor regex `:74` matches only a documented key whose value is a number literal), so keep the documented value a string. `TestDriftGoToDoc :232-263` / `TestDriftDocToGo :269-300` -- bidirectional, **no allowlist**. `maximalFixture fixtures_test.go:236-408` (52 keys; canonical, insert in sort position) + `TestDriftASTMatchesRuntimeEmission drift_test.go:402-450`. `TestWorkedExampleMatchesGoldenFixture goldenfixture_test.go:16-25` -- ⚠ **only reddens if the doc's worked example is edited; document the key in the field table instead.**
- `folio-format.md` (`_bmad-output/specs/spec-folio/folio-format.md`): the "common to all five types" field table **`:213-219`** is the natural home; MINOR/MAJOR rules `:67-72`; the version sentence enumerating what raises which version is **`:47`** and is itself an edit site.

### Golden-fixture registration -- thirteen sites

`fixtures/<slug>/` (`input.folio`, `expected.json`, `expected.pdf`, `README.md`) plus: (1) `goldenDigestRecord` `folio-go/byte_neutrality_test.go:92-459` -- ⚠ its completeness half `:693-718` scans `goldenDigestSearchScope :482` and **fails on any undeclared occurrence**, so a README quoting the digest makes the `readme` site mandatory; (2) `declaredEpic2GateObligations` `:789-819`; (3) `matrixDocuments` `folio-go/matrix_test.go:1414-1788` + a capture func (pattern `:696-720`); (4-8) `.github/workflows/matrix.yml` -- **five edits**: `:73`, `:114`, `:155`, `:196` and the `docs="…"` list at `:255`, tied by `TestMatrixDocumentSlugsAreRegisteredInCI` `matrix_registration_test.go:39`; (9) `missing_glyph_corpus_test.go:41-169` fixtures table; (10) `beyondBaselineAcceptance :192-198` -- ⚠ the identity at `:235-237` fatals unless `len(fixtures) == len(baselineAcceptanceFixtures) + len(beyondBaselineAcceptance)`, and `baselineAcceptanceFixtures` is **frozen at 5** (`first_baseline_acceptance_test.go:100`, `:258-259`) -- **never add there**; (11) the Go template const + data const pair, byte-identical to `input.folio` (pattern `alignment_rounding_template.go:3`/`:85`, drift assert `alignment_rounding_fixture_test.go:289-302`); (12) `render_test.go` subprocess selector const (newest `:868`); (13) its `TestMain` arm (`:920-932`).

### Canvas -- unchanged, and it does not preview grouping

- `addCanvasWindowCount` `folio-go/page_setup.go:627`, its two item literals `:647-657` and `:948-953` -- ⚠ **neither sets `Group`**, so the canvas paginates ungrouped. `CanvasProjection` `:266`; `ContentWindowHeight :306`, `ContentWindowCount :331` (floor doctrine `:307-330`), `ContentWindowOrigins :358` and `ContentWindowCountIsFloor :376` (both 7.6). Floor causes computed at `:635` -- grouping is **not** among them. `atomicSpansFor(..., nil)` at `:880` (D-7.4.4). Three browser seams confirmed: `wasm/engine.go:119`, `:255`, `:294`.
- Tests to keep green: `canvas_window_count_test.go:64/:102/:127/:161/:190/:213/:270/:299/:352/:395` -- none of their fixtures declares a group, so all are unaffected.
- Designer: **no grouping concept exists** (`component-command.ts` zero hits). `PropertyField` `component-property-command.ts:3` (23 keys) mirrors `component_commands.go:839`/`:896`. **No test asserts the designer offers every format key**, and `unbreakableValues` is the shipped precedent: a document-level format key with zero designer references. `engine-ownership-contract.test.ts:63` forbids a browser-side schema; unknown keys survive a designer open/save through Go's `Extra`, not through TypeScript.

## Tasks & Acceptance

**The ordering is normative.** Parts 1-2 must land before Part 3 can tag anything; Part 5's fixture is the proof and must land in the same commit as Parts 1-4.

**Execution -- Part 1: the format key.**
- `folio-go/internal/template/model.go` -- add `KeepTogether Presence[string]` to `Element` beside `VisibleIf`/`Style` (`:198-199`) -- an element-level optional key needs no cross-reference and cannot dangle when its element is deleted.
- `folio-go/internal/template/parse_bands.go` -- decode at `:210` on the `visibleIf` pattern (`:186-197`), adding `"keepTogether"` to `consumed` at `:124`; refuse a non-empty tag when `bandField` is not the content band, and when `el.Type` is a table, as located load errors naming the element and the field -- FR51 scopes the feature to content-band elements, and AC2 forbids the second grouping model a tagged table would require.
- `folio-go/internal/template/serialize.go` -- emit after `:257` **in the unkeyed literal form**, or `TestDriftGoToDoc`'s AST reader cannot see it.

**Part 2: the version rank.**
- `folio-go/internal/template/version.go` -- add the constant `"1.2"` for this key's MINOR, one `versionRank` member **between** `rankMinorFeature` and `rankMajorFeature`, its `versionForRank` entry, and a **new element-level probe** inside `versionRequiredByContent`'s element loop (`:201-217`) -- `styleVersionRank` (`:257-266`) is the wrong site because the key is not a style key. `SupportedMajor` stays 2 and `SupportedVersion` stays `"2.0"` (2.0 already exceeds 1.2); do not touch the `:20-54` doc comment except to extend its history.
  **Why `"1.2"` and not `"1.1"`:** D-1.4.13 makes a document declare *the lowest version its own content requires*, and the shipped precedent fixes what "requires" means -- `lineSpacing` and `color` took 1.1 even though a 1.0 reader loads such a file and merely ignores them (D-1.4.9). By exact parallel, a 1.1 reader would load a keep-together document and silently ignore the grouping, so 1.1 would be a document claiming a fidelity it does not have. This is additive-optional-key MINOR either way (`folio-format.md:67-72`); it extends **no** closed set, so it is not a D-1.4.12 question.

**Part 3: tag the column items.**
- `folio-go/render.go` -- add **one** derivation function returning `layout.ItemGroup{Present: true, Key: layout.ItemGroupKey{ElementID: <namespaced tag>, IsHeader: false, Index: <keep-together sentinel>}}`, with the composite literal lexically inside it and nowhere else. Build the element-id -> tag lookup once from the document (lookups only, never a map range that reaches output order -- D-1.3.5/R5). In `paginateDocument`, substitute it **only where the existing group is not `Present`**, at the rect site `:2033`, the text site `:2086` and the image site `:2119-2125` -- filling only the ungrouped case is what leaves every table row and every untagged document byte-identical.
- `folio-go/page_number.go` -- apply the identical substitution in `contentColumnItems` at `:99`, `:151` and the image block `:112-124`, calling the same derivation -- PHASE A and PHASE B must agree or the page count and the render disagree.
- `folio-go/render.go` -- add a **fourth role arm** to `clippedRowDiagnostic`'s switch (`:2296-2304`) and a group-appropriate remedy clause in the message (`:2310`), recognising the keep-together key. The code stays `TABLE_ROW_CLIPPED_HEIGHT` (Flag 2).

**Part 4: widen the tripwire deliberately.**
- `folio-go/table_row_clip_test.go` -- add the new derivation's name to `rowGroupDerivations` `:1076-1079`, raise `derivationFloor` `:1155-1168` from 2 to 3, **rename the test** (its current name asserts a property this story makes false), and rewrite the doc comment `:1042-1061` to state the new invariant: a present `ItemGroup` is a table row **or** an author-declared keep-together group, and nothing else. Keep both spelling checks and the vacuity floor intact.
- `_bmad-output/specs/spec-folio/folio-format.md` -- document the key in the field table `:213-219` with a **string**-valued example (a number-valued example would make it the 13th `numericFieldRegistry` entry), extend the version sentence `:47`, and extend D-4.6.2's authorship sentence so *"Folio absorbs what the data made too tall, and refuses what the author typed too tall"* also states what an over-tall declared group does.

**Part 5: the discriminating fixture.**
- `fixtures/keep-together/` + `folio-go/keep_together_template.go` -- a content band whose body text fills most of window one, followed by a three-element signature block positioned so that **ungrouped it splits across the window boundary and grouped it moves whole to window two.** Ship the template const byte-identical to `input.folio`, **plus a second const identical except that the keep-together keys are absent** -- the ungrouped twin is the discriminator.
- `folio-go/keep_together_fixture_test.go` -- assert the difference, not the feature: render both consts and assert (a) grouped, every signature element's runs are on page 2 and page 1 carries none of them; (b) ungrouped, the first signature element is on page 1 and the others on page 2; (c) the two renders differ in bytes. Add the over-tall case asserting one `TABLE_ROW_CLIPPED_HEIGHT` Warning with non-empty bytes and a message that does **not** contain "of the bound collection". Add the union-extent/no-contiguity case: a group whose members are separated by an intervening ungrouped element.
- Register the fixture at all thirteen sites listed in the Code Map, and add it to `beyondBaselineAcceptance` with a story-naming reason -- **never** to `baselineAcceptanceFixtures`.
- Version tests -- extend `linespacing_test.go:229` and `:246`, add cases to `:118-142` including an ordering case proving the new element-level probe participates in the maximum, and add an element-level document builder for `version_test.go:65-93`.

**Acceptance Criteria:**
- Given a set of content-band elements declared one keep-together group, when the document paginates, then every member is placed within one window or entirely within the next, at its own declared column position, with no sibling moved.
- Given the shipped `ItemGroup` machinery, when a keep-together group paginates, then `git diff --name-only` contains **no path under `folio-go/internal/layout/`**, and no new grouping type, field or key exists anywhere.
- Given a group taller than one window, when it paginates, then the render returns non-empty PDF bytes plus exactly one `TABLE_ROW_CLIPPED_HEIGHT` Warning located at the group's first member, and never an error.
- Given any template declaring no group, when it renders, then all twenty `goldenDigestRecord` entries and the nine reported digests are byte-identical, and its saved `version` is unchanged.
- Given the keep-together fixture, when the grouping tags are removed from the template, then the rendered page assignment **changes** and the test asserting the grouped placement **fails** -- the fixture discriminates rather than passing under any derivation.
- Given the widened tripwire, when a present `ItemGroup` is constructed in package folio outside the three named derivations, then the build fails naming the function and position.

## Spec Change Log

## Review Triage Log

### 2026-08-31 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 7: (high 1, medium 2, low 4)
- defer: 7: (high 1, medium 2, low 4)
- reject: 8: (high 0, medium 0, low 8)
- addressed_findings:
  - `[high]` `[patch]` PHASE A's keep-together grouping was entirely unverified — mutating all three `contentColumnItems` substitutions away left the whole `folio-go` suite green, so `{{pages}}` could print the ungrouped page count. Added `TestKeepTogetherPageCountReachesTheRenderedFooter` (a document whose grouping changes the page count, with a `Page {{page}} of {{pages}}` footer) and `TestBothPaginationPassesAgreeWithAKeepTogetherGroup`. Independently re-mutated after the fix: the 3-page document's footers print "of 2" and both tests fail.
  - `[medium]` `[patch]` `"keepTogether": null` bypassed all three load refusals — verified empirically that a `pageHeader`, a `pageFooter` and a `table` element each carrying the null form loaded clean, contradicting the matrix rows ("the key ... Refused at load") and the prose this story shipped in `folio-format.md`. Moved the band and table refusals above the `rawIsNull` branch so the key is refused wherever it can never be honoured; `null` still means "ungrouped" where the key is allowed. Refusal tests now drive both key forms.
  - `[medium]` `[patch]` A tagged image was wired in both passes and promised specific behaviour in `folio-format.md` ("removed, not moved") but was exercised by no render test; deleting both image `Group:` lines left the suite green. Added `TestKeepTogetherCarriesATaggedImage` and `TestKeepTogetherOverTallGroupDropsATaggedImage`, both mutation-verified.
  - `[low]` `[patch]` `version.go:277` cited `TestVersionForRankIsStrictlyAscending` as living in `version_test.go` while it was in `linespacing_test.go`; moved the test to `version_test.go` unchanged so the reference is true.
  - `[low]` `[patch]` Stale key counts: `maximalFixture` now emits 55 keys (measured, not assumed) while `fixtures_test.go` said 52 and `drift_test.go` said 52. Both corrected.
  - `[low]` `[patch]` `const contentBandField` had been inserted between `decodeBand`'s doc comment and `decodeBand`, orphaning the comment onto the const. Const moved.
  - `[low]` `[patch]` The fixture's line-extent table was wrong and duplicated in two files: it used descent 244 instead of Noto Sans's −293 (per-line extent is 14.982 pt, not 14.443) and gave `e1`'s bottom as the last line's top. Re-measured from the built column items (`e1` 0.000 … 554.334, union 706.000 … 754.982) and the duplication collapsed to one copy in `fixtures/keep-together/README.md`.

Rejected (with the authority each was tested against): an explicit `null` raising the document to `1.2` — the matrix's "Tag present anywhere | Document declaring the key" row covers a declared null, so this is the contract's answer, not a defect; `contentBandField` naming only one of three band fields (cosmetic; the const is used only for the comparison it names); `consumed["keepTogether"]` set unconditionally (the key is known either way); the matrix guard's literal `" re f\n"` operator count (a structural page-model assertion sits beside it); whitespace-only and case-differing tags (a tag is an opaque author string — trimming or folding would surprise); no length bound on the tag value (never projected, so no surface beyond template size); `folio-format.md` filing the key under "common to all five types" while the row's own text states the table exception; `keepTogetherUngroupedTemplateJSON` declaring `"1.2"` with no key present (deliberate, and mechanically enforced by `TestKeepTogetherTwinDiffersOnlyByTheTags`). Sprint status and the Delivery Log were left to the story closer by dispatch.

## Design Notes

### Ruling A -- the key needs no extension, because the paginator was already general

The obvious reading of AC2 is that `ItemGroupKey{ElementID, IsHeader, Index}` "groups exactly one row and never a run of rows" and must therefore be extended. **Re-measured at this baseline, that is false.** `Key`'s own doc (`paginate.go:167`) defines a group as *equality of Key*, and R7's contiguity premise was measured false and **removed** (`:619-635`) precisely because real table rows are not contiguous in column order. N loose elements sharing one `Key` value already *are* one group, and the union pre-pass, the ride-along and the R6 guard are all key-agnostic. So AC2 is satisfied in its strongest form: `internal/layout` is not merely "not a second model" -- it is **untouched**, and that is assertable by `git diff --name-only` rather than argued.

### Ruling B -- the declaration is an element key, not a document-level id list

FR51 (`epics.md:109`) ratifies no shape: *"Declare a group of content-band elements that paginate together."* Both an element-level tag and a document-level array of id lists satisfy it, and `unbreakableValues` (`model.go:73`) is a real precedent for the latter. The element key wins on a stated principle rather than taste: **a document-level id list is a second place element ids appear, and something must prune it when a component is deleted.** The engine owns the document and the designer has no grouping concept, so that pruning would be new engine logic serving a referential integrity problem the element key does not have -- a tag is deleted with its element, automatically, and can never dangle. It also needs no existence or band validation of referenced ids; the band and type refusals are local to the element being decoded, where `bandField` already is.

### Ruling C -- the group key lives in a namespace no element id can occupy

`validateElementID` (`ids.go:46-69`) admits only `^e[0-9a-z]+$` with no leading zero and a decoded counter >= 1, enforced document-wide at `parse.go:34-46`. A group key whose `ElementID` contains any character outside `[0-9a-z]` is therefore **provably** never equal to a real element's id. With `IsHeader: false`, that single choice makes every table-shaped path unreachable at once: `headerExtent` cannot match, so Gate B leaves `table == ""` and `ceilingFor` is unnarrowed; no `HeaderRepeats`, `RowDisplacement` or `TableHeaderSuppressed` can be produced; `headerContentOf` cannot match; `footerOrphanTargetsFrom` builds no target and `applyFooterMerge` cannot re-key it. This is a correctness argument with a grammar behind it, not a naming convention -- **assert the namespace in a test**, because the day someone "tidies" the prefix, every one of those paths reopens silently.

### Flag 1 for the engineering lead -- D-4.6.2's leniency, and my reading

**My reading: AC3's extension holds, but only because the intent takes the decision D-4.6.2 explicitly reserved -- and the log must be amended to say so.**

D-4.6.2 grounded the clip carve-out on **two** reasons, and the lead preferred the second: AD-14's ratified noun is *"Over-tall **rows**"* (`ARCHITECTURE-SPINE.md:320-321`), and *"Arm B would extend it beyond what any invariant grants."* Its documentation obligation states the principle as authorship: *"Folio absorbs what the data made too tall, and refuses what the author typed too tall."*

**On that principle alone the extension fails.** I tested the tempting counter-argument -- that a group's union extent is *derived* rather than typed, like a row's -- and it does not survive: because `ColumnItem.Group` is one key per item, a keep-together group **cannot contain a table** (its rows already carry their row key), so every member's height is author-typed. The union is derived from typed inputs by an author-declared membership. Under D-4.6.2's own test, an over-tall keep-together group is *"what the author typed too tall"*, and D-2.6.5 says such an item is a **located error**, not a clip.

**What selects the other way is the intent, twice and unqualified:** AC3 says *"the Story 4.6 exception applies unchanged"*, and the epic's own constraints repeat it. D-4.6.2's tripwire names exactly two lawful arms -- *"give the new grouping its own placement rule, or take the decision to widen the clip deliberately and update D-4.6.2"* -- and a ratified AC that postdates D-4.6.2 and speaks to this precise population **is** that deliberate decision. I have therefore planned the widening rather than halting, and I did not conclude the extension is unsafe.

**What remains the lead's, and why `Block If` holds the commit:** the amendment itself. Landing Part 4's widened tripwire while the decision log still says the carve-out is table-only would amend architecture silently -- the one outcome D-4.6.2 built the tripwire to prevent. The lead should record the amendment (quoted before/after, in the same commit as the code), and confirm one consequence AC3 does not mention: **an over-tall group's clip DROPS whole members' runs and images** (`paginate.go:900-931`, `:1137-1144`), so a signature *image* inside such a group is removed from the document. That follows from "unchanged", but it deserves a conscious yes.

### Flag 2 for the engineering lead -- the diagnostic code, and my dissent

**Applied as dispatched: no code is minted, and `TABLE_ROW_CLIPPED_HEIGHT` is reused with a fourth role arm.** I am recording the dissent rather than acting on it.

D-4.5.1's discriminator: *"Two conditions share a code only if the author would take the same action AND the same thing happened to their document."* Applied here, the first limb arguably holds (content present and truncated) and **the second fails**: the remedy for an over-tall row is *"reduce this row's height (font size or cell padding), shorten the data"*, while the remedy for an over-tall keep-together group is *remove a member, or stop declaring the group* -- a different action on a different object. D-000.65 puts the mint in the story where the condition first ships, which is this one. On the architecture's own tests the answer is **mint**.

Against that: minting costs seven registration sites plus a census witness, and the dispatch settles it as the lead's call. Reuse is also not indefensible -- both conditions are "an atomic unit exceeded the window and was truncated", which is arguably one condition with a second site.

**The part that is not optional either way:** without the fourth role arm, an author's signature block is reported as *"row 0 of the bound collection"* with a remedy about cell padding. That defect is fixed by Part 3 regardless of how the lead rules; only the `Code` string is in question.

### Limits to state, not to fix

- **The canvas does not preview grouping, and this story must not claim it does.** The canvas builds its column items ungrouped (`page_setup.go:647-657`, `:948-953`), and grouping is not among the three floor causes at `:635`. A declared group can therefore move a break the canvas does not draw. **I have not proved grouping preserves the count's floor property in both directions** -- a group can pull a later-visited member *earlier* via the ride-along -- so this story neither claims nor tests that. Do not add a floor cause and do not touch the projection: making the canvas group-aware is a separate story, and DW-40/DW-41 already own the neighbouring honesty questions.
- **D-7.4.4 stands.** The canvas breaks the raw template string with `nil` substitutions (`page_setup.go:880`), so nothing here claims anything about where a bound value breaks.
- **DW-43's owner is this plan gate.** Its subject -- no shipped fixture draws an in-sheet seam -- is answered: the keep-together fixture is exactly that instrument, and Part 5 ships one whose column really breaks inside a sheet. **DW-43 is not closed** (the dispatch fences 38-45); amend its entry to name this fixture and leave it open for whoever confirms the canvas draws that seam.
- **DW-37/DW-40/DW-41/DW-44 are read and left open**; none is touched, because the projection is not modified. DW-32 and DW-38 are untouched because no designer control and no palette kind is added.
- **No Go/TS mirror obligation arises.** The key is never projected, so `engine-protocol.ts`'s `hasOnly` allow-lists (`:207`, `:244`) are unchanged and `engine-bounds-mirror.test.ts`'s six pairs stay six. Confirm this by grepping the diff for `folio-designer/src` and expecting nothing -- if the designer must change, the plan was wrong.

## Verification

This story changes pagination inputs and the file format, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). **Report measured pass/fail counts, never "green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` -- expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:169`, P6g subtest, floor `:184`, `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` (`:243`) must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` -- expected: clean, exit 0.
- `gofmt -l folio-go` -- run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` -- run **once per leg** with `FOLIO_MATRIX_TARGET` set: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` (`matrix_test.go:69-74`). ⚠ **Unset, this test logs "asserts NOTHING" and returns -- a no-op is not a pass.** Grep each leg for `asserts NOTHING` and report the count per leg; name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` -- expected: pass; all four targets from one process.
- `cd lint && go test ./...` -- expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` -- expected: pass, **280 tests / 33 files unchanged** (this story adds none). **oxlint baseline is exactly 4 `only-export-components` warnings and 0 errors**; a fifth is a regression.
- `cd folio-designer && npm run test:e2e:compile` -- expected: pass. ⚠ `tsc --noEmit` only. **Browser e2e is deferred by D-000.4 and does NOT execute -- say so rather than implying it ran.**

**Nine digests to report byte-identical** (all **twenty** `goldenDigestRecord` entries must hold): statement-1 76,744 `114df1d6…`; statement-5 127,363 `70dce051…`; statement-20 269,884 `56bfbbd9…`; statement-50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`. **This is AC4 -- assert, do not assume.** Confirm `git status fixtures/` shows only the new directory before quoting any digest.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` (`fontgen_matrix_test.go:64`) fails under `-tags=matrix` without `fontTools`; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23).

**Manual checks:**
- **Confirm `git diff --name-only` contains no path under `folio-go/internal/layout/`.** This is AC2's proof and it is mechanical.
- **Confirm `README.md` appears in no commit and its md5 is still `078d7d80d518d54af2fc04fb270d46b8`.** Stage explicit paths; never `git add -A`.
- **Red-proof each of these and record what failed:** (1) remove the grouping tags from the fixture template -- the grouped placement assertion must fail, not the ungrouped one; (2) revert the Part 3 substitution so no item is tagged -- the difference assertion must fail; (3) revert the fourth role arm -- the over-tall message assertion must fail on "of the bound collection"; (4) restore `derivationFloor` to 2 and remove one derivation name -- the tripwire must fire naming the function.
- **Matrix Test Audit:** for every I/O matrix row, grep the covering test for a call to the symbol under test. A test that re-derives a rule instead of calling it is a false green, not a nit.
- **Verify the namespace claim rather than asserting it in prose:** a test showing the group key's `ElementID` is rejected by `validateElementID`.
- **Demonstrate end to end** that a signature block authored across a window boundary prints whole on the next page in the real PDF -- not that a conditional changed.

## Auto Run Result

### 2026-08-31 — Implementation dispatch (status: done)

Status: done
Blocking condition: none
Dispatch: existing spec at `ready-for-dev`, routed straight to implement (step-01 early exit). Baseline `45b11c5be341fd29fefad37200e805d14cdf0bbb`; HEAD verified unmoved through both subagent dispatches.
D-4.6.2: the `Block If` was DISCHARGED before this dispatch — the amendment is recorded in `folio-mvp-decision-log.md` and landed in the baseline commit itself.

**Implemented change.** An author declares `keepTogether: "<tag>"` on content-band elements; the tagged items are stamped with the *existing* `layout.ItemGroup` so the shipped grouping machinery (union extent, ride-along, the Story 4.6 over-tall clip) carries the whole feature. `internal/layout` gained nothing and changed nothing — `git diff --name-only` contains **no path under `folio-go/internal/layout/`**, which is AC2's mechanical proof.

**Files changed.**
- `internal/template/model.go` — `Element.KeepTogether Presence[string]`.
- `internal/template/parse_bands.go` — decode on the `visibleIf` shape; refuses the key outside the content band and on a table (both regardless of value, after review), and an empty tag.
- `internal/template/serialize.go` — emitted in the unkeyed literal form the drift AST reader can see.
- `internal/template/version.go` — new `"1.2"` rank between minor and major; new **element-level** probe in `versionRequiredByContent`. `SupportedMajor` 2 and `SupportedVersion` "2.0" unmoved; D-1.4.13's doc comment extended, not regressed.
- `render.go` — the single `keepTogetherGroup` derivation, the `keepTogether:` namespace, the `-2` index sentinel, substitution at the rect/text/image sites, and `clippedRowDiagnostic`'s **fourth role arm** (no code minted).
- `page_number.go` — the identical substitution in PHASE A's `contentColumnItems`.
- `table_row_clip_test.go` — the D-4.6.2 tripwire widened deliberately and renamed `TestAPresentItemGroupIsATableRowOrAKeepTogetherGroup`; whitelist gains `keepTogetherGroup`, `derivationFloor` 2 -> 3.
- `fixtures/keep-together/` + `keep_together_template.go` + `keep_together_fixture_test.go` — the discriminating fixture and its tests; registered at all thirteen sites.
- `folio-format.md`, `deferred-work.md` (DW-43 amended to name this fixture and left OPEN).
- Five call-site ripples passing `nil` for the new parameter (`table_pagination_test.go`, `table_footer_test.go`, `canvas_window_count_test.go`, `canvas_body_text_bounds_test.go`, `collect_text_runs_composition_test.go`) — no assertion or expectation altered.

**The co-extensiveness audit.** All four table-assuming sites were enumerated and proved unreachable: `paginate.go:833` (`!it.Group.Key.IsHeader`), `:839` (`tbl := it.Group.Key.ElementID`), `:949-950` (`headerPageOf[...]`), `:960-962` / Gate B `:956-964` (`headerExtent`). The group key's `ElementID` carries a `:`, which `validateElementID` (`ids.go:46-69`, grammar `^e[0-9a-z]+$`) proves no element id can equal, and `IsHeader` is false. Unreachability is **asserted, not relied upon**: `TestKeepTogetherGroupKeyIsNotAValidElementID` feeds the key through `ParseDocument` as an element id and requires refusal (with a loading control), and `TestKeepTogetherReachesNoTablePath` drives an over-tall group through the clip branch and requires zero `HeaderRepeats`, zero `RowDisplacement`, zero `Suppressed` and no footer-orphan target.

**Review findings.** 7 patched (high 1, medium 2, low 4), 7 deferred (high 1, medium 2, low 4), 8 rejected, 0 intent_gap, 0 bad_spec. See the Review Triage Log. Follow-up review recommended: **true** (a `high` was patched).

**Verification performed** (measured, re-run by the parent after patching):
- `go test -count=1 ./...` — exactly ONE red, `TestCorpusMeetsP6ExerciseFloors/P6g` (`got 7, need >=20`), the mandated permanent red, untouched; drift twin `TestCorpusP6StatsMatchDeclaredBaseline` green. Every other package ok.
- `go vet -tags=matrix ./...` exit 0; `gofmt -l folio-go` from the repo root: no output; `cd lint && go test -count=1 ./...` 4/4 ok.
- `TestTargetRenderHash` once per leg with `FOLIO_MATRIX_TARGET` set — darwin/arm64, linux/amd64, linux/arm64, js/wasm — all PASS with `asserts NOTHING` count **0** on every leg. `TestCrossTargetByteIdentity` ok.
- Nine digests re-asserted byte-identical after patching: statement-1 76,744 `114df1d6…`; -5 127,363 `70dce051…`; -20 269,884 `56bfbbd9…`; -50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`.
- Designer: typecheck clean; oxlint 4 `only-export-components` warnings / 0 errors (baseline); 280 tests / 33 files unchanged; `test:e2e:compile` clean — **`tsc --noEmit` only; the browser e2e did NOT execute** (D-000.4).
- `README.md` untouched, md5 still `078d7d80d518d54af2fc04fb270d46b8`.

**Mutation proofs run by the parent** (a green suite over correct code proves nothing):
- Fourth role arm neutered -> the over-tall message reads "row **-2** of the bound collection" and the test fails.
- `keepTogetherGroup` returning the zero group -> 7 tests red, including the golden fixture and the byte-difference assertion.
- PHASE A substitutions removed -> **before patching, the entire suite stayed green** (the finding); **after patching**, the 3-page document's footers print "of 2" and both new tests fail.
- `derivationFloor` back to 2 with the whitelist entry removed -> the tripwire fires naming `render.go:2054:9: keepTogetherGroup`.
- `versionForRank` ranks swapped -> both halves of `TestVersionForRankIsStrictlyAscending` fire.
All mutations restored byte-identically (`cmp`-verified) and the suite re-confirmed.

**The discriminating fixture.** `fixtures/keep-together/` ships the template and a twin identical except the tags are absent (`TestKeepTogetherTwinDiffersOnlyByTheTags` enforces that "identical except"). Asserted difference: **grouped**, page 1 carries no member of the signature block (no text, 0 rects) and page 2 carries the signature line, the date line and the ruled line's rect; **ungrouped**, the signature line is stranded on page 1 while the rule and date fall to page 2; and the two renders **differ in bytes**. A derivation that ignored grouping produces the split and fails.

**Residual risks.**
- The canvas does not preview grouping, and for a grouped document its window count and origins can be wrong while `ContentWindowCountIsFloor` reports the count as exact (measured; deferred, high — the intent forbids the fix here).
- The tagged-image behaviour is now tested at the render surface, but the browser e2e witness remains compile-only, so no executed proof crosses the browser boundary.
- `goldenDigestRecord` is now 21 entries; the twenty pre-existing entries are unmoved (the file's diff is insertions only).

Status: ready-for-dev
Blocking condition: none
Dispatch: classic intent, Story 7.7, Epic 7. `Halt after planning.` honoured — spec written, no implementation code, no commits.
Baseline: `ae8275205b576e3f8cfebe128a1e25e486952122`, working tree clean at dispatch and at halt (this spec is the only untracked addition).
Epic context: `epic-7-context.md` cache was VALID (mtime 2026-08-30 23:54:50 > newest planning artifact 2026-08-30 23:25:04) and was loaded; not recompiled, so it is unmodified.
Continuity source: `7-6-the-canvas-draws-every-page-the-document-will-produce.md` (`status: done`).
Verification: none run — planning-only dispatch executes no build or test commands.

Two items are flagged for the engineering lead at the plan gate; both are ruled in the spec so it is implementable as written, and both are the lead's to overturn:
1. **D-4.6.2's leniency extended to author-declared groups** (Design Notes, Flag 1). Planned as holding, on AC3's explicit ratified wording, which is the deliberate decision D-4.6.2's own tripwire reserved. It is NOT supported by D-4.6.2's authorship rationale, and the log must be amended. `Block If` holds the commit until it is.
2. **The diagnostic code** (Design Notes, Flag 2). Applied as dispatched — no mint; `TABLE_ROW_CLIPPED_HEIGHT` is reused with a fourth role arm. Dissent recorded: D-4.5.1's second limb and D-000.65 both point at a mint.

## Delivery Log

### 2026-08-31 — planned

Baseline `ae82752`. Plan-only dispatch; no code, no commits. The story carried an `oversized`
warning and kept it: a format key, a version rank, two pagination passes, a deliberately widened
architecture tripwire and a thirteen-site golden fixture are one story only because the fixture is
what proves the rest.

Two items went to the engineering lead at the plan gate and both came back ruled, in
`epic-7-8-decision-log.md` as **D-7.7.1 … D-7.7.5**.

**D-4.6.2's tripwire fired exactly as designed, and this is the second of the two arms it named.**
The entry it guards had ruled the over-tall clip carve-out table-only, and this story widens it to
author-declared keep-together groups: a page of its own, clipped, a Warning beside the bytes, never
fatal. That is applying D-4.6.2's own criterion — leniency follows authorship — not overriding a
reservation, and AD-14 grounds it without analogy. The lead recorded the amendment in
`folio-mvp-decision-log.md` **in the baseline commit**, before any code, which is what discharged the
`Block If` on `D-4.6.2 amendment not recorded`. Consciously accepted in the same ruling: the clip
drops whole members, so a signature **image** inside an over-tall group is removed, not moved.

**The Epic 7 scope fence was re-anchored from a filename to an invariant.** "`paginate.go` absent
from the diff" was always a proxy for "the window model must not change". At 7.7 the proxy and the
purpose come apart, so the fence became the property itself: the four pagination rules, window
advance to the first unplaced item, no page ever empty, one `Shift` per page, the column never
mutated — with `TestPaginateNeverProducesAnEmptyPage` and the four-rule tests green **and
unmodified**. It happens that the proxy held too, which the close verified independently.

The version rank was settled at `1.2` — its own new rank, reached independently by the plan gate and
the lead. Not 1.1, which would claim a reader that silently splits the block; not 2.0, which is
additive and would needlessly orphan 1.x readers.

### 2026-08-31 — built

Baseline `45b11c5`; story commit `ed485eb`, 31 files, one commit. An author declares
`keepTogether: "<tag>"` on content-band elements and the tagged column items are stamped with the
**existing** `layout.ItemGroup`, so the shipped grouping machinery — union extent, ride-along, the
Story 4.6 over-tall clip — carries the whole feature. `internal/layout` gained nothing: its tree hash
is byte-identical across the commit.

The group key's `ElementID` carries a `:`, which the element-id grammar `^e[0-9a-z]+$` proves no
element id can equal, and `IsHeader` is false. That single choice makes every table-shaped path
unreachable at once, and both halves are **asserted rather than argued** — one test feeds the key
through the public loader as an element id and requires refusal *with a loading control*, another
drives an over-tall group through the clip branch and requires zero header repeats, zero row
displacement, zero suppression records and no footer-orphan target.

`TABLE_ROW_CLIPPED_HEIGHT` gained a **fourth role arm** rather than a minted code (the lead's call;
the dissent is recorded in Design Notes). Without it a signature block is announced to its author as
a numbered row of a bound collection. The D-4.6.2 tripwire was widened deliberately and renamed to
the invariant it now holds, its whitelist gaining the one new derivation and its vacuity floor
raised from 2 to 3.

**One review pass: 7 patched (high 1, medium 2, low 4), 7 deferred (high 1, medium 2, low 4), 8
rejected, 0 intent gaps, 0 spec defects.** The high was found by mutation, not by reading: the
page-count pass was wired for grouping and **wholly unverified** — removing all three of its
substitutions left the entire suite green, so a document's own footers could print the ungrouped
page count while the render produced a different number of pages. Closed by a 3-page grouped
document whose footers would print "of 2".

### 2026-08-31 — done

Baseline `45b11c5`; story commit `ed485eb`; closed on `main` in its own commit, **not pushed**.
Status `done`; `sprint-status.yaml`'s story key set to `done`. **`epic-7` is left `in-progress`** —
the boundary is held on the unresolved canvas-honesty defect below, which is with the engineering
lead, and Story 7.8 is still `backlog`.

**`followup_review_recommended` was `true` and has been cleared to `false`.** The ground is that the
close re-derived the build's load-bearing claims by mutation and measurement rather than relaying
them, and every one held: the high gap's fix, the fourth role arm, the widened tripwire, the
namespace audit, the fence, and the fixture's discrimination. Nothing new was patched at the close.

**Gates measured at this revision, with their printed numbers. Nothing was carried forward.**

- `cd folio-go && go test -count=1 ./...` — **1560 pass, 5 skip, exactly ONE distinct failure**:
  `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)`, `got 7, need >=20`, the mandated
  permanent red, untouched. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` green. 13
  packages ok.
- `go vet -tags=matrix ./...` — exit 0. `gofmt -l folio-go` from the repository root — no output.
- `TestTargetRenderHash` — run **once per leg with `FOLIO_MATRIX_TARGET` exported**: `darwin/arm64`,
  `linux/amd64`, `linux/arm64`, `js/wasm`, all PASS, `grep -c "asserts NOTHING"` = **0** on every
  leg. **The unset run was executed as a control** and does print the line once, so the four legs are
  known not to have been no-ops.
- `TestCrossTargetByteIdentity` — PASS (21.2s). It prints the new fixture agreeing on all four
  targets: `keep-together` 82,825 bytes, `6ed495b4…`.
- `cd lint && go test -count=1 ./...` — 4 packages ok.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — typecheck clean; oxlint
  **exactly 4** pre-existing `only-export-components` warnings, 0 errors; **280 tests across 33 test
  files**, unchanged. `npm run test:e2e:compile` clean — **`tsc --noEmit` only; the browser e2e did
  NOT execute** (D-000.4).
- **The nine digests, measured by `shasum` against the artifacts rather than inferred from a green
  suite, all byte-identical:** `statement-1` 76,744 `114df1d6…`; `-5` 127,363 `70dce051…`; `-20`
  269,884 `56bfbbd9…`; `-50` 555,829 `5d090b0f…`; `mandatory-break` 56,681 `7cf743de…`;
  `line-spacing` 57,770 `de212115…`; `justified-text` 59,894 `6da3b12e…`; `alignment-rounding`
  61,346 `986400a1…`; `justified-thai` 15,079 `58ca4777…`. `goldenDigestRecord` went 20 → 21 and its
  diff is **insertions only**, so the twenty pre-existing entries are unmoved mechanically and not
  by inspection.
- **Nothing is deferred to a catch-up run.** This story's correctness is byte-identity- and
  pagination-shaped, so it carried the heavy suites regardless of Epic 7's per-epic cadence
  (D-R7.1), and so did every other Epic 7 story: 7.1, 7.2, 7.3, 7.4, 7.5 and 7.6 each record the
  four `FOLIO_MATRIX_TARGET` legs with a zero "asserts NOTHING" count, `TestCrossTargetByteIdentity`,
  the lint module and the designer suite in their own Delivery Logs. Epic 7 owes no heavy-test
  arrears.
- Known-environmental, not regressions: `TestShippedFacesReproduceFromUpstream` (no `fontTools`
  here); `lint/internal/rules/licencegraph_test.go` gofmt (DW-23, Story 15.2).

**The high gap was reproduced, not read.** With all three page-count-pass substitutions removed, the
3-page grouped document's footers print `Page 0 of 2` on every page and both new tests fail with the
message naming the disagreement between the two passes. The source was restored and `cmp`-verified
byte-identical.

**The fence was verified mechanically, not argued.** `git diff --name-only 45b11c5 ed485eb` contains
**zero** paths under `folio-go/internal/layout/`; the directory's tree object is the same hash at
both revisions and `paginate_test.go`'s blob is unchanged, so `TestPaginateNeverProducesAnEmptyPage`
and the four-rule tests are green **and provably unmodified** (26 tests pass in that package).

**Two more mutations were run at the close.** Neutering the fourth role arm makes the over-tall
warning read *"row **-2** of the bound collection"* with a remedy about cell padding, and the test
reddens. Restoring the tripwire's floor to 2 and removing its new whitelist entry fires it naming
`render.go:2054:9: keepTogetherGroup`. Both sources restored and `cmp`-verified.

**The fixture discriminates, reproduced end to end.** Grouped: 2 pages, page 1 carries no member of
the signature block (neither the signature text nor the date text, **0 rects**), page 2 carries all
three including the ruled line's rect; 82,825 bytes, `6ed495b4…`. Ungrouped twin: the signature line
is stranded at the foot of page 1 while the rule and the date fall to page 2; 82,824 bytes,
`16067ec1…`. The two renders differ in bytes. The twin's "identical except for the tags" claim is
enforced by reconstruction — the test rebuilds the twin from the grouped template by deleting
exactly the tag substring and requires string equality — so the pair cannot drift apart for a second
reason and quietly stop discriminating.

**The canvas honesty defect was reproduced with a control, and it is held at the boundary.** For a
document with a body element, a two-member group and an untagged tail, the render produces **3
pages** while the canvas reports `ContentWindowCount = 2` with `ContentWindowCountIsFloor = false` —
that is, *exact* — and origins `[0, 740000]` where the render's second window begins at the group's
earliest top, 700000. The same document with the tags removed renders 2 pages and the canvas is
right, which is the control that makes the failure the grouping's and not the fixture's. This is a
**fourth floor cause Story 7.6 does not know about**, and the wrong origins are a **separate** failure
that a floor flag does not disclose at all. Filed as **DW-46**, owner pending the lead's ruling. It
is why `epic-7` is not being marked done here.

**Seven deferrals filed into the register as DW-46 … DW-52**, where the build had recorded them only
in this spec's frontmatter. DW-43 was amended by the build to name this fixture and left OPEN.
DW-26/27/28/30/31/32/33/34/35/37/39-45 are untouched and OPEN; DW-29 remains Story 7.8's; DW-33 and
DW-38 still need rulings rather than patches, DW-38 routed to Epic 14.2 and DW-42 to a named Epic 15
story (D-7.7.3, D-7.7.4).

**Repository hygiene.** The story commit contains only this story's files and **no root
`README.md`**, whose md5 is still `078d7d80d518d54af2fc04fb270d46b8`. Both commits carry the
required `Co-Authored-By:` trailer. Nothing was pushed.
