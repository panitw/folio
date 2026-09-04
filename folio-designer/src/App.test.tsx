import { createEvent, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import App, { placementPoint } from './App'
import { shortcutHintsFor } from './shortcuts'
import { embeddedFaceFamily } from './embedded-face-family'
import { FileAccessCancelled, type FileAccess } from './file/file-access'
import type { EngineClient } from './engine-client'
import type { CanvasProjection } from './engine-protocol'
import { acceptSampleData } from './sample-data'
import { MAX_CANVAS_SHEETS } from './sheet-stack'
import { catalogueFaces } from './generated/font-catalogue'
import { IDBFactory as FakeIndexedDBFactory } from 'fake-indexeddb'

// STORY 16.5 — SOME OF THESE TESTS NEED A MACHINE THAT CAN KEEP A FACE.
//
// Installing IS the store write: there is no command behind it to succeed, so
// an install into a browser that will not keep anything is a REFUSED install
// (`App.font-store.test.tsx` asserts that refusal in its own right). jsdom 28.1.0
// provides no IndexedDB at all, so the web-tier tests below install a FRESH fake
// factory for their own duration and put back whatever was there. It is scoped
// per test rather than to the file, because the other hundred-odd tests here are
// about a designer with no font store and must stay that way.
const withMachineStore = (): (() => void) => {
  const previous = Object.getOwnPropertyDescriptor(globalThis, 'indexedDB')
  Object.defineProperty(globalThis, 'indexedDB', { value: new FakeIndexedDBFactory(), configurable: true, writable: true })
  return () => {
    if (previous) Object.defineProperty(globalThis, 'indexedDB', previous)
    else Reflect.deleteProperty(globalThis, 'indexedDB')
  }
}

// face() builds the PROJECTED shape of a named-face chain entry (Story 8.3:
// an entry is a discriminated object, not a string). A named face carries no
// family and no style — its name is its identity.
const face = (name: string) => ({ face: name, assetKey: '', family: '', style: '' })

// carried() is the projected shape of an EMBEDDED chain entry: no face name,
// an asset key, and the family/style Go read out of the asset's own `font`
// record for the panel to display. The key never becomes a family here — that
// derivation is embedded-face-family.ts's alone (D-8.4.1).
const carried = (assetKey: string) => ({ face: '', assetKey, family: 'Noto Sans Thai', style: 'Regular' })

// installStubFontSet installs the page font set jsdom does not implement and
// returns its own removal. `Object.defineProperty` because neither the face
// constructor nor the set exists to be assigned over.
//
// IT RECORDS WHAT THE SEAM DID TO IT, in order: the families added and the
// families removed. Registration and release are otherwise invisible — nothing
// in the DOM says a face was released — so a claim about the seam's LIFETIME
// can only be made against this record.
function installStubFontSet(): Readonly<{ restore: () => void; added: string[]; removed: string[] }> {
  class StubFace {
    readonly family: string
    constructor(family: string) { this.family = family }
    load(): Promise<StubFace> { return Promise.resolve(this) }
  }
  const added: string[] = []
  const removed: string[] = []
  const set = { add: (face: StubFace) => { added.push(face.family); return undefined }, delete: (face: StubFace) => { removed.push(face.family); return undefined } }
  Object.defineProperty(globalThis, 'FontFace', { value: StubFace, configurable: true, writable: true })
  Object.defineProperty(document, 'fonts', { value: set, configurable: true, writable: true })
  return { restore: () => { Reflect.deleteProperty(globalThis, 'FontFace'); Reflect.deleteProperty(document, 'fonts') }, added, removed }
}

// TWO text components drawing through ONE carried entry, so a per-component
// registration lifetime shows up as two asset requests instead of one.
const carriedFaceCanvas = (key: string) => {
  const paint = { overflow: false, truncated: false, lines: [{ top: 0, baseline: 12_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] }
  const component = (id: string, y: number) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height: 24_000, resizable: true, value: 'ignored', textPaint: paint })
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(key)] }], components: [component('e1', 0), component('e2', 30_000)] }
}

// TWO carried entries, ONE PER COMPONENT, so a claim about one key can be read
// off the DOM against a SETTLED outcome for the other. The fragment order in
// the container is the component order: `first` then `second`.
const twoCarriedFacesCanvas = (first: string, second: string) => {
  const paint = (key: string) => ({ overflow: false, truncated: false, lines: [{ top: 0, baseline: 12_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] })
  const component = (id: string, y: number, key: string) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height: 24_000, resizable: true, value: 'ignored', textPaint: paint(key) })
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(first), carried(second)] }], components: [component('e1', 0, first), component('e2', 30_000, second)] }
}

// ONE carried component that CROSSES A WINDOW SEAM, so the projection produces
// a home occurrence AND an echo of it on the next sheet. Every content window
// after the first is drawn by echoes, so a multi-sheet document is the ordinary
// case rather than an exotic one.
const carriedFaceEchoCanvas = (key: string) => {
  const paint = { overflow: false, truncated: false, lines: [{ top: 650_000, baseline: 662_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] }
  const spanning = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 650_000, width: 72_000, height: 100_000, resizable: true, value: 'ignored', textPaint: paint }
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans'), carried(key)] }], contentWindowCount: 3, contentWindowOrigins: [0, 700_000, 1_400_000], components: [spanning] }
}

// ONE SHIPPED-FACE component that CROSSES A WINDOW SEAM, drawing LATIN text
// through a chain that names only "Noto Sans Thai" — the I/O matrix's
// "Latin through a Thai-first chain" row, projected. The engine attributes the
// fragment to the face it measured with; nothing here carries a font, so the
// only identity on the wire is the shipped one.
const shippedFaceEchoCanvas = (name: string) => {
  const paint = { overflow: false, truncated: false, lines: [{ top: 650_000, baseline: 662_000, advance: 16_000, width: 24_000, fragments: [{ text: 'A5', x: 0, face: name }] }] }
  const spanning = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 650_000, width: 72_000, height: 100_000, resizable: true, value: 'ignored', textPaint: paint }
  return { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face(name)] }], contentWindowCount: 3, contentWindowOrigins: [0, 700_000, 1_400_000], components: [spanning] }
}

// The family sequence a rendered fragment asks for, quotes removed: jsdom
// re-spells single quotes as double ones when a declaration is read back, and
// the claim is about WHICH families are asked for and in what ORDER.
const familiesAskedFor = (node: HTMLElement) => node.style.fontFamily === '' ? [] : node.style.fontFamily.split(',').map((entry) => entry.trim().replace(/^['"]|['"]$/g, ''))

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => false,
  PDFPreviewViewer: ({ label, describedBy, onPageCount, onError }: { label: string; describedBy: string; onPageCount: (pages: number) => void; onError: (error: Error) => void }) => <><button type="button" aria-label={label} aria-describedby={describedBy} onClick={() => onPageCount(1)}>Admit local PDF</button><button type="button" aria-label="Fail local PDF viewer" onClick={() => onError(new Error('viewer rejected bytes'))}>Fail local PDF viewer</button></>,
}))

const bytes = new Uint8Array([1, 2, 3]).buffer
// The family combobox lists TWO groups since Story 8.6 — the chains the
// document declares, and the bundled catalogue it does not. Only a catalogue
// entry carries a source note, so that is what separates them here: the note is
// what the entry DOES, not how it looks, and a control that stopped
// distinguishing the two would fail this rather than restyle past it.
//
// MECHANICAL (Story 16.5): the note used to be the literal `add to document` on
// every catalogue arm. It now says which of two things a pick does — INSTALL a
// family this machine does not hold, or USE one it does — and the one phrase all
// three arms still share is the machine. `font-index.test.ts` is where the three
// sentences themselves are pinned; this only needs the partition.
// STORY 16.7 GAVE EVERY OPTION AN ARIA-HIDDEN SPECIMEN, so raw `textContent`
// now carries a sample the row's accessible name deliberately excludes (the
// honesty rule requires it: the specimen is decorative to assistive
// technology). `optionText` reads what a screen reader would — the row's name
// and its note, if it still has one — by dropping every `aria-hidden`
// descendant before reading text, so these assertions stay about what the row
// SAYS rather than about whether its specimen happened to finish loading.
const optionText = (option: Element): string => {
  const clone = option.cloneNode(true) as Element
  clone.querySelectorAll('[aria-hidden="true"]').forEach((node) => node.remove())
  return clone.textContent ?? ''
}
// AND `declaredOptions` NOW SCOPES BY GROUP MEMBERSHIP RATHER THAN BY TEXT.
// The "this machine" substring used to be true of every non-declared row and
// false of every declared one; Story 16.7 (D-16.R.72, narrowed) removed the
// per-row note from `AVAILABLE LOCALLY` too, so that heuristic would now let a
// local-tier row through as though it were declared. `IN THIS TEMPLATE` is the
// group's own accessible name and the one fact this control never overloads.
const declaredOptions = () => {
  const listbox = screen.getByRole('listbox', { name: 'Fonts' })
  const group = within(listbox).queryByRole('group', { name: 'IN THIS TEMPLATE' })
  return group ? within(group).queryAllByRole('option') : []
}
const sample = acceptSampleData('sample.json', new TextEncoder().encode('{"customer":{"name":"Preview customer"},"transactions":[]}').buffer)
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body', 'heading'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans Thai')] }], defaultFontSize: 12000, contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = (revision: number) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas })
const engine = (request = vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3 }, ...(operation === 'serialize' ? { bytes } : {}) }))) => ({ request }) as unknown as EngineClient

