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
// One rectangle, one command. Any anchor other than the bottom-right corner
// moves the origin while it sizes, and move-then-resize would be two history
// entries for one drag with an intermediate rectangle nobody asked for.
export function setComponentBoundsCommand(id: string, x: number, y: number, width: number, height: number, snap: boolean): ArrayBuffer {
  return encode(`{"kind":"setComponentBounds","version":1,"id":"${id}","x":${millipoints(x)},"y":${millipoints(y)},"width":${millipoints(width)},"height":${millipoints(height)},"snap":${snap}}`)
}
export function deleteComponentCommand(id: string): ArrayBuffer {
	return encode(`{"kind":"deleteComponent","version":1,"id":"${id}"}`)
}
export function duplicateComponentCommand(id: string, snap: boolean): ArrayBuffer {
	return encode(`{"kind":"duplicateComponent","version":1,"id":"${id}","snap":${snap}}`)
}

// The tree supplies decoded JSON object-key segments only. Go owns the
// expression grammar, root/params scope, target eligibility, canonical
// mutation, and diagnostics; this factory intentionally does none of those.
export function bindComponentScalarCommand(id: string, segments: ReadonlyArray<string>): ArrayBuffer {
	// This encodes JSON transport only; it does not turn a key into a Folio
	// expression. Complete escaping keeps decoded keys unambiguous until Go
	// verifies the exact segment sequence.
	return encode(`{"kind":"bindComponentScalar","version":1,"id":${quote(id)},"segments":[${segments.map(quote).join(',')}]}`)
}

function quote(value: string): string {
	let encoded = '"'
	for (const character of value) {
		const code = character.charCodeAt(0)
		switch (character) {
			case '\\': encoded += '\\\\'; break
			case '"': encoded += '\\"'; break
			case '\b': encoded += '\\b'; break
			case '\f': encoded += '\\f'; break
			case '\n': encoded += '\\n'; break
			case '\r': encoded += '\\r'; break
			case '\t': encoded += '\\t'; break
			default: encoded += code <= 0x1f || (code >= 0xd800 && code <= 0xdfff) ? `\\u${code.toString(16).padStart(4, '0')}` : character
		}
	}
	return `${encoded}"`
}
