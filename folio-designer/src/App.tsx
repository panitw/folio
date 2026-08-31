import './App.css'
import { useCallback, useEffect, useId, useLayoutEffect, useRef, useState, type ClipboardEvent as ReactClipboardEvent, type CSSProperties, type KeyboardEvent as ReactKeyboardEvent, type PointerEvent, type ReactNode } from 'react'
import { isProducerRenderFailure, type EngineClient } from './engine-client'
import type { CanvasProjection, EngineDiagnostic, EngineError, EngineSnapshot, TableColumns } from './engine-protocol'
import type { OfflineLifecycleState } from './offline-lifecycle'
import type { OfflineLifecycle } from './offline-lifecycle'
import { engineMayStart } from './offline-lifecycle'
import type { S1Payload } from './release-payload'
import { LoadScreen } from './LoadScreen'
import type { BindingErrorScope } from './DataPanel'
import { isFileAccessCancelled, type FileAccess, type FileTarget } from './file/file-access'
import { pageSetupCommand } from './page-setup-command'
import { bindComponentScalarCommand, createComponentCommand, deleteComponentCommand, dropComponentCommand, duplicateComponentCommand, moveComponentCommand, setComponentBoundsCommand, type PaletteKind } from './component-command'
import { updateComponentPropertiesCommand, type PropertyField, type PropertyIntent } from './component-property-command'
import { FontChainEditor } from './FontChainEditor'
import { type FontChainCommitError, type FontChainControl } from './font-chain-control'
import { proposedBounds, resizeAnchors, type DragAnchor, type DragLimit } from './resize-anchor'
import { columnEdgeAfterDrag, sheetStack, SHEET_STACK_GAP, type Sheet, type SheetOccurrence, type SheetStack } from './sheet-stack'
import { addTableColumnCommand, configureTableBindingCommand, moveTableColumnCommand, removeTableColumnCommand, updateTableColumnBindingCommand, updateTableColumnCommand, updateTableColumnFooterCommand } from './table-column-command'
import { TableEditor } from './TableEditor'
import { initialPDFPreviewViewState, PDFPreviewViewer, samePDFPreviewViewState, type PDFPreviewViewState } from './preview/pdf-viewer'
import { canInstallPreview, PREVIEW_DEBOUNCE_MS, PreviewWorkScheduler, staleCopy } from './preview/freshness'
import { PreviewDiagnostics, PreviewFailure, type DiagnosticLocation } from './preview/diagnostic-presenter'
import { isMacPlatform, primaryModifier, shortcutHintsFor } from './shortcuts'
import { DataPanel } from './DataPanel'
import { acceptSampleData, type SampleData, type SampleNode } from './sample-data'
import type { SampleFileAccess } from './sample-file'
import { assetBytesRequest, setComponentAssetCommand } from './component-asset-command'
import { embeddedFaceFamily } from './embedded-face-family'
import { registerCarriedFaces } from './embedded-face-registry'
import type { ImageFileAccess } from './image-file'

const paletteItems: ReadonlyArray<readonly [string, PaletteKind]> = [['Text', 'text'], ['Image', 'image'], ['Table', 'table'], ['Line', 'line'], ['Rectangle', 'rect']]
type InspectorTab = 'properties' | 'data'
const inspectorTabs: ReadonlyArray<readonly [InspectorTab, string, string]> = [['properties', 'PROPERTIES', 'INPUTS'], ['data', 'DATA', 'DATA']]
const EMPTY_PARAMETER_DOCUMENT = '{}'
const MAX_PARAMETER_DOCUMENT_BYTES = 8 * 1024 * 1024

// Sample inspection is strictly a local affordance. These values are never
// sent with a command; Go still owns all collection and field admission.
function tableSampleCandidates(root: SampleNode | undefined): ReadonlyArray<Readonly<{ collection: string; field: string }>> {
  if (!root) return []
  const candidates = new Map<string, Readonly<{ collection: string; field: string }>>()
  const visit = (node: SampleNode) => {
    if (node.kind === 'collection' && node.segments?.length && node.segments.every((part) => /^[A-Za-z_][A-Za-z0-9_]*$/.test(part))) {
      const collection = `${node.segments.join('.')}[]`
      const fields = (value: SampleNode, prefix: ReadonlyArray<string>): void => {
        if (value.kind === 'collection' || value.kind === 'truncated') return
        if (value.kind === 'object') { value.children.forEach((child) => fields(child, [...prefix, child.label])); return }
        if (prefix.length && prefix.every((part) => /^[A-Za-z_][A-Za-z0-9_]*$/.test(part))) {
          const field = prefix.join('.')
          candidates.set(`${collection}\u0000${field}`, { collection, field })
        }
      }
      node.children.forEach((item) => fields(item, []))
    }
    node.children.forEach(visit)
  }
  visit(root)
  return [...candidates.values()].sort((a, b) => a.collection.localeCompare(b.collection) || a.field.localeCompare(b.field)).slice(0, 50)
}

type ParameterReferenceState = Readonly<{ status: 'pending' | 'ready' | 'failed'; names: ReadonlyArray<string> }>

function Icon({ name }: { name: 'open' | 'save' }) {
  return <svg aria-hidden="true" className="icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.25"><path d={name === 'open' ? 'M2 5.5h4l1.2-2h6.8v9H2z M2 5.5h12' : 'M3 2h8l2 2v10H3z M5 2v4h6V2 M5 12h6'} /></svg>
}

const paletteGlyphs: Readonly<Record<PaletteKind, ReactNode>> = {
  text: <><path d="M3.5 4.5V3h9v1.5" /><path d="M8 3v10" /><path d="M5.75 13h4.5" /></>,
  image: <><path d="M2.5 3.5h11v9h-11z" /><path d="M2.5 10.25 5.75 7l2.25 2.25 2-2 3.5 3.5" /><circle cx="10.5" cy="6.25" r="1" /></>,
  table: <><path d="M2.5 3.5h11v9h-11z" /><path d="M2.5 6.5h11" /><path d="M6.5 6.5v6" /><path d="M10 6.5v6" /></>,
  line: <><path d="M3 12.5 13 3.5" /></>,
  rect: <><path d="M2.5 4.5h11v7h-11z" /></>,
}

function PaletteIcon({ kind }: { kind: PaletteKind }) {
  return <svg aria-hidden="true" className="palette-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.25" strokeLinecap="square">{paletteGlyphs[kind]}</svg>
}

type AppProps = Readonly<{ engine?: EngineClient; fileAccess?: FileAccess; sampleFileAccess?: SampleFileAccess; imageFileAccess?: ImageFileAccess; initialSnapshot?: EngineSnapshot; initialSampleData?: SampleData; blankBytes?: ArrayBuffer; initializationError?: string; offlineState?: OfflineLifecycleState; loadState?: OfflineLifecycle; payload?: S1Payload; engineState?: 'waiting' | 'starting' | 'failed'; onRetry?: () => void }>
// The document carries no readable face until one is registered, and this is
// the value that says so. A stable reference, so resetting it between
// documents is not itself a state change React has to re-render for.
const NO_CARRIED_FACES: ReadonlySet<string> = new Set()
type PreviewRecord = Readonly<{ bytes: ArrayBuffer; revision: number; identity: string; digest: string; diagnostics: ReadonlyArray<EngineDiagnostic>; token: number; generation: number }>
type PreviewFailureRecord = Readonly<{ error: EngineError; token: number; generation: number; revision: number }>

