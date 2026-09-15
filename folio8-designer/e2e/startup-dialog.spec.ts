import { expect, test, type Page } from '@playwright/test'

// STARTUP TEMPLATES, STORY 3 — THE BROWSER WITNESS FOR THE LAUNCH DIALOG.
//
// jsdom can prove the dialog's wiring against a fake engine; only a real build
// can prove the bundled thumbnails decode, that Escape leaves the REAL starter
// at revision 1, and that an example's template and sample render through the
// real wasm engine into an admitted production PDF.

const usableOfflineState = /^(Offline ready|Update available; current release remains usable)$/

const cardNames = ['Blank', 'Invoice', 'Bank Statement', 'Legal Contract', 'Electricity Bill']

async function expectLaunchDialog(page: Page) {
  const dialog = page.getByRole('dialog', { name: 'New template' })
  await expect(dialog).toBeVisible()
  const cards = dialog.getByRole('group', { name: 'Start from' }).getByRole('button')
  await expect(cards).toHaveCount(cardNames.length)
  for (const [index, name] of cardNames.entries()) await expect(cards.nth(index)).toHaveAccessibleName(name)
  const invoice = dialog.getByRole('button', { name: 'Invoice', exact: true })
  await expect(invoice).toHaveAttribute('aria-pressed', 'true')
  await expect(invoice).toBeFocused()
  // FIVE THUMBNAILS: Blank's drawn page and four engine-rendered PNGs, each
  // actually decoded — a broken image is still an <img>.
  await expect(dialog.getByTestId('startup-blank-page')).toBeVisible()
  const images = dialog.locator('img.startup-thumbnail')
  await expect(images).toHaveCount(4)
  await expect.poll(() => images.evaluateAll((nodes) => nodes.every((node) => (node as HTMLImageElement).complete && (node as HTMLImageElement).naturalWidth > 0))).toBe(true)
  return dialog
}

async function expectExampleInPreview(page: Page, name: string) {
  await expect(page.getByRole('dialog', { name: 'New template' })).toHaveCount(0)
  await expect(page.locator('.document-name')).toHaveText(name)
  await expect(page.getByText('Unsaved local changes')).toBeVisible()
  await expect(page.getByLabel('Preview region')).toBeVisible()
  await expect(page.getByRole('region', { name: /Current exact local production PDF, revision \d+/ })).toBeVisible({ timeout: 60_000 })
  await expect(page.getByRole('note', { name: 'No-data preview notice' })).toHaveCount(0)
}

test('launch shows the dialog, and Escape lands on the starter canvas at revision 1', async ({ page }) => {
  await page.goto('/')
  const dialog = await expectLaunchDialog(page)
  await expect(dialog.getByRole('button', { name: 'Open example' })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(dialog).toHaveCount(0)
  await expect(page.getByLabel('Canvas region')).toBeVisible()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await expect(page.locator('.document-name')).toHaveText('Untitled template')
})

test('opening Invoice lands in Preview with its sample loaded', async ({ page }) => {
  await page.goto('/')
  const dialog = await expectLaunchDialog(page)
  await expect(dialog.getByRole('status')).toContainText('Invoice opens in Preview with')
  await expect(dialog.getByRole('status')).toContainText('invoice.sample.json')
  await dialog.getByRole('button', { name: 'Open example' }).click()
  await expectExampleInPreview(page, 'Invoice')
  // The sample tree is the one Load sample JSON would have installed.
  await page.getByRole('tab', { name: 'DATA' }).click()
  await expect(page.getByRole('tree', { name: 'Sample data paths' })).toBeVisible()
})

test('offline after first load, the dialog and its thumbnails appear and an example opens in Preview', async ({ page, context }) => {
  await page.goto('/')
  await page.reload() // the first installation must activate before it can control a reload
  await expect(page.getByTestId('offline-status')).toHaveText(usableOfflineState)
  await expectLaunchDialog(page)
  await context.setOffline(true)
  await page.reload()
  const dialog = await expectLaunchDialog(page)
  await dialog.getByRole('button', { name: 'Bank Statement', exact: true }).click()
  await expect(dialog.getByRole('button', { name: 'Bank Statement', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await dialog.getByRole('button', { name: 'Open example' }).click()
  await expectExampleInPreview(page, 'Bank Statement')
})
