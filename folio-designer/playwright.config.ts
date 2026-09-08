import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  workers: 1,
  timeout: 90_000,
  expect: { timeout: 30_000 },
  use: {
    baseURL: 'http://127.0.0.1:4173',
    launchOptions: {
      executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH,
    },
  },
  // This command is not "start a server" -- it is a full cold production
  // build first, and `npm run build` runs build:wasm, which compiles the Go
  // engine to wasm. That is 141s on a warm developer laptop. On a bare
  // ubuntu-24.04 runner with no Go module or build cache (no setup-go step in
  // this workflow sets cache-dependency-path, so none of them cache), it is
  // slower still -- and at 180_000 it timed out on every push, failing the
  // folio-designer-e2e job four runs running with
  // `Timed out waiting 180000ms from config.webServer`.
  //
  // 600_000 is a budget, not a wait: a server that comes up in 200s costs
  // 200s. It stays well inside the job's 45-minute cap, so a build that
  // genuinely hangs still fails the job rather than the workflow.
  webServer: {
    command: 'npm run build && npm run preview -- --host 127.0.0.1 --port 4173',
    url: 'http://127.0.0.1:4173',
    reuseExistingServer: false,
    timeout: 600_000,
  },
})
