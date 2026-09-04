---
title: 'Story 17.3: Size and leading show the value they use'
type: 'feature'
created: '2026-09-04'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '85f213f'
context: []
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Font size and line spacing show their defaults as grey placeholder text — 12 and 1 — because the document
does not say anything about them and the box is genuinely empty. An author reading the panel cannot tell
whether the text is 12pt or whether nothing has been decided.

After this story both boxes carry the real number, and committing it writes it into the document.

One thing this story must not do is rewrite a file just because someone opened it.

<intent-contract>

## Intent

**Problem:** `App.tsx:1714` renders the size field with `empty: points(defaultFontSize)`, and `empty`
becomes `placeholder` (`:1835`). The value is `draft`, which starts from `committed` — `''` for an unset
field. Line spacing is the same shape (`:1651`, `empty: '1'`).

**Approach:** Show the effective value in both boxes, and write it to the document when the author
commits the field.

## Boundaries & Constraints

**Always:**
- **OPENING A DOCUMENT MAY NEVER MUTATE IT.** This is the story's safety property and it is assumed here
  rather than asked, because the alternative is every existing `.folio` silently rewriting itself on
  being looked at. **The value is written when the author COMMITS the field — never on load, never on
  selection, never on render.**
- **THE 24 GOLDEN DIGESTS MUST BE UNMOVED.** They carry their own `input.folio` and the designer does not
  re-save them, so movement means this story reached somewhere it should not have. **They are the proof.**
- **LEADING'S DEFAULT MUST COME FROM THE ENGINE, NOT FROM THIS FILE.** `lineSpacingField` (`:1651`)
  hard-codes `empty: '1'` while the engine owns the number (`folio-go/render.go:43`,
  `template.LineSpacingUnit = 1000`). Size already does this correctly — `defaultFontSize` is on the
  projection (`engine-protocol.ts:160`, `:271`) and comes from `render.go:36`. **Promoting a hard-coded
  string into a displayed VALUE would mint a second authority on a number the engine owns**, which is
  the defect this codebase has refused repeatedly. **Add `defaultLineSpacing` to the projection.**
- **THE UNSET/SET DISTINCTION IS REAL AND REACHES THE FILE.** Go's clear arm stores the zero `Presence`
  (`component_commands.go:1021-1027`, `:1209-1214`), so a cleared key is **omitted**
  (`internal/template/presence.go:16-20`). Clearing must still omit; this story adds a way to set the
  default explicitly, it does not remove the way to unset.
- **A DOCUMENT THAT SETS THE DEFAULT STOPS TRACKING IT.** If the engine's default ever moves, such a
  document keeps 12. The owner chose this knowingly over display-only. **Say it in the Delivery Log.**
- **The wire unit trap:** `lineSpacing` sends `1.5`, not `1500` (`component-property-command.ts:8-16`);
  `fontSize` is a point literal passed through unquoted and parsed by Go alone (`:29-33`).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Writing the default anywhere other than on an author's commit of that field.
- Changing what `clear` does, or removing the `×` control.
- Any field beyond `fontSize` and `lineSpacing`.

**Never:** a mutation caused by opening, loading or selecting · a second authority on a default the
engine owns · a moved golden digest · a value in the box that differs from what the canvas paints.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open a document with no `fontSize` | Any component | Box reads `12`; **the document is unchanged on disk** | — |
| Open a document with no `lineSpacing` | Any component | Box reads `1`, sourced from the engine's projection | — |
| Commit the shown default unchanged | Box reads `12`, author commits | `fontSize: 12` is written | — |
| Commit a changed value | `12` → `14` | `fontSize: 14` written, as today | — |
| Clear the field | `×`, or empty the box | Key **omitted** from the document, as today; box returns to `12` | — |
| Mixed selection | Two components, different sizes | `Mixed` as today; **no default is written to either** | — |
| A component that sets a value | `fontSize: 9` | Box reads `9`, not `12` | — |
| Golden corpus | The 24 fixtures | **Digests unmoved** | Movement is a defect, not a re-baseline |
| Engine refuses | Out-of-range leading | Existing `role="alert"`; box keeps the author's text | Existing error path |

</intent-contract>

## Code Map

Anchors at `85f213f`. **Re-verify — these are long lines.**

- `App.tsx:1646` `fontSizeField`; `:1651` `lineSpacingField` with the **hard-coded `empty: '1'`**.
- `App.tsx:1714` — `draftFor({ ...fontSizeField, empty: points(defaultFontSize) })` and `draftFor(lineSpacingField)`.
- `App.tsx:1835` — `placeholder: same ? empty : 'Mixed'`, `value: live ?? draft`. **The one-line change
  in appearance, and the whole risk is what it implies about the document.**
