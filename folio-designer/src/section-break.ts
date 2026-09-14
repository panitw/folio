import type { CanvasProjection } from './engine-protocol'
import { homeWindow } from './sheet-stack'

// spec-section-break CAP-1 / CAP-6. THE ARITHMETIC BEHIND THE SECTION BREAK
// LINE, in the same millipoints Go projected, and nothing else.
//
// It measures no DOM and re-derives nothing the engine owns. The offset, the
// content band's height and the window origins all come from the projection;
// the engine snaps, and the engine refuses a break through an element or
// outside the content band. What moves during a drag is ONE PROPOSED LINE over
// ONE projected number, exactly as band-boundary.ts does for a band boundary.
//
// THE TWO BOUNDS IT HOLDS carry no information beyond "no further": a break
// lies strictly inside the content band — the engine refuses 0 and the content
// height alike — so the proposal stops one millipoint inside each. A break dragged through an element is NOT clamped — the engine's
// refusal names the element in the way, and a silent drag limit would hide it.

function clamp(value: number, low: number, high: number): number { return Math.min(Math.max(value, low), Math.max(low, high)) }

// The content band's projected height: one window, the only bound used here.
export function contentBandHeight(canvas: CanvasProjection): number {
  return canvas.bands.find((band) => band.name === 'content')?.height ?? 0
}

// proposedSectionBreak turns pointer travel into a candidate offset. `dy` is
// DOWNWARD travel in millipoints, already through canvasDisplay.documentDelta.
// Rounded to a whole millipoint for the reason proposedBandHeight gives: at any
// zoom but 1 the product is a float artifact, not a JSON number.
export function proposedSectionBreak(original: number, dy: number, contentHeight: number): number {
  return clamp(Math.round(original + dy), 1, contentHeight - 1)
}

// The sheet the line is drawn on, once and with no echoes: the window that
// holds the offset (`homeWindow`), and the line's offset within that sheet.
export function sectionBreakPlacement(canvas: CanvasProjection): Readonly<{ sheet: number; y: number }> | undefined {
  const offset = canvas.sectionBreak
  if (offset === undefined) return undefined
  const sheet = homeWindow(canvas.contentWindowOrigins, offset)
  return { sheet, y: offset - (canvas.contentWindowOrigins[sheet] ?? 0) }
}
