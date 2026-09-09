import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App, { PROSE_COMMIT_DEBOUNCE_MS } from './App'
import { DataPanel } from './DataPanel'
import { acceptSampleData } from './sample-data'
import type { EngineClient } from './engine-client'
import { FileAccessCancelled } from './file/file-access'
import type { SampleFileAccess } from './sample-file'
import { PDF_FIXTURE_DIGEST, RENDER_ELAPSED_MS, RENDER_ENGINE_VERSION } from './test/pdf-fixture'

// face() builds the PROJECTED shape of a named-face chain entry (Story 8.3:
// an entry is a discriminated object, not a string). A named face carries no
// family and no style — its name is its identity.
const face = (name: string) => ({ face: name, assetKey: '', family: '', style: '', bold: '', italic: '', boldItalic: '' })

vi.mock('./preview/pdf-viewer', () => ({
  initialPDFPreviewViewState: { page: 1, scale: 1, ['scroll' + 'Top']: 0, ['scroll' + 'Left']: 0 },
  samePDFPreviewViewState: () => true,
  PDFPreviewViewer: () => null,
}))

const sampleBytes = new TextEncoder().encode('{"customer":{"name":"Ada"},"items":[{"sku":"A-1"}]}').buffer
const replacementBytes = new TextEncoder().encode('{"report":{"id":2}}').buffer
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, locale: 'en' as const, utcOffset: '+00:00', marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }], defaultFontSize: 12000, defaultLineSpacing: 1000, contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const snapshot = { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas }

const textCanvas = { ...canvas, components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Text' }] }
// STORY 14.4 / AC2. A canvas whose one component CANNOT receive a scalar
// binding, and a mixed one, so the fifth ladder arm can be exercised against a
// real selection rather than against a prop set by hand.
const lineCanvas = { ...canvas, components: [{ id: 'e1', type: 'line' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 1_000, resizable: true, background: '#000000' }] }
const tableCanvas = { ...canvas, components: [{ id: 'e1', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: false, tableBind: 'transactions[]' }] }
const mixedCanvas = { ...canvas, components: [...textCanvas.components, { id: 'e2', type: 'rect' as const, band: 'content' as const, x: 0, y: 100_000, width: 72_000, height: 24_000, resizable: true, background: '#1b2a4a' }] }
const rectCanvas = { ...canvas, components: [{ id: 'e1', type: 'rect' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, background: '#1b2a4a' }] }
const imageCanvas = { ...canvas, components: [{ id: 'e1', type: 'image' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true }] }

const openDataTab = () => fireEvent.click(screen.getByRole('tab', { name: 'DATA' }))

describe('docked sample data panel', () => {
  it('keeps authoring available when empty, loads a tree, keeps accepted bytes authoritative, and preserves a prior sample on cancel', async () => {
    const openSample = vi.fn<SampleFileAccess['openSample']>().mockResolvedValueOnce({ name: 'sample.json', bytes: sampleBytes }).mockRejectedValueOnce(new FileAccessCancelled()).mockResolvedValueOnce({ name: 'replacement.json', bytes: replacementBytes })
    const request = vi.fn(async (...args: [string, unknown?, AbortSignal?]) => args[0] === 'identity' ? { snapshot, preview: { revision: 1, identity: 'b'.repeat(64) } } : args[0] === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : args[0] === 'render' ? { snapshot, bytes: new Uint8Array([9]).buffer, preview: { revision: 1, identity: 'b'.repeat(64), pdfSha256: PDF_FIXTURE_DIGEST, elapsedMs: RENDER_ELAPSED_MS, version: RENDER_ENGINE_VERSION, diagnostics: [] } } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} sampleFileAccess={{ openSample }} />)
    openDataTab()
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
    // STORY 13.5 — ONE MODE CONTROL, AND ITS NAME DOES NOT MOVE. This used to
    // match by pattern because the preview heading's button carried two names —
    // `Cancel and return to Design` while a render was in flight, `Return to
    // Design` once a PDF was installed — which made the wording on screen at
    // this line a timing property of the render pipeline. That button is gone.
    // The document bar's DESIGN switch calls the same `returnToDesign` and is
    // named the same in both states, so the pattern is no longer needed.
    fireEvent.click(screen.getByRole('button', { name: 'DESIGN' }))
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

  it('restores the root tab stop when replacing a sample after nested navigation', async () => {
    const first = acceptSampleData('first.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    const replacement = acceptSampleData('replacement.json', replacementBytes)
    const { rerender } = render(<DataPanel sample={first} busy={false} available onLoad={() => undefined} />)
    const customer = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!
    customer.focus(); fireEvent.keyDown(customer, { key: 'ArrowRight' })
    const name = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!
    fireEvent.keyDown(name, { key: 'ArrowDown' })
    rerender(<DataPanel sample={replacement} busy={false} available onLoad={() => undefined} />)
    await waitFor(() => {
      const items = screen.getAllByRole('treeitem')
      expect(items.filter((item) => item.tabIndex === 0)).toHaveLength(1)
      expect(items[0]).toHaveAttribute('tabindex', '0')
    })
    const root = screen.getAllByRole('treeitem')[0]!
    root.focus(); fireEvent.keyDown(root, { key: 'ArrowDown' })
    await waitFor(() => expect(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('report'))).toHaveFocus())
  })

  it('shows a binding rejection only for its original sample, component, and picked path', () => {
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada","email":"a@example.test"}}').buffer)
    const error = { sample, componentID: 'e1', segments: ['customer', 'name'], message: 'e1: binding rejected' }
    render(<DataPanel sample={sample} busy={false} available selectedComponentId="e1" bindingError={error} onLoad={() => undefined} />)
    const customer = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!
    fireEvent.click(customer)
    const name = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!
    fireEvent.click(name)
    expect(screen.getByRole('alert')).toHaveTextContent('binding rejected')
    const email = screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('email'))!
    fireEvent.click(email)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('selects a root scalar by keyboard and sends one opaque binding command, then paints distinct binding state', async () => {
    const boundCanvas = { ...textCanvas, components: [{ ...textCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }] }
    const request = vi.fn(async (operation: string) => operation === 'command'
      ? { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 4, canUndo: true, canvas: boundCanvas } }
      : { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: textCanvas } })
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"},"items":[]}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} initialSampleData={sample} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    const customer = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('customer'))!
    customer.focus()
    fireEvent.keyDown(customer, { key: 'ArrowRight' })
    const name = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '3' && item.textContent?.startsWith('name'))!
    name.focus()
    fireEvent.keyDown(name, { key: 'Enter' })
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).not.toHaveLength(0))
    const commands = request.mock.calls.filter(([operation]) => operation === 'command') as unknown as Array<[string, ArrayBuffer]>
    expect(commands).toHaveLength(1)
    const [operation, payload] = commands[0]!
    expect(operation).toBe('command')
    expect(new TextDecoder().decode(payload)).toBe('{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}')
    await waitFor(() => expect(screen.getByText('Bound to').parentElement).toHaveTextContent('Bound to customer.name'))
    expect(screen.getByLabelText('text component e1; bound to customer.name')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Undo' })).toBeEnabled()
  })

  it('does not install a late binding response after Start blank replaces the document', async () => {
    let resolveBinding!: (value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof textCanvas } }) => void
    const boundCanvas = { ...textCanvas, components: [{ ...textCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }] }
    const request = vi.fn((operation: string) => {
      if (operation === 'command') return new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof textCanvas } }>((resolve) => { resolveBinding = resolve })
      if (operation === 'load') return Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 1, canvas: textCanvas } })
      return Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: textCanvas } })
    })
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: textCanvas }} initialSampleData={sample} blankBytes={new Uint8Array([7]).buffer} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    const customer = screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '2' && item.textContent?.startsWith('customer'))!
    fireEvent.click(customer)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.getAttribute('aria-level') === '3' && item.textContent?.startsWith('name'))!)
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(1))
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Started an unnamed local template')).toBeInTheDocument())
    resolveBinding({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 4, canvas: boundCanvas } })
    await Promise.resolve(); await Promise.resolve()
    expect(screen.queryByText('Bound to')).not.toBeInTheDocument()
    expect(screen.queryByText('customer.name')).not.toBeInTheDocument()
  })

  it('installs a committed binding after reselection while keeping the newer selection', async () => {
    const selectedCanvas = { ...textCanvas, components: [...textCanvas.components, { id: 'e2', type: 'text' as const, band: 'content' as const, x: 80_000, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Other' }] }
    const boundCanvas = { ...selectedCanvas, components: [{ ...selectedCanvas.components[0]!, value: '{{customer.name}}', binding: 'customer.name' }, selectedCanvas.components[1]!] }
    let resolveBinding!: (value: { snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof selectedCanvas } }) => void
    const request = vi.fn((operation: string) => operation === 'command' ? new Promise<{ snapshot: { documentState: 'loaded'; revision: number; byteLength: number; canvas: typeof selectedCanvas } }>((resolve) => { resolveBinding = resolve }) : Promise.resolve({ snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: selectedCanvas } }))
    const sample = acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: selectedCanvas }} initialSampleData={sample} />)
    openDataTab()
    fireEvent.click(screen.getByLabelText('text component e1'))
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
    fireEvent.click(screen.getByRole('button', { name: 'Connect selected path' }))
    await waitFor(() => expect(request.mock.calls.filter(([operation]) => operation === 'command')).toHaveLength(1))
    fireEvent.click(screen.getByLabelText('text component e2'))
    resolveBinding({ snapshot: { documentState: 'loaded', revision: 2, byteLength: 4, canvas: boundCanvas } })
    await waitFor(() => expect(screen.getByLabelText('text component e1; bound to customer.name')).toBeInTheDocument())
    expect(screen.getByLabelText('text component e2')).toHaveClass('canvas-component-selected')
  })

  it('revokes a pending picker when an equal-revision Start blank replaces the document', async () => {
    let release!: (value: { name: string; bytes: ArrayBuffer }) => void
    const openSample = vi.fn(() => new Promise<{ name: string; bytes: ArrayBuffer }>((resolve) => { release = resolve }))
    const request = vi.fn(async (operation: string) => operation === 'serialize' ? { snapshot, bytes: new Uint8Array([1]).buffer } : { snapshot })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={snapshot} blankBytes={new Uint8Array([7]).buffer} sampleFileAccess={{ openSample }} />)
    openDataTab()
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

// STORY 14.4 / AC2 — THE FIFTH ARM OF THE PRE-FLIGHT LADDER.
//
// WHAT WAS WRONG. The ladder above had four arms and none of them inspected the
// selected component's KIND, so with a Line selected and a root scalar picked,
// Connect was ENABLED, the command went to Go, and the author read
// *"e1: only text components can receive a scalar binding"* — a refusal the
// panel could have stated before the attempt, and had in fact invited by
// telling them to pick a path in the first place.
//
// WHAT THIS IS NOT. It is not a judgement about the PATH. D-6.2.1 keeps sample
// runtime kind out of command legality: `params.*` and an empty collection are
// both still offered by the tree and still refused (or accepted, then failed at
// render) by the engine. Those are Story 14.6's badges, deliberately untouched
// here.
describe('the data panel states a component-kind refusal before the attempt', () => {
  const pickCustomerName = () => {
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
  }
  const sample = () => acceptSampleData('keys.json', new TextEncoder().encode('{"customer":{"name":"Ada"}}').buffer)
  // ⚠ A SINGLE MICROTASK IS NOT A FLUSH — the same standard `binding-
  // vocabulary.test.tsx` argues for and, until P8, the one this file did not
  // meet: it waited a bare 20ms before a zero-dispatch assertion, which is
  // shorter than the panel's own prose debounce and would have gone green on a
  // dispatch merely armed on a timer. Read from the constant so it cannot drift
  // under the debounce it exists to outlast.
  const settle = async () => {
    await new Promise((resolve) => setTimeout(resolve, PROSE_COMMIT_DEBOUNCE_MS + 20))
    for (let turn = 0; turn < 4; turn++) await Promise.resolve()
  }
  const openApp = (canvasFixture: typeof textCanvas | typeof lineCanvas | typeof tableCanvas | typeof mixedCanvas | typeof rectCanvas | typeof imageCanvas) => {
    const request = vi.fn(async (operation: string) => operation === 'command'
      ? { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 4, canvas: canvasFixture } }
      : { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: canvasFixture } })
    render(<App engine={{ request } as unknown as EngineClient} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: canvasFixture }} initialSampleData={sample()} />)
    openDataTab()
    return request
  }
  const commandsFrom = (request: ReturnType<typeof openApp>) => request.mock.calls.filter(([operation]) => operation === 'command')

  // ALL FOUR REFUSED KINDS, not the two that happened to read well. P5 found
  // the article bug — "The selected component is a image." — surviving because
  // `image` was the one arm no case exercised. The kind whose wording is
  // awkward is exactly the kind most likely to go untested.
  it.each([
    ['line', lineCanvas, 'The selected component is a line.'],
    ['table', tableCanvas, 'The selected component is a table.'],
    ['rect', rectCanvas, 'The selected component is a rectangle.'],
    ['image', imageCanvas, 'The selected component is an image.'],
  ])('disables Connect for a selected %s and states the engine\'s own reason first', async (kind, fixture, tail) => {
    const request = openApp(fixture)
    fireEvent.click(screen.getByLabelText(new RegExp(`^${kind} component e1`)))
    pickCustomerName()
    const status = screen.getByText(`Only text components can receive a scalar binding. ${tail}`)
    expect(status).toHaveAttribute('role', 'status')
    const connect = screen.getByRole('button', { name: 'Connect selected path' })
    expect(connect).toBeDisabled()
    // AND THE GESTURE DISPATCHES NOTHING. ⚠ P11 — WHAT THIS DOES AND DOES NOT
    // ADD. React does not fire `onClick` on a disabled button, so given the
    // assertion above this count could not have been anything but zero: it is
    // a consequence of the mechanism, not independent evidence of it. It is
    // asserted anyway because "zero bindComponentScalar commands dispatched" is
    // the acceptance criterion's own wording and should be readable as such —
    // and because it would catch a future arm that announced a reason without
    // disabling the control. It is not a second, independent witness, and an
    // earlier comment here claiming it was overstated the case.
    fireEvent.click(connect)
    await settle()
    expect(commandsFrom(request)).toHaveLength(0)
    // NO ERROR SURFACE IS RENDERED. ⚠ This does NOT prove the post-hoc refusal
    // channel is unreachable: `openApp`'s fake answers every command
    // successfully and can never produce a refusal, so this assertion would
    // pass with or without the gate. What it witnesses is only that the panel
    // states its reason through `role="status"` rather than raising an alert
    // for a state that is not an error. The refusal path itself still exists
    // and is still needed by `params`; its coverage is the pre-existing test
    // above, which supplies a `bindingError` directly.
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  // THE NON-VACUOUS HALF. This fixture differs from the two above ONLY in the
  // selected component's kind, and the arm must NOT fire: an over-broad gate
  // that disabled Connect for everything would pass every assertion above.
  it('leaves Connect enabled for a text component and sends the same bytes it always sent', async () => {
    const request = openApp(textCanvas)
    fireEvent.click(screen.getByLabelText(/^text component e1/))
    pickCustomerName()
    expect(screen.queryByText(/Only text components can receive a scalar binding/)).not.toBeInTheDocument()
    const connect = screen.getByRole('button', { name: 'Connect selected path' })
    expect(connect).toBeEnabled()
    fireEvent.click(connect)
    await waitFor(() => expect(commandsFrom(request)).toHaveLength(1))
    const [, payload] = commandsFrom(request)[0] as unknown as [string, ArrayBuffer]
    expect(new TextDecoder().decode(payload)).toBe('{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["customer","name"]}')
  })

  // A MULTI-SELECTION HAS NO ONE KIND TO SPEAK FOR. `selectedComponentId` is
  // undefined for it, so the arm above it wins and the panel says the true
  // thing rather than picking one of the two kinds to complain about.
  it('keeps "select one component first" for a multi-selection carrying a non-text kind', () => {
    openApp(mixedCanvas)
    fireEvent.click(screen.getByLabelText(/^text component e1/))
    fireEvent.click(screen.getByLabelText(/^rect component e2/), { shiftKey: true })
    pickCustomerName()
    expect(screen.getByText('Binding unavailable: select one component first.')).toBeInTheDocument()
    expect(screen.queryByText(/Only text components can receive a scalar binding/)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Connect selected path' })).toBeDisabled()
  })

  // ARM ORDER, PROVED AT THE COMPONENT RATHER THAN INFERRED FROM THE APP. With
  // NOTHING selected the kind is unknown, and "select one component first" is
  // still the true first thing to say — so the fifth arm is placed after it.
  // Passing both props here is the only way to witness the ordering, because
  // App never supplies a type without an id.
  it('lets the no-selection arm win even when a kind is supplied', () => {
    render(<DataPanel sample={sample()} busy={false} available selectedComponentType="line" onLoad={() => undefined} />)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
    expect(screen.getByText('Binding unavailable: select one component first.')).toBeInTheDocument()
    expect(screen.queryByText(/Only text components can receive a scalar binding/)).not.toBeInTheDocument()
  })

  // P1 — THE FAIL-OPEN, AND THE STATE THAT REACHES IT.
  //
  // The first spelling of the fifth arm was `selectedComponentType !== undefined
  // && !…includes(…)`, so an UNDEFINED type skipped the arm entirely and handed
  // Connect back — restoring exactly the round-trip refusal this story exists to
  // remove. The state is reachable in `App.tsx`: the id is passed from the
  // selection unconditionally, while the kind is looked up in the projection, so
  // an absent canvas or an id no longer in the projection yields id-present /
  // kind-absent. A panel that does not know the kind cannot know the engine will
  // accept it.
  it('refuses to invite a pick when the selected component has no known kind', () => {
    render(<DataPanel sample={sample()} busy={false} available selectedComponentId="e1" onLoad={() => undefined} onConnect={() => undefined} />)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
    expect(screen.getByText('Only text components can receive a scalar binding. The selected component is not in the current projection.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Connect selected path' })).toBeDisabled()
    // NON-VACUITY: the identical render differing ONLY in a known text kind
    // leaves Connect enabled, so the refusal above is the unknown kind's doing
    // and not a fixture that disabled the button some other way.
    expect(screen.queryByText(/Ready to ask the engine to bind/)).not.toBeInTheDocument()
  })

  it('leaves Connect enabled for the same render once the kind is known to be text', () => {
    render(<DataPanel sample={sample()} busy={false} available selectedComponentId="e1" selectedComponentType="text" onLoad={() => undefined} onConnect={() => undefined} />)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('customer'))!)
    fireEvent.click(screen.getAllByRole('treeitem').find((item) => item.textContent?.startsWith('name'))!)
    expect(screen.getByRole('button', { name: 'Connect selected path' })).toBeEnabled()
    expect(screen.queryByText(/Only text components can receive a scalar binding/)).not.toBeInTheDocument()
  })

  // AND THE ARMS ABOVE IT ARE UNCHANGED: with a component selected but no path
  // picked, the ladder still asks for the path rather than pre-emptively
  // complaining about the kind.
  it('asks for a path before it speaks about the kind', () => {
    render(<DataPanel sample={sample()} busy={false} available selectedComponentId="e1" selectedComponentType="rect" onLoad={() => undefined} />)
    expect(screen.getByText('Choose an offered root scalar path.')).toBeInTheDocument()
    expect(screen.queryByText(/Only text components can receive a scalar binding/)).not.toBeInTheDocument()
  })
})
