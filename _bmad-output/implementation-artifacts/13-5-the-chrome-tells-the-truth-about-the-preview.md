---
title: 'Story 13.5: The chrome tells the truth about the preview'
type: 'feature'
created: '2026-09-08'
status: 'done'
baseline_commit: '0a1f1f73a557d672c944772b19fc22540b87a186'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** In Preview mode the frame around the page still describes Design mode. The document bar
reads `A4 · portrait` — a page-setup fact that says nothing about the render being looked at — while
render freshness lives only inside a paragraph in the preview pane, and goes on saying nothing about
how long ago that render happened. The application status bar never states the product's central
promise. And the preview heading carries a second way back to Design beside the document bar's mode
switch, so there are two mode controls where the design draws one.

**Approach:** In Preview mode the document bar's page-setup slot reads the render's own freshness —
a ticking `rendered <age> · current` / `rendered <age> · stale`; the status bar carries the standing
`no network · nothing left this machine` assurance, in space freed by dropping two Design-mode items
from the Preview bar; and the preview heading's Return-to-Design button is removed, leaving the
document-bar mode switch as the sole mode control. The switch already performs the identical, complete
cancellation, so the ability to abandon a render in progress survives the removal.

## Boundaries & Constraints

**Always:**

*Freshness is produced once, in one place.*
- **Both freshness renderings come from ONE exported function in `folio-designer/src/preview/freshness.ts`**, returning the document bar's short token and the status line's long copy together from one switch. **No call site may construct either string independently.** This makes "chrome and status line never disagree" true by construction rather than by review.
- **The status line's copy is UNCHANGED.** `Current exact local PDF` / `Current no-data layout PDF` / `staleCopy(...)` / `Rendering local PDF` / `Local Preview work failed…` / `Preview is waiting for local inputs` all keep their exact present wording.
- **`preview/freshness.ts` stays a `.ts` file.** A value export from a `.tsx` adds a fifth `react(only-export-components)` warning against a baseline of exactly 4.
- **The function takes `standIn` as an input even though the token does not vary on it**, so a future story that wants the bar to distinguish a no-data preview cannot do it by writing a second string elsewhere.
- **For a `standIn` preview whose status is `current`, the bar's head word is still `current`.** The token is a FRESHNESS claim and must never carry an exactness claim. Story 13.4's exactness disclosure stays exactly where 13.4 put it — the heading's `NO-DATA LAYOUT PREVIEW`, the status line's `Current no-data layout PDF`, and the stand-in notice. Do not add a second, narrower disclosure to the bar, and do not suppress the token.

*The age figure.*
- **AD-17 holds and is not in the way.** `ARCHITECTURE-SPINE.md:361-372` bans the browser measuring **text** — every metric and line break comes from the engine's measure API. A clock read is not a text metric, and `canvas-authority-contract.test.ts` names no clock API.
- **The browser stamps the time, not the engine.** `engine-protocol.ts:414`'s `isPreview` is a closed `hasOnly` allowlist, and a wasm-side clock would be the engine's time, not the browser's.
- **`elapsedMs` and the render age are DIFFERENT QUANTITIES and must never be conflated.** The engine's `elapsedMs` is how long the render took, already shown by the evidence rail as `elapsed` (`preview/evidence-rail.tsx:85` via `formatElapsed`, `preview/evidence-rail-facts.ts:63`). The new figure is how long ago that render finished. They print in the same 10px mono and are equal for exactly one instant after a render. **Do not reuse `formatElapsed`** — it has no minutes tier, and sharing it is exactly the conflation this boundary forbids.
- **The age formatter is pure and takes `(installedAt, now)`.** Only the ticking effect reads `Date.now()`. This is the repo's own idiom — `font-source.ts:403` already injects a clock read as a defaulted parameter — and it makes the ladder testable with zero fake timers.
- **Ladder and tick period, exactly:** `< 1000 ms` → `${ms} ms ago`, tick 100 ms · `< 60 s` → whole seconds, tick 1000 ms · `< 60 min` → whole minutes, tick 60000 ms · `>= 60 min` → whole hours, tick 60000 ms. **Invariant: the figure is never wrong by more than the unit it prints.**
- **Clamp a negative age to 0.** `Date.now()` is not monotonic; an NTP step must never print `-3 s ago`.
- **The interval mounts ONLY while `mode === 'preview'` AND a `PreviewRecord` exists**, in a `useEffect` whose cleanup clears it. **Design mode must never hold a live interval.**
- **The document-bar freshness element must NOT be `role="status"` and must NOT be `aria-live`.** A ticking live region announces the age every second. Its neighbours in that bar are full of live regions and copying the neighbour is the obvious mistake. It also takes no `title` and no `aria-describedby`: no AC requires either, and a cross-region ARIA association breaks silently if either element moves.

