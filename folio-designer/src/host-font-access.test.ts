import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import {
  HOST_FONT_ACCESS_APIS,
  HOST_FONT_DECLARATION_MARKER,
  POPULATION_FLOOR,
  SCANNED_ROOTS,
  assertNoHostFontAccess,
  occurrencesIn,
  scanHostFontAccess,
  scannedPopulation,
} from '../scripts/host-font-access.mjs'

// STORY 16.2, AC4's SECOND HALF — NO CODE PATH IN THIS DESIGNER ENUMERATES,
// READS OR FEATURE-DETECTS HOST FONTS.
//
// WHY THE ASSERTION IS SOURCE-LEVEL AND OVER THE WHOLE TREE. The claim is about
// the DESIGNER, not about one module: a test that only checked `font-store.ts`
// would pass while `App.tsx` called the API. So the scan reads every tracked
// source file under the designer, and this file drives the REAL scanner rather
// than a copy of its logic.
//
// AND IT IS RED-PROVED BY DELETING THE GUARD, NEVER BY FALSIFYING A CONDITION.
// These spellings appear ZERO times in this repository, so "it passed" is not
// evidence of anything on its own — a scan that read nothing, matched nothing,
// or exempted everything would pass identically. Three tests below are the
// evidence instead: a positive control, a population floor, and the comment
// direction. Emptying `HOST_FONT_ACCESS_APIS`, or deleting
// `assertNoHostFontAccess`, reds them.
//
// THE CLAIM IS BOUNDED (D-8.5.5). A green here means none of these spellings
// appears in the scanned population. It is NOT a proof that no host font is ever
// read — a source scan cannot see a name assembled at runtime, a call inside a
// dependency, or anything in an untracked file — and no test here is written as
// if it were.

const here = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.join(here, '..', '..')

