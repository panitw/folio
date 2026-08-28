import { brotliCompressSync, constants } from 'node:zlib'
import { existsSync, readdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
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
  const runtimeAssets = assetsFromDist(outputDir).filter((asset) => asset.url !== '/index.html')
  const pageId = pageIdentity(runtimeAssets, workerRevision)
  const index = join(outputDir, 'index.html')
  const html = readFileSync(index, 'utf8').replace('</head>', `<meta name="folio-page-release" content="${pageId}"></head>`)
  if (html === readFileSync(index, 'utf8')) throw new Error('production index has no head for page release identity')
  writeFileSync(index, html)
  const assets = assetsFromDist(outputDir)
  const thaiDictionary = readFileSync(join(root, '..', 'folio-go', 'internal', 'text', 'data', 'thai_words.trie'))
  const wasm = assets.find((asset) => asset.url.endsWith('.wasm'))
  if (!wasm) throw new Error('production build has no wasm runtime asset')
  const release = { version: 2, id: releaseIdentity(assets, workerRevision), pageId, workerRevision, thaiDictionary: { delivery: 'emitted-wasm-digest-witness', sha256: sha256(thaiDictionary), wasmUrl: wasm.url, proof: 'the emitted wasm offline-audit operation reports the embedded thai_words.trie digest' }, assets }
  for (const asset of release.assets.filter((asset) => asset.immutable)) {
    const output = join(outputDir, asset.url.slice(1) + '.br')
    rmSync(output, { force: true })
    writeFileSync(output, brotliCompressSync(readFileSync(join(outputDir, asset.url.slice(1))), { params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_GENERIC, [constants.BROTLI_PARAM_LGWIN]: 22 } }))
  }
  writeFileSync(join(outputDir, 'offline-release-manifest.json'), `${JSON.stringify(release, null, 2)}\n`)
  writeFileSync(join(outputDir, 'sw.js'), serviceWorkerSource(release))
  return release
}

if (process.argv[1] === fileURLToPath(import.meta.url)) generateOfflineRelease()
