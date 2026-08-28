import { useLayoutEffect, useRef, useState, type KeyboardEvent } from 'react'
import type { TableColumns } from './engine-protocol'

type Field = 'header' | 'width' | 'align'
type ActiveCell = Readonly<{ row: number; column: number }>
type Candidate = Readonly<{ collection: string; field: string }>
type Props = Readonly<{ projection: TableColumns; busy: boolean; error?: string; candidates: ReadonlyArray<Candidate>; sampleAvailable: boolean; onClose: () => void; onAdd: (index: number) => void; onRemove: (id: string) => void; onMove: (id: string, toIndex: number) => void; onUpdate: (id: string, field: Field, value: string | number) => void; onConfigure: (collection: string, alias: string) => void; onBind: (id: string, field: string) => void; onFooter: (id: string, footer: string, footerOf: string, footerFormat: string) => void }>

const cellCount = 11

export function TableEditor({ projection, busy, error, candidates, sampleAvailable, onClose, onAdd, onRemove, onMove, onUpdate, onConfigure, onBind, onFooter }: Props) {
  const columns = projection.table.columns
  const [active, setActive] = useState<ActiveCell>({ row: 0, column: 0 })
  const dialog = useRef<HTMLElement>(null)
	const matrixFocused = useRef(false)
  const focusCell = (next: ActiveCell) => {
    const row = Math.max(0, Math.min(next.row, Math.max(0, columns.length - 1)))
    const column = Math.max(0, Math.min(next.column, cellCount - 1))
		const targets = Array.from(dialog.current?.querySelectorAll<HTMLElement>('[data-matrix-cell]') ?? [])
		const preferred = targets.find((target) => target.dataset.matrixCell === `${row}:${column}`)
		const target = !preferred?.matches(':disabled') ? preferred : targets.find((target) => !target.matches(':disabled'))
    if (!target) return
		const [targetRow, targetColumn] = (target.dataset.matrixCell ?? '0:0').split(':').map(Number)
    setActive({ row: targetRow!, column: targetColumn! })
		target.focus()
  }
  // A worker re-projection replaces structural controls. Restore the logical
  // cell (or its nearest surviving neighbor) after both accepted and rejected
  // commits rather than leaving focus on a removed/disabled DOM node.
  useLayoutEffect(() => {
    if (columns.length && (!matrixFocused.current || (document.activeElement instanceof HTMLElement && document.activeElement.hasAttribute('data-matrix-cell')))) { focusCell(active); matrixFocused.current = true }
  }, [projection]) // eslint-disable-line react-hooks/exhaustive-deps
  const moveFocus = (event: KeyboardEvent<HTMLElement>, row: number, column: number) => {
    if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
		const enabled = (candidateRow: number, candidateColumn: number) => dialog.current?.querySelector<HTMLElement>(`[data-matrix-cell="${candidateRow}:${candidateColumn}"]`)?.matches(':disabled') === false
		if (event.key === 'Home' || event.key === 'End') {
			const start = event.key === 'Home' ? 0 : cellCount - 1; const step = event.key === 'Home' ? 1 : -1
			for (let candidate = start; candidate >= 0 && candidate < cellCount; candidate += step) if (enabled(row, candidate)) { focusCell({ row, column: candidate }); return }
			return
		}
		const vertical = event.key === 'ArrowUp' || event.key === 'ArrowDown'
		const step = event.key === 'ArrowUp' || event.key === 'ArrowLeft' ? -1 : 1
		for (let candidate = (vertical ? row : column) + step; candidate >= 0 && candidate < (vertical ? columns.length : cellCount); candidate += step) {
			const next = vertical ? { row: candidate, column } : { row, column: candidate }
			if (enabled(next.row, next.column)) { focusCell(next); return }
		}
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
      <div className="table-editor-heading"><div><p className="section-label">TABLE EDITOR</p><h2>Configure columns</h2><p id="table-editor-help">Collection and row alias are engine-owned. Sample data only suggests fields; it does not validate saved bindings.</p></div><button type="button" className="file-button" onClick={onClose}>Close Table Editor</button></div>
      <div className="table-editor-config" aria-label="Table row scope"><label>Root collection<input aria-label="Root collection" list="table-collection-candidates" defaultValue={projection.table.collection} disabled={busy} onBlur={(event) => { if (event.currentTarget.value !== projection.table.collection) onConfigure(event.currentTarget.value, projection.table.alias === 'row' ? '' : projection.table.alias) }} /></label><datalist id="table-collection-candidates">{[...new Set(candidates.map((candidate) => candidate.collection))].map((collection) => <option key={collection} value={collection} />)}</datalist><label>Row alias<input aria-label="Row alias" defaultValue={projection.table.alias} disabled={busy} onBlur={(event) => { if (event.currentTarget.value !== projection.table.alias) onConfigure(projection.table.collection, event.currentTarget.value === 'row' ? '' : event.currentTarget.value) }} /></label><p className="honest-note">{sampleAvailable ? 'Amber candidate values are local discovery hints only; the engine validates saved bindings.' : 'No sample data is loaded. Enter a root collection and row field; the engine validates saved bindings.'}</p></div>
      {columns.length === 0 ? <div className="table-editor-empty"><p>No columns yet. Add a column to start the matrix.</p><button type="button" className="file-button" disabled={busy} onClick={() => onAdd(0)}>Add column</button></div> : <div role="grid" aria-label="Table columns" aria-describedby="table-editor-help" aria-rowcount={columns.length + 1} aria-colcount={cellCount} className="table-matrix">
        <div role="row" aria-rowindex={1} className="matrix-header"><span role="columnheader">Header</span><span role="columnheader">Width (pt)</span><span role="columnheader">Cell alignment</span><span role="columnheader">Row field</span><span role="columnheader">Footer</span><span role="columnheader">Footer source</span><span role="columnheader">Footer format</span><span role="columnheader">Move earlier</span><span role="columnheader">Move later</span><span role="columnheader">Remove</span><span role="columnheader">Add after</span></div>
        {columns.map((column, index) => <div role="row" aria-rowindex={index + 2} aria-selected={active.row === index} className="matrix-row" key={column.id}>
          <label role="gridcell" aria-colindex={1}>Header<input {...matrixCell(index, 0)} aria-label={`Header for column ${index + 1}`} disabled={busy} defaultValue={column.header} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.header) onUpdate(column.id, 'header', event.currentTarget.value) }} /></label>
          <label role="gridcell" aria-colindex={2}>Width<input {...matrixCell(index, 1)} aria-label={`Width for column ${index + 1} in points`} disabled={busy} type="number" min="1" step="1" defaultValue={column.width / 1000} onBlur={(event) => { const value = Number(event.currentTarget.value); if (!busy && Number.isFinite(value) && value > 0 && value * 1000 !== column.width) onUpdate(column.id, 'width', value) }} /></label>
          <label role="gridcell" aria-colindex={3}>Alignment<select {...matrixCell(index, 2)} aria-label={`Cell alignment for column ${index + 1}`} disabled={busy} value={column.align} onChange={(event) => { if (!busy) onUpdate(column.id, 'align', event.target.value) }}><option value="left">Left</option><option value="center">Center</option><option value="right">Right</option></select></label>
          <label role="gridcell" aria-colindex={4}>Row field<input {...matrixCell(index, 3)} aria-label={`Row field for column ${index + 1}`} list={`table-field-candidates-${index}`} disabled={busy || !column.rowFieldEditable} aria-describedby={!column.rowFieldEditable ? `binding-display-${index}` : undefined} defaultValue={column.rowField} onBlur={(event) => { const field = event.currentTarget.value; if (!busy && column.rowFieldEditable && field && field !== column.rowField) onBind(column.id, field) }} /><output id={`binding-display-${index}`} aria-label={`Binding for column ${index + 1}`}>{column.binding}</output><datalist id={`table-field-candidates-${index}`}>{candidates.filter((candidate) => candidate.collection === projection.table.collection).map((candidate) => <option key={candidate.field} value={candidate.field} />)}</datalist></label>
          <label role="gridcell" aria-colindex={5}>Footer<select {...matrixCell(index, 4)} aria-label={`Footer aggregate for column ${index + 1}`} disabled={busy} value={column.footer ?? ''} onChange={(event) => { if (!busy) { const footer = event.target.value; onFooter(column.id, footer, footer === '' || footer === 'count' ? '' : column.footerOf, footer === '' ? '' : column.footerFormat) } }}><option value="">None</option><option value="sum">Sum</option><option value="avg">Average</option><option value="count">Count</option></select></label>
          <label role="gridcell" aria-colindex={6}>Footer source<input {...matrixCell(index, 5)} aria-label={`Footer source for column ${index + 1}`} defaultValue={column.footerOf ?? ''} disabled={busy || column.footer === 'count' || !column.footer} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.footerOf) onFooter(column.id, column.footer ?? '', event.currentTarget.value, column.footerFormat ?? '') }} /></label>
          <label role="gridcell" aria-colindex={7}>Footer format<input {...matrixCell(index, 6)} aria-label={`Footer format for column ${index + 1}`} defaultValue={column.footerFormat ?? ''} disabled={busy || !column.footer} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.footerFormat) onFooter(column.id, column.footer ?? '', column.footerOf ?? '', event.currentTarget.value) }} /></label>
          <span role="gridcell" aria-colindex={8} className="matrix-actions"><button {...matrixCell(index, 7)} type="button" aria-label={`Move column ${index + 1} earlier`} disabled={busy || index === 0} onClick={() => onMove(column.id, index - 1)}>↑</button></span>
          <span role="gridcell" aria-colindex={9} className="matrix-actions"><button {...matrixCell(index, 8)} type="button" aria-label={`Move column ${index + 1} later`} disabled={busy || index === columns.length - 1} onClick={() => onMove(column.id, index + 1)}>↓</button></span>
          <span role="gridcell" aria-colindex={10}><button {...matrixCell(index, 9)} type="button" aria-label={`Remove column ${index + 1}`} disabled={busy} onClick={() => onRemove(column.id)}>Remove</button></span>
          <span role="gridcell" aria-colindex={11}><button {...matrixCell(index, 10)} type="button" aria-label={`Add column after column ${index + 1}`} disabled={busy} onClick={() => onAdd(index + 1)}>Add after</button></span>
        </div>)}
      </div>}
      {error && <p role="alert" className="file-message">{error}</p>}
    </div>
  </section>
}
