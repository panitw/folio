import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, statSync } from 'node:fs'
import { dirname, extname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

// STORY 16.2 — THE HOST-FONT SCAN.
//
// WHAT IT GUARDS. SPEC-fonts' Non-goal *"No host fonts. Faces installed on the
// authoring or rendering machine are never enumerated or read."* is the ONE
// clause of that Non-goal D-16.1 left standing — every other clause of it was
// reversed by an owner decision — and Story 16.2 is the story most likely to
// break it, because it introduces a dropdown group whose heading reads
// AVAILABLE LOCALLY. That group means "typefaces this designer has fetched
// before". It does not mean "typefaces this computer has", and the distance
// between those two readings is one browser API call.
//
// WHY IT IS A SOURCE SCAN OVER THE WHOLE TREE, AND NOT A UNIT TEST.
// A test that only checked the store module would pass while `App.tsx` called
// the API. The claim is about the designer, so the scan reads the designer.
//
// IT WOULD PASS VACUOUSLY ON INTRODUCTION, which is the same trap
// `forbidden-font-hosts.mjs` is shaped around and the reason this file is
// shaped like it: these spellings appear ZERO times in this repository today,
// so a scan added now is green before it is correct and would stay green if it
// scanned nothing, matched nothing, or exempted everything. Three things are
// therefore load-bearing, and `src/host-font-access.test.ts` has one test each:
// a POSITIVE CONTROL, a POPULATION FLOOR, and the COMMENT DIRECTION.
//
// THE CLAIM IS BOUNDED, and every message below is worded to claim the bounded
// thing (D-8.5.5): what a green here means is that none of these spellings
// appears in the scanned population. It is not a proof that no host font is
// ever read — a source scan cannot see a name assembled at runtime or a call
// made by a dependency.
//
// ⚠ THE SPELLINGS ARE NOT WRITTEN IN ANY COMMENT, HERE OR ANYWHERE. The scan
// reads RAW source, so a spelling named in prose is an occurrence like any
// other; the exemption is computed over COMMENT-BLANKED source, exactly as the
// host scan's is, so a marker inside a comment declares nothing. Read the
// spellings off the array below.

/**
 * THE LOCAL FONT ACCESS SURFACE, BY EVERY SPELLING IT GOES BY.
 *
 * Patterns rather than substrings, because one of the near neighbours in this
 * repository really does collide: `standardFontDataUrl` (PDF.js) contains the
 * interface name this API exposes, and a substring scan would report the PDF
 * preview as a host-font reader. Word boundaries are the fix, and the
 * measured collision is why they are here rather than a precaution.
 *
 * EACH ENTRY CARRIES THE DECLARATION MARKER ON ITS OWN LINE, IN CODE — that is
 * how this file, which must spell what it forbids, is not itself a violation.
 */
export const HOST_FONT_ACCESS_APIS = [
  { name: 'queryLocalFonts', pattern: /\bqueryLocalFonts\b/, declaration: 'folio:host-font-declaration' },
  { name: 'navigator.fonts', pattern: /\bnavigator\s*\.\s*fonts\b/, declaration: 'folio:host-font-declaration' },
  { name: 'local-fonts permission', pattern: /(['"`])local-fonts\1/, declaration: 'folio:host-font-declaration' },
  { name: 'FontData', pattern: /\bFontData\b/, declaration: 'folio:host-font-declaration' },
]

/** The marker a line must carry, IN CODE, to name one of these deliberately. */
export const HOST_FONT_DECLARATION_MARKER = 'folio:host-font-declaration'

/**
 * The trees the scan reads: the whole designer, named positively rather than by
 * exclusion so the population is a decision somebody made.
 *
 * `src/` is the product. `scripts/` builds it. `e2e/` drives it. All three could
 * hold a call, and the browser specs in particular are where "just check what
 * fonts the machine has" would be most tempting to write.
 */
export const SCANNED_ROOTS = ['folio-designer/src', 'folio-designer/scripts', 'folio-designer/e2e']

export const SCANNED_EXTENSIONS = ['.ts', '.tsx', '.mts', '.cts', '.js', '.jsx', '.mjs', '.cjs', '.css', '.html', '.json']

/**
 * THE POPULATION FLOOR, AND THE FRACTION IS STATED BECAUSE THE FRACTION IS THE
 * WHOLE DESIGN OF IT.
 *
 * MEASURED AT STORY 16.2: **129** tracked files under `SCANNED_ROOTS` carry a
 * scanned extension — the number `npm run scan:host-fonts` prints. (An earlier
 * draft of this comment recorded 123 and set the floor at 50. That 123 was
 * measured while this story's own six new files were still UNTRACKED, so the
 * walk — which reads `git ls-files` — could not see them; the figure was
 * corrected to the measurement taken over the committed tree rather than
 * carried forward.)
 *
 * **86 is 2/3 of 129 (66.7%)**, deliberately the same fraction this guard's
 * model runs: `forbidden-font-hosts.mjs` floors 400 against a measured 579-600,
 * which is 67-69%. A floor at 50 was ~39% of the population, which means a walk
 * that collapsed to 60 files — half the designer unread — would still have
 * reported all-clear, and an all-clear from a scan that read half the tree is
 * the exact failure this guard is shaped around.
 *
 * It is still a FLOOR and not an equality: it must catch a walk that collapsed,
 * not fence in normal growth. Two thirds leaves room for a third of these files
 * to be deleted before the floor is the thing that reds.
 */
export const POPULATION_FLOOR = 86

/** Reuses the host scan's comment blanker, because the two guards want the same asymmetry. */
export { blankComments } from './forbidden-font-hosts.mjs'
import { blankComments } from './forbidden-font-hosts.mjs'

/**
 * The 1-based line numbers on which an API is DECLARED rather than used: the
 * line carries both the spelling and the marker, WITH COMMENTS BLANKED.
 *
 * The marker therefore has to be real code and cannot be written in a comment.
 * That asymmetry is the point: anyone can write a comment, and a comment is
 * exactly where somebody parks a call they mean to make later.
 */
export function exemptLineNumbers(source, pattern, extension = '.ts') {
  const lines = blankComments(source, extension).split('\n')
  const exempt = new Set()
  for (let index = 0; index < lines.length; index++) {
    if (pattern.test(lines[index]) && lines[index].includes(HOST_FONT_DECLARATION_MARKER)) exempt.add(index + 1)
  }
  return exempt
}

/** Every undeclared occurrence in `source`. The occurrence search runs over the RAW text. */
export function occurrencesIn(source, extension = '.ts') {
  const lines = source.split('\n')
  const found = []
  for (const { name, pattern } of HOST_FONT_ACCESS_APIS) {
    const exempt = exemptLineNumbers(source, pattern, extension)
    for (let index = 0; index < lines.length; index++) {
      if (!pattern.test(lines[index]) || exempt.has(index + 1)) continue
      found.push({ api: name, line: index + 1, text: lines[index].trim().slice(0, 160) })
    }
  }
  return found
}

const insideTheWalk = (path) => {
  const segments = path.split('/')
  if (segments.includes('.git') || segments.includes('node_modules')) return false
  return SCANNED_ROOTS.some((root) => path === root || path.startsWith(`${root}/`))
}

/**
 * The git-tracked files this scan reads, repo-root-relative.
 *
 * THROWS RATHER THAN RETURNING AN EMPTY LIST. A scan that could not look must
 * never read as an all-clear; that is the failure mode this whole guard is
 * shaped around.
 */
export function scannedPopulation(root) {
  let listing
  try {
    listing = execFileSync('git', ['-C', root, 'ls-files', '-z'], { encoding: 'utf8', maxBuffer: 256 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'] })
  } catch (error) {
    throw new Error(`host-font scan could not look: \`git ls-files\` failed in ${root} (${String(error)}), and an unobtainable population must never read as all-clear`)
  }
  const tracked = listing.split('\0').filter((entry) => entry !== '')
  if (tracked.length === 0) throw new Error(`host-font scan could not look: git tracks no file at all in ${root}`)
  const files = tracked
    .filter(insideTheWalk)
    .filter((path) => SCANNED_EXTENSIONS.includes(extname(path).toLowerCase()))
    .filter((path) => existsSync(join(root, path)) && statSync(join(root, path)).isFile())
  if (files.length === 0) throw new Error(`host-font scan could not look: no tracked file in ${root} under ${SCANNED_ROOTS.join(', ')} carries a scanned extension`)
  return files
}

export function scanHostFontAccess(root, { floor = POPULATION_FLOOR } = {}) {
  const files = scannedPopulation(root)
  const findings = []
  for (const path of files) {
    for (const occurrence of occurrencesIn(readFileSync(join(root, path), 'utf8'), extname(path).toLowerCase())) findings.push({ file: path, ...occurrence })
  }
  return { files: files.length, floor, findings }
}

export function assertNoHostFontAccess(root, options = {}) {
  const result = scanHostFontAccess(root, options)
  if (result.files < result.floor) {
    throw new Error(`host-font scan read only ${result.files} files, under its floor of ${result.floor} — refusing to report a clean population over one this small, because a scan that read almost nothing is indistinguishable from a scan that found nothing`)
  }
  if (result.findings.length > 0) {
    const named = result.findings.map((finding) => `  ${finding.file}:${finding.line}: ${finding.api}\n    ${finding.text}`).join('\n')
    throw new Error(`the Local Font Access API appears in the designer's source (${result.findings.length} occurrence(s) across ${result.files} files scanned):\n${named}\n`
      + 'SPEC-fonts forbids it outright: "No host fonts. Faces installed on the authoring or rendering machine are never enumerated or read." '
      + 'The designer\'s AVAILABLE LOCALLY group means typefaces this designer has FETCHED before — src/font-store.ts — and never typefaces this computer has installed. '
      + 'A commented-out call fails too: comments are stripped from the EXEMPTION, not from the scan.')
  }
  return result
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..')
  const result = assertNoHostFontAccess(root)
  // The population is REPORTED, not merely used. A scan whose only evidence is
  // "it passed" says nothing about what it looked at, and the claim it licenses
  // is bounded: about the scanned population, never about what the running
  // browser does.
  console.log(`host-font scan: ${result.files} files scanned under ${SCANNED_ROOTS.join(', ')}, floor ${result.floor}, 0 occurrences of ${HOST_FONT_ACCESS_APIS.length} spellings`)
}
