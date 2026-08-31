export const ENGINE_PROTOCOL_VERSION = 1 as const

export type EngineOperation = 'initialize' | 'load' | 'snapshot' | 'parameter-references' | 'table-columns' | 'validate' | 'serialize' | 'command' | 'undo' | 'redo' | 'identity' | 'render' | 'asset'

export const MAX_ENGINE_REQUEST_ID_LENGTH = 128
export const MAX_ENGINE_PAYLOAD_BYTES = 8 * 1024 * 1024
export const MAX_ENGINE_RENDER_PDF_BYTES = 32 * 1024 * 1024
export const MAX_ENGINE_DIAGNOSTICS = 256
export const MAX_ENGINE_ELEMENT_ID_LENGTH = 128
export const MAX_ENGINE_DATA_PATH_LENGTH = 256
export const MAX_ENGINE_BINDING_LENGTH = 256
export const MAX_ENGINE_PARAMETER_REFERENCES = 128
export const MAX_ENGINE_PARAMETER_NAME_LENGTH = 128
// The same bound Go projects the document's declared font chains under.
export const MAX_ENGINE_FONT_FAMILIES = 256
// And the bound ONE chain's entry list is projected under. Story 8.1 put the
// entries themselves on the wire, so the per-chain array needs the same
// treatment the chain list already had.
export const MAX_ENGINE_FONT_CHAIN_ENTRIES = 64
// A CHANNEL BACKSTOP, NOT A MIRROR. Go declares no maximum number of content
// windows — internal/layout bounds the count only by the column-item count —
// so this number is deliberately NOT in the pair list below: there is nothing
// on the Go side for it to drift against. It exists for the reason
// MAX_ENGINE_DIAGNOSTICS does, to keep an absurd array from being iterated,
// and it is set orders of magnitude above anything the projection produces
// because the cost of it biting is severe and silent: a rejected field
// discards the WHOLE snapshot and blanks the canvas. Epic 7's narrative
// target is forty pages; the canvas's own sheet budget is 120.
export const MAX_ENGINE_CONTENT_WINDOWS = 100_000

// ---------------------------------------------------------------------------
// THE CANVAS PROJECTION BOUNDS, MIRRORED FROM folio-go/page_setup.go.
//
// There is no shared source and no codegen: these are hand-copied, which is
// the drift pattern in pure form — the Go side can be raised and this side
// will silently keep rejecting, blanking the projection with no error anyone
// can attribute (D-7.4.5). They are hoisted out of the validators and named
// after their Go counterparts so `engine-bounds-mirror.test.ts` can read both
// files and assert the pairs are equal. Change one, change the other, in the
// same commit.
//
// A UNIT MISMATCH IS BUILT INTO THE TWO STRING BOUNDS, and it is recorded
// rather than "fixed": Go counts BYTES (`len()`), these count UTF-16 CODE
// UNITS (`.length`). For non-ASCII this side is the more permissive of the
// pair, so the Go side refuses first and nothing unrepresentable arrives.
// The tie assertion compares LITERALS, not quantities, and says so.
//
// maxCanvasBodyText — the body-text channel backstop (bytes/code units).
export const MAX_CANVAS_BODY_TEXT = 1048576
// maxCanvasBodyTextLines — 40 pages × 48 lines per A4 page at 11pt.
export const MAX_CANVAS_BODY_TEXT_LINES = 1920
// maxCanvasBodyTextFragments — CUMULATIVE across the whole component, which
// is the quantity counted below. Go's own maxCanvasTextFragments bounds one
// LINE and is deliberately NOT mirrored here; the two are different
// quantities, and pairing them would be a false tie.
export const MAX_CANVAS_BODY_TEXT_FRAGMENTS = 65536
// maxCanvasPropertyString — identifiers, colours and expressions only. Body
// text no longer shares it on either side of the channel (DW-25).
export const MAX_CANVAS_PROPERTY_STRING = 512
// THE FIFTH HAND-COPIED CROSS-LANGUAGE BOUND, and the only one that does not
// come from `page_setup.go`: `template.MinLineSpacingThousandths` and
// `template.MaxLineSpacingThousandths` (folio-go/internal/template/
// linespacing.go), which Story 7.4 projects across the channel for the first
// time. The Go comment there calls the maximum "A STATED SANITY CEILING, NOT
// A DERIVED SAFETY BOUND" — i.e. a number somebody will one day adjust — and
// a raised ceiling with these literals left behind would make `parseInbound`
// drop every snapshot of such a document silently, with no canvas and no
// error. So they are named here and tied to the Go declarations by
// engine-bounds-mirror.test.ts alongside the other four.
export const MIN_LINE_SPACING_THOUSANDTHS = 1
export const MAX_LINE_SPACING_THOUSANDTHS = 1000000

