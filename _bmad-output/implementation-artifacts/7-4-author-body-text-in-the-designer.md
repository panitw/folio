---
title: 'Story 7.4: Author body text in the designer'
type: 'feature'
created: '2026-08-30'
status: 'done'
baseline_revision: '813a414e12198be86d28f61af741b56fd93fb40e'
review_loop_iteration: 0
followup_review_recommended: true
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
  - '{project-root}/_bmad-output/implementation-artifacts/deferred-work.md'
warnings: ['multiple-goals', 'oversized'] # multiple-goals: DW-29 (the load-time refusal of `style.align: "justify"` on a table) is separably shippable from this story's designer surface and from DW-25, so the plan gate ROUTES IT OUT to a named Story 7.8 — see Design Notes, "The DW-29 judgment". DW-25 is NOT a second goal: it is AC1's own precondition. oversized: the DW-25 bounds rework, the four-way TypeScript mirror, and the designer editor are three wide surfaces that must be stated, not summarised.
deferred:
  - summary: >-
      The property-command encoder splices the author's typed text into the command JSON unquoted,
      so a non-numeric line-spacing or point value produces malformed bytes instead of a located
      engine error.
    evidence: |-
      `component-property-command.ts` routes `pointFields` and the new `ratioFields` through
      `rawNumberLiteral`, which returns the typed string verbatim: typing `abc` emits
      `{"op":"set","value":abc}`, which fails JSON parsing and yields a generic refusal rather than
      the field-located message the panel is built to show. The pattern is PRE-EXISTING for the
      point fields and the spec directed following it for line spacing, so it is not caused by this
      story - but this story widened the set of fields on that path.
    location: >-
      folio-designer/src/component-property-command.ts:28
    severity: medium
  - summary: >-
      When a text element's very first packed line already exceeds the per-line fragment guard, the
      element paints ZERO lines - the author sees an empty box whose only signal is the truncation
      notice.
    evidence: |-
      Measured during review: that path yields `Truncated=true, len(Lines)==0`. This is conformant
      with the contract as written (painting "stops at the last whole line that fits", and no whole
      line fits), and the state is still distinguishable from empty via the flag, so production
      behaviour was deliberately NOT changed. Whether such a line should paint a partial prefix is a
      design question the contract does not settle.
    location: >-
      folio-go/page_setup.go - the paint loop's per-line guard
    severity: medium
  - summary: >-
      The canvas renders one DOM span per fragment for every projected line, so a document at the
      new bounds can build tens of thousands of nodes with no virtualisation.
    evidence: |-
      The bounds this story raises make up to 1920 lines and 65536 fragments projectable where 256
      and 512 were the ceiling before, and the painting code maps every line unconditionally. Not a
      correctness defect and not newly introduced, but the scale at which it can now be reached is
      new.
    location: >-
      folio-designer/src/App.tsx - the textPaint painting path
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Non-normative: this section settles nothing. It describes what the story sets out to do.*

Today a contract clause cannot be typed into the designer at all. The one place to enter a component's words is a single-line box, which cannot hold a paragraph break; and the drawing surface refuses outright any text longer than about eighty English words — or about a hundred and seventy characters in Thai, where each character costs three times as much. Past that limit the whole page goes blank rather than showing what it can.

This story opens both. The text box becomes a real editor that keeps the lines an author types, and keeps the paragraph breaks of a clause pasted from a word processor while quietly dropping the pasted fonts and bold. Line spacing and alignment, justification included, join the existing typography controls. The refusal limits move up to figures derived from the epic's own forty-page target, and past those the drawing degrades honestly: it paints what it can and says it has been cut, instead of showing nothing at all.

The alignment control offers justification only for text, never for a table, where the setting means nothing. Making the engine refuse a hand-edited file that says otherwise is a separate change and becomes its own story.

Nothing here changes how pages are decided, and the canvas still draws a single page. One test stays red by design: a standing marker, not a failure.

<intent-contract>

## Intent

**Problem:** A multi-paragraph clause cannot be authored. The inspector's only content control is a single-line `<input>`, which cannot hold a line feed; line spacing has no control at all and is not even carried on the canvas projection; and the align control offers three of the four values the engine now accepts. Underneath, three canvas projection bounds *abort the entire projection* rather than degrading one element — and because the property command re-projects inside its own transaction, the 512-**byte** value cap does not merely blank the canvas, it **rejects the author's edit**. That cap is ~80 English words, or ~170 Thai/CJK characters, so the story's first acceptance criterion is unreachable until it lifts.

**Approach:** Split every projection bound that is doing two jobs — one set of bounds for identifiers, colours and expressions, a separate, derived set for document body text — and turn the body-text bounds from aborts into a per-element degradation that paints a prefix and flags itself, reusing the `fontChain` precedent already in the file. Mirror every changed bound in the TypeScript validator **in the same commit** and land a test that reads both sides. Then build the designer surface on top: a multi-line editor that preserves typed and pasted paragraph breaks, a line-spacing control (which requires projecting the field for the first time), and an align control that offers `justify` for text only. Every edit stays one opaque command; the canvas re-projects from the engine's answer.

## Boundaries & Constraints

**Always:**
- **Truncate the paint, never the value.** `CanvasTextPaint.Lines` is regenerated every projection and never written back; `component.Value` is what the properties panel edits and saves. Any degradation lives on the paint side. No code path may write a shortened value back into the document.
- **The degraded state must be distinguishable from the empty state.** Today an over-long element and an empty element both project `Lines: []`. A truncated element must paint its first N lines and carry a new projection flag beside `Overflow bool`, and the author must be *told* in text — not by colour alone, and not only by a CSS class (the existing `Overflow` flag sets a class for which no CSS rule exists; do not repeat that).
- **Every number replacing a bound is DERIVED and recorded, with the criterion written into the constant's own comment.** The const block at `page_setup.go:26-28` carries no comments today; the new ones must. No round numbers chosen for looking tidy.
- **Whichever Go bound changes, its TypeScript mirror changes in the same commit, and an executable assertion ties the two** — a test that reads both sources, not a comment asking the next person to remember.
- Degradation is a **projection** disposition, never a document validity rule: no new `diag.Diagnostic` code, no registry entry, no load-time refusal. The render path has no such cap and a 400-clause document must keep rendering.
- Every edit travels as one opaque command over the existing channel; the canvas re-projects from the engine's returned snapshot and never from a browser measurement (AD-15, AD-16, AD-17).
- Corpus byte-identity holds: all twenty recorded golden digests are **measured** unchanged and reported as measured.

**Block If:**
- Any recorded golden digest in `goldenDigestRecord` moves. The bounds and the designer surface are both byte-neutral by construction; a moved digest means something else changed.
- Preserving pasted paragraph breaks turns out to require a new runtime dependency (an HTML/RTF parser or sanitiser). `design-contract.test.ts:37-48` pins `package.json` and the lockfile, so this is a dependency decision, not an implementation choice.
- A canvas bound cannot be made to degrade without also changing what documents are *loadable* — that would invert the canvas-approximate/preview-exact asymmetry and is a design change, not an implementation detail.
- The Go↔TypeScript tie assertion cannot be expressed (e.g. the designer test runner cannot read `folio-go/page_setup.go`). Do not substitute a comment; the tie is the point of D-7.4.5.

**Never:**
- **Do not implement DW-29.** The load-time refusal of `style.align: "justify"` on a table element or its `headerStyle` is routed OUT of this story to a named **Story 7.8** — see Design Notes. This story's obligation for it is confined to (a) not *offering* `justify` for a table in the inspector, and (b) the record edits listed in Tasks.
- Do not touch `internal/layout/paginate.go`, the window model, or anything that decides where a page break falls.
- Do not close DW-26, DW-27, DW-28 or DW-31. DW-28 in particular: note it, do not fix it.
- Do not change `SupportedMajor` or `SupportedVersion`, and do not touch `versionRequiredByContent`'s ranking.
- Do not do Story 7.5's work (the content column past page one, the projected window count) or Story 7.6's (drawing multiple sheets). The canvas still draws one page.
- Do not touch `TestCorpusMeetsP6ExerciseFloors` or the P6g floor, and do not add a `beyondBaselineAcceptance` or `baselineAcceptanceFixtures` entry — this story adds **no golden fixture** (it changes no rendered byte).
- Do not raise a bound merely so a symptom stops reporting: the cliff is the defect, its position is not.
- **Never modify, move, delete or commit the repository-root `README.md`.** Stage explicit paths only; never `git add -A` or `git add .`.
- Do not claim the canvas shows where the engine will break a **bound** value. It does not — see D-7.4.4 in Design Notes.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Typed multi-line value | Author types three paragraphs separated by Enter into the content editor and blurs | One `updateComponentProperties` command carrying the value with `\n` separators; engine accepts; canvas re-projects with one `CanvasTextLine` per mandatory break | No error expected |
| Long clause, under the new bounds | A 40-page clause (≈1900 lines) pasted into a content-band text element | Command accepted; projection carries every line; no truncation flag | No error expected |
| Clause past the derived line bound | A value projecting more than `maxCanvasBodyTextLines` lines | The **first** `maxCanvasBodyTextLines` lines are painted, the truncation flag is set, the rest of the document projects normally, the author is told in text | Degrade, never abort. No diagnostic code, no load error |
| Justified clause past the cumulative fragment bound | A justified value whose word-fragments exceed `maxCanvasBodyTextFragments` across the element | Painting stops at the last whole line that fits the bound; truncation flag set; response still valid on the browser side | Degrade, never abort |
| Value past the channel ceiling | A value exceeding `maxCanvasBodyText` bytes | The projection still refuses — this one site remains a hard refusal by ruling (degradation lives on the paint side only). It is a channel-representability backstop at megabyte scale, unreachable by Epic 7's input | Recorded, not fixed; the comment must say so |
| Paste from a word processor | Clipboard carrying `text/html` + `text/plain`; the plain flavour has one line feed per paragraph | Paragraph breaks survive as mandatory breaks; bold/italic/fonts/indent are discarded silently | No error; discarding formatting is never an error |
| Paste containing CRLF | Value carrying `\r\n` between paragraphs | One mandatory break per `\r\n`, not two — the engine already folds it (`internal/text/opportunity.go:183-188`) | No error expected |
| Align control, text selected | A single text component selected | Four segments offered: left, center, right, **justify** | No error expected |
| Align control, table in the selection | A table component, or a mixed text+table selection | **Three** segments offered; `justify` is absent | The panel must not offer a value that is meaningless for that element type |
| Line spacing control | Author enters `1.5` and blurs | One command carrying `lineSpacing` as a raw JSON number in thousandths (`1500`); projection carries the committed value back | Out-of-range values are refused by the engine with the located `STYLE_LINE_SPACING_INVALID` message already shipped by 7.2; the panel shows it and keeps the author's text |
| Escape and blur in the multi-line editor | Author edits, presses Escape | Draft reverts, field blurs, nothing committed — unchanged from the single-line fields | No error expected |
| Existing single-page template | Any of the 20 corpus fixtures | Projection and rendered bytes unchanged | No error expected |

