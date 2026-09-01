import { brotliCompressSync, brotliDecompressSync, constants } from 'node:zlib'
import { existsSync, readFileSync, readdirSync, rmSync, unlinkSync, writeFileSync } from 'node:fs'
import { dirname, join, relative } from 'node:path'
import { tmpdir } from 'node:os'
import { fileURLToPath } from 'node:url'
import { execFileSync } from 'node:child_process'
import { serviceWorkerSource } from './offline-service-worker-template.mjs'
import { assertPinnedRuntime, generateOfflineRelease } from './generate-offline-release.mjs'
import { assertNoVCSStamp, buildEngineWasm } from './wasm-vcs-stamp.mjs'
import { RELEASE_RUNTIME, declaredCacheAssetBounds, pageIdentity, releaseIdentity, sha256 } from './offline-release-contract.mjs'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')
const dist = join(root, 'dist')
const fail = (message) => { throw new Error(`offline release verification failed: ${message}`) }
// Kept in sync with the `import.meta.env.DEV` offline bypass in src/main.tsx,
// src/App.tsx, and src/offline-lifecycle.ts.
const DEV_BYPASS_MARKERS = ['dev-bypass', 'Offline layer bypassed']
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
  if (release.version !== 3 || !Array.isArray(release.assets) || release.assets.length < 2 || !/^[a-f0-9]{64}$/.test(release.id) || !/^[a-f0-9]{64}$/.test(release.pageId) || !/^[a-f0-9]{64}$/.test(release.workerRevision)) fail('manifest has no complete release identity')
  const manifestUrls = new Set()
  for (const asset of release.assets) {
    if (!asset || typeof asset.url !== 'string' || !asset.url.startsWith('/') || asset.url.includes('\\') || manifestUrls.has(asset.url)) fail('manifest URL is duplicate or invalid')
    manifestUrls.add(asset.url)
  }
  // THE BOUND THE PARSER DECLARES, ENFORCED WHERE A RELEASE IS STILL REFUSABLE.
  // src/release-payload.ts refuses a payload outside [minimum, maximum] cache
  // assets; until this check existed nothing counted, so crossing it was a
  // build-time silence followed by a runtime one.
  //
  // (a) The count is `release.assets.length`, not `s1.assetCount`: the S1 block
  // below already ties them (`s1.assetCount !== release.assets.length` fails
  // there), so bounding this bounds exactly the quantity parseS1Payload bounds —
  // and it is available here, before anything that could trip first.
  //
  // (b) It sits BEFORE the manifest/output set comparison on the next line
  // ON PURPOSE. A release manufactured over the bound necessarily breaks that
  // exact-set check, and the per-asset digest and Brotli loops further down; a
  // bound check placed after any of them could never be red-proved on its own
  // message. Placement is what makes the proof mean something.
  const { minimumCacheAssets, maximumCacheAssets } = declaredCacheAssetBounds()
  if (release.assets.length > maximumCacheAssets) fail(`release carries ${release.assets.length} cache assets, over the declared maximum of ${maximumCacheAssets}`)
  if (release.assets.length < minimumCacheAssets) fail(`release carries ${release.assets.length} cache assets, under the declared minimum of ${minimumCacheAssets}`)
  const outputUrls = runtimeOutputUrls(outputDir)
  if (!sameSet(manifestUrls, outputUrls)) fail('manifest and production runtime output are not an exact set')
  if (!manifestUrls.has('/index.html')) fail('navigation entry is absent')
  for (const required of ['.wasm', '.css', '.js', '.ttf', '.folio']) if (!release.assets.some((asset) => asset.url.endsWith(required))) fail(`missing required runtime class ${required}`)
  if (!release.assets.some((asset) => /\/pdf\.worker-[A-Za-z0-9_-]+\.mjs$/.test(asset.url))) fail('missing local PDF.js worker runtime asset')
  if (release.assets.filter((asset) => asset.url.endsWith('.bcmap')).length < 4) fail('missing local PDF.js CMap runtime assets')
  if (!release.assets.some((asset) => /\/pdfjs-standard-fonts-[a-f0-9]{20}\/LiberationSans-Regular\.ttf$/.test(asset.url))) fail('missing local PDF.js standard-font runtime asset')
  if (release.id !== releaseIdentity(release.assets, release.workerRevision)) fail('release identity does not match canonical assets and worker revision')
  if (release.pageId !== pageIdentity(release.assets, release.workerRevision)) fail('page release identity does not match runtime assets and worker revision')
  if (!readFileSync(join(outputDir, 'index.html'), 'utf8').includes(`name="folio-page-release" content="${release.pageId}"`)) fail('page does not bind to its release identity')
  const indexHtml = readFileSync(join(outputDir, 'index.html'), 'utf8')
  const bootstrap = indexHtml.match(/<script id="folio-release-bootstrap" type="application\/json">([^<]+)<\/script>/)?.[1]
  if (!bootstrap) fail('page has no cached S1 release bootstrap')
  let bootS1
  try { bootS1 = JSON.parse(bootstrap).s1 } catch { fail('page S1 bootstrap is not JSON') }
  const sw = readFileSync(join(outputDir, 'sw.js'), 'utf8')
  const embedded = sw.match(/const RELEASE = (.+)\nconst CACHE_NAME/m)?.[1]
  if (!embedded || JSON.stringify(JSON.parse(embedded)) !== JSON.stringify(release)) fail('service worker and manifest release records differ')
  for (const required of ["const CACHE_NAME = 'folio-release-' + RELEASE.id", "credentials: 'omit'", 'url.origin === origin', 'paths.has(url.pathname)', 'RELEASE.pageId', 'windows.length === 0', 'offline asset integrity mismatch']) if (!sw.includes(required)) fail(`service worker lacks ${required}`)
  if (sw.includes('skipWaiting') || sw.includes('cache.addAll') || sw.includes('fetch(event.request)')) fail('service worker has unsafe activation or generic network fallback')
  const markerWrite = sw.indexOf("await cache.put(MARKER")
  const finalVerified = sw.indexOf("await progress('verified', activeAsset)")
  if (markerWrite < 0 || finalVerified < markerWrite) fail('emitted worker can report 100% before its complete marker')
  if (release.thaiDictionary?.delivery !== 'emitted-wasm-digest-witness' || !manifestUrls.has(release.thaiDictionary.wasmUrl) || !/^[a-f0-9]{64}$/.test(release.thaiDictionary.sha256)) fail('Thai dictionary containment is not declared against the emitted wasm')
  const s1 = release.s1
  const s1Ids = ['engine', 'latin-font', 'thai-font', 'cjk-font', 'thai-dictionary']
  if (!s1 || JSON.stringify(bootS1) !== JSON.stringify(s1) || s1.version !== 1 || s1.releaseId !== release.id || s1.pageId !== release.pageId || s1.unit !== 'MiB' || s1.decimals !== 2 || s1.assetCount !== release.assets.length || !Array.isArray(s1.cacheAssets) || !Array.isArray(s1.rows) || s1.rows.length !== s1Ids.length || s1.rows.map((row) => row.id).join(',') !== s1Ids.join(',')) fail('S1 payload metadata is incomplete or not exactly page/release bound')
  const semanticLabels = ['Engine', 'Latin font', 'Thai font', 'CJK font', 'Thai dictionary']
  if (s1.rows.map((row) => row.label).join(',') !== semanticLabels.join(',') || s1.rows.some((row) => /cloud|download|account|sync/i.test(row.label))) fail('S1 semantic labels contain delivery fiction')
  if (s1.cacheAssets.length !== release.assets.length || new Set(s1.cacheAssets.map((asset) => asset.assetUrl)).size !== release.assets.length) fail('S1 cache assets are incomplete')
  for (const asset of release.assets) {
    const cacheAsset = s1.cacheAssets.find((candidate) => candidate.assetUrl === asset.url)
    if (!cacheAsset || cacheAsset.bytes !== readFileSync(join(outputDir, asset.url.slice(1))).byteLength) fail(`S1 cache denominator is not the emitted release ${asset.url}`)
  }
  if (s1.cachedBytes !== s1.cacheAssets.reduce((total, asset) => total + asset.bytes, 0)) fail('S1 cache denominator is not all release assets')
  const cachedRows = s1.rows.filter((row) => row.delivery === 'cached-asset')
  if (cachedRows.length !== 4 || !cachedRows.every((row) => typeof row.assetUrl === 'string' && Number.isSafeInteger(row.bytes) && row.bytes > 0 && /^[a-f0-9]{64}$/.test(row.sha256))) fail('S1 cached rows are invalid')
  for (const row of cachedRows) {
    const asset = release.assets.find((candidate) => candidate.url === row.assetUrl)
    if (!asset || asset.sha256 !== row.sha256) fail(`S1 row is not bound to an emitted asset ${row.id}`)
    if (readFileSync(`${join(outputDir, row.assetUrl.slice(1))}.br`).byteLength !== row.bytes) fail(`S1 row size is not its emitted Brotli sidecar ${row.id}`)
  }
  const dictionaryRow = s1.rows[4]
  if (dictionaryRow.delivery !== 'embedded-in-engine' || dictionaryRow.assetUrl !== release.thaiDictionary.wasmUrl || dictionaryRow.bytes !== readFileSync(join(root, '..', 'folio-go', 'internal', 'text', 'data', 'thai_words.trie')).byteLength || dictionaryRow.sha256 !== release.thaiDictionary.sha256) fail('S1 Thai dictionary row is not a real embedded witness')
  if (release.s1VisibleBytes !== cachedRows.reduce((total, row) => total + row.bytes, 0)) fail('S1 visible payload total is not row arithmetic')
  const cjk = s1.rows.find((row) => row.id === 'cjk-font')
  if (!cjk || cjk.bytes !== Math.max(...cachedRows.filter((row) => row.id.endsWith('font')).map((row) => row.bytes))) fail('S1 CJK row is not the dominant font payload')
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
      const pdfjsCollection = /\/pdfjs-(?:cmaps|standard-fonts)-[a-f0-9]{20}\//.test(asset.url)
      if (!asset.url.startsWith(contract.immutableRuntime.urlPrefix) || (!/-[A-Za-z0-9_-]{8,}\.[A-Za-z0-9]+$/.test(asset.url) && !pdfjsCollection)) fail(`immutable URL is not content-addressed ${asset.url}`)
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
    assertEngineWasmIsTreeIndependent()
  }
  // The dev-server offline bypass is gated on `import.meta.env.DEV` so the
  // branch and its strings are eliminated from production bundles. Prove that
  // elimination against the emitted release rather than trusting the gate.
  for (const asset of release.assets) {
    if (!/\.(?:js|mjs|html|css)$/.test(asset.url)) continue
    const text = readFileSync(join(outputDir, asset.url.slice(1)), 'utf8')
    for (const marker of DEV_BYPASS_MARKERS) if (text.includes(marker)) fail(`development offline bypass shipped in ${asset.url}`)
  }
  if (!contract.immutableRuntime.cacheControl.includes('immutable') || contract.immutableRuntime.contentEncoding !== 'br') fail('immutable host policy is incomplete')
  if (!contract.updateableEntries.some((entry) => entry.url === '/index.html' && !entry.cacheControl.includes('immutable')) || !contract.updateableEntries.some((entry) => entry.url === '/sw.js' && !entry.cacheControl.includes('immutable'))) fail('updateable entries have immutable policy')
  return release
}

