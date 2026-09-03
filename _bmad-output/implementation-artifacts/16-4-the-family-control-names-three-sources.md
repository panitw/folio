---
title: 'Story 16.4: The family control names three sources'
type: 'feature'
created: '2026-09-02'
status: 'in-review'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'b8431c4e9d743cc2cde314c98688d393bfbbe828'
context:
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: []
deferred:
  - summary: >-
      epics.md contradicts five shipped stories: Story 16.5 has no entry at all, the file ends
      mid-Epic-16, and Story 16.4 AC1/AC2/AC5 still describe ADDED FROM WEB FONTS and embed-on-pick.
    evidence: >-
      Measured at b8431c4. Superseded by D-16.R.46 and D-16.R.72. Filed rather than noted because a
      note in this spec is read by whoever reads this spec, and the person harmed is whoever reads
      epics.md believing it. Owner: the post-16.4 infrastructure item beside DW-171/177/179/180/182.
  - summary: >-
      A pick from AVAILABLE TO INSTALL blocks up to 30 s on a stall and up to 180 s against a
      slow-but-alive host, with nothing bounding, reporting or cancelling the wait.
    evidence: >-
      Registered by Story 16.5 gated to 16.4, with a remedy D-16.R.72 forecloses: the web arm keeps
      its heading inside the control, so the rows and the pick both stay. The ruling supplied
      legibility, not a bound. Re-measured: fetchTimeoutMs = 30_000 armed PER REQUEST in timedFetcher
      (font-source.ts:313,343) across up to six call sites, so 16.5's recorded 30 s is the stall half.
      Owner: NOT Story 16.4 - the post-16.4 infrastructure item, or whoever next opens fetchWebFamily.
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The font menu has to answer a question it currently cannot: where is this typeface, and what happens to
my file if I pick it? Three answers, so three groups — already in this file; on this machine but not in
this file; not on this machine at all. A font moves between them because you did something, never on
its own.

Two things underneath it turn out to be broken. The list that feeds the menu says in its own
documentation that it returns the on-this-machine faces together, and it does not — a face you
downloaded sits wherever the web list happened to put it. And the menu only ever draws the first fifty
rows, which was harmless when they all sat under one heading and is not harmless under a heading that
promises a font is on your machine. Both are fixed here, because a heading that lies is worse than no
heading.

There is also an old sentence at the bottom of the menu saying a typeface from your own disk cannot be
embedded. It was true for a good reason, and the reason has moved twice since. This story does not
quietly carry it forward — it re-derives it and writes down the working.

<intent-contract>

## Intent

**Problem:** `FontFamilyProperty` draws two groups — the document's chains, and everything addable.
Epic 16 created three relationships to a font, and 16.5 made *installing* and *embedding* different
acts. The control also carries a disk-font decline derived from a premise D-16.1 reversed, a listbox
that breaks its required-owned-elements rule six times, and an ordering defect that makes any grouping
over its flat list draw a false heading.

**Approach:** Re-cut the control into the design's three groups on the axis the code already forks on;
repair `offeredFamilies` so the union arrives in its own documented order; partition before capping;
fix the listbox defect here rather than multiplying it; re-derive the disk-font decline against three
premises; and give the design's borrowed label back to the control it belongs to.

## Boundaries & Constraints

**Always:**
- **Three groups mean three relationships, and the axis is *where are the bytes*, never *when did it
  arrive*.**
  1. `IN THIS TEMPLATE` — the document's declared chains (`families.includes(name)`). **The bytes are
     in the file.**
  2. `AVAILABLE LOCALLY` — `familyIsInstalled(source)` is true: the `local` **and** `stored` arms.
     **On this machine, not in this file.** Picking embeds and commits the property in one gesture,
     two commands, no network.
  3. `AVAILABLE TO INSTALL` — the `web` arm. **Not on this machine.** Picking installs; nothing enters
     the file.
- **A row's group is a pure function of (declared?, `familyIsInstalled`?)** — never of a session-scoped set. The `local`/`stored` split stays invisible in the grouping: it is a provenance difference with
  no consequence at the moment of choosing, and surfacing it would be the fourth group this spec
  forbids.
