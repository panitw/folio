---
title: 'The preview navigates by page thumbnails'
type: 'feature'
created: '2026-09-09'
status: 'done'
baseline_commit: '62fd9379f41522c7096c09393d2ca4ff6726384b'
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Epic 13's goal promises "a page-thumbnail rail that doubles as a diagnostic map". Story 13.3
built the evidence rail and deferred the thumbnail rail (DW-304, D-13.3.1); nothing was ever built. In
Preview today the left grid column still renders the **design-mode component palette** — 180px of
controls that cannot place anything, because the canvas is gone — and the author's only way to reach page
7 of a 34-page render is to type into the status-bar page field. There is no way to see the shape of the
document.

**Approach:** Replace the palette column in Preview with a PAGES rail: one numbered thumbnail per page,
the current page marked, click to navigate, bounded with `… N more`. Thumbnails are rasterised by
**two files vendored from mozilla/pdf.js** (`pdf_thumbnail_view.js`, `renderable_view.js`, tag
`v6.2.108`, build SHA `0365cbde0`), with folio's own rail container above them. The rail is a pure
reader of `previewViewState`; it holds no page state of its own. **The diagnostic map is not built** —
it has no data source and is deferred to DW-311 by owner decision D-13.6.3.

## Boundaries & Constraints

**Always:**
- **`App.tsx:254`'s `previewViewState` remains the single page-state authority**, written only through
  its funnel `changePreviewViewState` (`App.tsx:814`). The rail reads the current page from it and
  requests changes through the same funnel. It stores no page number of its own.
- **The vendored files land as `.js`, in `src/vendor/pdfjs/`, outside `src/preview/`.** Both properties
  are load-bearing and measured, not stylistic — see Code Map §V.
- Vendored source keeps its Apache-2.0 header byte-for-byte, and every deviation from upstream is
  recorded in the provenance manifest with its reason.
- The rail's thumbnails come from `pdfPage.render(...)` on the PDF the engine already produced. No
  second render, no re-serialisation.
- `viewer-navigation.ts` and its twelve tests stay exactly as they are.

**Ask First:**
- Adding **any** npm dependency, devDependencies included — `src/font-store.test.ts:426` and
  `src/font-name-table.test.ts:160` both assert `dependencies` equals `['pdfjs-dist','react','react-dom']`
  by exact equality, and the orchestrator has ruled a new dependency is an escalation. If unit-testing
  the rail turns out to require `canvas` or a polyfill, **stop and ask**.
- Minting a design token. `design-contract.test.ts:22` asserts implemented token names equal
  `DESIGN.md`'s by exact equality per group, so a new token means editing a planning artifact from
  inside an implementation story. Use a literal. If a token is genuinely unavoidable, **stop and ask** —
  the orchestrator will make the `DESIGN.md` edit.
- Vendoring any **third** file from pdf.js. The owner authorised the thumbnail modules and their
  dependencies; the orchestrator narrowed that to two files (D-13.6.4). Taking fewer needed no
  permission; taking more does.
- Any change under `folio-go/`. The epic's standing promise is that it touches no engine byte.

**Never:**
- **Never build diagnostic-to-page marking, in any form.** No partial map, no marking from the two
  table-code sites, no disabled hook, no `page` field on `EngineDiagnostic`, no widening of
  `isDiagnostic`'s `hasExactKeys` set, no parsing of a diagnostic's message text. DW-311 owns this
  entirely. A map that marks 2 of 7 diagnostic sites lets an author read an unmarked page as clean.
- Never vendor `pdf_thumbnail_viewer.js`. Measured: 1,964 lines of page-editing UI (`#deletePages`,
  `#cutPages`, `#pastePages`, `#undo`, `#reportTelemetry`), 36 prohibited-identifier hits, and an
  internal `_currentPageNumber` that would be a second page-state authority.
- Never import `pdfjs-dist/web/pdf_viewer.mjs`, `PDFViewer`, or `pdf_viewer.css`. The CSS references 32
  unique images and `vite.config.ts:16` sets `assetsInlineLimit: 0`; against 3 free asset slots that
  fails the offline release outright.
- Never inline vendored code into `src/preview/pdf-viewer.tsx`. `pdf_thumbnail_view.js` contains
  `createObjectURL`, `revokeObjectURL`, `http://` and `https://` — all four on
  `preview-authority-contract.test.ts:7`'s forbidden list **for that one file**.
- Never add the substring `exclude` to `vite.config.ts` (`canvas-authority-contract.test.ts:795` is a
  bare whole-file substring check), and never add an `@media` rule to `App.css` (`:342-345` asserts the
  list equals exactly `['prefers-reduced-motion: reduce']`).
- Never touch `viewer-navigation.ts`, `sprint-status.yaml`, `deferred-work.md`,
  `fixtures/declared-variants/*`, or any `signoff.json`.
