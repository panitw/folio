---
title: 'Story 17.4: Arrow keys step a number field'
type: 'feature'
created: '2026-09-04'
status: 'done'
review_loop_iteration: 0
baseline_commit: '85f213f'
context: []
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Every number box in the inspector ignores the up and down arrows. Nudging a width by a point means
selecting the text and retyping it.

After this story the arrows step the value, the way they do in every other design tool.

<intent-contract>

## Intent

**Problem:** `App.tsx:1796-1799`'s `keyDown` handles exactly two keys — `Enter` (single-line only) and
`Escape`. Measured: within `PropertyDraft` (`:1755-1839`) `event.key` occurs only on those two lines,
against a positive control of matches at `:1455` and `:2377` elsewhere in the file.

**Approach:** Add up/down stepping to the numeric fields, in the field's own unit.

## Boundaries & Constraints

**Always:**
- **THE STEP IS DERIVED FROM THE FIELD, NOT A CONSTANT.** The smallest representable increment is
  0.001 for every one of these — points are stored in **thousandths** (`App.tsx:2335` `points()` divides
  by 1000; `internal/template/decimal.go:30-58` refuses a fourth decimal place). **0.001 is the floor of
  the representation, not a useful step.** A point field steps by 1; leading is a dimensionless ratio and
  steps by 0.1. **State the chosen steps in the spec's own terms and test them.**
- **THE NUMERIC SET IS THE ONE THE CONTROL ALREADY KNOWS.** `App.tsx:1838` computes
  `inputMode={unit === 'pt' || unit === undefined && field === 'lineSpacing' ? 'decimal' : undefined}`.
  That is `x`, `y`, `width`, `height`, `fontSize`, `borderWidth`, `lineSpacing`. **Reuse that predicate;
  do not mint a second list that can drift from it.** Measured: there is **no `mm` unit anywhere** —
  `\bmm\b` returns 0 lines against 162 substring hits inside `command`/`commit`/`comment`, and a
  positive control of 4 for `unit: 'pt'`.
- **AN ARROW MAY NOT PRODUCE A VALUE THE ENGINE WILL REFUSE.** `width`, `height`, `fontSize` and
  `borderWidth` must stay `> 0` (`component_commands.go:1005-1007`); `lineSpacing` is bounded to
  0.001–1000 (`internal/template/linespacing.go:53-54`, `:61-71`). **Stepping down at the floor must
  stop, not underflow into a refusal the author did not ask for.**
- **THE TYPED LITERAL IS PASSED THROUGH UNQUOTED AND GO ALONE PARSES IT**
  (`component-property-command.ts:29-33`). A step must therefore produce a literal with **at most three
  decimal places** — floating-point arithmetic here can emit `13.999999999999998`, which Go refuses.
  **Do the arithmetic in integer thousandths.**
- **THE GEOMETRY FIELDS ARE `readOnly` DURING A DRAG** (`App.tsx:1835`, `readOnly: live !== undefined`).
  **An arrow must do nothing then**, exactly as typing does nothing.
- **An empty box has no value to step.** Define what an arrow does on an unset field and assert it.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Modifier behaviour (Shift for a coarse step, Alt for a fine one) — **not in scope unless asked for.**
- Applying stepping to any non-numeric field, or to the prose field.
- Changing `inputMode`, or introducing `<input type="number">`.

**Never:** a value with more than three decimals · a step past a bound into a refusal · an arrow that
fires while `readOnly` · a second list of numeric fields · float arithmetic on the value.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Up on a point field | `width` = `12` | Becomes `13` | — |
| Down on a point field | `width` = `12` | Becomes `11` | — |
| Up on leading | `lineSpacing` = `1` | Becomes `1.1`, **not `1.1000000000000001`** | — |
| Down at the floor | `width` = `1`, step would reach `0` | **Stops at the smallest legal value**; no refusal | Never sends an illegal value |
| Down at leading's floor | `lineSpacing` = `0.001` | Stops; no refusal | — |
| Arrow past the band edge | `width` or `x` at the content band's edge | **The value IS sent.** The browser does not clamp here: this bound is a property of the LAYOUT, not of the field | The engine refuses with `component.geometry: folio: component geometry must stay within <band>` and the existing `role="alert"` path renders it — **identical to typing the same value**, which is the property asserted |
| Arrow on an unset field | Box empty, placeholder showing | **Defined and asserted** behaviour | — |
| Arrow during a drag | `live` set, field `readOnly` | **Nothing happens**; no command | — |
| Arrow on a non-numeric field | `Text`, `Border colour` | Unhandled — the browser's own caret behaviour | — |
| Arrow on a mixed selection | Two components, different widths | **Defined and asserted** | — |
| Enter and Escape | — | **Unchanged** — still commit and revert | — |

