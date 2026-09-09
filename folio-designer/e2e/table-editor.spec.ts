import { expect, test } from '@playwright/test'

// WHEN THIS FILE ACTUALLY RUNS. `npm run test:e2e` is its own CI job — it
// executes on every push to main and on every pull request, with no
// `continue-on-error` anywhere in the workflow — so this spec runs in a real
// Chromium there. Locally it is COMPILED at story cadence
// (`npm run test:e2e:compile`, which is `tsc --noEmit` over the spec files):
// that proves the spec still typechecks, which is not the same claim as the
// spec still passing, and DW-268 records the two days a broken roundtrip spec
// survived on exactly that difference. The older note here said the executable
// pass "remains Epic 6 boundary evidence under D-000.4", which stopped being
// true when the browser job was added.
//
// STORY 14.7 REWROTE THE ROVING WALK FOR THE SIX-COLUMN LATTICE. The matrix
// draws six columns — `#`, HEADER LABEL, BOUND FIELD · row scope, WIDTH, ALIGN,
// FOOTER AGGREGATE — while the KEYBOARD lattice behind them is wider than six,
// because the alignment control is three segments and the row's reorder/remove
// affordances are three more controls inside the `#` cell. Every control the
// eleven-column matrix could reach by arrow keys is still reachable by arrow
// keys (UX-DR25); what changed is which cell each one sits in.
test('table editor is a named keyboard-operable matrix', async ({ page }) => {
  await page.goto('/')
  await page.getByRole('button', { name: 'Place Table' }).click()
	await page.getByRole('region', { name: 'Content', exact: true }).click({ position: { x: 24, y: 24 } })
  await page.getByRole('button', { name: /table component/ }).click()
  await page.getByRole('button', { name: 'Configure columns' }).click()
  await expect(page.getByRole('button', { name: 'Add column' })).toBeVisible()
  await page.getByRole('button', { name: 'Add column' }).click()
  const grid = page.getByRole('grid', { name: 'Table columns' })
  await expect(grid).toBeVisible()
	await expect(grid).toHaveAttribute('aria-rowcount', '2')
	await expect(grid).toHaveAttribute('aria-colcount', '6')
	// The retired four are gone as COLUMNS and present as row affordances.
	await expect(page.getByRole('columnheader')).toHaveText(['#', 'HEADER LABEL', 'BOUND FIELD · row scope', 'WIDTH', 'ALIGN', 'FOOTER AGGREGATE'])
	await expect(page.getByRole('button', { name: 'Add column after column 1' })).toHaveCount(0)
	// The collection and the row alias are still edited here, and nowhere else.
	await expect(page.getByRole('combobox', { name: 'Root collection' })).toBeVisible()
	await expect(page.getByRole('textbox', { name: 'Row alias' })).toBeVisible()
	// The two read-outs this story added, stated rather than implied.
	await expect(page.getByRole('status', { name: 'Table scope' })).toContainText('band: content')
	await expect(page.getByRole('status', { name: 'Width budget' })).toContainText('Σ')
	await expect(page.getByRole('status', { name: 'Column summary' })).toContainText('1 column · 0 aggregates')
  const header = page.getByRole('textbox', { name: 'Header for column 1' })
  await header.focus()
	await page.keyboard.press('ArrowRight')
	await expect(page.getByRole('combobox', { name: 'Row field for column 1' })).toBeFocused()
	await page.keyboard.press('ArrowRight')
	await expect(page.getByRole('spinbutton', { name: 'Width for column 1 in points' })).toBeFocused()
	// THE ALIGNMENT CONTROL IS THREE REACHABLE SEGMENTS, not one select: each
	// takes its own lattice position, so no segment becomes unreachable.
	await page.keyboard.press('ArrowRight')
	await expect(page.getByRole('button', { name: 'Align left for column 1' })).toBeFocused()
	await page.keyboard.press('ArrowRight')
	await expect(page.getByRole('button', { name: 'Align center for column 1' })).toBeFocused()
	await page.keyboard.press('ArrowRight')
	await expect(page.getByRole('button', { name: 'Align right for column 1' })).toBeFocused()
	// End reaches the row's LAST ENABLED control. This column aggregates
	// nothing, so its source and its format do not exist and End stops at the
	// aggregate rather than landing on either hole.
	await page.keyboard.press('End')
	await expect(page.getByRole('combobox', { name: 'Footer aggregate for column 1' })).toBeFocused()
	await expect(page.getByRole('textbox', { name: 'Footer source for column 1' })).toHaveCount(0)
	await expect(page.getByRole('textbox', { name: 'Footer format for column 1' })).toHaveCount(0)
	// Home reaches the row's FIRST ENABLED control. A single column can move
	// neither earlier nor later, so both reorder affordances are disabled and
	// Home declines to land on either.
	await page.keyboard.press('Home')
	await expect(page.getByRole('button', { name: 'Remove column 1' })).toBeFocused()
	await expect(page.getByRole('button', { name: 'Move column 1 earlier' })).toBeDisabled()
	await expect(page.getByRole('button', { name: 'Move column 1 later' })).toBeDisabled()
	// `Close Table Editor` moved out of the heading and into the footer bar.
	await expect(page.getByRole('button', { name: 'Close Table Editor' })).toBeVisible()
	await page.keyboard.press('Escape')
	await expect(page.getByRole('dialog', { name: 'Table Editor' })).toHaveCount(0)
})
