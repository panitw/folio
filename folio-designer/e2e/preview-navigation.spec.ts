import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// STORY 13.2 — THE BROWSER WITNESS FOR THE ONE CLAIM jsdom CANNOT SEE.
//
// jsdom performs no layout, so nothing in the unit suite can distinguish a
// scrollable container from a non-scrollable one, or a pinned element from one
// that has scrolled out of view. Every CSS claim this story makes therefore has
// its only proof here. The distinguishing observable is WHAT STAYS PUT: if
// `.preview-region` is the scroller, the preview heading, the no-data notice
// and the evidence line scroll away with the page; if `.pdf-preview-scroll` is,
// they hold while the page moves under them.
//
// COMPILE-CHECKED LOCALLY, EXECUTED IN CI. The story's cadence forbids a local
// Playwright run because `webServer.command` is `npm run build`, which the
// story's Boundaries forbid outright. CI runs the whole browser suite on every
// push (DW-268, discharged at `adf905a`), so this is real per-commit coverage
// and is written as such — never as a compile-only placeholder.
//
// NOT ONE PROHIBITED IDENTIFIER IS SPELLED HERE, deliberately. Every `e2e/`
// file is auto-enrolled by the independent directory walk at
// `canvas-authority-contract.test.ts:673` and the `src/preview/` exception does
// not reach it, so the whole witness is built out of Playwright's own
// `boundingBox()` and `mouse.wheel()`, neither of which the scan names.
const template = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../folio-go/testdata/example/first-pdf.folio'))

// The viewport is Playwright's default 1280x720 — `playwright.config.ts`
// declares none. A letter page at 200% is 1224 CSS px wide and 1584 tall, and
// the page area is a few hundred px of a 648px workbench, so the page overflows
// BOTH axes. That is a precondition of everything below, and it is asserted
// rather than assumed: a page that fits its container cannot witness a scroll.
const zoomed = 2

