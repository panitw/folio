---
title: 'Story 12.4 — Padding is honoured outside a table, or the format says it is not'
type: 'feature'
created: '2026-09-05'
status: 'done'
review_loop_iteration: 0
baseline_commit: '75e5e24892c5f5763f33ba7f3c488f3466d89003'
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative, and rewritten after the story shipped. The frozen Intent below governs the
implementation; this section only says what came of it.*

A page's layout language has a setting called padding — the breathing room between the edge of a thing
and its contents. The engine accepted that setting on all five kinds of thing an author can place on a
page, but only one of them, a table, ever does anything with it. On the other four the value went into
the author's file, stayed there, and was read by nothing. The designer could persist a number that could
not mean anything.

This story stopped the designer writing that number, and stopped only that. A padding value aimed at a
text box, an image, a line or a rectangle is now refused, and the refusal names both the setting and the
kind of thing it was aimed at. Aimed at a table, it works exactly as before.

What the story deliberately did **not** do is make the engine stricter about reading. A hand-written file
already carrying padding on a text box still opens, still keeps the value when saved again, and still
accepts every other edit to that same element. The lopsidedness is the point, and it is the owner's
ruling rather than an oversight: the engine honours whatever it is handed, while the designer declines to
author what it knows is meaningless.

The record moved with the code — the format document now says where padding is honoured and where it is
merely tolerated. Two follow-ups were banked rather than fixed here.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The Go command layer accepts `paddingTop|paddingRight|paddingBottom|paddingLeft` on **all
five** element kinds, yet the only consumer on any render path is a table's cell chrome. **Measured at
`179a694` by an in-package probe, not assumed: all 20 kind×key combinations are accepted, and the
canonical bytes change every time.** So this is not "accepted and ignored" — **today the designer can
persist dead data into the author's file on four of the five kinds**: a value the file will carry, the
loader will keep, and nothing will ever paint. Meanwhile `folio-format.md`'s Style table scopes `padding`
to no element kind, and the only in-code record of the owner's ruling is a comment carrying an unsourced
date.

**Approach:** Execute **D-12.4.1** (owner decision, 2026-09-05). Split the existing all-kinds `allowed`
branch in `applyPropertyChanges` so the four padding keys are granted **only** on `ElementTable`, leaving
`background`/`borderWidth`/`borderColor`/`borderEdges` on the all-kinds branch. State the restriction in
the format's Style table **in the format's own voice**, following the `color` row's precedent. Replace the `App.tsx` comment with a
citation to D-12.4.1. Change nothing on the load, render or projection path.

### The ruling, verbatim (D-12.4.1, `epic-11-14-decision-log.md`)

> `style.padding` remains an engine property: **a loaded document keeps it and renders it**, and **the
> panel never offers a control for it**. In addition, the command layer stops accepting `padding*` on
> non-table kinds, which it currently accepts on any component.
>
> **The narrowing has a cost, and it is accepted.** A hand-written document carrying padding on a text
> box still **loads and renders**; what it can no longer do is receive a padding command from the
> designer. That is the intended asymmetry: the engine honours what it is given, and the designer
> refuses to author what it cannot mean.
>
> **Consequences.** No `FieldSpec` may author a padding field. The guard belongs in Go's command layer,
> not in the panel, because the panel not offering a control is not a guarantee — AD-15 says the engine
> owns the document, so the refusal must live where the document is written.

The epic's AC1 offers the owner a two-way fork (padding-as-inset vs. table-only). **That fork is DEAD** —
the ruling is taken, the table-only arm is the one that ships, and the insets branch is not to be specced,
stubbed, or mentioned in code. **Story 14.8's cell-padding acceptance criterion is DROPPED by the same
ruling** and is not this story's to implement or to remove.

## Boundaries & Constraints

**Always:**
- **This is a BRANCH SPLIT, not a narrowing of an existing condition.** The branch at
  `component_commands.go:1122-1126` grants eight keys behind a 5-way disjunction naming every member of
  the closed `ElementType` set. `background`, `borderWidth`, `borderColor`, `borderEdges` must remain
  granted on **all five** kinds; only the four padding keys move to a table-only branch. Calling this a
  "narrowing" implies a compatibility question it does not have.
- **Reuse the refusal that already exists.** `fmt.Errorf("property %s is not editable for %s", key,
  element.Type)` at `:1150`, wrapped by the caller as `componentFailure(id,
  "component."+propertyPath(changes), err.Error())` at `:1043-1045`. `propertyPath`'s canonical order
  (`:1054-1063`) already lists all four padding keys, so the DataPath is correct for free. **No new error
  type, no new wrapper, no new diagnostic code, no new machinery of any kind.**
- **Red-prove by mutating the SUBJECT.** Attempt a padding command on a `text` element and show it is
  **accepted before** the guard and **refused after**. Echo the mutated source line back so an edit that
  never applied cannot read as a pass. Mutating the test's *expectation* proves nothing.
- **The asymmetry must hold after the change, and it is testable.** A loaded document carrying padding on
  a non-table element must still load, round-trip through `SerializeTemplate`, project `paddingTop` onto
  the canvas, and accept **unrelated** property commands on that same element. This is the one way to
  implement the guard wrong: it must inspect the **`changes` map only**, never the element's existing
  style.
- **Keep the `App.tsx` replacement comment at exactly FOUR lines** (currently 1712–1715). `App.test.tsx`
  carries ~13 comments citing `App.tsx` by line number below 1715; a line-count change silently rots all
  of them (nothing reds — that is why it matters).
