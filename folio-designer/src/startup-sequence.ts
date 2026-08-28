import type { EngineClient } from './engine-client'
import type { EngineSnapshot } from './engine-protocol'

export async function loadStarterAfterEngineReady(clientPromise: Promise<EngineClient>, starterUrl: string, fetchStarter: (url: string) => Promise<Response>): Promise<Readonly<{ client: EngineClient; snapshot: EngineSnapshot }>> {
  const client = await clientPromise
  const source = await fetchStarter(starterUrl)
  if (!source.ok) throw new Error('starter template unavailable')
  const loaded = await client.request('load', await source.arrayBuffer())
  const serialized = await client.request('serialize')
  if (!serialized.bytes || serialized.bytes.byteLength !== loaded.snapshot.byteLength) throw new Error('canonical serialization unavailable')
  return { client, snapshot: loaded.snapshot }
}
