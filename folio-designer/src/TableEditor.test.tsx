import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import { TableEditor } from './TableEditor'
import { MAX_ENGINE_HISTORY_ENTRIES } from './engine-protocol'
import type { EngineClient } from './engine-client'
import type { FileAccess } from './file/file-access'
import { acceptSampleData } from './sample-data'
import { alignGlyphs, alignSegments } from './segmented-control'

// STORY 14.7 — THE TABLE EDITOR'S OWN TEST FILE, AND WHY IT EXISTS AT ALL.
//
// `DW-351` records that the matrix's keyboard lattice has never been driven
// over more than ONE ROW. Every table fixture in this repository carries a
// single column, so `moveFocus`'s vertical loop cannot step, and its
// `enabled()` skip — the branch that walks past a disabled or ABSENT cell —
// is never exercised. "The navigation survived the rebuild" measured against
// that is a comparison against nothing.
//
// So this file was written and watched to pass BEFORE the six-column rebuild,
// against the eleven-column matrix, over a THREE-column table whose footers
// differ from one another. Every proof here is about the lattice itself:
// vertical movement, the disabled skip in both axes, Home/End over an
// end-disabled row, and — after the rebuild — the revealed cells and the
// focus that must survive one of them disappearing.

const canvas = {
  width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, locale: 'en' as const, utcOffset: '+07:00',
  marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000,
  commandWidth: 595276, commandHeight: 841890,
  fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [{ face: 'Noto Sans', assetKey: '', family: '', style: '', bold: '', italic: '', boldItalic: '' }] }],
  defaultFontSize: 12000, defaultLineSpacing: 1000,
  contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true,
  bands: [
    { name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 },
    { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 },
    { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 },
  ],
  components: [{ id: 'e7', type: 'table' as const, band: 'content' as const, x: 23276, y: 0, width: 300000, height: 12000, resizable: false }],
}
const tableHeaderProjection = { headerHeight: 12000, altRowBackground: '', headerFontFamily: '', headerFontFamilyResolved: 'body', headerFontSize: 0, headerFontSizeResolved: 12000, headerLineSpacing: 0, headerLineSpacingResolved: 1000, headerBackground: '', headerBackgroundResolved: '', headerColor: '', headerColorResolved: '', headerValign: '', headerValignResolved: 'top', headerAlign: '', headerAlignResolved: 'left', headerBold: false, headerBoldResolved: false, headerItalic: false, headerItalicResolved: false }

type Footer = '' | 'sum' | 'avg' | 'count'
type ColumnFixture = Readonly<{ id: string; header: string; width: number; align: 'left' | 'center' | 'right'; rowField: string; footer: Footer; footerOf: string; footerFormat: string }>

// THREE COLUMNS, WITH THREE DIFFERENT FOOTER SHAPES, and each difference is
// load-bearing rather than decoration:
//   • column 1 aggregates a SUM, so it carries both a source and a format;
//   • column 2 aggregates NOTHING, so it carries neither — the row with the
//     holes the lattice has to walk past;
//   • column 3 aggregates a COUNT, which takes no source but does take a
//     format — the asymmetric middle case.
// Widths sum to 174pt (72 + 60 + 42) against a band remainder of 500pt: the
// content band is 523276 millipoints wide, the table sits at x = 23276, and
// `band.width − table.x` is 500000 millipoints — the same arithmetic
// `containComponent` does. So the budget's three states are reachable by moving
// one number, which is what the three budget tests below do.
const defaultColumns: ReadonlyArray<ColumnFixture> = [
  { id: 'c1', header: 'Amount', width: 72000, align: 'right', rowField: 'amount', footer: 'sum', footerOf: 'transactions.amount', footerFormat: '#,##0.00' },
  { id: 'c2', header: 'Date', width: 60000, align: 'left', rowField: 'date', footer: '', footerOf: '', footerFormat: '' },
  { id: 'c3', header: 'Note', width: 42000, align: 'center', rowField: 'note', footer: 'count', footerOf: '', footerFormat: '0' },
]

const projected = (columns: ReadonlyArray<ColumnFixture>, alias = 'row') => columns.map((column) => ({
  id: column.id, header: column.header, width: column.width, align: column.align,
  binding: column.rowField === '' ? '' : `{{${alias}.${column.rowField}}}`, rowField: column.rowField, rowFieldEditable: true,
  footer: column.footer, footerOf: column.footerOf, footerFormat: column.footerFormat,
}))

const snapshotOf = (over: Partial<typeof canvas> = {}) => ({ documentState: 'loaded' as const, revision: 1, byteLength: 3, canvas: { ...canvas, ...over } })

// The engine mock answers every `table-columns` call from a MUTABLE fixture, so
// a test can move a column's footer the way a command would and watch the panel
// re-project. `commands` records what actually went to the engine.
//
// STORY 14.7b TAUGHT IT TWO THINGS IT COULD NOT SAY BEFORE, and both were
// blocking rather than cosmetic.
//
// (1) A NO-OP. It used to answer `revision: ++state.revision` for EVERY
// command, so it could not express a legal command that changes nothing — and
// the edit count Cancel depends on is defined by exactly that distinction. THE
// ENGINE DECIDES WHAT IS A MUTATION, NOT THE UI (App.test.tsx:3955-3957), so
// this mock decides the same way Go does: it serializes its own state before
// and after `apply` and holds the revision STILL when they agree, which is
// `bytes.Equal(canonical, e.bytes)` returning before `pushUndo`.
//
// (2) A HISTORY. Cancel is a compensating sequence of `undo` operations, and a
// mock with no undo stack can only prove how many were SENT — never that the
// document came back to where it started. So it keeps the same two stacks Go
// keeps, restores canonical serialized state from them, leaves the revision
// MONOTONIC across an undo exactly as `install` does, pushes onto redo before
// restoring, and clears redo on the next committed change.
//
// The HEADER arms are applied for the same reason: `Clear Header text colour`
// on an unset field is the reachable no-op AC3 is driven through, and a mock
// that ignored header commands would have made every header edit a no-op and
// the proof vacuous.
const HEADER_STYLE_KEYS = { fontFamily: 'headerFontFamily', fontSize: 'headerFontSize', lineSpacing: 'headerLineSpacing', background: 'headerBackground', color: 'headerColor', valign: 'headerValign', align: 'headerAlign' } as const
const NUMERIC_HEADER_KEYS: ReadonlyArray<string> = ['headerFontSize', 'headerLineSpacing']

// (3) A REFUSAL. `refuseCommand` lets a test send a command that REACHES the
// engine and comes back an error — which is the only honest way to drive the
// matrix's "a rejected command is not counted" row. A dialog that declined to
// dispatch would prove nothing: a count keyed on dispatches and a count keyed on
// an observed revision change agree when nothing is sent at all. The thrown
// shape is the one `componentDiagnostic` renders — an `elementId` and a message
// — so the refusal arrives at the dialog's error surface located, exactly as
// `App.test.tsx`'s canvas-refusal test drives it.
const REFUSED_COMMAND = 'the engine refused this command'

