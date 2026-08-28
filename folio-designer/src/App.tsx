import './App.css'
import { useEffect, useRef, useState, type CSSProperties, type PointerEvent } from 'react'
import type { EngineClient } from './engine-client'
import type { CanvasProjection, EngineSnapshot } from './engine-protocol'
import type { OfflineLifecycleState } from './offline-lifecycle'
import type { OfflineLifecycle } from './offline-lifecycle'
import type { S1Payload } from './release-payload'
import { LoadScreen } from './LoadScreen'
import { isFileAccessCancelled, type FileAccess, type FileTarget } from './file/file-access'
import { pageSetupCommand } from './page-setup-command'
import { deleteComponentCommand, dropComponentCommand, moveComponentCommand, resizeComponentCommand, type PaletteKind } from './component-command'
import { updateComponentPropertiesCommand, type PropertyField, type PropertyIntent } from './component-property-command'

const paletteItems: ReadonlyArray<readonly [string, PaletteKind]> = [['Text', 'text'], ['Image', 'image'], ['Table', 'table'], ['Line', 'line'], ['Rectangle', 'rect']]

function Icon({ name }: { name: 'open' | 'save' }) {
  return <svg aria-hidden="true" className="icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.25"><path d={name === 'open' ? 'M2 5.5h4l1.2-2h6.8v9H2z M2 5.5h12' : 'M3 2h8l2 2v10H3z M5 2v4h6V2 M5 12h6'} /></svg>
}

type AppProps = Readonly<{ engine?: EngineClient; fileAccess?: FileAccess; initialSnapshot?: EngineSnapshot; blankBytes?: ArrayBuffer; initializationError?: string; offlineState?: OfflineLifecycleState; loadState?: OfflineLifecycle; payload?: S1Payload; engineState?: 'waiting' | 'starting' | 'failed'; onRetry?: () => void }>

