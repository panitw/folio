import { copyBytes, deepFreeze, ENGINE_PROTOCOL_VERSION, parseInbound, type EngineError, type EngineInbound, type EngineOperation, type EngineRequest, type EngineSnapshot, type IdentityPayload, type RenderPayload } from './engine-protocol'

export interface WorkerPort {
  postMessage(message: EngineRequest, transfer?: Transferable[]): void
  terminate(): void
  onmessage: ((event: MessageEvent<unknown>) => void) | null
  onerror: ((event: ErrorEvent) => void) | null
}

export type EngineResult = Readonly<{ snapshot: EngineSnapshot; bytes?: ArrayBuffer; preview?: Readonly<{ revision: number; identity: string; pdfSha256?: string; diagnostics?: ReadonlyArray<{ severity: 'warning'; code: string; elementId: string; dataPath: string; message: string }> }>; parameterReferences?: Readonly<{ revision: number; names: ReadonlyArray<string> }> }>

type Pending = { operation: EngineOperation; resolve: (result: EngineResult) => void; reject: (error: Error) => void }
type ClientState = 'starting' | 'ready' | 'failed' | 'terminated'

const errorFor = (code: string, message: string, dataPath?: string, elementId?: string) => Object.assign(new Error(message), { code, ...(dataPath ? { dataPath } : {}), ...(elementId ? { elementId } : {}) })

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

  request(operation: EngineOperation, payload?: ArrayBuffer | RenderPayload | IdentityPayload, signal?: AbortSignal): Promise<EngineResult> {
    if (this.#state !== 'ready') return Promise.reject(errorFor('ENGINE_NOT_READY', `Engine is ${this.#state}`))
    if (signal?.aborted) return Promise.reject(errorFor('REQUEST_ABORTED', 'Engine request was abandoned'))
    const requestId = `request-${++this.#nextRequest}`
    const payloadCopy = payload ? copyPayload(payload) : undefined
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
        operation,
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
    if (!message.ok) { pending.reject(errorFor(message.error.code, safeErrorMessage(message.error), message.error.dataPath, message.error.elementId)); return }
		const mismatch =
      (pending.operation === 'render' && (!message.bytes || !message.preview?.pdfSha256 || !message.preview.diagnostics || message.parameterReferences !== undefined)) ||
      (pending.operation === 'identity' && (!message.preview || message.bytes || message.preview.pdfSha256 !== undefined || message.preview.diagnostics !== undefined || message.parameterReferences !== undefined)) ||
      (pending.operation === 'parameter-references' && (!message.parameterReferences || message.bytes || message.preview)) ||
      (pending.operation === 'serialize' && (!message.bytes || message.preview || message.parameterReferences)) ||
      (!['render', 'identity', 'serialize', 'parameter-references'].includes(pending.operation) && (message.bytes || message.preview || message.parameterReferences))
		if (mismatch) { pending.reject(errorFor('PROTOCOL_OPERATION_MISMATCH', 'The engine response did not match its request')); this.#fail('PROTOCOL_OPERATION_MISMATCH', 'The engine response did not match its request'); return }
    const snapshot = deepFreeze({ ...message.snapshot }) as EngineSnapshot
    const bytes = message.bytes ? copyBytes(message.bytes) : undefined
		const preview = message.preview ? deepFreeze({ revision: message.preview.revision, identity: message.preview.identity, ...(message.preview.pdfSha256 ? { pdfSha256: message.preview.pdfSha256, diagnostics: message.preview.diagnostics!.map((diagnostic) => ({ ...diagnostic })) } : {}) }) : undefined
		const parameterReferences = message.parameterReferences ? deepFreeze({ revision: message.parameterReferences.revision, names: [...message.parameterReferences.names] }) : undefined
		pending.resolve(deepFreeze({ snapshot, ...(bytes ? { bytes } : {}), ...(preview ? { preview } : {}), ...(parameterReferences ? { parameterReferences } : {}) }))
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

function copyPayload(payload: ArrayBuffer | RenderPayload | IdentityPayload): ArrayBuffer | RenderPayload | IdentityPayload {
  return isArrayBuffer(payload) ? copyBytes(payload) : 'template' in payload ? { template: copyBytes(payload.template), data: copyBytes(payload.data), params: copyBytes(payload.params) } : { data: copyBytes(payload.data), params: copyBytes(payload.params) }
}

function workerPost(worker: WorkerPort, request: EngineRequest, payload?: ArrayBuffer | RenderPayload | IdentityPayload): void {
  worker.postMessage(request, isArrayBuffer(payload) ? [payload] : payload ? ('template' in payload ? [payload.template, payload.data, payload.params] : [payload.data, payload.params]) : [])
}

function isArrayBuffer(value: unknown): value is ArrayBuffer { return Object.prototype.toString.call(value) === '[object ArrayBuffer]' }

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