- `App.tsx:1750-1753` `committedValue`; `:2335` `points()` (÷1000, 3-dp, trailing zeros stripped).
- `App.tsx:1780` `commit` — sends `clear` when the box is emptied; `:1838` the `×` button.
- `engine-protocol.ts:160`, `:271` — `defaultFontSize` on the projection. **The shape to copy for leading.**
- `folio-go/render.go:36` `defaultFontSizePt = 12000`; `:43` `defaultLineSpacing = template.LineSpacingUnit`;
  `internal/template/linespacing.go:28` `LineSpacingUnit = 1000`, `:53-54` the legal range.
- `folio-go/page_setup.go:768` — where `DefaultFontSize` is emitted. **Leading joins it here.**
- `fixtures/*/expected.json` — **24** golden digests.

### RE-MEASURED AT THE DISPATCH BASELINE `8db26e0` (2026-09-04) — USE THESE, NOT THE LINE NUMBERS ABOVE

The Code Map above was measured at `85f213f`. **Three commits have landed since** (Stories 17.2 and
17.4), adding +154 lines to `App.tsx`, so every `App.tsx` anchor above is stale by ~+14 lines. Re-measured:

- `App.tsx:1660` `fontSizeField`; `:1665` `lineSpacingField` with the hard-coded `empty: '1'`.
- `App.tsx:1728` — the render site, `draftFor({ ...fontSizeField, empty: points(defaultFontSize) })` and `draftFor(lineSpacingField)`.
- `App.tsx:1975` — `placeholder: same ? empty : 'Mixed'`, `value: live ?? draft`.
- `App.tsx:1764` `committedValue`; `:2475` `points()`; `:1843` `commit()`; `:1804` `useState(committed)`.
- `App.tsx:1857` `stepThousandths` — **leading steps by 0.1, a point field by 1.**
- `engine-protocol.ts:160` (type), `:269` (`hasOnly` list), `:271` (integer validator) — **unchanged, exact.**
- `folio-go/render.go:36`, `:43` — **unchanged, exact.**

**A FIFTH PROTOCOL SITE THE CODE MAP ABOVE OMITS.** Adding `defaultLineSpacing` to the projection is a
protocol change across **five** files that must move in **ONE commit** — a protocol field that lands in
two commits is a protocol that is briefly wrong:
1. `folio-go/page_setup.go:373` — the struct field and its `json:` tag.
2. `folio-go/page_setup.go:768` — the `CanvasProjection{...}` construction.
3. `folio-designer/src/engine-protocol.ts:160` — the type.
4. `folio-designer/src/engine-protocol.ts:269` `hasOnly` **AND** `:271` the integer-key validator.
5. `folio-go/canvas_projection_wire_test.go:47` `canvasProjectionWireKeys` — the recorded sorted wire key
   set. **It reds until updated deliberately**, and it is the tripwire that catches a half-landed protocol
   change. Add `"defaultLineSpacing"` in sorted position.

## Tasks & Acceptance

**Execution:**
- [x] `folio-go` — put the default line spacing on the projection beside `defaultFontSize`, and extend
      `engine-protocol.ts`'s type and validator to match.
- [x] `App.tsx` — both boxes show the effective value; **remove the hard-coded `'1'`**.
- [x] `App.tsx` — committing the shown value writes it; clearing still omits.
- [x] **Prove no mutation on open** — a test that loads a document with neither key, asserts the boxes
      read the defaults, and asserts **no command was sent**.
- [x] Tests — one per matrix row.
- [x] Run the fixture corpus and report **unmoved, per fixture**.
- [x] **`committed` MUST STAY THE DOCUMENT'S OWN VALUE (`''` when unset); only the DRAFT initialises from
      the effective default.** `commit()` (`App.tsx:1843`) is `if (draft !== committed)`. Folding the
      default into `committed` makes that comparison false, so **committing the shown default would send
      NOTHING while every gate stayed green** — silently breaking matrix row 3. This is a boundary.
- [x] **SPLIT STORY 17.4's ONE-PREDICATE GUARD INTO TWO ARMS** (ruled 2026-09-04, see Design Notes):
      **drop** the unset arm — an unset `fontSize`/`lineSpacing` box now steps from its shown default;
      **keep** the mixed-selection arm — stepping a mixed selection still flattens every component onto
      one value, and nothing in this story touches that. A mixed field the author has TYPED into still
      steps, exactly as 17.4 has it.