describe('application shell', () => {
  it('hydrates engine-owned state when asynchronous startup replaces the loading shell', () => {
    const lifecycle = { state: 'ready' as const, cacheReady: true, verifiedAssetUrls: [] }
    const view = render(<App key="engine-loading" loadState={lifecycle} engineState="starting" />)
    expect(screen.getByRole('status', { name: 'Engine preparation status' })).toHaveTextContent('Starting local engine')
    view.rerender(<App key="engine-ready" engine={engine()} initialSnapshot={snapshot(1)} loadState={lifecycle} engineState="starting" />)
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 1')
    expect(screen.getByLabelText('Report page with Page Header, Content, and Page Footer')).toBeInTheDocument()
  })

  it('renders every persistent desktop landmark and honest later regions', () => {
    render(<App initialSnapshot={snapshot(1)} />)
    expect(screen.getByLabelText('Document bar')).toBeInTheDocument()
    expect(screen.getByLabelText('Component palette')).toBeInTheDocument()
    expect(screen.getByLabelText('Canvas region')).toBeInTheDocument()
    expect(screen.getByLabelText('Report page with Page Header, Content, and Page Footer')).toBeInTheDocument()
    expect(screen.getByLabelText('Properties panel')).toBeInTheDocument()
    expect(screen.getByLabelText('Status bar')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'PREVIEW' })).toHaveAttribute('aria-pressed', 'false')
  })

  // STORY 16.4 — THE STATUS BAR STATES THE FONT COUNT, AND IT IS THE SAME COUNT
  // THE DROPDOWN'S FIRST GROUP DRAWS.
  //
  // The two surfaces are asserted AGAINST EACH OTHER rather than each against a
  // literal, because the property is that they teach one model from one source:
  // `canvas.fontFamilies`, which is `IN THIS TEMPLATE`'s own predicate. A count
  // read from anywhere else — the fonts added this session, say, which is what
  // the mockup binds — would pass a two-literal test and fail this one.
  it('states the template font count in the status bar, from the source the first group groups on', () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const bar = screen.getByLabelText('Status bar')
    expect(within(bar).getByTestId('template-font-count')).toHaveTextContent('2 fonts in template')
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    expect(within(screen.getByRole('group', { name: 'IN THIS TEMPLATE' })).getAllByRole('option'), 'the bar and the first group must be counting the same thing').toHaveLength(2)
  })

  // AND IT AGREES WITH ITSELF AT ONE. A hardcoded "N fonts" reads "1 fonts",
  // which is the kind of small lie a status bar makes for ever.
  it('says one font in the singular', () => {
    const oneChain = { ...canvas, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: oneChain }} />)
    expect(screen.getByTestId('template-font-count')).toHaveTextContent('1 font in template')
  })

  it('opens an engine-projected, keyboard-operable table matrix with named controls', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const tableSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'table-columns') return { snapshot: tableSnapshot, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'right' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } }
      return { snapshot: tableSnapshot }
    })
    render(<App engine={engine(request)} initialSnapshot={tableSnapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
    const grid = await screen.findByRole('grid', { name: 'Table columns' })
    expect(grid).toHaveAttribute('aria-colcount', '11')
		expect(grid).toHaveAttribute('aria-rowcount', '2')
    const header = screen.getByRole('textbox', { name: 'Header for column 1' })
    header.focus(); fireEvent.keyDown(header, { key: 'ArrowRight' })
    expect(document.activeElement).toBe(screen.getByRole('spinbutton', { name: 'Width for column 1 in points' }))
    expect(screen.getByRole('button', { name: 'Move column 1 earlier' })).toBeDisabled()
		expect(screen.getByRole('button', { name: 'Move column 1 later' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Remove column 1' })).toBeInTheDocument()
  })

  it('traps the focused matrix, closes on Escape, and restores its invoking control', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const tableSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    render(<App engine={engine(vi.fn(async (operation: string) => operation === 'table-columns' ? { snapshot: tableSnapshot, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } } : { snapshot: tableSnapshot }))} initialSnapshot={tableSnapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    const invoker = screen.getByRole('button', { name: 'Configure columns' })
    invoker.focus(); fireEvent.click(invoker)
    const header = await screen.findByRole('textbox', { name: 'Header for column 1' })
    expect(document.activeElement).toBe(header)
    fireEvent.keyDown(header, { key: 'Tab' })
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Close Table Editor' }))
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).not.toBeInTheDocument())
    expect(document.activeElement).toBe(invoker)
  })

  it('admits a committed table snapshot after deselection and never reopens a closed session', async () => {
    const tableCanvas = { ...canvas, components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: false }] }
    const first = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: tableCanvas }
    const second = { documentState: 'loaded' as const, revision: 2, byteLength: 4, canvas: { ...tableCanvas, components: [{ ...tableCanvas.components[0]!, width: 144000 }] } }
    let releaseProjection!: () => void
    const delayedProjection = new Promise<{ snapshot: typeof second; tableColumns: { revision: number; table: { tableId: string; collection: string; alias: string; columns: { id: string; header: string; width: number; align: 'left'; binding: string; rowField: string; rowFieldEditable: boolean; footer: ''; footerOf: string; footerFormat: string }[] } } }>((resolve) => { releaseProjection = () => resolve({ snapshot: second, tableColumns: { revision: 2, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left', binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '', footerOf: '', footerFormat: '' }] } } }) })
    let queries = 0
    const request = vi.fn((operation: string) => {
      if (operation === 'table-columns') { queries++; return queries === 1 ? Promise.resolve({ snapshot: first, tableColumns: { revision: 1, table: { tableId: 'e7', collection: 'items[]', alias: 'row', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'left' as const, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '' as const, footerOf: '', footerFormat: '' }] } } }) : delayedProjection }
      if (operation === 'command') return Promise.resolve({ snapshot: second })
      return Promise.resolve({ snapshot: first })
    })
    render(<App engine={engine(request)} initialSnapshot={first} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
    await screen.findByRole('button', { name: 'Add column after column 1' })
    fireEvent.click(screen.getByRole('button', { name: 'Add column after column 1' }))
    fireEvent.click(screen.getByRole('button', { name: 'Close Table Editor' }))
    releaseProjection()
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 2'))
    expect(screen.queryByRole('dialog', { name: 'Table Editor' })).not.toBeInTheDocument()
  })

  it('replaces the canvas with Preview, cancels an older render, and never dirties or installs its late PDF', async () => {
    let releaseSerialize!: (value: { snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }) => void
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } })
      if (operation === 'serialize') return new Promise<{ snapshot: ReturnType<typeof snapshot>; bytes: ArrayBuffer }>((resolve) => { releaseSerialize = resolve })
      if (operation === 'render') return Promise.resolve({ snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    expect(screen.queryByLabelText('Canvas region')).not.toBeInTheDocument()
    expect(screen.getByText('Rendering local PDF')).toBeInTheDocument()
    await waitFor(() => expect(request).toHaveBeenCalledWith('serialize', undefined, expect.any(AbortSignal)))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel and return to Design' }))
    releaseSerialize({ snapshot: snapshot(1), bytes })
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.queryByText(/Go production digest/)).not.toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(request.mock.calls.map(([operation]) => operation)).toEqual(['parameter-references', 'identity', 'serialize'])
  })

  it('coalesces manual and debounced rerenders behind one active FIFO operation', async () => {
    let releaseIdentity!: () => void
    let identityCalls = 0
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') {
        identityCalls++
        if (identityCalls === 1) return new Promise<{ snapshot: ReturnType<typeof snapshot>; preview: { revision: number; identity: string } }>((resolve) => { releaseIdentity = () => resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }) })
        return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'c'.repeat(64) } })
      }
      if (operation === 'serialize') return Promise.resolve({ snapshot: snapshot(1), bytes })
      if (operation === 'render') return Promise.resolve({ snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'c'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    vi.useFakeTimers()
    try {
      render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
      fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[1]}' } })
      fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[2]}' } })
      fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
      await vi.runAllTimersAsync()
      expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(1)
      releaseIdentity()
      await vi.runAllTimersAsync()
      await Promise.resolve(); await Promise.resolve(); await Promise.resolve()
      expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('uses the engine reference projection for keyboard-operable parameter inputs and retains accepted bytes through an invalid draft', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') return { snapshot: snapshot(1), parameterReferences: { revision: 1, names: ['reportDate'] } }
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const named = await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.change(named, { target: { value: '"2026-08-28T00:00:00Z"' } })
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(2))
    const accepted = request.mock.calls.filter(([operation]) => operation === 'identity').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    expect(new TextDecoder().decode(accepted[1].params)).toContain('2026-08-28T00:00:00Z')
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{ nope' } })
    expect(screen.getByRole('alert')).toHaveTextContent('last accepted parameter document remains in Preview')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
  })

  it('states pending, failed, and empty parameter discovery without inventing fields', async () => {
    let release!: () => void
    const pending = new Promise<{ snapshot: ReturnType<typeof snapshot>; parameterReferences: { revision: number; names: string[] } }>((resolve) => { release = () => resolve({ snapshot: snapshot(1), parameterReferences: { revision: 1, names: [] } }) })
    const request = vi.fn((operation: string) => operation === 'parameter-references' ? pending : Promise.resolve(operation === 'identity' ? { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } } : { snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    expect(screen.getByText('Discovering parameter references from the local engine…')).toBeInTheDocument()
    release()
    await screen.findByText('The local engine found no parameter references in this template.')
  })

  it('states a failed parameter projection rather than calling it an empty projection', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') throw new Error('worker unavailable')
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await screen.findByText('The local engine could not provide parameter references. The raw parameter document is still available.')
    expect(screen.queryByText('The local engine found no parameter references in this template.')).not.toBeInTheDocument()
  })

  it('edits named parameters without rewriting raw numeric lexemes or special own keys', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') return { snapshot: snapshot(1), parameterReferences: { revision: 1, names: ['__proto__', 'constructor', 'reportDate'] } }
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const reportDate = await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"constructor":1.00e+2,"__proto__":-0,"other":123.4500}' } })
    expect(screen.getByRole('textbox', { name: 'Value for params.constructor' })).toHaveValue('1.00e+2')
    expect(screen.getByRole('textbox', { name: 'Value for params.__proto__' })).toHaveValue('-0')
    reportDate.focus()
    fireEvent.change(reportDate, { target: { value: '"2026-08-28T00:00:00Z"' } })
    expect(document.activeElement).toBe(reportDate)
    fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'identity').length).toBeGreaterThan(1))
    const accepted = request.mock.calls.filter(([operation]) => operation === 'identity').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    const exact = '{"constructor":1.00e+2,"__proto__":-0,"other":123.4500,"reportDate":"2026-08-28T00:00:00Z"}'
    expect(new TextDecoder().decode(accepted[1].params)).toBe(exact)
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render').length).toBeGreaterThan(0))
    const rendered = request.mock.calls.filter(([operation]) => operation === 'render').at(-1)! as unknown as [string, { params: ArrayBuffer }]
    expect(new TextDecoder().decode(rendered[1].params)).toBe(exact)
  })

  it('refreshes the engine reference projection after Undo while Preview remains open', async () => {
    let references = 0
    const historySnapshot = { ...snapshot(2), canUndo: false, canRedo: true }
    const initial = { ...snapshot(1), canUndo: true, canRedo: false }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'parameter-references') {
        references++
        return { snapshot: references === 1 ? initial : historySnapshot, parameterReferences: { revision: references === 1 ? 1 : 2, names: references === 1 ? ['reportDate'] : ['branch'] } }
      }
      if (operation === 'undo') return { snapshot: historySnapshot }
      if (operation === 'identity') return { snapshot: references > 1 ? historySnapshot : initial, preview: { revision: references > 1 ? 2 : 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: references > 1 ? historySnapshot : initial, bytes }
      if (operation === 'render') return { snapshot: references > 1 ? historySnapshot : initial, bytes: new Uint8Array([9]).buffer, preview: { revision: references > 1 ? 2 : 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: initial }
    })
    render(<App engine={engine(request)} initialSnapshot={initial} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await screen.findByRole('textbox', { name: 'Value for params.reportDate' })
    fireEvent.click(screen.getByRole('button', { name: /^Undo/ }))
    await screen.findByRole('textbox', { name: 'Value for params.branch' })
    expect(screen.queryByRole('textbox', { name: 'Value for params.reportDate' })).not.toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'parameter-references')).toHaveLength(2)
  })

  it('waits for matching PDF.js admission before claiming current exact output', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    expect(screen.getByText('LOCAL PDF PREVIEW')).toBeInTheDocument()
    expect(screen.queryByText('EXACT LOCAL PRODUCTION PDF')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    expect(screen.getByText('EXACT LOCAL PRODUCTION PDF')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Current exact local production PDF/ })).toHaveAttribute('aria-describedby', 'preview-freshness-status')
  })

  it('keeps producer diagnostics hidden and inert until their exact PDF is admitted, then revokes them on input change', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [{ severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'gone', dataPath: 'bands.content.gone', message: 'Content was clipped' }] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByLabelText('Render diagnostics')).toBeInTheDocument())
    fireEvent.change(screen.getByRole('textbox', { name: 'Raw parameter JSON' }), { target: { value: '{"transactions":[1]}' } })
    expect(screen.queryByLabelText('Render diagnostics')).not.toBeInTheDocument()
  })

  it('returns to Design and announces an unavailable authoritative warning target without selecting a substitute', async () => {
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [{ severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'gone', dataPath: '', message: 'Content was clipped' }] } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: /Stale historical PDF/ }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Locate in Design' })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Locate in Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.getByText('Locate unavailable: the authoritative element is no longer present.')).toHaveAttribute('role', 'status')
  })

  it('returns from a path-only render failure without requiring an element id', async () => {
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', dataPath: 'items[0]', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') throw failure
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(screen.getByLabelText('Local render failure')).toBeInTheDocument())
    fireEvent.click(within(screen.getByLabelText('Local render failure')).getByRole('button', { name: 'Return to Design' }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
  })

  it('returns from a located render failure by selecting only the current projected element', async () => {
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const })
    const locatedCanvas = { ...canvas, components: [{ id: 'e7', type: 'rect' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 12000, resizable: true }] }
    const locatedSnapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: locatedCanvas }
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: locatedSnapshot, preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: locatedSnapshot, bytes }
      if (operation === 'render') throw failure
      return { snapshot: locatedSnapshot }
    })
    render(<App engine={engine(request)} initialSnapshot={locatedSnapshot} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const card = await screen.findByLabelText('Local render failure')
    fireEvent.click(within(card).getByRole('button', { name: 'Return to Design' }))
    const component = await screen.findByRole('button', { name: 'rect component e7' })
    expect(component).toHaveClass('canvas-component-selected')
    expect(screen.getByText('Selected e7 in Design.')).toHaveAttribute('role', 'status')
  })

  it('retries an active failed render through the existing scheduler without mutating the document', async () => {
    let renders = 0
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', dataPath: 'params.reportDate', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') { renders++; if (renders === 1) throw failure; return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } } }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const card = await screen.findByLabelText('Local render failure')
    expect(card).toHaveTextContent('RENDER_INVALID')
    fireEvent.click(within(card).getByRole('button', { name: 'Retry preview' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(2))
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('forces a fresh FIFO render after a same-identity last-good PDF failure and retains that PDF as stale', async () => {
    let renders = 0
    const failure = Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'identity') return { snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } }
      if (operation === 'serialize') return { snapshot: snapshot(1), bytes }
      if (operation === 'render') {
        renders++
        if (renders > 1) throw failure
        return { snapshot: snapshot(1), bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
      }
      return { snapshot: snapshot(1) }
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    const admitted = await screen.findByRole('button', { name: /Stale historical PDF/ })
    fireEvent.click(admitted)
    await waitFor(() => expect(screen.getByRole('button', { name: /Current exact local production PDF/ })).toBeInTheDocument())
    fireEvent.click(screen.getByRole('button', { name: 'Fail local PDF viewer' }))
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(screen.getByText(/local PDF viewer could not display/i)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Render local PDF' }))
    const card = await screen.findByLabelText('Local render failure')
    expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument()
    const retry = within(card).getByRole('button', { name: 'Retry preview' })
    retry.focus()
    fireEvent.keyDown(retry, { key: 'Enter' })
    fireEvent.click(retry)
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(3))
    expect(request.mock.calls.filter(([operation]) => operation === 'identity')).toHaveLength(3)
    expect(request.mock.calls.filter(([operation]) => operation === 'serialize')).toHaveLength(3)
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
    expect(screen.getByLabelText('Local render failure')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Stale historical PDF/ })).toBeInTheDocument()
  })

  it('revokes a delayed render failure after leaving Preview so its actions cannot reach Design', async () => {
    let rejectRender!: (error: Error) => void
    const request = vi.fn((operation: string) => {
      if (operation === 'identity') return Promise.resolve({ snapshot: snapshot(1), preview: { revision: 1, identity: 'b'.repeat(64) } })
      if (operation === 'serialize') return Promise.resolve({ snapshot: snapshot(1), bytes })
      if (operation === 'render') return new Promise<never>((_, reject: (error: Error) => void) => { rejectRender = reject })
      return Promise.resolve({ snapshot: snapshot(1) })
    })
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} initialSampleData={sample} />)
    fireEvent.click(screen.getByRole('button', { name: 'PREVIEW' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'render')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel and return to Design' }))
    rejectRender(Object.assign(new Error('The template could not be processed'), { code: 'RENDER_INVALID', elementId: 'e7', producerRenderFailure: true as const }))
    await waitFor(() => expect(screen.getByLabelText('Canvas region')).toBeInTheDocument())
    expect(screen.queryByLabelText('Local render failure')).not.toBeInTheDocument()
    expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(0)
  })

  it('names local file controls, persistent unsaved state, and offline availability', () => {
    render(<App />)
    const open = screen.getByRole('button', { name: 'Open local template' })
    expect(open).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Save local template' })).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Start blank' })).toBeDisabled()
    expect(screen.getByRole('status', { name: 'Offline availability' })).toHaveTextContent('Offline cache unavailable')
  })

  it('labels the development bypass instead of claiming a verified cache', () => {
    render(<App offlineState="dev-bypass" />)
    expect(screen.getByRole('status', { name: 'Offline availability' })).toHaveTextContent('Offline layer bypassed (dev)')
  })

  it('announces the checking, ready, and waiting-update lifecycle states', () => {
    const { rerender } = render(<App offlineState="checking" />)
    const status = screen.getByRole('status', { name: 'Offline availability' })
    expect(status).toHaveAttribute('aria-live', 'polite')
    expect(status).toHaveTextContent('Offline cache checking')
    rerender(<App offlineState="ready" />)
    expect(status).toHaveTextContent('Offline ready')
    rerender(<App offlineState="update-available" />)
    expect(status).toHaveTextContent('Update available; current release remains usable')
  })

  it('bypasses S1 when the current cache and engine are already ready', () => {
    render(<App loadState={{ state: 'ready', cacheReady: true, verifiedAssetUrls: [] }} engineState="starting" />)
    expect(screen.getByRole('status', { name: 'Engine preparation status' })).toHaveTextContent('Starting local engine')
    expect(screen.queryByRole('heading', { name: 'Preparing Folio' })).not.toBeInTheDocument()
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument()
  })

  it('loads only opaque adapter bytes through Go, establishes a clean baseline, and dirties after a committed command', async () => {
    const request = vi.fn(async (operation: string) => ({ snapshot: snapshot(operation === 'command' ? 8 : 7), ...(operation === 'serialize' ? { bytes } : {}) }))
    const files: FileAccess = { open: vi.fn(async () => ({ bytes, name: 'report.folio' })), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Open local template' }))
    await waitFor(() => expect(screen.getByText('report.folio')).toBeInTheDocument())
    expect(request.mock.calls.map(([operation]) => operation)).toEqual(['load', 'serialize'])
    expect(screen.getByText('Saved local file')).toBeInTheDocument()
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByText('Unsaved local changes')).toBeInTheDocument())
  })

  it('keeps zoom, grid, and snap local while an explicit Go page-setup command alone dirties the document', async () => {
    const request = vi.fn(async (operation: string) => ({ snapshot: snapshot(operation === 'command' ? 2 : 1), ...(operation === 'serialize' ? { bytes } : {}) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
    fireEvent.click(screen.getByRole('button', { name: 'Grid on' }))
    fireEvent.click(screen.getByRole('button', { name: 'Snap on' }))
    expect(request).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Grid off' })).toHaveAttribute('aria-pressed', 'false')
    expect(screen.getByRole('button', { name: 'Snap off' })).toHaveAttribute('aria-pressed', 'false')
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37.125' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request).toHaveBeenCalledTimes(1))
    expect(request.mock.calls[0]![0]).toBe('command')
    const command = request.mock.calls[0] as unknown as [string, ArrayBuffer]
    expect(new TextDecoder().decode(command[1])).toContain('"top":37.125')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('drives the full authoritative undo/redo depth, bounds, and divergent branch from engine snapshots', async () => {
    let revision = 1
    let undoDepth = 0
    let redoDepth = 0
    const historySnapshot = () => ({ ...snapshot(revision), canUndo: undoDepth > 0, canRedo: redoDepth > 0 })
    const request = vi.fn(async (operation: string) => {
      if (operation === 'command') { revision++; undoDepth++; redoDepth = 0; return { snapshot: historySnapshot() } }
      if (operation === 'undo') { revision++; undoDepth--; redoDepth++; return { snapshot: historySnapshot() } }
      if (operation === 'redo') { revision++; undoDepth++; redoDepth--; return { snapshot: historySnapshot() } }
      return { snapshot: historySnapshot(), ...(operation === 'serialize' ? { bytes } : {}) }
    })
    render(<App engine={engine(request)} initialSnapshot={historySnapshot()} />)
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled())
    fireEvent.change(screen.getByRole('textbox', { name: 'Top margin (pt)' }), { target: { value: '38' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(2))
    fireEvent.click(screen.getByRole('button', { name: 'Undo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'undo')).toHaveLength(1))
    expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Redo' })).toBeEnabled()
    fireEvent.click(screen.getByRole('button', { name: 'Undo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'undo')).toHaveLength(2))
    expect(screen.getByRole('button', { name: 'Undo' })).toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: 'Redo' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'redo')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(3))
    expect(screen.getByRole('button', { name: 'Redo' })).toBeDisabled()
  })

  it('keeps a no-op command non-dirty and out of browser history when the engine returns its stable snapshot', async () => {
    const request = vi.fn(async () => ({ snapshot: { ...snapshot(1), canUndo: false, canRedo: true } }))
    render(<App engine={engine(request)} initialSnapshot={{ ...snapshot(1), canUndo: false, canRedo: true }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(screen.getByRole('button', { name: 'Undo' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Redo' })).toBeEnabled()
  })

  it.each([
    [true, { save: '⌘S', undo: '⌘Z', redo: '⇧⌘Z', preview: '⌥P', snap: '⌥S' }],
    [false, { save: 'Ctrl+S', undo: 'Ctrl+Z', redo: 'Ctrl+Y', preview: 'Alt+P', snap: 'Alt+S' }],
  ])('uses one platform-normalized shortcut map (%s)', (mac, expected) => {
    expect(shortcutHintsFor(mac)).toMatchObject(expected)
  })

  it.each([
    ['property draft', { key: 'z', ctrlKey: true }],
    ['IME composition', { key: 'z', ctrlKey: true, isComposing: true }],
  ])('does not route Undo through an editable %s', (_name, keyboard) => {
    const request = vi.fn(async () => ({ snapshot: { ...snapshot(1), canUndo: false, canRedo: true } }))
    render(<App engine={engine(request)} initialSnapshot={{ ...snapshot(1), canUndo: true }} />)
    fireEvent.keyDown(screen.getByRole('textbox', { name: 'Top margin (pt)' }), keyboard)
    expect(request).not.toHaveBeenCalled()
  })

  it('offers only the five fixed palette components and sends an opaque Go placement command', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    expect(screen.getAllByRole('button', { name: /Place / }).map((button) => button.getAttribute('aria-label'))).toEqual(['Place Text', 'Place Image', 'Place Table', 'Place Line', 'Place Rectangle'])
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    const [operation, payload] = request.mock.calls[0] as unknown as [string, ArrayBuffer]
    expect(operation).toBe('command')
    expect(new TextDecoder().decode(payload)).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })

  it('converts a local band pointer position through the shared display mapping before proposing placement', () => {
    const localX = ['offset', 'X'].join('')
    const localY = ['offset', 'Y'].join('')
    expect(placementPoint({ [localX]: 120, [localY]: 40 } as unknown as MouseEvent, canvas.bands[1]!, 1)).toEqual({ x: 156, y: 96 })
  })

  it('keeps selection local and deletes one unambiguous selected component through Go', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e9'))
    expect(request).not.toHaveBeenCalled()
    const region = screen.getByLabelText('Canvas region')
    region.focus()
    fireEvent.keyDown(region, { key: 'Delete' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"deleteComponent","version":1,"id":"e9"}')
  })

  it('does not send a move for a pointer selection, but commits one point-valued move after a real drag', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 10, clientY: 10 })
    expect(request).not.toHaveBeenCalled()
    fireEvent.pointerDown(component, { pointerId: 2, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 2, clientX: 13, clientY: 12 })
    fireEvent.pointerUp(component, { pointerId: 2, clientX: 13, clientY: 12 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"moveComponent","version":1,"id":"e9","x":3,"y":2,"snap":true}')
  })

  it('tracks a drag and a resize live in the geometry fields, then lands the accepted engine geometry', async () => {
    const placed = { id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true }
    const componentCanvas = { ...canvas, components: [placed] }
    const movedCanvas = { ...canvas, components: [{ ...placed, x: 6_000, y: 4_000 }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: movedCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.click(component)
    const x = screen.getByRole('textbox', { name: 'X (pt)' })
    const y = screen.getByRole('textbox', { name: 'Y (pt)' })
    expect(x).toHaveValue('0')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 13, clientY: 12 })
    // The transient proposal the canvas paints is the value the panel shows,
    // and it cannot be typed over while the pointer owns it.
    expect(x).toHaveValue('3')
    expect(y).toHaveValue('2')
    expect(x).toHaveAttribute('readonly')
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 19, clientY: 10 })
    expect(x).toHaveValue('9')
    expect(y).toHaveValue('0')
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 19, clientY: 10 })
    // Go's accepted geometry replaces the proposal; 9 was never committed.
    await waitFor(() => expect(x).toHaveValue('6'))
    expect(y).toHaveValue('4')
    expect(x).not.toHaveAttribute('readonly')

    const width = screen.getByRole('textbox', { name: 'Width (pt)' })
    fireEvent.pointerDown(screen.getByRole('button', { name: 'Resize e9' }), { pointerId: 2, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(screen.getByRole('button', { name: 'Resize e9' }), { pointerId: 2, clientX: 8, clientY: 5 })
    expect(width).toHaveValue('80')
    expect(screen.getByRole('textbox', { name: 'Height (pt)' })).toHaveValue('29')
  })

  it('keeps the drop proposal painted until the engine geometry lands', async () => {
    const placed = { id: 'e9', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true }
    const componentCanvas = { ...canvas, components: [placed] }
    const movedCanvas = { ...canvas, components: [{ ...placed, x: 6_000, y: 4_000 }] }
    let answer: (() => void) | undefined
    const request = vi.fn(async () => {
      await new Promise<void>((resolve) => { answer = resolve })
      return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: movedCanvas } }
    })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    const component = screen.getByLabelText('text component e9')
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 19, clientY: 10 })
    expect(component.style.getPropertyValue('--component-x')).toBe('9px')
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 19, clientY: 10 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // Red proof: dropping the proposal on pointer-up painted 0px here, so the
    // element visibly jumped back to where the drag started and stayed there
    // until Go answered. The proposal owns the paint until the answer lands.
    expect(component.style.getPropertyValue('--component-x')).toBe('9px')
    answer!()
    await waitFor(() => expect(component.style.getPropertyValue('--component-x')).toBe('6px'))
    expect(screen.getByRole('textbox', { name: 'X (pt)' })).not.toHaveAttribute('readonly')
  })

  it('toggles Shift-click selection once without engine traffic and clears it on an empty canvas click', () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true }, { id: 'e2', type: 'rect' as const, band: 'content' as const, x: 80000, y: 0, width: 72000, height: 24000, resizable: true }] }
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.pointerDown(screen.getByLabelText('text component e1'), { pointerId: 1, clientX: 1, clientY: 1 })
    fireEvent.pointerUp(screen.getByLabelText('text component e1'), { pointerId: 1, clientX: 1, clientY: 1 })
    fireEvent.pointerDown(screen.getByLabelText('rect component e2'), { pointerId: 2, clientX: 1, clientY: 1, shiftKey: true })
    fireEvent.pointerUp(screen.getByLabelText('rect component e2'), { pointerId: 2, clientX: 1, clientY: 1, shiftKey: true })
    expect(screen.getByLabelText('Resize e1')).toBeInTheDocument()
    expect(screen.getByLabelText('Resize e2')).toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('Report page with Page Header, Content, and Page Footer'))
    expect(screen.queryByLabelText('Resize e1')).not.toBeInTheDocument()
    expect(screen.queryByLabelText('Resize e2')).not.toBeInTheDocument()
    expect(request).not.toHaveBeenCalled()
  })

  it('leaves a dirty session untouched for an open cancellation or failure', async () => {
    const files: FileAccess = { open: vi.fn().mockRejectedValueOnce(new FileAccessCancelled()).mockRejectedValueOnce(new Error('denied')), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine()} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 2, byteLength: 3 }} />)
    const open = screen.getByRole('button', { name: 'Open local template' })
    fireEvent.click(open)
    await waitFor(() => expect(files.open).toHaveBeenCalledTimes(1))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    fireEvent.click(open)
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Could not open local file'))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('clears temporary busy wording after save cancellation without changing the session', async () => {
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget: vi.fn(async () => { throw new FileAccessCancelled() }), writeSave: vi.fn() }
    render(<App engine={engine()} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 2, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Save As' }))
    await waitFor(() => expect(files.acquireSaveTarget).toHaveBeenCalledOnce())
    expect(screen.queryByText(/Preparing Save As/)).not.toBeInTheDocument()
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('acquires a target before serialization, preserves dirty on failure, and handles the Save shortcut', async () => {
    let rejectSave = true
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 3, byteLength: 3 }, bytes }))
    const acquireSaveTarget = vi.fn(async () => ({ name: 'untitled.folio' }))
    const writeSave = vi.fn(async () => { if (rejectSave) throw new Error('denied'); return { name: 'untitled.folio' } })
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget, writeSave }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 3, byteLength: 3 }} />)
    fireEvent.keyDown(window, { key: 's', ctrlKey: true })
    await waitFor(() => expect(writeSave).toHaveBeenCalledTimes(1))
    expect(acquireSaveTarget).toHaveBeenCalledBefore(request)
    expect(request).toHaveBeenCalledBefore(writeSave)
    expect(screen.getByRole('alert')).toHaveTextContent('Could not save local file')
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
    rejectSave = false
    fireEvent.click(screen.getByRole('button', { name: 'Save local template' }))
    await waitFor(() => expect(screen.getByText('Downloaded local file untitled.folio')).toBeInTheDocument())
    expect(screen.getByText('Saved local file')).toBeInTheDocument()
  })

  it('routes Start blank through the engine and returns to an unnamed unsaved local workspace', async () => {
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 9, byteLength: 3 }, bytes }))
    render(<App engine={engine(request)} fileAccess={{ open: vi.fn(), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }} blankBytes={bytes} initialSnapshot={{ documentState: 'loaded', revision: 4, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(request).toHaveBeenCalledWith('load', bytes))
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('keeps a noncanonical valid open dirty until the canonical engine bytes are written', async () => {
    const canonical = new Uint8Array([9, 8, 7]).buffer
    const request = vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: 7, byteLength: 3 }, ...(operation === 'serialize' ? { bytes: canonical } : {}) }))
    const files: FileAccess = { open: vi.fn(async () => ({ bytes, name: 'noncanonical.folio' })), acquireSaveTarget: vi.fn(), writeSave: vi.fn() }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3 }} />)
    fireEvent.click(screen.getByRole('button', { name: 'Open local template' }))
    await waitFor(() => expect(screen.getByText(/canonical local changes need saving/)).toBeInTheDocument())
    expect(screen.getByText('Unsaved local changes')).toBeInTheDocument()
  })

  it('does not roll back or clean a newer engine revision after an older save settles', async () => {
    let releaseWrite: (() => void) | undefined
    let releaseCommit: (() => void) | undefined
    const writeSave = vi.fn(() => new Promise<{ name: string }>((resolve) => { releaseWrite = () => resolve({ name: 'untitled.folio' }) }))
    const request = vi.fn((operation: string): Promise<{ snapshot: ReturnType<typeof snapshot>; bytes?: ArrayBuffer }> => {
      if (operation === 'command') return new Promise((resolve) => { releaseCommit = () => resolve({ snapshot: snapshot(3) }) })
      return Promise.resolve({ snapshot: snapshot(2), bytes })
    })
    const files: FileAccess = { open: vi.fn(), acquireSaveTarget: vi.fn(async () => ({ name: 'untitled.folio' })), writeSave }
    render(<App engine={engine(request)} fileAccess={files} initialSnapshot={snapshot(2)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    fireEvent.click(screen.getByRole('button', { name: 'Save local template' }))
    await waitFor(() => expect(writeSave).toHaveBeenCalledOnce())
    releaseCommit!()
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 3'))
    releaseWrite!()
    await waitFor(() => expect(screen.getByText('Unsaved local changes')).toBeInTheDocument())
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('REVISION 3')
    expect(screen.getByText(/newer local changes need saving/)).toBeInTheDocument()
  })

  it('paints each Go band rectangle at its projected origin and uses one zoomed display scale for page and grid', () => {
    render(<App initialSnapshot={snapshot(1)} />)
    const page = screen.getByLabelText('Report page with Page Header, Content, and Page Footer')
    const header = screen.getByLabelText('Page Header')
    expect(page.style.getPropertyValue('--page-display-width')).toBe('595.276px')
    expect(page.style.getPropertyValue('--page-display-height')).toBe('841.89px')
    expect(page.style.getPropertyValue('--grid-display-pitch')).toBe('6px')
    expect(header.style.getPropertyValue('--band-x')).toBe('36px')
    expect(header.style.getPropertyValue('--band-y')).toBe('36px')
    expect(header.style.getPropertyValue('--band-width')).toBe('523.276px')
    expect(header.style.getPropertyValue('--band-height')).toBe('20px')
    fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
    expect(page.style.getPropertyValue('--page-display-width')).toBe('654.8036px')
    expect(page.style.getPropertyValue('--grid-display-pitch')).toBe('6.6px')
  })

  it('paints only pre-broken engine text lines without changing local document state', () => {
    const textCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72000, height: 24000, resizable: true, value: 'do not paint this value', textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 12000, advance: 16000, width: 24000, fragments: [{ text: 'engine ', x: 0 }, { text: 'line', x: 16000 }] }] } }] }
    const request = vi.fn()
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} />)
    expect(screen.getByText('engine', { exact: true })).toBeInTheDocument()
    expect(screen.getByText('line', { exact: true })).toBeInTheDocument()
    expect(screen.queryByText('do not paint this value')).not.toBeInTheDocument()
    expect(screen.getByLabelText('text component e1: engine line')).toBeInTheDocument()
    expect(document.querySelector('.canvas-text-line')).toHaveStyle({ '--text-line-baseline': '12px', '--text-line-advance': '16px' })
    expect(request).not.toHaveBeenCalled()
  })

  // STORY 8.4a. THE FACE THE DOCUMENT ITSELF CARRIES, PAINTED WITH.
  //
  // The engine measures and renders with a face out of the document's own
  // `assets` map and attributes each painted fragment to the asset it resolved
  // to. Without this the browser had no CSS family for such a face at all, so
  // the canvas rasterized at the engine's x-positions with a fallback's
  // metrics and the glyphs collided — the reported defect, rebuilt for a
  // document that carries its own typeface.
  //
  // THE PAGE'S FONT SET IS STUBBED because jsdom implements none: it defines
  // no `FontFace` and no font set on `document`, so both are installed with
  // `Object.defineProperty` (neither exists to be assigned over). Nothing here
  // measures anything, and the seam under test computes no metric either —
  // every x, advance and line break in the fixture is the engine's.
  it('registers a carried face once for the whole document and paints its fragments with the family derived from its key', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation === 'asset') { requested.push(new TextDecoder().decode(payload)); return { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer } }
        return { snapshot: snapshot(1) }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: carriedFaceCanvas(key) }} />)
      // ONE REQUEST FOR THE WHOLE DOCUMENT, not one per component. The nearest
      // precedent, ImagePaint, is mounted once per component AND once per
      // repeated sheet; copying that lifetime here would add one face many
      // times under a single global family name and let an unmounting instance
      // delete a face another is still painting with. The fixture has two text
      // components drawing through the same carried entry precisely so a
      // per-component lifetime would show up as two requests.
      await waitFor(() => expect(requested).toEqual([key]))
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      expect(painted().length).toBe(2)
      await waitFor(() => expect(painted().map((node) => node.style.fontFamily)).toEqual([embeddedFaceFamily(key), embeddedFaceFamily(key)]))
      // AND THE ENGINE'S OWN GEOMETRY IS UNTOUCHED BY IT (AD-17): the fragment
      // still paints at the x the engine measured, and the family is the only
      // thing this story added to it.
      expect(painted()[0]).toHaveStyle({ '--text-fragment-x': '0px' })
      expect(screen.getAllByLabelText(/text component e1/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // EVERY CONTENT WINDOW AFTER THE FIRST IS AN ECHO, so a document that runs
  // past one sheet paints most of itself through ComponentEcho rather than
  // through the home occurrence. A carried face that reached only the home
  // would leave sheets 2..N rasterizing at the ENGINE'S x-positions on the
  // fallback stack — precisely the collision this story exists to remove, on
  // most of the pages.
  //
  // MUTATION PROOF, RUN AND RECORDED: replacing `carriedFaces={carriedFaces}`
  // with `carriedFaces={NO_CARRIED_FACES}` in ComponentEcho reddens THIS test
  // on the echoed fragment and leaves every other designer test green — which
  // is why the assertion is written against the echoed node specifically and
  // not against the container's fragments as a set.
  it('paints a repeated sheet\'s echo with the carried face, not only the component\'s home', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const request = vi.fn(async (operation: string) => operation === 'asset'
        ? { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer }
        : { snapshot: snapshot(1) })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: carriedFaceEchoCanvas(key) }} />)
      const home = () => Array.from(view.container.querySelectorAll('.canvas-component:not(.canvas-component-echo) .canvas-text-fragment')) as HTMLElement[]
      const echoed = () => Array.from(view.container.querySelectorAll('.canvas-component-echo .canvas-text-fragment')) as HTMLElement[]
      // The fixture really does produce both, or the claim below is vacuous.
      expect(home()).toHaveLength(1)
      expect(echoed()).toHaveLength(1)
      await waitFor(() => expect(echoed()[0]!.style.fontFamily).toBe(embeddedFaceFamily(key)))
      expect(home()[0]!.style.fontFamily).toBe(embeddedFaceFamily(key))
      // AD-17 on the echo too: the engine's own x is what it paints at, and
      // the family is the only thing this story put on it.
      expect(echoed()[0]!).toHaveStyle({ '--text-fragment-x': '0px' })
    } finally {
      fontSet.restore()
    }
  })

  // STORY 8.4e. THE FACE THE BUILD SHIPS, ASKED FOR BY THE NAME THE ENGINE
  // MEASURED WITH — AT THE HOME FRAGMENT AND AT THE ECHO.
  //
  // Until this story a shipped fragment set no family at all and fell to
  // `.canvas-text-fragment`'s fixed stack, which names all three shipped faces
  // in one Latin-first order whatever order the document declared. All three
  // faces cover `A` and `5` (their cmaps overlap 339 / 529 / 230 codepoints
  // pairwise, measured), so a document whose chain is ["Noto Sans Thai"] had
  // its Latin MEASURED with Noto Sans Thai and RASTERIZED with Noto Sans:
  // right glyphs, wrong advances, creeping out of position at the engine's own
  // x. The fragment now names the attributed face FIRST.
  //
  // WHAT THIS LAYER CAN AND CANNOT PROVE, said out loud rather than implied.
  // jsdom applies no stylesheet and loads no font, so "rasterized with" is not
  // observable here; what is observable — and is what the fix consists of — is
  // that the engine's name reaches the element and the element asks for it
  // first. The executed browser assertion is owed at the epic gate, once CI
  // runs the Playwright suite (D-8.4.25(b), (d), (e)).
  //
  // MUTATION PROOF, RUN AND RECORDED: dropping the shipped branch from
  // TextPaint's inline style reddens this test at BOTH nodes; replacing
  // `carriedFaces={carriedFaces}` with `NO_CARRIED_FACES` in ComponentEcho
  // does NOT redden it, which is why the echo is asserted directly.
  it('paints a shipped-face fragment with the face the engine attributed it to, at the home occurrence and at the echo', async () => {
    const operations: string[] = []
    const request = vi.fn(async (operation: string) => { operations.push(operation); return { snapshot: snapshot(1) } })
    const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: shippedFaceEchoCanvas('Noto Sans Thai') }} />)
    const home = () => Array.from(view.container.querySelectorAll('.canvas-component:not(.canvas-component-echo) .canvas-text-fragment')) as HTMLElement[]
    const echoed = () => Array.from(view.container.querySelectorAll('.canvas-component-echo .canvas-text-fragment')) as HTMLElement[]
    // The fixture really does produce both, or the claim below is vacuous.
    expect(home()).toHaveLength(1)
    expect(echoed()).toHaveLength(1)
    for (const node of [home()[0]!, echoed()[0]!]) {
      expect(familiesAskedFor(node)[0]).toBe('Noto Sans Thai')
      // AND THE DECLARED STACK IS STILL BEHIND IT. An inline declaration
      // replaces the rule rather than extending it, so a codepoint the
      // attributed face does not cover must still reach the other shipped
      // faces rather than the browser's default.
      expect(familiesAskedFor(node)).toEqual(['Noto Sans Thai', 'Noto Sans', 'Noto Sans SC', 'sans-serif'])
      // AD-17 on both: the engine's own x is what it paints at, and the family
      // is the only thing this story put on it.
      expect(node).toHaveStyle({ '--text-fragment-x': '0px' })
    }
    // NOTHING WAS FETCHED AND NOTHING WAS REGISTERED. A shipped face is
    // declared at build time; it needs no `asset` request and no runtime seam,
    // and asking for one would be a per-document cost this story does not add.
    expect(operations).not.toContain('asset')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getAllByLabelText(/text component e1/)[0]).toBeInTheDocument()
  })

  // THE UNATTRIBUTED FRAGMENT, which is what the stylesheet's stack is FOR
  // now. A fragment carrying neither identity is legal on the wire — the
  // browser's guard admits it deliberately — and it must paint, on the
  // declared stack, with no inline family of its own. This is the direction
  // that keeps the assertion above from passing for the wrong reason.
  it('leaves a fragment the engine attributed to nothing on the stylesheet\'s declared stack', async () => {
    const bare = shippedFaceEchoCanvas('Noto Sans Thai')
    const unattributed = { ...bare, components: bare.components.map((component) => ({ ...component, textPaint: { ...component.textPaint, lines: component.textPaint.lines.map((line) => ({ ...line, fragments: line.fragments.map(({ text, x }) => ({ text, x })) })) } })) }
    const view = render(<App engine={engine(vi.fn(async () => ({ snapshot: snapshot(1) })))} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: unattributed }} />)
    const painted = Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
    expect(painted).toHaveLength(2)
    for (const node of painted) {
      expect(node.style.fontFamily).toBe('')
      expect(node).toHaveStyle({ '--text-fragment-x': '0px' })
    }
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  // THE DEGRADE PATH, stated as a claim about the SESSION. An inline family
  // REPLACES the stylesheet rule rather than extending it, so asking for a
  // family whose bytes never arrived would take the fragment off the declared
  // stack onto whatever the browser defaults to. A failed asset request must
  // therefore leave the fragment exactly as a shipped-face fragment: painted,
  // named, and on the declared stack — and it must never reach the engine's
  // failure channel.
  //
  // IT IS READ AFTER THE CHAIN HAS SETTLED, WHICH IS THE WHOLE DIFFICULTY.
  // "The fragment has no family" is ALSO true of a registration that simply
  // has not finished, so asserting it the instant the request was issued
  // proves "not yet", not "degraded" — the identical assertion passes at that
  // instant in the SUCCESS case. The document therefore carries TWO carried
  // faces: one whose bytes are withheld until the other has already been
  // REJECTED, so the first face's arrival in the font set is a positive
  // condition that cannot hold until the failure was handled.
  //
  // MUTATION PROOF, RUN AND RECORDED: returning bytes for `unfetchable`
  // instead of throwing reddens this test — the second fragment acquires its
  // family and the font set holds two families, not one.
  it('keeps painting on the declared stack when a carried face\'s bytes cannot be fetched', async () => {
    const fetchable = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const unfetchable = 'fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210'
    const fontSet = installStubFontSet()
    try {
      let release: () => void = () => undefined
      const withheld = new Promise<ArrayBuffer>((resolve) => { release = () => resolve(new Uint8Array([0, 1, 2, 3]).buffer) })
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation !== 'asset') return { snapshot: snapshot(1) }
        const key = new TextDecoder().decode(payload)
        requested.push(key)
        if (key === unfetchable) throw Object.assign(new Error('no such asset'), { code: 'ASSET_UNAVAILABLE' })
        return { snapshot: snapshot(1), bytes: await withheld }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: twoCarriedFacesCanvas(fetchable, unfetchable) }} />)
      // Both were asked for; the unfetchable one has already rejected, because
      // its bytes were never withheld behind anything.
      await waitFor(() => expect([...requested].sort()).toEqual([fetchable, unfetchable].sort()))
      release()
      // THE POSITIVE CONDITION. The font set cannot hold the fetchable face
      // until its bytes were released, which happened after the other request
      // rejected — so everything below is read on a settled chain.
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(fetchable)]))
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      expect(painted().length).toBe(2)
      await waitFor(() => expect(painted()[0]!.style.fontFamily).toBe(embeddedFaceFamily(fetchable)))
      expect(painted()[1]!.style.fontFamily).toBe('')
      expect(painted()[1]!).toHaveStyle({ '--text-fragment-x': '0px' })
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getAllByLabelText(/text component e1/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // THE KEY IS A STRING FROM THE DOCUMENT, AND IT BECOMES A CSS FAMILY. The
  // projection admits a FRAGMENT's `assetKey` only as 64 lowercase hex, but a
  // CHAIN ENTRY's key — the one this effect fetches bytes for and derives the
  // family from — is admitted on length alone, so the shape is asserted at the
  // derivation. A key that is not one is a carried face the browser declines:
  // no request, no family, no registration, and the fragment stays on the
  // stylesheet's declared stack. It is a DEGRADE and not a refusal — nothing
  // reaches the failure channel and the worker is untouched.
  it('declines a chain entry whose asset key is not an asset key, rather than turning it into a family', async () => {
    const wellFormed = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const injected = 'IBM Plex Sans, monospace'
    const fontSet = installStubFontSet()
    try {
      const requested: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation === 'asset') { requested.push(new TextDecoder().decode(payload)); return { snapshot: snapshot(1), bytes: new Uint8Array([0, 1, 2, 3]).buffer } }
        return { snapshot: snapshot(1) }
      })
      const view = render(<App engine={engine(request) } initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: twoCarriedFacesCanvas(wellFormed, injected) }} />)
      // The well-formed sibling settles, so the claim about the malformed one
      // is read after the effect has done everything it is going to do.
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(wellFormed)]))
      expect(requested).toEqual([wellFormed])
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      await waitFor(() => expect(painted()[0]!.style.fontFamily).toBe(embeddedFaceFamily(wellFormed)))
      // Not a family, not a partial family, not the string itself.
      expect(painted()[1]!.style.fontFamily).toBe('')
      expect(view.container.innerHTML).not.toContain(embeddedFaceFamily(injected))
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getAllByLabelText(/text component e2/)[0]).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // THE LISTING IS PART OF THE EFFECT'S KEY, AND THIS IS THE CASE THAT PROVES
  // IT. `documentGenerationValue` advances only when the document is REPLACED
  // — open a file, new template, undo/redo — and an ordinary property commit
  // is none of those: it commits through setCurrentSnapshot without clearing
  // the document interaction, so the generation does not move.
  //
  // Release is invisible in the DOM — the fragment simply stops asking for the
  // family — so the claim is made against the font set's own record.
  //
  // MUTATION PROOF, RUN AND RECORDED: reducing the effect's dependency array
  // to `[engine, documentGenerationValue]` reddens this test and leaves the
  // rest of the designer suite green.
  //
  // TRIGGER CHANGED BY STORY 16.9: this used to drop the carried entry through
  // the chain editor's own 'Remove entry' control, which no longer exists —
  // the story removed the UI, not the chain data or the property-commit path.
  // The effect under test reacts to ANY snapshot update that narrows
  // `fontChains`, regardless of which command produced it, so an ordinary
  // Bold toggle (still available, still a same-document 'command') is used to
  // deliver the identical projection change the removed control used to.
  it('releases a carried face when a same-document commit drops the chain entry that carried it', async () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const fontSet = installStubFontSet()
    try {
      const carrying = carriedFaceCanvas(key)
      const dropped = { ...carrying, fontChains: [{ name: 'body', entries: [face('Noto Sans')] }], components: carrying.components.map((component) => ({ ...component, bold: true })) }
      const loaded = (revision: number, projection: typeof carrying) => ({ documentState: 'loaded' as const, revision, byteLength: 3, canvas: projection })
      const request = vi.fn(async (operation: string) => {
        if (operation === 'asset') return { snapshot: loaded(1, carrying), bytes: new Uint8Array([0, 1, 2, 3]).buffer }
        if (operation === 'command') return { snapshot: loaded(2, dropped) }
        return { snapshot: loaded(1, carrying) }
      })
      render(<App engine={engine(request) } initialSnapshot={loaded(1, carrying)} />)
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(key)]))
      expect(fontSet.removed).toEqual([])
      fireEvent.click(screen.getAllByLabelText(/text component e1/)[0]!)
      // Any same-document command that returns a narrowed `fontChains`
      // exercises the same effect the deleted chain-editor control drove.
      fireEvent.click(screen.getByRole('button', { name: 'Bold' }))
      await waitFor(() => expect(fontSet.removed).toEqual([embeddedFaceFamily(key)]))
      // And it was released rather than re-registered: the document was never
      // replaced, so nothing asked for those bytes a second time.
      expect(fontSet.added).toEqual([embeddedFaceFamily(key)])
      expect(document.querySelectorAll('.canvas-text-fragment')).toHaveLength(2)
      expect(Array.from(document.querySelectorAll('.canvas-text-fragment')).map((node) => (node as HTMLElement).style.fontFamily)).toEqual(['', ''])
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  it('retains literal empty drafts, announces the precise engine diagnostic, and ignores a stale Apply draft reset', async () => {
    let resolveApply: ((value: { snapshot: ReturnType<typeof snapshot> }) => void) | undefined
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: ReturnType<typeof snapshot> }>((resolve) => { resolveApply = resolve }) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={snapshot(1)} />)
    const top = screen.getByRole('textbox', { name: 'Top margin (pt)' })
    fireEvent.change(top, { target: { value: '37' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    fireEvent.change(top, { target: { value: '38' } })
    resolveApply!({ snapshot: snapshot(2) })
    await waitFor(() => expect(top).toHaveValue('38'))
    request.mockRejectedValueOnce(Object.assign(new Error('must not be negative'), { code: 'PAGE_SETUP_INVALID', dataPath: 'page.margin.top' }))
    fireEvent.change(top, { target: { value: '' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply page setup' }))
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('page.margin.top: must not be negative'))
    expect(top).toHaveValue('')
  })

  it('keeps component drafts local, sends exactly one Enter/Blur commit, and locates a Go diagnostic', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello', fontFamily: 'body', fontSize: 12_000, borderEdges: ['bottom' as const] }] }
    const request = vi.fn((operation: string) => operation === 'command' ? Promise.reject(Object.assign(new Error('must fit the content band'), { elementId: 'e1', dataPath: 'component.x' })) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const x = screen.getByRole('textbox', { name: 'X (pt)' })
    fireEvent.change(x, { target: { value: '9999' } })
    expect(request).not.toHaveBeenCalled()
    fireEvent.keyDown(x, { key: 'Enter' })
    fireEvent.blur(x)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('e1: component.x: must fit the content band'))
    expect(x).toHaveValue('9999')
    expect(x).toHaveAttribute('aria-invalid', 'true')
  })

  it('routes the one CONTENT field to the value or expression command by what was typed', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    expect(screen.queryByRole('textbox', { name: 'Text expression' })).not.toBeInTheDocument()
    const field = screen.getByRole('textbox', { name: 'Text' })
    // Story 7.4: CONTENT is a textarea, so Enter puts a line feed in the
    // draft and commits NOTHING. Blur is this field's commit; every other
    // field keeps Enter.
    fireEvent.change(field, { target: { value: 'Customer: {{customer.name}}' } })
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(sent).toHaveLength(0)
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"expression":{"op":"set","value":"Customer: {{customer.name}}"}}}')
    const again = screen.getByRole('textbox', { name: 'Text' })
    fireEvent.change(again, { target: { value: 'Plain heading' } })
    fireEvent.blur(again)
    await waitFor(() => expect(sent).toHaveLength(2))
    expect(new TextDecoder().decode(sent[1]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"Plain heading"}}}')
  })

  it('marks the expression-bearing fields with fx, and lights it once the field holds one', () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const text = screen.getByRole('textbox', { name: 'Text' })
    const marker = (input: HTMLElement) => input.parentElement!.querySelector('.property-fx')
    expect(text).toHaveAttribute('aria-description', 'Accepts literal text, or {{ }} expressions')
    expect(marker(text)).toHaveTextContent('fx')
    expect(marker(text)).not.toHaveClass('property-fx-active')
    fireEvent.change(text, { target: { value: 'Customer: {{customer.name}}' } })
    expect(marker(text)).toHaveClass('property-fx-active')
    // A geometry field is a literal to Go, which rejects a placeholder in it:
    // it must carry no cue at all.
    expect(marker(screen.getByRole('textbox', { name: 'X (pt)' }))).toBeNull()
    const visible = screen.getByRole('textbox', { name: 'Visible if' })
    expect(visible).toHaveAttribute('aria-description', 'Accepts a boolean data path or call, written without {{ }} — the grammar has no comparisons')
    expect(marker(visible)).not.toHaveClass('property-fx-active')
    fireEvent.change(visible, { target: { value: 'customer.isActive' } })
    expect(marker(visible)).toHaveClass('property-fx-active')
  })

  it('authors BOX colours through the picker, states pt on an empty size, and drops the padding rows', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    // An unset size still says which unit it wants.
    const border = screen.getByRole('textbox', { name: 'Border width (pt)' })
    expect(border).toHaveValue('')
    expect(border.parentElement).toHaveTextContent('pt')
    for (const label of ['Padding top (pt)', 'Padding right (pt)', 'Padding bottom (pt)', 'Padding left (pt)']) expect(screen.queryByRole('textbox', { name: label })).not.toBeInTheDocument()
    const picker = screen.getByLabelText('Pick Border colour')
    fireEvent.change(picker, { target: { value: '#c81e1e' } })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"borderColor":{"op":"set","value":"#c81e1e"}}}')
    fireEvent.change(screen.getByLabelText('Pick Background'), { target: { value: '#0b1120' } })
    await waitFor(() => expect(sent).toHaveLength(2))
    expect(new TextDecoder().decode(sent[1]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"background":{"op":"set","value":"#0b1120"}}}')
  })

  it('paints the engine-projected box on the canvas, and nothing where none is projected', () => {
    const boxed = { id: 'e1', type: 'rect' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, background: '#1b2a4a', borderWidth: 2_000, borderColor: '#c81e1e', borderEdges: ['bottom' as const] }
    const plain = { id: 'e2', type: 'text' as const, band: 'content' as const, x: 0, y: 40_000, width: 72_000, height: 24_000, resizable: true, value: 'Plain' }
    render(<App initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [boxed, plain] } }} />)
    const box = screen.getByLabelText('rect component e1').querySelector('.canvas-box') as HTMLElement
    expect(box).not.toBeNull()
    expect(box.style.background).toBe('rgb(27, 42, 74)')
    expect(box.style.borderBottom).toBe('2px solid rgb(200, 30, 30)')
    // An edge the engine does not paint is not painted here either.
    expect(box.style.borderTop).toBe('0px')
    expect(screen.getByLabelText('text component e2').querySelector('.canvas-box')).toBeNull()
  })

  it('sets the text colour from TYPOGRAPHY and paints the canvas in it', async () => {
    const inked = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello', color: '#c81e1e', textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 10_000, advance: 12_000, width: 30_000, fragments: [{ text: 'Hello', x: 0 }] }] } }
    const sent: ArrayBuffer[] = []
    const componentCanvas = { ...canvas, components: [inked] }
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 1 + sent.length, byteLength: 3, canvas: componentCanvas } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    // The canvas paints the engine's ink, never a browser-side default.
    const element = screen.getByLabelText('text component e1: Hello')
    const paint = element.querySelector('.canvas-text-paint') as HTMLElement
    expect(paint.style.getPropertyValue('--text-ink')).toBe('#c81e1e')
    fireEvent.click(element)
    fireEvent.change(screen.getByLabelText('Pick Text colour'), { target: { value: '#1b2a4a' } })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"color":{"op":"set","value":"#1b2a4a"}}}')
  })

  it('keeps a newer property draft through an unrelated successful snapshot and exposes table truth', async () => {
    const componentCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }, { id: 'e2', type: 'table' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }] }
    let resolve: ((value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }) => void) | undefined
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof componentCanvas } }>((done) => { resolve = done }) : Promise.resolve({ snapshot: snapshot(1) }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const value = screen.getByRole('textbox', { name: 'Text' })
    fireEvent.change(value, { target: { value: 'newer literal' } })
    fireEvent.blur(value)
    resolve!({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 3, canvas: componentCanvas } })
    await waitFor(() => expect(value).toHaveValue('newer literal'))
    fireEvent.click(screen.getByLabelText('table component e2'))
    expect(screen.queryByRole('textbox', { name: 'Width (pt)' })).not.toBeInTheDocument()
    expect(screen.getByText('Table binding: transactions[] (display only)')).toBeInTheDocument()
  })
})

