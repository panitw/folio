import { expect, test } from '@playwright/test'

for (const snap of [false, true]) {
  test(`a pointer-placed table fills content with one blank editable column (snap=${snap})`, async ({ page }, testInfo) => {
    await page.goto('/')
    await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
    if (!snap) await page.getByRole('button', { name: /^Snap on/ }).click()
    const content = page.getByRole('region', { name: 'Content', exact: true })
    const band = await content.boundingBox()
    expect(band).not.toBeNull()
    await page.getByRole('button', { name: 'Place Table' }).click()
    // A real pointer close to the right edge used to place a zero-width box
    // there, and provisional free-box snapping could also refuse this drop.
    await content.click({ position: { x: band!.width - 5, y: 79 } })
    const table = content.getByRole('button', { name: /table component/ })
    await expect(table).toBeVisible()
    await expect(table).toHaveClass(/canvas-component-selected/)
    const placed = await table.boundingBox()
    expect(placed).not.toBeNull()
    expect(placed!.x).toBeCloseTo(band!.x, 0)
    expect(placed!.width).toBeCloseTo(band!.width, 0)
    const pixelsPerPoint = band!.width / 523.276
    const intendedY = snap ? Math.round(79 / pixelsPerPoint / 6) * 6 * pixelsPerPoint : 79
    expect(placed!.y - band!.y).toBeCloseTo(intendedY, 0)
    expect(placed!.height).toBeCloseTo(24 * pixelsPerPoint, 1)
    const chip = await table.locator('.canvas-table-chip').boundingBox()
    const selection = await page.locator('.canvas-selection-chrome').boundingBox()
    expect(chip).not.toBeNull()
    expect(selection).not.toBeNull()
    expect(chip!.y).toBeCloseTo(placed!.y, 1)
    expect(chip!.height).toBeCloseTo(placed!.height, 1)
    expect(selection!.height).toBeCloseTo(chip!.height, 1)
    const tableID = await table.getAttribute('data-component-id')
    const column = table.locator('.canvas-table-heading[data-column-id]')
    await expect(column).toHaveCount(1)
    const columnID = await column.getAttribute('data-column-id')
    expect(columnID).toBeTruthy()
    expect(columnID).not.toBe(tableID)
    await page.screenshot({ path: testInfo.outputPath('table-full-width.png'), fullPage: true })

    // Creation and its column are one engine history entry.
    await page.getByRole('button', { name: 'Undo', exact: true }).click()
    await expect(table).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Undo', exact: true })).toBeDisabled()
    await page.getByRole('button', { name: 'Redo', exact: true }).click()
    await expect(table).toHaveAttribute('data-component-id', tableID!)
    await expect(column).toHaveAttribute('data-column-id', columnID!)
    await table.click()
    await page.getByRole('button', { name: 'Configure columns' }).click()
    const dialog = page.getByRole('dialog', { name: 'Table Editor' })
    const width = dialog.getByRole('spinbutton', { name: 'Width for column 1 in points' })
    const header = dialog.getByRole('textbox', { name: 'Header for column 1' })
    await expect(width).toHaveValue('523.276')
    await expect(header).toHaveValue('')
    await expect(dialog.getByRole('grid', { name: 'Table columns' })).toHaveAttribute('aria-rowcount', '2')
    await width.fill('120')
    await width.press('Tab')
    await expect(dialog.getByRole('status', { name: 'Width budget' })).toContainText('Σ 120.0')
    await header.fill('Description')
    await header.press('Tab')
    await expect(header).toHaveValue('Description')
    await dialog.getByRole('button', { name: 'Done', exact: true }).click()
    await expect(column).toHaveText('Description')
    await expect.poll(async () => (await table.boundingBox())!.width).toBeCloseTo(120 * pixelsPerPoint, 0)
  })
}

test('the table bar and selection share the authored height at several zoom levels', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Place Table' }).click()
  await page.getByRole('region', { name: 'Content', exact: true }).click({ position: { x: 120, y: 96 } })
  const table = page.getByRole('button', { name: /table component/ })
  await expect(table).toBeVisible()
  let zoom = 10
  for (const height of [24, 12, 36]) {
    await page.getByRole('button', { name: 'Configure columns' }).click()
    const dialog = page.getByRole('dialog', { name: 'Table Editor' })
    const headerHeight = dialog.getByRole('spinbutton', { name: 'Header height in points' })
    if (height === 24) await expect(headerHeight).toHaveValue('24')
    else {
      await headerHeight.fill(String(height))
      await headerHeight.press('Tab')
      await expect(headerHeight).toHaveValue(String(height))
    }
    await dialog.getByRole('button', { name: 'Done', exact: true }).click()
    for (const target of [10, 5, 11, 20]) {
      while (zoom !== target) {
        await page.getByRole('button', { name: zoom < target ? 'Zoom in' : 'Zoom out', exact: true }).click()
        zoom += zoom < target ? 1 : -1
      }
      await expect(page.getByLabel('Canvas zoom')).toHaveText(`${target * 10}%`)
      await table.scrollIntoViewIfNeeded()
      await expect.poll(async () => (await table.boundingBox())!.height).toBeCloseTo(height * target / 10, 1)
      const box = (await table.boundingBox())!
      const chip = (await table.locator('.canvas-table-chip').boundingBox())!
      const selection = (await page.locator('.canvas-selection-chrome').boundingBox())!
      expect(chip.y).toBeCloseTo(box.y, 1)
      expect(chip.height).toBeCloseTo(box.height, 1)
      expect(selection.y).toBeCloseTo(chip.y, 1)
      expect(selection.height).toBeCloseTo(chip.height, 1)
      for (const selector of ['.canvas-table-collection', '.canvas-table-count', '.canvas-table-chip-icon']) {
        const content = (await table.locator(selector).boundingBox())!
        expect(content.y).toBeGreaterThanOrEqual(chip.y - 0.1)
        expect(content.y + content.height).toBeLessThanOrEqual(chip.y + chip.height + 0.1)
      }
    }
  }
})
