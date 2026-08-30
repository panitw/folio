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

  it('keeps the explicit format null operation distinct from clear and set', () => {
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'background', operation: 'null' }))).toContain('"background":{"op":"null"}')
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'borderEdges', operation: 'set', value: ['top', 'bottom'] }))).toContain('"value":["top","bottom"]')
  })
})
