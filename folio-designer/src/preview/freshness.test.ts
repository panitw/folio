import { describe, expect, it } from 'vitest'
import { canInstallPreview, PREVIEW_DEBOUNCE_MS, PreviewWorkScheduler, staleCopy } from './freshness'

describe('preview freshness authority', () => {
  it('documents the short debounce and never installs an outdated authority tuple', () => {
    expect(PREVIEW_DEBOUNCE_MS).toBe(250)
    const authority = { token: 4, generation: 7, revision: 11, identity: 'a'.repeat(64), mode: 'preview' as const }
    expect(canInstallPreview({ token: 4, generation: 7, revision: 11, identity: 'a'.repeat(64) }, authority)).toBe(true)
    expect(canInstallPreview({ token: 3, generation: 7, revision: 11, identity: 'a'.repeat(64) }, authority)).toBe(false)
    expect(canInstallPreview({ token: 4, generation: 8, revision: 11, identity: 'a'.repeat(64) }, authority)).toBe(false)
    expect(canInstallPreview({ token: 4, generation: 7, revision: 11, identity: 'b'.repeat(64) }, authority)).toBe(false)
  })

  it('uses permanent truthful stale language', () => {
    expect(staleCopy('inputs-changed')).toBe('STALE — inputs changed')
    expect(staleCopy('render-failed')).toBe('STALE — latest local render failed')
  })

  it('keeps only an active worker operation and its newest replacement', async () => {
    const scheduler = new PreviewWorkScheduler()
    const order: string[] = []
    let release!: () => void
    const active = () => new Promise<void>((resolve) => { order.push('active'); release = resolve })
    scheduler.submit(active)
    scheduler.submit(async () => { order.push('obsolete') })
    scheduler.submit(async () => { order.push('newest') })
    expect(scheduler.active).toBe(true)
    expect(scheduler.hasPending).toBe(true)
    expect(order).toEqual(['active'])
    release()
    await Promise.resolve()
    await Promise.resolve()
    expect(order).toEqual(['active', 'newest'])
    expect(scheduler.active).toBe(false)
  })
})
