import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const sourceDir = path.dirname(fileURLToPath(import.meta.url))
const designerRoot = path.dirname(sourceDir)
const production = fs.readdirSync(sourceDir, { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx|css)$/.test(entry) && !/\.test\.(?:ts|tsx)$/.test(entry))
  .map((entry) => path.join(sourceDir, entry))
const e2e = fs.readdirSync(path.join(designerRoot, 'e2e'), { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry))
  .map((entry) => path.join(designerRoot, 'e2e', entry))
const tests = fs.readdirSync(sourceDir, { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.test\.(?:ts|tsx)$/.test(entry) && entry !== 'canvas-authority-contract.test.ts')
  .map((entry) => path.join(sourceDir, entry))

const prohibited = [
  /(?:CanvasRenderingContext2D|\bctx)\.measureText/,
  /\b(?:getBoundingClientRect|getClientRects)\s*\(/,
  /\b(?:offset(?:Width|Height|Left|Top|Parent)|client(?:Width|Height|Left|Top)|scroll(?:Width|Height|Left|Top))\b/,
  /\boffset[XY]\b/,
  /\bResizeObserver\b/,
  // STORY 8.4a REPAIRED THIS RULE, WHICH HAD BEEN DEAD SINCE 7bfb076.
  // `violations()` used to rewrite EVERY `document.fonts` prefix in every file
  // to a throwaway token before applying any pattern, so this line could not
  // match anything at all: `document.fonts.add(face)` read as
  // `fontReadinessOnly.add(face)`. Three things confirmed it was dead rather
  // than merely broad — the token appeared nowhere else in the repository, the
  // mutation block below proved eleven other prohibitions and not this one, and
  // the rewrite had been appended as a drive-by unblock in a story about sample
  // data. The rewrite is now scoped to `document.fonts.ready` ALONE (readiness
  // is not measurement, and e2e/engine-worker.spec.ts legitimately awaits it),
  // so every other member — `add`, `delete`, `clear`, `load`, `check`, and
  // iteration — is caught here, and the proof that it is caught is in this
  // file rather than assumed.
  /\bdocument\.fonts\b/,
  // The other half of the same door, which was never guarded at all: a face can
  // be registered without touching `document.fonts` by name. Story 8.4a's seam
  // is carved out below, by file name and only while it still spells the
  // function that earns the exception.
  /\bnew FontFace\b/,
  /\bdevicePixelRatio\b/,
  /\b(?:Range|document\.createRange|getSelection|Selection)\s*\(/,
  /\bgetComputedStyle\s*\(/,
  /(?:white-space|text-wrap|overflow-wrap|word-break|line-clamp|text-align)\s*:\s*(?:normal|wrap|balance|pretty|anywhere|break-word|justify|\d+)/,
  // D-7.4.2 §5. The canvas paint is APPROXIMATE and, since Story 7.4, may be
  // a deliberately truncated prefix; pagination is EXACT and comes from
  // layout.Paginate. Deriving a height, a page count or a window count from
  // the paint's line count would make a truncated paint shorten the reported
  // column, and Story 7.6 would then draw the wrong number of sheets — the
  // canvas lying about pagination. The independence holds in Go today by
  // construction (nothing there reads CanvasTextPaint), but the data sits
  // right here in `line.advance`, one plausible line of designer code away.
  // A positive test would not catch that line being written; this does.
  /\b(?:textPaint|paint)\??\.lines\.length\b/,
  /\blines\.length\s*[*/]|[*/]\s*\blines\.length\b/,
  // Story 7.6 / AC2, and the same guard one step further along. A window's
  // COUNT may not come from the paint; a window's POSITION may not come from
  // arithmetic at all. The window height multiplied by an index is the closed form
  // internal/layout/paginate.go forbids by name — the window advances to the
  // top of the first item that did not fit, never by a fixed height — and it
  // is measurably wrong: 110 millipoints per window adrift on a column of
  // round 728pt spacing, and eleven windows where the engine says two on a
  // column with a declared gap. The origins are PROJECTED
  // (contentWindowOrigins); multiplying is the one plausible line of designer
  // code that would quietly replace them, in either operand order.
  /\b(?:contentWindowHeight|windowHeight)\s*\*|\*\s*[\w.?]*\b(?:contentWindowHeight|windowHeight)\b/,
]

// STORY 8.2. THE SECOND LOCK ON "THE BROWSER HOLDS NO ENGINE RULE".
//
// A chain edit is refused by the engine in the engine's own sentence, and the
// panel's only job is to place it. The failure mode that closes off is a
// TypeScript COPY of one of those rules — a duplicate-name check, an
// empty-chain check, an orphan check, a message table — which would pass every
// behavioural test the panel has, because "the error shows" is satisfied
// equally well by a local copy as by the engine's answer. App.test.tsx's
// anti-pre-emption assertions prove a command is still DISPATCHED for each of
// those cases; this scan proves the SENTENCES are not here.
//
// Each literal is a fragment of a real refusal in
// folio-go/component_commands.go: `a font chain named %q already exists`,
// `font chain %q is still named by ...`, `removing that entry would leave font
// chain %q with no entries`, `a font chain must declare at least one entry`,
// `entry index is out of range`, `font chain name exceeds the projection
// bound`, `no font chain named %q is declared`.
//
// SCOPE: PRODUCTION SOURCE ONLY, and deliberately so. A test that asserts the
// rendered refusal is `===` to the engine's `message` must be able to WRITE
// that message as a fixture — that is the evidence, not a copy of the rule —
// so the unit-test and e2e corpora are not scanned for these. Production is
// where a rule would have to live to reach an author.
//
// AND COMMENTS ARE STRIPPED FIRST. These sentences are ordinary English, and
// scanning raw file text made the guard collide with prose that has nothing to
// do with an engine rule: `createComponent` ALREADY EXISTS on the channel, the
// sheet gap IS DECLARED here as a number. Two unrelated comments were reworded
// to get the first version of this scan green, which is the guard editing the
// codebase rather than the codebase answering to the guard — and it would have
// happened again on every future `is declared` or `is out of range`. A COPY of
// an engine rule has to be a string or a template literal to reach an author,
// so those are the only places worth looking, and the two comments above have
// been restored to their original wording as this fix's own proof.
const refusalVocabulary = [
  /already exists/,
  /is still named by/,
  /with no entries/,
  /must declare at least one entry/,
  /is out of range/,
  /exceeds the projection bound/,
  /is declared/,
  // The rest of component_commands.go's font-chain vocabulary. These were
  // omitted while the scan read raw text because several of them are also
  // plausible English; with comments stripped they cost nothing to add, and
  // leaving them out left five refusals a browser copy could have used.
  /font chain entries are required/,
  /must be a string array/,
  /must be a non-empty string/,
  /declares more entries than the projection bound/,
  /declares more font chains than the projection bound/,
]

// withoutComments removes line and block comments while leaving string and
// template literals intact. It is a character scanner rather than a regex
// because a regex cannot tell `// a comment` from the `//` inside a URL string
// — and getting that backwards would make the guard vacuous in exactly the
// place it matters. Quotes are checked before comment openers, so an
// apostrophe inside a comment never opens a string and a `//` inside a string
// never opens a comment.
function withoutComments(source: string): string {
  let out = ''
  let index = 0
  let quote: string | undefined
  while (index < source.length) {
    const char = source[index] as string
    if (quote !== undefined) {
      out += char
      if (char === '\\') { out += source[index + 1] ?? ''; index += 2; continue }
      if (char === quote) quote = undefined
      index++
      continue
    }
    if (char === '"' || char === '\'' || char === '`') { quote = char; out += char; index++; continue }
    if (char === '/' && source[index + 1] === '/') { while (index < source.length && source[index] !== '\n') index++; continue }
    if (char === '/' && source[index + 1] === '*') { index += 2; while (index < source.length && !(source[index] === '*' && source[index + 1] === '/')) index++; index += 2; continue }
    out += char
    index++
  }
  return out
}

function refusalViolations(files: readonly string[]): string[] {
  return files.flatMap((file) => {
    const source = withoutComments(fs.readFileSync(file, 'utf8'))
    const name = path.relative(designerRoot, file)
    return refusalVocabulary.filter((pattern) => pattern.test(source)).map((pattern) => `${name}: ${pattern}`)
  })
}

// scanned is what every prohibition is actually applied to: the file's own
// text with the approved, individually justified exceptions removed and its
// COMMENTS STRIPPED. Comment stripping arrived with Story 8.4a and for the
// same reason it was already applied to the refusal vocabulary — these
// patterns are describable in English, and a guard that reddens on the prose
// explaining what not to write ends up editing the codebase instead of
// answering to it. `does not catch ordinary English in a comment` proves both
// directions.
//
// The pointer-input exception runs FIRST, on raw text, because its seam was
// written against the file as it stands and it asserts that seam is still
// there.
function scanned(file: string, source: string): string {
  return withoutApprovedRuntimeFaceRegistration(file, withoutComments(withoutApprovedLocalPointerInput(file, source)))
}

function violations(files: readonly string[]): string[] {
  return files.flatMap((file) => {
    const source = scanned(file, fs.readFileSync(file, 'utf8'))
    const name = path.relative(designerRoot, file)
    return prohibited.filter((pattern) => pattern.test(source)).map((pattern) => `${name}: ${pattern}`)
  })
}

describe('canvas projection authority contract', () => {
  it('scans a non-vacuous production, unit-test, and e2e corpus for browser measurement authority', () => {
    expect(production.length).toBeGreaterThan(10)
    expect(tests.length).toBeGreaterThan(10)
    expect(e2e.length).toBeGreaterThan(3)
    expect(violations([...production, ...tests, ...e2e])).toEqual([])
  })

  it('keeps the engine\'s own refusal vocabulary out of every production source file', () => {
    expect(production.length).toBeGreaterThan(10)
    expect(refusalViolations(production)).toEqual([])
  })

  it('turns a TypeScript copy of any engine refusal red', () => {
    // One realistic line per literal — the shape a browser-side rule would
    // actually take — so the scan is proved by its own reds, not assumed.
    expect(refusalForSource('if (chains.some((chain) => chain.name === name)) return `a font chain named "${name}" ' + 'already exists`')).not.toEqual([])
    expect(refusalForSource('const orphan = `font chain "${name}" ' + 'is still named by ${ids.join(", ")}`')).not.toEqual([])
    expect(refusalForSource('const empty = `removing that entry would leave font chain "${name}" ' + 'with no entries`')).not.toEqual([])
    expect(refusalForSource('const none = "a font chain ' + 'must declare at least one entry"')).not.toEqual([])
    expect(refusalForSource('const range = "entry index ' + 'is out of range"')).not.toEqual([])
    expect(refusalForSource('const bound = "font chain name ' + 'exceeds the projection bound"')).not.toEqual([])
    expect(refusalForSource('const missing = `no font chain named "${name}" ' + 'is declared`')).not.toEqual([])
    expect(refusalForSource('const required = "font chain ' + 'entries are required"')).not.toEqual([])
    expect(refusalForSource('const shape = "font chain entries ' + 'must be a string array"')).not.toEqual([])
    expect(refusalForSource('const face = "a font chain entry ' + 'must be a non-empty string"')).not.toEqual([])
    expect(refusalForSource('const many = "a font chain ' + 'declares more entries than the projection bound"')).not.toEqual([])
    expect(refusalForSource('const chains = "document ' + 'declares more font chains than the projection bound"')).not.toEqual([])
  })

  // THE OTHER HALF, and the one that keeps the guard from editing the prose it
  // is scanning. These are the two REAL comments in this repository that the
  // raw-text version of this scan reddened, restored verbatim.
  it('does not catch ordinary English in a comment', () => {
    expect(refusalForSource('// be created there rather than refused. `createComponent` already exists on\n// the channel, already carries the band NAME')).toEqual([])
    expect(refusalForSource('// between two sheets is declared HERE as a number and written out as a custom')).toEqual([])
    expect(refusalForSource('/* the index is out of range for this band, so nothing is drawn */')).toEqual([])
    // But a string on the same line as a comment is still read.
    expect(refusalForSource('const message = "already exists" // a comment')).not.toEqual([])
    // And a comment marker inside a STRING does not hide what follows it.
    expect(refusalForSource('const url = "https://example.test/x"; const message = "is out of range"')).not.toEqual([])
  })

  it('allows only the non-document reduced-motion media rule', () => {
    const css = fs.readFileSync(path.join(sourceDir, 'App.css'), 'utf8')
    expect([...css.matchAll(/@media\s*\(([^)]+)\)/g)].map((match) => match[1])).toEqual(['prefers-reduced-motion: reduce'])
  })

  it('turns realistic measurement, CSS, range, and event-coordinate mutations red', () => {
    expect(violationsForSource(`const width = ${['CanvasRenderingContext2D', 'measureText'].join('.')}("x")`)).not.toEqual([])
    expect(violationsForSource('.paint { white-space: normal }')).not.toEqual([])
    expect(violationsForSource(`const x = event.${['offset', 'X'].join('')}`)).not.toEqual([])
    expect(violationsForSource('const style = getComputedStyle(node)')).not.toEqual([])
    expect(violationsForSource(`const range = ${['document', 'createRange'].join('.') }()`)).not.toEqual([])
    expect(violationsForSource('.paint { text-align: justify }')).not.toEqual([])
    expect(violationsForSource('const height = component.textPaint.lines.length * line.advance')).not.toEqual([])
    expect(violationsForSource('const windows = Math.ceil(paint.lines.length / perPage)')).not.toEqual([])
    expect(violationsForSource('const top = canvas.contentWindowHeight * index')).not.toEqual([])
    expect(violationsForSource('const top = index * canvas.contentWindowHeight')).not.toEqual([])
    expect(violationsForSource('const top = sheet * windowHeight')).not.toEqual([])
  })

  // STORY 8.4a. THE MUTATION PROOFS THE `document.fonts` RULE NEVER HAD —
  // which is the defect being repaired, not a decoration on it. Before this,
  // deleting that rule outright left every assertion in this file green, so
  // the rule was indistinguishable from its own absence.
  it('turns runtime font registration red, allows readiness, and allows it only in the one approved seam', () => {
    expect(violationsForSource('document.fonts.add(face)')).not.toEqual([])
    expect(violationsForSource('document.fonts.delete(face)')).not.toEqual([])
    expect(violationsForSource('await document.fonts.load("12px x")')).not.toEqual([])
    expect(violationsForSource('if (document.fonts.check("12px x")) paint()')).not.toEqual([])
    expect(violationsForSource('const face = new FontFace(family, bytes)')).not.toEqual([])
    // READINESS STAYS LEGAL, in the one place that actually uses it and
    // anywhere else: it registers nothing and measures nothing.
    expect(violationsForFile('e2e/engine-worker.spec.ts', 'await document.fonts.ready')).toEqual([])
    // AND THE PROSE TAX IS GONE. A comment that names the mechanism is not the
    // mechanism, in either file.
    expect(violationsForSource('// the seam calls document.fonts.add once per carried face')).toEqual([])
    expect(violationsForSource('/* never write new FontFace outside that module */')).toEqual([])
  })

  // The seam exception is scoped to a file AND to the function inside it. Both
  // halves are proved: the same line is legal there and illegal everywhere
  // else, and it stops being legal there the moment the function it is
  // attached to is gone.
  it('scopes the registration exception to the named seam and to the function that earns it', () => {
    const seam = 'src/embedded-face-registry.ts'
    const registration = 'export function registerCarriedFaces(keys) {\n  const face = new FontFace(family, bytes)\n  document.fonts.add(face)\n}\n'
    expect(violationsForFile(seam, registration)).toEqual([])
    expect(violationsForFile('src/some-canvas-component.tsx', registration)).not.toEqual([])
    // OUTSIDE the approved function, inside the approved file.
    expect(violationsForFile(seam, `const stray = new FontFace(family, bytes)\n${registration}`)).not.toEqual([])
  })
})

function violationsForSource(source: string): RegExp[] { return violationsForFile('src/an-ordinary-module.ts', source) }
// The same scan a real file gets, addressed by NAME, so an exception that is
// scoped to a file can be proved to hold there and to hold nowhere else.
function violationsForFile(file: string, source: string): RegExp[] { return prohibited.filter((pattern) => pattern.test(scanned(path.join(designerRoot, file), source))) }
function refusalForSource(source: string): RegExp[] { return refusalVocabulary.filter((pattern) => pattern.test(withoutComments(source))) }

// THE TWO SCOPED EXCEPTIONS TO THE FONT RULES, each written the way
// withoutApprovedLocalPointerInput is: narrow to a named owner, and asserted
// to still have the thing that earns it, so the carve-out cannot outlive its
// reason.
function withoutApprovedRuntimeFaceRegistration(file: string, source: string): string {
  // READINESS IS NOT REGISTRATION AND NOT MEASUREMENT, and it is legal
  // everywhere. e2e/engine-worker.spec.ts awaits it so that in-flight requests
  // for the BUILD-TIME faces are not counted as offline failures; nothing is
  // measured, nothing is added, and no layout waits on it.
  const readiness = source.replace(/document\.fonts\.ready\b/g, 'fontReadinessOnly')
  // THE ONE SEAM THAT MAY REGISTER A FACE WHILE A DOCUMENT IS OPEN (Story
  // 8.4a). A face the DOCUMENT carries exists only inside that document, so it
  // cannot be declared at build time; it is fetched over the engine's own
  // `asset` operation and added to the page's font set for as long as that
  // document is open. The exception is the FUNCTION, not the file: everything
  // outside it is scanned normally, and if that function is renamed or removed
  // the exception dies with it. canvas-font-stack.test.ts separately asserts
  // this is the only site in the whole designer.
  if (path.basename(file) === 'embedded-face-registry.ts') {
    const seam = /export function registerCarriedFaces\([\s\S]*?\n}\n/
    expect(readiness).toMatch(seam)
    return readiness.replace(seam, '')
  }
  // THE DETECTOR'S OWN FIXTURES. canvas-font-stack.test.ts is the test that
  // proves runtime registration happens in exactly one place, and it cannot
  // prove its scanner detects a mechanism without spelling that mechanism. The
  // exception is narrowed to the two font spellings — every other prohibition
  // still applies to that file — and it holds only while the detector is
  // actually there.
  if (path.basename(file) === 'canvas-font-stack.test.ts') {
    expect(readiness).toMatch(/function registersAFaceAtRuntime\(source: string\): boolean/)
    return readiness.replace(/new FontFace\b/g, 'faceDetectorFixture').replace(/document\.fonts\b/g, 'fontSetDetectorFixture')
  }
  return readiness
}

function withoutApprovedLocalPointerInput(file: string, source: string): string {
  if (file.includes(`${path.sep}preview${path.sep}`)) {
    // PDF viewer scroll is deliberately transient viewer navigation, never a
    // document/canvas measurement. Keep that exception narrow to this owner.
    return source.replace(/scroll(?:Width|Height|Left|Top)\b/g, 'viewerTransientState')
  }
  if (path.basename(file) !== 'App.tsx') return source
  // The sole approved pointer coordinate is isolated to a named transient
  // proposal helper. It is not DOM measurement and never reaches paint.
  const seam = /export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/
  expect(source).toMatch(seam)
  return source.replace(seam, 'function pageStyle')
}
