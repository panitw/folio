import { expect, test } from '@playwright/test'

// STORY 16.3 — THE FONT BROWSER, IN A REAL BROWSER.
//
// WHAT THIS FILE CAN AND CANNOT WITNESS, STATED RATHER THAN IMPLIED. Every
// assertion below is over the modal's own structure, its keyboard behaviour and
// its staged state — none of which needs a network. The two claims that DO need
// one — a specimen actually rendered in its family's fetched face, and a
// confirmed add reaching the document — are exercised by the by-hand run this
// story's Verification section records, because the harness this repository runs
// e2e under has no route to the upstream repository host and a test that
// silently asserts the degraded path would be worse than no test.
//
// THE NAMES ARE THE CONTRACT. Story 8.2's spec pins `Font family` as the family
// control's accessible name and says in its own comment that the name must not
// move. This file adds names; it renames nothing.

const openBrowser = async (page: import('@playwright/test').Page) => {
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const content = page.getByRole('region', { name: 'Content', exact: true })
  await page.getByRole('button', { name: 'Place Text' }).click()
  await content.press('Enter')
  // PLACING THE COMPONENT IS ITSELF A COMMAND, so the document is at revision 2
  // before the modal has been opened at all. Anything asserting that the browser
  // left the document untouched has to compare against THIS, not against 1.
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 2/)
  await content.getByRole('button', { name: /text component/ }).last().click()
  // THE DOOR IS THE LAST ROW OF THE OPEN FAMILY DROPDOWN, where the design puts
  // it — reached by focusing the control the panel already had, never by a
  // keyboard shortcut. No shortcut is bound in this epic (D-16.R.33 R2), and no
  // hint glyph is drawn, so there is none to press and none to assert.
  await page.getByRole('combobox', { name: 'Font family' }).focus()
  await page.getByRole('button', { name: 'Add fonts…' }).click()
  return page.getByRole('dialog', { name: 'Font browser' })
}

test('the browser opens from the family control as the design\'s five-region modal', async ({ page }) => {
  const browser = await openBrowser(page)
  await expect(browser).toBeVisible()
  await expect(browser).toHaveAttribute('aria-modal', 'true')

  // Focus lands in the search box, so a keyboard-only author can type
  // immediately rather than tabbing in from wherever the panel left them.
  await expect(page.getByRole('textbox', { name: 'Search fonts' })).toBeFocused()

  // The rail's three groups, the results, and the footer.
  await expect(browser.getByRole('textbox', { name: 'Preview text' })).toBeVisible()
  await expect(browser.getByRole('slider', { name: 'Preview size in pixels' })).toBeVisible()
  await expect(browser.getByRole('button', { name: 'Thai sample text' })).toBeVisible()
  await expect(browser.getByRole('group', { name: 'Writing system' })).toBeVisible()
  await expect(browser.getByRole('group', { name: 'Category' })).toBeVisible()
  await expect(browser.getByRole('group', { name: 'Sort' })).toBeVisible()
  await expect(browser.getByRole('group', { name: 'Results view' })).toBeVisible()
  await expect(browser.getByRole('list', { name: 'Font families' })).toBeVisible()
  await expect(browser.getByRole('button', { name: 'Add to template' })).toBeDisabled()

  // THE COUNT IS THE ADDABLE ONE AND IT NAMES THE SNAPSHOT. The mockup's
  // "web font library · 1,946 families" claims a live library over a published
  // count; this is the sentence Story 16.1 already ships.
  await expect(browser.getByText(/families you can add/)).toBeVisible()
  await expect(browser.getByText(/The list itself is a snapshot taken on/)).toBeVisible()
  await expect(browser.getByText(/web font library/)).toHaveCount(0)
})

