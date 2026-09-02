import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { dirname, join } from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { assertNoVCSStamp, buildEngineWasm } from './wasm-vcs-stamp.mjs'
import { CATALOGUE_ASSET_PREFIX } from './offline-release-contract.mjs'
import { emitFontIndexModule } from './build-font-index.mjs'
// THE ONE sfnt `name`-TABLE READER (Story 16.1). This script used to carry
// its own hand-written `DataView` walk, and `src/font-catalogue.test.ts`
// carried a second one; Story 16.1 needed a THIRD, at runtime, and extracted
// the walk into one production module instead. Node 24 strips the types, so
// this plain-ESM build script imports the TypeScript directly.
import { faceCopyright as readFaceCopyright } from '../src/font-name-table.ts'

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

// THE SIX FAMILY NAMES THE HAND-WRITTEN RULES BELOW DECLARE.
//
// ⚠ THIS USED TO READ `new Set(Object.keys(assets))`, AND THAT WAS A GUARD THAT
// COULD NOT FIRE. `assets`' keys are SLOT names — `wasmExec`, `wasm`, `starter`,
// `sans`, `sansCjk`, `sansThai`, `mono`, `plexSans`, `plexSansThai` — and not one
// of them is a family name, so the half of the message promising "or over a
// family the six shipped rules already declare" was false: a catalogue entry
// declaring `Noto Sans` passed, and the browser was handed TWO `@font-face`
// rules for one family, the second silently winning. The intra-catalogue
// duplicate half worked; this half asserted nothing.
//
// IT IS A LITERAL LIST AND IT IS CHECKED AGAINST THE RULES, which is the only
// honest shape available here: the six rules must spell their families as
// literals (`src/font-binary-identity.test.ts` and `src/canvas-font-stack.test.ts`
// both parse them out of this file's TEXT, and an interpolated name is invisible
// to both). So the list cannot be derived from the template — instead the
// template is checked against the list, at the point of emission below, and a
// name that falls out of either side reds there rather than here.
const shippedFamilies = ['IBM Plex Sans', 'IBM Plex Mono', 'IBM Plex Sans Thai', 'Noto Sans', 'Noto Sans Thai', 'Noto Sans SC']

// A family name is interpolated UNESCAPED into a single-quoted CSS string
// (`font-family: '${face.family}'`). A quote closes it, a backslash escapes the
// closing quote, and a semicolon, brace or newline ends the declaration or the
// rule — so an unchecked name is not a typo, it is a way to write arbitrary CSS
// into `runtime-fonts.css`. Held to the printable, punctuation-free shape every
// real family name has, on the same reasoning the `id` field is held to
// `/^[a-z0-9]+$/`: a field that becomes syntax must be checked where it is read.
const familyShape = /^[A-Za-z0-9][A-Za-z0-9 .+-]*$/
// `directory` and `file` are `join()`ed into a filesystem path. A separator or a
// `..` segment reads a file from outside `public/fonts/`, which would ship bytes
// no NOTICE sits beside and no licence gate ever saw. One path segment, no dots
// leading, is the whole permitted vocabulary.
const segmentShape = /^[A-Za-z0-9][A-Za-z0-9._-]*$/

