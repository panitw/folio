import { copyFileSync, existsSync, mkdirSync, readFileSync, readdirSync, rmSync, writeFileSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { dirname, join } from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const designerRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const repoRoot = join(designerRoot, '..')
const generatedDir = join(designerRoot, 'src', 'generated')
const outputDir = join(generatedDir, 'runtime')
const goRoot = execFileSync('go', ['env', 'GOROOT'], { encoding: 'utf8' }).trim()
const wasmExec = [join(goRoot, 'lib', 'wasm', 'wasm_exec.js'), join(goRoot, 'misc', 'wasm', 'wasm_exec.js')].find(existsSync)

if (!wasmExec) throw new Error(`wasm_exec.js not found below ${goRoot}`)
rmSync(outputDir, { recursive: true, force: true })
mkdirSync(outputDir, { recursive: true })
const wasmPath = join(outputDir, 'folio-engine.wasm')
execFileSync('go', ['build', '-o', wasmPath, './wasm/cmd/engine'], {
  cwd: join(repoRoot, 'folio-go'),
  env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
  stdio: 'inherit',
})
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
}

rmSync(wasmPath, { force: true })
rmSync(gluePath, { force: true })
rmSync(starterPath, { force: true })
writeFileSync(join(generatedDir, 'offline-assets.ts'), Object.entries(assets)
  .map(([key, filename]) => `import ${key}Url from './runtime/${filename}?url'`).join('\n') + `\n\nexport const runtimeAssetUrls = { ${Object.keys(assets).map((key) => `${key}: ${key}Url`).join(', ')} } as const\n`)
writeFileSync(join(generatedDir, 'pdfjs-assets.ts'), `// Keep PDF.js CMaps and standard fonts in Vite's immutable asset graph.\nexport const pdfjsRuntimeAssets = import.meta.glob('./runtime/pdfjs-*/**/*', { eager: true, query: '?url', import: 'default' })\nexport const pdfjsViewerAssets = { cMapUrl: '/assets/${pdfjsCMapDirectory}/', standardFontDataUrl: '/assets/${pdfjsStandardFontDirectory}/', cMapPacked: true } as const\n`)
writeFileSync(join(generatedDir, 'runtime-fonts.css'), `@font-face { font-family: 'IBM Plex Sans'; src: url('./runtime/${assets.sans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Mono'; src: url('./runtime/${assets.sansCjk}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Sans Thai'; src: url('./runtime/${assets.sansThai}') format('truetype'); font-display: swap; }\n`)