// STORY 16.3 — THE FONT BROWSER AT THE APP SEAM.
//
// WHAT THESE COVER THAT `FontBrowser.test.tsx` CANNOT. That file renders the
// modal directly and supplies its own `sources`, `inTemplate` and
// `previewBytes`, so every wire between App and the modal is stubbed out. A
// review demonstrated the gap by DROPPING the second argument from
// `offeredFamilies(query, stored)` in `browsableFamilies`: it type-checks,
// because the parameter defaults to `[]`, and every face this machine already
// holds silently vanishes from the browser — with the whole suite green. These
// tests render the real `App` and address the modal through the door.
describe('the font browser opens from the family control', () => {
  // THE MODAL FETCHES THE MOMENT IT OPENS, and those fetches must not outlive
  // the test that started them. Every row on the first page asks
  // `browserSpecimenBytes` for a face, and for the web tier that is
  // `fetchWebFamily` — up to four probes each. Left to the real global `fetch`
  // they were still in flight when the test ended, and the next test's own stub
  // then counted them as its own: a measured 10 calls where one was expected,
  // in a test about a completely different control.
  //
  // A REJECTING FETCH IS THE FIX RATHER THAN A MUTE ONE, because
  // `fetchWebFamily` STOPS at the first probe that throws. There is no second
  // probe to leak, and every row settles on "cannot be shown set in itself" —
  // which is the correct rendering for a machine with no route upstream and the
  // state these tests read anyway.
  let restoreFetch: typeof globalThis.fetch
  beforeEach(() => {
    restoreFetch = globalThis.fetch
    globalThis.fetch = vi.fn(async () => { throw new TypeError('Failed to fetch') }) as never
  })
  afterEach(async () => {
    // Drain whatever the rejections queued before handing the global back.
    await new Promise((resolve) => setTimeout(resolve, 0))
    globalThis.fetch = restoreFetch
  })

  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  // `Arimo` is a COMMITTED local-tier face, so it is in the offered population
  // with no network at all, and declaring a chain of that name puts the same
  // family on both sides of the `In template` question.
  const withArimoDeclared = { ...canvas, components: [textComponent], fontFamilies: ['body', 'Arimo'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'Arimo', entries: [face('Noto Sans')] }] }

  const openDoor = (projection = withArimoDeclared) => {
    const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } }))
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: projection }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    fireEvent.click(screen.getByRole('button', { name: /^Add fonts…/ }))
    return request
  }

  it('mounts the dialog from the last row of the open dropdown', () => {
    const restore = installStubFontSet()
    try {
      openDoor()
      const dialog = screen.getByRole('dialog', { name: 'Font browser' })
      expect(dialog).toHaveAttribute('aria-modal', 'true')
      // The door closed the dropdown on its way out, so the listbox is gone and
      // the modal is what has focus.
      expect(screen.queryByRole('listbox', { name: 'Fonts' })).toBeNull()
    } finally {
      restore.restore()
    }
  })

  it('names the sub-label in the door\'s accessible name, not only on screen', () => {
    const restore = installStubFontSet()
    try {
      const request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } }))
      render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: withArimoDeclared }} />)
      fireEvent.click(screen.getByLabelText('text component e1'))
      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      // An `aria-label` REPLACES the contents, so a bare "Add fonts…" deleted
      // the one sentence saying what the row does for everybody who cannot see
      // it.
      const door = screen.getByRole('button', { name: 'Add fonts… Browse and embed web fonts' })
      expect(door).toBeInTheDocument()
      // AND NO SHORTCUT GLYPH, WHICH IS RULED RATHER THAN FORGOTTEN (D-16.R.33
      // R2). The mockup prints `⌘G` beside this row and no key is bound, so a
      // glyph here would be a false UI string. Story 16.4 restates the matrix
      // row; the assertion is added because nothing held it.
      expect(door.textContent, 'a hint glyph beside an unbound key is a false label').not.toMatch(/⌘|⌥/)
      expect(screen.getByRole('listbox', { name: 'Fonts' }).contains(door), 'the door is a real button OUTSIDE the listbox, never a non-option child of it').toBe(false)
    } finally {
      restore.restore()
    }
  })

  it('carries the DOCUMENT\'s declared chains into the modal as `In template`', async () => {
    const restore = installStubFontSet()
    // A STORE IS SUPPLIED SO THE DIALOG IS IN ITS ORDINARY MODE. jsdom provides
    // no IndexedDB, and a browser that cannot keep typefaces puts the confirm
    // control into Story 16.5's degraded model, where it names the count instead
    // of the action — which is asserted in `FontBrowser.test.tsx`, and is not
    // what this test is about.
    const restoreStore = withMachineStore()
    try {
      openDoor()
      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'Arimo' } })
      // `canvas.fontFamilies` reached the modal: the family the document
      // already declares cannot be staged again.
      const inTemplate = await screen.findByRole('button', { name: 'Arimo is in this template' })
      expect(inTemplate).toBeDisabled()
      expect(screen.getByRole('button', { name: 'Install on this machine' })).toBeDisabled()
    } finally {
      restore.restore()
      restoreStore()
    }
  })

  it('returns focus to the family control when the modal closes', async () => {
    const restore = installStubFontSet()
    try {
      openDoor()
      fireEvent.keyDown(screen.getByRole('dialog', { name: 'Font browser' }), { key: 'Escape' })
      await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Font browser' })).toBeNull())
      // The door unmounted with the dropdown, so without this focus would be on
      // `<body>` and a keyboard-only author would tab in from the top of the
      // page after every Escape (UX-DR25).
      await waitFor(() => expect(screen.getByRole('combobox', { name: 'Font family' })).toHaveFocus())
    } finally {
      restore.restore()
    }
  })
})

