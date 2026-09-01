import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// THE JOIN NOTHING IN THIS CODEBASE PERFORMS.
//
// `scripts/build-wasm.mjs` holds the two halves of the browser's font identity
// in two places that never meet. Its `assets` object maps OPAQUE SLOT NAMES
// (`sans`, `sansCjk`, `sansThai`) to source file paths and carries no family
// name at all; its `runtime-fonts.css` template maps FAMILY NAMES to those
// slots by hand, with no loop. Nothing composes the two, so nothing has ever
// been able to say WHICH FILE IS BEHIND A FAMILY NAME — and after Story 8.4b
// that question has a load-bearing answer: three files are each declared
// TWICE, once under the design system's family name and once under the
// engine's own face name, and the canvas asks for the engine's.
//
// This file performs the join and pins its result. Three claims:
//
//   1. THE DIVERGENCE, AND HOW FAR IT HAS GOT. Story 8.4b declared both
//      vocabularies over the SAME three files — a deliberate INTERVAL, in
//      which the IBM Plex names were IBM Plex in name only. Story 8.4c ends
//      it, one family at a time, and this file is where each split is
//      RECORDED rather than discovered by a designer squinting at glyphs. The
//      assertion below names the chrome families that still share a file with
//      an engine face name, so converting one is a one-line, deliberate edit
//      and forgetting one is a red test.
//
//   2. THE FACE THE ENGINE MEASURED. Every family named after a shipped face
//      is declared from bytes IDENTICAL to the bytes `folio-go/fonts/fonts.go`
//      embeds under that same face name. AD-17 makes the browser a rasterizer
//      only; a family called `Noto Sans` that is not the engine's `Noto Sans`
//      is worse than no family at all, because it fails silently.
//
//   3. THE BYTES ARE THE FACE THE NAME CLAIMS. A family name in a stylesheet
//      is an assertion about bytes, and until Story 8.4c nothing in this
//      repository could check it: `IBM Plex Mono` was Noto Sans SC, a CJK sans
//      with no monospacing, and every gate stayed green. The guard below opens
//      the declared file and reads its own `name` table, so the name over the
//      rule and the name inside the file have to agree.
//
// WHY THE GENERATOR SOURCE AND NOT ITS OUTPUT. `src/generated/runtime-fonts.css`
// and `src/generated/runtime/` are gitignored and only exist after `build:wasm`,
// so asserting against them would make this guard's strength depend on build
// order — and a missing file is the classic way a guard goes quietly vacuous.
// Every file read here is a tracked source, in either language.
const here = path.dirname(fileURLToPath(import.meta.url))
const designerRoot = path.join(here, '..')
const generatorPath = path.join(designerRoot, 'scripts', 'build-wasm.mjs')
const engineFontsDir = path.join(designerRoot, '..', 'folio-go', 'fonts')
const enginePath = path.join(engineFontsDir, 'fonts.go')

// The DESIGN SYSTEM's own three families — tokens.css's `--font-sans`,
// `--font-mono` and `--font-page`, through which every `--type-*` token
// resolves. They are spelled here because they are a fact about THIS
// repository's chrome, fixed by DESIGN.md and not derived from anything.
// Everything else the generator declares is claimed to be an ENGINE face name,
// and that claim is checked against `fonts.Shipped()` below rather than
// assumed. Note that this is emphatically NOT a family -> face mapping: no
// entry here says which engine face any of these corresponds to, and the pair
// structure below is derived from the generator's own slots instead.
const chromeFamilies = ['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai'] as const

