---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - _bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/prd.md
  - _bmad-output/planning-artifacts/prds/prd-folio-2026-08-22/addendum.md
  - _bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md
  - _bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/DESIGN.md
  - _bmad-output/planning-artifacts/ux-designs/ux-folio-2026-08-23/EXPERIENCE.md
  - _bmad-output/specs/spec-folio/SPEC.md
  - _bmad-output/specs/spec-folio/acceptance.md
  - _bmad-output/specs/spec-folio/glossary.md
excludedDocuments:
  - docs/folio-mvp-plan.md  # superseded by PRD §4 departures D1-D7
  - "**/review-*.md, **/.memlog.md"  # process records, not content
---

# folio - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for folio, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

**Folio Designer**

FR1: Lay out a report on a visual canvas with visible page boundaries, placing components at absolute coordinates by drag-and-drop, and resizing them directly.
FR2: Configure page setup — A4, Letter, or custom page size; portrait or landscape; page margins.
FR3: Align work using a grid with snapping.
FR4: Place components from an MVP palette of exactly five: Text, Image, Table, Line, Rectangle.
FR5: Edit component properties — position (X/Y), size (width/height), font family, font size, bold, italic, text alignment, vertical alignment, border, padding, background, visibility, data binding.
FR6: Compose the report from three bands: Page Header, Content, Page Footer. Group/report headers and group footers deferred.
FR7: Bind a component to a JSON path through the designer UI, choosing from paths discovered in a loaded sample JSON document.
FR8: Open a `.folio` file from the local machine and save it back, with no server round-trip and no account.
FR9: Load a sample JSON data document alongside the template, used for binding discovery and preview.
FR10: Edit table structure in the designer — add/remove columns, set each column's width, alignment, and header label, bind each column to a field of the row scope, configure per-column footer aggregates.

**Template Format**

FR11: Persist a report definition as a portable, human-readable text file with extension `.folio`, carrying a format `version`, page setup, and ordered band content.
FR12: Remain readable, diffable, and mergeable in Git; usable in CI/CD and over an API; editable directly by a human or an AI agent without the designer.
FR13: Carry a format version the library validates on load, so a template authored against a future format fails clearly rather than rendering incorrectly.

**Data Binding and Expressions**

FR14: Accept report data as JSON supplied by the calling application. Folio never connects to a database.
FR15: Bind a scalar field by dotted path, e.g. `{{customer.name}}`.
FR16: Bind a repeating region to a collection, e.g. `transactions[]`.
FR17: Reference the current row's fields inside a repeating region through a defined row scope.
FR18: Evaluate a small expression language inside `{{ }}` bindings, offering exactly eight functions: `sum()`, `count()`, `avg()`, `formatDate()`, `formatNumber()`, `upper()`, `lower()`, `if()`.
FR19: Aggregate over a collection field, e.g. `{{sum(transactions.amount)}}`.
FR20: Control component visibility conditionally. Conditional visibility in scope; conditional formatting out.

**Parameters**

FR21: Accept runtime parameters separately from report data, addressed through a distinct namespace: `{{params.reportDate}}`.

**Tables and Repeating Sections**

FR22: Generate table rows dynamically from a bound data collection.
FR23: Lay out columns at fixed widths with per-column alignment.
FR24: Render a header row, cell borders, and cell padding.
FR25: Break a table across pages: rows are atomic; a row that does not fit moves whole to the next page; a row taller than the content area renders at the top of a fresh page clipped to that page's content height with a diagnostic; a footer aggregate row is never orphaned; cell content wraps within its column width; column widths never negotiate against content.
FR26: Repeat the table header at the top of every continuation page.
FR27: Render footer aggregates — sum, count, average.
FR28: *(if capacity permits)* Alternating row styling.

**Rendering and Output**

FR29: Produce PDF as the only required output format for MVP.
FR30: Render multi-page documents with page headers and page footers on every page.
FR31: Render page numbers, including the `Page X of Y` form, resolved by a two-pass layout.
FR32: Embed and subset all fonts used in the document, so the PDF renders identically on a machine that has none installed, and so font subsetting is byte-reproducible.
FR33: Embed images inside the `.folio` template itself. Folio does not fetch images by URL and does not read them from the filesystem at render time.

**Preview**

FR34: Offer a fast Design Canvas view for interactive editing, understood to be an approximate visual representation.
FR35: Offer an exact PDF Preview produced by the real `folio-go` engine compiled to WebAssembly running in the browser tab — not a browser-based approximation, not a remote service.
FR36: Operate the designer and its exact preview fully offline, with neither template nor data leaving the user's machine.

**folio-go Rendering Library**

FR37: Load a template from a file.
FR38: Render a template plus data to PDF bytes in-process, through an "extremely simple" API — a load call and a render call.
FR39: Render to an `io.Writer`, so a caller can write a PDF directly to an HTTP response without a temporary file. An API shape, not a constant-memory guarantee.
FR40: Accept runtime parameters at render time alongside the data.
FR41: Fail with a located, actionable error — naming the template element and the data path — for malformed templates, unresolvable bindings, invalid expressions, missing glyphs, and content that cannot be laid out. Non-fatal diagnostics are reported without failing the render.
FR42: Validate a template independently of rendering, so a malformed `.folio` can be rejected in CI before production.
FR43: Render as a pure function of (template, data, parameters) — no wall clock, time zone, locale, environment variable, filesystem, or network on the render path.
FR44: Define overflow behaviour for absolutely-positioned content exceeding its declared bounds: clipped at the component boundary with a diagnostic. Never reflowed, never silently dropped.

**REST Rendering Service** *(optional)*

FR45: *(optional, only if capacity permits)* Expose an HTTP endpoint accepting parameters and data and responding with `application/pdf`. Template travels in the request or resolves from an operator-supplied path — no server-side report repository.

### NonFunctional Requirements

NFR1: **Byte-reproducible rendering.** Same template + data + parameters + `folio-go` version produces a byte-identical PDF regardless of OS, machine, or compilation target — including across compilation targets, so the WebAssembly preview hash-matches the native render. Decomposes into seven constraints:
- NFR1.a: Layout arithmetic is contraction-free — integer fixed-point for all position, advance, and dimension math.
- NFR1.b: Numeric emission is quantized through one normative formatter at fixed decimal precision.
- NFR1.c: Restricted numeric surface — only correctly-rounded/exact operations; `math` transcendentals prohibited, enforced by lint.
- NFR1.d: No unordered iteration reaches output; every emitted collection is deterministically sorted.
- NFR1.e: Font subsetting is byte-stable — no wall-clock timestamp; subset tags a deterministic hash of the glyph set.
- NFR1.f: Document metadata pinned — creation/modification dates and document ID omitted or content-derived; `SOURCE_DATE_EPOCH` adopted.
- NFR1.g: No encryption in MVP — a random IV per encrypted string/stream is mutually exclusive with byte-reproducibility.
- **Verification:** CI matrix must include an arm64 target alongside amd64 and wasm.

NFR2: **The browser is not the canonical renderer.** Layout authority belongs to `folio-go` alone; browser text and layout engines must never determine pagination, line breaking, or measurement — including in the design canvas.

NFR3: **Script and text support.** Correctly shape, measure, break, and render Latin, CJK, and Thai. Full OpenType GSUB/GPOS including mark-to-base positioning, without cgo. Line breaking whitespace-delimited for Latin, per-character for CJK, dictionary-based for Thai. RTL out of scope. Thai dictionary embedded in the binary and versioned with the library.

NFR4: **Memory and streaming.** `Page X of Y` requires total page count before the first page, so the pipeline is two-pass. `RenderTo`'s writer API is an ergonomic and API-stability choice, not a constant-memory guarantee. No memory ceiling or throughput target specified; 50-page statement is the working target.

NFR5: **Fidelity between design and production.** The exact preview and the production render are the same engine and produce the same bytes. No separate preview renderer that could drift.

NFR6: **Versioning.** Templates carry a format version; library ships as `folio-go v0.1`. Because golden-hash regression tests are the primary verification mechanism, any change to layout, subsetting, or emission is a breaking change for downstream test suites. Forward/backward compatibility rules undefined and need a policy before v1.

NFR7: **Font provisioning.** All fonts embedded and subsetted; subsetting byte-stable. Latin + CJK + Thai coverage requires a shipped font set embedded so the designer works fully offline. Measured budget: engine and font stack ~1.5 MB compressed, CJK face ~7.4 MB, Thai dictionary ~0.1 MB. Take the glyf/TrueType variable build over CFF/OpenType. Accepted: ~9 MB first load.

NFR8: **Privacy posture.** The draw.io model plus WebAssembly preview means templates and data never leave the user's machine during design — a property to state and protect.

### Additional Requirements

*Extracted from `ARCHITECTURE-SPINE.md` (24 ADs) and `addendum.md`.*

**🚨 STARTER TEMPLATE — impacts Epic 1 Story 1**

- **There is NO application starter template for the Go engine.** The architecture commits to Folio emitting its own PDF (AD-6) — no third-party PDF writer is a dependency — so the Go module is initialized from scratch: `go mod init github.com/panitw/folio/folio-go`, Go toolchain **pinned to an exact 1.26.x** via the `toolchain` directive (AD-22).
- **The designer is scaffolded from the standard Vite React-TS template** (`npm create vite`, React 19.2.x / Vite 7.3.x, Node 20.19+ or 22.12+) — the only "starter" in the project, and it pre-decides nothing architectural because the engine holds document state (AD-15).
- **Repository layout is a single repo**: Go module at root, designer in `designer/`, font binaries in `fonts/` as the single source of truth.

**Determinism infrastructure — must be wired before the first feature lands (AD-1, AD-21)**

