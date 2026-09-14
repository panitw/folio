import { expect, test, type Page } from '@playwright/test'

// SPEC-multi-pages story 2 — PAGES ON THE CANVAS, THROUGH THE REAL GO WORKER.
//
// App.test.tsx proves the commands and the selection state against a fake
// engine; this file proves the whole loop: Add page twice from a blank
// document, delete the middle page after confirming, undo it, clear page 2's
// Page Break, and save bytes that state `"pageBreak": false`.

const revision = (page: Page) => page.getByTestId('engine-snapshot')
const tools = (page: Page) => page.getByLabel('Canvas controls')
const labels = (page: Page) => page.locator('.page-label')

test('adds pages, deletes one after confirming, undoes it, and saves Page Break off', async ({ page }) => {
  await page.addInitScript(() => {
    const handle = {
      name: 'pages.folio',
      getFile: async () => new File([], 'pages.folio', { type: 'application/json' }),
      createWritable: async () => ({ write: async (written: ArrayBuffer) => { (window as typeof window & { __folioWrites?: number[][] }).__folioWrites = [Array.from(new Uint8Array(written))] }, close: async () => undefined }),
    }
    Object.assign(window, { showSaveFilePicker: async () => handle })
  })
  await page.goto('/')
  await expect(revision(page)).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await page.getByRole('button', { name: 'Start blank' }).click()
  await expect(labels(page)).toHaveText(['Page 1'])
  await expect(tools(page).getByRole('button', { name: 'Delete page' })).toBeDisabled()

  // ADD TWICE. Each new page arrives selected, and the next goes after it.
  await tools(page).getByRole('button', { name: 'Add page' }).click()
  await expect(labels(page)).toHaveText(['Page 1', 'Page 2'])
  await expect(page.getByRole('button', { name: 'Page 2' })).toHaveAttribute('aria-pressed', 'true')
  await tools(page).getByRole('button', { name: 'Add page' }).click()
  await expect(labels(page)).toHaveText(['Page 1', 'Page 2', 'Page 3'])

  // DELETE PAGE 2, CANCELLED FIRST. Escape changes nothing.
  await page.getByRole('button', { name: 'Page 2' }).click()
  const beforeDelete = (await revision(page).innerText()).trim()
  await tools(page).getByRole('button', { name: 'Delete page' }).click()
  const dialog = page.getByRole('dialog', { name: 'Delete page 2?' })
  await expect(dialog).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(dialog).toHaveCount(0)
  expect((await revision(page).innerText()).trim()).toBe(beforeDelete)

  // The Delete key never deletes a page.
  await page.getByRole('button', { name: 'Page 2' }).click()
  await page.keyboard.press('Delete')
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await expect(labels(page)).toHaveCount(3)

  // CONFIRMED, then UNDONE: one entry each way.
  await tools(page).getByRole('button', { name: 'Delete page' }).click()
  await page.getByRole('dialog', { name: 'Delete page 2?' }).getByRole('button', { name: 'Delete page' }).click()
  await expect(labels(page)).toHaveText(['Page 1', 'Page 2'])
  await page.keyboard.press('ControlOrMeta+z')
  await expect(labels(page)).toHaveText(['Page 1', 'Page 2', 'Page 3'])

  // PAGE BREAK. Disabled on page 1; cleared on page 2.
  await page.getByRole('button', { name: 'Page 1' }).click()
  await expect(page.getByRole('checkbox', { name: 'Page Break' })).toBeDisabled()
  await page.getByRole('button', { name: 'Page 2' }).click()
  const pageBreak = page.getByRole('checkbox', { name: 'Page Break' })
  await expect(pageBreak).toBeChecked()
  await pageBreak.click()
  await expect(pageBreak).not.toBeChecked()

  // SAVE. The written bytes state the setting.
  await page.getByRole('button', { name: 'Save As' }).click()
  await expect.poll(() => page.evaluate(() => (window as typeof window & { __folioWrites?: number[][] }).__folioWrites?.[0]?.length ?? 0)).toBeGreaterThan(0)
  const saved = await page.evaluate(() => new TextDecoder().decode(new Uint8Array((window as typeof window & { __folioWrites?: number[][] }).__folioWrites![0]!)))
  expect(saved).toContain('"pageBreak": false')
  expect(JSON.parse(saved).pages).toHaveLength(3)
})