export default function App({ engine, fileAccess, initialSnapshot, blankBytes, initializationError, offlineState = 'unavailable', loadState, payload, engineState = 'waiting', onRetry = () => undefined }: AppProps = {}) {
  const [snapshot, setSnapshot] = useState(initialSnapshot)
  const [commitError, setCommitError] = useState<string>()
  const [propertyError, setPropertyError] = useState<PropertyCommitError>()
  const [fileError, setFileError] = useState<string>()
  const [fileStatus, setFileStatus] = useState<string>()
  const [fileBusy, setFileBusy] = useState(false)
  const [title, setTitle] = useState('Untitled template')
  const [target, setTarget] = useState<FileTarget>()
  const [savedRevision, setSavedRevision] = useState<number>()
  const [zoom, setZoom] = useState(1)
  const [gridVisible, setGridVisible] = useState(true)
  const [snapEnabled, setSnapEnabled] = useState(true)
  const [placing, setPlacing] = useState<PaletteKind>()
  const [hoverBand, setHoverBand] = useState<CanvasProjection['bands'][number]['name']>()
  const [selected, setSelected] = useState<ReadonlyArray<string>>([])
  const [drag, setDrag] = useState<DragState>()
  const [preset, setPreset] = useState<string>(initialSnapshot?.canvas?.preset ?? 'A4')
  const [orientation, setOrientation] = useState<string>(initialSnapshot?.canvas?.orientation ?? 'portrait')
  const [draft, setDraft] = useState(() => draftFor(initialSnapshot?.canvas))
  const snapshotRef = useRef(snapshot)
  const saveInFlight = useRef(false)
  const draftGeneration = useRef(0)
  const documentGeneration = useRef(0)
  const [documentGenerationValue, setDocumentGenerationValue] = useState(0)
  const selectedRef = useRef(selected)
  useEffect(() => { selectedRef.current = selected }, [selected])
  const canvas = snapshot?.canvas
  const clearInteraction = () => { setPlacing(undefined); setHoverBand(undefined); setDrag(undefined) }
  const commitComponent = async (payload: ArrayBuffer, after?: () => void) => {
    if (!engine || fileBusy) return
    setCommitError(undefined)
    try { const result = await engine.request('command', payload); setCurrentSnapshot(result.snapshot); after?.() }
    catch (error) { setCommitError(componentDiagnostic(error)); clearInteraction() }
  }
  const place = (x: number, y: number) => {
    if (!placing) return
    const kind = placing
    clearInteraction()
    void commitComponent(dropComponentCommand(kind, x, y, snapEnabled))
  }
  const select = (id: string, extend: boolean) => setSelected((current) => extend ? (current.includes(id) ? current.filter((value) => value !== id) : [...current, id]) : [id])
  const deleteSelection = () => { if (selected.length === 1) void commitComponent(deleteComponentCommand(selected[0]!), () => setSelected([])) }
  const applyPageSetup = async () => {
    if (!engine || !canvas || fileBusy) return
    setCommitError(undefined)
    try {
      const requestGeneration = draftGeneration.current
      const result = await engine.request('command', pageSetupCommand(preset, orientation, preset === 'custom' ? draft.width : '0', preset === 'custom' ? draft.height : '0', draft))
      setCurrentSnapshot(result.snapshot, draftGeneration.current !== requestGeneration)
    } catch (error) { setCommitError(pageSetupDiagnostic(error)) }
  }
  const applyProperties = async (ids: ReadonlyArray<string>, intent: PropertyIntent, responseGeneration: number, selectionKey: string): Promise<CanvasProjection | undefined> => {
    if (!engine || fileBusy) return undefined
    setCommitError(undefined)
    setPropertyError(undefined)
    try {
      const result = await engine.request('command', updateComponentPropertiesCommand(ids, intent))
      // A command result is authoritative only if it advances this snapshot;
      // field drafts independently ignore a response from a replaced scope.
      if ((snapshotRef.current?.revision ?? -1) < result.snapshot.revision) setCurrentSnapshot(result.snapshot)
      return documentGeneration.current === responseGeneration && selectedRef.current.join(',') === selectionKey ? result.snapshot.canvas : undefined
    } catch (error) {
      if (documentGeneration.current === responseGeneration && selectedRef.current.join(',') === selectionKey) {
        const diagnostic = componentDiagnosticDetail(error)
        setPropertyError({ field: intent.field, selectionKey, ...diagnostic })
      }
      return undefined
    }
  }

  const setCurrentSnapshot = (next: EngineSnapshot | undefined, keepNewerDraft = false, clearDocumentInteraction = false) => { snapshotRef.current = next; setSnapshot(next); if (clearDocumentInteraction) { documentGeneration.current++; setDocumentGenerationValue(documentGeneration.current); setSelected([]); clearInteraction() }; if (next?.canvas) { setPreset(next.canvas.preset); setOrientation(next.canvas.orientation); if (!keepNewerDraft) setDraft(draftFor(next.canvas)) } }
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
      setCurrentSnapshot(loaded.snapshot, false, true)
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
      setCurrentSnapshot(loaded.snapshot, false, true)
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
      <nav className="palette-rail" aria-label="Component palette"><p className="section-label">PALETTE</p>{paletteItems.map(([label, kind]) => <button className="palette-item" type="button" key={kind} onPointerDown={() => { setPlacing(kind); setHoverBand(undefined) }} onClick={() => { setPlacing(kind); setHoverBand(undefined) }} aria-pressed={placing === kind} aria-label={`Place ${label}`}><span className="palette-icon" aria-hidden="true" />{label}<kbd>place</kbd></button>)}<p className="honest-note">Choose or drag a component, then choose a page band.</p></nav>
      <main className="canvas-region" aria-label="Canvas region" tabIndex={0} onClick={(event) => { if (event.target === event.currentTarget) setSelected([]) }} onKeyDown={(event) => { if ((event.key === 'Delete' || event.key === 'Backspace') && event.target === event.currentTarget && selected.length === 1) { event.preventDefault(); deleteSelection() } if (event.key === 'Escape') { clearInteraction(); setSelected([]) } }}>
        <div className="canvas-tools" aria-label="Canvas controls"><button type="button" onClick={() => setZoom((value) => Math.max(0.5, value - 0.1))} aria-label="Zoom out">−</button><output aria-label="Canvas zoom">{Math.round(zoom * 100)}%</output><button type="button" onClick={() => setZoom((value) => Math.min(2, value + 0.1))} aria-label="Zoom in">+</button><button type="button" onClick={() => setGridVisible((value) => !value)} aria-pressed={gridVisible}>Grid {gridVisible ? 'on' : 'off'}</button><button type="button" onClick={() => setSnapEnabled((value) => !value)} aria-pressed={snapEnabled}>Snap {snapEnabled ? 'on' : 'off'}</button></div>
        {canvas ? <section className={`page-surface${gridVisible ? ' page-grid' : ''}`} aria-label="Report page with Page Header, Content, and Page Footer" style={pageStyle(canvas, zoom)} onClick={() => setSelected([])}>
          {canvas.bands.map((band) => <section key={band.name} className={`page-band page-band-${band.name}${hoverBand === band.name ? ' page-band-target' : ''}`} aria-label={bandName(band.name)} aria-current={hoverBand === band.name ? 'true' : undefined} style={bandStyle(band, zoom)} tabIndex={0} onPointerEnter={() => placing && setHoverBand(band.name)} onPointerLeave={() => setHoverBand((current) => current === band.name ? undefined : current)} onPointerUp={(event) => { if (placing && event.currentTarget === event.target) place(band.x / 1000 + displayPoint(event.nativeEvent.offsetX, zoom), band.y / 1000 + displayPoint(event.nativeEvent.offsetY, zoom)) }} onKeyDown={(event) => { if (placing && (event.key === 'Enter' || event.key === ' ')) { event.preventDefault(); place(band.x / 1000, band.y / 1000) } }}><span>{bandName(band.name)}</span>{canvas.components.filter((component) => component.band === band.name).map((component) => <CanvasComponent key={component.id} component={component} zoom={zoom} selected={selected.includes(component.id)} preview={drag?.id === component.id ? drag : undefined} onSelect={select} onDelete={deleteSelection} onDragStart={setDrag} onDragEnd={(finished) => { setDrag(undefined); if (!finished.changed) return; if (finished.mode === 'move') void commitComponent(moveComponentCommand(component.id, finished.x, finished.y, snapEnabled)); else void commitComponent(resizeComponentCommand(component.id, finished.width, finished.height, snapEnabled)) }} />)}</section>)}
        </section> : <p className="canvas-awaiting" role="status">Waiting for Go page geometry.</p>}
        {commitError && <p role="alert" className="file-message">{commitError}</p>}{fileError && <p role="alert" className="file-message">{fileError}</p>}{fileStatus && <p role="status" aria-live="polite" className="file-message">{fileStatus}</p>}
      </main>
      <aside className="properties-panel" aria-label="Properties panel">{selected.length > 0 && canvas ? <ComponentProperties key={`${documentGenerationValue}:${selected.join(',')}`} components={canvas.components.filter((component) => selected.includes(component.id))} onCommit={applyProperties} documentGeneration={documentGenerationValue} propertyError={propertyError} /> : <PageSetup preset={preset} orientation={orientation} draft={draft} onPreset={setPreset} onOrientation={setOrientation} onDraft={updateDraft} onApply={applyPageSetup} disabled={!canvas || fileBusy} />}</aside>
    </div>
    <footer className="status-bar" aria-label="Status bar"><span>LOCAL SHELL</span><code data-testid="engine-snapshot">{engineLabel}</code><span className="status-spacer" /><span role="status" aria-live="polite" aria-label="Offline availability" data-testid="offline-status">{offlineLabel}</span><code>DESIGN MODE</code></footer>
  </div>
}

