---
title: 'Story 7.8: Refuse a justified table at load, in the author''s own terms'
type: 'bugfix'
created: '2026-08-31'
status: 'done'
baseline_revision: '0da98ecf4438375a38a6e5970783b3c21241be2a'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '{project-root}/_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md'
  - '{project-root}/_bmad-output/specs/spec-folio/folio-format.md'
warnings: ['oversized']
deferred:
  - summary: >-
      A rect, line or image element's style.align "justify" still loads and still
      stamps the document 2.0, which is DW-29's exact pathology one element type
      over from the one this story fixes.
    evidence: |-
      Measured at baseline 7c892f1 (before any of this story's code): a rect element
      carrying style.align "justify" returns from ParseDocument without error and
      versionRequiredByContent reports "2.0". Nothing draws it — page_setup.go projects
      Align for ElementText and ElementTable only. PRE-EXISTING, not caused by this
      story; surfaced because the story's stated rule is "partition by the code that
      consumes the value", which would also cover these three types.
    location: >-
      folio-go/internal/template/parse_bands.go decodeStyle
    severity: medium
  - summary: >-
      folio-go/wasm/cmd/engine/main_test.go is build-tagged js/wasm and nothing in
      go test ./... or in CI ever compiles or runs it, so every assertion about the
      wasm host boundary is dormant.
    evidence: |-
      The file carries //go:build js && wasm. .github/workflows/ci.yml runs go test
      -count=1 ./... with no GOOS/GOARCH; matrix.yml's js/wasm job runs only
      TestTargetRenderHash in the root package. The tests pass when built and run by
      hand under go_js_wasm_exec (verified this dispatch), but no automated path does
      that. Pre-existing — TestWasmHostReportsTheLineSpacingRefusalIntact has the same
      problem — but it is where this story's AC4 evidence would live.
    location: >-
      folio-go/wasm/cmd/engine/main_test.go
    severity: medium
  - summary: >-
      DW-52's bare type assertion in wrapTemplateError means a future wrapped load
      error silently loses its diagnostic code and is flattened at the host.
    evidence: |-
      folio-go/render_error.go uses err.(*template.LoadError) rather than errors.As.
      Harmless while nothing wraps, but the whole value of coding load errors depends
      on the assertion still matching. DW-52 already names its owner as "the story that
      next changes error wrapping in folio-go/internal/template".
    location: >-
      folio-go/render_error.go
    severity: low
  - summary: >-
      Designer TypeScript comments describe a justified table as a value that draws at
      the start edge, which the intended change would make a hard load refusal.
    evidence: |-
      folio-designer/src/App.tsx and src/preview/engine-protocol.ts explain admitting
      justify for any component, including type 'table'. Cannot be corrected by this
      story: the dispatch fences zero paths under folio-designer/.
    location: >-
      folio-designer/src/App.tsx, folio-designer/src/preview/engine-protocol.ts
    severity: low
  - summary: >-
      Documents already carrying a table's style.align/headerStyle.align "justify"
      become permanently unloadable, with no migration path and no stated format
      rule for NARROWING a closed set.
    evidence: |-
      Between Story 7.3 (which admitted justify) and this story, the designer's own
      engine could author such a file through the component_commands.go align arm —
      the in-memory door this story closes. Any file so written now fails ParseDocument
      forever. D-1.4.12 states that EXTENDING a closed set is MAJOR; folio-format.md
      says nothing about removing a member from one, so the version a narrowing lands
      under is unstated. Caused by this change, but the intent's I/O matrix mandates
      the refusal and its Never list excludes new designer work, so a migration path
      is out of scope here.
    location: >-
      folio-go/internal/template/parse_bands.go decodeStyle
    severity: medium
  - summary: >-
      wrapTemplateError passes LoadError.ElementID unbounded into the Diagnostic, and
      the wasm host byte-cuts it at 128, so a multi-byte element id is split mid-rune
      in the elementId field.
    evidence: |-
      folio-go/render_error.go passes le.ElementID straight through; the host applies
      bounded(elementId, 128) — a raw value[:max] byte slice — in wasm/cmd/engine/main.go.
      This is the same "runes, never bytes" property D-7.8.5 rules on, on a field the
      story bounded only inside the MESSAGE. Pre-existing and unchanged by this story:
      the elementId field was populated the same way before it.
    location: >-
      folio-go/render_error.go
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The intent contract below governs the implementation; where the two differ, the
contract wins.*

Until now, an author could tell a table to justify its text. The setting did nothing — every cell
still drew flush to the start edge — but accepting it stamped the whole document with a newer format
number, so it could no longer be opened by any older reader. The author paid the entire cost of an
incompatible file and got no change on the page and no warning.

The document is now refused when it is opened, and the refusal says which element and which field
are at fault and which settings are actually allowed. Justified text on an ordinary text block is
untouched and still works exactly as before.

Read this next part before you upgrade: **a file that already sets justify on a table will stop
opening.** There is no automatic repair and no migration path — the fix is to change that one
setting by hand. Only files written in a narrow window could be affected, and the gap has been
written up and handed to the engineering lead, but the break is real and deliberate rather than an
oversight.

One thing found along the way is worth knowing. Making the refusal reach a person meant letting
these messages through a channel that had previously been silencing them wholesale — which also
meant a document could get large chunks of itself quoted back. Every message is now capped, cut
cleanly, and marked where it was shortened.

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
- **D-7.8.5 AMENDS D-7.8.1's PREMISE — ruled 2026-08-31 after the first build dispatch halted on it.**
  The ruling's stated ground, *"its message never quotes the document"*, was **false**: 7 of
  `newLoadError`'s 105 non-test call sites pass `string(raw)` — an arbitrary JSON sub-object — as
  `value`, so moving `LoadError`s off `TEMPLATE_MALFORMED` switched off `reportableMessage`'s
  reflection guard for that whole population. Measured: a well-formed document with a 2048-byte
  `style` value went from a 35-character refusal to a 512-byte echo of the author's own file.
  **The resolution is to make the premise TRUE, and it is bound at the RENDERING, not at the
  constructor:** `LoadError.Value` stays **complete** in the struct, and `LoadError.Error()` bounds
  the value as it renders it into the sentence. One method, all 105 sites, no per-site judgement.
  **The 7 raw-carrying sites lose nothing** — the full sub-object stays on the error as structured
  data; it is relocated from the prose to the field.
  **Block if** the value is bounded in `newLoadError` instead (that truncates a Go integrator's CI
  evidence to fix a presentation problem — over-broad in the same way the boundary rule was);
  **block if** the bound counts **bytes** rather than **runes**, or can split a rune (a byte bound
  gives Thai and CJK authors a third of the budget — the script-dependence defect ruled on at 7.4,
  two floors down); **block if** the elision is invisible (a truncated value that looks whole is a
  new lie); **block if** the bound is a round number rather than **derived from a stated
  criterion** — *the message must stay dominated by the engine's own words (field, element, reason)
  inside `bounded(message, 512)`*. Put the criterion in the comment and one measured example in the
  story record. Around 96 runes is the expectation, **not the requirement**; a different number
  derived from the criterion is better than this one taken on faith.
  **`Value` is the author-supplied component identified so far. If the story finds another** — a
  `Field` path or a `Reason` that interpolates author content — **it gets the same treatment and the
  story says so.** The lead has already been wrong about the extent of this once; do not assume the
  list is complete.
- **Block if `reportableMessage`'s treatment of `TEMPLATE_MALFORMED` is changed.** The fence stands
  after the correction: genuinely unparseable documents keep their destroyed message and the good
  reason it was written for.
- **Block if `TestWasmHostSanitizesTemplateDiagnostics` carries only one arm.** It must assert
  **both**: *unparseable bytes* → `TEMPLATE_MALFORMED`, message destroyed (still in scope, keep it);
  **and** *parseable-but-invalid with a large `style` value* → the new code, message survives, **and
  the reflected fragment is bounded and visibly elided**. The first dispatch swapped the fixture to
  unparseable bytes only — the one shape still reaching the old guard — which leaves a green test
  measuring the residue. **General obligation, and this is its third occurrence in this epic:** when
  a change moves a population **out** of a guard's scope, the guard's test must be re-pointed to a
  member still in scope **and** the departed population asserted under its new treatment.
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

- **Implement D-7.8.5's bounding, in `LoadError.Error()`.** Runes, not bytes; never split a rune;
  visible elision marker; bound derived from the stated criterion with the derivation in the comment
  and one measured example in the story record. Assert it with an author-supplied value that is
  **multi-byte** — a Thai or CJK `style` value — so a byte-counting regression reddens rather than
  passing on ASCII.
- **Give `TestWasmHostSanitizesTemplateDiagnostics` both arms** per the `Block If` above, and prove
  the parseable arm non-vacuous by mutation.
- **File, do not fix:** `cmd/folio` already prints `err.Error()` with the full value to a terminal
  today, so terminal-escape content in a `.folio` is a **pre-existing** property of the CLI that this
  story neither creates nor fixes. One `deferred-work.md` line, not work in this story.

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

**2026-08-31 — one Code Map claim was measured and found half-wrong; scope unchanged.**

The Code Map states *"THE IN-MEMORY HOLE, and it is real … without this arm, AC1's 'never raised to
format 2.0 for that value' is false and the designer can author a file it cannot reopen."* That was
true at baseline. It stops being true the moment **Part 2** lands, because
`applyComponentProperties` (`component_commands.go:598-614`) and `wasm.Engine.Apply`
(`wasm/engine.go:236-252`) both **serialize and re-parse before installing** — so a document the
loader refuses cannot be installed however the `case "align":` arm behaves. Red-proof (2) was run as
written and confirms it: reverting the Part 3 arm leaves the command refused and the serialized
document at `1.0`; only the **message** assertion reddens.

**Part 3 was therefore implemented as specified, and it is not scope creep — but its product is the
located message, not the version closure.** Without it the author gets `component properties did not
pass format validation` for one bad value, naming neither the field nor the legal values, which is
the same class of defect this story exists to fix one layer down. It is also the layer that does not
depend on the round trip continuing to exist. Both facts are now recorded in the arm's own comment
and in `TestAPropertyCommandCannotStampATableDocumentAt2_0`'s doc, so no future reader mistakes that
test's green for a red-proof of the arm — `TestStyleAlignPropertyValidatesAgainstItsConsumersSet`'s
table leg is the live detector.

**Consequence for Part 0's blast radius, which the spec did not enumerate.** Making `newLoadError`
supply `TEMPLATE_FIELD_INVALID` moved every general load error off `TEMPLATE_MALFORMED`, which
falsified three shipped assertions the spec's falsified-test list did not name:
`render_error_test.go`'s `malformed template` cases (two) and
`wasm/cmd/engine/main_test.go`'s `TestWasmHostSanitizesTemplateDiagnostics`, plus the census's
`TEMPLATE_MALFORMED` trigger. None was deleted. Each was **re-pointed at a fixture that is still
genuinely malformed** — `unparseableTemplateJSON`, bytes that are not a `.folio` document — so the
destruction rule those tests exist to pin is still pinned, on the population that still carries it.
A `TEMPLATE_FIELD_INVALID` case was added beside each.

**2026-08-31 — D-7.8.5 implemented, and the ruling's "identify any other author-supplied
component" clause paid out THREE more.**

The bound lives in `LoadError.Error()` (`folio-go/internal/template/errors.go`). The struct fields
stay complete; only the sentence rendered for a person is bounded; `reportableMessage`'s treatment of
`TEMPLATE_MALFORMED` is untouched.

**The criterion, stated once and derived four ways.** *The message must stay dominated by the
engine's own words — the sentence frame, the field, the element and the reason — inside the wasm
host's `bounded(message, 512)` window.* The four bounds below were **derived** under the premise
that a rune costs at most 3 bytes, AD-12's closed locale set (`en`, `th`, `zh-Hans`, `ja`) being
entirely within the BMP. **That premise was itself corrected during the review pass and no longer
holds** (`errors.go` records the correction in full): the locale set constrains the document's
*language*, not the arbitrary JSON flowing through these fragments, where one non-BMP rune costs
`utf8.UTFMax` = 4 bytes. The four rune bounds are therefore a **fair-allocation** device — every
script gets the same number of runes — and not a byte guarantee. The byte guarantee is the
message-level `loadErrorMessageBytes` bound added by the review pass.

| Fragment | Bound | Derivation |
|---|---|---|
| `Value` | **84 runes** | Wholly the author's, and the fragment that regressed: it takes at most **half** the 512-byte window, so `3N + 3` (the elision marker) `<= 256` → `N <= 84.33`. |
| `ElementID` | **24 runes** | An id is short **by contract** (AD-10). The longest LEGAL one is 14 runes — `e` plus the 13 base-36 digits the int64 counter ceiling allows — so a longer one is by definition not an id. |
| `Field` | **96 runes** | The longest path the format can legitimately produce is `assets.<64-hex>.mediaType` = **81 runes**. No engine-authored path is ever truncated; a runaway author key is. |
| `Reason` | **256 runes** | The longest reason this loader can author is the asset-digest mismatch at **exactly 200 runes** (it names two 64-character digests), plus headroom for the stdlib text the `must be a …: ` arms append. |

**Around 96 was "the expectation, not the requirement"; 84 is the number the criterion actually
yields, and it is used.**

**Three more author-supplied components, found by measurement rather than by reading the format
string — which is the mistake D-7.8.5 exists to correct.** Measured at this story's tree before the
bound:

| Vector | Document | Message |
|---|---|---|
| `Value` (the known one) | a well-formed `.folio` whose `style` key holds 2048 Thai characters | **6,323 bytes**; the host reported 512 of them, **cut mid-rune** |
| `ElementID` **+ `Value` + `Reason`** | `"id"` holding 4096 characters | **12,424 bytes** — `claimID` passes the raw id as ElementID, as Value, and again inside the Reason via `validateElementID`'s `%q`, so one runaway id was reflected **three times in one message** |
| `Field` | an `assets` key of 4096 characters | **4,263 bytes** — `"assets."+k` splices an author-supplied object key |

`Reason` also interpolates a table's own collection path at `parse_bands.go`'s `footerOf` prefix
check. Nine `newLoadError` call sites (not seven) pass `string(raw)` as `value`.

**The measured example the ruling asked for.** The same 2048-Thai-character `style` document, through
the real wasm host under `go_js_wasm_exec`:

```
code=TEMPLATE_FIELD_INVALID
430 bytes, 83 author runes reflected, valid UTF-8, visibly elided
"template: field style (element e1): must be an object: template: expected a JSON object:
 json: cannot unmarshal string into Go value of type map[string]json.RawMessage (value: \"กกก…83 runes…\")"
```

**Corrected at close, 2026-08-31 (re-derived, not relayed).** An earlier version of this line read
*"347 of those 430 bytes are the engine's own words"*. That subtracted the author's **rune** count
(83) from a **byte** total (430). Re-measured: the 83 Thai runes occupy **249 bytes**, the elision
marker 3, so the engine's own words are **178 of 430 bytes (41%)** — the author's fragment is the
larger half of *this* message. What the design actually guarantees is the narrower claim the
derivation supports: the value can never exceed **256 bytes, half the host's 512-byte window**, and
no engine-authored word is ever truncated. The looser gloss — that the engine's words *dominate* any
given message — does not follow, and is not relied on anywhere. The whole message now fits inside
the host's window rather than being cut by it. Baseline: **35 characters**, all of them the
placeholder.

**`TestWasmHostSanitizesTemplateDiagnostics` carries both arms**, per the `Block If`. Arm one keeps
the destruction rule on the population that still carries it (unparseable bytes → `TEMPLATE_MALFORMED`,
message replaced). Arm two asserts the departed population under its new treatment (parseable-but-invalid
→ `TEMPLATE_FIELD_INVALID`, message survives naming element and field, reflection bounded in runes and
visibly elided). Arm two was proved non-vacuous twice by mutation — see *Auto Run Result*.

**`cmd/folio` filed, not fixed**, as instructed: `deferred-work.md` **DW-53**. The bound caps the
*volume* a hostile document pushes through the CLI, which is a side effect and not a fix —
**bounding is not escaping**, and a short escape sequence passes a length bound untouched.

## Review Triage Log

### 2026-08-31 — Review pass (re-dispatch)
- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 1, medium 1, low 3)
- defer: 2: (high 0, medium 1, low 1)
- reject: 7
- addressed_findings:
  - `[high]` `[patch]` `LoadError.Error()` bounded its four author-supplied fragments
    independently but never bounded the ASSEMBLED message, while D-7.8.5's criterion is
    stated over the message — *"inside `bounded(message, 512)`"*. Measured: a document
    whose element id is 2048 Thai runes produced a **1139-byte** message (`claimID`
    reflects the id as `ElementID`, as `Value`, and again inside `Reason`), which the
    host's `bounded(message, 512)` byte-slices to **invalid UTF-8 with no elision
    marker** — the two things D-7.8.5 explicitly blocks on. Added `boundBytes` and
    `loadErrorMessageBytes = 512` (copied from the host's own window, not chosen), applied
    last in `Error()`, cutting on a rune boundary with the marker's bytes budgeted.
    Same document now yields **511 bytes**, valid UTF-8, visibly elided; the host's cut
    can no longer fire. Also corrected the constants' false premise that *"a rune costs at
    most 3 bytes for every script the format admits"* — the AD-12 locale set constrains the
    document's language, not the arbitrary JSON flowing through these fragments, where one
    non-BMP rune is `utf8.UTFMax` = 4 bytes. New `TestLoadErrorMessageFitsTheHostWindow`
    (three legs, Thai content) and a third host arm.
  - `[medium]` `[patch]` `internal/template/linespacing_test.go:239` claimed
    `justifyHeaderStyleDoc` *"was REWRITTEN onto a text element"*; the const at `:499` is
    still a table, as its own doc comment at `:494` and the assertion above it require.
    Rewritten to state which of the story's two options was actually taken and why: the
    const was neither deleted nor rewritten, but kept as a table and repurposed into the
    refusal fixture.
  - `[low]` `[patch]` `component_commands.go:1073` and `component_properties_test.go:600`
    both named `applyComponentProperties`, which does not exist. The real symbol is
    `updateComponentProperties` (`component_commands.go:592`).
  - `[low]` `[patch]` `errors_test.go:107` doc comment read
    `TestTheOverRIDINGConstructorStillWins` above `func TestTheOverridingConstructorStillWins`.
  - `[low]` `[patch]` `folio-format.md:294` *"Neither does a table's own `style.align` or…"*
    → *"Nor does…"*.

**Rejected (7), with the authority each was tested against.** (1) *"The widened tripwire
lost the style set's exact-size pinning"* — **false**: `TestAlignSetsAreThreeSetsPinnedAgainstTheirMaps`
compares each set against an explicit `want` literal for both length and order, so a fifth
member reddens. (2) *"Arm 1 should reuse `unparseableTemplateJSON`"* — mechanically
impossible: that const is unexported in package `folio`, and `main_test.go` is package
`main` under `wasm/cmd/engine`. (3) *"`folio-format.md` promotes the rect/line/image gap to a
specification"* — `closedsets.go:45-47`'s invariant requires the document to describe the set
actually enforced; the gap itself is already carried in `deferred:`. (4) *"New tests repeat
DW-52's bare `err.(*LoadError)`"* — a package-wide pre-existing pattern (`ids_test.go:251`,
`keeptogether_test.go:192` both predate this story) that DW-52 files and the spec explicitly
defers. (5) *"`TestTheOverridingConstructorStillWins` calls the constructor rather than a
production trigger"* — the override's live sites are covered by the census trigger and the
line-spacing and footer-source tests. (6) *"The census trigger's whitespace-exact
`strings.Replace` is brittle"* — it fatals with an explicit fixture-precondition message
naming the cause. (7) *"`assertBoundedFragment` is coarse"* — the real gap it gestures at, no
message-level assertion, is closed by the high patch above.

