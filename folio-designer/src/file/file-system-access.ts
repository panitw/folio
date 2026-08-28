import { FileAccessCancelled, FileAccessFailure, folioName, type AcquiredSaveTarget, type FileAccess, type LocalFile, type LocalFileHandle, type SaveRequest, type SaveTargetRequest, type SavedLocalFile } from './file-access'

export type FileSystemPicker = Readonly<{
  showOpenFilePicker(options: OpenPickerOptions): Promise<ReadonlyArray<LocalFileHandle>>
  showSaveFilePicker(options: SavePickerOptions): Promise<LocalFileHandle>
}>

type PickerType = Readonly<{ description: string; accept: Readonly<Record<string, ReadonlyArray<string>>> }>
type OpenPickerOptions = Readonly<{ multiple: false; types: ReadonlyArray<PickerType> }>
type SavePickerOptions = Readonly<{ suggestedName: string; types: ReadonlyArray<PickerType> }>

const folioPickerType: PickerType = { description: 'Folio template', accept: { 'application/json': ['.folio'] } }

export class FileSystemAccess implements FileAccess {
  private readonly picker: FileSystemPicker

  constructor(picker: FileSystemPicker) { this.picker = picker }

  async open(): Promise<LocalFile> {
    try {
      const [handle] = await this.picker.showOpenFilePicker({ multiple: false, types: [folioPickerType] })
      if (!handle) throw new FileAccessCancelled()
      const file = await handle.getFile()
      return { bytes: await file.arrayBuffer(), name: file.name, target: { kind: 'in-place', name: file.name, handle } }
    } catch (error) {
      throw localFailure(error, 'Could not open local file')
    }
  }

  async acquireSaveTarget(request: SaveTargetRequest): Promise<AcquiredSaveTarget> {
    try {
      if (!request.saveAs && request.currentTarget?.kind === 'in-place') {
        return { name: request.currentTarget.name, target: request.currentTarget }
      }
      const handle = await this.picker.showSaveFilePicker({ suggestedName: folioName(request.suggestedName), types: [folioPickerType] })
      return { name: handle.name, target: { kind: 'in-place', name: handle.name, handle } }
    } catch (error) {
      throw localFailure(error, 'Could not prepare local save')
    }
  }

  async writeSave(acquired: AcquiredSaveTarget, request: SaveRequest): Promise<SavedLocalFile> {
    if (!acquired.target) throw new FileAccessFailure('Could not save local file')
    try {
      const target = acquired.target
      const writable = await target.handle.createWritable()
      await writable.write(request.bytes)
      await writable.close()
      return { name: target.name, target }
    } catch (error) {
      throw localFailure(error, 'Could not save local file')
    }
  }
}

function localFailure(error: unknown, fallback: string): Error {
  if (error instanceof FileAccessCancelled) return error
  if (error instanceof DOMException && error.name === 'AbortError') return new FileAccessCancelled()
  return new FileAccessFailure(fallback)
}