// SPEC-multi-pages story 3 — ELEMENTS ON A SPECIFIC PAGE, THROUGH THE REAL GO
// WORKER. Place Text on page 2, place Text on page 1 and drag it onto page 2,
// undo (it is back on page 1), redo, and save bytes listing both under pages[1].
test('places on page 2, drags a page-1 element onto page 2, undoes, and saves it under pages[1]', async ({ page }) => {
  await page.addInitScript(() => {
    const handle = {
      name: 'elements.folio',
      getFile: async () => new File([], 'elements.folio', { type: 'application/json' }),
      createWritable: async () => ({ write: async (written: ArrayBuffer) => { (window as typeof window & { __folioWrites?: number[][] }).__folioWrites = [Array.from(new Uint8Array(written))] }, close: async () => undefined }),
    }
    Object.assign(window, { showSaveFilePicker: async () => handle })
  })
  await page.goto('/')
  await expect(revision(page)).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await page.getByRole('button', { name: 'Start blank' }).click()
  await tools(page).getByRole('button', { name: 'Add page' }).click()
  await expect(labels(page)).toHaveText(['Page 1', 'Page 2'])
  const sheets = page.locator('.page-surface')
  const onSheet = (index: number) => sheets.nth(index).locator('.canvas-component:not(.canvas-component-echo)')

  // PLACE on page 2 from the keyboard: it lands on page 2's sheet.
  await page.getByRole('button', { name: 'Place Text' }).click()
  await page.getByLabel('Content on page 2 of 2').press('Enter')
  await expect(onSheet(1)).toHaveCount(1)
  await expect(onSheet(0)).toHaveCount(0)

  // PLACE on page 1, then DRAG it onto page 2's content band.
  await page.getByRole('button', { name: 'Place Text' }).click()
  await page.getByLabel('Content on page 1 of 2').press('Enter')
  await expect(onSheet(0)).toHaveCount(1)
  const source = onSheet(0)
  const from = await source.boundingBox()
  const target = await page.getByLabel('Content on page 2 of 2').boundingBox()
  if (!from || !target) throw new Error('canvas geometry is not visible')
  await page.mouse.move(from.x + 4, from.y + 4)
  await page.mouse.down()
  await page.mouse.move(from.x + 4, (from.y + target.y + 200) / 2, { steps: 5 })
  await page.mouse.move(from.x + 4, target.y + 200, { steps: 5 })
  // The preview follows the pointer onto page 2's sheet before release.
  await expect(onSheet(1)).toHaveCount(2)
  await page.mouse.up()
  await expect(onSheet(0)).toHaveCount(0)
  await expect(onSheet(1)).toHaveCount(2)

  // ONE UNDO puts it back on page 1; redo moves it again.
  await page.keyboard.press('ControlOrMeta+z')
  await expect(onSheet(0)).toHaveCount(1)
  await expect(onSheet(1)).toHaveCount(1)
  await page.keyboard.press('ControlOrMeta+Shift+z')
  await expect(onSheet(1)).toHaveCount(2)

  // SAVE. Both elements are listed under pages[1].
  await page.getByRole('button', { name: 'Save As' }).click()
  await expect.poll(() => page.evaluate(() => (window as typeof window & { __folioWrites?: number[][] }).__folioWrites?.[0]?.length ?? 0)).toBeGreaterThan(0)
  const saved = JSON.parse(await page.evaluate(() => new TextDecoder().decode(new Uint8Array((window as typeof window & { __folioWrites?: number[][] }).__folioWrites![0]!))))
  expect(saved.pages[0].elements).toHaveLength(0)
  expect(saved.pages[1].elements).toHaveLength(2)
})