// THE FIFTH MIRROR, and the first one that is a PREDICATE rather than a
// number: which bands cap a component vertically.
//
// DW-25 closed the four size caps above. Band containment is a different
// invariant that merely happens to live in the same file, and an audit closes
// only what it measured — so the standing obligation is widened here from
// "the size caps move together" to: ANY invariant duplicated across the
// Go/TypeScript boundary moves in ONE commit, with a test that reads both
// sides. `folio-go/component_commands.go` declares this same list under this
// same name and `engine-bounds-mirror.test.ts` reads both files.
//
// The content band is absent by MEANING. A page header and a page footer
// repeat on every page, so each is exactly one page tall; the content band is
// a COLUMN that Go's internal/layout slices into page-height windows, so a
// component below the foot of page one is on page two, not outside the
// document. What a stale copy of this list costs is not a hidden component:
// `isCanvas` returning false makes `parseInbound` return undefined, which
// terminates the worker, rejects every in-flight request and leaves the
// canvas blank — with no element id and no attributable error.
export const BANDS_CAPPING_VERTICALLY = ['pageHeader', 'pageFooter']

export type EngineError = Readonly<{
  code: string
  message: string
  elementId?: string
  dataPath?: string
}>

export type RenderPayload = Readonly<{ template: ArrayBuffer; data: ArrayBuffer; params: ArrayBuffer }>
export type IdentityPayload = Readonly<{ data: ArrayBuffer; params: ArrayBuffer }>
export type EngineDiagnostic = Readonly<{ severity: 'warning'; code: string; elementId: string; dataPath: string; message: string }>
export type PreviewEvidence = Readonly<{ revision: number; identity: string; pdfSha256?: string; diagnostics?: ReadonlyArray<EngineDiagnostic> }>
export type ParameterReferences = Readonly<{ revision: number; names: ReadonlyArray<string> }>
export type TableColumn = Readonly<{ id: string; header: string; width: number; align: 'left' | 'center' | 'right'; binding: string; rowField: string; rowFieldEditable: boolean; footer: '' | 'sum' | 'avg' | 'count'; footerOf: string; footerFormat: string }>
export type TableColumns = Readonly<{ revision: number; table: Readonly<{ tableId: string; collection: string; alias: string; columns: ReadonlyArray<TableColumn> }> }>

// Opaque bytes/JSON are deliberately the only document-bearing values on this
// boundary. These types describe transport, not the .folio file format.
export type EngineRequest = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'request'
  requestId: string
  operation: EngineOperation
  payload?: ArrayBuffer | RenderPayload | IdentityPayload
}>

export type EngineSnapshot = Readonly<{
  documentState: 'empty' | 'loaded'
  revision: number
  byteLength: number
	canUndo?: boolean
	canRedo?: boolean
	canvas?: CanvasProjection
}>