- [x] Rewrite `App.test.tsx:2547` `does nothing on an UNSET field, and sends no command` to the new
      behaviour, and **keep a test that still holds the mixed case**.
- [x] **Correct the now-false prose IN THE SAME COMMIT**: `App.tsx:1883-1885` and
      `App.test.tsx:2542-2546` both assert a rule the code will no longer follow. A comment is not a
      measurement, and a stale one is worse than none.
- [x] **ADD AN ASSERTION THAT ARROW AND TYPING AGREE ON AN UNSET FIELD**, mirroring Story 17.4's
      `expect(stepped).toEqual(typed)` shape. **This is the property the ruling rests on**, so it must be
      the thing that reddens if someone reverses it.

**Acceptance Criteria:**
- Given a document with neither key, when it opens, then both boxes show the engine's defaults and
  **no command has been sent**.
- Given the shown default is committed, when the document is saved, then the key is present.
- Given the field is cleared, when the document is saved, then the key is **absent**.
- Given the 24 fixtures, when they run, then **every digest is unmoved**.
- Given a document with neither key, when the author presses ArrowUp in either box, then the field steps
  from the shown default and sends the same command that TYPING that stepped value would send.
- Given a MIXED selection the author has not typed into, when an arrow is pressed, then **nothing steps
  and no command is sent**.

## Design Notes

**The interesting half of this story is Go, not React.** Showing a number is one line. Sourcing leading's
number from the engine is what stops the designer becoming a second place that knows what the default is
— and this codebase has twice shipped a defect of exactly that shape, most recently the browser header
whose sentence disagreed with the list beside it.

## Orchestrator Ruling — 2026-09-04: Story 17.4's unset-arrow guard retires, its mixed arm survives

**Ruled by the ORCHESTRATOR, not by the owner. Flagged to the owner in the report.**

**What was found.** This spec was measured at `85f213f`. Story 17.4 landed afterwards and shipped a guard
asserting that an UNSET field does not step from the arrow keys and sends no command. This story removes
the precondition that guard rests on: once the box carries the effective value, the draft parses and the
arrow steps. Measured by probe at `8db26e0`, not predicted — one ArrowUp on a document setting neither key
sends `{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":1.1}}}`.

**The ruling: drop the unset arm, keep the mixed arm.**

*Why not preserve the guard.* On 17.4's band-edge question the orchestrator ruled that **the arrow and the
keyboard are the same act**, carried by a single assertion — `expect(stepped).toEqual(typed)` over two
independent drives of the same field. That is now a shipped property of this control. Preserving the guard
breaks it: typing `1.1` into the Line spacing box would write `1.1`, while ArrowUp on the same visible `1`
would write nothing — same field, same visible state, same author intent, two outcomes depending on which
key was used. That is exactly the special case the band-edge test exists to catch, arriving through a door
the test does not watch. A principle shipped in one story is not carved out in the next.

*On 17.4's second rationale* — "a keypress the author reads as a nudge turns an inherited value into a
pinned one". True, and the strongest argument the other way. But an author pressing ArrowUp on leading is
not looking, they are CHANGING it; there is no way to increase leading without setting it. Pinning is not a
side effect of that gesture, it IS the gesture. Contrast the safety property this story is actually built
on — **opening a document must not mutate it** — which holds absolutely, because opening carries no intent.
An arrow press does.

*Why this is a guard retiring, not a decision being overturned.* 17.4's own stated precondition was "an
unset field has no value to step, and its placeholder is not one." After this story it HAS a value, sourced
from the engine's projection. The guard was correct for the world it shipped into and is obsolete in the
world this story creates.

*What survives.* 17.4 closed both rows with one predicate and that was right then. **Only ONE of the two
reasons dissolved.** Stepping a mixed selection still flattens every component onto one value — a
destructive edit fired by a nudge key — and nothing here touches that.

**Record this reversal in the story's Delivery Log and in the sprint-status comment for 17.4**, since 17.4
is the story whose decision moved. Do NOT write it into `epic-16-decision-log.md`; that log is Epic 16's.

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong.
- `cd folio-go && go test -count=1 ./...` — **this story touches Go.** The failing leaf set must stay
  exactly `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`; **anything else moving is this story's.**
  **`gofmt -l .` must be empty** — an unformatted file took the whole job down this week.
- `cd lint && go test -count=1 ./...` — `rc=0`.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.**
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`.

**A BROWSER RUN.** Photograph the typography panel on a document that sets neither key: both boxes carry
real numbers. `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (DW-180).

