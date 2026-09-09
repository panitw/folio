# Epic 13 Context: A template author can read, navigate and keep the exact PDF

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Preview is the screen where Folio's central claim lands: the document on screen is the production
document, produced in this tab, and the evidence for that is visible rather than asserted. When the epic
opened, only the middle column existed — the page itself. This epic builds the rest of the screen: a
page-thumbnail rail that doubles as a diagnostic map, an evidence rail carrying render facts and the
output hash as a first-class block, real PDF-viewer navigation (fit, typed zoom, typed page, persistent
scroll), an export path so the exact bytes can leave the tab, chrome that states freshness honestly, and
a preview that runs with no sample data at all. Stories 13.1–13.5 are delivered; the remaining work is
Story 13.6, the thumbnail rail, which was split off from 13.3 at its plan gate (DW-304, D-13.3.1) and
carries its own owner decision about how it is built. The epic's standing promise is that it touches no
engine byte: every PDF it displays, exports and describes is one the engine already produced. Story 13.4
carries a knowingly-taken owner exception to that promise, recorded below.

## Stories

- Story 13.1: The preview keeps the PDF
- Story 13.2: The viewer navigates like a PDF viewer
- Story 13.3: The preview screen is the evidence screen
- Story 13.4: Preview runs without sample data, and an absent value is empty
- Story 13.5: The chrome tells the truth about the preview
- Story 13.6: The preview navigates by page thumbnails

## Requirements & Constraints

- **Save the exact bytes.** The previewed PDF must be writable to a local file byte-for-byte as the
  engine produced it — never re-rendered for the save, never re-serialized. The saved file's digest is
  the digest the screen shows. Where no preview has rendered, the control is absent or disabled with
  its reason stated, never one that fails when pressed.
- **Real viewer navigation.** Fit-width, fit-page, a zoom the author can type or choose, a typed page
  number, and a scroll position that survives leaving and re-entering Preview.
- **Exactness is the product.** The interface must *earn* the byte-identity claim rather than assert
  it. An affirmation the product cannot substantiate is the one thing this screen must never print.
  Nothing in the interface may imply server rendering, a cloud round-trip, or an account — the standing
  "no network · nothing left this machine" assurance belongs where it is always visible.
- **Staleness is the state failure that matters.** A stale preview must be visibly invalidated or
  re-rendered, never presented unmarked — and that rule governs export and the evidence statistics
  exactly as it governs the page. Chrome, status line and rail must never disagree about freshness.
- **Canvas-approximate vs preview-exact must be legible without a tutorial.**
- **Preview must not be gated on sample data.** The engine imposes no such requirement (the CLI renders
  a bindings-free template with no data argument), so the screen must not disable itself ahead of it.
- **Diagnostics belong where the consequence is visible** — surfaced in Preview, non-blocking,
  dismissible, naming the offending element and path, and locating back to it on the canvas.
- **The offline promise is a measured budget, not an assumption.** Any payload the release grows by must
  be measured and recorded before it is accepted, because "no network · nothing left this machine" means
  the whole viewer ships to the author's machine.
- **Accessibility floor** applies to every new control: keyboard-reachable, operable, labelled, with
  visible focus, and diagnostics distinguished by shape before colour.

## Technical Decisions

- **Preview identity.** A rendered preview is keyed by a hash over serialized template ∥ data ∥ params
  ∥ engine version ∥ font-set identity; the key is recomputed on every committed command and any
  difference marks the preview stale. The preview surface is a controlled pdf.js canvas — never the
  browser's built-in viewer in an iframe or embed — because diagnostics overlay it and long renders
  need progress.
- **The browser never measures text.** Every metric and line break comes from the engine's measure API.
  `src/preview/` holds a narrow, explicitly-enumerated exception to the canvas-authority contract
  (today for `scroll*`); the fit-width/fit-page container measurement must be added to that exception
  **by name**, never by widening it with a wildcard — a rasterized PDF's display scale is viewer
  navigation, not document measurement. Guards are widened deliberately, never deleted.
- **File access is two-tier and capability-detected.** Where `showSaveFilePicker` exists, save through a
  held handle; otherwise fall back to a download. Save PDF must reuse the single file-access interface
  the designer already has for `.folio`, parameterising picker type and suggested filename — not a
  second, parallel download path.
- **Cross-target identity is a build property, not a live comparison.** The tab cannot compare itself to
  a native render. What is true is that the wasm engine is the same engine compiled to another target
  and that byte identity across `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm` is proven by
  the CI matrix for the release the browser is running. Wording on screen must say that, not more.
