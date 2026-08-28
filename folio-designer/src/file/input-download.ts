import { FileAccessCancelled, FileAccessFailure, folioName, type AcquiredSaveTarget, type FileAccess, type LocalFile, type SaveRequest, type SaveTargetRequest, type SavedLocalFile } from './file-access'

export type DownloadUrl = Readonly<{
  createObjectURL(object: Blob): string
  revokeObjectURL(url: string): void
}>

// This tier intentionally retains no target: browser downloads cannot promise
// that a previously opened file was overwritten or that permission survives.
export class InputDownloadAccess implements FileAccess {
  private readonly document: Document
  private readonly url: DownloadUrl

  constructor(document: Document = window.document, url: DownloadUrl = URL) {
    this.document = document
    this.url = url
  }

  open(): Promise<LocalFile> {
    return new Promise((resolve, reject) => {
      const input = this.document.createElement('input')
      input.type = 'file'
      input.accept = '.folio,application/json'
      input.value = '' // a second selection of the same file must still notify us
      input.style.display = 'none'
      const cleanup = () => input.remove()
      input.addEventListener('change', () => {
        const file = input.files?.item(0)
        cleanup()
        if (!file) { reject(new FileAccessCancelled()); return }
        void file.arrayBuffer().then((bytes) => resolve({ bytes, name: file.name }), () => reject(new FileAccessFailure('Could not open local file')))
      }, { once: true })
      input.addEventListener('cancel', () => { cleanup(); reject(new FileAccessCancelled()) }, { once: true })
      this.document.body.append(input)
      input.click()
    })
  }

  async acquireSaveTarget(request: SaveTargetRequest): Promise<AcquiredSaveTarget> {
    // Downloads have no retained overwrite permission. Naming is still chosen
    // synchronously at the user's Save/Save As gesture.
    return { name: folioName(request.suggestedName) }
  }

  async writeSave(target: AcquiredSaveTarget, request: SaveRequest): Promise<SavedLocalFile> {
    let href: string | undefined
    let anchor: HTMLAnchorElement | undefined
    try {
      const blob = new Blob([request.bytes], { type: 'application/json' })
      href = this.url.createObjectURL(blob)
      anchor = this.document.createElement('a')
      anchor.href = href
      anchor.download = target.name
      anchor.style.display = 'none'
      this.document.body.append(anchor)
      anchor.click()
      return { name: target.name }
    } catch {
      throw new FileAccessFailure('Could not download local file')
    } finally {
      anchor?.remove()
      if (href) queueMicrotask(() => this.url.revokeObjectURL(href!))
    }
  }
}
