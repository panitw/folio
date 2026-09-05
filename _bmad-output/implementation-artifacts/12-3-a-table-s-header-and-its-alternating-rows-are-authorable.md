---
title: "Story 12.3: A table's header and its alternating rows are authorable"
type: 'feature'
created: '2026-09-05'
status: 'done'
review_loop_iteration: 0
baseline_commit: 'f76b5bfde38fb8dafda2763aeacec536df3a48a0'
context: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs. Rewritten at close to
describe what actually shipped.*

A table in Folio could already have a taller header row, a styled header row and alternating row
colours. The engine read all three and rendered them, but nobody could set them without saving the
document out and editing it by hand. That gap is now closed. Three new commands let the table editor
write those values, the engine reports back what is set, and the editor has gained a small block of
controls beside its column matrix.

The subtle part worked as intended. A header style field left blank is not blank when the document
prints — it falls back to the table's own style, and then to a documented default. Working that out in
the browser would have been easy and wrong, so the engine now sends both the value that was set and
the value that will actually be used, and the panel only displays them.

Three things a later reader may misread. Two engine tests fail, and they failed the same way before
this story began; they belong to an unrelated corpus floor. A header border is still not settable —
deliberately left for later. And a bold or italic header remains impossible, because the engine has
nowhere to inherit those from yet; that is recorded as a consequence a later epic will meet, not an
oversight here. The story also documented, without changing, a rule the format enforced in silence.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The engine already accepts, stores and renders a table's `headerHeight`, its
`altRowBackground` (Story 4.8) and its `headerStyle` block through the two-level cascade Story 4.1
defines — but no command in the product writes any of them and no projection carries any of them, so
they are authorable only by hand-editing the file the designer just saved.

**Approach:** Add the command arms and the projection members that make the three authorable, and
render the controls in the existing table editor using its shipped commit-on-blur idiom. The engine
keeps sole ownership of the cascade: the projection carries the engine's **resolved** answer beside
the committed value, and the browser composes nothing.

## Boundaries & Constraints

**Always:**

- **Two projection members per header-style field** — the committed value (so the control can tell
  set from unset, and so *clearable back to absent* stays meaningful) and the resolved value (so the
  control shows what will actually be used). One member cannot serve both: a single resolved value
  makes *absent* unrepresentable, a single committed value forces the browser to resolve. (D-12.3.1.)
- **The resolved value is produced by the engine's existing cascade** — `table_render.go:resolveHeaderStyle`,
  called at the projection construction site. It is in `package folio`, so the projection can call it
  with nothing exported. Never re-run, re-derive, mirror or approximate that cascade in TypeScript.
- **`altRowBackground` and `headerHeight` carry ONE member each**, not two. Neither has a cascade to
  resolve through: `altRowBackground` is a flat override on odd zero-based collection indexes with no
  fallback level of its own, and `headerHeight` is required so it is never absent. Committed equals
  resolved for both. **Sixteen new members in total: 7 x 2 + 1 + 1.**
- **Clearing is to the ZERO `Presence`, never to an explicit null.** `Set:true, Null:true` serializes
  the key back as `"key": null`, which is still the key in the file, changes the bytes, burns an undo
  entry and raises the document's required format version. `op: "null"` is not offered for any field
  in this story.
- **The projection, the `isTableColumns` key list, the new wire record, and `engine-client.ts`'s
  settle path move in ONE commit** (D-7.4.5). A projection member that lands without its mirror is
  not a partial feature; it is a dead worker.
- **Commit-on-blur, no local uncommitted buffer.** AD-15 owns the document; the shipped
  `TableEditor.tsx` already commits every cell on blur and has no Apply and no Cancel. Follow it.
- Every new control is keyboard-reachable with visible focus. The matrix's roving-tabindex
  navigation and its `aria-colcount` of 11 are unchanged.
- Refusal copy stays terse and technical: state the fact, name the location, offer no comfort.

**Ask First:**

- Any need to change `TableExt`'s field set — reds `TestTableExtFieldSetIsPermanent` in both
  directions and needs an owner ruling.
- Any need to make a required format key optional, or to raise the document's format version.
- Any proposal to change `render.go:defaultFontSizePt`. See Design Notes; it is filed, not fixed.
- Any header-style field beyond the seven named below.
- Any change to the `spelled` lookup map's maximum key in `App.test.tsx`'s self-count guard beyond
  extending it by the entries this story needs.

**Never:**

- **Never compute a header-style fallback in TypeScript.** Handing the browser the ingredients — the
  committed field plus the table's own style — and letting it choose IS implementing the cascade in
  the browser, and violates AC2, AC3, AD-15 and AD-17.
- **Never add or remove a format key, and never raise the document version.** A document whose new
  values are not edited must serialize to identical bytes; no corpus digest may move.
- **Never author `headerStyle.padding` or `style.padding` from the panel** (D-12.4.1).
- **Never offer a clear affordance for `headerHeight`.** It is required.
- Never build the mockup's *"Show header row"* toggle or its three-way BORDERS preset — no format
  field backs either, and inventing a control the format cannot carry is the defect Epic 14's own
  AC4 exists to prevent.
- Never add a control as a matrix column; that is Story 14.7's surface, and `cellCount = 11` is
  load-bearing for `focusCell`, `moveFocus`, Home/End and post-reprojection refocus.
- **Never, against this repository: `git commit`, `git add`, `git stash`, `git checkout`, `git reset`,
  `git revert`, `git restore`, `git push`, `git merge`, `git rebase`, `git cherry-pick`, or create a
  branch.** The human makes every commit. Read-only git is fine. (Carve-out, already ratified:
  `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`, removed in a
  `finally`, is permitted — it is how the shipped scanner suite builds hermetic fixtures.)

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Set a header-style field | table `e1`, `fontSize` set to `14` | `headerStyle.fontSize` written; committed member reads 14; resolved member reads 14 | N/A |
| Clear a header-style field | `headerStyle.fontSize` present, cleared | key removed from `headerStyle` (zero `Presence`, not null); committed member absent; **resolved member reads the table's own `style.fontSize`, else the documented default** | N/A |
| Clear the last header-style field | `headerStyle` holds one field, cleared | `headerStyle` itself becomes absent rather than an empty object; bytes carry no `headerStyle` key | N/A |
| Set the alternating row background | `altRowBackground` absent, set to `#DDEEFF` | key written; odd zero-based collection indexes paint it, exactly as Story 4.8 already renders | N/A |
| Clear the alternating row background | `altRowBackground` present, cleared | key removed; rows revert to `style.background` | N/A |
| Set the header height | table `e1`, height set to 18pt | `headerHeight` written as millipoints | N/A |
| Attempt to clear the header height | clear op on `headerHeight` | refused, document unchanged | located refusal naming the field; no clear affordance is rendered, so this is reachable only from a hand-built command |
| Header height that overflows its band | height that pushes the table past its band | refused, document unchanged | located refusal; containment re-checked exactly as `updateTableColumn`'s width case does, because a table's projected height IS its `headerHeight` |
| Malformed colour | `altRowBackground` set to `not-a-colour` | refused by the engine, document unchanged | the engine's own located message is shown; the panel invents no second validation |
| Unknown header-style field name | `field` outside the seven | refused | located refusal; the arm's own validation is the only gate |
| `null` op on any 12.3 field | `op: "null"` | refused | located refusal naming the field |
| No edit at all | table opened, nothing changed | **bytes identical**; no revision, no dirty flag, no undo entry | N/A |
| Re-set a value to what it already is | `fontSize` already 14, set to 14 | canonical-bytes short-circuit: unchanged snapshot, no undo entry, **and no error** — a silent success | N/A |

