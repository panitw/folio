// This state vocabulary is intentionally independent of PDF.js and document
// commands. A preview can become stale synchronously, while the worker keeps
// its single FIFO operation in flight and the UI discards obsolete results.
export const PREVIEW_DEBOUNCE_MS = 250

export type PreviewFreshness = 'idle' | 'checking' | 'debouncing' | 'rendering' | 'current' | 'stale' | 'error'
export type StaleReason = 'inputs-changed' | 'render-failed'

export const staleCopy = (reason: StaleReason): string => reason === 'inputs-changed' ? 'STALE — inputs changed' : 'STALE — latest local render failed'

// The worker is FIFO and cannot pull a message back once posted.  Keep one
// request active and retain at most the newest replacement locally; when the
// active request settles, only that replacement is admitted.  This is the
// boundary that prevents manual clicks and debounce expiry from building an
// unbounded queue of abandoned worker messages.
export class PreviewWorkScheduler {
  #active = false
  #pending: (() => Promise<void>) | undefined

  submit(job: () => Promise<void>): void {
    if (this.#active) { this.#pending = job; return }
    this.#start(job)
  }

  clear(): void { this.#pending = undefined }

  get active(): boolean { return this.#active }
  get hasPending(): boolean { return this.#pending !== undefined }

  #start(job: () => Promise<void>): void {
    this.#active = true
    void job().finally(() => {
      const next = this.#pending
      this.#pending = undefined
      if (next) this.#start(next)
      else this.#active = false
    })
  }
}

export function canInstallPreview(candidate: Readonly<{ token: number; generation: number; revision: number; identity: string }>, authority: Readonly<{ token: number; generation: number; revision: number; identity: string; mode: 'design' | 'preview' }>): boolean {
  return authority.mode === 'preview' && candidate.token === authority.token && candidate.generation === authority.generation && candidate.revision === authority.revision && candidate.identity === authority.identity
}