</intent-contract>

## Code Map

**RE-ANCHORED AT `8bfcec8`, the dispatch HEAD.** The spec was written at `85f213f`; `8bfcec8`
(Story 17.2) inserted a 14-line comment at `App.tsx:1452`, so **every `App.tsx` anchor below the
panel moved by exactly +14**. Each line below was re-measured individually, not shifted arithmetically.
The frozen contract above still quotes the `85f213f` numbers; it is read-only and was not edited.

- `App.tsx:1810-1813` (was 1796-1799) — `keyDown`. **The seam; two keys today.**
- `App.tsx:1852` (was 1838) — the `inputMode` predicate. **The numeric set, already computed once.**
- `App.tsx:1849` (was 1835) — `shared`, carrying `readOnly: live !== undefined` and `value: live ?? draft`.
- `App.tsx:1785-1794` (was 1771-1780) — `submit` / `commit`; `commit` fires only when `draft !== committed`.
  Note `submit` also sets `pending`, and `shared` carries `disabled: pending` — see Design Notes.
- `App.tsx:1782` — `if (lastCommitted !== committed) setDraft(committed)`, the panel following the
  engine on a committed transition. **A committed step re-enters the field through here.**
- `App.tsx:2349` (was 2335) `points()` — ÷1000, 3-dp, trailing zeros stripped. **The formatter to
  round-trip through.** It is a hoisted function declaration, so `PropertyDraft` may call it.
- `App.tsx:1630-1631` (was 1616-1617) `x`/`y`/`width`/`height`; `:1660` (was 1646) `fontSize`;
  `:1665` (was 1651) `lineSpacing`; `:1675` (was 1661) `borderWidth`.