// THE SCRIPT VOCABULARY (Story 8.6), CLOSED AND SMALL. A face's declared
// coverage is what the designer proposes a fallback TAIL from: the shipped
// faces for the scripts the picked face does NOT cover, in this order. It is
// closed because an unrecognised script would silently propose no fallback for
// itself — the failure mode is a chain that draws tofu, and it would look like
// a correct pick. `font-catalogue.json`'s declaration is held to each binary's
// own `cmap` by `src/font-catalogue.test.ts`; here it is only held to the
// vocabulary.
const scriptFallbacks = { latin: 'Noto Sans', thai: 'Noto Sans Thai', cjk: 'Noto Sans SC' }
// AND THE THREE FALLBACK NAMES ARE HELD TO THE FAMILIES THAT ACTUALLY EXIST.
// These strings become CHAIN ENTRIES in the author's document — the engine
// resolves them against `fonts.Shipped()`'s own keys, and an entry naming a
// face nobody supplies is SKIPPED IN SILENCE (render.go's resolveRuneFace), so
// a typo or a rename here does not error anywhere: the proposed tail simply
// stops covering the script it was proposed for, and the chain draws tofu.
// That is the exact failure the closed `scripts` vocabulary above was justified
// by, reached from the other side.
//
// `shippedFamilies` is the right anchor because it is itself checked against
// the hand-written @font-face rules at the point of emission below, so this
// guard sits on a list that cannot quietly drift out of the stylesheet.
for (const [script, family] of Object.entries(scriptFallbacks)) {
  if (!shippedFamilies.includes(family)) throw new Error(`scriptFallbacks maps the script '${script}' to the face ${JSON.stringify(family)}, which shippedFamilies does not name. That string becomes a chain entry in the author's document, and the engine SKIPS an entry naming a face it was not given rather than failing — so a renamed face here would silently propose a fallback that draws nothing, and the chain would render tofu for exactly the script the fallback exists to cover.`)
}

const catalogueIds = new Set()
const catalogueFamilies = new Set(shippedFamilies)
const catalogueFaces = catalogue.map((entry) => {
  for (const field of ['id', 'directory', 'file', 'family', 'licence']) {
    if (typeof entry?.[field] !== 'string' || entry[field] === '') throw new Error(`font-catalogue.json entry is missing a ${field}: ${JSON.stringify(entry)}`)
  }
  if (!Array.isArray(entry.scripts) || entry.scripts.length === 0) throw new Error(`font-catalogue.json face ${entry.id} declares no scripts; the designer proposes a fallback tail from this list, and a face that claims nothing would be given a fallback for every script including its own`)
  for (const script of entry.scripts) {
    if (!Object.hasOwn(scriptFallbacks, script)) throw new Error(`font-catalogue.json face ${entry.id} declares the script ${JSON.stringify(script)}, which is not one of ${Object.keys(scriptFallbacks).join(', ')}; an unrecognised script proposes no fallback for itself and the chain draws tofu`)
  }
  if (new Set(entry.scripts).size !== entry.scripts.length) throw new Error(`font-catalogue.json face ${entry.id} declares a script twice`)
  // The id becomes a runtime filename stem AND the token the release manifest
  // recognises a catalogue asset by, so it is held to one shape here rather
  // than trusted to stay one.
  if (!/^[a-z0-9]+$/.test(entry.id)) throw new Error(`font-catalogue.json id ${JSON.stringify(entry.id)} is not lower-case alphanumeric`)
  if (catalogueIds.has(entry.id)) throw new Error(`font-catalogue.json declares the id ${JSON.stringify(entry.id)} twice`)
  for (const [field, value] of [['directory', entry.directory], ['file', entry.file]]) {
    if (!segmentShape.test(value) || value.includes('..')) throw new Error(`font-catalogue.json face ${entry.id} declares a ${field} ${JSON.stringify(value)} that is not a single plain path segment; it is joined into a filesystem path, so a separator or a '..' would read bytes from outside public/fonts/`)
  }
  if (!familyShape.test(entry.family)) throw new Error(`font-catalogue.json face ${entry.id} declares a family ${JSON.stringify(entry.family)} carrying a character that is CSS syntax; it is interpolated unescaped into font-family: '<name>' in the emitted stylesheet`)
  if (catalogueFamilies.has(entry.family)) throw new Error(`font-catalogue.json declares the family ${JSON.stringify(entry.family)} twice, or over a family the six shipped rules already declare`)
  catalogueIds.add(entry.id)
  catalogueFamilies.add(entry.family)
  if (!entry.file.endsWith('.ttf')) throw new Error(`font-catalogue.json face ${entry.id} is ${entry.file}; the emitted @font-face rule declares format('truetype') and the engine decodes only font/ttf and font/otf`)
  return { ...entry, filename: fingerprint(join(designerRoot, 'public', 'fonts', entry.directory, entry.file), `${CATALOGUE_ASSET_PREFIX}${entry.id}.ttf`) }
})

