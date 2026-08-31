// This is an opaque Go command vocabulary, not a browser-side property or
// style model. A caller sends exact local literals only on an explicit commit.
export type PropertyField = 'x' | 'y' | 'width' | 'height' | 'value' | 'expression' | 'visibleIf' | 'fontFamily' | 'fontSize' | 'lineSpacing' | 'bold' | 'italic' | 'align' | 'valign' | 'color' | 'background' | 'borderWidth' | 'borderColor' | 'borderEdges' | 'paddingTop' | 'paddingRight' | 'paddingBottom' | 'paddingLeft'

export type PropertyIntent = Readonly<{ field: PropertyField; operation: 'set' | 'clear' | 'null'; value?: string | boolean | ReadonlyArray<string> }>

const pointFields = new Set<PropertyField>(['x', 'y', 'width', 'height', 'fontSize', 'borderWidth', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'])
// lineSpacing travels unquoted like a point field but is NOT one, and the
// distinction is worth a second set rather than an entry in the first.
// It is a dimensionless RATIO. Go's decoder (template.DecodeLineSpacingRaw)
// reads the author's own literal and performs the x1000 to thousandths
// itself, exactly as it does for a `lineSpacing` written in a .folio file —
// so `1.5` on the wire is 1500 thousandths in the document, and sending
// `1500` would be refused as 1 500 000, outside the load-time range. Verified
// against the engine, not inferred: the two entry points share one decoder
// precisely so the inspector cannot mean something different from the file.
const ratioFields = new Set<PropertyField>(['lineSpacing'])

export function updateComponentPropertiesCommand(ids: ReadonlyArray<string>, intent: PropertyIntent): ArrayBuffer {
  const change = intent.operation === 'clear' || intent.operation === 'null'
    ? `{"op":${quote(intent.operation)}}`
    : pointFields.has(intent.field) || ratioFields.has(intent.field)
      ? `{"op":"set","value":${rawNumberLiteral(intent.value)}}`
      : `{"op":"set","value":${propertyValue(intent.value)}}`
  return new TextEncoder().encode(`{"kind":"updateComponentProperties","version":1,"ids":[${ids.map(quote).join(',')}],"changes":{${quote(intent.field)}:${change}}}`).buffer
}

function rawNumberLiteral(value: PropertyIntent['value']): string {
  // Preserve the typed literal, unquoted; Go alone decides whether it is a
  // valid number, in whatever unit that field is carried in.
  return typeof value === 'string' ? value : ''
}

function propertyValue(value: PropertyIntent['value']): string {
  if (typeof value === 'string') return quote(value)
  if (typeof value === 'boolean') return String(value)
  return `[${(value ?? []).map(quote).join(',')}]`
}

// JSON.stringify IS the escape table, and the hand-rolled one it replaces was
// a strict subset of it: `\ " \n \r \t` and nothing else, while JSON requires
// every code point in U+0000-U+001F to be escaped. A value carrying any other
// C0 control — U+0001 from a paste, most plausibly — emitted a raw control
// byte inside a JSON string, so the command was MALFORMED BEFORE Go could read
// the field, and the engine answered with a generic parse failure instead of
// the located refusal naming the field. Engine-side validation cannot
// substitute for this: the bytes never reach the rule.
// This is a MINIMAL fix, deliberately. `rawNumberLiteral` above and the four
// other designer encoders are Story 15.2a's shared-command-JSON authority
// (DW-32), which must re-read this file rather than assume its earlier shape.
function quote(value: string): string { return JSON.stringify(value) }
