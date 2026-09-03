---
title: 'Story 16.3: The font browser is the dialog the design drew'
type: 'feature'
created: '2026-09-02'
status: 'in-progress'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: '0e3a2913d96e1cac594fe76f743ea8eed8064592'
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

**Problem:** With D-16.1 the author can reach every addable family in the snapshot, and the family combobox — a flat list
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
- **The header's family count REUSES `familyIndexDisclosure()`, which already ships the correct
  sentence.** D-16.3: the index is a build-time snapshot, and D-16.R.2's consequence is that the
  browser's count is the **addable** count, not the published one. Story 16.1 already shipped
  `familyIndexDisclosure()` (`font-index.ts:282`) stating all of it in one sentence, derived from
  `addableFamilyCount`. **Do not mint a second sentence and do not hardcode a number** — two
  authorities on one count is the defect D-16.R.13 refused on `source` and D-16.R.6 refused on licence.
  The design's *"web font library · 1,946 families"* becomes that derived line.
- **Specimen rendering is bounded.** Over a thousand rows, each wanting a face, is a fetch storm and a
  memory problem. Register on demand — what is on screen — and say in code what the bound is.
- **The staged set is UI state and nothing else.** No partial document, no uncommitted buffer, no
  second document model (AD-15). Confirm dispatches one command per staged family through a **single
  named seam**, each its own history entry. **The seam's ACTION is provisional (D-16.R.46 Q2):** today it
  embeds, as Story 8.6 built; Story 16.5 changes it to install, moving the embed to first use. Build the
  seam as one function so 16.5 swaps a body rather than a state machine.
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
| Type in search | `sara` | Snapshot filtered by **family and category only** — the committed snapshot carries no designer field (D-16.R.33 R3) | — |
| Script + category chips | Thai + Serif | Intersection; `reset filters` visible exactly while a filter is active | — |
| Sort | Trending / A–Z | Ordered from snapshot fields — `popularity` and `family`. **No "Most styles" arm** (D-16.R.33 R3): this product embeds exactly one face per family (the upright Regular at weight 400, `font-source.ts:197`; `:314` refuses a family publishing none), so style count is not a difference the author can act on | — |
| Preview text and size | Thai text, 34px | Every specimen re-set; the Thai toggle switches the default sample | — |
| Stage several, confirm | 3 staged | 3 dispatches through the named seam, 3 history entries, progress stated. **Provisional per D-16.R.46 Q2** — 16.5 changes the seam's action to install | Per-family failure named; others proceed |
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
- `src/App.tsx:1611` — `FontFamilyProperty`, and its `openBrowser` seam in Story 16.4. **Re-verified at `9c2fbe6`.**
- `src/App.tsx:232-263` — the document-scoped face registration (**re-verified at `9c2fbe6`**), with its explicit reasoning about
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
- `src/App.tsx` — `Add fonts…` opens the browser from the existing family control. **No keyboard
  shortcut in this epic — the omission is hereby RULED and recorded** (D-16.R.33 R2, owner-confirmed):
  `⌘G` is the browser's Find Next, and this application's convention puts conventional document actions
  on Command (`⌘S`, `⌘Z`) and **app-specific actions on Option** (`⌥P`, `⌥S`). **No hint glyph is
  rendered**, because a `⌘G` label beside a key that does nothing is a false UI string. `⌥F` is named as
  its eventual shape; `src/shortcuts.ts` is untouched.
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

## Spec Change Log

### 2026-09-03 — applied at the plan gate (orchestrator)

**Contract amendments.** This spec was approved at a previous session's CHECKPOINT 1, so
`<intent-contract>` was locked. It is reopened deliberately here and the amendments recorded rather than
edited in silently:

1. **The *Most styles* sort arm is dropped** and the search predicate narrowed to family + category
   (D-16.R.33 R3). Measured: this product embeds exactly one face per family, so style count sorts on a
   difference the product erases; and `designer` exists in neither the generated module nor the raw
   snapshot, and obtaining it means a regeneration that breaks the `d6d51f1` pin and re-opens 16.1a's
   batch. The payload cost of adding a style count was **measured first** (+1,326 brotli bytes, ~0.008%
   of first load) so the record shows an **affordable decline made on the criterion**, not a budget one.
2. **The family count reuses `familyIndexDisclosure()`** instead of the three `~1,946` statements
   (D-16.R.2's consequence). The published figure is not the addable one, and Story 16.1 already ships
   the correct derived sentence.
3. **Confirm's action is marked PROVISIONAL** pending Story 16.5 (D-16.R.46 Q2). Recorded inside the
   contract because the Matrix row and the Boundary both asserted the embed mechanism, and shipping an
   AC that restates a mechanism under revision is how a provisional decision gets cited later as settled
   direction.

**Outside the contract.** Code Map anchors re-verified at `9c2fbe6` and corrected — all three designer
anchors had moved when Story 16.2 landed, which is why they were deliberately not corrected three
stories early (D-16.R.28). `⌘G` ruled out with its reasoning (D-16.R.33 R2, owner-confirmed). The browser
run scoped to discharge DW-161 and DW-176. `-count=1` written into the Go commands (DW-168).

### 2026-09-03 — corrected at the build dispatch (builder)

**`baseline_commit` corrected from `a40c34d` to `0e3a291`.** The recorded value was written when this
spec was created (2026-09-02) and is **52 commits stale**: `a40c34d` is *"Close Epics 9 and 10 at a
boundary gate"*, which **precedes the whole of Epic 16**. It therefore precedes this story's own
`Block If` preconditions — 16.0, 16.1, 16.1b, 16.1a and 16.2 all closed after it — so a review diff
taken from it would have swept 5,900 insertions of five other closed stories into this story's triage.
Step-03's preserve rule is conditioned on `(resumed run)`; this spec never reached `in-progress`.
`Block If` measured at this gate: all three named dependency stories are `done` in `sprint-status.yaml`.
Code Map anchors re-verified at `0e3a291`: all three designer anchors still hold (`App.tsx:232-263`
registration, `pickCatalogueFamily` `:778`, `FontFamilyProperty` `:1611`) — **none moved again** since
the plan gate's `9c2fbe6` re-verification.

## Verification

- `cd folio-designer && npm run test && npm run test:e2e:compile && npm run build`
- `cd lint && go test -count=1 ./...`
- `cd folio-go && go test -count=1 ./...`
  **`-count=1` is mandatory** (DW-168, narrowed by D-16.R.31): CI already passes it everywhere, so the
  live residue is exactly this by-hand path — and this story touches no Go, the condition under which a
  filesystem-walking Go test replays a stale green.
- A browser run: search, filter, sort, stage three, confirm, and one deliberate upstream failure.
  **This run also discharges DW-161 and DW-176** (D-16.R.42), because the container is already being
  paid for by this story's own cadence override. Four extra cases, each cheap:
  1. a pick with the network **up**, against the real hosts;
  2. a pick with the network **down**;
  3. a pick whose licence is **outside the allowlist**;
  4. **a stored face survives a reload and is offered with the network disabled** — the browser-witness
     residual 16.2 created (DW-176). This run already reloads to exercise the offline path, so this
     costs a assertion rather than a setup.
  **This is the first real-browser exercise of BOTH the web tier and the store**, which is how they are
  actually used together.
- Token fidelity checked against `DESIGN.md`, in the form `review-token-fidelity.md` already uses.

### Verification as run (builder, 2026-09-03)

| Gate | Result |
|---|---|
| `npm run test` (vitest, 54 files) | **634 passed, 1 failed.** The failure is `canvas-authority-contract.test.ts` reporting `e2e/e9-5-border-no-ink.spec.ts: getComputedStyle` and is **PRE-EXISTING at `0e3a291`** — reproduced by stashing this story's whole diff and re-running. It is Story 9.5's browser-witness assertion (a `page.evaluate` reading what the canvas painted) caught by a scan whose e2e corpus has no carve-out for it. Not repaired here: amending a canvas-authority guard for an unrelated story is the drive-by shape that guard's own comments deprecate. **Raised for a ruling.** |
| `npm run test:e2e:compile` | pass |
| `npm run build` | pass (host scans, wasm, `tsc -b`, vite, offline bundle, offline verify) |
| `cd lint && go test -count=1 ./...` | pass (4 packages) |
| `cd folio-go && go test -count=1 ./...` | **1 failed:** `internal/text` `TestCorpusMeetsP6ExerciseFloors/P6g (opaque names)` — got 7, need >=20. **PRE-EXISTING at `0e3a291`**, reproduced the same way. This story touches no Go. **Raised for a ruling.** |

**Real-browser run — PERFORMED, not skipped.** Chromium (Playwright, `chromium-1228`) against the
built bundle on `127.0.0.1:4173`. The committed `e2e/font-browser.spec.ts` — 6 tests — **passed in
the browser**, not only compile-checked, which discharges Story 16.0's compile-only finding for this
story's own spec.

A second, **temporary and uncommitted** witness spec then ran the cadence's six cases against the
**real upstream host**. All six passed; the file was deleted afterwards.

| # | Case | Outcome |
|---|---|---|
| 1 | Search, filter, sort, stage three (Kanit, Prompt, Mitr), confirm — network **up**, real hosts | Modal closed; document went from revision 2 to **revision 5** — one history entry per family — and all three appeared in the family control's own listbox. |
| 2 | A pick with the network **up** (DW-161) | Covered by case 1: three real fetches, three embeds. |
| 3 | One deliberate upstream failure inside a batch | Prompt's four `METADATA.pb` probes forced to 404. The other two embedded (revision 4), the modal stayed open, and the refusal was named **against Prompt** — *"Prompt is in this designer's snapshot of the family list but is no longer published upstream…"* — with only Prompt left staged. |
| 4 | A pick whose licence is **outside the allowlist** (DW-161) | `METADATA.pb` served with `license: "CC-BY-SA"`. Refused, unembedded, revision unmoved: *"Kanit cannot be added: its licence is published as \"CC-BY-SA\", and ShareAlike is a copyleft term…"*. Simulated at the network boundary because **no CC-BY-SA family is reachable**: `cc-by-sa/` upstream now holds only `knowledge`, whose `METADATA.pb` 404s, and no such family is in the snapshot. |
| 5 | A pick with the network **down** (DW-161) | `setOffline(true)` before the family was ever touched. Refused with the offline sentence and its second clause: *"…the faces this machine already holds are still offered."* |
| 6 | **A stored face survives a reload and is offered with the network disabled** (DW-176) | Kanit added with the network up, page reloaded (document back to revision 1), network disabled, browser reopened: the row read **`downloaded to this machine`**, its **specimen rendered from the stored bytes with no network**, and confirming embedded it — revision 3 — with the network still down. **First real-browser witness of IndexedDB persistence and of the store and the web tier used together.** |

**One finding raised by case 5, in Story 16.1's code and deliberately not repaired here.** On the
first attempt, the offline pick was reported as *"Kanit declares OFL-1.1 but publishes no OFL.txt
beside its face"* — because `METADATA.pb` came back from the browser's HTTP cache while `OFL.txt` did
not, and `readText` distinguishes only a **stall** from a failure, never an offline failure from a
missing file. That is precisely the "sends the reader upstream to look for a file that is sitting
there" hazard `font-source.ts`'s own stall comment describes, one branch over. Reachable whenever the
cache is partially warm. **Not fixed in this story** — it is `fetchWebFamily`'s, and repairing a
refusal message in the fetch module from a browser story is how a fix lands without a test that
would notice it regressing.