- Import-restriction lint banning `time`, `os`, `math/rand`, `net`, and all `math` transcendentals under `internal/`.
- Map-iteration check confirming no unordered iteration reaches PDF emission.
- CI matrix: `darwin/arm64`, `linux/amd64`, **`linux/arm64`**, and `js/wasm` under Node — hashes compared across all four.
- Golden-hash fixtures recorded per `folio-go` version **and** per Go toolchain version.
- No change lands without a fixture covering it.

**Core engine structure**

- Paradigm: functional core (`internal/`), imperative shell (`folio`, `cmd/folio`, `wasm/`, `designer/`).
- `internal/geom` defines `Length` as `int64` millipoints (1/1000 pt) and is the only package declaring a geometric scalar type (AD-2). No `float64` anywhere under `internal/`.
- Numbers reach PDF output in exactly two representations — a decimal emitter for geometric values and an integer writer for structural counts/offsets — both in one file; no other code writes a number to output (AD-3, as amended under D-1.1.b/D-000.6).
- Two-pass pipeline; pass two performs no measurement, line breaking, or pagination. Page numbers are late-bound slots. No expression may reference pagination (AD-4).
- `internal/pagemodel` is renderer-agnostic; `internal/layout` may not import `internal/pdf` (AD-5).
- Report data numbers parsed as exact scaled-integer decimals via `UseNumber`, never `float64` (AD-23).
- Element x/y is relative to its band's top-left; PageModel is top-left origin Y-down; the flip to PDF space happens in exactly one function (AD-24).

**Format and contracts**

- `.folio` has exactly one canonical serialization: sorted keys, two-space indent, LF, no trailing whitespace, trailing newline. Round-trip properties ship with the format (AD-9).
- Images: content-addressed `assets` map keyed by SHA-256, base64 hard-wrapped at 76 columns into an array of strings (AD-9).
- Element ids: opaque short strings from a monotonic per-document counter persisted in the template. No UUIDs (AD-10).
- Row scope: explicit alias — `{"bind":"transactions[]","as":"transaction"}`, default `row`. A row never shadows the root. Aggregates always whole-collection, never page-scoped (AD-11) — **resolves PRD Q3**.
- Locale: document declares one locale tag and fixed UTC offset; embedded table for a closed set `en`/`th`/`zh-Hans`/`ja`; `th` renders Buddhist-era years (AD-12) — **resolves PRD Q10**.
- PDF 1.7 target; content-derived `/ID`; `/CreationDate` and `/ModDate` omitted unless supplied via params; `cmd/folio` reads `SOURCE_DATE_EPOCH` and passes it in (AD-7) — **resolves PRD Q11**.
- One `Diagnostic` type: `Severity`, stable code from a closed registry, optional element id, optional data path, message. Absent path = Error; JSON `null` = empty; wrong-typed value = Error (AD-14).
- Table width is derived from the sum of column widths, never stored (AD-13).

**Fonts and text**

- The engine never embeds font data; `Render` takes an explicit `FontSet` (AD-8). Fonts live once in `fonts/`; the designer build copies from there.
- Ordered fallback chain declared per family and part of the FontSet's identity. Uncovered glyph = diagnostic naming element and rune.
- One subset per font per **document**, over the union of glyphs used.
- Shaping/subsetting: `boxesandglue/textshape` v0.0.15. Cross-validation only: `go-text/typesetting` v0.3.4.
- **Thai line breaking must be built** — embed **PyThaiNLP `words_th`** (62,106 words, CC0-1.0) as a `BytesTrie`, longest-match/DAG segmentation. Prerequisite of text wrapping, **not** a refinement of it. The deliverable is **break opportunities, not word segmentation** (AD-25): unknown runs are atomic and no break falls inside a Thai character cluster, so an unrecognised customer name overflows visibly under FR44 rather than being silently mis-split.
- Fonts: Noto Sans, Noto Sans Thai, Noto Sans SC (variable, glyf), all OFL 1.1.

**Licensing — PRD Q7 RESOLVED: MIT, open source (AD-26)**

- Folio ships under **MIT**. **No dependency may carry GPL, LGPL, AGPL, SSPL, or a commercial EULA at any depth** — Go links statically, so such a dependency attaches its obligations to the whole binary.
- A **CI licence check** over the Go module graph and the `designer/` lockfile fails the build rather than warning. This is not hypothetical: the only live Go Thai segmenter is MIT on its own tin but derived from LGPL-3.0 code.
- Redistributed assets keep their own terms and notices — Noto and IBM Plex under **SIL OFL 1.1** with licence text and copyright lines; `pdfjs-dist` under **Apache-2.0** with its NOTICE. A third-party licence manifest is a release artifact.

**Designer / wasm**

- The wasm engine **owns** the template document; TypeScript holds an immutable snapshot and sends committed mutations as commands. There is no TypeScript model of a `.folio` document (AD-15).
- One wasm module, one instance, one dedicated Worker. Render entry point consumes serialized `.folio` bytes, never a live document (AD-16).
- Preview render takes **three** inputs: template bytes, sample data, and an author-edited **parameter document** (AD-16).
- Preview identity = hash of (template ∥ data ∥ params ∥ engine version ∥ FontSet identity); mismatch marks stale (AD-18).
- Service worker precaches app shell, wasm, Thai dictionary, and font assets under content-hashed URLs; brotli + immutable cache headers (AD-19).
- File access is two-tier and capability-detected: File System Access API where present, `<input type="file">` + download fallback elsewhere. One interface, one branch at startup (AD-20).
- Expression parser is hand-written recursive descent — no generator, no dependency. Template validation is `internal/template` itself — **no JSON Schema library** (AD-9).
- Preview surface is a controlled `pdfjs-dist` 6.2.x canvas, never the browser's built-in viewer.

### UX Design Requirements

*Extracted from `EXPERIENCE.md` (5 surfaces, 6 component patterns, 10 state patterns) and `DESIGN.md` (60 colour tokens, 13 type roles, 15 component specs).*

**Design system foundation**

UX-DR1: Implement the DESIGN.md token file as the designer's single source of styling — 60 named colour tokens across seven groups (chrome, rules/edges, text, select, bind, status, page), 6 named rgba tint washes, 13 typography roles across two ramps (chrome and page), a 10-step spacing scale, and a zero-radius shape rule. No hard-coded hex anywhere in the app.
UX-DR2: Enforce the **two-accent grammar** in every surface without exception — `select` cyan means structure/focus/authority; `bind` amber means data, and only data. Selection handles stay cyan even on a bound element.
UX-DR3: Implement the single **dark-chrome / light-canvas** theme. Not a dual-theme system: the white page is the only bright surface. Contrast ratios must be met on dark chrome's own terms, with no light-mode fallback to lean on.
UX-DR4: Enforce the gradient rule — colour-ramp gradients forbidden; hard-stop patterns (page dot grid, transparency checkerboard) permitted.

**Surfaces**

UX-DR5: **S1 Load** — a first-run-only honest loading screen that names what is loading and why it happens once, itemising the real payload (engine 1.5 MB, Latin 0.4, Thai 0.1, CJK 7.4, Thai dictionary 0.12 = ~9.5 MB) and explaining why CJK dominates. Progress indication required; a spinner is not acceptable at this payload size. The one deliberately non-dense surface; introduces the progress-bar and manifest-row components and the only display-size type token.
UX-DR6: **S2 Workspace** — canvas, palette, properties, document bar as persistent regions. Page rendered at true proportions with visible boundaries, three bands, and a snapping grid.
UX-DR7: **S3 Binding Panel** — docked right, not a separate destination. Loads a sample JSON document and exposes its paths as a navigable tree; binding connects a path to the selected component.
UX-DR8: **S4 Table Editor** — a focused surface over the canvas, invoked from a selected Table. Column configuration presented as a **matrix** (columns × attributes), never as a repeated single-column form. The densest UI in the product.
UX-DR9: **S5 Preview** — a **mode switch, not a panel**. The canvas is replaced by the exact rendered PDF.

**Component patterns**

UX-DR10: Build the six component patterns with their DESIGN.md token counterparts — canvas component (selection handle, binding marker), band (band tab, band boundary), palette item, property field, binding tree node, table column matrix row.
UX-DR11: Make **band drop-targeting unambiguous** — the target band highlights before release. Flagged as the most consequential ambiguity on the canvas: which band a component lands in determines whether it repeats on every page, so an ambiguous boundary produces wrong output rather than a cosmetic error.
UX-DR12: Implement the `error-card` spec, marked "specified, not mocked" — no artboard shows a failed render, so `colors.danger` is unverified by any mockup.

**State patterns**

UX-DR13: Implement all ten declared states per surface: Loading (first run), Empty—no template, Empty—no sample data, Empty—table with no columns, Populated, Rendering, Diagnostic, Error, Unsaved changes, Stale preview.
UX-DR14: **Stale preview is the one to get right** — the preview must never silently present a stale document as the production artefact. Either invalidate visibly or re-render. Every other state failure is an inconvenience; this one breaks the product's central promise.
UX-DR15: Rendering state blocks **within the preview surface only** — the canvas remains interactive, and progress is indicated for a 50-page document.
UX-DR16: Unsaved-changes indicator is persistent, quiet, and always visible. **There is no autosave** in either file-access tier.
UX-DR17: Diagnostic state is non-blocking and dismissible, and **locates the offending element** back on the canvas. Error state blocks within Preview, names element and path, and offers return to canvas.

**Interaction**

UX-DR18: Implement the interaction primitives — select (click, shift-click extends, empty-canvas click clears), move (drag, grid-snapped), resize (handles, constrained to the containing band), drop (palette to band with pre-release highlight), bind, mode switch (preserves selection and scroll position), open/save.
UX-DR19: Provide shortcuts for save, undo, redo, delete, duplicate, nudge, toggle preview, toggle snapping, with hints in menus and tooltips. **No command palette in MVP.** Mouse-first with shortcuts, not keyboard-first.
UX-DR20: Undo covers every canvas and property mutation. **Loading sample data is not undoable and must not appear to be.**

