// The eight handles on a selected component. South-east is listed first only
// because it paints under the other seven: it keeps the larger hit target the
// image-asset e2e depends on, and the later siblings win where the two
// overlap on a small component.
export const resizeAnchors = ['se', 'nw', 'n', 'ne', 'w', 'e', 'sw', 's'] as const
export type ResizeAnchor = (typeof resizeAnchors)[number]
export type DragAnchor = ResizeAnchor | 'move'

// A drag proposal, in the same millipoints Go projected. Nothing here measures
// the DOM, reads layout, or decides what is legal: Go still snaps the numbers,
// contains them in the band, and is free to reject the whole rectangle.
//
// What the edges below do own is where a gesture *runs out*. A pointer that
// leaves the band should leave the component against that edge with its other
// axis still tracking the hand, not hand Go a rectangle it can only reject —
// a rejected drag lands the component back where it started, which reads as
// the drag having been thrown away. The floor keeps a rectangle from
// inverting under the pointer for the same reason.
const minimumSize = 1000

export type Bounds = Readonly<{ x: number; y: number; width: number; height: number }>
export type DragOrigin = Readonly<{ originalX: number; originalY: number; originalWidth: number; originalHeight: number }>
// The band the component is being dragged inside, in projected millipoints.
// Component geometry is band-relative, so the band's own size is the limit.
export type DragLimit = Readonly<{ width: number; height: number }>

function clamp(value: number, low: number, high: number): number { return Math.min(Math.max(value, low), Math.max(low, high)) }

export function proposedBounds(anchor: DragAnchor, origin: DragOrigin, dx: number, dy: number, limit?: DragLimit): Bounds {
  const limitWidth = limit ? limit.width : Number.POSITIVE_INFINITY
  const limitHeight = limit ? limit.height : Number.POSITIVE_INFINITY
  if (anchor === 'move') {
    return {
      x: clamp(origin.originalX + dx, 0, limitWidth - origin.originalWidth),
      y: clamp(origin.originalY + dy, 0, limitHeight - origin.originalHeight),
      width: origin.originalWidth,
      height: origin.originalHeight,
    }
  }
  // A resize moves the edges the anchor names and leaves the opposite ones
  // where they were, so the rectangle is easier to state as four edges than
  // as an origin plus a size.
  const west = anchor === 'nw' || anchor === 'w' || anchor === 'sw'
  const east = anchor === 'ne' || anchor === 'e' || anchor === 'se'
  const north = anchor === 'nw' || anchor === 'n' || anchor === 'ne'
  const south = anchor === 'sw' || anchor === 's' || anchor === 'se'
  const originalRight = origin.originalX + origin.originalWidth
  const originalBottom = origin.originalY + origin.originalHeight
  const left = west ? clamp(origin.originalX + dx, 0, originalRight - minimumSize) : origin.originalX
  const right = east ? clamp(originalRight + dx, origin.originalX + minimumSize, limitWidth) : originalRight
  const top = north ? clamp(origin.originalY + dy, 0, originalBottom - minimumSize) : origin.originalY
  const bottom = south ? clamp(originalBottom + dy, origin.originalY + minimumSize, limitHeight) : originalBottom
  return { x: left, y: top, width: right - left, height: bottom - top }
}
