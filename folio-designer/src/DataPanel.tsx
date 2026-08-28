import { useMemo, useState, type CSSProperties, type KeyboardEvent } from 'react'
import type { SampleData, SampleNode } from './sample-data'

type VisibleNode = Readonly<{ node: SampleNode; level: number; key: string; parent?: string }>

// This panel retains only local discovery/selection interaction. It never
// turns a display path into Folio syntax or decides whether a component/path
// pairing is legal; the one opaque command does that in Go.
export type BindingErrorScope = Readonly<{ sample: SampleData; componentID: string; segments: ReadonlyArray<string>; message: string }>

export function DataPanel({ sample, error, busy, available, selectedComponentId, selectedBinding, bindingError, bindingBusy, onLoad, onConnect }: Readonly<{ sample?: SampleData; error?: string; busy: boolean; available: boolean; selectedComponentId?: string; selectedBinding?: string; bindingError?: BindingErrorScope; bindingBusy?: boolean; onLoad: () => void; onConnect?: (segments: ReadonlyArray<string>) => void }>) {
  const action = sample ? 'Replace sample JSON' : 'Load sample JSON'
  const [pickedState, setPickedState] = useState<Readonly<{ sample: SampleData; node: SampleNode }>>()
  const picked = pickedState && pickedState.sample === sample ? pickedState.node : undefined
  const candidate = picked?.segments
  const currentBindingError = bindingError && sample === bindingError.sample && selectedComponentId === bindingError.componentID && candidate && sameSegments(candidate, bindingError.segments) ? bindingError.message : undefined
  const unavailable = !sample
    ? 'Binding unavailable: no sample data loaded.'
    : !picked
      ? 'Choose an offered root scalar path.'
      : !candidate
        ? 'The selected tree item is not a root scalar candidate.'
        : !selectedComponentId
          ? 'Binding unavailable: select one component first.'
          : undefined
  return <aside className="data-panel" aria-label="Data panel">
    <p className="section-label">DATA</p>
    <button className="file-button" type="button" onClick={onLoad} disabled={busy || !available}>{action}</button>
    {error && <p role="alert" className="data-message">{error}</p>}
    {!sample ? <><p className="data-empty" role="status">{unavailable}</p><p className="honest-note">{available ? 'Load one local JSON document to inspect its paths.' : 'Local sample selection is unavailable in this shell.'}</p></> : <>
      <p className="data-file" role="status">Local sample: <code>{sample.name}</code></p>
      <p className="honest-note">Choose an offered root scalar path, then connect it to the selected component. Go validates the command.</p>
      {sample.truncated && <p className="data-message" role="status">Tree inspection is truncated to keep this local panel responsive.</p>}
      <DataTree key={treeIdentity(sample.tree)} root={sample.tree} picked={picked} onPick={(node) => setPickedState({ sample, node })} />
      <p className="binding-status" role="status">{unavailable ?? (selectedBinding ? <>Current engine binding: <code>{selectedBinding}</code></> : <>Ready to ask the engine to bind <code>{picked!.path}</code>.</>)}</p>
      <button className="file-button binding-connect" type="button" disabled={Boolean(unavailable) || bindingBusy || !onConnect} onClick={() => candidate && onConnect?.(candidate)}>{bindingBusy ? 'Connecting…' : 'Connect selected path'}</button>
      {currentBindingError && <p className="data-message" role="alert">{currentBindingError}</p>}
    </>}
  </aside>
}