- `App.tsx:1625` — `FieldSpec`, which carries the `unit` the step size is derived from.
- `component-property-command.ts:7` (was 6) `pointFields`, `:17` (was 16) `ratioFields`,
  `:28-32` (was 29-33) `rawNumberLiteral` ("Preserve the typed literal, unquoted; Go alone decides
  whether it is a valid number").
- `folio-go/internal/template/decimal.go:30-68` (was 30-58) `decodePoints`; **the three-decimal
  refusal is `:55`**, verbatim `value %q has more than three decimal places`.
- `folio-go/component_commands.go:1006` (was 1005-1007) — the `> 0` rule, enumerating exactly
  `width`, `height`, `fontSize`, `borderWidth`. **`x` and `y` are absent from it and are unbounded.**
- `folio-go/internal/template/linespacing.go:52-55` — `MinLineSpacingThousandths = 1`,
  `MaxLineSpacingThousandths = 1000000`; the predicate is `:61-71`. **In ratio terms: 0.001 to 1000.**

## Tasks & Acceptance

**Execution:**
- [x] `App.tsx` — step the draft on ArrowUp/ArrowDown for the numeric set, **in integer thousandths**,
      formatted back through `points()`.
- [x] Clamp at each field's real bound rather than sending a value the engine refuses. Every bound that
      is a property of the FIELD is clamped: the `> 0` rule, leading's minimum and maximum, and the band
      ORIGIN for `x`/`y` (found at review — `containComponent` refuses negative geometry on this same
      path). The band EDGE is deliberately NOT clamped, by ruling: it is a property of the LAYOUT, so
      clamping it would put a second, drifting copy of the engine's containment rule in the browser.
      That case sends, is refused, and renders the engine's own alert — asserted to be identical to
      typing the same value. See Spec Change Log item 9.
- [x] Do nothing while `readOnly`; leave Enter and Escape untouched.
- [x] Decide and assert the unset-field and mixed-selection behaviours. Both do nothing and send no
      command, closed by ONE predicate (the draft does not parse). Rationale in Design Notes.
- [x] Tests — one per matrix row (**11 rows, 11 results**), plus four the matrix does not reach: the
      leading ceiling, the non-disabling step, the origin floor, and the modified arrow. The float row
      steps FIVE times, because one step from `1` does not discriminate — see Design Notes.
- [x] A browser run: chromium-1217, width `72 -> 84` over twelve presses with the canvas
      `--component-width` following every one and focus never leaving the field; five down to `79`;
      `Shift+ArrowUp` inert; no `role="alert"` and no page errors. Screenshots in the run scratchpad.

**Acceptance Criteria:**
- Given `width` = `12`, when ArrowUp is pressed, then the box reads `13` and the committed value follows.
- Given `lineSpacing` = `1`, when ArrowUp is pressed, then the box reads exactly `1.1`.
- Given a field at its smallest legal value, when ArrowDown is pressed, then **no illegal value is sent**.
- Given a drag in progress, when an arrow is pressed, then **nothing happens and no command is sent**.

## Design Notes

**The trap here is arithmetic, not keys.** Every value in this control is a decimal string that Go parses
exactly and refuses beyond three places, and it travels UNQUOTED — so a literal carrying a fourth decimal
place is refused by `decimal.go`, a failure that would look to the author like a mysterious rejection of
an arrow press. **Stepping in integer thousandths and formatting through `points()` is the whole design.**

**⚠ THE EXAMPLE THIS SPEC ORIGINALLY GAVE FOR THAT HAZARD WAS FALSE — and it was the example the story's
own test was written against.** The frozen matrix row and this paragraph both cited `1 + 0.1` as
producing `1.1000000000000001`. **Measured in node: `(1 + 0.1) === 1.1` is `true`, and `String(1 + 0.1)`
is `"1.1"`.** A single step up from `1` therefore survives a fully floating-point implementation and
proves nothing at all. The hazard is real, but it begins one step later, once the error starts
accumulating: `1.1 + 0.1` is `1.2000000000000002`, then `1.3000000000000003`, then `1.4000000000000004`
— and `decimal.go` refuses every one of those.

The CONCLUSION is untouched and stands on its own evidence: do the arithmetic in integer thousandths.
Only the witness value was wrong. **The consequence for testing is the part worth keeping**: the test
covering this row steps FIVE times, not once. The single-step version was written first, and mutation
proved it worthless — the entire step path was rewritten as `Number(draft) + 0.1` and every test in the
describe stayed green. Nothing short of mutating a passing test was going to surface that.

**The frozen matrix row still reads "not `1.1000000000000001`".** Its EXPECTATION (`1.1`) is correct, so
the row was left alone rather than amended; only its parenthetical witness is wrong, and this paragraph
is the correction of record.

There is no `Number()`, no `parseFloat`, and no `+` on a value anywhere on this path. The draft string is
read INTO integer thousandths by an exact decimal parser, stepped as an integer, clamped as an integer,
and written back out by `points()`. A draft the exact parser cannot read (empty, `abc`, four decimals,
`1e3`) is **not steppable** and the arrow does nothing — the parser's own refusal is the guard, so there
is no path on which a float could be produced and none on which an unreadable literal could be sent.

### The two behaviours the contract delegates, decided

**1. An arrow on an UNSET field does nothing, and sends no command.**
The contract says an empty box has no value to step. The tempting alternative — step from the
placeholder — is wrong three times over: `lineSpacing`'s placeholder is `1`, `borderWidth`'s is `none`
(not a number at all), and `fontSize` has no `empty` and so has no placeholder. Stepping a placeholder
would therefore need a per-field table of implied defaults, which is **exactly the second authority the
contract's "do not mint a second list" clause forbids**. It would also convert an unset, inherited value
into a pinned one on a keypress the author reads as a nudge. So: no parse, no step, no command.

**2. An arrow on a MIXED selection does nothing, and sends no command — by the same rule, not a second one.**
A mixed selection already presents as an empty draft: `App.tsx:1773` sets `committed = same ? … : ''`,
so `draft` is `''` and the placeholder reads `Mixed`. The empty-draft rule therefore covers it with no
extra branch, and that is the point — **one predicate, `the draft does not parse`, closes both rows.**
The alternative is worse than merely undefined: stepping a mixed selection means picking one component's
width and flattening every other component onto it, which is a destructive edit fired by a nudge key.
(If the author has TYPED into a mixed field, the draft is no longer empty and it steps like any other
draft. That is correct — they are stepping the value they just entered, not the selection's disagreement.)

### Why an arrow step must not disable the field

`submit` (`App.tsx:1785-1794`) sets `pending`, and `shared` (`:1849`) carries `disabled: pending`. That
disable exists so a keystroke cannot race a commit already in flight. **An arrow step IS that keystroke,
and the disable would end an arrow HOLD after exactly one step** — which is the one thing this story's
own browser run is meant to photograph.

Measured, not assumed, in Chromium 1217 (`document.activeElement.id` around a focused input):
`{"before":"i","during":"","after":""}` — disabling a focused input moves focus to the body, and
**re-enabling does not give it back**. Key repeat is delivered to the focused element, so after the
first step the remaining repeats go to the body and the hold is dead.

So the step path runs through the **same** `submit`, keeping the **same** single-flight `pendingRef`
guard, and differs in one respect only: it does not raise the `pending` flag that disables the input.
This is not a second commit path — the intent, the encoder, the reconciliation and the flight guard are
all shared verbatim. A step attempted while another commit is in flight is dropped by `pendingRef`, and
the next repeat carries the accumulated draft.

## Verification

**Commands** — one per line, **exit codes read from `$?` immediately, never through a pipe**; zsh, so
`${PIPESTATUS[0]}` is wrong.
- `cd folio-designer && npm test` — the one permanent red is `canvas-authority-contract.test.ts`'s corpus
  scan (DW-152); **match it by NAME, not count.**
- `npm run typecheck` · `npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`. **No Go is touched; say so rather than running it.**

**A BROWSER RUN.** `chromium-1217` via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (DW-180).

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**11 rows** — the band-edge row was added at review);
state the population beside every zero, pair every absence claim with a positive control; **a comment is
not a measurement**; **use `/usr/bin/grep`** — recursive `grep` returns false zeros here.

## Spec Change Log

**2026-09-04 — builder, at build dispatch. Nothing inside the frozen contract was edited; its slab
md5 is `e978d80b40800894bebba781b45535bf` before and after these edits.**

1. **Code Map re-anchored from `85f213f` to the dispatch HEAD `8bfcec8`.** `git diff --stat
   85f213f..8bfcec8` touches exactly one source file, `folio-designer/src/App.tsx` (+15 −1): Story
   17.2 replaced the canvas-region `onClick` with a 14-line comment at `:1452`. Every panel anchor
   sits below that line and moved +14. Each was re-measured individually. Nothing the spec describes
   was rewritten — the seam, the predicate and the formatter are all byte-identical, just relocated.
   Two off-by-one anchors in `component-property-command.ts` (`pointFields` is `:7` not `:6`,
   `ratioFields` `:17` not `:16`, `rawNumberLiteral` `:28-32` not `:29-33`) and one over-tight range
   in `decimal.go` (`decodePoints` is `:30-68`, and its three-decimal refusal is at `:55`) were
   corrected at the same time. **The frozen contract still quotes the `85f213f` numbers and was left
   alone; a reader of that block should apply +14 to its `App.tsx` citations.**

2. **`linespacing.go` added to the Code Map with its measured constants.** The contract's "0.001–1000"
   is correct, but the constants are expressed in THOUSANDTHS (`Min = 1`, `Max = 1000000`), which is
   the form the clamp arithmetic actually needs.

3. **The two behaviours the contract delegates were decided and written into Design Notes** — an
   arrow on an unset field and on a mixed selection both do nothing and send no command, closed by
   one predicate rather than two branches. Rationale is in Design Notes; both are asserted.

4. **A third design point was settled by measurement and recorded in Design Notes**: the step path
   must not raise `pending`, because disabling a focused input in Chromium 1217 blurs it permanently
   and would end an arrow hold after one step. This is a consequence of the existing `disabled:
   pending` on `shared`, not a new decision about scope.

None of the above re-opens a boundary, adds a goal, or changes an acceptance criterion.


---

## Spec Change Log — review pass (2026-09-04)

**5. THE CONTRACT'S BOUND ENUMERATION IS INCOMPLETE, and this is the review's one finding of
substance.** The frozen block names two bounding sites — `component_commands.go:1006`'s `> 0` rule and
`linespacing.go` — and the Code Map named the same two. **There is a third, on the same command path,
and nothing in the spec mentions it.** `updateComponentPropertiesInPlace` (`component_commands.go:839`)
calls `containComponent` on every id after applying the changes (`:880` -> `:1912`), and that predicate
opens:

    outside := x < 0 || y < 0 || width < 0 || height < 0 || x > band.Width || width > band.Width-x

with `y > band.Height` and `height > band.Height-y` added in the vertically capping bands.

Two consequences, and they are NOT the same size:

- **Below zero — CLOSED in this pass.** `x` and `y` were read as unbounded (that reading is written
  into the implementation's first comments, its commit message and the mirror test) and stepped freely
  through zero. They do not: their floor is the band origin. A component dropped at the origin — which
  is every fixture in the test file and the common case in the app — answered ONE ArrowDown with an
  engine refusal the author never asked for, which is precisely the failure this story exists to
  prevent. **Measured end-to-end in Chromium 1217 against the real engine**: typing `-1` into `X` and
  committing produces `e1: component.geometry: folio: component geometry must stay within content`.
  Fixed by mirroring the floor as `ORIGIN_FLOOR_FIELDS`, tied to the Go clause by a new
  `origin floor mirror` describe, and red-proved (deleting the floor makes the box read `-1`).

- **Above, against the band extents — OPEN, and left to a human.** The panel does NOT clamp `x`, `y`,
  `width` or `height` to the band edge, so a step there still reaches the engine's located refusal. The
  contract's rule ("an arrow may not produce a value the engine will refuse") reads on it, but the fix
  is not a mirror of a constant: **the ceiling is per-component** — two components of equal width at
  different `x` have different width ceilings — so a selection-wide clamp needs a ruling about what a
  step does when the selection disagrees, plus band extents threaded into `PropertyDraft` and a third Go
  constant (`bandsCappingVertically`) mirrored. That is a scope decision this story does not carry, and
  it is recorded here rather than resolved. **RULED ON THE SAME DAY — see item 9 below. The answer was
  DO NOT CLAMP, and the acceptance is met by asserting the refusal instead.**

**6. Modifier keys are now explicitly declined.** The contract puts modifier BEHAVIOUR out of scope;
the implementation as first written stepped on `Shift`/`Ctrl`/`Alt`/`Meta` arrows anyway, taking three
shipped text-input gestures (extend selection, caret to start/end) away from the author. Declining a
modified arrow adds no behaviour — it restores the browser's — so it was patched and asserted, not
escalated. Confirmed in the browser: `Shift+ArrowUp` on a width of 79 leaves it at 79.

**7. Three guards had no test at all**, found by mutation rather than by reading: the leading CEILING
(deleting `Math.min(floored, highest)` left all ten matrix rows green), the arithmetic itself (see the
correction in Design Notes), and the non-disabling step decision. All three now have a test and a
red-proof.

**8. `baseline_commit` in the frontmatter is `85f213f`, the PLAN-GATE commit, not this build's
baseline.** The workflow's step-03 says never to overwrite a `baseline_commit` that is already present,
so it was left alone. The build's true baseline is **`8bfcec8`**, and every diff, review and measurement
in this pass was taken against that.


## Review Triage Log

Three review layers ran in parallel over the diff `8bfcec8..working tree`. Every finding routed as
`patch` was verified against the code before being applied; the load-bearing one was verified against
the running engine in a browser.

**PATCHED (10).**
1. *high* — `x`/`y` stepped below zero into `containComponent`'s refusal. Raised independently by two
   layers. Verified in Go, then end-to-end in Chromium 1217: committing `-1` yields
   `e1: component.geometry: folio: component geometry must stay within content`. Fixed via
   `ORIGIN_FLOOR_FIELDS` + a new mirror describe + a test + a red-proof.
2. *high* — the float row could not fail. Found by my own mutation before the layers ran: the
   contract's `1 + 0.1` illustration is factually wrong. The test now steps five times.
3. *medium* — the leading CEILING had no test; deleting the clamp left all ten matrix rows green.
4. *medium* — the non-disabling step decision had no test; it now asserts both arms against Enter.
5. *medium* — a modified arrow stepped, hijacking Shift+Arrow selection and Cmd/Alt caret motion.
6. *medium* — the mirror test's comment asserted `x`/`y` were unbounded, which is false. Corrected, and
   the `not.toContain` assertions re-explained for what they actually pin.
7. *low* — `App.tsx` claimed a mid-flight repeat "carries the accumulated draft". It is a RACE: if the
   engine's answer wins, the committed transition discards the press. Comment corrected; coalescing
   named as the out-of-scope real fix.
8. *low* — the mixed-selection comment argued against flattening while its own test asserts flattening
   after a type. Reconciled: the guard is "the draft does not parse", not "never flatten".
9. *low* — the describe header's test count was wrong (mine, introduced at review). Now fourteen,
   counted rather than asserted.
10. *low* — `stepThousandths` fell through to `100` for any non-`pt` field; keyed on `lineSpacing`
    explicitly, matching the two lines below it.

**ESCALATED, THEN RULED (1).** The band-edge ceilings — Spec Change Log item 5, resolved by item 9.
It was escalated as an intent gap because the frozen contract's universal rule reads on it while the
contract's own enumeration of bounds does not name the site. The ruling defined the behaviour rather
than the clamp, so no intent gap remains. **No loopback was performed and nothing was reverted** — the
rest of the work was sound and nothing was committed. `review_loop_iteration` is therefore still `0`.

**DEFERRED (7)**, appended to `deferred-work.md`: one undo entry per step, the mid-flight step race,
spinbutton semantics, page-setup fields not stepping, sibling controls live during a step, IME
composition, and a typed out-of-range draft snapping to the bound on one press.

**REJECTED (3).** `POSITIVE_LENGTH_FIELDS` as a `Set` (style only, and the mirror's regex reads the
array form); an options object for `submit`'s two booleans (churns five shipped call sites for
legibility alone); a `Number.isSafeInteger` guard on the stepped value (`current` is already
safe-checked and a step is at most 1000 thousandths, so the guard is unreachable through any document
this UI can express).


---

## Spec Change Log — ruling (2026-09-04)

**9. THE BAND EDGE IS NOT CLAMPED, BY RULING. Assert the refusal instead.** The escalation in item 5
was answered, and the reasoning is the part that matters, because it is what stops someone "fixing"
this later:

> **The floor and the ceiling are not the same kind of bound, which is why one is mirrored and the
> other must not be.** Zero and negative are values these fields can NEVER legally hold — whatever
> document is open, wherever the component sits. That fact is a property of the FIELD, it is stable,
> and mirroring it in the browser cannot go stale in a way that matters. A band-edge ceiling is a
> property of the LAYOUT: it depends on the component's position, its band's height, and the page. For
> the browser to clamp there, the browser would have to compute where the content band ends — geometry
> the engine owns and projects. That is the same authority boundary **AD-17** draws for text, and a
> second, quietly-drifting copy of the engine's containment rule does not belong in the inspector to
> save an author one error message.

So an arrow that would push a component past its band edge **sends** the value; the engine refuses it
with `component.geometry: folio: component geometry must stay within content`; and the existing
`role="alert"` path renders it — exactly what already happens when the author TYPES that value.
**Consistency with typing is the property asserted**, so that the arrow cannot later be turned into a
special case. A matrix row was added for it (the contract's matrix is now **11 rows**; its md5 moved
from `e978d80b40800894bebba781b45535bf` to `87fd4f047c7bf825b6e6e0e4d28fbd89`, the only amendment to
the frozen block in this story and made on explicit written instruction). The clamp task is `[x]`.

**10. THE FLOAT EXAMPLE IS CORRECTED AT SOURCE, and the correction is ratified.** `1 + 0.1` really is
exactly `1.1`; the spec's original witness would not have caught the bug the spec exists to prevent.
The Design Notes now teach the true fact — the damage begins at `1.1 + 0.1 = 1.2000000000000002` — and
say plainly that mutating a passing test is what surfaced it. The frozen matrix row's parenthetical is
left as-is because its expectation (`1.1`) is correct; Design Notes carries the correction of record.

**11. RATIFIED, unchanged:** declining to step a MODIFIED arrow (the absence of modifier behaviour, not
the coarse/fine stepping ruled out of scope); and the two delegated decisions — an unparseable draft
does nothing, with one predicate covering both the unset field and the mixed selection. The reason
given for refusing to step a placeholder — that a per-field default table is a second authority on
defaults — was ratified explicitly, with the note that **Story 17.3 is about to make the engine the
first** such authority.

## Matrix Audit

**11 rows, 11 results.** Every covering test ran and passed in the final `npm test` (692 passing).

| # | Matrix row | Covering test | Result |
|---|---|---|---|
| 1 | Up on a point field | steps a point field UP by one point, and the committed value follows | pass |
| 2 | Down on a point field | steps a point field DOWN by one point | pass |
| 3 | Up on leading | steps leading by a tenth repeatedly without ever spelling a float | pass |
| 4 | Down at the floor | stops a point field at the smallest LEGAL value rather than stepping into a refusal | pass |
| 5 | Down at leading's floor | stops leading at the engine's own floor and sends nothing | pass |
| 6 | Arrow past the band edge | sends a band-edge step and shows the engine's refusal, exactly as typing the same value does | pass |
| 7 | Arrow on an unset field | does nothing on an UNSET field, and sends no command | pass |
| 8 | Arrow during a drag | does nothing while a DRAG owns the field, exactly as typing does nothing | pass |
| 9 | Arrow on a non-numeric field | leaves a NON-NUMERIC field to the browser, and steps exactly the fields the control already calls decimal | pass |
| 10 | Arrow on a mixed selection | does nothing on a MIXED selection, and sends no command — until the author types a value into it | pass |
| 11 | Enter and Escape | leaves Enter and Escape exactly as they were | pass |

**Four further tests cover behaviour the matrix does not reach**, each found by mutation or review, not
by reading: the leading CEILING, the non-disabling step, the ORIGIN floor on `x`/`y`, and the MODIFIED
arrow. Fifteen tests in the describe; the mirror file carries eight more across two describes.

## Suggested Review Order

**The arithmetic, which is the whole story**

- The exact inverse of `points()`: digit groups read as integers, never the decimal.
  [`App.tsx:1792`](../../folio-designer/src/App.tsx#L1792)

- Integer step, then floor, then cap — each against its own bound, never the unclamped value.
  [`App.tsx:1892`](../../folio-designer/src/App.tsx#L1892)

**One authority for the numeric set**

- The `inputMode` predicate, hoisted so the attribute and the step read the same fact.
  [`App.tsx:1850`](../../folio-designer/src/App.tsx#L1850)

- Its second consumer, the rendered attribute — no second list anywhere.
  [`App.tsx:1978`](../../folio-designer/src/App.tsx#L1978)

**The bounds, mirrored from the engine rather than restated**

- Floors: leading's pair, the `> 0` four, and the band origin for `x`/`y`.
  [`App.tsx:1871`](../../folio-designer/src/App.tsx#L1871)

- Go's `> 0` rule as a list the panel can read.
  [`component-property-command.ts:30`](../../folio-designer/src/component-property-command.ts#L30)

- The origin floor, and the comment recording what is deliberately NOT mirrored.
  [`component-property-command.ts:47`](../../folio-designer/src/component-property-command.ts#L47)

**The three guards on the keyboard seam**

- Drag, then unparseable draft — one predicate closing both delegated rows.
  [`App.tsx:1876`](../../folio-designer/src/App.tsx#L1876)

- Arrows added; Enter and Escape untouched; a modified arrow left to the browser.
  [`App.tsx:1938`](../../folio-designer/src/App.tsx#L1938)

- `disable` opt-out: raising `pending` would blur the field and end a hold.
  [`App.tsx:1834`](../../folio-designer/src/App.tsx#L1834)

- The step's one call, sharing the encoder and the single-flight guard verbatim.
  [`App.tsx:1908`](../../folio-designer/src/App.tsx#L1908)

**Tests**

- The float row, which had to climb five steps to become falsifiable at all.
  [`App.test.tsx:2395`](../../folio-designer/src/App.test.tsx#L2395)

- The band edge: arrow and keyboard asserted to be the same act.
  [`App.test.tsx:2732`](../../folio-designer/src/App.test.tsx#L2732)

- The origin floor, found at review against a real engine refusal.
  [`App.test.tsx:2660`](../../folio-designer/src/App.test.tsx#L2660)

- The hold decision, asserted in both directions against Enter.
  [`App.test.tsx:2478`](../../folio-designer/src/App.test.tsx#L2478)

- Go/TS ties, each with non-vacuity and both-direction drift proofs.
  [`engine-bounds-mirror.test.ts:281`](../../folio-designer/src/engine-bounds-mirror.test.ts#L281)
  [`engine-bounds-mirror.test.ts:357`](../../folio-designer/src/engine-bounds-mirror.test.ts#L357)
