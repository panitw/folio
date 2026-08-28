export type OfflineLifecycleState = 'checking' | 'ready' | 'update-available' | 'unavailable'

const MESSAGE_VERSION = 1
const releaseIdPattern = /^[a-f0-9]{64}$/

type WorkerStatus = Readonly<{ version: number; type: 'offline-status'; state: 'ready' | 'unavailable'; releaseId: string; pageId: string }>

export function parseWorkerStatus(value: unknown): WorkerStatus | undefined {
  if (!value || typeof value !== 'object') return undefined
  const candidate = value as Record<string, unknown>
  if (candidate.version !== MESSAGE_VERSION || candidate.type !== 'offline-status' || Object.keys(candidate).length !== 5) return undefined
  if (candidate.state !== 'ready' && candidate.state !== 'unavailable') return undefined
  if (typeof candidate.releaseId !== 'string' || !releaseIdPattern.test(candidate.releaseId) || typeof candidate.pageId !== 'string' || !releaseIdPattern.test(candidate.pageId)) return undefined
  return candidate as WorkerStatus
}

export function registerOfflineLifecycle(expectedPageId: string | undefined, onState: (state: OfflineLifecycleState) => void): void {
  if (!expectedPageId || !releaseIdPattern.test(expectedPageId) || !('serviceWorker' in navigator) || !window.isSecureContext) { onState('unavailable'); return }
  let current: OfflineLifecycleState | undefined
  const publish = (next: OfflineLifecycleState) => {
    // A complete waiting worker is more important than an active worker's
    // ready acknowledgement; never race that truthful update state away.
    if (current === 'update-available' && next === 'ready') return
    if (current === next) return
    current = next
    onState(next)
  }
  publish('checking')
  navigator.serviceWorker.addEventListener('message', (event: MessageEvent<unknown>) => {
    const status = parseWorkerStatus(event.data)
    if (!status) return
    if (status.pageId !== expectedPageId) { if (current !== 'update-available') publish('unavailable'); return }
    publish(status.state)
  })
  void navigator.serviceWorker.register('/sw.js', { scope: '/' }).then(async (registration) => {
    const observeInstalling = (installing: ServiceWorker | null) => {
      if (!installing) return
      installing.addEventListener('statechange', () => {
        if (installing.state === 'installed' && registration.waiting) publish('update-available')
        if (installing.state === 'redundant' && !registration.active && !navigator.serviceWorker.controller) publish('unavailable')
      })
    }
    registration.addEventListener('updatefound', () => observeInstalling(registration.installing))
    observeInstalling(registration.installing)
    if (registration.waiting) publish('update-available')
    const active = registration.active ?? navigator.serviceWorker.controller
    active?.postMessage({ version: MESSAGE_VERSION, type: 'get-offline-status' })
  }).catch(() => publish('unavailable'))
}
