import { createEvent, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App, { placementPoint } from './App'
import { shortcutHintsFor } from './shortcuts'
import { embeddedFaceFamily } from './embedded-face-family'
import { FileAccessCancelled, type FileAccess } from './file/file-access'
import type { EngineClient } from './engine-client'
import type { CanvasProjection } from './engine-protocol'
import { acceptSampleData } from './sample-data'
import { MAX_CANVAS_SHEETS } from './sheet-stack'

// face() builds the PROJECTED shape of a named-face chain entry (Story 8.3:
// an entry is a discriminated object, not a string). A named face carries no
// family and no style — its name is its identity.
const face = (name: string) => ({ face: name, assetKey: '', family: '', style: '' })

// carried() is the projected shape of an EMBEDDED chain entry: no face name,
// an asset key, and the family/style Go read out of the asset's own `font`
// record for the panel to display. The key never becomes a family here — that
// derivation is embedded-face-family.ts's alone (D-8.4.1).
const carried = (assetKey: string) => ({ face: '', assetKey, family: 'Noto Sans Thai', style: 'Regular' })

// installStubFontSet installs the page font set jsdom does not implement and
// returns its own removal. `Object.defineProperty` because neither the face
// constructor nor the set exists to be assigned over.
//
// IT RECORDS WHAT THE SEAM DID TO IT, in order: the families added and the
// families removed. Registration and release are otherwise invisible — nothing
// in the DOM says a face was released — so a claim about the seam's LIFETIME
// can only be made against this record.
function installStubFontSet(): Readonly<{ restore: () => void; added: string[]; removed: string[] }> {
  class StubFace {
    readonly family: string
    constructor(family: string) { this.family = family }
    load(): Promise<StubFace> { return Promise.resolve(this) }
  }
  const added: string[] = []
  const removed: string[] = []
  const set = { add: (face: StubFace) => { added.push(face.family); return undefined }, delete: (face: StubFace) => { removed.push(face.family); return undefined } }
  Object.defineProperty(globalThis, 'FontFace', { value: StubFace, configurable: true, writable: true })
  Object.defineProperty(document, 'fonts', { value: set, configurable: true, writable: true })
  return { restore: () => { Reflect.deleteProperty(globalThis, 'FontFace'); Reflect.deleteProperty(document, 'fonts') }, added, removed }
}

// TWO text components drawing through ONE carried entry, so a per-component
// registration lifetime shows up as two asset requests instead of one.
const carriedFaceCanvas = (key: string) => {
  const paint = { overflow: false, truncated: false, lines: [{ top: 0, baseline: 12_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] }
  const component = (id: string, y: number) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height: 24_000, resizable: true, value: 'ignored', textPaint: paint })
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(key)] }], components: [component('e1', 0), component('e2', 30_000)] }
}

// TWO carried entries, ONE PER COMPONENT, so a claim about one key can be read
// off the DOM against a SETTLED outcome for the other. The fragment order in
// the container is the component order: `first` then `second`.
const twoCarriedFacesCanvas = (first: string, second: string) => {
  const paint = (key: string) => ({ overflow: false, truncated: false, lines: [{ top: 0, baseline: 12_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] })
  const component = (id: string, y: number, key: string) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height: 24_000, resizable: true, value: 'ignored', textPaint: paint(key) })
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(first), carried(second)] }], components: [component('e1', 0, first), component('e2', 30_000, second)] }
}

// ONE carried component that CROSSES A WINDOW SEAM, so the projection produces
// a home occurrence AND an echo of it on the next sheet. Every content window
// after the first is drawn by echoes, so a multi-sheet document is the ordinary
// case rather than an exotic one.
const carriedFaceEchoCanvas = (key: string) => {
  const paint = { overflow: false, truncated: false, lines: [{ top: 650_000, baseline: 662_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] }
  const spanning = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 650_000, width: 72_000, height: 100_000, resizable: true, value: 'ignored', textPaint: paint }
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(key)] }], contentWindowCount: 3, contentWindowOrigins: [0, 700_000, 1_400_000], components: [spanning] }
}

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => false,
  PDFPreviewViewer: ({ label, describedBy, onPageCount, onError }: { label: string; describedBy: string; onPageCount: (pages: number) => void; onError: (error: Error) => void }) => <><button type="button" aria-label={label} aria-describedby={describedBy} onClick={() => onPageCount(1)}>Admit local PDF</button><button type="button" aria-label="Fail local PDF viewer" onClick={() => onError(new Error('viewer rejected bytes'))}>Fail local PDF viewer</button></>,
}))

const bytes = new Uint8Array([1, 2, 3]).buffer
const sample = acceptSampleData('sample.json', new TextEncoder().encode('{"customer":{"name":"Preview customer"},"transactions":[]}').buffer)
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body', 'heading'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans Thai')] }], defaultFontSize: 12000, contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = (revision: number) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas })
const engine = (request = vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3 }, ...(operation === 'serialize' ? { bytes } : {}) }))) => ({ request }) as unknown as EngineClient

