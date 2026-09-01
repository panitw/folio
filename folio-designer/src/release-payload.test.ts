import { afterEach, describe, expect, it, vi } from 'vitest'
import { loadS1Payload, parseS1Payload, type S1PayloadRejection } from './release-payload'

const hash = 'a'.repeat(64)
const cacheAssets = ['/index.html', '/engine', '/latin', '/thai', '/cjk', ...Array.from({ length: 15 }, (_, index) => `/asset-${index}`)]
const payload = () => ({ version: 1, releaseId: hash, pageId: 'b'.repeat(64), unit: 'MiB', decimals: 2, cachedBytes: 200, assetCount: 20, cacheAssets: cacheAssets.map((assetUrl) => ({ assetUrl, bytes: 10 })), rows: [
  { id: 'engine', label: 'Engine', delivery: 'cached-asset', assetUrl: '/engine', bytes: 10, sha256: hash }, { id: 'latin-font', label: 'Latin font', delivery: 'cached-asset', assetUrl: '/latin', bytes: 10, sha256: hash }, { id: 'thai-font', label: 'Thai font', delivery: 'cached-asset', assetUrl: '/thai', bytes: 10, sha256: hash }, { id: 'cjk-font', label: 'CJK font', delivery: 'cached-asset', assetUrl: '/cjk', bytes: 10, sha256: hash }, { id: 'thai-dictionary', label: 'Thai dictionary', delivery: 'embedded-in-engine', assetUrl: '/engine', bytes: 5, sha256: hash },
  ] })

// A rejection is only evidence if it NAMES its cause, so every assertion below
// reads the reason rather than the mere absence of a payload. `accepted` is
// returned for a success so a case that stops being rejected reddens loudly
// instead of quietly satisfying a `not.toBe` somewhere.
const reasonOf = (value: unknown): S1PayloadRejection | 'accepted' => { const result = parseS1Payload(value); return result.ok ? 'accepted' : result.reason }

const overBound = () => { const over = payload(); over.assetCount = 65; over.cacheAssets = Array.from({ length: 65 }, (_, index) => ({ assetUrl: `/asset-${index}`, bytes: 10 })); over.cachedBytes = 650; return over }
const underBound = () => { const under = payload(); under.assetCount = 9; under.cacheAssets = under.cacheAssets.slice(0, 9); under.cachedBytes = 90; return under }
const staleArithmetic = () => { const total = payload(); total.cachedBytes = 41; return total }
const deliveryFiction = () => { const delivery = payload(); delivery.rows[4].delivery = 'cached-asset'; return delivery }
const surplusField = () => ({ ...payload(), document: 'must-not-cross-boundary' })
const rowNotAnObject = () => { const rows = payload(); (rows.rows as unknown[])[2] = null; return rows }
const rowMislabelled = () => { const rows = payload(); rows.rows[1].label = 'Engine'; return rows }
const duplicateCacheAsset = () => { const duplicate = payload(); duplicate.cacheAssets[3] = { assetUrl: '/engine', bytes: 10 }; return duplicate }

// Every rejecting input this file asserts, in one table, so the exclusivity claim
// below is over the whole set rather than over whichever cases someone remembered.
const rejections: readonly (readonly [string, () => unknown, S1PayloadRejection])[] = [
  ['undefined', () => undefined, 'not-an-object'],
  ['null', () => null, 'not-an-object'],
  ['a bare string', () => 'not a payload', 'not-an-object'],
  ['a surplus field', surplusField, 'payload-shape'],
  ['a payload over the release bound', overBound, 'asset-count-over-maximum'],
  ['a payload under the release floor', underBound, 'asset-count-under-minimum'],
  ['a duplicated cache asset', duplicateCacheAsset, 'cache-assets-invalid'],
  ['stale cached-byte arithmetic', staleArithmetic, 'cached-bytes-mismatch'],
  ['a row that is not an object', rowNotAnObject, 'row-not-an-object'],
  ['a mislabelled row', rowMislabelled, 'row-shape'],
  ['dictionary delivery fiction', deliveryFiction, 'row-delivery-composition'],
]

