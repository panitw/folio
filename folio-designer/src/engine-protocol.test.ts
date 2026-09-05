import { describe, expect, it } from 'vitest'
import { ENGINE_PROTOCOL_VERSION, LOCALE_TAGS, MAX_CANVAS_BODY_TEXT_LINES, MAX_ENGINE_CONTENT_WINDOWS, MAX_ENGINE_FONT_CHAIN_ENTRIES, MAX_ENGINE_FONT_FAMILIES, MAX_CANVAS_PROPERTY_STRING, MAX_ENGINE_BINDING_LENGTH, MAX_ENGINE_DATA_PATH_LENGTH, MAX_ENGINE_ELEMENT_ID_LENGTH, MAX_ENGINE_PAYLOAD_BYTES, MAX_ENGINE_RENDER_PDF_BYTES, deepFreeze, parseInbound, parseRequest } from './engine-protocol'

// face() builds the PROJECTED shape of a named-face chain entry (Story 8.3:
// an entry is a discriminated object, not a string). A named face carries no
// family and no style — its name is its identity.
const face = (name: string) => ({ face: name, assetKey: '', family: '', style: '' })

const canvas = { width: 1000, height: 2000, orientation: 'portrait', preset: 'custom', locale: 'th', utcOffset: '+07:00', marginTop: 0, marginRight: 0, marginBottom: 0, marginLeft: 0, gridIncrement: 100, commandWidth: 1000, commandHeight: 2000, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }], defaultFontSize: 12000, defaultLineSpacing: 1000, contentWindowHeight: 1800, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader', x: 0, y: 0, width: 1000, height: 100 }, { name: 'content', x: 0, y: 100, width: 1000, height: 1800 }, { name: 'pageFooter', x: 0, y: 1900, width: 1000, height: 100 }], components: [] }

