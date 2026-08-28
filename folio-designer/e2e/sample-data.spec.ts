import { expect, test } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })
})

test('loads local sample JSON into the docked navigable discovery panel without opening a destination', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByLabel('Data panel')).toBeVisible()
  await expect(page.getByText('Binding unavailable: no sample data loaded.')).toBeVisible()
  await expect(page.getByLabel('Canvas region')).toBeVisible()
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Load sample JSON' }).click()
  await (await chooser).setFiles({ name: 'sample.json', mimeType: 'application/json', buffer: Buffer.from('{"customer":{"name":"Ada"},"items":[{"sku":"A-1"}]}') })
  await expect(page.getByRole('tree', { name: 'Sample data paths' })).toContainText('items[]')
  await expect(page.getByRole('button', { name: 'Replace sample JSON' })).toBeVisible()
  const tree = page.getByRole('tree', { name: 'Sample data paths' })
  // The tree is a projection that may start with a synthetic root; selecting
  // the visible scalar is stable across that implementation detail.
  await tree.getByRole('treeitem').filter({ hasText: /^customer/ }).click()
  await tree.getByRole('treeitem').filter({ hasText: /^name/ }).click()
  await expect(tree.getByRole('treeitem').filter({ hasText: /^name/ })).toBeFocused()

  const replacement = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Replace sample JSON' }).click()
  await (await replacement).setFiles({ name: 'invalid.json', mimeType: 'application/json', buffer: Buffer.from('{"customer":}') })
  await expect(page.getByRole('alert')).toContainText('one valid JSON document')
  await expect(tree).toContainText('Ada')
})
