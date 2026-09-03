import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { IDBFactory as FakeIndexedDBFactory, IDBObjectStore as FakeIndexedDBObjectStore } from 'fake-indexeddb'
import App from './App'
import type { EngineClient } from './engine-client'
import { sfntWithNames } from './test/sfnt-fixture'
import { embeddedFaceFamily } from './embedded-face-family'
import { previewFaceFamily } from './preview-face-family'
import { openFontStore, storedFaceKey, storeWriteRefusal, type StoredFaceRecord } from './font-store'
import { addableFamilyCount, webFamilies } from './font-index'
import { catalogueFaces } from './generated/font-catalogue'

/**
 * THE STORE PANEL'S HEADING, WRITTEN ONCE (Story 16.4).
 *
 * Two product sentences point the author at this panel — the quota refusal and
 * the late-embed refusal — so the name they use and the name the panel renders
 * are asserted against each other rather than each spelled out on its own.
 * Before 16.4 the panel borrowed `AVAILABLE LOCALLY` from a dropdown group that
 * lists strictly more than it does, and two differently populated regions shared
 * one name.
 */
const storePanelHeading = 'TYPEFACES THIS DESIGNER HAS DOWNLOADED'

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

// THE VARIABLE FACE THE INSTALL MUST REFUSE, AND IT IS THE FILE GO EMBEDS.
//
// `folio-go/testfont_embed_test.go:34` reaches this exact path by `//go:embed`,
// so the bytes the designer refuses at install are the same bytes
// `fontset.RefuseVariableFace` refuses at the command. A synthesised `fvar`
// would test the predicate against a fixture of our own making; this tests it
// against the one the authority is held to. `src/font-variable-face-tie.test.ts`
// carries the predicate-level tie and its over-broadness control; what is
// asserted HERE is the different claim the matrix row actually makes — that the
// INSTALL PATH refuses, names the family, and keeps nothing.
const variableFaceFixture = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..', 'folio-go', 'testdata', 'fonts', 'notosansthai-variable-testonly', 'NotoSansThai-VF.ttf')

/** Fails, never skips, if the shared fixture moves — a tie that cannot read its fixture is broken, not absent. */
const variableFaceBytes = (): ArrayBuffer => {
  expect(fs.existsSync(variableFaceFixture), `the shared fixture ${variableFaceFixture} is not there; Go embeds it by this exact path`).toBe(true)
  const held = fs.readFileSync(variableFaceFixture)
  return held.buffer.slice(held.byteOffset, held.byteOffset + held.byteLength) as ArrayBuffer
}

/** The Kanit upstream, but the face it serves is the variable one. Everything before the bytes still admits. */
const upstreamServingAVariableFace = (bytes: ArrayBuffer) => vi.fn(async (url: string) => {
  if (url.endsWith('/ofl/kanit/METADATA.pb')) return { ok: true, status: 200, text: async () => kanitMetadata }
  if (url.endsWith('/ofl/kanit/OFL.txt')) return { ok: true, status: 200, text: async () => kanitLicence }
  if (url.endsWith('/ofl/kanit/Kanit-Regular.ttf')) return { ok: true, status: 200, arrayBuffer: async () => bytes }
  return { ok: false, status: 404, text: async () => '' }
})

/** Everything the store holds right now, read past the designer rather than through its own list. */
const faceRecordsOnThisMachine = async (): Promise<ReadonlyArray<Readonly<{ family: string }>>> => {
  const opened = await openFontStore(globalThis.indexedDB)
  expect(opened.ok, 'the fake store must open, or reading it back asserts nothing').toBe(true)
  if (!opened.ok) return []
  const listed = await opened.value.list()
  expect(listed.ok).toBe(true)
  return listed.ok ? listed.value : []
}

// The page font set jsdom does not implement, installed for the one test that
// needs to watch a face reach it. Written the way `App.test.tsx` writes it —
// with neither of the two spellings `canvas-authority-contract.test.ts`
// forbids, so this file needs no carve-out from that contract.
function installStubFontSet(): Readonly<{ restore: () => void; added: string[] }> {
  class StubFace {
    readonly family: string
    constructor(family: string) { this.family = family }
    load(): Promise<StubFace> { return Promise.resolve(this) }
  }
  const added: string[] = []
  const set = { add: (loaded: StubFace) => { added.push(loaded.family); return undefined }, delete: () => undefined }
  Object.defineProperty(globalThis, 'FontFace', { value: StubFace, configurable: true, writable: true })
  Object.defineProperty(document, 'fonts', { value: set, configurable: true, writable: true })
  return { restore: () => { Reflect.deleteProperty(globalThis, 'FontFace'); Reflect.deleteProperty(document, 'fonts') }, added }
}

/**
 * A DOCUMENT THAT PAINTS A FACE IT DOES NOT CARRY.
 *
 * The chain declares one SHIPPED face and no carried entry, so
 * `carriedFaceKeys` — which reads `fontChains[].entries[].assetKey` — is empty
 * and the document-scoped registration has nothing to do. The projection still
 * attributes the fragment to `key`, so the only set that can supply a family
 * for it is the machine store's.
 */
const storedOnlyFaceCanvas = (key: string) => ({
  ...canvas,
  fontFamilies: ['body'],
  fontChains: [{ name: 'body', entries: [face('Noto Sans')] }],
  components: [{ id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'ignored', textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 12_000, advance: 16_000, width: 24_000, fragments: [{ text: 'สัญญา', x: 0, assetKey: key }] }] } }],
})

/** One complete stored record over `bytes`, written straight into the store before the designer opens it. */
const storedOnly = (key: string, bytes: ArrayBuffer): StoredFaceRecord => ({
  key,
  family: 'A Machine Face',
  style: 'Regular',
  licence: 'OFL-1.1',
  licenceText: kanitLicence,
  copyright: 'Copyright 2026 A Face Only This Machine Has',
  source: 'google/fonts — ofl/amachineface/AMachineFace-Regular.ttf, fetched 2026-09-03',
  mediaType: 'font/ttf',
  scripts: ['latin'],
  fetchedAt: '2026-09-03',
  byteLength: bytes.byteLength,
  bytes,
})

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

// STORY 16.3 — THE FONT BROWSER'S SPECIMEN BYTES COME FROM THE SAME THREE TIERS
// A PICK DOES, AND THE STORED TIER IS THE ONE THAT MUST NEED NO NETWORK.
//
// `App.tsx`'s `browserSpecimenBytes` is never executed by `FontBrowser.test.tsx`,
// which supplies its own resolver on every render. A review demonstrated the
// gap two ways — letting the `stored` branch fall through to `fetchWebFamily`,
// and reading `source.family` where it should read `source.record.key` — and in
// both the whole suite stayed green while every face this machine already holds
// would have needed the network. That is precisely the DW-176 property this
// story claims to have witnessed, so it is asserted here, where a fake store and
// a stub page font set already stand up.
describe('the font browser sets a stored family\'s specimen from the store', () => {
  it('reads a stored specimen with every fetch rejecting, and says a web row cannot be shown', async () => {
    const bytes = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2026 A Face Only This Machine Has' }])
    const key = await storedFaceKey(bytes)
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    const written = await opened.value.put(storedOnly(key, bytes))
    expect(written.ok, 'the fixture face must really be in the store before the designer opens it').toBe(true)

    const fontSet = installStubFontSet()
    try {
      // THE NETWORK IS GONE. Every request fails, so a specimen that renders can
      // only have come from the store.
      const offline = vi.fn(async () => { throw new TypeError('Failed to fetch') })
      globalThis.fetch = offline as never
      mount(commandRequest())
      await waitFor(() => expect(screen.getByText(/A Machine Face/, { selector: '.machine-font-name' })).toBeInTheDocument())

      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      fireEvent.click(screen.getByRole('button', { name: /^Add fonts…/ }))
      const dialog = screen.getByRole('dialog', { name: 'Font browser' })

      // THE STORED ROW. Searched by name because a family the snapshot does not
      // rank sorts last, which is the correct order and the wrong page.
      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'A Machine Face' } })
      expect(within(dialog).getByText('downloaded to this machine')).toBeInTheDocument()
      const specimen = await within(dialog).findByText('Everyone has the right to freedom of thought')
      expect(specimen.style.fontFamily, 'a family this machine already holds must be set in itself with no network').toBe(previewFaceFamily('A Machine Face'))
      expect(fontSet.added).toContain(previewFaceFamily('A Machine Face'))

      // AND A WEB ROW, ON THE SAME SCREEN, WITH THE SAME NETWORK. It says it
      // cannot be shown rather than borrowing the panel's typeface — which is
      // also what makes the line above a claim about the STORE and not about a
      // resolver that happens to succeed for everything.
      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'Kanit' } })
      expect(await within(dialog).findByText(/Kanit cannot be shown set in itself/)).toBeInTheDocument()
      expect(within(dialog).getByText('downloaded when you install it')).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })
})

