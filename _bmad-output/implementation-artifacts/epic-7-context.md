# Epic 7 Context: A template author can lay out a multi-page legal document

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Folio can render statements but cannot author contracts. A contract is body copy — numbered clauses
running for paragraphs, justified, set at a house-style line spacing — followed by a schedule and a
signature block that sit on later pages. Today a text element holds a single unbroken run and the
authoring canvas is exactly one page tall, so neither half is expressible. This epic adds both: the
typography that makes a text element hold a paragraph rather than a line (mandatory breaks, author-
controlled line spacing, justified edges, and a designer editor that can accept them), and an
authoring surface as long as the document's content column, with page breaks drawn where the engine
will actually take them. It deliberately changes **nothing** about the pagination model itself.

## Stories

- Story 7.1: Break a line where the author typed a break
- Story 7.2: Set the space between a paragraph's lines
- Story 7.3: Justify a paragraph's edges
- Story 7.4: Author body text in the designer
- Story 7.5: The content column runs past the first page
- Story 7.6: The canvas draws every page the document will produce
- Story 7.7: Keep a signature block together across a page break

## Requirements & Constraints

- A line feed inside a text value is a **mandatory** break the line packer may never decline; two
  consecutive feeds yield an empty line of one full advance; `\r\n` is one break, not two; the break
  character is drawn on neither line.
- Line spacing is an author-set integer ratio scaling the ruled baseline-to-baseline advance only.
  Absent means today's font-derived leading. Out-of-range or non-integral values are located load
  errors, never silent clamps.
- Justification distributes slack across a line's break opportunities so the line meets its declared
  width; the last line of a paragraph and any line ended by a mandatory break are set at the natural
  start edge, unjustified. Remainder distribution needs one stated, ordered integer rule.
- The align set stays closed — `justify` joins `left`/`center`/`right`; nothing else.
- Overflow behaviour is unchanged: content wider than its declared width is clipped with a located
  warning alongside successful output, never a fatal error.
- The designer must accept and preserve multi-line values, including a clause pasted from a word
  processor (paragraph breaks kept as mandatory breaks, other formatting silently discarded), and
  must expose line spacing and alignment in the inspector beside existing typography controls.
- Placement in the **content** band may extend past one page's content height; the `pageHeader` and
  `pageFooter` bands remain one page tall by definition and keep refusing such placements. Negative
  and beyond-safe-bound coordinates stay refused with unchanged messages.
- The canvas draws one sheet per pagination window with header/footer chrome repeated, break marks
  taken from the engine's projection, and must state plainly that a component's page is a consequence
  of the content above it — a column position, not a pin to a page number.
- Keep-together groups place entirely in one window or entirely in the next; a group taller than one
  window keeps the existing exception (its own page, clipped, recorded as a warning beside the bytes).
- **Every field this epic adds is optional and absent-by-default**, and the existing golden fixture
  corpus must hash identically after it on all four targets. This is an acceptance criterion.
- Every feature ships its golden fixture; a hash change is investigated as a defect until proven an
  intended, versioned behaviour change. Any change to layout or emission is a breaking release for
  downstream test suites.

## Technical Decisions

- **Determinism boundary (AD-1/AD-2/AD-3).** All geometry is `int64` millipoints from the shared
  geometry package; no `float64` in any layout, measurement, or emission signature; no transcendental
  math, clock, environment, or map-order dependence in the render path. Line-spacing arithmetic is
  therefore an integer ratio in thousandths applied to millipoints — a binary float here would break
  byte identity across targets. Numbers reach output through exactly two emitters (thousandths-scaled
  decimal, plain integer) and no other route.
- **Exact-decimal discipline (AD-23).** The same ban on binary floating point that covers report data
  covers any new ratio or slack arithmetic: exact scaled-integer decimals, round-half-to-even where a
  division is unavoidable, defined scale stated rather than incidental.
- **Break opportunities (AD-25).** The engine's contract is break *opportunities*, not segmentation.
  Unknown Thai runs are atomic, clusters are indivisible, and declared unbreakable values admit no
  interior break — but those constraints bind the breaks the engine *infers*. A line feed the caller
  supplied is a break the engine was **told** about, so it survives inside a declared value.
- **Stage ranks.** Strictly forward dependencies with `text` (5) below `layout` (7) below `pdf` (8);
  a stage needing something later receives it as a parameter, never an import. The rank table is
  executable and lives in the lint module, not in prose.
- **Two passes (AD-4) and absolute boxes (AD-24).** Pass one produces the complete page model; pass
  two serializes it and lays nothing out. Element coordinates are relative to their band, top-left
  origin with Y down, flipped to PDF space in exactly one function. Nothing in a band reflows.
  Pagination's window model over one unbounded content column is an **input** to this epic — no
  per-page layouts, no second grouping model in layout; keep-together reuses the existing row-group
  machinery.
- **Canonical file form (AD-9).** One legal serialization: sorted keys, two-space indent, LF, trailing
  newline; round-trip in both directions is a shipped test. New style fields are `lowerCamelCase` and
  must survive an edit-and-edit-back with identical bytes.
- **Engine owns the document (AD-15).** No TypeScript model of a template. Every committed change is
  an opaque command over the one channel; the UI holds an immutable snapshot for painting. Transient
  interaction state stays in the UI. Undo/redo is engine-side history over commands.
- **The browser never measures text (AD-17).** The canvas gets every metric, line break, and now every
  spaced advance and justified line position from the engine's measure/projection API, painting
  pre-broken lines with browser wrapping and browser justification disabled. The canvas and the PDF
  must consume the same advance.
- **Diagnostics (AD-14).** One diagnostic type, one channel, stable codes from a closed additive
  registry, element id attached. Errors abort; clipping and over-tall content are warnings returned
  beside successful output.

## UX & Interaction Patterns

- Single dark-chrome / light-canvas theme; the page is the only bright surface, with contrast met on
  dark chrome's own terms. Design tokens only — never hard-coded colour; cyan means structure, amber
  means data.
- The workspace is persistent regions (canvas, palette, properties, document bar) with the page at
  true proportions, visible boundaries, three bands, and a snapping grid — the multi-page canvas
  extends this rather than replacing it.
- New controls follow the established component and property-field patterns; refusals state the
  concrete reason in text at the control that caused it.
- The accessibility floor binds: every control keyboard-reachable, visible focus, accessible names on
  icon-only controls, diagnostics distinguished by shape before colour.
- A single-page template must look and behave exactly as it does today.

## Cross-Story Dependencies

- Two separable pillars: body text (7.1–7.4) and the multi-page canvas (7.5–7.6). 7.7 supplies the one
  legal-specific layout primitive neither pillar covers and is the story to cut first if the epic is
  trimmed.
- 7.4 depends on 7.1–7.3: the designer editor exposes multi-line values, line spacing and justify only
  once the engine accepts them.
- 7.6 depends on 7.5: the canvas can draw sheets only once the engine accepts and projects content-band
  placement beyond one page, including the window height and window count.
- 7.2 and 7.3 must hold the canvas/PDF agreement established earlier in the designer work — the canvas
  text paint plan consumes the same advances the renderer does.
- 7.7 reuses the existing table-row grouping machinery and the existing over-tall-group exception; it
  adds no new pagination concept.
