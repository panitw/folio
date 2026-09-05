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

// AD-12's CLOSED LOCALE SET, ONCE — the same discipline, for the same reason,
// on a different invariant.
//
// Go spells each of these four tags EXACTLY ONCE, as the right-hand side of a
// named constant in `folio-go/internal/template/locale.go`, and builds both
// `LocaleTags` and `closedLocales` from those constants so the set cannot drift
// by a spelling mistake. This is the browser's single spelling of the same set,
// tied to Go's by `engine-bounds-mirror.test.ts` on the same idiom the band
// list uses — it resolves Go's named constants before comparing, so
// `[]string{LocaleEN, LocaleTH, …}` and `['en', 'th', …]` are compared as the
// same CLAIM rather than as the same text.
//
// EVERYTHING ON THIS SIDE READS THIS ARRAY: the isCanvas guard's typed clause,
// the panel's <option> list, and the command factory's parameter type. A tag
// written out anywhere else would be a fourth spelling standing outside that
// census, which is the only place a stale copy can hide — and the cost of a
// stale copy here is the cost of every other one: isCanvas returns false,
// parseInbound returns undefined, the worker is terminated and the canvas is
// permanently blank with nothing to attribute it to.
//
// A FIFTH TAG IS NOT ADDED HERE. Widening the set is a MAJOR change under
// folio-format.md's MINOR-increment rule — every existing library validates it as a load error —
// so it is Go's decision and an owner's, and it reaches this file through the
// mirror rather than by an edit that starts here.
export const LOCALE_TAGS = ['en', 'th', 'zh-Hans', 'ja'] as const
export type LocaleTag = (typeof LOCALE_TAGS)[number]

