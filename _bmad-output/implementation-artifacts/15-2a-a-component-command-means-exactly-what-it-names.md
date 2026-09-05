---
title: 'Story 15.2a: A component command means exactly what it names'
type: 'bugfix'
created: '2026-09-05'
status: 'done'
baseline_revision: 'e8ff6e619e1531a2d477b715414076f46513bdcf'
baseline_commit: '3c7e2264046cdb471b03aab15436421fbed064f8'
review_loop_iteration: 1
context:
  - '{project-root}/_bmad-output/implementation-artifacts/epic-15-context.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-7-8-decision-log.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** A designer command can be made to name one component and property while changing a
different component's different property, and neither side stops it. **Measured by execution at
`7b03ea0`, not carried on report:** the designer's `rawNumberLiteral` splices an author's typed
string into command JSON **unquoted** (DW-32), `page-setup-command.ts` performs the identical splice
at a second front door (DW-73), two encoders escape non-BMP text from `charCodeAt(0)` and emit a
**lone surrogate** so a bind segment or asset key binds to a path the author never typed (DW-75), and
`component-command.ts` splices `id`, `type` and `band` raw into five builders. On the
Go side `ApplyComponentCommand` decodes into `map[string]json.RawMessage` at three nesting levels, so
a duplicate key is resolved **silently by last-wins** while the arity check still counts the
deduplicated map. Executed: a command naming `e1` in `pageHeader` mutated `e5` in `pageFooter` and
returned `nil`.

**Reachability, corrected — and this correction is the point.** The frozen text this spec derives
from calls the exposure *"keystroke-originated and self-inflicted in a local, serverless
application: HIGH by mechanism, low by encounter"*, and directed that the encounter half be
**measured rather than assumed**. It was, and **it is false in the direction that matters.** For
`rawNumberLiteral` itself the framing holds — no document value reaches it. But a **bind segment**
does. Bind segments are JSON object **keys taken verbatim from the author's sample-data file**,
nothing constrains a JSON key, and those keys travel through `component-command.ts`'s broken
`charCodeAt(0)` quoter. Open a data file, click a node, press Connect — no typing anywhere, and the
value that reaches the document is not the value the file held. Encounter is therefore **not low**:
it is whatever the data file the author opened says it is. This changes the **severity framing, not
the scope**, because that encoder is already one this story consolidates, so quoting it correctly
fixes it as a consequence. A later reader will believe the frozen block, so it says so here rather
than only in the Design Notes.

**Superseded at implementation, 2026-09-05.** This paragraph previously read that nothing constrains
`component.id`'s charset, so the defect was reachable by opening a crafted `.folio`, selecting an
element and pressing Delete. **That is false.** `internal/template/ids.go:validateElementID`
(AD-10/AC34) enforces `^e[0-9a-z]+$` on every element id at parse time, so such a document cannot be
loaded at all. The claim rested on `engine-protocol.ts:331`, which is the **inbound projection**
guard — a different direction and a different population, and one that can only ever carry ids the
loader already admitted. The conclusion it was offered for nevertheless holds, by the mechanism
above.

**Approach:** Make the property true on both sides in one commit — a **single shared command-JSON
authority** in the designer that every encoder routes through, so no call site can splice again; and
a **duplicate-key refusal** in the engine that narrows **both** exported command doors —
`ApplyComponentCommand` and `ApplyPageSetupCommand` — so the property is asserted against the engine
rather than against the fix. The encoder becomes **total**: every value spliced into command JSON
goes through the authority — strings quoted by `JSON.stringify`, and numerics **passed through
byte-for-byte when the draft is already a valid JSON number, or replaced by `null` when it is not**.
No `Number()` round-trip happens on the numeric path at all. Validity stays in Go, which alone holds
the numeric grammar; one Go diagnostic is split so it reports the cause it actually found.

## Boundaries & Constraints

**Always:**
- **The two halves land in ONE commit**, with a test that reads both sides. Without the Go half the
  only available test is *"the encoder produces well-formed JSON"* — a test of the fix, not of the
  property, which goes green the moment a future encoder regresses.
- **The engine test hands duplicate-key BYTES directly to `ApplyComponentCommand`** — never only the
  encoder.
- **Acceptance asserts a non-BMP round trip explicitly** (DW-75). A test that only proves "the
  payload is valid JSON" goes green without ever proving the round trip; every existing escaping
  test uses BMP-only inputs.
- **Existing valid output stays byte-identical.** 52 `toBe('{"kind":…')` assertions pin key order and
  the **unquoted number form** (`"value":13`, `"value":1.5`, `"value":0.001`). ~40 are in
  `App.test.tsx`. The authority must emit JSON *numbers*, not quoted strings, for numeric fields.
- **The engine refusal uses `componentFailure`**, not a bare `fmt.Errorf`. Every existing decode-site
  refusal is bare and therefore surfaces as `ENGINE_REJECTED` with no location.
- **The refusal carries the diagnostic code of the door it was raised at** — `COMPONENT_INVALID` for
  `ApplyComponentCommand`, `PAGE_SETUP_INVALID` for `ApplyPageSetupCommand`. Both codes are asserted,
  one per door. `engineFailure` matches `*ComponentCommandError` at `main.go:236` **before** the
  page-setup fallback at `:257`, so a page-setup refusal raised as a plain `componentFailure` reports
  the wrong code and misses the designer's only code-branching page-setup consumer.
- **The accept-set of both doors is narrowed or unchanged at every input.** No draft may be accepted
  that was refused before. This is the property the story is named for, and it is what the encoder's
  numeric path must not violate.
- **`ElementID` is empty on the refusal.** A duplicate key means the named id is untrustworthy —
  executed proof: the command named `e1` and changed `e5`. Naming any id names the wrong one.
- **Go stdlib only.** `folio-go/gomod_test.go:60` is a strict module allowlist; any new dependency
  reddens the suite by construction.
- **Both doors are narrowed.** `ApplyComponentCommand` and `ApplyPageSetupCommand` share one
  duplicate-key scan. A soleness guard covering one of two exported doors is a test that passes
  while the property is false.
- **The halves are one story.** One commit is better; two commits are acceptable. The component and
  page-setup halves must **never** be split across stories.
- **Red-prove the refusal at every level it exists**, not only the top one: top-level `kind` and
  `version`, `changes` in the component command, and `margin` in page setup. A guard proved only at
  the top level leaves the nested object as the surviving door.
- **The red-proof includes the document-originated leg**: a sample-data file whose JSON **key** is
  non-BMP, accepted through `acceptSampleData`, the node clicked and Connect pressed, asserting the
  bind segment on the wire is the author's own code points. The typed-draft leg proves the encoder;
  only this leg proves the severity.
- **The same-arity escalation leg must not be claimed where it does not apply.** Page setup has a
  single `kind`, so that leg is inapplicable there — say so explicitly rather than letting one leg's
  inapplicability read as coverage.
- **Red-prove every new test by mutation**, and echo the mutated line back, so a real green is
  distinguishable from an edit that never applied.

**Ask First:**
- Both plan-gate questions are **settled** — see `## Rulings`. Do not re-open them.
- Any finding that a *further* document-content path reaches a raw splice beyond the one recorded
  here.
- Any need to change a diagnostic string **other** than the one `## Rulings` names, or to change it
  in a way that alters what a *genuine* overflow reports.
- Any discovery that a test **does** pin one of the two diagnostic strings this story rewords. The
  ruling that the reword is free rests on nothing pinning them; if that is false, stop.

**Never:**
- **No assertion, test name, or acceptance criterion may contain the string `p`+`adding`.** Ruling
  D-D.1: Story 12.4 (`epics.md:4162`) makes the Go layer refuse those four fields on non-table kinds;
  measured CLOSED that no designer control authors them (`App.tsx`: 0 occurrences against a
  population of 17 `field: '<name>'` literals; `component-property-command.test.ts`: 0 against 3
  `fontSize|lineSpacing` hits), so such a test is vacuous today and would turn 12.4 from a new guard
  into a change to a passing acceptance. The four fields **stay** in `PropertyField` and
  `pointFields` unchanged and encode like any other point field — deleting them is equally forbidden,
  because 12.4 narrows by kind and does not retire them. This prohibition binds the Tasks,
  Acceptance Criteria and every test; it does not bind this constraint clause.
