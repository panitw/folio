import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

// STORY 8.5, AC4 — THE FORBIDDEN-HOST SOURCE SCAN.
//
// WHAT IT PROVES, AND THE SENTENCE IT MUST NEVER BE USED TO SAY (D-8.5.5).
// It proves that no forbidden font host appears in the SCANNED POPULATION, and
// nothing more. It is NOT a proof that "no request leaves the machine" — a
// source scan cannot see a host assembled at runtime, a host in an untracked
// file, a host in a dependency, or a request made by something this repository
// did not write. Every message below is worded to claim the bounded thing.
//
// WHY IT EXISTS. `SPEC.md`'s `## Non-goals` forbids a live font service, an
// arbitrary URL and a download on first use, and D-8.5.12 declined the owner
// brief's "size not paid for at first load" shape ON THE CONTRACT rather than
// by inventing a middle tier. The catalogue is therefore BUNDLED AND
// PRECACHED. This scan is the assertion that the decline was honoured in code
// rather than only in prose: the moment someone reaches for the cheap shape,
// the first thing they type is one of these two hosts.
//
// IT WOULD PASS VACUOUSLY ON INTRODUCTION, AND THAT IS THE TRAP (Design Note
// 2). These hosts appeared ZERO times in this repository's source before this
// story, so a scan added now is green before it is correct and would stay green
// if it scanned nothing at all. Three things are therefore load-bearing rather
// than optional, and each has its own test in `src/forbidden-font-hosts.test.ts`:
// a POSITIVE CONTROL (a fixture that does contain a host, asserted to fail), a
// POPULATION FLOOR (the scan reports how many files it read and refuses to
// report a clean population over an implausibly small one), and the
// COMMENT DIRECTION below.

/**
 * The hosts a bundled-and-precached catalogue must never reach for.
 *
 * EACH ENTRY CARRIES THE DECLARATION MARKER ON ITS OWN LINE, IN CODE — that is
 * how this file, which must spell the hosts to forbid them, is not itself a
 * violation. See `exemptLineNumbers` for why the marker has to be code and
 * cannot be a comment.
 */
export const FORBIDDEN_FONT_HOSTS = [
  { host: 'fonts.googleapis.com', declaration: 'folio:font-host-declaration' },
  { host: 'fonts.gstatic.com', declaration: 'folio:font-host-declaration' },
]

/** The marker a line must carry, IN CODE, to declare a host deliberately. */
export const DECLARATION_MARKER = 'folio:font-host-declaration'

/**
 * The trees the scan reads: THE SOURCE THAT BUILDS THE PRODUCT, named
 * positively rather than by exclusion so the population is a decision somebody
 * made rather than whatever survived a filter.
 *
 * WHAT IS OUT, AND WHAT WAS MEASURED THERE (D-8.5.5 — a scan's claim is
 * bounded, and the bound is stated rather than implied):
 *
 *   - `_bmad-output/` is the planning and implementation RECORD. It is where
 *     this repository writes down that these hosts are forbidden — the spec for
 *     this very story names both of them — so scanning it would make the record
 *     of the rule a violation of it. Measured at Story 8.5: 15 occurrences,
 *     all in archived UX mockup HTML from 2026-08-23 and in the story artifacts
 *     that quote them.
 *   - `docs/` is published prose. Measured at Story 8.5: 3 occurrences, all in
 *     `docs/expression-reference.html`, which really does link a Google Fonts
 *     stylesheet. That is a PRE-EXISTING fact about a documentation page, it
 *     predates this story, and this story may not fix it — it is reported
 *     rather than swept up, and rather than left implied by a silent exclusion.
 *   - `fixtures/` and `test-data/` hold PDFs and `.folio` documents, no source.
 *
 * NOTHING THAT BUILDS, TESTS OR SHIPS THE PRODUCT IS EXCLUDED. The designer,
 * the engine, the lint module, the hash matrix, the font tools and the CI
 * workflows are all in.
 */
export const SCANNED_ROOTS = ['folio-designer', 'folio-go', 'lint', 'hashmatrix', 'tools', '.github']

/**
 * The file classes the scan reads within those trees: source, in every language
 * this repository writes, plus the manifests and stylesheets a font host would
 * realistically be typed into.
 */
// `.mts`/`.cts` ARE HERE BECAUSE THE FIRST ONE ARRIVED WITH THIS SCAN.
// `scripts/forbidden-font-hosts.d.mts` — this module's own type sidecar — is
// the repository's only `.mts` file, and it sat OUTSIDE the population this
// scanner's comment claims covers "the source that builds the product". A
// blind spot the guard's own change created is the cheapest kind to leave in.
export const SCANNED_EXTENSIONS = ['.ts', '.tsx', '.mts', '.cts', '.js', '.jsx', '.mjs', '.cjs', '.go', '.css', '.html', '.json', '.yml', '.yaml', '.sh', '.py', '.mod', '.sum']

