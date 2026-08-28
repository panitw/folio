import type { EngineDiagnostic, EngineError } from '../engine-protocol'

export type DiagnosticLocation = Readonly<{ elementId?: string; dataPath?: string }>

function diagnosticLocation(diagnostic: DiagnosticLocation): string | undefined {
  return [diagnostic.elementId, diagnostic.dataPath].filter(Boolean).join(' · ') || undefined
}

// Presentation only: diagnostics arrive from the closed Go/wasm producer
// contract. This module deliberately has no code registry or template parser.
export function PreviewDiagnostics({ diagnostics, dismissed, onDismiss, onLocate }: { diagnostics: ReadonlyArray<EngineDiagnostic>; dismissed: ReadonlySet<string>; onDismiss: (key: string) => void; onLocate: (location: DiagnosticLocation) => void }) {
  const visible = diagnostics.map((diagnostic, index) => ({ diagnostic, key: `${index}:${diagnostic.severity}:${diagnostic.code}:${diagnostic.elementId}:${diagnostic.dataPath}:${diagnostic.message}` })).filter(({ key }) => !dismissed.has(key))
  if (!visible.length) return null
  // This text describes the admitted render generation, not the mutable local
  // dismissal view. Keeping it stable prevents every dismiss click from
  // re-announcing the same producer facts.
  const announcement = `${diagnostics.length} render ${diagnostics.length === 1 ? 'warning' : 'warnings'} available${diagnostics.length ? `: ${diagnostics.map((diagnostic) => diagnostic.code).join(', ')}` : ''}.`
  return <section className="diagnostic-list" aria-label="Render diagnostics"><p className="diagnostic-announcement" role="status" aria-live="polite" aria-atomic="true">{announcement}</p>{visible.map(({ diagnostic, key }) => {
    const location = diagnosticLocation(diagnostic)
    return <article className="diagnostic-card" key={key}><span className="diagnostic-triangle" aria-hidden="true">▲</span><div><p className="diagnostic-code">{diagnostic.severity} · {diagnostic.code}</p><p>{diagnostic.message}</p>{location && <code className="diagnostic-location">{location}</code>}{diagnostic.elementId && <button type="button" className="diagnostic-locate" onClick={() => onLocate(diagnostic)}>Locate in Design</button>}</div><button type="button" className="diagnostic-dismiss" aria-label={`Dismiss ${diagnostic.code} diagnostic`} onClick={() => onDismiss(key)}>Dismiss</button></article>
  })}</section>
}

export function PreviewFailure({ error, onReturn }: { error: EngineError; onReturn: () => void }) {
  const location = diagnosticLocation(error)
  return <section className="preview-failure" aria-label="Local render failure" role="status" aria-live="assertive" aria-atomic="true"><p>Local PDF render failed: {error.message}</p>{location && <code className="diagnostic-location">{location}</code>}<button type="button" className="file-button" onClick={onReturn}>Return to Design</button></section>
}
