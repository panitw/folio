import type { S1Payload } from './release-payload'

export type OfflineLifecycleState = 'checking' | 'caching' | 'ready' | 'update-available' | 'unavailable'
export type OfflineLifecycle = Readonly<{ state: OfflineLifecycleState; cacheReady: boolean; verifiedAssetUrls: readonly string[]; activeAssetUrl?: string; failedAssetUrl?: string; failure?: 'timeout' | 'install' | 'unsupported' }>
const MESSAGE_VERSION = 1
const releaseIdPattern = /^[a-f0-9]{64}$/
type WorkerStatus = Readonly<{ version: number; type: 'offline-status'; state: 'ready' | 'unavailable'; releaseId: string; pageId: string }>
type WorkerProgress = Readonly<{ version: number; type: 'offline-progress'; state: 'active' | 'verified' | 'failed'; assetUrl: string | null; releaseId: string; pageId: string }>

function identity(candidate: Record<string, unknown>): boolean { return candidate.version === MESSAGE_VERSION && typeof candidate.releaseId === 'string' && releaseIdPattern.test(candidate.releaseId) && typeof candidate.pageId === 'string' && releaseIdPattern.test(candidate.pageId) }
export function parseWorkerStatus(value: unknown): WorkerStatus | undefined { if (!value || typeof value !== 'object') return undefined; const candidate = value as Record<string, unknown>; return Object.keys(candidate).length === 5 && identity(candidate) && candidate.type === 'offline-status' && (candidate.state === 'ready' || candidate.state === 'unavailable') ? candidate as WorkerStatus : undefined }
export function parseWorkerProgress(value: unknown): WorkerProgress | undefined { if (!value || typeof value !== 'object') return undefined; const candidate = value as Record<string, unknown>; return Object.keys(candidate).length === 6 && identity(candidate) && candidate.type === 'offline-progress' && (candidate.state === 'active' || candidate.state === 'verified' || candidate.state === 'failed') && (candidate.assetUrl === null || (typeof candidate.assetUrl === 'string' && candidate.assetUrl.startsWith('/') && candidate.assetUrl.length <= 256)) ? candidate as WorkerProgress : undefined }
const matches = (event: WorkerStatus | WorkerProgress, payload: S1Payload) => event.releaseId === payload.releaseId && event.pageId === payload.pageId
const known = (url: string, payload: S1Payload) => payload.cacheAssets.some((asset) => asset.assetUrl === url)

export function reduceOfflineLifecycle(previous: OfflineLifecycle, event: WorkerStatus | WorkerProgress | 'checking' | 'update-available' | 'unavailable' | 'timeout', payload?: S1Payload): OfflineLifecycle {
  if (typeof event === 'string') {
    if (event === 'checking') return previous.cacheReady ? previous : { ...previous, state: 'checking' }
    if (event === 'update-available') return { ...previous, state: 'update-available' }
    return { ...previous, state: 'unavailable', cacheReady: false, failure: event === 'timeout' ? 'timeout' : previous.failure ?? 'install' }
  }
  if (!payload || !matches(event, payload)) return previous
  if (event.type === 'offline-status') return event.state === 'ready' ? { ...previous, state: previous.state === 'update-available' ? 'update-available' : 'ready', cacheReady: true, activeAssetUrl: undefined, failedAssetUrl: undefined, failure: undefined } : previous.cacheReady ? previous : { ...previous, state: 'unavailable', failure: 'install' }
  if (event.state === 'failed') return previous.cacheReady ? previous : { ...previous, state: 'unavailable', activeAssetUrl: undefined, failedAssetUrl: event.assetUrl ?? undefined, failure: 'install' }
  if (!event.assetUrl || !known(event.assetUrl, payload)) return previous
  if (event.state === 'active') return { ...previous, state: 'caching', activeAssetUrl: event.assetUrl }
  if (!previous.verifiedAssetUrls.includes(event.assetUrl)) return { ...previous, state: 'caching', activeAssetUrl: undefined, verifiedAssetUrls: [...previous.verifiedAssetUrls, event.assetUrl] }
  return previous
}

export function registerOfflineLifecycle(expectedPageId: string | undefined, payload: S1Payload | undefined, onState: (state: OfflineLifecycle) => void, timeoutMs = 8000): () => void {
  let current: OfflineLifecycle = { state: 'checking', cacheReady: false, verifiedAssetUrls: [] }
  let disposed = false
  let timeout: ReturnType<typeof setTimeout> | undefined
  const publish = (event: WorkerStatus | WorkerProgress | 'checking' | 'update-available' | 'unavailable' | 'timeout') => { if (disposed) return; const next = reduceOfflineLifecycle(current, event, payload); if (JSON.stringify(next) !== JSON.stringify(current)) { current = next; onState(next) } }
  if (!expectedPageId || !payload || expectedPageId !== payload.pageId || !releaseIdPattern.test(expectedPageId) || !('serviceWorker' in navigator) || !window.isSecureContext) { publish('unavailable'); return () => { disposed = true } }
  onState(current)
  const onMessage = (event: MessageEvent<unknown>) => { const message = parseWorkerStatus(event.data) ?? parseWorkerProgress(event.data); if (message) publish(message) }
  navigator.serviceWorker.addEventListener('message', onMessage)
  timeout = setTimeout(() => publish('timeout'), timeoutMs)
  let registration: ServiceWorkerRegistration | undefined
  let installing: ServiceWorker | null = null
  const onStateChange = () => { if (!registration || !installing || disposed) return; if (installing.state === 'installed' && registration.waiting) publish('update-available'); if (installing.state === 'redundant' && !navigator.serviceWorker.controller) publish('unavailable') }
  const onUpdateFound = () => { if (!registration) return; if (installing) installing.removeEventListener('statechange', onStateChange); installing = registration.installing; installing?.addEventListener('statechange', onStateChange); onStateChange() }
  void navigator.serviceWorker.register('/sw.js', { scope: '/' }).then((registered) => { if (disposed) return; registration = registered; registration.addEventListener('updatefound', onUpdateFound); onUpdateFound(); if (registration.waiting) publish('update-available'); (registration.active ?? navigator.serviceWorker.controller)?.postMessage({ version: MESSAGE_VERSION, type: 'get-offline-status' }) }).catch(() => publish('unavailable'))
  return () => { disposed = true; if (timeout) clearTimeout(timeout); navigator.serviceWorker.removeEventListener('message', onMessage); registration?.removeEventListener('updatefound', onUpdateFound); installing?.removeEventListener('statechange', onStateChange) }
}