**Determinism in the interface**

UX-DR21: Make the canvas-approximate / preview-exact asymmetry legible **without a tutorial**. A user must never believe the canvas is authoritative.
UX-DR22: Surface output-affecting diagnostics — clipped element, over-tall row, glyph with no coverage — in the preview where the consequence is visible, and locate back to the offending element on the canvas.
UX-DR23: **Nothing in the interface may imply server rendering, a cloud round-trip, or an account.** The engine runs in the tab; the design must not borrow the visual language of remote processing.

**Voice and accessibility**

UX-DR24: Voice is **terse and technical** — states the fact, names the location, offers no comfort (e.g. `No binding for transactions[].amount`).
UX-DR25: Meet the accessibility floor as behavioural obligations, independent of formal conformance being out of scope: every interactive element keyboard-reachable and operable (palette, properties, binding tree, table columns); visible focus on every focusable element using `colors.select`; accessible names on every icon-only control; errors and diagnostics announced and distinguished by **shape before colour**; canvas handle hit targets larger than their visual footprint; the Table Editor behaving as a data grid under keyboard navigation.

### FR Coverage Map

### FR Coverage Map

Every FR maps to exactly one epic. FR45 is the only requirement not scheduled — it is optional
in the PRD and a Non-goal in `SPEC.md`.

| FR | Epic | What lands |
|---|---|---|
| FR11 | Epic 1 | `.folio` as a portable versioned text file |
| FR12 | Epic 1 | Diffable, mergeable, agent-editable — canonical serialization |
| FR13 | Epic 1 | Format version validated on load |
| FR14 | Epic 1 | JSON report data accepted; no database |
| FR15 | Epic 1 | Scalar binding by dotted path |
| FR29 | Epic 1 | PDF output |
| FR32 | Epic 1 | Font embedding + byte-stable subsetting (Latin) |
| FR33 | Epic 1 | Images embedded in the template, never fetched |
| FR37 | Epic 1 | Load a template from a file |
| FR38 | Epic 1 | Render to PDF bytes in-process |
| FR39 | Epic 1 | Render to an `io.Writer` |
| FR40 | Epic 1 | Runtime parameters accepted at render time |
| FR43 | Epic 1 | Render is a pure function |
| FR6 | Epic 2 | Three bands: Page Header, Content, Page Footer |
| FR30 | Epic 2 | Multi-page with headers and footers on every page |
| FR31 | Epic 2 | Page numbers including `Page X of Y` |
| FR44 | Epic 2 | Overflow clipped at the component boundary with a diagnostic |
| FR16 | Epic 3 | Collection binding |
| FR17 | Epic 3 | Row scope inside a repeating region |
| FR18 | Epic 3 | The eight expression functions |
| FR19 | Epic 3 | Aggregation over a collection field |
| FR20 | Epic 3 | Conditional visibility |
| FR21 | Epic 3 | Runtime parameter namespace |
| FR41 | Epic 3 | Located, actionable errors |
| FR42 | Epic 3 | Standalone template validation |
| FR22 | Epic 4 | Rows generated from a bound collection |
| FR23 | Epic 4 | Fixed column widths with per-column alignment |
| FR24 | Epic 4 | Header row, cell borders, cell padding |
| FR25 | Epic 4 | The page-break contract (5 sub-rules) |
| FR26 | Epic 4 | Header repeated on every continuation page |
| FR27 | Epic 4 | Footer aggregates — sum, count, average |
| FR28 | Epic 4 | Alternating row styling *(if capacity permits)* |
| FR1 | Epic 5 | Visual canvas, absolute placement, drag and resize |
| FR2 | Epic 5 | Page setup — size, orientation, margins |
| FR3 | Epic 5 | Grid with snapping |
| FR4 | Epic 5 | The five-component palette |
| FR5 | Epic 5 | Component property editing |
| FR8 | Epic 5 | Local open and save, no account |
| FR34 | Epic 5 | Approximate Design Canvas view |
| FR35 | Epic 5 | Exact PDF Preview via wasm |
| FR36 | Epic 5 | Fully offline operation |
| FR7 | Epic 6 | Bind a component to a JSON path via the UI |
| FR9 | Epic 6 | Load a sample JSON document |
| FR10 | Epic 6 | Table structure editing |
| FR45 | — | *Not scheduled. Optional in PRD, Non-goal in SPEC.* |

**NFR coverage:** NFR1 and NFR6 are established in Epic 1 and defended by every epic thereafter
(AD-21: no change lands without a fixture). NFR3 and NFR7 land in Epic 2. NFR4 lands in Epic 2
(two-pass). NFR2, NFR5 and NFR8 land in Epic 5.

**UX-DR coverage:** UX-DR1–6, 9, 10, 11, 13–19, 21–25 land in Epic 5. UX-DR7, 8, 12, 20 land in
Epic 6.

## Epic List

**Six epics.** Engine first, designer second — but not the source plan's order. Two of the PRD's
open questions are resolved by this sequencing and are recorded below.

> **PRD Q5 — RESOLVED.** The source plan put tables at Phase 3 and the expression engine at
> Phase 4, but table footer aggregates *are* `sum()`, `count()` and `avg()`. The plan's order was
> simply wrong. Expressions (Epic 3) now precede tables (Epic 4). Scalar binding moves earlier
> still, into Epic 1, because a template that cannot resolve `{{customer.name}}` renders nothing
> worth hashing.

> **PRD Q6 — RESOLVED, and partially against the source plan.** Engine-first stands: counter-metric
> C4 is honoured, and no designer work begins until Epic 4 renders the golden report. But the
> designer is *not* one trailing phase. It is two epics, and the exact preview ships **with** the
> first of them rather than after it. The source plan's Phase 6 — preview last — is rejected:
> building a canvas you cannot preview with means authoring blind, and it would leave NFR2 and
> NFR5 unvalidated until the final phase of the project.

> **File-churn check — consolidation considered and rejected.** Two overlaps exist. Epics 2 and 4
> both touch `internal/layout`, but band layout and table layout are separate files with no
> feedback loop between them; merging would produce one epic carrying both R2 and the MVP's
> highest-risk engineering area. Epics 5 and 6 both touch `folio-designer/`, but canvas and
> chrome versus binding panels and table editor are distinct surfaces. In both cases the split
> buys a genuine risk boundary, so it stands.

> **Build-order gap closed.** Thai line breaking appears in no phase of the source plan. It is a
> **prerequisite** of Epic 2's text wrapping, not a refinement of it, and is scheduled as such.

### Epic 1: A Go developer can render a deterministic PDF

Anan adds `folio-go` to his module, loads a `.folio` file, renders it with his JSON data, and gets
PDF bytes — with the same SHA-256 on his Mac, on the Linux CI runner, on arm64, and under Node.
The library is integrable from this epic onward; everything after it adds capability without
changing the contract.

This epic is deliberately whole. Byte-identity cannot be retrofitted: the fixed-point unit, the
single numeric emitter, the ordering discipline, and the four-target harness have to exist before
the first feature, or every feature after them has to be re-verified. It carries the project's
Critical risk (R1) and retires it on the smallest possible artefact.

**FRs covered:** FR11, FR12, FR13, FR14, FR15, FR29, FR32, FR33, FR37, FR38, FR39, FR40, FR43
**Also lands:** NFR1.a–g, NFR6, AD-1, AD-2, AD-3, AD-6, AD-7, AD-9, AD-10, AD-21, AD-22, AD-23, AD-26

### Epic 2: A Go developer can render real multi-page documents in three scripts

A statement with a page header and footer on every page, correct `Page X of Y`, and Latin, Thai
and CJK text that wraps and breaks where a human would break it. Thai marks sit on their base
characters; Thai words break by dictionary; CJK breaks per character.

Split from Epic 1 because it carries the project's second-highest risk (R2: Thai segmentation must
be written from scratch, with no usable Go library to lean on), and because the two-pass pipeline
is the structural decision that `Page X of Y` forces.

**FRs covered:** FR6, FR30, FR31, FR44
**Also lands:** NFR3, NFR4, NFR7, AD-4, AD-5, AD-8, AD-24, AD-25
**Opens with:** a Thai segmenter spike gated on a corpus that includes Thai proper nouns, before the rest of the epic commits.

### Epic 3: A Go developer can render computed, parameterised documents

Values that are calculated rather than copied — totals, averages, formatted dates and numbers,
conditionally hidden components — and runtime parameters supplied separately from report data. A
Thai statement shows a Buddhist-era date. A malformed template is rejected in CI with a message
naming the element and the path, before it ever reaches production.

**FRs covered:** FR16, FR17, FR18, FR19, FR20, FR21, FR41, FR42
**Also lands:** AD-11, AD-12, AD-14

### Epic 4: A Go developer can render the golden report

The transaction table: rows from a bound collection, fixed columns, a header repeated on every
continuation page, sum footers that never orphan, and rows that never split across a page. At the
end of this epic the Customer Account Statement renders at 1, 5, 20 and 50 pages with identical
hashes on all four targets.

The PRD calls table pagination the highest-risk engineering area in the MVP and rules that correct
pagination beats table styling wherever the two compete. **This epic is the C4 gate: designer work
begins only once it passes.**

**FRs covered:** FR22, FR23, FR24, FR25, FR26, FR27, FR28
**Also lands:** AD-13, and success measures S1, S2, S3, S4

### Epic 5: A template author can lay out a report and see the real PDF

