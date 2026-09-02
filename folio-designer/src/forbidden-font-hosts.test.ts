import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import {
  DECLARATION_MARKER,
  DECLARED_ONLY_FONT_HOSTS,
  FORBIDDEN_FONT_HOSTS,
  SCANNED_FONT_HOSTS,
  POPULATION_FLOOR,
  SCANNED_EXTENSIONS,
  SCANNED_ROOTS,
  assertNoForbiddenFontHosts,
  blankComments,
  exemptLineNumbers,
  occurrencesIn,
  scanForbiddenFontHosts,
  scannedPopulation,
} from '../scripts/forbidden-font-hosts.mjs'

// STORY 8.5, AC4. THE SCAN THAT WOULD PASS VACUOUSLY, AND THE THREE THINGS
// THAT STOP IT (Design Note 2).
//
// The two forbidden hosts appear ZERO times in the trees this scan reads, so a
// scan introduced now is green BEFORE it is correct, and it would stay green if
// it scanned nothing at all, matched nothing at all, or exempted everything.
// "It passed" is therefore not evidence of anything here. What this file
// establishes instead:
//
//   1. A POSITIVE CONTROL — a fixture that really does contain a host, asserted
//      to fail, so the matcher is shown to match.
//   2. A POPULATION FLOOR — the scan reports how many files it read and refuses
//      to report a clean population over an implausibly small one, so a scan
//      that stopped looking reds instead of passing.
//   3. THE COMMENT DIRECTION — a host inside a comment STILL FAILS. Comments
//      are stripped from the EXEMPTION, never from the scan, and deleting that
//      stripping step reds a named test below on its own message.
//
// AND THE CLAIM IS BOUNDED (D-8.5.5). What a green here means is that no
// forbidden host appears in the scanned population. It is never a claim that no
// request leaves the machine, and no test here should ever be written as if it
// were.

const here = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.join(here, '..', '..')
const hosts = FORBIDDEN_FONT_HOSTS.map((entry) => entry.host)
const declaredOnly = DECLARED_ONLY_FONT_HOSTS.map((entry) => entry.host)

