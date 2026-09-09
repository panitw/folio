---
title: 'A Line is a thickness and a colour; a Rectangle is a fill and a border'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: '9e50b51631a9d8b55787fcc1f67542e4ebb9d6cf'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative, and rewritten at close to describe what actually shipped. The frozen Intent below is
what governed the implementation.*

A Line here is drawn as a very short, very wide filled box, and the panel used to say so out loud: you thinned a hairline by editing a field called H and coloured it by editing one called
Background. That is fixed. A selected Line now offers Thickness, Colour, Length and an orientation;
a selected Rectangle says Fill. Each still writes exactly the value it always wrote, so the same
edits produce the same saved document as before — and this ships with a proof of that.

Three things may look wrong later and are deliberate. The orientation control swaps the two
dimensions in one edit, so it is a single undo step; the panel's side of the command was widened to
carry more than one change at once to allow that, which the engine always permitted. On a square
rule that control is switched off on purpose, with a stated reason: swapping equal sides changes
nothing. And when an edit that moved both dimensions is refused, the panel
withholds only the part of the message that could name the wrong field, and still prints the rest.

Nothing about the file format moved and no engine code changed. Six follow-ups were registered
rather than fixed here, including the engine's own mislabel. The story's own lesson
is that three of its guards passed while the thing they guarded was genuinely broken, because each
checked a name the screen never shows; all three now fail when it is broken.

---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The inspector spells a Line's properties in the engine's implementation terms rather
than the author's. A Line's thickness is offered as `H`, its colour as `Background`, and its length
as `W`; it is additionally offered border and edge controls that belong to the vocabulary of a
rectangle. A Rectangle's fill is likewise called `Background`. Drawing a rule therefore requires
knowing it is implemented as a filled box.

**Approach:** A presentation-layer remap in the panel only. For a single selected Line, the POSITION
section offers **Length** and **Thickness** — mapped onto `width` and `height` according to an
orientation derived from the committed box — plus an orientation control that swaps the two; the BOX
section offers **Colour** in place of Background and withholds the border and edge controls. For a
single selected Rectangle, Background is spelled **Fill** and everything else is untouched. Every
control keeps writing the field it writes today through the existing
`updateComponentProperties` command. No new `PropertyField`, no new command kind, no new serialized
key, no engine change.

The one mechanism this story adds is on the panel's side of that command: the encoder is widened
**additively** so a single `updateComponentProperties` can carry more than one change, which is what
lets the orientation toggle move `width` and `height` together as **one undo step**. The engine
already accepts this — its only cardinality rule on `changes` is that it be non-empty — so nothing
new reaches it, and no existing single-field call site moves.

## Boundaries & Constraints

**Always:**
- **Presentation only.** Every control maps onto an existing `PropertyField` and reaches the engine
  through an existing command kind. **For every gesture that exists today, the wire bytes are
  unchanged** — the relabels write the same field with the same encoding. The orientation toggle is a
  *new* gesture and its bytes are new by definition: one `updateComponentProperties` carrying two keys
  the engine already accepts. Neither case adds a field, a kind, or a serialized key.
- **AD-9** — `.folio` has exactly one legal serialization; **AD-21** — a hash change is a defect
  until proven intended. This story moves neither, and says so as a checked claim.
- **I-5 / AC4** — the panel narrowing its vocabulary is never a document migration. Selecting an
  element writes nothing. A control this story withholds never removes the value the document carries.
- **UX-DR25** — no control loses its accessible name. Names may be renamed, never dropped.
- **D-000.33** — `## Verification` carries unit tests, typecheck, lint and `test:e2e:compile` only.
- **D-000.32** — name every suite that did **not** run, in those words.
- **Millipoints stay the stored unit**; points stay the displayed unit (**D-14.2.Q3**, below).
- Kind-specific spelling is gated on a **single** selection (`single?.type`). A multi-selection keeps
  the generic vocabulary.

**The six guardrails on the encoder widening (D-14.2.Q2) — not optional, they bound the blast radius:**
1. **The widening is ADDITIVE.** The existing single-field call shape must keep compiling and behaving
   identically. Keep `PropertyIntent` as-is and accept `PropertyIntent | ReadonlyArray<PropertyIntent>`,
   or add an overload — **do not convert the singular form into a one-element array at the 13
   construction sites.** **If the widening cannot be made additive, STOP and bring it back** — that is
   a different and larger story, and it stops at this guardrail rather than growing through it.
2. **`PropertyCommitError` matches any field the failing intent carried**, so `errorFor('width')` and
   `errorFor('height')` both resolve for an orientation refusal, and the orientation control declares
   both. A refusal matching neither is the vacuous case — **assert it cannot happen.**
3. **Red-prove the encoder on ordering.** `changes` is a JSON object and Go walks `propertyOrder`, not
   insertion order, so `{width, height}` and `{height, width}` must produce the same document — or the
   encoder has acquired an ordering dependency the engine does not honour.
4. **`command-json-soleness.test.ts` moves with it and must still be able to fail.** It exists to prove
   `command-json.ts` is the sole encoder, and a widened signature quietly dropping a value out of that
   path is exactly what it guards. **Re-run its red proof; do not merely re-green it.**
5. **One undo step is the assertion, not the design note.** Test it as a revision delta: toggle, assert
   the snapshot advanced by exactly one, undo once, assert both dimensions returned. Asserting only
   that the shape swapped would pass on a two-command implementation.
6. **`resizeComponentCommand` stays inert and is NOT retired here.** Retiring it was (b)'s argument and
   (b) is rejected. An export with a definition, one test import and zero production call sites is a
   real finding and stays visible as one; the orchestrator has registered it as **DW-334**.

**Ask First:**
- Anything that would add a `PropertyField`, a command kind, a serialized key, or an engine change.
- **Any Go edit at all, including the `propertyPath` first-key fix** — registered, not this story's.
- **Converting any existing single-field control to a batched intent**, or anything that makes those
  13 intent construction sites move.
- **Retiring `resizeComponentCommand`** or deleting any other export.
- Widening the swept population of `control-vocabulary-contract.test.tsx` beyond `button` /
  `[role="button"]` — that is DW-331 and it is **ruled out of this story**.
- Editing `DESIGN.md`, `EXPERIENCE.md`, `epics.md`, `Main.dc.html`, `sprint-status.yaml` or
  `deferred-work.md`. Send wording to the orchestrator instead.

**Never:**
- Never commit, `git add`, stash, checkout, reset, revert, restore or clean. Never push. Never branch.
- Never touch `fixtures/declared-variants/expected.pdf`, `input.folio` or `signoff.json`.
- Never normalise a document on selection, and never give a withheld field a default that a focus-and-blur
  would commit (`shown: true` — see Design Notes).
- No new `@media` query in `App.css`; no hex/rgb/hsl literal anywhere in it.
- Out of scope, even where trivially adjacent: DW-305, DW-324, DW-325, DW-326, DW-327, DW-328,
  DW-330, DW-331, DW-145, DW-146; the arrow-step granularity of a Thickness field (**D-14.2.Q5** / **DW-335**); the
  identity row's raw `line` / `rect` spelling at `App.tsx:2717`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Horizontal line selected | single `type: 'line'`, `width` 72000, `height` 1000 | POSITION offers `Length (pt)`→`width` and `Thickness (pt)`→`height`; orientation reads horizontal | N/A |