- **Do not bundle Story 12.4's refusal.** It is a different subject with its own red-proof, not a
  third member of the welded set.
- **Do not fold `ratioFields` into `pointFields`.** Verified at `component-property-command.ts:12-19`:
  `template.DecodeLineSpacingRaw` performs the ×1000 itself, so `1.5` on the wire is 1500
  thousandths and `1500` would be refused. Merging the sets is a silent unit bug.
- **Do not touch `POSITIVE_LENGTH_FIELDS` or `ORIGIN_FLOOR_FIELDS`** — membership *or* formatting.
  `engine-bounds-mirror.test.ts:264/:318` and `:352/:385` read them with anchored single-line
  regexes and drift-prove by substitution; `:310`/`:381` additionally pin `App.tsx:2150`'s exact
  text. Re-wrapping a line or changing quote style breaks a source-text guard. That mirrored Go/TS
  invariant moves in one commit with both sides (D-7.4.5); this is not that commit.
- **Do not re-hand-roll an escape table.** `quote()` stays routed through `JSON.stringify` (Story
  8.2, deliberate).
- **No additions to or removals from `pointFields`.** New numeric fields are Epic 12/14's business.
- **Do not unify with `engine-protocol.ts`.** That is the INBOUND projection validator (engine →
  browser): a different direction and a different population. Coupling them makes a projection change
  fail an outbound-command test.
- **No new npm or Go dependency.** `design-contract.test.ts:71` pins package + lockfile.
- **No browser-side rule about what a number *means*.** **D-D.2**: the panel must not become a second
  authority on validity. Go's grammar is specific and non-obvious — it rejects exponents, rejects more
  than three decimal places, and bounds against `MaxCanvasMillipoints` — so a browser rule that
  duplicated it would invite D-7.4.5's mirror-test burden for a whole grammar, and one that
  approximated it would refuse what the engine accepts. One authority, and it is Go.
  **The shape check the encoder does apply is not such a rule**: testing that a draft is a valid JSON
  number decides only whether bytes can *reach* Go's rule, never what they mean. Anything that matches
  travels verbatim; the encoder never rewrites the author's literal. Do not narrow that check below
  the JSON number grammar — a narrower one would send `1e3` as `null` and cost the located
  *"must be a decimal with at most three places"* Go answers today.
- **Do not add a charset constraint on `component.id` at load — one already exists.**
  `internal/template/ids.go:validateElementID` (AD-10/AC34) enforces `^e[0-9a-z]+$` at parse time.
  Nothing is to be built here and **nothing is to be deferred**: filing a register entry saying the
  constraint is missing would write a falsehood into the record.
- **Do not delete the overflow detector while fixing its message.** The mixed branch carries a real
  overflow guard in the same condition; splitting means two branches, not a changed string.
- **No branch, no push, no commit.** The orchestrator makes every commit.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Injection via numeric draft | Author types `0}},"ids":["other"],"changes":{"width":{"op":"set","value":10` into `width` and blurs | The emitted command names exactly the selected ids and the `width` change; no second `"ids"` or `"changes"` key appears at any level. **Executed today:** the payload is valid JSON whose parse collapses to `ids:["other"], changes:{width:{op:set,value:10}}` — the attacker controls **both** the target and the change, and the author's own selection vanishes entirely | Unparseable drafts encode as `null`; the engine refuses, per `## Rulings` |
| Non-BMP bind segment | `bindComponentScalarCommand('e1', ['a\u{1F600}b'])` | `JSON.parse` of the payload returns the **three code points the author typed**, `U+1F600` intact — not `U+D83D` | none; it must round-trip |
| Non-BMP asset key | `setComponentAssetCommand` with an astral character in id or media type | same round trip | none |
| Non-BMP bind segment from a data file | A sample-data file whose JSON **key** is non-BMP; accepted through `acceptSampleData`, node clicked, Connect pressed — **no typing anywhere** | The bind segment on the wire is the author's own code points | **The document-originated leg.** A JSON key is unconstrained and reaches the broken quoter verbatim |
| Page-setup splice | Author types `0,"preset":"custom","orientation":"landscape","height":9999` into page width | The page-setup command carries one `preset`, one `orientation`, one `width` | Unparseable drafts encode as `null`; the engine refuses, per `## Rulings` |
| Duplicate key, top level | Bytes with `"ids"` and `"changes"` twice reach `ApplyComponentCommand` | **Refused.** Today: accepted, `e5` mutated, `nil` returned | `componentFailure`, `ElementID: ""`, document-scoped `DataPath` |
| Duplicate key, `changes` object | `{"value":{…"FIRST"},"value":{…"SECOND"}}` | **Refused.** Today: accepted, `SECOND` applied | as above |
| Duplicate key, operation object | `{"op":"set","value":"FIRST","value":"SECOND"}`; and `"op"` twice | **Refused.** Today: both accepted | as above |
| Duplicate `version` | `{"version":0, …, "version":1}` | **Refused.** Today: **accepted** — the `version == 1` gate reads the last one, so a `version:0` command is admitted by appending a second key | as above |
| Same-arity kind escalation | `deleteComponent` (arity 3) bytes whose duplicate resolves the kind to `deleteFontChain` (also arity 3) | **Refused at the duplicate.** Today: **accepted** — `componentFields` passes and dispatch lands in the wrong handler, stopped only later by that handler's own field names. Arity is a coincidence, not a check | as above |
| Object inside an array | A duplicate key in an object nested inside an array | **Refused** — the scan must reach every nesting level, arrays included | as above |
| Unparseable numeric draft | Author types `abc`, `0x10`, `.5`, `007`, `+5` or a blank into `width` — none is a valid JSON number | The encoder emits `null`; the engine refuses with `must be a number` | At baseline these produced malformed bytes and a generic parse failure. `null` is the improvement |
| Valid JSON number the engine refuses | `1e3` and `1e21` typed into `width` | **Both travel byte-for-byte** and earn Go's own located refusal, `must be a decimal with at most three places`. They must behave **the same as each other** | **This pair is the regression's signature.** An implementation where `1e3` is accepted and `1e21` refused has reintroduced the coercion |
| Large integer literal | `9007199254740993` typed into `width` | Reaches Go **verbatim**, no precision loss — no `Number()` round-trip occurs | A `Number()` path silently yields `...992` |
| Genuine numeric overflow | `width` literal `99999999999999999999` | Still refused as an **overflow** — the message must not move | **Executed today:** `width overflows millipoints`. The red-proof runs both directions against this exact pair |
| Empty margin draft | Author clears a margin field; `page-setup-command.ts:6` already emits `null` | Refused as not-a-number, matching the comment that already claims a field-specific diagnostic | **Executed today:** `page.margin.<name> overflows millipoints` — a shipped wrong-cause diagnostic nobody filed |
| Duplicate key, page setup top level | Bytes with a repeated key reach `ApplyPageSetupCommand` | **Refused.** Today: `len(raw) != 7` cannot trip, because the map already deduped | `ElementID: ""`, document-scoped `DataPath` |
| Duplicate key, `margin` object | `{"top":1,"top":2,"right":…,"bottom":…,"left":…}` | **Refused.** Today: `len(margins) != 4` cannot trip either — the nested door | as above |
| Valid command, unchanged | Any command the designer emits today, either door | Byte-identical payload; accepted; same projection | none |

</frozen-after-approval>

## Rulings — settled at this plan gate, 2026-09-05

**Both plan-gate questions are closed. They are recorded here because the reasoning, not just the
answer, is what a later reader needs.**

