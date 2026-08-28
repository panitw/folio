import { describe, expect, it, vi } from 'vitest'
import { parseWorkerProgress, parseWorkerStatus, reduceOfflineLifecycle, registerOfflineLifecycle } from './offline-lifecycle'
import type { S1Payload } from './release-payload'

const release = 'a'.repeat(64)
const page = 'b'.repeat(64)
const payload: S1Payload = { version: 1, releaseId: release, pageId: page, unit: 'MiB', decimals: 2, cachedBytes: 100, assetCount: 10, cacheAssets: ['/index.html', '/assets/engine.wasm', '/assets/latin.ttf', '/assets/thai.ttf', '/assets/cjk.ttf', '/assets/a.js', '/assets/b.js', '/assets/c.js', '/assets/d.js', '/assets/e.js'].map((assetUrl) => ({ assetUrl, bytes: 10 })), rows: [
  { id: 'engine', label: 'Engine', delivery: 'cached-asset', assetUrl: '/assets/engine.wasm', bytes: 10, sha256: release },
  { id: 'latin-font', label: 'Latin font', delivery: 'cached-asset', assetUrl: '/assets/latin.ttf', bytes: 10, sha256: release },
  { id: 'thai-font', label: 'Thai font', delivery: 'cached-asset', assetUrl: '/assets/thai.ttf', bytes: 10, sha256: release },
  { id: 'cjk-font', label: 'CJK font', delivery: 'cached-asset', assetUrl: '/assets/cjk.ttf', bytes: 10, sha256: release },
  { id: 'thai-dictionary', label: 'Thai dictionary', delivery: 'embedded-in-engine', assetUrl: '/assets/engine.wasm', bytes: 5, sha256: release },
] }

describe('offline lifecycle message boundary', () => {
  it('accepts only a bounded status for the page release that requested it', () => {
    expect(parseWorkerStatus({ version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: page })).toEqual({ version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: page })
    expect(parseWorkerStatus({ version: 1, type: 'offline-status', state: 'ready', releaseId: 'not-a-release', pageId: page, template: '{customer: secret}' })).toBeUndefined()
    expect(parseWorkerStatus({ version: 2, type: 'offline-status', state: 'ready', releaseId: release, pageId: page })).toBeUndefined()
    expect(parseWorkerStatus({ version: 1, type: 'cache-all', state: 'ready', releaseId: release, pageId: page })).toBeUndefined()
  })

  it('keeps a waiting update visible when the active worker reports ready and rejects a mismatched page release', async () => {
    const channel = new EventTarget() as EventTarget & { register: ReturnType<typeof vi.fn>; controller?: ServiceWorker }
    const installing = new EventTarget() as EventTarget & { state: ServiceWorkerState }
    installing.state = 'installed'
    const active = { postMessage: vi.fn() } as unknown as ServiceWorker
    const registration = new EventTarget() as ServiceWorkerRegistration
    Object.assign(registration, { active, waiting: null, installing })
    channel.register = vi.fn(async () => registration)
    Object.defineProperty(navigator, 'serviceWorker', { configurable: true, value: channel })
    Object.defineProperty(window, 'isSecureContext', { configurable: true, value: true })
    const states: string[] = []
    registerOfflineLifecycle(page, payload, (state) => states.push(state.state))
    await Promise.resolve()
    expect(active.postMessage).toHaveBeenCalledWith({ version: 1, type: 'get-offline-status' })
    Object.assign(registration, { waiting: installing })
    registration.dispatchEvent(new Event('updatefound'))
    installing.dispatchEvent(new Event('statechange'))
    channel.dispatchEvent(new MessageEvent('message', { data: { version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: page } }))
    expect(states).toEqual(['checking', 'update-available', 'update-available'])
    channel.dispatchEvent(new MessageEvent('message', { data: { version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: 'c'.repeat(64) } }))
    expect(states).toEqual(['checking', 'update-available', 'update-available'])
  })

  it('counts only verified, payload-bound asset events and never treats engine readiness as cache completion', () => {
    const initial = { state: 'checking' as const, cacheReady: false, verifiedAssetUrls: [] }
    const verified = parseWorkerProgress({ version: 1, type: 'offline-progress', state: 'verified', assetUrl: '/assets/engine.wasm', releaseId: release, pageId: page })!
    expect(reduceOfflineLifecycle(initial, verified, payload)).toMatchObject({ state: 'caching', verifiedAssetUrls: ['/assets/engine.wasm'], cacheReady: false })
    const malformed = parseWorkerProgress({ version: 1, type: 'offline-progress', state: 'verified', assetUrl: '/assets/engine.wasm', releaseId: release, pageId: page, document: 'secret' })
    expect(malformed).toBeUndefined()
    const unknown = parseWorkerProgress({ version: 1, type: 'offline-progress', state: 'verified', assetUrl: '/assets/not-in-release.js', releaseId: release, pageId: page })!
    expect(reduceOfflineLifecycle(initial, unknown, payload)).toEqual(initial)
    const mismatchedRelease = parseWorkerProgress({ version: 1, type: 'offline-progress', state: 'verified', assetUrl: '/assets/engine.wasm', releaseId: 'c'.repeat(64), pageId: page })!
    expect(reduceOfflineLifecycle(initial, mismatchedRelease, payload)).toEqual(initial)
    expect(reduceOfflineLifecycle(initial, { version: 1, type: 'offline-status', state: 'unavailable', releaseId: release, pageId: page }, payload)).toMatchObject({ state: 'unavailable', cacheReady: false })
  })

  it('removes every live observer callback and converts a stalled status exchange into retryable timeout', async () => {
    vi.useFakeTimers()
    const channel = new EventTarget() as EventTarget & { register: ReturnType<typeof vi.fn>; controller?: ServiceWorker }
    const active = { postMessage: vi.fn() } as unknown as ServiceWorker
    const registration = new EventTarget() as ServiceWorkerRegistration
    Object.assign(registration, { active, waiting: null, installing: null })
    channel.register = vi.fn(async () => registration)
    Object.defineProperty(navigator, 'serviceWorker', { configurable: true, value: channel })
    Object.defineProperty(window, 'isSecureContext', { configurable: true, value: true })
    const states: string[] = []
    const stop = registerOfflineLifecycle(page, payload, (state) => states.push(state.state), 10)
    await Promise.resolve()
    vi.advanceTimersByTime(10)
    expect(states).toContain('unavailable')
    const before = states.length
    stop()
    channel.dispatchEvent(new MessageEvent('message', { data: { version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: page } }))
    expect(states).toHaveLength(before)
    vi.useRealTimers()
  })
})