describe('typography controls over the engine-projected closed sets', () => {
  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  const select = (request = vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } }))) => {
    const componentCanvas = { ...canvas, components: [textComponent] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    return request
  }

  it('offers the document\'s declared font chains, searched, and commits the chosen one', async () => {
    const request = select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    // SCOPED TO THE DECLARED GROUP. Since Story 8.6 the listbox carries a
    // second group — the bundled catalogue — and this test is about the first:
    // a catalogue entry is not a chain the document declares, and folding the
    // two together would make "the document's declared chains" mean "every
    // family that exists".
    const listed = () => declaredOptions().map(optionText)
    expect(listed()).toEqual(['body', 'heading'])
    fireEvent.change(combobox, { target: { value: 'head' } })
    expect(listed()).toEqual(['heading'])
    fireEvent.click(screen.getByRole('option', { name: 'heading' }))
    await waitFor(() => expect(request).toHaveBeenCalledWith('command', expect.anything()))
    // The typed search text is never a value: only a listed family is sent.
    expect(screen.queryByRole('listbox', { name: 'Fonts' })).not.toBeInTheDocument()
  })

  it('states a search that matches neither a declared chain nor a family on this machine, instead of offering to invent one', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    fireEvent.change(combobox, { target: { value: 'Helvetica' } })
    expect(screen.queryByRole('option')).not.toBeInTheDocument()
    // CORRECTED WITH THE CHANGE, NOT EDITED QUIETLY (Story 16.4, narrowed by
    // 16.9). This pinned `Nothing in this document or the catalogue matches …`
    // — a sentence naming ONE of the three places the control searched at
    // 16.4, written when the catalogue was the whole offer. Story 16.9 then
    // dropped the control's own third place, "or in the list you can install"
    // — this control no longer searches there at all, `Add fonts…` does — so
    // the sentence names only the two groups drawn above it now.
    expect(screen.getByText('Nothing in this template or on this machine matches "Helvetica".')).toBeInTheDocument()
  })

  // STORY 16.4 — THREE GROUPS, ON THE AXIS THE CODE ALREADY FORKS ON.
  //
  // The grouping key is (declared?, `familyIsInstalled`?) and nothing else, so
  // these assertions are written against what each group's rows DO — the note a
  // row carries says whether picking it uses the face or downloads it — rather
  // than against a class name, which would pass on a control that grouped
  // alphabetically.
  const dropdown = () => screen.getByRole('listbox', { name: 'Fonts' })
  const groupRows = (label: string) => within(screen.getByRole('group', { name: label })).getAllByRole('option').map(optionText)

  // STORY 16.9 NARROWED THIS FROM THREE GROUPS TO TWO: `AVAILABLE TO INSTALL`
  // (the `web` arm, ~1,273 families not on this machine) is gone from this
  // dropdown. `Add fonts…` is now the only door to that relationship; this
  // control offers only what is usable now, with no network and no wait.
  it('draws the two groups, disjoint and complete, on the where-are-the-bytes axis', () => {
    select()
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    // THE ORDER IS THE MODEL: in the file, on the machine.
    expect(within(dropdown()).getAllByRole('group').map((group) => group.getAttribute('aria-label'))).toEqual(['IN THIS TEMPLATE', 'AVAILABLE LOCALLY'])
    const template = groupRows('IN THIS TEMPLATE')
    const local = groupRows('AVAILABLE LOCALLY')
    // COMPLETE: every option the listbox owns sits in exactly one group.
    expect(template.length + local.length).toBe(within(dropdown()).getAllByRole('option').length)
    // DISJOINT: no family is offered twice under two different promises.
    const names = [...template, ...local].map((text) => text.split(' — ')[0])
    expect(new Set(names).size).toBe(names.length)
    // AND EACH HEADING TELLS THE TRUTH ABOUT ITS OWN ROWS.
    expect(template).toEqual(['body', 'heading'])
    // STORY 16.7 (Design Note 2, narrowing D-16.R.72): the per-row note that
    // used to restate "needs no download" is now the specimen instead, so the
    // truthful claim left to make about AVAILABLE LOCALLY's OWN ROWS is that
    // none of them still carries any note at all — every row this dropdown
    // draws is one this machine already holds, so no row needs one any more.
    expect(local.every((text) => !text.includes(' — ')), `no AVAILABLE LOCALLY row may still carry a per-row note: ${local.slice(0, 3).join(' / ')}`).toBe(true)
    // THE SECOND GROUP IS POPULATED ON A FRESH MACHINE, which is the half of
    // D-16.R.72 a store-shaped reading of the heading would have got wrong: the
    // committed faces ship inside the release, so they are always on it.
    expect(local, 'the committed faces are on this machine whether or not anything was ever downloaded').toHaveLength(catalogueFaces.length)
  })

  // STORY 16.9 — THERE IS NO THIRD GROUP LEFT TO CAP. Both remaining groups
  // render in full, unconditionally: neither is bounded any more, because
  // the population that once needed a bound (~1,273 web-tier rows) is no
  // longer offered here at all.
  it('renders both remaining groups in full, with no cap and no "Showing N of M" note', () => {
    select()
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    expect(groupRows('IN THIS TEMPLATE')).toHaveLength(2)
    expect(groupRows('AVAILABLE LOCALLY')).toHaveLength(catalogueFaces.length)
    expect(screen.queryByRole('group', { name: 'AVAILABLE TO INSTALL' })).not.toBeInTheDocument()
    expect(screen.queryByText(/Showing \d+ of \d+/)).not.toBeInTheDocument()
  })

  // STORY 16.9 ALSO REMOVED TWO OTHER CONTROLS FROM THIS FIELD — THE CLEAR
  // BUTTON, AND THE DISCLOSURE THAT REVEALED THE FONT-CHAIN EDITOR — AND
  // UNTIL NOW NEITHER HAD A GUARD OF ITS OWN, ONLY A RETIREMENT COMMENT. A
  // comment cannot fail a build; this can.
  //
  // POPULATION STATED BESIDE THE ZERO, PER STANDING RULE: with the combobox
  // focused, this property renders exactly one button (`buttonNames` below) —
  // the disclosure fixed by the chevron correction above. If either removed
  // control ever came back it would be a SECOND or THIRD entry in that same
  // array, not a query silently matching nothing; the assertions below are
  // therefore not the only thing that would catch a regression, but they name
  // the two controls this story specifically removed rather than leaving that
  // to an incidental count.
  it('offers no control to clear the family, and none to reveal a font-chain editor', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    // SCOPED TO THIS FIELD, NOT THE WHOLE PAGE: the property panel around it
    // carries its own buttons (Bold, alignment, Add fonts…) that have nothing
    // to do with what this test is about. `.property-combobox` is the field's
    // own wrapper, rendered by `FontFamilyProperty` itself.
    const field = combobox.closest('.property-combobox')
    if (!field) throw new Error('the font family combobox is no longer inside .property-combobox; re-scope this test rather than deleting it')
    const buttonNames = within(field as HTMLElement).getAllByRole('button').map((button) => button.getAttribute('aria-label') ?? button.textContent)
    // `.property-combobox` wraps the field AND its open overlay (the shell
    // holding the listbox and, below it, `Add fonts…`), so both legitimate
    // buttons appear here — the population is two, not one, and that is
    // stated rather than assumed.
    expect(buttonNames, 'the population this zero is measured against — every button this field renders with the dropdown open').toEqual(['Hide fonts', 'Add fonts… Browse and embed web fonts'])
    expect(screen.queryByRole('button', { name: 'Clear Font family' }), 'text always has a typeface; there is no control to clear it').not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Edit font chains' }), '`FontChainEditor.tsx` and its render site are deleted').not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Hide font chains' }), 'the same disclosure, in its other label').not.toBeInTheDocument()
  })

  // STORY 16.9 DISCHARGES A REGISTERED HAZARD: a pick from the removed
  // `AVAILABLE TO INSTALL` group used to block up to 30s on a stall and 180s
  // against a slow host, because the dropdown could trigger a fetch that only
  // a web-tier row could reach (`fetchWebFamily`, against the one declared
  // repository host). With no web row left in this control, no code path
  // from opening or filtering the dropdown can reach that fetch at all —
  // never mind stall or succeed.
  //
  // "NO REQUEST" MEANS NO REQUEST TO A THIRD PARTY, NOT LITERALLY ZERO CALLS
  // TO `fetch`. `AVAILABLE LOCALLY`'s own specimens (Story 16.7) read their
  // bytes from THIS RELEASE'S OWN bundle, over a relative URL, the same read
  // `runtimeAssetUrls` assets get behind the service worker — offline-capable,
  // no third party, and unrelated to the hazard this story closes. So this
  // installs the same page-font stub the specimen tests use (`installStubFontSet`,
  // above) rather than leaving `FontFace` undefined, which would make every
  // specimen fetch silently never fire and prove nothing about THIS story's
  // claim. The assertion is that among however many same-origin reads happen,
  // NONE is an absolute URL — which is exactly what would change if a web-tier
  // row, and the fetch it can reach, ever came back.
  //
  // THE OBSERVER IS PROVEN ALIVE, NOT ASSUMED (the "vacuous survivor" this
  // suite has been burned by before): the same spy that records only
  // same-origin calls from opening and filtering is proven capable of
  // recording a call at all, from the very rows this dropdown draws.
  it('reaches no third-party host opening and filtering the dropdown, though the same spy can see the local reads it does make', async () => {
    const bundleAsset = new Uint8Array([0x00, 0x01, 0x00, 0x00, 0x7f]).buffer
    const fetchStub = vi.fn(async (_url: string) => ({ ok: true, arrayBuffer: async () => bundleAsset }))
    const restore = globalThis.fetch
    globalThis.fetch = fetchStub as never
    const fontSet = installStubFontSet()
    try {
      select()
      const combobox = screen.getByRole('combobox', { name: 'Font family' })
      fireEvent.focus(combobox)
      // SETTLES ON A REAL SPECIMEN, so the assertions below run after
      // whatever this dropdown is going to fetch has actually been asked
      // for — never on a synchronous race that would pass by never giving
      // the effect a chance to run.
      await waitFor(() => expect(fetchStub).toHaveBeenCalled())
      fireEvent.change(combobox, { target: { value: 'lora' } })
      fireEvent.change(combobox, { target: { value: 'Kanit' } })
      fireEvent.change(combobox, { target: { value: '' } })
      await waitFor(() => expect(groupRows('AVAILABLE LOCALLY').length).toBeGreaterThan(0))
      // THE CORE CLAIM: not one of however many calls happened names an
      // external host. `fetchWebFamily` is the one function in this codebase
      // that ever builds an absolute URL to the declared repository host, and
      // nothing reachable from this dropdown calls it any more.
      const urls = fetchStub.mock.calls.map((call) => String(call[0]))
      expect(urls.length, 'the observer must have seen something, or the claim below is vacuous').toBeGreaterThan(0)
      expect(urls.every((url) => !/^https?:\/\//.test(url)), `every request must be same-origin, never a third-party host: ${urls.filter((url) => /^https?:\/\//.test(url)).slice(0, 3).join(', ')}`).toBe(true)
    } finally {
      globalThis.fetch = restore
      fontSet.restore()
    }
  })

  // A HEADING IS SUPPRESSED ONLY WHEN ITS OWN GROUP IS EMPTY AFTER FILTERING —
  // never because another group emptied, and never while it still owns a row.
  // STORY 16.9 removed the third group; 'sara' (Sarabun/Sarala) matches
  // neither remaining group — it exists only in the web-tier snapshot this
  // control no longer offers — so filtering it now empties the dropdown
  // entirely rather than isolating a third heading.
  it('suppresses each heading only on its own empty group', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    // A declared chain name that matches no typeface at all: group 1 alone.
    fireEvent.change(combobox, { target: { value: 'heading' } })
    expect(within(dropdown()).getAllByRole('group').map((group) => group.getAttribute('aria-label'))).toEqual(['IN THIS TEMPLATE'])
    // A family on this machine, matching no declared chain: group 2 alone.
    fireEvent.change(combobox, { target: { value: 'lora' } })
    expect(within(dropdown()).getAllByRole('group').map((group) => group.getAttribute('aria-label'))).toEqual(['AVAILABLE LOCALLY'])
    // A query matching neither remaining group empties the dropdown, rather
    // than falling through to a group this control no longer draws.
    fireEvent.change(combobox, { target: { value: 'sara' } })
    expect(within(dropdown()).queryAllByRole('group')).toEqual([])
    expect(screen.getByText('Nothing in this template or on this machine matches "sara".')).toBeInTheDocument()
  })

  // THE KEYBOARD IS LINEAR EVEN THOUGH THE LIST IS NOT. Two groups is two
  // headings interleaved into ONE option sequence, and 8.6's reason for the
  // flat list survives unchanged — which is exactly why the heading
  // interleave had to go: it was the one element of the walk that read a
  // position semantically.
  it('walks both groups in one arrow-key sequence, and wraps', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    const owner = () => document.getElementById(combobox.getAttribute('aria-activedescendant') ?? '')?.closest('[role="group"]')?.getAttribute('aria-label') ?? undefined
    const total = within(dropdown()).getAllByRole('option').length
    expect(total, 'a walk over an empty list would prove nothing').toBeGreaterThan(2)
    const visited: Array<string | undefined> = []
    for (let step = 0; step < total; step += 1) {
      visited.push(owner())
      fireEvent.keyDown(combobox, { key: 'ArrowDown' })
    }
    // ONE CONTIGUOUS RUN PER GROUP, IN THE ORDER THEY ARE DRAWN. Three runs
    // would mean the walk crossed a group boundary and came back.
    expect(visited.filter((label, index) => index === 0 || label !== visited[index - 1])).toEqual(['IN THIS TEMPLATE', 'AVAILABLE LOCALLY'])
    // AND THE SEQUENCE IS ONE SEQUENCE: the step past the last row is the first.
    expect(owner()).toBe('IN THIS TEMPLATE')
    // Backwards from the top lands on the last row of the last group, which is
    // the same walk read the other way.
    fireEvent.keyDown(combobox, { key: 'ArrowUp' })
    expect(owner()).toBe('AVAILABLE LOCALLY')
  })

  // STORY 8.6's DEFERRAL, CLOSED HERE RATHER THAN MULTIPLIED (it would have gone
  // from six non-option children to seven). NOTHING PINNED THE PRESENTATION ROLE
  // BEFORE THIS TEST — measured: zero assertions over the whole test corpus — so
  // the fix would otherwise have been unguarded, and a note dropped back into the
  // list would break the listbox again in silence.
  it('owns only groups of options, with every note outside the list and referenced from it', () => {
    select()
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    const listbox = dropdown()
    expect(listbox.querySelectorAll('[role="presentation"]'), 'a listbox may not own a presentational child').toHaveLength(0)
    expect(Array.from(listbox.children).map((child) => child.getAttribute('role')), 'every child of the listbox is a group').toEqual(['group', 'group'])
    for (const group of within(listbox).getAllByRole('group')) {
      for (const child of Array.from(group.children)) {
        expect(child.getAttribute('role') === 'option' || child.getAttribute('aria-hidden') === 'true', `${group.getAttribute('aria-label')} owns a child that is neither an option nor hidden from the tree: ${child.outerHTML.slice(0, 80)}`).toBe(true)
      }
    }
    // POSITIVE CONTROL, IN THE ONE STATE THAT STILL DRAWS A NOTE. The standing
    // explanation was removed from this dropdown by OWNER decision, so the notes
    // node exists only when the search matches nothing. The rule it was written
    // for is unchanged and is asserted where it can still bite: a note is never a
    // row in the list, and the list names the note that describes it.
    expect(listbox.getAttribute('aria-describedby'), 'with rows to show there is no note to describe them').toBeNull()
    fireEvent.change(screen.getByRole('combobox', { name: 'Font family' }), { target: { value: 'Helvetica' } })
    const empty = screen.getByText(/Nothing in this template or on this machine matches/)
    const emptied = dropdown()
    expect(emptied, 'the empty-state sentence is about the list and is not a row in it').not.toContainElement(empty)
    const notes = document.getElementById(emptied.getAttribute('aria-describedby') ?? '')
    expect(notes, 'the list must name the note that describes it').not.toBeNull()
    expect(notes).toContainElement(empty)
  })

  // MATRIX ROW: TWO COMPONENTS WITH DIFFERENT FAMILIES STILL SAY `Mixed`, AND
  // THE THREE GROUPS DO NOT CHANGE THAT. The row was carried as "as today" and
  // "as today" was asserted by nothing, so a control rebuilt around a partition
  // could have lost it in silence.
  it('keeps the Mixed placeholder over a selection with two different families', () => {
    const two = [
      { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'One', fontFamily: 'body' },
      { id: 'e2', type: 'text' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 24_000, resizable: true, value: 'Two', fontFamily: 'heading' },
    ]
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: two } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByLabelText('text component e2'), { shiftKey: true })
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    expect(combobox).toHaveAttribute('placeholder', 'Mixed')
    expect(combobox).toHaveAttribute('aria-description', 'Mixed value')
    // AND NO ROW CLAIMS TO BE THE SELECTED ONE, because none of them is.
    fireEvent.focus(combobox)
    expect(within(screen.getByRole('listbox', { name: 'Fonts' })).queryAllByRole('option', { selected: true })).toEqual([])
  })

  // STORY 8.6, AC4. THE TWO GROUPS ARE DIFFERENT KINDS OF THING, and the
  // difference is asserted by what each does when it is picked — not by a
  // class name, which would pass on a control where both options committed the
  // same property.
  it('offers the bundled catalogue as a second, visibly distinct group whose entries the document does not declare', () => {
    select()
    const combobox = screen.getByRole('combobox', { name: 'Font family' })
    fireEvent.focus(combobox)
    // A family the document declares no chain for. It is offered, it is marked
    // as an addition rather than a selection, and the headings say which group
    // is which.
    // The accessible name collapses the note's leading space, so the pattern is
    // whitespace-tolerant rather than pinned to one spelling of the gap.
    //
    // STORY 16.1: THE NOTE NOW SAYS WHICH TIER THE ROW COMES FROM. `Inter` is
    // one of the 21 committed faces — the LOCAL FACE TIER — so its row states
    // that nothing is downloaded. A family that exists only in the build-time
    // index snapshot carries the plain note, because picking it fetches.
    const inter = screen.getByRole('option', { name: /^Inter$/ })
    expect(inter).toBeInTheDocument()
    // STORY 16.4 DREW THREE HEADINGS ON THE AXIS `WHERE ARE THE BYTES`; STORY
    // 16.9 NARROWED THE DROPDOWN BACK TO TWO. The two 8.6 shipped — `In this
    // document` and `Catalogue — not yet in this document` — named a WHEN and
    // a source that was never the only one; the group a row sits in is a pure
    // function of (declared?, `familyIsInstalled`?), and each heading is the
    // accessible name of the group that owns those rows. `AVAILABLE TO
    // INSTALL` is gone from this control — `Add fonts…` is its door now.
    expect(screen.getByRole('group', { name: 'IN THIS TEMPLATE' })).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'AVAILABLE LOCALLY' })).toBeInTheDocument()
    expect(screen.queryByRole('group', { name: 'AVAILABLE TO INSTALL' })).not.toBeInTheDocument()
    // RETIRED, AND THE SUBJECT NO LONGER EXISTS: the disclosure stopped rendering
    // in this dropdown (OWNER, 2026-09-03 — the standing explanation above
    // `Add fonts…` was cut), and Story 16.10 then removed its last render site,
    // the font browser's header, because the design draws no paragraph there.
    // `familyIndexDisclosure()` IS DELETED — measured at zero consumers — so
    // there is no sentence left for any surface to assert. The COUNT it quoted
    // survives: `addableFamilyCount`'s own two facts are pinned in
    // `font-index.test.ts`, and `resultLine`, the line that still prints that
    // count in the browser's results toolbar, has its output pinned separately in
    // `font-browser-model.test.ts:285-288`. Two different subjects, two files.
    // THE DISK-FONT DECLINE IS RETIRED FROM THIS SURFACE (OWNER, 2026-09-03).
    // It was re-derived at 16.4 rather than carried, and it was the only place
    // this product answered "where do I add my own font file?" — there is no
    // import control to be found missing. The owner cut the standing explanation
    // above `Add fonts…` and this sentence went with it, so the question now has
    // no answer in the UI. Recorded here rather than left to be rediscovered.
    // The declared group never carries the addition note: picking one of those
    // sets a property, and it is already in the file.
    expect(declaredOptions().map(optionText)).toEqual(['body', 'heading'])
    // A SNAPSHOT-ONLY FAMILY IS NO LONGER REACHABLE BY TYPING HERE AT ALL
    // (Story 16.9). `Kanit` is in the published web-tier snapshot and not
    // among the 31 local/stored faces, so it is not offered by this control
    // under any query — `Add fonts…` is the only door to it now.
    fireEvent.change(combobox, { target: { value: 'Kanit' } })
    expect(screen.queryByRole('option', { name: /^Kanit/ }), 'a family not on this machine is not offered by this dropdown at all').not.toBeInTheDocument()
    expect(screen.getByText('Nothing in this template or on this machine matches "Kanit".')).toBeInTheDocument()
  })

  // STORY 8.6's AC1/AC3 AT THE BROWSER BOUNDARY, BEHAVIOUR-CHANGED BY STORY 16.5
  // INTO THE THIRD ARM'S WITNESS.
  //
  // `Inter` is a LOCAL-TIER face: it ships inside the release, so this machine
  // already holds it and there is nothing to install. Picking it is therefore
  // FIRST USE, and first use is two commands — `embedFontFamily` and then
  // `updateComponentProperties` — with two undo entries. The order is forced by
  // the engine (`canvas.fontFamilies` is the closed set `style.fontFamily` may
  // name), and asserting it ON THE WIRE, in order, is what makes this a claim
  // about behaviour rather than about a class name.
  //
  // What changed is the SECOND command. Under Story 8.6 a catalogue pick
  // deliberately did not set `fontFamily`; under embed-on-use, setting a
  // component's family to a face this machine holds IS the author asking for
  // both, so both happen — as two commands, never fused into one.
  it('embeds and then commits the property, as two commands, when a family this machine holds is picked', async () => {
    const face = new Uint8Array([0x00, 0x01, 0x00, 0x00, 0x7f]).buffer
    const fetchStub = vi.fn(async (_url: string) => ({ ok: true, arrayBuffer: async () => face }))
    const restore = globalThis.fetch
    globalThis.fetch = fetchStub as never
    try {
      // `select`'s mock is declared with no parameters, so its recorded calls
      // are an empty tuple to TypeScript. The arguments are read back through
      // a widened view rather than by changing that shared signature, which
      // every other test in this describe block is written against.
      const request = select()
      const sent = request.mock.calls as unknown as ReadonlyArray<readonly [string, ArrayBuffer]>
      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      fireEvent.click(screen.getByRole('option', { name: /^Inter$/ }))
      await waitFor(() => expect(request).toHaveBeenCalledWith('command', expect.anything()))
      // THE BYTES CAME FROM THE BUNDLE, not from a host. The URL is one of the
      // release's own content-addressed assets, which is what makes the pick
      // work offline — SPEC-fonts requires no call to either Google Fonts host
      // at any point, and `forbidden-font-hosts.test.ts` is what names them.
      //
      // STORY 16.1 MAKES THIS THE LOCAL-TIER WITNESS (D-16.R.3): `Inter` is one
      // of the 21 committed faces, so the pick issues EXACTLY ONE request, to a
      // relative release asset — no `METADATA.pb`, no licence file, no third
      // party. `Inter` is variable-only on the `google/fonts` mirror and static
      // here, which is precisely why the snapshot's `axes` field is not
      // consulted for a family the local tier holds.
      expect(fetchStub).toHaveBeenCalledTimes(1)
      expect(String(fetchStub.mock.calls[0][0])).not.toMatch(/^https?:/)
      // TWO COMMANDS, IN THIS ORDER. Not one fused command, and not the embed
      // alone: the property is committed only after the chain is declared,
      // because the engine refuses it otherwise.
      await waitFor(() => expect(sent).toHaveLength(2))
      const kinds = sent.map(([, buffer]) => (JSON.parse(new TextDecoder().decode(buffer)) as Record<string, unknown>)['kind'])
      expect(kinds).toEqual(['embedFontFamily', 'updateComponentProperties'])
      const property = JSON.parse(new TextDecoder().decode(sent[1][1])) as Record<string, unknown>
      expect(property['changes']).toEqual({ fontFamily: { op: 'set', value: 'Inter' } })
      expect(sent[0][0]).toBe('command')
      const payload = JSON.parse(new TextDecoder().decode(sent[0][1])) as Record<string, unknown>
      expect(payload['kind']).toBe('embedFontFamily')
      expect(payload['name']).toBe('Inter')
      expect(payload['family']).toBe('Inter')
      // The three keys the ENGINE REFUSES TO LOAD A DOCUMENT WITHOUT. A pick
      // that omitted any of them would produce a document the engine's own
      // parser rejects, so their presence is the browser's half of that
      // contract and not a detail.
      expect(payload['licence']).toBeTruthy()
      expect(payload['licenceText']).toBeTruthy()
      expect(payload['copyright']).toBeTruthy()
      expect(payload['data']).toBe('AAEAAH8=')
      // AC3: Inter covers Latin and nothing else, so the proposed tail is the
      // shipped faces for the scripts it does NOT cover, in order.
      expect(payload['tail']).toEqual(['Noto Sans Thai', 'Noto Sans SC'])
    } finally {
      globalThis.fetch = restore
    }
  })

  // RETIRED (Story 16.9): five tests drove `AVAILABLE TO INSTALL` — the
  // dropdown's third group, which no longer exists — by focusing the
  // combobox, typing 'Kanit', and clicking the row labelled 'install on
  // this machine'. That row is gone: 'Add fonts...' is now the only door to
  // a family not on this machine, so the scenarios these tests drove
  // (resolving METADATA.pb into an install; one pick at a time across the
  // web-tier resolution; releasing the pick hold on a mid-resolution
  // document replacement; a directory disagreeing with its declared
  // licence token; installing with no network) can no longer be reached
  // from this control at all. OF THE THREE MECHANISMS THEY EXERCISED, TWO ARE
  // UNIT-TESTED DIRECTLY AND ONE IS NOT, and the difference is now stated
  // instead of averaged over: `fetchWebFamily` / `timedFetcher`'s stall
  // handling is covered in `font-source.test.ts`, and the licence
  // classification in `font-licence.test.ts`.
  //
  // THE ONE-RESOLUTION-AT-A-TIME GUARD IS IN NEITHER FILE, and this comment
  // used to say it was. It cannot be: the guard is not in either module. It is
  // `App.tsx`'s own `fontChainBusyRef` / `holdFontChain` (App.tsx:114-115),
  // read by both surviving doors — `addFamilyToDocument` (App.tsx:885) and
  // `embedInstalledFamily` (App.tsx:1073) — and handed back on a document
  // replacement by `setCurrentSnapshot` rather than by either `finally`, which
  // is generation-guarded and deliberately declines to release a moved hold.
  // Measured against the claim: searching both named files for
  // `at a time|concurrent|overlapping|in flight|busy|re-entran` returns 0 hits,
  // against 29 `fetchWebFamily` hits in `font-source.test.ts` as the positive
  // control that the search itself works.
  //
  // WHERE IT IS COVERED NOW: two tests in `App.font-store.test.tsx`, driven
  // through the surviving `AVAILABLE LOCALLY` first-use door — 'releases the
  // pick hold when the document is replaced mid-resolution, so a later pick
  // still commits' (the flag stranded, so every later pick is dead for the
  // session) and 'refuses a pick made against another component while the
  // first embed is still in flight, and sends one embed' (the flag never
  // taken, so two overlapping picks both resolve and two embeds commit). The
  // deleted row is what went; the mechanism did not.
  //
  // AND ALL THREE ARE STILL REACHABLE END TO END through the font browser's
  // own 'Add fonts...' flow, which this story does not touch.

  it('shows the engine\'s own default size for an element that commits none, as a placeholder and not a value', () => {
    select()
    const size = screen.getByRole('textbox', { name: 'Font size (pt)' })
    expect(size).toHaveValue('')
    expect(size).toHaveAttribute('placeholder', '12')
  })

  it('commits alignment from the closed sets, and clears it by pressing the active segment again', async () => {
    const request = select()
    for (const name of ['Align left', 'Align center', 'Align right', 'Align justify', 'Vertical align top', 'Vertical align middle', 'Vertical align bottom']) {
      expect(screen.getByRole('button', { name })).toHaveAttribute('aria-pressed', 'false')
    }
    fireEvent.click(screen.getByRole('button', { name: 'Align center' }))
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
  })

  // Story 7.4 / AC3. `style.align` admits `justify` for a text element, and
  // has since 7.3; a table's cells draw a justified value at their start
  // edge, so the value is meaningless there and the control must not offer
  // it. A MIXED selection is the case that decides the rule: one command goes
  // to every id, so the segment is offered only when EVERY selected component
  // is text.
  it('offers justify for text alone, and never for a table or a mixed selection', async () => {
    const table = { id: 'e2', type: 'table' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    const componentCanvas = { ...canvas, components: [textComponent, table] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)

    const alignNames = () => within(screen.getByRole('group', { name: 'Align' })).getAllByRole('button').map((button) => button.getAttribute('aria-label'))
    fireEvent.click(screen.getByLabelText('text component e1'))
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right', 'Align justify'])
    // The glyph is an SVG path, never a CSS declaration asking the browser to
    // justify: canvas-authority-contract.test.ts bans that across every
    // production, unit and e2e source.
    expect(screen.getByRole('button', { name: 'Align justify' }).querySelector('svg path')).toBeInTheDocument()

    fireEvent.click(screen.getByLabelText('table component e2'))
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right'])

    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByLabelText('table component e2'), { shiftKey: true })
    expect(alignNames()).toEqual(['Align left', 'Align center', 'Align right'])

    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Align justify' }))
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"align":{"op":"set","value":"justify"}}}')
  })

  // Story 7.4 / AC4. lineSpacing is a dimensionless RATIO carried as a RAW,
  // UNQUOTED JSON number: Go's own decoder performs the x1000 to thousandths,
  // exactly as it does for a value written in a .folio file. Quoting it, or
  // pre-multiplying it here, is refused by the engine.
  it('shows an unset ratio as a placeholder and commits a typed one as a raw unquoted number', async () => {
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const leading = screen.getByRole('textbox', { name: 'Line spacing' })
    // Empty is the leading the declared chain itself rules — a ratio of 1 —
    // shown as a placeholder and never as a value.
    expect(leading).toHaveValue('')
    expect(leading).toHaveAttribute('placeholder', '1')
    expect(leading).toHaveAttribute('inputMode', 'decimal')
    fireEvent.change(leading, { target: { value: '1.5' } })
    fireEvent.keyDown(leading, { key: 'Enter' })
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":1.5}}}')
  })

  it('reads a committed line spacing back in the author\'s units and clears it from the same row', async () => {
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    const spaced = { ...canvas, components: [{ ...textComponent, lineSpacing: 1_500 }] }
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: spaced }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    // The engine carries thousandths; the author is shown the ratio.
    expect(screen.getByRole('textbox', { name: 'Line spacing' })).toHaveValue('1.5')
    fireEvent.click(screen.getByRole('button', { name: 'Clear Line spacing' }))
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"clear"}}}')
  })

  // The mock reproduces the engine's REAL rejection, measured by running the
  // command through Go: applyPropertyChanges prefixes the command key and
  // template's validator words its reason in terms of that same key, so the
  // message the browser receives really does begin `lineSpacing: lineSpacing`.
  // ComponentCommandError carries it verbatim, with DataPath
  // `component.lineSpacing` on element `e1`; the Go half is pinned by
  // TestLineSpacingPropertyCommandDecodesThroughTheOneLoaderValidator.
  it('shows the engine\'s located line-spacing refusal and keeps the author\'s text', async () => {
    const engineMessage = 'lineSpacing: lineSpacing must be between 1 and 1000000 thousandths (0.001 to 1000); 0 is outside that range'
    const request = vi.fn((operation: string) => operation === 'command'
      ? Promise.reject(Object.assign(new Error(engineMessage), { elementId: 'e1', dataPath: 'component.lineSpacing' }))
      : Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3 } }))
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const leading = screen.getByRole('textbox', { name: 'Line spacing' })
    fireEvent.change(leading, { target: { value: '0' } })
    fireEvent.keyDown(leading, { key: 'Enter' })
    // The WHOLE located sentence, not a prefix of it: the element, the field
    // it was located to, and the engine's own reason including the offending
    // value the author typed.
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent(`e1: component.lineSpacing: ${engineMessage}`))
    expect(leading).toHaveValue('0')
    expect(leading).toHaveAttribute('aria-invalid', 'true')
  })

  it('presses the segment the engine has committed, and clears it from the same control', async () => {
    const aligned = { ...canvas, components: [{ ...textComponent, align: 'right' as const, valign: 'middle' as const }] }
    const sent: ArrayBuffer[] = []
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: aligned }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    const right = screen.getByRole('button', { name: 'Align right' })
    expect(right).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Vertical align middle' })).toHaveAttribute('aria-pressed', 'true')
    fireEvent.click(right)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(sent.map((payload) => new TextDecoder().decode(payload)).join('')).toContain('"op":"clear"')
  })
})