- Never run a git write command. See Design Notes.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Rail enumerates a render | Preview mode, `previewPages = 5` | Five rail entries, numbered `1`–`5`, in page order | N/A |
| Current page marked | `previewViewState.page = 3` | Entry 3 alone carries the current marking and `aria-current="page"`; the other four carry neither | N/A |
| Click navigates | Author activates entry 4 | `changePreviewViewState` called once with `page: 4`, every other field of `previewViewState` unchanged | N/A |
| Bound respected | `previewPages = 34`, bound 12 | Twelve entries, then the literal text `… 22 more`; entries 13–34 are not rendered | N/A |
| Exactly at the bound | `previewPages = 12`, bound 12 | Twelve entries and **no** `… N more` text at all | N/A |
| One page | `previewPages = 1` | One entry numbered `1`, no truncation text | N/A |
| No render yet | Preview mode, `previewPages` undefined | The rail's PAGES heading renders with no entries and no truncation text | Never a thumbnail for a page that does not exist |
| Palette absent in Preview | `mode === 'preview'` | The component palette is not in the document at all | N/A |
| Palette present in Design | `mode === 'design'` | The palette renders unchanged, and the rail is not in the document | N/A |
| Page beyond the bound stays reachable | `previewPages = 34`, author types `30` in the status bar | Navigation succeeds; the rail still shows the first twelve and marks none as current | The rail never becomes the only route to a page |
| Rasterisation fails | `pdfPage.render` rejects for one entry | That entry keeps its placeholder; the rail and the other entries stay usable | A thumbnail failure never blanks the rail or the page area |
| Failed render | `previewError` set, no bytes | No rail at all — there is no document to enumerate | Zero diagnostics and zero bytes; the failure card renders instead |

</frozen-after-approval>

## Code Map

Every anchor re-measured at **`a1dff17`**, the dispatch baseline, tree clean. **`folio-designer/src/App.tsx`
contains two literal NUL bytes near lines 3201–3202, so the whole file is invisible to `grep -I` — exit 1,
no output, no warning. Use `grep -a` for every search in this repository, and state the population searched
in the same sentence as any absence reported.**

### §V — The vendored files, and why their placement and extension are load-bearing

Fetched and hashed from `https://raw.githubusercontent.com/mozilla/pdf.js/v6.2.108/web/`:

| file | lines | bytes | SHA-256 (upstream, before modification) |
|---|---|---|---|
| `pdf_thumbnail_view.js` | 557 | 17,105 | `8bb39945f9199f8c35fc1cb7999dc0542ab5859691365dab15e43be241093526` |
| `renderable_view.js` | 72 | 1,617 | `b1f630e45648c765a02c8733412118f9a245d1e9dd6b9d034d4e46eadb1cfd79` |

- `renderable_view.js` has **zero import statements**. It exports `RenderableView` and `RenderingStates`
  and is fully self-contained.
- `pdf_thumbnail_view.js` has **exactly three** import lines (`:24-26`) and needs exactly three
  modifications — see the Execution list. `AppOptions` is touched at only `:95` and `:96`, both supplying
  a default for a value that is already a constructor option.
- **`.js`, not `.ts`, is measured.** `canvas-authority-contract.test.ts:8-16` builds its three corpora
  with `/\.(?:ts|tsx|css)$/`, `/\.(?:ts|tsx)$/` and `/\.test\.(?:ts|tsx)$/`. A `.js` file matches none,
  so it enters no arm. `tsconfig.app.json` sets no `allowJs`, so `tsc -b` never compiles it. `vitest`'s
  `test.include` is `['src/**/*.test.{ts,tsx}', 'scripts/**/*.test.mjs']`, so it is never collected.
  Measured: `npx oxlint` over both real files at a non-ignored `src/` path emits **0 warnings, 0 errors**.
- **`src/vendor/pdfjs/`, not `src/preview/`, is also measured.**
  `withoutApprovedLocalPointerInput` (`canvas-authority-contract.test.ts:866-911`) waives
  `scroll(?:Width|Height|Left|Top)` for **every** file whose path satisfies
  `file.includes(`${path.sep}preview${path.sep}`)`. A `.d.ts` sidecar placed under `src/preview/` would
  inherit that waiver **by accident** — precisely what AC3 forbids. Outside that directory it inherits
  nothing.
- ⚠ **A `.d.ts` IS a `.ts`, so each sidecar is in the AD-17 production corpus.** Keep every prohibited
  spelling out of them: no `clientWidth`/`clientHeight`/`scrollTop`/`offsetWidth`/`getBoundingClientRect`/
  `getComputedStyle`/`devicePixelRatio`/`ResizeObserver` member on any declared type.
- Measured across all 17 `prohibited` regexes and all 19 `refusalVocabulary` regexes, through the
  contract's own comment-stripping scanner: **`pdf_thumbnail_view.js` 0 hits, `renderable_view.js` 0 hits.**
  So **no exclusion is required at all** — see Design Notes for how AC3's intent is honoured positively.
- `OutputScale` and `RenderingCancelledException` are both **confirmed exported** from
  `pdfjs-dist/build/pdf.mjs` (runtime probe: both `function`), so the `pdfjs-lib` rewrite resolves.
- `Promise.withResolvers` at `pdf_thumbnail_view.js:341` is ES2024 and Safari 16 lacks it. Vite's default
  target is `["chrome107","edge107","firefox104","safari16"]` and esbuild passes the call through
  verbatim. **This introduces no new browser-support class**: `node_modules/pdfjs-dist/build/pdf.mjs`
  already contains 27 occurrences and is already shipped. Do not "fix" it here.

### §A — The collaborator surface `PDFThumbnailView` actually needs

