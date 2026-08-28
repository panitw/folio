import { describe, expect, it } from 'vitest'
import { normalizePublicPath, pageIdentity, releaseIdentity } from './offline-release-contract.mjs'

describe('offline release contract', () => {
  it('normalizes the single Windows separator emitted by path.relative', () => {
    expect(normalizePublicPath('assets\\worker-abcdef12.js')).toBe('/assets/worker-abcdef12.js')
  })

  it('makes worker-only changes produce a distinct page and cache identity', () => {
    const assets = [{ url: '/index.html', sha256: 'a'.repeat(64) }, { url: '/assets/app-abcdef12.js', sha256: 'b'.repeat(64) }]
    expect(releaseIdentity(assets, 'c'.repeat(64))).not.toBe(releaseIdentity(assets, 'd'.repeat(64)))
    expect(pageIdentity(assets, 'c'.repeat(64))).not.toBe(pageIdentity(assets, 'd'.repeat(64)))
  })
})
