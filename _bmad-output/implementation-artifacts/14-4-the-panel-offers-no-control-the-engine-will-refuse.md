---
title: 'The panel offers no control the engine will refuse'
type: 'feature'
created: '2026-09-09'
status: 'done'
review_loop_iteration: 0
baseline_commit: '82466e71df8a378ace909cabce6c72e0278fc5f8'
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative — a summary for a reader who is not implementing this. The frozen Intent below governs
implementation. Rewritten at close to describe what shipped.*

Selecting a Line, a Rectangle, an Image or a Table used to show a BINDING section that held no control and
invited the author to connect a data path — which the engine then refused. That section is now absent for
those kinds, and the Data tab states the reason before the pick instead of after it: the connect control is
disabled, naming the kind that cannot take a binding. A table's binding, stated three times and
editable in none of them, is now stated once, naming the table editor as where it is edited. No document
bytes, no command, and nothing in the engine changed.

The rule about which kinds may take a binding is written down once, read by both designer surfaces that
need it, and tied to the engine's own rule by a check that fails if either side moves alone. That tie was a
condition of the pre-flight shipping at all.

Three things are intended, not oversights. The story asks that selecting a component preserves
everything the document stores; that is proven at the panel and against an engine verified unchanged, but
**not** at the save path, because these tests cannot observe a saved file's bytes. The refusal wording is
temporary — a later story replaces the presentation while inheriting the rule. And the panel still tells
every author to pick a path before telling some it was pointless; that is registered and owned by
the same later story.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The inspector's BINDING section (`App.tsx:2946`) has **no kind gate** — for a Line, Rectangle,
Image or Table it renders *"No engine binding on this component. Pick a root scalar in the Data tab."*, which
is an instruction to attempt something Go refuses at `component_commands.go:775-777` (*"only text components
can receive a scalar binding"*). The Data panel's pre-flight ladder (`DataPanel.tsx:17-25`) has four arms and
**none inspects the selected component's kind**, so the refusal arrives only after the command round-trips.
Separately, a Table's binding is stated three times in the inspector (`App.tsx:2945`, `:2946`, `:2947`) and
edited in none of them; the sole editable site is `TableEditor.tsx:194`.

**Approach:** Kind-gate the BINDING section, add a fifth arm to the Data panel's existing `unavailable`
ladder keyed on the selected component's type, and reduce the Table's three statements to one that names
where the value is edited. The type gate becomes **one** exported TypeScript constant consumed at both the
existing projection guard (`engine-protocol.ts:583`) and the new panel gate, tied to Go by a new `describe`
in `engine-bounds-mirror.test.ts`. No Go change, no command change, no document byte changes.

## Boundaries & Constraints

**Always:**

- **I-5 (inherited verbatim from Story 14.2).** *The panel narrowing its vocabulary is never a document
  migration. Selecting an element writes nothing. A control this story withholds never removes the value the
  document carries.* Its accepted correction (D-14.2.Q1) applies here too: hiding does **not** collide with
  I-5, because I-5 forbids *deleting* a value and `PropertyDraft` writes nothing on mount. **The cost of
  hiding is discoverability, not preservation** — which is why disclosure is part of the remedy, not an extra.

- **The pre-flight rule is a mirror and must be registered as one.** `engine-protocol.ts:583`
  (`component.type !== 'text' && component.binding !== undefined → false`) is **already** a TypeScript
  re-derivation of Go's rule, shipped and outside the mirror census. This story does not *create* the DW-336
  shape; it **joins an existing instance of it**. Therefore: exactly **one** exported constant, consumed at
  both sites, with a new `describe` in `engine-bounds-mirror.test.ts` carrying a non-vacuity assertion, an
  agreement assertion, a `sites` consumption assertion, and a two-directional red-proof. **If the mirror
  `describe` does not ship, the panel pre-flight does not ship either.** *Ratified at CHECKPOINT 1 as a
  **constraint, not a task** — it is not something to trade away under schedule pressure.*

- **Mirror the `bindComponentScalar` gate, and name it so.** "Binding" is overloaded. A Table legally takes a
  binding through `configureTableBinding` / `updateTableColumnBinding`
  (`component_commands.go:545`, `:596`). A constant named for *bindability in general* would be wrong on the
  Table row. Name it for scalar binding.

- **The rule being mirrored is the component-type gate and nothing else.** Whether the *picked path* yields a
  scalar at render time is runtime data, decided by `internal/bind` at render, and **D-6.2.1 deliberately
  keeps sample runtime kind out of command legality** (`component_commands_test.go:279-282`). Widening the
  panel's judgement to "is this path bindable" crosses the authority boundary and is forbidden.

- **Voice is UX-DR24** (`epics.md:282`): *states the fact, names the location, offers no comfort.* Reuse the
  four-clause shape 14.2 shipped at `App.tsx:2944` — fact → withholding plus the true reason → the stored
  value is unchanged → the epistemic bound — and reuse the bare `.honest-note` class (`App.css:110`).
  **A second disclosure idiom in this inspector is a vocabulary regression and is not acceptable.**

- **Preservation is by construction and must be red-proved anyway.** There is no TypeScript document model:
  `App.tsx:1941-1943` sends `serialize` bytes straight to the file handle, and `wasm/engine.go:189-195`
  returns a copy of `e.bytes`. Nothing this story does can drop a key. **But 14.2's preservation half was
  never mutation-proved** (its disclosure half was, only after a review catch). Do not inherit that green.

**Ask First:**

- Any change to Go, to `component-command.ts`'s command shape, or to any wire payload. None is needed.
- Adding any new `FieldSpec`. This story authors **no** new field; see the `shown: true` hazard in Design Notes.
- Making a table's binding editable in the main window. That is new capability, it contradicts Story 14.7's
  premise, and it is outside this story's fence (see Open Question 2).
- Any change to the Data panel beyond the fifth `unavailable` arm and its supporting prop — the tree's row
  treatment, badges, dimming and context bar belong to Story 14.6.

**Never:**

- Never change document bytes, the wire format, or any command payload.
- Never give a withheld field a default that a focus-and-blur would commit (`shown: true`).
- Never make the panel judge a *path's* runtime kind. That is D-6.2.1's boundary.
- Never take the `params` pickability defect or the empty-collection pickability defect. Both are real, both
  are registered below, and **both are owned by Story 14.6's acceptance criteria** (`epics.md`, Story 14.6,
  the RUNTIME-badge and TABLE-ONLY-badge rows).