The vendored class touches its injected collaborators at exactly three sites, so folio's adapters are tiny:

- `:374` — `this.renderingQueue.isHighestPriority(this)`
- `:415` — `this.eventBus.dispatch('thumbnailrendered', { source, pageNumber, pdfPage })`
- `:538` — `this.linkService.pagesCount`

Constructor options it reads: `container`, `eventBus`, `id`, `defaultViewport`,
`optionalContentConfigPromise`, `linkService`, `renderingQueue`, `maxCanvasPixels`, `maxCanvasDim`,
`pageColors`, `enableSplitMerge` (leave `false`; it is what creates the selection checkbox).
Public methods folio drives: `setPdfPage`, `draw`, `toggleCurrent`, `reset`, `destroy`, `setPageLabel`.

⚠ **`grep -ac currentPageNumber pdf_thumbnail_view.js` is 0.** The class holds no current-page value;
`toggleCurrent(isCurrent)` is driven entirely by the caller. This is why V-B satisfies the one-authority
AC *by construction* rather than by discipline — the `_currentPageNumber` field lives in
`pdf_thumbnail_viewer.js`, the file this story does not vendor.

DOM it builds: `div.thumbnail[page-number]` → `div.thumbnailImageContainer.missingThumbnailImage[role=button][page-number]`
→ `img`. It sets `data-l10n-id`/`data-l10n-args` attributes for pdf.js's localiser, which folio does not
have; they are inert. Folio supplies its own accessible name.

### §B — Page state, and the funnel the rail must use

- `App.tsx:254` — `const [previewViewState, setPreviewViewState] = useState<PDFPreviewViewState>(initialPDFPreviewViewState)`. **The single authority.**
- `App.tsx:814` — `changePreviewViewState`, the single write funnel, `useCallback(…, [])` and stable.
- `App.tsx:259` — `previewPages`, the page **count**, set only by `viewerPages` (`:810-813`).
- `App.tsx:2186-2239` — the status-bar `previewNavigation`: `goToPreviewPage` (`:2188`) is the existing
  page-change helper and is the function the rail should reuse rather than reimplement.
- `PDFPreviewViewState` is `{ page; scale; scrollTop; scrollLeft; fit? }` (`preview/pdf-viewer.tsx:12`).
  A page change must preserve every other field — `samePDFPreviewViewState` (`:18`) is the identity guard
  the funnel uses.

### §C — Where the rail mounts, and the grid

- `App.tsx:2260` — `<div className="workbench" id="future-features">`.
- `App.tsx:2261` — `<nav className="palette-rail" aria-label="Component palette">` … **an unconditional
  direct child**. The mode ternary does not begin until `App.tsx:2276`. This is the line AC5 is about.
- `App.css:46` — `.workbench { … grid-template-columns: var(--palette-width) minmax(0,1fr) var(--panel-width) }`,
  the **only** `grid-template-columns` on `.workbench` and the declaration Story 13.3 deliberately left
  free for this rail (its reasoning is at `App.css:554-566`).
- `App.css:28` — `.app-shell-preview` exists already but touches `grid-template-rows` only.
- `tokens.css:16` — `--palette-width: 180px; --panel-width: 300px`. **There is no rail-width token and
  none may be minted**; the design's 132px goes in as a literal.

### §D — The design, and the tokens that already carry it

`…/mockups/Preview.dc.html:50-77`. Every colour maps to an existing token — verified against `tokens.css`:

| mockup | token | use |
|---|---|---|
| `132px` | *(literal)* | rail column width |
| `#1a1e23` | `--color-panel` | rail ground |
| `#272c33` | `--color-line` | rail right border |
| `#737c86` | `--color-ink-low` | `PAGES` label, with `--type-label` |
| `44×62px`, `#e9ecef` | `--color-page-shell` | thumbnail placeholder |
| `#2f353d` | `--color-line-strong` | thumbnail border, non-current |
| `#58a6c4`, 2px | `--color-select` | **current** page border |
| `#ffffff` | `--color-page` | current thumbnail ground |
| `#5e666f` / `#e6e9ec` | `--color-ink-faint` / `--color-ink-high` | page number, other / current |
| `#4e565f`, mono 10px | `--color-ink-ghost` + `--type-mono` | `… N more` |
| `8px` gap, `10px 12px` padding | `--space-3`, `--space-4`/`--space-5` | spacing |

⚠ The mockup also draws a **`#c9a758` (`--color-bind`) diagnostic mark on page 4** (`:66-71`).
**Do not build it.** D-13.6.3 defers it to DW-311. The mockup's `… 29 more` beside five thumbnails
implies a bound of 5 at its own fixed height; that is an illustration, not a specified bound (Design Notes).

### §E — Guards that constrain this story

- `canvas-authority-contract.test.ts:762-766` — file-count floors, **all `toBeGreaterThanOrEqual`**, so
  adding files is safe. Measured headroom now: production 70 (floor 58), tests 63 (51), e2e 20 (15),
  `.tsx` 9 (8), **`.css` 3 (floor 3 — AT the floor, so no `.css` file may be removed)**.
- `canvas-authority-contract.test.ts:748-749` — the `tests` and `e2e` arms must contain only `.ts`/`.tsx`;
  `production` deliberately also allows `.css`. A `.js` enters no arm.
