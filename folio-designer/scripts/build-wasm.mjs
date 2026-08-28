import { copyFileSync, existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const designerRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const repoRoot = join(designerRoot, '..')
const outputDir = join(designerRoot, 'public', 'wasm')
const templateDir = join(designerRoot, 'public', 'templates')
const goRoot = execFileSync('go', ['env', 'GOROOT'], { encoding: 'utf8' }).trim()
const wasmExec = [join(goRoot, 'lib', 'wasm', 'wasm_exec.js'), join(goRoot, 'misc', 'wasm', 'wasm_exec.js')].find(existsSync)

if (!wasmExec) throw new Error(`wasm_exec.js not found below ${goRoot}`)
mkdirSync(outputDir, { recursive: true })
mkdirSync(templateDir, { recursive: true })
execFileSync('go', ['build', '-o', join(outputDir, 'folio-engine.wasm'), './wasm/cmd/engine'], {
  cwd: join(repoRoot, 'folio-go'),
  env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
  stdio: 'inherit',
})
copyFileSync(wasmExec, join(outputDir, 'wasm_exec.js'))
// Whitespace makes this a deliberately non-canonical input; only the Go
// serializer determines the bytes that the application subsequently reads.
writeFileSync(join(templateDir, 'starter.folio'), Buffer.concat([Buffer.from('\n  '), readFileSync(join(repoRoot, 'folio-go', 'testdata', 'template', 'golden', 'worked-example.json'))]))
