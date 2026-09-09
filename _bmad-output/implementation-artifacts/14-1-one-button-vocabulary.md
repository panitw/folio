---
title: '14.1 — One button vocabulary'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: 'dbe058bf33411dfbd42824cb1d6e36f11a1ef686'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The designer has no written rule for when a control is spelled as a word and when it is
spelled as a glyph, so each story picked one and the seams show. Measured at `924195b`: the document
bar's six local-file controls carry three treatments in one row — Open and Save are icon-only
`.icon-button`s, Save As and Start blank are words, Undo and Redo are words with a visible `<kbd>` —
and the TYPOGRAPHY section's `.property-grid` puts three-or-four SVG align icons in the left cell
beside the words TOP / MID / BOT in the right cell, both rendered by the same component through the
same `.property-segment` class. Nothing in `DESIGN.md`, `EXPERIENCE.md` or `docs/` states which is
right, and no test can tell. Without a written, checked rule the next four Epic 14 stories each
re-decide it.

**Approach:** Write the rule down first, in the vocabulary `DESIGN.md` already establishes, and make
it checkable by a contract test that **derives its population from the rendered designer** rather
than naming controls — the failure shape DW-305 recorded, where a 4.5:1 contrast floor asserted over
a hand-written five-pair list let Story 13.5 ship a 2.25:1 pairing with nothing red. Then apply the
rule to exactly the two surfaces the acceptance criteria name — the document bar's action family and
the two TYPOGRAPHY segmented controls — and **report, rather than fix**, every other control the
audit turns up.

## Boundaries & Constraints

**Always:**
- The rule is written before the sweep, and the sweep is derived from the rule. If a control changes,
  the rule says why.
- The guard **derives** its population. No clause may be enforced over a hand-written list of
  controls. A control the guard cannot see must be visible as an absence, not as silence (DW-305,
  D-000.32).
- Every guard clause is **executed against a planted violation** and against the nearest legitimate
  spelling of the same thing, and each red proof names the specific clause it must wake — not merely
  "something failed" (D-11.3.7; `canvas-authority-contract.test.ts:811-815` is the harness idiom).
- Accessible names are preserved or improved, never lost (UX-DR25, `EXPERIENCE.md:282`). A control
  that loses its visible text keeps the `aria-label` it already had.
- Icons are 16px-grid stroked SVG written inline as components (`DESIGN.md:602`), never an emoji,
  never a file on disk (D-14.0.1 — the offline release has two free cache slots and
  `vite.config.ts:16` sets `assetsInlineLimit: 0`, so any emitted file is its own row).
- Only committed engine values and presentation change. No document-model change, no command-surface
  change, no format change, no rendered byte changes.

**Ask First:**
- Extending the sweep to any control the acceptance criteria do not name, beyond the one enlargement
  ruled in R-Q4 below. The audit's other findings are reported as a list, not fixed.
- **If `role="group"` on `.mode-switch` costs more than one attribute** — if it changes any existing
  accessible name or reds any test — STOP and report. That would make it real scope and the ruling
  changes (R-Q4).
- Anything that would make an open DW entry trivially fixable as a side effect — report it, do not
  take it.
- Any control whose acceptance criterion cannot be met without work outside this fence.

**Never:**
- **Never edit `DESIGN.md`** (R-Q1). The rule is expressed in its vocabulary and lives in this spec.
- Never make `design-contract.test.ts:165`'s contrast-pair list exhaustive. DW-305 is an explicit
  deferral and reds on pre-existing violations this story did not cause.
- Never take DW-281, DW-300/301/302, DW-303, DW-306..310, DW-313..324, or any other open entry an AC
  does not name.
- Never add a `.svg`, image or font file. Never add a dependency.
- Never edit `sprint-status.yaml`, `deferred-work.md`, `epics.md`, or any file under
  `fixtures/declared-variants/`.
- Never commit, stage, stash, branch, push, or open a pull request. The orchestrator makes every
  commit.
- Never write a compile-only e2e placeholder and call it coverage.

## Rulings at CHECKPOINT 1

Four questions were put to the orchestrator at the plan gate and all four were ruled. Recorded here
verbatim in substance, inside the frozen block, so the implementer inherits the ruling rather than
re-deriving it.

