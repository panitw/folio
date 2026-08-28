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

  it('keeps the explicit format null operation distinct from clear and set', () => {
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'background', operation: 'null' }))).toContain('"background":{"op":"null"}')
    expect(decode(updateComponentPropertiesCommand(['e1'], { field: 'borderEdges', operation: 'set', value: ['top', 'bottom'] }))).toContain('"value":["top","bottom"]')
  })
})
