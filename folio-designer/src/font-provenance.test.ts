import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { catalogueFaces } from './generated/font-catalogue'
import { fontHostDeclarations, webFaceSource } from './font-source'

// STORY 16.1a — THE TRIPWIRE UNDER `source`, ON BOTH TIERS (D-16.R.13, DW-160).
//
// `source` is one of the twelve wire fields a `.folio` records for an embedded
// face, and until this story its two writers disagreed in KIND:
//
//   committed tier  `folio-designer/public/fonts/<dir>/<file> — see that
//                    directory's NOTICE.md …`   (a path into THIS repository,
//                    naming a file the recipient of a `.folio` does not have)
//   fetched tier    `<the declared fetch host>/google/fonts/main/<path>`
//                    (a bare MUTABLE BRANCH URL)
//
// D-16.R.13's verdict: `source` carries the **upstream project**, the **path
// within it**, and the **fetch date** — and three things it must never carry.
//
//   - **No resolvable-looking URL.** A string shaped like an address is a
//     promise of fetchability, and a promise that decays is worse than none: a
//     dead link in a year reads as "this provenance is broken" when the
//     provenance is intact.
//   - **No branch name.** `main` does not identify bytes.
//   - **No SHA-256.** The face is stored under its digest as its asset key, and
//     the key travels with the record when an asset is lifted into another
//     document. Restating it here would create two authorities on one fact that
//     can disagree.
//
// AND THIS FILE EXISTS BECAUSE THE CONVENTION ALONE WILL NOT HOLD. The ruling
// asked for the tripwire by name. Both tiers are asserted here, together, on
// purpose: the defect being guarded against is not "a bad string" but "the two
// writers drifting apart", which no test looking at one tier can see.

const here = path.dirname(fileURLToPath(import.meta.url))

/**
 * A HOST, RECOGNISED BY ITS TLD RATHER THAN BY A LIST OF KNOWN FONT HOSTS.
 *
 * A list of the hosts this product does reach would pass the moment somebody
 * reached for a new one, which is the whole failure shape. The TLD set is
 * deliberately narrow so that a FILENAME does not read as a host: `.ttf`,
 * `.pb`, `.md` and a version like `4.005_Desktop` are not in it.
 */
const hostShaped = /\b[a-z0-9][a-z0-9-]*\.(?:com|org|net|io|dev|co|app|sh|xyz)\b/i
const schemeShaped = /[a-z][a-z0-9+.-]*:\/\//i
const digestShaped = /\b[0-9a-f]{64}\b/i
const branchShaped = /(?:^|[/@])(?:main|master|HEAD)(?:[/\s]|$)/

function assertProvenanceShape(tier: string, subject: string, value: string): void {
  expect(value, `${tier}: ${subject} publishes an empty source`).not.toBe('')
  expect(value, `${tier}: ${subject} carries a URL scheme in \`source\`. A resolvable-looking string is a promise of fetchability, and a promise that decays reads as broken provenance (D-16.R.13).`).not.toMatch(schemeShaped)
  expect(value, `${tier}: ${subject} carries a HOST in \`source\`. \`source\` names provenance, not a retrieval path (D-16.R.13).`).not.toMatch(hostShaped)
  expect(value, `${tier}: ${subject} carries a branch name in \`source\`. A branch does not identify the bytes the field claims to describe — that was the defect (D-16.R.13).`).not.toMatch(branchShaped)
  expect(value, `${tier}: ${subject} restates a SHA-256 in \`source\`. The digest is already the asset key, and duplicating it puts two authorities on one fact (D-16.R.13).`).not.toMatch(digestShaped)
  // AND IT IS STILL PROVENANCE, not merely a string that avoids three
  // prohibitions: a project, a path within it, and a fetch date.
  expect(value, `${tier}: ${subject} does not name a fetch date`).toMatch(/, fetched \d{4}-\d{2}-\d{2}$/)
  expect(value.split(' — ')[0], `${tier}: ${subject} does not name an upstream project`).not.toBe('')
}

describe('`source` names provenance and never a retrieval path', () => {
  // NON-VACUITY FIRST. Every assertion below is inside a loop over the
  // generated catalogue, and a module that emitted nothing would satisfy them
  // all in silence — the exact shape this suite's other guards are written
  // against.
  it('is asserted over the whole committed tier, not over a sample of it', () => {
    expect(catalogueFaces.length, 'the generated catalogue is empty, so every committed-tier assertion below is vacuous').toBeGreaterThanOrEqual(31)
  })

  it('carries no scheme and no host on the committed tier', () => {
    for (const face of catalogueFaces) assertProvenanceShape('committed tier', face.family, face.source)
  })

  // AND IT NO LONGER POINTS AT A FILE THAT DOES NOT TRAVEL. The old string named
  // `NOTICE.md` in this repository; a `.folio` reaches its recipient without it.
  it('inlines the pinned upstream release rather than pointing at this repository', () => {
    for (const face of catalogueFaces) {
      expect(face.source, `${face.family}: \`source\` points at a NOTICE.md the recipient of a .folio does not have`).not.toContain('NOTICE.md')
      expect(face.source, `${face.family}: \`source\` points into this repository's own tree`).not.toContain('folio-designer/')
      expect(face.source, `${face.family}: \`source\` names no pinned upstream release`).toMatch(/^[^\s@]+@[^\s@]+ — \S+, fetched \d{4}-\d{2}-\d{2}$/)
    }
  })

  it('carries no scheme and no host on the fetched tier', () => {
    assertProvenanceShape('fetched tier', 'a fetched face', webFaceSource('ofl/kanit/Kanit-Regular.ttf', '2026-09-03'))
    // The path within the project survives, because it is the half of the
    // record that says WHICH face — dropping it would make the field name a
    // project and nothing else.
    expect(webFaceSource('ofl/kanit/Kanit-Regular.ttf', '2026-09-03')).toContain('ofl/kanit/Kanit-Regular.ttf')
  })

  // THE HOST CONSTANTS ARE THE THING THE FETCHED TIER USED TO INTERPOLATE, so
  // the tripwire is stated against them by identity rather than by spelling: a
  // rename in `fontHostDeclarations` cannot walk out from under this.
  it('never interpolates a declared font host into the field', () => {
    expect(fontHostDeclarations.length, 'no host is declared, so the check below compares against nothing').toBeGreaterThan(0)
    const written = [...catalogueFaces.map((face) => face.source), webFaceSource('ofl/kanit/Kanit-Regular.ttf', '2026-09-03')]
    for (const { host } of fontHostDeclarations) {
      for (const value of written) expect(value, `\`source\` names the fetch host ${host}`).not.toContain(host)
    }
  })

  // AND THE ONE WRITER ON EACH TIER IS THE ONE THIS FILE CHECKS. Both
  // assertions above read a FUNCTION; a second assignment site writing its own
  // string would be invisible to them, which is how the branch URL survived a
  // suite that already had opinions about this field.
  it('has exactly one writer per tier, and this file checks it', () => {
    const fetched = fs.readFileSync(path.join(here, 'font-source.ts'), 'utf8')
    const assignments = [...fetched.matchAll(/^\s*source: (.+),$/gm)].map((match) => match[1])
    expect(assignments, 'font-source.ts must build `source` through webFaceSource and nowhere else').toEqual(['webFaceSource(`${directory}/${slug}/${filename}`, today)'])

    const build = fs.readFileSync(path.join(here, '..', 'scripts', 'build-wasm.mjs'), 'utf8')
    expect(build, 'build-wasm.mjs must build the committed tier\'s `source` through committedFaceSource').toContain('source: ${JSON.stringify(committedFaceSource(face))}')
  })
})