</intent-contract>

## Code Map

**Every anchor below was verified in the tree at `9f08adf`. Several anchors quoted in DW-25 and in D-7.4.2 have rotted since Story 7.3 touched `page_setup.go`; the corrections are marked ⚠ and are load-bearing.**

### The canvas projection bounds — `folio-go/page_setup.go`

- `:21-28` — the const block. `MaxCanvasMillipoints` `:25`, `maxCanvasPropertyString = 512` `:26`, `maxCanvasTextLines = 256` `:27`, `maxCanvasTextFragments = 512` `:28`. **The comment block at `:21-24` belongs to `MaxCanvasMillipoints` alone — `:26-28` carry no comments at all.** D-7.4.2 §4 requires the criterion in the constant's comment; there is nothing to amend, so the comments must be written.
- ⚠ **`maxCanvasPropertyString` has NINE enforcement sites, not the eight DW-25 lists, and every line number in that list has rotted.** Re-derived at HEAD:
  - **Body text (2 sites — DW-25 names only one):** `:581` the element **value**, in `canvasComponents` (`:560`), `return nil, err` aborting the whole component list; and **`:522`** `len(fragment.text)`, inside `addCanvasTextPaint`, OR'd with the fragment-count check. **`:522` is the site nobody named.** A value longer than 512 bytes that got past `:581` would abort here instead — so the two must move together.
  - **Identifiers (2):** `:211` font-family name; `:640` `fontFamily`.
  - **Colours (3):** `:665` `color`, `:671` `background`, `:686` `border.color`.
  - **Expressions (2):** `:590` `visibleIf`, `:596` table `bind`.
  - The seven non-body-text sites keep 512 and keep aborting. **Record the population; Epic 7 makes none of them newly reachable.**
- `:455-458` — `maxCanvasTextLines` enforcement inside `addCanvasTextPaint` (`:404`). `return fmt.Errorf("folio: canvas text element %s exceeds the line projection bound", element.ID)`. Caller chain, all abort: `CanvasWithTextPaint:278-280` → `wasm/engine.go:119-122`, `:255-258`, `:294`, each `return Snapshot{}, err`. **A breaching document cannot be opened at all.**
- `:520-524` — `maxCanvasTextFragments`, **enforced PER LINE**: `paintLine` is constructed fresh at `:520` inside the per-line loop opened at `:479`, so the counter resets every line. This is the asymmetry against the TypeScript mirror.
- **`:427-435` — THE DEGRADATION PRECEDENT TO REUSE, verbatim.** `fontChain` failure sets `component.TextPaint = &CanvasTextPaint{Lines: []CanvasTextLine{}}` and `continue`s the inner `for _, element := range band.elements` loop (opened `:419`). Its comment ("They remain loadable; there is simply no honest measured paint to display yet") covers this case verbatim. **Reuse this shape; do not invent a second disposition.** Note `:436-440` (the empty-value path) assigns the *identical* value — which is exactly why the degraded state needs its own flag.
- `:30-53` — `CanvasTextFragment{Text,X}` `:31-34`; `CanvasTextLine{Top,Baseline,Advance,Width,Fragments}` `:39-45`; `CanvasTextPaint{Overflow bool, Lines []CanvasTextLine}` `:50-53`. `CanvasComponent.TextPaint` `:102`.
- **`Overflow`'s full path — the five stops a sibling field must follow:** set at `:459` (`detectWidthOverflow`, the only write); **no Go read anywhere**; TS type `engine-protocol.ts:61`; TS admission `engine-protocol.ts:231` `hasOnly(value, ['overflow','lines'])` — **exact-key, so a Go-only field addition drops the whole response**; TS consumption `engine-protocol.ts:238` and `App.tsx:1102` (class `canvas-component-text-overflow`). ⚠ **There is no CSS rule for that class anywhere** — `Overflow` is currently invisible to the author. A truncation flag that only sets a class repeats the bug.
- `:449` — `atomicSpansFor(t.doc.UnbreakableValues, nil)`. Substitutions are **nil**: D-7.4.4 confirmed.
- `:539-544` `canvasDerived`, `:546-551` `canvasDerivedSum`, `:553-558` `canvasLineTop` — per-value `MaxCanvasMillipoints` guards, each with its own error. Untouched by this story.
- `:445` — `shapeSegments(..., breaksAreConsumed)`; `:449-455` `text.Opportunities` → `packLines`. **A mandatory break in a value already projects as a new canvas line today.** The seam is `lineBreakHandling` (`render.go:1187-1210`); of six `shapeSegments` callers four consume breaks (`page_setup.go:445`, `render.go:794`, `table_render.go:897`, `:1149`) and two draw them (`page_number.go:431`, `table_render.go:682` the column label). This story adds no caller.

### The TypeScript mirror — `folio-designer/src/engine-protocol.ts`

⚠ **There are FOUR hand-copied mirrors, not the three D-7.4.5 names.** The fourth is the one that would still blank the canvas after a correct Go-side fix:

1. `:231` — `value.lines.length > 256` (mirrors `maxCanvasTextLines`).
2. `:243` — `fragments <= 512` (mirrors `maxCanvasTextFragments`) — but **CUMULATIVE across the whole component**: `fragments` is declared at `:234` in `isTextPaint`'s body, incremented at `:242` for every fragment of every line, never reset. Go's `:522` is per line.
3. `:243` — `fragment.text.length <= 512` (mirrors `page_setup.go:522`).
4. ⚠ **`:152-154` — `optionalString` caps at 512 and is applied to `value` at `:154`**, alongside `binding`, `visibleIf`, `fontFamily`, `color`, `background`, `borderColor`, `tableBind`. **This is `maxCanvasPropertyString`'s two-jobs conflation reproduced exactly on the TypeScript side**, and it mirrors `page_setup.go:581`. Lifting the Go value cap without splitting this one leaves the projection dropping at 512 anyway, with no attributable error.
- **Blast radius of any one failure:** `isTextPaint` false → `:169` → the `components.every` at `:140` → `isCanvas` `:113` → `isSnapshot` `:247` → `parseInbound` `:269` falls to `:271` `return undefined`. **The whole engine response is discarded; the designer sees no snapshot.**
- `:61` — the `CanvasProjection` component type. `align` already admits `'justify'`; there is **no `lineSpacing`** and no truncation flag.
- `:142` — `hasOnly` allow-list for a component; exact-key. **Any new projected field must be added here or the response drops.**
- `:163` — component align validator, already admits `justify` (7.3), with the comment at `:156-162` saying "The inspector control that OFFERS the choice is Story 7.4's".
- `:29`, `:111` — the **column** align vocabulary, deliberately still the triple. Leave both.
- **No Go↔TS tie exists anywhere today.** No designer test reads a `.go` file; no Go test reads a `.ts` file; there is no codegen; the three `isTextPaint` literals are not even hoisted to named constants. **The shape to copy** is `canvas-authority-contract.test.ts` / `engine-ownership-contract.test.ts`, which both `readFileSync` over source files and assert regex properties.

### The measured numbers (re-measured at HEAD — several published figures are wrong)

- ⚠ **Today's peak cumulative fragment count for one component is 256, not the ~249 recorded in DW-25.** Measured by sweeping element width 1–600pt over an adversarial 512-byte value (256 single-character words), justified, shipped chain. The peak is the degenerate one-word-per-line case, jointly pinned by both caps: 512 bytes allows at most `⌊(512+1)/2⌋ = 256` words, and `maxCanvasTextLines = 256` caps the degenerate case at 256. Against the browser's cumulative 512 that is **2× headroom**.
- ⚠ **"Roughly 73 justified lines" is true only for a ~240pt column.** Measured fragments-per-line, justified, on realistic prose: 16.72 at 11pt/523pt (full A4 content width) — **⚠ SUPERSEDED at the implementing dispatch: 16.72 came from a 13-line sample and understates the figure, which rises with sample length; re-measured on the shipped `["Noto Sans"]` chain it is 18.05 at 523pt (and the 19.35 recorded mid-run was the `Roboto-Regular` TEST face, not the shipped one). See the Spec Change Log** → cumulative 512 crossed at **~31 lines**; 13.99 at 468pt → ~37; 11.17 at 360pt → ~46; 6.99 at 240pt → ~73. At 12pt: 15.48/523pt → ~33; 6.99/240pt → ~73. **The general law is geometry-free: cumulative fragments ≈ the word count of the value**, so 512 is crossed at ~512 words at any width. Quote the law, not the 73.
- ⚠ **The decision log's vertical metrics are mislabelled.** `epic-7-8-decision-log.md:891` and `7-2-…:701` attribute `FirstBaseline: 11759, Advance: 14982` to **12pt**; they are the **11pt** values (the fixture they came from, `line_spacing_template.go:60-75`, declares `"fontSize": 11`). Measured at HEAD on `["Noto Sans"]`: **11pt → FirstBaseline 11759, Advance 14982, LastDescent 3223; 12pt → 12828 / 16344 / 3516.**
- ⚠ **"~45 lines per A4 page at 11pt" (D-7.4.2 §4) is the 12pt figure.** A4 page height 841890 mp (`render.go:26`); content-band height `ContentHeight(g)` (`internal/layout/band.go:75-77`) = 729890 mp for the canonical 36pt margins + 20pt header + 20pt footer (the exact value already shipped in `App.test.tsx:17`). `⌊729890/14982⌋ = **48** lines/page at 11pt`; `⌊729890/16344⌋ = 44` at 12pt.

