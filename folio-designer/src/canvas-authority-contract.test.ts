import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const sourceDir = path.dirname(fileURLToPath(import.meta.url))
const designerRoot = path.dirname(sourceDir)
const production = fs.readdirSync(sourceDir, { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx|css)$/.test(entry) && !/\.test\.(?:ts|tsx)$/.test(entry))
  .map((entry) => path.join(sourceDir, entry))
const e2e = fs.readdirSync(path.join(designerRoot, 'e2e'), { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry))
  .map((entry) => path.join(designerRoot, 'e2e', entry))
const tests = fs.readdirSync(sourceDir, { recursive: true })
  .filter((entry): entry is string => typeof entry === 'string' && /\.test\.(?:ts|tsx)$/.test(entry) && entry !== 'canvas-authority-contract.test.ts')
  .map((entry) => path.join(sourceDir, entry))

const prohibited = [
  /(?:CanvasRenderingContext2D|\bctx)\.measureText/,
  /\b(?:getBoundingClientRect|getClientRects)\s*\(/,
  /\b(?:offset(?:Width|Height|Left|Top|Parent)|client(?:Width|Height|Left|Top)|scroll(?:Width|Height|Left|Top))\b/,
  /\boffset[XY]\b/,
  /\bResizeObserver\b/,
  /\bdocument\.fonts\b/,
  /\bdevicePixelRatio\b/,
  /\b(?:Range|document\.createRange|getSelection|Selection)\s*\(/,
  /\bgetComputedStyle\s*\(/,
  /(?:white-space|text-wrap|overflow-wrap|word-break|line-clamp|text-align)\s*:\s*(?:normal|wrap|balance|pretty|anywhere|break-word|justify|\d+)/,
  // D-7.4.2 §5. The canvas paint is APPROXIMATE and, since Story 7.4, may be
  // a deliberately truncated prefix; pagination is EXACT and comes from
  // layout.Paginate. Deriving a height, a page count or a window count from
  // the paint's line count would make a truncated paint shorten the reported
  // column, and Story 7.6 would then draw the wrong number of sheets — the
  // canvas lying about pagination. The independence holds in Go today by
  // construction (nothing there reads CanvasTextPaint), but the data sits
  // right here in `line.advance`, one plausible line of designer code away.
  // A positive test would not catch that line being written; this does.
  /\b(?:textPaint|paint)\??\.lines\.length\b/,
  /\blines\.length\s*[*/]|[*/]\s*\blines\.length\b/,
  // Story 7.6 / AC2, and the same guard one step further along. A window's
  // COUNT may not come from the paint; a window's POSITION may not come from
  // arithmetic at all. The window height multiplied by an index is the closed form
  // internal/layout/paginate.go forbids by name — the window advances to the
  // top of the first item that did not fit, never by a fixed height — and it
  // is measurably wrong: 110 millipoints per window adrift on a column of
  // round 728pt spacing, and eleven windows where the engine says two on a
  // column with a declared gap. The origins are PROJECTED
  // (contentWindowOrigins); multiplying is the one plausible line of designer
  // code that would quietly replace them, in either operand order.
  /\b(?:contentWindowHeight|windowHeight)\s*\*|\*\s*[\w.?]*\b(?:contentWindowHeight|windowHeight)\b/,
]

function violations(files: readonly string[]): string[] {
  return files.flatMap((file) => {
    const source = withoutApprovedLocalPointerInput(file, fs.readFileSync(file, 'utf8')).replace(/document\.fonts\b/g, 'fontReadinessOnly')
    const name = path.relative(designerRoot, file)
    return prohibited.filter((pattern) => pattern.test(source)).map((pattern) => `${name}: ${pattern}`)
  })
}

describe('canvas projection authority contract', () => {
  it('scans a non-vacuous production, unit-test, and e2e corpus for browser measurement authority', () => {
    expect(production.length).toBeGreaterThan(10)
    expect(tests.length).toBeGreaterThan(10)
    expect(e2e.length).toBeGreaterThan(3)
    expect(violations([...production, ...tests, ...e2e])).toEqual([])
  })

  it('allows only the non-document reduced-motion media rule', () => {
    const css = fs.readFileSync(path.join(sourceDir, 'App.css'), 'utf8')
    expect([...css.matchAll(/@media\s*\(([^)]+)\)/g)].map((match) => match[1])).toEqual(['prefers-reduced-motion: reduce'])
  })

  it('turns realistic measurement, CSS, range, and event-coordinate mutations red', () => {
    expect(violationsForSource(`const width = ${['CanvasRenderingContext2D', 'measureText'].join('.')}("x")`)).not.toEqual([])
    expect(violationsForSource('.paint { white-space: normal }')).not.toEqual([])
    expect(violationsForSource(`const x = event.${['offset', 'X'].join('')}`)).not.toEqual([])
    expect(violationsForSource('const style = getComputedStyle(node)')).not.toEqual([])
    expect(violationsForSource(`const range = ${['document', 'createRange'].join('.') }()`)).not.toEqual([])
    expect(violationsForSource('.paint { text-align: justify }')).not.toEqual([])
    expect(violationsForSource('const height = component.textPaint.lines.length * line.advance')).not.toEqual([])
    expect(violationsForSource('const windows = Math.ceil(paint.lines.length / perPage)')).not.toEqual([])
    expect(violationsForSource('const top = canvas.contentWindowHeight * index')).not.toEqual([])
    expect(violationsForSource('const top = index * canvas.contentWindowHeight')).not.toEqual([])
    expect(violationsForSource('const top = sheet * windowHeight')).not.toEqual([])
  })
})

function violationsForSource(source: string): RegExp[] { return prohibited.filter((pattern) => pattern.test(source)) }

function withoutApprovedLocalPointerInput(file: string, source: string): string {
  if (file.includes(`${path.sep}preview${path.sep}`)) {
    // PDF viewer scroll is deliberately transient viewer navigation, never a
    // document/canvas measurement. Keep that exception narrow to this owner.
    return source.replace(/scroll(?:Width|Height|Left|Top)\b/g, 'viewerTransientState')
  }
  if (path.basename(file) !== 'App.tsx') return source
  // The sole approved pointer coordinate is isolated to a named transient
  // proposal helper. It is not DOM measurement and never reaches paint.
  const seam = /export function placementPoint\(event: Pick<MouseEvent,[\s\S]*?\n}\nfunction pageStyle/
  expect(source).toMatch(seam)
  return source.replace(seam, 'function pageStyle')
}
