import type { CanvasProjection } from './engine-protocol'

// THE SHEET STACK, AS ARITHMETIC.
//
// Everything here is a pure function of numbers the ENGINE projected — the
// window origins, the window height, the page height, the band rectangles —
// plus this module's own two declared display constants and the local zoom.
// Nothing measures the DOM, and nothing may: `canvas-authority-contract.
// test.ts` bans getBoundingClientRect, every offset*/client*/scroll* metric
// and getComputedStyle across the whole designer, which is why the gap
// between two sheets is declared HERE as a number and written out as a custom
// property rather than read back from a CSS token.
//
// It is a `.ts` and not a `.tsx` on purpose: oxlint's `only-export-components`
// baseline is exactly four warnings, and a non-component export added beside
// a component would make a fifth. A pure module is also the only way to unit
// test this geometry at all.

// MAX_CANVAS_SHEETS bounds the DRAWING, never the projected value. D-7.4.2's
// settled shape: truncate the drawing, keep the number, make the degraded
// state distinguishable from the empty one, and derive the bound rather than
// picking a round one.
//
// THE DERIVATION. Epic 7's narrative target is forty pages. The projection's
// own body-text paint budget is 1920 lines (maxCanvasBodyTextLines), which at
// the decision log's corrected forty-to-fifty lines per A4 window is under
// fifty windows of solid prose — so past that, a window can only come from a
// DECLARED placement gap, not from text anyone is reading. 120 is three times
// the epic's stated target and more than twice what the paint budget can
// fill. Each sheet is a page's worth of DOM and the canvas is unvirtualised
// (DW-34), which is the cost this bounds; the count itself is untouched, and
// the interface says out loud that it is showing the first N of M.
export const MAX_CANVAS_SHEETS = 120

// The vertical gap between two sheets, in CSS pixels at zoom 1 — the stack's
// own constant, not `--space-5`. The stack must be able to invert its own
// display geometry to keep a drag tracking the pointer across a seam, and a
// token the browser would have to read back is exactly the measurement this
// canvas may not make.
export const SHEET_STACK_GAP = 24

export type CanvasComponentProjection = CanvasProjection['components'][number]

// One occurrence of a component on one sheet. `home` is true for exactly one
// occurrence per component: the window its own top falls in. Only the home
// occurrence is interactive and accessibly named — two identical accessible
// names for one component would break selection, RTL's getByLabelText and
// Playwright's strict mode alike (Ruling G).
export type SheetOccurrence = Readonly<{ component: CanvasComponentProjection; y: number; home: boolean }>

export type Sheet = Readonly<{
  // The window this sheet draws, 0-based, and where that window begins in the
  // content column — straight from contentWindowOrigins, never index * height.
  index: number
  origin: number
  // Where the NEXT window begins, measured down this sheet's content band,
  // when that falls inside the band. Absent when the next window begins past
  // this sheet's own foot — a declared gap — in which case the band's foot IS
  // the boundary and the skipped column region is drawn by nobody.
  seam?: number
  content: ReadonlyArray<SheetOccurrence>
}>

export type SheetStack = Readonly<{
  sheets: ReadonlyArray<Sheet>
  // What the projection said, before the drawing budget was applied.
  windowCount: number
  truncated: boolean
  isFloor: boolean
}>

const contentComponents = (canvas: CanvasProjection): ReadonlyArray<CanvasComponentProjection> => canvas.components.filter((component) => component.band === 'content')

// The window a component BELONGS to: the last one that begins at or above its
// own top. origins[0] is 0 and a component's y is non-negative, so this always
// answers, and for every component the engine actually paginated it answers
// the window that contains the component's top.
function homeWindow(origins: ReadonlyArray<number>, y: number): number {
  let home = 0
  for (let index = 1; index < origins.length; index += 1) {
    if ((origins[index] as number) <= y) home = index
    else break
  }
  return home
}