### The designer surface — `folio-designer/src/App.tsx`

- `:890-920` `ComponentProperties`. `:892` `types`, `:893` `all`, `:896` `table`, `:898` `typographic = all(type => type === 'text' || type === 'table')`.
- `:855` `contentField = { field: 'value', label: 'Text', affix: 'Text', fx: 'placeholder' }`; rendered `:912` only when `single && types.has('text')`.
- `:953-991` `PropertyDraft`. **`:990` is the whole control: a plain `<input>`** — `type="text"`, so **it cannot hold a newline at all**. Commit on `onBlur`; **`Enter` calls `preventDefault()` and commits**; `Escape` reverts and blurs. `:978` `commit()`, `:969-977` `submit()` (single-flight, reconciles through `canonicalValue` `:992`).
- `:842` — `FieldSpec` already carries an **unused `prose?: true` flag**, consumed at `:954`/`:990` → class `property-value-prose`, styled `App.css:152`. No spec sets it. It is the hook a multi-line field was anticipated through.
- `:881-885` — `alignGlyphs` (typed `Record<'left'|'center'|'right', string>`), `AlignIcon` (`:882-884`, same union), `alignSegments` (`:885`, three entries). **No `justify` glyph, no `justify` segment.** `:886` `valignSegments`.
- `:913` — the TYPOGRAPHY section, one line: font family, font size, bold/italic, colour, then `SegmentedProperty` for align and valign. **`:1052-1067` `SegmentedProperty` takes `field` and `segments` and never sees `component.type`** — the segment list is a module constant. Widening it per type is a one-line derivation inside `ComponentProperties`, which already computes `types`/`all`/`table`.
- ⚠ **The comment at `:862-866` still says the align set is left/center/right "so the control can offer exactly the accepted values and nothing else"** — stale since 7.3. Correct it.
- `:1168-1171` `TextPaint`; the whole render is `:1170`. `:1102` the overflow class. `:1174` `componentAccessibleName` (joins fragment text, `.slice(0,160)`).
- `:1223-1226` `isEditableTarget` **already includes `TEXTAREA`**, and the global key handler (`:649-673`) returns early for editable targets — so a textarea will not steal ⌘Z/arrows for free. ⌘S still fires (`:654-658`).
- `App.css:75` `.canvas-text-paint,.canvas-text-line,.canvas-text-fragment{position:absolute;display:block;white-space:pre}`; `:79-82` the paint/line/fragment rules; `:224` `textarea{min-height:120px;resize:vertical}` **already exists**; `:140-141` `.property-field:focus-within` focus ring; `:230` `input,select,textarea:focus-visible`.

### The command channel

