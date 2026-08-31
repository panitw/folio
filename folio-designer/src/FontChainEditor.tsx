import { useId, useLayoutEffect, useRef, type KeyboardEvent } from 'react'
import type { CanvasProjection } from './engine-protocol'
import { addFontChainCommand, addFontChainEntryCommand, deleteFontChainCommand, moveFontChainEntryCommand, removeFontChainEntryCommand, renameFontChainCommand } from './font-chain-command'
import { sameFontChainControl, type FontChainCommitError, type FontChainControl } from './font-chain-control'

// THE CHAIN EDITOR HOLDS NO MODEL (AD-15). Every name, every entry and every
// ordering drawn here is read from `chains` — CanvasProjection.fontChains on
// the latest snapshot — on every render. There is no local list, no optimistic
// mutation and no cache that survives a snapshot: an edit dispatches one
// opaque command (AD-16) and what the author sees afterwards is the engine's
// answer, not what they typed.
//
// AND IT HOLDS NO RULE ABOUT THE DOCUMENT. It never checks for a duplicate
// name, never walks the document for elements referencing a chain, never
// checks whether a remove would empty a chain, and never checks whether this
// build ships a face. Every one of those is refused by the engine, in the
// engine's own sentence, which this component only places.
// canvas-authority-contract.test.ts keeps the engine's refusal vocabulary out
// of this file by scanning for it — that is the one rule enforced
// mechanically, and it is narrower than the paragraph above.
//
// "NO RULE" IS ABOUT AUTHORITY, NOT ABOUT CODE. `entryLabel` below does branch
// on `assetKey` and does compose `family` and `style` into one string, and
// saying otherwise would be a comment that reads its own code wrong (review
// finding 5). What it never does is DECIDE anything: the discriminant it
// branches on is projected, the strings it composes are projected, and the
// engine has already resolved every case where a value could be missing —
// which is why a family is never empty here and never needs a fallback
// invented in the browser.
//
// It is INLINE, in the TYPOGRAPHY section it is revealed from: no dialog, no
// separate mode, no focus trap (AC1). TableEditor is the ordered
// add/remove/reorder precedent and its keyboard idioms are copied — 1-indexed
// positional accessible names, per-entry move-earlier/move-later buttons — but
// deliberately NOT its `role="dialog" aria-modal="true"` shell.
//
// THE DISPLAYED TEXT OF AN ENTRY IS THE PROJECTED ENTRY, UNMODIFIED. No
// parsing, no key detection, no extension stripping, no file-name handling.
//
// Story 8.3 gave an entry the richer shape this note was written for, and the
// promise held: an entry is now an object carrying a `face`, an `assetKey`, a
// `family` and a `style`, and this component still decides nothing about them.
// Which KIND of entry it is comes from the projection (`assetKey` non-empty),
// never from
// inspecting a value — a 64-character face name is a legal face name, so
// "looks like a digest" was never a test this side could have made. What an
// embedded entry DISPLAYS is `family` and `style`, both read by Go from the
// asset's own `font` record; when the document declares no family Go projects
// the asset key as one, so the browser never has to decide what to draw for a
// nameless face. `entryLabel` below is the whole of it.
type Props = Readonly<{
  chains: CanvasProjection['fontChains']
  busy: boolean
  error?: FontChainCommitError
  onCommand: (payload: ArrayBuffer, control: FontChainControl) => void
}>

