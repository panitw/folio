import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// STORY 13.4 — THE BROWSER WITNESS FOR A PREVIEW WITH NO SAMPLE DATA.
//
// COMPILE-CHECKED LOCALLY, EXECUTED IN CI. The story's heavy-test cadence runs
// unit + lint + typecheck on the authoring machine; the Playwright suite builds
// the release, which that cadence forbids locally. CI runs the whole browser
// suite on every push (DW-268, discharged at `adf905a`), so this test executes
// per-commit and is written as real coverage — never as a compile-only
// placeholder.
//
// It is the one witness jsdom cannot give: the REAL Go engine generating the
// stand-in document, the real wasm render, and the real PDF.js rasterizer
// admitting bytes produced from values no author supplied.
const template = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../folio-go/testdata/example/first-pdf.folio'))

test('previews a bound template with no sample data, and claims nothing about production', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const templateChooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await templateChooser).setFiles({ name: 'statement.folio', mimeType: 'application/json', buffer: template })
  await expect(page.locator('.document-name')).toHaveText('statement.folio')

  // STRAIGHT TO PREVIEW. No JSON is loaded, and the template's one text element
  // binds {{customer.name}} — the shape that used to refuse three ways over.
  await page.getByRole('button', { name: 'PREVIEW' }).click()

  // A REAL PDF ARRIVES, AND PDF.js ADMITS IT. The alternation this line used
  // to accept — `Stale historical PDF|Current no-data layout PDF` — is
  // satisfied by the FIRST name, which the viewer carries the moment bytes are
  // handed to it. Admission is what promotes it to `Current`: App.tsx only
  // reaches `'current'` from the viewer's own onPageCount, so waiting for the
  // admitted name is the assertion the comment was claiming to make.
  await expect(page.getByRole('region', { name: /Stale historical PDF/ })).toBeVisible({ timeout: 60_000 })
  await expect(page.getByRole('region', { name: /Current no-data layout PDF, revision \d+/ })).toBeVisible({ timeout: 60_000 })
  await expect(page.getByRole('note', { name: 'No-data preview notice' })).toBeVisible()
  await expect(page.getByText('NO-DATA LAYOUT PREVIEW')).toBeVisible()
  await expect(page.getByText(/Stand-in local digest [a-f0-9]{64}/)).toBeVisible()

  // AND NOTHING ON THE SCREEN CLAIMS PRODUCTION.
  await expect(page.getByText('EXACT LOCAL PRODUCTION PDF')).toHaveCount(0)
  await expect(page.getByRole('region', { name: /Current exact local production PDF/ })).toHaveCount(0)
  await expect(page.getByText(/Historical producer digest/)).toHaveCount(0)
  await expect(page.getByText('Preview unavailable: no sample data loaded')).toHaveCount(0)

  // The control the third gate disabled is reachable and usable, and the
  // export names what it would be writing.
  await page.getByRole('tab', { name: 'INPUTS' }).click()
  await expect(page.getByRole('button', { name: 'Render local PDF' })).toBeEnabled()
  await expect(page.getByRole('button', { name: 'Save no-data PDF' })).toBeVisible()
})
