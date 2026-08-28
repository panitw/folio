/// <reference lib="webworker" />

import { ENGINE_PROTOCOL_VERSION, type EngineError, type EngineRequest, type EngineSnapshot } from './engine-protocol'
import { EngineRequestAdmission } from './engine-worker-admission'
import { EngineWorkerQueue } from './engine-worker-queue'
import { runtimeAssetUrls } from './generated/offline-assets'

declare const Go: new () => { importObject: WebAssembly.Imports; run(instance: WebAssembly.Instance): void }

type WasmHost = { handle(request: string): string }
type WasmResponse = { ok: boolean; snapshot?: EngineSnapshot; bytesBase64?: string; diagnosticCode?: string; message?: string; elementId?: string; dataPath?: string }

const worker = self as unknown as DedicatedWorkerGlobalScope
let host: WasmHost | undefined
let bootFailure: EngineError | undefined
let lifecycleState: 'ready' | 'failed' | undefined
let booted: Promise<void> | undefined
const admission = new EngineRequestAdmission()
const queue = new EngineWorkerQueue<EngineRequest>(async (request) => {
  await booted
  if (bootFailure) { respondFailure(request.requestId, bootFailure); return }
  await execute(request)
})

function lifecycle(state: 'ready' | 'failed', error?: EngineError) {
  if (lifecycleState === state) return
  lifecycleState = state
  worker.postMessage({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'lifecycle', state, ...(error ? { error } : {}) })
}

async function boot(): Promise<void> {
  try {
    importScripts(runtimeAssetUrls.wasmExec)
    const go = new Go()
    const result = await WebAssembly.instantiateStreaming(fetch(runtimeAssetUrls.wasm), go.importObject)
    go.run(result.instance)
    const candidate = (globalThis as typeof globalThis & { FolioWasmHost?: WasmHost }).FolioWasmHost
    if (!candidate) throw new Error('wasm host did not register')
    if (bootFailure) return
    host = candidate
    lifecycle('ready')
  } catch {
    failWorker({ code: 'WASM_INITIALIZATION_FAILED', message: 'The Folio engine could not start' })
  }
}

worker.onmessage = (event: MessageEvent<unknown>) => {
  const result = admission.admit(event.data)
  if (result.kind === 'enqueue') { queue.enqueue(result.request); return }
  if (result.fatal) { failWorker(result.error); return }
  respondFailure(result.requestId!, result.error)
}

async function execute(request: EngineRequest): Promise<void> {
  try {
    const payloadBase64 = request.payload ? bytesToBase64(request.payload) : undefined
    const raw = host!.handle(JSON.stringify({ operation: request.operation, ...(payloadBase64 ? { payloadBase64 } : {}) }))
    const result = JSON.parse(raw) as WasmResponse
    if (!result.ok || !result.snapshot) {
      respondFailure(request.requestId, {
        code: result.diagnosticCode ?? 'ENGINE_REJECTED',
        message: (result.message ?? 'Engine rejected request').slice(0, 512),
        ...(result.elementId ? { elementId: result.elementId.slice(0, 128) } : {}),
        ...(result.dataPath ? { dataPath: result.dataPath.slice(0, 256) } : {}),
      })
      return
    }
    const bytes = result.bytesBase64 ? base64ToBytes(result.bytesBase64) : undefined
    worker.postMessage({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: request.requestId, ok: true, snapshot: result.snapshot, ...(bytes ? { bytes } : {}) }, bytes ? [bytes] : [])
  } catch {
    respondFailure(request.requestId, { code: 'WASM_PROTOCOL_FAILURE', message: 'The engine returned an invalid response' })
  }
}

function respondFailure(requestId: string, error: EngineError) {
  worker.postMessage({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId, ok: false, error })
}

function failWorker(error: EngineError): void {
  if (bootFailure) return
  bootFailure = error
  lifecycle('failed', error)
}

function bytesToBase64(bytes: ArrayBuffer): string { let text = ''; for (const byte of new Uint8Array(bytes)) text += String.fromCharCode(byte); return btoa(text) }
function base64ToBytes(value: string): ArrayBuffer { const text = atob(value); const bytes = new Uint8Array(text.length); for (let index = 0; index < text.length; index++) bytes[index] = text.charCodeAt(index); return bytes.buffer }

booted = boot()
