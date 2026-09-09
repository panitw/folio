import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import App from './App'
import type { EngineClient } from './engine-client'

// STORY 14.1 — THE CONTROL VOCABULARY, IN ITS CHECKABLE FORM.
//
// The rule this file enforces is written in the story spec
// (`_bmad-output/implementation-artifacts/14-1-one-button-vocabulary.md`,
// `## Design Notes` → THE RULE) in `DESIGN.md`'s own vocabulary:
//
//   V1  A control is a word.
//   V2  A control is a glyph only as one member of a segmented control.
//   V3  Uniformity: (a) every control sharing a control class is spelled the
//       same way; (b) every control inside one named control group is spelled
//       the same way.
//   V4  A glyph always carries an accessible name.
//
// V3(a), V3(b) and V4 are enforced here as clauses R1, R2 and R3. V2 is
// recorded rather than enforced, as clause R4 — a pin, by accessible name,
// over the glyph controls that are not members of a segmented control. Story
// 14.1 changes none of them; 14.2, 14.3, 14.4 and 14.7 each own one of those
// surfaces and will rule on it with the surface in front of them.
//
// ⚠ THE POPULATION IS DERIVED, NOT LISTED. Every clause below runs over the
// controls a RENDER produces — never over a hand-written list of control names.
// That is the failure shape DW-305 recorded: `design-contract.test.ts:165`
// asserts a 4.5:1 contrast floor over five hand-written pairs, and a 2.25:1
// pairing shipped in Story 13.5 with nothing red, because the pair was not on
// the list. A list cannot see what it does not name.
//
// ⚠ AND IT RENDERS RATHER THAN SCANS SOURCE, deliberately. `SegmentedProperty`
// (`App.tsx`) is a SINGLE JSX site whose children are `{segment.content}`.
// Align and Vertical align differ only in the DATA handed to it — `alignSegments`
// versus `valignSegments` — so the exact defect AC3 names (icons in one cell,
// the words TOP / MID / BOT in the cell beside it, same class, same row, same
// size) is invisible to a source-text or AST scan of the markup. Only rendered
// output can tell an `<svg>` child from the string `TOP`. The cost is that the
// guard sees only what it renders, which is why the coverage assertions below
// pin the VISITED GROUP NAMES and not merely a count: a count floor cannot see
// a render state quietly dropping out.
//
// ⚠ NO MEASUREMENT IDENTIFIERS. This file is inside the corpus
// `canvas-authority-contract.test.ts` scans, so it must not name
// `getComputedStyle`, `getBoundingClientRect`, `measureText`, `document.fonts`,
// `devicePixelRatio`, or any of the `offset*` / `client*` / `scroll*`
// identifiers. Nothing here needs them: the subject is what a control SAYS, not
// how large it is.

type Treatment = 'word' | 'glyph' | 'empty'

type Control = Readonly<{
  element: Element
  state: string
  treatment: Treatment
  name: string
  classes: ReadonlyArray<string>
  segmented: boolean
  where: string
}>

// A letter or a digit in any script. `×`, `∅`, `−`, `+`, `◀`, `▶` and `·` carry
// none, which is exactly the line V1 and V2 draw: a typographic character
// standing in for an icon is a GLYPH, not a word, however it is encoded.
const LETTER_OR_DIGIT = /[\p{L}\p{N}]/u

// THE VISUALLY-HIDDEN CLASSES, DERIVED FROM THE STYLESHEET rather than listed.
//
// ⚠ WITHOUT THIS, THE CLASSIFIER HAS A BLIND SPOT WIDE ENOUGH TO DRIVE A GLYPH
// THROUGH. `<button><svg aria-hidden="true"/><span class="sr-only">Zoom in</span></button>`
// is a control that LOOKS like an icon and READS like a word: strip only the
// `aria-hidden` subtree and the remaining text is `Zoom in`, so it classifies as
// a word and escapes R1, R2, R3 and R4's census together. That is DW-305's shape
// again — a guard that passes because it cannot see the thing. `.sr-only`
// (`App.css:7`) is real, Story 13.5 introduced it, `.diagnostic-announcement`
// (`App.css:594`) is a second copy of the same idiom, and neither sits inside a
// control TODAY. The point is that neither has to.
//
// The set is read out of `App.css` by the declarations that DEFINE the idiom —
// clipped to nothing while staying in the accessibility tree — so a third class
// written tomorrow is covered the day it is written, with no edit here.
// `getComputedStyle` is not available to this file (it is inside
// `canvas-authority-contract.test.ts`'s scanned corpus) and would not help in
// jsdom regardless; the stylesheet source is the honest reading.
const stylesheet = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'App.css'), 'utf8')
const VISUALLY_HIDDEN_CLASSES: ReadonlySet<string> = new Set(
  Array.from(stylesheet.matchAll(/([^{}]+)\{([^{}]*)\}/g))
    .filter((rule) => /clip-path\s*:\s*inset\(\s*50%\s*\)/.test(rule[2] ?? '') && /position\s*:\s*absolute/.test(rule[2] ?? ''))
    .flatMap((rule) => Array.from((rule[1] ?? '').matchAll(/\.([A-Za-z0-9_-]+)/g)).map((match) => match[1] as string))
    .sort(),
)
// `[hidden]` joins them for the same reason: the platform removes it from the
// page and from the tree, so text inside one is not what the control says.
const HIDDEN_SUBTREES = ['[aria-hidden="true"]', '[hidden]', ...[...VISUALLY_HIDDEN_CLASSES].map((token) => `.${token}`)].join(', ')

