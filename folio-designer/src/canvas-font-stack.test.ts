import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { createElement } from 'react'
import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { TextPaint } from './App'
import { embeddedFaceFamily } from './embedded-face-family'
import type { CanvasProjection } from './engine-protocol'

// The canvas paints each engine-supplied fragment as an absolutely
// positioned span at the x the ENGINE measured. The browser therefore
// contributes rasterization only (AD-17) — but it must rasterize with the
// SAME faces the engine measured with, or every fragment's drawn width
// disagrees with the x of the fragment after it and the two collide.
//
// THE DEFECT THIS PINS, because it shipped and a person reported it.
// `scripts/build-wasm.mjs` registered the three shipped Noto faces under
// IBM PLEX family names only (the design system's vocabulary). `App.css`
// asked for them under NOTO names, which nothing declared a face for, so the
// browser fell through to generic `sans-serif` — a system Thai face with
// different metrics. Latin looked fine (the fallback's Latin is close enough
// to pass a glance); Thai overlapped at exactly the fragment boundaries,
// which sit at spaces. Reported as "letters rendered on top of each other"
// around "พระราชบัญญัติ การทวงถามหนี้".
//
// AND WHAT STORY 8.4b CHANGED. Asking under the Noto names is now CORRECT:
// the generator declares the same three files a second time under the
// engine's own face names, so the stack below names exactly what the engine
// measures with. The historical defect is preserved here because the guards
// in this file are shaped by it — the failure was silent, cosmetic-looking,
// and script-dependent, which is why every claim here is a checked one.
//
// WHY THIS TEST READS THE GENERATOR AND NOT ITS OUTPUT.
// `src/generated/runtime-fonts.css` is gitignored and only exists after
// `build:wasm`. Asserting against it would make this guard's strength
// depend on build order, and a missing file is the classic way a guard
// goes quietly vacuous. Both files read here are tracked sources.
const here = path.dirname(fileURLToPath(import.meta.url))
const generatorPath = path.join(here, '..', 'scripts', 'build-wasm.mjs')
const cssPath = path.join(here, 'App.css')
// THE ENGINE'S OWN AUTHORITY FOR A FACE NAME, read rather than restated.
// `fonts.Shipped()` is the single machine-readable enumeration of the faces
// this build measures with; there is no exported constant, JSON registry or
// generated artifact carrying the same three names. An earlier form of the
// test below hardcoded them, which meant it would have gone on passing —
// while having become false — the moment folio-go shipped a different set.
const enginePath = path.join(here, '..', '..', 'folio-go', 'fonts', 'fonts.go')
// The design system's own vocabulary, read rather than restated: the three
// IBM Plex families must remain named by a `--font-*` token here, or the two
// vocabularies this story deliberately keeps apart have been collapsed.
const tokensPath = path.join(here, 'tokens.css')

/** The face names `fonts.Shipped()` keys its FontSet by, in the order it writes them. */
function shippedFaceNames(fontsGo: string): ReadonlyArray<string> {
  const body = /func Shipped\(\) folio\.FontSet \{[\s\S]*?\n\}/.exec(fontsGo)?.[0]
  if (body === undefined) throw new Error(`no Shipped() function in ${enginePath}`)
  return [...body.matchAll(/"([^"]+)":\s*\w+,/g)].map((m) => m[1])
}

// THE ONE MODULE ALLOWED TO REGISTER A FACE WHILE A DOCUMENT IS OPEN
// (Story 8.4a). Named here rather than described, because the claim these
// tests make is not "registration is rare" but "registration is HERE".
const runtimeRegistrationSeam = 'embedded-face-registry.ts'

/**
 * Family names the generator actually declares an @font-face for.
 *
 * HAND IT `withoutComments(...)` OUTPUT, ALWAYS. It reads TEXT, so a rule that
 * has been commented out reads exactly like a live one — measured, not feared:
 * commenting out the three engine-named rules and leaving them as comment text
 * left every test in this file and in font-binary-identity.test.ts green while
 * the emitted stylesheet dropped to three rules, reproducing the very
 * Thai-overlap defect these guards exist to prevent. The red-proof for that is
 * its own test below.
 */
function declaredFamilies(generator: string): ReadonlyArray<string> {
  return [...generator.matchAll(/@font-face \{ font-family: '([^']+)'/g)].map((m) => m[1])
}

/**
 * Every `--font-*` custom property in tokens.css, with the value it is given.
 *
 * `var(--font-sans)` inside a `--type-*` token is deliberately not matched: the
 * pattern requires the `:` of a DECLARATION, so this answers to where a family
 * is NAMED rather than to where one is referenced.
 */
function fontTokenValues(tokens: string): ReadonlyArray<readonly [string, string]> {
  return [...tokens.matchAll(/(--font-[\w-]+)\s*:\s*([^;}]+)/g)].map((m) => [m[1], m[2].trim()] as const)
}