- **A font changes group because the author acted.** Story 8.6's rule, carried in its own code comment
  at `App.tsx:1901-1902` — *"nothing says 'added', the entry simply moves"* — extends to three groups
  unchanged. Install moves a row 3 → 2; first use moves it 2 → 1.
- **The fork at `choose()` keeps its shape.** Two decisions, two undos; fusing them was refused at 8.6,
  refused again at 16.5, and is refused here.
- **A heading tells the truth about every row it could own.** No group may be capped, filtered or
  re-ordered in a way that leaves a member of it undrawn while its heading is shown.
- **The accessible names Playwright addresses must not move.** `combobox` "Font family", "Edit font
  chains", the listbox's own name "Fonts".
- **The listbox is fixed, not extended.** Use `role="group"` with `aria-label`, or move the notes out of the `<ul>` and reference them with `aria-describedby`.
- **The disk-font decline is re-derived or lifted, never carried unread**, and the reasoning is
  recorded in the code comment, which is where 8.6 put the last one.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Any change that would fuse the embed and the property commit into one command.
- Any fourth group, or any grouping key other than (declared?, `familyIsInstalled`?).

**Block If:**
- **An accessible name in the e2e specs would move.**
- **The control would offer a font the engine will refuse** — Story 14.4's rule, applied here.
- **A heading would be drawn over a group whose membership the render cannot show in full.**

