---
title: 'The viewer navigates like a PDF viewer'
type: 'feature'
created: '2026-09-07'
status: 'in-progress'
baseline_commit: '15b80c66809e4b7f56c239bcd882a6af3b94ea1f'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Reading a 34-page statement in Preview is Previous/Next one page at a time and `−`/`+` in
ten-point steps — fifteen clicks end to end, with no fit, no typed zoom and no typed page. The viewer's
own scroll container is never height-constrained, so the outer region scrolls in its place and the page
scrolls away from the chrome that describes it; the page/zoom/scroll the view state already carries is
thrown away on every render, so leaving Preview and coming back loses the author's place. And the render
effect lists the whole view-state object in its dependency array (**DW-191**), so **every horizontal
scroll of a zoomed page today destroys the `PDFDocumentProxy` and re-rasterizes the PDF** — live now, on
the horizontal axis, not latent behind the height fix.

**Approach:** Give the viewer the navigation a PDF viewer has, in the place the design puts it. Narrow
the render effect to what rendering actually reads, height-constrain the viewer's own scroll container
so it scrolls on both axes with the page safely centred, stop discarding the view state on re-render,
and move the page stepper, the page indicator and the zoom into the application's bottom status bar —
adding a typed page, a typed zoom, and fit-width/fit-page that persist across page changes until the
author zooms by hand. Fit needs the container's pixel size; that measurement is taken inside
`src/preview/` under the canvas-authority exception that directory already holds, extended by the
property names used and by nothing else.

## Boundaries & Constraints

**Always:**
- **DW-191 is fixed before or with the height cap, never after it** (D-000.8). Capping the height first
  turns an already-firing tear-down into one that fires on every vertical scroll too.
- Every read of a container's pixel size lives in a file under `folio-designer/src/preview/`, and
  `canvas-authority-contract.test.ts`'s preview exception is extended **by explicit property name**,
  never by widening its group to a wildcard. `clientLeft`, `clientTop`, every `offset*`,
  `getBoundingClientRect`, `ResizeObserver`, `getComputedStyle` and `devicePixelRatio` must all still
  red inside `src/preview/`, **proved by running the scan against each, not by reading it** (D-11.3.7).
- The preview exception gains a **seam assertion** in the shape of its three siblings, so it dies with its
  reason; and the narrowing is **mutation-proved** — `clientLeft`, `clientTop` and `offsetWidth` still red
  *after* the change, exactly as measured before it. A widening that slipped past would be invisible.
- **New status-bar content is additive.** No existing status-bar item is removed, moved or renamed:
  `data-testid="offline-status"`, `data-testid="engine-snapshot"` and `data-testid="template-font-count"`
  are all located by e2e specs, and removal appears in no 13.2 criterion.
- **`App.css` gains no `@media` rule.** `canvas-authority-contract.test.ts` asserts the extracted media
  list *equals* `['prefers-reduced-motion: reduce']`; a responsive breakpoint reds it.
- `App.css` carries **no raw colour literal and no un-tokenised `border-radius`** — the scans at
  `design-contract.test.ts:87–88` read the whole file, comments included.
- Any new stale transition keeps the existing ternary spelling `setPreviewStatus(previewRef.current ?
  'stale' : …)`. **A second literal `setPreviewStatus('stale')` in `App.tsx` reds
  `preview-authority-contract.test.ts:29`**, whose `String.replace` substitutes only the first
  occurrence — a failure whose name explains nothing.
- A new `App.test.tsx` block goes **above** the `// STORY 17.1: THE CANVAS FOLLOWS THE CONTENT FIELD.`
  comment, and must contain neither the literal `describe('Story 17.1` nor any `// <ALLCAPS> tests.`
  sentence. Appending after the file's last describe, or writing either of those two strings, breaks the
  self-count guard at `App.test.tsx:7576` with a message that points at the wrong block.
- The viewer's `<section className="pdf-preview" aria-label={label}>` keeps its region role and its
  three label spellings, and the scroll host keeps `role="img"` with the same name — five e2e specs and
  `App.test.tsx` locate the preview by exactly those.
- `pdf-viewer.tsx` keeps exactly **two** value exports. A third adds a fifth
  `react(only-export-components)` warning and moves the lint gate off 4; shared helpers go in a new
  module under `src/preview/`, which is inside the same exception directory.
- Story 13.4's claims are not regressed: `PreviewRecord.standIn`, the amber `role="note"` notice, the
  `NO-DATA LAYOUT PREVIEW` heading, and the stand-in-qualified export wording.
- Every new control is keyboard-reachable, operable, labelled and shows visible focus (UX-DR25).
- Mutation-prove every new or changed guard **by deletion**, restore with `cp`, and report the restore
  digest.

**Ask First:**
- Admitting any property name into the canvas-authority exception beyond the ones this story's fit
  arithmetic actually consumes, or admitting any *other* prohibition into it.
- Any recompute-on-resize behaviour, or a `window` resize listener — **ruled out for this story**
  (Q5(a), below); the gap is registered instead.
- Any change to the engine, the wasm boundary, the engine protocol, or `Render` — **this story touches
  no engine byte**.
- Any change to `.folio` open/save behaviour, or to Save PDF's wording or placement (13.1's surface;
  **DW-281** is open against it — an **owner request**, verbatim *"Later this button should be moved to
  the preview area."* Do not act on it here; 13.1 owns that surface).
- Changing `previewOversample` or anything else about raster fidelity.
- Removing any existing status-bar item, or any content the design assigns to Story 13.3 or 13.5.

**Never:**
- **NEVER COMMIT.** No commit, `git add`, `git add -A`, stash, checkout, reset, revert, restore, branch,
  push, or pull request. Leave the work in the tree and name the files. The single carve-out is
  `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`, removed in a `finally`.
- Never touch `fixtures/declared-variants/expected.pdf`, `input.folio` or `signoff.json` — human-attested
  and bound to the live PDF hash.