describe('no forbidden font host reaches the scanned source', () => {
  it('forbids exactly the two Google Fonts hosts, spelled here as well as there', () => {
    // Both hosts are spelled in this test file, so this line declares them the
    // way the scanner requires: IN CODE, on the same line, with the marker.
    // A comment could not do it — see the comment-direction test below.
    expect([...hosts].sort(), 'folio:font-host-declaration').toEqual(['fonts.googleapis.com', 'fonts.gstatic.com'])
    expect(DECLARATION_MARKER).toBe('folio:font-host-declaration')
  })

  // THE REAL SCAN, OVER THE REAL TREE. Its population size is asserted first,
  // because the finding count is only meaningful next to what was read.
  it('reads a real population of the product source trees and finds no forbidden host', () => {
    const result = scanForbiddenFontHosts(repoRoot)
    expect(result.files, `the scan read only ${result.files} files, which cannot support a claim about this repository`).toBeGreaterThan(POPULATION_FLOOR)
    expect(result.findings, 'a forbidden font host appears in the product source. The catalogue is bundled and precached (D-8.5.12).').toEqual([])
    // AND IT SPANS EVERY TREE IT CLAIMS TO, EXACTLY.
    //
    // THE EXPECTATION IS DERIVED FROM `SCANNED_ROOTS`, not a subset written out
    // here. It used to name three of the six by hand and pipe them through
    // `expect.arrayContaining`, which made the `.sort()` beside it a no-op and
    // read as an exact-set check while being a containment one: `hashmatrix`,
    // `tools` and `.github` could have dropped out of the walk entirely and
    // this stayed green. Now adding a root the walk cannot reach reds, and so
    // does reaching a directory no root declares.
    const reached = [...new Set(scannedPopulation(repoRoot).map((file) => file.split('/')[0]))].sort()
    expect(reached, 'every declared scan root must contribute at least one file, and nothing outside them may').toEqual([...SCANNED_ROOTS].sort())
    expect(SCANNED_ROOTS.length, 'the scan roots must not have collapsed to a single tree').toBeGreaterThanOrEqual(3)
    // And what it deliberately does NOT read, so the bound is a decision on the
    // record rather than an accident of a filter (D-8.5.5).
    expect([...reached]).not.toContain('_bmad-output')
    expect([...reached]).not.toContain('docs')
  })

  // 1 — THE POSITIVE CONTROL. Without this, every assertion above is satisfied
  // by a scanner that matches nothing.
  it('fails, naming the file and the line, on a source file that contains a forbidden host', () => {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-font-host-'))
    try {
      execFileSync('git', ['-C', scratch, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      const tree = path.join(scratch, 'folio-designer', 'src')
      fs.mkdirSync(tree, { recursive: true })
      fs.writeFileSync(path.join(tree, 'clean.ts'), 'export const ok = 1\n')
      fs.writeFileSync(path.join(tree, 'offender.ts'), `const a = 1\nconst url = 'https://${hosts[0]}/css2?family=Inter'\n`)
      execFileSync('git', ['-C', scratch, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })

      const result = scanForbiddenFontHosts(scratch, { floor: 1 })
      expect(result.findings).toEqual([{ file: 'folio-designer/src/offender.ts', host: hosts[0], half: 'forbidden', line: 2, text: `const url = 'https://${hosts[0]}/css2?family=Inter'` }])
      // The thrown message NAMES the file and the line (AC4), not merely a count.
      expect(() => assertNoForbiddenFontHosts(scratch, { floor: 1 })).toThrow(/folio-designer\/src\/offender\.ts:2/)
      // And the clean file is not reported, so the matcher discriminates.
      expect(result.findings.map((finding) => finding.file)).not.toContain('folio-designer/src/clean.ts')
    } finally {
      fs.rmSync(scratch, { recursive: true, force: true })
    }
  })

  // 2 — THE POPULATION FLOOR. A scan that read almost nothing is
  // indistinguishable from a scan that found nothing, and it must not be
  // allowed to report the second.
  it('refuses to report a clean population over an implausibly small one', () => {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-font-host-floor-'))
    try {
      execFileSync('git', ['-C', scratch, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      const tree = path.join(scratch, 'folio-go')
      fs.mkdirSync(tree, { recursive: true })
      fs.writeFileSync(path.join(tree, 'main.go'), 'package main\n')
      execFileSync('git', ['-C', scratch, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })

      // Clean by content — and still refused, because one file is not a population.
      expect(scanForbiddenFontHosts(scratch, { floor: 1 }).findings).toEqual([])
      expect(() => assertNoForbiddenFontHosts(scratch)).toThrow(/read only 1 files, under its floor of/)
    } finally {
      fs.rmSync(scratch, { recursive: true, force: true })
    }
  })

  // AND THE POPULATION READER REFUSES TO GO QUIET IN THE OTHER TWO DIRECTIONS
  // TOO: a directory git does not track at all, and a repository with nothing
  // in the scanned trees. An unobtainable population must never read as an
  // all-clear.
  it('throws rather than reporting an empty population it could not obtain', () => {
    const notARepository = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-font-host-nogit-'))
    const emptyRepository = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-font-host-empty-'))
    try {
      expect(() => scannedPopulation(notARepository)).toThrow(/could not look/)
      execFileSync('git', ['-C', emptyRepository, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      fs.mkdirSync(path.join(emptyRepository, 'docs'), { recursive: true })
      fs.writeFileSync(path.join(emptyRepository, 'docs', 'page.html'), '<!doctype html>\n')
      execFileSync('git', ['-C', emptyRepository, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })
      expect(() => scannedPopulation(emptyRepository)).toThrow(/carries a scanned extension/)
    } finally {
      fs.rmSync(notARepository, { recursive: true, force: true })
      fs.rmSync(emptyRepository, { recursive: true, force: true })
    }
  })

  // 3 — THE COMMENT DIRECTION, the one that is easy to get exactly backwards.
  //
  // A HOST INSIDE A COMMENT STILL FAILS. The obvious implementation of "strip
  // comments" would exempt it, which is precisely the shape someone reaching
  // for the declined middle tier would leave behind: the host parked in a
  // commented-out line, ready to be uncommented.
  it('fails on a forbidden host that appears inside a comment, in every comment form', () => {
    for (const form of [
      `// const url = 'https://${hosts[0]}/css2'`,
      `/* const url = 'https://${hosts[0]}/css2' */`,
      `const a = 1 // https://${hosts[1]}`,
    ]) {
      expect(occurrencesIn(form).length, `a forbidden host commented out as ${JSON.stringify(form)} must still be reported`).toBeGreaterThan(0)
    }
  })

  // AND THE EXEMPTION IS WHAT COMMENTS ARE STRIPPED FROM. This is the mutation
  // proof for the stripping step: make `exemptLineNumbers` read the RAW text
  // instead of the comment-blanked text and this test reds on its own message,
  // because a marker written in a comment would then exempt the host beside it.
  it('honours a declaration written in code and refuses the same declaration written in a comment', () => {
    const inCode = `const entry = { host: '${hosts[0]}', declaration: '${DECLARATION_MARKER}' }`
    const inComment = `// '${hosts[0]}' ${DECLARATION_MARKER}`
    const inCommentTrailing = `const url = 'https://${hosts[0]}/css2' // ${DECLARATION_MARKER}`

    expect(exemptLineNumbers(inCode, hosts[0]), 'a declaration written in code exempts its own line').toEqual(new Set([1]))
    expect(occurrencesIn(inCode)).toEqual([])

    expect(exemptLineNumbers(inComment, hosts[0]), 'a declaration written in a comment declares nothing — comments are stripped FROM THE EXEMPTION').toEqual(new Set())
    expect(occurrencesIn(inComment).length).toBe(1)
    expect(exemptLineNumbers(inCommentTrailing, hosts[0]), 'a trailing comment cannot exempt a live line').toEqual(new Set())
    expect(occurrencesIn(inCommentTrailing).length).toBe(1)
  })

  // THE COMMENT BLANKER ITSELF, shown to blank comments AND to leave strings
  // alone. A blanker that ate the `//` inside a URL string would strip the rest
  // of that line and make the exemption fire on text nobody wrote.
  it('blanks comments, preserves line numbering, and does not mistake a URL for a comment', () => {
    // LINE NUMBERING IS THE PROPERTY, not the exact run of spaces: the
    // exemption is computed per line and reported against the RAW text's line
    // numbers, so a blanker that dropped a newline would silently exempt the
    // wrong line. Asserted structurally so the check says what it means.
    const blanked = (source: string, extension?: string) => blankComments(source, extension).split('\n')

    expect(blanked('a\n// b\nc'), 'a line comment leaves an empty line where it was').toEqual(['a', '    ', 'c'])
    expect(blanked('a\n/* b\nc */\nd'), 'a block comment blanks every line it spans').toEqual(['a', '    ', '    ', 'd'])
    expect(blanked(`const u = 'https://example.test/x' // gone`), 'the // inside a URL string is not a comment').toEqual([`const u = 'https://example.test/x'        `])
    expect(blanked('a: 1 # note\nb: 2', '.yml')).toEqual(['a: 1       ', 'b: 2'])
    expect(blanked('a: 1 # note\nb: 2', '.ts'), 'a hash is not a comment in TypeScript').toEqual(['a: 1 # note', 'b: 2'])
    expect(blanked('<p>x</p><!-- y -->\n<p>z</p>', '.html')).toEqual(['<p>x</p>          ', '<p>z</p>'])
  })


  // ─────────────────────────────────────────────────────────────────────────
  // STORY 16.1 — THE SECOND HALF (D-16.4). The guard was AMENDED, not deleted.
  //
  // D-16.1 reversed "no live font service", so this scan's original subject is
  // gone and the naive reading is that the guard has nothing left to say. The
  // amendment is the answer: two hosts are now ALLOWED, and allowed in exactly
  // one module, so the question the scan asks changed from "is this reached at
  // all" to "is this reached from anywhere but its one declared home".
  //
  // AND IT LANDED IN THE SAME STORY AS ITS SUBJECT, deliberately. A guard that
  // arrives after the population it polices is how D-8.6.5 shipped green.
  // ─────────────────────────────────────────────────────────────────────────

  it('declares exactly the two allowed hosts as declared-only, alongside the two forbidden outright', () => {
    // Both halves are spelled in this file, so this line declares them the way
    // the scanner requires: IN CODE, on the same line, with the marker.
    expect([...declaredOnly].sort(), 'folio:font-host-declaration').toEqual(['fonts.google.com', 'raw.githubusercontent.com'])
    expect(SCANNED_FONT_HOSTS.map((entry) => entry.half)).toEqual(['forbidden', 'forbidden', 'declared-only', 'declared-only'])
    // THE TWO HALVES ARE DISJOINT. A host in both would make the message a lie
    // whichever sentence it printed.
    expect(declaredOnly.filter((host) => hosts.includes(host))).toEqual([])
  })

  // THE POSITIVE CONTROL FOR THE NEW HALF. Without it, every assertion about
  // declared-only hosts is satisfied by a scanner that matches nothing — which
  // is exactly the vacuity the first half was written around, one host list
  // later.
  it('fails, naming the file, the line and the half, on an allowed host used outside its declaring module', () => {
    const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'folio-font-host-declared-'))
    try {
      execFileSync('git', ['-C', scratch, 'init', '-q'], { stdio: ['ignore', 'pipe', 'pipe'] })
      const tree = path.join(scratch, 'folio-designer', 'src')
      fs.mkdirSync(tree, { recursive: true })
      // The SAME host, twice: once used, once declared. Only the used one is a
      // finding, which is what makes this a discrimination rather than a match.
      fs.writeFileSync(path.join(tree, 'second-fetch-site.ts'), `const url = 'https://${declaredOnly[0]}/google/fonts/main/ofl/x/x.ttf'\n`)
      fs.writeFileSync(path.join(tree, 'font-source.ts'), `export const host = { host: '${declaredOnly[0]}', declaration: '${DECLARATION_MARKER}' }.host\n`)
      execFileSync('git', ['-C', scratch, 'add', '-A'], { stdio: ['ignore', 'pipe', 'pipe'] })

      const result = scanForbiddenFontHosts(scratch, { floor: 1 })
      expect(result.findings.map((finding) => finding.file)).toEqual(['folio-designer/src/second-fetch-site.ts'])
      expect(result.findings[0].half).toBe('declared-only')
      expect(() => assertNoForbiddenFontHosts(scratch, { floor: 1 })).toThrow(/folio-designer\/src\/second-fetch-site\.ts:1/)
      // AND THE TWO HALVES SAY DIFFERENT THINGS. A reader told "this host is
      // forbidden" about a host the product legitimately fetches from would
      // either delete a working call site or delete the guard.
      expect(() => assertNoForbiddenFontHosts(scratch, { floor: 1 })).toThrow(/DECLARED-ONLY/)
      expect(() => assertNoForbiddenFontHosts(scratch, { floor: 1 })).not.toThrow(/FORBIDDEN OUTRIGHT/)
    } finally {
      fs.rmSync(scratch, { recursive: true, force: true })
    }
  })

  // THE MUTATION PROOF FOR THE NEW HALF, TAKEN BY DELETING THE HALF rather than
  // by falsifying a condition: with `DECLARED_ONLY_FONT_HOSTS` empty, the same
  // offending source produces no finding at all. That is what makes the green
  // above a claim about the scanner and not about the fixture.
  it('reds only because the second half exists: removing it makes the same offender invisible', () => {
    const offender = `const url = 'https://${declaredOnly[0]}/google/fonts/main/ofl/x/x.ttf'`
    expect(occurrencesIn(offender).map((finding) => finding.host)).toEqual([declaredOnly[0]])
    // The deletion, simulated over the same matcher the scan uses: a host in
    // neither list is not a finding, whatever it is.
    expect(occurrencesIn(offender.replace(declaredOnly[0], 'example.test')).length).toBe(0)
  })

  // AND THE MARKER DIRECTION IS PRESERVED ACROSS THE NEW HALF. The scan runs
  // over RAW source while the exemption runs over COMMENT-BLANKED source, so a
  // declaration written in a comment declares nothing — for an allowed host
  // exactly as for a forbidden one. A comment is precisely where someone parks
  // a second fetch site.
  it('honours a declared-only host declared in code and refuses the same declaration written in a comment', () => {
    for (const host of declaredOnly) {
      const inCode = `const entry = { host: '${host}', declaration: '${DECLARATION_MARKER}' }`
      expect(exemptLineNumbers(inCode, host)).toEqual(new Set([1]))
      expect(occurrencesIn(inCode)).toEqual([])

      const inComment = `// '${host}' ${DECLARATION_MARKER}`
      expect(exemptLineNumbers(inComment, host)).toEqual(new Set())
      expect(occurrencesIn(inComment).length).toBe(1)

      const commentedOut = `// const url = 'https://${host}/x'`
      expect(occurrencesIn(commentedOut).length, 'a commented-out use still fails: comments are stripped from the EXEMPTION, not from the scan').toBe(1)
    }
  })

  // THE REAL SUBJECT, IN THE REAL TREE. `src/font-source.ts` is the single
  // declared home of every allowed host, and this asserts that it EXISTS and
  // really carries the declarations — a second half whose only subject was a
  // fixture would be green over a repository that had none.
  it('finds its declared subject in the real source, and reds when that subject stops declaring', () => {
    const fontSource = fs.readFileSync(path.join(here, 'font-source.ts'), 'utf8')
    for (const host of declaredOnly) {
      expect(exemptLineNumbers(fontSource, host, '.ts').size, `${host} must be declared in code in src/font-source.ts`).toBeGreaterThan(0)
      expect(occurrencesIn(fontSource, '.ts').filter((finding) => finding.host === host)).toEqual([])
    }
    // AND STRIPPING THE DECLARATION REDS IT. The marker is removed from the
    // module's own source and the same scanner is re-run over it: every host it
    // spells becomes a finding. This is the red-proof for the half, taken by
    // removing the declaration rather than by weakening the check.
    const undeclared = fontSource.split(DECLARATION_MARKER).join('undeclared')
    const findings = occurrencesIn(undeclared, '.ts')
    expect(findings.length, 'with its declarations stripped, font-source.ts must fail the scan').toBeGreaterThanOrEqual(declaredOnly.length)
    expect([...new Set(findings.map((finding) => finding.host))].sort()).toEqual([...declaredOnly].sort())
    expect(findings.every((finding) => finding.half === 'declared-only')).toBe(true)
  })

  it('scans the source extensions a font host would realistically be typed into', () => {
    for (const extension of ['.ts', '.tsx', '.js', '.mjs', '.go', '.css', '.html', '.json']) expect(SCANNED_EXTENSIONS).toContain(extension)
  })
})
