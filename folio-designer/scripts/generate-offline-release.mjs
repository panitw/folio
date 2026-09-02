import { brotliCompressSync, constants } from 'node:zlib'
import { existsSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import { serviceWorkerSource } from './offline-service-worker-template.mjs'
import { RELEASE_RUNTIME, isCatalogueAssetUrl, normalizePublicPath, pageIdentity, releaseIdentity, sha256 } from './offline-release-contract.mjs'

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
  // THE SIDECAR SIZES ARE RECORDED AS THEY ARE WRITTEN (AC5, Story 8.5).
  // Until this story the release said what every asset WEIGHS ON DISK
  // (`s1.cacheAssets`) and what four of them weigh COMPRESSED (the S1 rows),
  // and nothing at all about the compressed weight of the other nineteen — the
  // figure that decides what a first load actually costs over the wire. A
  // catalogue of twenty-one faces makes that gap the majority of the payload.
  //
  // Recorded here rather than re-measured later ON PURPOSE: this is the loop
  // that produces the bytes, so the record cannot describe a sidecar that was
  // never written. `verify-offline-release.mjs` then re-stats every one of them
  // and refuses a release whose record has drifted (`brotli-record-drift`).
  const brotliSidecarBytes = new Map()
  for (const asset of initialAssets.filter((asset) => asset.immutable)) {
    const output = join(outputDir, asset.url.slice(1) + '.br')
    rmSync(output, { force: true })
    writeFileSync(output, brotliCompressSync(readFileSync(join(outputDir, asset.url.slice(1))), { params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_GENERIC, [constants.BROTLI_PARAM_LGWIN]: 22 } }))
    brotliSidecarBytes.set(asset.url, statSync(output).size)
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
  // Regeneration is part of normal local verification. Strip our previous
  // generated bootstrap so a second build:offline run replaces it instead of
  // nesting stale S1 records before the current page identity.
  const originalHtml = readFileSync(index, 'utf8').replace(/<meta name="folio-page-release"[^>]*><script id="folio-release-bootstrap" type="application\/json">[^<]*<\/script>/, '')
  let bootstrappedHtml = originalHtml
  // The bootstrap is cached inside index.html. Its own byte length is the only
  // self-reference, so converge that decimal value before hashing the final page.
  // The S1 payload includes index.html's own emitted size. Keep iterating
  // through the self-reference until both the page bootstrap and manifest
  // describe the same final output (the larger application bundle can require
  // more than the former four passes to settle its decimal widths).
  for (let attempt = 0; attempt < 8; attempt++) {
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
  // EVERY IMMUTABLE ASSET CARRIES ITS OWN COMPRESSED WEIGHT (AC5). A missing
  // entry throws rather than defaulting to zero: a record that silently reads
  // "0 bytes" for an asset nobody compressed is worse than no record, because
  // it sums into a total somebody will quote.
  for (const asset of assets) {
    if (!asset.immutable) continue
    const bytes = brotliSidecarBytes.get(asset.url)
    if (!Number.isSafeInteger(bytes) || bytes <= 0) throw new Error(`no Brotli sidecar size was recorded for immutable asset ${asset.url}`)
    asset.brotliBytes = bytes
  }
  // AND THE CATALOGUE'S SHARE OF IT, AS ONE NUMBER (AC5, D-8.4j.8). Story 8.4d
  // owns the threshold and sets it last against the finished weight; this story
  // RECORDS the weight and sets nothing. A subtotal spread over twenty-one rows
  // is a subtotal nobody adds up, which is precisely how 8.4d would inherit a
  // figure it could not use.
  //
  // The catalogue is recognised by the asset-URL prefix `build-wasm.mjs`
  // fingerprints it under, and the count is held to `font-catalogue.json`'s own
  // length — so a prefix change, a dropped face or a Vite naming change reds
  // here rather than quietly reporting the catalogue as weighing nothing.
  const declaredCatalogue = JSON.parse(readFileSync(join(root, 'font-catalogue.json'), 'utf8'))
  const catalogueAssets = assets.filter((asset) => isCatalogueAssetUrl(asset.url))
  if (catalogueAssets.length !== declaredCatalogue.length) throw new Error(`font-catalogue.json declares ${declaredCatalogue.length} catalogue faces and the emitted release carries ${catalogueAssets.length} assets under the catalogue prefix`)
  const immutableAssets = assets.filter((asset) => asset.immutable)
  const brotli = {
    version: 1,
    immutableAssetCount: immutableAssets.length,
    totalBytes: immutableAssets.reduce((total, asset) => total + asset.brotliBytes, 0),
    catalogue: {
      familyCount: catalogueAssets.length,
      // THE ONE NUMBER. Total Brotli bytes the Story 8.5 catalogue adds to the
      // offline release. It is a MEASUREMENT, not a budget, and nothing in this
      // repository compares it to a threshold.
      totalBytes: catalogueAssets.reduce((total, asset) => total + asset.brotliBytes, 0),
    },
  }
  const release = { version: 3, brotli, id: releaseId, pageId, workerRevision, thaiDictionary: { delivery: 'emitted-wasm-digest-witness', sha256: sha256(thaiDictionary), wasmUrl: wasm.url, proof: 'the emitted wasm offline-audit operation reports the embedded thai_words.trie digest' }, assets, s1, s1VisibleBytes: visibleBytes }
  writeFileSync(join(outputDir, 'offline-release-manifest.json'), `${JSON.stringify(release, null, 2)}\n`)
  writeFileSync(join(outputDir, 'sw.js'), serviceWorkerSource(release))
  return release
}

if (process.argv[1] === fileURLToPath(import.meta.url)) generateOfflineRelease()
