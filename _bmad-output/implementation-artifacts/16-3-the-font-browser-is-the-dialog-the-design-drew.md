---
title: 'Story 16.3: The font browser is the dialog the design drew'
type: 'feature'
created: '2026-09-02'
status: 'done'
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

**The footer's face line, and the question it is the cheapest place to disclose.** `weightLine` states
one fact — one upright Regular per staged family, no bold and no italic — and deliberately names no
destination, because Story 16.5 inverts the destination (confirm will install rather than embed) and
destination language written here is language 16.5 must invert. It stays in `confirmLabel` and
`pendingLine`, which 16.5 revises in one place.

That fact has a consequence this story does **not** own and must not re-own. The property panel ships
Bold and Italic toggles (`src/App.tsx:1414`), while `SPEC.md:151` records the standing Non-goal — *"No
synthetic bold or oblique, and no variable-font axes. A weight is a face or it does not exist"* — and
`folio-go/internal/fontset/fontset.go:715` says bold *"when it arrives"* will be a `wght` instance
shipped as its own static face, in the future tense. Measured at this story's gate: all 31 catalogue
faces are `style: "Regular"` and **0 of 45** committed font binaries are bold, italic or oblique, so no
weighted face ships anywhere today. **Epic 16 therefore does not change the proportion — it was already
universal — it changes the population**, from 31 local families to ~1,305 offerable ones, roughly a 42x
growth in how often an author can set Bold on a family that has no bold face.

**`epics.md:521-525` already owns this**: bold and italic are *"not in this epic"* and the
realize-or-retire decision belongs to **Epic 11 (FR57)**, with SPEC-fonts recording the same question as
open. Nothing is registered here, because a second entry at a gate that does not own the question is the
two-authorities defect this epic has refused twice. It is recorded only that **this footer slot is the
cheapest place in the product to disclose the answer** should Epic 11 rule unfavourably — which is a
second reason the slot is built rather than deleted.

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

### Review patches applied (builder, 2026-09-03)

Step-04's three review layers returned no `intent_gap` and no `bad_spec`, so no loopback. All
findings were applied as patches. **Each fix was red-proofed** — reverted, watched fail, restored —
because a guard nobody has seen red is a guard nobody has tested. Unit tests went 641 → 659.

**Two of the review's own premises turned out to be wrong, and both are corrected rather than
carried:**

1. **D3 was priced as unreachable and is a live defect.** The review measured 0 of 1,811 *index*
   rows as thai-without-latin, which is right about the index and wrong about the browser: a
   **local-tier** row's coverage is read off the committed face, not off the snapshot, and two
   committed faces record `thai` alone — `Noto Sans Thai Looped` and `Noto Serif Thai`. The browser
   was printing **`Thai + Latin` beside two shipped faces whose own record claims no Latin**. Both
   are now asserted by name.
2. **F2's fix silently landed in the wrong function.** `setFontBrowserOpen(false)` was appended to
   the first `setTableEditor(undefined); setTableEditorError(undefined)` pair in the file — which is
   `closeTableEditor`, not `clearDocumentInteraction`. It compiled, and the red-proof was *also* red
   with the fix "in place", which is the only reason it was caught. Moved, and red-proofed again
   from the correct site.

**One defect the patches introduced and closed.** The new `App.test.tsx` cases open the real modal,
which starts up to four upstream probes per row. Left on the global `fetch` they outlived their own
test, and the next test's stub counted them as its own — 10 calls where one was expected, in a test
about an unrelated control. The describe now installs a **rejecting** fetch (which stops
`fetchWebFamily` at the first probe, so there is no second one to leak) and drains before restoring.

