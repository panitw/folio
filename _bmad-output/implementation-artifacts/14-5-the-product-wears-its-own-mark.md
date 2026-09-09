---
title: 'The product wears its own mark'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: 'f3d6c3676658db8585d108393e8c9c4fc4b08356'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative, and rewritten after delivery to describe what actually shipped. The frozen Intent
below is what governs the implementation.*

The product now wears its own mark. A small square outline with a solid block inside it sits beside
the word FOLIO in two places: the bar above an open document, and the screen shown while the
application is still starting up. Both are the same drawing, produced by one component that is handed
a size and derives every coordinate from it, so the two renderings cannot quietly drift apart. Its
colour arrives from the design system's existing selection token through a single styling rule, which
means no colour value is written into the drawing itself.

The epic asked for three sizes and only two were built, deliberately. Measuring the third against the
design showed it was not the logo at all but the status marker used on the starting-up screen's file
rows, sitting in a row beside a tick and a dash and shaped differently from the mark. Building it as
written would have placed the brand inside a vocabulary that means "in progress".

The mark is decorative and deliberately silent, so the product name is announced once rather than
twice. Two smaller mismatches against the mockups on these screens were measured, judged to belong to
no requirement this story owns, and left visibly alone rather than quietly changed.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The designer chrome has never worn the mark the design gives the product. The document
bar shows the word FOLIO alone (`App.tsx:2370`) and the load screen shows `FOLIO / OFFLINE` alone
(`LoadScreen.tsx:24`). The approved mockups draw a mark before the word at both sites — a square
outline containing a smaller solid block, centred, stroked 1.5px in the design's cyan — and a brand
that appears only after loading is not the thing the author first sees. The mark has **two** sizes:
22px on the load screen and 18px in the document bar (`epics.md` AC3 as corrected at `6fa386d`;
`DESIGN.md:436-438`, which adds "Nowhere else").

**Approach:** Add one inline-SVG React component, a square outline with a centred solid block,
parameterised by a single `size` prop, and place it before the wordmark in the document bar (18px)
and on the load screen (22px). Its colour arrives as `currentColor` from an App.css class using
`var(--color-select)`, so no colour literal is written. It is decorative — `aria-hidden`, no name of
its own — so the pair announces the product name once.

## Boundaries & Constraints

**Always:**
- **One drawing.** The mark is a single component with one geometry rule parameterised by `size`.
  Not two drawings, not a lookup table of hand-tuned integers per site.
- **Inline SVG, never a file on disk.** The offline release's cache manifest counts emitted files and
  `vite.config.ts` sets `assetsInlineLimit: 0`. This is the standing rule already written verbatim at
  `App.tsx:2897-2899` (D-14.0.1) and restated by D-13.6.7. **This story's asset-slot cost is zero**,
  and that is a claim to be checked, not a formality.
- **Colour reaches the mark only as `currentColor`**, set by an App.css class that uses
  `var(--color-select)`. No colour literal in any file this story adds or edits.
- **The mark carries `aria-hidden="true"`** and no accessible name, `role`, or `<title>` of its own.
- Follow the house SVG conventions already in use (`App.tsx:94` is the model): `aria-hidden="true"`,
  `fill="none"` on the stroked shape, `stroke="currentColor"`, JSX `strokeWidth`.

**Ask First:**
- Minting a design token, or editing `DESIGN.md`, `design-tokens.ts`, or `tokens.css`.
- Widening the colour-literal prohibition beyond `App.css` and `BrandMark.tsx`.
- Any change to `sprint-status.yaml`, `deferred-work.md`, `epics.md`, or `DESIGN.md` — these are the
  owner's files. Send wording; do not edit them.

**Never:**
- **Never place the mark anywhere but the two call sites** — in particular never on the load screen's
  manifest rows, and never as a manifest row bullet. Ruled 2026-09-09, and `epics.md`'s AC3 *Given*
  was corrected on disk at `6fa386d` to match: the mark has **two** sizes; `DESIGN.md:436-438` adds
  "Nowhere else"; and the 13px shape at `Load.dc.html:78-80` is the CJK row's **in-progress marker**,
  one value of a three-value status vocabulary beside a green tick (`:41`) and a grey dash (`:96`) —
  not a brand instance. Putting the mark there would make the product's brand read as "loading".
- Never change shipped wordmark copy (`FOLIO`, `FOLIO / OFFLINE`), any shipped colour, or the manifest
  row markers (`✓ → × —`).
- Never put the mark inside a `<button>` or `[role="button"]`, and never give its container
  `role="group"` or `role="tablist"`.
