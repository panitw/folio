import { expect, test } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const fixture = readFileSync(path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../fixtures/statement-1/input.folio'))

// Compiled but not executed until the Epic 5 D-000.4 browser gate. These use
// real worker and browser picker/download seams; no TypeScript bytes are made.
test('the fallback opens a local file, saves worker output as a download, and remains local while offline', async ({ page, context }) => {
  await page.goto('/')
  await context.setOffline(true)
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await chooser).setFiles({ name: 'statement.folio', mimeType: 'application/json', buffer: fixture })
  await expect(page.getByText('statement.folio')).toBeVisible()
  const download = page.waitForEvent('download')
  await page.keyboard.press(process.platform === 'darwin' ? 'Meta+S' : 'Control+S')
  await expect((await download).suggestedFilename()).resolves.toBe('statement.folio')
  await expect(page.getByRole('status')).toContainText(/Downloaded local file|Saved local file/)
})

test('the activation-gated tier opens, Save As picks before the worker write, and writes opaque worker bytes', async ({ page }) => {
  await page.addInitScript((rawBytes) => {
    const bytes = new Uint8Array(rawBytes)
    const handle = {
      name: 'native.folio',
      getFile: async () => new File([bytes], 'native.folio', { type: 'application/json' }),
      createWritable: async () => ({ write: async (written: ArrayBuffer) => { (window as typeof window & { __folioWrites?: number[][] }).__folioWrites = [Array.from(new Uint8Array(written))] }, close: async () => undefined }),
    }
    Object.assign(window, { showOpenFilePicker: async () => [handle], showSaveFilePicker: async () => handle })
  }, [...fixture])
  await page.goto('/')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await expect(page.getByText('native.folio')).toBeVisible()
  await page.getByRole('button', { name: 'Save As' }).click()
  await expect.poll(() => page.evaluate(() => (window as typeof window & { __folioWrites?: number[][] }).__folioWrites?.[0]?.length ?? 0)).toBeGreaterThan(0)
  await expect(page.getByRole('status')).toContainText(/Saved locally as|Saved local file/)
})

test('a native Save As cancellation keeps the opened local identity and unsaved indicator honest', async ({ page }) => {
  await page.addInitScript((rawBytes) => {
    const bytes = new Uint8Array(rawBytes)
    Object.assign(window, {
      showOpenFilePicker: async () => [{ name: 'cancelled.folio', getFile: async () => new File([bytes], 'cancelled.folio', { type: 'application/json' }), createWritable: async () => { throw new Error('not reached') } }],
      showSaveFilePicker: async () => { throw new DOMException('cancel', 'AbortError') },
    })
  }, [...fixture])
  await page.goto('/')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await expect(page.getByText('cancelled.folio')).toBeVisible()
  await page.getByRole('button', { name: 'Save As' }).click()
  await expect(page.getByText('cancelled.folio')).toBeVisible()
  await expect(page.getByText('Saved local file')).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})
