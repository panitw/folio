import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App, { placementPoint } from './App'
import { shortcutHintsFor } from './shortcuts'
import { FileAccessCancelled, type FileAccess } from './file/file-access'
import type { EngineClient } from './engine-client'

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => false,
  PDFPreviewViewer: ({ label, describedBy, onPageCount }: { label: string; describedBy: string; onPageCount: (pages: number) => void }) => <button type="button" aria-label={label} aria-describedby={describedBy} onClick={() => onPageCount(1)}>Admit local PDF</button>,
}))

const bytes = new Uint8Array([1, 2, 3]).buffer
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = (revision: number) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas })
const engine = (request = vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3 }, ...(operation === 'serialize' ? { bytes } : {}) }))) => ({ request }) as unknown as EngineClient

describe('application shell', () => {
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

  it('replaces the canvas with Preview, cancels an older render, and never dirties or installs its late PDF', async () => {
    let releaseSerialize!: (value: { snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }) => void
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } })
      if (operation === 'serialize') return new Promise<{ snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }>((resolve) => { releaseSerialize = resolve })
      if (operation === 'render') return Promise.resolve({ snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    expect(screen.queryByLabelText('Canvas region')).not.toBeInTheDocument()
    expect(screen.getByText('Rendering local PDF')).toBeInTheDocument()
    await waitFor(() => expect(request).toHaveBeenCalledWith('serialize', undefined, expect.any(AbortSignal)))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel and return to Design' }))
    releaseSerialize({ snapshot: snapshot(1), bytes })
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.queryByText(/Go production digest/)).not.toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(request.mock.calls.map(([operation]) => operation)).toEqual(['identity', 'serialize'])
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
      render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
      fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw sample data JSON' }), { target: { value: '{"transactions":[1]}' } })
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw sample data JSON' }), { target: { value: '{"transactions":[2]}' } })
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

  it('waits for matching PDF.js admission before claiming current exact output', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
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
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByLabelText('Render diagnostics')).toBeInTheDocument())
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw sample data JSON' }), { target: { value: '{"transactions":[1]}' } })
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
  })

  it('returns to Design and announces an unavailable authoritative warning target without selecting a substitute', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [{ severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'gone', dataPath: '', message: 'Content was clipped' }] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Locate in Design' })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Locate in Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.getByText('Locate unavailable: the authoritative element is no longer present.')).toHaveAttribute('role', 'status')
  })

  it('returns from a path-only render failure without requiring an element id', async () => {
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', dataPath: 'items[0]' })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') throw failure
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByLabelText('Local render failure')).toBeInTheDocument())
    fireEvent.click(within(screen.getByLabelText('Local render failure')).getByRole('button', { name: 'Return to Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
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