function PageSetup({ preset, orientation, draft, onPreset, onOrientation, onDraft, onApply, disabled }: { preset: string; orientation: string; draft: Draft; onPreset: (value: string) => void; onOrientation: (value: string) => void; onDraft: (key: keyof Draft, value: string) => void; onApply: () => void; disabled: boolean }) {
  return <><p className="section-label">PAGE SETUP</p><p className="honest-note">Component properties require a selection.</p><label>Preset<select aria-label="Page preset" value={preset} onChange={(event) => onPreset(event.target.value)}><option value="A4">A4</option><option value="Letter">Letter</option><option value="custom">Custom</option></select></label><label>Orientation<select aria-label="Page orientation" value={orientation} onChange={(event) => onOrientation(event.target.value)}><option value="portrait">Portrait</option><option value="landscape">Landscape</option></select></label>{preset === 'custom' && <><Field label="Width (pt)" value={draft.width} onChange={(value) => onDraft('width', value)}/><Field label="Height (pt)" value={draft.height} onChange={(value) => onDraft('height', value)}/></>}<Field label="Top margin (pt)" value={draft.top} onChange={(value) => onDraft('top', value)}/><Field label="Right margin (pt)" value={draft.right} onChange={(value) => onDraft('right', value)}/><Field label="Bottom margin (pt)" value={draft.bottom} onChange={(value) => onDraft('bottom', value)}/><Field label="Left margin (pt)" value={draft.left} onChange={(value) => onDraft('left', value)}/><button type="button" className="file-button" onClick={onApply} disabled={disabled}>Apply page setup</button><p className="honest-note">Grid and snap are editor preferences. Undo and redo arrive later.</p></>
}

