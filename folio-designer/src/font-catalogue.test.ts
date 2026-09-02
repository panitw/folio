import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
// THE GENERATED MODULE IS A SUBJECT HERE, not a convenience. Nothing observed
// it before, which is exactly how a build that gave 17 of 21 faces another
// project's licence text shipped green: `font-catalogue.json` was right, every
// binary was right, and the artifact BETWEEN them — the only thing the pick
// actually reads — was checked by nothing.
import { catalogueFaces as generatedFaces } from './generated/font-catalogue'
// THE SHARED sfnt `name`-TABLE READER (Story 16.1). This file used to write out
// its own `DataView` walk, byte-identical to a second copy in
// `scripts/build-wasm.mjs`; Story 16.1 needed a THIRD at runtime, over bytes
// fetched from a third party, and extracted the walk here instead of adding one.
import { nameTableString, requireStaticTrueTypeTables, type SfntTable } from './font-name-table'

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
//   STORY 8.6 ADDED A FOURTH: `scripts` — what each face covers, which the
//   designer proposes a fallback tail from. It is checked against that face's
//   OWN `cmap`, in BOTH directions: a script the manifest claims and the
//   binary cannot draw would give a document a chain with no fallback for
//   runes nothing in it covers, and a script the binary DOES cover that the
//   manifest omits would staple a redundant shipped face onto every chain
//   picking it. Both are silent; neither is visible in a rendered page until
//   somebody types in that script.

const here = path.dirname(fileURLToPath(import.meta.url))
const designerRoot = path.join(here, '..')
const cataloguePath = path.join(designerRoot, 'font-catalogue.json')
const generatorPath = path.join(designerRoot, 'scripts', 'build-wasm.mjs')
const fontsRoot = path.join(designerRoot, 'public', 'fonts')

/** The six families the generator declares by hand, which the catalogue must not collide with. */
const shippedFamilies = ['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai', 'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC']

interface CatalogueFace { id: string; directory: string; file: string; family: string; licence: string; scripts: ReadonlyArray<string> }

const catalogue: ReadonlyArray<CatalogueFace> = JSON.parse(fs.readFileSync(cataloguePath, 'utf8'))
const faceDirectory = (face: CatalogueFace) => path.join(fontsRoot, face.directory)
const faceFile = (face: CatalogueFace) => path.join(faceDirectory(face), face.file)
const digest = (file: string) => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex')

// ---------------------------------------------------------------------------
// THE SMALLEST sfnt READ THAT ANSWERS "IS THIS ONE UPRIGHT STATIC REGULAR".
// No font library, deliberately, on the ground `src/font-binary-identity.test.ts`
// already records: reading four integers out of a table directory is a short
// DataView walk, and a parser dependency added to check them would put a new
// package in the designer's graph. Story 16.1 moved the `name`-table half of
// that walk into `src/font-name-table.ts` and left the OS/2, head and post reads
// here, where they are this file's own question.
//
// ⚠ ONE READER, AND THE COST IS STATED. The comment that used to sit on
// `copyright` below claimed the comparison was "between two independent readers
// rather than one reader agreeing with itself". After the extraction that is no
// longer true of the browser side, and the claim is corrected rather than left
// standing: the generated catalogue and this test now read nameID 0 through the
// SAME module. What was bought is that the runtime reader — the one that decides
// what `font.copyright` a fetched face publishes — is the reader these 21
// committed faces exercise on every run, instead of a third hand-copy nobody
// checks. The independence that remains is Go's: `internal/fontset` walks the
// same table again, from the bytes, for its own different question.
//
// THE SWITCH WAS WITNESSED, not assumed. This file's assertions were run against
// the generated catalogue emitted by the OLD hand-written reader and again by the
// shared one, and all 21 `copyright` values were byte-identical.
// ---------------------------------------------------------------------------

function fontView(file: string): DataView {
  const bytes = fs.readFileSync(file)
  return new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
}

const sfntTables = (view: DataView): Readonly<Record<string, SfntTable>> => requireStaticTrueTypeTables(view)

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
    // nameID 13 is the LICENCE DESCRIPTION the face carries in its own bytes.
    licenceDescription: nameTableString(view, tables, 13) ?? '<the file declares no licence description>',
    // nameID 0 is the COPYRIGHT, and it is the value the generated catalogue
    // publishes as `copyright`. Read through the shared reader — see the note
    // above for what that costs and what it bought.
    copyright: nameTableString(view, tables, 0)?.trim(),
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

