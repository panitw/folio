import { describe, expect, it } from 'vitest'
import { bindComponentScalarCommand, dropComponentCommand, moveComponentCommand, resizeComponentCommand } from './component-command'

const text = (value: ArrayBuffer) => new TextDecoder().decode(value)

describe('opaque component commands', () => {
  it('converts projection millipoints to exact point literals once for move and resize', () => {
    expect(text(moveComponentCommand('e9', 1001, 2002, false))).toBe('{"kind":"moveComponent","version":1,"id":"e9","x":1.001,"y":2.002,"snap":false}')
    expect(text(resizeComponentCommand('e9', 73003, 25004, true))).toBe('{"kind":"resizeComponent","version":1,"id":"e9","width":73.003,"height":25.004,"snap":true}')
  })

  it('sends a global document point to the Go-owned drop hit test', () => {
    expect(text(dropComponentCommand('text', 36, 56, true))).toBe('{"kind":"dropComponent","version":1,"type":"text","x":36,"y":56,"snap":true}')
  })

  it('encodes decoded picker segments with complete JSON escaping', () => {
    expect(text(bindComponentScalarCommand('e1', ['a.b', 'line\nbreak', '\u0000']))).toBe('{"kind":"bindComponentScalar","version":1,"id":"e1","segments":["a.b","line\\nbreak","\\u0000"]}')
  })
})