**Not routed as an intent gap, and why.** The intent-alignment layer framed a genuine fork:
whether *"partition by the code that CONSUMES the value"* refuses `justify` on `rect`, `line`
and `image` too. The contract's I/O matrix enumerates only `table` and `text` rows, and its
Never list forbids *"a blanket ban wearing a narrow name"*, so the matrix settles the scope at
table-vs-text. The gap was already measured at baseline and already carried in `deferred:` by
the planning dispatch; it is pre-existing, not caused here.

### 2026-08-31 — Review pass
- intent_gap: 1: (high 1, medium 0, low 0)
- bad_spec: 0
- patch: 7: (high 0, medium 1, low 6)
- defer: 4: (high 0, medium 2, low 2)
- reject: 5
- addressed_findings:
  - none

The first dispatch's implementation (1886 lines of diff) was **reverted in full** and the tree
returned to `7c892f1`. It was reverted not because the code was wrong but because the ruling it was
built on rested on a false premise: D-7.8.1 moved every `LoadError` off `TEMPLATE_MALFORMED` on the
stated ground that *"its message never quotes the document"*, and nine `newLoadError` call sites
pass an arbitrary JSON sub-object as `value`. Shipping it would have switched off the host's
reflection guard for that whole population. The halt sent the premise back to the lead, which
produced **D-7.8.5**, and the work was re-derived on top of the corrected ruling in the re-dispatch
below. Patch and defer findings from this pass were moot under the cascading rule and were not
applied. *(The reverted diff was briefly committed as a `.patch` artifact; it was removed at the
story's close — a superseded, reverted attempt does not belong in history, and this paragraph is the
record of it.)*

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

