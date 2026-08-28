import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const sourceRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const production = fs.readdirSync(sourceRoot, { recursive: true }).filter((entry): entry is string => typeof entry === 'string' && /\.(?:ts|tsx)$/.test(entry) && !entry.includes('.test.')).map((entry) => ({ name: String(entry), source: fs.readFileSync(path.join(sourceRoot, String(entry)), 'utf8') }))

function forbiddenFileState(files: ReadonlyArray<Readonly<{ name: string; source: string }>>): string[] {
  return files.filter(({ source }) => /\b(?:localStorage|sessionStorage|indexedDB|caches\.open|showDirectoryPicker)\b|navigator\.storage\.getDirectory\s*\(|\b(?:cloud|sync|recent files|collaborator|account)\b/i.test(source)).map(({ name }) => name)
}

function extraCapabilityChecks(files: ReadonlyArray<Readonly<{ name: string; source: string }>>): string[] {
  return files.filter(({ name, source }) => name !== 'file/capability.ts' && /(?:if|\?)\s*\([^\n]*show(?:Open|Save)FilePicker|window\.show(?:Open|Save)FilePicker/.test(source)).map(({ name }) => name)
}

function opaqueByteRewrites(files: ReadonlyArray<Readonly<{ name: string; source: string }>>): string[] {
  return files.filter(({ source }) => /\bTextDecoder\b|replace\([^)]*\\n[^)]*\\r\\n|new\s+Blob\s*\(\s*\[\s*(?:['"`]|(?:text|decoded|content)\b)/.test(source)).map(({ name }) => name)
}

function fileNetworkCalls(files: ReadonlyArray<Readonly<{ name: string; source: string }>>): string[] {
  return files.filter(({ name, source }) => name.startsWith('file/') && /\bfetch\s*\(/.test(source)).map(({ name }) => name)
}

describe('local file policy structure', () => {
  it('has production witnesses while keeping capability selection in one composition seam and no durable/browser-cloud fiction', () => {
    expect(production.some(({ name }) => name === 'file/capability.ts')).toBe(true)
    expect(production.some(({ name }) => name === 'file/file-system-access.ts')).toBe(true)
    expect(production.some(({ name }) => name === 'file/input-download.ts')).toBe(true)
    expect(extraCapabilityChecks(production)).toEqual([])
    expect(forbiddenFileState(production)).toEqual([])
    expect(opaqueByteRewrites(production)).toEqual([])
    expect(fileNetworkCalls(production)).toEqual([])
  })

  it('red-proves an accidental second capability branch and prohibited durable file state', () => {
    expect(extraCapabilityChecks([{ name: 'App.tsx', source: 'if (window.showSaveFilePicker) save()' }])).toEqual(['App.tsx'])
    expect(forbiddenFileState([{ name: 'file/cache.ts', source: 'localStorage.setItem("template", "bytes")' }])).toEqual(['file/cache.ts'])
    expect(forbiddenFileState([{ name: 'file/opfs.ts', source: 'navigator.storage.getDirectory()' }])).toEqual(['file/opfs.ts'])
    expect(fileNetworkCalls([{ name: 'file/upload.ts', source: 'fetch("/template", { method: "POST" })' }])).toEqual(['file/upload.ts'])
    expect(opaqueByteRewrites([{ name: 'file/rewrite.ts', source: 'const text = new TextDecoder().decode(bytes); new Blob([text.replace(/\\n/g, "\\r\\n")])' }])).toEqual(['file/rewrite.ts'])
  })
})