Ploy opens Folio in a browser, opens a `.folio` from her laptop, sets up an A4 page, drops a logo
and text fields into the three bands, and switches to Preview to see the exact production document
— the same engine, in her tab, with nothing sent anywhere. It works with the network disconnected.

Preview ships in this epic, not after it. The canvas-approximate / preview-exact asymmetry is the
product concept, and a canvas without a preview cannot validate NFR2 or NFR5.

**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR8, FR34, FR35, FR36
**Also lands:** NFR2, NFR5, NFR8, AD-15, AD-16, AD-17, AD-18, AD-19, AD-20
**UX-DRs:** UX-DR1–6, 9, 10, 11, 13–19, 21–25

### Epic 6: A template author can bind a report to data and build the golden report

Ploy loads the sample JSON her developer gave her, walks its paths as a tree, binds each field by
picking rather than typing, drops a table into the Content band, and configures its five columns
as a matrix — widths, alignment, headers, row-scoped bindings, and a sum footer. She saves the file
and commits it. Anan renders it, and the hash matches what she saw.

This epic closes the loop the whole product exists to prove, and ends on success measure S5.

**FRs covered:** FR7, FR9, FR10
**UX-DRs:** UX-DR7, 8, 12, 20
**Also lands:** success measures S5, S7, S8, S9

---

## Epic 1: A Go developer can render a deterministic PDF

Anan adds `folio-go` to his module, loads a `.folio` file, renders it with his JSON data, and gets PDF bytes — with the same SHA-256 on his Mac, on the Linux CI runner, on arm64, and under Node. The library is integrable from this epic onward; everything after it adds capability without changing the contract.

### Story 1.1: A minimal PDF, reproducible on one machine

As an integrating Go developer,
I want `folio-go` to emit a minimal valid PDF whose bytes are identical every time I render the same input,
So that hash comparison becomes my regression test before any feature exists to regress.

**Covers:** FR29, FR43 · NFR1.a, NFR1.b, NFR1.d, NFR1.f, NFR6 · AD-1, AD-2, AD-3, AD-6, AD-7, AD-22

**Acceptance Criteria:**

**Given** a new Go module `github.com/panitw/folio/folio-go` with an exact `toolchain go1.26.x` directive
**When** the module is built
**Then** the pinned toolchain version is used, and the version is recorded in a file the golden fixtures reference
**And** no dependency on any third-party PDF writer exists in `go.mod`

**Given** the `internal/geom` package
**When** any position, advance, or dimension is computed
**Then** it is an `int64` `Length` in millipoints (1/1000 pt)
**And** `internal/geom` is the only package declaring a geometric scalar type
**And** font scaling is a single exported function with round-half-to-even on the exact integer quotient

**Given** a hard-coded page containing one filled rectangle
**When** `Render` is called
**Then** a valid PDF 1.7 document is produced with catalog, page tree, content stream, and cross-reference table
**And** every geometric number in the output was written by the single unexported decimal emitter as sign + integer part + up to three trimmed fractional digits, with structural counts, offsets and object/generation numbers routed through a separate integer writer in the same file (D-1.1.b: AD-3 governs numeric representations, not the number of functions)
**And** no `float64` appears in any signature under `internal/`

**Given** the same input rendered twice in two separate OS processes
**When** the two outputs are compared
**Then** they are byte-identical
**And** `/CreationDate` and `/ModDate` are absent
**And** `/ID` is the first 16 bytes of a SHA-256 over the pre-`/ID` body, with both array entries identical

**Given** the render completes
**When** the recorded golden hash is compared
**Then** it matches, and the fixture records both the `folio-go` version and the Go toolchain version

### Story 1.2: Cross-target byte-identity proven in CI

As an integrating Go developer,
I want CI to prove the same render produces identical bytes on every architecture and compilation target,
So that the PDF my Linux service returns is provably the same file that was verified on a developer's Mac.

**Covers:** NFR1 · AD-21 · retires Risk R1

**Acceptance Criteria:**

**Given** the CI configuration
**When** a build runs
**Then** it executes on `darwin/arm64`, `linux/amd64`, `linux/arm64`, and `js/wasm`
**And** the `js/wasm` target is executed under Node or an equivalent runtime
**And** `CGO_ENABLED=0` on every target

**Given** Story 1.1's minimal document rendered on all four targets
**When** the four outputs are hashed
**Then** all four hashes are identical
**And** the job fails if any pair differs

**Given** a deliberately introduced floating-point multiply-add in layout arithmetic
**When** CI runs
**Then** the arm64 job produces a different hash from the wasm job and the build fails
**And** this negative test is retained as proof the matrix can actually detect contraction

**Given** the reference render used for comparison
**When** it is validated
**Then** it is confirmed non-blank and page-count-correct, so hash equality cannot pass on two identical empty files

### Story 1.3: Guardrails that fail the build

As an integrating Go developer,
I want the determinism and licence rules enforced mechanically rather than by convention,
So that a reasonable-looking commit cannot quietly erode the property the whole library rests on.

**Covers:** NFR1.c, NFR1.d · AD-1, AD-26

**Acceptance Criteria:**

**Given** the import-restriction lint
**When** any package under `internal/` imports `time`, `os`, `math/rand`, `net`, or a `math` transcendental (`Sin`, `Cos`, `Tan`, `Log`, `Exp`, `Pow`, `Sinh`, `Erf`, …)
**Then** the build fails naming the file and the offending import
**And** the allow-listed numeric surface is `+ - * /`, comparison, `Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`

**Given** the map-iteration check
**When** a `range` over a map can reach an output byte
**Then** the build fails naming the location

**Given** the licence check
**When** any dependency in the Go module graph or the `designer/` lockfile carries GPL, LGPL, AGPL, SSPL, or a commercial EULA at any depth
**Then** the build fails rather than warning
**And** a third-party licence manifest is produced as a release artifact, carrying OFL 1.1 text for shipped faces and the Apache-2.0 NOTICE for `pdfjs-dist`

**Given** each guardrail
**When** it is added
**Then** a deliberately violating fixture proves it fires, and the fixture is retained

### Story 1.4: Load, validate, and round-trip a `.folio` template

As an integrating Go developer,
I want to load a `.folio` file and have malformed or future-versioned templates rejected with a located error,
So that a bad template is caught in CI rather than in production.

**Covers:** FR11, FR12, FR13, FR37 · NFR6 · AD-9, AD-10

**Acceptance Criteria:**

**Given** a well-formed `.folio` file carrying a `version`, page setup, and ordered band content
**When** `folio.LoadTemplate(path)` is called
**Then** a `Template` is returned with no error

**Given** a template declaring a format version newer than the library supports
**When** it is loaded
**Then** loading fails with an error naming the declared version and the supported version
**And** no render is attempted

**Given** any parsed template `d` and any canonical byte sequence `b`
**When** round-trip properties are tested
**Then** `Parse(Serialize(d)) == d` and `Serialize(Parse(b)) == b` hold
**And** these tests ship with the format, not after it

**Given** a template serialized by the library
**When** the bytes are inspected
**Then** object keys are sorted, indentation is two spaces, line endings are LF, there is no trailing whitespace, and the file ends with a newline

**Given** any element in a loaded template
**When** its identity is inspected
**Then** it carries an opaque short id from a monotonic counter persisted in the document
**And** ids are not UUIDs, are never reused, and are unchanged by a save

### Story 1.5: Embed and subset a font, byte-stably

As an integrating Go developer,
I want every font the document uses embedded as a subset that is identical on every run,
So that the PDF renders correctly on a machine with no fonts installed without breaking hash equality.

**Covers:** FR32 · NFR1.e, NFR7 · AD-8

**Acceptance Criteria:**

**Given** a `FontSet` value supplied to `Render`
**When** the render executes
**Then** fonts are resolved by pure lookup against that `FontSet`
**And** no package under `internal/` embeds font data
**And** no host font query occurs

**Given** a document using a subset of a Latin face
**When** the PDF is produced
**Then** the face is embedded as a `Type0`/`Identity-H` composite font with a `FontFile2` and a `ToUnicode` CMap
**And** the six-letter subset tag is derived from a hash of the sorted glyph-id set
**And** one subset is emitted per font per document over the union of glyphs used, never per page

**Given** the same document rendered in two separate processes
**When** the embedded font programs are compared
**Then** they are byte-identical
**And** no wall-clock timestamp appears in the subset output

**Given** the resulting PDF opened on a machine with none of its fonts installed
**When** it is displayed
**Then** the text renders correctly

### Story 1.6: Bind scalar JSON values into text

As an integrating Go developer,
I want to supply my data as JSON and have `{{customer.name}}` resolve to the right value,
So that the template is a template rather than a fixed document.

**Covers:** FR14, FR15, FR38, FR43 · AD-14, AD-23

**Acceptance Criteria:**

**Given** a JSON document supplied as report data
**When** it is decoded
**Then** number literals are preserved and converted to exact scaled-integer decimals carrying the literal's own precision
**And** no `float64` is produced anywhere in the binding path
**And** a literal too large for the representation is an error, never a silent narrowing

**Given** a text element bound to `{{customer.name}}`
**When** the document renders
**Then** the value at that dotted path appears in the output

**Given** a binding whose path is absent from the data
**When** the document renders
**Then** the render fails with an error naming the path and the element id

**Given** a binding whose value is JSON `null`
**When** the document renders
**Then** it renders as empty and is not an error

**Given** a binding whose value is of the wrong kind for its element
**When** the document renders
**Then** the render fails with an error; the value is never coerced

**Given** a render in progress
**When** the render path is inspected
**Then** it has read no wall clock, time zone, locale, environment variable, filesystem path, or network resource

### Story 1.7: Render to an `io.Writer` with runtime parameters

As an integrating Go developer,
I want to write a PDF straight to my HTTP response and pass runtime values separately from report data,
So that I can serve a generated document without a temporary file and without smuggling dates into the data.

