import { render, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

const state = vi.hoisted(() => ({ loadingDestroy: vi.fn(), documentDestroy: vi.fn(), cancel: vi.fn(), render: vi.fn(), getDocument: vi.fn(), workerSrc: '' }))

vi.mock('pdfjs-dist/build/pdf.mjs', () => ({
  GlobalWorkerOptions: { get workerSrc() { return state.workerSrc }, set workerSrc(value: string) { state.workerSrc = value } },
  getDocument: state.getDocument,
}))
vi.mock('pdfjs-dist/build/pdf.worker.mjs?url', () => ({ default: '/assets/pdf.worker-local.mjs' }))

import { PDFPreviewViewer } from './pdf-viewer'

// The mock's viewport scales, so a test can tell the raster viewport from
// the displayed one — the whole point of the oversampled canvas.
const pdf = (numPages = 1) => ({
  numPages,
  getPage: vi.fn(async () => ({ getViewport: ({ scale }: { scale: number }) => ({ width: 20 * scale, height: 30 * scale }), render: state.render })),
  destroy: state.documentDestroy,
})
const viewerProps = (overrides = {}) => ({ bytes: new Uint8Array([1, 2, 3]).buffer, label: 'Exact PDF', describedBy: 'preview-freshness-status', state: { page: 1, scale: 1, scrollTop: 0, scrollLeft: 0 }, onStateChange: vi.fn(), onError: vi.fn(), onPageCount: vi.fn(), ...overrides })

describe('local PDF preview owner', () => {
  afterEach(() => { vi.restoreAllMocks(); state.loadingDestroy.mockReset(); state.documentDestroy.mockReset(); state.cancel.mockReset(); state.render.mockReset(); state.getDocument.mockReset(); state.workerSrc = '' })

  it('copies opaque bytes, uses only the local worker, and destroys every owned resource once', async () => {
    const task = { promise: Promise.resolve(), cancel: state.cancel }
    state.render.mockReturnValue(task)
    state.getDocument.mockReturnValue({ promise: Promise.resolve(pdf()), destroy: state.loadingDestroy })
    const context = {} as CanvasRenderingContext2D
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(context)
    const create = vi.spyOn(URL, 'createObjectURL')
    const revoke = vi.spyOn(URL, 'revokeObjectURL')
    const source = new Uint8Array([1, 2, 3]).buffer
    const view = render(<PDFPreviewViewer {...viewerProps({ bytes: source })} />)
    await waitFor(() => expect(state.render).toHaveBeenCalledOnce())
    const options = state.getDocument.mock.calls[0]![0] as { data: Uint8Array; disableAutoFetch: boolean; disableStream: boolean; useWorkerFetch: boolean; isEvalSupported: boolean; cMapUrl: string; cMapPacked: boolean; standardFontDataUrl: string }
    expect(options.data).toEqual(new Uint8Array([1, 2, 3]))
    expect(options.data.buffer).not.toBe(source)
    expect(options).toMatchObject({ disableAutoFetch: true, disableStream: true, useWorkerFetch: false, isEvalSupported: false, cMapUrl: expect.stringMatching(/^\/assets\/pdfjs-cmaps-[a-f0-9]{20}\/$/), cMapPacked: true, standardFontDataUrl: expect.stringMatching(/^\/assets\/pdfjs-standard-fonts-[a-f0-9]{20}\/$/) })
    expect(state.workerSrc).toBe('/assets/pdf.worker-local.mjs')
    expect(create).not.toHaveBeenCalled(); expect(revoke).not.toHaveBeenCalled()
    view.unmount()
    await waitFor(() => expect(state.cancel).toHaveBeenCalledOnce())
    expect(state.documentDestroy).toHaveBeenCalledOnce(); expect(state.loadingDestroy).toHaveBeenCalledOnce()
  })

  it('does not install a late document after unmount and reports a real PDF.js failure', async () => {
    let resolve!: (value: ReturnType<typeof pdf>) => void
    state.getDocument.mockReturnValue({ promise: new Promise((done) => { resolve = done }), destroy: state.loadingDestroy })
    const onError = vi.fn()
    const view = render(<PDFPreviewViewer {...viewerProps({ bytes: new Uint8Array([7]).buffer, onError })} />)
    await waitFor(() => expect(state.getDocument).toHaveBeenCalledOnce())
    view.unmount(); resolve(pdf())
    await Promise.resolve(); await Promise.resolve()
    expect(state.documentDestroy).toHaveBeenCalledOnce()
    expect(onError).not.toHaveBeenCalled()

    state.getDocument.mockReturnValue({ promise: Promise.reject(new Error('bad PDF')), destroy: state.loadingDestroy })
    render(<PDFPreviewViewer {...viewerProps({ bytes: new Uint8Array([8]).buffer, onError })} />)
    await waitFor(() => expect(onError).toHaveBeenCalledWith(expect.objectContaining({ message: 'bad PDF' })))
  })

  it('reports an honest 50-page sequence and delegates page navigation as transient state', async () => {
    const onStateChange = vi.fn()
    state.render.mockReturnValue({ promise: Promise.resolve(), cancel: state.cancel })
    state.getDocument.mockReturnValue({ promise: Promise.resolve(pdf(50)), destroy: state.loadingDestroy })
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({} as CanvasRenderingContext2D)
    const { getByRole } = render(<PDFPreviewViewer {...viewerProps({ onStateChange })} />)
    await waitFor(() => expect(getByRole('status', { name: 'PDF page status' })).toHaveTextContent('Page 1 of 50'))
    getByRole('button', { name: 'Next PDF page' }).click()
    expect(onStateChange).toHaveBeenCalledWith({ page: 2, scale: 1, scrollTop: 0, scrollLeft: 0 })
  })

  it('rasterizes above the displayed size, so the page is not magnified on a denser display', async () => {
    state.render.mockReturnValue({ promise: Promise.resolve(), cancel: state.cancel })
    state.getDocument.mockReturnValue({ promise: Promise.resolve(pdf()), destroy: state.loadingDestroy })
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue({} as CanvasRenderingContext2D)
    const view = render(<PDFPreviewViewer {...viewerProps()} />)
    await waitFor(() => expect(state.render).toHaveBeenCalledOnce())
    const canvas = view.container.querySelector('canvas')!
    // Displayed at the page's own size; rasterized at twice it.
    expect(canvas.style.width).toBe('20px')
    expect(canvas.style.height).toBe('30px')
    expect(canvas.width).toBe(40)
    expect(canvas.height).toBe(60)
    // And pdf.js is asked to paint the LARGER viewport, not the shown one.
    expect((state.render.mock.calls[0]![0] as { viewport: { width: number } }).viewport.width).toBe(40)
  })
})