- **`Render`'s semantics are untouched.** No-data empty-value substitution is scoped strictly to
  *preview with no data supplied*. The absent-binding error contract, its diagnostic codes and the
  golden corpus stay exactly as they are; the same template rendered by the Go library with absent data
  still fails. A path absent from data that *was* supplied remains a located error — absent data and
  absent-from-present-data are different conditions.
- **Story 13.4's owner exception (D-13.4.1, 2026-09-07).** The epic's "no engine byte" promise has one
  knowingly-taken exception, and only this one: 13.4 adds a **new read-only Go function** — a
  type-directed stand-in document generator — plus a wasm op to expose it. It was taken because the
  story's empty-value criterion is not otherwise deliverable: no single value renders empty
  (`upper`/`lower` reject null; `formatNumber` rejects both null and `""`; `formatDate` accepts neither
  and has no empty form; and a null under `visibleIf` silently deletes the element). `Render` itself is
  still untouched — it is invoked with genuinely supplied data — so the promise's *intent*, that every
  displayed PDF is one the engine really produced, holds; its literal wording does not. Do not treat
  this as licence to widen engine changes elsewhere in the epic.
- **Story 13.6's owner decision — SUPERSEDED TWICE; read D-13.6.2 and D-13.6.4, not this.** D-13.6.1
  (2026-09-08) said the rail is built on `pdfjs-dist`'s viewer components using `PDFViewer` and
  `PDFThumbnailViewer`. **`PDFThumbnailViewer` is not in the published package** — 0 occurrences against a
  positive control of 4 for `PDFViewer`, and the bundle exports 22 names without it. The orchestrator
  supplied that false premise. As shipped: **two files, 629 lines, VENDORED** from mozilla/pdf.js
  `v6.2.108` into `src/vendor/pdfjs/` (D-13.6.2, owner; narrowed by D-13.6.4). `PDFViewer` was never
  adopted and `pdf_thumbnail_viewer.js` was never vendored — at that version it is Firefox's
  page-organiser, not a rail. The
  orchestrator recommended the core-API route and was overruled; the consequences below are constraints,
  not preferences, precisely because they are the costs that choice carries.
- **One page-state authority.** `PDFViewer` owns page state itself, so the preview must have exactly one
  page-state authority: Story 13.2's `viewer-navigation.ts` is retired or subordinated to it, never run
  in parallel. Two authorities for the same fact is the defect shape this run has found most often, and
  adopting the viewer bundle is what makes it a live risk rather than a hypothetical one.
- **The diagnostic-to-page derivation was NOT built, and no story owns it. See DW-311.** It is true that
  `pdfjs-dist` cannot supply it. What this line got wrong is that 13.6 would therefore write it: measured
  at 13.6's plan gate, **there is no data to write it from.** `EngineDiagnostic` carries no page,
  `isDiagnostic`'s `hasExactKeys` REJECTS one, and five of seven `Diagnostic{` sites in `folio-go/render.go`
  run *before pagination*, so no page exists yet to record. There is no join anywhere in the wasm reply
  between a page index and an `elementId`. **Owner ruling D-13.6.3 deferred it out of the story entirely**,
  rejecting a partial map that would mark two of seven sites and let an unmarked page read as clean.
- **The viewer bundle's payload is measured, not assumed.** Measured before dispatch at `pdf_viewer.mjs`
  307 KB + `pdf_viewer.css` 160 KB + 328 KB of images, against the 853 KB core `pdf.mjs` already shipped
  — roughly a 90% increase in the PDF.js payload before minification and gzip. The actual release-size
  change must be measured and recorded, never asserted.
- **A no-data preview's hash is not evidence** of cross-target equality, because the inputs were not the
  production inputs, and the screen must say plainly that it is a no-data preview whose empty values
  are stand-ins.
- **Layout frame.** Preview mode swaps the palette rail for a ~132 px page-thumbnail rail and the
  properties panel for a ~320 px render panel, and the status bar grows to ~32 px for page navigation —
  the one place the frame changes height. The component palette must not render in preview mode at all,
  since nothing there can be placed.

## UX & Interaction Patterns

- **Preview is a mode switch, not a panel.** The canvas is replaced by the rendered PDF. The mode switch
  in the document bar is the single way in and out; the separate in-heading return button goes away —
  while keeping a reachable way to abandon a render in progress.
