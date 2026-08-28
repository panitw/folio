export const ENGINE_PROTOCOL_VERSION = 1 as const

export type EngineOperation = 'initialize' | 'load' | 'snapshot' | 'validate' | 'serialize' | 'command'

export const MAX_ENGINE_REQUEST_ID_LENGTH = 128
export const MAX_ENGINE_PAYLOAD_BYTES = 8 * 1024 * 1024

export type EngineError = Readonly<{
  code: string
  message: string
  elementId?: string
  dataPath?: string
}>

// Opaque bytes/JSON are deliberately the only document-bearing values on this
// boundary. These types describe transport, not the .folio file format.
export type EngineRequest = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'request'
  requestId: string
  operation: EngineOperation
  payload?: ArrayBuffer
}>

export type EngineSnapshot = Readonly<{
  documentState: 'empty' | 'loaded'
  revision: number
  byteLength: number
	canvas?: CanvasProjection
}>

// This is paint-only output from Go, not a .folio page model. Values are
// millipoints and are never used to derive a browser document layout.
export type CanvasProjection = Readonly<{
	width: number; height: number; orientation: 'portrait' | 'landscape'; preset: 'A4' | 'Letter' | 'custom'
	marginTop: number; marginRight: number; marginBottom: number; marginLeft: number; gridIncrement: number; commandWidth: number; commandHeight: number
	bands: ReadonlyArray<Readonly<{ name: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number }>>
}>

export type EngineSuccess = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'response'
  requestId: string
  ok: true
  snapshot: EngineSnapshot
  bytes?: ArrayBuffer
}>

export type EngineFailure = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'response'
  requestId: string
  ok: false
  error: EngineError
}>

export type EngineLifecycle = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'lifecycle'
  state: 'ready' | 'failed'
  error?: EngineError
}>

export type EngineInbound = EngineSuccess | EngineFailure | EngineLifecycle

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null && Object.getPrototypeOf(value) === Object.prototype
const isArrayBuffer = (value: unknown): value is ArrayBuffer => Object.prototype.toString.call(value) === '[object ArrayBuffer]'
export const isEngineRequestId = (value: unknown): value is string => typeof value === 'string' && value.length > 0 && value.length <= MAX_ENGINE_REQUEST_ID_LENGTH && /^[A-Za-z0-9][A-Za-z0-9._:-]*$/.test(value)

const isError = (value: unknown): value is EngineError => isRecord(value) && typeof value.code === 'string' && value.code.length > 0 && value.code.length <= 96 && typeof value.message === 'string' && value.message.length <= 512 && (value.elementId === undefined || typeof value.elementId === 'string') && (value.dataPath === undefined || typeof value.dataPath === 'string')
const isCanvas = (value: unknown): value is CanvasProjection => {
  if (!isRecord(value) || !['A4', 'Letter', 'custom'].includes(value.preset as string) || (value.orientation !== 'portrait' && value.orientation !== 'landscape')) return false
  const integer = (key: string, positive = false) => typeof value[key] === 'number' && Number.isSafeInteger(value[key]) && (positive ? value[key] > 0 : value[key] >= 0)
  if (!['width', 'height', 'gridIncrement', 'commandWidth', 'commandHeight'].every((key) => integer(key, true)) || !['marginTop', 'marginRight', 'marginBottom', 'marginLeft'].every((key) => integer(key))) return false
  if (!Array.isArray(value.bands) || value.bands.length !== 3) return false
  const names = ['pageHeader', 'content', 'pageFooter']
  const page = value as Record<string, number>
  return value.bands.every((band, index) => {
    if (!isRecord(band) || band.name !== names[index] || !['x', 'y', 'width', 'height'].every((key) => typeof band[key] === 'number' && Number.isSafeInteger(band[key]))) return false
    const paint = band as Record<string, number>
    return paint.x >= 0 && paint.y >= 0 && paint.width > 0 && paint.height >= 0 && paint.x + paint.width <= page.width && paint.y + paint.height <= page.height
  })
}
const isSnapshot = (value: unknown): value is EngineSnapshot => isRecord(value) && (value.documentState === 'empty' || value.documentState === 'loaded') && typeof value.revision === 'number' && Number.isSafeInteger(value.revision) && value.revision >= 0 && typeof value.byteLength === 'number' && Number.isSafeInteger(value.byteLength) && value.byteLength >= 0 && (value.canvas === undefined || isCanvas(value.canvas))

export function requestCorrelationId(value: unknown): string | undefined {
  return isRecord(value) && isEngineRequestId(value.requestId) ? value.requestId : undefined
}

export function parseRequest(value: unknown): EngineRequest | undefined {
  if (!isRecord(value) || value.protocolVersion !== ENGINE_PROTOCOL_VERSION || value.kind !== 'request' || !isEngineRequestId(value.requestId)) return undefined
  if (!['initialize', 'load', 'snapshot', 'validate', 'serialize', 'command'].includes(value.operation as string)) return undefined
  if (value.payload !== undefined && (!isArrayBuffer(value.payload) || value.payload.byteLength > MAX_ENGINE_PAYLOAD_BYTES)) return undefined
  return value as EngineRequest
}

export function parseInbound(value: unknown): EngineInbound | undefined {
  if (!isRecord(value) || value.protocolVersion !== ENGINE_PROTOCOL_VERSION || typeof value.kind !== 'string') return undefined
  if (value.kind === 'lifecycle') {
    if ((value.state === 'ready' && value.error === undefined) || (value.state === 'failed' && isError(value.error))) return value as EngineLifecycle
    return undefined
  }
  if (value.kind !== 'response' || typeof value.requestId !== 'string' || typeof value.ok !== 'boolean') return undefined
  if (value.ok && isSnapshot(value.snapshot) && (value.bytes === undefined || isArrayBuffer(value.bytes))) return value as EngineSuccess
  if (!value.ok && isError(value.error)) return value as EngineFailure
  return undefined
}

export function copyBytes(bytes: ArrayBuffer): ArrayBuffer { return bytes.slice(0) }

export function deepFreeze<T>(value: T): Readonly<T> {
  if (value && typeof value === 'object' && !Object.isFrozen(value)) {
    Object.freeze(value)
    for (const child of Object.values(value)) deepFreeze(child)
  }
  return value as Readonly<T>
}
