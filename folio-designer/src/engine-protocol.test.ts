import { describe, expect, it } from 'vitest'
import { ENGINE_PROTOCOL_VERSION, MAX_ENGINE_BINDING_LENGTH, MAX_ENGINE_DATA_PATH_LENGTH, MAX_ENGINE_ELEMENT_ID_LENGTH, MAX_ENGINE_PAYLOAD_BYTES, MAX_ENGINE_RENDER_PDF_BYTES, deepFreeze, parseInbound, parseRequest } from './engine-protocol'

const canvas = { width: 1000, height: 2000, orientation: 'portrait', preset: 'custom', marginTop: 0, marginRight: 0, marginBottom: 0, marginLeft: 0, gridIncrement: 100, commandWidth: 1000, commandHeight: 2000, fontFamilies: ['body'], defaultFontSize: 12000, bands: [{ name: 'pageHeader', x: 0, y: 0, width: 1000, height: 100 }, { name: 'content', x: 0, y: 100, width: 1000, height: 1800 }, { name: 'pageFooter', x: 0, y: 1900, width: 1000, height: 100 }], components: [] }

describe('canvas projection protocol guard', () => {
  it('accepts and deeply freezes the exact three bounded bands', () => {
    const inbound = parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas } })
    expect(inbound).toBeDefined()
    const frozen = deepFreeze(canvas)
    expect(Object.isFrozen(frozen.bands)).toBe(true)
    expect(Object.isFrozen(frozen.bands[0])).toBe(true)
  })

  it.each([
    [{ ...canvas, bands: [canvas.bands[1], canvas.bands[0], canvas.bands[2]] }],
    [{ ...canvas, bands: [{ ...canvas.bands[0], name: 'content' }, ...canvas.bands.slice(1)] }],
    [{ ...canvas, bands: [{ ...canvas.bands[0], x: 1_000 }, ...canvas.bands.slice(1)] }],
    [{ ...canvas, width: Number.MAX_SAFE_INTEGER + 1 }],
  ])('rejects structurally false paint geometry', (bad) => {
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })).toBeUndefined()
  })

  it.each([
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true }, { id: 'e1', type: 'rect', band: 'content', x: 20, y: 0, width: 10, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 991, y: 0, width: 10, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'table', band: 'content', x: 0, y: 0, width: 0, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: false }] },
  ])('rejects ambiguous, out-of-band, or incoherent component paint geometry', (bad) => {
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })).toBeUndefined()
  })

  it('bounds opaque producer failure provenance at the main-thread boundary', () => {
    const response = (error: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: false, error })
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH), dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH) })).toBeDefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH + 1) })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH + 1) })).toBeUndefined()
    expect(response({ code: 'C'.repeat(97), message: 'invalid' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'm'.repeat(513) })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 7 })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: [] })).toBeUndefined()
  })

  it('rejects surplus authority-bearing fields at every projection level', () => {
    const response = (bad: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })
    expect(response({ ...canvas, style: {} })).toBeUndefined()
    expect(response({ ...canvas, bands: [{ ...canvas.bands[0], extra: true }, ...canvas.bands.slice(1)] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, Extra: {} }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, fontFamily: 'body' }] })).toBeUndefined()
  })

  it('accepts only bounded, ordered engine text paint and rejects browser-shaped substitutes', () => {
    const textPaint = { overflow: false, lines: [{ top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'engine line', x: 0 }] }] }
    const response = (projection: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: projection } })
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint }] })).toBeDefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, viewportWidth: 100 } }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, lines: [{ ...textPaint.lines[0], width: 11 }] } }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, lines: [{ ...textPaint.lines[0], fragments: [{ text: 'engine line', x: 0, fontMetrics: 1 }] }] } }] })).toBeUndefined()
  })

  it('admits one bounded text-binding paint label but rejects an editable projection', () => {
    const response = (component: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [component] } } })
    const text = { id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, binding: 'customer.name', textPaint: { overflow: false, lines: [] } }
    expect(response(text)).toBeDefined()
    expect(response({ ...text, binding: 'a'.repeat(MAX_ENGINE_BINDING_LENGTH + 1) })).toBeUndefined()
    expect(response({ ...text, binding: '' })).toBeUndefined()
    expect(response({ ...text, binding: { path: 'customer.name' } })).toBeUndefined()
    expect(response({ ...text, type: 'image' })).toBeUndefined()
  })

  it('rejects non-advancing or out-of-box text paint geometry', () => {
    const response = (textPaint: object) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint }] } },
    })
    const line = { top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'engine', x: 0 }] }
    expect(response({ overflow: false, lines: [line, { ...line, top: 11, baseline: 19 }] })).toBeUndefined()
    expect(response({ overflow: false, lines: [{ ...line, baseline: 13 }] })).toBeUndefined()
    expect(response({ overflow: false, lines: [{ ...line, fragments: [{ text: 'engine', x: 11 }] }] })).toBeUndefined()
    expect(response({ overflow: false, lines: [{ ...line, top: -1, baseline: 8 }] })).toBeUndefined()
  })

  it('accepts a well-formed image paint inside its own box and rejects malformed or out-of-box substitutes', () => {
    const box = { x: 0, y: 0, width: 100, height: 50 }
    // Finding 12 (review of 2026-08-29): the wire key is the FULL 64-hex
    // digest (D-5.13.2 amendment) — a truncated key passed admission here
    // before the fix, then could never resolve through the per-key 'asset'
    // fetch. This fixture's OWN well-formedness now pins that width.
    const image = { mediaType: 'image/png', assetKey: 'ab'.repeat(32), width: 300, height: 150, drawX: 0, drawY: 0, drawWidth: 100, drawHeight: 50 }
    const response = (img: object | undefined) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', resizable: true, ...box, ...(img === undefined ? {} : { image: img }) }] } },
    })
    expect(response(image)).toBeDefined()
    expect(response(undefined)).toBeDefined()
    expect(response({ ...image, mediaType: '' })).toBeUndefined()
    expect(response({ ...image, assetKey: '' })).toBeUndefined()
    expect(response({ ...image, assetKey: 'z'.repeat(64) })).toBeUndefined() // 64 chars, not hex
    // A well-formed but TRUNCATED key (12 hex characters, the inspector's
    // own display-abbreviation width) must be rejected, not merely a
    // wrong-length string of the wrong shape.
    expect(response({ ...image, assetKey: 'abcdef012345' })).toBeUndefined()
    expect(response({ ...image, width: 0 })).toBeUndefined()
    expect(response({ ...image, height: -1 })).toBeUndefined()
    expect(response({ ...image, drawWidth: 0 })).toBeUndefined()
    expect(response({ ...image, drawX: -1 })).toBeUndefined()
    expect(response({ ...image, drawX: 1, drawWidth: 100 })).toBeUndefined() // spills past box.x+box.width
    expect(response({ ...image, drawY: 1, drawHeight: 50 })).toBeUndefined() // spills past box.y+box.height
    expect(response({ ...image, extra: true })).toBeUndefined()
    // A non-image component must never carry an image paint.
    expect(parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, lines: [] }, image }] } },
    })).toBeUndefined()
  })

  it("admits Finding 9's imageUnavailable discriminant only for an image component with no image paint, and only its two named values", () => {
    const box = { x: 0, y: 0, width: 100, height: 50 }
    const response = (component: object) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', resizable: true, ...box, ...component }] } },
    })
    expect(response({ imageUnavailable: 'missing' })).toBeDefined()
    expect(response({ imageUnavailable: 'undecodable' })).toBeDefined()
    expect(response({})).toBeDefined() // absent is legal (an image paint present, or nothing yet)
    expect(response({ imageUnavailable: 'something-else' })).toBeUndefined()
    // Not a text component.
    expect(parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, lines: [] }, imageUnavailable: 'missing' }] } },
    })).toBeUndefined()
    // Not alongside a PRESENT image paint — the two are one signal.
    const image = { mediaType: 'image/png', assetKey: 'ab'.repeat(32), width: 300, height: 150, drawX: 0, drawY: 0, drawWidth: 100, drawHeight: 50 }
    expect(response({ image, imageUnavailable: 'missing' })).toBeUndefined()
  })

  it('admits only operation-coherent closed worker requests and responses', () => {
    const load = new Uint8Array([1]).buffer
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'load-1', operation: 'load', payload: load })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'load-2', operation: 'load' })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'snapshot-1', operation: 'snapshot', payload: load })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'snapshot-2', operation: 'snapshot', viewport: 900 })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'undo-1', operation: 'undo' })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'redo-1', operation: 'redo' })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'undo-2', operation: 'undo', payload: load })).toBeUndefined()
    const dpr = ['device', 'PixelRatio'].join('')
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'command-1', operation: 'command', payload: load, [dpr]: 2 })).toBeUndefined()
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'response-1', ok: false, error: { code: 'NO', message: 'no' }, font: 'browser' })).toBeUndefined()
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'lifecycle', state: 'ready', snapshot: {} })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'asset-1', operation: 'asset', payload: load })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'asset-2', operation: 'asset' })).toBeUndefined()
    const assetResponse = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response' as const, requestId: 'asset-1', ok: true as const, snapshot: { documentState: 'loaded' as const, revision: 3, byteLength: 1 }, bytes: load }
    expect(parseInbound(assetResponse)).toBeDefined()
  })

  it('accepts only a correlated, bounded three-byte render envelope and producer digest', () => {
    const part = new Uint8Array([1]).buffer
    const render = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'render-1', operation: 'render', payload: { template: part, data: part.slice(0), params: part.slice(0) } }
    expect(parseRequest(render)).toBeDefined()
    expect(parseRequest({ ...render, payload: { template: part, data: part.slice(0) } })).toBeUndefined()
    expect(parseRequest({ ...render, viewport: 900 })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'render-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, bytes: part, preview: { revision: 7, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, pdfSha256: 'not-a-digest' } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, identity: 'not-an-identity' } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, revision: -1 } })).toBeUndefined()
    expect(parseInbound({ ...response, bytes: undefined })).toBeUndefined()
    expect(parseRequest({ ...render, payload: { template: new ArrayBuffer(MAX_ENGINE_PAYLOAD_BYTES + 1), data: part.slice(0), params: part.slice(0) } })).toBeUndefined()
    expect(parseInbound({ ...response, bytes: new ArrayBuffer(MAX_ENGINE_RENDER_PDF_BYTES + 1) })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, diagnostics: [{}] } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, diagnostics: [], snapshot: {} } })).toBeUndefined()
  })

  it('accepts only an identity-only engine response with revision-bound opaque evidence', () => {
    const part = new Uint8Array([1]).buffer
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'identity-1', operation: 'identity', payload: { data: part, params: part.slice(0) } }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: { data: part } })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'identity-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, preview: { revision: 7, identity: 'a'.repeat(64) } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, pdfSha256: 'b'.repeat(64) } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, revision: 6 } })).toBeUndefined()
  })

  it('admits only a bounded, revision-correlated engine parameter-reference projection', () => {
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'params-1', operation: 'parameter-references' }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: new Uint8Array([1]).buffer })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'params-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, parameterReferences: { revision: 7, names: ['branch', 'reportDate'] } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 6, names: ['branch'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate', 'branch'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate', 'reportDate'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate'], expression: 'params.reportDate' } })).toBeUndefined()
  })

  it('admits only a selected, revision-correlated table-column paint projection', () => {
    const payload = new TextEncoder().encode('{"id":"e7"}').buffer
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'table-1', operation: 'table-columns', payload }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: undefined })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'table-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, tableColumns: { revision: 7, table: { tableId: 'e7', collection: 'transactions[]', alias: 'transaction', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'right', binding: '{{transaction.amount}}', rowField: 'amount', rowFieldEditable: true, footer: 'sum', footerOf: 'transactions.amount', footerFormat: '#,##0.00' }] } } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, revision: 6 } })).toBeUndefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, table: { ...response.tableColumns.table, columns: [{ ...response.tableColumns.table.columns[0], bind: 'row.amount' }] } } })).toBeUndefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, table: { ...response.tableColumns.table, columns: [{ ...response.tableColumns.table.columns[0], width: 0 }] } } })).toBeUndefined()
  })
})
