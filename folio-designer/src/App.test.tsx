import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { FileAccessCancelled, type FileAccess } from './file/file-access'
import type { EngineClient } from './engine-client'

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
    expect(screen.getByText('PREVIEW · later')).toBeInTheDocument()
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
})
