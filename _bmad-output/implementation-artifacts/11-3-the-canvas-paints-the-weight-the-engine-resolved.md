---
title: 'The canvas paints the weight the engine resolved'
type: 'feature'
created: '2026-09-06'
status: 'done'
baseline_commit: 'fa931076369e17b6f7feab7dedebbbe740b207da'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Not normative, and rewritten after the fact. The frozen Intent below governs implementation.*

The product forbids faking a bold typeface by smearing a normal one thicker. **The design canvas has been
doing exactly that, and since the previous story it has been doing it twice over.**

Story 11.2 taught the engine to pick a real bold cut and to tell the browser which one it picked. The
browser has been ignoring that and applying its own thickening instruction, computed from the checkbox
rather than from the answer. Once the engine started naming a genuinely bold typeface, the browser began
thickening **that** — so a bold heading on the canvas has been drawn in the bold cut and then emboldened
again on top of it. Nothing was violated when it shipped: 11.2's acceptance never reached the canvas.
But the honest description is that the previous story **made the canvas worse in exchange for making the
engine right**, and nobody noticed until this story traced the path the paint actually takes.

This story deletes the thickening, lets the real cut through untouched, teaches the panel to read a
document's declared cuts back out so it can say when a typeface simply has no bold, and fills in the
starter document so a brand-new file can bold at all.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The canvas paints synthetic bold. `App.tsx:3394` sets `--text-font-weight` from
`component.bold` — **the requested flag** — and `App.css:152` feeds it to `font-weight`, with
`--text-font-style` / `font-style` the same story. That is browser emboldening and obliquing, forbidden by
I-2 and AD-17 alike. Since Story 11.2 it is also actively wrong: `CanvasTextFragment.Face`
(`page_setup.go:1649`) already carries the **resolved** face, and the generated `@font-face` rules declare
no `font-weight` descriptor (asserted at `font-catalogue.test.ts:536`), so a real `Roboto Bold` fragment is
today **synthetically emboldened on top of itself**.

Two read-back holes ride the same surface. **DW-239:** `CanvasFontChainEntry` (`page_setup.go:593`)
projects only `face`/`assetKey`/`family`/`style`, so the panel cannot see that a chain declares a bold cut,
and cannot honour the epic's rule that an absent cut is *stated* rather than shown as a reachable on-state.
**DW-240:** `TableColumnsProjection` carries the committed/resolved pair for seven header-style fields while
`tableHeaderStyleFields` is nine — a header weight the designer can author and cannot read back. And no
document in the product declares a variant at all: `starter.folio` is three bare regular faces, so a new
file cannot bold (D-11.0.1).

**Approach:** Delete the synthetic CSS and let the engine's resolved face reach the browser unmodified,
with a contract test naming the forbidden CSS properties the way the canvas-authority contract already
names the banned measurement APIs. Project each entry's declared variants onto the wire, and derive the
B / I controls' third state from that projection rather than from the requested flag. Give
`TableColumnsProjection` the missing pair and tie its member list to `tableHeaderStyleFields`. Write
`starter.folio`'s variants first, so the story has a document to demonstrate against.

## Boundaries & Constraints

**Always:**
- **The canvas paints the RESOLVED OUTCOME, never the requested flag.** `component.bold` and
  `component.italic` may drive the panel's control state and nothing that paints. The face is
  `fragment.face` / `fragment.assetKey`, which the engine already sets.
- **No synthetic bold or oblique, anywhere (I-2).** Never `font-weight` or `font-style` on painted
  document text, in a stylesheet rule or an inline style, directly or through a custom property. This is
  the sharp end of the rule the emit path already obeys.
- **AD-17 is untouched.** This story adds no browser measurement. Attribution is not measurement.
- **The engine already resolves; do NOT write a second resolver.** `fontChain` (`render.go:1194`) returns
  `(base, styled)` and `addCanvasTextPaint` (`page_setup.go:1347`) already passes `styledChain` into
  `shapeSegments` at `:1377`. If you find yourself resolving a weight in TypeScript or in a second Go
  site, stop and ask.
- **Resolution is DECLARED, never constructed, parsed, or inferred from binaries (D-11.2.1 / D-11.2.2).**
  The projection copies `entry.Bold` / `entry.Italic` / `entry.BoldItalic` verbatim. No `+ " Bold"`, no
  name parsing, no reading a name table, no sniffing `OS/2`.
- **A sibling's namespace matches its entry's discriminant (AD-8).** A `face` entry's variants are FontSet
  face names; an `asset` entry's variants are asset keys. The projection must not cross them, and the
  browser guard must not admit a crossing.
- **`CanvasFontChainEntry`'s new keys are ALWAYS PRESENT — no `omitempty`.** The browser checks the object
  with `hasExactKeys` (`engine-protocol.ts:334`); a key that appears only for some entries **rejects the
  whole snapshot for those documents, and the symptom is a blank canvas.** The struct says so at
  `page_setup.go:590`. Absence is the zero value, as it already is for `family`/`style`.
- **`starter.folio` must declare no variant it cannot honour.** Under **D-11.2.11** a variant naming its
  own base face is a load error once 11.4 lands, so never write `{"face":"X","bold":"X"}` as a
  placeholder. **Absence is the correct declaration** for a cut that does not exist, and AC3's engine
  Warning is the correct outcome. Noto Sans Thai has **Bold only** (no upstream italic); Noto Sans SC has
  **no cut at all** and must declare none.
- **`starter.folio` is written FIRST, not as a closing tidy-up (D-11.2.12).** The story's acceptance needs
  a document that declares variants.
- **The B / I absence statement follows DESIGN.md:** anything declined or disabled carries its reason
  beside it, in the product's terse, technical, located voice — never a bare grey-out.
- **Never pin a predicate against itself, and never ship an assertion whose two sides could be equal
  (D-11.2.8).** Named traps are in Design Notes; both are live here.

**Ask First:**
- Running `npm run build` (the offline release build). `starter.folio` is fingerprinted into the release
  manifest at `build-wasm.mjs:99`, so this story plausibly needs one — **ask before running it** so two
  builds do not race into one `dist` (DW-100).
- Any change to the **font-chain command surface** — `setFontChain`, `addFontChain`, `embedFontFamily`,
  `font-chain-command.ts`. That is **Story 11.4 (DW-238)** and is out of scope here.
- Any change that moves a **golden digest** or a rendered byte. This story runs **no matrix suite**
  (D-11.2.12); if you believe you need one, the seam was drawn wrong — say so rather than running it.
- Any change to `fonts.Shipped()`, the shipped-face count, or the `@font-face` generator's face list. That
  is Story 11.1's guard table and a move means this story has drifted.
- Widening `CanvasTextFragment`, `CanvasTextPaint`, or the canvas fragment fallback stack
  (`shipped-face-family.ts:63` and `App.css:154`, tied to each other by a guard).
- Anything outside this spec's Tasks: a file the tasks do not name, an adjacent bug, a refactor, a guard
  for a case no acceptance criterion demonstrates, or a new dependency.

