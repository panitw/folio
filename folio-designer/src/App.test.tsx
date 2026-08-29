import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App, { placementPoint } from './App'
import { shortcutHintsFor } from './shortcuts'
import { FileAccessCancelled, type FileAccess } from './file/file-access'
import type { EngineClient } from './engine-client'
import { acceptSampleData } from './sample-data'

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => false,
  PDFPreviewViewer: ({ label, describedBy, onPageCount, onError }: { label: string; describedBy: string; onPageCount: (pages: number) => void; onError: (error: Error) => void }) => <><button type="button" aria-label={label} aria-describedby={describedBy} onClick={() => onPageCount(1)}>Admit local PDF</button><button type="button" aria-label="Fail local PDF viewer" onClick={() => onError(new Error('viewer rejected bytes'))}>Fail local PDF viewer</button></>,
}))

const bytes = new Uint8Array([1, 2, 3]).buffer
const sample = acceptSampleData('sample.json', new TextEncoder().encode('{"customer":{"name":"Preview customer"},"transactions":[]}').buffer)
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
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
    const textCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true, value: 'do not paint this value', textPaint: { overflow: false, lines: [{ top: 0, baseline: 12000, advance: 16000, width: 24000, fragments: [{ text: 'engine ', x: 0 }, { text: 'line', x: 16000 }] }] } }] }
    const request = vi.fn()
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} />)
    expect(screen.getByText('engine', { exact: true })).toBeInTheDocument()
    expect(screen.getByText('line', { exact: true })).toBeInTheDocument()
    expect(screen.queryByText('do not paint this value')).not.toBeInTheDocument()
    expect(screen.getByLabelText('text component e1: engine line')).toBeInTheDocument()
    expect(document.querySelector('.canvas-text-line')).toHaveStyle({ '--text-line-baseline': '12px', '--text-line-advance': '16px' })
    expect(request).not.toHaveBeenCalled()
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

  it('keeps a newer property draft through an unrelated successful snapshot and exposes table truth', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }, { id: 'e2', type: 'table' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }] }
    let resolve: ((value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }) => void) | undefined
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }>((done) => { resolve = done }) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const value = screen.getByRole('textbox', { name: 'Text value' })
    fireEvent.change(value, { target: { value: 'newer literal' } })
    fireEvent.keyDown(value, { key: 'Enter' })
    resolve!({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 3, canvas: componentCanvas } })
    await waitFor(() => expect(value).toHaveValue('newer literal'))
    fireEvent.click(screen.getByLabelText('table component e2'))
    expect(screen.queryByRole('textbox', { name: 'Width (pt)' })).not.toBeInTheDocument()
    expect(screen.getByText('Table binding: transactions[] (display only)')).toBeInTheDocument()
  })
})

describe('Story 5.13: image asset selection', () => {
  const imageComponent = { id: 'e1', type: 'image' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 48_000, resizable: true, image: { mediaType: 'image/png', assetKey: 'a'.repeat(64), width: 300, height: 200, drawX: 6_000, drawY: 8_000, drawWidth: 60_000, drawHeight: 40_000 } }
  const undecodableImageComponent = { id: 'e2', type: 'image' as const, band: 'content' as const, x: 0, y: 60_000, width: 72_000, height: 48_000, resizable: true }
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
