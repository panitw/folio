import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { IDBFactory as FakeIndexedDBFactory } from 'fake-indexeddb'
import App from './App'
import type { EngineClient } from './engine-client'
import { sfntWithNames } from './test/sfnt-fixture'

// STORY 16.2 AT THE BROWSER BOUNDARY — A FETCHED FACE STAYS ON THIS MACHINE.
//
// These tests drive the whole designer, because the claims are about the
// DESIGNER and not about the store module: "picking it again fetches nothing",
// "it works with the network down", "storage that cannot be opened leaves a
// working designer". None of those can be asserted from `font-store.ts` alone.
//
// THE BACKING STORE IS `fake-indexeddb`, installed onto `globalThis` for this
// file only. jsdom 28.1.0 provides no IndexedDB at all, so without it the
// designer would take its own degraded path here and every test below would be
// asserting the degradation. A FRESH factory per test, because the store's
// whole property is that it survives — including, otherwise, into the next
// test.
//
// AND THE RESIDUAL IS NAMED: real browser IndexedDB is proven only by a
// browser, and that witness is routed to Story 16.3's run. Nothing here claims
// otherwise.

const face = (name: string) => ({ face: name, assetKey: '', family: '', style: '' })
const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }], defaultFontSize: 12000, contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
const engine = (request: unknown) => ({ request }) as unknown as EngineClient

const kanitFace = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2020 The Kanit Project Authors' }])
const kanitMetadata = 'name: "Kanit"\nlicense: "OFL"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "Kanit-Regular.ttf"\n}\n'
const kanitLicence = 'This Font Software is licensed under the SIL Open Font License, Version 1.1.'

/** The upstream a Kanit pick reads: three round-trips, in this order. */
const upstreamFetch = () => vi.fn(async (url: string) => {
  if (url.endsWith('/ofl/kanit/METADATA.pb')) return { ok: true, status: 200, text: async () => kanitMetadata }
  if (url.endsWith('/ofl/kanit/OFL.txt')) return { ok: true, status: 200, text: async () => kanitLicence }
  if (url.endsWith('/ofl/kanit/Kanit-Regular.ttf')) return { ok: true, status: 200, arrayBuffer: async () => kanitFace }
  return { ok: false, status: 404, text: async () => '' }
})

const timeoutError = () => new DOMException('The operation was aborted due to timeout', 'TimeoutError')

let restoreFetch: typeof globalThis.fetch
let restoreIndexedDB: PropertyDescriptor | undefined

beforeEach(() => {
  restoreFetch = globalThis.fetch
  restoreIndexedDB = Object.getOwnPropertyDescriptor(globalThis, 'indexedDB')
  Object.defineProperty(globalThis, 'indexedDB', { value: new FakeIndexedDBFactory(), configurable: true, writable: true })
})

afterEach(() => {
  globalThis.fetch = restoreFetch
  if (restoreIndexedDB) Object.defineProperty(globalThis, 'indexedDB', restoreIndexedDB)
  else Reflect.deleteProperty(globalThis, 'indexedDB')
})

const mount = (request: unknown, chains = canvas.fontChains) => {
  const componentCanvas = { ...canvas, fontChains: chains, components: [textComponent] }
  render(<App engine={engine(request)} blankBytes={new Uint8Array([1, 2, 3]).buffer} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: componentCanvas }} />)
  fireEvent.click(screen.getByLabelText('text component e1'))
}

const commandRequest = () => vi.fn(async (operation: string) => ({ snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } } }))

/** Drives the family control the way an author does. Returns false when the row is not offered. */
const pick = (query: string, name: RegExp) => {
  const combobox = screen.getByRole('combobox', { name: 'Font family' })
  fireEvent.focus(combobox)
  fireEvent.change(combobox, { target: { value: query } })
  const option = screen.queryByRole('option', { name })
  if (option) fireEvent.click(option)
  return option !== null
}

/** The store panel's own live region, so this is never confused with the panel's other status lines. */
const storeNote = () => within(screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })).getByRole('status')

const embedPayloads = (request: { mock: { calls: unknown[][] } }) => request.mock.calls
  .filter((call) => call[0] === 'command')
  .map((call) => JSON.parse(new TextDecoder().decode(call[1] as ArrayBuffer)) as Record<string, unknown>)

