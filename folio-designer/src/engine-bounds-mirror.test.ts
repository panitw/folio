import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// D-7.4.5 / DW-25, AS WIDENED BY STORY 7.5.
//
// THE STANDING OBLIGATION, in its current form: ANY invariant duplicated
// across the Go/TypeScript boundary moves in ONE commit, with a test that
// reads both sides. It used to read "the size caps move together", and that
// wording was the defect — DW-25 audited and closed the four SIZE CAPS, and
// band containment, a different invariant that merely lives in the same file,
// was left untied and drifted anyway. An audit closes only what it measured.
//
// `engine-protocol.ts` hand-copies the canvas projection bounds
// `folio-go/page_setup.go` declares, and the band-containment predicate
// `folio-go/component_commands.go` declares. There is no shared source and no
// codegen, so the only thing standing between the files is this test: it
// reads BOTH sides and asserts they agree. A comment asking the next person
// to remember is precisely what this replaces.
//
// WHAT THE TIE COMPARES, AND WHAT IT DOES NOT. It compares LITERALS, not
// quantities. Go's `len()` counts BYTES; TypeScript's `.length` counts UTF-16
// CODE UNITS, so for non-ASCII the browser side is the more permissive of the
// two string bounds. That asymmetry is pre-existing, is safe in that
// direction — Go refuses first, so nothing unrepresentable crosses — and is
// recorded here rather than "fixed", because equal numerals is exactly the
// property that keeps the two sides from drifting apart.
//
// Go's per-line `maxCanvasTextFragments` is deliberately absent: the browser
// counts fragments CUMULATIVELY across a component, which is a different
// quantity, and pairing them would assert a tie that does not exist.
//
// NOT EVERY MIRROR COMES FROM ONE GO FILE. Story 7.4 projects `lineSpacing`
// across the channel, whose range is declared in
// `folio-go/internal/template/linespacing.go` and whose maximum that file
// explicitly calls "A STATED SANITY CEILING" — a number somebody will adjust.
// Raising it in Go alone would leave the browser silently dropping every
// snapshot of such a document, so its pair is tied here too, read from ITS
// OWN source rather than assumed to live beside the others.

const sourceDir = path.dirname(fileURLToPath(import.meta.url))
const goSources = {
  pageSetup: path.resolve(sourceDir, '../../folio-go/page_setup.go'),
  lineSpacing: path.resolve(sourceDir, '../../folio-go/internal/template/linespacing.go'),
  // Story 7.5's mirror is a PREDICATE, not a numeral, and it lives in a third
  // Go file. Reading it here is the point: a tie list that only ever reaches
  // the files it already knew about is the audit's blind spot restated.
  componentCommands: path.resolve(sourceDir, '../../folio-go/component_commands.go'),
} as const
const tsPath = path.join(sourceDir, 'engine-protocol.ts')

type GoSource = keyof typeof goSources
type Pair = Readonly<{ go: string; source: GoSource; ts: string; sites: ReadonlyArray<RegExp> }>

// Each pair names a Go constant, the Go FILE it is declared in, its
// TypeScript mirror, and the validator sites that must actually CONSUME the
// mirror. Without the site regexes a hoisted constant could sit in the file
// unused while the validator kept a stale inline literal, and the tie would
// pass while the canvas blanked.
const pairs: ReadonlyArray<Pair> = [
  { go: 'maxCanvasBodyText', source: 'pageSetup', ts: 'MAX_CANVAS_BODY_TEXT', sites: [/boundedString\('value', MAX_CANVAS_BODY_TEXT\)/, /fragment\.text\.length <= MAX_CANVAS_BODY_TEXT\b/] },
  { go: 'maxCanvasBodyTextLines', source: 'pageSetup', ts: 'MAX_CANVAS_BODY_TEXT_LINES', sites: [/value\.lines\.length > MAX_CANVAS_BODY_TEXT_LINES\b/] },
  { go: 'maxCanvasBodyTextFragments', source: 'pageSetup', ts: 'MAX_CANVAS_BODY_TEXT_FRAGMENTS', sites: [/fragments <= MAX_CANVAS_BODY_TEXT_FRAGMENTS\b/] },
  { go: 'maxCanvasPropertyString', source: 'pageSetup', ts: 'MAX_CANVAS_PROPERTY_STRING', sites: [/boundedString\(key, MAX_CANVAS_PROPERTY_STRING\)/] },
  { go: 'MinLineSpacingThousandths', source: 'lineSpacing', ts: 'MIN_LINE_SPACING_THOUSANDTHS', sites: [/component\.lineSpacing < MIN_LINE_SPACING_THOUSANDTHS\b/] },
  { go: 'MaxLineSpacingThousandths', source: 'lineSpacing', ts: 'MAX_LINE_SPACING_THOUSANDTHS', sites: [/component\.lineSpacing > MAX_LINE_SPACING_THOUSANDTHS\b/] },
]

// Both spellings a Go declaration can take: a file-scope `const NAME = N` and
// a member of a grouped `const ( … )` block, which is how linespacing.go
// declares its pair.
function goConstant(source: string, name: string): string | undefined {
  return source.match(new RegExp(`^(?:const[ \\t]+|[ \\t]+)${name} = (\\d+)$`, 'm'))?.[1]
}