- `folio-designer/src/component-property-command.ts:4` `PropertyField` union — **`lineSpacing` is absent**. `:8` `pointFields` (millipoint literals, emitted unquoted by `pointLiteral` `:19-22`); `:24-28` `propertyValue`; `:29` `quote()` **already escapes `\n`, `\r`, `\t`** — the wire format needs no change for multi-line.
- `:9-17` `updateComponentPropertiesCommand` — **one field per command**; `changes` always has exactly one key. Go's `applyPropertyChanges` would accept several (`component_commands.go:918-926`) but no TS caller builds one.
- `App.tsx:456-476` `applyProperties` → `engine.request('command', …)`; `engine-client.ts:52-72`; `engine.worker.ts:63`; `wasm/cmd/engine/main.go:144` → `wasm/engine.go:210-262`, which **re-projects at `:255` via `folio.CanvasWithTextPaint`** — that is the answer the canvas re-projects from.
- `folio-go/component_commands.go:894-1189` `applyPropertyChanges`. `:895-916` the `allowed` map (element type IS in scope: `:897`, `:900`, `:904`, `:909`). `value` text-only `:900-903`, set at `:1021`, placeholder-guarded `:1015-1020` — **no length cap and no newline handling on the command path itself**. `align` arm `:1048-1070` (7.3's `template.IsStyleAlign` check at `:1066-1068`). `lineSpacing` arm `:1120-1143` — **expects a raw JSON number in thousandths**, via `template.DecodeLineSpacingRaw` (`internal/template/linespacing.go:100-118`, range `[1,1000000]`).
- ⚠ **`:833` calls `Canvas(t)` inside the same transaction.** This is why the 512-byte cap is not merely a paint limit: it **fails the property command**. AC1 is literally unimplementable until it lifts.

### Story 7.5 independence — the evidence, already gathered

- **Nothing reads `CanvasTextPaint` in Go.** `grep TextPaint folio-go/*.go` (non-test) yields writes only: `page_setup.go:102`, `:433`, `:436`, `:438`, `:533`, and the type at `:50`.
- Component extent is `projectedSize(element)` (`component_commands.go:1735-1744`) — the **document's** width/height, called at `page_setup.go:570`. Band geometry is `page_setup.go:251-255` from page size and margins. Windows are `internal/layout.Paginate` (`paginate.go:548`) over `layout.ColumnItem`s built in `render.go:2025/:2079/:2119` and `page_number.go:91/:122/:142` from the render/shaping path — **never from the paint array**.
- TypeScript reads `textPaint` at exactly three places: `App.tsx:1101` (overflow class), `:1169-1170` (painting; the wrapper is `inset:0`, so it takes its box from the component, not the line count), `:1174` (accessible name). **No JS reads `paint.lines.length`.**
- **So a truncated paint shortens nothing today.** The plausible future breach is a designer-side `lines.length × advance` height derivation — the data sits right there in `line.advance`. The assertion must therefore be a **negative contract**, not just a positive test.

### Contracts any change must survive

- `folio-designer/src/canvas-authority-contract.test.ts:18-29` — the `prohibited` regex list, scanned over **all** non-test `src` sources, all `src` tests, and **all of `e2e/`**; non-vacuity at `:41-43`; red-proofs `:52-59` (`text-align: justify` at `:58`).
  - ✅ `<textarea>` is **not** banned; `white-space: pre-wrap`/`pre`/`pre-line` are safe (`:28`'s alternation does not match them).
  - ❌ `:21` bans `scrollHeight` — **auto-growing a textarea trips it**; use the existing CSS `resize: vertical`.
  - ❌ `:26` bans `getSelection`/`createRange`; `:27` `getComputedStyle`; `:28` `text-align: justify` — **the justify segment must be an icon or label, never CSS `text-align: justify`**, in production *or test or e2e* sources.
  - ⚠ **`:64-76`, the `App.tsx` escape hatch, is positional:** the regex at `:73` requires `export function placementPoint(…)` to be followed *immediately* by `\n}\nfunction pageStyle`. **Inserting any function between them breaks this test.**
- `folio-designer/src/design-contract.test.ts:50-57` — `:53` forbids any colour literal (`#rgb`, `rgb(`, `hsl(`) in `App.css`; `:54` forbids a non-token `border-radius`. `:21-23` requires a new token name to be added to `DESIGN.md` too. `:37-48` pins `package.json`/lockfile versions.
- `folio-designer/src/engine-ownership-contract.test.ts:51-55` (single Worker owner), `:64-71` (`JSON.parse/stringify` allow-list).
- **Accessibility conventions** (no dedicated file; enforced through `App.test.tsx` by role+name queries): `aria-label` **and** `title` on every icon-only control (`App.tsx:1066`, `:990`, `:1037`); `aria-pressed` on segments; mixed states in `aria-label`/`aria-description`, never colour; errors as `role="alert"` + `aria-invalid` + `aria-errormessage` (`:984`, `:990`); reasons stated in text (`App.test.tsx:989`).

### Tests that pin current behaviour and will need to change

- `App.test.tsx:780-794` (one Enter/Blur commit + located diagnostic) and `:796-813` ("routes the one CONTENT field to the value or expression command by what was typed", asserting exact command bytes) — **both assume Enter commits on the content field.**
- `App.test.tsx:941`/`:950` — the align segments; `:943` enumerates the six accessible names, `:961` the clear payload. A fourth text-only segment changes `:943`.
- `App.test.tsx:751` — "paints only pre-broken engine text lines".
- `component-property-command.test.ts` (3 cases) — exact encoded bytes.
- `engine-protocol.test.ts` — projection admission; `:146` already mentions `style.lineSpacing`.
- **Go has no test at all for the `lineSpacing` property-command path** (`component_commands_test.go` and `component_properties_test.go` contain zero `lineSpacing` hits).

### Verification surface (all confirmed at HEAD)

- `folio-go/byte_neutrality_test.go:92-460` `goldenDigestRecord` — **20 entries**; all nine digests named in the dispatch confirmed exactly. The "invalidated IN WHOLE" clause is `:225`.
- `folio-go/matrix_test.go:69-74` `matrixTargets` (4); `:1414-1788` `matrixDocuments` (21); `TestTargetRenderHash` `:1979`, gate `:1980-1990` (**`t.Log("… asserts NOTHING …"); return` when `FOLIO_MATRIX_TARGET` is unset**); `TestCrossTargetByteIdentity` `:1802`.
- `folio-go/matrix_registration_test.go:40` — **untagged**, so it runs in the ordinary suite and cross-checks `matrixDocuments` against `matrix.yml`.
- `folio-go/internal/text/corpus_test.go:169` `TestCorpusMeetsP6ExerciseFloors`; the P6g floor line is **`:185`** `{"P6g (opaque names)", stats.P6g, 20}`; measured at HEAD: **`got 7, need >=20`**, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`; the other six subtests PASS. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` `:243` (baseline P6g = 7 at `:251`) **must stay green**.
- `folio-go/fontgen_matrix_test.go:64` `TestShippedFacesReproduceFromUpstream` — never skips; without `fontTools` it fatals at `:99-108`.
- `lint/internal/rules/licencegraph_test.go:112` — the one gofmt-dirty file (DW-23). `gofmt -l folio-go` is clean.

## Tasks & Acceptance

**Execution — Part 1: the projection bounds (DW-25). Do this first; the designer surface depends on it.**

- `folio-go/page_setup.go` — **Split the const block.** Keep `maxCanvasPropertyString = 512` for identifiers, colours and expressions. Add `maxCanvasBodyText` (bytes), `maxCanvasBodyTextLines`, and `maxCanvasBodyTextFragments` (cumulative per element), each with a doc comment stating **the criterion and the arithmetic**. Derivations to record, using the corrected inputs in the Code Map: `maxCanvasBodyTextLines = 1920`, from `40 pages × ⌊729890 mp content-band height ÷ 14982 mp advance at 11pt on the shipped chain⌋ = 40 × 48`; `maxCanvasBodyTextFragments = 65536` (**corrected at the implementing dispatch from the 32768 first specified here**), from the same forty-page document justified at full A4 content width, where a line averages 18.05 word fragments re-measured on the shipped chain — `1920 × 18.05 = 34 656`, taken to the next power of two. The 32768 first written here derived from the understated 16.72 and would have bound at ~1 690 lines, short of the forty-page criterion it claimed; `maxCanvasBodyText = 1048576` bytes, a **channel-representability backstop** sized so it cannot bind before the paint bounds do (1920 lines × ~90 characters × 3 bytes for Thai/CJK ≈ 519 KB). -- Rationale: `maxCanvasPropertyString` does two jobs; splitting at the declaration is what makes an identifier bound structurally unable to govern body text. The numbers are derived so the next reader can re-derive them.
- `folio-go/page_setup.go` — **Point `:581` and `:522` at `maxCanvasBodyText`.** Both are body-text sites; a value that passed `:581` and failed `:522` would abort in the paint loop instead. Leave `:211`, `:590`, `:596`, `:640`, `:665`, `:671`, `:686` at `maxCanvasPropertyString`. -- Rationale: DW-25's enumeration named only one of the two body-text sites; re-derive by grep and record the corrected population in the closing note, per DW-31's own lesson.
- `folio-go/page_setup.go` — **Turn `:455-458` and `:520-524` into per-element degradation, reusing `:427-435`'s shape exactly.** Instead of `return err`: paint the first `maxCanvasBodyTextLines` lines, stop at the last whole line before `maxCanvasBodyTextFragments` cumulative fragments would be exceeded, set the new truncation flag, and continue the element loop. `maxCanvasTextFragments` stays as the **per-line** guard. Do not mint a diagnostic code and do not add a registry entry. -- Rationale: D-7.4.2 — the precedent is eleven lines above the site and its stated rationale covers this case verbatim; a second disposition is the thing that makes a codebase incoherent.
- `folio-go/page_setup.go` — **Add `Truncated bool \`json:"truncated"\`` to `CanvasTextPaint` beside `Overflow`**, set only at the degradation site. -- Rationale: today a 400-line element and an empty element both project `Lines: []` — the all-clear wearing the face of could-not-look.
- `folio-go/page_setup.go` — **Project `lineSpacing`** on `CanvasComponent`, set in `applyCanvasStyle` alongside the other style fields. -- Rationale: the field is absent from the projection entirely, so an inspector control would have nothing to read back; this is a precondition of AC3, not a nicety.
- `folio-designer/src/engine-protocol.ts` — **Update all four mirrors in this same commit**, and split `optionalString`: `value` takes the body-text bound, the other seven keys keep 512. Then `:231` `lines.length > 1920`; `:243` `fragments <= 65536` (see the corrected derivation above) and `fragment.text.length <= 1048576`. Add `truncated` and `lineSpacing` to the component `hasOnly` allow-list (`:142`), to `isTextPaint`'s `hasOnly` (`:231`), and to the type at `:61`, with a `typeof … === 'boolean'`/number check. **Note the unit mismatch in the two string bounds and record it rather than "fixing" it:** Go counts **bytes** (`len()`), TypeScript counts **UTF-16 code units** (`.length`), so for non-ASCII the TS side is the more permissive of the pair. This asymmetry is pre-existing at 512/512 and is safe in that direction — the Go side refuses first — but the tie assertion compares literals, not quantities, and must say so in its own comment. -- Rationale: `hasOnly` is exact-key, so a Go-only field addition silently drops the entire snapshot; `:152-154` is `maxCanvasPropertyString`'s conflation reproduced on the TS side — the fourth mirror D-7.4.5 did not name, without which the Go split changes nothing observable; and an undocumented unit mismatch inside a test that claims to tie two constants is exactly the false assurance the tie exists to prevent.
- `folio-designer/src/` (new contract test, e.g. `engine-bounds-mirror.test.ts`) — **Read `folio-go/page_setup.go` and this repo's `engine-protocol.ts`, extract the four Go constants and the four TS literals by regex, and assert equality pairwise.** Include a non-vacuity assertion that all four were actually found, and a red-proof that a deliberately mismatched literal fails. Follow `canvas-authority-contract.test.ts`'s `readFileSync` shape. -- Rationale: D-7.4.5 requires a test that reads both, not a comment; and without the non-vacuity guard a regex that stops matching passes silently, which is the failure mode the tie exists to prevent.
- `folio-go/page_setup_test.go` (or nearest) — **Assert the per-line/cumulative asymmetry explicitly**: a projection whose per-line fragment counts are all legal but whose cumulative total exceeds the browser bound must be degraded Go-side, not emitted. -- Rationale: the two sides bound different quantities; the Go side must not emit what the TS side will reject, or the canvas blanks with no attributable error.
- `folio-go/` (nearest layout/canvas test) + `folio-designer/src/canvas-authority-contract.test.ts` — **Assert Story 7.5's window count is independent of the paint.** Positive half: a document whose paint truncates produces an identical `Paginate` page count to the same document untruncated. Negative half: add a prohibited pattern banning any height or window derivation from `textPaint…lines.length`, with its own red-proof. -- Rationale: D-7.4.2 §5 — the independence holds today but is one plausible line of designer code away from breaking, and 7.6 would then draw the wrong number of sheets. A positive test alone would not catch its introduction.

**Execution — Part 2: the designer surface.**

- `folio-designer/src/App.tsx` — **Make the CONTENT field multi-line.** Render a `<textarea>` (not an `<input>`) for the field carrying the existing unused `prose` flag, keeping every other `PropertyDraft` behaviour identical: commit on blur, Escape reverts and blurs, single-flight `submit`, `canonicalValue` reconciliation, `aria-label`/`aria-description`/`aria-invalid`/`aria-errormessage`, the `fx` marker and the mixed-value placeholder. **`Enter` must insert a newline** in this field while the single-line fields keep Enter→commit. Do not auto-grow via `scrollHeight`; `App.css:224` already supplies `min-height` and `resize: vertical`. -- Rationale: `type="text"` cannot hold a line feed at all, which is the whole of AC1; and the minimal-change reading keeps every other committed behaviour, so the blast radius is one field's Enter semantics.
- `folio-designer/src/App.tsx` — **Handle paste on the content editor by taking `text/plain` only**, discarding every other clipboard flavour without error. Do not read `text/html` or `text/rtf` and do not add a parsing dependency. Normalise nothing except what the engine does not already fold — the engine folds `\r\n` itself. -- Rationale: "other formatting is discarded without error" is achieved by never reading the flavours that carry it; a sanitiser would be a new dependency, and `design-contract.test.ts:37-48` pins the lockfile.
- `folio-designer/src/App.tsx` — **Add a `justify` segment to the align control, offered only when the selection is entirely text.** Widen `alignGlyphs`/`AlignIcon`'s literal unions, derive the segment list inside `ComponentProperties` from the already-computed `types`/`all`, and offer three segments for a table or a mixed text+table selection. **Correct the stale comment at `:862-866`.** The glyph must be an SVG path; **never CSS `text-align: justify`**, in production, test or e2e sources. -- Rationale: AC3 scopes justify to "a text component is selected", and Epic 14.4's principle forbids offering a value that is meaningless for the element type; `canvas-authority-contract.test.ts:58` red-proofs the CSS ban across all three source trees.
- `folio-designer/src/App.tsx` + `component-property-command.ts` — **Add a line-spacing control** to the TYPOGRAPHY section following the `fontSize` `PropertyDraft` pattern (decimal `inputMode`, placeholder showing the inherited default, clear action). Add `lineSpacing` to `PropertyField` and emit it as a **raw unquoted JSON number in thousandths** (`1.5` → `1500`) — a third numeric encoding alongside `pointFields`' millipoints. -- Rationale: `component_commands.go:1120-1143` decodes a raw number via `DecodeLineSpacingRaw`; a quoted string is refused, and the existing `pointLiteral` scale is wrong by three orders of magnitude.
- `folio-designer/src/App.tsx` + `App.css` — **Surface the truncation flag to the author in text.** State the concrete reason at the component, distinguished by shape and words before colour, using design tokens only. Give the flag an accessible statement, not just a class. **While here, note in the spec's Delivery Log that `Overflow`'s existing class has no CSS rule at all** — do not fix it, but do not repeat it. -- Rationale: D-7.4.2 §2 requires the author be *told* the text is cut; `design-contract.test.ts:53` forbids colour literals in `App.css`; the accessibility floor requires shape-before-colour.
- `folio-designer/src/App.test.tsx` — **Update `:780-794`, `:796-813`, `:941`/`:943`/`:950`** for the new Enter semantics and the fourth segment, and add cases for: a typed multi-line value committing as one command; a paste preserving paragraph breaks and dropping `text/html`; `justify` present for a text selection and **absent** for a table and for a mixed selection; the line-spacing command's exact encoded bytes; keyboard reachability and accessible names for both new controls. -- Rationale: these are the tests that currently assert the behaviour being changed; leaving them would either fail or, worse, be weakened into vacuity.
- `folio-designer/src/engine-protocol.test.ts` — Assert a projection carrying `truncated: true` with a painted prefix is admitted; one carrying an unknown key is rejected; `lineSpacing` is admitted; and a component `value` longer than 512 but under the new bound is admitted. -- Rationale: the exact-key `hasOnly` and the split `optionalString` are the two places a silent drop would reappear.
- `folio-go/component_commands_test.go` / `component_properties_test.go` — Add the missing `lineSpacing` property-command coverage, and a test that a value carrying mandatory breaks commits and re-projects as multiple `CanvasTextLine`s. -- Rationale: the `lineSpacing` command arm has zero test coverage today, and AC1's engine half is otherwise unpinned.

**Execution — Part 3: the record.**

- `_bmad-output/implementation-artifacts/deferred-work.md` — **Close DW-25**, recording: the re-derived nine-site `maxCanvasPropertyString` enumeration with the grep that produced it; the corrected fragment peak (**256**, not 249) and the corrected crossing point (**~31 lines at full A4 width**, not 73, with the geometry-free law that cumulative fragments ≈ word count); the derived bounds and their arithmetic; and the one site that still refuses (`maxCanvasBodyText`, recorded not fixed). **Amend DW-29** to record that Story 7.4's plan gate judged it `multiple-goals` and routed it to a named **Story 7.8**, per the entry's own escalation clause — it is not a further deferral. -- Rationale: DW-25's own anchors had rotted and its enumeration was incomplete; closing against the hand-list would repeat DW-24's failure.
- `_bmad-output/implementation-artifacts/epic-7-8-decision-log.md` — Append a correction note (do not rewrite the rulings): D-7.4.2 §4's "~45 lines per A4 page at 11pt" is the 12pt figure (11pt is 48); `:891`'s `FirstBaseline 11759 / Advance 14982` are the **11pt** values, not 12pt; and D-7.4.5's "three hand-copied constants" is **four** — `engine-protocol.ts:152-154`'s `optionalString` is the fourth. -- Rationale: three later stories will reuse these numbers; a mislabelled font size propagates into every derivation built on it.
- `_bmad-output/planning-artifacts/epics.md` — Add **Story 7.8** to Epic 7, carrying DW-29: the loader must refuse `style.align: "justify"` on a table element and its `headerStyle` with a located error naming the element and the field, validating a style block's `align` against **the set its consumer accepts** rather than its JSON key path. Record the three things it inherits: the existing tests it must invert (`closedsets_test.go:287-292`, `linespacing_test.go:168-175` with its `justifyHeaderStyleDoc` const at `:311-331`, and `table_render_test.go:1338` `TestTableCellsCascadedJustifyIsDrawnAtTheStartEdge`); that a **text** element's `justify` must stay accepted, so the fix cannot be a blanket ban wearing a narrow name; and that it must decide whether to mint a **third** per-field style diagnostic code, which `internal/diag/diag.go:249-252` explicitly reserves for a deliberate decision. -- Rationale: DW-29's own escalation clause names Story 7.8 as the alternative to absorbing it, and an unwritten story is a deferral by another name.

**Acceptance Criteria:**

- Given a text component and a clause of forty pages pasted into the inspector, when the edit is committed, then the command is **accepted** (not merely the canvas surviving), the value round-trips with its mandatory breaks intact, and the canvas paints it — demonstrated end to end with a real multi-paragraph value, never by asserting that a constant changed.
- Given any body-text value that exceeds a painting bound, when it is projected, then that element paints its first N lines with the truncation flag set, **every other component in the document projects normally**, and no error is returned from the projection.
- Given a truncated element and an empty element in the same document, when their projections are compared, then they are distinguishable by a field, and the author is told in text — not by colour alone and not only by a CSS class.
- Given any Go bound this story changes, when the tie assertion runs, then it reads both `page_setup.go` and `engine-protocol.ts`, finds all four pairs (asserted non-vacuously), and fails if either side is edited alone.
- Given the same document projected with and without paint truncation, when `Paginate`'s page count is compared, then it is identical; and given the designer sources, when the authority contract scans them, then no height or window count is derived from `textPaint…lines.length`.
- Given a table component or a mixed text+table selection, when the inspector renders, then the align control offers exactly three segments; given an all-text selection, then it offers four, and the justify glyph is an SVG path with no CSS `text-align: justify` anywhere in production, unit or e2e sources.
- Given the twenty recorded golden digests, when the corpus is rendered, then all twenty are **measured** unchanged and reported as measured, not assumed.
- Given the working tree at the end of the story, when it is inspected, then the repository-root `README.md` is byte-identical to its committed state and appears in no commit.

## Spec Change Log

*Append-only.*

### 2026-08-30 — the `lineSpacing` wire unit, corrected against the engine

**Tasks Part 2 and the I/O matrix both say the command carries `lineSpacing` "as a raw JSON number in
thousandths (`1.5` → `1500`)". The parenthetical is wrong on the facts, and implementing it as
written would have shipped a control the engine refuses every time.**

`template.DecodeLineSpacingRaw` → `decodeNumberRaw` → `DecodeLineSpacing` → `decodePoints`, and
`decodePoints` performs the ×1000 itself — the very ×1000 a thousandths ratio needs, inherited rather
than restated so there is no second decimal parser to drift from (`internal/template/linespacing.go`).
So the wire carries **the author's own ratio**, and Go converts:

| wire literal | committed | note |
|---|---|---|
| `1.5` | 1500 thousandths | correct |
| `1500` | — | **refused**: 1 500 000 is outside D-7.2.3's `[1, 1000000]` |

Verified by execution against `ApplyComponentCommand` before a line of the control was written, and
pinned from both sides afterwards:
`TestLineSpacingPropertyCommandDecodesThroughTheOneLoaderValidator` (`folio-go/component_properties_test.go`)
asserts `1.5 → 1500` **and** that `1500` is refused; `component-property-command.test.ts` asserts the
exact bytes the designer emits.

**Everything else in that task holds unchanged**: it is a raw, UNQUOTED JSON number (a quoted `"1.5"`
is refused as a non-number), and it is not a millipoint field. It is carried in its own
`ratioFields` set rather than added to `pointFields`, because the mechanism is shared but the unit is
not — `pointFields` means millipoints, and putting a dimensionless ratio in it would say the wrong
thing about the number.

### 2026-08-30 — the cumulative-fragment assertion is a rule test, not a document test

Tasks Part 1 asks for "a projection whose per-line fragment counts are all legal but whose cumulative
total exceeds the browser bound". A document that reaches 32 768 cumulative fragments cannot be
projected inside an ordinary unit test: `packLines` is **superlinear in a value's break-opportunity
count**, and a justified component's cumulative fragment count is ~ its word count. Measured at HEAD
through `CanvasWithTextPaint` on one justified element: **1.2 s at 4,000 word opportunities, 9.8 s at
8,000** — so ~33,000 would cost minutes of wall clock on every run.

What shipped instead asserts the same claim in two halves, neither of which is weaker for it:

1. `canvasFragmentBudget` — the projection's own rule, extracted as a named type and called at the
   one enforcement site — is exercised directly with the **real constants**: a sequence of lines each
   carrying exactly `maxCanvasTextFragments` fragments (per-line legal, none refused for its own
   size) stops on the **cumulative** bound after exactly `65536 / 512` lines, and a fresh budget with
   its whole allowance untouched still refuses one line past the per-line guard. That is the
   asymmetry, stated on the quantity that carries it.
2. Every projection-level paint test runs `assertWithinBrowserFragmentBounds`, so no test in the file
   can emit a paint `engine-protocol.ts` would reject — per line or cumulatively.

Plus `TestCanvasBodyTextDegradesPastThePerLineFragmentGuard`, which drives the refusal through a
**real** projection at a feasible size. The `packLines` characteristic is pre-existing, is not made
reachable by anything in Epic 7, and is recorded in the decision log rather than fixed here.

### 2026-08-30 — Dispatch 2, review pass: `maxCanvasBodyTextFragments` corrected to 65536

**Triggering finding.** The constant's own doc comment quoted "measured 16.72 fragments per line at
11pt across 523pt", the figure this spec's Code Map carried, while the SAME commit's record files
(`deferred-work.md`, `epic-7-8-decision-log.md`) reported a re-measurement of 19.35 at the same
width. Three documents, two numbers, one commit — and materially, `1920 × 19.35 = 37 152` is above
32 768, so the bound did **not** cover the forty-page criterion its comment claimed for it.

**What was amended.** This spec's Tasks section prescribed `32768` and the arithmetic `1920 × 17 =
32 640`; that instruction derived from the understated 16.72 and is corrected here to `65536`. The
Code Map's measurement line and Dispatch 1's derived-bounds table are annotated as superseded rather
than rewritten, so the figure that was believed at planning time stays legible.

**The measurement that settles it.** Re-measured at the implementing dispatch through
`CanvasWithTextPaint`, counting off `CanvasTextPaint.Lines`, justified English contract prose at
11pt on the **shipped `["Noto Sans"]` chain** across 523.276pt of A4 content width: **18.05**
fragments/line at 1824 words (17.86 at 912 words; 8.10 at a 240pt column; 30.86 in a short-word
worst case). The figure rises with sample length because a short final line drags a small sample
down — which is why the briefing's 16.72, taken from a 13-line sample, was low. The 19.35 recorded
mid-run was measured on the `Roboto-Regular` TEST face, not the shipped chain, and is not the figure
the constant should carry. `1920 × 18.05 = 34 656`, to the next power of two: **65 536**, which also
clears the short-word worst case (`1920 × 30.86 = 59 251`).

**Known-bad state avoided.** Shipping a bound that binds at ~1 690 lines while its own comment
claims forty pages — the precise failure the contract's "every number is DERIVED and recorded, with
the criterion written into the constant's own comment" clause exists to prevent.

**Carried through in the same edit,** because the contract requires a bound and its mirror to move
together: `folio-go/page_setup.go`, `folio-designer/src/engine-protocol.ts`, the numeral pinned in
`engine-bounds-mirror.test.ts`, and both record files.

### 2026-08-30 — Dispatch 2: the `lineSpacing` wire unit (recorded, contract not amended)

The I/O matrix row for the line-spacing control states the command carries "a raw JSON number in
thousandths (`1500`)". That parenthetical is **factually wrong about the engine** and was not
implementable as written. Measured directly against `template.DecodeLineSpacingRaw`:

    raw 1.5  -> 1500 thousandths, no error
    raw 1500 -> refused: "lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000);
                1500000 is outside that range"

`decodePoints` performs the ×1000 itself, so the wire carries the author's **ratio** and the
*document* holds thousandths. The row's substance — one command, the committed value read back — is
satisfied; only its parenthetical about the encoding is wrong. Recorded here rather than routed to
`intent_gap`, which would have reverted a correct implementation over a factual error in a
non-normative aside. The encoding is now pinned from both sides.

## Review Triage Log

*Append-only.*

### 2026-08-30 — Review pass
- intent_gap: 0
- bad_spec: 0
- patch: 13: (high 1, medium 8, low 4)
- defer: 3: (high 0, medium 2, low 1)
- reject: 12
- addressed_findings:
  - `[high]` `[patch]` The browser e2e witness was left broken by this story's own Enter change. `browser-native-roundtrip.spec.ts:151` committed the CONTENT field with `press('Enter')`; `App.tsx:1026` now suppresses Enter-commit for the `prose` field, so the repository's ONLY cross-boundary authoring witness (browser wasm → saved bytes → native render → byte identity) would time out at its first of six calls. `test:e2e:compile` is `tsc --noEmit` and cannot see it. Changed to `field.blur()`, the idiom already used four times in that same spec. **Not executed** — browser e2e is deferred by D-000.4.
  - `[medium]` `[patch]` `maxCanvasBodyTextFragments` derived from a stale measurement; corrected 32768 → 65536. See the Spec Change Log entry above for the measurement and the amendment.
  - `[medium]` `[patch]` A FIFTH hand-copied cross-language bound shipped with none of the tie this story built for the other four: `engine-protocol.ts` hard-coded `lineSpacing < 1 || > 1000000`, mirroring `MinLineSpacingThousandths`/`MaxLineSpacingThousandths` — which `linespacing.go:44` itself calls "a STATED SANITY CEILING", i.e. a number that gets adjusted. Raising the Go ceiling would have passed every test while the browser silently dropped every snapshot. Hoisted to exported constants, consumed at the validator site, and tied by `engine-bounds-mirror.test.ts`, which now reads TWO Go sources; pair count 4 → 6, non-vacuity guard extended to assert both sources are represented. Red-proofed in three directions.
  - `[medium]` `[patch]` `pasteProse` returned BEFORE `preventDefault()` when the clipboard had no `text/plain` flavour, handing the insertion to the browser's default paste — which inserts text derived from `text/html`, contradicting both the matrix row and the handler's own comment that other flavours are never read. Ordering fixed; test added.
  - `[medium]` `[patch]` `pasteProse` never restored the caret. The textarea is controlled, so pasting into the middle of a clause moved the cursor to the end and the next keystrokes landed in the wrong paragraph. Caret restored via a ref and `useLayoutEffect`; red-proofed (29 vs the expected 20).
  - `[medium]` `[patch]` The panel's line-spacing error test asserted a message the engine cannot produce. `component_commands.go:1140` wraps the validator error as `fmt.Errorf("%s: %w", key, err)`, so the real message DOUBLES the key (`lineSpacing: lineSpacing must be between…`); the mock omitted the prefix, so the test passed against a fiction. Corrected from a real run, and the Go-side test strengthened from `err != nil` to asserting `Message`, `DataPath` and `ElementID`.
  - `[medium]` `[patch]` Nothing pinned that Go actually emits the `truncated` key, which the browser now REQUIRES (exact-key `hasOnly` plus a boolean check). Adding `,omitempty` — which every neighbouring optional field carries — would have kept both suites green while `parseInbound` discarded every real snapshot. Added a marshalling test; red-proofed by adding `,omitempty`.
  - `[medium]` `[patch]` A vacuous assertion hid that one degradation path paints nothing. Measured: the per-line-guard path yields `Truncated=true, len(Lines)==0`, so `assertWithinBrowserFragmentBounds` was certifying an EMPTY slice. The helper now requires a caller-stated expected line count and fails on mismatch; all three callers state theirs, and the zero is explained rather than accidental.
  - `[medium]` `[patch]` `TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve` — the executable form of "the seven surviving sites are RECORDED, not fixed" — probed only four of the seven and asserted merely that SOME error came back, so any unrelated parse failure kept it green. All seven sites now probed, each asserting its own refusal message.
  - `[low]` `[patch]` `property-field-prose` was emitted on the prose row with no CSS rule anywhere — the same class-with-no-rule shape this story's contract names and forbids repeating. Rule added (tokens only) plus prose `min-height`/`resize`.
  - `[low]` `[patch]` The typography-leak assertion (`component_properties_test.go:151`) was not extended to the newly projected `LineSpacing`. Added — and it was VACUOUS as written: the worked example carries only text and table elements, so a `rect` leak probe and a presence precondition were added with it.
  - `[low]` `[patch]` Two false statements in constant comments the contract requires to be correct: `maxCanvasBodyText`'s "1 MiB is the next power of two above 519 KB" (it is 524 288), and `maxCanvasTextFragments`'s "256 lines is about six pages at 11pt" (this file's own 48 lines/page makes it five). Both corrected; neither constant changed.
  - `[low]` `[patch]` Record and test tidy-ups: `epics.md`'s pillar still read "(7.1–7.4)" a sentence before Story 7.8 was added to the epic; two adjacent line-spacing tests both claimed to read the committed value back when only one did; and `TestMultiParagraphValue…` described a six-paragraph fixture as "four" while asserting only `len(Lines) < 4` — tightened to the claim the fixture actually guarantees (each of the six paragraph openings must start a painted line).

**Rejected on verification, worth recording** (both were confidently asserted by a review layer and are wrong on the facts): that `deferred-work.md`'s DW-25 closing block breaks the file's heading hierarchy — `## DW-24 IS CLOSED — Story 7.3` at line 408 establishes H2 as that file's existing convention for a closing block; and that the decision log's correction §2 says "five pages" contradicting the code's "six" — no such correction exists, though the underlying arithmetic was wrong and was fixed on its own merits.

## Design Notes

**The DW-29 judgment — why this story does not implement it.**

The plan gate judged DW-29 to make this story `multiple-goals` and routed it to a named **Story 7.8**, which is the disposition the entry's own escalation clause provides for. Three reasons, in order of weight:

1. **The reachability argument that put DW-25 here runs the other way for DW-29.** D-7.4.1 placed DW-25 in 7.4 because "7.4 is where the condition first becomes reachable through the product". 7.4 does **not** make DW-29's condition newly reachable — precisely because AC3 obliges the panel not to offer `justify` for a table. After this story the defective document is still reachable only by hand-editing a `.folio`, exactly as it is today. So 7.4 discharges the **product** half; the **format** half is a separate change.
2. **It is separably shippable.** It depends on neither the editor nor DW-25: it is a load-time validation change in `internal/template`, threading the consumer's element type into `decodeStyle` (both callers already have it: `parse_bands.go:204` has `el.Type`, `:338` is statically `ElementTable`), inverting three shipped tests and deleting a version-test fixture.
3. **It carries a policy question a builder should not settle unattended.** A load error raised with `newLoadError` is **uncoded**, and `wasm/cmd/engine/main.go:276-281` destroys every uncoded load error into "The template could not be processed" — so "a located error naming the element and the field" cannot reach a designer author without a **third** per-field style diagnostic code. `internal/diag/diag.go:249-252` explicitly reserves that: *"Before a THIRD is minted, someone must decide whether the general form is right or whether AD-14's closed registry accretes one entry per style field forever."* That is a lead call.

By contrast **DW-25 is not a second goal at all**: `component_commands.go:833` re-projects inside the property-command transaction, so the 512-byte cap rejects the author's edit outright. AC1 cannot be written without it.

**Why the value cap alone stays a refusal.** D-7.4.2 §1 is explicit that "degradation lives on the paint side only", and `component.Value` is what the panel edits and saves — truncating it would write the truncation into the document. So `maxCanvasBodyText` remains a hard refusal, sized as a channel-representability backstop at megabyte scale where Epic 7's input cannot reach it. That residue is **recorded, not fixed**, and its comment must say so, following D-7.2.3's precedent for a stated sanity ceiling.

**D-7.4.4 — a limit to state, not to fix.** `page_setup.go:449` passes **nil** substitutions to `atomicSpansFor`, so no atomic span exists on the canvas and it breaks the **raw template string** (`{{…}}` and all), not the bound data. Correct as designed. This story must not claim the canvas shows where the engine will break a *bound* value.

**DW-28 — note it, do not fix it.** Any Thai sequence stacking two marks over a base (`ครั้ง`, `ทั้งนี้`, `ตั้งแต่`) is a hard `Render` error, not a diagnostic. A multi-line editor makes an author far more likely to type such text. The canvas projection does **not** go through `internal/pdf`, so the canvas will paint it happily and only the preview will fail — which Story 6.6's honest-failure path already handles. Record that observation in the Delivery Log; change nothing.

**Two traps that will bite silently.** `folio-designer/src/canvas-authority-contract.test.ts:73` neutralises exactly one `App.tsx` seam by requiring `export function placementPoint(…)` to be followed *immediately* by `\n}\nfunction pageStyle` — inserting a function between them breaks that test with a message about pointer input. And adding a new `export` to `App.tsx` adds a **fifth** `only-export-components` lint warning to the four pre-existing ones; keep new helpers unexported.

## Verification

7.4 touches the engine's canvas projection and a channel schema on both sides, so it carries the heavy tests regardless of the per-epic cadence (D-R7.1). **Report measured pass/fail counts, never "green".**

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected: **exactly ONE** failure, `TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:169`; the P6g subtest, `got 7, need >=20`), the **mandated permanent red**. Never touch it or the P6g floor. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` (`:243`) must stay **green**. Anything else red is a defect.
- `cd folio-go && go vet -tags=matrix ./...` — expected: clean.
- `gofmt -l folio-go` — run **from the repo root**; expected: no output.
- `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` — run **once per leg** with `FOLIO_MATRIX_TARGET` set: `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` (the list is `matrix_test.go:69-74`). **Unset, this test logs "asserts NOTHING" and returns — a no-op is not a pass.** Name the legs that ran.
- `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` — expected: pass; the all-four-in-one-process local gate.
- `cd lint && go test ./...` — expected: pass.
- `cd folio-designer && npm run typecheck && npm run lint && npm test` — expected: pass, **215 tests at baseline** plus whatever this story adds. The **4** pre-existing `only-export-components` lint warnings (`preview/pdf-viewer.tsx:16`, `:17`, `App.tsx:1078`, `:1085`) are not a regression; **a fifth would be**.
- `cd folio-designer && npm run test:e2e:compile` — expected: pass. Browser e2e is deferred by D-000.4 and does not run in CI; if the editor path warrants a Playwright spec, add it and compile it, but do not claim it executed unless it did.

**Nine digests to report byte-identical** (all twenty in `goldenDigestRecord` must hold; these are the ones this run must quote): statement-1 76,744 `114df1d6…`; statement-5 127,363 `70dce051…`; statement-20 269,884 `56bfbbd9…`; statement-50 555,829 `5d090b0f…`; mandatory-break 56,681 `7cf743de…`; line-spacing 57,770 `de212115…`; justified-text 59,894 `6da3b12e…`; alignment-rounding 61,346 `986400a1…`; justified-thai 15,079 `58ca4777…`.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` (`fontgen_matrix_test.go:64`) fails under `-tags=matrix` because `fontTools` is not installed here; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, owned by Story 15.2).