**Ruling 1 (D-D.2) — the encoder is total; validity stays in Go; one Go diagnostic is split.**
⚠ **The mechanism this ruling originally prescribed — `JSON.stringify(Number(v))` — was SUPERSEDED at
review loop 1 (see `## Spec Change Log`); it took from Go the authority the same ruling reserved to
it.** The mechanism is now: a draft matching the **JSON number grammar**
`/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/` travels **verbatim**; anything else becomes
`null`. The ruling's substance is unchanged and is what forced the correction — A
browser-side pre-refusal — DW-32's original prescription — was **rejected**: it makes the panel a
second authority on what a number is, and Go's grammar is specific and non-obvious.

But `null` alone does not deliver a located message, and the ruling depends on it doing so.
**Executed through the exported door**, not read: `null` and a 20-digit overflow **both** return
`width overflows millipoints`. The cause is `parseMillipoints`'s whole-part loop, whose single
condition `c < '0' || c > '9' || whole > (1<<63-1)/10` mixes *not a digit* with *genuine overflow*.
**So: split that branch** — a non-digit reports that the value is not a number, a real overflow keeps
reporting overflow. Red-prove **both directions**; fixing the message must not delete the detector.

This also repairs a **shipped defect nobody filed**: `page-setup-command.ts:6` already emits `null`
for an empty draft, under a comment claiming Go returns the field-specific diagnostic — and what an
author with an empty margin gets today is `page.margin.top overflows millipoints`. The convention was
right; the diagnostic under it has been wrong the whole time.

**The reword is free because nothing pins it.** Measured: `overflows millipoints` occurs **twice**,
both in `folio-go/page_setup.go` (`:2058`, `:2076`); `must be a finite number` **once** (`:2035`).
**Zero** test files and **zero** `.ts`/`.tsx` files reference either. Positive control, because a bare
zero proves nothing: tests *do* pin engine message strings.

⚠ **RE-MEASURED AT IMPLEMENTATION, because two figures in this paragraph disagreed with the ledger
entry quoting them.** All counts below are **FILES, not occurrences**, over the repository's test
files only (`folio-go/**/*_test.go` and `folio-designer/src/*.test.{ts,tsx}`), taken with `git grep -l`
at the baseline commit `3c7e226`:

| string | test files at baseline |
|---|---|
| `overflows millipoints` | **0** |
| `must be a finite number` | **0** |
| `must stay within` (positive control) | **3** |
| `is required` (positive control) | **14** |

The paragraph above said 2 and 13 for the two controls; DW-196 said 3 and 105 — the 105 was the
whole-repository file count for `is required`, not the test-file count, and the two were never
comparable. Both are now stated with their population. The zeros are unchanged, so the ruling holds.
**If implementation finds a test that does pin either string, stop and ask** — the ruling rests on
this.

**DW-32's "what closing it requires" sentence is superseded, deliberately.** Replace it with, verbatim:

> A total encoder: every value spliced into command JSON goes through `JSON.stringify` — strings
> quoted, numerics passed through **verbatim** when the draft already matches the JSON number
> grammar and replaced by `null` when it does not — never re-parsed, so the author's literal is never
> rewritten. The
> refusal stays in Go, which alone holds the numeric grammar (no exponent, at most three decimal
> places, bounded by MaxCanvasMillipoints); the encoder's job is to make bytes that reach that rule,
> not to duplicate it. A browser-side pre-refusal was the original prescription and was rejected at
> 15.2a's plan gate (D-D.2) because it makes the panel a second authority on validity.

Keep that entry's severity paragraph and **add the bind-segment finding to it** — a sample-data JSON
key is document-originated, reaches the broken quoter with no typing, and is the part of the
mechanism the entry underestimated.

**Ruling 2 (also D-D.2 — the same log entry as Ruling 1) — narrow both doors. The frozen AC is
amended at the plan gate, deliberately.** Verified by reading:
`ApplyPageSetupCommand` (`page_setup.go:1903`) decodes with `dec.Decode(&raw)` into
`map[string]json.RawMessage` — last-wins — and gates on `len(raw) != 7`, an arity check a duplicate
key **cannot** trip because the map already deduped. The identical shape recurs **nested** at
`margin` (`len(margins) != 4`, `:1938-1940`). And `equalNumber(raw["version"], "1")` means the
duplicate-`version` trick works on this door too.

Three grounds. The property the story is *named for* is a property of the command **channel**, so a
soleness guard over one of two exported doors passes while the property is false. The cost is
asymmetric and time-boxed: narrowing an exported decoder is free before the tag and breaking after
it, and **Story 15.3 cuts that tag inside this same epic** — "file it before the tag" is one story
away from "forever". And DW-73 already puts page setup's encoder in scope on the TS side, so doing
the TS half and deferring the Go half **manufactures** the asymmetry DW-73 exists to prevent: quoted
on the wire, last-wins at the door.

**Two things the page-setup half must cover that the original analysis missed:** `preset` and
`orientation` are spliced raw **inside quotes** at `page-setup-command.ts:7`, so they need quoting
like any other string — `engine-protocol.ts:275` is a *remote* defence, and a check must sit where
its operands are. And `width`/`height`/`margin.*` take the same numeric treatment, which is what
the existing `value === '' ? 'null' : value` was reaching for.


## Code Map

**Measured at `e8ff6e6`.** HEAD moved three times during investigation (`28cd225` → `7b03ea0` →
`e8ff6e6`); `git diff --name-only 28cd225..e8ff6e6 -- folio-designer/src folio-go` is **empty**, so
every anchor below holds.

### The designer's encoders — SIX modules, three answers (the register says five)

| File | Escaper | Answer |
|---|---|---|
| `folio-designer/src/component-property-command.ts` | `quote()` `:81` | **A — `JSON.stringify`** |
| `folio-designer/src/table-column-command.ts` | `quote` `:3`, `number` `:4` | **A, for STRINGS.** ⚠ **Corrected at implementation: it was NOT wholly correct.** `number` (`:4`) existed but **two builders bypassed it** — `addTableColumnCommand` spliced `${index}` and `moveTableColumnCommand` spliced `${toIndex}` raw. Both are browser-derived list positions, so nothing reached them, but "the shipped model" overstated it: a file with a correct helper that two of its own seven builders ignore. |
| `folio-designer/src/font-chain-command.ts` | `quote` `:20`, `index` `:24` | **A** — its header `:8-13` already declares this "the only correct answer". **No production caller** (Story 16.9 deleted the chain editor); test-only, 7 tests. |
| `folio-designer/src/component-command.ts` | `quote()` `:52-67` | **B — hand-rolled, `charCodeAt(0)`** |
| `folio-designer/src/component-asset-command.ts` | `quote()` `:18-33` | **B** — byte-identical to the above but for indentation (tabs vs spaces) |
| `folio-designer/src/page-setup-command.ts` | none | **C — no escaping at all**; `:6-7` `const number = (value) => value === '' ? 'null' : value`, spliced at six sites. **The only encoder with no test file.** |

**The four defects, each proved by execution, not by reading:**

1. **DW-32 — `component-property-command.ts:58-62`** `rawNumberLiteral` returns the typed string
   verbatim; spliced at `:53` for `pointFields` (`:7`) ∪ `ratioFields` (`:19`). Executed: the story's
   payload yields valid JSON whose parse collapses to `ids:["other"], changes:{width:{op:set,value:10}}`
   — the attacker controls **both** target and change, and the legitimate selection vanishes.
2. **DW-73 — `page-setup-command.ts:6-7`.** Executed: one typed width overrides `preset` *and*
   `orientation`. **Worse than DW-73 records:** `preset` and `orientation` are themselves spliced raw
   *inside quotes* (`"preset":"${preset}"`), taken straight from the projection at `App.tsx:136-137`
   and `:1227`, defended only by a **remote** allowlist at `engine-protocol.ts:275`.