describe('a fetched face stays on this machine', () => {
  // AC1. The pick fetches, embeds, and KEEPS the face — under the SHA-256 of
  // its bytes, with everything the embed command requires beside it.
  it('keeps a fetched face, with its licence record, and offers it back as already downloaded', async () => {
    const fetchStub = upstreamFetch()
    globalThis.fetch = fetchStub as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/), 'a snapshot family must be offered before it is stored').toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    // THE ROW SAYS WHAT IT IS: the family, the style, its size and the day it
    // was downloaded — the two facts an author needs to decide which to let go.
    expect(screen.getByText(/Regular · \d+ KB · downloaded \d{4}-\d{2}-\d{2}/)).toBeInTheDocument()
    // AND THE FAMILY CONTROL NOW OFFERS IT FROM THE STORE. The document
    // declares a chain for it too, so it is searched by a row that is offered
    // as an ADDITION only while the chain is absent — the store's own listing
    // is what this asserts.
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    fireEvent.change(screen.getByRole('combobox', { name: 'Font family' }), { target: { value: 'Kanit' } })
    expect(screen.getByRole('option', { name: /^Kanit\s*—\s*add to document, already downloaded to this machine$/ })).toBeInTheDocument()
  })

  // AC2 AND AC3 TOGETHER, WHICH IS THE WHOLE POINT OF THE STORE. A second
  // document picks the same family with the NETWORK REMOVED ENTIRELY: every
  // fetch throws. The pick must still embed, from the stored bytes, with the
  // same licence record.
  it('embeds a stored family in a second document with no network at all, fetching nothing', async () => {
    globalThis.fetch = upstreamFetch() as never
    const first = commandRequest()
    mount(first)
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    const fetched = embedPayloads(first)
    expect(fetched).toHaveLength(1)
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())

    // THE NETWORK IS GONE. Any request at all now fails, so a pick that
    // succeeds can only have come from the store.
    const offline = vi.fn(async () => { throw new TypeError('Failed to fetch') })
    globalThis.fetch = offline as never
    fireEvent.click(screen.getByLabelText('text component e1'))
    const before = embedPayloads(first).length
    expect(pick('Kanit', /^Kanit\s*—\s*add to document, already downloaded to this machine$/), 'the stored family must still be offered offline').toBe(true)
    await waitFor(() => expect(embedPayloads(first).length).toBeGreaterThan(before))
    expect(offline, 'a stored pick must not reach the network at all').not.toHaveBeenCalled()

    // AND IT IS THE SAME DOCUMENT CONTENT, not a degraded one: the same bytes,
    // the same three fields the engine refuses a document without, and the same
    // provenance string.
    const stored = embedPayloads(first).at(-1)!
    expect(stored['kind']).toBe('embedFontFamily')
    expect(stored['data']).toBe(fetched[0]['data'])
    expect(stored['licence']).toBe(fetched[0]['licence'])
    expect(stored['licenceText']).toBe(fetched[0]['licenceText'])
    expect(stored['copyright']).toBe(fetched[0]['copyright'])
    expect(stored['source']).toBe(fetched[0]['source'])
    expect(Object.keys(stored)).toHaveLength(12)
  })

  // A FAMILY THIS MACHINE DOES NOT HOLD, WITH NO NETWORK, IS A STATED
  // DEGRADATION — never a document that will not render, and never a silence.
  it('states that a family cannot be added right now when it is neither stored nor reachable', async () => {
    globalThis.fetch = (vi.fn(async () => { throw new TypeError('Failed to fetch') })) as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toMatch(/without a network connection/)
    expect(alert.textContent).toMatch(/faces this machine already holds are still offered/)
  })

  // THE REMOVAL AFFORDANCE. It names what it removes, and it states — in the
  // panel, not in a decision log — that documents are unaffected. An author who
  // has to GUESS whether a delete button reaches their saved work will simply
  // never press it.
  it('removes a face by name, and says documents that embed it are unchanged', async () => {
    globalThis.fetch = upstreamFetch() as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    // The standing note is present before anything is removed, because it is
    // what the list IS, not a consequence of pressing the button.
    expect(screen.getByText(/not the fonts installed on your computer, which this designer never looks at/)).toBeInTheDocument()
    const embedsBefore = embedPayloads(request).length
    fireEvent.click(screen.getByRole('button', { name: 'Remove Kanit (Regular) from this machine' }))
    await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit \(Regular\) was removed from this machine/))
    expect(storeNote().textContent).toMatch(/Documents that already embed it are unchanged/)
    // NOT A DOCUMENT CHANGE. No command, no revision, no undo entry: removing a
    // cached copy is a machine action, and a `.folio` carries its own faces.
    expect(embedPayloads(request)).toHaveLength(embedsBefore)
    await waitFor(() => expect(screen.queryByText(/Kanit/, { selector: '.machine-font-name' })).not.toBeInTheDocument())
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
  })

  // STORAGE THAT CANNOT BE OPENED LEAVES A WORKING DESIGNER. A private window
  // or blocked site data is a standing condition, not an error the author can
  // act on — so the group is empty, picks still fetch, and the message states
  // which of those two things is true.
  it('still designs, and still picks, when the store cannot be opened at all', async () => {
    Reflect.deleteProperty(globalThis, 'indexedDB')
    const fetchStub = upstreamFetch()
    globalThis.fetch = fetchStub as never
    const request = commandRequest()
    mount(request)
    await waitFor(() => expect(storeNote().textContent).toMatch(/not letting the designer keep typefaces on this machine/))
    expect(storeNote().textContent).toMatch(/picking a family still fetches it/i)
    // AND THE PICK STILL WORKS. The store is a cache and a source, never a
    // precondition.
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(embedPayloads(request)).toHaveLength(1))
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
  })
})