// This is paint-only output from Go, not a .folio page model. Values are
// millipoints and are never used to derive a browser document layout.
export type CanvasProjection = Readonly<{
	width: number; height: number; orientation: 'portrait' | 'landscape'; preset: 'A4' | 'Letter' | 'custom'
	marginTop: number; marginRight: number; marginBottom: number; marginLeft: number; gridIncrement: number; commandWidth: number; commandHeight: number
	// contentWindowHeight is ONE page's worth of content column, and
	// contentWindowCount is how many of those windows the column occupies —
	// both from Go, neither derived here. The count is a claim about the
	// column as the ENGINE currently paints it, and a floor rather than a
	// prediction wherever a bound table is involved: the canvas has no data,
	// so a table contributes its header and none of its rows.
	contentWindowHeight: number; contentWindowCount: number
	// contentWindowOrigins is where each of those windows BEGINS, in the
	// content column's own band-relative frame — one entry per window,
	// origins[0] === 0, strictly increasing. It comes from Go's own
	// PageAssignment.Shift and is NEVER the window height multiplied by an
	// index: that
	// closed form is the spelling internal/layout/paginate.go forbids by
	// name, and it is wrong by 110 millipoints per window on a column of
	// round 728pt spacing and by nine whole windows on a column with a
	// declared gap. contentWindowCountIsExact is Go saying the count can be
	// TRUSTED — false wherever a registered cause applies: a bound table, a
	// pagination that degraded, text that could not be shaped, or an element
	// whose visibility depends on data. Its sense is deliberately this way
	// round so that its zero value, false, is the SAFE claim; direction —
	// whether the true number is higher or lower — is deliberately not
	// carried, because neither side is safe to act on. Both are engine facts,
	// and neither is a rule this side gets to restate.
	contentWindowOrigins: ReadonlyArray<number>; contentWindowCountIsExact: boolean
	// fontFamilies is the closed set style.fontFamily may name in THIS
	// document, from Go, sorted; defaultFontSize is the size the producer
	// draws an element that commits none at. Neither is restated here.
	fontFamilies: ReadonlyArray<string>; defaultFontSize: number
	// fontChains is the SAME set of chains, with the ordered faces behind each
	// name: fontChains.map(c => c.name) is fontFamilies, entry for entry, and
	// the validator asserts it rather than trusting it. Entry order is the
	// document's own authored order and is never re-sorted here.
	fontChains: ReadonlyArray<Readonly<{ name: string; entries: ReadonlyArray<string> }>>
	bands: ReadonlyArray<Readonly<{ name: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number }>>
	components: ReadonlyArray<Readonly<{ id: string; type: 'text' | 'image' | 'table' | 'line' | 'rect'; band: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number; resizable: boolean; value?: string; binding?: string; visibleIf?: string; fontFamily?: string; fontSize?: number; lineSpacing?: number; bold?: boolean; italic?: boolean; align?: 'left' | 'center' | 'right' | 'justify'; valign?: 'top' | 'middle' | 'bottom'; color?: string; background?: string; borderWidth?: number; borderColor?: string; borderEdges?: ReadonlyArray<'top' | 'right' | 'bottom' | 'left'>; paddingTop?: number; paddingRight?: number; paddingBottom?: number; paddingLeft?: number; tableBind?: string; textPaint?: Readonly<{ overflow: boolean; truncated: boolean; lines: ReadonlyArray<Readonly<{ top: number; baseline: number; advance: number; width: number; fragments: ReadonlyArray<Readonly<{ text: string; x: number }>> }>> }>; image?: Readonly<{ mediaType: string; assetKey: string; width: number; height: number; drawX: number; drawY: number; drawWidth: number; drawHeight: number }>; imageUnavailable?: 'missing' | 'undecodable' }>>
}>

