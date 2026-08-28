import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import { getEngineClient } from './engine-client.ts'

const root = createRoot(document.getElementById('root')!)

// The composition root owns the one client/worker lifetime. The starter file
// is opaque bytes to TypeScript; Go parses, canonicalizes, snapshots and then
// serializes it before React receives the small immutable projection.
void (async () => {
  try {
    const client = await getEngineClient()
    const source = await fetch('/templates/starter.folio')
    if (!source.ok) throw new Error('starter template unavailable')
    const input = await source.arrayBuffer()
    const loaded = await client.request('load', input)
    const serialized = await client.request('serialize')
    if (!serialized.bytes || serialized.bytes.byteLength !== loaded.snapshot.byteLength) throw new Error('canonical serialization unavailable')
    root.render(<StrictMode><App engine={client} initialSnapshot={loaded.snapshot} /></StrictMode>)
  } catch {
    root.render(<StrictMode><App initializationError="Engine unavailable" /></StrictMode>)
  }
})()
