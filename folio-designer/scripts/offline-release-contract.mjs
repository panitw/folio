import { createHash } from 'node:crypto'
import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

export const RELEASE_RUNTIME = 'v24.16.0'

export const sha256 = (bytes) => createHash('sha256').update(bytes).digest('hex')

export function normalizePublicPath(value) {
  return `/${value.replaceAll('\\', '/').replace(/^\/+/, '')}`
}

export function canonicalAssetRows(assets) {
  return [...assets]
    .map((asset) => `${asset.url}:${asset.sha256}`)
    .sort()
}

export function releaseIdentity(assets, workerRevision) {
  // index.html is the release bootstrap: it carries the identity it binds to.
  // Excluding that self-describing mutable shell avoids an impossible hash cycle;
  // the verifier still hashes and exact-set checks the emitted HTML itself.
  return sha256(Buffer.from([...canonicalAssetRows(assets.filter((asset) => asset.url !== '/index.html')), `worker:${workerRevision}`].join('\n')))
}

export function pageIdentity(assets, workerRevision) {
  return releaseIdentity(assets, workerRevision)
}

// THE CACHE-ASSET BOUND HAS ONE AUTHORITY: the `const` lines in
// src/release-payload.ts. The verifier must not re-type the number. `npm run
// build` is `build:wasm && tsc -b && vite build && build:offline &&
// verify:offline` — it never runs Vitest — so a second copy of the constant plus
// a tie test would leave a drifted build GREEN. The TypeScript module cannot be
// imported here either (this module needs node:crypto; tsconfig.app.json includes
// only src/; src/ imports nothing from scripts/), so the number is DERIVED from
// the declaration, on the idiom verify-offline-release.mjs already uses for the
// index.html bootstrap and for sw.js's `const RELEASE = …`.
const releasePayloadSource = join(dirname(fileURLToPath(import.meta.url)), '..', 'src', 'release-payload.ts')

// Line-anchored and multiline ON PURPOSE, and it must match EXACTLY ONCE.
// A false failure here stops a build loudly; a false success silently re-creates
// the very defect the bound exists to prevent, so every ambiguous case fails:
//   - `// const maximumCacheAssets = 64` is a commented-out copy, not a bound;
//     `^const` refuses it rather than counting dead code as live.
//   - two live declarations means first-match-wins would pick one and hide the
//     other, so two is a failure rather than a preference.
//   - zero means the constant was renamed, deleted or reformatted, and a reader
//     that answered "no bound" there would disable the guard in silence.
function readDeclaredConstant(source, name) {
  const matches = [...source.matchAll(new RegExp(String.raw`^const ${name} = (\d+)$`, 'gm'))]
  if (matches.length !== 1) throw new Error(`src/release-payload.ts does not declare \`${name}\` as a single live constant: found ${matches.length} line-anchored \`const ${name} = <digits>\` declarations, expected exactly 1${matches.length === 0 ? ' (a renamed, deleted, reformatted or commented-out constant reads as none)' : ' (a second live declaration makes the authority ambiguous)'}`)
  return Number(matches[0][1])
}

export function declaredCacheAssetBounds(source = readFileSync(releasePayloadSource, 'utf8')) {
  return { minimumCacheAssets: readDeclaredConstant(source, 'minimumCacheAssets'), maximumCacheAssets: readDeclaredConstant(source, 'maximumCacheAssets') }
}