- `canvas-authority-contract.test.ts:342-345` — `App.css`'s `@media` list must equal exactly
  `['prefers-reduced-motion: reduce']`. **The rail's width cannot be responsive.**
- `canvas-authority-contract.test.ts:793-795` — `vite.config.ts` must match `/include:\s*\[/` and must
  **not** match `/exclude/` anywhere, comments included.
- `preview-authority-contract.test.ts:7` — 13 forbidden tokens over **`pdf-viewer.tsx` only**, including
  `createObjectURL`, `revokeObjectURL`, `fetch(` (a bare substring, so `prefetch(` fails too), `http://`,
  `https://`. `:22` pins the contiguous run
  `id="preview-freshness-status" className="preview-status" role="status"` in `App.tsx`. `:29`'s red proof
  requires `setPreviewStatus('stale')` to appear **exactly once** in `App.tsx` — do not add a second call.
- **oxlint baseline is exactly 4 `react(only-export-components)` warnings**, re-measured at `a1dff17`:
  `preview/pdf-viewer.tsx:17:14`, `:18:14`, `App.tsx:4030:14`, `:4037:17`. It is a **convention, not a
  mechanical gate** — no `--max-warnings` anywhere, and no test asserts the count. A fifth appears the
  moment a `.tsx` exports a non-component value, so every non-component helper goes in a `.ts`. The
  precedent is already in `src/preview/`: `evidence-rail.tsx` (components + types) beside
  `evidence-rail-facts.ts` (everything else). **Re-measure and report the new anchors; do not repeat these.**
- `src/font-store.test.ts:426` and `src/font-name-table.test.ts:160` — `dependencies` equals
  `['pdfjs-dist','react','react-dom']` by exact equality. Vendoring source adds no package; keep it that way.
- `src/font-store.test.ts:433-451` — the one source walk that *does* include `.js`. It looks only for the
  string `fake-indexeddb`; measured, neither vendored file contains it.
- `scripts/forbidden-font-hosts.mjs` / `scripts/host-font-access.mjs` — both scan `.js`. Measured: zero
  matches in either vendored file. Both population floors are minimums that new files only help.
- `lint/internal/rules/licencegraph.go:89-110` — `ScanPDFJSNotice` checks that four files exist and are
  non-empty under `folio-designer/third-party-notices/pdfjs-dist/`. **Existence only; no content
  assertion.** So nothing reds if the NOTICE is not updated — but its text currently claims folio
  redistributes "only the local PDF.js worker, four Adobe binary CMaps, and four Liberation Sans
  standard-font files", which vendoring falsifies. AD-26 requires the licence material to be true.

### §F — Test-surface facts

- Suite baseline at `a1dff17`: **68 files / 1102 tests / 0 failures**, `npx oxlint` exit 0 with the 4
  warnings above.
- `preview/pdf-viewer.test.tsx:4-10` — the module-mock precedent to follow: `vi.mock('pdfjs-dist/build/pdf.mjs', …)`
  plus `vi.mock('pdfjs-dist/build/pdf.worker.mjs?url', …)`, `getContext` stubbed via
  `vi.spyOn(HTMLCanvasElement.prototype, 'getContext')` (`:36`, `:68`, `:306`), and layout faked by
  redefining `clientWidth`/`clientHeight` on `HTMLDivElement.prototype` (`:31-35`).
- jsdom 28.1.0, measured: `createImageBitmap`, `OffscreenCanvas`, `ImageBitmap`, `structuredClone`,
  `IntersectionObserver`, `ResizeObserver`, `matchMedia` are all `undefined`, and
  `canvas.getContext('2d')` returns `null`. `canvas` is **not** installed and must not be added.
  Therefore `draw()` is stubbed in unit tests and rasterisation is proven only in `e2e`.
- `src/test/setup.ts` is 5 lines (jest-dom + RTL `cleanup`) and polyfills nothing.
- `App.test.tsx:7090-7365` — `describe('Story 13.2: the viewer navigates from the status bar')`, 14 `it`
  blocks. These cover the page-state funnel the rail now also uses; they must stay green untouched.
- `App.test.tsx:200` — asserts the palette is present, inside a **Design-mode** test (`:205` asserts
  `PREVIEW` is `aria-pressed=false`). It still passes after AC5. `e2e/application-shell.spec.ts:13` is
  likewise Design mode.
- `App.test.tsx:7120` — `it('leaves Design mode with the canvas zoom alone and no PDF controls at all')`
  is the existing "absent in the other mode" idiom to copy for the palette assertion.
- ⚠ Any new or amended `e2e` spec **must force the fallback file tier first** — headless Chromium *has*
  the File System Access API, so `showOpenFilePicker()` is called and no `filechooser` event is ever
  emitted (a 90s timeout on a working app):
  `await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })`.
  Model at `e2e/local-file-actions.spec.ts:11`.
- `e2e/preview-navigation.spec.ts:127-132` — the `boundingBox()` geometry idiom, including
  `expect(before.statusBar.height).toBe(32)`.

## Tasks & Acceptance

**Execution:**

