import { describe, expect, it, vi } from 'vitest'
import { parseWorkerStatus, registerOfflineLifecycle } from './offline-lifecycle'

const release = 'a'.repeat(64)
const page = 'b'.repeat(64)

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
    registerOfflineLifecycle(page, (state) => states.push(state))
    await Promise.resolve()
    expect(active.postMessage).toHaveBeenCalledWith({ version: 1, type: 'get-offline-status' })
    Object.assign(registration, { waiting: installing })
    registration.dispatchEvent(new Event('updatefound'))
    installing.dispatchEvent(new Event('statechange'))
    channel.dispatchEvent(new MessageEvent('message', { data: { version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: page } }))
    expect(states).toEqual(['checking', 'update-available'])
    channel.dispatchEvent(new MessageEvent('message', { data: { version: 1, type: 'offline-status', state: 'ready', releaseId: release, pageId: 'c'.repeat(64) } }))
    expect(states).toEqual(['checking', 'update-available'])
  })
})