### The correction to that ruling's premise (D-7.8.5, 2026-08-31)

The first build dispatch implemented D-7.8.1 exactly as ruled, and it worked — then halted, because
**the ruling's stated factual ground was false**. This section records why, because the mechanism is
more reusable than the fix.

**The false premise.** The ruling argued the wasm boundary rule was *"over-broad by accident, not by
design"* on the ground that *"a `LoadError` is not [document-quoting]"*. The lead had read
`LoadError.Error()`'s **format string** and concluded about the **data flowing through it** —
without checking what callers pass as `value`. Its own words on the error: *"That is the same
failure I recorded against myself at Epic 4 — measuring one population and reporting on a wider
one."* `reportableMessage`'s comment — *"that message quotes the offending document back, so a large
or hostile one would be reflected instead of described"* — was **right**, and the ruling was wrong
about it.

**Measured at `7c892f1`:** 7 of `newLoadError`'s 105 non-test call sites pass `string(raw)`, an
arbitrary JSON sub-object, as `value` — `parse_bands.go:583`/`:718`/`:749`,
`decodehelpers.go:158`, `parse.go:82` and two more. A well-formed document with a 2048-byte `style`
value: baseline `TEMPLATE_MALFORMED`, 35-character message, no reflection; with the change,
`TEMPLATE_FIELD_INVALID`, 512-character message, **reflected**.

