import { describe, expect, it } from 'vitest'
import vm from 'node:vm'
import { isCacheableStaticRequest, isStatusRequest, serviceWorkerSource } from './offline-service-worker-template.mjs'

const release = (id = 'a'.repeat(64), workerRevision = 'b'.repeat(64)) => ({ version: 2, id, pageId: 'c'.repeat(64), workerRevision, assets: [{ url: '/index.html', sha256: 'd'.repeat(64), immutable: false }, { url: '/assets/app-abc12345.js', sha256: 'e'.repeat(64), immutable: true }] })
const request = (overrides = {}) => ({ url: 'https://folio.test/assets/app-abc12345.js', method: 'GET', credentials: 'omit', mode: 'cors', ...overrides })

function workerHarness(workerRelease, options = {}) {
  const handlers = {}
  const cacheData = options.cacheData ?? new Map()
  const deleted = []
  const caches = {
    open: async (name) => {
      if (!cacheData.has(name)) cacheData.set(name, new Map())
      const store = cacheData.get(name)
      return { match: async (key) => store.get(key), put: async (key, value) => store.set(key, value) }
    },
    delete: async (name) => { deleted.push(name); return cacheData.delete(name) },
    keys: async () => [...cacheData.keys()],
  }
  const clients = { claim: async () => {}, matchAll: async () => options.windows ?? [] }
  const self = { location: { origin: 'https://folio.test' }, addEventListener: (name, handler) => { handlers[name] = handler } }
  vm.runInNewContext(serviceWorkerSource(workerRelease), { self, caches, clients, fetch: options.fetch ?? (async () => { throw new Error('network unavailable') }), crypto: globalThis.crypto, Response, URL, Set, Promise, Uint8Array })
  return { handlers, cacheData, deleted }
}

describe('service worker static policy', () => {
  it('allows only credential-omitting manifest static requests', () => {
    const paths = new Set(['/assets/app-abc12345.js'])
    expect(isCacheableStaticRequest(request(), 'https://folio.test', paths)).toBe(true)
    expect(isCacheableStaticRequest(request({ credentials: 'same-origin' }), 'https://folio.test', paths)).toBe(false)
    expect(isCacheableStaticRequest(request({ credentials: 'include' }), 'https://folio.test', paths)).toBe(false)
    expect(isCacheableStaticRequest(request({ url: 'https://evil.test/assets/app-abc12345.js' }), 'https://folio.test', paths)).toBe(false)
    expect(isCacheableStaticRequest(request({ url: 'https://folio.test/documents/customer.folio' }), 'https://folio.test', paths)).toBe(false)
    expect(isCacheableStaticRequest(request({ method: 'POST' }), 'https://folio.test', paths)).toBe(false)
  })

  it('executes the emitted status handler with no module closure', async () => {
    const workerRelease = release()
    const cacheName = `folio-release-${workerRelease.id}`
    const marker = `/__folio-release__/${workerRelease.id}`
    const cacheData = new Map([[cacheName, new Map([[marker, { text: async () => workerRelease.id }], ...workerRelease.assets.map((asset) => [asset.url, {}])])]])
    const harness = workerHarness(workerRelease, { cacheData })
    const messages = []
    let completion
    harness.handlers.message({ data: { version: 1, type: 'get-offline-status' }, source: { postMessage: (message) => messages.push(message) }, waitUntil: (promise) => { completion = promise } })
    await completion
    expect(messages).toEqual([{ version: 1, type: 'offline-status', state: 'ready', releaseId: workerRelease.id, pageId: workerRelease.pageId }])
    expect(isStatusRequest({ version: 1, type: 'get-offline-status' })).toBe(true)
  })

  it('keeps an old complete cache when a same-assets changed-worker install fails', async () => {
    const old = release('a'.repeat(64), 'b'.repeat(64))
    const next = release('f'.repeat(64), 'g'.repeat(64))
    const oldCache = `folio-release-${old.id}`
    const cacheData = new Map([[oldCache, new Map([['/old', {}]])]])
    const harness = workerHarness(next, { cacheData })
    let completion
    harness.handlers.install({ waitUntil: (promise) => { completion = promise } })
    await expect(completion).rejects.toThrow('network unavailable')
    expect(cacheData.has(oldCache)).toBe(true)
    expect(harness.deleted).toEqual([`folio-release-${next.id}`])
  })

  it('does not retire an old cache while a window can still rely on it, then retires it with no clients', async () => {
    const current = release()
    const currentCache = `folio-release-${current.id}`
    const marker = `/__folio-release__/${current.id}`
    const complete = new Map([[marker, { text: async () => current.id }], ...current.assets.map((asset) => [asset.url, {}])])
    const oldCache = 'folio-release-old-complete'
    const cacheData = new Map([[currentCache, complete], [oldCache, new Map()]])
    const retained = workerHarness(current, { cacheData, windows: [{ postMessage: () => {} }] })
    let completion
    retained.handlers.activate({ waitUntil: (promise) => { completion = promise } })
    await completion
    expect(cacheData.has(oldCache)).toBe(true)
    const retired = workerHarness(current, { cacheData, windows: [] })
    retired.handlers.activate({ waitUntil: (promise) => { completion = promise } })
    await completion
    expect(cacheData.has(oldCache)).toBe(false)
  })
})
