// Go writes its build info into the emitted binary as newline-separated
// `build\t<key>=<value>` records. `go version -m` CANNOT read a wasm binary — it
// answers `unrecognized file format` — so grepping its output for `vcs` prints
// nothing BECAUSE THE TOOL FAILED, not because the stamp is absent. The stamp is
// therefore read from the raw bytes, which is the only reading that is a
// measurement rather than a hope.
//
// This lives apart from build-wasm.mjs because that script is all top-level side
// effects: importing it to test the detector would run a build.
const VCS_SETTINGS = [
  { setting: 'vcs.revision', needle: 'vcs.revision=' },
  { setting: 'vcs.time', needle: 'vcs.time=' },
  { setting: 'vcs.modified', needle: 'vcs.modified=' },
  // The bare `vcs` setting names the version-control system rather than the tree,
  // and `vcs=` is short enough to occur in arbitrary payload bytes, so it counts
  // only where it appears as a build-info record.
  { setting: 'vcs', needle: 'build\tvcs=' },
]

// A stamped value runs to the record's newline; stop at any non-printable byte so
// a needle that landed in binary payload yields a short, readable report instead
// of a screenful of noise.
const readValue = (bytes, start) => {
  let end = start
  while (end < bytes.length && end - start < 200 && bytes[end] >= 0x20 && bytes[end] <= 0x7e) end += 1
  return bytes.toString('latin1', start, end)
}

export function findVCSStampSettings(source) {
  // Every search below is `latin1`, a byte-exact 1:1 mapping, so a needle can
  // never straddle a multi-byte decode. A string source MUST be decoded the same
  // way: Buffer.from's default is UTF-8, which would silently disagree with the
  // searches and shift every offset past the first non-ASCII byte.
  const bytes = Buffer.isBuffer(source) ? source : typeof source === 'string' ? Buffer.from(source, 'latin1') : Buffer.from(source)
  const found = []
  for (const { setting, needle } of VCS_SETTINGS) {
    for (let index = bytes.indexOf(needle, 0, 'latin1'); index !== -1; index = bytes.indexOf(needle, index + needle.length, 'latin1')) {
      found.push({ setting, value: readValue(bytes, index + needle.length) })
    }
  }
  return found
}

export function assertNoVCSStamp(source, label) {
  const found = findVCSStampSettings(source)
  if (found.length === 0) return
  throw new Error(`${label} carries a Go build-info VCS stamp, so its bytes are a function of the working tree rather than of the source: ${found.map(({ setting, value }) => `${setting}=${value}`).join(', ')}. Build it with -buildvcs=false.`)
}