describe('application shell', () => {
  it('hydrates engine-owned state when asynchronous startup replaces the loading shell', () => {
    const lifecycle = { state: 'ready' as const, cacheReady: true, verifiedAssetUrls: [] }
    const view = render(<App key="engine-loading" loadState={lifecycle} engineState="starting" />)
    expect(screen.getByRole('status', { name: 'Engine preparation status' })).toHaveTextContent('Starting local engine')
    view.rerender(<App key="engine-ready" engine={engine()} initialSnapshot={snapshot(1)} loadState={lifecycle} engineState="starting" />)
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 1')
    expect(screen.getByLabelText('Report page with Page Header, Content, and Page Footer')).toBeInTheDocument()
  })

  it('renders every persistent desktop landmark and honest later regions', () => {
    render(<App initialSnapshot={snapshot(1)} />)
    expect(screen.getByLabelText('Document bar')).toBeInTheDocument()
    expect(screen.getByLabelText('Component palette')).toBeInTheDocument()
    expect(screen.getByLabelText('Canvas region')).toBeInTheDocument()
    expect(screen.getByLabelText('Report page with Page Header, Content, and Page Footer')).toBeInTheDocument()
    expect(screen.getByLabelText('Properties panel')).toBeInTheDocument()
    expect(screen.getByLabelText('Status bar')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'PREVIEW' })).toHaveAttribute('aria-pressed', 'false')
  })

  it('opens an engine-projected, keyboard-operable table matrix with named controls', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const tableSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'table-columns') return { snapshot: tableSnapshot, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'right' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } }
      return { snapshot: tableSnapshot }
    })
    render(<App engine={engine(request)} initialSnapshot={tableSnapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
    const grid = await screen.findByRole('grid', { name: 'Table columns' })
    expect(grid).toHaveAttribute('aria-colcount', '11')
		expect(grid).toHaveAttribute('aria-rowcount', '2')
    const header = screen.getByRole('textbox', { name: 'Header for column 1' })
    header.focus(); fireEvent.keyDown(header, { key: 'ArrowRight' })
    expect(document.activeElement).toBe(screen.getByRole('spinbutton', { name: 'Width for column 1 in points' }))
    expect(screen.getByRole('button', { name: 'Move column 1 earlier' })).toBeDisabled()
		expect(screen.getByRole('button', { name: 'Move column 1 later' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Remove column 1' })).toBeInTheDocument()
  })

  it('traps the focused matrix, closes on Escape, and restores its invoking control', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const tableSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    render(<App engine={engine(vi.fn(async (operation: string) => operation === 'table-columns' ? { snapshot: tableSnapshot, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } } : { snapshot: tableSnapshot }))} initialSnapshot={tableSnapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    const invoker = screen.getByRole('button', { name: 'Configure columns' })
    invoker.focus(); fireEvent.click(invoker)
    const header = await screen.findByRole('textbox', { name: 'Header for column 1' })
    expect(document.activeElement).toBe(header)
    fireEvent.keyDown(header, { key: 'Tab' })
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Close Table Editor' }))
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).not.toBeInTheDocument())
    expect(document.activeElement).toBe(invoker)
  })

  it('admits a committed table snapshot after deselection and never reopens a closed session', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const first = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    const second = { documentState: 'loaded' as const, revision: 2, byteLength: 4, canvas: { ...tableCanvas, components: [{ ...tableCanvas.components[0]!, width: 144000 }] } }
    let releaseProjection!: () => void
    const delayedProjection = new Promise<{ snapshot: typeof second; tableColumns: { revision: number; table: { tableId: string; collection: string; alias: string; columns: { id: string; header: string; width: number; align: 'left'; binding: string; rowField: string; rowFieldEditable: boolean; footer: ''; footerOf: string; footerFormat: string }[] } } }>((resolve) => { releaseProjection = () => resolve({ snapshot: second, tableColumns: { revision: 2, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left', binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '', footerOf: '', footerFormat: '' }] } } }) })
    let queries = 0
    const request = vi.fn((operation: string) => {
      if (operation === 'table-columns') { queries++; return queries === 1 ? Promise.resolve({ snapshot: first, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } }) : delayedProjection }
      if (operation === 'command') return Promise.resolve({ snapshot: second })
      return Promise.resolve({ snapshot: first })
    })
    render(<App engine={engine(request)} initialSnapshot={first} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
    await screen.findByRole('button', { name: 'Add column after column 1' })
    fireEvent.click(screen.getByRole('button', { name: 'Add column after column 1' }))
    fireEvent.click(screen.getByRole('button', { name: 'Close Table Editor' }))
    releaseProjection()
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 2'))
    expect(screen.queryByRole('dialog', { name: 'Table Editor' })).not.toBeInTheDocument()
  })

  it('replaces the canvas with Preview, cancels an older render, and never dirties or installs its late PDF', async () => {
    let releaseSerialize!: (value: { snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }) => void
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } })
      if (operation === 'serialize') return new Promise<{ snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }>((resolve) => { releaseSerialize = resolve })
      if (operation === 'render') return Promise.resolve({ snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    expect(screen.queryByLabelText('Canvas region')).not.toBeInTheDocument()
    expect(screen.getByText('Rendering local PDF')).toBeInTheDocument()
    await waitFor(() => expect(request).toHaveBeenCalledWith('serialize', undefined, expect.any(AbortSignal)))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel and return to Design' }))
    releaseSerialize({ snapshot: snapshot(1), bytes })
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.queryByText(/Go production digest/)).not.toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(request.mock.calls.map(([operation]) => operation)).toEqual(['parameter-references', 'identity', 'serialize'])
  })

  it('coalesces manual and debounced rerenders behind one active FIFO operation', async () => {
    let releaseIdentity!: () => void
    let identityCalls = 0
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') {
        identityCalls++
        if (identityCalls === 1) return new Promise<{ snapshot: ReturnType<typeof snapshot>; preview: { revision: number; identity: string } }>((resolve) => { releaseIdentity = () => resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }) })
        return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'c'.repeat(64) } })
      }
      if (operation === 'serialize') return Promise.resolve({ snapshot: snapshot(1), bytes })
      if (operation === 'render') return Promise.resolve({ snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'c'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    vi.useFakeTimers()
    try {
      render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
      fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[1]}' } })
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[2]}' } })
      fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
      await vi.runAllTimersAsync()
      expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(1)
      releaseIdentity()
      await vi.runAllTimersAsync()
      await Promise.resolve(); await Promise.resolve(); await Promise.resolve()
      expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('uses the engine reference projection for keyboard-operable parameter inputs and retains accepted bytes through an invalid draft', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') return { snapshot: snapshot(1), parameterReferences: { revision: 1, names: ['reportDate'] } }
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const named = await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.change(named, { target: { value: '"2026-08-28T00:00:00Z"' } })
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(2))
    const accepted = request.mock.calls.filter(([operation]) => operation === 'identity').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    expect(new TextDecoder().decode(accepted[1].params)).toContain('2026-08-28T00:00:00Z')
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{ nope' } })
    expect(screen.getByRole('alert')).toHaveTextContent('last accepted parameter document remains in Preview')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
  })

  it('states pending, failed, and empty parameter discovery without inventing fields', async () => {
    let release!: () => void
    const pending = new Promise<{ snapshot: ReturnType<typeof snapshot>; parameterReferences: { revision: number; names: string[] } }>((resolve) => { release = () => resolve({ snapshot: snapshot(1), parameterReferences: { revision: 1, names: [] } }) })
    const request = vi.fn((operation: string) => operation === 'parameter-references' ? pending : Promise.resolve(operation === 'identity' ? { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } } : { snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    expect(screen.getByText('Discovering parameter references from the local engine…')).toBeInTheDocument()
    release()
    await screen.findByText('The local engine found no parameter references in this template.')
  })

  it('states a failed parameter projection rather than calling it an empty projection', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') throw new Error('worker unavailable')
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await screen.findByText('The local engine could not provide parameter references. The raw parameter document is still available.')
    expect(screen.queryByText('The local engine found no parameter references in this template.')).not.toBeInTheDocument()
  })

  it('edits named parameters without rewriting raw numeric lexemes or special own keys', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') return { snapshot: snapshot(1), parameterReferences: { revision: 1, names: ['__proto__', 'constructor', 'reportDate'] } }
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const reportDate = await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"constructor":1.00e+2,"__proto__":-0,"other":123.4500}' } })
    expect(screen.getByRole('textbox', { name: 'Value for params.constructor' })).toHaveValue('1.00e+2')
    expect(screen.getByRole('textbox', { name: 'Value for params.__proto__' })).toHaveValue('-0')
    reportDate.focus()
    fireEvent.change(reportDate, { target: { value: '"2026-08-28T00:00:00Z"' } })
    expect(document.activeElement).toBe(reportDate)
    fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'identity').length).toBeGreaterThan(1))
    const accepted = request.mock.calls.filter(([operation]) => operation === 'identity').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    const exact = '{"constructor":1.00e+2,"__proto__":-0,"other":123.4500,"reportDate":"2026-08-28T00:00:00Z"}'
    expect(new TextDecoder().decode(accepted[1].params)).toBe(exact)
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render').length).toBeGreaterThan(0))
    const rendered = request.mock.calls.filter(([operation]) => operation === 'render').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    expect(new TextDecoder().decode(rendered[1].params)).toBe(exact)
  })

  it('refreshes the engine reference projection after Undo while Preview remains open', async () => {
    let references = 0
    const historySnapshot = { ...snapshot(2), canUndo: false, canRedo: true }
    const initial = { ...snapshot(1), canUndo: true, canRedo: false }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') {
        references++
        return { snapshot: references === 1 ? initial : historySnapshot, parameterReferences: { revision: references === 1 ? 1 : 2, names: references === 1 ? ['reportDate'] : ['branch'] } }
      }
      if (operation === 'undo') return { snapshot: historySnapshot }
      if (operation === 'identity') return { snapshot: references > 1 ? historySnapshot : initial, preview: { revision: references > 1 ? 2 : 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: references > 1 ? historySnapshot : initial, bytes }
      if (operation === 'render') return { snapshot: references > 1 ? historySnapshot : initial, bytes: new Uint8Array([9]).buffer, preview: { revision: references > 1 ? 2 : 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: initial }
    })
    render(<App engine={engine(request)} initialSnapshot={initial} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.click(screen.getByRole('button', { name: /^Undo/ }))
    await screen.findByRole('textbox', { name: 'Value for params.branch' })
    expect(screen.queryByRole('textbox', { name: 'Value for params.reportDate' })).not.toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'parameter-references')).toHaveLength(2)
  })

  it('waits for matching PDF.js admission before claiming current exact output', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    expect(screen.getByText('LOCAL PDF PREVIEW')).toBeInTheDocument()
    expect(screen.queryByText('EXACT LOCAL PRODUCTION PDF')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    expect(screen.getByText('EXACT LOCAL PRODUCTION PDF')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Current exact local production PDF/ })).toHaveAttribute('aria-describedby', 'preview-freshness-status')
  })

  it('keeps producer diagnostics hidden and inert until their exact PDF is admitted, then revokes them on input change', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [{ severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'gone', dataPath: 'bands.content.gone', message: 'Content was clipped' }] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByLabelText('Render diagnostics')).toBeInTheDocument())
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[1]}' } })
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
  })

  it('returns to Design and announces an unavailable authoritative warning target without selecting a substitute', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [{ severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'gone', dataPath: '', message: 'Content was clipped' }] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Locate in Design' })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Locate in Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.getByText('Locate unavailable: the authoritative element is no longer present.')).toHaveAttribute('role', 'status')
  })

  it('returns from a path-only render failure without requiring an element id', async () => {
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', dataPath: 'items[0]', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') throw failure
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByLabelText('Local render failure')).toBeInTheDocument())
    fireEvent.click(within(screen.getByLabelText('Local render failure')).getByRole('button', { name: 'Return to Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
  })

  it('returns from a located render failure by selecting only the current projected element', async () => {
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const })
    const locatedCanvas = { ...canvas, components: [{ id: 'e7', type: 'rect' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: true }] }
    const locatedSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: locatedCanvas }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: locatedSnapshot, preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: locatedSnapshot, bytes }
      if (operation === 'render') throw failure
      return { snapshot: locatedSnapshot }
    })
    render(<App engine={engine(request)} initialSnapshot={locatedSnapshot} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const card = await screen.findByLabelText('Local render failure')
    fireEvent.click(within(card).getByRole('button', { name: 'Return to Design' }))
    const component = await screen.findByRole('button', { name: 'rect component e7' })
    expect(component).toHaveClass('canvas-component-selected')
    expect(screen.getByText('Selected e7 in Design.')).toHaveAttribute('role', 'status')
  })

  it('retries an active failed render through the existing scheduler without mutating the document', async () => {
    let renders = 0
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', dataPath: 'params.reportDate', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') { renders++; if (renders === 1) throw failure; return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const card = await screen.findByLabelText('Local render failure')
    expect(card).toHaveTextContent('RENDER_INVALID')
    fireEvent.click(within(card).getByRole('button', { name: 'Retry preview' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(2))
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('forces a fresh FIFO render after a same-identity last-good PDF failure and retains that PDF as stale', async () => {
    let renders = 0
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') {
        renders++
        if (renders > 1) throw failure
        return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const admitted = await screen.findByRole('button', { name: /Stale historical PDF/ })
    fireEvent.click(admitted)
    await waitFor(() => expect(screen.getByRole('button', { name: /Current exact local production PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Fail local PDF viewer' }))
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(screen.getByText(/local PDF viewer could not display/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
    const card = await screen.findByLabelText('Local render failure')
    expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument()
    const retry = within(card).getByRole('button', { name: 'Retry preview' })
    retry.focus()
    fireEvent.keyDown(retry, { key: 'Enter' })
    fireEvent.click(retry)
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(3))
    expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(3)
    expect(request.mock.calls.filter(([operation]) => operation === 'serialize')).toHaveLength(3)
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
    expect(screen.getByLabelText('Local render failure')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument()
  })

  it('revokes a delayed render failure after leaving Preview so its actions cannot reach Design', async () => {
    let rejectRender!: (error: Error) => void
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } })
      if (operation === 'serialize') return Promise.resolve({ snapshot: snapshot(1), bytes })
      if (operation === 'render') return new Promise<never>((_, reject: (error: Error) => void) => { rejectRender = reject })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel and return to Design' }))
    rejectRender(Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
  })

  it('names local file controls, persistent unsaved state, and offline availability', () => {
    render(<App />)
    const open = screen.getByRole('button', { name: 'Open local template' })
    expect(open).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Save local template' })).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Start blank' })).toBeDisabled()
    expect(screen.getByRole('status', { name: 'Offline availability' })).toHaveTextContent('Offline cache unavailable')
  })

  it('labels the development bypass instead of claiming a verified cache', () => {
    render(<App offlineState="dev-bypass" />)
    expect(screen.getByRole('status', { name: 'Offline availability' })).toHaveTextContent('Offline layer bypassed (dev)')
  })

  it('announces the checking, ready, and waiting-update lifecycle states', () => {
    const { rerender } = render(<App offlineState="checking" />)
    const status = screen.getByRole('status', { name: 'Offline availability' })
    expect(status).toHaveAttribute('aria-live', 'polite')
    expect(status).toHaveTextContent('Offline cache checking')
    rerender(<App offlineState="ready" />)
    expect(status).toHaveTextContent('Offline ready')
    rerender(<App offlineState="update-available" />)
    expect(status).toHaveTextContent('Update available; current release remains usable')
  })

  it('bypasses S1 when the current cache and engine are already ready', () => {
    render(<App loadState={{ state: 'ready', cacheReady: true, verifiedAssetUrls: [] }} engineState="starting" />)
    expect(screen.getByRole('status', { name: 'Engine preparation status' })).toHaveTextContent('Starting local engine')
    expect(screen.queryByRole('heading', { name: 'Preparing Folio' })).not.toBeInTheDocument()
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
  })

  it('loads only opaque adapter bytes through Go, establishes a clean baseline, and dirties after a committed command', async () => {
    const request = vi.fn(async (operation: string) => ({ snapshot: snapshot(operation === 'command' ? 8 : 7), ...(operation === 'serialize' ? { bytes } : {}) }))
    const files: FileAccess = { open: vi.fn(async () => ({ bytes, name: 'report.folio' })), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Open local template' }))
    await waitFor(() => expect(screen.getByText('report.folio')).toBeInTheDocument())
    expect(request.mock.calls.map(([operation]) => operation)).toEqual(['load', 'serialize'])
    expect(screen.getByText('Saved local file')).toBeInTheDocument()
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByText('Unsaved local changes')).toBeInTheDocument())
  })

  it('keeps zoom, grid, and snap local while an explicit Go page-setup command alone dirties the document', async () => {
    const request = vi.fn(async (operation: string) => ({ snapshot: snapshot(operation === 'command' ? 2 : 1), ...(operation === 'serialize' ? { bytes } : {}) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
    fireEvent.click(screen.getByRole('button', { name: 'Grid on' }))
    fireEvent.click(screen.getByRole('button', { name: 'Snap on' }))
    expect(request).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Grid off' })).toHaveAttribute('aria-pressed', 'false')
    expect(screen.getByRole('button', { name: 'Snap off' })).toHaveAttribute('aria-pressed', 'false')
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37.125' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request).toHaveBeenCalledTimes(1))
    expect(request.mock.calls[0]![0]).toBe('command')
    const command = request.mock.calls[0] as unknown as [string, ArrayBuffer]
    expect(new TextDecoder().decode(command[1])).toContain('"top":37.125')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('drives the full authoritative undo/redo depth, bounds, and divergent branch from engine snapshots', async () => {
    let revision = 1
    let undoDepth = 0
    let redoDepth = 0
    const historySnapshot = () => ({ ...snapshot(revision), canUndo: undoDepth > 0, canRedo: redoDepth > 0 })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'command') { revision++; undoDepth++; redoDepth = 0; return { snapshot: historySnapshot() } }
      if (operation === 'undo') { revision++; undoDepth--; redoDepth++; return { snapshot: historySnapshot() } }
      if (operation === 'redo') { revision++; undoDepth++; redoDepth--; return { snapshot: historySnapshot() } }
      return { snapshot: historySnapshot(), ...(operation === 'serialize' ? { bytes } : {}) }
    })
    render(<App engine={engine(request)} initialSnapshot={historySnapshot()} />)
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled())
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '38' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(2))
    fireEvent.click(screen.getByRole('button', { name: 'Undo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'undo')).toHaveLength(1))
    expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Redo' })).toBeEnabled()
    fireEvent.click(screen.getByRole('button', { name: 'Undo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'undo')).toHaveLength(2))
    expect(screen.getByRole('button', { name: 'Undo' })).toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: 'Redo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'redo')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(3))
    expect(screen.getByRole('button', { name: 'Redo' })).toBeDisabled()
  })

  it('keeps a no-op command non-dirty and out of browser history when the engine returns its stable snapshot', async () => {
    const request = vi.fn(async () => ({ snapshot: { ...snapshot(1), canUndo: false, canRedo: true } }))
    render(<App engine={engine(request)} initialSnapshot={{ ...snapshot(1), canUndo: false, canRedo: true }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(screen.getByRole('button', { name: 'Undo' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Redo' })).toBeEnabled()
  })

  it.each([
    [true, { save: '⌘S', undo: '⌘Z', redo: '⇧⌘Z', preview: '⌥P', snap: '⌥S' }],
    [false, { save: 'Ctrl+S', undo: 'Ctrl+Z', redo: 'Ctrl+Y', preview: 'Alt+P', snap: 'Alt+S' }],
  ])('uses one platform-normalized shortcut map (%s)', (mac, expected) => {
    expect(shortcutHintsFor(mac)).toMatchObject(expected)
  })

  it.each([
    ['property draft', { key: 'z', ctrlKey: true }],
    ['IME composition', { key: 'z', ctrlKey: true, isComposing: true }],
  ])('does not route Undo through an editable %s', (_name, keyboard) => {
    const request = vi.fn(async () => ({ snapshot: { ...snapshot(1), canUndo: false, canRedo: true } }))
    render(<App engine={engine(request)} initialSnapshot={{ ...snapshot(1), canUndo: true }} />)
    fireEvent.keyDown(screen.getByRole('textbox', { name: 'Top margin (pt)' }), keyboard)
    expect(request).not.toHaveBeenCalled()
  })

  it('offers only the five fixed palette components and sends an opaque Go placement command', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    expect(screen.getAllByRole('button', { name: /Place / }).map((button) => button.getAttribute('aria-label'))).toEqual(['Place Text', 'Place Image', 'Place Table', 'Place Line', 'Place Rectangle'])
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    const [operation, payload] = request.mock.calls[0] as unknown as [string, ArrayBuffer]
    expect(operation).toBe('command')
    expect(new TextDecoder().decode(payload)).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })

  it('converts a local band pointer position through the shared display mapping before proposing placement', () => {
    const localX = ['offset', 'X'].join('')
    const localY = ['offset', 'Y'].join('')
    expect(placementPoint({ [localX]: 120, [localY]: 40 } as unknown as MouseEvent, canvas.bands[1]!, 1)).toEqual({ x: 156, y: 96 })
  })

  it('keeps selection local and deletes one unambiguous selected component through Go', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e9'))
    expect(request).not.toHaveBeenCalled()
    const region = screen.getByLabelText('Canvas region')
    region.focus()
    fireEvent.keyDown(region, { key: 'Delete' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"deleteComponent","version":1,"id":"e9"}')
  })

  it('does not send a move for a pointer selection, but commits one point-valued move after a real drag', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 10, clientY: 10 })
    expect(request).not.toHaveBeenCalled()
    fireEvent.pointerDown(component, { pointerId: 2, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 2, clientX: 13, clientY: 12 })
    fireEvent.pointerUp(component, { pointerId: 2, clientX: 13, clientY: 12 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"moveComponent","version":1,"id":"e9","x":3,"y":2,"snap":true}')
  })

  it('tracks a drag and a resize live in the geometry fields, then lands the accepted engine geometry', async () => {
    const placed = { id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true }
    const componentCanvas = { ...canvas, components: [placed] }
    const movedCanvas = { ...canvas, components: [{ ...placed, x: 6_000, y: 4_000 }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: movedCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.click(component)
    const x = screen.getByRole('textbox', { name: 'X (pt)' })
    const y = screen.getByRole('textbox', { name: 'Y (pt)' })
    expect(x).toHaveValue('0')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 13, clientY: 12 })
    // The transient proposal the canvas paints is the value the panel shows,
    // and it cannot be typed over while the pointer owns it.
    expect(x).toHaveValue('3')
    expect(y).toHaveValue('2')
    expect(x).toHaveAttribute('readonly')
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 19, clientY: 10 })
    expect(x).toHaveValue('9')
    expect(y).toHaveValue('0')
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 19, clientY: 10 })
    // Go's accepted geometry replaces the proposal; 9 was never committed.
    await waitFor(() => expect(x).toHaveValue('6'))
    expect(y).toHaveValue('4')
    expect(x).not.toHaveAttribute('readonly')

    const width = screen.getByRole('textbox', { name: 'Width (pt)' })
    fireEvent.pointerDown(screen.getByRole('button', { name: 'Resize e9' }), { pointerId: 2, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(screen.getByRole('button', { name: 'Resize e9' }), { pointerId: 2, clientX: 8, clientY: 5 })
    expect(width).toHaveValue('80')
    expect(screen.getByRole('textbox', { name: 'Height (pt)' })).toHaveValue('29')
  })

  it('keeps the drop proposal painted until the engine geometry lands', async () => {
    const placed = { id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true }
    const componentCanvas = { ...canvas, components: [placed] }
    const movedCanvas = { ...canvas, components: [{ ...placed, x: 6_000, y: 4_000 }] }
    let answer: (() => void) | undefined
    const request = vi.fn(async () => {
      await new Promise<void>((resolve) => { answer = resolve })
      return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: movedCanvas } }
    })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 19, clientY: 10 })
    expect(component.style.getPropertyValue('--component-x')).toBe('9px')
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 19, clientY: 10 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // Red proof: dropping the proposal on pointer-up painted 0px here, so the
    // element visibly jumped back to where the drag started and stayed there
    // until Go answered. The proposal owns the paint until the answer lands.
    expect(component.style.getPropertyValue('--component-x')).toBe('9px')
    answer!()
    await waitFor(() => expect(component.style.getPropertyValue('--component-x')).toBe('6px'))
    expect(screen.getByRole('textbox', { name: 'X (pt)' })).not.toHaveAttribute('readonly')
  })

  it('toggles Shift-click selection once without engine traffic and clears it on an empty canvas click', () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }, { id: 'e2', type: 'rect' as const, band: 'content' as const, x: 80000, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.pointerDown(screen.getByLabelText('text component e1'), { pointerId: 1, clientX: 1, clientY: 1 })
    fireEvent.pointerUp(screen.getByLabelText('text component e1'), { pointerId: 1, clientX: 1, clientY: 1 })
    fireEvent.pointerDown(screen.getByLabelText('rect component e2'), { pointerId: 2, clientX: 1, clientY: 1, shiftKey: true })
    fireEvent.pointerUp(screen.getByLabelText('rect component e2'), { pointerId: 2, clientX: 1, clientY: 1, shiftKey: true })
    expect(screen.getByLabelText('Resize e1')).toBeInTheDocument()
    expect(screen.getByLabelText('Resize e2')).toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('Report page with Page Header, Content, and Page Footer'))
    expect(screen.queryByLabelText('Resize e1')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Resize e2')).not.toBeInTheDocument()
    expect(request).not.toHaveBeenCalled()
  })

  it('leaves a dirty session untouched for an open cancellation or failure', async () => {
    const files: FileAccess = { open: vi.fn().mockRejectedValueOnce(new FileAccessCancelled()).mockRejectedValueOnce(new Error('denied')), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine()} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 2, byteLength: 3 }} />)
    const open = screen.getByRole('button', { name: 'Open local template' })
    fireEvent.click(open)
    await waitFor(() => expect(files.open).toHaveBeenCalledTimes(1))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    fireEvent.click(open)
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Could not open local file'))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('clears temporary busy wording after save cancellation without changing the session', async () => {
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget: vi.fn(async () => { throw new FileAccessCancelled() }), writeSave: vi.fn() }
    render(<App engine={engine()} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 2, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Save As' }))
    await waitFor(() => expect(files.acquireSaveTarget).toHaveBeenCalledOnce())
    expect(screen.queryByText(/Preparing Save As/)).not.toBeInTheDocument()
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('acquires a target before serialization, preserves dirty on failure, and handles the Save shortcut', async () => {
    let rejectSave = true
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 3, byteLength: 3 }, bytes }))
    const acquireSaveTarget = vi.fn(async () => ({ name: 'untitled.folio' }))
    const writeSave = vi.fn(async () => { if (rejectSave) throw new Error('denied'); return { name: 'untitled.folio' } })
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget, writeSave }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 3, byteLength: 3 }} />)
    fireEvent.keyDown(window, { key: 's', ctrlKey: true })
    await waitFor(() => expect(writeSave).toHaveBeenCalledTimes(1))
    expect(acquireSaveTarget).toHaveBeenCalledBefore(request)
    expect(request).toHaveBeenCalledBefore(writeSave)
    expect(screen.getByRole('alert')).toHaveTextContent('Could not save local file')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    rejectSave = false
    fireEvent.click(screen.getByRole('button', { name: 'Save local template' }))
    await waitFor(() => expect(screen.getByText('Downloaded local file untitled.folio')).toBeInTheDocument())
    expect(screen.getByText('Saved local file')).toBeInTheDocument()
  })

  it('routes Start blank through the engine and returns to an unnamed unsaved local workspace', async () => {
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 9, byteLength: 3 }, bytes }))
    render(<App engine={engine(request)} fileAccess={{ open: vi.fn(), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }} blankBytes={bytes} initialSnapshot={{ documentState: 'loaded', revision: 4, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(request).toHaveBeenCalledWith('load', bytes))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('keeps a noncanonical valid open dirty until the canonical engine bytes are written', async () => {
    const canonical = new Uint8Array([9, 8, 7]).buffer
    const request = vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: 7, byteLength: 3 }, ...(operation === 'serialize' ? { bytes: canonical } : {}) }))
    const files: FileAccess = { open: vi.fn(async () => ({ bytes, name: 'noncanonical.folio' })), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Open local template' }))
    await waitFor(() => expect(screen.getByText(/canonical local changes need saving/)).toBeInTheDocument())
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('does not roll back or clean a newer engine revision after an older save settles', async () => {
    let releaseWrite: (() => void) | undefined
    let releaseCommit: (() => void) | undefined
    const writeSave = vi.fn(() => new Promise<{ name: string }>((resolve) => { releaseWrite = () => resolve({ name: 'untitled.folio' }) }))
    const request = vi.fn((operation: string): Promise<{ snapshot: ReturnType<typeof snapshot>; bytes?: ArrayBuffer }> => {
      if (operation === 'command') return new Promise((resolve) => { releaseCommit = () => resolve({ snapshot: snapshot(3) }) })
      return Promise.resolve({ snapshot: snapshot(2), bytes })
    })
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget: vi.fn(async () => ({ name: 'untitled.folio' })), writeSave }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={snapshot(2)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    fireEvent.click(screen.getByRole('button', { name: 'Save local template' }))
    await waitFor(() => expect(writeSave).toHaveBeenCalledOnce())
    releaseCommit!()
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 3'))
    releaseWrite!()
    await waitFor(() => expect(screen.getByText('Unsaved local changes')).toBeInTheDocument())
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 3')
    expect(screen.getByText(/newer local changes need saving/)).toBeInTheDocument()
  })

  it('paints each Go band rectangle at its projected origin and uses one zoomed display scale for page and grid', () => {
    render(<App initialSnapshot={snapshot(1)} />)
    const page = screen.getByLabelText('Report page with Page Header, Content, and Page Footer')
    const header = screen.getByLabelText('Page Header')
    expect(page.style.getPropertyValue('--page-display-width')).toBe('595.276px')
    expect(page.style.getPropertyValue('--page-display-height')).toBe('841.89px')
    expect(page.style.getPropertyValue('--grid-display-pitch')).toBe('6px')
    expect(header.style.getPropertyValue('--band-x')).toBe('36px')
    expect(header.style.getPropertyValue('--band-y')).toBe('36px')
    expect(header.style.getPropertyValue('--band-width')).toBe('523.276px')
    expect(header.style.getPropertyValue('--band-height')).toBe('20px')
    fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
    expect(page.style.getPropertyValue('--page-display-width')).toBe('654.8036px')
    expect(page.style.getPropertyValue('--grid-display-pitch')).toBe('6.6px')
  })

  it('paints only pre-broken engine text lines without changing local document state', () => {
    const textCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true, value: 'do not paint this value', textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 12000, advance: 16000, width: 24000, fragments: [{ text: 'engine ', x: 0 }, { text: 'line', x: 16000 }] }] } }] }
    const request = vi.fn()
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} />)
    expect(screen.getByText('engine', { exact: true })).toBeInTheDocument()
    expect(screen.getByText('line', { exact: true })).toBeInTheDocument()
    expect(screen.queryByText('do not paint this value')).not.toBeInTheDocument()
    expect(screen.getByLabelText('text component e1: engine line')).toBeInTheDocument()
    expect(document.querySelector('.canvas-text-line')).toHaveStyle({ '--text-line-baseline': '12px', '--text-line-advance': '16px' })
    expect(request).not.toHaveBeenCalled()
  })

  // STORY 8.4a. THE FACE THE DOCUMENT ITSELF CARRIES, PAINTED WITH.
  //
  // The engine measures and renders with a face out of the document's own
  // `assets` map and attributes each painted fragment to the asset it resolved
  // to. Without this the browser had no CSS family for such a face at all, so
  // the canvas rasterized at the engine's x-positions with a fallback's
  // metrics and the glyphs collided — the reported defect, rebuilt for a
  // document that carries its own typeface.
  //
  // THE PAGE'S FONT SET IS STUBBED because jsdom implements none: it defines
  // no `FontFace` and no font set on `document`, so both are installed with
  // `Object.defineProperty` (neither exists to be assigned over). Nothing here
  // measures anything, and the seam under test computes no metric either —
  // every x, advance and line break in the fixture is the engine's.
  it('registers a carried face once for the whole document and paints its fragments with the family derived from its key', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation === 'asset') { requested.push(new TextDecoder().decode(payload)); return { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer } }
        return { snapshot: snapshot(1) }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: carriedFaceCanvas(key) }} />)
      // ONE REQUEST FOR THE WHOLE DOCUMENT, not one per component. The nearest
      // precedent, ImagePaint, is mounted once per component AND once per
      // repeated sheet; copying that lifetime here would add one face many
      // times under a single global family name and let an unmounting instance
      // delete a face another is still painting with. The fixture has two text
      // components drawing through the same carried entry precisely so a
      // per-component lifetime would show up as two requests.
      await waitFor(() => expect(requested).toEqual([key]))
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      expect(painted().length).toBe(2)
      await waitFor(() => expect(painted().map((node) => node.style.fontFamily)).toEqual([embeddedFaceFamily(key), embeddedFaceFamily(key)]))
      // AND THE ENGINE'S OWN GEOMETRY IS UNTOUCHED BY IT (AD-17): the fragment
      // still paints at the x the engine measured, and the family is the only
      // thing this story added to it.
      expect(painted()[0]).toHaveStyle({ '--text-fragment-x': '0px' })
      expect(screen.getAllByLabelText(/text component e1/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // EVERY CONTENT WINDOW AFTER THE FIRST IS AN ECHO, so a document that runs
  // past one sheet paints most of itself through ComponentEcho rather than
  // through the home occurrence. A carried face that reached only the home
  // would leave sheets 2..N rasterizing at the ENGINE'S x-positions on the
  // fallback stack — precisely the collision this story exists to remove, on
  // most of the pages.
  //
  // MUTATION PROOF, RUN AND RECORDED: replacing `carriedFaces={carriedFaces}`
  // with `carriedFaces={NO_CARRIED_FACES}` in ComponentEcho reddens THIS test
  // on the echoed fragment and leaves every other designer test green — which
  // is why the assertion is written against the echoed node specifically and
  // not against the container's fragments as a set.
  it('paints a repeated sheet\'s echo with the carried face, not only the component\'s home', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const request = vi.fn(async (operation: string) => operation === 'asset'
        ? { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer }
        : { snapshot: snapshot(1) })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: carriedFaceEchoCanvas(key) }} />)
      const home = () => Array.from(view.container.querySelectorAll('.canvas-component:not(.canvas-component-echo) .canvas-text-fragment')) as HTMLElement[]
      const echoed = () => Array.from(view.container.querySelectorAll('.canvas-component-echo .canvas-text-fragment')) as HTMLElement[]
      // The fixture really does produce both, or the claim below is vacuous.
      expect(home()).toHaveLength(1)
      expect(echoed()).toHaveLength(1)
      await waitFor(() => expect(echoed()[0]!.style.fontFamily).toBe(embeddedFaceFamily(key)))
      expect(home()[0]!.style.fontFamily).toBe(embeddedFaceFamily(key))
      // AD-17 on the echo too: the engine's own x is what it paints at, and
      // the family is the only thing this story put on it.
      expect(echoed()[0]!).toHaveStyle({ '--text-fragment-x': '0px' })
    } finally {
      fontSet.restore()
    }
  })

  // THE DEGRADE PATH, stated as a claim about the SESSION. An inline family
  // REPLACES the stylesheet rule rather than extending it, so asking for a
  // family whose bytes never arrived would take the fragment off the declared
  // stack onto whatever the browser defaults to. A failed asset request must
  // therefore leave the fragment exactly as a shipped-face fragment: painted,
  // named, and on the declared stack — and it must never reach the engine's
  // failure channel.
  //
  // IT IS READ AFTER THE CHAIN HAS SETTLED, WHICH IS THE WHOLE DIFFICULTY.
  // "The fragment has no family" is ALSO true of a registration that simply
  // has not finished, so asserting it the instant the request was issued
  // proves "not yet", not "degraded" — the identical assertion passes at that
  // instant in the SUCCESS case. The document therefore carries TWO carried
  // faces: one whose bytes are withheld until the other has already been
  // REJECTED, so the first face's arrival in the font set is a positive
  // condition that cannot hold until the failure was handled.
  //
  // MUTATION PROOF, RUN AND RECORDED: returning bytes for `unfetchable`
  // instead of throwing reddens this test — the second fragment acquires its
  // family and the font set holds two families, not one.
  it('keeps painting on the declared stack when a carried face\'s bytes cannot be fetched', async () => {
    const fetchable = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const unfetchable = 'fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210'
    const fontSet = installStubFontSet()
    try {
      let release: () => void = () => undefined
      const withheld = new Promise<ArrayBuffer>((resolve) => { release = () => resolve(new Uint8Array([0, 1, 2, 3]).buffer) })
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation !== 'asset') return { snapshot: snapshot(1) }
        const key = new TextDecoder().decode(payload)
        requested.push(key)
        if (key === unfetchable) throw Object.assign(new Error('no such asset'), { code: 'ASSET_UNAVAILABLE' })
        return { snapshot: snapshot(1), bytes: await withheld }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: twoCarriedFacesCanvas(fetchable, unfetchable) }} />)
      // Both were asked for; the unfetchable one has already rejected, because
      // its bytes were never withheld behind anything.
      await waitFor(() => expect([...requested].sort()).toEqual([fetchable, unfetchable].sort()))
      release()
      // THE POSITIVE CONDITION. The font set cannot hold the fetchable face
      // until its bytes were released, which happened after the other request
      // rejected — so everything below is read on a settled chain.
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(fetchable)]))
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      expect(painted().length).toBe(2)
      await waitFor(() => expect(painted()[0]!.style.fontFamily).toBe(embeddedFaceFamily(fetchable)))
      expect(painted()[1]!.style.fontFamily).toBe('')
      expect(painted()[1]!).toHaveStyle({ '--text-fragment-x': '0px' })
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getAllByLabelText(/text component e1/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // THE KEY IS A STRING FROM THE DOCUMENT, AND IT BECOMES A CSS FAMILY. The
  // projection admits a FRAGMENT's `assetKey` only as 64 lowercase hex, but a
  // CHAIN ENTRY's key — the one this effect fetches bytes for and derives the
  // family from — is admitted on length alone, so the shape is asserted at the
  // derivation. A key that is not one is a carried face the browser declines:
  // no request, no family, no registration, and the fragment stays on the
  // stylesheet's declared stack. It is a DEGRADE and not a refusal — nothing
  // reaches the failure channel and the worker is untouched.
  it('declines a chain entry whose asset key is not an asset key, rather than turning it into a family', async () => {
    const wellFormed = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const injected = 'IBM Plex Sans, monospace'
    const fontSet = installStubFontSet()
    try {
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation === 'asset') { requested.push(new TextDecoder().decode(payload)); return { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer } }
        return { snapshot: snapshot(1) }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: twoCarriedFacesCanvas(wellFormed, injected) }} />)
      // The well-formed sibling settles, so the claim about the malformed one
      // is read after the effect has done everything it is going to do.
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(wellFormed)]))
      expect(requested).toEqual([wellFormed])
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      await waitFor(() => expect(painted()[0]!.style.fontFamily).toBe(embeddedFaceFamily(wellFormed)))
      // Not a family, not a partial family, not the string itself.
      expect(painted()[1]!.style.fontFamily).toBe('')
      expect(view.container.innerHTML).not.toContain(embeddedFaceFamily(injected))
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getAllByLabelText(/text component e2/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // THE LISTING IS PART OF THE EFFECT'S KEY, AND THIS IS THE CASE THAT PROVES
  // IT. `documentGenerationValue` advances only when the document is REPLACED
  // — open a file, new template, undo/redo — and a font-chain command is none
  // of those: applyFontChain commits through setCurrentSnapshot without
  // clearing the document interaction, so the generation does not move. The
  // author can nonetheless remove the entry that carries the face, from a
  // control that is right there in the panel.
  //
  // Release is invisible in the DOM — the fragment simply stops asking for the
  // family — so the claim is made against the font set's own record.
  //
  // MUTATION PROOF, RUN AND RECORDED: reducing the effect's dependency array
  // to `[engine, documentGenerationValue]` reddens this test and leaves the
  // rest of the designer suite green.
  it('releases a carried face when a chain edit drops it, though the document was never replaced', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const carrying = carriedFaceCanvas(key)
      const dropped = { ...carrying, fontChains: [{ name: 'body', entries: [face('Noto Sans')] }] }
      const loaded = (revision: number, projection: typeof carrying) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas: projection })
      const request = vi.fn(async (operation: string) => {
        if (operation === 'asset') return { snapshot: loaded(1, carrying), bytes: new Uint8Array([0, 1, 2, 3]).buffer }
        if (operation === 'command') return { snapshot: loaded(2, dropped) }
        return { snapshot: loaded(1, carrying) }
      })
      render(<App engine={engine(request) } initialSnapshot={loaded(1, carrying)} />)
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(key)]))
      expect(fontSet.removed).toEqual([])
      fireEvent.click(screen.getAllByLabelText(/text component e1/)[0]!)
      fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
      // The carried entry is the second of the one chain, and this is the
      // author-reachable control that drops it.
      fireEvent.click(screen.getByRole('button', { name: 'Remove entry 2 of font chain 1' }))
      await waitFor(() => expect(screen.queryByRole('button', { name: 'Remove entry 2 of font chain 1' })).not.toBeInTheDocument())
      await waitFor(() => expect(fontSet.removed).toEqual([embeddedFaceFamily(key)]))
      // And it was released rather than re-registered: the document was never
      // replaced, so nothing asked for those bytes a second time.
      expect(fontSet.added).toEqual([embeddedFaceFamily(key)])
      expect(document.querySelectorAll('.canvas-text-fragment')).toHaveLength(2)
      expect(Array.from(document.querySelectorAll('.canvas-text-fragment')).map((node) => (node as HTMLElement).style.fontFamily)).toEqual(['', ''])
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  it('retains literal empty drafts, announces the precise engine diagnostic, and ignores a stale Apply draft reset', async () => {
    let resolveApply: ((value: { snapshot: ReturnType<typeof snapshot> }) => void) | undefined
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: ReturnType<typeof snapshot> }>((resolve) => { resolveApply = resolve }) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    const top = screen.getByRole('textbox', { name: 'Top margin (pt)' })
    fireEvent.change(top, { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    fireEvent.change(top, { target: { value: '38' } })
    resolveApply!({ snapshot: snapshot(2) })
    await waitFor(() => expect(top).toHaveValue('38'))
    request.mockRejectedValueOnce(Object.assign(new Error('must not be negative'), { code: 'PAGE_SETUP_INVALID', dataPath: 'page.margin.top' }))
    fireEvent.change(top, { target: { value: '' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('page.margin.top: must not be negative'))
    expect(top).toHaveValue('')
  })

  it('keeps component drafts local, sends exactly one Enter/Blur commit, and locates a Go diagnostic', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello', fontFamily: 'body', fontSize: 12_000, borderEdges: ['bottom' as const] }] }
    const request = vi.fn((operation: string) => operation === 'command' ? Promise.reject(Object.assign(new Error('must fit the content band'), { elementId: 'e1', dataPath: 'component.x' })) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const x = screen.getByRole('textbox', { name: 'X (pt)' })
    fireEvent.change(x, { target: { value: '9999' } })
    expect(request).not.toHaveBeenCalled()
    fireEvent.keyDown(x, { key: 'Enter' })
    fireEvent.blur(x)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('e1: component.x: must fit the content band'))
    expect(x).toHaveValue('9999')
    expect(x).toHaveAttribute('aria-invalid', 'true')
  })

  it('routes the one CONTENT field to the value or expression command by what was typed', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    expect(screen.queryByRole('textbox', { name: 'Text expression' })).not.toBeInTheDocument()
    const field = screen.getByRole('textbox', { name: 'Text' })
    // Story 7.4: CONTENT is a textarea, so Enter puts a line feed in the
    // draft and commits NOTHING. Blur is this field's commit; every other
    // field keeps Enter.
    fireEvent.change(field, { target: { value: 'Customer: {{customer.name}}' } })
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(sent).toHaveLength(0)
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"expression":{"op":"set","value":"Customer: {{customer.name}}"}}}')
    const again = screen.getByRole('textbox', { name: 'Text' })
    fireEvent.change(again, { target: { value: 'Plain heading' } })
    fireEvent.blur(again)
    await waitFor(() => expect(sent).toHaveLength(2))
    expect(new TextDecoder().decode(sent[1]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"Plain heading"}}}')
  })

  it('marks the expression-bearing fields with fx, and lights it once the field holds one', () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const text = screen.getByRole('textbox', { name: 'Text' })
    const marker = (input: HTMLElement) => input.parentElement!.querySelector('.property-fx')
    expect(text).toHaveAttribute('aria-description', 'Accepts literal text, or {{ }} expressions')
    expect(marker(text)).toHaveTextContent('fx')
    expect(marker(text)).not.toHaveClass('property-fx-active')
    fireEvent.change(text, { target: { value: 'Customer: {{customer.name}}' } })
    expect(marker(text)).toHaveClass('property-fx-active')
    // A geometry field is a literal to Go, which rejects a placeholder in it:
    // it must carry no cue at all.
    expect(marker(screen.getByRole('textbox', { name: 'X (pt)' }))).toBeNull()
    const visible = screen.getByRole('textbox', { name: 'Visible if' })
    expect(visible).toHaveAttribute('aria-description', 'Accepts a boolean data path or call, written without {{ }} — the grammar has no comparisons')
    expect(marker(visible)).not.toHaveClass('property-fx-active')
    fireEvent.change(visible, { target: { value: 'customer.isActive' } })
    expect(marker(visible)).toHaveClass('property-fx-active')
  })

  it('authors BOX colours through the picker, states pt on an empty size, and drops the padding rows', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    // An unset size still says which unit it wants.
    const border = screen.getByRole('textbox', { name: 'Border width (pt)' })
    expect(border).toHaveValue('')
    expect(border.parentElement).toHaveTextContent('pt')
    for (const label of ['Padding top (pt)', 'Padding right (pt)', 'Padding bottom (pt)', 'Padding left (pt)']) expect(screen.queryByRole('textbox', { name: label })).not.toBeInTheDocument()
    const picker = screen.getByLabelText('Pick Border colour')
    fireEvent.change(picker, { target: { value: '#c81e1e' } })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"borderColor":{"op":"set","value":"#c81e1e"}}}')
    fireEvent.change(screen.getByLabelText('Pick Background'), { target: { value: '#0b1120' } })
    await waitFor(() => expect(sent).toHaveLength(2))
    expect(new TextDecoder().decode(sent[1]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"#0b1120"}}}')
  })

  it('paints the engine-projected box on the canvas, and nothing where none is projected', () => {
    const boxed = { id: 'e1', type: 'rect' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, background: '#1b2a4a', borderWidth: 2_000, borderColor: '#c81e1e', borderEdges: ['bottom' as const] }
    const plain = { id: 'e2', type: 'text' as const, band: 'content' as const, x: 0, y: 40_000, width: 72_000, height: 24_000, resizable: true, value: 'Plain' }
    render(<App initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [boxed, plain] } }} />)
    const box = screen.getByLabelText('rect component e1').querySelector('.canvas-box') as HTMLElement
    expect(box).not.toBeNull()
    expect(box.style.background).toBe('rgb(27, 42, 74)')
    expect(box.style.borderBottom).toBe('2px solid rgb(200, 30, 30)')
    // An edge the engine does not paint is not painted here either.
    expect(box.style.borderTop).toBe('0px')
    expect(screen.getByLabelText('text component e2').querySelector('.canvas-box')).toBeNull()
  })

  it('sets the text colour from TYPOGRAPHY and paints the canvas in it', async () => {
    const inked = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello', color: '#c81e1e', textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 10_000, advance: 12_000, width: 30_000, fragments: [{ text: 'Hello', x: 0 }] }] } }
    const sent: ArrayBuffer[] = []
    const componentCanvas = { ...canvas, components: [inked] }
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    // The canvas paints the engine's ink, never a browser-side default.
    const element = screen.getByLabelText('text component e1: Hello')
    const paint = element.querySelector('.canvas-text-paint') as HTMLElement
    expect(paint.style.getPropertyValue('--text-ink')).toBe('#c81e1e')
    fireEvent.click(element)
    fireEvent.change(screen.getByLabelText('Pick Text colour'), { target: { value: '#1b2a4a' } })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"color":{"op":"set","value":"#1b2a4a"}}}')
  })

  it('keeps a newer property draft through an unrelated successful snapshot and exposes table truth', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }, { id: 'e2', type: 'table' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }] }
    let resolve: ((value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }) => void) | undefined
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }>((done) => { resolve = done }) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const value = screen.getByRole('textbox', { name: 'Text' })
    fireEvent.change(value, { target: { value: 'newer literal' } })
    fireEvent.blur(value)
    resolve!({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 3, canvas: componentCanvas } })
    await waitFor(() => expect(value).toHaveValue('newer literal'))
    fireEvent.click(screen.getByLabelText('table component e2'))
    expect(screen.queryByRole('textbox', { name: 'Width (pt)' })).not.toBeInTheDocument()
    expect(screen.getByText('Table binding: transactions[] (display only)')).toBeInTheDocument()
  })
})