**Covers:** FR39, FR40

**Acceptance Criteria:**

**Given** an `http.ResponseWriter`
**When** `folio.RenderTo(w, template, data, params)` is called
**Then** PDF bytes are written directly to it with no temporary file created

**Given** the same inputs
**When** rendered through `Render` and through `RenderTo`
**Then** the resulting bytes are identical

**Given** a parameter supplied at render time
**When** a `{{params.reportDate}}` binding is resolved
**Then** it resolves from the parameter namespace, not from report data
**And** report data cannot shadow the `params.` namespace

**Given** the public API surface
**When** an integrator reads the README
**Then** producing a first PDF requires a load call and a render call, and nothing ceremonial

### Story 1.8: Embed and render an image from the template

As a template author,
I want my logo carried inside the `.folio` file itself,
So that the template is self-contained and rendering never depends on a URL or a file on disk.

**Covers:** FR33 · AD-9, AD-24

**Acceptance Criteria:**

**Given** a template containing an image
**When** the file is inspected
**Then** the image lives in a top-level `assets` object keyed by the SHA-256 of its bytes
**And** its base64 is hard-wrapped at 76 columns into an array of strings
**And** the file remains valid JSON and its non-asset content stays readable in a text diff

**Given** two elements referencing the same image
**When** the template is serialized
**Then** the bytes are stored once and both elements reference the same asset key

**Given** a template whose image is rendered
**When** the PDF is produced
**Then** the image appears as an image XObject
**And** the render made no network request and read no file from disk

**Given** an image drawn into a declared box
**When** its placement is computed
**Then** it is scaled to fit preserving aspect ratio and centred within the box, computed in integer millipoints
**And** it is never stretched, never cropped, and never sized from its intrinsic pixel dimensions

---

## Epic 2: A Go developer can render real multi-page documents in three scripts

A statement with a page header and footer on every page, correct `Page X of Y`, and Latin, Thai and CJK text that wraps and breaks where a human would break it. Thai marks sit on their base characters; Thai words break by dictionary; CJK breaks per character.

### Story 2.1: Thai break-opportunity spike

As a solo builder,
I want to prove the Thai break approach works against real Thai text including proper nouns before committing the rest of this epic,
So that the project's second-highest risk is retired cheaply rather than discovered halfway through the layout engine.

**Covers:** NFR3 · AD-25 · retires Risk R2

**Acceptance Criteria:**

**Given** the PyThaiNLP `words_th` wordlist (62,106 words, CC0-1.0)
**When** it is compiled
**Then** it is a `BytesTrie` embedded in the binary via `go:embed`
**And** it is never read from disk, and no code path uses `runtime.Caller`
**And** the trie loads and queries correctly under `js/wasm`

**Given** a Thai evaluation corpus that deliberately includes Thai personal names, place names, and transaction descriptions
**When** break opportunities are computed
**Then** results are recorded and hand-reviewed, and the review is kept as the basis of the S4 fixture

**Given** a run of Thai characters the dictionary cannot cover
**When** break opportunities are computed
**Then** the run yields no interior break opportunities and is treated as atomic
**And** no unrecognised name is split at a guess

**Given** any Thai character cluster — a consonant with its vowels and tone marks
**When** break opportunities are computed
**Then** no break falls inside it

**Given** the spike's findings
**When** they are reviewed
**Then** either the approach is confirmed and the remaining stories proceed, or the deviation is recorded and the epic is re-planned before further work

### Story 2.2: The shipped font set and its fallback chain

As a template author,
I want Latin, Thai and CJK glyphs available without installing anything,
So that my statement renders the same for me as it does on a server that has no fonts at all.

**Covers:** NFR7 · AD-8

**Acceptance Criteria:**

**Given** the repository
**When** font binaries are located
**Then** they exist in exactly one place, `fonts/`, and the `folio/fonts` package wraps them with `go:embed` for native callers

**Given** the shipped set
**When** it is inspected
**Then** it covers Latin (Noto Sans), Thai (Noto Sans Thai), and CJK (Noto Sans SC, variable, glyf outline format), all SIL OFL 1.1
**And** each face travels with its licence text and copyright lines

**Given** a template naming a font family
**When** the family is resolved
**Then** it resolves against an ordered fallback chain declared in the `FontSet`
**And** the chain is part of the `FontSet`'s identity, so the same template with a different chain is a different render

**Given** a glyph covered by no font in the chain
**When** the document renders
**Then** a diagnostic is reported naming the element id and the offending rune
**And** no blank box is silently emitted

### Story 2.3: Shape Latin, Thai and CJK text

As a template author,
I want Thai vowels and tone marks to sit correctly on their base characters,
So that my customers' names are not merely present but correct.

**Covers:** NFR3

**Acceptance Criteria:**

**Given** text in Latin, Thai or CJK
**When** it is shaped
**Then** shaping runs through `boxesandglue/textshape` with full OpenType `GSUB` and `GPOS`
**And** no cgo is used and the build succeeds for `js/wasm`

**Given** Thai text containing vowels and tone marks
**When** it is shaped
**Then** marks are positioned by `GPOS` mark-to-base positioning
**And** output is cross-validated against `go-text/typesetting` for the same input

**Given** the same text shaped in two separate processes
**When** the resulting glyph runs are compared
**Then** glyph ids and positions are byte-identical

### Story 2.4: Break and measure lines in all three scripts

As a template author,
I want text to wrap where a reader would expect it to,
So that a multi-line address or transaction description is legible rather than merely fitting.

**Covers:** NFR3 · AD-2, AD-25 · S4

**Acceptance Criteria:**

**Given** Latin text
**When** lines are broken
**Then** breaks occur at whitespace per UAX #14

**Given** CJK text
**When** lines are broken
**Then** breaks occur per character

**Given** Thai text
**When** lines are broken
**Then** breaks come from the embedded dictionary, constrained by Story 2.1's two rules — unknown runs atomic, never inside a character cluster

**Given** the frozen expected-break fixture for Thai and CJK
**When** the breaker runs against it
**Then** results match the fixture exactly
**And** the fixture is hand-checked once and then never regenerated to make a test pass

**Given** any text measurement
**When** it is computed
**Then** it is exact integer arithmetic in millipoints with no `float64` anywhere in the path

### Story 2.5: Compose a page from three bands

As a template author,
I want my report composed of a Page Header, a Content region and a Page Footer,
So that which band I place something in decides whether it repeats on every page.

**Covers:** FR6 · AD-5, AD-24

**Acceptance Criteria:**

**Given** a template
**When** its structure is inspected
**Then** it has exactly three bands — Page Header, Content, Page Footer
**And** group headers, report headers and group footers are absent and unsupported

**Given** an element within a band
**When** its coordinates are resolved
**Then** x and y are relative to that band's top-left corner, never to the page
**And** bands are placed on the page by `internal/layout` alone

**Given** the page model produced by layout
**When** it is inspected
**Then** its origin is top-left with Y increasing downward
**And** it names only geometry, glyph runs, images and vector primitives — no PDF object references, resource dictionaries or content-stream operators
**And** `internal/layout` does not import `internal/pdf`

**Given** the flip from page-model space to PDF user space
**When** the code is inspected
**Then** it occurs in exactly one function in `internal/pdf` and nowhere else

**Given** page dimensions, margins and band heights
**When** the content area is computed
**Then** it is derived by one function as page height minus margins minus header height minus footer height

### Story 2.6: Render multi-page documents with headers and footers on every page

As an integrating Go developer,
I want a document longer than one page to carry its header and footer throughout,
So that page 34 of a statement is as complete as page 1.

**Covers:** FR30 · NFR4 · AD-4

**Acceptance Criteria:**

**Given** content exceeding one page
**When** the document renders
**Then** it paginates into multiple pages
**And** the Page Header and Page Footer bands appear on every page

**Given** the pipeline
**When** it executes
**Then** pass one produces a complete page model and pass two serializes it
**And** pass two performs no measurement, no line breaking and no pagination

**Given** a document that grows
**When** the Content band overflows
**Then** it grows by producing more pages, never by displacing sibling components within a page

### Story 2.7: Render `Page X of Y`

As a template author,
I want a footer that says which page of how many the reader is holding,
So that a printed statement is verifiably complete.

**Covers:** FR31 · AD-4

**Acceptance Criteria:**

**Given** a footer containing a `Page X of Y` construct
**When** the document renders
**Then** X is the current page and Y the document total, correct on every page

**Given** the page model after pass one
**When** page-number text is inspected
**Then** it is carried as a late-bound slot whose box is already measured
**And** it is resolved between the passes by substituting pre-measured glyphs

**Given** the expression language
**When** an author attempts to reference pagination
**Then** no `page` namespace exists and none can be added

**Given** documents of 1, 5, 20 and 50 pages
**When** each renders
**Then** `Page X of Y` is correct throughout and hashes match recorded goldens on all four targets

### Story 2.8: Clip overflowing content and say so

As a template author,
I want content that exceeds its box cut at the boundary with a diagnostic rather than silently lost,
So that I find out on the preview instead of from a customer.

**Covers:** FR44 · AD-14, AD-24

**Acceptance Criteria:**

**Given** absolutely-positioned content exceeding its declared bounds
**When** the document renders
**Then** it is clipped at the component boundary
**And** a diagnostic is reported naming the element id
**And** it is never reflowed and never silently dropped

**Given** the diagnostic
**When** the render completes
**Then** PDF bytes are still returned — clipping degrades output but does not fail the render

---

## Epic 3: A Go developer can render computed, parameterised documents

Values that are calculated rather than copied — totals, averages, formatted dates and numbers, conditionally hidden components — and runtime parameters supplied separately from report data.

