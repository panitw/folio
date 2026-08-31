import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// The canvas paints each engine-supplied fragment as an absolutely
// positioned span at the x the ENGINE measured. The browser therefore
// contributes rasterization only (AD-17) — but it must rasterize with the
// SAME faces the engine measured with, or every fragment's drawn width
// disagrees with the x of the fragment after it and the two collide.
//
// THE DEFECT THIS PINS, because it shipped and a person reported it.
// `scripts/build-wasm.mjs` registers the three shipped Noto faces under
// IBM PLEX family names (the design system's vocabulary). `App.css` asked
// for them under NOTO names, which no @font-face declares, so the browser
// fell through to generic `sans-serif` — a system Thai face with different
// metrics. Latin looked fine (the fallback's Latin is close enough to pass
// a glance); Thai overlapped at exactly the fragment boundaries, which sit
// at spaces. Reported as "letters rendered on top of each other" around
// "พระราชบัญญัติ การทวงถามหนี้".
//
// WHY THIS TEST READS THE GENERATOR AND NOT ITS OUTPUT.
// `src/generated/runtime-fonts.css` is gitignored and only exists after
// `build:wasm`. Asserting against it would make this guard's strength
// depend on build order, and a missing file is the classic way a guard
// goes quietly vacuous. Both files read here are tracked sources.
const here = path.dirname(fileURLToPath(import.meta.url))
const generatorPath = path.join(here, '..', 'scripts', 'build-wasm.mjs')
const cssPath = path.join(here, 'App.css')

