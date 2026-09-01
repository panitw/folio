import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { declaredCacheAssetBounds, normalizePublicPath, pageIdentity, releaseIdentity } from './offline-release-contract.mjs'

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

// An INDEPENDENT re-read of the real file: line splitting and prefix matching
// rather than the helper's own regex, so this asserts the declared numbers
// without re-typing them. A literal here would be a second authority — the
// exact drift the helper exists to make impossible.
const reReadDeclared = (name) => {
  const line = readFileSync(join(import.meta.dirname, '..', 'src', 'release-payload.ts'), 'utf8').split('\n').filter((candidate) => candidate.startsWith(`const ${name} = `))
  expect(line).toHaveLength(1)
  return Number(line[0].slice(`const ${name} = `.length))
}

describe('declared cache-asset bounds', () => {
  it('reads the two bounds src/release-payload.ts actually declares', () => {
    expect(declaredCacheAssetBounds()).toEqual({ minimumCacheAssets: reReadDeclared('minimumCacheAssets'), maximumCacheAssets: reReadDeclared('maximumCacheAssets') })
  })

  // One mutation at a time, and each held to ITS OWN message: a reader that
  // failed for a shared reason would prove only that it failed, not that it can
  // tell a rename from a duplicate.
  it('throws when the constant is absent, naming it and reporting none found', () => {
    expect(() => declaredCacheAssetBounds('const minimumCacheAssets = 10\nconst cacheCeiling = 64\n')).toThrow(/`maximumCacheAssets` as a single live constant: found 0 .*reads as none/)
  })

  it('throws when the only occurrence is commented out', () => {
    expect(() => declaredCacheAssetBounds('const minimumCacheAssets = 10\n// const maximumCacheAssets = 64\n')).toThrow(/`maximumCacheAssets` as a single live constant: found 0 /)
    expect(() => declaredCacheAssetBounds('const minimumCacheAssets = 10\n  const maximumCacheAssets = 64\n')).toThrow(/`maximumCacheAssets` as a single live constant: found 0 /)
  })

  it('throws when the constant is declared twice, reporting both', () => {
    expect(() => declaredCacheAssetBounds('const minimumCacheAssets = 10\nconst maximumCacheAssets = 64\nconst maximumCacheAssets = 4096\n')).toThrow(/`maximumCacheAssets` as a single live constant: found 2 .*authority ambiguous/)
  })

  it('reports the minimum by its own name rather than by the maximum', () => {
    expect(() => declaredCacheAssetBounds('const maximumCacheAssets = 64\n')).toThrow(/`minimumCacheAssets` as a single live constant: found 0 /)
  })
})