**Manual checks:**
- **Re-derive the `maxCanvasPropertyString` enumeration by grep at the closing revision** and paste the output into DW-25's closing note. Close against **that**, never against the hand-list — DW-25's own list was both stale and incomplete.
- **Re-measure the cumulative fragment peak** after the value cap lifts, at full A4 content width and at a 240pt clause column, and record both against the new browser bound in DW-25's closing note.
- **Demonstrate end to end** that a real pasted multi-paragraph clause reaches the canvas — a constant edit with no end-to-end demonstration is the shape that lets this defect survive (D-7.4.1).

## Auto Run Result

### Dispatch 1 — 2026-08-30, plan only

Status: `ready-for-dev`
Blocking condition: none
Baseline: `9f08adf191918992a7970b0789393bf30b331497` on `main`, tree clean
Directive: `Halt after planning.` — spec written, **no implementation code, no commits**.

**Plan-gate judgments made in this dispatch:**

1. **DW-29 judged `multiple-goals` and routed to a named Story 7.8**, per DW-29's own escalation clause (`deferred-work.md:2209-2212`). Not a deferral. This story keeps only the product half — the inspector must not offer `justify` for a table — plus the record edits in Tasks Part 3. Reasoning in Design Notes.
2. **DW-25 kept as this story's own acceptance**, not a second goal: `component_commands.go:833` re-projects inside the property-command transaction, so the 512-byte cap rejects the author's edit outright and AC1 is unimplementable without it.

