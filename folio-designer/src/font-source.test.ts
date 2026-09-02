import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { familyDirectorySlug, fetchWebFamily, fontHostDeclarations, parseFamilyMetadata, probeDirectories, regularFilename } from './font-source'
import { sfntWithNames } from './test/sfnt-fixture'
import { blankComments } from '../scripts/forbidden-font-hosts.mjs'

// STORY 16.1 — THE FETCH PATH AND ITS ADMISSION DECISION.
//
// EVERY TEST HERE DRIVES A STUB FETCHER, NOT THE NETWORK. What is being asserted
// is this module's own decisions — which directory it derives, what it accepts as
// confirmation, what it refuses and in which words, and above all THE ORDER:
// classify, then embed. A test that reached upstream would assert Google's
// current state rather than this module's rules, and would go red for reasons
// that are nobody's defect.

const here = path.dirname(fileURLToPath(import.meta.url))
// THE HOST IS SPELLED HERE WITH THE SCANNER'S MARKER, IN CODE, on this one
// line — the same shape `FORBIDDEN_FONT_HOSTS` uses. A marker in a comment
// would declare nothing: the scan reads RAW source while the exemption reads
// COMMENT-BLANKED source, deliberately.
const base = { url: 'https://raw.githubusercontent.com/google/fonts/main', declaration: 'folio:font-host-declaration' }.url

const kanitMetadata = `name: "Kanit"
designer: "Cadson Demak"
license: "OFL"
category: "SANS_SERIF"
date_added: "2015-11-04"
fonts {
  name: "Kanit"
  style: "normal"
  weight: 100
  filename: "Kanit-Thin.ttf"
  post_script_name: "Kanit-Thin"
}
fonts {
  name: "Kanit"
  style: "italic"
  weight: 400
  filename: "Kanit-Italic.ttf"
}
fonts {
  name: "Kanit"
  style: "normal"
  weight: 400
  filename: "Kanit-Regular.ttf"
  post_script_name: "Kanit-Regular"
}
subsets: "latin"
source {
  repository_url: "https://github.com/cadsondemak/Kanit"
  files {
    source_file: "Kanit-Regular.ttf"
    dest_file: "Kanit-Regular.ttf"
  }
}
`

const face = sfntWithNames([{ platform: 3, nameID: 0, value: 'Copyright 2020 The Kanit Project Authors' }])

type StubFile = Readonly<{ status?: number; body?: string | ArrayBuffer }>

/** A fetcher over a fixed map of paths, recording every URL it was asked for. */
function stub(files: Readonly<Record<string, StubFile>>) {
  const asked: string[] = []
  const fetcher = async (url: string) => {
    asked.push(url)
    const file = Object.hasOwn(files, url) ? files[url] : undefined
    if (file === undefined) return { ok: false, status: 404, text: async () => '', arrayBuffer: async () => new ArrayBuffer(0) } as unknown as Response
    const status = file.status ?? 200
    return {
      ok: status >= 200 && status < 300,
      status,
      text: async () => (typeof file.body === 'string' ? file.body : ''),
      arrayBuffer: async () => (typeof file.body === 'string' ? new ArrayBuffer(0) : file.body ?? new ArrayBuffer(0)),
    } as unknown as Response
  }
  return { fetcher, asked }
}

const kanitUpstream = (overrides: Readonly<Record<string, StubFile>> = {}) => ({
  [`${base}/ofl/kanit/METADATA.pb`]: { body: kanitMetadata },
  [`${base}/ofl/kanit/OFL.txt`]: { body: 'Copyright 2020 The Kanit Project Authors\n\nThis Font Software is licensed under the SIL Open Font License, Version 1.1.\n' },
  [`${base}/ofl/kanit/Kanit-Regular.ttf`]: { body: face },
  ...overrides,
})

