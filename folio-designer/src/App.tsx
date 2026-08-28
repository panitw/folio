import './App.css'
import { useEffect, useRef, useState, type CSSProperties } from 'react'
import type { EngineClient } from './engine-client'
import type { CanvasProjection, EngineSnapshot } from './engine-protocol'
import type { OfflineLifecycleState } from './offline-lifecycle'
import type { OfflineLifecycle } from './offline-lifecycle'
import type { S1Payload } from './release-payload'
import { LoadScreen } from './LoadScreen'
import { isFileAccessCancelled, type FileAccess, type FileTarget } from './file/file-access'
import { pageSetupCommand } from './page-setup-command'

const paletteItems = ['Text', 'Image', 'Table', 'Line', 'Rectangle']

function Icon({ name }: { name: 'open' | 'save' }) {
  return <svg aria-hidden="true" className="icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.25"><path d={name === 'open' ? 'M2 5.5h4l1.2-2h6.8v9H2z M2 5.5h12' : 'M3 2h8l2 2v10H3z M5 2v4h6V2 M5 12h6'} /></svg>
}

type AppProps = Readonly<{ engine?: EngineClient; fileAccess?: FileAccess; initialSnapshot?: EngineSnapshot; blankBytes?: ArrayBuffer; initializationError?: string; offlineState?: OfflineLifecycleState; loadState?: OfflineLifecycle; payload?: S1Payload; engineState?: 'waiting' | 'starting' | 'failed'; onRetry?: () => void }>