function DataTree({ root, picked, onPick }: Readonly<{ root: SampleNode; picked?: SampleNode; onPick: (node: SampleNode) => void }>) {
  const [expanded, setExpanded] = useState<ReadonlySet<string>>(() => new Set([keyFor(root, 0)]))
  const visible = useMemo(() => flatten(root, expanded), [root, expanded])
  const [active, setActive] = useState(() => keyFor(root, 0))
  const focus = (key: string) => { setActive(key); requestAnimationFrame(() => Array.from(document.querySelectorAll<HTMLElement>('[data-tree-key]')).find((element) => element.dataset.treeKey === key)?.focus()) }
  const toggle = (key: string) => setExpanded((value) => { const next = new Set(value); if (next.has(key)) next.delete(key); else next.add(key); return next })
  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, current: VisibleNode) => {
    // Tree navigation is local interaction. Never let an arrow intended for a
    // focused discovery node become a canvas nudge shortcut.
    event.stopPropagation()
    const index = visible.findIndex((entry) => entry.key === current.key)
    const move = (next: number) => { event.preventDefault(); if (visible[next]) focus(visible[next]!.key) }
    const branch = current.node.children.length > 0
    if (event.key === 'ArrowDown') return move(index + 1)
    if (event.key === 'ArrowUp') return move(index - 1)
    if (event.key === 'Home') return move(0)
    if (event.key === 'End') return move(visible.length - 1)
    if (event.key === 'ArrowRight' && branch) { event.preventDefault(); if (!expanded.has(current.key)) setExpanded((value) => new Set(value).add(current.key)); else if (visible[index + 1]?.parent === current.key) focus(visible[index + 1]!.key); return }
    if (event.key === 'ArrowLeft') { event.preventDefault(); if (branch && expanded.has(current.key)) setExpanded((value) => { const next = new Set(value); next.delete(current.key); return next }); else if (current.parent) focus(current.parent); return }
    if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); if (branch) toggle(current.key); else if (current.node.segments) onPick(current.node) }
  }
  return <ul className="data-tree" role="tree" aria-label="Sample data paths">{visible.map((entry) => <li className={`data-tree-node${picked === entry.node ? ' data-tree-picked' : ''}`} role="none" key={entry.key} style={{ '--tree-level': entry.level } as CSSProperties}><button className="tree-item" type="button" role="treeitem" data-tree-key={entry.key} aria-level={entry.level} aria-expanded={entry.node.children.length ? expanded.has(entry.key) : undefined} aria-selected={entry.node.segments ? picked === entry.node : undefined} tabIndex={active === entry.key ? 0 : -1} onFocus={() => setActive(entry.key)} onClick={() => entry.node.children.length ? toggle(entry.key) : entry.node.segments && onPick(entry.node)} onKeyDown={(event) => onKeyDown(event, entry)}><span className="tree-label">{entry.node.label}</span><code>{entry.node.path}</code><span className="tree-detail">{describe(entry.node)}{entry.node.segments && ' · root scalar candidate'}</span></button></li>)}</ul>
}

function flatten(root: SampleNode, expanded: ReadonlySet<string>): VisibleNode[] {
  const result: VisibleNode[] = []
  const visit = (node: SampleNode, level: number, parent?: string, ordinal = 0) => {
    const key = parent ? `${parent}/${keyFor(node, ordinal)}` : keyFor(node, ordinal); result.push({ node, level, key, parent })
    if (node.children.length && expanded.has(key)) node.children.forEach((child, index) => visit(child, level + 1, key, index))
  }
  visit(root, 1); return result
}

const keyFor = (node: SampleNode, ordinal: number): string => `${node.path}:${ordinal}`
const describe = (node: SampleNode): string => [node.kind, node.count === undefined ? undefined : `${node.count} ${node.kind === 'collection' ? 'items' : 'properties'}`, node.preview, node.truncated ? 'truncated' : undefined].filter(Boolean).join(' · ')
const sameSegments = (left: ReadonlyArray<string>, right: ReadonlyArray<string>) => left.length === right.length && left.every((segment, index) => segment === right[index])
// A changed visible path set remounts this transient discovery widget, so its
// initial root remains the one roving tab stop after sample replacement.
const treeIdentity = (node: SampleNode): string => `${node.path}:${node.children.map(treeIdentity).join('|')}`
