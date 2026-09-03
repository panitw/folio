import { describe, expect, it } from 'vitest'
import { addableFamilyCount, indexCategories, indexScripts, offeredFamilies, type FamilySource } from './font-index'
import { browserRows, buttonLabel, buttonName, confirmLabel, emptyStateHeading, familiesPerPage, filterRows, filtersActive, latinSample, noFilters, pageCount, pageLine, pageOf, pendingLine, resultLine, rowState, rowTierNote, scriptBadge, sortRows, specimenFor, thaiSample, weightLine, type BrowserRow } from './font-browser-model'
import type { StoredFace } from './font-store'

// THE FONT BROWSER'S LOGIC, ASSERTED AGAINST THE DESIGN IT WAS PORTED FROM
// (Story 16.3).
//
// `Font Browser.dc.html`'s `renderVals()` settles the edge cases; these are the
// assertions that the port kept them, plus the three places the port
// DELIBERATELY departs from the mockup — the dropped `Most styles` arm, the
// search predicate narrowed to family and category, and a footer that states
// what this product actually embeds.

const webRow = (family: string, category: string, scripts: ReadonlyArray<'latin' | 'thai'>, popularity: number): FamilySource =>
  ({ tier: 'web', family, row: { family, category, scripts, variable: false, popularity } })

const storedRecord = (family: string, scripts: ReadonlyArray<string>): StoredFace => ({
  key: 'a'.repeat(64), family, style: 'Regular', licence: 'OFL-1.1', licenceText: 'terms', copyright: 'c',
  source: 'google/fonts — ofl/x/X-Regular.ttf, fetched 2026-09-03', mediaType: 'font/ttf', scripts,
  fetchedAt: '2026-09-03', byteLength: 4,
})

const row = (family: string, category: string | undefined, scripts: ReadonlyArray<string>, popularity?: number): BrowserRow =>
  ({ family, source: webRow(family, category ?? 'Serif', [], popularity ?? 0), ...(category === undefined ? {} : { category }), ...(popularity === undefined ? {} : { popularity }), scripts })

describe('the font browser describes the families it is given', () => {
  it('reads category, popularity and scripts off the snapshot, and says nothing it cannot read', () => {
    const [sarabun] = browserRows([webRow('Sarabun', 'Sans Serif', ['latin', 'thai'], 12)])
    expect(sarabun?.category).toBe('Sans Serif')
    expect(sarabun?.scripts).toEqual(['latin', 'thai'])
    // A family the snapshot has no row for keeps its tier's own scripts and
    // carries no category at all — never a guessed one.
    const [invented] = browserRows([{ tier: 'stored', family: 'A Face Only This Machine Has', record: storedRecord('A Face Only This Machine Has', ['latin']) }])
    expect(invented?.category).toBeUndefined()
    expect(invented?.popularity).toBeUndefined()
    expect(invented?.scripts).toEqual(['latin'])
  })

  it('names every tier a row can come from, and cannot gain a fourth silently', () => {
    expect(rowTierNote(webRow('Kanit', 'Sans Serif', ['latin', 'thai'], 8))).toBe('downloaded when you add it')
    expect(rowTierNote({ tier: 'stored', family: 'Kanit', record: storedRecord('Kanit', ['latin']) })).toBe('downloaded to this machine')
    expect(() => rowTierNote({ tier: 'gossip' } as unknown as FamilySource)).toThrow(/gossip/)
  })

  it('derives its chip vocabularies from the snapshot rather than from the mockup', () => {
    // The mockup draws four category chips and six writing-system ones
    // (All/Latin/Thai/Cyrillic/Greek). Neither list is this snapshot's.
    expect(indexCategories).toContain('Handwriting')
    expect(indexCategories.length).toBeGreaterThanOrEqual(5)
    expect([...indexScripts]).toEqual(['latin', 'thai'])
    expect(indexScripts).not.toContain('cyrillic')
  })
})

