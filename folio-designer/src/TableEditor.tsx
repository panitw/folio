import { useLayoutEffect, useRef, useState, type KeyboardEvent } from 'react'
import type { TableColumns } from './engine-protocol'

type Field = 'header' | 'width' | 'align'
type ActiveCell = Readonly<{ row: number; column: number }>
type Props = Readonly<{ projection: TableColumns; busy: boolean; error?: string; onClose: () => void; onAdd: (index: number) => void; onRemove: (id: string) => void; onMove: (id: string, toIndex: number) => void; onUpdate: (id: string, field: Field, value: string | number) => void }>

const cellCount = 7

export function TableEditor({ projection, busy, error, onClose, onAdd, onRemove, onMove, onUpdate }: Props) {
  const columns = projection.table.columns
  const [active, setActive] = useState<ActiveCell>({ row: 0, column: 0 })
  const dialog = useRef<HTMLElement>(null)
  const focusCell = (next: ActiveCell) => {
    const row = Math.max(0, Math.min(next.row, Math.max(0, columns.length - 1)))
    const column = Math.max(0, Math.min(next.column, cellCount - 1))
		const target = dialog.current?.querySelector<HTMLElement>(`[data-matrix-cell="${row}:${column}"]`)
		if (target?.matches(':disabled')) return
    setActive({ row, column })
		target?.focus()
  }
  // A worker re-projection replaces structural controls. Restore the logical
  // cell (or its nearest surviving neighbor) after both accepted and rejected
  // commits rather than leaving focus on a removed/disabled DOM node.
  useLayoutEffect(() => { if (columns.length) focusCell(active) }, [projection, error]) // eslint-disable-line react-hooks/exhaustive-deps
  const moveFocus = (event: KeyboardEvent<HTMLElement>, row: number, column: number) => {
    if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    focusCell({ row: event.key === 'ArrowUp' ? row - 1 : event.key === 'ArrowDown' ? row + 1 : row, column: event.key === 'ArrowLeft' ? column - 1 : event.key === 'ArrowRight' ? column + 1 : event.key === 'Home' ? 0 : event.key === 'End' ? cellCount - 1 : column })
  }
  const trapDialog = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') { event.preventDefault(); onClose(); return }
    if (event.key !== 'Tab') return
    const focusable = Array.from(dialog.current?.querySelectorAll<HTMLElement>('button:not([disabled]), input:not([disabled]), select:not([disabled])') ?? []).filter((element) => element.tabIndex >= 0)
    if (!focusable.length) return
    const index = focusable.indexOf(document.activeElement as HTMLElement)
    if ((!event.shiftKey && index === focusable.length - 1) || (event.shiftKey && index <= 0)) { event.preventDefault(); focusable[event.shiftKey ? focusable.length - 1 : 0]?.focus() }
  }
  const matrixCell = (row: number, column: number) => ({ 'data-matrix-cell': `${row}:${column}`, tabIndex: active.row === row && active.column === column ? 0 : -1, onFocus: () => setActive({ row, column }), onKeyDown: (event: KeyboardEvent<HTMLElement>) => moveFocus(event, row, column) })
  return <section ref={dialog} className="table-editor-backdrop" role="dialog" aria-modal="true" aria-label="Table Editor" aria-busy={busy || undefined} onKeyDownCapture={trapDialog}>
    <div className="table-editor">
      <div className="table-editor-heading"><div><p className="section-label">TABLE EDITOR</p><h2>Configure columns</h2><p id="table-editor-help">Column widths define the table width. Changes are committed by the local engine.</p></div><button type="button" className="file-button" onClick={onClose}>Close Table Editor</button></div>
      {columns.length === 0 ? <div className="table-editor-empty"><p>No columns yet. Add a column to start the matrix.</p><button type="button" className="file-button" disabled={busy} onClick={() => onAdd(0)}>Add column</button></div> : <div role="grid" aria-label="Table columns" aria-describedby="table-editor-help" aria-rowcount={columns.length + 1} aria-colcount={cellCount} className="table-matrix">
        <div role="row" aria-rowindex={1} className="matrix-header"><span role="columnheader">Header</span><span role="columnheader">Width (pt)</span><span role="columnheader">Cell alignment</span><span role="columnheader">Move earlier</span><span role="columnheader">Move later</span><span role="columnheader">Remove</span><span role="columnheader">Add after</span></div>
        {columns.map((column, index) => <div role="row" aria-rowindex={index + 2} aria-selected={active.row === index} className="matrix-row" key={column.id}>
          <label role="gridcell" aria-colindex={1}>Header<input {...matrixCell(index, 0)} aria-label={`Header for column ${index + 1}`} defaultValue={column.header} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.header) onUpdate(column.id, 'header', event.currentTarget.value) }} /></label>
          <label role="gridcell" aria-colindex={2}>Width<input {...matrixCell(index, 1)} aria-label={`Width for column ${index + 1} in points`} type="number" min="1" step="1" defaultValue={column.width / 1000} onBlur={(event) => { const value = Number(event.currentTarget.value); if (!busy && Number.isFinite(value) && value > 0 && value * 1000 !== column.width) onUpdate(column.id, 'width', value) }} /></label>
          <label role="gridcell" aria-colindex={3}>Alignment<select {...matrixCell(index, 2)} aria-label={`Cell alignment for column ${index + 1}`} value={column.align} onChange={(event) => { if (!busy) onUpdate(column.id, 'align', event.target.value) }}><option value="left">Left</option><option value="center">Center</option><option value="right">Right</option></select></label>
          <span role="gridcell" aria-colindex={4} className="matrix-actions"><button {...matrixCell(index, 3)} type="button" aria-label={`Move column ${index + 1} earlier`} disabled={busy || index === 0} onClick={() => onMove(column.id, index - 1)}>↑</button></span>
          <span role="gridcell" aria-colindex={5} className="matrix-actions"><button {...matrixCell(index, 4)} type="button" aria-label={`Move column ${index + 1} later`} disabled={busy || index === columns.length - 1} onClick={() => onMove(column.id, index + 1)}>↓</button></span>
          <span role="gridcell" aria-colindex={6}><button {...matrixCell(index, 5)} type="button" aria-label={`Remove column ${index + 1}`} disabled={busy} onClick={() => onRemove(column.id)}>Remove</button></span>
          <span role="gridcell" aria-colindex={7}><button {...matrixCell(index, 6)} type="button" aria-label={`Add column after column ${index + 1}`} disabled={busy} onClick={() => onAdd(index + 1)}>Add after</button></span>
        </div>)}
      </div>}
      {error && <p role="alert" className="file-message">{error}</p>}
    </div>
  </section>
}