/**
 * THE LICENCE A FACE DECLARES IN ITS OWN `name` TABLE, per SPDX identifier.
 *
 * WHY THIS EXISTS, AND WHY IT IS THE `licence` FIELD'S FIRST CONSUMER.
 * `src/font-binary-identity.test.ts` already holds each chrome face's nameID 13
 * to the SIL OFL, on the stated ground that a redistributed asset's terms
 * travel in its `name` table as well as in the `LICENSE*`/`NOTICE*` beside it.
 * This file adopted that suite's DIGEST tie and not its LICENCE tie, and the
 * gap is not theoretical: swap a face's binary and its NOTICE together — same
 * family, different terms — and the digest check, the name check and the
 * instance checks all stay green while `lint/MANIFEST.md` publishes a licence
 * the binary itself contradicts.
 *
 * AND IT IS WHAT MAKES `font-catalogue.json`'s `licence` FIELD LOAD-BEARING.
 * Until this table existed nothing in the designer read that field at all: it
 * restated a fact already held in three other places (the NOTICE, the manifest
 * row, `pinnedCensus`) and could disagree with all three in silence. Keyed off
 * it, the field now has exactly one job and reds when it is wrong.
 *
 * MEASURED over all 21 committed faces before being written, not assumed: the
 * 19 OFL-1.1 faces all carry the SIL sentence in nameID 13 — including
 * `cascadiacode` and `cascadiamono`, whose description OPENS "Microsoft
 * supplied font..." and carries the OFL sentence further in, which is why this
 * is a substring match rather than a prefix or an equality — and both
 * Ubuntu-font-1.0 faces read "Licensed under the Ubuntu Font Licence 1.0."
 *
 * A CLOSED TABLE, DELIBERATELY: an id with no entry here fails rather than
 * skipping, so admitting a new licence to the catalogue is a decision somebody
 * makes here rather than a silent hole. It is NOT the licence gate's allowlist
 * and must not be treated as one — `lint` owns admission (D-000.11); this only
 * asks whether the bytes agree with the label already admitted.
 */
const licenceSignatures: Readonly<Record<string, RegExp>> = {
  'OFL-1.1': /SIL Open Font License/i,
  'Ubuntu-font-1.0': /Ubuntu Font Licence/i,
}

/**
 * THE UNICODE `cmap`, AS A SET OF THE CODEPOINTS THE FACE ACTUALLY MAPS.
 *
 * Formats 4 and 12 only, and that is measured rather than assumed: all 21
 * committed faces carry one or the other under a (3,1), (3,10) or (0,x)
 * subtable. A face carrying neither throws here instead of being scored zero —
 * a coverage check that silently reads no subtable would report every script
 * uncovered and pass nothing, which is the vacuous-green shape this file's
 * other guards are written against.
 *
 * Format 12 groups are bounded per group, because a CJK face's cmap is
 * hundreds of thousands of codepoints and this test only ever asks about a
 * handful of them; no committed catalogue face is CJK today, and the bound is
 * what keeps that from becoming a minutes-long test if one ever is.
 */
function cmapCoverage(file: string): ReadonlySet<number> {
  const view = fontView(file)
  const tables = sfntTables(view)
  const cmap = tables['cmap']
  if (cmap === undefined) throw new Error(`${file} has no cmap table`)
  const subtables = view.getUint16(cmap.offset + 2)
  let chosen = -1
  for (let index = 0; index < subtables; index++) {
    const record = cmap.offset + 4 + index * 8
    const platform = view.getUint16(record)
    const encoding = view.getUint16(record + 2)
    if (!((platform === 3 && (encoding === 1 || encoding === 10)) || platform === 0)) continue
    const subtable = cmap.offset + view.getUint32(record + 4)
    const format = view.getUint16(subtable)
    if (format === 4 || format === 12) chosen = subtable
  }
  if (chosen < 0) throw new Error(`${file} carries no Unicode cmap subtable in format 4 or 12`)
  const covered = new Set<number>()
  if (view.getUint16(chosen) === 4) {
    const segmentBytes = view.getUint16(chosen + 6)
    const endOffset = chosen + 14
    const startOffset = endOffset + segmentBytes + 2
    const deltaOffset = startOffset + segmentBytes
    const rangeOffset = deltaOffset + segmentBytes
    for (let segment = 0; segment < segmentBytes / 2; segment++) {
      const start = view.getUint16(startOffset + segment * 2)
      const end = view.getUint16(endOffset + segment * 2)
      if (start === 0xffff) continue
      const delta = view.getInt16(deltaOffset + segment * 2)
      const range = view.getUint16(rangeOffset + segment * 2)
      for (let codepoint = start; codepoint <= end; codepoint++) {
        let glyph: number
        if (range === 0) glyph = (codepoint + delta) & 0xffff
        else {
          const at = rangeOffset + segment * 2 + range + (codepoint - start) * 2
          if (at + 1 >= view.byteLength) continue
          glyph = view.getUint16(at)
          if (glyph !== 0) glyph = (glyph + delta) & 0xffff
        }
        // GLYPH 0 IS .notdef — a mapping to it is the absence of a mapping,
        // and counting it would score every face as covering everything.
        if (glyph !== 0) covered.add(codepoint)
      }
    }
  } else {
    const groups = view.getUint32(chosen + 12)
    for (let group = 0; group < groups; group++) {
      const record = chosen + 16 + group * 12
      const start = view.getUint32(record)
      const end = view.getUint32(record + 4)
      for (let codepoint = start; codepoint <= end && codepoint - start < 70000; codepoint++) covered.add(codepoint)
    }
  }
  return covered
}

