import { expect, test } from '@playwright/test'

test('the initial shell exposes its desktop landmarks and honest deferred controls', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByLabel('Document bar')).toBeVisible()
  await expect(page.getByLabel('Component palette')).toBeVisible()
  await expect(page.getByLabel('Canvas region')).toBeVisible()
  await expect(page.getByLabel('Properties panel')).toBeVisible()

  await expect(page.getByRole('button', { name: 'Open local template unavailable until local files are added' })).toBeDisabled()
  await expect(page.getByText('PREVIEW · later')).toBeVisible()
})