**R-Q1 — The rule is expressed in `DESIGN.md`'s vocabulary. `DESIGN.md` is NOT edited.** The Epic 16
discriminator settles it (`epic-16-decision-log.md:2745-2777`): *"transcribe a declared value, refuse
an undeclared one."* A word-versus-glyph rule is undeclared — that file has no `components.button`
and no such rule. Supporting measurements: AC1 and `epic-14-context.md:105` both say *in the terms
of*, not *into*; and `DESIGN.md` has **exactly one commit in its whole history**, the initial one.
The orchestrator's added reason, which belongs in the spec: the two available ways to put this rule
into `DESIGN.md` are **"breaks the build"** and **"is invisible to the guard"** — a new top-level key
reds `design-contract.test.ts:21`'s exact per-group name set equality unless `src/design-tokens.ts`
moves with it, and per DW-282 a **nested** sub-key is not checked at all, so it would ship silently
unguarded. Neither is acceptable, so the rule lives here.

**R-Q2 — Pin the V2 violators, do not fix them.** Ruled by the orchestrator directly, as a
*narrowing*: it defers roughly a dozen judgements rather than making them, and 14.2, 14.3, 14.4 and
14.7 each own one of those surfaces and will make the call with the surface in front of them. Two
conditions are explicit and the review layers hold the implementer to them:
- **Pin by accessible name, never by count.** A count floor passes when one control is renamed and
  another appears.
- **The pin must red on a thirteenth, and that must be EXECUTED, not asserted** — add one, watch it
  fail, remove it, report the numbers. *"A characterisation test nobody has seen fail is a list, not
  a guard."*

**R-Q3 — AC2 reaches word-versus-icon only; the shortcut asymmetry is a finding, not this story's
work.** Keep Save's `title`, add no `<kbd>`. `EXPERIENCE.md:238-239` documents hints as belonging "in
menus and tooltips" while the shipped inline `<kbd>` on Undo/Redo/PREVIEW already contradicts that,
so fixing it in either direction — making Save's ⌘S visible, or removing Undo's `<kbd>` to match the
document — is a design ruling rather than a sweep. *"A story whose subject is consistency must not
pick one silently."* The orchestrator has registered the asymmetry.

**R-Q4 — `.mode-switch` gets `role="group"` too, by orchestrator ruling, and no AC covers it.** Both
`App.tsx:2244` and `:2259` are roleless `<div>`s carrying an `aria-label` the accessibility tree
drops; they sit in the same header; `.document-actions` must be fixed regardless. Shipping a story
whose subject is "no two members of the same family are spelled differently" having fixed one of two
identical defects in front of it would reintroduce the drift the story removes. It is one attribute,
its two members already pass, and it is trivially revertible. **The Delivery Log must say it was
outside the ACs and included by orchestrator ruling**, so it reads as a judgement rather than as an
absorbed sweep. If it costs more than an attribute, the Ask First fence above applies.

## I/O & Edge-Case Matrix

The unit under test is the control-vocabulary guard. "Input" is a control as it renders; "Output" is
the guard's verdict.

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Uniform class family | Every rendered control carrying `.file-button` is a word | Clause R1 passes | N/A |
| Split class family | One `.property-segment` renders an `<svg>`, a sibling renders the text `TOP` | R1 fails, naming `property-segment` and both treatments | The failure message names the class and the two treatments found |
| Uniform named group | `role="group"` "Local file actions" holds six words | Clause R2 passes | N/A |
| Split named group | The same group holds two glyph controls and four words | R2 fails, naming the group's accessible name | Message names the group and the mixed members |
| Glyph without a name | A control renders only an `<svg>` and carries no `aria-label`/`aria-labelledby` | Clause R3 fails | Message names the control's position and its empty name |
| Glyph with a name | A control renders only an `<svg>` and carries `aria-label="Align left"` | R3 passes | N/A |
| Shortcut hint is not text | Undo renders `Undo` plus `<kbd aria-hidden="true">⌘Z</kbd>` | Classified **word**; the `aria-hidden` subtree is excluded before classification | N/A |
| Symbol standing in for an icon | A control's only visible text is `×` or `−` | Classified **glyph**, not word | Counted into the R4 census below |
| Ungrouped glyph appears | A glyph control outside any segmented control that is not in the recorded census | Clause R4 fails: the closed set moved | Message names the new control, so a human rules on it |
| Population shrinks | A render state is dropped, or a control class stops rendering | The coverage floors and the visited-name set fail | Message shows which names went missing, not only a count |
| Empty control | A control renders no text and no `<svg>` (drag handles) | Classified **nameless-by-content**; R3 applies, R1/R2 do not | N/A |

</frozen-after-approval>

## Code Map

Anchors verified at `924195b` and re-confirmed at the baseline
`dbe058bf33411dfbd42824cb1d6e36f11a1ef686`. HEAD moved four times under planning (from the dispatch's
`f7766fb`) and **no `folio-designer`, `folio-go`, `lint/` or `tools/` file changed in any of them** —
measured with `git diff --name-only f7766fb..dbe058b` filtered to those four trees, which returned
nothing. Every line anchor below therefore still holds. `App.tsx` holds two literal NUL bytes near
`3201-3202` — every search in it must use `grep -a`.

