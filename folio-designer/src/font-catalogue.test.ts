import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// STORY 8.5 — THE CATALOGUE, HELD TO ITS OWN RECORD.
//
// `font-catalogue.json` is the single declaration of which faces ship
// (Design Note 4): `scripts/build-wasm.mjs` loops over it, fingerprints each
// binary into `src/generated/runtime/`, and emits one `@font-face` rule per
// entry. That makes the manifest a claim about twenty-one committed binaries,
// and this file is where the claim meets the bytes.
//
// THREE CLAIMS, ONE PER ACCEPTANCE CRITERION:
//
//   AC1 — every face travels with the unmodified upstream `LICENSE*` and a
//   `NOTICE.md` recording the pinned upstream version, the upstream archive
//   digest, the committed digest, the byte size, the fetch date and the path
//   inside the archive; and `shasum -a 256` of the binary EQUALS the digest its
//   own NOTICE records. A provenance record nothing checks becomes a false
//   statement the first time a binary is swapped — and swapping a weight or an
//   italic keeps every name-table check green, which is exactly why the digest
//   is the tie.
//
//   AC6 — every face is a SINGLE UPRIGHT STATIC REGULAR, read out of its own
//   `name` and `OS/2` tables: no bold, no italic, no oblique, no variable axis.
//   Epic 11 (FR57) owns realize-vs-retire and the owner ruling has not been
//   made (D-000.7), so a weight matrix must not arrive here by accident. And
//   the generated `@font-face` rules declare no `font-weight` and no
//   `font-style`, so every other weight stays browser-synthesised from one face.
//
//   AC3 — at least twenty NEW families beyond the six already shipped, each
//   with bytes of its own.
//
// WHAT IS DELIBERATELY NOT HERE: anything about picking a family. Nothing in
// this story makes a catalogue face selectable — that is Story 8.6.

const here = path.dirname(fileURLToPath(import.meta.url))
const designerRoot = path.join(here, '..')
const cataloguePath = path.join(designerRoot, 'font-catalogue.json')
const generatorPath = path.join(designerRoot, 'scripts', 'build-wasm.mjs')
const fontsRoot = path.join(designerRoot, 'public', 'fonts')

/** The six families the generator declares by hand, which the catalogue must not collide with. */
const shippedFamilies = ['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai', 'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC']

interface CatalogueFace { id: string; directory: string; file: string; family: string; licence: string }

const catalogue: ReadonlyArray<CatalogueFace> = JSON.parse(fs.readFileSync(cataloguePath, 'utf8'))
const faceDirectory = (face: CatalogueFace) => path.join(fontsRoot, face.directory)
const faceFile = (face: CatalogueFace) => path.join(faceDirectory(face), face.file)
const digest = (file: string) => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex')

// ---------------------------------------------------------------------------
// THE SMALLEST sfnt READ THAT ANSWERS "IS THIS ONE UPRIGHT STATIC REGULAR".
// No font library, deliberately, on the ground `src/font-binary-identity.test.ts`
// already records: reading four integers out of a table directory is a short
// DataView walk, and a parser dependency added to check them would put a new
// package in the designer's graph.
// ---------------------------------------------------------------------------
type SfntTable = { readonly offset: number; readonly length: number }

function fontView(file: string): DataView {
  const bytes = fs.readFileSync(file)
  return new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
}