- Never remove `TableEditor.tsx:194`'s editing capability — Story 14.7 relocates those controls and needs
  them to still exist.
- Never remove `TableEditor.tsx:201`'s `<output aria-label="Binding for column N">` display — Story 14.10
  requires the editor to keep *showing* a column's bound field.
- Never edit `sprint-status.yaml`, `deferred-work.md`, `epics.md` or `DESIGN.md`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| BINDING on a text component | single `type: 'text'`, no binding | Section renders unchanged, including the existing note pointing at the Data tab | N/A |
| BINDING on a bound text component | single `type: 'text'`, `binding: 'customer.name'` | The bind chip renders unchanged | N/A |
| BINDING on a non-text component | single `type` in `line`/`rect`/`image`/`table` | The BINDING section is **not rendered at all**, and no text anywhere invites a pick | N/A |
| Multi-selection | `selected.length > 1` | Existing `'Binding is shown for one selected component.'` behaviour is unchanged | N/A |
| Data panel, non-text selected | one `line`/`rect`/`image`/`table` selected, a root scalar picked | Connect is **disabled** and the ladder states the reason *before* the attempt; zero `bindComponentScalar` commands dispatched | Pre-flight, not error |
| Data panel, text selected | single `type: 'text'`, a root scalar picked | Connect is **enabled**; behaviour byte-identical to today | Go still refuses `params`, rendered as today |
| Data panel, nothing selected | no selection, a path picked | Existing `'Binding unavailable: select one component first.'` arm still wins | N/A |
| Table binding present | single `type: 'table'`, `tableBind: 'transactions[]'` | The value is stated **once**, and the statement names where it is edited | N/A |
| Table binding absent | single `type: 'table'`, `tableBind` absent or `''` | The single statement says so without inviting an inspector-side edit | N/A |
| Selecting any component | any kind clicked on the canvas | **Zero** `command` requests dispatched; the projection still carries every stored value | N/A |

</frozen-after-approval>

## Rulings (plan gate — SETTLED. Carry these; do not re-open them.)

All four were ruled by the engineering lead at CHECKPOINT 1, 2026-09-09, after verifying each premise at
HEAD. **The ruling text below is theirs, quoted; the surrounding evidence is this spec's.**

**D-14.4.Q1 — AC1: HIDE the BINDING section for non-text kinds.** Ruled:
> *"The section holds no control and can never hold a value for these kinds. The epic's own constraint is
> 'absent rather than disabled-and-mysterious', and a section that exists only to explain why it is empty is
> the disabled-and-mysterious case wearing a paragraph."*

Evidence: `App.tsx:2946` contains two `<p>` elements and no control. Go writes `component.Binding` at exactly
one site (`page_setup.go:1784`), inside `if element.Type == template.ElementText` (`:1767`) — grep-verified
sole write; the only other `Binding:` in the tree is `TableColumnProjection`'s, a different value
(`column.Bind`). **So a non-text component can never carry a binding in the projection, hiding suppresses no
value, and AC4 does not bite on AC1.**

**D-14.4.Q2 — AC3: option (a). The editor stays the single editable statement.** Ruled:
> *"Delete the display-only text at `:2945`, keep `Configure columns`, amend `:2947` to name the editor as
> where the binding is edited. **(b) is refused**: it would build a new capability, contradict 14.7's premise
> that collection and row alias stay editable inside the editor, and it is outside your fence."*

**The epic context file was wrong on this point and HAS SINCE BEEN CORRECTED** in commit `82466e7`
(`epic-14-context.md`, the `14.4 ↔ 14.10` bullet, now marked *CORRECTED 2026-09-09*). Its old wording said
14.10 rules that "where" is the main window. 14.10's actual subject is the **per-column bound field**
(`updateTableColumnBinding`); AC3's subject is the **table's own collection** (`configureTableBinding`) —
different values, different commands, and Story 14.7 explicitly keeps collection and row alias editable
inside the table editor. **Read the corrected bullet, not any cached memory of the old one.**

**D-14.4.Q3 — AC2: option (a). 14.4 owns the RULE; Story 14.6 owns the PRESENTATION.** Ruled:
> *"One fifth arm on `DataPanel.tsx`'s existing four-arm `unavailable` ladder, one prop, plus the mirror.
> Deferring AC2 wholly to 14.6 would leave the Go/TS tie unregistered while a third copy of the rule gets
> written."*

**Inheritance note, required by the ruling and addressed to Story 14.6's builder:** the fifth `unavailable`
arm and its wording are **presentation, and 14.6 is expected to replace them.** What 14.6 must inherit and
must NOT re-derive is **the rule** — the exported constant and its mirror `describe`. Reading the fifth arm
as territory to preserve would be a misreading; reading the constant as territory to re-spell inline would
red the `sites` assertion, which is exactly what it is there for.

**D-14.4.Q4 — INCLUDE the adjacent pair.** Ruled:
> *"`maxCanvasBindingString = 256` ↔ `MAX_ENGINE_BINDING_LENGTH = 256`; I confirmed **zero** occurrences of
> either name in `engine-bounds-mirror.test.ts`. One row in a table you are already editing."*

On why this is inside the fence when 14.3's comparable deferral was not:
> *"The difference is not cost. DW-337 was a different subject in a story about selection. This pair is **the
> same subject** — Go/TS agreement on binding validation — in **the same test file** this story is already
> extending. That is inside the fence, not adjacent to it."*