</frozen-after-approval>

## Code Map

Measured at `f76b5bf` (the dispatch baseline; tree clean apart from step-01's own recompiled
`epic-12-context.md`). Counts marked CLOSED are fully enumerated. **Cited by symbol, not line number**
(D-12.C.4: this run has already found citations that were correct when written and false on arrival,
because the block citing them moved the target).


**HEAD has since moved to `910741c` and every anchor below still holds.** The three intervening
commits touch only `_bmad-output/` — `epic-11-14-decision-log.md` (+209) and `epics.md` (+17), the
latter adding Story 14.8's owner-decision block. **`git diff --stat f76b5bf..HEAD` shows no source
file of any kind.** Story 12.3's own text in `epics.md` is byte-identical across the move (diffed).
So the Code Map was measured at `f76b5bf`, `baseline_commit` stays `f76b5bf`, and neither is stale.

### Go — the model and what already exists

- `folio-go/internal/template/model.go:TableExt` -- the three properties, and their types decide the
  whole story: `HeaderHeight geom.Length` (**plain value, required**), `AltRowBackground Presence[string]`,
  `HeaderStyle Presence[Style]`. `model.go:Style` -- **11 fields, CLOSED** (Align, Background, Color,
  Bold, Italic, Border, FontFamily, FontSize, LineSpacing, Padding, Valign; `Extra` excluded as the
  documented passthrough carrier). `headerStyle` reuses `Style` verbatim — it is not a separate type.
- `folio-go/internal/template/parse_bands.go:decodeTableExt` -- `headerHeight` is pre-seeded into
  `consumed` beside `bind` and `columns`; a missing key is `newLoadError("headerHeight", id, "",
  "missing required field for a table")`. **Exactly three table fields carry that error, CLOSED.**
  Also: `altRowBackground: null` is a load error, so `AltRowBackground`'s `Null` state is unreachable
  from the loader.
- `folio-go/internal/template/serialize.go:writeElement` -- `headerHeight` is emitted
  **unconditionally**; `altRowBackground` is guarded on `.Set` with **no `.Null` branch** (so a
  `Set+Null` value would serialize as `""`, which the loader then rejects as a non-hex colour — one
  more reason clearing must be to the zero `Presence`); `headerStyle` is guarded on `.Set` with a
  `.Null` branch. `serialize.go:writeObject` sorts keys with `sort.Slice`, so canonical order is
  computed, not hand-maintained — **no ordering work is needed and no ordering test can break.**
- `folio-go/table_render.go:resolveHeaderStyle` -> `resolvedHeaderStyle` -- **the cascade, and there is
  exactly one of it.** Non-test callers, CLOSED over the module: `resolveHeaderStyle` 1,
  `resolveBodyStyle` 1. Both are `package folio`, as is the projection, so **nothing needs exporting.**
  **9 arms** (FontFamily, FontSize, LineSpacing, Border, Background, Color, Padding, Valign, Align);
  Align resolves into `alignFallback`, which a column's own `align` then beats per column.
  **No Bold arm and no Italic arm** — see Design Notes.
- `folio-go/table_render.go:collectBandTableRuns` -- the alternating-row decision, ~6 inline lines:
  `hasAltBackground := tbl.AltRowBackground.Set && !tbl.AltRowBackground.Null`, applied on
  `rowIdx%2 == 1`. **Zero fallback levels of its own.** Do not touch; this story adds no rendering rule.
- `folio-go/component_commands.go:projectedSize` -- a table's projected size is
  (sum of column widths, `Table.Value.HeaderHeight`). **`HeaderHeight` IS the table's height** — which
  is why `width`/`height` are withheld from tables in `applyPropertyChanges`, and why a header-height
  arm inherits `updateTableColumn`'s containment obligation via `containComponent`.

### Go — the door

- `folio-go/component_commands.go:ApplyComponentCommand` -- **27 case arms, CLOSED** (re-measured
  directly; a `case "` count over the whole file returns 30 at one indent and 50 at any indent,
  because `updateTableColumn` has its own `field` switch — do not count that way). The last three in
  source order are `setBandHeight` (12.1), `setDocumentLocale` and `setDocumentUTCOffset` (12.2). New
  arms append after them. Pre-switch gates in order: nil template; `refuseDuplicateCommandKeys`;
  decode; surplus-token check; `version == 1`; `kind`. `default` returns a **plain** error → unlocated
  `ENGINE_REJECTED` at the host with no elementId and no dataPath.
- **The seven existing table arms, all routed through `applyTableColumnCommand`** (serialize → reparse
  → apply → serialize → reparse → project → install): `addTableColumn`, `removeTableColumn`,
  `moveTableColumn`, `updateTableColumn`, `configureTableBinding`, `updateTableColumnBinding`,
  `updateTableColumnFooter`. **Every one addresses the table by element `id` via `findComponent`** and
  repeats the same two-line gate — not found → `componentFailure(id, "table.id", "table was not found")`;
  not a table → `componentFailure(id, "table.id", "component is not a table")`. **Copy this gate.**
  Their DataPaths are literal constants; no path builder is needed because nothing wire-supplied is
  interpolated into a path. (`bandHeightPath` and `fontChainPath` exist only because those two DO
  interpolate, and both wrap `truncateAtRuneBoundary` against `maxComponentDataPathBytes`.)
- `folio-go/component_commands.go:applyPropertyChanges` -- **`propertyOrder`, 23 keys, CLOSED, and none
  of them reaches `element.Table`** (measured over the function body; positive control `element.` = 16
  hits in the same range). Its `op` grammar is the precedent to copy: `set` → `Presence{Set:true,Value}`,
  `clear` → the **zero** `Presence`, `null` → `Set+Null`; `clear`/`null` must carry no value; only
  `visibleIf`, `color` and `background` accept `null` at all. `duplicateComponent` carries the
  canonical statement of the clear idiom in its doc comment. `setComponentAsset`'s doc comment is a
  pre-written ruling for this story's structural question: anything the `{op,value}` grammar cannot
  express, or where clear must stay inexpressible, becomes **its own top-level command kind, not a key
  threaded through `applyPropertyChanges`.**
- `folio-go/component_commands.go:renameFontChain` -- **the ONE non-test writer of
  `HeaderStyle.FontFamily`**, CLOSED. `fontChainReferences` (used by `deleteFontChain`) is a **read**,
  not a write. Verified at the ruling's own tree with `git show 71627a5` — identical, so D-12.3.0's
  "two sites" was an over-count when written, **not drift**. Other non-test writers of the three
  fields, CLOSED: `parse_bands.go` (loader) for all three, and
  `component_commands.go:createComponentInBand` seeding `HeaderHeight: 12000` on a new table.
  **No command mutates any of the three today.**
