// STORY 12.3. A table's header height, its alternating row background, and one
// field of its header style — as opaque Go-defined bytes.
//
// THREE KINDS, NOT ONE, because Go has three arms and not one: a command names
// exactly what it changes (Story 15.2a). They are also three top-level kinds
// rather than three keys threaded through `updateComponentProperties`, and that
// is `setComponentAsset`'s own shipped ruling rather than a preference —
// anything the `{op,value}` grammar cannot express, or where CLEAR must stay
// inexpressible, becomes its own kind. `headerHeight` is exactly that: the
// format requires it, so there is no clear for it to have, and NO CLEAR
// AFFORDANCE IS OFFERED ANYWHERE.
//
// THIS MODULE HOLDS NO RULE OF ITS OWN. It does not clamp, normalise, validate
// a colour, or decide what a legal font chain is. Those are engine rules — one
// predicate each, asked by the loader and the command door alike — and the
// panel renders the engine's own located sentence. What it does decide is
// TRANSPORT: which of the two JSON scalar types each field travels as, because
// Go decodes `fontSize` with a length decoder and `align` with a string one,
// and a value in the wrong JSON type could not reach either rule to be judged
// by it.
//
// DRAFTS TRAVEL AS TYPED. `jsonNumber` tests the author's literal against the
// JSON number grammar and sends it byte for byte or sends `null`; nothing here
// runs Number(). An emptied box therefore reaches Go as `null` (numeric) or
// `""` (string) and the engine names the field — never as a value nobody typed.
// That is band-height-command.ts's promise, in both the shapes it takes.
//
// CLEARING IS `op: "clear"`, WHICH REMOVES THE KEY. It is not `op: "null"` —
// no field in this story accepts that, and all three arms refuse it — because
// an explicit null is still the key in the file: it changes the bytes, burns an
// undo entry and raises the document's required format version.
import { commandBytes, jsonNumber, jsonString, type JsonField } from './command-json'

// The seven header-style fields a command may author. The four `Style` fields
// absent from this union are each a ruling, not an oversight: `border` is
// deferred to Story 14.8's BORDERS section, `padding` is forbidden outright by
// D-12.4.1, and `bold`/`italic` have no arm in the engine's header cascade to
// resolve from — a header style declaring either would be stored and read by
// nothing.
export type TableHeaderStyleField = 'fontFamily' | 'fontSize' | 'lineSpacing' | 'background' | 'color' | 'valign' | 'align'

// The two header-style fields whose value is a NUMBER on the wire: a length in
// points and a dimensionless ratio. Everything else is a string. This is a
// transport fact about Go's decoders, not a second opinion about what a legal
// value is.
const NUMERIC_HEADER_STYLE_FIELDS: ReadonlyArray<TableHeaderStyleField> = ['fontSize', 'lineSpacing']

// `height` is the author's DRAFT, in points, passed as typed. There is no
// clear: `headerHeight` is required by the format (`parse_bands.go` hard-errors
// on its absence and `serialize.go` emits it unconditionally), so a cleared one
// is a document that cannot be reopened. The engine refuses a clear op on it
// too — this factory simply cannot build one.
export function tableHeaderHeightCommand(id: string, height: string): ArrayBuffer {
  return commandBytes('setTableHeaderHeight', [['id', jsonString(id)], ['height', jsonNumber(height)]])
}

// The alternating row background: a `#RRGGBB` draft, or a clear that removes
// the key and returns those rows to `style.background`. Story 4.8 already
// renders it; this only writes it.
export function tableAltRowBackgroundCommand(id: string, operation: 'set' | 'clear', value = ''): ArrayBuffer {
  return commandBytes('setTableAltRowBackground', [['id', jsonString(id)], ...operationFields(operation, jsonString(value))])
}

// One field of the header-only Style block. A cleared field is removed from the
// block, and clearing the LAST one removes the block itself — both are the
// engine's doing, not this module's.
export function tableHeaderStyleCommand(id: string, field: TableHeaderStyleField, operation: 'set' | 'clear', value = ''): ArrayBuffer {
  const encoded = NUMERIC_HEADER_STYLE_FIELDS.includes(field) ? jsonNumber(value) : jsonString(value)
  return commandBytes('updateTableHeaderStyle', [['id', jsonString(id)], ['field', jsonString(field)], ...operationFields(operation, encoded)])
}

// The `{op[, value]}` tail both clearable kinds share. A clear carries NO
// value, and that is arity rather than politeness: Go counts every top-level
// key and refuses any other count, so a clear that carried one would be refused
// whole.
function operationFields(operation: 'set' | 'clear', encoded: string): ReadonlyArray<JsonField> {
  return operation === 'clear' ? [['op', jsonString('clear')]] : [['op', jsonString('set')], ['value', encoded]]
}