// ─────────────────────────────────────────────────────────────────────────────
// THE STALL, AND THE THREE THINGS THAT MUST BE TRUE OF IT (D-16.R.14/.15).
//
// A fetch that rejects releases the pick's hold through its `finally`. A fetch
// that HANGS never settles, so the hold is never released and the font control
// is disabled for the rest of the session, with no message and no way back. The
// timeout turns that into an ordinary stated degradation — and the RELEASE is
// the half that has to be red-proved, because it is invisible when it works.
describe('a pick that stalls rather than failing', () => {
  // (a) RELEASE. The timeout's rejection reaches the pick, the hold comes off,
  // the control is usable again and the degradation is stated.
  //
  // RED-PROOF: delete the `holdFontChain(false)` in `pickCatalogueFamily`'s
  // `finally` and this reds on the disabled combobox.
  it('releases the hold and states the degradation when the fetch times out', async () => {
    let stall: (() => void) | undefined
    const fetchStub = vi.fn(async (url: string) => {
      if (url.endsWith('/ofl/kanit/METADATA.pb')) return new Promise<never>((_resolve, reject) => { stall = () => reject(timeoutError()) })
      return { ok: false, status: 404, text: async () => '' }
    })
    globalThis.fetch = fetchStub as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(stall).toBeDefined())
    // THE CONTROL IS HELD WHILE THE FETCH IS IN FLIGHT — which is the state
    // that, without a timeout, would never end.
    expect(screen.getByRole('combobox', { name: 'Font family' })).toBeDisabled()
    stall!()
    // AND THE HOLD COMES OFF.
    await waitFor(() => expect(screen.getByRole('combobox', { name: 'Font family' })).not.toBeDisabled())
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toMatch(/stopped responding/)
    expect(alert.textContent).toMatch(/waited 30 seconds/)
    // AND IT IS NOT THE OFFLINE SENTENCE, which would be false here: the
    // network is up and the host is hanging.
    expect(alert.textContent).not.toMatch(/without a network connection/)
  })

  // (b) THE SECOND CARRIER — the instance D-16.R.15 predicted. The hold is
  // backed by BOTH a ref and React state, and the document-reset path once
  // cleared only the state copy: replacing the document mid-stall left the ref
  // stuck `true` for the session, with the control looking perfectly enabled
  // while every later pick silently did nothing. Proving (a) alone re-proves a
  // shape that was already fixed; this is the one that can regress.
  it('is clear after the document is replaced mid-stall, so a later pick still commits', async () => {
    let stall: (() => void) | undefined
    let first = true
    const fetchStub = vi.fn(async (url: string) => {
      if (url.endsWith('/ofl/kanit/METADATA.pb')) {
        if (first) { first = false; return new Promise<never>((_resolve, reject) => { stall = () => reject(timeoutError()) }) }
        return { ok: true, status: 200, text: async () => kanitMetadata }
      }
      if (url.endsWith('/ofl/kanit/OFL.txt')) return { ok: true, status: 200, text: async () => kanitLicence }
      if (url.endsWith('/ofl/kanit/Kanit-Regular.ttf')) return { ok: true, status: 200, arrayBuffer: async () => kanitFace }
      return { ok: false, status: 404, text: async () => '' }
    })
    globalThis.fetch = fetchStub as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(stall).toBeDefined())

    // THE DOCUMENT IS REPLACED WHILE THE PICK IS STILL STALLED.
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())
    stall!()

    fireEvent.click(screen.getByLabelText('text component e1'))
    const before = embedPayloads(request).length
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/), 'the pick must still be offered in the new document').toBe(true)
    await waitFor(() => expect(embedPayloads(request).length).toBeGreaterThan(before))
  })

  // (c) NO SILENT RETRY, WITH THE COUNT AS A LITERAL. A retry over a
  // deterministic stall hides it — Story 16.0's `Never:` clause, same reasoning
  // — and the count is written out rather than derived from the code path that
  // drives the fetches, because a derived count agrees with any implementation.
  it('makes exactly one request and never repeats it after the timeout', async () => {
    let stall: (() => void) | undefined
    const fetchStub = vi.fn(async (url: string) => {
      if (url.endsWith('/ofl/kanit/METADATA.pb')) return new Promise<never>((_resolve, reject) => { stall = () => reject(timeoutError()) })
      return { ok: false, status: 404, text: async () => '' }
    })
    globalThis.fetch = fetchStub as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*add to document$/)).toBe(true)
    await waitFor(() => expect(stall).toBeDefined())
    stall!()
    await screen.findByRole('alert')
    // ONE. The first probe stalled, so the chain ends there: no further probe,
    // no licence read, no byte read, and no second attempt at the one that
    // stalled.
    expect(fetchStub).toHaveBeenCalledTimes(1)
    // And still one a moment later, so "no retry" is a property of the code
    // rather than of when the assertion happened to run.
    await Promise.resolve()
    expect(fetchStub).toHaveBeenCalledTimes(1)
  })
})