describe('the mockup filter predicate, narrowed to what the snapshot carries', () => {
  const rows = [row('Sarabun', 'Sans Serif', ['latin', 'thai'], 1), row('Lora', 'Serif', ['latin'], 2), row('Chonburi', 'Display', ['latin', 'thai'], 3)]

  it('matches family and category, and has no designer field to match', () => {
    expect(filterRows(rows, { ...noFilters, query: 'sara' }).map((entry) => entry.family)).toEqual(['Sarabun'])
    expect(filterRows(rows, { ...noFilters, query: 'display' }).map((entry) => entry.family)).toEqual(['Chonburi'])
    // `Cadson Demak` designs both Sarabun and Chonburi. The mockup's predicate
    // would have found them; this snapshot carries no designer at all.
    expect(filterRows(rows, { ...noFilters, query: 'cadson' })).toEqual([])
  })

  it('intersects the script chip with the category chips', () => {
    expect(filterRows(rows, { ...noFilters, script: 'thai' }).map((entry) => entry.family)).toEqual(['Sarabun', 'Chonburi'])
    expect(filterRows(rows, { ...noFilters, script: 'thai', categories: ['Display'] }).map((entry) => entry.family)).toEqual(['Chonburi'])
    expect(filterRows(rows, { ...noFilters, script: 'thai', categories: ['Serif'] })).toEqual([])
  })

  it('shows `reset filters` exactly while a filter is active', () => {
    expect(filtersActive(noFilters)).toBe(false)
    expect(filtersActive({ ...noFilters, query: '  ' })).toBe(false)
    expect(filtersActive({ ...noFilters, query: 'sara' })).toBe(true)
    expect(filtersActive({ ...noFilters, script: 'thai' })).toBe(true)
    expect(filtersActive({ ...noFilters, categories: ['Serif'] })).toBe(true)
  })
})

describe('the two sort arms, and the one the port dropped', () => {
  const rows = [row('Zed', 'Serif', ['latin'], 9), row('Alpha', 'Serif', ['latin'], 40), row('Unranked', undefined, ['latin'])]

  it('orders Trending by the snapshot rank, most popular first', () => {
    expect(sortRows(rows, 'Trending').map((entry) => entry.family)).toEqual(['Zed', 'Alpha', 'Unranked'])
  })

  it('sorts a family the snapshot does not rank LAST, never first', () => {
    // An absent rank read as zero would put the family the snapshot says least
    // about at the head of the list it is ordering.
    expect(sortRows(rows, 'Trending').at(-1)?.family).toBe('Unranked')
  })

  it('orders A – Z by family name', () => {
    expect(sortRows(rows, 'A – Z').map((entry) => entry.family)).toEqual(['Alpha', 'Unranked', 'Zed'])
  })
})

describe('the specimen, the badge and the row button', () => {
  const sarabun = row('Sarabun', 'Sans Serif', ['latin', 'thai'], 1)
  const lora = row('Lora', 'Serif', ['latin'], 2)

  it('gives a Thai-covering family the Thai sample while the toggle is on', () => {
    expect(specimenFor(sarabun, '', true)).toBe(thaiSample)
    expect(specimenFor(sarabun, '', false)).toBe(latinSample)
    expect(specimenFor(lora, '', true)).toBe(latinSample)
  })

  it('lets the author\'s own words win over both samples', () => {
    expect(specimenFor(sarabun, 'ธนาคารกรุงศรี', true)).toBe('ธนาคารกรุงศรี')
    expect(specimenFor(lora, '  spaced  ', true)).toBe('spaced')
  })

  it('badges Thai coverage the design\'s way, and says so when the snapshot states nothing', () => {
    expect(scriptBadge(sarabun)).toBe('Thai + Latin')
    expect(scriptBadge(lora)).toBe('Latin')
    expect(scriptBadge(row('Allkin', 'Display', []))).toBe('script not stated')
  })

  it('carries the mockup\'s three button states in its own order of precedence', () => {
    expect(rowState('Lora', ['Lora'], ['Lora'])).toBe('in-template')
    expect(rowState('Lora', [], ['Lora'])).toBe('staged')
    expect(rowState('Lora', [], [])).toBe('addable')
    expect(buttonLabel('in-template')).toBe('In template')
    expect(buttonLabel('staged')).toBe('✓ Added')
    expect(buttonLabel('addable')).toBe('+ Add')
    // The visible label is the design's; the accessible name has to name the
    // family, because a screen reader hears twelve buttons all called `+ Add`.
    expect(buttonName('Lora', 'addable')).toBe('Add Lora to this template')
    expect(buttonName('Lora', 'staged')).toBe('Remove Lora from the families to add')
    expect(buttonName('Lora', 'in-template')).toBe('Lora is in this template')
  })
})