- `folio-go/wasm/engine.go:Engine.Apply` -- the canonical-bytes short-circuit runs **before**
  `pushUndo`/`install`, so a no-op edit is free and silent. `wasm/engine.go`'s dispatch is a binary
  if/else on the literal `"pageSetup"`; **every other kind reaches `ApplyComponentCommand`, so a new
  arm needs no edit here.**
- `folio-go/wasm/cmd/engine/main.go:engineFailure` -- `errors.As(*ComponentCommandError)` →
  `COMPONENT_INVALID`, with Message/ElementID/DataPath bounded at **512 / 128 / 256 BYTES**.
- **No Go census of command kinds exists** — re-verified at this tree: `component_commands.go` is the
  only non-test file carrying all four of `"createComponent"`, `"setBandHeight"`, `"setDocumentLocale"`,
  `"addFontChain"`. A 28th, 29th and 30th arm join no list.

### Go — the projection surface (and why it is NOT the one D-12.3.1 named)

- `folio-go/table_columns_projection.go:TableColumnsProjection` -- **`{TableID, Collection, Alias,
  Columns}`; this is the surface.** It is per-table, requested by element id when the editor opens,
  and it **already carries table-level non-column data** (`Collection`, `Alias`), so a table-level
  member is not a new kind of thing here. `TableColumnProjection` is 10 keys.
- `folio-go/page_setup.go:CanvasProjection` -- **23 keys, document-level, none `omitempty`.** The
  wire test asserts a **zero** projection marshals the same key set as a projected one, which
  **forbids `omitempty` at its top level** and therefore cannot represent absence. This — plus the
  fact that these three properties are per-table while this projection is per-document — is why
  D-12.3.1's five-site checklist does not transfer. See Design Notes.
- `folio-go/canvas_projection_wire_test.go` -- pins **exactly four** key lists, each twice (Go side by
  `json.Marshal` byte-derived key set; TS side by anchored regexp reading `engine-protocol.ts` off
  disk): `canvasProjectionWireKeys` (23), `canvasFontChainWireKeys` (2), `canvasFontChainEntryWireKeys`
  (4), `canvasTextFragmentWireKeys` (4). **`isTableColumns` and `TableColumnsProjection` are pinned by
  nothing** — measured CLOSED: `isTableColumns` in any `*.go` = **0** (positive control `isCanvas` in
  `*.go` = 12); `TableColumnsProjection` in any `*_test.go` = **0** (positive control
  `CanvasProjection` in `*_test.go` = 38). The nested-record mechanism to copy is
  `marshalledObjectKeys` + a `…FromProjectionBytes` digger + an anchored regexp. **Do not reuse the
  parameter names `chain` or `fragment`** in a new guard — `canvasGuardChainKeyList` and
  `canvasGuardFragmentKeyList` are anchored on those identifiers alone and would silently read the
  wrong list.

### Designer — protocol, client, worker

- `folio-designer/src/engine-protocol.ts:isTableColumns` -- uses **`hasExactKeys`**, on the envelope,
  on `table`, and on every `column`. `hasExactKeys` = same length AND every listed key present, so it
  rejects in **both** directions; `hasOnly` (used by `isCanvas`) is a subset check that rejects an
  unknown key but **accepts a missing one**. The "typed-key validator" is the block of clauses after
  the key-list call inside the same guard.
  **⚠ The wire test file's own header comment calls `hasOnly` "an EXACT-KEY check" and says the symptom
  is "the canvas blanks". Both are wrong. Do not inherit that wording.**
- **What a guard failure actually does:** guard false → `parseInbound` returns `undefined` →
  `engine-client.ts:#fail('PROTOCOL_INVALID')` → state `failed`, handlers detached, **`worker.terminate()`**,
  every pending request rejected, **no re-spawn exists**. Not a blank canvas — a dead session. And for
  a **first** table-editor open it is completely silent: `openTableEditor`'s catch sets
  `tableEditorError`, which renders only inside `<TableEditor>`, which does not render because the
  editor never opened. **Assert the guard's return value, not a visual symptom.**
- 🔴 `folio-designer/src/engine-client.ts:#settle` -- **hand-enumerates the table object's members**:
  `{ tableId, collection, alias, columns: [...map(c => ({...c}))] }`. Columns ride a spread and
  survive; **a new table-LEVEL member passes the guard and is then silently dropped before `App.tsx`
  sees it.** No test covers this. All sixteen of this story's members ride that path. **This is the
  most dangerous site in the story and it must be fixed here, with its own test.**
- `folio-designer/src/engine.worker.ts:WasmResponse` -- re-declares the table-columns shape
  structurally inline. Missed by every existing record; it must move too.

### Designer — the table editor

- `folio-designer/src/TableEditor.tsx` -- 81 lines, one exported component. DOM order:
  `table-editor-heading` (section label, `Configure columns`, honest note, `Close Table Editor`) →
  `table-editor-config` (`Root collection`, `Row alias`, both **commit on blur** behind a
  value-changed guard) → either `table-editor-empty` or `table-matrix`
  (`role="grid"`, `aria-colcount={11}`) → `{error && <p role="alert">}`.
  **There is no draft and no Apply**: every control is uncontrolled (`defaultValue`), text commits on
  blur, selects commit on change, and `App.tsx:commitTableColumn` sends the command, installs the
  snapshot and **re-requests `table-columns`**.
- `folio-designer/src/TableEditor.tsx:trapDialog` -- builds its focus list in **document order**
  filtered to `tabIndex >= 0`, which excludes every non-active matrix cell. Today that list is
  `[Close Table Editor, Root collection, Row alias, <the one active matrix cell>]`, so the active cell
  is **last** — which is the only reason `App.test.tsx`'s "Tab from the header cell → Close Table
  Editor" assertion passes (it is exercising the **wrap** branch, though it reads like a forward tab).
  **Placing new controls after the matrix moves the active cell off the end of that list and changes
  the focus order. That is intended here — see Tasks.**
- `folio-designer/src/TableEditor.tsx:matrixCell`/`moveFocus`/`focusCell` -- roving tabindex, arrow /
  Home / End handling that skips disabled cells, and a `useLayoutEffect` on `[projection]` that
  restores the logical cell after every accepted **and** rejected commit. Untouched by this story.
- `folio-designer/src/table-column-command.ts` -- seven one-line factories, all
  `commandBytes(kind, [[key, encoded]])` from `./command-json`. **The template to copy.**
  `table-column-command.test.ts` pins **one** byte-exact wire string (`addTableColumn`) and the rest by
  decoded object equality — weaker than 12.1's band-height precedent, which pinned full decoded
  strings in order. **Pin the new factories byte-exact.**
- `folio-designer/src/App.tsx:PropertyDraft` -- the colour affordance to reuse *by pattern, not by
  import*: a hex text `<input className="property-value">` plus a native
  `<input type="color" className="property-swatch…" aria-label={`Pick ${label}`}>`, with
  `App.tsx:swatchColor` (the only local hex predicate — `/^#[0-9a-fA-F]{6}$/`, else `#000000`) and the
  `property-swatch-unset` dashed/15%-opacity class at `App.css` for the absent state.
  **`swatchColor('')` returns `#000000`, so an unset colour renders as BLACK unless the unset class is
  applied** — the same defect class as 12.2's missing `Not set` placeholder, wearing a picker.
  `PropertyDraft` itself is **not** reusable: it reads `committedValue` off the canvas projection and
  takes a different commit path. Reuse `swatchColor`, the regex, the class names and the CSS; expect
  to re-implement the two-control row.
