import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { catalogueFaces } from './generated/font-catalogue'
import { familyIndex, familyIndexPublishedFamilies, familyIndexSnapshotDate } from './generated/font-index'
import { blankComments } from '../scripts/forbidden-font-hosts.mjs'
import { addableFamilyCount, familyIndexDisclosure, indexExcludedCjkFamilies, localTierHolds, offeredFamilies, webFamilies } from './font-index'

// STORY 16.1 — THE TWO TIERS AND THE JOIN BETWEEN THEM (D-16.R.3, D-16.R.2).

const here = path.dirname(fileURLToPath(import.meta.url))
const manifest: ReadonlyArray<{ family: string; licence: string }> = JSON.parse(fs.readFileSync(path.join(here, '..', 'font-catalogue.json'), 'utf8'))

/**
 * THE STORY 16.1a BATCH, WRITTEN OUT RATHER THAN DERIVED.
 *
 * Derived from the catalogue it would be a tautology — "the families in the
 * catalogue are in the catalogue" — and the property under test is that these
 * TEN SPECIFIC families, the refused head of the popularity distribution, are
 * the ones an author can now reach. The membership rule that produced it is
 * D-16.R.16's, corrected by D-16.R.19: the refused families within the top 20
 * by `popularity` on the committed snapshot, minus CJK, minus the families the
 * tier already held, minus the `shippedFamilies` collisions, minus anything with
 * no obtainable static from its own project upstream (`Google Sans`, which
 * publishes none, and `Jost`, whose static names itself `Jost*`).
 */
const batchFamilies: ReadonlyArray<string> = [
  'Arimo', 'DM Sans', 'Lora', 'Montserrat', 'Open Sans',
  'Oswald', 'Plus Jakarta Sans', 'Roboto Condensed', 'Roboto Mono', 'Roboto Slab',
]

