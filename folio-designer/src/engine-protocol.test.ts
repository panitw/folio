import { describe, expect, it } from 'vitest'
import { ENGINE_PROTOCOL_VERSION, MAX_ENGINE_DATA_PATH_LENGTH, MAX_ENGINE_ELEMENT_ID_LENGTH, deepFreeze, parseInbound } from './engine-protocol'

const canvas = { width: 1000, height: 2000, orientation: 'portrait', preset: 'custom', marginTop: 0, marginRight: 0, marginBottom: 0, marginLeft: 0, gridIncrement: 100, commandWidth: 1000, commandHeight: 2000, bands: [{ name: 'pageHeader', x: 0, y: 0, width: 1000, height: 100 }, { name: 'content', x: 0, y: 100, width: 1000, height: 1800 }, { name: 'pageFooter', x: 0, y: 1900, width: 1000, height: 100 }], components: [] }

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

  it('bounds optional component diagnostic fields at the main-thread boundary', () => {
    const response = (error: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: false, error })
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH), dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH) })).toBeDefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH + 1) })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH + 1) })).toBeUndefined()
  })

  it('rejects surplus authority-bearing fields at every projection level', () => {
    const response = (bad: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })
    expect(response({ ...canvas, style: {} })).toBeUndefined()
    expect(response({ ...canvas, bands: [{ ...canvas.bands[0], extra: true }, ...canvas.bands.slice(1)] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, Extra: {} }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, fontFamily: 'body' }] })).toBeUndefined()
  })
})