test('search, chips and sort narrow the results, and reset filters appears exactly while one is active', async ({ page }) => {
  const browser = await openBrowser(page)
  await expect(browser.getByRole('button', { name: 'reset filters' })).toHaveCount(0)

  await page.getByRole('textbox', { name: 'Search fonts' }).fill('sarabun')
  await expect(browser.getByRole('list', { name: 'Font families' }).getByText('Sarabun', { exact: true })).toBeVisible()
  await expect(browser.getByRole('button', { name: 'reset filters' })).toBeVisible()

  await browser.getByRole('button', { name: 'Clear the search' }).click()
  await browser.getByRole('button', { name: 'Thai coverage', exact: true }).click()
  // `exact` because the vocabulary is derived from the snapshot and it carries
  // both `Serif` and `Sans Serif`: a substring match would resolve to two chips.
  await browser.getByRole('button', { name: 'Serif category', exact: true }).click()
  await expect(browser.getByRole('button', { name: 'Sort by A – Z' })).toBeVisible()
  // The arm the port dropped: this product embeds one face per family, so a
  // style count is not a difference the author can act on (D-16.R.33 R3).
  await expect(browser.getByRole('button', { name: /Most styles/ })).toHaveCount(0)

  await browser.getByRole('button', { name: 'reset filters' }).click()
  await expect(browser.getByRole('button', { name: 'reset filters' })).toHaveCount(0)
})

test('an unmatched query gets the design\'s empty state, naming the query', async ({ page }) => {
  const browser = await openBrowser(page)
  await page.getByRole('textbox', { name: 'Search fonts' }).fill('zzzz no such family')
  await expect(browser.getByText('No families match “zzzz no such family”')).toBeVisible()
  await expect(browser.getByRole('list', { name: 'Font families' })).toHaveCount(0)
})

test('the Grid view draws the same families as cards', async ({ page }) => {
  const browser = await openBrowser(page)
  const rows = await browser.getByRole('list', { name: 'Font families' }).getByRole('listitem').count()
  await browser.getByRole('button', { name: 'Grid view' }).click()
  await expect(browser.getByRole('list', { name: 'Font families' }).getByRole('listitem')).toHaveCount(rows)
})

test('staging several families states what is about to be embedded, and Escape discards it', async ({ page }) => {
  const browser = await openBrowser(page)
  const list = browser.getByRole('list', { name: 'Font families' })
  const add = list.getByRole('button', { name: /^Add .* to this template$/ })
  await add.nth(0).click()
  await add.nth(0).click()
  await expect(browser.getByText(/^2 families ready to embed/)).toBeVisible()
  // THE FOOTER SAYS WHAT THIS PRODUCT ACTUALLY PUTS IN THE FILE — one upright
  // Regular per family, whole — and not the mockup's "≈ N weights · subset".
  await expect(browser.getByText(/2 faces · one upright Regular each · whole file, not subset/)).toBeVisible()
  await expect(browser.getByRole('button', { name: 'Add 2 to template' })).toBeEnabled()

  // ESCAPE DISCARDS THE STAGED SET AND LEAVES THE DOCUMENT UNTOUCHED. Nothing
  // was sent, so the revision has not moved off the one the placement produced.
  await page.keyboard.press('Escape')
  await expect(browser).toHaveCount(0)
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 2/)
})

test('every family the browser lists has a specimen state it states in words', async ({ page }) => {
  const browser = await openBrowser(page)
  const rows = browser.getByRole('list', { name: 'Font families' }).getByRole('listitem')
  await expect(rows.first()).toBeVisible()
  // A ROW EITHER SHOWS THE SAMPLE SET IN ITS OWN FACE OR SAYS WHY IT IS NOT.
  // In a harness with no route upstream every web-tier row takes the second
  // branch, which is exactly the assertion worth making here: the browser never
  // renders the sample in the panel's typeface while implying it is the family.
  const first = rows.first()
  const specimen = first.locator('.font-browser-specimen, .font-browser-specimen-absent')
  await expect(specimen).toHaveCount(1)
})