**ADDED 2026-09-04 (dispatch preflight).**
- **DIFF FROM `8db26e0`, NOT FROM THE FRONTMATTER.** `baseline_commit: '85f213f'` is preserved per the
  workflow's never-overwrite rule, but it is the PLAN-GATE commit. Diffing from it sweeps Stories 17.2 and
  17.4 (1887 lines) in as this story's.
- **`gofmt -l .` must print NOTHING**, run from `folio-go`. Read `$?` directly; never through a pipe.
- Every Go gate runs with **`-count=1`** in BOTH modules — a cached `ok` here is not a measurement.
- **Red-prove every new guard**, and mutate by DELETION where an arm's reachability is in question.
- The mixed-selection arm must have a test that **still reds if the arm is deleted**.

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**9 rows**);
state the population beside every zero, pair every absence claim with a positive control; **a reported
digest names its algorithm and byte range**; **a comment is not a measurement**; **use `/usr/bin/grep`**.

## Delivery Log

### What was built

**The interesting half really was Go.** `defaultLineSpacing` now travels on the canvas projection
beside `defaultFontSize`, and the designer's line-spacing field no longer spells `'1'` for itself.
The protocol change moved across all **five** sites in one edit, exactly as the Code Map's
fifth-site note demanded: the struct field and `json:` tag (`folio-go/page_setup.go:374-383`), the
`CanvasProjection{...}` construction (`:778`), the TypeScript type
(`folio-designer/src/engine-protocol.ts:166`), `hasOnly` **and** the positive-integer key list
(`:275`, `:277`), and the recorded wire key set (`folio-go/canvas_projection_wire_test.go:57`).

In the panel, `FieldSpec` gained one flag, `shown`, which says the field's `empty` string is not
placeholder chrome standing in for behaviour but **the engine's own effective value**, so the box
carries it as real text. Only `fontSize` and `lineSpacing` set it, and both take their string from
the projection at the render site — never from a literal in `App.tsx`.

### The boundary the spec named, and how it is held

`committed` **stays the document's own value** (`''` when the key is absent). The default lands in
the **draft** only, through one helper (`inherited`), so `commit()`'s `if (draft !== committed)`
still fires when the author commits the shown default. The spec predicted this would be the failure
that stayed green; **it was mutation-proved, not reasoned about** — folding the default into
`committed` reds exactly three tests, two of them the "writes the shown default" rows.

`inherited` deliberately does **not** fill a MIXED draft. Deleting its `same &&` arm reds one test.

### The consequence the owner chose knowingly

**A document that commits the default stops tracking it.** Once `fontSize: 12` is in the file, that
element is 12pt for good; if the engine's `defaultFontSizePt` ever moves, this document keeps 12
while a document that never committed follows the engine. That is the price of writing the value
rather than displaying it, and the owner took it over display-only.

### One behaviour change not in the matrix, flagged

**`×` now appears on the size and leading rows even when the key is absent**, because `canClear`
keys off the box having a value and the box now always has one. `clear` itself is untouched — it
still sends `op:"clear"`, Go still stores the zero `Presence`, and the key is still omitted from the
file — but pressing `×` on an already-unset field now sends a redundant, idempotent clear command.
The row must not go blank behind it, which is what the accepted-command reconciliation now
guarantees; `clears an ALREADY-ABSENT key without blanking the row` reds if that wrapper is removed.
Removing the control was on the spec's **Ask First** list, so it was not removed.

### Story 17.4's guard: unset arm retired, mixed arm kept

Applied as the orchestrator ruled. `App.test.tsx`'s `does nothing on an UNSET field, and sends no
command` is rewritten to `steps an UNSET field from the value the box shows, and sends what typing
that value would send`, carrying `expect(stepped).toEqual(typed)` over two independent drives —
the property the ruling rests on, so it is the thing that reddens if someone reverses it.
Reinstating the guard reds three tests. The false prose in `App.tsx`'s `step()` and above the test
was corrected in the same edit; a stale comment is worse than none.

A **positive control** was added for the retirement's scope: `Border width (pt)` has no projected
default, so its draft is still empty and its arrow still does nothing.

### Verification actually run

- `cd folio-go && go test -count=1 ./...` → **rc=1**, failing leaf set **exactly**
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` (2 `FAIL:` lines, parent + leaf; the
  permitted one). `gofmt -l .` → **rc=0, printed nothing**.
