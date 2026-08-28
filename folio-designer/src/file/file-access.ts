// The file boundary carries opaque bytes and honest local-file metadata only.
// It deliberately has no knowledge of the .folio document structure.
export type FileTarget = Readonly<{
  kind: 'in-place'
  name: string
  handle: LocalFileHandle
}>

export type LocalFileHandle = Readonly<{
  name: string
  getFile(): Promise<File>
  createWritable(): Promise<WritableFile>
}>

export type WritableFile = Readonly<{
  write(bytes: ArrayBuffer): Promise<void>
  close(): Promise<void>
}>

export type LocalFile = Readonly<{
  bytes: ArrayBuffer
  name: string
  target?: FileTarget
}>

export type SavedLocalFile = Readonly<{
  name: string
  target?: FileTarget
}>

export type SaveRequest = Readonly<{
  bytes: ArrayBuffer
}>

export type SaveTargetRequest = Readonly<{
  suggestedName: string
  currentTarget?: FileTarget
  saveAs: boolean
}>

// Target acquisition is deliberately separate from writing. Chromium's save
// picker needs the click's transient user activation; serialization may take
// long enough that waiting for it first would lose that activation.
export type AcquiredSaveTarget = Readonly<{
  name: string
  target?: FileTarget
}>

export interface FileAccess {
  open(): Promise<LocalFile>
  acquireSaveTarget(request: SaveTargetRequest): Promise<AcquiredSaveTarget>
  writeSave(target: AcquiredSaveTarget, request: SaveRequest): Promise<SavedLocalFile>
}

export class FileAccessCancelled extends Error {
  constructor() { super('Local file selection was cancelled') }
}

export class FileAccessFailure extends Error {
  constructor(message: string) { super(message) }
}

export function isFileAccessCancelled(error: unknown): error is FileAccessCancelled {
  return error instanceof FileAccessCancelled
}

export function folioName(name: string): string {
  const base = name.trim().replace(/[\\/]/g, '') || 'untitled'
  return base.toLowerCase().endsWith('.folio') ? base : `${base}.folio`
}
