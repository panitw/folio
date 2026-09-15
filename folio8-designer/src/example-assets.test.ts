import { describe, expect, it } from 'vitest'
import { exampleAssets } from './generated/example-assets'

describe('exampleAssets', () => {
  it('lists the Invoice example with its template, sample and thumbnail URLs', () => {
    expect(exampleAssets.map((example) => example.id)).toEqual(['invoice'])
    const [invoice] = exampleAssets
    expect(invoice.template).toMatch(/invoice\.[a-f0-9]{20}\.folio/)
    expect(invoice.sample).toMatch(/invoice\.sample\.[a-f0-9]{20}\.json/)
    expect(invoice.thumbnail).toMatch(/invoice\.thumbnail\.[a-f0-9]{20}\.png/)
    expect(new Set([invoice.template, invoice.sample, invoice.thumbnail]).size).toBe(3)
  })
})
