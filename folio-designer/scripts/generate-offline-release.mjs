import { brotliCompressSync, constants } from 'node:zlib'
import { existsSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import { serviceWorkerSource } from './offline-service-worker-template.mjs'
import { RELEASE_RUNTIME, normalizePublicPath, pageIdentity, releaseIdentity, sha256 } from './offline-release-contract.mjs'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const dist = join(root, 'dist')

const walk = (directory) => readdirSync(directory, { withFileTypes: true }).flatMap((entry) => entry.isDirectory() ? walk(join(directory, entry.name)) : [join(directory, entry.name)])

export function assetsFromDist(outputDir) {
  const assetsDir = join(outputDir, 'assets')
  const assetFiles = walk(assetsDir).filter((file) => !file.endsWith('.br')).sort()
  if (assetFiles.length === 0) throw new Error('production build emitted no runtime assets')
  return [join(outputDir, 'index.html'), ...assetFiles].map((file) => ({ url: normalizePublicPath(relative(outputDir, file)), sha256: sha256(readFileSync(file)), immutable: file !== join(outputDir, 'index.html') }))
}

export function assertPinnedRuntime(version = process.version) {
  if (version !== RELEASE_RUNTIME) throw new Error(`offline release generation requires Node ${RELEASE_RUNTIME}; received ${version}`)
}

export function generateOfflineRelease(outputDir = dist) {
  assertPinnedRuntime()
  if (!existsSync(join(outputDir, 'index.html'))) throw new Error('build output is required before offline release generation')
  const workerRevision = sha256(readFileSync(join(root, 'scripts', 'offline-service-worker-template.mjs')))
  const initialAssets = assetsFromDist(outputDir)
  const runtimeAssets = initialAssets.filter((asset) => asset.url !== '/index.html')
  const pageId = pageIdentity(runtimeAssets, workerRevision)
  const thaiDictionary = readFileSync(join(root, '..', 'folio-go', 'internal', 'text', 'data', 'thai_words.trie'))
  const wasm = initialAssets.find((asset) => asset.url.endsWith('.wasm'))
  if (!wasm) throw new Error('production build has no wasm runtime asset')
  const releaseId = releaseIdentity(initialAssets, workerRevision)
  for (const asset of initialAssets.filter((asset) => asset.immutable)) {
    const output = join(outputDir, asset.url.slice(1) + '.br')
    rmSync(output, { force: true })
    writeFileSync(output, brotliCompressSync(readFileSync(join(outputDir, asset.url.slice(1))), { params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_GENERIC, [constants.BROTLI_PARAM_LGWIN]: 22 } }))
  }
  const find = (needle) => {
    const asset = initialAssets.find((candidate) => candidate.url.includes(needle))
    if (!asset) throw new Error(`production build has no ${needle} runtime asset`)
    return asset
  }
  const cachedRow = (id, label, asset) => ({ id, label, delivery: 'cached-asset', assetUrl: asset.url, bytes: statSync(join(outputDir, `${asset.url.slice(1)}.br`)).size, sha256: asset.sha256 })
  const engine = find('.wasm')
  const latin = find('/noto-sans.')
  const thai = find('/noto-sans-thai.')
  const cjk = find('/noto-sans-cjk.')
  const rows = [cachedRow('engine', 'Engine', engine), cachedRow('latin-font', 'Latin font', latin), cachedRow('thai-font', 'Thai font', thai), cachedRow('cjk-font', 'CJK font', cjk)]
  const visibleBytes = rows.reduce((total, row) => total + row.bytes, 0)
  if (!rows.every((row) => Number.isSafeInteger(row.bytes) && row.bytes > 0)) throw new Error('S1 payload rows are incomplete')
  const s1 = { version: 1, releaseId, pageId, unit: 'MiB', decimals: 2, cachedBytes: 0, assetCount: initialAssets.length, cacheAssets: initialAssets.map((asset) => ({ assetUrl: asset.url, bytes: statSync(join(outputDir, asset.url.slice(1))).size })), rows: [...rows, { id: 'thai-dictionary', label: 'Thai dictionary', delivery: 'embedded-in-engine', assetUrl: engine.url, bytes: thaiDictionary.byteLength, sha256: sha256(thaiDictionary) }] }
  const index = join(outputDir, 'index.html')
  const originalHtml = readFileSync(index, 'utf8')
  let bootstrappedHtml = originalHtml
  // The bootstrap is cached inside index.html. Its own byte length is the only
  // self-reference, so converge that decimal value before hashing the final page.
  for (let attempt = 0; attempt < 4; attempt++) {
    const otherBytes = initialAssets.filter((asset) => asset.url !== '/index.html').reduce((total, asset) => total + statSync(join(outputDir, asset.url.slice(1))).size, 0)
    const indexAsset = s1.cacheAssets.find((asset) => asset.assetUrl === '/index.html')
    if (!indexAsset) throw new Error('S1 cache asset list has no navigation entry')
    indexAsset.bytes = s1.cachedBytes - otherBytes
    const bootstrap = `<meta name="folio-page-release" content="${pageId}"><script id="folio-release-bootstrap" type="application/json">${JSON.stringify({ s1 })}</script>`
    bootstrappedHtml = originalHtml.replace('</head>', `${bootstrap}</head>`)
    if (bootstrappedHtml === originalHtml) throw new Error('production index has no head for release bootstrap')
    const nextBytes = otherBytes + Buffer.byteLength(bootstrappedHtml)
    if (nextBytes === s1.cachedBytes) break
    s1.cachedBytes = nextBytes
  }
  writeFileSync(index, bootstrappedHtml)
  const assets = assetsFromDist(outputDir)
  const release = { version: 3, id: releaseId, pageId, workerRevision, thaiDictionary: { delivery: 'emitted-wasm-digest-witness', sha256: sha256(thaiDictionary), wasmUrl: wasm.url, proof: 'the emitted wasm offline-audit operation reports the embedded thai_words.trie digest' }, assets, s1, s1VisibleBytes: visibleBytes }
  writeFileSync(join(outputDir, 'offline-release-manifest.json'), `${JSON.stringify(release, null, 2)}\n`)
  writeFileSync(join(outputDir, 'sw.js'), serviceWorkerSource(release))
  return release
}

if (process.argv[1] === fileURLToPath(import.meta.url)) generateOfflineRelease()
