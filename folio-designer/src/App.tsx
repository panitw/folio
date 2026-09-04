import './App.css'
import { useCallback, useEffect, useId, useLayoutEffect, useMemo, useRef, useState, type ClipboardEvent as ReactClipboardEvent, type CSSProperties, type KeyboardEvent as ReactKeyboardEvent, type PointerEvent, type ReactNode } from 'react'
import { isProducerRenderFailure, type EngineClient } from './engine-client'
import { MAX_LINE_SPACING_THOUSANDTHS, MIN_LINE_SPACING_THOUSANDTHS, type CanvasProjection, type EngineDiagnostic, type EngineError, type EngineSnapshot, type TableColumns } from './engine-protocol'
import type { OfflineLifecycleState } from './offline-lifecycle'
import type { OfflineLifecycle } from './offline-lifecycle'
import { engineMayStart } from './offline-lifecycle'
import type { S1Payload } from './release-payload'
import { LoadScreen } from './LoadScreen'
import type { BindingErrorScope } from './DataPanel'
import { isFileAccessCancelled, type FileAccess, type FileTarget } from './file/file-access'
import { pageSetupCommand } from './page-setup-command'
import { bindComponentScalarCommand, createComponentCommand, deleteComponentCommand, dropComponentCommand, duplicateComponentCommand, moveComponentCommand, setComponentBoundsCommand, type PaletteKind } from './component-command'
import { ORIGIN_FLOOR_FIELDS, POSITIVE_LENGTH_FIELDS, updateComponentPropertiesCommand, type PropertyField, type PropertyIntent } from './component-property-command'
import { FontBrowser } from './FontBrowser'
import { type FontChainCommitError, type FontChainControl } from './font-chain-control'
import { embedFontFamilyCommand } from './font-chain-command'
import { catalogueFaces, scriptFallbackFaces } from './generated/font-catalogue'
import { familyIsInstalled, indexRowFor, offeredFamilies, type FamilySource } from './font-index'
import { browserRows } from './font-browser-model'
import { fetchWebFamily } from './font-source'
import { openFontStore, storeWriteRefusal, storedFaceKey, type FontStore, type StoredFace } from './font-store'
import { previewFaceFamily } from './preview-face-family'
import { openPreviewFaceRegistry, type PreviewFaceBytes, type PreviewFaceRegistry, type PreviewFaceStatus } from './preview-face-registry'
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
import { embeddedFaceFamily, isCarriedFaceAssetKey } from './embedded-face-family'
import { isShippedFaceName, shippedFaceFamily } from './shipped-face-family'
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
// The same trick for the machine store's listing: an empty store and a store
// that could not be opened both render nothing, and neither should re-render
// the tree for the privilege.
const NO_STORED_FACES: ReadonlyArray<StoredFace> = []
type PreviewRecord = Readonly<{ bytes: ArrayBuffer; revision: number; identity: string; digest: string; diagnostics: ReadonlyArray<EngineDiagnostic>; token: number; generation: number }>
type PreviewFailureRecord = Readonly<{ error: EngineError; token: number; generation: number; revision: number }>