- **Three columns.** Left: PAGES thumbnail rail, one numbered thumbnail per page, current page marked in
  the select accent, a page carrying a diagnostic marked in the bind accent, clicking navigates, and a
  long document truncates rather than rendering every thumbnail. **(The mockup's `… 29 more` is an
  illustration constrained by its own fixed 764px canvas, not a specified bound. As shipped the bound is
  **12**, chosen as both a layout and a rasterisation budget, and it is one constant in
  `page-rail-facts.ts`.)** Middle: the page on the
  dark ground, carrying the PRODUCTION OUTPUT badge and nothing competing with it. Right: the evidence
  rail.
- **Evidence rail** carries RENDER (engine version, target, pages, rows, elapsed, byte size), OUTPUT
  HASH as its own bordered block with the digest in mono wrapped across two lines so it can be compared
  by eye, DIAGNOSTICS with counts in the header (total, and errors separately), and a paired Re-render /
  Save PDF action row at its foot. Re-render moves here from the INPUTS tab.
- **Diagnostics** use the shape-before-colour legend — triangle/dashed for a render that proceeded,
  square/solid for one that failed — and a zero state that states zero explicitly so a clean render
  reads as *checked*, not as *nothing here*. A card names its location as bound path ·
  element kind · band - THREE parts, not four - and carries Locate on canvas and Dismiss. **Corrected
  2026-09-09:** this line said "page · bound path · element kind · band". The page part came from the UX
  mockup and has never been true of the code: `evidence-rail-facts.ts:113-117` renders three parts, and
  `diagnostic-presenter.tsx:11-21` states in a comment that the page is deliberately absent. The owner ruled
  it out for Story 13.3 and again for 13.6 (D-13.6.3); the deferred work is DW-311.
- **Navigation controls live in the bottom status bar** — page stepper, page indicator, zoom — not on a
  toolbar above the page, so the page area carries the page alone. A fit choice persists across page
  changes until the author zooms manually; beyond viewport size the viewer's own height-constrained
  container scrolls on both axes with the page centred when smaller, so the outer region never scrolls
  in its place.
- **Chrome states the Preview-mode fact.** In Preview the document bar shows render freshness
  ("rendered 412 ms ago · current", or stale in the same words the status line uses) rather than the
  Design-mode page setup.
- **Two-accent grammar holds:** cyan means structure, focus and authority; amber means data only. Every
  machine-readable value — hash, path, byte count, filename — is set in mono.
- **Voice is terse and technical**: state the fact, name the location, offer no comfort.

## Cross-Story Dependencies

- 13.1–13.5 are delivered; 13.6 is the remaining story and lands on top of all of them.
- 13.6 completes the PAGES rail that 13.3 scoped but did not build — 13.3 owns the evidence rail and the
  screen frame, 13.6 owns the thumbnail rail itself. **It does NOT own the diagnostic-to-page marking**,
  which D-13.6.3 deferred to DW-311 because no data source for it exists.
- 13.6 must reconcile with 13.2 — **but not as written here.** `PDFViewer` was never adopted, and
  `viewer-navigation.ts` was never a page-state authority in the first place: measured, it is 108 lines of
  pure arithmetic with no imports, no React, no DOM and no state at all. It **shipped untouched.** The sole
  page-state authority is `App.tsx`'s `previewViewState`, written through one funnel, and the rail reads
  from it. The status-bar page/zoom controls 13.2 placed must read from the same single authority,
  and the `src/preview/` exception list in the canvas-authority contract test — a shared guard other
  epics also assert against — must still be extended by name rather than by wildcard.
- ~~13.6's diagnostic marking depends on 13.3's diagnostic model and on the page/path/element/band location
  the diagnostic cards already carry.~~ **FALSE ON BOTH HALVES.** There is no diagnostic marking in 13.6
  (DW-311), and the diagnostic cards carry a **three-part** location — `dataPath · kind · band`. There is no
  page in it, deliberately; `diagnostic-presenter.tsx` says so in a comment. The four-part claim came from
  the UX mockup and was never true of the code.
- 13.3 re-dresses the diagnostic card and its Locate-on-canvas / Dismiss behaviour already built earlier
  in the designer; it changes presentation, not that behaviour.
- 13.1's Save PDF and 13.3's evidence rail land the same paired action row — 13.3 owns its placement,
  13.1 owns the export itself; both are governed by the freshness rule.
- 13.5 removes the in-preview return button that 13.3's screen rework also touches; the document-bar
  mode switch becomes the sole mode control.
- 13.4 depends on the preview freshness/identity key so that loading sample data later marks the
  no-data preview stale like any other input change, and must not disturb the engine's error contract or
  the golden corpus owned by the engine epics.
- 13.2 depends on the `src/preview/` exception list in the canvas-authority contract test being extended
  by name, which is a shared guard other epics also assert against.