- **Validation is not local.** The hex regex only decides the swatch chip's appearance and the colour
  the picker opens on; Go's `parseHexColor` is the gate and the panel shows its located diagnostic.

### Guards, floors and traps — re-measured at this tree

- `folio-designer/src/command-json-soleness.test.ts` -- AST-scans every production `.ts`/`.tsx` under
  `src/` (TableEditor.tsx included): only `command-json.ts` may build command JSON. **Its 8-name
  factory list is SPELLED TWICE in the file and both copies must move together** — the file says so
  itself; miss one and the `\bString\(` check never looks at the new factory.
- `folio-designer/src/canvas-authority-contract.test.ts` -- comments ARE stripped (quote-aware
  `withoutComments`, applied at all three scan sites), so a comment cannot trip the prohibited-token
  list. **No DOM measurement of any kind.** Population floors, measured now: production **61** (floor
  58), tests **57** (floor 51), e2e **15** (floor 15, **ON IT**), production `.tsx` **8** (floor 8,
  **ON IT**), `.css` **3** (floor 3, **ON IT**). App.css `@media` set is an exact `toEqual` of one
  entry — **a responsive breakpoint for a wide section reds it.** All the file-count floors are
  minimums: they red on shrink, not growth.
- `folio-designer/src/file/file-access-contract.test.ts` -- scans raw text **including comments** for
  `cloud`, `sync`, `account`, `collaborator`, `recent files`, case-insensitive. **A comment saying
  "keep the draft in sync with the projection" reds it.** It has no population floor at all.
- `folio-designer/src/canvas-font-stack.test.ts` -- an exact, ordered **6-entry** `font-family`
  position list, zero headroom. A *correct* new `font-family` declaration anywhere in `src` reds it.
- `folio-designer/src/design-contract.test.ts` -- reads App.css and never TableEditor.tsx: zero
  hex/`rgb(`/`hsl(` in App.css, `border-radius` must be `var(--radius…)`, exactly one
  `var(--type-display)` and one `var(--type-numeric-lg)`. **A static colour literal for a swatch
  cannot live in CSS.**
- `folio-designer/src/engine-ownership-contract.test.ts` -- any TS type or object literal carrying
  **≥2** of `{version, page, bands, elements, assets}` is a schema mirror and fails.
- 🔴 **`App.test.tsx`'s self-count guard (DW-199).** The spelled header word is `TWENTY-THREE` against
  **23** measured `it(`s from the Story 17.1 describe **to EOF**, and **`TWENTY-THREE` is the largest
  key in the `spelled` lookup map** — so bumping the number is not enough, the map must be extended
  too. Appending a describe below that point reds an unrelated test. Either append **above** the
  Story 17.1 describe, or extend the map. DW-199 names "Epic 14's first designer story" as the
  trigger; **12.3 dispatches first, so that trigger is stale.**
- `folio-go/render_arch_test.go:TestFolioMethodNamesAreInjective` -- walks non-test `*.go` **directly
  under `folio-go/`** and keys methods by NAME alone. `Error` and `Unwrap` each have exactly one
  declarer (`*RenderError`), so **a second error type in a root file collides immediately** and
  `Error` cannot be renamed. If a new error type is needed, declare it in `internal/*`.
- `folio-go/internal/bandcomposition_arch_test.go:bandHeightOperandNames` -- a hardcoded 7-name map
  that **contains `headerHeight`**, meaning the *band's*. A receiverless helper in `package folio`
  naming a local `headerHeight` for the *table's* trips `TestNoBandOriginArithmeticInPackageFolio` as
  a **false positive by name collision**. The walk skips methods, so handlers on the transaction
  receiver are safe.
- `folio-go/table_behaviour_suite_test.go:TestTableBehaviourSuiteIsNotSupersededByTheGolden` -- a
  7-name literal set compared bidirectionally against `filepath.Glob("table_*_test.go")`. **A new root
  test file named `table_*_test.go` reds on the commit that adds it** unless the list moves with it.
- `folio-go/internal/template/table_geometry_test.go:TestTableExtFieldSetIsPermanent` -- reflects over
  `TableExt` against a literal field set. **This story adds no `TableExt` field, so it stays quiet.**
- `folio-go/internal/template/drift_test.go` -- `extractGoKeys` AST-parses `serialize.go` and only
  recognises the **positional** `kv{"key", …}` spelling. This story changes no serializer key.

### Read-only / do not touch

- `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — the mandated permanent red.
  **Exactly two failures at baseline; a third is a hard stop.**
- `folio-designer/e2e/browser-native-roundtrip.spec.ts` — hung, pre-existing, out of scope.
- `e2e/` is **not** typechecked by `npx tsc -b` and Playwright runs in no workflow. **Put every real
  assertion in vitest, not e2e.** `e2e/table-editor.spec.ts` exists and pins an arrow walk plus
  `aria-colcount`; it may be updated but it is not evidence.
- The matrix suite and Playwright are **not in scope** — end-of-epic cadence, owed at the Epic 12
  boundary.

## Tasks & Acceptance

**Execution:**

- [x] `folio-go/component_commands.go` -- add three arms after `setDocumentUTCOffset`:
      `setTableHeaderHeight`, `setTableAltRowBackground`, `updateTableHeaderStyle`. Each addresses the
      table by element `id` and repeats the two-line `findComponent` + `is not a table` gate the seven
      existing table arms share. `updateTableHeaderStyle` accepts a `field` from the seven named in
      Design Notes and an op of `set` or `clear`; `setTableAltRowBackground` accepts `set` or `clear`;
      `setTableHeaderHeight` accepts neither. **`null` is refused by all three.** Clear writes the
      **zero `Presence`**; when the last header-style field is cleared, `headerStyle` itself becomes
      absent rather than an empty object. -- these are the missing writers; the `{op,value}` grammar
      cannot express a required field, so per `setComponentAsset`'s ruling they are their own kinds.
- [x] `folio-go/component_commands.go` -- `setTableHeaderHeight` re-checks containment via
      `containComponent`, exactly as `updateTableColumn`'s width case does. -- a table's projected
      height IS its `headerHeight`; without this an author can grow a table out of its band.
- [x] `folio-go/table_columns_projection.go` -- widen `TableColumnsProjection` by **sixteen** members:
      committed + resolved for each of the seven header-style fields, one for `altRowBackground`, one
      for `headerHeight`. Resolved values come from calling `table_render.go:resolveHeaderStyle` at
      this site. -- the panel must show what the document will use, sourced from the engine.
- [x] `folio-go/canvas_projection_wire_test.go` -- add the **fifth** record: a sorted wire-key set for
      the table projection, pinned on the Go side from marshalled bytes and on the TS side by a new
      anchored regexp over `isTableColumns`'s `hasExactKeys` list. Use a new guard-list identifier;
      do **not** reuse the `chain` or `fragment` parameter names. -- this surface is pinned by nothing
      today, and 12.3 is the first story to put load-bearing data on it, so the pin is part of it.
- [x] `folio-designer/src/engine-protocol.ts` -- widen the `TableColumns` type and `isTableColumns`'s
      `hasExactKeys` key list, and add typed clauses for the new members. -- `hasExactKeys` rejects in
      both directions, so this must land in the same commit as the Go widening.
- [x] `folio-designer/src/engine-client.ts` -- **fix `#settle`'s hand-enumeration** so table-level
      members are carried through rather than dropped. -- a guarded member is currently discarded
      silently before `App.tsx` sees it; silence is worse than a loud protocol failure.
