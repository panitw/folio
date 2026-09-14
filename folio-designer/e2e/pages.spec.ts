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
