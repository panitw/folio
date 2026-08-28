import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Compiled source for the Epic 6 boundary run. The fixture is altered only as
// picker input; the browser never reads its template structure.
const source = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../folio-go/testdata/example/first-pdf.folio'), 'utf8')
const parameterTemplate = Buffer.from(source.replace('{{customer.name}}', '{{params.reportDate}}'))

test('discovers reportDate from Go, keeps an absent value as a located engine failure, and makes supplied input stale', async ({ page }) => {
  await page.goto('/')
  const templateChooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await templateChooser).setFiles({ name: 'parameter-preview.folio', mimeType: 'application/json', buffer: parameterTemplate })
  const sampleChooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Load sample JSON' }).click()
  await (await sampleChooser).setFiles({ name: 'sample.json', mimeType: 'application/json', buffer: Buffer.from('{}') })
  await page.getByRole('button', { name: 'PREVIEW' }).click()
  const reportDate = page.getByRole('textbox', { name: 'Value for params.reportDate' })
  await expect(reportDate).toBeVisible()
  await expect(page.getByLabel('Local render failure')).toContainText(/e1.*params\.reportDate|params\.reportDate.*e1/)
  await reportDate.fill('"2026-08-28T00:00:00Z"')
  await expect(page.getByText('STALE — inputs changed')).toBeVisible()
  await page.getByRole('button', { name: 'Render local PDF' }).click()
  await expect(page.getByText('EXACT LOCAL PRODUCTION PDF')).toBeVisible()
})