export default function App({ engine, fileAccess, sampleFileAccess, imageFileAccess, initialSnapshot, initialSampleData, blankBytes, initializationError, offlineState = 'unavailable', loadState, payload, engineState = 'waiting', onRetry = () => undefined }: AppProps = {}) {
  const [snapshot, setSnapshot] = useState(initialSnapshot)
  const [commitError, setCommitError] = useState<string>()
  const [propertyError, setPropertyError] = useState<PropertyCommitError>()
  const [fontChainError, setFontChainError] = useState<FontChainCommitError>()
  const [fontChainBusy, setFontChainBusy] = useState(false)
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
  // Where the armed palette kind is following the pointer. One transient
  // client coordinate for chrome that never touches document geometry: it
  // places no component and proposes nothing to Go.
  const [placingAt, setPlacingAt] = useState<Readonly<{ x: number; y: number }>>()
  // The hovered band, keyed by SHEET as well as by band: three sheets carry
  // three content bands, and highlighting all of them would be exactly the
  // ambiguous drop target this epic's own rule forbids.
  const [hoverBand, setHoverBand] = useState<string>()
  const [selected, setSelected] = useState<ReadonlyArray<string>>([])
  const [drag, setDrag] = useState<DragState>()
  const [preset, setPreset] = useState<string>(initialSnapshot?.canvas?.preset ?? 'A4')
  const [orientation, setOrientation] = useState<string>(initialSnapshot?.canvas?.orientation ?? 'portrait')
  const [draft, setDraft] = useState(() => draftFor(initialSnapshot?.canvas))
  const [mode, setMode] = useState<'design' | 'preview'>('design')
  // One right-hand inspector with two tabs, per the UX design. The tab is a
  // purely local view preference; it never changes what the engine owns.
  const [inspectorTab, setInspectorTab] = useState<InspectorTab>('properties')
  const [preview, setPreview] = useState<PreviewRecord | undefined>(undefined)
  const [previewStatus, setPreviewStatus] = useState<'idle' | 'checking' | 'debouncing' | 'rendering' | 'current' | 'stale' | 'error'>('idle')
  const [staleReason, setStaleReason] = useState<'inputs-changed' | 'render-failed'>('inputs-changed')
  const [previewError, setPreviewError] = useState<PreviewFailureRecord>()
  const [previewIssue, setPreviewIssue] = useState<string>()
  const [dismissedDiagnostics, setDismissedDiagnostics] = useState<ReadonlySet<string>>(new Set())
  const [undoAvailable, setUndoAvailable] = useState(initialSnapshot?.canUndo === true)
  const [redoAvailable, setRedoAvailable] = useState(initialSnapshot?.canRedo === true)
  const [locateStatus, setLocateStatus] = useState<string>()
  const [previewViewState, setPreviewViewState] = useState<PDFPreviewViewState>(initialPDFPreviewViewState)
  // Accepted bytes and editor draft are intentionally separate. The engine
  // receives only accepted raw text; invalid local input cannot silently turn
  // into an alternate runtime value.
  const [previewParams, setPreviewParams] = useState(EMPTY_PARAMETER_DOCUMENT)
  const [previewParamsDraft, setPreviewParamsDraft] = useState(EMPTY_PARAMETER_DOCUMENT)
  const [previewParamsError, setPreviewParamsError] = useState<string>()
  const [parameterReferenceState, setParameterReferenceState] = useState<ParameterReferenceState>({ status: 'pending', names: [] })
  const [sampleData, setSampleData] = useState<SampleData | undefined>(initialSampleData)
  const [sampleError, setSampleError] = useState<string>()
  const [sampleBusy, setSampleBusy] = useState(false)
  const [bindingError, setBindingError] = useState<BindingErrorScope>()
  const [bindingBusy, setBindingBusy] = useState(false)
  const [assetError, setAssetError] = useState<Readonly<{ id: string; message: string }>>()
  const [assetBusy, setAssetBusy] = useState(false)
  const [tableEditor, setTableEditor] = useState<TableColumns>()
  const [tableEditorBusy, setTableEditorBusy] = useState(false)
  const [tableEditorError, setTableEditorError] = useState<string>()
  const snapshotRef = useRef(snapshot)
  const saveInFlight = useRef(false)
  const draftGeneration = useRef(0)
  const documentGeneration = useRef(0)
  const [documentGenerationValue, setDocumentGenerationValue] = useState(0)
  // The asset keys whose faces have ACTUALLY reached the page's font set.
  // Not the keys the document declares: a fragment may only ask for a derived
  // family once there is a face registered under it, or a failed fetch would
  // move it off the stylesheet's declared stack onto a family nothing
  // declares. See the registration effect below.
  const [carriedFaces, setCarriedFaces] = useState<ReadonlySet<string>>(NO_CARRIED_FACES)
  const selectedRef = useRef(selected)
  const previewToken = useRef(0)
  const previewAbort = useRef<AbortController | undefined>(undefined)
  const previewTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)
  const previewScheduler = useRef(new PreviewWorkScheduler())
  const previewRef = useRef<PreviewRecord | undefined>(undefined)
  const previewGeneration = useRef(0)
  const retryingFailure = useRef<number | undefined>(undefined)
  const previewNeedsFreshRender = useRef(false)
  const sampleDataRef = useRef(sampleData)
  const bindingInFlight = useRef(false)
	const tableEditorSession = useRef(0)
	const tableEditorInvoker = useRef<HTMLElement | undefined>(undefined)
  // Local picker results are authority-scoped independently of React renders.
  // A document replacement revokes a picker started for the old document.
  const sampleLoadGeneration = useRef(0)
  const previewParamsRef = useRef(previewParams)
  const parameterReferenceRequest = useRef(0)
  const modeRef = useRef(mode)
  const canvasRegionRef = useRef<HTMLElement>(null)
  useEffect(() => { modeRef.current = mode }, [mode])
  useEffect(() => { selectedRef.current = selected }, [selected])
  const canvas = snapshot?.canvas
  // STORY 8.4a — THE FACES THIS DOCUMENT CARRIES, REGISTERED ONCE FOR THE
  // WHOLE DOCUMENT.
  //
  // The engine measures and renders with a face the document carries in its
  // `assets` map, and now attributes each painted fragment to the asset it
  // resolved. The browser had no CSS family for such a face at all — no
  // `@font-face`, no name, no bytes — so the canvas rasterized at the engine's
  // x-positions with a fallback face's metrics and the glyphs collided. This
  // fetches those bytes over the SAME `asset` operation images already use and
  // registers each one under the family embedded-face-family.ts derives from
  // its key.
  //
  // IT IS DOCUMENT-SCOPED, WHICH IS THE ONE STRUCTURAL DECISION HERE.
  // ImagePaint is the nearest precedent and the wrong lifetime: it is mounted
  // once per component AND once per repeated sheet, so it makes an independent
  // request per mounted instance. `document.fonts` is a global, name-keyed
  // registry, so that shape would add one face many times under one family and
  // let an unmounting instance delete a face another is still painting with.
  //
  // THE LISTING IS PART OF THE KEY, AND `documentGenerationValue` ALONE IS NOT
  // ENOUGH. The generation advances only when the document is REPLACED (open a
  // file, new template, undo/redo); an ordinary font-chain command commits
  // through setCurrentSnapshot without it and can add or remove a carried
  // entry. The listing is over the carried keys only, sorted and de-duplicated
  // — precisely the input this effect reads — so a chain rename or a shipped
  // entry's edit does not needlessly re-register a face that is already good.
  const carriedFaceKeys = [...new Set((canvas?.fontChains ?? []).flatMap((chain) => chain.entries).map((entry) => entry.assetKey).filter((key) => key.length > 0))].sort()
  const carriedFaceListing = carriedFaceKeys.join('\u0000')
  useEffect(() => {
    setCarriedFaces(NO_CARRIED_FACES)
    if (!engine || carriedFaceListing === '') return
    // A rejected request or a response with no bytes is a document fact, not a
    // session fault: registerCarriedFaces reports the key as unregistered, the
    // fragment keeps the stylesheet's declared stack, and nothing reaches the
    // engine's failure channel.
    return registerCarriedFaces(carriedFaceListing.split('\u0000'), async (assetKey) => (await engine.request('asset', assetBytesRequest(assetKey))).bytes, setCarriedFaces)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [engine, documentGenerationValue, carriedFaceListing])
  const installPreview = (next: PreviewRecord | undefined) => { previewRef.current = next; setPreview(next) }
  const cancelPreviewWork = () => {
    previewToken.current++
    previewAbort.current?.abort()
    previewAbort.current = undefined
    if (previewTimer.current !== undefined) clearTimeout(previewTimer.current)
    previewTimer.current = undefined
    previewScheduler.current.clear()
    retryingFailure.current = undefined
  }
  const invalidatePreview = (clear = false) => {
    // A request already admitted to the FIFO worker must drain; invalidation
    // revokes its authority synchronously and lets the scheduler coalesce a
    // single newest replacement behind it instead of posting duplicates.
    previewToken.current++
    previewGeneration.current++
    if (clear) {
      installPreview(undefined)
      setPreviewViewState(initialPDFPreviewViewState)
      setPreviewStatus('idle')
    } else {
      setPreviewStatus(previewRef.current ? 'stale' : 'idle')
      setStaleReason('inputs-changed')
    }
    setPreviewError(undefined)
    setPreviewIssue(undefined)
    retryingFailure.current = undefined
    setDismissedDiagnostics(new Set())
  }
  const renderPreview = (force = false) => {
    if (!sampleDataRef.current) { setPreviewStatus('idle'); return }
    if (previewTimer.current !== undefined) clearTimeout(previewTimer.current)
    previewTimer.current = undefined
    previewScheduler.current.submit(() => runPreview(force))
  }
  const runPreview = async (force = false) => {
    const sample = sampleDataRef.current
    if (!engine || !snapshotRef.current || !sample) return
    const generation = previewGeneration.current
    const documentAtStart = documentGeneration.current
    const revisionAtStart = snapshotRef.current.revision
    // The engine receives the accepted file bytes, not the local inspection
    // projection. ArrayBuffer slicing is a transport copy, never a rewrite.
    const data = sample.bytes.slice(0)
    const params = new TextEncoder().encode(previewParamsRef.current).buffer
    const token = ++previewToken.current
    const controller = new AbortController()
    previewAbort.current = controller
    const current = (identity?: string) => token === previewToken.current && !controller.signal.aborted && modeRef.current === 'preview' && previewGeneration.current === generation && documentGeneration.current === documentAtStart && snapshotRef.current?.revision === revisionAtStart && (!identity || canInstallPreview({ token, generation, revision: revisionAtStart, identity }, { token: previewToken.current, generation: previewGeneration.current, revision: snapshotRef.current?.revision ?? -1, identity, mode: modeRef.current }))
    const mustRender = force || previewNeedsFreshRender.current
    setPreviewStatus(previewRef.current ? 'stale' : 'checking'); setPreviewError(undefined); setPreviewIssue(undefined)
    try {
      const checked = await engine.request('identity', { data, params }, controller.signal)
      const identity = checked.preview?.identity
      if (!identity || checked.preview.revision !== revisionAtStart || !current(identity)) return
      if (!mustRender && previewStatus === 'current' && previewRef.current && previewRef.current.identity === identity && previewRef.current.revision === revisionAtStart && previewRef.current.generation === generation) {
        setPreviewStatus('current')
        return
      }
      const canonical = await engine.request('serialize', undefined, controller.signal)
      if (!canonical.bytes) throw new Error('Current canonical document is unavailable')
      const revision = canonical.snapshot.revision
      if (revision !== revisionAtStart || !current(identity)) return
      setPreviewStatus(previewRef.current ? 'stale' : 'rendering')
      const result = await engine.request('render', { template: canonical.bytes, data, params }, controller.signal)
      if (!result.bytes || !result.preview?.pdfSha256 || !result.preview.diagnostics || result.preview.identity !== identity || result.preview.revision !== revision || !current(identity)) return
      installPreview({ bytes: result.bytes.slice(0), revision, identity, digest: result.preview.pdfSha256, diagnostics: result.preview.diagnostics, token, generation })
      previewNeedsFreshRender.current = false
      setDismissedDiagnostics(new Set())
      setPreviewViewState(initialPDFPreviewViewState)
      // PDF.js is a separate boundary. The bytes become current only when its
      // matching document has admitted successfully through onPageCount. The
      // candidate remains marked stale while it is visible but unconfirmed.
      setStaleReason('inputs-changed')
      setPreviewStatus('stale')
    } catch (error) {
      if (token !== previewToken.current || controller.signal.aborted) return
      if (!current()) return
      if (isProducerRenderFailure(error)) {
        setStaleReason('render-failed')
        setPreviewStatus(previewRef.current ? 'stale' : 'error'); setPreviewError({ error: previewFailure(error), token, generation, revision: revisionAtStart })
      } else {
        // This failed before/around the closed producer response boundary. It
        // remains a local Preview issue rather than invented render provenance.
        setStaleReason('inputs-changed')
        setPreviewStatus(previewRef.current ? 'stale' : 'error')
        setPreviewIssue(localPreviewIssue(error))
      }
    } finally {
      if (previewAbort.current === controller) previewAbort.current = undefined
    }
  }
  const schedulePreview = () => {
    if (modeRef.current !== 'preview') return
    if (previewTimer.current !== undefined) clearTimeout(previewTimer.current)
    setPreviewStatus(previewRef.current ? 'stale' : 'debouncing')
    previewTimer.current = setTimeout(() => { previewTimer.current = undefined; renderPreview() }, PREVIEW_DEBOUNCE_MS)
  }
  const acceptPreviewParameters = (draftValue: string) => {
    setPreviewParamsDraft(draftValue)
    try {
      const parsed: unknown = JSON.parse(draftValue)
      if (parsed === null || Array.isArray(parsed) || typeof parsed !== 'object') throw new Error('Parameter input must be a JSON object')
      if (new TextEncoder().encode(draftValue).byteLength > MAX_PARAMETER_DOCUMENT_BYTES) throw new Error('Parameter input exceeds the local engine limit')
      previewParamsRef.current = draftValue
      setPreviewParams(draftValue)
      setPreviewParamsError(undefined)
      invalidatePreview()
      schedulePreview()
    } catch (error) {
      setPreviewParamsError(error instanceof Error ? error.message : 'Parameter input must be valid JSON')
    }
  }
  const setNamedParameter = (name: string, value: string) => {
    try {
      const document: unknown = JSON.parse(previewParamsRef.current)
      JSON.parse(value)
      if (document === null || Array.isArray(document) || typeof document !== 'object') throw new Error('Parameter input must be a JSON object')
      // Never stringify the accepted document here. Its numeric lexemes and
      // untouched source bytes are runtime input evidence, not UI state to
      // normalize. The small JSON token locator only replaces this one value.
      const next = replaceTopLevelJSONValue(previewParamsRef.current, name, value)
      if (next === undefined) throw new Error('Parameter input must be a JSON object')
      acceptPreviewParameters(next)
    } catch {
      setPreviewParamsError(`Value for params.${name} must be valid JSON`)
    }
  }
  const clearPreviewParameters = (clearReferences = false) => {
    previewParamsRef.current = EMPTY_PARAMETER_DOCUMENT
    setPreviewParams(EMPTY_PARAMETER_DOCUMENT)
    setPreviewParamsDraft(EMPTY_PARAMETER_DOCUMENT)
    setPreviewParamsError(undefined)
    if (clearReferences) { parameterReferenceRequest.current++; setParameterReferenceState({ status: 'pending', names: [] }) }
  }
  const loadParameterReferences = async () => {
    const currentSnapshot = snapshotRef.current
    const generation = documentGeneration.current
    const request = ++parameterReferenceRequest.current
    if (!engine || !currentSnapshot) {
      setParameterReferenceState({ status: 'failed', names: [] })
      return
    }
    setParameterReferenceState({ status: 'pending', names: [] })
    try {
      const result = await engine.request('parameter-references')
      if (parameterReferenceRequest.current === request && documentGeneration.current === generation) {
        if (snapshotRef.current?.revision === currentSnapshot.revision && result.snapshot.revision === currentSnapshot.revision && result.parameterReferences?.revision === currentSnapshot.revision) setParameterReferenceState({ status: 'ready', names: result.parameterReferences.names })
        else setParameterReferenceState({ status: 'failed', names: [] })
      }
    } catch {
      // Never turn an unavailable projection into a guessed empty one.
      if (parameterReferenceRequest.current === request && documentGeneration.current === generation) setParameterReferenceState({ status: 'failed', names: [] })
    }
  }
  const enterPreview = () => {
    modeRef.current = 'preview'
    setMode('preview')
    void loadParameterReferences()
    renderPreview()
  }
  const returnToDesign = () => { cancelPreviewWork(); modeRef.current = 'design'; setPreviewStatus(previewRef.current ? 'stale' : 'idle'); setPreviewError(undefined); setMode('design') }
  const viewerError = useCallback((token: number, error: Error) => {
    // Invalidate before unmounting the failed viewer: an older PDF.js callback
    // must never turn a newer render into an error state.
    if (token !== previewToken.current || modeRef.current !== 'preview') return
    cancelPreviewWork()
    const current = previewRef.current
    if (!current) return
    previewNeedsFreshRender.current = true
    setPreviewError(undefined)
    setPreviewIssue(`The local PDF viewer could not display the admitted PDF: ${error.message.slice(0, 160)}`)
    setStaleReason('inputs-changed')
    setPreview((current) => { previewRef.current = current; setPreviewStatus(current ? 'stale' : 'error'); return current })
  }, [])
  const viewerPages = useCallback((token: number, pages: number) => {
    const current = previewRef.current
    if (pages > 0 && current && current.token === token && token === previewToken.current && modeRef.current === 'preview' && canInstallPreview({ token, generation: current.generation, revision: current.revision, identity: current.identity }, { token: previewToken.current, generation: previewGeneration.current, revision: snapshotRef.current?.revision ?? -1, identity: current.identity, mode: modeRef.current })) { previewNeedsFreshRender.current = false; setPreviewIssue(undefined); setPreviewStatus('current') }
  }, [])
  const changePreviewViewState = useCallback((next: PDFPreviewViewState) => setPreviewViewState((current) => samePDFPreviewViewState(current, next) ? current : next), [])
  const clearInteraction = () => { setPlacing(undefined); setPlacingAt(undefined); setHoverBand(undefined); setDrag(undefined) }
  const commitComponent = async (payload: ArrayBuffer, after?: () => void) => {
    if (!engine || fileBusy) return
    setCommitError(undefined)
    try {
      const priorRevision = snapshotRef.current?.revision
      const result = await engine.request('command', payload)
      if (result.snapshot.revision !== priorRevision) invalidatePreview()
      setCurrentSnapshot(result.snapshot)
      after?.()
    }
    catch (error) { setCommitError(componentDiagnostic(error)); clearInteraction() }
  }
  const openTableEditor = async (id: string) => {
    if (!engine || fileBusy || selectedRef.current.length !== 1 || selectedRef.current[0] !== id) return
    const generation = documentGeneration.current
    const revision = snapshotRef.current?.revision
		const session = ++tableEditorSession.current
		tableEditorInvoker.current = document.activeElement instanceof HTMLElement ? document.activeElement : undefined
    setTableEditorError(undefined); setTableEditorBusy(true)
    try {
      const result = await engine.request('table-columns', new TextEncoder().encode(JSON.stringify({ id })).buffer)
      if (tableEditorSession.current === session && documentGeneration.current === generation && selectedRef.current.length === 1 && selectedRef.current[0] === id && snapshotRef.current?.revision === revision && result.snapshot.revision === revision && result.tableColumns?.revision === revision && result.tableColumns.table.tableId === id) setTableEditor(result.tableColumns)
    } catch (error) { if (tableEditorSession.current === session && documentGeneration.current === generation) setTableEditorError(componentDiagnostic(error))
    } finally { if (tableEditorSession.current === session && documentGeneration.current === generation) setTableEditorBusy(false) }
  }
	const revokeTableEditor = () => {
		tableEditorSession.current++
		setTableEditor(undefined); setTableEditorError(undefined); setTableEditorBusy(false)
	}
	const closeTableEditor = () => {
		revokeTableEditor()
		queueMicrotask(() => {
			const invoker = tableEditorInvoker.current
			if (invoker?.isConnected) invoker.focus()
			else canvasRegionRef.current?.focus()
		})
	}
  const commitTableColumn = async (payload: ArrayBuffer) => {
    const current = tableEditor
    if (!engine || fileBusy || !current || tableEditorBusy) return
    const generation = documentGeneration.current
    const revision = snapshotRef.current?.revision
    const id = current.table.tableId
		const session = tableEditorSession.current
    let accepted = false
    setTableEditorError(undefined); setTableEditorBusy(true)
    try {
      const committed = await engine.request('command', payload)
			accepted = true
			// A committed canonical document is not scoped to transient selection or
			// editor visibility. Admit it whenever it follows the expected document
			// generation/revision; only the editor's re-projection remains scoped.
			if (documentGeneration.current === generation && snapshotRef.current?.revision === revision) {
				if (committed.snapshot.revision !== revision) invalidatePreview()
				setCurrentSnapshot(committed.snapshot)
			}
      const projected = await engine.request('table-columns', new TextEncoder().encode(JSON.stringify({ id })).buffer)
      if (tableEditorSession.current === session && documentGeneration.current === generation && selectedRef.current.length === 1 && selectedRef.current[0] === id && snapshotRef.current?.revision === committed.snapshot.revision && projected.snapshot.revision === committed.snapshot.revision && projected.tableColumns?.revision === committed.snapshot.revision && projected.tableColumns.table.tableId === id) setTableEditor(projected.tableColumns)
			else if (tableEditorSession.current === session) revokeTableEditor()
    } catch (error) { if (accepted) { if (tableEditorSession.current === session) revokeTableEditor() } else if (tableEditorSession.current === session && documentGeneration.current === generation && selectedRef.current.length === 1 && selectedRef.current[0] === id && snapshotRef.current?.revision === revision) setTableEditorError(componentDiagnostic(error))
    } finally { if (tableEditorSession.current === session && documentGeneration.current === generation) setTableEditorBusy(false) }
  }
  const bindPickedPath = async (segments: ReadonlyArray<string>) => {
    const id = selectedRef.current.length === 1 ? selectedRef.current[0] : undefined
    if (!engine || fileBusy || !id || bindingInFlight.current) return
    const requestGeneration = documentGeneration.current
    const priorRevision = snapshotRef.current?.revision
    const requestSample = sampleDataRef.current
    const requestSegments = [...segments]
    setBindingError(undefined)
    bindingInFlight.current = true
    setBindingBusy(true)
    try {
      const result = await engine.request('command', bindComponentScalarCommand(id, segments))
      // Selection is transient and cannot revoke an already committed engine
      // command. Only a document replacement or a newer authoritative view
      // can prevent this response from becoming the current projection.
      if (documentGeneration.current === requestGeneration && snapshotRef.current?.revision === priorRevision) {
        if (result.snapshot.revision !== priorRevision) invalidatePreview()
        setCurrentSnapshot(result.snapshot)
      }
    } catch (error) {
      if (requestSample && documentGeneration.current === requestGeneration && sampleDataRef.current === requestSample && selectedRef.current.length === 1 && selectedRef.current[0] === id && snapshotRef.current?.revision === priorRevision) setBindingError({ sample: requestSample, componentID: id, segments: requestSegments, message: componentDiagnostic(error) })
    } finally {
      bindingInFlight.current = false
      if (documentGeneration.current === requestGeneration) setBindingBusy(false)
    }
  }
  const place = (x: number, y: number) => {
    if (!placing) return
    const kind = placing
    clearInteraction()
    void commitComponent(dropComponentCommand(kind, x, y, snapEnabled))
  }
  // THE SECOND PLACEMENT SPELLING, for the sheets that did not exist before
  // (Ruling H). `dropComponent` carries a PAGE point and Go hit-tests it, and
  // hitTestBand's rectangle is one page tall — so a point on sheet three
  // would resolve to whichever band of page ONE it happened to land in, and
  // be created there rather than refused. `createComponent` already exists on
  // the channel, already carries the band NAME and a band-relative
  // coordinate, and Go already handles it; the first sheet keeps today's
  // dropComponent payload byte for byte, because a single-page template must
  // behave exactly as it did.
  const placeInBand = (band: CanvasProjection['bands'][number]['name'], x: number, y: number) => {
    if (!placing) return
    const kind = placing
    clearInteraction()
    void commitComponent(createComponentCommand(kind, band, x, y, snapEnabled))
  }
  const select = (id: string, extend: boolean) => { setBindingError(undefined); revokeTableEditor(); setSelected((current) => extend ? (current.includes(id) ? current.filter((value) => value !== id) : [...current, id]) : [id]) }
  const deleteSelection = () => { if (selected.length === 1) void commitComponent(deleteComponentCommand(selected[0]!), () => { revokeTableEditor(); setSelected([]) }) }
  const duplicateSelection = () => { if (selected.length === 1) void commitComponent(duplicateComponentCommand(selected[0]!, snapEnabled)) }
  const nudgeSelection = (dx: number, dy: number) => {
    const component = snapshotRef.current?.canvas?.components.find((candidate) => candidate.id === selectedRef.current[0])
    if (component && selectedRef.current.length === 1) void commitComponent(moveComponentCommand(component.id, component.x + dx, component.y + dy, snapEnabled))
  }
  const applyPageSetup = async () => {
    if (!engine || !canvas || fileBusy) return
    setCommitError(undefined)
    try {
      const requestGeneration = draftGeneration.current
      const priorRevision = snapshotRef.current?.revision
      const result = await engine.request('command', pageSetupCommand(preset, orientation, preset === 'custom' ? draft.width : '0', preset === 'custom' ? draft.height : '0', draft))
      if (result.snapshot.revision !== priorRevision) invalidatePreview()
      setCurrentSnapshot(result.snapshot, draftGeneration.current !== requestGeneration)
    } catch (error) { setCommitError(pageSetupDiagnostic(error)) }
  }
  const applyProperties = async (ids: ReadonlyArray<string>, intent: PropertyIntent, responseGeneration: number, selectionKey: string): Promise<CanvasProjection | undefined> => {
    if (!engine || fileBusy) return undefined
    setCommitError(undefined)
    setPropertyError(undefined)
    // Symmetry with applyFontChain, which clears propertyError's sibling: a
    // refusal is about the edit that caused it, and leaving a stale chain
    // refusal standing in the TYPOGRAPHY section while the author goes on
    // committing font size or bold shows them a sentence about an edit they
    // have moved past.
    setFontChainError(undefined)
    try {
      const priorRevision = snapshotRef.current?.revision
      const result = await engine.request('command', updateComponentPropertiesCommand(ids, intent))
      if (result.snapshot.revision !== priorRevision) invalidatePreview()
      setHistoryAvailability(result.snapshot)
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
  // The chain editor's six commands travel the SAME path a property commit
  // takes — one opaque command, one revision, one undo entry — and differ only
  // in where a refusal is anchored: by the CONTROL the panel dispatched from,
  // never by what the engine returned. That is what lets an unlocated
  // ENGINE_REJECTED (componentFields arity, a projection bound) be shown in
  // exactly the same place a located COMPONENT_INVALID is, with one rule.
  // The message is the engine's own string and is stored unprefixed, because
  // the control already says where it belongs.
  const applyFontChain = async (payload: ArrayBuffer, control: FontChainControl, responseGeneration: number, selectionKey: string) => {
    if (!engine || fileBusy || fontChainBusy) return
    setCommitError(undefined)
    setFontChainError(undefined)
    setFontChainBusy(true)
    try {
      const priorRevision = snapshotRef.current?.revision
      const result = await engine.request('command', payload)
      if (result.snapshot.revision !== priorRevision) invalidatePreview()
      setHistoryAvailability(result.snapshot)
      if ((snapshotRef.current?.revision ?? -1) < result.snapshot.revision) setCurrentSnapshot(result.snapshot)
    } catch (error) {
      if (documentGeneration.current === responseGeneration && selectedRef.current.join(',') === selectionKey) setFontChainError({ control, selectionKey, message: componentDiagnosticDetail(error).message })
    } finally {
      if (documentGeneration.current === responseGeneration) setFontChainBusy(false)
    }
  }

  // Story 5.13: choosing a local image is a two-step boundary crossing — the
  // browser reads bytes (imageFileAccess), then sends ONE opaque committed
  // command carrying those bytes and the browser's own declared media type
  // (AC1). This function does not hash, sniff, or decide legality; Go alone
  // does, through the ordinary command/diagnostic path every other mutation
  // already uses.
  const applyImageAsset = async (id: string) => {
    if (!engine || !imageFileAccess || fileBusy || assetBusy) return
    // Finding 4 (review of 2026-08-29): this closure spans the two longest
    // awaits in the application — an OS file dialog, then an engine command
    // carrying up to megabytes — and element ids are reused across
    // documents. Capture the generation/revision BEFORE the picker await,
    // matching bindPickedPath's shape exactly, so a result that resolves
    // after an Open/Start-blank/undo/newer-command document replacement is
    // never installed over the authoritative snapshot (AC1's named red
    // proof: "a command result installed after document replacement").
    const requestGeneration = documentGeneration.current
    const priorRevision = snapshotRef.current?.revision
    setAssetError(undefined)
    setAssetBusy(true)
    try {
      const picked = await imageFileAccess.openImage()
      const result = await engine.request('command', setComponentAssetCommand(id, picked.mediaType, picked.bytes))
      if (documentGeneration.current === requestGeneration && snapshotRef.current?.revision === priorRevision) {
        if (result.snapshot.revision !== priorRevision) invalidatePreview()
        setCurrentSnapshot(result.snapshot)
      }
    } catch (error) {
      if (!isFileAccessCancelled(error) && documentGeneration.current === requestGeneration && snapshotRef.current?.revision === priorRevision) {
        setAssetError({ id, message: componentDiagnostic(error) })
      }
    } finally {
      if (documentGeneration.current === requestGeneration) setAssetBusy(false)
    }
  }

  const setHistoryAvailability = (next: EngineSnapshot | undefined) => { setUndoAvailable(next?.canUndo === true); setRedoAvailable(next?.canRedo === true) }
  const setCurrentSnapshot = (next: EngineSnapshot | undefined, keepNewerDraft = false, clearDocumentInteraction = false) => { snapshotRef.current = next; setSnapshot(next); setHistoryAvailability(next); if (clearDocumentInteraction) { documentGeneration.current++; tableEditorSession.current++; setDocumentGenerationValue(documentGeneration.current); setSelected([]); setBindingError(undefined); setBindingBusy(false); setTableEditor(undefined); setTableEditorError(undefined); setAssetError(undefined); setAssetBusy(false); setFontChainError(undefined); setFontChainBusy(false); clearInteraction() }; if (next?.canvas) { setPreset(next.canvas.preset); setOrientation(next.canvas.orientation); if (!keepNewerDraft) setDraft(draftFor(next.canvas)) } }
  const updateDraft = (key: keyof Draft, value: string) => { draftGeneration.current++; setDraft((current) => ({ ...current, [key]: value })) }
  const announceFailure = (message: string) => { setFileStatus(undefined); setFileError(message) }
  const revokeSampleLoad = () => { sampleLoadGeneration.current++; setSampleBusy(false) }
  const clearSampleData = () => {
    revokeSampleLoad()
    sampleDataRef.current = undefined; setSampleData(undefined); setSampleError(undefined); setBindingError(undefined)
    clearPreviewParameters()
    // Clearing a previously accepted sample is itself a Preview input change.
    invalidatePreview(true)
  }
  const loadSample = async () => {
    if (!sampleFileAccess || sampleBusy) return
    const authority = ++sampleLoadGeneration.current
    setSampleBusy(true); setSampleError(undefined)
    try {
      const selected = await sampleFileAccess.openSample()
      const accepted = acceptSampleData(selected.name, selected.bytes)
      if (authority !== sampleLoadGeneration.current) return
      // Replacement is atomic: only a fully accepted raw file and its bounded
      // projection can replace the prior local sample.
      sampleDataRef.current = accepted; setSampleData(accepted); setBindingError(undefined)
      clearPreviewParameters()
      invalidatePreview()
      schedulePreview()
    } catch (error) {
      if (authority === sampleLoadGeneration.current && !isFileAccessCancelled(error)) setSampleError(error instanceof Error ? error.message : 'Could not read local sample data')
    } finally { if (authority === sampleLoadGeneration.current) setSampleBusy(false) }
  }
  const open = async () => {
    if (!engine || !fileAccess || fileBusy) return
    revokeSampleLoad()
    invalidatePreview(true)
    clearPreviewParameters(true)
    setFileBusy(true); setFileError(undefined); setFileStatus('Opening local file…')
    try {
      const opened = await fileAccess.open()
      const loaded = await engine.request('load', opened.bytes)
      const canonical = await engine.request('serialize')
      if (!canonical.bytes) throw new Error('Local file could not be serialized')
      const inputWasCanonical = equalBytes(opened.bytes, canonical.bytes)
      setCurrentSnapshot(loaded.snapshot, false, true)
      clearSampleData()
      setTitle(opened.name)
      setTarget(opened.target)
      setSavedRevision(inputWasCanonical ? canonical.snapshot.revision : undefined)
      setFileStatus(inputWasCanonical ? `Opened local file ${opened.name}` : `Opened local file ${opened.name}; canonical local changes need saving`)
      if (modeRef.current === 'preview') { void loadParameterReferences(); void renderPreview() }
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
    revokeSampleLoad()
    invalidatePreview(true)
    clearPreviewParameters(true)
    setFileBusy(true); setFileError(undefined); setFileStatus('Starting blank local template…')
    try {
      const loaded = await engine.request('load', blankBytes)
      setCurrentSnapshot(loaded.snapshot, false, true)
      clearSampleData()
      setTitle('Untitled template'); setTarget(undefined); setSavedRevision(undefined)
      setFileStatus('Started an unnamed local template')
      if (modeRef.current === 'preview') { void loadParameterReferences(); void renderPreview() }
    } catch { announceFailure('Could not start a blank local template')
    } finally { setFileBusy(false) }
  }

  const applyHistory = async (operation: 'undo' | 'redo') => {
    if (!engine || fileBusy || !(operation === 'undo' ? undoAvailable : redoAvailable)) return
    setCommitError(undefined)
    try {
      const result = await engine.request(operation)
      invalidatePreview(); setCurrentSnapshot(result.snapshot, false, true)
      // Undo/Redo installs a different canonical revision while Preview stays
      // open. Re-query the Go projection instead of keeping fields attributed
      // to the revision we just left.
      if (modeRef.current === 'preview') void loadParameterReferences()
    } catch (error) {
      const received = error as { code?: string }
      if (received.code === 'UNDO_UNAVAILABLE') setUndoAvailable(false)
      else if (received.code === 'REDO_UNAVAILABLE') setRedoAvailable(false)
      else setCommitError(componentDiagnostic(error))
    }
  }
  const returnWithOptionalSelection = (location?: DiagnosticLocation, announceUnavailable = false) => {
    const id = location?.elementId
    const current = snapshotRef.current?.canvas?.components
    clearInteraction()
    if (id && current?.some((component) => component.id === id)) { revokeTableEditor(); setSelected([id]); setLocateStatus(`Selected ${id} in Design.`) }
    else if (id && announceUnavailable) setLocateStatus('Locate unavailable: the authoritative element is no longer present.')
    else setLocateStatus(undefined)
    returnToDesign()
    setTimeout(() => canvasRegionRef.current?.focus(), 0)
  }
  const admittedPreview = (candidate: PreviewRecord | undefined): candidate is PreviewRecord => Boolean(candidate && previewStatus === 'current' && modeRef.current === 'preview' && previewRef.current === candidate && candidate.token === previewToken.current && candidate.generation === previewGeneration.current && candidate.revision === snapshotRef.current?.revision && candidate.identity === previewRef.current.identity)
  const activeFailure = (candidate: PreviewFailureRecord | undefined): candidate is PreviewFailureRecord => Boolean(candidate && previewError === candidate && ['error', 'stale'].includes(previewStatus) && modeRef.current === 'preview' && candidate.token === previewToken.current && candidate.generation === previewGeneration.current && candidate.revision === snapshotRef.current?.revision)
  const locateDiagnostic = (candidate: PreviewRecord, location: DiagnosticLocation) => { if (admittedPreview(candidate)) returnWithOptionalSelection(location, true) }
  const retryFromFailure = (failure: PreviewFailureRecord) => {
    if (!activeFailure(failure) || retryingFailure.current === failure.token) return
    retryingFailure.current = failure.token
    // Bypass only the retained same-identity PDF shortcut. The scheduler, FIFO
    // worker, serializer, and PDF.js admission path remain unchanged.
    renderPreview(true)
  }
  const returnFromFailure = (failure: PreviewFailureRecord) => { if (activeFailure(failure)) returnWithOptionalSelection(failure.error, true) }

  useEffect(() => {
    const shortcut = (event: KeyboardEvent) => {
      const editing = isEditableTarget(event.target) || event.isComposing
      const mac = isMacPlatform()
      const modifier = primaryModifier(event, mac)
      if (modifier && event.key.toLowerCase() === 's' && engine && fileAccess && !fileBusy) {
        event.preventDefault()
        void save(false)
        return
      }
      if (editing) return
      if (modifier && event.key.toLowerCase() === 'z' && !event.shiftKey && undoAvailable) { event.preventDefault(); void applyHistory('undo'); return }
      if ((modifier && event.shiftKey && event.key.toLowerCase() === 'z' || !mac && modifier && event.key.toLowerCase() === 'y') && redoAvailable) { event.preventDefault(); void applyHistory('redo'); return }
      if (modifier && event.key.toLowerCase() === 'd' && modeRef.current === 'design' && selectedRef.current.length === 1) { event.preventDefault(); duplicateSelection(); return }
      if (modeRef.current === 'design' && selectedRef.current.length === 1 && ['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key)) { event.preventDefault(); const step = event.shiftKey ? 10_000 : 1_000; if (event.key === 'ArrowLeft') nudgeSelection(-step, 0); if (event.key === 'ArrowRight') nudgeSelection(step, 0); if (event.key === 'ArrowUp') nudgeSelection(0, -step); if (event.key === 'ArrowDown') nudgeSelection(0, step); return }
      if (event.altKey && event.key.toLowerCase() === 's' && modeRef.current === 'design') { event.preventDefault(); setSnapEnabled((value) => !value); return }
      if (event.altKey && event.key.toLowerCase() === 'p' && engine && snapshotRef.current) {
        event.preventDefault()
        if (mode === 'preview') returnToDesign()
        else enterPreview()
      }
    }
    window.addEventListener('keydown', shortcut)
    return () => window.removeEventListener('keydown', shortcut)
  })

  useEffect(() => () => cancelPreviewWork(), [])

  if (loadState && !engine) {
    if (engineMayStart(loadState) && engineState !== 'failed') return <main className="engine-starting" aria-label="Engine preparation"><p role="status" aria-live="polite" aria-label="Engine preparation status">Starting local engine</p></main>
    return <LoadScreen lifecycle={loadState} payload={payload} engineState={engineState} onRetry={onRetry} />
  }

  const shortcuts = shortcutHintsFor()
  const currentDiagnostics = previewStatus === 'current' && mode === 'preview' && preview?.revision === snapshot?.revision ? preview : undefined
  const currentFailure = previewError && ['error', 'stale'].includes(previewStatus) && mode === 'preview' && previewError.revision === snapshot?.revision ? previewError : undefined
  const engineLabel = initializationError ? 'ENGINE UNAVAILABLE' : snapshot ? `GO SNAPSHOT · REVISION ${snapshot.revision}` : 'ENGINE STARTING'
  const offlineLabel = import.meta.env.DEV && offlineState === 'dev-bypass' ? 'Offline layer bypassed (dev)' : offlineState === 'ready' ? 'Offline ready' : offlineState === 'checking' ? 'Offline cache checking' : offlineState === 'update-available' ? 'Update available; current release remains usable' : 'Offline cache unavailable'
  const dirty = !snapshot || savedRevision === undefined || snapshot.revision !== savedRevision
  const saveLabel = dirty ? 'Unsaved local changes' : 'Saved local file'
  // ONE SHEET PER PROJECTED WINDOW, from the projection alone. The model is
  // built in sheet-stack.ts, which is pure arithmetic over Go's numbers: the
  // window origins, the window height and the page geometry. Nothing here
  // measures the DOM and nothing multiplies a window height by an index.
  const stack = canvas ? sheetStack(canvas) : undefined
  const disclosure = stack ? sheetStackDisclosure(stack) : undefined
  const sheetSurface = (projection: CanvasProjection, model: SheetStack, sheet: Sheet) => {
    const sheets = model.sheets.length
    // A single-sheet document must render exactly the DOM and exactly the
    // accessible names it rendered before this story, so the page qualifier
    // appears only once there is more than one page to be ambiguous about —
    // and RTL and Playwright both fail outright on a duplicate exact label.
    const many = sheets > 1
    const pageOf = many ? ` ${sheet.index + 1} of ${sheets}` : ''
    return <section key={sheet.index} className={`page-surface${gridVisible ? ' page-grid' : ''}`} aria-label={`Report page${pageOf} with Page Header, Content, and Page Footer`} style={pageStyle(projection, zoom)} onClick={() => { revokeTableEditor(); setSelected([]) }}>
      {projection.bands.map((band) => {
        const content = band.name === 'content'
        // The content band is one WINDOW of the column, so a component's
        // in-sheet position is its column position minus this window's
        // origin. A repeating band is the same band on every sheet, so its
        // origin is zero on all of them.
        const origin = content ? sheet.origin : 0
        // The two repeating bands are drawn on every sheet because the engine
        // repeats them — but exactly ONE occurrence of each of their
        // components is interactive and accessibly named, the same rule a
        // content component spanning two windows obeys (Ruling G). Two
        // identical accessible names for one component would break selection,
        // getByLabelText and Playwright's strict mode alike.
        const occurrences = content ? sheet.content : projection.components.filter((component) => component.band === band.name).map((component) => ({ component, y: component.y, home: sheet.index === 0 }))
        const target = `${sheet.index}:${band.name}`
        const paint = (occurrence: SheetOccurrence) => occurrence.home
          ? <CanvasComponent key={occurrence.component.id} component={occurrence.component} carriedFaces={carriedFaces} origin={origin} note={content && sheet.index > 0 ? canvasColumnPositionNotice(sheet.index + 1, sheets) : undefined} limit={{ band: band.name, width: band.width, height: band.height }} zoom={zoom} selected={selected.includes(occurrence.component.id)} preview={drag?.id === occurrence.component.id ? drag : undefined} engine={engine} generation={documentGenerationValue} trackColumn={content && many ? (edge: number, delta: number) => columnEdgeAfterDrag(model, projection, zoom, edge, delta) : undefined} onSelect={select} onDelete={deleteSelection} onDragStart={setDrag} onDragEnd={(finished) => { if (!finished.changed) { setDrag(undefined); return } const command = finished.mode === 'move' ? moveComponentCommand(occurrence.component.id, finished.x, finished.y, snapEnabled) : setComponentBoundsCommand(occurrence.component.id, finished.x, finished.y, finished.width, finished.height, snapEnabled); void commitComponent(command, () => setDrag(undefined)).finally(() => setDrag(undefined)) }} />
          : <ComponentEcho key={`${occurrence.component.id}@${sheet.index}`} component={occurrence.component} carriedFaces={carriedFaces} y={occurrence.y} zoom={zoom} engine={engine} generation={documentGenerationValue} />
        // A band holding a drag in progress stops clipping for the duration.
        // The dragged component's rendered position already tracks the
        // pointer down the whole stack, so clipping it to one window would
        // make it vanish at the very seam it is being dragged across. The
        // element is NOT moved in the tree to achieve this — a reparented
        // node loses its pointer capture mid-gesture — only the clip is
        // lifted, on the one band that needs it.
        const dragging = occurrences.some((occurrence) => drag?.id === occurrence.component.id)
        return <section key={band.name} className={`page-band page-band-${band.name}${hoverBand === target ? ' page-band-target' : ''}`} aria-label={many ? `${bandName(band.name)} on page ${sheet.index + 1} of ${sheets}` : bandName(band.name)} aria-current={hoverBand === target ? 'true' : undefined} style={bandStyle(band, zoom)} tabIndex={0} onPointerEnter={() => placing && setHoverBand(target)} onPointerLeave={() => setHoverBand((current) => current === target ? undefined : current)} onPointerUp={(event) => { if (placing && event.currentTarget === event.target) { const point = placementPoint(event.nativeEvent, band, zoom); if (sheet.index === 0) place(point.x, point.y); else placeInBand(band.name, point.x - band.x / 1000, origin / 1000 + point.y - band.y / 1000) } }} onKeyDown={(event) => { if (placing && (event.key === 'Enter' || event.key === ' ')) { event.preventDefault(); if (sheet.index === 0) place(band.x / 1000, band.y / 1000); else placeInBand(band.name, 0, origin / 1000) } }}><span>{bandName(band.name)}</span>{many ? <div className={`band-window${dragging ? ' band-window-open' : ''}`}>{occurrences.map(paint)}</div> : occurrences.map(paint)}{content && sheet.seam !== undefined ? <span className="page-seam" aria-hidden="true" style={{ '--seam-display-y': canvasDisplay.css(sheet.seam, zoom) } as CSSProperties} /> : undefined}</section>
      })}
    </section>
  }
  return <div className="app-shell" aria-label="Folio designer application shell" aria-busy={fileBusy}>
    <header className="document-bar" aria-label="Document bar">
      <span className="brand">FOLIO</span><span className="document-name">{title}</span><span className={`status-dot${dirty ? '' : ' status-clean'}`} aria-hidden="true" /><span className="status-copy" role="status">{saveLabel}</span>
      <div className="document-actions" aria-label="Local file actions"><button className="icon-button" type="button" onClick={() => void open()} disabled={!engine || !fileAccess || fileBusy} aria-label="Open local template"><Icon name="open" /></button><button className="icon-button" type="button" onClick={() => void save(false)} disabled={!engine || !fileAccess || fileBusy} aria-label="Save local template" title={`Save (${shortcuts.save})`}><Icon name="save" /></button><button className="file-button" type="button" onClick={() => void save(true)} disabled={!engine || !fileAccess || fileBusy}>Save As</button><button className="file-button" type="button" onClick={() => void startBlank()} disabled={!engine || !blankBytes || fileBusy}>Start blank</button><button className="file-button" type="button" onClick={() => void applyHistory('undo')} disabled={!undoAvailable || fileBusy}>Undo <kbd aria-hidden="true">{shortcuts.undo}</kbd></button><button className="file-button" type="button" onClick={() => void applyHistory('redo')} disabled={!redoAvailable || fileBusy}>Redo <kbd aria-hidden="true">{shortcuts.redo}</kbd></button></div>
      <span className="later-control" aria-label="Current page setup">{canvas ? `${canvas.preset} · ${canvas.orientation}` : 'Page setup unavailable'}</span>
      <div className="mode-switch" aria-label="Designer mode"><button className={mode === 'design' ? 'mode-active' : ''} type="button" aria-pressed={mode === 'design'} onClick={returnToDesign}>DESIGN</button><button className={mode === 'preview' ? 'mode-active' : ''} type="button" aria-pressed={mode === 'preview'} onClick={enterPreview}>PREVIEW <kbd aria-hidden="true">{shortcuts.preview}</kbd></button></div>
    </header>
    <div className="workbench" id="future-features">
      <nav className="palette-rail" aria-label="Component palette"><p className="section-label">PALETTE</p>{paletteItems.map(([label, kind]) => <button className="palette-item" type="button" key={kind} onPointerDown={() => { setPlacing(kind); setHoverBand(undefined) }} onClick={() => { setPlacing(kind); setHoverBand(undefined) }} aria-pressed={placing === kind} aria-label={`Place ${label}`}><PaletteIcon kind={kind} />{label}<kbd>place</kbd></button>)}<p className="honest-note">Choose or drag a component, then choose a page band.</p></nav>
      {mode === 'design' ? <main ref={canvasRegionRef} className={`canvas-region${placing ? ' canvas-region-placing' : ''}`} aria-label="Canvas region" tabIndex={0} onPointerMove={(event) => { if (placing) setPlacingAt({ x: event.clientX, y: event.clientY }) }} onPointerLeave={() => setPlacingAt(undefined)} onClick={(event) => { if (event.target === event.currentTarget) { revokeTableEditor(); setSelected([]) } }} onKeyDown={(event) => { if ((event.key === 'Delete' || event.key === 'Backspace') && event.target === event.currentTarget && selected.length === 1) { event.preventDefault(); deleteSelection() } if (event.key === 'Escape') { clearInteraction(); revokeTableEditor(); setSelected([]) } }}>
        <div className="canvas-tools" aria-label="Canvas controls"><button type="button" onClick={() => setZoom((value) => Math.max(0.5, value - 0.1))} aria-label="Zoom out">−</button><output aria-label="Canvas zoom">{Math.round(zoom * 100)}%</output><button type="button" onClick={() => setZoom((value) => Math.min(2, value + 0.1))} aria-label="Zoom in">+</button><button type="button" onClick={() => setGridVisible((value) => !value)} aria-pressed={gridVisible}>Grid {gridVisible ? 'on' : 'off'}</button><button type="button" onClick={() => setSnapEnabled((value) => !value)} aria-pressed={snapEnabled}>Snap {snapEnabled ? 'on' : 'off'} <kbd aria-hidden="true">{shortcuts.snap}</kbd></button><button type="button" onClick={duplicateSelection} disabled={selected.length !== 1}>Duplicate <kbd aria-hidden="true">{shortcuts.duplicate}</kbd></button><button type="button" onClick={deleteSelection} disabled={selected.length !== 1}>Delete <kbd aria-hidden="true">{shortcuts.delete}</kbd></button><span>Nudge <kbd aria-hidden="true">{shortcuts.nudge}</kbd></span></div>
        {disclosure && <p className="canvas-disclosure" role="status" aria-live="polite" aria-label="Canvas sheet disclosure">{disclosure}</p>}
        {canvas && stack ? (stack.sheets.length === 1 ? sheetSurface(canvas, stack, stack.sheets[0] as Sheet) : <div className="sheet-stack" style={{ '--sheet-stack-gap': `${SHEET_STACK_GAP}px` } as CSSProperties}>{stack.sheets.map((sheet) => sheetSurface(canvas, stack, sheet))}</div>) : <p className="canvas-awaiting" role="status">Waiting for Go page geometry.</p>}
        {placing && placingAt && <span className="placement-ghost" aria-hidden="true" style={{ '--ghost-x': `${placingAt.x}px`, '--ghost-y': `${placingAt.y}px` } as CSSProperties}><PaletteIcon kind={placing} />{paletteItems.find(([, kind]) => kind === placing)?.[0]}</span>}
        {commitError && <p role="alert" className="file-message">{commitError}</p>}{fileError && <p role="alert" className="file-message">{fileError}</p>}{fileStatus && <p role="status" aria-live="polite" className="file-message">{fileStatus}</p>}{locateStatus && <p role="status" aria-live="polite" className="file-message">{locateStatus}</p>}
      </main> : <main className="preview-region" aria-label="Preview region"><div className="preview-heading"><p>{previewStatus === 'current' ? 'EXACT LOCAL PRODUCTION PDF' : 'LOCAL PDF PREVIEW'}</p><button type="button" className="file-button" onClick={returnToDesign}>{['checking', 'debouncing', 'rendering'].includes(previewStatus) ? 'Cancel and return to Design' : 'Return to Design'}</button></div><p id="preview-freshness-status" className="preview-status" role="status" aria-live="polite" aria-atomic="true">{!sampleData ? 'Preview unavailable: no sample data loaded' : previewStatus === 'current' ? 'Current exact local PDF' : previewStatus === 'stale' ? `${staleCopy(staleReason)}${currentFailure ? `; local PDF render failed: ${currentFailure.error.message}` : previewIssue ? `; ${previewIssue}` : ''}` : ['checking', 'debouncing', 'rendering'].includes(previewStatus) ? 'Rendering local PDF' : previewStatus === 'error' ? `Local Preview work failed${previewIssue ? `: ${previewIssue}` : currentFailure ? `: ${currentFailure.error.message}` : ''}` : 'Preview is waiting for local inputs'}</p>{currentFailure && <PreviewFailure error={currentFailure.error} onRetry={() => retryFromFailure(currentFailure)} onReturn={() => returnFromFailure(currentFailure)} />}{preview && <><PDFPreviewViewer bytes={preview.bytes} label={previewStatus === 'current' ? `Current exact local production PDF, revision ${preview.revision}` : `Stale historical PDF, revision ${preview.revision}`} describedBy="preview-freshness-status" state={previewViewState} onStateChange={changePreviewViewState} onError={(error) => viewerError(preview.token, error)} onPageCount={(pages) => viewerPages(preview.token, pages)} />{currentDiagnostics && <PreviewDiagnostics diagnostics={currentDiagnostics.diagnostics} dismissed={dismissedDiagnostics} onDismiss={(key) => setDismissedDiagnostics((current) => new Set([...current, key]))} onLocate={(location) => locateDiagnostic(currentDiagnostics, location)} />}</>}<p className="preview-evidence">{preview ? `Historical producer digest ${preview.digest}` : 'Go production digest pending'}{preview ? ` · ${preview.diagnostics.length} diagnostics retained` : ''}</p></main>}
      <aside className="inspector-panel" aria-label="Inspector">
        <div className="panel-tabs" role="tablist" aria-label="Inspector tabs">{inspectorTabs.map(([tab, designLabel, previewLabel]) => <button key={tab} type="button" role="tab" id={`inspector-tab-${tab}`} aria-controls={`inspector-panel-${tab}`} aria-selected={inspectorTab === tab} tabIndex={inspectorTab === tab ? 0 : -1} className={`panel-tab panel-tab-${tab}${inspectorTab === tab ? ' panel-tab-active' : ''}`} onClick={() => setInspectorTab(tab)} onKeyDown={(event) => { const next = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0; if (!next) return; event.preventDefault(); const order = inspectorTabs.map(([name]) => name); const target = order[(order.indexOf(tab) + next + order.length) % order.length]!; setInspectorTab(target); requestAnimationFrame(() => document.getElementById(`inspector-tab-${target}`)?.focus()) }}>{mode === 'preview' ? previewLabel : designLabel}</button>)}</div>
        <div className="panel-body" role="tabpanel" id="inspector-panel-properties" aria-label={mode === 'preview' ? 'Preview inputs' : 'Properties panel'} hidden={inspectorTab !== 'properties'}>{mode === 'preview' ? <><p className="section-label">PREVIEW INPUTS</p><ParameterEditor referenceState={parameterReferenceState} accepted={previewParams} draft={previewParamsDraft} error={previewParamsError} onDraft={acceptPreviewParameters} onNamedValue={setNamedParameter} /><button type="button" className="file-button" onClick={() => void renderPreview(true)} disabled={!sampleData}>Render local PDF</button><p className="honest-note">Parameters are local Preview input and are not part of the template.</p></> : selected.length > 0 && canvas ? <ComponentProperties key={`${documentGenerationValue}:${selected.join(',')}`} components={canvas.components.filter((component) => selected.includes(component.id))} fontFamilies={canvas.fontFamilies} fontChains={canvas.fontChains} defaultFontSize={canvas.defaultFontSize} onCommit={applyProperties} onFontChainCommand={(payload, control) => void applyFontChain(payload, control, documentGeneration.current, selected.join(','))} fontChainError={fontChainError} fontChainBusy={fontChainBusy || fileBusy} documentGeneration={documentGenerationValue} propertyError={propertyError} drag={drag} onEditTable={(id) => void openTableEditor(id)} onPickImage={(id) => void applyImageAsset(id)} imageAvailable={imageFileAccess !== undefined} assetBusy={assetBusy} assetError={assetError} /> : <PageSetup preset={preset} orientation={orientation} draft={draft} onPreset={setPreset} onOrientation={setOrientation} onDraft={updateDraft} onApply={applyPageSetup} disabled={!canvas || fileBusy} />}</div>
        <div className="panel-body" role="tabpanel" id="inspector-panel-data" aria-labelledby="inspector-tab-data" hidden={inspectorTab !== 'data'}><DataPanel sample={sampleData} error={sampleError} busy={sampleBusy} available={Boolean(sampleFileAccess)} selectedComponentId={selected.length === 1 ? selected[0] : undefined} selectedBinding={selected.length === 1 ? canvas?.components.find((component) => component.id === selected[0])?.binding : undefined} bindingError={bindingError} bindingBusy={bindingBusy} onLoad={() => void loadSample()} onConnect={(segments) => void bindPickedPath(segments)} /></div>
      </aside>
    </div>
    {tableEditor && <TableEditor projection={tableEditor} busy={tableEditorBusy} error={tableEditorError} candidates={tableSampleCandidates(sampleData?.tree)} sampleAvailable={Boolean(sampleData)} onClose={closeTableEditor} onAdd={(index) => void commitTableColumn(addTableColumnCommand(tableEditor.table.tableId, index))} onRemove={(columnId) => void commitTableColumn(removeTableColumnCommand(tableEditor.table.tableId, columnId))} onMove={(columnId, index) => void commitTableColumn(moveTableColumnCommand(tableEditor.table.tableId, columnId, index))} onUpdate={(columnId, field, value) => void commitTableColumn(updateTableColumnCommand(tableEditor.table.tableId, columnId, field, value))} onConfigure={(collection, alias) => void commitTableColumn(configureTableBindingCommand(tableEditor.table.tableId, collection, alias))} onBind={(columnId, field) => void commitTableColumn(updateTableColumnBindingCommand(tableEditor.table.tableId, columnId, field))} onFooter={(columnId, footer, footerOf, footerFormat) => void commitTableColumn(updateTableColumnFooterCommand(tableEditor.table.tableId, columnId, footer, footerOf, footerFormat))} />}
    <footer className="status-bar" aria-label="Status bar"><span>LOCAL SHELL</span><code data-testid="engine-snapshot">{engineLabel}</code><span className="status-spacer" /><span role="status" aria-live="polite" aria-label="Offline availability" data-testid="offline-status">{offlineLabel}</span><code>{mode.toUpperCase()} MODE</code></footer>
  </div>
}

function ParameterEditor({ referenceState, accepted, draft, error, onDraft, onNamedValue }: { referenceState: ParameterReferenceState; accepted: string; draft: string; error?: string; onDraft: (value: string) => void; onNamedValue: (name: string, value: string) => void }) {
  const values = parameterValues(accepted)
  const references = referenceState.names
  return <><p className="panel-heading">Runtime parameters</p>{referenceState.status === 'pending' ? <p className="honest-note" role="status">Discovering parameter references from the local engine…</p> : referenceState.status === 'failed' ? <p className="file-message" role="alert">The local engine could not provide parameter references. The raw parameter document is still available.</p> : references.length > 0 ? <div aria-label="Engine-discovered parameter references">{references.map((name) => <ParameterValueInput key={name} name={name} acceptedValue={values[name]} onAccept={onNamedValue} />)}</div> : <p className="honest-note">The local engine found no parameter references in this template.</p>}<label>Raw parameter JSON<textarea aria-label="Raw parameter JSON" aria-invalid={Boolean(error)} aria-describedby={error ? 'parameter-input-error' : undefined} value={draft} onChange={(event) => onDraft(event.target.value)} /></label>{error && <p id="parameter-input-error" role="alert" className="file-message">{error}. The last accepted parameter document remains in Preview.</p>}</>
}

function ParameterValueInput({ name, acceptedValue, onAccept }: { name: string; acceptedValue?: string; onAccept: (name: string, value: string) => void }) {
	const accepted = acceptedValue ?? ''
	return <label>params.{name}<input aria-label={`Value for params.${name}`} value={accepted} onChange={(event) => onAccept(name, event.target.value)} /></label>
}

function parameterValues(raw: string): Record<string, string> {
  const values = Object.create(null) as Record<string, string>
  for (const property of topLevelJSONProperties(raw) ?? []) values[property.name] = raw.slice(property.valueStart, property.valueEnd)
  return values
}

type JSONPropertySpan = Readonly<{ name: string; valueStart: number; valueEnd: number }>

// This is intentionally a token locator, not a template parser or a second
// parameter schema. JSON.parse remains the acceptance authority; these spans
// merely let a named control replace one top-level raw JSON value without
// rewriting numeric lexemes, whitespace, unrelated keys, or special keys.
function topLevelJSONProperties(raw: string): ReadonlyArray<JSONPropertySpan> | undefined {
  let cursor = skipJSONWhitespace(raw, 0)
  if (raw[cursor++] !== '{') return undefined
  const properties: JSONPropertySpan[] = []
  cursor = skipJSONWhitespace(raw, cursor)
  if (raw[cursor] === '}') return properties
  while (cursor < raw.length) {
    const keyStart = cursor
    const keyEnd = scanJSONString(raw, cursor)
    if (keyEnd === undefined) return undefined
    let name: unknown
    try { name = JSON.parse(raw.slice(keyStart, keyEnd)) } catch { return undefined }
    if (typeof name !== 'string') return undefined
    cursor = skipJSONWhitespace(raw, keyEnd)
    if (raw[cursor++] !== ':') return undefined
    cursor = skipJSONWhitespace(raw, cursor)
    const valueStart = cursor
    const valueEnd = scanJSONValue(raw, cursor)
    if (valueEnd === undefined) return undefined
    properties.push({ name, valueStart, valueEnd })
    cursor = skipJSONWhitespace(raw, valueEnd)
    if (raw[cursor] === '}') return properties
    if (raw[cursor++] !== ',') return undefined
    cursor = skipJSONWhitespace(raw, cursor)
  }
  return undefined
}

function replaceTopLevelJSONValue(raw: string, name: string, value: string): string | undefined {
  const properties = topLevelJSONProperties(raw)
  if (!properties) return undefined
  // Duplicate JSON keys resolve last-wins in the production decoder; retain
  // that same value while preserving every other occurrence byte-for-byte.
  const existing = properties.filter((property) => property.name === name).at(-1)
  if (existing) return raw.slice(0, existing.valueStart) + value + raw.slice(existing.valueEnd)
  const close = skipJSONWhitespace(raw, properties.length === 0 ? 1 : properties.at(-1)!.valueEnd)
  const beforeClose = raw.indexOf('}', close)
  if (beforeClose < 0) return undefined
  return raw.slice(0, beforeClose) + `${properties.length > 0 ? ',' : ''}${JSON.stringify(name)}:${value}` + raw.slice(beforeClose)
}

function skipJSONWhitespace(raw: string, cursor: number): number {
  while (cursor < raw.length && /[\t\n\r ]/.test(raw[cursor]!)) cursor++
  return cursor
}

function scanJSONString(raw: string, cursor: number): number | undefined {
  if (raw[cursor] !== '"') return undefined
  for (cursor++; cursor < raw.length; cursor++) {
    if (raw[cursor] === '\\') { cursor++; continue }
    if (raw[cursor] === '"') return cursor + 1
  }
  return undefined
}

function scanJSONValue(raw: string, cursor: number): number | undefined {
  const first = raw[cursor]
  if (first === '"') return scanJSONString(raw, cursor)
  if (first === '{' || first === '[') {
    const close = first === '{' ? '}' : ']'
    cursor = skipJSONWhitespace(raw, cursor + 1)
    if (raw[cursor] === close) return cursor + 1
    while (cursor < raw.length) {
      if (first === '{') {
        const keyEnd = scanJSONString(raw, cursor)
        if (keyEnd === undefined) return undefined
        cursor = skipJSONWhitespace(raw, keyEnd)
        if (raw[cursor++] !== ':') return undefined
        cursor = skipJSONWhitespace(raw, cursor)
      }
      const valueEnd = scanJSONValue(raw, cursor)
      if (valueEnd === undefined) return undefined
      cursor = skipJSONWhitespace(raw, valueEnd)
      if (raw[cursor] === close) return cursor + 1
      if (raw[cursor++] !== ',') return undefined
      cursor = skipJSONWhitespace(raw, cursor)
    }
    return undefined
  }
  const token = raw.slice(cursor).match(/^(?:true|false|null|-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?)/)?.[0]
  return token ? cursor + token.length : undefined
}

function PageSetup({ preset, orientation, draft, onPreset, onOrientation, onDraft, onApply, disabled }: { preset: string; orientation: string; draft: Draft; onPreset: (value: string) => void; onOrientation: (value: string) => void; onDraft: (key: keyof Draft, value: string) => void; onApply: () => void; disabled: boolean }) {
  return <><p className="section-label">PAGE SETUP</p><p className="honest-note">Component properties require a selection.</p><label>Preset<select aria-label="Page preset" value={preset} onChange={(event) => onPreset(event.target.value)}><option value="A4">A4</option><option value="Letter">Letter</option><option value="custom">Custom</option></select></label><label>Orientation<select aria-label="Page orientation" value={orientation} onChange={(event) => onOrientation(event.target.value)}><option value="portrait">Portrait</option><option value="landscape">Landscape</option></select></label>{preset === 'custom' && <><Field label="Width (pt)" value={draft.width} onChange={(value) => onDraft('width', value)}/><Field label="Height (pt)" value={draft.height} onChange={(value) => onDraft('height', value)}/></>}<Field label="Top margin (pt)" value={draft.top} onChange={(value) => onDraft('top', value)}/><Field label="Right margin (pt)" value={draft.right} onChange={(value) => onDraft('right', value)}/><Field label="Bottom margin (pt)" value={draft.bottom} onChange={(value) => onDraft('bottom', value)}/><Field label="Left margin (pt)" value={draft.left} onChange={(value) => onDraft('left', value)}/><button type="button" className="file-button" onClick={onApply} disabled={disabled}>Apply page setup</button><p className="honest-note">Grid and snap are editor preferences; document undo is available in the document bar.</p></>
}

type PanelComponent = CanvasProjection['components'][number]
type PropertyCommitError = Readonly<{ field: PropertyField; selectionKey: string; elementId?: string; dataPath?: string; message: string }>
type CommitProperties = (ids: ReadonlyArray<string>, intent: PropertyIntent, generation: number, key: string) => Promise<CanvasProjection | undefined>
// Panel sections mirror the UX design's inspector: an identity row, then
// POSITION / CONTENT / TYPOGRAPHY / BOX / BINDING. Each field keeps its exact
// engine field name and accessible label; only the presentation is grouped.
// `empty` is the engine's behaviour when the field carries no committed value —
// no border, no padding, no fill, always visible. It is shown as a placeholder,
// never as a value: the field stays empty and nothing is written to the
// document until the author types.
// `fx` marks a field the engine will read as an expression rather than as a
// literal, and says which of the two spellings it accepts — so the cue is on
// the exact fields that accept one, and never on a field where Go rejects a
// placeholder outright.
type FieldExpression = 'placeholder' | 'condition'
type FieldSpec = Readonly<{ field: PropertyField; label: string; affix?: string; unit?: string; swatch?: true; prose?: true; empty?: string; fx?: FieldExpression }>
const fxHint: Readonly<Record<FieldExpression, string>> = { placeholder: 'Accepts literal text, or {{ }} expressions', condition: 'Accepts a boolean data path or call, written without {{ }} — the grammar has no comparisons' }
// A condition field IS the expression, so any text in it is one; a text field
// holds an expression only where a placeholder is spelled.
function holdsExpression(fx: FieldExpression, text: string): boolean { return fx === 'placeholder' ? containsPlaceholder(text) : text !== '' }
const positionFields: ReadonlyArray<FieldSpec> = [{ field: 'x', label: 'X (pt)', affix: 'X', unit: 'pt' }, { field: 'y', label: 'Y (pt)', affix: 'Y', unit: 'pt' }]
const sizeFields: ReadonlyArray<FieldSpec> = [{ field: 'width', label: 'Width (pt)', affix: 'W', unit: 'pt' }, { field: 'height', label: 'Height (pt)', affix: 'H', unit: 'pt' }]
// One CONTENT field, not two. Go keeps two commands behind the same
// element.Value — `value` rejects a placeholder, `expression` requires one —
// and splitting the panel along that seam made the author pick the command
// before typing, with the same committed text showing in both rows. The field
// routes on what was actually typed instead; the engine's two guards, and the
// two spellings they accept, are unchanged.
// `prose` is what makes this a TEXTAREA rather than an <input type="text">,
// which cannot hold a line feed at all — the whole of Story 7.4's first AC.
// The flag was declared on FieldSpec long before anything set it; this is the
// field it was anticipated for, and it stays the only one that sets it.
const contentField: FieldSpec = { field: 'value', label: 'Text', affix: 'Text', prose: true, fx: 'placeholder' }
// The picker can only carry a well-formed #RRGGBB — exactly what Go's
// parseHexColor accepts — so an unset or half-typed field opens it on black
// while the text beside it stays the committed truth.
function swatchColor(text: string): string { return /^#[0-9a-fA-F]{6}$/.test(text) ? text : '#000000' }
function containsPlaceholder(text: string): boolean { return text.includes('{{') || text.includes('}}') }
function contentCommand(field: PropertyField, text: string): PropertyField { return field === 'value' && containsPlaceholder(text) ? 'expression' : field }
// TYPOGRAPHY is laid out as Main.dc.html draws it: the family row spans the
// panel, the size sits beside the B/I pair, and align/valign are the design's
// two segmented controls instead of free-text fields. Both are closed sets Go
// already validates, and the control offers exactly the values ITS OWN
// selection accepts and nothing else.
//
// That is no longer one list. Since Story 7.3 `style.align` admits FOUR
// values for text — left/center/right/justify — while a table's cells draw a
// justified value at the start edge, so it means nothing there. The align
// segments are therefore derived per selection in ComponentProperties, not a
// module constant; valign stays the one triple top/middle/bottom.
const fontSizeField: FieldSpec = { field: 'fontSize', label: 'Font size (pt)', unit: 'pt' }
// Story 7.4. A dimensionless ratio, shown in the author's own units: the
// engine carries thousandths and `points` already divides by 1000, so 1500
// reads back as "1.5". Empty is the leading the declared font chain itself
// rules, which is exactly a ratio of 1.
const lineSpacingField: FieldSpec = { field: 'lineSpacing', label: 'Line spacing', affix: 'Leading', empty: '1' }
// Story 10.1: the ink, in TYPOGRAPHY where the rest of the type lives —
// it colours the glyphs, not the box, so it belongs beside the family and
// the size rather than beside Background. Empty is the engine's own
// behaviour with no colour declared: the PDF's initial fill, black.
const colorField: FieldSpec = { field: 'color', label: 'Text colour', affix: 'Colour', swatch: true, empty: 'black' }
// BOX: a Border row, the edge set, then the Background and Visibility rows the
// design shows as label-and-value. The four padding rows are deliberately not
// here (owner's call, 2026-08-30): style.padding stays an engine property that
// a loaded document keeps and renders — the panel simply does not author it.
const borderFields: ReadonlyArray<FieldSpec> = [{ field: 'borderWidth', label: 'Border width (pt)', affix: 'Border', unit: 'pt', empty: 'none' }, { field: 'borderColor', label: 'Border colour', affix: 'Border colour', swatch: true, empty: 'none' }]
const backgroundField: FieldSpec = { field: 'background', label: 'Background', affix: 'Background', swatch: true, empty: 'none' }
const visibilityField: FieldSpec = { field: 'visibleIf', label: 'Visible if', affix: 'Visibility', empty: 'always', fx: 'condition' }
type SegmentSpec = Readonly<{ value: string; label: string; content: ReactNode }>
// The justify glyph is FOUR FLUSH RULES, drawn as an SVG path like its three
// siblings. It is never the CSS justify declaration: the browser must not be
// asked to justify anything, in production, unit or e2e sources, and
// canvas-authority-contract.test.ts bans the property/value pair outright —
// in comments too, which is why this sentence spells neither.
type AlignVariant = 'left' | 'center' | 'right' | 'justify'
const alignGlyphs: Readonly<Record<AlignVariant, string>> = { left: 'M2 4h12M2 8h8M2 12h11', center: 'M2 4h12M4 8h8M3 12h10', right: 'M2 4h12M6 8h8M3 12h11', justify: 'M2 3h12M2 7h12M2 11h12M2 15h12' }
function AlignIcon({ variant }: { variant: AlignVariant }) {
  return <svg aria-hidden="true" className="segment-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.2"><path d={alignGlyphs[variant]} /></svg>
}
const alignSegments: ReadonlyArray<SegmentSpec> = [{ value: 'left', label: 'Align left', content: <AlignIcon variant="left" /> }, { value: 'center', label: 'Align center', content: <AlignIcon variant="center" /> }, { value: 'right', label: 'Align right', content: <AlignIcon variant="right" /> }]
// Offered only when every selected component is text. A table element's
// style.align cascades to its cells, which draw a justified value at their
// start edge — so justify means nothing there, and a control must not offer a
// value that is meaningless for the element type. A MIXED text+table
// selection gets the triple too: one command goes to every id in it.
const justifySegment: SegmentSpec = { value: 'justify', label: 'Align justify', content: <AlignIcon variant="justify" /> }
const valignSegments: ReadonlyArray<SegmentSpec> = [{ value: 'top', label: 'Vertical align top', content: 'TOP' }, { value: 'middle', label: 'Vertical align middle', content: 'MID' }, { value: 'bottom', label: 'Vertical align bottom', content: 'BOT' }]
function PropertySection({ title, tone, children }: { title: string; tone?: 'bind'; children: ReactNode }) {
  return <section className={`property-section property-section-${title.toLowerCase()}${tone === 'bind' ? ' property-section-bind' : ''}`}><p className="section-label">{title}</p>{children}</section>
}
function ComponentProperties({ components, fontFamilies, fontChains, defaultFontSize, onCommit, onFontChainCommand, fontChainError, fontChainBusy, documentGeneration, propertyError, drag, onEditTable, onPickImage, imageAvailable, assetBusy, assetError }: { components: ReadonlyArray<PanelComponent>; fontFamilies: ReadonlyArray<string>; fontChains: CanvasProjection['fontChains']; defaultFontSize: number; onCommit: CommitProperties; onFontChainCommand: (payload: ArrayBuffer, control: FontChainControl) => void; fontChainError?: FontChainCommitError; fontChainBusy: boolean; documentGeneration: number; propertyError?: PropertyCommitError; drag?: DragState; onEditTable: (id: string) => void; onPickImage: (id: string) => void; imageAvailable: boolean; assetBusy: boolean; assetError?: Readonly<{ id: string; message: string }> }) {
  const ids = components.map((component) => component.id)
  const types = new Set(components.map((component) => component.type))
  const all = (predicate: (type: PanelComponent['type']) => boolean) => [...types].every(predicate)
  const single = components.length === 1 ? components[0]! : undefined
  const scopedError = propertyError?.selectionKey === ids.join(',') ? propertyError : undefined
  const scopedChainError = fontChainError?.selectionKey === ids.join(',') ? fontChainError : undefined
  // The chain editor is REVEALED, not opened: it renders inside this same
  // TYPOGRAPHY section, from a third button on the family control. AC1 forbids
  // a dialog, so there is no separate mode and no dialog stack — just a
  // disclosure whose state is local to this panel.
  const [chainsOpen, setChainsOpen] = useState(false)
  const table = single?.type === 'table' ? single : undefined
  const image = single?.type === 'image' ? single : undefined
  const typographic = all((type) => type === 'text' || type === 'table')
  // FOUR segments for an all-text selection, THREE for anything carrying a
  // table. SegmentedProperty never sees component.type — the widening is a
  // derivation here, where the selection's types are already known.
  const alignChoices = all((type) => type === 'text') ? [...alignSegments, justifySegment] : alignSegments
  // A drag is the same transient local proposal the canvas is already
  // painting, shown in the same units. It is never committed from here; the
  // pointer release sends the one command and Go's accepted geometry then
  // replaces it through the committed value below.
  const dragging = single && drag?.id === single.id ? drag : undefined
  const live = (field: PropertyField): string | undefined => dragging && (field === 'x' || field === 'y' || field === 'width' || field === 'height') ? points(dragging[field]) : undefined
  // The CONTENT field sends `value` or `expression` depending on the typed
  // text, so it owns the rejection of either command.
  const errorFor = (field: PropertyField) => scopedError && (scopedError.field === field || (field === 'value' && scopedError.field === 'expression')) ? scopedError : undefined
  const draftFor = (spec: FieldSpec) => <PropertyDraft key={spec.field} spec={spec} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} live={live(spec.field)} error={errorFor(spec.field)} />
  return <>
    <div className="component-identity">{single ? <PaletteIcon kind={single.type} /> : undefined}<span className="component-identity-name">{single ? single.type : `${components.length} selected`}</span><span className="component-identity-meta">{single ? `${single.id} · band: ${single.band}` : [...types].join(' · ')}</span></div>
    <PropertySection title="POSITION"><div className="property-grid">{positionFields.map(draftFor)}{all((type) => type !== 'table') && sizeFields.map(draftFor)}</div></PropertySection>
    {single && types.has('text') && <PropertySection title="CONTENT">{draftFor(contentField)}<p className="honest-note">Literal text, or {'{{ }}'} placeholders for data.</p></PropertySection>}
    {typographic && <PropertySection title="TYPOGRAPHY"><FontFamilyProperty families={fontFamilies} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'fontFamily' ? scopedError : undefined} chainsOpen={chainsOpen} onToggleChains={() => setChainsOpen((open) => !open)} />{chainsOpen && <FontChainEditor chains={fontChains} busy={fontChainBusy} error={scopedChainError} onCommand={onFontChainCommand} />}<div className="property-size-row">{draftFor({ ...fontSizeField, empty: points(defaultFontSize) })}<div className="property-toggle-row"><BooleanProperty label="Bold" field="bold" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'bold' ? scopedError : undefined} /><BooleanProperty label="Italic" field="italic" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'italic' ? scopedError : undefined} /></div></div>{draftFor(lineSpacingField)}{draftFor(colorField)}<div className="property-grid"><SegmentedProperty label="Align" field="align" segments={alignChoices} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'align' ? scopedError : undefined} /><SegmentedProperty label="Vertical align" field="valign" segments={valignSegments} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'valign' ? scopedError : undefined} /></div></PropertySection>}
    {image && <ImageSection component={image} onPick={onPickImage} available={imageAvailable} busy={assetBusy} error={assetError?.id === image.id ? assetError.message : undefined} />}
    <PropertySection title="BOX">{borderFields.map(draftFor)}<BorderEdgesProperty components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'borderEdges' ? scopedError : undefined} />{draftFor(backgroundField)}{draftFor(visibilityField)}<p className="honest-note">Visibility takes a boolean field or call — {'e.g. customer.isActive'}. Empty is always visible.</p></PropertySection>
    {table && <PropertySection title="TABLE"><button type="button" className="file-button" onClick={() => onEditTable(table.id)}>Configure columns</button><p className="honest-note">Table binding: {table.tableBind ?? 'Not set'} (display only)</p></PropertySection>}
    <PropertySection title="BINDING" tone="bind">{single?.binding ? <p className="binding-chip"><span className="binding-dot" aria-hidden="true" />Bound to <code>{single.binding}</code></p> : <p className="honest-note">{single ? 'No engine binding on this component. Pick a root scalar in the Data tab.' : 'Binding is shown for one selected component.'}</p>}</PropertySection>
    <p className="honest-note">{types.has('table') ? 'Table size and binding are not editable here; table geometry is derived from columns.' : 'Only committed engine values are shown. Arbitrary CSS is not editable here.'}</p>
  </>
}