*The status bar.*
- **In Preview mode ONLY, exactly two items are dropped — `LOCAL SHELL` and `template-font-count` — and nothing else.** `engine-snapshot`, `offline-status` (visible, unchanged, still `role="status"`) and `PREVIEW MODE` all stay. **Design mode's status bar is untouched.**
- **The Preview status bar renders at exactly 32px.** `e2e/preview-navigation.spec.ts:131` asserts `statusBar.height === 32`; the token is `--status-bar-height-preview` (`tokens.css:16`), applied by `.app-shell-preview` (`App.css:28`). Nothing added may wrap or grow it.
- **Anything added to the status bar is non-interactive.** `App.test.tsx:7101-7102` maps every `button, input, select` in the bar to `aria-label` and asserts equality with an exact 7-element list.
- **`App.css` may not gain a second `@media` rule.** `canvas-authority-contract.test.ts:344` asserts the extracted media-query list equals exactly `['prefers-reduced-motion: reduce']`. A responsive breakpoint is not available as a fix.
- **`App.css` may carry no literal colour and no second use of the display/numeric-lg type tokens.** `design-contract.test.ts:84-90` bans `#hex`/`rgb(`/`hsl(`; `:145-150` pins `var(--type-display)` and `var(--type-numeric-lg)` to exactly one occurrence each. `--color-ink-ghost` (`tokens.css:6`) is the mockup's assurance colour and already exists, so **no new design token is needed** — which avoids `design-contract.test.ts:21-23`'s exact token-name equality against DESIGN.md.

*The heading button.*
- **The freshness paragraph keeps its exact attribute substring, in order.** `preview/preview-authority-contract.test.ts:22` pins `id="preview-freshness-status" className="preview-status" role="status"` as one contiguous string. It sits on the **same physical source line** (`App.tsx:2188`) as the button being removed, so the edit must not disturb it.
- **`#preview-freshness-status` text is asserted with `toHaveText` (exact)** at `e2e/application-shell.spec.ts:43`. Its text may not gain a suffix.
- **Every surviving `Return to Design` selector must be scoped** with `within(...)` on the failure card. After the heading button is gone, an unscoped `getByRole('button', { name: 'Return to Design' })` silently retargets the failure card's identically-named button (`preview/diagnostic-presenter.tsx:39`) — a guard that keeps passing while testing a different control.

*Process.*
- **NEVER COMMIT.** Do not run `git add`, `git stash`, `git checkout`, `git reset`, `git revert`, `git restore`, or any other git write command. Never push and never create a branch. Reading git state (`git log`, `git status`, `git show`, `git diff`) is permitted and encouraged. The single carve-out is `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`, removed in a `finally`, for a test fixture.