export function sheetStack(canvas: CanvasProjection): SheetStack {
  const origins = canvas.contentWindowOrigins
  // Named for what it is, not shortened to `height`: the authority contract
  // bans a window position derived by multiplying the window height by an
  // index, and a text guard can only catch the spelling it can see.
  const windowHeight = canvas.contentWindowHeight
  const drawn = Math.min(origins.length, MAX_CANVAS_SHEETS)
  const components = contentComponents(canvas)
  const homes = new Map<string, number>()
  for (const component of components) homes.set(component.id, homeWindow(origins, component.y))
  const sheets: Sheet[] = []
  for (let index = 0; index < drawn; index += 1) {
    const origin = origins[index] as number
    const next = origins[index + 1]
    const content: SheetOccurrence[] = []
    for (const component of components) {
      const home = homes.get(component.id) === index
      // Drawn on every window its extent intersects, and unconditionally on
      // its home window — so a component the engine never paginated (a text
      // element whose chain would not resolve contributes no extents) still
      // has exactly one occurrence somewhere rather than vanishing.
      const intersects = component.y < origin + windowHeight && component.y + component.height > origin
      if (home || intersects) content.push({ component, y: component.y - origin, home })
    }
    sheets.push({ index, origin, ...(next !== undefined && next - origin <= windowHeight ? { seam: next - origin } : {}), content })
  }
  return { sheets, windowCount: origins.length, truncated: origins.length > drawn, isFloor: canvas.contentWindowCountIsFloor }
}

// THE STACK'S DISPLAY-SPACE INVERSE, in document millipoints.
//
// A drag accumulates a linear pointer delta, which knows nothing about the
// repeated page-footer, the gap and the repeated page-header standing between
// one window's foot and the next window's head. Across a seam the component
// would drift from the hand by exactly that much. resize-anchor.ts's own
// header states the principle this repairs: a pointer that leaves the band
// should leave the component against that edge with its other axis still
// TRACKING THE HAND.
//
// The pitch is one whole page plus the stack's declared gap. The gap is the
// one display-space constant in the arithmetic, so it is divided by the zoom
// exactly once, here, and everything else is projected millipoints.
export function sheetPitch(canvas: CanvasProjection, zoom: number): number {
  return canvas.height + Math.round((SHEET_STACK_GAP / zoom) * 1000)
}

const contentBandTop = (canvas: CanvasProjection): number => (canvas.bands[1] as CanvasProjection['bands'][number]).y

// A column offset, as a distance down the whole stack from the top edge of
// sheet one's page.
export function stackYForColumn(stack: SheetStack, canvas: CanvasProjection, zoom: number, columnY: number): number {
  const origins = stack.sheets.map((sheet) => sheet.origin)
  const index = Math.min(homeWindow(origins, Math.max(columnY, 0)), stack.sheets.length - 1)
  return index * sheetPitch(canvas, zoom) + contentBandTop(canvas) + (columnY - (origins[index] as number))
}

// And back: which sheet a point down the stack falls on, and what column
// offset that is. The two are inverses on every point of every drawn sheet,
// which is the property the drag depends on and sheet-stack.test.ts asserts.
export function columnForStackY(stack: SheetStack, canvas: CanvasProjection, zoom: number, stackY: number): Readonly<{ window: number; columnY: number }> {
  const pitch = sheetPitch(canvas, zoom)
  const index = Math.min(Math.max(Math.floor(stackY / pitch), 0), stack.sheets.length - 1)
  const sheet = stack.sheets[index] as Sheet
  return { window: index, columnY: sheet.origin + (stackY - index * pitch - contentBandTop(canvas)) }
}

// What the drag actually asks for: given the column offset of the edge the
// gesture is moving and the raw document-space delta the pointer travelled,
// where does that edge end up in the COLUMN? One opaque number goes to Go,
// and it is a column coordinate, never a pin to a sheet.
export function columnEdgeAfterDrag(stack: SheetStack, canvas: CanvasProjection, zoom: number, columnEdge: number, delta: number): number {
  return columnForStackY(stack, canvas, zoom, stackYForColumn(stack, canvas, zoom, columnEdge) + delta).columnY
}