- [x] `folio-designer/src/engine.worker.ts` -- update `WasmResponse`'s inline structural
      re-declaration of the table-columns shape. -- missed by every existing record.
- [x] `folio-designer/src/table-style-command.ts` (new) -- three factories over
      `commandBytes(kind, fields)`, importing from `./command-json` and building no JSON themselves.
      -- the soleness guard permits exactly one JSON authority.
- [x] `folio-designer/src/command-json-soleness.test.ts` -- add the new module to the factory list in
      **BOTH** places it is spelled. -- the file states the requirement itself; miss one and the
      `\bString\(` check never inspects the new factory.
- [x] `folio-designer/src/TableEditor.tsx` -- add a HEADER section **after the matrix, where the
      design places it**, carrying: header height (a required number, **no clear affordance**), the
      alternating row background (hex text + swatch, clearable), and the seven header-style fields
      (each clearable). Follow the editor's commit-on-blur idiom — no draft, no Apply, no Cancel.
      Every control keyboard-reachable with visible focus. -- placement follows the design; the focus
      order changes deliberately, see the next task.
- [x] `folio-designer/src/TableEditor.tsx` + `folio-designer/src/App.test.tsx` -- **amend the Tab-wrap
      assertion to the new focus order and state the new order in a comment as a decision.** -- the
      old assertion passed only because the active matrix cell happened to be last in document order.
      Choosing a layout to keep a test green would let a test dictate the product; a guard that
      quietly becomes a layout requirement has stopped being a guard, so the break is made visible
      and decided rather than routed around.
- [x] `folio-designer/src/TableEditor.tsx` -- an unset colour must render with the
      `property-swatch-unset` treatment, not as black. -- `swatchColor('')` returns `#000000`; this is
      12.2's `Not set` defect wearing a picker, and it is invisible to any command assertion.
- [x] `_bmad-output/specs/spec-folio/folio-format.md` -- mark `headerHeight` **Required** in the table
      field table. -- it is required by silence today: no optional marker, no default, enforced by
      code and documented by omission. This is the document describing what already ships, not a
      format change, and it is the same move 12.2 made on the `utcOffset` row.
- [x] `folio-go/table_header_style_test.go` (new) -- unit-test every row of the I/O matrix, and add
      the writer assertion scoped to *no command **authors** `headerStyle`*, **naming
      `renameFontChain` singular as the known writer**. **If this file is named `table_*_test.go` it
      must be added to `table_behaviour_suite_test.go`'s 7-name literal set in the same commit**, or
      `TestTableBehaviourSuiteIsNotSupersededByTheGolden` reds bidirectionally. -- phrased as
      "nothing writes `headerStyle`" the assertion is false at this tree; phrased as "two writers" it
      is false in the other direction.
- [x] `folio-designer/src/App.test.tsx` -- append new tests **above** the Story 17.1 describe, or
      extend the `spelled` map past `TWENTY-THREE`. -- DW-199; the map's largest key is the ceiling,
      so bumping the count alone does not work.
- [x] `_bmad-output/implementation-artifacts/deferred-work.md` -- append four entries: (1)
      `headerStyle.border`, **trigger: Story 14.8's BORDERS section, or the first request for a header
      border** — registered, not dropped; (2) **`resolveHeaderStyle` has no Bold and no Italic arm**,
      trigger keyed on Epic 11's resolution story, so 11.2 meets it as a named consequence; (3) the
      `fontSize` default disagreement (see Design Notes); (4) a re-key of DW-199's trigger onto
      **purpose — "the first story that adds a designer UI section"** — rather than an epic number,
      cross-referencing DW-199. -- a deferral without a trigger is a deletion with paperwork.

**Acceptance Criteria:**

- Given a table with no header styling, when the table editor is opened, then it offers header height,
  the alternating row background and the seven header-style fields, each showing the engine's
  committed value, and each clearable back to absent **where the format permits absence**.
  `headerHeight` is required — `parse_bands.go` hard-errors on its absence and the format's table
  table gives it no optional marker and no default — so it is offered with no clear affordance.
  (AC1's original "each … clearable" was an over-generalisation in the criterion, not a property of
  the product; `x`, `y`, `width`, `height` and `value` are already non-clearable.)
- Given an alternating row background, when the document renders, then it paints exactly as Story 4.8
  already renders it: this story adds no rendering rule and no second implementation of the cascade.
  **This is a prohibition on what 12.3 may build, not a satisfied feature** — it is discharged by the
  absence of any new resolution logic and by the browser holding none, never by a preservation test
  over the engine.
- Given a header style field left absent, when the document renders, then it falls back to the table's
  own style and then to that field's documented default — the cascade Story 4.1 already defines — and
  **the editor shows that resolved value because the engine sent it**, not because the browser worked
  it out. Likewise a prohibition, and it binds on the first criterion.
- Given a table whose header and row styling are not edited, when it is serialized, then the bytes are
  unchanged, no corpus digest moves, and the document's format version is not raised.
- Given a projected member missing from `isTableColumns`'s key list, **or** a guard key with no
  projected member, when a table-columns response is parsed, then `parseInbound` returns `undefined`
  and the worker is terminated. **Both directions must be red-proved**, because `hasExactKeys`
  rejects both and only the first was possible under `hasOnly`.
- Given a new table-level projection member that passes the guard, when the response travels through
  `engine-client.ts`'s settle path, then the member is present in what `App.tsx` receives — asserted
  by its own test, because that path drops table-level members today.

## Spec Change Log

**2026-09-05 — review pass 1. `review_loop_iteration` stays 0: NO loopback.** Three review layers ran
(blind-hunter, edge-case-hunter, verification-gap). **No `intent_gap` and no `bad_spec`** — the frozen
intent held, and the non-frozen sections did not mislead the implementer. The architecture the plan gate
settled was verified correct under review: the `TableColumnsProjection` surface, the two-member design,
the three-arm command shape and the new wire pin all stood. Findings routed **22 patch (21 applied, 1 refused on measurement) · 6 defer
(DW-220–DW-225) · 2 reject**; the implementation was not reverted and no section was re-derived.

**One patch was refused on measurement rather than applied, and that was right.** P6 alleged the colour
picker lacked a synchronous in-flight guard, so a drag would emit one command per intermediate colour.
Instrumented, the handler is entered three times and **exactly one command goes out**: React classifies
`change` as discrete and flushes the state update before the next event, so `busy` is already true on
entry two. A `pendingRef` would have been a third gate deciding nothing, whose only release path is a
later render. The burst was kept as a property test and shown non-vacuous — removing BOTH existing
guards yields three commands.

