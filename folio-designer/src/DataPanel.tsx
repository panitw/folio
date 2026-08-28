import { useMemo, useState, type CSSProperties, type KeyboardEvent } from 'react'
import type { SampleData, SampleNode } from './sample-data'

type VisibleNode = Readonly<{ node: SampleNode; level: number; key: string; parent?: string }>

export function DataPanel({ sample, error, busy, available, onLoad }: Readonly<{ sample?: SampleData; error?: string; busy: boolean; available: boolean; onLoad: () => void }>) {
  const action = sample ? 'Replace sample JSON' : 'Load sample JSON'
  return <aside className="data-panel" aria-label="Data panel">
    <p className="section-label">DATA</p>
    <button className="file-button" type="button" onClick={onLoad} disabled={busy || !available}>{action}</button>
    {error && <p role="alert" className="data-message">{error}</p>}
    {!sample ? <><p className="data-empty" role="status">Binding unavailable: no sample data loaded.</p><p className="honest-note">{available ? 'Load one local JSON document to inspect its paths. Binding actions are unavailable in this story.' : 'Local sample selection is unavailable in this shell.'}</p></> : <><p className="data-file" role="status">Local sample: <code>{sample.name}</code></p><p className="honest-note">Discovery only. Display paths are not connected to a component.</p>{sample.truncated && <p className="data-message" role="status">Tree inspection is truncated to keep this local panel responsive.</p>}<DataTree root={sample.tree} /></>}
  </aside>
}

function DataTree({ root }: Readonly<{ root: SampleNode }>) {
  const [expanded, setExpanded] = useState<ReadonlySet<string>>(() => new Set([keyFor(root, 0)]))
  const visible = useMemo(() => flatten(root, expanded), [root, expanded])
  const [active, setActive] = useState(() => keyFor(root, 0))
  const focus = (key: string) => { setActive(key); requestAnimationFrame(() => Array.from(document.querySelectorAll<HTMLElement>('[data-tree-key]')).find((element) => element.dataset.treeKey === key)?.focus()) }
  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, current: VisibleNode) => {
    const index = visible.findIndex((entry) => entry.key === current.key)
    const move = (next: number) => { event.preventDefault(); if (visible[next]) focus(visible[next]!.key) }
    const branch = current.node.children.length > 0
    if (event.key === 'ArrowDown') return move(index + 1)
    if (event.key === 'ArrowUp') return move(index - 1)
    if (event.key === 'Home') return move(0)
    if (event.key === 'End') return move(visible.length - 1)
    if (event.key === 'ArrowRight' && branch) { event.preventDefault(); if (!expanded.has(current.key)) setExpanded((value) => new Set(value).add(current.key)); else if (visible[index + 1]?.parent === current.key) focus(visible[index + 1]!.key); return }
    if (event.key === 'ArrowLeft') { event.preventDefault(); if (branch && expanded.has(current.key)) setExpanded((value) => { const next = new Set(value); next.delete(current.key); return next }); else if (current.parent) focus(current.parent); return }
    if ((event.key === 'Enter' || event.key === ' ') && branch) { event.preventDefault(); setExpanded((value) => { const next = new Set(value); if (next.has(current.key)) next.delete(current.key); else next.add(current.key); return next }) }
  }
  return <ul className="data-tree" role="tree" aria-label="Sample data paths">{visible.map((entry) => <li className="data-tree-node" role="none" key={entry.key} style={{ '--tree-level': entry.level } as CSSProperties}><button className="tree-item" type="button" role="treeitem" data-tree-key={entry.key} aria-level={entry.level} aria-expanded={entry.node.children.length ? expanded.has(entry.key) : undefined} tabIndex={active === entry.key ? 0 : -1} onFocus={() => setActive(entry.key)} onClick={() => entry.node.children.length && setExpanded((value) => { const next = new Set(value); if (next.has(entry.key)) next.delete(entry.key); else next.add(entry.key); return next })} onKeyDown={(event) => onKeyDown(event, entry)}><span className="tree-label">{entry.node.label}</span><code>{entry.node.path}</code><span className="tree-detail">{describe(entry.node)}</span></button></li>)}</ul>
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