3. **DW-75 — `component-command.ts:52-67` / `component-asset-command.ts:18-33`.** `for (const
   character of value)` iterates by code point; `character.charCodeAt(0)` reads only the high unit;
   the `0xd800..0xdfff` branch emits it and the low unit is never visited. Executed:
   `quote('a\u{1F600}b')` → `"a\ud83db"`, which **parses**, so there is no error — the emoji is
   silently mutilated to `U+61 U+D83D U+62` and Go substitutes `U+FFFD`.
4. **UNTRACKED — `component-command.ts:18,21,24,27,33,36,39`** splice `id`, `type`, `band` raw with
   zero escaping (only `bindComponentScalarCommand:49` uses `quote()`). Worth quoting as hygiene, but
   **not document-reachable**: `internal/template/ids.go:validateElementID` (AD-10/AC34) enforces
   `^e[0-9a-z]+$` at parse time, so a crafted id cannot be loaded. `engine-protocol.ts:331` is the
   **inbound projection** guard, a different direction and a different population — see
   `## Design Notes`.
5. **THE DOCUMENT-ORIGINATED ROUTE, and the one that carries the severity — `bindComponentScalarCommand`.**
   Bind segments are JSON object **keys taken verbatim from the author's sample-data file**; nothing
   constrains a JSON key, and they pass through the broken `charCodeAt(0)` quoter at `:52-67`. Open a
   data file, click a node, press Connect: no typing, and the segment that reaches the document is not
   the one the file held. This is the leg the acceptance must exercise.

**The reachability answer for `rawNumberLiteral` itself: NO.** Traced every `PropertyIntent`
producer. `updateComponentPropertiesCommand` has exactly one production caller (`App.tsx:760` in
`applyProperties:749`, wired as `onCommit` only at `App.tsx:1478`). Only `PropertyDraft.commit()`
(`App.tsx:2103`) and `.step()` (`:2196`) reach point/ratio fields. Every hop is number-typed and
re-formatted locally: `committedValue` (`App.tsx:1805-1809`) returns `points(value)` for numbers;
`points` (`:2891`) emits only `-?\d+(\.\d{1,3})?`; `engine-protocol.ts:331/:356/:373` type every
point and ratio field as `number`. Arrow step is gated by `draftThousandths` (`:1834`). `onPaste`
exists on the **textarea only** (`:2389`), not the numeric input. **But the input is a plain text
`<input>` with `onChange={writeDraft(event.target.value)}` and no validation before send** — so DW-32
is author-reachable in one gesture (paste, blur).

### The engine — where last-wins lives

- `folio-go/component_commands.go:38` `ApplyComponentCommand(t *Template, command []byte) (CanvasProjection, error)`.
  `:42-44` `json.NewDecoder` + `UseNumber` + `var raw map[string]json.RawMessage` — **L1, last key wins.**
  `:47-48` trailing-value guard. `:50` `equalNumber(raw["version"], "1")`. `:53` kind. `:56` dispatch.
- `folio-go/component_commands.go:1302` `componentFields(raw, want)` — `len(raw) != want` on the
  **deduplicated** map. This is why a four-field injection still counts four.
- **Three object levels, and arrays hold only scalars.** L1 `raw` `:44`; L2 `changes` `:862`; L3 the
  operation object `:899` (whose own `len(value) != 1 && != 2` check is the same dedup-blind count).
  Verified: the only slice decodes are `[]string` (`ids:847`, `segments:548`, `entries:2122`,
  `edges:1251`, `tail:2585`) — zero `[]json.RawMessage`, against a population of 5 as control.
  `page_setup.go:1938` adds an L2′ `margin` object with the same defect.
- **`ids` uniqueness (`:851-856`) is a semantic rule, not duplicate-key detection** — it does nothing
  for L1/L2/L3.
- **Executed at HEAD:** L1 duplicate `ids`+`changes` → accepted, `e1` (pageHeader) named, `e5`
  (pageFooter) mutated, `err == nil`. L2 and L3 duplicates → accepted, last wins. Duplicate `op`
  (`clear` then `set`) → accepted. **Duplicate `version` (`0` then `1`) → accepted.** `createComponent`
  with `"band"` twice → placed in the second band. Kind escalation to a **different arity** is refused
  (arity coincidence only); escalation to the **same** arity (`deleteComponent` 3 → `deleteFontChain`
  3) **passes the arity check and lands in the wrong handler**, failing only later on field names.
- The wasm path adds no check either: `folio-go/wasm/engine.go:224` `json.Unmarshal(command,
  &struct{Kind string})` also routes on the **last** `"kind"`.
- **24 dispatched kinds**, arities 3–12, at `component_commands.go:147,188,217,261,329,384,433,536,
  608,709,839,1381,1437,1606,1651,1694,1758,1774,2107,2165,2202,2221,2244,2272,2360`; plus
  `ApplyPageSetupCommand` (`page_setup.go:1903`) = the 25 the epic-11-14 log names.

### The refusal channel (confirmed against source, every claim)

- `componentFailure` → `*folio.ComponentCommandError`, `component_commands.go:24-32`.
- `wasm/cmd/engine/main.go:236-244` matches it **before** `*RenderError` (`:245-254`) and emits
  `DiagnosticCode: "COMPONENT_INVALID"` `:239`, `Message: bounded(…, 512)` `:240`, `ElementID:
  bounded(…, 128)` `:241`, `DataPath: bounded(…, 256)` `:242`. `bounded` slices by **bytes**.
- `COMPONENT_INVALID` is a **host-local literal**, absent from `internal/diag/diag.go`'s 18 registered
  codes. **Minting a registry code is structurally blocked:** `diagnostic_registry_census_test.go:131`
  requires `errors.As(err, &renderErr)` for every registered code, and a `ComponentCommandError` is
  not a `*RenderError`.
- ⚠ **Every existing decode-site refusal (`:45`, `:48`, `:51`, `:54`) is a bare `fmt.Errorf`**, so it
  falls through to `ENGINE_REJECTED` with no location. The new refusal must use `componentFailure`.
- Precedent for `ElementID: ""`: `componentFailure("", "component.ids", …)` at `:844,:849,:859,:863`.
  Precedent for a document-scoped path: page setup's `"page.setup"` fallback (`main.go:256-265`).
- Module-side bounds `maxComponentFailureMessageBytes` / `maxComponentDataPathBytes` at
  `component_commands.go:1974-1975`, tied to the host's literals by
  `component_commands_test.go:1738` — which pins **Message and DataPath only, not ElementID**.

### Detecting duplicate keys — mechanism and prior art

`encoding/json` does not report duplicates. Of the three stdlib options, **`json.Decoder` token
streaming (`Token()`/`More()`) with a per-object `seen` set** is the only one that works: a
`RawMessage` re-walk is structurally blind because the duplicate is already collapsed by the time you
hold the map. **Measured cost: 2.55 µs** for a realistic property command (100k iterations in 255 ms);
27 ms for an 8 MB asset envelope. Verified to detect duplicates at `$`, `$.changes`,
`$.changes.value`, `$.list[0]` and `$.k[0][0]` — arrays included.

⚠ **CORRECTED AT IMPLEMENTATION.** This paragraph originally added *"it also subsumes the
trailing-value guard at `:47-48`"*. **It does not, as implemented**: the scan reads the first
top-level value and stops, and never inspects the bytes after it. The existing trailing-value guard
therefore stays and is still the only thing checking for a second document — it was left in place
rather than replaced, which is the correct outcome, but not for the reason the sentence gave.

**Reuse, do not invent** — the idiom already ships: `internal/bind/value.go:161-183` (Decoder +
UseNumber + `Token()` vs `io.EOF`), `internal/template/rawvalue.go:51` (canonicalising recursive
walk), `component_commands.go:519-521` (`decoder.More()`), `lint/internal/licence/graph.go:50-51`.

⚠ **`lint/internal/rules/maprange.go:35-45` totally bans `range` over a map in non-test `.go`.** A
`seen map[string]bool` used only for lookup and insert is compliant; do not range it.

### Both-sides test precedent — copy this mechanism

