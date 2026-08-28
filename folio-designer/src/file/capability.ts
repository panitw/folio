import { type FileAccess } from './file-access'
import { FileSystemAccess, type FileSystemPicker } from './file-system-access'
import { InputDownloadAccess } from './input-download'
import type { DownloadUrl } from './input-download'

export type FileAccessBrowser = Partial<FileSystemPicker> & Readonly<{ document: Document; url: DownloadUrl }>

function currentBrowser(): FileAccessBrowser {
  const pickerWindow = window as typeof window & Partial<FileSystemPicker>
  return { document: window.document, url: URL, showOpenFilePicker: pickerWindow.showOpenFilePicker, showSaveFilePicker: pickerWindow.showSaveFilePicker }
}

// Called once from the composition root. The application receives only FileAccess.
export function selectFileAccess(browser: FileAccessBrowser = currentBrowser()): FileAccess {
  if (typeof browser.showOpenFilePicker === 'function' && typeof browser.showSaveFilePicker === 'function') {
    return new FileSystemAccess({ showOpenFilePicker: browser.showOpenFilePicker.bind(browser), showSaveFilePicker: browser.showSaveFilePicker.bind(browser) })
  }
  return new InputDownloadAccess(browser.document, browser.url)
}