| Vertical line selected | single `type: 'line'`, `width` 1000, `height` 72000 | POSITION offers `Length (pt)`→`height` and `Thickness (pt)`→`width`; orientation reads vertical | N/A |
| Square line (`width == height`) | single `type: 'line'`, 10000 × 10000 | Reads **horizontal** (tie rule); committed values shown unchanged; nothing written | N/A |
| Thickness edited | horizontal line, author types `2` into Thickness | Emits exactly `{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"height":{"op":"set","value":2}}}` — byte-identical to what the `H` field emitted before this story | field-anchored `property-error` on refusal |
| Colour picked on a line | single `type: 'line'` | Emits a `background` `set` change; swatch is named `Pick Colour`, null action `Set Colour null` | field-anchored |
| Line's BOX section | any single `type: 'line'` | Border width, border colour and the four edge checkboxes are **not rendered** | N/A |
| Line that carries a border | single `type: 'line'` with `borderWidth`/`borderColor`/`borderEdges` projected | The withheld capability is **disclosed** by an honest-note; the stored border is untouched | N/A |
| Square rule, orientation offered | single `type: 'line'`, 10000 × 10000 | The non-current segment is **disabled** and carries an accessible reason: a square rule has no orientation to change. No command is sent (**D-14.2.Q7**) | N/A — disabled from mount, so patch 5's focus hazard does not arise |
| Orientation toggled | horizontal line 72000 × 1000, author toggles to vertical | **One** `updateComponentProperties` carrying both `width` and `height`; box becomes 1000 × 72000; Length and Thickness keep their values; snapshot revision advances by exactly **one** | field-anchored on both fields |
| Orientation undone | after the toggle, one undo | Both dimensions return together — proving it was one history entry, not two | N/A |
| Orientation refused | long horizontal rule in a short band, toggled to vertical | Engine refuses via `containComponent`; the refusal is reported beside the control as `component.geometry` | `component.geometry` is accurate for a w/h pair |
| Multi-key refusal path | a `{width, height}` intent refused by the `height` change | Panel prints `elementId` and `message` and **suppresses the field path**, because `propertyPath` would name `component.width` | D-14.2.Q2b |
| Thickness exceeds Length | horizontal line 72000 × 1000, author sets Thickness to 80000 | Commit **succeeds**; the derived orientation now reads vertical and the Length/Thickness labels swap over the two values. **Accepted and stated, not prevented** | N/A |
| Rectangle selected | single `type: 'rect'` | Background spelled **Fill** (`Pick Fill`, `Set Fill null`); border and edge controls unchanged; `W`/`H` unchanged | field-anchored |
| Selection only | a Line is clicked on the canvas | **Zero** `'command'` requests are dispatched | N/A |
| Multi-selection incl. a line | `[text, line]` shift-selected | Generic vocabulary: `Width (pt)`, `Height (pt)`, `Background`; border controls present | unchanged |

</frozen-after-approval>

## Rulings at the Open-Questions gate

Recorded verbatim in effect so the implementer inherits them rather than rediscovering them.

- **D-14.2.1 (orchestrator, against itself) — the NUL-byte trap does not reproduce.** The two NUL
  bytes in `folio-designer/src/App.tsx` are at lines **3700–3701**, not near 3201. Measured here:
  this machine's `grep` is **ugrep 7.8.4**; `grep -In` returned matches both before (`:2564`) and
  after (`:4019`) the NULs with exit 0. The orchestrator confirmed independently
  (`grep -Ic "return"` = 306, exit 0, identical to plain `grep -c`). The "invisible to `grep -I`,
  exit 1, no output" claim had been carried through roughly six dispatches on citation without
  re-measurement. **Still use `grep -a` unconditionally** — a subagent elsewhere may run a different
  grep — but do not repeat the false premise.
- **D-14.2.Q1 — AC2 stands as an outcome; its stated reason is replaced.** D-000.9 item 1
  (`epic-11-14-decision-log.md:861-865`) impeached the AC's because-clause, not its outcome: a border
  on a Line **paints** (`folio-go/element_box.go:124` `elementBoxDeclaration` is kind-agnostic;
  `borderPaints` `:151-157` returns true for a present, non-null, non-empty edge set — re-verified at
  HEAD). Ruling: **hide the border and edge controls, drop the false reason** ("a filled bar has no
  edge set to draw"), state the true one — *a Line is authored as a thickness and a colour* — and
  **render an honest-note when the selected Line actually carries a border**, so a painting property
  is disclosed rather than vanished. The AC text is **not** amended.
  Accepted correction to D-000.9's wording: hiding does **not** collide with I-5, because I-5 forbids
  *deleting* a value and `PropertyDraft` writes nothing on mount (`App.tsx:2810-2818`). The cost of
  hiding is **discoverability, not preservation** — which is why the note is part of the ruling.
- **D-14.2.Q3 — the display unit is points, product-wide.** This discharges one of the three rulings
  `epic-14-context.md:113-115` said were owed before their story is built. `Thickness (pt)`,
  `Length (pt)`. The mockup's `mm` (`Main.dc.html:274, 279, 284, 289`) is mockup fidelity against
  **zero** millimetres in 106 tracked files (D-000.9 item 3). A future story wanting millimetres is
  proposing a product-wide change, not a label.
- **D-14.2.Q2 (engineering lead; recorded in the program log as D-14.2.2) — the orientation
  control commits through route (c): widen
  `updateComponentPropertiesCommand` to carry several changes in one `changes` object.** (a)
  `setComponentBounds`, (b) `resizeComponent` and (d) two sequential commands are rejected.
  **It is not a new command, settled on a fact rather than a reading** — re-verified here at HEAD:
  the engine's *only* cardinality rule on `changes` is a **floor**,
  `folio-go/component_commands.go:1055` refuses `len(changes) == 0`, not `> 1`; `:1068` applies every
  change via `applyPropertyChanges`; and `containComponent` runs **once per id after all changes are
  applied** (`:1071-1072`), which is precisely the transient-shape hazard that killed the
  two-command sequence. The engine was written for multi-key changes and has been waiting for one.
  Against AC1's three prohibitions: the `kind` is untouched; `width` and `height` are already in
  `PropertyField` and in Go's `propertyOrder`; and "what is serialized" is the `.folio` document,
  which ends carrying the same two numbers either way. **Command bytes are not serialization.**
  AC1's own wording points at (c): *"mapped onto the element's height and width **by the panel**"* —
  the panel's vocabulary is `PropertyIntent` over `PropertyField`; the bounds commands are the
  **canvas's** vocabulary, reached from drag. And (c) **never snaps structurally**, because
  `updateComponentProperties` has no `snap` parameter to get wrong, whereas (a)/(b) would make this
  the sole caller in the codebase passing `snap: false` — a latent defect whose trigger is the next
  person tidying the inconsistency and silently destroying every thin rule in every document.
  Correct-by-construction beats correct-by-remembering.
- **D-14.2.Q2b — `propertyPath` mislabels a multi-key refusal, and the remedy is panel-side.**
  `folio-go/component_commands.go:1079-1087` returns **the first key in canonical order**, not the key
  that failed — correct while `changes` always had one member, a mislabel the moment it has two. With
  `{width, height}` a refusal raised by `height` reports `component.width`, and that reaches the
  screen. **Ruled remedy: suppress the path only when BOTH hold — the intent carried more than one
  field, AND the path's last segment is a member of `PropertyField`.** The rule is *never print a
  field name that may be wrong*, **not** *never print anything*. Consequences, each pinned by a test:
  `component.width` on a `{width, height}` intent is **suppressed**; `component.geometry` is
  **printed**, because `geometry` is not a `PropertyField` and it is accurate for a w/h pair — and it
  is the refusal an author will actually hit; `component.changes` is **printed**, because
  `propertyPath`'s fallback names no specific field and so cannot mislabel one; any single-field
  intent is **printed**, exactly as before this story.
  Single-field intents are unchanged. **Do NOT fix `propertyPath` in Go** — that is the correct fix
  and it is an engine change AC1 does not name; the orchestrator has registered it as **DW-333**. The case that
  matters most is unaffected: a `containComponent` refusal reports `component.geometry`
  (`component_commands.go:1073`), which is accurate for a width/height pair.
- **Measured corrections to D-14.2.Q2b's anchors** (reported to the orchestrator; the ruling's
  *intent* is unaffected and stands, only its cited facts move):
  1. The inline error does **not** render `dataPath` alone. Measured at `App.tsx:3341`:
     `` {error.elementId ? `${error.elementId}: ` : ''}{error.dataPath ? `${error.dataPath}: ` : ''}{error.message} `` —
     it prints **both**, concatenated, so a height-caused refusal on a `{width, height}` intent would
     read `e1: component.width: …`. Note this differs from `componentDiagnostic` (`:4098`), which uses
     `elementId ?? dataPath`. **Therefore this spec implements the narrower form of the ruling:
     suppress only `dataPath`, and keep `elementId`, which is accurate.** "Render `message` alone"
     would discard a true term along with the false one. Flagged for overrule.
  2. `field: intent.field` is at **`App.tsx:1199`**, not `:1201`.
  3. The "31 `field:` intent constructions" is the count of the bare token `field:`, which includes
     type declarations (`FieldSpec`, `PropertyCommitError`, function params). The **measured intent
     constructions are 13**, over 9 lines (`:3013, 3040, 3148, 3341, 3493, 3769, 3824, 4013, 4019`) —
     19 of the 31 are `field: '<literal>'` and 10 are the `field, operation` shorthand. The guardrail
     is unchanged and cheaper than priced: **do not touch any of those 13.**
