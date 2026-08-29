// Story 5.13's asset-authoring command factory. It stays exactly as opaque
// as component-command.ts/component-property-command.ts: this module knows
// nothing about the .folio format, does not hash, does not sniff the file's
// real format, and does not decide legality — it only base64-encodes the
// bytes the browser already read and carries the browser/OS's OWN declared
// media type opaquely to Go, which is the sole authority (component_commands.go's
// setComponentAsset, D-5.13.1).
export function setComponentAssetCommand(id: string, mediaType: string, bytes: ArrayBuffer): ArrayBuffer {
  return new TextEncoder().encode(`{"kind":"setComponentAsset","version":1,"id":${quote(id)},"mediaType":${quote(mediaType)},"data":${quote(bytesToBase64(bytes))}}`).buffer
}

function bytesToBase64(bytes: ArrayBuffer): string {
  let text = ''
  for (const byte of new Uint8Array(bytes)) text += String.fromCharCode(byte)
  return btoa(text)
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

// assetBytesRequest is the opaque per-key paintable-bytes request payload
// (D-5.13.2's "Producer" clause): the engine's 'asset' operation decodes
// this UTF-8 key directly, mirroring how 'command' payloads are opaque
// bytes Go alone parses.
export function assetBytesRequest(key: string): ArrayBuffer {
  return new TextEncoder().encode(key).buffer
}