// The visible text a reader is left with once the hidden subtrees are gone.
// `Undo <kbd aria-hidden="true">⌘Z</kbd>` reads as `Undo`, so a shortcut hint
// never turns a word into something else and never turns a glyph into a word.
function visibleText(control: Element): string {
  const clone = control.cloneNode(true) as Element
  clone.querySelectorAll(HIDDEN_SUBTREES).forEach((node) => node.remove())
  return (clone.textContent ?? '').replace(/\s+/g, ' ').trim()
}

// THE CLASSIFIER — one helper, applied to every control the sweep visits.
//
// ⚠ THE `aria-hidden` STRIP APPLIES TO TEXT, AND THE `<svg>` TEST DOES NOT, and
// that asymmetry is the whole point rather than an oversight. Every icon in this
// designer is written `<svg aria-hidden="true">` and every icon-bearing control
// carries its own `aria-label` — that IS the correct spelling (UX-DR25,
// `EXPERIENCE.md:282`), and it is what makes the glyph decorative to a screen
// reader while remaining the only thing a sighted reader sees. Stripping the
// icon before asking "is this a glyph?" would classify every correctly-written
// icon control as empty and leave R4's census permanently blank — a guard that
// passes because it looks in the wrong place. So: strip `aria-hidden` to read
// the TEXT, and look for the `<svg>` in the control as it actually renders.
export function treatmentOf(control: Element): Treatment {
  const text = visibleText(control)
  if (LETTER_OR_DIGIT.test(text)) return 'word'
  if (control.querySelector('svg') !== null) return 'glyph'
  if (text.length > 0) return 'glyph'
  return 'empty'
}

// The accessible name, computed over the forms this designer actually uses:
// `aria-label`, then `aria-labelledby`, then name-from-content with the
// `aria-hidden` subtrees excluded, then `title`.
//
// ⚠ IT IS NOT TRUSTED ON ITS OWN. The row below titled "agrees with the
// accessibility tree about every swept control's name" asserts this function's
// answer against jest-dom's `toHaveAccessibleName` for EVERY control the sweep
// visits, in every state. A resolver that quietly disagreed with the
// accessibility tree would make R3 and R4 pin the wrong strings, and nothing
// else in this file would notice.
export function accessibleName(element: Element): string {
  const label = element.getAttribute('aria-label')
  if (label !== null && label.trim() !== '') return label.replace(/\s+/g, ' ').trim()
  const ids = (element.getAttribute('aria-labelledby') ?? '').split(/\s+/).filter((id) => id !== '')
  if (ids.length > 0) {
    const referenced = ids.map((id) => element.ownerDocument.getElementById(id)?.textContent ?? '').join(' ').replace(/\s+/g, ' ').trim()
    if (referenced !== '') return referenced
  }
  // ⚠ NAME-FROM-CONTENT READS THE `.sr-only` CAPTION, and `visibleText` does not.
  // The two are deliberately different readings of the same control: what a
  // SIGHTED reader sees (which is the subject of V1/V2/V3) versus what a SCREEN
  // READER announces (which is the subject of V4). A visually-hidden caption is
  // invisible to the first and load-bearing for the second. Only `aria-hidden`
  // and `[hidden]` are dropped here, because those leave the tree too.
  const named = element.cloneNode(true) as Element
  named.querySelectorAll('[aria-hidden="true"], [hidden]').forEach((node) => node.remove())
  const content = (named.textContent ?? '').replace(/\s+/g, ' ').trim()
  if (content !== '') return content
  const title = element.getAttribute('title')
  return title === null ? '' : title.replace(/\s+/g, ' ').trim()
}

// A container is a NAMED CONTROL GROUP if the accessibility tree gives it both a
// grouping role and a name. `aria-label` on a roleless `<div>` is neither — the
// tree drops it — which is the defect Story 14.1 fixed on `.document-actions`
// and, by orchestrator ruling, on `.mode-switch`.
function groupsIn(root: ParentNode): ReadonlyArray<Element> {
  return Array.from(root.querySelectorAll('[role="group"], [role="tablist"]'))
}

function groupName(group: Element): string {
  const label = group.getAttribute('aria-label')
  if (label !== null && label.trim() !== '') return label.replace(/\s+/g, ' ').trim()
  const ids = (group.getAttribute('aria-labelledby') ?? '').split(/\s+/).filter((id) => id !== '')
  const referenced = ids.map((id) => group.ownerDocument.getElementById(id)?.textContent ?? '').join(' ').replace(/\s+/g, ' ').trim()
  return referenced
}

function controlsIn(root: ParentNode): ReadonlyArray<Element> {
  return Array.from(root.querySelectorAll('button, [role="button"]'))
}

// SEGMENTED-CONTROL MEMBERSHIP, DERIVED — never a class name, because the class
// is exactly the thing a future control could quietly adopt or shed. A group is
// a `{components.segmented-control}` when it holds two or more controls and
// EVERY one of them carries `aria-pressed`: that is what "the mutually-exclusive
// values of a single closed-set property" looks like in the tree. Align,
// Vertical align and the DESIGN/PREVIEW mode switch qualify; `PDF navigation`,
// `Local file actions`, `Border edges` and the inspector tablist do not.
function isSegmentedControl(group: Element): boolean {
  const members = controlsIn(group)
  return members.length >= 2 && members.every((member) => member.hasAttribute('aria-pressed'))
}

