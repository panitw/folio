// Table column commands are opaque engine intent, never a browser table model.
//
// STORY 15.2a: this file WAS NOT wholly correct, and the story's Code Map
// called it "the shipped model" — the correction is worth keeping because it is
// what the consolidation is for. Its `quote` and `number` helpers were both
// JSON.stringify, and then TWO of its seven builders ignored `number` and
// spliced a raw `${index}`/`${toIndex}` anyway. Nothing reached those two —
// both are browser-derived list positions, never author text — so it was a
// latent hole rather than a live defect, and it survived precisely because the
// file read as the one that had got it right.
//
// That is the argument for routing all SIX encoders through one module rather
// than auditing them: a correct helper sitting beside a call site that does not
// use it looks identical to a correct file.
import { commandBytes, jsonNumber, jsonString } from './command-json'

export function addTableColumnCommand(id: string, index: number): ArrayBuffer { return commandBytes('addTableColumn', [['id', jsonString(id)], ['index', jsonNumber(index)]]) }
export function removeTableColumnCommand(id: string, columnId: string): ArrayBuffer { return commandBytes('removeTableColumn', [['id', jsonString(id)], ['columnId', jsonString(columnId)]]) }
export function moveTableColumnCommand(id: string, columnId: string, toIndex: number): ArrayBuffer { return commandBytes('moveTableColumn', [['id', jsonString(id)], ['columnId', jsonString(columnId)], ['toIndex', jsonNumber(toIndex)]]) }
export function updateTableColumnCommand(id: string, columnId: string, field: 'header' | 'width' | 'align', value: string | number): ArrayBuffer { return commandBytes('updateTableColumn', [['id', jsonString(id)], ['columnId', jsonString(columnId)], ['field', jsonString(field)], ['value', typeof value === 'string' ? jsonString(value) : jsonNumber(value)]]) }
export function configureTableBindingCommand(id: string, collection: string, alias: string): ArrayBuffer { return commandBytes('configureTableBinding', [['id', jsonString(id)], ['collection', jsonString(collection)], ['alias', jsonString(alias)]]) }
export function updateTableColumnBindingCommand(id: string, columnId: string, field: string): ArrayBuffer { return commandBytes('updateTableColumnBinding', [['id', jsonString(id)], ['columnId', jsonString(columnId)], ['field', jsonString(field)]]) }
export function updateTableColumnFooterCommand(id: string, columnId: string, footer: string, footerOf: string, footerFormat: string): ArrayBuffer { return commandBytes('updateTableColumnFooter', [['id', jsonString(id)], ['columnId', jsonString(columnId)], ['footer', jsonString(footer)], ['footerOf', jsonString(footerOf)], ['footerFormat', jsonString(footerFormat)]]) }
