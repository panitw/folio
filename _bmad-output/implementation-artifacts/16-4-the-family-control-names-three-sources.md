---
title: 'Story 16.4: The family control names three sources'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: 'a40c34db6cff7372363b2a553710eff48759bef1'
context:
  - '{project-root}/_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The font menu has to answer a question it currently cannot: where is this typeface, and what happens to
my file if I choose it? Three groups, three different answers — already in this template; fetched from
the web and now in this template; on this machine but not in this file yet. A font moves between them
because you did something, never on its own.

There is also an old sentence at the bottom of that menu saying a typeface from your own disk cannot be
embedded. It was true for a good reason, and the reason has just changed. This story does not quietly
carry it forward — it either restates it with the new reason or takes it away deliberately.

<intent-contract>

## Intent

**Problem:** `FontFamilyProperty` draws two groups — the document's chains, and the bundled catalogue.
Epic 16 creates a third relationship (fetched, on this machine, not in this file), and the design draws
all three. The control also carries a disk-font decline derived from a premise D-16.1 has changed, and
a listbox that breaks its required-owned-elements rule four times, registered as deferred at 8.6.

**Approach:** Rebuild the control to the design's three groups plus its filter field and `Add fonts…`
footer; route the third group to Story 16.2's store; re-derive the disk-font decline against D-16.1;
and fix the listbox defect here rather than multiplying it.

## Boundaries & Constraints

**Always:**
- **Three groups mean three relationships, and the difference must be legible without explanation.**
  `IN THIS TEMPLATE` is in the file. `ADDED FROM WEB FONTS` is what this session fetched into the file.
  `AVAILABLE LOCALLY` is on this machine and **not** in the file.
- **A font changes group because the author acted.** Story 8.6's rule — *"nothing says 'added', the
  entry simply moves"* — extends to three groups unchanged.
- **The fork at `choose()` keeps its shape.** A declared name is a `fontFamily` property commit. A
  family not yet in the document is a document change through the embed command. Two decisions, two
  undos; fusing them was refused at 8.6 and is refused here.
- **The accessible names Playwright addresses must not move.** `combobox` "Font family", "Edit font
  chains", "Show fonts"/"Hide fonts", "Clear Font family" — `e2e/component-properties.spec.ts` and
  `browser-native-roundtrip.spec.ts` address them, and the latter is the repository's only
  cross-boundary authoring witness.
- **The listbox defect is fixed, not extended.** 8.6 registered it: four `role="presentation"` children
  inside `role="listbox"`. Three groups plus a filter field plus a footer row makes it worse. Use
  `role="group"` with `aria-label`, or move the notes out of the `<ul>` and reference them with
  `aria-describedby`.
- **The disk-font decline is re-derived or lifted, never carried unread.** D-8.6.1 declined it because
  a file off the author's disk supplies no licence text and no copyright, so the designer would have to
  invent terms or write a document its own parser refuses. **That reasoning still stands on its own
  terms** — a fetched face brings `OFL.txt` and a `name` table with it, a dropped file brings neither —
  so the likely outcome is *restated with the new contrast*, which is more useful than the old wording.
  It is not, however, a foregone conclusion, and it must be written down either way.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **Stories 16.2 and 16.3 are not closed.** The third group has no source and the footer no destination.
- **An accessible name in the e2e specs would move.**
- **The control would offer a font the engine will refuse** — 14.4's rule, applied here.

**Never:** a host font in any group · a group whose membership changes without an author action · a
fourth group.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open with an empty store | Fresh machine, 1 chain | `IN THIS TEMPLATE` only; the other groups absent rather than empty-with-a-heading; `Add fonts…` present | — |
| Open with a populated store | 4 stored, 2 in document | Two groups populated, membership disjoint | — |
| Pick from `AVAILABLE LOCALLY` | Stored face | Embedded with no fetch; row moves to `IN THIS TEMPLATE` | Refusal anchored at the control |
| Pick a declared chain | Name in the document | `fontFamily` property commit, as today | Existing refusal path |
| Filter field | `sara` | All groups filtered; headings for empty groups suppressed | — |
| Add fonts… | Click, or `⌘G` | Opens the browser (16.3) | — |
| Mixed selection | Two components, different families | `Mixed` placeholder, as today | — |
| Keyboard walk | Arrow keys | One linear walk across all three groups, as 8.6's flat `matches` list already does | — |

</intent-contract>

