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
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas }

describe('docked sample data panel', () => {
  it('keeps authoring available when empty, loads a tree, keeps accepted bytes authoritative, and preserves a prior sample on cancel', async () => {
    const openSample = vi.fn<SampleFileAccess['openSample']>().mockResolvedValueOnce({ name: 'sample.json', bytes: sampleBytes }).mockRejectedValueOnce(new FileAccessCancelled()).mockResolvedValueOnce({ name: 'replacement.json', bytes: replacementBytes })
    const request = vi.fn(async (...args: [string, unknown?, AbortSignal?]) => args[0] === 'identity' ? { snapshot, preview: { revision: 1, identity: 'b'.repeat(64) } } : args[0] === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : args[0] === 'render' ? { snapshot, bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} sampleFileAccess={{ openSample }} />)
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

  it('revokes a pending picker when an equal-revision Start blank replaces the document', async () => {
    let release!: (value: { name: string; bytes: ArrayBuffer }) => void
    const openSample = vi.fn(() => new Promise<{ name: string; bytes: ArrayBuffer }>((resolve) => { release = resolve }))
    const request = vi.fn(async (operation: string) => operation === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} blankBytes={new Uint8Array([7]).buffer} sampleFileAccess={{ openSample }} />)
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