function sweepControls(root: ParentNode, state: string): ReadonlyArray<Control> {
  const segmented = groupsIn(root).filter(isSegmentedControl)
  const elements = controlsIn(root)
  return elements.map((element, index) => ({
    element,
    state,
    treatment: treatmentOf(element),
    name: accessibleName(element),
    classes: (element.getAttribute('class') ?? '').split(/\s+/).filter((token) => token !== ''),
    segmented: segmented.some((group) => group.contains(element)),
    where: `${state} · control ${index + 1} of ${elements.length}`,
  }))
}

const describeControl = (control: Control) => `${control.name === '' ? '(no accessible name)' : `"${control.name}"`} [${control.treatment}]`

// R1 — V3(a). Every control that shares a control class is spelled the same way.
// `empty` controls (the drag handles) are excluded: they have no spelling to
// disagree about, and R3 is what holds them to account.
function r1Violations(controls: ReadonlyArray<Control>): ReadonlyArray<string> {
  const byClass = new Map<string, Control[]>()
  for (const control of controls) {
    if (control.treatment === 'empty') continue
    for (const token of control.classes) byClass.set(token, [...(byClass.get(token) ?? []), control])
  }
  return [...byClass.entries()].filter(([, members]) => members.length >= 2 && new Set(members.map((member) => member.treatment)).size > 1)
    .map(([token, members]) => `R1 class "${token}" mixes treatments: ${members.map(describeControl).join(', ')}`)
    .sort()
}

// R2 — V3(b). Every control inside one named control group is spelled the same
// way. A row that puts icons beside words at the same size is this clause.
function r2Violations(root: ParentNode, state: string): ReadonlyArray<string> {
  return groupsIn(root).map((group) => {
    const members = controlsIn(group).map((element) => ({ element, state, treatment: treatmentOf(element), name: accessibleName(element), classes: [], segmented: false, where: state } as Control)).filter((member) => member.treatment !== 'empty')
    if (members.length < 2 || new Set(members.map((member) => member.treatment)).size === 1) return undefined
    return `R2 group "${groupName(group) === '' ? '(no accessible name)' : groupName(group)}" mixes treatments: ${members.map(describeControl).join(', ')}`
  }).filter((message): message is string => message !== undefined).sort()
}

// R3 — V4 / UX-DR25 / `EXPERIENCE.md:282`. A control a reader cannot read has to
// be a control a screen reader can announce.
function r3Violations(controls: ReadonlyArray<Control>): ReadonlyArray<string> {
  return controls.filter((control) => control.treatment !== 'word' && control.name === '')
    .map((control) => `R3 ${control.treatment} control with an empty accessible name at ${control.where}`)
    .sort()
}

// R4 — the V2 census. Pinned BY ACCESSIBLE NAME, never by count: a count floor
// passes when one control is renamed and another appears in its place. This
// story records the set and changes none of it; one more reds and forces a
// human ruling.
//
// GLYPH, not glyph-or-empty, and R4's own wording is why: "the set of GLYPH
// controls that are not members of a segmented control". An `empty` control has
// no spelling to be wrong about — it is a drag handle or a canvas element, not a
// picture standing in for a word — and R3 is what holds it to an accessible
// name. Folding empties in here would also pull the canvas's own elements into a
// census of chrome controls, which is a different subject.
function censusMembers(controls: ReadonlyArray<Control>): ReadonlyArray<Control> {
  return controls.filter((control) => control.treatment === 'glyph' && !control.segmented)
}
// ⚠ A MULTISET KEYED BY STATE AND NAME, NOT A SET OF NAMES, and that is R-Q2's
// ruling carried all the way through rather than half way. A `Set` of names
// collapses a repeat: a SECOND glyph control appearing under a name the census
// already holds produces neither an arrival nor a departure, so the clause whose
// whole purpose is "never by count" would have been silently insensitive to
// exactly one more control. Keying by state as well as name also means a glyph
// that stops rendering in ONE state while surviving in another still reds.
const censusKey = (control: Control) => `${control.state} · ${control.name === '' ? '(no accessible name)' : control.name}`
const tally = (keys: ReadonlyArray<string>): ReadonlyMap<string, number> => {
  const counts = new Map<string, number>()
  for (const key of keys) counts.set(key, (counts.get(key) ?? 0) + 1)
  return counts
}
function r4Violations(controls: ReadonlyArray<Control>, census: ReadonlyArray<string>): ReadonlyArray<string> {
  const found = tally(censusMembers(controls).map(censusKey))
  const recorded = tally(census)
  return [...new Set([...found.keys(), ...recorded.keys()])].sort().flatMap((key) => {
    const now = found.get(key) ?? 0
    const before = recorded.get(key) ?? 0
    if (now === before) return []
    return [now > before
      ? `R4 the closed set moved: ${now} glyph control(s) outside every segmented control now render as ${key}, where the census records ${before}`
      : `R4 the closed set shrank: ${now} glyph control(s) outside every segmented control now render as ${key}, where the census records ${before}`]
  })
}

// ---------------------------------------------------------------------------
// THE DECLARED RENDER STATES.
// ---------------------------------------------------------------------------

