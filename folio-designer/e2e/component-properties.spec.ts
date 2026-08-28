import { expect, test } from '@playwright/test'

// Compile-covered at the Epic 5 cadence. Browser execution intentionally
// remains deferred by D-000.4; Go is the property-validation authority.
test('a selected component exposes committed properties and a mixed selection omits non-shared size', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Place Text' }).click()
  await page.getByLabel('Content').press('Enter')
  const text = page.getByLabel(/text component e/)
  await text.click()
  await expect(page.getByRole('textbox', { name: 'X (pt)' })).toBeVisible()
  await page.getByRole('textbox', { name: 'X (pt)' }).fill('12')
  await page.getByRole('textbox', { name: 'X (pt)' }).press('Enter')

  await page.getByRole('button', { name: 'Place Table' }).click()
  await page.getByLabel('Content').press('Enter')
  const table = page.getByLabel(/table component e/)
  await table.click({ modifiers: ['Shift'] })
  await expect(page.getByRole('textbox', { name: 'Width (pt)' })).toHaveCount(0)
  await expect(page.getByText(/Table size, binding, columns/)).toBeVisible()
})
