import { expect, test, type Page } from '@playwright/test'
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'

const sample = Buffer.from(`{
  "customer": {"name": "Ada Lovelace"},
  "account": {"number": "001-9"},
  "period": {"from": "2026-08-01", "to": "2026-08-29"},
  "transactions": [
    {"date":"2026-08-01","description": "Opening balance — สวัสดี 世界", "debitDisplay":"", "creditDisplay":"1,250.10", "balanceDisplay":"1,250.10", "amount": 1250.10},
    {"date":"2026-08-29","description": "Service charge — 支払い", "debitDisplay":"480.33", "creditDisplay":"", "balanceDisplay":"769.77", "amount": 480.33}
  ]
}\n`)
const params = Buffer.from(`{\n  "reportDate": "2026-08-29",\n  "generatedDate": "2026-08-29",\n  "documentDate": "2026-08-29T00:00:00Z"\n}\n`)
const root = path.resolve(import.meta.dirname, '../..')
const goRoot = path.join(root, 'folio-go')

type CapturedRequest = Readonly<{ sessionId: string; operation: string; requestId: string; command?: number[]; render?: Readonly<{ template: number[]; data: number[]; params: number[] }> }>
type Captured = Readonly<{ requests: ReadonlyArray<CapturedRequest>; responses: ReadonlyArray<string>; serializations: ReadonlyArray<number[]>; renders: ReadonlyArray<number[]>; failures: ReadonlyArray<string> }>
type BrowserWitness = Readonly<{ template: Buffer; data: Buffer; params: Buffer; pdf: Buffer; repeatedPDF: Buffer; requests: ReadonlyArray<CapturedRequest>; serialized: number }>

// The observer wraps the existing Worker constructor before the app starts.
// It copies messages for evidence only; it neither constructs another worker
// nor changes an opaque request/response byte array. PDF.js admission is still
// proved by the visible "EXACT LOCAL PRODUCTION PDF" state below.
async function observeOneWorker(page: Page, sessionId: string): Promise<void> {
  await page.addInitScript((session) => {
    // The proof deliberately exercises the fallback file picker/download
    // route so Playwright can retain the exact downloaded bytes.
    Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined })
    const NativeWorker = window.Worker
    const requestOperations = new Map<string, string>()
    const capture = { requests: [] as Array<CapturedRequest>, responses: [] as string[], serializations: [] as number[][], renders: [] as number[][], failures: [] as string[] }
    class ObservedWorker extends NativeWorker {
      constructor(...args: ConstructorParameters<typeof Worker>) {
        super(...args)
        this.addEventListener('message', (event: MessageEvent<unknown>) => {
          const message = event.data as { kind?: string; ok?: boolean; requestId?: string; bytes?: ArrayBuffer; error?: { code?: string; message?: string } }
          const operation = message.requestId ? requestOperations.get(message.requestId) : undefined
          if (message.kind === 'response') capture.responses.push(`${operation ?? 'unknown'}:${message.ok ? 'ok' : 'error'}`)
          if (message.kind === 'response' && message.ok && message.bytes && operation === 'serialize') capture.serializations.push(Array.from(new Uint8Array(message.bytes)))
          if (message.kind === 'response' && message.ok && message.bytes && operation === 'render') capture.renders.push(Array.from(new Uint8Array(message.bytes)))
          if (message.kind === 'response' && !message.ok) capture.failures.push(`${operation ?? 'unknown'}:${message.error?.code ?? 'unknown'}:${message.error?.message ?? 'missing message'}`)
        })
        this.addEventListener('error', (event) => capture.failures.push(`worker-runtime:${event.message || 'missing message'}`))
      }
      postMessage(message: unknown, transfer: Transferable[]): void
      postMessage(message: unknown, options?: StructuredSerializeOptions): void
      postMessage(message: unknown, transferOrOptions?: Transferable[] | StructuredSerializeOptions): void {
        const request = message as { kind?: string; operation?: string; requestId?: string; payload?: ArrayBuffer | { template: ArrayBuffer; data: ArrayBuffer; params: ArrayBuffer } }
        if (request.kind === 'request' && request.operation && request.requestId) {
          requestOperations.set(request.requestId, request.operation)
          const bytes = (value: ArrayBuffer) => Array.from(new Uint8Array(value.slice(0)))
          const record: { sessionId: string; operation: string; requestId: string; command?: number[]; render?: { template: number[]; data: number[]; params: number[] } } = { sessionId: session, operation: request.operation, requestId: request.requestId }
          if (request.operation === 'command' && request.payload instanceof ArrayBuffer) record.command = bytes(request.payload)
          if (request.operation === 'render' && request.payload && !(request.payload instanceof ArrayBuffer)) record.render = { template: bytes(request.payload.template), data: bytes(request.payload.data), params: bytes(request.payload.params) }
          capture.requests.push(record)
        }
        if (Array.isArray(transferOrOptions)) super.postMessage(message, transferOrOptions)
        else super.postMessage(message, transferOrOptions)
      }
    }
    Object.assign(window, { Worker: ObservedWorker, __folioRoundTripCapture: capture })
  }, sessionId)
}

