---
title: 'Story 7.8: Refuse a justified table at load, in the author''s own terms'
type: 'bugfix'
created: '2026-08-31'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
warnings: ['oversized']
deferred: []
---

<intent-contract>

## Intent

**Problem:** A table element's `style.align: "justify"` or `headerStyle.align: "justify"` loads
without error, raises the document to format **2.0** — unreadable to every 1.x reader — and then
renders identically to `align: left` with no diagnostic. The author pays a MAJOR bump and receives
nothing. The root cause is D-7.3.1's guardrail partitioning the alignment closed set **by JSON key
location** (`style`/`headerStyle` vs `columns[]`) instead of **by the code that consumes the value**:
a table's `style.align` and its `columns[].align` are consumed at the same site through
`r.alignFallback`, so the guard meant to make justified table cells impossible let the value in
through the other door.

**Approach:** Re-partition the alignment vocabulary **by consumer, keyed on element type**. Add a
third closed set for a table's style-attachment points, thread the consumer's element type into
`decodeStyle` (both callers already hold it), and refuse `justify` on a table's `style.align` and
`headerStyle.align` as a **located** load error naming the element and the field. A text element's
`justify` is untouched. The format-version half then closes by construction on the file path, and is
closed on the in-memory path by making the property-command arm validate through the same
consumer-keyed source the loader uses.

## Boundaries & Constraints

**Always:**
- A **text** element's `style.align: "justify"` stays accepted exactly as today, still raises the
  document to `2.0`, and the `justified-text` and `justified-thai` goldens render byte-identically.
- All **twenty-one** `goldenDigestRecord` entries (`folio-go/byte_neutrality_test.go:92`) stay
  byte-identical, verified by `shasum` against the shipped artifacts. AD-21/AD-22 byte identity is
  the product; a moved digest is a defect until proven intended.
- The refusal is **located**: it names the element id and the field (`style.align` or
  `headerStyle.align`), never a bare "invalid value".
- No message may claim a set of legal values that differs from the set actually enforced
  (`internal/template/closedsets.go:45-47`). Every rejection message is derived from an ordered token
  slice via `closedSetMessage`, never hand-written.
- The property-command path validates `align` against **the same single source the loader does** —
  the invariant `IsStyleAlign`'s own doc comment states (`closedsets.go:75-78`).
- No `float64` under `internal/` (AD-23). No new binary-float arithmetic anywhere.
- `fixtures/statement-signoff.json` is a human attestation. No agent writes its `reader`, `date` or
  `examined` fields.
- `README.md` is the user's file: never modified, moved, deleted, or committed. Stage explicit paths;
  never `git add -A` or `git add .`.

**Block If:**
- **THE DIAGNOSTIC-CODE QUESTION IS SETTLED — D-7.8.1, ruled 2026-08-31. Implement the ruling; do
  not re-derive it and do not substitute a cheaper local option.** Mint **exactly one** code,
  `TEMPLATE_FIELD_INVALID`, for the category *a well-formed template carries a field value that is
  not acceptable*, and have **`newLoadError` itself supply it**, so every uncoded site becomes coded
  by construction — no enumeration, no per-site judgement. `newLoadErrorCoded` stays as the override
  for conditions that genuinely need discrimination. **Block if** the implementation instead mints a
  per-field code, invents an unregistered string literal, reuses `CodeStyleLineSpacingInvalid`,
  downgrades AC4 to an uncoded error, or changes `reportableMessage`'s treatment of
  `TEMPLATE_MALFORMED` (that code keeps destroying its own messages, for the reason it was written —
  what changes is that `LoadError`s stop being bucketed there).
- **Block if this story retires or re-means `STYLE_COLOR_INVALID` or `STYLE_LINE_SPACING_INVALID`.**
  Auditing those two against the new rule is a named obligation triggered by the `folio-go/v0.1.0`
  tag (D-7.8.2), not this story's work, and it must not happen here by accident.
- **Block if any of the three falsified tests is DELETED, or if any is reported as having no
  meaningful inverted form.** All three invert. `table_render_test.go:1338`'s subject becomes
  unloadable, so its assertion **moves down to the load layer** — "unreachable at this layer" is a
  reason to relocate an assertion, never a licence to drop it. If the implementer judges an inverted
  form impossible, that returns to the engineering lead rather than being resolved in the diff.
- A golden digest moves. Stop and report; do **not** re-record, and do **not** re-attest.

**Never:**
- Never delete the three shipped tests this story falsifies — **invert** them. They are correct
  against the contract as it was written and they pin today's acceptance. Deleting a failing test
  instead of inverting it is the specific failure this story must not commit.
- Never ban `justify` at the `Style` type or the `closedStyleAligns` map wholesale — that is a
  blanket ban wearing a narrow name and it breaks Story 7.3, Story 7.4 and two goldens.
- Never touch `folio-go/internal/layout/` — zero paths there. `paginate.go`'s prohibition on a
  closed-form page count stands untouched.
- Never re-do Story 7.4's product half. The inspector already offers `justify` for text alone; no
  change to `folio-designer/src/App.tsx`'s `alignChoices`, and no new designer feature work.
- Never widen this into `valign`, `columns[].footer`, or any other closed set.
- Never touch `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` (mandated red) or its drift
  twin.
