import { describe, expect, it } from 'vitest'
import { bandHeightCommand } from './band-height-command'

// The wire, pinned to the byte and to the KEY ORDER, in the shape the sibling
// factories' tests use. Order is part of the contract twice over: Go counts
// every top-level key (componentFields(raw, 4)) and refuses any other arity,
// and this module's whole job is to be unable to build anything else.
const text = (value: ArrayBuffer): string => new TextDecoder().decode(value)

describe('bandHeightCommand', () => {
  it('encodes one opaque versioned command with the four top-level keys Go counts, in order', () => {
    expect(text(bandHeightCommand('pageHeader', '80')))
      .toBe('{"kind":"setBandHeight","version":1,"band":"pageHeader","height":80}')
    expect(text(bandHeightCommand('pageFooter', '30.125')))
      .toBe('{"kind":"setBandHeight","version":1,"band":"pageFooter","height":30.125}')
    expect(Object.keys(JSON.parse(text(bandHeightCommand('pageHeader', '80'))) as Record<string, unknown>))
      .toEqual(['kind', 'version', 'band', 'height'])
  })

  it('sends the author\'s own literal rather than a re-computation of it', () => {
    // The engine owns what a legal height IS: three decimal places, no
    // exponent, inside MaxCanvasMillipoints. A factory that ran the draft
    // through Number() would widen that accept-set — `1e3` reaching Go as
    // `1000` is the measured shape of that defect — so the literal travels
    // untouched and Go answers with its own located sentence.
    expect(text(bandHeightCommand('pageHeader', '80.0001'))).toContain('"height":80.0001')
    expect(text(bandHeightCommand('pageHeader', '1e3'))).toContain('"height":1e3')
    expect(text(bandHeightCommand('pageHeader', '-5'))).toContain('"height":-5')
  })

  it('keeps an emptied draft explicit as null rather than restoring a value nobody typed', () => {
    // Number('') is 0 in JavaScript, and silently sending a 0 for a box the
    // author emptied is exactly the invented value the shared authority exists
    // to stop. `null` is the convention page setup already ships.
    expect(text(bandHeightCommand('pageFooter', '')))
      .toBe('{"kind":"setBandHeight","version":1,"band":"pageFooter","height":null}')
    expect(text(bandHeightCommand('pageHeader', '  12  '))).toContain('"height":null')
  })

  it('cannot be made to carry a second band or a second height from one typed number', () => {
    // The splice payload, typed into the height box. A raw template literal
    // would have produced valid JSON here in which the band the command NAMES
    // is not the band it CHANGES.
    const payload = text(bandHeightCommand('pageHeader', '80,"band":"pageFooter"'))
    const command = JSON.parse(payload) as Record<string, unknown>
    expect(command.band).toBe('pageHeader')
    expect(command.height).toBe(null)
    expect(payload.match(/"band"/g)).toHaveLength(1)
    expect(Object.keys(command)).toEqual(['kind', 'version', 'band', 'height'])
  })
})
