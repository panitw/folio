import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import type { EngineClient } from './engine-client'
import type { CanvasProjection } from './engine-protocol'
import { moveComponentsCommand } from './component-command'
import { enclosedComponents, selectionRectangle, type SelectionPoint, type SelectionRectangle } from './canvas-selection'

type PointerInput = Readonly<{ pointerId: number; clientX: number; clientY: number }>
type GestureBase = { pointerId: number; clientX: number; clientY: number; canvas: CanvasProjection; revision: number; generation: number; zoom: number; selectionKey: string; ids: ReadonlyArray<string>; changed: boolean }
type RectangleGesture = GestureBase & { kind: 'rectangle'; start: SelectionPoint; end: SelectionPoint; additive: boolean; clearOnClick: boolean }
type GroupGesture = GestureBase & { kind: 'group'; referenceId: string; snap: boolean; dx: number; dy: number; sequence: number; acceptedSequence: number; acceptedDX: number; acceptedDY: number; inFlight: boolean; released: boolean; committing: boolean }
type Gesture = RectangleGesture | GroupGesture
export type GroupPreview = Readonly<{ ids: ReadonlyArray<string>; dx: number; dy: number }>
type Context = Readonly<{ engine?: EngineClient; canvas?: CanvasProjection; revision: number; generation: number; zoom: number; selection: ReadonlyArray<string>; enabled: boolean; snap: boolean; documentDelta: (pixels: number, zoom: number) => number; capture: (id: number) => void; release: (id: number) => void; onSelection: (ids: ReadonlyArray<string>) => void; onCommit: (payload: ArrayBuffer) => Promise<unknown>; onError: (error: unknown) => void }>

