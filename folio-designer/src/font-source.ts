import { classifyLicenceToken, type LicenceClassification } from './font-licence.ts'
import { faceCopyright } from './font-name-table.ts'

// THE FONT SOURCE — THE ONE MODULE IN THIS REPOSITORY THAT NAMES A FONT HOST.
//
// Every host this designer is allowed to reach is spelled once, below, on a line
// carrying the `folio:font-host-declaration` marker IN CODE. That is not a
// formality: `scripts/forbidden-font-hosts.mjs` fails the build on either host
// appearing anywhere else under its scanned roots, and it computes the exemption
// over comment-blanked source, so the marker cannot be written in a comment.
// Anyone reaching for a second fetch site has to either put it here or break the
// build.
//
// WHY THE REPOSITORY HOST AND NEVER THE STYLESHEET ENDPOINT (D-16.3, measured
// 2026-09-02; the two hosts are named in the array below, and deliberately
// nowhere else in this file — the source scan reads RAW text, so a host spelled
// in prose here would be an undeclared occurrence like any other).
// The `css2` endpoint under a modern browser
// user-agent returns `woff2`, which the engine's `decodeRecognisedFont` refuses
// by design — its accepted media types are exactly `font/ttf` and `font/otf`,
// with `font/woff2` deliberately outside the set — and it returns it SPLIT BY
// `unicode-range` INTO PER-SCRIPT SUBSETS, which would embed partial coverage
// into a document naming the whole family. The full TTF that endpoint serves to
// a legacy user-agent is unreachable from a browser, which cannot set
// `User-Agent`. The `css2` host is forbidden outright by the scan's first half
// and that is what keeps this trap shut.
//
// WHAT IS FETCHED, AND IN WHAT ORDER: `METADATA.pb` first, because it decides
// admission; then, only if the terms are admitted, the licence file and the
// bytes. THE ORDER IS NON-NEGOTIABLE — classify, then embed. A face may not
// reach the document before its licence and copyright are in hand.

/**
 * THE HOSTS. Declared here, in code, with the marker, and nowhere else.
 *
 * `fontsRepositoryHost` serves the bytes, the licence text and the per-family
 * metadata, with `access-control-allow-origin: *`, so a browser may read it.
 *
 * `familyIndexHost` serves the family list and CANNOT be read by a browser at
 * all — it sends no `access-control-allow-origin` (D-16.3, measured). It is
 * named here only because `scripts/build-font-index.mjs` snapshots it at build
 * time; nothing at runtime reaches for it, which is exactly why the word "live"
 * is qualified everywhere it appears in this product.
 */
export const fontHostDeclarations = [
  { host: 'raw.githubusercontent.com', declaration: 'folio:font-host-declaration', role: 'face bytes, licence text and per-family metadata, read at the moment a family is picked' },
  { host: 'fonts.google.com', declaration: 'folio:font-host-declaration', role: 'the family list, read ONLY by scripts/build-font-index.mjs at build time; unreadable from a browser' },
] as const

export const fontsRepositoryHost = fontHostDeclarations[0].host
export const familyIndexHost = fontHostDeclarations[1].host

/** The branch of `google/fonts` the snapshot and the fetches both read. */
const fontsRepositoryBase = `https://${fontsRepositoryHost}/google/fonts/main`

/**
 * THE PROBE ORDER (D-16.R.6). The index carries no path field, so the directory
 * is derived and then confirmed; these are the four top-level directories
 * `google/fonts` publishes.
 *
 * THE DIRECTORY IS NEVER EVIDENCE OF THE TERMS. Measured: upstream MOVES
 * families between directories — `apache/roboto` now 404s and Roboto lives in
 * `ofl/` — so reading layout as a licence assertion would let a family that
 * moved silently change the terms a document publishes. `METADATA.pb` always
 * wins; a family resolved at `ofl/x` whose token says `APACHE2` is admitted as
 * `Apache-2.0`, and the divergence is RECORDED rather than refused.
 */
export const probeDirectories = ['ofl', 'apache', 'ufl', 'cc-by-sa'] as const

