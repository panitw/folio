import { describe, expect, it } from 'vitest'
import { createEngineClientSingleton, EngineClient, type WorkerPort } from './engine-client'
import { ENGINE_PROTOCOL_VERSION, type EngineRequest } from './engine-protocol'

class FakeWorker implements WorkerPort {
  onmessage: ((event: MessageEvent<unknown>) => void) | null = null
  onerror: ((event: ErrorEvent) => void) | null = null
  readonly sent: EngineRequest[] = []
  terminated = 0
  postMessage(message: EngineRequest): void { this.sent.push(message) }
  terminate(): void { this.terminated++ }
  emit(data: unknown): void { this.onmessage?.({ data } as MessageEvent<unknown>) }
  ready(): void { this.emit({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'lifecycle', state: 'ready' }) }
  respond(requestId: string, revision: number, bytes?: ArrayBuffer): void { this.emit({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId, ok: true, snapshot: { documentState: 'loaded', revision, byteLength: 10 }, ...(bytes ? { bytes } : {}) }) }
}

describe('engine client protocol and lifecycle', () => {
  it('shares one worker and initialization promise across repeated consumers', async () => {
    let constructed = 0
    const worker = new FakeWorker()
    const getClient = createEngineClientSingleton(() => { constructed++; return worker })
    const firstPromise = getClient()
    const secondPromise = getClient()
    worker.ready()
    const [first, second] = await Promise.all([firstPromise, secondPromise])
    expect(first).toBe(second)
    expect(constructed).toBe(1) // Red proof target: a second constructor fails here.
  })

  it('rejects calls before startup and settles FIFO-correlated immutable results', async () => {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    await expect(client.request('snapshot')).rejects.toMatchObject({ code: 'ENGINE_NOT_READY' })
    worker.ready()
    const first = client.request('snapshot')
    const second = client.request('serialize')
    expect(worker.sent.map((request) => request.requestId)).toEqual(['request-1', 'request-2'])
    worker.respond('request-1', 1)
    worker.respond('request-2', 2, new Uint8Array([1, 2]).buffer)
    const [one, two] = await Promise.all([first, second])
    expect(one.snapshot.revision).toBe(1)
    expect(two.snapshot.revision).toBe(2)
    expect(Object.isFrozen(one.snapshot)).toBe(true)
    expect(() => { (one.snapshot as { revision: number }).revision = 99 }).toThrow()
    expect(new Uint8Array(two.bytes!)).toEqual(new Uint8Array([1, 2]))
  })

  it('fails closed rather than applying an out-of-order response', async () => {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    worker.ready()
    const first = client.request('snapshot')
    const second = client.request('serialize')
    worker.respond('request-2', 2)
    await expect(first).rejects.toMatchObject({ code: 'PROTOCOL_OUT_OF_ORDER' })
    await expect(second).rejects.toMatchObject({ code: 'PROTOCOL_OUT_OF_ORDER' })
    expect(client.state).toBe('failed')
  })

  it('fails closed on a duplicate, unknown version, and termination, rejecting pending work once', async () => {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    worker.ready()
    const pending = client.request('snapshot')
    worker.respond('request-1', 1)
    await pending
    worker.respond('request-1', 1) // Red proof: duplicate response cannot settle twice.
    expect(client.state).toBe('failed')
    expect(worker.terminated).toBe(1)

    const secondWorker = new FakeWorker()
    const secondClient = new EngineClient(secondWorker)
    secondWorker.ready()
    const outstanding = secondClient.request('snapshot')
    secondClient.terminate()
    await expect(outstanding).rejects.toMatchObject({ code: 'ENGINE_TERMINATED' })

    const thirdWorker = new FakeWorker()
    const thirdClient = new EngineClient(thirdWorker)
    thirdWorker.emit({ protocolVersion: 999, kind: 'lifecycle', state: 'ready' })
    expect(thirdClient.state).toBe('failed')
  })

  it('abandons a caller without cancelling or mutating worker state', async () => {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    worker.ready()
    const controller = new AbortController()
    const pending = client.request('snapshot', undefined, controller.signal)
    controller.abort()
    await expect(pending).rejects.toMatchObject({ code: 'REQUEST_ABORTED' })
    worker.respond('request-1', 1) // late, legitimate terminal response is consumed
    expect(client.state).toBe('ready')
    expect(worker.terminated).toBe(0)
  })

  it('preserves bounded component diagnostic location on a command rejection', async () => {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    worker.ready()
    const pending = client.request('command', new Uint8Array([1]).buffer)
    worker.emit({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'request-1', ok: false, error: { code: 'COMPONENT_INVALID', message: 'bad x', elementId: 'e1', dataPath: 'component.x' } })
    await expect(pending).rejects.toMatchObject({ code: 'COMPONENT_INVALID', elementId: 'e1', dataPath: 'component.x', message: 'bad x' })
  })

  it('fails closed when every operation receives surplus table metadata', async () => {
		for (const operation of ['render', 'identity', 'serialize', 'parameter-references', 'snapshot', 'command', 'undo', 'redo'] as const) {
    const worker = new FakeWorker()
    const client = new EngineClient(worker)
    worker.ready()
			const payload = operation === 'render' ? { template: new Uint8Array([1]).buffer, data: new Uint8Array([2]).buffer, params: new Uint8Array([3]).buffer } : operation === 'identity' ? { data: new Uint8Array([1]).buffer, params: new Uint8Array([2]).buffer } : operation === 'command' ? new Uint8Array([1]).buffer : undefined
			const pending = client.request(operation, payload)
			const base = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response' as const, requestId: 'request-1', ok: true as const, snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 10 }, tableColumns: { revision: 1, table: { tableId: 'e7', columns: [] } } }
			worker.emit(operation === 'render' ? { ...base, bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'a'.repeat(64), pdfSha256: 'b'.repeat(64), diagnostics: [] } } : operation === 'identity' ? { ...base, preview: { revision: 1, identity: 'a'.repeat(64) } } : operation === 'serialize' ? { ...base, bytes: new Uint8Array([9]).buffer } : operation === 'parameter-references' ? { ...base, parameterReferences: { revision: 1, names: [] } } : base)
			await expect(pending).rejects.toMatchObject({ code: 'PROTOCOL_OPERATION_MISMATCH' })
			expect(client.state).toBe('failed')
		}
  })

  it('fails closed when a table-column reply carries unrelated success payloads', async () => {
		const worker = new FakeWorker()
		const client = new EngineClient(worker)
		worker.ready()
		const pending = client.request('table-columns', new TextEncoder().encode('{"id":"e7"}').buffer)
		worker.emit({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'request-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 10 }, tableColumns: { revision: 1, table: { tableId: 'e7', columns: [] } }, parameterReferences: { revision: 1, names: [] } })
		await expect(pending).rejects.toMatchObject({ code: 'PROTOCOL_OPERATION_MISMATCH' })
    expect(client.state).toBe('failed')
  })

  it('rejects singleton startup failure and clears listeners on every terminal state', async () => {
    const worker = new FakeWorker()
    const getClient = createEngineClientSingleton(() => worker)
    const starting = getClient()
    worker.emit({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'lifecycle', state: 'failed', error: { code: 'WASM_INITIALIZATION_FAILED', message: 'safe' } })
    await expect(starting).rejects.toMatchObject({ code: 'WASM_INITIALIZATION_FAILED' })
    expect(worker.onmessage).toBeNull()
    expect(worker.onerror).toBeNull()

    const terminated = new FakeWorker()
    const client = new EngineClient(terminated)
    client.terminate()
    expect(terminated.onmessage).toBeNull()
    expect(terminated.onerror).toBeNull()
  })
})
