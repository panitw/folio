import { useEffect, useRef, useState } from 'react'
import type { PDFDocumentLoadingTask, PDFDocumentProxy, RenderTask } from 'pdfjs-dist/build/pdf.mjs'
import workerUrl from 'pdfjs-dist/build/pdf.worker.mjs?url'
import { pdfjsRuntimeAssets, pdfjsViewerAssets } from '../generated/pdfjs-assets'

// PDF.js is strictly a local rasterizer for bytes the Go renderer returned.
// The worker URL is a Vite-managed immutable release asset, never a CDN URL.

type ActiveDocument = { loading: PDFDocumentLoadingTask; document?: PDFDocumentProxy; task?: RenderTask; canvas?: HTMLCanvasElement }

export type PDFPreviewViewState = Readonly<{ page: number; scale: number; scrollTop: number; scrollLeft: number }>
export const initialPDFPreviewViewState: PDFPreviewViewState = { page: 1, scale: 1, scrollTop: 0, scrollLeft: 0 }
export const samePDFPreviewViewState = (left: PDFPreviewViewState, right: PDFPreviewViewState) => left.page === right.page && left.scale === right.scale && left.scrollTop === right.scrollTop && left.scrollLeft === right.scrollLeft

export type PDFPreviewViewerProps = Readonly<{ bytes: ArrayBuffer; label: string; state: PDFPreviewViewState; onStateChange: (state: PDFPreviewViewState) => void; onError: (error: Error) => void; onPageCount: (pages: number) => void }>

export function PDFPreviewViewer({ bytes, label, state, onStateChange, onError, onPageCount }: PDFPreviewViewerProps) {
  const host = useRef<HTMLDivElement>(null)
  const active = useRef<ActiveDocument | undefined>(undefined)
  const [pages, setPages] = useState<number>()
  const page = state.page
  const scale = state.scale

  useEffect(() => {
    let cancelled = false
    const dispose = async () => {
      const current = active.current
      active.current = undefined
      current?.task?.cancel()
      current?.canvas?.remove()
      try { await current?.document?.destroy() } catch { /* cancellation is expected */ }
      try { await current?.loading.destroy() } catch { /* cancellation is expected */ }
    }
    const render = async () => {
      await dispose()
      // Referencing the generated map keeps every support byte in the release
      // set. The two URLs are local, content-addressed directories whose
      // canonical filenames PDF.js appends itself.
      void pdfjsRuntimeAssets
      const pdfjs = await import('pdfjs-dist/build/pdf.mjs')
      pdfjs.GlobalWorkerOptions.workerSrc = workerUrl
      if (cancelled) return
      const loading = pdfjs.getDocument({ data: new Uint8Array(bytes.slice(0)), disableAutoFetch: true, disableStream: true, stopAtErrors: true, isEvalSupported: false, useWorkerFetch: false, cMapUrl: pdfjsViewerAssets.cMapUrl, cMapPacked: pdfjsViewerAssets.cMapPacked, standardFontDataUrl: pdfjsViewerAssets.standardFontDataUrl })
      const current: ActiveDocument = { loading }
      active.current = current
      try {
        const document = await loading.promise
        if (cancelled || active.current !== current) { await document.destroy(); return }
        current.document = document
        setPages(document.numPages); onPageCount(document.numPages)
        const safePage = Math.min(Math.max(1, page), document.numPages)
        if (safePage !== page) onStateChange({ ...state, page: safePage })
        const pdfPage = await document.getPage(safePage)
        if (cancelled || active.current !== current) return
        const viewport = pdfPage.getViewport({ scale })
        const canvas = window.document.createElement('canvas')
        canvas.width = Math.ceil(viewport.width); canvas.height = Math.ceil(viewport.height)
        canvas.setAttribute('aria-label', `${label}, page ${safePage} of ${document.numPages}; visual PDF canvas`)
        const context = canvas.getContext('2d')
        if (!context) throw new Error('PDF canvas is unavailable')
        current.canvas = canvas
        host.current?.replaceChildren(canvas)
        current.task = pdfPage.render({ canvasContext: context, viewport })
        await current.task.promise
      } catch (reason) {
        if (!cancelled && active.current === current) onError(reason instanceof Error ? reason : new Error('PDF preview could not be rendered'))
      }
    }
    void render()
    return () => { cancelled = true; void dispose() }
  }, [bytes, label, page, scale, state, onError, onPageCount, onStateChange])

  useEffect(() => {
    if (!host.current) return
    host.current.scrollTop = state.scrollTop
    host.current.scrollLeft = state.scrollLeft
  }, [state.scrollLeft, state.scrollTop])

  return <section className="pdf-preview" aria-label="Exact local production PDF preview">
    <div className="pdf-preview-controls" aria-label="PDF viewer controls">
      <button type="button" onClick={() => onStateChange({ ...state, page: Math.max(1, page - 1) })} disabled={page <= 1} aria-label="Previous PDF page">Previous</button>
      <output aria-live="polite" aria-label="PDF page status">{pages ? `Page ${page} of ${pages}` : 'Rendering PDF'}</output>
      <button type="button" onClick={() => onStateChange({ ...state, page: pages ? Math.min(pages, page + 1) : page + 1 })} disabled={!pages || page >= pages} aria-label="Next PDF page">Next</button>
      <button type="button" onClick={() => onStateChange({ ...state, scale: Math.max(0.5, scale - 0.1) })} aria-label="Zoom out PDF">−</button>
      <output aria-label="PDF zoom">{Math.round(scale * 100)}%</output>
      <button type="button" onClick={() => onStateChange({ ...state, scale: Math.min(2, scale + 0.1) })} aria-label="Zoom in PDF">+</button>
    </div>
    <div className="pdf-preview-scroll" ref={host} role="img" aria-label={label} tabIndex={0} onScroll={(event) => onStateChange({ ...state, scrollTop: event.currentTarget.scrollTop, scrollLeft: event.currentTarget.scrollLeft })} />
  </section>
}