- **`folio-format.md` is a CONTRACT, not prose — four Go tests read it and CI runs them with `-count=1`
  for exactly that reason.** Two constraints follow, and they are invariants, not notes:
  (a) **`drift_test.go`'s `proseAddsKey` scanner runs on EVERY non-fence line, table rows included** —
  ``regexp.MustCompile("adds\\s+`\"([A-Za-z][A-Za-z0-9_]*)\"`")`` at `:117`, applied at `:196` — so no
  new prose anywhere in that file may contain the literal shape ``adds `"someKey"` ``.
  (b) **The `### style` fence (433–447) and everything under `## Worked example` are BYTE-LOCKED** to
  `worked-example.json` / `asset-example.json` by `goldenfixture_test.go`. Line 796 carries a padding
  block. Editing either fence without regenerating its golden reds two tests.
  What IS safe: a row's **description cell** — the drift extractor reads only the first cell
  (`cells := strings.SplitN(line, "|", 3)`, then `firstCell := cells[1]`, `:180-184`), which is why the
  `color` row already names `` `rect` ``, `` `line` `` and `` `image` `` with the suite green.
- **Version control — every clause. You may NOT `commit`, `add`, `stash`, `checkout`, `reset`, `revert`,
  or `restore`.** Leave all work in the working tree. **Never push. Never create a branch.**