export default function App({ engine, fileAccess, sampleFileAccess, imageFileAccess, initialSnapshot, initialSampleData, blankBytes, initializationError, offlineState = 'unavailable', loadState, payload, engineState = 'waiting', onRetry = () => undefined }: AppProps = {}) {
  const [snapshot, setSnapshot] = useState(initialSnapshot)
  const [commitError, setCommitError] = useState<string>()
  const [propertyError, setPropertyError] = useState<PropertyCommitError>()
  const [fontChainError, setFontChainError] = useState<FontChainCommitError>()
  const [fontChainBusy, setFontChainBusy] = useState(false)
  // THE BUSY FLAG IS ALSO HELD IN A REF, because the window it now guards is a
  // NETWORK CHAIN. Since Story 16.1 a web-tier pick awaits up to six sequential
  // cross-origin round-trips before any command is sent, and a React state read
  // inside an event handler is the value that handler CLOSED OVER: two picks
  // dispatched before the re-render both see `false`, both resolve, and two
  // embeds commit. The ref is the same value read at the instant of the call.
  const fontChainBusyRef = useRef(false)
  const holdFontChain = (busy: boolean) => { fontChainBusyRef.current = busy; setFontChainBusy(busy) }
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
  // STORY 16.2 — THE MACHINE FONT STORE. What this machine holds, kept for the
  // family control's `AVAILABLE LOCALLY` group. A private window, cleared site
  // data or a quota refusal leaves a WORKING designer with an empty group —
  // Story 16.6 deleted the panel that used to say so on screen, deliberately
  // reversing 16.2's stated-degradation clause.
  const [storedFaces, setStoredFaces] = useState<ReadonlyArray<StoredFace>>(NO_STORED_FACES)
  // WHETHER THIS BROWSER CAN KEEP TYPEFACES AT ALL. Optimistic until the store
  // answers, because the modal cannot be open before it has: flashing the
  // degraded copy and then withdrawing it would be a worse lie than either state.
  const [storeKeepsFaces, setStoreKeepsFaces] = useState(true)
  // STORY 16.3 — THE FONT BROWSER IS OPEN OR IT IS NOT, AND THAT IS ALL THE
  // STATE IT HAS UP HERE. Everything the modal knows — the query, the chips,
  // the sort, the staged families — lives inside it and dies with it, which is
  // what keeps a Cancel/Apply pair from becoming a second document model
  // (AD-15). The one thing App owns is whether it is on screen.
  const [fontBrowserOpen, setFontBrowserOpen] = useState(false)
  const [machineFaces, setMachineFaces] = useState<ReadonlySet<string>>(NO_CARRIED_FACES)
  // THE HANDLE IS A PROMISE, NOT A RESOLVED VALUE, AND THAT IS A CORRECTNESS
  // POINT RATHER THAN A STYLE ONE. Opening a database is asynchronous, so a ref
  // holding the OPENED store is empty for the first moments of the session — and
  // a pick in that window would silently skip the store, keeping nothing, with
  // no failure anywhere to show it. Holding the OPENING lets every caller await
  // the same one open, whenever it asks.
  const fontStore = useRef<Promise<FontStore | undefined> | undefined>(undefined)
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
  //
  // THE KEY'S SHAPE IS ADMITTED HERE, NOT ASSUMED. `isCarriedFaceAssetKey` is
  // the derivation module's own predicate, and this is the one production
  // caller that hands it a key the FRAGMENT guard never saw: a fragment's
  // `assetKey` is admitted as 64 lowercase hex at the protocol boundary, a
  // chain ENTRY's is admitted on length alone. An entry whose key is not that
  // shape is a carried face the browser declines: no bytes are fetched, no
  // family is derived, nothing is registered, and every fragment keeps the
  // stylesheet's declared stack. It is a degrade, not a refusal — the
  // projection is already admitted and the session is untouched.
  const carriedFaceKeys = [...new Set((canvas?.fontChains ?? []).flatMap((chain) => chain.entries).map((entry) => entry.assetKey).filter(isCarriedFaceAssetKey))].sort()
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

  // STORY 16.2 — THE MACHINE STORE IS OPENED ONCE, FOR THE SESSION.
  //
  // THE LIFETIME IS THE THIRD ONE IN THIS FILE AND IT IS ARGUED, NOT ASSUMED,
  // because the effect above is the precedent for exactly that obligation.
  // `registerCarriedFaces` is DOCUMENT-scoped: it is re-run when the document
  // is replaced and when the set of carried entries changes, and its release
  // removes what it added. This one is MACHINE-scoped — strictly, the browser
  // profile's origin — and is deliberately NOT keyed on
  // `documentGenerationValue`: the whole property of the store is that it
  // survives the document, so re-opening it per document would be re-asking a
  // question whose answer cannot have changed, and clearing it per document
  // would delete the feature.
  //
  // AN OPEN FAILURE SAYS NOTHING ON SCREEN (Story 16.6, reversing 16.2's
  // stated-degradation clause by owner decision). A browser with storage
  // blocked simply keeps `storedFaces` empty; the designer works and picks
  // still embed straight into the document.
  useEffect(() => {
    let live = true
    const opening = openFontStore().then((opened) => {
      if (live) setStoreKeepsFaces(opened.ok)
      if (opened.ok) return opened.value
      return undefined
    })
    fontStore.current = opening
    void (async () => {
      const store = await opening
      if (!live || !store) return
      const listed = await store.list()
      if (!live || !listed.ok) return
      setStoredFaces(listed.value)
    })()
    return () => { live = false }
  }, [])

  // AND THE FACES THIS MACHINE HOLDS ARE REGISTERED FOR PREVIEW, ALONGSIDE THE
  // DOCUMENT'S OWN, WITHOUT DISTURBING THE EFFECT ABOVE.
  //
  // Two separate registrations rather than one merged list, and the separation
  // is the point: the document effect's key is `documentGenerationValue` plus
  // the carried listing, and folding a machine-scoped input into it would make
  // a store write re-register every carried face in the open document. They
  // meet only at the canvas, in the union below.
  //
  // THE FAMILY NAME IS THE SAME ON BOTH SIDES BECAUSE THE KEY IS. A stored face
  // and a carried face of the same bytes share a content address, so they share
  // the family `embedded-face-family.ts` derives, and `document.fonts` — a
  // global, name-keyed registry — simply holds one name over two identical
  // faces. That is a duplicate, not a conflict: either release removes only the
  // `FontFace` object it added, and the surviving one draws the same glyphs.
  //
  // THE COST, STATED: this reads every stored face's bytes once per session.
  // The store grows by one face per pick of a new family, so that is a handful
  // of megabytes for an author who has picked a handful of families, and it
  // buys a preview that does not wait for a network the store exists to avoid.
  const machineFaceListing = storedFaces.map((face) => face.key).sort().join(' ')
  useEffect(() => {
    setMachineFaces(NO_CARRIED_FACES)
    if (machineFaceListing === '') return
    return registerCarriedFaces(machineFaceListing.split(' '), async (key) => {
      const read = await (await fontStore.current)?.get(key)
      return read?.ok ? read.value?.bytes : undefined
    }, setMachineFaces)
  }, [machineFaceListing])

  // The canvas asks one question — "is there a face registered under this asset
  // key" — and both registrations can answer it.
  const paintableFaces = useMemo(() => machineFaces.size === 0 ? carriedFaces : new Set([...carriedFaces, ...machineFaces]), [carriedFaces, machineFaces])

  // The listing is re-read from the store rather than patched in memory, so the
  // store stays the single authority on what this machine holds. A refresh that
  // fails leaves the previous listing standing, silently — a stale listing
  // still picks correctly, because every pick re-reads the bytes.
  const refreshStoredFaces = async () => {
    const store = await fontStore.current
    if (!store) return
    const listed = await store.list()
    if (listed.ok) setStoredFaces(listed.value)
  }

  // STORY 16.3 — WHAT THE FONT BROWSER OFFERS, AND WHAT ITS SPECIMENS ARE SET IN.
  //
  // THE OFFERED LIST IS `offeredFamilies` WITH AN EMPTY QUERY, WHICH IS THE SAME
  // FUNCTION THE FAMILY CONTROL ASKS. There is one answer to "which families may
  // this author add", and a browser that computed its own would be a second one
  // — with its own opinion about variable-only families and its own tier order.
  // The browser filters and sorts what it is given; it never decides what it is
  // given.
  const browsableFamilies = useMemo(() => offeredFamilies('', storedFaces), [storedFaces])

  // AND THE BYTES A SPECIMEN IS SET IN COME FROM THE SAME THREE TIERS A PICK
  // RESOLVES FROM, THROUGH THE SAME THREE READS.
  //
  // NOTHING BUT BYTES TRAVELS. No licence record is kept, nothing is written to
  // the machine store and no command is sent: a preview is a face for one
  // `<span>` in one modal, and `preview-face-registry.ts` releases it when the
  // row leaves the page.
  //
  // THE STORE IS READ AND NOT WRITTEN, AND THAT ASYMMETRY IS THE POINT. THE
  // STORE HOLDS FACES THE AUTHOR CHOSE, NOT FACES THEY SCROLLED PAST. Reading it
  // is free and makes a family this machine already holds cost no network at
  // all; writing it would fill a store that carries a slot and byte budget
  // (Story 16.2) with every family that happened to cross the viewport, so the
  // budget would be spent by browsing rather than by deciding. A face earns its
  // place on this machine by being picked.
  //
  // THE WEB TIER GOES THROUGH `fetchWebFamily` — THE FULL RESOLUTION, LICENCE
  // CLASSIFICATION AND ALL — AND THAT IS NOT AN OVERSIGHT. A cheaper
  // bytes-only fetch would be a SECOND fetch path in a designer whose whole
  // host discipline rests on there being one, and it would let the browser set
  // a specimen in a face the pick would then refuse on its terms: `+ Add`
  // promising something the product declines. Reusing the pick's own resolution
  // means a specimen appears exactly for the families that can actually be
  // added. The cost is the pick's cost, bounded by `familiesPerPage`.
  const browserSpecimenBytes = async (family: string): Promise<ArrayBuffer | undefined> => {
    const source = browsableFamilies.find((entry) => entry.family === family)
    if (source === undefined) return undefined
    if (source.tier === 'local') {
      try {
        const response = await fetch(source.face.url)
        return response.ok ? await response.arrayBuffer() : undefined
      } catch {
        return undefined
      }
    }
    if (source.tier === 'stored') {
      const read = await (await fontStore.current)?.get(source.record.key)
      return read?.ok ? read.value?.bytes : undefined
    }
    const outcome = await fetchWebFamily(source.family)
    return outcome.ok ? outcome.face.bytes : undefined
  }

  // THE FAMILY CONTROL'S OWN READER (Story 16.7) — deliberately a SEPARATE
  // function from `browserSpecimenBytes` above, so that this story's own
  // divergence never has to edit the browser's (Ask First on that one is a
  // boundary of this story, not an invitation to fork a shared branch inside
  // it).
  //
  // LOCAL AND STORED READ EXACTLY AS THE BROWSER'S DO — no network for a face
  // this machine already holds, a store read for one it fetched before. `web`
  // IS WHERE THIS READER DIVERGES, ON PURPOSE: fetching to draw a specimen for
  // a family not on this machine is the thing Design Note (1) refuses — a
  // pick already blocks up to 30s on a stall and 180s against a slow host,
  // and a MENU must never cost that. So a `web` row resolves to `undefined`
  // with NO CALL TO `fetchWebFamily` AT ALL. STORY 16.9 removed the
  // dropdown's own web-tier group entirely, so this branch is now reachable
  // only through `familyControlSpecimenBytes`'s general contract (any
  // `family` string), never through a row this control renders.
  const familyControlSpecimenBytes = async (family: string): Promise<ArrayBuffer | undefined> => {
    const source = browsableFamilies.find((entry) => entry.family === family)
    if (source === undefined || source.tier === 'web') return undefined
    if (source.tier === 'local') {
      try {
        const response = await fetch(source.face.url)
        return response.ok ? await response.arrayBuffer() : undefined
      } catch {
        return undefined
      }
    }
    const read = await (await fontStore.current)?.get(source.record.key)
    return read?.ok ? read.value?.bytes : undefined
  }

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
    // Clears propertyError's sibling too: a refusal is about the edit that
    // caused it, and leaving a stale chain/embed refusal standing in the
    // TYPOGRAPHY section while the author goes on committing font size or
    // bold shows them a sentence about an edit they have moved past.
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
  // A FONT COMMAND TRAVELS THE SAME PATH A PROPERTY COMMIT TAKES — one opaque
  // command, one revision, one undo entry — and differs only in where a
  // refusal is anchored: by the CONTROL the panel dispatched from, never by
  // what the engine returned. That is what lets an unlocated ENGINE_REJECTED
  // (componentFields arity, a projection bound) be shown in exactly the same
  // place a located COMPONENT_INVALID is, with one rule. The message is the
  // engine's own string and is stored unprefixed, because the control already
  // says where it belongs.
  //
  // STORY 16.9 DELETED THIS FUNCTION'S OTHER TWO CALLERS, `applyFontChain` and
  // `pickCatalogueFamily`, WITH THE CONTROLS THEY SERVED — the chain editor
  // and the dropdown's install-tier pick. `engine-protocol.ts` still names the
  // six chain commands the removed editor issued, untouched, because this was
  // a UI removal and not a protocol change (see its own note). The one caller
  // left below is `dispatchEmbed`.
  //
  // STORY 16.3 GAVE IT A RETURN VALUE AND A SECOND ANNOUNCER, AND NEITHER
  // CHANGES AN EXISTING CALLER. The font browser dispatches one command per
  // staged family and has to name a refusal AGAINST THE FAMILY THAT EARNED IT —
  // three refusals cannot all be the panel's one `fontChainError`, and the two
  // that were overwritten would simply vanish. So the message is RETURNED, and
  // `announce` says whether this call also writes it to the panel.
  const sendFontChain = async (payload: ArrayBuffer, control: FontChainControl, responseGeneration: number, selectionKey: string, announce: 'panel' | 'caller' = 'panel'): Promise<string | undefined> => {
    if (!engine) return undefined
    setCommitError(undefined)
    if (announce === 'panel') setFontChainError(undefined)
    try {
      const priorRevision = snapshotRef.current?.revision
      const result = await engine.request('command', payload)
      if (result.snapshot.revision !== priorRevision) invalidatePreview()
      setHistoryAvailability(result.snapshot)
      if ((snapshotRef.current?.revision ?? -1) < result.snapshot.revision) setCurrentSnapshot(result.snapshot)
    } catch (error) {
      const message = componentDiagnosticDetail(error).message
      if (announce === 'panel' && documentGeneration.current === responseGeneration && selectedRef.current.join(',') === selectionKey) setFontChainError({ control, selectionKey, message })
      return message
    }
    return undefined
  }

  // STORY 8.6, WIDENED BY STORY 16.1: PICKING A FAMILY, FROM EITHER OF TWO TIERS.
  //
  // THE FUNCTION GAINED A SOURCE; IT DID NOT SWAP ONE (D-16.R.3). The local
  // tier is the 21 committed faces, read from the release bundle's own
  // content-addressed assets exactly as before — the same read `runtimeAssetUrls`
  // assets get, behind the service worker, so it works with the browser offline
  // and NO THIRD PARTY IS CONTACTED AT ALL. The web tier is a family from the
  // build-time index snapshot, whose metadata, licence text and bytes are
  // fetched at the moment of the pick.
  //
  // LOCAL WINS AND IS NEVER RE-FETCHED. `offeredFamilies` has already removed
  // the web duplicate of a family the local tier holds, so a `Roboto` or an
  // `Inter` pick never reaches `fetchWebFamily`, and the snapshot's `axes`
  // opinion about them is never consulted. The committed bytes carry a stronger
  // record than any fetch can produce; preferring a fetch would replace a
  // verified record with an unverified one.
  //
  // CLASSIFY, THEN EMBED. For the web tier every licence decision is made inside
  // `fetchWebFamily` BEFORE a byte is fetched, and a refusal comes back as a
  // sentence rather than as a face. No licence is knowable before a pick — the
  // index publishes no licence field — so this refusal is necessarily post-pick,
  // and it is surfaced at the control the author acted on.
  //
  // THE COMMAND IS UNTOUCHED, AT TWELVE FIELDS. `embedFontFamilyCommand` already
  // demands exactly what a `.folio` requires and Go refuses the pick without it,
  // so changing the SOURCE of those values while leaving that guard in place is
  // what keeps this story from reaching the format.
  //
  // THE PROPOSED TAIL IS COMPUTED HERE and is a PROPOSAL, not a rule: the
  // shipped faces for the scripts the picked face does not cover, in the order
  // scriptFallbackFaces names them. The document's chain data still carries it
  // exactly as ANY other chain; this designer ships no UI to hand-edit that
  // order (Story 16.9 removed the chain editor), so the proposal stands until
  // a future UI, or the engine's own commands, revise it (AC3).
  // THE BUSY FLAG IS HELD ACROSS THE WHOLE RESOLUTION, NOT ONLY THE COMMAND.
  // Before Story 16.1 the awaited work in front of this hold was ONE
  // same-origin read of a precached bundle asset, so the window in which a
  // second pick could pass this guard was negligible. It is now a chain of up to
  // six sequential cross-origin round-trips — up to four `METADATA.pb` probes,
  // then the licence file, then the face bytes — and a second pick during that
  // window would resolve concurrently and commit a second embed. So the flag is
  // taken HERE and released in the `finally`, and the command half is called
  // through `sendFontChain`, which does not take it again.
  //
  // STORY 16.3 BUILT THIS SEAM AND STORY 16.5 SWAPPED ITS BODY, WHICH IS EXACTLY
  // WHAT IT WAS BUILT FOR (D-16.R.46).
  //
  // IT USED TO EMBED. The bytes travelled into the document at the moment of a
  // pick, so a family the author merely tried landed in the `.folio` and stayed
  // there. IT NOW INSTALLS: fetch, classify, keep on this machine, AND SEND NO
  // ENGINE COMMAND AT ALL. The embed moved to first use — see
  // `embedInstalledFamily` below — so an installed-but-unused face is never in
  // the file to be pruned.
  //
  // BOTH CALLERS ARE UNCHANGED BY THAT SWAP, which is the property the seam was
  // built to have: the family control's pick and the font browser's confirm name
  // no mechanism, so neither had to learn a new one.
  //
  // THE NAME IS NOW WRONG AND IT IS KEPT ANYWAY, deliberately: it is the seam
  // D-16.R.46 Q2 named, two specs point at it by that name, and renaming it in
  // the same commit that inverts it would make the diff unreadable at exactly
  // the moment the record needs to be legible. It reads `installFamily` below,
  // which is what it does.
  //
  // NO REVISION, NO HISTORY ENTRY, NO UNDO, BY CONSTRUCTION rather than by
  // suppression. History is whole canonical `.folio` byte snapshots and `Apply`
  // short-circuits when the bytes do not move (`folio-go/wasm/engine.go`), so an
  // action that sends no command cannot move any of the three.
  const addFamilyToDocument = async (source: FamilySource, responseGeneration: number, selectionKey: string, announce: 'panel' | 'caller' = 'panel'): Promise<string | undefined> => {
    if (!engine) return 'This designer has no engine to send the change to.'
    if (fileBusy || fontChainBusyRef.current) return `${source.family} was not installed: the designer was busy with another change. Try it again.`
    holdFontChain(true)
    try {
      return await installFamily(source, responseGeneration, selectionKey, announce)
    } finally {
      if (documentGeneration.current === responseGeneration) holdFontChain(false)
    }
  }

  /**
   * THE REFUSAL SURFACE, HOISTED SO BOTH HALVES OF THE SPLIT SAY IT THE SAME WAY.
   *
   * A refusal that resolves after the selection or the document moved on is NOT
   * shown in the panel — `sendFontChain`'s own rule, applied to the half of the
   * flow that happens before it is called. It is still RETURNED, because the
   * browser's caller is a modal that is still open and still owns a row for this
   * family: "the panel is looking at something else now" is a reason not to
   * paint an error on the panel, never a reason to swallow one.
   */
  const refuseFontChain = (message: string, responseGeneration: number, selectionKey: string, announce: 'panel' | 'caller'): string => {
    if (announce === 'panel' && documentGeneration.current === responseGeneration && selectedRef.current.join(',') === selectionKey) {
      setFontChainError({ control: { action: 'embed' }, selectionKey, message })
    }
    return message
  }

  /** A face resolved from any tier, carrying everything `embedFontFamily` refuses a document without. */
  type ResolvedFace = Readonly<{ family: string; style: string; licence: string; licenceText: string; copyright: string; source: string; mediaType: string; bytes: ArrayBuffer; scripts: ReadonlyArray<string> }>

  /**
   * THE ONE EMBED DISPATCH, so the three paths that can reach the command cannot
   * drift in what they send.
   *
   * THE PROPOSED TAIL IS COMPUTED HERE and is a PROPOSAL, not a rule: the shipped
   * faces for the scripts the picked face does not cover, in the order
   * `scriptFallbackFaces` names them. It lands in the document's chain data
   * like any other chain; this designer ships no UI to hand-edit that order.
   */
  const dispatchEmbed = async (face: ResolvedFace, responseGeneration: number, selectionKey: string, announce: 'panel' | 'caller'): Promise<string | undefined> => {
    const tail = scriptFallbackFaces.filter(([script]) => !face.scripts.includes(script)).map(([, shipped]) => shipped)
    return sendFontChain(embedFontFamilyCommand({ chain: face.family, family: face.family, style: face.style, licence: face.licence, licenceText: face.licenceText, copyright: face.copyright, source: face.source, mediaType: face.mediaType, bytes: face.bytes, tail }), { action: 'embed' }, responseGeneration, selectionKey, announce)
  }

  /**
   * INSTALLING A FACE IS NOT EMBEDDING IT (Story 16.5, D-16.R.46).
   *
   * FETCH, CLASSIFY, STORE. NO ENGINE COMMAND. Three of the four steps the old
   * pick took are here unchanged, and the fourth — the embed — is gone from this
   * path entirely. The document's revision, history and asset map are untouched,
   * which is what makes "install what looks promising" cost the author's file
   * nothing.
   *
   * INSTALL RUNS EVERY ADMISSION CHECK THAT CAN RUN AT INSTALL. Moving the embed
   * to first use moves every refusal the command makes to a later moment than
   * the one the author acted in, and a refusal that arrives at first use is a
   * worse refusal. So `fetchWebFamily` does the whole admission it always did —
   * the closed licence-token table, the upstream licence text, nameID 0 from the
   * bytes — plus Story 16.5's refuse-only `fvar` filter over the fetched bytes.
   * Nothing here ADMITS anything: `embedFontFamily` still decides what enters a
   * document.
   *
   * ONE CHECK CANNOT MOVE AND IT IS NAMED RATHER THAN HIDDEN. Go's nameID-13
   * licence-signature tie (`internal/fontset/licencesignature.go`) compares a
   * face's declared licence against the licence written inside the face, and
   * porting it would need a second name-table reader and two regex tables in
   * this designer — a competing authority over what enters a document, which is
   * the one thing this story may not build. So a face CAN install successfully
   * and still be refused the first time it is used, and `embedInstalledFamily`
   * below says exactly that when it happens.
   *
   * ONLY A `web` ROW HAS ANYTHING TO INSTALL. The local tier ships inside the
   * release and the stored tier is already here, so both are routed to
   * `embedInstalledFamily` by the family control's fork and are not stageable in
   * the browser. Reaching this function with one of them is a routing defect, and
   * it is NAMED rather than silently answered `undefined`, which would report a
   * no-op as a success.
   */
  const installFamily = async (source: FamilySource, responseGeneration: number, selectionKey: string, announce: 'panel' | 'caller' = 'panel'): Promise<string | undefined> => {
    const refuse = (message: string) => refuseFontChain(message, responseGeneration, selectionKey, announce)
    // Spelled as the discriminant rather than through `familyIsInstalled` so the
    // narrowing below is the compiler's and not a comment's.
    if (source.tier !== 'web') return refuse(`${source.family} is already on this machine, so there is nothing to install. Pick it again to use it in this document.`)
    const outcome = await fetchWebFamily(source.family)
    if (!outcome.ok) return refuse(outcome.reason)
    // LAYOUT DIVERGENCE IS AN OBSERVATION, AND AN OBSERVATION NEEDS A READER.
    // `fetchWebFamily` records when the directory a family was resolved in
    // disagrees with the licence its own metadata declares — never a refusal
    // (D-16.R.6: METADATA.pb wins, the directory is only where the files sit),
    // but worth seeing, because systematically it means the probe order is
    // costing round-trips. It is written to the browser's own log, which is
    // where a person already looks for what the designer did on a pick; it is
    // deliberately NOT a UI surface, because nothing here is wrong and an
    // author has no decision to make about it.
    if (outcome.face.layoutDivergence !== undefined) console.info(outcome.face.layoutDivergence)
    const face: ResolvedFace = { ...outcome.face, scripts: source.row.scripts }

    // A STORE THAT CANNOT BE OPENED DEGRADES TO THE PRE-16.5 MODEL RATHER THAN
    // REFUSING (orchestrator ruling, 2026-09-03). NEITHER OBVIOUS ANSWER WAS RIGHT.
    //
    // Refusing contradicts Story 16.2's locked contract — *"given storage that
    // cannot be opened or written, the designer still works and says what is
    // degraded"* — and would mean a private window could add no font at all.
    // But 16.2's degradation taken LITERALLY fails too: under the old model the
    // store was a convenience because the pick embedded immediately, while under
    // embed-on-use it is load-bearing, so "install anyway, keep nothing" stores
    // nothing and nothing can ever be used. That is the dead end D-16.R.46 Q4
    // forbids.
    //
    // SO THIS BROWSER GETS THE OLD MODEL: the pick puts the font straight into
    // the document, exactly as it did before this story. It is also the honest
    // description of the mode — with nowhere to keep a face there is nothing to
    // install, and embedding at once is the only way a web font can be used
    // here. The machine store stays a convenience and never becomes a
    // dependency. Story 16.6 deleted the note that used to say this on screen
    // (deliberate reversal, D-16.R.82) — nothing is said, here or anywhere.
    //
    // The property is NOT committed, because this is still the second arm of the
    // fork: Story 8.6's *"carry this typeface"* and *"draw this box with it"* stay
    // two decisions, and the degradation may not quietly fuse them.
    if (!(await fontStore.current)) {
      const rejected = await dispatchEmbed(face, responseGeneration, selectionKey, announce)
      if (rejected !== undefined) return rejected
      return undefined
    }

    // OTHERWISE THE STORE WRITE IS THE WHOLE ACT, AND THAT INVERTS 16.2's RULING.
    //
    // 16.2 put this write AFTER the embed and ruled a quota refusal a
    // DEGRADATION rather than a failed pick, on the ground that *"the face was
    // fetched, the terms were admitted and the document has it"* — there was
    // something left to degrade FROM. Here there is not. No command follows this
    // line, no document has the face, and if the write is refused the author has
    // nothing at all. So it is a REFUSAL, stated at the control they acted on.
    //
    // The ordering question 16.2 answered has dissolved rather than flipped:
    // with one act there is no second act to order it against.
    const kept = await keepOnThisMachine(face)
    if (kept !== undefined) return refuse(storeWriteRefusal(face.family, kept))
    return undefined
  }

  /**
   * FIRST USE — THE MOMENT A FONT STARTS TRAVELLING INSIDE THE TEMPLATE.
   *
   * Story 8.6's guarantee is untouched: send the file to a colleague and the
   * pages come out identical, because a font still travels inside the `.folio`
   * (CAP-2/AD-8). Only the moment it starts travelling moves — from the pick to
   * the first time something in the template is actually set in the family.
   *
   * TWO COMMANDS, NEVER ONE, AND THE ORDER IS FORCED BY THE ENGINE.
   * `canvas.fontFamilies` is the closed set `style.fontFamily` may name, so
   * `updateComponentProperties` is refused unless the chain is already declared.
   * This half sends the embed; the family control commits the property after it
   * returns, and only if it returns nothing. Two commands means two unambiguous
   * undo entries — "carry this typeface" and "draw this box with it" are two
   * decisions and fusing them would make one undo ambiguous. There is no
   * compound-command mechanism in this product and Story 8.6's refused fusion is
   * not reopened.
   *
   * THE BYTES COME FROM THIS MACHINE FIRST, AND A MISS FALLS THROUGH TO THE
   * FETCH RATHER THAN REFUSING (Story 16.2's contract, restored by orchestrator
   * ruling 2026-09-03 after this story briefly removed it).
   *
   * A `local` row is read from the release's own content-addressed assets behind
   * the service worker; a `stored` row is read out of the machine font store. A
   * stored read can MISS: the entry may have been dropped as unsound between the
   * listing and this read — the store self-heals by dropping, silently — or
   * removed in another tab. 16.2's matrix is explicit that this is SELF-HEALING,
   * *"entry treated as absent and dropped; refetch on next pick"*, and that is a
   * shipped contract this story may not amend. Refusing instead would convert a
   * path that repairs itself into a permanent local failure the author would
   * have no way to clear, since Story 16.6 removed the only control that ever
   * touched the store.
   *
   * THE FACE IS THEN WRITTEN BACK, which is the half that makes it self-healing
   * rather than merely survivable. That write follows an embed, so the document
   * already has the face; a write-back failure is silent (Story 16.6) rather
   * than stated, same as every other store degradation.
   */
  const embedInstalledFamily = async (source: FamilySource, responseGeneration: number, selectionKey: string): Promise<string | undefined> => {
    const refuse = (message: string) => refuseFontChain(message, responseGeneration, selectionKey, 'panel')
    // EVERY EXIT GOES THROUGH THE REFUSAL SURFACE, INCLUDING THE EARLY ONES.
    // These two returned a sentence that nothing painted and nothing read: the
    // family control discards the value, so clicking an installed family while
    // the designer was busy produced no command, no property commit and no
    // message anywhere. A guard the author cannot see is a guard that looks
    // like a broken control.
    if (!engine) return refuse('This designer has no engine to send the change to.')
    if (fileBusy || fontChainBusyRef.current) return refuse(`${source.family} was not used: the designer was busy with another change. Try it again.`)
    holdFontChain(true)
    try {
      let embedded: ResolvedFace
      // Set when the bytes had to come over the network because the stored entry
      // was gone. They are worth writing back; a face read successfully OUT of
      // the store is deliberately not rewritten to it, because the record is
      // already there byte for byte.
      let refetched = false
      if (source.tier === 'stored') {
        const read = await (await fontStore.current)?.get(source.record.key)
        if (read?.ok && read.value !== undefined) {
          const held = read.value
          embedded = { family: held.family, style: held.style, licence: held.licence, licenceText: held.licenceText, copyright: held.copyright, source: held.source, mediaType: held.mediaType, bytes: held.bytes, scripts: held.scripts }
        } else {
          const outcome = await fetchWebFamily(source.family)
          if (!outcome.ok) return refuse(outcome.reason)
          if (outcome.face.layoutDivergence !== undefined) console.info(outcome.face.layoutDivergence)
          embedded = { ...outcome.face, scripts: source.record.scripts }
          refetched = true
        }
      } else if (source.tier === 'local') {
        const face = source.face
        let bytes: ArrayBuffer
        try {
          const response = await fetch(face.url)
          if (!response.ok) throw new Error(`the bundled face responded ${response.status}`)
          bytes = await response.arrayBuffer()
        } catch (error) {
          return refuse(`${face.family} could not be read from the offline bundle: ${error instanceof Error ? error.message : String(error)}`)
        }
        embedded = { family: face.family, style: face.style, licence: face.licence, licenceText: face.licenceText, copyright: face.copyright, source: face.source, mediaType: 'font/ttf', bytes, scripts: face.scripts }
      } else {
        // UNREACHABLE BY EITHER CALLER, AND NAMED RATHER THAN SKIPPED. The family
        // control routes a `web` row to `installFamily`, so arriving here with one
        // is a routing defect; answering it by fetching would restore the fused
        // pick this story exists to split.
        return refuse(`${source.family} is not on this machine yet, so it cannot be put into this document. Install it first.`)
      }
      // THE ANNOUNCER IS `'caller'` SO THE PANEL IS PAINTED ONCE, WITH THE WHOLE
      // SENTENCE. `sendFontChain` would otherwise write the engine's bare
      // refusal and the disclosure below would have to overwrite it — two
      // messages for one event, in whichever order React settled.
      setFontChainError(undefined)
      const rejected = await dispatchEmbed(embedded, responseGeneration, selectionKey, 'caller')
      if (rejected !== undefined) return refuse(lateEmbedRefusal(embedded.family, rejected))
      // THE STORE HEALS AFTER THE EMBED. The document already has the face, so
      // a write-back failure here is silent (Story 16.6) rather than stated —
      // there is nothing left for the author to act on.
      if (refetched) await keepOnThisMachine(embedded)
      // THE PROPERTY COMMIT IS AUTHORISED HERE, NOT IN THE FAMILY CONTROL, AND
      // ONLY IF THE DOCUMENT AND THE SELECTION ARE STILL THE ONES THE AUTHOR
      // ACTED IN.
      //
      // The embed is a chain of awaits — a store read or a whole refetch — and
      // an Open, a Start blank or an undo can land in the middle of it. The
      // family control cannot see that: its `documentGeneration` and its `ids`
      // are the render's, frozen in the closure, so comparing them to
      // themselves would always agree. The live values are `documentGeneration`
      // and `selectedRef` HERE, which is why the check belongs here.
      //
      // AND IT MATTERS BECAUSE `applyProperties` SENDS FIRST AND GUARDS AFTER:
      // it dispatches the command unconditionally and only declines to install
      // the RESULT against a moved document. So a stale commit is not a dropped
      // response — it is `updateComponentProperties` reaching the engine with
      // the previous document's element ids.
      //
      // A NON-`undefined` RETURN SUPPRESSES THE COMMIT, and `refuse` declines to
      // paint anything whose generation has moved, so this is silent by
      // construction — which is right: the author is looking at something else.
      if (documentGeneration.current !== responseGeneration || selectedRef.current.join(',') !== selectionKey) {
        return refuse(`${embedded.family} was embedded, but no component was set in it: the document or the selection moved while the face was being written.`)
      }
      return undefined
    } finally {
      if (documentGeneration.current === responseGeneration) holdFontChain(false)
    }
  }

  /**
   * KEEPING THE FACE, AND FAILING TO KEEP IT IS NOW A REFUSAL RATHER THAN A
   * DEGRADATION (Story 16.5 — see `installFamily` for why the ruling inverted).
   *
   * IT RETURNS THE REASON AND RENDERS NOTHING, because its two callers owe the
   * author two different treatments. `installFamily` turns a non-`undefined`
   * reason into a refusal via `storeWriteRefusal`, because the write IS the
   * whole act there. `embedInstalledFamily`'s write-back after a refetch is the
   * one place under embed-on-use where the document already has the face, so it
   * discards the reason and stays silent (Story 16.6) rather than stating it.
   * `undefined` means the face is on this machine.
   *
   * Everything `embedFontFamily` requires travels into the store WITH the bytes
   * — licence identifier, licence text and copyright — because a face offered
   * from the store must be embeddable without a network, and the command
   * refuses without all three. A store that kept the bytes and dropped the
   * terms would put a document its own parser refuses one step away.
   *
   * The key is computed HERE, from the bytes, by `crypto.subtle`. It is the
   * store's own address and it agrees with the one Go derives for the same
   * bytes — `src/font-store.test.ts` asserts that agreement against a digest Go
   * itself produced, so the two addressings cannot drift.
   */
  const keepOnThisMachine = async (face: ResolvedFace): Promise<string | undefined> => {
    const store = await fontStore.current
    if (!store) return 'this browser is not letting the designer keep typefaces on this machine'
    let key: string
    try {
      key = await storedFaceKey(face.bytes)
    } catch (error) {
      return componentDiagnostic(error)
    }
    const written = await store.put({ ...face, key, byteLength: face.bytes.byteLength, fetchedAt: new Date().toISOString().slice(0, 10), bytes: face.bytes })
    if (!written.ok) return written.reason
    await refreshStoredFaces()
    return undefined
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
  const setCurrentSnapshot = (next: EngineSnapshot | undefined, keepNewerDraft = false, clearDocumentInteraction = false) => { snapshotRef.current = next; setSnapshot(next); setHistoryAvailability(next); if (clearDocumentInteraction) { documentGeneration.current++; tableEditorSession.current++; setDocumentGenerationValue(documentGeneration.current); setSelected([]); setBindingError(undefined); setBindingBusy(false); setTableEditor(undefined); setTableEditorError(undefined); setFontBrowserOpen(false); setAssetError(undefined); setAssetBusy(false); setFontChainError(undefined); holdFontChain(false); clearInteraction() }; if (next?.canvas) { setPreset(next.canvas.preset); setOrientation(next.canvas.orientation); if (!keepNewerDraft) setDraft(draftFor(next.canvas)) } }
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
          ? <CanvasComponent key={occurrence.component.id} component={occurrence.component} carriedFaces={paintableFaces} origin={origin} note={content && sheet.index > 0 ? canvasColumnPositionNotice(sheet.index + 1, sheets) : undefined} limit={{ band: band.name, width: band.width, height: band.height }} zoom={zoom} selected={selected.includes(occurrence.component.id)} preview={drag?.id === occurrence.component.id ? drag : undefined} engine={engine} generation={documentGenerationValue} trackColumn={content && many ? (edge: number, delta: number) => columnEdgeAfterDrag(model, projection, zoom, edge, delta) : undefined} onSelect={select} onDelete={deleteSelection} onDragStart={setDrag} onDragEnd={(finished) => { if (!finished.changed) { setDrag(undefined); return } const command = finished.mode === 'move' ? moveComponentCommand(occurrence.component.id, finished.x, finished.y, snapEnabled) : setComponentBoundsCommand(occurrence.component.id, finished.x, finished.y, finished.width, finished.height, snapEnabled); void commitComponent(command, () => setDrag(undefined)).finally(() => setDrag(undefined)) }} />
          : <ComponentEcho key={`${occurrence.component.id}@${sheet.index}`} component={occurrence.component} carriedFaces={paintableFaces} y={occurrence.y} zoom={zoom} engine={engine} generation={documentGenerationValue} />
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
      {/* NO onClick HERE, DELIBERATELY (Story 17.2). The backdrop — the grey
          space around the page — used to clear the selection when the click
          landed on the <main> itself. That guard (target === currentTarget)
          was written to spare the toolbar and the sheet stack, and so it fired
          on precisely the region where a click is most often a miss. Clicking
          the page surface (the `page-surface` onClick) and Escape (below, on
          this element, and NOT gated on the target test) remain the two ways
          to deselect. The revokeTableEditor call went with the clear: the
          editor is bound to the one selected component, and that selection now
          survives the click, so there is nothing left here for it to unbind.
          (Its own modal is `position: fixed; inset: 0; z-index: 20` at
          App.css:304, so in a real browser a click cannot reach this element
          while the editor is open — that is a reading of the stylesheet, which
          jsdom does not apply and no test measures.) */}
      {mode === 'design' ? <main ref={canvasRegionRef} className={`canvas-region${placing ? ' canvas-region-placing' : ''}`} aria-label="Canvas region" tabIndex={0} onPointerMove={(event) => { if (placing) setPlacingAt({ x: event.clientX, y: event.clientY }) }} onPointerLeave={() => setPlacingAt(undefined)} onKeyDown={(event) => { if ((event.key === 'Delete' || event.key === 'Backspace') && event.target === event.currentTarget && selected.length === 1) { event.preventDefault(); deleteSelection() } if (event.key === 'Escape') { clearInteraction(); revokeTableEditor(); setSelected([]) } }}>
        <div className="canvas-tools" aria-label="Canvas controls"><button type="button" onClick={() => setZoom((value) => Math.max(0.5, value - 0.1))} aria-label="Zoom out">−</button><output aria-label="Canvas zoom">{Math.round(zoom * 100)}%</output><button type="button" onClick={() => setZoom((value) => Math.min(2, value + 0.1))} aria-label="Zoom in">+</button><button type="button" onClick={() => setGridVisible((value) => !value)} aria-pressed={gridVisible}>Grid {gridVisible ? 'on' : 'off'}</button><button type="button" onClick={() => setSnapEnabled((value) => !value)} aria-pressed={snapEnabled}>Snap {snapEnabled ? 'on' : 'off'} <kbd aria-hidden="true">{shortcuts.snap}</kbd></button><button type="button" onClick={duplicateSelection} disabled={selected.length !== 1}>Duplicate <kbd aria-hidden="true">{shortcuts.duplicate}</kbd></button><button type="button" onClick={deleteSelection} disabled={selected.length !== 1}>Delete <kbd aria-hidden="true">{shortcuts.delete}</kbd></button><span>Nudge <kbd aria-hidden="true">{shortcuts.nudge}</kbd></span></div>
        {disclosure && <p className="canvas-disclosure" role="status" aria-live="polite" aria-label="Canvas sheet disclosure">{disclosure}</p>}
        {canvas && stack ? (stack.sheets.length === 1 ? sheetSurface(canvas, stack, stack.sheets[0] as Sheet) : <div className="sheet-stack" style={{ '--sheet-stack-gap': `${SHEET_STACK_GAP}px` } as CSSProperties}>{stack.sheets.map((sheet) => sheetSurface(canvas, stack, sheet))}</div>) : <p className="canvas-awaiting" role="status">Waiting for Go page geometry.</p>}
        {placing && placingAt && <span className="placement-ghost" aria-hidden="true" style={{ '--ghost-x': `${placingAt.x}px`, '--ghost-y': `${placingAt.y}px` } as CSSProperties}><PaletteIcon kind={placing} />{paletteItems.find(([, kind]) => kind === placing)?.[0]}</span>}
        {commitError && <p role="alert" className="file-message">{commitError}</p>}{fileError && <p role="alert" className="file-message">{fileError}</p>}{fileStatus && <p role="status" aria-live="polite" className="file-message">{fileStatus}</p>}{locateStatus && <p role="status" aria-live="polite" className="file-message">{locateStatus}</p>}
      </main> : <main className="preview-region" aria-label="Preview region"><div className="preview-heading"><p>{previewStatus === 'current' ? 'EXACT LOCAL PRODUCTION PDF' : 'LOCAL PDF PREVIEW'}</p><button type="button" className="file-button" onClick={returnToDesign}>{['checking', 'debouncing', 'rendering'].includes(previewStatus) ? 'Cancel and return to Design' : 'Return to Design'}</button></div><p id="preview-freshness-status" className="preview-status" role="status" aria-live="polite" aria-atomic="true">{!sampleData ? 'Preview unavailable: no sample data loaded' : previewStatus === 'current' ? 'Current exact local PDF' : previewStatus === 'stale' ? `${staleCopy(staleReason)}${currentFailure ? `; local PDF render failed: ${currentFailure.error.message}` : previewIssue ? `; ${previewIssue}` : ''}` : ['checking', 'debouncing', 'rendering'].includes(previewStatus) ? 'Rendering local PDF' : previewStatus === 'error' ? `Local Preview work failed${previewIssue ? `: ${previewIssue}` : currentFailure ? `: ${currentFailure.error.message}` : ''}` : 'Preview is waiting for local inputs'}</p>{currentFailure && <PreviewFailure error={currentFailure.error} onRetry={() => retryFromFailure(currentFailure)} onReturn={() => returnFromFailure(currentFailure)} />}{preview && <><PDFPreviewViewer bytes={preview.bytes} label={previewStatus === 'current' ? `Current exact local production PDF, revision ${preview.revision}` : `Stale historical PDF, revision ${preview.revision}`} describedBy="preview-freshness-status" state={previewViewState} onStateChange={changePreviewViewState} onError={(error) => viewerError(preview.token, error)} onPageCount={(pages) => viewerPages(preview.token, pages)} />{currentDiagnostics && <PreviewDiagnostics diagnostics={currentDiagnostics.diagnostics} dismissed={dismissedDiagnostics} onDismiss={(key) => setDismissedDiagnostics((current) => new Set([...current, key]))} onLocate={(location) => locateDiagnostic(currentDiagnostics, location)} />}</>}<p className="preview-evidence">{preview ? `Historical producer digest ${preview.digest}` : 'Go production digest pending'}{preview ? ` · ${preview.diagnostics.length} diagnostics retained` : ''}</p></main>}
      <aside className="inspector-panel" aria-label="Inspector">
        <div className="panel-tabs" role="tablist" aria-label="Inspector tabs">{inspectorTabs.map(([tab, designLabel, previewLabel]) => <button key={tab} type="button" role="tab" id={`inspector-tab-${tab}`} aria-controls={`inspector-panel-${tab}`} aria-selected={inspectorTab === tab} tabIndex={inspectorTab === tab ? 0 : -1} className={`panel-tab panel-tab-${tab}${inspectorTab === tab ? ' panel-tab-active' : ''}`} onClick={() => setInspectorTab(tab)} onKeyDown={(event) => { const next = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0; if (!next) return; event.preventDefault(); const order = inspectorTabs.map(([name]) => name); const target = order[(order.indexOf(tab) + next + order.length) % order.length]!; setInspectorTab(target); requestAnimationFrame(() => document.getElementById(`inspector-tab-${target}`)?.focus()) }}>{mode === 'preview' ? previewLabel : designLabel}</button>)}</div>
        <div className="panel-body" role="tabpanel" id="inspector-panel-properties" aria-label={mode === 'preview' ? 'Preview inputs' : 'Properties panel'} hidden={inspectorTab !== 'properties'}>{mode === 'preview' ? <><p className="section-label">PREVIEW INPUTS</p><ParameterEditor referenceState={parameterReferenceState} accepted={previewParams} draft={previewParamsDraft} error={previewParamsError} onDraft={acceptPreviewParameters} onNamedValue={setNamedParameter} /><button type="button" className="file-button" onClick={() => void renderPreview(true)} disabled={!sampleData}>Render local PDF</button><p className="honest-note">Parameters are local Preview input and are not part of the template.</p></> : selected.length > 0 && canvas ? <ComponentProperties key={`${documentGenerationValue}:${selected.join(',')}`} components={canvas.components.filter((component) => selected.includes(component.id))} fontFamilies={canvas.fontFamilies} fontChains={canvas.fontChains} carriedFaces={paintableFaces} specimenBytes={familyControlSpecimenBytes} defaultFontSize={canvas.defaultFontSize} defaultLineSpacing={canvas.defaultLineSpacing} onCommit={applyProperties} onUseFamily={(source) => embedInstalledFamily(source, documentGeneration.current, selected.join(','))} onOpenFontBrowser={() => setFontBrowserOpen(true)} browserOpen={fontBrowserOpen} storedFaces={storedFaces} fontChainError={fontChainError} fontChainBusy={fontChainBusy || fileBusy} documentGeneration={documentGenerationValue} propertyError={propertyError} drag={drag} onEditTable={(id) => void openTableEditor(id)} onPickImage={(id) => void applyImageAsset(id)} imageAvailable={imageFileAccess !== undefined} assetBusy={assetBusy} assetError={assetError} /> : <PageSetup preset={preset} orientation={orientation} draft={draft} onPreset={setPreset} onOrientation={setOrientation} onDraft={updateDraft} onApply={applyPageSetup} disabled={!canvas || fileBusy} />}</div>
        <div className="panel-body" role="tabpanel" id="inspector-panel-data" aria-labelledby="inspector-tab-data" hidden={inspectorTab !== 'data'}><DataPanel sample={sampleData} error={sampleError} busy={sampleBusy} available={Boolean(sampleFileAccess)} selectedComponentId={selected.length === 1 ? selected[0] : undefined} selectedBinding={selected.length === 1 ? canvas?.components.find((component) => component.id === selected[0])?.binding : undefined} bindingError={bindingError} bindingBusy={bindingBusy} onLoad={() => void loadSample()} onConnect={(segments) => void bindPickedPath(segments)} /></div>
      </aside>
    </div>
    {tableEditor && <TableEditor projection={tableEditor} busy={tableEditorBusy} error={tableEditorError} candidates={tableSampleCandidates(sampleData?.tree)} sampleAvailable={Boolean(sampleData)} onClose={closeTableEditor} onAdd={(index) => void commitTableColumn(addTableColumnCommand(tableEditor.table.tableId, index))} onRemove={(columnId) => void commitTableColumn(removeTableColumnCommand(tableEditor.table.tableId, columnId))} onMove={(columnId, index) => void commitTableColumn(moveTableColumnCommand(tableEditor.table.tableId, columnId, index))} onUpdate={(columnId, field, value) => void commitTableColumn(updateTableColumnCommand(tableEditor.table.tableId, columnId, field, value))} onConfigure={(collection, alias) => void commitTableColumn(configureTableBindingCommand(tableEditor.table.tableId, collection, alias))} onBind={(columnId, field) => void commitTableColumn(updateTableColumnBindingCommand(tableEditor.table.tableId, columnId, field))} onFooter={(columnId, footer, footerOf, footerFormat) => void commitTableColumn(updateTableColumnFooterCommand(tableEditor.table.tableId, columnId, footer, footerOf, footerFormat))} />}
    {fontBrowserOpen && canvas && <FontBrowser sources={browsableFamilies} inTemplate={canvas.fontFamilies} previewBytes={browserSpecimenBytes} onAddFamily={(source) => addFamilyToDocument(source, documentGeneration.current, selected.join(','), 'caller')} storeKeepsFaces={storeKeepsFaces} onClose={() => setFontBrowserOpen(false)} />}
    {/* THE FONT COUNT, AND NOTHING ELSE NEW (Story 16.4). It is read off
        `canvas.fontFamilies`, which is `IN THIS TEMPLATE`'s own predicate, so
        the dropdown's first group and this line teach one model from one
        source and cannot drift apart.

        THE MOCKUP'S BINDING IS REFUSED, DELIBERATELY. `statusFontLine` counts
        `s.added.length` — the fonts added THIS SESSION — which is the
        session-scoped set this story forbids as a grouping key and would be no
        better here; and its else-branch is a hardcoded "3 fonts in template",
        which is placeholder data rather than a specification. No grid reading,
        no snap state and no selection content is added: those are three more
        claims about the canvas, and this story is not the place to make them. */}
    <footer className="status-bar" aria-label="Status bar"><span>LOCAL SHELL</span><code data-testid="engine-snapshot">{engineLabel}</code><span className="status-spacer" />{canvas && <span data-testid="template-font-count">{`${canvas.fontFamilies.length} font${canvas.fontFamilies.length === 1 ? '' : 's'} in template`}</span>}<span role="status" aria-live="polite" aria-label="Offline availability" data-testid="offline-status">{offlineLabel}</span><code>{mode.toUpperCase()} MODE</code></footer>
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
// `empty` is what the row says when it holds NOTHING — 'none', 'black',
// 'always', a grey word standing in for behaviour the document does not
// author. STORY 17.3 adds `shown`, which says that string is not a stand-in at
// all but the ENGINE'S OWN EFFECTIVE VALUE for this field, so the box carries
// it as real text the author can read, step and commit. Only `fontSize` and
// `lineSpacing` set it, and both take their string from the projection — never
// from a literal in this file.
type FieldSpec = Readonly<{ field: PropertyField; label: string; affix?: string; unit?: string; swatch?: true; prose?: true; empty?: string; shown?: true; fx?: FieldExpression }>
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
// reads back as "1.5".
//
// STORY 17.3 TOOK THE `'1'` OUT OF THIS LINE. The neutral ratio is the
// ENGINE'S number — `defaultLineSpacing` in render.go, which is
// template.LineSpacingUnit — and spelling it here made the designer a second
// authority on it: if the engine's default ever moved, this string would have
// gone on claiming the old one and nothing would have reddened. It is now
// projected (CanvasProjection.defaultLineSpacing) and supplied at the render
// site, exactly as `defaultFontSize` already was for the size beside it.
const lineSpacingField: FieldSpec = { field: 'lineSpacing', label: 'Line spacing', affix: 'Leading' }
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
function ComponentProperties({ components, fontFamilies, fontChains, carriedFaces, specimenBytes, defaultFontSize, defaultLineSpacing, onCommit, onUseFamily, onOpenFontBrowser, browserOpen, storedFaces, fontChainError, fontChainBusy, documentGeneration, propertyError, drag, onEditTable, onPickImage, imageAvailable, assetBusy, assetError }: { components: ReadonlyArray<PanelComponent>; fontFamilies: ReadonlyArray<string>; fontChains: CanvasProjection['fontChains']; carriedFaces: ReadonlySet<string>; specimenBytes: PreviewFaceBytes; defaultFontSize: number; defaultLineSpacing: number; onCommit: CommitProperties; onUseFamily: (source: FamilySource) => Promise<string | undefined>; onOpenFontBrowser: () => void; browserOpen: boolean; storedFaces: ReadonlyArray<StoredFace>; fontChainError?: FontChainCommitError; fontChainBusy: boolean; documentGeneration: number; propertyError?: PropertyCommitError; drag?: DragState; onEditTable: (id: string) => void; onPickImage: (id: string) => void; imageAvailable: boolean; assetBusy: boolean; assetError?: Readonly<{ id: string; message: string }> }) {
  const ids = components.map((component) => component.id)
  const types = new Set(components.map((component) => component.type))
  const all = (predicate: (type: PanelComponent['type']) => boolean) => [...types].every(predicate)
  const single = components.length === 1 ? components[0]! : undefined
  const scopedError = propertyError?.selectionKey === ids.join(',') ? propertyError : undefined
  const scopedChainError = fontChainError?.selectionKey === ids.join(',') ? fontChainError : undefined
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
    {typographic && <PropertySection title="TYPOGRAPHY"><FontFamilyProperty families={fontFamilies} fontChains={fontChains} carriedFaces={carriedFaces} specimenBytes={specimenBytes} components={components} ids={ids} onCommit={onCommit} onUseFamily={onUseFamily} onOpenFontBrowser={onOpenFontBrowser} browserOpen={browserOpen} storedFaces={storedFaces} pickBusy={fontChainBusy} pickError={scopedChainError?.control.action === 'embed' ? scopedChainError : undefined} documentGeneration={documentGeneration} error={scopedError?.field === 'fontFamily' ? scopedError : undefined} /><div className="property-size-row">{draftFor({ ...fontSizeField, empty: points(defaultFontSize), shown: true })}<div className="property-toggle-row"><BooleanProperty label="Bold" field="bold" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'bold' ? scopedError : undefined} /><BooleanProperty label="Italic" field="italic" components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'italic' ? scopedError : undefined} /></div></div>{draftFor({ ...lineSpacingField, empty: points(defaultLineSpacing), shown: true })}{draftFor(colorField)}<div className="property-grid"><SegmentedProperty label="Align" field="align" segments={alignChoices} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'align' ? scopedError : undefined} /><SegmentedProperty label="Vertical align" field="valign" segments={valignSegments} components={components} ids={ids} onCommit={onCommit} documentGeneration={documentGeneration} error={scopedError?.field === 'valign' ? scopedError : undefined} /></div></PropertySection>}
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
// STORY 17.4. THE EXACT INVERSE OF `points`, and the only reader on the arrow
// step's path: a plain decimal with AT MOST THREE PLACES, read into INTEGER
// thousandths.
//
// The trap this exists to close is arithmetic, not keys. Every value in the
// inspector is a decimal string Go parses exactly and refuses beyond three
// places (`internal/template/decimal.go`: `has more than three decimal
// places`), and it is passed through UNQUOTED. So the decimal itself is never
// added to: only its DIGIT GROUPS are read as integers, and every step, clamp
// and comparison downstream is integer arithmetic on thousandths.
//
// MEASURED, because the obvious illustration of the hazard is wrong: `1 + 0.1`
// is EXACTLY `1.1` in IEEE doubles, so a single step off a round number would
// survive float arithmetic and prove nothing. The damage begins on the SECOND
// step — `1.1 + 0.1` is `1.2000000000000002` — and compounds from there. That
// is why the guard is the arithmetic itself rather than a check on the result,
// and why the test that covers it steps repeatedly.
//
// A draft this refuses is NOT STEPPABLE and the arrow does nothing: an empty
// box, a mixed selection (which presents as an empty draft), `abc`, a fourth
// decimal place, `1e3`, or a magnitude past the exact-integer range. That
// refusal IS the guard — there is no path on which a float could be produced,
// and none on which an unreadable literal could be sent.
function draftThousandths(text: string): number | undefined {
  const parts = /^(-?)(\d+)(?:\.(\d{1,3}))?$/.exec(text)
  if (!parts) return undefined
  const magnitude = Number.parseInt(parts[2] as string, 10) * 1000 + Number.parseInt(((parts[3] ?? '') as string).padEnd(3, '0'), 10)
  if (!Number.isSafeInteger(magnitude)) return undefined
  return parts[1] === '-' ? -magnitude : magnitude
}
function PropertyDraft({ spec, components, ids, onCommit, documentGeneration, live, error }: { spec: FieldSpec; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; documentGeneration: number; live?: string; error?: PropertyCommitError }) {
  const { field, label, affix, unit, swatch, prose, empty, shown, fx } = spec
  const values = components.map((component) => committedValue(component, field))
  const same = values.every((value) => value === values[0])
  // `committed` IS THE DOCUMENT'S OWN VALUE AND MUST STAY SO — `''` when the
  // key is absent. Story 17.3 puts the engine's default in the BOX, never in
  // here: `commit()` below is `if (draft !== committed)`, so folding the
  // default into `committed` would make that comparison false and committing
  // the shown default would send NOTHING while every gate stayed green.
  const committed = same ? values[0] ?? '' : ''
  // STORY 17.3. What the box READS when the document says nothing. For every
  // field but two that is the empty string and this is the identity function;
  // for `fontSize` and `lineSpacing` it is the engine's own effective value,
  // arriving as `empty` from the projection with `shown` marking it real.
  //
  // NOT APPLIED TO A MIXED SELECTION. `same` is false there, the components
  // genuinely disagree, and filling in one number would both lie about them
  // and — through the arrow step, which reads the draft — put a flattening
  // edit one nudge key away. Mixed keeps its empty draft and its `Mixed`
  // placeholder.
  //
  // AND IT WRITES NOTHING BY ITSELF. This is display state: no command is sent
  // until the author commits the field, which is the safety property the whole
  // story rests on — opening a document may never mutate it.
  const inherited = (text: string): string => text === '' && same && shown === true && empty !== undefined ? empty : text
  const [draft, setDraft] = useState(inherited(committed))
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  // A canvas drag, resize or nudge commits geometry without this field ever
  // being touched. Follow the engine on a committed transition; a draft the
  // engine has not accepted (a rejected commit leaves the value unchanged)
  // still survives, because nothing transitioned.
  const [lastCommitted, setLastCommitted] = useState(committed)
  if (lastCommitted !== committed) { setLastCommitted(committed); setDraft(inherited(committed)) }
  const selectionKey = ids.join(',')
  const revert = () => setDraft(inherited(committed))
  // `disable` is the ONE thing the arrow step varies, and it is not a second
  // commit path: the intent, the encoder, the reconciliation and the
  // single-flight `pendingRef` guard are all shared verbatim. `shared` carries
  // `disabled: pending`, which exists so a keystroke cannot race a commit
  // already in flight — but an arrow step IS that keystroke, and disabling a
  // FOCUSED input moves focus to the body and does not give it back when the
  // input is re-enabled (measured in Chromium 1217). Key repeat is delivered to
  // the focused element, so raising `pending` here would end an arrow HOLD after
  // exactly one step.
  //
  // A step that arrives mid-flight is still dropped by `pendingRef`, and what
  // happens to it then is a RACE, not an accumulation: the draft has already
  // advanced, so if the next repeat beats the engine's answer it carries the
  // accumulated value, but if the answer wins, the committed transition below
  // (`lastCommitted !== committed`) rewrites the draft to the engine's value
  // and that press is lost. A hold therefore steps at the round-trip rate, not
  // the repeat rate. Coalescing repeats into one command is the real fix and is
  // out of this story's scope — it also bears on undo, since each step is its
  // own revision and its own undo entry.
  const submit = async (intent: PropertyIntent, reconcileDraft: boolean, disable = true) => {
    if (pendingRef.current) return
    pendingRef.current = true
    if (disable) setPending(true)
    const accepted = await onCommit(ids, intent, documentGeneration, selectionKey)
    pendingRef.current = false
    if (disable) setPending(false)
    // Through `inherited` for the CLEAR case: the engine's answer for a
    // cleared key is `''`, and the box must come back to the shown default
    // rather than to an empty row. Without this wrapper the committed
    // transition above would set the default and this line would immediately
    // overwrite it with the empty string, whichever order they landed in.
    if (accepted && reconcileDraft) setDraft(inherited(canonicalValue(accepted, ids, field) ?? draft))
  }
  // The `else` arm is Story 17.3's, and it is not a commit: emptying a box
  // whose key is ALREADY absent leaves `draft === committed === ''`, so there
  // is nothing to send — and nothing was sent before this story either. What
  // changes is what the author is left looking at. The row must come back to
  // the value it inherits, the same one it opened on, instead of sitting blank
  // beside a canvas that is still painting 12.
  const commit = async () => { if (draft !== committed) await submit({ field: contentCommand(field, draft), operation: draft === '' && field !== 'value' && field !== 'expression' ? 'clear' : 'set', value: draft }, true); else setDraft(inherited(committed)) }
  // STORY 17.4: ARROWS STEP A NUMBER FIELD.
  //
  // THE NUMERIC SET IS THE ONE THE CONTROL ALREADY KNOWS. This predicate was
  // computed inline for `inputMode`; it is hoisted rather than restated, so a
  // field can never be typeable as a decimal and unsteppable, or the reverse.
  // It is exactly x, y, width, height, fontSize, borderWidth and lineSpacing.
  const numeric = unit === 'pt' || unit === undefined && (field === 'lineSpacing')
  // THE STEP IS DERIVED FROM THE FIELD, NOT A CONSTANT. A point field steps by
  // one POINT, the same increment the canvas nudge already uses; leading is a
  // dimensionless RATIO and steps by a tenth. Both are written in the
  // thousandths the arithmetic runs in. 0.001 is the floor of the
  // representation, never a step.
  const stepThousandths = field === 'lineSpacing' ? 100 : 1_000
  // The bounds are the ENGINE'S OWN, read from the places that declare them
  // rather than restated: `POSITIVE_LENGTH_FIELDS` is the four keys
  // `component_commands.go` refuses at or below zero, so their smallest legal
  // value is one thousandth; `ORIGIN_FLOOR_FIELDS` is `x` and `y`, which
  // `containComponent` refuses BELOW ZERO on this same command path; and
  // lineSpacing's pair is engine-protocol's mirror of `linespacing.go`.
  //
  // ⚠ THIS IS NOT THE WHOLE OF `containComponent`. It also bounds x, y, width
  // and height ABOVE against the band extents, and the arrow step does NOT
  // clamp to those — a step at the band edge still reaches the engine's own
  // located refusal. That is an OPEN question recorded in the story's Spec
  // Change Log, not a settled exclusion: the bound is per-component (two
  // components with equal widths at different x have different width ceilings),
  // so a selection-wide clamp needs a ruling this story does not carry.
  const lowest = field === 'lineSpacing' ? MIN_LINE_SPACING_THOUSANDTHS : POSITIVE_LENGTH_FIELDS.includes(field) ? 1 : ORIGIN_FLOOR_FIELDS.includes(field) ? 0 : undefined
  const highest = field === 'lineSpacing' ? MAX_LINE_SPACING_THOUSANDTHS : undefined
  // Returns whether the arrow was HANDLED, which is what suppresses the
  // browser's own caret jump — an unhandled arrow keeps it, on a non-numeric
  // field, during a drag, and on a draft with no value in it to step.
  const step = (direction: 1 | -1): boolean => {
    // A drag owns the geometry fields: they are `readOnly`, typing does
    // nothing, and an arrow does nothing either.
    if (!numeric || live !== undefined) return false
    // ONE PREDICATE, NOW CLOSING ONE ROW. Story 17.4 wrote this guard to close
    // TWO — an unset field and a mixed selection, both of which presented as an
    // empty draft the exact parser refuses. STORY 17.3 DISSOLVED THE FIRST OF
    // THEM, and the orchestrator retired that arm on 2026-09-04 rather than
    // carving it out: `fontSize` and `lineSpacing` now carry the engine's own
    // effective value as text, so an unset field HAS a value, it parses, and
    // ArrowUp steps from it. The precondition 17.4 rested on — "an unset field
    // has no value to step, and its placeholder is not one" — is simply no
    // longer true of these two fields, and keeping the guard would have made
    // typing `1.1` into a box reading `1` write 1.1 while ArrowUp on the same
    // visible `1` wrote nothing. That is the special case 17.4's own
    // `expect(stepped).toEqual(typed)` exists to forbid.
    //
    // 17.4's SECOND reason did not dissolve and this line still carries it. A
    // MIXED selection also presents as an empty draft, and stepping it would
    // flatten every component onto one value — a destructive edit fired by a
    // nudge key, on a field the author has not touched. Nothing in 17.3
    // reaches that: `inherited` above deliberately does not fill a mixed
    // draft. A mixed field the author has TYPED into is no longer empty and
    // steps like any other draft — they are stepping the value they entered.
    const current = draftThousandths(draft)
    if (current === undefined) return false
    const stepped = current + direction * stepThousandths
    // Floored, THEN capped, each against its own bound and neither against the
    // unclamped step: folding both into one expression let the absent ceiling
    // fall back to `stepped` and quietly undo the floor.
    const floored = lowest === undefined ? stepped : Math.max(stepped, lowest)
    const next = highest === undefined ? floored : Math.min(floored, highest)
    // Already at the bound: the arrow is still handled — the caret stays put —
    // but nothing changed, so no command is sent.
    if (next === current) return true
    const value = points(next)
    setDraft(value)
    // `reconcileDraft` is false BY DESIGN. `points` already emits the
    // canonical spelling, so there is nothing for the engine to normalise, and
    // reconciling from a resolved step would overwrite a draft a later repeat
    // had already advanced. The committed transition above
    // (`lastCommitted !== committed`) is what lands the engine's answer.
    void submit({ field, operation: 'set', value }, false, false)
    return true
  }
  // The reset action appears with the value it resets, and with a mixed
  // selection, which also has committed values behind it. An unset row is
  // otherwise empty chrome with nothing to clear on it.
  //
  // STORY 17.3 gives `fontSize` and `lineSpacing` a value even when the
  // document is silent, so `×` now appears on those two rows unconditionally.
  // That is the matrix's own reading of the control — the box says 12, so the
  // row offers to reset it, and clearing an already-absent key lands back on
  // the same 12. `clear` itself is untouched: it still sends `op:"clear"`, and
  // Go still stores the zero Presence that omits the key from the file.
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
    // Story 17.4. Enter and Escape above are untouched; the arrows are the
    // only addition, and only where the step actually took the key.
    //
    // A MODIFIED arrow is left entirely alone. This is not modifier BEHAVIOUR,
    // which the story puts out of scope — it is the absence of it: inside a
    // text input Shift+Arrow extends the selection and (on macOS) Cmd+Arrow and
    // Alt+Arrow move the caret, so stepping on a modified arrow would take
    // three shipped editing gestures away from the author to no end. Adding a
    // coarse or fine step on a modifier is the thing that needs asking for.
    if ((event.key === 'ArrowUp' || event.key === 'ArrowDown') && !event.shiftKey && !event.ctrlKey && !event.altKey && !event.metaKey) { if (step(event.key === 'ArrowUp' ? 1 : -1)) event.preventDefault() }
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
    : <input className="property-value" {...shared} inputMode={numeric ? 'decimal' : undefined} onChange={(event) => setDraft(event.target.value)} />}{fx && <span className={`property-fx${holdsExpression(fx, live ?? draft) ? ' property-fx-active' : ''}`} title={fxHint[fx]} aria-hidden="true">fx</span>}{swatch && <input type="color" className={`property-swatch${/^#[0-9a-fA-F]{6}$/.test(live ?? draft) ? '' : ' property-swatch-unset'}`} aria-label={`Pick ${label}`} value={swatchColor(live ?? draft)} disabled={pending || live !== undefined} onChange={(event) => { setDraft(event.target.value); void submit({ field, operation: 'set', value: event.target.value }, true) }} onBlur={() => void commit()} />}{unit && <span className="property-unit">{unit}</span>}{canClear && <button type="button" className="property-inline-action" aria-label={`Clear ${label}`} title={`Clear ${label}`} disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'clear' }, true)}>×</button>}{canNull && <button type="button" className="property-inline-action" aria-label={`Set ${label} null`} title={`Set ${label} null`} disabled={pending} onMouseDown={(event) => event.preventDefault()} onClick={() => void submit({ field, operation: 'null' }, true)}>∅</button>}</div>{error && <p id={errorId} role="alert" className="property-error">{error.elementId ? `${error.elementId}: ` : ''}{error.dataPath ? `${error.dataPath}: ` : ''}{error.message}</p>}</div>
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
//
// STORY 8.6 GAVE IT A SECOND GROUP; STORY 16.4 MADE IT THREE, ON THE AXIS THE
// CODE ALREADY FORKS ON — WHERE ARE THE BYTES, never when did they arrive.
// STORY 16.9 NARROWED THE DROPDOWN BACK TO TWO: the third relationship still
// exists in the code below — a family can still be not on this machine at
// all — but this control no longer OFFERS it. `Add fonts…` (the font
// browser) is the surface built for that relationship, and a menu listing
// ~1,273 rows nobody can use without a download, with no network and no
// wait, was never what the owner asked for.
//
//   1. IN THIS TEMPLATE — `families.includes(name)`, the document's own
//      declared chains. The bytes are IN THE FILE. Picking commits `fontFamily`,
//      exactly as it did before 8.6.
//   2. AVAILABLE LOCALLY — `familyIsInstalled(source)`, which is the `local` and
//      `stored` arms together. ON THIS MACHINE, NOT IN THIS FILE. Picking embeds
//      from the machine and then commits the property: two commands, two undos,
//      no network.
//
// A ROW'S GROUP IS A PURE FUNCTION OF (declared?, `familyIsInstalled`?) AND OF
// NOTHING ELSE — never of a set built up over this session. The `local`/`stored`
// split is deliberately invisible here: it is a provenance difference with no
// consequence at the moment of choosing. It used to surface at REMOVAL, in
// `lateEmbedRefusal`'s two branches — Story 16.6 deleted the removal control
// that distinction served, and the refusal below now says one sentence for
// both tiers.
//
// A FONT STILL CHANGES GROUP BECAUSE THE AUTHOR ACTED, even with the third
// group gone from this menu (Story 8.6's rule: nothing says "added", the
// entry simply moves). Installing through the font browser still moves a
// family into AVAILABLE LOCALLY the next time this dropdown opens; first use
// still moves it into IN THIS TEMPLATE. Only the affordance for the FIRST of
// those two moves left THIS control — the relationship the code forks on,
// and `offeredFamilies` itself, are unchanged.
//
// The engine stays the authority on what `fontFamily` may name. The designer
// never invents a family name — an offered name reaches the document only by
// going through a command, and the projection is what says it arrived.