async function captured(page: Page): Promise<Captured> {
  return page.evaluate(() => (window as typeof window & { __folioRoundTripCapture: Captured }).__folioRoundTripCapture)
}

// The inspector is one tabbed panel: PROPERTIES (INPUTS in Preview) and DATA.
// Each authoring helper opens the tab holding the control it drives.
async function openTab(page: Page, name: 'PROPERTIES' | 'DATA' | 'INPUTS'): Promise<void> {
  await page.getByRole('tab', { name }).click()
}

async function setFontFamily(page: Page): Promise<void> {
  await openTab(page, 'PROPERTIES')
  const font = page.getByRole('textbox', { name: 'Font family' })
  await font.fill('body')
  await font.press('Enter')
  await expect(font).toHaveValue('body')
}

async function loadSample(page: Page): Promise<void> {
  await openTab(page, 'DATA')
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Load sample JSON' }).click()
  await (await chooser).setFiles({ name: 'authored-session-data.json', mimeType: 'application/json', buffer: sample })
  await expect(page.getByRole('tree', { name: 'Sample data paths' })).toBeVisible()
}

async function revision(page: Page): Promise<number> {
  const text = await page.getByTestId('engine-snapshot').textContent()
  const found = text?.match(/REVISION (\d+)/)
  if (!found) throw new Error(`could not read engine revision from ${text}`)
  return Number(found[1])
}

async function waitForRevisionAdvance(page: Page, before: number): Promise<void> {
  await expect.poll(() => revision(page), { timeout: 12_000 }).toBeGreaterThan(before)
}

async function bindTextToCustomer(page: Page, content: ReturnType<Page['getByRole']>): Promise<void> {
  const texts = content.getByRole('button', { name: /text component/ })
  const beforeCount = await texts.count()
  await page.getByRole('button', { name: 'Place Text' }).click()
  await content.press('Enter')
  await expect(texts).toHaveCount(beforeCount + 1, { timeout: 12_000 })
  await texts.last().click()
  await setFontFamily(page)
  await openTab(page, 'DATA')
  const tree = page.getByRole('tree', { name: 'Sample data paths' })
  const customer = tree.getByRole('treeitem').filter({ hasText: /^customer/ })
  await customer.click()
  const name = tree.getByRole('treeitem').filter({ hasText: /^name/ })
  await expect(name).toBeVisible()
  await name.click()
  await page.getByRole('button', { name: 'Connect selected path' }).click()
  await openTab(page, 'PROPERTIES')
  await expect(page.getByText('Bound to').locator('..')).toContainText('customer.name')
}

async function placeStatementText(page: Page, band: ReturnType<Page['getByRole']>, value: string, y: number, expression = false): Promise<void> {
  const texts = band.getByRole('button', { name: /text component/ })
  const beforeCount = await texts.count()
  await page.getByRole('button', { name: 'Place Text' }).click()
  await band.press('Enter')
  await expect(texts).toHaveCount(beforeCount + 1, { timeout: 12_000 })
  const text = texts.last()
  await text.click()
  // A palette text starts deliberately style-free. Use the visible authoring
  // control so each statement field is independently renderable.
  await setFontFamily(page)
  if (y !== 0) {
    const yField = page.getByRole('textbox', { name: 'Y (pt)' })
    const beforeY = await revision(page)
    await yField.fill(String(y))
    await yField.press('Enter')
    await waitForRevisionAdvance(page, beforeY)
  }
  const field = page.getByRole('textbox', { name: expression ? 'Text expression' : 'Text value' })
  const beforeValue = await revision(page)
  await field.fill(value)
  await field.press('Enter')
  await waitForRevisionAdvance(page, beforeValue)
}

