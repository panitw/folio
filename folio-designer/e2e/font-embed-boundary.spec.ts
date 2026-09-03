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
const LOCAL_NOTE = ' — use it, already on this machine'
const WEB_NOTE = ' — install on this machine'

test('the designer offers the whole local face tier, marked as needing no download', async ({ page }) => {
  await placeAndSelectText(page)
  await page.getByRole('combobox', { name: 'Font family' }).click()
  const options = await page.getByRole('listbox', { name: 'Fonts' }).getByRole('option').allTextContents()
  const local = options.filter((text) => text.includes(LOCAL_NOTE)).map((text) => text.replace(LOCAL_NOTE, '').trim())
  expect([...local].sort()).toEqual([...families].sort())
  // AND THE SECOND TIER IS REALLY THERE, distinct from the first: rows that
  // carry the install note are families whose bytes are not on this machine.
  const web = options.filter((text) => text.includes(WEB_NOTE) && !text.includes(LOCAL_NOTE))
  expect(web.length, 'the snapshot tier must contribute rows of its own').toBeGreaterThan(0)
  // AND THE COUNT IS THE ADDABLE COUNT, WITH THE LIST'S STALENESS STATED. The
  // family list is a build-time snapshot — its endpoint sends no CORS header —
  // and only the typeface is fetched.
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
    await page.getByRole('option', { name: `${family}${LOCAL_NOTE}` }).click()
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
