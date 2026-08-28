import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

describe('Preview local-only authority boundary', () => {
  it('contains no browser viewer, network, storage, or object-URL fallback', () => {
    const source = readFileSync('src/preview/pdf-viewer.tsx', 'utf8')
    const forbidden = ['<iframe', '<embed', 'createObjectURL', 'revokeObjectURL', 'fetch(', 'XMLHttpRequest', 'WebSocket', 'EventSource', 'localStorage', 'sessionStorage', 'indexedDB', 'http://', 'https://']
    for (const token of forbidden) expect(source).not.toContain(token)
    expect(source).toContain('disableAutoFetch: true')
    expect(source).toContain('useWorkerFetch: false')
    expect(source).toContain('isEvalSupported: false')
  })
})
