import { describe, expect, it } from 'vitest'
import { bindComponentScalarCommand, dropComponentCommand, moveComponentCommand, moveComponentsCommand, resizeComponentCommand } from './component-command'

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

it('encodes a captured group movement as one relative, revision-fenced command', () => {
  expect(new TextDecoder().decode(moveComponentsCommand(['e1', 'e2'], 'e2', -1125, 2227, true, 17))).toBe('{"kind":"moveComponents","version":1,"ids":["e1","e2"],"referenceId":"e2","dx":-1.125,"dy":2.227,"snap":true,"expectedRevision":17}')
})

// Pointer-only policy remains optional for continuous-coordinate callers.
it('encodes the current-window constraint only when requested', () => {
  const decode = (value: ArrayBuffer) => JSON.parse(new TextDecoder().decode(value))
  expect(decode(moveComponentsCommand(['e1'], 'e1', 0, 9000, false, 1))).not.toHaveProperty('constrainToWindow')
  expect(decode(moveComponentsCommand(['e1'], 'e1', 0, 9000, false, 1, true))).toHaveProperty('constrainToWindow', true)
})