**Never:**
- **Never commit, `git add`, stash, checkout, reset, revert, or restore. The orchestrator makes every
  commit.** Leave all work in the working tree and report what to commit. Reading git state is fine and
  encouraged. One carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` dir under `os.tmpdir()`,
  removed in `finally`.
- **Never create a branch. Never push. Never open a pull request.**
- **Never `git add -A`.**
- **Never edit `worked-example.json`, and never give it an object-form chain entry.** Three tests bind it:
  `goldenfixture_test.go:16` byte-compares it to the `## Worked example` fence (heading
  `folio-format.md:846`, body **851-957** — the dispatch's `:876` lands mid-table),
  `goldenfixture_test.go:198-200` re-asserts the same equality as a control, and
  `browser_roundtrip_witness_test.go:196-206` requires it to be canonical serializer output. Worse:
  **`diagnostic_registry_census_test.go:208-209` is an explicit tripwire** — *"fixture precondition:
  worked-example.json's fonts block now carries an object-form entry, so the absence arm may not be
  reached"*. It is the only production witness for `DiagCodeTextStyleFaceUndeclared`; an object-form entry
  there destroys it.
- Never add a golden fixture or an `expected.pdf`. **That is Story 11.5**, split out by D-11.2.12, and it
  carries an owner attestation this story does not.
- Never implement D-11.2.11's self-referential-variant load error. **That is Story 11.4's** parse check.
- Never build a face-name → CSS-family mapping table. `shipped-face-family.ts` is shape-based by ruling
  (D-8.4.14) and already admits `Roboto Bold`.
- Never treat a green e2e run as evidence. **DW-192/DW-193: `test:e2e:compile` is a TYPECHECK, and the
  suite is exercised at epic boundaries — not in CI and not per story.** (Corrected by the ORCHESTRATOR
  under D-000.30, 2026-09-06, after approval: the frozen text said "compiled and never executed", which
  is DW-193's wording with four load-bearing words dropped. CI has indeed never run it — 42 cases across
  16 files — but the suite is *not* unexecuted: `0c0f3e9` is "Fix the two e2e failures Epic 12's boundary
  gate found, one of them a regression". Only a human may reopen a frozen block, and the builder correctly
  declined to; this is that reopening, recorded rather than silent.) The `App.tsx` pointer carve-out
  waives every AD-17 prohibition inside that file. Say so in the report rather than letting a green read
  as coverage.
- Never rewrite an existing document on open (I-5). `starter.folio` is a shipped template we author
  deliberately; do not generalise from it to a migration.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Bold element, variant declared | Chain entry `{"face":"Roboto","bold":"Roboto Bold"}`; element declares bold | Fragment carries `face: "Roboto Bold"`; browser asks for `'Roboto Bold', …`; **no `font-weight` anywhere** | N/A |
| Bold element, no variant declared | Bare `"Roboto"`; element declares bold | Fragment carries `face: "Roboto"`; painted in Regular. Canvas and PDF agree — both took the engine's answer | Engine Warning (already shipped); canvas discards diags |
| No style declared | Element declares neither | Byte-identical projection to today apart from the three new always-present keys | N/A |
| Chain declares a bold somewhere | `["Roboto"(+bold), "Noto Sans SC"(none)]` | B control is **available**. Bolding works for Latin; the CJK shortfall is the engine's per-rune Warning, not a panel state | N/A |
| **No entry declares a bold** | Chain is `["Noto Sans SC"]` | B control shows the **absent** state and states the reason beside it. It never renders as plainly on | N/A |
| Absent cut, document already declares it | Family switched to a cut-less chain after bolding | Control shows the absent state **and remains operable so the author can clear it** — no state that can be entered and not left | N/A |
| Italic on Thai | `"Noto Sans Thai"` (Bold, no italic) | I control shows the absent state; B does not | N/A |
| Embedded entry with a variant | `{"asset":"<k1>","bold":"<k2>"}` | Projection carries `bold: "<k2>"` — an **asset key**, not a face name | N/A |
| Table header weight | `headerStyle.bold` declared | `TableColumnsProjection` carries `headerBold` **and** `headerBoldResolved` from `resolveHeaderStyle` | N/A |
| Table header weight absent | `headerStyle.bold` absent, `style.bold` set | `headerBold` false (absent); `headerBoldResolved` true (cascaded) | N/A |
| New key on one side only | A variant member added to Go and not to `engine-protocol.ts` | `hasExactKeys` rejects the snapshot → **blank canvas**. Both sides and `canvasFontChainEntryWireKeys` move in one commit | Wire test reds |
| Starter opened as a new document | `starter.folio` at `2.0` with object-form entries | Loads; a text element can be bolded and paints in Roboto Bold | N/A |

</frozen-after-approval>

## Code Map

Every anchor **measured at `fa93107`** with a clean tree (D-000.4). Implementation dispatches from
**`42f0d44`**, which adds only `epic-11-14-decision-log.md` — **zero code delta**, verified with
`git diff --name-only fa93107..42f0d44 | grep -v '^_bmad-output/'` (empty), so every anchor below holds at
either commit. `baseline_commit` stays `fa93107` per the workflow's preserve rule.

⚠ **Story 11.2's Code Map was measured at `102e1fc`, BEFORE its own implementation commit `d0ded7e`** — so
every Go anchor in it moved by the time it closed. That is structural, not carelessness: a Code Map is
measured at the plan gate and the story then edits the files it anchored (**D-11.3.4**). It is why D-000.4
says *cite by symbol* — the symbol is the only thing that survives the story's own diff. **Re-derive every
line below before editing it**, including the ones in this map.

### What is ALREADY DONE, and must not be re-done

This is the story's most important finding: **AC1's engine half needs no code.**

- `folio-go/render.go:1194` `fontChain(doc, el) (base, styled []string, err error)` — returns both lists.
- `folio-go/render.go:2000-2030` `shapeSegments` — `face = chain[index]`, then `face = variant` when
  `styled[index]` is non-empty **and covers the rune**. `faceSegment{face: seg.face}` at `:2145`.
- `folio-go/page_setup.go:1347` `addCanvasTextPaint` takes `chain, styledChain`; passes **both** to
  `shapeSegments` at `:1377`; emits `CanvasTextFragment{… Face: shipped}` at `:1649`.
- `folio-designer/src/shipped-face-family.ts:65` — the family rule is a **shape regex**
  (`/^[A-Za-z][A-Za-z0-9]*(?:[ -][A-Za-z0-9]+)*$/`), not a list, so `"Roboto Bold"` already yields
  `'Roboto Bold', 'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC', sans-serif`. **No list to extend.**
- Story 11.1 already generates an `@font-face` per cut (`build-wasm.mjs`), forced by the mirror guard
  `familiesWithNoRule` in `font-binary-identity.test.ts`.

**So the resolved bold face is on the wire and resolvable by the browser today.** The only thing standing
between it and the screen is the synthetic CSS below.

### The synthetic weight — the whole of AC1

- `folio-designer/src/App.tsx:3394` — one very long line. It sets
  `'--text-font-weight': component.bold ? 700 : 400` and
  `'--text-font-style': component.italic ? 'italic' : 'normal'` on `.canvas-text-paint`, **from the
  requested flag**. The same line already sets `fontFamily` correctly from `fragment.assetKey` /
  `fragment.face`. Delete the two weight/style properties; leave the family logic alone.
- `folio-designer/src/App.css:152` — `.canvas-text-paint { … font-weight: var(--text-font-weight);
  font-style: var(--text-font-style); }`. Delete both declarations.
- ⚠ **`--text-font-size` on the same rule stays.** It is a size, not a weight, and the engine's own.

### AC2's contract test — copy the mechanism that exists

