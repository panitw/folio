import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { dirname, join } from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { assertNoVCSStamp, buildEngineWasm } from './wasm-vcs-stamp.mjs'
import { CATALOGUE_ASSET_PREFIX } from './offline-release-contract.mjs'

const designerRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const generatedDir = join(designerRoot, 'src', 'generated')
const outputDir = join(generatedDir, 'runtime')
const goRoot = execFileSync('go', ['env', 'GOROOT'], { encoding: 'utf8' }).trim()
const wasmExec = [join(goRoot, 'lib', 'wasm', 'wasm_exec.js'), join(goRoot, 'misc', 'wasm', 'wasm_exec.js')].find(existsSync)

if (!wasmExec) throw new Error(`wasm_exec.js not found below ${goRoot}`)
rmSync(outputDir, { recursive: true, force: true })
mkdirSync(outputDir, { recursive: true })
const wasmPath = join(outputDir, 'folio-engine.wasm')
// `-buildvcs=false` CLOSES ONE INPUT: THE TREE'S STATE. Go's default stamps
// `vcs.revision`, `vcs.time` and `vcs.modified` into the binary, and it derives
// `vcs.modified` from `git status`, where a single UNTRACKED file is enough — so
// without this flag the engine wasm, and every byte figure measured over the
// bundle it dominates, is a fact about whoever last wrote a scratch file into the
// checkout (D-8.5.7). With it, a clean tree, a stray untracked file and a modified
// tracked file all produce the same bytes — measured, not assumed.
//
// IT DOES NOT MAKE THE BUNDLE A FUNCTION OF THE SOURCE ALONE. The checkout's
// absolute PATH is still an input: two copies of one commit at two paths, both
// unstamped, build to different bytes, because Go embeds absolute source paths.
// That residual is measured and recorded as DW-105; `-trimpath` is its candidate
// remedy and is deliberately not applied here.
//
// Dropping `vcs.revision` is a DELIBERATE TRADE, NOT A FREE WIN: the artifact no
// longer self-identifies its commit. It is an acceptable one because the release
// manifest already carries `releaseId` and `pageId` derived from asset hashes,
// which identify the BUNDLE more precisely than a commit does, and because AD-22
// pins an exact `toolchain` directive, making the compiler behind these bytes a
// release event rather than an ambient fact.
// The flag and the argv live in wasm-vcs-stamp.mjs so the release verifier can
// run this exact build with the flag dropped, and run it twice with the flag, as
// executable proofs rather than as prose in a delivery log.
buildEngineWasm(wasmPath)
// Enforce the flag at its point of use, not only in a test: `build:wasm` is a
// dependency of `typecheck`, `test` and `build`, so dropping the flag reddens
// every designer gate and fails fast — and this must run here, while the raw wasm
// still exists, because it is deleted after fingerprinting below.
assertNoVCSStamp(readFileSync(wasmPath), 'src/generated/runtime/folio-engine.wasm')
const gluePath = join(outputDir, 'wasm-exec.js')
copyFileSync(wasmExec, gluePath)
// PDF.js support files are copied into the same generated, immutable runtime
// tree as wasm and fonts. The viewer imports their generated URL map so Vite
// emits and the release verifier precaches every local CMap/font asset.
const pdfjsRoot = join(designerRoot, 'node_modules', 'pdfjs-dist')
const copyPDFJSRuntime = (directory, files) => {
  const target = join(outputDir, directory)
  mkdirSync(target, { recursive: true })
  for (const file of files) copyFileSync(join(pdfjsRoot, directory.replace('pdfjs-', ''), file), join(target, file))
}
// Folio PDFs embed their own shipped faces; these local fallback assets cover
// the four CID collections PDF.js may need to inspect and its standard sans
// family without shipping its unneeded browser-viewer resources or licences.
copyPDFJSRuntime('pdfjs-cmaps', ['Adobe-GB1-0.bcmap', 'Adobe-CNS1-0.bcmap', 'Adobe-Japan1-0.bcmap', 'Adobe-Korea1-0.bcmap'])
copyPDFJSRuntime('pdfjs-standard_fonts', ['LiberationSans-Regular.ttf', 'LiberationSans-Bold.ttf', 'LiberationSans-Italic.ttf', 'LiberationSans-BoldItalic.ttf'])
// The starter is an empty, author-owned canvas. Whitespace keeps it a
// deliberately non-canonical input; only the Go serializer determines the
// bytes the application subsequently reads and saves.
const starterPath = join(outputDir, 'starter.folio')
writeFileSync(starterPath, Buffer.concat([Buffer.from('\n  '), readFileSync(join(designerRoot, 'public', 'templates', 'starter.folio'))]))