const canvas = { width: 595276, height: 841890, orientation: 'portrait' as const, preset: 'A4' as const, locale: 'en' as const, utcOffset: '+07:00', marginTop: 36000, marginRight: 36000, marginBottom: 36000, marginLeft: 36000, gridIncrement: 6000, commandWidth: 595276, commandHeight: 841890, fontFamilies: ['body'], fontChains: [{ name: 'body', entries: [{ face: 'Noto Sans', assetKey: '', family: '', style: '', bold: '', italic: '', boldItalic: '' }] }], defaultFontSize: 12000, defaultLineSpacing: 1000, contentWindowHeight: 729890, contentWindowCount: 1, contentWindowOrigins: [0], contentWindowCountIsExact: true, bands: [{ name: 'pageHeader' as const, x: 36000, y: 36000, width: 523276, height: 20000 }, { name: 'content' as const, x: 36000, y: 56000, width: 523276, height: 729890 }, { name: 'pageFooter' as const, x: 36000, y: 785890, width: 523276, height: 20000 }], components: [] }
const textComponent = { id: 'e1', type: 'text' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 24_000, resizable: true, value: 'Hello' }
const tableComponent = { id: 'e7', type: 'table' as const, band: 'content' as const, x: 0, y: 0, width: 72_000, height: 12_000, resizable: false }
const engine = () => ({ request: vi.fn(async () => ({ snapshot: { documentState: 'loaded' as const, revision: 2, byteLength: 3 } })) }) as unknown as EngineClient

const mount = (components: ReadonlyArray<typeof textComponent | typeof tableComponent>) =>
  render(<App engine={engine()} initialSnapshot={{ documentState: 'loaded', revision: 1, byteLength: 3, canvas: { ...canvas, components: [...components] } }} />)

// Each state is a NAMED render, so a state that stops producing controls is
// named in the failure rather than absorbed into a smaller total.
const states: ReadonlyArray<Readonly<{ name: string; open: () => Element }>> = [
  {
    name: 'design · nothing selected',
    open: () => mount([textComponent, tableComponent]).container,
  },
  {
    name: 'design · a text element selected',
    open: () => {
      const view = mount([textComponent, tableComponent])
      fireEvent.click(within(view.container).getByLabelText('text component e1'))
      return view.container
    },
  },
  {
    name: 'design · a table element selected',
    open: () => {
      const view = mount([textComponent, tableComponent])
      fireEvent.click(within(view.container).getByLabelText('table component e7'))
      return view.container
    },
  },
  {
    name: 'preview',
    open: () => {
      const view = mount([textComponent, tableComponent])
      fireEvent.click(within(view.container).getByRole('button', { name: 'PREVIEW' }))
      return view.container
    },
  },
]

type Swept = Readonly<{ state: string; root: Element; controls: ReadonlyArray<Control> }>
const sweepStates = (declared: ReadonlyArray<Readonly<{ name: string; open: () => Element }>>): ReadonlyArray<Swept> =>
  declared.map((state) => {
    const root = state.open()
    return { state: state.name, root, controls: sweepControls(root, state.name) }
  })
const sweepEveryState = (): ReadonlyArray<Swept> => sweepStates(states)

// THE V2 CENSUS AS IT STANDS at Story 14.1 — eleven distinct names over twenty
// renderings, DERIVED from the sweep above and then written down here so that
// one more reds. Story 14.1
// changes none of them; 14.2, 14.3, 14.4 and 14.7 each own one of these surfaces
// and will rule on it with the surface in front of them (R-Q2).
//
// Two items the story's prose enumerates are deliberately ABSENT from this set,
// and the classifier is why rather than an oversight:
//   • the `B` / `I` weight and slope toggles classify as WORDS. `B` and `I` are
//     letters, and V1's own test is "contains a letter or digit" — they are the
//     initial of the property they set, not a picture of it. Whether an initial
//     is a word is a judgement, and this story does not make it silently.
//   • the drag handles (`Resize e1`, `Resize the page header`, `Resize the page
//     footer`) and the canvas's own elements classify as EMPTY: they render no
//     text and no `<svg>`, so they have no spelling to be wrong. R3 still holds
//     each of them to an accessible name, and every one passes.
// Both are reported to the orchestrator with the rest of the audit rather than
// quietly folded in here.
const V2_CENSUS: ReadonlyArray<string> = [
  // `.canvas-tools` — a roleless `<div>` whose `aria-label` the tree drops, so
  // R2 cannot see that it also puts these two beside Grid / Snap / Duplicate /
  // Delete. Reported, not fixed: no AC names it. (Design's three states only.)
  'design · nothing selected · Zoom out',
  'design · nothing selected · Zoom in',
  'design · a text element selected · Zoom out',
  'design · a text element selected · Zoom in',
  'design · a table element selected · Zoom out',
  'design · a table element selected · Zoom in',
  // `.property-inline-action` — the inspector's `×` clear, `∅` null and the
  // font-family disclosure chevron. Uniform within their class, so R1 is green.
  // They render only while something is selected.
  'design · a text element selected · Clear Font size (pt)',
  'design · a text element selected · Clear Line spacing',
  'design · a text element selected · Set Background null',
  'design · a text element selected · Set Visible if null',
  'design · a text element selected · Show fonts',
  'design · a table element selected · Clear Font size (pt)',
  'design · a table element selected · Clear Line spacing',
  'design · a table element selected · Set Background null',
  'design · a table element selected · Set Visible if null',
  'design · a table element selected · Show fonts',
  // The PDF navigation group — uniform within its group, so R2 is green.
  'preview · Previous PDF page',
  'preview · Next PDF page',
  'preview · Zoom out PDF',
  'preview · Zoom in PDF',
]

