export type S1Row = Readonly<{ id: 'engine' | 'latin-font' | 'thai-font' | 'cjk-font' | 'thai-dictionary'; label: 'Engine' | 'Latin font' | 'Thai font' | 'CJK font' | 'Thai dictionary'; delivery: 'cached-asset' | 'embedded-in-engine'; assetUrl: string; bytes: number; sha256: string }>
export type S1Payload = Readonly<{ version: 1; releaseId: string; pageId: string; unit: 'MiB'; decimals: 2; cachedBytes: number; assetCount: number; cacheAssets: readonly Readonly<{ assetUrl: string; bytes: number }>[]; rows: readonly S1Row[] }>
// EVERY REJECTION CARRIES ITS OWN NAME. The bound and the fifteen unrelated shape
// checks used to share one bare `undefined`, so "this release lists more assets
// than the reader accepts" and "this is not a payload at all" were the same
// value — an all-clear indistinguishable from a couldn't-look, on the first
// screen a user sees. `asset-count-over-maximum` in particular is reachable by
// that cause and no other, which is what makes it evidence.
export type S1PayloadRejection =
  | 'not-an-object'
  | 'payload-shape'
  | 'asset-count-over-maximum'
  | 'asset-count-under-minimum'
  | 'cache-assets-invalid'
  | 'cached-bytes-mismatch'
  | 'row-not-an-object'
  | 'row-shape'
  | 'row-delivery-composition'
  | 'no-bootstrap'
  | 'malformed-json'
export type S1PayloadResult = Readonly<{ ok: true; payload: S1Payload }> | Readonly<{ ok: false; reason: S1PayloadRejection }>

const hash = /^[a-f0-9]{64}$/
const ids = ['engine', 'latin-font', 'thai-font', 'cjk-font', 'thai-dictionary'] as const
const labels = ['Engine', 'Latin font', 'Thai font', 'CJK font', 'Thai dictionary'] as const
// scripts/offline-release-contract.mjs DERIVES these two numbers from these two
// lines and scripts/verify-offline-release.mjs fails the build on a release
// outside them, so a release over the bound can no longer be emitted in silence.
// Keep each on its own line as `const <name> = <digits>`: that reader is
// line-anchored and requires exactly one live match, so reformatting, renaming
// or duplicating either line fails the build loudly rather than disabling it.
const minimumCacheAssets = 10
const maximumCacheAssets = 64
const reject = (reason: S1PayloadRejection): S1PayloadResult => ({ ok: false, reason })

export function parseS1Payload(value: unknown): S1PayloadResult {
  if (!value || typeof value !== 'object') return reject('not-an-object')
  const candidate = value as Record<string, unknown>
  if (Object.keys(candidate).length !== 9 || candidate.version !== 1 || !hash.test(String(candidate.releaseId)) || !hash.test(String(candidate.pageId)) || candidate.unit !== 'MiB' || candidate.decimals !== 2 || typeof candidate.cachedBytes !== 'number' || !Number.isSafeInteger(candidate.cachedBytes) || candidate.cachedBytes <= 0 || typeof candidate.assetCount !== 'number' || !Number.isSafeInteger(candidate.assetCount) || !Array.isArray(candidate.cacheAssets) || candidate.cacheAssets.length !== candidate.assetCount || !Array.isArray(candidate.rows) || candidate.rows.length !== ids.length) return reject('payload-shape')
  // The bound lives OUTSIDE the shape condition above so each arm is reachable by
  // its own cause alone. The condition above has already established that
  // `assetCount` is a safe integer and that `cacheAssets` has exactly that length.
  if (candidate.assetCount > maximumCacheAssets) return reject('asset-count-over-maximum')
  if (candidate.assetCount < minimumCacheAssets) return reject('asset-count-under-minimum')
  const cacheAssets = candidate.cacheAssets as unknown[]
  if (new Set(cacheAssets.map((asset) => typeof asset === 'object' && asset ? (asset as Record<string, unknown>).assetUrl : undefined)).size !== candidate.assetCount || !cacheAssets.every((asset) => { const item = asset as Record<string, unknown>; return asset && typeof asset === 'object' && Object.keys(item).length === 2 && typeof item.assetUrl === 'string' && item.assetUrl.startsWith('/') && item.assetUrl.length <= 256 && typeof item.bytes === 'number' && Number.isSafeInteger(item.bytes) && item.bytes > 0 })) return reject('cache-assets-invalid')
  if (cacheAssets.reduce<number>((total, asset) => total + (asset as { bytes: number }).bytes, 0) !== candidate.cachedBytes) return reject('cached-bytes-mismatch')
  const rows: S1Row[] = []
  for (let index = 0; index < ids.length; index++) {
    const row = candidate.rows[index]
    if (!row || typeof row !== 'object') return reject('row-not-an-object')
    const item = row as Record<string, unknown>
    if (Object.keys(item).length !== 6 || item.id !== ids[index] || item.label !== labels[index] || (item.delivery !== 'cached-asset' && item.delivery !== 'embedded-in-engine') || typeof item.assetUrl !== 'string' || !item.assetUrl.startsWith('/') || item.assetUrl.length > 256 || typeof item.bytes !== 'number' || !Number.isSafeInteger(item.bytes) || item.bytes <= 0 || typeof item.sha256 !== 'string' || !hash.test(item.sha256)) return reject('row-shape')
    rows.push(item as S1Row)
  }
  const cached = rows.filter((row) => row.delivery === 'cached-asset')
  if (cached.length !== 4 || rows[4].delivery !== 'embedded-in-engine' || rows[4].assetUrl !== rows[0].assetUrl) return reject('row-delivery-composition')
  return { ok: true, payload: { version: 1, releaseId: candidate.releaseId as string, pageId: candidate.pageId as string, unit: 'MiB', decimals: 2, cachedBytes: candidate.cachedBytes, assetCount: candidate.assetCount, cacheAssets: cacheAssets as { assetUrl: string; bytes: number }[], rows } }
}