async function authorStatementFraming(page: Page): Promise<void> {
  const header = page.getByRole('region', { name: 'Page Header', exact: true })
  const content = page.getByRole('region', { name: 'Content', exact: true })
  const footer = page.getByRole('region', { name: 'Page Footer', exact: true })
  await page.getByRole('button', { name: 'Place Image' }).click()
  await header.press('Enter')
  await expect(header.getByRole('button', { name: /image component/ })).toHaveCount(1)
  await placeStatementText(page, header, 'CUSTOMER ACCOUNT STATEMENT', 30)
  await placeStatementText(page, content, 'Customer: {{customer.name}}', 30, true)
  await placeStatementText(page, content, 'Account: {{account.number}}', 60, true)
  await placeStatementText(page, content, 'Statement period: {{period.from}} to {{period.to}}', 90, true)
  await placeStatementText(page, footer, 'Confidential — generated {{params.generatedDate}} / report {{params.reportDate}}', 0, true)
  await placeStatementText(page, footer, 'Page {{page}} of {{pages}}', 12, true)
}

async function authorTableWithFooter(page: Page, content: ReturnType<Page['getByRole']>): Promise<void> {
  await page.getByRole('button', { name: 'Place Table' }).click()
  await content.press('Enter')
  const table = content.getByRole('button', { name: /table component/ }).last()
  await table.click()
  await openTab(page, 'PROPERTIES')
  const yField = page.getByRole('textbox', { name: 'Y (pt)' })
  await yField.fill('120')
  await yField.press('Enter')
  await setFontFamily(page)
  await page.getByRole('button', { name: 'Configure columns' }).click()
  const dialog = page.getByRole('dialog', { name: 'Table Editor' })
  // A datalist input has the browser's `combobox` accessibility role.
  const collection = dialog.getByRole('combobox', { name: 'Root collection' })
  const beforeCollection = await revision(page)
  await collection.fill('transactions[]', { timeout: 12_000 })
  // TableEditor commits its normal, user-facing scope field on focus exit.
  // Tab both supplies that real interaction and avoids racing a synthetic blur
  // with the worker's busy-state transition.
  await collection.press('Tab')
  await waitForRevisionAdvance(page, beforeCollection)
  await expect(dialog.getByRole('combobox', { name: 'Root collection' })).toHaveValue('transactions[]', { timeout: 12_000 })
  await dialog.getByRole('button', { name: 'Add column' }).click()
  await expect(dialog.getByRole('grid', { name: 'Table columns' })).toBeVisible({ timeout: 12_000 })
  const field1 = dialog.getByRole('combobox', { name: 'Row field for column 1' })
  const beforeFirstField = await revision(page)
  await field1.fill('date', { timeout: 12_000 })
  await field1.blur()
  await waitForRevisionAdvance(page, beforeFirstField)
  for (const [column, rowField] of ['description', 'debitDisplay', 'creditDisplay', 'balanceDisplay'].entries()) {
    const index = column + 1
    const beforeColumn = await revision(page)
    await dialog.getByRole('button', { name: `Add column after column ${index}` }).click()
    await waitForRevisionAdvance(page, beforeColumn)
    const field = dialog.getByRole('combobox', { name: `Row field for column ${index + 1}` })
    const beforeField = await revision(page)
    await field.fill(rowField, { timeout: 12_000 })
    await field.blur()
    await waitForRevisionAdvance(page, beforeField)
  }
  const headers = ['Date', 'Description', 'Debit', 'Credit', 'Balance']
  for (const [index, header] of headers.entries()) {
    const input = dialog.getByRole('textbox', { name: `Header for column ${index + 1}` })
    const beforeHeader = await revision(page)
    await input.fill(header)
    await input.blur()
    await waitForRevisionAdvance(page, beforeHeader)
  }
  const beforeFooter = await revision(page)
  await dialog.getByRole('combobox', { name: 'Footer aggregate for column 5' }).selectOption('sum')
  await waitForRevisionAdvance(page, beforeFooter)
  const footerSource = dialog.getByRole('textbox', { name: 'Footer source for column 5' })
  const beforeFooterSource = await revision(page)
  await footerSource.fill('transactions.amount', { timeout: 12_000 })
  await footerSource.blur()
  await waitForRevisionAdvance(page, beforeFooterSource)
  await dialog.getByRole('button', { name: 'Close Table Editor' }).click()
}

