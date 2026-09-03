import { catalogueFaces, type CatalogueFace, type CatalogueScript } from './generated/font-catalogue'
import { familyIndex, familyIndexExcludedCjkFamilies, familyIndexPublishedFamilies, familyIndexSnapshotDate, type IndexFamily } from './generated/font-index'
import type { StoredFace } from './font-store'

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

// AND A THIRD TIER SINCE STORY 16.2 (D-16.R.33 R1): THE MACHINE STORE.
//
// A face this designer has fetched BEFORE, kept in origin-scoped browser storage
// under the SHA-256 of its bytes. It is a source exactly as the other two are:
// picking it fetches nothing, works with the network down, and embeds the same
// three-part licence record the fetch would have produced — because the store
// keeps that record beside the bytes.
//
// IT SITS BETWEEN THE TWO, AND THE ORDER IS THE HONEST ONE. The local tier
// wins over it: those 31 faces carry a REVIEWED licence identifier, the
// upstream licence file committed beside the binary, and a build-time gate over
// all of it — a stronger record than any fetch can produce, including the fetch
// that filled this store. The store wins over the web tier for the plain reason
// that it is the same bytes without the round-trip.
//
// THE SEAM IS BUILT HERE, NOT LEFT TO STORY 16.4. 16.4 adds the headings that
// GROUP these three; it does not reshape the union. Every exhaustive switch
// over `FamilySource` reds until the new arm is handled, which is what makes
// the hand-off a mechanism rather than a sentence in a spec.
//
// `AVAILABLE LOCALLY` IS FETCHED FACES — THIS ARM. The frozen intent contract
// says so in those words ("`AVAILABLE LOCALLY` is fetched faces, never host
// fonts", D-16.2), the story's plain-terms opener says it means "typefaces this
// designer has fetched before", and the panel heading `App.tsx` renders says
// TYPEFACES THIS DESIGNER HAS DOWNLOADED. All three agree, and this comment
// agrees with them: it does NOT mean this arm plus the local arm.
//
// WHAT THE DROPDOWN GROUP OF THAT NAME ENDS UP CONTAINING IS STORY 16.4'S TO
// DECIDE, AND IS NOT DECIDED HERE. 16.4 adds the headings that group this
// union; 16.2 builds the union and no group at all. Whether the 31 shipped
// local-tier faces are shown under that heading, under their own, or under
// neither is a presentation question this module has no opinion on — it joins
// three lists and returns them in one order.
//
// WHAT IS SETTLED, IN EVERY READING: it never means "the fonts installed on
// this computer". SPEC-fonts' *"No host fonts"* Non-goal is the one clause
// D-16.1 left standing and it is untouched: the Local Font Access API is not
// used, referenced or feature-detected anywhere in this designer, and
// `src/host-font-access.test.ts` is the tripwire.
export type FamilySource =
  /** A face this machine already holds. Picking it fetches nothing. */
  | Readonly<{ tier: 'local'; family: string; face: CatalogueFace }>
  /** A face this designer fetched before and kept. Picking it fetches nothing. */
  | Readonly<{ tier: 'stored'; family: string; record: StoredFace }>
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
 * ONE ORDERED LIST OF EVERY FAMILY THE AUTHOR MAY PICK, local tier first, then
 * the faces this machine already holds, then the rest of the snapshot.
 *
 * Local first is the honest order rather than a preference: those rows need no
 * network, and the join above has already removed their web duplicates, so a
 * family present in both appears once, from the tier that can serve it offline.
 * The stored tier extends exactly that reasoning to a face the author fetched
 * last week.
 *
 * A STORED FAMILY REPLACES ITS WEB ROW; IT DOES NOT SIT BESIDE IT. One family,
 * one row, from the cheapest tier that can serve it — which is the same rule
 * `webFamilies` already applies to the local tier. A row offered twice, once as
 * "already here" and once as "will be downloaded", would make the author choose
 * between two spellings of one thing.
 *
 * A STORED FAMILY WITH NO WEB ROW IS STILL OFFERED. The index is a build-time
 * snapshot that ages, so a family fetched under one release can be withdrawn or
 * renamed upstream before the next. Its bytes are here, its licence record is
 * here, and refusing to offer it because a dated list no longer mentions it
 * would be the store failing at the one job it exists for.
 *
 * THE STORE'S LISTING IS PASSED IN, NEVER READ FROM HERE. The store's reads are
 * asynchronous and this function is called on every keystroke of a combobox.
 * The caller owns the read, its lifetime and its degradation; this module
 * remains a pure join over three inputs, which is also what keeps it testable
 * without a database.
 */
