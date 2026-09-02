import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { expect, test } from '@playwright/test'

// STORY 16.0 — THE PROBE HARNESS, AND IT IS EXECUTED.
//
// Every other spec in this directory carries the D-000.4 banner: compiled by
// `tsc --noEmit`, never run in a browser. THIS ONE IS DIFFERENT, and the
// difference is the whole point. The defect it exists for — `embedFontFamily`
// throwing at the WASM/worker boundary and reporting only "The engine returned
// an invalid response" — reached the owner PRECISELY BECAUSE the e2e suite is
// compile-checked and never executed. A verification pass that proved this
// story without a real browser would reproduce the conditions that produced
// the bug. Run it with:
//
//   PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH="$HOME/Library/Caches/ms-playwright/\
//   chromium-1208/chrome-mac-arm64/Google Chrome for Testing.app/Contents/\
//   MacOS/Google Chrome for Testing" \
//     npx playwright test e2e/font-embed-boundary.spec.ts --reporter=list
//
// THE CATALOGUE IS NEVER HARDCODED HERE, in either of the two places it could
// have been. The family list is read from `font-catalogue.json` — the build's
// own input, not a copy of it — and it is ALSO enumerated out of the running
// application's own listbox and compared against that file. A twenty-second
// family added to the catalogue therefore cannot escape this harness: it
// appears in the file, it appears in the listbox, and it gets picked. The
// generated `src/generated/font-catalogue.ts` cannot be imported here because
// it imports `.ttf?url` modules that only Vite resolves.
type CatalogueRow = Readonly<{ id: string; family: string }>
const cataloguePath = fileURLToPath(new URL('../font-catalogue.json', import.meta.url))
const catalogue = JSON.parse(readFileSync(cataloguePath, 'utf8')) as ReadonlyArray<CatalogueRow>
const families = catalogue.map((row) => row.family)

// One row per family, in the shape the story asks to be reported: embedded, or
// refused with its located text, or the bare protocol failure that is the
// defect. `outcome` is deliberately the engine's own sentence and not a
// classification of it — a diagnosis cannot be read out of a label this
// harness invented.
type Row = { family: string; outcome: string }

const PROTOCOL_FAILURE = 'The engine returned an invalid response'

async function placeAndSelectText(page: import('@playwright/test').Page) {
  await page.goto('/')
  await expect(page.getByTestId('engine-snapshot')).toHaveText(/GO SNAPSHOT · REVISION 1/)
  const content = page.getByRole('region', { name: 'Content', exact: true })
  await page.getByRole('button', { name: 'Place Text' }).click()
  await content.press('Enter')
  await content.getByRole('button', { name: /text component/ }).last().click()
  await expect(page.getByRole('combobox', { name: 'Font family' })).toBeVisible()
}

async function currentRevision(page: import('@playwright/test').Page): Promise<number> {
  const label = await page.getByTestId('engine-snapshot').textContent()
  const match = /REVISION (\d+)/.exec(label ?? '')
  return match ? Number(match[1]) : -1
}

test('the designer offers exactly the families the catalogue declares', async ({ page }) => {
  await placeAndSelectText(page)
  await page.getByRole('combobox', { name: 'Font family' }).click()
  const options = await page.getByRole('listbox', { name: 'Fonts' }).getByRole('option').allTextContents()
  const offered = options.filter((text) => text.includes('— add to document')).map((text) => text.replace(' — add to document', '').trim())
  expect([...offered].sort()).toEqual([...families].sort())
})

test('every catalogue family is embedded or located-refused, and none returns the bare protocol failure', async ({ page }) => {
  // 21 families, each on its OWN blank document: a shared document would let
  // one pick's document state (twenty-one embedded assets, twenty-one chains)
  // confound the next pick's result, and the report would no longer be one
  // measurement per family.
  test.setTimeout(900_000)
  const rows: Row[] = []
  for (const family of families) {
    await placeAndSelectText(page)
    const before = await currentRevision(page)
    const combobox = page.getByRole('combobox', { name: 'Font family' })
    await combobox.click()
    await combobox.fill(family)
    await page.getByRole('option', { name: `${family} — add to document` }).click()
    let outcome = 'TIMED OUT with no engine answer'
    try {
      await expect
        .poll(async () => {
          const alerts = await page.locator('p.property-error').allTextContents()
          if (alerts.length > 0) return `REFUSED: ${alerts[0]}`
          return (await currentRevision(page)) > before ? 'EMBEDDED' : 'pending'
        }, { timeout: 120_000, intervals: [250] })
        .not.toBe('pending')
      const alerts = await page.locator('p.property-error').allTextContents()
      outcome = alerts.length > 0 ? `REFUSED: ${alerts[0]}` : 'EMBEDDED'
    } catch {
      // Left as the timeout sentence above; a family that never answers is a
      // finding, not a reason to abandon the other twenty.
    }
    rows.push({ family, outcome })
    console.log(`${family.padEnd(24)} ${outcome}`)
  }
  console.log('\nPER-FAMILY TABLE')
  for (const row of rows) console.log(`| ${row.family} | ${row.outcome} |`)

  // THE ACCEPTANCE. Each family is embedded or refused with a located reason,
  // and none produces the boundary's own least-useful sentence.
  const swallowed = rows.filter((row) => row.outcome.includes(PROTOCOL_FAILURE))
  expect(swallowed.map((row) => `${row.family}: ${row.outcome}`)).toEqual([])
  const silent = rows.filter((row) => row.outcome.startsWith('TIMED OUT'))
  expect(silent.map((row) => row.family)).toEqual([])
  expect(rows).toHaveLength(families.length)
})