- Never create a branch. Never push. `main` only.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Table `style.align` justify | `.folio` with `{"type":"table", "style":{"align":"justify"}}` | `ParseDocument` refuses | Located load error: field `style.align`, element id named, reason lists `left, center, right` |
| Table `headerStyle.align` justify | `.folio` with a table's `"headerStyle":{"align":"justify"}` | `ParseDocument` refuses | Located load error: field `headerStyle.align`, element id named, reason lists `left, center, right` |
| Text `style.align` justify | `.folio` with `{"type":"text", "style":{"align":"justify"}}` | Loads unchanged; `versionRequiredByContent` = `"2.0"` | No error expected |
| Table style justify never reaches 2.0 | The refused document above | No `*Document` exists, so no version is ever computed for it | No error expected — refusal precedes version entirely |
| Property command, table + justify | `updateComponentProperties{id: <table>, align: "justify"}` | Command refused; the table's `style.align` is unchanged and the serialized document stays below `2.0` | Command failure naming the field and the legal values for a table |
| Property command, text + justify | `updateComponentProperties{id: <text>, align: "justify"}` | Accepted exactly as today; canonical bytes carry `"align": "justify"` | No error expected |
| Table `columns[].align` justify | `{"columns":[{"align":"justify"}]}` | Refused, unchanged from today | Located at the column id, reason lists `left, center, right` |
| Table style `"middle"` | `{"type":"table","style":{"align":"middle"}}` | Refused | Message lists **`left, center, right`** — it must **not** name `justify` as legal for a table |
| Text style `"middle"` | `{"type":"text","style":{"align":"middle"}}` | Refused, unchanged from today | Message lists `left, center, right, justify` |
| Every shipped golden | The 21 `goldenDigestRecord` fixtures | Byte-identical digests; no fixture contains `justify` on a table | No error expected |

</intent-contract>

## Code Map

**Every anchor below was verified by reading the file at baseline `53b2c1f`.** Three anchors in the
epic text are **wrong** and are corrected here — see *Anchor corrections* at the end of this section.

### The validation seam — `folio-go/internal/template/`

- **`closedsets.go:33-92`** — the alignment vocabulary. Token constants `:48-53`
  (`AlignLeft`/`AlignCenter`/`AlignRight`/`AlignJustify`). Two ordered slices: `StyleAlignTokens`
  `:60` = `{left, center, right, justify}`; `ColumnAlignTokens` `:65` = `{left, center, right}`. Two
  maps: `closedStyleAligns` `:67-69`, `closedColumnAligns` `:71-73`. `IsStyleAlign` `:79`.
  `closedSetMessage(tokens []string) string` `:86-88` — the shared derived-message helper (D-7.3.1
  guardrail 2), already present.
  ⚠ **THE RULE THAT DECIDES THE DESIGN**, doc comment `:45-47`: *"the rejection MESSAGE is derived
  from the slice rather than restated as a literal: after Story 7.3 no message may claim a set of
  legal values that differs from the set actually enforced."*
  ⚠ The `StyleAlignTokens` doc `:56-59` claims the set is what *"`style.align` and
  `headerStyle.align` admit"*. **That sentence becomes false** and must be rewritten.
  ⚠ `closedValigns` `:90-92` is still a bare map with **no** ordered-slice twin and a hand-written
  message (`parse_bands.go:608`). Out of scope — do not touch, do not "fix" it in passing.
- **`parse_bands.go:580`** — `func decodeStyle(elementID string, raw json.RawMessage, fieldPrefix string) (Style, error)`, doc `:570-579`, body to ~`:713`. **It receives no element type. That is the
  whole structural defect.** The align check is `:596-598`:
  ```go
  if !closedStyleAligns[s] {
      return Style{}, newLoadError(fieldPrefix+".align", elementID, s, closedSetMessage(StyleAlignTokens))
  }
  ```
- **`decodeStyle` has exactly TWO callers, both in this file** (repo-wide grep; it is unexported and
  no test calls it):
  - `:210` — `st, err := decodeStyle(string(id), styRaw, "style")`, inside `decodeElement` (`:111`).
    **`el.Type` is live**: set at `:143` (`el := Element{ID: id, Type: ElementType(typeStr)}`) after
    the `closedElementTypes` check `:140-142`, and the function already branches on it at `:164`,
    `:254` and `:271`.
  - `:397` — `hs, err := decodeStyle(id, hsRaw, "headerStyle")`, inside `decodeTableExt` (`:328`),
    whose single call site `:307` sits under `case ElementTable:` `:306`. **Statically a table**; no
    type variable is in scope, so the constant is passed.
- **`parse_bands.go:452-466`** — `decodeColumn` (`:411`), the `columns[].align` check `:462` against
  `closedColumnAligns`, error `:463`. **READ-ONLY. Already correct; it must be absent from the diff.**
- **`errors.go:58-63`** `newLoadError(field, elementID, value, reason string) error` → `*LoadError`
  with `Code` left `""`. **`errors.go:73`** `newLoadErrorCoded(..., code diag.Code)` is the only way
  over the uncoded/coded line. `LoadError.Error()` `:51-56` is where location already lives:
  `"template: field %s (element %s): %s (value: %s)"`.
- **`model.go`** — `ElementType` `:174`; constants `:176-182` (`ElementText` `:177`, **`ElementTable`
  `:179`**). `Element.Type` `:193`; `Element.Style Presence[Style]` `:199` (**every** element type).
  **`TableExt.HeaderStyle Presence[Style]` `:249` — table-only.** `Style.Align` `:271`;
  `Column.Align` `:257`.

