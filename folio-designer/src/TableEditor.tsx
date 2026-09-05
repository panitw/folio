import { useLayoutEffect, useRef, useState, type FocusEvent, type KeyboardEvent } from 'react'
import type { TableColumns } from './engine-protocol'
import { isHexColour, swatchColor } from './swatch-color'
import type { TableHeaderStyleField } from './table-style-command'

type Field = 'header' | 'width' | 'align'
type ActiveCell = Readonly<{ row: number; column: number }>
type Candidate = Readonly<{ collection: string; field: string }>
type Props = Readonly<{ projection: TableColumns; busy: boolean; error?: string; candidates: ReadonlyArray<Candidate>; sampleAvailable: boolean; onClose: () => void; onAdd: (index: number) => void; onRemove: (id: string) => void; onMove: (id: string, toIndex: number) => void; onUpdate: (id: string, field: Field, value: string | number) => void; onConfigure: (collection: string, alias: string) => void; onBind: (id: string, field: string) => void; onFooter: (id: string, footer: string, footerOf: string, footerFormat: string) => void; onHeaderHeight: (height: string) => void; onAltRowBackground: (operation: 'set' | 'clear', value?: string) => void; onHeaderStyle: (field: TableHeaderStyleField, operation: 'set' | 'clear', value?: string) => void }>

const cellCount = 11

// A projected thousandths count as the author reads and types it. The engine
// carries lengths in millipoints and the line-spacing ratio in thousandths, and
// both divide by the same 1000 to reach the number a person types — points for
// the two lengths, a bare ratio for the spacing. Every box in this panel is in
// author units, exactly as the matrix's own Width column already is.
const authored = (thousandths: number): string => String(thousandths / 1000)

// What the document WILL USE for a field the author has not set — the engine's
// own answer, projected beside the committed one, never worked out here.
//
// AN EMPTY RESOLVED STRING DOES NOT MEAN THE SAME THING FOR EVERY FIELD, so
// each caller says what its own empty means rather than sharing one sentence.
// For a BACKGROUND, empty is literally nothing: the cascade found no fill and
// none is painted. For TEXT COLOUR it is not — the header still draws, in the
// renderer's own default ink — and "Using: nothing" claimed the header would
// print with no colour at all, which is the one thing that cannot happen. For a
// FONT FAMILY, empty means no chain is declared anywhere on this table, which
// is a third thing again.
const resolvedNote = (value: string, whenEmpty: string): string => value === '' ? whenEmpty : `Using: ${value}`