// ImageSection is AC2's IMAGE section: it shows the CURRENTLY set asset's
// identity straight from the engine snapshot (never a local model of it),
// plus one named control that opens the local image picker. Every
// unavailable/failed state states its concrete reason in text — never
// colour alone — and the control is a plain, visibly-labelled button, so
// both the accessible name and keyboard/focus behaviour come from the same
// existing button pattern every other file/pick control in this panel uses.
function ImageSection({ component, onPick, available, busy, error }: { component: PanelComponent; onPick: (id: string) => void; available: boolean; busy: boolean; error?: string }) {
  const image = component.image
  return <PropertySection title="IMAGE">
    {image
      ? <p className="honest-note">{image.mediaType} · {image.width}×{image.height}px · asset {image.assetKey.slice(0, 12)}…</p>
      // Finding 9 (review of 2026-08-29): the paint field's absence used to
      // drive ONE fixed string here, which was FALSE for a dangling asset
      // reference — the media type is fine, the asset is simply gone.
      // Go's imageUnavailable discriminant (still one signal alongside the
      // absent paint, D-5.13.2) now says which of the two applies. With no
      // discriminant at all the box is simply unfilled: a placed image now
      // starts with a null asset and waits for a file.
      : <p className="honest-note" role="status">{component.imageUnavailable === 'missing' ? "This element's asset is not present in the document." : component.imageUnavailable === 'undecodable' ? "This version cannot render this asset's media type." : 'No image chosen yet. This box stays empty, and prints nothing, until you choose one.'}</p>}
    <button type="button" className="file-button" disabled={!available || busy} onClick={() => onPick(component.id)}>Choose image…</button>
    {!available && <p className="honest-note">No local file picker is available in this browser tier.</p>}
    {error && <p role="alert" className="property-error">{error}</p>}
  </PropertySection>
}

