// spec-section-break CAP-1. The two Section Break commands, as opaque
// Go-defined bytes.
//
// In the file the break is the content band's `sectionBreak` key, but on the
// canvas it is placed, dragged, nudged, typed and deleted like an element.
// Each of those is ONE command, so each is one undo entry.
//
// THIS MODULE HOLDS NO BOUND AND NO GRID. The engine snaps (`snap` is true for
// a placement and a drag, false for a typed Y and an arrow nudge) and the
// engine refuses a break outside the content band or through an element, with
// a located sentence naming what is in the way.
import { commandBytes, jsonBoolean, jsonNumber } from './command-json'

// `offset` is points, as the author typed it or as `points()` spelled a
// gesture's proposal. jsonNumber sends the literal byte for byte or `null`.
export function setSectionBreakCommand(offset: string, snap: boolean): ArrayBuffer {
  return commandBytes('setSectionBreak', [['offset', jsonNumber(offset)], ['snap', jsonBoolean(snap)]])
}

export function removeSectionBreakCommand(): ArrayBuffer {
  return commandBytes('removeSectionBreak', [])
}