describe('typography controls over the engine-projected closed sets', () => {
  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  const select = (request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } }))) => {
    const componentCanvas = { ...canvas, components: [textComponent] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    return request
  }

  it('offers the document\'s declared font chains, searched, and commits the chosen one', async () => {
    const request = select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    const listed = () => within(screen.getByRole('listbox', { name: 'Declared fonts' })).queryAllByRole('option').map((option) => option.textContent)
    expect(listed()).toEqual(['body', 'heading'])
    fireEvent.change(combobox, { target: { value: 'head' } })
    expect(listed()).toEqual(['heading'])
    fireEvent.click(screen.getByRole('option', { name: 'heading' }))
    await waitFor(() => expect(request).toHaveBeenCalledWith('command', expect.anything()))
    // The typed search text is never a value: only a listed family is sent.
    expect(screen.queryByRole('listbox', { name: 'Declared fonts' })).not.toBeInTheDocument()
  })

  it('states a search that matches no declared chain instead of offering to invent one', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    fireEvent.change(combobox, { target: { value: 'Helvetica' } })
    expect(screen.queryByRole('option')).not.toBeInTheDocument()
    expect(screen.getByText('No declared font matches "Helvetica".')).toBeInTheDocument()
  })

  it('shows the engine\'s own default size for an element that commits none, as a placeholder and not a value', () => {
    select()
    const size = screen.getByRole('textbox', { name: 'Font size (pt)' })
    expect(size).toHaveValue('')
    expect(size).toHaveAttribute('placeholder', '12')
  })

  it('commits alignment from the closed sets, and clears it by pressing the active segment again', async () => {
    const request = select()
    for (const name of ['Align left', 'Align center', 'Align right', 'Align justify', 'Vertical align top', 'Vertical align middle', 'Vertical align bottom']) {
      expect(screen.getByRole('button', { name })).toHaveAttribute('aria-pressed', 'false')
    }
    fireEvent.click(screen.getByRole('button', { name: 'Align center' }))
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
  })

  // Story 7.4 / AC3. `style.align` admits `justify` for a text element, and
  // has since 7.3; a table's cells draw a justified value at their start
  // edge, so the value is meaningless there and the control must not offer
  // it. A MIXED selection is the case that decides the rule: one command goes
  // to every id, so the segment is offered only when EVERY selected component
  // is text.
  it('offers justify for text alone, and never for a table or a mixed selection', async () => {
    const table = { id: 'e2', type: 'table' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    const componentCanvas = { ...canvas, components: [textComponent, table] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)

    const alignNames = () => within(screen.getByRole('group', { name: 'Align' })).getAllByRole('button').map((button) => button.getAttribute('aria-label'))
    fireEvent.click(screen.getByLabelText('text component e1'))
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right', 'Align justify'])
    // The glyph is an SVG path, never a CSS declaration asking the browser to
    // justify: canvas-authority-contract.test.ts bans that across every
    // production, unit and e2e source.
    expect(screen.getByRole('button', { name: 'Align justify' }).querySelector('svg path')).toBeInTheDocument()

    fireEvent.click(screen.getByLabelText('table component e2'))
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right'])

    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByLabelText('table component e2'), { shiftKey: true })
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right'])

    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Align justify' }))
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"align":{"op":"set","value":"justify"}}}')
  })

  // Story 7.4 / AC4. lineSpacing is a dimensionless RATIO carried as a RAW,
  // UNQUOTED JSON number: Go's own decoder performs the x1000 to thousandths,
  // exactly as it does for a value written in a .folio file. Quoting it, or
  // pre-multiplying it here, is refused by the engine.
  it('shows an unset ratio as a placeholder and commits a typed one as a raw unquoted number', async () => {
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const leading = screen.getByRole('textbox', { name: 'Line spacing' })
    // Empty is the leading the declared chain itself rules — a ratio of 1 —
    // shown as a placeholder and never as a value.
    expect(leading).toHaveValue('')
    expect(leading).toHaveAttribute('placeholder', '1')
    expect(leading).toHaveAttribute('inputMode', 'decimal')
    fireEvent.change(leading, { target: { value: '1.5' } })
    fireEvent.keyDown(leading, { key: 'Enter' })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":1.5}}}')
  })

  it('reads a committed line spacing back in the author\'s units and clears it from the same row', async () => {
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    const spaced = { ...canvas, components: [{ ...textComponent, lineSpacing: 1_500 }] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: spaced }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    // The engine carries thousandths; the author is shown the ratio.
    expect(screen.getByRole('textbox', { name: 'Line spacing' })).toHaveValue('1.5')
    fireEvent.click(screen.getByRole('button', { name: 'Clear Line spacing' }))
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"clear"}}}')
  })

  // The mock reproduces the engine's REAL rejection, measured by running the
  // command through Go: applyPropertyChanges prefixes the command key and
  // template's validator words its reason in terms of that same key, so the
  // message the browser receives really does begin `lineSpacing: lineSpacing`.
  // ComponentCommandError carries it verbatim, with DataPath
  // `component.lineSpacing` on element `e1`; the Go half is pinned by
  // TestLineSpacingPropertyCommandDecodesThroughTheOneLoaderValidator.
  it('shows the engine\'s located line-spacing refusal and keeps the author\'s text', async () => {
    const engineMessage = 'lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); 0 is outside that range'
    const request = vi.fn((operation: string) => operation === 'command'
      ? Promise.reject(Object.assign(new Error(engineMessage), { elementId: 'e1', dataPath: 'component.lineSpacing' }))
      : Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3 } }))
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const leading = screen.getByRole('textbox', { name: 'Line spacing' })
    fireEvent.change(leading, { target: { value: '0' } })
    fireEvent.keyDown(leading, { key: 'Enter' })
    // The WHOLE located sentence, not a prefix of it: the element, the field
    // it was located to, and the engine's own reason including the offending
    // value the author typed.
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(`e1: component.lineSpacing: ${engineMessage}`))
    expect(leading).toHaveValue('0')
    expect(leading).toHaveAttribute('aria-invalid', 'true')
  })

  it('presses the segment the engine has committed, and clears it from the same control', async () => {
    const aligned = { ...canvas, components: [{ ...textComponent, align: 'right' as const, valign: 'middle' as const }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: aligned }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const right = screen.getByRole('button', { name: 'Align right' })
    expect(right).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Vertical align middle' })).toHaveAttribute('aria-pressed', 'true')
    fireEvent.click(right)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(sent.map((payload) => new TextDecoder().decode(payload)).join('')).toContain('"op":"clear"')
  })
})