describe('the family directory slug', () => {
  // D-16.R.6, verified 8 of 8 on deliberately awkward families. These are the
  // measured pairs, not invented ones: digits survive, spaces and punctuation
  // do not, and nothing is inserted.
  it('lowercases the family name and deletes every character outside [a-z0-9]', () => {
    for (const [family, slug] of [
      ['Press Start 2P', 'pressstart2p'],
      ['Baloo Bhai 2', 'baloobhai2'],
      ['Alegreya SC', 'alegreyasc'],
      ['Source Serif 4', 'sourceserif4'],
      ['DM Sans', 'dmsans'],
      ['Ma Shan Zheng', 'mashanzheng'],
      ['Playpen Sans Thai', 'playpensansthai'],
      ['Noto Sans Thai Looped', 'notosansthailooped'],
    ] as const) expect(familyDirectorySlug(family), family).toBe(slug)
  })

  it('deletes rather than transliterates, so a name with nothing to keep produces nothing', () => {
    expect(familyDirectorySlug('Röboto')).toBe('rboto')
    expect(familyDirectorySlug('— —')).toBe('')
  })

  it('probes the four upstream directories, in the order the ruling names', () => {
    expect([...probeDirectories]).toEqual(['ofl', 'apache', 'ufl', 'cc-by-sa'])
  })
})

describe('reading METADATA.pb', () => {
  // DEPTH IS THE WHOLE DIFFICULTY. `name:` appears at the top level (the
  // family) AND inside every `fonts { … }` block (the face) AND inside nested
  // `source { files { … } }` blocks. A flat scan would confirm the directory
  // against the wrong string, which is the exact failure the confirmation
  // exists to prevent.
  it('reads the FAMILY name from the top level and never a face name or a nested one', () => {
    const metadata = parseFamilyMetadata(kanitMetadata)
    expect(metadata?.name).toBe('Kanit')
    expect(metadata?.licence).toBe('OFL')
    expect(metadata?.faces).toHaveLength(3)
  })

  it('reads the Regular filename from the style:"normal" weight:400 entry rather than constructing one', () => {
    const metadata = parseFamilyMetadata(kanitMetadata)
    expect(regularFilename(metadata!)).toBe('Kanit-Regular.ttf')
    // AND IT IS NOT THE FIRST ENTRY, NOR THE ONE THE FAMILY NAME WOULD SUGGEST
    // IF THE FILES WERE NAMED DIFFERENTLY: the Thin comes first in the file and
    // the italic 400 sits between them.
    expect(metadata!.faces[0].filename).toBe('Kanit-Thin.ttf')
  })

  it('has no Regular to offer when upstream declares no upright 400', () => {
    const variableOnly = parseFamilyMetadata('name: "Anuphan"\nlicense: "OFL"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: ""\n}\n')
    expect(regularFilename(variableOnly!)).toBeUndefined()
  })

  it('returns nothing at all for a file that declares neither a name nor a licence', () => {
    expect(parseFamilyMetadata('# a comment\n')).toBeUndefined()
  })
})