// STORY 16.3 — A REFUSAL REACHED THROUGH THE `'caller'` ANNOUNCER IS RETURNED,
// AND THE MODAL IS WHERE IT LANDS.
//
// The seam has two announcers: the family control's pick writes a refusal to the
// panel, and the browser's confirm gets it BACK so it can name it against the
// row that earned it. Nothing asserted the return value on the App side, and a
// review demonstrated what that costs: restore `refuse(reason); return` in place
// of `return refuse(reason)` and every family reports success — `confirm` clears
// the staged set and closes the modal, so a licence refusal or an upstream 404
// dismisses the browser with the author believing everything was embedded, and
// the `'caller'` path paints no panel alert either. The suite stayed green.
describe('the font browser names a refusal the seam returned', () => {
  it('keeps the modal open, names the family, and sends no command', async () => {
    const fontSet = installStubFontSet()
    try {
      // Nothing is published upstream: every probe 404s, which is the refusal
      // `fetchWebFamily` states in its own words.
      const gone = vi.fn(async () => ({ ok: false, status: 404, text: async () => '' }))
      globalThis.fetch = gone as never
      const request = commandRequest()
      mount(request)
      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      fireEvent.click(screen.getByRole('button', { name: /^Add fonts…/ }))
      const dialog = screen.getByRole('dialog', { name: 'Font browser' })

      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'Kanit' } })
      fireEvent.click(await within(dialog).findByRole('button', { name: 'Install Kanit on this machine' }))
      fireEvent.click(within(dialog).getByRole('button', { name: 'Install 1 on this machine' }))

      // THE ENGINE'S OWN SENTENCE, INSIDE THE DIALOG, AGAINST THE FAMILY.
      expect(await within(dialog).findByText(/Kanit: Kanit is in this designer's snapshot of the family list but is no longer published upstream/)).toBeInTheDocument()
      // AND THE MODAL IS STILL THERE, with Kanit still staged, so the author
      // can read the refusal and retry without finding the family again.
      expect(screen.getByRole('dialog', { name: 'Font browser' })).toBeInTheDocument()
      expect(within(dialog).getByRole('button', { name: 'Remove Kanit from the families to install' })).toBeInTheDocument()
      // NEUTRALISED VACUOUS SURVIVOR (Story 16.5). This line was
      // `expect(embedPayloads(request)).toEqual([])` — "nothing reached the
      // document" — which was a real claim while confirming embedded and is
      // TRIVIALLY TRUE now that confirming never sends a command at all. It is
      // kept, because a confirm that started sending commands would still be a
      // defect, but it can no longer carry the test on its own.
      //
      // THE CLAIM THAT CAN STILL FAIL IS ABOUT THE MACHINE, WHICH IS WHERE A
      // CONFIRM NOW ACTS. A refused install must leave NOTHING on this machine:
      // no store row, no listing entry, and the family still offered as one to
      // install rather than one already downloaded. An install that wrote a
      // partial record before failing — bytes without terms, which is the
      // record shape the engine refuses a document over — passes the line
      // above and fails every line below it.
      expect(embedPayloads(request)).toEqual([])
      expect(screen.getByText(/No typefaces have been downloaded to this machine yet/), 'a refused install must keep nothing').toBeInTheDocument()
      expect(within(screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })).queryAllByRole('button', { name: /^Remove / })).toEqual([])
      // AND THE FAMILY CONTROL STILL OFFERS IT AS AN INSTALL. `offeredFamilies`
      // reads the store's own listing, so a row that had moved to the
      // already-downloaded note would mean a record survived the refusal.
      fireEvent.keyDown(dialog, { key: 'Escape' })
      await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Font browser' })).toBeNull())
      const combobox = screen.getByRole('combobox', { name: 'Font family' })
      fireEvent.focus(combobox)
      fireEvent.change(combobox, { target: { value: 'Kanit' } })
      expect(screen.getByRole('option', { name: /^Kanit\s*—\s*install on this machine$/ })).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  it('does not re-materialise over a replaced document carrying the previous one\'s staged set', async () => {
    const fontSet = installStubFontSet()
    try {
      globalThis.fetch = upstreamFetch() as never
      mount(commandRequest())
      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      fireEvent.click(screen.getByRole('button', { name: /^Add fonts…/ }))
      const dialog = screen.getByRole('dialog', { name: 'Font browser' })
      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'Kanit' } })
      fireEvent.click(await within(dialog).findByRole('button', { name: 'Install Kanit on this machine' }))
      expect(within(dialog).getByText(/^1 family ready to install/)).toBeInTheDocument()

      // REPLACING THE DOCUMENT IS THE END OF THIS MODAL, not a pause in it. The
      // open flag lived outside `clearDocumentInteraction`, so the modal
      // vanished only while `canvas` was momentarily undefined and then came
      // BACK over the new document — still carrying a staged set assembled
      // against the old one.
      fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
      await screen.findByText('Started an unnamed local template')
      // The settle condition is the modal GOING, not the document name — which
      // starts out as `Untitled template` and would have made this pass before
      // the replacement had happened at all.
      await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Font browser' })).toBeNull())
      // And it does not come back. The new document has a canvas of its own, so
      // the render guard alone never hid the modal for more than an instant.
      await waitFor(() => expect(screen.queryByText(/family ready to install/)).toBeNull())
      expect(screen.queryByRole('dialog', { name: 'Font browser' })).toBeNull()
    } finally {
      fontSet.restore()
    }
  })
})

