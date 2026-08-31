import { expect, test } from '@playwright/test'

// THIS FILE IS COMPILED, NEVER EXECUTED. `npm run test:e2e:compile` is
// `tsc --noEmit` over this directory; `npm run test:e2e` (Playwright) appears
// in no workflow and browser e2e remains deferred by D-000.4. So what follows
// is a typed witness that the roles, names and flows below EXIST in the
// source — not evidence that any of them work in a browser. Nothing here may
// be reported as verified behaviour.
// Compile-covered at the Epic 5 cadence; Go is the property-validation
// authority.
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

// Story 8.2. The chain editor is revealed from a third button on the family
// control and renders INSIDE the same typography section — no dialog, no
// separate mode (AC1). Every control it draws is addressed here by the
// accessible name the panel gives it.
test('the font chain editor is revealed in the typography section and exposes every chain control', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const content = page.getByRole('region', { name: 'Content', exact: true })
  await page.getByRole('button', { name: 'Place Text' }).click()
  await content.press('Enter')
  await content.getByRole('button', { name: /text component/ }).last().click()

  // THE NAME THAT MUST NOT MOVE. browser-native-roundtrip.spec.ts addresses
  // the family control by exactly this accessible name, and it is the
  // repository's only cross-boundary authoring witness. The chain-editor
  // affordance is a THIRD button beside it, not a rename of it.
  const family = page.getByRole('combobox', { name: 'Font family' })
  await expect(family).toBeVisible()

  await page.getByRole('button', { name: 'Edit font chains' }).click()
  await expect(page.getByRole('dialog')).toHaveCount(0)
  const chains = page.getByRole('group', { name: 'Font chains' })
  await expect(chains).toBeVisible()
  await expect(family).toBeVisible()

  // Rename: the name control, then its action.
  const name = chains.getByRole('textbox', { name: 'Font chain 1 name' })
  await name.fill('text')
  await page.getByRole('button', { name: 'Rename font chain 1' }).click()

  // An ordered entry list with per-entry move and remove controls.
  await expect(chains.getByRole('list', { name: 'Entries of font chain 1' })).toBeVisible()
  await page.getByRole('button', { name: 'Move entry 1 of font chain 1 later' }).click()
  await page.getByRole('button', { name: 'Move entry 2 of font chain 1 earlier' }).click()
  await page.getByRole('button', { name: 'Remove entry 1 of font chain 1' }).click()

  // Add an entry, then a whole chain.
  await chains.getByRole('textbox', { name: 'New entry for font chain 1' }).fill('Noto Sans SC')
  await page.getByRole('button', { name: 'Add entry to font chain 1' }).click()
  await chains.getByRole('textbox', { name: 'New font chain name' }).fill('display')
  await chains.getByRole('textbox', { name: 'First entry for the new font chain' }).fill('Noto Sans')
  await page.getByRole('button', { name: 'Add font chain' }).click()

  // Delete, and the disclosure closing again.
  await page.getByRole('button', { name: 'Delete font chain 1' }).click()
  await page.getByRole('button', { name: 'Hide font chains' }).click()
  await expect(page.getByRole('group', { name: 'Font chains' })).toHaveCount(0)
})
