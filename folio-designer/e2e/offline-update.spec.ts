import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { expect, test } from '@playwright/test'
// @ts-expect-error The source module is executed by Vitest and Node; this
// browser-only spec is compile-covered through the Epic 5 boundary.
import { serviceWorkerSource } from '../scripts/offline-service-worker-template.mjs'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

async function currentRelease() {
  return JSON.parse(await readFile(join(root, 'dist', 'offline-release-manifest.json'), 'utf8')) as { id: string; pageId: string; workerRevision: string; assets: { url: string }[] }
}

function workerOnlyRevision(release: Awaited<ReturnType<typeof currentRelease>>) {
  return { ...release, id: 'f'.repeat(64), pageId: 'e'.repeat(64), workerRevision: 'd'.repeat(64) }
}

// These are intentionally executable Playwright lifecycle proofs. Execution
// is deferred by D-000.4 until Epic 5 closes; local Chromium absence is not a
// substitute for the unit-level lifecycle proofs in this story.
test('a complete worker-only update waits behind an old tab and preserves its cache', async ({ page }) => {
  await page.goto('/')
  await page.reload()
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
  const oldRelease = await currentRelease()
  const oldCache = `folio-release-${oldRelease.id}`
  await page.route('**/sw.js', async (route) => route.fulfill({ contentType: 'application/javascript', body: serviceWorkerSource(workerOnlyRevision(oldRelease)) }))
  await page.evaluate(async () => { await (await navigator.serviceWorker.getRegistration())?.update() })
  await expect.poll(() => page.evaluate(async () => Boolean((await navigator.serviceWorker.getRegistration())?.waiting))).toBe(true)
  expect(await page.evaluate(async (name) => (await caches.keys()).includes(name), oldCache)).toBe(true)
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
})

test('a failed replacement keeps the old complete release usable', async ({ page }) => {
  await page.goto('/')
  await page.reload()
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
  const oldRelease = await currentRelease()
  const oldCache = `folio-release-${oldRelease.id}`
  const doomed = oldRelease.assets.find((asset) => asset.url.endsWith('.wasm'))!
  await page.route(`**${doomed.url}`, async (route) => route.fulfill({ status: 500, body: 'install failure' }))
  await page.route('**/sw.js', async (route) => route.fulfill({ contentType: 'application/javascript', body: serviceWorkerSource(workerOnlyRevision(oldRelease)) }))
  await page.evaluate(async () => { await (await navigator.serviceWorker.getRegistration())?.update() })
  await expect.poll(() => page.evaluate(async (name) => (await caches.keys()).includes(name), oldCache)).toBe(true)
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
})