- [x] `folio-designer/src/vendor/pdfjs/pdf_thumbnail_view.js`, `renderable_view.js` — vendor both files
  verbatim from tag `v6.2.108`, Apache-2.0 headers byte-for-byte intact. Apply **exactly three**
  modifications to `pdf_thumbnail_view.js`, each marked with a `FOLIO MODIFICATION` comment naming its
  reason: (1) `:24` `from "pdfjs-lib"` → `from 'pdfjs-dist/build/pdf.mjs'`; (2) `:26` delete the
  `AppOptions` import; (3) `:95-96` replace the two `AppOptions.get(...)` fallbacks with declared
  constants carrying pdf.js's own upstream defaults. Change nothing else. — Rationale: `pdfjs-lib` is a
  pdf.js build alias that does not exist here, and `app_options.js` supplies only these two numbers, so
  dropping it removes 1,082 lines from the closure.
- [x] `folio-designer/src/vendor/pdfjs/pdf_thumbnail_view.d.ts`, `renderable_view.d.ts` — sibling
  declarations using **top-level `export declare`**, not the `declare module` shape of
  `src/preview/pdfjs-dist.d.ts` (which only works for bare package specifiers). Declare only the surface
  §A lists. **No prohibited identifier may appear in either file** — they are in the AD-17 production
  corpus. — Rationale: `allowJs` is false and `strict` is on, so an undeclared `.js` import is a hard
  `TS7016`; measured, a sibling `.d.ts` makes `tsc -b` exit 0 and keeps the `.js` off `--listFiles`.
- [x] `folio-designer/src/vendor/pdfjs/PROVENANCE.md` — the manifest: upstream repository, tag
  `v6.2.108`, build SHA `0365cbde0` (present in both `build/pdf.mjs` and `web/pdf_viewer.mjs`), the
  two-file list with the upstream SHA-256 of each as fetched, the three recorded modifications with
  reasons, the licence (Apache-2.0, unchanged), an explicit note that `pdf_thumbnail_viewer.js` was
  **deliberately not** vendored and why, and a step-by-step re-vendoring procedure. — Rationale: AC1;
  shape borrowed from `folio-go/internal/fontset/vendor-boundary.md`.
- [x] `folio-designer/src/vendor/pdfjs/vendor-pin.test.ts` — the executable half of the audit, following
  `2-3a`'s AC8 ("load-bearing rows are executable, not narrated"). Assert: (1) the vendored directory
  contains **exactly** the four expected files plus the manifest — a fifth file fails, which is what
  stops the fork growing unnoticed; (2) each `.js` file's current SHA-256 equals the value recorded in
  `PROVENANCE.md`, **read out of the manifest** rather than restated in the test; (3) the version pin —
  `PROVENANCE.md`'s tag matches `pdfjs-dist`'s installed version from `package.json`, so a
  `pdfjs-dist` bump reds here; (4) each file still contains its Apache-2.0 header; (5) **each `.js` file
  contains none of the AD-17 prohibited identifiers**, importing the prohibition list from the contract
  rather than restating it. — Rationale: AC1 and AC2. Row (5) is how AC3's intent is met positively —
  see Design Notes.
- [x] `folio-designer/src/preview/page-rail-facts.ts` — the rail's non-component logic, as pure
  functions: `PAGE_RAIL_BOUND = 12` (the arithmetic is in Design Notes; change the constant, not the
  call sites, if the value is revised), `railEntries(pages, currentPage, bound)` returning the entry list
  and the truncation remainder (`undefined` when `pages <= bound`), and the `… N more` string builder. Also
  the three collaborator adapters §A requires (`isHighestPriority`, a `dispatch` sink, `pagesCount`),
  which are ~10 lines total. — Rationale: a `.ts` costs no `oxlint` warning; a `.tsx` value export would
  make a fifth.
- [x] `folio-designer/src/preview/page-rail.tsx` — the rail component: a `<nav aria-label="Page thumbnails">`
  with the `PAGES` label, one entry per `railEntries` row, each a real `<button>` carrying its page
  number and `aria-current="page"` when current, calling the funnel on activation. Mount one
  `PDFThumbnailView` per entry and drive `setPdfPage` / `draw` / `toggleCurrent` / `destroy`.
  **Export components and types only.** — Rationale: AC4, AC5 and the click/current-page ACs; keyboard
  reachability comes free from using `<button>` rather than a `div[role=button]`.
- [x] `folio-designer/src/App.tsx:2261` — gate the palette on `mode === 'design'` and render `<PageRail>`
  in Preview in the same grid slot. Pass `previewPages`, `previewViewState.page`, the funnel, and the
  document. **Do not add a second `setPreviewStatus('stale')` call**, and do not disturb the pinned
  attribute run at `:2288`. — Rationale: AC5; measured, the palette is unconditional today.
- [x] `folio-designer/src/App.css:46` — add a Preview-mode `grid-template-columns` on
  `.app-shell-preview .workbench` (or equivalent) using the literal `132px` for the first column, plus
  the rail's own rules using only the tokens in §D. **No `@media` rule. No new token.** — Rationale: the
  shared declaration 13.3 left free; the authority contract forbids a second `@media`.
- [x] `folio-designer/third-party-notices/pdfjs-dist/NOTICE` — add a line stating that folio also
  redistributes two vendored pdf.js source files, naming them and the tag. — Rationale: the existing text
  says "only the … worker, four CMaps, and four Liberation Sans files", which vendoring makes untrue.
  Nothing reds if this is skipped, which is exactly why it must be done deliberately (AD-26).