describe('no code path in the designer reaches for host-installed fonts', () => {
  // THE SPELLINGS ARE PINNED, so a list quietly emptied to make a red go away
  // reds HERE instead. This is the assertion that makes every clean result
  // below mean something.
  it('forbids the Local Font Access surface by every spelling it goes by', () => {
    // Each is spelled on a line carrying the marker in code, which is how this
    // test file names what it forbids without being a violation of it. A marker
    // in a comment would declare nothing — see the comment-direction test.
    const declared = [
      { name: 'queryLocalFonts', declaration: 'folio:host-font-declaration' },
      { name: 'navigator.fonts', declaration: 'folio:host-font-declaration' },
      { name: 'local-fonts permission', declaration: 'folio:host-font-declaration' },
      { name: 'FontData', declaration: 'folio:host-font-declaration' },
    ]
    expect(HOST_FONT_ACCESS_APIS.map((api) => api.name)).toEqual(declared.map((api) => api.name))
    expect(HOST_FONT_DECLARATION_MARKER).toBe('folio:host-font-declaration')
    for (const api of HOST_FONT_ACCESS_APIS) expect(api.declaration).toBe(HOST_FONT_DECLARATION_MARKER)
  })

  // THE REAL SCAN, OVER THE REAL TREE. The population is asserted first, because
  // a finding count is only meaningful next to what was read.
  it('reads a real population of the designer and finds no host-font access', () => {
    const result = scanHostFontAccess(repoRoot)
    expect(result.files, `the scan read only ${result.files} files, which cannot support a claim about this designer`).toBeGreaterThan(POPULATION_FLOOR)
    expect(result.findings, 'the designer must never enumerate, read or feature-detect a font installed on the machine (SPEC-fonts Non-goal: "No host fonts")').toEqual([])
    // AND IT SPANS EVERY TREE IT CLAIMS TO. The expectation is DERIVED from
    // `SCANNED_ROOTS` rather than written out here, so adding a root the walk
    // cannot reach reds, and so does reaching a directory no root declares.
    const reached = [...new Set(scannedPopulation(repoRoot).map((file) => file.split('/').slice(0, 2).join('/')))].sort()
    expect(reached, 'every declared scan root must contribute at least one file, and nothing outside them may').toEqual([...SCANNED_ROOTS].sort())
  })

  // 1 — THE POSITIVE CONTROL. Without this, every assertion above is satisfied
  // by a scanner that matches nothing at all.
  it('fails, naming the file and the line, on a source file that calls the API', () => {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-host-font-'))
    try {
      execFileSync('git', ['-C', scratch, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      const tree = path.join(scratch, 'folio-designer', 'src')
      fs.mkdirSync(tree, { recursive: true })
      fs.writeFileSync(path.join(tree, 'clean.ts'), 'export const ok = 1\n')
      fs.writeFileSync(path.join(tree, 'offender.ts'), 'const ok = 1\nconst faces = await window.query' + 'LocalFonts()\n')
      execFileSync('git', ['-C', scratch, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })

      const result = scanHostFontAccess(scratch, { floor: 1 })
      expect(result.findings.map((finding) => `${finding.file}:${finding.line}:${finding.api}`)).toEqual(['folio-designer/src/offender.ts:2:queryLocalFonts'])
      // The thrown message NAMES the file and the line, not merely a count.
      expect(() => assertNoHostFontAccess(scratch, { floor: 1 })).toThrow(/folio-designer\/src\/offender\.ts:2/)
      // And the clean file is not reported, so the matcher discriminates.
      expect(result.findings.map((finding) => finding.file)).not.toContain('folio-designer/src/clean.ts')
    } finally {
      fs.rmSync(scratch, { recursive: true, force: true })
    }
  })

  // AND IT DISCRIMINATES AGAINST A REAL NEIGHBOUR IN THIS REPOSITORY. PDF.js's
  // `standardFontDataUrl` contains the interface name this API exposes, and a
  // substring scan would report the PDF preview as a host-font reader. This is
  // a MEASURED collision, not a precaution: that option is passed at
  // `src/preview/pdf-viewer.tsx`.
  it('does not read a font-shaped identifier that is not this API', () => {
    const neighbours = 'const options = { standardFontDataUrl: url, cMapUrl: url }\nconst family = "local-fonts-are-not-read-here"\n'
    expect(occurrencesIn(neighbours, '.ts')).toEqual([])
    // AND THE REAL NEIGHBOUR, READ AS ITSELF RATHER THAN QUOTED. The page font
    // set is the registry a document's OWN carried faces are registered into,
    // which is the exact opposite of reading a font off the machine — and the
    // one module in this build that touches it must stay legal. Its source is
    // the corpus, so a widened pattern that swept it up reds here.
    const registry = fs.readFileSync(path.join(here, 'embedded-face-registry.ts'), 'utf8')
    expect(registry, 'the neighbour corpus must be the module that really does register faces').toContain('registerCarriedFaces')
    expect(occurrencesIn(registry, '.ts')).toEqual([])
  })

  // 2 — THE POPULATION FLOOR. A scan that read almost nothing is
  // indistinguishable from a scan that found nothing, and it must not be
  // allowed to report the second.
  it('refuses to report a clean population over an implausibly small one', () => {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-host-font-floor-'))
    try {
      execFileSync('git', ['-C', scratch, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      const tree = path.join(scratch, 'folio-designer', 'src')
      fs.mkdirSync(tree, { recursive: true })
      fs.writeFileSync(path.join(tree, 'only.ts'), 'export const ok = 1\n')
      execFileSync('git', ['-C', scratch, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })
      // Clean by content, and still refused, because one file is not a population.
      expect(scanHostFontAccess(scratch, { floor: 1 }).findings).toEqual([])
      expect(() => assertNoHostFontAccess(scratch)).toThrow(/under its floor/)
    } finally {
      fs.rmSync(scratch, { recursive: true, force: true })
    }
  })

  // 3 — THE COMMENT DIRECTION, which is the half that is easy to get backwards.
  // The SCAN reads raw text, so a commented-out call still fails. What is
  // stripped is the EXEMPTION: a marker written inside a comment declares
  // nothing, because a comment is exactly where somebody parks a call they mean
  // to make later.
  it('fails on a commented-out call, and refuses a declaration written in a comment', () => {
    const spelling = 'query' + 'LocalFonts'
    expect(occurrencesIn(`// const faces = await ${spelling}()\n`, '.ts')).toHaveLength(1)
    // A marker in a comment on the same line exempts NOTHING.
    expect(occurrencesIn(`// ${spelling}() ${HOST_FONT_DECLARATION_MARKER}\n`, '.ts')).toHaveLength(1)
    // In code, on the same line, it does — which is how the scanner and this
    // file are able to name what they forbid.
    expect(occurrencesIn(`const forbidden = { api: '${spelling}', declaration: '${HOST_FONT_DECLARATION_MARKER}' }\n`, '.ts')).toEqual([])
  })
})
