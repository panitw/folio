import { describe, expect, it } from 'vitest'
import { addTableColumnCommand, moveTableColumnCommand, removeTableColumnCommand, updateTableColumnCommand } from './table-column-command'

const decode = (command: ArrayBuffer): Record<string, unknown> => JSON.parse(new TextDecoder().decode(command)) as Record<string, unknown>

describe('table-column command bytes', () => {
  it('uses complete JSON string encoding while retaining the closed envelope order', () => {
    const control = 'A\u0000\b\f\n\r\t\\/" สวัสดี'
    expect(decode(updateTableColumnCommand('e7', 'e8', 'header', control))).toEqual({ kind: 'updateTableColumn', version: 1, id: 'e7', columnId: 'e8', field: 'header', value: control })
    expect(new TextDecoder().decode(addTableColumnCommand('e7', 1))).toBe('{"kind":"addTableColumn","version":1,"id":"e7","index":1}')
    expect(decode(removeTableColumnCommand('e7', 'e8'))).toMatchObject({ kind: 'removeTableColumn', id: 'e7', columnId: 'e8' })
    expect(decode(moveTableColumnCommand('e7', 'e8', 0))).toMatchObject({ kind: 'moveTableColumn', id: 'e7', columnId: 'e8', toIndex: 0 })
  })

  it('does not construct an invalid JSON number for non-finite local input', () => {
    expect(decode(updateTableColumnCommand('e7', 'e8', 'width', Number.NaN)).value).toBeNull()
  })
})