// withoutComments strips line and block comments while leaving string and
// template literals intact, so the parses below answer to the CODE that emits
// the stylesheet rather than to prose that merely spells a rule.
//
// IT IS DUPLICATED FROM canvas-font-stack.test.ts (which in turn duplicates it
// from canvas-authority-contract.test.ts), DELIBERATELY. Importing it from
// another test file would register that file's whole suite a second time under
// this one, and hoisting it into a shared non-test module would put a test
// helper into `src/` — the very production corpus those suites scan. Each guard
// stays independently self-contained: this file must be able to redden on its
// own, without depending on another suite's helper staying correct.
//
// A character scanner rather than a regex because a regex cannot tell
// `// a comment` from the `//` inside a URL string, and getting that backwards
// makes the guard vacuous exactly where it matters.
//
// WHY THIS IS LOAD-BEARING HERE, measured: with the three engine-named rules
// commented out — left in the file as comment text — every test in this file
// stayed green while the emitted stylesheet dropped to three rules, which is
// the exact state that shipped the reported Thai overlap. Red-proof below.
function withoutComments(source: string): string {
  let out = ''
  let index = 0
  let quote: string | undefined
  while (index < source.length) {
    const char = source[index] as string
    if (quote !== undefined) {
      out += char
      if (char === '\\') { out += source[index + 1] ?? ''; index += 2; continue }
      if (char === quote) quote = undefined
      index++
      continue
    }
    if (char === '"' || char === '\'' || char === '`') { quote = char; out += char; index++; continue }
    if (char === '/' && source[index + 1] === '/') { while (index < source.length && source[index] !== '\n') index++; continue }
    if (char === '/' && source[index + 1] === '*') { index += 2; while (index < source.length && !(source[index] === '*' && source[index + 1] === '/')) index++; index += 2; continue }
    out += char
    index++
  }
  return out
}

/**
 * The shape `familySourcePaths` yields for a rule whose `assets` slot did not
 * resolve. It is NOT a file path, and must never be treated as one: two rules
 * naming the same unresolvable slot would otherwise group into a well-formed
 * "pair" and satisfy the interval guard while resolving to nothing at all.
 */
const sentinelPrefix = '<no assets slot named '
const isSentinel = (file: string) => file.startsWith(sentinelPrefix)

/**
 * The `assets` half: slot name -> source file path, relative to the designer
 * root, for every slot fingerprinted out of `public/fonts`. The other slots
 * (`wasm`, `wasmExec`, `starter`) are fingerprinted from build products rather
 * than from a committed path and are deliberately not matched.
 */
