import { expect, test } from '@playwright/test'

// Compile-covered at the Epic 5 cadence. Browser execution intentionally
// remains deferred by D-000.4; Go is the property-validation authority.
test('a selected component exposes committed properties and a mixed selection omits non-shared size', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const content = page.getByRole('region', { name: 'Content', exact: true })
  await page.getByRole('button', { name: 'Place Text' }).click()
  await content.press('Enter')
  const text = content.getByRole('button', { name: /text component/ }).last()
  await text.click()
  await expect(page.getByRole('textbox', { name: 'X (pt)' })).toBeVisible()
  await page.getByRole('textbox', { name: 'X (pt)' }).fill('12')
  await page.getByRole('textbox', { name: 'X (pt)' }).press('Enter')

  await page.getByRole('button', { name: 'Place Table' }).click()
  await content.press('Enter')
  const table = content.getByRole('button', { name: /table component/ }).last()
  await text.click()
  await table.click({ modifiers: ['Shift'] })
  await expect(page.getByRole('textbox', { name: 'Width (pt)', exact: true })).toHaveCount(0)
  await expect(page.getByText(/Table size and binding are not editable here/)).toBeVisible()
})
