import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: { assetsInlineLimit: 0 },
  // Runtime inputs are copied through Vite's asset graph by build-wasm.mjs.
  // Never publish the mutable source files from public/ as application URLs.
  publicDir: false,
  test: { environment: 'jsdom', setupFiles: ['./src/test/setup.ts'], include: ['src/**/*.test.{ts,tsx}', 'scripts/**/*.test.mjs'] },
})