// One stable ancestor owns pointer capture. One preview query may be in flight;
// newer pointer positions replace queued work, and release waits for the final
// requested position. No browser geometry or snapping policy lives here.
export function useCanvasSelection(context: Context) {
  const current = useRef(context)
  useLayoutEffect(() => { current.current = context })
  const gesture = useRef<Gesture | undefined>(undefined)
  const [rectangle, setRectangle] = useState<SelectionRectangle>()
  const [group, setGroup] = useState<GroupPreview>()
  const suppressClick = useRef(false)
  const commitPending = useRef(false)
  const clear = () => {
    const old = gesture.current
    gesture.current = undefined
    setRectangle(undefined); setGroup(undefined)
    if (old) current.current.release(old.pointerId)
  }
  const cancel = () => {
    if (!gesture.current) return false
    suppressClick.current = true
    // A dispatched transaction cannot be cancelled by a view gesture. Keep
    // its accepted geometry until it settles; context replacement is handled
    // separately below and never installs a response over a newer document.
    if (gesture.current.kind === 'group' && gesture.current.committing) return true
    clear(); return true
  }
  const valid = (operation: Gesture) => {
    const now = current.current
    return gesture.current === operation && now.enabled && now.canvas === operation.canvas && now.revision === operation.revision && now.generation === operation.generation && now.zoom === operation.zoom && now.selection.join(',') === operation.selectionKey
  }
  useEffect(() => {
    const operation = gesture.current
    if (operation && !valid(operation)) { suppressClick.current = true; clear() }
  })
  useEffect(() => () => { gesture.current = undefined }, [])
  useEffect(() => {
    const scroll = () => { cancel() }
    const escape = (event: KeyboardEvent) => { if (event.key === 'Escape' && cancel()) { event.preventDefault(); event.stopPropagation() } }
    window.addEventListener('scroll', scroll, true)
    window.addEventListener('keydown', escape, true)
    return () => { window.removeEventListener('scroll', scroll, true); window.removeEventListener('keydown', escape, true) }
  }, [])
  const base = (input: PointerInput, ids: ReadonlyArray<string>): GestureBase | undefined => {
    const now = current.current
    if (!now.enabled || !now.canvas || commitPending.current) return undefined
    cancel(); suppressClick.current = false
    return { pointerId: input.pointerId, clientX: input.clientX, clientY: input.clientY, canvas: now.canvas, revision: now.revision, generation: now.generation, zoom: now.zoom, selectionKey: ids.join(','), ids: [...ids], changed: false }
  }
  const beginRectangle = (input: PointerInput, start: SelectionPoint, additive: boolean, clearOnClick = true) => {
    const captured = base(input, current.current.selection)
    if (!captured) return
    gesture.current = { ...captured, kind: 'rectangle', start, end: start, additive, clearOnClick }
    current.current.capture(input.pointerId)
  }
  const beginGroup = (input: PointerInput, ids: ReadonlyArray<string>, referenceId: string) => {
    const captured = base(input, ids)
    if (!captured) return
    gesture.current = { ...captured, kind: 'group', referenceId, snap: current.current.snap, dx: 0, dy: 0, sequence: 0, acceptedSequence: 0, acceptedDX: 0, acceptedDY: 0, inFlight: false, released: false, committing: false }
    current.current.capture(input.pointerId)
  }
  const finishGroup = (operation: GroupGesture) => {
    if (!valid(operation) || operation.inFlight || operation.committing || operation.acceptedSequence !== operation.sequence) return
    const payload = moveComponentsCommand(operation.ids, operation.referenceId, operation.dx, operation.dy, operation.snap, operation.revision, true)
    const commit = operation.changed && (operation.acceptedDX !== 0 || operation.acceptedDY !== 0)
    if (!commit) { clear(); return }
    operation.committing = true
    commitPending.current = true
    // Keep accepted geometry until the authoritative command settles. A new
    // gesture cannot start from the old snapshot during that round trip.
    void current.current.onCommit(payload).catch((error: unknown) => {
      if (valid(operation)) current.current.onError(error)
    }).finally(() => { commitPending.current = false; if (gesture.current === operation) clear() })
  }
  const pump = async (operation: GroupGesture): Promise<void> => {
    if (!valid(operation) || operation.inFlight || operation.committing) return
    const engine = current.current.engine
    if (!engine) { cancel(); return }
    if (operation.acceptedSequence === operation.sequence) { if (operation.released) finishGroup(operation); return }
    operation.inFlight = true
    const sequence = operation.sequence
    try {
      const result = await engine.request('group-move-preview', moveComponentsCommand(operation.ids, operation.referenceId, operation.dx, operation.dy, operation.snap, operation.revision, true))
      operation.inFlight = false
      if (!valid(operation)) return
      if (!result.groupMove || result.groupMove.revision !== operation.revision) { cancel(); return }
      if (sequence > operation.acceptedSequence) {
        operation.acceptedSequence = sequence
        operation.acceptedDX = result.groupMove.dx; operation.acceptedDY = result.groupMove.dy
        setGroup({ ids: operation.ids, dx: operation.acceptedDX, dy: operation.acceptedDY })
      }
      if (operation.acceptedSequence !== operation.sequence) void pump(operation)
      else if (operation.released) finishGroup(operation)
    } catch (error) {
      operation.inFlight = false
      if (!valid(operation)) return
      cancel(); current.current.onError(error)
    }
  }
  const move = (input: PointerInput) => {
    const operation = gesture.current
    if (!operation || operation.pointerId !== input.pointerId || operation.kind === 'group' && operation.released) return
    if (!valid(operation)) { cancel(); return }
    const rawDX = input.clientX - operation.clientX, rawDY = input.clientY - operation.clientY
    operation.changed ||= Math.abs(rawDX) >= 2 || Math.abs(rawDY) >= 2
    if (!operation.changed) return
    const dx = current.current.documentDelta(rawDX, operation.zoom) * 1000, travel = current.current.documentDelta(rawDY, operation.zoom) * 1000
    if (operation.kind === 'rectangle') {
      operation.end = { x: operation.start.x + dx, y: operation.start.y + travel }
      setRectangle(selectionRectangle(operation.start, operation.end))
    } else {
      const dy = travel
      if (dx !== operation.dx || dy !== operation.dy) { operation.dx = dx; operation.dy = dy; operation.sequence++ }
      // Show zero while the first engine query is pending; geometry controls
      // must be unavailable as soon as movement owns the selection.
      setGroup({ ids: operation.ids, dx: operation.acceptedDX, dy: operation.acceptedDY })
      void pump(operation)
    }
  }
  const finish = (input: PointerInput) => {
    const operation = gesture.current
    if (!operation || operation.pointerId !== input.pointerId) return
    move(input)
    if (!valid(operation)) return
    suppressClick.current = true
    if (operation.kind === 'rectangle') {
      const found = operation.changed ? enclosedComponents(operation.canvas, operation.zoom, selectionRectangle(operation.start, operation.end)) : []
      const wanted = new Set(operation.additive ? [...operation.ids, ...found] : found)
      clear()
      if (operation.changed || operation.clearOnClick && !operation.additive) current.current.onSelection(operation.canvas.components.filter((component) => wanted.has(component.id)).map((component) => component.id))
    } else { operation.released = true; void pump(operation) }
  }
  const lostCapture = () => { if (gesture.current?.kind === 'group' && gesture.current.released) return; cancel() }
  const freshPointer = () => { suppressClick.current = false }
  const consumeClick = () => { const suppressed = suppressClick.current; suppressClick.current = false; return suppressed }
  const blocksPointer = () => commitPending.current
  return { rectangle, group, beginRectangle, beginGroup, move, finish, cancel, lostCapture, freshPointer, consumeClick, blocksPointer }
}
