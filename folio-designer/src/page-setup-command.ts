// The command is deliberately assembled as opaque Go-defined bytes. It is
// not a TypeScript document interface and does not JSON stringify any .folio.
export function pageSetupCommand(preset: string, orientation: string, width: string, height: string, margin: Readonly<Record<'top' | 'right' | 'bottom' | 'left', string>>): ArrayBuffer {
  // Empty drafts stay explicit (`null`) so Go can return the field-specific
  // diagnostic instead of the UI silently restoring a previous value.
  const number = (value: string) => value === '' ? 'null' : value
  const text = `{"kind":"pageSetup","version":1,"preset":"${preset}","orientation":"${orientation}","width":${number(width)},"height":${number(height)},"margin":{"top":${number(margin.top)},"right":${number(margin.right)},"bottom":${number(margin.bottom)},"left":${number(margin.left)}}}`
  return new TextEncoder().encode(text).buffer
}