### The version half — `folio-go/internal/template/version.go`

- `SupportedMajor = 2` `:64`, `SupportedVersion = "2.0"` `:65`. Version strings `:88-93`. Rank ladder
  `:263-268` is four rungs (`rankKeepTogether` `:266` shipped with Story 7.7).
- `versionRequiredByContent` `:221-250` — walks bands → elements. Three probes: `KeepTogether` `:231`,
  `styleVersionRank(el.Style.Value)` `:235`, `styleVersionRank(el.Table.Value.HeaderStyle.Value)`
  `:242`.
- `styleVersionRank(st Style) versionRank` `:296-305`. Deciding line `:301`:
  `if st.Align.Set && !st.Align.Null && st.Align.Value == AlignJustify { rank = rankMajorFeature }`.
  **It is type-blind by signature** — one bare `Style`, no element, no type.
- ⚠ **THE ORDERING FACT THAT MAKES THE VERSION HALF FREE.** `versionRequiredByContent` has exactly one
  non-test caller, `versionForSave` `:169`; `versionForSave` has exactly one non-test caller,
  `serialize.go:118`. **It never runs on the load path.** Load touches version only at `parse.go:72`
  `checkVersionLoadable(version)`, on the *declared string* alone, before `bands` is decoded
  (`:133-142`). So once `decodeStyle` refuses, `ParseDocument` returns `nil, err` and **no `*Document`
  ever exists** for the version probe to observe. **`styleVersionRank` needs no change for the file
  path — assert this, do not code it.**
- ⚠ **THE IN-MEMORY HOLE, and it is real.** `folio-go/component_commands.go:909` puts `"align"` in
  `allowed` for `ElementText || ElementTable`; the `case "align":` arm `:1048-1071` validates only
  `if !template.IsStyleAlign(text)` `:1066` — **no element-type check**. So
  `updateComponentProperties{align:"justify"}` on a **table** is accepted today, `styleFor(element)`
  writes it, and the next `SerializeTemplate` stamps the document `2.0`. Story 7.4 closed only the
  **UI** door (`folio-designer/src/App.tsx:992`), not the engine's. Without this arm, AC1's *"never
  raised to format 2.0 for that value"* is false and the designer can author a file it cannot reopen.

### The consumption site — `folio-go/table_render.go` (READ-ONLY evidence)

- `resolvedHeaderStyle.alignFallback` `:294-296`; `resolvedBodyStyle.alignFallback` `:412`; both
  seeded `"left"` at `:312` / `:423`.
- Header assignment **`:372-377`** (`headerStyle.align` wins, else `style.align`); body assignment
  **`:440-442`** (`style.align` only — deliberately no headerStyle arm, doc `:383-392`).
- Consumed at **four** sites, identical shape `align := <x>.alignFallback; if col.Align.Set { align = col.Align.Value }`: `:672-675` header, `:863-866` body, `:1110-1113` and `:1117-1120` footer.
  **This is the proof of the "same consumer" claim: `style.align` and `columns[].align` meet here.**
- The three `default:` arms whose comments Story 7.3 corrected: header `:706-714`, body `:1047-1055`,
  footer `:1231-1239`. Each says *"the contract forbids IMPLEMENTING justified table cells"* and
  *"Any OTHER value the load-time closed-set check already rejected."* **After this story `justify` is
  also rejected at load, so all three comments become stale and must be updated.**

### The diagnostic path (see Block If — decision RESERVED)

- `internal/diag/diag.go` — 16 registered codes, const block `:62-254`; **the reservation `:249-252`**;
  `CodeStyleLineSpacingInvalid` `:253` (load-time, the true precedent);
  `CodeStyleColorInvalid` `:149` (render-time). `allCodes` `:261-278`; `dispositions` `:295-312`.
  Additive-only rule stated in-file `:56-58`.
- `render_error.go:97-106` `wrapTemplateError` — a `LoadError` with `Code == ""` becomes
  `DiagCodeTemplateMalformed`; note it passes `dataPath = ""`, so a load error's **field** survives
  only inside `Message`.
- `folio-go/wasm/cmd/engine/main.go:272-281` `reportableMessage` — replaces the message for
  `DiagCodeTemplateMalformed` **and only that code**; every other code passes through
  `bounded(message, 512)` intact.
- The working precedent, end to end: `parse_bands.go:685` raises `CodeStyleLineSpacingInvalid` via
  `newLoadErrorCoded`, with the in-situ rationale at `:677-684`; proven to survive to the host by
  `folio-go/wasm/cmd/engine/main_test.go:63-81`. **The designer needs no TypeScript change for a new
  code** — `folio-designer/src/preview/diagnostic-presenter.tsx:10` is deliberately code-agnostic.
- Minting a third code touches: the const block; `allCodes` `:261-278`; `dispositions` `:295-312`
  (else `diagnostic_registry_census_test.go:20` fatals); `codePins` `diag_test.go:25-46`; the public
  bridge const in `folio-go/diagnostic.go`; `diagCodeBridgePins` `diag_bridge_test.go:27-48` (its AST
  scan `:61-101` auto-fires on any new exported `DiagCode*`); and the census **error trigger** map
  `diagnostic_registry_census_test.go:24-90` (fatals at `:97`/`:102`; the trigger must produce a real
  `*RenderError` with a non-blank message and a location).