- Never run `npm run build`.
- No new dependency, and no new file-access, network, storage or object-URL path in `preview/` — the 13
  forbidden tokens stand.
- Not the PAGES thumbnail rail, not the evidence rail, not removing the palette rail from preview mode
  (all Story 13.3). Not the freshness chrome, the mode-switch rework or the "no network · nothing left
  this machine" assurance (Story 13.5).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Fit-width | container's usable width 800px, page 600 CSS px wide at scale 1 | scale resolves to 800/600; the zoom readout shows that resolved percentage, not the word "fit" alone | N/A |
| Fit-page | usable box 800×500, page 600×900 at scale 1 | scale resolves to the smaller of the two ratios, so the whole page fits | N/A |
| Container unmeasurable | fit chosen while the measurement reads 0 (jsdom, hidden, unmounted) | scale is left exactly as it was and the fit choice is retained | no 0, negative or NaN scale ever reaches the rasterizer |
| Fit persists across a page change | fit-page active, author presses Next | the new page is scaled by the same fit rule, re-resolved for that page's own size | N/A |
| Manual zoom clears fit | fit active, author presses `+` or types a zoom | fit is cleared and the scale steps from the currently resolved scale | N/A |
| Typed page | `"7"` on a 34-page document | the viewer navigates to page 7 | N/A |
| Typed page out of range or non-numeric | `"0"`, `"35"`, `"abc"`, `""` | no navigation; the field returns to the current page | nothing is written to the view state |
| Typed zoom | `"250"` / `"10"` / `"abc"` | clamped to the zoom bounds / rejected | the field returns to the current zoom |
| Scroll of a zoomed page | author scrolls horizontally at scale 2 | `scrollLeft` is recorded **and the document is not re-loaded or re-rasterized** — the PDF.js document-open count does not move | N/A |
| Leave Preview and return | page 7, zoom 150%, scrolled, then DESIGN then PREVIEW | page, zoom, scroll position and fit choice are the ones the author left | N/A |
| Re-render onto a shorter document | current page 30, new document has 2 pages | the page clamps to the new last page | the existing `safePage` clamp, unchanged |
| Preview cleared | `invalidatePreview(clear)` | the view state resets to its initial value | N/A |
| Container resized under an active fit | the window or a panel changes width while fit-width is set | the previously resolved scale stands until the next re-render (page, bytes, zoom or fit change), which re-resolves it | documented, not a failure — ruled Q5(a) and registered |
| Fit chosen from the zoom control | the author picks `Fit width` where a percentage would be | the fit is stored on the view state and the readout shows the *resolved* percentage | N/A |

</frozen-after-approval>

## Code Map

