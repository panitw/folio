// THE ONE sfnt `name`-TABLE READER, AND WHY IT IS PRODUCTION SOURCE.
//
// Story 16.1 needs nameID 0 — a face's own copyright — AT RUNTIME, over bytes
// fetched seconds earlier from a third party. Two hand-written `DataView` walks
// answering that question already existed in this repository and NEITHER could
// be imported by the browser: one is a vitest file (`font-catalogue.test.ts`),
// one is a build script (`scripts/build-wasm.mjs`). A third hand-copy would have
// been the first one whose divergence nobody would notice, because the two
// existing copies are checked against each other by the very test that holds one
// of them. So the walk is extracted HERE, once, and both existing callers are
// switched onto it.
//
// NO PARSER DEPENDENCY, AND THAT IS A STANDING DECISION rather than an omission.
// `src/font-catalogue.test.ts` records the ground: `package.json`'s
// `dependencies` are exactly `pdfjs-dist`, `react` and `react-dom`, the bundle
// sits under a measured payload budget, and reading four integers out of a table
// directory is a short walk. Story 16.1 adds no dependency.
//
// NO `Buffer`, DELIBERATELY. Both copies this replaces used `Buffer.swap16()`,
// which is Node-only. This module runs in the browser, in vitest and under plain
// `node` (`scripts/build-wasm.mjs` imports it directly — Node 24 strips the
// types), so the two encodings the `name` table actually uses are decoded by
// hand below.
//
// THIS MODULE ANSWERS A DIFFERENT QUESTION FROM GO'S READER, and the duplication
// is forced rather than chosen. `copyright` is one of `embedFontFamily`'s twelve
// wire fields, so Go cannot supply an input to itself; Go reads the name table
// again, from the same bytes, for its own question (Story 16.1b's licence tie).
// Two readers answering two questions is correct here — one reader would be a
// check over its own input.

export type SfntTable = Readonly<{ offset: number; length: number }>

/**
 * The table directory: every four-character tag the file declares, with the
 * offset and length it declares for it. No version check — some callers want
 * one and some do not, so it is `requireStaticTrueTypeTables` that refuses.
 */
export function sfntTableDirectory(view: DataView): Readonly<Record<string, SfntTable>> {
  const tables: Record<string, SfntTable> = {}
  const count = view.getUint16(4)
  for (let index = 0; index < count; index++) {
    const record = 12 + index * 16
    let tag = ''
    for (let byte = 0; byte < 4; byte++) tag += String.fromCharCode(view.getUint8(record + byte))
    tables[tag] = { offset: view.getUint32(record + 8), length: view.getUint32(record + 12) }
  }
  return tables
}

/**
 * The same directory, but only for a file whose sfnt version says it is a
 * static TrueType outline font — `0x00010000` or the older `true`. An
 * OpenType/CFF (`OTTO`) or WOFF wrapper is refused here rather than being read
 * as if its offsets meant the same thing.
 */
export function requireStaticTrueTypeTables(view: DataView): Readonly<Record<string, SfntTable>> {
  const version = view.getUint32(0)
  if (version !== 0x00010000 && version !== 0x74727565) throw new Error(`not a static TrueType sfnt: version 0x${version.toString(16).padStart(8, '0')}`)
  return sfntTableDirectory(view)
}

// The two encodings a `name` record actually uses, decoded without `Buffer`.
//
// UTF-16BE for the Windows (3) and Unicode (0) platforms, latin1 for the
// Macintosh (1) platform's Roman script. The odd-length guard is not decorative:
// a record claiming platform 3 with an odd byte count is malformed, and pairing
// its bytes anyway invents a character.
const utf16BE = (view: DataView, at: number, length: number): string => {
  let out = ''
  for (let byte = 0; byte + 1 < length; byte += 2) out += String.fromCharCode(view.getUint16(at + byte))
  return out
}
const latin1 = (view: DataView, at: number, length: number): string => {
  let out = ''
  for (let byte = 0; byte < length; byte++) out += String.fromCharCode(view.getUint8(at + byte))
  return out
}

/**
 * The first readable value the `name` table carries for `nameID`, preferring a
 * Unicode-platform record and falling back to a single-byte one.
 *
 * Returns `undefined` when the face carries no `name` table at all, or carries
 * one with no record for this id. A caller that requires the value says so
 * itself — silence and absence are different answers and this reader does not
 * collapse them.
 */
export function nameTableString(view: DataView, tables: Readonly<Record<string, SfntTable>>, nameID: number): string | undefined {
  const name = tables['name']
  if (name === undefined) return undefined
  const count = view.getUint16(name.offset + 2)
  const storage = name.offset + view.getUint16(name.offset + 4)
  let singleByte: string | undefined
  for (let index = 0; index < count; index++) {
    const record = name.offset + 6 + index * 12
    if (view.getUint16(record + 6) !== nameID) continue
    const platform = view.getUint16(record)
    const length = view.getUint16(record + 8)
    const offset = view.getUint16(record + 10)
    if ((platform === 3 || platform === 0) && length % 2 === 0) return utf16BE(view, storage + offset, length)
    singleByte ??= latin1(view, storage + offset, length)
  }
  return singleByte
}

/** A `DataView` over a face's bytes, whatever container they arrived in. */
export function fontView(bytes: ArrayBuffer | ArrayBufferView): DataView {
  return ArrayBuffer.isView(bytes) ? new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength) : new DataView(bytes)
}

/**
 * nameID 0 — THE ONE STATEMENT OF PROVENANCE THAT CANNOT BE EDITED FROM OUTSIDE
 * THE BINARY, and the value that reaches a `.folio` as `font.copyright`.
 *
 * It is read from the face's own bytes and never from a family's upstream
 * metadata, for the reason `font-catalogue.md` already gives: a value copied
 * from somewhere else is a second authority on the face's provenance, and the
 * first binary swap makes the document publish a statement its own bytes
 * contradict.
 *
 * ABSENT IS A REFUSAL, NOT A BLANK. The engine refuses to load a document that
 * embeds a face with an empty `copyright` (`internal/template/parse.go`'s
 * `requireEmbeddedFaceLicence`), so a pick that carried one would write a file
 * the product's own parser rejects. Blank counts as absent: `" "` states
 * exactly as much as `""`.
 */
export function faceCopyright(bytes: ArrayBuffer | ArrayBufferView): string {
  const view = fontView(bytes)
  const copyright = nameTableString(view, sfntTableDirectory(view), 0)?.trim()
  if (!copyright) throw new Error('this face declares no copyright in its own `name` table (nameID 0); a face embedded into a document must state whose it is, and the engine refuses to load a document that does not')
  return copyright
}