function sfntTables(view: DataView): Readonly<Record<string, SfntTable>> {
  const version = view.getUint32(0)
  if (version !== 0x00010000 && version !== 0x74727565) throw new Error(`not a static TrueType sfnt: version 0x${version.toString(16).padStart(8, '0')}`)
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

function nameTableString(view: DataView, tables: Readonly<Record<string, SfntTable>>, nameID: number): string | undefined {
  const name = tables['name']
  if (name === undefined) throw new Error('font has no name table')
  const count = view.getUint16(name.offset + 2)
  const storage = name.offset + view.getUint16(name.offset + 4)
  let singleByte: string | undefined
  for (let index = 0; index < count; index++) {
    const record = name.offset + 6 + index * 12
    if (view.getUint16(record + 6) !== nameID) continue
    const platform = view.getUint16(record)
    const length = view.getUint16(record + 8)
    const offset = view.getUint16(record + 10)
    const bytes = Buffer.from(view.buffer.slice(view.byteOffset + storage + offset, view.byteOffset + storage + offset + length))
    if ((platform === 3 || platform === 0) && length % 2 === 0) return bytes.swap16().toString('utf16le')
    singleByte ??= bytes.toString('latin1')
  }
  return singleByte
}

/** Everything AC6 asks a face about itself, read from its own bytes. */
function instanceOfFile(file: string) {
  const view = fontView(file)
  const tables = sfntTables(view)
  const os2 = tables['OS/2']
  const head = tables['head']
  const post = tables['post']
  if (os2 === undefined || head === undefined || post === undefined) throw new Error(`${file} is missing an OS/2, head or post table`)
  return {
    family: nameTableString(view, tables, 16) ?? nameTableString(view, tables, 1) ?? '<the file declares no family name>',
    subfamily: nameTableString(view, tables, 17) ?? nameTableString(view, tables, 2) ?? '<the file declares no subfamily name>',
    usWeightClass: view.getUint16(os2.offset + 4),
    // fsSelection bit 0 ITALIC, bit 5 BOLD, bit 6 REGULAR, bit 9 OBLIQUE.
    fsSelection: view.getUint16(os2.offset + 62),
    // head.macStyle bit 0 bold, bit 1 italic.
    macStyle: view.getUint16(head.offset + 44),
    italicAngle: view.getInt32(post.offset + 4) / 65536,
    variableTables: ['fvar', 'gvar', 'avar', 'HVAR', 'MVAR'].filter((tag) => tag in tables),
    outlineTables: ['glyf', 'CFF ', 'CFF2'].filter((tag) => tag in tables),
  }
}

/**
 * The digest a face's own `NOTICE.md` records for the file it ships, on exactly
 * the reader `src/font-binary-identity.test.ts` uses for the six shipped faces —
 * duplicated rather than imported, so each guard reddens on its own without
 * depending on the other suite's helper staying correct, and so importing it
 * does not register that suite a second time under this one.
 */
function recordedShippedDigest(noticeFile: string): string {
  const notice = fs.readFileSync(noticeFile, 'utf8')
  const rows = [...notice.matchAll(/^\|[^|\n]*sha256 of the SHIPPED[^|\n]*\|\s*`([0-9a-f]{64})`\s*\|/gm)]
  if (rows.length !== 1) throw new Error(`${noticeFile} must record exactly one 'sha256 of the SHIPPED …' table row carrying a 64-hex digest, and records ${rows.length}`)
  return rows[0][1]
}

/** The byte count a NOTICE records for the file it ships, as a number. */
function recordedShippedSize(noticeFile: string): number {
  const notice = fs.readFileSync(noticeFile, 'utf8')
  const rows = [...notice.matchAll(/^\| Size \| ([0-9,]+) bytes \|$/gm)]
  if (rows.length !== 1) throw new Error(`${noticeFile} must record exactly one '| Size | <n> bytes |' row, and records ${rows.length}`)
  return Number(rows[0][1].replaceAll(',', ''))
}

describe('the Story 8.5 catalogue ships the faces its manifest declares', () => {
  // NON-VACUITY FIRST. Every loop below is over `catalogue`, and an empty or
  // truncated manifest would satisfy all of them silently — the exact shape of
  // vacuous green this story's design notes are written against.
  it('declares at least twenty NEW families, none of them a family the six shipped rules already declare', () => {
    expect(catalogue.length, 'AC3 requires at least 20 new families beyond the 6 already shipped').toBeGreaterThanOrEqual(20)
    const families = catalogue.map((face) => face.family)
    expect(new Set(families).size, 'two catalogue entries declare the same family').toBe(families.length)
    expect(families.filter((family) => shippedFamilies.includes(family)), 'a catalogue face must not redeclare one of the six shipped families').toEqual([])
    const directories = catalogue.map((face) => face.directory)
    expect(new Set(directories).size, 'two catalogue entries share a directory, so two families would resolve to one file').toBe(directories.length)
    const ids = catalogue.map((face) => face.id)
    expect(new Set(ids).size).toBe(ids.length)
    // Every id is the runtime filename stem AND the token the release manifest
    // recognises a catalogue asset by, so its shape is asserted, not assumed.
    expect(ids.filter((id) => !/^[a-z0-9]+$/.test(id)), 'a catalogue id must be lower-case alphanumeric').toEqual([])
  })

  // AC1. THE PROVENANCE RECORD IS TRUE OF THE BYTES BESIDE IT.
  it('gives every catalogue face a LICENSE, a NOTICE, and a NOTICE that describes the binary it sits next to', () => {
    for (const face of catalogue) {
      const directory = faceDirectory(face)
      const file = faceFile(face)
      expect(fs.existsSync(file), `${face.id}: ${path.relative(designerRoot, file)} is declared in font-catalogue.json and is not committed`).toBe(true)

      // `manifest.ResolveAssets` looks for a file whose name STARTS WITH
      // "LICENSE" — uppercase, no other spelling. A face carrying `LICENCE.txt`
      // or `OFL.txt` would fail the licence gate at build time with a message
      // about a missing licence file, which is a confusing way to learn that a
      // filename was copied verbatim from upstream.
      const licences = fs.readdirSync(directory).filter((name) => name.startsWith('LICENSE'))
      expect(licences.length, `${face.id}: exactly one LICENSE* file is expected beside the binary; found ${JSON.stringify(licences)}`).toBe(1)

      const notice = path.join(directory, 'NOTICE.md')
      expect(fs.existsSync(notice), `${face.id}: no NOTICE.md, which manifest.ResolveAssets (AC25, AD-26) already requires`).toBe(true)
      const text = fs.readFileSync(notice, 'utf8')

      // The copyright line the licence gate will publish in lint/MANIFEST.md.
      expect(text.split('\n').some((line) => line.includes('Copyright')), `${face.id}: NOTICE.md carries no line containing "Copyright", so the licence gate fails the build`).toBe(true)

      // AC1's six recorded facts, each asserted by the row that carries it.
      expect(text, `${face.id}: NOTICE.md records no pinned upstream release`).toMatch(/^\| Upstream project \| .+release `[^`]+` \|$/m)
      expect(text, `${face.id}: NOTICE.md records no download URL`).toMatch(/^\| Download URL \| https:\/\/\S+ \|$/m)
      expect(text, `${face.id}: NOTICE.md records no path inside the archive`).toMatch(/^\| Path inside the archive \| `\S+` \|$/m)
      expect(text, `${face.id}: NOTICE.md records no fetch date`).toMatch(/^\| Fetched \| \d{4}-\d{2}-\d{2} \|$/m)
      expect(text, `${face.id}: NOTICE.md records no upstream archive digest`).toMatch(/^\| sha256 of the release archive \| `[0-9a-f]{64}` \([\d,]+ bytes\) \|$/m)
      expect(text, `${face.id}: NOTICE.md does not state its relation to the source`).toMatch(/copied unmodified, no derivation/)

      // AND THE TWO FACTS THAT CAN BE FALSE OF THE FILE ITSELF.
      expect(
        digest(file),
        `${face.id}: the binary's sha256 is not the digest its own NOTICE.md records. Either the binary was swapped without `
        + 'amending its provenance record — a different weight, a different style, a subset cut — or the record was amended '
        + 'without the binary. Both make the NOTICE a false statement about the bytes beside it.',
      ).toBe(recordedShippedDigest(notice))
      expect(recordedShippedSize(notice), `${face.id}: the recorded byte size is not the committed file's size`).toBe(fs.statSync(file).size)
    }
  })

  // AC6. ONE UPRIGHT STATIC REGULAR PER FAMILY, READ FROM THE BYTES.
  it('ships every catalogue face as a single upright static Regular, with no bold, italic, oblique or variable axis', () => {
    for (const face of catalogue) {
      const file = faceFile(face)
      const instance = instanceOfFile(file)
      const where = `${face.id} (${path.relative(designerRoot, file)})`

      expect(instance.family, `${where}: font-catalogue.json declares the family '${face.family}' and the file's own name table says '${instance.family}'. A family name is an assertion about bytes.`).toBe(face.family)
      expect(instance.subfamily, `${where}: is not a Regular instance`).toBe('Regular')
      expect(instance.usWeightClass, `${where}: OS/2.usWeightClass is not 400`).toBe(400)
      expect(instance.fsSelection & 0x40, `${where}: OS/2.fsSelection does not set the REGULAR bit`).toBe(0x40)
      expect(instance.fsSelection & 0x01, `${where}: OS/2.fsSelection sets the ITALIC bit`).toBe(0)
      expect(instance.fsSelection & 0x20, `${where}: OS/2.fsSelection sets the BOLD bit`).toBe(0)
      expect(instance.fsSelection & 0x200, `${where}: OS/2.fsSelection sets the OBLIQUE bit`).toBe(0)
      expect(instance.macStyle, `${where}: head.macStyle declares a bold or italic style`).toBe(0)
      expect(instance.italicAngle, `${where}: post.italicAngle is not upright`).toBe(0)
      expect(instance.variableTables, `${where}: carries variable-font tables. Epic 11 (FR57) owns realize-vs-retire and the owner ruling has not been made (D-000.7).`).toEqual([])
      // NFR7's operative choice: the glyf/TrueType static build, not CFF.
      expect(instance.outlineTables, `${where}: is not a glyf/TrueType static build`).toEqual(['glyf'])
      expect(path.extname(face.file), `${where}: the engine decodes only font/ttf and font/otf, and the emitted rule declares format('truetype')`).toBe('.ttf')
    }
  })

  // AND THE RULES THE GENERATOR EMITS DECLARE NO WEIGHT AND NO STYLE, so every
  // other weight the chrome asks for stays browser-synthesised from the one
  // face above. Read out of the generator SOURCE, which is tracked, rather than
  // out of `src/generated/runtime-fonts.css`, which is gitignored and only
  // exists after `build:wasm` — a guard whose strength depends on build order
  // is a guard that goes quietly vacuous.
  it('emits one catalogue rule per declared face, carrying no font-weight and no font-style', () => {
    const generator = fs.readFileSync(generatorPath, 'utf8')

    // THE LOOP IS THE MANIFEST'S, non-vacuously: the emitter must interpolate
    // the catalogue's own family and filename, so a template that had drifted
    // onto a literal list reds here rather than silently emitting six rules.
    const emitter = /catalogueFaces\.map\(\(face\) => `(@font-face \{[^`]*\})\\n`\)/.exec(generator)
    expect(emitter, `no catalogue @font-face emitter found in ${generatorPath}; the parse below would assert nothing`).not.toBeNull()
    const rule = (emitter as RegExpExecArray)[1]
    expect(rule).toContain('${face.family}')
    expect(rule).toContain('${face.filename}')
    expect(rule).toContain("format('truetype')")
    expect(rule, 'a font-weight descriptor would declare a weight matrix this story does not ship (AC6)').not.toContain('font-weight')
    expect(rule, 'a font-style descriptor would declare an italic or oblique this story does not ship (AC6)').not.toContain('font-style')

    // And the generator reads the manifest rather than a hardcoded list.
    expect(generator).toContain("readFileSync(join(designerRoot, 'font-catalogue.json'), 'utf8')")
  })
})
