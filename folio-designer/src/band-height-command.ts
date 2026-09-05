// STORY 12.1. The band-height command, as opaque Go-defined bytes.
//
// It is a COMPONENT command, not a page-setup one, and that is the whole reason
// this file exists rather than two more keys on pageSetupCommand: the engine's
// page-setup door gates on a seven-key arity that every caller's shape depends
// on, while the component door dispatches on `kind` and counts each arm's own
// fields. The seven font-chain factories are the shipped precedent for a
// document-level mutation encoded here.
//
// THIS MODULE HOLDS NO BOUND OF ITS OWN, and that is a ruling rather than an
// omission (Story 17.4 item 9, restated by 12.1's Q4). A floor at the lowest
// occupied edge of the band is computable from the projection alone — and a
// bound the browser can only derive from LAYOUT does not belong in the
// inspector, however convenient. The panel proposes the number the author
// typed, the engine refuses it with a located sentence, and the existing
// role="alert" path renders that sentence. Consistency with typing is the
// property the tests assert.
import { commandBytes, jsonNumber, jsonString } from './command-json'
import type { CappingBand } from './engine-protocol'

// Only the bands that CAP VERTICALLY have a height a command may set. `content`
// is absent by MEANING, not by omission: its height is derived from the page
// and the two bands above and below it, and the engine refuses a command naming
// it.
//
// THE LIST IS NOT SPELLED AGAIN HERE. Go's bandsCappingVertically is mirrored
// in engine-protocol.ts as BANDS_CAPPING_VERTICALLY, and
// engine-bounds-mirror.test.ts reads both sides; a union written out in this
// file would be a fourth copy standing outside that census, which is the only
// place a stale copy of this list can hide. CappingBand is that same list as a
// type.

// `height` is the author's DRAFT, passed as typed. jsonNumber tests it against
// the JSON number grammar and sends it byte for byte or sends `null` — an
// emptied box becomes `null` and the engine names the field, exactly as page
// setup already behaves. Nothing here re-computes it.
export function bandHeightCommand(band: CappingBand, height: string): ArrayBuffer {
  return commandBytes('setBandHeight', [['band', jsonString(band)], ['height', jsonNumber(height)]])
}