/** Quoted families the canvas fragment rule asks for, in order. */
function requestedFamilies(css: string): ReadonlyArray<string> {
  const rule = css.split('\n').find((line) => line.startsWith('.canvas-text-fragment {'))
  if (rule === undefined) throw new Error('no .canvas-text-fragment rule in App.css')
  const declaration = /font-family:([^;]+);/.exec(rule)
  if (declaration === null) throw new Error('.canvas-text-fragment declares no font-family')
  return [...declaration[1].matchAll(/'([^']+)'/g)].map((m) => m[1])
}

/**
 * Whether a source registers a font face AT RUNTIME — the mechanism Story 8.4a
 * needs and this build has none of.
 *
 * Three spellings, because the disclosure names three: the `FontFace`
 * constructor, `document.fonts.add`, and an `@font-face` rule injected as text
 * whose `src` is a `data:` or `blob:` URL (a build-time `@font-face` points at
 * a bundled asset path, so the URL scheme is what separates the two).
 */
function registersAFaceAtRuntime(source: string): boolean {
  if (/new FontFace\b|document\.fonts\.add\b/.test(source)) return true
  // The injected-rule form: an `@font-face` and, within the same rule, a `src`
  // fed from a data/blob URL. Bounded rather than greedy so a build-time
  // `@font-face` early in a file cannot pair with an unrelated `data:` URL far
  // below it.
  return /@font-face[\s\S]{0,400}?src\s*:[^;}]{0,200}?(?:data:|blob:)/.test(source)
}

// withoutComments strips line and block comments while leaving string and
// template literals intact, so a scan over source text answers to the CODE
// rather than to the prose describing it.
//
// IT IS DUPLICATED FROM canvas-authority-contract.test.ts, DELIBERATELY, and
// the alternatives are both worse. Importing it from that file would register
// its whole suite a second time under this one; hoisting it into a shared
// non-test module would put a test helper into `src/`, where it would enter
// the very production corpus these scans walk. A character scanner rather than
// a regex because a regex cannot tell `// a comment` from the `//` inside a
// URL string, and getting that backwards makes the guard vacuous exactly where
// it matters.
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

/** Every non-test designer source, PAIRED WITH ITS NAME so a scan can say where. */
function designerSources(): ReadonlyArray<readonly [string, string]> {
  return fs.readdirSync(here, { recursive: true })
    .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry) && !/\.test\.(?:ts|tsx)$/.test(entry))
    .map((entry) => [entry, fs.readFileSync(path.join(here, entry), 'utf8')] as const)
}

// The named sources that register a face at runtime, ANSWERING TO THE CODE
// RATHER THAN TO THE PROSE. Comments are stripped before the detector sees
// them, and that is not tidiness: measured, this file's own seam module
// describes its exception in a comment that spells `new FontFace`, so a raw
// scan reported the seam as registering even after the registration had been
// taken out of it — a guard that would have stayed green through the removal
// of the thing it guards. The detector itself is untouched (it is proved
// against its own fixtures below); only what is handed to it is.
function runtimeRegistrationSites(sources: ReadonlyArray<readonly [string, string]>): string[] {
  return sources.filter(([, source]) => registersAFaceAtRuntime(withoutComments(source))).map(([name]) => name)
}