- Never add a second `@media` rule to `App.css`, a third visually-hidden class, a colour literal, or
  a raw `border-radius`.
- Never insert a top-level function between `export function placementPoint(` and `function pageStyle`
  in `App.tsx`.
- Never commit, stage, stash, checkout, reset, revert, restore, clean, push, or create a branch.
- No format change, no engine surface, no new command, no rendered PDF byte.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Document bar | `<BrandMark size={18} />` beside the wordmark | `<svg>` 18×18, stroke-width 1.5; outer rect at 0.75/0.75, 16.5×16.5; inner solid rect 6×8 at 6,5 | N/A |
| Load screen | `<BrandMark size={22} />` beside the wordmark | `<svg>` 22×22, stroke-width 1.5; outer rect at 0.75/0.75, 20.5×20.5; inner solid rect 7.333×9.778 at 7.333,6.111 | N/A |
| Colour | Any size | Outer `stroke="currentColor" fill="none"`; inner `fill="currentColor"`. No literal anywhere in the component | N/A |
| Assistive technology | Either site | The mark contributes nothing to the accessible tree; the pair's accessible text is the wordmark alone | N/A |
| Stroke does not scale | 18 vs 22 | `strokeWidth` is `1.5` at both sizes — the mockups draw 1.5px at every size | N/A |

</frozen-after-approval>

## Code Map

**The two call sites**
- `folio-designer/src/App.tsx:2370` — the document bar's one line. The wordmark is a bare
  `<span className="brand">FOLIO</span>`, a direct flex child of
  `<header className="document-bar" aria-label="Document bar">` (`:2369`), sibling to
  `.document-name`, `.status-dot`, `.status-copy` and `.document-actions` (`:2393`).
  **No test anywhere asserts the wordmark's text, class, or name** — verified: `grep -rn FOLIO` over
  all test files matches only `e2e/browser-native-roundtrip.spec.ts:410` (an env var name).
- `folio-designer/src/LoadScreen.tsx:24` — `<p className="load-brand">FOLIO / OFFLINE</p>`, inside
  `<main className="load-screen" aria-labelledby="load-title">` (`:22`). `LoadScreen` is already its
  own exported component (`:13`), imported by `App.tsx:9` and rendered at `App.tsx:2174`.
  **No test asserts `.load-brand` or its text** — `LoadScreen.test.tsx` covers the progressbar,
  manifest and retry button only.

**Where the new component goes**
- A new `folio-designer/src/BrandMark.tsx`. A sibling module is the established pattern
  (`LoadScreen.tsx`, `DataPanel.tsx`, `TableEditor.tsx`, `FontBrowser.tsx`). It cannot live in
  `App.tsx` — `App.tsx` already imports `LoadScreen`, so a reverse import is a cycle.
- **House SVG idiom to copy — `folio-designer/src/App.tsx:85-95` (`PaletteIcon`).** Every production
  `<svg>` is inline JSX, `aria-hidden="true"`, `fill="none"`, `stroke="currentColor"`, JSX
  `strokeWidth`. There is no shared icon module; small local icon components are the norm.
- **Sizing deviates deliberately.** Every existing glyph is sized by a CSS class
  (`.palette-icon { width: var(--icon-size) }`). This mark takes `width`/`height`/`viewBox` as SVG
  **attributes** instead — see Design Notes.

**Styling**
- `folio-designer/src/App.css:29-35` — the whole document-bar block. `:30`
  `.brand { color: var(--color-ink-high); font: var(--type-brand); letter-spacing: var(--tracking-brand); }`
- `folio-designer/src/App.css:594` — `.load-brand { margin: 0; color: var(--color-select); … }`
- `folio-designer/src/tokens.css:7` — `--color-select: #58A6C4;`. **Re-measured: the file spells it
  UPPERCASE.** A `toContain('#58a6c4')` written in lowercase would fail. Its value is confirmed
  against `DESIGN.md:36` `select: '#58A6C4'`. **Any assertion touching this value compares
  case-insensitively and says so in its failure message**, so a later re-case of the token file does
  not read as a colour change.
- **CSS specificity check is mandatory** (DW-226 class of defect). Before adding a rule, grep
  `App.css` for every selector that could match the new elements, including ancestor-plus-type
  selectors that name no class (`.page-band > span` is the known precedent that has burned two
  stories). The document bar's rules are class-only today, but verify rather than assume.