- `cd lint && go test -count=1 ./...` → **rc=0**.
- `cd folio-designer && npm test` → 708 tests, **1 failure, matched by NAME**:
  `canvas-authority-contract.test.ts` → `scans a non-vacuous production, unit-test, and e2e corpus
  for browser measurement authority` (DW-152, permanent).
- `npm run typecheck` **rc=0** · `npx oxlint` **rc=0, exactly 4 warnings** · `npm run build`
  **rc=0** · `npm run test:e2e:compile` **rc=0**.
- **Golden corpus: 24 digests, 24 unmoved.** `git diff --stat -- fixtures` printed **nothing**
  (rc=0), and the fixture tests recompute and re-compare under `-count=1` — 76 `--- PASS` lines
  across the fixture run, rc=0. The digest is **sha256 over the whole produced PDF byte stream**
  (`crypto/sha256.Sum256` on the rendered bytes, `fixture_test.go:325`), recorded as the `sha256`
  field of each `fixtures/<name>/expected.json`.
- **A browser run was taken.** Chromium `chromium-1217` (Chrome for Testing, 336M — `chromium-1208`
  is the 428K stub, DW-180) via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH`, against the real `vite
  preview` build. On a placed text element committing neither key, the typography panel reads
  **Font size = `"12"` and Line spacing = `"1"` as INPUT VALUES**, not placeholders, while Colour
  beside them still shows `black` as grey placeholder chrome — the visual proof the change is
  scoped. The engine snapshot stayed at `REVISION 2` (the placement) across an idle wait: selecting
  and rendering the panel sent nothing.

### Red-proofs (every new guard, by deletion or reversal)

| Mutation | Reds |
|---|---|
| Delete `same &&` from `inherited` (mixed protection) | `leaves a MIXED selection MIXED` |
| Fold the default into `committed` | the 2 "writes the shown default" rows + the empty-box row |
| Reinstate 17.4's unset-arrow guard | the arrow/typing agreement row + 2 more |
| Drop `inherited()` from the accepted-command reconcile | `clears an ALREADY-ABSENT key without blanking the row` |
| Drop `inherited()` from the committed transition | `CLEARS to an omitted key, and the box returns to the engine default` |
| Drop the `else` arm of `commit()` | `sends NOTHING when the author empties an already-unset box` |
| Re-mint the hard-coded `'1'` at the render site | `reads BOTH defaults off the projection and not off a constant in the panel` |
| Drop `shown` from the font-size render site | 9 tests |
| Drop `inherited()` from `revert` (Escape) | `reverts on Escape to the inherited value rather than to nothing` |
| Go: `DefaultLineSpacing: 999` | `TestCanvasProjectionCarriesTheDefaultLineSpacing` |
| Go: remove `"defaultLineSpacing"` from the recorded wire keys | both `TestCanvasProjectionWireKeys…` tests, one of which reads `engine-protocol.ts` |

### Matrix audit — 9 rows, 9 results

| # | Row | Result | Where |
|---|---|---|---|
| 1 | Open with no `fontSize` → box `12`, document unchanged | **PASS** | `OPENS A DOCUMENT THAT SETS NEITHER KEY…` (asserts `request` never called) |
| 2 | Open with no `lineSpacing` → box `1`, from the projection | **PASS** | same test + `reads BOTH defaults off the projection…` |
| 3 | Commit the shown default unchanged → `fontSize: 12` written | **PASS** | `WRITES THE SHOWN DEFAULT…` / `…SHOWN LEADING…` |
| 4 | Commit a changed value `12`→`14` | **PASS** | `commits a CHANGED value exactly as it did before` |
| 5 | Clear → key omitted, box returns to `12` | **PASS** | `CLEARS to an omitted key…`, `clears an ALREADY-ABSENT key…`, `sends NOTHING when the author empties…` |
| 6 | Mixed selection → `Mixed`, no default written to either | **PASS** | `leaves a MIXED selection MIXED…` |
| 7 | A component that sets `fontSize: 9` → box reads `9` | **PASS** | `shows the COMPONENT'S OWN value…` |
| 8 | Golden corpus → 24 digests unmoved | **PASS** | `git diff --stat -- fixtures` empty; fixture suite green under `-count=1` |
| 9 | Engine refuses → existing `role="alert"`, box keeps the author's text | **PASS** | `keeps the author's text on the existing refusal path` |

### Not done

The work is **uncommitted, by instruction** — every change is left unstaged in the working tree for
review. `_bmad-output/implementation-artifacts/sprint-status.yaml` still reads `in-progress` for
17.3; advancing it is the closer's job, not this run's.
