import { expect, test } from '@playwright/test'

// This package is compile-covered each story; execution is deferred to the
// Epic 5 D-000.4 boundary because it needs a real browser SW lifecycle.
test('the built application loads, serializes, and commits through the real Go worker', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByLabel('Document bar')).toBeVisible()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await expect(page.getByRole('button', { name: 'Commit engine snapshot' })).toBeVisible()
  await page.getByRole('button', { name: 'Commit engine snapshot' }).click()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 2/)
})

test('a complete release reloads through the service worker while offline without a network failure', async ({ page, context }) => {
  const failures: string[] = []
  page.on('requestfailed', (request) => failures.push(request.url()))
  await page.goto('/')
  await page.reload() // the first installation must activate before it can control a reload
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
  await context.setOffline(true)
  await page.reload()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await expect(page.getByTestId('offline-status')).toHaveText('Offline ready')
  expect(failures).toEqual([])
})