function tsConstant(source: string, name: string): string | undefined {
  return source.match(new RegExp(`^export const ${name} = (\\d+)$`, 'm'))?.[1]
}

describe('canvas projection bounds mirror', () => {
  const sources: Record<GoSource, string> = { pageSetup: fs.readFileSync(goSources.pageSetup, 'utf8'), lineSpacing: fs.readFileSync(goSources.lineSpacing, 'utf8'), componentCommands: fs.readFileSync(goSources.componentCommands, 'utf8') }
  const goValue = (pair: Pair) => goConstant(sources[pair.source], pair.go)
  const ts = fs.readFileSync(tsPath, 'utf8')

  it('finds every declared pair on both sides of the channel', () => {
    // Non-vacuity. A regex that quietly stops matching is the exact failure
    // mode a tie assertion exists to prevent, so the found set is asserted
    // whole before any equality is claimed from it.
    expect(pairs.map((pair) => [pair.go, goValue(pair)])).toEqual(pairs.map((pair) => [pair.go, expect.stringMatching(/^\d+$/)]))
    expect(pairs.map((pair) => [pair.ts, tsConstant(ts, pair.ts)])).toEqual(pairs.map((pair) => [pair.ts, expect.stringMatching(/^\d+$/)]))
    // The count, and the fact that BOTH Go sources are actually read: a pair
    // list that quietly lost its only linespacing.go member would otherwise
    // leave this file reading one source and still calling itself the tie.
    expect(pairs).toHaveLength(6)
    // The NUMERAL pairs read two sources; componentCommands carries the
    // predicate tie below and deliberately contributes no pair.
    expect(new Set(pairs.map((pair) => pair.source))).toEqual(new Set(['pageSetup', 'lineSpacing']))
  })

  it('holds every Go bound and its TypeScript mirror at the same number', () => {
    expect(pairs.map((pair) => `${pair.go}=${goValue(pair)}`)).toEqual(pairs.map((pair) => `${pair.go}=${tsConstant(ts, pair.ts)}`))
    // The derivations, pinned as numerals so a silent joint edit of both
    // files still has to face the arithmetic recorded in the Go comments.
    expect(goConstant(sources.pageSetup, 'maxCanvasBodyTextLines')).toBe('1920')
    expect(goConstant(sources.pageSetup, 'maxCanvasBodyTextFragments')).toBe('65536')
    expect(goConstant(sources.pageSetup, 'maxCanvasBodyText')).toBe('1048576')
    expect(goConstant(sources.pageSetup, 'maxCanvasPropertyString')).toBe('512')
    expect(goConstant(sources.lineSpacing, 'MinLineSpacingThousandths')).toBe('1')
    expect(goConstant(sources.lineSpacing, 'MaxLineSpacingThousandths')).toBe('1000000')
  })

  it('consumes every mirrored constant at the validator site it bounds', () => {
    for (const pair of pairs) for (const site of pair.sites) expect(ts).toMatch(site)
  })

  it('keeps body text and identifiers on separate bounds on both sides', () => {
    // The split is the whole point (D-7.4.2 §3): the element VALUE must not
    // be governed by the identifier bound on either side of the channel.
    expect(sources.pageSetup).toMatch(/if len\(element\.Value\.Value\) > maxCanvasBodyText \{/)
    const identifiers = ts.match(/\[([^[\]]*)\]\.every\(optionalString\)/)?.[1] ?? ''
    expect(identifiers).toContain("'fontFamily'")
    expect(identifiers).not.toContain("'value'")
  })

  it('turns a one-sided edit of any pair red', () => {
    // The red-proof: mutate the TypeScript literal alone, exactly as a
    // careless raise-the-number change would, and the comparison must fail.
    for (const pair of pairs) {
      const drifted = ts.replace(new RegExp(`^export const ${pair.ts} = (\\d+)$`, 'm'), `export const ${pair.ts} = 7`)
      expect(drifted).not.toBe(ts)
      expect(tsConstant(drifted, pair.ts)).not.toBe(goValue(pair))
    }
    // And the same in the other direction, from the Go side — once per Go
    // SOURCE, because the linespacing.go pair is declared in a grouped const
    // block and a regex that only understood the file-scope spelling would
    // pass this test while tying nothing.
    const driftedLines = sources.pageSetup.replace(/^const maxCanvasBodyTextLines = (\d+)$/m, 'const maxCanvasBodyTextLines = 256')
    expect(driftedLines).not.toBe(sources.pageSetup)
    expect(goConstant(driftedLines, 'maxCanvasBodyTextLines')).not.toBe(tsConstant(ts, 'MAX_CANVAS_BODY_TEXT_LINES'))
    const driftedCeiling = sources.lineSpacing.replace(/^(\s+)MaxLineSpacingThousandths = (\d+)$/m, '$1MaxLineSpacingThousandths = 2000000')
    expect(driftedCeiling).not.toBe(sources.lineSpacing)
    expect(goConstant(driftedCeiling, 'MaxLineSpacingThousandths')).not.toBe(tsConstant(ts, 'MAX_LINE_SPACING_THOUSANDTHS'))
  })
})

// Go declares the list through named constants, so the identifiers are
// resolved before comparison: `[]string{bandPageHeader, bandPageFooter}` and
// `['pageHeader', 'pageFooter']` are the same claim spelled two ways, and it
// is the CLAIM that has to match, not the spelling.
function goBandsCappingVertically(source: string): ReadonlyArray<string> {
  const names = new Map<string, string>()
  for (const match of source.matchAll(/^[ \t]*(band[A-Za-z]+)\s+= "([^"]+)"$/gm)) names.set(match[1] as string, match[2] as string)
  const list = source.match(/^var bandsCappingVertically = \[\]string\{([^}]*)\}$/m)?.[1]
  if (list === undefined) return []
  return list.split(',').map((entry) => entry.trim()).filter((entry) => entry.length > 0).map((entry) => names.get(entry) ?? entry)
}