**The two surfaces the ACs name**

- `folio-designer/src/App.tsx:2242-2260` — the document bar `<header>`. `:2244` is one dense line
  holding `<div className="document-actions" aria-label="Local file actions">` and all six controls:
  Open and Save as `className="icon-button"` with `<Icon name="open"/>`/`<Icon name="save"/>` and
  `aria-label="Open local template"` / `"Save local template"` (Save alone also carries
  `title={\`Save (${shortcuts.save})\`}`); Save As, Start blank, Undo, Redo as `className="file-button"`
  with visible text, Undo/Redo each ending `<kbd aria-hidden="true">{shortcuts.undo|redo}</kbd>`.
  **`.document-actions` is a plain `<div>`** — its `aria-label` is dropped by the accessibility tree
  because a roleless `div` takes no name. `:2259` `.mode-switch` has the identical defect.
- `folio-designer/src/App.tsx:85-87` — `Icon({ name: 'open' | 'save' })`, two hardcoded `path d`
  strings, `className="icon"`. **Its only consumers are the two doc-bar buttons at `:2244`.**
- `folio-designer/src/App.tsx:2629-2641` — `AlignVariant`, `alignGlyphs` (four `path d` strings),
  `AlignIcon` (`className="segment-icon"`, `viewBox="0 0 16 16"`, `strokeWidth="1.2"`,
  `aria-hidden="true"`), `alignSegments` (three), `justifySegment` (`:2640`), and `valignSegments`
  (`:2641`) whose three `content` values are the **string literals** `'TOP'`, `'MID'`, `'BOT'` and
  whose `label` values are `Vertical align top|middle|bottom`.
- `folio-designer/src/App.tsx:2658` — `alignChoices` is **four** segments for an all-text selection
  and **three** otherwise. AC3 says "three"; that is ruled false by **D-000.9 item 9**
  (`epic-11-14-decision-log.md:883`) and is inherited here rather than rediscovered. The Epic 14
  preamble said "three" too and **the orchestrator corrected `epics.md` in `8ea9e56`** while this
  spec was at the plan gate. This story does not touch `epics.md`.
- `folio-designer/src/App.tsx:2681` — the TYPOGRAPHY `PropertySection`; both `SegmentedProperty`
  instances sit in one `.property-grid`, which `App.css:284` lays out as `1fr 1fr`, so Align and
  Vertical align are literally side by side at equal width.
- `folio-designer/src/App.tsx:3964-3979` — `SegmentedProperty`. **One JSX site renders both
  controls**: `div.property-editor > div.property-segmented[role="group"][aria-label] >
  button.property-segment[aria-pressed][aria-label={segment.label}][title]`, children
  `{segment.content}`. Consequence: **the Align/Vertical-align divergence lives in data
  (`alignSegments` vs `valignSegments`), not in JSX**, so a source-text or AST scan of the markup
  cannot see it. The guard must observe rendered output.

**Stylesheet**

- `folio-designer/src/App.css:35-45` — `.document-actions` (`:35`), `.icon-button` + `:disabled`
  (`:36-37`), `.file-button` + `:disabled` (`:38-39`), `.icon` (`:40`), `.mode-switch` (`:42-45`).
- `folio-designer/src/App.css:20-27` — a Story 13.2 comment that names `.icon-button`,
  `.file-button` and `.mode-switch button` as the three consumers of `--status-bar-height`. **If
  `.icon-button` is removed this comment becomes false** and must be corrected in the same change.
- `folio-designer/src/App.css:460-464` — `.property-segmented` (`:460`), `.property-segment`
  (`:461`), `[aria-pressed="true"]` (`:462`), `:disabled` (`:463`), `.segment-icon`
  (`width: 12px; height: 12px`, `:464`). New vertical-align icons reuse `.segment-icon` unchanged.

**The guards this change must not break, and the one it imitates**