**KEEP on any future re-derivation:** the resolved half of every pair comes from `resolveHeaderStyle`
called at the projection site, never recomputed in TypeScript; `headerHeight` and `altRowBackground`
carry ONE member each, because absence for them resolves to nothing rather than to an inherited value;
clearing writes the zero `Presence` and never an explicit null; the `isTableColumns` guard must admit
exactly what the file door admits (a negative `headerHeight` loads, so a `>= 0` browser bound killed the
worker silently — that is what P2 fixed); `tableCommandTarget` is called before any op or field
validation in all three arms, so a refusal never names a field on an element that does not exist; and
`headerStyleFor` stays the sole source of a writable `HeaderStyle` pointer — the writer census caught a
patch that took a raw pointer around it.

**The audit that produced the most value was mutation, not reading.** Six guards were mutated at step-03
before any reviewer ran; five reddened and one did not — `setTableHeaderHeight`'s clear/null refusal could
be **deleted outright with every test still green**, because the command then fell through to
"headerHeight must be a positive length", a different refusal for a different reason located at the same
`table.headerHeight`. `refusalSaysWhy` now pins the sentence. **A refusal asserted only by its DataPath
is not asserted.**

## Design Notes

**The surface D-12.3.1 named is not the surface the table editor reads, and the correction upgrades
the story's guard obligation.** D-12.3.1's checklist — the Go `CanvasProjection` struct, its
construction in `page_setup.go`, the TS type, `hasOnly` plus the typed-key validator, and the sorted
set in `canvas_projection_wire_test.go` — was inherited from Story 17.3 and picks the surface before
measuring which projection the table editor reads. **Its principle transfers exactly and is
unchanged: project the resolved value, two members per property, projection and mirror in one
commit.** Its five sites do not, for three measured reasons: `CanvasProjection` is document-level
while these properties are per-table and nothing restricts a document to one table; its zero-value
parity assertion forbids `omitempty` at its top level, so it cannot represent absence; and the editor
does not read it. The correct surface is `TableColumnsProjection`, and it costs more, not less:
`hasExactKeys` is stricter in both directions than `hasOnly`, so the red-proof gains an arm; the
surface is pinned by nothing, so **12.3 is the first story to put load-bearing data on it and adding
the pin is part of this story**; and `engine-client.ts:#settle` silently drops table-level members,
which fails *quietly* — worse than a `hasOnly` mismatch that at least kills the worker loudly.

**The seven fields, and the eleven-row map that discharges D-8.1.2.** `epic-7-8-decision-log.md`'s
D-8.1.2 is a standing rule: *"Any story that walks a `style.X` must state whether it also walks
`headerStyle.X`, and say why."* 12.3 is that rule's archetype and its inverse, so the map is stated
once, completely, for all 11 `Style` fields:

| `Style` field | 12.3 walks `headerStyle.X`? | Why |
|---|---|---|
| `fontFamily` | **in scope** | resolver arm; note `renameFontChain` already rewrites it |
| `fontSize` | **in scope** | resolver arm |
| `lineSpacing` | **in scope** | resolver arm |
| `background` | **in scope** | resolver arm |
| `color` | **in scope** | resolver arm |
| `valign` | **in scope** | resolver arm |
| `align` | **in scope** | resolver arm; a column's own `align` still wins for that column |
| `border` | **deferred** | resolver arm exists, but it is a nested block the cascade treats block-granularly. Trigger: Story 14.8's BORDERS section, or the first request for a header border |
| `padding` | **forbidden** | D-12.4.1 — the panel never authors padding, on a table or anywhere else |
| `bold` | **no arm exists** | `resolveHeaderStyle` has no `Bold` arm at all — see below |
| `italic` | **no arm exists** | `resolveHeaderStyle` has no `Italic` arm at all — see below |

That is 7 in scope, 1 deferred, 1 forbidden, 2 with nowhere to resolve from. **Sixteen projection
members: 7 x 2, plus one each for `altRowBackground` and `headerHeight`** — one, not two, because
neither has a cascade to resolve through, so committed equals resolved for both.

**Do not "fix" that asymmetry.** A resolved member exists to answer *"what will actually be used when
this is absent"*. Where nothing can be absent (`headerHeight` is required), or where absence resolves
to nothing rather than to an inherited value (`altRowBackground` is a flat override with no fallback
level of its own), **the question has no content and a second member would be ceremony** — a
duplicate of the committed value, carried on every projection, that some later reader has to keep in
sync with itself. The seven header-style fields get two because for them the question is real.

**The forward finding, registered here because nobody else has it on a list.** `resolveHeaderStyle`
has no `Bold` and no `Italic` arm. A table's `style.bold` is settable through the property command
today and serializes, but the header cascade never reads it. Epic 11 is about to make bold and italic
real on the render path, and when it lands **a table's header row still will not be able to be bold**,
because the cascade has nowhere to resolve it from. This is not 12.3's to fix — adding arms would be
a rendering change, which this story forbids itself — but it is 12.3's to record, keyed on Epic 11's
resolution story so 11.2 meets it as a named consequence rather than a surprise.

**The `fontSize` default disagreement: filed, not fixed, and here is why that is not inconsistent with
D-12.C.** `folio-format.md` documents a `fontSize` default of **10**; `render.go:defaultFontSizePt` is
**12000** and `resolveHeaderStyle` seeds from it; `CanvasProjection.DefaultFontSize` already ships
**12** to the browser. So the panel presents 12 today. In D-12.C the code was loose and tightening it
moved no rendered bytes. **Here `defaultFontSizePt` seeds the render path, so "fixing" it to 10 moves
every document that does not declare a size — an AD-21 event that relocates the whole golden corpus.**
The deferral entry must name both sources and **state which one ships today (12)**, or a future reader
finds "the format says 10" and treats the code as the bug.

**Why three arms and not one key in `applyPropertyChanges`.** `applyPropertyChanges` never touches
`element.Table` (measured CLOSED), its 23 keys are a flat list with no nesting for a header-style
block, and `width`/`height` are withheld from tables precisely because `projectedSize` derives a
table's height from `HeaderHeight`. `setComponentAsset`'s own doc comment already rules the shape:
anything the `{op,value}` grammar cannot express, or where clear must stay inexpressible, becomes its
own top-level kind. `headerHeight` is exactly that — required, plain `geom.Length`, never absent.

**The spec pins the op grammar and the gate; it does NOT pin the wire arity — that is deliberate.**
This spec is the authority on the shared `findComponent` + `is not a table` gate, on which ops each
arm accepts (`set`/`clear`, `null` refused everywhere, no clear for `headerHeight`), and on the
clear spelling. **The factory's byte-exact test is the single authority on the wire bytes.** Pinning
arity here as well would create a second spelling of one rule — the defect Story 12.2 spent its whole
life on (one predicate, one place) and the defect `command-json-soleness.test.ts` exists to catch. If
the spec and the byte-exact test ever disagreed, the test would be right and the spec would be the
thing nobody updated. **One authority, and it is the executable one.**

**Clearing, and why the spelling matters more than it looks.** A clear writes the zero `Presence`. A
`Set+Null` clear would emit `"altRowBackground": null` — which the serializer's alt-row branch does
not even handle (it would write `""`, which the loader then rejects as a non-hex colour) — and would
change the bytes, burn an undo entry, and raise the document's required format version. It is also
what makes "clear an already-absent field" a genuine no-op under `wasm/engine.go`'s canonical-bytes
short-circuit, which returns an unchanged snapshot with **no error**: a silent success. A panel that
infers "the write landed" from a changed revision will be wrong for a no-op.

