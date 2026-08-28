import { copyFileSync, existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
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
// Whitespace makes this a deliberately non-canonical input; only the Go
// serializer determines the bytes that the application subsequently reads.
const starterPath = join(outputDir, 'starter.folio')
writeFileSync(starterPath, Buffer.concat([Buffer.from('\n  '), readFileSync(join(repoRoot, 'folio-go', 'testdata', 'template', 'golden', 'worked-example.json'))]))

const fingerprint = (source, label) => {
  const bytes = readFileSync(source)
  const digest = createHash('sha256').update(bytes).digest('hex')
  const extension = label.slice(label.lastIndexOf('.'))
  const stem = label.slice(0, -extension.length)
  const target = `${stem}.${digest.slice(0, 20)}${extension}`
  copyFileSync(source, join(outputDir, target))
  return target
}

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
writeFileSync(join(generatedDir, 'runtime-fonts.css'), `@font-face { font-family: 'IBM Plex Sans'; src: url('./runtime/${assets.sans}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Mono'; src: url('./runtime/${assets.sansCjk}') format('truetype'); font-display: swap; }\n@font-face { font-family: 'IBM Plex Sans Thai'; src: url('./runtime/${assets.sansThai}') format('truetype'); font-display: swap; }\n`)