export default function App({ engine, fileAccess, initialSnapshot, blankBytes, initializationError, offlineState = 'unavailable', loadState, payload, engineState = 'waiting', onRetry = () => undefined }: AppProps = {}) {
  const [snapshot, setSnapshot] = useState(initialSnapshot)
  const [commitError, setCommitError] = useState<string>()
  const [fileError, setFileError] = useState<string>()
  const [fileStatus, setFileStatus] = useState<string>()
  const [fileBusy, setFileBusy] = useState(false)
  const [title, setTitle] = useState('Untitled template')
  const [target, setTarget] = useState<FileTarget>()
  const [savedRevision, setSavedRevision] = useState<number>()
  const [zoom, setZoom] = useState(1)
  const [gridVisible, setGridVisible] = useState(true)
  const [snapEnabled, setSnapEnabled] = useState(true)
  const [preset, setPreset] = useState<string>(initialSnapshot?.canvas?.preset ?? 'A4')
  const [orientation, setOrientation] = useState<string>(initialSnapshot?.canvas?.orientation ?? 'portrait')
  const [draft, setDraft] = useState(() => draftFor(initialSnapshot?.canvas))
  const snapshotRef = useRef(snapshot)
  const saveInFlight = useRef(false)
  const draftGeneration = useRef(0)
  const canvas = snapshot?.canvas
  const applyPageSetup = async () => {
    if (!engine || !canvas || fileBusy) return
    setCommitError(undefined)
    try {
      const requestGeneration = draftGeneration.current
      const result = await engine.request('command', pageSetupCommand(preset, orientation, preset === 'custom' ? draft.width : '0', preset === 'custom' ? draft.height : '0', draft))
      setCurrentSnapshot(result.snapshot, draftGeneration.current !== requestGeneration)
    } catch (error) { setCommitError(pageSetupDiagnostic(error)) }
  }

  const setCurrentSnapshot = (next: EngineSnapshot | undefined, keepNewerDraft = false) => { snapshotRef.current = next; setSnapshot(next); if (next?.canvas) { setPreset(next.canvas.preset); setOrientation(next.canvas.orientation); if (!keepNewerDraft) setDraft(draftFor(next.canvas)) } }
  const updateDraft = (key: keyof Draft, value: string) => { draftGeneration.current++; setDraft((current) => ({ ...current, [key]: value })) }
  const announceFailure = (message: string) => { setFileStatus(undefined); setFileError(message) }
  const open = async () => {
    if (!engine || !fileAccess || fileBusy) return
    setFileBusy(true); setFileError(undefined); setFileStatus('Opening local file…')
    try {
      const opened = await fileAccess.open()
      const loaded = await engine.request('load', opened.bytes)
      const canonical = await engine.request('serialize')
      if (!canonical.bytes) throw new Error('Local file could not be serialized')
      const inputWasCanonical = equalBytes(opened.bytes, canonical.bytes)
      setCurrentSnapshot(loaded.snapshot)
      setTitle(opened.name)
      setTarget(opened.target)
      setSavedRevision(inputWasCanonical ? canonical.snapshot.revision : undefined)
      setFileStatus(inputWasCanonical ? `Opened local file ${opened.name}` : `Opened local file ${opened.name}; canonical local changes need saving`)
    } catch (error) {
      if (isFileAccessCancelled(error)) setFileStatus(undefined)
      else announceFailure('Could not open local file')
    } finally { setFileBusy(false) }
  }

  const save = async (saveAs: boolean) => {
    if (!engine || !fileAccess || saveInFlight.current) return
    saveInFlight.current = true; setFileBusy(true); setFileError(undefined); setFileStatus(saveAs ? 'Preparing Save As…' : 'Preparing local save…')
    try {
      // Must run inside the gesture before awaiting the worker: the native
      // picker is activation-gated. Cancellation leaves every session field as-is.
      const acquired = await fileAccess.acquireSaveTarget({ suggestedName: title, currentTarget: target, saveAs })
      setFileStatus('Saving local file…')
      const serialized = await engine.request('serialize')
      if (!serialized.bytes) throw new Error('Local file could not be serialized')
      const saved = await fileAccess.writeSave(acquired, { bytes: serialized.bytes })
      setTitle(saved.name)
      setTarget(saved.target)
      // Completion establishes only the written revision. It must never repaint
      // an older snapshot over a newer engine commit or call that newer state clean.
      const wroteCurrentRevision = snapshotRef.current?.revision === serialized.snapshot.revision
      if (wroteCurrentRevision) setSavedRevision(serialized.snapshot.revision)
      setFileStatus(wroteCurrentRevision ? (saved.target ? `Saved locally as ${saved.name}` : `Downloaded local file ${saved.name}`) : `Saved revision ${serialized.snapshot.revision}; newer local changes need saving`)
    } catch (error) {
      if (isFileAccessCancelled(error)) setFileStatus(undefined)
      else announceFailure('Could not save local file')
    } finally { saveInFlight.current = false; setFileBusy(false) }
  }

  const startBlank = async () => {
    if (!engine || !blankBytes || fileBusy) return
    setFileBusy(true); setFileError(undefined); setFileStatus('Starting blank local template…')
    try {
      const loaded = await engine.request('load', blankBytes)
      setCurrentSnapshot(loaded.snapshot)
      setTitle('Untitled template'); setTarget(undefined); setSavedRevision(undefined)
      setFileStatus('Started an unnamed local template')
    } catch { announceFailure('Could not start a blank local template')
    } finally { setFileBusy(false) }
  }

  useEffect(() => {
    const shortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's' && engine && fileAccess && !fileBusy) {
        event.preventDefault()
        void save(false)
      }
    }
    window.addEventListener('keydown', shortcut)
    return () => window.removeEventListener('keydown', shortcut)
  })

  if (loadState && !engine) {
    if (loadState.cacheReady && engineState !== 'failed') return <main className="engine-starting" aria-label="Engine preparation"><p role="status" aria-live="polite" aria-label="Engine preparation status">Starting local engine</p></main>
    return <LoadScreen lifecycle={loadState} payload={payload} engineState={engineState} onRetry={onRetry} />
  }

  const engineLabel = initializationError ? 'ENGINE UNAVAILABLE' : snapshot ? `GO SNAPSHOT · REVISION ${snapshot.revision}` : 'ENGINE STARTING'
  const offlineLabel = offlineState === 'ready' ? 'Offline ready' : offlineState === 'checking' ? 'Offline cache checking' : offlineState === 'update-available' ? 'Update available; current release remains usable' : 'Offline cache unavailable'
  const dirty = !snapshot || savedRevision === undefined || snapshot.revision !== savedRevision
  const saveLabel = dirty ? 'Unsaved local changes' : 'Saved local file'
  return <div className="app-shell" aria-label="Folio designer application shell" aria-busy={fileBusy}>
    <header className="document-bar" aria-label="Document bar">
      <span className="brand">FOLIO</span><span className="document-name">{title}</span><span className={`status-dot${dirty ? '' : ' status-clean'}`} aria-hidden="true" /><span className="status-copy" role="status">{saveLabel}</span>
      <div className="document-actions" aria-label="Local file actions"><button className="icon-button" type="button" onClick={() => void open()} disabled={!engine || !fileAccess || fileBusy} aria-label="Open local template"><Icon name="open" /></button><button className="icon-button" type="button" onClick={() => void save(false)} disabled={!engine || !fileAccess || fileBusy} aria-label="Save local template"><Icon name="save" /></button><button className="file-button" type="button" onClick={() => void save(true)} disabled={!engine || !fileAccess || fileBusy}>Save As</button><button className="file-button" type="button" onClick={() => void startBlank()} disabled={!engine || !blankBytes || fileBusy}>Start blank</button></div>
      <span className="later-control" aria-label="Current page setup">{canvas ? `${canvas.preset} · ${canvas.orientation}` : 'Page setup unavailable'}</span>
      <div className="mode-switch" aria-label="Designer mode"><span className="mode-active">DESIGN</span><span className="mode-disabled">PREVIEW · later</span></div>
    </header>
    <div className="workbench" id="future-features">
      <nav className="palette-rail" aria-label="Component palette"><p className="section-label">PALETTE</p>{paletteItems.map((item) => <div className="palette-item" key={item}><span className="palette-icon" aria-hidden="true" />{item}<kbd>later</kbd></div>)}<p className="honest-note">Placement arrives later.</p></nav>
      <main className="canvas-region" aria-label="Canvas region" tabIndex={0}>
        <div className="canvas-tools" aria-label="Canvas controls"><button type="button" onClick={() => setZoom((value) => Math.max(0.5, value - 0.1))} aria-label="Zoom out">−</button><output aria-label="Canvas zoom">{Math.round(zoom * 100)}%</output><button type="button" onClick={() => setZoom((value) => Math.min(2, value + 0.1))} aria-label="Zoom in">+</button><button type="button" onClick={() => setGridVisible((value) => !value)} aria-pressed={gridVisible}>Grid {gridVisible ? 'on' : 'off'}</button><button type="button" onClick={() => setSnapEnabled((value) => !value)} aria-pressed={snapEnabled}>Snap {snapEnabled ? 'on' : 'off'}</button></div>
        {canvas ? <section className={`page-surface${gridVisible ? ' page-grid' : ''}`} aria-label="Report page with Page Header, Content, and Page Footer" style={pageStyle(canvas, zoom)}>
          {canvas.bands.map((band) => <section key={band.name} className={`page-band page-band-${band.name}`} aria-label={bandName(band.name)} style={bandStyle(band, zoom)}><span>{bandName(band.name)}</span></section>)}
        </section> : <p className="canvas-awaiting" role="status">Waiting for Go page geometry.</p>}
        {commitError && <p role="alert" className="file-message">{commitError}</p>}{fileError && <p role="alert" className="file-message">{fileError}</p>}{fileStatus && <p role="status" aria-live="polite" className="file-message">{fileStatus}</p>}
      </main>
      <aside className="properties-panel" aria-label="Properties panel"><p className="section-label">PAGE SETUP</p><label>Preset<select aria-label="Page preset" value={preset} onChange={(event) => setPreset(event.target.value)}><option value="A4">A4</option><option value="Letter">Letter</option><option value="custom">Custom</option></select></label><label>Orientation<select aria-label="Page orientation" value={orientation} onChange={(event) => setOrientation(event.target.value)}><option value="portrait">Portrait</option><option value="landscape">Landscape</option></select></label>{preset === 'custom' && <><Field label="Width (pt)" value={draft.width} onChange={(value) => updateDraft('width', value)}/><Field label="Height (pt)" value={draft.height} onChange={(value) => updateDraft('height', value)}/></>}<Field label="Top margin (pt)" value={draft.top} onChange={(value) => updateDraft('top', value)}/><Field label="Right margin (pt)" value={draft.right} onChange={(value) => updateDraft('right', value)}/><Field label="Bottom margin (pt)" value={draft.bottom} onChange={(value) => updateDraft('bottom', value)}/><Field label="Left margin (pt)" value={draft.left} onChange={(value) => updateDraft('left', value)}/><button type="button" className="file-button" onClick={() => void applyPageSetup()} disabled={!canvas || fileBusy}>Apply page setup</button><p className="honest-note">Grid and snap are editor preferences. Component placement arrives later.</p></aside>
    </div>
    <footer className="status-bar" aria-label="Status bar"><span>LOCAL SHELL</span><code data-testid="engine-snapshot">{engineLabel}</code><span className="status-spacer" /><span role="status" aria-live="polite" aria-label="Offline availability" data-testid="offline-status">{offlineLabel}</span><code>DESIGN MODE</code></footer>
  </div>
}