export function FontChainEditor({ chains, busy, error, onCommand }: Props) {
  const uid = useId()
  // The text fields are UNCONTROLLED and are read by id at commit time rather
  // than mirrored into state. A draft the panel held would have to be
  // invalidated on every snapshot to keep AD-15's promise, and the field the
  // author is typing into is the one place a stale copy would be visible.
  const typed = (id: string): string => (document.getElementById(id) as HTMLInputElement | null)?.value ?? ''
  // WHERE FOCUS GOES WHEN A COMMAND SETTLES, and it settles two ways.
  //
  // Pressing any control here sets `busy`, which sets `disabled`, which BLURS
  // it. So focus has to be put back deliberately, or the author is left on
  // `document.body` with the refusal they just caused unreachable by Tab and
  // an `aria-errormessage` hanging off a control they are no longer on.
  //
  // ACCEPTED: the list moved, and a reordered entry now sits under a DIFFERENT
  // button than the one that was pressed — so keyboard-only reordering follows
  // the entry. At the ends of the list the control it lands under is DISABLED
  // (an entry moved to last has no "move later"), so the row's counterpart is
  // the fallback, and the originating control the fallback after that.
  // REFUSED: the list did not move, so focus goes straight back to the control
  // that asked — the one the refusal is anchored to.
  //
  // The request is cleared as soon as it SETTLES, not when the list next
  // changes. Clearing it on `[chains]` alone left a refused command's targets
  // alive indefinitely, so the next unrelated re-projection — an ordinary
  // property commit, an undo, a file load — stole focus to a move button the
  // author never pressed.
  //
  // "The list moved" is decided on the projected VALUES, not on the array's
  // identity. Referential comparison looked equivalent and is not: Go is free
  // to hand back a chain list that is equal to the one already installed, and
  // a reference check would then call an accepted no-op edit a move. Reading
  // the names and entries says what is actually being asked — did what the
  // author is looking at change? — and it is the same answer in a test, where
  // a fixture may legitimately reuse one array across two snapshots.
  //
  // Story 8.3: the entries are OBJECTS, so `entries.join()` here would produce
  // `[object Object]` for every one of them — a signature under which two
  // chains that differ only in an entry's family, style or asset key compare
  // EQUAL, so a landed edit would be read as "the list did not move" and focus
  // would go to the wrong control. The signature is built field by field
  // instead, over every value the panel can display or act on.
  const listing = chains.map((chain) => [chain.name, ...chain.entries.map((entry) => [entry.face, entry.assetKey, entry.family, entry.style].join('\u0002'))].join('\u0000')).join('\u0001')
  type Request = Readonly<{ origin: string; targets: ReadonlyArray<string>; clear: ReadonlyArray<string>; listing: string }>
  const pending = useRef<Request | undefined>(undefined)
  useLayoutEffect(() => {
    const request = pending.current
    if (!request || busy) return
    pending.current = undefined
    const landed = request.listing !== listing
    // A field is emptied only when its add ACTUALLY LANDED. Leaving the text
    // in place after an accepted add invites a second press, and the engine
    // has no duplicate-entry rule to refuse it with — it would simply append
    // the same face twice.
    if (landed) for (const id of request.clear) { const field = document.getElementById(id) as HTMLInputElement | null; if (field) field.value = '' }
    for (const id of landed ? [...request.targets, request.origin] : [request.origin]) {
      const node = document.getElementById(id)
      if (node && !node.matches(':disabled')) { node.focus(); return }
    }
  }, [listing, busy])

  // entryLabel is the ONE place an entry becomes text, and it holds no rule:
  // it reads the discriminant the engine projected and the display strings the
  // engine read from the document. A named face shows its name; an embedded
  // face shows the family and, when the document declared one, the style.
  const entryLabel = (entry: Props['chains'][number]['entries'][number]): string =>
    entry.assetKey.length > 0 ? [entry.family, entry.style].filter((part) => part.length > 0).join(' ') : entry.face

  const errorId = `${uid}-refusal`
  const anchored = (control: FontChainControl) => error && sameFontChainControl(error.control, control) ? error : undefined
  const flag = (control: FontChainControl) => anchored(control) ? { 'aria-invalid': 'true' as const, 'aria-errormessage': errorId } : {}
  // The refusal is the ENGINE'S string and is rendered exactly as it arrived.
  const refusal = (control: FontChainControl) => anchored(control) ? <p id={errorId} role="alert" className="property-error">{error?.message}</p> : undefined

  const dispatch = (payload: ArrayBuffer, control: FontChainControl, origin: string, targets: ReadonlyArray<string> = [], clear: ReadonlyArray<string> = []) => {
    pending.current = { origin, targets, clear, listing }
    onCommand(payload, control)
  }
  const rename = (chain: string, at: number, origin: string) => dispatch(renameFontChainCommand(chain, typed(`${uid}-name-${at}`)), { chain, action: 'rename' }, origin)
  const addEntry = (chain: string, at: number, position: number, origin: string) => dispatch(addFontChainEntryCommand(chain, at, typed(`${uid}-face-${position}`)), { chain, action: 'addEntry' }, origin, [], [`${uid}-face-${position}`])
  const addChain = (origin: string) => dispatch(addFontChainCommand(typed(`${uid}-new-name`), [typed(`${uid}-new-face`)]), { action: 'add' }, origin, [], [`${uid}-new-name`, `${uid}-new-face`])
  // Enter commits the field it is pressed in, so every control here is
  // reachable and operable from the keyboard alone (UX-DR25) without the
  // author having to leave the field to find its button.
  const enter = (event: KeyboardEvent<HTMLInputElement>, run: () => void) => { if (event.key === 'Enter') { event.preventDefault(); run() } }

  return <div className="font-chain-editor" aria-label="Font chains" role="group">
    <p className="section-label">FONT CHAINS</p>
    {chains.length === 0
      ? <p className="honest-note">This document declares no font chains.</p>
      : <ul className="font-chain-list">{chains.map((chain, position) => {
        const nameId = `${uid}-name-${position}`
        return <li key={chain.name} className="font-chain">
          <div className="font-chain-heading">
            <div className="property-field">
              {/* Uncontrolled, keyed by the PROJECTED name: when the engine
                  accepts a rename the key changes, the field remounts, and it
                  shows the engine's name rather than the author's draft. */}
              <input id={nameId} className="property-value" aria-label={`Font chain ${position + 1} name`} defaultValue={chain.name} disabled={busy} onKeyDown={(event) => enter(event, () => rename(chain.name, position, nameId))} {...flag({ chain: chain.name, action: 'rename' })} />
            </div>
            <button type="button" id={`${uid}-rename-${position}`} className="file-button" aria-label={`Rename font chain ${position + 1}`} disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => rename(chain.name, position, `${uid}-rename-${position}`)}>Rename</button>
            <button type="button" id={`${uid}-delete-${position}`} className="file-button" aria-label={`Delete font chain ${position + 1}`} disabled={busy} onClick={() => dispatch(deleteFontChainCommand(chain.name), { chain: chain.name, action: 'delete' }, `${uid}-delete-${position}`)} {...flag({ chain: chain.name, action: 'delete' })}>Delete</button>
          </div>
          {refusal({ chain: chain.name, action: 'rename' })}
          {refusal({ chain: chain.name, action: 'delete' })}
          <ol className="font-chain-entries" aria-label={`Entries of font chain ${position + 1}`}>
            {chain.entries.map((face, entry) => <li key={`${entry}:${face.face}:${face.assetKey}`} className="font-chain-entry">
              <span className="font-chain-entry-face">{entryLabel(face)}</span>
              <button type="button" id={`${uid}-earlier-${position}-${entry}`} className="property-inline-action" aria-label={`Move entry ${entry + 1} of font chain ${position + 1} earlier`} title="Move earlier" disabled={busy || entry === 0} onClick={() => dispatch(moveFontChainEntryCommand(chain.name, entry, entry - 1), { chain: chain.name, entry, action: 'moveEntryEarlier' }, `${uid}-earlier-${position}-${entry}`, [`${uid}-earlier-${position}-${entry - 1}`, `${uid}-later-${position}-${entry - 1}`])} {...flag({ chain: chain.name, entry, action: 'moveEntryEarlier' })}>↑</button>
              <button type="button" id={`${uid}-later-${position}-${entry}`} className="property-inline-action" aria-label={`Move entry ${entry + 1} of font chain ${position + 1} later`} title="Move later" disabled={busy || entry === chain.entries.length - 1} onClick={() => dispatch(moveFontChainEntryCommand(chain.name, entry, entry + 1), { chain: chain.name, entry, action: 'moveEntryLater' }, `${uid}-later-${position}-${entry}`, [`${uid}-later-${position}-${entry + 1}`, `${uid}-earlier-${position}-${entry + 1}`])} {...flag({ chain: chain.name, entry, action: 'moveEntryLater' })}>↓</button>
              <button type="button" id={`${uid}-remove-${position}-${entry}`} className="property-inline-action" aria-label={`Remove entry ${entry + 1} of font chain ${position + 1}`} title="Remove entry" disabled={busy} onClick={() => dispatch(removeFontChainEntryCommand(chain.name, entry), { chain: chain.name, entry, action: 'removeEntry' }, `${uid}-remove-${position}-${entry}`)} {...flag({ chain: chain.name, entry, action: 'removeEntry' })}>×</button>
              {refusal({ chain: chain.name, entry, action: 'moveEntryEarlier' })}
              {refusal({ chain: chain.name, entry, action: 'moveEntryLater' })}
              {refusal({ chain: chain.name, entry, action: 'removeEntry' })}
            </li>)}
          </ol>
          <div className="font-chain-add-entry">
            <div className="property-field">
              <input id={`${uid}-face-${position}`} className="property-value" aria-label={`New entry for font chain ${position + 1}`} placeholder="Face name" disabled={busy} onKeyDown={(event) => enter(event, () => addEntry(chain.name, chain.entries.length, position, `${uid}-face-${position}`))} {...flag({ chain: chain.name, action: 'addEntry' })} />
            </div>
            <button type="button" id={`${uid}-add-entry-${position}`} className="file-button" aria-label={`Add entry to font chain ${position + 1}`} disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => addEntry(chain.name, chain.entries.length, position, `${uid}-add-entry-${position}`)}>Add entry</button>
          </div>
          {refusal({ chain: chain.name, action: 'addEntry' })}
        </li>
      })}</ul>}
    <div className="font-chain-new">
      <div className="property-field">
        <input id={`${uid}-new-name`} className="property-value" aria-label="New font chain name" placeholder="Chain name" disabled={busy} onKeyDown={(event) => enter(event, () => addChain(`${uid}-new-name`))} {...flag({ action: 'add' })} />
      </div>
      <div className="property-field">
        <input id={`${uid}-new-face`} className="property-value" aria-label="First entry for the new font chain" placeholder="Face name" disabled={busy} onKeyDown={(event) => enter(event, () => addChain(`${uid}-new-face`))} {...flag({ action: 'add' })} />
      </div>
      <button type="button" id={`${uid}-add-chain`} className="file-button" disabled={busy} onMouseDown={(event) => event.preventDefault()} onClick={() => addChain(`${uid}-add-chain`)}>Add font chain</button>
    </div>
    {refusal({ action: 'add' })}
    <p className="honest-note">Chain names and faces are the engine's. A refused edit is shown here in the engine's own words, and the document is unchanged.</p>
  </div>
}
