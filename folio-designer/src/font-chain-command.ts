// Font chain commands are OPAQUE ENGINE INTENT, never a browser-side font
// model. Nothing here knows what a duplicate name is, what an empty chain is,
// which chains an element still names, or which faces this build ships: every
// one of those is a rule the engine owns and answers with its own sentence.
// These six builders only put the author's ask on the wire, byte for byte, in
// the exact shape folio-go/component_commands.go's six handlers read.
//
// EVERY value is encoded with JSON.stringify — the table-column encoder's
// model, and the only correct answer to "escape this for JSON". A chain name
// is author-typed and unconstrained (Go accepts any non-empty UTF-8 string up
// to 512 bytes), so `"`, `\` and a pasted C0 control all reach this module,
// and a hand-rolled escape table that misses one produces bytes Go cannot
// parse — which loses the located refusal the panel exists to display.
//
// THE FIELD COUNT IS PART OF THE CONTRACT. componentFields(raw, N) counts
// every top-level key, `kind` and `version` included, and refuses anything
// else: add 4, rename 4, delete 3, addEntry 5, moveEntry 5, removeEntry 4.
// An extra field is not ignored, it is a refusal.
const encode = (value: string): ArrayBuffer => new TextEncoder().encode(value).buffer
const quote = (value: string): string => JSON.stringify(value)
// Go reads these with commandInt, which requires an integer literal. They are
// browser-derived list positions, never author text, and they are still
// encoded rather than spliced so this module has exactly one encoder.
const index = (value: number): string => JSON.stringify(value)

export function addFontChainCommand(name: string, entries: ReadonlyArray<string>): ArrayBuffer {
  return encode(`{"kind":"addFontChain","version":1,"name":${quote(name)},"entries":[${entries.map(quote).join(',')}]}`)
}

export function renameFontChainCommand(name: string, to: string): ArrayBuffer {
  return encode(`{"kind":"renameFontChain","version":1,"name":${quote(name)},"to":${quote(to)}}`)
}

export function deleteFontChainCommand(name: string): ArrayBuffer {
  return encode(`{"kind":"deleteFontChain","version":1,"name":${quote(name)}}`)
}

export function addFontChainEntryCommand(name: string, at: number, face: string): ArrayBuffer {
  return encode(`{"kind":"addFontChainEntry","version":1,"name":${quote(name)},"index":${index(at)},"face":${quote(face)}}`)
}

export function moveFontChainEntryCommand(name: string, from: number, to: number): ArrayBuffer {
  return encode(`{"kind":"moveFontChainEntry","version":1,"name":${quote(name)},"from":${index(from)},"to":${index(to)}}`)
}

export function removeFontChainEntryCommand(name: string, at: number): ArrayBuffer {
  return encode(`{"kind":"removeFontChainEntry","version":1,"name":${quote(name)},"index":${index(at)}}`)
}
