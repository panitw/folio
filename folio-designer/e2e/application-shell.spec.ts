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

test('the real worker projects page bands and accepts local page setup without changing transient controls', async ({ page }) => {
  await page.goto('/')
  const report = page.getByLabel('Report page with Page Header, Content, and Page Footer')
  await expect(report).toBeVisible()
  await expect(page.getByLabel('Page Header')).toBeVisible()
  await expect(page.getByLabel('Content')).toBeVisible()
  await expect(page.getByLabel('Page Footer')).toBeVisible()
  await page.getByRole('button', { name: 'Zoom in' }).click()
  await expect(page.getByLabel('Canvas zoom')).toHaveText('110%')
  await page.getByRole('button', { name: 'Grid on' }).click()
  await expect(page.getByRole('button', { name: 'Grid off' })).toHaveAttribute('aria-pressed', 'false')
  await page.getByRole('button', { name: 'Snap on' }).click()
  await expect(page.getByRole('button', { name: 'Snap off' })).toHaveAttribute('aria-pressed', 'false')
  await page.getByRole('textbox', { name: 'Top margin (pt)' }).fill('37')
  await page.getByRole('button', { name: 'Apply page setup' }).click()
  await expect(page.getByText('Unsaved local changes')).toBeVisible()
})

test('canvas keeps a keyboard-visible scroll target at a narrow viewport', async ({ page }) => {
  await page.setViewportSize({ width: 360, height: 700 })
  await page.goto('/')
  const canvas = page.getByLabel('Canvas region')
  await expect(canvas).toBeVisible()
  await canvas.focus()
  await expect(canvas).toBeFocused()
  await expect(page.getByRole('button', { name: 'Zoom in' })).toBeVisible()
})
