import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { BANDS_CAPPING_VERTICALLY, CAPPING_BANDS, type CappingBand } from './engine-protocol'

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
// Story 7.6's THIRD consumer of the band-containment tie. The drag clamp used
// to cap every band vertically with its own inline rule — a fourth spelling
// of an invariant three files already state — and lifting the content band in
// Go and in the protocol while leaving it clamped here would have shipped a
// column reachable by command and not by hand. It reads the list; this test
// reads it reading the list.
const dragClampPath = path.join(sourceDir, 'resize-anchor.ts')

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
  // Story 8.1. maxCanvasFontFamilies had a TypeScript mirror and no pair from
  // the day it was written — the one-sided edit this list exists to catch,
  // sitting inside the list itself. It is tied here alongside the per-chain
  // entry bound the same story added, so the projection's font half is not
  // half-tied.
  { go: 'maxCanvasFontFamilies', source: 'pageSetup', ts: 'MAX_ENGINE_FONT_FAMILIES', sites: [/value\.fontFamilies\.length > MAX_ENGINE_FONT_FAMILIES\b/] },
  { go: 'maxCanvasFontChainEntries', source: 'componentCommands', ts: 'MAX_ENGINE_FONT_CHAIN_ENTRIES', sites: [/chain\.entries\.length <= MAX_ENGINE_FONT_CHAIN_ENTRIES\b/] },
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
    expect(pairs).toHaveLength(8)
    // The NUMERAL pairs now read all three sources: componentCommands carries
    // the band-containment predicate tie below AND, since Story 8.1, the
    // per-chain entry bound the font-chain commands and the projection share.
    expect(new Set(pairs.map((pair) => pair.source))).toEqual(new Set(['pageSetup', 'lineSpacing', 'componentCommands']))
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
  const dragClamp = fs.readFileSync(dragClampPath, 'utf8')

  it('reads a non-empty list from every side', () => {
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

  it('consumes the list at the site it governs, in all three consumers', () => {
    // A list nothing reads would tie dead declarations together while the
    // real gates kept their own inline spellings.
    expect(go).toMatch(/if !outside && slices\.Contains\(bandsCappingVertically, band\.Name\) \{/)
    expect(go).toMatch(/func containEdgeY\(band CanvasBand, value, limit geom\.Length\) geom\.Length \{\n\tif !slices\.Contains\(bandsCappingVertically, band\.Name\) \{/)
    expect(ts).toMatch(/BANDS_CAPPING_VERTICALLY\.includes\(component\.band as string\) && !\(box\.y \+ box\.height <= band\.height\)/)
    // The DRAG CLAMP (DW-36), which reads the list from engine-protocol.ts
    // rather than restating it, and gates the ONE vertical limit both of its
    // vertical clamps consume.
    expect(dragClamp).toMatch(/import \{ BANDS_CAPPING_VERTICALLY, type CanvasProjection \} from '\.\/engine-protocol'/)
    expect(dragClamp).toMatch(/^ {2}const limitHeight = limit && BANDS_CAPPING_VERTICALLY\.includes\(limit\.band\) \? limit\.height : Number\.POSITIVE_INFINITY$/m)
  })

  // STORY 12.1's REVIEW ADDED TWO MORE SPELLINGS OF THIS LIST AND THIS IS WHERE
  // THEY WERE PUT BACK.
  //
  // The band-height command takes a BAND as a parameter, so it needed the list
  // as a type and the panel needed it as something to iterate. Both were first
  // written out again — `SettableBand` in band-height-command.ts and
  // `settableBands` in App.tsx — which made four and five copies of a list whose
  // entire safety property is that this file reads it on both sides. A copy
  // outside the census is the only kind that can go stale unnoticed, and the
  // cost of this one going stale is the cost of every other copy of it: the
  // browser drops the snapshot, terminates the worker and blanks the canvas.
  it('hands the SAME list to the type and the array the command path consumes', () => {
    // The array is the same object, not an equal one: a member added to
    // BANDS_CAPPING_VERTICALLY is in CAPPING_BANDS by identity.
    expect(CAPPING_BANDS).toBe(BANDS_CAPPING_VERTICALLY)
    expect([...CAPPING_BANDS]).toEqual(goBandsCappingVertically(go))
    // AND THE UNION, TIED THE SAME WAY THE ARRAYS ARE: by reading the source.
    // A type cannot be enumerated at run time, so the DERIVATION is read out of
    // engine-protocol.ts instead — CappingBand must be the projection's own
    // band-name union (which canvas_projection_wire_test.go pins against Go)
    // minus the one band with no height to cap with — and the union that
    // derivation produces is then computed from that same union's source text
    // and compared to Go's list. A third band added in Go reaches the projection
    // union, reaches CappingBand through Exclude, and reds here.
    expect(ts).toMatch(/^export type CappingBand = Exclude<CanvasProjection\['bands'\]\[number\]\['name'\], 'content'>$/m)
    const projectedBandNames = ts.match(/^[ \t]+bands: ReadonlyArray<Readonly<\{ name: ([^;]+);/m)?.[1]
    expect(projectedBandNames).toBeDefined()
    const derivedUnion = (projectedBandNames ?? '').split('|').map((entry) => entry.trim().replace(/^'|'$/g, '')).filter((entry) => entry !== '' && entry !== 'content')
    expect(derivedUnion.sort()).toEqual([...goBandsCappingVertically(go)].sort())
    // The type-level half of the same claim, which tsc checks and vitest cannot:
    // a Record keyed by the union must carry EXACTLY the union's members, so a
    // member gained or lost on either side stops this file compiling.
    const everyCappingBand: Record<CappingBand, true> = { pageHeader: true, pageFooter: true }
    expect(Object.keys(everyCappingBand).sort()).toEqual([...goBandsCappingVertically(go)].sort())
    // And the consumers hold no copy of their own: they import these two.
    const app = fs.readFileSync(path.join(sourceDir, 'App.tsx'), 'utf8')
    const factory = fs.readFileSync(path.join(sourceDir, 'band-height-command.ts'), 'utf8')
    expect(app).toContain('CAPPING_BANDS')
    expect(app).not.toMatch(/\['pageHeader', 'pageFooter'\]/)
    expect(factory).toMatch(/^import type \{ CappingBand \} from '\.\/engine-protocol'$/m)
    expect(factory).not.toMatch(/'pageHeader' \| 'pageFooter'/)
  })

  it('keeps the HORIZONTAL cap universal in every consumer', () => {
    // The column is unbounded vertically, never horizontally — so neither
    // side may guard its x check by band.
    expect(go).toMatch(/outside := x < 0 \|\| y < 0 \|\| width < 0 \|\| height < 0 \|\| x > geom\.Length\(band\.Width\) \|\| width > geom\.Length\(band\.Width\)-x$/m)
    expect(ts).toMatch(/^ {4}if \(!\(box\.x \+ box\.width <= band\.width\)\) return false$/m)
    // The drag clamp's width limit takes the band's size unconditionally,
    // with no band in the condition at all.
    expect(dragClamp).toMatch(/^ {2}const limitWidth = limit \? limit\.width : Number\.POSITIVE_INFINITY$/m)
  })

  it('turns a one-sided edit of the predicate red', () => {
    const driftedTs = ts.replace(/^export const BANDS_CAPPING_VERTICALLY = \[([^\]]*)\]$/m, "export const BANDS_CAPPING_VERTICALLY = ['pageHeader', 'content', 'pageFooter']")
    expect(driftedTs).not.toBe(ts)
    expect(tsBandsCappingVertically(driftedTs)).not.toEqual(goBandsCappingVertically(go))
    const driftedGo = go.replace(/^var bandsCappingVertically = \[\]string\{([^}]*)\}$/m, 'var bandsCappingVertically = []string{bandPageHeader}')
    expect(driftedGo).not.toBe(go)
    expect(goBandsCappingVertically(driftedGo)).not.toEqual(tsBandsCappingVertically(ts))
    // And the third consumer's own drift: a drag clamp that stopped reading
    // the list and re-stated the rule inline is the shape DW-36 named — "a
    // fourth spelling of the tie" — and it would be invisible to every
    // assertion above, because both declarations would still agree.
    const driftedClamp = dragClamp.replace(/const limitHeight = limit && BANDS_CAPPING_VERTICALLY\.includes\(limit\.band\) \? limit\.height : Number\.POSITIVE_INFINITY/, 'const limitHeight = limit ? limit.height : Number.POSITIVE_INFINITY')
    expect(driftedClamp).not.toBe(dragClamp)
    expect(driftedClamp).not.toMatch(/BANDS_CAPPING_VERTICALLY\.includes\(limit\.band\)/)
  })
})

// Go declares the rule as a disjunction inside the property command's length
// arm; TypeScript declares it as a list the inspector reads. Both are resolved
// to a set of KEY NAMES before comparison, because it is the CLAIM — which
// fields the engine refuses at or below zero — that has to match, not the
// spelling.
function goPositiveLengthFields(source: string): ReadonlyArray<string> {
  const clause = source.match(/^\t+if \((key == "\w+"(?: \|\| key == "\w+")*)\) && length <= 0 \{$/m)?.[1]
  if (clause === undefined) return []
  return [...clause.matchAll(/key == "(\w+)"/g)].map((match) => match[1] as string)
}

function tsPositiveLengthFields(source: string): ReadonlyArray<string> {
  const list = source.match(/^export const POSITIVE_LENGTH_FIELDS: ReadonlyArray<PropertyField> = \[([^\]]*)\]$/m)?.[1]
  if (list === undefined) return []
  return list.split(',').map((entry) => entry.trim().replace(/^'|'$/g, '')).filter((entry) => entry.length > 0)
}

// STORY 17.4's MIRROR, and the second one here that ties a PREDICATE.
//
// The invariant: which property keys the engine refuses at or below zero. Go
// enforces it on the COMMAND path; the inspector's ARROW STEP reads it to know
// where to stop, so that a keypress never proposes a value the command path
// will refuse. The asymmetry is what makes a one-sided edit dangerous in BOTH
// directions: adding a key in Go alone would let an arrow step a field into a
// refusal the author reads as a mysteriously rejected nudge, and dropping one
// in Go alone would leave the panel clamping a field the engine no longer
// bounds. `x` and `y` are on NEITHER list, which is the reason the tie is over
// a list at all rather than over "the numeric fields" — their own floor comes
// from `containComponent` instead, tied in the describe below.
describe('positive length rule mirror', () => {
  const go = fs.readFileSync(goSources.componentCommands, 'utf8')
  const ts = fs.readFileSync(path.join(sourceDir, 'component-property-command.ts'), 'utf8')
  const panel = fs.readFileSync(path.join(sourceDir, 'App.tsx'), 'utf8')

  it('reads a non-empty list from every side', () => {
    // Non-vacuity first: a regex that quietly stops matching would make the
    // equality below true and meaningless.
    expect(goPositiveLengthFields(go)).toEqual(['width', 'height', 'fontSize', 'borderWidth'])
    expect(tsPositiveLengthFields(ts)).toEqual(['width', 'height', 'fontSize', 'borderWidth'])
  })

  it('agrees on which property keys must stay positive', () => {
    expect(goPositiveLengthFields(go)).toEqual(tsPositiveLengthFields(ts))
    // x and y are on neither side of THIS rule. That does NOT make them
    // unbounded, which is what an earlier reading concluded: they are bounded
    // below at the band origin by `containComponent`, tied in the describe
    // below. What these four assertions pin is only that x and y are not
    // subject to the strictly tighter `> 0` rule.
    expect(goPositiveLengthFields(go)).not.toContain('x')
    expect(tsPositiveLengthFields(ts)).not.toContain('x')
    expect(goPositiveLengthFields(go)).not.toContain('y')
    expect(tsPositiveLengthFields(ts)).not.toContain('y')
  })

  it('consumes the list at the site it governs, and the line-spacing pair beside it', () => {
    // A list nothing reads would tie a dead declaration to a live Go rule
    // while the step kept its own inline spelling.
    expect(go).toMatch(/if \(key == "width" \|\| key == "height" \|\| key == "fontSize" \|\| key == "borderWidth"\) && length <= 0 \{\n\t+return fmt\.Errorf\("%s must be positive", key\)$/m)
    expect(panel).toMatch(/POSITIVE_LENGTH_FIELDS\.includes\(field\) \? 1 : /)
    // The step's OTHER bound comes from the already-tied line-spacing pair
    // rather than from two fresh literals in the panel.
    expect(panel).toMatch(/field === 'lineSpacing' \? MIN_LINE_SPACING_THOUSANDTHS :/)
    expect(panel).toMatch(/field === 'lineSpacing' \? MAX_LINE_SPACING_THOUSANDTHS :/)
  })

  it('turns a one-sided edit of the rule red', () => {
    const driftedTs = ts.replace(/^export const POSITIVE_LENGTH_FIELDS: ReadonlyArray<PropertyField> = \[([^\]]*)\]$/m, "export const POSITIVE_LENGTH_FIELDS: ReadonlyArray<PropertyField> = ['width', 'height', 'fontSize']")
    expect(driftedTs).not.toBe(ts)
    expect(tsPositiveLengthFields(driftedTs)).not.toEqual(goPositiveLengthFields(go))
    const driftedGo = go.replace(/if \(key == "width" \|\| key == "height" \|\| key == "fontSize" \|\| key == "borderWidth"\) && length <= 0 \{/, 'if (key == "width" || key == "height" || key == "fontSize" || key == "borderWidth" || key == "x") && length <= 0 {')
    expect(driftedGo).not.toBe(go)
    expect(goPositiveLengthFields(driftedGo)).not.toEqual(tsPositiveLengthFields(ts))
  })
})

// THE SECOND BOUND ON THE PROPERTY PATH, added at Story 17.4's review.
//
// `updateComponentPropertiesInPlace` does not stop at the `> 0` rule above: it
// calls `containComponent` on every id after applying the changes
// (`component_commands.go:880`), and that predicate opens by refusing NEGATIVE
// geometry. So `x` and `y` — absent from the `> 0` list, and therefore read as
// "unbounded" when this story was planned — have a floor of ZERO on exactly the
// path an arrow step uses. The panel mirrors that floor through
// `ORIGIN_FLOOR_FIELDS` rather than restating a literal.
//
// ⚠ WHAT THIS TIE DELIBERATELY DOES NOT COVER: the same predicate also bounds
// x, y, width and height ABOVE, against the band extents. The panel does not
// clamp to those, so a step at the band edge still reaches the engine's own
// located refusal. The ceiling is PER-COMPONENT — two components of equal width
// at different x have different width ceilings — so a selection-wide clamp
// needs a ruling this story does not carry. It is recorded as an open question
// in the story's Spec Change Log, and named here so the omission is visible
// rather than inferred from silence.
function goNonNegativeGeometryFields(source: string): ReadonlyArray<string> {
  const clause = source.match(/^\toutside := ((?:\w+ < 0 \|\| )+)/m)?.[1]
  if (clause === undefined) return []
  return [...clause.matchAll(/(\w+) < 0/g)].map((match) => match[1] as string)
}

function tsOriginFloorFields(source: string): ReadonlyArray<string> {
  const list = source.match(/^export const ORIGIN_FLOOR_FIELDS: ReadonlyArray<PropertyField> = \[([^\]]*)\]$/m)?.[1]
  if (list === undefined) return []
  return list.split(',').map((entry) => entry.trim().replace(/^'|'$/g, '')).filter((entry) => entry.length > 0)
}

describe('origin floor mirror', () => {
  const go = fs.readFileSync(goSources.componentCommands, 'utf8')
  const ts = fs.readFileSync(path.join(sourceDir, 'component-property-command.ts'), 'utf8')
  const panel = fs.readFileSync(path.join(sourceDir, 'App.tsx'), 'utf8')

  it('reads the negative-geometry refusal from Go and the floor list from the panel', () => {
    // Non-vacuity: a regex that quietly stopped matching would make every
    // comparison below true and meaningless.
    expect(goNonNegativeGeometryFields(go)).toEqual(['x', 'y', 'width', 'height'])
    expect(tsOriginFloorFields(ts)).toEqual(['x', 'y'])
  })

  it('mirrors exactly the fields whose floor comes from containment alone', () => {
    // Go refuses four fields below zero. Two of them, width and height, are
    // already held to a STRICTLY TIGHTER floor by the `> 0` rule, so the panel
    // clamps them there instead; the remaining two are the origin-floor list.
    // Stated as a partition, so a field moving between the two Go rules cannot
    // leave both TypeScript lists satisfied.
    expect([...tsOriginFloorFields(ts), ...tsPositiveLengthFields(ts)].filter((field) => goNonNegativeGeometryFields(go).includes(field)).sort())
      .toEqual([...goNonNegativeGeometryFields(go)].sort())
    expect(tsOriginFloorFields(ts).filter((field) => tsPositiveLengthFields(ts).includes(field))).toEqual([])
  })

  it('consumes the list at the site it governs', () => {
    expect(panel).toMatch(/ORIGIN_FLOOR_FIELDS\.includes\(field\) \? 0 : undefined$/m)
  })

  it('turns a one-sided edit of the rule red', () => {
    const driftedTs = ts.replace(/^export const ORIGIN_FLOOR_FIELDS: ReadonlyArray<PropertyField> = \[([^\]]*)\]$/m, "export const ORIGIN_FLOOR_FIELDS: ReadonlyArray<PropertyField> = ['x']")
    expect(driftedTs).not.toBe(ts)
    expect(tsOriginFloorFields(driftedTs)).not.toEqual(tsOriginFloorFields(ts))
    const driftedGo = go.replace(/^\toutside := x < 0 \|\| y < 0 \|\| /m, '\toutside := x < 0 || ')
    expect(driftedGo).not.toBe(go)
    expect(goNonNegativeGeometryFields(driftedGo)).not.toEqual(goNonNegativeGeometryFields(go))
  })
})