`folio-go/canvas_projection_wire_test.go:373` `TestCanvasProjectionWireKeysAreTheOnesTheDesignerAccepts`:
`repoRootFromTest(t)` (`fixture_test.go:32-50`, walks up to a dir holding both `folio-go/` and
`fixtures/`) → `os.ReadFile` the `.ts` → **anchored** `regexp` (`:346`) → `extractedKeyList` →
`reflect.DeepEqual` against a Go-side record. Two standing rules it carries and a copy must too: a
missing file is **`t.Fatalf`, never `t.Skip`** (`:377-380`), and a non-matching regexp is `t.Fatal`
saying *"re-derive this extraction rather than deleting the check"* (`:383`). The anchoring rule is
stated at `:366`: an unanchored match reads some other guard's list and compares it to this record —
a green test asserting the wrong thing. A closer Go→host instance is
`component_commands_test.go:1738`.

### Designer-side guards a new file or a changed line will trip

- **`engine-ownership-contract.test.ts:66` — the allowlist you MUST edit.** `:8-10` scans
  `readdirSync(recursive)` over `.ts|.tsx` minus tests, so a new module is picked up automatically;
  `:27` is an AST detector for `JSON.parse`/`JSON.stringify`; `:66` filters against
  `['engine.worker.ts','offline-lifecycle.ts','release-payload.ts','sample-data.ts','App.tsx',
  'table-column-command.ts','component-property-command.ts','font-chain-command.ts']`. **A new shared
  module calling `JSON.stringify` fails this test until it is added.** The filter is one-directional,
  so stale entries do not fail — but the trailing comment naming "the **three** command factories" is
  a factual claim the consolidation invalidates and must be rewritten. Same test also requires
  `schemaMirrors == []`: no interface/type/object literal may name ≥2 of `version, page, bands,
  elements, assets`.
- `engine-bounds-mirror.test.ts:264/:310/:318/:352/:381/:385` — the source-text guards on
  `POSITIVE_LENGTH_FIELDS`, `ORIGIN_FLOOR_FIELDS`, and `App.tsx:2150`. See **Never**.
- `canvas-authority-contract.test.ts` — the AD-17 scan; `:18`/`:191`/`:200` prohibited-regex arm with
  non-vacuity floors at `:197-199`; `:203-205` forbids engine refusal vocabulary in production source.
- `design-contract.test.ts:71` pins package + lockfile + strict compiler metadata.
- oxlint baseline is **exactly 4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`;
  `App.tsx:2896,2903`). New non-component exports go in `.ts`, **never** `.tsx`.

### Conventions the new module must follow

Flat in `folio-designer/src/` (45 non-test `.ts` sit there; the only subdirectories are domains, not
util buckets). No barrel — import by relative path. Named exports only, no default. A file-head block
comment stating the ownership boundary (universal across all six encoders). A colocated
`<name>.test.ts`. 2-space indentation (`component-command.ts`'s tabs are the outlier).

### The ledger this story owes

Per **D-000.15**, a before-the-tag exported narrowing is registered at story close as a deferred entry
owned by **Story 15.3**. Template: **DW-153** (`deferred-work.md:7131-7164`) and **DW-167** (`:7666`)
— *"not a defect and nothing to fix, a LEDGER ENTRY"*, `Status: OPEN until Story 15.3 reads it`, with
sections *What was narrowed* / *The refusing surface, exactly* / *What did NOT move, deliberately* /
*The one behaviour change visible to an existing caller* / *What discharges it*. **D-8.2.2's two-part
test** (`epic-7-8-decision-log.md:2461-2477`): (a) reachable through the module's exported API — yes;
(b) a narrowing that has not landed — yes. **Both limbs satisfied.** There is **no** API-surface
census test and no public-API inventory file in the repo (searched; 14 hits, all doc-comment prose),
and `version.go:9` (`0.0.0-dev`) is **Story 15.3's** to stamp, not this story's.

### The second exported door, and the diagnostic that reports the wrong cause

- `folio-go/page_setup.go:1903` `ApplyPageSetupCommand(t *Template, command []byte)` — decodes with
  `dec.Decode(&raw)` into `map[string]json.RawMessage`, then gates on `len(raw) != 7`, the kind, and
  `equalNumber(raw["version"], "1")`. **Both the arity gate and the version gate are duplicate-blind**,
  exactly as on the component door, because the map has already deduped.
- `folio-go/page_setup.go:1938-1940` — `json.Unmarshal(marginRaw, &margins)` into a second
  `map[string]json.RawMessage` with `len(margins) != 4`. **The nested door.** A guard proved only at
  the top level leaves this one open.
- `folio-go/page_setup.go:2021-2045` `lengthField` — the grammar, and the reason no browser rule may
  duplicate it: rejects `eE` (`:2029`), rejects more than three decimal places (`:2032`), then
  `json.Number` (`:2034-2036`, the sole *must be a finite number* site), then `parseMillipoints`, then
  the `MaxCanvasMillipoints` bound (`:2041`).
- `folio-go/page_setup.go:2046-2085` `parseMillipoints` — **four message sites, three vocabularies**:
  `:2053` *must be a number* (malformed shape), **`:2058` the MIXED branch**, `:2066` *must be a
  number* (fraction loop, non-digit), `:2076` *overflows millipoints* (genuine overflow only, must not
  move). The mixed condition is `c < '0' || c > '9' || whole > (1<<63-1)/10` — **splitting it means
  two branches, not a changed string**, because the second clause is a real overflow guard.

**Executed through the exported door** (scratch module with a `replace` onto `folio-go`, deleted
after; tree confirmed clean). A `pageSetup` command's `width` literal, and the error it returns today:

| literal | today |
|---|---|
| `null` — **what the chosen encoder emits for an unparseable draft** | `page.width: width overflows millipoints` |
| `99999999999999999999` | `page.width: width overflows millipoints` |
| a quoted `12` | `page.width: width overflows millipoints` |
| `1e3` | `page.width: width must be a decimal with at most three places` |
| `1.2345` | `page.width: width must have at most three decimal places` |
| `abc` | `page setup command is malformed` — bare `abc` is not JSON, so it never reaches `lengthField` |

⚠ **The first two rows are the red-proof pair**: indistinguishable today, they must diverge after,
with neither going silent. Note `abc` does **not** reach *must be a finite number* through this door —
that site is reached only by valid JSON that is not a number — so do not write a test asserting it does.

⚠ **A wording fork the ruling does not settle, flagged rather than decided.** `parseMillipoints`'s own
established wording for a non-digit is *must be a number* (`:2053` and `:2066`, the latter being the
fraction loop directly below the mixed branch). The ruling's prescribed text, *must be a finite
number*, is the **caller's** wording at `:2035` for a `json.Unmarshal` failure — a genuinely
different cause. **RULED at CHECKPOINT 1: use `must be a number`, matching `:2066`.** The reasoning
matters more than the choice: the fraction-digit loop at `:2066` already answers this exact question
for a non-digit, and `:2058` calls the **identical condition** something else one loop apart. That
disagreement **is** the defect. Making the whole loop say what the fraction loop already says is not
picking a vocabulary — it is removing an inconsistency that was already there. Importing `:2035`'s
wording would add a third vocabulary and leave the two digit loops still disagreeing.

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/command-json.ts` -- NEW. The single command-JSON authority: a string quoter
      (`JSON.stringify`), a number emitter modelled on `table-column-command.ts:4`, and whatever
      envelope helper lets a builder stop writing template literals with `${…}` in a JSON position.
      File-head comment stating the ownership boundary. -- One place to be right; D-8.1.3's exact shape.