/**
 * WHICH OF TWO STORED FACES OF ONE FAMILY IS OFFERED — A RULE, NOT AN ACCIDENT.
 *
 * The store is keyed by the SHA-256 of the bytes, so ONE FAMILY CAN HONESTLY
 * HAVE MORE THAN ONE ENTRY: upstream re-cut the face between two fetches, or
 * the author has a Regular and an Italic of it. Under AD-8 those are DIFFERENT
 * FACES, not versions of one, and this function is `offeredFamilies` — one row
 * per family — so exactly one of them is offered.
 *
 * IT USED TO BE WHICHEVER ONE CAME LAST IN `list()` ORDER, which is the store's
 * family-then-key sort — that is, the hash, that is, arbitrary. A menu that
 * silently hands the author one of two faces on the strength of a digest
 * ordering is the "silent substitution" the contract forbids in the very clause
 * that makes the key a content hash.
 *
 * THE RULE IS: THE MOST RECENTLY FETCHED WINS. `fetchedAt` is `YYYY-MM-DD`, so
 * a plain string comparison is chronological. It is the right default because
 * the newer entry is the one the author's last pick of that family actually
 * produced, so it is the one they last saw work.
 *
 * TIES GO TO THE LEXICOGRAPHICALLY SMALLER KEY. Two faces fetched on the same
 * day are common (a Regular and an Italic within one session), and the tie must
 * break on something stable rather than on arrival order. The smaller key is
 * also the one `font-store.ts`'s `list()` sorts FIRST within a family, so what
 * the menu offers and what the listing shows first are the same face rather
 * than two answers to one question.
 *
 * NEITHER OF THESE IS A DROPDOWN GROUP, AND 16.2 DOES NOT BUILD ONE. Offering
 * both styles as separate rows is Story 16.4's, which owns the grouping; this
 * story owes only a choice that is deterministic and stated.
 */
function mostRecentlyFetched(left: StoredFace, right: StoredFace): StoredFace {
  if (left.fetchedAt !== right.fetchedAt) return left.fetchedAt > right.fetchedAt ? left : right
  return left.key <= right.key ? left : right
}

export function offeredFamilies(query: string, stored: ReadonlyArray<StoredFace> = []): ReadonlyArray<FamilySource> {
  const needle = query.trim().toLowerCase()
  const hit = (family: string) => needle === '' || family.toLowerCase().includes(needle)
  // THE LOCAL TIER IS NOT DISPLACED BY THE STORE. Its record is the stronger
  // one (see the note on `FamilySource`), and it needs no network either, so
  // there is nothing to win by preferring a fetched copy of the same family.
  //
  // AND WHERE THE STORE HOLDS TWO FACES OF ONE FAMILY, WHICH ONE IS OFFERED IS
  // CHOSEN AND WRITTEN DOWN — see `mostRecentlyFetched`. It used to be whichever
  // `list()` happened to return last, which is the silent substitution the
  // content-address key exists to refuse.
  const storedByFamily = new Map<string, StoredFace>()
  for (const record of stored) {
    if (localTierHolds(record.family)) continue
    const held = storedByFamily.get(record.family)
    storedByFamily.set(record.family, held === undefined ? record : mostRecentlyFetched(held, record))
  }
  const local: ReadonlyArray<FamilySource> = catalogueFaces.filter((face) => hit(face.family)).map((face) => ({ tier: 'local', family: face.family, face }))
  const web: FamilySource[] = []
  const offeredFromStore = new Set<string>()
  for (const row of webFamilies) {
    if (!hit(row.family)) continue
    const record = storedByFamily.get(row.family)
    if (record === undefined) { web.push({ tier: 'web', family: row.family, row }); continue }
    offeredFromStore.add(row.family)
    web.push({ tier: 'stored', family: row.family, record })
  }
  const orphaned: ReadonlyArray<FamilySource> = [...storedByFamily.values()]
    .filter((record) => !offeredFromStore.has(record.family) && hit(record.family))
    .map((record) => ({ tier: 'stored', family: record.family, record }))
  // Stored-but-unlisted families come first among the non-local rows for the
  // same reason the local tier does: they are here, and the rows after them are
  // not.
  return [...local, ...orphaned, ...web]
}