// DW-106. build-wasm.mjs guards ONE CAUSE — the absence of four VCS needles in the
// emitted wasm. The PROPERTY that matters is that two builds of one commit agree,
// and a guard keyed on a proxy rather than on its purpose is a defect the
// engineering lead recorded against its own ruling (D-8.4.30). So assert the
// property itself: build the engine twice, with the tree state DELIBERATELY
// DIFFERENT between the two runs, and require the digests to agree — reporting
// both, so a failure names what moved.
//
// ITS LIMIT, STATED RATHER THAN OVERCLAIMED: this would NOT have caught DW-105.
// It holds the checkout PATH fixed by construction, and DW-105 is path dependence
// — one commit built from three different paths gives three different binaries,
// each embedding 87 occurrences of its own absolute root. The value of this check
// is the NEXT tree-dependent input, not the one already measured.
//
// It runs only under --wasm-witness. `build:wasm` is a dependency of `typecheck`,
// `test` AND `build`, so a second engine build there would tax every designer
// gate; this is the home that runs once.
function assertEngineWasmIsTreeIndependent() {
  const first = join(tmpdir(), `folio-engine-tree-state-a-${process.pid}.wasm`)
  const second = join(tmpdir(), `folio-engine-tree-state-b-${process.pid}.wasm`)
  // An untracked file is enough: Go derives `vcs.modified` from `git status`.
  const stray = join(root, '..', 'folio-go', `.folio-tree-state-probe-${process.pid}`)
  try {
    buildEngineWasm(first, { stdio: 'pipe' })
    writeFileSync(stray, 'transient tree-state probe written by verify-offline-release.mjs\n')
    buildEngineWasm(second, { stdio: 'pipe' })
    const before = sha256(readFileSync(first))
    const after = sha256(readFileSync(second))
    if (before !== after) fail(`the engine wasm is a function of the working tree rather than of the source: two builds of this commit differing only by one stray untracked file digest ${before} and ${after}`)
  } finally {
    rmSync(stray, { force: true })
    rmSync(first, { force: true })
    rmSync(second, { force: true })
  }
}