- `folio-designer/src/canvas-authority-contract.test.ts` — the mechanism to copy. `:6-16` derives
  three populations with `fs.readdirSync(dir, { recursive: true })` plus an extension filter;
  `:294-296` and `:743-774` put `toBeGreaterThanOrEqual` floors on the counts **and** compare an
  independently-walked `e2e` set so no file drops out by name; `:228-248` `withoutComments` is a
  hand-written scanner (not a regex) so `//` inside a string is not a comment; `:811-815`
  `violationsForFile(name, source)` is the harness every red proof runs through, and `:393`/`:736`
  show the `expect(...).toContain(String(rule))` idiom that makes a proof name the rule it wakes.
  **It also scans `src/**/*.test.ts(x)`**, so the new test file is inside its corpus: it must not
  contain `getComputedStyle`, `getBoundingClientRect`, `measureText`, `document.fonts`,
  `devicePixelRatio`, or the `offset*`/`client*`/`scroll*` identifiers.
  `:456-457` and `:477` pin the literal CSS source text of `.property-segment[aria-pressed="true"]`
  and `.property-toggle-unavailable` — **those rules must not be edited.**
- `folio-designer/src/design-contract.test.ts` — `:14-18` `namesFromDesign(group)` parses `DESIGN.md`
  frontmatter by regex; `:21-23` asserts **exact per-group token-name set equality** against
  `src/design-tokens.ts`, so any new key at two-space indent in `DESIGN.md` reds unless
  `design-tokens.ts` moves with it; `:84-91` forbids any `#hex`/`rgb(`/`hsl(` literal in `App.css`
  and any non-`var(--radius…)` `border-radius`; `:145-151` pins `var(--type-display)` and
  `var(--type-numeric-lg)` at **exactly one** occurrence each; `:165` is DW-305's hand-written
  five-pair contrast list — **out of scope, do not extend**.
- `folio-designer/src/App.test.tsx:3132` asserts the seven segment accessible names
  `Align left|center|right|justify` and `Vertical align top|middle|bottom` as a set. Those names are
  preserved by this change, so it stays green.
- Population searched for the visible strings `TOP`/`MID`/`BOT` as an assertion: **all 122
  `folio-designer/src/**/*.test.ts(x)` files and all 21 `folio-designer/e2e/*.spec.ts` files** —
  **no test asserts them.** Positive control for that search: `App.test.tsx:3132` was found by the
  same command. So replacing the words with icons breaks no existing assertion.
- `folio-designer/src/test/setup.ts` registers `@testing-library/jest-dom/vitest` globally, so
  `toHaveAccessibleName` is available; it is used at `App.test.tsx:5290` and `FontBrowser.test.tsx:504`.
- `folio-designer/vite.config.ts:33` — vitest `include: ['src/**/*.test.{ts,tsx}', …]`, no `exclude`,
  so a new file under `src/` is picked up with no config change.

**Design record (what it does and does not say)**

- `…/ux-designs/ux-folio-2026-08-23/DESIGN.md` — 604 lines, YAML frontmatter `1-298` then prose.
  Relevant vocabulary: `typography.label` / `label-tight` / `body` / `body-em` / `title` / `mono` /
  `mono-em`; `colors.ink-high|ink|ink-low|ink-ghost|ink-disabled`, `raised`, `active`, `edge`,
  `line-strong`; `spacing.1|3|6`; `components.doc-bar` (`176-181`), `components.segmented-control`
  (`208-220`), `components.property-field` (`200-207`), `components.palette-item` (`194-199`).
  `:433` "Uppercase belongs only to section labels and mode switches". `:602` "Don't add an emoji,
  anywhere. Icons are stroked SVG on a 16 px grid." **There is no `components.button` and no
  word-versus-glyph rule anywhere.**
- `…/EXPERIENCE.md:282` "Every icon-only control carries an accessible name." `:145` "Prefer the noun
  over the sentence for labels." `:238-239` places shortcut hints in "menus and tooltips", which the
  shipped Undo/Redo/PREVIEW inline `<kbd>` already contradicts.
- `…/mockups/Main.dc.html:38-41` draws **only** Open and Save, both as bare words in
  `padding: 4px 8px; color: #aab2bb`, with no border, no icon and no shortcut. **Population searched
  for `save as|start blank|undo|redo` across all six `mockups/*.dc.html`: no match.** The mockup
  therefore supports AC2's Open/Save half directly and says nothing about the rest of the family —
  the preamble's claim at `epics.md:4825-4826` that "the design's own document bar draws all of them
  as text" is a generalisation, not a drawn fact. Positive control for that search: the same command
  returned `Main.dc.html:39` and `Binding.dc.html:35-36` for `open`/`save`.
- `…/mockups/Main.dc.html:314-323` draws Align as three 12px SVG icons and Vertical align as the
  words TOP/MID/BOT, in one `1fr 1fr` row with identical frames. **The mockup is itself the source of
  the inconsistency AC3 names**, and AC3's "both iconic" therefore departs from the drawn design
  deliberately. Recorded here so the departure is stated rather than discovered.

**Rulings inherited**