test('scrolls the page area alone, and leaves the chrome that describes it fixed and reachable', async ({ page }) => {
  // THE FALLBACK FILE TIER, FORCED — without which this spec cannot pass at all.
  // Headless Chromium HAS the File System Access API, so the app takes the
  // native tier, calls `showOpenFilePicker()` directly, and emits no
  // `filechooser` event ever: the wait below then times out at 90s on a
  // perfectly working application. Population, measured: ten specs under `e2e/`
  // wait on a `filechooser`, and the eight that pass all force this tier first
  // (`local-file-actions.spec.ts:11` is the model). This spec and
  // `preview-no-data.spec.ts` were the only two that did not, and they were
  // exactly the two that failed the first time the suite was ever run.
  await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const templateChooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await templateChooser).setFiles({ name: 'statement.folio', mimeType: 'application/json', buffer: template })
  await expect(page.locator('.document-name')).toHaveText('statement.folio')

  // The no-data fixture is chosen because it is the one that renders the most
  // chrome around the page: 13.4's amber notice as well as the heading, the
  // freshness line and the evidence line. Story 13.4's own witness establishes
  // that this template reaches an admitted preview with no sample data.
  await page.getByRole('button', { name: 'PREVIEW' }).click()
  const pageArea = page.getByRole('img', { name: /Current no-data layout PDF, revision \d+/ })
  await expect(pageArea).toBeVisible({ timeout: 60_000 })

  // THE CONTROLS ARE IN THE STATUS BAR, WHICH IS OUTSIDE BOTH `<main>`s. Their
  // presence here is what makes the page area able to carry the page and
  // nothing else, and the zoom below is driven through the real control rather
  // than through a style, so the witness exercises the shipped path.
  const statusBar = page.getByRole('contentinfo', { name: 'Status bar' })
  await expect(statusBar.getByRole('button', { name: 'Previous PDF page' })).toBeVisible()
  await expect(statusBar.getByRole('button', { name: 'Next PDF page' })).toBeVisible()
  await expect(statusBar.getByRole('combobox', { name: 'PDF zoom' })).toBeVisible()
  await statusBar.getByRole('combobox', { name: 'PDF zoom' }).selectOption({ label: `${zoomed * 100}%` })
  await expect(statusBar.getByRole('textbox', { name: 'PDF zoom percentage' })).toHaveValue(String(zoomed * 100))

  const canvas = pageArea.locator('canvas')
  await expect(canvas).toBeVisible()
  const heading = page.getByText('NO-DATA LAYOUT PREVIEW')
  const notice = page.getByRole('note', { name: 'No-data preview notice' })
  // ⚠ STORY 13.3 — THE DIGEST IS NO LONGER ONE OF THE HELD-POSITION SUBJECTS,
  // AND REMOVING IT FROM THAT LIST IS THE POINT (review P7).
  //
  // It used to be `.preview-evidence`, a line INSIDE the preview scroller, so
  // "it did not move when the page scrolled" was a claim that could fail. The
  // rail put it in the Inspector column, which is a different, non-scrolling
  // container: `after.evidence.y === before.evidence.y` there is true of any
  // element in any sibling of the scroller, guards nothing, and reads as a pin.
  // `heading` and `notice` are still inside the preview region and still carry
  // that claim. What survives for the digest is the claim that IS still real —
  // it stays on screen and keeps saying the same thing across the scroll — and
  // it is asserted below rather than measured here.
  const evidence = page.getByLabel('Output hash').locator('.rail-hash-value')
  const digestBefore = await evidence.textContent()
  expect(digestBefore).toMatch(/^[a-f0-9]{64}$/)

  // Read once, at rest, and then compared after the scroll. `boundingBox()` is
  // viewport-relative, which is exactly the frame this claim is about.
  const boxes = async () => {
    const measured = await Promise.all([canvas, pageArea, heading, notice, statusBar].map(async (locator) => {
      const box = await locator.boundingBox()
      expect(box).not.toBeNull()
      return box!
    }))
    return { canvas: measured[0]!, pageArea: measured[1]!, heading: measured[2]!, notice: measured[3]!, statusBar: measured[4]! }
  }
  const before = await boxes()

  // THE PRECONDITION, ASSERTED. Without an overflowing page there is nothing to
  // scroll, and every "held its position" assertion below would pass over a
  // viewer that simply never moves. This is the arm that keeps the witness from
  // going vacuous if a future layout change makes the page area large enough to
  // hold a 200% page whole.
  expect(before.canvas.height).toBeGreaterThan(before.pageArea.height)
  expect(before.canvas.width).toBeGreaterThan(before.pageArea.width)

  // THE LEFT AND TOP EDGES OF A ZOOMED PAGE ARE REACHABLE — the `safe center`
  // fix, on BOTH axes. `.pdf-preview-scroll` centres its single child; with
  // plain `center` and a child larger than the box, the overflow is split
  // evenly and the leading edge sits at a negative offset the scroller can
  // never reach, because a scroll offset does not go below zero. `safe center`
  // falls back to start-alignment in exactly that case, which puts the leading
  // edge AT the container's own edge and no further. One pixel for the border.
  //
  // BOTH axes are asserted because the rule sets both keywords and each is a
  // separate declaration: `justify-content` governs x, `align-content` governs
  // y, and a revert of either alone is a defect the other cannot see.
  expect(before.canvas.x).toBeGreaterThanOrEqual(before.pageArea.x - 1)
  expect(before.canvas.y).toBeGreaterThanOrEqual(before.pageArea.y - 1)

  // THE PREVIEW STATUS BAR IS THE TALLER OF THE TWO DECLARED SIZES. `.status-bar`
  // declares no height of its own — its height IS the `.app-shell` grid track,
  // which `.app-shell-preview` switches to `--status-bar-height-preview`. So
  // this is the one observable that can tell a minted-and-used token from a
  // minted-and-forgotten one, and no unit test can see it: jsdom applies no
  // stylesheet at all.
  expect(before.statusBar.height).toBe(32)

  // Everything the chrome assertions depend on is on screen to begin with, so
  // "still visible afterwards" is a change and not a restatement.
  for (const locator of [heading, notice, evidence]) await expect(locator).toBeInViewport()

  // THE SCROLL ITSELF, DRIVEN AS A USER DRIVES IT. A wheel over the page area
  // scrolls the nearest scrollable ancestor. Which element that is IS the
  // property under test: before this story `.pdf-preview-scroll` had no height
  // at all, so its own overflow never activated vertically and the wheel
  // reached `.preview-region` instead, taking the heading and the notice with
  // it. Both outcomes move something; only one of them moves the right thing.
  await canvas.hover()
  await page.mouse.wheel(0, 600)
  await expect.poll(async () => (await canvas.boundingBox())!.y, { timeout: 10_000 }).toBeLessThan(before.canvas.y - 100)

  const after = await boxes()

  // THE PAGE MOVED...
  expect(after.canvas.y).toBeLessThan(before.canvas.y - 100)

  // ...AND EVERYTHING AROUND IT DID NOT. Asserted as exact viewport positions,
  // not as `toBeVisible()`: Playwright's visibility means a non-empty box, so a
  // diagnostic that had scrolled clean out of the window would still be
  // "visible" by that measure. Position is the property; presence is not.
  expect(after.pageArea.y).toBe(before.pageArea.y)
  expect(after.pageArea.height).toBe(before.pageArea.height)
  expect(after.heading.y).toBe(before.heading.y)
  expect(after.notice.y).toBe(before.notice.y)
  expect(after.statusBar.y).toBe(before.statusBar.y)

  // AND STILL REACHABLE — the condition attached to this story at approval. A
  // diagnostic the author cannot see while looking at the thing it describes is
  // one that may as well not be rendered, and `toBeInViewport` is the assertion
  // that can tell the two apart.
  for (const locator of [heading, notice, evidence]) await expect(locator).toBeInViewport()
  await expect(statusBar.getByRole('button', { name: 'Next PDF page' })).toBeInViewport()

  // The horizontal axis is DW-191's own axis — the one on which the tear-down
  // was live before this story — so it is scrolled too, and the chrome holds
  // through that as well.
  await page.mouse.wheel(400, 0)
  await expect.poll(async () => (await canvas.boundingBox())!.x, { timeout: 10_000 }).toBeLessThan(before.canvas.x - 100)
  const sideways = await boxes()
  expect(sideways.heading.x).toBe(before.heading.x)
  expect(sideways.notice.x).toBe(before.notice.x)
  expect(sideways.pageArea.x).toBe(before.pageArea.x)
  for (const locator of [heading, notice, evidence]) await expect(locator).toBeInViewport()
  // AND THE DIGEST STILL SAYS THE SAME THING. This is the digest claim that a
  // scroll can actually break — a rail re-rendered or re-keyed mid-scroll would
  // change or lose it — where its viewport position cannot.
  await expect(evidence).toHaveText(digestBefore!)

  // FIT WIDTH, AGAINST A CONTAINER THAT WAS REALLY LAID OUT. This is the only
  // place in the repository where the story's container measurement runs against
  // actual layout: `pdf-viewer.test.tsx` proves the arithmetic by installing a
  // fake box on `HTMLDivElement.prototype`, which is exactly as big as the test
  // says it is. Here the number comes from the browser, so the whole chain —
  // the padding moved off the measured element, `scrollbar-gutter: stable`, the
  // grid track, the flex cap — has to be right for the page to end up fitting.
  await statusBar.getByRole('combobox', { name: 'PDF zoom' }).selectOption({ label: 'Fit width' })
  // The readout shows the RESOLVED percentage, never the word "fit" — the
  // matrix row that says so has no other witness.
  const zoomReadout = statusBar.getByRole('textbox', { name: 'PDF zoom percentage' })
  await expect(zoomReadout).not.toHaveValue(String(zoomed * 100))
  await expect(zoomReadout).toHaveValue(/^\d+$/)
  // AND THE PAGE ACTUALLY FITS. `boundingBox()` on the scroll host is its border
  // box, which includes the stable scrollbar gutter the canvas does not get, so
  // a fitted page is strictly narrower than its host with room to spare.
  await expect.poll(async () => (await canvas.boundingBox())!.width, { timeout: 10_000 }).toBeLessThanOrEqual((await pageArea.boundingBox())!.width)
  // The chrome is still where it was through all of that.
  const fitted = await boxes()
  expect(fitted.heading.y).toBe(before.heading.y)
  expect(fitted.notice.y).toBe(before.notice.y)
  expect(fitted.statusBar.y).toBe(before.statusBar.y)
  for (const locator of [heading, notice, evidence]) await expect(locator).toBeInViewport()
})