**Findings registered by the orchestrator, NOT by this story. Do not implement them and do not edit
`deferred-work.md`:**
1. **`params.*` is offered by the picker with no dimming.** `sample-data.ts:98,107` attach `segments` to
   params children, so `params.region` is labelled "root scalar candidate"; Go refuses at
   `component_commands.go:749-751` — **before** the element lookup, so a Line plus a params path yields the
   *params* message, not the type message. Owned by Story 14.6's RUNTIME-badge AC.
2. **An empty collection is pickable and Go *accepts* it.** `segments:["items"]` succeeds and canonicalises
   to `"value": "{{items}}"` (`component_commands_test.go:283-287`), deliberately per **D-6.2.1**, and the
   failure surfaces only at render (`internal/bind/text.go:420-423`). **This is the one place the panel
   offers something the engine accepts and the render then fails**, which no AC in this epic names in those
   terms. Owned by Story 14.6's TABLE-ONLY-badge AC.
3. **`TableEditor.tsx` has no test file at all.** The only editable table-binding site in the product is
   covered by nothing but the encoder unit test (`table-column-command.test.ts:13`).

## Code Map

**AC1 — the BINDING section and its missing kind gate**
- `folio-designer/src/App.tsx:2946` — the whole section, one line:
  `<PropertySection title="BINDING" tone="bind">{single?.binding ? <p className="binding-chip">…</p> : <p className="honest-note">{single ? 'No engine binding on this component. Pick a root scalar in the Data tab.' : 'Binding is shown for one selected component.'}</p>}</PropertySection>`
  Its **only** conditions are `single?.binding` and `single`. No kind gate.
- `folio-designer/src/App.tsx:2903` — `const table = single?.type === 'table' ? single : undefined`. The
  established shape for a kind-derived local in this component.
- **The precedent for kind-gating a section** is its own neighbours: `:2941` `{single && types.has('text') && …}`
  (CONTENT), `:2940` `{image && …}`, `:2945` `{table && …}`. Copy one of those, do not invent a new form.
- `folio-go/page_setup.go:1767` — `if element.Type == template.ElementText && element.Value.Set && !element.Value.Null {`
  and `:1784` `component.Binding = stringPointer(binding)`. **Read-only evidence, verified by grep: this is the
  ONLY write to `component.Binding` in the whole Go tree.** Therefore `single.binding` is *always* `undefined`
  for a non-text component, the bind-chip branch is unreachable for them, and **hiding the section hides no
  value**. This is why AC4 does not bite on AC1.

**AC2 — the Data panel's pre-flight ladder**
- `folio-designer/src/DataPanel.tsx:17-25` — the entire `unavailable` cascade, four arms, none kind-aware.
  **This is the existing idiom AC2 needs; add a fifth arm, do not build a parallel mechanism.**
- `folio-designer/src/DataPanel.tsx:34` — `<p className="binding-status" role="status">{unavailable ?? …}</p>`.
  The reason is already announced here.
- `folio-designer/src/DataPanel.tsx:35` — the Connect button, `disabled={Boolean(unavailable) || bindingBusy || !onConnect}`.
  **A fifth arm disables it for free.**
- `folio-designer/src/App.tsx:2464` — the `DataPanel` call site. Passes `selectedComponentId` but **never the
  type**. This is the one prop to add; the type is already in hand on the projection.
- `folio-designer/src/DataPanel.tsx:6-8` — the file header comment: *"It never turns a display path into Folio
  syntax or decides whether a component/path pairing is legal; the one opaque command does that in Go."*
  **AC2 narrows this claim and the comment must be amended in the same commit**, or the file will assert
  something false about itself. Note it is *already* partly false — arm 3 (`'The selected tree item is not a
  root scalar candidate.'`) is a legality judgement.
- `folio-designer/src/App.tsx:900-923` — `bindPickedPath`. Checks engine, busy, single selection, in-flight.
  No kind check. **Read-only: the pre-flight belongs in the ladder, not here** — the ladder is what disables
  the button, and a guard here would refuse silently.
- `folio-go/component_commands.go:775-777` — the authority:
  `if element.Type != template.ElementText { … "only text components can receive a scalar binding" }`.
  **Verified by repo-wide grep: this string occurs exactly once in the tree.** Fires after `findComponent`
  (`:771`), before mutation (`:782`).
- `folio-designer/src/DataPanel.tsx:36` + `App.tsx:4439` — where a refusal renders today
  (`role="alert"`, prefixed `elementId ?? dataPath`). **The post-hoc path AC2 makes unreachable for this case.
  Do not delete it** — it still serves `params`.

**The mirror — the instrument, and the copy that already exists**
- `folio-designer/src/engine-protocol.ts:583` — `if (component.type !== 'text' && component.binding !== undefined) return false`.
  **The shipped, unregistered TS copy of Go's rule.** Inside `isCanvas`; a false return drops the whole
  snapshot and blanks the canvas.
- `folio-designer/src/engine-bounds-mirror.test.ts` — 6 `describe` blocks at `:110`, `:207`, `:321`, `:414`,
  `:490`, `:563`. **Follow `positive length rule mirror` (`:414`) exactly** — it is the predicate-shaped one.
- `folio-designer/src/engine-bounds-mirror.test.ts:43-58` — `goSources`, four hardcoded Go paths resolved from
  `import.meta.url`. A fifth entry is expected practice, not a smell.
- `:102-108` — `goConstant` / `tsConstant`, line-anchored `^…$` with `m`.
- `:181-193` — `goPositiveLengthFields`, the **predicate** extractor to copy.
- `:78-88` — the `sites` regex column. **This is the clause most likely to be dropped and is the one that
  matters**: without it a hoisted constant can sit unused while the validator keeps a stale inline literal.
- **Non-vacuity is a separate `it` in every describe** (`:115`, `:212`, `:336`, `:419`, `:495`, `:567`), some
  with custom failure messages telling the maintainer to re-derive rather than delete. **Read-only evidence:
  there is no "reads zero rules and asserts all zero agree" hole in this file. Do not introduce one.**
