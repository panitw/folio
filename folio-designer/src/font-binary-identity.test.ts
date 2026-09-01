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
// This file performs the join and pins its result. Two claims:
//
//   1. THE DELIBERATE INTERVAL. Each of the three shipped files is reached by
//      exactly two family names, and both names resolve to the SAME file. That
//      is what makes "the canvas asks for the engine's names" a rename rather
//      than a rasterization change. It is an INTERVAL, not an invariant:
//      Story 8.4c puts real IBM Plex bytes behind the IBM Plex names, and this
//      is the guard that must be edited — deliberately, with the divergence
//      recorded — rather than discovered by a designer squinting at glyphs.
//
//   2. THE FACE THE ENGINE MEASURED. Every family named after a shipped face
//      is declared from bytes IDENTICAL to the bytes `folio-go/fonts/fonts.go`
//      embeds under that same face name. AD-17 makes the browser a rasterizer
//      only; a family called `Noto Sans` that is not the engine's `Noto Sans`
//      is worse than no family at all, because it fails silently.
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

/**
 * Source files whose family count is not exactly two, i.e. where the interval
 * has broken — AND every unresolved slot, which is not a source file at all.
 * A sentinel that grouped like a path would let two rules over one broken slot
 * form a well-formed "pair" that resolves to no bytes whatsoever.
 */
function filesOutsideTheInterval(generator: string): ReadonlyArray<string> {
  return Object.entries(familiesPerSourceFile(generator)).filter(([file, families]) => isSentinel(file) || families.length !== 2).map(([file]) => file)
}

const isChrome = (family: string) => (chromeFamilies as ReadonlyArray<string>).includes(family)

/**
 * Source files NOT reached by exactly one design-system family and one engine
 * face name — the strongest half of the interval, since a file reached by two
 * chrome names has the right COUNT with the interval gone.
 *
 * A NAMED HELPER RATHER THAN AN INLINE LOOP, deliberately: written inline in the
 * real assertion and re-expressed differently in the red-proof fixture, this
 * half had never been shown to redden through the code the real assertion runs.
 * Both now drive this one function. Each offender carries the families that
 * reached it, so a failure still prints WHAT reached the file, not just a count.
 */
