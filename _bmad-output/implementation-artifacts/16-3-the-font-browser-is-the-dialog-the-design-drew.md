---
title: 'Story 16.3: The font browser is the dialog the design drew'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md'
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Choosing a typeface from a list of names is guessing. This is the screen that stops it: every family
is shown set in itself, in your own words, at the size you are going to print. Type a customer's name
in Thai and you see which faces can actually set it before you commit to one.

Around that: a search, filters for writing system and category, a sort, and a way to line up several
families and add them in one go — with a footer that tells you what is about to go into your file.

<intent-contract>

## Intent

**Problem:** With D-16.1 the author can reach ~1,946 families, and the family combobox — a flat list
of names in the panel's own typeface — is not a way to choose among them. `Font Browser.dc.html` draws
the screen that is; the product has no equivalent of it.

**Approach:** Build the design's modal over Story 16.1's snapshot and Story 16.2's store: specimens
rendered in each family at an author-set size in author-set text, script and category chips, sort, Row
and Grid views, staged multi-select, and a footer that states what will be embedded before it is.

## Boundaries & Constraints

**Always:**
- **A specimen must be set in its own family**, which means each rendered family's face is registered
  for **preview** independently of any document embedding it. Today the browser registers a CSS family
  only for faces the *document* carries (`App.tsx:186-224`); that is the wrong lifetime here and a
  third one has to be argued, not assumed.
- **Preview registration must not create a second authority on what a document contains.** A face
  registered for preview is not embedded, is not in `fontFamilies`, and must never appear as if it
  were.
- **The header's family count is the SNAPSHOT's count and says so.** D-16.3: the index is a build-time
  snapshot. The design's *"web font library · 1,946 families"* becomes a line that does not claim to
  be live.
- **Specimen rendering is bounded.** ~1,946 rows, each wanting a face, is a fetch storm and a memory
  problem. Register on demand — what is on screen — and say in code what the bound is.
- **The staged set is UI state and nothing else.** No partial document, no uncommitted buffer, no
  second document model (AD-15). Confirm dispatches the same one-command-per-family embed Story 8.6
  built, each its own history entry.
- **Every refusal is per family and named.** One family failing to fetch does not abandon the others.
- **Keyboard and assistive technology are contract, not polish** (UX-DR25): focus trapped, Escape
  closes, every chip and toggle a real named control, staged state announced.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **Stories 16.0, 16.1 and 16.2 are not closed.** This story renders their output; it cannot be
  verified over a boundary that throws or a source that does not exist.
- **A Cancel/Apply buffer would appear anywhere else in the panel.** The staged set is scoped to this
  modal because the modal is where the design puts it; it is not a precedent for the inspector, which
  AD-15 governs.
- **Fidelity would be claimed against tokens the design system does not define.** The mockup's hex
  values are checked against `DESIGN.md`'s tokens; a colour with no token is a gap to raise, not a
  literal to paste.

**Never:** a second document model · an OS font list · rendering a specimen in a fallback face while
implying it is the family · blocking the UI on a fetch.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open from `Add fonts…` | Dropdown open | Modal, focus trapped, search focused | — |
| Type in search | `sara` | Snapshot filtered by family and designer, as the mockup's own predicate does | — |
| Script + category chips | Thai + Serif | Intersection; `reset filters` visible exactly while a filter is active | — |
| Sort | Trending / A–Z / Most styles | Ordered from snapshot fields | — |
| Preview text and size | Thai text, 34px | Every specimen re-set; the Thai toggle switches the default sample | — |
| Stage several, confirm | 3 staged | 3 embeds, 3 history entries, progress stated | Per-family failure named; others proceed |
| A family fails to fetch | Upstream 404 | That row reports it; the rest are added | Named refusal |
| No results | Query matches nothing | The design's empty state, naming the query | — |
| Grid view | Toggle | The design's 3-column cards | — |
| Escape / Cancel | Modal open, families staged | Closes, stages discarded, document untouched | — |
| Offline | No network | Stated; stored families still addable (16.2) | Degradation |

</intent-contract>

## Code Map