- [x] `folio-designer/src/command-json.test.ts` -- NEW. Non-BMP round trip (`U+1F600` through a bind
      segment and an asset key, asserted by `JSON.parse` returning the author's code points); the
      duplicate-key injection payload; the hostile-id payload. -- DW-75 and DW-32 are invisible to
      every existing test.
- [x] `folio-designer/src/component-command.ts` -- Delete the hand-rolled `quote()` at `:52-67`;
      route strings AND the raw `id`/`type`/`band` splices at `:18,21,24,27,33,36,39` through the
      authority. -- DW-75, plus the raw id/type/band splices as hygiene; this file also carries
      the bind segments that are the story's document-originated leg.
- [x] `folio-designer/src/component-asset-command.ts` -- Delete the hand-rolled `quote()` at `:18-33`;
      route through the authority. -- DW-75, second copy.
- [x] `folio-designer/src/component-property-command.ts` -- Route `quote` and `rawNumberLiteral`
      (`:53`, `:58-62`) through the authority; a numeric draft matching the JSON number grammar
      travels verbatim, anything else becomes `null`. No `Number()` on this path.
      `POSITIVE_LENGTH_FIELDS`,
      `ORIGIN_FLOOR_FIELDS`, `pointFields` and `ratioFields` are untouched. -- DW-32.
- [x] `folio-designer/src/page-setup-command.ts` -- Route all six numeric splices AND the raw
      `preset`/`orientation` string splices at `:6-7` through the authority. -- DW-73.
- [x] `folio-designer/src/page-setup-command.test.ts` -- NEW. The only encoder with no test file. -- DW-73
      requires its acceptance to exercise this file **by name**.
- [x] `folio-designer/src/table-column-command.ts`, `folio-designer/src/font-chain-command.ts` --
      Route through the authority. `font-chain-command.ts` is already correct; `table-column-command.ts`
      is correct for strings but **two of its builders splice a raw `${index}`/`${toIndex}` past its own
      `number` helper** (corrected at implementation — the Code Map called it "the shipped model").
      Routing them is what makes soleness a real property rather than a claim about the other four of
      the six. -- DW-73's closure condition.
- [x] `folio-designer/src/engine-ownership-contract.test.ts` -- Add the authority to the `:66`
      allowlist; remove the encoders that no longer call `JSON.stringify` directly; rewrite the
      "three command factories" comment. -- The test fails otherwise.
- [x] `folio-designer/src/command-json-soleness.test.ts` -- NEW guard: **an allowlist, not a denylist** — assert
      that the set of production files building command JSON is exactly {the authority} and that no
      other production file interpolates into a JSON string or number position. Red-prove by
      reintroducing one splice. -- A guard enumerating what is forbidden cannot see what nobody
      thought to forbid.
- [x] `folio-go/component_commands.go` -- A stdlib token-streaming duplicate-key scan run on the
      command bytes before/at `:44`, refusing at **every** nesting level including objects inside
      arrays, via `componentFailure` with `ElementID: ""` and a document-scoped `DataPath`. It is
      **one shared scan called from BOTH `ApplyComponentCommand` and
      `ApplyPageSetupCommand`** (`page_setup.go:1903`), reaching the nested `margin` object at
      `:1938-1940`. -- The Go half; it is what makes the property assertable at all.
- [x] `folio-go/component_commands_test.go` -- Hand `ApplyComponentCommand` duplicate-key **bytes**
      directly at L1, L2, L3, inside an array, and on `version`; assert refusal and that the
      previously-mutated element is unchanged. Site it near `TestComponentFailureBoundsMatchTheHostsOwnLiterals`
      (`:1738`); the fixture helper is `componentTemplate(t)` (`:25-35`). -- The AC requires the test
      to hand the **engine** the bytes.
- [x] `folio-go/command_json_authority_wire_test.go` -- The one test that reads both languages: copy
      `canvas_projection_wire_test.go:373`'s mechanism (`repoRootFromTest` + **anchored** regexp,
      `t.Fatalf` on a missing file, `t.Fatal` telling a future reader to re-derive rather than delete).
      -- "They land in one commit, with a test that reads both sides" is a frozen AC.
- [x] `folio-go/wasm/engine_test.go` -- Cover `Engine.Apply`'s own last-wins `kind` read
      (`wasm/engine.go:224`). -- The wasm path routes on the last `"kind"` and adds no check.
- [x] `_bmad-output/implementation-artifacts/deferred-work.md` -- Close DW-32, DW-73, DW-75; append
      the **D-000.15 ledger entry** for Story 15.3 in DW-153's shape, covering **both** narrowings;
      replace DW-32's superseded closure sentence with the verbatim text in `## Rulings` and add the
      bind-segment finding to its severity paragraph; append any entry the rulings
      defers. -- The narrowing must be handed over, not just performed.

- [x] `folio-go/page_setup.go` -- Split the mixed branch at `:2058` into two branches: a non-digit
      reports **`must be a number`** (matching `:2066`, ruled at CHECKPOINT 1), a genuine overflow
      keeps reporting `overflows millipoints`. `:2076` is untouched.
      -- The ruling depends on `null` producing a right-cause message; today it does not.
- [x] `folio-go/page_setup_test.go` (or the existing page-setup test file) -- Assert the split **both
      directions** against the executed pair: `null` says not-a-number, `99999999999999999999` still
      says overflows. -- Fixing a message must not delete the detector.
- [x] `folio-designer/src/page-setup-command.ts` -- Also quote `preset` and `orientation`, spliced raw
      *inside quotes* at `:7`. -- `engine-protocol.ts:275` is a remote defence; a check belongs where
      its operands are.

**Acceptance Criteria:**

- Given duplicate-key **bytes** at any nesting level — top level, the `changes` object, the operation
  object, an object inside an array, or a repeated `version` — when they reach `ApplyComponentCommand`
  directly (not via the encoder), then it **refuses**, and the element a last-wins resolution would
  have mutated is provably unchanged.
- Given bytes carrying a repeated `version` (`0` then `1`), when they reach
  `ApplyComponentCommand`, then it **refuses** — asserted directly, not left to be repaired by
  accident at the top level, because a fix nothing asserts can regress silently. Today it is
  **accepted**: the `version == 1` gate reads the last key, so a `version:0` command is admitted
  by appending a second one.
- Given bytes whose duplicate resolves one command kind into a **different kind of the same
  arity** (`deleteComponent` 3 → `deleteFontChain` 3), when they reach `ApplyComponentCommand`,
  then it **refuses at the duplicate** — not at the second handler's field names, which is the
  arity coincidence that happens to stop it today and stops nothing in general.
- Given the duplicate-key refusal, when it surfaces at the wasm host, then it arrives as
  `COMPONENT_INVALID` with an **empty** `ElementID` — never naming an id the duplicate has made
  untrustworthy.
- Given a bind segment or asset key containing `U+1F600`, when the designer encodes the command, then
  `JSON.parse` of the payload returns **the author's own code points**, and a mutation reverting the
  fix reddens that assertion.
- Given DW-32's own duplicate-key payload typed into a numeric field — not `abc`, which proves only
  that malformed bytes are produced — when the property command is encoded, then it names exactly the
  selected ids and the one edited field, with no second `"ids"` or `"changes"` key. This case pins at
  least one **named literal** field that Story 12.4 never touches (`x` or `fontSize`); a table-driven
  suite derives its table from the field set rather than transcribing it, and keeps the literal-pinned
  case so the suite cannot quietly shrink when the set later moves.
- Given the authority, when the soleness guard runs, then the set of production files building command
  JSON is exactly the authority, `page-setup-command.ts` is exercised **by name**, and reintroducing a
  single splice anywhere reddens it.
- Given every command the designer emits today for valid input, when it is re-encoded through the
  authority, then the bytes are **identical** — key order and the unquoted number form included — and
  all 766 existing designer tests still pass.
- Given the Go and TypeScript halves, when they land, then they land in **one commit** with a test
  that reads both sides, and the exported narrowing is registered on D-000.15's before-the-tag ledger
  for Story 15.3.
- Given bytes carrying a duplicate key at any level of a **page setup** command — top level, the
  nested `margin` object, or a repeated `version` — when they reach `ApplyPageSetupCommand`, then it
  **refuses**. Page setup has a single `kind`, so the same-arity escalation leg is **inapplicable
  here and is stated as inapplicable**, never counted as covered.
