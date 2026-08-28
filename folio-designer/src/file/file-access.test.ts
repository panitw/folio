import { describe, expect, it, vi } from 'vitest'
import { selectFileAccess } from './capability'
import { FileAccessCancelled, type LocalFileHandle } from './file-access'
import { FileSystemAccess } from './file-system-access'
import { InputDownloadAccess } from './input-download'

const bytes = new Uint8Array([0, 255, 7]).buffer
const file = () => new File([bytes], 'report.folio', { type: 'application/json' })

function handle(name = 'report.folio', events: string[] = []): LocalFileHandle {
  return {
    name,
    getFile: async () => file(),
    createWritable: async () => ({
      write: async (written) => { events.push(`write:${Array.from(new Uint8Array(written)).join(',')}`) },
      close: async () => { events.push('close') },
    }),
  }
}

describe('local file access boundary', () => {
  it('selects exactly one capability tier from complete picker capabilities and falls back for incomplete APIs', () => {
    const url = { createObjectURL: vi.fn(() => 'blob:local'), revokeObjectURL: vi.fn() }
    const complete = { document, url, showOpenFilePicker: vi.fn(), showSaveFilePicker: vi.fn() }
    expect(selectFileAccess(complete)).toBeInstanceOf(FileSystemAccess)
    expect(complete.showOpenFilePicker).toHaveBeenCalledTimes(0)
    expect(selectFileAccess({ document, url, showOpenFilePicker: vi.fn() })).toBeInstanceOf(InputDownloadAccess)
  })

  it('returns opaque selected bytes and an in-memory target from the File System Access tier', async () => {
    const selected = handle()
    const access = new FileSystemAccess({ showOpenFilePicker: vi.fn(async () => [selected]), showSaveFilePicker: vi.fn() })
    const opened = await access.open()
    expect(opened.name).toBe('report.folio')
    expect(opened.target?.handle).toBe(selected)
    expect(new Uint8Array(opened.bytes)).toEqual(new Uint8Array(bytes))
  })

  it('silently classifies picker aborts and writes then closes before reporting an in-place save', async () => {
    const events: string[] = []
    const selected = handle('saved.folio', events)
    const picker = { showOpenFilePicker: vi.fn(async () => { throw new DOMException('cancel', 'AbortError') }), showSaveFilePicker: vi.fn(async () => selected) }
    const access = new FileSystemAccess(picker)
    await expect(access.open()).rejects.toBeInstanceOf(FileAccessCancelled)
    const target = await access.acquireSaveTarget({ suggestedName: 'ignored.folio', currentTarget: { kind: 'in-place', name: selected.name, handle: selected }, saveAs: false })
    await expect(access.writeSave(target, { bytes })).resolves.toMatchObject({ name: 'saved.folio', target: { handle: selected } })
    expect(events).toEqual(['write:0,255,7', 'close'])
    expect(picker.showSaveFilePicker).not.toHaveBeenCalled()
  })

  it('uses a fresh File System Access target for Save As and does not report success when close fails', async () => {
    const broken: LocalFileHandle = { name: 'new.folio', getFile: async () => file(), createWritable: async () => ({ write: async () => undefined, close: async () => { throw new Error('media removed') } }) }
    const picker = { showOpenFilePicker: vi.fn(), showSaveFilePicker: vi.fn(async () => broken) }
    const access = new FileSystemAccess(picker)
    const target = await access.acquireSaveTarget({ suggestedName: 'old.folio', saveAs: true })
    await expect(access.writeSave(target, { bytes })).rejects.toThrow('Could not save local file')
    expect(picker.showSaveFilePicker).toHaveBeenCalledWith(expect.objectContaining({ suggestedName: 'old.folio' }))
  })

  it('keeps each denied, stale, write, and close failure as a failed local save', async () => {
    const failedHandles: LocalFileHandle[] = [
      { name: 'denied.folio', getFile: async () => file(), createWritable: async () => { throw new Error('denied') } },
      { name: 'stale.folio', getFile: async () => file(), createWritable: async () => ({ write: async () => { throw new Error('stale') }, close: async () => undefined }) },
      { name: 'close.folio', getFile: async () => file(), createWritable: async () => ({ write: async () => undefined, close: async () => { throw new Error('close') } }) },
    ]
    for (const failed of failedHandles) {
      const access = new FileSystemAccess({ showOpenFilePicker: vi.fn(), showSaveFilePicker: vi.fn() })
      const target = await access.acquireSaveTarget({ suggestedName: 'report.folio', currentTarget: { kind: 'in-place', name: failed.name, handle: failed }, saveAs: false })
      await expect(access.writeSave(target, { bytes })).rejects.toThrow('Could not save local file')
    }
  })

  it('downloads exact opaque bytes without inventing an overwrite target', async () => {
    const anchor = { href: '', download: '', style: { display: '' }, click: vi.fn(), remove: vi.fn() }
    const fakeDocument = { body: { append: vi.fn() }, createElement: vi.fn(() => anchor) } as unknown as Document
    let captured: Blob | undefined
    const url = { createObjectURL: vi.fn((blob: Blob) => { captured = blob; return 'blob:local' }), revokeObjectURL: vi.fn() }
    const access = new InputDownloadAccess(fakeDocument, url)
    const target = await access.acquireSaveTarget({ suggestedName: 'report', saveAs: false })
    await expect(access.writeSave(target, { bytes })).resolves.toEqual({ name: 'report.folio' })
    expect(anchor.download).toBe('report.folio')
    expect(anchor.click).toHaveBeenCalledOnce()
    expect(new Uint8Array(await captured!.arrayBuffer())).toEqual(new Uint8Array(bytes))
  })

  it('uses an accept-filtered input and accepts a same-file selection as a fresh opaque open', async () => {
    const access = new InputDownloadAccess()
    const opening = access.open()
    const input = document.body.querySelector<HTMLInputElement>('input[type="file"]')!
    expect(input.accept).toContain('.folio')
    Object.defineProperty(input, 'files', { configurable: true, value: { item: () => file() } })
    input.dispatchEvent(new Event('change'))
    await expect(opening).resolves.toMatchObject({ name: 'report.folio' })
    expect(document.body.querySelector('input[type="file"]')).toBeNull()
  })

  it('acquires a native save picker before any serialization/write work', async () => {
    const events: string[] = []
    const picker = { showOpenFilePicker: vi.fn(), showSaveFilePicker: vi.fn(async () => { events.push('picker'); return handle('picked.folio', events) }) }
    const access = new FileSystemAccess(picker)
    const target = await access.acquireSaveTarget({ suggestedName: 'picked', saveAs: true })
    await access.writeSave(target, { bytes })
    expect(events).toEqual(['picker', 'write:0,255,7', 'close'])
  })

  it('revokes the object URL and removes the anchor after every post-creation failure', async () => {
    const anchor = { href: '', download: '', style: { display: '' }, click: vi.fn(() => { throw new Error('blocked') }), remove: vi.fn() }
    const fakeDocument = { body: { append: vi.fn() }, createElement: vi.fn(() => anchor) } as unknown as Document
    const url = { createObjectURL: vi.fn(() => 'blob:local'), revokeObjectURL: vi.fn() }
    const access = new InputDownloadAccess(fakeDocument, url)
    const target = await access.acquireSaveTarget({ suggestedName: 'report', saveAs: false })
    await expect(access.writeSave(target, { bytes })).rejects.toThrow('Could not download local file')
    await Promise.resolve()
    expect(anchor.remove).toHaveBeenCalledOnce()
    expect(url.revokeObjectURL).toHaveBeenCalledWith('blob:local')
  })
})