const fingerprint = (source, label) => {
  const bytes = readFileSync(source)
  const digest = createHash('sha256').update(bytes).digest('hex')
  const extension = label.slice(label.lastIndexOf('.'))
  const stem = label.slice(0, -extension.length)
  const target = `${stem}.${digest.slice(0, 20)}${extension}`
  copyFileSync(source, join(outputDir, target))
  return target
}

// PDF.js resolves CMaps and standard fonts by appending their canonical file
// names to these bases. Give each collection one content-addressed directory
// (rather than content-addressing individual filenames), preserving that
// contract while still making the directory immutable when any copied byte
// changes.
const directoryFingerprint = (directory) => createHash('sha256').update(readdirSync(directory).sort().map((file) => `${file}:${createHash('sha256').update(readFileSync(join(directory, file))).digest('hex')}`).join('\n')).digest('hex').slice(0, 20)
const pdfjsCMapDirectory = `pdfjs-cmaps-${directoryFingerprint(join(outputDir, 'pdfjs-cmaps'))}`
const pdfjsStandardFontDirectory = `pdfjs-standard-fonts-${directoryFingerprint(join(outputDir, 'pdfjs-standard_fonts'))}`

const assets = {
  wasmExec: fingerprint(gluePath, 'wasm-exec.js'),
  wasm: fingerprint(wasmPath, 'folio-engine.wasm'),
  starter: fingerprint(starterPath, 'starter.folio'),
  sans: fingerprint(join(designerRoot, 'public', 'fonts', 'notosans', 'NotoSans-Regular.ttf'), 'noto-sans.ttf'),
  sansCjk: fingerprint(join(designerRoot, 'public', 'fonts', 'notosanssc', 'NotoSansSC-Regular.ttf'), 'noto-sans-cjk.ttf'),
  sansThai: fingerprint(join(designerRoot, 'public', 'fonts', 'notosansthai', 'NotoSansThai-Regular.ttf'), 'noto-sans-thai.ttf'),
  mono: fingerprint(join(designerRoot, 'public', 'fonts', 'ibmplexmono', 'IBMPlexMono-Regular.ttf'), 'ibm-plex-mono.ttf'),
  plexSans: fingerprint(join(designerRoot, 'public', 'fonts', 'ibmplexsans', 'IBMPlexSans-Regular.ttf'), 'ibm-plex-sans.ttf'),
  plexSansThai: fingerprint(join(designerRoot, 'public', 'fonts', 'ibmplexsansthai', 'IBMPlexSansThai-Regular.ttf'), 'ibm-plex-sans-thai.ttf'),
}

// ───────────────────────────────────────────────────────────────────────────
// THE CATALOGUE (Story 8.5). The six slots above are HARDCODED BY NAME because
// each one is load-bearing under a name: `src/main.tsx` and
// `src/engine.worker.ts` import them out of `runtimeAssetUrls`,
// `generate-offline-release.mjs` finds three of them by URL substring to build
// the S1 payload, and `src/font-binary-identity.test.ts` pins the six-family
// join family by family. They are a vocabulary, not a list.
//
// THE CATALOGUE IS A LIST, and so it is driven by one. `font-catalogue.json` is
// the single place a face is declared; adding one is a directory, a NOTICE and
// a row there, never an edit in three places (Design Note 4). Twenty-seven
// hardcoded keys emitting a twenty-seven-line CSS string is the shape this
// deliberately does not take.
//
// THE EMITTED CSS SHAPE IS IDENTICAL to the six rules below it — one static
// Regular per family, `format('truetype')`, `font-display: swap`, and NO
// `font-weight` and NO `font-style` descriptor — so AC6's "no bold, no italic,
// no variable axis" stays observable from the generated file itself rather than
// from prose. `src/font-catalogue.test.ts` reads each committed binary's own
// `name` and `OS/2` tables and holds the catalogue to that.
//
// These faces reach Vite's asset graph through the `url()` in the emitted
// stylesheet alone — they are deliberately NOT added to `runtimeAssetUrls`,
// which exists for the assets application code names in an import.
const catalogue = JSON.parse(readFileSync(join(designerRoot, 'font-catalogue.json'), 'utf8'))
if (!Array.isArray(catalogue) || catalogue.length === 0) throw new Error('font-catalogue.json declares no catalogue faces')
const catalogueIds = new Set()
const catalogueFamilies = new Set(Object.keys(assets))
const catalogueFaces = catalogue.map((entry) => {
  for (const field of ['id', 'directory', 'file', 'family']) {
    if (typeof entry?.[field] !== 'string' || entry[field] === '') throw new Error(`font-catalogue.json entry is missing a ${field}: ${JSON.stringify(entry)}`)
  }
  // The id becomes a runtime filename stem AND the token the release manifest
  // recognises a catalogue asset by, so it is held to one shape here rather
  // than trusted to stay one.
  if (!/^[a-z0-9]+$/.test(entry.id)) throw new Error(`font-catalogue.json id ${JSON.stringify(entry.id)} is not lower-case alphanumeric`)
  if (catalogueIds.has(entry.id)) throw new Error(`font-catalogue.json declares the id ${JSON.stringify(entry.id)} twice`)
  if (catalogueFamilies.has(entry.family)) throw new Error(`font-catalogue.json declares the family ${JSON.stringify(entry.family)} twice, or over a family the six shipped rules already declare`)
  catalogueIds.add(entry.id)
  catalogueFamilies.add(entry.family)
  if (!entry.file.endsWith('.ttf')) throw new Error(`font-catalogue.json face ${entry.id} is ${entry.file}; the emitted @font-face rule declares format('truetype') and the engine decodes only font/ttf and font/otf`)
  return { ...entry, filename: fingerprint(join(designerRoot, 'public', 'fonts', entry.directory, entry.file), `${CATALOGUE_ASSET_PREFIX}${entry.id}.ttf`) }
})