const bootstrapNode = (text: string) => {
  const node = document.createElement('script')
  node.id = 'folio-release-bootstrap'
  node.type = 'application/json'
  node.textContent = text
  document.body.append(node)
  return node
}

afterEach(() => { document.getElementById('folio-release-bootstrap')?.remove(); vi.restoreAllMocks() })

describe('release-derived S1 payload boundary', () => {
  it('accepts the current bounded generated shape', () => expect(parseS1Payload(payload())).toMatchObject({ ok: true, payload: { cachedBytes: 200, assetCount: 20, rows: expect.any(Array) } }))

  it('names the cause of every rejection rather than answering the same value to all of them', () => {
    expect(rejections.map(([name, input]) => [name, reasonOf(input())])).toEqual(rejections.map(([name, , reason]) => [name, reason]))
  })

  it('rejects cache asset counts outside the bounded release envelope, one reason per end', () => {
    expect(reasonOf(underBound())).toBe('asset-count-under-minimum')
    expect(reasonOf(overBound())).toBe('asset-count-over-maximum')
  })

  // THE DISCRIMINATION CLAIM. A named reason that several unrelated causes can
  // also produce is no better than the bare `undefined` it replaced.
  it('produces the over-the-bound reason from that cause and from no other asserted input', () => {
    expect(rejections.filter(([, input]) => reasonOf(input()) === 'asset-count-over-maximum').map(([name]) => name)).toEqual(['a payload over the release bound'])
    expect(rejections.filter(([, input]) => reasonOf(input()) === 'asset-count-under-minimum').map(([name]) => name)).toEqual(['a payload under the release floor'])
  })
})

describe('S1 bootstrap reader', () => {
  it('reports an absent bootstrap node by its own reason', () => expect(loadS1Payload()).toEqual({ ok: false, reason: 'no-bootstrap' }))

  it('reports an empty bootstrap node by that same reason', () => {
    bootstrapNode('')
    expect(loadS1Payload()).toEqual({ ok: false, reason: 'no-bootstrap' })
  })

  it('reads the emitted bootstrap shape', () => {
    bootstrapNode(JSON.stringify({ s1: payload() }))
    expect(loadS1Payload()).toMatchObject({ ok: true, payload: { assetCount: 20 } })
  })

  // The two causes that used to be one bare `undefined`.
  it('tells a bootstrap that is not JSON apart from one that is over the bound', () => {
    bootstrapNode('{"s1": {"version": 1, "cache')
    const malformed = loadS1Payload()
    document.getElementById('folio-release-bootstrap')!.remove()
    bootstrapNode(JSON.stringify({ s1: overBound() }))
    const over = loadS1Payload()
    expect(malformed).toEqual({ ok: false, reason: 'malformed-json' })
    expect(over).toEqual({ ok: false, reason: 'asset-count-over-maximum' })
    expect(malformed).not.toEqual(over)
  })

  // The catch exists for JSON.parse's SyntaxError and for nothing else. Anything
  // else thrown here is a fault in the page, and disguising it as a malformed
  // bootstrap is how a defect becomes invisible.
  it('rethrows a throw that is not a JSON.parse SyntaxError instead of converting it to a rejection', () => {
    bootstrapNode('{"s1": null}')
    vi.spyOn(JSON, 'parse').mockImplementation(() => { throw new RangeError('bootstrap read faulted') })
    expect(() => loadS1Payload()).toThrow(RangeError)
  })

  // `JSON.parse('null')` succeeds, so reading `.s1` off it would throw a TypeError
  // the narrowed catch no longer admits. It must still reach the parser as a
  // rejection: narrowing may not change what a user sees on this path.
  it('keeps a null bootstrap a rejection rather than a propagating TypeError', () => {
    bootstrapNode('null')
    expect(loadS1Payload()).toEqual({ ok: false, reason: 'not-an-object' })
  })
})