- Given an unparseable numeric draft, when the encoder emits `null` and the engine refuses it, then
  the message is **`must be a number`** — and a genuine overflow still reports `overflows
  millipoints`, both directions red-proved against the pair the Code Map records. The second half is
  the one that silently disappears if the split is done carelessly.
- Given a sample-data file carrying a non-BMP JSON **key**, when it is accepted through
  `acceptSampleData`, the node clicked and Connect pressed, then the bind segment on the wire is the
  author's **own code points**. The typed-draft leg proves the encoder; only this document-originated
  leg proves the severity.

## Spec Change Log

### 2026-09-05 — review loop 1: the numeric path, and the second door's diagnostic code

**Triggering finding.** All three review layers and the Matrix Test Audit independently reported that
`jsonNumber = JSON.stringify(Number(v))` makes the browser a second authority on what a number is.
Executed: `1e3` reaches the wire as `1000`, `0x10` as `16`, `0b101` as `5`, `.5` as `0.5`, `007` as
`7`, `+5` as `5`, `" 12 "` as `12`, and `9007199254740993` as `9007199254740992`. At baseline every
one of those reached Go verbatim and earned a located refusal. **The change widened the accept-set of
both exported doors in a story whose purpose is to narrow them.** The signature is `1e3` being
accepted while `1e21` is still refused — the same input class splitting two ways is never a designed
behaviour, only ever an artifact.

**Root cause: a contradiction inside `<frozen-after-approval>`, authored at the plan gate.** The
Approach prescribed `JSON.stringify(Number(v))` while the Never clause forbade a browser-side rule
about what a number is. Both were frozen; the prescribed expression is what violated the prohibition.
The implementer followed the spec exactly. **This is a spec defect, not an implementation defect.**

**What was amended, by explicit human authorisation to reopen the frozen block:**
1. **Approach** — was *"numerics as `JSON.stringify(Number(v))`"*; now *"numerics passed through
   byte-for-byte when the draft is already a valid JSON number, or replaced by `null` when it is
   not"*, with `Number()` deleted from the numeric path entirely.
2. **Never** — the browser-side-rule clause now distinguishes a rule about what a number **means**
   (forbidden) from a shape check deciding whether bytes can **reach** Go's rule (required), and
   forbids narrowing that check below the JSON number grammar.
3. **Always** — a new clause: the refusal carries the diagnostic code of **the door it was raised
   at**. The `componentFailure` constraint was written when only one door was in scope, so extending
   it verbatim to page setup was a faithful reading of a clause that had not been updated. Also a new
   clause making "the accept-set is narrowed or unchanged at every input" explicit.
4. **D-D.2's prescribed expression is superseded** by the JSON-number passthrough. The ruling's own
   words are the reason: the encoder's job is *"to make bytes that reach that rule, not to duplicate
   it"* — and the set of bytes that reach the rule is exactly the set of valid JSON numbers.

**Also recorded here, previously unrecorded:** the frozen block was reopened once before, on
2026-09-05, to correct the reachability claim after `internal/template/ids.go:validateElementID` was
found to enforce `^e[0-9a-z]+$`. That renegotiation was authorised but never logged; this entry is
its record.

**Known-bad state avoided.** Shipping an encoder that silently converts `1e3` into a 1000pt width and
`0x10` into 16pt, while a story titled *a component command means exactly what it names* claims to
have narrowed both doors.

**KEEP — what worked and must survive re-derivation.**
- The duplicate-key scan, its three nesting levels, the `version` and same-arity legs, and the
  executed proof that `ApplyComponentCommand` refuses what it used to resolve last-wins.
- Routing `table-column-command.ts` and `font-chain-command.ts` through the authority even though they
  were already correct — that is what makes soleness a property rather than a claim about four files.
- The soleness guard as a true **allowlist** with a non-vacuity floor over the production corpus.
- The both-sides test's mechanism: `repoRootFromTest`, **anchored** regexps, `t.Fatalf` never
  `t.Skip`, and a failure message telling a future reader to re-derive rather than delete.
- `blankIsNotZero`'s **intent** — a blank draft must never become `0`. It is now subsumed: blank fails
  the shape check and becomes `null`. Do not reintroduce it as a separate coercion step.
- Leaving `assetBytesRequest` outside the authority: a bare key is not command JSON.
- The `parseMillipoints` split and its both-directions red-proof.
## Design Notes

**The condition the story told us to measure came back positive — at a different site.** The frozen
text says: *"if ANY path lets DOCUMENT CONTENT reach `rawNumberLiteral` … then a hostile `.folio`
mutates arbitrary components on edit, and this jumps the queue immediately. The type signature and
the projection's numeric typing were measured; not every panel path was."* Measured: for
`rawNumberLiteral` the answer is **no** — every numeric hop is `number`-typed and re-formatted through
`points()`. The same hazard **is** document-originated, but by a route neither the story nor the
register named: a **bind segment** is a JSON object key taken verbatim from the author's sample-data
file, nothing constrains a JSON key, and it travels through `component-command.ts`'s broken
`charCodeAt(0)` quoter. Open a data file, click a node, press Connect. This is **not a scope
expansion** — that encoder is already one the consolidation routes through. What changes is the
*urgency framing*: the story's "keystroke-originated and self-inflicted" characterisation is true of
DW-32 and **false of the story as a whole**.

**The falsified claim, kept visible because the lesson outlives it.** This spec asserted at the plan
gate that a crafted `component.id` in a `.folio` reached the same splice with no typing. **It does
not:** `internal/template/ids.go:validateElementID` (AD-10/AC34) enforces `^e[0-9a-z]+$` at parse
time, so the document cannot be opened. The claim rested on `engine-protocol.ts:331` having no
charset check — but that guard governs the **inbound projection** (engine to browser), which can only
ever carry ids the loader already admitted. **The falsifying fact was already in this story's own
record:** D-D.1 says, of that very file, *"that is the INBOUND projection validator — a different
direction and a different population"*. It was written about field lists, and three readers failed to
carry it across to `component.id`.

**The reusable rule: when citing a validator as ABSENT, name the direction it governs and confirm
that is the direction under discussion.** A guard cited for a direction it does not govern is the
same error class as a guard whose expectation comes from its own input — the citation looks like
evidence and is about something else.

**Why quoting is still the complete fix** for the raw `id`/`type`/`band` splices: `JSON.stringify`
escapes the quote, so any such value becomes one string the engine resolves literally. Correct
behaviour, not a workaround — it is simply hygiene rather than the severity leg.

**A stale comment worth not propagating:** `component-property-command.ts:22-27` says `x` and `y`
"are absent from it and are unbounded, negatives included." That is false — `containComponent`
(`folio-go/component_commands.go:880` → `:1912`) floors both at the band origin, which is why Story
17.4 added `ORIGIN_FLOOR_FIELDS` seventeen lines below. Do not repeat the claim in new prose; do not
rewrite the comment either (it sits inside `engine-bounds-mirror.test.ts`'s guarded region).

**Ruling D-D.1's premises were re-measured before being written into this spec**, per the standing
rule that a settled ruling resting on a false factual premise is an intent gap rather than a
constraint. All held: the four fields are in `PropertyField` and `pointFields`; `App.tsx` has **0**
occurrences of them against a population of **17** `field: '<name>'` literals (the ruling said 16 —
harmless drift, same conclusion); `component-property-command.test.ts` has **0** against **3**
`fontSize|lineSpacing` hits; the `ratioFields` ×1000 comment is verbatim at `:12-19`; the two
source-text guards and their drift-proofs are at `engine-bounds-mirror.test.ts:264/:318` and
`:352/:385`; Story 12.4 is real, at `epics.md:4162`.

## Verification

Each command's exit code is read from `$?` **directly, never through a pipe** (the shell is zsh, so
`${pipestatus[1]}`, not `PIPESTATUS`). A `cd` earlier in a compound command silently re-roots later
relative paths, and a chained `&&` that stops early looks like success — so these run as separate
steps with separate exit codes. The run's heavy-test cadence is **end of epic**, so the matrix suite
is **not** in this story's set.

