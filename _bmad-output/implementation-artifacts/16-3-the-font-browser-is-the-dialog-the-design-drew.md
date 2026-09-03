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

## PENDING GATE AMENDMENTS — read before implementing (added 2026-09-03 by the orchestrator)

*Non-normative. These are RULED changes to this spec, recorded here so they are not lost, but NOT yet
applied. They are applied at this story's own plan gate, not now — see the reason below.*

**Why they are not applied yet.** This spec was planned at `baseline_commit: a40c34d` and approved
before 16.0, 16.1, 16.1b and 16.1a landed. Its `## Code Map` line anchors are therefore stale by
construction (D-16.R.28), and Story 16.2 lands before this one, which will rot them again. Re-verifying
them now would be work done twice and trusted once. Anchors are re-verified at this story's gate,
against the tree as it stands then.

**Anchors known stale at `8a9e297`** (indicative, re-measure at the gate — do not trust these numbers
either): the document-scoped face registration cited as `:186-224` is at `:232-243`; `pickCatalogueFamily`
cited as `:608-627` is at `:660`; `FontFamilyProperty` cited as `:1296-1380` / `:1284-1380` begins at
`:1385`; the disk-font decline cited as `:1366` is at `:1474`. `App.css:210-223` and
`e2e/component-properties.spec.ts:35-60` were checked and are still correct.

**Ruled changes to apply at the gate (D-16.R.33):**

1. **R3 — drop the *Most styles* sort arm, and drop `designer` from the search predicate.** This
   product embeds **exactly one face per family** (the upright Regular at weight 400 —
   `font-source.ts:197`, and `:314` refuses a family that publishes none), so style count is not a
   difference the author can act on; a family with eighteen styles and one with a single style deliver
   the identical thing. And `designer` is not a field addition but a **snapshot regeneration**: the
   committed `font-index.json` carries no designer field, obtaining it means re-running
   `refresh:font-index`, and that breaks the `d6d51f1` pin (D-16.R.23), fires DW-166 trigger 1, and per
   D-16.R.26 inverts Story 16.1a's landed batch on exactly the axis this epic has already litigated
   twice. **The payload was measured first so this is a priced decline, not a budget one:** `+1,326`
   brotli bytes for a style count, ~0.008% of a 15,729,262-byte first load. **Affordable, and declined
   on the criterion.** Do not reverse it when 15.0 frees room.
2. **R3 — the header's family count reuses `familyIndexDisclosure()`, which already ships the correct
   sentence.** This spec still says *"~1,946 families"* in three places; D-16.R.2's own consequence is
   that the browser's count is the **addable** count. Measured at `8a9e297`: 1,811 snapshot rows
   (1,946 published, 135 CJK excluded), 1,273 web-addable, 31 local — **1,304**. Two authorities on one
   count is the defect D-16.R.13 refused on `source` and D-16.R.6 refused on licence.
3. **R2 — `Add fonts…` is this story's entry point, and there is NO keyboard shortcut in this epic.**
   The spec's *"`⌘G` opens it, or the omission is ruled and recorded"* is hereby the omission, ruled:
   `⌘G` is the browser's Find Next, and this application's own convention (`src/shortcuts.ts`) puts
   conventional document actions on Command and **app-specific actions on Option** (`⌥P`, `⌥S`). No
   hint glyph is rendered, because a `⌘G` label beside a key that does nothing is the false-UI-string
   class this epic has ruled against three times. Registered deferred with `⌥F` named as its shape.
4. **The browser run also discharges DW-161** — a pick with the network up, a pick with it down, and a
   pick whose licence is outside the allowlist, against the real hosts. The container is already being
   paid for by this story's own override; scoping it wider costs almost nothing and converts the
   epic's largest evidence gap from a note into a measurement.

**Items 1 and 2 are CONTRACT AMENDMENTS.** This spec's `<intent-contract>` spans lines 27-91, and the
Sort row (`:82`) and the `~1,946` statements (`:31`, `:50`, `:52`) are **inside** it, on a spec already
at `ready-for-dev`. They are reopened deliberately at the gate and recorded as amendments in a
`## Spec Change Log`, not edited in silently. Item 3 touches Tasks only; item 4 touches Verification only.

## Verification

- `cd folio-designer && npm run test && npm run test:e2e:compile && npm run build`
- `cd lint && go test -count=1 ./...`
- `cd folio-go && go test -count=1 ./...`
  **`-count=1` is mandatory** (DW-168, narrowed by D-16.R.31): CI already passes it everywhere, so the
  live residue is exactly this by-hand path — and this story touches no Go, the condition under which a
  filesystem-walking Go test replays a stale green.
- A browser run: search, filter, sort, stage three, confirm, and one deliberate upstream failure.
- Token fidelity checked against `DESIGN.md`, in the form `review-token-fidelity.md` already uses.