describe('fetching a family from the web tier', () => {
  it('resolves a static family end to end and hands the command exactly what it requires', async () => {
    const { fetcher, asked } = stub(kanitUpstream())
    const outcome = await fetchWebFamily('Kanit', fetcher)
    expect(outcome.ok).toBe(true)
    if (!outcome.ok) return
    expect(outcome.face.family).toBe('Kanit')
    expect(outcome.face.style).toBe('Regular')
    // THE SPDX ID, NEVER THE UPSTREAM TOKEN. This value is literally Go's input
    // at component_commands.go's RefuseContradictedLicence call.
    expect(outcome.face.licence).toBe('OFL-1.1')
    // THE UPSTREAM FILE, NEVER A HAND-COPY — a hand-copy would be a second
    // authority on the terms.
    expect(outcome.face.licenceText).toContain('SIL Open Font License')
    // nameID 0 FROM THE FACE'S OWN BYTES, never from METADATA.pb.
    expect(outcome.face.copyright).toBe('Copyright 2020 The Kanit Project Authors')
    expect(outcome.face.mediaType).toBe('font/ttf')
    expect(outcome.face.source).toContain('ofl/kanit/Kanit-Regular.ttf')
    expect(outcome.face.layoutDivergence).toBeUndefined()
    // PROBING IS ONCE PER PICK: one metadata read, one licence file, one face.
    expect(asked).toHaveLength(3)
    expect(asked[0]).toContain('/ofl/kanit/METADATA.pb')
  })

  it('walks the probe order and stops at the directory that answers', async () => {
    const apache = kanitMetadata.replace('license: "OFL"', 'license: "APACHE2"').replace(/Kanit/g, 'Roboto Slab').replace(/robotoslab/g, 'robotoslab')
    const { fetcher, asked } = stub({
      [`${base}/apache/robotoslab/METADATA.pb`]: { body: apache },
      [`${base}/apache/robotoslab/LICENSE.txt`]: { body: 'Apache License, Version 2.0' },
      [`${base}/apache/robotoslab/Roboto Slab-Regular.ttf`]: { body: face },
    })
    const outcome = await fetchWebFamily('Roboto Slab', fetcher)
    expect(outcome.ok).toBe(true)
    expect(asked[0]).toContain('/ofl/robotoslab/METADATA.pb')
    expect(asked[1]).toContain('/apache/robotoslab/METADATA.pb')
    // AND IT DID NOT KEEP PROBING once a directory answered.
    expect(asked.filter((url) => url.endsWith('METADATA.pb'))).toHaveLength(2)
  })

  // D-16.R.6: THE DIRECTORY IS DERIVED THEN CONFIRMED. A mismatch is a refusal,
  // never a fallback to the next directory — continuing past a disagreement
  // would turn "derived then confirmed" back into the guess it replaces.
  it('refuses when METADATA.pb names a different family, and does not try the next directory', async () => {
    const { fetcher, asked } = stub({ [`${base}/ofl/geist/METADATA.pb`]: { body: 'name: "Geist Mono"\nlicense: "OFL"\n' } })
    const outcome = await fetchWebFamily('Geist', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toContain('Geist Mono')
    expect(outcome.reason).toMatch(/refused rather than guessed past/)
    expect(asked).toHaveLength(1)
  })

  // METADATA.pb ALWAYS WINS ON LICENCE; THE PROBE RESULT IS NEVER EVIDENCE OF
  // TERMS. Upstream moves families between directories — `apache/roboto` now
  // 404s and Roboto lives in `ofl/` — so reading layout as a licence assertion
  // would let a family that moved silently change the terms a document
  // publishes.
  it('admits a family whose directory disagrees with its token, and RECORDS the divergence', async () => {
    const { fetcher } = stub({
      [`${base}/ofl/movedfamily/METADATA.pb`]: { body: 'name: "Moved Family"\nlicense: "APACHE2"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "MovedFamily-Regular.ttf"\n}\n' },
      [`${base}/ofl/movedfamily/LICENSE.txt`]: { body: 'Apache License, Version 2.0' },
      [`${base}/ofl/movedfamily/MovedFamily-Regular.ttf`]: { body: face },
    })
    const outcome = await fetchWebFamily('Moved Family', fetcher)
    expect(outcome.ok, 'a layout disagreement is an observation, not a refusal').toBe(true)
    if (!outcome.ok) return
    expect(outcome.face.licence).toBe('Apache-2.0')
    expect(outcome.face.layoutDivergence).toContain('ofl/')
    expect(outcome.face.layoutDivergence).toContain('APACHE2')
  })

  it('refuses a mapped-but-unacceptable licence by name, with its reason, before any byte is fetched', async () => {
    const { fetcher, asked } = stub({
      [`${base}/ofl/sharealike/METADATA.pb`]: { body: 'name: "ShareAlike"\nlicense: "CC-BY-SA"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "ShareAlike-Regular.ttf"\n}\n' },
      [`${base}/ofl/sharealike/ShareAlike-Regular.ttf`]: { body: face },
    })
    const outcome = await fetchWebFamily('ShareAlike', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.classification?.state).toBe('refused')
    expect(outcome.reason).toContain('CC-BY-SA')
    // CLASSIFY, THEN EMBED. The face was never fetched.
    expect(asked.some((url) => url.endsWith('.ttf'))).toBe(false)
  })

  it('refuses an unrecognised token as NOT RECOGNISED rather than as forbidden, before any byte is fetched', async () => {
    const { fetcher, asked } = stub({
      [`${base}/ofl/fifthdirectory/METADATA.pb`]: { body: 'name: "Fifth Directory"\nlicense: "SOMETHINGNEW"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "FifthDirectory-Regular.ttf"\n}\n' },
      [`${base}/ofl/fifthdirectory/FifthDirectory-Regular.ttf`]: { body: face },
    })
    const outcome = await fetchWebFamily('Fifth Directory', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.classification?.state).toBe('unrecognised')
    expect(outcome.reason).toMatch(/does not recognise/)
    expect(outcome.reason).toMatch(/not the same as being forbidden/)
    expect(asked.some((url) => url.endsWith('.ttf'))).toBe(false)
  })

  it('refuses a family that publishes no upright static Regular', async () => {
    const { fetcher } = stub({ [`${base}/ofl/anuphan/METADATA.pb`]: { body: 'name: "Anuphan"\nlicense: "OFL"\nfonts {\n  style: "normal"\n  weight: 200\n  filename: "Anuphan-ExtraLight.ttf"\n}\n' } })
    const outcome = await fetchWebFamily('Anuphan', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toMatch(/no upright Regular/)
  })

  // THE `woff2` ROUTE, SHUT AT THE FILE LEVEL AS WELL AS AT THE HOST LEVEL. The
  // engine's accepted media types are exactly font/ttf and font/otf.
  it('refuses a Regular published in a format the engine does not read', async () => {
    const { fetcher } = stub({ [`${base}/ofl/subsetted/METADATA.pb`]: { body: 'name: "Subsetted"\nlicense: "OFL"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "Subsetted-Regular.woff2"\n}\n' } })
    const outcome = await fetchWebFamily('Subsetted', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toMatch(/not a font file this engine reads/)
  })

  it('refuses a family whose licence file is missing or empty, stating why', async () => {
    for (const overrides of [{ [`${base}/ofl/kanit/OFL.txt`]: { status: 404 } }, { [`${base}/ofl/kanit/OFL.txt`]: { body: '   \n' } }]) {
      const { fetcher, asked } = stub(kanitUpstream(overrides))
      const outcome = await fetchWebFamily('Kanit', fetcher)
      expect(outcome.ok).toBe(false)
      if (outcome.ok) return
      expect(outcome.reason).toMatch(/publishes no OFL\.txt/)
      expect(outcome.reason).toMatch(/may not carry a face without the text of its licence/)
      // AND THE BYTES WERE NEVER FETCHED: a face may not reach the document
      // before its terms are in hand.
      expect(asked.some((url) => url.endsWith('.ttf'))).toBe(false)
    }
  })

  // THE LICENCE FILE IS NAMED BY THE DECLARED TERMS, NOT BY THE DIRECTORY, so a
  // family sitting in `ofl/` under Apache terms is asked for LICENSE.txt.
  it('asks for the licence file the declared terms name, not the one the directory suggests', async () => {
    const { fetcher, asked } = stub({
      [`${base}/ofl/movedfamily/METADATA.pb`]: { body: 'name: "Moved Family"\nlicense: "APACHE2"\nfonts {\n  style: "normal"\n  weight: 400\n  filename: "MovedFamily-Regular.ttf"\n}\n' },
    })
    await fetchWebFamily('Moved Family', fetcher)
    expect(asked.some((url) => url.endsWith('/ofl/movedfamily/LICENSE.txt'))).toBe(true)
    expect(asked.some((url) => url.endsWith('/ofl/movedfamily/OFL.txt'))).toBe(false)
  })

  it('refuses a face that carries no copyright in its own name table', async () => {
    const { fetcher } = stub(kanitUpstream({ [`${base}/ofl/kanit/Kanit-Regular.ttf`]: { body: sfntWithNames([], { omitNameTable: true }) } }))
    const outcome = await fetchWebFamily('Kanit', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toMatch(/declares no copyright in its own `name` table \(nameID 0\)/)
  })

  // OFFLINE DEGRADES, NEVER BREAKS. No network means no NEW family; it never
  // means a document that will not render.
  it('states that a family cannot be added right now when the network is down', async () => {
    const outcome = await fetchWebFamily('Kanit', async () => { throw new TypeError('Failed to fetch') })
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toMatch(/You cannot add a family without a network connection/)
    expect(outcome.reason).toMatch(/faces this machine already holds are still offered/)
  })

  // A STALE SNAPSHOT IS A NAMED, ACTIONABLE REFUSAL. The list ships with the
  // designer and ages between releases, so it can name a family upstream has
  // since renamed or withdrawn — and the message says exactly that rather than
  // reporting a bare 404.
  it('says the snapshot is stale when every probe 404s', async () => {
    const { fetcher, asked } = stub({})
    const outcome = await fetchWebFamily('Withdrawn Family', fetcher)
    expect(outcome.ok).toBe(false)
    if (outcome.ok) return
    expect(outcome.reason).toMatch(/no longer published upstream/)
    expect(outcome.reason).toMatch(/ages between releases/)
    expect(asked).toHaveLength(probeDirectories.length)
  })
})

describe('the declared hosts', () => {
  const source = fs.readFileSync(path.join(here, 'font-source.ts'), 'utf8')

  // THE MARKER MUST BE REAL CODE ON THE LINE NAMING THE HOST. The scan computes
  // its exemption over COMMENT-BLANKED source, so a marker written in a comment
  // declares nothing — see forbidden-font-hosts.mjs's blankComments.
  it('declares both hosts in code, on the line that names them', () => {
    expect(fontHostDeclarations.map((entry) => entry.host), 'folio:font-host-declaration').toEqual(['raw.githubusercontent.com', 'fonts.google.com'])
    for (const entry of fontHostDeclarations) {
      expect(entry.declaration).toBe('folio:font-host-declaration')
      const line = source.split('\n').find((text) => text.includes(`'${entry.host}'`) && !text.trimStart().startsWith('//'))
      expect(line, `${entry.host} must be spelled on a line that also carries the marker, in code`).toContain('folio:font-host-declaration')
    }
  })

  // THE `woff2` ROUTE ASSERTED ABSENT FROM SOURCE. The stylesheet endpoint
  // returns woff2 under a modern browser UA — which the engine refuses by design
  // — split by `unicode-range` into per-script subsets, which would embed partial
  // coverage into a document naming the whole family. The full TTF that endpoint
  // serves to a legacy UA is unreachable, because a browser cannot set
  // `User-Agent`.
  it('reaches for no stylesheet endpoint, no woff2 and no unicode-range subset', () => {
    // OVER COMMENT-BLANKED SOURCE, using the scanner's own blanker. This module
    // has to NAME the refused route to explain why it is refused, and a check
    // that could not tell an explanation from a call site would either fire on
    // the prose or be silently satisfied by deleting it.
    const code = blankComments(source, '.ts')
    expect(code, 'the blanker must not have eaten the file').toContain('fetchWebFamily')
    for (const forbidden of ['css2', 'woff2', 'unicode-range', 'googleapis', 'gstatic']) {
      expect(code, `font-source.ts must not reach for ${forbidden}`).not.toMatch(new RegExp(forbidden, 'i'))
    }
  })
})