function tsBandsCappingVertically(source: string): ReadonlyArray<string> {
  const list = source.match(/^export const BANDS_CAPPING_VERTICALLY = \[([^\]]*)\]$/m)?.[1]
  if (list === undefined) return []
  return list.split(',').map((entry) => entry.trim().replace(/^'|'$/g, '')).filter((entry) => entry.length > 0)
}

// STORY 7.5's MIRROR, and the first one here that ties a PREDICATE rather
// than a numeral.
//
// The invariant: which bands cap a component's vertical extent. Go enforces
// it on the COMMAND path (containComponent) and TypeScript enforces it again
// on the PROJECTION path (isCanvas) — and the asymmetry is what makes a
// one-sided edit dangerous. Go's canvasComponents does not re-check
// containment when projecting, so Go already projects an out-of-band element
// happily and TypeScript alone kills it: lifting the cap in Go without
// lifting it here would ship a story that is invisible in the running app,
// because parseInbound would return undefined, terminate the worker and blank
// the canvas for exactly the documents the story exists to make authorable.
describe('band containment mirror', () => {
  const go = fs.readFileSync(goSources.componentCommands, 'utf8')
  const ts = fs.readFileSync(tsPath, 'utf8')

  it('reads a non-empty list from both sides', () => {
    // Non-vacuity first: a regex that quietly stops matching would make every
    // equality below true and meaningless.
    expect(goBandsCappingVertically(go)).toEqual(['pageHeader', 'pageFooter'])
    expect(tsBandsCappingVertically(ts)).toEqual(['pageHeader', 'pageFooter'])
  })

  it('agrees on which bands cap a component vertically', () => {
    expect(goBandsCappingVertically(go)).toEqual(tsBandsCappingVertically(ts))
    // The content band is on NEITHER list, which is the whole of Story 7.5:
    // a column has no height to be inside of.
    expect(goBandsCappingVertically(go)).not.toContain('content')
    expect(tsBandsCappingVertically(ts)).not.toContain('content')
  })

  it('consumes the list at the validator site it governs, on both sides', () => {
    // A list nothing reads would tie two dead declarations together while the
    // real gate kept its own inline spelling.
    expect(go).toMatch(/if !outside && slices\.Contains\(bandsCappingVertically, band\.Name\) \{/)
    expect(go).toMatch(/func containEdgeY\(band CanvasBand, value, limit geom\.Length\) geom\.Length \{\n\tif !slices\.Contains\(bandsCappingVertically, band\.Name\) \{/)
    expect(ts).toMatch(/BANDS_CAPPING_VERTICALLY\.includes\(component\.band as string\) && !\(box\.y \+ box\.height <= band\.height\)/)
  })

  it('keeps the HORIZONTAL cap universal on both sides', () => {
    // The column is unbounded vertically, never horizontally — so neither
    // side may guard its x check by band.
    expect(go).toMatch(/outside := x < 0 \|\| y < 0 \|\| width < 0 \|\| height < 0 \|\| x > geom\.Length\(band\.Width\) \|\| width > geom\.Length\(band\.Width\)-x$/m)
    expect(ts).toMatch(/^ {4}if \(!\(box\.x \+ box\.width <= band\.width\)\) return false$/m)
  })

  it('turns a one-sided edit of the predicate red', () => {
    const driftedTs = ts.replace(/^export const BANDS_CAPPING_VERTICALLY = \[([^\]]*)\]$/m, "export const BANDS_CAPPING_VERTICALLY = ['pageHeader', 'content', 'pageFooter']")
    expect(driftedTs).not.toBe(ts)
    expect(tsBandsCappingVertically(driftedTs)).not.toEqual(goBandsCappingVertically(go))
    const driftedGo = go.replace(/^var bandsCappingVertically = \[\]string\{([^}]*)\}$/m, 'var bandsCappingVertically = []string{bandPageHeader}')
    expect(driftedGo).not.toBe(go)
    expect(goBandsCappingVertically(driftedGo)).not.toEqual(tsBandsCappingVertically(ts))
  })
})
