import { expect, test } from '@playwright/test'

test('the initial shell exposes desktop landmarks and honest local-file controls', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByLabel('Document bar')).toBeVisible()
  await expect(page.getByLabel('Component palette')).toBeVisible()
  await expect(page.getByLabel('Canvas region')).toBeVisible()
  await expect(page.getByLabel('Properties panel')).toBeVisible()

  await expect(page.getByRole('button', { name: 'Open local template' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Save local template' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Save As' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Start blank' })).toBeVisible()
  await expect(page.getByText('Unsaved local changes')).toBeVisible()
  await expect(page.getByText('PREVIEW · later')).toBeVisible()
})
