import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { designTokenSets } from './design-tokens'

const sourceDir = path.dirname(fileURLToPath(import.meta.url))
const designPath = path.resolve(sourceDir, '../../_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md')
const cssPath = path.resolve(sourceDir, 'tokens.css')
const appCssPath = path.resolve(sourceDir, 'App.css')
const packagePath = path.resolve(sourceDir, '../package.json')
const lockPath = path.resolve(sourceDir, '../package-lock.json')

function namesFromDesign(group: string) {
  const source = fs.readFileSync(designPath, 'utf8')
  const block = source.match(new RegExp(`^${group}:\\n([\\s\\S]*?)(?=^[a-z]+:|^---$)`, 'm'))?.[1] ?? ''
  return [...block.matchAll(/^  ['"]?([\w-]+)['"]?:/gm)].map((match) => match[1]).sort()
}

describe('DESIGN.md token contract', () => {
  it('has exact token-name equality with the independent design source', () => {
    for (const [group, implemented] of Object.entries(designTokenSets)) expect([...implemented].sort()).toEqual(namesFromDesign(group))
  })

  it('routes every authoritative colour value into the consumed CSS token source', () => {
    const source = fs.readFileSync(designPath, 'utf8')
    const css = fs.readFileSync(cssPath, 'utf8')
    const colors = source.match(/^colors:\n([\s\S]*?)(?=^tints:)/m)?.[1] ?? ''
    for (const [, name, value] of colors.matchAll(/^  ([\w-]+): '([^']+)'/gm)) {
      expect(css).toContain(`--color-${name}: ${value}`)
    }
    expect(css).toContain('--type-page-eyebrow:')
    expect(css).toContain('--type-page-fine:')
    expect(css).toContain("@import './generated/runtime-fonts.css'")
  })

  it('pins package, lockfile, and strict compiler metadata independently', () => {
    const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf8'))
    const lock = JSON.parse(fs.readFileSync(lockPath, 'utf8'))
    const compiler = fs.readFileSync(path.resolve(sourceDir, '../tsconfig.app.json'), 'utf8')
    expect(pkg.engines.node).toBe('24.16.0')
    expect(pkg.dependencies.react).toBe('19.2.0')
    expect(pkg.dependencies['react-dom']).toBe('19.2.0')
    expect(pkg.devDependencies.vite).toBe('7.3.6')
    expect(lock.packages[''].dependencies.react).toBe('19.2.0')
    expect(lock.packages['node_modules/react'].version).toBe('19.2.0')
    expect(compiler).toMatch(/"strict":\s*true/)
  })

  it('keeps colour literals and curved radii inside the token definition only', () => {
    const shellCss = fs.readFileSync(appCssPath, 'utf8')
    const tokensCss = fs.readFileSync(cssPath, 'utf8')
    expect(shellCss).not.toMatch(/#[0-9a-f]{3,8}|\b(?:rgb|hsl)\(/i)
    expect(shellCss).not.toMatch(/border-radius:(?!\s*var\(--radius)/)
    expect(tokensCss).toContain('--radius-default: 0')
    expect(tokensCss).toContain('--radius-dot: 50%')
  })

  it('retains the accent grammar and permitted hard-stop page grid', () => {
    const shellCss = fs.readFileSync(appCssPath, 'utf8')
    expect(shellCss).toContain('outline: 2px solid var(--color-select)')
    expect(shellCss).toContain('background: var(--color-bind)')
    expect(shellCss).not.toMatch(/linear-gradient|conic-gradient/i)
    expect(shellCss).toContain('.page-surface')
    expect(shellCss).toContain('radial-gradient(var(--color-page-dot)')
    expect(shellCss).not.toMatch(/\.canvas-region[^}]*--color-page-/)
		expect(shellCss).toContain('.tree-item:focus-visible { outline: 2px solid var(--color-select); outline-offset: -2px; }')
  })

  it('reserves the solid danger card and square marker for a failed local render', () => {
    const shellCss = fs.readFileSync(appCssPath, 'utf8')
    expect(shellCss).toContain('.preview-failure { display: grid; grid-template-columns: auto minmax(0, 1fr) auto;')
    expect(shellCss).toContain('border: 1px solid var(--color-danger); border-left: 3px solid var(--color-danger);')
    expect(shellCss).toContain('.preview-failure-marker { color: var(--color-danger);')
    expect(shellCss).toContain('.preview-failure button:focus-visible { outline: 2px solid var(--color-select);')
  })

  it('limits the display and large numeric exception to S1', () => {
    const shellCss = fs.readFileSync(appCssPath, 'utf8')
    expect(shellCss.match(/var\(--type-display\)/g)).toHaveLength(1)
    expect(shellCss.match(/var\(--type-numeric-lg\)/g)).toHaveLength(1)
    expect(shellCss).toContain('.load-column h1')
    expect(shellCss).toContain('.load-numeric')
  })

  it('keeps actual shell foreground/background pairings above the usability floor', () => {
    const tokensCss = fs.readFileSync(cssPath, 'utf8')
    const channel = (name: string) => {
      const value = tokensCss.match(new RegExp(`--color-${name}: #(\\w{6})`))?.[1]
      if (!value) throw new Error(`missing colour token ${name}`)
      return value.match(/\w\w/g)!.map((part) => Number.parseInt(part, 16) / 255)
    }
    const luminance = (name: string) => channel(name).map((value) => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4).reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0)
    const contrast = (foreground: string, background: string) => (Math.max(luminance(foreground), luminance(background)) + 0.05) / (Math.min(luminance(foreground), luminance(background)) + 0.05)
    const shellCss = fs.readFileSync(appCssPath, 'utf8')
    expect(shellCss).toContain('background: var(--color-panel)')
    expect(shellCss).toContain('background-color: var(--color-page)')
    for (const [foreground, background] of [['ink', 'panel'], ['ink-high', 'raised'], ['page-ink', 'page'], ['page-ink-body', 'page'], ['page-ink-muted', 'page']]) {
      expect(contrast(foreground, background)).toBeGreaterThanOrEqual(4.5)
    }
  })
})