- **D-000.9 item 9** (`epic-11-14-decision-log.md:883`) — AC3's "three SVG icons" is wrong; it is
  four for an all-text selection since justify landed.
- **D-14.0.1** (`epic-11-14-decision-log.md:5629`) — every Epic 14 story states its offline-release
  slot cost in its spec; the release carries 62 assets against a hard maximum of 64.
- **D-000.33** — `## Verification` carries unit tests, typecheck, lint and `test:e2e:compile` only.
- **D-000.32 / D-000.34** — a guard nothing runs, and an index nothing checks, decay identically.
- **D-11.3.7** — run a guard against the defect it forbids *and* the nearest legitimate spelling.
- **DW-326** (registered `8ea9e56`) — "the document bar discloses keyboard shortcuts three different
  ways, and the design doc contradicts the code". This is R-Q3's finding, already registered by the
  orchestrator. **It is not this story's work**: keep Save's `title`, add no `<kbd>`, change no
  shortcut disclosure anywhere.

## Tasks & Acceptance

**Execution:**

- [x] `_bmad-output/implementation-artifacts/14-1-one-button-vocabulary.md` -- record **THE RULE**
      verbatim in `## Design Notes` below, in `DESIGN.md`'s vocabulary, before any code changes --
      AC1 makes the rule the deliverable and the sweep its consequence, so the rule text is written
      and frozen first and every later task cites the clause it serves.

- [x] `folio-designer/src/control-vocabulary-contract.test.tsx` -- NEW. The checkable form of the
      rule. Render the designer across a declared set of states, sweep every control the render
      produces (`button`, `[role="button"]`), classify each, and enforce clauses R1-R4 over the
      **derived** population. A new contract file rather than an addition to `design-contract.test.ts`,
      because the subject is rendered output and that file is pure `fs` reads — matching the
      precedent by which `canvas-authority-contract`, `engine-ownership-contract`,
      `file-access-contract` and `preview-authority-contract` are separate files per subject.
      Requirements:
      - **Classification** (one exported-in-file helper, applied to every control): strip
        `[aria-hidden="true"]` subtrees, then — *word* if the remaining visible text contains a
        letter or digit; *glyph* if it contains an `<svg>` with no such text, or its remaining text
        is entirely symbol characters; *empty* if neither.
      - **R1 (family):** for every `class` token carried by two or more swept controls, all of them
        share one treatment.
      - **R2 (group):** for every `[role="group"]` / `[role="tablist"]` container in the sweep, all
        controls inside it share one treatment.
      - **R3 (name, UX-DR25):** every *glyph* and every *empty* control has a non-empty accessible
        name.
      - **R4 (closed set):** the set of glyph controls that are **not** members of a segmented
        control is pinned **by accessible name, never by count** (R-Q2 — a count floor passes when
        one control is renamed and another appears). This story records the set as it stands and
        fixes none of it. The pin must be **observed to red on a thirteenth**: add one, watch it
        fail, remove it, and report the numbers in the completion report.
      - **Coverage honesty:** assert floors on the number of controls swept, on the number of
        distinct class families seen, and on the number of groups seen — **and** assert the set of
        visited group accessible names by name, so a dropped render state shows which names went
        missing rather than only a smaller count. A count floor alone cannot see a shrink.
      - **Red proofs, executed:** for each of R1, R2, R3 and R4, run the classifier and rule over a
        planted violating fragment and assert the result names **that** rule, plus a matching
        green over the nearest legitimate spelling. `canvas-authority-contract.test.ts:811-815`
        and `:393` are the harness and assertion idiom.
      - Must not use `getComputedStyle`, `getBoundingClientRect`, `measureText`, `document.fonts`,
        `devicePixelRatio` or any `offset*`/`client*`/`scroll*` identifier — this file is inside
        `canvas-authority-contract.test.ts`'s scanned corpus.

- [x] `folio-designer/src/App.tsx` -- doc bar: give `.document-actions` at `:2244` `role="group"` so
      its existing `aria-label="Local file actions"` reaches the accessibility tree, and respell Open
      and Save as words (`className="file-button"`, visible text `Open` and `Save`) -- AC2. **Keep
      both existing `aria-label`s unchanged** (`Open local template`, `Save local template`): they
      contain the new visible text, so the accessible name is unchanged, WCAG label-in-name holds,
      and the ~27 existing tests that query those names stay green (AC4). **Keep Save's `title` and
      add no `<kbd>` to either control** — the shortcut-disclosure asymmetry is a registered finding,
      not this story's work (R-Q3).