type Draft = { width: string; height: string; top: string; right: string; bottom: string; left: string }
function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) { return <label>{label}<input aria-label={label} inputMode="decimal" value={value} onChange={(event) => onChange(event.target.value)} /></label> }
function bandName(name: CanvasProjection['bands'][number]['name']): string { return name === 'pageHeader' ? 'Page Header' : name === 'pageFooter' ? 'Page Footer' : 'Content' }
function points(value: number): string { const negative = value < 0; const magnitude = Math.abs(value); const whole = Math.floor(magnitude / 1000); const fraction = String(magnitude % 1000).padStart(3, '0').replace(/0+$/, ''); return `${negative ? '-' : ''}${whole}${fraction ? `.${fraction}` : ''}` }
function draftFor(canvas?: CanvasProjection): Draft { return canvas ? { width: points(canvas.commandWidth), height: points(canvas.commandHeight), top: points(canvas.marginTop), right: points(canvas.marginRight), bottom: points(canvas.marginBottom), left: points(canvas.marginLeft) } : { width: '', height: '', top: '', right: '', bottom: '', left: '' } }
function displayPx(millipoints: number, zoom: number): string { return `${Math.round(millipoints * zoom * 1000) / 1_000_000}px` }
function pageStyle(canvas: CanvasProjection, zoom: number): CSSProperties { return { '--page-display-width': displayPx(canvas.width, zoom), '--page-display-height': displayPx(canvas.height, zoom), '--grid-display-pitch': displayPx(canvas.gridIncrement, zoom) } as CSSProperties }
function bandStyle(band: CanvasProjection['bands'][number], zoom: number): CSSProperties { return { '--band-x': displayPx(band.x, zoom), '--band-y': displayPx(band.y, zoom), '--band-width': displayPx(band.width, zoom), '--band-height': displayPx(band.height, zoom) } as CSSProperties }
function pageSetupDiagnostic(error: unknown): string { const received = error as { code?: string; dataPath?: string; message?: string }; if (received.code === 'PAGE_SETUP_INVALID') return received.dataPath ? `${received.dataPath}: ${received.message ?? 'invalid value'}` : received.message ?? 'Page setup is invalid.'; return 'Page setup is invalid. Check the selected size and margins.' }

function equalBytes(left: ArrayBuffer, right: ArrayBuffer): boolean {
  const a = new Uint8Array(left)
  const b = new Uint8Array(right)
  return a.length === b.length && a.every((value, index) => value === b[index])
}
