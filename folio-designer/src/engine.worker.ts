/// <reference lib="webworker" />

import { ENGINE_PROTOCOL_VERSION, MAX_ENGINE_RENDER_PDF_BYTES, type EngineDiagnostic, type EngineError, type EngineRequest, type EngineSnapshot, type IdentityPayload, type RenderPayload } from './engine-protocol'
import { EngineRequestAdmission } from './engine-worker-admission'
import { EngineWorkerQueue } from './engine-worker-queue'
import { runtimeAssetUrls } from './generated/offline-assets'

declare const Go: new () => { importObject: WebAssembly.Imports; run(instance: WebAssembly.Instance): void }

type WasmHost = { handle(request: string): string }
type WasmResponse = { ok: boolean; snapshot?: EngineSnapshot; bytesBase64?: string; diagnosticCode?: string; message?: string; elementId?: string; dataPath?: string; pdfSha256?: string; previewIdentity?: string; renderRevision?: number; diagnostics?: EngineDiagnostic[]; parameterReferences?: string[]; parameterReferenceRevision?: number }

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
    const render = request.operation === 'render' ? request.payload as RenderPayload : undefined
    const identity = request.operation === 'identity' ? request.payload as IdentityPayload : undefined
    const payloadBase64 = request.payload instanceof ArrayBuffer ? bytesToBase64(request.payload) : undefined
    const raw = host!.handle(JSON.stringify({ operation: request.operation, ...(payloadBase64 ? { payloadBase64 } : {}), ...(render ? { templateBase64: bytesToBase64(render.template), dataBase64: bytesToBase64(render.data), paramsBase64: bytesToBase64(render.params) } : {}), ...(identity ? { dataBase64: bytesToBase64(identity.data), paramsBase64: bytesToBase64(identity.params) } : {}) }))
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
    // Reject a hostile/buggy producer before atob allocates its decoded
    // buffer. The protocol repeats this guard on the main-thread boundary.
    const bytes = result.bytesBase64 ? base64ToBytesBounded(result.bytesBase64, request.operation === 'render' ? MAX_ENGINE_RENDER_PDF_BYTES : undefined) : undefined
    const preview = request.operation === 'render' ? { revision: result.renderRevision, identity: result.previewIdentity, pdfSha256: result.pdfSha256, diagnostics: result.diagnostics } : request.operation === 'identity' ? { revision: result.renderRevision, identity: result.previewIdentity } : undefined
    const parameterReferences = request.operation === 'parameter-references' ? { revision: result.parameterReferenceRevision, names: result.parameterReferences } : undefined
    worker.postMessage({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: request.requestId, ok: true, snapshot: result.snapshot, ...(bytes ? { bytes } : {}), ...(preview ? { preview } : {}), ...(parameterReferences ? { parameterReferences } : {}) }, bytes ? [bytes] : [])
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
function base64ToBytesBounded(value: string, max?: number): ArrayBuffer {
  const padding = value.endsWith('==') ? 2 : value.endsWith('=') ? 1 : 0
  const byteLength = value.length % 4 === 0 ? value.length / 4 * 3 - padding : -1
  if (byteLength < 0 || (max !== undefined && byteLength > max)) throw new Error('WASM byte response exceeds its transport limit')
  const text = atob(value)
  if (max !== undefined && text.length > max) throw new Error('WASM byte response exceeds its transport limit')
  const bytes = new Uint8Array(text.length); for (let index = 0; index < text.length; index++) bytes[index] = text.charCodeAt(index); return bytes.buffer
}

booted = boot()
