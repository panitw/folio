import type { CanvasProjection } from './engine-protocol'
import { sheetPitch, sheetStack } from './sheet-stack'

export type SelectionPoint = Readonly<{ x: number; y: number }>
export type SelectionRectangle = Readonly<{ left: number; top: number; right: number; bottom: number }>

export function selectionRectangle(a: SelectionPoint, b: SelectionPoint): SelectionRectangle {
  return { left: Math.min(a.x, b.x), top: Math.min(a.y, b.y), right: Math.max(a.x, b.x), bottom: Math.max(a.y, b.y) }
}

export function enclosedComponents(canvas: CanvasProjection, zoom: number, rect: SelectionRectangle): string[] {
  const stack = sheetStack(canvas)
  const pitch = sheetPitch(canvas, zoom)
  const encloses = (x: number, y: number, width: number, height: number) => rect.left <= x && rect.top <= y && rect.right >= x + width && rect.bottom >= y + height
  return canvas.components.filter((component) => {
    const band = canvas.bands.find((entry) => entry.name === component.band)!
    const x = band.x + component.x
    if (component.band !== 'content') return stack.sheets.some((sheet) => encloses(x, sheet.index * pitch + band.y + component.y, component.width, component.height))
    if (component.height === 0) return stack.sheets.some((sheet) => component.y >= sheet.origin && component.y <= sheet.origin + Math.min(canvas.contentWindowHeight, sheet.seam ?? canvas.contentWindowHeight) && encloses(x, sheet.index * pitch + band.y + component.y - sheet.origin, component.width, 0))
    let covered = component.y
    const end = component.y + component.height
    for (const sheet of stack.sheets) {
      const start = Math.max(component.y, sheet.origin)
      const stop = Math.min(end, sheet.origin + Math.min(canvas.contentWindowHeight, sheet.seam ?? canvas.contentWindowHeight))
      if (stop <= start) continue
      // Seam masks, skipped column intervals and the drawing cap must not
      // turn a clipped continuation into an enclosed complete component.
      if (start > covered || !encloses(x, sheet.index * pitch + band.y + start - sheet.origin, component.width, stop - start)) return false
      covered = Math.max(covered, stop)
    }
    return covered >= end
  }).map((component) => component.id)
}

// Move paint's absolute coordinates with the box while retaining its offsets.
// Geometry remains a captured projection plus the engine's accepted delta.
export function translatedCanvas(canvas: CanvasProjection, ids: ReadonlyArray<string>, dx: number, dy: number): CanvasProjection {
  const targets = new Set(ids)
  return { ...canvas, components: canvas.components.map((component) => targets.has(component.id) ? {
    ...component, x: component.x + dx, y: component.y + dy,
    ...(component.image ? { image: { ...component.image, drawX: component.image.drawX + dx, drawY: component.image.drawY + dy } } : {}),
    ...(component.textPaint ? { textPaint: { ...component.textPaint, lines: component.textPaint.lines.map((line) => ({ ...line, top: line.top + dy, baseline: line.baseline + dy, fragments: line.fragments.map((fragment) => ({ ...fragment, x: fragment.x + dx })) })) } } : {}),
  } : component) }
}