function pairsOutsideTheInterval(generator: string): ReadonlyArray<string> {
  return Object.entries(familiesPerSourceFile(generator))
    .filter(([file, families]) => isSentinel(file) || families.filter(isChrome).length !== 1 || families.filter((family) => !isChrome(family)).length !== 1)
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
      'IBM Plex Mono': 'public/fonts/notosanssc/NotoSansSC-Regular.ttf',
      'IBM Plex Sans Thai': 'public/fonts/notosansthai/NotoSansThai-Regular.ttf',
      'Noto Sans': 'public/fonts/notosans/NotoSans-Regular.ttf',
      'Noto Sans Thai': 'public/fonts/notosansthai/NotoSansThai-Regular.ttf',
      'Noto Sans SC': 'public/fonts/notosanssc/NotoSansSC-Regular.ttf',
    })
  })

  // THE INTERVAL, PINNED. Three files, six names, two names per file — and the
  // two names over one file resolve to one and the same path. This is TRUE
  // ONLY UNTIL STORY 8.4c, which is the whole point of pinning it.
  it('keeps two family names — one chrome, one engine — over each single shipped file, an interval Story 8.4c ends', () => {
    const perFile = familiesPerSourceFile(generator)
    expect(Object.keys(perFile).length, 'the six @font-face rules must reach exactly three source files').toBe(3)
    expect(
      filesOutsideTheInterval(generator),
      'Every shipped font file is declared under exactly TWO family names: the design system\'s and the engine\'s own. '
      + 'A file listed here has gained or lost a rule. If this is STORY 8.4c putting real IBM Plex bytes behind the IBM '
      + 'Plex names, that story SPLITS each pair into two files and this assertion is what it must deliberately rewrite; '
      + 'if it is not 8.4c, a family name has silently changed which bytes the browser rasterizes with.',
    ).toEqual([])

    // EVERY GROUPING KEY IS A REAL FILE. `familySourcePaths` falls back to a
    // sentinel string when a rule names an `assets` slot that does not resolve,
    // and a sentinel groups like a path: two rules over one broken slot would
    // otherwise form a well-formed "pair" resolving to no bytes at all. The
    // helpers reject sentinels too; this states it where a failure can name one.
    expect(
      Object.keys(perFile).filter(isSentinel),
      'an @font-face rule names an `assets` slot that does not exist, so nothing resolves the bytes behind that family',
    ).toEqual([])

    // AND THE TWO NAMES ARE ONE OF EACH. A file reached by two chrome families,
    // or by two engine names, is the same count with the interval gone. Through
    // the named helper, so the red-proof below drives this exact code path.
    expect(
      pairsOutsideTheInterval(generator),
      'Every shipped font file must be reached by exactly one design-system family and one engine face name — the '
      + 'two-names-one-file interval Story 8.4c ends. A file listed here is reached by two names of the same kind, which '
      + 'is the right COUNT with the interval gone.',
    ).toEqual([])
  })

  // THE RED-PROOF FOR THE INTERVAL CHECKER. The real generator has, and must
  // keep having, exactly three well-formed pairs, so the corpus itself can
  // never exercise the failing direction. A checker that only ever passes has
  // not been shown to discriminate.
  it('reports a mismatch when a pair is repointed or a rule deleted', () => {
    const rule = (family: string, slot: string) => `@font-face { font-family: '${family}'; src: url('./runtime/\${assets.${slot}}') format('truetype'); font-display: swap; }`
    const assetsBlock = "  sans: fingerprint(join(designerRoot, 'public', 'fonts', 'notosans', 'NotoSans-Regular.ttf'), 'noto-sans.ttf'),\n"
      + "  sansThai: fingerprint(join(designerRoot, 'public', 'fonts', 'notosansthai', 'NotoSansThai-Regular.ttf'), 'noto-sans-thai.ttf'),\n"
    const fixture = (...rules: ReadonlyArray<string>) => `${assetsBlock}${rules.join('\\n')}`

    // The healthy shape, so the checker is shown to pass something.
    const healthy = fixture(rule('IBM Plex Sans', 'sans'), rule('Noto Sans', 'sans'), rule('IBM Plex Sans Thai', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesOutsideTheInterval(healthy)).toEqual([])
    expect(pairsOutsideTheInterval(healthy)).toEqual([])
    expect(familySourcePaths(healthy)['Noto Sans']).toBe('public/fonts/notosans/NotoSans-Regular.ttf')
    expect(familySourcePaths(healthy)['Noto Sans']).toBe(familySourcePaths(healthy)['IBM Plex Sans'])

    // A PAIR REPOINTED AT DIFFERING FILES — the mutation Story 8.4c will make
    // on purpose and anyone else would make by accident.
    const repointed = fixture(rule('IBM Plex Sans', 'sans'), rule('Noto Sans', 'sansThai'), rule('IBM Plex Sans Thai', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesOutsideTheInterval(repointed)).not.toEqual([])
    expect(familySourcePaths(repointed)['Noto Sans']).not.toBe(familySourcePaths(repointed)['IBM Plex Sans'])

    // A RULE DELETED — the "simplification" that would leave a family the
    // canvas asks for with no face behind it.
    const deleted = fixture(rule('IBM Plex Sans', 'sans'), rule('IBM Plex Sans Thai', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesOutsideTheInterval(deleted)).toEqual(['public/fonts/notosans/NotoSans-Regular.ttf'])

    // AND A FILE REACHED BY TWO NAMES OF THE SAME KIND — the right count with
    // the interval gone, which the count alone would wave through.
    const bothChrome = fixture(rule('IBM Plex Sans', 'sans'), rule('IBM Plex Mono', 'sans'), rule('IBM Plex Sans Thai', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(filesOutsideTheInterval(bothChrome)).toEqual([])
    expect(familiesPerSourceFile(bothChrome)['public/fonts/notosans/NotoSans-Regular.ttf'].filter((family) => !isChrome(family))).toEqual([])
    // THROUGH THE HELPER THE REAL ASSERTION RUNS, not a second expression of
    // the same idea: this is the code path that must redden in production.
    expect(pairsOutsideTheInterval(bothChrome)).toEqual(['public/fonts/notosans/NotoSans-Regular.ttf (reached by: IBM Plex Sans, IBM Plex Mono)'])

    // AND A RULE OVER A SLOT THAT DOES NOT RESOLVE. Two such rules group under
    // one sentinel string and would otherwise look like a well-formed pair —
    // right count, one chrome name, one engine name, and no bytes anywhere.
    const unresolvable = fixture(rule('IBM Plex Sans', 'nope'), rule('Noto Sans', 'nope'), rule('IBM Plex Sans Thai', 'sansThai'), rule('Noto Sans Thai', 'sansThai'))
    expect(Object.keys(familiesPerSourceFile(unresolvable)).filter(isSentinel)).toEqual(['<no assets slot named nope>'])
    expect(filesOutsideTheInterval(unresolvable)).toEqual(['<no assets slot named nope>'])
    expect(pairsOutsideTheInterval(unresolvable)).toEqual(['<no assets slot named nope> (reached by: IBM Plex Sans, Noto Sans)'])

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
    const engineNamed = Object.keys(declaredPaths).filter((family) => !(chromeFamilies as ReadonlyArray<string>).includes(family))
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