**Corrections to the ruling record, established by measurement at this baseline** (tasked in Part 3, not yet applied):

- `maxCanvasPropertyString` has **9** enforcement sites, not the 8 DW-25 lists, and every published line number has rotted. There are **two** body-text sites — `page_setup.go:581` (element value) and **`:522`** (fragment text) — where DW-25 names only one.
- **D-7.4.5 undercounts the TypeScript mirrors: there are 4, not 3.** The fourth is `engine-protocol.ts:152-154`'s `optionalString`, which caps `value` at 512 alongside seven identifier/colour keys — `maxCanvasPropertyString`'s two-jobs conflation reproduced exactly on the browser side. Without splitting it, a correct Go-side fix changes nothing observable.
- Today's cumulative fragment peak is **256**, not the ~249 recorded; the cumulative cap is crossed at **~31 justified lines at full A4 content width**, not 73 (73 holds only for a ~240pt column). The geometry-free law: cumulative fragments ≈ the value's word count, so 512 is crossed at ~512 words at any width.
- `epic-7-8-decision-log.md:891` mislabels `FirstBaseline 11759 / Advance 14982` as 12pt; they are the **11pt** values. Consequently D-7.4.2 §4's "~45 lines per A4 page at 11pt" is the 12pt figure — 11pt is **48**.

