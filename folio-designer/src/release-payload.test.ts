import { describe, expect, it } from 'vitest'
import { parseS1Payload } from './release-payload'

const hash = 'a'.repeat(64)
const cacheAssets = ['/index.html', '/engine', '/latin', '/thai', '/cjk', ...Array.from({ length: 15 }, (_, index) => `/asset-${index}`)]
const payload = () => ({ version: 1, releaseId: hash, pageId: 'b'.repeat(64), unit: 'MiB', decimals: 2, cachedBytes: 200, assetCount: 20, cacheAssets: cacheAssets.map((assetUrl) => ({ assetUrl, bytes: 10 })), rows: [
  { id: 'engine', label: 'Engine', delivery: 'cached-asset', assetUrl: '/engine', bytes: 10, sha256: hash }, { id: 'latin-font', label: 'Latin font', delivery: 'cached-asset', assetUrl: '/latin', bytes: 10, sha256: hash }, { id: 'thai-font', label: 'Thai font', delivery: 'cached-asset', assetUrl: '/thai', bytes: 10, sha256: hash }, { id: 'cjk-font', label: 'CJK font', delivery: 'cached-asset', assetUrl: '/cjk', bytes: 10, sha256: hash }, { id: 'thai-dictionary', label: 'Thai dictionary', delivery: 'embedded-in-engine', assetUrl: '/engine', bytes: 5, sha256: hash },
  ] })
describe('release-derived S1 payload boundary', () => {
  it('accepts the current bounded generated shape', () => expect(parseS1Payload(payload())).toMatchObject({ cachedBytes: 200, assetCount: 20, rows: expect.any(Array) }))
  it('rejects stale arithmetic, dictionary delivery fiction, and surplus fields', () => {
    const total = payload(); total.cachedBytes = 41; expect(parseS1Payload(total)).toBeUndefined()
    const delivery = payload(); delivery.rows[4].delivery = 'cached-asset'; expect(parseS1Payload(delivery)).toBeUndefined()
    const surplus = { ...payload(), document: 'must-not-cross-boundary' }; expect(parseS1Payload(surplus)).toBeUndefined()
  })
  it('rejects cache asset counts outside the bounded release envelope', () => {
    const tooFew = payload(); tooFew.assetCount = 9; tooFew.cacheAssets = tooFew.cacheAssets.slice(0, 9); tooFew.cachedBytes = 90
    expect(parseS1Payload(tooFew)).toBeUndefined()
    const tooMany = payload(); tooMany.assetCount = 65; tooMany.cacheAssets = Array.from({ length: 65 }, (_, index) => ({ assetUrl: `/asset-${index}`, bytes: 10 })); tooMany.cachedBytes = 650
    expect(parseS1Payload(tooMany)).toBeUndefined()
  })
})