**Ask First:**
- Any change to a file the Execution list does not name.
- Widening any prohibition carve-out in `canvas-authority-contract.test.ts`.
- Any new field on the wasm preview payload (`engine-protocol.ts:414`'s closed `hasOnly` allowlist).
- Dropping, hiding or renaming any status-bar item other than the two this spec names.
- Any change to the status line's existing copy.

**Never:**
- No engine change of any kind. No `folio-go` edit, no wasm-boundary edit, no golden-corpus edit.
- Never touch `fixtures/declared-variants/expected.pdf`, `input.folio`, or `signoff.json`.
- Never edit `sprint-status.yaml`, `deferred-work.md`, or `epics.md`.
- **Never touch `folio-designer/e2e/preview-no-data.spec.ts`.** It is another story's, it was rewritten after this spec's baseline, and it is not a signal about this work.
- Do not fix DW-300, DW-301, DW-302 (evidence-rail follow-ups), DW-303, or DW-304 (the PAGES rail).
- Never remove or rename the failure card's own `Return to Design` button (`preview/diagnostic-presenter.tsx:39`) — it is a different control with the same name.
- Never introduce `getBoundingClientRect`, `offset*`/`client*`/`scroll*` dimension reads, `ResizeObserver`, `getComputedStyle`, `measureText`, or `document.fonts` anywhere. Geometry assertions use Playwright's `boundingBox()`.
- Do not change the document bar's Open/Save/Undo/Redo cluster, and do not build the PAGES thumbnail rail.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Design mode, bar | `mode === 'design'` | Document bar reads `${preset} · ${orientation}`; status bar carries `LOCAL SHELL` and `n fonts in template` and **no** assurance line | `Page setup unavailable` when `canvas` is absent, unchanged |
| Fresh render | Preview, `current`, installed 412 ms ago | `rendered 412 ms ago · current`; status line `Current exact local PDF` | N/A |
| No-data render | Preview, `current`, `standIn` | `rendered <age> · current` — head word unchanged; status line `Current no-data layout PDF` | N/A |
| Age ladder | 999 / 1000 / 59_999 / 60_000 / 3_599_999 / 3_600_000 ms | `999 ms ago` / `1 s ago` / `59 s ago` / `1 min ago` / `59 min ago` / `1 h ago` | N/A |
| Clock steps backwards | `now < installedAt` | `0 ms ago` | Clamp to 0; never a negative figure |
| Inputs changed | Preview, `stale`, `inputs-changed`, record present | `rendered <age> · stale`; status line `STALE — inputs changed` | N/A |
| Render failed | Preview, `stale`, `render-failed`, record present | `rendered <age> · stale`; status line `STALE — latest local render failed` | N/A |
| Digest mismatch over a good preview | Preview, `error`, record present | `rendered <age> · stale` — the bar never affirms a render the screen refused; status line `Local Preview work failed…` | The bar is conservative where the status line is specific; neither says `current` |
| Rendering, no prior PDF | Preview, `checking`/`debouncing`/`rendering`, no record | `no render yet` — no age, no token, because no render has finished | N/A |
| Idle | Preview, `idle`, no record | `no render yet` | N/A |
| Leave and re-enter | Preview → Design → Preview, record survives | Age keeps counting from the original render's stamp, not from re-entry | N/A |
| Cancel mid-render | Preview, render in flight, author presses DESIGN | Token bumped, `AbortController` aborted, debounce cleared, scheduler cleared; a late-arriving result is **never installed** | N/A |
| Assurance line fits | Preview at a 1024px-wide viewport | The assurance element's right edge <= the VIEWPORT's right edge and the leftmost surviving item's left edge >= the viewport's left edge, measured from `page.viewportSize()`. **Never against the status bar's own box:** `.app-shell` is `display: grid; overflow: hidden` and `.status-bar` is its item, so an overfull `nowrap` bar grows its border box and both sides of that comparison move together. AMENDED BY THE ORCHESTRATOR 2026-09-08 — see Spec Change Log | Restoring **both** dropped items must make this fail. AMENDED 2026-09-08 (second amendment): this clause read "either dropped item", which the owner's ruling to hide `offline-status` made false — that returned ~90px plus its 12px gap, taking the margin to 124.94px, so a restored 78px `LOCAL SHELL` now genuinely fits with 46.94px to spare. The governing property is that the assertion is **non-vacuous** — it must red on a bar that overflows the viewport — not that any particular single item overflows it. Measured: unmutated GREEN 1024.00 (margin 124.94) · `LOCAL SHELL` restored GREEN 1024.00 (margin 46.94) · both restored RED 1097.06 > 1024 |

</frozen-after-approval>

## Rulings resolved at the plan gate

Recorded verbatim so the implementer inherits the reasoning, not a pointer to it. Two premises argued
during the gate were **false and are corrected here** rather than left standing.

**CORRECTION 1 — "dropping `offline-status` from Preview costs two e2e specs" is FALSE.** Population
searched: `e2e/offline-update.spec.ts` and `e2e/engine-worker.spec.ts`, the only two specs reading that
testid. Neither ever enters Preview mode; both exercise Design only. **The e2e cost is zero.** The real
and unchanged reason to keep it is accessibility: `offlineLabel` (`App.tsx:2018`) has five states
including `'Update available; current release remains usable'` and `'Offline cache unavailable'`, which
can transition while an author sits in Preview.

**CORRECTION 2 — the "genuine first" framing overstated the novelty of a ticker.** The counts are
right: `Date.now()` appears 0 times in all of `folio-designer/src`, and `setInterval` 0 times in `src`
and `e2e` outside the vendored wasm shim. But `requestAnimationFrame` has two precedents
(`App.tsx:2205`, `DataPanel.tsx:45`), so a browser scheduling API is not novel here. **"It's a first"
is not an argument against the ticker.**

**Q1 — RULING: drop exactly two items.** AC4 forces *some* removal, because the bar is full — that part
is inside the fence. *How much* is not forced. The two-item drop fits with 35px spare, so anything
beyond that minimum is product-surface change no AC asks for. The epic cites `Preview.dc.html` for **the
assurance string only**, not for the mockup's whole composition. No single drop closes the 163px deficit
(the largest, `engine-snapshot`, is 156); of the two-item sets that fit, `LOCAL SHELL`+`offline` kills
the live region and `LOCAL SHELL`+`PREVIEW MODE` misses by 1px. `LOCAL SHELL`+`font count` (198px) is
the only minimal set preserving both testid-bearing elements, and it is independently justified — the
assurance line states in words what `LOCAL SHELL` states in shorthand, and the font count is a template
fact whose render-side siblings moved to 13.3's evidence rail.

**Q2 — RULING: one function; the premise that `epics.md` needs amending is REJECTED.** The proof is
inside the AC set: AC1 mandates the bar read `· current`, and the status line contains no bare
`current` either. If *"the same words the status line uses"* meant string identity, **AC1 and AC5 would
contradict each other on the fresh case, in the same story, before any pixel is measured.** The reading
that uses every clause is the one where the bar carries the same HEAD WORD and the status line carries
the qualifier, with AC5's purpose clause — *"so the two never disagree"* — as the operative test,
satisfied by construction under the one-function rule. **The planning record stays true as written. Do
not touch `epics.md`.**

**Q2 AMENDED AT CHECKPOINT 1 — the "prefix-case-match for all seven states" test rule was WITHDRAWN,
and the formulation in the Tasks below is the RULED one.** The original clause was jointly
unsatisfiable with "the status line's copy is UNCHANGED": measured against the actual copy at
`App.tsx:2188`, five states match but `error` reads `Local Preview work failed` and `idle` reads
`Preview is waiting for local inputs`, whose head words `Local` and `Preview` are not freshness tokens.
The stronger reason: **`error` can coexist with an installed record** — `App.tsx:664`'s digest mismatch
sets `error` *without* installing, over a surviving preview — so for error-with-a-record the bar must
read `stale`, deliberately diverging from the status line's more specific copy. That is the bar being
conservative about a render the app refused, which is the exact failure Story 13.3 spent a patch on
(the evidence screen affirming a render it had just refused for corruption). **A rule forcing the bar
to mirror the status line there would have re-introduced it.** The replacement keeps every property the
original was protecting: one producer, state-list-driven, exhaustive, non-vacuous in both directions,
no hand-maintained mirror.

**Q3 — RULING: it ticks.** Not ticking makes AC1's stated output true for one instant and false forever
after, in the story named "the chrome tells the truth about the preview"; dropping the figure abandons
AC1's own worked example. Neither makes AC1 pass.

**Q4 — RULING: no compensating copy.** AC3 requires the ABILITY to abandon a render, not the wording,
and the ability is proved in the Code Map. **Consequence to disclose, not to fix here:** after this
story nothing anywhere tells an author that a render can be abandoned. That is a shortfall for the
orchestrator to register; 13.5 does not absorb a fix for it.

**Open, and NOT blocking:** whether to adopt the mockup's full Preview status bar instead of this
minimal drop is with the owner. An owner answer folds in without rework — three more `mode === 'preview'`
drops plus one `.sr-only` span. **Build the two-item drop now.**

## Code Map

Every anchor re-measured at **`ab3fc1a`**, the dispatch baseline. `git diff --stat 0d395ca ab3fc1a` is
one file — `deferred-work.md`, 38 insertions / 11 deletions — with **no `folio-designer/` change**, so
the browser measurements taken at `7280caa` and the source anchors both still hold; do not re-verify
from scratch. **`folio-designer/src/App.tsx` contains two literal NUL bytes near lines 3201–3202, so the
whole file is invisible to `grep -I` — use `grep -a` for every search in this repository, and state the
population searched in the same sentence as any absence reported.**

**The four edit sites in `App.tsx`**
- `:155` — `type PreviewRecord = Readonly<{ bytes; revision; identity; digest; diagnostics; token; generation; standIn; elapsedMs; version }>`. **No completion timestamp of any kind.** The docblock at `:150-154` already states the rule a stamp must obey: these fields "describe the render that produced THESE bytes, so a stale record keeps reporting its own render's numbers."
- `:667` — `installPreview({ …, elapsedMs: result.preview.elapsedMs, version: result.preview.version })`, the **single** install site.
- `:2163` — `<span className="later-control" aria-label="Current page setup">{canvas ? \`${canvas.preset} · ${canvas.orientation}\` : 'Page setup unavailable'}</span>`. Rendered in **both** modes; a Preview-only swap needs a `mode` branch. **Unguarded:** no test in `src/**/*.test.ts(x)` or `e2e/*.spec.ts` references `later-control`, `Current page setup`, or `Page setup unavailable`.
- `:2188` — one physical line holding `.preview-heading`, its `<p>`, the Return-to-Design button, the pinned `#preview-freshness-status` paragraph, the stand-in notice, the failure card and the viewer. Edit surgically.
- `:2254` — the whole `<footer className="status-bar" aria-label="Status bar">`. Children in order: `LOCAL SHELL` · `<code data-testid="engine-snapshot">` · `{mode === 'preview' && previewNavigation}` · `.status-spacer` · `template-font-count` · `offline-status` (a `role="status"` live region) · `<code>{mode.toUpperCase()} MODE</code>`.

**Measured geometry at a 1024px viewport, in Preview** (Chromium 1217, `7280caa`; the arithmetic every
sizing decision here rests on)
- Status bar: `.status-spacer` = **76.9px**, the only free space. Assurance line = **228px** + one 12px gap = **240px**. Deficit **163px**. Dropping `LOCAL SHELL` (66+12) and `template-font-count` (108+12) frees **198px**, leaving **35px** spare.
- Other items, each with its gap: `GO SNAPSHOT · REVISION n` 156 · `Offline ready` 90 · `PREVIEW MODE` 84.
- Document bar: **179px** slack + the 78px `.later-control` = **257px available**. `rendered 412 ms ago · current` = **174px** (83px spare). For contrast, `rendered 4 min ago · STALE — inputs changed` = 258px and `rendered 412 ms ago · Current exact local PDF` = 270px — both overflow, which is why the bar carries the head word and not the qualifier.
- Both bars compute to `10px / 13px "IBM Plex Mono", ui-monospace, monospace`.

**Cancellation — AC3's evidence**
- `:769` — `returnToDesign` is one line: `cancelPreviewWork(); modeRef.current = 'design'; setPreviewStatus(previewRef.current ? 'stale' : 'idle'); setPreviewError(undefined); setMode('design')`.
- `:562-565` — `cancelPreviewWork`: bumps `previewToken`, calls `previewAbort.current?.abort()` (the controller passed to every `engine.request` in `runPreview`), clears the debounce timer, clears the scheduler's pending job.
- **Three author-reachable exits, all calling `returnToDesign`:** the heading button (`:2188`), the document-bar DESIGN button (`:2166`), and `Alt+P` (`:1961-1966`). The DESIGN button's `onClick` is the *same reference*; removing the heading button removes no capability. Its one asymmetry: the heading button unmounts itself so focus falls to `<body>`, whereas DESIGN persists and keeps focus.
- **The cancel wording is already nearly unreachable.** Every status setter uses `previewRef.current ? 'stale' : X`, so once any PDF is installed `checking`/`debouncing`/`rendering` are unreachable and `Cancel and return to Design` appears only on the first render of a session. `DataPanel.test.tsx:47-52` and `App.test.tsx:7562-7564` both document this with case-insensitive regexes.

**Freshness state and vocabulary**
- `folio-designer/src/preview/freshness.ts` — `PreviewFreshness` (7 states, `:6`), `StaleReason` (`:7`), `staleCopy` (`:9`) returning `'STALE — inputs changed'` / `'STALE — latest local render failed'` (em dash U+2014, pinned at `freshness.test.ts:15-16` and `e2e/application-shell.spec.ts:44`). `PREVIEW_DEBOUNCE_MS = 250` (`:4`). **This is where the shared function goes** — already a `.ts`, already owns `staleCopy`.
- `App.tsx:2188` — status-line copy per state, all unchanged by this story.
- `App.tsx:219-220` — `previewStatus`, `staleReason`. `App.tsx:785` (`viewerPages`) is the **only** path to a durable `'current'`.
- **Which states can coexist with a record:** with a record installed, `current`, `stale` and `error` are reachable (`:664`'s digest mismatch sets `error` without installing, over a surviving previous record). Without a record: `idle`, `checking`, `debouncing`, `rendering`, `error`.
- ⚠ `engine-ownership-contract.test.ts:11,25` flags any type literal carrying **>=2** of `version`/`page`/`bands`/`elements`/`assets`. `PreviewRecord` already has `version`, which is why the new field is named `installedAt`.

**The quantity that must not be confused with it**
- `preview/evidence-rail-facts.ts:63-68` — `formatElapsed`: `< 1000` → `` `${ms} ms` ``; otherwise one-decimal seconds with `.0` trimmed. **No minutes tier.** Rendered as `elapsed 412 ms` at `preview/evidence-rail.tsx:85`, fed from `App.tsx:2226`.

**Timer precedent**
- `App.tsx:305`, `:700-705` — the debounce: a `useRef` handle, cleared-then-armed imperatively, torn down at `:561`/`:591`. There is **no** unmount-cleanup effect for it; a ticking display needs the opposite shape.
- `App.test.tsx:4827` / `:7696` — the suite's only time helper: `const elapse = async (ms) => { await act(async () => { await vi.advanceTimersByTimeAsync(ms) }) }`. `vi.useFakeTimers` appears 3 times (`:701`, `:4846`, `:7734`); `vi.setSystemTime` 0 times.
- `App.tsx:2205`, `DataPanel.tsx:45` — the two `requestAnimationFrame` precedents.
- `font-source.ts:403` — the injected-clock idiom: `today = new Date().toISOString().slice(0,10)` as a defaulted parameter.

**Guards that will move**
- `preview/preview-authority-contract.test.ts` — 13 forbidden tokens over `pdf-viewer.tsx` (`:7`); 5 positive `App.tsx` literals (`:17-20`, `:22`); 2 negatives (`:21`, `:24`); 2 red-proof mutations (`:29-30`). Only `:19` and `:22` touch the freshness `<p>`.
- `canvas-authority-contract.test.ts:762-766` — floors: `production >= 58`, `tests >= 51`, `e2e >= 15`, `.tsx >= 8`, `.css >= 3`. All `>=`, so **adding** files is safe.
- `canvas-authority-contract.test.ts:344` — exactly one `@media` in `App.css`.
- `App.test.tsx:7086-7091` — the Preview status-bar census: asserts `LOCAL SHELL`, `engine-snapshot`, `template-font-count`, `offline-status`, `PREVIEW MODE`, `Next PDF page` all present in Preview. **Rewritten by this story.**
- `App.test.tsx:7101-7102` — the exhaustive 7-label list of the bar's `button, input, select`.
- `App.test.tsx:7057-7060` — asserts `Zoom in` / `Zoom out` / `Canvas zoom` are absent in Preview.
- `App.test.tsx:701-715` — the **one** fake-timer block that enters Preview and installs a preview; `:708` and `:711` call `await vi.runAllTimersAsync()`, which does not terminate against a live rescheduling interval. `:4846` and `:7734` are Design-mode blocks — **verify that rather than assume it.**
- Broken by removing the heading button: `App.test.tsx:680`, `:966` (exact `'Cancel and return to Design'`), `:6485`, `:6600` (exact `'Return to Design'`, both **unscoped**); `e2e/application-shell.spec.ts:42` (asserts it is **visible** — an existence contract to retire deliberately) and `:46` (clicks it).
- Untouched by that removal (a different button): `App.test.tsx:876`, `:893`, `preview/diagnostic-presenter.test.tsx:71`, `e2e/preview-parameters.spec.ts:40`.
- `App.test.tsx:7198` — `getByRole('button', { name: 'DESIGN' })`, the only existing test that drives the surviving mode control.
- `e2e/preview-navigation.spec.ts:127-132` — the existing `boundingBox()` geometry idiom to copy, including `:131`'s `expect(before.statusBar.height).toBe(32)`.
- **Any new or amended `e2e` spec must force the fallback file tier first**: headless Chromium *has* the File System Access API, so `showOpenFilePicker()` is called and no `filechooser` event is ever emitted — a 90s timeout on a working app. `await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })`; model at `e2e/local-file-actions.spec.ts:11`. Eleven of the twenty specs already do this.

**CSS**
- `App.css:433` `.status-bar` (flex, `gap: var(--space-5)` = 12px, `padding: 0 var(--space-5)`, no wrap, no overflow, no min-width) · `:435` `.status-spacer { flex: 1 }` · `:440` `.preview-nav` (no `flex-shrink`).
- `App.css:29` `.document-bar` (flex, `gap: var(--space-3)` = 8px) · `:41` `.later-control { margin-left: auto }` · `:31` sets its mono type.
- `DESIGN.md:181-188` already declares the `status-bar` component group including both heights.

**The design being implemented**
- `…/mockups/Preview.dc.html:282-305` — the assurance string, last and right-most, in `#4e565f` (= `--color-ink-ghost`). **Cited for the string only, not for the composition.**
- `…/mockups/Preview.dc.html:20-45` — `rendered 412 ms ago` · `·` · `current`, then the mode switch.

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/preview/freshness.ts` — add two exports beside `staleCopy`. (1) `freshnessChrome({ status, staleReason, standIn, hasRecord })` → `{ token: 'current' | 'stale' | undefined; statusLine: string }`, one switch: `status === 'current'` → token `'current'`; `hasRecord && status !== 'current'` → token `'stale'`; `!hasRecord` → token `undefined`; `statusLine` reproduces today's copy exactly. (2) `formatRenderAge(installedAt: number, now: number): string` — pure, the four-tier ladder, negatives clamped to 0. Also export the tick period for a given age so the effect and the ladder cannot drift apart. — Rationale: one producer means chrome and status line cannot disagree; a `.ts` file costs no lint warning.
- [x] `folio-designer/src/preview/freshness.test.ts` — cover the ladder at `999`/`1000`/`59_999`/`60_000`/`3_599_999`/`3_600_000` ms, the negative clamp, the tick period per tier, and the **exhaustive state invariant**: iterating the seven `PreviewFreshness` values **from the module's own state list** (never a restated literal), token `'current'` if and only if `statusLine` starts with `Current`; token `'stale'` implies `statusLine` never starts with `Current`; token `undefined` exactly when `hasRecord` is false. — Rationale: a hand-maintained mirror is the failure this run has been bitten by three times.
- [x] `folio-designer/src/App.tsx:155,667` — add `installedAt: number` to `PreviewRecord`, stamped once with `Date.now()` at the single install site, never re-stamped. — Rationale: leave-and-re-enter then falls out free, because the stamp is the render's and not the entry's.
- [x] `folio-designer/src/App.tsx:2163` — branch `.later-control` on mode: Design keeps `${preset} · ${orientation}`; Preview renders `rendered <age> · <token>` from `freshnessChrome` + `formatRenderAge`, or `no render yet` when there is no record. **No `role="status"`, no `aria-live`, no `title`, no `aria-describedby`.** — Rationale: AC1; a ticking live region would announce the age every second.
- [x] `folio-designer/src/App.tsx` — the ticking `useEffect`: mounted only while `mode === 'preview'` **and** a `PreviewRecord` exists; period from the ladder tier; cleanup clears the interval. — Rationale: Design mode must never hold a live interval.
- [x] `folio-designer/src/App.tsx:2188` — remove the `.preview-heading` Return-to-Design button and route the status paragraph's copy through `freshnessChrome(...).statusLine`, **without disturbing** the pinned `id="preview-freshness-status" className="preview-status" role="status"` substring on that same line. — Rationale: AC2; `preview-authority-contract.test.ts:22`.
- [x] `folio-designer/src/App.tsx:2254` and `folio-designer/src/App.css` — in Preview mode only, drop `LOCAL SHELL` and `template-font-count`, and add a **non-interactive** `<span>` carrying `no network · nothing left this machine` styled with existing tokens. No new `@media`, no literal colour, no new design token. — Rationale: AC4 and the 163px deficit.
- [x] `folio-designer/src/App.test.tsx` — the test surface, all of it: rewrite the `:7086-7091` census to assert the three that stay **present in Preview** and the two that are **absent in Preview and still present in Design**; convert `:708`/`:711` from `runAllTimersAsync` to `advanceTimersByTimeAsync` (and **verify**, do not assume, that `:4846` and `:7734` are Design-mode); delete the `:680`/`:966` heading-button clicks; scope `:6485`/`:6600` with `within(...)` on the failure card; add the first test of the Design-mode `${preset} · ${orientation}` string; add the ticker and no-record tests. — Rationale: the Design-mode absence assertion is what stops the drop leaking out of Preview; the scoping is what stops a guard passing against a different control.
- [x] `folio-designer/src/App.test.tsx` — add a **cancellation** test: with a render in flight, pressing DESIGN bumps the token, aborts the controller, clears the debounce and scheduler, and a late-arriving result is **never installed**. — Rationale: AC3. Asserting only that the mode changed is a test that passes on a broken cancel.
- [x] `folio-designer/e2e/application-shell.spec.ts` — amend `:42` (the button no longer exists) and `:46` (drive DESIGN instead). — Rationale: AC2.
- [x] `folio-designer/e2e/preview-navigation.spec.ts` — add the 1024px geometry assertion in the `boundingBox()` shape `:127-132` already uses. **CORRECTED AT REVIEW — see Spec Change Log entry 1.** The observable is the **viewport**, never the status bar's own box: `bar.width <= viewportWidth`, `assurance.right <= viewportWidth`, `bar.x >= 0`, `leftmost.x >= 0`, `toBeInViewport({ ratio: 1 })`, and a width floor reflecting the ~228px the line needs. Carry an **executed** red proof: with one dropped item restored it must fail. Never `getBoundingClientRect`. — Rationale: jsdom applies no stylesheet, so no unit test can ever see this fit; and `.app-shell` is a grid whose `.status-bar` item GROWS when overfull, so comparing the assurance against the bar compares two sides that move together — which is not an assertion.

**Acceptance Criteria:**
- Given a fresh render in Preview, when 3.2 s pass with no new render, then the document bar's figure advances from `412 ms ago` to `3.2 s ago` on its own, and the evidence rail's `elapsed` does **not** change.
- Given any of the seven `PreviewFreshness` states, when the chrome and the status line are rendered, then both come from a single `freshnessChrome` call and the bar reads `current` exactly when the status line begins with `Current`.
- Given a Preview whose render failed or whose digest mismatched, when the chrome is shown, then the bar does not read `current`.
- Given Design mode, when the status bar is shown, then `LOCAL SHELL` and the font count are present and the assurance line is absent; and given Preview, the reverse.
- Given a render in flight and no other exit, when the author presses DESIGN, then the render is abandoned and its late result is never installed.
- Given `npm run lint`, when it runs, then there are exactly 4 `react(only-export-components)` warnings and no errors.

## Spec Change Log

**Entry 1 — 2026-09-08, step-04 review. The fit assertion named an observable that cannot fail.**

*Triggering finding.* The verification-gap layer found, and a browser measurement confirmed, that
`promise.x + promise.width <= bar.x + bar.width` is structurally incapable of failing. `.app-shell`
(`App.css:19`) is a grid with `min-width: 1024px` and `overflow: hidden`; `.status-bar` is its item, so
an overfull `nowrap` row **grows the bar's own border box** and both sides of that inequality move
together. Measured with both dropped items restored: the bar is 163px overfull, two thirds of the
assurance is clipped off-screen, and **all five assertions passed**, `toBeInViewport()` included
(its default ratio accepts any intersection) and `bar.height === 32` included (nowrap correctly forces
*horizontal* overflow, which a horizontal assertion against the grown box cannot see).

*What was amended.* The Execution task's wording only — the observable is now the viewport, which does
not move with the content. The frozen I/O matrix row is untouched and cannot be amended here; see the
note below.

*Known-bad state avoided.* Shipping a geometry guard that reports green while the product's central
trust claim is clipped off the screen — D-000.32's exact shape (a guard that cannot fail is
indistinguishable from a passing one), inside the one story whose subject is chrome that tells the truth.

*KEEP — these survived the finding and must survive any re-derivation.* The two-item drop and its
arithmetic; `white-space: nowrap` on `.status-bar` (load-bearing: without it an overfull bar absorbs the
excess vertically and is clipped by the fixed 32px track, where no horizontal assertion could see it);
the `bar.height === 32` assertion; and the decision to place the assurance last so it is the element
pushed out first.

*⚠ FROZEN-BLOCK DEFECT — ESCALATED, AND NOW AMENDED.* The `## I/O & Edge-Case Matrix` row
*"Assurance line fits"* stated the same unfailable observable inside `<frozen-after-approval>`, which is
read-only to every agent. Its **Error Handling** column stated the property that actually governs —
*"Restoring either dropped item must make this fail"* — and that is the clause the implementation
satisfies, so the row was internally inconsistent rather than merely ambiguous.

**Amended by the orchestrator, 2026-09-08.** The Expected Behaviour column now measures against
`page.viewportSize()` and names why the status bar's own box is the wrong reference: `.app-shell` is
`display: grid; overflow: hidden` and `.status-bar` is its item, so an overfull `nowrap` bar grows its
border box and both sides of the comparison move together.

**This is the frozen block working, not failing.** The builder found a defect in human-owned intent, could
not fix it, did not quietly reinterpret it, did not revert 1128 lines of correct work over one comparison,
and escalated it with the evidence. A human then renegotiated the wording — which is exactly the
transaction `<frozen-after-approval>` exists to force. The measured proof that the old wording was
unfailable: with both dropped items restored the bar ran 163px overfull with two thirds of the assurance
clipped off-screen, and all five assertions passed, `toBeInViewport()` included at an intersection ratio of
0.337.

**Entry 2 — 2026-09-08. A false premise inside this spec's own Q1 ruling record.**

The Q1 ruling text (quoted verbatim above, and left unaltered because it is a quotation) justifies
dropping `template-font-count` partly on the ground that *"the font count is a template fact whose
render-side siblings moved to 13.3's evidence rail."* **The rail carries no font fact at all** —
`grep -ain font` over `preview/evidence-rail.tsx` and `preview/evidence-rail-facts.ts` returns zero
hits; the rail carries engine, target, pages, elapsed, size, hash and diagnostics. The *ruling stands* —
the count is genuinely a template fact rather than a render fact, which is sufficient on its own — but
the relocation half of its reasoning is false, and the code comment that had repeated it was corrected.
Consequence to disclose: in Preview the font count now has no home anywhere on screen.

## Design Notes

**The seam, decided in advance.** The lead judges this two goals built as one story, and names the
seam now so a mid-flight cut is a known seam rather than an improvised one (D-13.3.1's lesson applied
before the fact instead of after):

- **Goal A — the chrome states the render's freshness.** AC1/2/3/5: the document-bar swap, the ticker,
  the shared function, the heading-button removal, the cancel test. Converges on `App.tsx:2163`/`:2188`
  and `preview/freshness.ts`. UX-DR14.
- **Goal B — the app states its standing promise.** AC4: the assurance line and the two-item drop.
  Different element (`App.tsx:2254`), different CSS rule, different test block, different design
  citation (UX-DR23). Needs **none** of Goal A's machinery.

**If 13.5 must be cut mid-flight, Goal B is the clean cut and Goal A is the remainder**, and the
deferral is registered in a tracker a planner reads — never only in this file.

**The heading-button removal is NOT a safe separate cut**, despite looking like the smallest piece. It
sits on the same physical source line (`App.tsx:2188`) as the `id="preview-freshness-status" …
role="status"` substring that `preview-authority-contract.test.ts:22` pins contiguously. It is the most
entangled edit in the story, not the least.

## Verification

Per owner ruling **D-000.33** (2026-09-08), the heavy suites moved to the Epic 13 boundary gate. This
section carries unit tests, typecheck, lint and e2e compile only. **`npm run test:e2e`, the integration
suites, `npm run build` as a gate, the `verify:offline*` chain and the font-host scans DO NOT RUN in
this story, and must be reported as "did not run", in those words.** The e2e specs written and amended
here are **written-and-compiled-only** and do not execute until the Epic 13 boundary gate.

**Commands:**
- `cd folio-designer && npm run typecheck` — expected: clean, exit 0.
- `cd folio-designer && npm run lint` — expected: exactly **4** `react(only-export-components)` warnings, no errors. Baseline anchors measured at `0d395ca`: `preview/pdf-viewer.tsx:17:14`, `:18:14`, `App.tsx:3895:14`, `App.tsx:3902:17`. **Re-measure the anchors and report the new ones; do not repeat these line numbers.** A fifth warning means a value export was added to a `.tsx` — put shared helpers in a `.ts`.
- `cd folio-designer && npm test -- --run` — baseline measured at `0d395ca`: **68 files / 1076 tests, 0 failures**. Report the new counts, and **diff the test *name* sets, not the totals** — equal totals can hide a swap.
- `cd folio-designer && npm run test:e2e:compile` — expected: clean, exit 0. **This is only `tsc -p tsconfig.e2e.json --noEmit`. It is not a browser run and must never be described as one.**

**Manual checks:**
- The ticking effect is absent in Design mode — verified by reading the effect's guard, and by the fact that `App.test.tsx:4846`/`:7734` still terminate.
- `git status --porcelain` is limited to the files this story's Execution list names.
- The `<frozen-after-approval>` block is byte-identical to its approved state.

## Suggested Review Order

**The one producer — read this first**

- The whole story's coherence rests here: one switch yields both renderings.
  [`freshness.ts:95`](../../folio-designer/src/preview/freshness.ts#L95)

- The state list the exhaustiveness guard iterates, instead of restating it.
  [`freshness.ts:11`](../../folio-designer/src/preview/freshness.ts#L11)

- Where both renderings are consumed — the single call site.
  [`App.tsx:2092`](../../folio-designer/src/App.tsx#L2092)

**The age is the render's, not the screen's**

- Stamped once, at the only install site; never re-read from live state.
  [`App.tsx:694`](../../folio-designer/src/App.tsx#L694)

- The record carries `installedAt`, so a stale record keeps dating itself.
  [`App.tsx:165`](../../folio-designer/src/App.tsx#L165)

- The interval, fenced by one expression so Design can never hold it.
  [`App.tsx:2020`](../../folio-designer/src/App.tsx#L2020)

- Pure formatter with the ruled four-tier ladder and the negative clamp.
  [`freshness.ts:147`](../../folio-designer/src/preview/freshness.ts#L147)

- Token absent means no record: the bar never affirms a render that never happened.
  [`App.tsx:2097`](../../folio-designer/src/App.tsx#L2097)

**The chrome swap**

- Preview reads freshness; Design keeps page setup, and the label branches too.
  [`App.tsx:2256`](../../folio-designer/src/App.tsx#L2256)

- Two items fenced to Design, assurance appended last so it is pushed out first.
  [`App.tsx:2377`](../../folio-designer/src/App.tsx#L2377)

- `nowrap` is load-bearing: it forces horizontal overflow a guard can see.
  [`App.css:433`](../../folio-designer/src/App.css#L433)

**The guard that had to be rewritten**

- Measures the viewport, not the bar's own box, which grows when overfull.
  [`preview-navigation.spec.ts:314`](../../folio-designer/e2e/preview-navigation.spec.ts#L314)

- `ratio: 1` — the default accepted a two-thirds-clipped line at ratio 0.337.
  [`preview-navigation.spec.ts:328`](../../folio-designer/e2e/preview-navigation.spec.ts#L328)

**Tests worth reading**

- The row that was vacuous: `error` WITH a surviving record, driven end to end.
  [`App.test.tsx:7864`](../../folio-designer/src/App.test.tsx#L7864)

- Proves a new render re-stamps, so the figure is about the render on screen.
  [`App.test.tsx:7990`](../../folio-designer/src/App.test.tsx#L7990)

- The exhaustive invariant, iterated off the module's own state list.
  [`freshness.test.ts:89`](../../folio-designer/src/preview/freshness.test.ts#L89)