- [x] `folio-designer/src/preview/page-rail.test.tsx` and `page-rail-facts.test.ts` — cover **every row of
  the I/O matrix**, with `PDFThumbnailView.prototype.draw` stubbed and no rasterisation asserted. The
  boundary rows matter most: `pages === bound` must produce **no** truncation text, and `pages === bound + 1`
  must produce `… 1 more`. Assert the click path calls the funnel **once** with only `page` changed, by
  comparing the whole next state object against the previous one — not by asserting `page` alone. Assert
  the rail reads the current page from the prop and **stores none**: re-rendering with a new current page
  must move the marking with no rail-internal state involved. — Rationale: the funnel assertion is what
  proves the one-authority AC; asserting `page` alone would pass on a rail that silently dropped `scale`.
- [x] `folio-designer/src/App.test.tsx` — add the palette mode-partition test in the `:7120` idiom: in
  Preview the palette is **absent** and the rail is **present**; in Design the reverse. Verify, do not
  assume, that `:200` and the other palette assertions are Design-mode and still pass. — Rationale: AC5.
  The Design-mode half is what stops the removal leaking out of Preview.
- [x] `folio-designer/e2e/preview-page-rail.spec.ts` — a **real** browser spec: the rail renders, a
  thumbnail actually rasterises (assert the `img` acquires a `src` and the container loses
  `missingThumbnailImage`), clicking entry 3 moves the page indicator, and the rail column measures 132px
  via `boundingBox()`. Force the fallback file tier first per §F. **This spec does not execute in this
  story** — it is written and compile-checked only, and runs at the Epic 13 boundary gate. Never describe
  compile-clean as passing. — Rationale: rasterisation is unprovable in jsdom, so this is the only witness
  that the vendored code works at all.

**Acceptance Criteria:**
- Given the vendored tree, when `vendor-pin.test.ts` runs, then a changed byte in either `.js`, a fifth
  file in the directory, a `pdfjs-dist` version that no longer matches the recorded tag, a stripped
  Apache-2.0 header, or a prohibited identifier appearing in a vendored file each make it fail — and each
  of those five failures is **executed and observed**, not asserted to be possible.
- Given `previewViewState` is the sole page-state authority, when the rail is searched for stored page
  state, then it holds none: `grep -a` over the rail's own files returns no `useState`/`useRef` holding a
  page number, and the marking is a pure function of the prop.
- Given Preview mode, when the status bar navigates to a page the rail has truncated away, then navigation
  still succeeds — the rail is never the only route to a page.
- Given `npm run lint`, when it runs, then there are exactly 4 `react(only-export-components)` warnings
  and no errors, with the new anchors reported rather than the ones in §E repeated.
- Given the diagnostic map, when the whole diff is searched, then there is no page field on a diagnostic,
  no widening of `isDiagnostic`, no marking code and no disabled hook for one — DW-311 owns it entirely.

## Spec Change Log

- **Review triage (14 patches), after implementation.** Non-frozen sections only; the
  `<frozen-after-approval>` block is byte-identical to its approved state (sha256
  `14f9a42f5c071dc1b6a16b2f9880a7a2c134ce0930e65bc3ce3c0dcce49ad613`, 6,779 bytes).
  - **Design Notes / Verification — the payload prediction was corrected from "zero new assets,
    `assetCount` 61, three free slots" to the MEASURED "one new asset row, `assetCount` 62, two free
    slots".** The original reasoning assumed a static import; the implementation must use a dynamic one.
  - **Code Map §V's line counts are one high** for both vendored files (557/72 recorded; `wc -l` measures
    556/71 — a `split('\n')` count includes the empty element after the trailing newline). The Code Map
    is left as the record of what the story measured at dispatch;
    `src/vendor/pdfjs/PROVENANCE.md` carries the corrected figures and names the discrepancy.
  - No change to Intent, Boundaries & Constraints, the I/O matrix, or Tasks & Acceptance.

## Design Notes

**Why the vendored files are `.js` in `src/vendor/pdfjs/`, and how AC3 is satisfied without an exclusion.**
AC3 asks for a narrow, justified AD-17 exclusion because vendored pdf.js "measures its own rendered
output". Measured, that premise does not hold for the two files V-B actually takes: across all 17
prohibited regexes and all 19 refusal regexes, `pdf_thumbnail_view.js` and `renderable_view.js` produce
**zero hits**. The 36 hits live in `pdf_thumbnail_viewer.js` — the file the D-13.6.4 narrowing removed
from scope. And as `.js` they are outside all three corpora anyway.

So this story writes **no exclusion**, and that is the stronger outcome: there is no carve-out for a
future file to inherit. AC3's real worry — that the scan stops seeing something — is answered positively
instead, by `vendor-pin.test.ts` row (5), which asserts the vendored files are *clean* rather than
*exempt*, and by row (1), which fails if a third file appears in the directory. A future re-vendor that
drags in a measuring file reds immediately. Placing the tree outside `src/preview/` matters for the same
reason: that directory carries a **directory-wide** `scroll*` waiver
(`canvas-authority-contract.test.ts:866-911`), and a `.d.ts` dropped inside it would inherit an exemption
by accident.