- `:326-330` — the stated policy: *"WRAP-FRAGILE AND LOUD ABOUT IT (D-000.27)"*.
- `folio-go/internal/template/model.go:406-410` — the closed five `ElementText/Image/Table/Line/Rect`.
- `folio-go/page_setup.go:337` (`maxCanvasBindingString = 256`) ↔ `folio-designer/src/engine-protocol.ts:11`
  (`MAX_ENGINE_BINDING_LENGTH = 256`). **Verified: zero occurrences of either name in the mirror test.** Q4.

**AC3 — the Table's three statements and its one editable site**
- `folio-designer/src/App.tsx:2945` — `{table && <PropertySection title="TABLE"><button …>Configure columns</button><p className="honest-note">Table binding: {table.tableBind ?? 'Not set'} (display only)</p></PropertySection>}`
- `folio-designer/src/App.tsx:2946` — the ungated BINDING section, which also speaks for a Table (AC1 removes it).
- `folio-designer/src/App.tsx:2947` — `{types.has('table') ? 'Table size and binding are not editable here; table geometry is derived from columns.' : 'Only committed engine values are shown. …'}`
- `folio-designer/src/TableEditor.tsx:194` — **the only editable site**: `aria-label="Root collection"` and
  `aria-label="Row alias"`, `onBlur → onConfigure` → `configureTableBindingCommand`
  (`table-column-command.ts:21`) → Go `configureTableBinding` (`component_commands.go:521`).
- Both sites read the **same** value: `tableBind` is `element.Table.Value.Bind` (`page_setup.go:1797`); the
  editor's `table.collection` is the same field (`table_columns_projection.go:171,186`). AC3's premise holds.
- `folio-designer/src/App.test.tsx:2871` — `expect(screen.getByText('Table binding: transactions[] (display only)'))`,
  in the test at `:2858`. **The one test that pins the text AC3 removes. It must be updated, not deleted.**

**AC4 — preservation, and why it is by construction**
- `folio-designer/src/App.tsx:1941-1943` — `serialize` bytes go straight to the file handle. **No TypeScript
  document model exists anywhere in the round trip.**
- `folio-go/wasm/engine.go:189-195` — `Serialize` returns a copy of `e.bytes`.
- `folio-go/internal/template/passthrough_test.go:140` — `TestPassthroughAtEveryObjectLevel`: unknown keys are
  preserved opaquely at ten of eleven object levels (D-1.4.9).
- `folio-designer/src/App.tsx:2849-2851` — `borderProjected`, the predicate shape to copy if a disclosure note
  is needed; `:2845-2848` names its blind spots in a comment. **This is DW-336's subject** — the same species
  as the constant this story registers.
- `folio-designer/src/line-rect-vocabulary.test.tsx:257-269`, `:280-288`, `:291-297`, `:301-306` — 14.2's
  disclosure tests: the `it.each` over predicate arms, the non-vacuous negative differing only in the withheld
  key, and the **DOM-order placement assertion**. Copy all four shapes.
- `folio-designer/src/line-rect-vocabulary.test.tsx:105-118` — `settle()`, deliberately longer than
  `PROSE_COMMIT_DEBOUNCE_MS`. Use it; a shorter flush makes a zero-dispatch assertion vacuous.
- `folio-designer/src/App.css:110` — `.honest-note { … }`. **One rule. Reuse it. Add no CSS.**
  `design-contract.test.ts:84-91` and `canvas-authority-contract.test.ts:342-345` both read `App.css` as raw text.

**Guards this story must stay green against**
- `folio-designer/src/control-vocabulary-contract.test.tsx:177` — sweeps `button, [role="button"]` across four
  render states; `V2_CENSUS` is a multiset keyed `state · name`. **Removing the BINDING section changes no
  button** (it contains none), but the TABLE section's `Configure columns` button is in the population — do
  not rename it.
- `folio-designer/src/canvas-authority-contract.test.ts` — the AD-17 prohibitions and its population floors.
- `folio-designer/src/command-json-soleness.test.ts` — the command encoder is sole. Nothing here adds one.

## Tasks & Acceptance