function slotSourcePaths(generator: string): Readonly<Record<string, string>> {
  const entries = [...generator.matchAll(/(\w+):\s*fingerprint\(join\(designerRoot,\s*((?:'[^']*'\s*,\s*)*'[^']*')\)\s*,/g)]
  return Object.fromEntries(entries.map((match) => [
    match[1],
    [...match[2].matchAll(/'([^']*)'/g)].map((segment) => segment[1]).join('/'),
  ]))
}

/**
 * The `@font-face` half: family name -> the `assets` slot its `src` interpolates.
 *
 * The spelling matched here is the generator's exact, deliberate one-line-per-rule
 * form. Drift in it does not weaken this guard quietly: the non-vacuity floors
 * below assert the parse found rules at all, and the exact map assertion
 * asserts it found ALL of them.
 */
function familySlots(generator: string): ReadonlyArray<readonly [string, string]> {
  return [...generator.matchAll(/@font-face \{ font-family: '([^']+)'; src: url\('\.\/runtime\/\$\{assets\.(\w+)\}'\) format\('truetype'\); font-display: swap; \}/g)]
    .map((match) => [match[1], match[2]] as const)
}

/** The join: family name -> the source file the browser will be handed for it. */
function familySourcePaths(generator: string): Readonly<Record<string, string>> {
  const slots = slotSourcePaths(generator)
  return Object.fromEntries(familySlots(generator).map(([family, slot]) => [family, slots[slot] ?? `${sentinelPrefix}${slot}>`]))
}

/**
 * The families that reach each SOURCE FILE, keyed by the file. The PAIR
 * STRUCTURE IS DERIVED, never listed: "two names over one file" is exactly a
 * file with two families, and a deleted rule, a repointed rule or a slot moved
 * onto another file all show up as a file with the wrong number. Listing the
 * pairs instead would be a second authority on which browser family
 * corresponds to which engine face — exactly the mapping table Story 8.4b's
 * verdict rejects by name.
 *
 * Grouped by the RESOLVED FILE rather than by the `assets` slot, deliberately:
 * families sharing a slot resolve to one path by construction, so a slot-level
 * check would be a tautology. The file is the layer the claim is about.
 */
function familiesPerSourceFile(generator: string): Readonly<Record<string, ReadonlyArray<string>>> {
  const grouped: Record<string, string[]> = {}
  for (const [family, file] of Object.entries(familySourcePaths(generator))) (grouped[file] ??= []).push(family)
  return grouped
}

const isChrome = (family: string) => (chromeFamilies as ReadonlyArray<string>).includes(family)

/**
 * Source files reached by MORE THAN ONE family name — the RESIDUE of the
 * two-names-one-file interval Story 8.4b pinned and Story 8.4c ends. Once every
 * design-system family has its own IBM Plex bytes this is empty, and every
 * entry in it before then is a family still waiting to be converted.
 *
 * A NAMED HELPER RATHER THAN AN INLINE LOOP, deliberately: the real assertion
 * and the red-proof fixture below drive this one function, so the code path
 * that must redden in production is the code path shown to redden. Each
 * offender carries the families that reached it, so a failure prints WHAT
 * shares the file rather than just a count.
 *
 * EVERY UNRESOLVED SLOT IS REPORTED TOO. `familySourcePaths` falls back to a
 * sentinel string when a rule names an `assets` slot that does not exist, and a
 * sentinel groups like a path: a lone rule over a broken slot would otherwise
 * pass as a happily-unshared file resolving to no bytes at all.
 */
function filesReachedByMoreThanOneFamily(generator: string): ReadonlyArray<string> {
  return Object.entries(familiesPerSourceFile(generator))
    .filter(([file, families]) => isSentinel(file) || families.length > 1)
    .map(([file, families]) => `${file} (reached by: ${families.join(', ')})`)
}

/** The face names `fonts.Shipped()` keys its FontSet by, in the order it writes them. */
function shippedFaceNames(fontsGo: string): ReadonlyArray<string> {
  const body = /func Shipped\(\) folio\.FontSet \{[\s\S]*?\n\}/.exec(fontsGo)?.[0]
  if (body === undefined) throw new Error(`no Shipped() function in ${enginePath}`)
  return [...body.matchAll(/"([^"]+)":\s*\w+,/g)].map((match) => match[1])
}

/** Face name -> the file `fonts.go` embeds for it, joining `Shipped()`'s map through the //go:embed directives. */
function shippedFacePaths(fontsGo: string): Readonly<Record<string, string>> {
  const embeds = Object.fromEntries([...fontsGo.matchAll(/\/\/go:embed\s+(\S+)\s*\nvar\s+(\w+)\s+\[\]byte/g)].map((match) => [match[2], match[1]]))
  const body = /func Shipped\(\) folio\.FontSet \{[\s\S]*?\n\}/.exec(fontsGo)?.[0] ?? ''
  return Object.fromEntries([...body.matchAll(/"([^"]+)":\s*(\w+),/g)].map((match) => [match[1], embeds[match[2]] ?? `<no //go:embed for ${match[2]}>`]))
}

const digest = (file: string) => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex')

// ---------------------------------------------------------------------------
// THE SMALLEST sfnt READ THAT ANSWERS "WHAT DOES THIS FILE SAY IT IS".
//
// WHY HERE AND NOT IN `lint`. `lint/internal/rules` already owns "a font binary
// is what it claims to be" (`looksLikeSfnt`) and already reaches designer paths
// elsewhere, so it was the obvious home. It was rejected on a measured ground:
// `lint/go.mod` requires only `golang.org/x/tools`, so the check would need a
// hand-rolled Go `name`-table parser or a NEW MODULE DEPENDENCY — in the very
// module whose own dependency graph `ScanLicenceGraph` audits. On this side a
// `name`-table read is a short DataView walk with no dependency at all, and it
// sits beside the parse of the generator that decides which file to read.
//
// NO FONT LIBRARY, DELIBERATELY. Everything below is offsets from the OpenType
// spec: the table directory (12-byte header, 16-byte records), `name` (format,
// count, stringOffset, then 12-byte records), `post` (isFixedPitch at +12), and
// `cmap` subtable formats 4 and 12. Adding a parser dependency to read four
// integers would put a new package in the designer's graph to check a fact
// about three committed files.
// ---------------------------------------------------------------------------

type SfntTable = { readonly offset: number; readonly length: number }

function fontView(file: string): DataView {
  const bytes = fs.readFileSync(file)
  return new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength)
}

/** The table directory: four-character tag -> where that table starts and how long it is. */
function sfntTables(view: DataView): Readonly<Record<string, SfntTable>> {
  const version = view.getUint32(0)
  // 0x00010000 is TrueType outlines; 'true' is the legacy Apple spelling. A
  // 'ttcf' collection or an 'OTTO' CFF font is neither of the two things this
  // repository commits, and reading one as if it were is how a guard goes
  // quietly wrong rather than loudly.
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

/**
 * One `name`-table string, preferring the Windows (platform 3) UTF-16BE record
 * and falling back to the Macintosh (platform 1) single-byte one, which is the
 * only other encoding these files carry. Returns undefined when the font
 * declares no such name at all — reported by the caller rather than defaulted,
 * because a missing name and a wrong name are different defects.
 */
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
    const bytes = Buffer.from(view.buffer.slice(view.byteOffset + storage + view.getUint16(record + 10), view.byteOffset + storage + view.getUint16(record + 10) + length))
    if ((platform === 3 || platform === 0) && length % 2 === 0) return bytes.swap16().toString('utf16le')
    singleByte ??= bytes.toString('latin1')
  }
  return singleByte
}

/**
 * The family the file declares FOR ITSELF: nameID 16 (typographic family) when
 * present, else nameID 1 (family). Both IBM Plex and Noto ship Regular-only
 * static instances that carry nameID 1 and no nameID 16, so the fallback is the
 * live path here — the preference is stated because a font that DOES carry
 * nameID 16 means it by it.
 */
function declaredFamilyOfFile(file: string): string {
  const view = fontView(file)
  const tables = sfntTables(view)
  return nameTableString(view, tables, 16) ?? nameTableString(view, tables, 1) ?? '<the file declares no family name>'
}

/** `post.isFixedPitch` — non-zero exactly when the face claims to be monospaced. */
function isFixedPitch(file: string): number {
  const view = fontView(file)
  const post = sfntTables(view)['post']
  if (post === undefined) throw new Error(`${file} has no post table`)
  return view.getUint32(post.offset + 12)
}

/** nameID 13, the licence description the face carries in its own bytes. */
function licenceDescriptionOfFile(file: string): string {
  const view = fontView(file)
  return nameTableString(view, sfntTables(view), 13) ?? '<the file declares no licence description>'
}

/**
 * THE CHECKER THE WHOLE STORY IS ABOUT: the families whose declared source file
 * does not, in its own `name` table, call itself by the name the stylesheet
 * declares it under. Empty means every checked family name is an assertion the
 * bytes agree with.
 *
 * Takes the generator TEXT rather than reading it, so the red-proof below can
 * drive this exact function with a fixture generator — an identity guard that
 * only ever passes has not been shown to discriminate.
 */
function familiesDeclaredFromForeignBytes(generator: string, families: ReadonlyArray<string>): ReadonlyArray<string> {
  const declared = familySourcePaths(generator)
  return families.flatMap((family) => {
    const relative = declared[family]
    if (relative === undefined) return [`${family}: no @font-face rule declares it`]
    if (isSentinel(relative)) return [`${family}: declared from ${relative}, so no bytes resolve at all`]
    const file = path.join(designerRoot, relative)
    if (!fs.existsSync(file)) return [`${family}: declared from ${relative}, which does not exist`]
    const actual = declaredFamilyOfFile(file)
    return actual === family ? [] : [`${family}: declared from ${relative}, whose name table says '${actual}'`]
  })
}

describe('the family names the browser is given are the files the engine measured with', () => {
  // THE GENERATOR IS PARSED WITHOUT ITS COMMENTS. Both halves of the join are
  // read out of source text by regex, and comment text is source text: a rule
  // that has been commented out reads exactly like a live one. Measured — see
  // the helper's own note — commenting out the three engine-named rules left
  // this whole file green over a stylesheet that had lost half its faces.
  const generator = withoutComments(fs.readFileSync(generatorPath, 'utf8'))
  const fontsGo = fs.readFileSync(enginePath, 'utf8')

  // NON-VACUITY ON BOTH PARSED HALVES, FIRST AND SEPARATELY. Each half is
  // parsed out of a source file by regex, and a regex that stops matching
  // yields an empty object over which every assertion below passes. These two
  // are what turn "the generator was reformatted" into a red test rather than
  // a silent one.
  it('reads both halves of the generator, and neither is empty', () => {
    const slots = slotSourcePaths(generator)
    expect(Object.keys(slots).length, `read no font asset slots out of ${generatorPath}`).toBeGreaterThanOrEqual(3)
    expect(familySlots(generator).length, `read no @font-face rules out of ${generatorPath}`).toBe(6)
    expect(shippedFaceNames(fontsGo).length, `read no face names out of Shipped() in ${enginePath}`).toBe(3)
  })

  // AND NEITHER HALF COUNTS COMMENTED-OUT TEXT. The floor of six above is a
  // text count, so without the strip a rule that had been commented out still
  // satisfied it — the measured mutation that left every guard here green while
  // the emitted stylesheet dropped to three rules.
  it('reads neither a commented-out @font-face rule nor a commented-out assets slot as live', () => {
    const emitted = "@font-face { font-family: 'Noto Sans'; src: url('./runtime/${assets.sans}') format('truetype'); font-display: swap; }"
    const slot = "  sans: fingerprint(join(designerRoot, 'public', 'fonts', 'notosans', 'NotoSans-Regular.ttf'), 'noto-sans.ttf'),"
    // The live direction first, so both parses are shown to find anything.
    expect(familySlots(withoutComments(emitted)).length).toBe(1)
    expect(Object.keys(slotSourcePaths(withoutComments(slot)))).toEqual(['sans'])
    // Then the two comment forms a generator can hide either half in.
    expect(familySlots(withoutComments(`// ${emitted}`)).length).toBe(0)
    expect(familySlots(withoutComments(`/* ${emitted} */`)).length).toBe(0)
    expect(Object.keys(slotSourcePaths(withoutComments(`// ${slot}`)))).toEqual([])
    // AND THE DEFECT ITSELF: unstripped, the commented-out rule counted.
    expect(familySlots(`// ${emitted}`).length).toBe(1)
  })

  // THE WHOLE MAP, EXACTLY. Not a containment and not a count: an exact
  // `toEqual` over the composed join, so an added rule, a removed rule or a
  // rule repointed at another slot all redden here, and the failure prints the
  // family and the file rather than a number.
  it('binds every declared family to the source file it is declared from', () => {
    expect(familySourcePaths(generator)).toEqual({
      'IBM Plex Sans': 'public/fonts/notosans/NotoSans-Regular.ttf',
      'IBM Plex Mono': 'public/fonts/ibmplexmono/IBMPlexMono-Regular.ttf',
      'IBM Plex Sans Thai': 'public/fonts/notosansthai/NotoSansThai-Regular.ttf',
      'Noto Sans': 'public/fonts/notosans/NotoSans-Regular.ttf',
      'Noto Sans Thai': 'public/fonts/notosansthai/NotoSansThai-Regular.ttf',
      'Noto Sans SC': 'public/fonts/notosanssc/NotoSansSC-Regular.ttf',
    })
  })

  // THE INTERVAL, AS FAR AS 8.4c HAS ENDED IT. Story 8.4b pinned three files
  // reached by two names each. Story 8.4c splits each pair by giving the
  // design system's family its own IBM Plex bytes, and this list is the
  // RESIDUE: the files still shared, named with the families that share them.
  // Converting a family is a one-line generator edit and a one-line edit here;
  // forgetting one, or repointing a family at bytes it does not name, is red.
  it('names exactly the chrome families that have not yet been given their own IBM Plex bytes', () => {
    const perFile = familiesPerSourceFile(generator)
    expect(Object.keys(perFile).length, 'the six @font-face rules must reach four source files at this point in Story 8.4c').toBe(4)

    // EVERY GROUPING KEY IS A REAL FILE. `familySourcePaths` falls back to a
    // sentinel string when a rule names an `assets` slot that does not
    // resolve, and a sentinel groups like a path — so a rule resolving to no
    // bytes whatsoever would otherwise read as an ordinary unshared file.
    expect(
      Object.keys(perFile).filter(isSentinel),
      'an @font-face rule names an `assets` slot that does not exist, so nothing resolves the bytes behind that family',
    ).toEqual([])

    expect(
      filesReachedByMoreThanOneFamily(generator),
      'These files are still reached by BOTH a design-system family name and an engine face name — the interval Story '
      + '8.4b pinned and Story 8.4c ends, family by family. `IBM Plex Mono` has already been split off onto real IBM '
      + 'Plex bytes; the two listed here are what 8.4c\'s second commit converts. If you are that commit, this list '
      + 'becomes empty. If you are not, a family name has silently changed which bytes the browser rasterizes with.',
    ).toEqual([
      'public/fonts/notosans/NotoSans-Regular.ttf (reached by: IBM Plex Sans, Noto Sans)',
      'public/fonts/notosansthai/NotoSansThai-Regular.ttf (reached by: IBM Plex Sans Thai, Noto Sans Thai)',
    ])
  })

  // THE RED-PROOF FOR THE SHARING CHECKER. The real generator has, and will
  // increasingly have, no offenders at all, so the corpus itself cannot
  // exercise the failing direction. A checker that only ever passes has not
  // been shown to discriminate.
  it('reports a shared file, a repointed rule and an unresolvable slot', () => {
    const rule = (family: string, slot: string) => `@font-face { font-family: '${family}'; src: url('./runtime/\${assets.${slot}}') format('truetype'); font-display: swap; }`
    const assetsBlock = "  sans: fingerprint(join(designerRoot, 'public', 'fonts', 'notosans', 'NotoSans-Regular.ttf'), 'noto-sans.ttf'),\n"
      + "  sansThai: fingerprint(join(designerRoot, 'public', 'fonts', 'notosansthai', 'NotoSansThai-Regular.ttf'), 'noto-sans-thai.ttf'),\n"
    const fixture = (...rules: ReadonlyArray<string>) => `${assetsBlock}${rules.join('\\n')}`

    // THE FULLY DIVERGED SHAPE, so the checker is shown to pass something: one
    // family per file, no sharing anywhere. This is what the real generator
    // looks like once 8.4c's second commit lands.
    const diverged = fixture(rule('Noto Sans', 'sans'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesReachedByMoreThanOneFamily(diverged)).toEqual([])
    expect(familySourcePaths(diverged)['Noto Sans']).toBe('public/fonts/notosans/NotoSans-Regular.ttf')

    // A FILE STILL SHARED — the state this story is in the middle of ending,
    // and the state a forgotten conversion leaves behind.
    const shared = fixture(rule('IBM Plex Sans', 'sans'), rule('Noto Sans', 'sans'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesReachedByMoreThanOneFamily(shared)).toEqual(['public/fonts/notosans/NotoSans-Regular.ttf (reached by: IBM Plex Sans, Noto Sans)'])

    // A PAIR REPOINTED AT DIFFERING FILES — the mutation that leaves a family
    // the canvas asks for rasterizing with someone else's bytes.
    const repointed = fixture(rule('Noto Sans', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesReachedByMoreThanOneFamily(repointed)).toEqual(['public/fonts/notosansthai/NotoSansThai-Regular.ttf (reached by: Noto Sans, Noto Sans Thai)'])
    expect(familySourcePaths(repointed)['Noto Sans']).toBe('public/fonts/notosansthai/NotoSansThai-Regular.ttf')

    // A RULE DELETED — the "simplification" that would leave a family the
    // canvas asks for with no face behind it at all.
    const deleted = fixture(rule('Noto Sans Thai', 'sansThai'))
    expect(Object.keys(familySourcePaths(deleted))).toEqual(['Noto Sans Thai'])

    // AND A RULE OVER A SLOT THAT DOES NOT RESOLVE. It reaches no bytes, and
    // without the sentinel check it would group like a perfectly good file.
    const unresolvable = fixture(rule('IBM Plex Sans', 'nope'), rule('Noto Sans Thai', 'sansThai'))
    expect(Object.keys(familiesPerSourceFile(unresolvable)).filter(isSentinel)).toEqual(['<no assets slot named nope>'])
    expect(filesReachedByMoreThanOneFamily(unresolvable)).toEqual(['<no assets slot named nope> (reached by: IBM Plex Sans)'])

    // AND THE PARSER ITSELF GOING BLIND, which must not read as health.
    expect(familySlots('writeFileSync(cssPath, `body { font-family: sans-serif }`)').length).toBe(0)
  })

  // THE FILE-SWAP TIE, AND THE POINT OF THE WHOLE STORY. A family called
  // `Noto Sans` is only useful because the bytes behind it are the bytes the
  // ENGINE measured `Noto Sans` with. Names are cheap; this compares files.
  //
  // SCOPED TO THE ENGINE-NAMED HALF ON PURPOSE. The IBM Plex families are the
  // DESIGN SYSTEM's vocabulary, and it is a deliberate temporary fact — not a
  // requirement — that they currently sit over the same Noto files. Story 8.4c
  // puts real IBM Plex bytes behind those names, at which point asserting
  // engine identity for them would be false. Asserting it for the engine-named
  // half is true now AND stays true across 8.4c, which is exactly why the tie
  // is drawn here.
  it('declares each engine face from bytes identical to the ones fonts.Shipped() embeds', () => {
    const declaredPaths = familySourcePaths(generator)
    const engineNamed = Object.keys(declaredPaths).filter((family) => !isChrome(family))
    const shipped = shippedFacePaths(fontsGo)

    // THE NAME SETS ARE EQUAL. Add a fourth face to `fonts.Shipped()`, or drop
    // one, and the generator no longer declares the set it claims to.
    expect([...engineNamed].sort()).toEqual(Object.keys(shipped).sort())
    expect(Object.keys(shipped).length).toBe(3)

    // AND SO ARE THE BYTES, face by face.
    for (const face of Object.keys(shipped)) {
      const browserFile = path.join(designerRoot, declaredPaths[face])
      const engineFile = path.join(engineFontsDir, shipped[face])
      expect(fs.existsSync(browserFile), `${browserFile} (declared for '${face}') must exist`).toBe(true)
      expect(fs.existsSync(engineFile), `${engineFile} (embedded for '${face}') must exist`).toBe(true)
      expect(
        digest(browserFile),
        `the browser declares '${face}' from ${declaredPaths[face]}, which must be byte-identical to the file folio-go embeds under that face name (${shipped[face]}) — AD-17 makes the browser a rasterizer only, so a same-named different face fails silently`,
      ).toBe(digest(engineFile))
    }
  })
})

// THE FIRST TEST IN THIS REPOSITORY THAT CAN TELL WHETHER THE BYTES BEHIND A
// FONT NAME ARE THE FONT THAT NAME CLAIMS.
//
// Everything above joins a family name to a FILE PATH. That is a claim about
// the generator, and it was true of the defect too: `IBM Plex Mono` resolved,
// perfectly consistently, to `noto-sans-cjk.ttf` — Noto Sans SC, a 10.6 MB
// Chinese sans with no monospacing at all, drawing every number, tab label and
// brand mark in the chrome. Nothing was red. The assertions below open the
// declared file and read what the FILE says about itself.
//
// WHAT THESE GATES CAN AND CANNOT PROVE. Vitest under jsdom applies no
// stylesheet and loads no font, and this repository has no gate that executes
// a real font load or rasterizes a glyph (`test:e2e:compile` is `tsc
// --noEmit`). So this proves the BINDING — which family name resolves to which
// bytes, and what those bytes declare themselves to be. It does not prove that
// a browser visibly draws the chrome in IBM Plex.
describe('the bytes behind a chrome family name are the face that name claims', () => {
  const generator = withoutComments(fs.readFileSync(generatorPath, 'utf8'))

  // SCOPED TO `IBM Plex Mono` FOR NOW. It is the family Story 8.4c converts
  // first, on its own, because it was the worst of the three and because one
  // binary is the cheapest end-to-end proof the IBM Plex pipeline works.
  // `IBM Plex Sans` and `IBM Plex Sans Thai` still sit over the engine's Noto
  // files, and asserting self-identity for them would be false until the
  // second commit lands — at which point this list becomes `chromeFamilies`.
  const convertedChromeFamilies = ['IBM Plex Mono'] as const

  it('declares each converted chrome family from a file whose own name table carries that family', () => {
    expect(
      familiesDeclaredFromForeignBytes(generator, convertedChromeFamilies),
      'A design-system family name is an assertion about bytes. Each family listed here is declared over a file that '
      + 'calls itself something else — which is exactly the shipped defect this guard exists to make observable, and '
      + 'which no gate in this repository could see before it.',
    ).toEqual([])
  })

  it('gives the mono family a genuinely monospaced face, and one that carries the OFL in its own bytes', () => {
    const file = path.join(designerRoot, familySourcePaths(generator)['IBM Plex Mono'])

    // THE DEFECT'S OWN SHAPE. `--font-mono` — and through it `--type-mono`,
    // `--type-mono-em`, `--type-numeric-lg`, `--type-brand`,
    // `--type-brand-load`, `--type-band-tab` and `--type-page-mono` — resolves
    // through this family. A face that is not fixed-pitch satisfies the name
    // check above and still draws every column of digits ragged.
    expect(isFixedPitch(file), `${file} is declared as 'IBM Plex Mono' but its post table does not claim fixed pitch`).not.toBe(0)

    // AD-26: the redistributed asset carries its own terms, in the bytes as
    // well as in the LICENSE*/NOTICE* beside them.
    expect(licenceDescriptionOfFile(file)).toContain('SIL Open Font License, Version 1.1')
  })

  // THE RED-PROOF. The corpus cannot exercise the failing direction — the
  // whole point of the story is that it no longer holds — so the checker is
  // driven with a fixture generator that reinstates the exact defect: the
  // family `IBM Plex Mono` bound back at the Noto CJK file.
  it('reports a mismatch when IBM Plex Mono is bound back at the Noto CJK file', () => {
    const rule = (family: string, slot: string) => `@font-face { font-family: '${family}'; src: url('./runtime/\${assets.${slot}}') format('truetype'); font-display: swap; }`
    const assetsBlock = "  sansCjk: fingerprint(join(designerRoot, 'public', 'fonts', 'notosanssc', 'NotoSansSC-Regular.ttf'), 'noto-sans-cjk.ttf'),\n"
      + "  mono: fingerprint(join(designerRoot, 'public', 'fonts', 'ibmplexmono', 'IBMPlexMono-Regular.ttf'), 'ibm-plex-mono.ttf'),\n"

    // The healthy direction first, through the same function, so the checker
    // is shown to pass something as well as to fail something.
    expect(familiesDeclaredFromForeignBytes(`${assetsBlock}${rule('IBM Plex Mono', 'mono')}`, ['IBM Plex Mono'])).toEqual([])

    // AND THE DEFECT, VERBATIM: the state this repository shipped in.
    expect(familiesDeclaredFromForeignBytes(`${assetsBlock}${rule('IBM Plex Mono', 'sansCjk')}`, ['IBM Plex Mono'])).toEqual([
      "IBM Plex Mono: declared from public/fonts/notosanssc/NotoSansSC-Regular.ttf, whose name table says 'Noto Sans SC'",
    ])
    // The same file, checked the second way the mono slot is checked: the CJK
    // sans is not monospaced, so the fixed-pitch assertion discriminates too.
    expect(isFixedPitch(path.join(designerRoot, 'public/fonts/notosanssc/NotoSansSC-Regular.ttf'))).toBe(0)

    // A rule over bytes that do not exist, and a family with no rule at all,
    // are reported rather than silently skipped — the two ways this checker
    // could go vacuous while returning an empty list.
    expect(familiesDeclaredFromForeignBytes(`${assetsBlock}${rule('IBM Plex Mono', 'nope')}`, ['IBM Plex Mono'])).toEqual([
      'IBM Plex Mono: declared from <no assets slot named nope>, so no bytes resolve at all',
    ])
    expect(familiesDeclaredFromForeignBytes(assetsBlock, ['IBM Plex Mono'])).toEqual(['IBM Plex Mono: no @font-face rule declares it'])
  })

  // AND THE READER ITSELF, SHOWN TO READ. Every assertion above is only worth
  // what the `name`-table walk is worth, and a reader that returned the same
  // string for everything would satisfy them all.
  it('reads a family name out of every committed face, and tells them apart', () => {
    expect(declaredFamilyOfFile(path.join(designerRoot, 'public/fonts/ibmplexmono/IBMPlexMono-Regular.ttf'))).toBe('IBM Plex Mono')
    expect(declaredFamilyOfFile(path.join(designerRoot, 'public/fonts/notosans/NotoSans-Regular.ttf'))).toBe('Noto Sans')
    expect(declaredFamilyOfFile(path.join(designerRoot, 'public/fonts/notosanssc/NotoSansSC-Regular.ttf'))).toBe('Noto Sans SC')
    expect(declaredFamilyOfFile(path.join(designerRoot, 'public/fonts/notosansthai/NotoSansThai-Regular.ttf'))).toBe('Noto Sans Thai')
  })
})