// (4) A HELD UNDO. Story 14.7b's `Done` and Escape are live gestures WHILE the
// compensating sequence runs — the dialog is still on screen, because closing is
// the success path only — so proving they cannot tear it down needs the sequence
// to actually be in flight when they are pressed. `pauseUndoAt` parks one `undo`
// on a promise the test resolves by hand; nothing else about the mock changes.
//
// (5) A COMPONENT MOVE. It is the one committed edit reachable AFTER the dialog
// has closed and with no selection change — an arrow nudge on the still-selected
// table — which is what the discard sentence's own promise ("until your next
// committed edit") has to be measured against. Its x rides in canonical form so
// the mock treats it as a real change, exactly as Go would.
function tableEngine(initial: ReadonlyArray<ColumnFixture> = defaultColumns, over: Partial<typeof canvas> = {}, options: Readonly<{ failUndoAt?: number; pauseUndoAt?: number; refuseCommand?: (command: Readonly<Record<string, unknown>>) => boolean }> = {}) {
  const state = { columns: [...initial], collection: 'transactions[]', alias: 'row', header: { ...tableHeaderProjection }, componentX: 23276, revision: 1 }
  const history = { undo: [] as string[], redo: [] as string[] }
  const commands: string[] = []
  let undos = 0
  let releaseUndo!: () => void
  const heldUndo = new Promise<void>((resolve) => { releaseUndo = resolve })
  // THE CANONICAL FORM, and the revision is deliberately NOT in it: Go compares
  // document BYTES, and a comparison that included the revision would call every
  // command a change.
  const canonical = () => JSON.stringify({ columns: state.columns, collection: state.collection, alias: state.alias, header: state.header, componentX: state.componentX })
  const restore = (serialized: string) => { const parsed = JSON.parse(serialized) as { columns: ColumnFixture[]; collection: string; alias: string; header: typeof tableHeaderProjection; componentX: number }; state.columns = parsed.columns; state.collection = parsed.collection; state.alias = parsed.alias; state.header = parsed.header; state.componentX = parsed.componentX }
  // THE MOCK APPLIES WHAT IT IS SENT. A frozen projection cannot prove a
  // READ-BACK — the uncontrolled boxes in this panel keep whatever was typed
  // into them whether or not the document took it — so every assertion about a
  // committed value below reads the value back through a re-projection that
  // this little in-memory engine actually computed from the command.
  const apply = (command: Readonly<Record<string, unknown>>) => {
    const columnId = String(command.columnId ?? '')
    const edit = (change: (column: ColumnFixture) => ColumnFixture) => { state.columns = state.columns.map((column) => column.id === columnId ? change(column) : column) }
    // The nudge's command, in the mock's own canonical form. `x` arrives in
    // POINTS (the command layer divides millipoints by 1000), and it is a change
    // like any other: history entry, revision, the lot.
    if (command.kind === 'moveComponent') state.componentX = Number(command.x) * 1000
    if (command.kind === 'configureTableBinding') { state.collection = String(command.collection); state.alias = String(command.alias) === '' ? 'row' : String(command.alias) }
    if (command.kind === 'updateTableColumnFooter') edit((column) => ({ ...column, footer: command.footer as Footer, footerOf: String(command.footerOf), footerFormat: String(command.footerFormat) }))
    if (command.kind === 'updateTableColumnBinding') edit((column) => ({ ...column, rowField: String(command.field) }))
    if (command.kind === 'updateTableColumn' && command.field === 'align') edit((column) => ({ ...column, align: command.value as ColumnFixture['align'] }))
    if (command.kind === 'updateTableColumn' && command.field === 'header') edit((column) => ({ ...column, header: String(command.value) }))
    if (command.kind === 'updateTableColumn' && command.field === 'width') edit((column) => ({ ...column, width: Number(command.value) * 1000 }))
    if (command.kind === 'removeTableColumn') state.columns = state.columns.filter((column) => column.id !== columnId)
    if (command.kind === 'addTableColumn') state.columns = [...state.columns.slice(0, Number(command.index)), { id: `n${state.columns.length + 1}`, header: '', width: 72000, align: 'left', rowField: '', footer: '', footerOf: '', footerFormat: '' }, ...state.columns.slice(Number(command.index))]
    if (command.kind === 'moveTableColumn') { const moving = state.columns.find((column) => column.id === columnId); if (moving) { const rest = state.columns.filter((column) => column.id !== columnId); state.columns = [...rest.slice(0, Number(command.toIndex)), moving, ...rest.slice(Number(command.toIndex))] } }
    if (command.kind === 'setTableHeaderHeight') state.header = { ...state.header, headerHeight: Number(command.height) * 1000 }
    if (command.kind === 'setTableAltRowBackground') state.header = { ...state.header, altRowBackground: command.op === 'clear' ? '' : String(command.value) }
    if (command.kind === 'updateTableHeaderStyle') {
      const key = HEADER_STYLE_KEYS[String(command.field) as keyof typeof HEADER_STYLE_KEYS]
      // A clear REMOVES the key, which the projection reports as the field's
      // empty value — '' for a string, 0 for a length. Clearing what is already
      // empty therefore leaves canonical form untouched, which is the whole of
      // the no-op arm.
      if (key !== undefined) state.header = { ...state.header, [key]: command.op === 'clear' ? (NUMERIC_HEADER_KEYS.includes(key) ? 0 : '') : (NUMERIC_HEADER_KEYS.includes(key) ? Number(command.value) * 1000 : String(command.value)) }
    }
  }
  const snap = () => ({ ...snapshotOf(over), revision: state.revision, canUndo: history.undo.length > 0, canRedo: history.redo.length > 0 })
  const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
    if (operation === 'command') {
      const text = new TextDecoder().decode(payload); commands.push(text)
      const parsed = JSON.parse(text) as Readonly<Record<string, unknown>>
      // A REFUSED COMMAND IS RECORDED IN `commands` — it did reach the engine —
      // and then rejects before `apply`. So no state moves, no history entry is
      // pushed and the revision stands still: `install` is never reached, and
      // there is nothing for Cancel to unwind.
      if (options.refuseCommand?.(parsed) === true) throw Object.assign(new Error(REFUSED_COMMAND), { elementId: 'e7' })
      const before = canonical()
      apply(parsed)
      // NO CHANGE, NO REVISION AND NO HISTORY ENTRY — Go returns
      // `e.Snapshot(), nil` before pushUndo, before `e.redo = nil` and before
      // `install`, which is the sole site of `e.revision++`.
      if (canonical() !== before) { history.undo.push(before); history.redo = []; state.revision++ }
      return { snapshot: snap() }
    }
    if (operation === 'undo') {
      undos++
      if (options.pauseUndoAt === undos) await heldUndo
      if (options.failUndoAt === undos || history.undo.length === 0) throw Object.assign(new Error('Nothing to undo'), { code: 'UNDO_UNAVAILABLE' })
      history.redo.push(canonical())
      restore(history.undo.pop() as string)
      state.revision++
      return { snapshot: snap() }
    }
    if (operation === 'redo') {
      if (history.redo.length === 0) throw Object.assign(new Error('Nothing to redo'), { code: 'REDO_UNAVAILABLE' })
      history.undo.push(canonical())
      restore(history.redo.pop() as string)
      state.revision++
      return { snapshot: snap() }
    }
    if (operation === 'table-columns') return { snapshot: snap(), tableColumns: { revision: state.revision, table: { tableId: 'e7', collection: state.collection, alias: state.alias, ...state.header, columns: projected(state.columns, state.alias) } } }
    return { snapshot: snap() }
  })
  return { state, commands, request, canonical, releaseUndo, engine: { request } as unknown as EngineClient, snapshot: snapshotOf(over) }
}

const openEditor = async (harness: ReturnType<typeof tableEngine>, sampleJson?: string) => {
  const sample = sampleJson === undefined ? undefined : acceptSampleData('c.json', new TextEncoder().encode(sampleJson).buffer)
  render(<App engine={harness.engine} initialSnapshot={harness.snapshot} initialSampleData={sample} />)
  fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
  fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
  return screen.findByRole('dialog', { name: 'Table Editor' })
}

const cellOf = (element: Element | null) => (element as HTMLElement | null)?.dataset.matrixCell
const activeCell = () => cellOf(document.activeElement)
const press = (key: string) => fireEvent.keyDown(document.activeElement!, { key })

describe('the table editor matrix lattice', () => {
  it('steps between rows with ArrowUp and ArrowDown, which no single-column fixture can exercise', async () => {
    await openEditor(tableEngine())
    const first = screen.getByRole('textbox', { name: 'Header for column 1' })
    first.focus()
    expect(activeCell()).toBe(cellOf(first))
    press('ArrowDown')
    expect(document.activeElement).toBe(screen.getByRole('textbox', { name: 'Header for column 2' }))
    press('ArrowDown')
    expect(document.activeElement).toBe(screen.getByRole('textbox', { name: 'Header for column 3' }))
    // The lattice does not wrap: the last row's ArrowDown is a no-op, and the
    // first row's ArrowUp is too.
    press('ArrowDown')
    expect(document.activeElement).toBe(screen.getByRole('textbox', { name: 'Header for column 3' }))
    press('ArrowUp')
    press('ArrowUp')
    expect(document.activeElement).toBe(first)
    press('ArrowUp')
    expect(document.activeElement).toBe(first)
  })

  it('skips a cell it may not land on vertically rather than stopping at it', async () => {
    // Column 2 aggregates nothing, so its footer format is not a place focus
    // may land — disabled today, absent after the rebuild. Either way the
    // vertical scan must walk PAST it to column 3 rather than stopping.
    await openEditor(tableEngine())
    const format = screen.getByRole('textbox', { name: 'Footer format for column 1' })
    format.focus()
    press('ArrowDown')
    expect(document.activeElement).not.toBe(document.body)
    expect(document.activeElement).toBe(screen.getByRole('textbox', { name: 'Footer format for column 3' }))
  })

  it('skips a cell it may not land on horizontally, and Home and End reach the row\'s first and last enabled controls', async () => {
    await openEditor(tableEngine())
    // COUNT TAKES NO SOURCE, so column 3's source is not a landing place —
    // disabled today, absent after the rebuild. Walking LEFT from its format
    // must arrive at the aggregate rather than at the hole between them.
    const format = screen.getByRole('textbox', { name: 'Footer format for column 3' })
    format.focus()
    press('ArrowLeft')
    expect(document.activeElement).toBe(screen.getByRole('combobox', { name: 'Footer aggregate for column 3' }))
    // Row 1 cannot move earlier and row 3 cannot move later, so each end row
    // has one disabled affordance that Home/End must decline to land on.
    const firstRowHeader = screen.getByRole('textbox', { name: 'Header for column 1' })
    firstRowHeader.focus()
    press('Home')
    expect(document.activeElement).not.toBe(screen.getByRole('button', { name: 'Move column 1 earlier' }))
    expect(screen.getByRole('button', { name: 'Move column 1 earlier' })).toBeDisabled()
    expect((document.activeElement as HTMLElement).matches(':disabled')).toBe(false)
    const lastRowHeader = screen.getByRole('textbox', { name: 'Header for column 3' })
    lastRowHeader.focus()
    press('End')
    expect(document.activeElement).not.toBe(screen.getByRole('button', { name: 'Move column 3 later' }))
    expect(screen.getByRole('button', { name: 'Move column 3 later' })).toBeDisabled()
    expect((document.activeElement as HTMLElement).matches(':disabled')).toBe(false)
  })

  it('reaches every enabled control in a row by arrow key from that row\'s first cell', async () => {
    const dialog = await openEditor(tableEngine())
    // Row 1 carries the maximal shape — an aggregate with BOTH a source and a
    // format — so its walk visits the widest lattice the panel can draw.
    const header = screen.getByRole('textbox', { name: 'Header for column 1' })
    header.focus()
    press('Home')
    const visited = new Set<string>()
    for (let step = 0; step < 40; step++) {
      const cell = activeCell()
      if (cell === undefined || visited.has(cell)) break
      visited.add(cell)
      press('ArrowRight')
    }
    const enabledInRow = Array.from(dialog.querySelectorAll<HTMLElement>('[data-matrix-cell^="0:"]')).filter((cell) => !cell.matches(':disabled'))
    expect(enabledInRow.length).toBeGreaterThan(1)
    expect([...visited].sort()).toEqual(enabledInRow.map((cell) => cell.dataset.matrixCell!).sort())
  })
})

