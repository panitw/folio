import { expect, test, type Locator, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// THE DOCUMENT BAR'S DOCUMENTATION LINK OPENS THE BUNDLED GUIDE IN A NEW TAB.
//
// The claim is two-sided: the guide really loads (as an emitted asset of this
// build, with the guide's own `<title>`), and the tab the author was editing is
// left EXACTLY as it was — same document, same unsaved state, same selected
// component, same undo history. Both activation paths are driven: a click and
// Enter on the focused link.
//
// The expected title is read from `docs/rendering-library.html` itself, so the
// assertion follows the page rather than restating its wording.
const here = path.dirname(fileURLToPath(import.meta.url))
const template = readFileSync(path.resolve(here, '../../folio8-go/testdata/example/first-pdf.folio8'))
const titleOf = (file: string): string => {
  const raw = /<title>([^<]*)<\/title>/i.exec(readFileSync(path.resolve(here, '../../docs', file), 'utf8'))?.[1]?.trim() ?? ''
  return raw.replaceAll('&lt;', '<').replaceAll('&gt;', '>').replaceAll('&quot;', '"').replaceAll('&#39;', '\'').replaceAll('&amp;', '&')
}
const guideTitle = titleOf('rendering-library.html')
const documentationStems = ['rendering-library', 'folio8-format', 'expression-reference']

type EditorState = Readonly<{ url: string; name: string | null; status: string | null; snapshot: string | null; components: string[]; selected: string[]; undo: boolean; redo: boolean }>

const idsOf = (locator: Locator): Promise<string[]> => locator.evaluateAll((elements) => elements.map((element) => element.getAttribute('data-component-id') ?? ''))

async function editorState(page: Page): Promise<EditorState> {
  const actions = page.getByRole('group', { name: 'Local file actions' })
  return {
    url: page.url(),
    name: await page.locator('.document-name').textContent(),
    status: await page.locator('.status-copy').textContent(),
    snapshot: await page.getByTestId('engine-snapshot').textContent(),
    components: await idsOf(page.locator('[data-component-id]')),
    selected: await idsOf(page.locator('.canvas-component-selected[data-component-id]')),
    undo: await actions.getByRole('button', { name: 'Undo', exact: true }).isEnabled(),
    redo: await actions.getByRole('button', { name: 'Redo', exact: true }).isEnabled(),
  }
}

// An opened template with an UNSAVED edit and a SELECTED component, so every
// part of the state the new tab must not disturb is actually present.
async function openEditedTemplate(page: Page): Promise<EditorState> {
  await page.addInitScript(() => { Object.assign(window, { showOpenFilePicker: undefined, showSaveFilePicker: undefined }) })
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const chooser = page.waitForEvent('filechooser')
  await page.getByRole('button', { name: 'Open local template' }).click()
  await (await chooser).setFiles({ name: 'edited.folio8', mimeType: 'application/json', buffer: template })
  await expect(page.locator('.document-name')).toHaveText('edited.folio8')

  const content = page.getByRole('region', { name: 'Content', exact: true })
  const before = await content.locator('[data-component-id]').count()
  await page.getByRole('button', { name: 'Place Rectangle' }).click()
  await content.press('Enter')
  await expect(content.locator('[data-component-id]')).toHaveCount(before + 1)
  await content.locator('[data-component-id]').last().click()
  await expect(page.locator('.canvas-component-selected[data-component-id]')).toHaveCount(1)
  await expect(page.locator('.status-copy')).toHaveText('Unsaved local changes')

  const state = await editorState(page)
  // Non-vacuous: the state compared afterwards really carries each fact.
  expect(state.status).toBe('Unsaved local changes')
  expect(state.selected).toHaveLength(1)
  expect(state.undo).toBe(true)
  return state
}

const documentationLink = (page: Page): Locator => page.getByRole('banner', { name: 'Document bar' })
  .getByRole('group', { name: 'Documentation' })
  .getByRole('link', { name: 'Rendering library documentation' })

async function expectGuide(guide: Page): Promise<void> {
  await guide.waitForLoadState('domcontentloaded')
  expect(new URL(guide.url()).pathname).toMatch(/\/assets\/rendering-library-[a-f0-9]{20}\.html$/)
  await expect(guide).toHaveTitle(guideTitle)
}

test.beforeEach(async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 800 })
})

test('clicking the documentation link opens the guide in a new tab and leaves unsaved edits, selection and undo untouched', async ({ page }) => {
  const before = await openEditedTemplate(page)
  const popup = page.waitForEvent('popup')
  await documentationLink(page).click()
  const guide = await popup
  await expectGuide(guide)
  expect(await editorState(page)).toEqual(before)
  await guide.close()
  expect(await editorState(page)).toEqual(before)
})

test('pressing Enter on the focused documentation link opens the guide in a new tab and leaves the editor untouched', async ({ page }) => {
  const before = await openEditedTemplate(page)
  const link = documentationLink(page)
  await link.focus()
  await expect(link).toBeFocused()
  const popup = page.waitForEvent('popup')
  await page.keyboard.press('Enter')
  const guide = await popup
  await expectGuide(guide)
  expect(await editorState(page)).toEqual(before)
  await guide.close()
  expect(await editorState(page)).toEqual(before)
})

// THE BUNDLED COPIES LINK TO EACH OTHER BY THEIR FINGERPRINTED NAMES. Every link
// in the guide that targets one of the three pages must name an emitted file of
// this build, and none may still spell a canonical `docs/` name — that would be
// a link into a file the application does not serve.
test('the bundled guide links to the bundled format and expression references by emitted names', async ({ page }) => {
  await page.goto('/')
  const href = await documentationLink(page).getAttribute('href')
  expect(href).not.toBeNull()
  const guide = await page.context().newPage()
  await guide.goto(new URL(href!, page.url()).toString())
  await expectGuide(guide)
  const hrefs = await guide.locator('a[href]').evaluateAll((anchors) => anchors.map((anchor) => anchor.getAttribute('href') ?? ''))
  const crossPage = hrefs.filter((target) => documentationStems.some((stem) => target.includes(stem)))
  expect(crossPage.length, 'the guide must link to the format reference').toBeGreaterThan(0)
  for (const target of crossPage) {
    expect(target, 'a cross-page link must be rewritten to a fingerprinted name').toMatch(/^(?:rendering-library|folio8-format|expression-reference)-[a-f0-9]{20}\.html(?:#.*)?$/)
    const response = await guide.request.get(new URL(target, guide.url()).toString())
    expect(response.ok(), `${target} must be an emitted file`).toBe(true)
    expect(response.headers()['content-type']).toMatch(/text\/html/)
  }
  const format = crossPage.find((target) => target.startsWith('folio8-format-'))
  expect(format, 'the guide must link to the format reference').toBeDefined()
  await guide.goto(new URL(format!, guide.url()).toString())
  await expect(guide).toHaveTitle(titleOf('folio8-format.html'))
})