### Story 3.1: Bind a repeating region to a collection with an explicit row scope

As a template author,
I want to bind a region to `transactions[]` and address each row's fields unambiguously,
So that the rule for what `{{transaction.amount}}` means is written down rather than guessed.

**Covers:** FR16, FR17, FR21 · AD-11

**Acceptance Criteria:**

**Given** a repeating region declaring `{"bind": "transactions[]", "as": "transaction"}`
**When** its contents are evaluated for a row
**Then** `transaction.*` resolves to that row's fields

**Given** a repeating region that omits `as`
**When** its contents are evaluated
**Then** the alias defaults to `row`

**Given** an unqualified path inside a repeating region
**When** it is resolved
**Then** it resolves from the document root — a row never shadows the root

**Given** a `params.` path anywhere, including inside a row scope
**When** it is resolved
**Then** it resolves from the parameter namespace and can be shadowed by nothing

**Given** a collection binding whose path is not an array
**When** the document renders
**Then** the render fails with an error naming the path and the element id

### Story 3.2: Evaluate the expression language

As a template author,
I want a small set of expressions inside my bindings,
So that I can compute a value without the template becoming a program.

**Covers:** FR18 · AD-9 · counter-metric C1

**Acceptance Criteria:**

**Given** the expression implementation
**When** it is inspected
**Then** it is a hand-written recursive-descent parser with no generator and no third-party dependency
**And** the function table is closed and contains exactly eight entries

**Given** `upper()`, `lower()` and `if()`
**When** they are evaluated
**Then** each behaves per its specification and has tests covering its edge cases

**Given** an expression with a syntax error
**When** the template is loaded
**Then** it fails with an error naming the element id and the offending expression text

**Given** an attempt to add a ninth function
**When** CI runs
**Then** the closed function table makes the addition visible in a diff, per counter-metric C1

### Story 3.3: Aggregate over a collection

As a template author,
I want a statement total that adds up,
So that the sum footer on my transaction table is correct to the last satang.

**Covers:** FR19 · AD-11, AD-23

**Acceptance Criteria:**

**Given** `{{sum(transactions.amount)}}`
**When** it is evaluated
**Then** the result matches a hand-computed total exactly
**And** the arithmetic is exact decimal, never binary floating point

**Given** `count()` and `avg()`
**When** they are evaluated
**Then** `count()` returns the collection length and `avg()` divides at a defined scale with round-half-to-even

**Given** an aggregate evaluated inside a row scope
**When** it resolves
**Then** it is computed over the whole collection, never over the rows on the current page
**And** no per-page subtotal behaviour exists

**Given** an aggregate over an empty collection
**When** it is evaluated
**Then** `sum` and `count` return zero and `avg` reports a diagnostic rather than dividing by zero

### Story 3.4: Format dates and numbers by declared locale

As a template author in Bangkok,
I want a Thai statement to show a Buddhist-era date and correctly grouped numbers,
So that the document reads as a Thai document rather than a translated one.

**Covers:** FR18 · AD-1, AD-12

**Acceptance Criteria:**

**Given** a document declaring a `locale` tag and a fixed UTC offset
**When** `formatDate` or `formatNumber` is evaluated
**Then** symbols and calendar come from the embedded locale table for that tag
**And** no host locale, host time zone or host clock is consulted

**Given** the locale table
**When** it is inspected
**Then** it covers exactly `en`, `th`, `zh-Hans` and `ja`
**And** an unlisted tag is a load error, never a silent fallback

**Given** the `th` locale
**When** a date is formatted
**Then** the year is rendered in the Buddhist era

**Given** a date value
**When** it is passed to `formatDate`
**Then** only an RFC 3339 string or an epoch-millisecond number is accepted; anything else is an error

**Given** `formatNumber` scaling by powers of ten
**When** the implementation is inspected
**Then** it uses an integer lookup table and never `math.Pow`

### Story 3.5: Hide a component conditionally

As a template author,
I want to hide a component when the data says it is irrelevant,
So that an empty section does not print as an empty box.

**Covers:** FR20 · AD-24

**Acceptance Criteria:**

**Given** a component with a visibility condition
**When** the condition evaluates false
**Then** the component is absent from the page model entirely

**Given** a hidden component
**When** the page is laid out
**Then** its siblings do not move — nothing in a band ever reflows

**Given** a visibility condition applied to a table row
**When** the template is validated
**Then** it is rejected: visibility applies to elements only, because row-level visibility would make pagination a function of data

**Given** conditional formatting — data-driven styling
**When** it is attempted
**Then** it is unsupported and out of scope; only visibility is conditional

### Story 3.6: Fail with located, actionable diagnostics

As an integrating Go developer,
I want every failure to tell me which element and which data path caused it,
So that debugging a template is reading one message rather than bisecting a document.

**Covers:** FR41 · AD-14

**Acceptance Criteria:**

**Given** the diagnostic type
**When** it is inspected
**Then** one `Diagnostic` value carries `Severity`, a stable code from a closed registry, an optional element id, an optional data path, and a message

**Given** severity
**When** a render runs
**Then** `Error` aborts it and `Warning` accompanies successfully returned PDF bytes

**Given** each failure mode named in FR41 — malformed template, unresolvable binding, invalid expression, missing glyph, unlayoutable content
**When** it occurs
**Then** it produces a distinct stable code, and a test asserts the message locates the problem

**Given** a caller matching on a failure
**When** they write the check
**Then** they match on the code, never on message text

**Given** an existing diagnostic code
**When** the registry changes
**Then** codes are additive only; changing a code's meaning is a breaking change

### Story 3.7: Validate a template and render it from the command line

As an integrating Go developer,
I want to reject a malformed template in CI before it reaches production, and to pin document dates reproducibly,
So that a bad template fails my pipeline rather than my customers' statements.

**Covers:** FR42 · AD-7

**Acceptance Criteria:**

**Given** a malformed `.folio` file
**When** `folio.Validate` is called
**Then** it is rejected with located diagnostics and no render is attempted

**Given** `cmd/folio`
**When** it is run
**Then** it offers a validate command and a render command

**Given** `SOURCE_DATE_EPOCH` set in the environment
**When** `cmd/folio render` runs
**Then** the CLI reads it and passes it in as a parameter
**And** the library core still reads no environment variable

**Given** no date supplied by any route
**When** the PDF is produced
**Then** `/CreationDate` and `/ModDate` are omitted

---

## Epic 4: A Go developer can render the golden report

The transaction table: rows from a bound collection, fixed columns, a header repeated on every continuation page, sum footers that never orphan, and rows that never split across a page. **This epic is the C4 gate: designer work begins only once it passes.**

### Story 4.1: Lay out a table's columns, header and cell chrome

As a template author,
I want fixed-width columns with their own alignment, a header row, borders and padding,
So that a transaction table reads as a table rather than as text in rows.

**Covers:** FR23, FR24 · AD-13

**Acceptance Criteria:**

**Given** a table with declared column widths
**When** its geometry is computed
**Then** the table's width is the sum of its column widths
**And** no independent table-width field is stored anywhere

**Given** each column
**When** it is laid out
**Then** it honours its own alignment and its declared width
**And** column widths are never negotiated against content

**Given** a table
**When** it renders
**Then** a header row, cell borders and cell padding are produced per the template's styling

### Story 4.2: Generate rows from a bound collection

As a template author,
I want one row per item in my data,
So that a statement with 400 transactions prints 400 rows without me placing any of them.

**Covers:** FR22, FR25 · AD-11

**Acceptance Criteria:**

**Given** a table bound to a collection
**When** the document renders
**Then** one row is generated per element of the collection, in data order

**Given** a cell whose content exceeds its column width
**When** the row is laid out
**Then** the content wraps within the column width and the row grows taller

**Given** each cell binding
**When** it is resolved
**Then** it resolves within the table's row scope per Story 3.1

**Given** a table bound to an empty collection
**When** the document renders
**Then** the header renders, no data rows are produced, and the render succeeds

### Story 4.3: Break a table across pages without splitting a row

As a template author,
I want a row to stay whole,
So that a transaction never appears half on one page and half on the next.

**Covers:** FR25

**Acceptance Criteria:**

**Given** a row that does not fit in the remaining content height
**When** pagination runs
**Then** the row moves whole to the next page
**And** no row is ever split across a page boundary

**Given** a table longer than one page
**When** it paginates
**Then** rows continue in data order across pages with no duplication and no omission

### Story 4.4: Repeat the table header on every continuation page

As a reader of a 50-page statement,
I want to know what each column means on page 34,
So that I do not have to page back to the start to read my own statement.

**Covers:** FR26

**Acceptance Criteria:**

**Given** a table continuing onto a further page
**When** that page renders
**Then** the table header row is repeated at the top of the table on that page

**Given** the repeated header
**When** the content height available for data rows is computed
**Then** the header's height is accounted for on every continuation page, not only the first

### Story 4.5: Render footer aggregates without orphaning them

As a template author,
I want the sum row to stay with its data,
So that a total never appears alone at the top of a page with nothing above it.

**Covers:** FR25, FR27

**Acceptance Criteria:**

**Given** a table with per-column footer aggregates
**When** it renders
**Then** sum, count and average render per the column's configuration

**Given** a footer aggregate row that does not fit in the remaining space
**When** pagination runs
**Then** it moves to the next page together with at least one preceding data row

**Given** the aggregate values
**When** they are computed
**Then** they cover the whole collection, matching Story 3.3

### Story 4.6: Handle a row taller than the page

As an integrating Go developer,
I want a pathological row to produce a diagnostic rather than an infinite loop,
So that one bad record cannot hang my statement service.

**Covers:** FR25, FR41 · AD-14

**Acceptance Criteria:**