**Commands:**
- `cd folio-designer && npm test` -- the **full suite, no filter**. Baseline at `e8ff6e6`:
  **55 files / 766 tests passed, exit 0.** **Measured on the delivered tree: 58 files / 804 tests
  passed, exit 0** (+3 files, +38 tests, no regressions).
- `cd folio-designer && npx tsc --noEmit` -- expected: no errors, exit 0.
- `cd folio-designer && npm run lint` -- expected: **exactly 4** `only-export-components` warnings
  (`preview/pdf-viewer.tsx:16,17`; `App.tsx:2896,2903`), 0 errors, exit 0. A fifth is a regression.
- `cd folio-go && go test -count=1 ./...` -- **measured on the delivered tree: 1950 pass / 2 fail /
  5 skip, exit 1** (+35 passing; the two fails are the sanctioned red and its parent, no third).
  Baseline at `e8ff6e6`: **1915 pass / 2 fail / 5 skip**,
  exit 1. The two fails are `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` and its enclosing
  parent — **the one sanctioned red, never to be "fixed"**. There are **no** other reds at baseline;
  any third is this story's. Use `-json` and parse test-level `Action` — package `ok` lines give no
  counts and a non-`-v` run parses as `pass=0`.
- `cd lint && go build ./...` -- exit 0 -- **separate step, separate exit code.**
- `cd lint && go vet ./...` -- exit 0 -- **separate step.**
- `cd lint && test -z "$(gofmt -l .)"` -- exit 0 and an empty listing -- **separate step.** The `lint`
  CI job broke on gofmt this week and a chained `&&` hid it.
- `gofmt -l folio-go` **run from the repo root** -- empty listing, exit 0. Running it inside
  `folio-go/` produces `lstat folio-go: no such file or directory`, which reads like "clean" and is a
  non-measurement.

**Manual checks:**
- Every new test is **red-proved by mutation**: mutate, run, confirm red, **echo the mutated line
  back** to prove the edit applied, restore with an **absolute** path, then `git status --porcelain`
  before the next probe.

## Review Triage Log

**Review loop 1, 2026-09-05.** Three layers (blind hunter, edge-case hunter, verification-gap) plus a
Matrix Test Audit, all four run against the full 186KB diff. Outcome: **1 intent_gap · 13 patch · 3
defer · 1 reject.**

**intent_gap (1) — the only one, and it was a spec defect.** `jsonNumber = JSON.stringify(Number(v))`
made the browser a second authority on what a number is: executed, `1e3` reached the wire as `1000`,
`0x10` as `16`, `9007199254740993` as `...992`, where at baseline all three travelled verbatim and
earned a located refusal. **The change widened the accept-set of both doors in a story whose purpose
is to narrow them.** The root cause was a contradiction between two clauses both inside the frozen
block — see `## Spec Change Log`. The implementer followed the spec exactly. Resolved by the JSON
number grammar passthrough; `1e3` and `1e21` are now pinned to behave identically, which is the
signature that catches a reintroduction.

**patch (13), all applied.** The most serious was the duplicate-key scanner **desynchronising its
decoder** at the depth bound — found independently by three reviewers, fixed by draining rather than
refusing (refusing would report a duplicate-key cause for what is a depth problem). Also: the
`numericFields` table derived rather than transcribed, 7 cases becoming 11; a `"value":null` probe on
the component door; `jsonObject` now throws on a duplicate key so the authority cannot emit what the
engine refuses; booleans routed through `jsonBoolean`, red-proved by a structural soleness arm rather
than a behavioural no-op; the false "subsumes the trailing-value guard" claim corrected in both
places; the positive-control figures re-measured as files at the baseline commit and reconciled; and
the host-source assertion made whitespace-insensitive so gofmt realignment cannot red it.

⚠ **Carry to the Delivery Log: the diagnostic was wrong on BOTH doors, and only one door was in the
report.** The `parseMillipoints` branch split was specified to fix page setup. Executed at baseline,
the **component** door returned `width overflows millipoints` for `"value":null`; it now returns
`width must be a number` with `DataPath component.width`. That was not predicted — it fell out
because the defect lived in shared code and had only ever been looked at through one door.

**defer (3).** DW-197, the `parseMillipoints` overflow off-by-one — **verified pre-existing at the
baseline commit** (`3c7e226:folio-go/page_setup.go:2057` carries the identical expression), so this
story neither introduced nor fixed it. DW-198, `go test ./...` failing on a tagged `folio-go/v0.1.0`
outside the monorepo, addressed to **Story 15.3 by name**, and filed with the measurement rather than
the finding: 4 cross-language source-reading files, but **79 `repoRootFromTest` call sites across 50
files** are the larger half and are entirely pre-existing. And the `App.test.tsx` header-count guard,
whose expectation is right but whose **extent** is wrong — it counts to end of file, so the next
person to append loses a gate run; filed by the coordinator, not here, to avoid two writers.

**reject (1), worth recording because it would have been costly.** A reviewer reported that
`pageSetupCommand` has no production caller. **False** — `App.tsx:12` imports it, `App.tsx:744` calls
it. Believing it would have gutted DW-73's severity and written a non-existent front door into the
register, which is a *false closure*: worse than an open entry, because it stops anyone looking again.

**Two false-zero traps bit this story, and they are different mechanisms.** Recursive `grep` returns
false zeros here, which is the known one. But `git grep` has its own: it searches **tracked files
only**, so a newly created file is invisible — and those are exactly the files a story just wrote.
`git grep --untracked` when the question could involve new files, and say which you used.

## Suggested Review Order

**The encoder's authority — start here**

- The whole design in four lines: shape-test the draft, pass it through, never rewrite it.
  [`command-json.ts:82`](../../folio-designer/src/command-json.ts#L82)

- Why no `Number()`: the accept-set must narrow or stay level, never widen.
  [`command-json.ts:88`](../../folio-designer/src/command-json.ts#L88)

- The authority is made incapable of emitting what the engine now refuses.
  [`command-json.ts:104`](../../folio-designer/src/command-json.ts#L104)

**The engine's refusal — the half that makes the property assertable**

- One scan, two doors; the door decides the diagnostic code.
  [`component_commands.go:91`](../../folio-go/component_commands.go#L91)

- Walks arrays and nested objects, so no level is a special case.
  [`component_commands.go:115`](../../folio-go/component_commands.go#L115)

- Drains rather than refuses at the depth bound — a depth problem is not a duplicate.
  [`component_commands.go:179`](../../folio-go/component_commands.go#L179)

- The component door, narrowed.
  [`component_commands.go:206`](../../folio-go/component_commands.go#L206)

- The second door, narrowed — the amendment the plan gate missed.
  [`page_setup.go:1914`](../../folio-go/page_setup.go#L1914)

- The split branch: a non-digit is not an overflow, and both must stay distinguishable.
  [`page_setup.go:2063`](../../folio-go/page_setup.go#L2063)

**The encoders routed through it**

- The hand-rolled quoter that corrupted non-BMP text is gone.
  [`component-command.ts:5`](../../folio-designer/src/component-command.ts#L5)

- The second front door, including `preset` and `orientation`.
  [`page-setup-command.ts:11`](../../folio-designer/src/page-setup-command.ts#L11)

**Guards — what stops it regressing**

- Soleness as an allowlist, not a denylist: what nobody thought to forbid still fails.
  [`command-json-soleness.test.ts:80`](../../folio-designer/src/command-json-soleness.test.ts#L80)

- A Go test reading TypeScript source, anchored, fatal rather than skipped.
  [`command_json_authority_wire_test.go:60`](../../folio-go/command_json_authority_wire_test.go#L60)

- The document-originated leg: a non-BMP JSON key, through the real gesture.
  [`App.test.tsx:4146`](../../folio-designer/src/App.test.tsx#L4146)