**Derived bounds specified** (each with its criterion required in the constant's comment):

| Constant | Value | Derivation |
|---|---|---|
| `maxCanvasBodyTextLines` | 1920 | 40 pages × ⌊729890 mp content-band height ÷ 14982 mp advance at 11pt⌋ = 40 × 48 |
| `maxCanvasBodyTextFragments` | ~~32768~~ **superseded → 65536** | same 40-page document justified at full A4 width; the ~17 fragments/line specified here was understated, re-measured 18.05 on the shipped chain → 1920 × 18.05 = 34 656, to the next power of two |
| `maxCanvasBodyText` | 1048576 bytes | channel-representability backstop, sized above 1920 lines × ~90 chars × 3 bytes ≈ 519 KB so it cannot bind before the paint bounds |

Verification was **not** run in this dispatch: no code changed. The `## Verification` section states what the implementing dispatch must measure.

### Dispatch 2 — 2026-08-30, implement, review and commit

Status: `done`
Blocking condition: none
Baseline: `813a414e12198be86d28f61af741b56fd93fb40e` on `main`, tree clean at start
Directive: implement from `ready-for-dev`; no halt-after-planning.

**Summary.** The two conflated canvas projection bounds were split at the declaration and the body-text pair turned from whole-projection aborts into per-element degradation, reusing the `fontChain` precedent; `lineSpacing` is projected for the first time; all four TypeScript mirrors were split and tied to Go by an executable assertion; and the designer gained a multi-line content editor, plain-text-only paste, a `justify` align segment offered for text alone, a line-spacing control, and a truncation notice stated in words.

**Files changed**

- `folio-go/page_setup.go` — const block split (`maxCanvasBodyText`, `maxCanvasBodyTextLines`, `maxCanvasBodyTextFragments`, each with its criterion and arithmetic in its own doc comment; `maxCanvasTextLines` gone, not renamed); both body-text sites repointed; the two paint bounds degrade per element via the extracted `canvasFragmentBudget`; `CanvasTextPaint.Truncated` added; `lineSpacing` projected.
- `folio-designer/src/engine-protocol.ts` — all four mirrors hoisted to exported constants including `optionalString`'s split, plus the line-spacing range; `truncated` and `lineSpacing` through the exact-key `hasOnly` gates.
- `folio-designer/src/engine-bounds-mirror.test.ts` (new) — reads `page_setup.go`, `internal/template/linespacing.go` and `engine-protocol.ts`, asserts six pairs non-vacuously, checks each constant is consumed at its validator site, and red-proofs one-sided edits in both directions.
- `folio-go/canvas_body_text_bounds_test.go` (new) — degradation, distinguishability, the per-line/cumulative asymmetry, pagination independence, the seven surviving identifier sites, the channel-ceiling boundary, the at-the-bound untruncated case, and the marshalled wire keys.
- `folio-designer/src/App.tsx`, `App.css`, `component-property-command.ts` — textarea content field, plain-only paste with caret preservation, the fourth align segment, the line-spacing control, the truncation notice.
- `folio-go/component_properties_test.go`, `folio-designer/src/App.test.tsx`, `engine-protocol.test.ts`, `component-property-command.test.ts`, `canvas-authority-contract.test.ts` — coverage for the above.
- `folio-designer/e2e/browser-native-roundtrip.spec.ts` — the witness's CONTENT commit gesture follows the field's new semantics.
- `deferred-work.md` (DW-25 closed, DW-29 amended), `epic-7-8-decision-log.md` (corrections appended), `epics.md` (Story 7.8 written in).

**Review findings breakdown:** 13 patches applied (1 high, 8 medium, 4 low); 3 deferred; 12 rejected; 0 intent_gap; 0 bad_spec. Detail in the Review Triage Log.

**Follow-up review recommendation: `true`.** Patched severities: high 1, medium 8, low 4. A high-severity patched finding sets it true on its own; the score `3 × 8 + 1 × 4 = 28` is also at or above 5.

**Verification performed (measured, at the committed tree)**

- `cd folio-go && go test -count=1 ./...` — 16 packages ok; exactly ONE failure, `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)` (got 7, need >=20), the mandated permanent red. Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` PASS.
- `go vet -tags=matrix ./...` — exit 0, no output. `gofmt -l folio-go` from the repo root — empty.
- `go test -tags=matrix -run TestTargetRenderHash` — all four legs run individually with `FOLIO_MATRIX_TARGET` set (`darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm`); each `ok`, and each grepped for "asserts NOTHING" with a count of 0, so no leg was a no-op.
- `go test -tags=matrix -run TestCrossTargetByteIdentity .` — ok (21.1s).
- `cd lint && go test ./...` — 4 packages ok.
- `cd folio-designer && npm run typecheck` clean; `npm run lint` exactly 4 warnings, all pre-existing `only-export-components`, no fifth; `npm test` **235 passed / 235, 31 files** (215 at the spec's stated baseline); `npm run test:e2e:compile` clean.
- Nine golden digests re-measured from `fixtures/<name>/expected.pdf` at the final tree, all byte-identical to the recorded values: statement-1 76,744 `114df1d6`; statement-5 127,363 `70dce051`; statement-20 269,884 `56bfbbd9`; statement-50 555,829 `5d090b0f`; mandatory-break 56,681 `7cf743de`; line-spacing 57,770 `de212115`; justified-text 59,894 `6da3b12e`; alignment-rounding 61,346 `986400a1`; justified-thai 15,079 `58ca4777`. `git status fixtures/` empty. All twenty entries in `goldenDigestRecord` green via `TestGoldenDigestAgreesAtEveryDeclaredSite`.
- Matrix Test Audit: all twelve I/O rows covered by tests that ran and passed. Two rows had no covering test at first pass — "value past the channel ceiling" and the untruncated side of "long clause under the new bounds" — and both were closed with boundary tests before review.
- Known-environmental, not regressions: `TestShippedFacesReproduceFromUpstream` (no `fontTools` locally); `lint/internal/rules/licencegraph_test.go` not gofmt-clean (DW-23, Story 15.2).

**Residual risks**

- **Browser e2e was compiled, never executed** (D-000.4). The witness's commit gesture was corrected by reasoning, not by a browser run; it is the one change in this story with no executed proof.
- The canvas breaks the **raw template string**, not bound data (`page_setup.go` passes nil substitutions to `atomicSpansFor`) — D-7.4.4. This story makes no parity claim.
- **DW-28 is made more likely, not fixed:** a multi-line editor invites Thai text stacking two marks over a base, a hard `Render` error. The canvas projection does not go through `internal/pdf`, so the canvas paints it and only the preview fails — Story 6.6's honest-failure path.
- `CanvasTextPaint.Overflow`'s CSS class still has no rule anywhere; recorded, deliberately not fixed, and the new truncation flag does not repeat the shape.

## Delivery Log

### Dispatch 2 — 2026-08-30, implementation

Baseline `813a414`. Implemented in one pass, in the spec's own order: the projection bounds first,
then the four-way TypeScript mirror and its tie assertion, then the designer surface, then the
record.

**Demonstrated end to end, which is the thing D-7.4.1 said a constant edit alone would not prove.**
A real six-paragraph clause of 585 bytes — past the 512-byte cap that used to **reject the edit
outright**, because `component_commands.go:833` re-projects inside the property command's own
transaction — is committed through `ApplyComponentCommand`, round-trips byte-for-byte, comes back on
the projection for the panel to read, and paints as six-plus `CanvasTextLine`s with its mandatory
breaks intact (`TestMultiParagraphValueCommitsAndReProjectsAsSeveralCanvasLines`). The same test
proves a CRLF pair folds to **one** break, not two. On the designer side a 1,900-paragraph value
typed into the editor commits as **one** command, and a word-processor paste keeps its paragraph
breaks while `text/html` and `text/rtf` are never read at all (`App.test.tsx`, Story 7.4 block).

**What the bounds became.** `maxCanvasPropertyString` was split at the declaration, not raised;
`maxCanvasTextLines = 256` is gone rather than renamed. Nine enforcement sites re-derived by grep at
the closing revision, two of them body text; the seven identifier, colour and expression sites keep
512 and keep aborting, and `TestCanvasIdentifierBoundsStillRefuseAtFiveHundredAndTwelve` is where
that record is executable. Full arithmetic in each constant's own comment and in DW-25's closing
note.

**Two observations recorded, deliberately not fixed.**

1. **`Overflow`'s CSS class does not exist.** `CanvasTextPaint.Overflow` sets
   `canvas-component-text-overflow` on the component and **there is no rule for that class anywhere
   in `App.css`** — the older degradation flag has always been invisible to the author. Out of this
   story's contract to fix, but not repeated: the new `Truncated` flag states its reason **in words**
   at the component and carries the same sentence into the component's accessible name, so it is
   never colour-only and never class-only.
2. **DW-28 will bite an author far sooner now.** Any Thai sequence stacking two marks over a base
   (`ครั้ง`, `ทั้งนี้`, `ตั้งแต่`) is a hard `Render` error, not a diagnostic, and a multi-line editor
   makes such text far more likely to be typed. The canvas projection does not go through
   `internal/pdf`, so **the canvas will paint it happily and only the preview will fail** — which
   Story 6.6's honest-failure path already handles. Noted; nothing changed.

**Two claims this story does NOT make.** The canvas breaks the **raw template string**, `{{…}}` and
all, because `page_setup.go` passes nil substitutions to `atomicSpansFor` (D-7.4.4) — it does not
show where the engine will break a *bound* value. And a truncated paint says nothing about
pagination: `Paginate`'s page count is asserted independent of it from both directions — positively
by `TestPaginationIsIndependentOfCanvasPaintTruncation`, and negatively by a new prohibited pattern
in `canvas-authority-contract.test.ts` that bans any height or window derivation from
`textPaint…lines.length`, with its own red-proof.

**Two spec corrections, both established by execution rather than by reading, are in the Spec Change
Log**: the `lineSpacing` wire unit (the engine takes the author's ratio and converts; sending
thousandths is refused), and why the cumulative-fragment assertion tests the projection's own budget
rule rather than a 33,000-word document.

**DW-29 was not implemented**, per the Never list. This story discharged its product half only — the
align control offers `justify` for an all-text selection and three segments for a table or a mixed
selection — and the format half is now **Story 7.8**, written into `epics.md` with its three
inheritances named, including the third-diagnostic-code decision that `internal/diag/diag.go:249-252`
reserves for a lead.

### Verification — measured

| Command | Result |
|---|---|
| `cd folio-go && go test -count=1 ./...` | **1 failure**, and it is the mandated permanent red: `TestCorpusMeetsP6ExerciseFloors/P6g` (`got 7, need >=20`). Its drift twin `TestCorpusP6StatsMatchDeclaredBaseline` **green**. Nothing else red. |
| `cd folio-go && go vet -tags=matrix ./...` | clean |
| `gofmt -l folio-go` (from repo root) | no output |
| `cd folio-go && go test -tags=matrix -run TestTargetRenderHash -v .` | **all four legs run individually with `FOLIO_MATRIX_TARGET` set** — `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` — all **ok**; grep for "asserts NOTHING" returns **0**, so no leg was a no-op |
| `cd folio-go && go test -tags=matrix -run TestCrossTargetByteIdentity .` | ok (20.2 s) |
| `cd lint && go test ./...` | ok, 4 packages |
| `cd folio-designer && npm run typecheck` | clean |
| `cd folio-designer && npm run lint` | **exactly 4** pre-existing `only-export-components` warnings (`preview/pdf-viewer.tsx:16`, `:17`, `App.tsx:1136`, `:1143`) — no fifth |
| `cd folio-designer && npm test` | **233 passed / 233**, 31 files (baseline was 215 / 30) |
| `cd folio-designer && npm run test:e2e:compile` | clean. No Playwright spec added; browser e2e was **not executed** (D-000.4) |
| `FOLIO_HEAVY=1 go test -run "Fixture\|Golden\|Statement" -v .` | all pass, **zero SKIPs** — the heavy statement legs actually ran |

**Golden digests — measured, not assumed.** All twenty entries in `goldenDigestRecord` hold
(`TestGoldenDigestAgreesAtEveryDeclaredSite`, plus every per-fixture test re-rendering its input and
comparing bytes). The nine this run was asked to quote, measured at the closing revision:

| fixture | bytes | sha256 |
|---|---|---|
| statement-1 | 76,744 | `114df1d6…` |
| statement-5 | 127,363 | `70dce051…` |
| statement-20 | 269,884 | `56bfbbd9…` |
| statement-50 | 555,829 | `5d090b0f…` |
| mandatory-break | 56,681 | `7cf743de…` |
| line-spacing | 57,770 | `de212115…` |
| justified-text | 59,894 | `6da3b12e…` |
| alignment-rounding | 61,346 | `986400a1…` |
| justified-thai | 15,079 | `58ca4777…` |

Every one byte-identical to its recorded value. No golden fixture was added; no rendered byte moved.

**Known-environmental, not regressions:** `TestShippedFacesReproduceFromUpstream` needs `fontTools`,
absent here; `lint/internal/rules/licencegraph_test.go` is not gofmt-clean (DW-23, Story 15.2).

**Repository-root `README.md` is byte-identical to its committed state** and appears in no staged
change: `git status` lists only the fourteen modified and two added files this story owns.
