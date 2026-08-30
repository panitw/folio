import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { DataPanel } from './DataPanel'
import { acceptSampleData } from './sample-data'
import type { EngineClient } from './engine-client'
import { FileAccessCancelled } from './file/file-access'
import type { SampleFileAccess } from './sample-file'

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => true,
  PDFPreviewViewer: () => null,
}))

const sampleBytes = new TextEncoder().encode('{"customer":{"name":"Ada"},"items":[{"sku":"A-1"}]}').buffer
const replacementBytes = new TextEncoder().encode('{"report":{"id":2}}').buffer
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body'], defaultFontSize: 12000, contentWindowHeight: 729890, contentWindowCount: 1, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas }

const textCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Text' }] }

const openDataTab = () => fireEvent.click(screen.getByRole('tab', { name: 'DATA' }))

describe('docked sample data panel', () => {
  it('keeps authoring available when empty, loads a tree, keeps accepted bytes authoritative, and preserves a prior sample on cancel', async () => {
    const openSample = vi.fn<SampleFileAccess['openSample']>().mockResolvedValueOnce({ name: 'sample.json', bytes: sampleBytes }).mockRejectedValueOnce(new FileAccessCancelled()).mockResolvedValueOnce({ name: 'replacement.json', bytes: replacementBytes })
    const request = vi.fn(async (...args: [string, unknown?, AbortSignal?]) => args[0] === 'identity' ? { snapshot, preview: { revision: 1, identity: 'b'.repeat(64) } } : args[0] === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : args[0] === 'render' ? { snapshot, bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} sampleFileAccess={{ openSample }} />)
    openDataTab()
    expect(screen.getByLabelText('Data panel')).toBeInTheDocument()
    expect(screen.getByText('Binding unavailable: no sample data loaded.')).toBeInTheDocument()
    expect(screen.getByLabelText('Canvas region')).toBeInTheDocument()
    const load = screen.getByRole('button', { name: 'Load sample JSON' }); load.focus(); expect(load).toHaveFocus()
    fireEvent.click(load)
    await waitFor(() => expect(screen.getByText('Local sample:')).toBeInTheDocument())
    expect(screen.getByRole('tree', { name: 'Sample data paths' })).toHaveTextContent('items[]')
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(request.mock.calls.some(([operation]) => operation === 'identity')).toBe(true))
    const data = request.mock.calls.find(([operation]) => operation === 'identity')![1] as unknown as { data: ArrayBuffer }
    expect(new Uint8Array(data.data)).toEqual(new Uint8Array(sampleBytes))
    fireEvent.click(screen.getByRole('button', { name: 'Return to Design' }))
    fireEvent.click(screen.getByRole('button', { name: 'Replace sample JSON' }))
    await waitFor(() => expect(screen.getByText('sample.json')).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Replace sample JSON' }))
    await waitFor(() => expect(screen.getByText('replacement.json')).toBeInTheDocument())
  })

  it('provides one roving tree-item route through branches and scalar leaves', async () => {
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada","none":null},"items":[]}').buffer)
    render(<DataPanel sample={sample} busy={false} available onLoad={() => undefined} />)
    const root = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '1')!
    root.focus(); expect(root).toHaveFocus()
    fireEvent.keyDown(root, { key: 'ArrowDown' })
    const customer = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('customer'))!
    await waitFor(() => expect(customer).toHaveFocus())
    fireEvent.keyDown(customer, { key: 'ArrowRight' })
    fireEvent.keyDown(customer, { key: 'ArrowDown' })
    const name = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '3' && item.textContent?.startsWith('name'))!
    await waitFor(() => expect(name).toHaveFocus())
    fireEvent.keyDown(name, { key: 'End' })
    const items = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('items'))!
    await waitFor(() => expect(items).toHaveFocus())
    expect(customer).toHaveAttribute('aria-expanded', 'true')
  })

  it('restores the root tab stop when replacing a sample after nested navigation', async () => {
    const first = acceptSampleData('first.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    const replacement = acceptSampleData('replacement.json', replacementBytes)
    const { rerender } = render(<DataPanel sample={first} busy={false} available onLoad={() => undefined} />)
    const customer = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!
    customer.focus(); fireEvent.keyDown(customer, { key: 'ArrowRight' })
    const name = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!
    fireEvent.keyDown(name, { key: 'ArrowDown' })
    rerender(<DataPanel sample={replacement} busy={false} available onLoad={() => undefined} />)
    await waitFor(() => {
      const items = screen.getAllByRole('treeitem')
      expect(items.filter((item) => item.tabIndex === 0)).toHaveLength(1)
      expect(items[0]).toHaveAttribute('tabindex', '0')
    })
    const root = screen.getAllByRole('treeitem')[0]!
    root.focus(); fireEvent.keyDown(root, { key: 'ArrowDown' })
    await waitFor(() => expect(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('report'))).toHaveFocus())
  })

  it('shows a binding rejection only for its original sample, component, and picked path', () => {
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada","email":"a@example.test"}}').buffer)
    const error = { sample, componentID: 'e1', segments: ['customer', 'name'], message: 'e1: binding rejected' }
    render(<DataPanel sample={sample} busy={false} available selectedComponentId="e1" bindingError={error} onLoad={() => undefined} />)
    const customer = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!
    fireEvent.click(customer)
    const name = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!
    fireEvent.click(name)
    expect(screen.getByRole('alert')).toHaveTextContent('binding rejected')
    const email = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('email'))!
    fireEvent.click(email)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('selects a root scalar by keyboard and sends one opaque binding command, then paints distinct binding state', async () => {
    const boundCanvas = { ...textCanvas, components: [{ ...textCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }] }
    const request = vi.fn(async (operation: string) => operation === 'command'
      ? { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 4, canUndo: true, canvas: boundCanvas } }
      : { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: textCanvas } })
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"},"items":[]}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} initialSampleData={sample} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    const customer = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('customer'))!
    customer.focus()
    fireEvent.keyDown(customer, { key: 'ArrowRight' })
    const name = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '3' && item.textContent?.startsWith('name'))!
    name.focus()
    fireEvent.keyDown(name, { key: 'Enter' })
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).not.toHaveLength(0))
    const commands = request.mock.calls.filter(([operation]) => operation === 'command') as unknown as Array<[string, ArrayBuffer]>
    expect(commands).toHaveLength(1)
    const [operation, payload] = commands[0]!
    expect(operation).toBe('command')
    expect(new TextDecoder().decode(payload)).toBe('{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}')
    await waitFor(() => expect(screen.getByText('Bound to').parentElement).toHaveTextContent('Bound to customer.name'))
    expect(screen.getByLabelText('text component e1; bound to customer.name')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled()
  })

  it('does not install a late binding response after Start blank replaces the document', async () => {
    let resolveBinding!: (value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof textCanvas } }) => void
    const boundCanvas = { ...textCanvas, components: [{ ...textCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }] }
    const request = vi.fn((operation: string) => {
      if (operation === 'command') return new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof textCanvas } }>((resolve) => { resolveBinding = resolve })
      if (operation === 'load') return Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 1, canvas: textCanvas } })
      return Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: textCanvas } })
    })
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} initialSampleData={sample} blankBytes={new Uint8Array([7]).buffer} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    const customer = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('customer'))!
    fireEvent.click(customer)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '3' && item.textContent?.startsWith('name'))!)
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Started an unnamed local template')).toBeInTheDocument())
    resolveBinding({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 4, canvas: boundCanvas } })
    await Promise.resolve(); await Promise.resolve()
    expect(screen.queryByText('Bound to')).not.toBeInTheDocument()
    expect(screen.queryByText('customer.name')).not.toBeInTheDocument()
  })

  it('installs a committed binding after reselection while keeping the newer selection', async () => {
    const selectedCanvas = { ...textCanvas, components: [...textCanvas.components, { id: 'e2', type: 'text' as const, band: 'content' as const, x: 80_000, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Other' }] }
    const boundCanvas = { ...selectedCanvas, components: [{ ...selectedCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }, selectedCanvas.components[1]!] }
    let resolveBinding!: (value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof selectedCanvas } }) => void
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof selectedCanvas } }>((resolve) => { resolveBinding = resolve }) : Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: selectedCanvas } }))
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: selectedCanvas }} initialSampleData={sample} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(1))
    fireEvent.click(screen.getByLabelText('text component e2'))
    resolveBinding({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 4, canvas: boundCanvas } })
    await waitFor(() => expect(screen.getByLabelText('text component e1; bound to customer.name')).toBeInTheDocument())
    expect(screen.getByLabelText('text component e2')).toHaveClass('canvas-component-selected')
  })

  it('revokes a pending picker when an equal-revision Start blank replaces the document', async () => {
    let release!: (value: { name: string; bytes: ArrayBuffer }) => void
    const openSample = vi.fn(() => new Promise<{ name: string; bytes: ArrayBuffer }>((resolve) => { release = resolve }))
    const request = vi.fn(async (operation: string) => operation === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} blankBytes={new Uint8Array([7]).buffer} sampleFileAccess={{ openSample }} />)
    openDataTab()
    fireEvent.click(screen.getByRole('button', { name: 'Load sample JSON' }))
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Started an unnamed local template')).toBeInTheDocument())
    release({ name: 'late.json', bytes: sampleBytes })
    await Promise.resolve(); await Promise.resolve()
    expect(screen.queryByText('late.json')).not.toBeInTheDocument()
    expect(screen.getByText('Binding unavailable: no sample data loaded.')).toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(0)
  })
})