**Guards this story must not red — each verified this run**
- `design-contract.test.ts:84-91` — bans `/#[0-9a-f]{3,8}|\b(?:rgb|hsl)\(/i` and non-token
  `border-radius`, **over `src/App.css` only**. Does not strip comments. `rgba(`, `hsla(` and named
  colours are NOT caught even there, and no `.ts`/`.tsx` file is read.
- `design-contract.test.ts:21-23` — exact token-**name** set equality against `DESIGN.md`;
  `:25-35` — every `DESIGN.md` colour value must appear in `tokens.css`; `:153-168` — a contrast floor
  that re-parses `--color-<name>: #<6 hex>`. Minting a token means three coordinated edits, one of
  them to the owner's `DESIGN.md`. **This story mints none.**
- `design-contract.test.ts:145-151` — `var(--type-display)` and `var(--type-numeric-lg)` must each
  appear in `App.css` **exactly once**.
- `canvas-authority-contract.test.ts:344` — `App.css` may contain **exactly one** `@media`, and it
  must be `prefers-reduced-motion: reduce`.
- `canvas-authority-contract.test.ts:905-911` — asserts `App.tsx` matches
  `/export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/`.
- `canvas-authority-contract.test.ts:762-766` — file-count floors: production ≥58, tests ≥51,
  e2e ≥15, production `.tsx` ≥8, `.css` ≥3. **All are `>=` — growth is free by design** (the file's
  own comment at `:759-761` says so). New files enter the 17-regex prohibition corpus, which bans
  `getComputedStyle`, `getBoundingClientRect`, `offset*`/`client*`/`scroll*`, `ResizeObserver` and
  more. A decorative SVG trips none of them.
- `control-vocabulary-contract.test.tsx:176-177` — sweeps `button, [role="button"]` **only**; `:164-166`
  sweeps `[role="group"], [role="tablist"]`. A decorative `<svg>` in a `<span>` is swept by neither.
  `:530-538` checks the group name sets in **both directions**, so a new named group reds.
  `:557` — an exact `toEqual(['diagnostic-announcement', 'sr-only'])` over App.css classes carrying
  both `clip-path: inset(50%)` and `position: absolute`.
  `:382-441` — the V2 glyph census, keyed `state · name`; only `button`/`[role=button]` members enter it.
- `App.test.tsx:1050` — no `<svg>` inside any of the six `Local file actions` buttons. The mark is a
  sibling of that group, not a child; adjacent but not violated.
- `LoadScreen.test.tsx:20` — `expect(getByLabelText('Offline payload manifest')).toHaveTextContent('✓')`.
  **This is what a manifest-row bullet change would break**, and it is the mechanical reason the
  manifest is behind an Ask First.

**Read-only evidence for the two absence claims**
- The AC2 prohibition is **not** currently repo-wide. `design-contract.test.ts:87` is the only
  colour-literal ban in the tree. Production hex outside `tokens.css` exists today and is legitimate
  document-model colour, not chrome: `App.tsx:4700` (`component.borderColor ?? '#000000'`) and
  `swatch-color.ts:20`. `engine-client.ts` has a false-positive `#abandoned` private field that any
  naive `#[0-9a-f]{3,8}` scan hits.
- The AC4 claim has no incumbent. Nothing today asserts an accessible name over the brand area at
  either site, so both fences are new and both must be shown to be able to fail.

## Tasks & Acceptance

**Operational constraints for the implementer (D-14.4.1 channel — these bind you):**
- Do **not** run any git write command. No commit, no `git add`, no stash, checkout, reset, revert,
  restore, `git clean`, no push, no branch. Reading git state is fine.
- Do **not** edit `sprint-status.yaml`, `deferred-work.md`, `epics.md`, or `DESIGN.md`.
- Do **not** touch `fixtures/declared-variants/expected.pdf`, `input.folio`, or `signoff.json`.
- `App.tsx` contains two raw NUL bytes at lines 3990-3991, so git treats it as binary. **Verify every
  edit to it with `diff -a` or `cmp`** — plain `diff` prints "Binary files differ" with zero changed
  lines, making a landed edit and a missing one look identical.
- Shell is zsh: `${PIPESTATUS[0]}` is empty. Use `cmd > log 2>&1; echo $?`, never `$?` after a pipe.
  Quote every glob (`--include='*.test.ts'`).
- Export **only** the `BrandMark` component from `BrandMark.tsx`. A second export from a file that
  also exports a component adds an oxlint `only-export-components` warning, and the warning count is
  an anchor (measured this run: exactly 4).
- If the right change turns out to be outside these tasks, **stop and report it** — do not implement
  it and do not implement a smaller version of it to avoid asking.