describe('the footer states what is about to go into the file', () => {
  it('carries the mockup\'s pending line and confirm label', () => {
    expect(pendingLine(0)).toBe('Select families to add to this template')
    expect(pendingLine(1)).toBe('1 family ready to embed')
    expect(pendingLine(3)).toBe('3 families ready to embed')
    expect(confirmLabel(0)).toBe('Add to template')
    expect(confirmLabel(3)).toBe('Add 3 to template')
  })

  it('corrects the mockup\'s weight line to what this product embeds', () => {
    // The mockup prints `≈ N weights · subset latin+thai`. This product embeds
    // one upright Regular per family and subsets nothing, so a footer repeating
    // the mockup would be the one region that lies about the file.
    expect(weightLine(0)).toBe('')
    expect(weightLine(1)).toBe('1 face · one upright Regular each, no bold or italic')
    expect(weightLine(3)).toBe('3 faces · one upright Regular each, no bold or italic')
    // THE FOOTER MAY NOT SPEAK ABOUT SUBSETTING AT ALL, in either direction.
    // The product DOES subset (at PDF render, over the glyphs the document
    // uses), so both "subset latin+thai" and "whole file, not subset" are false
    // sentences — the second shipped briefly and this is what catches it.
    expect(weightLine(3)).not.toMatch(/weights|subset/i)
    // AND IT MAY NOT NAME A DESTINATION, because Story 16.5 inverts it.
    expect(weightLine(3)).not.toMatch(/template|file|document|embed|install/i)
  })
})

describe('the result line, the empty state and the page bound', () => {
  it('states the shown count against the matching count and the addable total', () => {
    expect(resultLine(5, 5)).toBe(`5 of ${addableFamilyCount} families`)
    expect(resultLine(12, 340)).toBe(`12 of 340 matching families, out of ${addableFamilyCount}`)
    // NEVER 1,946: that is what the source PUBLISHED on the snapshot date.
    expect(resultLine(5, 5)).not.toContain('1946')
  })

  it('names the query in the empty state, and does not invent one when there is none', () => {
    expect(emptyStateHeading('qqq')).toBe('No families match “qqq”')
    expect(emptyStateHeading('  ')).toBe('No families match these filters')
  })

  it('bounds a page at `familiesPerPage` and clamps a page index past the end', () => {
    const rows = Array.from({ length: 30 }, (_, index) => row(`Family ${index}`, 'Serif', ['latin'], index))
    expect(pageOf(rows, 0)).toHaveLength(familiesPerPage)
    expect(pageOf(rows, 2)).toHaveLength(30 - 2 * familiesPerPage)
    expect(pageCount(30)).toBe(3)
    expect(pageCount(0)).toBe(1)
    expect(pageOf(rows, 99).map((entry) => entry.family)).toEqual(pageOf(rows, 2).map((entry) => entry.family))
    expect(pageLine(0, 30)).toBe('Page 1 of 3')
    expect(pageLine(9, 30)).toBe('Page 3 of 3')
  })

  it('never puts more families on a page than may be registered for preview at once', () => {
    // The bound is the whole reason the browser pages rather than scrolls: every
    // row on screen wants a real face registered for it.
    const everything = browserRows(offeredFamilies(''))
    expect(everything.length).toBeGreaterThan(1000)
    expect(pageOf(everything, 0).length).toBe(familiesPerPage)
  })
})
