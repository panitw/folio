import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import type { EngineClient } from './engine-client'
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
function tableEngine(initial: ReadonlyArray<ColumnFixture> = defaultColumns, over: Partial<typeof canvas> = {}) {
  const state = { columns: [...initial], collection: 'transactions[]', alias: 'row', revision: 1 }
  const commands: string[] = []
  // THE MOCK APPLIES WHAT IT IS SENT. A frozen projection cannot prove a
  // READ-BACK — the uncontrolled boxes in this panel keep whatever was typed
  // into them whether or not the document took it — so every assertion about a
  // committed value below reads the value back through a re-projection that
  // this little in-memory engine actually computed from the command.
  const apply = (command: Readonly<Record<string, unknown>>) => {
    const columnId = String(command.columnId ?? '')
    const edit = (change: (column: ColumnFixture) => ColumnFixture) => { state.columns = state.columns.map((column) => column.id === columnId ? change(column) : column) }
    if (command.kind === 'configureTableBinding') { state.collection = String(command.collection); state.alias = String(command.alias) === '' ? 'row' : String(command.alias) }
    if (command.kind === 'updateTableColumnFooter') edit((column) => ({ ...column, footer: command.footer as Footer, footerOf: String(command.footerOf), footerFormat: String(command.footerFormat) }))
    if (command.kind === 'updateTableColumnBinding') edit((column) => ({ ...column, rowField: String(command.field) }))
    if (command.kind === 'updateTableColumn' && command.field === 'align') edit((column) => ({ ...column, align: command.value as ColumnFixture['align'] }))
    if (command.kind === 'updateTableColumn' && command.field === 'header') edit((column) => ({ ...column, header: String(command.value) }))
    if (command.kind === 'updateTableColumn' && command.field === 'width') edit((column) => ({ ...column, width: Number(command.value) * 1000 }))
    if (command.kind === 'removeTableColumn') state.columns = state.columns.filter((column) => column.id !== columnId)
    if (command.kind === 'addTableColumn') state.columns = [...state.columns.slice(0, Number(command.index)), { id: `n${state.columns.length + 1}`, header: '', width: 72000, align: 'left', rowField: '', footer: '', footerOf: '', footerFormat: '' }, ...state.columns.slice(Number(command.index))]
    if (command.kind === 'moveTableColumn') { const moving = state.columns.find((column) => column.id === columnId); if (moving) { const rest = state.columns.filter((column) => column.id !== columnId); state.columns = [...rest.slice(0, Number(command.toIndex)), moving, ...rest.slice(Number(command.toIndex))] } }
  }
  const request = vi.fn(async (operation: string, payload?: ArrayBuffer) => {
    if (operation === 'command') { const text = new TextDecoder().decode(payload); commands.push(text); apply(JSON.parse(text) as Readonly<Record<string, unknown>>); return { snapshot: { ...snapshotOf(over), revision: ++state.revision } } }
    if (operation === 'table-columns') return { snapshot: { ...snapshotOf(over), revision: state.revision }, tableColumns: { revision: state.revision, table: { tableId: 'e7', collection: state.collection, alias: state.alias, ...tableHeaderProjection, columns: projected(state.columns, state.alias) } } }
    return { snapshot: { ...snapshotOf(over), revision: state.revision } }
  })
  return { state, commands, request, engine: { request } as unknown as EngineClient, snapshot: snapshotOf(over) }
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
    expect(within(screen.getByRole('dialog', { name: 'Table Editor' })).getByRole('button', { name: 'Close Table Editor' })).toBeInTheDocument()
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