/**
 * THE LICENCE FILE IS NAMED BY THE DECLARED TERMS, NOT BY THE DIRECTORY, for
 * the same reason the directory is not evidence: a family resolved at `ofl/x`
 * that declares `APACHE2` publishes Apache terms, and carrying `OFL.txt` beside
 * them would make the document state terms its own record contradicts. If the
 * file the declared licence names is not there, the pick is REFUSED — a
 * document may not carry a face without its terms.
 *
 * THE MAP HOLDS EXACTLY THE IDS THE TOKEN TABLE CAN EMIT, AND NO OTHERS. It is
 * keyed on `classifyLicenceToken`'s admitted output, not on D-8.5.3's four
 * identifiers, and those are different sets on purpose: `font-licence.ts`
 * deliberately has no `MIT` row and argues at length that this is ABSENCE, NOT
 * NARROWING — `google/fonts` publishes no MIT token, so nothing here can ever
 * be asked for one. A speculative `MIT` row would be dead on arrival and, worse,
 * would be the mapping a future MIT token silently inherited without anybody
 * reviewing which file upstream actually publishes. If the token table ever
 * gains a row, this map gains one in the same change, and until then a missing
 * row is a stated refusal below rather than a URL ending in `undefined`.
 */
const licenceFileFor: Readonly<Record<string, string>> = {
  'OFL-1.1': 'OFL.txt',
  'Apache-2.0': 'LICENSE.txt',
  'Ubuntu-font-1.0': 'UFL.txt',
}

/**
 * THE SLUG RULE, EXACT (D-16.R.6): lowercase the family name, then delete every
 * character outside `[a-z0-9]`.
 *
 * Verified 8 of 8 on deliberately awkward families — `Press Start 2P` →
 * `pressstart2p`, `Baloo Bhai 2` → `baloobhai2`, `Alegreya SC` → `alegreyasc`,
 * `Source Serif 4` → `sourceserif4`, `DM Sans` → `dmsans`, `Ma Shan Zheng` →
 * `mashanzheng`, `Playpen Sans Thai` → `playpensansthai`, `Noto Sans Thai
 * Looped` → `notosansthailooped`.
 *
 * This is a derivation CLOSED BY VERIFICATION, which is why it is not the guess
 * this module forbids one level down for the Regular filename: the directory it
 * proposes is accepted only if that directory's `METADATA.pb` `name` string-
 * equals the family the author picked.
 */
export function familyDirectorySlug(family: string): string {
  let slug = ''
  for (const character of family.toLowerCase()) if (character >= 'a' && character <= 'z' || character >= '0' && character <= '9') slug += character
  return slug
}

export type FamilyMetadata = Readonly<{
  /** The family name `METADATA.pb` itself declares. The confirmation compares against this. */
  name: string
  /** The upstream licence token — `OFL`, `APACHE2`, `UFL`, … — never an SPDX id. */
  licence: string
  /** Every `fonts { … }` block, in file order. */
  faces: ReadonlyArray<Readonly<{ style: string; weight: number; filename: string }>>
}>

/**
 * A READER FOR `METADATA.pb`'s TEXT PROTO, and only for the four things this
 * story asks it.
 *
 * DEPTH IS TRACKED RATHER THAN IGNORED, because `name:` appears BOTH at the top
 * level (the family) and inside every `fonts { … }` block (the face). A flat
 * scan would read the last face's name as the family's and confirm a directory
 * against the wrong string — the exact failure the confirmation exists to
 * prevent.
 *
 * No protobuf dependency: this reads four fields out of a line-oriented text
 * format, and `package.json`'s three dependencies are a standing decision.
 */
export function parseFamilyMetadata(source: string): FamilyMetadata | undefined {
  let name: string | undefined
  let licence: string | undefined
  const faces: Array<{ style: string; weight: number; filename: string }> = []
  let block: { style: string; weight: number; filename: string } | undefined
  let depth = 0
  for (const rawLine of source.split('\n')) {
    const line = rawLine.trim()
    if (line === '' || line.startsWith('#')) continue
    if (line === '}') {
      // DEPTH HAS A FLOOR, AND THE FLOOR IS WHAT KEEPS A MALFORMED FILE FROM
      // RESOLVING TO THE WRONG STRING. Unfloored, one stray `}` drives the
      // depth negative, the next `{` returns it to 0 without opening a block,
      // and the `name:` inside a `fonts { … }` entry — upstream blocks really
      // do carry one — is then read as the FAMILY name and confirmed against.
      // That is precisely the confusion the name-equality confirmation exists
      // to prevent. Floored, an unbalanced file can only fail to resolve.
      depth = Math.max(0, depth - 1)
      if (depth === 0 && block !== undefined) {
        if (block.filename !== '') faces.push(block)
        block = undefined
      }
      continue
    }
    if (line.endsWith('{')) {
      const opened = line.slice(0, -1).trim()
      if (depth === 0 && opened === 'fonts') block = { style: '', weight: 0, filename: '' }
      depth += 1
      continue
    }
    const colon = line.indexOf(':')
    if (colon === -1) continue
    const key = line.slice(0, colon).trim()
    const raw = line.slice(colon + 1).trim()
    const value = raw.startsWith('"') && raw.endsWith('"') && raw.length >= 2 ? raw.slice(1, -1) : raw
    if (depth === 0) {
      if (key === 'name') name ??= value
      if (key === 'license') licence ??= value
      continue
    }
    if (depth === 1 && block !== undefined) {
      if (key === 'style') block.style = value
      if (key === 'weight') block.weight = Number(value)
      if (key === 'filename') block.filename = value
    }
  }
  if (name === undefined || licence === undefined) return undefined
  return { name, licence, faces }
}