// THE SAME LIST, ONCE AS A TYPE AND ONCE AS AN ARRAY A CALLER MAY ITERATE —
// and neither of them is a second copy of it.
//
// Story 12.1 first shipped `settableBands` in App.tsx and `SettableBand` in
// band-height-command.ts, each spelling the two names out again. That made four
// and five copies of a list whose whole safety property is that
// engine-bounds-mirror.test.ts reads it on BOTH sides of the Go/TypeScript
// boundary and refuses to let it drift: two of the five were outside that
// census, which is the only place a stale copy can hide.
//
// CAPPING_BANDS is the SAME ARRAY OBJECT, narrowed to the element type; the
// union is taken out of the projection's own band-name union, which
// canvas_projection_wire_test.go pins against Go, minus the one band that has
// no height to cap with. Both are tied back to Go's list by name in
// engine-bounds-mirror.test.ts — the union through a Record whose keys must be
// exactly the union's, so a member gained or lost on either side is a
// compile-time error and a red test rather than a silent widening.
export type CappingBand = Exclude<CanvasProjection['bands'][number]['name'], 'content'>
export const CAPPING_BANDS = BANDS_CAPPING_VERTICALLY as ReadonlyArray<CappingBand>

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
// STORY 12.3 — the table's own header and row properties, beside its columns.
//
// TWO MEMBERS PER HEADER-STYLE FIELD, and the pair is the whole point. The bare
// name (`headerAlign`) is what the DOCUMENT DECLARES — '' or 0 when the key is
// absent — so a control can tell set from unset and offer to clear back to
// absent. The `…Resolved` twin is what the document WILL USE, which for an
// absent field is the table's own `style.<field>` and then that field's
// documented default.
//
// THE RESOLVED HALF IS THE ENGINE'S ANSWER AND IS NEVER RECOMPUTED HERE. Go's
// resolveHeaderStyle is the one cascade in this program; it runs at the
// projection's construction site and its result travels on the wire. Handing
// the browser the committed field plus the table's own style and letting it
// choose IS implementing the cascade in the browser — forbidden by AD-15 and
// AD-17 and by this story's own AC2 and AC3.
//
// `headerHeight` AND `altRowBackground` CARRY ONE MEMBER EACH. Neither has a
// cascade to resolve through — the first is required so it is never absent, the
// second is a flat override with no fallback level of its own — so committed IS
// resolved and a second member would be ceremony: a duplicate of the committed
// value, carried on every projection, that a later reader has to keep agreeing
// with itself.
export type TableHeaderStyle = Readonly<{
  headerFontFamily: string; headerFontFamilyResolved: string
  headerFontSize: number; headerFontSizeResolved: number
  headerLineSpacing: number; headerLineSpacingResolved: number
  headerBackground: string; headerBackgroundResolved: string
  headerColor: string; headerColorResolved: string
  headerValign: string; headerValignResolved: string
  headerAlign: string; headerAlignResolved: string
}>
export type TableColumns = Readonly<{ revision: number; table: Readonly<{ tableId: string; collection: string; alias: string; headerHeight: number; altRowBackground: string; columns: ReadonlyArray<TableColumn> }> & TableHeaderStyle }>

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
	// The DOCUMENT's two declared formatting authorities (Story 12.2). `locale`
	// is one of AD-12's four tags and decides how every formatDate and
	// formatNumber in the document renders — including the Buddhist-era year
	// under `th`; `utcOffset` is the ±HH:MM string those dates are resolved in.
	// Both come from Go and neither is defaulted, derived or validated here: the
	// panel shows what the engine holds, proposes what the author typed, and
	// lets the engine refuse it in the engine's own sentence.
	locale: LocaleTag; utcOffset: string
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
	//
	// defaultLineSpacing (Story 17.3) is the same promise for LEADING: the
	// ratio the producer measures an element with when its style declares
	// none, in THOUSANDTHS, so 1000 is a ratio of 1.0. It is here because the
	// inspector used to spell that `1` itself, and a designer-side copy of an
	// engine-owned default is a second authority that can drift silently.
	fontFamilies: ReadonlyArray<string>; defaultFontSize: number; defaultLineSpacing: number
	// fontChains is the SAME set of chains, with the ordered ENTRIES behind each
	// name: fontChains.map(c => c.name) is fontFamilies, entry for entry, and
	// the validator asserts it rather than trusting it. Entry order is the
	// document's own authored order and is never re-sorted here.
	//
	// An entry is a discriminated OBJECT since Story 8.3: `face` names a face the
	// renderer is given, `assetKey` names one the document carries, exactly one of
	// them is non-empty, and `family`/`style` are what the panel DISPLAYS for an
	// embedded entry — read by Go from the asset's own `font` record, never
	// derived in the browser.
	fontChains: ReadonlyArray<Readonly<{ name: string; entries: ReadonlyArray<Readonly<{ face: string; assetKey: string; family: string; style: string }>> }>>
	bands: ReadonlyArray<Readonly<{ name: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number }>>
	components: ReadonlyArray<Readonly<{ id: string; type: 'text' | 'image' | 'table' | 'line' | 'rect'; band: 'pageHeader' | 'content' | 'pageFooter'; x: number; y: number; width: number; height: number; resizable: boolean; value?: string; binding?: string; visibleIf?: string; fontFamily?: string; fontSize?: number; lineSpacing?: number; bold?: boolean; italic?: boolean; align?: 'left' | 'center' | 'right' | 'justify'; valign?: 'top' | 'middle' | 'bottom'; color?: string; background?: string; borderWidth?: number; borderColor?: string; borderEdges?: ReadonlyArray<'top' | 'right' | 'bottom' | 'left'>; paddingTop?: number; paddingRight?: number; paddingBottom?: number; paddingLeft?: number; tableBind?: string; textPaint?: Readonly<{ overflow: boolean; truncated: boolean; lines: ReadonlyArray<Readonly<{ top: number; baseline: number; advance: number; width: number; fragments: ReadonlyArray<Readonly<{ text: string; x: number; face?: string; assetKey?: string }>> }>> }>; image?: Readonly<{ mediaType: string; assetKey: string; width: number; height: number; drawX: number; drawY: number; drawWidth: number; drawHeight: number }>; imageUnavailable?: 'missing' | 'undecodable' }>>
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

