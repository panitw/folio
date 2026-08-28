import { copyBytes, deepFreeze, ENGINE_PROTOCOL_VERSION, parseInbound, type EngineError, type EngineInbound, type EngineOperation, type EngineRequest, type EngineSnapshot } from './engine-protocol'

export interface WorkerPort {
  postMessage(message: EngineRequest, transfer?: Transferable[]): void
  terminate(): void
  onmessage: ((event: MessageEvent<unknown>) => void) | null
  onerror: ((event: ErrorEvent) => void) | null
}

export type EngineResult = Readonly<{ snapshot: EngineSnapshot; bytes?: ArrayBuffer }>

type Pending = { resolve: (result: EngineResult) => void; reject: (error: Error) => void }
type ClientState = 'starting' | 'ready' | 'failed' | 'terminated'

const errorFor = (code: string, message: string, dataPath?: string) => Object.assign(new Error(message), { code, ...(dataPath ? { dataPath } : {}) })

export class EngineClient {
	#state: ClientState = 'starting'
	#nextRequest = 0
	#pending = new Map<string, Pending>()
	#order: string[] = []
	#abandoned = new Set<string>()
	#ready: Promise<EngineClient>
	#resolveReady!: (client: EngineClient) => void
	#rejectReady!: (error: Error) => void
	private readonly worker: WorkerPort

	constructor(worker: WorkerPort) {
    this.worker = worker
    this.#ready = new Promise<EngineClient>((resolve, reject) => { this.#resolveReady = resolve; this.#rejectReady = reject })
    // Direct clients are useful in focused tests and do not always await the
    // lifecycle promise. Keep a handled observer while callers still receive
    // the original rejection from whenReady()/the singleton.
    void this.#ready.catch(() => undefined)
    worker.onmessage = (event) => this.#onMessage(event.data)
    worker.onerror = () => this.#fail('WORKER_RUNTIME_ERROR', 'The engine worker failed')
  }

  get state(): ClientState { return this.#state }
  whenReady(): Promise<EngineClient> { return this.#ready }

  request(operation: EngineOperation, payload?: ArrayBuffer, signal?: AbortSignal): Promise<EngineResult> {
    if (this.#state !== 'ready') return Promise.reject(errorFor('ENGINE_NOT_READY', `Engine is ${this.#state}`))
    if (signal?.aborted) return Promise.reject(errorFor('REQUEST_ABORTED', 'Engine request was abandoned'))
    const requestId = `request-${++this.#nextRequest}`
    const payloadCopy = payload ? copyBytes(payload) : undefined
    const request: EngineRequest = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId, operation, ...(payloadCopy ? { payload: payloadCopy } : {}) }
    return new Promise<EngineResult>((resolve, reject) => {
      const abort = () => {
        if (this.#pending.delete(requestId)) {
          this.#abandoned.add(requestId)
          reject(errorFor('REQUEST_ABORTED', 'Engine request was abandoned'))
        }
      }
      signal?.addEventListener('abort', abort, { once: true })
      this.#pending.set(requestId, {
        resolve: (result) => { signal?.removeEventListener('abort', abort); resolve(result) },
        reject: (error) => { signal?.removeEventListener('abort', abort); reject(error) },
      })
      this.#order.push(requestId)
      try { workerPost(this.worker, request, payloadCopy) } catch { this.#fail('WORKER_POST_FAILED', 'Could not send engine request') }
    })
  }

  terminate(): void {
    if (this.#state === 'terminated') return
    this.#state = 'terminated'
    this.#detach()
    this.worker.terminate()
    this.#rejectReady(errorFor('ENGINE_TERMINATED', 'The engine worker was terminated'))
    this.#rejectPending('ENGINE_TERMINATED', 'The engine worker was terminated')
  }

  #onMessage(raw: unknown): void {
    const message = parseInbound(raw)
    if (!message) { this.#fail('PROTOCOL_INVALID', 'The engine sent an invalid protocol message'); return }
    if (message.kind === 'lifecycle') {
      if (message.state === 'ready' && this.#state === 'starting') { this.#state = 'ready'; this.#resolveReady(this); return }
      this.#fail(message.error?.code ?? 'LIFECYCLE_INVALID', message.error?.message ?? 'Invalid engine lifecycle transition')
      return
    }
    this.#settle(message)
  }

  #settle(message: Exclude<EngineInbound, { kind: 'lifecycle' }>): void {
    if (this.#order[0] !== message.requestId) {
      this.#fail('PROTOCOL_OUT_OF_ORDER', 'The engine sent responses out of FIFO order')
      return
    }
    this.#order.shift()
    if (this.#abandoned.delete(message.requestId)) return
    const pending = this.#pending.get(message.requestId)
    if (!pending) { this.#fail('PROTOCOL_DUPLICATE_OR_UNKNOWN', 'The engine sent an unknown or duplicate response'); return }
    this.#pending.delete(message.requestId)
    if (!message.ok) { pending.reject(errorFor(message.error.code, safeErrorMessage(message.error), message.error.dataPath)); return }
    const snapshot = deepFreeze({ ...message.snapshot }) as EngineSnapshot
    const bytes = message.bytes ? copyBytes(message.bytes) : undefined
    pending.resolve(deepFreeze({ snapshot, ...(bytes ? { bytes } : {}) }))
  }

  #fail(code: string, message: string): void {
    if (this.#state === 'failed' || this.#state === 'terminated') return
    this.#state = 'failed'
    this.#detach()
    this.worker.terminate()
    this.#rejectReady(errorFor(code, message))
    this.#rejectPending(code, message)
  }

  #rejectPending(code: string, message: string): void {
    for (const pending of this.#pending.values()) pending.reject(errorFor(code, message))
    this.#pending.clear()
    this.#abandoned.clear()
    this.#order = []
  }

  #detach(): void { this.worker.onmessage = null; this.worker.onerror = null }
}

function workerPost(worker: WorkerPort, request: EngineRequest, payload?: ArrayBuffer): void {
  worker.postMessage(request, payload ? [payload] : [])
}

export function createEngineClientSingleton(createWorker: () => WorkerPort): () => Promise<EngineClient> {
	let singleton: Promise<EngineClient> | undefined
	return () => {
		if (!singleton) {
      try {
        const client = new EngineClient(createWorker())
        singleton = client.whenReady()
      } catch {
        singleton = Promise.reject(errorFor('WORKER_CONSTRUCTION_FAILED', 'The engine worker could not be constructed'))
      }
    }
		return singleton
	}
}

function safeErrorMessage(error: EngineError): string {
  return error.message.slice(0, 512)
}

// This is the single discoverable Worker construction site in production.
const appEngineClient = createEngineClientSingleton(() => new Worker(new URL('./engine.worker.ts', import.meta.url), { name: 'folio-engine' }))
export const getEngineClient = (): Promise<EngineClient> => appEngineClient()