- `folio-designer/src/canvas-authority-contract.test.ts:18` `prohibited` — the regex list.
  `:132` `withoutComments` is a **quote-aware character scanner** (it does strip comments; the common
  claim that it does not is false). `:195` the corpus floors: production > 10, tests > 10, e2e > 3.
- **The CSS precedent is already in the list**, at the `white-space|text-wrap|…` row: it bans
  **property:value pairs**, scoping by VALUE.
- ⚠ **VALUE SCOPING DOES NOT WORK FOR THIS ONE.** The chrome legitimately writes the same values —
  `font-weight: 500` at `App.css:35`, `:42`, `:199`, `:354`, `:434`, `:528`, `:542`;
  `font-weight: 600` at `:350`, `:407`; `font-style: italic` at `:333` (`.property-fx`). A value-scoped
  ban reds nine legitimate chrome rules. **The prohibition must be scoped to the painted-document
  surface**, not to the value. See Design Notes.
- ⚠ **THE PROHIBITION MUST REACH TWO DIFFERENT SPELLINGS, AND A LINE-46-SHAPED PATTERN REACHES
  NEITHER.** The value today is not a literal: `App.css:152` reads `font-weight: var(--text-font-weight)`
  and `App.tsx:3394` writes `'--text-font-weight': component.bold ? 700 : 400`. A pattern matching
  `font-weight:\s*(bold|\d+)` matches **neither of the two lines this story deletes**, so it would ship
  green over the defect. Write the pattern against both spellings and give **each its own
  `violationsForSource` row**.
- ⚠ Two more false-positive sources to survive: `App.test.tsx:3254` and `:3274` carry
  `font-weight:700` inside **pasted-HTML test fixtures** (input data, not a rule), and
  `font-catalogue.test.ts:41-42,524,536-537` legitimately names both properties while asserting their
  **absence** from `@font-face` descriptors.
- **Both directions are mandatory**, matching `:208` `turns a TypeScript copy of any engine refusal red`
  and `:228` `does not catch ordinary English in a comment`. The red-proof block for CSS-shaped rules is
  `:243` `turns realistic measurement, CSS, range, and event-coordinate mutations red` — the new rows
  belong there.
- ⚠ **The two custom properties are asserted by NOTHING today**, so deleting them reds nothing and the
  deletion is invisible to the suite. Measured: `--text-font-weight` / `--text-font-style` are set only
  at `App.tsx:3394` and read only at `App.css:152`; **positive control** — the sibling
  `--text-line-baseline` IS asserted at `App.test.tsx:1433` and `--text-ink` at `:2293`, same file, same
  `toHaveStyle` idiom. So this is a genuine coverage hole, not a search artefact, and the deletion needs
  a new **positive** assertion of its own — that a bold component's painted node carries no weight or
  slope — not merely the contract scan.

### DW-239 — the chain-entry projection, and the blank-canvas hazard

Three sites move in ONE commit; two of the three are already tied by a test that will red if you miss one.

- `folio-go/page_setup.go:593` `CanvasFontChainEntry{Face, AssetKey, Family, Style}` — **no `omitempty` on
  any field, deliberately.** `:590` states why: `hasExactKeys` blanks the canvas for documents missing a
  key. Add `Bold`, `Italic`, `BoldItalic` the same way.
