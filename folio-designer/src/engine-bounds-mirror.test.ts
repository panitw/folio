import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// D-7.4.5 / DW-25. `engine-protocol.ts` hand-copies the canvas projection
// bounds `folio-go/page_setup.go` declares. There is no shared source and no
// codegen, so the only thing standing between the two files is this test: it
// reads BOTH sources and asserts the pairs are equal. A comment asking the
// next person to remember is precisely what this replaces.
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
  const sources: Record<GoSource, string> = { pageSetup: fs.readFileSync(goSources.pageSetup, 'utf8'), lineSpacing: fs.readFileSync(goSources.lineSpacing, 'utf8') }
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