async function authorAlternateReport(page: Page): Promise<void> {
  const header = page.getByRole('region', { name: 'Page Header', exact: true })
  const content = page.getByRole('region', { name: 'Content', exact: true })
  const footer = page.getByRole('region', { name: 'Page Footer', exact: true })
  await placeStatementText(page, header, 'ACCOUNT NOTICE', 30)
  await placeStatementText(page, footer, 'Archive copy — {{params.reportDate}}', 0, true)
  await page.getByRole('button', { name: 'Place Rectangle' }).click()
  await content.press('Enter')
  await expect(content.getByRole('button', { name: /rect component/ })).toHaveCount(1)
}

function assertNativePreflight(output: string, name: string, template: Buffer, data: Buffer, parameterBytes: Buffer): void {
  const templatePath = path.join(output, `${name}.folio`)
  const dataPath = path.join(output, `${name}.data.json`)
  const paramsPath = path.join(output, `${name}.params.json`)
  const pdfPath = path.join(output, `${name}.preflight.pdf`)
  writeFileSync(templatePath, template)
  writeFileSync(dataPath, data)
  writeFileSync(paramsPath, parameterBytes)
  try {
    execFileSync(path.join(output, 'folio'), ['validate', '-data', dataPath, '-params', paramsPath, templatePath], { cwd: goRoot, stdio: 'pipe' })
    execFileSync(path.join(output, 'folio'), ['render', '-data', dataPath, '-params', paramsPath, '-o', pdfPath, templatePath], { cwd: goRoot, stdio: 'pipe' })
  } catch (error) {
    const processError = error as { stdout?: Buffer; stderr?: Buffer; message: string }
    throw new Error(`native preflight rejected the exact browser-saved ${name} input: ${processError.stderr?.toString('utf8') || processError.stdout?.toString('utf8') || processError.message}`, { cause: error })
  }
}

async function savePreviewAndCapture(page: Page, fileName: string, output: string, name: string): Promise<BrowserWitness> {
  const download = page.waitForEvent('download')
  await page.getByRole('button', { name: 'Save As' }).click()
  const saved = await download
  const savedBytes = Buffer.from(await saved.createReadStream().then(async (stream) => {
    if (!stream) throw new Error('browser returned no saved stream')
    const chunks: Buffer[] = []
    for await (const chunk of stream) chunks.push(Buffer.from(chunk))
    return Buffer.concat(chunks)
  }))
  expect(saved.suggestedFilename()).toBe(fileName)

  await page.getByRole('button', { name: 'PREVIEW' }).click()
  await openTab(page, 'INPUTS')
  const rawParameters = page.getByRole('textbox', { name: 'Raw parameter JSON' })
  await rawParameters.fill(params.toString('utf8'))
  // Fail with the public CLI's concrete diagnostic before asking the WASM
  // adapter to map it to its intentionally bounded response schema. The
  // input bytes here are exactly those downloaded from this browser session.
  assertNativePreflight(output, name, savedBytes, sample, params)
  await page.getByRole('button', { name: 'Render local PDF' }).click()
  try {
    // Identity deliberately hashes the complete shipped font set before the
    // one Go render and PDF.js admission. This is runtime work, not a locator
    // ambiguity (all selector-facing steps above retain short timeouts).
    await expect(page.getByText('EXACT LOCAL PRODUCTION PDF')).toBeVisible({ timeout: 60_000 })
  } catch (error) {
    const proof = await captured(page)
    throw new Error(`Preview was not admitted; worker evidence: ${JSON.stringify({ requests: proof.requests.map(({ operation }) => operation), responses: proof.responses, failures: proof.failures })}`, { cause: error })
  }
  const proof = await captured(page)
  const serialized = proof.serializations.at(-1)
  const pdf = proof.renders.at(-1)
  const render = proof.requests.filter((request) => request.operation === 'render').at(-1)?.render
  if (!serialized || !pdf || !render) throw new Error('the one worker did not produce canonical, render-request, and PDF bytes')
  expect(Buffer.from(serialized)).toEqual(savedBytes)
  expect(Buffer.from(render.template)).toEqual(savedBytes)
  expect(Buffer.from(render.data)).toEqual(sample)
  expect(Buffer.from(render.params)).toEqual(params)
  await page.getByRole('button', { name: 'Render local PDF' }).click()
  await expect.poll(async () => (await captured(page)).renders.length).toBeGreaterThan(proof.renders.length)
  const repeated = (await captured(page)).renders.at(-1)
  if (!repeated) throw new Error('repeated browser render did not return PDF bytes')
  expect(Buffer.from(repeated)).toEqual(Buffer.from(pdf))
  return { template: Buffer.from(render.template), data: Buffer.from(render.data), params: Buffer.from(render.params), pdf: Buffer.from(pdf), repeatedPDF: Buffer.from(repeated), requests: (await captured(page)).requests, serialized: proof.serializations.length }
}