- **D-14.2.Q7 (orchestrator, at step-04) — a square rule disables the orientation segment, with a
  stated reason.** Swapping the dimensions of a square is the identity, so no implementation makes the
  press meaningful; the only choice is whether the panel admits it. An enabled control that visibly
  does nothing is an instrument whose silence is its answer, in the epic whose subject is the panel
  telling the truth. **Disable the non-current segment when `width === height`, give it an accessible
  reason** (a square rule has no orientation to change), and pin both the disabled state and the
  reason with a test. Because it is disabled from mount rather than transiently, the focus hazard in
  patch 5 does not apply. This contradicts no frozen row: the "Square line" row covers only the read,
  and its silence on the toggle is what the builder correctly declined to fill on its own. It sits at
  the same tie boundary as the "Thickness exceeds Length" row.
- **D-14.2.Q4 — add Line and Rectangle render states to `control-vocabulary-contract.test.tsx`.**
  14.1's V2 deferral (R-Q2) was made on the explicit promise that 14.2, 14.3, 14.4 and 14.7 "each own
  one of those surfaces and will make the call with the surface in front of them". This story is the
  first of the four. Shipping new controls under a vocabulary guard that cannot see them would make
  that deferral retroactively hollow.
- **D-14.2.Q6 — the byte-identity proof is wire-byte equality, not document bytes.** A vitest/jsdom
  test **cannot** obtain real serialized bytes: the save path is `App.tsx:1826
  engine.request('serialize')` answered by `wasm/engine.go:189`, the only wasm instantiation in
  production source is `engine.worker.ts` via `importScripts`, and that soleness is *enforced* by
  `engine-ownership-contract.test.ts:56-58`; every designer unit test injects a fake `EngineClient`
  returning a fixed byte constant (`App.test.tsx:209`). Since this story changes **no Go code**,
  identical command bytes for the same gesture is exactly what makes the document bytes identical.
  A Go byte-identity test is **ruled out**: it would have passed before this story existed, and
  D-000.33 would keep it from being run — D-000.32's shape exactly.
- **D-14.2.Q5 — the Thickness arrow step is out of the fence.** `stepThousandths` (`App.tsx:3086`) is
  keyed on `field`, so `height` steps 1pt per arrow press on a field whose sensible range is
  0.25–1pt. Finding accepted, fix **not** in this story; registered as **DW-335**.
- **DW-331 stays open.** Widening the swept population from buttons to checkboxes is the same shape
  as enforcing V2, which 14.1 deliberately deferred. Do not fold it in.

## Code Map

Anchors measured at the dispatch baseline `303b8071abf71bb9fee0dd61992bdfeaa721b16b`, tree clean.
**Every `App.tsx` search used `grep -a`** (see D-14.2.1 — the trap's premise is false but the habit
stands). `App.tsx` is 4350 lines; NUL bytes at 3700–3701.

**The two sections the ACs name — and the precedent for gating them**

- `folio-designer/src/App.tsx:2684` — `ComponentProperties`, the only place the inspector branches on
  element kind. `:2685` `ids`, `:2686` `types`, `:2687` `all(predicate)`, `:2688` `single`
  (`components.length === 1 ? components[0]! : undefined`), `:2691` `table`, `:2692` `image`,
  `:2693` `typographic`. **There is no `line` or `rect` branch anywhere in the panel today** —
  population searched: every `grep -a` hit for `'line'`, `'rect'`, `types.has`, `single?.type` in
  `App.tsx`; those two strings appear only in `paletteItems` (`:53`) and `paletteGlyphs` (`:89-90`).
- `folio-designer/src/App.tsx:2718` — POSITION.
  `{positionFields.map(draftFor)}{all((type) => type !== 'table') && sizeFields.map(draftFor)}`.
  **This line is already the precedent for kind-gating a field group.** This story adds a second
  predicate in the same idiom; it does not invent a mechanism.
- `folio-designer/src/App.tsx:2722` — BOX.
  `{borderFields.map(draftFor)}<BorderEdgesProperty …/>{draftFor(backgroundField)}{draftFor(visibilityField)}<p className="honest-note">…</p>`.
  **BOX has no kind gate at all today** — a Line gets the full border stack.
- `folio-designer/src/App.tsx:2707` — `draftFor(spec)` renders `<PropertyDraft key={spec.field} … />`.
  The React key is the **field**, so a relabel reuses the same instance and the same committed value;
  a reorder of the two size specs is a reorder, not a remount.
- `folio-designer/src/App.tsx:2717` — the identity row renders `single.type` **raw**, so a Line reads
  as the lowercase string `line`. The palette already spells them properly (`:53`).
  **Out of scope** — named here so review does not report it as missed.

**The field specs this story edits**

- `:2558` `FieldSpec = Readonly<{ field: PropertyField; label: string; affix?: string; unit?: string;
  swatch?: true; prose?: true; empty?: string; shown?: true; fx?: FieldExpression }>`.
  **`affix` is the visible text; `label` is the accessible name and is never rendered.** That split is
  the whole mechanism this story uses.
- `:2563` `positionFields` — `X`, `Y`. Untouched.
- `:2564` `sizeFields` — `[{ field: 'width', label: 'Width (pt)', affix: 'W', unit: 'pt' },
  { field: 'height', label: 'Height (pt)', affix: 'H', unit: 'pt' }]`. The epic's complaint, literally.
- `:2638` `borderFields` — `Border width (pt)` / `Border colour`. Withheld for a Line.
- `:2639` `backgroundField` — `{ field: 'background', label: 'Background', affix: 'Background',
  swatch: true, empty: 'none' }`. Relabelled per kind.
- `:2640` `visibilityField`. Untouched.
- `:4019` `BorderEdgesProperty` — one line. `role="group" aria-label="Border edges"`, affix `Edges`,
  four `<input type="checkbox" aria-label={\`Border ${edge}\`}>`, a `Mixed` span, and a
  `Clear Border edges` button that renders only once an edge is set. Withheld for a Line.

**The single render site every one of those flows through**

- `folio-designer/src/App.tsx:2791` `PropertyDraft`; the rendered markup is `:3339-3341`.
  `:3338` `shared` carries `'aria-label': label`. `:3341` derives three further names from the same
  string: `` `Pick ${label}` `` (the swatch), and at the inline actions `` `Clear ${label}` `` and
  `` `Set ${label} null` ``. **Relabelling one `FieldSpec.label` renames up to four accessible names
  at once** — that is the mechanism behind the census work below.
- `:3161` `canClear` — false for `x`, `y`, `width`, `height`, `value`, `expression`; otherwise
  requires a non-empty value. **So Length and Thickness carry no `×`, exactly as `W`/`H` do not.**
- `:3162` `canNull` — `field === 'visibleIf' || field === 'background'`. **Field-keyed, not
  label-keyed**, so a Line's Colour and a Rect's Fill each keep the `∅` action under the new name.
- `:3163` `errorId` is `property-error-${field}` — field-keyed, so error routing survives a relabel
  verbatim.
- `:3168` `description` is derived from `same`/`fx`, **not** from `label`. A relabel does not move it.
- `:3080` `numeric`, `:3086` `stepThousandths` (`field === 'lineSpacing' ? 100 : 1_000`),
  `:3101` `lowest` (`POSITIVE_LENGTH_FIELDS` includes `height`, so the floor is 1 thousandth).
  **`:3086` is D-14.2.Q5 / DW-335 and is not edited.**

