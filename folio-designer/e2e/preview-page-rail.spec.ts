import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// STORY 13.6 — THE BROWSER WITNESS THAT THE VENDORED PDF.JS CODE ACTUALLY WORKS.
//
// This is the ONLY test that can see a thumbnail. Measured in jsdom 28.1.0:
// `canvas.getContext('2d')` returns null, and `OffscreenCanvas`,
// `createImageBitmap` and `ImageBitmap` are all undefined — so
// `PDFThumbnailView.draw` cannot complete there and the unit suite stubs it. Every
// claim below therefore has its only proof here: that the two vendored files link
// against `pdfjs-dist/build/pdf.mjs` at all, that `#convertCanvasToImage` produces
// an image, and that the rail is 132px of real layout rather than a rule nothing
// applies.
//
// ⚠ WRITTEN AND COMPILE-CHECKED IN 13.6, EXECUTED AT THE EPIC 13 BOUNDARY GATE.
// Under D-000.33 the story's own cadence runs `npm test`, `tsc -b` and
// `tsc -p tsconfig.e2e.json --noEmit` only; `npm run test:e2e` needs
// `webServer.command` (`npm run build`), which the gate owns. Compile-clean is
// NOT passing, and this file has never been executed by anything.
//
// NOT ONE PROHIBITED IDENTIFIER IS SPELLED HERE. Every `e2e/` file is
// auto-enrolled in the AD-17 corpus by an independent directory walk, and the
// `src/preview/` waiver does not reach it, so the geometry is Playwright's own
// `boundingBox()` and nothing else.

// FIVE PAGES, FROM STATIC TEXT AND NO DATA BINDING AT ALL. The rail's whole
// point is a document longer than one page, and page-count-5's five markers are
// positioned by y-offset rather than produced by a repeat over data — so this
// spec witnesses the rail without also depending on sample data being loaded.
const template = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../fixtures/page-count-5/input.folio'))