describe('canvas projection protocol guard', () => {
  // STORY 12.2: THE DOCUMENT'S TWO DECLARED FORMATTING AUTHORITIES.
  //
  // The projection gained `locale` and `utcOffset` so the PAGE SETUP panel can
  // show what the engine holds. `hasOnly` cannot carry them: it is a SUBSET
  // check, so a key Go simply failed to send passes it and arrives at the panel
  // as `undefined` — a locale row with no value, and an offset row that would
  // send the string "undefined" straight back to the engine. THE ABSENCE CASES
  // BELOW ARE THE ONLY THING THAT CATCHES THAT, and they are the failure this
  // story could otherwise have shipped in silence.
  it('requires both document-settings fields, and requires the locale to be one of AD-12\'s four tags', () => {
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    // The positive control first: the fixture as it stands is accepted, so
    // every rejection below is attributable to the patch and not to the base.
    expect(projection(canvas)).toBeDefined()
    // EVERY TAG IN THE CLOSED SET IS ACCEPTED, enumerated from LOCALE_TAGS
    // rather than written out again — a guard narrowed to one tag would blank
    // the canvas for the documents this story exists to make authorable.
    expect(LOCALE_TAGS.length).toBeGreaterThan(0)
    for (const tag of LOCALE_TAGS) expect(projection({ ...canvas, locale: tag })).toBeDefined()
    // ABSENT. Neither key can be caught by hasOnly.
    const { locale: _locale, ...noLocale } = canvas
    const { utcOffset: _offset, ...noOffset } = canvas
    expect(projection(noLocale)).toBeUndefined()
    expect(projection(noOffset)).toBeUndefined()
    // ILLEGAL. A tag outside AD-12's set — one Go's loader would refuse — and
    // an empty offset, which is what an emptied box would round-trip as if the
    // engine ever echoed one back.
    expect(projection({ ...canvas, locale: 'fr' })).toBeUndefined()
    expect(projection({ ...canvas, locale: 'EN' })).toBeUndefined()
    expect(projection({ ...canvas, locale: '' })).toBeUndefined()
    expect(projection({ ...canvas, locale: 7 })).toBeUndefined()
    expect(projection({ ...canvas, locale: null })).toBeUndefined()
    expect(projection({ ...canvas, utcOffset: '' })).toBeUndefined()
    expect(projection({ ...canvas, utcOffset: 7 })).toBeUndefined()
    expect(projection({ ...canvas, utcOffset: null })).toBeUndefined()
    expect(projection({ ...canvas, utcOffset: 'x'.repeat(MAX_CANVAS_PROPERTY_STRING + 1) })).toBeUndefined()
    // AND THE OFFSET'S GRAMMAR IS NOT RESTATED HERE. ±HH:MM is the engine's
    // rule, asked through the one predicate its loader and its command door
    // share; a browser-side copy could refuse a snapshot Go legitimately sent,
    // and the symptom would be a permanently blank canvas. So a value this side
    // cannot judge is ACCEPTED on shape alone.
    expect(projection({ ...canvas, utcOffset: '+99:99' })).toBeDefined()
    expect(projection({ ...canvas, utcOffset: 'Z' })).toBeDefined()
    // The extra-key direction, on the two new keys' own account: a THIRD
    // document-settings key Go started sending drops the snapshot.
    expect(projection({ ...canvas, timeZone: 'Asia/Bangkok' })).toBeUndefined()
  })

  it('accepts and deeply freezes the exact three bounded bands', () => {
    const inbound = parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas } })
    expect(inbound).toBeDefined()
    const frozen = deepFreeze(canvas)
    expect(Object.isFrozen(frozen.bands)).toBe(true)
    expect(Object.isFrozen(frozen.bands[0])).toBe(true)
  })

  it.each([
    [{ ...canvas, bands: [canvas.bands[1], canvas.bands[0], canvas.bands[2]] }],
    [{ ...canvas, bands: [{ ...canvas.bands[0], name: 'content' }, ...canvas.bands.slice(1)] }],
    [{ ...canvas, bands: [{ ...canvas.bands[0], x: 1_000 }, ...canvas.bands.slice(1)] }],
    [{ ...canvas, width: Number.MAX_SAFE_INTEGER + 1 }],
  ])('rejects structurally false paint geometry', (bad) => {
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })).toBeUndefined()
  })

  it.each([
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true }, { id: 'e1', type: 'rect', band: 'content', x: 20, y: 0, width: 10, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 991, y: 0, width: 10, height: 10, resizable: true }] },
    // The two REPEATING bands still cap vertically, and nothing in this file
    // exercised that before Story 7.5: every content component here sat at
    // y: 0, so the vertical conjunct was never reached and a Y-only lift
    // would have left the whole file green and vacuous.
    { ...canvas, components: [{ id: 'e1', type: 'rect', band: 'pageHeader', x: 0, y: 0, width: 10, height: 101, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'rect', band: 'pageFooter', x: 0, y: 95, width: 10, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'table', band: 'content', x: 0, y: 0, width: 0, height: 10, resizable: true }] },
    { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: false }] },
  ])('rejects ambiguous, out-of-band, or incoherent component paint geometry', (bad) => {
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })).toBeUndefined()
  })

  it('admits a content component below the foot of page one, and only in the content band', () => {
    // Story 7.5. The content band is a COLUMN: a component five windows down
    // is on a later page, not outside the document. Dropping the snapshot for
    // it would terminate the worker and blank the canvas with no attributable
    // error, so the browser's copy of the band-containment gate has to lift
    // with Go's — in the same commit, which engine-bounds-mirror.test.ts
    // reads both sides to enforce.
    const tall = { ...canvas, components: [{ id: 'e1', type: 'rect', band: 'content', x: 0, y: 9_000, width: 10, height: 10, resizable: true }] }
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: tall } })).toBeDefined()
    // And the lift is vertical only: the column is unbounded downwards, never
    // sideways.
    const wide = { ...canvas, components: [{ id: 'e1', type: 'rect', band: 'content', x: 0, y: 9_000, width: 1_001, height: 10, resizable: true }] }
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: wide } })).toBeUndefined()
  })

  it('requires the engine-owned window height and window count', () => {
    // WHICH GUARD DOES WHAT, because the two are easy to confuse and only one
    // of them is doing the work here. `hasOnly` is a SUBSET check — it rejects
    // keys the build does not know, not keys it is missing (`hasExactKeys` is
    // the strict sibling, and the canvas is not checked with it). What rejects
    // an OMITTED field is `integer(key, true)`, which reads `undefined` and
    // fails both `Number.isSafeInteger` and `> 0`. So both fields are required
    // and both must be strictly positive: a zero count is not a document,
    // since internal/layout answers ONE page for a column with no items.
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    const { contentWindowCount: _count, ...noCount } = canvas
    const { contentWindowHeight: _height, ...noHeight } = canvas
    expect(projection(noCount)).toBeUndefined()
    // The second half, which the first version of this test never reached:
    // omitting the HEIGHT has to be refused on the same terms, or a Go build
    // that shipped one field and not the other would be admitted.
    expect(projection(noHeight)).toBeUndefined()
    for (const bad of [0, -1, 1.5, '4', null]) {
      expect(projection({ ...canvas, contentWindowCount: bad })).toBeUndefined()
      expect(projection({ ...canvas, contentWindowHeight: bad })).toBeUndefined()
    }
    expect(projection({ ...canvas, contentWindowCount: 4, contentWindowOrigins: [0, 1800, 3600, 5400] })).toBeDefined()
  })

  // Story 7.6. The origins are what the canvas draws every sheet boundary
  // from, so a malformed sequence is not a cosmetic problem: the browser
  // would either draw a boundary in the wrong place or, worse, derive one
  // itself. Every shape below is refused by the SAME path as any other
  // malformed projection field — parseInbound returns undefined and the
  // whole snapshot is discarded — which is what makes an honest Go field the
  // only way to get a drawing at all.
  it('requires one window origin per window, starting at zero and strictly increasing', () => {
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    const three = { ...canvas, contentWindowCount: 3, contentWindowOrigins: [0, 1800, 5400] }
    // The positive case first, so every rejection below is a discrimination
    // rather than a fixture that never parsed.
    expect(projection(three)).toBeDefined()
    // Windows do NOT have to be one window apart: the engine advances to the
    // top of the first item that did not fit, so a declared gap is a legal
    // and expected sequence. A validator that required a fixed stride would
    // be the forbidden closed form wearing a guard's clothes.
    expect(projection({ ...canvas, contentWindowCount: 2, contentWindowOrigins: [0, 7_280_000] })).toBeDefined()
    const { contentWindowOrigins: _origins, ...noOrigins } = canvas
    const { contentWindowCountIsExact: _exact, ...noExact } = canvas
    // Absent entirely. `hasOnly` is a subset check and says nothing about a
    // MISSING key; these two value predicates are the whole guard.
    expect(projection(noOrigins)).toBeUndefined()
    expect(projection(noExact)).toBeUndefined()
    // A nil Go slice marshals to null, not to [].
    expect(projection({ ...canvas, contentWindowOrigins: null })).toBeUndefined()
    expect(projection({ ...canvas, contentWindowOrigins: 1 })).toBeUndefined()
    expect(projection({ ...canvas, contentWindowOrigins: [] })).toBeUndefined()
    // Wrong length, both directions.
    expect(projection({ ...three, contentWindowOrigins: [0, 1800] })).toBeUndefined()
    expect(projection({ ...three, contentWindowOrigins: [0, 1800, 5400, 9000] })).toBeUndefined()
    // Not starting at zero: window one begins at the top of the column,
    // unconditionally, and internal/layout guarantees it.
    expect(projection({ ...three, contentWindowOrigins: [900, 1800, 5400] })).toBeUndefined()
    // Not increasing, and not strictly increasing.
    expect(projection({ ...three, contentWindowOrigins: [0, 5400, 1800] })).toBeUndefined()
    expect(projection({ ...three, contentWindowOrigins: [0, 1800, 1800] })).toBeUndefined()
    // Entries that are not safe non-negative integers.
    for (const bad of [-1, 1.5, '1800', null, Number.MAX_SAFE_INTEGER + 1]) {
      expect(projection({ ...three, contentWindowOrigins: [0, 1800, bad] })).toBeUndefined()
    }
    // The honesty flag is a boolean and nothing else — never a truthy string
    // a disclosure would then render. 0 and '' matter twice over here: the
    // sense is inverted from the field this replaced, so a falsy non-boolean
    // slipping through would read as "do not trust this count" on a document
    // that is exact.
    for (const bad of [0, 1, 'true', null, '']) {
      expect(projection({ ...canvas, contentWindowCountIsExact: bad })).toBeUndefined()
    }
    expect(projection({ ...canvas, contentWindowCountIsExact: false })).toBeDefined()
    // The declared cap, at its edge on both sides.
    const long = (count: number) => ({ ...canvas, contentWindowCount: count, contentWindowOrigins: Array.from({ length: count }, (_value, index) => index * 1800) })
    expect(projection(long(MAX_ENGINE_CONTENT_WINDOWS))).toBeDefined()
    expect(projection(long(MAX_ENGINE_CONTENT_WINDOWS + 1))).toBeUndefined()
  })

  // STORY 8.1. fontChains is the first projection field that carries the
  // document's font MAP rather than a name list, and both of its failure modes
  // are silent: hasOnly is a SUBSET check, so a key Go sends and this file does
  // not list drops the whole snapshot and blanks the canvas with no
  // attributable error; and a fontChains/fontFamilies disagreement would let
  // the chain editor offer a name the engine does not hold. Both are measured
  // here, not assumed.
  it('accepts the projected font chains only when they agree with fontFamilies', () => {
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    const two = { ...canvas, fontFamilies: ['body', 'heading'], fontChains: [{ name: 'body', entries: [face('Noto Sans')] }, { name: 'heading', entries: [face('Noto Sans'), face('Noto Sans Thai')] }] }
    expect(projection(two)).toBeDefined()
    // THE KEY, BOTH WAYS. Go omits it; and Go sends it while the guard's own
    // hasOnly list does not name it — the second direction is asserted against
    // the real list by reading engine-protocol.ts in
    // canvas_projection_wire_test.go, and here by the extra-key case below.
    const { fontChains: _chains, ...noChains } = canvas
    expect(projection(noChains)).toBeUndefined()
    expect(projection({ ...canvas, extraProjectionKey: 1 })).toBeUndefined()
    // And the POSITIVE case the extra-key assertion used to be conflated
    // with: a document declaring `"fonts": {}` — the component-asset-import
    // and image-embed fixtures both do — projects no chains and no families,
    // and that is VALID. Carrying the extra key made that assertion pass on
    // the key alone, so it said nothing either way about a zero-chain
    // projection, and a guard that rejected one would have gone unnoticed.
    expect(projection({ ...canvas, fontChains: [], fontFamilies: [] })).toBeDefined()
    // Disagreement with fontFamilies, in each of its three shapes: a different
    // name, a different length, and the same names in a different order.
    expect(projection({ ...canvas, fontChains: [{ name: 'brand', entries: [face('Noto Sans')] }] })).toBeUndefined()
    expect(projection({ ...two, fontChains: [two.fontChains[0]] })).toBeUndefined()
    expect(projection({ ...two, fontChains: [two.fontChains[1], two.fontChains[0]] })).toBeUndefined()
    // A chain with no entries is not one Go projects, because it is not one
    // style.fontFamily may name.
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [] }] })).toBeUndefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [face('')] }] })).toBeUndefined()
    // Shape and bounds.
    expect(projection({ ...canvas, fontChains: null })).toBeUndefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body' }] })).toBeUndefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [face('Noto Sans')], extra: 1 }] })).toBeUndefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [7] }] })).toBeUndefined()
    const entries = (count: number) => ({ ...canvas, fontChains: [{ name: 'body', entries: Array.from({ length: count }, (_value, index) => face(`face-${index}`)) }] })
    expect(projection(entries(MAX_ENGINE_FONT_CHAIN_ENTRIES))).toBeDefined()
    expect(projection(entries(MAX_ENGINE_FONT_CHAIN_ENTRIES + 1))).toBeUndefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [face('f'.repeat(MAX_CANVAS_PROPERTY_STRING))] }] })).toBeDefined()
    expect(projection({ ...canvas, fontChains: [{ name: 'body', entries: [face('f'.repeat(MAX_CANVAS_PROPERTY_STRING + 1))] }] })).toBeUndefined()
    // The count bound the mirror test ties to Go's maxCanvasFontFamilies, at
    // its edge — fontFamilies and fontChains cross the boundary together.
    const families = (count: number) => {
      const names = Array.from({ length: count }, (_value, index) => `f${String(index).padStart(6, '0')}`)
      return { ...canvas, fontFamilies: names, fontChains: names.map((name) => ({ name, entries: [face('Noto Sans')] })) }
    }
    expect(projection(families(MAX_ENGINE_FONT_FAMILIES))).toBeDefined()
    expect(projection(families(MAX_ENGINE_FONT_FAMILIES + 1))).toBeUndefined()
  })

  // Story 8.3. A chain entry is a DISCRIMINATED OBJECT, not a string. The Go
  // projection and this guard changed in one commit for the usual reason: the
  // old `typeof face === 'string'` clause rejected an object entry outright,
  // isCanvas returned false, parseInbound returned undefined, engine-client
  // terminated the worker, and the canvas was permanently blank with nothing
  // to attribute it to.
  it('accepts the projected chain ENTRY only in its discriminated shape', () => {
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    const chain = (...entries: ReadonlyArray<unknown>) => ({ ...canvas, fontChains: [{ name: 'body', entries }] })
    const key = 'c'.repeat(64)

    // Both legal shapes, and a chain mixing them.
    expect(projection(chain(face('Noto Sans')))).toBeDefined()
    expect(projection(chain({ face: '', assetKey: key, family: 'Inter', style: 'Regular' }))).toBeDefined()
    expect(projection(chain({ face: '', assetKey: key, family: 'Inter', style: '' }))).toBeDefined()
    expect(projection(chain(face('Noto Sans'), { face: '', assetKey: key, family: 'Inter', style: 'Regular' }))).toBeDefined()

    // Not an object at all: the shapes the pre-8.3 wire could carry, and the
    // ones a hostile or stale sender might.
    expect(projection(chain('Noto Sans'))).toBeUndefined()
    expect(projection(chain(7))).toBeUndefined()
    expect(projection(chain(null))).toBeUndefined()
    expect(projection(chain([face('Noto Sans')]))).toBeUndefined()

    // The key set is EXACT, both directions: a key Go stops sending fails as
    // surely as a key Go starts sending.
    expect(projection(chain({ assetKey: key, family: 'Inter', style: 'Regular' }))).toBeUndefined()
    expect(projection(chain({ face: 'Noto Sans', assetKey: '', family: '', style: '', weight: 700 }))).toBeUndefined()
    expect(projection(chain({ face: 'Noto Sans', assetKey: '', family: '' }))).toBeUndefined()

    // Every value is a string.
    expect(projection(chain({ face: 'Noto Sans', assetKey: null, family: '', style: '' }))).toBeUndefined()
    expect(projection(chain({ face: '', assetKey: key, family: 7, style: '' }))).toBeUndefined()

    // EXACTLY ONE of face and assetKey. Neither is an entry of no kind;
    // both is an entry of two.
    expect(projection(chain({ face: '', assetKey: '', family: '', style: '' }))).toBeUndefined()
    expect(projection(chain({ face: 'Noto Sans', assetKey: key, family: 'Inter', style: '' }))).toBeUndefined()

    // A named face carries no display strings — its name IS its identity —
    // and an embedded one always carries a family, because Go falls back to
    // the asset key rather than sending a name the panel cannot draw.
    expect(projection(chain({ face: 'Noto Sans', assetKey: '', family: 'Inter', style: '' }))).toBeUndefined()
    expect(projection(chain({ face: 'Noto Sans', assetKey: '', family: '', style: 'Regular' }))).toBeUndefined()
    expect(projection(chain({ face: '', assetKey: key, family: '', style: 'Regular' }))).toBeUndefined()

    // The per-string bound applies to EVERY projected string, not only the
    // face name — a bound on three of four fields is a bound on nothing.
    const long = 'f'.repeat(MAX_CANVAS_PROPERTY_STRING + 1)
    expect(projection(chain({ face: '', assetKey: key, family: long, style: '' }))).toBeUndefined()
    expect(projection(chain({ face: '', assetKey: key, family: 'Inter', style: long }))).toBeUndefined()
    expect(projection(chain({ face: '', assetKey: long, family: 'Inter', style: '' }))).toBeUndefined()
  })

  // DW-70. Go sorts these keys with slices.Sorted over Go strings — BY BYTE —
  // and those keys are the canonical `.folio`'s own `fonts` key order under
  // AD-9, so Go's order IS the document's and is NORMATIVE. The guard used
  // `>=` on JavaScript strings, which compares UTF-16 CODE UNITS, and the two
  // disagree wherever a name mixes the astral planes with U+E000-U+FFFF: a
  // surrogate pair sorts BELOW U+E000 in UTF-16 and ABOVE it in UTF-8.
  //
  // The consequence was not a dropped frame. isCanvas false makes parseInbound
  // return undefined, which engine-client raises as PROTOCOL_INVALID, which
  // TERMINATES the worker and leaves the canvas permanently blank. Story 8.2
  // is what lets an author type a chain name, so two keystrokes reached it.
  it('accepts the projected chain names in Go\'s byte order, not the browser\'s UTF-16 order', () => {
    const projection = (patch: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: patch } })
    const ordered = (...names: ReadonlyArray<string>) => ({ ...canvas, fontFamilies: names, fontChains: names.map((name) => ({ name, entries: [face('Noto Sans')] })) })
    // THE MEASURED PAIR. '\uE000' is EE 80 80 and '\u{1F600}' is F0 9F 98 80,
    // so Go sends them in this order. In UTF-16 the emoji begins 0xD83D, which
    // is BELOW 0xE000 — so `>=` called this pair out of order and dropped the
    // whole snapshot.
    expect(projection(ordered('\uE000', '\u{1F600}'))).toBeDefined()
    // And the reverse pair is still rejected, so the fix widened the accepted
    // set rather than removing the check: a genuinely out-of-order projection
    // is a channel fault and is still not trusted.
    expect(projection(ordered('\u{1F600}', '\uE000'))).toBeUndefined()
    // The ordinary cases the check has always covered, both ways.
    expect(projection(ordered('body', 'heading'))).toBeDefined()
    expect(projection(ordered('heading', 'body'))).toBeUndefined()
    // Equal names are neither ascending nor unique.
    expect(projection(ordered('body', 'body'))).toBeUndefined()
    // A prefix sorts before its extension in both orders, and does here.
    expect(projection(ordered('body', 'bodyweight'))).toBeDefined()
    expect(projection(ordered('bodyweight', 'body'))).toBeUndefined()
    // A second astral pair, on the other side of the boundary: U+FFFD (EF BF
    // BD) still precedes the emoji in byte order.
    expect(projection(ordered('\uFFFD', '\u{1F600}'))).toBeDefined()
    expect(projection(ordered('\u{1F600}', '\uFFFD'))).toBeUndefined()
  })

  it('bounds opaque producer failure provenance at the main-thread boundary', () => {
    const response = (error: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: false, error })
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH), dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH) })).toBeDefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 'e'.repeat(MAX_ENGINE_ELEMENT_ID_LENGTH + 1) })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: 'p'.repeat(MAX_ENGINE_DATA_PATH_LENGTH + 1) })).toBeUndefined()
    expect(response({ code: 'C'.repeat(97), message: 'invalid' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'm'.repeat(513) })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: '' })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', elementId: 7 })).toBeUndefined()
    expect(response({ code: 'COMPONENT_INVALID', message: 'invalid', dataPath: [] })).toBeUndefined()
  })

  it('rejects surplus authority-bearing fields at every projection level', () => {
    const response = (bad: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: bad } })
    expect(response({ ...canvas, style: {} })).toBeUndefined()
    expect(response({ ...canvas, bands: [{ ...canvas.bands[0], extra: true }, ...canvas.bands.slice(1)] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, Extra: {} }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, fontFamily: 'body' }] })).toBeUndefined()
  })

  it('accepts only bounded, ordered engine text paint and rejects browser-shaped substitutes', () => {
    const textPaint = { overflow: false, truncated: false, lines: [{ top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'engine line', x: 0 }] }] }
    const response = (projection: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: projection } })
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint }] })).toBeDefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, viewportWidth: 100 } }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, lines: [{ ...textPaint.lines[0], width: 11 }] } }] })).toBeUndefined()
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { ...textPaint, lines: [{ ...textPaint.lines[0], fragments: [{ text: 'engine line', x: 0, fontMetrics: 1 }] }] } }] })).toBeUndefined()
  })

  // STORY 8.4a. A fragment may carry the ASSET KEY of the face the engine
  // resolved it to, and the key is OPTIONAL: its absence is the projection's
  // own statement that this fragment is a SHIPPED face, so both shapes have to
  // be admitted and the optional one has to be proved optional.
  //
  // WHAT A WRONG ANSWER COSTS HERE, and it is why the shape is checked rather
  // than merely typed. `hasOnly` rejects a key it does not list, isCanvas then
  // fails, parseInbound returns undefined, and engine-client raises
  // PROTOCOL_INVALID — which TERMINATES the worker and rejects every pending
  // request. Not a blank canvas: a dead session.
  it('admits a paint fragment attributed to a carried face, and one attributed to none', () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const response = (fragment: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 8, advance: 12, width: 10, fragments: [fragment] }] } }] } } })
    expect(response({ text: 'engine line', x: 0, assetKey: key })).toBeDefined()
    // THE SHIPPED-FACE PATH: no key at all, which is the common case and the
    // one that must not have become mandatory.
    expect(response({ text: 'engine line', x: 0 })).toBeDefined()
    // An explicit `undefined` is the same statement, and is what a projection
    // reconstructed in JavaScript will hand this guard.
    expect(response({ text: 'engine line', x: 0, assetKey: undefined })).toBeDefined()
    // AND THE KEY IS THE FORMAT'S OWN SHAPE — 64 lowercase hex characters, the
    // same rule the image projection's key is held to. Anything else is a
    // producer that has drifted, not an older one to tolerate: the browser
    // hands this string straight back to the `asset` operation and derives a
    // CSS family from it.
    expect(response({ text: 'engine line', x: 0, assetKey: '' })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, assetKey: key.toUpperCase() })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, assetKey: key.slice(0, 63) })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, assetKey: 'body' })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, assetKey: 7 })).toBeUndefined()
  })

  // STORY 8.4e. A fragment may instead carry the ENGINE'S OWN FontSet NAME for
  // the SHIPPED face it was measured with — the other half of the same
  // attribution, and `assetKey`'s mutually exclusive twin. The pair
  // discriminates exactly as a chain ENTRY's `face`/`assetKey` pair does one
  // level up, with one deliberate difference: NEITHER is legal here, because
  // the absence of both is the wire's statement that the projection did not
  // attribute this fragment, and such a fragment must still paint on the
  // stylesheet's declared stack rather than kill the session.
  //
  // THE BOUND IS CHECKED HERE EVEN THOUGH GO CANNOT BREACH IT TODAY. The
  // engine can only put a key of the FontSet it was given on this field; that
  // is the ENGINE's guarantee, and a guard's job is to hold when the other
  // side is wrong. The value is written into an inline `font-family`
  // declaration on the browser, so an unbounded string here is an unbounded
  // string in a stylesheet.
  it('admits a paint fragment attributed to a shipped face, and refuses one carrying both identities', () => {
    const key = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'
    const response = (fragment: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 8, advance: 12, width: 10, fragments: [fragment] }] } }] } } })
    expect(response({ text: 'engine line', x: 0, face: 'Noto Sans Thai' })).toBeDefined()
    expect(response({ text: 'engine line', x: 0, face: undefined })).toBeDefined()
    // THE SAME BOUND A CHAIN ENTRY'S `face` ALREADY USES — no new numeral, and
    // both of its directions.
    expect(response({ text: 'engine line', x: 0, face: 'f'.repeat(MAX_CANVAS_PROPERTY_STRING) })).toBeDefined()
    expect(response({ text: 'engine line', x: 0, face: 'f'.repeat(MAX_CANVAS_PROPERTY_STRING + 1) })).toBeUndefined()
    // An empty string is not an absence: absence is spelled by omission, and a
    // producer sending '' has drifted rather than said anything.
    expect(response({ text: 'engine line', x: 0, face: '' })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, face: 7 })).toBeUndefined()
    expect(response({ text: 'engine line', x: 0, face: null })).toBeUndefined()
    // EXACTLY ONE OF THE TWO. Both is a producer contradicting itself about
    // which face drew this fragment, and the browser would have to pick.
    expect(response({ text: 'engine line', x: 0, face: 'Noto Sans', assetKey: key })).toBeUndefined()
    // NEITHER IS LEGAL, deliberately — see above. Re-asserted here so the
    // exclusivity is never tightened into a requirement by accident.
    expect(response({ text: 'engine line', x: 0 })).toBeDefined()
  })

  // Story 7.3 / FR47. The alignment vocabulary is TWO closed sets on this
  // boundary as well as in Go: a COMPONENT may be justified, a table
  // COLUMN may not. The validator gates the projection — an unrecognised
  // value drops the whole response — so a justified document that this
  // check refused would blank the entire canvas rather than merely lose
  // its alignment.
  it('admits a justified component, refuses a justified table column, and accepts word-grained fragments', () => {
    const response = (projection: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: projection } })
    const emptyPaint = { overflow: false, truncated: false, lines: [] }
    const component = (align: string) => ({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, align, textPaint: emptyPaint }] })
    for (const align of ['left', 'center', 'right', 'justify']) expect(response(component(align))).toBeDefined()
    for (const align of ['middle', 'JUSTIFY', 'flush', '']) expect(response(component(align))).toBeUndefined()

    // The COLUMN set stays the triple, on its own projection.
    const columns = (align: string) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'table-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, tableColumns: { revision: 7, table: { tableId: 'e7', collection: 'rows[]', alias: 'row', headerHeight: 12000, altRowBackground: '', headerFontFamily: '', headerFontFamilyResolved: 'body', headerFontSize: 0, headerFontSizeResolved: 12000, headerLineSpacing: 0, headerLineSpacingResolved: 1000, headerBackground: '', headerBackgroundResolved: '', headerColor: '', headerColorResolved: '', headerValign: '', headerValignResolved: 'top', headerAlign: '', headerAlignResolved: 'left', columns: [{ id: 'e8', header: 'Amount', width: 72000, align, binding: '{{row.amount}}', rowField: 'amount', rowFieldEditable: true, footer: '', footerOf: '', footerFormat: '' }] } } })
    expect(columns('right')).toBeDefined()
    expect(columns('justify')).toBeUndefined()

    // A justified line arrives as SEVERAL fragments with ascending x — the
    // engine positions each word; the browser never justifies anything.
    const wordGrained = { overflow: false, truncated: false, lines: [{ top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'one', x: 0 }, { text: ' two', x: 4 }, { text: ' three', x: 8 }] }] }
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, align: 'justify', textPaint: wordGrained }] })).toBeDefined()
    // …and a fragment placed outside the component's own box is still
    // refused, word-grained or not.
    expect(response({ ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, align: 'justify', textPaint: { ...wordGrained, lines: [{ ...wordGrained.lines[0], fragments: [...wordGrained.lines[0].fragments, { text: 'far', x: 11 }] }] } }] })).toBeUndefined()
  })

  // Story 7.4 / DW-25. The three places a body-text projection could still be
  // dropped silently: the exact-key `hasOnly` on the paint, the split
  // `optionalString`, and the line bound. Each of these failures blanks the
  // WHOLE canvas — isTextPaint false fails the component, which fails
  // isCanvas, isSnapshot and finally parseInbound — so there is no
  // attributable error, only a designer with no snapshot.
  it('admits a truncated prefix paint, a clause past the identifier bound, and a projected line spacing', () => {
    const response = (component: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [component] } } })
    const line = { top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'engine line', x: 0 }] }
    const text = (extra: object) => ({ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, ...extra })

    // A PREFIX with the flag set is the degraded state, and it is admitted.
    expect(response(text({ textPaint: { overflow: false, truncated: true, lines: [line] } }))).toBeDefined()
    // Both flags are required, exactly as `overflow` always was: a producer
    // that has stopped emitting one has drifted from this contract.
    expect(response(text({ textPaint: { overflow: false, lines: [line] } }))).toBeUndefined()
    expect(response(text({ textPaint: { overflow: false, truncated: 'yes', lines: [line] } }))).toBeUndefined()
    // And an unknown key still drops the response — hasOnly is exact-key.
    expect(response(text({ textPaint: { overflow: false, truncated: false, clipped: true, lines: [line] } }))).toBeUndefined()

    // THE FOURTH MIRROR. An element's value is BODY TEXT and no longer shares
    // the identifier bound; the seven identifier keys still keep it.
    const clause = 'x'.repeat(MAX_CANVAS_PROPERTY_STRING + 1)
    expect(response(text({ value: clause, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeDefined()
    expect(response(text({ fontFamily: clause, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeUndefined()
    expect(response(text({ color: clause, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeUndefined()

    // style.lineSpacing, projected for the first time: thousandths, inside
    // the range the engine's one validator enforces on both entry points.
    expect(response(text({ lineSpacing: 1500, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeDefined()
    expect(response(text({ lineSpacing: 0, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeUndefined()
    expect(response(text({ lineSpacing: 1000001, textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeUndefined()
    expect(response(text({ lineSpacing: '1.5', textPaint: { overflow: false, truncated: false, lines: [] } }))).toBeUndefined()
  })

  it('admits a forty-page paint and refuses one line past the mirrored bound', () => {
    const response = (count: number) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 1800, resizable: true, textPaint: { overflow: false, truncated: false, lines: Array.from({ length: count }, (_value, index) => ({ top: index, baseline: index, advance: 1, width: 10, fragments: [] })) } }] } } })
    expect(response(MAX_CANVAS_BODY_TEXT_LINES)).toBeDefined()
    expect(response(MAX_CANVAS_BODY_TEXT_LINES + 1)).toBeUndefined()
    // The old bound must no longer bite: 257 lines is about six pages, and
    // refusing them here is what blanked the canvas after a paste.
    expect(response(257)).toBeDefined()
  })

  it('admits one bounded text-binding paint label but rejects an editable projection', () => {
    const response = (component: object) => parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true, snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [component] } } })
    const text = { id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, binding: 'customer.name', textPaint: { overflow: false, truncated: false, lines: [] } }
    expect(response(text)).toBeDefined()
    expect(response({ ...text, binding: 'a'.repeat(MAX_ENGINE_BINDING_LENGTH + 1) })).toBeUndefined()
    expect(response({ ...text, binding: '' })).toBeUndefined()
    expect(response({ ...text, binding: { path: 'customer.name' } })).toBeUndefined()
    expect(response({ ...text, type: 'image' })).toBeUndefined()
  })

  // admittedTextLines pulls the first component's paint lines out of a
  // parsed inbound, failing the test if anything on the way is missing.
  // It exists so the tight-leading cases can assert on VALUES: parseInbound
  // is a type guard that returns its input unchanged, so `toBeDefined()`
  // passes for anything non-undefined and would keep passing if the
  // geometry were quietly rewritten on the way through.
  const admittedTextLines = (parsed: ReturnType<typeof parseInbound>) => {
    if (parsed === undefined || parsed.kind !== 'response' || !parsed.ok) throw new Error('the projection was rejected at the boundary')
    const lines = parsed.snapshot.canvas?.components[0]?.textPaint?.lines
    if (lines === undefined) throw new Error('the admitted snapshot carries no text paint lines')
    return lines
  }

  it('rejects non-advancing or out-of-box text paint geometry', () => {
    const response = (textPaint: object) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint }] } },
    })
    const line = { top: 0, baseline: 8, advance: 12, width: 10, fragments: [{ text: 'engine', x: 0 }] }
    expect(response({ overflow: false, truncated: false, lines: [line, { ...line, top: 11, baseline: 19 }] })).toBeUndefined()
    expect(response({ overflow: false, truncated: false, lines: [{ ...line, fragments: [{ text: 'engine', x: 11 }] }] })).toBeUndefined()
    expect(response({ overflow: false, truncated: false, lines: [{ ...line, top: -1, baseline: 8 }] })).toBeUndefined()
    // Story 7.2 / D-7.2.2, INVERTED DELIBERATELY. This assertion used to
    // read `.toBeUndefined()`: a baseline of 13 against top 0 and advance
    // 12 sits below the next line's top, so the line boxes overlap. That
    // IS tight leading, it is what the PDF draws, and the clause refusing
    // it restated an engine invariant on the browser's side of the
    // channel — blanking the entire projection over one line. It is
    // pinned here as deliberately gone rather than left silently untested.
    //
    // Asserted on the PARSED VALUES, not with `toBeDefined()`: the
    // predicate is a type guard that returns the input unchanged, so
    // `toBeDefined()` would pass for anything at all that is not
    // `undefined` and would keep passing if the geometry were quietly
    // rewritten on the way through.
    const tightLine = { ...line, baseline: 13 }
    const accepted = response({ overflow: false, truncated: false, lines: [tightLine] })
    expect(admittedTextLines(accepted)).toEqual([tightLine])
  })

  it('accepts a whole snapshot whose text lines are set at tight leading', () => {
    // The projection-level half of the assertion above: a MULTI-LINE
    // component at an advance tighter than its own first-baseline
    // offset. Every line here has baseline > top + advance, and
    // consecutive tops still step by exactly one advance — the shape
    // page_setup.go emits for `style.lineSpacing: 0.6`. Before D-7.2.2
    // one such line failed isTextPaint, then isCanvas, then isSnapshot,
    // and parseInbound dropped the whole response.
    const tight = {
      overflow: false, truncated: false,
      lines: [
        { top: 0, baseline: 11, advance: 9, width: 10, fragments: [{ text: 'first', x: 0 }] },
        { top: 9, baseline: 20, advance: 9, width: 10, fragments: [{ text: 'second', x: 0 }] },
        { top: 18, baseline: 29, advance: 9, width: 10, fragments: [{ text: 'third', x: 0 }] },
      ],
    }
    const parsed = parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 40, resizable: true, textPaint: tight }] } },
    })
    // Every line's own numbers must survive the boundary intact — the
    // canvas paints from these, and the engine is the only thing allowed
    // to have decided them (AD-17).
    const lines = admittedTextLines(parsed)
    expect(lines).toEqual(tight.lines)
    expect(lines.map((l) => [l.top, l.baseline, l.advance])).toEqual([[0, 11, 9], [9, 20, 9], [18, 29, 9]])
    // And the overlap really is present in what was admitted, so this
    // case cannot quietly stop exercising tight leading.
    expect(lines.every((l) => l.baseline > l.top + l.advance)).toBe(true)
    // A non-advancing projection is still refused, so the acceptance
    // above is not "the predicate stopped checking anything".
    expect(parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 40, resizable: true, textPaint: { ...tight, lines: tight.lines.map((l) => ({ ...l, advance: 0 })) } }] } },
    })).toBeUndefined()
  })

  it('accepts a well-formed image paint inside its own box and rejects malformed or out-of-box substitutes', () => {
    const box = { x: 0, y: 0, width: 100, height: 50 }
    // Finding 12 (review of 2026-08-29): the wire key is the FULL 64-hex
    // digest (D-5.13.2 amendment) — a truncated key passed admission here
    // before the fix, then could never resolve through the per-key 'asset'
    // fetch. This fixture's OWN well-formedness now pins that width.
    const image = { mediaType: 'image/png', assetKey: 'ab'.repeat(32), width: 300, height: 150, drawX: 0, drawY: 0, drawWidth: 100, drawHeight: 50 }
    const response = (img: object | undefined) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', resizable: true, ...box, ...(img === undefined ? {} : { image: img }) }] } },
    })
    expect(response(image)).toBeDefined()
    expect(response(undefined)).toBeDefined()
    expect(response({ ...image, mediaType: '' })).toBeUndefined()
    expect(response({ ...image, assetKey: '' })).toBeUndefined()
    expect(response({ ...image, assetKey: 'z'.repeat(64) })).toBeUndefined() // 64 chars, not hex
    // A well-formed but TRUNCATED key (12 hex characters, the inspector's
    // own display-abbreviation width) must be rejected, not merely a
    // wrong-length string of the wrong shape.
    expect(response({ ...image, assetKey: 'abcdef012345' })).toBeUndefined()
    expect(response({ ...image, width: 0 })).toBeUndefined()
    expect(response({ ...image, height: -1 })).toBeUndefined()
    expect(response({ ...image, drawWidth: 0 })).toBeUndefined()
    expect(response({ ...image, drawX: -1 })).toBeUndefined()
    expect(response({ ...image, drawX: 1, drawWidth: 100 })).toBeUndefined() // spills past box.x+box.width
    expect(response({ ...image, drawY: 1, drawHeight: 50 })).toBeUndefined() // spills past box.y+box.height
    expect(response({ ...image, extra: true })).toBeUndefined()
    // A non-image component must never carry an image paint.
    expect(parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, truncated: false, lines: [] }, image }] } },
    })).toBeUndefined()
  })

  it("admits Finding 9's imageUnavailable discriminant only for an image component with no image paint, and only its two named values", () => {
    const box = { x: 0, y: 0, width: 100, height: 50 }
    const response = (component: object) => parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'image', band: 'content', resizable: true, ...box, ...component }] } },
    })
    expect(response({ imageUnavailable: 'missing' })).toBeDefined()
    expect(response({ imageUnavailable: 'undecodable' })).toBeDefined()
    expect(response({})).toBeDefined() // absent is legal (an image paint present, or nothing yet)
    expect(response({ imageUnavailable: 'something-else' })).toBeUndefined()
    // Not a text component.
    expect(parseInbound({
      protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'canvas-1', ok: true,
      snapshot: { documentState: 'loaded', revision: 1, byteLength: 1, canvas: { ...canvas, components: [{ id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 10, height: 10, resizable: true, textPaint: { overflow: false, truncated: false, lines: [] }, imageUnavailable: 'missing' }] } },
    })).toBeUndefined()
    // Not alongside a PRESENT image paint — the two are one signal.
    const image = { mediaType: 'image/png', assetKey: 'ab'.repeat(32), width: 300, height: 150, drawX: 0, drawY: 0, drawWidth: 100, drawHeight: 50 }
    expect(response({ image, imageUnavailable: 'missing' })).toBeUndefined()
  })

  it('admits only operation-coherent closed worker requests and responses', () => {
    const load = new Uint8Array([1]).buffer
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'load-1', operation: 'load', payload: load })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'load-2', operation: 'load' })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'snapshot-1', operation: 'snapshot', payload: load })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'snapshot-2', operation: 'snapshot', viewport: 900 })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'undo-1', operation: 'undo' })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'redo-1', operation: 'redo' })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'undo-2', operation: 'undo', payload: load })).toBeUndefined()
    const dpr = ['device', 'PixelRatio'].join('')
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'command-1', operation: 'command', payload: load, [dpr]: 2 })).toBeUndefined()
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'response-1', ok: false, error: { code: 'NO', message: 'no' }, font: 'browser' })).toBeUndefined()
    expect(parseInbound({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'lifecycle', state: 'ready', snapshot: {} })).toBeUndefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'asset-1', operation: 'asset', payload: load })).toBeDefined()
    expect(parseRequest({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'asset-2', operation: 'asset' })).toBeUndefined()
    const assetResponse = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response' as const, requestId: 'asset-1', ok: true as const, snapshot: { documentState: 'loaded' as const, revision: 3, byteLength: 1 }, bytes: load }
    expect(parseInbound(assetResponse)).toBeDefined()
  })

  it('accepts only a correlated, bounded three-byte render envelope and producer digest', () => {
    const part = new Uint8Array([1]).buffer
    const render = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'render-1', operation: 'render', payload: { template: part, data: part.slice(0), params: part.slice(0) } }
    expect(parseRequest(render)).toBeDefined()
    expect(parseRequest({ ...render, payload: { template: part, data: part.slice(0) } })).toBeUndefined()
    expect(parseRequest({ ...render, viewport: 900 })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'render-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, bytes: part, preview: { revision: 7, identity: 'b'.repeat(64), pdfSha256: 'a'.repeat(64), diagnostics: [] } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, pdfSha256: 'not-a-digest' } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, identity: 'not-an-identity' } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, revision: -1 } })).toBeUndefined()
    expect(parseInbound({ ...response, bytes: undefined })).toBeUndefined()
    expect(parseRequest({ ...render, payload: { template: new ArrayBuffer(MAX_ENGINE_PAYLOAD_BYTES + 1), data: part.slice(0), params: part.slice(0) } })).toBeUndefined()
    expect(parseInbound({ ...response, bytes: new ArrayBuffer(MAX_ENGINE_RENDER_PDF_BYTES + 1) })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, diagnostics: [{}] } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, diagnostics: [], snapshot: {} } })).toBeUndefined()
  })

  it('accepts only an identity-only engine response with revision-bound opaque evidence', () => {
    const part = new Uint8Array([1]).buffer
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'identity-1', operation: 'identity', payload: { data: part, params: part.slice(0) } }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: { data: part } })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'identity-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, preview: { revision: 7, identity: 'a'.repeat(64) } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, pdfSha256: 'b'.repeat(64) } })).toBeUndefined()
    expect(parseInbound({ ...response, preview: { ...response.preview, revision: 6 } })).toBeUndefined()
  })

  it('admits only a bounded, revision-correlated engine parameter-reference projection', () => {
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'params-1', operation: 'parameter-references' }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: new Uint8Array([1]).buffer })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'params-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, parameterReferences: { revision: 7, names: ['branch', 'reportDate'] } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 6, names: ['branch'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate', 'branch'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate', 'reportDate'] } })).toBeUndefined()
    expect(parseInbound({ ...response, parameterReferences: { revision: 7, names: ['reportDate'], expression: 'params.reportDate' } })).toBeUndefined()
  })

  it('admits only a selected, revision-correlated table-column paint projection', () => {
    const payload = new TextEncoder().encode('{"id":"e7"}').buffer
    const request = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'request', requestId: 'table-1', operation: 'table-columns', payload }
    expect(parseRequest(request)).toBeDefined()
    expect(parseRequest({ ...request, payload: undefined })).toBeUndefined()
    const response = { protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'table-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, tableColumns: { revision: 7, table: { tableId: 'e7', collection: 'transactions[]', alias: 'transaction', headerHeight: 12000, altRowBackground: '', headerFontFamily: '', headerFontFamilyResolved: 'body', headerFontSize: 0, headerFontSizeResolved: 12000, headerLineSpacing: 0, headerLineSpacingResolved: 1000, headerBackground: '', headerBackgroundResolved: '', headerColor: '', headerColorResolved: '', headerValign: '', headerValignResolved: 'top', headerAlign: '', headerAlignResolved: 'left', columns: [{ id: 'e8', header: 'Amount', width: 72000, align: 'right', binding: '{{transaction.amount}}', rowField: 'amount', rowFieldEditable: true, footer: 'sum', footerOf: 'transactions.amount', footerFormat: '#,##0.00' }] } } }
    expect(parseInbound(response)).toBeDefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, revision: 6 } })).toBeUndefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, table: { ...response.tableColumns.table, columns: [{ ...response.tableColumns.table.columns[0], bind: 'row.amount' }] } } })).toBeUndefined()
    expect(parseInbound({ ...response, tableColumns: { ...response.tableColumns, table: { ...response.tableColumns.table, columns: [{ ...response.tableColumns.table.columns[0], width: 0 }] } } })).toBeUndefined()
  })

  // STORY 12.3 — THE TABLE OBJECT'S KEY SET, RED-PROVED IN BOTH DIRECTIONS.
  //
  // `hasExactKeys` is a LENGTH check AND a membership check, so it rejects a
  // key Go stops sending exactly as hard as one it starts sending. Only the
  // second direction was possible under `hasOnly`, which is what isCanvas uses,
  // so both arms are asserted here rather than assumed from the canvas guard.
  //
  // WHAT A FAILURE COSTS, because it is not the blank canvas the wire test's
  // header describes: parseInbound returns undefined, engine-client raises
  // PROTOCOL_INVALID, the worker is TERMINATED and every pending request
  // rejected, and no re-spawn exists. On a FIRST table-editor open it is
  // silent — openTableEditor's catch sets an error that renders only inside
  // <TableEditor>, which never mounts. So the assertion is on parseInbound's
  // RETURN VALUE and never on a visual symptom.
  it('refuses a table projection with a missing member and one with a surplus key alike', () => {
    const table = { tableId: 'e7', collection: 'transactions[]', alias: 'transaction', headerHeight: 12000, altRowBackground: '', headerFontFamily: '', headerFontFamilyResolved: 'body', headerFontSize: 0, headerFontSizeResolved: 12000, headerLineSpacing: 0, headerLineSpacingResolved: 1000, headerBackground: '', headerBackgroundResolved: '', headerColor: '', headerColorResolved: '', headerValign: '', headerValignResolved: 'top', headerAlign: '', headerAlignResolved: 'left', columns: [] }
    const responseFor = (value: Record<string, unknown>) => ({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'table-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, tableColumns: { revision: 7, table: value } })
    expect(parseInbound(responseFor(table))).toBeDefined()
    // DIRECTION ONE — a projected member the guard's list does not name. Modelled
    // as the guard being BEHIND Go: the response carries a seventeenth member.
    expect(parseInbound(responseFor({ ...table, headerBorder: '' }))).toBeUndefined()
    // DIRECTION TWO — a guard key with no projected member, one per member, so
    // the proof is not carried by whichever key happens to be checked first.
    for (const key of Object.keys(table)) {
      const missing: Record<string, unknown> = { ...table }
      delete missing[key]
      expect(parseInbound(responseFor(missing)), `dropping ${key} must be refused`).toBeUndefined()
    }
    // And the typed clauses behind the key list, so a member of the right NAME
    // and the wrong shape is refused too.
    expect(parseInbound(responseFor({ ...table, headerHeight: '12000' }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerFontSize: 12.5 }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerAlignResolved: '' }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerValignResolved: 'centre' }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerAlign: 'justify' }))).toBeUndefined()
    // A COMMITTED alignment of '' is ABSENT and must stay admissible: refusing
    // it would make an unstyled table's own projection unparseable.
    expect(parseInbound(responseFor({ ...table, headerAlign: '', headerValign: '' }))).toBeDefined()
  })

  // A NEGATIVE LENGTH THE FILE DOOR ADMITS MUST NOT KILL THE WORKER.
  //
  // The guard used to require `>= 0` for every projected length, but the loader
  // bounds NEITHER headerHeight NOR style.fontSize: decimal.go negates on
  // `sign < 0`, parse_bands.go assigns `t.HeaderHeight = hh` unchecked, and the
  // style decoder assigns `st.FontSize = present(v)` the same way. So a
  // hand-authored `"headerHeight": -5` loads and renders TODAY, and after this
  // story opening its table editor failed the guard — parseInbound undefined,
  // PROTOCOL_INVALID, worker.terminate(), no re-spawn, and on a first open
  // nothing shown at all because <TableEditor> never mounts to carry the error.
  //
  // The remedy is here rather than at the loader on purpose: bounding the
  // loader would narrow the format, which the story forbids itself. The guard's
  // job is to admit exactly what the file door admits.
  it('admits the negative lengths the loader itself admits, and still refuses a negative line spacing', () => {
    const table = { tableId: 'e7', collection: 'transactions[]', alias: 'transaction', headerHeight: 12000, altRowBackground: '', headerFontFamily: '', headerFontFamilyResolved: 'body', headerFontSize: 0, headerFontSizeResolved: 12000, headerLineSpacing: 0, headerLineSpacingResolved: 1000, headerBackground: '', headerBackgroundResolved: '', headerColor: '', headerColorResolved: '', headerValign: '', headerValignResolved: 'top', headerAlign: '', headerAlignResolved: 'left', columns: [] }
    const responseFor = (value: Record<string, unknown>) => ({ protocolVersion: ENGINE_PROTOCOL_VERSION, kind: 'response', requestId: 'table-1', ok: true, snapshot: { documentState: 'loaded', revision: 7, byteLength: 1 }, tableColumns: { revision: 7, table: value } })
    // The document that loads today: a negative headerHeight, and the negative
    // fontSize that cascades into its resolved twin.
    expect(parseInbound(responseFor({ ...table, headerHeight: -5000 }))).toBeDefined()
    expect(parseInbound(responseFor({ ...table, headerFontSize: -4000, headerFontSizeResolved: -4000 }))).toBeDefined()
    // THE LINE-SPACING PAIR IS NOT RELAXED, because for it the file door really
    // does bound: DecodeLineSpacingRaw refuses anything outside [1, 1000000]
    // thousandths, so a negative one cannot come from a loaded document and
    // admitting it would only widen the guard past its source.
    expect(parseInbound(responseFor({ ...table, headerLineSpacing: -1 }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerLineSpacingResolved: -1000 }))).toBeUndefined()
    // And the shape clauses still hold on the relaxed members: a length is
    // still an INTEGER count of millipoints and still a number.
    expect(parseInbound(responseFor({ ...table, headerHeight: -5.5 }))).toBeUndefined()
    expect(parseInbound(responseFor({ ...table, headerFontSizeResolved: null }))).toBeUndefined()
  })
})