**The truncation bound is a judgement, and here is the arithmetic.** The AC requires that a bound exist,
not a particular value. A thumbnail row is 62px plus an 8px gap = 70px. The rail cannot measure its own
height — browser measurement is banned outside the one named `measuredViewerBox` seam — so the bound must
be a fixed count rather than a fit. The bound is also a rasterisation budget: each entry is one
`pdfPage.render` call, where today's viewer rasterises exactly one page at a time. **12** is proposed:
840px of rail, which fills a typical viewport without truncating most documents, and 12 page rasters at
preview open. The mockup's five-thumbnails-plus-`… 29 more` is an illustration constrained by its own
fixed 764px canvas, not a specified bound. If the orchestrator prefers a different number it is a
one-constant change in `page-rail-facts.ts`.

**Why the click assertion compares whole state objects.** `PDFPreviewViewState` carries `page`, `scale`,
`scrollTop`, `scrollLeft` and `fit`. A rail that navigated by calling the funnel with a freshly built
object would silently reset zoom and scroll, and a test asserting `next.page === 4` would pass over that
bug. Comparing the whole object against the previous one is what makes the assertion able to see the
defect.

**The payload ACs are discharged at the boundary gate, not here, and that is a conflict worth naming.**
Two ACs require the offline release to still build and pass with the measured `assetCount` reported. But
`npm run build` and the `verify:offline*` chain run at the Epic 13 boundary gate under D-000.33, so this
story **cannot** measure a built `assetCount`, and Verification says so in those words rather than
implying a measurement it did not take.

⚠ **THE FIRST VERSION OF THIS PARAGRAPH PREDICTED ZERO NEW ASSETS AND WAS WRONG.** It reasoned that two
vendored `.js` modules imported from TypeScript are "bundled by Vite into an existing chunk". They are
not: `page-rail.tsx` reaches the vendored module through a **dynamic** `import()`, so Rollup emits it as
its own chunk and the offline manifest carries it as its own asset row. The dynamic import cannot be
removed — a static one puts `pdfjs-dist/build/pdf.mjs` into `App.tsx`'s module graph, and pdf.js touches
`DOMMatrix` while evaluating, which jsdom does not define (measured: three unit files fail to load).

So the figure below is **measured, not predicted** — `npx vite build` then `npm run build:offline`, run
during review triage and therefore **outside this story's own gate cadence**, which is why Verification
still reports the offline chain as "did not run":

- `dist/assets/pdf_thumbnail_view-<hash>.js` — **7.56 kB raw / 2.62 kB gzip, its own chunk**
- `s1.assetCount` = **62** against `maximumCacheAssets = 64` (`src/release-payload.ts:42`) — **two free
  slots, not three**
- `brotli.immutableAssetCount` = 61, `brotli.totalBytes` = 19,048,890

This story adds no image, font or CSS file — the rail's rules go into the existing `App.css`, and
`pdf_viewer.css` and its 32 images are explicitly out of scope — so the one new row is the whole delta.
If the gate measures anything other than **62**, that is a finding, not a rounding difference.

Note also that the final AC's premise is stale after the D-13.6.4 narrowing: it describes "the viewer
bundle is added" and prices `pdf_viewer.mjs` 307 KB + `pdf_viewer.css` 160 KB + 328 KB of images at a ~90%
PDF.js payload increase. **None of that is added.** The honest measurement this story owes is the one
above — 629 vendored lines that share the already-shipped `build/pdf.mjs` — and the ~90% figure describes a
route that was ruled out.

**The seam, named in advance,** so a mid-flight cut is a known seam rather than an improvised one:
- **Goal A — the palette gives way to a rail.** AC5 plus the grid change: `App.tsx:2261`, `App.css:46`.
  Needs none of the vendored code and could ship with a placeholder rail.
- **Goal B — the thumbnails rasterise.** The vendored files, the provenance audit, the adapters.
If 13.6 must be cut, **Goal A is the clean cut and Goal B is the remainder**, and the deferral goes in a
tracker a planner reads — never only in this file.

**Verbatim, into every step-03 subagent handoff, unaltered:** *You never commit. You never run `git add`,
`git stash`, `git checkout`, `git reset`, `git revert`, `git restore`, or any other git write command. You
never push and you never create a branch. The orchestrator makes every commit in this run.* Reading git
state (`git log`, `git status`, `git show`, `git diff`) is permitted and encouraged. The one ratified
carve-out: `git init`/`add`/`commit` inside a `mkdtempSync` directory under `os.tmpdir()`, removed in a
`finally`, is permitted for a test fixture.

## Verification

Per owner ruling **D-000.33**, the heavy suites run at the Epic 13 boundary gate, which the orchestrator
runs. This section carries unit tests, typecheck, lint and e2e compile only. **`npm run test:e2e`, the Go
suites, the matrix legs, `npm run build` as a gate, the `verify:offline*` chain and the font-host scans DO
NOT RUN in this story, and must be reported as "did not run", in those words.** The e2e spec written here
is **written-and-compiled-only** and does not execute until that gate.