**Never:** a host font in any group · a group whose membership changes without an author action · a
fourth group · a session-scoped set anywhere in the grouping.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open on a fresh machine | Empty store, 1 chain | All three groups drawn: `IN THIS TEMPLATE`, a **populated** `AVAILABLE LOCALLY` (the 31 committed faces are always present), `AVAILABLE TO INSTALL`; `Add fonts…` present | — |
| Open with a populated store | 4 stored, 2 in document | Three groups, membership disjoint and complete; stored rows sit under `AVAILABLE LOCALLY` regardless of their index position | — |
| Pick from `AVAILABLE LOCALLY` | Stored or local face | Embedded from the machine with no fetch, then the property committed — **two commands, two undos**; row moves 2 → 1 | Engine refusal anchored at the control; the property commit is suppressed |
| Pick from `AVAILABLE TO INSTALL` | Web-tier family | Face installed. **No engine command, no property commit.** Row moves 3 → 2 | Refusal anchored at the control |
| Pick from `AVAILABLE TO INSTALL`, store unavailable | No IndexedDB | Degrades to embed-on-pick; row moves 3 → 1; `storeUnavailableEmbedNote` says so | — |
| Pick a declared chain | Name in the document | `fontFamily` property commit, as today | Existing refusal path |
| More than 50 web families offered | Empty query, 1,273 web rows | Groups 1 and 2 render **in full**; the 50-row cap applies to `AVAILABLE TO INSTALL` alone; the *"Showing N of M"* line sits under that group and names **its** population, not the union's | — |
| A stored face ranked deep in the web list | 1 stored family at web index ~900, 32 installed rows against a 50-row cap | It draws under `AVAILABLE LOCALLY` and **is not dropped by the cap**; all 32 installed rows render | This is the regression the repair exists for: a heading is never drawn over a group it cannot show in full |
| Filter field | `sara` | All groups filtered; a heading is suppressed **only when its own group is empty after filtering** | — |
| Add fonts… | Click | Opens the browser (16.3). **No `⌘G`, no hint glyph** | — |
| Mixed selection | Two components, different families | `Mixed` placeholder, as today | — |
| Keyboard walk | Arrow keys | One linear walk across all three groups in one sequence | — |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `604ae8f` and re-confirmed at `b8431c4`** — the diff between
them is one file, `epic-16-decision-log.md`, +145 lines, touching nothing under `folio-designer/`
(D-16.R.28's rule, discharged by measurement rather than by assumption).

**The control — `folio-designer/src/App.tsx`**
- `:1941-2109` — `FontFamilyProperty` in full. Note its two-group doc comment is **detached**: it sits at `:1890-1906`, above `renderedFamilyLimit`, not above the function.
- `:1969` `declared` · `:1980` `addable` · `:1986` `shown = addable.slice(0, renderedFamilyLimit)` · `:1990` the flat `matches` list · `:1911` `renderedFamilyLimit = 50`.
- `:2041-2046` `choose()` — **three** arms since 16.5, forking on `familyIsInstalled(match.source)` then on `match.source`. `:2011-2024` `commitFirstUse` — embed, then property only if no refusal.
- `:2063` the `<ul role="listbox" aria-label="Fonts">`; `:2083` its close.
- **The six `role="presentation"` children: `:2064`, `:2070`, `:2071`, `:2077`, `:2078`, `:2082`.**
  Four are conditional; two always paint. Max five render at once.
- `:2065-2070` the `.flatMap((node, index) => index === declared.length ? [heading, node] : [node])`
  heading interleave — **the one element in the walk that reads position semantically**, and the one
  the `role="group"` fix replaces.
- `:2082` the disk-font decline, verbatim: *"Fonts come from this catalogue. A typeface on your own disk cannot be embedded."* Its rationale comment is at `:2079-2081`.
- `:2077` `familyIndexDisclosure()` · `:2078` the "Showing N of M" line · `:2102` the `Add fonts…` button, already a sibling **outside** the `<ul>`.
- `:1754-1769` `MachineFontStore`; `:1759` its label, which **borrows** `AVAILABLE LOCALLY`. `:1938` `lateEmbedRefusal`'s removable branch names that label; its non-removable branch is **already correct**. `:1479` the status bar. `:184`/`:292-310`/`:350-356` the store state and its one session-scoped loader — **no new async work is needed; `storedFaces` is already in state.** **The union — `folio-designer/src/font-index.ts`**
- `:148-150` the doc comment that lies: *"ONE ORDERED LIST … local tier first, then the faces this
  machine already holds, then the rest of the snapshot."*
- `:212-246` `offeredFamilies`. **`:231-235` pushes a `stored` row inside the `for (const row of webFamilies)` loop**, so a stored family with an index row lands at its *web* position. Measured by execution: 4 alternation runs under `familyIsInstalled`, not 2; a planted stored face landed at
  index 900 of 1304; 32 installed rows existed and 31 fell inside the 50-row cap.
- `:264` `familyIsInstalled = source.tier !== 'web'` · `:283-297` `familySourceNote`, exhaustive with `const unhandled: never` at `:289` · `:308-312` `familyIndexDisclosure` · `:50-62` the comment
  contradicting the shipped heading.

**Read-only evidence**
- `folio-designer/src/App.css:208-223` the option rules; `:352-355` `.property-add-fonts`. Token vocabulary is `tokens.css`, name-gated against `DESIGN.md` by `design-contract.test.ts`.
- `folio-designer/src/font-browser-model.ts:198` `sortRows` — **both arms are total orders ending in `localeCompare`**, so 16.3's browser order is a function of `sortRows` alone and is immune to the `offeredFamilies` repair. `:143` `rowTierNote`, a second exhaustive switch over the union.
- `font-store.ts:470` `storeWriteRefusal` and `:502` `storeUnavailableEmbedNote`.

## Tasks & Acceptance

**Execution:**
- [x] `src/font-index.ts` — **repair `offeredFamilies` to return the order its own doc comment already documents.** Collect stored rows into their own list inside the `webFamilies` loop and return `[...local, ...stored, ...orphanedStored, ...web]`. **This is a defect fix with independent standing: the function contradicts its own documentation at `:148-150`, and the grouping ruling
      was grounded on the half that lies.** After it, the control groups and labels the union and does
      not reorder it.
- [x] `src/font-index.test.ts` — **assert the union arrives as exactly TWO runs under `familyIsInstalled` (installed, then not), with a stored face planted at a deep web position.**
      Assert the **run structure, not a row list** — a test naming families re-pins to the snapshot and
      rots at the next index bump. This test reds today.
- [x] `src/font-index.ts` — amend the comment at `:50-62`, which states the name means the `stored` arm
      and *"does NOT mean this arm plus the local arm"*. **A module that contradicts the shipped
      heading is worse than either reading.** The same claim is made in three further places and all
      four are amended together: `scripts/host-font-access.mjs:13` and `:183`, and `font-browser-model.ts:17`. **A defect that exists in four places is four.**
- [x] `src/App.tsx` — partition `matches` into the three groups, then render. **Groups 1 and 2 render in full; the cap becomes `webGroup.slice(0, renderedFamilyLimit)`.** The cap's own comment says
      *"this bounds the DOM and never the claim"* — a cap applied before the partition breaks that.
      Move the "Showing N of M" line into group 3 and have it name the population it counts.
- [x] `src/App.tsx` — **group 2 is deliberately uncapped, with a NAMED revisit trigger written into the
      comment: revisit when a store can hold on the order of 200 entries.** A silent unboundedness is
      the thing this epic has spent the week correcting.
- [x] `src/App.tsx` — fix the listbox: **all six** `role="presentation"` children plus whatever the third heading adds. `role="group"` with `aria-label`, or the notes moved out of the `<ul>` and referenced by `aria-describedby`. **The heading interleave at `:2065-2070` is replaced by this
      change, which is what keeps the keyboard walk order-agnostic.** Add the assertion that reds if
      non-option children return — nothing pins the presentation role today (0 assertions over 67
      test/spec files), so the fix is otherwise unguarded.
- [x] `src/App.tsx` — the three headings, `IN THIS TEMPLATE` / `AVAILABLE LOCALLY` / `AVAILABLE TO INSTALL`; a heading suppressed only when its own group is empty after filtering. **Do not rewrite `familySourceNote` and do not duplicate it into the heading** — the heading names the
      place, the per-row note names the act, and both already ship.
- [x] `src/App.tsx:1759` — **rename the store panel's label to `TYPEFACES THIS DESIGNER HAS DOWNLOADED`**, dropping the borrowed `AVAILABLE LOCALLY —` prefix. Measured: `AVAILABLE LOCALLY` occurs **exactly once across all six mockup files**, at `Font Browser.dc.html:219`, and it is the
      dropdown's group heading; **the design draws no machine-store panel at all.** The store panel
      borrowed a label from a control that did not yet exist. This removes a deviation rather than
      adding one.
- [x] `src/App.tsx:1938`, `src/font-store.ts:470` — follow the rename in the **removable branch only**. `lateEmbedRefusal`'s non-removable branch is already correct and **must not be touched**: a
      local-tier face genuinely has nothing to remove.
- [x] `src/font-store.test.ts:284`, `src/App.font-store.test.tsx:665` — re-point the two `/AVAILABLE LOCALLY/` assertions. **Each must still assert a pointer to a removal control that
      actually exists, not merely that some string is present** — re-pointing a place-keyed guard
      relocates its blind spot unless the property it asserts is restated with it.
- [x] `src/App.tsx:2082` and its comment at `:2079-2081` — **re-derive the disk-font decline against
      THREE premises and record the working in the code comment.** (1) REVERSED, D-16.1: the catalogue
      is no longer the only source. (2) STANDING, D-16.2: faces already on the authoring machine are
      never enumerated or read. (3) NEW, D-16.R.46 Q4: *"installing is only ever a precursor to
      embedding, and embedding is the step the licence requirement gates."* The conclusion holds and
      premise 3 is why it holds more firmly: a fetched face arrives with its `OFL.txt` and its name
      table, a file dragged off a desktop arrives with neither, so it cannot be embedded — and because
      installing exists only to lead to embedding, it cannot usefully be installed either. **It would
      be a dead end with a friendlier first click.**
- [x] `src/App.tsx:1479` — the status bar states the font count **and nothing else new**: *"N fonts in template"* off `families.length`, which is `IN THIS TEMPLATE`'s own predicate, so
      both surfaces teach one model from one source. **No grid, no snap, no selection content.** The
      mockup's `s.added.length` binding is refused for D-16.R.72's reason, and its hardcoded `"3 fonts in template"` else-branch is placeholder data, not a spec. **If it needs more than the
      count, it splits and returns to the orchestrator.**
- [x] `src/App.css` — a treatment for the third group in the existing token vocabulary. **No new token, no colour literal, no `border-radius` that is not `var(--radius-*)`, no gradient, and no `@media`** — `canvas-authority-contract.test.ts:230` permits App.css exactly one, and it is `prefers-reduced-motion: reduce`. Leave `.property-options`' box-shadow alone (DW-178).
- [x] Tests — group membership disjoint and complete; a heading suppressed only on its own empty group;
      picking from `AVAILABLE LOCALLY` embeds with no fetch and moves the row; picking from `AVAILABLE TO INSTALL` sends no engine command and no property commit; the keyboard walk crosses all three
      groups in one sequence; the six-child listbox regression reds.

**Acceptance Criteria:**
- Given the control, when opened, then it draws three groups on the (declared?, `familyIsInstalled`?) axis, its filter and its `Add fonts…` footer, and no group's heading is drawn over a member it does
  not show.
- Given a face this machine holds whose family sits deep in the index, when the control is opened with
  no query, then that face appears under `AVAILABLE LOCALLY` — which is false at HEAD and is the
  measurement this story turns green.
- Given the union, when `offeredFamilies` returns, then rows satisfying `familyIsInstalled` form
  exactly one contiguous run, matching the order the module documents.
- Given a font in any group, when the author reads it, then its relationship to the file is legible
  without explanation, and it changes group only because the author acted.
- Given the disk-font decline, when the control is rebuilt, then it is restated or removed against all
  three premises, and the derivation is in the code comment.
- Given the listbox, when read by assistive technology, then it owns only options, and a test reds if
  that regresses.
- Given the store panel and the dropdown group, when both are on screen, then no two differently
  populated regions share one name.

## Spec Change Log

- **2026-09-03, plan gate at `b8431c4` — the `<intent-contract>` was TRANSCRIBED, not preserved
  verbatim.** Step-02's draft-resume rule preserves the frozen block; it is declined here because the
  orchestrator reset this spec from `ready-for-dev` to `draft` (D-16.R.58) precisely so its contract
  could be amended, and D-16.R.33 parked nine ruled changes for application at this gate. Pre-edit
  contract md5 `88c7de19c0dcf8a0a8292cab82152574` over lines 29-90. `baseline_commit` moved from `a40c34d` (where the contract was written) to `b8431c4` (where its anchors were measured).
- **Six contract edits, all ruled.** Always bullet 1 replaced with D-16.R.72's three-group text · the
  *"Open with an empty store"* matrix row was **false** (the 31 local faces are always present) · the
  `AVAILABLE LOCALLY` pick row said **"one command and one undo"**, which **16.5 falsified — it is two commands and two undos** (`commitFirstUse`, `App.tsx:2011-2024`); this is a closed story's frozen
  text falsified by a later story, recorded here rather than corrected silently · a new matrix row for
  the install pick · `⌘G` struck from the `Add fonts…` row (D-16.R.33 R2) · a new Always clause and a
  new Block If for the heading-truth rule that Q2's measurement forced.
- **Two premises in the record were measured false at this gate.** `font-index.ts:148-150` documents an order `offeredFamilies` does not produce, and D-16.R.33 R1's *"16.4 adds headings rather than
  plumbing"* rested on it. The repair is now a task with independent standing. Separately, the
  amendment block's own line anchors were stale as it predicted; all were re-measured.
- **The keyboard-walk rebuild I proposed was withdrawn after an element-by-element audit.** `move()`, `active`, the option ids, `aria-activedescendant`, `aria-selected` and `choose()` are all order-agnostic; the single position-reading element is the heading interleave at `:2065-2070`, which the listbox fix replaces. The story grows by a reorder, a partition and one `slice` — not a second
  deliverable.

## Design Notes

**Why the flat list survives three groups.** 8.6's comment is explicit: the options are one flat list
because the keyboard is linear even when the list is not. Three groups is more headings interleaved
into the same single list, not three lists — which is exactly why the `role="presentation"` defect gets
worse and has to be fixed here.

**Why the order repair is not scope creep.** Grouping over a flat list is only *labelling* if the list
is already ordered by group. It is not, and the sentence claiming it is was written above the function
rather than measured over it. **A comment is not a measurement.** Every earlier instance of this
epic's signature defect had a tool answering more narrowly than its name implied; this one had no tool
at all, and was findable only by execution.

**Why `AVAILABLE LOCALLY` contains 31 faces nobody fetched.** D-16.2 (OWNER) says the group *"is
fetched faces, never host fonts"*, and this heading widens that. It is a ruling and not an override
because 16.2 delegated this exact question rather than settling it (`font-index.ts:57-63`), and the
clause's load-bearing half — *never the OS font list* — is untouched, with
`src/host-font-access.test.ts` still its tripwire. Deviation row ten, owner-visible at the epic gate. **Why `lateEmbedRefusal`'s two branches are not contradictory.** The `local`/`stored` split has no
consequence when *choosing* a font and a real one when *removing* it. The grouping hides it at pick
time; that sentence correctly surfaces it at removal time. The product already speaks about the
asymmetry in the right place.

## Verification

**Nothing in Epic 16 is CI-verified.** The designer CI job halts at `npm run test` (DW-171), so the six
later steps have not run for as long as DW-152 has been red. Every gate below is a local measurement
with no machine watching. Repair is scheduled immediately after this story (D-16.R.44).

**Commands** — run each on its own line; a conjunction silently drops everything after its first
known-failing term, and `npm test` exits 1 on DW-152.
- `cd folio-designer && npm test` — expect **672 passed / 1 failed of 673 across 55 files**, `rc=1`. Match the red by NAME, not count: `canvas-authority-contract.test.ts:190`, received array exactly `["e2e/e9-5-border-no-ink.spec.ts: /\bgetComputedStyle\s*\(/"]`. **A matching count with a different
  name is a regression.** The count will rise as this story adds tests; the name set must not.
- `cd folio-designer && npm run typecheck` — `rc=0`.
- `cd folio-designer && npx oxlint` — `rc=0`, **exactly 4** `only-export-components` warnings.
  **Take this one from the tool directly, not from `npm run lint`.** The gate pins the count and the
  rule, never the line numbers (they moved to `App.tsx:2273,2280` in this story and the gate still
  matches). `npm run lint` is the convenience spelling and `package.json:23` makes it a bare `oxlint`,
  but a runner can be rewritten by a command hook that fails on the tool's non-JSON output and exits
  **on its own behalf** — measured this story: the same command on the same tree returned `rc=0` in one
  agent's context and `rc=1` in another's.
- **AN EXIT CODE IS NOT PORTABLE BETWEEN AGENTS.** This run relays measurements implementer → builder
  → closer → decision log, and a number that depends on who ran it cannot survive that chain, with
  nothing in its appearance saying which context produced it. For any tool whose runner may be
  rewritten, the **recorded** measurement is taken from the tool directly.
- `cd folio-designer && npm run test:e2e:compile` — `rc=0`.
- `cd folio-designer && npm run build` — `rc=0`.
- `cd lint && go test -count=1 ./...` — four `ok`. **`-count=1` is mandatory** (DW-168): the rules package walks the directory and Go's cache does not track `ReadDir`, so a cached `ok` is no
  measurement at all.
- `cd folio-go && go test -count=1 ./...` — expect `rc=1` with the failing leaf set **exactly** `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`, the mandated permanent red. This story
  touches no Go; if anything else moves, the cause is upstream of this epic.
- Take every exit code from `$?` **immediately** after the command, never through a pipe or a trailing `echo`. The shell is zsh: `${PIPESTATUS[0]}` is wrong here. Run each module's commands in their own invocation — a `cd` persists and breaks a later relative path into a message that reads like a pass.

**Standing rules, re-run and never cited:**
- **The per-row matrix audit reports N rows, N results — never a single verdict.** Any row whose only
  assertion calls the production function directly rather than driving the path is **FAILED, not
  passed**. A row covered only at module level is a **PARTIAL**; never upgrade a partial to a pass.
- **The identifier sweep:** every `DW-\d+` cited in this story's artifacts must resolve to a **definition line** in `deferred-work.md` (182 definitions at this gate). **A citation is verified by
  locating its definition, never by re-reading the citation.**
- **A COMMENT IS NOT A MEASUREMENT.** Where a claim depends on an ordering, an invariant or a bound,
  the evidence is a **run** — not the sentence above the function.
- **State the population beside every zero**, read the matches rather than the count, and pair every
  absence claim with a positive control that must produce hits. An unquoted `grep --include=*.ts` dies
  in zsh printing nothing, which reads exactly like a clean result.

**Guards that red on PROSE, including comments, in the file this story edits:**
- `file/file-access-contract.test.ts:70` bans `cloud|sync|recent files|collaborator|account` case-insensitively over App.tsx raw text. `asScanned` blanks comments **only** for `font-store.ts`.
- `host-font-access.test.ts` bans `queryLocalFonts`, `navigator.fonts`, a **quoted** `local-fonts`, and `FontData`. The re-derivation comment must state its rationale without naming the API — which is possible: `App.tsx:1767` already ships that rationale in prose and the guard is green over 129 files.
- `property-prose-height.test.ts:38-43` — exactly ONE `<input>` may wear `property-value-prose`.
- `canvas-authority-contract.test.ts:230` — App.css may contain exactly one `@media`. **Expected reds, by design, not surprises:** `App.test.tsx:1504` pins the disk-font sentence as an
**exact whole-element** `getByText` match, so restating **or deleting** it reds (a `getBy*` throws on
zero). `App.test.tsx:1488-1489` pins today's two heading literals. Correct these with the change and
say so in the Delivery Log; **if an existing test asserts today's interleaved order, it is pinning the
divergence from the doc comment and is corrected with it — say so rather than editing it quietly.**

**Manual checks:**
- A browser run: open with an empty store and with a populated one; pick from each group; walk the list
  by keyboard end to end across all three groups.
- `e2e/font-embed-boundary.spec.ts:128` is a **measured browser red** at the baseline. **Do not touch
  it** — it is assigned to the post-16.4 CI repair, and two owners is how a ruled deferral vanishes.
- The 23 golden digests, unmoved (`shasum` from the repo root; they live at `<root>/fixtures/`).

**A BROWSER RUN WAS TAKEN, and it is what witnesses this story's accessibility repair.** Chromium `chromium-1217` (Chrome for Testing 147.0.7727.15, 336M) via `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` — the pinned `chromium-1208` is a 428K stub and `chromium_headless_shell-1208` is absent, which is DW-180 biting a third time; **a directory existing is not a browser, and `du -sh` is the ten-second check.** Measured in the real accessibility tree: the listbox's direct child roles are exactly `["group","group","group"]` labelled `IN THIS TEMPLATE` / `AVAILABLE LOCALLY` / `AVAILABLE TO INSTALL`, **`role="presentation"` children = 0** (all six gone), 82 options reachable through the listbox (1 declared + 31 local + the 50-row cap = 82, so partition-then-cap holds in a browser too), and the disclosure is outside the list and reached by `aria-describedby`. `browser-native-roundtrip.spec.ts` — the only cross-boundary authoring witness — **passes**, resolving `getByRole('option', {exact: true})` through the new `div role="listbox" > div role="group" > div role="option"` structure.

**STILL NOT RUN HERE — deferred to the epic catch-up that gates `epic-16: done`:** the matrix corpora, the four AD-21 legs, and `TestCrossTargetByteIdentity`. **Naming them is the point.** With CI halted at step 2 (DW-171) there is no machine that would say otherwise. A green gate list in this story means the gates in this list, and a report that does not say so is making the wider claim by omission.