**Ask First:**
- Any edit to `folio-format.md` outside line 457's **description cell** (the text after the second `|`).
- Any new diagnostic code, error type, or format **version bump**. (`align`'s `justify` raises a document
  to `2.0`; padding's restriction is a **command** restriction, not a format change — no bump.)
- Any change to the loader, the renderer, the canvas projection, or the panel's field set.
- Any case where a stated fact in this spec turns out to be false at the tree. **Say so rather than
  implementing it.**

**Never:**
- **`folio-designer/src/engine-protocol.ts`'s four padding keys (`:179`, `:331`, `:356`).** That file is
  the **INBOUND** projection validator (engine→browser), proven by the type chain `CanvasProjection`
  (`:130`) → `EngineSnapshot` (`:119`) → `EngineSuccess` (`:182`) → `EngineInbound` (`:209`) →
  `parseInbound` (`:488`), and by its own header comment at `:128` ("paint-only output from Go"). Padding
  appears **nowhere** on the outbound `EngineRequest`/`parseRequest` side. A loaded document keeps and
  renders its padding, and `page_setup.go` projects it — touching this breaks the loaded-document half of
  the ruling. **Conflating an inbound guard with an outbound one produced a false finding earlier in this
  programme.**
- **`pointFields`, `PropertyField`, `POSITIVE_LENGTH_FIELDS`, `ORIGIN_FLOOR_FIELDS` in
  `component-property-command.ts` — membership AND formatting.** `engine-bounds-mirror.test.ts:264` and
  `:352` read two of them with `^…$`-anchored single-line regexes (`m` flag, `[^\]]*` bodies) and
  drift-prove them by substitution with `not.toBe` guards, so a re-wrap reds **twice**.
- **`parse_bands.go`'s `decodePadding` (`:740-769`) and its single call site (`:714-721`).** Kind-agnostic
  by design; that is the ruling's explicit asymmetry.
- **`page_setup.go:1848-1878`** (kind-blind padding projection) and its wire shape at `:271-274`.
- **`folio-format.md`** line 457's **first** cell, the `### style` fence at 433–447, and anything under
  `## Worked example` (line 796's padding block is byte-locked to `worked-example.json` by
  `goldenfixture_test.go`).
- **Any mention of the command layer, the designer, or a refusal in `folio-format.md`.** RULED by the
  coordinator at CHECKPOINT 1: **the format document describes THE FILE, and it keeps one voice.** It has
  never mentioned the command layer anywhere (measured: zero hits for `updateComponentProperties` /
  `not editable` / `command layer` over the whole file, against 15 for `load error`), and it will not
  acquire a second voice because one story found it convenient. **The command layer's refusal is a
  property of the designer, not of the format** — a hand-authored `.folio` carrying padding on a text
  element is still a valid document, and the format spec must keep saying so. The AC is satisfied by the
  inertness sentence alone: *padding is accepted and inert on non-table kinds* **is** the format saying it
  is not honoured. A bare D-12.4.1 citation is the only pointer the row carries.
- Story 14.8's cell-padding AC — dropped by the ruling, not this story's to touch.
- Any golden fixture, any `.folio` file, any `expected.pdf`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Padding on a table | `updateComponentProperties` with `paddingTop\|Right\|Bottom\|Left` on an `ElementTable` | Accepted, written, canonical bytes change | N/A |
| Padding on the four non-table kinds | Same command on `text`, `image`, `line`, `rect` — 4 kinds × 4 keys = **16 cases** | **Refused**, template bytes unchanged | `*ComponentCommandError` with `Message` = `property <key> is not editable for <kind>`, `DataPath` = `component.<key>`, `ElementID` = the element |
| The other four box keys, unmoved | `background`, `borderWidth`, `borderColor`, `borderEdges` on all five kinds | Still accepted on **all five** | N/A |
| A loaded document already carrying text padding | Hand-authored `style.padding` on a `text` element | Loads, round-trips with padding intact, projects `paddingTop` to the canvas, and **accepts an unrelated `x` command** | N/A — refusal must key off the `changes` map, never the element's existing style |
| The corpus | All 31 `.folio` files + `worked-example.json` | Byte-identical renders; **no** document declares padding off a table (AD-21) | N/A |

</frozen-after-approval>

## Code Map

Every anchor below was re-measured at `179a694` (HEAD, clean tree). The survey behind them (D-12.A) was
taken at `71627a5`; the six intervening commits touched **no** file named here.

- `folio-go/component_commands.go:1112-1431` — `applyPropertyChanges`. Structure: `:1113` seeds
  `x`/`y`/`visibleIf`; `:1114` `propertyOrder` (23 keys, closed); `:1115-1117` `if element.Type !=
  template.ElementTable` grants `width`/`height` — **the direct precedent, a kind-conditional branch in
  the mirror shape**; `:1118-1121` text-only grants `value`/`expression`; **`:1122-1126` the 5-way
  disjunction that is the subject of this story**; `:1127-1134` text-or-table grants the typography set,
  with Story 10.1's rationale comment at `:1128-1130` — **the comment style to imitate**. `allowed` is
  consumed at exactly one site, `:1149-1151`, inside the `for _, key := range propertyOrder` loop
  (`:1144`).
- `folio-go/component_commands.go:1150` — the located refusal, already exact.
- `folio-go/component_commands.go:1043-1045` — `updateComponentPropertiesInPlace`'s wrapper into
  `componentFailure`. `componentFailure` is `:31-33`; `*ComponentCommandError` is `:24-29`.
- `folio-go/component_commands.go:1054-1063` — `propertyPath`; its slice is byte-identical to
  `propertyOrder` and already ends with the four padding keys.
- `folio-go/internal/template/model.go:252-261` — `ElementType`, the closed set of exactly five, with the
  comment `// ElementType is the closed set of element kinds (FR4).`
- `folio-go/component_properties_test.go` (**`package folio`**, in-package) — where the new tests go.
  `:77-97` `TestUpdateComponentPropertiesRejectsTableGeometryAndRollsBackBatch` is the refusal precedent:
  it refuses `width` on the table `e2`, asserts the `"not editable"` substring, and asserts canonical-byte
  neutrality. **Improve on it**: assert the full message and the `DataPath`, not a substring.
- `folio-go/component_commands_test.go:61-83` — `for _, kind := range []string{"text","image","table",
  "line","rect"}` over `createComponent`. **The natural shape for this story**, since padding must be
  refused on four kinds and accepted on one.
- `folio-go/testdata/template/golden/worked-example.json` — what `componentTemplate(t)` loads
  (`component_commands_test.go:25-36`). `e1` = text (pageHeader), `e2` = table (already carrying
  `padding {left:3, right:3}`), `e5` = text (pageFooter). A red-first test can use `e1` for the refusal
  and `e2` for the still-accepted case **without touching any fixture bytes**.
- `_bmad-output/specs/spec-folio/folio-format.md:449-463` — the Style table. `:457` is the bare `padding`
  row. `:463` is the `color` row, the **drafting precedent** (added by `dfb55d8`, 2026-09-02, at the Epic
  10 reconstruction). `:455` is the `align` row, the precedent for *refusal* phrasing. These two are the
  **only** kind-scoped rows in the table's 13 data rows.
- `folio-designer/src/App.tsx:1712-1715` — the comment above `borderFields` (`:1716`). Nine `FieldSpec`
  declarations in the file; none authors a padding field (positive control: `borderWidth` hits at
  `:1716`).
- `folio-designer/src/App.test.tsx:1422,1432` — the existing behavioural guard for the ruling's panel
  half: *"drops the padding rows"*, asserting the four `Padding … (pt)` textboxes are not in the document.
  **Read-only. It already passes and must keep passing.**
- `_bmad-output/implementation-artifacts/sprint-status.yaml:450-451` — the stale pre-ruling comment.
  Nothing mechanical parses it: the only readers are BMAD skill scripts under `.agents/skills/`, and no
  test, lint rule, build script or CI step reads the file.

**Read-only evidence gathered at the plan gate (do not re-derive; re-measure only if you doubt it):**

- **The premise is measured, not assumed.** A throwaway in-package probe at `179a694` created one
  component of each of the five kinds and sent all four padding keys to each: **20 of 20 ACCEPTED**, and
  the canonical bytes changed for every kind. The same probe confirmed the refusal machinery already
  works — `width` on the table `e2` returns `folio: property width is not editable for table` with
  `DataPath = "component.width"`, `ElementID = "e2"`. Probe deleted; tree verified clean.
- **The asymmetry holds today.** A second probe hand-added `"padding": {"left":2,"top":4}` to `e1` (a
  `text` element) in `worked-example.json` in memory: it **loaded**, **round-tripped with padding
  intact**, **projected `paddingTop` onto the canvas**, and **accepted an unrelated `x` command** with
  the padding surviving. All four must still be true after the guard.
- **Byte-neutrality re-verified at the plan gate.** Population: **31** `.folio` files outside
  `node_modules`. Exactly **4** contain `padding` — `fixtures/statement-{1,5,20,50}/input.folio`, all at
  line 19, all on element `e8` whose declaration on line 18 reads `"type": "table"`. Positive control:
  24 of the 31 contain `style`. **Zero non-table padding declarations.** The one non-`.folio` document
  fixture that carries padding, `worked-example.json`, has it on `e2`, also a table. So this story is
  genuinely red-first-then-green with **no golden churn**.
- **Padding-command test coverage today is ZERO in both languages.** Population 224 Go `*_test.go` files
  and 70 designer `*.test.ts*`/`*.spec.ts` files; grep for the four key names returns rc=1 in both.
  Positive control `borderWidth` fires in both (`folio-go/element_box_test.go`; three designer files).
  No test anywhere asserts the `property %s is not editable for %s` sentence.
- **Nothing Story 15.2a added sits underneath this.** Its 167 lines in `component_commands.go` are all
  above dispatch (`refuseDuplicateCommandKeys` at `:91-109` and its helpers, plus a 3-line call at
  `:206-208`). Every function between command entry and `applyPropertyChanges` was read: none rejects,
  renames or transforms a padding key.
- **`folio-format.md` IS asserted over — by four Go tests, run in CI with `-count=1` for exactly that
  reason.** `drift_test.go`'s `TestDriftGoToDoc`/`TestDriftDocToGo`, `goldenfixture_test.go`, and
  `numeric_classification_test.go`. **The table extractor reads only the FIRST cell** —
  `cells := strings.SplitN(line, "|", 3)`, then `firstCell := cells[1]` (`drift_test.go:180-184`) — which
  is precisely why the `color` row can already name `` `rect` ``, `` `line` `` and `` `image` `` in its
  description cell with the drift test green. **One prose scanner does run on every non-fence line,
  table rows included:** ``proseAddsKey = regexp.MustCompile("adds\\s+`\"([A-Za-z][A-Za-z0-9_]*)\"`")``
  (`:117`, applied at `:196`). The new sentence must not contain the literal shape ``adds `"someKey"` ``.
- **Editing the `App.tsx` comment cannot red a source-text guard.** `design-contract.test.ts` **does not
  read `App.tsx` at all** (its six file handles are DESIGN.md, tokens.css, App.css, package.json,
  package-lock.json, tsconfig.app.json). `property-prose-height.test.ts:30` does read it, but its ten
  assertions are all content patterns about the prose textarea — none is a line count, a comment scan, or
  an exact-once count. The one live hazard is `canvas-authority-contract.test.ts`, which scans raw source
  **without stripping comments** against a fixed prohibited-token list (`measureText`,
  `getBoundingClientRect`, `offsetWidth`, `getComputedStyle(`, `ResizeObserver`, `document.fonts`, a
  `text-align: justify`-shaped CSS pair, …). **The replacement comment must not spell any of those.**

## Tasks & Acceptance

**Execution:**

- [x] `folio-go/component_properties_test.go` — **FIRST, and it must be RED.** Add a table-driven test
      that loops the closed kind set (`text`, `image`, `table`, `line`, `rect`) × the four padding keys.
      For the four non-table kinds assert: the command is refused; the error is a `*ComponentCommandError`
      whose `Message` is exactly `property <key> is not editable for <kind>`, whose `DataPath` is exactly
      `component.<key>`, and whose `ElementID` is the element; and `SerializeTemplate` bytes are unchanged
      across the refusal. For `table` assert the command is **accepted**. Follow `:77-97`'s shape but
      assert the **full** message, not the `"not editable"` substring (the unkeyed `:1142` message also
      contains that substring, so a substring assertion cannot tell the two refusals apart). Use `e1`
      (text) and `e2` (table) from the fixture; create `image`/`line`/`rect` with `createComponent` per
      `component_commands_test.go:61-83`. **Run it and record that it FAILS** before writing any
      production code.
- [x] `folio-go/component_properties_test.go` — add a second test pinning the two invariants the guard
      could break: (a) `background`, `borderWidth`, `borderColor` and `borderEdges` are still accepted on
      **all five** kinds; (b) a document hand-carrying `style.padding` on a `text` element still parses,
      round-trips with the padding intact, projects `paddingTop` through `Canvas`, and accepts an
      unrelated `x` property command with the padding surviving. Both are green today — say so in a
      comment, and say which mutation each would catch.
- [x] `folio-go/component_commands.go:1122-1126` — **split the branch.** Leave `background`,
      `borderWidth`, `borderColor`, `borderEdges` on the existing 5-way disjunction; move
      `paddingTop`/`paddingRight`/`paddingBottom`/`paddingLeft` into a new
      `if element.Type == template.ElementTable { … }` block. Add a short rationale comment in the style
      of `:1128-1130`, **citing D-12.4.1 by id**. Re-run the red test and record that it now passes.
- [x] **Red-prove the guard by mutating the SUBJECT.** Revert the branch condition to the all-kinds form
      (or delete the new table-only branch outright — deletion proves the branch is *reached*, not merely
      ordered), re-run, and show the new test goes RED. **Echo the mutated line back** in your report.
      Restore, re-run green, and confirm `git status --porcelain` shows only the intended files.
- [x] `_bmad-output/specs/spec-folio/folio-format.md:457` — extend the `padding` row's **description
      cell only** (everything after the second `|`; leave `` | `padding` | `` byte-identical). State: the
      default; that padding is honoured **only** inside a table's cell chrome, cascading to data cells
      with the rest of `style` (see `:388`); and that on a `text`, `image`, `line` or `rect` a declared
      padding is **accepted and inert** — it loads, it round-trips, and nothing ever consumes it.
      (CORRECTED at review: this clause originally read "it loads, it round-trips, **it projects to the
      canvas**, and nothing ever consumes it". "Canvas" is designer vocabulary that appears nowhere else
      in that document, so the clause violated the very one-voice ruling this task cites. The `color`
      precedent stops at "nothing ever paints it". See the Spec Change Log.) **Write NOTHING about the command layer, the designer, or a refusal** — see
      Boundaries/Never; that inertness sentence IS the format saying padding is not honoured there, and
      the document keeps one voice. Carry a bare **D-12.4.1 (2026-09-05)** citation so a reader can find
      the ruling, and note the **field is unchanged and the document version is not raised**. Follow the
      `color` row's shape at
      `:463`. **Do not write the literal phrase ``adds `"…"` ``.** Keep it to one row on one line.
- [x] `folio-designer/src/App.tsx:1712-1715` — **replace, do not delete**, the comment above
      `borderFields`. It is still the only in-code statement of the ruling; deleting it recreates
      D-000.10's exact failure mode (a rule whose sole record is a comment, one refactor from being
      dropped as stale) in the other direction. The replacement must cite **D-12.4.1** by id and name the
      decision log, keep the substance (the panel does not author padding; a loaded document keeps and
      renders it), and add that the Go command layer now refuses it off a table. **Exactly four lines.**
      Spell none of `canvas-authority-contract.test.ts`'s prohibited tokens.