// DW-107. Story 8.4g red-proved its VCS-stamp guard BY HAND and recorded the
// result in prose, so the evidence could never fail again. This is that proof,
// executed: drop `-buildvcs=false` from the one argv that declares it, build, and
// require the detector to report every setting Go then stamps in.
//
// It deliberately does NOT go through `redProof`: that harness mutates dist/ and
// re-runs verifyOfflineRelease, and the guard under proof lives in build-wasm.mjs.
// The shape is kept — mutate, observe a failure, hold it to the guard's OWN
// message, restore — and the probe builds to a temp file, so src/generated/ is
// never touched and there is nothing to restore by hand.
function proveVCSStampGuardDiscriminates() {
  if (!existsSync(join(root, '..', '.git'))) fail('red proof vcs-stamp-guard could not look: Go stamps VCS build info only inside a version-controlled tree, and this checkout has no .git — an all-clear here would be indistinguishable from a couldn\'t-look')
  const probe = join(tmpdir(), `folio-engine-vcs-probe-${process.pid}.wasm`)
  try {
    buildEngineWasm(probe, { flags: [], stdio: 'pipe' })
    let message
    try { assertNoVCSStamp(readFileSync(probe), 'engine wasm built without -buildvcs=false') } catch (error) { message = error.message }
    if (!message) fail('red proof vcs-stamp-guard escaped verification: the engine wasm built WITHOUT -buildvcs=false carries no VCS stamp, so the flag build-wasm.mjs depends on is buying nothing')
    for (const setting of ['vcs.revision=', 'vcs.time=', 'vcs.modified=', 'vcs=']) if (!message.includes(setting)) fail(`red proof vcs-stamp-guard failed for the wrong reason: expected the detector to report ${setting}, got: ${message}`)
  } finally { rmSync(probe, { force: true }) }
}

