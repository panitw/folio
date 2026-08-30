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