- [x] **ALREADY DONE — do not repeat.** `_bmad-output/implementation-artifacts/sprint-status.yaml` — the
      workflow's own `sync-sprint-status` sub-step performed this at the start of step-03 and it is
      verified: `epic-12: in-progress` (line 454) and
      `12-4-padding-is-honoured-outside-a-table-or-the-format-says-it-is: in-progress` (line 458), both
      flat under `development_status`, and the stale comment now cites D-12.4.1 as settled. Recorded here
      for the closer. The original instruction follows for the record only:
      replace the stale comment
      (*"12.4 is a RULING story: the owner decides whether padding insets…"*) with one citing D-12.4.1 as
      settled. Set `12-4-padding-is-honoured-outside-a-table-or-the-format-says-it-is` to `in-progress`
      and lift `epic-12` from `backlog` to `in-progress` (12.4 is the epic's first story). **Both are FLAT
      keys under `development_status` — do not nest.** Read the two lines back after editing and assert
      the values actually changed.

**Acceptance Criteria:**

- Given the command layer, when `paddingTop`, `paddingRight`, `paddingBottom` or `paddingLeft` is sent to
  a `text`, `image`, `line` or `rect` element, then it is refused with `Message` exactly
  `property <key> is not editable for <kind>` and `DataPath` exactly `component.<key>`, and the
  template's canonical bytes are unchanged — all **16** combinations.
- Given the command layer, when any of the four padding keys is sent to an `ElementTable`, then it is
  accepted and applied, exactly as today.
- Given a `.folio` document hand-authored with `style.padding` on a `text` element, when it is loaded,
  when it is serialized, when its canvas is projected, and when an unrelated property command is applied
  to that element, then all four succeed and the padding survives — the engine honours what it is given.
- Given the corpus, when every fixture is rendered, then no output byte moves and every golden digest is
  unchanged (AD-21), because no document in the corpus declares padding on a non-table element.
- Given the format's Style table, when a reader looks up `padding`, then the row says it is honoured only
  inside a table's cell chrome and is accepted-and-inert on the other four kinds, and cites D-12.4.1 —
  **and says nothing about the command layer**, because the row describes the file.
- Given `App.tsx`, when a reader reaches `borderFields`, then a four-line comment cites D-12.4.1 by id
  and states the ruling — and `App.test.tsx`'s existing "drops the padding rows" guard still passes.

## Spec Change Log

- **2026-09-05, plan gate → build. HEAD moved between the two; the Code Map is unaffected.** Every anchor
  in this spec was measured at `179a694`. While the spec sat at CHECKPOINT 1 the coordinator committed
  `75e5e24` ("Correct D-12.A: design-contract.test.ts never reads App.tsx"), which is now
  `baseline_commit`. `git diff --stat 179a694..75e5e24` is **one file, +34 lines,
  `epic-11-14-decision-log.md`** — a decision-log correction only. It touches **no** file this spec names
  (checked by name against `component_commands.go`, `model.go`, `parse_bands.go`, `page_setup.go`,
  `App.tsx`, `folio-format.md`, `component-property-command.ts`, `engine-protocol.ts`,
  `component_properties_test.go`, `worked-example.json` — rc=1, none matched). **So every Code Map anchor
  and every baseline figure in `## Verification` remains exact at `baseline_commit`.** The commit is the
  coordinator recording the correction this story's plan gate produced: D-12.A had claimed
  `design-contract.test.ts` asserts over `App.tsx`'s raw source, and it does not read that file at all.

### Review round 1 — 2026-09-05, three layers, no loopback

`review_loop_iteration` stays **0**: no `intent_gap` and no `bad_spec`. Three layers ran (blind-hunter,
edge-case-hunter, verification-gap). **8 patched, 1 deferred, 6 rejected.** Every finding was re-measured
before triage; two reviewer claims were **false at the tree** and are recorded as such below.

**The finding that mattered, and it was demonstrated rather than argued.** The table arm of the new
`TestPaddingPropertyCommandsAreGrantedOnATableAlone` asserted only *"accepted, and some byte moved"*.
Rotating the key→edge map at `component_commands.go:1222` so **every** padding command lands on the wrong
edge left the whole suite **GREEN, exit 0** — a mis-wired table inset would have shipped. Patched with a
`serializedPaddingEdges` read-back that asserts the commanded edge carries the commanded value and the
fixture's other authored edges are undisturbed; the identical mutation now fails with
`paddingTop wrote the wrong edge: padding = map[bottom:5 left:3 right:3], want "top" = 5`. Re-proved by
me after the patch, not taken on report.

**Patched (8):** the edge read-back above · `subjectIDForKind` now asserts `e1`/`e2` really are the kinds
it claims, so fixture drift fails loudly instead of going vacuous · `clear` and `null` ops pinned on all
five kinds (measured: both refused off a table; `clear` accepted on a table and prunes an emptied
`style.padding`; `null` on a table still yields its own `paddingLeft does not support null`, which the new
branch must not shadow) · the mixed-selection refusal pinned as transactional · the format row's
"it projects to the canvas" clause removed · `headerStyle.padding` wins for the header row added, matching
what the `color` row already states and verified at `table_render.go:379-384` · the row's closing sentence
rewritten in the `color` row's cadence · the Go comment corrected from "renders padding on any kind" to
keeps-and-round-trips on any kind, renders where a table consumes it.

**Deferred (1):** `DW-203` — "No `FieldSpec` may author a padding field" is enforced only by
`App.test.tsx:1432`, which checks labels for the one element it selects. A padding row offered **only for
tables** would satisfy Go's new grant and red nothing. Surfaced, not caused, by this story.

**Rejected (6), with the measurement that killed each:** negative padding on a table is **already** a
located command refusal (`component paddingTop exceeds the projection bound`), not the unlocated `Canvas`
abort a reviewer predicted — **false at the tree** · the `DataPath`-mislocation finding is `DW-202`,
already filed at the plan gate · the 5-way disjunction being a tautology is pre-existing and `ElementType`
is closed at five · the untracked spec files are an artifact of this run's version-control prohibition,
which reserves every commit to the human · the review diff omitting two `_bmad-output` files was my diff
scoping, not a defect in the change · "nothing user-facing documents the refusal" was ruled at
CHECKPOINT 1, deliberately.

**One regression the P1 fix flushed out, worth carrying forward.** The first read-back draft used
`map[string]float64` and tripped `folio-go/internal/arch_test.go`'s `TestNoFloat64UnderModule`, which
**scans `_test.go` files too** — a third failure, 1955/3. Reworked to `map[string]json.Number` compared by
canonical spelling, which is exact anyway. **This module has no float escape hatch, not even in tests.**

## Design Notes

**Why a branch split and not a narrowed condition.** The eight keys in `:1122-1126` sit behind one
disjunction only because every kind happened to want all eight. Editing the disjunction to exclude
non-table kinds would strip `background` and the three border keys from four kinds too — a silent,
unrelated regression that the suite might not catch, since Story 9.1's box keys have their own coverage
elsewhere. Two branches, two rules.

**A known pre-existing limitation, and it is NOT this story's to fix.** `propertyPath` (`:1054-1063`)
returns the **first present** key in canonical order, not the **offending** one. A mixed change set such
as `{"x": …, "paddingTop": …}` on a text element would refuse at `paddingTop` but report
`DataPath = "component.x"`. This is latent today for `width` on a table and is unreachable from the
designer — `component-property-command.ts:19` types `PropertyIntent` with a single `field`, so every
emitted `changes` object has exactly one key. **Do not fix it here** (it changes an error location for
commands the product cannot send, and touching `propertyPath` risks the DataPath this story depends on).
**It is already filed as `DW-202` in `deferred-work.md`**, logged at this plan gate on the coordinator's
instruction precisely so a reviewer meeting it at step-04 does not have to re-derive it. If it is raised
in review, the answer is "see DW-202" — it is not a finding against this story.

**The format doc's voice — RULED, and the answer went against my first draft.** I had specced the row as
a hybrid: `color`-shaped on the load path plus one `align`-shaped clause naming the designer's refusal.
**The coordinator ruled that clause OUT at CHECKPOINT 1**, and the reasoning is now the constraint:
`folio-format.md` describes THE FILE, it has never mentioned the command layer (measured: zero hits for
`updateComponentProperties`/`not editable`/`command layer`, against 15 `load error` hits), and it must not
acquire a second voice because one story found it convenient. A hand-authored `.folio` carrying padding on
a text element **is still a valid document**, and the format spec has to keep saying so. So the row takes
the `color` shape only — positive scope, inert kinds in backticks, the three-verb inertness clause, a
dated provenance line — plus a bare D-12.4.1 citation. The refusal's record lives where it belongs: the
decision log, the code comment, and the Go test.

## Verification

The heavy-test cadence for this epic is **end of epic**, so **no matrix suite and no Playwright**.
Capture each exit code on the command's own line (`cmd; rc=$?`) or with `if cmd; then … else … fi` —
`$?` is clobbered by any intervening command, `echo` included. **Never chain with `&&`**, which silently
drops everything after its first failing term.

**Baseline measured at `179a694` on a clean tree at the plan gate. If your own run disagrees with this
table, the disagreement is itself the halt.**

**Commands:**
- `cd folio-go && go test -count=1 ./...` — expected **exit 1**, **1952 pass / 2 fail / 5 skip**. The two
  are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child (`got 7, need >=20`), the one
  sanctioned permanent red. **A third failure is a real regression.** The pass count must rise by the
  tests this story adds, and by nothing else.
- `cd folio-designer && npx vitest run` — expected **exit 0**, **58 files / 806 tests**, all passing.
- `cd folio-designer && npx tsc --noEmit` — expected exit 0, `TypeScript: No errors found`.
- `cd folio-designer && npx oxlint` — expected exit 0 with **exactly 4** `only-export-components`
  warnings (`src/preview/pdf-viewer.tsx:16,17`; `src/App.tsx:2896,2903` — the line numbers move, the
  count and the rule are the invariant). **A fifth warning is a regression.**
- `cd lint && go build ./...` — expected exit 0. **Its own command, its own exit code.**
- `cd lint && go vet ./...` — expected exit 0. **Its own command, its own exit code.**
- `cd lint && test -z "$(gofmt -l .)"` — expected exit 0. **Its own command, its own exit code.**
- `cd lint && go test -count=1 ./...` — expected exit 0, four `ok` lines. `-count=1` is mandatory: this
  module walks directories, and Go's test cache does not track `ReadDir` results, so a cached `ok` is no
  measurement at all.

**Manual checks:**
- `git status --porcelain` lists **only** the files this story's tasks name (plus the workflow's own
  `_bmad-output` artifacts). **Nothing is committed, added, stashed, checked out, reset, reverted or
  restored; no branch is created; nothing is pushed.**
