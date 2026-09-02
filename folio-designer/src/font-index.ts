import { catalogueFaces, type CatalogueFace } from './generated/font-catalogue'
import { familyIndex, familyIndexExcludedCjkFamilies, familyIndexPublishedFamilies, familyIndexSnapshotDate, type IndexFamily } from './generated/font-index'

// THE TWO TIERS, AND THE JOIN BETWEEN THEM (D-16.R.3).
//
// THE BUNDLED CATALOGUE DID NOT GO AWAY. The 21 committed faces survive Epic 16
// unchanged, as the LOCAL FACE TIER — `font-catalogue.json`, the per-face
// `LICENSE*` and `NOTICE.md` beside each binary, and the build-time gate over
// all of it. The pick GAINS A SOURCE; it does not swap one.
//
// (It is the "local face tier" and not the "derived-static tier". Measured:
// `tools/fontgen/instance_faces.py` drives a hardcoded three-entry list of
// ENGINE faces and none of the 21 designer faces is derived — every `NOTICE.md`
// says NO DERIVATION APPLIES. Calling them derived would carry an error into the
// name.)
//
// LOCAL WINS, WITH NO FETCH AT ALL — no `METADATA.pb`, no licence file, no
// bytes. The committed bytes carry a STRONGER record than any fetch can produce:
// a reviewed licence identifier, the upstream licence file committed beside the
// binary, and a provenance note. Preferring a fetch would replace a verified
// record with an unverified one.
//
// DIVERGENCE IS DELIBERATELY NOT RECONCILED. Under AD-8 and D-16.2 a face is
// identified by the SHA-256 of its bytes, so "upstream released a newer version"
// is a DIFFERENT FACE, not a newer one. There is no staleness check, no update
// prompt and no version compare in this epic; the deferral is registered in
// `deferred-work.md` with its trigger.

export type FamilySource =
  /** A face this machine already holds. Picking it fetches nothing. */
  | Readonly<{ tier: 'local'; family: string; face: CatalogueFace }>
  /** A family from the build-time index snapshot. Picking it fetches. */
  | Readonly<{ tier: 'web'; family: string; row: IndexFamily }>

/** Restated here so the UI can say how old the list is without importing the generated module. */
export const indexSnapshotDate = familyIndexSnapshotDate
export const indexPublishedFamilies = familyIndexPublishedFamilies
export const indexExcludedCjkFamilies = familyIndexExcludedCjkFamilies

/**
 * THE JOIN KEY IS EXACT `family` STRING EQUALITY. No case-folding, no whitespace
 * normalisation, no fuzzy match.
 *
 * Measured ground for why looseness is refused rather than merely unnecessary:
 * `Inter Display` and `Source Serif 4 Display` have NO INDEX ROW AT ALL, so 2 of
 * the 21 are unjoinable under any normalisation — while `Geist` / `Geist Mono` /
 * `Geist Pixel` is exactly the neighbourhood a loose matcher gets wrong. A LOCAL
 * FACE WITH NO INDEX ROW IS LOCAL-TIER-ONLY, AND THAT IS CORRECT BEHAVIOUR, NOT
 * A DEFECT.
 */
const localByFamily: ReadonlyMap<string, CatalogueFace> = new Map(catalogueFaces.map((face) => [face.family, face]))

export const localTierHolds = (family: string): boolean => localByFamily.has(family)