export type EngineSuccess = Readonly<{
  protocolVersion: typeof ENGINE_PROTOCOL_VERSION
  kind: 'response'
  requestId: string
  ok: true
  snapshot: EngineSnapshot
  bytes?: ArrayBuffer
	preview?: PreviewEvidence
	parameterReferences?: ParameterReferences
	tableColumns?: TableColumns
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
const isRenderPayload = (value: unknown): value is RenderPayload => isRecord(value) && hasExactKeys(value, ['template', 'data', 'params']) && ['template', 'data', 'params'].every((key) => isArrayBuffer(value[key]) && value[key].byteLength > 0 && value[key].byteLength <= MAX_ENGINE_PAYLOAD_BYTES)
const isIdentityPayload = (value: unknown): value is IdentityPayload => isRecord(value) && hasExactKeys(value, ['data', 'params']) && ['data', 'params'].every((key) => isArrayBuffer(value[key]) && value[key].byteLength > 0 && value[key].byteLength <= MAX_ENGINE_PAYLOAD_BYTES)
const hasOnly = (value: Record<string, unknown>, keys: readonly string[]) => Object.keys(value).every((key) => keys.includes(key))
const hasExactKeys = (value: Record<string, unknown>, keys: readonly string[]) => Object.keys(value).length === keys.length && keys.every((key) => Object.prototype.hasOwnProperty.call(value, key))
export const isEngineRequestId = (value: unknown): value is string => typeof value === 'string' && value.length > 0 && value.length <= MAX_ENGINE_REQUEST_ID_LENGTH && /^[A-Za-z0-9][A-Za-z0-9._:-]*$/.test(value)

// A supplied provenance field is a producer fact, not a display hint.  Empty
// values have no useful meaning in the closed contract, so reject them at the
// boundary instead of accepting and silently dropping them later by truthiness.
const isError = (value: unknown): value is EngineError => isRecord(value) && hasOnly(value, ['code', 'message', 'elementId', 'dataPath']) && typeof value.code === 'string' && value.code.length > 0 && value.code.length <= 96 && typeof value.message === 'string' && value.message.length > 0 && value.message.length <= 512 && (value.elementId === undefined || typeof value.elementId === 'string' && value.elementId.length > 0 && value.elementId.length <= MAX_ENGINE_ELEMENT_ID_LENGTH) && (value.dataPath === undefined || typeof value.dataPath === 'string' && value.dataPath.length > 0 && value.dataPath.length <= MAX_ENGINE_DATA_PATH_LENGTH)
const isDiagnostic = (value: unknown): value is EngineDiagnostic => isRecord(value) && hasExactKeys(value, ['severity', 'code', 'elementId', 'dataPath', 'message']) && value.severity === 'warning' && typeof value.code === 'string' && value.code.length > 0 && value.code.length <= 96 && typeof value.elementId === 'string' && value.elementId.length <= MAX_ENGINE_ELEMENT_ID_LENGTH && typeof value.dataPath === 'string' && value.dataPath.length <= MAX_ENGINE_DATA_PATH_LENGTH && typeof value.message === 'string' && value.message.length <= 512
const isPreview = (value: unknown): value is PreviewEvidence => isRecord(value) && hasOnly(value, ['revision', 'identity', 'pdfSha256', 'diagnostics']) && typeof value.revision === 'number' && Number.isSafeInteger(value.revision) && value.revision >= 0 && typeof value.identity === 'string' && /^[a-f0-9]{64}$/.test(value.identity) && ((value.pdfSha256 === undefined && value.diagnostics === undefined) || (typeof value.pdfSha256 === 'string' && /^[a-f0-9]{64}$/.test(value.pdfSha256) && Array.isArray(value.diagnostics) && value.diagnostics.length <= MAX_ENGINE_DIAGNOSTICS && value.diagnostics.every(isDiagnostic)))
const isParameterReferences = (value: unknown): value is ParameterReferences => isRecord(value) && hasExactKeys(value, ['revision', 'names']) && typeof value.revision === 'number' && Number.isSafeInteger(value.revision) && value.revision >= 0 && Array.isArray(value.names) && value.names.length <= MAX_ENGINE_PARAMETER_REFERENCES && value.names.every((name) => typeof name === 'string' && name.length > 0 && name.length <= MAX_ENGINE_PARAMETER_NAME_LENGTH && /^[A-Za-z_][A-Za-z0-9_]*$/.test(name)) && new Set(value.names).size === value.names.length && value.names.every((name, index, names) => index === 0 || names[index - 1]! < name)
const isTableColumns = (value: unknown): value is TableColumns => {
  if (!isRecord(value) || !hasExactKeys(value, ['revision', 'table']) || typeof value.revision !== 'number' || !Number.isSafeInteger(value.revision) || value.revision < 0 || !isRecord(value.table) || !hasExactKeys(value.table, ['tableId', 'collection', 'alias', 'columns'])) return false
  const table = value.table
  return typeof table.tableId === 'string' && table.tableId.length > 0 && table.tableId.length <= MAX_ENGINE_ELEMENT_ID_LENGTH && typeof table.collection === 'string' && table.collection.length > 0 && table.collection.length <= MAX_ENGINE_BINDING_LENGTH && typeof table.alias === 'string' && table.alias.length > 0 && table.alias.length <= 64 && Array.isArray(table.columns) && table.columns.length <= 128 && table.columns.every((column) => isRecord(column) && hasExactKeys(column, ['id', 'header', 'width', 'align', 'binding', 'rowField', 'rowFieldEditable', 'footer', 'footerOf', 'footerFormat']) && typeof column.id === 'string' && column.id.length > 0 && column.id.length <= MAX_ENGINE_ELEMENT_ID_LENGTH && typeof column.header === 'string' && column.header.length <= 256 && typeof column.width === 'number' && Number.isSafeInteger(column.width) && column.width > 0 && ['left', 'center', 'right'].includes(column.align as string) && typeof column.binding === 'string' && column.binding.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.rowField === 'string' && column.rowField.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.rowFieldEditable === 'boolean' && ['','sum','avg','count'].includes(column.footer as string) && typeof column.footerOf === 'string' && column.footerOf.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.footerFormat === 'string' && column.footerFormat.length <= 256) && new Set(table.columns.map((item) => (item as Record<string, unknown>).id)).size === table.columns.length
}
const isCanvas = (value: unknown): value is CanvasProjection => {
  if (!isRecord(value) || !hasOnly(value, ['width', 'height', 'orientation', 'preset', 'marginTop', 'marginRight', 'marginBottom', 'marginLeft', 'gridIncrement', 'commandWidth', 'commandHeight', 'fontFamilies', 'fontChains', 'defaultFontSize', 'contentWindowHeight', 'contentWindowCount', 'contentWindowOrigins', 'contentWindowCountIsExact', 'bands', 'components']) || !['A4', 'Letter', 'custom'].includes(value.preset as string) || (value.orientation !== 'portrait' && value.orientation !== 'landscape')) return false
  const integer = (key: string, positive = false) => typeof value[key] === 'number' && Number.isSafeInteger(value[key]) && (positive ? value[key] > 0 : value[key] >= 0)
  if (!['width', 'height', 'gridIncrement', 'commandWidth', 'commandHeight', 'defaultFontSize', 'contentWindowHeight', 'contentWindowCount'].every((key) => integer(key, true)) || !['marginTop', 'marginRight', 'marginBottom', 'marginLeft'].every((key) => integer(key))) return false
  // The declared font chain names, as Go sorted them: bounded in count and
  // length like every other list on this projection, unique, and in the order
  // Go sent so the browser never re-sorts an engine-owned set.
  if (!Array.isArray(value.fontFamilies) || value.fontFamilies.length > MAX_ENGINE_FONT_FAMILIES || !value.fontFamilies.every((name) => typeof name === 'string' && name.length > 0 && name.length <= MAX_CANVAS_PROPERTY_STRING) || value.fontFamilies.some((name, index, names) => index > 0 && (names[index - 1] as string) >= (name as string))) return false
  // The chains those names stand for. Bounded in count and in per-chain entry
  // count, every entry a non-empty bounded string, no chain empty — an empty
  // chain is not one Go projects, because it is not one style.fontFamily may
  // name. The last clause is the cross-check the two lists exist to give each
  // other: Go builds fontFamilies FROM fontChains, so any disagreement here is
  // a channel fault and the snapshot is not trusted.
  const chains = value.fontChains
  if (!Array.isArray(chains) || chains.length !== value.fontFamilies.length) return false
  if (!chains.every((chain, index) => isRecord(chain) && hasExactKeys(chain, ['name', 'entries']) && chain.name === (value.fontFamilies as ReadonlyArray<unknown>)[index] && Array.isArray(chain.entries) && chain.entries.length > 0 && chain.entries.length <= MAX_ENGINE_FONT_CHAIN_ENTRIES && chain.entries.every((face) => typeof face === 'string' && face.length > 0 && face.length <= MAX_CANVAS_PROPERTY_STRING))) return false
  // The window origins, in the same shape: bounded in count, every entry a
  // safe non-negative integer, and in the order and at the length Go's own
  // pagination fixes. `hasOnly` is a SUBSET check, so an origins key Go
  // simply failed to send is caught HERE and nowhere else — as is a `nil`
  // slice, which marshals to null and is not an array.
  const origins = value.contentWindowOrigins
  if (!Array.isArray(origins) || origins.length === 0 || origins.length > MAX_ENGINE_CONTENT_WINDOWS || origins.length !== value.contentWindowCount) return false
  if (!origins.every((origin) => typeof origin === 'number' && Number.isSafeInteger(origin) && origin >= 0) || origins[0] !== 0 || origins.some((origin, index) => index > 0 && (origins[index - 1] as number) >= (origin as number))) return false
  if (typeof value.contentWindowCountIsExact !== 'boolean') return false
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
	if (!isRecord(component) || !hasOnly(component, ['id', 'type', 'band', 'x', 'y', 'width', 'height', 'resizable', 'value', 'binding', 'visibleIf', 'fontFamily', 'fontSize', 'lineSpacing', 'bold', 'italic', 'align', 'valign', 'color', 'background', 'borderWidth', 'borderColor', 'borderEdges', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft', 'tableBind', 'textPaint', 'image', 'imageUnavailable']) || typeof component.id !== 'string' || component.id.length === 0 || component.id.length > MAX_ENGINE_ELEMENT_ID_LENGTH || ids.has(component.id) || !componentTypes.includes(component.type as string) || !bandNames.includes(component.band as string) || typeof component.resizable !== 'boolean' || !['x', 'y', 'width', 'height'].every((key) => typeof component[key] === 'number' && Number.isSafeInteger(component[key]) && (component[key] as number) >= 0)) return false
    ids.add(component.id)
    const bandIndex = bandNames.indexOf(component.band as string)
    if (bandIndex < priorBand) return false
    priorBand = bandIndex
    const band = bands[bandIndex] as Record<string, number>
    const box = component as Record<string, number>
    const table = component.type === 'table'
    if (table ? component.resizable || box.height <= 0 : !component.resizable || box.width <= 0 || box.height <= 0) return false
    // THE HORIZONTAL CAP IS UNIVERSAL; the vertical one is not. A band is as
    // wide as the printable page and nothing may hang off its side, in any
    // band. The vertical cap belongs only to the bands that HAVE a capacity —
    // see BANDS_CAPPING_VERTICALLY, which Go's containComponent mirrors.
    if (!(box.x + box.width <= band.width)) return false
    if (BANDS_CAPPING_VERTICALLY.includes(component.band as string) && !(box.y + box.height <= band.height)) return false
    // THE FOURTH HAND-COPIED MIRROR (DW-25). This one predicate used to cap
    // `value` — the document's BODY TEXT — at the same 512 as seven
    // identifier and colour keys: maxCanvasPropertyString's two-jobs
    // conflation, reproduced exactly on the browser side. Splitting Go's
    // constant without splitting this one would have changed nothing
    // observable: the browser would go on dropping the whole response at 512
    // bytes of clause text, with no attributable error.
    const boundedString = (key: string, limit: number) => component[key] === undefined || typeof component[key] === 'string' && (component[key] as string).length <= limit
    const optionalString = (key: string) => boundedString(key, MAX_CANVAS_PROPERTY_STRING)
    const optionalLength = (key: string) => component[key] === undefined || typeof component[key] === 'number' && Number.isSafeInteger(component[key]) && (component[key] as number) >= 0
	if (!boundedString('value', MAX_CANVAS_BODY_TEXT) || !['binding', 'visibleIf', 'fontFamily', 'color', 'background', 'borderColor', 'tableBind'].every(optionalString) || !['fontSize', 'borderWidth', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'].every(optionalLength) || (component.bold !== undefined && typeof component.bold !== 'boolean') || (component.italic !== undefined && typeof component.italic !== 'boolean')) return false
	if (component.binding !== undefined && (typeof component.binding !== 'string' || component.binding.length === 0 || component.binding.length > MAX_ENGINE_BINDING_LENGTH)) return false
    // Story 7.3 / FR47: the COMPONENT alignment vocabulary admits
    // `justify`; the COLUMN one (isTableColumns, above) deliberately does
    // not. This validator GATES the projection — an unrecognised value
    // drops the whole response and blanks the canvas — so a justified
    // document would otherwise show as nothing at all rather than as
    // itself. The inspector control that OFFERS the choice is Story 7.4's;
    // admitting the value is not offering it.
    if (component.align !== undefined && !['left', 'center', 'right', 'justify'].includes(component.align as string) || component.valign !== undefined && !['top', 'middle', 'bottom'].includes(component.valign as string)) return false
    if (component.borderEdges !== undefined && (!Array.isArray(component.borderEdges) || component.borderEdges.length === 0 || component.borderEdges.some((edge) => !['top', 'right', 'bottom', 'left'].includes(edge)))) return false
    // style.lineSpacing, projected for the first time by Story 7.4: a
    // dimensionless ratio in THOUSANDTHS, positive, and bounded by the same
    // range the engine's one validator enforces at load and on the property
    // command alike (template.MinLineSpacingThousandths ..
    // MaxLineSpacingThousandths, D-7.2.3). Admitting the value is not
    // adjudicating it — a value Go committed is a value Go already ruled on.
    if (component.lineSpacing !== undefined && (typeof component.lineSpacing !== 'number' || !Number.isSafeInteger(component.lineSpacing) || component.lineSpacing < MIN_LINE_SPACING_THOUSANDTHS || component.lineSpacing > MAX_LINE_SPACING_THOUSANDTHS)) return false
	if (component.type !== 'text' && component.value !== undefined) return false
	if (component.type !== 'text' && component.binding !== undefined) return false
	if (component.type !== 'table' && component.tableBind !== undefined) return false
	if (!['text', 'table'].includes(component.type as string) && ['fontFamily', 'fontSize', 'bold', 'italic', 'align', 'valign'].some((key) => component[key] !== undefined)) return false
	if (!isTextPaint(component.textPaint, box)) return false
	if (component.type === 'text' ? component.textPaint === undefined : component.textPaint !== undefined) return false
	if (!isImagePaint(component.image, box)) return false
	if (component.type !== 'image' && component.image !== undefined) return false
	// Finding 9 (review of 2026-08-29): the bounded, enumerated reason
	// discriminant Go emits alongside an absent image paint — only legal
	// for an 'image' component whose image paint is itself absent (the
	// two are the same "one Go-side signal", D-5.13.2), never alongside a
	// present paint and never for a non-image component.
	if (component.imageUnavailable !== undefined && (component.type !== 'image' || component.image !== undefined || !['missing', 'undecodable'].includes(component.imageUnavailable as string))) return false
	return true
  })
}

// isImagePaint admits Story 5.13's optional per-component paint-only image
// projection. Absence is always legal for an 'image' component (D-5.13.2:
// "absence, not zero" — an unrecognised or undecodable asset simply has no
// paint). When present, every field must be a bounded, positive, in-box
// value: the draw rectangle is asserted to sit INSIDE the component's own
// box, exactly like the fit-and-centre invariant it is meant to project.
const isImagePaint = (value: unknown, box: Record<string, number>): boolean => {
  if (value === undefined) return true
  if (!isRecord(value) || !hasOnly(value, ['mediaType', 'assetKey', 'width', 'height', 'drawX', 'drawY', 'drawWidth', 'drawHeight'])) return false
  if (typeof value.mediaType !== 'string' || value.mediaType.length === 0 || value.mediaType.length > 128) return false
  // Finding 12 (review of 2026-08-29): D-5.13.2's amendment settled the wire
  // key as the FULL 64-hex digest — a per-key bytes request cannot address
  // an asset by a prefix (Go's isAssetKeyShape, asset_bytes.go, requires
  // exactly 64). This admission previously accepted 1..64, so a truncated
  // key passed straight through to a per-key fetch that could never
  // succeed (Finding 13's permanent "Loading image…").
  if (typeof value.assetKey !== 'string' || value.assetKey.length !== 64 || !/^[a-f0-9]{64}$/.test(value.assetKey)) return false
  const integer = (key: string, positive = false) => typeof (value as Record<string, unknown>)[key] === 'number' && Number.isSafeInteger((value as Record<string, unknown>)[key]) && (positive ? ((value as Record<string, unknown>)[key] as number) > 0 : ((value as Record<string, unknown>)[key] as number) >= 0)
  if (!['width', 'height', 'drawWidth', 'drawHeight'].every((key) => integer(key, true)) || !['drawX', 'drawY'].every((key) => integer(key))) return false
  const paint = value as Record<string, number>
  return paint.drawX >= box.x && paint.drawY >= box.y && paint.drawX + paint.drawWidth <= box.x + box.width && paint.drawY + paint.drawHeight <= box.y + box.height
}

// isTextPaint admits the engine's own honest measurement and checks only
// what the JS boundary can genuinely go wrong at. It deliberately does
// NOT check `paint.baseline > paint.top + paint.advance` any more
// (Story 7.2, D-7.2.2): the engine emits `baseline = top + FirstBaseline`
// while `advance` is the SCALED value, so that clause reduced to
// `FirstBaseline <= Advance` — an ENGINE invariant restated on the
// browser's side of the channel, and one `style.lineSpacing`
// deliberately dissolves.
//
// `FirstBaseline > Advance` means one line's baseline sits below the
// next line's top: the line boxes overlap. That IS tight leading, it is
// what the PDF draws, and refusing it here failed one line, then
// isCanvas, then isSnapshot, and blanked the WHOLE projection. AD-17
// says the canvas takes every text metric FROM the engine; the browser
// adjudicating them was that invariant inverted, not enforced.
//
// The real invariants all survive on the line below and must stay:
// `paint.advance <= 0`, `paint.baseline < paint.top` (FirstBaseline is
// an ascent clamped at zero, and lineSpacing scales only Advance), the
// Number.isSafeInteger checks (the actual JS-boundary concern), and
// `paint.top < priorTop + priorAdvance` — which is `originY+i·A <
// originY+i·A`, false for any positive advance, so it does not become
// the next cliff.
const isTextPaint = (value: unknown, component: Record<string, number>): boolean => {
  if (value === undefined) return true
  // `truncated` is required exactly as `overflow` is: Go emits both
  // unconditionally, and a paint arriving without it is a producer that has
  // drifted from this contract, not an older one to be tolerated.
  if (!isRecord(value) || !hasOnly(value, ['overflow', 'truncated', 'lines']) || typeof value.overflow !== 'boolean' || typeof value.truncated !== 'boolean' || !Array.isArray(value.lines) || value.lines.length > MAX_CANVAS_BODY_TEXT_LINES) return false
  let priorTop = -1
  let priorAdvance = 0
  // CUMULATIVE across every line of the component, never reset — which is a
  // different quantity from Go's per-line maxCanvasTextFragments. Go carries
  // its own cumulative counter (maxCanvasBodyTextFragments) precisely so it
  // never emits a projection this line would discard.
  let fragments = 0
  return value.lines.every((line) => {
    if (!isRecord(line) || !hasOnly(line, ['top', 'baseline', 'advance', 'width', 'fragments']) || !['top', 'baseline', 'advance', 'width'].every((key) => typeof line[key] === 'number' && Number.isSafeInteger(line[key]))) return false
    const paint = line as Record<string, number>
    if (paint.top < component.y || paint.baseline < paint.top || paint.advance <= 0 || paint.width < 0 || (priorTop >= 0 && paint.top < priorTop + priorAdvance) || (!value.overflow && paint.width > component.width) || !Array.isArray(line.fragments)) return false
    priorTop = paint.top
    priorAdvance = paint.advance
    return line.fragments.every((fragment) => {
      fragments++
      return fragments <= MAX_CANVAS_BODY_TEXT_FRAGMENTS && isRecord(fragment) && hasOnly(fragment, ['text', 'x']) && typeof fragment.text === 'string' && fragment.text.length > 0 && fragment.text.length <= MAX_CANVAS_BODY_TEXT && typeof fragment.x === 'number' && Number.isSafeInteger(fragment.x) && fragment.x >= component.x && fragment.x <= component.x + Math.max(paint.width, component.width)
    })
  })
}
const isSnapshot = (value: unknown): value is EngineSnapshot => isRecord(value) && hasOnly(value, ['documentState', 'revision', 'byteLength', 'canUndo', 'canRedo', 'canvas']) && (value.documentState === 'empty' || value.documentState === 'loaded') && typeof value.revision === 'number' && Number.isSafeInteger(value.revision) && value.revision >= 0 && typeof value.byteLength === 'number' && Number.isSafeInteger(value.byteLength) && value.byteLength >= 0 && (value.canUndo === undefined || typeof value.canUndo === 'boolean') && (value.canRedo === undefined || typeof value.canRedo === 'boolean') && (value.canvas === undefined || isCanvas(value.canvas))

export function requestCorrelationId(value: unknown): string | undefined {
  return isRecord(value) && isEngineRequestId(value.requestId) ? value.requestId : undefined
}

export function parseRequest(value: unknown): EngineRequest | undefined {
  if (!isRecord(value) || !hasOnly(value, ['protocolVersion', 'kind', 'requestId', 'operation', 'payload']) || value.protocolVersion !== ENGINE_PROTOCOL_VERSION || value.kind !== 'request' || !isEngineRequestId(value.requestId)) return undefined
	if (!['initialize', 'load', 'snapshot', 'parameter-references', 'table-columns', 'validate', 'serialize', 'command', 'undo', 'redo', 'identity', 'render', 'asset'].includes(value.operation as string)) return undefined
  if (value.payload !== undefined && (!isArrayBuffer(value.payload) || value.payload.byteLength > MAX_ENGINE_PAYLOAD_BYTES) && !(value.operation === 'render' && isRenderPayload(value.payload)) && !(value.operation === 'identity' && isIdentityPayload(value.payload))) return undefined
	const needsPayload = value.operation === 'initialize' || value.operation === 'load' || value.operation === 'command' || value.operation === 'table-columns' || value.operation === 'asset'
  if (value.operation === 'render' ? !isRenderPayload(value.payload) : value.operation === 'identity' ? !isIdentityPayload(value.payload) : needsPayload !== (value.payload !== undefined)) return undefined
  return value as EngineRequest
}

export function parseInbound(value: unknown): EngineInbound | undefined {
  if (!isRecord(value) || value.protocolVersion !== ENGINE_PROTOCOL_VERSION || typeof value.kind !== 'string') return undefined
  if (value.kind === 'lifecycle') {
    if ((hasExactKeys(value, ['protocolVersion', 'kind', 'state']) && value.state === 'ready') || (hasExactKeys(value, ['protocolVersion', 'kind', 'state', 'error']) && value.state === 'failed' && isError(value.error))) return value as EngineLifecycle
    return undefined
  }
  if (value.kind !== 'response' || !isEngineRequestId(value.requestId) || typeof value.ok !== 'boolean') return undefined
	if (value.ok && hasOnly(value, ['protocolVersion', 'kind', 'requestId', 'ok', 'snapshot', 'bytes', 'preview', 'parameterReferences', 'tableColumns']) && isSnapshot(value.snapshot) && (value.bytes === undefined || isArrayBuffer(value.bytes) && value.bytes.byteLength <= MAX_ENGINE_RENDER_PDF_BYTES) && (value.preview === undefined || isPreview(value.preview)) && (value.parameterReferences === undefined || isParameterReferences(value.parameterReferences)) && (value.tableColumns === undefined || isTableColumns(value.tableColumns)) && (value.preview === undefined || value.preview.revision === value.snapshot.revision) && (value.parameterReferences === undefined || value.parameterReferences.revision === value.snapshot.revision) && (value.tableColumns === undefined || value.tableColumns.revision === value.snapshot.revision) && (value.preview?.pdfSha256 === undefined || value.bytes !== undefined)) return value as EngineSuccess
  if (!value.ok && hasExactKeys(value, ['protocolVersion', 'kind', 'requestId', 'ok', 'error']) && isError(value.error)) return value as EngineFailure
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