// Story 7.4: authoring body text in the designer. The CONTENT control was an
// <input type="text">, which cannot hold a line feed at all, so a
// multi-paragraph clause could not be typed OR pasted — the story's first AC
// in one sentence.
describe('Story 7.4: authoring a multi-paragraph clause', () => {
  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  const openEditor = (sent: ArrayBuffer[]) => {
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    return screen.getByRole('textbox', { name: 'Text' })
  }

  it('is a textarea, so Enter inserts a paragraph break instead of committing', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    expect(field.tagName).toBe('TEXTAREA')
    // Three paragraphs, typed. Enter commits nothing; blur sends ONE command
    // carrying the whole value with its line feeds intact.
    fireEvent.change(field, { target: { value: 'First clause.\nSecond clause.\nThird clause.' } })
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(sent).toHaveLength(0)
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"First clause.\\nSecond clause.\\nThird clause."}}}')
  })

  it('commits a forty-page clause as one command, with every paragraph break intact', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    // Well past the 512 BYTES the projection used to cap an element's value
    // at — about eighty English words, less than one numbered clause — which
    // did not merely blank the canvas but REJECTED the edit, because the
    // property command re-projects inside its own transaction.
    const clause = Array.from({ length: 1_900 }, (_value, index) => `Clause ${index + 1}. The parties agree as set out above.`).join('\n')
    expect(new TextEncoder().encode(clause).byteLength).toBeGreaterThan(512)
    fireEvent.change(field, { target: { value: clause } })
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    const command = new TextDecoder().decode(sent[0]!)
    expect(JSON.parse(command).changes.value.value).toBe(clause)
    expect(sent).toHaveLength(1)
  })

  it('takes only the plain flavour of a word-processor paste, keeping paragraph breaks and dropping the formatting', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: '' } })
    const flavours: Record<string, string> = {
      'text/plain': 'Clause 1.\nClause 2.',
      'text/html': '<p style="font-weight:700;font-family:Georgia">Clause 1.</p><p><em>Clause 2.</em></p>',
      'text/rtf': '{\\rtf1 \\b Clause 1.\\par}',
    }
    const read: string[] = []
    fireEvent.paste(field, { clipboardData: { getData: (flavour: string) => { read.push(flavour); return flavours[flavour] ?? '' } } })
    // The formatting is discarded by never being looked at: no sanitiser, no
    // new dependency, nothing to go wrong on an unusual clipboard.
    expect(read).toEqual(['text/plain'])
    expect(field).toHaveValue('Clause 1.\nClause 2.')
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(JSON.parse(new TextDecoder().decode(sent[0]!)).changes.value.value).toBe('Clause 1.\nClause 2.')
  })

  it('inserts nothing when the clipboard has no plain flavour, instead of letting the browser paste the HTML', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: 'Clause 1.' } })
    const read: string[] = []
    const flavours: Record<string, string> = {
      'text/html': '<p style="font-weight:700;font-family:Georgia">Pasted from a word processor</p>',
      'text/rtf': '{\\rtf1 \\b Pasted from a word processor\\par}',
    }
    // fireEvent returns false when the event was cancelled, which is the only
    // way to see from here that the BROWSER's own paste was refused. Without
    // preventDefault the browser inserts text it derives from the text/html
    // flavour — formatting laundered in through the door the story closed.
    const dispatched = fireEvent.paste(field, { clipboardData: { getData: (flavour: string) => { read.push(flavour); return flavours[flavour] ?? '' } } })
    expect(dispatched).toBe(false)
    expect(read).toEqual(['text/plain'])
    expect(field).toHaveValue('Clause 1.')
    expect(sent).toHaveLength(0)
  })

  it('leaves the caret at the end of the pasted text, not at the end of the field', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent) as HTMLTextAreaElement
    fireEvent.change(field, { target: { value: 'Clause 1.\nClause 3.' } })
    // The caret sits at the start of the LAST paragraph, and a clause is
    // pasted in front of it. A controlled textarea re-rendered with a new
    // value drops the caret at the very end unless it is restored, which would
    // put the author's next keystroke in the wrong paragraph.
    field.setSelectionRange(10, 10)
    fireEvent.paste(field, { clipboardData: { getData: () => 'Clause 2.\n' } })
    expect(field).toHaveValue('Clause 1.\nClause 2.\nClause 3.')
    expect(field.selectionStart).toBe(20)
    expect(field.selectionEnd).toBe(20)
  })

  it('reverts and blurs on Escape, exactly as every single-line field does', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: 'a draft\nnobody wants' } })
    fireEvent.keyDown(field, { key: 'Escape' })
    expect(field).toHaveValue('Hello')
    expect(sent).toHaveLength(0)
  })

  it('paints one canvas line per engine line and says, in words, when the paint is only a prefix', () => {
    const line = (top: number, text: string) => ({ top, baseline: top + 10_000, advance: 14_000, width: 30_000, fragments: [{ text, x: 0 }] })
    const cut = { ...canvas, components: [{ ...textComponent, textPaint: { overflow: false, truncated: true, lines: [line(0, 'First clause.'), line(14_000, 'Second clause.')] } }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: cut }} />)
    expect(document.querySelectorAll('.canvas-text-line')).toHaveLength(2)
    // Stated in words at the component, and in the same sentence a screen
    // reader gets — never by colour, and never only by a CSS class.
    const notice = 'Canvas preview cut short. The whole text is in the document and prints in full.'
    expect(screen.getByText(notice)).toBeInTheDocument()
    expect(screen.getByLabelText(`text component e1: First clause. Second clause.; ${notice}`)).toBeInTheDocument()
  })

  it('says nothing about truncation for an untruncated paint or an empty one', () => {
    const notice = 'Canvas preview cut short. The whole text is in the document and prints in full.'
    const whole = { ...canvas, components: [
      { ...textComponent, textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 10_000, advance: 14_000, width: 30_000, fragments: [{ text: 'First clause.', x: 0 }] }] } },
      { id: 'e2', type: 'text' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 24_000, resizable: true, textPaint: { overflow: false, truncated: false, lines: [] } },
    ] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: whole }} />)
    expect(screen.queryByText(notice)).not.toBeInTheDocument()
  })
})

