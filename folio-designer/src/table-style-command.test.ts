import { describe, expect, it } from 'vitest'
import { tableAltRowBackgroundCommand, tableHeaderHeightCommand, tableHeaderStyleCommand } from './table-style-command'

// THIS FILE IS THE SINGLE AUTHORITY ON STORY 12.3's WIRE BYTES, and the story
// spec deliberately does not restate the arity: the spec owns the op grammar
// and the shared `findComponent` gate, and pinning the arity in prose as well
// would be a second spelling of one rule — the defect Story 12.2 spent its
// whole life on. If the two ever disagreed, this file would be right.
//
// Order is part of the contract twice over: Go counts every top-level key
// (componentFields) and refuses any other count, and the factory's whole job is
// to be unable to build anything else.
const text = (value: ArrayBuffer): string => new TextDecoder().decode(value)
const keys = (value: ArrayBuffer): string[] => Object.keys(JSON.parse(text(value)) as Record<string, unknown>)

describe('tableHeaderHeightCommand', () => {
  it('encodes one opaque versioned command with the four top-level keys Go counts, in order', () => {
    expect(text(tableHeaderHeightCommand('e7', '18')))
      .toBe('{"kind":"setTableHeaderHeight","version":1,"id":"e7","height":18}')
    expect(text(tableHeaderHeightCommand('e7', '12.5')))
      .toBe('{"kind":"setTableHeaderHeight","version":1,"id":"e7","height":12.5}')
    expect(keys(tableHeaderHeightCommand('e7', '18'))).toEqual(['kind', 'version', 'id', 'height'])
  })

  it('has no clear to offer, and sends the author\'s own literal rather than a re-computation of it', () => {
    // There is no `op` on this kind at all — `headerHeight` is required by the
    // format, so a cleared one is a document that cannot be reopened. The
    // factory cannot build the command that would ask for it.
    expect(text(tableHeaderHeightCommand('e7', '18'))).not.toContain('"op"')
    expect(text(tableHeaderHeightCommand('e7', '1e3'))).toContain('"height":1e3')
    expect(text(tableHeaderHeightCommand('e7', '18.0001'))).toContain('"height":18.0001')
    // Number('') is 0 in JavaScript. An emptied box must not silently restore a
    // height nobody typed.
    expect(text(tableHeaderHeightCommand('e7', ''))).toBe('{"kind":"setTableHeaderHeight","version":1,"id":"e7","height":null}')
  })
})

describe('tableAltRowBackgroundCommand', () => {
  it('encodes set with a value and clear without one, in order', () => {
    expect(text(tableAltRowBackgroundCommand('e7', 'set', '#DDEEFF')))
      .toBe('{"kind":"setTableAltRowBackground","version":1,"id":"e7","op":"set","value":"#DDEEFF"}')
    expect(text(tableAltRowBackgroundCommand('e7', 'clear')))
      .toBe('{"kind":"setTableAltRowBackground","version":1,"id":"e7","op":"clear"}')
    expect(keys(tableAltRowBackgroundCommand('e7', 'set', '#DDEEFF'))).toEqual(['kind', 'version', 'id', 'op', 'value'])
    expect(keys(tableAltRowBackgroundCommand('e7', 'clear'))).toEqual(['kind', 'version', 'id', 'op'])
  })

  it('never spells `null`, and passes a malformed colour through for the engine to refuse', () => {
    // Clearing is the ZERO presence, which removes the key. `"altRowBackground":
    // null` would still be the key in the file — different bytes, an undo
    // entry, and a raised format version — and the serializer has no null
    // branch for it, so the loader would refuse the document it just wrote.
    expect(text(tableAltRowBackgroundCommand('e7', 'clear'))).not.toContain('null')
    // The panel invents no second validation: Go's parseHexColor is the gate
    // and its located sentence is what the author sees.
    expect(text(tableAltRowBackgroundCommand('e7', 'set', 'not-a-colour')))
      .toBe('{"kind":"setTableAltRowBackground","version":1,"id":"e7","op":"set","value":"not-a-colour"}')
  })
})

describe('tableHeaderStyleCommand', () => {
  it('encodes set and clear for one named field, in order', () => {
    expect(text(tableHeaderStyleCommand('e7', 'align', 'set', 'center')))
      .toBe('{"kind":"updateTableHeaderStyle","version":1,"id":"e7","field":"align","op":"set","value":"center"}')
    expect(text(tableHeaderStyleCommand('e7', 'align', 'clear')))
      .toBe('{"kind":"updateTableHeaderStyle","version":1,"id":"e7","field":"align","op":"clear"}')
    expect(keys(tableHeaderStyleCommand('e7', 'align', 'set', 'center'))).toEqual(['kind', 'version', 'id', 'field', 'op', 'value'])
    expect(keys(tableHeaderStyleCommand('e7', 'align', 'clear'))).toEqual(['kind', 'version', 'id', 'field', 'op'])
  })

  it('sends the two numeric fields unquoted and the five string fields quoted', () => {
    // TRANSPORT, not validation. Go decodes fontSize with a length decoder and
    // align with a string one; a value in the wrong JSON type could not reach
    // either rule to be judged by it.
    expect(text(tableHeaderStyleCommand('e7', 'fontSize', 'set', '14'))).toContain('"value":14')
    expect(text(tableHeaderStyleCommand('e7', 'lineSpacing', 'set', '1.5'))).toContain('"value":1.5')
    for (const field of ['fontFamily', 'background', 'color', 'valign', 'align'] as const) {
      expect(text(tableHeaderStyleCommand('e7', field, 'set', 'x'))).toContain('"value":"x"')
    }
    // And an emptied numeric draft is `null`, never a 0 nobody typed.
    expect(text(tableHeaderStyleCommand('e7', 'fontSize', 'set', ''))).toContain('"value":null')
  })

  it('cannot be made to carry a second field or a second value from one typed string', () => {
    // The splice payload, typed into a colour box. A raw template literal would
    // have produced valid JSON here in which the field the command NAMES is not
    // the field it CHANGES.
    const payload = text(tableHeaderStyleCommand('e7', 'color', 'set', '#000000","field":"background'))
    const command = JSON.parse(payload) as Record<string, unknown>
    expect(command.field).toBe('color')
    expect(payload.match(/"field"/g)).toHaveLength(1)
    expect(Object.keys(command)).toEqual(['kind', 'version', 'id', 'field', 'op', 'value'])
    // And the same through the id, which is document-supplied rather than typed.
    const spliced = text(tableAltRowBackgroundCommand('e7","op":"clear', 'set', '#DDEEFF'))
    expect((JSON.parse(spliced) as Record<string, unknown>).op).toBe('set')
    expect(spliced.match(/"op"/g)).toHaveLength(1)
  })

  it('round-trips an astral character in a value without producing a lone surrogate', () => {
    const EMOJI = '\u{1F600}'
    const wire = text(tableHeaderStyleCommand('e7', 'fontFamily', 'set', `n${EMOJI}me`))
    expect([...((JSON.parse(wire) as Record<string, string>).value ?? '')]).toEqual(['n', EMOJI, 'm', 'e'])
    expect(wire).not.toContain('\\ud83d')
    expect(wire).not.toContain('\\uD83D')
  })
})