export function TableEditor({ projection, busy, error, candidates, sampleAvailable, onClose, onAdd, onRemove, onMove, onUpdate, onConfigure, onBind, onFooter, onHeaderHeight, onAltRowBackground, onHeaderStyle }: Props) {
  const table = projection.table
  const columns = table.columns
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
  // ONE COMMAND PER BURST, AND `busy` IS ENOUGH TO GIVE IT — MEASURED, because
  // it does not look like it should be.
  //
  // <input type="color"> fires `onChange` continuously while the author drags,
  // and each of those is an engine command and an undo entry; worse,
  // commitTableColumn's revision-mismatch branch calls revokeTableEditor(), so
  // a burst could close the panel out from under the author. The obvious
  // reading is that `busy` — React state — lags a tick and cannot stop the
  // second event, and that a synchronous ref (App.tsx:PropertyDraft's
  // `pendingRef`) is needed instead.
  //
  // IT IS NOT, HERE, and that was measured rather than assumed. Three change
  // events dispatched inside ONE act() — the most batching-friendly shape
  // available — enter this handler three times and produce exactly ONE command.
  // React classifies `change` and `input` as DISCRETE and flushes their state
  // updates before delivering the next event, so `busy` is already true on the
  // second entry. App.tsx:commitTableColumn's own `tableEditorBusy` guard holds
  // the same property a second time: removing EITHER leaves one command,
  // removing BOTH gives three, which is how App.test.tsx's burst test was shown
  // not to be vacuous.
  //
  // So no synchronous ref is added. It would be a third gate deciding nothing,
  // and its release path — a ref can only be cleared by a later render — could
  // strand the control on the one branch of commitTableColumn that returns
  // without changing any state.
  const dispatchOnce = (send: () => void) => {
    if (busy) return
    send()
  }
  // COMMIT ON BLUR, NO DRAFT, exactly as every other control in this panel
  // does: AD-15 owns the document, so an emptied box IS a clear and a box the
  // author did not touch sends nothing. There is no Apply and no Cancel to add
  // one to.
  //
  // A BLUR THAT LANDS WHILE A COMMAND IS IN FLIGHT IS DISCARDED, AND IT MUST
  // NOT BE DISCARDED SILENTLY. `if (busy) return` alone left the box holding
  // text the document does not hold and never would — the author saw their
  // value sitting in the field with nothing to say it had gone nowhere. Bumping
  // `restore` re-keys these boxes so React re-applies each one's
  // `defaultValue`, which IS the committed value. It is a remount nonce and NOT
  // a draft: it carries no author data and there is still nothing to Apply.
  //
  // AND A NUMBER INPUT REPORTING `badInput` COMMITS NOTHING. A browser reports
  // an unparseable number input's value as '', which this handler would read as
  // an emptied box and therefore as a CLEAR — so typing garbage into the font
  // size deleted the field. The garbage stays on screen for the author to fix
  // and the document is left alone.
  const [restore, setRestore] = useState(0)
  const boxKey = (committed: string | number) => `${restore}:${committed}`
  const commitStyleText = (field: TableHeaderStyleField, committed: string) => (event: FocusEvent<HTMLInputElement>) => {
    const input = event.currentTarget
    if (busy) { setRestore((count) => count + 1); return }
    if (input.validity?.badInput) return
    const value = input.value
    if (value === committed) return
    if (value === '') onHeaderStyle(field, 'clear')
    else onHeaderStyle(field, 'set', value)
  }
  // A clearable SELECT spells absence as its own empty option rather than as a
  // separate button: the option's label carries the engine's resolved value, so
  // the one control answers both "what is set" and "what will be used".
  const styleSelect = (field: TableHeaderStyleField, label: string, committed: string, resolved: string, options: ReadonlyArray<readonly [string, string]>) =>
    <label className="table-header-field">{label}
      <select aria-label={label} disabled={busy} value={committed} onChange={(event) => { const value = event.target.value; dispatchOnce(() => { if (value === '') onHeaderStyle(field, 'clear'); else onHeaderStyle(field, 'set', value) }) }}>
        <option value="">Not set</option>
        {options.map(([value, text]) => <option key={value} value={value}>{text}</option>)}
      </select>
      <output aria-label={`Resolved ${label}`}>{resolvedNote(resolved, 'Using: nothing')}</output>
    </label>
  // A clearable COLOUR is the inspector's shipped two-control row, re-implemented
  // rather than imported (PropertyDraft reads the canvas projection and commits
  // by a different path). The unset treatment is not decoration: swatchColor('')
  // is BLACK, so an absent colour without the dashed chip reads as a colour the
  // author chose.
  //
  // `key={boxKey(committed)}` IS THE FIX FOR THE HALF-CONTROLLED ROW, and it is
  // not cosmetic. The text box is uncontrolled (`defaultValue`) while the chip
  // beside it is controlled (`value`), so after a swatch pick or a × clear
  // committed and re-projected, the chip moved and the BOX STILL SHOWED THE OLD
  // HEX. The author's next blur on that box then compared stale DOM text
  // against the new committed value, found them different, and sent
  // `op: "set"` with the OLD colour — silently undoing the pick they had just
  // made. Keying the input on the committed value remounts it whenever the
  // engine's answer changes, so the two halves of the row cannot disagree after
  // ANY commit, and no draft is introduced to do it (AD-15 still owns the
  // document; the box is still uncontrolled between commits). Every keyed box in
  // this section shares one spelling of the key, so the busy-restore above rides
  // the same mechanism.
  const styleColour = (field: TableHeaderStyleField, label: string, committed: string, resolved: string, whenUnresolved: string) =>
    <label className="table-header-field">{label}
      <span className="table-header-control">
        <input key={boxKey(committed)} aria-label={label} disabled={busy} defaultValue={committed} onBlur={commitStyleText(field, committed)} />
        <input type="color" className={`property-swatch${isHexColour(committed) ? '' : ' property-swatch-unset'}`} aria-label={`Pick ${label}`} value={swatchColor(committed)} disabled={busy} onChange={(event) => { const value = event.target.value; dispatchOnce(() => onHeaderStyle(field, 'set', value)) }} />
        <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onHeaderStyle(field, 'clear'))}>×</button>
      </span>
      <output aria-label={`Resolved ${label}`}>{resolvedNote(resolved, whenUnresolved)}</output>
    </label>
  // `min` IS PER-CONTROL AND IT IS THE SMALLEST VALUE THE ENGINE ACCEPTS AT
  // THAT STEP. It used to be `min="0"` on every number in this section while
  // both arms behind them require a POSITIVE length — so the control advertised
  // a value the engine refuses, which the matrix's own Width cell already knew
  // not to do (`min="1"`).
  const styleNumber = (field: TableHeaderStyleField, label: string, committed: number, resolved: string, step: string, min: string) =>
    <label className="table-header-field">{label}
      <span className="table-header-control">
        <input key={boxKey(committed)} aria-label={label} type="number" min={min} step={step} disabled={busy} defaultValue={committed === 0 ? '' : authored(committed)} onBlur={commitStyleText(field, committed === 0 ? '' : authored(committed))} />
        <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onHeaderStyle(field, 'clear'))}>×</button>
      </span>
      <output aria-label={`Resolved ${label}`}>{resolvedNote(resolved, 'Using: nothing')}</output>
    </label>
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
      {/* THE HEADER SECTION SITS AFTER THE MATRIX, WHERE THE DESIGN PLACES IT,
          AND THAT DELIBERATELY CHANGES THE TAB ORDER — ruled and recorded as
          D-12.3.2. The new order is [Close Table Editor, Root collection, Row
          alias, the one active matrix cell, these controls in document order],
          so the active cell is no longer last in trapDialog's list and the wrap
          happens at "Header alignment" instead. App.test.tsx asserts both ends
          of that list.

          role="group" IS LOAD-BEARING, not decoration: an aria-label on a plain
          div with NO role is dropped by the accessibility tree, so the section
          named nothing to a screen reader. */}
      <div className="table-editor-header" role="group" aria-label="Table header and rows">
        <p className="section-label">HEADER AND ROWS</p>
        <label className="table-header-field">Header height (pt)
          <input key={boxKey(table.headerHeight)} aria-label="Header height in points" type="number" min="1" step="1" disabled={busy} defaultValue={authored(table.headerHeight)} onBlur={(event) => { const input = event.currentTarget; if (busy) { setRestore((count) => count + 1); return } if (input.validity?.badInput) return; if (input.value !== authored(table.headerHeight)) onHeaderHeight(input.value) }} />
          <output aria-label="Header height note">Required by the format, so it has no clear.</output>
        </label>
        <label className="table-header-field">Alternating row background
          <span className="table-header-control">
            <input key={boxKey(table.altRowBackground)} aria-label="Alternating row background" disabled={busy} defaultValue={table.altRowBackground} onBlur={(event) => { const input = event.currentTarget; if (busy) { setRestore((count) => count + 1); return } const value = input.value; if (value === table.altRowBackground) return; if (value === '') onAltRowBackground('clear'); else onAltRowBackground('set', value) }} />
            <input type="color" className={`property-swatch${isHexColour(table.altRowBackground) ? '' : ' property-swatch-unset'}`} aria-label="Pick Alternating row background" value={swatchColor(table.altRowBackground)} disabled={busy} onChange={(event) => { const value = event.target.value; dispatchOnce(() => onAltRowBackground('set', value)) }} />
            <button type="button" className="property-inline-action" aria-label="Clear Alternating row background" title="Clear Alternating row background" disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onAltRowBackground('clear'))}>×</button>
          </span>
          <output aria-label="Alternating row background note">Odd rows only; cleared rows use the table background.</output>
        </label>
        <label className="table-header-field">Header font family
          <span className="table-header-control">
            <input key={boxKey(table.headerFontFamily)} aria-label="Header font family" disabled={busy} defaultValue={table.headerFontFamily} onBlur={commitStyleText('fontFamily', table.headerFontFamily)} />
            <button type="button" className="property-inline-action" aria-label="Clear Header font family" title="Clear Header font family" disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onHeaderStyle('fontFamily', 'clear'))}>×</button>
          </span>
          <output aria-label="Resolved Header font family">{resolvedNote(table.headerFontFamilyResolved, 'Using: nothing — this table names no font chain')}</output>
        </label>
        {styleNumber('fontSize', 'Header font size (pt)', table.headerFontSize, `${authored(table.headerFontSizeResolved)}pt`, '0.5', '0.5')}
        {styleNumber('lineSpacing', 'Header line spacing', table.headerLineSpacing, authored(table.headerLineSpacingResolved), '0.1', '0.1')}
        {styleColour('background', 'Header background', table.headerBackground, table.headerBackgroundResolved, 'Using: nothing — no fill is painted')}
        {/* NOT "Using: nothing". An unresolved header COLOUR does not mean the
            header prints with no colour — it prints in the renderer's own
            default ink. The background above is the case where nothing really
            is used; this one is not, and one sentence for both was wrong for
            this one. */}
        {styleColour('color', 'Header text colour', table.headerColor, table.headerColorResolved, "Using: the renderer's default ink")}
        {styleSelect('valign', 'Header vertical alignment', table.headerValign, table.headerValignResolved, [['top', 'Top'], ['middle', 'Middle'], ['bottom', 'Bottom']])}
        {styleSelect('align', 'Header alignment', table.headerAlign, table.headerAlignResolved, [['left', 'Left'], ['center', 'Center'], ['right', 'Right']])}
        <p className="honest-note">A field left blank falls back to the table's own style and then to the format's default. The engine resolves it; the note under each control is the engine's answer, not this panel's.</p>
      </div>
      {error && <p role="alert" className="file-message">{error}</p>}
    </div>
  </section>
}