- [x] `folio-designer/src/App.tsx` -- give `.mode-switch` at `:2259` `role="group"` so its existing
      `aria-label="Designer mode"` reaches the accessibility tree -- **no AC covers this; it is
      included by orchestrator ruling R-Q4.** Identical roleless-`div` defect to `.document-actions`,
      same header, and its two members already pass every clause. Change nothing else about it. **If
      this costs more than the one attribute — if it alters any existing accessible name or reds any
      test — STOP and report** rather than working it out (R-Q4). The Delivery Log must record it as
      outside the ACs and included by ruling.

- [x] `folio-designer/src/App.tsx` -- TYPOGRAPHY: replace `valignSegments`' `'TOP'`/`'MID'`/`'BOT'`
      string contents at `:2641` with an inline `ValignIcon` component written in `AlignIcon`'s exact
      idiom (`:2631-2633`) -- AC3, both segmented controls iconic. Reuse `className="segment-icon"`,
      `viewBox="0 0 16 16"`, `strokeWidth="1.2"`, `aria-hidden="true"`, and a `valignGlyphs` record
      beside `alignGlyphs`. The three `label` values are unchanged, so each segment's accessible name
      is unchanged (AC4) and `App.test.tsx:3132` stays green. Inline SVG, never a file (D-14.0.1).

- [x] `folio-designer/src/App.tsx`, `folio-designer/src/App.css` -- remove `Icon` (`:85-87`),
      `.icon-button` (`App.css:36-37`) and `.icon` (`:40`) **only if** the sweep leaves each with
      zero remaining references, proven by a stated `grep -a` over a stated population, and correct
      the Story 13.2 comment at `App.css:20-27` which names `.icon-button` as a consumer of
      `--status-bar-height` -- a comment that outlives its subject is a false index (D-000.34).
      Leave `--icon-size` and every token in `tokens.css` alone.

- [x] `folio-designer/src/App.test.tsx` -- add the behavioural assertions the contract test cannot
      make: Open and Save are reachable by their unchanged accessible names **and** now expose the
      visible text `Open`/`Save`; the six local-file controls resolve as one named group; each
      vertical-align segment renders an `<svg>` and still answers to `Vertical align top|middle|bottom`
      -- AC2, AC3, AC4. Diff test **name** sets against the baseline, never totals.

**Acceptance Criteria:**

- Given the rule for when a control is a word and when it is a glyph, when this story is planned,
  then the rule exists as written text in this spec's `## Design Notes` before any control changes,
  expressed in `DESIGN.md`'s token and component vocabulary, and every control change below cites the
  clause it serves. (AC1)
- Given the guard that checks the rule, when a control is added, renamed or respelled anywhere in the
  swept population, then it is classified and checked without any edit to a list of controls — and
  when a control that previously rendered stops rendering, the guard fails naming what went missing
  rather than passing over a smaller population.
- Given each of the guard's four clauses, when the story is complete, then a planted violation of
  that clause has been **executed** and observed to fail naming that clause, and the nearest
  legitimate spelling of the same thing has been executed and observed to pass. Reading a regex is
  not testing a regex (D-11.3.7).
- Given the document bar, when it is shown, then Open, Save, Save As, Start blank, Undo and Redo are
  all spelled as words, they resolve as one named group in the accessibility tree, and no member of
  the family is spelled differently from another. (AC2)
- Given the TYPOGRAPHY section's two segmented controls, when a text component is selected, then
  Align renders four icon segments and Vertical align renders three icon segments — one vocabulary,
  no words beside icons at the same size in the same row — and when a table is in the selection Align
  renders three. (AC3, corrected by D-000.9 item 9)
- Given every control this story changes, when it is queried by role and name, then its accessible
  name is byte-identical to the name it had before the change, and no control is left glyph-only
  without a name. (AC4, UX-DR25)
- Given the audit of every other control the sweep visits, when the story is complete, then the
  controls that violate the rule but are outside this story's acceptance criteria are **reported to
  the orchestrator as a list and pinned by R4**, and none of them is changed.

## Design Notes

### THE RULE — the designer's control vocabulary

Written in `DESIGN.md`'s terms. Clause numbers are cited by the tasks above and by the guard.

**V1 — A control is a word.** By default every control is spelled with the word for what it does, in
`{typography.body}` on the ink ramp, following `EXPERIENCE.md`'s "prefer the noun over the sentence".
`{typography.label}` uppercase is used only where `DESIGN.md:433` already permits it — section labels
and mode switches.

**V2 — A control is a glyph only as one member of a `{components.segmented-control}`**, that is, only
when it is one of the mutually-exclusive values of a single closed-set property and its siblings are
the other values. A glyph is a stroked SVG on the 16px grid (`DESIGN.md:602`), written inline as a
component; never an emoji, never a typographic character standing in for an icon, never a file on
disk.