// THE COPYRIGHT LINE AND THE LICENCE TEXT, READ OFF COMMITTED BYTES (Story 8.6).
//
// A `.folio` that carries a face must state its terms — the engine refuses to
// load one that does not — so the designer has to be able to supply them at the
// moment of the pick. Neither is hand-copied into `font-catalogue.json`, and
// that is the whole point: a hand-copied licence is a SECOND authority on what
// the terms are, and the first time a binary is swapped the document would
// publish terms its own bytes contradict.
//
//   licenceText — the unmodified upstream `LICENSE*` file committed beside the
//   binary. It is the same file `manifest.ResolveAssets` (AD-26) already
//   requires and `src/font-catalogue.test.ts` already counts.
//
//   copyright — nameID 0 of the face's OWN `name` table, which is the one
//   statement of a face's provenance that cannot be edited from outside the
//   binary. Measured, not assumed: all 21 committed faces carry it.
//
// This is ~4 KB of licence text per DISTINCT LICENCE, not per face: the texts
// are keyed by SPDX identifier and the faces reference them, so 21 faces over
// two licences emit two copies here. (The DOCUMENT still carries one copy per
// embedded face — deliberately, because an asset passed on alone must carry its
// own terms — but there is no reason for the BUNDLE to pay for that.)
// THE COPYRIGHT READER IS NOW SHARED (Story 16.1). `sfntTableDirectory`,
// `nameTableString` and `faceCopyright` used to be written out here, by hand,
// beside a byte-identical second copy in `src/font-catalogue.test.ts`. Both are
// now `src/font-name-table.ts`, which the designer also uses AT RUNTIME to read
// nameID 0 out of a face fetched seconds earlier. One walk, three callers, no
// font-parsing dependency added.
const faceCopyright = (file) => {
  try {
    return readFaceCopyright(readFileSync(file))
  } catch (error) {
    throw new Error(`${file}: ${error instanceof Error ? error.message : String(error)}`)
  }
}

// PER FACE, NEVER PER IDENTIFIER. This was keyed by SPDX id and filled from
// whichever face reached that id first, which gave 17 of 21 faces ANOTHER
// PROJECT'S licence text — every OFL-1.1 face emitted cascadiacode's LICENSE,
// "with Reserved Font Name Cascadia Code" and all. That inverts the whole point
// of the story: a document embedding Inter would have travelled stating terms
// naming Microsoft's font.
//
// "The OFL is the OFL" is FALSE OF THE FILES, and that is the trap. The SIL OFL
// carries a per-project preamble — a copyright line and a Reserved Font Name —
// so two OFL-1.1 faces ship two DIFFERENT texts, and the identifier is a
// classification of the terms, never a substitute for them.
//
// It costs bundle bytes: 21 texts of ~4 KB rather than 2. That is the correct
// trade and it is stated rather than left to be rediscovered — a smaller bundle
// is not a reason to publish the wrong terms. The `?url` imports below are
// unaffected, so no build ASSET is added and the release cache does not grow
// by one slot on account of this module (it is 54 since Story 16.1a's batch).
const licenceTextOf = (face) => {
  const directory = join(designerRoot, 'public', 'fonts', face.directory)
  const licences = readdirSync(directory).filter((name) => name.startsWith('LICENSE'))
  // Now runs for EVERY face rather than for the first face of each identifier,
  // which is the second thing the cache was quietly costing.
  if (licences.length !== 1) throw new Error(`font-catalogue.json face ${face.id} has ${licences.length} LICENSE* files beside it (${JSON.stringify(licences)}); exactly one is the text that travels into every document embedding this face`)
  return readFileSync(join(directory, licences[0]), 'utf8').trimEnd()
}