// R0 — COVERAGE HONESTY, and it is a CLAUSE like the other four rather than a
// bare assertion, so that it too can be run against a planted shrink and watched
// to fail. A guard whose own coverage check has never been seen to red is the
// same list-shaped trap the rest of this file is written against.
//
// ⚠ THE GROUPS ARE SPLIT INTO CHECKED AND PRESENT-BUT-UNDER-ARITY, because
// R2 returns early on a group holding fewer than two non-empty swept controls
// and would otherwise report such a group as covered when nothing checked it.
// `Border edges` is exactly that: its four edge controls are
// `<input type="checkbox">` and its only button is a `×` clear that renders
// solely once an edge is set, so in every declared state R2 sees zero members
// there. Recording it in the second bucket makes the hole VISIBLE — and a group
// silently sliding from the first bucket into the second now reds.
//
// (`PDF navigation` is in the FIRST bucket, measured: its `<select>` and two
// text inputs are outside this story's swept population, but its four `◀ ▶ − +`
// buttons are not, so R2 does check it over four members.)
const GROUP_ARITY_FLOOR = 2
const CHECKED_GROUPS: ReadonlySet<string> = new Set([
  'Local file actions',
  'Designer mode',
  'Inspector tabs',
  'Align',
  'Vertical align',
  'PDF navigation',
  'Render actions',
])
const UNDER_ARITY_GROUPS: ReadonlySet<string> = new Set(['Border edges'])
// Counts enumerated at the Story 14.1 baseline (125 controls, 17 class tokens,
// 20 group instances, smallest state 17). Floors, not equalities, so ordinary
// growth never churns the guard while any shrink reddens.
const PER_STATE_CONTROL_FLOOR = 15
const CONTROL_FLOOR = 100
const CLASS_FAMILY_FLOOR = 15
const GROUP_INSTANCE_FLOOR = 18

// The most non-empty swept controls any single state puts inside a group of this
// name — the arity R2 actually got to work with at its best.
function groupArity(swept: ReadonlyArray<Swept>): ReadonlyMap<string, number> {
  const best = new Map<string, number>()
  for (const entry of swept) {
    for (const group of groupsIn(entry.root)) {
      const members = controlsIn(group).filter((element) => treatmentOf(element) !== 'empty').length
      best.set(groupName(group), Math.max(best.get(groupName(group)) ?? 0, members))
    }
  }
  return best
}

function r0Violations(swept: ReadonlyArray<Swept>): ReadonlyArray<string> {
  const messages: string[] = []
  const controls = swept.flatMap((entry) => entry.controls)
  const families = new Set(controls.flatMap((control) => control.classes))
  const groupInstances = swept.flatMap((entry) => groupsIn(entry.root))
  for (const entry of swept) {
    if (entry.controls.length < PER_STATE_CONTROL_FLOOR) messages.push(`R0 state "${entry.state}" swept ${entry.controls.length} controls, under its floor of ${PER_STATE_CONTROL_FLOOR}`)
  }
  if (controls.length < CONTROL_FLOOR) messages.push(`R0 the sweep visited ${controls.length} controls, under the floor of ${CONTROL_FLOOR}`)
  if (families.size < CLASS_FAMILY_FLOOR) messages.push(`R0 the sweep visited ${families.size} class families, under the floor of ${CLASS_FAMILY_FLOOR}`)
  if (groupInstances.length < GROUP_INSTANCE_FLOOR) messages.push(`R0 the sweep visited ${groupInstances.length} group instances, under the floor of ${GROUP_INSTANCE_FLOOR}`)
  // ⚠ AND THE FLOORS ABOVE CANNOT SEE A SHRINK BY NAME, which is what this
  // clause exists for: drop the preview state and every total above still clears
  // its floor while `PDF navigation` and `Render actions` quietly stop being
  // checked. These say WHICH group went, and whether it went missing or merely
  // fell below the arity R2 needs.
  const arity = groupArity(swept)
  const checked = new Set([...arity].filter(([, members]) => members >= GROUP_ARITY_FLOOR).map(([name]) => name))
  const underArity = new Set([...arity].filter(([, members]) => members < GROUP_ARITY_FLOOR).map(([name]) => name))
  for (const name of [...CHECKED_GROUPS].sort()) {
    if (checked.has(name)) continue
    messages.push(arity.has(name)
      ? `R0 the group "${name}" is no longer checked by R2: it renders ${arity.get(name)} non-empty control(s), under the arity of ${GROUP_ARITY_FLOOR}`
      : `R0 the group "${name}" renders in no declared state, so nothing checked it`)
  }
  for (const name of [...checked].sort()) {
    if (!CHECKED_GROUPS.has(name)) messages.push(`R0 a group R2 now checks is not recorded as checked: "${name}"`)
  }
  for (const name of [...UNDER_ARITY_GROUPS].sort()) {
    if (!underArity.has(name)) messages.push(`R0 the group "${name}" is recorded as present-but-unchecked and no longer renders that way`)
  }
  for (const name of [...underArity].sort()) {
    if (!UNDER_ARITY_GROUPS.has(name)) messages.push(`R0 a group is present but under R2's arity and is not recorded as such: "${name}" (${arity.get(name)} non-empty control(s))`)
  }
  // Every treatment the classifier can return is actually exercised, so a
  // classifier that had collapsed to one answer would red here rather than pass
  // every clause vacuously.
  const treatments = new Set(controls.map((control) => control.treatment))
  for (const treatment of ['word', 'glyph', 'empty'] as const) {
    if (!treatments.has(treatment)) messages.push(`R0 no swept control classified as ${treatment}, so the clauses over that treatment ran vacuously`)
  }
  return messages
}