/**
 * THE PROBE PER SCRIPT: codepoints a face claiming that script must ALL map,
 * and a face not claiming it must map NONE of.
 *
 * They are ordinary letters rather than rarities on purpose. The question is
 * "can this face draw text in this script at all", not "is its coverage
 * complete" — a partial Latin face is still the right first entry in a chain,
 * whereas one that maps no Latin letter at all must not be. Each probe spans
 * more than one block of its script so a face carrying, say, only ASCII digits
 * does not pass as Latin.
 *
 * MEASURED over all 21 committed faces before being written: 19 map every
 * Latin probe and no Thai one; notosansthailooped and notoserifthai map every
 * Thai probe and no Latin one. No committed face is ambiguous under it, and no
 * committed face is CJK.
 */
const scriptProbes: Readonly<Record<string, ReadonlyArray<number>>> = {
  // A, Z, a, z, 0, 9 — Basic Latin letters and digits.
  latin: [0x41, 0x5a, 0x61, 0x7a, 0x30, 0x39],
  // ko kai, so suea, a vowel sign and a tone mark: consonants, vowel and tone.
  thai: [0x0e01, 0x0e2a, 0x0e30, 0x0e48],
  // Four common Han ideographs.
  cjk: [0x4e00, 0x4e8c, 0x6c34, 0x9fa5],
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

      // AND THE TERMS THE BINARY ITSELF DECLARES (nameID 13), which is the one
      // statement of a face's licence that cannot be edited from outside the
      // binary. `font-catalogue.json`'s `licence` field is the key, so the
      // field is load-bearing rather than decorative.
      const signature = licenceSignatures[face.licence]
      expect(signature, `${face.id}: font-catalogue.json declares the licence '${face.licence}', which no entry in licenceSignatures recognises. Admitting a licence to the catalogue is a decision to record here, not a hole to fall through.`).toBeDefined()
      expect(
        instanceOfFile(file).licenceDescription,
        `${face.id}: font-catalogue.json declares '${face.licence}' and the binary's own name table (nameID 13) does not say so. `
        + 'A redistributed asset carries its terms in its name table as well as in the LICENSE*/NOTICE* beside it, and swapping '
        + 'a binary and its NOTICE together — same family, different terms — passes every other check in this file while '
        + 'lint/MANIFEST.md publishes a licence the bytes contradict.',
      ).toMatch(signature as RegExp)

      // The three records must also agree with each other: the NOTICE names the
      // same identifier the manifest declares.
      expect(text, `${face.id}: NOTICE.md does not name the SPDX identifier font-catalogue.json declares`).toContain(`\`${face.licence}\``)

      // STORY 8.6 — AND THE GENERATED MODULE PUBLISHES THIS FACE'S OWN TERMS.
      //
      // The designer sends `licenceText` and `copyright` with every pick, and
      // the engine refuses to load a document that embeds a face without them,
      // so these two strings ARE the terms a `.folio` travels under. They are
      // asserted per face, against this face's own directory and this face's
      // own bytes, because the failure they exist to catch is not "the field
      // is empty" — it is "the field is FULL, and it belongs to another
      // project".
      const generated = generatedFaces.find((emitted) => emitted.id === face.id)
      expect(generated, `${face.id}: font-catalogue.json declares it and src/generated/font-catalogue.ts emits no row for it, so the pick could not embed it`).toBeDefined()

      // (a) THE LICENCE TEXT IS THE ONE COMMITTED BESIDE THIS BINARY.
      // NOT keyed by SPDX identifier: the SIL OFL carries a per-project
      // preamble — a copyright line and a Reserved Font Name — so two OFL-1.1
      // faces ship two DIFFERENT texts, and an identifier classifies terms
      // rather than standing in for them.
      const licenceFile = path.join(directory, licences[0])
      expect(
        generated?.licenceText,
        `${face.id}: the generated catalogue publishes a licence text that is not the one committed beside this binary (${path.relative(designerRoot, licenceFile)}). Every document embedding this face would travel stating another project's terms.`,
      ).toBe(fs.readFileSync(licenceFile, 'utf8').trimEnd())
      expect(generated?.licence, `${face.id}: the generated catalogue and the manifest disagree on the SPDX identifier`).toBe(face.licence)

      // (b) THE COPYRIGHT IS THIS BINARY'S OWN nameID 0.
      const declaredCopyright = instanceOfFile(file).copyright
      expect(declaredCopyright, `${face.id}: the binary declares no copyright in its own name table (nameID 0), so there is nothing for the document to record`).toBeTruthy()
      expect(
        generated?.copyright,
        `${face.id}: the generated catalogue publishes a copyright that is not the one this face's own name table declares. nameID 0 is the one statement of provenance that cannot be edited from outside the binary.`,
      ).toBe(declaredCopyright)
    }
  })

  // NON-VACUITY FOR THE TWO ASSERTIONS ABOVE. Both are inside a loop over
  // `catalogue` and both reach the generated module through `.find()`, so a
  // module that emitted nothing at all would make them assert on `undefined`
  // rows — caught by the `toBeDefined()` above, but only face by face. This
  // states the population once, at the top level, so a truncated or stale
  // generated module reds here with one clear sentence.
  it('emits exactly one generated row per declared catalogue face', () => {
    expect(generatedFaces).toHaveLength(catalogue.length)
    expect(generatedFaces.map((face) => face.id).sort()).toEqual(catalogue.map((face) => face.id).sort())
    // And every row carries a URL into the fingerprinted runtime directory —
    // the bytes the pick reads. A row with no URL is a family the author can
    // see and cannot embed.
    expect(generatedFaces.filter((face) => typeof face.url !== 'string' || face.url === '')).toEqual([])
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

  // STORY 8.6. THE DECLARED COVERAGE AGREES WITH THE BINARY'S OWN cmap.
  it('declares, for every catalogue face, exactly the scripts its own cmap can draw', () => {
    // NON-VACUITY: the probe table is what every assertion below reads, and a
    // manifest declaring a script it does not name would be scored against
    // nothing at all.
    const vocabulary = Object.keys(scriptProbes)
    expect(vocabulary.length, 'the probe table is empty, so every assertion below is vacuous').toBeGreaterThan(0)

    for (const face of catalogue) {
      const where = `${face.id} (${path.relative(designerRoot, faceFile(face))})`
      expect(Array.isArray(face.scripts) && face.scripts.length > 0, `${where}: font-catalogue.json declares no scripts; the designer proposes a fallback tail from this list, and a face claiming nothing would be given a fallback for every script including its own`).toBe(true)
      expect(face.scripts.filter((script) => !vocabulary.includes(script)), `${where}: declares a script outside the closed vocabulary ${vocabulary.join(', ')}. An unrecognised script proposes no fallback for itself and the chain draws tofu.`).toEqual([])

      const covered = cmapCoverage(faceFile(face))
      expect(covered.size, `${where}: its cmap maps no codepoint at all, so the coverage comparison below would assert nothing`).toBeGreaterThan(0)

      for (const script of vocabulary) {
        const probes = scriptProbes[script] as ReadonlyArray<number>
        const mapped = probes.filter((codepoint) => covered.has(codepoint))
        const hex = (list: ReadonlyArray<number>) => list.map((codepoint) => `U+${codepoint.toString(16).toUpperCase().padStart(4, '0')}`).join(', ')
        if (face.scripts.includes(script)) {
          // CLAIMED: the binary must draw all of them. A face claiming a
          // script it cannot draw gets NO shipped fallback for that script,
          // so the document renders tofu where it promised coverage.
          expect(mapped, `${where}: font-catalogue.json claims the script '${script}' and the face's own cmap maps only ${hex(mapped)} of ${hex(probes)}. A claimed script gets no fallback entry, so the chain would draw tofu.`).toEqual(probes)
        } else {
          // NOT CLAIMED: the binary must draw none of them. A face that does
          // cover a script it does not claim gets a redundant shipped face
          // stapled behind it in every chain that picks it.
          expect(mapped, `${where}: font-catalogue.json does not claim the script '${script}' and the face's own cmap maps ${hex(mapped)}. Every chain picking this family would carry a redundant shipped fallback for a script it already covers.`).toEqual([])
        }
      }
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