type PanelComponent = CanvasProjection['components'][number]
type PropertyCommitError = Readonly<{ field: PropertyField; selectionKey: string; elementId?: string; dataPath?: string; message: string }>
type CommitProperties = (ids: ReadonlyArray<string>, intent: PropertyIntent, generation: number, key: string) => Promise<CanvasProjection | undefined>
const baseFields: ReadonlyArray<readonly [PropertyField, string]> = [['x', 'X (pt)'], ['y', 'Y (pt)'], ['visibleIf', 'Visible if']]
const styleFields: ReadonlyArray<readonly [PropertyField, string]> = [['background', 'Background'], ['borderWidth', 'Border width (pt)'], ['borderColor', 'Border colour'], ['paddingTop', 'Padding top (pt)'], ['paddingRight', 'Padding right (pt)'], ['paddingBottom', 'Padding bottom (pt)'], ['paddingLeft', 'Padding left (pt)']]
const typographyFields: ReadonlyArray<readonly [PropertyField, string]> = [['fontFamily', 'Font family'], ['fontSize', 'Font size (pt)'], ['align', 'Align'], ['valign', 'Vertical align']]
function ComponentProperties({ components, onCommit, documentGeneration, propertyError }: { components: ReadonlyArray<PanelComponent>; onCommit: CommitProperties; documentGeneration: number; propertyError?: PropertyCommitError }) {
  const ids = components.map((component) => component.id)
  const types = new Set(components.map((component) => component.type))
  const all = (predicate: (type: PanelComponent['type']) => boolean) => [...types].every(predicate)
  const fields: Array<readonly [PropertyField, string]> = [...baseFields]
  if (all((type) => type !== 'table')) fields.splice(2, 0, ['width', 'Width (pt)'], ['height', 'Height (pt)'])
  if (components.length === 1 && types.has('text')) fields.push(['value', 'Text value'])
  if (all((type) => type === 'text' || type === 'table')) fields.push(...typographyFields)
  fields.push(...styleFields)
  const scopedError = propertyError?.selectionKey === ids.join(',') ? propertyError : undefined
  return <><p className="section-label">{components.length === 1 ? 'COMPONENT' : 'COMPONENTS'}</p><p className="panel-heading">{components.length === 1 ? `${components[0]!.id} · ${components[0]!.type}` : `${components.length} selected`}</p>{fields.map(([field, label]) => <PropertyDraft key={field} label={label} field={field} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === field ? scopedError : undefined} />)}<BorderEdgesProperty components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'borderEdges' ? scopedError : undefined} />{all((type) => type === 'text' || type === 'table') && <BooleanProperty label="Bold" field="bold" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'bold' ? scopedError : undefined} />}{all((type) => type === 'text' || type === 'table') && <BooleanProperty label="Italic" field="italic" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'italic' ? scopedError : undefined} />}{components.length === 1 && components[0]!.type === 'table' && <p className="honest-note">Table binding: {components[0]!.tableBind ?? 'Not set'} (display only)</p>}<p className="honest-note">{types.has('table') ? 'Table size, binding, columns, and header style are not editable here; table geometry is derived.' : 'Only committed engine values are shown. Asset import and arbitrary CSS are not editable here.'}</p></>
}

