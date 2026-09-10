import { useEffect, useLayoutEffect, useRef, useState, type FocusEvent, type KeyboardEvent } from 'react'
import { MAX_ENGINE_HISTORY_ENTRIES, type TableColumns } from './engine-protocol'
import { alignSegments, SegmentedControl } from './segmented-control'
import { isHexColour, swatchColor } from './swatch-color'
import type { TableHeaderStyleField } from './table-style-command'

type Field = 'header' | 'width' | 'align'
type ActiveCell = Readonly<{ row: number; column: number }>
type Candidate = Readonly<{ collection: string; field: string }>
type Props = Readonly<{ projection: TableColumns; busy: boolean; fileBusy: boolean; discarding: boolean; error?: string; candidates: ReadonlyArray<Candidate>; sampleAvailable: boolean; band?: string; availableWidth?: number; sampleItemCount?: number; editCount: number; onClose: () => void; onCancel: () => void; onAdd: (index: number) => void; onRemove: (id: string) => void; onMove: (id: string, toIndex: number) => void; onUpdate: (id: string, field: Field, value: string | number) => void; onConfigure: (collection: string, alias: string) => void; onFooter: (id: string, footer: string, footerOf: string, footerFormat: string) => void; onHeaderHeight: (height: string) => void; onAltRowBackground: (operation: 'set' | 'clear', value?: string) => void; onHeaderStyle: (field: TableHeaderStyleField, operation: 'set' | 'clear', value?: string) => void }>

// STORY 14.7 — SIX LABELLED COLUMNS ON SCREEN, TWELVE LATTICE CELLS BEHIND
// THEM, and the two numbers are different on purpose.
//
// `aria-colcount` counts what the DESIGN draws: `#`, HEADER LABEL, BOUND
// FIELD · row scope, WIDTH, ALIGN, FOOTER AGGREGATE. `cellCount` counts what a
// KEYBOARD must reach, which is every control inside those six cells — the
// three reorder/remove affordances that used to wear column headers of their
// own, the three segments of the alignment control, and the two footer fields
// the aggregate reveals.
//
// THE LATTICE COVERS THE MAXIMAL ROW SHAPE and a smaller row simply has holes
// in it. That is free rather than clever: `moveFocus`'s `enabled()` already
// treats an ABSENT cell exactly like a disabled one (`undefined === false`),
// so the scan walks past a hole with no per-row arithmetic anywhere.
//
// ⚠ THE FOOTER CELLS ARE DERIVED FROM `alignSegments.length`, NOT WRITTEN DOWN.
// The alignment control owns `align`, `align + 1`, `align + 2`; a fourth
// segment would have written into the aggregate's slot and given two elements
// the SAME `data-matrix-cell`, which `querySelector` resolves by silently
// picking one — a roving lattice with two cells at one address and no error
// anywhere. Deriving the offset makes a widened control impossible to get wrong
// here, and `TableEditor.test.tsx` asserts every address in the dialog is
// unique so the property is checked rather than merely intended.
//
// ⚠ STORY 14.10 EMPTIED ADDRESS 4 AND MOVED NOTHING. `CELL.bound` was the Row
// field input, which [D-14.10.1] removed: binding is edited in the main window
// now. These are CONSTANTS, not an occupancy count, so address 4 is simply a
// hole in every row — designed behaviour under the paragraph above, and
// `moveFocus`'s `enabled()` walks past it exactly as it walks past a disabled
// cell. `CELL.bound` is kept rather than deleted so the addresses either side
// of it keep the numbers four pinned traversal walks were written against.
const ALIGN_CELL = 6
const CELL = { moveEarlier: 0, moveLater: 1, remove: 2, header: 3, bound: 4, width: 5, align: ALIGN_CELL, aggregate: ALIGN_CELL + alignSegments.length, footerOf: ALIGN_CELL + alignSegments.length + 1, footerFormat: ALIGN_CELL + alignSegments.length + 2 }
const cellCount = CELL.footerFormat + 1
const COLUMN_COUNT = 6

// THE FOUR BORDER EDGES, IN THE FORMAT'S OWN ORDER, which is the order the
// checkboxes are drawn in AND the order the engine's projection joins them in.
// The closed set itself is the engine's — `internal/template/closedsets.go` —
// and the loader is the door that refuses an unknown name; this is the panel's
// rendering order, exactly as `alignSegments` is the alignment control's rather
// than a second opinion about what a legal alignment is.
const BORDER_EDGES = ['top', 'right', 'bottom', 'left'] as const

// The display unit is POINTS, one decimal (D-14.2.Q3, settled product-wide).
// The stored value is millipoints and is untouched by anything in this file.
const pointsOf = (millipoints: number): string => (millipoints / 1000).toFixed(1)
const plural = (count: number, word: string): string => `${count} ${word}${count === 1 ? '' : 's'}`

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