**The command surface — nothing here changes**

- `folio-designer/src/component-property-command.ts:17` — `PropertyField`, 23 members, mirrored exactly
  by Go's `propertyOrder` (`folio-go/component_commands.go:1139`). `:19` `PropertyIntent` is
  **single-field**. `:62-69` `updateComponentPropertiesCommand` emits `changes` with exactly one key.
  `:44` `POSITIVE_LENGTH_FIELDS`. `:72-78` `numberLiteral` writes the author's literal byte-for-byte as
  an unquoted JSON number in **points** — no `Number()` round trip.
- `folio-designer/src/App.tsx:1178` `applyProperties` → `:1189`
  `engine.request('command', updateComponentPropertiesCommand(ids, intent))`. **The single property
  commit site.** Refusals surface as the field-anchored `propertyError`.
- Engine allowlist for a line/rect (`folio-go/component_commands.go:1136-1147`): `x`, `y`, `visibleIf`,
  `width`, `height`, `background`, `borderWidth`, `borderColor`, `borderEdges`. Typography is refused
  on a line/rect; padding is table-only (D-12.4.1).
- **`Colour` on a Line maps to `background`, NOT to `style.color`.** DW-146 records that `style.color`
  on a line or rect loads, serializes and round-trips and is **never painted, never validated, never
  projected** (`folio-go/component_commands.go:1163` refuses `color` off text/table). The
  obvious-looking field is the inert one. Do not reach for it.
**The commit-and-error vocabulary the widening moves — five declarations, one predicate**

Measured; this is what route (c) actually costs beyond the encoder itself.

- `folio-designer/src/component-property-command.ts:19` — `PropertyIntent.field` is **singular**.
  `:62-69` `updateComponentPropertiesCommand(ids, intent)` builds `jsonObject([[intent.field, change]])`
  — exactly one key. Both widen, **additively** (guardrail 1).
- `folio-designer/src/App.tsx:1199` — `setPropertyError({ field: intent.field, selectionKey, ...diagnostic })`.
  The inline error is routed by **the intent's own field**, not by the engine's returned path.
- `folio-designer/src/App.tsx:2537` — `PropertyCommitError = Readonly<{ field: PropertyField;
  selectionKey: string; elementId?: string; dataPath?: string; message: string }>`. `field` is singular.
- `folio-designer/src/App.tsx:2706` — `errorFor(field)` matches on **equality**
  (`scopedError.field === field`, plus the `value`/`expression` alias). This is the one predicate.
- `folio-designer/src/App.tsx:3341` (tail) — the render:
  `` {error && <p id={errorId} role="alert" className="property-error">{error.elementId ? `${error.elementId}: ` : ''}{error.dataPath ? `${error.dataPath}: ` : ''}{error.message}</p>} ``.
  **It prints `elementId` AND `dataPath`, concatenated** — unlike `componentDiagnostic` (`:4098`),
  which uses `elementId ?? dataPath`. D-14.2.Q2b's suppression applies to **`dataPath` only**.
- The 13 intent construction sites that must NOT move: `App.tsx:3013, 3040, 3148, 3341, 3493, 3769,
  3824, 4013, 4019`. `:4019` is `BorderEdgesProperty`, which this story withholds for a Line but does
  not edit.

- `folio-designer/src/App.tsx:813` `commitComponent` — the **non-property** commit path used by canvas
  drag (`:2130`). Its failures render as `commitError` at **`App.tsx:2314`**,
  `<p role="alert" className="file-message">` **in the document bar**, not beside the control. This is
  the fact that decides the orientation route.
- `folio-designer/src/component-command.ts:43` `resizeComponentCommand(id, width, height, snap)` —
  written and unit-tested, **zero production call sites**. Verified: the definition, one import at
  `component-command.test.ts:2`, one assertion at `:9`, nothing else in `src` or `e2e`.
  `:49` `setComponentBoundsCommand` — shipped, used at `App.tsx:2130`.
  Both take `snap`; the 6pt grid (`folio-go/page_setup.go:19-21`) would destroy a 1pt thickness, so
  either would need `snap: false` — a deviation every other caller does not make.
  `updateComponentProperties` never snaps.

**The engine's own confirmation of the mapping**

- `folio-go/internal/template/model.go:404-411` — `ElementType` is a closed set of five;
  `:419-461` one `Element` struct for all of them. `internal/template/parse_bands.go:315` —
  `case ElementLine, ElementRect:` with the body comment `// no extra fields`.
- Format spec, `_bmad-output/specs/spec-folio/folio-format.md:465` — *"`line`, `rect` — no extra
  fields; both are drawn from `style.border` and `style.background`."*
- `folio-go/component_commands.go:1729-1740` — the only place the two kinds differ: a created Line
  gets `Background: "#000000"`, a created Rect gets a 1pt black `Border`. `:1639`
  `lineDropHeight = 1000` (1pt). `folio-go/element_box_test.go:333-334` carries the sentence
  *"a rule's declared height is its thickness"*. **Height is thickness, width is length, background is
  colour — already true in the engine; this story only says so in the panel.**
- `folio-go/internal/template/serialize.go:44` sorts keys at every level; `:16-19` records that there
  is no `omitempty` and absence is carried by `Presence.Set` alone; `internal/template/decimal.go:220`
  formats from int64 millipoints with no float. `width`, `height`, `style.background` and
  `style.border.*` are all format-1.0 keys, so nothing this story maps can move the version
  (`internal/template/version.go:188`).
- `folio-go/wasm/engine.go:288-292` — a command producing bytes equal to the current bytes is **not
  committed at all**: no revision bump, no undo entry, no dirty flag.

**The guard that must see this change, and what it costs**

- `folio-designer/src/control-vocabulary-contract.test.tsx` — 14.1's rendered-output contract.
  - `:294-323` `states` — **four**: `design · nothing selected`, `design · a text element selected`,
    `design · a table element selected`, `preview`. `:285-286` the fixture holds only `e1` (text) and
    `e7` (table). **No Line state and no Rectangle state exist**, so every change this story makes is
    invisible to the guard until D-14.2.Q4's states are added.
  - `:176-178` `controlsIn` sweeps `'button, [role="button"]'` **only**. `<input type="checkbox">` is
    not in the population — which is why removing the four edge checkboxes for a Line reds nothing,
    and why `Border edges` sits in `UNDER_ARITY_GROUPS` (DW-331).
  - `:120-126` `treatmentOf` — `word` if the visible text matches `/[\p{L}\p{N}]/u` (`:68`), else
    `glyph` if it contains an `<svg>` or any non-alphanumeric text, else `empty`.
  - `:211-220` R1 (V3a, per class token), `:224-230` R2 (V3b, per named group), `:234-238` R3 (V4,
    a non-word control with no accessible name), `:267-278` R4 (the V2 census, a **multiset** diff
    keyed `` `${state} · ${name}` `` at `:261`).
  - **R1, R2 and R3 run per render state (`:495`, `:499`, `:503`); R4 runs across states (`:507`).**
    That asymmetry is **DW-328** and is **out of scope** — but it is why a family spelled one way in
    one state and another way in a second state would not red. If this story's controls land in a
    class family that spans states, say so; do not fix DW-328.
  - `:351-379` `V2_CENSUS` — 20 entries, 11 distinct names. The two that live in **BOX** and therefore
    move under a relabel are `Set Background null` and `Set Visible if null`; the other three
    (`Clear Font size (pt)`, `Clear Line spacing`, `Show fonts`) are TYPOGRAPHY-only and a Line renders
    no TYPOGRAPHY section (`typographic`, `:2693`, gates `:2720`).
  - `:398-415` the floors: `PER_STATE_CONTROL_FLOOR = 15`, `CONTROL_FLOOR = 100`,
    `CLASS_FAMILY_FLOOR = 15`, `GROUP_INSTANCE_FLOOR = 18`, plus `CHECKED_GROUPS` and
    `UNDER_ARITY_GROUPS`. `:449-463` R0 asserts **four exact set relations** over group names — a new
    named group that R2 now checks **must** be added to `CHECKED_GROUPS` or R0 reds.
  - **Kind-gating protects the existing census.** Relabelling `background` only when
    `single?.type === 'line' | 'rect'` leaves the text and table states spelling it `Background`, so
    the four existing `Set Background null` / `Set Visible if null` rows survive untouched.

