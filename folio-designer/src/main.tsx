import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import { getEngineClient, type EngineClient } from './engine-client.ts'
import type { EngineSnapshot } from './engine-protocol.ts'
import { registerOfflineLifecycle, type OfflineLifecycle } from './offline-lifecycle.ts'
import { loadS1Payload, type S1Payload } from './release-payload.ts'
import { runtimeAssetUrls } from './generated/offline-assets.ts'
import { loadStarterAfterEngineReady } from './startup-sequence.ts'
import { selectFileAccess, selectSampleFileAccess } from './file/capability.ts'

const root = createRoot(document.getElementById('root')!)
let lifecycle: OfflineLifecycle = { state: 'checking', cacheReady: false, verifiedAssetUrls: [] }
let payload: S1Payload | undefined
let engine: EngineClient | undefined
let snapshot: EngineSnapshot | undefined
let blankBytes: ArrayBuffer | undefined
let engineState: 'waiting' | 'starting' | 'failed' = 'waiting'
let started = false
let stopObservation: (() => void) | undefined
let observationInFlight = false
const fileAccess = selectFileAccess()
const sampleFileAccess = selectSampleFileAccess()
const render = () => root.render(<StrictMode><App key={engine ? 'engine-ready' : 'engine-loading'} engine={engine} fileAccess={fileAccess} sampleFileAccess={sampleFileAccess} initialSnapshot={snapshot} blankBytes={blankBytes} offlineState={lifecycle.state} loadState={lifecycle} payload={payload} engineState={engineState} onRetry={startObservation} /></StrictMode>)
async function startEngine() {
  if (started || !lifecycle.cacheReady) return
  started = true; engineState = 'starting'; render()
  try {
    const startedEngine = await loadStarterAfterEngineReady(getEngineClient(), runtimeAssetUrls.starter, (url) => fetch(url, { credentials: 'omit' }))
    engine = startedEngine.client; snapshot = startedEngine.snapshot; blankBytes = startedEngine.blankBytes; render()
  } catch { engineState = 'failed'; render() }
}
async function startObservation() {
  if (engineState === 'failed') { window.location.reload(); return }
  if (observationInFlight) return
  observationInFlight = true
  stopObservation?.()
  lifecycle = { state: 'checking', cacheReady: false, verifiedAssetUrls: [] }; engineState = 'waiting'; render()
  try {
    payload = loadS1Payload()
    const expectedPageId = document.querySelector('meta[name="folio-page-release"]')?.getAttribute('content') ?? undefined
    stopObservation = registerOfflineLifecycle(expectedPageId, payload, (next) => { lifecycle = next; render(); void startEngine() })
    render()
  } finally { observationInFlight = false }
}
void startObservation()