**On Story 14.8 — the seam is ruled, not open.** 12.3 is 14.8's prerequisite: it owns the storage and
the commands. **The owner has ruled that 14.8 becomes a restyle story and that the mockup is
corrected to draw what exists.** So 12.3's controls ship now, in the existing editor idiom, and 14.8
will later move them into the drawn HEADER/CELLS/BORDERS grouping. *Show header row* and the
three-way borders preset **are not coming** — the format has no fields for them and inventing them
was refused. Nothing in this spec is scoped around the seam, and no later reader should re-open it.

**The mockup does not draw these fields.** `TableEditor.dc.html`'s HEADER and CELLS sections contain
"Show header row", "Repeat on continuation pages", "Padding", "Row height" and a BORDERS preset —
measured CLOSED, with positive controls on the same file: `altRow`/`alternating`/`zebra` = 0,
"Header height" = 0, any colour control = 0, against 108 hex literals and "Row height" = 1. Those
sections belong to Story 14.8 and draw a different field set. **There is no drawn design for 12.3's
three subjects**; the shipped `TableEditor.tsx` is the only authority for the section's shape, and its
existing `table-editor-config` rows are the pattern to follow. The *placement* — after the matrix —
follows the design.

## Verification

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- expected: rc 1 with **EXACTLY TWO** failures,
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child. **A THIRD failure is a hard
  stop: report before triaging.** (Measured at baseline `f76b5bf`: exactly these two, 14 `ok`.)
- `cd folio-go && gofmt -l .` -- expected: no output. (Baseline: empty.)
- `cd folio-go && go vet ./...` -- expected: no output. (Baseline: empty.)
- `cd folio-designer && npx vitest run` -- expected: rc 0, at or above the floor of **61 files / 866
  tests**. (Measured at baseline: exactly 61 / 866.)
- `cd folio-designer && npx tsc -b` -- expected: rc 0. **Use `tsc -b`, never `npx tsc --noEmit`,
  which typechecks ZERO files here; add `--force` for a gate that matters, since plain `-b` is
  incremental and can exit 0 having checked nothing.** (Baseline: `tsc -b --force` rc 0.)
- `cd folio-designer && npx oxlint` -- expected: rc 0 with **exactly 4** `only-export-components`
  warnings, freshly measured, never carried forward. (Baseline: rc 0, 4 warnings — two in
  `src/preview/pdf-viewer.tsx`, two in `src/App.tsx`.)
- `cd folio-designer && npm run scan:font-hosts` -- expected: rc 0, 0 occurrences. (Baseline: 631
  files scanned, floor 400.)
- `cd folio-designer && npm run scan:host-fonts` -- expected: rc 0, 0 occurrences. (Baseline: 146
  files scanned, floor 86.)
- `cd lint && go build ./... && go vet ./... && gofmt -l . && go test -count=1 ./...` -- expected:
  build/vet clean, gofmt silent, 4 packages `ok`. (Baseline: all four `ok`.)

**Measurement rules that bind on every command above** (this run has paid for each):

- `grep` in this shell is a **function shadowing the binary**; on `folio-designer/src/App.tsx` it
  prints ZERO matching lines while `grep -c` on the same file reports the true count. Use
  `/usr/bin/grep` for anything a conclusion rests on, and give every negative result a **positive
  control on the same file**. `grep -P` is unavailable (BSD grep).
- `git grep` / `git ls-files` are tracked-only, so a population from them is **SAMPLED, not CLOSED**.
  An absence claim must use `/usr/bin/grep -r` or `git grep --untracked`, or state the exclusion.
- Capture `rc` on each command's **own line** — `$?` is clobbered by any intervening command, `echo`
  included. Never `cd` in a compound command; it re-roots later relative paths.

**Not in scope:** the matrix suite and Playwright — end-of-epic cadence, owed at the Epic 12 boundary.

**Manual checks:**

- The unset colour swatch renders with the `property-swatch-unset` treatment, not black.
- Every new control is reachable by keyboard with a visible focus ring, and the matrix's arrow /
  Home / End navigation and `aria-colcount` of 11 are unchanged.

## Suggested Review Order

**The cascade, projected once**