describe('the build-time index snapshot', () => {
  // NON-VACUITY FIRST. Every filter below is over `familyIndex`, and an empty
  // or truncated snapshot satisfies all of them silently.
  it('ships a dated snapshot of a real population, and never fetches the list at runtime', () => {
    expect(familyIndex.length).toBeGreaterThan(1000)
    expect(familyIndexPublishedFamilies).toBeGreaterThanOrEqual(familyIndex.length)
    expect(familyIndexSnapshotDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    // THE LIST IS NOT FETCHABLE BY A BROWSER — the endpoint sends no
    // access-control-allow-origin — so no production module may reach for it.
    // `font-source.ts` names the host only because the BUILD script imports the
    // constant from there; nothing calls it.
    const production = fs.readdirSync(here, { recursive: true })
      .filter((entry): entry is string => typeof entry === 'string' && /\.tsx?$/.test(entry) && !/\.test\.tsx?$/.test(entry))
    expect(production.length).toBeGreaterThan(5)
    for (const entry of production) {
      // Comment-blanked, using the host scanner's own blanker: the generated
      // module NAMES the endpoint in its header to say the list came from a
      // build-time snapshot of it, and a check that could not tell that
      // disclosure from a call site would be satisfied by deleting the
      // disclosure.
      const code = blankComments(fs.readFileSync(path.join(here, entry), 'utf8'), '.ts')
      expect(code, `${entry} must not fetch the family index at runtime`).not.toMatch(/metadata\/fonts/)
    }
  })

  // THE NON-GOAL IS HONOURED IN THE SNAPSHOT, not in the browser: CJK stays on
  // the shipped-face path, so those rows never ship at all.
  it('excludes CJK families from the snapshot rather than filtering them in the browser', () => {
    expect(indexExcludedCjkFamilies).toBeGreaterThan(0)
    // A COROLLARY, NOT THE CHECK. `scriptsOf` emits only `latin` and `thai`, so
    // this line is structurally incapable of failing whatever the filter does —
    // the filter itself is exercised by the test below, over a fixture.
    expect(familyIndex.some((row) => row.scripts.includes('cjk'))).toBe(false)
  })

  // THE REAL FILTER, RUN. `trimSnapshot` — not a restatement of it — over a
  // hand-written `familyMetadataList` carrying one family per CJK subset the
  // exclusion list names, plus two that must survive. Asserting the returned
  // families AND the excluded count directly is what makes deleting a subset
  // from `cjkSubsets` red: nothing else in this suite reads that list, because
  // the shipped snapshot is a fixed artefact that a narrower filter would not
  // change until it were regenerated.
  it('runs the CJK exclusion over every subset it names, and counts what it excluded', async () => {
    const { trimSnapshot } = await import('../scripts/build-font-index.mjs')
    const cjk = ['chinese-simplified', 'chinese-traditional', 'chinese-hongkong', 'japanese', 'korean']
    const published = {
      familyMetadataList: [
        ...cjk.map((subset, index) => ({ family: `CJK ${subset}`, category: 'SANS_SERIF', subsets: [subset, 'latin', 'menu'], fonts: { 400: {} }, popularity: index + 1 })),
        { family: 'Variable Latin', category: 'SANS_SERIF', subsets: ['latin', 'menu'], axes: [{ tag: 'wght' }], fonts: { 400: {} }, popularity: 6 },
        { family: 'Static Thai', category: 'SERIF', subsets: ['latin', 'thai', 'menu'], fonts: { 400: {}, 700: {} }, popularity: 7 },
      ],
    }
    const snapshot = trimSnapshot(published, '2026-09-03')
    expect(snapshot.families.map((entry) => entry.family), 'every CJK subset in the list is excluded and nothing else is').toEqual(['Variable Latin', 'Static Thai'])
    expect(snapshot.publishedFamilies).toBe(7)
    expect(snapshot.excludedCjkFamilies).toBe(5)
    expect(snapshot.snapshotDate).toBe('2026-09-03')
    // AND THE TRIM ITSELF: `menu` is a subsetting artefact and is dropped, and
    // `axes` is carried as the variable-only PREDICTION the browser filters on.
    expect(snapshot.families[0].axes).toEqual(['wght'])
    expect(snapshot.families[1].axes).toEqual([])
    expect(snapshot.families[1].subsets).toEqual(['latin', 'thai'])
    expect(snapshot.families[1].styles).toEqual(['400', '700'])
  })

  // AN OFFLINE RELEASE BUILD IS A SHIPPED GATE, so the step that emits this
  // module MUST NOT REACH THE NETWORK. Asserted by running the real emit step
  // with `fetch` replaced by a thrower — a check over source text would pass on
  // an indirection, and this one cannot.
  it('emits the module from committed data with no network at all', async () => {
    const { emitFontIndexModule } = await import('../scripts/build-font-index.mjs')
    const restore = globalThis.fetch
    globalThis.fetch = (() => { throw new Error('the build step must not reach the network') }) as never
    try {
      const emitted = emitFontIndexModule()
      expect(emitted.families).toBe(familyIndex.length)
      expect(emitted.snapshotDate).toBe(familyIndexSnapshotDate)
    } finally {
      globalThis.fetch = restore
    }
  })

  it('carries no licence field, because no licence is knowable before a pick', () => {
    const row = familyIndex[0] as unknown as Record<string, unknown>
    expect(Object.keys(row).sort()).toEqual(['category', 'family', 'popularity', 'scripts', 'variable'])
  })
})

describe('the local face tier', () => {
  // THE 21 COMMITTED FACES SURVIVE EPIC 16 UNCHANGED. `pickCatalogueFamily`
  // gained a source; it did not swap one.
  it('is the whole bundled catalogue, unchanged', () => {
    // THE POPULATION FLOOR — ONE OF THREE, AND ALL THREE MOVE TOGETHER.
    // The other two are `src/font-catalogue.test.ts` ("declares at least
    // twenty NEW families") and `src/font-name-table.test.ts` ("reads a
    // copyright out of every committed catalogue face"). Raised 20 -> 31 by
    // Story 16.1a, which added ten families to the local face tier.
    // D-16.R.12: "a floor left at 21 while the tier grows to 30 is a floor
    // that stops measuring the thing it was built to measure" — and
    // D-16.R.18's correction to it: a floor that exists in three files is
    // three floors, so a batch that raises one and leaves two behind is
    // silently unmeasured at the two it left.
    expect(catalogueFaces.length, 'the local face tier population floor; Story 16.1a raised it 20 -> 31').toBeGreaterThanOrEqual(31)
    expect(catalogueFaces.length).toBe(manifest.length)
    for (const face of catalogueFaces) expect(localTierHolds(face.family)).toBe(true)
  })

  // JOIN KEY: EXACT `family` STRING EQUALITY. No case-folding, no whitespace
  // normalisation, no fuzzy match. `Geist` / `Geist Mono` / `Geist Pixel` is
  // exactly the neighbourhood a loose matcher gets wrong.
  it('joins on exact string equality and on nothing looser', () => {
    for (const face of catalogueFaces) {
      expect(localTierHolds(face.family.toLowerCase()), `${face.family} must not join case-insensitively`).toBe(face.family === face.family.toLowerCase())
      expect(localTierHolds(` ${face.family}`)).toBe(false)
      expect(localTierHolds(face.family.replace(/ /g, ''))).toBe(face.family.indexOf(' ') === -1)
    }
  })

  // A LOCAL FACE WITH NO INDEX ROW IS LOCAL-TIER-ONLY, AND THAT IS CORRECT
  // BEHAVIOUR, NOT A DEFECT. Measured: `Inter Display` and `Source Serif 4
  // Display` have no index row at all, so 2 of 21 are unjoinable under any
  // normalisation.
  it('still offers a local face the published index has never heard of', () => {
    const indexed = new Set(familyIndex.map((row) => row.family))
    const unjoinable = catalogueFaces.filter((face) => !indexed.has(face.family))
    expect(unjoinable.length, 'this assertion is only meaningful while some local face is absent from the index').toBeGreaterThan(0)
    for (const face of unjoinable) {
      expect(offeredFamilies(face.family).some((source) => source.tier === 'local' && source.family === face.family)).toBe(true)
      expect(webFamilies.some((row) => row.family === face.family)).toBe(false)
    }
  })

  // "VARIABLE-ONLY" IS A PROPERTY OF THE BYTE SOURCE, NOT OF THE FAMILY
  // (D-16.R.2a). `Roboto` and `Inter` are committed here as byte-for-byte
  // upstream STATICS and appear among the mirror's variable-only rows only
  // because `google/fonts` carries VF-only builds of them. The index's `axes`
  // field is not consulted for a family the local tier holds.
  it('offers a locally-held family from the local tier even when the index calls it variable-only', () => {
    const variableUpstream = catalogueFaces.filter((face) => familyIndex.some((row) => row.family === face.family && row.variable))
    expect(variableUpstream.map((face) => face.family), 'the measured cases this rule exists for').toEqual(expect.arrayContaining(['Roboto', 'Inter', ...batchFamilies]))
    for (const face of variableUpstream) {
      const offered = offeredFamilies(face.family).filter((source) => source.family === face.family)
      expect(offered, `${face.family} must be offered exactly once`).toHaveLength(1)
      expect(offered[0].tier).toBe('local')
    }
  })

  // STORY 16.1a — THE BATCH, NAMED, AND WHAT ADDING IT DID TO THE COUNT.
  //
  // Every one of these ten is variable-only on the `google/fonts` mirror and was
  // therefore already filtered OUT of `webFamilies` before this story; each one's
  // own project publishes an ordinary static Regular, which is what the local
  // tier now carries. So ADDING A FAMILY LOCALLY IS WHAT MAKES IT OFFERED, and
  // the addable count rises by exactly the batch size rather than by some
  // number that depends on what the mirror happened to hold.
  it('offers every family the batch added, exactly once and from the local tier', () => {
    expect(batchFamilies.length, 'the batch list is empty, so every assertion over it is vacuous').toBe(10)
    for (const family of batchFamilies) {
      expect(localTierHolds(family), `${family} was added by Story 16.1a and the local tier does not hold it`).toBe(true)
      const offered = offeredFamilies(family).filter((source) => source.family === family)
      expect(offered, `${family} must be offered exactly once`).toHaveLength(1)
      expect(offered[0].tier, `${family} must be offered from the local tier, with no fetch`).toBe('local')
      expect(webFamilies.some((row) => row.family === family), `${family} must not also be offered from the web tier`).toBe(false)
    }
  })

  // AND THE COUNT ROSE BY EXACTLY THE BATCH SIZE. Recomputed from the index and
  // from the catalogue MINUS the batch, rather than compared against a number
  // typed in from a previous run — a hardcoded "before" is a second authority
  // that ages, and this one cannot.
  it('raised the addable count by exactly the batch size', () => {
    const beforeLocal = new Set(catalogueFaces.map((face) => face.family).filter((family) => !batchFamilies.includes(family)))
    expect(beforeLocal.size, 'the pre-batch tier is the catalogue minus the ten').toBe(catalogueFaces.length - batchFamilies.length)
    const beforeWeb = familyIndex.filter((row) => !row.variable && !beforeLocal.has(row.family))
    const beforeAddable = beforeWeb.length + beforeLocal.size
    expect(
      addableFamilyCount - beforeAddable,
      'the batch adds ten families the web tier could not offer, so the addable count must rise by exactly ten. A batch family that were statically published on the mirror would have been addable already, and the delta would be short by one.',
    ).toBe(batchFamilies.length)
  })

  it('offers a family the local tier holds exactly once, never from both tiers', () => {
    const offered = offeredFamilies('').map((source) => source.family)
    expect(new Set(offered).size).toBe(offered.length)
    for (const face of catalogueFaces) expect(webFamilies.some((row) => row.family === face.family)).toBe(false)
  })
})

describe('what the browser shows and what it says about it', () => {
  // A FAMILY THAT CANNOT BE ADDED IS FILTERED OUT, NOT LISTED AND REFUSED
  // (D-16.R.2, owner). Measured: 37 of the 50 most popular families are
  // variable-only, so listing and refusing means the most common first action
  // in the product fails.
  it('filters out variable-only families the local tier does not hold', () => {
    const variable = familyIndex.filter((row) => row.variable)
    expect(variable.length, 'about a quarter of the library ships variable-only').toBeGreaterThan(300)
    for (const row of webFamilies) expect(row.variable, `${row.family} is variable-only and must not be listed`).toBe(false)
    // AND THE HIDDEN ROWS ARE HIDDEN, not merely marked.
    expect(offeredFamilies('').some((source) => source.tier === 'web' && source.row.variable)).toBe(false)
  })

  it('reports the ADDABLE count, and says which count it is', () => {
    expect(addableFamilyCount).toBe(webFamilies.length + catalogueFaces.length)
    expect(addableFamilyCount).toBeLessThan(familyIndexPublishedFamilies)
    const disclosure = familyIndexDisclosure()
    expect(disclosure).toContain(`${addableFamilyCount} families you can add`)
    expect(disclosure).not.toContain(String(familyIndexPublishedFamilies))
    expect(disclosure).toContain(`${catalogueFaces.length} already on this machine`)
  })

  // THE WORD "LIVE" IS QUALIFIED WHEREVER IT WOULD OTHERWISE BE READ IN. The
  // LIST is a build-time snapshot with a date; only the typeface is fetched.
  it('says the list is a dated snapshot and only the typeface is fetched', () => {
    const disclosure = familyIndexDisclosure()
    expect(disclosure).toContain(`snapshot taken on ${familyIndexSnapshotDate}`)
    expect(disclosure).toMatch(/changes only when the designer is released/)
    expect(disclosure).toMatch(/fetched at the moment you pick one/)
    expect(disclosure).toMatch(/single variable file are not shown/)
  })

  it('searches both tiers, local first, and matches on substring', () => {
    const results = offeredFamilies('kanit')
    expect(results.some((source) => source.family === 'Kanit')).toBe(true)
    const all = offeredFamilies('')
    const firstWeb = all.findIndex((source) => source.tier === 'web')
    const lastLocal = all.map((source) => source.tier).lastIndexOf('local')
    expect(lastLocal).toBeLessThan(firstWeb)
  })
})