- ⚠ **AD-14 (`ARCHITECTURE-SPINE.md:311-324`) says nothing normative about *when* a new entry may be
  added.** It constrains only mutation: *"Codes are additive only; changing a code's meaning is a
  breaking change."* There is **no** general-form code in the tree today — grep for
  `STYLE_FIELD*`/`*FIELD_INVALID`/`PROPERTY_INVALID` returns zero. No registry code carries a field
  name as structured data; `Diagnostic.DataPath` exists but both style codes pass `""`.

### The tests this story falsifies

- **`closedsets_test.go:287-291`** — leg (b) of `TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage` (doc `:277-281`, body `:282-347`). Asserts `headerStyle.align: "justify"` **must load**.
  **INVERT.** Note it is a 6-line leg inside a 66-line test, not a whole function. Legs (a) `:283-286`
  text-style justify loads, (c) `:293-313` column refusal, (d) `:314-326`, (e) `:328-335`, (f)
  `:337-346` message derivation — all **stay**, though (d)'s expected message changes if the element
  it uses is a table (it is not: `alignDocWithStyle` `:150-172` builds a **text** element).
  Helper `alignDocWithTable` `:175-200` emits **no table `style` block** — only `headerStyle` and
  `columns[]` — so **nothing currently pins a table's own `style.align: "justify"`**; that leg must be
  **added**, not inverted.
- **`internal/template/linespacing_test.go:230-237`** — asserts `justifyHeaderStyleDoc` loads
  (`t.Fatalf` on error) **and** that `versionRequiredByContent` = `"2.0"`. **INVERT.** The const
  `justifyHeaderStyleDoc` `:479-503` is the version-test fixture the epic says is deleted with them —
  it is the only justify-bearing fixture anywhere that puts `justify` on a **table**, and it has
  exactly **one** consumer (`:231`).
- **`table_render_test.go:1338-1368`** `TestTableCellsCascadedJustifyIsDrawnAtTheStartEdge`, doc
  `:1322-1337`, helper `justifyCascadeColumns` `:1313-1318`. **It goes through the REAL load path** —
  `tableHeaderDocFull(style, style, ...)` → `ParseTemplate` → `Render` at `:1340-1352` — and injects
  the same style at **both** attachment points. So it fails at `t.Fatalf` on `ParseTemplate` before
  any byte comparison. **INVERT**: its byte-identity claim becomes unreachable because no such
  document can exist.
- ⚠ **A FOURTH TEST THE EPIC DOES NOT LIST.** `closedsets_test.go:215-275`
  `TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps` — its **name** asserts there are two sets, and
  `:261-263` asserts `len(StyleAlignTokens) == len(ColumnAlignTokens)+1`. A third set breaks both.
  **Rename and extend** (the 7.7 `rowGroupDerivations` tripwire-widening precedent); keep every
  both-directions map pinning and the `IsStyleAlign` agreement checks `:266-274` intact.
- ⚠ **A FIFTH.** `component_properties_test.go:217-249`
  `TestStyleAlignPropertyValidatesAgainstTheStyleSetOnly` runs on `worked-example.json`'s `e1`, a
  **text** element, so it stays green — but its **name and doc become false** under the command-path
  fix. Rename and add a **table** leg.

### Read-only, must be absent from the diff