describe('a fetched face stays on this machine', () => {
  // AC1, BEHAVIOUR-CHANGED BY STORY 16.5 IN ITS SECOND HALF. The pick fetches
  // and KEEPS the face — under the SHA-256 of its bytes, with everything the
  // embed command requires beside it — AND IT NO LONGER EMBEDS. Installing is
  // a machine action: no command, no revision, no history entry, no undo.
  it('keeps a fetched face with its licence record, sends no command for it, and offers it back as already downloaded', async () => {
    const fetchStub = upstreamFetch()
    globalThis.fetch = fetchStub as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/), 'a snapshot family must be offered before it is stored').toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    // THE DOCUMENT DID NOT MOVE. The store row above is the positive settle
    // condition — the install really finished — so this is a claim about a
    // completed install and not about an assertion that ran too early.
    expect(embedPayloads(request), 'installing a family sends no engine command at all').toEqual([])
    expect(request.mock.calls.filter((call) => call[0] === 'command'), 'no command means no revision, no history entry and no undo').toEqual([])
    // THE ROW SAYS WHAT IT IS: the family, the style, its size and the day it
    // was downloaded — the two facts an author needs to decide which to let go.
    expect(screen.getByText(/Regular · \d+ KB · downloaded \d{4}-\d{2}-\d{2}/)).toBeInTheDocument()
    // AND THE FAMILY CONTROL NOW OFFERS IT FROM THE STORE. The document
    // declares a chain for it too, so it is searched by a row that is offered
    // as an ADDITION only while the chain is absent — the store's own listing
    // is what this asserts.
    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    fireEvent.change(screen.getByRole('combobox', { name: 'Font family' }), { target: { value: 'Kanit' } })
    expect(screen.getByRole('option', { name: /^Kanit\s*—\s*use it, already downloaded to this machine$/ })).toBeInTheDocument()
  })

  // AC2 AND AC3 TOGETHER, WHICH IS THE WHOLE POINT OF THE STORE, RE-ANCHORED BY
  // STORY 16.5 ONTO THE MOMENT THAT NOW CARRIES THE EMBED.
  //
  // BEHAVIOUR-CHANGED. The first pick used to embed AND keep, so the reference
  // payload this test compared against was the first document's own command.
  // Installing sends no command, so that reference no longer exists — the
  // comparison is now against THE FIXTURE BYTES THEMSELVES, which is a stronger
  // anchor anyway: it cannot agree with a payload the designer produced twice
  // from the same defect.
  //
  // The claim is unchanged and is the one Story 16.2 exists for: a second
  // document, with THE NETWORK REMOVED ENTIRELY, still gets the face and its
  // terms out of the store. Only the trigger moved, from the pick to first use.
  it('embeds a stored family in a second document with no network at all, fetching nothing', async () => {
    globalThis.fetch = upstreamFetch() as never
    const first = commandRequest()
    mount(first)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    // THE SETTLE CONDITION IS THE STORE ROW, and it has to be: there is no
    // command to wait on any more, so waiting on one would wait for ever.
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    expect(embedPayloads(first), 'the install must not have embedded anything').toEqual([])
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())

    // THE NETWORK IS GONE. Any request at all now fails, so a pick that
    // succeeds can only have come from the store.
    const offline = vi.fn(async () => { throw new TypeError('Failed to fetch') })
    globalThis.fetch = offline as never
    fireEvent.click(screen.getByLabelText('text component e1'))
    const before = embedPayloads(first).length
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/), 'the stored family must still be offered offline').toBe(true)
    await waitFor(() => expect(embedPayloads(first).length).toBeGreaterThan(before))
    expect(offline, 'a stored pick must not reach the network at all').not.toHaveBeenCalled()

    // AND IT IS THE SAME DOCUMENT CONTENT, not a degraded one: the same bytes
    // the upstream served, the three fields the engine refuses a document
    // without, and the same twelve-field arity.
    const stored = embedPayloads(first).find((payload) => payload['kind'] === 'embedFontFamily')!
    expect(stored, 'first use of a stored family must send the embed command').toBeDefined()
    expect(stored['data']).toBe(btoa(String.fromCharCode(...new Uint8Array(kanitFace))))
    expect(stored['licence']).toBe('OFL-1.1')
    expect(stored['licenceText']).toBe(kanitLicence)
    expect(stored['copyright']).toBe('Copyright 2020 The Kanit Project Authors')
    expect(String(stored['source'])).toContain('ofl/kanit/Kanit-Regular.ttf')
    expect(Object.keys(stored)).toHaveLength(12)

    // TWO COMMANDS, IN THIS ORDER, AND TWO UNDO ENTRIES. `canvas.fontFamilies`
    // is the closed set `style.fontFamily` may name, so the property command is
    // refused until the chain is declared — the engine forces the order and
    // nothing in the designer chooses it.
    await waitFor(() => expect(embedPayloads(first).map((payload) => payload['kind'])).toEqual(['embedFontFamily', 'updateComponentProperties']))
    const property = embedPayloads(first).at(-1)!
    expect(property['changes']).toEqual({ fontFamily: { op: 'set', value: 'Kanit' } })
    expect(property['ids']).toEqual(['e1'])
  })

  // A FAMILY THIS MACHINE DOES NOT HOLD, WITH NO NETWORK, IS A STATED
  // DEGRADATION — never a document that will not render, and never a silence.
  it('states that a family cannot be installed right now when it is neither stored nor reachable', async () => {
    globalThis.fetch = (vi.fn(async () => { throw new TypeError('Failed to fetch') })) as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toMatch(/You cannot install a family without a network connection/)
    expect(alert.textContent).toMatch(/faces this machine already holds are still offered/)
  })

  // THE REMOVAL AFFORDANCE. It names what it removes, and it states — in the
  // panel, not in a decision log — that documents are unaffected. An author who
  // has to GUESS whether a delete button reaches their saved work will simply
  // never press it.
  // NEUTRALISED VACUOUS SURVIVOR (Story 16.5), AND THE FIX IS TO MAKE THE
  // DOCUMENT ACTUALLY CARRY THE FACE FIRST.
  //
  // `expect(embedPayloads(request)).toHaveLength(embedsBefore)` was a real claim
  // while a pick embedded: `embedsBefore` was 1, the document had the face, and
  // the line said removing the machine copy did not reach it. A pick installs
  // now, so `embedsBefore` is 0 and the line reads "0 commands before, 0 after"
  // — TRUE OF A DESIGNER THAT CANNOT SEND COMMANDS AT ALL.
  //
  // So the test now does what its own name has always claimed: it puts the face
  // IN THE DOCUMENT — by using it, which is two commands — and only then removes
  // the machine copy. `embedsBefore` is 2, the count must still be 2 afterwards,
  // and the sentence about documents being unchanged is finally being asserted
  // over a document that has something to be unchanged about.
  it('removes a face by name, and says documents that embed it are unchanged', async () => {
    globalThis.fetch = upstreamFetch() as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    // The standing note is present before anything is removed, because it is
    // what the list IS, not a consequence of pressing the button.
    expect(screen.getByText(/not the fonts installed on your computer, which this designer never looks at/)).toBeInTheDocument()

    // FIRST USE, so the document really carries the face: the embed and then the
    // property commit, two commands and two undo entries.
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/), 'the installed family must be offered for use').toBe(true)
    await waitFor(() => expect(embedPayloads(request).map((payload) => payload['kind'])).toEqual(['embedFontFamily', 'updateComponentProperties']))
    const embedsBefore = embedPayloads(request).length
    expect(embedsBefore, 'the count this test guards must be non-zero, or guarding it says nothing').toBe(2)

    fireEvent.click(screen.getByRole('button', { name: 'Remove Kanit (Regular) from this machine' }))
    await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit \(Regular\) was removed from this machine/))
    expect(storeNote().textContent).toMatch(/Documents that already embed it are unchanged/)
    // NOT A DOCUMENT CHANGE. No command, no revision, no undo entry: removing a
    // cached copy is a machine action, and a `.folio` carries its own faces.
    expect(embedPayloads(request)).toHaveLength(embedsBefore)
    await waitFor(() => expect(screen.queryByText(/Kanit/, { selector: '.machine-font-name' })).not.toBeInTheDocument())
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
  })

  // STORY 16.4 — A FONT CHANGES GROUP BECAUSE THE AUTHOR ACTED, AND THE WHOLE
  // JOURNEY IS ONE TEST.
  //
  // 8.6's rule was "nothing says added, the entry simply moves", and three
  // groups extend it rather than replace it: install moves a row 3 → 2, first
  // use moves it 2 → 1. Both halves of what a pick does are asserted at the
  // same time as the move, because the move is only honest if the act behind it
  // is — a row that jumped to AVAILABLE LOCALLY without the bytes arriving, or
  // to IN THIS TEMPLATE without a command, would be a heading lying twice.
  it('moves a row 3 → 2 when it is installed and 2 → 1 when it is first used', async () => {
    globalThis.fetch = upstreamFetch() as never
    let declaredChains = ['body']
    const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
      if (operation === 'command' && payload) {
        const parsed = JSON.parse(new TextDecoder().decode(payload)) as Record<string, unknown>
        if (parsed['kind'] === 'embedFontFamily') declaredChains = ['body', 'Kanit']
      }
      return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: { ...canvas, fontFamilies: declaredChains, components: [textComponent] } } }
    })
    mount(request)
    const combobox = () => screen.getByRole('combobox', { name: 'Font family' })
    const groupOf = (family: string) => {
      fireEvent.focus(combobox())
      fireEvent.change(combobox(), { target: { value: family } })
      // The accessible name collapses the note's leading space, so the pattern
      // is whitespace-tolerant rather than pinned to one spelling of the gap —
      // and it accepts the bare name too, which is what group 1's rows carry.
      return screen.getByRole('option', { name: new RegExp(`^${family}\\s*(—|$)`) }).closest('[role="group"]')?.getAttribute('aria-label')
    }

    // GROUP 3 — not on this machine.
    expect(groupOf('Kanit')).toBe('AVAILABLE TO INSTALL')
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    // NOTHING ENTERED THE DOCUMENT. No engine command at all, so no property
    // commit either — which is what makes the row's new heading true.
    expect(embedPayloads(request), 'installing sends no command: the document did not move').toEqual([])

    // GROUP 2 — on this machine, not in this file.
    expect(groupOf('Kanit')).toBe('AVAILABLE LOCALLY')
    // AND PICKING FROM GROUP 2 REACHES NO NETWORK. Every fetch from here on
    // fails, so an embed that completes can only have read the machine's copy.
    const refuse = vi.fn(async () => { throw new TypeError('a face already on this machine must not be fetched') })
    globalThis.fetch = refuse as never
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/)).toBe(true)
    // TWO COMMANDS AND TWO UNDO ENTRIES, in the order the engine forces.
    await waitFor(() => expect(embedPayloads(request).map((sent) => sent['kind'])).toEqual(['embedFontFamily', 'updateComponentProperties']))
    expect(refuse, 'a pick from AVAILABLE LOCALLY must not touch the network').not.toHaveBeenCalled()

    // GROUP 1 — in the file. The projection is what says it arrived.
    await waitFor(() => expect(groupOf('Kanit')).toBe('IN THIS TEMPLATE'))
    // AND IT IS OFFERED ONCE, not once per group it has passed through.
    expect(screen.getAllByRole('option', { name: /^Kanit\s*(—|$)/ })).toHaveLength(1)
  })

  // THE REGRESSION THE ORDER REPAIR AND THE PARTITION EXIST FOR, DRIVEN THROUGH
  // THE CONTROL RATHER THAN ASSERTED OVER `offeredFamilies`.
  //
  // At HEAD a stored family took the INDEX POSITION of the web row it replaced.
  // `Philosopher` sits at offset 891 of 1273, so with a 50-row cap applied to
  // the union it was painted nowhere at all — while AVAILABLE LOCALLY was drawn
  // over a group that was missing it. A heading is never drawn over a member the
  // render cannot show.
  it('draws a deeply-ranked stored family under AVAILABLE LOCALLY, and the cap does not swallow it', async () => {
    const bytes = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2026 A Face Only This Machine Has' }])
    const key = await storedFaceKey(bytes)
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    const written = await opened.value.put({ ...storedOnly(key, bytes), family: 'Philosopher' })
    expect(written.ok, 'the fixture face must really be in the store before the designer opens it').toBe(true)
    // WITH NO NETWORK, so nothing here can be explained by a fetch.
    globalThis.fetch = vi.fn(async () => { throw new TypeError('Failed to fetch') }) as never
    mount(commandRequest())
    await waitFor(() => expect(screen.getByText(/Philosopher/, { selector: '.machine-font-name' })).toBeInTheDocument())

    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    const local = within(screen.getByRole('group', { name: 'AVAILABLE LOCALLY' })).getAllByRole('option')
    const install = within(screen.getByRole('group', { name: 'AVAILABLE TO INSTALL' })).getAllByRole('option')
    // IT IS UNDER THE HEADING THAT SAYS THE BYTES ARE HERE, AND IT SAYS SO.
    expect(local.map((option) => option.textContent)).toContain('Philosopher — use it, already downloaded to this machine')
    // AND THE GROUP IS DRAWN IN FULL — 31 committed faces plus this one against
    // a 50-row cap that used to be applied to the union before the split.
    expect(local, 'every installed row renders; the cap belongs to the other group').toHaveLength(catalogueFaces.length + 1)
    expect(install, 'the cap is still doing its job where it belongs').toHaveLength(50)
    // AND IT IS OFFERED ONCE: from the store, never also from the snapshot.
    expect(screen.getAllByRole('option', { name: /^Philosopher/ })).toHaveLength(1)
    // THE CAPPED GROUP'S NOTE COUNTS THE CAPPED GROUP. Group 2 and group 3
    // together are the whole addable union, so naming the union would be
    // counting rows the cap never touched.
    const counted = /Showing (\d+) of (\d+) families you can install/.exec(screen.getByText(/families you can install/).textContent ?? '')
    expect(counted).not.toBeNull()
    expect(Number(counted![2])).not.toBe(addableFamilyCount)
    expect(Number(counted![2]) + local.length).toBe(addableFamilyCount)
  })

  // AND WITH NO STORE AT ALL THE ROW MOVES 3 → 1 IN ONE STEP.
  //
  // `storeUnavailableEmbedNote` says the font went straight into the document,
  // and this asserts the consequence that sentence describes rather than the
  // sentence again: with nowhere to install to, there is no middle group for the
  // row to rest in, so it crosses group 2 without stopping. A row that landed
  // under AVAILABLE LOCALLY here would be claiming a machine copy that this
  // browser explicitly cannot keep.
  it('moves a row 3 → 1 in one step when the store cannot be opened at all', async () => {
    Reflect.deleteProperty(globalThis, 'indexedDB')
    globalThis.fetch = upstreamFetch() as never
    let declaredChains = ['body']
    const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
      if (operation === 'command' && payload) {
        const parsed = JSON.parse(new TextDecoder().decode(payload)) as Record<string, unknown>
        if (parsed['kind'] === 'embedFontFamily') declaredChains = ['body', 'Kanit']
      }
      return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: { ...canvas, fontFamilies: declaredChains, components: [textComponent] } } }
    })
    mount(request)
    await waitFor(() => expect(storeNote().textContent).toMatch(/not letting the designer keep typefaces on this machine/))
    const combobox = () => screen.getByRole('combobox', { name: 'Font family' })
    const groupOf = (family: string) => {
      fireEvent.focus(combobox())
      fireEvent.change(combobox(), { target: { value: family } })
      return screen.getByRole('option', { name: new RegExp(`^${family}\\s*(—|$)`) }).closest('[role="group"]')?.getAttribute('aria-label')
    }
    expect(groupOf('Kanit')).toBe('AVAILABLE TO INSTALL')
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(embedPayloads(request).map((sent) => sent['kind'])).toEqual(['embedFontFamily']))
    await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit went straight into this document/))
    // STRAIGHT TO GROUP 1, and never to group 2 — nothing is on this machine.
    await waitFor(() => expect(groupOf('Kanit')).toBe('IN THIS TEMPLATE'))
    expect(screen.queryByRole('group', { name: 'AVAILABLE LOCALLY' }), 'no family matches Kanit on a machine that keeps nothing').toBeNull()
  })

  // AND THE CAP IS APPLIED AFTER THE PARTITION, WHICH ONLY SHOWS WHEN THE
  // INSTALLED GROUP IS BIGGER THAN THE CAP.
  //
  // With 31 committed faces and one stored family the two repairs are redundant
  // — the repaired order puts all 32 inside the first 50 rows either way. This
  // is the fixture that tells them apart: 25 stored families on top of the 31
  // takes group 2 to 56, above `renderedFamilyLimit`. A cap over the union would
  // paint 50 rows in total and NOTHING under AVAILABLE TO INSTALL, with both
  // headings still drawn. It is also the measurement behind that group being
  // deliberately uncapped, and behind the named trigger to revisit it.
  it('renders every installed row even when the installed group is larger than the cap', async () => {
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    const from = Math.floor(webFamilies.length * 0.7)
    const deep = webFamilies.slice(from, from + 25)
    expect(deep, 'the fixture needs enough stored families to overflow a union-wide cap').toHaveLength(25)
    for (const row of deep) {
      const bytes = sfntWithNames([{ platform: 3, nameID: 0, value: `Copyright 2026 ${row.family}` }])
      const written = await opened.value.put({ ...storedOnly(await storedFaceKey(bytes), bytes), family: row.family })
      expect(written.ok, `the fixture face for ${row.family} must reach the store`).toBe(true)
    }
    globalThis.fetch = vi.fn(async () => { throw new TypeError('Failed to fetch') }) as never
    mount(commandRequest())
    await waitFor(() => expect(document.querySelectorAll('.machine-font-name')).toHaveLength(deep.length))

    fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
    const local = within(screen.getByRole('group', { name: 'AVAILABLE LOCALLY' })).getAllByRole('option')
    expect(local, 'every row under a heading promising the bytes are here must be drawn').toHaveLength(catalogueFaces.length + deep.length)
    expect(local.length, 'the fixture must really exceed the cap, or this measures nothing').toBeGreaterThan(50)
    expect(within(screen.getByRole('group', { name: 'AVAILABLE TO INSTALL' })).getAllByRole('option'), 'the third group keeps its own bound').toHaveLength(50)
  })

  // THE OTHER HALF OF BOTH RE-POINTED REFUSALS (Story 16.4).
  //
  // `storeWriteRefusal` and `lateEmbedRefusal` both tell the author to go to a
  // named panel and press a remove control. Their guards were keyed on the
  // PLACE NAME — `/AVAILABLE LOCALLY/` — which the rename invalidated, and a
  // re-pointed place-keyed guard relocates its blind spot unless the property it
  // asserts is restated with it. The property is this: the region those
  // sentences name is really on the screen, and it is really the region that
  // carries a per-face remove control.
  it('points both refusals at the one region that actually carries a remove control', async () => {
    globalThis.fetch = upstreamFetch() as never
    mount(commandRequest())
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    const panel = screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })
    // The heading the sentences name, drawn by the panel that holds the control.
    expect(within(panel).getByText(storePanelHeading)).toBeInTheDocument()
    expect(within(panel).getByRole('button', { name: 'Remove Kanit (Regular) from this machine' })).toBeInTheDocument()
    expect(storeWriteRefusal('Kanit', 'the origin is out of space'), 'the quota refusal must name this region').toContain(storePanelHeading)
    // AND THE PANEL NO LONGER ANSWERS TO THE BORROWED NAME. `AVAILABLE LOCALLY`
    // is the DROPDOWN's group heading and belongs to it alone; the panel having
    // given it back is the point of the rename, not a side effect of it.
    expect(within(panel).queryByText(/AVAILABLE LOCALLY/)).toBeNull()
  })

  // STORY 16.2's SELF-HEALING CONTRACT, WHICH STORY 16.5 BRIEFLY BROKE AND THEN
  // RESTORED — AND WHICH NOTHING HAD EVER ASSERTED.
  //
  // 16.2's matrix: *"Stored bytes fail to decode later | Corrupt entry | Entry
  // treated as absent and dropped; refetch on next pick | Self-healing, logged
  // honestly."* This story's first pass replaced that fallback with a refusal,
  // and the suite stayed green — because no test covered it. That is the same
  // gap the matrix audit found on two other rows, so the restored behaviour gets
  // a test rather than a comment.
  //
  // THE ENTRY IS DELETED BEHIND THE DESIGNER'S BACK, which is the real shape of
  // this: another tab removed it, or the store dropped it as unsound between the
  // listing and the read. The designer still believes it is installed — that is
  // the state the fallback exists for.
  //
  // BOTH HALVES ARE ASSERTED. The refetch must happen (a refusal would red the
  // embed assertion) AND the store must be healed (a fetch with no write back
  // would leave the next use fetching again for ever).
  it('refetches and heals when a stored entry has gone missing, rather than refusing first use', async () => {
    globalThis.fetch = upstreamFetch() as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    const installed = await faceRecordsOnThisMachine()
    expect(installed.map((record) => record.family)).toEqual(['Kanit'])

    // THE ENTRY VANISHES WITHOUT THE DESIGNER BEING TOLD. Its own listing still
    // has the row, so the family is still offered as one to use.
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    await opened.value.remove((installed[0] as unknown as { key: string }).key)
    expect(await faceRecordsOnThisMachine(), 'the fixture must really have emptied the store').toEqual([])

    // FIRST USE STILL WORKS. A refusal here would be the permanent local failure
    // the ruling refused: the author would have to find and press a removal
    // control on a row whose face is already gone.
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/), 'the stale listing must still offer the family').toBe(true)
    await waitFor(() => expect(embedPayloads(request).map((payload) => payload['kind'])).toEqual(['embedFontFamily', 'updateComponentProperties']))
    expect(screen.queryByRole('alert'), 'a self-healing path states no refusal').not.toBeInTheDocument()

    // AND THE STORE IS HEALED, not merely bypassed.
    await waitFor(async () => expect((await faceRecordsOnThisMachine()).map((record) => record.family)).toEqual(['Kanit']))
  })

  // MATRIX ROW: "Install a variable face | Fetched bytes carry an `fvar` table |
  // REFUSED AT INSTALL, NAMED, NOTHING STORED".
  //
  // THE SECOND HALF IS THE HALF WITH TEETH. A refusal that had already written
  // the record would satisfy every message assertion and leave a variable face
  // sitting in AVAILABLE LOCALLY, offered for a first use that Go would then
  // refuse — the exact dead end D-16.R.46 Q4 forbids. So the store is read back
  // PAST the designer, through `openFontStore`, and must be empty.
  //
  // NOTHING BEFORE THE BYTES REFUSES THIS PICK, which is what makes it a claim
  // about the `fvar` filter and not about some earlier guard: METADATA.pb
  // resolves, the licence token classifies, the OFL.txt is served, and all three
  // requests are made. Only then are the bytes looked at.
  it('refuses a variable face at install, names it, and keeps nothing on this machine', async () => {
    const bytes = variableFaceBytes()
    const fetchStub = upstreamServingAVariableFace(bytes)
    globalThis.fetch = fetchStub as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)

    // NAMED, AND THE CAUSE NAMED WITH IT. An author who is told only "could not
    // be added" has nothing to act on.
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toMatch(/^Kanit cannot be installed/)
    expect(alert.textContent).toMatch(/VARIABLE font/)
    expect(alert.textContent).toMatch(/`fvar` table/)
    expect(alert.textContent, 'the refusal must say the machine was left alone').toMatch(/Nothing was kept on this machine/)

    // THE WHOLE RESOLUTION RAN. Three round-trips: if the pick had been refused
    // by the licence table or the metadata probe this would be fewer, and the
    // assertions above would be about a different guard.
    expect(fetchStub.mock.calls.map((call) => String(call[0]).split('/').pop())).toEqual(['METADATA.pb', 'OFL.txt', 'Kanit-Regular.ttf'])

    // NOTHING STORED — read out of the store itself, not off the screen.
    expect(await faceRecordsOnThisMachine()).toEqual([])
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
    expect(within(screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })).queryAllByRole('button', { name: /^Remove / })).toEqual([])
    // AND NOTHING REACHED THE DOCUMENT EITHER.
    expect(embedPayloads(request)).toEqual([])
    // AND THE ROW STILL SAYS "INSTALL", because nothing was installed. A row
    // that had moved to the already-downloaded note would mean a record survived.
    fireEvent.change(screen.getByRole('combobox', { name: 'Font family' }), { target: { value: 'Kanit' } })
    expect(screen.getByRole('option', { name: /^Kanit\s*—\s*install on this machine$/ })).toBeInTheDocument()
  })

  // MATRIX ROW: "First use, licence contradiction | Go's nameID-13 tie refuses at
  // embed | Refusal states the face IS INSTALLED ON THIS MACHINE AND CANNOT BE
  // EMBEDDED, and points at the removal control | Property command not sent;
  // face stays installed and removable".
  //
  // THIS IS THE ONE ADMISSION CHECK THAT COULD NOT MOVE TO INSTALL, and the
  // whole reason the story is allowed to leave it in Go is that the residue is
  // DISCLOSED rather than left to surprise the author. An undisclosed dead end
  // is what D-16.R.46 Q4 forbids; a dead end the author can see and clear is a
  // stated limit. So the disclosure is the deliverable, and it is asserted here.
  //
  // THE POINTER IS CHECKED AGAINST A CONTROL THAT REALLY EXISTS. Naming a button
  // in prose is worth nothing if the button is called something else — so the
  // sentence's own words are handed to `getByRole` as an accessible name. That
  // is what stops this from being a spelling assertion.
  //
  // The engine's refusal is stubbed rather than provoked with a mislabelled
  // binary: the tie itself is Go's, proven in
  // `folio-go/internal/fontset/licencesignature_test.go` over committed bytes.
  // What is unproven anywhere else — and what this owns — is what the DESIGNER
  // does with a refusal that arrives at first use.
  it('says an installed face cannot be embedded, points at the removal control, and leaves it installed', async () => {
    globalThis.fetch = upstreamFetch() as never
    const sent: Record<string, unknown>[] = []
    const contradiction = 'fonts: font "Kanit": the face\'s own name table records it under the SIL Open Font License while this document declares Apache-2.0'
    const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
      if (operation === 'command' && payload) {
        const parsed = JSON.parse(new TextDecoder().decode(payload)) as Record<string, unknown>
        sent.push(parsed)
        // THE ENGINE REFUSES THE EMBED, exactly as the nameID-13 tie does: a
        // rejection carrying a located message, through the ordinary
        // command/diagnostic path.
        if (parsed['kind'] === 'embedFontFamily') throw { dataPath: 'fonts.Kanit', message: contradiction }
      }
      return { snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } } }
    })
    mount(request)

    // INSTALLED FIRST, and the install is unaffected: this face passes every
    // check that CAN run at install.
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    expect(sent, 'the install must not have sent anything for the engine to refuse').toEqual([])

    // FIRST USE — and this is where the refusal arrives.
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/)).toBe(true)
    const alert = await screen.findByRole('alert')

    // THE THREE THINGS THE ENGINE'S OWN SENTENCE CANNOT SAY.
    expect(alert.textContent, 'the author must be told the face IS here').toMatch(/Kanit is installed on this machine and cannot be embedded in this document/)
    expect(alert.textContent, "the engine's own reason must survive, not be replaced by a friendlier one").toContain(contradiction)
    expect(alert.textContent, 'the author must be told their file did not move').toMatch(/Nothing was written to the document/)

    // AND THE POINTER NAMES A CONTROL THAT IS REALLY ON THE SCREEN. The
    // sentence's own words are used as the accessible name, so a renamed or
    // missing affordance reds this rather than leaving prose pointing at nothing.
    const pointer = /remove it with the “(.+?)” control/.exec(alert.textContent ?? '')
    expect(pointer, `the refusal must point at a named control: ${alert.textContent}`).not.toBeNull()
    expect(screen.getByRole('button', { name: pointer![1] })).toBeInTheDocument()

    // THE PROPERTY COMMAND WAS NOT SENT. `canvas.fontFamilies` never gained the
    // chain, so committing `style.fontFamily` would have been refused anyway —
    // and sending it would put a second, unrelated refusal in front of the
    // author on top of this one.
    expect(sent.map((payload) => payload['kind'])).toEqual(['embedFontFamily'])

    // AND THE FACE IS STILL INSTALLED AND STILL REMOVABLE, which is what makes
    // the pointer above actionable rather than decorative.
    expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument()
    expect((await faceRecordsOnThisMachine()).map((record) => record.family)).toEqual(['Kanit'])
    fireEvent.click(screen.getByRole('button', { name: pointer![1] }))
    await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit \(Regular\) was removed from this machine/))
    expect(await faceRecordsOnThisMachine(), 'the control the refusal points at must actually clear the dead end').toEqual([])
  })

  // MATRIX ROW 7: "Store write fails at install | Origin quota refuses the `put` |
  // THE INSTALL FAILS AND SAYS SO | Refusal, not a silent success".
  //
  // THE ONLY ASSERTION THAT EXISTED FOR THIS ROW CALLED `storeWriteRefusal`
  // DIRECTLY and checked the string — the module-level defect Row 5's own task
  // forbids. Nothing drove a failing `put` through `installFamily` in a mounted
  // designer, so `if (kept !== undefined) return refuse(...)` could be changed to
  // `return undefined` and the whole suite stayed green while the install
  // reported success and kept nothing.
  //
  // THE REFUSAL IS INJECTED INTO THE REAL PLUMBING at the exact place a browser
  // raises it — `font-store.test.ts`'s own technique — so the transaction, the
  // request and the store's error path are all genuinely exercised.
  it('refuses the install when the origin will not take the face, and says so at the control', async () => {
    globalThis.fetch = upstreamFetch() as never
    const request = commandRequest()
    mount(request)
    const original = FakeIndexedDBObjectStore.prototype.put
    FakeIndexedDBObjectStore.prototype.put = function refuse(): never { throw new DOMException('the origin has no room left for this face', 'QuotaExceededError') }
    try {
      expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
      // IT SAYS SO, AT THE CONTROL THE AUTHOR ACTED ON — not in a log, not
      // nowhere. A silent success is what this row exists to forbid.
      const alert = await screen.findByRole('alert')
      expect(alert.textContent).toMatch(/Kanit was not installed on this machine/)
      expect(alert.textContent).toMatch(/no room left/)
      expect(alert.textContent, 'nothing may claim a document changed: no command is sent at install').toMatch(/no document was changed/)
      // AND THE PLACE IT NAMES IS ON THE SCREEN. This read `/AVAILABLE LOCALLY/`
      // until 16.4 renamed the panel; re-pointing the regex alone would have
      // moved the blind spot, so the heading is read OFF THE RENDERED PANEL and
      // matched against the sentence. `getByText` throws if the panel stops
      // drawing it, which is the half a string match cannot see.
      const panel = screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })
      const heading = within(panel).getByText(storePanelHeading)
      expect(alert.textContent, 'the author needs an action, and the removal control is the one that frees space').toContain(heading.textContent)
    } finally {
      FakeIndexedDBObjectStore.prototype.put = original
    }
    // AND NOTHING WAS KEPT AND NOTHING WAS SENT.
    expect(await faceRecordsOnThisMachine()).toEqual([])
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
    expect(embedPayloads(request)).toEqual([])
  })

  // THE FAILED HEAL DISCLOSES, AND IT IS A DEGRADATION RATHER THAN A REFUSAL.
  //
  // `storeWriteDegradation` has exactly one live path left: the write-back after
  // a refetch at first use. Deleting that `setStoreNote` made a failed heal
  // completely silent with the suite green — the author's document gets the face
  // and their machine quietly does not, for ever, with no way to tell.
  //
  // AND THE SENTENCE MUST NOT BE THE REFUSAL'S. The document HAS the face here,
  // so a message saying "no document was changed" would be false; this is the
  // one place Story 16.2's original distinction still holds.
  it('states a failed heal as a degradation, not a refusal, when the document already has the face', async () => {
    globalThis.fetch = upstreamFetch() as never
    const request = commandRequest()
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    const installed = await faceRecordsOnThisMachine()
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    await opened.value.remove((installed[0] as unknown as { key: string }).key)

    // THE REFETCH SUCCEEDS AND THE HEAL DOES NOT.
    const original = FakeIndexedDBObjectStore.prototype.put
    FakeIndexedDBObjectStore.prototype.put = function refuse(): never { throw new DOMException('the origin has no room left for this face', 'QuotaExceededError') }
    try {
      expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/)).toBe(true)
      await waitFor(() => expect(embedPayloads(request).map((payload) => payload['kind'])).toEqual(['embedFontFamily', 'updateComponentProperties']))
      await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit is in this document, but it could not be kept on this machine/))
    } finally {
      FakeIndexedDBObjectStore.prototype.put = original
    }
    expect(storeNote().textContent).toMatch(/no room left/)
    // THE DISTINCTION, ASSERTED IN BOTH DIRECTIONS. This is not the install
    // refusal's sentence, and it must never claim the document was left alone.
    expect(storeNote().textContent).not.toMatch(/was not installed on this machine/)
    expect(storeNote().textContent).not.toMatch(/no document was changed/)
    expect(screen.queryByRole('alert'), 'a degradation is not a refusal and does not alert at the control').not.toBeInTheDocument()
  })

  // THE LOCAL-TIER ARM OF THE LATE REFUSAL, which the Row 5 test cannot reach.
  //
  // `lateEmbedRefusal` takes a `removable` flag and Row 5 drives only the STORED
  // tier, so hoisting that flag to `true` passed every assertion there while a
  // local-tier refusal pointed the author at a "Remove …" control that is not on
  // the screen — the exact "prose naming a control that is not there" failure
  // that test exists to catch, one tier over.
  //
  // A bundled face ships inside the release rather than in the machine store, so
  // there is nothing to remove and the sentence says that instead of naming a
  // button the author will hunt for and not find.
  it('does not point at a removal control for a bundled face, which has none', async () => {
    const bundled = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2020 The Inter Project Authors' }])
    // The local tier reads its bytes from the release's own content-addressed
    // assets — a relative URL, never a host.
    globalThis.fetch = vi.fn(async (url: string) => {
      if (/^https?:/.test(String(url))) throw new TypeError('no third party may be contacted for a bundled face')
      return { ok: true, status: 200, arrayBuffer: async () => bundled }
    }) as never
    const contradiction = 'fonts: font "Inter": the face declares a licence its own name table contradicts'
    const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
      if (operation === 'command' && payload) {
        const parsed = JSON.parse(new TextDecoder().decode(payload)) as Record<string, unknown>
        if (parsed['kind'] === 'embedFontFamily') throw { dataPath: 'fonts.Inter', message: contradiction }
      }
      return { snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3, canvas: { ...canvas, components: [textComponent] } } }
    })
    mount(request)
    expect(pick('Inter', /^Inter\s*—\s*use it, already on this machine$/), 'a committed local-tier family must be offered for use').toBe(true)
    const alert = await screen.findByRole('alert')
    expect(alert.textContent).toMatch(/Inter is installed on this machine and cannot be embedded in this document/)
    expect(alert.textContent).toContain(contradiction)
    // THE TAIL IS THE OTHER ARM: it says why there is nothing to remove.
    expect(alert.textContent).toMatch(/ships inside this designer rather than in the machine store, so there is nothing to remove/)
    // AND IT NAMES NO CONTROL, because none exists — the store list is empty.
    expect(alert.textContent, 'a bundled face has no removal control to point at').not.toMatch(/remove it with the/)
    expect(within(screen.getByRole('group', { name: 'Typefaces downloaded to this machine' })).queryAllByRole('button', { name: /^Remove / })).toEqual([])
  })

  // THE DOCUMENT MAY BE REPLACED WHILE THE EMBED IS IN FLIGHT, AND THE PROPERTY
  // COMMIT MUST NOT FOLLOW IT INTO THE NEW ONE.
  //
  // Every other async commit in `App.tsx` guards this way and there are tests
  // named for it; the third arm is a NEW async path and did not inherit it. It
  // matters because `applyProperties` SENDS FIRST AND GUARDS AFTER — it
  // dispatches the command unconditionally and only declines to install the
  // RESULT — so a stale commit is not a dropped response, it is
  // `updateComponentProperties` reaching the engine carrying the PREVIOUS
  // document's element ids.
  it('does not commit the property when the document is replaced while the embed is in flight', async () => {
    globalThis.fetch = upstreamFetch() as never
    let release: () => void = () => {}
    const held = new Promise<void>((resolve) => { release = resolve })
    const sent: Record<string, unknown>[] = []
    let holdTheEmbed = false
    const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
      if (operation === 'command' && payload) {
        const parsed = JSON.parse(new TextDecoder().decode(payload)) as Record<string, unknown>
        sent.push(parsed)
        if (holdTheEmbed && parsed['kind'] === 'embedFontFamily') await held
      }
      return { snapshot: { documentState: 'loaded' as const, revision: operation === 'command' ? 2 : 1, byteLength: 3, canvas: { ...canvas, components: [textComponent] } } }
    })
    mount(request)
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())

    holdTheEmbed = true
    expect(pick('Kanit', /^Kanit\s*—\s*use it, already downloaded to this machine$/)).toBe(true)
    await waitFor(() => expect(sent.map((payload) => payload['kind'])).toEqual(['embedFontFamily']))

    // THE DOCUMENT IS REPLACED WHILE THE EMBED IS STILL IN FLIGHT.
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())
    release()

    // AND THE PROPERTY NEVER FOLLOWS IT. Settled twice over — once through the
    // event loop and once through a fresh selection round-trip — so this is a
    // claim about the guard and not about when the assertion happened to run.
    await new Promise((resolve) => setTimeout(resolve, 0))
    fireEvent.click(screen.getByLabelText('text component e1'))
    await new Promise((resolve) => setTimeout(resolve, 0))
    expect(sent.map((payload) => payload['kind']), 'the property commit belongs to a document that is no longer open').toEqual(['embedFontFamily'])
    // AND NOTHING IS SHOWN AGAINST A CONTROL THE AUTHOR IS NO LONGER LOOKING AT.
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  // THE DEGRADED CONFIRM WARNING, ASSERTED AGAINST THE CODE THAT DECIDES IT AND
  // NOT THE CODE THAT DISPLAYS IT.
  //
  // `FontBrowser.test.tsx` proves the dialog renders the warning when it is
  // HANDED `storeKeepsFaces={false}`. That proves a component renders what it is
  // told and nothing about what it is told: hardcode `storeKeepsFaces={true}` at
  // the `App.tsx` call site and that test stays green while a private-window
  // author sees `Install 5 on this machine` on a button that writes five faces
  // into their document. The clause would exist, a test would assert it, and the
  // alibi would still hold.
  //
  // SO THIS DRIVES THE WHOLE DESIGNER WITH NO STORE AT ALL and reads the name off
  // the real control, which is the only version of the claim the acceptance
  // criterion actually makes.
  it('computes the degraded confirm warning from a store it really could not open', async () => {
    Reflect.deleteProperty(globalThis, 'indexedDB')
    const fontSet = installStubFontSet()
    try {
      globalThis.fetch = upstreamFetch() as never
      mount(commandRequest())
      await waitFor(() => expect(storeNote().textContent).toMatch(/not letting the designer keep typefaces on this machine/))

      fireEvent.focus(screen.getByRole('combobox', { name: 'Font family' }))
      fireEvent.click(screen.getByRole('button', { name: /^Add fonts…/ }))
      const dialog = screen.getByRole('dialog', { name: 'Font browser' })
      fireEvent.change(screen.getByRole('textbox', { name: 'Search fonts' }), { target: { value: 'Kanit' } })
      fireEvent.click(await within(dialog).findByRole('button', { name: 'Install Kanit on this machine' }))

      // THE COUNT TRAVELS IN THE NAME OF THE CONTROL THAT DOES IT, and the name
      // CONTAINS the visible label rather than replacing it (WCAG 2.5.3), so a
      // speech-input author can still say what they can see.
      const confirm = within(dialog).getByRole('button', { name: 'Install 1 on this machine — this browser will not keep fonts, so confirming adds 1 family to this document' })
      expect(confirm.textContent).toBe('Install 1 on this machine')
      expect(within(dialog).getByText(/This browser will not keep typefaces, so these go straight into the document/)).toBeInTheDocument()
    } finally {
      fontSet.restore()
    }
  })

  // STORAGE THAT CANNOT BE OPENED LEAVES A WORKING DESIGNER — STORY 16.2's
  // CONTRACT, KEPT, BY DEGRADING TO THE PRE-16.5 MODEL (orchestrator ruling,
  // 2026-09-03).
  //
  // BEHAVIOUR-CHANGED, THEN CORRECTED, AND THE CORRECTION IS THE INTERESTING
  // PART. This story first made an unopenable store REFUSE the install, on the
  // ground that installing IS the store write. That was refused in review: it
  // contradicts 16.2's locked *"still works and says what is degraded"*, and it
  // would mean a private window could add no font at all. But 16.2's degradation
  // taken literally fails too — "install anyway, keep nothing" leaves a face that
  // can never be used, the dead end D-16.R.46 Q4 forbids. The third option is
  // what is asserted here: this browser gets the OLD MODEL, the pick embeds
  // directly, and the note SAYS THAT IS WHAT HAPPENED.
  //
  // TWO CLAIMS, AND BOTH ARE LOAD-BEARING. A degradation that embedded silently
  // would leave the author unable to tell why this browser behaves unlike the
  // last one; a message with no embed behind it would be a lie.
  it('still designs and still adds fonts when the store cannot be opened at all, by embedding directly and saying so', async () => {
    Reflect.deleteProperty(globalThis, 'indexedDB')
    const fetchStub = upstreamFetch()
    globalThis.fetch = fetchStub as never
    const request = commandRequest()
    mount(request)
    await waitFor(() => expect(storeNote().textContent).toMatch(/not letting the designer keep typefaces on this machine/))
    // THE PICK STILL PUTS THE FONT IN THE DOCUMENT. IndexedDB is a convenience
    // here and never a dependency.
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(embedPayloads(request).map((payload) => payload['kind'])).toEqual(['embedFontFamily']))
    expect(embedPayloads(request)[0]!['family']).toBe('Kanit')
    // AND THE PROPERTY IS STILL NOT COMMITTED: this is the fork's SECOND arm
    // running in its old shape, not the third. The degradation may not quietly
    // fuse the two decisions Story 8.6 kept apart.
    expect(embedPayloads(request).map((payload) => payload['kind'])).toEqual(['embedFontFamily'])
    // AND THE AUTHOR IS TOLD WHICH MODEL THEY ARE IN, in the product's own terms.
    await waitFor(() => expect(storeNote().textContent).toMatch(/Kanit went straight into this document/))
    expect(storeNote().textContent).toMatch(/nowhere to keep it there is nothing to install/)
    expect(storeNote().textContent, 'the author must be told their document is still self-contained').toMatch(/carries its own copy/)
    expect(screen.getByText(/No typefaces have been downloaded to this machine yet/)).toBeInTheDocument()
  })

  // THE MACHINE-SCOPED PREVIEW REGISTRATION, PINNED IN BOTH DIRECTIONS.
  //
  // `App.tsx` keeps TWO registrations — the document's carried faces and this
  // machine's stored faces — and unions them into `paintableFaces`, which is
  // what the canvas is given. Until this test, that union was worth nothing to
  // any assertion: a reviewer mutation-proved it by handing the canvas
  // `carriedFaces` instead of `paintableFaces` at BOTH call sites, and the
  // whole suite stayed green. The one machine-scoped registration in the
  // application had no test at all.
  //
  // THE FIXTURE IS THE ASSERTION. The document declares ONE chain entry, a
  // shipped face, so it carries no face of its own and `carriedFaces` is empty
  // for the life of this test. The store holds one face, and the projection
  // attributes its fragment to that face's asset key. So the fragment can only
  // acquire a family if the set the canvas is given is LARGER than the set the
  // document carries — which is the exact property the mutation removes.
  //
  // Neither registration is asked to do the other's job: nothing here goes
  // through the engine's `asset` operation, and the assertion below that the
  // engine was never asked for one is what says so.
  it('registers a face this machine holds but this document does not carry, and paints with it', async () => {
    const bytes = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2026 A Face Only This Machine Has' }])
    const key = await storedFaceKey(bytes)
    const opened = await openFontStore(globalThis.indexedDB)
    if (!opened.ok) throw new Error(opened.reason)
    const written = await opened.value.put(storedOnly(key, bytes))
    expect(written.ok, 'the fixture face must really be in the store before the designer opens it').toBe(true)

    const fontSet = installStubFontSet()
    try {
      // Every `asset` request is recorded, because the assertion at the end of
      // this test is that there were NONE: the document carries no face, so the
      // document-scoped registration has nothing to ask the engine for.
      const assetRequests: string[] = []
      const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
        if (operation === 'asset') assetRequests.push(new TextDecoder().decode(payload))
        return { snapshot: { documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: storedOnlyFaceCanvas(key) } }
      })
      const view = render(<App engine={engine(request)} blankBytes={new Uint8Array([1, 2, 3]).buffer} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: storedOnlyFaceCanvas(key) }} />)
      // THE POSITIVE SETTLE CONDITION, so what follows is read on a chain that
      // has actually reached the machine registration rather than on a race:
      // the face is in the page's font set, under the family derived from the
      // content address, and exactly once.
      await waitFor(() => expect(fontSet.added).toEqual([embeddedFaceFamily(key)]))
      // AND IT REACHED THE CANVAS. This is the assertion the mutation reds: with
      // `carriedFaces` passed instead of `paintableFaces` the fragment keeps the
      // stylesheet's declared stack and this stays `''` forever.
      const painted = () => Array.from(view.container.querySelectorAll('.canvas-text-fragment')) as HTMLElement[]
      expect(painted().length).toBe(1)
      await waitFor(() => expect(painted()[0]!.style.fontFamily, 'a face held only by the machine store must still reach the canvas').toBe(embeddedFaceFamily(key)))
      // THE DOCUMENT CARRIED NOTHING, which is what makes the line above a
      // claim about the machine registration and not about the document one.
      // The document-scoped effect asks the ENGINE for bytes; it was never
      // asked, because the document declares no carried entry.
      expect(assetRequests, 'the document carries no face, so the engine must never be asked for one').toEqual([])
    } finally {
      fontSet.restore()
    }
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
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
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
    // AND IT IS NOT THE OFFLINE SENTENCE, which this designer has no ground to
    // say: a timeout does not establish that the network is down.
    expect(alert.textContent).not.toMatch(/without a network connection/)
    // NOR THE OPPOSITE CLAIM, which it has no ground for either. A timeout
    // cannot know the network is reachable — the same abort fires on a
    // blackholed link, a hanging DNS lookup and a connection merely too slow.
    expect(alert.textContent, 'a timeout cannot diagnose the network in either direction').not.toMatch(/network is reachable/i)
  })

  // (b) THE SECOND CARRIER — the instance D-16.R.15 predicted. The hold is
  // backed by BOTH a ref and React state, and the document-reset path once
  // cleared only the state copy: replacing the document mid-stall left the ref
  // stuck `true` for the session, with the control looking perfectly enabled
  // while every later pick silently did nothing. Proving (a) alone re-proves a
  // shape that was already fixed; this is the one that can regress.
  //
  // RE-ANCHORED BY STORY 16.5 (MECHANICAL). The witness for "the later pick
  // really happened" used to be an embed command, because a pick embedded. A
  // pick now INSTALLS, so the witness is the store row appearing — the same
  // observable the install path actually produces. The claim is unchanged: a
  // hold left behind by the replaced document would make the later pick do
  // NOTHING, and doing nothing is exactly what a witness must be able to see.
  it('is clear after the document is replaced mid-stall, so a later pick still installs', async () => {
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
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
    await waitFor(() => expect(stall).toBeDefined())

    // THE DOCUMENT IS REPLACED WHILE THE PICK IS STILL STALLED.
    fireEvent.click(screen.getByRole('button', { name: 'Start blank' }))
    await waitFor(() => expect(screen.getByText('Untitled template')).toBeInTheDocument())
    stall!()

    fireEvent.click(screen.getByLabelText('text component e1'))
    const probesBefore = fetchStub.mock.calls.length
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/), 'the pick must still be offered in the new document').toBe(true)
    // THE FACE REACHES THE MACHINE. A stuck hold returns before the fetch, so
    // neither of these can be satisfied by a pick that silently did nothing.
    await waitFor(() => expect(screen.getByText(/Kanit/, { selector: '.machine-font-name' })).toBeInTheDocument())
    expect(fetchStub.mock.calls.length).toBeGreaterThan(probesBefore)
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
    expect(pick('Kanit', /^Kanit\s*—\s*install on this machine$/)).toBe(true)
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