**The fix makes the premise true rather than working around its falsity.** The premise was a claim
about **the message**; bounding in `Error()` makes it true *of the message*, which is the thing that
gets reflected. Bounding `Value` in the constructor was rejected for a reason worth keeping: **it
repeats the mistake it would be fixing.** A Go integrator's CI log legitimately wants the whole
offending JSON, and the hazard exists only where the message is rendered to a person who may not
have authored the document — so truncating the struct field is over-broad in precisely the way the
boundary rule was. One over-broad rule is not corrected by writing another.

**Why this was not escalated to the owner**, which the lead nearly did, since a reflection question
with no threat model (PRD §13 records that MVP deliberately has none) is normally the owner's:

1. **The project has already decided this exact question, in this exact file.** `main.go`'s
   `ENGINE_REJECTED` path does `bounded(err.Error(), 512)` and returns it, reasoning: *"The engine
   authored this text about a template the caller already holds. Withholding it left the panel with
   nothing to act on, so report it bounded."* So bounded, engine-authored text about a document the
   caller already holds **is** reported — established direction. `TEMPLATE_MALFORMED`'s exception is
   not an exception to that principle; it exists because that message **quotes** the document, which
   is different from **mentioning** it. This ruling brings `LoadError` into compliance with the
   principle rather than asking for an exception to it, so no new risk is accepted and there is
   nothing for the owner to accept.