- Start here: the engine's answer is asked for at ONE site and shipped beside the committed value.
  [`table_columns_projection.go:152`](../../folio-go/table_columns_projection.go#L152)

- Sixteen members, not eighteen: `headerHeight` and `altRowBackground` resolve to nothing, so they carry one each.
  [`table_columns_projection.go:17`](../../folio-go/table_columns_projection.go#L17)

- Both legs of the two-level cascade, one subtest per resolved member — four had zero coverage before review.
  [`table_header_style_test.go:636`](../../folio-go/table_header_style_test.go#L636)

**The door: three arms**

- The shared gate all three now reach before any op or field validation.
  [`component_commands.go:2482`](../../folio-go/component_commands.go#L2482)

- Required, so no clear and no null — and containment re-checked, since a table's height IS its header height.
  [`component_commands.go:2553`](../../folio-go/component_commands.go#L2553)

- The op grammar, refusing `null` once for all three arms.
  [`component_commands.go:2501`](../../folio-go/component_commands.go#L2501)

- The sole source of a writable `HeaderStyle` pointer; the writer census enforces it.
  [`component_commands.go:2745`](../../folio-go/component_commands.go#L2745)

- Clearing the last authorable field drops the block, but a hand-authored border survives.
  [`component_commands.go:2761`](../../folio-go/component_commands.go#L2761)

**The wire, pinned for the first time**

- The fifth record: this surface carried load-bearing data with no pin until now.
  [`canvas_projection_wire_test.go:474`](../../folio-go/canvas_projection_wire_test.go#L474)

- `hasExactKeys` rejects both ways, so the guard is red-proved in both directions.
  [`engine-protocol.ts:355`](../../folio-designer/src/engine-protocol.ts#L355)

- The silent-drop fix: a table-level member used to be discarded before `App.tsx` saw it.
  [`engine-client.ts:129`](../../folio-designer/src/engine-client.ts#L129)

**The panel**

- The section sits after the matrix, where the design places it; focus order amended deliberately.
  [`TableEditor.tsx:222`](../../folio-designer/src/TableEditor.tsx#L222)

- Text box and swatch keyed on the committed value, so a commit cannot leave them disagreeing.
  [`TableEditor.tsx:168`](../../folio-designer/src/TableEditor.tsx#L168)

- "Using: …" distinguishes no paint from the renderer's own default ink.
  [`TableEditor.tsx:31`](../../folio-designer/src/TableEditor.tsx#L31)

**Guards and supporting changes**

- A refusal asserted only by its DataPath is not asserted; this pins the sentence.
  [`table_header_style_test.go:110`](../../folio-go/table_header_style_test.go#L110)

- No command may author a header style except this one; `renameFontChain` is the named exception.
  [`table_header_style_test.go:467`](../../folio-go/table_header_style_test.go#L467)

- Restores a scan this change had blinded by declaring the set through a builder.
  [`closedsets_test.go:768`](../../folio-go/internal/template/closedsets_test.go#L768)

## Delivery Log

### 2026-09-05 — done

Baseline `f76b5bf`. Shipped as one commit, `56a903d` on `main`, 26 files, already pushed. Closed by the
closer; the frozen intent block was not touched (verified byte-identical by sha256 before and after this
entry: `5c60b158…bc19b938`, 91 lines / 7395 bytes).

**What shipped, and where the thinking stayed.** Three command arms, a sixteen-member table projection,
a first-ever wire pin on that projection, and a controls block in the table editor after the column
matrix. The load-bearing decision was that **the cascade stays in the engine**. A header style field left
absent falls back to the table's own `style`, then to that field's documented default; the projection
therefore ships the *resolved value* beside the committed one rather than shipping the ingredients and
letting TypeScript re-derive the answer. `resolveHeaderStyle` is called at the projection site and
nowhere else. Had the ingredients gone over the wire instead, the cascade would exist twice, and the
browser's copy would be the one nobody re-checked when the engine's changed.

**Two members per property — except for two properties, deliberately.** Seven header-style fields carry
a committed and a resolved member each; `altRowBackground` and `headerHeight` carry **one each**, for
sixteen rather than eighteen. A resolved member answers *"what will be used when this is absent"*, and
for these two the question has no content: `headerHeight` is required so it can never be absent, and
`altRowBackground` is a flat override with no fallback level of its own, so absence resolves to nothing
rather than to an inherited value. For both, committed **is** resolved. A second member would be a
duplicate carried on every projection that some later reader has to keep in sync with itself. This
asymmetry is intentional and is not to be "fixed".

**The surface correction, and what it cost.** D-12.3.1's checklist named the document-level
`CanvasProjection` and its five sites — inherited from Story 17.3, which picked a surface before
measuring which one the table editor reads. **Its principle transferred unchanged** (project the resolved
value, two members per property, projection and mirror in one commit); **its sites did not.** These
properties are per-table and nothing restricts a document to one table; that projection's zero-value
parity assertion forbids `omitempty` at its top level, so it cannot represent absence; and the table
editor does not read it at all. The correct surface is `TableColumnsProjection`, and it cost *more*, not
less: `hasExactKeys` is stricter in both directions than `hasOnly`, so the red-proof gained an arm.

**12.3 is the first story to put load-bearing data on that surface, so adding the pin was part of the
story.** The projection had shipped unpinned because nothing before this depended on its shape.
Introducing the wire pin in `canvas_projection_wire_test.go` was not incidental tidying; without it the
sixteen members would have been the only unguarded contract in the change.

**The silent-drop fix in `engine-client.ts`, and why it outranked a guard mismatch.** `#settle` discarded
table-level members before `App.tsx` ever saw them. That is the worse failure mode of the two available:
a `hasOnly` mismatch at least kills the worker loudly and blanks the canvas, so somebody notices within
one interaction. A silent drop returns a plausible panel showing stale or empty values, and nothing
anywhere reports a fault. The fix landed in the same commit as the members it carries.

**`headerHeight` marked Required in `folio-format.md` documents what already ships; it changes nothing.**
`parse_bands.go` has always refused a table without it and `serialize.go` has always emitted it
unconditionally — the field was required *by silence*, with no marker and no sentence saying why. Field,
code and document version are all unchanged; only their documentation moved. Same move Story 12.2 made on
the `utcOffset` row.

**The finding that justified the whole review, and it came from mutation rather than reading.** Six guards
were mutated at step-03 before any reviewer ran. Five reddened; one did not. `setTableHeaderHeight`'s
clear/null refusal could be **deleted outright with the entire suite still green**, because the command
then fell through to "headerHeight must be a positive length" — a *different* refusal, for a *different*
reason, located at the *same* `table.headerHeight` DataPath. Every assertion keyed on the path, so the
path kept matching and nothing noticed the sentence had changed. `refusalSaysWhy` now pins the sentence
itself. The fix was verified by **re-running the identical mutation**: green before, red after. **A
refusal asserted only by its DataPath is not asserted.**

**Triage.** 22 patch (21 applied, 1 refused on measurement) · 6 defer (DW-220–DW-225) · 2 reject.
`review_loop_iteration` stayed 0 — no `intent_gap`, no `bad_spec`, no loopback, no section re-derived. The
one refused patch alleged a missing in-flight guard on the colour picker; instrumented, the handler is
entered three times and exactly one command goes out, and the burst test was shown non-vacuous by
removing both existing guards (three commands). DW-216 through DW-225 were filed by the builder and are
already in this commit.

**Process deviation, logged by the builder against itself.** Step-04 directs spawning all three review
layers before reading any output; the builder launched one, read it, then launched the other two. It
judged no contamination — the reviewers are context-free and it triaged nothing until all three had
returned. Recorded here because the story's record carried it nowhere else.

**Gates, measured at close on `24758aa`** (the two commits above `56a903d` touch only decision-log prose;
`epics.md` and every source file are byte-identical to `56a903d`). `folio-go`: rc 1 with **exactly the two
expected failures** — `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child — and **14
packages `ok`** (14 is the count of packages passing; 15 is the count of packages *with* tests, one of
which is the expected failure). `gofmt -l .` empty, `go vet` rc 0. `folio-designer`: `vitest` rc 0, **62
files / 893 tests**, above the 61/866 floor; `tsc -b` and `tsc -b --force` both rc 0; `oxlint` rc 0 with
**exactly 4** `only-export-components` warnings, and the set was confirmed **identical to baseline rather
than merely equinumerous** — `pdf-viewer.tsx` is untouched since `f76b5bf`, and App.tsx from the first
warning to EOF is byte-identical to baseline (sha256 `4cff49ad…8e6179e1`), so the two warnings there are
the same `canvasDisplay` and `placementPoint` exports renumbered by the −2 line shift from four hunks all
at or above line 2513. Scans: `scan:font-hosts` 635 files (635 tracked + 0 untracked), `scan:host-fonts`
149 files (149 tracked + 0 untracked), 0 occurrences each. Both totals rose from baseline (631 / 146) by
exactly the count of new files this story added in each scan's root — 4 repo-wide, 3 under
`folio-designer/src` — and all of them now read as *tracked* rather than *untracked*, which is the
widened population from Story 15.2b behaving as designed. `lint`: build, vet, gofmt clean, 4 packages
`ok`. **Matrix and Playwright were not run** — end-of-epic cadence, owed at the Epic 12 boundary gate.

**Deferred, with owners.** DW-216 (`headerStyle.border` unauthorable; trigger: Story 14.8's BORDERS
section or the first request for a header border) · DW-217 (`resolveHeaderStyle` has no `Bold`/`Italic`
arm, so a table header cannot be bold even after Epic 11 lands; keyed to Epic 11's resolution story) ·
DW-218 (`folio-format.md` says the `fontSize` default is 10, the engine ships 12, and 12 is what every
document already renders with — filed rather than fixed, because changing it is an AD-21 event that
relocates the golden corpus) · DW-219 through DW-225 as filed. Story 14.8's seam is **ruled, not open**:
it becomes a restyle story and the mockup is corrected to draw what exists.