**The guards a new test file must not break**

- `folio-designer/src/canvas-authority-contract.test.ts` — any new `*.test.ts(x)` under `src/` joins its
  corpus (`:14-16`) and is scanned with comments stripped (`:228-248`, `:280-282`). It must not name
  `getComputedStyle`, `getBoundingClientRect`, `measureText`, `document.fonts`, `devicePixelRatio`, or
  any `offset*` / `client*` / `scroll*` identifier (`:19-45`). Population floors are `>=` (`:743-775`)
  so adding files is free, but `tests` must contain only `.ts`/`.tsx`. `:483-485` forbids reusing the
  basenames `App.css`, `App.tsx`, `App.test.tsx`, `font-catalogue.test.ts`. `:342-345` — `App.css` must
  contain **exactly one** `@media`, and it must be `prefers-reduced-motion: reduce`.
- `folio-designer/src/design-contract.test.ts:84-91` forbids any `#hex`/`rgb(`/`hsl(` literal in
  `App.css`. D-000.9's closing note warns that this file and `property-prose-height.test.ts` both assert
  over **raw source text** and *"any Epic 14 restyle fires them."*
- `folio-designer/vite.config.ts:36` — `include: ['src/**/*.test.{ts,tsx}', …]`, no `exclude`, so a new
  file under `src/` is collected with no config change.

**Existing tests that touch this territory**

- **No test anywhere selects a Line.** Population searched: all of `folio-designer/src` and
  `folio-designer/e2e`, recursive, `grep -a "type: 'line'"` — zero hits, exit 0. Positive control on
  the identical command shape for `'rect'` returned 10+ hits (`App.test.tsx:915, 1276, 1297, 2367,
  3909, 6857`; `engine-protocol.test.ts:79, 85, 86, 100`). **This story writes the project's first
  Line-selection coverage from scratch, which is exactly how much the suite's greenness is worth
  here: on this surface, nothing was being watched before.**
- `folio-designer/src/App.test.tsx:2346` — the BOX-section test. Drives `Pick Border colour` (`:2357`)
  and `Pick Background` (`:2361`) and asserts exact wire bytes (`:2360`, `:2363`). Selection is a
  **text** element, so kind-gating leaves it green.
- `folio-designer/src/App.test.tsx:3985` — **the brittle one.** An exact-equality array
  `['X (pt)', 'Y (pt)', 'Width (pt)', 'Height (pt)', 'Font size (pt)', 'Line spacing',
  'Border width (pt)']`. A kind-gated change should leave it alone; **if it moves, that is a signal
  that the gate leaked, not a chore.**
- `folio-designer/src/App.test.tsx:3906` — a text **+ rect** mixed selection asserting `Width (pt)`.
  This is why kind spelling is gated on `single`, not on `types.has`.
- `folio-designer/src/component-property-command.test.ts:8` — the exact-wire-byte idiom this story's
  byte proof copies.
- Engine fake: `folio-designer/src/App.test.tsx:209` —
  `({ request }) as unknown as EngineClient`, truth injected through `initialSnapshot`. No wasm, no
  `vi.mock` of the engine. Same one-liner at `control-vocabulary-contract.test.tsx:287`.

**The design record — what it says and, mostly, does not**