| Group | What changed |
|---|---|
| **A** | Stale prose corrected in three places: `App.css`'s shadow comment (the sheet elevation is DESIGN.md's third, transcribed, so the count stays three), the fidelity review's Recommendation (the token was minted; what remains is DW-178), and `font-index.ts`'s orphaned doc block. |
| **B** | The preview registry could strand a row on `Fetching…` for ever, and a comment claimed otherwise. The seam now reports **declines** as well as registrations (`onDeclined`), the derivation is checked **before** any fetch is started, and a `face.load()` rejection reaches the row instead of being swallowed. The comment that claimed a safety the code did not have is rewritten to describe what happens. |
| **C** | `confirm()` gained a `try`/`finally` and a per-family `catch` (a seam that *rejects* is now a named refusal, not a permanently disabled modal), clears the batch count when the batch ends, gates Escape on `!busy` like every other control, and names the unreachable missing-row branch instead of skipping it silently. |
| **D** | The elevation guard now asserts the minted token is **used**, not merely defined. `never` guards added to the sort arms and the results view. `scriptBadge` no longer claims Latin it cannot see. |
| **E** | The App seam had no coverage at all. Seven new cases across `App.test.tsx` and `App.font-store.test.tsx`: the door mounts the dialog, the document's chains reach it as `In template`, focus returns to the family control, a **stored** family's specimen renders with every fetch rejecting while a web row says it cannot be shown, and a refusal returned through the `'caller'` announcer is named inside the dialog with the dialog still open and no command sent. Each was red-proofed against the exact mutation the review named — including the dropped `offeredFamilies` second argument. |
| **F** | The Grid cap and the rail's readout now have one authority (`specimenSize` / `sizeReadout`), so the number printed is the number in use; the modal closes with the document; the door's sub-label is in its accessible name; focus returns to the family combobox on close; and the e2e specimen test addresses rows by name and asserts **every** row rather than the first. |

**Gates, run as separate commands.** `npm run test`: 659 passed, **1 failed** — `canvas-authority-contract.test.ts`, the pre-existing DW-152 red, unchanged. `npm run test:e2e:compile`: exit 0. `npm run build`: exit 0. `npm run lint`: clean. The committed e2e spec was re-run **in a real browser** (Chromium 1228) with the strengthened assertions: **6 passed**.

## Delivery Log

**Gates, measured locally on `e3f655b` by the builder — not relayed.** Nothing in this epic may be
called CI-verified: the designer CI job halts at step 2 on DW-152's standing red, and that repair lands
after Story 16.4.

| Gate | Result |
|---|---|
| `npm run test` | **659 passed, 1 failed** (54 files, 660 tests) |
| `npm run test:e2e:compile` | EXIT=0 |
| `npm run build` | EXIT=0 |
| `cd lint && go test -count=1 ./...` | EXIT=0, 4 packages |
| `cd folio-go && go test -count=1 ./...` | EXIT=1, 16 packages, 1 failure |
| Real browser (`chromium-1217`) | **6/6 passed**, EXIT=0, against the built bundle |