describe('Story 5.13: image asset selection', () => {
  const imageComponent = { id: 'e1', type: 'image' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 48_000, resizable: true, image: { mediaType: 'image/png', assetKey: 'a'.repeat(64), width: 300, height: 200, drawX: 6_000, drawY: 8_000, drawWidth: 60_000, drawHeight: 40_000 } }
  const undecodableImageComponent = { id: 'e2', type: 'image' as const, band: 'content' as const, x: 0, y: 60_000, width: 72_000, height: 48_000, resizable: true, imageUnavailable: 'undecodable' as const }
  const textComponent = { id: 'e3', type: 'text' as const, band: 'content' as const, x: 0, y: 120_000, width: 72_000, height: 24_000, resizable: true }
  const imageFileAccess = (openImage: () => Promise<{ bytes: ArrayBuffer; mediaType: string; name: string }>) => ({ openImage }) as unknown as import('./image-file').ImageFileAccess

  it('shows the IMAGE section carrying the engine snapshot identity for a single image selection, and never for other selections', () => {
    const componentCanvas = { ...canvas, components: [imageComponent, textComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    // Empty selection: no IMAGE section.
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
    // Single image selection: IMAGE section with identity from the snapshot.
    fireEvent.click(screen.getByLabelText('image component e1'))
    expect(screen.getByText('IMAGE')).toBeInTheDocument()
    expect(screen.getByText('image/png · 300×200px · asset aaaaaaaaaaaa…')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Choose image…' })).toBeInTheDocument()
    // Mixed selection (image + text): no IMAGE section.
    fireEvent.click(screen.getByLabelText('text component e3'), { shiftKey: true })
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
    // Non-image single selection: no IMAGE section.
    fireEvent.click(screen.getByLabelText('text component e3'))
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
  })

  it('states the concrete reason for an asset this version cannot render, distinguished by text, not colour alone', () => {
    const componentCanvas = { ...canvas, components: [undecodableImageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    fireEvent.click(screen.getByLabelText('image component e2'))
    expect(screen.getByText("This version cannot render this asset's media type.")).toBeInTheDocument()
  })

  it('shows an empty, choosable box for a placed image with no file yet', () => {
    // Go projects neither a paint nor an unavailable reason for a box the
    // author has not filled: nothing to draw, and nothing wrong.
    const emptyImageComponent = { id: 'e2', type: 'image' as const, band: 'content' as const, x: 0, y: 60_000, width: 72_000, height: 48_000, resizable: true }
    const componentCanvas = { ...canvas, components: [emptyImageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    expect(screen.getByText('No image')).toBeInTheDocument()
    expect(screen.queryByText(/Image unavailable/)).not.toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('image component e2'))
    expect(screen.getByText(/No image chosen yet/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Choose image…' })).toBeEnabled()
  })

  it('states the concrete reason when no local picker capability is available in this browser tier', () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    const button = screen.getByRole('button', { name: 'Choose image…' })
    expect(button).toBeDisabled()
    expect(screen.getByText('No local file picker is available in this browser tier.')).toBeInTheDocument()
  })

  it('drives the keyboard path from a real keyboard SELECTION to a focus-visible picker control, then commits through it', async () => {
    // Finding 15 (review of 2026-08-29): the original version of this test
    // selected the component with fireEvent.click (a mouse-shaped
    // interaction) and only PROVED the picker button was focusable, never
    // dispatching a key event at all. CanvasComponent has its own
    // onKeyDown handler for Enter/Space selection (App.tsx) — exercise
    // THAT, not a click, for the selection half of "selection to picker".
    const componentCanvas = { ...canvas, components: [imageComponent] }
    const openImage = vi.fn(async () => ({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' }))
    const request = vi.fn(async (operation: string, _payload?: ArrayBuffer) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(openImage)} />)
    const component = screen.getByLabelText('image component e1')
    component.focus()
    fireEvent.keyDown(component, { key: 'Enter' })
    expect(screen.getByText('IMAGE')).toBeInTheDocument()
    const button = screen.getByRole('button', { name: 'Choose image…' })
    // "Keyboard-reachable... with visible colors.select focus" (AC2):
    // proved by moving focus WITHOUT a pointer event and checking it
    // landed. A native <button> converts Enter/Space into a real 'click'
    // event by construction in every browser — jsdom does not synthesise
    // that translation from a bare keyDown, so the committed activation
    // below stands in for it here; the Playwright suite drives the SAME
    // control with a real OS-level key press (image-asset.spec.ts).
    button.focus()
    expect(document.activeElement).toBe(button)
    fireEvent.click(button)
    await waitFor(() => expect(openImage).toHaveBeenCalledOnce())
    await waitFor(() => expect(request.mock.calls.some(([operation]) => operation === 'command')).toBe(true))
    const [, payload] = request.mock.calls.find(([operation]) => operation === 'command')!
    const command = new TextDecoder().decode(payload)
    expect(command).toContain('"kind":"setComponentAsset"')
    expect(command).toContain('"id":"e1"')
    expect(command).toContain('"mediaType":"image/jpeg"')
  })

  it('shows a located diagnostic when the command rejects the picked file, and shows nothing when the picker is cancelled', async () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    const request = vi.fn(async (operation: string) => operation === 'command' ? Promise.reject(Object.assign(new Error('asset exceeds the 8388608-byte supported size'), { elementId: 'e1', dataPath: 'component.data' })) : { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: componentCanvas } })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'huge.png' }))} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('e1: asset exceeds the 8388608-byte supported size'))
  })

  it('shows no error when the local picker is cancelled', async () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => { throw new FileAccessCancelled() })} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Choose image…' })).not.toBeDisabled())
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('does not install a setComponentAsset result that resolves after a document replacement (Finding 4)', async () => {
    // AC1's own named red proof: "a command result installed after
    // document replacement". Element ids are reused across documents
    // (e1, e2, ...), and this closure spans the two longest awaits in the
    // app — an OS file dialog, then an engine command carrying up to
    // megabytes — so if Open/Start blank/undo lands in between, a stale
    // command result must never overwrite the newer, authoritative
    // document. Before the fix, applyImageAsset called setCurrentSnapshot
    // unconditionally with no generation/revision guard, matching every
    // OTHER committed-command path's (bindPickedPath, etc.) shape only in
    // that one respect being ABSENT.
    const componentCanvas = { ...canvas, components: [imageComponent] }
    let resolveOpenImage: ((value: { bytes: ArrayBuffer; mediaType: string; name: string }) => void) | undefined
    const openImage = vi.fn(() => new Promise<{ bytes: ArrayBuffer; mediaType: string; name: string }>((resolve) => { resolveOpenImage = resolve }))
    const request = vi.fn(async (operation: string) => {
      if (operation === 'load') return { snapshot: { documentState: 'loaded' as const, revision: 50, byteLength: 3 } }
      if (operation === 'command') return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }
      return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: componentCanvas } }
    })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(openImage)} blankBytes={bytes} />)

    // Start the asset pick — this awaits openImage(), which we hold open.
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(openImage).toHaveBeenCalledOnce())

    // A DOCUMENT REPLACEMENT lands while the picker is still open.
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 50'))

    // NOW the picker resolves and the stale command completes.
    resolveOpenImage!({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' })
    await waitFor(() => expect(request.mock.calls.some(([operation]) => operation === 'command')).toBe(true))

    // The blank document (revision 50) must still be showing — the stale
    // command's revision-2 result must never have been installed.
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 50')
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('paints the image inside the Go-owned draw rectangle at zoom, fetched per asset key, shows an honest placeholder for an undecodable asset, and revokes its object URL on every trigger AC3 names', async () => {
    let paintCounter = 0
    const createObjectURL = vi.fn(() => `blob:paint-${++paintCounter}`)
    const revokeObjectURL = vi.fn()
    const priorCreate = URL.createObjectURL; const priorRevoke = URL.revokeObjectURL
    ;(URL as unknown as { createObjectURL: typeof createObjectURL }).createObjectURL = createObjectURL
    ;(URL as unknown as { revokeObjectURL: typeof revokeObjectURL }).revokeObjectURL = revokeObjectURL
    try {
      const canvasWithKey = (assetKey: string, extra: ReadonlyArray<typeof undecodableImageComponent> = [undecodableImageComponent]) =>
        ({ ...canvas, components: [{ ...imageComponent, image: { ...imageComponent.image, assetKey } }, ...extra] })
      const canvasA = canvasWithKey('a'.repeat(64))
      // Finding 11: asset REPLACEMENT (same element id, new assetKey) is a
      // distinct trigger from document replacement — exercised via a
      // committed setComponentAsset, never a document generation bump.
      const canvasB = canvasWithKey('b'.repeat(64))
      const openImage = vi.fn(async () => ({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' }))
      const request = vi.fn(async (operation: string) => {
        if (operation === 'asset') return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3 }, bytes }
        if (operation === 'command') return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: canvasB } }
        // Finding 11: document REPLACEMENT (Start blank) with the SAME
        // assetKey as canvasB — isolates the `generation` dependency from
        // `assetKey`, so a future edit that dropped `generation` from
        // ImagePaint's effect deps would leak here with nothing else red.
        if (operation === 'load') return { snapshot: { documentState: 'loaded' as const, revision: 9, byteLength: 3, canvas: canvasB } }
        return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: canvasA } }
      })
      const view = render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: canvasA }} imageFileAccess={imageFileAccess(openImage)} blankBytes={bytes} />)
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledOnce())
      const img = () => view.container.querySelector('img.canvas-image-paint') as HTMLImageElement
      expect(img().src).toContain('blob:paint-1')
      // The undecodable second element paints an honest, named placeholder,
      // never a blank box and never a crash.
      expect(screen.getByText(/Image unavailable/)).toBeInTheDocument()

      // Finding 3: the painted element's geometry must equal the engine's
      // OWN draw rectangle (image.drawX/Y/W/H relative to component.x/y),
      // mapped through canvasDisplay's zoom rule — never object-fit, never
      // a browser-computed fit. Checked at the default zoom (1) first.
      expect(img().style.left).toBe('6px'); expect(img().style.top).toBe('8px')
      expect(img().style.width).toBe('60px'); expect(img().style.height).toBe('40px')

      // Finding 3/10: AC3 says the rectangle is "mapped through the
      // existing zoom rule" — this was asserted at NO zoom before. One
      // step of "Zoom in" (+0.1) must scale every one of the four values
      // by exactly the same factor the engine's own zoom rule uses.
      fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
      await waitFor(() => expect(screen.getByLabelText('Canvas zoom')).toHaveTextContent('110%'))
      expect(img().style.left).toBe('6.6px'); expect(img().style.top).toBe('8.8px')
      expect(img().style.width).toBe('66px'); expect(img().style.height).toBe('44px')

      // Finding 11, trigger 1 of 3: ASSET REPLACEMENT. Pick a new file for
      // the same element; the committed command repoints its assetKey, and
      // that alone (not a document generation bump) must revoke the first
      // URL and fetch a second.
      fireEvent.click(screen.getByLabelText('image component e1'))
      fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
      await waitFor(() => expect(openImage).toHaveBeenCalledOnce())
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledTimes(2))
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-1')
      expect(img().src).toContain('blob:paint-2')

      // Finding 11, trigger 2 of 3: DOCUMENT REPLACEMENT. Start blank loads
      // a canvas whose e1 element carries the SAME assetKey as canvasB —
      // only `generation` changed, isolating it from the assetKey trigger
      // just exercised above.
      fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledTimes(3))
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-2')
      expect(img().src).toContain('blob:paint-3')

      // Finding 11, trigger 3 of 3: DELETION (unmount) revokes the URL this
      // effect most recently created — no accumulation across a session.
      view.unmount()
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-3')
    } finally {
      (URL as unknown as { createObjectURL: typeof priorCreate }).createObjectURL = priorCreate
      ;(URL as unknown as { revokeObjectURL: typeof priorRevoke }).revokeObjectURL = priorRevoke
    }
  })
})