2. **Measured: no injection vector.** `grep -rn "dangerouslySetInnerHTML\|innerHTML"
   folio-designer/src/` returns nothing — diagnostics reach React as text nodes and are escaped. The
   residual reflection is inert display of the user's own file on the user's own screen, with no
   server and no third party anywhere in the product.

**What would reopen it** — recorded, not acted on: a rendering surface that **interprets** rather
than escapes (any `innerHTML`, a Markdown renderer, a `title`/attribute sink), or **FR45's REST
service**, which PRD §13 names as the thing that brings a threat model with it. At that point
bounded reflection of author content becomes an owner question with a standard to judge against.

**Two things the lead explicitly did not claim**, so this is not over-read: it audited two injection
spellings, not every sink in the designer; and `cmd/folio` already prints the full value to a
terminal today, so terminal-escape content in a `.folio` is a pre-existing CLI property this story
neither creates nor fixes — a `deferred-work.md` line, not work here.

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

Status: done

**This supersedes the previous (blocked) Auto Run Result.** The intent gap it raised was ruled on as
**D-7.8.5** and written into the intent contract before this dispatch; the ruling is implemented, and
the premise D-7.8.1 rested on is now true of the message rather than merely asserted about it.

### What was done

The first dispatch's reverted implementation was re-derived as the `Resuming` section directs, and
D-7.8.5 was implemented on top of it: `LoadError.Error()` bounds every
author-supplied fragment in **runes**, never splitting one, with a visible `…`; the struct fields stay
complete. Three author-supplied components beyond `Value` were found by measurement and given the same
treatment (see the Spec Change Log entry above). `TestWasmHostSanitizesTemplateDiagnostics` was given
both arms.

### Verification — measured, not "green"