function committedValue(component: PanelComponent, field: PropertyField): string | undefined {
  const value = component[field === 'expression' ? 'value' : field as keyof PanelComponent]
  if (typeof value === 'number') return points(value)
  return typeof value === 'string' ? value : undefined
}
function PropertyDraft({ spec, components, ids, onCommit, documentGeneration, live, error }: { spec: FieldSpec; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; live?: string; error?: PropertyCommitError }) {
  const { field, label, affix, unit, swatch, prose, empty, fx } = spec
  const values = components.map((component) => committedValue(component, field))
  const same = values.every((value) => value === values[0])
  const committed = same ? values[0] ?? '' : ''
  const [draft, setDraft] = useState(committed)
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  // A canvas drag, resize or nudge commits geometry without this field ever
  // being touched. Follow the engine on a committed transition; a draft the
  // engine has not accepted (a rejected commit leaves the value unchanged)
  // still survives, because nothing transitioned.
  const [lastCommitted, setLastCommitted] = useState(committed)
  if (lastCommitted !== committed) { setLastCommitted(committed); setDraft(committed) }
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
  const commit = async () => { if (draft !== committed) await submit({ field: contentCommand(field, draft), operation: draft === '' && field !== 'value' && field !== 'expression' ? 'clear' : 'set', value: draft }, true) }
  // The design draws an unset row as empty chrome, and there is nothing to
  // clear on one: the reset action appears with the value it resets, and with a
  // mixed selection, which also has committed values behind it.
  const canClear = field !== 'x' && field !== 'y' && field !== 'width' && field !== 'height' && field !== 'value' && field !== 'expression' && (!same || (live ?? draft) !== '')
  const canNull = field === 'visibleIf' || field === 'background'
  const errorId = error ? `property-error-${field}` : undefined
  // The fx cue is a marker, not a control: it states, in the row itself, that
  // this field is read as an expression. The same sentence reaches a screen
  // reader through the input's description, so the cue is never colour- or
  // sight-only.
  const description = [same ? undefined : 'Mixed value', fx ? fxHint[fx] : undefined].filter((part) => part !== undefined).join('. ') || undefined
  // Enter COMMITS in a single-line field and INSERTS A LINE FEED in a prose
  // one — the one behaviour that differs between the two controls. Escape
  // still reverts and blurs, blur still commits, and the single-flight submit
  // and canonicalValue reconciliation are shared verbatim.
  const keyDown = (event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    if (event.key === 'Enter' && !prose) { event.preventDefault(); void commit() }
    if (event.key === 'Escape') { event.preventDefault(); revert(); event.currentTarget.blur() }
  }
  const proseField = useRef<HTMLTextAreaElement>(null)
  const proseCaret = useRef<number | undefined>(undefined)
  useLayoutEffect(() => {
    const caret = proseCaret.current
    if (caret === undefined || proseField.current === null) return
    proseCaret.current = undefined
    proseField.current.setSelectionRange(caret, caret)
  })
  // ONLY the plain flavour is ever read. A clipboard from a word processor
  // also carries text/html and text/rtf, which is where its fonts, bold,
  // italics and indents live; discarding them "without error" is achieved by
  // never looking at them, and never by adding a sanitiser — a parser would
  // be a new runtime dependency, and design-contract.test.ts pins the
  // lockfile. Paragraph breaks survive because they are in the plain text,
  // and a CRLF pair is folded into ONE mandatory break by the engine itself.
  const pasteProse = (event: ReactClipboardEvent<HTMLTextAreaElement>) => {
    // preventDefault FIRST, and UNCONDITIONALLY. A clipboard carrying only
    // text/html or text/rtf has no plain flavour to insert, and returning
    // before this call handed that paste to the BROWSER, which inserts text it
    // derived from the HTML — the one outcome "only the plain flavour is ever
    // read" exists to forbid. Nothing to insert must mean nothing inserted.
    event.preventDefault()
    const plain = event.clipboardData.getData('text/plain')
    if (plain === '') return
    const field = event.currentTarget
    const head = field.selectionStart ?? field.value.length
    const tail = field.selectionEnd ?? field.value.length
    setDraft(`${field.value.slice(0, head)}${plain}${field.value.slice(tail)}`)
    // The textarea is CONTROLLED, so React rewrites its value on the next
    // render and the caret goes to the end of the whole field. Pasting into
    // the middle of a long clause would then land the author's next keystroke
    // in the wrong paragraph, so the caret is put back at the end of what was
    // just inserted, in the layout effect that runs once the render commits.
    proseCaret.current = head + plain.length
  }
  const shared = { 'aria-label': label, 'aria-description': description, 'aria-invalid': error ? ('true' as const) : undefined, 'aria-errormessage': errorId, readOnly: live !== undefined, value: live ?? draft, placeholder: same ? empty : 'Mixed', disabled: pending, onBlur: () => void commit(), onKeyDown: keyDown }
  return <div className="property-editor"><div className={`property-field${prose ? ' property-field-prose' : ''}${live === undefined ? '' : ' property-field-live'}`}>{affix && <span className="property-affix">{affix}</span>}{prose
    ? <textarea ref={proseField} className="property-value property-value-prose" rows={4} {...shared} onChange={(event) => setDraft(event.target.value)} onPaste={pasteProse} />
    : <input className="property-value" {...shared} inputMode={unit === 'pt' || unit === undefined && (field === 'lineSpacing') ? 'decimal' : undefined} onChange={(event) => setDraft(event.target.value)} />}{fx && <span className={`property-fx${holdsExpression(fx, live ?? draft) ? ' property-fx-active' : ''}`} title={fxHint[fx]} aria-hidden="true">fx</span>}{swatch && <input type="color" className={`property-swatch${/^#[0-9a-fA-F]{6}$/.test(live ?? draft) ? '' : ' property-swatch-unset'}`} aria-label={`Pick ${label}`} value={swatchColor(live ?? draft)} disabled={pending || live !== undefined} onChange={(event) => { setDraft(event.target.value); void submit({ field, operation: 'set', value: event.target.value }, true) }} onBlur={() => void commit()} />}{unit && <span className="property-unit">{unit}</span>}{canClear && <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'clear' }, true)}>×</button>}{canNull && <button type="button" className="property-inline-action" aria-label={`Set ${label} null`} title={`Set ${label} null`} disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'null' }, true)}>∅</button>}</div>{error && <p id={errorId} role="alert" className="property-error">{error.elementId ? `${error.elementId}: ` : ''}{error.dataPath ? `${error.dataPath}: ` : ''}{error.message}</p>}</div>
}
function canonicalValue(canvas: CanvasProjection, ids: ReadonlyArray<string>, field: PropertyField): string | undefined { const values = canvas.components.filter((component) => ids.includes(component.id)).map((component) => committedValue(component, field)); return values.length === ids.length && values.every((value) => value === values[0]) ? values[0] ?? '' : undefined }
function BooleanProperty({ label, field, components, ids, onCommit, documentGeneration, error }: { label: string; field: 'bold' | 'italic'; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) {
  const values = components.map((component) => component[field])
  const uniform = values.every((value) => value === values[0])
  const active = uniform && values[0] === true
  const [pending, setPending] = useState(false); const pendingRef = useRef(false); const commit = async (operation: PropertyIntent['operation'], value?: boolean) => { if (pendingRef.current) return; pendingRef.current = true; setPending(true); await onCommit(ids, { field, operation, value }, documentGeneration, ids.join(',')); pendingRef.current = false; setPending(false) }
  const set = values.some((value) => value !== undefined)
  return <div className="property-editor"><div className="property-toggle-group"><button type="button" className="property-toggle" disabled={pending} aria-pressed={active} aria-label={uniform ? label : `${label}, mixed`} onClick={() => void commit('set', !active)}>{label.slice(0, 1)}</button>{!uniform && <span className="property-toggle-mixed" aria-hidden="true">·</span>}{set && <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={pending} onClick={() => void commit('clear')}>×</button>}</div>{error && <p role="alert" className="property-error">{error.message}</p>}</div>
}
// The font family is a closed set too, but a per-DOCUMENT one: style.fontFamily
// must name a declared, non-empty font chain, and Go now projects exactly those
// names (CanvasProjection.fontFamilies). So this is a search-and-select over the
// engine's own list rather than a free-text field whose every typo is a round
// trip to a rejection. The typed text filters; it is never committed as a value.
function FontFamilyProperty({ families, components, ids, onCommit, documentGeneration, error, chainsOpen, onToggleChains }: { families: ReadonlyArray<string>; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError; chainsOpen: boolean; onToggleChains: () => void }) {
  const values = components.map((component) => committedValue(component, 'fontFamily'))
  const uniform = values.every((value) => value === values[0])
  const committed = uniform ? values[0] ?? '' : ''
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [active, setActive] = useState(0)
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  const listId = useId()
  const needle = query.trim().toLowerCase()
  const matches = needle === '' ? families : families.filter((name) => name.toLowerCase().includes(needle))
  const close = () => { setOpen(false); setQuery(''); setActive(0) }
  const commit = async (intent: PropertyIntent) => {
    if (pendingRef.current) return
    pendingRef.current = true
    setPending(true)
    await onCommit(ids, intent, documentGeneration, ids.join(','))
    pendingRef.current = false
    setPending(false)
  }
  const choose = (name: string) => { close(); void commit({ field: 'fontFamily', operation: 'set', value: name }) }
  const move = (step: number) => { if (matches.length > 0) setActive((current) => (current + step + matches.length) % matches.length) }
  const errorId = error ? 'property-error-fontFamily' : undefined
  return <div className="property-editor property-combobox" onBlur={(event) => { if (!event.currentTarget.contains(event.relatedTarget as Node | null)) close() }}>
    <div className="property-field">
      <input className="property-value property-value-prose" role="combobox" aria-label="Font family" aria-expanded={open} aria-controls={listId} aria-autocomplete="list" aria-activedescendant={open && matches.length > 0 ? `${listId}-${active}` : undefined} aria-description={uniform ? undefined : 'Mixed value'} aria-invalid={error ? 'true' : undefined} aria-errormessage={errorId} disabled={pending} value={open ? query : committed} placeholder={!uniform ? 'Mixed' : open ? 'Search fonts' : families.length > 0 ? 'Choose a font' : 'No font declared'} onFocus={() => setOpen(true)} onChange={(event) => { setOpen(true); setQuery(event.target.value); setActive(0) }} onKeyDown={(event) => {
        if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); setOpen(true); move(event.key === 'ArrowDown' ? 1 : -1); return }
        if (event.key === 'Enter') { event.preventDefault(); const name = matches[active]; if (open && name) choose(name); return }
        if (event.key === 'Escape') { event.preventDefault(); close() }
      }} />
      <button type="button" className="property-inline-action property-disclosure" aria-label={open ? 'Hide fonts' : 'Show fonts'} title={open ? 'Hide fonts' : 'Show fonts'} disabled={pending} tabIndex={-1} onMouseDown={(event) => event.preventDefault()} onClick={() => (open ? close() : setOpen(true))}>{open ? '⌃' : '⌄'}</button>
      {/* THE THIRD BUTTON. The set this combobox searches is the document's
          own declared chains, so the place to edit that set is here, beside
          it — revealed in the same panel section, never in a dialog (AC1). */}
      <button type="button" className="property-inline-action" aria-label={chainsOpen ? 'Hide font chains' : 'Edit font chains'} title={chainsOpen ? 'Hide font chains' : 'Edit font chains'} aria-expanded={chainsOpen} disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => { close(); onToggleChains() }}>≡</button>
      {(committed !== '' || !uniform) && <button type="button" className="property-inline-action" aria-label="Clear Font family" title="Clear Font family" disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => { close(); void commit({ field: 'fontFamily', operation: 'clear' }) }}>×</button>}
    </div>
    {open && <ul className="property-options" id={listId} role="listbox" aria-label="Declared fonts">
      {matches.map((name, index) => <li key={name} id={`${listId}-${index}`} role="option" aria-selected={name === committed} className={`property-option${index === active ? ' property-option-active' : ''}`} onMouseDown={(event) => event.preventDefault()} onMouseEnter={() => setActive(index)} onClick={() => choose(name)}>{name}</li>)}
      {matches.length === 0 && <li className="property-option property-option-empty" role="presentation">{families.length === 0 ? 'This document declares no font chains.' : `No declared font matches "${query.trim()}".`}</li>}
    </ul>}
    {error && <p id={errorId} role="alert" className="property-error">{error.message}</p>}
  </div>
}
// The design's segmented control over one closed-set engine field. There is no
// separate clear button in the design and no room for one in a 1fr cell, so
// pressing the active segment again clears the property — the only way back to
// the value the document inherits, and the state the control already shows as
// "no segment pressed".
function SegmentedProperty({ label, field, segments, components, ids, onCommit, documentGeneration, error }: { label: string; field: 'align' | 'valign'; segments: ReadonlyArray<SegmentSpec>; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) {
  const values = components.map((component) => committedValue(component, field))
  const uniform = values.every((value) => value === values[0])
  const current = uniform ? values[0] : undefined
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  const commit = async (value: string) => {
    if (pendingRef.current) return
    pendingRef.current = true
    setPending(true)
    await onCommit(ids, current === value ? { field, operation: 'clear' } : { field, operation: 'set', value }, documentGeneration, ids.join(','))
    pendingRef.current = false
    setPending(false)
  }
  return <div className="property-editor"><div className="property-segmented" role="group" aria-label={uniform ? label : `${label}, mixed`}>{segments.map((segment) => <button key={segment.value} type="button" className="property-segment" disabled={pending} aria-pressed={current === segment.value} aria-label={segment.label} title={current === segment.value ? `${segment.label}, press again to clear` : segment.label} onClick={() => void commit(segment.value)}>{segment.content}</button>)}{!uniform && <span className="property-toggle-mixed" aria-hidden="true">·</span>}</div>{error && <p role="alert" className="property-error">{error.message}</p>}</div>
}
function BorderEdgesProperty({ components, ids, onCommit, documentGeneration, error }: { components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; error?: PropertyCommitError }) { const values = components.map((component) => (component.borderEdges ?? []).join(',')); const same = values.every((value) => value === values[0]); const [edges, setEdges] = useState<string[]>(same && values[0] ? values[0].split(',') : []); const pending = useRef(false); const update = async (next: string[]) => { if (pending.current) return; pending.current = true; setEdges(next); await onCommit(ids, { field: 'borderEdges', operation: next.length ? 'set' : 'clear', ...(next.length ? { value: next } : {}) }, documentGeneration, ids.join(',')); pending.current = false }; return <div className="property-editor"><div className="property-edges" role="group" aria-label="Border edges"><span className="property-affix">Edges</span>{['top', 'right', 'bottom', 'left'].map((edge) => <label key={edge}><input type="checkbox" aria-label={`Border ${edge}`} checked={edges.includes(edge)} onChange={() => void update(edges.includes(edge) ? edges.filter((value) => value !== edge) : [...edges, edge])} />{edge}</label>)}{!same && <span aria-label="Border edges mixed">Mixed</span>}{(edges.length > 0 || !same) && <button type="button" className="property-inline-action" aria-label="Clear Border edges" title="Clear Border edges" onClick={() => void update([])}>×</button>}</div>{error && <p role="alert" className="property-error">{error.message}</p>}</div> }

type Draft = { width: string; height: string; top: string; right: string; bottom: string; left: string }
function Field({ label, value, onChange }: { label: string; value: string; onChange: (value: string) => void }) { return <label>{label}<input aria-label={label} inputMode="decimal" value={value} onChange={(event) => onChange(event.target.value)} /></label> }
function bandName(name: CanvasProjection['bands'][number]['name']): string { return name === 'pageHeader' ? 'Page Header' : name === 'pageFooter' ? 'Page Footer' : 'Content' }
function points(value: number): string { const negative = value < 0; const magnitude = Math.abs(value); const whole = Math.floor(magnitude / 1000); const fraction = String(magnitude % 1000).padStart(3, '0').replace(/0+$/, ''); return `${negative ? '-' : ''}${whole}${fraction ? `.${fraction}` : ''}` }
function draftFor(canvas?: CanvasProjection): Draft { return canvas ? { width: points(canvas.commandWidth), height: points(canvas.commandHeight), top: points(canvas.marginTop), right: points(canvas.marginRight), bottom: points(canvas.marginBottom), left: points(canvas.marginLeft) } : { width: '', height: '', top: '', right: '', bottom: '', left: '' } }
// The canvas has one deliberately lossy display rounding rule. It maps only
// Go-owned millipoints plus local zoom; viewport, DPR, font metrics and DOM
// geometry are not inputs to painting or hit/drag proposals.
export const canvasDisplay = Object.freeze({
  css: (millipoints: number, zoom: number): string => `${Math.round(millipoints * zoom * 1000) / 1_000_000}px`,
  documentDelta: (pixels: number, zoom: number): number => Math.round((pixels / zoom) * 1000) / 1000,
})
// Pointer offsets are one local input event coordinate, not a layout query.
// They enter only this transient drop proposal and are converted through the
// same zoom mapping used by paint and drag; Go validates/snap-containes it.
export function placementPoint(event: Pick<MouseEvent, 'offsetX' | 'offsetY'>, band: CanvasProjection['bands'][number], zoom: number): Readonly<{ x: number; y: number }> {
  return { x: band.x / 1000 + canvasDisplay.documentDelta(event.offsetX, zoom), y: band.y / 1000 + canvasDisplay.documentDelta(event.offsetY, zoom) }
}
function pageStyle(canvas: CanvasProjection, zoom: number): CSSProperties { return { '--page-display-width': canvasDisplay.css(canvas.width, zoom), '--page-display-height': canvasDisplay.css(canvas.height, zoom), '--grid-display-pitch': canvasDisplay.css(canvas.gridIncrement, zoom), '--page-margin-left': canvasDisplay.css(canvas.marginLeft, zoom), '--page-margin-right': canvasDisplay.css(canvas.marginRight, zoom) } as CSSProperties }
function bandStyle(band: CanvasProjection['bands'][number], zoom: number): CSSProperties { return { '--band-x': canvasDisplay.css(band.x, zoom), '--band-y': canvasDisplay.css(band.y, zoom), '--band-width': canvasDisplay.css(band.width, zoom), '--band-height': canvasDisplay.css(band.height, zoom) } as CSSProperties }
function pageSetupDiagnostic(error: unknown): string { const received = error as { code?: string; dataPath?: string; message?: string }; if (received.code === 'PAGE_SETUP_INVALID') return received.dataPath ? `${received.dataPath}: ${received.message ?? 'invalid value'}` : received.message ?? 'Page setup is invalid.'; return 'Page setup is invalid. Check the selected size and margins.' }
function componentDiagnosticDetail(error: unknown): Readonly<{ elementId?: string; dataPath?: string; message: string }> { const received = error as { elementId?: string; dataPath?: string; message?: string }; return { ...(received.elementId ? { elementId: received.elementId } : {}), ...(received.dataPath ? { dataPath: received.dataPath } : {}), message: received.message ?? 'Component change was rejected.' } }
function componentDiagnostic(error: unknown): string { const received = componentDiagnosticDetail(error); const prefix = received.elementId ?? received.dataPath; return prefix ? `${prefix}: ${received.message}` : received.message }

type DragState = Readonly<{ id: string; mode: DragAnchor; startClientX: number; startClientY: number; x: number; y: number; width: number; height: number; originalX: number; originalY: number; originalWidth: number; originalHeight: number; changed: boolean }>
function CanvasComponent({ component, carriedFaces, origin, note, limit, zoom, selected, preview, engine, generation, trackColumn, onSelect, onDelete, onDragStart, onDragEnd }: { component: CanvasProjection['components'][number]; carriedFaces: ReadonlySet<string>; origin: number; note?: string; limit: DragLimit; zoom: number; selected: boolean; preview?: DragState; engine?: EngineClient; generation: number; trackColumn?: (edge: number, delta: number) => number; onSelect: (id: string, extend: boolean) => void; onDelete: () => void; onDragStart: (drag: DragState | undefined) => void; onDragEnd: (drag: DragState) => void }) {
  const selectedByPointer = useRef(false)
  const proposal = preview ?? { x: component.x, y: component.y, width: component.width, height: component.height }
  // Component geometry is COLUMN geometry, in every band; this sheet shows one
  // window of it, and `origin` is where that window begins. It is 0 for the
  // repeating bands and 0 for the first window, which is why a single-sheet
  // document paints at exactly the coordinates it painted at before.
  const active = { ...proposal, y: proposal.y - origin }
  const begin = (event: PointerEvent, mode: DragAnchor) => { event.stopPropagation(); selectedByPointer.current = true; onSelect(component.id, event.shiftKey); if (event.shiftKey) return; event.currentTarget.setPointerCapture?.(event.pointerId); onDragStart({ id: component.id, mode, startClientX: event.clientX, startClientY: event.clientY, x: component.x, y: component.y, width: component.width, height: component.height, originalX: component.x, originalY: component.y, originalWidth: component.width, originalHeight: component.height, changed: false }) }
  // Ruling I. A linear pixel delta knows nothing about the page footer, the
  // gap and the page header standing between one window's foot and the next
  // window's head, so across a seam the component drifts from the hand by
  // exactly that much. `trackColumn` is the stack's own inverse: it converts
  // the distance the pointer travelled DOWN THE STACK into the distance it
  // travelled down the COLUMN, applied to the edge this anchor actually
  // moves. What Go receives is still one opaque command carrying a column
  // coordinate.
  const move = (event: PointerEvent) => { if (!preview) return; const rawDX = event.clientX - preview.startClientX; const rawDY = event.clientY - preview.startClientY; const changed = preview.changed || Math.abs(rawDX) >= 2 || Math.abs(rawDY) >= 2; const dx = canvasDisplay.documentDelta(rawDX, zoom) * 1000; const travelled = canvasDisplay.documentDelta(rawDY, zoom) * 1000; const edge = preview.mode === 'sw' || preview.mode === 's' || preview.mode === 'se' ? preview.originalY + preview.originalHeight : preview.originalY; const dy = trackColumn ? trackColumn(edge, travelled) - edge : travelled; onDragStart({ ...preview, changed, ...proposedBounds(preview.mode, preview, dx, dy, limit) }) }
  const finish = (event: PointerEvent) => { if (!preview) return; event.stopPropagation(); onDragEnd(preview) }
  const paint = component.textPaint
  return <div className={`canvas-component canvas-component-${component.type}${paint?.overflow ? ' canvas-component-text-overflow' : ''}${selected ? ' canvas-component-selected' : ''}`} aria-label={componentAccessibleName(component, note)} role="button" tabIndex={0} style={componentStyle(active, zoom)} onClick={(event) => { event.stopPropagation(); if (!selectedByPointer.current) onSelect(component.id, event.shiftKey); selectedByPointer.current = false }} onKeyDown={(event) => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelect(component.id, event.shiftKey) } if (selected && (event.key === 'Delete' || event.key === 'Backspace')) { event.preventDefault(); event.stopPropagation(); onDelete() } }} onPointerDown={(event) => begin(event, 'move')} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => onDragStart(undefined)}><ComponentBox component={component} zoom={zoom} />{paint?.truncated ? <span className="canvas-text-truncated">{canvasTruncationNotice}</span> : undefined}{paint ? <TextPaint component={component} carriedFaces={carriedFaces} zoom={zoom} /> : component.type === 'image' ? <ImagePaint component={component} zoom={zoom} engine={engine} generation={generation} /> : component.type === 'table' ? 'Table' : ''}{selected && <span className="canvas-dimension" aria-hidden="true">{points(active.width)} × {points(active.height)}</span>}{selected && component.resizable && resizeAnchors.map((anchor) => anchor === 'se'
      ? <button key={anchor} type="button" className="resize-handle" aria-label={`Resize ${component.id}`} onPointerDown={(event) => begin(event, anchor)} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => onDragStart(undefined)} />
      : <span key={anchor} className={`selection-handle selection-handle-${anchor}`} aria-hidden="true" onPointerDown={(event) => begin(event, anchor)} onPointerMove={move} onPointerUp={finish} onPointerCancel={() => onDragStart(undefined)} />)}</div>
}
// ImagePaint is Story 5.13's canvas producer: it paints ONLY inside the
// fit-and-centre draw rectangle Go already computed (component.image), never
// a rectangle CSS or this component negotiates on its own (AD-17/guardrail
// 5) — object-fit is never used; the <img> is sized to the EXACT draw
// rectangle Go supplied, so the browser has no fitting decision left to
// make. Paintable bytes are fetched separately, per asset key, on a fresh
// effect run keyed by (assetKey, generation) — never cached across a
// document replacement — and the object URL this effect creates is
// revoked in its own cleanup, covering deletion (unmount), asset
// replacement (assetKey changes) and document replacement (generation
// changes) alike, with no accumulation across a session.
function ImagePaint({ component, zoom, engine, generation }: { component: CanvasProjection['components'][number]; zoom: number; engine?: EngineClient; generation: number }) {
  const image = component.image
  const [url, setUrl] = useState<string>()
  // Finding 13 (review of 2026-08-29): a failed per-key 'asset' fetch used
  // to leave `url` undefined forever, which rendered as "Loading image…"
  // permanently — the opposite of AC3's "honest placeholder" for a
  // component that cannot be painted. Track failure as its own state,
  // distinct from "still in flight", so the placeholder can say which one
  // it is. Causes include a stale/malformed key (Finding 12, now rejected
  // earlier by admission) and any transport failure.
  const [failed, setFailed] = useState(false)
  useEffect(() => {
    setUrl(undefined)
    setFailed(false)
    if (!engine || !image) return
    let active = true
    let created: string | undefined
    void engine.request('asset', assetBytesRequest(image.assetKey)).then((result) => {
      if (!active) return
      if (!result.bytes) { setFailed(true); return }
      created = URL.createObjectURL(new Blob([result.bytes], { type: image.mediaType }))
      setUrl(created)
    }).catch(() => { if (active) setFailed(true) })
    return () => { active = false; if (created) URL.revokeObjectURL(created) }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [engine, image?.assetKey, image?.mediaType, generation])
  if (!image) {
    // Finding 9: echo which of the two Go-side reasons applies, matching
    // ImageSection's text — one Go signal drives both surfaces. No reason at
    // all is the empty box a placed image starts as, which the design draws
    // as a dashed placeholder rather than a failure.
    if (!component.imageUnavailable) return <ImagePlaceholder>No image</ImagePlaceholder>
    return <ImagePlaceholder>{component.imageUnavailable === 'missing' ? 'Image unavailable — its asset is not present in the document' : 'Image unavailable — this version cannot render its media type'}</ImagePlaceholder>
  }
  const style: CSSProperties = { position: 'absolute', left: canvasDisplay.css(image.drawX - component.x, zoom), top: canvasDisplay.css(image.drawY - component.y, zoom), width: canvasDisplay.css(image.drawWidth, zoom), height: canvasDisplay.css(image.drawHeight, zoom) }
  if (url) return <img src={url} alt="" aria-hidden="true" draggable={false} className="canvas-image-paint" style={style} />
  return <ImagePlaceholder>{failed ? 'Image unavailable — could not load its bytes' : 'Loading image…'}</ImagePlaceholder>
}
// The page's placeholder frame: the palette's own image glyph over the reason
// this box has nothing to paint. The reason text is unchanged — Go states
// which of its two cases applies and this still echoes it.
function ImagePlaceholder({ children }: { children: string }) {
  return <span className="canvas-image-placeholder" aria-hidden="true"><svg className="canvas-placeholder-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.2" strokeLinecap="square">{paletteGlyphs.image}</svg><span>{children}</span></span>
}
// Display-only reading aid. Go painted this text and owns the expression
// grammar outright — scope, validation, evaluation, diagnostics. Tinting the
// delimiters it already painted decides nothing and parses nothing beyond the
// literal braces standing in the run.
const expressionRun = /(\{\{[^{}]*\}\})/g
function textRuns(text: string): ReadonlyArray<string> { return text.split(expressionRun).filter((part) => part !== '') }
function isExpressionRun(part: string): boolean { return part.startsWith('{{') && part.endsWith('}}') }
// STORY 8.4a — THE ONE FONT-FAMILY THE CANVAS EVER ASKS FOR BY NAME AT
// RUNTIME, and it is set PER FRAGMENT because a fragment is exactly one face
// by construction (the engine emits at most one run per face segment and never
// merges adjacent runs) while a component is not: a mixed-script element draws
// Latin through one chain entry and Thai through another.
//
// It is set only when the engine attributed this fragment to an asset the
// document CARRIES *and* that asset's face has actually reached the page's
// font set. Both halves matter. Without the first, a shipped face would be
// asked for under a family nothing declares; without the second, a fetch that
// failed would take the fragment OFF the stylesheet's declared stack — an
// inline declaration replaces the rule rather than extending it — and the
// canvas would paint with whatever the browser defaults to. Absent, the
// fragment falls to `.canvas-text-fragment`'s declared stack in App.css, which
// is exactly the shipped-face path and exactly the degrade path.
//
// The family is derived FROM THE ASSET KEY (D-8.4.1) by the one module that
// makes that decision. Nothing here reads `family`, `style`, `face` or a chain
// name, and the stylesheet still holds no document input at all.
//
// TextPaint is EXPORTED so the per-fragment face attribution can be asserted
// on a real DOM node rather than by scanning this file's text: what the canvas
// asks for is a rendered fact, and canvas-font-stack.test.ts reads it off the
// element.
export function TextPaint({ component, carriedFaces, zoom }: { component: CanvasProjection['components'][number]; carriedFaces: ReadonlySet<string>; zoom: number }) {
  const paint = component.textPaint!
  return <span className="canvas-text-paint" aria-hidden="true" style={{ '--text-font-size': canvasDisplay.css(component.fontSize ?? 12000, zoom), '--text-font-weight': component.bold ? 700 : 400, '--text-font-style': component.italic ? 'italic' : 'normal', ...(component.color === undefined ? {} : { '--text-ink': component.color }) } as CSSProperties}>{paint.lines.map((line, lineIndex) => <span className="canvas-text-line" key={`${component.id}-${lineIndex}`} style={{ '--text-line-baseline': canvasDisplay.css(line.baseline - component.y, zoom), '--text-line-advance': canvasDisplay.css(line.advance, zoom) } as CSSProperties}>{line.fragments.map((fragment, fragmentIndex) => <span className="canvas-text-fragment" key={`${component.id}-${lineIndex}-${fragmentIndex}`} style={{ '--text-fragment-x': canvasDisplay.css(fragment.x - component.x, zoom), ...(fragment.assetKey !== undefined && carriedFaces.has(fragment.assetKey) ? { fontFamily: embeddedFaceFamily(fragment.assetKey) } : {}) } as CSSProperties}>{textRuns(fragment.text).map((part, partIndex) => isExpressionRun(part) ? <span className="canvas-text-expression" key={`${component.id}-${lineIndex}-${fragmentIndex}-${partIndex}`}>{part}</span> : part)}</span>)}</span>)}</span>
}
// The engine says this element's paint is a PREFIX. It is stated in words, at
// the component, in the same sentence a screen reader gets — not by colour,
// and not only by a class. (The older `overflow` flag sets
// `canvas-component-text-overflow` and NOTHING ELSE: there is no CSS rule for
// that class anywhere, so it is invisible to the author. Noted, deliberately
// not fixed here; repeating the shape is what this avoids.)
//
// It says what is true of the CANVAS, and what is still true of the document:
// the value is intact and prints whole. Nothing here is derived from how many
// lines were painted — the canvas must never turn a truncated paint into a
// number about the document.
const canvasTruncationNotice = 'Canvas preview cut short. The whole text is in the document and prints in full.'
// AC4's PER-COMPONENT half, folded into the accessible name exactly as the
// truncation notice above is — the shipped idiom, and the reason this is an
// obligation rather than a tooltip nobody reads. A component that sits on a
// later sheet is not pinned there: the sheet it lands on is a consequence of
// everything above it in the column, and the canvas has no data, so the page
// it prints on can differ from the page drawn here.
const canvasColumnPositionNotice = (page: number, pages: number): string => `on canvas page ${page} of ${pages}, which is a consequence of the content above it and can change when the data does — a column position, not a pin to page ${page}`
// AC4's CANVAS-WIDE half. The claim the sheets make is narrower than the
// drawing looks: they show the pages this content column occupies AS THE
// CANVAS HAS LAID IT OUT, never a forecast of the printed document. Where the
// engine says the count is NOT EXACT — a bound table, a degraded pagination,
// text that could not be shaped, an element whose visibility depends on data —
// the printed document runs to some other number of pages, and the four
// shipped statement templates are one proof: byte-identical files that print
// one, five, twenty and fifty pages and project ONE window each.
const canvasColumnClaim = "A component's page is a consequence of the content above it and can change when the data does — it is a column position, not a pin to page three."
// ⚠ DIRECTION-FREE ON PURPOSE. The previous wording here said the document
// "prints more pages than are shown", which is true of a bound table and FALSE
// of an element the data hides — the engine places it and the render omits it,
// so the printed document is SHORTER. The causes do not agree on a direction
// and a document can carry two that disagree, so the only honest sentence is
// that the number is not the printed one.
const canvasInexactClaim = 'This count depends on data the canvas does not have, so the printed document can run to a different number of pages.'
function sheetStackDisclosure(stack: SheetStack): string | undefined {
  // Silent for the document that has nothing to disclose: one window, and an
  // exact count. Saying it anyway would be noise on every
  // single-page template, and AC5 requires that template's accessible surface
  // to be what it was.
  if (stack.sheets.length <= 1 && stack.isExact) return undefined
  const shown = stack.truncated ? `Showing the first ${stack.sheets.length} sheets of ${stack.windowCount}.` : `Showing ${stack.sheets.length} ${stack.sheets.length === 1 ? 'sheet' : 'sheets'}.`
  return `${shown} These are the pages this content column occupies as the canvas has laid it out, not a prediction of the printed document. ${canvasColumnClaim}${stack.isExact ? '' : ` ${canvasInexactClaim}`}`
}
function componentAccessibleName(component: CanvasProjection['components'][number], note?: string): string {
  const page = note ? `; ${note}` : ''
  if (component.type !== 'text') return `${component.type} component ${component.id}${page}`
  const text = component.textPaint?.lines.map((line) => line.fragments.map((fragment) => fragment.text).join('').trim()).filter(Boolean).join(' ').slice(0, 160)
  const binding = component.binding ? `; bound to ${component.binding}` : ''
  const cut = component.textPaint?.truncated ? `; ${canvasTruncationNotice}` : ''
  return text ? `text component ${component.id}: ${text}${binding}${cut}${page}` : `text component ${component.id}${binding}${cut}${page}`
}
// A component whose extent crosses a window boundary is drawn on every window
// it intersects, because leaving the later sheets empty would make the
// drawing a lie in a second way. Only the HOME occurrence carries the role,
// the tab stop and the name; these echoes are decoration — aria-hidden, no
// handlers, no handles — so one component never presents two identical
// accessible names (Ruling G).
function ComponentEcho({ component, carriedFaces, y, zoom, engine, generation }: { component: CanvasProjection['components'][number]; carriedFaces: ReadonlySet<string>; y: number; zoom: number; engine?: EngineClient; generation: number }) {
  const paint = component.textPaint
  return <span className={`canvas-component canvas-component-echo canvas-component-${component.type}`} aria-hidden="true" style={componentStyle({ x: component.x, y, width: component.width, height: component.height }, zoom)}><ComponentBox component={component} zoom={zoom} />{paint ? <TextPaint component={component} carriedFaces={carriedFaces} zoom={zoom} /> : component.type === 'image' ? <ImagePaint component={component} zoom={zoom} engine={engine} generation={generation} /> : component.type === 'table' ? 'Table' : ''}</span>
}
// Story 9.2: the box the engine paints — style.background and
// style.border — drawn on the canvas from the ENGINE's own projection, so
// what the author sees is what will print. Painted as a child beneath the
// content rather than on the component itself, which leaves the selection
// tint and the dotted placement outline the canvas already draws alone.
//
// The engine's own defaults are mirrored where a border declares less than
// all of itself: 0.5pt, #000000, all four edges (buildCellRectWithBackground-
// Field). A border declared with NO fields at all (`"border": {}`, reachable
// only by hand — the designer's own clear drops an empty border) projects
// nothing at all, so the canvas cannot see it; the PDF still draws it.
const boxEdges = ['top', 'right', 'bottom', 'left'] as const
function ComponentBox({ component, zoom }: { component: CanvasProjection['components'][number]; zoom: number }) {
  const bordered = component.borderWidth !== undefined || component.borderColor !== undefined || component.borderEdges !== undefined
  if (component.background === undefined && !bordered) return undefined
  const edges = component.borderEdges ?? boxEdges
  // The declared width, through the one zoom mapping every other painted
  // length goes through — not floored at a device pixel. A 0.5pt hairline
  // draws as the sub-pixel line it is, because inflating it would show the
  // author a border the PDF will not print.
  const stroke = `${canvasDisplay.css(component.borderWidth ?? 500, zoom)} solid ${component.borderColor ?? '#000000'}`
  const style: Record<string, string> = {}
  if (component.background !== undefined) style.background = component.background
  if (bordered) for (const edge of boxEdges) style[`border${edge[0]!.toUpperCase()}${edge.slice(1)}`] = edges.includes(edge) ? stroke : '0'
  return <span className="canvas-box" aria-hidden="true" style={style as CSSProperties} />
}
function componentStyle(component: { x: number; y: number; width: number; height: number }, zoom: number): CSSProperties { return { '--component-x': canvasDisplay.css(component.x, zoom), '--component-y': canvasDisplay.css(component.y, zoom), '--component-width': canvasDisplay.css(component.width, zoom), '--component-height': canvasDisplay.css(component.height, zoom) } as CSSProperties }

function equalBytes(left: ArrayBuffer, right: ArrayBuffer): boolean {
  const a = new Uint8Array(left)
  const b = new Uint8Array(right)
  return a.length === b.length && a.every((value, index) => value === b[index])
}

function previewFailure(error: EngineError): EngineError {
  // Values reached here were bounded and validated at the worker boundary.
  // Preserve their presence and content; do not turn absent provenance into a
  // browser-owned replacement code, message, or location.
  return { code: error.code, message: error.message, ...(error.elementId !== undefined ? { elementId: error.elementId } : {}), ...(error.dataPath !== undefined ? { dataPath: error.dataPath } : {}) }
}

function localPreviewIssue(error: unknown): string {
  return error instanceof Error && error.message ? error.message.slice(0, 512) : 'The local Preview work did not complete'
}

function isEditableTarget(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  return target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)
}