**Given** a single row taller than a full content area
**When** pagination runs
**Then** it renders at the top of a fresh page, clipped to that page's content height
**And** a diagnostic is reported naming the row and the element id
**And** the render completes and returns PDF bytes

**Given** the same pathological row
**When** pagination runs
**Then** the algorithm terminates — no loop and no silent truncation

### Story 4.7: The golden report renders and hashes match

As a solo builder,
I want the Customer Account Statement to render correctly and reproducibly at four page counts,
So that the MVP's central claim is demonstrated rather than asserted.

**Covers:** S1, S2, S3, S4 · the C4 gate

**Acceptance Criteria:**

**Given** the golden report fixture — logo, customer block, account block, statement period, five-column transaction table with repeated header and sum footer, and a footer carrying confidentiality text, a params-supplied generated date and `Page X of Y`
**When** it renders at 1, 5, 20 and 50 pages
**Then** all four render correctly against recorded reference renders, not merely producing a file

**Given** the fixture data
**When** it is inspected
**Then** it contains Latin, Thai and CJK text in the same table, at least one row wrapping to multiple lines, and at least one embedded image
**And** the generated date arrives through `params` and is never read from a clock

**Given** the four page counts
**When** each is rendered on `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`
**Then** every hash is identical across all four targets
**And** each reference render is confirmed non-blank and page-count-correct

**Given** the Thai text in the fixture
**When** its line breaks are compared to the frozen expected-break fixture
**Then** they match exactly, and Thai marks are positioned by `GPOS`

### Story 4.8: Alternating row styling

As a template author,
I want alternating row shading,
So that a dense transaction table is easier to read across.

**Covers:** FR28 · optional, first to be cut

**Acceptance Criteria:**

**Given** a table with alternating row styling enabled
**When** it renders
**Then** alternate rows carry the configured background

**Given** a table paginating across pages
**When** alternation is applied
**Then** the alternation follows row index in the collection, so it does not reset per page

**Given** this story
**When** capacity is assessed
**Then** it is understood as optional — correct pagination takes priority over table styling wherever the two compete, and this story is cut before any of Stories 4.1 to 4.7 are compromised

---

## Epic 5: A template author can lay out a report and see the real PDF

Ploy opens Folio in a browser, opens a `.folio` from her laptop, sets up an A4 page, drops a logo and text fields into the three bands, and switches to Preview to see the exact production document — the same engine, in her tab, with nothing sent anywhere. It works with the network disconnected.

### Story 5.1: The design system and application shell

As a template author,
I want an interface that reads as a precision instrument rather than a web page,
So that the report I am building is the only thing competing for my attention.

**Covers:** UX-DR1, UX-DR2, UX-DR3, UX-DR4 · the only starter template in the project

**Acceptance Criteria:**

**Given** a new `folio-designer/` workspace
**When** it is scaffolded
**Then** it is created from the standard Vite React-TS template (`npm create vite`), React 19.2.x, Vite 7.3.x, Node 20.19+ or 22.12+
**And** it is the only starter template in the project — the Go module is initialized from scratch (Story 1.1)
**And** TypeScript strict mode is on

**Given** the designer's styling
**When** it is inspected
**Then** every colour, type role, spacing step and radius comes from the `DESIGN.md` token file
**And** no hard-coded hex value appears anywhere in the application

**Given** the token implementation
**When** it is inspected
**Then** it carries the 60 named colour tokens across chrome, rules, text, select, bind, status and page groups; the six named rgba tint washes; the 13 typography roles across the chrome and page ramps; the 10-step spacing scale; and the zero-radius shape rule

**Given** any surface
**When** accent colour is applied
**Then** `select` cyan means structure, focus and authority, and `bind` amber means data and only data
**And** selection handles remain cyan even on a bound element

**Given** the theme
**When** it renders
**Then** it is the single dark-chrome / light-canvas theme, with the page as the only bright surface
**And** contrast ratios are met on dark chrome's own terms with no light-mode fallback

**Given** the gradient rule
**When** it is applied
**Then** colour-ramp gradients are absent, while hard-stop patterns such as the page dot grid and the transparency checkerboard are permitted

### Story 5.2: The engine worker and command channel

As a template author,
I want the application to hold exactly one copy of my document,
So that what I see, what I save and what renders can never disagree.

**Covers:** AD-15, AD-16

**Acceptance Criteria:**

**Given** the designer at runtime
**When** the engine is instantiated
**Then** exactly one wasm module instance exists, in one dedicated Worker

**Given** the document
**When** ownership is inspected
**Then** the wasm engine parses, holds, mutates, validates and serializes it
**And** no TypeScript model of a `.folio` document exists anywhere in the codebase

**Given** the user interface
**When** it paints
**Then** it holds an immutable snapshot and sends every committed mutation as a command over one channel

**Given** transient interaction state — a drag in flight, a resize preview, an uncommitted property keystroke
**When** it occurs
**Then** it lives in the UI and never enters the document

**Given** any engine call
**When** it is made
**Then** it is asynchronous over a request/response channel and the UI never blocks on it

### Story 5.3: Work offline

As a template author,
I want the designer and its preview to work with the network disconnected,
So that my templates and my customers' data never depend on anyone else's servers.

**Covers:** FR36 · NFR8 · AD-19

**Acceptance Criteria:**

**Given** a completed first load
**When** the network is disconnected and the page is reloaded
**Then** the application loads, opens a template and produces a preview

**Given** the service worker
**When** it installs
**Then** it precaches the application shell, the wasm module, the Thai dictionary and the font assets under content-hashed URLs, and serves them cache-first

**Given** the assets
**When** they are served
**Then** they carry brotli compression and immutable long-lived cache headers

**Given** the application at any point after load
**When** network activity is observed
**Then** no request is made at render or preview time
**And** no template or data byte leaves the machine

### Story 5.4: An honest first-run load screen

As a first-time user,
I want to be told what is downloading and why it happens only once,
So that a nine-megabyte wait feels like an explanation rather than a hang.

**Covers:** UX-DR5

**Acceptance Criteria:**

**Given** a first visit
**When** the application loads
**Then** a load screen names what is loading and states that it happens once per browser cache lifetime

**Given** the payload
**When** it is presented
**Then** it is itemised with real sizes — engine 1.5 MB, Latin 0.4 MB, Thai 0.1 MB, CJK 7.4 MB, Thai dictionary 0.12 MB — and explains why CJK dominates

**Given** the download in progress
**When** it is displayed
**Then** actual progress is shown; a spinner without progress is not acceptable at this payload size

**Given** this surface
**When** its density is compared to the rest of the application
**Then** it is deliberately the one non-dense surface, and it is the only place the display-size type token appears

**Given** a subsequent visit with the cache warm
**When** the application loads
**Then** the load screen does not appear

### Story 5.5: Open and save `.folio` files locally

As a template author,
I want to work on files on my own laptop with no account,
So that my templates are mine and live in my own Git repository.

**Covers:** FR8 · AD-9, AD-20 · UX-DR13, UX-DR16

**Acceptance Criteria:**

**Given** a browser implementing the File System Access API
**When** a file is opened and later saved
**Then** the file handle is held and the file is saved in place

**Given** a browser without it — Firefox or Safari, which ship only OPFS
**When** a file is opened and later saved
**Then** open falls back to `<input type="file">` and save falls back to a download
**And** the capability check happens once at startup, after which the rest of the application talks to one file-access interface and never branches again

**Given** unsaved changes
**When** the workspace is displayed
**Then** a persistent, quiet, always-visible indicator shows them
**And** no autosave exists in either tier

**Given** a file saved by the designer
**When** its bytes are compared to what the engine serializes
**Then** they are identical, because the engine performed the serialization

**Given** the application opened with no template
**When** the workspace is displayed
**Then** it shows an unnamed blank template and offers both open-a-file and start-blank
**And** the author is never left on a dead canvas with no path forward

### Story 5.6: The canvas — page, bands, grid and page setup

As a template author,
I want to see my page at its true proportions with its three bands clearly separated,
So that I always know which band I am working in.

**Covers:** FR2, FR3, FR6 · UX-DR6

**Acceptance Criteria:**

**Given** the workspace
**When** the canvas renders
**Then** the page appears at true proportions with visible page boundaries, on a dark surround

**Given** the three bands
**When** they are displayed
**Then** Page Header, Content and Page Footer are legible as distinct regions at all times

**Given** page setup
**When** it is configured
**Then** A4, Letter or a custom size may be chosen, with portrait or landscape orientation and page margins

**Given** the grid
**When** snapping is enabled
**Then** movement snaps to it, and snapping can be toggled

### Story 5.7: Place and manipulate the five components

As a template author,
I want to drop a text field or a logo onto the page and move it where I want it,
So that laying out a statement is direct manipulation rather than configuration.

**Covers:** FR1, FR4 · UX-DR10, UX-DR11, UX-DR18, UX-DR25

**Acceptance Criteria:**

**Given** the palette
**When** it is displayed
**Then** it offers exactly five components — Text, Image, Table, Line, Rectangle
**And** the set is closed and the interface does not imply extensibility

**Given** a component dragged from the palette toward a band
**When** it hovers before release
**Then** the target band highlights, so the author knows which band will receive it before committing
**And** an ambiguous drop is not possible

**Given** a placed component
**When** it is manipulated
**Then** click selects, shift-click extends, clicking empty canvas clears, drag moves with grid snapping, and handles resize
**And** resizing is constrained to the containing band

**Given** a selection handle
**When** its hit target is measured
**Then** it is larger than its visual footprint

### Story 5.8: Edit component properties

As a template author,
I want to set a component's exact position, size and typography,
So that I can be precise where dragging is only approximate.

**Covers:** FR5 · UX-DR10, UX-DR25

**Acceptance Criteria:**

