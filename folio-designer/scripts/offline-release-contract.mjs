import { createHash } from 'node:crypto'

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