**Execution:**
- [x] `folio-designer/src/BrandMark.tsx` -- new file: `export function BrandMark({ size }: { size: number })`
      returning one inline `<svg aria-hidden="true" className="brand-mark" width={size} height={size}
      viewBox={\`0 0 ${size} ${size}\`} fill="none">` containing exactly two `<rect>` elements — the
      outer stroked outline and the inner solid block — with geometry from the single rule in Design
      Notes, each value rounded to 3 decimal places. No colour literal; no accessible name; no
      `<title>`; no `role`. **Carry a code comment stating that the mockup's integers were compared
      and consciously not matched: at 22px the formula gives 7.333×9.778 against the mockup's 7×10,
      deltas of 0.34px and 0.22px.** -- one drawing, one rule, so the two sites cannot drift; the
      comment stops a later reader "fixing" the formula to the mockup's integers.
- [x] `folio-designer/src/App.tsx:2370` -- import `BrandMark` and render `<BrandMark size={18} />`
      immediately before the wordmark, wrapping mark and word in one `<span className="brand-lockup">`.
      Keep `<span className="brand">FOLIO</span>` byte-identical inside it. -- the mark sits before
      the word as a pair. Place the import beside the existing `LoadScreen` import at `:9`; define
      nothing new between `placementPoint` and `pageStyle`.
- [x] `folio-designer/src/LoadScreen.tsx:24` -- import `BrandMark` and render `<BrandMark size={22} />`
      immediately before the existing brand text, wrapping both in one `<span className="brand-lockup">`
      inside the existing `<p className="load-brand">`. Do not change the text. -- the author sees the
      brand before loading finishes.
- [x] `folio-designer/src/App.css` -- add `.brand-lockup` (inline-flex, centred, a token gap) and
      `.brand-mark { color: var(--color-select); flex: none; }`. **First grep App.css for every
      selector that could match these elements, including ancestor-plus-type selectors naming no
      class**, and record what you found. No colour literal, no `border-radius`, no second `@media`,
      no third visually-hidden class. -- the token is the only source of the colour.