// Every CSS font-family DECLARATION position in a TypeScript source: the
// `fontFamily:` of an inline style object and the `font-family:` of a CSS
// string, with the value it is being given. It deliberately does not match
// `fontFamily?:` (a type member), `'fontFamily',` (a key in a list) or
// `'fontFamily' |` (a union member) — none of those declares anything to the
// browser, and the projection's `fontFamily` field, which names a document's
// CHAIN, is read all over this codebase without ever reaching CSS.
//
// The lookbehind is measured rather than defensive: without it the scan read
// `'property-error-fontFamily' : undefined` — a ternary on an identifier that
// merely ENDS in the word — as a declaration of `undefined`, which is a green
// guard reporting a position that does not exist and a red one the moment an
// unrelated id is renamed.
function fontFamilyDeclarations(source: string): ReadonlyArray<string> {
  return [...withoutComments(source).matchAll(/(?<![\w-])(?:font-family|fontFamily)['"\]]?\s*:\s*([^,;}\n]*)/g)].map((match) => match[1].trim())
}

// The ONLY value a font-family position may be given in designer source: the
// family derived from the asset key the ENGINE attributed this fragment to.
const assetKeyDerivedFamily = /^embeddedFaceFamily\([A-Za-z][A-Za-z0-9_]*\.assetKey\)$/

function unapprovedFontFamilyDeclarations(source: string): ReadonlyArray<string> {
  return fontFamilyDeclarations(source).filter((value) => !assetKeyDerivedFamily.test(value))
}

// A 64-character asset key, the shape the engine projects and the format's own
// rule produces (the lowercase hex SHA-256 of the face's decoded bytes).
const carriedKey = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef'

// paintedFragmentFamilies renders the real component the canvas renders and
// reads the family off the DOM. It is the one assertion in this file that is
// not a text scan: what the browser is ASKED for is a rendered fact, and a
// scan of App.tsx could only ever say that a plausible line is present.
function paintedFragmentFamilies(assetKey: string | undefined, registered: ReadonlySet<string>): ReadonlyArray<string> {
  const component = {
    id: 'e1', type: 'text', band: 'content', x: 0, y: 0, width: 72_000, height: 24_000, resizable: true,
    textPaint: { overflow: false, truncated: false, lines: [{ top: 0, baseline: 10_000, advance: 12_000, width: 30_000, fragments: [{ text: 'สัญญา', x: 0, ...(assetKey === undefined ? {} : { assetKey }) }] }] },
  } as unknown as CanvasProjection['components'][number]
  const { container } = render(createElement(TextPaint, { component, carriedFaces: registered, zoom: 1 }))
  return Array.from(container.querySelectorAll('.canvas-text-fragment')).map((node) => (node as HTMLElement).style.fontFamily)
}

describe('the canvas paints with the faces the engine measured', () => {
  const generator = fs.readFileSync(generatorPath, 'utf8')
  const css = fs.readFileSync(cssPath, 'utf8')
  // COMMENTS STRIPPED BEFORE THE GENERATOR IS PARSED, for the same measured
  // reason the designer-source scans below strip them: the parse must answer to
  // the CODE that emits the stylesheet, not to prose that merely spells a rule.
  const declared = declaredFamilies(withoutComments(generator))
  const requested = requestedFamilies(css)
  const engineFaces = shippedFaceNames(fs.readFileSync(enginePath, 'utf8'))

  // Vacuity guard: neither side may be empty, or the assertion below
  // passes by having nothing to compare. A regex that stops matching
  // because a file's shape changed is the failure mode this catches.
  //
  // THE FLOOR IS SIX, RAISED FROM THREE BY STORY 8.4b, because six is the
  // true count: three files, each declared twice — once under the design
  // system's family name and once under the engine's own face name. A floor
  // that sits below the real number has stopped discriminating; at three it
  // would have survived the deletion of the entire engine-named half.
  it('reads a non-empty declaration set from the generator', () => {
    expect(declared.length).toBeGreaterThanOrEqual(6)
  })

  it('reads a non-empty request list from the canvas rule', () => {
    expect(requested.length).toBeGreaterThanOrEqual(3)
  })

  // THE GENERATOR PARSE ANSWERS TO THE CODE, NOT TO THE PROSE.
  //
  // MEASURED, NOT FEARED. Commenting out the three engine-named `@font-face`
  // rules in `scripts/build-wasm.mjs` — leaving the text in place as a comment
  // — left every test in this file and in `font-binary-identity.test.ts` green
  // while the emitted stylesheet dropped from six rules to three, which is
  // exactly the state that shipped the reported Thai overlap. The floor of six
  // above cannot catch it on its own: a commented-out rule still counts as a
  // declaration to a raw text scan. This is the direction that proves it does
  // not any more.
  it('does not count a commented-out @font-face rule as a declared family', () => {
    const emitted = "@font-face { font-family: 'Noto Sans'; src: url('./runtime/x.ttf') format('truetype'); font-display: swap; }"
    // The live direction, so the parse is shown to find a rule at all.
    expect(declaredFamilies(withoutComments(emitted))).toEqual(['Noto Sans'])
    // Both comment forms a generator can hide a rule in.
    expect(declaredFamilies(withoutComments(`// ${emitted}`))).toEqual([])
    expect(declaredFamilies(withoutComments(`/* ${emitted} */`))).toEqual([])
    // AND THE DEFECT ITSELF: without the strip, commented-out text counted.
    expect(declaredFamilies(`// ${emitted}`)).toEqual(['Noto Sans'])
  })

  // NO CHROME TOKEN IS EDITED — CHECKED, NOT ASSERTED IN PROSE.
  //
  // Story 8.4b's whole claim is that the engine's vocabulary became nameable in
  // the browser with NO chrome token touched. Nothing checked the second half:
  // `design-contract.test.ts` pins the token NAMES and the `@import` line, not
  // their values, so repointing `--font-sans` at `'Noto Sans'` — collapsing the
  // two vocabularies this story deliberately keeps separate — would have passed
  // every existing test. tokens.css is READ here and never edited.
  it('keeps each design-system family named by a --font-* token in tokens.css', () => {
    const fontTokens = fontTokenValues(fs.readFileSync(tokensPath, 'utf8'))
    // NON-VACUITY FLOOR. A reformat that stops the parse matching yields an
    // empty list over which every `not.toEqual([])` below would fail — which is
    // the point: this must redden rather than pass over an empty parse.
    expect(fontTokens.length, `read no --font-* tokens out of ${tokensPath}`).toBeGreaterThanOrEqual(3)
    for (const family of ['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai']) {
      expect(
        fontTokens.filter(([, value]) => value.includes(`'${family}'`)).map(([name]) => name),
        `no --font-* token in tokens.css names '${family}' any more. Story 8.4b declares the ENGINE's face names alongside `
        + 'the design system\'s; it does not replace them. A chrome token pointed at an engine face name collapses the two '
        + 'vocabularies, and every --type-* token resolves through these three.',
      ).not.toEqual([])
    }
    // AND THE OTHER DIRECTION, which is what the hazard actually looks like:
    // no chrome token may name an ENGINE face at all. Measured — the
    // containment check above alone is not enough, because repointing
    // `--font-sans` at 'Noto Sans' leaves `--font-page` still naming
    // 'IBM Plex Sans', so the three families stay present while a chrome token
    // has been collapsed onto the engine's vocabulary. The face names come from
    // `fonts.Shipped()`, not from a literal restated here.
    expect(engineFaces.length, 'the engine face names must have been read').toBe(3)
    for (const face of engineFaces) {
      expect(
        fontTokens.filter(([, value]) => value.includes(`'${face}'`)).map(([name]) => name),
        `a --font-* token in tokens.css names the ENGINE face '${face}'. The chrome and the engine keep separate `
        + 'vocabularies: Story 8.4b declares the engine\'s names for the CANVAS, and a chrome token pointed at one puts '
        + 'the design system\'s type on whichever bytes the engine happens to ship.',
      ).toEqual([])
    }

    // THE RED DIRECTION, through the same helper: a token repointed at the
    // engine's vocabulary no longer names the chrome family.
    expect(fontTokenValues("--font-sans: 'Noto Sans', system-ui, sans-serif;").filter(([, value]) => value.includes("'IBM Plex Sans'"))).toEqual([])
    expect(fontTokenValues("--font-sans: 'IBM Plex Sans', system-ui, sans-serif;").filter(([, value]) => value.includes("'IBM Plex Sans'")).length).toBe(1)
  })

  // GUARD 1, WIDENED BY STORY 8.4a. It used to say only that every family the
  // STYLESHEET asks for is declared by an `@font-face`. That was a tie between
  // two files and nothing more: a family set from TypeScript — which is
  // precisely what 8.4a introduces — escaped it entirely, so the old form
  // would have gone on passing while the canvas asked for a family nothing had
  // registered. It now ties, for a CARRIED face, the family the fragment
  // actually asks for to the ASSET THE ENGINE RESOLVED IT TO.
  //
  // THE TIE IS SCOPED TO THE CARRIED CASE, AND THAT IS STILL DELIBERATE —
  // BUT NOT FOR THE REASON IT USED TO BE. Until Story 8.4b this comment said
  // the universal form was FALSE, because for a shipped face the rule asked
  // for 'IBM Plex Sans' while the engine measured 'Noto Sans', two disjoint
  // vocabularies. THAT IS NO LONGER TRUE. 8.4b registers the same three
  // shipped files a SECOND time under the engine's own face names and points
  // the fragment rule at those names, so the shipped half now asks for
  // exactly the names the engine measures with — checked by the test below,
  // and tied to the engine's bytes by src/font-binary-identity.test.ts.
  //
  // WHAT KEEPS THE TIE SCOPED IS NOW THE OTHER RESIDUAL. The fragment stack
  // is a FIXED constant naming all three faces in one order, while a document
  // may declare a chain like ["Noto Sans Thai"] whose covering face is not
  // the stack's first; the three faces' cmaps genuinely overlap (339 / 529 /
  // 230 codepoints pairwise, measured), so the engine can measure a Latin run
  // with 'Noto Sans Thai' while the browser's Latin-first stack rasterizes it
  // with 'Noto Sans'. A shipped fragment carries no face identity on the wire
  // — only carried faces carry an `assetKey` — so nothing here can state the
  // per-fragment claim yet. Closing that needs shipped-face attribution on
  // the projection, the shape 8.4a built for carried faces; it is DW-35 cause
  // one's REMAINING half, and is not what this assertion claims. The carried
  // half is where the per-fragment claim is checkable today, and it is
  // checked here at full strength.
  it('asks only for families the browser has a face for, and ties every runtime one to the engine\'s own attribution', () => {
    // (a) THE STYLESHEET HALF, unchanged: every family the fragment rule names
    // is one the generator declares an @font-face for.
    expect(requested.filter((family) => !declared.includes(family))).toEqual([])

    // (b) THE RUNTIME HALF. A fragment the engine attributed to an asset the
    // document carries asks for that asset's own derived family — not a
    // stylesheet constant, not a chain name, not the asset's `font.family`.
    expect(paintedFragmentFamilies(carriedKey, new Set([carriedKey]))).toEqual([embeddedFaceFamily(carriedKey)])
    // A fragment the engine attributed to NOTHING is a shipped face and asks
    // for nothing of its own: it falls to App.css's declared stack, checked in
    // (a). This is the shipped-face path, unchanged by this story.
    expect(paintedFragmentFamilies(undefined, new Set([carriedKey]))).toEqual([''])
    // AND THE DEGRADE PATH. An inline declaration REPLACES the rule rather
    // than extending it, so asking for a family whose bytes never arrived
    // would take the fragment off the declared stack onto the browser's
    // default. The family is asked for only once the face is registered.
    expect(paintedFragmentFamilies(carriedKey, new Set())).toEqual([''])

    // (c) THE OTHER END OF THE TIE: the seam registers under the SAME
    // derivation of the SAME key, through the SAME module. Two derivations
    // that merely agree today are two derivations.
    const seam = fs.readFileSync(path.join(here, runtimeRegistrationSeam), 'utf8')
    const app = fs.readFileSync(path.join(here, 'App.tsx'), 'utf8')
    expect(seam).toMatch(/new FontFace\(embeddedFaceFamily\(assetKey\), bytes\)/)
    expect(seam).toContain('from \'./embedded-face-family\'')
    expect(app).toContain('from \'./embedded-face-family\'')

    // (d) AND IT CANNOT COLLIDE WITH A BUILD-TIME FAMILY. D-8.4.1's own
    // hazard: `document.fonts` is a global name-keyed registry, so a derived
    // family that happened to equal a declared one would silently substitute.
    for (const family of declared) expect(embeddedFaceFamily(carriedKey)).not.toBe(family)
  })

  // DW-35 TRIPWIRE, RE-RECORDED AT STORY 8.4b. It had two causes; cause two is
  // closed, and cause one is now HALF closed. Conflating the closed half with
  // the open one is how the open one disappears.
  //
  // CAUSE ONE, VOCABULARY LAYER (CLOSED BY STORY 8.4b). Until 8.4b the two
  // sides did not merely differ in stack ORDER — they used different NAMES for
  // the same three shipped files: the generator registered them under IBM Plex
  // family names while a chain's entries are the ENGINE's face names, so a
  // chain entry could not be used as a CSS family name AT ALL. The earlier
  // form of this comment called the fix a design-system decision above a
  // builder's authority, needing either a rename of the generated families
  // (rippling into tokens.css and design-contract.test.ts) or a face-name ->
  // CSS-family map. MEASURED FALSE, per D-8.4.14: it needed neither. 8.4b adds
  // a SECOND @font-face over each of the SAME three files under the engine's
  // own face names, so the engine's vocabulary is nameable in the browser with
  // no chrome token edited, no binary added and no mapping table built — a
  // mapping table being a second authority on which browser family is which
  // engine face, rejected by name. The browser family now IS the engine's
  // name, so there is nothing to map.
  //
  // CAUSE ONE, ATTRIBUTION LAYER (STILL OPEN). The fragment stack below is a
  // fixed constant, not the document's chain, and a shipped-face fragment
  // carries no face identity on the wire. So for a document whose chain is
  // `["Noto Sans Thai"]` the engine still measures with a face the browser's
  // fixed Latin-first order may not reach first, and the faces' coverage
  // overlaps enough for that to bite. Closing it needs per-fragment
  // shipped-face attribution on the projection — 8.4a's carried-face shape,
  // extended to shipped faces — which is a different story. DW-35 cause one
  // should be NARROWED to this residual, not closed; see deferred-work.md.
  //
  // CAUSE TWO (Story 8.4, CLOSED BY STORY 8.4a). The engine renders — and
  // measures — with a face the DOCUMENT ITSELF CARRIES, decoded out of its
  // `assets` map, and the browser had NOTHING for it: no `@font-face`, no
  // family name, no bytes at all, so the fragment stack fell straight through
  // to generic `sans-serif`. 8.4a carries each fragment's ASSET KEY through
  // the projection, fetches the bytes over the existing `asset` operation, and
  // registers a `FontFace` under a family derived from that key. The guard
  // above is the tie; the guards below are what keep the mechanism from
  // spreading.
  //
  // THE DESIGN DECISION 8.4a INHERITED WAS ALREADY MADE (D-8.4.1): a carried
  // face's CSS family name derives from its ASSET KEY, never from the asset's
  // `font.family`. AD-8 makes the asset key the resolver, and deriving from
  // `font.family` would let a document's "Inter" collide with a shipped
  // "Inter" in the browser's own font registry — AD-8's hazard, one layer down.
  //
  // THE OBSTACLE THAT WAS MEASURED AND IS NOW GONE, kept because the shape of
  // its removal is what a reader needs. It said the two sides did not merely
  // differ in stack ORDER but used different NAMES for the same shipped files,
  // so a chain entry could not be used as a CSS family name at all — and that
  // the fix therefore needed a face-name -> CSS-family mapping existing on
  // NEITHER side, or a rename of the generated families rippling into the
  // design tokens and their contract test. Story 8.4b did a THIRD thing: it
  // added a second `font-face` rule per file under the engine's own name,
  // leaving the IBM Plex rules and every token untouched. A chain entry is now
  // a usable CSS family name.
  //
  // THE ALIASING TRAP IS STILL LIVE, and is now the only reason the two halves
  // must never be collapsed: the generator's `'IBM Plex Mono'` is Noto Sans SC,
  // not a mono face. That defect belongs to Story 8.4c, which puts real IBM
  // Plex bytes behind the IBM Plex names; until then the pairing is pinned,
  // file by file, in src/font-binary-identity.test.ts.
  it('records that the fragment stack is a stylesheet constant with no document input', () => {
    // NON-VACUITY FIRST. `find(...) ?? ''` yields an empty string the moment
    // the rule is reformatted onto several lines, and `expect('').not.toMatch`
    // passes while proving nothing at all. Both halves are asserted to have
    // been FOUND before anything is asserted about them.
    const rule = css.split('\n').find((line) => line.startsWith('.canvas-text-fragment {'))
    expect(rule, 'the single-line .canvas-text-fragment rule must exist').toBeDefined()
    const declaration = /font-family:([^;]+);/.exec(rule as string)?.[1]
    expect(declaration, '.canvas-text-fragment must declare a font-family').toBeDefined()
    // Every family is a literal. No custom property, no interpolation, and no
    // way for a projected chain to reach this declaration. Story 8.4a's derived
    // family is set INLINE on the one fragment it belongs to and never here:
    // one stylesheet rule cannot vary per fragment, and a mixed-script element
    // needs it to.
    expect(declaration as string).not.toMatch(/var\(/)
    expect(requested.length).toBeGreaterThanOrEqual(3)
  })

  // THE SUCCESSOR OF THE DISJOINTNESS RECORD (Story 8.4b). What stood here
  // asserted, in both directions, that the engine's face names and the
  // browser's declared families do NOT intersect — the deliberate disjointness
  // that was DW-35 cause one. 8.4b reverses exactly that, so those two
  // assertions could not be edited: they had to go red and be replaced by
  // their opposite. THE CHROME HALF DID NOT GO WITH THEM. The IBM Plex
  // families are still declared and are still what every `--type-*` token in
  // tokens.css resolves through; nothing in this story renames, removes or
  // repoints them, and the arrayContaining floor below is what would notice.
  //
  // THE ENGINE'S NAMES ARE READ, NOT RESTATED. The form this replaces compared
  // a dynamically-read `declared` against a hardcoded three-element literal —
  // which would have kept passing, while having become false, if folio-go ever
  // shipped a different FontSet. Both halves of the claim below come from
  // `fonts.Shipped()` itself.
  it('declares the engine\'s own face names and asks the canvas for exactly them', () => {
    // NON-VACUITY BEFORE ANYTHING ELSE. A parse that yields nothing makes every
    // `filter(...).toEqual([])` below pass over an empty set, which is the
    // classic way this shape of guard goes quiet.
    expect(engineFaces.length, `read no face names out of Shipped() in ${enginePath}`).toBe(3)

    // THE CHROME HALF, UNWEAKENED. The design system's vocabulary must remain
    // declared; the canvas no longer asks for it, but every type token does.
    expect(declared).toEqual(expect.arrayContaining(['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai']))

    // THE BROWSER CAN NAME THE FACE THE ENGINE MEASURED WITH. Every face in
    // the shipped FontSet has an @font-face of its own. That those faces are
    // declared from the ENGINE'S OWN BYTES is the separate, stronger claim
    // made in src/font-binary-identity.test.ts.
    expect(engineFaces.filter((face) => !declared.includes(face))).toEqual([])

    // AND THE CANVAS ASKS FOR THEM. Containment, not equality: the stack ends
    // in the generic `sans-serif` keyword, which is guarded separately below.
    expect(engineFaces.filter((face) => !requested.includes(face))).toEqual([])

    // AND IN THE ENGINE'S OWN ORDER, which containment alone does not say.
    // The acceptance criterion and the I/O matrix both require the ORDER, and
    // for good reason: a CSS stack is a first-match-wins search per codepoint,
    // the three faces' cmaps overlap (339 / 529 / 230 codepoints pairwise,
    // measured) and all three cover `A` and `5`. Reordering the stack CJK-first
    // changes which face rasterizes every overlapping codepoint — a metric
    // change the containment assertion above waves straight through. The
    // expected order is `fonts.Shipped()`'s own source order, parsed above, so
    // this ties the browser's search order to the engine's declaration order
    // rather than to a literal restated here.
    expect(
      requested.slice(0, engineFaces.length),
      'the .canvas-text-fragment stack must name the engine\'s faces first and in the order fonts.Shipped() writes them',
    ).toEqual([...engineFaces])
  })

  // STORY 8.4a'S POSITIVE TWIN OF STORY 8.4'S DISCLOSURE OF ABSENCE.
  //
  // 8.4 recorded that NOTHING in this build registers a face at runtime, and
  // said in its own words that the assertion "will have to be deleted by 8.4a
  // rather than merely edited". It was: an absence assertion cannot survive
  // the thing whose absence it asserts. What survives is its SCANNING POWER,
  // re-pointed from "nowhere" to "exactly here". The detector is unchanged and
  // is still proved against itself below.
  //
  // WHY AN EXACT LIST RATHER THAN A COUNT. The hazard is not that registration
  // happens twice; it is that it happens somewhere with the WRONG LIFETIME.
  // `document.fonts` is a global, name-keyed registry, and the obvious place
  // to copy from — ImagePaint — is mounted once per component AND once per
  // repeated sheet, so a face registered there would be added N x M times
  // under one family and deleted by whichever instance unmounted first, while
  // another was still painting with it. Naming the seam is what makes that
  // mistake visible.
  it('registers a face at runtime in exactly one named seam and nowhere else', () => {
    // Non-vacuity: the generator really does declare the build-time faces, so
    // what follows is a statement about a corpus that has font registration in
    // it, not about an empty or unread one.
    expect(declared.length).toBeGreaterThanOrEqual(3)
    for (const declaration of declared) {
      expect(generator).toContain(`@font-face { font-family: '${declaration}'`)
    }
    // And every @font-face src in the GENERATOR is still a build-time asset,
    // never a document's bytes: the build-time path did not acquire a second
    // mechanism while the runtime one was being added.
    expect(registersAFaceAtRuntime(withoutComments(generator))).toBe(false)
    const sources = designerSources()
    expect(sources.length).toBeGreaterThan(10)
    expect(runtimeRegistrationSites(sources)).toEqual([runtimeRegistrationSeam])
  })

  // THE REPLACEMENT'S OWN RED-PROOF. An exact list is only a guard if both of
  // its failure directions actually fail, and neither can be shown by the real
  // corpus — which has, and must keep having, exactly one site.
  it('turns a second registration site red, and a seam that stops registering red too', () => {
    const seam = [runtimeRegistrationSeam, "const face = new FontFace(embeddedFaceFamily(assetKey), bytes)"] as const
    const benign = ['App.tsx', 'const style = { fontSize: 12 }'] as const
    expect(runtimeRegistrationSites([seam, benign])).toEqual([runtimeRegistrationSeam])
    // A SECOND SITE. This is the mutation the exact list exists to catch —
    // ImagePaint growing a font branch, say.
    const second = ['ImagePaint.tsx', 'document.fonts.add(face)'] as const
    expect(runtimeRegistrationSites([seam, second])).not.toEqual([runtimeRegistrationSeam])
    // AND THE SEAM ITSELF GOING QUIET, which would mean the feature had been
    // removed or renamed out from under this list.
    expect(runtimeRegistrationSites([benign])).toEqual([])
    // A MENTION IS NOT A REGISTRATION. This is the mutation that was actually
    // getting through: a module that only DESCRIBES the mechanism in a comment
    // — as the seam's own header does, and as this file does throughout — must
    // not be counted, or the seam could stop registering and stay on the list.
    expect(runtimeRegistrationSites([['a-module.ts', '// the seam calls new FontFace and document.fonts.add once per carried face']])).toEqual([])
  })

  // THE DETECTOR IS ITSELF ASSERTED, because a negative scan is only as strong
  // as the pattern behind it and a regex that matches nothing passes every
  // "no offender" assertion in this file. Each mechanism the comment above
  // names is shown to be caught, and ordinary source is shown not to be.
  it('detects each of the three runtime-registration mechanisms it claims to scan for', () => {
    for (const mechanism of [
      "const face = new FontFace('x', 'url(data:font/ttf;base64,AA)')",
      "document.fonts.add(face)",
      "style.textContent = `@font-face { font-family: 'x'; src: url(data:font/ttf;base64,AA); }`",
      "sheet.insertRule(\"@font-face { font-family: 'x'; src: url(blob:http://a/b) }\")",
    ]) {
      expect(registersAFaceAtRuntime(mechanism), mechanism).toBe(true)
    }
    for (const benign of [
      "const url = 'data:image/png;base64,AA'",
      "createObjectURL(new Blob([bytes], { type: image.mediaType }))",
      "// the generator writes an @font-face per shipped face at build time",
    ]) {
      expect(registersAFaceAtRuntime(benign), benign).toBe(false)
    }
  })

  // GUARD 2, WIDENED BY STORY 8.4a. It used to say only that no source names a
  // CHAIN ENTRY in a font-family position — a tripwire on one obvious route,
  // which an asset-key-derived family would have walked straight past, green,
  // by simply not being spelled `chain.entries[0]`. It now says the opposite
  // way round, which is the strictly stronger claim: a font-family position
  // may name ONE thing, the family derived from the asset key the engine
  // attributed, and every other spelling — a chain entry, a chain name, an
  // asset's `font.family`, the projected `fontFamily` field, a literal — is a
  // violation.
  //
  // AND IT NO LONGER TAXES PROSE. The old form scanned raw file text, so a
  // COMMENT explaining what not to write reddened it; the two directions are
  // both proved in the test below this one.
  it('permits only an asset-key-derived family in a font-family position, in every designer source', () => {
    const sources = designerSources()
    expect(sources.length).toBeGreaterThan(10)
    const positions = sources.flatMap(([name, source]) => fontFamilyDeclarations(source).map((value) => `${name}: ${value}`))
    // NON-VACUITY AND THE WHOLE CLAIM IN ONE LINE: there is exactly one CSS
    // font-family position in the designer's TypeScript, it is the canvas
    // fragment's, and it names the derived family. `App.css`'s own literal
    // stack is a stylesheet constant and is guarded separately, above and
    // below.
    expect(positions).toEqual(['App.tsx: embeddedFaceFamily(fragment.assetKey)'])
  })

  it('turns a document-supplied family in a font-family position red, and leaves the prose describing one alone', () => {
    // Every route a document's own vocabulary could reach CSS by, including
    // the one the old form of this guard named.
    expect(unapprovedFontFamilyDeclarations('style={{ fontFamily: chain.entries[0] }}')).not.toEqual([])
    expect(unapprovedFontFamilyDeclarations('style={{ fontFamily: entry.family }}')).not.toEqual([])
    expect(unapprovedFontFamilyDeclarations('style={{ fontFamily: component.fontFamily }}')).not.toEqual([])
    expect(unapprovedFontFamilyDeclarations('node.style.fontFamily = ""; const rule = `font-family: ${chain.name}`')).not.toEqual([])
    expect(unapprovedFontFamilyDeclarations('const css = ".x { font-family: \'IBM Plex Sans\' }"')).not.toEqual([])
    // THE APPROVED ONE, and only in its derived form.
    expect(unapprovedFontFamilyDeclarations('style={{ fontFamily: embeddedFaceFamily(fragment.assetKey) }}')).toEqual([])
    // THE NEGATIVE CASES. A scan that only ever reddens has not been shown to
    // discriminate. Prose describing the forbidden route is not the route, and
    // the projection's `fontFamily` FIELD — a document's chain name — is read
    // all over this codebase without ever reaching a CSS declaration.
    expect(unapprovedFontFamilyDeclarations('// never write fontFamily: chain.entries[0] here')).toEqual([])
    expect(unapprovedFontFamilyDeclarations('/* fontFamily: entry.family would collide with a shipped face */')).toEqual([])
    expect(unapprovedFontFamilyDeclarations('const values = components.map((c) => committedValue(c, \'fontFamily\'))')).toEqual([])
    expect(unapprovedFontFamilyDeclarations('type Component = Readonly<{ fontFamily?: string; fontSize?: number }>')).toEqual([])
    expect(unapprovedFontFamilyDeclarations('void commit({ field: \'fontFamily\', operation: \'set\', value: name })')).toEqual([])
  })

  // The generic keyword is a last resort and must stay last. If it moved
  // ahead of a declared family the browser would never reach the real
  // face, reproducing the same defect with the stack looking correct.
  it('keeps the generic fallback last', () => {
    const rule = css.split('\n').find((line) => line.startsWith('.canvas-text-fragment {')) ?? ''
    const declaration = /font-family:([^;]+);/.exec(rule)?.[1] ?? ''
    const entries = declaration.split(',').map((entry) => entry.trim())
    expect(entries[entries.length - 1]).toBe('sans-serif')
    expect(entries.slice(0, -1).every((entry) => entry.startsWith("'"))).toBe(true)
  })
})