// Story 7.4: authoring body text in the designer. The CONTENT control was an
// <input type="text">, which cannot hold a line feed at all, so a
// multi-paragraph clause could not be typed OR pasted — the story's first AC
// in one sentence.
describe('Story 7.4: authoring a multi-paragraph clause', () => {
  const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
  const openEditor = (sent: ArrayBuffer[]) => {
    const request = vi.fn(async (_operation: string, payload?: ArrayBuffer) => { if (payload) sent.push(payload); return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } } })
    render(<App engine={engine(request as never)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } }} />)
    fireEvent.click(screen.getByLabelText('text component e1'))
    return screen.getByRole('textbox', { name: 'Text' })
  }

  it('is a textarea, so Enter inserts a paragraph break instead of committing', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    expect(field.tagName).toBe('TEXTAREA')
    // Three paragraphs, typed. Enter commits nothing; blur sends ONE command
    // carrying the whole value with its line feeds intact.
    fireEvent.change(field, { target: { value: 'First clause.\nSecond clause.\nThird clause.' } })
    fireEvent.keyDown(field, { key: 'Enter' })
    expect(sent).toHaveLength(0)
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(new TextDecoder().decode(sent[0]!)).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"First clause.\\nSecond clause.\\nThird clause."}}}')
  })

  it('commits a forty-page clause as one command, with every paragraph break intact', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    // Well past the 512 BYTES the projection used to cap an element's value
    // at — about eighty English words, less than one numbered clause — which
    // did not merely blank the canvas but REJECTED the edit, because the
    // property command re-projects inside its own transaction.
    const clause = Array.from({ length: 1_900 }, (_value, index) => `Clause ${index + 1}. The parties agree as set out above.`).join('\n')
    expect(new TextEncoder().encode(clause).byteLength).toBeGreaterThan(512)
    fireEvent.change(field, { target: { value: clause } })
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    const command = new TextDecoder().decode(sent[0]!)
    expect(JSON.parse(command).changes.value.value).toBe(clause)
    expect(sent).toHaveLength(1)
  })

  it('takes only the plain flavour of a word-processor paste, keeping paragraph breaks and dropping the formatting', async () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: '' } })
    const flavours: Record<string, string> = {
      'text/plain': 'Clause 1.\nClause 2.',
      'text/html': '<p style="font-weight:700;font-family:Georgia">Clause 1.</p><p><em>Clause 2.</em></p>',
      'text/rtf': '{\\rtf1 \\b Clause 1.\\par}',
    }
    const read: string[] = []
    fireEvent.paste(field, { clipboardData: { getData: (flavour: string) => { read.push(flavour); return flavours[flavour] ?? '' } } })
    // The formatting is discarded by never being looked at: no sanitiser, no
    // new dependency, nothing to go wrong on an unusual clipboard.
    expect(read).toEqual(['text/plain'])
    expect(field).toHaveValue('Clause 1.\nClause 2.')
    fireEvent.blur(field)
    await waitFor(() => expect(sent).toHaveLength(1))
    expect(JSON.parse(new TextDecoder().decode(sent[0]!)).changes.value.value).toBe('Clause 1.\nClause 2.')
  })

  it('inserts nothing when the clipboard has no plain flavour, instead of letting the browser paste the HTML', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: 'Clause 1.' } })
    const read: string[] = []
    const flavours: Record<string, string> = {
      'text/html': '<p style="font-weight:700;font-family:Georgia">Pasted from a word processor</p>',
      'text/rtf': '{\\rtf1 \\b Pasted from a word processor\\par}',
    }
    // fireEvent returns false when the event was cancelled, which is the only
    // way to see from here that the BROWSER's own paste was refused. Without
    // preventDefault the browser inserts text it derives from the text/html
    // flavour — formatting laundered in through the door the story closed.
    const dispatched = fireEvent.paste(field, { clipboardData: { getData: (flavour: string) => { read.push(flavour); return flavours[flavour] ?? '' } } })
    expect(dispatched).toBe(false)
    expect(read).toEqual(['text/plain'])
    expect(field).toHaveValue('Clause 1.')
    expect(sent).toHaveLength(0)
  })

  it('leaves the caret at the end of the pasted text, not at the end of the field', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent) as HTMLTextAreaElement
    fireEvent.change(field, { target: { value: 'Clause 1.\nClause 3.' } })
    // The caret sits at the start of the LAST paragraph, and a clause is
    // pasted in front of it. A controlled textarea re-rendered with a new
    // value drops the caret at the very end unless it is restored, which would
    // put the author's next keystroke in the wrong paragraph.
    field.setSelectionRange(10, 10)
    fireEvent.paste(field, { clipboardData: { getData: () => 'Clause 2.\n' } })
    expect(field).toHaveValue('Clause 1.\nClause 2.\nClause 3.')
    expect(field.selectionStart).toBe(20)
    expect(field.selectionEnd).toBe(20)
  })

  it('reverts and blurs on Escape, exactly as every single-line field does', () => {
    const sent: ArrayBuffer[] = []
    const field = openEditor(sent)
    fireEvent.change(field, { target: { value: 'a draft\nnobody wants' } })
    fireEvent.keyDown(field, { key: 'Escape' })
    expect(field).toHaveValue('Hello')
    expect(sent).toHaveLength(0)
  })

  it('paints one canvas line per engine line and says, in words, when the paint is only a prefix', () => {
    const line = (top: number, text: string) => ({ top, baseline: top + 10_000, advance: 14_000, width: 30_000, fragments: [{ text, x: 0 }] })
    const cut = { ...canvas, components: [{ ...textComponent, textPaint: { overflow: false, truncated: true, lines: [line(0, 'First clause.'), line(14_000, 'Second clause.')] } }] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: cut }} />)
    expect(document.querySelectorAll('.canvas-text-line')).toHaveLength(2)
    // Stated in words at the component, and in the same sentence a screen
    // reader gets — never by colour, and never only by a CSS class.
    const notice = 'Canvas preview cut short. The whole text is in the document and prints in full.'
    expect(screen.getByText(notice)).toBeInTheDocument()
    expect(screen.getByLabelText(`text component e1: First clause. Second clause.; ${notice}`)).toBeInTheDocument()
  })

  it('says nothing about truncation for an untruncated paint or an empty one', () => {
    const notice = 'Canvas preview cut short. The whole text is in the document and prints in full.'
    const whole = { ...canvas, components: [
      { ...textComponent, textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 10_000, advance: 14_000, width: 30_000, fragments: [{ text: 'First clause.', x: 0 }] }] } },
      { id: 'e2', type: 'text' as const, band: 'content' as const, x: 0, y: 30_000, width: 72_000, height: 24_000, resizable: true, textPaint: { overflow: false, truncated: false, lines: [] } },
    ] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: whole }} />)
    expect(screen.queryByText(notice)).not.toBeInTheDocument()
  })
})

