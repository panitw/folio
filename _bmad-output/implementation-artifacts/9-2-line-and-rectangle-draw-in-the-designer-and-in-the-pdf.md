---
title: 'Story 9.2: Line and Rectangle draw, in the designer and in the PDF'
type: 'feature'
created: '2026-09-02'
status: 'done'
review_loop_iteration: 1
followup_review_recommended: true
reconstructed: true
reconstructed_from_commit: '791ed002462807b71bb2223e1fe4f0900121f0e5'
baseline_commit: 'be40f248f3f7ff9749a54f4a070e54d8e5326300'
baseline_revision: 'be40f248f3f7ff9749a54f4a070e54d8e5326300'
audit_revision: '6e06cc784417bed50396ff742440a8fa0ef875fc'
context:
  - '{project-root}/_bmad-output/planning-artifacts/epics.md'
  - '{project-root}/_bmad-output/implementation-artifacts/9-1-the-engine-paints-a-component-s-background-and-border.md'
warnings: [reconstructed-after-the-fact, shared-commit, unreviewed-at-ship, carries-unattributed-work]
---

## ⚠ This spec was written after the code shipped

**Nothing here was authored before implementation.** Stories 9.1 and 9.2 were implemented on
2026-08-30 in a single commit, `791ed00` *("Make a component's box print, on the page and on the
canvas", 22 files, +1649/−162)*, with **no story file, no acceptance criteria, no review and no
retrospective**. Both sat at `review` in `sprint-status.yaml` until this reconstruction, which exists
because **Story 15.3 cuts a release tag over this code** (owner decision D-000.5).

**The contract below was derived by reading the shipped diff.** Where the original intent could not
be recovered, this spec says so under `## What could not be determined` rather than inventing one.

**Which parts of `791ed00` this story owns** (Story 9.1 owns the engine box paint on the page):

*Owned by 9.2 — line and rectangle reaching the designer canvas and the PDF:*
- `folio-go/component_commands.go` (+24) — `lineDropHeight`, the line's drop height override in
  `dropComponent`, and the starting `style` a dropped line and rect receive in
  `createComponentInBand`.
- `folio-go/element_box_test.go`'s `TestRectAndLineElementsDrawTheirBox` and
  `TestAPlacedLineAndRectArriveVisible` (the rest of that file is 9.1's).
- `folio-go/page_setup.go` — **only** the `CanvasComponent` style projection arm
  (`background` / `borderWidth` / `borderColor` / `borderEdges`) the canvas box reads.
- `folio-designer/src/App.tsx` — `ComponentBox`, the BOX inspector section (`borderFields`,
  `backgroundField`, `BorderEdgesProperty`), and the CONTENT/`fx`/focus-ring inspector rework.
- `folio-designer/src/App.css` (+65/−6), `App.test.tsx` (+138/−2),
  `e2e/browser-native-roundtrip.spec.ts` (+14/−10).
- `folio-designer/src/preview/pdf-viewer.tsx` (+19/−1) and its test — the preview oversample
  constant. **See the caveat below: this belongs to no acceptance criterion in either story.**

*Owned by neither story, carried in the same commit* — recorded in full in Story 9.1's spec:
`folio-go/text_alignment.go` (+89) and its callers (this, not the box paint, is what moved four
statement goldens by +4 bytes per page — D-15.1.1, intended per D-R7.6, goldens re-recorded);
`CanvasProjection.FontFamilies`/`DefaultFontSize` and the matching `engine-protocol.ts` validator;
and `_bmad-output/implementation-artifacts/evidence/story-6.7-roundtrip-manifest.json` (204 lines),
which Epic 8's boundary gate proved stale and impossible to have come from a real Playwright run —
Playwright had never executed in this repository until 2026-09-02. **Noted; not Epic 9's to fix.**

## In plain terms (read this first if you just want the gist)

*This section is background, not a requirement; the contract below governs.*

The palette offers five things to place. Two of them printed nothing at all. A rectangle and a rule
could be dropped, moved, resized and saved, and the finished document came out as though they had
never been placed — they were not merely unstyled, they never reached the page in any form. That
made the palette five items with three that worked.

They work now, and they work by being ordinary. A rectangle is simply a box, so it prints the way
every other box prints. A rule is a box too, filled rather than outlined, which means the height you
give it is its thickness: make it thinner and the line gets finer. Neither needed a rule of its own.

Placing one now gives you something you can see. Before, a freshly dropped rectangle carried no
appearance at all, so the only feedback was an outline the design surface draws for its own
purposes; you had to go and set a colour before anything existed on the page. A dropped rule arrives
one point thick and filled, a dropped rectangle arrives outlined, and both are ordinary settings you
can change or clear like any other.

The design surface draws the same box, from the engine's own description of it rather than from any
guess of its own — so what you see is meant to be what prints. One case where that is not yet true
is written down in the review below.

<intent-contract>

## Intent

**Problem:** Line and Rectangle are two of FR4's five palette components and neither reaches the
page model at all — `render_visibility.go` records their verdicts as "LATENT, not broken". A rect
prints nothing whatever it is styled with. Separately, a newly dropped line or rect carries no style
at all, so even once the engine paints boxes there is nothing to paint: the author must style an
invisible element before anything appears.

**Approach:** Add **no kind-specific rule**. A rect *is* a box, so Story 9.1's element-box collector
covers it with no rect arm. A line *is* a box too — a filled bar — so its **declared height is its
thickness** and it needs no thickness field. Give each kind the one declaration that makes it the
shape its palette entry names, at drop time, as an ordinary editable style value. Paint the same box
on the canvas from the **engine's own projection**, never from a browser-side model of style.

## Boundaries & Constraints

**Always:**
- **No rect-specific and no line-specific paint path.** Both go through
  `collectElementBoxRects` → `buildCellRectWithBackgroundField` like every other kind.
- **A line's declared height is its thickness.** Nothing else expresses it.
- **A dropped component's starting style is an ordinary style value** the author edits or clears like
  any other — never a hidden default the engine substitutes at render time.
- **The canvas paints from `CanvasProjection`**, never from a TypeScript model of `style` (AD-15,
  AD-17).
- **The engine's defaults are mirrored, not re-decided**, where a border declares less than all of
  itself: 0.5pt, `#000000`, all four edges — the numbers
  `buildCellRectWithBackgroundField` uses.
- **A refused commit leaves the last accepted box showing**, as every other property already behaves.

**Block If:**
- A second implementation of "what does `style.border` mean" would appear in TypeScript.
- A line would gain a thickness field distinct from its height.
- The canvas would measure anything in the browser (AD-17 / Story 5.9).

**Never:** a browser-side style model · `window.devicePixelRatio` or any display-derived number ·
a kind-specific rect or line renderer.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Rect with a box | `type: "rect"` + background and/or border | Prints as that box, through 9.1's collector, with no rect-specific rule | none |
| Line with a background | `type: "line"`, `height: 1` | Prints as a **filled bar** 1pt tall — declared height is thickness | none |
| Drop a Line from the palette | `dropComponent {"type":"line"}` | Arrives with `style.background = "#000000"` and `height = 1000` millipoints (1pt), **off the 6pt grid on purpose** — a rule's thickness is not a position, and snapping applies to x/y alone | none |
| Drop a Rectangle | `dropComponent {"type":"rect"}` | Arrives with `style.border = {color:"#000000", width:1000}` (1pt) at the default 72×24pt box | none |
| Canvas paints a box | any component carrying background/border in the projection | Same fill, border width, colour and edge set the engine will paint, from `CanvasProjection` | none |
| Border declaring less than all of itself | e.g. only `borderColor` | Canvas mirrors the engine's defaults: 0.5pt, `#000000`, all four edges | none |
| `"border": {}` — no fields at all | reachable only by hand-editing | **Canvas paints nothing; the PDF still strokes a 0.5pt black box.** Disclosed in `App.tsx:1624-1626` as a known divergence | none |
| `border.edges: []` | present, non-null, empty | **AS SHIPPED: the canvas paints a full four-sided border; the PDF paints none.** Finding 5 in Story 9.1's review — a defect, not an intended arm | none |
| Engine refuses a committed box value | e.g. a bad colour | Canvas keeps showing the engine's last accepted box | refusal anchored at the control |
| Preview rasterization | any document | Rasterized at a **fixed constant multiple** (`previewOversample = 2`) of the displayed size, never a display-derived ratio (AD-17 bans the pixel-ratio property by name, and its scanner reads comments too) | none |

</intent-contract>

## Code Map

**Go — engine (`folio-go/`), at audit revision `6e06cc7`**
- `component_commands.go` — `const lineDropHeight geom.Length = 1000`, applied in `dropComponent`
  when `elementType == template.ElementLine`.
- `component_commands.go` — in `createComponentInBand`, the starting declarations:
  `styleFor(&element).Background = present("#000000")` for a line;
  `styleFor(&element).Border = present(Border{Color:"#000000", Width:1000})` for a rect.
- `page_setup.go:1740-1768` — the `CanvasComponent` style projection: `Background`, `BorderWidth`
  (through `canvasPropertyLength`, which **bounds and rejects out-of-range values**), `BorderColor`,
  `BorderEdges`. `BorderEdges` carries `json:"borderEdges,omitempty"` — the source of Finding 5.
- `element_box.go` — 9.1's collector; rect and line are two of its four eligible kinds and get no
  special case.

**Designer (`folio-designer/src/`)**
- `App.tsx:1616-1640` — `ComponentBox`: reads `component.background`, `borderWidth`, `borderColor`,
  `borderEdges` off the projection and emits per-edge CSS border shorthand; painted as a child
  beneath the content so the selection tint and dotted placement outline stay the canvas's own.
- `App.tsx` — `borderFields` / `backgroundField` / `BorderEdgesProperty`, the BOX inspector section.
  Padding rows were removed; `style.padding` remains an engine property a loaded document keeps and
  renders.
- `engine-protocol.ts` — `isCanvasProjection` / component validation; the `hasOnly` key list is the
  contract for what a component may carry.
- `preview/pdf-viewer.tsx:12-15,56-78` — `const previewOversample = 2`; the page is rasterized at
  `scale * previewOversample` and the canvas is given an explicit CSS size at `scale`.

## Tasks & Acceptance

*Reconstructed. These describe what the shipped code does.*

**AC1 — A rect prints as its box, with no rect-specific rule.** *Shipped; `element_box.go` has no
rect arm, and `element_box_test.go:177` asserts a rect with background+border produces one `Rect`
carrying both.*

**AC2 — A line prints as a filled bar of its declared box.** *Shipped;
`element_box_test.go:177` asserts the line element yields `HasFill` with `W=300000, H=1000`
millipoints — "its declared height IS its thickness".*

**AC3 — A dropped Line or Rectangle arrives visible.** *Shipped;
`TestAPlacedLineAndRectArriveVisible` (`element_box_test.go:322`) drops each kind through the real
`dropComponent` command and asserts, off the resulting **projection**: line → `background
"#000000"`, height `1000`; rect → `borderColor "#000000"`, `borderWidth 1000`, height `24000`.*

**AC4 — The canvas paints the engine's box.** *Shipped; `ComponentBox` consumes only projection
fields and mirrors `buildCellRectWithBackgroundField`'s defaults (`?? 500`, `?? '#000000'`,
`?? boxEdges`). **Partially defeated for `edges: []`** — see Finding 5 in Story 9.1's review.*

**AC5 — A refused box value leaves the last accepted box showing.** *Shipped; the BOX controls use
the same `documentGeneration` / `onCommit` refusal-anchoring path as every other property.*

**AC6 (unattributed) — the PDF preview rasterizes at a fixed multiple of its displayed size.**
This shipped in the same commit and is **covered by no acceptance criterion in either story's epic
text**. It is recorded here so it is not lost, and flagged as scope that arrived without a contract.

## Review Triage Log

### 2026-09-02 — First review, at audit revision `6e06cc7`

**This code had never been reviewed.** The full ranked finding list — with reproduced evidence — is
in Story 9.1's spec, because the engine and the canvas are two ends of one mechanism and splitting
the evidence would hide the pairing. **Nothing was fixed**; this pass had no mandate to modify source.

The findings that land on **this** story:

- **Finding 5 (medium, undisclosed canvas/PDF divergence).** `border.edges: []` projects as *no*
  `borderEdges` key (`omitempty` drops an empty slice), so `App.tsx:1631`'s
  `component.borderEdges ?? boxEdges` defaults to **all four edges**. Measured at `6e06cc7` with
  `border:{"edges":[],"color":"#ff0000"}` on a rect: the projection is
  `{"id":"e1","type":"rect",…,"borderColor":"#ff0000"}` with no `borderEdges`, while the rendered PDF
  contains **0 fills and 0 strokes**. The canvas paints a full red rectangle the printed document
  does not have. The neighbouring `"border": {}` divergence **is** disclosed in the code
  (`App.tsx:1624-1626`); this one is not. Both cut against AD-17 / Story 5.9's "the canvas shows what
  prints"; only one is written down.
- **Finding 1 (high, fail-open).** The same `edges: []` document *also* occupies a real column item
  and adds a **blank second page** to the PDF. Full evidence in Story 9.1's spec.
- **Finding 3 (high, fail-open).** Because 9.1's collector accepts **text**, a styled text element
  makes `ContentWindowCountIsExact` claim exactness for a count that is wrong. Reproduced. The
  divergence is documented in `page_setup.go`'s comment and is not one of the four `exact` causes.
- **Finding 2 (high).** A negative `style.border.width` reaches the content stream as `-5 w`, which
  ISO 32000-1 §8.4.3.2 forbids — **while `canvasPropertyLength` on this story's own projection path
  already refuses the same value** (`folio: component borderWidth exceeds the projection bound`).
  The designer cannot open a document the engine will render into an invalid PDF. The guard exists;
  it is simply on the wrong path.
- **Finding 7 (low).** `engine-protocol.ts`'s `isTextPaint` fragment bound was **relaxed** from
  `component.x + paint.width` to `component.x + Math.max(paint.width, component.width)`. That
  widening belongs to the text-alignment work, which has no story, so a browser-side validation bound
  was loosened with no contract behind it.

**Comment claims checked.** `pdf-viewer.tsx`'s claim that the multiplier is "a CONSTANT, never read
from the display" is **true** — `previewOversample` is a module-level literal and no
`devicePixelRatio` reference exists in the file. `App.tsx:1624-1626`'s disclosure about `"border":
{}` is **true and reproduced** (canvas projects nothing; PDF strokes one box).

**Coverage.** `App.test.tsx` (+138) and `element_box_test.go`'s two 9.2 tests assert real
properties — none is vacuous. But **no test pairs the canvas projection against the rendered PDF for
a partial or empty edge set**, which is exactly where Finding 5 lives, and none covers a dropped
component's style being subsequently cleared.

## Design Notes

**Why a line has no thickness field.** A line is a box; giving it a separate thickness would create a
second geometry authority for the same number and a second thing to keep in step across the format,
the engine, the projection and the inspector. Height already means "how tall this is". The cost is
that a rule dropped from the palette must break the 6pt grid (`lineDropHeight = 1000`), which the
code states as deliberate: thickness is not a position, and snapping applies to x/y alone.

**Why the drop carries a real style rather than a render-time default.** A hidden default would make
"unstyled" and "black" indistinguishable in the file and unclearable in the inspector. An ordinary
declaration is editable, clearable, serializable and round-trippable by machinery that already
exists.

**Padding.** The four padding rows were removed from the inspector by owner call on 2026-08-30.
`style.padding` remains an engine property a loaded document keeps and renders — the removal is of
the control, not of the field.

## What could not be determined

Stated plainly, because "could not look" is not "all clear":

1. **Why `1pt` was chosen for both the line's drop thickness and the rect's drop border width.** No
   comment, test or epic sentence gives a reason; the values read as a designer's judgement with no
   record.
2. **Whether the `edges: []` canvas/PDF divergence (Finding 5) was noticed.** The author disclosed
   the neighbouring `"border": {}` divergence in the same function's comment, which makes silence on
   this one ambiguous between "not noticed" and "noticed and not written down". The record cannot
   distinguish them.
3. **Whether the preview-oversample change (AC6) belonged to this story at all.** The commit message
   presents it as a separate concern under its own heading; no epic acceptance criterion covers it.
4. **Whether the inspector rework** — CONTENT collapsed to one Text field, the `fx` marker, the
   `Visible if` wording, the focus-ring inset — **was specified anywhere.** The commit message
   attributes it to "reported confusion" with no ticket, and it appears in no epic acceptance
   criterion. It is real user-facing behaviour shipped with no contract.
5. **What the original acceptance criteria were.** The epic text was written in the same session as
   the implementation; there is no earlier draft in the repository.

## Verification

**Run at audit revision `6e06cc7` (`main`, tree clean), 2026-09-02.** Gate results, working
directories and the 23 golden digests are recorded once, in Story 9.1's `## Verification` — the same
single run covers both stories, and duplicating the numbers here would create two records that can
drift. Summary, for reading convenience:

- `folio-go` (`/Users/panitw/Projects/folio/folio-go`): `go test -count=1 ./...` → **1877 pass /
  2 fail / 5 skip**, the two failures being the documented permanent red
  (`TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`). `go vet ./...` → clean.
- `lint` (`/Users/panitw/Projects/folio/lint`): `go test -count=1 ./...` → **red on 2 of 5 runs**,
  from a cross-package race (Finding 8 in Story 9.1's spec), not a code regression.
- `gofmt -l folio-go lint` from **the repo root** → exactly
  `lint/internal/rules/licencegraph_test.go`. ⚠ From `lint/` it prints `lstat` errors that read as
  clean.
- `folio-designer` (`/Users/panitw/Projects/folio/folio-designer`): `npm run typecheck` exit 0;
  `npm run lint` → **exactly 4** `only-export-components`; `npm test` → **42 files / 432 tests
  passed**.
- `shasum -a 256 fixtures/*/expected.pdf` from the repo root → **23 digests, all holding**.

**Per-epic cadence: the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and Playwright
were NOT run** — those are Epic 9/10's boundary gate, after Epic 10.

**Tree state.** No source file was modified, nothing was committed, nothing was pushed, no branch was
created.

## Delivery Log

### 2026-08-30 — shipped, unrecorded
Implemented and committed as part of `791ed00` with no story file, no acceptance criteria, no review
and no retrospective. Left at `review` in `sprint-status.yaml`.

### 2026-09-02 — reconstructed and reviewed
Spec written retroactively from the shipped diff at audit revision `6e06cc7`. First review of the
code; findings recorded in Story 9.1's spec and cross-referenced above. **None fixed.** Finding 5 —
the canvas painting a border the PDF will not print — should be triaged before Story 15.3 cuts a tag
over this code.
