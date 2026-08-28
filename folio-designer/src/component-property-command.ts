// This is an opaque Go command vocabulary, not a browser-side property or
// style model. A caller sends exact local literals only on an explicit commit.
export type PropertyField = 'x' | 'y' | 'width' | 'height' | 'value' | 'visibleIf' | 'fontFamily' | 'fontSize' | 'bold' | 'italic' | 'align' | 'valign' | 'background' | 'borderWidth' | 'borderColor' | 'borderEdges' | 'paddingTop' | 'paddingRight' | 'paddingBottom' | 'paddingLeft'

export type PropertyIntent = Readonly<{ field: PropertyField; operation: 'set' | 'clear' | 'null'; value?: string | boolean | ReadonlyArray<string> }>

const pointFields = new Set<PropertyField>(['x', 'y', 'width', 'height', 'fontSize', 'borderWidth', 'paddingTop', 'paddingRight', 'paddingBottom', 'paddingLeft'])

export function updateComponentPropertiesCommand(ids: ReadonlyArray<string>, intent: PropertyIntent): ArrayBuffer {
  const change = intent.operation === 'clear' || intent.operation === 'null'
    ? `{"op":${quote(intent.operation)}}`
    : pointFields.has(intent.field)
      ? `{"op":"set","value":${pointLiteral(intent.value)}}`
      : `{"op":"set","value":${propertyValue(intent.value)}}`
  return new TextEncoder().encode(`{"kind":"updateComponentProperties","version":1,"ids":[${ids.map(quote).join(',')}],"changes":{${quote(intent.field)}:${change}}}`).buffer
}

function pointLiteral(value: PropertyIntent['value']): string {
  // Preserve the typed literal; Go alone decides whether it is a valid point.
  return typeof value === 'string' ? value : ''
}

function propertyValue(value: PropertyIntent['value']): string {
  if (typeof value === 'string') return quote(value)
  if (typeof value === 'boolean') return String(value)
  return `[${(value ?? []).map(quote).join(',')}]`
}

function quote(value: string): string { return `"${value.replaceAll('\\', '\\\\').replaceAll('"', '\\"').replaceAll('\n', '\\n').replaceAll('\r', '\\r').replaceAll('\t', '\\t')}"` }
