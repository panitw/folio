import { brotliCompressSync, brotliDecompressSync, constants } from 'node:zlib'
import { existsSync, readFileSync, readdirSync, unlinkSync, writeFileSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'
import { execFileSync } from 'node:child_process'
import { serviceWorkerSource } from './offline-service-worker-template.mjs'
import { assertPinnedRuntime } from './generate-offline-release.mjs'
import { RELEASE_RUNTIME, pageIdentity, releaseIdentity, sha256 } from './offline-release-contract.mjs'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const dist = join(root, 'dist')
const fail = (message) => { throw new Error(`offline release verification failed: ${message}`) }
const sameSet = (left, right) => left.size === right.size && [...left].every((value) => right.has(value))
const brotliOptions = { params: { [constants.BROTLI_PARAM_QUALITY]: 11, [constants.BROTLI_PARAM_MODE]: constants.BROTLI_MODE_GENERIC, [constants.BROTLI_PARAM_LGWIN]: 22 } }
const walk = (directory) => readdirSync(directory, { withFileTypes: true }).flatMap((entry) => entry.isDirectory() ? walk(join(directory, entry.name)) : [join(directory, entry.name)])
const runtimeOutputUrls = (outputDir) => new Set(['/index.html', ...walk(join(outputDir, 'assets')).filter((file) => !file.endsWith('.br')).map((file) => `/${relative(outputDir, file).replaceAll('\\', '/')}`)])

export function verifyOfflineRelease(outputDir = dist, { wasmWitness = false } = {}) {
  assertPinnedRuntime()
  const manifestFile = join(outputDir, 'offline-release-manifest.json')
  if (!existsSync(manifestFile)) fail('missing generated manifest')
  const release = JSON.parse(readFileSync(manifestFile, 'utf8'))
  const contract = JSON.parse(readFileSync(join(root, 'static-host-contract.json'), 'utf8'))
  if (release.version !== 2 || !Array.isArray(release.assets) || release.assets.length < 2 || !/^[a-f0-9]{64}$/.test(release.id) || !/^[a-f0-9]{64}$/.test(release.pageId) || !/^[a-f0-9]{64}$/.test(release.workerRevision)) fail('manifest has no complete release identity')
  const manifestUrls = new Set()
  for (const asset of release.assets) {
    if (!asset || typeof asset.url !== 'string' || !asset.url.startsWith('/') || asset.url.includes('\\') || manifestUrls.has(asset.url)) fail('manifest URL is duplicate or invalid')
    manifestUrls.add(asset.url)
  }
  const outputUrls = runtimeOutputUrls(outputDir)
  if (!sameSet(manifestUrls, outputUrls)) fail('manifest and production runtime output are not an exact set')
  if (!manifestUrls.has('/index.html')) fail('navigation entry is absent')
  for (const required of ['.wasm', '.css', '.js', '.ttf', '.folio']) if (!release.assets.some((asset) => asset.url.endsWith(required))) fail(`missing required runtime class ${required}`)
  if (release.id !== releaseIdentity(release.assets, release.workerRevision)) fail('release identity does not match canonical assets and worker revision')
  if (release.pageId !== pageIdentity(release.assets, release.workerRevision)) fail('page release identity does not match runtime assets and worker revision')
  if (!readFileSync(join(outputDir, 'index.html'), 'utf8').includes(`name="folio-page-release" content="${release.pageId}"`)) fail('page does not bind to its release identity')
  const sw = readFileSync(join(outputDir, 'sw.js'), 'utf8')
  const embedded = sw.match(/const RELEASE = (.+)\nconst CACHE_NAME/m)?.[1]
  if (!embedded || JSON.stringify(JSON.parse(embedded)) !== JSON.stringify(release)) fail('service worker and manifest release records differ')
  for (const required of ["const CACHE_NAME = 'folio-release-' + RELEASE.id", "credentials: 'omit'", "credentials === 'omit'", 'RELEASE.pageId', 'windows.length === 0', 'offline asset integrity mismatch']) if (!sw.includes(required)) fail(`service worker lacks ${required}`)
  if (sw.includes('skipWaiting') || sw.includes('cache.addAll') || sw.includes('fetch(event.request)')) fail('service worker has unsafe activation or generic network fallback')
  if (release.thaiDictionary?.delivery !== 'emitted-wasm-digest-witness' || !manifestUrls.has(release.thaiDictionary.wasmUrl) || !/^[a-f0-9]{64}$/.test(release.thaiDictionary.sha256)) fail('Thai dictionary containment is not declared against the emitted wasm')
  const thaiSource = readFileSync(join(root, '..', 'folio-go', 'internal', 'text', 'data', 'thai_words.trie'))
  if (sha256(thaiSource) !== release.thaiDictionary.sha256) fail('Thai dictionary audit digest is stale')
  for (const asset of release.assets) {
    const file = join(outputDir, asset.url.slice(1))
    if (!existsSync(file)) fail(`missing manifest asset ${asset.url}`)
    if (sha256(readFileSync(file)) !== asset.sha256) fail(`stale asset digest ${asset.url}`)
  }
  for (const asset of release.assets) {
    const file = join(outputDir, asset.url.slice(1))
    if (asset.immutable) {
      if (!asset.url.startsWith(contract.immutableRuntime.urlPrefix) || !/-[A-Za-z0-9_-]{8,}\.[A-Za-z0-9]+$/.test(asset.url)) fail(`immutable URL is not content-addressed ${asset.url}`)
      const sidecar = `${file}${contract.immutableRuntime.sidecarSuffix}`
      if (!existsSync(sidecar)) fail(`missing Brotli sidecar ${asset.url}`)
      const original = readFileSync(file)
      const compressed = readFileSync(sidecar)
      if (!brotliDecompressSync(compressed).equals(original)) fail(`stale Brotli sidecar ${asset.url}`)
      if (!brotliCompressSync(original, brotliOptions).equals(compressed)) fail(`non-deterministic Brotli sidecar ${asset.url} under ${RELEASE_RUNTIME}`)
    } else if (asset.url !== '/index.html') fail(`unexpected mutable runtime entry ${asset.url}`)
  }
  if (wasmWitness) {
    const glue = release.assets.find((asset) => asset.url.includes('/wasm-exec.'))
    if (!glue) fail('wasm runtime glue is absent from the release')
    const wasmDigest = execFileSync(process.execPath, [join(root, 'scripts', 'verify-wasm-dictionary.mjs'), join(outputDir, glue.url.slice(1)), join(outputDir, release.thaiDictionary.wasmUrl.slice(1))], { encoding: 'utf8', timeout: 30000 }).trim()
    if (wasmDigest !== release.thaiDictionary.sha256) fail('emitted wasm dictionary witness does not match the shipped dictionary')
  }
  if (!contract.immutableRuntime.cacheControl.includes('immutable') || contract.immutableRuntime.contentEncoding !== 'br') fail('immutable host policy is incomplete')
  if (!contract.updateableEntries.some((entry) => entry.url === '/index.html' && !entry.cacheControl.includes('immutable')) || !contract.updateableEntries.some((entry) => entry.url === '/sw.js' && !entry.cacheControl.includes('immutable'))) fail('updateable entries have immutable policy')
  return release
}

function rewriteRelease(outputDir, release) {
  release.pageId = pageIdentity(release.assets, release.workerRevision)
  release.id = releaseIdentity(release.assets, release.workerRevision)
  writeFileSync(join(outputDir, 'offline-release-manifest.json'), `${JSON.stringify(release, null, 2)}\n`)
  writeFileSync(join(outputDir, 'sw.js'), serviceWorkerSource(release))
}

function redProof(name, mutate) {
  let restore
  try {
    restore = mutate(dist)
    let failed = false
    try { verifyOfflineRelease(dist) } catch { failed = true }
    if (!failed) fail(`red proof ${name} escaped verification`)
  } finally { restore?.() }
}

export function runRedProofs(baseline = verifyOfflineRelease()) {
  const readRelease = (outputDir) => JSON.parse(readFileSync(join(outputDir, 'offline-release-manifest.json'), 'utf8'))
  for (const extension of ['.wasm', '.ttf']) redProof(`stale-${extension.slice(1)}-byte`, (outputDir) => { const witness = baseline.assets.find((candidate) => candidate.immutable && candidate.url.endsWith(extension)); if (!witness) fail(`red proof has no ${extension} witness`); const file = join(outputDir, witness.url.slice(1)); const original = readFileSync(file); writeFileSync(file, Buffer.concat([original, Buffer.from([0])])); return () => writeFileSync(file, original) })
  for (const [name, extension] of [['missing-wasm-row', '.wasm'], ['missing-font-row', '.ttf']]) redProof(name, (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const worker = join(outputDir, 'sw.js'); const oldManifest = readFileSync(manifest); const oldWorker = readFileSync(worker); const release = readRelease(outputDir); release.assets = release.assets.filter((asset) => !asset.url.endsWith(extension)); rewriteRelease(outputDir, release); return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) } })
  redProof('unmanifested-runtime-output', (outputDir) => { const injected = join(outputDir, 'assets', 'injected-12345678.js'); writeFileSync(injected, 'export {}'); return () => unlinkSync(injected) })
  redProof('stale-release-identity', (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const original = readFileSync(manifest); const release = JSON.parse(original); release.id = '0'.repeat(64); writeFileSync(manifest, JSON.stringify(release)); return () => writeFileSync(manifest, original) })
  redProof('worker-manifest-drift', (outputDir) => { const worker = join(outputDir, 'sw.js'); const original = readFileSync(worker, 'utf8'); const drifted = original.replace(`"pageId":"${baseline.pageId}"`, `"pageId":"${'0'.repeat(64)}"`); if (drifted === original) fail('red proof could not locate the worker release record'); writeFileSync(worker, drifted); return () => writeFileSync(worker, original) })
  redProof('dictionary-witness-mismatch', (outputDir) => {
    const release = readRelease(outputDir)
    const manifest = join(outputDir, 'offline-release-manifest.json')
    const worker = join(outputDir, 'sw.js')
    const oldManifest = readFileSync(manifest)
    const oldWorker = readFileSync(worker)
    release.thaiDictionary.sha256 = '0'.repeat(64)
    rewriteRelease(outputDir, release)
    return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) }
  })
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const redOnly = process.argv.includes('--red-only')
  const wasmWitness = process.argv.includes('--wasm-witness')
  const baseline = redOnly ? JSON.parse(readFileSync(join(dist, 'offline-release-manifest.json'), 'utf8')) : verifyOfflineRelease(dist, { wasmWitness })
  if (process.argv.includes('--red-proof') || redOnly) runRedProofs(baseline)
}