function committedValue(component: PanelComponent, field: PropertyField): string | undefined {
  const value = component[field as keyof PanelComponent]
  if (typeof value === 'number') return points(value)
  return typeof value === 'string' ? value : undefined
}
function PropertyDraft({ label, field, components, ids, onCommit, documentGeneration, error }: { label: string; field: PropertyField; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) {
  const values = components.map((component) => committedValue(component, field))
  const same = values.every((value) => value === values[0])
  const committed = same ? values[0] ?? '' : ''
  const [draft, setDraft] = useState(committed)
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  const selectionKey = ids.join(',')
  const revert = () => setDraft(committed)
  const submit = async (intent: PropertyIntent, reconcileDraft: boolean) => {
    if (pendingRef.current) return
    pendingRef.current = true
    setPending(true)
    const accepted = await onCommit(ids, intent, documentGeneration, selectionKey)
    pendingRef.current = false
    setPending(false)
    if (accepted && reconcileDraft) setDraft(canonicalValue(accepted, ids, field) ?? draft)
  }
  const commit = async () => { if (draft !== committed) await submit({ field, operation: draft === '' && field !== 'value' ? 'clear' : 'set', value: draft }, true) }
  const canNull = field === 'visibleIf' || field === 'background'
  const errorId = error ? `property-error-${field}` : undefined
  return <div className="property-editor"><label>{label}<input aria-label={label} aria-description={same ? undefined : 'Mixed value'} aria-invalid={error ? 'true' : undefined} aria-errormessage={errorId} inputMode={label.includes('(pt)') ? 'decimal' : undefined} value={draft} placeholder={same ? undefined : 'Mixed'} disabled={pending} onChange={(event) => setDraft(event.target.value)} onBlur={() => void commit()} onKeyDown={(event) => { if (event.key === 'Enter') { event.preventDefault(); void commit() } if (event.key === 'Escape') { event.preventDefault(); revert(); (event.target as HTMLInputElement).blur() } }} /></label>{field !== 'x' && field !== 'y' && field !== 'width' && field !== 'height' && field !== 'value' && <button type="button" className="property-action" disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'clear' }, true)}>Clear {label}</button>}{canNull && <button type="button" className="property-action" disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'null' }, true)}>Set {label} null</button>}{error && <p id={errorId} role="alert" className="property-error">{error.elementId ? `${error.elementId}: ` : ''}{error.dataPath ? `${error.dataPath}: ` : ''}{error.message}</p>}</div>
}
function canonicalValue(canvas: CanvasProjection, ids: ReadonlyArray<string>, field: PropertyField): string | undefined { const values = canvas.components.filter((component) => ids.includes(component.id)).map((component) => committedValue(component, field)); return values.length === ids.length && values.every((value) => value === values[0]) ? values[0] ?? '' : undefined }
function BooleanProperty({ label, field, components, ids, onCommit, documentGeneration, error }: { label: string; field: 'bold' | 'italic'; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) {
  const values = components.map((component) => component[field])
  const uniform = values.every((value) => value === values[0])
  const active = uniform && values[0] === true
  const [pending, setPending] = useState(false); const pendingRef = useRef(false); const commit = async (operation: PropertyIntent['operation'], value?: boolean) => { if (pendingRef.current) return; pendingRef.current = true; setPending(true); await onCommit(ids, { field, operation, value }, documentGeneration, ids.join(',')); pendingRef.current = false; setPending(false) }
  return <div className="property-editor"><button type="button" className="file-button property-toggle" disabled={pending} aria-pressed={active} aria-label={uniform ? label : `${label}, mixed`} onClick={() => void commit('set', !active)}>{uniform ? label : `${label} · Mixed`}</button><button type="button" className="property-action" disabled={pending} onClick={() => void commit('clear')}>Clear {label}</button>{error && <p role="alert" className="property-error">{error.message}</p>}</div>
}
function BorderEdgesProperty({ components, ids, onCommit, documentGeneration, error }: { components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) { const values = components.map((component) => (component.borderEdges ?? []).join(',')); const same = values.every((value) => value === values[0]); const [edges, setEdges] = useState<string[]>(same && values[0] ? values[0].split(',') : []); const pending = useRef(false); const update = async (next: string[]) => { if (pending.current) return; pending.current = true; setEdges(next); await onCommit(ids, { field: 'borderEdges', operation: next.length ? 'set' : 'clear', ...(next.length ? { value: next } : {}) }, documentGeneration, ids.join(',')); pending.current = false }; return <fieldset className="property-editor"><legend>Border edges</legend>{['top', 'right', 'bottom', 'left'].map((edge) => <label key={edge}><input type="checkbox" aria-label={`Border ${edge}`} checked={edges.includes(edge)} onChange={() => void update(edges.includes(edge) ? edges.filter((value) => value !== edge) : [...edges, edge])} />{edge}</label>)}{!same && <span aria-label="Border edges mixed">Mixed</span>}<button type="button" className="property-action" onClick={() => void update([])}>Clear Border edges</button>{error && <p role="alert" className="property-error">{error.message}</p>}</fieldset> }

type Draft = { width: string; height: string; top: string; right: string; bottom: string; left: string }
function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) { return <label>{label}<input aria-label={label} inputMode="decimal" value={value} onChange={(event) => onChange(event.target.value)} /></label> }
function bandName(name: CanvasProjection['bands'][number]['name']): string { return name === 'pageHeader' ? 'Page Header' : name === 'pageFooter' ? 'Page Footer' : 'Content' }
function points(value: number): string { const negative = value < 0; const magnitude = Math.abs(value); const whole = Math.floor(magnitude / 1000); const fraction = String(magnitude % 1000).padStart(3, '0').replace(/0+$/, ''); return `${negative ? '-' : ''}${whole}${fraction ? `.${fraction}` : ''}` }
function draftFor(canvas?: CanvasProjection): Draft { return canvas ? { width: points(canvas.commandWidth), height: points(canvas.commandHeight), top: points(canvas.marginTop), right: points(canvas.marginRight), bottom: points(canvas.marginBottom), left: points(canvas.marginLeft) } : { width: '', height: '', top: '', right: '', bottom: '', left: '' } }
function displayPx(millipoints: number, zoom: number): string { return `${Math.round(millipoints * zoom * 1000) / 1_000_000}px` }
function pageStyle(canvas: CanvasProjection, zoom: number): CSSProperties { return { '--page-display-width': displayPx(canvas.width, zoom), '--page-display-height': displayPx(canvas.height, zoom), '--grid-display-pitch': displayPx(canvas.gridIncrement, zoom) } as CSSProperties }
function bandStyle(band: CanvasProjection['bands'][number], zoom: number): CSSProperties { return { '--band-x': displayPx(band.x, zoom), '--band-y': displayPx(band.y, zoom), '--band-width': displayPx(band.width, zoom), '--band-height': displayPx(band.height, zoom) } as CSSProperties }
function pageSetupDiagnostic(error: unknown): string { const received = error as { code?: string; dataPath?: string; message?: string }; if (received.code === 'PAGE_SETUP_INVALID') return received.dataPath ? `${received.dataPath}: ${received.message ?? 'invalid value'}` : received.message ?? 'Page setup is invalid.'; return 'Page setup is invalid. Check the selected size and margins.' }
function componentDiagnosticDetail(error: unknown): Readonly<{ elementId?: string; dataPath?: string; message: string }> { const received = error as { elementId?: string; dataPath?: string; message?: string }; return { ...(received.elementId ? { elementId: received.elementId } : {}), ...(received.dataPath ? { dataPath: received.dataPath } : {}), message: received.message ?? 'Component change was rejected.' } }
function componentDiagnostic(error: unknown): string { const received = componentDiagnosticDetail(error); const prefix = received.elementId ?? received.dataPath; return prefix ? `${prefix}: ${received.message}` : received.message }
function displayPoint(pixel: number, zoom: number): number { return Math.round((pixel / zoom) * 1000) / 1000 }

type DragState = Readonly<{ id: string; mode: 'move' | 'resize'; startClientX: number; startClientY: number; x: number; y: number; width: number; height: number; originalX: number; originalY: number; originalWidth: number; originalHeight: number; changed: boolean }>
function CanvasComponent({ component, zoom, selected, preview, onSelect, onDelete, onDragStart, onDragEnd }: { component: CanvasProjection['components'][number]; zoom: number; selected: boolean; preview?: DragState; onSelect: (id: string, extend: boolean) => void; onDelete: () => void; onDragStart: (drag: DragState | undefined) => void; onDragEnd: (drag: DragState) => void }) {
  const selectedByPointer = useRef(false)
  const active = preview ?? { x: component.x, y: component.y, width: component.width, height: component.height }
  const begin = (event: PointerEvent, mode: 'move' | 'resize') => { event.stopPropagation(); selectedByPointer.current = true; onSelect(component.id, event.shiftKey); if (event.shiftKey) return; event.currentTarget.setPointerCapture?.(event.pointerId); onDragStart({ id: component.id, mode, startClientX: event.clientX, startClientY: event.clientY, x: component.x, y: component.y, width: component.width, height: component.height, originalX: component.x, originalY: component.y, originalWidth: component.width, originalHeight: component.height, changed: false }) }
  const move = (event: PointerEvent) => { if (!preview) return; const rawDX = event.clientX - preview.startClientX; const rawDY = event.clientY - preview.startClientY; const changed = preview.changed || Math.hypot(rawDX, rawDY) >= 2; const dx = displayPoint(rawDX, zoom) * 1000; const dy = displayPoint(rawDY, zoom) * 1000; onDragStart({ ...preview, changed, x: preview.originalX + (preview.mode === 'move' ? dx : 0), y: preview.originalY + (preview.mode === 'move' ? dy : 0), width: preview.originalWidth + (preview.mode === 'resize' ? dx : 0), height: preview.originalHeight + (preview.mode === 'resize' ? dy : 0) }) }
  const finish = (event: PointerEvent) => { if (!preview) return; event.stopPropagation(); onDragEnd(preview) }
  return <div className={`canvas-component canvas-component-${component.type}${selected ? ' canvas-component-selected' : ''}`} aria-label={`${component.type} component ${component.id}`} role="button" tabIndex={0} style={componentStyle(active, zoom)} onClick={(event) => { event.stopPropagation(); if (!selectedByPointer.current) onSelect(component.id, event.shiftKey); selectedByPointer.current = false }} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelect(component.id, event.shiftKey) } if (selected && (event.key === 'Delete' || event.key === 'Backspace')) { event.preventDefault(); event.stopPropagation(); onDelete() } }} onPointerDown={(event) => begin(event, 'move')} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => onDragStart(undefined)}>{component.type === 'text' ? 'Text' : component.type === 'image' ? 'Image' : component.type === 'table' ? 'Table' : ''}{selected && component.resizable && <button type="button" className="resize-handle" aria-label={`Resize ${component.id}`} onPointerDown={(event) => begin(event, 'resize')} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => onDragStart(undefined)} />}</div>
}
function componentStyle(component: { x: number; y: number; width: number; height: number }, zoom: number): CSSProperties { return { '--component-x': displayPx(component.x, zoom), '--component-y': displayPx(component.y, zoom), '--component-width': displayPx(component.width, zoom), '--component-height': displayPx(component.height, zoom) } as CSSProperties }

function equalBytes(left: ArrayBuffer, right: ArrayBuffer): boolean {
  const a = new Uint8Array(left)
  const b = new Uint8Array(right)
  return a.length === b.length && a.every((value, index) => value === b[index])
}