describe('Story 5.13: image asset selection', () => {
  const imageComponent = { id: 'e1', type: 'image' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 48_000, resizable: true, image: { mediaType: 'image/png', assetKey: 'a'.repeat(64), width: 300, height: 200, drawX: 6_000, drawY: 8_000, drawWidth: 60_000, drawHeight: 40_000 } }
  const undecodableImageComponent = { id: 'e2', type: 'image' as const, band: 'content' as const, x: 0, y: 60_000, width: 72_000, height: 48_000, resizable: true, imageUnavailable: 'undecodable' as const }
  const textComponent = { id: 'e3', type: 'text' as const, band: 'content' as const, x: 0, y: 120_000, width: 72_000, height: 24_000, resizable: true }
  const imageFileAccess = (openImage: () => Promise<{ bytes: ArrayBuffer; mediaType: string; name: string }>) => ({ openImage }) as unknown as import('./image-file').ImageFileAccess

  it('shows the IMAGE section carrying the engine snapshot identity for a single image selection, and never for other selections', () => {
    const componentCanvas = { ...canvas, components: [imageComponent, textComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    // Empty selection: no IMAGE section.
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
    // Single image selection: IMAGE section with identity from the snapshot.
    fireEvent.click(screen.getByLabelText('image component e1'))
    expect(screen.getByText('IMAGE')).toBeInTheDocument()
    expect(screen.getByText('image/png · 300×200px · asset aaaaaaaaaaaa…')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Choose image…' })).toBeInTheDocument()
    // Mixed selection (image + text): no IMAGE section.
    fireEvent.click(screen.getByLabelText('text component e3'), { shiftKey: true })
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
    // Non-image single selection: no IMAGE section.
    fireEvent.click(screen.getByLabelText('text component e3'))
    expect(screen.queryByText('IMAGE')).not.toBeInTheDocument()
  })

  it('states the concrete reason for an asset this version cannot render, distinguished by text, not colour alone', () => {
    const componentCanvas = { ...canvas, components: [undecodableImageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    fireEvent.click(screen.getByLabelText('image component e2'))
    expect(screen.getByText("This version cannot render this asset's media type.")).toBeInTheDocument()
  })

  it('shows an empty, choosable box for a placed image with no file yet', () => {
    // Go projects neither a paint nor an unavailable reason for a box the
    // author has not filled: nothing to draw, and nothing wrong.
    const emptyImageComponent = { id: 'e2', type: 'image' as const, band: 'content' as const, x: 0, y: 60_000, width: 72_000, height: 48_000, resizable: true }
    const componentCanvas = { ...canvas, components: [emptyImageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'logo.png' }))} />)
    expect(screen.getByText('No image')).toBeInTheDocument()
    expect(screen.queryByText(/Image unavailable/)).not.toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('image component e2'))
    expect(screen.getByText(/No image chosen yet/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Choose image…' })).toBeEnabled()
  })

  it('states the concrete reason when no local picker capability is available in this browser tier', () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    const button = screen.getByRole('button', { name: 'Choose image…' })
    expect(button).toBeDisabled()
    expect(screen.getByText('No local file picker is available in this browser tier.')).toBeInTheDocument()
  })

  it('drives the keyboard path from a real keyboard SELECTION to a focus-visible picker control, then commits through it', async () => {
    // Finding 15 (review of 2026-08-29): the original version of this test
    // selected the component with fireEvent.click (a mouse-shaped
    // interaction) and only PROVED the picker button was focusable, never
    // dispatching a key event at all. CanvasComponent has its own
    // onKeyDown handler for Enter/Space selection (App.tsx) — exercise
    // THAT, not a click, for the selection half of "selection to picker".
    const componentCanvas = { ...canvas, components: [imageComponent] }
    const openImage = vi.fn(async () => ({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' }))
    const request = vi.fn(async (operation: string, _payload?: ArrayBuffer) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3, canvas: componentCanvas } }))
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(openImage)} />)
    const component = screen.getByLabelText('image component e1')
    component.focus()
    fireEvent.keyDown(component, { key: 'Enter' })
    expect(screen.getByText('IMAGE')).toBeInTheDocument()
    const button = screen.getByRole('button', { name: 'Choose image…' })
    // "Keyboard-reachable... with visible colors.select focus" (AC2):
    // proved by moving focus WITHOUT a pointer event and checking it
    // landed. A native <button> converts Enter/Space into a real 'click'
    // event by construction in every browser — jsdom does not synthesise
    // that translation from a bare keyDown, so the committed activation
    // below stands in for it here; the Playwright suite drives the SAME
    // control with a real OS-level key press (image-asset.spec.ts).
    button.focus()
    expect(document.activeElement).toBe(button)
    fireEvent.click(button)
    await waitFor(() => expect(openImage).toHaveBeenCalledOnce())
    await waitFor(() => expect(request.mock.calls.some(([operation]) => operation === 'command')).toBe(true))
    const [, payload] = request.mock.calls.find(([operation]) => operation === 'command')!
    const command = new TextDecoder().decode(payload)
    expect(command).toContain('"kind":"setComponentAsset"')
    expect(command).toContain('"id":"e1"')
    expect(command).toContain('"mediaType":"image/jpeg"')
  })

  it('shows a located diagnostic when the command rejects the picked file, and shows nothing when the picker is cancelled', async () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    const request = vi.fn(async (operation: string) => operation === 'command' ? Promise.reject(Object.assign(new Error('asset exceeds the 8388608-byte supported size'), { elementId: 'e1', dataPath: 'component.data' })) : { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: componentCanvas } })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => ({ bytes, mediaType: 'image/png', name: 'huge.png' }))} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('e1: asset exceeds the 8388608-byte supported size'))
  })

  it('shows no error when the local picker is cancelled', async () => {
    const componentCanvas = { ...canvas, components: [imageComponent] }
    render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(async () => { throw new FileAccessCancelled() })} />)
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Choose image…' })).not.toBeDisabled())
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('does not install a setComponentAsset result that resolves after a document replacement (Finding 4)', async () => {
    // AC1's own named red proof: "a command result installed after
    // document replacement". Element ids are reused across documents
    // (e1, e2, ...), and this closure spans the two longest awaits in the
    // app — an OS file dialog, then an engine command carrying up to
    // megabytes — so if Open/Start blank/undo lands in between, a stale
    // command result must never overwrite the newer, authoritative
    // document. Before the fix, applyImageAsset called setCurrentSnapshot
    // unconditionally with no generation/revision guard, matching every
    // OTHER committed-command path's (bindPickedPath, etc.) shape only in
    // that one respect being ABSENT.
    const componentCanvas = { ...canvas, components: [imageComponent] }
    let resolveOpenImage: ((value: { bytes: ArrayBuffer; mediaType: string; name: string }) => void) | undefined
    const openImage = vi.fn(() => new Promise<{ bytes: ArrayBuffer; mediaType: string; name: string }>((resolve) => { resolveOpenImage = resolve }))
    const request = vi.fn(async (operation: string) => {
      if (operation === 'load') return { snapshot: { documentState: 'loaded' as const, revision: 50, byteLength: 3 } }
      if (operation === 'command') return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: componentCanvas } }
      return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: componentCanvas } }
    })
    render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} imageFileAccess={imageFileAccess(openImage)} blankBytes={bytes} />)

    // Start the asset pick — this awaits openImage(), which we hold open.
    fireEvent.click(screen.getByLabelText('image component e1'))
    fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
    await waitFor(() => expect(openImage).toHaveBeenCalledOnce())

    // A DOCUMENT REPLACEMENT lands while the picker is still open.
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())
    await waitFor(() => expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 50'))

    // NOW the picker resolves and the stale command completes.
    resolveOpenImage!({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' })
    await waitFor(() => expect(request.mock.calls.some(([operation]) => operation === 'command')).toBe(true))

    // The blank document (revision 50) must still be showing — the stale
    // command's revision-2 result must never have been installed.
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent('GO SNAPSHOT · REVISION 50')
    expect(screen.getByText('Untitled template')).toBeInTheDocument()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('paints the image inside the Go-owned draw rectangle at zoom, fetched per asset key, shows an honest placeholder for an undecodable asset, and revokes its object URL on every trigger AC3 names', async () => {
    let paintCounter = 0
    const createObjectURL = vi.fn(() => `blob:paint-${++paintCounter}`)
    const revokeObjectURL = vi.fn()
    const priorCreate = URL.createObjectURL; const priorRevoke = URL.revokeObjectURL
    ;(URL as unknown as { createObjectURL: typeof createObjectURL }).createObjectURL = createObjectURL
    ;(URL as unknown as { revokeObjectURL: typeof revokeObjectURL }).revokeObjectURL = revokeObjectURL
    try {
      const canvasWithKey = (assetKey: string, extra: ReadonlyArray<typeof undecodableImageComponent> = [undecodableImageComponent]) =>
        ({ ...canvas, components: [{ ...imageComponent, image: { ...imageComponent.image, assetKey } }, ...extra] })
      const canvasA = canvasWithKey('a'.repeat(64))
      // Finding 11: asset REPLACEMENT (same element id, new assetKey) is a
      // distinct trigger from document replacement — exercised via a
      // committed setComponentAsset, never a document generation bump.
      const canvasB = canvasWithKey('b'.repeat(64))
      const openImage = vi.fn(async () => ({ bytes, mediaType: 'image/jpeg', name: 'logo.jpg' }))
      const request = vi.fn(async (operation: string) => {
        if (operation === 'asset') return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3 }, bytes }
        if (operation === 'command') return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: canvasB } }
        // Finding 11: document REPLACEMENT (Start blank) with the SAME
        // assetKey as canvasB — isolates the `generation` dependency from
        // `assetKey`, so a future edit that dropped `generation` from
        // ImagePaint's effect deps would leak here with nothing else red.
        if (operation === 'load') return { snapshot: { documentState: 'loaded' as const, revision: 9, byteLength: 3, canvas: canvasB } }
        return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: canvasA } }
      })
      const view = render(<App engine={engine(request)} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: canvasA }} imageFileAccess={imageFileAccess(openImage)} blankBytes={bytes} />)
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledOnce())
      const img = () => view.container.querySelector('img.canvas-image-paint') as HTMLImageElement
      expect(img().src).toContain('blob:paint-1')
      // The undecodable second element paints an honest, named placeholder,
      // never a blank box and never a crash.
      expect(screen.getByText(/Image unavailable/)).toBeInTheDocument()

      // Finding 3: the painted element's geometry must equal the engine's
      // OWN draw rectangle (image.drawX/Y/W/H relative to component.x/y),
      // mapped through canvasDisplay's zoom rule — never object-fit, never
      // a browser-computed fit. Checked at the default zoom (1) first.
      expect(img().style.left).toBe('6px'); expect(img().style.top).toBe('8px')
      expect(img().style.width).toBe('60px'); expect(img().style.height).toBe('40px')

      // Finding 3/10: AC3 says the rectangle is "mapped through the
      // existing zoom rule" — this was asserted at NO zoom before. One
      // step of "Zoom in" (+0.1) must scale every one of the four values
      // by exactly the same factor the engine's own zoom rule uses.
      fireEvent.click(screen.getByRole('button', { name: 'Zoom in' }))
      await waitFor(() => expect(screen.getByLabelText('Canvas zoom')).toHaveTextContent('110%'))
      expect(img().style.left).toBe('6.6px'); expect(img().style.top).toBe('8.8px')
      expect(img().style.width).toBe('66px'); expect(img().style.height).toBe('44px')

      // Finding 11, trigger 1 of 3: ASSET REPLACEMENT. Pick a new file for
      // the same element; the committed command repoints its assetKey, and
      // that alone (not a document generation bump) must revoke the first
      // URL and fetch a second.
      fireEvent.click(screen.getByLabelText('image component e1'))
      fireEvent.click(screen.getByRole('button', { name: 'Choose image…' }))
      await waitFor(() => expect(openImage).toHaveBeenCalledOnce())
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledTimes(2))
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-1')
      expect(img().src).toContain('blob:paint-2')

      // Finding 11, trigger 2 of 3: DOCUMENT REPLACEMENT. Start blank loads
      // a canvas whose e1 element carries the SAME assetKey as canvasB —
      // only `generation` changed, isolating it from the assetKey trigger
      // just exercised above.
      fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
      await waitFor(() => expect(createObjectURL).toHaveBeenCalledTimes(3))
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-2')
      expect(img().src).toContain('blob:paint-3')

      // Finding 11, trigger 3 of 3: DELETION (unmount) revokes the URL this
      // effect most recently created — no accumulation across a session.
      view.unmount()
      expect(revokeObjectURL).toHaveBeenCalledWith('blob:paint-3')
    } finally {
      (URL as unknown as { createObjectURL: typeof priorCreate }).createObjectURL = priorCreate
      ;(URL as unknown as { revokeObjectURL: typeof priorRevoke }).revokeObjectURL = priorRevoke
    }
  })
})