- `folio-go/page_setup.go:651` `projectFontChainEntry` — fills the entry and applies
  `maxCanvasPropertyString` to **every** string ("a bound applied to three of four fields is a bound on
  nothing"). Read the variants with `entry.Variant(s)` (`model.go:293`) or the fields directly
  (`model.go` `FontChainEntry.Bold/Italic/BoldItalic`, plain strings, `""` = absent). Add the three to the
  bound loop.
- `folio-go/canvas_projection_wire_test.go:98`
  `canvasFontChainEntryWireKeys = []string{"assetKey","face","family","style"}` — the recorded protocol
  set. `:200` ties it to the **Go struct's marshalled keys**; `:414` ties it to **the designer's guard**,
  and its failure message says a one-sided field "blanks the canvas exactly as a chain-level or a
  top-level one would". Sorted order.
- `folio-designer/src/engine-protocol.ts:333` `isFontChainEntry` — `hasExactKeys(value, ['face',
  'assetKey', 'family', 'style'])`, then per-field `typeof` and `MAX_CANVAS_PROPERTY_STRING`, then the
  discriminant rule `(face.length > 0) === (assetKey.length > 0)` → false, then
  `assetKey.length > 0 ? family.length > 0 : family.length === 0 && style.length === 0`.
- `folio-designer/src/engine-protocol.ts:285` — the TS **type**:
  `entries: ReadonlyArray<Readonly<{ face: string; assetKey: string; family: string; style: string }>>`.
  A second place to add the three members.
- ⚠ Do **not** shape-check a variant asset key as 64-hex here. `:325-328` records the ruling: a
  64-character face name is a legal face name, so "looks like a digest" was never available as a test.

### AC3 — the B / I third state

- `folio-designer/src/App.tsx:2692` `BooleanProperty` — the generic control. Renders
  `label.slice(0, 1)` (so "B" / "I"), `aria-pressed={active}`, `disabled={pending}`, and toggles between
  `{operation:'clear'}` and `{operation:'set', value:true}`. It receives `components`, `ids`, `field`,
  `onCommit`, `documentGeneration`, `error` — **and no chain**. It needs the absence fact passed in.
- `folio-designer/src/App.tsx:2058` — the call site inside `PropertySection title="TYPOGRAPHY"`, where
  `fontChains` is already in scope (it is handed to `FontFamilyProperty` on the same line).
- ⚠ `folio-designer/src/App.tsx:2795` `declaredChainEntry(name, chains)` returns
  **`chain.entries[0]` — the FIRST entry only.** It is the obvious reuse point and it is the **wrong**
  rule for this question. See Design Notes; this is a vacuity trap as well as a correctness one.
- `folio-designer/src/App.css:350` `.property-toggle[aria-pressed="true"]` — the existing on-state. The
  absent state needs its own rule; **check specificity and source order by hand** (jsdom applies no
  stylesheet, so CSS correctness here has zero executable coverage).
- DESIGN.md's rule: disabled drops to `0.42` opacity **and must carry a stated reason, never a bare
  grey-out**; "State the reason next to anything disabled."
- The mockup (`Main.dc.html:296-313`) shows only the two-state pair — **there is no third state drawn**,
  so it is built from the DESIGN.md rule, not copied. (The mockup's own `font-weight: 600` on its "B"
  glyph is chrome, and is not a precedent for the page.)

### DW-240 — `TableColumnsProjection`

- `folio-go/table_columns_projection.go:17-80` `TableColumnsProjection` — its comment at `:22` says
  **"sixteen members, and the arithmetic is 7 x 2 + 1 + 1"**. Seven `Header*`/`Header*Resolved` pairs at
  `:64-77` (`FontFamily`, `FontSize`, `LineSpacing`, `Background`, `Color`, `Valign`, `Align`) plus
  `HeaderHeight` (`:61`) and `AltRowBackground` (`:62`) singletons. `grep -n 'Bold\|Italic'` on the file
  → **0**; positive control `grep -n 'json:'` → **30**. Adding the two pairs makes it
  **9 x 2 + 1 + 1 = 20**; the arithmetic sentence must move with it.
- ⚠ **DW-240 has FIVE mirror sites, not two.** The register named the struct; these were measured here:
  `folio-go/canvas_projection_wire_test.go:474` `tableColumnsProjectionWireKeys` (asserted `:574-580`,
  cross-checked against the TS guard at `:598-600`); `folio-designer/src/engine-protocol.ts:199-206`
  `type TableHeaderStyle` (fourteen members, seven pairs); `engine-protocol.ts:378` `isTableColumns`'s
  `hasExactKeys` list, with `:410` and `:417`; and
  **`folio-designer/src/table-style-command.ts:40`** `TableHeaderStyleField` — still a **seven**-member
  union, whose comment at `:34-39` is now **stale and false** (*"`bold`/`italic` have no arm in the
  engine's header cascade to resolve from"* — 11.2 gave them one). Retire that comment by editing it, the
  way 11.2 retired its twin at `component_commands.go:2510`.
- ✅ **`TableEditor.tsx` needs no new control.** `grep -c -i 'bold\|italic'` → **0**; positive control
  `grep -c 'headerAlign\|headerValign\|headerFontFamily'` → **4**. DW-240 is read-back plumbing only;
  do **not** add a header B/I control here — that is not in this story's ACs.
- The resolved half calls **`resolveHeaderStyle`** (`table_render.go:321`) — the one cascade, unexported,
  same package. **Nothing in TypeScript may re-derive it.** 11.2 already added its bold and italic arms.
- `folio-go/component_commands.go` `tableHeaderStyleFields = []string{"fontFamily","fontSize",
  "lineSpacing","background","color","valign","align","bold","italic"}` — **nine**. The authoring half.
- ⚠ **The durable half of DW-240 is the TIE, not the two members.** Nothing today relates
  `tableHeaderStyleFields` to the projection's member list, which is exactly why no test caught the
  asymmetry. Add that tie or the tenth field repeats this. **Its home is
  `folio-go/table_header_style_test.go:705-775`** `TestEveryResolvedProjectionMemberFollowsBothLegsOfThe
  Cascade`, whose `rows` table at `:711-728` has **exactly seven rows** — a hand-maintained list that is
  itself the second copy of the asymmetry. `:302` already iterates `tableHeaderStyleFields`, but only for
  *reachability*, which is why it did not catch this.
- Absence convention: the struct states **"ABSENCE IS SPELLED AS THE ZERO VALUE"**, forced by the exact
  key set. For a bool that collapses committed-absent with committed-`false`; see Design Notes.

### `starter.folio`, and its consumers — measured, because the dispatch's list was short

`folio-designer/public/templates/starter.folio`, 20 lines. Chain at `:9`
(`"Roboto": ["Roboto", "Noto Sans Thai", "Noto Sans SC"]`), `"version": "1.0"` at `:19`, zero elements.

`git grep -n 'starter\.folio'` over the whole tree (rc=0) returns these **non-`_bmad-output`** consumers:

- `folio-designer/scripts/build-wasm.mjs:74-75` writes it into the release output, **:99
  `starter: fingerprint(starterPath, 'starter.folio')`** — its bytes feed the release manifest. ⚠ This is
  why a release build may be needed and is **not** in the dispatch's list.
- `folio-designer/e2e/font-embed-boundary.spec.ts:75` reads the real file and, at `:362`, reasons about
  *"starter.folio's declared chains"*. Its `:65-66` comment says it derives from the file rather than
  from a premise about its content — **verify that survives an object-form entry**, including whether it
  types the entries as `string[]`. Compile-only, but `test:e2e:compile` must stay green.
- `folio-designer/src/startup-sequence.test.ts:11,15` — the **path** only, not the content.
- `folio-go/fonts/roboto/NOTICE.md:28` — prose quoting the starter's chain verbatim. Licence-accounted;
  check whether the census reads it.
- `folio-go/fonts/fonts.go:141` — a comment mentioning the starter.

**Nothing pins the starter's bytes or its version string — measured here, with positive controls:**

- The file's sha256 appears **nowhere** in the tree (population: whole repo minus `node_modules`/`.git`).
- The build fingerprints the `"\n  "`-prefixed copy into
  `folio-designer/src/generated/offline-assets.ts`, which is **gitignored** (`.gitignore:99`) and
  regenerated by `build:wasm`, so a starter edit re-derives it automatically. **Positive control:**
  `git ls-files folio-designer/src/generated/` returns exactly one tracked file, `pdfjs-assets.ts` — a
  tracked generated file *would* have shown up.
- `verify-offline-release.mjs:113` checks only that some asset **ends with** `.folio` — a class check, not
  a digest. **Positive control that this script does pin things:** it and
  `offline-release-contract.mjs:65-70` line-anchor-read `^const maximumCacheAssets = (\d+)$` out of
  `release-payload.ts` and fail the build on drift. The mechanism exists and was deliberately not applied
  here.
- No snapshot test exists for it. `startup-sequence.test.ts:9` deliberately stubs the fetch with
  `new Uint8Array([1,2,3])` and never reads the real file.
- Nothing on the designer side asserts a **document** version string at all: every `"version"` hit under
  `src`/`e2e`/`scripts` is the command envelope `"version":1`, and `'2.0'` appears **0** times.

**So a release build is probably NOT required** — it is still Ask First, but say this measurement when
you ask rather than asking blind.

**`2.0` is accepted, and the bump is mandatory.** `internal/template/version.go:140` `checkVersionLoadable`
refuses only `major > SupportedMajor`; `SupportedMajor` is 2 and `SupportedVersion` is `"2.0"` at
`:76-79`, and **they do not move.** `versionRequiredByContent` → `fontsRequireMajor` (`:382-391`) raises to
`2.0` whenever any entry `SerialisesAsObject()`, so the engine will stamp `2.0` on the first save; leaving
`1.0` in the committed file ships the "version that lies" this run already caught once.
`folio-go/wasm/cmd/engine/main_test.go:229-238` asserts **only `loaded.OK`** — no content, no version, no
digest — and `2.0` passes it.

⚠ **`{"face":"Noto Sans SC"}` with no siblings canonicalises straight back to the bare string
`"Noto Sans SC"`.** `SerialisesAsObject()` (`model.go:272`) is false for a variant-free object, and
`font_chain_variants_test.go:249-272` pins exactly that ("if it serialised as an object it would raise
itself to 2.0 on every save"). **Write SC as a bare string**, or accept a first-save diff.
Serializer key order is alphabetical: `bold`, `boldItalic`, `face`, `italic`.

**The exact face names** (`folio-go/fonts/fonts.go:150-164`, eleven keys):

| entry | `face` | `bold` | `italic` | `boldItalic` |
|---|---|---|---|---|
| Roboto | `Roboto` | `Roboto Bold` | `Roboto Italic` | `Roboto Bold Italic` |
| Noto Sans Thai | `Noto Sans Thai` | `Noto Sans Thai Bold` | **none** | **none** |
| Noto Sans SC | bare string | **none** | **none** | **none** |

Both absences are rulings, stated in `fonts.go` itself: SC at `:102-107` (D-A — its Regular alone is
10.6 MB, so cuts were declined and *"a family with no face at a requested weight is a permanent, shipped
condition"*), Thai italic at `:121-124` (*"upstream publishes none. Seven is the whole realizable set"*).

### Guards that will fire, and one that must not

- `canvas_projection_wire_test.go:200` and `:414` — will red until all three chain-entry sites agree.
  **This is the story's safety net; do not weaken it.**
- `canvas-authority-contract.test.ts` corpus floors (production > 10, tests > 10, e2e > 3) — adding files
  is free; deleting is not.
- `canvas-font-stack.test.ts` — a contract/census test with **no production module beside it** (there is
  no `canvas-font-stack.ts`; `find` returns only the `.test.ts` — positive control: `shipped-face-family*`
  returns both). It reads `build-wasm.mjs`, `App.css`, `fonts.go`, `tokens.css` and `font-catalogue.json`
  as text and ties them together. **It must stay green untouched.** Three rows bind this story:
  - ⚠ **`:288` `shippedFaceDerivedFamily = /^shippedFaceFamily\(fragment\.face\)$/` is anchored to the
    identifier `fragment` BY NAME**, and `:282-287` says why: *"A chain ENTRY has a `face` too, so
    `shippedFaceFamily(entry.face)` is a per-COMPONENT, chain-entry-derived family — the exact evasion
    this census exists to catch."* The B / I work reads chain entries; **never put one in a font-family
    position.**
  - `:1043` permits only a closed set of approved family derivations in a font-family position **in every
    designer source**; `:1070` proves it reds on a document-supplied family.
  - `:796` `toBe(11)` and `:819` `toHaveLength(10)` are 11.1's face counts and must not move.
- `font-catalogue.test.ts:524-537` — `@font-face` rules carry no `font-weight`/`font-style` **descriptor**.
  Green and unrelated, but it is the nearest neighbour to AC2 and must not be confused with it: a
  descriptor on a rule is a different thing from a property on painted text.
- The 11.1 shipped-face guard table (`shipped_faces_ext_test.go:72`, `shipped_faces_test.go:525`,
  `fontgen_matrix_test.go:264` `wantDerivedShippedFaces = 7`, the designer `toBe(11)` / `toBe(13)` counts)
  — **all must stay green untouched.** A move means the story drifted into 11.1.

## Tasks & Acceptance

**Execution:** (in this order — `starter.folio` first, by D-11.2.12)

- [x] `folio-designer/public/templates/starter.folio` -- Rewrite `:9`'s chain per the Code Map's face
      table: Roboto as an object with all three variants, Noto Sans Thai as an object with `bold` only,
      **Noto Sans SC as a BARE STRING** (a variant-free object canonicalises back to one on the first
      save). Move `:19` to `"2.0"` — mandatory, because `fontsRequireMajor` raises on any object-form
      entry and leaving `1.0` ships a version that lies. **Declare no cut that does not exist**, and
      **never name a family's own base face as its variant** (D-11.2.11). Re-confirm the Code Map's
      unpinned-bytes measurement yourself before editing; report the search, not the conclusion.
- [x] `folio-go/page_setup.go` -- Add `Bold`, `Italic`, `BoldItalic` to `CanvasFontChainEntry` with
      **always-present** JSON keys, and fill them in `projectFontChainEntry` from the entry's own
      variants, verbatim, per the entry's discriminant. Extend the `maxCanvasPropertyString` loop to
      cover all seven strings. Amend the struct's doc comment.
- [x] `folio-go/canvas_projection_wire_test.go` -- Add the three keys to `canvasFontChainEntryWireKeys`,
      keeping sorted order, so both the Go-struct tie (`:200`) and the designer-guard tie (`:414`) hold.
- [x] `folio-designer/src/engine-protocol.ts` -- Add the three members to the `fontChains` entry **type**
      (`:285`) and to `isFontChainEntry` (`:333`): the exact-key list, per-field `typeof`, the length
      bound, and the namespace rule (a `face` entry's variants are face names, an `asset` entry's are
      asset keys). Do not shape-check an asset key as a digest.
- [x] `folio-go/table_columns_projection.go` -- Add `HeaderBold`/`HeaderBoldResolved` and
      `HeaderItalic`/`HeaderItalicResolved`, resolved through `resolveHeaderStyle` and never re-derived.
      Update the "sixteen members … 7 x 2 + 1 + 1" comment to the new arithmetic, and disclose the
      bool-absence limit the way the struct already discloses `headerStyle.fontSize: 0`.
- [x] DW-240's four mirrors, **in the same change** -- `canvas_projection_wire_test.go:474`
      `tableColumnsProjectionWireKeys`; `engine-protocol.ts:199-206` `TableHeaderStyle`;
      `engine-protocol.ts:378` `isTableColumns`'s `hasExactKeys` list (with `:410`, `:417`); and
      `table-style-command.ts:40` `TableHeaderStyleField` — **retire its stale `:34-39` comment by
      editing it**, the way 11.2 retired its twin at `component_commands.go:2510`.
- [x] `folio-designer/e2e/font-embed-boundary.spec.ts:74` -- Widen `StarterTemplate`'s
      `ReadonlyArray<string>` so the type stops lying about the file it parses. It reads only
      `Object.keys(fonts)`, so this is a truthfulness fix, not a behaviour one, and the compile stays
      green either way.
- [x] `folio-go/table_columns_projection_*_test.go` (or the nearest existing home) -- Add the **tie**
      between `tableHeaderStyleFields` and the projection's member list, so a tenth header-style field
      cannot repeat DW-240. This is the durable half of the fix; the two members are the cheap half.
- [x] `folio-designer/src/App.tsx` -- Delete `--text-font-weight` and `--text-font-style` from `:3394`.
      Pass the per-cut absence fact into `BooleanProperty` at `:2058`, computed from the component's
      family and `fontChains` over **every entry in the chain, not `entries[0]`** (do not reuse
      `declaredChainEntry` at `:2792`). Render a **genuine third state** in `BooleanProperty` (`:2692`)
      per CHECKPOINT 1's Q1(c) ruling — neither on nor off, its reason stated beside the control, and
      **always operable** so a declared-but-unavailable cut can be cleared. Never a disabled two-state
      control, and never a bare grey-out.
- [x] `folio-designer/src/App.css` -- Delete `font-weight` and `font-style` from `.canvas-text-paint`
      (`:152`); keep `font-size`. Add the absent-state rule for the toggle. **Compute specificity and
      source order by hand against every rule that could match** — jsdom verifies no CSS.
- [x] `folio-designer/src/canvas-authority-contract.test.ts` -- Add the AC2 prohibition, scoped to the
      painted-document surface rather than to the property's value, and prove it **in both directions**:
      a realistic mutation reds, and each of the legitimate chrome rules, the pasted-HTML fixtures and
      `font-catalogue.test.ts`'s absence assertions stays green.
- [x] Tests for the I/O matrix rows -- Cover: **a positive assertion that a bold component's painted node
      carries no weight and no slope** (the two custom properties are asserted by nothing today, so their
      deletion is otherwise invisible — use the `toHaveStyle` idiom `App.test.tsx:1433` already uses for
      `--text-line-baseline`); the resolved-face paint; the variant round-trip through the projection; the
      third state present **and** absent; the chain-whose-FIRST-entry-lacks-bold row that distinguishes
      the all-entries rule from the `entries[0]` rule; the table header committed/resolved pair; and the
      wire-key agreement on both projections.
      ⚠ **Drive the declared-but-unavailable state by its REAL route** — bold a Roboto element, then
      switch its family to a CJK-only chain — rather than hand-building a projection in that state.

**Acceptance Criteria:**

- Given an element declaring bold whose chain declares a bold cut, when the canvas paints it, then the
  fragment is painted in the face the engine resolved and **no `font-weight` or `font-style` is applied to
  it by the browser**, directly or through a custom property.
- Given the designer's production and test sources and its stylesheet, when the canvas-authority contract
  runs, then a synthetic weight or slope on painted document text is a **failure**, proved by a mutation
  that reds and by the legitimate chrome uses of the same properties that stay green.
- Given a chain in which **no entry** declares a bold cut, when the inspector shows the element, then the
  B control states that this family has no bold face and never renders as plainly on; and given a chain
  in which **some** entry declares one, the control is available.
- Given a document whose chain entry declares style variants, when the canvas projection is built, then
  the panel can read those variants back — the wire, the browser's guard and the Go struct agreeing by
  test — and given a header style declaring bold or italic, `TableColumnsProjection` carries both the
  committed and the resolved value.
- Given a brand-new document from `starter.folio`, when its author adds a text element and presses B,
  then the text paints in a real bold face; and the file declares no variant naming its own base.

## Spec Change Log

### 2026-09-06 — the I/O matrix has no both-flags row, and the fix is a patch (review round 1)

**Triggering finding (F1).** `boldItalic` is projected across the whole new seam and read by nothing:
`chainDeclaresCut` is typed `field: 'bold' | 'italic'`. An element with **both** flags set, on a chain
declaring `bold` and `italic` but **not** `boldItalic`, resolves to the base face and warns — while both
controls read plainly on. That is the state AC3 exists to prevent, arriving through the one combination
**this spec's I/O matrix never enumerated. The omission is the spec's.**

**Why this is a patch and not a loopback.** Under D-11.1.24 the test is *would re-deriving from a corrected
spec produce different code?* It would produce **this** code plus one generalised predicate. Reverting 17
files and ~965 lines to re-derive a predicate is the wrong trade, and D-11.2.12 already fenced this story
as indivisible. The matrix row is frozen and cannot be added here, so the rule is recorded in Design Notes
instead and is **binding there**.

**KEEP — what must survive re-derivation.** The two-slice engine seam is untouched and correct. The
wire-key ties across Go, the recorded set and the TS guard are right, and caught nothing only because they
were obeyed. The `table_header_style_test.go` tie between `tableHeaderStyleFields` and the projection's
member list is DW-240's durable half. The clearable-state test already drives Q1's real route — set bold,
then move the family under it — rather than hand-building the projection; keep that shape.

**Known-bad state avoided.** Applying the generalised predicate *alone*, which is what a naive reading of
the finding produces, makes B enter the unavailable state and say *"No bold face in this family"* when the
chain **does** declare a bold. It fixes the state and breaks the statement. **A panel that lies precisely
is not better than one that lies vaguely.**

### 2026-09-06 — the frozen block carries the drifted e2e claim, and only a human can correct it

`## Boundaries & Constraints` → **Never** contains *"DW-192/DW-193: the e2e suite is compiled and never
executed"*. **That is DW-193's wording with four load-bearing words dropped, and it is false** — see
**D-000.30**. The suite IS executed, at epic boundaries; `0c0f3e9` fixed two real e2e failures found at
Epic 12's boundary gate. What is true is narrower: CI runs only `test:e2e:compile`
(`.github/workflows/ci.yml:249`), so 42 cases across 16 spec files are typechecked and not executed **in
CI and not per story**.

It sits **inside `<frozen-after-approval>`**, so this story may not edit it. Both non-frozen sites — the
Verification Commands bullet and the Manual-checks bullet — now carry the corrected wording, and **the
corrected wording governs.** Flagged for the human who next reopens the block; the implementer was told to
follow the corrected statement, and did. *Caught by the implementation subagent, not by review — the
correction had been applied to one of the two non-frozen sites and not the other.*


## Design Notes

### F1 — THE COMBINED CUT. Ruled at review round 1. Binding; all three parts required.

The I/O matrix could not be amended (frozen), so this is where the missing row lives.

1. **Predicate — computed against the cut the element's RESULTING `(bold, italic)` combination requires,
   not against the control's own axis.** Both flags set → the required cut is `boldItalic`. `cutAbsent`
   asks whether *that* cut is declared by any entry in the chain.
2. **The sentence names the MISSING CUT, never the control.** `boldItalic` absent → **"No bold italic face
   in this family."** `bold` absent → "No bold face in this family." ⚠ **This part is not optional and the
   predicate is wrong without it.** With the predicate alone, B enters the unavailable state and says
   *"No bold face in this family"* on a chain that **declares a bold** — the statement is simply false.
   Deriving the wording from the **cut** gives one sentence per cut instead of one per control, and it
   cannot go false that way.
3. **When the missing cut is the combined one, BOTH controls enter the unavailable state — both are
   implicated, and marking only one implies the other is fine — but the reason is stated ONCE for the
   pair.** The announcement path is **one place**; this is the same fix as P9's double-announcement bug
   (the sentence is currently folded into `aria-label` *and* rendered as a visible non-`aria-hidden` `<p>`,
   phrased three ways), and the combined case would have made it a quadruple. Fix them together.

**What makes this state acceptable at all, and it belongs in the copy as well as here:** both controls stay
operable, and here that is not merely Q1's rule — **it is the escape route.** Turning off *either* B or I
moves the element to a combination the chain **does** declare, so the combined-absence state is
self-resolving through the controls the author is already looking at. The honest message is not *"this
family cannot do what you asked"* but *"this family cannot do **both** of those at once"* — the second is
the accurate one and it tells the author what to do next.

### P3 — the embedded arm, and the compensation that was holding it up

`carriedFaceKeys` (`App.tsx`) collected only `entry.assetKey`, which was complete while an entry named
one face. An entry may now name four, and for `{"asset": K1, "bold": K2}` the engine resolves a bold run
to `K2` — so with `K2` unfetched the fragment gets **no** `fontFamily` at all and falls to the
stylesheet's stack, while the document's own bold bytes sit in its `assets` map. **This is the same shape
as D-11.3.1: a compensation removed without supplying what it compensated for** — before this story that
fragment at least got `font-weight: 700`, so deleting the synthetic weight alone would have made embedded
bold *strictly worse*. The variants are collected per the entry's **discriminant**, never by the key's
shape: a 64-character face name is a legal face name (the same ruling `isFontChainEntry` records), so
filtering the whole population by `isCarriedFaceAssetKey` would cross the two namespaces on exactly the
value that looks like it could not.

### Rulings applied at CHECKPOINT 1 (2026-09-06) — apply, do not re-open

Recorded as **D-11.3.1 … D-11.3.5** in
[`epic-11-14-decision-log.md`](./epic-11-14-decision-log.md), committed at `42f0d44`. Summarised here so
the implementer needs no second file; the log is authoritative if they ever disagree.

- **Q1 → (c). A GENUINE THIRD STATE, always operable.** Not a disabled two-state control. The AC's own
  words decide it: *"states that this family has no bold face **rather than appearing to be on**"* — a
  disabled control still renders as on-or-off, so that sentence rules out both. **Refused outright:**
  disabling whenever the cut is absent, because `bold: true` would then be unclearable — *a control that
  has taken the document hostage*, against I-5's posture that the panel must never leave the author
  unable to reach what the document carries.
  ⚠ **The state must be driven by its REAL route in the test** — bold a Roboto element, then switch its
  family to a CJK-only chain — **not by hand-constructing the projection.** A state only reachable
  through a fabricated fixture is a state nobody has shown is reachable.
- **Q2 → ratified: no entry in the chain declares the cut.** See the `entries[0]` trap below.
- **Q3 → ratified (zero value, limit disclosed), but the GROUND IS NOT "no command can write it".** That
  justification does not hold: 11.2 shipped `tableHeaderStyleFields` with nine members, so the Go command
  layer *does* accept a header bold, and `"bold": null` decoding to `present(false)` makes committed-false
  representable. **The real ground: `CanvasProjection` is not the file format.** It is engine↔browser,
  both in this repo, moving in one commit — **not tag-bound**, so a tri-state can be added at any time,
  before or after 15.3, for free. DW-241's narrowing asymmetry does not apply. Consistency inside one
  struct beats a second idiom for one field, the same reasoning that chose plain `string` over `Presence`
  in 11.2.
  ⚠ **VERIFY BEFORE BUILDING ON IT, do not assume it because it was ruled:** the third state must be
  derivable **without** distinguishing committed-absent from committed-`false`. It needs
  *declared-true-but-no-face*, which survives the collapse. **If you find a case where the control
  genuinely needs absent-vs-false, the collapse is lossy where it matters and Q3 flips to a tri-state —
  stop and report it.**
- **No release build. Ruled**, on the evidence in the Code Map: the starter is fingerprinted only into
  gitignored regenerated output and `verify-offline-release.mjs:113` class-checks `.folio` rather than
  pinning a digest. **Do not run `npm run build`.** (The frozen Ask First was honoured — it was asked and
  answered.)


**Why AC2 cannot be scoped by value, and what to scope by instead.** The existing CSS prohibition bans
`property: value` pairs because the forbidden values (`white-space: normal`) appear nowhere legitimate.
Here they do: the chrome writes `font-weight: 500` in seven rules and `font-weight: 600` in two, and
`.property-fx` is deliberately italic. The invariant is not "this value is wrong" but **"this property may
not be applied to painted document text"** — a page/chrome distinction DESIGN.md already draws ("Don't mix
page tokens into chrome or chrome tokens onto the page"). So scope the prohibition to the canvas text
surface: the `.canvas-text-*` rules in `App.css`, and the inline style objects on those elements in
`App.tsx`. A name-based ban on `--text-font-weight` alone is **not** acceptable — a rename defeats it, and
the guard must catch the property arriving by any spelling. Because CSS here has zero executable coverage,
this guard is the only thing standing there; treat it as the deliverable, not as paperwork.

**The `entries[0]` trap, and why the starter cannot test it. `App.tsx:2792` `declaredChainEntry` IS THE
WRONG FUNCTION TO REUSE HERE — named explicitly, because it is the one an implementer reaches for.** It
returns the first entry and sits three functions from the call site. Using it would mean a chain `["Noto Sans SC","Roboto"]`
reports "no bold face" while Latin would bold perfectly well. **The starter's own chain cannot detect the
error** — its first entry (Roboto) declares bold, so the first-entry rule and the all-entries rule return
the same answer. An assertion whose two sides could be equal is not an assertion (D-11.2.8): the test for
this row must use a chain whose first entry has no bold and whose later entry does, and something must pin
that the two rules genuinely disagree on it.

**The bool on a zero-value wire.** `TableColumnsProjection` spells absence as the zero value because its
key set is pinned exactly in both directions, so `omitempty` is unavailable. For a bool that makes
committed-absent and committed-`false` the same wire value. It is acceptable here for the same reason the
struct already accepts `headerStyle.fontSize: 0`: **no command can write it.** The designer's standing rule
is that off is absent, not false (commit `3bacfec`), and 11.2 measured that `"bold": null` decodes to
`present(false)` rather than erroring. Disclose the limit in the struct comment rather than inventing a
tri-state; an undisclosed limit ages into false reassurance (D-11.2.11's guardrail).

**Why this story runs no matrix suite.** D-11.2.12: of everything originally assigned to 11.3, only the
bold golden moved a rendered byte, and it is now **Story 11.5**. Splitting *removed* the four-target gate
from this half. This is the **inverse** of D-11.1.18, where splitting would have *multiplied* the same
gate — same gate, opposite arithmetic, opposite answer. Recorded here so a later reader does not read the
two rulings as inconsistent.

**Why the projection and the control cannot be split apart.** The canvas is told the resolved outcome, not
the requested flag, so the control's third state is *derived from* the projection. A projection shipped
without its consumer is unconsumed; a control shipped without its projection is guessing (D-7.4.5).

## Verification

Run each module's commands in **its own** invocation — a `cd` persists through a compound command and makes
later relative paths resolve elsewhere, printing `lstat …` lines that read like passes. Use `-count=1` on
every Go run: the cache serves a stale PASS for tests that walk the filesystem. Never `&&`-chain these; a
conjunction silently drops everything after its first failing term.

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- baseline at `fa93107` is **2187 pass / 2 fail / 5 skip**,
  exit 1. The ONLY acceptable failures are `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)`
  child — a mandated permanent red, never to be "fixed". **A third distinct failure is a HARD STOP:
  report it, do not triage around it.** Count from `go test -json` `Action` events carrying a `Test`
  field; a plain run prints no totals and a carried-forward figure is not a measurement.
- `cd lint && go test -count=1 ./...` -- four `ok`, **227 tests**, exit 0.
- `gofmt -l /Users/panitw/Projects/folio/folio-go /Users/panitw/Projects/folio/lint` -- from the **repo
  root with absolute paths**; empty output. Any `lstat` line is a non-measurement.
- `cd folio-designer && npx tsc -b --force` -- exit 0. **`--force` is required**: `tsc -b` exits 0 without
  typechecking when its build info is current.
- `cd folio-designer && npm test` -- baseline **64 files / 953 tests**, all passing, exit 0. Never run
  `npx vitest` from the repo root: it picks up the Playwright specs and produces a false mass-failure.
- `cd folio-designer && npx oxlint` -- expect exactly the pre-existing `only-export-components` warnings;
  quote the count rather than calling it clean.
- `cd folio-designer && npm run test:e2e:compile` -- must stay green (`font-embed-boundary.spec.ts` reads
  the starter). ⚠ **State it as D-000.30 states it: a green `test:e2e:compile` is a TYPECHECK and not
  coverage — the suite is exercised at EPIC BOUNDARIES, not in CI and not per story.** Do **not** write the
  shorter "compiled and never executed": that is DW-193's wording with four load-bearing words dropped, it
  is false, and it invites a reader to conclude the suite has rotted (it has not — `0c0f3e9` fixed two real
  e2e failures found at Epic 12's boundary gate). CI runs only `test:e2e:compile`
  (`.github/workflows/ci.yml:249`), typechecking **42 cases across 16 spec files** without executing them.
- **DO NOT run the `-tags=matrix` suite.** D-11.2.12 removed it from this story. If you believe you need
  it, something here moves a rendered byte and the seam was mis-drawn: **stop and say so.**
- **DO NOT run `npm run build`. RULED at CHECKPOINT 1** on the Code Map's measurement: the starter is
  fingerprinted only into gitignored regenerated output, and `verify-offline-release.mjs:113` class-checks
  `.folio` rather than pinning a digest. The question was asked and answered; do not re-ask it.

**Manual checks:**

- **Mutation-proof every new guard; a green suite over correct code proves nothing.** For each of: the
  AC2 prohibition, the third-state rule, the wire-key ties, and the DW-240 tie — reverse the behaviour at
  the production site and confirm the named test reds, then restore and confirm it greens. **Anchor every
  mutation and check WHERE it landed**; an un-anchored edit reds an unrelated test and proves the wrong
  thing. Report which test caught which mutation.
- **Mutate by deletion as well as substitution** for the third state: delete the absence check entirely
  and confirm a test reds, not merely flip its condition.
- **Red-prove AC2 in both directions.** Write `font-weight: 700` into `.canvas-text-paint` and confirm the
  contract reds; separately confirm that `App.css`'s nine legitimate chrome rules, `App.test.tsx`'s
  pasted-HTML fixtures and `font-catalogue.test.ts`'s absence assertions all stay green. A prohibition
  that reds the chrome is not scoped; one that misses the mutation is not a prohibition.
- **Read the new CSS by hand.** jsdom parses no stylesheet, so the absent-state rule's correctness is
  invisible to the suite. For each new declaration find every other rule setting the same property on the
  same elements — **including type selectors inside existing rules, e.g. `.foo > span`** — and compare
  specificity AND source order. Assert the toggled class, which is the only handle the suite has.
- **STATED LIMIT, recorded here rather than left as prose in a report nobody re-reads.** The
  `.property-toggle-unavailable` / `.property-unavailable` rules — including the hand-computed specificity
  argument against `.property-toggle[aria-pressed="true"]` — are **witnessed by nothing that executes.**
  Deleting either rule leaves the third state visually identical to a plain toggle with the whole suite
  green. The tests can assert only that the class name is on the element, and they do. **This is a
  permanent property of CSS in this repo, not a gap this story can close.**
  ⚠ **Hand-computed specificity says nothing about LAYOUT.** Check separately that the new `<p>` does not
  break the B / I row: `.property-toggle-row` is `display: flex; align-items: center`, so making one of the
  two editors taller can push B and I off each other's line. jsdom cannot see that either.
- **A green `test:e2e:compile` is a TYPECHECK and not coverage — the suite is exercised at EPIC BOUNDARIES,
  not in CI and not per story (D-000.30).** CI runs only `test:e2e:compile`
  (`.github/workflows/ci.yml:249`), typechecking 42 cases across 16 spec files without executing them; the
  suite itself is not rotten (`0c0f3e9` fixed two real e2e failures found at Epic 12's boundary gate). Use
  **this** wording in the report, never the shorter "compiled and never executed". Separately, `App.tsx`
  carries a pointer carve-out that waives every AD-17 prohibition inside that file — say so too.
- **Confirm the starter's claims yourself before editing it**, with the search and a positive control
  recorded for each absence: nothing pins its bytes, nothing pins its version string, and the loader
  accepts `2.0`. D-000.4 — an absence relayed from a prompt or a subagent is a lead, not a result.
- **Discharge Q3's open check before relying on the zero-value collapse.** Confirm the third state is
  derivable from *declared-true-but-no-face* alone, never needing committed-absent vs committed-`false`.
  If it does need that distinction, the collapse is lossy where it matters — **stop and report it**; Q3
  flips to a tri-state and that is a spec change, not an implementation choice.
- Confirm the 11.1 shipped-face guard table did not move. If any count changed, the story drifted.
- Confirm no golden digest moved and **no fixture was added** — that is Story 11.5's.
- Report `git status --porcelain` and confirm **nothing was committed**: HEAD must still be `fa93107`.
  Note that `epic-11-context.md` was recompiled by this workflow's own step 1 and is a legitimate,
  uncommitted modification.

## Suggested Review Order

**Start here — the whole story in one deletion**

- The synthetic weight is gone; only the engine's own `font-size` survives on the painted surface.
  [`App.css:170`](../../folio-designer/src/App.css#L170)

- Its twin: the fragment's family comes from the resolved face, and no weight is written beside it.
  [`App.tsx:3559`](../../folio-designer/src/App.tsx#L3559)

**The guard that is the deliverable — three rules, because one shape could not cover it**

- Longhand, selector bound widened so a grouped selector across lines cannot slip through.
  [`canvas-authority-contract.test.ts:109`](../../folio-designer/src/canvas-authority-contract.test.ts#L109)

- Shorthand, value-scoped: `font: bold …` was the live idiom this commit's own CSS already uses.
  [`canvas-authority-contract.test.ts:124`](../../folio-designer/src/canvas-authority-contract.test.ts#L124)

- JSX, bounded by the opening tag — the first `}` bound could not see past a conditional spread.
  [`canvas-authority-contract.test.ts:146`](../../folio-designer/src/canvas-authority-contract.test.ts#L146)

**The projection — DW-239, three sites that blank the canvas if they disagree**

- Three always-present keys; `omitempty` here would reject whole documents at the browser guard.
  [`page_setup.go:615`](../../folio-go/page_setup.go#L615)

- Variants copied verbatim, per the entry's own discriminant, and bounded like every other string.
  [`page_setup.go:678`](../../folio-go/page_setup.go#L678)

- The recorded protocol set, tied to both the Go struct and the browser's guard.
  [`canvas_projection_wire_test.go:98`](../../folio-go/canvas_projection_wire_test.go#L98)

- The browser half of that tie: exact keys, per-arm namespace, no digest-shape check on an asset key.
  [`engine-protocol.ts:365`](../../folio-designer/src/engine-protocol.ts#L365)

**The third state — F1's rule, where the review round landed**

- Asks the chain for a named cut; `boldItalic` is a cut, not a combination of two answers.
  [`App.tsx:2750`](../../folio-designer/src/App.tsx#L2750)

- The cut the RESULTING (bold, italic) combination needs — the predicate, part one of three.
  [`App.tsx:2776`](../../folio-designer/src/App.tsx#L2776)

- The sentence is derived from the CUT, never the control; naming the control makes it false.
  [`App.tsx:2813`](../../folio-designer/src/App.tsx#L2813)

- Conservative across a selection: absent only when no selected component's chain declares it.
  [`App.tsx:2791`](../../folio-designer/src/App.tsx#L2791)

- One announcement path — `aria-describedby` at the paragraph, plain accessible names.
  [`App.tsx:2821`](../../folio-designer/src/App.tsx#L2821)

- Structural fix, not a specificity argument: the note stacks below rather than inside the flex row.
  [`App.css:388`](../../folio-designer/src/App.css#L388)

**The regression review caught — embedded bold had lost its face entirely**

- Variant asset keys are registered too, gated by the discriminant and never by key shape.
  [`App.tsx:311`](../../folio-designer/src/App.tsx#L311)

**DW-240 — the read-back, and the tie that stops the tenth field repeating it**

- The committed/resolved pair for bold and italic, resolved through the one cascade.
  [`table_columns_projection.go:93`](../../folio-go/table_columns_projection.go#L93)

- The durable half: the projection's member list is tied to `tableHeaderStyleFields` itself.
  [`table_header_style_test.go:831`](../../folio-go/table_header_style_test.go#L831)

**The starter, and the test that makes shipping it a checked act**

- Real cuts, no self-referential variant, SC bare, version raised to what the content requires.
  [`starter.folio:9`](../../folio-designer/public/templates/starter.folio#L9)

- Nothing that runs had ever parsed the shipped file; now three tests do.
  [`starter_template_test.go:62`](../../folio-go/starter_template_test.go#L62)