/**
 * THE TIER A ROW IS OFFERED FROM, IN THE AUTHOR'S OWN TERMS — one exhaustive
 * switch over `FamilySource`, so the union cannot gain an arm that nothing
 * describes.
 *
 * This is the seam D-16.R.33 R1 asked to be built here rather than left to
 * Story 16.4: adding a fourth tier without handling it stops compiling at the
 * `never`, which is what makes the hand-off enforceable.
 */
export function familySourceNote(source: FamilySource): string {
  switch (source.tier) {
    case 'local': return ' — add to document, already on this machine'
    case 'stored': return ' — add to document, already downloaded to this machine'
    case 'web': return ' — add to document'
    default: {
      const unhandled: never = source
      // NO `JSON.stringify` HERE. `engine-ownership-contract.test.ts` keeps
      // JSON out of every module but the protocol envelopes and the three
      // command factories, and it is right to: this module joins three lists
      // and has no business serialising anything.
      throw new Error(`a FamilySource tier nothing describes reached the family control: ${String((unhandled as FamilySource).tier)}`)
    }
  }
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

/**
 * THE SNAPSHOT ROW FOR A FAMILY, WHATEVER TIER OFFERS IT (Story 16.3).
 *
 * `webFamilies` above is the ADDABLE-FROM-THE-WEB list: it has already dropped
 * the variable-only rows and every family the local tier holds. That is the
 * right list to OFFER from and the wrong one to DESCRIBE from, because the
 * facts a browser prints beside a family — its category, how popular it is,
 * which scripts it covers — are properties of the family and are carried by the
 * snapshot for local-tier families too.
 *
 * SO THIS READS THE WHOLE `familyIndex`, INCLUDING ROWS `webFamilies` REMOVED,
 * AND THAT IS NOT A HOLE IN THE FILTER. A row reachable here can never become a
 * pick: `offeredFamilies` is the only source of a `FamilySource`, nothing here
 * produces one, and a local-tier family is offered from the local tier whatever
 * the snapshot's `variable` flag says about the mirror's build of it (D-16.R.2a).
 * This function answers "what does the snapshot say about this name", never
 * "may the author add it".
 *
 * A FAMILY WITH NO ROW IS `undefined` AND THE CALLER MUST SAY SO IN WORDS. Two
 * of the local-tier families have no index row at all — the join note above
 * names them — so "no category" is a real and permanent state, not a loading
 * one, and printing a guessed category for those two would be the browser
 * inventing a fact about a typeface.
 */
const indexByFamily: ReadonlyMap<string, IndexFamily> = new Map(familyIndex.map((row) => [row.family, row]))

export function indexRowFor(family: string): IndexFamily | undefined {
  return indexByFamily.get(family)
}

/**
 * THE CATEGORY VOCABULARY, READ OFF THE SNAPSHOT RATHER THAN TYPED OUT.
 *
 * `Font Browser.dc.html` draws four category chips — Sans Serif, Serif, Display,
 * Monospace — over fourteen placeholder families. The snapshot carries FIVE
 * categories, and the fifth (Handwriting) is the third largest of them. A hand
 * copy of the mockup's four would hide 337 families behind chips that look
 * exhaustive, which is the failure mode a derived vocabulary cannot have.
 */
export const indexCategories: ReadonlyArray<string> = [...new Set(familyIndex.map((row) => row.category))].sort()

/**
 * AND THE WRITING-SYSTEM VOCABULARY, ON THE SAME GROUND.
 *
 * The mockup also draws Cyrillic and Greek chips. This snapshot records no such
 * coverage for any family — `CatalogueScript` is `latin`, `thai` and `cjk`, and
 * CJK is excluded from the snapshot by SPEC-fonts' own non-goal — so those two
 * chips would filter every result away every time they were pressed. A control
 * that can only ever empty the list is a false affordance, so the chips are the
 * scripts the snapshot and the local tier actually name.
 */
export const indexScripts: ReadonlyArray<CatalogueScript> = [...new Set([...familyIndex.flatMap((row) => row.scripts), ...catalogueFaces.flatMap((face) => face.scripts)])].sort()
