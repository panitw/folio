import { describe, expect, it } from 'vitest'
import { updateComponentPropertiesCommand } from './component-property-command'

const decode = (value: ArrayBuffer): string => new TextDecoder().decode(value)

describe('updateComponentPropertiesCommand', () => {
  it('encodes one opaque versioned command without a document model', () => {
    expect(decode(updateComponentPropertiesCommand(['e1', 'e2'], { field: 'x', operation: 'set', value: '12.125' }))).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1","e2"],"changes":{"x":{"op":"set","value":12.125}}}')
  })

  it('keeps clear distinct from a literal empty text value', () => {
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'visibleIf', operation: 'clear' }))).toContain('"op":"clear"')
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'value', operation: 'set', value: '' }))).toContain('"value":""')
  })

  // Story 7.4. Two encodings that are easy to get wrong in opposite ways.
  it('sends a ratio unquoted and a multi-line clause with its breaks escaped', () => {
    // lineSpacing is a RAW, UNQUOTED number carrying the author's own ratio.
    // Go multiplies by 1000 itself, so 1.5 becomes 1500 thousandths in the
    // document; sending 1500 would be refused as 1 500 000, outside the
    // load-time range, and sending "1.5" quoted is refused as a non-number.
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'lineSpacing', operation: 'set', value: '1.5' }))).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"lineSpacing":{"op":"set","value":1.5}}}')
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'lineSpacing', operation: 'clear' }))).toContain('"lineSpacing":{"op":"clear"}')
    // A clause's paragraph breaks are QUOTED text and survive as \n escapes;
    // a CRLF pair travels as it was typed, and the engine folds it into ONE
    // mandatory break rather than two.
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'value', operation: 'set', value: 'One.\nTwo.\r\nThree.' }))).toBe('{"kind":"updateComponentProperties","version":1,"ids":["e1"],"changes":{"value":{"op":"set","value":"One.\\nTwo.\\r\\nThree."}}}')
  })

  // Story 8.2. quote() escaped `\ " \n \r \t` and NOTHING else, while JSON
  // requires every code point in U+0000-U+001F. A value carrying any other C0
  // control — U+0001 from the prose field's paste path, most plausibly — put a
  // RAW control byte inside a JSON string, so the command was malformed before
  // Go could read the field and the engine answered with a generic parse
  // failure instead of the located refusal naming it. Routing quote() through
  // JSON.stringify WIDENS what is escaped and must not narrow it, so the five
  // it already handled are re-asserted here beside the ones it missed.
  it('escapes the whole of U+0000-U+001F, and still escapes the five it always did', () => {
    const payload = (value: string) => decode(updateComponentPropertiesCommand(['e1'], { field: 'value', operation: 'set', value }))
    expect(() => JSON.parse(payload('a\u0001b'))).not.toThrow()
    expect(JSON.parse(payload('a\u0001b')).changes.value.value).toBe('a\u0001b')
    expect(payload('a\u0001b')).toContain('\\u0001')
    expect(payload('\u0000\u001f')).toContain('\\u0000\\u001f')
    // The departed population: the five characters quote() already handled,
    // and a value made only of JSON's own syntax characters.
    expect(payload('a\\b"c\nd\re\tf')).toContain('"value":"a\\\\b\\"c\\nd\\re\\tf"')
    // A lone surrogate is well-formed JSON only when escaped; it used to
    // travel raw and be replaced by U+FFFD at the encoder.
    expect(() => JSON.parse(payload('a\uD800b'))).not.toThrow()
    // A field name and an operation take the same encoder, unchanged.
    expect(payload('x')).toContain('"changes":{"value":{"op":"set"')
  })

  it('keeps the explicit format null operation distinct from clear and set', () => {
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'background', operation: 'null' }))).toContain('"background":{"op":"null"}')
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'borderEdges', operation: 'set', value: ['top', 'bottom'] }))).toContain('"value":["top","bottom"]')
  })
})