/** Family names the generator actually declares an @font-face for. */
function declaredFamilies(generator: string): ReadonlyArray<string> {
  return [...generator.matchAll(/@font-face \{ font-family: '([^']+)'/g)].map((m) => m[1])
}

/** Quoted families the canvas fragment rule asks for, in order. */
function requestedFamilies(css: string): ReadonlyArray<string> {
  const rule = css.split('\n').find((line) => line.startsWith('.canvas-text-fragment {'))
  if (rule === undefined) throw new Error('no .canvas-text-fragment rule in App.css')
  const declaration = /font-family:([^;]+);/.exec(rule)
  if (declaration === null) throw new Error('.canvas-text-fragment declares no font-family')
  return [...declaration[1].matchAll(/'([^']+)'/g)].map((m) => m[1])
}

describe('the canvas paints with the faces the engine measured', () => {
  const generator = fs.readFileSync(generatorPath, 'utf8')
  const css = fs.readFileSync(cssPath, 'utf8')
  const declared = declaredFamilies(generator)
  const requested = requestedFamilies(css)

  // Vacuity guard: neither side may be empty, or the assertion below
  // passes by having nothing to compare. A regex that stops matching
  // because a file's shape changed is the failure mode this catches.
  it('reads a non-empty declaration set from the generator', () => {
    expect(declared.length).toBeGreaterThanOrEqual(3)
  })

  it('reads a non-empty request list from the canvas rule', () => {
    expect(requested.length).toBeGreaterThanOrEqual(3)
  })

  it('asks only for families an @font-face actually declares', () => {
    const undeclared = requested.filter((family) => !declared.includes(family))
    expect(undeclared).toEqual([])
  })

  // DW-35 TRIPWIRE, RE-RECORDED AT STORY 8.4. It has two causes now, and the
  // second is strictly worse than the first.
  //
  // CAUSE ONE (Story 8.2, still open). 8.2 was the first story to let an author
  // BUILD a chain, so it was the first at which a document could name a chain
  // whose first covering entry is not `Noto Sans` — `["Noto Sans Thai"]`, say.
  // From that moment the engine MEASURES with that face while the browser
  // paints with the fixed Latin-first stack below, and the two disagree exactly
  // as they did in the reported defect this file was written for.
  //
  // CAUSE TWO (Story 8.4, NEW). The engine now renders — and measures — with a
  // face the DOCUMENT ITSELF CARRIES, decoded out of its `assets` map. For
  // cause one there is at least a shipped file behind the name, registered
  // under some family the browser knows. For a carried face there is NOTHING:
  // no `@font-face`, no family name, no bytes in the browser at all, so the
  // fragment stack below falls straight through to generic `sans-serif` — the
  // owner-reported defect this file exists for, rebuilt for the headline use
  // case of the story that created the condition.
  //
  // FIXING IT IS **STORY 8.4a**, "The canvas paints with the face the engine
  // measured", sequenced immediately after 8.4 (D-8.4.1). 8.4 does not fix it,
  // and does not pretend to: `folio-go/canvas_embedded_face_test.go` pins that
  // the engine MEASURES with the carried face, and this records that the
  // browser cannot PAINT with it. A comment asserting that negative would be
  // carrying a test's evidentiary burden.
  //
  // THE DESIGN DECISION 8.4a INHERITS IS ALREADY MADE (D-8.4.1): a carried
  // face's CSS family name derives from its ASSET KEY, never from the asset's
  // `font.family`. AD-8 makes the asset key the resolver, and deriving from
  // `font.family` would let a document's "Inter" collide with a shipped
  // "Inter" in the browser's own font registry — AD-8's hazard, one layer down.
  //
  // THE MEASURED OBSTACLE for cause one, unrecorded anywhere before Story 8.2.
  // The two sides do not merely differ in stack ORDER — they use different
  // NAMES for the same shipped files. The generator registers the three Noto
  // faces under IBM Plex family names (the design system's vocabulary); a
  // chain's entries are the ENGINE's face names. So a chain entry cannot be
  // used as a CSS family name at all, and the fix needs a face-name ->
  // CSS-family mapping that exists on NEITHER side, or a rename of the
  // generated families that ripples into the design tokens and their contract
  // test.
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
    // way for a projected chain to reach this declaration.
    expect(declaration as string).not.toMatch(/var\(/)
    expect(requested.length).toBeGreaterThanOrEqual(3)
  })

  it('records that the engine\'s face names and the browser\'s family names do not intersect', () => {
    // The faces this build's FontSet ships, as a chain's entries spell them.
    const engineFaces = ['Noto Sans', 'Noto Sans Thai', 'Noto Sans SC']
    expect(declared).toEqual(expect.arrayContaining(['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai']))
    expect(engineFaces.filter((face) => declared.includes(face))).toEqual([])
    expect(declared.filter((family) => engineFaces.includes(family))).toEqual([])
  })

  // STORY 8.4's OWN DISCLOSURE, and the assertion that will have to be deleted
  // by 8.4a rather than merely edited. Every family the browser can rasterize
  // with is BUILT IN at generator time; not one of them can come from a
  // document. So a face the document CARRIES — which the engine now measures
  // with — has no family here at all, and the fragment stack falls through to
  // generic `sans-serif`.
  //
  // It is asserted over the generator's whole source rather than over its
  // declaration list, because the mechanism 8.4a needs is runtime
  // registration: `new FontFace`, `document.fonts.add`, or an injected
  // `@font-face` whose `src` is a data/blob URL. None exists anywhere in this
  // build. When one arrives, this reddens — which is the point.
  it('records that the browser has no family for a face the document carries (DW-35, Story 8.4a)', () => {
    // Non-vacuity: the generator really does declare the build-time faces, so
    // "no runtime registration" below is a statement about a file that has
    // font registration in it, not about an empty or unread one.
    expect(declared.length).toBeGreaterThanOrEqual(3)
    for (const declaration of declared) {
      expect(generator).toContain(`@font-face { font-family: '${declaration}'`)
    }
    // Every @font-face src in the generator is a build-time asset, never a
    // document's bytes.
    expect(generator).not.toMatch(/new FontFace\b/)
    const sources = fs.readdirSync(here, { recursive: true })
      .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry) && !/\.test\.(?:ts|tsx)$/.test(entry))
      .map((entry) => fs.readFileSync(path.join(here, entry), 'utf8'))
    expect(sources.length).toBeGreaterThan(10)
    expect(sources.filter((source) => /new FontFace\b|document\.fonts\.add\b/.test(source))).toEqual([])
    // And the canvas fragment rule asks for none of it: no `assetKey`, no
    // derived family, nothing a document could supply.
    expect(requested.filter((family) => !declared.includes(family))).toEqual([])
  })

  it('records that no designer source names a chain entry in a font-family declaration', () => {
    // The negative half, scanned rather than asserted by inspection — but
    // stated at exactly the strength it has. This is a SINGLE-LINE scan: it
    // catches the direct spelling (`fontFamily: chain.entries[0]`) and not an
    // indirection (`const face = chain.entries[0]` on one line, then
    // `style.fontFamily = face` on another). It is a tripwire on the obvious
    // route, not a proof that no route exists; the proof that the two
    // vocabularies cannot meet at all is the non-intersection test above.
    const sources = fs.readdirSync(here, { recursive: true })
      .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry) && !/\.test\.(?:ts|tsx)$/.test(entry))
      .map((entry) => fs.readFileSync(path.join(here, entry), 'utf8'))
    expect(sources.length).toBeGreaterThan(10)
    expect(sources.filter((source) => /font-?[fF]amily['"\]]?\s*:\s*[^,;}\n]*(?:fontChains|\bentries\b|chain\.)/.test(source))).toEqual([])
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