**Given** a selected component
**When** the properties panel renders
**Then** it is contextual to the selection and exposes position (X/Y), size (width/height), font family, font size, bold, italic, text alignment, vertical alignment, border, padding, background and visibility

**Given** a property edit
**When** it is committed
**Then** it is sent as a command to the engine and the snapshot updates
**And** an uncommitted keystroke does not enter the document

**Given** every property field
**When** navigated by keyboard
**Then** it is reachable and operable, with visible focus using the select token

### Story 5.9: A canvas the browser never measures

As a template author,
I want the canvas to wrap text where the engine would wrap it,
So that the approximate view is approximate in scale, not wrong about where my content lands.

**Covers:** NFR2 · AD-17 · UX-DR21

**Acceptance Criteria:**

**Given** any text painted on the canvas
**When** its metrics are obtained
**Then** they come from the engine's measure API

**Given** text rendered on the canvas
**When** the DOM is inspected
**Then** it is painted as pre-broken lines with browser wrapping disabled — `white-space: pre`, no `text-wrap`, no justification

**Given** the browser
**When** its role is assessed
**Then** it contributes rasterization only, and determines no pagination, no line breaking and no measurement

**Given** Thai or CJK text on the canvas
**When** it wraps
**Then** it wraps at the same opportunities the engine would choose

**Given** the single-line text painting introduced with the palette in Story 5.7
**When** this story lands
**Then** it is superseded by engine-measured, pre-broken multi-line painting
**And** no code path remains in which the browser decides where a line ends

### Story 5.10: Preview the exact production document

As a template author,
I want to see the real PDF my developer's service will produce,
So that I am the first person to see the document rather than the last.

**Covers:** FR34, FR35 · NFR5 · AD-16, AD-18 · UX-DR9, UX-DR15, UX-DR23

**Acceptance Criteria:**

**Given** the workspace
**When** the author switches to Preview
**Then** the canvas is replaced by the rendered PDF as a full-surface mode switch, not a panel
**And** selection and scroll position are preserved where meaningful

**Given** a preview render
**When** it is invoked
**Then** it takes three inputs — the serialized `.folio` bytes from the same save path, the sample data document, and an author-supplied parameter document
**And** it never receives a live in-memory document

**Given** the previewed PDF
**When** its hash is compared to a native `folio-go` render of the same three inputs
**Then** they are identical

**Given** the preview surface
**When** it is implemented
**Then** it is a controlled `pdfjs-dist` canvas, never the browser's built-in viewer in an `iframe` or `embed`

**Given** a 50-page document rendering
**When** it is in progress
**Then** progress is indicated, the blocking is confined to the preview surface, and the canvas remains interactive

**Given** the interface as a whole
**When** it is reviewed
**Then** nothing implies server rendering, a cloud round-trip or an account
**And** the canvas-approximate / preview-exact asymmetry is legible without a tutorial

### Story 5.11: Never show a stale preview

As a template author,
I want a preview that has gone out of date to say so,
So that I never commit a template believing I have seen its output when I have not.

**Covers:** AD-18 · UX-DR14

**Acceptance Criteria:**

**Given** a rendered preview
**When** its identity is computed
**Then** it is a hash of the serialized template, the data, the parameters, the engine version and the `FontSet` identity

**Given** a committed command that changes any of those inputs
**When** the identity is recomputed
**Then** it differs and the preview is marked stale immediately

**Given** a stale preview
**When** it is displayed
**Then** it is either visibly invalidated or re-rendered
**And** it is never presented unmarked as the production artefact

### Story 5.12: Diagnostics that locate, and an interface that can be driven from the keyboard

As a template author,
I want a render complaint to point at the thing that caused it, and to be able to work without a mouse where it matters,
So that fixing a problem is one click and the tool does not exclude people who cannot drag.

**Covers:** FR41 · AD-10 · UX-DR17, UX-DR19, UX-DR20, UX-DR22, UX-DR25

**Acceptance Criteria:**

**Given** a render returning warnings — clipped content, an over-tall row, a glyph with no coverage
**When** they are presented
**Then** they surface in the preview where the consequence is visible, non-blocking and dismissible
**And** selecting one locates the offending element back on the canvas by its element id

**Given** a render that fails
**When** the error is presented
**Then** it blocks within the preview surface, names the element and the data path, and offers return to the canvas

**Given** any diagnostic or error
**When** it is displayed
**Then** it is announced, and distinguished by shape before colour

**Given** the application
**When** it is driven by keyboard
**Then** the palette, properties, binding tree and table columns are all reachable and operable
**And** every focusable element shows visible focus using the select token
**And** every icon-only control carries an accessible name

**Given** the frequent operations
**When** shortcuts are used
**Then** save, undo, redo, delete, duplicate, nudge, toggle preview and toggle snapping are covered, with hints in menus and tooltips
**And** no command palette exists

**Given** undo
**When** it is exercised
**Then** it covers every canvas and property mutation as engine-side history over committed commands
**And** loading sample data is not undoable and does not appear to be

---

## Epic 6: A template author can bind a report to data and build the golden report

Ploy loads the sample JSON her developer gave her, walks its paths as a tree, binds each field by picking rather than typing, drops a table into the Content band, and configures its five columns as a matrix. She saves the file and commits it. Anan renders it, and the hash matches what she saw.

### Story 6.1: Load sample data and browse its paths

As a template author,
I want to see the shape of the data my developer will supply,
So that I am binding to something real rather than to a name I invented.

**Covers:** FR9 · UX-DR7, UX-DR13

**Acceptance Criteria:**

**Given** the binding panel before any data is loaded
**When** it is displayed
**Then** it states that binding is unavailable and offers the load action
**And** components can still be placed on the canvas

**Given** a sample JSON document loaded from the local machine
**When** the panel renders
**Then** its paths appear as a navigable tree, including nested objects and collections

**Given** the panel
**When** its placement is inspected
**Then** it is docked to the workspace, not a separate destination

### Story 6.2: Bind a component by picking a path

As a template author,
I want to bind a field by choosing it from the tree,
So that I never mistype a path and discover it at render time.

**Covers:** FR7 · UX-DR7

**Acceptance Criteria:**

**Given** a selected component and a path in the tree
**When** they are connected
**Then** the component becomes bound to that path and the binding is sent to the engine as a command

**Given** a bound component
**When** it is displayed on the canvas
**Then** its bound state is marked using the bind accent, while its selection handles remain the select accent

**Given** every scalar field in the golden report
**When** the template is authored
**Then** each was bound by selecting a path, with no path typed by hand

### Story 6.3: Supply parameters for preview

As a template author,
I want to provide the runtime values my developer's service will supply,
So that I can preview a statement whose generated date is real rather than blank.

**Covers:** FR21 · AD-16

**Acceptance Criteria:**

**Given** a template using `{{params.reportDate}}`
**When** the author opens the parameter document
**Then** they can supply values for the parameters the template references

**Given** a preview render
**When** it runs
**Then** the parameter document is passed as the third input alongside template bytes and sample data

**Given** a parameter referenced by the template but absent from the parameter document
**When** the preview renders
**Then** the failure is located, naming the parameter and the element id

### Story 6.4: Configure table columns as a matrix

As a template author,
I want to configure all my columns in one grid,
So that setting up five columns is one task rather than five repetitions of a form.

**Covers:** FR10 · UX-DR8, UX-DR25

**Acceptance Criteria:**

**Given** a selected Table component
**When** the table editor is invoked
**Then** it opens as a focused surface over the canvas

**Given** the editor with no columns yet
**When** it is displayed
**Then** add-column is the only meaningful action and is unmistakable

**Given** columns and their attributes
**When** they are presented
**Then** they appear as a matrix of columns against attributes — never as a repeated single-column form

**Given** the editor
**When** it is navigated by keyboard
**Then** it behaves as a data grid

**Given** each column
**When** it is configured
**Then** width, alignment and header label can all be set, and columns can be added and removed

### Story 6.5: Bind columns in row scope and configure footer aggregates

As a template author,
I want each column bound to a field of the current row, with a sum on the columns that need one,
So that my transaction table fills itself and totals itself.

**Covers:** FR10 · AD-11, AD-13

**Acceptance Criteria:**

**Given** a table bound to a collection
**When** a column's binding is configured
**Then** the fields offered are those of the row scope, expressed through the region's declared alias

**Given** a column
**When** a footer aggregate is configured
**Then** sum, count or average can be selected per column

**Given** the configured table
**When** it is previewed
**Then** rows generate from the collection and footer aggregates compute over the whole collection

### Story 6.6: Present a failed render honestly

As a template author,
I want a render failure explained in place,
So that I can fix it without leaving the tool or guessing.

**Covers:** UX-DR12, UX-DR24

**Acceptance Criteria:**

**Given** a render that fails
**When** the error card is displayed
**Then** it states the fact, names the location, and offers no comfort — matching the product's terse, technical voice
**And** it uses the danger token, which no mockup previously exercised

**Given** the error card
**When** it is presented
**Then** it names the element and the data path and offers return to the canvas

### Story 6.7: The round trip closes

As a solo builder,
I want a template authored in the browser to render byte-identically through the Go library,
So that the claim the entire product rests on is demonstrated end to end.

**Covers:** S5, S7, S8, S9

**Acceptance Criteria:**

**Given** the golden report authored in this session in the designer — not merely re-saved
**When** it is saved locally and rendered by native `folio-go` with the same data and parameters
**Then** the output is byte-identical to the preview the author saw

**Given** a template hand-edited in a text editor and never opened in the designer
**When** it is loaded and rendered
**Then** it succeeds

**Given** a second, structurally different template
**When** it is authored and rendered
**Then** it renders correctly and reproducibly

**Given** every error case in the diagnostic registry
**When** the suite runs
**Then** each produces a located, actionable message
