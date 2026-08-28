import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { LoadScreen } from './LoadScreen'
import type { S1Payload } from './release-payload'

const payload: S1Payload = { version: 1, releaseId: 'a'.repeat(64), pageId: 'b'.repeat(64), unit: 'MiB', decimals: 2, cachedBytes: 100, assetCount: 10, cacheAssets: ['/index.html', '/engine', '/latin', '/thai', '/cjk', '/a', '/b', '/c', '/d', '/e'].map((assetUrl) => ({ assetUrl, bytes: 10 })), rows: [
  { id: 'engine', label: 'Engine', delivery: 'cached-asset', assetUrl: '/engine', bytes: 10, sha256: 'a'.repeat(64) }, { id: 'latin-font', label: 'Latin font', delivery: 'cached-asset', assetUrl: '/latin', bytes: 10, sha256: 'a'.repeat(64) }, { id: 'thai-font', label: 'Thai font', delivery: 'cached-asset', assetUrl: '/thai', bytes: 10, sha256: 'a'.repeat(64) }, { id: 'cjk-font', label: 'CJK font', delivery: 'cached-asset', assetUrl: '/cjk', bytes: 10, sha256: 'a'.repeat(64) }, { id: 'thai-dictionary', label: 'Thai dictionary', delivery: 'embedded-in-engine', assetUrl: '/engine', bytes: 5, sha256: 'a'.repeat(64) },
] }
describe('honest first-run load screen', () => {
  it('uses named live numeric progress and never invents a second Thai request', () => {
    render(<LoadScreen lifecycle={{ state: 'caching', cacheReady: false, verifiedAssetUrls: ['/engine'] }} payload={payload} engineState="waiting" onRetry={vi.fn()} />)
    expect(screen.getByRole('status', { name: 'Offline preparation status' })).toHaveTextContent('Caching verified assets')
    expect(screen.getByRole('progressbar', { name: 'Verified offline cache progress' })).toHaveAttribute('aria-valuenow', '10')
    expect(screen.getByLabelText('Offline payload manifest')).toHaveTextContent('embedded in engine; no second request')
    expect(screen.queryByText(/spinner/i)).not.toBeInTheDocument()
  })
  it('uses active and failed words before colour and never calls an unmarked cache complete', () => {
    render(<LoadScreen lifecycle={{ state: 'caching', cacheReady: false, verifiedAssetUrls: ['/engine', '/latin', '/thai', '/cjk', '/a', '/b', '/c', '/d', '/e'], activeAssetUrl: '/index.html' }} payload={payload} engineState="waiting" onRetry={vi.fn()} />)
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '90')
    expect(screen.getByLabelText('Offline payload manifest')).toHaveTextContent('✓')
    expect(screen.getByText(/10 release assets/)).toBeInTheDocument()
  })
  it('shows a keyboard retry only for a bounded failure', () => {
    const retry = vi.fn(); render(<LoadScreen lifecycle={{ state: 'unavailable', cacheReady: false, verifiedAssetUrls: [] }} payload={payload} engineState="waiting" onRetry={retry} />)
    expect(screen.getByRole('button', { name: 'Retry preparation' })).toHaveFocus()
  })
})
