import { useMemo, useState, type CSSProperties, type KeyboardEvent } from 'react'
import { SCALAR_BINDING_COMPONENT_TYPES, type CanvasComponentType } from './engine-protocol'
import type { SampleData, SampleNode } from './sample-data'

type VisibleNode = Readonly<{ node: SampleNode; level: number; key: string; parent?: string }>

// This panel retains only local discovery/selection interaction. It never
// turns a display path into Folio syntax, and it never decides whether a
// PATH is bindable; the one opaque command does that in Go.
//
// ⚠ AMENDED BY STORY 14.4, AND THE AMENDMENT IS THE POINT. This comment used
// to end "…or decides whether a component/path pairing is legal", which was
// already partly false when it was written — the third `unavailable` arm below
// ("not a root scalar candidate") is a legality judgement — and is now false in
// one further, deliberate way: the FIFTH arm judges the selected component's
// KIND against `SCALAR_BINDING_COMPONENT_TYPES`, before the command is sent.
//
// What the panel still never judges is the PATH. D-6.2.1 keeps sample runtime
// kind out of command legality on purpose: Go accepts `{{items}}` for an empty
// collection and reports the mismatch at render, so a panel that pre-judged
// path kind would be a second, drifting copy of the binder and would be wrong
// the moment the runtime data differed from the sample.
export type BindingErrorScope = Readonly<{ sample: SampleData; componentID: string; segments: ReadonlyArray<string>; message: string }>

// The noun phrase the fifth arm names the selected kind with, ARTICLE INCLUDED
// — because "a"/"an" is a property of the WORD, and a template that supplied
// one article for every kind rendered "a image".
//
// `rect` is spelled out as "a rectangle": the projection's own vocabulary is
// the honest one for every other kind, but `rect` is an abbreviation rather
// than a word. NOTE this does NOT match the component-identity row in the
// inspector, which renders `single.type` verbatim and therefore still says
// `rect` — an earlier version of this comment claimed the two agreed, and they
// do not. Story 14.2 spells the same kind "Rectangle" in the property
// vocabulary, so the inspector's raw row is the odd one out, not this.
//
// `text` is UNREACHABLE here — the arm fires only for kinds absent from
// SCALAR_BINDING_COMPONENT_TYPES, and `text` is the sole member. It is kept
// because the Record is total, so a kind added to the projection cannot reach
// this map without a word.
const kindWord: Readonly<Record<CanvasComponentType, string>> = { text: 'a text component', image: 'an image', table: 'a table', line: 'a line', rect: 'a rectangle' }

export function DataPanel({ sample, error, busy, available, selectedComponentId, selectedComponentType, selectedBinding, bindingError, bindingBusy, onLoad, onConnect }: Readonly<{ sample?: SampleData; error?: string; busy: boolean; available: boolean; selectedComponentId?: string; selectedComponentType?: CanvasComponentType; selectedBinding?: string; bindingError?: BindingErrorScope; bindingBusy?: boolean; onLoad: () => void; onConnect?: (segments: ReadonlyArray<string>) => void }>) {
  const action = sample ? 'Replace sample JSON' : 'Load sample JSON'
  const [pickedState, setPickedState] = useState<Readonly<{ sample: SampleData; node: SampleNode }>>()
  const picked = pickedState && pickedState.sample === sample ? pickedState.node : undefined
  const candidate = picked?.segments
  const currentBindingError = bindingError && sample === bindingError.sample && selectedComponentId === bindingError.componentID && candidate && sameSegments(candidate, bindingError.segments) ? bindingError.message : undefined
  // STORY 14.4 / AC2 — THE FIFTH ARM. Before it, this ladder had four arms and
  // none inspected the selected component's KIND, so the inspector invited a
  // pick for a Line, a Rectangle, an Image or a Table and Go's refusal — *"only
  // text components can receive a scalar binding"* — arrived only after the
  // command round-tripped. The reason is now stated BEFORE the attempt, and
  // because `unavailable` also disables the Connect button, no command is
  // dispatched for the gesture at all.
  //
  // It sits AFTER the `!selectedComponentId` arm on purpose: with nothing
  // selected there is no kind to speak for, and "select one component first"
  // is still the true first thing to say.
  //
  // ⚠ D-14.4.Q3's INHERITANCE NOTE, ADDRESSED TO STORY 14.6's BUILDER. This arm
  // and its wording are PRESENTATION, and 14.6 is expected to replace them with
  // its own row treatment, badges and dimming. What 14.6 must inherit and must
  // NOT re-derive is THE RULE: `SCALAR_BINDING_COMPONENT_TYPES` and its mirror
  // `describe` in `engine-bounds-mirror.test.ts`. Reading this arm as territory
  // to preserve is a misreading; re-spelling the constant as an inline
  // comparison against the kind literal reds the mirror's `sites` assertion,
  // which is exactly what it is there for. (That mirror's census also asserts
  // this FILE holds no such comparison, so the drift cannot hide in a comment
  // either — which is why this sentence does not quote one.)
  //
  // ⚠ IT FAILS CLOSED ON AN UNKNOWN KIND, AND THAT IS THE WHOLE OF P1. The
  // first spelling of this arm was `selectedComponentType !== undefined && !…
  // .includes(…)`, which SKIPPED the arm when the type was absent and handed
  // Connect back — re-enabling the round-trip refusal this story exists to
  // remove. The state is reachable: the id is passed from the selection
  // unconditionally, while the kind comes from the projection, so an absent
  // canvas or an id no longer in the projection yields id-present /
  // kind-absent. A panel that does not know the kind cannot know the engine
  // will accept it, so it says so rather than inviting the attempt.
  const bindableKind = selectedComponentType !== undefined && SCALAR_BINDING_COMPONENT_TYPES.includes(selectedComponentType)
  const selectedKindReason = selectedComponentType === undefined ? 'The selected component is not in the current projection.' : `The selected component is ${kindWord[selectedComponentType]}.`
  const unavailable = !sample
    ? 'Binding unavailable: no sample data loaded.'
    : !picked
      ? 'Choose an offered root scalar path.'
      : !candidate
        ? 'The selected tree item is not a root scalar candidate.'
        : !selectedComponentId
          ? 'Binding unavailable: select one component first.'
          : !bindableKind
            ? `Only text components can receive a scalar binding. ${selectedKindReason}`
            : undefined
  return <div className="data-panel" aria-label="Data panel">
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
  </div>
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