Anchors re-derived by symbol at `ea7f31d` (D-000.4), each one read directly rather than relayed; every
search used `grep -a` (DW-256). **HEAD moved twice under this plan gate**, `ea7f31d` → `820db0a` ("Close
13.4, and strip a fabricated owner request out of the register") → `41ba28d` ("Mark 13.2 in-progress"),
both the orchestrator's and both touching only `_bmad-output/`. Every code anchor below was measured at
`ea7f31d` and **stands unchanged at `41ba28d`**, verified by diffing the two commits. `820db0a` renumbered
DW-276 to **DW-281**. ⚠ **`a5b7a10` ("Restore the owner attribution I wrongly stripped") REVERSED that
ruling**: the owner did say it, verbatim, in another session; D-13.4.3 is withdrawn, DW-281 carries the
attribution again, and DW-284 is folded into it. Cite DW-281 as an **owner request**, not a finding.

**The viewer — `folio-designer/src/preview/pdf-viewer.tsx` (108 lines).**
- `PDFPreviewViewState` (`:11`) — `{ page; scale; scrollTop; scrollLeft }`; `initialPDFPreviewViewState`
  (`:16`) and `samePDFPreviewViewState` (`:17`) are the file's **only two value exports**, and are
  exactly the two `oxlint` warnings the gate expects.
- `PDFPreviewViewerProps` (`:19`) — `bytes, label, describedBy, state, onStateChange, onError,
  onPageCount`.
- **`:89` — the DW-191 dependency array**, `[bytes, label, page, scale, state, onError, onPageCount,
  onStateChange]`. `state` is a fresh object literal on every `onStateChange({ ...state, … })`, so a
  scroll re-runs the effect, `dispose()`s the document and re-rasterizes. The three callbacks are
  **already stable** — `viewerError` (`App.tsx:721`), `viewerPages` (`App.tsx:734`) and
  `changePreviewViewState` (`App.tsx:738`) are all `useCallback(…, [])` — so only `state` has to leave.
  <!-- AMENDED BY THE ORCHESTRATOR AFTER APPROVAL, 2026-09-08 (D-13.2.1). THE SENTENCE ABOVE IS FALSE and
  is left in place only so the amendment is legible against it. Only `changePreviewViewState` is passed
  directly. `viewerError` and `viewerPages` are wrapped in INLINE ARROWS at the JSX site (`App.tsx:2148`)
  to bind `preview.token`, so both are FRESH IDENTITIES ON EVERY App RENDER and both sit in the dep array.
  Removing `state` alone therefore did NOT fix DW-191: a scroll still re-rendered App, changed those two
  deps, disposed the PDFDocumentProxy and zeroed the scroll — proved by instrumenting the running page
  (`set-300 -> 300`, `after-800ms -> 0`, `scroll-events = 2, top = 0`, with the CSS verified correct and
  the whole ancestor chain non-scrolling). No unit test saw it because `viewerProps()` builds
  `onError: vi.fn(), onPageCount: vi.fn()` ONCE and reuses that object across rerenders, so the suite
  tests a viewer given props the application never supplies. RULED: fix in `pdf-viewer.tsx` by holding the
  callbacks in refs and dropping them from the dep array — rendering CALLS them, it is not DRIVEN by them
  — which makes the property hold regardless of caller, rather than by a `useCallback` discipline the next
  JSX edit can undo. The DW-191 test gains an arm that rerenders with FRESH callback identities, since
  that is what App actually does. -->
- ⚠ **`:56` is the trap in that fix.** `if (safePage !== page) onStateChange({ ...state, page: safePage })`
  spreads `state` from the effect's closure. Once `state` leaves the dep array that closure is stale, and
  the write silently reverts whatever scroll or zoom happened in between. The **separating input** is: scroll,
  then let a clamp fire — a fix that reads a live view state keeps the scroll, a fix that spreads the stale
  closure discards it.
- `:91–95` — the scroll-restore effect, deps `[state.scrollLeft, state.scrollTop]`, writing
  `host.current.scrollTop/scrollLeft`. It is the code the epic calls dead; it is dead only vertically.
- `:97–107` — `<section className="pdf-preview">`, then `.pdf-preview-controls` (`:98`) carrying all six
  controls to be moved: `Previous PDF page` (`:99`), the `PDF page status` output (`:100`), `Next PDF
  page` (`:101`), `Zoom out PDF` (`:102`, `Math.max(0.5, scale - 0.1)`), the `PDF zoom` output (`:103`),
  `Zoom in PDF` (`:104`, `Math.min(2, scale + 0.1)`), then the scroll host `.pdf-preview-scroll` (`:106`,
  `role="img"`, `tabIndex={0}`, `onScroll`).
- `previewOversample = 2` (`:15`) and the comment at `:59–70` — the standing precedent for *why* the
  browser is not asked to measure. A fit measurement is a different claim and needs the exception, not
  this constant.

**App — `folio-designer/src/App.tsx` (4011 lines).**
- `:35` the viewer import; `:218` `const [previewViewState, setPreviewViewState] = useState(...)`.
- **`:628` `setPreviewViewState(initialPDFPreviewViewState)`** inside `runPreview`, immediately after
  `installPreview(...)` — the reset that defeats persistence. Leaving Preview sets status `'stale'`
  (`returnToDesign`, `:720`), so `enterPreview` (`:714`) → `renderPreview` → `runPreview`'s
  nothing-changed early return (gated on `previewStatus === 'current'`) **cannot fire**, and execution
  always reaches `:628`.
- `:556` the same reset inside `invalidatePreview(clear)` — that one is correct and stays.
- `:734` `viewerPages(token, pages)` — **it does not store the count**, only promotes the status to
  `'current'`. A status-bar page indicator needs the count in App state, which does not exist today.
- `:738` `changePreviewViewState` — de-dupes through `samePDFPreviewViewState`.
- `:2072` — the whole preview `<main>` on one line: heading, the pinned freshness `<p>`, 13.4's
  `preview-standin-notice`, `PreviewFailure`, `PDFPreviewViewer` (a **bare child of `<main>`**, no
  wrapper), `PreviewDiagnostics`, the evidence `<p>`, and 13.1's file-message pair.
- **`:2104` `<footer className="status-bar" aria-label="Status bar">`** — the target. It is a sibling of
  `.workbench`, outside both `<main>`s, and is the **same element in both modes**; it carries `LOCAL
  SHELL`, the engine snapshot `<code>`, `.status-spacer`, the template font count, the offline status,
  and `<code>{mode.toUpperCase()} MODE</code>`. Nothing in it is preview-conditional today.
- `:2067` `.canvas-tools` — the design-mode zoom trio, accessible names `Zoom out` / `Canvas zoom` /
  `Zoom in`, distinguished from the viewer's only by the missing `PDF` suffix. Do not collide.
- `:2051` the palette rail, rendered unconditionally in both modes — **Story 13.3's**, not this one's.
- `:633` is the **only** literal `setPreviewStatus('stale')` in the file (verified twice: `grep -an`, and
  `str.count` in `python3` over the raw bytes). The other eleven calls use the ternary form.

**Layout — `folio-designer/src/App.css` (623 lines).** The chain, read end to end:
- `:19` `.app-shell { height: 100%; overflow: hidden; display: grid; grid-template-rows:
  var(--doc-bar-height) minmax(0, 1fr) var(--status-bar-height); }` — the only height-constrained node.
- `:37` `.workbench { min-height: 0; overflow: hidden; display: grid; grid-template-columns: … }`.
- `:451` `.preview-region { min-width: 0; overflow: auto; padding: var(--space-6); … }` — **this is what
  actually scrolls today.**
- `section.pdf-preview` has **no rule at all** (searched all 623 lines; the only `.pdf-preview` selector
  is `:36`, a `:focus-visible` outline shared with `.mode-switch button`).
- `:457` `.pdf-preview-scroll { min-height: 240px; overflow: auto; display: grid; justify-content:
  center; padding: var(--space-6); … }` — **no `height`, no `max-height`**, so its `overflow: auto` never
  activates vertically. `:458` `.pdf-preview-scroll canvas { max-width: none; … }` is why it already
  overflows *horizontally*.
- ⚠ Two consequences of `:457` that the story inherits. Its **`padding: var(--space-6)` is inside
  `clientWidth`**, so the measured element must carry no padding or border — move the visual padding to
  an enclosing element rather than subtracting a hardcoded mirror of the token. And **`justify-content:
  center` is the unsafe keyword**: with horizontal overflow — which happens today — the left edge of the
  page is centred out of reach. `safe center` on both axes is the fix, and it is a precondition of the
  centring the AC asks for.
- `:424` `.status-bar` declares **no height of its own**; its height is the `.app-shell` grid track,
  `--status-bar-height: 24px` (`tokens.css:16`). That token is **also** the size of `.icon-button`
  (`:27`), `.file-button` (`:29`) and `.mode-switch button` (`:34`), so it cannot be raised in place.
  `DESIGN.md:182–188` declares `status-bar.heightPreview: '32px'`; no such token exists in `tokens.css`,
  and `design-contract.test.ts:22`'s token-name equality harvests only 2-space-indented keys, so a
  `heightPreview` sub-key is invisible to it and a new size token trips nothing.
- `:452`/`:455`/`:456` the `.pdf-preview-controls` rules — they retire with the controls.

**The guard to extend — `folio-designer/src/canvas-authority-contract.test.ts` (827 lines).**
- `:21` the one prohibition line: `/\b(?:offset(?:Width|Height|Left|Top|Parent)|client(?:Width|Height|Left|Top)|scroll(?:Width|Height|Left|Top))\b/`.
- `:768–774` `withoutApprovedLocalPointerInput` — the preview arm, scoped by `file.includes(`${path.sep}preview${path.sep}`)`,
  rewriting `/scroll(?:Width|Height|Left|Top)\b/g` to `viewerTransientState` **before** comment-stripping
  (composition order at `:281`). It is the **one exception in this file that asserts no seam** — its three
  siblings (`:750` `embedded-face-registry.ts`, `:761` `canvas-font-stack.test.ts`, the Story 17.6 e2e
  carve-out) each `expect(source).toMatch(seam)` so the carve-out dies with its reason. D-000.6 names that
  as the shape to copy.
- **Plan-gate measurement, taken (throwaway `node -e`, nothing written to the tree).** Inside
  `src/preview/` today: `clientWidth` **RED**, `clientHeight` **RED**, `clientTop` **RED**, `clientLeft`
  **RED**, `offsetWidth` **RED**, `scrollLeft` ok. With `\bclientWidth\b` and `\bclientHeight\b` added as
  two separately-named rewrites: `clientWidth` ok, `clientHeight` ok, and `clientTop`, `clientLeft`,
  `offsetWidth` **still RED**. `getBoundingClientRect` is caught by a different pattern and is untouched
  either way. If the build's own run disagrees with this record, the disagreement is itself the halt.
- `:342–345` the `@media` equality assertion. `:668–669` the corpus floors (`production ≥ 58`,
  `tests ≥ 51`) — `toBeGreaterThanOrEqual`, so new files are safe. `:673–675` the independent e2e
  directory walk: **every new `e2e/*.spec.ts` is auto-enrolled** and the `src/preview/` exception does not
  reach it. `:582–584` is the standing proof that an e2e carve-out is per *file*, never per folder.

**Tests and their traps.**
- `folio-designer/src/preview/pdf-viewer.test.tsx` — mocks `pdfjs-dist/build/pdf.mjs`; `pdf(numPages)`
  (`:16`) whose `getViewport({scale})` returns `{width: 20*scale, height: 30*scale}`, so a fit
  computation is fully expressible against it; `viewerProps` (`:21`); the page-navigation test at `:64`
  is the only test in the repository that touches any viewer control.
- `folio-designer/src/App.test.tsx` — `vi.mock('./preview/pdf-viewer', …)` at `:129`. **The stub takes no
  `state`/`onStateChange` props at all** and renders two buttons; `initialPDFPreviewViewState` is spelled
  with computed keys `['scroll' + 'Top']` to dodge a source scan, and `samePDFPreviewViewState: () =>
  false` deliberately defeats the de-dupe. A test that drives page/zoom/fit through App must extend this
  mock. `showRenderedPreview` (`:6358`) is the canonical "render then admit" helper.
- `folio-designer/src/preview/preview-authority-contract.test.ts:29` — the single-occurrence
  `String.replace` red proof described in Boundaries.
- `folio-designer/e2e/preview-no-data.spec.ts` — the model for a browser witness: opens
  `folio-go/testdata/example/first-pdf.folio` through the real file chooser, clicks `PREVIEW`, and waits
  on `getByRole('region', { name: /Current no-data layout PDF, revision \d+/ })`.
  `e2e/pdf-export.spec.ts`, `e2e/preview-parameters.spec.ts`, `e2e/browser-native-roundtrip.spec.ts` and
  `e2e/application-shell.spec.ts` locate the same region by name.

**Absence, with its population.** Searched all 61 `*.test.ts(x)` files under `folio-designer/src` and all
18 specs under `folio-designer/e2e` with `grep -arn` for `Previous PDF page|Next PDF page|PDF page
status|Zoom in PDF|Zoom out PDF|PDF zoom|PDF viewer controls|previewViewState|PDFPreviewViewState`: the
only hits are `pdf-viewer.test.tsx:70–72` and the mock at `App.test.tsx:130–132`. **Nothing asserts the
view-state reset at `App.tsx:628`, nothing asserts the zoom controls, and no e2e spec touches the viewer
controls or the status bar.** So no existing test needs deliberate editing — and nothing would have
caught a regression either.

## Tasks & Acceptance

**Execution:**
- [ ] `folio-designer/src/canvas-authority-contract.test.ts` -- extend the `src/preview/` arm of
  `withoutApprovedLocalPointerInput` (`:768`) with the fit measurement's property names as **separate,
  individually-spelled rewrites**, never by widening the `client(?:…)` group; add a seam assertion in the
  shape of the `embedded-face-registry.ts` sibling so the carve-out dies with its reason; and add
  falsifier rows proving `clientLeft`, `clientTop`, `offsetWidth` and `getBoundingClientRect` still red
  inside `src/preview/`, and that the admitted names still red in a non-preview file. The two admitted
  names are `\bclientWidth\b` and `\bclientHeight\b`, **each its own `.replace(...)`**; the seam
  assertion follows the `embedded-face-registry.ts` sibling at `:750`. **This task lands
  first** — a scan that has not been widened deliberately is the thing that must fail the moment the
  measurement is written.
- [ ] `folio-designer/src/preview/pdf-viewer.tsx` -- **fix DW-191**: narrow the render effect's
  dependencies (`:89`) to what rendering actually reads, so a scroll or a fit write no longer disposes
  and re-rasterizes the document; and rewrite `:56`'s clamp write so it reads the live view state rather
  than the effect's now-stale closure. Remove `.pdf-preview-controls` (`:98–105`) — the controls move to
  the status bar — while keeping the `<section className="pdf-preview">` region and the scroll host's
  `role`, name, `tabIndex` and `onScroll` exactly as they are. Take the container measurement here and
  resolve the fit scale, keeping the file at two value exports.
- [ ] `folio-designer/src/preview/<new module>.ts` -- the pure fit arithmetic: given a page's intrinsic
  CSS size at scale 1 and the container's usable box, return the fit-width and fit-page scales, the
  clamp for a typed zoom, and the clamp/parse for a typed page. Pure and total: an unmeasurable (zero or
  non-finite) box yields "no change", never a zero or NaN scale. Keeping it out of `pdf-viewer.tsx`
  keeps the lint count at four and gives the arithmetic a test that needs no DOM.
- [ ] `folio-designer/src/App.tsx` -- carry the fit choice on the view state alongside page/zoom/scroll;
  hold the viewer's page count in state so the indicator can read it; stop discarding the view state at
  `:628` while leaving `invalidatePreview`'s reset at `:556` alone; and render the page stepper, a typed
  page, the page indicator, `−`/`+`, a typed zoom and the fit choices **inside the bottom status bar
  (`:2104`), in preview mode only**, each with an accessible name that does not collide with
  `.canvas-tools`' `Zoom in`/`Zoom out`/`Canvas zoom`. A manual zoom clears the fit choice. The zoom is a
  `<select>` offering **Fit width / Fit page / 50% / 75% / 100% / 150% / 200%** beside a typed percentage
  input, in the mockup's left-to-right order `◀ · page · ▶ · divider · zoom` — **the orchestrator's design
  call, not a mockup-derived requirement**; see Design Notes. The addition is **purely additive**: no
  existing status-bar item is removed, moved or renamed. Add no second literal
  `setPreviewStatus('stale')`.
- [ ] `folio-designer/src/tokens.css` + `folio-designer/src/App.css` -- mint
  **`--status-bar-height-preview: 32px`**, the value `DESIGN.md:182–188` already declares, and switch
  `.app-shell`'s third grid row by a mode class **derived from the same `mode` value the rest of the shell
  already switches on** — never a second parallel answer to "are we in preview", which is the mirror-drift
  shape this run keeps hitting. `--status-bar-height: 24px` **cannot** be raised in
  place: it is also the size of `.icon-button` (`:27`), `.file-button` (`:29`) and `.mode-switch button`
  (`:34`), so a new token is the only correct move, not a convenience. Then make `.pdf-preview-scroll` the
  bounded scroller instead of `.preview-region`: constrain the height through the preview `<main>` so the
  page area scrolls under fixed chrome, switch both centring keywords to `safe center` so the left edge of
  an overflowing page stays reachable, and move the padding off the measured element. Style the new
  status-bar controls. **No `@media` rule, no raw colour, no un-tokenised radius.**
- [ ] `folio-designer/src/preview/pdf-viewer.test.tsx` -- cover the DW-191 fix by counting PDF.js document
  opens across a scroll write (the count must not move), the fit resolution against the mock's scaling
  viewport, the unmeasurable-container case, and the clamp write's separating input: a clamp that fires
  after a scroll must not revert the scroll.
- [ ] `folio-designer/src/preview/<new module>.test.ts` -- the I/O matrix's arithmetic rows, one test per
  row, against literal expectations rather than values recomputed by the code under test (D-11.2.2).
- [ ] `folio-designer/src/App.test.tsx` -- extend the viewer mock so a test can drive page, zoom, fit and
  scroll; cover the status-bar controls' presence, names, keyboard operation, the typed-entry rejection
  rows, fit-cleared-by-manual-zoom, and view-state survival across DESIGN → PREVIEW. **Insert the block
  above the `// STORY 17.1: THE CANVAS FOLLOWS THE CONTENT FIELD.` comment**, with no `describe('Story
  17.1` literal and no `// <ALLCAPS> tests.` sentence anywhere in it.
- [ ] `folio-designer/e2e/<new spec>.spec.ts` -- the browser witness for the one claim jsdom cannot see:
  that the **viewer's own container** scrolls and the outer region does not. Written as real coverage —
  CI executes the whole Playwright suite on every push since `adf905a`. Assert it **without spelling any
  prohibited property**: every `e2e/` file is auto-enrolled by the directory walk at
  `canvas-authority-contract.test.ts:673`. Zoom past fit, scroll the page area, and show that the preview
  heading and the status bar hold their position while the page canvas moves — Playwright's own
  `boundingBox()` is not on the prohibited list, and the heading holding still *is* the difference
  between the inner container scrolling and the outer region scrolling. **The same witness proves a second
  property**: that the chrome which now surrounds the scrolling page area stays **visible and reachable**
  while the page is scrolled — at minimum 13.4's amber no-data notice or `PreviewDiagnostics`, whichever
  the fixture can reach.

**Acceptance Criteria:**
- Given a preview at a zoom where the page overflows its container, when the author scrolls it, then the
  number of PDF.js document opens across the scroll **does not move** — mutation-prove this by restoring
  `state` to the effect's dependency array and confirming the count assertion reds. A fix whose
  correctness no test can distinguish from the bug is not a fix.
- Given the render effect no longer depends on the whole view state, when a page clamp fires after the
  author has scrolled, then the scroll position survives — the named separating input for `:56`'s stale
  closure, asserted directly rather than inferred.
- Given every property name admitted into the canvas-authority exception, when the scan is **run** against
  each admitted name, each neighbouring name that must stay banned (`clientLeft`, `clientTop`,
  `offsetWidth`, `getBoundingClientRect`), and the same admitted name in a file outside `src/preview/`,
  then each lands on the expected side. Read nothing; run everything (D-11.3.7).
- Given the exception's new seam assertion, when the code that earns the exception is renamed or deleted,
  then the exception's own assertion fails — the carve-out cannot outlive its reason.
- Given an author who sets a page, a zoom or a fit and a scroll position, when they switch to DESIGN and
  back to PREVIEW, then all four are the ones they left. Mutation-prove it by restoring
  `setPreviewViewState(initialPDFPreviewViewState)` at `App.tsx:628` and confirming the assertion reds.
- Given each new control, when it is reached and operated by keyboard alone, then it is reachable in
  document order, carries an accessible name distinct from the design-mode canvas controls, and acts —
  asserted by role and accessible name, never by class or test id.
- Given the whole designer suite, when `App.test.tsx`'s self-count guard and
  `preview-authority-contract.test.ts` run, then both are green **for their own reasons** — the story's
  new tests are inside the Story 17.1 count window's exclusion or the count is amended deliberately, and
  `App.tsx` still contains exactly one literal `setPreviewStatus('stale')`.
- Given the preview chrome that now surrounds a bounded, scrolling page area — 13.4's amber no-data
  notice, `PreviewFailure`, `PreviewDiagnostics`, the evidence line and 13.1's file-message pair — when the
  page area is scrolled to its far extent, then that chrome is still **visible and reachable**, proved in a
  real browser by the same "what stays put" observable as the scroll witness. **A unit assertion that each
  still renders is a different property and must not stand in for this one: jsdom performs no layout, so it
  cannot tell a fixed element from one that has scrolled out of view.** A diagnostic that scrolls away with
  the page is one the author cannot see while looking at the thing it describes.
- Given every guard this story adds or changes, when the line it guards is deleted, then that guard fails;
  the source is restored by `cp` from a snapshot and the restore digest is reported.

## Spec Change Log

- **Plan gate, 2026-09-07 (not a review loopback). AC3's "the new property" is singular and the arity is
  wrong.** The criterion reads *"the exception names the new property explicitly rather than being widened
  by wildcard"*, but AC1 requires **fit-page** as well as fit-width, and the available height is not
  derivable from the width and the page's aspect ratio. Two names are forced: `clientWidth` **and**
  `clientHeight`, each its own separately-spelled rewrite. The orchestrator ruled Q1(c) and directed this
  be said plainly rather than letting the AC's wording quietly stretch to cover a second name. Known-bad
  state avoided: an implementer reading "the property" as a licence to reach for the
  `client(?:Width|Height|Left|Top)` group, which would waive `clientLeft` and `clientTop` at the same
  time. **KEEP on any re-derivation:** the two names spelled individually and never as a group; the
  before/after measurement that `clientLeft`, `clientTop` and `offsetWidth` still red; and the seam
  assertion, which this exception has never had while its three siblings all do.

- **Plan gate, 2026-09-07. The control shape has no mockup, and is recorded as the orchestrator's design
  call.** `Preview.dc.html` is a static picture: **no `<button>`, `<input>` or `<select>` anywhere in its
  309 lines**, and the strings "fit width"/"fit page" appear **nowhere in the `ux-designs` tree**
  (population: `grep -arn "Fit width|Fit page|fit-width|fit width"` over `ux-folio-2026-08-23/**`;
  positive control — the only hits are the bundled design-canvas editor's own "Fit artboards" toolbar,
  which is not a Folio surface). The mockup fixes only the left-to-right order and the 32px height. The
  select-plus-typed-input idiom was **chosen by the orchestrator at this gate**, is flagged to the owner
  as a UI shape chosen without a mockup, and is open to objection before it ships. Known-bad state
  avoided: a choice the orchestrator made reading, later, as one the design specified — **D-13.4.3's rule
  applies to design provenance as much as to owner requests**. **KEEP on any re-derivation:** the
  attribution line, and the population-plus-control sentence that establishes the mockup really is silent.

- **Plan gate, 2026-09-07. Three behaviours this story is the first to guard, and two live defects kept as
  tasks rather than deferrals.** (i) The view-state reset at `App.tsx:628` is tested by nothing —
  population: all 61 `*.test.ts(x)` under `folio-designer/src` and all 18 specs under `folio-designer/e2e`,
  searched with `grep -arn` for seven identifiers, hits only at `pdf-viewer.test.tsx:70–72` and a mock at
  `App.test.tsx:130–132` that takes **no `state`/`onStateChange` props at all**. Ruled Q6(a): delete the
  reset. **A behaviour with no guard is not a behaviour anyone chose, and the next person should know 13.2
  is where the choice was made.** (ii) `.pdf-preview-scroll` (`App.css:457`) uses the **unsafe**
  `justify-content: center`, so with the horizontal overflow `:458` guarantees, **the left edge of a
  zoomed page is unreachable today** — a user-facing bug in the exact surface this story rebuilds, so
  deferring it would ship a navigation story that cannot navigate to the left edge. (iii) The same rule's
  `padding: var(--space-6)` sits inside any `clientWidth` reading, and `getComputedStyle` is banned, so it
  cannot be subtracted at runtime and the arithmetic must account for it by construction. Both are tasks.

- **Step-04, 2026-09-08 (not a review loopback). `npm run build` was run, as an
  orchestrator-authorized exception to the frozen Never list.** The Boundaries forbid `npm run build`
  outright, and `npm run test:e2e` invokes it through `webServer.command`. The orchestrator ran the
  Playwright suite by hand, found this story's witness and Story 13.4's both failing, and directed a
  fix plus a full re-run reporting 40/40 — which is unobtainable without the build. **The exception was
  granted, not taken**; the record shows a granted exception rather than a violated constraint. Two
  things came out of the run that nothing else could have. (i) Both specs waited on a `filechooser`
  without first forcing the fallback file tier; headless Chromium HAS the File System Access API, so
  the app took the native tier, called `showOpenFilePicker()`, emitted no `filechooser` event and the
  wait timed out at 90 s. Population: ten `e2e/` specs wait on a `filechooser`, the eight that pass all
  force the tier (`local-file-actions.spec.ts:11`), and the two that did not were the two that failed.
  (ii) With that fixed, this story's witness failed on a **real product defect** — see the orchestrator's
  amendment inside the Code Map (D-13.2.1): DW-191 was not actually fixed, because `App.tsx:2148` hands
  the viewer two fresh callback identities on every render and both sat in the render effect's
  dependency array. **KEEP on any re-derivation:** that `npm run test:e2e` needs explicit authorization
  because it builds; that a `filechooser` spec must force the fallback tier first; and that
  *"CI executes it"* was false for this repository — the jobs added at `adf905a` had never run, because
  this pipeline never pushes (`git rev-list --count origin/main..HEAD` = 40 at the time). **A test that
  has never been executed is not coverage, whatever the config says.**

## Design Notes

**Plan-gate rulings, carried verbatim (orchestrator, 2026-09-07).**

- Q1 — the exception's arity: *"(c). Two separately-named rewrites plus the seam assertion. (b) is not
  implementable — available height is not derivable from width and aspect ratio — so (a) is forced by the
  AC's own fit-page requirement, and the AC's singular 'the new property' is simply wrong about the arity.
  … The seam assertion is the part I most want: this is the **one** carve-out in
  `canvas-authority-contract.test.ts` that asserts no seam while its three siblings each
  `expect(source).toMatch(seam)`, so today the exception cannot die with its reason. … **Mutation-prove
  the narrowing**: confirm `clientLeft`, `clientTop` and `offsetWidth` still red after your change,
  exactly as you measured before it — a widening that slips past would be invisible."*
- Q2 — the control shape: *"(a). The select-plus-typed-input idiom. It satisfies 'type or choose' with one
  control per concept, preserves the mockup's left-to-right order, and is the interaction every author
  already knows. … That is not a licence to invent freely; it is a gap I am filling with a conventional
  idiom, and I am flagging it to the owner as a UI shape chosen without a mockup so they can object before
  it ships. Record it as **the orchestrator's design call, not a mockup-derived requirement** — the
  D-13.4.3 rule applies to design provenance too: do not let a choice I made read as one the design
  specified."*
- Q3 — the status bar's existing content: *"(a). Additive only. Every item you would remove is owned by
  another story or located by an e2e test (`offline-status`, `engine-snapshot`, `template-font-count`),
  and removing them appears in no 13.2 criterion. The mockup is a target state assembled across
  13.2/13.3/13.5."*
- Q4 — the preview status-bar height: *"(a). Mint `--status-bar-height-preview: 32px`. `DESIGN.md:182-188`
  already declares the number, so this is implementing a declared value rather than inventing one. … a new
  token is the only correct move, not a convenience."*
- Q5 — resize: *"(a). Recompute only when the viewer already re-renders. It is exactly what the AC asks
  and nothing more, in the epic's largest story. **Register the gap.**"*
- Q6 — view-state persistence: *"(a), delete the reset. It is what a PDF viewer does, and (b) invents a
  'why did this render happen' signal to distinguish two paths the AC does not distinguish. The `safePage`
  clamp already handles a document that got shorter."*
- CHECKPOINT 1 / check 1 — all six controls move: *"AC6's consequent is 'so the page area carries the page
  **and nothing else**'. 'Nothing else' cannot coexist with previous/next and `−`/`+` staying in the page
  area. AC1's 'alongside the existing previous/next and `−`/`+`' then reads as those controls being
  alongside the new ones in their new home, not as them staying put."*
- CHECKPOINT 1 / check 2 — the mode class: *"acceptable … **One condition:** the class must be derived from
  the same `mode` value the rest of the shell already switches on, not a second parallel source of truth
  about which mode we are in. Two places that independently decide 'are we in preview' is the mirror-drift
  shape this run has hit repeatedly."*
- CHECKPOINT 1 / check 3 — the restructure and its condition: *"the `preview-region` restructure is
  REQUIRED by AC4, not merely enabled by it … Everything that becomes fixed chrome around the scrolling
  page area … must be asserted to remain **visible and reachable while the page area is scrolled**, not
  merely to still render. … And a unit assertion that each still renders is **not** the same property and
  must not be allowed to stand in for it: jsdom performs no layout, so it cannot tell a fixed element from
  one that has scrolled out of view."*
- Q7 — scope: *"[K] Keep. … (A) and (B) are coupled through the view-state shape, D-000.8 forbids
  separating the DW-191 fix from the height cap, and (B) alone is what the title promises. One user goal."*

**For Story 13.5's builder, so the reasoning is inherited rather than rediscovered.** The Preview status
bar in `Preview.dc.html` is a **target state assembled across three stories**: 13.2 adds the page stepper,
the page indicator and the zoom and removes nothing; 13.3 owns the PAGES rail and the evidence rail; 13.5
owns the freshness chrome, the mode-switch rework and the *"no network · nothing left this machine"*
assurance, and is the story that may decide the Design-mode items (`LOCAL SHELL`, the engine snapshot, the
font count, the offline status, `PREVIEW MODE`) should not be in the Preview bar at all. Nothing in 13.2
prejudges that; the additive rule here is a boundary, not a preference.

**Why the controls move at all.** AC6 is explicit — the design puts the page stepper, the page indicator
and the zoom in the application's **bottom status bar**, "so the page area carries the page and nothing
else". The bar is `App.tsx:2104`, outside both `<main>`s, so App must own the controls and the viewer
must give them up. The view state already lives in App (`:218`), which is why this is a move rather than
a lift; the one thing App does not have is the page count, because `viewerPages` (`:734`) uses it only to
promote the status.

**What the mockup does and does not settle.** `Preview.dc.html:294–307` is a static picture: every
element is a styled `<div>`/`<span>`, there is no `<button>`, `<input>` or `<select>` in the whole
309-line file, and the strings "fit width" and "fit page" appear **nowhere in the entire `ux-designs`
tree** (population: `grep -arn` over `ux-folio-2026-08-23/**`; the only "Fit" hits are in the bundled
design-canvas editor's own toolbar). It fixes the *order* — `◀ · Page 3 of 34 · ▶ · divider · 100%` — and
the 32px preview height, and it is silent on control types. **Where the mockup is silent the AC governs;
where the AC is also silent — which is the case for the control types — the shape is the orchestrator's
call at this gate, recorded as such in the Spec Change Log and open to the owner's objection.**

**The fit measurement, and why it is a real exception rather than a constant.** `previewOversample = 2`
is a constant precisely because `devicePixelRatio` is banned — but that is *image fidelity*, and the
comment at `pdf-viewer.tsx:59–70` says so. A fit scale is not fidelity: it is the container's actual
pixel size, and no constant can stand in for it. AC3 anticipates exactly this and rules the shape: taken
inside `src/preview/`, under the exception that directory already holds, naming the property explicitly —
though it names **one** property and fit-page forces **two**, which the Spec Change Log records.
The measured element must have **no padding and no border**, because `clientWidth` includes padding and
`getComputedStyle` is banned — so there is no way to subtract the token's value at runtime, and
hardcoding a mirror of `--space-6` would be a constant that drifts silently.

**The scrollbar feedback loop, named so it is not discovered at review.** `clientWidth` excludes a
vertical scrollbar. Fit-width scaled to the width measured *before* a vertical scrollbar appears
over-shoots once it does. `scrollbar-gutter: stable` on the measured element removes the loop at the CSS
layer and is not a prohibited property; the alternative — a fudge allowance — is a magic number nothing
can test.

**Two latent `App.css` defects this story necessarily touches.** `.pdf-preview-scroll` (`:457`) uses the
**unsafe** `justify-content: center`, so with the horizontal overflow that `:458` guarantees, the left
edge of a zoomed page is already unreachable today. And its `padding: var(--space-6)` sits inside any
`clientWidth` reading. Both are pre-existing and both are on this story's critical path; `safe center` on
both axes and relocating the padding are the fixes, not scope creep.

**How the browser witness proves the thing jsdom cannot.** jsdom performs no layout, so nothing in this
repository can distinguish a scrollable container from a non-scrollable one — that is DW-191's own
sentence and it is why the CSS half of this story has no unit-level proof. The distinguishing observable
is *what stays put*: if `.preview-region` scrolls, the preview heading and the freshness line scroll away
with the page; if `.pdf-preview-scroll` scrolls, they hold. Playwright's `boundingBox()` reads that
without spelling a single prohibited identifier, so the spec needs no e2e carve-out. Avoid `page.evaluate`
bodies naming `scrollTop`/`clientWidth` — the scan reads raw source text and every `e2e/` file is enrolled
by an independent directory walk.

**The lint gate will move, and that is expected.** `oxlint`'s four warnings are all
`react(only-export-components)`, at `pdf-viewer.tsx:16,17` and `App.tsx:3745,3752`. Editing either file
above those lines moves the anchors; a **third value export** from `pdf-viewer.tsx` would make it five.
The count stays four; the line numbers do not, and the new anchors are reported rather than assumed.

## Verification

**Commands** (run from the repository root unless stated; baselines measured by the orchestrator at
`ea7f31d`):

- `cd /Users/panitw/Projects/folio/folio-designer && npm test` -- expected: **65 files / 1002 tests** at
  baseline, plus this story's new tests; **zero failures**. Report files/tests as a diff of test *names*
  against the baseline, not as a total (a miscount hides behind a matching total).
- `cd /Users/panitw/Projects/folio/folio-designer && npx tsc -b --force` -- expected: exit 0, no output.
- `cd /Users/panitw/Projects/folio/folio-designer && npx oxlint` -- expected: **exactly 4** warnings, all
  `react(only-export-components)`, two in `src/preview/pdf-viewer.tsx` and two in `src/App.tsx`. **Report
  the new line:column anchors**; they move and that is not a regression. Five warnings means a third value
  export crept into `pdf-viewer.tsx`.
- `cd /Users/panitw/Projects/folio/folio-designer && npm run test:e2e:compile` -- expected: exit 0.
- `git -C /Users/panitw/Projects/folio status --porcelain` -- expected: only this story's files, and **no
  commit**. `git add -A` is forbidden; the tree is handed over, not committed.
- **Do NOT run `npm run build`** and do not run `npm run test:e2e` locally — it builds. The Playwright
  witness is compile-checked here and **executed by CI on every push** (`adf905a`), so it is written as
  real coverage; report it as compile-checked locally and CI-executed, never as locally executed.
- Go and `lint` suites are **not** expected to run: this story changes no Go file. If any Go file is
  touched, the gates are `go test -count=1 ./...` → 2273/2/5 failing **only**
  `TestCorpusMeetsP6ExerciseFloors` and `P6g_(opaque_names)` **by enumerated name** — a third failure by
  name is a hard stop — and `cd lint && go test -count=1 ./...` → four `ok`.
- `gofmt -l` over absolute paths -- expected: empty output, **checked by the shape of the output rather
  than by the exit code**.

**Manual checks:**
- Every guard added or changed is mutation-proved **by deletion**, restored with `cp`, and its restore
  digest reported. A guard that survives deletion of the line it guards is removed and the measurement
  recorded (D-000.9).
- The canvas-authority exception is proved by **running** the scan against each admitted name, each
  neighbour that must stay banned, and one non-preview file — not by reading the regex.
- **The narrowing is mutation-proved against the plan-gate baseline.** Before the change, inside
  `src/preview/`: `clientWidth` RED, `clientHeight` RED, `clientTop` RED, `clientLeft` RED, `offsetWidth`
  RED, `scrollLeft` ok. After it: `clientWidth` ok, `clientHeight` ok, and `clientTop`, `clientLeft`,
  `offsetWidth` **still RED**. Report both columns. If the build's own run disagrees with this record, the
  disagreement is itself the halt.
- The new `--status-bar-height-preview` token is confirmed to be **used, not merely defined** — an unused
  token is indistinguishable from its own absence (the lesson `design-contract.test.ts:60` records).