describe('the table editor scope, budget and summary', () => {
  it('states the collection, the sample item count and the band', async () => {
    // 34 items, of which the parser keeps only SAMPLE_LIMITS.items as children:
    // the count on the node is the true one and is what must be reported.
    const items = Array.from({ length: 34 }, (_, index) => `{"amount":${index}}`).join(',')
    await openEditor(tableEngine(), `{"transactions":[${items}]}`)
    expect(screen.getByRole('status', { name: 'Table scope' })).toHaveTextContent('transactions[] · 34 items in sample · band: content')
  })

  it('says the item count is unknown rather than showing a number when no sample is loaded', async () => {
    await openEditor(tableEngine())
    const scope = screen.getByRole('status', { name: 'Table scope' })
    expect(scope).toHaveTextContent('transactions[]')
    expect(scope).toHaveTextContent('item count unknown')
    expect(scope).toHaveTextContent('band: content')
    expect(scope.textContent).not.toMatch(/\d+ items/)
  })

  it('says the item count is unknown when the sampled collection carries no count', async () => {
    // A collection the parser truncated away entirely — the node exists with no
    // `count`, and absence there means unknown, never zero.
    await openEditor(tableEngine(), '{"other":[{"a":1}]}')
    const scope = screen.getByRole('status', { name: 'Table scope' })
    expect(scope).toHaveTextContent('item count unknown')
    expect(scope.textContent).not.toContain('0 items')
  })

  it('states the width budget under, exactly at, and over the space the band leaves', async () => {
    // 72 + 60 + 42 = 174pt against a band remainder of 500pt.
    await openEditor(tableEngine())
    const budget = screen.getByRole('status', { name: 'Width budget' })
    expect(budget).toHaveTextContent('174.0')
    expect(budget).toHaveTextContent('500.0')
    expect(budget).not.toHaveTextContent('exact')
    expect(budget).toHaveTextContent('326.0 to spare')
    // THE UNIT IS NOT SPELLED ON ONE FIGURE OUT OF THREE. All three are points
    // and none of them says so here; the `pt` an author reads sits beside the
    // WIDTH box they type into.
    expect(budget.textContent).not.toContain('pt')
  })

  it('carries the exact badge only when the columns fill the band remainder exactly', async () => {
    const exact: ReadonlyArray<ColumnFixture> = [
      { ...defaultColumns[0]!, width: 250000 },
      { ...defaultColumns[1]!, width: 150000 },
      { ...defaultColumns[2]!, width: 100000 },
    ]
    await openEditor(tableEngine(exact))
    const budget = screen.getByRole('status', { name: 'Width budget' })
    expect(budget).toHaveTextContent('exact')
    expect(budget).toHaveTextContent('500.0 of 500.0')
  })

  it('states the overflow in the author\'s terms when a loaded file exceeds the band', async () => {
    const over: ReadonlyArray<ColumnFixture> = [
      { ...defaultColumns[0]!, width: 400000 },
      { ...defaultColumns[1]!, width: 150000 },
      { ...defaultColumns[2]!, width: 100000 },
    ]
    await openEditor(tableEngine(over))
    const budget = screen.getByRole('status', { name: 'Width budget' })
    expect(budget).not.toHaveTextContent('exact')
    expect(budget).toHaveTextContent('150.0 over')
    expect(budget.textContent).not.toContain('pt')
  })

  it('summarises the columns and the aggregates in the footer bar', async () => {
    await openEditor(tableEngine())
    expect(screen.getByRole('status', { name: 'Column summary' })).toHaveTextContent('3 columns · 2 aggregates')
    // STORY 14.7b — THE ONE `Close Table Editor` IS NOW THE `Cancel` / `Done`
    // PAIR, and the summary is unchanged beside it.
    const footer = within(screen.getByRole('dialog', { name: 'Table Editor' }))
    expect(footer.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    expect(footer.getByRole('button', { name: 'Done' })).toBeInTheDocument()
    expect(footer.queryByRole('button', { name: 'Close Table Editor' })).toBeNull()
  })
})

// ---------------------------------------------------------------------------
// THE ABSENCE FENCES (D-14.5.1).
//
// ⚠ THE FENCES READ THE DOM DIRECTLY RATHER THAN THROUGH ROLE QUERIES, and
// both halves of that are deliberate. `within(x)` cannot see a violation ON
// `x`, and Testing Library's role queries skip `aria-hidden` subtrees by
// default — so a fence written as `within(grid).queryAllByRole('columnheader')`
// is blind to the container itself and to anything hidden inside it. Each fence
// below is therefore proved by ITS OWN REDS: every forbidden thing is planted
// back at every position it could occupy — on the scanned root, as a direct
// child, inside a row, inside a cell, and inside an `aria-hidden` subtree — and
// the fence is asserted to red at all five.
// ---------------------------------------------------------------------------

const namesOf = (root: Element, selector: string): ReadonlyArray<string> =>
  [...(root.matches(selector) ? [root] : []), ...Array.from(root.querySelectorAll(selector))]
    .map((node) => (node.getAttribute('aria-label') ?? node.textContent ?? '').replace(/\s+/g, ' ').trim())

const columnHeaderNames = (root: Element) => namesOf(root, '[role="columnheader"]')
const RETIRED_COLUMN_HEADERS = ['Move earlier', 'Move later', 'Remove', 'Add after']
const retiredColumnHeaders = (root: Element) => columnHeaderNames(root).filter((name) => RETIRED_COLUMN_HEADERS.includes(name))

// Every position a planted violation could occupy inside a grid, as five roots
// the fence is run over one at a time.
const plantedEverywhere = (offender: string, markup: (id: string) => string): ReadonlyArray<readonly [string, Element]> => {
  const build = (inner: string) => { const host = document.createElement('div'); host.innerHTML = inner; return host.firstElementChild! }
  return [
    [`${offender} · on the scanned root itself`, build(markup('root'))],
    [`${offender} · as a direct child of the grid`, build(`<div role="grid">${markup('child')}</div>`)],
    [`${offender} · inside a row`, build(`<div role="grid"><div role="row">${markup('row')}</div></div>`)],
    [`${offender} · inside a cell`, build(`<div role="grid"><div role="row"><span role="gridcell">${markup('cell')}</span></div></div>`)],
    [`${offender} · inside an aria-hidden subtree`, build(`<div role="grid"><div aria-hidden="true">${markup('hidden')}</div></div>`)],
  ]
}

describe('the six columns the design draws, and the four that left', () => {
  it('carries exactly six columnheaders, spelled as the design spells them', async () => {
    await openEditor(tableEngine())
    const grid = screen.getByRole('grid', { name: 'Table columns' })
    expect(columnHeaderNames(grid)).toEqual(['#', 'HEADER LABEL', 'BOUND FIELD · row scope', 'WIDTH', 'ALIGN', 'FOOTER AGGREGATE'])
    expect(retiredColumnHeaders(grid)).toEqual([])
    expect(grid).toHaveAttribute('aria-colcount', '6')
  })

  it('reds when any of the four retired columnheaders is put back, at every position it could occupy', () => {
    for (const offender of RETIRED_COLUMN_HEADERS) {
      for (const [where, root] of plantedEverywhere(offender, () => `<span role="columnheader">${offender}</span>`)) {
        expect(retiredColumnHeaders(root), where).toEqual([offender])
      }
    }
  })

  it('keeps reorder and remove as named, keyboard-operable row affordances', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    expect(screen.getByRole('button', { name: 'Move column 1 earlier' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Move column 1 earlier' })).toHaveAttribute('title', 'Column 1 is already first')
    expect(screen.getByRole('button', { name: 'Move column 3 later' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Move column 3 later' })).toHaveAttribute('title', 'Column 3 is already last')
    const later = screen.getByRole('button', { name: 'Move column 1 later' })
    later.focus()
    fireEvent.keyDown(later, { key: ' ' })
    fireEvent.click(later)
    await waitFor(() => expect(harness.state.columns.map((column) => column.id)).toEqual(['c2', 'c1', 'c3']))
    await waitFor(() => expect(screen.getByRole('textbox', { name: 'Header for column 1' })).toHaveValue('Date'))
  })

  it('appends through one Add column control below the grid, and reaches a position by reorder', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    expect(screen.queryByRole('button', { name: 'Add column after column 1' })).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Add column' }))
    await waitFor(() => expect(harness.state.columns).toHaveLength(4))
    expect(harness.state.columns[3]!.id).toBe('n4')
    await screen.findByRole('textbox', { name: 'Header for column 4' })
  })

  it('keeps the empty state, its stated reason and its own Add column control', async () => {
    await openEditor(tableEngine([]))
    expect(screen.getByText('No columns yet. Add a column to start the matrix.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add column' })).toBeEnabled()
    expect(screen.queryByRole('grid', { name: 'Table columns' })).toBeNull()
  })
})

describe('the ALIGN cell is the inspector\'s segmented control', () => {
  const alignGroups = (root: Element) => [...(root.matches('[role="group"]') ? [root] : []), ...Array.from(root.querySelectorAll('[role="group"]'))].filter((group) => (group.getAttribute('aria-label') ?? '').startsWith('Cell alignment for column '))
  const ALLOWED = ['Align left', 'Align center', 'Align right']
  const alignVocabularyViolations = (root: Element): ReadonlyArray<string> =>
    alignGroups(root).flatMap((group) => Array.from(group.querySelectorAll('button')).map((button) => button.getAttribute('aria-label') ?? ''))
      .filter((label) => !ALLOWED.some((allowed) => label.startsWith(`${allowed} for column `)))

  it('renders the shared module\'s three segments, and no fourth', async () => {
    const dialog = await openEditor(tableEngine())
    const groups = alignGroups(dialog)
    expect(groups).toHaveLength(3)
    const segments = Array.from(groups[0]!.querySelectorAll('button'))
    expect(segments.map((segment) => segment.getAttribute('aria-label'))).toEqual(['Align left for column 1', 'Align center for column 1', 'Align right for column 1'])
    expect(alignVocabularyViolations(dialog)).toEqual([])
    // LITERALLY THE INSPECTOR'S CONTROL: the same class, inside the same
    // container class, drawing the same paths out of the same module. A copy
    // that merely looked alike would have to reproduce all three.
    expect(groups[0]!.className).toBe('property-segmented')
    expect(segments.map((segment) => segment.className)).toEqual(['property-segment', 'property-segment', 'property-segment'])
    expect(segments.map((segment) => segment.querySelector('svg.segment-icon path')?.getAttribute('d'))).toEqual([alignGlyphs.left, alignGlyphs.center, alignGlyphs.right])
    expect(alignSegments.map((segment) => segment.value)).toEqual(['left', 'center', 'right'])
  })

  it('reds when a fourth segment is offered, at every position it could occupy', () => {
    // AC2 IS PROVED BY ADDING `justify` BACK, not by reverting the control. A
    // column's align is `ColumnAlignTokens`, which has three members; the
    // four-segment array belongs to the inspector's all-text selection alone.
    for (const [where, root] of plantedEverywhere('Align justify', () => '<div role="group" aria-label="Cell alignment for column 1"><button aria-label="Align justify for column 1"></button></div>')) {
      expect(alignVocabularyViolations(root), where).toEqual(['Align justify for column 1'])
    }
  })

  it('shows the committed alignment and commits a new one through the engine', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    expect(screen.getByRole('button', { name: 'Align right for column 1' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Align left for column 1' })).toHaveAttribute('aria-pressed', 'false')
    fireEvent.click(screen.getByRole('button', { name: 'Align center for column 1' }))
    await waitFor(() => expect(harness.commands).toContain('{"kind":"updateTableColumn","version":1,"id":"e7","columnId":"c1","field":"align","value":"center"}'))
    await waitFor(() => expect(screen.getByRole('button', { name: 'Align center for column 1' })).toHaveAttribute('aria-pressed', 'true'))
  })
})

describe('the collapsed FOOTER AGGREGATE and what it reveals', () => {
  const footerControlNames = (root: Element) => namesOf(root, 'input, select, textarea')
  const revealedFooterControls = (root: Element, column: number) => footerControlNames(root).filter((name) => name === `Footer source for column ${column}` || name === `Footer format for column ${column}`)
  const rowOf = (dialog: Element, index: number) => dialog.querySelectorAll('.matrix-row')[index]!

  it('offers one control reading none, and no source and no format anywhere in that row', async () => {
    const dialog = await openEditor(tableEngine())
    const aggregate = screen.getByRole('combobox', { name: 'Footer aggregate for column 2' })
    expect(aggregate).toHaveValue('')
    expect(Array.from(aggregate.querySelectorAll('option')).map((option) => option.textContent)).toEqual(['none', 'sum', 'avg', 'count'])
    expect(revealedFooterControls(rowOf(dialog, 1), 2)).toEqual([])
  })

  it('reveals a format but no source for count, and both for sum', async () => {
    const dialog = await openEditor(tableEngine())
    expect(revealedFooterControls(rowOf(dialog, 2), 3)).toEqual(['Footer format for column 3'])
    expect(revealedFooterControls(rowOf(dialog, 0), 1)).toEqual(['Footer source for column 1', 'Footer format for column 1'])
    expect(screen.getByRole('textbox', { name: 'Footer source for column 1' })).toHaveValue('transactions.amount')
  })

  it('reds when a source or a format is put back into a row that has no aggregate, at every position', () => {
    for (const control of ['Footer source for column 2', 'Footer format for column 2']) {
      for (const [where, root] of plantedEverywhere(control, () => `<input aria-label="${control}" />`)) {
        expect(revealedFooterControls(root, 2), where).toEqual([control])
      }
    }
  })

  it('keeps focus in the row when the aggregate is cleared out from under the source field', async () => {
    // THE HAZARD `focusCell` CARRIED UNTIL THIS STORY. A revealed cell that
    // disappears while it holds focus left `document.activeElement` on
    // `document.body`, because `focusCell` returned early on an ABSENT cell
    // instead of falling back the way it does for a disabled one.
    await openEditor(tableEngine())
    const source = screen.getByRole('textbox', { name: 'Footer source for column 1' })
    source.focus()
    expect(activeCell()).toBe('0:10')
    fireEvent.change(screen.getByRole('combobox', { name: 'Footer aggregate for column 1' }), { target: { value: '' } })
    await waitFor(() => expect(screen.queryByRole('textbox', { name: 'Footer source for column 1' })).toBeNull())
    expect(document.activeElement).not.toBe(document.body)
    expect(activeCell()).toMatch(/^0:/)
    expect(document.activeElement).toBe(screen.getByRole('combobox', { name: 'Footer aggregate for column 1' }))
  })

  it('commits a revealed source through the engine and reads it back', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    const source = screen.getByRole('textbox', { name: 'Footer source for column 1' })
    fireEvent.blur(source, { target: { value: 'transactions.gross' } })
    await waitFor(() => expect(harness.state.columns[0]!.footerOf).toBe('transactions.gross'))
    await waitFor(() => expect(screen.getByRole('status', { name: 'Column summary' })).toHaveTextContent('3 columns · 2 aggregates'))
  })
})

describe('the row scope stays editable here, and the document answers', () => {
  it('commits a new collection and reads the engine\'s answer back out of the scope header', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    const collection = screen.getByRole('combobox', { name: 'Root collection' })
    fireEvent.blur(collection, { target: { value: 'invoices[]' } })
    await waitFor(() => expect(harness.commands).toContain('{"kind":"configureTableBinding","version":1,"id":"e7","collection":"invoices[]","alias":""}'))
    // READ BACK THROUGH THE PROJECTION, never off the box that was typed into:
    // an uncontrolled input keeps whatever was typed whether or not the
    // document took it, so asserting its value would prove nothing.
    await waitFor(() => expect(screen.getByRole('status', { name: 'Table scope' })).toHaveTextContent('invoices[]'))
  })

  it('commits a new row alias and reads it back out of every column\'s projected binding', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    const alias = screen.getByRole('textbox', { name: 'Row alias' })
    fireEvent.blur(alias, { target: { value: 'txn' } })
    await waitFor(() => expect(harness.commands).toContain('{"kind":"configureTableBinding","version":1,"id":"e7","collection":"transactions[]","alias":"txn"}'))
    await waitFor(() => expect(screen.getByRole('status', { name: 'Binding for column 1' })).toHaveTextContent('{{txn.amount}}'))
  })

  it('names itself to the accessibility tree and says where a collection is changed', async () => {
    await openEditor(tableEngine())
    const group = screen.getByRole('group', { name: 'Table row scope' })
    expect(group).toContainElement(screen.getByRole('combobox', { name: 'Root collection' }))
    expect(group).toContainElement(screen.getByRole('textbox', { name: 'Row alias' }))
    expect(group.textContent).toContain('the only place a table’s collection and row alias can be changed')
    expect(group.textContent).not.toContain('local discovery hints only')
  })
})

// ---------------------------------------------------------------------------
// STEP-04 REVIEW FINDINGS, EACH WITH THE TEST THAT WOULD HAVE CAUGHT IT.
// ---------------------------------------------------------------------------

describe('the matrix sends no command it does not have to', () => {
  it('does not re-commit an alignment that is already committed', async () => {
    // Every other commit path in this panel suppresses an unchanged value; the
    // ALIGN cell did not, so clicking the segment already pressed sent a
    // command. THE COST IS NOT AN UNDO ENTRY — `folio-go/wasm/engine.go`
    // short-circuits on `bytes.Equal(canonical, e.bytes)` and returns BEFORE
    // `pushUndo`, before the redo branch is cleared and before the revision
    // moves — it is a wasted engine round trip and a `busy` flicker over the
    // whole dialog. The guard is a convention this file already keeps
    // everywhere else, and no test clicked a pressed segment until this one.
    const harness = tableEngine()
    await openEditor(harness)
    const pressed = screen.getByRole('button', { name: 'Align right for column 1' })
    expect(pressed).toHaveAttribute('aria-pressed', 'true')
    fireEvent.click(pressed)
    await Promise.resolve()
    expect(harness.commands).toEqual([])
    // And the neighbouring segment still commits, so the guard is a guard and
    // not a control that stopped working.
    fireEvent.click(screen.getByRole('button', { name: 'Align left for column 1' }))
    await waitFor(() => expect(harness.commands).toHaveLength(1))
  })
})

describe('focus never lands on nobody inside an open modal', () => {
  it('lands on the empty state\'s Add column when the last column is removed', async () => {
    // ⚠ AND ESCAPE IS THE REASON THIS MATTERS. `trapDialog` is bound as
    // `onKeyDownCapture` on the dialog element, so a key pressed while focus
    // sits on `document.body` never reaches it: a stranded author could not
    // close the dialog with the keyboard at all.
    const harness = tableEngine([defaultColumns[0]!])
    await openEditor(harness)
    const remove = screen.getByRole('button', { name: 'Remove column 1' })
    remove.focus()
    fireEvent.click(remove)
    await waitFor(() => expect(harness.state.columns).toHaveLength(0))
    await screen.findByText('No columns yet. Add a column to start the matrix.')
    expect(document.activeElement).not.toBe(document.body)
    expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Add column' }))
    // The trap can hear a key again, which is the property the strand removed.
    fireEvent.keyDown(document.activeElement!, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).not.toBeInTheDocument())
  })

  it('keeps focus in the same row when a row is removed out from under it', async () => {
    // The case `nearestInRow` was rewritten for and the one the sibling test
    // does not reach: the surviving rows RENUMBER, so the row the author was on
    // is now a different column, and the cell they were on may not exist there.
    const harness = tableEngine()
    await openEditor(harness)
    const middleHeader = screen.getByRole('textbox', { name: 'Header for column 2' })
    middleHeader.focus()
    expect(activeCell()).toBe('1:3')
    fireEvent.click(screen.getByRole('button', { name: 'Remove column 2' }))
    await waitFor(() => expect(harness.state.columns.map((column) => column.id)).toEqual(['c1', 'c3']))
    await waitFor(() => expect(screen.queryByRole('textbox', { name: 'Header for column 3' })).toBeNull())
    expect(document.activeElement).not.toBe(document.body)
    // Row 1 is now the old column 3, and focus stayed on row 1 rather than
    // being thrown to the top of the matrix.
    expect(activeCell()).toMatch(/^1:/)
    expect(screen.getByRole('textbox', { name: 'Header for column 2' })).toHaveValue('Note')
  })

  it('does not reclaim focus that has already left the dialog', async () => {
    // The dialog's own `onFocusCapture` fires only for targets INSIDE the
    // dialog, so focus moving to something behind the modal left the panel
    // still believing the matrix held it — and the next re-projection dragged
    // the author back into the matrix from outside. The document-level
    // `focusin` listener is the only place that transition is observable.
    await openEditor(tableEngine())
    const cell = screen.getByRole('textbox', { name: 'Header for column 1' })
    cell.focus()
    const outside = screen.getByRole('button', { name: 'PREVIEW' })
    expect(screen.getByRole('dialog', { name: 'Table Editor' })).not.toContainElement(outside)
    outside.focus()
    outside.blur()
    expect(document.activeElement).toBe(document.body)
    fireEvent.change(screen.getByRole('combobox', { name: 'Footer aggregate for column 1' }), { target: { value: 'avg' } })
    await waitFor(() => expect(screen.getByRole('combobox', { name: 'Footer aggregate for column 1' })).toHaveValue('avg'))
    expect(document.activeElement).toBe(document.body)
  })
})

describe('the width budget states what it can and cannot know', () => {
  it('agrees with itself when the difference is smaller than the figure it prints', async () => {
    // 500.04pt against 500.0pt available: the printed numbers round to the same
    // figure while the stored values differ, so the line used to read
    // `Σ 500.0 of 500.0 available · 0.0 pt over` with NO exact badge — the
    // numbers saying exact and the badge saying not.
    const nearly: ReadonlyArray<ColumnFixture> = [
      { ...defaultColumns[0]!, width: 250040 },
      { ...defaultColumns[1]!, width: 150000 },
      { ...defaultColumns[2]!, width: 100000 },
    ]
    await openEditor(tableEngine(nearly))
    const budget = screen.getByRole('status', { name: 'Width budget' })
    expect(budget).toHaveTextContent('Σ 500.0 of 500.0 available')
    expect(budget).not.toHaveTextContent('exact')
    expect(budget).toHaveTextContent('under 0.1 over')
    expect(budget.textContent).not.toContain('0.0 over')
  })

  it('states that a table starting outside its band has no width that fits', async () => {
    // `band.width − table.x` goes NEGATIVE whenever the table's x sits past its
    // band's right edge — a hand-edited file, or a margin change that shrank the
    // band under an x the document already held. The read-out used to print
    // `of -50.0 available`, which is a negative quantity of space.
    const outside = { components: [{ ...canvas.components[0]!, x: 573276 }] }
    await openEditor(tableEngine(defaultColumns, outside))
    const budget = screen.getByRole('status', { name: 'Width budget' })
    expect(budget.textContent).not.toContain('-')
    expect(budget).toHaveTextContent('this table starts 50.0 outside its band, so no column width fits')
    expect(budget).not.toHaveTextContent('exact')
  })

  it('states the budget in the empty state, which is when it is most worth reading', async () => {
    // The foot carried both `Add column` and the budget and was gated on
    // `columns.length > 0`, so an author deciding how many columns will fit was
    // told nothing at exactly that moment.
    await openEditor(tableEngine([]))
    expect(screen.getByRole('status', { name: 'Width budget' })).toHaveTextContent('Σ 0.0 of 500.0 available · 500.0 to spare')
    // And still exactly ONE control named `Add column`: the empty state's own.
    expect(screen.getAllByRole('button', { name: 'Add column' })).toHaveLength(1)
  })
})

describe('the roving lattice addresses every control exactly once', () => {
  it('gives no two controls the same lattice address, and seats the aggregate after the last align segment', async () => {
    // `querySelector` resolves a duplicated `data-matrix-cell` by silently
    // picking one, so a lattice with two cells at one address has no failure of
    // its own to report. The align control owns a RUN of cells derived from
    // `alignSegments.length`; widening it used to have written the fourth
    // segment straight into the aggregate's slot.
    const dialog = await openEditor(tableEngine())
    const addresses = Array.from(dialog.querySelectorAll<HTMLElement>('[data-matrix-cell]')).map((cell) => cell.dataset.matrixCell!)
    expect(new Set(addresses).size).toBe(addresses.length)
    const columnOf = (element: Element) => Number((element as HTMLElement).dataset.matrixCell!.split(':')[1])
    const segments = Array.from(dialog.querySelectorAll('[aria-label^="Align "][aria-label$="for column 1"]'))
    expect(segments).toHaveLength(alignSegments.length)
    const segmentCells = segments.map(columnOf)
    expect(segmentCells).toEqual(Array.from({ length: alignSegments.length }, (_, offset) => segmentCells[0]! + offset))
    const aggregate = screen.getByRole('combobox', { name: 'Footer aggregate for column 1' })
    expect(columnOf(aggregate)).toBe(segmentCells[0]! + alignSegments.length)
  })
})

// ---------------------------------------------------------------------------
// STORY 14.7b — THE FOOTER'S `Cancel` / `Done` PAIR, AND THE COUNT BEHIND IT.
//
// Every arm here is driven through the SHIPPED GESTURE rather than through a
// prop: a changing edit is a real blur on a real matrix cell, and the no-op is a
// real click on the `×` beside `Header text colour`, which is `disabled={busy}`
// only and is never disabled on "already unset". The two boundary cases are the
// exception and say why in place.
// ---------------------------------------------------------------------------

// `Cancel` is `disabled={busy || overHistoryBound}`, so with the count under the
// bound its enabled state IS the dialog's idle state. Awaiting it is awaiting
// both the command and the re-projection behind it.
const idle = async () => { await waitFor(() => expect(screen.getByRole('button', { name: 'Cancel' })).toBeEnabled()) }

// A CHANGING EDIT, through the matrix cell an author types in.
const retitle = async (column: number, value: string) => {
  const box = screen.getByRole('textbox', { name: `Header for column ${column}` })
  fireEvent.change(box, { target: { value } })
  fireEvent.blur(box)
  await idle()
}

// THE REACHABLE NO-OP. `Clear Header text colour` on a field the projection
// reports as '' sends a LEGAL `clear` that removes a key which is not there, so
// canonical bytes do not move and the engine's revision stands still.
const clearHeaderColour = async () => {
  fireEvent.click(screen.getByRole('button', { name: 'Clear Header text colour' }))
  await idle()
}

const operations = (harness: ReturnType<typeof tableEngine>, operation: string) => harness.request.mock.calls.filter(([sent]) => sent === operation)
const reopen = async () => {
  fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
  return screen.findByRole('dialog', { name: 'Table Editor' })
}

// ⚠ THIS WHOLE BLOCK CARRIES A RAISED TIMEOUT, AND THE REASON IS A MEASUREMENT,
// NOT A HUNCH. Every claim here drives the real dialog through several committed
// edits — each one a type, a blur and an engine round trip — so the tests are
// legitimately slow rather than accidentally slow. Locally the heaviest are
// ~1.4s, ~1.3s, ~1.2s, ~1.2s and ~1.0s against a 5s default, which reads as
// comfortable and is not: CI's runner is 3-5x slower on jsdom with `userEvent`
// (measured — simple tests that cost ~100ms here cost 300-900ms there), which
// puts five of them between 4s and 6s. Two of them TIMED OUT on CI at `482ea5d`
// while passing locally, and the rest were one scheduling hiccup behind.
//
// The block timeout is deliberately per-BLOCK rather than sprinkled over the two
// that happened to fail first. Fixing only those would have left four tests
// sitting just inside the limit — a suite-level instance of the same shape
// D-14.7.4 names, a guard that can only just pass, where the next slightly
// slower runner reds a claim that is true. The margin is wide for the reason the
// 101-edit test below states in its own comment: **a slow machine should red the
// claim, not the clock.**
describe('the table editor\'s Cancel discards what it counted', { timeout: 30_000 }, () => {
  it('counts only the edits the engine agreed changed the document, and unwinds exactly those', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    // ONE EDIT IN AN EARLIER SESSION, KEPT WITH `Done`. It is what makes the
    // control at the foot of this test mean anything: without history from
    // BEFORE the dialog, an over-count would merely run out of undo rather than
    // reach back into the author's own work.
    await retitle(1, 'Committed earlier')
    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    const earlier = harness.canonical()

    await reopen()
    const atOpen = harness.canonical()
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    // TWO NO-OPS, and they are dispatched commands the engine accepted — the
    // count must not move for either.
    await clearHeaderColour()
    await clearHeaderColour()
    expect(harness.commands).toHaveLength(5)

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    // EXACTLY TWO, not four.
    expect(operations(harness, 'undo')).toHaveLength(2)
    expect(harness.canonical()).toBe(atOpen)
    expect(harness.canonical()).toBe(earlier)

    // THE CONTROL. Had the count come from DISPATCHES (5) or from accepted
    // commands (4) rather than from an observed revision change (2), the extra
    // steps would have run — and this is where they land: past the dialog-open
    // state, with the header an EARLIER session committed thrown away. That is
    // the single destructive failure guardrail 1 exists to prevent.
    await harness.request('undo')
    expect(harness.canonical()).not.toBe(atOpen)
    expect(harness.canonical()).not.toContain('Committed earlier')
  })

  it('does not count a command the engine refused, and leaves the existing error surface standing', async () => {
    // MATRIX ROW: "a rejected command is not counted". THE REFUSAL COMES FROM
    // THE ENGINE, not from the dialog declining to dispatch — that is what makes
    // the input SEPARATING. Three commands reach the engine and two of them move
    // the document, so a count keyed on dispatches (3) and a count keyed on an
    // observed revision change (2) give different answers here; a refusal the
    // dialog swallowed before sending would have them agree, and the test would
    // measure nothing. Proved the way the counted-discard test above proves it:
    // two genuinely changing edits, the refused one, then `Cancel`.
    const harness = tableEngine(defaultColumns, {}, { refuseCommand: (command) => command.value === 'Refused' })
    await openEditor(harness)
    const atOpen = harness.canonical()
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Refused')
    // ALL THREE REACHED THE ENGINE. The refused one is not a command that was
    // never sent.
    expect(harness.commands).toHaveLength(3)

    // THE EXISTING ERROR SURFACE IS UNCHANGED — the other half of this row. The
    // dialog's own `role="alert"` carries the engine's LOCATED sentence, and the
    // dialog stays open on it rather than tearing itself down.
    const failure = await screen.findByRole('alert')
    expect(failure).toHaveTextContent(`e7: ${REFUSED_COMMAND}`)
    expect(failure.className).toContain('file-message')
    expect(screen.getByRole('dialog', { name: 'Table Editor' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    // TWO, NOT THREE. The refusal never reached `install`, so it left no history
    // entry, and a third `undo` would have reached past the dialog-open state.
    expect(operations(harness, 'undo')).toHaveLength(2)
    expect(harness.canonical()).toBe(atOpen)
    // AND THE UNDO STACK IS EMPTY AT DIALOG-OPEN STATE, which is the same claim
    // read from the other side: had the count been 3, the third step would have
    // had nothing to consume and Cancel would have reported a failure instead of
    // a discard.
    await expect(harness.request('undo')).rejects.toThrow('Nothing to undo')
  })

  it('keeps every edit when the author presses Done, and sends no undo at all', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    const kept = harness.canonical()
    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(harness.canonical()).toBe(kept)
    expect(harness.canonical()).toContain('Total')
  })

  it('treats Escape as Done rather than as Cancel, so it closes and keeps', async () => {
    const harness = tableEngine()
    const dialog = await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    const kept = harness.canonical()
    fireEvent.keyDown(dialog, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(harness.canonical()).toBe(kept)
  })

  it('starts a reopened dialog at zero, because the count is a function of the session', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    const kept = harness.canonical()

    await reopen()
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(harness.canonical()).toBe(kept)
  })

  it('says nothing and sends nothing when there is nothing to discard', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(screen.queryByText(/Discarded/)).toBeNull()
  })

  it('states the discard and the honest limit on it, and the discarded edits are still redoable', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    const beforeCancel = harness.canonical()
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(3)
    expect(harness.canonical()).not.toBe(beforeCancel)

    // THE NUMBER, AND THE LIMIT ON THE PROMISE. `Undo()` pushes onto redo before
    // restoring, so the discarded edits survive — but only until the next
    // committed command clears the engine's redo stack.
    const stated = screen.getByText(/Discarded 3 table editor edits/)
    expect(stated).toHaveAttribute('role', 'status')
    expect(stated).toHaveAttribute('aria-live', 'polite')
    expect(stated.textContent).toContain('Redo restores them until your next committed edit')

    // AC7, AND IT IS N PRESSES AND AN EXACT DOCUMENT, NOT ONE PRESS AND AN
    // INEQUALITY. "Redo restores them" is a claim about ALL THREE: a single redo
    // followed by `not.toEqual` is satisfied by an engine that restored one edit,
    // or a different edit, or half of one. Three presses and byte-equality
    // against the canonical form the document held immediately before `Cancel`
    // is the sentence's actual promise.
    for (const press of [1, 2, 3]) {
      const redo = screen.getByRole('button', { name: 'Redo' })
      await waitFor(() => expect(redo).toBeEnabled())
      fireEvent.click(redo)
      await waitFor(() => expect(operations(harness, 'redo')).toHaveLength(press))
    }
    expect(harness.canonical()).toBe(beforeCancel)
  })

  it('says it in the singular when exactly one edit is discarded', async () => {
    // THE SINGULAR ARM OF `discarded === 1 ? 'edit' : 'edits'` AND OF `'it' :
    // 'them'`. Every other discard here is 3, so both words rendered only in
    // their plural form and the ternaries were untested in one direction.
    const harness = tableEngine()
    await openEditor(harness)
    const atOpen = harness.canonical()
    await retitle(1, 'Total')

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(operations(harness, 'undo')).toHaveLength(1)
    expect(harness.canonical()).toBe(atOpen)
    const stated = screen.getByText(/Discarded 1 table editor edit\./)
    expect(stated.textContent).toContain('Redo restores it until your next committed edit')
    // AND NOT THE PLURAL, in either place: a sentence that read "1 edits" or
    // "restores them" would satisfy a looser matcher.
    expect(stated.textContent).not.toContain('edits')
    expect(stated.textContent).not.toContain('them')
  })

  it('stops where it actually got to when an undo in the sequence fails, and stays open to say so', async () => {
    // The 3rd `undo` returns UNDO_UNAVAILABLE against a count of 4.
    const harness = tableEngine(defaultColumns, {}, { failUndoAt: 3 })
    await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    fireEvent.click(screen.getByRole('button', { name: 'Add column' }))
    await idle()

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    const failure = await screen.findByRole('alert')
    // THE DIALOG IS STILL OPEN. A closed dialog can state nothing, so closing is
    // the success path only.
    expect(screen.getByRole('dialog', { name: 'Table Editor' })).toBeInTheDocument()
    // BOTH NUMBERS, so a message claiming a completed discard cannot pass for
    // this one.
    expect(failure.textContent).toContain('Discarded 2 of 4 edits')
    expect(failure.textContent).toContain('The other 2 still stand')
    expect(failure.textContent).toContain('nothing left to undo')
    expect(operations(harness, 'undo')).toHaveLength(3)
    // AND ITS PROJECTION IS THE DOCUMENT IT REACHED, not the one it set out
    // from: two of the four edits were undone, so the fourth column is gone
    // again and the third column's header is back to 'Note'.
    //
    // ⚠ THE DOCUMENT IS READ FROM THE ENGINE AND THE COLUMN COUNT FROM THE DOM,
    // and the split is not a preference. The matrix's header box is UNCONTROLLED
    // and carries no `key`, so it keeps whatever was typed into it across a
    // re-projection — asserting its `value` would measure the author's typing,
    // not the document. The row COUNT is structural and the DOM does show it.
    await waitFor(() => expect(screen.queryByRole('textbox', { name: 'Header for column 4' })).toBeNull())
    expect(screen.getAllByRole('textbox', { name: /^Header for column/ })).toHaveLength(3)
    expect((JSON.parse(harness.canonical()) as { columns: ColumnFixture[] }).columns.map((column) => column.header)).toEqual(['Total', 'When', 'Note'])
  })

  // -------------------------------------------------------------------------
  // THE UNWIND IS NOT A MOMENT THE DIALOG CAN BE CLOSED IN.
  //
  // The sequence runs with the dialog STILL OPEN — closing is the success path
  // only — so `Done` and Escape are live gestures over a loop that is mid-flight.
  // Both tear the session down (`revokeTableEditor` advances
  // `tableEditorSession`), and the loop's own session guard then returns BEFORE
  // it installs the snapshot it reached: the engine ends k undos back while the
  // canvas, `snapshotRef` and the preview still show the pre-Cancel document,
  // with nothing on screen saying the two have parted company. That is what
  // `discarding` shuts, and it is `discarding` and never `busy` — `busy` is not
  // cleared by `setCurrentSnapshot`'s `clearDocumentInteraction` branch, so
  // gating a way out on it could leave a modal that cannot be closed at all.
  // -------------------------------------------------------------------------
  it.each([
    ['Done', (dialog: HTMLElement) => { const done = screen.getByRole('button', { name: 'Done' }); expect(done, 'Done must be disabled while the discard is unwinding').toBeDisabled(); fireEvent.click(done); expect(dialog).toBeInTheDocument() }],
    ['Escape', (dialog: HTMLElement) => { fireEvent.keyDown(dialog, { key: 'Escape' }) }],
  ])('does not let %s tear the discard down mid-unwind, and the engine never gets ahead of the screen', async (_name, dismiss) => {
    // THE SECOND UNDO OF THREE IS PARKED, so the loop is genuinely in flight when
    // the gesture lands: one undo has been applied to the document, two have not,
    // and the reached snapshot has not been installed anywhere yet.
    const harness = tableEngine(defaultColumns, {}, { pauseUndoAt: 2 })
    const dialog = await openEditor(harness)
    const atOpen = harness.canonical()
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    await retitle(1, 'And one more')

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(operations(harness, 'undo')).toHaveLength(2))
    dismiss(dialog)
    await act(async () => { await Promise.resolve() })
    // THE DIALOG IS STILL HERE. Whether the gesture was refused at the control or
    // swallowed at the trap, what it must not do is end the session.
    expect(screen.getByRole('dialog', { name: 'Table Editor' })).toBeInTheDocument()

    harness.releaseUndo()
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    // ALL FOUR RAN AND THE DOCUMENT IS BACK AT DIALOG-OPEN STATE.
    expect(operations(harness, 'undo')).toHaveLength(4)
    expect(harness.canonical()).toBe(atOpen)
    // ⚠ AND THE SCREEN AGREES WITH THE ENGINE, which is the defect stated as an
    // assertion rather than as a count. A teardown mid-sequence makes the loop
    // return before `setCurrentSnapshot`, so the engine sits four undos back
    // while this read-out still shows the revision the last commit installed.
    expect(screen.getByTestId('engine-snapshot')).toHaveTextContent(`REVISION ${harness.state.revision}`)
    expect(screen.getByText(/Discarded 4 table editor edits/)).toBeInTheDocument()
  })

  it.each([
    ['Done', () => { expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled(); fireEvent.click(screen.getByRole('button', { name: 'Done' })) }],
    ['Escape', (dialog: HTMLElement) => { fireEvent.keyDown(dialog, { key: 'Escape' }) }],
  ])('leaves %s working while an ordinary blur commit is in flight, and when nothing is', async (_name, dismiss) => {
    // THE CONTROL FOR THE PAIR ABOVE, and the reason the gate is `discarding` and
    // not `busy`. An ordinary commit raises `busy` too — `Cancel` is disabled
    // right here, which is how this test knows it is in flight — and the two ways
    // out must be untouched by it. A `busy`-gated Escape would also be a modal
    // with no exit the moment `busy` latched, which
    // `setCurrentSnapshot`'s `clearDocumentInteraction` branch can do.
    const harness = tableEngine()
    const dialog = await openEditor(harness)
    // IDLE FIRST, so "still works" has a baseline in this same test.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeEnabled()

    const box = screen.getByRole('textbox', { name: 'Header for column 1' })
    fireEvent.change(box, { target: { value: 'Total' } })
    fireEvent.blur(box)
    expect(screen.getByRole('button', { name: 'Cancel' }), 'the commit must actually be in flight for this test to mean anything').toBeDisabled()
    dismiss(dialog)
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    // BOTH ARE `Done`: they close and KEEP, so no undo is sent by either.
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(screen.queryByText(/Discarded/)).toBeNull()
  })

  it('withdraws the discard sentence once it stops being true', async () => {
    const harness = tableEngine()
    await openEditor(harness)
    await retitle(1, 'Total')
    await retitle(2, 'When')
    await retitle(3, 'Remark')
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Table Editor' })).toBeNull())
    expect(screen.getByText(/Discarded 3 table editor edits/)).toBeInTheDocument()

    // ⚠ THE SENTENCE PROMISES SOMETHING WITH AN EXPIRY: "Redo restores them until
    // your next committed edit". Go's next `install` sets `e.redo = nil`, so the
    // moment the document moves again the promise is FALSE — and a live region
    // still asserting it is worse than silence. An arrow nudge on the
    // still-selected table is the committed edit that ends it: it changes no
    // selection and never touches the edit count, which is why clearing the
    // sentence only where the COUNT is cleared would leave this case standing.
    fireEvent.keyDown(screen.getByLabelText('Canvas region'), { key: 'ArrowRight' })
    await waitFor(() => expect(harness.commands.filter((text) => text.includes('"kind":"moveComponent"'))).toHaveLength(1))
    await waitFor(() => expect(screen.queryByText(/Discarded 3 table editor edits/)).toBeNull())
  })

  it('carries the count the application accumulated into the footer, over the engine\'s bound', async () => {
    // ⚠ THIS IS THE ONLY EXECUTING PROOF THAT `tableEditorEditCount` — the STATE
    // MIRROR beside the ref — reaches the dialog at all. Every other arm here
    // measures undo operations, which come from the REF; the two boundary tests
    // below render the dialog with a literal. So the mirror could be dropped
    // (`setTableEditorEdits` writing only `tableEditorEdits.current`) and nothing
    // would red, while a real author at 101 edits would see an enabled `Cancel`
    // that silently does nothing, because the loop refuses above the bound.
    //
    // ONE COLUMN, and the count is driven through the same shipped blur every
    // other arm uses — 101 of them, which is what the claim needs: the footer's
    // reason cannot be reached with fewer.
    const harness = tableEngine([defaultColumns[0]!])
    await openEditor(harness)
    const box = () => screen.getByRole('textbox', { name: 'Header for column 1' })
    // NOT `idle()`, AND NOT `waitFor`. That helper waits on `Cancel` becoming
    // enabled, which is the very thing this test drives to FALSE; and a hundred
    // `waitFor` polls cost more than the hundred commits do. `commitTableColumn`
    // awaits exactly two engine round trips against a synchronous mock, so
    // draining the microtask queue inside `act` settles each edit — and if it ever
    // stopped settling, the command count asserted immediately after the loop
    // would say so rather than the assertions quietly measuring a smaller count.
    const settle = async () => { await act(async () => { for (let turn = 0; turn < 6; turn++) await Promise.resolve() }) }
    for (let edit = 1; edit <= MAX_ENGINE_HISTORY_ENTRIES + 1; edit++) {
      const target = box()
      fireEvent.change(target, { target: { value: `Header ${edit}` } })
      fireEvent.blur(target)
      await settle()
      expect(box(), `edit ${edit} never settled`).toBeEnabled()
    }
    expect(harness.commands).toHaveLength(MAX_ENGINE_HISTORY_ENTRIES + 1)

    const note = screen.getByText(/Cancel is unavailable/)
    expect(note.textContent).toContain(`${MAX_ENGINE_HISTORY_ENTRIES + 1} edits`)
    expect(note.textContent).toContain(`${MAX_ENGINE_HISTORY_ENTRIES}-step history`)
    const cancel = screen.getByRole('button', { name: 'Cancel' })
    expect(cancel).toBeDisabled()
    expect(cancel).toHaveAttribute('aria-describedby', note.id)
    // AND THE WAY OUT THAT KEEPS THE WORK IS STILL THERE, which is the whole
    // reason a disabled Cancel is safe.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    // ⚠ ITS OWN TIMEOUT, BECAUSE A HUNDRED AND ONE REAL COMMITS COST REAL TIME —
    // measured at ~6s here, against the 5s default. The alternative was to fake
    // the count, and a faked count cannot prove that the APPLICATION's mirror is
    // what the footer reads. The margin is deliberately wide so a slower machine
    // reds the claim rather than the clock.
  }, 30_000)

  it('refuses Cancel in the footer while a local file operation is in flight, not only in the handler', async () => {
    // `cancelTableEditor` returns early on `fileBusy`, so without the same flag on
    // the button the author met an available-looking `Cancel` during a save that
    // swallowed the click and discarded nothing. Mirrored in both directions, as
    // the history bound already is.
    const harness = tableEngine()
    // A save target that never arrives holds `fileBusy` up for the whole test.
    const fileAccess = { open: vi.fn(), acquireSaveTarget: vi.fn(() => new Promise<never>(() => {})), writeSave: vi.fn() } as unknown as FileAccess
    render(<App engine={harness.engine} initialSnapshot={harness.snapshot} fileAccess={fileAccess} />)
    fireEvent.click(screen.getByRole('button', { name: 'table component e7' }))
    fireEvent.click(screen.getByRole('button', { name: 'Configure columns' }))
    await screen.findByRole('dialog', { name: 'Table Editor' })
    await retitle(1, 'Total')

    // Cmd/Ctrl+S sits ABOVE the modal guard and still saves from inside the
    // dialog, which is what puts this dialog in the state being tested.
    fireEvent.keyDown(screen.getByRole('button', { name: 'Done' }), { key: 's', ctrlKey: true })
    await waitFor(() => expect(fileAccess.acquireSaveTarget).toHaveBeenCalledOnce())
    const cancel = screen.getByRole('button', { name: 'Cancel' })
    expect(cancel).toBeDisabled()
    fireEvent.click(cancel)
    await act(async () => { await Promise.resolve() })
    expect(operations(harness, 'undo')).toHaveLength(0)
    expect(screen.getByRole('dialog', { name: 'Table Editor' })).toBeInTheDocument()
    // `Done` is not a document mutation, so a save in flight does not take it
    // away: the dialog still has a way out.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
  })

  // THE TWO SIDES OF THE HISTORY BOUND, AND THIS PAIR RENDERS THE DIALOG
  // DIRECTLY RATHER THAN DRIVING App — deliberately. The claim is about what the
  // FOOTER DOES WITH A COUNT, and reaching 101 committed edits through the real
  // gestures would spend a hundred round trips to set up a prop this dialog
  // simply receives. Everything else in this describe goes through App.
  const directProjection = { revision: 1, table: { tableId: 'e7', collection: 'transactions[]', alias: 'row', ...tableHeaderProjection, columns: projected(defaultColumns) } }
  // `busy`, `fileBusy` AND `discarding` ALL DEFAULT TO FALSE, so every call site
  // keeps exactly the meaning it had and each arm below passes only the one flag
  // it is about. `unmount` is returned so an arm can hold its own control — the
  // SAME count with the flag down — in one test without two dialogs sharing
  // `screen`.
  const renderFooter = (editCount: number, busy = false, fileBusy = false, discarding = false) => {
    const onCancel = vi.fn()
    const onClose = vi.fn()
    const { unmount } = render(<TableEditor projection={directProjection} busy={busy} fileBusy={fileBusy} discarding={discarding} candidates={[]} sampleAvailable={false} editCount={editCount} onClose={onClose} onCancel={onCancel} onAdd={vi.fn()} onRemove={vi.fn()} onMove={vi.fn()} onUpdate={vi.fn()} onConfigure={vi.fn()} onBind={vi.fn()} onFooter={vi.fn()} onHeaderHeight={vi.fn()} onAltRowBackground={vi.fn()} onHeaderStyle={vi.fn()} />)
    return { onCancel, onClose, unmount }
  }

  it('keeps Cancel available at exactly the engine\'s history limit', () => {
    const { onCancel, onClose } = renderFooter(MAX_ENGINE_HISTORY_ENTRIES)
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeEnabled()
    expect(screen.queryByText(/Cancel is unavailable/)).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(onCancel).toHaveBeenCalledOnce()
    // Done and Escape are available here too, and they are the same act.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('disables Cancel one above the limit and states the reason in visible text', () => {
    const { onCancel, onClose } = renderFooter(MAX_ENGINE_HISTORY_ENTRIES + 1)
    const cancel = screen.getByRole('button', { name: 'Cancel' })
    expect(cancel).toBeDisabled()
    // A VISIBLE SENTENCE, not a bare grey-out and not a `title`. It names the
    // author's number and the engine's, and says what would go wrong.
    const note = screen.getByText(/Cancel is unavailable/)
    expect(note.textContent).toContain(`${MAX_ENGINE_HISTORY_ENTRIES + 1} edits`)
    expect(note.textContent).toContain(`${MAX_ENGINE_HISTORY_ENTRIES}-step history`)
    expect(note.textContent).toContain('land part-way')
    expect(cancel).toHaveAttribute('aria-describedby', note.id)
    // DONE AND ESCAPE STAY AVAILABLE. This is the state that forces Escape to
    // mean Done: if it meant Cancel the dialog would not be keyboard-dismissible
    // here at all.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    expect(onClose).toHaveBeenCalledOnce()
    expect(onCancel).not.toHaveBeenCalled()

    // AND THE TRAP STILL WRAPS AT BOTH ENDS WITH `Cancel` GONE FROM ITS LIST.
    // `trapDialog` selects `button:not([disabled])`, so a disabled Cancel is not
    // in it and the wrap ends move — the third intended re-ordering of this
    // list. Both ends are re-derived from the DOM.
    const dialog = screen.getByRole('dialog', { name: 'Table Editor' })
    const tabbable = Array.from(dialog.querySelectorAll<HTMLElement>('button:not([disabled]), input:not([disabled]), select:not([disabled])')).filter((element) => element.tabIndex >= 0)
    expect(tabbable).not.toContain(cancel)
    const last = tabbable[tabbable.length - 1] as HTMLElement
    expect(last).toBe(screen.getByRole('button', { name: 'Done' }))
    last.focus()
    fireEvent.keyDown(last, { key: 'Tab' })
    expect(document.activeElement).toBe(tabbable[0])
    fireEvent.keyDown(document.activeElement!, { key: 'Tab', shiftKey: true })
    expect(document.activeElement).toBe(last)
  })

  it('disables Cancel while a command is in flight, at a count nowhere near the history bound', () => {
    // MATRIX ROW: "Cancel while a command is in flight". `Cancel` is
    // `disabled={busy || overHistoryBound}`, so THE COUNT IS THE SEPARATING
    // INPUT, not the flag: a test that set `busy` alongside a count of 101 could
    // not say which of the two conditions disabled the button, and would still
    // pass against a `Cancel` that ignored `busy` entirely. TWO is far under
    // MAX_ENGINE_HISTORY_ENTRIES, so only `busy` can be doing it here — and the
    // control at the foot of this test renders the SAME count with `busy` false
    // and finds Cancel enabled.
    const inFlight = renderFooter(2, true)
    const cancel = screen.getByRole('button', { name: 'Cancel' })
    expect(cancel).toBeDisabled()
    // AND IT IS NOT THE BOUND SAYING SO. The over-bound note is the only thing
    // that ever explains a disabled Cancel by the bound, and it is absent, as is
    // the `aria-describedby` that points at it.
    expect(screen.queryByText(/Cancel is unavailable/)).toBeNull()
    expect(cancel).not.toHaveAttribute('aria-describedby')
    // THE COUNT CANNOT MOVE UNDER THE LOOP, because the loop cannot be started:
    // a disabled button dispatches nothing.
    fireEvent.click(cancel)
    expect(inFlight.onCancel).not.toHaveBeenCalled()
    // `Done` stays available, so an in-flight command does not make the dialog a
    // trap — the same reason `Done` survives the over-bound state.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()

    // THE CONTROL. Same count, nothing in flight: enabled, and it dispatches.
    // Without this arm the assertions above are satisfied by a `Cancel` that is
    // disabled for some other reason, or always.
    inFlight.unmount()
    const idleFooter = renderFooter(2)
    const enabled = screen.getByRole('button', { name: 'Cancel' })
    expect(enabled).toBeEnabled()
    fireEvent.click(enabled)
    expect(idleFooter.onCancel).toHaveBeenCalledOnce()
  })

  it('disables Cancel while a local file operation is in flight, and not by the bound', () => {
    // The third condition on this button, and the same separating shape the arm
    // above uses: TWO is far under MAX_ENGINE_HISTORY_ENTRIES and `busy` is false,
    // so only `fileBusy` can be disabling it here.
    const saving = renderFooter(2, false, true)
    const cancel = screen.getByRole('button', { name: 'Cancel' })
    expect(cancel).toBeDisabled()
    expect(screen.queryByText(/Cancel is unavailable/)).toBeNull()
    expect(cancel).not.toHaveAttribute('aria-describedby')
    fireEvent.click(cancel)
    expect(saving.onCancel).not.toHaveBeenCalled()
    // `Done` and Escape are unaffected: a save is not a reason to trap the author
    // in the dialog.
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    expect(saving.onClose).toHaveBeenCalledOnce()

    // THE CONTROL. Same count, no file operation: enabled, and it dispatches.
    saving.unmount()
    const idleFooter = renderFooter(2)
    const enabled = screen.getByRole('button', { name: 'Cancel' })
    expect(enabled).toBeEnabled()
    fireEvent.click(enabled)
    expect(idleFooter.onCancel).toHaveBeenCalledOnce()
  })

  it('shuts both ways out while the discard is unwinding, and only then', () => {
    // THE DIALOG-LEVEL HALF of the mid-unwind proof above: `discarding` is the
    // ONLY flag that takes `Done` and Escape away. The count is 2 and `busy` is
    // false in the first render, so nothing else could be doing it — and the
    // control that follows sets `busy` INSTEAD, at the same count, and finds both
    // ways out live. That pairing is the whole claim: a `busy`-gated Escape plus a
    // `busy` that can latch is a modal with no exit at all.
    const unwinding = renderFooter(2, false, false, true)
    const dialog = screen.getByRole('dialog', { name: 'Table Editor' })
    expect(screen.getByRole('button', { name: 'Done' })).toBeDisabled()
    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    fireEvent.keyDown(dialog, { key: 'Escape' })
    expect(unwinding.onClose).not.toHaveBeenCalled()

    unwinding.unmount()
    const committing = renderFooter(2, true)
    expect(screen.getByRole('button', { name: 'Done' })).toBeEnabled()
    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(committing.onClose).toHaveBeenCalledOnce()
    fireEvent.keyDown(screen.getByRole('dialog', { name: 'Table Editor' }), { key: 'Escape' })
    expect(committing.onClose).toHaveBeenCalledTimes(2)
  })
})
