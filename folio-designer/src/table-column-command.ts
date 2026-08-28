// Table column commands are opaque engine intent, never a browser table model.
const encode = (value: string): ArrayBuffer => new TextEncoder().encode(value).buffer
const quote = (value: string): string => JSON.stringify(value)
const number = (value: number): string => JSON.stringify(value)

export function addTableColumnCommand(id: string, index: number): ArrayBuffer { return encode(`{"kind":"addTableColumn","version":1,"id":${quote(id)},"index":${index}}`) }
export function removeTableColumnCommand(id: string, columnId: string): ArrayBuffer { return encode(`{"kind":"removeTableColumn","version":1,"id":${quote(id)},"columnId":${quote(columnId)}}`) }
export function moveTableColumnCommand(id: string, columnId: string, toIndex: number): ArrayBuffer { return encode(`{"kind":"moveTableColumn","version":1,"id":${quote(id)},"columnId":${quote(columnId)},"toIndex":${toIndex}}`) }
export function updateTableColumnCommand(id: string, columnId: string, field: 'header' | 'width' | 'align', value: string | number): ArrayBuffer { return encode(`{"kind":"updateTableColumn","version":1,"id":${quote(id)},"columnId":${quote(columnId)},"field":${quote(field)},"value":${typeof value === 'string' ? quote(value) : number(value)}}`) }
export function configureTableBindingCommand(id: string, collection: string, alias: string): ArrayBuffer { return encode(`{"kind":"configureTableBinding","version":1,"id":${quote(id)},"collection":${quote(collection)},"alias":${quote(alias)}}`) }
export function updateTableColumnBindingCommand(id: string, columnId: string, field: string): ArrayBuffer { return encode(`{"kind":"updateTableColumnBinding","version":1,"id":${quote(id)},"columnId":${quote(columnId)},"field":${quote(field)}}`) }
export function updateTableColumnFooterCommand(id: string, columnId: string, footer: string, footerOf: string, footerFormat: string): ArrayBuffer { return encode(`{"kind":"updateTableColumnFooter","version":1,"id":${quote(id)},"columnId":${quote(columnId)},"footer":${quote(footer)},"footerOf":${quote(footerOf)},"footerFormat":${quote(footerFormat)}}`) }