- [x] `folio-designer/src/BrandMark.test.tsx` -- new file covering the I/O matrix and both absence
      claims. Assert **literal** expected attribute strings at 18 and 22 (never recompute them from
      the component's own formula — a derived expectation is vacuous). Include: the geometry rows;
      `strokeWidth` `"1.5"` at both sizes; outer `stroke="currentColor"` `fill="none"`, inner
      `fill="currentColor"`; a source-text scan of `BrandMark.tsx` for `#[0-9a-fA-F]{3,8}`, `rgb(`,
      `rgba(`, `hsl(`, `hsla(` **with a negative control** proving it does not fire on the file as
      shipped and a positive control proving it fires on a hex; and a source-text assertion that the
      `<rect` markup for this mark appears in **exactly one** production file. -- the absence claims
      need a guard that can see an addition, not only a removal.
- [x] `folio-designer/src/App.test.tsx` -- add the document-bar half of AC4: render `<App>`, locate
      the `.brand-lockup` in the `Document bar` landmark, and assert (a) the `<svg>` carries
      `aria-hidden="true"`, (b) `within(lockup).queryAllByRole('img')` is empty, (c) the lockup's
      accessible text is exactly `FOLIO`, (d) the svg's `width` attribute is `"18"`. -- the fence must
      render the surface that carries the name.
- [x] `folio-designer/src/LoadScreen.test.tsx` -- add the load-screen half of AC4 and AC3: render
      `<LoadScreen>`, assert the same three decorative properties, and assert the svg's `width`
      attribute is `"22"`. Leave the existing `✓` assertion at `:20` untouched. -- the same fence at
      the second site, and the proof that size is a parameter rather than a second drawing.
- [x] `folio-designer/e2e/brand-mark.spec.ts` -- new Playwright spec: the document bar's mark is
      visible and its `boundingBox()` is 18×18. **Use `boundingBox()`; never write `getComputedStyle`
      or `getBoundingClientRect` in source text** — both are banned by the corpus prohibition scan
      (`canvas-authority-contract.test.ts`), whose one carve-out is `e9-5-border-no-ink.spec.ts`.
      Follow `e2e/application-shell.spec.ts` for setup. -- jsdom cannot see that a 1.5px stroke
      renders at 18px; this is the only witness for the visual claim.

**Red proofs — each must be EXECUTED and observed, and each names the assertion it reds.**
Two of the four ACs are absence claims, and a mutation that merely reverts the implementation cannot
falsify an absence (D-14.4.3): removing the mark makes "no hex" and "no second announcement" pass
*more* easily. So:
- [x] **ADD a hex to App.css.** Write `#58A6C4` into the `.brand-mark` rule. Expect
      `design-contract.test.ts:84-91` to red. Revert.
- [x] **ADD a hex to the component.** Change `stroke="currentColor"` to `stroke="#58A6C4"` in
      `BrandMark.tsx`. Expect **both** the new source scan and the rendered-attribute assertion in
      `BrandMark.test.tsx` to red. Revert. (Confirm the scan is not vacuous by also running its
      negative control.)
- [x] **ADD an accessible name to the mark.** Give the `<svg>` `role="img" aria-label="Folio"`.
      Expect a **named** assertion to red in **both** `App.test.tsx` and `LoadScreen.test.tsx`.
      Revert. If either stays green, the fence cannot see its target and must be rewritten before
      the story proceeds.
      **EXECUTED AND OBSERVED — and the fence WAS rewritten because of it. The spec's original
      wording named two assertions and both were wrong**, so the record is corrected here rather
      than left standing:
      - `queryAllByRole('img')` **in its default spelling stays GREEN.** Testing Library's role
        queries exclude `aria-hidden` subtrees, so the fence as originally specified could not see
        the mutation it was written to catch — a guard that cannot fail (D-14.4.3), sitting inside
        the assertion that guards the absence claim. The fence now asserts **both** spellings; the
        `hidden: true` one is what reds.
      - The **accessible-text assertion never fires** on this mutation and structurally cannot: the
        clone-and-strip step removes the `[aria-hidden="true"]` subtree before reading
        `textContent`, so a name added under a still-present `aria-hidden` leaves it unchanged.
      - Measured: reds `App.test.tsx:1086` and `LoadScreen.test.tsx:132` — the `hidden: true` role
        sweep — at 2 failed / 363 passed.
      - **Proof 3b, `aria-label` only with no `role`, run separately:** reds at both sites on a
        *different* assertion, the contributed-name sweep. The two halves fire independently.
      - **Hole found at review and patched:** every name sweep queried *descendants* of the lockup,
        so a name on the lockup element itself escaped all of them — `role="img" aria-label="Folio"`
        on `<span className="brand-lockup">` passed **373/373**. On the load screen that is a real
        regression: `role="img"` makes children presentational, so AT would announce "Folio" instead
        of "FOLIO / OFFLINE". The sweep now includes the lockup element itself.
- [x] **ADD a second drawing.** Replace `<BrandMark size={22} />` in `LoadScreen.tsx` with a
      hand-written inline `<svg>` using the mockup's rounded integers (inner 7×10). Expect the
      exactly-one-production-file assertion **and** the literal geometry assertion to red. Revert.
- [x] **DELETE the token reference.** Remove `.brand-mark { color: var(--color-select) }` from
      App.css. Expect the App.css source-text pin to red. Revert.
- [x] Presence proof (revert is valid here): remove `<BrandMark size={18} />` from `App.tsx` and
      expect the document-bar test to red. Revert.

**Acceptance Criteria:**
- Given the document bar, when it is shown, then the mark sits before the word FOLIO as a square
  outline containing a smaller centred solid block, 18×18 with a 1.5px stroke.
- Given the mark's colour, when it is implemented, then it arrives only as `currentColor` from
  `.brand-mark { color: var(--color-select) }`, and no colour literal is written in `BrandMark.tsx`,
  in the `App.css` rules this story adds, or at either call site — proved by a scan that reds when a
  hex is **added**. This AC is **partial preservation**: the App.css half is preserved (an incumbent
  guard already covers it), the `BrandMark.tsx` half is constructed (nothing covers `.tsx` today).
  Both halves get an ADD mutation.
- Given the mark at two sizes, when it is built, then one component parameterised by `size` produces
  both — **rather than a separate drawing per site**, `epics.md` AC3 as corrected at `f3d6c36` — no
  second drawing of the mark exists in any production file, and it is used on the load screen as well
  as the document bar.
- Given the mark, when it is read by assistive technology, then it carries `aria-hidden="true"`,
  exposes no role or accessible name of its own, and the lockup's accessible text at each site is the
  wordmark alone — proved at **both** sites by a mutation that **adds** a name.

## Spec Change Log

## Design Notes

### The geometry rule — one formula, and why the mockups are not three similar drawings

`Main.dc.html:24-25` draws 18px outer / 1.5px border / inner **6×8**. `Load.dc.html:24-25` draws 22px
outer / 1.5px border / inner **7×10**. `Load.dc.html:78-80` draws a 13px outer / inner **5×5**.
Measured, these are **not geometrically similar**: inner-width÷outer is 0.333, 0.318, 0.385 and
inner-height÷outer is 0.444, 0.455, 0.385. The 18 and 22 pair fits one rule; the 13 does not, and its
inner block is square where the mark's is portrait.

The rule that fits both declared sizes, with the stroke held constant as the mockups hold it:

```
stroke     = 1.5           (constant at every size — the mockups do not scale it)
outer rect = x,y 0.75      w,h size − 1.5      (inset by half the stroke, so the outer edge lands on the box)
inner rect = w size/3      h size×4/9          x size/3      y size×5/18
```

At 18 this is exact: inner 6×8 at 6,5 — the mockup's own integers. At 22 it gives 7.333×9.778 at
7.333,6.111 against the mockup's rounded 7×10, a difference of 0.34px and 0.22px. **A lookup table of
the mockup's two integer pairs would satisfy the pixels and violate the AC** — that is two drawings
wearing one function signature. The formula is the deliverable.

### Why SVG attributes rather than the house CSS-class sizing

Every existing glyph is sized by a CSS class. This one must not be, and the reason is testability:
**vitest runs in jsdom, which parses no stylesheet and computes no layout**, so any correctness placed
in CSS has zero executable coverage in this project (the `.canvas-component-echo` specificity defect
shipped with a fully green suite for exactly this reason). If `size` reached the mark through a class,
"one component parameterised by size" would be unprovable by the suite. As SVG `width`/`height`/
`viewBox`/`x`/`y` attributes it is directly readable in jsdom, and the parameterisation becomes a
literal assertion. The colour still lives in CSS — it has to, to stay under the token rule — and that
half is therefore pinned by source text and named in Verification as not behaviourally proven.

### Observed and deliberately untouched — not omissions

Two mismatches against the mockups sit inside the surfaces this story edits. Both were measured, both
are outside the fence, and both are recorded here so a reviewer does not report them as oversights.

- **The load screen's wordmark reads `FOLIO / OFFLINE`** (`LoadScreen.tsx:24`) where `Load.dc.html:27`
  draws `FOLIO` alone. No AC asks for a copy change, and shipped wording is a content decision that is
  not this story's. AC4 is therefore read as *"the mark adds no second announcement"*, and the
  load-screen assertion expects the accessible text `FOLIO / OFFLINE`, not `FOLIO`.
- **`.load-brand` is `var(--color-select)`** (`App.css:594`) where the mockup's load-screen wordmark is
  `#e6e9ec` = `--color-ink-high`. Also unnamed by any AC. Left as shipped.

### The colour-literal ban is narrower than it looks — DW-358 and DW-359

AC2's "no hex is written anywhere in the app" is **not** satisfied today, so it is specced as partial
preservation rather than as preservation. Two registered deferrals hold the remainder, and this story
closes neither:

- **DW-358** — `design-contract.test.ts:87` is too strict and too loose at once. Its regex
  `/#[0-9a-f]{3,8}|\b(?:rgb|hsl)\(/i` requires `(` **immediately** after `rgb`/`hsl`, so `rgba(` and
  `hsla(` cannot match; and it reads the raw file without stripping comments, so a hex inside a CSS
  comment reds it. It reads `src/App.css` alone.
- **DW-359** — closing UX-DR1 repo-wide needs a policy on which colours are chrome and which are
  document model. `App.tsx:4700` and `swatch-color.ts:20` are the document's own `'#000000'` default
  mirroring the Go engine and must not be tokenised, so the eventual ban can never be absolute; it has
  to be an allowlist with stated reasons.

### Asset-slot cost: zero, and why that is a claim rather than a formality

D-13.6.7 requires every Epic 14 story to state its slot cost. This one is **zero**: the mark is inline
SVG, adds no file Vite emits as an asset, and touches only `.tsx`, `.css` and test/e2e files. The
ruling's usual check is a clean `npm run build` re-measuring `s1.assetCount`, but `npm run build` is
outside the per-story cadence (D-000.33). The claim is therefore checked structurally instead — the
diff adds no file outside `src/**/*.tsx` and `e2e/**/*.ts` — and `npm run build` is named below as not
run. The release currently carries 62 assets against a warning threshold of 56, so this story spends
none of that margin.

## Verification

**On the two commit ids below — not an error.** The baselines in this section were
measured at `630ecc3` while this story's `baseline_commit` frontmatter is `f3d6c36`. The two are
equivalent for the purpose of every count here: `git diff --name-only 630ecc3 f3d6c36 -- folio-designer/`
reports **0 files**, so nothing under the designer moved between them and the 74/1251/0 baseline
holds at either id. Verified this run.

**Commands (D-000.33 cadence — run all four from `folio-designer/`):**
- `npx vitest run` -- expected: 0 failures. **Baseline re-measured at `630ecc3`: 74 files / 1251 tests
  / 0 failures.** Report the new counts and diff the test-name list against the baseline as a
  **multiset** — duplicate `it.each` titles mean a set diff silently absorbs a lost test. Report
  GONE/NEW with counts.
- `npx tsc -b --force` -- expected: no output, exit 0. `--force` is mandatory. Baseline: clean.
- `npx oxlint` -- expected: 0 errors and **exactly 4** `only-export-components` warnings.
  **Re-measured at `630ecc3`: 4, anchored at `src/preview/pdf-viewer.tsx:17,18` and
  `src/App.tsx:4462,4469`.** Re-measure rather than quoting; a fifth warning means `BrandMark.tsx`
  gained a second export.
- `npm run test:e2e:compile` -- expected: exit 0. Baseline: clean.

**What is proven and what is not.**

Proven by the commands above: the component exists; it is used at both call sites; `size` is a real
parameter producing the two declared geometries; the outer and inner rects carry `currentColor` and
`fill="none"`; the mark carries `aria-hidden` and exposes no name at either site; no colour literal
exists in the component or in the App.css rules added; no second drawing exists in any production
file — each of these tied to a mutation that was executed and observed to red.

**Not proven:** that the mark *looks right*. A square outline rendering crisply at 18px with a 1.5px
stroke, the optical centring of the inner block, and the fact that `--color-select` actually paints
cyan are all beyond jsdom, which applies no stylesheet and computes no layout. `e2e/brand-mark.spec.ts`
is the only witness for the rendered box, and it is **compiled here, scheduled to execute at the epic
boundary gate on push** — it is not run by this story and must not be reported as covered.

**Suites that did not run, in these words:** the browser suite, the Go suites, the matrix legs,
`npm run build` as a gate, the `verify:offline*` chain, and the font-host scans.


## Suggested Review Order

**The one drawing**

- The geometry rule and why the mockup's integers were consciously not matched.
  [`BrandMark.tsx:9`](../../folio-designer/src/BrandMark.tsx#L9)

- The whole component: `size` in, two rects out, colour deferred to `currentColor`.
  [`BrandMark.tsx:50`](../../folio-designer/src/BrandMark.tsx#L50)

**The two call sites, and only two**

- 18px in the document bar; `.brand` stays byte-identical inside the new lockup.
  [`App.tsx:2377`](../../folio-designer/src/App.tsx#L2377)

- 22px on the load screen, from the same component; wordmark copy untouched.
  [`LoadScreen.tsx:35`](../../folio-designer/src/LoadScreen.tsx#L35)

- The fence that makes "two sizes, nowhere else" executable rather than a comment.
  [`BrandMark.test.tsx:203`](../../folio-designer/src/BrandMark.test.tsx#L203)

**The colour, which has exactly one source**

- The token reaches the mark here and by no other route.
  [`App.css:57`](../../folio-designer/src/App.css#L57)

- The specificity sweep, with line numbers that sit below the comment itself.
  [`App.css:35`](../../folio-designer/src/App.css#L35)

- The class attribute is the middle link; deleting it used to leave the suite green.
  [`BrandMark.test.tsx:98`](../../folio-designer/src/BrandMark.test.tsx#L98)

- The comment-stripped literal scan, with a two-way discrimination control.
  [`BrandMark.test.tsx:127`](../../folio-designer/src/BrandMark.test.tsx#L127)

**Decorative, and provably so — the absence claim**

- The document bar fence: the sweep includes the wrapper, not just its descendants.
  [`App.test.tsx:1072`](../../folio-designer/src/App.test.tsx#L1072)

- The same fence at the second site, plus the rendered geometry literals at 22.
  [`LoadScreen.test.tsx:108`](../../folio-designer/src/LoadScreen.test.tsx#L108)

- One drawing, whatever element a copy might use — widened past `<rect`.
  [`BrandMark.test.tsx:176`](../../folio-designer/src/BrandMark.test.tsx#L176)

**Peripheral — the browser witness that has not run**

- Box, order and colour in a real browser; compiled, unexecuted until the boundary push.
  [`brand-mark.spec.ts:22`](../../folio-designer/e2e/brand-mark.spec.ts#L22)


## Delivery Log

### 2026-09-09 — done

Baseline `f3d6c36`. Shipped as one component drawn once and parameterised by `size`, placed at exactly
two call sites — 18px in the document bar, 22px on the load screen — with every coordinate derived
from the single geometry rule rather than transcribed from the mockups, so the two renderings cannot
diverge. Colour reaches it as `currentColor` from one App.css rule over `var(--color-select)`; no
colour literal was added. The epic's third size was **correctly excluded**: measurement identified the
13px load-screen file-row shape as the in-progress status marker sitting between a tick and a dash,
with a square inner block where the mark's is portrait, so building it as written would have enrolled
the logo into a status vocabulary. That exclusion is a narrowing of AC3, not a gap — `epics.md` was
corrected for it at `6fa386d` and `f3d6c36` before implementation began.

**Triage:** 12 patched / 2 deferred / 6 rejected. Both deferrals were registered by the engineering
lead as **DW-360** (a constant `viewBox` would make the mark literally one set of coordinates) and
**DW-361** (the brand is painted with the selection token, so retheming selection would retint the
logo). `intent_gap` 0, `bad_spec` 0, `review_loop_iteration` 0 — the spec survived implementation
without renegotiation.

**What the review caught — the story's real lesson.** The fence guarding AC4's absence claim was
**inert twice**, and neither failure was visible by reading it. First, Testing Library's role queries
skip `aria-hidden` subtrees, so the fence stayed green against the exact mutation it existed to
catch — the mark could be given a name and the guard would not notice. Second, every sweep queried
*descendants* of the lockup, so a name placed on the wrapper itself passed **373/373**. That second
direction is the harmful one: `role="img"` on the wrapper makes its children presentational and would
silence "OFFLINE" on the load screen, which is the one string that screen exists to say. Both defects
appeared only on execution. This is recorded as **D-14.5.1** (a descendant sweep does not cover the
element it descends from), itself a refinement of **D-14.4.3** (an absence claim cannot be red-proved
by reverting the implementation — it needs a mutation that ADDS the forbidden thing).

**Measured gates**, re-run at `b505260` from `folio-designer/` with captured exit codes: `npx vitest
run` exit 0 — **75 files / 1263 tests / 0 failures** (baseline 74/1251/0 at `630ecc3`; the +1 file is
`BrandMark.test.tsx` and the +12 tests decompose by name as 10 added there, 1 in `App.test.tsx`, 1 in
`LoadScreen.test.tsx`, with 0 removed). `npx tsc -b --force` exit 0, zero bytes of output. `npx
oxlint` exit 0, 0 errors and exactly 4 `only-export-components` warnings — re-anchored this run to
`src/preview/pdf-viewer.tsx:17,18` and `src/App.tsx:4469,4476`; the App.tsx pair moved down 7 lines
from the `4462,4469` measured at `630ecc3`, which is this story's own insertion and not a new warning.
`npm run test:e2e:compile` exit 0.

**What is not coverage.** `e2e/brand-mark.spec.ts` is compiled and has never executed against a
browser — it is scheduled for the epic boundary gate on push and is not coverage until then. The
App.css colour rule is pinned by **source text only, not behaviourally**: the review demonstrated that
appending `.document-bar svg { color: … }` left the suite at 452/452 green, so the guard proves the
declaration is written, never that the token is what actually paints. And **"it looks right" is not
proven** — jsdom applies no stylesheet and computes no layout, so the crispness of a 1.5px stroke at
18px, the optical centring of the inner block, and the fact that `--color-select` renders as the
design's cyan are all beyond every gate that ran.

**Suites that did not run, in these words:** the browser suite, the Go suites, the matrix legs, `npm
run build` as a gate, the `verify:offline*` chain, and the font-host scans. `npm run build` matters
here beyond routine: D-13.6.7's asset-slot check normally re-measures `s1.assetCount` through it, so
this story's zero-slot claim rests on the structural argument in the Design Notes rather than a
measured count, against a release already carrying 62 assets versus a warning threshold of 56.

Deferred, with owners: **DW-360** and **DW-361** both stand open in `deferred-work.md`, owner
unassigned and neither closed by this story. **DW-358** and **DW-359** continue to hold the remainder
of AC2's colour-literal ban; this story closed neither and did not narrow either.

Commit `b505260` — "Give the product its mark, at two sizes from one drawing". Tracker advanced
`review` → `done`. `epic-14` deliberately stays `in-progress`: the boundary-gate suites above have not
run.