`folio-go/internal/layout/**` (zero paths). `parse_bands.go`'s `decodeColumn` align check `:452-466`.
`component_commands.go:290-295` (the column arm; already refuses `justify`).
`folio-designer/src/App.tsx` (Story 7.4's product half). `fixtures/statement-signoff.json`.
`README.md`. `folio-designer/src/engine-protocol.ts:278` — its `includes('justify')` gate has no
per-type branch, but after this story no table can carry `justify`, so it is harmless; **note it, do
not change it** (designer scope fence).

### Anchor corrections to the epic text (`epics.md:2313-2377`)

1. **`folio-go/line_spacing_test.go:168-175` with its const at `:311-331` is WRONG.** That file exists
   (935 lines, root package) and contains **zero** occurrences of `justify`; its `:168-175` is
   `TestNeutralRatioNeverRefusesADegeneracyItDidNotCause`. The real site is
   **`folio-go/internal/template/linespacing_test.go:230-237`**, const at **`:479-503`**.
2. `parse_bands.go:204` / `:338` for the two `decodeStyle` callers → actually **`:210`** and
   **`:397`**. The epic's *claim* holds exactly: `el.Type` is live at `:210`; `:397` is statically
   `ElementTable`.
3. `closedsets_test.go:287-292` → the block is **`:287-291`**.
4. Correct as written: `table_render_test.go:1338`; `table_render.go:373-376` and `:440-441` (the
   enclosing statements open at `:372` and `:440`); `diag.go:249-252`;
   `wasm/cmd/engine/main.go:276-281` (full function `:272-281`).

## Tasks & Acceptance

**The ordering is normative.** Part 0 gates everything. Parts 1–2 must land before Part 3 can assert
anything, and Part 5's fixture must land in the same commit as Parts 1–4.

**Part 0 — the settled decision, and the two things it obliges. DO THIS FIRST.**
- Implement **D-7.8.1** as ruled (see Design Notes, *"The reserved decision, and how it was
  settled"*): one code `TEMPLATE_FIELD_INVALID`, supplied **by `newLoadError` itself**. The
  registry-policy rule this establishes must be written into `internal/diag/diag.go` in place of the
  reservation at `:249-252`, so the next reader finds the answer where they would have found the
  question: **the general code is the default; a specific code is minted only when a named consumer
  must BRANCH on it to behave differently.** Everything else discriminates on the `Field` datum,
  which can grow freely without touching a closed registry.
- **Assert the property END TO END, through the wasm entry point** — a template in, the author's
  words out — not at the constructor. The requirement is that **a `LoadError` must not reach the
  wasm boundary carrying `TEMPLATE_MALFORMED`**. The defect being fixed lives at the boundary, so a
  Go-side-only assertion would pass while the author still sees one sentence.
  `folio-go/wasm/cmd/engine/main_test.go:63-81` is the shipped shape to follow.
- **Re-point and RE-MEASURE AC41's enumeration test.** Its subject is "call sites of the uncoded
  constructor", and after this change that set is empty or means something different. A test that
  keeps passing while measuring nothing is the dead-detector failure recorded at D-7.4.2. Prove the
  re-pointed test is non-vacuous by breaking it deliberately and showing it reddens; report the
  mutation and its output.
- **Load-stage only.** `TEMPLATE_FIELD_INVALID` does not absorb render-stage conditions;
  `diag.go`'s own scope note for `STYLE_LINE_SPACING_INVALID` already draws that line (a zero
  resolved advance and int64 overflow are render-stage with a different remedy). The code names the
  **condition**, not the field and not the call site — the same discipline that gave one code to two
  `TABLE_FOOTER_SOURCE_FORBIDDEN` sites.

**Part 1 — the third closed set.**
- `folio-go/internal/template/closedsets.go` — add `TableStyleAlignTokens` (ordered slice,
  `{AlignLeft, AlignCenter, AlignRight}`) and its `closedTableStyleAligns` map, both built from the
  existing token constants, beside the two shipped sets. Rewrite the `:33-47` header comment to state
  the **consumer** partition and D-7.3.1's reusable lesson, and rewrite `StyleAlignTokens`' doc
  `:56-59` so it no longer claims to be what `headerStyle.align` admits. Add an `IsTableStyleAlign`
  predicate mirroring `IsStyleAlign` `:79` — the command path needs it and `closedSetMessage` must
  serve the new slice unchanged. Do **not** touch `closedValigns`.

**Part 2 — thread the consumer into the loader.**
- `folio-go/internal/template/parse_bands.go` — give `decodeStyle` (`:580`) an element-type
  parameter, and at `:596-598` select the set by that type: a table validates against
  `closedTableStyleAligns` and its message is `closedSetMessage(TableStyleAlignTokens)`; every other
  type keeps `closedStyleAligns` / `StyleAlignTokens`. Pass `el.Type` at `:210` and the
  `ElementTable` constant at `:397`. The error stays `newLoadError`-shaped **except** for the code
  question settled in Part 0. Leave `decodeColumn` `:452-466` and the `valign` arm `:601-611`
  untouched.

**Part 3 — close the in-memory door.**
- `folio-go/component_commands.go` — in the `case "align":` arm `:1048-1071`, select the predicate by
  `element.Type` (already in scope, used at `:897`–`:909`), and name the legal values from the
  matching ordered slice. Rationale is not preference: `IsStyleAlign`'s own doc comment
  (`closedsets.go:75-78`) requires the property-command path to *"validate it against the same single
  source the loader does"*, and without this AC1's *"never raised to format 2.0"* is false for every
  designer-authored document. Do **not** touch the column arm `:290-295`.

**Part 4 — invert the falsified tests, and widen the two tripwires.**
- `folio-go/internal/template/closedsets_test.go` — **invert** leg (b) `:287-291` to assert
  `headerStyle.align: "justify"` is now a **located** load error naming the element and
  `headerStyle.align`, with a message listing `left, center, right`. **Add** a leg for a table's own
  `style.align: "justify"` (nothing pins it today — `alignDocWithTable` `:175-200` emits no table
  `style` block, so extend the helper). Keep legs (a), (c), (d), (e), (f). **Rename and extend**
  `TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps` `:215-275` for three sets: replace the
  `len(StyleAlignTokens) == len(ColumnAlignTokens)+1` assertion `:261-263` with pinning that states
  the new invariant, and keep every both-directions map check and the `IsStyleAlign` checks
  `:266-274`.
- `folio-go/internal/template/linespacing_test.go` — **invert** `:230-237` into a refusal assertion,
  and **delete** the now-unreachable `justifyHeaderStyleDoc` const `:479-503` **or** rewrite it onto a
  text element; state in the test's comment which was chosen and why. Every other justify version
  fixture in this file (`:137`, `:140`, `:141`, `:192-205`, `:245-259`, `:308`) is built on
  `"type": "text"` and **stays green with no edit** — verify that rather than assuming it.
- `folio-go/table_render_test.go` — **invert** `TestTableCellsCascadedJustifyIsDrawnAtTheStartEdge`
  `:1338-1368` into an assertion that such a document is refused at load, naming element and field.
  **Rename it** — its current name asserts a property this story makes unreachable. Preserve the
  `center`-vs-`left` non-vacuity leg for the values that still load, so the cascade is still proven
  to reach the cell align switches.
- `folio-go/table_render.go` — update the three stale `default:`-arm comments (`:706-714`,
  `:1047-1055`, `:1231-1239`): `justify` no longer reaches a table cell through `alignFallback`,
  because it is now refused at load. **Comment-only; the code is unchanged and the arms stay.**
- `folio-go/component_properties_test.go` — rename
  `TestStyleAlignPropertyValidatesAgainstTheStyleSetOnly` `:217-249` and add a **table** leg asserting
  the command refusal.

**Part 5 — the discriminating fixture and the version proof.**
- `folio-go/internal/template/` — a test asserting the **version half** directly: build a table
  document carrying `style.align: "justify"`, assert `ParseDocument` refuses it, and assert that the
  refusal is what prevents `2.0` — i.e. that no `*Document` reaches `versionRequiredByContent`. Pair
  it with a **text** document carrying `justify` asserting `versionRequiredByContent` still returns
  `majorFeatureVersion`. **The two legs together are the discriminator**; either alone is consistent
  with a blanket ban, which the Never list forbids.
- `folio-go/` — an end-to-end test that the property command on a table cannot produce a `2.0`
  document: issue `updateComponentProperties{align:"justify"}` against a table element, assert the
  command is refused **and** that `SerializeTemplate` still stamps the document below `2.0`.
- `folio-go/wasm/cmd/engine/main_test.go` — on the pattern of
  `TestWasmHostReportsTheLineSpacingRefusalIntact` `:63-81`, assert the table-justify refusal reaches
  the host **with element id and field intact** and is **not** flattened to *"The template could not
  be processed"*. **This test is AC4 and it is only writable once Part 0 is settled.**
- `_bmad-output/specs/spec-folio/folio-format.md` — amend the sentences this story falsifies: `:47`
  (*"`2.0` if any style sets `align: "justify"`"*), `:69-80` (align is now **three** closed sets, and
  the "This has happened once" note), `:267` (`headerStyle` is *"same vocabulary as an element's own
  `style`"* — no longer true), and `:340` (*"Declaring `justify` anywhere raises the document to
  version `2.0`"* — no longer true for a table). State the consumer partition as the rule.
- `_bmad-output/implementation-artifacts/deferred-work.md` — close DW-29, naming this story and the
  commit.

**Acceptance Criteria:**
- Given a `.folio` file setting `style.align: "justify"` on a table element, or on its `headerStyle`,
  when it is loaded, then it is refused with a located error naming the element and the field, and no
  `*Document` exists for `versionRequiredByContent` to raise to `2.0`.
- Given a `.folio` file setting `style.align: "justify"` on a **text** element, when it is loaded,
  then it is accepted exactly as today, `versionRequiredByContent` returns `"2.0"`, and the
  `justified-text` and `justified-thai` goldens render byte-identically.
- Given the alignment vocabulary, when any align value is validated — at load **or** through the
  property-command path — then it is validated against the set its **consumer** accepts, keyed on the
  element type, and the rejection message lists exactly the members of the set that rejected it.
- Given a designer author who opens such a file, when the engine reports the refusal, then the
  element id and the field reach the host intact rather than being flattened to *"The template could
  not be processed"*.
- Given any template not setting `justify` on a table, when it renders, then all **twenty-one**
  `goldenDigestRecord` entries are byte-identical and its saved `version` is unchanged.
- Given the three falsified tests, when the diff is inspected, then each is **inverted or renamed**
  and none is deleted; and `git diff --name-only` contains **no path under
  `folio-go/internal/layout/`**.

## Spec Change Log

## Review Triage Log

## Design Notes

### The reserved decision, and how it was settled (D-7.8.1, 2026-08-31)

AC4 requires the refusal to reach a designer author *in words*. An **uncoded** `newLoadError` cannot:
`wrapTemplateError` (`render_error.go:97-106`) turns it into `TEMPLATE_MALFORMED`, and
`reportableMessage` (`folio-go/wasm/cmd/engine/main.go:272-281`) replaces that one code's message with
*"The template could not be processed"*. `internal/diag/diag.go:249-252` reserved the choice; the
engineering lead has now made it.

**The ruling: one code, `TEMPLATE_FIELD_INVALID`, supplied by `newLoadError` itself.** Not a
per-field code. Not a change to the wasm boundary rule.

**What the lead measured, and why it made the answer cheaper than any option this spec had framed.**
Reading `internal/template/errors.go:42-75`:

- **`LoadError` is already structured** — `{Field, ElementID, Value, Reason, Code}`. The general
  option's selling point, "carry element/field/value as data", is **already true today**. The only
  thing missing is the `Code`.
- **Its message never quotes the document.** `Error()` renders `"template: field %s (element %s):
  %s (value: %s)"` — one field, one bounded value. The reflection hazard `reportableMessage` guards
  against is a **document-quoting** message, and a `LoadError` is not one. **The boundary rule is
  over-broad by accident, not by design.**
- **`newLoadError` is a SINGLE constructor** that every uncoded load-error site in the package calls,
  and **AC41's enumeration test already walks those call sites** — so the instrument that would
  otherwise have to be built already exists.
- **`newLoadErrorCoded` already exists** and already codes three footer-source conditions, so coding
  a load error is established practice rather than a new mechanism.

Together those make the fix structural rather than enumerative: put the code in the constructor and
every uncoded site becomes coded at once, with no per-site decision and no accretion.

**The registry-policy rule the reservation asked someone to decide:**

> **The general code is the default. A specific code is minted only when a named consumer must
> BRANCH on it to behave differently.** Everything else discriminates on the `Field` data, which can
> grow freely without touching a closed registry.

This is D-7.3.1's own lesson applied one level up — partition by what the consumer **does**, not by
where the value is written. A designer receiving any of these does exactly one thing: locate the
element, name the field, show the value and the reason. One behaviour, one code.
`STYLE_COLOR_INVALID` versus `STYLE_LINE_SPACING_INVALID` buys a consumer nothing it cannot get from
`Field`, which is why the registry was heading for one entry per style field forever.

**Why the other two options were rejected.** Minting a third per-field code answers *"is the general
form right?"* with *"not yet"*, which is how the reservation's own worry comes true — the third mint
is the moment the pattern becomes a policy by default. Changing the wasm boundary rule was rejected
too, **and this ruling gets its benefit without its blast radius**: `TEMPLATE_MALFORMED` keeps
destroying its own messages, for the good reason it was written; what changes is that `LoadError`s
stop being bucketed there. Nothing else reaching `reportableMessage` is affected, so there is no
unaudited population.

**What this story does NOT do.** It does not retire or re-mean the two existing style codes. At
least `STYLE_COLOR_INVALID` is a **render** error by Epic 10's own AC, so it is not in
`TEMPLATE_FIELD_INVALID`'s load-stage category at all and may be correctly specific;
`STYLE_LINE_SPACING_INVALID` spans a load path and a property-command path. Neither is a clean
migration candidate on today's evidence. Auditing both against the rule above — *does any consumer
branch on them?* — is a **named obligation triggered by the `folio-go/v0.1.0` tag** (D-7.8.2),
because AD-14 makes removing a code a breaking change and that is free exactly once.

### Why a third SET rather than a type-guard bolted above the existing two

Both readings refuse the same documents, so the discriminator is the **message**, and
`closedsets.go:45-47` states the rule as an invariant: *"no message may claim a set of legal values
that differs from the set actually enforced."* Under a bare type-guard, a table's
`style.align: "middle"` would still be rejected by `closedStyleAligns` and would report *"not one of
the closed set left, center, right, **justify**"* — a message naming `justify` as legal for an element
that can never carry it. That is exactly the lie the rule forbids, and it is the same class of defect
as the one this story exists to fix. The third set also keeps `closedSetMessage` as the single
derivation, so the message cannot drift from the set.

The cost is honest and is booked as a task: `TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps` is named
and written for two sets and must be widened for three.

### Why the version half needs no version-code change on the file path

Traced and verified: `versionRequiredByContent` runs **only** at save
(`version.go:169` ← `versionForSave` ← `serialize.go:118`), on a fully parsed and fully validated
`*Document`. The load path touches version only at `parse.go:72`, on the declared string, before
`bands` is decoded. A loader refusal therefore closes the 2.0 raise **by construction** — there is no
second place to patch, and `styleVersionRank`'s type-blindness is not a defect to fix. This is why
Part 5 asserts the property rather than coding it.

The property-command path is the genuine exception, and it is why Part 3 exists rather than being
scope creep: it is the one remaining route by which a table document reaches `2.0`.

### Scope fence

This is the **format** half only. Story 7.4 discharged the product half — the inspector no longer
offers `justify` for a table or a mixed selection (`App.tsx:992`, pinned by `App.test.tsx:950-988`).
Part 3 is engine code (`folio-go/component_commands.go`), not designer work, and no path under
`folio-designer/` is touched.

## Verification

This story changes the file format's accepted set and a pagination-adjacent input, so it carries the
heavy tests regardless of the per-epic cadence (**D-R7.1**). **Report measured pass/fail counts, never
"green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected: **exactly ONE** failure,
  `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go`, `P6g_(opaque_names)` subtest,
  `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin
  `TestCorpusP6StatsMatchDeclaredBaseline` must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` — expected: clean, exit 0.
- `gofmt -l folio-go` — run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` — run **once per leg** with
  `FOLIO_MATRIX_TARGET` **exported**: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`
  (`matrix_test.go:69-74`). ⚠ **Unset, this test logs "asserts NOTHING" and returns — a no-op is not a
  pass.** Also run **one unset control** and show it printing `asserts NOTHING`, proving the four legs
  are not no-ops. Grep each leg for `asserts NOTHING` and report the count per leg; name the legs.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` — expected: pass; all four
  targets from one process.
- `cd lint && go test ./...` — expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — expected: pass, test count
  unchanged (this story adds none). **oxlint baseline is exactly 4 `only-export-components` warnings
  and 0 errors**; a fifth is a regression.
- `cd folio-designer && npm run test:e2e:compile` — expected: pass. ⚠ `tsc --noEmit` only. **Browser
  e2e is deferred by D-000.4 and does NOT execute — say so rather than implying it ran.**

**All twenty-one `goldenDigestRecord` entries must be verified byte-identical by `shasum` against the
shipped artifacts** (`byte_neutrality_test.go:92`): minimal-rect, font-text, image-embed,
multi-script-fallback, shaped-text, three-band-page, multi-page, page-count-20, wrapped-text,
statement-1, statement-5, statement-20, statement-50, alternating-rows, component-asset-import,
mandatory-break, line-spacing, justified-text, justified-thai, alignment-rounding, keep-together.
**A changed digest is a defect until proven intended — stop and report; never re-record and never
re-attest.** Confirm `git status fixtures/` is clean before quoting any digest.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` fails under
`-tags=matrix` without `fontTools`; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean
(DW-23).

**Manual checks:**
- **Confirm `git diff --name-only` contains no path under `folio-go/internal/layout/`.** Mechanical.
- **Confirm `README.md` appears in no commit and its md5 is still
  `078d7d80d518d54af2fc04fb270d46b8`.** Stage explicit paths; never `git add -A`.
- **Confirm `fixtures/statement-signoff.json` is unmodified.**
- **Confirm no shipped fixture gained `justify` on a table**: `grep -rn justify fixtures/` must show
  only `justified-text` and `justified-thai` (both `"type": "text"`) plus README prose. Verified at
  baseline: every table-bearing fixture has a `justify` count of **0**, so no golden is at risk.
- **Red-proof each of these and record what failed:** (1) revert the `decodeStyle` type parameter —
  the table refusal legs must fail while the text legs stay green; (2) revert the Part 3 command arm —
  the property-command table test must fail and the serialize-below-2.0 assertion with it; (3) revert
  the third set and reuse `StyleAlignTokens` for tables — the message assertion must fail on
  `justify` appearing in a table's rejection text; (4) change the fixture's element from `table` to
  `text` — the refusal assertion must fail, proving the test keys on the consumer and not on the key
  path.
- **Matrix Test Audit:** for every I/O matrix row, grep the covering test for a call to the symbol
  under test. A test that re-derives a rule instead of calling it is a false green, not a nit.
- **Mutation-proof the version claim in both directions:** a green suite over correct code proves
  nothing here. Show that the table leg fails when the loader accepts the value, and that the text leg
  fails when `styleVersionRank`'s `:301` justify arm is removed.
- **Demonstrate end to end** that a hand-edited `.folio` with a justified table produces a message
  naming the element and the field at the wasm host boundary — not that a conditional changed.

## Auto Run Result

Status: blocked
Blocking condition: intent gap

### The unanswered question

**Whether to mint a THIRD per-field style diagnostic code — and if so, in which of two forms.**
`internal/diag/diag.go:249-252` reserves this decision verbatim; Story 7.8's epic text
(`epics.md`, inherited item 3) states *"That is a lead call and this story must not settle it
unattended"*; DW-29 states *"That is a lead call, not a builder's"*; and `epic-7-context.md`
records it as an open lead decision. The spec states the three options in Design Notes
(*"The reserved decision"*) and its `Block If` and picks none. It sits in **Part 0** of
Tasks & Acceptance, which gates every other task.

The options, restated for the ruling:
1. Mint a per-field code (e.g. `STYLE_ALIGN_INVALID`), on `CodeStyleLineSpacingInvalid`'s shipped
   precedent — and thereby accept that the registry accretes one entry per style field.
2. Mint one general load-time style code carrying the field as payload. Constraint: AD-14 is
   additive-only, so `STYLE_LINE_SPACING_INVALID` cannot be retired — a general code coexists with
   the per-field one rather than replacing it.
3. Do not mint. Directly contradicts AC4: the refusal is flattened to *"The template could not be
   processed"* at `folio-go/wasm/cmd/engine/main.go:272-281` and never reaches the author in words.

Why this is not rulable here: the outcomes differ observably and **permanently** (a shipped code
string is irreversible under AD-14's additive-only rule), and nothing in the intent selects between
options 1 and 2. D-4.5.1's local discriminator (`diag.go:192-194`) would mechanically favour
option 1, but the reservation deliberately asks a higher registry-policy question than D-4.5.1
answers, so applying it would settle the reserved question rather than resolve it.

### Evidence gathered this dispatch

The full investigation is preserved in `## Code Map` and `## Design Notes` and is **not** repeated
here. Findings that the dispatch did not anticipate:

- **Two of the epic's three test anchors are wrong.** `folio-go/line_spacing_test.go:168-175` with
  its const at `:311-331` does not exist as described — that file contains zero occurrences of
  `justify`. The real site is `folio-go/internal/template/linespacing_test.go:230-237`, const
  `justifyHeaderStyleDoc` at `:479-503`. Corrected in the Code Map.
- **A fourth and a fifth test are affected**, neither listed in the epic:
  `closedsets_test.go:215-275` `TestAlignSetsAreTwoSetsPinnedAgainstTheirMaps` (its name and its
  `len(StyleAlignTokens) == len(ColumnAlignTokens)+1` assertion at `:261-263`), and
  `component_properties_test.go:217-249` `TestStyleAlignPropertyValidatesAgainstTheStyleSetOnly`.
- **The format-version half is free on the file path.** `versionRequiredByContent` runs only at
  save, on a fully validated `*Document`; a loader refusal closes the 2.0 raise by construction.
- **But an in-memory hole exists and the story's own AC1 requires closing it.**
  `component_commands.go:909` allows `align` on a table and the `case "align":` arm `:1048-1071`
  validates with no element-type check, so the designer's *engine* can still stamp a table document
  2.0. Story 7.4 closed only the UI door. Ruled in scope, selected by `IsStyleAlign`'s own doc
  comment (`closedsets.go:75-78`): the property-command path *"must validate it against the same
  single source the loader does"*.
- **No golden is at risk.** Every table-bearing fixture has a `justify` count of 0; the only
  justify-bearing fixtures are the two text goldens the story must preserve.

### Resuming

The spec is otherwise complete and meets the READY FOR DEVELOPMENT standard on every criterion
except **Sufficient**. To resume: record the lead's ruling in the spec (amending `Block If` and
Part 0), then set `status` to `ready-for-dev` and re-dispatch. A re-dispatch against this file
while `status: blocked` will halt at step-01 on `blocked spec supplied`.