// isFontChainEntry is the projected chain entry's own guard (Story 8.3).
// Split out of the isCanvas one-liner rather than inlined there because it
// carries a RULE — the discriminated shape — and a rule buried inside a chain
// of `&&` is a rule nobody edits deliberately.
//
// THE DISCRIMINANT IS PROJECTED, NEVER DERIVED HERE. Exactly one of `face` and
// `assetKey` is non-empty, and this ASSERTS that rather than guessing from a
// value's shape: a 64-character face name is a legal face name, so "looks like
// a digest" was never available as a test, and FontChainEditor is forbidden a
// rule of its own.
//
// An embedded entry always carries a non-empty `family` — Go decides what the
// panel shows and falls back to the asset key — and a named face carries no
// family and no style at all, because its name IS its identity.
const isFontChainEntry = (value: unknown): boolean => {
  if (!isRecord(value) || !hasExactKeys(value, ['face', 'assetKey', 'family', 'style'])) return false
  const { face, assetKey, family, style } = value
  if (typeof face !== 'string' || typeof assetKey !== 'string' || typeof family !== 'string' || typeof style !== 'string') return false
  if ([face, assetKey, family, style].some((text) => text.length > MAX_CANVAS_PROPERTY_STRING)) return false
  if ((face.length > 0) === (assetKey.length > 0)) return false
  if (assetKey.length > 0) return family.length > 0
  return family.length === 0 && style.length === 0
}

