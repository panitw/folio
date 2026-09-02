import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { catalogueFaces } from './generated/font-catalogue'
import { fontHostDeclarations, webFaceSource } from './font-source'
import { assertProvenanceShape } from './test/provenance-shape'

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

// THE PREDICATE ITSELF LIVES IN `src/test/provenance-shape.ts`, shared with
// `font-source.test.ts`, which asserts the same shape over the FETCHED tier's
// real write path. Shared rather than copied on purpose: the defect this
// tripwire guards is the two tiers drifting apart, and two copies of the rule
// can drift exactly as the two writers did.

describe('`source` names provenance and never a retrieval path', () => {
  // NON-VACUITY FIRST. Every assertion below is inside a loop over the
  // generated catalogue, and a module that emitted nothing would satisfy them
  // all in silence — the exact shape this suite's other guards are written
  // against.
  //
  // AND IT IS THE POPULATION FLOOR — ONE OF FOUR, AND ALL FOUR MOVE TOGETHER.
  // The other three are `src/font-catalogue.test.ts` ("declares at least twenty
  // NEW families"), `src/font-index.test.ts` ("is the whole bundled catalogue,
  // unchanged") and `src/font-name-table.test.ts` ("reads a copyright out of
  // every committed catalogue face"). Raised 20 -> 31 by Story 16.1a.
  // D-16.R.18: "a floor that exists in three files is three floors, and a
  // ruling that says *the floor* has already lost track of one of them" — this
  // site is the fourth, added by the same story, and is named at the other
  // three so the count cannot quietly go stale again.
  it('is asserted over the whole committed tier, not over a sample of it', () => {
    expect(catalogueFaces.length, 'the generated catalogue is empty, so every committed-tier assertion below is vacuous; this is one of FOUR population floors and all four move together').toBeGreaterThanOrEqual(31)
  })

  it('carries no scheme and no host on the committed tier', () => {
    for (const face of catalogueFaces) assertProvenanceShape(expect, 'committed tier', face.family, face.source)
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
    assertProvenanceShape(expect, 'fetched tier', 'webFaceSource, called directly', webFaceSource('ofl/kanit/Kanit-Regular.ttf', '2026-09-03'))
    // The path within the project survives, because it is the half of the
    // record that says WHICH face — dropping it would make the field name a
    // project and nothing else.
    expect(webFaceSource('ofl/kanit/Kanit-Regular.ttf', '2026-09-03')).toContain('ofl/kanit/Kanit-Regular.ttf')
    // AND THE REAL WRITE PATH IS ASSERTED ELSEWHERE, deliberately. This case
    // calls `webFaceSource` with an EXPLICIT date, so it cannot observe
    // `fetchWebFamily`'s default `today` expression — measured: deleting the
    // `.slice(0, 10)` from that default publishes a full ISO timestamp on every
    // fetched pick and leaves this file entirely green. `font-source.test.ts`
    // drives `fetchWebFamily` with NO `today` argument against the kanit stub
    // and applies this same predicate to the outcome, which is the case that
    // reds under that mutation.
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