/**
 * THE LATE REFUSAL, DISCLOSED RATHER THAN LEFT TO SURPRISE (Story 16.5).
 *
 * Install runs every admission check that can run at install. ONE CANNOT MOVE:
 * Go's nameID-13 licence-signature tie reads the licence written inside the face
 * and refuses a face whose own bytes contradict its declared terms
 * (`folio-go/internal/fontset/licencesignature.go`). Porting it would need a
 * second name-table reader and two regex tables in this designer — a competing
 * authority over what enters a document — so it stays in Go, and the residue is
 * that a face can install successfully and be refused the first time it is used.
 *
 * A DEAD END THE AUTHOR CAN SEE IS A STATED LIMIT; A DEAD END THAT SIMPLY FAILS
 * IS NOT. So the sentence says what the engine's own refusal cannot: that the
 * face IS on this machine and that nothing was written to the document.
 *
 * ONE SENTENCE FOR BOTH TIERS (Story 16.6). It used to point at a per-face
 * remove control for a stored face and say a bundled one had none — the
 * distinction the family control's `local`/`stored` split still tracks above,
 * for grouping only. That control is gone, so there is no remedy left to
 * offer and nothing left to tell the two tiers apart by.
 */
const lateEmbedRefusal = (family: string, engineMessage: string): string =>
  `${family} is installed on this machine and cannot be embedded in this document: ${engineMessage} Nothing was written to the document, and the face is still on this machine.`