- The red-first proof is recorded with the mutated source line echoed back: padding on a `text` element
  **accepted before**, **refused after**.
- `git diff` touches none of: `engine-protocol.ts`, `component-property-command.ts`, `parse_bands.go`,
  `page_setup.go`, any `.folio` file, any `expected.pdf`, any golden fixture.
- `folio-format.md` line 457's first cell (`` | `padding` | ``) is byte-identical, and the diff touches no
  fenced block in that file.
- The `App.tsx` comment replacement is exactly four lines, so line numbers below 1715 do not shift.

## Suggested Review Order

**The rule itself**

- Start here: the branch split, and the comment that is now the durable in-code record of D-12.4.1.
  [`component_commands.go:1128`](../../folio-go/component_commands.go#L1128)

- What did NOT move: the four box keys stay granted on all five kinds. The split's whole point.
  [`component_commands.go:1122`](../../folio-go/component_commands.go#L1122)

- The refusal and its DataPath are reused, not rebuilt — no new error type, no new diagnostic code.
  [`component_commands.go:1150`](../../folio-go/component_commands.go#L1150)

**The record, in the two places a human reads it**

- The format row, in the format's own voice: honoured in a table's cell chrome, inert elsewhere.
  [`folio-format.md:457`](../../_bmad-output/specs/spec-folio/folio-format.md#L457)

- The panel-side comment, replaced not deleted — a rule whose only record is a comment is one refactor from gone.
  [`App.tsx:1713`](../../folio-designer/src/App.tsx#L1713)

**The tests, and the one that nearly shipped blind**

- The 16 refusals plus the 4 table grants; the table arm now reads back WHICH edge it wrote.
  [`component_properties_test.go:754`](../../folio-go/component_properties_test.go#L754)

- The read-back helper the review added; without it a rotated key-to-edge map stayed green.
  [`component_properties_test.go:680`](../../folio-go/component_properties_test.go#L680)

- The two ways to build this wrong: narrowing instead of splitting, and keying off existing style.
  [`component_properties_test.go:830`](../../folio-go/component_properties_test.go#L830)

- Subject lookup asserts the fixture's kinds, so drift fails loudly instead of going vacuous.
  [`component_properties_test.go:715`](../../folio-go/component_properties_test.go#L715)

**Peripherals**

- `clear` and `null` obey the same table-only grant; table `clear` prunes an emptied padding object.
  [`component_properties_test.go:921`](../../folio-go/component_properties_test.go#L921)

- A mixed selection is refused wholesale and moves no canonical byte.
  [`component_properties_test.go:978`](../../folio-go/component_properties_test.go#L978)

- Two deferrals banked, both with the clause that says when to re-price them.
  [`deferred-work.md:8802`](deferred-work.md#L8802)

- Tracker: the pre-ruling comment replaced, epic-12 lifted, 12.4 in flight.
  [`sprint-status.yaml:454`](sprint-status.yaml#L454)

## Delivery Log

### 2026-09-05 — done

Baseline `75e5e24`. Shipped as **one commit, `fd4da07`**, "Stop the designer authoring padding it cannot
mean" — already on `main` and already pushed. **Closed late.** No closer was dispatched when the story
shipped; it slipped between 12.1's close and 12.2's dispatch, and this entry is written after `52d0509`
(12.2), `56a903d` (12.3) and `3c4abcb` (15.2b) had all landed on top of it. **Every figure below is
labelled with the tree it was measured on.** Nothing here was measured at `fd4da07` by running anything —
no worktree was created and no revision was checked out — so the 12.4-era figures are static reads of the
commit object, and the gate figures are today's tree at `68e21ac`.

**What shipped, and what it means — the short version for anyone arriving from a D-12.4.1 citation.**
D-12.4.1 is **not** "padding is forbidden". It is an asymmetry with two halves, and both halves shipped:

- **The engine keeps honouring what it is given.** The loader, the serializer, the canvas projection and
  the table's cell chrome were **not touched** — `git show fd4da07 --stat` names eight files and none of
  them is on a load, render or projection path. A hand-authored document carrying `style.padding` on a
  `text` element still parses, still round-trips with the padding intact, still projects `paddingTop`, and
  still accepts unrelated property commands on that same element.
- **The designer stops authoring what it cannot mean.** The command layer's all-kinds `allowed` branch was
  **split in two**: `background`/`borderWidth`/`borderColor`/`borderEdges` stay granted on all five kinds,
  and the four padding keys moved to a table-only branch. **Padding on an `ElementTable` is still an
  accepted, applied command** — that is the point of the split, and it is what
  `TestPaddingPropertyCommandsAreGrantedOnATableAlone`'s table arm asserts. The refusal covers the four
  non-table kinds only.

So a later story that reads D-12.4.1 as barring padding everywhere is reading it too widely; what the
ruling bars everywhere is a **panel control**, and what it bars in the command layer is padding **off a
table**. The ruling explicitly rejected the framing that "the command must not be stricter than the
loader" — that is **not this project's rule**. The engine and the designer are allowed to disagree, and
here they deliberately do.

**Decisions applied.** `D-12.4.1` (owner, 2026-09-05, `epic-11-14-decision-log.md`) — executed in full,
and now cited by id in three places rather than by an undated comment: the Go branch's rationale comment,
the designer's `borderFields` comment, and the format's `padding` row. The epic's AC1 two-way fork
(padding-as-inset vs. table-only) is **dead**; the table-only arm is the one that shipped. `DW-202` and
`DW-203` were both **filed by this story** rather than fixed by it.

**12.4-era figures — static reads of the commit object, not runs.** `git show fd4da07 --stat`: **8 files,
+939 / −9**. Production surface is small and the test surface is not: `component_commands.go` **+13/−1**
(the branch split and its seven-line rationale comment), `App.tsx` **+3/−3** (the comment replacement,
with `git show 75e5e24:…App.tsx | wc -l` and `git show fd4da07:…App.tsx | wc -l` both **3157**, so the
"exactly four lines, no line-number shift" constraint held and the ~13 line-number citations in
`App.test.tsx` did not rot), `folio-format.md` **one line changed** with its first cell byte-identical
(`cut -d'|' -f2 | od -c` gives the identical 14-byte dump either side of the commit),
`component_properties_test.go` **+342** carrying **four** new top-level tests
(`…AreGrantedOnATableAlone`, `…LeavesTheBoxKeysAndAnAlreadyLoadedPaddingAlone`,
`…CommandOpsFollowTheSameTableOnlyGrant`, `…RefusalAcrossAMixedSelectionIsTransactional`). The remaining
four files are `_bmad-output` records.

**Review, and the finding that actually mattered.** One round, three layers (blind-hunter, edge-case,
verification-gap), no loopback — `review_loop_iteration` stayed **0**. **8 patched / 1 deferred / 6
rejected**, per the Spec Change Log above; those counts are the build's own and are **transcribed here,
not re-derived** — this close did not re-open triage. The finding worth remembering is the one the
reviewer *demonstrated*: the table arm originally asserted only "accepted, and some byte moved", so
rotating the key-to-edge map left the entire suite **green** — a table commanded to a 5pt top inset would
have taken a 5pt bottom inset and shipped clean. The patch reads the edge back by name and also asserts
the fixture's *other* authored edges are undisturbed, which is the half that catches a rotation rather
than a drop. A first draft of that read-back used `map[string]float64` and tripped
`internal/arch_test.go`'s `TestNoFloat64UnderModule`, which scans `_test.go` files too; it was reworked to
`json.Number`. **This module has no float escape hatch, not even in tests.**

**Gates — measured today, 2026-09-05, on `68e21ac`, a clean tree. These are NOT 12.4's gates.** Three
stories have landed since `fd4da07`, so these numbers describe the tree 12.4 lives in now, not the tree it
shipped into; the build's own contemporaneous claim was 1956 Go passes and 806 designer tests, and that
claim is left as the build's, unverified by this close.

- `folio-go`: `go test -count=1 ./...` → **rc 1**, **2105 pass / 2 fail / 5 skip**, **14 packages ok**,
  1 package failing. The two failures are the single sanctioned permanent red —
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child. No third failure. Counted from
  `-json` events, since `go test ./...` prints no pass/skip totals.
- `folio-go`: `gofmt -l .` → **empty**. `go vet ./...` → **rc 0**.
- `folio-designer`: `npx vitest run` → **rc 0**, **62 files / 893 tests**, all passing.
- `folio-designer`: `npx tsc -b --force` → **rc 0**. (`--force` deliberately: `tsc -b` alone can exit 0
  off the incremental cache without typechecking, and `npx tsc --noEmit` typechecks zero files in this
  project.)
- `folio-designer`: `npx oxlint` → **rc 0**, **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:3059,3066` — the line numbers moved since the spec was
  written, the count and the rule did not).
- `lint`: `go build ./...` **rc 0** · `go vet ./...` **rc 0** · `test -z "$(gofmt -l .)"` **rc 0** ·
  `go test -count=1 ./...` **rc 0**, four `ok` lines.

**Suites NOT run, and when they come due.** The matrix suite and Playwright were out of scope by this
epic's heavy-test cadence, which is **end of epic**. They remain owed at the **Epic 12 boundary gate**,
which cannot be reached yet: 12.5 is still `backlog`. They were not run at `fd4da07` either, so no run of
them has ever covered this change.

**Deferred, both still OPEN and both unassigned.**

- **`DW-202`** — `propertyPath` locates a refusal at the first key present in canonical order, not at the
  key actually refused, so a mixed change set reports the wrong `DataPath`. Latent and unreachable from
  the product: `component-property-command.ts`'s `PropertyIntent` carries a single `field`, so every
  emitted change set has exactly one key. **Re-checked at close: still single-field at `68e21ac`, so the
  deferral's unreachability premise still holds.**
- **`DW-203`** — "No `FieldSpec` may author a padding field" is enforced only by a behavioural test that
  checks labels for the one element it selects. A padding row offered **only when a table is selected**
  would satisfy Go's new grant and red nothing. **Sharpened by this story, not caused by it** — before
  the split, any padding row would have "worked". **Re-checked at close and it has become less
  hypothetical:** `epics.md`'s Story 14.8 section still carries an unstruck acceptance criterion
  instructing exactly that control, underneath a newer owner preamble that rules it out. Flagged to the
  orchestrator; not this story's to reconcile.

**Housekeeping done at close.** The plain-terms opener did not exist and was **written**, not rewritten.
The tracker key moved `review` → `done`; `epic-12` stays `in-progress` because 12.5 is `backlog`. The
frozen block was verified byte-identical across these edits by sha256 either side
(`0631cfbc54e1662397161eefb1815aed26f0c21fdcb18d9c984c73c756b9865f`). Nothing was committed — every
version-control verb is reserved to the human on this run, so the two files this close touched are left
dirty in the working tree.