| Command | Result |
|---|---|
| `go test -count=1 ./...` (folio-go) | **1575 pass, 5 skip, 2 fail** — both fails are `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)`, `got 7, need >=20`, the mandated permanent red (parent + subtest). Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` **PASS**. |
| `go vet -tags=matrix ./...` | exit 0, no output |
| `gofmt -l folio-go` (from repo root) | no output |
| `TestTargetRenderHash`, four legs, `FOLIO_MATRIX_TARGET` exported | `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` — each **1 PASS, 0 FAIL, 0 `asserts NOTHING`** |
| unset control | **1 `asserts NOTHING`** — the four legs are not no-ops |
| `TestCrossTargetByteIdentity -tags=matrix` | PASS (21.1s), all four targets from one process |
| `lint` module `go test ./...` | 4 packages ok |
| designer `npm run typecheck` | clean |
| designer `npm run lint` (oxlint) | **exactly 4 `only-export-components` warnings, 0 errors** — the baseline |
| designer `npm test` | **33 files, 280 tests pass** — unchanged, this story adds none |
| designer `npm run test:e2e:compile` | clean. **`tsc --noEmit` only; browser e2e is deferred by D-000.4 and did NOT execute.** |
| wasm host tests under `go_js_wasm_exec` (node v24.16.0) | 5 pass, 0 fail — **executed, not merely compiled** |

**All twenty-one `goldenDigestRecord` digests verified byte-identical by hashing the shipped
`fixtures/<dir>/expected.pdf`: 21 OK, 0 moved.** `git status fixtures/` clean before quoting.

**Manual checks.** `git diff --name-only`: **0 paths** under `folio-go/internal/layout/`, **0** under
`folio-designer/`. `README.md` md5 still `078d7d80d518d54af2fc04fb270d46b8`.
`fixtures/statement-signoff.json` unmodified. `grep -rln justify fixtures/` returns only
`justified-text` and `justified-thai` (`"type": "text"` in both) plus README prose in those two and
`statement-20`.

### Red-proofs, each run and each recorded

| Mutation | What reddened |
|---|---|
| **(1)** `decodeStyle` no longer selects by element type | `TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage`, `TestATableStyleJustifyIsRefusedBeforeAnyVersionIsComputed` (both legs), `TestLoadErrorsCarryFieldValueAndTheGeneralCode/table-style-justify`, `TestVersionForSaveIsRaisedOnlyByContent`, `TestATableWhoseCascadedAlignIsJustifyIsRefusedAtLoad`, the census. **Every text-justify leg stayed PASS** — verified subtest by subtest. |
| **(2)** Part 3 command arm reverted to the single predicate | `TestStyleAlignPropertyValidatesAgainstItsConsumersSet`. As the earlier Spec Change Log entry records, `TestAPropertyCommandCannotStampATableDocumentAt2_0` stays green because the round trip re-parses — the message is the arm's product, and that test's doc says so. |
| **(3)** table message derived from `StyleAlignTokens` | `TestAlignClosedSetsRejectAtTheRightSiteWithTheirOwnMessage`, `TestATableWhoseCascadedAlignIsJustifyIsRefusedAtLoad` — `justify` appears in a table's rejection text |
| **(4a)** fixture element `table` → `text` | refusal assertions fail: `refusal = field "width" element "e1", want field "style.align"` |
| **(4b)** partition by KEY PATH (`fieldPrefix == "headerStyle"`) instead of consumer | `a table carrying justify at style.align must be refused at load` — **the sharp proof: the tests key on the consumer, not on the key path, which is the exact defect D-7.3.1 shipped** |
| **(V)** `styleVersionRank`'s justify arm deleted | `TestVersionForSaveIsRaisedOnlyByContent/justify_alone,_1.0_in` (`"1.0", want "2.0"`), `/justify,_1.1_in`, `TestNewContentIsNeverStampedWithTheLibraryCeiling/the_2.0_alignment`. The version claim is mutation-proved in **both** directions. |
| **(C)** `newLoadError` reverted to leave `Code` empty (AC41 re-measurement) | `TestLoadErrorsCarryFieldValueAndTheGeneralCode` on **6 of 7** cases, `TestDiagnosticRegistryErrorCensus/TEMPLATE_FIELD_INVALID`, `TestFourErrorModesCarrySeverityErrorDiagnostics/invalid_template_field`, and at the host `TestWasmHostSanitizesTemplateDiagnostics/parseable-but-invalid…` (`code = "TEMPLATE_MALFORMED"`) plus `TestWasmHostReportsTheTableJustifyRefusalIntact`. The re-pointed enumeration test is **not** a dead detector. |
| **(B1)** `boundRunes` counts BYTES | all four bounding legs: *"only 27 runes survived a bound of 84 … the bound is counting BYTES"*, and *"the message is not valid UTF-8: a bound split a rune"* |
| **(B2)** bounding removed entirely | *"the message reflected 2048 runes of the author's own document; the bound is 84"*; at the host, *"the host reported 111 runes … the bound is 84"* |

### AC4, end to end at the wasm boundary

Executed under `go_js_wasm_exec`, not merely compiled:

```
code=TEMPLATE_FIELD_INVALID
message="template: field style.align (element e1): not one of the closed set left, center, right (value: justify)"
; the same value on a text element loaded
```

### Nothing left blocking

Deferred, unchanged and recorded in this spec's frontmatter: the rect/line/image `justify` gap, the
dormancy of `main_test.go` in CI (it is build-tagged `js && wasm`; run by hand this dispatch),
DW-52's bare type assertion, and the designer's stale TypeScript comments. **DW-53 is new** and is
this story's "file, do not fix" line.

### Review pass addendum (2026-08-31, re-dispatch)

Four review layers ran in parallel over the diff since `0da98ec`. **One high-severity patch was
applied after the implementation commit**, plus four low/medium ones; see the `## Review Triage Log`.
The headline finding was corroborated independently by three of the four layers and then reproduced
by the orchestrating dispatch before being routed:

> `LoadError.Error()` bounded its four author-supplied fragments but never the assembled message.
> A document whose element id is 2048 Thai runes rendered a **1139-byte** message, which the wasm
> host's `bounded(message, 512)` byte-sliced into **invalid UTF-8 with no elision marker** — both of
> the conditions D-7.8.5 blocks on, at the surface AC4 names. Now **511 bytes**, valid UTF-8, visibly
> elided, with the host's cut unable to fire.

**Verification re-run in full after the patches** (measured, not adjectives):

- `go test -count=1 ./...` — **1579 pass, 5 skip, 2 fail**, both the mandated permanent
  `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` red (got 7, need >=20). One distinct
  failure; drift twin `TestCorpusP6StatsMatchDeclaredBaseline` **PASS**.
- `go vet -tags=matrix ./...` exit **0**. `gofmt -l folio-go` from the repo root: **no output**
  (repo-wide, only the known DW-23 `lint/internal/rules/licencegraph_test.go` reports).
- `TestTargetRenderHash` per leg with `FOLIO_MATRIX_TARGET` exported — `darwin/arm64`,
  `linux/amd64`, `linux/arm64`, `js/wasm`: each **1 PASS, 0 `asserts NOTHING`**. Unset control:
  **1 `asserts NOTHING`**, proving the four legs are not no-ops.
- `TestCrossTargetByteIdentity` **PASS** (21.5s, all four targets from one process).
- `cd lint && go test -count=1 ./...` — 4 packages **ok**.
- Designer: typecheck **0**, oxlint **0 errors / exactly 4** `only-export-components` warnings,
  **280 tests / 33 files pass**, `test:e2e:compile` **0** (`tsc --noEmit` only — browser e2e is
  deferred by D-000.4 and **did not execute**).
- `wasm/cmd/engine` **executed** under `go_js_wasm_exec`, not merely compiled: 5 tests / 3 subtests
  pass, printing the located refusal at the host and both bounded-reflection measurements.
- Heavy tests with `FOLIO_HEAVY=1`: the three table-shaped gated tests plus all four
  `TestStatementGoldenFixtures` subtests **pass**.
- **All 21 `goldenDigestRecord` digests re-verified by SHA-256 against the shipped `expected.pdf`
  after the patches: 21 match, 0 moved.** `git status fixtures/` clean; `statement-signoff.json`
  untouched and unmodified since `36bb3f5`.

**Mutation proofs reproduced independently by the dispatch** (not accepted on report):

| Mutation | Result |
|---|---|
| `boundRunes` counts bytes instead of runes | host arm 2 red — *"only 27 runes survived: the bound is counting BYTES, not runes"*; arm 1 stayed green, so the two arms measure different populations |
| `newLoadError` leaves `Code` empty | AC41's re-point red on **6 of 7** cases plus the coverage witness — *"zero \*LoadErrors reached the CODE assertion, so the re-pointed half of this test measured nothing (D-7.4.2)"* |
| `boundBytes` neutralised to identity | both new message-window legs red at **1142** and **1050** bytes; the under-the-bound leg correctly stayed green |

Every mutation was restored and confirmed byte-identical by `cmp`.

**Scope fences confirmed mechanically:** zero paths under `folio-go/internal/layout/` and zero under
`folio-designer/` in the change set; `README.md` in no commit, md5 still
`078d7d80d518d54af2fc04fb270d46b8`; every `justify` in `fixtures/` sits on a `"type": "text"` element
or in README prose.


## Delivery Log

### 2026-08-31 — planned

Dispatched from Epic 7 against baseline `7c892f1` to close DW-29. The plan gate re-verified every
anchor the epic text carried and found three wrong (D-7.8.4) — the second time this run an epic's
anchors had rotted, which is where *an anchor written at a plan gate is a claim with an expiry date*
was recorded. The gate also found a **second door** the epic had not listed: the property-command
path could stamp a table document `2.0` from the designer's own engine, so AC1 was not satisfiable
without it. The reserved diagnostic-code question went to the lead and came back as **D-7.8.1**: one
general code supplied by the constructor, not a per-field code.

### 2026-08-31 — built

**The first dispatch halted, and that was the right halt.** D-7.8.1's ruling rested on a stated
ground — *a load error's message never quotes the document* — and the build measured it FALSE: nine
constructor call sites pass an arbitrary JSON sub-object as the value. Implementing the ruling as
written would have switched off the host's reflection guard for that entire population; a well-formed
document with a 2048-character style value went from a 35-character refusal to 512 bytes of the
author's own file, cut mid-rune. The build **reverted 1886 lines rather than working around its own
ruling**, and sent the premise back. That is the behaviour the halt exists for: the cheap move was to
special-case the nine sites and ship.

**The lead's answer moved the seam.** D-7.8.5 made the premise true instead of routing around it, and
located the bound at the **rendering** rather than at the **constructor**. The struct keeps the
complete value for a Go integrator's CI log; only the sentence rendered for a person is bounded. One
method, all call sites, no per-site judgement — where bounding in the constructor would have
truncated a datum to fix a presentation problem, over-broad in the same way the original boundary
rule was. The story then found **three more** author-supplied fragments beyond the one the ruling
named, by measurement rather than by reading the format string.