## Code Map

**Designer (`folio-designer/`)**
- `src/App.tsx:1284-1380` — `FontFamilyProperty` in full: the header comment stating the two-group
  rule, `declared` / `catalogue` at `:1310-1315`, the flat `matches` list at `:1317` (**and why it is
  flat — the keyboard walk**), `choose()`'s fork at `:1333-1337`, the combobox at `:1342`, the three
  inline action buttons at `:1358-1362`, the listbox at `:1354-1367`, and the
  `role="presentation"` children at `:1355`, `:1361`, `:1363` and `:1366` — **the registered defect**.
- `src/App.tsx:1366` — the disk-font decline string: *"Fonts come from this catalogue. A typeface on
  your own disk cannot be embedded."* **The sentence D-16.1 puts back in question.**
- `src/App.css:210-223` — `.property-options`, `.property-option`, `.property-option-catalogue`,
  `.property-option-note`. The third group needs a treatment that is distinct without inventing a
  vocabulary.
- `src/App.tsx:608-627` — `pickCatalogueFamily`, the embed path both non-document groups share.
- `src/font-chain-control.ts` — `FontChainControl.action`, refusal anchoring.
- `e2e/component-properties.spec.ts:35-60`, `e2e/browser-native-roundtrip.spec.ts` — the accessible
  names that must not move.
- `_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md` frontmatter
  `deferred:` — the listbox finding, its evidence and its suggested shapes.

## Tasks & Acceptance

**Execution:**
- `src/App.tsx` — three groups over one flat `matches` list, preserving the linear keyboard walk;
  suppress a heading whose group is empty; the design's filter field at the top and `Add fonts…`
  pinned at the bottom.
- `src/App.tsx` — the third group's source is Story 16.2's store, filtered to exclude anything already
  declared, exactly as `catalogue` already excludes declared families.
- `src/App.tsx` — the listbox fix from 8.6's deferred entry, with a test that reds if the
  non-option children return.
- `src/App.tsx` — the disk-font decline: re-derived and restated against D-16.1, or lifted; **and the
  reasoning recorded in the code comment**, which is where 8.6 put the last one.
- `src/App.tsx` — `Add fonts…` and `⌘G` open the browser; `src/shortcuts.ts` if the shortcut lands.
- `src/App.tsx` — the status line's font count, as the design's own status bar states it.
- `src/App.css` — a treatment for the third group in the existing token vocabulary.
- Tests: group membership is disjoint and complete; picking from the store embeds without a fetch and
  moves the row; the keyboard walk crosses all three groups in one sequence; accessible names unchanged;
  the listbox reds if regressed.

**Acceptance Criteria:**
- Given the control, when opened, then it draws the design's three groups, its filter field and its
  `Add fonts…` footer.
- Given a font in any group, when the author reads it, then its relationship to the file is legible
  without explanation, and it changes group only because the author acted.
- Given a font in `AVAILABLE LOCALLY`, when picked, then it is embedded from the store with no fetch,
  as one command and one undo, and it moves to `IN THIS TEMPLATE`.
- Given the disk-font decline, when the control is rebuilt, then it is restated with its current reason
  or removed deliberately, and the choice is recorded.
- Given the listbox, when read by assistive technology, then it owns only options, and a test reds if
  that regresses.
- Given the e2e specs, when they compile, then every accessible name they address still resolves.

## Design Notes

**Why the flat list survives three groups.** 8.6's comment is explicit: the options are one flat list
because the keyboard is linear even when the list is not, and splitting the markup would split the
arrow-key walk with it. Three groups is more headings interleaved into the same single list, not three
lists — which is also exactly why the `role="presentation"` defect gets worse and has to be fixed here.

**The decline is the interesting part of this story.** D-8.6.1 declined disk fonts because the
catalogue was the only source that could supply licence terms. D-16.1 removes "the catalogue is the
only source" — but not the reason: a fetched face arrives **with** its `OFL.txt` and its `name` table,
and a file dragged off a desktop arrives with neither. The premise moved and the conclusion probably
did not, which is precisely the case where a team carries a sentence forward without re-reading it.
Writing the new derivation down is cheap; discovering later that nobody checked is not.

## Verification

- `cd folio-designer && npm run test && npm run test:e2e:compile && npm run build`
- A browser run: open with an empty store and with a populated one; pick from each group; walk the list
  by keyboard end to end.
- The 23 golden digests, unmoved.