/** Extensions whose language comments a line out with `#` rather than `//`. */
const HASH_COMMENT_EXTENSIONS = ['.yml', '.yaml', '.sh', '.py']

/**
 * `source` with every COMMENT's characters replaced by spaces, newlines kept,
 * so line and column numbers still line up with the original.
 *
 * ⚠ THIS IS NOT APPLIED TO THE SCAN. Read the direction carefully, because
 * getting it backwards is the whole defect this guard is shaped around: the
 * SCAN reads the RAW text, so a host commented out still fails. What is
 * stripped is the EXEMPTION — a `folio:font-host-declaration` written inside a
 * comment declares nothing, because a comment is exactly where someone parks a
 * host they mean to use later. Comments are stripped FROM THE EXEMPTION, NOT
 * FROM THE SCAN.
 *
 * DELETE THIS STEP — pass the raw text to `exemptLineNumbers` instead — and the
 * comment-direction test reds on its own message. That is the mutation proof;
 * a guard that survives its own deletion is decoration.
 *
 * A character scanner rather than a regex, because a regex cannot tell
 * `// a comment` from the `//` inside `https://fonts.example`, and getting that
 * backwards makes the exemption fire on the very string it must not.
 */
export function blankComments(source, extension = '.ts') {
  const hash = HASH_COMMENT_EXTENSIONS.includes(extension)
  let out = ''
  let index = 0
  let quote
  const skip = (from) => { for (let at = from; at < source.length && source[at] !== '\n'; at++) out += ' '; return source.indexOf('\n', from) === -1 ? source.length : source.indexOf('\n', from) }
  while (index < source.length) {
    const char = source[index]
    if (quote !== undefined) {
      out += char
      if (char === '\\') { out += source[index + 1] ?? ''; index += 2; continue }
      if (char === quote || (char === '\n' && quote !== '`')) quote = undefined
      index++
      continue
    }
    if (char === '"' || char === '\'' || char === '`') { quote = char; out += char; index++; continue }
    if (char === '/' && source[index + 1] === '/') { index = skip(index); continue }
    if (hash && char === '#') { index = skip(index); continue }
    if (char === '<' && source.startsWith('<!--', index)) {
      const end = source.indexOf('-->', index)
      const stop = end === -1 ? source.length : end + 3
      for (let at = index; at < stop; at++) out += source[at] === '\n' ? '\n' : ' '
      index = stop
      continue
    }
    if (char === '/' && source[index + 1] === '*') {
      const end = source.indexOf('*/', index + 2)
      const stop = end === -1 ? source.length : end + 2
      for (let at = index; at < stop; at++) out += source[at] === '\n' ? '\n' : ' '
      index = stop
      continue
    }
    out += char
    index++
  }
  return out
}

/**
 * The 1-based line numbers on which `host` is DECLARED rather than used: the
 * line carries both the host and `DECLARATION_MARKER` **with comments blanked**.
 *
 * The marker therefore has to be real code — a string literal, an identifier —
 * and cannot be `// folio:font-host-declaration`. That asymmetry is the point:
 * anyone can write a comment.
 */
export function exemptLineNumbers(source, host, extension = '.ts') {
  const stripped = blankComments(source, extension)
  const lines = stripped.split('\n')
  const exempt = new Set()
  for (let index = 0; index < lines.length; index++) {
    if (lines[index].includes(host) && lines[index].includes(DECLARATION_MARKER)) exempt.add(index + 1)
  }
  return exempt
}

/**
 * Every occurrence of a forbidden host in `source` that is not declared, as
 * `{ host, line, text }`. The occurrence search runs over the RAW source.
 */
export function occurrencesIn(source, extension = '.ts') {
  const lines = source.split('\n')
  const found = []
  for (const { host } of FORBIDDEN_FONT_HOSTS) {
    const exempt = exemptLineNumbers(source, host, extension)
    for (let index = 0; index < lines.length; index++) {
      if (!lines[index].includes(host) || exempt.has(index + 1)) continue
      found.push({ host, line: index + 1, text: lines[index].trim().slice(0, 160) })
    }
  }
  return found
}

/** Whether a repo-relative path is inside one of the scanned trees. */
const insideTheWalk = (path) => {
  const segments = path.split('/')
  if (segments.includes('.git') || segments.includes('node_modules')) return false
  return SCANNED_ROOTS.includes(segments[0])
}

