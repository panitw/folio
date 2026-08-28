// Component commands are opaque, versioned Go vocabulary. This module does
// not describe a .folio element or serialize a document; it only encodes the
// final intent that crosses the existing worker FIFO.
export type PaletteKind = 'text' | 'image' | 'table' | 'line' | 'rect'

const encode = (text: string): ArrayBuffer => new TextEncoder().encode(text).buffer
const point = (value: number): string => {
  const rounded = Math.round(value * 1000) / 1000
  return Number.isFinite(rounded) ? String(rounded) : 'null'
}

// Canvas projections and interaction drafts use millipoints. Commands use
// canonical point literals, so the conversion occurs once at this opaque
// boundary rather than accidentally sending projection units to Go.
const millipoints = (value: number): string => point(value / 1000)

export function createComponentCommand(type: PaletteKind, band: 'pageHeader' | 'content' | 'pageFooter', x: number, y: number, snap: boolean): ArrayBuffer {
  return encode(`{"kind":"createComponent","version":1,"type":"${type}","band":"${band}","x":${point(x)},"y":${point(y)},"width":72,"height":24,"snap":${snap}}`)
}
export function dropComponentCommand(type: PaletteKind, pageX: number, pageY: number, snap: boolean): ArrayBuffer {
  return encode(`{"kind":"dropComponent","version":1,"type":"${type}","x":${point(pageX)},"y":${point(pageY)},"snap":${snap}}`)
}
export function moveComponentCommand(id: string, x: number, y: number, snap: boolean): ArrayBuffer {
  return encode(`{"kind":"moveComponent","version":1,"id":"${id}","x":${millipoints(x)},"y":${millipoints(y)},"snap":${snap}}`)
}
export function resizeComponentCommand(id: string, width: number, height: number, snap: boolean): ArrayBuffer {
  return encode(`{"kind":"resizeComponent","version":1,"id":"${id}","width":${millipoints(width)},"height":${millipoints(height)},"snap":${snap}}`)
}
export function deleteComponentCommand(id: string): ArrayBuffer {
	return encode(`{"kind":"deleteComponent","version":1,"id":"${id}"}`)
}
export function duplicateComponentCommand(id: string, snap: boolean): ArrayBuffer {
	return encode(`{"kind":"duplicateComponent","version":1,"id":"${id}","snap":${snap}}`)
}
