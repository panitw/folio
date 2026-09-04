---
title: 'Story 16.9: The typography panel offers what you can use'
type: 'refactor'
created: '2026-09-04'
status: 'in-progress'
review_loop_iteration: 0
baseline_commit: '4d2b27e'
context:
  - '{project-root}/_bmad-output/implementation-artifacts/16-4-the-family-control-names-three-sources.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

Four things the owner did not ask for come out of the typography panel, and one misaligned control gets
fixed.

The font-chain editor goes — the panel with Rename, Delete and Add entry, and the button that opens it.
The fallback itself stays: a document still says Roboto first, then Noto Sans Thai, then Noto Sans SC, so
Thai and Chinese text renders exactly as it does now. What goes is the ability to hand-edit that order.

The Clear button on the font field goes. Text always has a typeface; there is no such thing as text with
none, so a control that appears to remove one was offering something the product cannot do.

The dropdown stops listing the thousand-odd families that are not on this machine. Installing is what the
font browser is for, and `Add fonts…` at the foot of the menu is the door to it. The menu now offers only
what can be used immediately, with no network and no waiting.

And the chevron on the font field sits off-centre; the design has it centred.

<intent-contract>

## Intent

**Problem:** The typography panel carries three controls the owner never asked for — a font-chain editor, a
Clear button on the font field, and a dropdown group listing ~1,273 families that are not on this machine
and cannot be used without a download. The field's chevron is also not vertically centred, unlike the
design.

**Approach:** Remove all three, keeping the document's font-chain *data* untouched, and centre the chevron.

## Boundaries & Constraints

**Always:**
- **THE CHAIN DATA STAYS. ONLY ITS EDITOR GOES.** A document's `fonts` map keeps its fallback arrays —
  `{"Roboto": ["Roboto", "Noto Sans Thai", "Noto Sans SC"]}` — and **Thai and CJK text must render exactly
  as it does today.** Removing the editor may not change one rendered byte.
- **The dropdown offers only what is usable now:** `IN THIS TEMPLATE`, `AVAILABLE LOCALLY`, and the
  `Add fonts…` row. **No `AVAILABLE TO INSTALL` group, and no web-tier row anywhere in the control.**
- **`Add fonts…` stays and still opens the browser.** It is now the only route to a family that is not on
  this machine, so it may not be weakened.
- **The font browser is untouched.** It keeps every family, its own specimens and its install controls.
- **Text always has a typeface.** With `Clear` gone, no control offers to leave an element with none.
- **Deleted code is deleted** — no unreachable components, dead props, or orphaned CSS.
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Ask First:**
- Any change to a document's `fonts` map, its shape, or how the engine resolves it.
- Removing `AVAILABLE LOCALLY` or `IN THIS TEMPLATE`.
- Any change to the font browser.

**Never:** a second chain editor elsewhere · a web-tier row in the dropdown · reducing Thai or CJK
coverage · changing the render of any existing document · a fetch triggered by opening the dropdown.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Open the typography panel | Any selection | **No FONT CHAINS panel and no control that opens one** | — |
| Open the font field | Any document | Chevron **vertically centred** in the field | — |
| Field with a committed family | `Roboto` | **No Clear control** beside the field | — |
| Open the dropdown | Empty store | `IN THIS TEMPLATE` + `AVAILABLE LOCALLY` + `Add fonts…`; **no third group** | — |
| Open the dropdown | 3 stored faces | Those faces appear under `AVAILABLE LOCALLY`; still no third group | — |
| Open the dropdown | — | **Zero network requests** — asserted with a fetch spy | Never fetched |
| Filter the list | `sara` | Both remaining groups filter; a heading is suppressed only when its own group is empty | — |
| `Add fonts…` | Click | Opens the browser, unchanged | — |
| Thai + CJK document | Chain with Noto fallbacks | **Renders byte-identically to before this story** | — |

</intent-contract>

## Code Map

Anchors at `4d2b27e`. **Re-verify before editing** — several move under this change.

**The chain editor**
- `folio-designer/src/FontChainEditor.tsx` — the whole component; `:58` `FontChainEditor`. **Deleted.**
- `folio-designer/src/App.tsx` — its render site inside the `TYPOGRAPHY` section, the `chainsOpen` /
  `onToggleChains` state and props threaded to `FontFamilyProperty`, and the disclosure button whose
  accessible name is `Edit font chains` / `Hide font chains`.
- **`engine-protocol.ts` also names `FontChainEditor`-related commands — the ENGINE SIDE IS NOT TOUCHED.**
  Chain commands remain part of the protocol; only the UI that issued them goes.

**The Clear control**
- `folio-designer/src/App.tsx:2242` — `{(committed !== '' || !uniform) && <button … aria-label="Clear Font
  family" …>}`. **Story 16.4 already measured `Clear Font family` as having zero assertions across 67 test
  and spec files**, so it was dropped from the protected accessible-name list then.