// THE COMMITTED TIER'S `source`, INLINED FROM THE NOTICE RATHER THAN POINTING
// AT IT (D-16.R.13, DW-160).
//
// What was here until Story 16.1a:
//
//   `folio-designer/public/fonts/<dir>/<file> — see that directory's NOTICE.md
//    for the pinned upstream release and digest`
//
// Honest, and incomplete in the one way that matters: **the recipient of a
// `.folio` does not have that NOTICE.md.** The file travels alone (CAP-2), so a
// `source` that points into this repository's tree names a fact its reader
// cannot reach. The fetched tier had the mirror-image defect — a bare mutable
// branch URL — and the two disagreed in KIND, so a reader could not tell which
// tier a face came from. Both halves are corrected in one story because a field
// whose two writers use different vocabularies is uninterpretable, not merely
// inconsistent.
//
// The shape, on BOTH tiers: the **upstream project**, the **path within it**,
// and the **fetch date**. No scheme, no host, no branch name — a
// resolvable-looking string is a promise of fetchability, and a promise that
// decays reads as broken provenance when the provenance is intact. And **no
// SHA-256**: the face is stored under its digest as its asset key, and
// restating it here would put two authorities on one fact.
//
// PARSED, NEVER RETYPED. Every value comes out of the face's own NOTICE.md —
// the same rows `src/font-catalogue.test.ts` already holds each NOTICE to — so
// a provenance record and the string a document publishes cannot drift apart.
// An unparseable NOTICE throws here rather than emitting a vaguer string,
// because a silently degraded provenance line is the failure this field's whole
// correction is about.
const committedFaceSource = (face) => {
  const notice = join(designerRoot, 'public', 'fonts', face.directory, 'NOTICE.md')
  const text = readFileSync(notice, 'utf8')
  const row = (label, pattern) => {
    const match = pattern.exec(text)
    if (match === null) throw new Error(`public/fonts/${face.directory}/NOTICE.md records no ${label}, so the '${face.family}' face cannot state where it came from in the documents that embed it (D-16.R.13). Every catalogue NOTICE carries this row and src/font-catalogue.test.ts asserts it.`)
    return match[1]
  }
  // The project is written `github.com/<owner>/<repo>` in every NOTICE; the
  // host is dropped here, where the string becomes a document field, rather
  // than in the record, where it is a true statement about where to look.
  const project = row('upstream project', /^\| Upstream project \| `([^`]+)`/m).replace(/^github\.com\//, '')
  const release = row('pinned upstream release', /^\| Upstream project \| .+release `([^`]+)` \|$/m)
  const path = row('path inside the archive', /^\| Path inside the archive \| `(\S+)` \|$/m)
  const fetched = row('fetch date', /^\| Fetched \| (\d{4}-\d{2}-\d{2}) \|$/m)
  return `${project}@${release} — ${path}, fetched ${fetched}`
}