export function loadS1Payload(): S1PayloadResult {
  const node = document.getElementById('folio-release-bootstrap')
  const text = node?.textContent
  if (!text) return reject('no-bootstrap')
  let bootstrap: unknown
  // THE TRY WRAPS EXACTLY ONE CALL, AND THE CATCH ADMITS EXACTLY ONE ERROR.
  // It used to wrap the parse AND parseS1Payload and convert anything thrown to a
  // bare `undefined` — but parseS1Payload never throws, so the only throw this
  // ever existed for is JSON.parse's SyntaxError. Anything else is a fault in the
  // page, not a malformed bootstrap, and is rethrown rather than disguised as one.
  //
  // It is narrowed rather than removed because removal is measurably worse.
  // main.tsx calls render() BEFORE its try, the block has a finally and no catch,
  // and startObservation is invoked as `void startObservation()`, so a propagating
  // SyntaxError would never reach registerOfflineLifecycle: instead of the stated
  // "Offline cache unavailable" phase and its Retry preparation button, the load
  // screen would sit on "Checking cache" forever with no message and no retry —
  // an unstated hang traded for a stated failure, in the name of removing silence.
  try { bootstrap = JSON.parse(text) } catch (error) { if (error instanceof SyntaxError) return reject('malformed-json'); throw error }
  // `JSON.parse('null')` succeeds, and reading `.s1` off it would throw a
  // TypeError the narrowed catch no longer covers. Guard the read instead of
  // widening the catch back: a null bootstrap must keep reaching the parser as a
  // rejection, exactly as it did before, or narrowing would have changed what a
  // user sees on a path this story is not scoped to change.
  return parseS1Payload(bootstrap && typeof bootstrap === 'object' ? (bootstrap as { s1?: unknown }).s1 : undefined)
}

// THE TWO DECISIONS main.tsx MAKES ABOUT A RESULT, WHERE A TEST CAN EXECUTE THEM.
// Nothing in this repo imports main.tsx, and Vitest collects only
// `src/**/*.test.{ts,tsx}` and `scripts/**/*.test.mjs`, so a decision left inline
// there is run by NO gate: `payloadForLifecycle` collapsed to a bare `undefined`
// would type-check, pass every suite, and ship an app permanently reporting
// "Offline cache unavailable". Both decisions are pure functions of the result,
// so they live here and are asserted in release-payload.test.ts.
export function payloadForLifecycle(result: S1PayloadResult): S1Payload | undefined {
  // registerOfflineLifecycle and the load screen already speak `undefined` for
  // "no usable payload", so a rejection maps to that at exactly one place and
  // nothing downstream changes. The reason is not discarded on the way past — it
  // is what the predicate below reads.
  return result.ok ? result.payload : undefined
}

// The dev-server bypass covers ONE cause: the dev server emits no bootstrap node
// at all. A bootstrap that is malformed, or over the release bound, is a real
// fault and must never be read as "the dev server did not emit one" and quietly
// bypassed — so this is true for `no-bootstrap` and for no other reason.
export function isDevBypassReason(result: S1PayloadResult): boolean {
  return !result.ok && result.reason === 'no-bootstrap'
}

export function formatMiB(bytes: number, decimals = 2): string { return `${(bytes / (1024 * 1024)).toFixed(decimals)} MiB` }