/**
 * WHICH FILE IS REGULAR IS READ, NOT CONSTRUCTED.
 *
 * It is the `fonts { style: "normal", weight: 400 }` entry's own `filename`. A
 * filename assembled from the family name — `${family}-Regular.ttf` — is a guess,
 * and it is wrong often enough that the whole pick would fail on families whose
 * files are named for something other than their display name.
 */
export function regularFilename(metadata: FamilyMetadata): string | undefined {
  return metadata.faces.find((face) => face.style === 'normal' && face.weight === 400)?.filename
}

/**
 * THE MEDIA TYPE IS READ FROM THE FILE THE METADATA NAMES, and it is one of the
 * two the engine accepts or the pick is refused here rather than at the
 * boundary. This is also the second place the `woff2` route is shut: a `.woff2`
 * filename has no media type in this table.
 */
const mediaTypes: Readonly<Record<string, string>> = { '.ttf': 'font/ttf', '.otf': 'font/otf' }
const mediaTypeOf = (filename: string): string | undefined => {
  const dot = filename.lastIndexOf('.')
  const extension = dot === -1 ? '' : filename.slice(dot).toLowerCase()
  return Object.hasOwn(mediaTypes, extension) ? mediaTypes[extension] : undefined
}

/** Everything `embedFontFamilyCommand` requires of a fetched face, and the divergence note. */
export type FetchedFace = Readonly<{
  family: string
  style: string
  licence: string
  licenceText: string
  copyright: string
  source: string
  mediaType: string
  bytes: ArrayBuffer
  /** Set when the resolved directory disagrees with the declared token. An observation, never a refusal. */
  layoutDivergence?: string
}>

export type FetchOutcome =
  | Readonly<{ ok: true; face: FetchedFace }>
  | Readonly<{ ok: false; reason: string; classification?: LicenceClassification }>

type Fetcher = (url: string) => Promise<Response>

const refuse = (reason: string, classification?: LicenceClassification): FetchOutcome => ({ ok: false, reason, classification })

/**
 * ONE PICK, ONE RESOLUTION. Probing is once per pick — never at index render and
 * never on a keystroke: four probes across 1,946 families must not become a
 * browsing cost.
 */