function rewriteRelease(outputDir, release) {
  release.pageId = pageIdentity(release.assets, release.workerRevision)
  release.id = releaseIdentity(release.assets, release.workerRevision)
  writeFileSync(join(outputDir, 'offline-release-manifest.json'), `${JSON.stringify(release, null, 2)}\n`)
  writeFileSync(join(outputDir, 'sw.js'), serviceWorkerSource(release))
}

function redProof(name, mutate, expected) {
  let restore
  try {
    restore = mutate(dist)
    let message
    try { verifyOfflineRelease(dist) } catch (error) { message = error.message }
    if (!message) fail(`red proof ${name} escaped verification`)
    // A mutation can trip several guards at once. Where a proof names the guard
    // it is proving, hold it to that guard rather than to any failure at all.
    if (expected && !message.includes(expected)) fail(`red proof ${name} failed for the wrong reason: ${message}`)
  } finally { restore?.() }
}

export function runRedProofs(baseline = verifyOfflineRelease()) {
  const readRelease = (outputDir) => JSON.parse(readFileSync(join(outputDir, 'offline-release-manifest.json'), 'utf8'))
  for (const extension of ['.wasm', '.ttf']) redProof(`stale-${extension.slice(1)}-byte`, (outputDir) => { const witness = baseline.assets.find((candidate) => candidate.immutable && candidate.url.endsWith(extension)); if (!witness) fail(`red proof has no ${extension} witness`); const file = join(outputDir, witness.url.slice(1)); const original = readFileSync(file); writeFileSync(file, Buffer.concat([original, Buffer.from([0])])); return () => writeFileSync(file, original) })
  for (const [name, extension] of [['missing-wasm-row', '.wasm'], ['missing-font-row', '.ttf']]) redProof(name, (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const worker = join(outputDir, 'sw.js'); const oldManifest = readFileSync(manifest); const oldWorker = readFileSync(worker); const release = readRelease(outputDir); release.assets = release.assets.filter((asset) => !asset.url.endsWith(extension)); rewriteRelease(outputDir, release); return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) } })
  for (const [name, matcher] of [['missing-pdfjs-worker', (url) => /\/pdf\.worker-[A-Za-z0-9_-]+\.mjs$/.test(url)], ['missing-pdfjs-cmap', (url) => url.endsWith('.bcmap')], ['missing-pdfjs-standard-font', (url) => /\/pdfjs-standard-fonts-[a-f0-9]{20}\/LiberationSans-Regular\.ttf$/.test(url)]]) redProof(name, (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const worker = join(outputDir, 'sw.js'); const oldManifest = readFileSync(manifest); const oldWorker = readFileSync(worker); const release = readRelease(outputDir); release.assets = release.assets.filter((asset) => !matcher(asset.url)); rewriteRelease(outputDir, release); return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) } })
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
  redProof('s1-total-mismatch', (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const original = readFileSync(manifest); const release = JSON.parse(original); release.s1.cachedBytes++; writeFileSync(manifest, JSON.stringify(release)); return () => writeFileSync(manifest, original) })
  redProof('s1-delivery-fiction', (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const original = readFileSync(manifest); const release = JSON.parse(original); release.s1.rows[4].delivery = 'cached-asset'; writeFileSync(manifest, JSON.stringify(release)); return () => writeFileSync(manifest, original) })
  // THE ENVELOPE, BOTH ENDS, PROVED BY MANUFACTURING A RELEASE OUTSIDE IT.
  // These follow the `missing-*-row` shape rather than the `s1-*` one: the bound
  // is a property of the asset POPULATION, so the only faithful mutation is a
  // release that really carries the wrong number of assets. `rewriteRelease`
  // recomputes `id` and `pageId` so the identity checks cannot trip first, and
  // `expected` holds each proof to the bound guard's own message rather than to
  // any failure at all.
  redProof('asset-count-over-bound', (outputDir) => {
    const manifest = join(outputDir, 'offline-release-manifest.json')
    const worker = join(outputDir, 'sw.js')
    const oldManifest = readFileSync(manifest)
    const oldWorker = readFileSync(worker)
    const release = readRelease(outputDir)
    const { maximumCacheAssets } = declaredCacheAssetBounds()
    while (release.assets.length <= maximumCacheAssets) release.assets.push({ url: `/assets/over-bound-${release.assets.length}-0123456789ab.js`, sha256: '0'.repeat(64), immutable: true })
    rewriteRelease(outputDir, release)
    return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) }
  }, 'over the declared maximum of')
  redProof('asset-count-under-bound', (outputDir) => {
    const manifest = join(outputDir, 'offline-release-manifest.json')
    const worker = join(outputDir, 'sw.js')
    const oldManifest = readFileSync(manifest)
    const oldWorker = readFileSync(worker)
    const release = readRelease(outputDir)
    const { minimumCacheAssets } = declaredCacheAssetBounds()
    release.assets = release.assets.slice(0, minimumCacheAssets - 1)
    rewriteRelease(outputDir, release)
    return () => { writeFileSync(manifest, oldManifest); writeFileSync(worker, oldWorker) }
  }, 'under the declared minimum of')
  redProof('s1-cloud-label', (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const original = readFileSync(manifest); const release = JSON.parse(original); release.s1.rows[0].label = 'Cloud download'; writeFileSync(manifest, JSON.stringify(release)); return () => writeFileSync(manifest, original) })
  redProof('s1-progress-denominator', (outputDir) => { const manifest = join(outputDir, 'offline-release-manifest.json'); const original = readFileSync(manifest); const release = JSON.parse(original); release.s1.cacheAssets.pop(); writeFileSync(manifest, JSON.stringify(release)); return () => writeFileSync(manifest, original) })
  redProof('s1-bootstrap-drift', (outputDir) => { const index = join(outputDir, 'index.html'); const original = readFileSync(index); writeFileSync(index, original.toString().replace('"releaseId":"', '"releaseId":"0')); return () => writeFileSync(index, original) })
  redProof('dev-bypass-shipped', (outputDir) => {
    const witness = baseline.assets.find((candidate) => candidate.url.startsWith('/assets/') && candidate.url.endsWith('.js'))
    if (!witness) fail('red proof has no script witness')
    const file = join(outputDir, witness.url.slice(1))
    const original = readFileSync(file)
    // Regenerating after the mutation restores every digest, identity, and S1
    // figure, so the bypass guard is the only one left that can trip.
    writeFileSync(file, Buffer.concat([original, Buffer.from(`\n// ${DEV_BYPASS_MARKERS[0]}\n`)]))
    generateOfflineRelease(outputDir)
    return () => { writeFileSync(file, original); generateOfflineRelease(outputDir) }
  }, 'development offline bypass shipped in')
  proveVCSStampGuardDiscriminates()
  redProof('worker-progress-before-marker', (outputDir) => { const worker = join(outputDir, 'sw.js'); const original = readFileSync(worker, 'utf8'); const moved = original.replace("    await cache.put(MARKER, new Response(RELEASE.id, { headers: { 'content-type': 'text/plain' } }))\n    await progress('verified', activeAsset)", "    await progress('verified', activeAsset)\n    await cache.put(MARKER, new Response(RELEASE.id, { headers: { 'content-type': 'text/plain' } }))"); if (moved === original) fail('red proof could not find final marker ordering'); writeFileSync(worker, moved); return () => writeFileSync(worker, original) })
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const redOnly = process.argv.includes('--red-only')
  const wasmWitness = process.argv.includes('--wasm-witness')
  const baseline = redOnly ? JSON.parse(readFileSync(join(dist, 'offline-release-manifest.json'), 'utf8')) : verifyOfflineRelease(dist, { wasmWitness })
  if (process.argv.includes('--red-proof') || redOnly) runRedProofs(baseline)
}