test('enumerates the render as rasterised page thumbnails, and navigates by them', async ({ page }) => {
  // THE FALLBACK FILE TIER, FORCED — without which this spec cannot pass at all.
  // Headless Chromium HAS the File System Access API, so the app takes the native
  // tier, calls `showOpenFilePicker()` directly, and emits no `filechooser` event
  // ever: the wait below then times out at 90s on a perfectly working
  // application. `local-file-actions.spec.ts:11` is the model.
  await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await chooser).setFiles({ name: 'five-pages.folio', mimeType: 'application/json', buffer: template })
  await expect(page.locator('.document-name')).toHaveText('five-pages.folio')

  // IN DESIGN, THE PALETTE IS THE FIRST COLUMN. Asserting it here is what makes
  // its absence in Preview a change rather than a restatement of a template that
  // never had one.
  await expect(page.getByLabel('Component palette')).toBeVisible()
  await expect(page.getByLabel('Page thumbnails')).toHaveCount(0)

  await page.getByRole('button', { name: 'PREVIEW' }).click()
  const pageArea = page.getByRole('img', { name: /PDF, revision \d+/ })
  await expect(pageArea).toBeVisible({ timeout: 60_000 })
  const statusBar = page.getByLabel('Status bar')
  await expect(statusBar.getByLabel('PDF page status')).toHaveText('Page 1 of 5', { timeout: 60_000 })

  // THE COLUMN SWAPPED, AND THE PALETTE IS GONE FROM THE DOCUMENT.
  const rail = page.getByLabel('Page thumbnails')
  await expect(rail).toBeVisible()
  await expect(rail.getByText('PAGES')).toBeVisible()
  await expect(page.getByLabel('Component palette')).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Place Text' })).toHaveCount(0)

  // FIVE ENTRIES, AND NO TRUNCATION LINE — five is under the bound of twelve.
  //
  // ⚠ `exact: true` ON EVERY ENTRY QUERY. Playwright's `getByRole({ name })`
  // matches a SUBSTRING by default, so `Page 1` also matches `Page 10`–`Page 12`.
  // On this five-page fixture that is invisible; on any document past nine pages
  // every one of these becomes a strict-mode violation. The bound is twelve, so
  // that is a document this rail is expected to show.
  for (const number of [1, 2, 3, 4, 5]) await expect(rail.getByRole('button', { name: `Page ${number}`, exact: true })).toBeVisible()
  await expect(rail.getByRole('button', { name: 'Page 6', exact: true })).toHaveCount(0)
  await expect(rail.locator('.page-rail-more')).toHaveCount(0)

  // THE RASTERISATION ITSELF, WHICH IS WHY THIS FILE EXISTS. `#convertCanvasToImage`
  // sets the `img`'s `src` to an object URL and removes `missingThumbnailImage`
  // from the container, and it does that ONLY after `pdfPage.render` resolved. So
  // the two assertions together are the claim that the vendored code ran end to
  // end against the PDF the Go engine produced.
  const firstWell = rail.locator('.page-rail-well').first()
  await expect(firstWell.locator('img')).toHaveAttribute('src', /^blob:/, { timeout: 60_000 })
  await expect(firstWell.locator('.missingThumbnailImage')).toHaveCount(0)
  // Every entry, not just the first: the pass draws all five.
  await expect(rail.locator('.page-rail-well .missingThumbnailImage')).toHaveCount(0, { timeout: 60_000 })

  // AND THE MARKING IS ON PAGE ONE ALONE, from `previewViewState` and nothing
  // else.
  await expect(rail.locator('[aria-current="page"]')).toHaveCount(1)
  await expect(rail.getByRole('button', { name: 'Page 1', exact: true })).toHaveAttribute('aria-current', 'page')

  // NAVIGATION: clicking entry 3 moves the page the status bar reports, which is
  // the same `previewViewState` the viewer reads. A rail with its own page value
  // would move one and not the other.
  await rail.getByRole('button', { name: 'Page 3', exact: true }).click()
  await expect(statusBar.getByLabel('PDF page status')).toHaveText('Page 3 of 5')
  await expect(statusBar.getByRole('textbox', { name: 'PDF page number' })).toHaveValue('3')
  await expect(rail.getByRole('button', { name: 'Page 3', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(rail.locator('[aria-current="page"]')).toHaveCount(1)

  // And back the other way: the status bar writes and the rail's marking follows,
  // which is the reader-only property stated as an observation rather than a
  // comment.
  await statusBar.getByRole('textbox', { name: 'PDF page number' }).fill('5')
  await statusBar.getByRole('textbox', { name: 'PDF page number' }).press('Enter')
  await expect(rail.getByRole('button', { name: 'Page 5', exact: true })).toHaveAttribute('aria-current', 'page')
  await expect(rail.locator('[aria-current="page"]')).toHaveCount(1)

  // THE COLUMN IS 132px, WHICH NO UNIT TEST CAN SEE — jsdom applies no
  // stylesheet at all, so `.app-shell-preview .workbench`'s
  // `grid-template-columns` has exactly one possible witness and this is it. The
  // thumbnail well's 44x62 is measured for the same reason: an inline height
  // pdf.js sets from its own 126px thumbnail width would otherwise make the well
  // ~178px tall and nothing in the unit suite would notice.
  const railBox = await rail.boundingBox()
  expect(railBox).not.toBeNull()
  expect(railBox!.width).toBe(132)
  const wellBox = await firstWell.boundingBox()
  expect(wellBox).not.toBeNull()
  expect(wellBox!.width).toBe(44)
  expect(wellBox!.height).toBe(62)

  // The rail is a sibling of the page area, not an overlay on it: the page area
  // starts to the right of where the rail ends.
  const region = page.getByLabel('Preview region')
  const areaBox = await region.boundingBox()
  expect(areaBox).not.toBeNull()
  expect(areaBox!.x).toBeGreaterThanOrEqual(railBox!.x + railBox!.width)

  // ⚠ AND THE COLUMN ASSIGNMENT SURVIVES THE RAIL GOING AWAY, which is the case
  // the story's own final matrix row describes: a failed render has no document
  // to enumerate, so NO rail is in the document at all, and `.workbench` is left
  // with two children in a three-column grid. Under auto-placement the preview
  // region would slide into the 132px track and the inspector into the `1fr` — a
  // visibly broken screen. `.app-shell-preview` assigns all three columns by
  // name so it cannot.
  //
  // The rail is removed here rather than reached through a real render failure:
  // the resulting DOM is exactly the failed-render DOM (workbench > main +
  // inspector), and it is deterministic, whereas manufacturing an engine refusal
  // would need a fixture invented for this test. jsdom applies no stylesheet, so
  // there is no unit-level equivalent of this measurement.
  await page.evaluate(() => { document.querySelector('.page-rail')?.remove() })
  await expect(page.getByLabel('Page thumbnails')).toHaveCount(0)
  const railless = await region.boundingBox()
  expect(railless).not.toBeNull()
  // Still in the second track — it did NOT slide left into the rail's 132px.
  expect(railless!.x).toBe(areaBox!.x)
  expect(railless!.width).toBe(areaBox!.width)
  expect(railless!.x).toBeGreaterThanOrEqual(132)
  // And the inspector is still the last track, at its declared 300px.
  const inspector = await page.getByLabel('Inspector', { exact: true }).boundingBox()
  expect(inspector).not.toBeNull()
  expect(inspector!.width).toBe(300)
  expect(inspector!.x).toBeGreaterThanOrEqual(railless!.x + railless!.width)
})