> **VERSION CONTROL — ABSOLUTE, and it overrides any habit or workflow step.** Do **not** run `git commit`,
> `git add`, `git stash`, `git checkout`, `git reset`, `git revert`, `git restore`, `git clean`, `git push`,
> or `git branch`. The human makes every commit in this repository. Leave your work **unstaged** in the
> working tree and say so in your report. Reading git state (`log`, `status`, `diff`, `show`) is encouraged.
> The one carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`, removed
> in a `finally`, is permitted.
>
> **Also never touch** `fixtures/declared-variants/expected.pdf`, `input.folio`, or `signoff.json` — they are
> human-attested. And never edit `sprint-status.yaml`, `deferred-work.md`, `epics.md` or `DESIGN.md`.

**Execution:**

- [x] `folio-designer/src/engine-protocol.ts` — export **one** constant naming the scalar-binding-capable
      component types (a closed, ordered array), and make `:583` consume it instead of its inline `'text'`
      literal. Behaviour must be byte-identical; this is a hoist, not a change. *Rationale: one TS site for the
      rule, so the mirror has something single to tie.*

- [x] `folio-designer/src/engine-bounds-mirror.test.ts` — add `describe('scalar binding legality mirror')`
      following `positive length rule mirror` (`:414`) **exactly**: (1) a `goScalarBindingTypes` extractor over
      `component_commands.go:775-777`, resolving the `template.ElementText` constant name through
      `internal/template/model.go` as a **fifth** `goSources` entry; (2) a **separate non-vacuity `it`**
      asserting the Go side extracts to a non-empty, expected value, with a custom failure message telling the
      maintainer to re-derive rather than delete; (3) an agreement assertion including the negative half
      (`not.toContain('table')`); (4) a **`sites` consumption assertion naming BOTH consumers** — the
      `engine-protocol.ts:583` guard and the new `DataPanel` gate; (5) a two-directional red-proof, including
      the drift this story specifically invites — a panel that stops reading the constant and re-spells
      `type === 'text'` inline. **Also add the `maxCanvasBindingString` (`page_setup.go:337`) ↔
      `MAX_ENGINE_BINDING_LENGTH` (`engine-protocol.ts:11`) row to the existing pair table — D-14.4.Q4, ruled
      in scope.**

- [x] `folio-designer/src/App.tsx:2946` — kind-gate the BINDING section so it renders **only** for a single
      text component (**D-14.4.Q1: HIDE**). Use the neighbours' existing gate form (`:2940`/`:2941`/`:2945`),
      not a new one. *Rationale: AC1.*

- [x] `folio-designer/src/App.tsx:2464` — pass the single selection's `type` to `DataPanel`. *Rationale: the
      panel cannot state a reason it is not told.*

- [x] `folio-designer/src/DataPanel.tsx:17-25` — add the fifth `unavailable` arm, keyed on the passed type
      against the exported constant. Wording per UX-DR24 — state the fact and name the location, e.g.
      *"A Line cannot take a scalar binding. Only text components can."* Place it **after** the
      `!selectedComponentId` arm so "select one component first" still wins when nothing is selected.
      **Amend the file header comment (`:6-8`) in the same commit** so it stops claiming the panel never judges
      legality. **Add a short comment recording D-14.4.Q3's inheritance note**: this arm is presentation that
      Story 14.6 is expected to replace; the exported constant and its mirror are the part 14.6 must inherit
      rather than re-spell. *Rationale: AC2.*

- [x] `folio-designer/src/App.tsx:2945`/`:2947` — **D-14.4.Q2(a)**: delete the `Table binding: X (display
      only)` text at `:2945`, **keep** the `Configure columns` button, and amend `:2947` so it names the table
      editor as where the binding is edited. Reuse `.honest-note` and the four-clause voice. **Do not build a
      main-window editor — option (b) was explicitly refused.** *Rationale: AC3.*

- [x] `folio-designer/src/App.test.tsx:2858-2871` — update the test pinning
      `'Table binding: transactions[] (display only)'` to assert the new single statement. **Update, do not
      delete** — deleting it removes the only coverage of that surface.

- [x] **Tests for every I/O matrix row.** Extend `folio-designer/src/DataPanel.test.tsx` for the ladder arms
      (including the text-selected row proving the arm does **not** fire) and `line-rect-vocabulary.test.tsx`
      or a sibling for the inspector rows. Assert **rendered text**, never `FieldSpec.label` — 14.2's three
      false greens were all assertions resolving through an accessible name that is never rendered.

- [x] **RED PROOFS — executed, not written.** Each must be proved landed with `diff -a` before its result is
      trusted, and proved restored with `cmp` after. `App.tsx` carries two NUL bytes near lines 3700-3701, so
      **plain `diff` reports "Binary files differ" with zero changed lines** — `diff -a`/`cmp` are mandatory
      and every path must be absolute. Required:
      1. Revert the BINDING kind gate → the AC1 test reds.
      2. Revert the fifth ladder arm → the AC2 tests red **and** a `bindComponentScalar` command is observed
         dispatched for a non-text component.
      3. **Delete Go's `component_commands.go:775-777` gate in memory** → the new mirror `describe` reds.
      4. **Re-spell the panel gate as an inline `type === 'text'`** (constant left hoisted but unused) → the
         `sites` consumption assertion reds. *If this stays green the mirror is decorative.*
      5. **Empty/rename the Go source the extractor reads** → the non-vacuity `it` reds, not the agreement one.
      6. **AC4, and do not inherit 14.2's green here:** add a command dispatch on selection → the
         zero-dispatch assertion reds. 14.2 required this mutation and recorded no outcome for it.

**Acceptance Criteria:**

- Given a selected Line, Rectangle, Image or Table, when the inspector is shown, then no text anywhere in it
  invites the author to pick a data path, and no BINDING control is offered.
- Given the Data panel with a non-text component selected and a root scalar path picked, when the author looks
  at the panel, then the reason it cannot be connected is already stated and Connect is disabled — and no
  `bindComponentScalar` command is dispatched for that gesture.
- Given a selected Table, when the inspector is shown, then its binding is stated exactly once and that
  statement names where the value is edited.
- Given a component of any kind authored before this story, or hand-edited into a shape the panel's vocabulary
  does not describe, when it is selected, then the panel dispatches **zero** commands and the projection still
  carries every stored value.
- Given Go's scalar-binding gate, when it is changed on one side only, then `engine-bounds-mirror.test.ts`
  reds — proved by executing the mutation, not by inspection.

## Spec Change Log

**Entry 1 — 2026-09-09, step-04 review. Patch round; no loopback, no re-derivation.**

*Correction to the SPEC's own wording, not a deviation by the implementation (red proof 5).* The spec
expected that emptying the Go source the extractor reads would red the **non-vacuity `it` and not the
agreement one**. Both red, and that is correct: with the Go extraction returning `[]`,
`toEqual(['text'])` cannot pass either. Making the agreement assertion survive a broken extractor would
build precisely the *"reads zero rules, asserts all zero agree"* hole `engine-bounds-mirror.test.ts`
exists to forbid, and the existing `content-window ceiling` and `locale tag` mirrors behave the same
way. **The spec's phrasing was wrong; the implementation was right. Do not distort the test to match
the spec.**

*Triggering findings.* Three review layers ran. The verification-gap layer **demonstrated four false
greens with executed mutations** — an inspector-side collection editor (the thing D-14.4.Q2 explicitly
refused) shipping green at 1246/0; deleting the `isCanvas` binding guard leaving 1245/1246 green with
only a source-text match reding; dropping a table's stored border from the panel leaving AC4's own
"keeps every projected value" test green; plus a reachable fail-open where `selectedComponentId` is
present and `selectedComponentType` is undefined.

*Why patch and not bad_spec.* Considered and rejected. The root cause is **assertion strength, not
design**: the exported constant, the mirror `describe`, the kind gate and the ladder arm are all
correct and were independently validated (red proof 4 reds only the `sites` assertion while every
behavioural test stays green — the discriminating result showing the mirror catches an untying no
behavioural test can see). Every fix is local and additive. Re-deriving would discard a working mirror
to repair test assertions.

*The spec-level lesson, recorded so a future story inherits it.* This spec's red-proof list required
only **reverting the implementation**. Every one of the four false greens was an **absence or
preservation** claim, and those are invisible to a revert — they are falsified only by **ADDING the
forbidden thing** (an editor that should not exist, a value the panel should still show). **A spec that
asks a story to withhold or preserve something must require mutations that ADD, not only mutations that
REVERT.**

*KEEP on any future re-derivation:* the single exported `SCALAR_BINDING_COMPONENT_TYPES` consumed at
both the `isCanvas` guard and the panel gate; the mirror `describe`'s per-block non-vacuity `it` and its
`sites` consumption column; D-14.4.Q3's inheritance note as a source comment; and the ladder arm's
placement after the `!selectedComponentId` arm.

## Design Notes

**Why this is not a new DW-336, and why that changes the recommendation.** The instinct on reading AC2 is
*"the panel is about to re-derive a Go rule in TypeScript — refuse or demand a mirror."* But
`engine-protocol.ts:583` **already** re-derives it, shipped, untied, and in the harshest possible failure mode:
a one-sided Go edit makes `isCanvas` return false, `parseInbound` return `undefined`, and the canvas go blank —
not a wrong tooltip, a dead editor. The choice is therefore not *"one copy or two"* but *"two copies untied, or
one constant tied."* That is why the mirror is a hard gate on shipping the pre-flight rather than a nice-to-have.

**Why this rule is mirrorable when the band edge was not.** Story 17.4's ruling (its Spec Change Log item 9),
as narrowed by D-12.5.1, splits bounds by what the engine's refusal would have *told* the author: clamp where
the refusal carries no information, send where it names something they need. Go's gate reads
`element.Type` alone — no band, no geometry, no sibling, no page — so it is a property of the FIELD, illegal
whatever document is open and wherever the component sits. And its refusal (`field: "component.id"`, text
restating the selection the author is already looking at) carries **nothing** a pre-flight would destroy.
Contrast the strand floor, whose refusal names *which* component is in the way and which D-12.5.1 therefore
ruled must not be clamped. This rule is on the clamp side of that line, by the ruling's own discriminator.

**The `shown: true` hazard, and why this story is immune by construction.** `FieldSpec.empty` is a placeholder
hint; `FieldSpec.shown` puts the value **in the box**, and `blur()` then commits it because
`draft !== committed` — a focus-and-tab-away silently writes the document. That is the one mechanism by which a
*hiding* story can genuinely violate I-5, and 14.2 named it as such. **This story authors no `FieldSpec` at
all** — it removes a section, adds a ladder arm, and hoists a constant. Keep it that way; the moment a task
grows a `FieldSpec`, this note becomes live.

**What "stated before the attempt" must not become.** The panel may state that *this component kind* cannot
take a scalar binding. It must not state that *this path* is unbindable. D-6.2.1 keeps sample runtime kind out
of command legality on purpose — Go accepts `{{items}}` for a collection and reports the mismatch at render
through AD-14's located diagnostic. A panel that pre-judged path kind would be a second, drifting copy of the
binder, and would be wrong the moment the runtime data differs from the sample.

**What jsdom can and cannot prove here.** It can see which section renders, the ladder's announced text, the
button's `disabled` attribute, and the exact command bytes (or their absence). It **cannot** prove anything
about serialized `.folio` bytes — the only place real serialize bytes are observable is
`e2e/browser-native-roundtrip.spec.ts`, which this cadence **compiles and does not run**. Say so; do not let
AC4 be described as save-path proven. 14.2's equivalent gap is named in its own test source at
`line-rect-vocabulary.test.tsx:356-375`, and that honesty is the thing to copy.

## Verification

**Commands (the per-story cadence, D-000.33):**
- `cd folio-designer && npx vitest run` — expected: all pass. **Baseline: 73 files / 1221 tests / 0 failures.**
  *Measured at `b32d832`. `baseline_commit` is `82466e7`, which is one commit later; the delta touches only
  `_bmad-output/` (verified `git diff --name-only b32d832..82466e7`), so no code moved and the measurement
  and every Code Map anchor below still hold at `82466e7`.* Diff the test **name set**, never totals, and report GONE/NEW. **Note: 1221 assertions resolve
  to only 1213 unique `file :: fullName` strings — 8 names are duplicated, so a SET diff can silently absorb a
  lost test. Diff as a MULTISET and report GONE/NEW with counts.** *Raised at CHECKPOINT 1 and adopted by the
  orchestrator as a standing instruction for every story, superseding the previous "diff name sets" rule.*
- `cd folio-designer && npx tsc -b --force` — expected: clean. **`--force` is mandatory**; `tsc -b` is
  incremental and can exit 0 having checked nothing.
- `cd folio-designer && npx oxlint` — expected: 0 errors and **exactly 4** `only-export-components` warnings.
  Measured at `b32d832` (unchanged at `82466e7`): `preview/pdf-viewer.tsx:17,18` and `App.tsx:4425,4432`. **Re-measure the line
  numbers; they move.**
- `cd folio-designer && npm run test:e2e:compile` — expected: clean. **This is `tsc --noEmit` only. It proves
  compilation; it does not run anything and is not coverage.**

Capture each as `cmd > log 2>&1; echo $?`. The shell is **zsh**: `${PIPESTATUS[0]}` is empty and `$?` after a
pipe is the last stage's. **Quote every glob** (`--include='*.test.ts'`) — unquoted, zsh fails the command
while a following `echo` still prints its reassuring message.

**Suites that DID NOT RUN in this story, named per D-000.32:** the browser suite, the Go suites, the matrix
legs, `npm run build` as a gate, the `verify:offline*` chain, and the font-host scans. These run at the Epic 14
boundary gate, which is the owner's.

**Do not claim any failure is pre-existing** unless it was checked at the baseline commit. If it was not checked, say so.

## Suggested Review Order

**The rule, and the single place it is written down**

- Start here: the whole story reduces to this one closed list.
  [`engine-protocol.ts:194`](../../folio-designer/src/engine-protocol.ts#L194)

- The pre-existing untied copy of Go's rule, now consuming the constant instead of an inline literal.
  [`engine-protocol.ts:620`](../../folio-designer/src/engine-protocol.ts#L620)

- Go's authority, unchanged by this story — the rule the two TS sites answer to.
  [`component_commands.go:775`](../../folio-go/component_commands.go#L775)

**The tie that keeps the two languages honest**

- The new mirror; follows the `positive length rule` block's shape exactly.
  [`engine-bounds-mirror.test.ts:719`](../../folio-designer/src/engine-bounds-mirror.test.ts#L719)

- Non-vacuity first: a broken extractor reds here rather than passing silently.
  [`engine-bounds-mirror.test.ts:726`](../../folio-designer/src/engine-bounds-mirror.test.ts#L726)

- The clause that earns the mirror: catches an inline re-spelling no behavioural test can see.
  [`engine-bounds-mirror.test.ts:758`](../../folio-designer/src/engine-bounds-mirror.test.ts#L758)

- The adjacent untied pair closed under D-14.4.Q4, same subject, same file.
  [`engine-bounds-mirror.test.ts:110`](../../folio-designer/src/engine-bounds-mirror.test.ts#L110)

**AC2 — the reason stated before the attempt**

- One projection lookup feeding id, type and binding; removes the id/type skew.
  [`App.tsx:505`](../../folio-designer/src/App.tsx#L505)

- The fifth ladder arm, failing closed — unknown kind is not bindable.
  [`DataPanel.tsx:81`](../../folio-designer/src/DataPanel.tsx#L81)

- Names the kind, or says the projection lacks it; never invents a kind word.
  [`DataPanel.tsx:82`](../../folio-designer/src/DataPanel.tsx#L82)

**AC1 and AC3 — what the inspector stops offering**

- BINDING withheld for kinds the projection can never carry a binding for.
  [`App.tsx:2973`](../../folio-designer/src/App.tsx#L2973)

- TABLE reduced to the one control that does something.
  [`App.tsx:2972`](../../folio-designer/src/App.tsx#L2972)

- The table's binding stated once, in 14.2's four-clause honest-note voice.
  [`App.tsx:2984`](../../folio-designer/src/App.tsx#L2984)

**The three assertions that were false greens until step-04 caught them**

- Guards a ruling: an inspector-side collection editor now reds this.
  [`binding-vocabulary.test.tsx:165`](../../folio-designer/src/binding-vocabulary.test.tsx#L165)

- The behavioural witness the `isCanvas` guard previously lacked entirely.
  [`engine-protocol.test.ts:564`](../../folio-designer/src/engine-protocol.test.ts#L564)

- Pins the fail-open's reachability: id present, kind unknown.
  [`DataPanel.test.tsx:341`](../../folio-designer/src/DataPanel.test.tsx#L341)

## Delivery Log

### 2026-09-09 — done

Baseline `82466e7`. Shipped in `e70d3a3` on `main`, one commit, ten files (eight of them under
`folio-designer/src`, plus this spec and the tracker). `git diff -- folio-go/` is empty at close: no Go
changed, which is what makes the mirror below a tie rather than a co-edit.

**What shipped.** One exported constant now names the component kinds that may take a scalar binding, and
it is read at both TypeScript sites: the pre-existing `isCanvas` projection guard, which previously carried
an inline re-spelling of Go's rule, and the new Data-panel pre-flight. The inspector's BINDING section is
kind-gated using the neighbours' existing gate form, so it does not render for a Line, Rectangle, Image or
Table. The Data panel gained a fifth arm on its existing four-arm `unavailable` ladder, placed after the
`!selectedComponentId` arm so "select one component first" still wins; the arm disables Connect for free and
announces the reason through the existing `role="status"` line. The Table section keeps `Configure columns`
and loses its display-only binding text; the single surviving statement names the table editor as where the
collection and row alias are edited. `DataPanel.tsx`'s file header comment was amended in the same commit,
because AC2 makes its old claim about never judging legality false — and it now records that the comment was
*already* partly false before this story. A new `describe` in `engine-bounds-mirror.test.ts` ties the
constant to Go, and D-14.4.Q4's adjacent pair (`maxCanvasBindingString` ↔ `MAX_ENGINE_BINDING_LENGTH`) was
added to the existing pair table in the same file.

**Decisions applied, by ID.** D-14.4.Q1 (HIDE the section rather than disable it — a section that exists only
to explain why it is empty is the disabled-and-mysterious case wearing a paragraph). D-14.4.Q2(a) (the table
editor stays the single editable statement; option (b), a main-window editor, was explicitly refused as new
capability contradicting 14.7's premise and outside the fence). D-14.4.Q3(a) (14.4 owns the RULE, 14.6 owns
the PRESENTATION — carried into the source as a comment addressed to 14.6's builder, so the fifth arm is not
misread as territory to preserve). D-14.4.Q4 (the adjacent pair is the same subject in the same file, so
inside the fence — the discriminator is subject, not cost, which is why 14.3's comparable deferral was not).
D-6.2.1 (the panel judges component KIND and never a path's runtime kind; widening it would cross the
authority boundary and build a second, drifting copy of the binder). D-14.2.Q1 and I-5 (hiding is not
deleting; the cost of hiding is discoverability, not preservation, which is why disclosure is part of the
remedy). UX-DR24 (the four-clause voice and the bare `.honest-note` class reused; a second disclosure idiom
in this inspector would have been a vocabulary regression). D-1.4.9 (opaque passthrough is why AC4 holds by
construction). D-000.27 (the mirror is wrap-fragile and loud about it). D-14.2.1 (`diff -a` and `cmp` for
every mutation proof touching `App.tsx`, whose two NUL bytes make plain `diff` report "Binary files differ"
with zero changed lines — a landed and a missing edit look identical). D-000.32 and D-000.33 (unrun suites
named in their own words; heavy suites are the epic-boundary gate).

**Triage.** **11 patched, 5 deferred, 2 rejected. Zero `intent_gap`, zero `bad_spec`, and
`review_loop_iteration` stays 0** — patch round only, no loopback and no re-derivation. The five deferrals
were registered by the orchestrator as **DW-352** (the panel's unconditional note still tells every author to
pick, and the new refusal arm sits after the `!picked` arm, so a Line author is told it was pointless only
after picking), **DW-353** (the refusal message is generic for the one kind that legally does take a binding
— a table, through a different command), **DW-354** (a multi-selection branch is now reachable only for a
mixed selection and is asserted by nothing), **DW-355** (an unvalidated field is narrowed by a cast inside
the very guard that validates it) and **DW-356** (the disabled Connect button's stated reason is not
programmatically tied to it). DW-352 and DW-353 are owned by Story 14.6, which already owns this surface's
presentation. Three further findings were registered by the orchestrator *before* implementation and are not
this story's to fix: **DW-349**, **DW-350** and **DW-351**. The triage tallies above are the orchestrator's,
recorded as handed down; this close did not re-derive them and is not a second review.

**A spec correction, not a deviation (red proof 5).** The spec predicted that emptying the Go source the
extractor reads would red the non-vacuity `it` and *not* the agreement one. Both red, and that is correct:
with the Go extraction returning `[]`, the agreement assertion cannot pass either, and making it survive a
broken extractor would build exactly the "reads zero rules, asserts all zero agree" hole this test file
exists to forbid. The spec's phrasing was wrong and the implementation was right; the test was not distorted
to match the spec.

**What the review caught, and it is the story's real lesson.** The verification-gap layer **demonstrated four
false greens by executing mutations against a green suite** — not by inspection, and not by prediction. The
sharpest one fenced a ruling of mine: AC3's claim that no inspector-side collection editor exists was
asserted by querying an accessible name that exists **only inside a dialog the test never renders**, so
adding the editor D-14.4.Q2 had explicitly refused shipped **green at 1246 passed / 0 failed**. The others:
deleting the `isCanvas` binding guard left 1245 of 1246 green, with only a source-text match reding; dropping
a table's stored border from the panel left AC4's own "keeps every projected value" test green; and a
reachable fail-open where the selection id is present while the kind is absent handed Connect back and
re-enabled the round-trip refusal this story exists to remove. That last one is now the arm's default: an
unknown kind is not bindable. All four now red when the forbidden thing is added.

**The generalisable half, and it belongs to specs rather than to this story.** Every one of the four was an
**absence or preservation** claim, and those are invisible to a revert — the only mutation this spec's
red-proof list required. An absence claim is falsified by **ADDING the forbidden thing**, not by removing the
implementation. A spec that asks a story to withhold or preserve something must require mutations that ADD.

**AC4 is NOT proven at the save path, and must not be described as if it were.** jsdom cannot observe
serialized `.folio` bytes; the only real serialize path in the product is exercised by
`e2e/browser-native-roundtrip.spec.ts`, which this cadence **compiles and does not run**. What *is* proven is
narrower and worth stating exactly: **zero commands dispatched on selection** (flushed past the prose commit
debounce, so the assertion is not vacuous), and a **projection still carrying the values the panel withholds**,
measured **against an engine verified unchanged** by an empty `git diff -- folio-go/`. That is the honest
limit. It is the same gap 14.2 named in its own test source, and naming it is the part to copy.

**Gates, re-measured at close at `e70d3a3` with captured exit codes (`cmd > log 2>&1; echo $?`; the shell is
zsh, where `${PIPESTATUS[0]}` is empty and silently wrong, so no `$?` was taken after a pipe).**
`cd folio-designer && npx vitest run` → **0**, **74 files / 1251 tests / 0 failures**.
`npx tsc -b --force` → **0**, zero bytes of output (`--force` is mandatory; `tsc -b` is incremental and can
exit 0 having checked nothing). `npx oxlint` → **0**, **0 errors and exactly 4 `only-export-components`
warnings**, anchored at close to `preview/pdf-viewer.tsx:17,18` and `App.tsx:4462,4469` (the Verification
section's `App.tsx:4425,4432` is its `b32d832` baseline record; the anchors moved, as it said they would).
`npm run test:e2e:compile` → **0**. `git diff -- folio-go/` → empty.

**The baseline test-name MULTISET diff the Verification section asks for was NOT re-derived at close.** It
needs the suite run at `82466e7`, which needs a checkout or a worktree, and this close is under a standing
prohibition on all git state changes. What was checked instead, without changing git state: the tracked test
file population at `82466e7` versus `e70d3a3` is a strict superset — **zero GONE, one NEW**
(`src/binding-vocabulary.test.tsx`) — which bounds any loss to *within* a surviving file and cannot rule one
out. Totals moved 73 → 74 files and 1221 → 1251 tests; at close the five touched files hold 467 tests
(`binding-vocabulary` 15, `DataPanel` 18, `engine-bounds-mirror` 33, `engine-protocol` 42, `App` 359). A
total is weaker than a name multiset, and the spec's own warning stands: 1221 assertions resolved to 1213
unique names, so a SET diff can silently absorb a lost test.

**Suites that DID NOT RUN in this story, named per D-000.32:** the browser suite, the Go suites, the matrix
legs, `npm run build` as a gate, the `verify:offline*` chain, and the font-host scans. These come due at the
Epic 14 boundary gate, which is the owner's — and it is the reason `epic-14` stays `in-progress` at this
close rather than moving to `done`.
