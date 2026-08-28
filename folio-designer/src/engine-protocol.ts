export const ENGINE_PROTOCOL_VERSION = 1 as const

export type EngineOperation = 'initialize' | 'load' | 'snapshot' | 'validate' | 'serialize' | 'command'

export const MAX_ENGINE_REQUEST_ID_LENGTH = 128
export const MAX_ENGINE_PAYLOAD_BYTES = 8 * 1024 * 1024
export const MAX_ENGINE_ELEMENT_ID_LENGTH = 128
export const MAX_ENGINE_DATA_PATH_LENGTH = 256

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
	components: ReadonlyArray<Readonly<{ id: string; type: 'text' | 'image' | 'table' | 'line' | 'rect'; band: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number; resizable: boolean; value?: string; visibleIf?: string; fontFamily?: string; fontSize?: number; bold?: boolean; italic?: boolean; align?: 'left' | 'center' | 'right'; valign?: 'top' | 'middle' | 'bottom'; background?: string; borderWidth?: number; borderColor?: string; borderEdges?: ReadonlyArray<'top' | 'right' | 'bottom' | 'left'>; paddingTop?: number; paddingRight?: number; paddingBottom?: number; paddingLeft?: number; tableBind?: string }>>
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
const hasOnly = (value: Record<string, unknown>, keys: readonly string[]) => Object.keys(value).every((key) => keys.includes(key))
export const isEngineRequestId = (value: unknown): value is string => typeof value === 'string' && value.length > 0 && value.length <= MAX_ENGINE_REQUEST_ID_LENGTH && /^[A-Za-z0-9][A-Za-z0-9._:-]*$/.test(value)

const isError = (value: unknown): value is EngineError => isRecord(value) && hasOnly(value, ['code', 'message', 'elementId', 'dataPath']) && typeof value.code === 'string' && value.code.length > 0 && value.code.length <= 96 && typeof value.message === 'string' && value.message.length <= 512 && (value.elementId === undefined || typeof value.elementId === 'string' && value.elementId.length <= MAX_ENGINE_ELEMENT_ID_LENGTH) && (value.dataPath === undefined || typeof value.dataPath === 'string' && value.dataPath.length <= MAX_ENGINE_DATA_PATH_LENGTH)
const isCanvas = (value: unknown): value is CanvasProjection => {
  if (!isRecord(value) || !hasOnly(value, ['width', 'height', 'orientation', 'preset', 'marginTop', 'marginRight', 'marginBottom', 'marginLeft', 'gridIncrement', 'commandWidth', 'commandHeight', 'bands', 'components']) || !['A4', 'Letter', 'custom'].includes(value.preset as string) || (value.orientation !== 'portrait' && value.orientation !== 'landscape')) return false
  const integer = (key: string, positive = false) => typeof value[key] === 'number' && Number.isSafeInteger(value[key]) && (positive ? value[key] > 0 : value[key] >= 0)
  if (!['width', 'height', 'gridIncrement', 'commandWidth', 'commandHeight'].every((key) => integer(key, true)) || !['marginTop', 'marginRight', 'marginBottom', 'marginLeft'].every((key) => integer(key))) return false
  const bands = value.bands
  const components = value.components
  if (!Array.isArray(bands) || bands.length !== 3 || !Array.isArray(components)) return false
  const names = ['pageHeader', 'content', 'pageFooter']
  const page = value as Record<string, number>
  const bandsValid = bands.every((band, index) => {
    if (!isRecord(band) || !hasOnly(band, ['name', 'x', 'y', 'width', 'height']) || band.name !== names[index] || !['x', 'y', 'width', 'height'].every((key) => typeof band[key] === 'number' && Number.isSafeInteger(band[key]))) return false
    const paint = band as Record<string, number>
    if (!(paint.x >= 0 && paint.y >= 0 && paint.width > 0 && paint.height >= 0 && paint.x + paint.width <= page.width && paint.y + paint.height <= page.height)) return false
    if (index > 0) {
      const prior = bands[index - 1] as Record<string, number>
      if (paint.x !== prior.x || paint.width !== prior.width || paint.y !== prior.y + prior.height) return false
    }
    return true
  })
  const componentTypes = ['text', 'image', 'table', 'line', 'rect']
  const bandNames = ['pageHeader', 'content', 'pageFooter']
  if (!bandsValid) return false
  const ids = new Set<string>()
  let priorBand = -1
  return components.every((component) => {
    if (!isRecord(component) || !hasOnly(component, ['id', 'type', 'band', 'x', 'y', 'width', 'height', 'resizable', 'value', 'visibleIf', 'fontFamily', 'fontSize', 'bold', 'italic', 'align', 'valign', 'background', 'borderWidth', 'borderColor', 'borderEdges', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft', 'tableBind']) || typeof component.id !== 'string' || component.id.length === 0 || component.id.length > MAX_ENGINE_ELEMENT_ID_LENGTH || ids.has(component.id) || !componentTypes.includes(component.type as string) || !bandNames.includes(component.band as string) || typeof component.resizable !== 'boolean' || !['x', 'y', 'width', 'height'].every((key) => typeof component[key] === 'number' && Number.isSafeInteger(component[key]) && (component[key] as number) >= 0)) return false
    ids.add(component.id)
    const bandIndex = bandNames.indexOf(component.band as string)
    if (bandIndex < priorBand) return false
    priorBand = bandIndex
    const band = bands[bandIndex] as Record<string, number>
    const box = component as Record<string, number>
    const table = component.type === 'table'
    if (table ? component.resizable || box.height <= 0 : !component.resizable || box.width <= 0 || box.height <= 0) return false
    if (!(box.x + box.width <= band.width && box.y + box.height <= band.height)) return false
    const optionalString = (key: string) => component[key] === undefined || typeof component[key] === 'string' && (component[key] as string).length <= 512
    const optionalLength = (key: string) => component[key] === undefined || typeof component[key] === 'number' && Number.isSafeInteger(component[key]) && (component[key] as number) >= 0
    if (!['value', 'visibleIf', 'fontFamily', 'background', 'borderColor', 'tableBind'].every(optionalString) || !['fontSize', 'borderWidth', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'].every(optionalLength) || (component.bold !== undefined && typeof component.bold !== 'boolean') || (component.italic !== undefined && typeof component.italic !== 'boolean')) return false
    if (component.align !== undefined && !['left', 'center', 'right'].includes(component.align as string) || component.valign !== undefined && !['top', 'middle', 'bottom'].includes(component.valign as string)) return false
    if (component.borderEdges !== undefined && (!Array.isArray(component.borderEdges) || component.borderEdges.length === 0 || component.borderEdges.some((edge) => !['top', 'right', 'bottom', 'left'].includes(edge)))) return false
    if (component.type !== 'text' && component.value !== undefined) return false
    if (component.type !== 'table' && component.tableBind !== undefined) return false
    if (!['text', 'table'].includes(component.type as string) && ['fontFamily', 'fontSize', 'bold', 'italic', 'align', 'valign'].some((key) => component[key] !== undefined)) return false
    return true
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