**V3 — Uniformity.** (a) Every control that shares a control class is spelled the same way.
(b) Every control inside one named control group is spelled the same way. A row that mixes icons and
words at the same size is a violation of (b); two controls of one class that disagree is a violation
of (a).

**V4 — A glyph always carries an accessible name** (`EXPERIENCE.md:282`, UX-DR25). A respelling never
removes or narrows a name; a control that loses visible text keeps the name it had.

**What this story enforces and what it only records.** V3(a), V3(b) and V4 are enforced over the full
swept population as clauses R1, R2 and R3 — no control is named, so a control added tomorrow is
checked tomorrow. **V2 is not enforced as a pass/fail**, because doing so today would red on roughly
a dozen controls this story's acceptance criteria do not name. **Measured after implementation, the
census is eleven names**, not the dozen this paragraph estimated at planning time: the canvas `Zoom
out` / `Zoom in` and the preview `Previous PDF page` / `Next PDF page` / `Zoom out PDF` / `Zoom in
PDF` steppers, the inspector's four `×` clear and `∅` null actions, and the font-family disclosure
chevron. **Two things this paragraph guessed at planning time are NOT in it, and the classifier is
why.** The `B`/`I` toggles classify as **words** — V1's own test is "contains a letter or digit", and
a single initial passes it, even though the design draws them as letter-glyphs
(`Main.dc.html:309-310`). The drag handles classify as **empty** — no text and no `<svg>` — and are
held by R3, which they pass, not by R4. Whether a single initial is a word is a design judgement and
it was not made silently; both are documented in the census block in the contract file. Instead R4
pins that set **by accessible name,
derived from the sweep** (never by count — a count floor passes when one control is renamed and
another appears): the set as it stands is recorded, none of it is changed, and adding a thirteenth
reds and forces a decision. That red is **executed and observed**, not asserted; a characterisation
test nobody has seen fail is a list, not a guard.

This is the deliberate line between the rule and the sweep — the rule is written and the bleeding
stops, while the existing violations go to the orchestrator as a list. Ruled at the plan gate
(R-Q2) as a **narrowing**: it defers roughly a dozen judgements rather than making them, and Stories
14.2, 14.3, 14.4 and 14.7 each own one of those surfaces and will make the call with the surface in
front of them, rather than have it made by a sweep passing through.

### Why the guard renders rather than scans source

`SegmentedProperty` (`App.tsx:3964-3979`) is a **single JSX site** whose children are
`{segment.content}`. Align and Vertical align differ only in the data passed to it (`alignSegments`
vs `valignSegments`), so the exact defect AC3 names is invisible to a source-text or AST scan of the
markup — the project's usual `canvas-authority-contract` idiom cannot see it. Only rendered output
distinguishes an `<svg>` child from the string `TOP`. The cost is that the guard checks what it
renders: that is a coverage limit, not a listed-population limit, and it is why the coverage floors
must assert **visited names**, not merely counts.

### The known residual

The sweep can only see states it renders. A control that renders in no declared state is unchecked
and the guard is silent about it. That silence is the residual risk of this design and is named here
so review hunts it rather than discovering it later. It is bounded by the visited-name assertions and
by R4's closed set, and it is strictly better than the alternative the project already has evidence
against — a hand-written list of controls, which is DW-305's shape exactly.

## Verification

**Commands** (D-000.33 — this story runs unit tests, typecheck, lint and the e2e compile only):

- `cd folio-designer && npm test` -- expected: every test file passes. Baseline at the dispatch
  commit was **71 files / 1144 tests / 0 failures**; report the new counts and **diff the test *name*
  sets**, never the totals. The new contract file must appear in the run.
- `cd folio-designer && npx tsc -b --force` -- expected: exit 0 with no diagnostics. `--force` is
  required: `npm run typecheck` is `build:wasm && tsc -b`, and `tsc -b` is incremental, so it can
  exit 0 having typechecked nothing.
- `cd folio-designer && npm run lint` -- expected: `oxlint` reports the same set of
  `react(only-export-components)` warnings as the baseline. The baseline was **exactly 4**, anchored
  at `preview/pdf-viewer.tsx:17:14`, `:18:14`, `App.tsx:4045:14`, `:4052:17` — **the anchors move, so
  re-measure and report the number found rather than quoting these.** There is no `--max-warnings`
  and no CI assertion behind this count; a fifth warning is a fact to report, not a broken build.
