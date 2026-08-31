import { describe, expect, it } from 'vitest'
import { addFontChainCommand, addFontChainEntryCommand, deleteFontChainCommand, moveFontChainEntryCommand, removeFontChainEntryCommand, renameFontChainCommand } from './font-chain-command'

const text = (payload: ArrayBuffer): string => new TextDecoder().decode(payload)
const parsed = (payload: ArrayBuffer): Record<string, unknown> => JSON.parse(text(payload)) as Record<string, unknown>

// The two values the engine will never see if the encoder is wrong. `a"b\c`
// exercises the two characters JSON's own syntax is made of; U+0001 is a C0
// control that the encoder this story replaced did NOT escape — it emitted the
// raw byte inside a JSON string, which is malformed, so Go answered with a
// generic parse failure instead of the located refusal naming the field. An
// author pastes such a byte far more easily than they type one.
const awkward = 'a"b\\c'
const controlled = 'a\u0001b'

describe('the six font chain commands are exactly the payloads the engine reads', () => {
  it('encodes every kind with its exact field arity', () => {
    // componentFields(raw, N) counts EVERY top-level key, `kind` and `version`
    // included, and an extra or missing one is a refusal rather than an
    // ignored field. These are the six counts folio-go's handlers require.
    expect(text(addFontChainCommand('heading', ['Noto Sans', 'Noto Sans Thai']))).toBe('{"kind":"addFontChain","version":1,"name":"heading","entries":["Noto Sans","Noto Sans Thai"]}')
    expect(text(renameFontChainCommand('body', 'text'))).toBe('{"kind":"renameFontChain","version":1,"name":"body","to":"text"}')
    expect(text(deleteFontChainCommand('body'))).toBe('{"kind":"deleteFontChain","version":1,"name":"body"}')
    expect(text(addFontChainEntryCommand('body', 1, 'Noto Sans SC'))).toBe('{"kind":"addFontChainEntry","version":1,"name":"body","index":1,"face":"Noto Sans SC"}')
    expect(text(moveFontChainEntryCommand('body', 2, 0))).toBe('{"kind":"moveFontChainEntry","version":1,"name":"body","from":2,"to":0}')
    expect(text(removeFontChainEntryCommand('body', 1))).toBe('{"kind":"removeFontChainEntry","version":1,"name":"body","index":1}')

    const arity: ReadonlyArray<readonly [ArrayBuffer, number, readonly string[]]> = [
      [addFontChainCommand('heading', ['Noto Sans']), 4, ['kind', 'version', 'name', 'entries']],
      [renameFontChainCommand('body', 'text'), 4, ['kind', 'version', 'name', 'to']],
      [deleteFontChainCommand('body'), 3, ['kind', 'version', 'name']],
      [addFontChainEntryCommand('body', 0, 'Noto Sans'), 5, ['kind', 'version', 'name', 'index', 'face']],
      [moveFontChainEntryCommand('body', 0, 1), 5, ['kind', 'version', 'name', 'from', 'to']],
      [removeFontChainEntryCommand('body', 0), 4, ['kind', 'version', 'name', 'index']],
    ]
    for (const [payload, fields, keys] of arity) {
      const object = parsed(payload)
      expect(Object.keys(object)).toEqual([...keys])
      expect(Object.keys(object)).toHaveLength(fields)
      expect(object.version).toBe(1)
    }
  })

  it('produces valid JSON for a value carrying a quote, a backslash and a C0 control, per builder', () => {
    // Every builder, every author-supplied position: the name, the rename
    // destination, a face in the create list, and a face added later.
    for (const value of [awkward, controlled, `${awkward}${controlled}`]) {
      expect(parsed(addFontChainCommand(value, [value]))).toEqual({ kind: 'addFontChain', version: 1, name: value, entries: [value] })
      expect(parsed(renameFontChainCommand(value, `${value}!`))).toEqual({ kind: 'renameFontChain', version: 1, name: value, to: `${value}!` })
      expect(parsed(deleteFontChainCommand(value))).toEqual({ kind: 'deleteFontChain', version: 1, name: value })
      expect(parsed(addFontChainEntryCommand(value, 3, value))).toEqual({ kind: 'addFontChainEntry', version: 1, name: value, index: 3, face: value })
      expect(parsed(moveFontChainEntryCommand(value, 1, 0))).toEqual({ kind: 'moveFontChainEntry', version: 1, name: value, from: 1, to: 0 })
      expect(parsed(removeFontChainEntryCommand(value, 2))).toEqual({ kind: 'removeFontChainEntry', version: 1, name: value, index: 2 })
    }
    // The C0 control travels as its \u escape, not as a raw byte.
    expect(text(deleteFontChainCommand(controlled))).toContain('\\u0001')
    expect(text(deleteFontChainCommand(controlled))).not.toContain('\u0001')
  })

  it('measures the departed encoder rather than asserting the new one is better', () => {
    // The escape table component-property-command.ts's quote() carried before
    // this story, reproduced here so the fix is proved against what it
    // replaced. It handles `\ " \n \r \t` and NOTHING else, while JSON
    // requires all of U+0000-U+001F.
    const departed = (value: string) => `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"').replaceAll('\n', '\\n').replaceAll('\r', '\\r').replaceAll('\t', '\\t')}"`
    // It was right about the five it knew, which is the population that must
    // not be narrowed by the fix — JSON.stringify still escapes all of them.
    for (const value of [awkward, 'a\nb', 'a\rb', 'a\tb', 'a\\b', 'a"b']) {
      expect(departed(value)).toBe(JSON.stringify(value))
      expect(JSON.parse(departed(value))).toBe(value)
    }
    // And wrong about every other control character, which is the defect.
    expect(() => JSON.parse(`{"name":${departed(controlled)}}`)).toThrow()
    expect(JSON.parse(`{"name":${JSON.stringify(controlled)}}`)).toEqual({ name: controlled })
    // A lone surrogate is the second value the departed table passed through
    // raw. JSON.stringify escapes it into a well-formed document instead.
    expect(text(deleteFontChainCommand('a\uD800b'))).toContain('\\ud800')
    expect(() => JSON.parse(text(deleteFontChainCommand('a\uD800b')))).not.toThrow()
  })
})