function fingerprint(value: Buffer): Readonly<{ length: number; sha256: string }> {
  return { length: value.length, sha256: createHash('sha256').update(value).digest('hex') }
}

function runNativeCLI(output: string, name: string): Buffer {
  const binary = path.join(output, 'folio')
  const template = path.join(output, `${name}.folio`)
  const data = path.join(output, `${name}.data.json`)
  const parameterFile = path.join(output, `${name}.params.json`)
  const pdf = path.join(output, `${name}.native-cli.pdf`)
  execFileSync(binary, ['validate', '-data', data, '-params', parameterFile, template], { cwd: goRoot, stdio: 'pipe' })
  execFileSync(binary, ['render', '-data', data, '-params', parameterFile, '-o', pdf, template], { cwd: goRoot, stdio: 'pipe' })
  return readFileSync(pdf)
}

test('fresh authored sessions close exactly through admitted Preview and native Folio', async ({ browser }, testInfo) => {
  test.setTimeout(300_000)
  const output = testInfo.outputPath('browser-native-roundtrip')
  mkdirSync(output, { recursive: true })
  execFileSync('go', ['build', '-o', path.join(output, 'folio'), './cmd/folio'], { cwd: goRoot, stdio: 'pipe' })

  const goldenContext = await browser.newContext()
  const goldenPage = await goldenContext.newPage()
  await observeOneWorker(goldenPage, 'golden-fresh-session')
  await goldenPage.goto('/')
  await expect(goldenPage.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const goldenStartup = await captured(goldenPage)
  expect(goldenStartup.requests.map(({ operation }) => operation)).toEqual(['initialize', 'serialize'])
  expect(await goldenPage.getByRole('button', { name: /component/ }).count()).toBe(0)
  await loadSample(goldenPage)
  const goldenContent = goldenPage.getByRole('region', { name: 'Content', exact: true })
  await authorStatementFraming(goldenPage)
  await bindTextToCustomer(goldenPage, goldenContent)
  await authorTableWithFooter(goldenPage, goldenContent)
  const golden = await savePreviewAndCapture(goldenPage, 'Untitled template.folio', output, 'golden')
  const goldenSession = golden.requests
  expect(goldenSession.filter(({ operation, command }) => operation === 'command' && command).length).toBeGreaterThanOrEqual(8)
  expect(goldenSession.filter(({ operation }) => operation === 'load' || operation === 'initialize').map(({ operation }) => operation)).toEqual(['initialize'])
  await goldenContext.close()

  const alternateContext = await browser.newContext()
  const alternatePage = await alternateContext.newPage()
  await observeOneWorker(alternatePage, 'alternate-fresh-session')
  await alternatePage.goto('/')
  await expect(alternatePage.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const alternateStartup = await captured(alternatePage)
  expect(alternateStartup.requests.map(({ operation }) => operation)).toEqual(['initialize', 'serialize'])
  await loadSample(alternatePage)
  const alternateContent = alternatePage.getByRole('region', { name: 'Content', exact: true })
  await bindTextToCustomer(alternatePage, alternateContent)
  await authorAlternateReport(alternatePage)
  const alternate = await savePreviewAndCapture(alternatePage, 'Untitled template.folio', output, 'alternate')
  const alternateSession = alternate.requests
  expect(alternateSession.filter(({ operation, command }) => operation === 'command' && command).length).toBeGreaterThanOrEqual(3)
  expect(alternateSession.filter(({ operation }) => operation === 'load' || operation === 'initialize').map(({ operation }) => operation)).toEqual(['initialize'])
	const historyKeys = (requests: ReadonlyArray<CapturedRequest>) => new Set(requests.map(({ sessionId, requestId }) => `${sessionId}:${requestId}`))
	const goldenHistory = historyKeys(goldenSession)
	for (const key of historyKeys(alternateSession)) expect(goldenHistory.has(key)).toBe(false)
  await alternateContext.close()

  for (const [name, witness] of Object.entries({ golden, alternate })) {
    writeFileSync(path.join(output, `${name}.folio`), witness.template)
    writeFileSync(path.join(output, `${name}.data.json`), witness.data)
    writeFileSync(path.join(output, `${name}.params.json`), witness.params)
    writeFileSync(path.join(output, `${name}.browser.pdf`), witness.pdf)
  }
  const goldenNative = runNativeCLI(output, 'golden')
  const alternateNative = runNativeCLI(output, 'alternate')
  expect(goldenNative).toEqual(golden.pdf)
  expect(alternateNative).toEqual(alternate.pdf)
  execFileSync('go', ['test', '.', '-run', '^TestBrowserAuthoredRoundTripWitness$', '-count=1'], { cwd: goRoot, env: { ...process.env, FOLIO_ROUNDTRIP_DIR: output }, stdio: 'pipe' })

  const evidence = {
    retention: 'Committed manifest contains only hashes, small raw render inputs, and command provenance. Playwright output/PDFs/native binary are disposable and ignored.',
    runtime: { platform: process.platform, arch: process.arch, node: process.version, chromium: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH ? 'explicit Chromium executable' : 'Playwright default' },
    rawRenderInputs: { dataUTF8: golden.data.toString('utf8'), paramsUTF8: golden.params.toString('utf8') },
    golden: { template: fingerprint(golden.template), data: fingerprint(golden.data), params: fingerprint(golden.params), browserPDF: fingerprint(golden.pdf), repeatedBrowserPDF: fingerprint(golden.repeatedPDF), nativePDF: fingerprint(goldenNative), serializations: golden.serialized, commands: golden.requests.filter(({ operation, command }) => operation === 'command' && command).map(({ requestId, command }) => ({ requestId, utf8: new TextDecoder().decode(new Uint8Array(command!)) })), operations: golden.requests.map(({ operation, requestId }) => ({ operation, requestId })) },
    alternate: { template: fingerprint(alternate.template), data: fingerprint(alternate.data), params: fingerprint(alternate.params), browserPDF: fingerprint(alternate.pdf), repeatedBrowserPDF: fingerprint(alternate.repeatedPDF), nativePDF: fingerprint(alternateNative), serializations: alternate.serialized, commands: alternate.requests.filter(({ operation, command }) => operation === 'command' && command).map(({ requestId, command }) => ({ requestId, utf8: new TextDecoder().decode(new Uint8Array(command!)) })), operations: alternate.requests.map(({ operation, requestId }) => ({ operation, requestId })) },
    handEditedNativeTemplate: 'folio-go/testdata/template/golden/worked-example.json (native test only; never uploaded by this browser test)',
  }
  writeFileSync(path.join(output, 'evidence.json'), `${JSON.stringify(evidence, null, 2)}\n`)
	// This is deliberately the only durable artifact. It contains no PDF,
	// browser-result path, or platform executable; future runs replace it with
	// the exact current provenance rather than retaining unbounded test output.
	const durableEvidence = path.join(root, '_bmad-output/implementation-artifacts/evidence/story-6.7-roundtrip-manifest.json')
	mkdirSync(path.dirname(durableEvidence), { recursive: true })
	writeFileSync(durableEvidence, `${JSON.stringify(evidence, null, 2)}\n`)
  await testInfo.attach('browser-native-roundtrip-evidence', { path: path.join(output, 'evidence.json'), contentType: 'application/json' })
})