**The third group**
- `folio-designer/src/App.tsx:2105` — the `groups` array; the `install` entry is the third element.
  `toInstall`, `shownToInstall`, `renderedFamilyLimit` and the cap note (`Showing N of M …`) exist only to
  serve it.
- `folio-designer/src/font-index.ts` — `offeredFamilies` still returns every tier **for the browser**; the
  control filters to `familyIsInstalled`. **Do not narrow `offeredFamilies` itself.**

**The chevron**
- `folio-designer/src/App.css:209` `.property-disclosure { font: var(--type-body); }` — the only rule on it.
- The design: `_bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/mockups/Font Browser.dc.html`.
  **Read the mockup's own alignment rather than guessing a value.**

## Tasks & Acceptance

**Execution:**
- [ ] `folio-designer/src/FontChainEditor.tsx` — delete the file.
- [ ] `folio-designer/src/App.tsx` — remove its render site, the disclosure button, and the `chainsOpen` /
      `onToggleChains` state and props threaded for it.
- [ ] `folio-designer/src/App.tsx` — remove the Clear control.
- [ ] `folio-designer/src/App.tsx` — remove the `install` group, `toInstall`, `shownToInstall`, the cap and
      its note; the control's rows become `declared` + `onThisMachine`.
- [ ] `folio-designer/src/App.css` — centre the chevron; delete rules orphaned by the removals.
- [ ] Tests — retire what asserted the removed controls, **each with a one-line reason**; keep everything
      about picking, specimens and the browser. **Add a fetch-spy assertion that opening the dropdown makes
      zero requests**, since the group that could fetch is gone.
- [ ] A browser run: photograph the panel and the open dropdown; report the chevron's box against the
      field's.

**Acceptance Criteria:**
- Given any selection, when the typography panel renders, then no chain editor and no control opening one
  exists.
- Given a document with Thai and CJK text, when it renders, then its bytes are **identical to before this
  story**.
- Given the dropdown open, when its groups are read, then they are exactly `IN THIS TEMPLATE` and
  `AVAILABLE LOCALLY`, and **no request has been made**.
- Given the font field, when it is photographed in a browser, then the chevron is vertically centred within
  it.

## Design Notes

**This narrows D-16.R.72's ruling from three groups to two, and that is deliberate.** The axis survives —
rows are still grouped by *where the bytes are* — but the third answer, *not on this machine*, is no longer
offered here. It is offered in the browser, which is the surface built for it.

**IT ALSO DISCHARGES A REGISTERED HAZARD.** The deferral that a pick from `AVAILABLE TO INSTALL` blocks up
to 30 s on a stall and 180 s against a slow host existed **because the dropdown could trigger a fetch.**
With no web row in the control, that pick cannot happen. **Close that entry against this story rather than
leaving it open against a control that no longer offers the path** — and the fetch-spy assertion is what
keeps it closed.

**The chain editor's removal is a UI removal and nothing else.** `engine-protocol.ts` keeps its chain
commands, the document keeps its `fonts` map, and the engine keeps resolving fallbacks. **A story that
removed the data would be a different story and is forbidden above.**

## Verification

**Nothing in Epic 16 is CI-verified** (DW-171); every gate is a local measurement with no machine watching.
Say what you did not run.

**Commands** — one per line, exit codes from `$?` immediately, never through a pipe or wrapper; zsh, so
`${PIPESTATUS[0]}` is wrong. **Read lint from `npx oxlint` directly, never `npm run lint`.**
- `cd folio-designer && npm test` — the count **falls** as removed controls' tests retire. The one
  permanent red is `canvas-authority-contract.test.ts:190` (DW-152); **match it by NAME, not count.**
- `cd folio-designer && npm run typecheck` — **the gate that catches dead props and orphaned state.**
- `cd folio-designer && npx oxlint` (**exactly 4** warnings) · `npm run build` · `npm run test:e2e:compile`
  — all `rc=0`.
- `cd folio-go && go test -count=1 ./...` — no Go here; the failing leaf set must stay exactly
  `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, **and the golden digests must be unmoved**, which is
  what proves the chain data survived.
- `cd lint && go test -count=1 ./...` — `rc=0`.

**A BROWSER RUN IS REQUIRED for the chevron.** No test here can see alignment: jsdom does no layout and
`getComputedStyle` is banned outside the canvas. Use `chromium-1217` via
`PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH` (`chromium-1208` is a 428K stub, DW-180). **Report the chevron's
bounding box and the field's, and show the centres agree** — a screenshot alone is not the measurement.

**Standing rules — re-run, never cite:** the per-row matrix audit reports **N rows, N results** (this
matrix is **9 rows**); a deletion needs a guard, so **assert each removed control's absence and red-prove
that assertion by restoring the control**; state the population beside every zero and pair every absence
claim with a positive control; **a comment is not a measurement.**