/**
 * `axes` IS A PREDICTION, AND THIS IS THE ONLY PLACE IT IS CONSULTED.
 *
 * About a quarter of the published library ships as a single variable file
 * holding every weight at once. This product does not accept one: accepting it
 * would mean guessing which weight the author meant, and would make the same
 * template print differently on different machines. Browser-side instancing is
 * refused by ruling (D-16.5(c)) — it makes the embedded face a function of the
 * author's runtime — and Folio has no backend, so there is no third place to do
 * it.
 *
 * Measured: all 558 axes-declaring families still list a `400` key under
 * `fonts`, so the offered-weights map cannot answer this and `axes != []` is the
 * only signal the snapshot carries. It is a good heuristic, verified on Roboto
 * and six others, AND IT IS A HEURISTIC. The authority stays Go: the engine
 * refuses a variable face if one ever reaches it, and hiding a row here does not
 * relax that by one byte.
 *
 * IT IS NOT CONSULTED FOR A FAMILY THE LOCAL TIER HOLDS. "Variable-only" is a
 * property of the BYTE SOURCE, not of the family (D-16.R.2a): `Roboto` and
 * `Inter` are committed here as byte-for-byte upstream STATICS, from
 * `googlefonts/roboto-classic` and `rsms/inter` v4.1, and appear among the
 * variable-only rows only because the `google/fonts` mirror carries VF-only
 * builds of them. A family in the local tier is offered from the local tier and
 * the index's opinion about it is never consulted.
 */
const addableFromTheWeb = (row: IndexFamily): boolean => !row.variable

/**
 * A FAMILY THAT CANNOT BE ADDED IS FILTERED OUT, NOT LISTED AND REFUSED
 * (D-16.R.2, owner).
 *
 * The refusal was priced as a long tail and it is not one: measured, 37 of the
 * 50 most popular families are variable-only — Roboto, Open Sans, Inter,
 * Montserrat, Raleway, Nunito, Oswald, Playfair Display. Listing them and
 * refusing means the most common first action in the product fails. A row the
 * author cannot act on is a row that should not be there.
 *
 * A HIDDEN ROW IS A PRESENTATION CHOICE, NEVER A GUARD. The engine's refusal
 * stays, for anything that reaches it by any other door.
 */
export const webFamilies: ReadonlyArray<IndexFamily> = familyIndex.filter((row) => addableFromTheWeb(row) && !localTierHolds(row.family))

/**
 * THE COUNT THE BROWSER REPORTS IS THE ADDABLE COUNT, and the caller is expected
 * to say which it is. It is not 1,946: that is how many families the source
 * published on the snapshot date, and it is carried separately
 * (`indexPublishedFamilies`) so the two can never be confused for each other.
 */
export const addableFamilyCount = webFamilies.length + catalogueFaces.length

/**
 * ONE ORDERED LIST OF EVERY FAMILY THE AUTHOR MAY PICK, local tier first.
 *
 * Local first is the honest order rather than a preference: those rows need no
 * network, and the join above has already removed their web duplicates, so a
 * family present in both appears once, from the tier that can serve it offline.
 */
export function offeredFamilies(query: string): ReadonlyArray<FamilySource> {
  const needle = query.trim().toLowerCase()
  const hit = (family: string) => needle === '' || family.toLowerCase().includes(needle)
  const local: ReadonlyArray<FamilySource> = catalogueFaces.filter((face) => hit(face.family)).map((face) => ({ tier: 'local', family: face.family, face }))
  const web: ReadonlyArray<FamilySource> = webFamilies.filter((row) => hit(row.family)).map((row) => ({ tier: 'web', family: row.family, row }))
  return [...local, ...web]
}

/**
 * THE SENTENCE THE BROWSER SHOWS ABOUT ITS OWN LIST.
 *
 * It says the count is the addable one, and it qualifies the word "live": the
 * LIST is a build-time snapshot with a date, because the endpoint that publishes
 * it sends no CORS header and a browser cannot read it. Only the typeface is
 * fetched at the moment of a pick. A product that let this read as "live font
 * browser" would be saying something untrue.
 */
export function familyIndexDisclosure(): string {
  return `${addableFamilyCount} families you can add — ${catalogueFaces.length} already on this machine, ${webFamilies.length} downloaded when you pick them. `
    + `The list itself is a snapshot taken on ${indexSnapshotDate} and ships with this designer, so it changes only when the designer is released; the typefaces are fetched at the moment you pick one. `
    + `Families published only as a single variable file are not shown, because this product embeds one static weight.`
}
