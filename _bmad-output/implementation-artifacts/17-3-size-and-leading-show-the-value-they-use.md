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

## Tasks & Acceptance

**Execution:**
- [ ] `folio-go` — put the default line spacing on the projection beside `defaultFontSize`, and extend
      `engine-protocol.ts`'s type and validator to match.
- [ ] `App.tsx` — both boxes show the effective value; **remove the hard-coded `'1'`**.
- [ ] `App.tsx` — committing the shown value writes it; clearing still omits.
- [ ] **Prove no mutation on open** — a test that loads a document with neither key, asserts the boxes
      read the defaults, and asserts **no command was sent**.
- [ ] Tests — one per matrix row.
- [ ] Run the fixture corpus and report **unmoved, per fixture**.

**Acceptance Criteria:**
- Given a document with neither key, when it opens, then both boxes show the engine's defaults and
  **no command has been sent**.
- Given the shown default is committed, when the document is saved, then the key is present.
- Given the field is cleared, when the document is saved, then the key is **absent**.
- Given the 24 fixtures, when they run, then **every digest is unmoved**.

## Design Notes

**The interesting half of this story is Go, not React.** Showing a number is one line. Sourcing leading's
number from the engine is what stops the designer becoming a second place that knows what the default is
— and this codebase has twice shipped a defect of exactly that shape, most recently the browser header
whose sentence disagreed with the list beside it.

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

**Standing rules — re-run, never cite:** the matrix audit reports **N rows, N results** (**9 rows**);
state the population beside every zero, pair every absence claim with a positive control; **a reported
digest names its algorithm and byte range**; **a comment is not a measurement**; **use `/usr/bin/grep`**.
