import { readFileSync } from 'node:fs'
import path from 'node:path'
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
//   chromium-1217/chrome-mac-arm64/Google Chrome for Testing.app/Contents/\
//   MacOS/Google Chrome for Testing" \
//     npx playwright test e2e/font-embed-boundary.spec.ts --reporter=list
//
// THE PINNED REVISION 1208 IS NOT USABLE ON THIS MACHINE and must not be named
// here: that directory is a 428 KB truncated download with no
// `Contents/Frameworks/...framework`, and launching it aborts in `dlopen`
// (1217 is 336 MB and complete). Do NOT run `npx playwright install chromium`
// to repair it — that archive has previously returned HTTP 400 from the
// download host. PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH stays the mechanism;
// only the revision it points at changed.
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

// AND THE BOUNDARY'S SENTENCES ARE NOT HARDCODED EITHER — for the same reason,
// and it is the sharper of the two.
//
// The first version of this harness rejected exactly one string, "The engine
// returned an invalid response". That was already stale when it was written:
// this story's own change RENAMES the production fault, so a recurrence of the
// wasm out-of-memory now reaches the panel as "The engine's response could not
// be parsed: SyntaxError …". The harness would have recorded that as an
// ordinary located refusal and passed green over the exact defect it exists
// for. Four of the five stages evaded it.
//
// So the set is READ OUT OF engine.worker.ts's own `boundarySentences` table.
// A sixth stage added there enters this rejection set in the same commit that
// adds it, with nobody having to remember. If that table is renamed or
// restructured this throws rather than silently matching nothing — an empty
// rejection set is the failure mode being designed against, not an outcome.
const workerSource = readFileSync(fileURLToPath(new URL('../src/engine.worker.ts', import.meta.url)), 'utf8')
const boundarySentences = readBoundarySentences(workerSource)