- `…/mockups/Main.dc.html` draws a **Text** inspector and only that (`:264-267`). Line and Rectangle
  appear solely as palette rows (`:89-94`, `:95-100`). **There is no drawn Line inspector, no
  Thickness, no Fill, no Length, no orientation control, and no edge checkboxes anywhere in the file.**
  BOX draws four rows: `Border` collapsed to `none` (`:332`), `Padding` (`:336`, already ruled out by
  D-12.4.1), `Background` (`:341`, that literal word), `Visibility` (`:347`). Population searched: the
  whole file for `POSITION|BOX|Background|Border|Fill|Thickness|Line|Rect`; positive control —
  `Line`/`Rectangle` hit twice each and `Background` hit at `:341`. Per D-14.8.1's principle
  (*"a mockup that draws capabilities the product does not have is a drawing to correct, not a
  specification to implement"*) this is the mirror case, and **the orchestrator is annotating the
  story's `Design:` line in `epics.md`. Do not touch the mockup.**
- `EXPERIENCE.md:86` — *"**Contextual to selection.** Position, size, typography, alignment, border,
  padding, background, visibility, binding (FR5)"*. "Contextual to selection" is the clause that
  authorises per-kind vocabulary. `:282` — every icon-only control carries an accessible name.
- `DESIGN.md` rules on **nothing** here: every hit for line/rect/fill/border is a colour token or the
  1px border-weight rule. **Do not edit it** (Epic 16's recorded discriminator is unrelated and 14.1
  was ruled not to edit it either).

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/App.tsx` -- derive the Line's orientation from the committed box and add the
      per-kind field specs -- a `lineOrientation(component)` helper returning horizontal when
      `height <= width` (**ties read horizontal**, stated), and spec builders that map
      `Length (pt)`/`Thickness (pt)` onto `width`/`height` for a horizontal line and onto
      `height`/`width` for a vertical one. `affix` carries the visible word, `label` the accessible
      name; **`field` is untouched in every case** — that is what makes AC1's "no new field, no new
      command" true rather than asserted. Derivation is read-only: it never writes.

- [x] `folio-designer/src/App.tsx:2718` -- gate POSITION's size fields on kind -- a second predicate in
      the idiom the line already uses for tables. A single Line gets Length/Thickness; everything else,
      including any multi-selection, keeps `sizeFields` verbatim so `App.test.tsx:3985` and `:3906`
      stay green.

- [x] `folio-designer/src/App.tsx:2722` -- gate BOX on kind, relabel the background field, and add the
      disclosure note -- withhold `borderFields` and `BorderEdgesProperty` for a single Line
      (**D-14.2.Q1**); spell `background` as `Colour` for a Line and `Fill` for a Rectangle, leaving
      every other selection spelling it `Background`; render an honest-note **only when the selected
      Line actually carries a border** in the projection (`borderWidth`, `borderColor` or
      `borderEdges` present — `engine-protocol.ts:309`). Voice per UX-DR24: state the fact, name the
      location, offer no comfort. Withholding a control **must not** clear the stored value, and the
      new specs must use `empty` **without** `shown: true` (see Design Notes — `shown` is a
      commit-on-blur hazard).

- [x] `folio-designer/src/component-property-command.ts` -- widen the encoder to carry several changes
      in one `changes` object (**D-14.2.Q2**, **guardrail 1: additive**) -- accept
      `PropertyIntent | ReadonlyArray<PropertyIntent>` (or add an overload) so the 13 existing
      single-field construction sites keep compiling and emitting **byte-identical** commands. The
      multi-intent form emits one `changes` object with one key per intent. **If this cannot be made
      additive, STOP and return it to the orchestrator** — that is a different, larger story.
      `numberLiteral` (`:72-78`) and `POSITIVE_LENGTH_FIELDS` (`:44`) are unchanged.

- [x] `folio-designer/src/component-property-command.test.ts` -- extend the exact-wire-byte assertions
      (**guardrail 3**) -- assert the single-field form's bytes are **unchanged**, assert the
      multi-intent form's literal bytes, and **red-prove the ordering property**: `{width, height}` and
      `{height, width}` must produce the same document, because Go walks `propertyOrder`
      (`component_commands.go:1082`), not insertion order. An encoder with an ordering dependency the
      engine does not honour is the defect this proof exists to catch.

- [x] `folio-designer/src/command-json-soleness.test.ts` -- move it with the widened signature
      (**guardrail 4**) -- it proves `command-json.ts` is the sole encoder, and a widened signature
      quietly dropping a value out of that path is exactly what it guards. **Re-run its red proof;
      do not merely re-green it.**

- [x] `folio-designer/src/App.tsx:1199, 2537, 2706` -- widen the commit-and-error vocabulary
      (**guardrail 2**) -- `PropertyCommitError` records **every** field the failing intent carried, and
      `errorFor` matches if the queried field is among them, so `errorFor('width')` and
      `errorFor('height')` both resolve for an orientation refusal and the orientation control declares
      both. Single-field behaviour is bit-for-bit unchanged. **Assert that a refusal matching neither
      field cannot happen** — that is the vacuous case guardrail 2 names.

- [x] `folio-designer/src/App.tsx:3341` -- suppress a field path that may be wrong (**D-14.2.Q2b**) --
      when the failing intent carried more than one field, do **not** render `dataPath`; keep
      `elementId` and `message`, both of which stay accurate. Single-field intents render exactly as
      today. Rationale to carry in the code: `propertyPath` returns the **first key in canonical
      order**, so a height-caused refusal on `{width, height}` would print `component.width` — a false
      statement about which field was refused, in the epic whose subject is the panel telling the
      truth. Declining to print a path known to be unreliable is strictly more truthful than printing
      it. **Do not fix `propertyPath` in Go.**

- [x] `folio-designer/src/App.tsx` -- the orientation control -- derives its state read-only from the
      committed box and, when toggled, commits `width` and `height` as **one multi-intent
      `updateComponentProperties`**, which is **one undo step** and never snaps. `SegmentedProperty`
      (`App.tsx:4003`) cannot host it unmodified — its `field` prop is typed `'align' | 'valign'` and
      it emits a single-field intent; widen it or add a sibling, and if it lands as a `role="group"`
      register the group name in `CHECKED_GROUPS` (R0 relation #2 reds otherwise). A refusal —
      `containComponent` will refuse rotating a long rule inside a short band — must surface beside the
      control, reported as `component.geometry` (`component_commands.go:1073`), which is accurate.
      Offer the control for a **single** Line only.

- [x] `folio-designer/src/line-rect-vocabulary.test.tsx` -- NEW. The story's own coverage, and the
      project's first Line-selection tests. Cover every row of the I/O matrix: the two orientations
      and the square tie; the withheld border and edge controls for a Line; the disclosure note firing
      only when a border is projected; `Fill` on a Rectangle with its border controls intact; the
      generic vocabulary surviving on a multi-selection; and **accessible names present on every
      renamed control** (UX-DR25). Build a `lineComponent` / `rectComponent` fixture in the shape of
      `CanvasProjection['components'][number]`; drive selection by clicking the canvas element's
      accessible name, as `App.test.tsx:2370` does. **Name no measurement identifier**
      (`canvas-authority-contract.test.ts:19-45`).

- [x] `folio-designer/src/line-rect-vocabulary.test.tsx` -- the byte proof (**D-14.2.Q6**) -- assert the
      **exact wire bytes** emitted for each remapped control against a literal JSON string, in the
      `component-property-command.test.ts:8` idiom, and assert they are **byte-identical to what the
      pre-story `W`/`H`/`Background` controls emitted for the same gesture**. Separately assert that
      selecting a Line dispatches **zero** `'command'` requests (AC4). State in a comment what this
      does and does not prove, and name the residual: the real serialize path is exercised only by
      `e2e/browser-native-roundtrip.spec.ts:328-329`, which this story **compiles and does not run**.

- [x] `folio-designer/src/line-rect-vocabulary.test.tsx` -- pin the derived-label flip as its own case
      -- setting Thickness above Length makes `height > width`, the derivation re-reads as vertical,
      and **the two labels swap over the two values** with no command in between. The commit succeeds;
      this is **accepted and stated, not prevented** — latching the orientation would need stored
      state, which AC1 forbids. **Put the reason in the test's comment**: a derived label is a function
      of committed state and therefore moves when that state moves. That comment is what stops the next
      reader from "fixing" it into a latch. **If in practice the flip is worse than it reads on paper —
      e.g. it fires mid-typing through the draft/blur cycle so a value moves under the cursor — STOP
      and return it to the orchestrator.** That is a design fork for the owner, not a thing to absorb.

- [x] `folio-designer/src/line-rect-vocabulary.test.tsx` -- the one-undo-step assertion
      (**guardrail 5**) -- test it as a **revision delta**: toggle orientation, assert the snapshot
      advanced by exactly one, undo once, assert **both** dimensions returned together. Asserting only
      that the shape swapped would pass on a two-command implementation, which is the thing this
      guardrail exists to exclude. Also assert the refusal path resolves for **both** `width` and
      `height` (guardrail 2) and that a refusal matching neither cannot occur.

- [x] `folio-designer/src/control-vocabulary-contract.test.tsx` -- add the Line and Rectangle render
      states (**D-14.2.Q4**) -- extend the fixture with a line and a rect component, add the two states
      to `states` (`:294-323`) and to the declared-name assertion (`:477`), and **derive the new census
      rows by executing the sweep and reading what it reports** — do not hand-write them from this
      spec. Register any new named group in `CHECKED_GROUPS`, confirm each new state clears
      `PER_STATE_CONTROL_FLOOR = 15`, and confirm `Border edges` keeps its `UNDER_ARITY_GROUPS`
      standing (text and rect still render it, so its arity is unchanged). **Do not widen the swept
      population** — that is DW-331 and it is ruled out.

- [x] `folio-designer/src/App.css` -- only if the disclosure note or the orientation control needs it --
      tokens only, **no hex/rgb/hsl literal** (`design-contract.test.ts:84-91`) and **no new `@media`**
      (`canvas-authority-contract.test.ts:342-345`). Prefer reusing `.honest-note`, which already exists
      at the two BOX/BINDING sites, and add nothing.

- [x] RED PROOFS -- **executed, not written** -- mutate, run, observe the red, prove the mutation
      landed, then revert and re-run green. At minimum: (1) revert the POSITION gate so a Line renders
      `W`/`H`; (2) revert the BOX gate so a Line renders the border stack; (3) revert the Rect `Fill`
      relabel; (4) change one emitted command byte (e.g. make Thickness send `borderWidth`); (5) add a
      command dispatch on selection; (6) **split the orientation commit into two sequential
      single-field commands** — guardrail 5's revision-delta assertion must red, and if it stays green
      the assertion is measuring the shape rather than the history; (7) **reverse the encoder's key
      order** — guardrail 3's ordering proof must stay green, and if it reds the encoder has an
      ordering dependency; (8) **restore `dataPath` printing on a multi-field intent** — D-14.2.Q2b's
      assertion must red on the `component.width` mislabel. **Before trusting any result, diff the file
      and confirm the edit is present** — 14.1 hit a mutation that looked proved by a green run where
      the edit never landed. Each proof must name the assertion it woke.

**Acceptance Criteria:**

- Given a selected Line, when the inspector is shown, then it offers Thickness, Colour, Length and an
  orientation, and the emitted command bytes for each are byte-identical to those the `H`/`Background`/`W`
  controls emitted before this story — no new `PropertyField`, no new command kind, no new serialized key.
- Given a selected Line, when the inspector is shown, then the border controls and the four edge
  checkboxes are not offered, and a Line that carries a border discloses it rather than hiding it silently.
- Given a selected Rectangle, when the inspector is shown, then its background is labelled Fill and its
  border and edge controls remain.
- Given a Line or Rectangle authored before this story, or hand-edited to a shape the panel's vocabulary
  does not describe, when it is selected, then the panel shows the engine's committed values without
  rewriting them, and dispatches zero commands.
- Given every control this story adds, renames or withholds, when it is reached by keyboard or read by a
  screen reader, then its accessible name is present and unambiguous (UX-DR25), and the control-vocabulary
  contract sweeps it in a declared Line or Rectangle state.
- Given a Line whose orientation is toggled, when the change commits, then it is **one** command and
  **one** history entry — a single undo returns both dimensions together — and no snapping occurs.
- Given any single-field property edit anywhere in the panel, when it commits, then its command bytes and
  its error presentation are **identical to before this story**: the encoder widening is additive and the
  13 existing intent construction sites are untouched.
- Given a refusal of a multi-field intent, when it is shown, then the panel states the element and the
  message and **does not name a field path**, because the path the engine returns is the first key in
  canonical order rather than the key that failed.

## Spec Change Log

- **2026-09-09, orchestrator, at CHECKPOINT 1 before approval.** Two edits inside the frozen block,
  made while the builder held.
  1. **Renumbered two spec-local rulings that collided with the program decision log.** The log's
     `D-14.2.2` is the route-(c) ruling; this spec used `D-14.2.2` for the display unit and
     `D-14.2.3` for the arrow step. Spec-local rulings are now uniformly `D-14.2.Q<n>` —
     display unit is **Q3**, arrow step is **Q5** — and Q2 carries a pointer to the log's D-14.2.2.
     `D-14.2.1` is unchanged because it genuinely *is* the log's D-14.2.1. Deferred findings now
     carry their registered numbers: **DW-333** (`propertyPath`), **DW-334**
     (`resizeComponentCommand`), **DW-335** (arrow step).
  2. **Added one I/O matrix row: Thickness exceeding Length.** Orientation is *derived* from the
     committed box, so setting Thickness above Length flips the derivation and the two labels swap
     over the two values under the author's hands. Latching the orientation would require stored
     state, which AC1 forbids — so the behaviour is **accepted and stated rather than prevented**,
     and it needs a test that pins it. Left undefined it would have been discovered by whoever typed
     a large number into a Thickness field.

- **2026-09-09, orchestrator ruling relayed by the builder, at step-04 triage. Patch, not a
  loopback — and the triage label was corrected on the record.** The builder offered the
  `component.geometry` suppression as an `intent_gap`; it is not one. An intent gap is *the frozen
  block does not say*. **Frozen row 10 says exactly what should happen** and the shipped code did
  something else — a defect **against** intent, whose remedy is a patch. `bad_spec` is for a frozen
  intent that turned out wrong; this frozen intent was right, and the implementation drifted from it
  through a shorthand in a **non-frozen** section. Reverting a sound, heavily-verified diff to
  re-derive it nearly identically would have been the workflow's letter defeating its purpose.
  1. **Amended `D-14.2.Q2b`** from the shorthand "do not print `dataPath`" to the literal rule the AC
     bullet already stated — *does not name **a field path***. Verified independently before ruling:
     `PropertyField` (`component-property-command.ts:17`) has 23 members and **`geometry` is not among
     them**, `propertyPath` returns one of those 23 or the literal `changes`, and `containComponent`
     returns `component.geometry`. The panel can separate them exactly, with a membership test
     against a union that already exists. The code comment claiming "the panel cannot tell the two
     apart" was false, and a false premise carrying a user-visible suppression is the shape this run
     keeps catching. **The frozen block was not touched for this** — rows 10, 11 and the AC bullet
     were consistent all along under a literal reading.
  2. **Added `D-14.2.Q7` and one frozen matrix row** for the square rule: the orientation segment is
     disabled with a stated reason rather than enabled and inert. The frozen block was reopened by
     the orchestrator to add that row; the builder did not modify it on its own authority.
  **KEEP on any future re-derivation:** the kind-gating on `single?.type`; `field` untouched in every
  `FieldSpec`; `empty` without `shown`; the additive encoder widening; the derived-not-stored
  orientation; and the census rows derived by running the sweep rather than hand-written.

## Design Notes

### `shown: true` is the trap, and this story must not step in it

`FieldSpec.empty` is a **placeholder hint**; `FieldSpec.shown` puts the value **in the box**
(`App.tsx:2813`, `useState(inherited(committed))`). With `shown: true`, `blur()` (`:3062`) calls
`commit()`, and because `draft ('#000000') !== committed ('')` **it submits** — a focus-and-tab-away
silently writes the document. Today that is deliberate and confined to `fontSize` and `lineSpacing`,
which write the engine's own projected default. If a Line's Thickness or Colour were given a
`shown: true` invented default, selecting a Line and tabbing through it would normalise the document
and break AC4. **Use `empty` without `shown`.**

### Why the disclosure note is bounded, and where it cannot see

The note fires on what the **projection** carries. Two known blind spots, both out of scope and both
named so review does not report them as defects of this story:

- `folio-go/page_setup.go:1917-1929` — a border that paints no ink projects nothing.
- **DW-145** — an all-edges border declared as `{}` prints, but the canvas shows nothing. Such a Line
  would paint a border the note cannot see.

The note is therefore an honest disclosure of what the panel knows, not a guarantee about the PDF.
Say so in the note's own vicinity rather than implying completeness.

### Why the byte proof is at the wire, not the document

Stated as a ruling above (D-14.2.Q6); repeated here because it is the thing most likely to be
"improved" by a reviewer into something that cannot exist. The engine is unchanged by this story.
Therefore: *same gesture → same command bytes → same document bytes*. The middle term is the only one
a jsdom test can observe, and it is the one this story is actually responsible for. A test that
appeared to compare document bytes would be comparing a fake's fixed constant to itself.

## Verification

**Commands** (D-000.33 — unit tests, typecheck, lint and the e2e compile only):

- `cd folio-designer && npx vitest run` -- expected: every file passes. Baseline at the dispatch commit
  `303b807` was **72 files / 1169 tests / 0 failures**. Report the new counts and **diff the test *name*
  sets** — never the totals — reporting GONE and NEW explicitly.
- `cd folio-designer && npx tsc -b --force` -- expected: exit 0, no diagnostics. **`--force` is
  mandatory**: `npm run typecheck` is `build:wasm && tsc -b`, and `tsc -b` is incremental — it can exit
  0 having checked nothing.
- `cd folio-designer && npx oxlint` -- expected: **exactly 4** `react(only-export-components)` warnings.
  **Re-measure the anchors; never quote them** — they move every story. There is no `--max-warnings`, no
  CI assertion and no test behind this count. If the measurement is 5, report it rather than assuming the
  build broke.
- `cd folio-designer && npm run test:e2e:compile` -- (`tsc -p tsconfig.e2e.json --noEmit`) expected: exit 0.

Capture exit codes **without a pipe** — `$?` after a pipe is the pipe's last command, not the tool's.

**Suites that did NOT run, in those words** (Epic 14 boundary gate, the orchestrator's to run):
the browser suite, the Go suites, the matrix legs, `npm run build` as a gate, the `verify:offline*`
chain, and the font-host scans. Standing Go reds that are **not** regressions and must never be
"fixed": `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)`.

**Offline-release slot cost (D-14.0.1): zero.** This story edits `.tsx` and at most `App.css`; it adds
no emitted asset and every glyph it might add is inline JSX (`vite.config.ts:16`
`assetsInlineLimit: 0`). Current `s1.assetCount` is **62** against `maximumCacheAssets = 64`
(`release-payload.ts:42`), `warnCacheAssets = 56` (`:57`) — margin 2, already over the warning line
before this story. **That 62 was read from the existing `dist/offline-release-manifest.json` in the
tree, not from a clean build.** Per D-000.33 no build is run to confirm a zero.

**Manual checks:**
- Confirm the working tree contains no change under `folio-go/`, `fixtures/`, `_bmad-output/planning-artifacts/`,
  `DESIGN.md`, `EXPERIENCE.md`, `sprint-status.yaml` or `deferred-work.md`. This story is presentation-only;
  a diff outside `folio-designer/src/` is a finding, not a convenience.

## Suggested Review Order

**The vocabulary remap — start here**

- The whole story in one function: new words, existing `field` values, no new anything.
  [`App.tsx:2618`](../../folio-designer/src/App.tsx#L2618)

- Orientation is derived from the committed box, never stored; ties read horizontal.
  [`App.tsx:2613`](../../folio-designer/src/App.tsx#L2613)

- `Colour` for a Line, `Fill` for a Rect, `Background` for everything else — spread, so `field` survives.
  [`App.tsx:2712`](../../folio-designer/src/App.tsx#L2712)

- POSITION gates size fields on kind, in the idiom the table gate already used.
  [`App.tsx:2822`](../../folio-designer/src/App.tsx#L2822)

- BOX withholds the border stack for a Line and discloses a border the document still carries.
  [`App.tsx:2826`](../../folio-designer/src/App.tsx#L2826)

- The disclosure's predicate — three arms, each one now separately pinned.
  [`App.tsx:2731`](../../folio-designer/src/App.tsx#L2731)

**The encoder widening, and the diagnostic it put at risk**

- One list, type derived from it: drift between the union and the runtime check is unrepresentable.
  [`component-property-command.ts:31`](../../folio-designer/src/component-property-command.ts#L31)

- A non-empty tuple makes the empty-intent refusal unrepresentable rather than merely untested.
  [`component-property-command.ts:46`](../../folio-designer/src/component-property-command.ts#L46)

- Additive: the singular form still emits byte-identical commands at all 13 call sites.
  [`component-property-command.ts:117`](../../folio-designer/src/component-property-command.ts#L117)

- The two-part rule: suppress a path only if multi-field AND it names a real field.
  [`App.tsx:3473`](../../folio-designer/src/App.tsx#L3473)

- The error now carries every field the intent did, so both dimensions anchor a refusal.
  [`App.tsx:2552`](../../folio-designer/src/App.tsx#L2552)

**The orientation control**

- A sibling of `SegmentedProperty`, not a widening of it; disabled from mount on a square rule.
  [`App.tsx:4207`](../../folio-designer/src/App.tsx#L4207)

**Guards — the part that decides whether any of the above is real**

- The visible-text assertions; without these the whole remap could silently revert.
  [`line-rect-vocabulary.test.tsx:148`](../../folio-designer/src/line-rect-vocabulary.test.tsx#L148)

- Wire bytes identical to what `H` sent — the byte-identity claim, at the only layer that can observe it.
  [`line-rect-vocabulary.test.tsx:392`](../../folio-designer/src/line-rect-vocabulary.test.tsx#L392)

- The square rule states its reason instead of offering a control that does nothing.
  [`line-rect-vocabulary.test.tsx:506`](../../folio-designer/src/line-rect-vocabulary.test.tsx#L506)

- Line and Rect join the vocabulary sweep, so 14.1's deferral is honoured.
  [`control-vocabulary-contract.test.tsx:331`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L331)

- Floors re-measured at 14.2; a comment that misstates its own baseline is worse than no ratchet.
  [`control-vocabulary-contract.test.tsx:488`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L488)

- Key order is meaningless on the wire, proved from the engine's own source.
  [`component-property-command.test.ts:98`](../../folio-designer/src/component-property-command.test.ts#L98)

## Delivery Log

### 2026-09-09 — done

Baseline `9e50b51`. Shipped in **one commit, `8187a25`**, on `main`, unpushed. Seven files: the
inspector remap and the encoder widening in `folio-designer/src/`, this spec, and the tracker hop.
Nothing under `folio-go/`, `fixtures/`, `planning-artifacts/`, `DESIGN.md` or `EXPERIENCE.md` moved —
the presentation-only fence held, and the manual check in Verification is satisfied by
`git show --stat 8187a25`.

**What actually shipped.** The vocabulary remap is one function: new words over the existing
`PropertyField` values, orientation derived from the committed box and never stored, ties reading
horizontal. Three things beyond the relabel. (1) The encoder was widened **additively** so one
`updateComponentProperties` carries several changes — route (c) of [D-14.2.Q2] / program-log
[D-14.2.2] — which is what makes the orientation toggle one undo step; the singular form still emits
byte-identical commands at all 13 call sites. (2) The `printsDataPath` **two-part rule** from
[D-14.2.Q2b] as it was amended at triage: suppress the path only when the intent carried more than
one field **and** the path's last segment is a `PropertyField` member. `component.geometry` and
`component.changes` print; single-field intents are untouched. (3) [D-14.2.Q7]'s square rule — the
non-current orientation segment is **disabled from mount** with an accessible reason, rather than
enabled and inert. Also applied: [D-14.2.Q1] (border and edge controls withheld, false reason
dropped, an honest note when a Line does carry a border), [D-14.2.Q3] (display unit is points,
product-wide — this discharges one of the three rulings `epic-14-context.md` records as owed),
[D-14.2.Q4] (Line and Rectangle joined the vocabulary sweep, honouring 14.1's V2 deferral),
[D-14.2.Q6] (byte identity proved at the wire, the only layer jsdom can observe). [D-14.2.1] stands
as a correction to this run, not to the code: the NUL-byte trap in `App.tsx` does not reproduce as
it had been cited for six dispatches — the bytes are at 3700–3701 and this host's `grep` reads
through them. **DW-331 was deliberately left open**; [D-14.2.Q5] was accepted as a finding and
registered, not fixed.

**Triage.** 14 patches applied, 6 deferred, **0 rejected**, **0 loopbacks**, `review_loop_iteration`
**0**. The one call that mattered was a label, not a fix: the builder offered the
`component.geometry` suppression as an `intent_gap`; it was ruled a **patch**, because frozen matrix
row 10 said exactly what should happen and the implementation drifted from it through a shorthand in
a non-frozen section. A loopback would have reverted a sound, mutation-verified diff to re-derive it
nearly identically.

**What the review caught, and the lesson this story is worth keeping for.** **Three guards were green
over real defects** — the visible words, the `pt` unit, and the border-disclosure predicate. Every
one of those assertions resolved by `FieldSpec.label`, an accessible name that **is never rendered**,
while the word actually on screen is `affix`. The verification-gap reviewer mutated the Line's words
back to `W`/`H` and both kinds back to `Background` and the whole suite stayed green: the story's
entire visible deliverable was unfalsifiable, inside the guard built to prevent exactly that, in the
epic whose subject is the panel telling the truth. All three were repatched to assert **rendered
text**, and **all three now red when reverted**. A fourth mutation did not land — a regex missed —
and the builder discarded its own green rather than reporting it, which is the failure mode 14.1
shipped past.

**Gates, re-measured at `8187a25` on a clean tree, exit codes captured without a pipe** (D-000.33):
`npx vitest run` **exit 0 — 73 files, 1208 tests, 0 failures**; `npx tsc -b --force` **exit 0**, no
diagnostics; `npx oxlint` **exit 0**, exactly **4** `react(only-export-components)` warnings
(anchors re-measured, not quoted: `pdf-viewer.tsx:17,18`, `App.tsx:4307,4314`);
`npm run test:e2e:compile` **exit 0**. Test **name** sets diffed against the dispatch baseline
`303b807` rather than totals: **zero GONE**, 36 new names plus one new file
(`line-rect-vocabulary.test.tsx`); the runtime delta is +39 because the vocabulary sweep is
table-driven and gained cases without gaining names. The 73 files are 69 under `src/` plus 4
`scripts/*.test.mjs`.

**Suites that did NOT run, in those words:** the browser suite, the Go suites, the matrix legs,
`npm run build` as a gate, the `verify:offline*` chain, and the font-host scans. These are the Epic
14 boundary gate's, and the epic cannot close until they are run. Standing Go reds that are **not**
regressions: `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)`.

**Deferred — all six registered by the owner, all `Status: OPEN`, and all carrying
`Owner: unassigned`, which is a gap the register cannot close on its own.** **DW-336** (`borderProjected`
re-derives a Go invariant in TypeScript with no mirror test), **DW-337** (derived orientation ignores
live drag geometry, so mid-resize the labels contradict the drawn box), **DW-338** (the single-flight
pending block is now copied verbatim four times), **DW-339** (`OrientationProperty` renders
full-bleed outside `.property-grid`, unverified by any run), **DW-340** (independent `pendingRef`s
let a Thickness blur-commit and an orientation click be in flight together), **DW-341** (nothing
asserts the Line panel omits TYPOGRAPHY and `Text colour`). Registered earlier at the plan gate and
still open: **DW-333** (the Go `propertyPath` fix this story muted panel-side), **DW-334**, **DW-335**.