const isRecord = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null && Object.getPrototypeOf(value) === Object.prototype
const isArrayBuffer = (value: unknown): value is ArrayBuffer => Object.prototype.toString.call(value) === '[object ArrayBuffer]'
const isRenderPayload = (value: unknown): value is RenderPayload => isRecord(value) && hasExactKeys(value, ['template', 'data', 'params']) && ['template', 'data', 'params'].every((key) => isArrayBuffer(value[key]) && value[key].byteLength > 0 && value[key].byteLength <= MAX_ENGINE_PAYLOAD_BYTES)
const isIdentityPayload = (value: unknown): value is IdentityPayload => isRecord(value) && hasExactKeys(value, ['data', 'params']) && ['data', 'params'].every((key) => isArrayBuffer(value[key]) && value[key].byteLength > 0 && value[key].byteLength <= MAX_ENGINE_PAYLOAD_BYTES)
// DW-70. Go sorts the projected chain names with slices.Sorted over Go
// strings, which compares them BY BYTE — and those keys are the canonical
// `.folio`'s own `fonts` key order under AD-9, so Go's order IS the document's
// order and is NORMATIVE. JavaScript's `<` compares UTF-16 CODE UNITS, and the
// two disagree wherever a name mixes the astral planes with U+E000-U+FFFF: a
// surrogate pair (0xD800-) sorts BELOW U+E000 in UTF-16 and ABOVE it in UTF-8.
// The measured pair that motivated this is `'\uE000'` before `'\u{1F600}'` —
// Go's order, and the order this guard used to REJECT, taking the whole
// snapshot with it (isCanvas false -> parseInbound undefined -> PROTOCOL_INVALID
// -> engine-client terminates the worker and the canvas is permanently blank).
// Comparing by CODE POINT is the same sequence as comparing UTF-8 bytes, so the
// browser adopts Go's order. Never the reverse: changing Go's comparator would
// move golden bytes for any document whose chain names cross the boundary.
const compareCodePoints = (left: string, right: string): number => {
  const a = Array.from(left, (unit) => unit.codePointAt(0) as number)
  const b = Array.from(right, (unit) => unit.codePointAt(0) as number)
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return (a[index] as number) < (b[index] as number) ? -1 : 1
  return a.length === b.length ? 0 : a.length < b.length ? -1 : 1
}
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
  if (!isRecord(value) || !hasExactKeys(value, ['revision', 'table']) || typeof value.revision !== 'number' || !Number.isSafeInteger(value.revision) || value.revision < 0 || !isRecord(value.table) || !hasExactKeys(value.table, ['tableId', 'collection', 'alias', 'headerHeight', 'altRowBackground', 'headerFontFamily', 'headerFontFamilyResolved', 'headerFontSize', 'headerFontSizeResolved', 'headerLineSpacing', 'headerLineSpacingResolved', 'headerBackground', 'headerBackgroundResolved', 'headerColor', 'headerColorResolved', 'headerValign', 'headerValignResolved', 'headerAlign', 'headerAlignResolved', 'columns'])) return false
  const table = value.table
  // THE TYPED CLAUSES FOR STORY 12.3's SIXTEEN MEMBERS. Every one is REQUIRED
  // and never optional: hasExactKeys above already refuses a response that
  // omits one, and the Go side has no `omitempty` for exactly that reason
  // (canvas_projection_wire_test.go's fifth record pins both directions).
  //
  // ADMITTING A VALUE IS NOT ADJUDICATING IT. These bounds exist so a malformed
  // response cannot reach React, not so the browser can second-guess the
  // engine: a value Go committed is a value Go already ruled on. '' and 0 are
  // admitted throughout because that is how this projection spells ABSENT.
  const headerString = (key: keyof TableHeaderStyle | 'altRowBackground') => typeof table[key] === 'string' && (table[key] as string).length <= MAX_CANVAS_PROPERTY_STRING
  // THE GUARD MUST ADMIT EXACTLY WHAT THE FILE DOOR ADMITS, and for these two
  // lengths the file door admits a NEGATIVE one. `internal/template/decimal.go`
  // negates on `sign < 0` and neither `parse_bands.go` (`t.HeaderHeight = hh`)
  // nor the style decoder (`st.FontSize = present(v)`) bounds the result, so a
  // hand-authored `"headerHeight": -5` loads and renders today. A guard
  // requiring `>= 0` therefore refused a document the engine had already
  // accepted — and the symptom was not a warning: parseInbound returns
  // undefined, engine-client terminates the worker with no re-spawn, and on a
  // FIRST table-editor open nothing is shown at all, because the panel that
  // would render the error never mounts.
  //
  // The bound is relaxed HERE rather than added at the loader on purpose:
  // bounding the loader is a format narrowing, and this story's Never list
  // forbids one. DW-26 already records those lengths as unbounded at load.
  const headerLength = (key: keyof TableHeaderStyle | 'headerHeight') => typeof table[key] === 'number' && Number.isSafeInteger(table[key])
  // The line-spacing pair keeps `>= 0`, because for IT the file door really
  // does bound: DecodeLineSpacingRaw refuses anything outside
  // [MinLineSpacingThousandths, MaxLineSpacingThousandths] = [1, 1000000], and
  // 0 is this projection's spelling of absent.
  const headerRatio = (key: keyof TableHeaderStyle) => headerLength(key) && (table[key] as number) >= 0
  if (!(['altRowBackground', 'headerFontFamily', 'headerFontFamilyResolved', 'headerBackground', 'headerBackgroundResolved', 'headerColor', 'headerColorResolved'] as const).every(headerString)) return false
  if (!(['headerHeight', 'headerFontSize', 'headerFontSizeResolved'] as const).every(headerLength)) return false
  if (!(['headerLineSpacing', 'headerLineSpacingResolved'] as const).every(headerRatio)) return false
  // A COMMITTED alignment may be '' — that is what absent looks like — while a
  // RESOLVED one never is: resolveHeaderStyle seeds `left` and `top` before it
  // cascades anything, so an empty resolved value would mean the engine skipped
  // its own default.
  if (!['', 'left', 'center', 'right'].includes(table.headerAlign as string) || !['left', 'center', 'right'].includes(table.headerAlignResolved as string)) return false
  if (!['', 'top', 'middle', 'bottom'].includes(table.headerValign as string) || !['top', 'middle', 'bottom'].includes(table.headerValignResolved as string)) return false
  return typeof table.tableId === 'string' && table.tableId.length > 0 && table.tableId.length <= MAX_ENGINE_ELEMENT_ID_LENGTH && typeof table.collection === 'string' && table.collection.length > 0 && table.collection.length <= MAX_ENGINE_BINDING_LENGTH && typeof table.alias === 'string' && table.alias.length > 0 && table.alias.length <= 64 && Array.isArray(table.columns) && table.columns.length <= 128 && table.columns.every((column) => isRecord(column) && hasExactKeys(column, ['id', 'header', 'width', 'align', 'binding', 'rowField', 'rowFieldEditable', 'footer', 'footerOf', 'footerFormat']) && typeof column.id === 'string' && column.id.length > 0 && column.id.length <= MAX_ENGINE_ELEMENT_ID_LENGTH && typeof column.header === 'string' && column.header.length <= 256 && typeof column.width === 'number' && Number.isSafeInteger(column.width) && column.width > 0 && ['left', 'center', 'right'].includes(column.align as string) && typeof column.binding === 'string' && column.binding.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.rowField === 'string' && column.rowField.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.rowFieldEditable === 'boolean' && ['','sum','avg','count'].includes(column.footer as string) && typeof column.footerOf === 'string' && column.footerOf.length <= MAX_ENGINE_BINDING_LENGTH && typeof column.footerFormat === 'string' && column.footerFormat.length <= 256) && new Set(table.columns.map((item) => (item as Record<string, unknown>).id)).size === table.columns.length
}
const isCanvas = (value: unknown): value is CanvasProjection => {
  if (!isRecord(value) || !hasOnly(value, ['width', 'height', 'orientation', 'preset', 'locale', 'utcOffset', 'marginTop', 'marginRight', 'marginBottom', 'marginLeft', 'gridIncrement', 'commandWidth', 'commandHeight', 'fontFamilies', 'fontChains', 'defaultFontSize', 'defaultLineSpacing', 'contentWindowHeight', 'contentWindowCount', 'contentWindowOrigins', 'contentWindowCountIsExact', 'bands', 'components']) || !['A4', 'Letter', 'custom'].includes(value.preset as string) || (value.orientation !== 'portrait' && value.orientation !== 'landscape')) return false
  // THE TWO DOCUMENT-SETTINGS CLAUSES ARE LOAD-BEARING, and `hasOnly` above
  // cannot stand in for them: it is a SUBSET check, so a key Go simply failed
  // to send passes it and reaches the panel as `undefined` — a locale row with
  // no value and an offset row that would send the string "undefined" back.
  // Only a typed clause catches an ABSENT key, which is the failure this story
  // could otherwise ship in silence.
  //
  // `locale` is checked against LOCALE_TAGS, the browser's one spelling of
  // AD-12's closed set, exactly as `preset` and `orientation` are checked
  // against theirs on the line above. `utcOffset` is checked only for shape —
  // a non-empty string within the projection's ordinary string bound. ITS
  // GRAMMAR IS NOT RESTATED HERE: ±HH:MM is the engine's rule
  // (template.IsUTCOffset, the one predicate the loader and the command door
  // both ask), and a browser-side copy of it would be a second authority that
  // could refuse a document Go admits.
  if (!LOCALE_TAGS.includes(value.locale as LocaleTag)) return false
  if (typeof value.utcOffset !== 'string' || value.utcOffset.length === 0 || value.utcOffset.length > MAX_CANVAS_PROPERTY_STRING) return false
  const integer = (key: string, positive = false) => typeof value[key] === 'number' && Number.isSafeInteger(value[key]) && (positive ? value[key] > 0 : value[key] >= 0)
  if (!['width', 'height', 'gridIncrement', 'commandWidth', 'commandHeight', 'defaultFontSize', 'defaultLineSpacing', 'contentWindowHeight', 'contentWindowCount'].every((key) => integer(key, true)) || !['marginTop', 'marginRight', 'marginBottom', 'marginLeft'].every((key) => integer(key))) return false
  // The declared font chain names, as Go sorted them: bounded in count and
  // length like every other list on this projection, unique, and in the order
  // Go sent so the browser never re-sorts an engine-owned set. The ordering
  // check is compareCodePoints, which is Go's byte order — see DW-70 above;
  // `>=` on JavaScript strings is NOT, and dropped a legitimate snapshot.
  if (!Array.isArray(value.fontFamilies) || value.fontFamilies.length > MAX_ENGINE_FONT_FAMILIES || !value.fontFamilies.every((name) => typeof name === 'string' && name.length > 0 && name.length <= MAX_CANVAS_PROPERTY_STRING) || value.fontFamilies.some((name, index, names) => index > 0 && compareCodePoints(names[index - 1] as string, name as string) >= 0)) return false
  // The chains those names stand for. Bounded in count and in per-chain entry
  // count, no chain empty — an empty chain is not one Go projects, because it
  // is not one style.fontFamily may name. The `chain.name === fontFamilies[i]`
  // clause is the cross-check the two lists exist to give each other: Go builds
  // fontFamilies FROM fontChains, so any disagreement here is a channel fault
  // and the snapshot is not trusted.
  //
  // AN ENTRY IS AN OBJECT, NOT A STRING (Story 8.3). It used to be
  // `typeof face === 'string'`, and that clause rejected an object entry
  // outright — isCanvas false, parseInbound undefined, the worker terminated
  // and the canvas permanently blank with no element id and nothing to
  // attribute it to. That is why the Go projection and this guard change in
  // ONE commit; canvas_projection_wire_test.go reddens if only one of them
  // moves, at the entry level as well as at the chain level.
  const chains = value.fontChains
  if (!Array.isArray(chains) || chains.length !== value.fontFamilies.length) return false
  if (!chains.every((chain, index) => isRecord(chain) && hasExactKeys(chain, ['name', 'entries']) && chain.name === (value.fontFamilies as ReadonlyArray<unknown>)[index] && Array.isArray(chain.entries) && chain.entries.length > 0 && chain.entries.length <= MAX_ENGINE_FONT_CHAIN_ENTRIES && chain.entries.every((entry) => isFontChainEntry(entry)))) return false
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
    // A FRAGMENT'S TWO ATTRIBUTION KEYS ARE A DISCRIMINATED PAIR, on the model
    // isFontChainEntry already applies one level up: `assetKey` names a face
    // the DOCUMENT carries, `face` (Story 8.4e) names one the ENGINE ships,
    // and no fragment may carry both. NEITHER is legal and deliberately so —
    // that is the wire's own statement of "unattributed", and such a fragment
    // paints on the stylesheet's declared stack rather than terminating the
    // worker. `face` is bounded by MAX_CANVAS_PROPERTY_STRING, the same bound
    // a chain entry's `face` already uses, and is checked HERE rather than
    // trusted: Go can only put a FontSet key on this field today, but a
    // guard's job is to hold when the other side is wrong.
    return line.fragments.every((fragment) => {
      fragments++
      return fragments <= MAX_CANVAS_BODY_TEXT_FRAGMENTS && isRecord(fragment) && hasOnly(fragment, ['text', 'x', 'face', 'assetKey']) && typeof fragment.text === 'string' && fragment.text.length > 0 && fragment.text.length <= MAX_CANVAS_BODY_TEXT && typeof fragment.x === 'number' && Number.isSafeInteger(fragment.x) && fragment.x >= component.x && fragment.x <= component.x + Math.max(paint.width, component.width) && (fragment.assetKey === undefined || (typeof fragment.assetKey === 'string' && /^[a-f0-9]{64}$/.test(fragment.assetKey))) && (fragment.face === undefined || (typeof fragment.face === 'string' && fragment.face.length > 0 && fragment.face.length <= MAX_CANVAS_PROPERTY_STRING)) && !(fragment.face !== undefined && fragment.assetKey !== undefined)
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
