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
//
// The message names the source it actually read. It used to hardcode
// `src/release-payload.ts` even when a caller injected a fixture string — which
// is how every failure case is driven — so a failure report could name a file
// that was never opened.
function readDeclaredConstant(source, name, label) {
  const matches = [...source.matchAll(new RegExp(String.raw`^const ${name} = (\d+)$`, 'gm'))]
  if (matches.length !== 1) throw new Error(`${label} does not declare \`${name}\` as a single live constant: found ${matches.length} line-anchored \`const ${name} = <digits>\` declarations, expected exactly 1${matches.length === 0 ? ' (a renamed, deleted, reformatted or commented-out constant reads as none)' : ' (a second live declaration makes the authority ambiguous)'}`)
  return Number(matches[0][1])
}

export function declaredCacheAssetBounds(source, label = source === undefined ? 'src/release-payload.ts' : 'the injected release-payload source') {
  const text = source === undefined ? readFileSync(releasePayloadSource, 'utf8') : source
  const minimumCacheAssets = readDeclaredConstant(text, 'minimumCacheAssets', label)
  const maximumCacheAssets = readDeclaredConstant(text, 'maximumCacheAssets', label)
  // BOTH NUMBERS CAN BE READABLE AND THE ENVELOPE STILL IMPOSSIBLE. An edit that
  // left the floor above the ceiling would make every release fail verification
  // with a message blaming the RELEASE for carrying the wrong number of assets,
  // when the fault is in this declaration. Say which it is, here, once.
  if (minimumCacheAssets > maximumCacheAssets) throw new Error(`${label} declares an inverted cache-asset envelope: \`minimumCacheAssets\` is ${minimumCacheAssets} and \`maximumCacheAssets\` is ${maximumCacheAssets}, so no release can satisfy both and the fault is in the declaration rather than in any release`)
  return { minimumCacheAssets, maximumCacheAssets }
}