- `cd folio-designer && npm run test:e2e:compile` -- expected: exit 0. Compile-clean is **not**
  passing; this story adds no e2e spec, so nothing here is a coverage claim.

**Suites that did NOT run, named in those words** — the browser suite (`npm run test:e2e`), every Go
suite in `folio-go/` and `lint/`, the hashmatrix matrix legs, `npm run build` as a gate, the
`verify:offline*` chain, and the `scan:font-hosts` / `scan:host-fonts` font-host scans. All of them
run at the Epic 14 boundary gate, which the orchestrator runs. A guard never invoked is
indistinguishable from a passing one (D-000.32).

**Standing Go reds that are not regressions:** `TestCorpusMeetsP6ExerciseFloors` and its subtest
`P6g_(opaque_names)`.

**Offline-release slot cost (D-14.0.1): ZERO.** This story touches `folio-designer/src/App.tsx`,
`folio-designer/src/App.css`, `folio-designer/src/App.test.tsx` and one new
`folio-designer/src/control-vocabulary-contract.test.tsx`. It adds **no** emitted asset: the three
vertical-align glyphs are inline SVG inside a TSX component in the same idiom as `AlignIcon`
(`App.tsx:2631-2633`), not a `.svg` file, and `vite.config.ts:16` sets `assetsInlineLimit: 0` so only
emitted files take a cache row. The release stays at 62 assets against the maximum of 64, with the
margin of 2 unspent. **What was checked:** the file list this story writes to, and that every icon it
introduces is a JSX component rather than an imported file. Per D-14.0.1 no clean build is run to
confirm a zero, since `npm run build` is the boundary gate's under D-000.33.

**Manual checks:**

- The rule text in `## Design Notes` is written and unchanged before any control edit lands.
- `DESIGN.md` is byte-identical to its state at `924195b` unless the orchestrator rules otherwise.
- `git status` shows no staged change and no commit made by this workflow.

## Suggested Review Order

**The rule itself — read this before any code**

- The rule in `DESIGN.md`'s vocabulary, V1–V4, and what is enforced versus recorded.
  [`14-1-one-button-vocabulary.md` → `## Design Notes`](14-1-one-button-vocabulary.md)

- The same rule restated where it executes; each clause names the V-clause it serves.
  [`control-vocabulary-contract.test.tsx:13`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L13)

**The guard — how the population is derived rather than listed**

- The four render states. This list IS the coverage claim; everything unrendered is unchecked.
  [`control-vocabulary-contract.test.tsx:294`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L294)

- Word / glyph / empty, computed per control — the one judgement the whole guard rests on.
  [`control-vocabulary-contract.test.tsx:120`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L120)

- Visually-hidden classes parsed out of `App.css`, not listed; pinned for non-vacuity.
  [`control-vocabulary-contract.test.tsx:89`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L89)

- R0: coverage as a clause, so a dropped state names the group that vanished.
  [`control-vocabulary-contract.test.tsx:430`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L430)

- R1 and R2 — the two uniformity clauses that make AC3 and AC2 checkable.
  [`control-vocabulary-contract.test.tsx:211`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L211)

- R4's census: a multiset keyed state·name, so a repeat under a held name still reds.
  [`control-vocabulary-contract.test.tsx:267`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L267)

- The eleven recorded V2 violators — the audit's output, deferred by ruling R-Q2.
  [`control-vocabulary-contract.test.tsx:351`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L351)

- `Border edges` is disclosed as unchecked rather than counted as covered.
  [`control-vocabulary-contract.test.tsx:408`](../../folio-designer/src/control-vocabulary-contract.test.tsx#L408)

**The sweep — the two surfaces the ACs name**

- Open and Save become words in a real group; both `aria-label`s byte-identical.
  [`App.tsx:2262`](../../folio-designer/src/App.tsx#L2262)

- Why `.file-button`'s border departs from the mockup, and why AC2 outranks the drawing.
  [`App.tsx:2242`](../../folio-designer/src/App.tsx#L2242)

- `.mode-switch` gains the same role. Outside the ACs, included by ruling R-Q4.
  [`App.tsx:2277`](../../folio-designer/src/App.tsx#L2277)

- The three vertical-align glyphs, inline SVG in `AlignIcon`'s idiom; labels unchanged.
  [`App.tsx:2676`](../../folio-designer/src/App.tsx#L2676)

**Peripherals**

- `.icon-button` and `.icon` removed; the Story 13.2 comment corrected with them.
  [`App.css:21`](../../folio-designer/src/App.css#L21)

- The behavioural rows the contract test cannot make, naming controls directly.
  [`App.test.tsx:1038`](../../folio-designer/src/App.test.tsx#L1038)