**The review layers converged on what the per-fragment bounds still missed.** Three of four layers
independently raised it: four bounds sharing no budget do not discharge a criterion stated over the
assembled message. Re-derived at close rather than accepted — a document whose element id is 2048
Thai runes reflects that id three times in one sentence and produced **1139 bytes**, which the host's
byte slice cut into **invalid UTF-8 with no elision marker**, both conditions the ruling blocks on, at
the surface AC4 names. Now **511 bytes**, valid UTF-8, visibly elided.

**D-7.8.7 is the entry with reach beyond this story.** The first dispatch had narrowed a guard's
fixture to the one shape still tripping the old guard, leaving a green test measuring the residue —
the third instance of that shape in one epic, which makes it a rate rather than an anecdote. The
standing obligation: *when a change moves a population out of a guard's scope, the guard's test must
be re-pointed to a member still in scope AND the departed population asserted under its new
treatment.*

### 2026-08-31 — done

Baseline `0da98ec`; commits `93c9bdd` and `7ab6a09`, closed at `HEAD`. Decisions applied: **D-7.8.1**
(one general code from the constructor), **D-7.8.4** (three anchor corrections plus the second door),
**D-7.8.5** (bound at the rendering, in runes), **D-7.8.7** (both guard arms). **D-7.8.2**'s audit of
the two existing style codes was deliberately NOT done — it is triggered by the `folio-go/v0.1.0`
tag, and both codes were verified unretired and unchanged in meaning at close.

Findings across two review passes, and the two passes do not simply add up: the first pass raised
7 patch / 4 defer / 5 reject, but its patch and defer findings were **moot and never applied** — the
intent gap it also raised reverted the code they described. The findings that actually landed are the
re-dispatch's: **5 patched (1 high) / 2 deferred / 7 rejected**. Two
rejections were spot-checked at close and both refute the specific claim at the cited location; the
widened tripwire does still pin each set's exact length and order, and the suggested fixture reuse is
mechanically impossible across the package boundary.

**Gates measured at close, not relayed.** Full Go suite: 13 packages ok, **exactly one distinct red**
— the mandated `P6g_(opaque_names)` floor, got 7 need >=20, untouched. `go vet -tags=matrix` exit 0,
no output; `gofmt -l folio-go` from the repo root, no output (only the known DW-23 file reports).
Matrix render hash on all four legs with the target exported — darwin/arm64, linux/amd64,
linux/arm64, js/wasm — each PASS with **0 `asserts NOTHING`**, against an unset control that printed
`asserts NOTHING` and asserted nothing, proving the legs are not no-ops. Cross-target byte identity
PASS. Lint module **117 tests / 4 packages**. Designer **280 tests / 33 files**, typecheck exit 0,
oxlint exactly 4 baseline warnings and 0 errors, e2e **compile only** — the browser suite is deferred
by D-000.4 and **did not execute**. The wasm host tests were **executed** under `go_js_wasm_exec`, not
merely compiled: **6 tests / 3 subtests pass**. **All 21 golden digests byte-identical by SHA-256
against the shipped artifacts; none moved.**

Four claims were re-derived by mutation rather than read from the report. Making the value bound count
bytes reddens the host's second arm at 26 surviving runes — so the multi-byte assertion has real
teeth. Removing the message-level bound reproduces the exact pathology: invalid UTF-8 ending in a
partial rune with no marker. Stripping the general code from the constructor reddens 6 of 7
enumerated cases plus the coverage witness, and both host arms with it. Five tests moved and **zero
were deleted** — four renames verified 1:1 by diffing test-function names across the commits, and the
fifth inverted in place with its table fixture repurposed rather than rewritten.

**Two record corrections made at close.** The measured example claimed *347 of 430 bytes are the
engine's own words*; that subtracted a rune count from a byte total, and the true figure is **178 of
430 (41%)** — the author's fragment is the larger half of that particular message. The guarantee the
derivation actually supports is narrower and still holds: the value can never exceed half the host's
window, and no engine-authored word is ever truncated. Corrected in both this spec and the DW-29
closure note. The spec also still asserted the BMP premise that the review pass had already recorded
as false in the code; corrected to match. A comment in the diagnostic registry restated D-7.8.1's
falsified premise flatly, so the next reader would have re-derived it — annotated to say the premise
holds only because the rendering bound makes it hold.

Deferred with owners: **DW-54** (a file already carrying a table justify is permanently unloadable,
and the format doc has a rule for extending a closed set but none for narrowing one) — **the
engineering lead**, live now rather than at the tag, because D-7.8.3's before-the-tag set already
holds two more narrowings. **DW-55** (the element id reaches the host unbounded and is byte-cut at
128) — the story that next changes diagnostic construction, or Story 15.3 before the tag. **DW-53**
(the CLI prints the full error to a terminal) was filed and not fixed by instruction. All three were
found recorded only in the spec's frontmatter and were filed into the standing register at close.

Housekeeping: the first dispatch's reverted 1886-line attempt had been committed as a 109KB `.patch`
artifact. It is removed in the closing commit — forward, not by rewriting history — and the spec's
two references to it are now prose that carries the same record.

**Epic 7 stays `in-progress`.** Stories 7.9 and 7.10 are open, and 7.9 gates the boundary (D-7.7.8).