function readBoundarySentences(source: string): ReadonlyArray<string> {
  const table = /const boundarySentences[^=]*=\s*\{([\s\S]*?)\n\}/.exec(source)
  if (!table) throw new Error('engine.worker.ts no longer declares a `boundarySentences` table this harness can read; re-derive the rejection set rather than deleting the check')
  const found = [...table[1].matchAll(/:\s*(['"])((?:\\.|(?!\1).)*)\1/g)].map((match) => match[2].replace(/\\(['"])/g, '$1'))
  if (found.length < 5) throw new Error(`read ${found.length} boundary sentences from engine.worker.ts, expected at least the five stages; the extraction has gone stale`)
  return found
}

// One row per family, in the shape the story asks to be reported: embedded, or
// refused with its located text, or a boundary sentence — which is the defect.
// `outcome` is deliberately the engine's own words and not a classification of
// them: a diagnosis cannot be read out of a label this harness invented.
type Row = { family: string; outcome: string }

const boundaryAnswers = (rows: ReadonlyArray<Row>): ReadonlyArray<Row> =>
  rows.filter((row) => boundarySentences.some((sentence) => row.outcome.includes(sentence)))

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

// STORY 16.1 SPLIT THE SECOND GROUP IN TWO, AND THIS TEST NOW ASSERTS THE
// TIER RATHER THAN THE WHOLE LIST.
//
// The control used to offer exactly the committed families and nothing else.
// It now offers those PLUS families from the build-time index snapshot, whose
// bytes are fetched at the moment of a pick — and the painted list is capped
// while the count is not. So "exactly the families the catalogue declares" is
// no longer a true statement about the listbox, and the property that replaced
// it is the one the ruling actually cares about: THE LOCAL FACE TIER IS OFFERED
// IN FULL, IT SAYS THAT IT NEEDS NO DOWNLOAD, AND NOTHING IS MISSING FROM IT.
// A twenty-second committed family still cannot escape this harness.
//
// STORY 16.5 CHANGED WHAT THE NOTES SAY, BECAUSE IT CHANGED WHAT A PICK DOES. A
// family this machine already holds is USED when it is picked — the face is
// embedded and the property committed, two commands — while one it does not hold
// is INSTALLED and reaches no document at all. The local tier is the first case,
// which is why this file's every-family-embedded measurement below survives the
// split intact.
//
// STORY 16.7 REMOVED `LOCAL_NOTE` FROM THE ROW (Design Note 2): a specimen
// replaced it, so a local-tier row's own text is now just its family name.
// The LOCAL half of this measurement finds its rows by GROUP MEMBERSHIP and
// reads each one's name off `.property-option-name`, never off a note this
// group no longer carries.
//
// STORY 16.9 REMOVED THE SNAPSHOT TIER FROM THIS DROPDOWN ENTIRELY. The
// control no longer offers "those PLUS families from the build-time index
// snapshot" — that PLUS is gone, `Add fonts…` is its door now, and `WEB_NOTE`
// below is asserted ABSENT rather than present.
const WEB_NOTE = ' — install on this machine'

test('the designer offers the whole local face tier, marked as needing no download', async ({ page }) => {
  await placeAndSelectText(page)
  await page.getByRole('combobox', { name: 'Font family' }).click()
  const local = (await page.getByRole('group', { name: 'AVAILABLE LOCALLY' }).getByRole('option').locator('.property-option-name').allTextContents()).map((text) => text.trim())
  expect([...local].sort()).toEqual([...families].sort())
  // AND THE THIRD TIER IS GONE (Story 16.9): the dropdown no longer offers a
  // row for any family not on this machine at all — `Add fonts…` is the
  // only door to one now — so no option may carry the install note.
  const options = await page.getByRole('listbox', { name: 'Fonts' }).getByRole('option').allTextContents()
  const web = options.filter((text) => text.includes(WEB_NOTE))
  expect(web.length, 'Story 16.9 removed the install-tier group; no row may carry its note').toBe(0)
  await expect(page.getByRole('group', { name: 'AVAILABLE TO INSTALL' })).toHaveCount(0)
  // AND THE COUNT IS THE ADDABLE COUNT, WITH THE LIST'S STALENESS STATED. The
  // family list is a build-time snapshot — its endpoint sends no CORS header —
  // and only the typeface is fetched.
  //
  // RETIRED, PRE-EXISTING (OWNER, 2026-09-03): the standing disclosure above
  // `Add fonts…` was cut before this story started (baseline `cd7e683`), so
  // this assertion was already false when Story 16.7 opened. It is a defect
  // this story found rather than caused, and is left exactly as it was found
  // rather than repaired as a drive-by — see the story's Delivery Log.
  const disclosure = options.find((text) => text.includes('families you can add'))
  expect(disclosure, 'the control must state how many families can be added').toBeTruthy()
  expect(disclosure).toMatch(/snapshot taken on \d{4}-\d{2}-\d{2}/)
  expect(disclosure).toMatch(/changes only when the designer is released/)
})

// STORY 16.5: THIS MEASUREMENT IS UNCHANGED, AND THAT IS THE POINT OF STATING IT.
//
// Install/embed separation moves WHEN a face enters a document for a family this
// machine does not hold. The committed catalogue families are the LOCAL FACE TIER
// — they ship inside the release, so this machine already holds them — and picking
// one is FIRST USE, which embeds. So the after-state is still every one of them
// EMBEDDED and nothing less, over whatever `families` currently holds — the
// assertion below reads the list rather than a numeral, which is why it cannot
// rot into a floor the way the counts in the comments below had.
//
// WHAT THE REVISION POLL NOW WATCHES IS TWO COMMANDS RATHER THAN ONE: the embed
// and then the `fontFamily` commit. `> before` is satisfied by either, and it is
// deliberately not tightened to `+2` here — this file's subject is the WASM
// boundary answering at all, and the two-command order is asserted on the wire in
// `src/App.test.tsx` and `src/App.font-store.test.tsx` where a payload can be read.
test('every catalogue family embeds, and no pick is answered by a boundary sentence', async ({ page }) => {
  // ONE FAMILY PER BLANK DOCUMENT: a shared document would let one pick's
  // document state (every earlier family's asset and chain) confound the next
  // pick's result, and the report would no longer be one measurement per family.
  //
  // STALE NUMERAL CORRECTED (Story 16.5, D-16.R.35's rule). This said "21
  // families" in three places; Story 16.1a grew the local face tier and
  // `font-catalogue.json` has carried more than 21 entries since. The count is
  // deliberately not restated as a new numeral — `families` is the list and
  // every assertion below is written against it.
  test.setTimeout(900_000)
  const rows: Row[] = []
  for (const family of families) {
    await placeAndSelectText(page)
    const before = await currentRevision(page)
    const combobox = page.getByRole('combobox', { name: 'Font family' })
    await combobox.click()
    await combobox.fill(family)
    await page.getByRole('option', { name: family, exact: true }).click()
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

  // THE ACCEPTANCE, IN TWO HALVES, AND BOTH ARE LOAD-BEARING.
  //
  // First: NO pick may be answered by ANY boundary sentence. A boundary
  // sentence is never an engine refusal — the engine's refusals carry a
  // diagnosticCode and a located message through a different arm entirely —
  // so one appearing here is always this story's defect wearing a new stage
  // name, and must never be counted as a located refusal.
  expect(boundaryAnswers(rows).map((row) => `${row.family}: ${row.outcome}`)).toEqual([])
  const silent = rows.filter((row) => row.outcome.startsWith('TIMED OUT'))
  expect(silent.map((row) => row.family)).toEqual([])

  // Second: THE OVER-BROADNESS CONTROL. Everything above is satisfied by a
  // regression that refuses every family with a tidy located reason — the
  // browser half had no answer to that, while the Go half has had one since
  // TestStaticFaceIsStillEmbeddedAtBothDoors. Every catalogue face is static
  // by construction (each public/fonts/*/NOTICE.md asserts no `fvar`, and
  // scripts/build-wasm.mjs validates it at build time), so the measured
  // after-state of this story is EVERY committed family EMBEDDED, and nothing
  // less will do.
  expect(rows.filter((row) => row.outcome !== 'EMBEDDED').map((row) => `${row.family}: ${row.outcome}`)).toEqual([])
  expect(rows).toHaveLength(families.length)
})

// THE DISCLOSURE CHEVRON IS THE MOCKUP'S OWN SVG, MEASURED IN A REAL BROWSER,
// AND THE COMPARISON IS READ OUT OF THE MOCKUP RATHER THAN COPIED FROM IT.
//
// A first version of this button drew the chevron as a text character
// (`▲`/`▼`), justified by a canvas ink-extent measurement showing it centred
// in THIS font. That measurement was true and the conclusion was wrong: text
// glyphs centre by font metrics, a UI typeface swap could silently reopen the
// gap, and a filled triangle is the wrong shape next to the design's thin
// stroked chevron in `Font Browser.dc.html:185`. The corrected button draws
// that exact `<svg>` — one glyph, unconditionally, since the mockup's own
// `dropdownOpen` flag there gates the menu panel, not the chevron — and
// leans on `.property-inline-action`'s `display: grid; place-items: center`
// for a BOX-level centering that cannot drift with the font, the same way
// the mockup's `align-items: center` centres it in its row.
//
// The `viewBox`/`path d` this test compares against are read out of the
// mockup file, not retyped here, for the same reason `families` above is read
// out of `font-catalogue.json`: a hand-copied second source of truth goes
// stale silently, while a read that finds nothing throws.
const mockupPath = path.resolve(fileURLToPath(new URL('.', import.meta.url)), '..', '..', '_bmad-output', 'planning-artifacts', 'ux-designs', 'ux-folio-2026-08-23', 'mockups', 'Font Browser.dc.html')
const mockupSource = readFileSync(mockupPath, 'utf8')

function readMockupChevron(source: string): Readonly<{ viewBox: string; path: string }> {
  const match = /<svg width="8" height="8" viewBox="([^"]+)"[^>]*><path d="([^"]+)"><\/path><\/svg>/.exec(source)
  if (!match) throw new Error('the mockup no longer draws the 8x8 chevron this test compares the field against at Font Browser.dc.html:185; re-derive the comparison rather than deleting the check')
  return { viewBox: match[1], path: match[2] }
}
const mockupChevron = readMockupChevron(mockupSource)

test('the disclosure chevron is the mockup\'s own SVG, centred structurally in both states', async ({ page }) => {
  await placeAndSelectText(page)
  const button = page.locator('button.property-disclosure')
  const svg = button.locator('svg')

  const measure = async () => {
    const buttonBox = await button.boundingBox()
    const svgBox = await svg.boundingBox()
    if (!buttonBox || !svgBox) throw new Error('the disclosure button or its svg has no box; it is not laid out')
    const deltaX = (svgBox.x + svgBox.width / 2) - (buttonBox.x + buttonBox.width / 2)
    const deltaY = (svgBox.y + svgBox.height / 2) - (buttonBox.y + buttonBox.height / 2)
    return { deltaX, deltaY, viewBox: await svg.getAttribute('viewBox'), path: await svg.locator('path').getAttribute('d') }
  }

  // CLOSED STATE.
  await expect(button).toHaveAttribute('aria-label', 'Show fonts')
  const closed = await measure()
  expect(closed.viewBox, 'the rendered chevron must be the mockup\'s own viewBox, not a stand-in').toBe(mockupChevron.viewBox)
  expect(closed.path, 'the rendered chevron must be the mockup\'s own path, not a stand-in').toBe(mockupChevron.path)
  expect(Math.abs(closed.deltaX), `the chevron must be centred in the field; measured deltaX=${closed.deltaX}`).toBeLessThan(0.5)
  expect(Math.abs(closed.deltaY), `the chevron must be centred in the field; measured deltaY=${closed.deltaY}`).toBeLessThan(0.5)

  // OPEN STATE — SAME GLYPH, NO FLIP. The mockup draws one chevron regardless
  // of `dropdownOpen`; a two-state flip here would be inventing state the
  // design does not have. The accessible name still flips: that is
  // screen-reader state the design does not govern.
  await button.click()
  await expect(button).toHaveAttribute('aria-label', 'Hide fonts')
  const open = await measure()
  expect(open.path, 'the design draws one chevron, not two; this button must not flip shape when open').toBe(closed.path)
  expect(open.viewBox).toBe(closed.viewBox)
  expect(Math.abs(open.deltaX), `the chevron must stay centred when open; measured deltaX=${open.deltaX}`).toBeLessThan(0.5)
  expect(Math.abs(open.deltaY), `the chevron must stay centred when open; measured deltaY=${open.deltaY}`).toBeLessThan(0.5)
})