Both failures are **pre-existing, verified at `0e3a291` in an isolated worktree** rather than inferred:
`canvas-authority-contract.test.ts` (DW-152's standing red) and `TestCorpusMeetsP6ExerciseFloors/P6g`
(got 7, need >=20). The failing-test NAME SET is unchanged from baseline — diffed by name, not by total.
This story changed **0 Go files**.

**Two false greens were found in the measuring itself.** `PIPESTATUS` is bash and this shell is zsh, so
an exit-code capture printed empty while a piped `tail` reported success over a suite that had failed;
and `npm run test && npm run build` short-circuits, so the build never ran behind the standing red. Every
gate above was re-run as a separate command with no pipe.

**Deferrals: DW-176 DISCHARGED. DW-161 stays OPEN and its own text is why** — its ordering clause binds
it to DW-101 ("the specs exist and CI never runs them"), which is open until the post-16.4 CI repair. Of
its three cases, two ran against the real host and the licence case was **simulated at the network
boundary because it is impossible, not skipped**: no CC-BY-SA family is reachable upstream, and a random
sample of **60 of the 1,273** addable families resolved 60/60 with every token already in the closed table
(`OFL` 58, `APACHE2` 1, `UFL` 1). **DW-177, DW-178 and DW-179 registered.**

### The pattern behind three of this story's defects: the thing was changed and its explanation was not

`App.css`'s comment still said the modal *"takes the page's, which is the only shadow token the
stylesheet has"* while sitting directly above the line that had been changed to `--shadow-sheet`; the
token-fidelity review's row recorded the token as minted while its own Recommendation still said it was
not this story's to add; and a doc block was left orphaned from the constant it describes by an insertion
placed between them. **The most serious code finding is the same species one level down:**
`embedded-face-registry.ts` claimed a declined derivation *"takes the same degrade a fetch returning no
bytes already takes"* when it did not — **a comment claiming a safety the code does not have, found
inside the guard built for that exact defect.** Third instance in one story.

### What the red-proofs bought, which is not what they are usually credited with

F2's fix — resetting `fontBrowserOpen` when the document is replaced — **appended to the first matching
pair in the file and landed in `closeTableEditor` instead of `clearDocumentInteraction`.** It compiled
cleanly, the suite was green, and **its red-proof came back red WITH the fix supposedly in place.** No
reviewer reading that diff would have seen it. It was caught only because the mutation was run *before*
the fix rather than after, and the proof refused to change colour. A red-proof's value here was not
*"the test is real"* but ***"the fix is where you think it is."***

### A measurement over the wrong population, three times

`indexCategories`/`indexScripts` derived from `familyIndex` (1,811) while the chips filter the offered
population (1,273 web + 31 local); a comment cited Handwriting at **337** when the browser offers **259**;
and the `'Thai + Latin'` badge was called unreachable on the strength of *"0 of 1,811 rows are
thai-without-latin"* — a measurement over the **index**, offered as evidence about the **browser**, which
also carries 31 local faces that are not in the index at all. **Two of them, `Noto Sans Thai Looped` and
`Noto Serif Thai`, record `thai` alone**, so the badge was claiming Latin coverage beside two shipped
faces whose own record denies it. It was a live user-visible defect, not a stylistic tidy. The rule this
story pays for: **when two populations differ by a filter, every claim about one is silent about the
other, and the silence is invisible at the point of use.**

### Three re-proofs run at the close, by deletion, by the builder

| # | Mutation | Result |
|---|---|---|
| 1 | `return refuse(outcome.reason)` -> `refuse(...); return undefined` (both sites) | **RED** — `keeps the modal open, names the family, and sends no command` |
| 2 | `App.css` `.font-browser` `--shadow-sheet` -> `--shadow-page` | **RED** — the elevation routing case |
| 3a | delete the pre-fetch decline guard in `show()` | **RED** — `never fetches for a family whose name the derivation declines...` |
| 3b | delete the load-rejection `onDeclined` in the seam | **RED** — 2 tests, incl. `reports a face whose bytes will not parse as unavailable, never as still fetching` |

**The seam test renders the real `App` and clicks the real `Add fonts…` button** — it is not a stub one
level down. `App.font-store.test.tsx` and `App.test.tsx` both mount `<App />`, and the new cases address
the door by its accessible name and assert the real dialog.

**One re-proof initially reported a false pass and was re-run.** A mutation regex matched nothing, so the
file was never modified and the resulting green was a non-measurement rather than evidence. Re-targeted at
the real site — the decline had moved into the seam behind a new `onDeclined` parameter — it reddened two
tests. **A mutation that does not apply proves nothing, and its green looks exactly like a passing proof.**

## Suggested Review Order

**The one genuinely new mechanism: a third face lifetime**

- The argument for a third lifetime, and why it cannot collide with the other two
  [`preview-face-family.ts:40`](../../folio-designer/src/preview-face-family.ts#L40)

- The seam widened to take the derivation as a parameter, defaulting to the document's
  [`embedded-face-registry.ts:94`](../../folio-designer/src/embedded-face-registry.ts#L94)

- The decline is checked BEFORE any byte is fetched, so a row cannot sit on `preparing` for ever
  [`preview-face-registry.ts:124`](../../folio-designer/src/preview-face-registry.ts#L124)

- The exact-list census gains one anchored position and four near-miss red-proofs
  [`canvas-font-stack.test.ts:269`](../../folio-designer/src/canvas-font-stack.test.ts#L269)

**Saying only what the data supports**

- Chip vocabularies derived from the offered population, not the wider index
  [`font-index.ts:335`](../../folio-designer/src/font-index.ts#L335)

- The footer states one fact about what a face is, and names no destination
  [`font-browser-model.ts:316`](../../folio-designer/src/font-browser-model.ts#L316)

**The confirm batch and its refusals**

- One named seam per staged family; refusals per family, the rest proceed
  [`FontBrowser.tsx:137`](../../folio-designer/src/FontBrowser.tsx#L137)

- Specimen bytes resolve through the same three tiers a pick does; the store is read, never written
  [`App.tsx:403`](../../folio-designer/src/App.tsx#L403)

**The elevation, transcribed rather than invented**

- DESIGN.md's third declared elevation, implemented so the count stays three
  [`tokens.css:15`](../../folio-designer/src/tokens.css#L15)

- The guard asserts both the token's value and its use, from DESIGN.md's own text
  [`design-contract.test.ts:52`](../../folio-designer/src/design-contract.test.ts#L52)

- The surface that consumes it
  [`App.css:380`](../../folio-designer/src/App.css#L380)

**Supporting**

- The App seam under a real `App` render, through the real door
  [`App.font-store.test.tsx:203`](../../folio-designer/src/App.font-store.test.tsx#L203)

- The executed browser witness, six cases
  [`font-browser.spec.ts:38`](../../folio-designer/e2e/font-browser.spec.ts#L38)
