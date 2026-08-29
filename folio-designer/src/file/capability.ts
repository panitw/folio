import { type FileAccess } from './file-access'
import { FileSystemAccess, type FileSystemPicker } from './file-system-access'
import { InputDownloadAccess } from './input-download'
import type { DownloadUrl } from './input-download'
import { FileSystemSampleAccess, InputSampleAccess, type SampleFileAccess, type SamplePicker } from '../sample-file'

export type FileAccessBrowser = Partial<FileSystemPicker> & Readonly<{ document: Document; url: DownloadUrl }>

function currentBrowser(): FileAccessBrowser {
  const pickerWindow = window as typeof window & Partial<FileSystemPicker>
  // The pickers are Window methods and the browser brand-checks their receiver.
  // Bind them to the window here, where it is still in hand; carrying the bare
  // references out on this record would strand them on a plain object.
  return { document: window.document, url: URL, showOpenFilePicker: pickerWindow.showOpenFilePicker?.bind(pickerWindow), showSaveFilePicker: pickerWindow.showSaveFilePicker?.bind(pickerWindow) }
}

// Called once from the composition root. The application receives only FileAccess.
export function selectFileAccess(browser: FileAccessBrowser = currentBrowser()): FileAccess {
  if (typeof browser.showOpenFilePicker === 'function' && typeof browser.showSaveFilePicker === 'function') {
    return new FileSystemAccess({ showOpenFilePicker: browser.showOpenFilePicker.bind(browser), showSaveFilePicker: browser.showSaveFilePicker.bind(browser) })
  }
  return new InputDownloadAccess(browser.document, browser.url)
}

// Sample selection shares the same single capability decision as template
// selection, but exposes no template target, save, or document semantics.
export function selectSampleFileAccess(browser: FileAccessBrowser = currentBrowser()): SampleFileAccess {
  if (typeof browser.showOpenFilePicker === 'function') return new FileSystemSampleAccess({ showOpenFilePicker: browser.showOpenFilePicker.bind(browser) } as SamplePicker)
  return new InputSampleAccess(browser.document)
}