rmSync(wasmPath, { force: true })
rmSync(gluePath, { force: true })
rmSync(starterPath, { force: true })
writeFileSync(join(generatedDir, 'offline-assets.ts'), Object.entries(assets)
  .map(([key, filename]) => `import ${key}Url from './runtime/${filename}?url'`).join('\n') + `\n\nexport const runtimeAssetUrls = { ${Object.keys(assets).map((key) => `${key}: ${key}Url`).join(', ')} } as const\n`)
writeFileSync(join(generatedDir, 'pdfjs-assets.ts'), `// Keep PDF.js CMaps and standard fonts in Vite's immutable asset graph.\nexport const pdfjsRuntimeAssets = import.meta.glob('./runtime/pdfjs-*/**/*', { eager: true, query: '?url', import: 'default' })\nexport const pdfjsViewerAssets = { cMapUrl: '/assets/${pdfjsCMapDirectory}/', standardFontDataUrl: '/assets/${pdfjsStandardFontDirectory}/', cMapPacked: true } as const\n`)
// SIX RULES, TWO VOCABULARIES (Story 8.4b), NOW OVER DIFFERENT FILES (8.4c).
// The first three register the DESIGN SYSTEM's family names, which is what every
// `--type-*` token in tokens.css resolves through. The second three register the
// ENGINE's own face names — the exact spellings `fonts.Shipped()` keys its FontSet
// by — so the canvas can ASK FOR THE FACE THE ENGINE MEASURED WITH by name (AD-17
// makes the browser a rasterizer only, and it cannot rasterize with the engine's
// face while it has no way to name it).
//
// Story 8.4b registered both halves over THE SAME THREE FILES, a deliberate
// interval in which the IBM Plex names were IBM Plex in name only — `IBM Plex
// Mono` was Noto Sans SC, a CJK sans with no monospacing. Story 8.4c ended it:
// SIX RULES OVER SIX FILES, each family declared from bytes that call themselves
// by that family's name.
//
// The three Noto slots STAY whatever the chrome points at. The engine half
// declares them, generate-offline-release.mjs requires `/noto-sans.`,
// `/noto-sans-thai.` and `/noto-sans-cjk.` each by name, NFR7 requires CJK
// coverage in the shipped set, and the release verifier requires the CJK face to
// remain the dominant font payload. `sansCjk` in particular now backs ONE rule
// rather than two; deleting it would throw at release-build time.
//
// ONE STATIC REGULAR PER FAMILY, and no `font-weight`/`font-style` descriptor on
// any rule, so every weight and italic the design system asks for is
// browser-synthesised from one face — exactly as it already was from the Noto
// files. Which file is behind which family name is pinned, family by family, by
// src/font-binary-identity.test.ts, which opens each file and reads its own
// `name` table: a family name is an assertion about bytes, and that is where it
// is checked rather than discovered by a designer squinting at glyphs.
writeFileSync(join(generatedDir, 'runtime-fonts.css'), `@font-face { font-family: 'IBM Plex Sans'; src: url('./runtime/${assets.plexSans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Mono'; src: url('./runtime/${assets.mono}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Sans Thai'; src: url('./runtime/${assets.plexSansThai}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans'; src: url('./runtime/${assets.sans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans Thai'; src: url('./runtime/${assets.sansThai}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans SC'; src: url('./runtime/${assets.sansCjk}') format('truetype'); font-display: swap; }\n`
  // AND THE CATALOGUE, one rule per declared face, emitted from the manifest
  // rather than written out. Same shape as the six above, deliberately: no
  // `font-weight`, no `font-style`, one static Regular per family (AC6).
  + catalogueFaces.map((face) => `@font-face { font-family: '${face.family}'; src: url('./runtime/${face.filename}') format('truetype'); font-display: swap; }\n`).join(''))