// STORY 7.6 — THE CANVAS DRAWS EVERY PAGE THE DOCUMENT WILL PRODUCE.
//
// Every projection below is a fixture: the sheets, the seams and the
// disclosures are read from `contentWindowOrigins` and
// `contentWindowCountIsExact`, so a test that changed the origins and saw the
// same drawing would be a test of nothing.
describe('canvas sheet stack', () => {
  const origins = [0, 700_000, 1_400_000]
  const threeWindows = { ...canvas, contentWindowCount: 3, contentWindowOrigins: origins }
  const header = { id: 'h1', type: 'text' as const, band: 'pageHeader' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: true }
  const footer = { id: 'f1', type: 'text' as const, band: 'pageFooter' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: true }
  const at = (id: string, y: number, height = 24_000) => ({ id, type: 'text' as const, band: 'content' as const, x: 0, y, width: 72_000, height, resizable: true })
  const snapshotOf = (projection: CanvasProjection) => ({ documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: projection })
  const sheetLabels = () => Array.from(document.querySelectorAll('.page-surface')).map((surface) => surface.getAttribute('aria-label'))

  it('draws one sheet per projected window, in order, repeating the two bands the engine repeats', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [header, at('e1', 0), footer] })} />)
    expect(sheetLabels()).toEqual([
      'Report page 1 of 3 with Page Header, Content, and Page Footer',
      'Report page 2 of 3 with Page Header, Content, and Page Footer',
      'Report page 3 of 3 with Page Header, Content, and Page Footer',
    ])
    // Each sheet carries all three bands, and the repeating bands carry their
    // components — because the engine repeats them onto every printed page.
    expect(document.querySelectorAll('.page-band-pageHeader')).toHaveLength(3)
    expect(document.querySelectorAll('.page-band-pageFooter')).toHaveLength(3)
    expect(document.querySelectorAll('.page-band-content')).toHaveLength(3)
    // One accessible name each, though: a repeated component is one
    // component, and two identical names would make selection ambiguous.
    expect(screen.getByLabelText('text component h1')).toBeInTheDocument()
    expect(screen.getByLabelText('text component f1')).toBeInTheDocument()
    expect(document.querySelectorAll('.canvas-component-echo')).toHaveLength(4)
  })

  it('marks the seam at the projected origin of the NEXT window, and nowhere when that is past the foot', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf(threeWindows)} />)
    const seams = Array.from(document.querySelectorAll('.page-seam')).map((seam) => (seam as HTMLElement).style.getPropertyValue('--seam-display-y'))
    // origins[1] − origins[0] and origins[2] − origins[1], at zoom 1. The
    // last sheet has no next window, so it draws no marker.
    expect(seams).toEqual(['700px', '700px'])
    // RED PROOF, run and recorded: substituting the window height multiplied
    // by an index for the projected origin gives 729.89px here — the wrong
    // place by a tenth of an inch, and by nine whole sheets on a column with
    // a declared gap. The origins move the marker; nothing else does.
    expect(seams).not.toContain(`${canvas.contentWindowHeight / 1000}px`)
  })

  it('draws no in-sheet seam where the next window begins past the sheet own foot', () => {
    // A declared gap: the next window begins far below this sheet, so the
    // band's own foot IS the boundary and the skipped column region is drawn
    // by nobody.
    const declaredGap = { ...canvas, contentWindowCount: 2, contentWindowOrigins: [0, 7_280_000] }
    render(<App engine={engine()} initialSnapshot={snapshotOf(declaredGap)} />)
    expect(document.querySelectorAll('.page-surface')).toHaveLength(2)
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
  })

  it('draws a component that crosses a seam on both windows, with one interactive home', () => {
    const spanning = at('e2', 650_000, 100_000)
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [spanning] })} />)
    // getByLabelText throws on more than one match, so this assertion IS the
    // uniqueness claim.
    const home = screen.getByLabelText('text component e2')
    expect(home.style.getPropertyValue('--component-y')).toBe('650px')
    const echoes = Array.from(document.querySelectorAll('.canvas-component-echo')) as HTMLElement[]
    expect(echoes).toHaveLength(1)
    // The echo is positioned at the component's column offset MINUS the
    // window it is drawn on, so the run reads as continuous across the seam.
    expect(echoes[0]?.style.getPropertyValue('--component-y')).toBe('-50px')
    expect(echoes[0]?.getAttribute('aria-hidden')).toBe('true')
    expect(echoes[0]?.getAttribute('role')).toBeNull()
  })

  it('sends a later-sheet placement as the band-aware createComponent command, carrying a COLUMN coordinate', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content on page 3 of 3'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // 1400pt is contentWindowOrigins[2] in points — a position in the column,
    // never a pin to sheet three. Go's hitTestBand rectangle is one page tall
    // and is not moved: this routes around it rather than through it.
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"createComponent","version":1,"type":"text","band":"content","x":0,"y":1400,"width":72,"height":24,"snap":true}')
  })

  it('sends a later-sheet POINTER placement through the same column translation as the keyboard one', async () => {
    // The pointer branch and the keyboard branch are two different expressions
    // on the same handler, and only the keyboard one was exercised: replacing
    // the pointer branch with `placeInBand(band.name, point.x, point.y)` — no
    // column origin, page-absolute x — left the whole designer suite green
    // while a mouse-dropped component silently landed on sheet one.
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    const band = screen.getByLabelText('Content on page 3 of 3')
    const localX = ['offset', 'X'].join('')
    const localY = ['offset', 'Y'].join('')
    // jsdom exposes these two as prototype getters that always answer 0, so a
    // plain fireEvent property bag is silently dropped; they have to be
    // defined on the native event the handler actually reads.
    const released = createEvent.pointerUp(band)
    Object.defineProperty(released, localX, { value: 120 })
    Object.defineProperty(released, localY, { value: 40 })
    fireEvent(band, released)
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // x is band-relative (156 page-absolute less the band's own 36), and y is
    // the COLUMN offset: contentWindowOrigins[2] of 1400pt plus the 40pt the
    // pointer sat below the band's head.
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"createComponent","version":1,"type":"text","band":"content","x":120,"y":1440,"width":72,"height":24,"snap":true}')
  })

  it('keeps the FIRST sheet on today dropComponent payload even when the stack is deep', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf(threeWindows)} />)
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content on page 1 of 3'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })

  it('drags a component across a seam onto a later sheet and commits a column coordinate', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0)] })} />)
    const component = screen.getByLabelText('text component e1')
    // One whole sheet-plus-gap down the stack: 841.890pt of page and the 24px
    // gap the stack declares.
    fireEvent.pointerDown(component, { pointerId: 1, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(component, { pointerId: 1, clientX: 10, clientY: 875.89 })
    // Tracking the hand means landing one WINDOW lower in the column — 700pt,
    // the projected origin of window two — not 865.89pt, which is what the
    // linear pixel delta this replaced would have proposed. The difference is
    // exactly the page footer, the gap and the page header the pointer
    // crossed.
    expect(component.style.getPropertyValue('--component-y')).toBe('700px')
    // The clip has to lift for the duration of the gesture, or the component
    // is clipped out of view at the very seam it is being dragged across.
    // jsdom applies no stylesheet, so the class is the only observable the
    // suite has: pinning it to the constant `band-window` left every numeric
    // assertion here green while the gesture was visually broken.
    expect(document.querySelectorAll('.band-window-open').length).toBeGreaterThan(0)
    fireEvent.pointerUp(component, { pointerId: 1, clientX: 10, clientY: 875.89 })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    // And it drops again once the gesture has settled, so the clip is back for
    // every component that is not being dragged.
    await waitFor(() => expect(document.querySelectorAll('.band-window-open')).toHaveLength(0))
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"moveComponent","version":1,"id":"e1","x":0,"y":700,"snap":true}')
  })

  it('lets a drag leave the content band vertically while still capping the repeating bands', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0), footer] })} />)
    const content = screen.getByLabelText('text component e1')
    fireEvent.pointerDown(content, { pointerId: 1, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(content, { pointerId: 1, clientX: 0, clientY: 800 })
    // 800pt down a 729.89pt window. RED PROOF, run and recorded: restoring
    // the clamp DW-36 lifted pins this at 705.89px — the band's own height
    // less the component's — which is the whole of why the lifted column was
    // reachable by command and not by hand.
    expect(content.style.getPropertyValue('--component-y')).toBe('800px')
    const repeated = screen.getByLabelText('text component f1')
    fireEvent.pointerDown(repeated, { pointerId: 2, clientX: 0, clientY: 0 })
    fireEvent.pointerMove(repeated, { pointerId: 2, clientX: 0, clientY: 400 })
    // The page footer is 20pt tall and the component is 12pt tall, so the
    // drag rests at the band foot exactly as it did before this story.
    expect(repeated.style.getPropertyValue('--component-y')).toBe('8px')
  })

  it('states what the sheets claim, in accessible text, whenever there is more than one or the count is not exact', () => {
    const claim = "A component's page is a consequence of the content above it and can change when the data does — it is a column position, not a pin to page three."
    const inexact = 'This count depends on data the canvas does not have, so the printed document can run to a different number of pages.'
    render(<App engine={engine()} initialSnapshot={snapshotOf(threeWindows)} />)
    const disclosure = screen.getByRole('status', { name: 'Canvas sheet disclosure' })
    expect(disclosure).toHaveTextContent('Showing 3 sheets.')
    expect(disclosure).toHaveTextContent('These are the pages this content column occupies as the canvas has laid it out, not a prediction of the printed document.')
    expect(disclosure).toHaveTextContent(claim)
    // Exact, so it must NOT disclaim: a disclosure that always said
    // everything would say nothing.
    expect(disclosure).not.toHaveTextContent(inexact)
  })

  it('says a data-length document can print a different number of pages than are drawn, even when it draws only one sheet', () => {
    // DIRECTION-FREE, deliberately. The old sentence here promised MORE
    // pages, which is true of this bound table and false of an element the
    // data hides — and one document can carry both.
    const inexact = 'This count depends on data the canvas does not have, so the printed document can run to a different number of pages.'
    // The shipped statement shape: four byte-identical templates that print
    // one, five, twenty and fifty pages and project ONE window each, because
    // the canvas has never been given the data.
    const bound = { ...canvas, contentWindowCountIsExact: false, components: [{ id: 'e8', type: 'table' as const, band: 'content' as const, x: 0, y: 54_000, width: 400_000, height: 28_000, resizable: false, tableBind: 'transactions[]' }] }
    render(<App engine={engine()} initialSnapshot={snapshotOf(bound)} />)
    const disclosure = screen.getByRole('status', { name: 'Canvas sheet disclosure' })
    expect(disclosure).toHaveTextContent('Showing 1 sheet.')
    expect(disclosure).toHaveTextContent(inexact)
    // One window, so no seam and no page qualifier — the honesty is in the
    // words, not in a drawing that would be a second lie.
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
    expect(sheetLabels()).toEqual(['Report page with Page Header, Content, and Page Footer'])
  })

  it('carries the same claim into the accessible name of a component on a later sheet, and not onto a first-sheet one', () => {
    render(<App engine={engine()} initialSnapshot={snapshotOf({ ...threeWindows, components: [at('e1', 0), at('e3', 1_450_000)] })} />)
    // The exact sentence, so deleting it turns this red rather than merely
    // shortening a string nobody asserted.
    expect(screen.getByLabelText('text component e3; on canvas page 3 of 3, which is a consequence of the content above it and can change when the data does — a column position, not a pin to page 3')).toBeInTheDocument()
    // The first sheet's component keeps the name it had before this story.
    expect(screen.getByLabelText('text component e1')).toBeInTheDocument()
  })

  it('draws the first budgeted sheets, says it is showing the first N of M, and never blanks', () => {
    const many = MAX_CANVAS_SHEETS + 5
    const budgeted = { ...canvas, contentWindowCount: many, contentWindowOrigins: Array.from({ length: many }, (_value, index) => index * 700_000) }
    render(<App engine={engine()} initialSnapshot={snapshotOf(budgeted)} />)
    expect(document.querySelectorAll('.page-surface')).toHaveLength(MAX_CANVAS_SHEETS)
    expect(screen.getByRole('status', { name: 'Canvas sheet disclosure' })).toHaveTextContent(`Showing the first ${MAX_CANVAS_SHEETS} sheets of ${many}.`)
    expect(screen.queryByText('Waiting for Go page geometry.')).not.toBeInTheDocument()
  })

  // AC5, ASSERTED RATHER THAN ASSUMED. A template whose content column
  // occupies one window and whose content band holds nothing whose length
  // comes from data must render what it rendered at 2c5cfa1: the same DOM,
  // the same accessible names, the same command payload. The compile-only e2e
  // specs address these labels in Playwright STRICT MODE, where a duplicate
  // or a renamed label would be caught by nothing executable.
  it('renders a genuinely single-page template exactly as it did before this story', async () => {
    const request = vi.fn(async () => ({ snapshot: snapshot(2) }))
    render(<App engine={engine(request)} initialSnapshot={snapshotOf({ ...canvas, components: [at('e1', 0)] })} />)
    expect(sheetLabels()).toEqual(['Report page with Page Header, Content, and Page Footer'])
    expect(screen.getByLabelText('Page Header')).toBeInTheDocument()
    expect(screen.getByLabelText('Content')).toBeInTheDocument()
    expect(screen.getByLabelText('Page Footer')).toBeInTheDocument()
    expect(screen.getByLabelText('text component e1')).toBeInTheDocument()
    // No seam, no stack wrapper, no clipping window, no echo, no disclosure.
    expect(document.querySelectorAll('.page-seam')).toHaveLength(0)
    expect(document.querySelectorAll('.sheet-stack')).toHaveLength(0)
    expect(document.querySelectorAll('.band-window')).toHaveLength(0)
    expect(document.querySelectorAll('.canvas-component-echo')).toHaveLength(0)
    expect(screen.queryByRole('status', { name: 'Canvas sheet disclosure' })).not.toBeInTheDocument()
    // And the payload, byte for byte.
    fireEvent.click(screen.getByRole('button', { name: 'Place Text' }))
    fireEvent.keyDown(screen.getByLabelText('Content'), { key: 'Enter' })
    await waitFor(() => expect(request).toHaveBeenCalledOnce())
    expect(new TextDecoder().decode((request.mock.calls[0] as unknown as [string, ArrayBuffer])[1])).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })
})

// RETIRED (Story 16.9): this file carried a ~560-line describe block,
// 'the font chain editor, where fonts are chosen', covering every control
// FontChainEditor.tsx drew — rename, delete, add/move/remove entry, the
// empty-document state, undo/redo interaction, and the disclosure button
// beside the family combobox that revealed it. `FontChainEditor.tsx` is
// deleted along with its render site, so none of those controls exist to
// assert any more. THE DATA THE EDITOR EDITED IS UNTOUCHED: a document's
// `fonts` map still carries its fallback chains exactly as before — Thai
// and CJK text still renders through them — and `engine-protocol.ts` still
// names the six chain commands the editor used to issue; only the UI that
// issued them is gone. Chain-command construction itself is still tested
// directly in `font-chain-command.test.ts`, independent of any UI.