**Design**
- `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html` — the
  whole modal. Its `renderVals()` is the behavioural spec: `specimenFor`, the filter predicate, the
  sort arms, `chip()`, `buttonLabel` states (`+ Add` / `✓ Added` / `In template`), `pendingLine`,
  `weightLine`, `confirmLabel`, `resultLine`. **Its `FAMILIES` (14) and `SYSTEM_FONTS` (6) are
  placeholder data, not requirements** — 14 stands in for the snapshot, and `SYSTEM_FONTS` is
  superseded by D-16.2.
- `.../DESIGN.md` — the token file fidelity is judged against.

**Designer (`folio-designer/`)**
- `src/App.tsx:1296-1380` — `FontFamilyProperty`, and its `openBrowser` seam in Story 16.4.
- `src/App.tsx:186-224` — the document-scoped face registration, with its explicit reasoning about
  lifetimes. **The preview registration is a sibling of this, not a change to it.**
- `src/embedded-face-registry.ts` — `document.fonts` registration; **name-keyed and global**, which is
  why a preview lifetime must not collide with the document one.
- `src/App.css:210-223` — the combobox listbox styling the modal's rows should not silently diverge
  from.
- `src/FontChainEditor.tsx` — the closest existing multi-control panel; the shape to match for
  keyboard behaviour and refusal anchoring.
- `e2e/component-properties.spec.ts:35-60` — how the family control is addressed by accessible name.
  **Those names must not move**; the browser adds names, it does not rename.

## Tasks & Acceptance

**Execution:**
- `src/` — a new modal component rendering the design's five regions: header, rail, results, empty
  state, footer.
- `src/` — a **preview face registry** with its own lifetime, registering on-demand for visible rows
  and releasing what scrolls away, with the bound written down.
- `src/` — the filter, sort and staging logic, ported from the mockup's `renderVals()` rather than
  re-invented, so the design and the product agree on edge cases.
- `src/` — confirm: dispatch one embed per staged family through Story 16.1's path, reporting progress
  and per-family refusals.
- `src/App.tsx` — `Add fonts…` opens it; `⌘G` opens it, or the omission is ruled and recorded.
- `src/App.css` — styles in the existing token vocabulary; any mockup colour without a token is raised.
- Tests: filter/sort/staging as units; the modal's accessible names; focus trap and Escape; a staged
  confirm producing N history entries; one failing family not abandoning the rest.
- e2e: a compile-checked spec, and — per Story 16.0's finding about compile-only e2e — an actual
  browser run recorded, or a plain statement that none was performed.

**Acceptance Criteria:**
- Given the browser, when it opens, then it is the design's modal with its five regions and its states.
- Given each result, when rendered, then its specimen is set in that family, at the chosen size, in the
  chosen text, and never silently in a fallback.
- Given the Thai toggle and a Thai-covering family, when it is on, then the specimen is the Thai sample
  and the family carries the `Thai + Latin` badge.
- Given filters and sort, when used, then results and `resultLine` follow, and `reset filters` appears
  exactly while a filter is active.
- Given several staged families, when confirmed, then each is embedded as its own history entry, the
  footer states progress, and a failure is named per family.
- Given keyboard-only operation, when the modal is open, then focus is trapped, Escape closes, and
  every control has a name.
- Given the header count, when it is shown, then it names the snapshot rather than claiming a live
  library.

## Design Notes

**The one genuinely new mechanism is preview registration, and it is where this story can go wrong.**
`document.fonts` is a global, name-keyed registry — the comment at `App.tsx:186-224` says so and says
why the document effect is scoped the way it is. A preview registry that registers under the same
family names would let a modal's scroll position affect what the canvas paints. The two must not be
able to collide, and a test should prove they cannot.

**Why the mockup's own logic is ported rather than reimplemented.** `renderVals()` already settles a
dozen edge cases — what the button says when a family is in the template versus staged, what the
footer says at zero, which sample text a Thai family gets. Re-deriving those from the screenshot is how
a product drifts from its design while everyone believes it matches.

## Verification

- `cd folio-designer && npm run test && npm run test:e2e:compile && npm run build`
- A browser run: search, filter, sort, stage three, confirm, and one deliberate upstream failure.
- Token fidelity checked against `DESIGN.md`, in the form `review-token-fidelity.md` already uses.