/**
 * The git-tracked files this scan reads, repo-root-relative.
 *
 * THROWS RATHER THAN RETURNING AN EMPTY LIST, in three places. A scan that
 * could not look must never read as an all-clear — that is the failure mode
 * this whole guard is shaped around, and returning `[]` here would make every
 * assertion downstream pass over a repository nothing ever opened.
 */
export function scannedPopulation(root) {
  let listing
  try {
    listing = execFileSync('git', ['-C', root, 'ls-files', '-z'], { encoding: 'utf8', maxBuffer: 256 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'] })
  } catch (error) {
    throw new Error(`forbidden font host scan could not look: \`git ls-files\` failed in ${root} (${String(error)}), and an unobtainable population must never read as all-clear`)
  }
  const tracked = listing.split('\0').filter((entry) => entry !== '')
  if (tracked.length === 0) throw new Error(`forbidden font host scan could not look: git tracks no file at all in ${root}`)
  const files = tracked
    .filter(insideTheWalk)
    .filter((path) => SCANNED_EXTENSIONS.includes(extname(path).toLowerCase()))
    .filter((path) => existsSync(join(root, path)) && statSync(join(root, path)).isFile())
  if (files.length === 0) throw new Error(`forbidden font host scan could not look: no tracked file in ${root} under ${SCANNED_ROOTS.join(', ')} carries a scanned extension (${SCANNED_EXTENSIONS.join(' ')})`)
  return files
}

/**
 * THE POPULATION FLOOR. A count written next to the thing it counts stops being
 * true the moment the thing grows, so this is a FLOOR rather than an equality —
 * but a floor is what turns "the scan found nothing" into a claim.
 *
 * MEASURED AT STORY 8.5: 579 tracked files under `SCANNED_ROOTS` carry a
 * scanned extension — 578 before `.mts`/`.cts` joined `SCANNED_EXTENSIONS`,
 * plus the one `.mts` file in the repository, which is this module's own type
 * sidecar. The number `npm run scan:font-hosts` prints, re-measured
 * on the committed catalogue at the build gate rather than estimated. (An
 * earlier draft of this comment recorded 1,058; that figure matches neither the
 * scanned population nor the repository-wide extension count of 704, so it was
 * corrected to the measurement rather than carried forward. A count written
 * beside the thing it counts is a claim, and this one was checked.) The floor
 * sits well under the measurement on purpose: it must catch a walk that
 * collapsed, not fence in normal growth.
 */
export const POPULATION_FLOOR = 400

export function scanForbiddenFontHosts(root, { floor = POPULATION_FLOOR } = {}) {
  const files = scannedPopulation(root)
  const findings = []
  for (const path of files) {
    const source = readFileSync(join(root, path), 'utf8')
    for (const occurrence of occurrencesIn(source, extname(path).toLowerCase())) findings.push({ file: path, ...occurrence })
  }
  return { files: files.length, floor, findings }
}

export function assertNoForbiddenFontHosts(root, options = {}) {
  const result = scanForbiddenFontHosts(root, options)
  if (result.files < result.floor) {
    throw new Error(`forbidden font host scan read only ${result.files} files, under its floor of ${result.floor} — refusing to report a clean population over one this small, because a scan that read almost nothing is indistinguishable from a scan that found nothing`)
  }
  if (result.findings.length > 0) {
    const named = result.findings.map((finding) => `  ${finding.file}:${finding.line}: ${finding.host}\n    ${finding.text}`).join('\n')
    throw new Error(`forbidden font host in scanned source (${result.findings.length} occurrence(s) across ${result.files} files scanned):\n${named}\n`
      + 'The catalogue is BUNDLED AND PRECACHED (D-8.5.12): SPEC.md\'s ## Non-goals forbids a live font service, an arbitrary '
      + 'URL and a download on first use, and no middle tier was invented. A commented-out host fails too — comments are '
      + 'stripped from the EXEMPTION, not from the scan. If a line must name a host, declare it in code with the marker '
      + `\`${DECLARATION_MARKER}\` on that same line.`)
  }
  return result
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..')
  const result = assertNoForbiddenFontHosts(root)
  // The population size is REPORTED, not merely used (AC4). A scan whose only
  // evidence is "it passed" says nothing about what it looked at. The claim is
  // bounded on purpose (D-8.5.5): about the scanned population, never about
  // what leaves the machine.
  console.log(`forbidden font host scan: ${result.findings.length} occurrence(s) in ${result.files} tracked source files under ${relative(process.cwd(), root) || '.'} (floor ${result.floor}). `
    + 'This bounds the SCANNED POPULATION only; it is not a claim that no request leaves the machine.')
}