**Commands:**
- `cd folio-designer && npm run typecheck` — expected: clean, exit 0. `tsc -b` is **incremental and can
  exit 0 without checking anything**; use `npx tsc -b --force` for proof that the `.d.ts` sidecars
  actually satisfy the imports.
- `cd folio-designer && npm run lint` — expected: exactly **4** `react(only-export-components)` warnings,
  no errors. Baseline anchors at `a1dff17`: `preview/pdf-viewer.tsx:17:14`, `:18:14`, `App.tsx:4030:14`,
  `App.tsx:4037:17`. **Re-measure and report the new anchors; do not repeat these.** A fifth warning means
  a value export reached a `.tsx` — move it to a `.ts`.
- `cd folio-designer && npm test -- --run` — baseline at `a1dff17`: **68 files / 1102 tests / 0 failures**.
  Report the new counts, and **diff the test *name* sets, not the totals** — equal totals can hide a swap.
- `cd folio-designer && npm run test:e2e:compile` — expected: clean, exit 0. **This is only
  `tsc -p tsconfig.e2e.json --noEmit`. It is not a browser run and must never be described as one.**

**Explicitly deferred to the Epic 13 boundary gate, and to be reported as "did not run":**
- `npm run build`, the `verify:offline*` chain and the resulting `s1.assetCount` — so the two payload ACs
  are **not discharged by this story**. The figure to confirm there is `s1.assetCount` = **62** with
  **one** new asset row (`pdf_thumbnail_view-<hash>.js`, 7.56 kB) and two free slots against
  `maximumCacheAssets = 64`. That figure was measured during review triage with `npx vite build` +
  `npm run build:offline`, which is NOT this story's gate cadence: the `verify:offline*` chain itself and
  `npm run build` as a gate still **did not run**, and must be reported that way rather than as "the
  release still builds".
- `npm run test:e2e` — `e2e/preview-page-rail.spec.ts` is written and compile-checked only. It has never
  been executed by anything. Do not describe compile-clean as passing.
- The Go suites, the matrix legs and the font-host scans. Standing Go reds that are **not** regressions
  and must not be reported as such: `TestCorpusMeetsP6ExerciseFloors` and its subtest `P6g_(opaque_names)`.

**Manual checks:**
- The five `vendor-pin.test.ts` failure modes were each **executed** and observed red, and the numbers
  reported — not asserted to be possible.
- `git status --porcelain` is limited to the files this story's Execution list names.
- The `<frozen-after-approval>` block is byte-identical to its approved state.
- `grep -a` was used for every search touching `App.tsx`, and any absence reported names its population.
- `vite.config.ts` still contains no `exclude` substring, and `App.css` still has exactly one `@media`.

## Suggested Review Order

**The vendored boundary — read this first, it is the story's only third-party code**

- The whole design in one line: two files, from one tag, with three recorded edits.
  [`PROVENANCE.md`](../../folio-designer/src/vendor/pdfjs/PROVENANCE.md)

- Modification 1 of 3; the other two are the `AppOptions` drop and its two constants.
  [`pdf_thumbnail_view.js:24`](../../folio-designer/src/vendor/pdfjs/pdf_thumbnail_view.js#L24)

- AC3's answer: the vendored files are proved *clean*, never *exempt*.
  [`vendor-pin.test.ts:111`](../../folio-designer/src/vendor/pdfjs/vendor-pin.test.ts#L111)

- A directory negation only; every file inside still faces `.env`, `*.key`, `*.pem`.
  [`.gitignore:74`](../../.gitignore#L74)

**Page state — the AC that says there may be exactly one authority**

- The rail is mounted as a pure reader; `previewViewState` stays the only owner.
  [`App.tsx:2276`](../../folio-designer/src/App.tsx#L2276)

- Deps `[bytes, count]` are what re-arm the pass when the page count arrives late.
  [`page-rail.tsx:39`](../../folio-designer/src/preview/page-rail.tsx#L39)

- Marking is driven from the prop; the rail stores no page number.
  [`page-rail.tsx:133`](../../folio-designer/src/preview/page-rail.tsx#L133)

**Behaviour the review changed**

- `reset()` is what makes the frozen matrix row true; upstream clears the placeholder.
  [`page-rail.tsx:103`](../../folio-designer/src/preview/page-rail.tsx#L103)

- Clears `ariaCurrent`; `false` reflects as the string `"false"`, so only `null` removes it.
  [`page-rail-facts.ts:109`](../../folio-designer/src/preview/page-rail-facts.ts#L109)

- Explicit tracks, so the rail-less failed render cannot slide left into 132px.
  [`App.css:75`](../../folio-designer/src/App.css#L75)

- All nine options, including the three that make CJK rasterise offline.
  [`page-rail.tsx:62`](../../folio-designer/src/preview/page-rail.tsx#L62)

**Supporting**

- The bound is one constant and a rasterisation budget, not a fit calculation.
  [`page-rail-facts.ts:18`](../../folio-designer/src/preview/page-rail-facts.ts#L18)

- The rail's markup: a real `<button>`, with pdf.js chrome sealed in an `aria-hidden` well.
  [`page-rail.tsx:140`](../../folio-designer/src/preview/page-rail.tsx#L140)

- The licence material now names what is actually redistributed.
  [`NOTICE:13`](../../folio-designer/third-party-notices/pdfjs-dist/NOTICE#L13)