// THE TYPED CATALOGUE MODULE. `src` could not enumerate the catalogue at all
// before this: `offline-assets.ts` exports the nine named slots and nothing
// else, and the catalogue faces reached Vite only through the `url()` in the
// emitted stylesheet. They still do for the CSS; this module adds the second
// thing the pick needs — the URL to READ THE BYTES FROM, content-addressed and
// precached exactly as `runtimeAssetUrls` assets are, so the pick reads bytes
// already on the machine and fetches nothing.
//
// NO NEW BUILD ASSET. Every `?url` import below names a file the catalogue loop
// above already fingerprinted into `src/generated/runtime/`; this module names
// them, it does not create them. The release cache's slot count is unchanged
// by this module — 54 of the 64 since Story 16.1a's batch, and it was 44 before.
writeFileSync(join(generatedDir, 'font-catalogue.ts'),
  `// GENERATED by scripts/build-wasm.mjs from font-catalogue.json. Do not edit.\n`
  + catalogueFaces.map((face, index) => `import catalogueUrl${index} from './runtime/${face.filename}?url'`).join('\n')
  + `\n\nexport type CatalogueScript = ${Object.keys(scriptFallbacks).map((script) => JSON.stringify(script)).join(' | ')}\n\n`
  + `export type CatalogueFace = Readonly<{ id: string; family: string; style: string; licence: string; licenceText: string; copyright: string; source: string; scripts: ReadonlyArray<CatalogueScript>; url: string }>\n\n`
  + `// The shipped face that covers each script, in the order a proposed tail\n`
  + `// names them. A face's tail is the entries for the scripts it does NOT cover.\n`
  + `export const scriptFallbackFaces: ReadonlyArray<readonly [CatalogueScript, string]> = [${Object.entries(scriptFallbacks).map(([script, face]) => `[${JSON.stringify(script)}, ${JSON.stringify(face)}]`).join(', ')}]\n\n`
  + `export const catalogueFaces: ReadonlyArray<CatalogueFace> = [\n`
  + catalogueFaces.map((face, index) => `  { id: ${JSON.stringify(face.id)}, family: ${JSON.stringify(face.family)}, style: "Regular", licence: ${JSON.stringify(face.licence)}, licenceText: ${JSON.stringify(licenceTextOf(face))}, copyright: ${JSON.stringify(faceCopyright(join(designerRoot, 'public', 'fonts', face.directory, face.file)))}, source: ${JSON.stringify(committedFaceSource(face))}, scripts: [${face.scripts.map((script) => JSON.stringify(script)).join(', ')}], url: catalogueUrl${index} },`).join('\n')
  + `\n]\n`)

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
// THE COLLISION GUARD'S LIST IS HELD TO THE RULES IT CLAIMS TO DESCRIBE.
// `shippedFamilies` above is what stops a catalogue entry redeclaring one of
// these six; a name that drifts out of either side would make that guard silent
// again, in exactly the way `Object.keys(assets)` did. Checked here, where both
// the list and the template are in scope.
const shippedRules = `@font-face { font-family: 'IBM Plex Sans'; src: url('./runtime/${assets.plexSans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Mono'; src: url('./runtime/${assets.mono}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Sans Thai'; src: url('./runtime/${assets.plexSansThai}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans'; src: url('./runtime/${assets.sans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans Thai'; src: url('./runtime/${assets.sansThai}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'Noto Sans SC'; src: url('./runtime/${assets.sansCjk}') format('truetype'); font-display: swap; }\n`
for (const family of shippedFamilies) if (!shippedRules.includes(`font-family: '${family}'`)) throw new Error(`shippedFamilies names ${JSON.stringify(family)} and no hand-written @font-face rule declares it, so the catalogue's collision guard is describing a family that is not there`)
if (shippedRules.split('@font-face').length - 1 !== shippedFamilies.length) throw new Error(`the hand-written stylesheet emits ${shippedRules.split('@font-face').length - 1} rules and shippedFamilies names ${shippedFamilies.length}, so a rule exists that the catalogue's collision guard does not know about`)
writeFileSync(join(generatedDir, 'runtime-fonts.css'), shippedRules
  // AND THE CATALOGUE, one rule per declared face, emitted from the manifest
  // rather than written out. Same shape as the six above, deliberately: no
  // `font-weight`, no `font-style`, one static Regular per family (AC6).
  + catalogueFaces.map((face) => `@font-face { font-family: '${face.family}'; src: url('./runtime/${face.filename}') format('truetype'); font-display: swap; }\n`).join(''))

// THE FAMILY INDEX SNAPSHOT MODULE, emitted beside the catalogue module and
// from committed data alone — no network, because an offline release build is a
// shipped gate. See scripts/build-font-index.mjs for what is committed and why.
emitFontIndexModule()