// THE FAMILY CONTROL'S OWN SAMPLE TEXT (Story 16.7), SHORT ON PURPOSE AND
// DELIBERATELY NOT `font-browser-model.ts`'s `latinSample`/`thaiSample`.
// The design draws "a few letters set in that typeface" beside the name —
// `Aa Bb 123`, `กขค Aa` — never the browser's full sentence, which measured
// true in a real dropdown: at the panel's width it left three or four
// characters of the NAME before the specimen's ellipsis took over, on every
// row, which is the opposite of a control whose job is to show a name AND a
// face. Reusing the mechanism (the registry, the honesty rule, `lang="th"`)
// is Story 16.7's contract; reusing the browser's own sample sentence is not
// part of it, and the mockup never drew that sentence here to begin with.
const familyControlLatinSample = 'Aa Bb 123'
const familyControlThaiSample = 'กขค Aa'

// STORY 16.7 — RESOLVING A DECLARED CHAIN TO THE FACE IT PAINTS WITH, THE SAME
// WAY `TextPaint` RESOLVES A FRAGMENT (below, in this file): a carried entry
// counts only once its asset key is actually in `carriedFaces` (the paintable
// set built at the top of this component from `carriedFaceKeys`), and a
// shipped entry counts by its own engine name. ONLY THE CHAIN'S FIRST ENTRY IS
// CONSULTED. The engine alone decides, per glyph, which entry a real paragraph
// actually falls back to — that decision is AD-17's, and re-deriving it here
// against a fixed sample string is exactly the browser-side measurement AD-17
// forbids. The first entry is the chain's own declared primary, so a specimen
// set in it is an honest answer to "what face is this" even where a longer
// paragraph might fall through to the second.
//
// RETURNED AS `declaredEntry`, DELIBERATELY NOT `entry`. `canvas-font-stack.
// test.ts`'s authority census anchors its shipped-face pattern to the literal
// identifier `fragment` and poisons `entry.face` by name, because a document's
// declared chain entry is not the engine's attribution for a PAINTED fragment
// and must never stand in for one in the canvas's own paint path — the
// hazard argued at `TextPaint`, below. This value never reaches that path: it
// sets one `aria-hidden` specimen `<span>` with FIXED sample text, nothing is
// measured, and there is no sibling fragment for a mismatched entry to
// collide with. `declaredEntry` names that difference rather than hiding it.
function declaredChainEntry(name: string, chains: CanvasProjection['fontChains']): CanvasProjection['fontChains'][number]['entries'][number] | undefined {
  return chains.find((chain) => chain.name === name)?.entries[0]
}