export async function fetchWebFamily(family: string, fetcher: Fetcher = (url) => fetch(url)): Promise<FetchOutcome> {
  const slug = familyDirectorySlug(family)
  if (slug === '') return refuse(`${family} has no directory this designer can derive from its name`)

  let directory: string | undefined
  let metadata: FamilyMetadata | undefined
  let sawSomething = false
  for (const candidate of probeDirectories) {
    let response: Response
    try {
      response = await fetcher(`${fontsRepositoryBase}/${candidate}/${slug}/METADATA.pb`)
    } catch (error) {
      return refuse(`${family} could not be reached right now (${error instanceof Error ? error.message : String(error)}). You cannot add a family without a network connection; the faces this machine already holds are still offered.`)
    }
    if (response.status === 404) continue
    if (!response.ok) return refuse(`${family}'s upstream metadata responded ${response.status}`)
    sawSomething = true
    const parsed = parseFamilyMetadata(await response.text())
    if (parsed === undefined) return refuse(`${family}'s upstream METADATA.pb could not be read`)
    // THE CONFIRMATION. A mismatch is a REFUSAL, never a fallback to the next
    // directory: the slug is a derivation and this is the check that closes it,
    // so continuing past a disagreement would turn "derived then confirmed"
    // back into the guess it exists to replace.
    if (parsed.name !== family) {
      return refuse(`${family} does not match the upstream directory ${candidate}/${slug}, which publishes "${parsed.name}". The directory is derived from the family name and confirmed by the family's own metadata, and a disagreement is refused rather than guessed past.`)
    }
    directory = candidate
    metadata = parsed
    break
  }
  if (directory === undefined || metadata === undefined) {
    return refuse(sawSomething
      ? `${family} could not be resolved upstream`
      : `${family} is in this designer's snapshot of the family list but is no longer published upstream. The list ships with the designer and ages between releases, so it can name a family that has since been renamed or withdrawn.`)
  }

  // CLASSIFY, THEN EMBED. Nothing below this line is fetched until the terms are
  // admitted, and no byte reaches the document before its licence and copyright
  // are in hand.
  const classification = classifyLicenceToken(metadata.licence)
  if (classification.state !== 'admitted') return refuse(`${family} cannot be added: ${classification.reason}`, classification)

  const filename = regularFilename(metadata)
  if (filename === undefined) {
    return refuse(`${family} publishes no upright Regular (a static face at weight 400) upstream, so there is no single face to embed`)
  }
  const mediaType = mediaTypeOf(filename)
  if (mediaType === undefined) return refuse(`${family}'s Regular is published as ${filename}, which is not a font file this engine reads`)

  // A MISSING ROW IS A STATED REFUSAL, NEVER A MALFORMED FETCH. Unguarded, this
  // lookup would build a URL ending in `undefined`, fetch it, and refuse the
  // family by naming a licence file that does not exist anywhere — a message
  // that sends the reader upstream to look for a file nobody ever published.
  // The admitted set and this map are meant to be the same set; if they ever
  // diverge, this says so in those words.
  if (!Object.hasOwn(licenceFileFor, classification.spdx)) {
    return refuse(`${family} declares ${classification.spdx}, which this designer admits but has no licence file name for, so its terms cannot be fetched to travel with it`, classification)
  }
  const licenceFile = licenceFileFor[classification.spdx]
  const licenceText = await readText(fetcher, `${fontsRepositoryBase}/${directory}/${slug}/${licenceFile}`)
  if (licenceText === undefined || licenceText.trim() === '') {
    return refuse(`${family} declares ${classification.spdx} but publishes no ${licenceFile} beside its face, so its terms cannot travel with it. A document may not carry a face without the text of its licence.`)
  }

  let bytes: ArrayBuffer
  try {
    const response = await fetcher(`${fontsRepositoryBase}/${directory}/${slug}/${filename}`)
    if (!response.ok) return refuse(`${family}'s face ${filename} responded ${response.status}`)
    bytes = await response.arrayBuffer()
  } catch (error) {
    return refuse(`${family} could not be fetched right now (${error instanceof Error ? error.message : String(error)}). You cannot add a family without a network connection; the faces this machine already holds are still offered.`)
  }

  // nameID 0, FROM THE BYTES THAT ARE ABOUT TO BE EMBEDDED. Absent is a refusal,
  // because the engine refuses to load a document that embeds a face with no
  // copyright, so admitting one here would write a file this product cannot open.
  let copyright: string
  try {
    copyright = faceCopyright(bytes)
  } catch (error) {
    return refuse(`${family} cannot be added: ${error instanceof Error ? error.message : String(error)}`)
  }

  // LAYOUT DISAGREEMENT IS AN OBSERVATION, NOT A REFUSAL. Recorded because
  // systematically it means the probe order is costing round-trips.
  const expected = { 'OFL-1.1': 'ofl', 'Apache-2.0': 'apache', 'Ubuntu-font-1.0': 'ufl' }[classification.spdx]
  const layoutDivergence = expected !== undefined && expected !== directory
    ? `${family} is published under ${directory}/ while its own metadata declares ${classification.token} (${classification.spdx}); the metadata is the authority on the terms and the directory is only where the files sit`
    : undefined

  return {
    ok: true,
    face: {
      family,
      style: 'Regular',
      licence: classification.spdx,
      licenceText,
      copyright,
      source: `${fontsRepositoryHost}/google/fonts/main/${directory}/${slug}/${filename}`,
      mediaType,
      bytes,
      layoutDivergence,
    },
  }
}

async function readText(fetcher: Fetcher, url: string): Promise<string | undefined> {
  try {
    const response = await fetcher(url)
    if (!response.ok) return undefined
    return await response.text()
  } catch {
    return undefined
  }
}
