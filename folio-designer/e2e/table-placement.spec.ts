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
    expect(placed!.height).toBeGreaterThan(0)
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
