export type S1Row = Readonly<{ id: 'engine' | 'latin-font' | 'thai-font' | 'cjk-font' | 'thai-dictionary'; label: 'Engine' | 'Latin font' | 'Thai font' | 'CJK font' | 'Thai dictionary'; delivery: 'cached-asset' | 'embedded-in-engine'; assetUrl: string; bytes: number; sha256: string }>
export type S1Payload = Readonly<{ version: 1; releaseId: string; pageId: string; unit: 'MiB'; decimals: 2; cachedBytes: number; assetCount: number; cacheAssets: readonly Readonly<{ assetUrl: string; bytes: number }>[]; rows: readonly S1Row[] }>

const hash = /^[a-f0-9]{64}$/
const ids = ['engine', 'latin-font', 'thai-font', 'cjk-font', 'thai-dictionary'] as const
const labels = ['Engine', 'Latin font', 'Thai font', 'CJK font', 'Thai dictionary'] as const

export function parseS1Payload(value: unknown): S1Payload | undefined {
  if (!value || typeof value !== 'object') return undefined
  const candidate = value as Record<string, unknown>
  if (Object.keys(candidate).length !== 9 || candidate.version !== 1 || !hash.test(String(candidate.releaseId)) || !hash.test(String(candidate.pageId)) || candidate.unit !== 'MiB' || candidate.decimals !== 2 || typeof candidate.cachedBytes !== 'number' || !Number.isSafeInteger(candidate.cachedBytes) || candidate.cachedBytes <= 0 || typeof candidate.assetCount !== 'number' || !Number.isSafeInteger(candidate.assetCount) || candidate.assetCount !== 10 || !Array.isArray(candidate.cacheAssets) || candidate.cacheAssets.length !== candidate.assetCount || !Array.isArray(candidate.rows) || candidate.rows.length !== ids.length) return undefined
  const cacheAssets = candidate.cacheAssets as unknown[]
  if (new Set(cacheAssets.map((asset) => typeof asset === 'object' && asset ? (asset as Record<string, unknown>).assetUrl : undefined)).size !== candidate.assetCount || !cacheAssets.every((asset) => { const item = asset as Record<string, unknown>; return asset && typeof asset === 'object' && Object.keys(item).length === 2 && typeof item.assetUrl === 'string' && item.assetUrl.startsWith('/') && item.assetUrl.length <= 256 && typeof item.bytes === 'number' && Number.isSafeInteger(item.bytes) && item.bytes > 0 })) return undefined
  if (cacheAssets.reduce<number>((total, asset) => total + (asset as { bytes: number }).bytes, 0) !== candidate.cachedBytes) return undefined
  const rows: S1Row[] = []
  for (let index = 0; index < ids.length; index++) {
    const row = candidate.rows[index]
    if (!row || typeof row !== 'object') return undefined
    const item = row as Record<string, unknown>
    if (Object.keys(item).length !== 6 || item.id !== ids[index] || item.label !== labels[index] || (item.delivery !== 'cached-asset' && item.delivery !== 'embedded-in-engine') || typeof item.assetUrl !== 'string' || !item.assetUrl.startsWith('/') || item.assetUrl.length > 256 || typeof item.bytes !== 'number' || !Number.isSafeInteger(item.bytes) || item.bytes <= 0 || typeof item.sha256 !== 'string' || !hash.test(item.sha256)) return undefined
    rows.push(item as S1Row)
  }
  const cached = rows.filter((row) => row.delivery === 'cached-asset')
  if (cached.length !== 4 || rows[4].delivery !== 'embedded-in-engine' || rows[4].assetUrl !== rows[0].assetUrl) return undefined
  return { version: 1, releaseId: candidate.releaseId as string, pageId: candidate.pageId as string, unit: 'MiB', decimals: 2, cachedBytes: candidate.cachedBytes, assetCount: candidate.assetCount, cacheAssets: cacheAssets as { assetUrl: string; bytes: number }[], rows }
}

export function loadS1Payload(): S1Payload | undefined {
  const node = document.getElementById('folio-release-bootstrap')
  if (!node?.textContent) return undefined
  try { return parseS1Payload((JSON.parse(node.textContent) as { s1?: unknown }).s1) } catch { return undefined }
}

export function formatMiB(bytes: number, decimals = 2): string { return `${(bytes / (1024 * 1024)).toFixed(decimals)} MiB` }