describe('control vocabulary contract', () => {
  it('R0 — the sweep is non-vacuous, and every recorded group is still covered the way it was', () => {
    const swept = sweepEveryState()
    expect(swept.map((entry) => entry.state)).toEqual(states.map((state) => state.name))
    expect(r0Violations(swept)).toEqual([])
    // The clause's own inputs are non-vacuous: a `VISUALLY_HIDDEN_CLASSES` that
    // parsed to nothing would silently restore the P3 blind spot, and no other
    // assertion in this file would notice.
    expect([...VISUALLY_HIDDEN_CLASSES]).toEqual(['diagnostic-announcement', 'sr-only'])
  })

  it('agrees with the accessibility tree about every swept control\'s name', () => {
    for (const entry of sweepEveryState()) {
      for (const control of entry.controls) {
        if (control.name === '') expect(control.element, control.where).not.toHaveAccessibleName()
        else expect(control.element, control.where).toHaveAccessibleName(control.name)
      }
    }
  })

  it('R1 — every control sharing a control class is spelled the same way', () => {
    for (const entry of sweepEveryState()) expect(r1Violations(entry.controls), entry.state).toEqual([])
  })

  it('R2 — every control inside one named control group is spelled the same way', () => {
    for (const entry of sweepEveryState()) expect(r2Violations(entry.root, entry.state), entry.state).toEqual([])
  })

  it('R3 — every glyph and every empty control carries an accessible name', () => {
    for (const entry of sweepEveryState()) expect(r3Violations(entry.controls), entry.state).toEqual([])
  })

  it('R4 — the set of glyph controls outside every segmented control has not moved', () => {
    expect(r4Violations(sweepEveryState().flatMap((entry) => entry.controls), V2_CENSUS)).toEqual([])
  })

  // -------------------------------------------------------------------------
  // THE CLAUSES, RUN AGAINST A PLANTED VIOLATION AND AGAINST THE NEAREST
  // LEGITIMATE SPELLING OF THE SAME THING (D-11.3.7). Reading a rule is not
  // testing a rule; each red below is EXECUTED and each is asserted to name the
  // clause it must wake and NO OTHER, so "something failed" cannot pass for a
  // proof.
  // -------------------------------------------------------------------------

  // The harness every planted fragment runs through, and the reason each proof
  // can claim ONE clause: `census` defaults to the fragment's own glyph set, so
  // R4 is silent by construction unless a proof perturbs it deliberately.
  const rulesOver = (root: Element, census?: ReadonlyArray<string>): ReadonlyArray<string> => {
    const controls = sweepControls(root, 'planted')
    const settled = census ?? censusMembers(controls).map(censusKey)
    return [...r1Violations(controls), ...r2Violations(root, 'planted'), ...r3Violations(controls), ...r4Violations(controls, settled)]
  }
  const clausesWoken = (root: Element, census?: ReadonlyArray<string>): ReadonlyArray<string> => [...new Set(rulesOver(root, census).map((message) => message.slice(0, 2)))].sort()
  const glyph = <svg aria-hidden="true" viewBox="0 0 16 16"><path d="M2 8h12" /></svg>

  it('R1 reds on the exact defect AC3 names — one class, an icon in one cell and a word in the next', () => {
    // The shipped defect, transcribed: ONE `.property-segment` class, one row,
    // an `<svg>` on one side and the string TOP on the other.
    const { container } = render(<div>
      <button type="button" className="property-segment" aria-label="Align left">{glyph}</button>
      <button type="button" className="property-segment" aria-label="Vertical align top">TOP</button>
    </div>)
    expect(rulesOver(container)).toEqual(['R1 class "property-segment" mixes treatments: "Align left" [glyph], "Vertical align top" [word]'])
    expect(clausesWoken(container)).toEqual(['R1'])
  })

  it('R1 passes over the nearest legitimate spelling — the same class, both cells iconic', () => {
    const { container } = render(<div>
      <button type="button" className="property-segment" aria-label="Align left">{glyph}</button>
      <button type="button" className="property-segment" aria-label="Vertical align top">{glyph}</button>
    </div>)
    expect(rulesOver(container)).toEqual([])
  })

  it('R2 reds on the exact defect AC2 names — one named group holding glyphs beside words', () => {
    // The document bar as it shipped: Open and Save as icon-only buttons in the
    // same named family as Save As and Start blank.
    const { container } = render(<div role="group" aria-label="Local file actions">
      <button type="button" aria-label="Open local template">{glyph}</button>
      <button type="button" aria-label="Save local template">{glyph}</button>
      <button type="button">Save As</button>
      <button type="button">Start blank</button>
    </div>)
    expect(rulesOver(container)).toEqual([
      'R2 group "Local file actions" mixes treatments: "Open local template" [glyph], "Save local template" [glyph], "Save As" [word], "Start blank" [word]',
    ])
    expect(clausesWoken(container)).toEqual(['R2'])
  })

  it('R2 passes over the nearest legitimate spelling — the same group, every member a word', () => {
    const { container } = render(<div role="group" aria-label="Local file actions">
      <button type="button" aria-label="Open local template">Open</button>
      <button type="button" aria-label="Save local template">Save</button>
      <button type="button">Save As</button>
      <button type="button">Start blank</button>
    </div>)
    expect(rulesOver(container)).toEqual([])
  })

  it('R2 does not mistake a shortcut hint for a treatment — an `aria-hidden` <kbd> beside a word is still a word', () => {
    // The nearest legitimate spelling of the R2 red above, and the one the
    // document bar actually ships: Undo and Redo end in an `aria-hidden` <kbd>
    // while Save As does not. A classifier that read raw `textContent` would
    // still call all three words; one that stripped the whole control would call
    // two of them glyphs and red a shipped, correct row.
    const { container } = render(<div role="group" aria-label="Local file actions">
      <button type="button">Save As</button>
      <button type="button">Undo <kbd aria-hidden="true">⌘Z</kbd></button>
      <button type="button">Redo <kbd aria-hidden="true">⇧⌘Z</kbd></button>
    </div>)
    expect(rulesOver(container)).toEqual([])
    expect(treatmentOf(screen.getByRole('button', { name: 'Undo' }))).toEqual('word')
  })

  it('R3 reds on a glyph with no accessible name, and on an empty control with none', () => {
    const { container } = render(<div>
      <button type="button">{glyph}</button>
      <button type="button" className="resize-handle" />
    </div>)
    expect(rulesOver(container)).toEqual([
      'R3 empty control with an empty accessible name at planted · control 2 of 2',
      'R3 glyph control with an empty accessible name at planted · control 1 of 2',
    ])
    expect(clausesWoken(container)).toEqual(['R3'])
  })

  it('R3 passes over the nearest legitimate spelling — the same two controls, each named', () => {
    const { container } = render(<div>
      <button type="button" aria-label="Align left">{glyph}</button>
      <button type="button" className="resize-handle" aria-label="Resize e1" />
    </div>)
    expect(rulesOver(container)).toEqual([])
  })

  it('sees a glyph wearing a visually-hidden caption, and does not mistake it for a word', () => {
    // THE P3 BLIND SPOT, PLANTED. `<svg aria-hidden/>` plus a `.sr-only` caption
    // is a control that looks like an icon and reads like a word: strip only the
    // `aria-hidden` subtree and the leftover text is `Zoom in`, so it would
    // classify as a WORD and escape R1, R2, R3 and the R4 census together.
    const { container } = render(<div>
      <button type="button" className="canvas-stepper">{glyph}<span className="sr-only">Zoom in</span></button>
      <button type="button" className="canvas-stepper">Grid on</button>
    </div>)
    const stepper = screen.getAllByRole('button')[0] as Element
    expect(treatmentOf(stepper)).toEqual('glyph')
    // Its NAME still comes from that caption — the two readings are different on
    // purpose, and V4 must keep working through a `.sr-only` label.
    expect(accessibleName(stepper)).toEqual('Zoom in')
    expect(stepper).toHaveAccessibleName('Zoom in')
    // And now the class family it shares with a word is a live R1 violation
    // instead of two words agreeing about nothing.
    expect(rulesOver(container)).toEqual(['R1 class "canvas-stepper" mixes treatments: "Zoom in" [glyph], "Grid on" [word]'])
    expect(clausesWoken(container)).toEqual(['R1'])
  })

  it('R4 reds when a glyph control appears outside every segmented control, and names it', () => {
    const { container } = render(<div>
      <button type="button" aria-label="Collapse the inspector">×</button>
    </div>)
    expect(rulesOver(container, [])).toEqual([
      'R4 the closed set moved: 1 glyph control(s) outside every segmented control now render as planted · Collapse the inspector, where the census records 0',
    ])
    expect(clausesWoken(container, [])).toEqual(['R4'])
  })

  it('R4 reds when a recorded member stops rendering, rather than passing over a smaller set', () => {
    const { container } = render(<div />)
    expect(rulesOver(container, ['planted · Zoom in'])).toEqual([
      'R4 the closed set shrank: 0 glyph control(s) outside every segmented control now render as planted · Zoom in, where the census records 1',
    ])
  })

  // R-Q2 IN FULL: "pin by accessible name, NEVER BY COUNT". A `Set` of names
  // would pass this — a second control under a name the census already holds is
  // neither an arrival nor a departure — which is the same shape as the count
  // floor the ruling rejected. The multiset sees it.
  it('R4 reds when a SECOND control appears under a name the census already holds', () => {
    const { container } = render(<div>
      <button type="button" aria-label="Zoom in">+</button>
      <button type="button" aria-label="Zoom in">+</button>
    </div>)
    expect(rulesOver(container, ['planted · Zoom in'])).toEqual([
      'R4 the closed set moved: 2 glyph control(s) outside every segmented control now render as planted · Zoom in, where the census records 1',
    ])
    expect(clausesWoken(container, ['planted · Zoom in'])).toEqual(['R4'])
  })

  it('R4 passes over the nearest legitimate spelling — a glyph INSIDE a segmented control is not in the census', () => {
    // The same glyph, one row over. `Align left` is a member of a
    // `{components.segmented-control}` — two or more mutually-exclusive values of
    // one closed-set property, each carrying `aria-pressed` — which is the one
    // place V2 permits a glyph, so it never enters R4's set at all.
    const { container } = render(<div role="group" aria-label="Align">
      <button type="button" aria-pressed={true} aria-label="Align left">{glyph}</button>
      <button type="button" aria-pressed={false} aria-label="Align center">{glyph}</button>
    </div>)
    expect(rulesOver(container, [])).toEqual([])
  })

  it('R0 reds when a render state is dropped, and NAMES the groups that stopped being checked', () => {
    // THE FAILURE THE COVERAGE CLAUSE EXISTS FOR, EXECUTED rather than claimed.
    // Drop the preview state — the cheapest way for this guard to quietly get
    // smaller. The control and class-family totals STILL CLEAR their floors: the
    // remaining three states sweep 108 controls between them, over the floor of
    // 100. The group-instance floor does red, but only as a smaller number; it
    // cannot say what left. The by-name half names both groups that stopped
    // being checked, which is the whole point of recording them by name.
    const shrunk = sweepStates(states.slice(0, 3))
    expect(shrunk.flatMap((entry) => entry.controls).length).toBeGreaterThanOrEqual(CONTROL_FLOOR)
    expect(new Set(shrunk.flatMap((entry) => entry.controls).flatMap((control) => control.classes)).size).toBeGreaterThanOrEqual(CLASS_FAMILY_FLOOR)
    expect(r0Violations(shrunk)).toEqual([
      'R0 the sweep visited 15 group instances, under the floor of 18',
      'R0 the group "PDF navigation" renders in no declared state, so nothing checked it',
      'R0 the group "Render actions" renders in no declared state, so nothing checked it',
    ])
  })

  it('R0 reds when a recorded group slips below the arity R2 needs, rather than reporting it covered', () => {
    // The other half of P4's disclosure: a group that still RENDERS but stops
    // holding two swept controls is one R2 returns early on. Before this clause
    // it would have kept its place in the visited set and read as checked.
    const { container } = render(<div>
      <div role="group" aria-label="Designer mode"><button type="button">DESIGN</button></div>
    </div>)
    expect(r0Violations([{ state: 'planted', root: container, controls: sweepControls(container, 'planted') }])
      .filter((message) => message.includes('Designer mode'))).toEqual([
      'R0 the group "Designer mode" is no longer checked by R2: it renders 1 non-empty control(s), under the arity of 2',
      'R0 a group is present but under R2\'s arity and is not recorded as such: "Designer mode" (1 non-empty control(s))',
    ])
  })

  // -------------------------------------------------------------------------
  // AC2 / AC3 / AC4 — the two surfaces this story respells, asserted where the
  // clauses above cannot: a clause says "no member disagrees", not "this member
  // says Open".
  // -------------------------------------------------------------------------

  it('spells all six local-file controls as words inside one named group', () => {
    const root = states[0]!.open()
    const group = screen.getByRole('group', { name: 'Local file actions' })
    expect(root.contains(group)).toBe(true)
    const members = controlsIn(group).map((element) => ({ name: accessibleName(element), treatment: treatmentOf(element), text: visibleText(element) }))
    expect(members).toEqual([
      { name: 'Open local template', treatment: 'word', text: 'Open' },
      { name: 'Save local template', treatment: 'word', text: 'Save' },
      { name: 'Save As', treatment: 'word', text: 'Save As' },
      { name: 'Start blank', treatment: 'word', text: 'Start blank' },
      { name: 'Undo', treatment: 'word', text: 'Undo' },
      { name: 'Redo', treatment: 'word', text: 'Redo' },
    ])
  })

  it('draws both TYPOGRAPHY segmented controls in one vocabulary, with every accessible name intact', () => {
    states[1]!.open()
    const align = controlsIn(screen.getByRole('group', { name: 'Align' }))
    const valign = controlsIn(screen.getByRole('group', { name: 'Vertical align' }))
    expect(align.map(accessibleName)).toEqual(['Align left', 'Align center', 'Align right', 'Align justify'])
    expect(valign.map(accessibleName)).toEqual(['Vertical align top', 'Vertical align middle', 'Vertical align bottom'])
    expect([...align, ...valign].map(treatmentOf)).toEqual(Array.from({ length: 7 }, () => 'glyph'))
    // No visible word survives anywhere in either row — the defect was words
    // BESIDE icons at the same size, so an icon plus a caption would not fix it.
    expect([...align, ...valign].map(visibleText)).toEqual(Array.from({ length: 7 }, () => ''))
    // ⚠ AND THE SEVEN PICTURES ARE SEVEN DIFFERENT PICTURES. Everything above is
    // satisfied by a control drawing the same glyph three times: each segment
    // would still hold one `svg.segment-icon`, still show no text, still classify
    // `glyph`, still answer to its own name — and Vertical align would be
    // unreadable. `valignGlyphs` is a one-line record of three path strings,
    // which is precisely where a copy/paste slip lives; nothing else in this
    // suite would see it. Asserted over the path data, so two segments cannot
    // draw the same stroke.
    const paths = [...align, ...valign].map((segment) => segment.querySelector('svg.segment-icon path')?.getAttribute('d') ?? '')
    expect(paths.filter((data) => data === '')).toEqual([])
    expect(new Set(paths).size).toEqual(paths.length)
  })

  it('offers Align three ways when a table is in the selection, still in one vocabulary', () => {
    states[2]!.open()
    const align = controlsIn(screen.getByRole('group', { name: 'Align' }))
    expect(align.map(accessibleName)).toEqual(['Align left', 'Align center', 'Align right'])
    expect(align.map(treatmentOf)).toEqual(['glyph', 'glyph', 'glyph'])
  })
})