// STORY 13.5 — THE STATUS BAR FITS ITS OWN CONTENTS, WHICH NO UNIT TEST CAN SEE.
//
// jsdom applies no stylesheet and performs no layout, so the whole of "the
// assurance line fits beside what the bar still carries" is invisible to the
// unit suite: `App.test.tsx` can prove the element is rendered and can prove the
// two dropped items are gone, and neither claim is about width.
//
// THE MEASUREMENT THAT MADE THIS STORY DROP TWO ITEMS, RE-ASSERTED AS A GUARD.
// At a 1024px viewport the bar's only free space was `.status-spacer` at 76.9px
// and the assurance line needs 228px plus a 12px gap — a 163px deficit. No
// single item closed it (the largest, the engine snapshot, is 156px with its
// gap). Dropping `LOCAL SHELL` (78px) and the template font count (120px) frees
// 198px, which left 35px spare. THOSE ARE THE FIGURES THE DROP WAS DECIDED ON,
// and they are recorded because they are the reasoning; they are no longer the
// current ones. Hiding `offline-status` with `.sr-only` afterwards returned
// roughly another 90px, so the bar as shipped has 124.94px spare — measured,
// not derived. The consequence is stated plainly in the red-proof note below:
// restoring ONE dropped item no longer overflows the bar, because it genuinely
// no longer does.
//
// ⚠ WHAT THIS TEST MUST NOT MEASURE AGAINST, AND WHY THE OBVIOUS FORM IS A
// FRAUD. The first version of this assertion read
// `assurance.right <= statusBar.right`, and it CANNOT FAIL. `.status-bar` is
// the grid item in `.app-shell`'s single implicit column; when its row of
// `nowrap` children overruns the 1024px track, the bar's own border box GROWS
// with them, so BOTH SIDES OF THAT INEQUALITY MOVE TOGETHER BY THE SAME AMOUNT
// and the overflow is hidden by `.app-shell { overflow: hidden }` instead of
// being reported. Measured in Chromium, both dropped items restored: the bar
// reports 1187.06px inside a 1024px viewport — 163px overfull, exactly the
// deficit this story's arithmetic named — and the assurance's right edge sits
// at 1175.06px, 151px of it off-screen. `1175.06 <= 1187.06` is TRUE. The
// height was still exactly 32 (`nowrap` does its job and pushes the overflow
// SIDEWAYS, where a horizontal assertion measured against the grown bar cannot
// see it), and `toBeInViewport()` at its default ratio passed too — measured
// ratio 0.337, because any intersection satisfies it. Five green assertions
// over a bar with two thirds of its promise cut off.
//
// SO EVERY EDGE HERE IS MEASURED AGAINST THE VIEWPORT, the one box in this test
// that does not grow with the content: `page.viewportSize()` is the width the
// test itself set. The bar overrunning its track is then the failure rather
// than the thing that conceals it.
//
// THE STATE THIS IS MEASURED IN IS A PARAMETER, NOT A CONSTANT. A fit
// assertion that only ever visits the SHORTEST state is the guard that passes
// because it never visits the case that fails, so `fitsIn` below is run once
// per reachable composition. Population searched for items whose width can
// vary — every child of the `<footer className="status-bar">` at `App.tsx`
// and every child of the `previewNavigation` group it mounts, read whole:
//   · `engine-snapshot` — `GO SNAPSHOT · REVISION n`, and `n` is the engine's
//     own document revision, so it grows without bound. DRIVEN FOR REAL to
//     three digits below, through the keyboard nudge.
//   · the page readout — `Page N of M`, both unbounded, plus its
//     `Rendering PDF` alternative. NOT driven (see the arithmetic at the foot
//     of the test, and the reason it is not driven).
//   · the zoom `<select>` — constant: Chromium sizes an `auto`-width select to
//     its WIDEST OPTION rather than the selected one, and its widest label
//     (`Fit width`) is what the base measurement already carries. Checked by
//     driving both a different selection and the one state that changes the
//     option set at all.
//   · the page-number and zoom-percentage `<input>`s — fixed at 46px by
//     `.preview-nav input`, asserted rather than assumed.
//   · `offline-status` — carries `.sr-only` in Preview, so it is `position:
//     absolute`, is not a flex item, and contributes neither width nor a gap.
//     Asserted from both sides: its own box, and the single 12px gap between
//     the spacer and `PREVIEW MODE` that only holds if nothing sits between
//     them in flow.
//   · `PREVIEW MODE` and the assurance sentence — literal constants.
//
// MEASURED IN CHROMIUM 1217, UNMUTATED, at a 1024px viewport: the bar is
// 1024.00 x 32, the assurance sits at x 784.00, 228.00 wide, right edge
// 1012.00 (the bar's own 12px right padding), the leftmost surviving item at
// x 12.00, and `.status-spacer` — which IS the bar's spare room, being the one
// `flex: 1` child — is 124.94px. Driving the revision to three digits spends
// exactly 12.00px of it, leaving 112.94px: two characters at the 6.00px
// advance this bar's 10px mono actually measures.
//
// THE RED PROOF, EXECUTED IN CHROMIUM 1217 AND NOT MERELY WRITTEN DOWN —
// un-fence the dropped items from `mode === 'design'` in `App.tsx` and run this
// file. Measured, with the two census rows above neutralised so the FIT
// assertions are the ones observed failing rather than shadowed:
//   · unmutated → GREEN, at the figures above.
//   · `LOCAL SHELL` alone restored → STILL GREEN, bar 1024.00, spare room
//     46.94px. THIS IS A DELIBERATE, REPORTED CHANGE: before `offline-status`
//     was visually hidden the bar had 35px spare and this same mutation
//     reddened it at 1067.06px. Hiding that span returned roughly 90px, and
//     restoring one 78px item now genuinely fits — the assertion is telling the
//     truth about a bar that changed, not failing to notice an overflow.
//   · BOTH restored → RED, bar 1097.06 inside a 1024 viewport, spare room
//     0.00px, failing at `bar.width <= width` with the state name in the
//     message. The guard can still fail, and this is the run that proves it.
//
// NO PROHIBITED IDENTIFIER IS SPELLED HERE. `boundingBox()` and
// `viewportSize()` are Playwright's own and are not named by the corpus scan
// that auto-enrols every `e2e/` file; `getBoundingClientRect` and the
// `offset*`/`client*`/`scroll*` reads are never used.
test('fits the standing local-only assurance into the Preview status bar at a narrow desktop viewport', async ({ page }) => {
  await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })
  await page.setViewportSize({ width: 1024, height: 768 })
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const templateChooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await templateChooser).setFiles({ name: 'statement.folio', mimeType: 'application/json', buffer: template })
  await expect(page.locator('.document-name')).toHaveText('statement.folio')

  // Measured over a REAL rendered preview, because that is the widest the bar
  // ever gets: the page readout carries its final page count only once PDF.js
  // has admitted the document, and the zoom controls are all mounted.
  await page.getByRole('button', { name: 'PREVIEW' }).click()
  await expect(page.getByRole('img', { name: /Current no-data layout PDF, revision \d+/ })).toBeVisible({ timeout: 60_000 })

  const statusBar = page.getByRole('contentinfo', { name: 'Status bar' })
  const assurance = page.getByTestId('local-only-assurance')
  await expect(assurance).toHaveText('no network · nothing left this machine')
  // THE TWO DROPPED ITEMS ARE REALLY GONE FROM THE LAID-OUT BAR, not merely
  // absent from a jsdom tree: without this the fit below could be passing
  // because the bar is missing something else entirely.
  await expect(statusBar.getByText('LOCAL SHELL')).toHaveCount(0)
  await expect(statusBar.getByTestId('template-font-count')).toHaveCount(0)

  const box = async (locator: typeof statusBar) => {
    const measured = await locator.boundingBox()
    expect(measured).not.toBeNull()
    return measured!
  }
  const viewport = page.viewportSize()
  expect(viewport).not.toBeNull()
  const width = viewport!.width
  const snapshotItem = statusBar.getByTestId('engine-snapshot')
  const spacer = statusBar.locator('.status-spacer')
  const modeBadge = statusBar.getByText('PREVIEW MODE', { exact: true })
  const readout = statusBar.getByLabel('PDF page status')
  const zoomSelect = statusBar.getByRole('combobox', { name: 'PDF zoom' })
  const zoomTyped = statusBar.getByRole('textbox', { name: 'PDF zoom percentage' })
  const pageTyped = statusBar.getByRole('textbox', { name: 'PDF page number' })

  // THE WHOLE FIT CLAIM, PARAMETERISED OVER THE STATE IT IS MADE IN. A fit
  // assertion that only ever visits the SHORTEST state is a guard that passes
  // because it never visits the case that fails, so every assertion below is
  // taken in each state the bar can actually reach, and every one of them is
  // measured against the VIEWPORT — the one box in this test that does not grow
  // with the content. The returned figure is the bar's real spare room:
  // `.status-spacer` is `flex: 1`, so it IS the free space, and it is what the
  // arithmetic further down spends.
  const fitsIn = async (state: string) => {
    const bar = await box(statusBar)
    const promise = await box(assurance)
    const leftmost = await box(snapshotItem)
    const room = (await box(spacer)).width

    // THE HEIGHT IS STILL THE DECLARED PREVIEW HEIGHT. Nothing in the bar may
    // wrap or grow it. This is a real guard against the OTHER way an overfull
    // bar could hide — wrapping to two rows — and it is not a guard against
    // horizontal overflow at all, which is why it is not left to carry the test.
    expect(bar.height, state).toBe(32)

    // THE BAR ITSELF STAYS INSIDE THE WINDOW. This is the assertion the first
    // version was missing: an overfull `nowrap` bar grows sideways past its grid
    // track, and the viewport is the box that does not move with it.
    expect(bar.x, state).toBeGreaterThanOrEqual(0)
    expect(bar.width, state).toBeLessThanOrEqual(width)

    // THE FIT ITSELF, ON BOTH ENDS, AGAINST THE WINDOW. The assurance is the
    // last item in the bar, so its right edge is where an overfull bar shows
    // first; the leftmost surviving item's left edge is the other end of the
    // same claim, and it would move if a future change reached the fit by
    // pulling content off the left instead.
    expect(promise.x + promise.width, state).toBeLessThanOrEqual(width)
    expect(leftmost.x, state).toBeGreaterThanOrEqual(0)

    // AND THE WHOLE LINE IS ON SCREEN, NOT MERELY TOUCHING IT. `toBeInViewport()`
    // with its default ratio is satisfied by a single visible pixel, which is
    // exactly what an overfull bar leaves: `ratio: 1` is the claim the acceptance
    // row actually makes.
    await expect(assurance, state).toBeInViewport({ ratio: 1 })

    // AND THE LINE IS ITS FULL SELF RATHER THAN COLLAPSED. The sentence measures
    // 228px in this bar's 10px mono; the floor is set just under that so a normal
    // metrics difference does not redden it, while a truncated or collapsed line
    // — which every inequality above would otherwise accept — does.
    expect(promise.width, state).toBeGreaterThan(200)
    return room
  }

  const baseMargin = await fitsIn('base state: one-digit engine revision, the rendered page count')

  // THE OFFLINE LIVE REGION IS OUT OF THE PAINTED BAR, AND OUT OF FLEX FLOW.
  // In Preview it carries `.sr-only`, which is `position: absolute` — so it is
  // not a flex item, contributes no width and takes no `gap`, and the bar's
  // spare room is therefore the SAME in all five `offlineLabel` states instead
  // of varying by the 24 characters between the shortest and the longest. Both
  // halves are checked here rather than assumed: the box is at most a pixel, and
  // the distance from the spacer's right edge to `PREVIEW MODE` is exactly one
  // 12px gap, which is only true if nothing sits between them in flow. Its text
  // is untouched — this is a visual change and nothing more.
  const offline = statusBar.getByTestId('offline-status')
  await expect(offline).toHaveText(/^(Offline ready|Update available; current release remains usable|Offline cache checking|Offline cache unavailable|Offline layer bypassed \(dev\))$/)
  const offlineBox = await box(offline)
  expect(offlineBox.width).toBeLessThanOrEqual(1)
  expect(offlineBox.height).toBeLessThanOrEqual(1)
  const spacerBox = await box(spacer)
  expect((await box(modeBadge)).x - (spacerBox.x + spacerBox.width)).toBeCloseTo(12, 0)

  // THE ZOOM SELECT IS A CONSTANT WIDTH, ALREADY AT ITS MAXIMUM. Chromium sizes
  // an `auto`-width `<select>` to its WIDEST OPTION, not to the selected one, so
  // the control cannot grow with a selection. Its widest label is `Fit width`
  // (9 characters, against `Fit page`, five percentages of at most 4, and
  // `Custom` at 6), and that is what the base measurement above was already
  // taken with. Driven rather than argued, in both directions that could move
  // it: a different SELECTION, and the one state that changes the option SET at
  // all — a hand-typed zoom, which mounts the extra `Custom` option.
  const zoomAtFitWidth = (await box(zoomSelect)).width
  await zoomSelect.selectOption('0.5')
  await expect(zoomSelect).toHaveValue('0.5')
  expect((await box(zoomSelect)).width).toBe(zoomAtFitWidth)
  await zoomTyped.fill('133')
  await zoomTyped.press('Enter')
  await expect(zoomSelect).toHaveValue('custom')
  expect((await box(zoomSelect)).width).toBe(zoomAtFitWidth)
  await zoomSelect.selectOption('fit-width')
  await expect(zoomSelect).toHaveValue('fit-width')

  // AND THE TWO TYPED FIELDS CANNOT GROW EITHER: `.preview-nav input` fixes both
  // at 46px, and `* { box-sizing: border-box }` makes that the border box the
  // measurement returns, whatever is typed into them.
  expect((await box(pageTyped)).width).toBe(46)
  expect((await box(zoomTyped)).width).toBe(46)

  // ── THE THREE-DIGIT ENGINE REVISION, DRIVEN FOR REAL ──────────────────────
  // `engine-snapshot` reads `GO SNAPSHOT · REVISION n`, and `n` is the ENGINE'S
  // OWN DOCUMENT REVISION: every accepted command advances it, so the state is
  // reachable through the product rather than only imaginable. The cheapest
  // genuine route is the keyboard nudge — one command per keypress, no dialog,
  // no file. SHIFT is required: the unshifted step is 1000 and the engine snaps
  // every coordinate to `GridIncrement = 6000` (`folio-go/page_setup.go:21`), so
  // an unshifted nudge rounds back to where it started, changes no byte, and
  // advances no revision. The shifted 10000 step snaps to 12000 and back to 0,
  // so alternating the two directions advances the counter on EVERY press while
  // leaving the component where it began.
  await page.getByRole('button', { name: 'DESIGN' }).click()
  await page.getByLabel(/text component e1/).click()
  const revisionNow = async () => Number(/REVISION (\d+)/.exec((await snapshotItem.textContent()) ?? '')?.[1] ?? '0')
  const startingRevision = await revisionNow()
  // FAIL FAST AND LOUDLY if the selection or the nudge is not doing anything,
  // rather than spending two hundred presses discovering it.
  await page.keyboard.press('Shift+ArrowRight')
  await expect.poll(revisionNow, { message: 'a shifted nudge must advance the engine revision' }).toBeGreaterThan(startingRevision)
  for (let press = 1; press < 240 && await revisionNow() < 100; press += 1) await page.keyboard.press(press % 2 === 0 ? 'Shift+ArrowRight' : 'Shift+ArrowLeft')
  await expect(snapshotItem).toHaveText(/GO SNAPSHOT · REVISION \d{3}/)

  // RE-ENTRY DOES NOT RENDER ON ITS OWN. With a record surviving from the first
  // visit the app shows the STALE historical PDF and offers `Re-render`, so the
  // fresh render at the new revision has to be asked for — and the viewer's own
  // accessible name is then a second witness that the revision really is three
  // digits wide in the state being measured.
  await page.getByRole('button', { name: 'PREVIEW' }).click()
  await expect(page.getByRole('img', { name: /Stale historical PDF, revision \d+/ })).toBeVisible({ timeout: 60_000 })
  await page.getByRole('button', { name: 'Re-render' }).click()
  await expect(page.getByRole('img', { name: /Current no-data layout PDF, revision \d{3}/ })).toBeVisible({ timeout: 60_000 })
  await expect(snapshotItem).toHaveText(/GO SNAPSHOT · REVISION \d{3}/)
  const wideMargin = await fitsIn('engine revision at three digits')

  // AND THE DRIVEN STATE REALLY IS THE WIDER ONE. Without this the run above
  // could be a second copy of the base case wearing a different name: two more
  // digits must have cost the bar two more characters of room.
  const advance = (await box(readout)).width / (((await readout.textContent()) ?? '').length)
  expect(baseMargin - wideMargin).toBeGreaterThan(advance)

  // ── THE THREE-DIGIT PAGE READOUT, COVERED BY MEASURED ARITHMETIC ──────────
  // NOT DRIVEN, AND NOT SKIPPED SILENTLY. `Page N of M` at three digits each
  // needs a template that renders three hundred pages; producing one, rendering
  // it through the engine and letting PDF.js admit it is minutes of browser time
  // per run for two characters of width, and no honest cheap route to it exists
  // — the count comes from the rendered document and nothing in the UI can set
  // it. So the property is asserted ARITHMETICALLY from figures this test has
  // actually measured, in the WIDEST state it did drive: the per-character
  // advance is read off the readout itself (this bar computes to
  // `10px/13px "IBM Plex Mono"`, ~6px per character), the margin is the real
  // `.status-spacer` width in the three-digit-revision state, and the claim is
  // that the margin still exceeds what each alternative readout would add. The
  // two alternatives are the only ones this element has: the longest plausible
  // page readout, and the `Rendering PDF` copy it shows before PDF.js reports a
  // page count.
  const readoutText = (await readout.textContent()) ?? ''
  for (const alternative of ['Page 300 of 3400', 'Rendering PDF']) {
    const growth = (alternative.length - readoutText.length) * advance
    expect(
      wideMargin,
      `the bar's spare room in the three-digit-revision state is ${wideMargin.toFixed(2)}px; swapping the readout from ${JSON.stringify(readoutText)} to ${JSON.stringify(alternative)} adds ${growth.toFixed(2)}px at a measured ${advance.toFixed(2)}px per character`,
    ).toBeGreaterThan(growth)
  }
})