// STORY 7.6 — THE CANVAS DRAWS EVERY PAGE THE DOCUMENT WILL PRODUCE.
//
// Every projection below is a fixture: the sheets, the seams and the
// disclosures are read from `contentWindowOrigins` and
// `contentWindowCountIsExact`, so a test that changed the origins and saw the
// same drawing would be a test of nothing.
describe('canvas sheet stack', () => {
  const origins = [0, 700_000, 1_400_000]
  const threeWindows = { ...canvas, contentWindowCount: 3, contentWindowOrigins: origins }
  const header = { id: 'h1', type: 'text' as const, band: 'pageHeader' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: true }
  const footer = { id: 'f1', type: 'text' as const, band: 'pageFooter' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: true }
  const at = (id: string, y: number, height = 24_000) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height, resizable: true })
  const snapshotOf = (projection: CanvasProjection) => ({ documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: projection })
  const sheetLabels = () => Array.from(document.querySelectorAll('.page-surface')).map((surface) => surface.getAttribute('aria-label'))

  it('draws one sheet per projected window, in order, repeating the two bands the engine repeats', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [header, at('e1', 0), footer] })} />)
    expect(sheetLabels()).toEqual([
      'Report page 1 of 3 with Page Header, Content, and Page Footer',
      'Report page 2 of 3 with Page Header, Content, and Page Footer',
      'Report page 3 of 3 with Page Header, Content, and Page Footer',
    ])
    // Each sheet carries all three bands, and the repeating bands carry their
    // components — because the engine repeats them onto every printed page.
    expect(document.querySelectorAll('.page-band-pageHeader')).toHaveLength(3)
    expect(document.querySelectorAll('.page-band-pageFooter')).toHaveLength(3)
    expect(document.querySelectorAll('.page-band-content')).toHaveLength(3)
    // One accessible name each, though: a repeated component is one
    // component, and two identical names would make selection ambiguous.
    expect(screen.getByLabelText('text component h1')).toBeInTheDocument()
    expect(screen.getByLabelText('text component f1')).toBeInTheDocument()
    expect(document.querySelectorAll('.canvas-component-echo')).toHaveLength(4)
  })

  it('marks the seam at the projected origin of the NEXT window, and nowhere when that is past the foot', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf(threeWindows)} />)
    const seams = Array.from(document.querySelectorAll('.page-seam')).map((seam) => (seam as HTMLElement).style.getPropertyValue('--seam-display-y'))
    // origins[1] − origins[0] and origins[2] − origins[1], at zoom 1. The
    // last sheet has no next window, so it draws no marker.
    expect(seams).toEqual(['700px', '700px'])
    // RED PROOF, run and recorded: substituting the window height multiplied
    // by an index for the projected origin gives 729.89px here — the wrong
    // place by a tenth of an inch, and by nine whole sheets on a column with
    // a declared gap. The origins move the marker; nothing else does.
    expect(seams).not.toContain(`${canvas.contentWindowHeight / 1000}px`)
  })

  it('draws no in-sheet seam where the next window begins past the sheet own foot', () => {
    // A declared gap: the next window begins far below this sheet, so the
    // band's own foot IS the boundary and the skipped column region is drawn
    // by nobody.
    const declaredGap = { ...canvas, contentWindowCount: 2, contentWindowOrigins: [0, 7_280_000] }
    render(<App engine={engine()} initialSnapshot={snapshotOf(declaredGap)} />)
    expect(document.querySelectorAll('.page-surface')).toHaveLength(2)
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
  })

  it('draws a component that crosses a seam on both windows, with one interactive home', () => {
    const spanning = at('e2', 650_000, 100_000)
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [spanning] })} />)
    // getByLabelText throws on more than one match, so this assertion IS the
    // uniqueness claim.
    const home = screen.getByLabelText('text component e2')
    expect(home.style.getPropertyValue('--component-y')).toBe('650px')
    const echoes = Array.from(document.querySelectorAll('.canvas-component-echo')) as HTMLElement[]
    expect(echoes).toHaveLength(1)
    // The echo is positioned at the component's column offset MINUS the
    // window it is drawn on, so the run reads as continuous across the seam.
    expect(echoes[0]?.style.getPropertyValue('--component-y')).toBe('-50px')
    expect(echoes[0]?.getAttribute('aria-hidden')).toBe('true')
    expect(echoes[0]?.getAttribute('role')).toBeNull()
  })

  it('sends a later-sheet placement as the band-aware createComponent command, carrying a COLUMN coordinate', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content on page 3 of 3'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // 1400pt is contentWindowOrigins[2] in points — a position in the column,
    // never a pin to sheet three. Go's hitTestBand rectangle is one page tall
    // and is not moved: this routes around it rather than through it.
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"createComponent","version":1,"type":"text","band":"content","x":0,"y":1400,"width":72,"height":24,"snap":true}')
  })

  it('sends a later-sheet POINTER placement through the same column translation as the keyboard one', async () => {
    // The pointer branch and the keyboard branch are two different expressions
    // on the same handler, and only the keyboard one was exercised: replacing
    // the pointer branch with `placeInBand(band.name, point.x, point.y)` — no
    // column origin, page-absolute x — left the whole designer suite green
    // while a mouse-dropped component silently landed on sheet one.
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    const band = screen.getByLabelText('Content on page 3 of 3')
    const localX = ['offset', 'X'].join('')
    const localY = ['offset', 'Y'].join('')
    // jsdom exposes these two as prototype getters that always answer 0, so a
    // plain fireEvent property bag is silently dropped; they have to be
    // defined on the native event the handler actually reads.
    const released = createEvent.pointerUp(band)
    Object.defineProperty(released, localX, { value: 120 })
    Object.defineProperty(released, localY, { value: 40 })
    fireEvent(band, released)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // x is band-relative (156 page-absolute less the band's own 36), and y is
    // the COLUMN offset: contentWindowOrigins[2] of 1400pt plus the 40pt the
    // pointer sat below the band's head.
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"createComponent","version":1,"type":"text","band":"content","x":120,"y":1440,"width":72,"height":24,"snap":true}')
  })

  it('keeps the FIRST sheet on today dropComponent payload even when the stack is deep', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content on page 1 of 3'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })

  it('drags a component across a seam onto a later sheet and commits a column coordinate', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0)] })} />)
    const component = screen.getByLabelText('text component e1')
    // One whole sheet-plus-gap down the stack: 841.890pt of page and the 24px
    // gap the stack declares.
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 10, clientY: 875.89 })
    // Tracking the hand means landing one WINDOW lower in the column — 700pt,
    // the projected origin of window two — not 865.89pt, which is what the
    // linear pixel delta this replaced would have proposed. The difference is
    // exactly the page footer, the gap and the page header the pointer
    // crossed.
    expect(component.style.getPropertyValue('--component-y')).toBe('700px')
    // The clip has to lift for the duration of the gesture, or the component
    // is clipped out of view at the very seam it is being dragged across.
    // jsdom applies no stylesheet, so the class is the only observable the
    // suite has: pinning it to the constant `band-window` left every numeric
    // assertion here green while the gesture was visually broken.
    expect(document.querySelectorAll('.band-window-open').length).toBeGreaterThan(0)
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 10, clientY: 875.89 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // And it drops again once the gesture has settled, so the clip is back for
    // every component that is not being dragged.
    await waitFor(() => expect(document.querySelectorAll('.band-window-open')).toHaveLength(0))
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"moveComponent","version":1,"id":"e1","x":0,"y":700,"snap":true}')
  })

  it('lets a drag leave the content band vertically while still capping the repeating bands', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0), footer] })} />)
    const content = screen.getByLabelText('text component e1')
    fireEvent.pointerDown(content, { pointerId: 1, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(content, { pointerId: 1, clientX: 0, clientY: 800 })
    // 800pt down a 729.89pt window. RED PROOF, run and recorded: restoring
    // the clamp DW-36 lifted pins this at 705.89px — the band's own height
    // less the component's — which is the whole of why the lifted column was
    // reachable by command and not by hand.
    expect(content.style.getPropertyValue('--component-y')).toBe('800px')
    const repeated = screen.getByLabelText('text component f1')
    fireEvent.pointerDown(repeated, { pointerId: 2, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(repeated, { pointerId: 2, clientX: 0, clientY: 400 })
    // The page footer is 20pt tall and the component is 12pt tall, so the
    // drag rests at the band foot exactly as it did before this story.
    expect(repeated.style.getPropertyValue('--component-y')).toBe('8px')
  })

  it('states what the sheets claim, in accessible text, whenever there is more than one or the count is not exact', () => {
    const claim = "A component's page is a consequence of the content above it and can change when the data does — it is a column position, not a pin to page three."
    const inexact = 'This count depends on data the canvas does not have, so the printed document can run to a different number of pages.'
    render(<App engine={engine()} initialSnapshot={snapshotOf(threeWindows)} />)
    const disclosure = screen.getByRole('status', { name: 'Canvas sheet disclosure' })
    expect(disclosure).toHaveTextContent('Showing 3 sheets.')
    expect(disclosure).toHaveTextContent('These are the pages this content column occupies as the canvas has laid it out, not a prediction of the printed document.')
    expect(disclosure).toHaveTextContent(claim)
    // Exact, so it must NOT disclaim: a disclosure that always said
    // everything would say nothing.
    expect(disclosure).not.toHaveTextContent(inexact)
  })

  it('says a data-length document can print a different number of pages than are drawn, even when it draws only one sheet', () => {
    // DIRECTION-FREE, deliberately. The old sentence here promised MORE
    // pages, which is true of this bound table and false of an element the
    // data hides — and one document can carry both.
    const inexact = 'This count depends on data the canvas does not have, so the printed document can run to a different number of pages.'
    // The shipped statement shape: four byte-identical templates that print
    // one, five, twenty and fifty pages and project ONE window each, because
    // the canvas has never been given the data.
    const bound = { ...canvas, contentWindowCountIsExact: false, components: [{ id: 'e8', type: 'table' as const, band: 'content' as const, x: 0, y: 54_000, width: 400_000, height: 28_000, resizable: false, tableBind: 'transactions[]' }] }
    render(<App engine={engine()} initialSnapshot={snapshotOf(bound)} />)
    const disclosure = screen.getByRole('status', { name: 'Canvas sheet disclosure' })
    expect(disclosure).toHaveTextContent('Showing 1 sheet.')
    expect(disclosure).toHaveTextContent(inexact)
    // One window, so no seam and no page qualifier — the honesty is in the
    // words, not in a drawing that would be a second lie.
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
    expect(sheetLabels()).toEqual(['Report page with Page Header, Content, and Page Footer'])
  })

  it('carries the same claim into the accessible name of a component on a later sheet, and not onto a first-sheet one', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0), at('e3', 1_450_000)] })} />)
    // The exact sentence, so deleting it turns this red rather than merely
    // shortening a string nobody asserted.
    expect(screen.getByLabelText('text component e3; on canvas page 3 of 3, which is a consequence of the content above it and can change when the data does — a column position, not a pin to page 3')).toBeInTheDocument()
    // The first sheet's component keeps the name it had before this story.
    expect(screen.getByLabelText('text component e1')).toBeInTheDocument()
  })

  it('draws the first budgeted sheets, says it is showing the first N of M, and never blanks', () => {
    const many = MAX_CANVAS_SHEETS + 5
    const budgeted = { ...canvas, contentWindowCount: many, contentWindowOrigins: Array.from({ length: many }, (_value, index) => index * 700_000) }
    render(<App engine={engine()} initialSnapshot={snapshotOf(budgeted)} />)
    expect(document.querySelectorAll('.page-surface')).toHaveLength(MAX_CANVAS_SHEETS)
    expect(screen.getByRole('status', { name: 'Canvas sheet disclosure' })).toHaveTextContent(`Showing the first ${MAX_CANVAS_SHEETS} sheets of ${many}.`)
    expect(screen.queryByText('Waiting for Go page geometry.')).not.toBeInTheDocument()
  })

  // AC5, ASSERTED RATHER THAN ASSUMED. A template whose content column
  // occupies one window and whose content band holds nothing whose length
  // comes from data must render what it rendered at 2c5cfa1: the same DOM,
  // the same accessible names, the same command payload. The compile-only e2e
  // specs address these labels in Playwright STRICT MODE, where a duplicate
  // or a renamed label would be caught by nothing executable.
  it('renders a genuinely single-page template exactly as it did before this story', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf({ ...canvas, components: [at('e1', 0)] })} />)
    expect(sheetLabels()).toEqual(['Report page with Page Header, Content, and Page Footer'])
    expect(screen.getByLabelText('Page Header')).toBeInTheDocument()
    expect(screen.getByLabelText('Content')).toBeInTheDocument()
    expect(screen.getByLabelText('Page Footer')).toBeInTheDocument()
    expect(screen.getByLabelText('text component e1')).toBeInTheDocument()
    // No seam, no stack wrapper, no clipping window, no echo, no disclosure.
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
    expect(document.querySelectorAll('.sheet-stack')).toHaveLength(0)
    expect(document.querySelectorAll('.band-window')).toHaveLength(0)
    expect(document.querySelectorAll('.canvas-component-echo')).toHaveLength(0)
    expect(screen.queryByRole('status', { name: 'Canvas sheet disclosure' })).not.toBeInTheDocument()
    // And the payload, byte for byte.
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })
})