/**
 * BEST-EFFORT SCRIPT COVERAGE FOR A DECLARED ROW'S SAMPLE TEXT. A chain
 * entry's own `family` is DISPLAY identity — the name a pick recorded, e.g.
 * `Inter` — so it is looked up the same two places every other row's coverage
 * comes from (`font-index.ts`'s local tier, then the snapshot), never
 * invented. A miss here — the two committed faces with no index row, or a
 * built-in engine name like `Noto Sans Thai` that names no offered family at
 * all — costs only which SAMPLE prints, never which face the row is set in.
 */
function scriptsForFamilyName(family: string): ReadonlyArray<string> {
  if (family === '') return []
  return catalogueFaces.find((face) => face.family === family)?.scripts ?? indexRowFor(family)?.scripts ?? []
}

function FontFamilyProperty({ families, fontChains, carriedFaces, specimenBytes, components, ids, onCommit, onUseFamily, onOpenFontBrowser, browserOpen, storedFaces, pickBusy, pickError, documentGeneration, error }: { families: ReadonlyArray<string>; fontChains: CanvasProjection['fontChains']; carriedFaces: ReadonlySet<string>; specimenBytes: PreviewFaceBytes; components: ReadonlyArray<PanelComponent>; ids: ReadonlyArray<string>; onCommit: CommitProperties; onUseFamily: (source: FamilySource) => Promise<string | undefined>; onOpenFontBrowser: () => void; browserOpen: boolean; storedFaces: ReadonlyArray<StoredFace>; pickBusy: boolean; pickError?: FontChainCommitError; documentGeneration: number; error?: PropertyCommitError }) {
  const values = components.map((component) => committedValue(component, 'fontFamily'))
  const uniform = values.every((value) => value === values[0])
  const committed = uniform ? values[0] ?? '' : ''
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [active, setActive] = useState(0)
  const [pending, setPending] = useState(false)
  const pendingRef = useRef(false)
  const listId = useId()
  // FOCUS RETURNS HERE WHEN THE BROWSER CLOSES (UX-DR25).
  //
  // THE CONTROL THAT OPENED IT CANNOT TAKE FOCUS BACK, WHICH IS WHY THIS IS THE
  // OWNER. The `Add fonts…` button lives inside the open dropdown and the
  // dropdown is closed on the way into the modal, so by the time the modal
  // unmounts its invoker is gone from the document — focus would land on
  // `<body>`, and a keyboard-only author would have to tab in from the top of
  // the page after every Escape. The family combobox is the control the door
  // belonged to and the one the author was working in, so it is where focus
  // goes: on Escape, on Cancel, on ×, and after a successful add.
  const field = useRef<HTMLInputElement>(null)
  const browserWasOpen = useRef(false)
  useEffect(() => {
    if (browserWasOpen.current && !browserOpen) field.current?.focus()
    browserWasOpen.current = browserOpen
  }, [browserOpen])
  // STORY 16.7 — THIS CONTROL'S OWN PREVIEW-FACE REGISTRY, ITS OWN INSTANCE.
  //
  // OPENED ON MOUNT AND CLOSED ON UNMOUNT, the same shape `FontBrowser.tsx`
  // opens its own in — this component just stays mounted across many opens and
  // closes of the DROPDOWN, where the modal only ever exists for one. "A face
  // is registered only while the dropdown is open, and released when it
  // closes" is enforced by the SHOW effect below rather than by tearing the
  // registry down every time: `show([])` releases every family it holds, so a
  // closed dropdown holds exactly zero preview faces on `document.fonts` —
  // observably identical to a torn-down registry — without re-losing every
  // in-flight fetch (and re-declining every failed one) on each reopen.
  //
  // THE RESOLVER IS READ THROUGH A REF for the same reason `FontBrowser` reads
  // its own through one: a caller re-creating the closure every render must
  // not tear the registry down and re-fetch everything on every keystroke.
  const specimenBytesRef = useRef(specimenBytes)
  useEffect(() => { specimenBytesRef.current = specimenBytes })
  const [specimenTick, setSpecimenTick] = useState(0)
  const [registry, setRegistry] = useState<PreviewFaceRegistry>()
  useEffect(() => {
    const opened = openPreviewFaceRegistry((family) => specimenBytesRef.current(family), () => setSpecimenTick((tick) => tick + 1))
    setRegistry(opened)
    return () => opened.close()
  }, [])
  const specimenStatus = (family: string): PreviewFaceStatus => {
    void specimenTick
    return registry?.statusOf(family) ?? 'preparing'
  }
  const needle = query.trim().toLowerCase()
  const hit = (name: string) => needle === '' || name.toLowerCase().includes(needle)
  const declared = families.filter(hit)
  // A catalogue family the document ALREADY declares a chain for is not
  // offered twice: the pick named the chain after the family, so the entry has
  // moved into the first group and showing it in both would make "declared"
  // and "not yet declared" stop meaning anything.
  // STORY 16.1: THE SECOND GROUP IS NO LONGER 21 ROWS. It is the local face
  // tier — the same 21 committed faces, which need no network — followed by the
  // families in the designer's build-time snapshot of the published library.
  // `offeredFamilies` owns the join and the filter; this control owns only the
  // one exclusion that is about THIS DOCUMENT: a family the document already
  // declares a chain for has moved into the first group and is not offered twice.
  // THE PARTITION. One predicate, two groups: `declared` above is
  // `families.includes(name)`, and `familyIsInstalled` is the one definition
  // of "this machine already holds it", shared with the browser's row state
  // so the two surfaces cannot disagree.
  //
  // STORY 16.9 REMOVED THE THIRD GROUP, AVAILABLE TO INSTALL. `offeredFamilies`
  // itself is untouched and still returns the web tier for the font browser,
  // which still draws it (`FontBrowser.tsx`) — this control simply never asks
  // for it, by filtering to `familyIsInstalled` before anything else. Nothing
  // downstream of this line ever sees a web-tier row, which is what makes
  // opening this dropdown a zero-fetch operation: a row that is never in
  // `onThisMachine` is never registered for a specimen, never rendered, and
  // never reachable by a pick.
  const onThisMachine = offeredFamilies(query, storedFaces).filter((source) => !families.includes(source.family) && familyIsInstalled(source))
  // THE REGISTRY HOLDS EXACTLY THE FAMILIES THIS RENDER CAN SHOW A SPECIMEN
  // FOR — `onThisMachine`, and nothing else. NUL-JOINED for the reason
  // `FontBrowser.tsx`'s own key is: family names contain spaces. EMPTY
  // WHENEVER THE DROPDOWN IS CLOSED, so a stale key from the last time it was
  // open cannot re-open the registry's `show`.
  const installedFamilyKey = open ? onThisMachine.map((source) => source.family).join(' ') : ''
  useEffect(() => { registry?.show(installedFamilyKey === '' ? [] : installedFamilyKey.split(' ')) }, [registry, installedFamilyKey])
  // THE SCRIPT COVERAGE FOR EVERY `AVAILABLE LOCALLY` ROW, FROM THE SAME
  // DERIVATION THE FONT BROWSER'S OWN ROWS USE, so the two surfaces cannot
  // describe one family's coverage two different ways.
  const installedRowByFamily = new Map(browserRows(onThisMachine).map((row) => [row.family, row]))
  // GROUP 2 IS DELIBERATELY UNCAPPED, AND THE REVISIT TRIGGER IS NAMED RATHER
  // THAN LEFT TO BE NOTICED. Its population is the 31 committed faces plus
  // whatever this designer has downloaded, so it is tens of rows and not
  // thousands, and a heading saying a font is already on your machine may not
  // hide one. REVISIT WHEN THE MACHINE STORE CAN HOLD ON THE ORDER OF 200
  // ENTRIES: at that size the group needs its own bound, and the bound has to
  // arrive with something that still tells the truth about what it hid.
  //
  // THERE IS NO THIRD GROUP TO CAP (Story 16.9). The ~1,273 families this
  // machine does not hold are no longer offered here at all — `Add fonts…`
  // below is the only door to them — so the render-limit problem a cap once
  // solved does not exist in this control any more.
  //
  // ONE flat option list behind two visible groups, because the keyboard is
  // linear even when the list is not: `active` indexes this, arrow keys walk
  // it, and Enter dispatches whichever kind it lands on.
  const matches: ReadonlyArray<{ name: string; source?: FamilySource }> = [...declared.map((name) => ({ name })), ...onThisMachine.map((source) => ({ name: source.family, source }))]
  // THE TWO GROUPS, EACH OVER ITS OWN SLICE OF `matches`, so a row's option id
  // and its arrow-key position stay the flat index whatever the grouping does.
  // A heading is drawn only when its own group has rows after filtering.
  const groups: ReadonlyArray<{ key: string; label: string; from: number; rows: ReadonlyArray<{ name: string; source?: FamilySource }> }> = [
    { key: 'template', label: 'IN THIS TEMPLATE', from: 0, rows: matches.slice(0, declared.length) },
    { key: 'local', label: 'AVAILABLE LOCALLY', from: declared.length, rows: matches.slice(declared.length) },
  ]
  const close = () => { setOpen(false); setQuery(''); setActive(0) }
  const commit = async (intent: PropertyIntent) => {
    if (pendingRef.current) return
    pendingRef.current = true
    setPending(true)
    await onCommit(ids, intent, documentGeneration, ids.join(','))
    pendingRef.current = false
    setPending(false)
  }
  // FIRST USE — THE FORK'S OWN HALF FOR AN INSTALLED FAMILY. `onUseFamily` sends the embed and
  // resolves to a refusal sentence or to nothing; the property is committed ONLY
  // after it returns nothing, because `canvas.fontFamilies` is the closed set
  // `style.fontFamily` may name and the property command is refused until the
  // chain is declared. The engine forces the order; nothing here chooses it.
  //
  // TWO COMMANDS AND TWO UNDO ENTRIES, NOT ONE OF EITHER. There is no compound
  // command in this product and Story 8.6's refused fusion is not reopened.
  //
  // THE PENDING FLAG IS RELEASED BEFORE `commit`, because `commit` takes it
  // itself; holding it across both would make the second command drop silently.
  const commitFirstUse = async (source: FamilySource) => {
    if (pendingRef.current) return
    pendingRef.current = true
    setPending(true)
    let refusal: string | undefined
    try {
      refusal = await onUseFamily(source)
    } finally {
      pendingRef.current = false
      setPending(false)
    }
    if (refusal !== undefined) return
    await commit({ field: 'fontFamily', operation: 'set', value: source.family })
  }
  // THE FORK. A declared name is a property commit — today's behaviour, byte
  // for byte. A row carrying a `source` is always a family this machine
  // already holds: STORY 16.9 removed the dropdown's install-tier group, so
  // `match.source` can no longer name a family that is NOT on this machine
  // (`onThisMachine`, above, is the only source of a `source`-bearing row).
  //
  // TAKING IT IS THE FIRST USE — the moment the font starts travelling inside
  // the template — so `commitFirstUse` embeds from the machine and then
  // commits the property: the two decisions ("carry this typeface" and "draw
  // this box with it") are taken together here because the author took them
  // together. They are still TWO COMMANDS and two separately undoable
  // entries; only the trigger is one gesture. Fusing them into one command
  // would make that undo ambiguous, and there is no mechanism to fuse them
  // with (Story 8.6's refused fusion is not reopened).
  const choose = (match: { name: string; source?: FamilySource }) => {
    close()
    if (match.source) { void commitFirstUse(match.source); return }
    void commit({ field: 'fontFamily', operation: 'set', value: match.name })
  }
  const move = (step: number) => { if (matches.length > 0) setActive((current) => (current + step + matches.length) % matches.length) }
  // THE SPECIMEN, DRAWN FOR BOTH GROUPS THIS CONTROL RENDERS (Design Note 2 of
  // Story 16.7). STORY 16.9 REMOVED THE THIRD GROUP, `AVAILABLE TO INSTALL`,
  // which used to keep a per-row note instead of a specimen — that branch and
  // its row are both gone from this control now, not merely unreached.
  //
  // OMITTED ENTIRELY UNLESS THE FACE IS READY OR ALREADY ON THE PAGE. A row
  // whose face is still preparing, or that never resolves to one at all, shows
  // no specimen and never a sample set in a substitute face — the honesty rule
  // `FontBrowser.tsx:195-201` states in words, restated here as an omission
  // because a dropdown row has no room for the sentence.
  const specimenNode = (match: { name: string; source?: FamilySource }) => {
    if (match.source) {
      // EVERY `match.source` HERE IS A FACE THIS MACHINE ALREADY HOLDS.
      // `onThisMachine` is the only source of a `source`-bearing row now that
      // the third group is gone, so `installedRowByFamily` always has an
      // entry for one.
      const row = installedRowByFamily.get(match.source.family)
      if (row === undefined || specimenStatus(row.family) !== 'ready' || previewFaceFamily(row.family) === undefined) return undefined
      return <span className="property-option-specimen" aria-hidden="true" lang={row.scripts.includes('thai') ? 'th' : undefined} style={{ fontFamily: previewFaceFamily(row.family) } as CSSProperties}>{row.scripts.includes('thai') ? familyControlThaiSample : familyControlLatinSample}</span>
    }
    // A DECLARED CHAIN SHOWS THE FACE IT PAINTS WITH (Design Note 3), resolved
    // the way `TextPaint` resolves a fragment: a carried entry counts only
    // once its asset key is actually in `carriedFaces`, a shipped entry counts
    // by its own engine name — and NEITHER GOES THROUGH THIS CONTROL'S OWN
    // REGISTRY AT ALL, because the face is already on `document.fonts`,
    // registered for the open document or shipped with the build. Only the
    // chain's first entry is consulted; the engine alone decides per glyph
    // which entry a real paragraph actually uses (AD-17), and re-deriving
    // that here from a fixed sample string is the measurement AD-17 forbids.
    //
    // THE TWO BRANCHES ARE WRITTEN OUT RATHER THAN FUNNELLED THROUGH A SHARED
    // "resolve to a family" HELPER, so that each `fontFamily:` position in
    // this file names the exact approved derivation it uses in the clear —
    // `canvas-font-stack.test.ts`'s authority census reads designer source as
    // text, and a helper that returned a plain string would read as an
    // unapproved `css` at the one place that census actually looks.
    const declaredEntry = declaredChainEntry(match.name, fontChains)
    if (declaredEntry === undefined) return undefined
    const scripts = scriptsForFamilyName(declaredEntry.family)
    const lang = scripts.includes('thai') ? 'th' : undefined
    const sample = scripts.includes('thai') ? familyControlThaiSample : familyControlLatinSample
    if (isCarriedFaceAssetKey(declaredEntry.assetKey) && carriedFaces.has(declaredEntry.assetKey)) {
      return <span className="property-option-specimen" aria-hidden="true" lang={lang} style={{ fontFamily: embeddedFaceFamily(declaredEntry.assetKey) } as CSSProperties}>{sample}</span>
    }
    if (isShippedFaceName(declaredEntry.face)) {
      return <span className="property-option-specimen" aria-hidden="true" lang={lang} style={{ fontFamily: shippedFaceFamily(declaredEntry.face) } as CSSProperties}>{sample}</span>
    }
    return undefined
  }
  const errorId = error ? 'property-error-fontFamily' : undefined
  return <div className="property-editor property-combobox" onBlur={(event) => { if (!event.currentTarget.contains(event.relatedTarget as Node | null)) close() }}>
    <div className="property-field">
      <input ref={field} className="property-value property-value-prose" role="combobox" aria-label="Font family" aria-expanded={open} aria-controls={listId} aria-autocomplete="list" aria-activedescendant={open && matches.length > 0 ? `${listId}-${active}` : undefined} aria-description={uniform ? undefined : 'Mixed value'} aria-invalid={error ? 'true' : undefined} aria-errormessage={errorId} disabled={pending || pickBusy} value={open ? query : committed} placeholder={!uniform ? 'Mixed' : open ? 'Search fonts' : 'Choose a font'} onFocus={() => setOpen(true)} onChange={(event) => { setOpen(true); setQuery(event.target.value); setActive(0) }} onKeyDown={(event) => {
        if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); setOpen(true); move(event.key === 'ArrowDown' ? 1 : -1); return }
        if (event.key === 'Enter') { event.preventDefault(); const match = matches[active]; if (open && match) choose(match); return }
        if (event.key === 'Escape') { event.preventDefault(); close() }
      }} />
      {/* THE GLYPH IS THE MOCKUP'S OWN SVG PATH, NOT A TEXT CHARACTER.
          `Font Browser.dc.html:185` draws this exact chevron —
          `<svg width="8" height="8" viewBox="0 0 8 8" ...><path d="M1.5
          3l2.5 2.5L6.5 3">` — as ONE glyph, unconditionally: the mockup's
          `dropdownOpen` flag there gates the menu panel, not the chevron, so
          this button draws the same downward stroke whether open or closed.
          A first attempt swapped the text glyphs `⌃`/`⌄` (off-centre in
          OPPOSITE directions when measured by canvas ink-extent sampling —
          U+2303/U+2304 are accent-style glyphs anchored near cap-height, not
          shapes meant to fill their line) for `▲`/`▼`, which measured centred
          — but a filled triangle is the wrong SHAPE next to the design's thin
          stroke, the fix kept a two-state flip the design never had, and any
          text glyph centres by the FONT's metrics, not the box, so a UI
          typeface swap could silently reintroduce an off-centre glyph. An
          SVG box has no baseline: `.property-inline-action`'s own `display:
          grid; place-items: center` centres it structurally, the same way
          `align-items: center` centres the mockup's chevron in its row — a
          box property, not a font property, so it cannot drift when a face
          changes. `stroke="currentColor"` inherits this button's own
          ghost/hover/disabled color from `.property-inline-action` rather
          than hard-coding the mockup's static `#aab2bb`. */}
      <button type="button" className="property-inline-action property-disclosure" aria-label={open ? 'Hide fonts' : 'Show fonts'} title={open ? 'Hide fonts' : 'Show fonts'} disabled={pending} tabIndex={-1} onMouseDown={(event) => event.preventDefault()} onClick={() => (open ? close() : setOpen(true))}><svg aria-hidden="true" width="8" height="8" viewBox="0 0 8 8" fill="none" stroke="currentColor" strokeWidth="1.2"><path d="M1.5 3l2.5 2.5L6.5 3" /></svg></button>
      {/* STORY 16.9 REMOVED THE THIRD BUTTON THAT REVEALED THE CHAIN EDITOR,
          AND THE CLEAR BUTTON BESIDE IT. Text always has a typeface — there
          is no such thing as text with none — so a control offering to leave
          the field empty was offering something the product cannot do; see
          `FontChainEditor.tsx`'s own deletion and the code comment on
          `fontFamilies` for the chain data this UI removal leaves untouched. */}
    </div>
    {/* THE LISTBOX OWNS OPTIONS AND GROUPS OF OPTIONS, AND NOTHING ELSE (Story
        16.4, closing 8.6's deferral rather than multiplying it; narrowed to
        two groups by Story 16.9). It carried SIX `role="presentation"`
        children — two headings, an empty state, the disclosure, the cap note
        and the disk-font decline — which breaks a listbox's
        required-owned-elements rule. Both sanctioned repairs are used, each
        where it fits:

          · THE HEADINGS BECOME `role="group"` WITH AN `aria-label`. The visible
            heading is `aria-hidden` and the group carries the same words as its
            name, so the accessibility tree sees listbox → group → option with no
            stray text node in it, and a sighted reader still reads the heading.
          · THE NOTES MOVE OUT OF THE LIST ENTIRELY and are referenced with
            `aria-describedby`. They were never options: they describe the list.

        AND THE KEYBOARD WALK IS UNCHANGED, WHICH IS WHY THIS IS THE FIX AND NOT
        A SECOND DELIVERABLE. The one element in the walk that read POSITION
        semantically was the heading interleave this replaces (`index ===
        declared.length`); `move`, `active`, the option ids, `aria-activedescendant`
        and `choose` are all order-agnostic and are untouched. `active` still
        indexes the one flat `matches` array, and the groups are drawn over
        contiguous slices of it. */}
    {open && <div className="property-options">
      {/* NO CLASS ON THIS ELEMENT, DELIBERATELY. It carried
          `property-option-groups`, which `App.css` styled in zero places — a
          name that looks like a styling hook and is not one costs a reader the
          search. The shell above owns every box property, so the list needs no
          rule of its own; it is addressable by its role and its id. */}
      <div id={listId} role="listbox" aria-label="Fonts" aria-describedby={matches.length === 0 ? `${listId}-notes` : undefined}>
        {groups.filter((group) => group.rows.length > 0).map((group) => <div key={group.key} className={`property-option-group property-option-group-${group.key}`} role="group" aria-label={group.label}>
          <p className="property-option-heading" aria-hidden="true">{group.label}</p>
          {group.rows.map((match, offset) => {
            const index = group.from + offset
            // THE SPECIMEN REPLACES THE PER-ROW NOTE (Design Note 2 of Story
            // 16.7): a family this machine already holds, or a chain the
            // document already declares, draws a specimen instead of
            // restating what its own group heading already says. STORY 16.9
            // removed the one group whose rows fell back to a per-row note
            // instead — `AVAILABLE TO INSTALL` — so every row here now draws
            // a specimen or nothing, never a note.
            return <div key={`${group.key}:${match.name}`} id={`${listId}-${index}`} role="option" aria-selected={match.source === undefined && match.name === committed} className={`property-option${index === active ? ' property-option-active' : ''}${match.source ? ' property-option-catalogue' : ''}`} onMouseDown={(event) => event.preventDefault()} onMouseEnter={() => setActive(index)} onClick={() => choose(match)}><span className="property-option-name">{match.name}</span>{specimenNode(match)}</div>
          })}
        </div>)}
      </div>
      {matches.length === 0 && <div id={`${listId}-notes`} className="property-option-notes">
        {/* THE EMPTY STATE NAMES THE TWO PLACES IT LOOKED, because those are
            the two groups above it. STORY 16.9 dropped the third clause,
            "or in the list you can install" — this control no longer looks
            there at all, and `Add fonts…` below is where that search lives
            now. A control may not claim to have searched a place it never
            queried. */}
        {matches.length === 0 && <p className="property-option property-option-empty">{`Nothing in this template or on this machine matches "${query.trim()}".`}</p>}
      </div>}
      {/* STORY 16.3 — THE DOOR TO THE BROWSER, WHERE THE DESIGN PUTS IT: the last
          row of the open family dropdown, INSIDE the floating panel (Story 16.6 follow-on).
          It used to be a sibling of that panel, and `.property-combobox` is the
          positioning context, so `top: 100%` put the panel below the button and the
          design's last row rendered FIRST. Measured in Chromium before the move:
          button at y=433, list at y=481. The scroll now belongs to the listbox, so
          this row stays pinned at the foot of the panel instead of scrolling away. It is a real button OUTSIDE the
          `role="listbox"` rather than a fourth `role="presentation"` child inside
          it, because a listbox's children are options and this is not one — the
          keyboard walk over `matches` above must not land on it.

          NO KEYBOARD SHORTCUT AND NO HINT GLYPH, AND THE OMISSION IS RULED RATHER
          THAN FORGOTTEN (D-16.R.33 R2, owner-confirmed). The mockup prints `⌘G`
          beside this row; `⌘G` is the browser's own Find Next, and this
          application's convention puts conventional document actions on Command
          (⌘S, ⌘Z) and app-specific ones on Option (⌥P, ⌥S). `⌥F` is named as the
          eventual shape and is not bound in this epic — so no glyph is drawn,
          because a `⌘G` label beside a key that does nothing is a false UI
          string. `src/shortcuts.ts` is untouched. */}
      {/* THE ACCESSIBLE NAME CARRIES BOTH LINES. An `aria-label` REPLACES the
          element's contents for assistive technology, so naming this "Add fonts…"
          alone deleted the only sentence saying what the row does — the sub-label
          is visible to a sighted reader and was inaudible to everybody else. */}
      <button type="button" className="property-add-fonts" aria-label="Add fonts… Browse and embed web fonts" disabled={pending || pickBusy} onMouseDown={(event) => event.preventDefault()} onClick={() => { close(); onOpenFontBrowser() }}>
        <span className="property-add-fonts-label">Add fonts…</span>
        <span className="property-add-fonts-note">Browse and embed web fonts</span>
        </button>
    </div>}
    {pickError && <p role="alert" className="property-error">{pickError.message}</p>}
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
// TWO POPULATIONS, TWO SEAMS, ONE EXPRESSION (Story 8.4e). A fragment carries
// exactly one of the engine's two identities for the face it was measured
// with, and each has its own derivation module and nothing else derives it:
//
//	assetKey -> embedded-face-family.ts  (the face the DOCUMENT carries)
//	face     -> shipped-face-family.ts   (the face the BUILD ships)
//
// THE CARRIED BRANCH is set only when the engine attributed this fragment to
// an asset the document carries *and* that asset's face has actually reached
// the page's font set. Both halves matter. Without the first, a face would be
// asked for under a family nothing declares; without the second, a fetch that
// failed would take the fragment OFF the stylesheet's declared stack — an
// inline declaration replaces the rule rather than extending it — and the
// canvas would paint with whatever the browser defaults to.
//
// THE SHIPPED BRANCH needs no such registration check: Story 8.4b declares an
// `@font-face` for each of the engine's own face names at build time, over the
// engine's own bytes, so the name the engine measured with is already
// resolvable. It carries the declared stack as its own tail, so a codepoint
// the attributed face does not cover still reaches the other shipped faces
// rather than the browser's default. Until this story it was set to NOTHING,
// and the fragment fell to one fixed stylesheet stack whatever order the
// document declared — which for a chain like ["Noto Sans Thai"] rasterized
// Latin with Noto Sans while the engine had measured it with Noto Sans Thai.
//
// WITH NEITHER identity the fragment falls to `.canvas-text-fragment`'s
// declared stack in App.css, which is the degrade path and the only path left
// for an unattributed fragment.
//
// Each family is derived FROM THE ENGINE'S OWN IDENTITY (D-8.4.1, D-8.4.14) by
// the one module that makes that decision. Nothing here reads a chain entry's
// `family`, `style` or a chain name, and the stylesheet still holds no
// document input at all.
//
// TextPaint is EXPORTED so the per-fragment face attribution can be asserted
// on a real DOM node rather than by scanning this file's text: what the canvas
// asks for is a rendered fact, and canvas-font-stack.test.ts reads it off the
// element.
export function TextPaint({ component, carriedFaces, zoom }: { component: CanvasProjection['components'][number]; carriedFaces: ReadonlySet<string>; zoom: number }) {
  const paint = component.textPaint!
  return <span className="canvas-text-paint" aria-hidden="true" style={{ '--text-font-size': canvasDisplay.css(component.fontSize ?? 12000, zoom), '--text-font-weight': component.bold ? 700 : 400, '--text-font-style': component.italic ? 'italic' : 'normal', ...(component.color === undefined ? {} : { '--text-ink': component.color }) } as CSSProperties}>{paint.lines.map((line, lineIndex) => <span className="canvas-text-line" key={`${component.id}-${lineIndex}`} style={{ '--text-line-baseline': canvasDisplay.css(line.baseline - component.y, zoom), '--text-line-advance': canvasDisplay.css(line.advance, zoom) } as CSSProperties}>{line.fragments.map((fragment, fragmentIndex) => <span className="canvas-text-fragment" key={`${component.id}-${lineIndex}-${fragmentIndex}`} style={{ '--text-fragment-x': canvasDisplay.css(fragment.x - component.x, zoom), ...(fragment.assetKey !== undefined && carriedFaces.has(fragment.assetKey) ? { fontFamily: embeddedFaceFamily(fragment.assetKey) } : isShippedFaceName(fragment.face) ? { fontFamily: shippedFaceFamily(fragment.face) } : {}) } as CSSProperties}>{textRuns(fragment.text).map((part, partIndex) => isExpressionRun(part) ? <span className="canvas-text-expression" key={`${component.id}-${lineIndex}-${fragmentIndex}-${partIndex}`}>{part}</span> : part)}</span>)}</span>)}</span>
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