export function TableEditor({ projection, busy, fileBusy, discarding, error, candidates, sampleAvailable, band, availableWidth, sampleItemCount, editCount, onClose, onCancel, onAdd, onRemove, onMove, onUpdate, onConfigure, onFooter, onHeaderHeight, onAltRowBackground, onHeaderStyle }: Props) {
  const table = projection.table
  const columns = table.columns
  // THE MATRIX OPENS ON THE FIRST EDITABLE CELL, NOT ON CELL ZERO. Cell zero is
  // now `Move column 1 earlier`, which is disabled on the first row — and the
  // fallback beside it is `Remove column 1`. Opening a dialog with focus parked
  // on a destructive affordance is not a thing to do by accident, and the
  // header label is where an author's attention goes anyway.
  const [active, setActive] = useState<ActiveCell>({ row: 0, column: CELL.header })
  const dialog = useRef<HTMLElement>(null)
	const matrixFocused = useRef(false)
  // Whether the MATRIX held focus at the moment this render began. Set on a
  // cell's own focus and cleared by the dialog's capture handler for any other
  // control, so a re-projection can tell "the author was in the matrix and the
  // cell they were on has just been removed" from "the author is typing in the
  // HEADER AND ROWS section and must be left alone".
  const cellHeldFocus = useRef(false)
  const emptyAdd = useRef<HTMLButtonElement>(null)
  // ⚠ AND ONE LISTENER ON THE DOCUMENT, BECAUSE THE DIALOG'S OWN CAPTURE
  // HANDLER CANNOT SEE FOCUS LEAVE IT. React's `onFocusCapture` on the dialog
  // fires only for targets INSIDE the dialog, so focus moving to something
  // behind the modal left `cellHeldFocus` still true — and a later
  // re-projection would then reclaim focus from outside the dialog and drag the
  // author back into the matrix. `focusin` on the document, in the capture
  // phase, is the only place that transition is observable.
  useEffect(() => {
    // `Event`, not `FocusEvent`: this file imports React's `FocusEvent` type,
    // which shadows the DOM one. Only `event.target` is read.
    const record = (event: Event) => {
      const target = event.target
      if (!(target instanceof Node) || dialog.current === null || !dialog.current.contains(target)) cellHeldFocus.current = false
    }
    document.addEventListener('focusin', record, true)
    return () => document.removeEventListener('focusin', record, true)
  }, [])
  // ⚠ AN ABSENT CELL AND A DISABLED CELL ARE THE SAME THING TO THIS PANEL, AND
  // UNTIL STORY 14.7 ONLY ONE OF THE TWO SCANS AGREED. `moveFocus`'s
  // `enabled()` reads `…?.matches(':disabled') === false`, so an absent cell
  // yields `undefined === false` → `false` and is skipped exactly like a
  // disabled one. `focusCell` read `!preferred?.matches(':disabled')`, which is
  // `true` when `preferred` is `undefined` — so it took the preferred branch,
  // found nothing, and returned EARLY instead of falling back the way it does
  // for a disabled cell. That single asymmetry stranded focus on
  // `document.body` the moment a revealed cell disappeared while holding it,
  // and it would have shipped green: nothing in the suite removed a cell that
  // had focus.
  //
  // The fallback now prefers the NEAREST SURVIVING CELL IN THE SAME ROW. A row
  // that loses its footer source keeps focus in its own row rather than being
  // thrown to the top of the matrix.
  const focusCell = (next: ActiveCell) => {
    const row = Math.max(0, Math.min(next.row, Math.max(0, columns.length - 1)))
    const column = Math.max(0, Math.min(next.column, cellCount - 1))
		const targets = Array.from(dialog.current?.querySelectorAll<HTMLElement>('[data-matrix-cell]') ?? [])
		const usable = (candidate: HTMLElement) => !candidate.matches(':disabled')
		const columnOf = (candidate: HTMLElement) => Number((candidate.dataset.matrixCell ?? '0:0').split(':')[1])
		const preferred = targets.find((target) => target.dataset.matrixCell === `${row}:${column}`)
		const nearestInRow = targets.filter((target) => target.dataset.matrixCell?.startsWith(`${row}:`) && usable(target))
			.sort((a, b) => Math.abs(columnOf(a) - column) - Math.abs(columnOf(b) - column))[0]
		const target = preferred !== undefined && usable(preferred) ? preferred : nearestInRow ?? targets.find(usable)
    if (!target) return
		const [targetRow, targetColumn] = (target.dataset.matrixCell ?? '0:0').split(':').map(Number)
    setActive({ row: targetRow!, column: targetColumn! })
		target.focus()
  }
  // A worker re-projection replaces structural controls. Restore the logical
  // cell (or its nearest surviving neighbor) after both accepted and rejected
  // commits rather than leaving focus on a removed/disabled DOM node.
  //
  // THE THIRD BRANCH IS STORY 14.7'S: a cell that held focus and no longer
  // exists leaves `document.activeElement` on `document.body`, which the
  // `hasAttribute('data-matrix-cell')` test cannot see. Focus is then on
  // NOBODY, so reclaiming it into the row is safe — and it is the only branch
  // that reaches the "aggregate set to none while the source held focus" case.
  useLayoutEffect(() => {
    const focused = document.activeElement
    const stranded = cellHeldFocus.current && (focused === null || focused === document.body)
    // ⚠ THE LAST COLUMN LEAVING IS NOT "NOTHING TO DO". `focusCell` has no cell
    // to land on, so this used to return and leave focus on `document.body`
    // INSIDE AN OPEN MODAL — and `trapDialog` is bound as `onKeyDownCapture` on
    // the dialog element, so a key pressed on `body` never reaches it and
    // ESCAPE STOPPED CLOSING THE DIALOG. The empty state's own `Add column` is
    // the one control left and is where the author's next action is anyway.
    if (!columns.length) {
      if (matrixFocused.current && stranded) emptyAdd.current?.focus()
      return
    }
    if (!matrixFocused.current || stranded || (focused instanceof HTMLElement && focused.hasAttribute('data-matrix-cell'))) { focusCell(active); matrixFocused.current = true }
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
  // ⚠ A VISIBLE REASON, NOT A BARE GREY-OUT AND NOT A `title`. Above the
  // engine's history bound a discard cannot land where it claims to: the ring
  // buffer has already evicted the oldest entry, so the sequence would stop one
  // edit short and the last undo would fail. The author is told that, on screen,
  // beside the button it disables — a `title` is not readable by keyboard and a
  // grey button with no sentence is a refusal with no reason.
  const overHistoryBound = editCount > MAX_ENGINE_HISTORY_ENTRIES
  const trapDialog = (event: KeyboardEvent<HTMLElement>) => {
    // ⚠ ESCAPE IS SWALLOWED WHILE THE DISCARD IS UNWINDING, AND ONLY THEN.
    // Escape is `Done`, and `Done` tears the session down: mid-sequence that
    // advances `tableEditorSession`, so App's Cancel loop hits its own
    // session-teardown guard and returns BEFORE it installs the snapshot it
    // reached — the engine k undos back while the canvas and the preview still
    // show the pre-Cancel document, with nothing on screen saying so.
    //
    // GATED ON `discarding` AND NEVER ON `busy`. `busy` is not cleared by
    // `setCurrentSnapshot`'s `clearDocumentInteraction` branch, so a latched
    // `busy` plus a gated Escape would make this modal impossible to close at
    // all — a worse defect than the one being fixed. `discarding` is the
    // compensating sequence's own in-flight flag and nothing else's; an
    // ordinary blur commit leaves Escape working exactly as it did.
    if (event.key === 'Escape') { event.preventDefault(); if (discarding) return; onClose(); return }
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
  //
  // `whenUnresolved` IS PER-CONTROL for the same reason `styleColour`'s is: an
  // empty resolved string does not mean the same thing for every number. For the
  // header font size it means the cascade found nothing, which is what the
  // default sentence says; for the header BORDER width it means no border is
  // painted at all, which is a different fact and gets its own sentence.
  //
  // ⚠ `committed` IS `number | string`, AND THE UNION IS THE WHOLE OF THIS
  // CONTROL'S ABSENCE HANDLING. Two of the three lengths this factory draws
  // spell absence as `0` because a zero of their own is not a meaningful
  // declaration (a zero font size, a zero line spacing); the header border width
  // spells absence as `''` because a zero width IS a declaration — the thinnest
  // device line PDF can draw — and a number whose absence is spelled `0` cannot
  // tell the two apart. `Number()` on the way in would fold `'0'` straight back
  // to the absent case and undo the projection's string spelling ONE LAYER UP,
  // which is exactly the defect the string spelling was introduced to remove:
  // a declared zero-width border would render an EMPTY box, indistinguishable
  // from unauthored, and clearing a declared `'0'` would not even remount the
  // box because `boxKey(0)` is the key it already had.
  //
  // So the branch is on the STRING, in `authoredBox` below, and `boxKey` is
  // handed the committed value UNCONVERTED so that `''` and `'0'` are two keys.
  const authoredBox = (committed: number | string): string => typeof committed === 'string' ? (committed === '' ? '' : authored(Number(committed))) : (committed === 0 ? '' : authored(committed))
  const styleNumber = (field: TableHeaderStyleField, label: string, committed: number | string, resolved: string, step: string, min: string, whenUnresolved = 'Using: nothing') =>
    <label className="table-header-field">{label}
      <span className="table-header-control">
        <input key={boxKey(committed)} aria-label={label} type="number" min={min} step={step} disabled={busy} defaultValue={authoredBox(committed)} onBlur={commitStyleText(field, authoredBox(committed))} />
        <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onHeaderStyle(field, 'clear'))}>×</button>
      </span>
      <output aria-label={`Resolved ${label}`}>{resolvedNote(resolved, whenUnresolved)}</output>
    </label>
  // A FACT THE ENGINE DERIVES, STATED AS ONE — never offered as a control and
  // never drawn as a disabled control either. `Row height` and `Repeat on
  // continuation pages` are both things the design drew as settings and the
  // format has no field for: a row is as tall as its content, and the header
  // repeat is the engine's own pagination behaviour. The idiom is the shipped
  // one (`Binding.dc.html`'s badge, plain-sentence reason, dimmed value) and it
  // honours DESIGN.md's "state the reason next to anything disabled" — a greyed
  // box with no reason is the thing that rule exists against.
  const derivedFact = (name: string, value: string, badge: string, reason: string) =>
    <div className="table-header-field table-header-fact">
      <span className="table-header-fact-name">{name}<span className="table-header-fact-badge">{badge}</span></span>
      <output className="table-header-fact-value" aria-label={`${name} note`} aria-live="off">{value}</output>
      <p className="table-header-fact-reason">{reason}</p>
    </div>
  const matrixCell = (row: number, column: number) => ({ 'data-matrix-cell': `${row}:${column}`, tabIndex: active.row === row && active.column === column ? 0 : -1, onFocus: () => { cellHeldFocus.current = true; setActive({ row, column }) }, onKeyDown: (event: KeyboardEvent<HTMLElement>) => moveFocus(event, row, column) })
  // ONE ALIGNMENT CONTROL IN THE PRODUCT — this is the inspector's own module,
  // imported, not a copy. Only the labels are per-row, because three buttons
  // named "Align left" in one dialog name nothing in particular.
  //
  // THREE SEGMENTS, NEVER FOUR. `justifySegment` is not imported here at all:
  // `ColumnAlignTokens` has three members and a table cell would draw a
  // justified value at its start edge regardless.
  const columnAlignSegments = (index: number) => alignSegments.map((segment) => ({ ...segment, label: `${segment.label} for column ${index + 1}` }))
  // THE WIDTH BUDGET, AND ITS AUTHORITY IS THE ENGINE'S OWN RULE RATHER THAN A
  // second opinion invented here. `containComponent` refuses a component whose
  // `width > band.Width - x`, and a table's projected width is Σ its column
  // widths — so the space the columns have is the band's width less the table's
  // own x, which is what `availableWidth` carries. Nothing is computed about
  // the document; this states, before the render fails, the arithmetic the
  // author would otherwise meet as a refusal afterwards.
  const totalWidth = columns.reduce((sum, column) => sum + column.width, 0)
  const exactFit = availableWidth !== undefined && availableWidth >= 0 && totalWidth === availableWidth
  // A DELTA THIS READ-OUT CANNOT PRINT IS SAID IN WORDS, NEVER ROUNDED TO
  // `0.0`. Widths are millipoints and this line shows one decimal, so a
  // difference under 100 millipoints rounds away — and `Σ 500.0 of 500.0
  // available · 0.0 pt over` with no `exact` badge is the read-out contradicting
  // itself: the numbers say exact, the badge says not. The badge stays on TRUE
  // equality (the engine's `containComponent` compares the stored values, not
  // the printed ones) and the sentence stops claiming a figure it does not have.
  const deltaOf = (millipoints: number): string => millipoints < 100 ? 'under 0.1' : pointsOf(millipoints)
  // ⚠ AND THE UNIT IS NOT SPELLED ON ONE FIGURE OUT OF THREE. It was `pt` on
  // the delta alone, which reads as if the other two were something else. The
  // figures are all points, they are all unit-less here, and the unit is
  // carried where an author types one: the `pt` beside every WIDTH box.
  const budgetLine = availableWidth === undefined
    ? `Σ ${pointsOf(totalWidth)} · the band's remaining width is unknown here`
    // NEGATIVE AVAILABLE WIDTH IS A REAL STATE, NOT A GUARD AGAINST NOTHING.
    // `band.width − table.x` goes negative whenever the table's x sits past its
    // band's right edge — reachable by loading a hand-edited file, and reachable
    // by a margin change that shrinks the band under an x the document already
    // holds. Printing `of -50.0 available` would state a negative quantity of
    // space; this states the fact that produced it.
    : availableWidth < 0 ? `Σ ${pointsOf(totalWidth)} · this table starts ${pointsOf(-availableWidth)} outside its band, so no column width fits`
    : exactFit ? `Σ ${pointsOf(totalWidth)} of ${pointsOf(availableWidth)} available`
    : totalWidth < availableWidth ? `Σ ${pointsOf(totalWidth)} of ${pointsOf(availableWidth)} available · ${deltaOf(availableWidth - totalWidth)} to spare`
    : `Σ ${pointsOf(totalWidth)} of ${pointsOf(availableWidth)} available · ${deltaOf(totalWidth - availableWidth)} over — this table will not fit its band`
  // THE ITEM COUNT IS UNKNOWN OR IT IS A NUMBER, and it is never 0 standing in
  // for either. `sampleItemCount` is `SampleNode.count`, which is the TRUE
  // count of a sampled collection rather than the handful of items the parser
  // kept; it is absent when no sample is loaded and absent again when the node
  // itself was truncated away. An empty collection really does report 0.
  const sampleLine = !sampleAvailable ? 'item count unknown · no sample loaded'
    : sampleItemCount === undefined ? 'item count unknown'
    : `${plural(sampleItemCount, 'item')} in sample`
  // NO `collection === ''` ARM. `isTableColumns` (`engine-protocol.ts`) admits a
  // projection only when `table.collection.length > 0`, so an empty collection
  // is a shape the validator refuses before this panel ever sees it. A branch
  // for it would read as a state the product can reach.
  const scopeLine = `${table.collection} · ${sampleLine} · band: ${band ?? 'unknown'}`
  // `footer` IS A NON-OPTIONAL `'' | 'sum' | 'avg' | 'count'` on the wire
  // (`engine-protocol.ts:TableColumn`), so `!== undefined` here — and the
  // `?? ''` this file used to spell on `footer`, `footerOf` and `footerFormat` —
  // described a partial projection that cannot arrive. Empty is `''`.
  const aggregateCount = columns.filter((column) => column.footer !== '').length
  // The reveal predicate IS the predicate that used to disable these two boxes,
  // unchanged: a source is live once an aggregate is chosen and that aggregate
  // is not `count` (count takes none); a format is live once an aggregate is
  // chosen. Story 14.7 changes only whether the box is PRESENT — the engine's
  // rules do not move.
  const showsFooterOf = (footer: string) => footer !== '' && footer !== 'count'
  const showsFooterFormat = (footer: string) => footer !== ''
  // THE HEADER BORDER'S THREE READINGS, AND NONE OF THEM RE-DERIVES THE ENGINE.
  //
  // `committedEdges` is what the DOCUMENT declares, which is what the boxes
  // above it must show: a checkbox reflecting the RESOLVED set could never be
  // unchecked back to absent.
  //
  // `borderPainted` is read off `edgesResolved` and off nothing else, because it
  // is still the only member that can answer it — though NOT for the reason it
  // once was. The resolved width no longer collapses the two states: it is a
  // string, so `''` is "no border reaches this row" and `'0'` is "a declared
  // zero-width border". What the width cannot report is the SECOND way to paint
  // nothing: a border that does resolve while its declared `edges` names no
  // side. An empty resolved edge list is exactly the emitter's own no-stroke
  // condition (`internal/pdf/rectdoc.go`: `HasStroke && (Top || Right || Bottom
  // || Left)`), and it covers both ways at once.
  //
  // ⚠ `borderAuthored` READS THE THREE COMMITTED MEMBERS, AND ALL THREE SPELL
  // ABSENCE AS `''`. The width used to be a NUMBER whose absence was 0 — and
  // `parse_bands.go` accepts a zero width as "the thinnest device line PDF can
  // draw, not an absent border", so a header border authored as nothing but
  // `{"width": 0}` read as unauthored and this panel then said "nothing here is
  // set, so this header row takes the table's own border" about a header that
  // had taken the border over. The projection now spells the width as a string
  // in the SAME thousandths — `''` absent, `'0'` a declared zero — so absence
  // and a declared zero are two different values here, and each of the three
  // disjuncts below is independently sufficient.
  //
  // ⚠ AND THE UN-AUTHORED BRANCH MAKES NO CLAIM ABOUT PROVENANCE, DELIBERATELY.
  // It once said the header "takes the table's own border". It cannot know that:
  // a header declaring `{"border": {}}` — an empty block, which the loader
  // admits and `resolveHeaderStyle` takes WHOLE — has already taken the border
  // over while committing no sub-field, so all three members read `''` and this
  // flag reads false. `{"border": {"edges": []}}` is the same class.
  // NOR CAN THE RESOLVED TRIO SETTLE IT: a header that genuinely inherits the
  // table's border also has a non-empty `edgesResolved`, so `resolved !== ''`
  // cannot tell "declares an empty block" from "inherits the table's" either.
  // A flat projection of N sub-field members cannot carry the block's N+1 bits
  // of state — its own presence plus each sub-field's — and a presence member is
  // refused from both sides of `table_header_style_test.go`'s pair/unpaired tie.
  // So the branch states what it CAN know: that nothing is authored, and what
  // the engine resolved. Provenance is left unsaid rather than guessed.
  // DO NOT REINSTATE A SENTENCE ABOUT INHERITANCE HERE — the test named
  // "makes no claim about provenance" exists to catch exactly that.
  const committedEdges: ReadonlyArray<string> = table['headerBorder.edges'] === '' ? [] : table['headerBorder.edges'].split(',')
  const borderPainted = table['headerBorder.edgesResolved'] !== ''
  const borderAuthored = table['headerBorder.width'] !== '' || table['headerBorder.color'] !== '' || table['headerBorder.edges'] !== ''
  return <section ref={dialog} className="table-editor-backdrop" role="dialog" aria-modal="true" aria-label="Table Editor" aria-busy={busy || undefined} onKeyDownCapture={trapDialog} onFocusCapture={(event) => { if (!(event.target instanceof HTMLElement) || !event.target.hasAttribute('data-matrix-cell')) cellHeldFocus.current = false }}>
    <div className="table-editor">
      <div className="table-editor-heading"><div><p className="section-label">TABLE EDITOR</p><h2>Configure columns</h2><p id="table-editor-help">Structure only: columns, widths, alignment and footer totals. A column’s bound field is set in the main window — select the column on the canvas, then pick a path in the DATA tab.</p></div><output className="table-editor-scope" aria-label="Table scope" aria-live="off">{scopeLine}</output></div>
      {/* THE COLLECTION AND THE ROW ALIAS STAY EDITABLE HERE. `DW-351` records
          this as the ONLY site in the product where a table's collection can be
          changed, and Story 14.6's context bar now routes authors to it, so the
          restyle is a labelled group and nothing more — they are not moved and
          they do not become read-only. The mockup draws the collection as plain
          text and draws no alias at all; that is the mockup being silent about a
          capability, not the product losing one.

          role="group" IS LOAD-BEARING: an aria-label on a plain div with NO role
          is dropped by the accessibility tree, so this section named nothing to
          a screen reader while carrying an aria-label that read as if it did.

          ⚠ STORY 14.10 SCOPED THIS CLAIM AND DID NOT REVOKE IT. [D-14.10.1]
          moved the PER-COLUMN BOUND FIELD out of this dialog and left the
          COLLECTION and the ROW ALIAS exactly where they are, so "the only site
          in the product where a table's collection can be changed" is still
          true and DW-351 still stands. What is no longer true anywhere in this
          file is that a column's field is edited here. */}
      <div className="table-editor-config" role="group" aria-label="Table row scope"><p className="section-label">ROW SCOPE</p><label>Root collection<input aria-label="Root collection" list="table-collection-candidates" defaultValue={projection.table.collection} disabled={busy} onBlur={(event) => { if (event.currentTarget.value !== projection.table.collection) onConfigure(event.currentTarget.value, projection.table.alias === 'row' ? '' : projection.table.alias) }} /></label><datalist id="table-collection-candidates">{[...new Set(candidates.map((candidate) => candidate.collection))].map((collection) => <option key={collection} value={collection} />)}</datalist><label>Row alias<input aria-label="Row alias" defaultValue={projection.table.alias} disabled={busy} onBlur={(event) => { if (event.currentTarget.value !== projection.table.alias) onConfigure(projection.table.collection, event.currentTarget.value === 'row' ? '' : event.currentTarget.value) }} /></label><p className="honest-note">{sampleAvailable ? 'This is the only place a table’s collection and row alias can be changed. Candidate collections come from the loaded sample; the engine validates every saved binding.' : 'This is the only place a table’s collection and row alias can be changed. No sample data is loaded, so nothing is suggested; the engine validates every saved binding.'}</p></div>
      {columns.length === 0 ? <div className="table-editor-empty"><p>No columns yet. Add a column to start the matrix.</p><button ref={emptyAdd} type="button" className="file-button" disabled={busy} onClick={() => dispatchOnce(() => onAdd(0))}>Add column</button></div> : <div role="grid" aria-label="Table columns" aria-describedby="table-editor-help" aria-rowcount={columns.length + 1} aria-colcount={COLUMN_COUNT} className="table-matrix">
        <div role="row" aria-rowindex={1} className="matrix-header"><span role="columnheader">#</span><span role="columnheader">HEADER LABEL</span><span role="columnheader">BOUND FIELD · row scope</span><span role="columnheader">WIDTH</span><span role="columnheader">ALIGN</span><span role="columnheader">FOOTER AGGREGATE</span></div>
        {columns.map((column, index) => <div role="row" aria-rowindex={index + 2} aria-selected={active.row === index} className="matrix-row" key={column.id}>
          {/* THE ROW'S OWN AFFORDANCES, NOT FOUR MORE COLUMNS. Reorder and
              remove act on THIS ROW, so they live in the row's identity cell
              beside its number rather than wearing column headers that describe
              no column. Each keeps an accessible name and a lattice position,
              so nothing the eleven-column matrix could reach by keyboard has
              become unreachable. A disabled end states WHICH end it is. */}
          <span role="gridcell" aria-colindex={1} className="matrix-rail"><span className="matrix-ordinal">{index + 1}</span><span className="matrix-actions"><button {...matrixCell(index, CELL.moveEarlier)} type="button" className="matrix-affordance" aria-label={`Move column ${index + 1} earlier`} title={index === 0 ? `Column ${index + 1} is already first` : `Move column ${index + 1} earlier`} disabled={busy || index === 0} onClick={() => dispatchOnce(() => onMove(column.id, index - 1))}>↑</button><button {...matrixCell(index, CELL.moveLater)} type="button" className="matrix-affordance" aria-label={`Move column ${index + 1} later`} title={index === columns.length - 1 ? `Column ${index + 1} is already last` : `Move column ${index + 1} later`} disabled={busy || index === columns.length - 1} onClick={() => dispatchOnce(() => onMove(column.id, index + 1))}>↓</button><button {...matrixCell(index, CELL.remove)} type="button" className="matrix-affordance" aria-label={`Remove column ${index + 1}`} title={`Remove column ${index + 1}`} disabled={busy} onClick={() => dispatchOnce(() => onRemove(column.id))}>×</button></span></span>
          <span role="gridcell" aria-colindex={2}><input {...matrixCell(index, CELL.header)} aria-label={`Header for column ${index + 1}`} disabled={busy} defaultValue={column.header} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.header) onUpdate(column.id, 'header', event.currentTarget.value) }} /></span>
          {/* STORY 14.10 / [D-14.10.1] — BOUND FIELD IS DISPLAY-ONLY, AND THE
              REMOVAL IS THE POINT RATHER THAN A SIDE EFFECT.
              This cell used to carry an `<input list=…>` — the ONE value in the
              product edited somewhere other than the main window, and the one
              thing `TableEditor.dc.html` never drew: the design draws BOUND
              FIELD as a dense binding chip with no input border, no chevron and
              muted ink, visibly unlike every editable neighbour in the same row.
              The shipped input was the thing that departed from the design.
              A column's field is now bound where every other binding is made:
              select the column on the canvas, then pick a path in the DATA tab.
              This editor is structure — add, remove, reorder, width, align,
              footer aggregate — plus the table's collection and row alias, which
              [D-14.10.1] explicitly leaves editable here.
              ⚠ THE `<output>` STAYS, AND SO DOES ITS ADDRESS IN THE LATTICE'S
              ARITHMETIC. `CELL.bound = 4` is a CONSTANT, not a count of
              occupants, so address 4 is simply a hole now and `moveFocus`'s
              `enabled()` already treats an absent cell exactly like a disabled
              one. Nothing else moved. */}
          <span role="gridcell" aria-colindex={3} className="matrix-bound"><output id={`binding-display-${index}`} aria-label={`Binding for column ${index + 1}`}>{column.binding}</output></span>
          {/* THE UNIT IS SHOWN BESIDE THE NUMBER because the columnheader is
              `WIDTH`, as the design spells it, and a bare number in a design
              tool is ambiguous. It is `pt` and not `mm`: D-14.2.Q3 settled the
              display unit product-wide, and the mockup's millimetres are
              mockup fidelity against zero millimetres in the product. */}
          <span role="gridcell" aria-colindex={4} className="matrix-width"><input {...matrixCell(index, CELL.width)} aria-label={`Width for column ${index + 1} in points`} disabled={busy} type="number" min="1" step="1" defaultValue={column.width / 1000} onBlur={(event) => { const value = Number(event.currentTarget.value); if (!busy && Number.isFinite(value) && value > 0 && value * 1000 !== column.width) onUpdate(column.id, 'width', value) }} /><span className="matrix-unit">pt</span></span>
          <span role="gridcell" aria-colindex={5}><SegmentedControl label={`Cell alignment for column ${index + 1}`} segments={columnAlignSegments(index)} current={column.align} disabled={busy} onPick={(value) => { if (value !== column.align) dispatchOnce(() => onUpdate(column.id, 'align', value)) }} segmentProps={(segment) => matrixCell(index, CELL.align + segment)} /></span>
          {/* ONE CONTROL WHERE THERE WERE THREE. The source and the format are
              REVEALED by the aggregate that needs them rather than sitting
              there greyed out: a control that can never hold a value for this
              aggregate is absent, not disabled-and-mysterious. */}
          <span role="gridcell" aria-colindex={6} className="matrix-footer-cell"><select {...matrixCell(index, CELL.aggregate)} aria-label={`Footer aggregate for column ${index + 1}`} disabled={busy} value={column.footer} onChange={(event) => { if (!busy) { const footer = event.target.value; onFooter(column.id, footer, showsFooterOf(footer) ? column.footerOf : '', showsFooterFormat(footer) ? column.footerFormat : '') } }}><option value="">none</option><option value="sum">sum</option><option value="avg">avg</option><option value="count">count</option></select>{showsFooterOf(column.footer) && <label className="matrix-reveal">source<input {...matrixCell(index, CELL.footerOf)} aria-label={`Footer source for column ${index + 1}`} defaultValue={column.footerOf} disabled={busy} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.footerOf) onFooter(column.id, column.footer, event.currentTarget.value, column.footerFormat) }} /></label>}{showsFooterFormat(column.footer) && <label className="matrix-reveal">format<input {...matrixCell(index, CELL.footerFormat)} aria-label={`Footer format for column ${index + 1}`} defaultValue={column.footerFormat} disabled={busy} onBlur={(event) => { if (!busy && event.currentTarget.value !== column.footerFormat) onFooter(column.id, column.footer, column.footerOf, event.currentTarget.value) }} /></label>}</span>
        </div>)}
      </div>}
      {/* ADD-AFTER LEFT THE ROW AND BECAME ONE CONTROL BELOW THE GRID. A
          position is reached by reorder — a step longer, never a capability
          lost — and the four row-action columns stop wearing column headers
          that describe no column. */}
      {/* ⚠ THE BUDGET RENDERS IN THE EMPTY STATE TOO, WHICH IS THE ONE MOMENT
          IT IS MOST WORTH READING: an author with no columns yet is deciding
          how many will fit, and the foot used to be gated on
          `columns.length > 0` so the answer was withheld exactly then. Only the
          BUTTON is gated — the empty state carries its own, and two controls
          with one name in one dialog name nothing.

          `aria-live="off"` ON ALL THREE READ-OUTS, DELIBERATELY. `<output>`
          carries an implicit `role="status"`, which is an assertive-by-default
          polite live region: a screen-reader user editing widths would hear the
          whole budget re-read after every committed keystroke, and the scope
          line and the column summary the same. They are read on demand, where
          the author goes looking for them; the refusals in this dialog are
          `role="alert"` and still announce. */}
      <div className="table-matrix-foot">{columns.length > 0 && <button type="button" className="file-button" disabled={busy} onClick={() => dispatchOnce(() => onAdd(columns.length))}>Add column</button>}<output className={`table-budget${exactFit ? ' table-budget-exact' : ''}`} aria-label="Width budget" aria-live="off">{budgetLine}{exactFit && <span className="table-budget-badge">exact</span>}</output></div>
      {/* THE HEADER SECTION SITS AFTER THE MATRIX, WHERE THE DESIGN PLACES IT,
          AND THAT DELIBERATELY CHANGES THE TAB ORDER — ruled and recorded as
          D-12.3.2. Story 14.7 moved `Close Table Editor` out of the heading and
          into the footer bar below, and Story 14.7b turned that one button into
          the `Cancel` / `Done` pair — so the order is now [Root collection, Row
          alias, the one active matrix cell, Add column, these controls in
          document order, Cancel, Done]. `Cancel` DROPS OUT of it whenever it is
          disabled, because the trap's query is `button:not([disabled])`.
          App.test.tsx asserts both ends of that list, re-derived from the DOM.

          role="group" IS LOAD-BEARING, not decoration: an aria-label on a plain
          div with NO role is dropped by the accessibility tree, so the section
          named nothing to a screen reader. */}
      <div className="table-editor-header" role="group" aria-label="Table header and rows">
        {/* THREE HEADINGS INSIDE ONE GROUP, NOT THREE GROUPS, AND THAT IS A
            RULING RATHER THAN A LAYOUT PREFERENCE (Story 14.8).

            Three `role="group"`s would take the shrunk sweep in
            `control-vocabulary-contract.test.tsx` from 32 group instances to 34
            and so CLEAR `GROUP_INSTANCE_FLOOR = 33` — the pinned clause would
            stop proving anything about a dropped render state and become a guard
            that cannot fail, arriving as a side effect of markup. Raising the
            floor would have been defensible, since the counted population really
            did grow; it is the wrong bar when an arm exists that spends nothing.

            PLAIN `<p>` HEADINGS WERE REFUSED ON THE ACCESSIBILITY AXIS. The
            design draws three sections; a screen-reader user would have
            perceived one undifferentiated group — this epic's own subject
            failing inside the epic. `<h3>` is a correct non-skipping descent
            under the dialog's `<h2>Configure columns</h2>`, nothing pins
            `section-label` to `<p>`, and no heading-order contract exists.

            ⚠ DO NOT PROMOTE ANY OF THESE TO A GROUP WITH `aria-labelledby`.
            That is the same arm through the back door: it is the ROLE that the
            sweep counts, not the label. */}
        <h3 className="section-label">HEADER</h3>
        <label className="table-header-field">Header height (pt)
          <input key={boxKey(table.headerHeight)} aria-label="Header height in points" type="number" min="1" step="1" disabled={busy} defaultValue={authored(table.headerHeight)} onBlur={(event) => { const input = event.currentTarget; if (busy) { setRestore((count) => count + 1); return } if (input.validity?.badInput) return; if (input.value !== authored(table.headerHeight)) onHeaderHeight(input.value) }} />
          <output aria-label="Header height note">Required by the format, so it has no clear.</output>
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
        {/* NOT A SETTING, AND SAID SO RATHER THAN DRAWN AS A DISABLED CHECKBOX.
            The mockup drew `Repeat on continuation pages` as a ticked box with a
            `REQUIRED` badge; there is no format field behind it, and DESIGN.md's
            "don't draw an affordance the product cannot honour" settles that
            against the drawing. The reason line is the shipped idiom for a
            derived value, and DESIGN.md's "state the reason next to anything
            disabled" is the rule it honours.

            ⚠ AND NOTHING HERE IS WORDED AS AN ABSOLUTE — the BADGE included.
            The engine carries `DiagCodeTableHeaderRepeatSuppressed`, its own
            record of the page where reserving the header would leave no room for
            a row, so the repeat is dropped for that page and a warning is raised.
            The mockup's `REQUIRED` promised a guarantee the engine can suspend;
            so would a badge reading `ALWAYS`, which is why this one says
            `DERIVED` — the honest claim is that the repeat is the engine's
            behaviour rather than the author's choice, and the reason line names
            the exception rather than merely avoiding the word. */}
        {derivedFact('Repeat on continuation pages', 'on', 'DERIVED', 'The header is redrawn at the top of each page a table continues onto. On a page where reserving it would leave no room for a row, the engine drops the repeat for that page only and warns.')}
        <h3 className="section-label">CELLS</h3>
        <label className="table-header-field">Alternating row background
          <span className="table-header-control">
            <input key={boxKey(table.altRowBackground)} aria-label="Alternating row background" disabled={busy} defaultValue={table.altRowBackground} onBlur={(event) => { const input = event.currentTarget; if (busy) { setRestore((count) => count + 1); return } const value = input.value; if (value === table.altRowBackground) return; if (value === '') onAltRowBackground('clear'); else onAltRowBackground('set', value) }} />
            <input type="color" className={`property-swatch${isHexColour(table.altRowBackground) ? '' : ' property-swatch-unset'}`} aria-label="Pick Alternating row background" value={swatchColor(table.altRowBackground)} disabled={busy} onChange={(event) => { const value = event.target.value; dispatchOnce(() => onAltRowBackground('set', value)) }} />
            <button type="button" className="property-inline-action" aria-label="Clear Alternating row background" title="Clear Alternating row background" disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => dispatchOnce(() => onAltRowBackground('clear'))}>×</button>
          </span>
          <output aria-label="Alternating row background note">Odd rows only; cleared rows use the table background.</output>
        </label>
        {/* The mockup drew `Row height` with the SAME dropdown chevron as the
            footer aggregate pickers, so `auto` read as one option among several.
            There is no other option and no field to hold one. */}
        {derivedFact('Row height', 'auto', 'DERIVED', 'Every row is as tall as the content it holds, so there is nothing to choose. Only the HEADER row has a height of its own.')}
        <h3 className="section-label">BORDERS</h3>
        {/* THE TABLE'S OWN BORDER IS NOT RESTATED HERE. It is authored in the
            inspector's BOX section for every non-line component including a
            table (D-14.4.Q2(a)); this section adds the HEADER-ROW override and
            nothing else. */}
        {/* `min="0"` AND `step="0.001"`, AND BOTH COME FROM THE SAME PLACE: THE
            SMALLEST VALUE AND THE FINEST GRANULARITY THE ENGINE ARM ACCEPTS.
            Every other number in this panel advertises the smallest value its
            arm accepts, and for a border width that value is ZERO —
            `parse_bands.go` accepts it as the thinnest device line PDF can draw
            and refuses only a NEGATIVE one, so a control that refused zero
            locally would refuse a border the format defines.

            ⚠ THE STEP IS `0.001` AND NOT `0.5` FOR THE IDENTICAL REASON. The arm
            reads this value through `propertyLength`, which accepts THREE DECIMAL
            PLACES, so `0.3pt` and `0.125pt` are legal border widths. A `step` of
            `0.5` advertised a granularity the engine does not enforce, and it
            advertised it toothlessly: `stepMismatch` is not `badInput`, so the
            blur handler committed the off-step value anyway. The control was
            drawing a rule that neither it nor the engine applied — the same
            defect as a `min` the arm does not have, in the other axis. */}
        {/* THE COMMITTED VALUE IS PASSED AS THE STRING IT IS, NOT THROUGH
            `Number()`. The projection spells this one length as a STRING so that
            `''` (absent) and `'0'` (a declared zero — a legal border, the
            thinnest line PDF can draw) are different values, and `Number()` here
            would fold them back together and reintroduce the very conflation the
            re-spelling removed. `styleNumber` branches on the string; see
            `authoredBox`. The RESOLVED half is a different question and `Number()`
            is right there — absence on that half is `''`, which `borderPainted`
            has already answered before this expression is reached. */}
        {styleNumber('border.width', 'Header border width (pt)', table['headerBorder.width'], borderPainted ? `${authored(Number(table['headerBorder.widthResolved']))}pt` : '', '0.001', '0', 'Using: nothing — no border is painted')}
        {styleColour('border.color', 'Header border colour', table['headerBorder.color'], borderPainted ? table['headerBorder.colorResolved'] : '', 'Using: nothing — no border is painted')}
        {/* ⚠ FOUR BARE CHECKBOXES, WITH NO `role="group"` AND NO `×` CLEAR, AND
            BOTH ABSENCES ARE DELIBERATE. The obvious move is to copy the
            inspector's `BorderEdgesProperty`, and copying it verbatim breaks two
            guards at once: its `<div role="group" aria-label="Border edges">`
            takes the shrunk sweep's group count 32 -> 33 and CLEARS
            `GROUP_INSTANCE_FLOOR`, which is the arm this section refused for its
            headings, arriving through the back door; and its
            `.property-inline-action ×` is a glyph button outside any segmented
            control, so it moves `V2_CENSUS`. Four checkboxes with individual
            accessible names cost nothing on either guard — the sweep's
            population is `button, [role="button"]`, so a checkbox is never
            swept.

            AND THE CLEAR AFFORDANCE IS NOT MISSING: unchecking every edge IS
            the clear. The engine refuses an empty edge array, so an emptied set
            can only mean "remove the attribute", and that is what is sent. */}
        <div className="table-header-field">Header border edges
          <span className="table-header-control table-header-edges">
            {BORDER_EDGES.map((edge) => <label key={edge}>
              <input type="checkbox" aria-label={`Header border ${edge} edge`} disabled={busy} checked={committedEdges.includes(edge)} onChange={() => { const next = BORDER_EDGES.filter((name) => name === edge ? !committedEdges.includes(name) : committedEdges.includes(name)); dispatchOnce(() => next.length === 0 ? onHeaderStyle('border.edges', 'clear') : onHeaderStyle('border.edges', 'set', next.join(','))) }} />
              {edge}
            </label>)}
          </span>
          <output aria-label="Resolved Header border edges">{resolvedNote(table['headerBorder.edgesResolved'], 'Using: nothing — no border is painted')}</output>
        </div>
        {/* THE TAKEOVER, IN WORDS, AND IT IS THIS STORY'S PRINCIPAL CORRECTNESS
            RISK RATHER THAN A COURTESY. The engine resolves the header's border
            BLOCK-GRANULARLY: `resolveHeaderStyle` takes `headerStyle.border`
            whole, with the table's own `style.border` only as a sibling case and
            never a field-by-field merge. So authoring ONE attribute stops the
            table's border reaching the header row entirely, and the other two
            fall to the format's defaults rather than to the table's values.
            Three resolved numbers tell an author what is drawn; they do not say
            THAT, and it is not visible from the field's name.

            ⚠ A `<p>` AND NOT AN `<output>`, WHICH IS A CONSTRAINT RATHER THAN A
            preference. Every `<output>` in this section is a one-line read-out
            with `text-overflow: ellipsis; white-space: nowrap`, and a sentence
            needs to wrap — but `white-space: normal` is in
            `canvas-authority-contract.test.ts`'s prohibited vocabulary, because
            the ENGINE owns text wrapping and a CSS file re-deciding it is the
            defect that scan exists for. `.honest-note` already spans the grid
            and wraps, so the sentence takes the element that fits it. It
            therefore carries NO `aria-label`: one on a bare `<p>` is dropped by
            the accessibility tree, and this panel has been burned by that twice
            already — the text itself is what a reader and a test both read. */}
        <p className="honest-note">{borderAuthored
          ? 'This header border is authored as a whole: it no longer follows the table’s border, and anything left blank above falls to the format’s own default rather than to the table’s value. Clear all three to give the table’s border back.'
          : 'No header border attribute is authored here. The note under each control is the engine’s resolved answer for that attribute. Setting any one of the three authors the header’s border as a whole, and from then on the other two fall to the format’s own default.'}</p>
        <p className="honest-note">A field left blank falls back to the table's own style and then to the format's default — except the three border attributes, which the engine takes as one block, as the line above says. The engine resolves every one of them; the note under each control is the engine's answer, not this panel's.</p>
      </div>
      {error && <p role="alert" className="file-message">{error}</p>}
      {/* THE FOOTER BAR: THE SUMMARY, AND THE TWO WAYS OUT (Story 14.7b).
          `Close Table Editor` moved down out of the heading at Story 14.7 and
          has now become a PAIR, which is D-14.7.1's ruling.

          WHAT THE TWO WORDS MEAN, because they are not symmetric and a reader
          should not have to find that out by pressing one:
            • `Done` closes and KEEPS every edit. So does ESCAPE — Escape and
              Done are THE SAME ACT here, deliberately. A modal whose Escape is
              not its Cancel is unusual, and it is forced: Cancel is disabled
              above the engine's history bound, and if Escape meant Cancel the
              dialog would stop being keyboard-dismissible in exactly that state.
            • `Cancel` is THE ONLY DISCARD. It undoes precisely the commits this
              session made that the ENGINE agreed changed the document, then
              closes. Nothing here counts anything: `editCount` is the
              application's integer (App.tsx), and this dialog neither reads nor
              could read a revision.

          ⚠ AND `Done` IS DISABLED WHILE THE DISCARD IS UNWINDING — `discarding`,
          which is the compensating sequence's own in-flight flag and NOT the
          general `busy`. Closing mid-sequence advances `tableEditorSession`, so
          App's loop returns at its teardown guard before installing the snapshot
          it reached, leaving the engine k undos behind a canvas and a preview
          that still show the pre-Cancel document. Escape is gated on the same
          flag, for the same reason, in `trapDialog` above. `busy` deliberately
          does NOT gate either one: it is not cleared by
          `setCurrentSnapshot`'s `clearDocumentInteraction` branch, and a latched
          `busy` would make this modal undismissable.

          ⚠ `Cancel` ALSO DISABLES ON `fileBusy`, mirroring the refusal
          `cancelTableEditor` already makes: a save or an export in flight makes
          the handler return, so without this the button looked available during
          one and swallowed the click. The toolbar's Undo/Redo pair
          (`disabled={!undoAvailable || fileBusy}`) is the precedent, and it is
          mirrored in BOTH directions exactly as the history bound is.

          ⚠ THE ACTIONS WRAPPER IS A PLAIN `<div>` AND MUST STAY ONE. The bar is
          `justify-content: space-between`, so the summary and the actions have
          to be its two children or four items would scatter across it. It is
          NOT `role="group"`: `control-vocabulary-contract.test.tsx:885-889`
          pins the exact string 'R0 the sweep visited 32 group instances, under
          the floor of 33', and one more group instance takes that shrunk count
          to 33, clears the floor and reds the `toEqual`. The fix is a plain
          div, never an edit to that pinned string.

          ⚠ AND THIS RE-ORDERS `trapDialog`'S FOCUSABLE LIST A THIRD TIME —
          D-12.3.2 did it once, Story 14.7 again, this is the third — INTENDED
          each time. `Done` is now the list's last member, and because the query
          is `button:not([disabled])`, A DISABLED `Cancel` DROPS OUT OF THE LIST
          ALTOGETHER and the wrap ends move again. `App.test.tsx` re-derives both
          ends from the DOM rather than naming them. */}
      <div className="table-editor-footer"><output aria-label="Column summary" aria-live="off">{`${plural(columns.length, 'column')} · ${plural(aggregateCount, 'aggregate')}`}</output><div className="table-editor-actions">{overHistoryBound && <p className="table-editor-cancel-note" id="table-editor-cancel-note">{`Cancel is unavailable: ${editCount} edits exceed the engine's ${MAX_ENGINE_HISTORY_ENTRIES}-step history, so a discard would land part-way. Use Done and undo what you want by hand.`}</p>}<button type="button" className="file-button" disabled={busy || fileBusy || overHistoryBound} aria-describedby={overHistoryBound ? 'table-editor-cancel-note' : undefined} onClick={onCancel}>Cancel</button><button type="button" className="file-button" disabled={discarding} onClick={onClose}>Done</button></div></div>
    </div>
  </section>
}