// STORY 8.2. The chain editor is the first designer surface that EDITS the
// document's own `fonts` map, and the whole of its correctness is that it
// holds nothing: no list, no ordering, no rule, no sentence. Each test below
// is written to fail if the panel starts holding one of them.
describe('the font chain editor, where fonts are chosen', () => {
  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  type Chains = typeof canvas.fontChains
  const chainSnapshot = (revision: number, chains: Chains) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas: { ...canvas, fontFamilies: chains.map((chain) => chain.name), fontChains: chains, components: [textComponent] } })
  const declared: Chains = canvas.fontChains
  const commands = (request: { mock: { calls: unknown[][] } }) => request.mock.calls.filter((call) => call[0] === 'command').map((call) => new TextDecoder().decode(call[1] as ArrayBuffer))
  const open = (request: ReturnType<typeof vi.fn>, chains: Chains = declared) => {
    render(<App engine={engine(request as never)} initialSnapshot={chainSnapshot(1, chains)} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    return request
  }
  const accepts = (chains: Chains = declared) => vi.fn(async (operation: string) => operation === 'command' ? chainSnapshotResponse(chains) : { snapshot: chainSnapshot(1, declared) })
  const chainSnapshotResponse = (chains: Chains) => ({ snapshot: chainSnapshot(2, chains) })
  const refuses = (message: string, provenance: object = {}) => vi.fn(async (operation: string) => {
    if (operation === 'command') throw Object.assign(new Error(message), provenance)
    return { snapshot: chainSnapshot(1, declared) }
  })

  it('reveals the editor inside the typography section, never as a dialog, showing the projection\'s own chains', () => {
    open(accepts())
    // AC1: the same panel section. No dialog is opened, and the editor's own
    // element is INSIDE the TYPOGRAPHY section it was revealed from.
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    const editor = screen.getByRole('group', { name: 'Font chains' })
    expect(editor.closest('.property-section-typography')).not.toBeNull()
    // The list is the projection's, name for name and entry for entry.
    expect(screen.getAllByRole('textbox', { name: /^Font chain \d+ name$/ }).map((field) => (field as HTMLInputElement).value)).toEqual(['body', 'heading'])
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Noto Sans↑↓×'])
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 2' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Noto Sans↑↓×', 'Noto Sans Thai↑↓×'])
    // The combobox this affordance sits beside keeps its accessible name,
    // which is the one e2e witness addresses it by.
    expect(screen.getByRole('combobox', { name: 'Font family' })).toBeInTheDocument()
  })

  // Story 8.3. The panel holds NO rule about which kind an entry is: the
  // engine projects the discriminant (`assetKey`) and the display strings
  // (`family`, `style`), and the panel places them. A named face still shows
  // its name and nothing else.
  const embedded = (assetKey: string, family: string, style: string) => ({ face: '', assetKey, family, style })

  it('draws an embedded entry as its projected family and style, and a named face as its name', () => {
    const key = 'c'.repeat(64)
    open(accepts(), [{ name: 'body', entries: [face('Noto Sans'), embedded(key, 'Inter', 'Regular')] }])
    const items = within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)
    expect(items).toEqual(['Noto Sans↑↓×', 'Inter Regular↑↓×'])
    // The asset key is never shown when the document named the face: the
    // author reads a typeface, not a digest.
    expect(screen.queryByText(key)).not.toBeInTheDocument()
  })

  it('shows the family alone when the embedded face declares no style', () => {
    // Go projects the asset key AS the family when the document declares no
    // `font.family`, so the panel always has a name and never has to decide
    // what to draw for an empty one — the browser derives nothing either way.
    const key = 'd'.repeat(64)
    open(accepts(), [{ name: 'body', entries: [embedded(key, 'Inter', ''), embedded(key, key, '')] }])
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Inter↑↓×', `${key}↑↓×`])
  })

  // THE CHANGE SIGNATURE IS VALUE-BASED, and this is the snapshot that proves
  // it. The two entries differ ONLY in a field `entries.join(' ')` would have
  // flattened — both are embedded, so both stringify to `[object Object]` —
  // so under the old signature the accepted edit would have read as "the list
  // did not move" and focus would have stayed on the pressed control instead
  // of following the entry.
  it('reads a moved entry as a move even when only a projected field differs', async () => {
    const one = 'a'.repeat(64)
    const two = 'b'.repeat(64)
    const three = 'c'.repeat(64)
    // THREE entries, so the moved one lands in the MIDDLE and "move later" is
    // still enabled there — the end-of-list fallback would otherwise mask
    // which control focus actually went to.
    const before: Chains = [{ name: 'body', entries: [embedded(one, 'Inter', 'Regular'), embedded(two, 'Lora', 'Italic'), embedded(three, 'Cardo', 'Bold')] }]
    const after: Chains = [{ name: 'body', entries: [embedded(two, 'Lora', 'Italic'), embedded(one, 'Inter', 'Regular'), embedded(three, 'Cardo', 'Bold')] }]
    const request = vi.fn(async (operation: string) => operation === 'command' ? { snapshot: chainSnapshot(2, after) } : { snapshot: chainSnapshot(1, before) })
    open(request, before)
    const later = screen.getByRole('button', { name: 'Move entry 1 of font chain 1 later' })
    later.focus()
    fireEvent.click(later)
    await waitFor(() => expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Lora Italic↑↓×', 'Inter Regular↑↓×', 'Cardo Bold↑↓×']))
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Move entry 2 of font chain 1 later' }))
  })

  it('states an empty document rather than drawing a chain that is not there', () => {
    open(accepts([]), [])
    expect(screen.getByText('This document declares no font chains.')).toBeInTheDocument()
    expect(screen.queryByRole('textbox', { name: /^Font chain \d+ name$/ })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add font chain' })).toBeInTheDocument()
  })

  it('dispatches each of the six commands, byte for byte, from its own control', async () => {
    const request = open(accepts())
    const step = async (act: () => void, expected: string, count: number) => {
      act()
      await waitFor(() => expect(commands(request)).toHaveLength(count))
      expect(commands(request)[count - 1]).toBe(expected)
    }
    await step(() => {
      fireEvent.change(screen.getByRole('textbox', { name: 'Font chain 1 name' }), { target: { value: 'text' } })
      fireEvent.click(screen.getByRole('button', { name: 'Rename font chain 1' }))
    }, '{"kind":"renameFontChain","version":1,"name":"body","to":"text"}', 1)
    await step(() => fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' })), '{"kind":"deleteFontChain","version":1,"name":"body"}', 2)
    await step(() => {
      fireEvent.change(screen.getByRole('textbox', { name: 'New entry for font chain 2' }), { target: { value: 'Noto Sans SC' } })
      fireEvent.click(screen.getByRole('button', { name: 'Add entry to font chain 2' }))
    }, '{"kind":"addFontChainEntry","version":1,"name":"heading","index":2,"face":"Noto Sans SC"}', 3)
    await step(() => fireEvent.click(screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' })), '{"kind":"moveFontChainEntry","version":1,"name":"heading","from":0,"to":1}', 4)
    await step(() => fireEvent.click(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 earlier' })), '{"kind":"moveFontChainEntry","version":1,"name":"heading","from":1,"to":0}', 5)
    await step(() => fireEvent.click(screen.getByRole('button', { name: 'Remove entry 1 of font chain 2' })), '{"kind":"removeFontChainEntry","version":1,"name":"heading","index":0}', 6)
    await step(() => {
      fireEvent.change(screen.getByRole('textbox', { name: 'New font chain name' }), { target: { value: 'a"b\\c' } })
      fireEvent.change(screen.getByRole('textbox', { name: 'First entry for the new font chain' }), { target: { value: 'Noto Sans' } })
      fireEvent.click(screen.getByRole('button', { name: 'Add font chain' }))
    }, '{"kind":"addFontChain","version":1,"name":"a\\"b\\\\c","entries":["Noto Sans"]}', 7)
  })

  // THE DISCRIMINATING PROOF (AD-15). A test that types a value and reads the
  // same value back cannot tell a re-projection from a local model: both pass.
  // So the engine answers with a DIFFERENT name than the one typed, and the
  // panel must show the engine's.
  it('renders the name the engine returned, not the name the author typed', async () => {
    const request = vi.fn(async (operation: string) => operation === 'command'
      ? { snapshot: chainSnapshot(2, [{ name: 'engine-chose', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans Thai')] }]) }
      : { snapshot: chainSnapshot(1, declared) })
    open(request)
    fireEvent.change(screen.getByRole('textbox', { name: 'Font chain 1 name' }), { target: { value: 'typed-by-the-author' } })
    fireEvent.click(screen.getByRole('button', { name: 'Rename font chain 1' }))
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'Font chain 1 name' })).toHaveValue('engine-chose'))
    expect(screen.queryByDisplayValue('typed-by-the-author')).not.toBeInTheDocument()
    // Exactly one command, and the revision the engine returned is the one on
    // the status line: one edit, one revision, one undo entry.
    expect(commands(request)).toHaveLength(1)
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 2')
  })

  // AC: the string the author reads is `===` to the engine's `message` field,
  // at the control that dispatched it. The three named refusals, each at its
  // own control, and each one PROVED TO HAVE BEEN SENT — a passing "the error
  // shows" assertion is satisfied just as well by a TypeScript copy of the
  // rule, so the dispatch is what distinguishes them.
  it('shows a duplicate-name refusal verbatim at the add control, having sent the command anyway', async () => {
    const engineMessage = 'a font chain named "body" already exists'
    const request = open(refuses(engineMessage, { dataPath: 'fonts.body' }))
    fireEvent.change(screen.getByRole('textbox', { name: 'New font chain name' }), { target: { value: 'body' } })
    fireEvent.change(screen.getByRole('textbox', { name: 'First entry for the new font chain' }), { target: { value: 'Noto Sans' } })
    fireEvent.click(screen.getByRole('button', { name: 'Add font chain' }))
    // ANTI-PRE-EMPTION: `body` is already declared and the command still goes.
    await waitFor(() => expect(commands(request)).toEqual(['{"kind":"addFontChain","version":1,"name":"body","entries":["Noto Sans"]}']))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    const field = screen.getByRole('textbox', { name: 'New font chain name' })
    expect(field).toHaveAttribute('aria-invalid', 'true')
    expect(field.getAttribute('aria-errormessage')).toBe(alert.id)
    // The document is unchanged — read off the PROJECTION, not off the name
    // fields, which are uncontrolled and are still showing the author's draft.
    // (`['body','heading']` in those fields would have looked identical had
    // the add succeeded, so they are no evidence either way.)
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem')).toHaveLength(1)
    expect(screen.queryByRole('list', { name: 'Entries of font chain 3' })).not.toBeInTheDocument()
  })

  it('shows an orphaning-delete refusal verbatim at the delete control, with the engine\'s own id list', async () => {
    const engineMessage = 'font chain "body" is still named by e1, e2 and 3 more'
    const request = open(refuses(engineMessage, { dataPath: 'fonts.body' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' }))
    await waitFor(() => expect(commands(request)).toEqual(['{"kind":"deleteFontChain","version":1,"name":"body"}']))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    const control = screen.getByRole('button', { name: 'Delete font chain 1' })
    expect(control).toHaveAttribute('aria-invalid', 'true')
    expect(control.getAttribute('aria-errormessage')).toBe(alert.id)
  })

  it('shows an emptying-remove refusal verbatim at that entry\'s remove control, having sent the command anyway', async () => {
    const engineMessage = 'removing that entry would leave font chain "body" with no entries'
    const request = open(refuses(engineMessage, { dataPath: 'fonts.body' }))
    // ANTI-PRE-EMPTION: chain 1 holds exactly one entry and the remove is
    // still dispatched. Nothing in the panel knows that emptying is refused.
    fireEvent.click(screen.getByRole('button', { name: 'Remove entry 1 of font chain 1' }))
    await waitFor(() => expect(commands(request)).toEqual(['{"kind":"removeFontChainEntry","version":1,"name":"body","index":0}']))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    const control = screen.getByRole('button', { name: 'Remove entry 1 of font chain 1' })
    expect(control).toHaveAttribute('aria-invalid', 'true')
    expect(control.getAttribute('aria-errormessage')).toBe(alert.id)
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem')).toHaveLength(1)
  })

  // The rename row of the matrix. Its refusal is the SAME sentence a duplicate
  // create gets, but located at the DESTINATION — DataPath `fonts.<to>`, not
  // `fonts.<name>` — because the destination key is the one that is taken. The
  // panel must not care: it anchors by the control it dispatched from, so a
  // located refusal lands at the rename control that asked, and the dataPath
  // does not drag the message to some other row.
  it('shows a rename-onto-a-declared-key refusal verbatim at the rename control, having sent the command anyway', async () => {
    const engineMessage = 'a font chain named "heading" already exists'
    const request = open(refuses(engineMessage, { dataPath: 'fonts.heading' }))
    fireEvent.change(screen.getByRole('textbox', { name: 'Font chain 1 name' }), { target: { value: 'heading' } })
    fireEvent.click(screen.getByRole('button', { name: 'Rename font chain 1' }))
    // ANTI-PRE-EMPTION: `heading` is right there in the list the panel is
    // drawing, and the panel still does not look. Renaming onto a declared key
    // is the engine's rule, so the command goes and the engine answers.
    await waitFor(() => expect(commands(request)).toEqual(['{"kind":"renameFontChain","version":1,"name":"body","to":"heading"}']))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    // Anchored at the RENAME control — the name field it was dispatched from.
    const control = screen.getByRole('textbox', { name: 'Font chain 1 name' })
    expect(control).toHaveAttribute('aria-invalid', 'true')
    expect(control.getAttribute('aria-errormessage')).toBe(alert.id)
    // The refusal located `fonts.heading` — chain 2 — and the message is still
    // beside chain 1, which is what asked. Placement is the browser's
    // knowledge, never the engine's DataPath.
    expect(screen.getByRole('textbox', { name: 'Font chain 2 name' })).not.toHaveAttribute('aria-invalid')
    expect(screen.getByRole('button', { name: 'Delete font chain 1' })).not.toHaveAttribute('aria-invalid')
    // Exactly one alert. And the document is unchanged — asserted against the
    // PROJECTION, because chain 1's name field is uncontrolled and is showing
    // the author's typed draft: `['heading','heading']` in those two fields
    // would have looked exactly the same had the rename SUCCEEDED, so they
    // are no evidence either way. The entry lists and the declared-font list
    // come from the snapshot.
    expect(screen.getAllByRole('alert')).toHaveLength(1)
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem')).toHaveLength(1)
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 2' })).getAllByRole('listitem')).toHaveLength(2)
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    expect(within(screen.getByRole('listbox', { name: 'Declared fonts' })).getAllByRole('option').map((option) => option.textContent)).toEqual(['body', 'heading'])
  })

  // The reorder row's ERROR cell. `fontChainIndex` refuses an out-of-range
  // `from`/`to` with this exact sentence, located at `fonts.<name>`; the panel
  // shows it at the move control that asked, unchanged and unprefixed.
  it('shows an out-of-range reorder refusal verbatim at the move control', async () => {
    const engineMessage = 'entry index is out of range'
    const request = open(refuses(engineMessage, { dataPath: 'fonts.heading' }))
    fireEvent.click(screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' }))
    await waitFor(() => expect(commands(request)).toEqual(['{"kind":"moveFontChainEntry","version":1,"name":"heading","from":0,"to":1}']))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    const control = screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' })
    expect(control).toHaveAttribute('aria-invalid', 'true')
    expect(control.getAttribute('aria-errormessage')).toBe(alert.id)
    // The counterpart control on the same row did not dispatch, so it is not
    // flagged; nor is the same-numbered entry of the other chain.
    expect(screen.getByRole('button', { name: 'Move entry 1 of font chain 2 earlier' })).not.toHaveAttribute('aria-invalid')
    expect(screen.getByRole('button', { name: 'Remove entry 1 of font chain 1' })).not.toHaveAttribute('aria-invalid')
    // The order shown afterwards is still the projection's: a refused move
    // moves nothing, because there was no local splice to undo. (The refused
    // row carries the alert, so the faces are read off the head of each item
    // rather than its whole text.)
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 2' })).getAllByRole('listitem').map((item) => item.textContent?.split('\u2191')[0])).toEqual(['Noto Sans', 'Noto Sans Thai'])
  })

  // An arity refusal, a projection-bound refusal and a serialize failure all
  // reach the browser as ENGINE_REJECTED with NO dataPath. The anchor is what
  // the panel asked, not what the engine returned, so they land identically.
  it('shows an unlocated refusal at the control that dispatched it', async () => {
    const engineMessage = 'folio: component command has unknown or missing fields'
    const request = open(refuses(engineMessage))
    fireEvent.click(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 earlier' }))
    await waitFor(() => expect(commands(request)).toHaveLength(1))
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toBe(engineMessage)
    expect(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 earlier' })).toHaveAttribute('aria-invalid', 'true')
    // And it is not shown on the OTHER move control for the same entry.
    expect(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 later' })).not.toHaveAttribute('aria-invalid')
  })

  it('reorders by keyboard alone and keeps the moved entry under the hand', async () => {
    const three: Chains = [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('A'), face('B'), face('C')] }]
    const moved: Chains = [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('B'), face('A'), face('C')] }]
    const request = vi.fn(async (operation: string) => operation === 'command' ? { snapshot: chainSnapshot(2, moved) } : { snapshot: chainSnapshot(1, three) })
    open(request, three)
    const later = screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' })
    later.focus()
    // Enter on a focused button is a click: the whole reorder is operable
    // from the keyboard alone (UX-DR25), with no pointer anywhere in it.
    fireEvent.click(later)
    await waitFor(() => expect(within(screen.getByRole('list', { name: 'Entries of font chain 2' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['B↑↓×', 'A↑↓×', 'C↑↓×']))
    // Focus followed the entry to its new position, so a second press moves
    // the SAME entry again rather than whichever one slid under the cursor.
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 later' }))
  })

  it('falls back to the counterpart control when the moved entry lands at an end', async () => {
    const moved: Chains = [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans Thai'), face('Noto Sans')] }]
    const request = vi.fn(async (operation: string) => operation === 'command' ? { snapshot: chainSnapshot(2, moved) } : { snapshot: chainSnapshot(1, declared) })
    open(request)
    const later = screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' })
    later.focus()
    fireEvent.click(later)
    await waitFor(() => expect(within(screen.getByRole('list', { name: 'Entries of font chain 2' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Noto Sans Thai↑↓×', 'Noto Sans↑↓×']))
    // The entry is now last, so "move later" is disabled there. Focus lands
    // on the row's other move control rather than on the document body.
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 earlier' }))
  })

  // P2. Pressing a control sets `busy`, which sets `disabled`, which blurs it
  // in a real browser. On an ACCEPTED command the follow-up focus puts the
  // hand back; on a REFUSED one nothing did, and the author was left on
  // `document.body` — unable to Tab to the `role="alert"` they had just
  // caused, with `aria-errormessage` hanging off a control they were no
  // longer on. jsdom does not blur a control the moment it is disabled, so
  // the browser's blur is applied explicitly here; the RESTORE is what is
  // under test.
  it('puts focus back on the refused control so the alert it points at is reachable', async () => {
    let settle!: (value: unknown) => void
    const request = vi.fn((operation: string) => operation === 'command'
      ? new Promise((_resolve, reject) => { settle = () => reject(Object.assign(new Error('font chain "body" is still named by e1'), { dataPath: 'fonts.body' })) })
      : Promise.resolve({ snapshot: chainSnapshot(1, declared) }))
    open(request as unknown as ReturnType<typeof vi.fn>)
    const control = screen.getByRole('button', { name: 'Delete font chain 1' })
    control.focus()
    fireEvent.click(control)
    await waitFor(() => expect(screen.getByRole('button', { name: 'Delete font chain 1' })).toBeDisabled())
    // Focus is demonstrably ELSEWHERE when the refusal lands — a real browser
    // puts it on `document.body` when the pressed control is disabled; here it
    // is moved to another control, which is the same starting position and is
    // deterministic in jsdom.
    const elsewhere = screen.getByRole('textbox', { name: 'Font size (pt)' })
    elsewhere.focus()
    expect(document.activeElement).toBe(elsewhere)
    settle(undefined)
    const alert = await screen.findByRole('alert')
    const restored = screen.getByRole('button', { name: 'Delete font chain 1' })
    expect(document.activeElement).toBe(restored)
    expect(restored.getAttribute('aria-errormessage')).toBe(alert.id)
  })

  // P1. The follow-up focus target used to be cleared only when `chains`
  // next changed. A refusal produces no new snapshot, so a refused move left
  // its targets alive indefinitely and the NEXT re-projection from any
  // unrelated cause — an ordinary property commit, an undo, a file load —
  // stole focus to a move button the author never pressed.
  it('does not steal focus on a later unrelated re-projection after a refused move', async () => {
    let refuse = true
    const request = vi.fn(async (operation: string) => {
      if (operation !== 'command') return { snapshot: chainSnapshot(1, declared) }
      if (refuse) throw Object.assign(new Error('entry index is out of range'), { dataPath: 'fonts.heading' })
      // The property commit's response carries a DIFFERENT chain listing —
      // the shape an undo or a file load has, delivered through a path that
      // keeps the panel mounted so the theft is observable at all. A listing
      // that did not change could not steal focus under any implementation,
      // so a fixture that reused the same chains would test nothing.
      return { snapshot: chainSnapshot(2, [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans SC')] }]) }
    })
    open(request)
    fireEvent.click(screen.getByRole('button', { name: 'Move entry 1 of font chain 2 later' }))
    await screen.findByRole('alert')
    // The author moves on to an ordinary property edit, which re-projects.
    refuse = false
    const elsewhere = screen.getByRole('textbox', { name: 'Font size (pt)' })
    elsewhere.focus()
    fireEvent.click(screen.getByRole('button', { name: 'Align center' }))
    // Wait for the RE-PROJECTION, not for the dispatch: `mock.calls` grows the
    // instant the command is sent, which is before the snapshot that would
    // trigger the theft has been installed.
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 2'))
    expect(document.activeElement).toBe(elsewhere)
    expect(document.activeElement).not.toBe(screen.getByRole('button', { name: 'Move entry 2 of font chain 2 earlier' }))
  })

  // P9. A refusal is about the edit that caused it. applyFontChain already
  // clears propertyError's sibling; the reverse was missing, so a refused
  // chain edit stayed rendered in the TYPOGRAPHY section while the author went
  // on committing font size and bold beside it.
  it('clears a standing chain refusal when an ordinary property commit succeeds', async () => {
    let refuse = true
    const request = vi.fn(async (operation: string) => {
      if (operation !== 'command') return { snapshot: chainSnapshot(1, declared) }
      if (refuse) throw Object.assign(new Error('font chain "body" is still named by e1'), { dataPath: 'fonts.body' })
      return { snapshot: chainSnapshot(2, declared) }
    })
    open(request)
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' }))
    await screen.findByRole('alert')
    refuse = false
    fireEvent.click(screen.getByRole('button', { name: 'Align center' }))
    await waitFor(() => expect(request.mock.calls.filter((call) => call[0] === 'command')).toHaveLength(2))
    await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument())
    expect(screen.getByRole('button', { name: 'Delete font chain 1' })).not.toHaveAttribute('aria-invalid')
  })

  // P10. The add fields are uncontrolled, so an accepted add used to leave the
  // author's text sitting in them. A second press then re-dispatched — and for
  // an entry the engine has NO duplicate rule to refuse it with, so it would
  // simply append the same face twice.
  it('empties the add fields once the add has actually landed, and not before', async () => {
    let accept = false
    const grown = [{ name: 'body', entries: [face('Noto Sans'), face('Noto Sans SC')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans Thai')] }]
    const request = vi.fn(async (operation: string) => {
      if (operation !== 'command') return { snapshot: chainSnapshot(1, declared) }
      if (!accept) throw Object.assign(new Error('font chain entry exceeds the projection bound'), { dataPath: 'fonts.body' })
      return { snapshot: chainSnapshot(2, grown) }
    })
    open(request)
    const field = screen.getByRole('textbox', { name: 'New entry for font chain 1' })
    fireEvent.change(field, { target: { value: 'Noto Sans SC' } })
    fireEvent.click(screen.getByRole('button', { name: 'Add entry to font chain 1' }))
    // REFUSED: the text stays, so the author can correct it rather than retype.
    await screen.findByRole('alert')
    expect(screen.getByRole('textbox', { name: 'New entry for font chain 1' })).toHaveValue('Noto Sans SC')
    accept = true
    fireEvent.click(screen.getByRole('button', { name: 'Add entry to font chain 1' }))
    await waitFor(() => expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem')).toHaveLength(2))
    // ACCEPTED: the field is empty, so a second press cannot silently append
    // the same face again.
    expect(screen.getByRole('textbox', { name: 'New entry for font chain 1' })).toHaveValue('')
  })

  it('empties both add-chain fields once the chain has actually landed', async () => {
    const grown = [...declared, { name: 'display', entries: [face('Noto Sans')] }]
    const request = vi.fn(async (operation: string) => operation === 'command' ? { snapshot: chainSnapshot(2, grown) } : { snapshot: chainSnapshot(1, declared) })
    open(request)
    fireEvent.change(screen.getByRole('textbox', { name: 'New font chain name' }), { target: { value: 'display' } })
    fireEvent.change(screen.getByRole('textbox', { name: 'First entry for the new font chain' }), { target: { value: 'Noto Sans' } })
    fireEvent.click(screen.getByRole('button', { name: 'Add font chain' }))
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'Font chain 3 name' })).toHaveValue('display'))
    expect(screen.getByRole('textbox', { name: 'New font chain name' })).toHaveValue('')
    expect(screen.getByRole('textbox', { name: 'First entry for the new font chain' })).toHaveValue('')
  })

  // P3. applyFontChain refuses to dispatch while the FILE path is busy, so a
  // control that stays enabled through a save or an open is a click that
  // silently does nothing and says nothing.
  it('disables every chain control while a file operation holds the document', async () => {
    const files: FileAccess = { open: vi.fn(() => new Promise<never>(() => undefined)), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    const request = accepts()
    render(<App engine={engine(request as never)} fileAccess={files} initialSnapshot={chainSnapshot(1, declared)} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    expect(screen.getByRole('button', { name: 'Delete font chain 1' })).toBeEnabled()
    fireEvent.click(screen.getByRole('button', { name: 'Open local template' }))
    await waitFor(() => expect(files.open).toHaveBeenCalledOnce())
    for (const name of ['Delete font chain 1', 'Rename font chain 1', 'Add font chain', 'Move entry 1 of font chain 2 later', 'Remove entry 1 of font chain 1']) {
      expect(screen.getByRole('button', { name })).toBeDisabled()
    }
    expect(screen.getByRole('textbox', { name: 'Font chain 1 name' })).toBeDisabled()
    expect(screen.getByRole('textbox', { name: 'New font chain name' })).toBeDisabled()
  })

  // P4. An accepted chain edit is a document mutation like any other: it makes
  // a rendered PDF stale, and it carries the engine's own undo availability.
  it('marks a rendered local PDF stale when a chain edit is accepted', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: chainSnapshot(1, declared), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: chainSnapshot(1, declared), bytes }
      if (operation === 'render') return { snapshot: chainSnapshot(1, declared), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      if (operation === 'command') return { snapshot: chainSnapshot(2, declared) }
      return { snapshot: chainSnapshot(1, declared) }
    })
    render(<App engine={engine(request as never)} initialSnapshot={chainSnapshot(1, declared)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    fireEvent.click(await screen.findByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Current exact local production PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Return to Design' }))
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' }))
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 2'))
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    // The admitted PDF is retained, and it is retained as STALE: the chain
    // edit changed the document the exact PDF was produced from.
    expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /Current exact local production PDF/ })).not.toBeInTheDocument()
  })

  // Both halves of the same response, and deliberately at an UNCHANGED
  // revision — which is what separates them. `setHistoryAvailability` runs on
  // every accepted response; the snapshot install runs only when the revision
  // ADVANCES, so a stale or replayed response cannot overwrite what is shown.
  it('takes undo availability from the response but does not install a revision that has not advanced', async () => {
    const other = [{ name: 'somewhere-else', entries: [face('Noto Sans')] }]
    const request = vi.fn(async (operation: string) => operation === 'command'
      ? { snapshot: { ...chainSnapshot(1, other), canUndo: true, canRedo: false } }
      : { snapshot: chainSnapshot(1, declared) })
    open(request)
    expect(screen.getByRole('button', { name: /^Undo/ })).toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 2' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /^Undo/ })).toBeEnabled())
    // The response's own chains are NOT shown: its revision (1) did not
    // advance past the installed one (1), so it was never installed.
    expect(screen.getAllByRole('textbox', { name: /^Font chain \d+ name$/ }).map((field) => (field as HTMLInputElement).value)).toEqual(['body', 'heading'])
    expect(screen.queryByDisplayValue('somewhere-else')).not.toBeInTheDocument()
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 1')
  })

  // P5. The asset path's own red proof (see "does not install a
  // setComponentAsset result that resolves after a document replacement"),
  // applied to the chain path. Two things must survive it: the stale result
  // must not be installed over the newer document, and `fontChainBusy` must
  // not be left stuck true — its reset is generation-guarded, so the document
  // replacement has to clear it or every chain control is dead for the session.
  it('does not install a chain result that resolves after a document replacement, and leaves no control stuck', async () => {
    const replaced = [{ name: 'from-the-undo', entries: [face('Noto Sans')] }]
    let settle!: () => void
    const held = new Promise((resolve) => { settle = () => resolve({ snapshot: chainSnapshot(2, declared) }) })
    const request = vi.fn((operation: string) => {
      if (operation === 'command') return held
      if (operation === 'undo') return Promise.resolve({ snapshot: { ...chainSnapshot(50, replaced), canUndo: false, canRedo: true } })
      return Promise.resolve({ snapshot: chainSnapshot(1, declared) })
    })
    render(<App engine={engine(request as never)} initialSnapshot={{ ...chainSnapshot(1, declared), canUndo: true, canRedo: false }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Delete font chain 1' })).toBeDisabled())

    // A DOCUMENT REPLACEMENT lands while the chain command is still in flight.
    fireEvent.click(screen.getByRole('button', { name: /^Undo/ }))
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 50'))

    // NOW the stale command completes.
    settle()
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 50')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()

    // And the editor still works: the busy flag was reset by the replacement,
    // not by the stale response's generation-guarded finally.
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    expect(screen.getByRole('textbox', { name: 'Font chain 1 name' })).toHaveValue('from-the-undo')
    expect(screen.getByRole('button', { name: 'Delete font chain 1' })).toBeEnabled()
  })

  // The same replacement, with the stale command REFUSING rather than
  // succeeding: the refusal belongs to a document that is no longer open, so
  // it must not be rendered against a control in the one that is.
  it('does not show a chain refusal that resolves after a document replacement', async () => {
    // The replacing document declares a chain of the SAME NAME, so the stale
    // refusal's control still EXISTS to be anchored to. A replacement that
    // renamed everything would drop the message for the wrong reason — no
    // matching control — and prove nothing about the generation guard.
    const replaced = [{ name: 'body', entries: [face('Noto Sans SC')] }]
    let settle!: () => void
    const held = new Promise((_resolve, reject) => { settle = () => reject(Object.assign(new Error('font chain "body" is still named by e1'), { dataPath: 'fonts.body' })) })
    const request = vi.fn((operation: string) => {
      if (operation === 'command') return held
      if (operation === 'undo') return Promise.resolve({ snapshot: { ...chainSnapshot(50, replaced), canUndo: false, canRedo: true } })
      return Promise.resolve({ snapshot: chainSnapshot(1, declared) })
    })
    render(<App engine={engine(request as never)} initialSnapshot={{ ...chainSnapshot(1, declared), canUndo: true, canRedo: false }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    fireEvent.click(screen.getByRole('button', { name: 'Delete font chain 1' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Delete font chain 1' })).toBeDisabled())
    fireEvent.click(screen.getByRole('button', { name: /^Undo/ }))
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 50'))
    settle()
    await new Promise((resolve) => setTimeout(resolve, 0))
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Edit font chains' }))
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Noto Sans SC↑↓×'])
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Delete font chain 1' })).not.toHaveAttribute('aria-invalid')
  })

  it('accepts a chain entry naming a face this build does not ship, displayed as the projection spells it', () => {
    open(accepts(), [{ name: 'body', entries: [face('Helvetica'), face('Noto Sans')] }])
    expect(within(screen.getByRole('list', { name: 'Entries of font chain 1' })).getAllByRole('listitem').map((item) => item.textContent)).toEqual(['Helvetica↑↓×', 'Noto Sans↑↓×'])
  })
})
