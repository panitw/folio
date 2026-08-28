import { expect, test } from '@playwright/test'

// This package is deliberately compile-covered on Story 5.2. Execution is
// deferred with the full browser suite to the Epic 5 D-000.4 boundary.
test('the built application loads, serializes, and commits through the real Go worker', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByLabel('Document bar')).toBeVisible()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  await expect(page.getByRole('button', { name: 'Commit engine snapshot' })).toBeVisible()
  await page.getByRole('button', { name: 'Commit engine snapshot' }).click()
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 2/)
  await expect.poll(async () => page.evaluate(async () => {
    const [wasm, glue] = await Promise.all([fetch('/wasm/folio-engine.wasm'), fetch('/wasm/wasm_exec.js')])
    return wasm.ok && glue.ok
  })).toBe(true)
})
