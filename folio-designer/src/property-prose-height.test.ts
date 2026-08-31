import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

// The font-family combobox and the CONTENT field share `property-value-prose`
// because both want body type instead of the mono the other property values
// use. They do NOT share a height contract: CONTENT is a multi-line textarea
// with a four-row floor and a resize grip, and the combobox is a single-line
// input. When the floor lived on the shared class the combobox rendered at
// roughly twice its intended height.
//
// These tests tie the class to the ELEMENT that has the multi-line contract.
// They read the two tracked sources rather than any built artefact, so they
// fail on the edit that would reintroduce the bug rather than on a build.

const here = path.dirname(fileURLToPath(import.meta.url))
const css = fs.readFileSync(path.join(here, 'App.css'), 'utf8')
const tsx = fs.readFileSync(path.join(here, 'App.tsx'), 'utf8')

describe('property-value-prose height is scoped to the textarea', () => {
  it('does not put a height floor or a resize grip on the shared class', () => {
    const shared = css.match(/^\.property-value-prose \{[^}]*\}/m)
    expect(shared, 'the shared .property-value-prose rule must exist').not.toBeNull()
    expect(shared![0]).not.toMatch(/min-height/)
    expect(shared![0]).not.toMatch(/resize/)
  })

  it('scopes the height floor and resize grip to textarea', () => {
    const scoped = css.match(/^textarea\.property-value-prose \{[^}]*\}/m)
    expect(scoped, 'the textarea-scoped rule must exist').not.toBeNull()
    expect(scoped![0]).toMatch(/min-height:\s*72px/)
    expect(scoped![0]).toMatch(/resize:\s*vertical/)
  })

  // Non-vacuity: the two tests above would both pass if nothing used the class
  // at all. These pin that the pairing they are about is still real.
  it('is still shared by exactly one textarea and one input', () => {
    const textareas = tsx.match(/<textarea[^>]*property-value-prose/g) ?? []
    const inputs = tsx.match(/<input[^>]*property-value-prose/g) ?? []
    expect(textareas).toHaveLength(1)
    expect(inputs).toHaveLength(1)
  })

  it('keeps body type on the shared class, which is why it is shared at all', () => {
    expect(css).toMatch(/\.property-value-prose \{[^}]*font:\s*var\(--type-body-em\)/)
  })
})
