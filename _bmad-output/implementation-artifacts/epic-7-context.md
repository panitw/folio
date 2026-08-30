# Epic 7 Context: A template author can lay out a multi-page legal document

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Folio can author a statement but not a contract. A contract is mostly body copy — numbered clauses
running for paragraphs, justified, set at a house-style line spacing — followed by a schedule and a
signature block on later pages. Today a text element holds one run of text with no way to express a
break, and the canvas is one page tall, so neither half can be authored. This epic supplies both: it
makes a text element hold a *paragraph* (mandatory breaks, author-controlled line spacing, justified
edges, and a designer that can type and paste them), and it makes the authoring surface as long as
the document the engine will actually produce. It changes **nothing** about pagination — the engine
already treats the content band as a window onto one unbounded column, and that window model is an
input to this epic, not a target of it.

## Stories

- Story 7.1: Break a line where the author typed a break
- Story 7.2: Set the space between a paragraph's lines
- Story 7.3: Justify a paragraph's edges
- Story 7.4: Author body text in the designer
- Story 7.5: The content column runs past the first page
- Story 7.6: The canvas draws every page the document will produce
- Story 7.7: Keep a signature block together across a page break
- Story 7.8: Refuse a justified table at load, in the author's own terms

## Requirements & Constraints

- A mandatory line break authored inside a text value survives to the PDF, so one component can hold
  more than one paragraph. Two consecutive breaks produce an empty line, giving a paragraph gap.
- A text element's line spacing is author-controlled, expressed as an integer ratio applied to the
  ruled inter-baseline advance. Absent means today's font-derived leading.
- A text element can be justified to its declared width; the last line of a paragraph is left
  unjustified.
- Multi-paragraph body text, its spacing and its alignment are all authorable in the designer, not
  only by hand-editing the file.
- The canvas extends across every page the content column currently occupies, with page-header and
  page-footer chrome repeated per page and breaks drawn where the engine will take them.
- A group of content-band elements can be declared to paginate together, so a signature block is
  never severed by a page boundary.
- **Every field this epic adds is optional and absent-by-default. The existing golden corpus must
  hash identically afterwards on all four targets. That is an acceptance criterion, not an
  aspiration** — and every feature here ships its own golden fixture.
- Existing overflow behaviour is unchanged: content that exceeds its box is clipped with a located
  warning beside the PDF bytes, never a fatal error. The over-tall-unit exception (a page of its own,
  clipped, recorded as a warning) applies unchanged to keep-together groups.

## Technical Decisions

- **Integer arithmetic only.** All geometry is int64 millipoints; binary floats appear nowhere in the
  engine core, for geometry or for data. Line spacing is therefore an integer ratio in thousandths,
  and justification slack is distributed in integer millipoints by one stated, ordered remainder rule
  — anything else breaks byte-identity across the four targets.
- **Layout happens once.** Pass one produces the complete page model; pass two only serializes it.
  The PDF writer performs no measurement, no line breaking, and no pagination.
- **Boxes are absolute and nothing negotiates.** Coordinates are relative to the band's top-left, Y
  down; elements never reflow or displace siblings. Line spacing must scale the baseline-to-baseline
  advance and nothing else, so the first baseline — and the element's top edge — does not move.
- **Mandatory breaks are not optional opportunities.** The text layer must report them as mandatory
  and the line packer must always take them; a `\r\n` pair is one break; the break character itself is
  consumed and drawn on neither line, as whitespace breaks already are.
- **The engine owns the document.** There is no TypeScript model of a template. The designer holds an
  immutable snapshot for painting and sends every committed mutation as an opaque command over one
  channel; undo/redo is engine-side history. Transient interaction state stays in the UI.
- **The browser never measures text.** The canvas gets every metric and line break from the engine's
  measure API and paints pre-broken lines with browser wrapping and justification disabled. Anything
  the canvas draws about page windows and break positions must come from the engine's projection, not
  from a browser computation.
- **Reuse the existing grouping machinery** for keep-together rather than introducing a second
  grouping model into layout.
- **Diagnostics are one type on one channel**, carrying a stable code from a closed registry, an
  element id and a path. Out-of-range or malformed values are located load errors naming the element
  and the field — never a silent clamp or no-op. An *uncoded* load error is flattened to a generic
  message before it reaches a designer author, so any refusal that must be explained in the author's
  terms needs a code; whether to mint another per-field style code is an open lead-level decision.
- **Validate a closed set against the set its consumer accepts**, keyed on element type — not on where
  the value sits in the JSON. That mis-partition is the root cause Story 7.8 exists to fix.
- The file format has one canonical byte form; adding an optional field must not raise the format
  version for documents that do not use it.

## UX & Interaction Patterns

- The three bands are structural, not decorative — which band a component sits in determines whether
  it repeats. Band identity and drop targeting stay unambiguous as the canvas grows past one page.
- A component on a later page is at a **column position**, not pinned to page three. The interface
  must say plainly that its page is a consequence of the content above it and can change with the
  data.
- The canvas is an approximation, the preview is exact, and the interface must make that asymmetry
  legible without a tutorial.
- Property fields commit on blur or Enter; a multi-selection marks divergent values as mixed rather
  than picking one. The inspector must not offer a setting the renderer cannot honour for the
  selected element type.
- Pasting from a word processor preserves paragraph breaks and discards other formatting silently.
- Undo covers every canvas and property mutation, and a single-page template behaves as it does today.

## Cross-Story Dependencies

- Two separable pillars: **body text** (7.1–7.4, plus 7.8) and **the multi-page canvas** (7.5–7.6).
  7.7 is the one legal-specific layout primitive neither pillar supplies and is the story to cut first
  if the epic needs trimming.
- 7.1 (mandatory breaks) underpins 7.3, whose last-line rule must also treat a line ended by a
  mandatory break as unjustified, and 7.4, whose paste path produces them.
- 7.2 and 7.3 must both be reflected by the canvas text paint plan, which must consume the same
  advance and the same line breaks the renderer does.
- 7.4 needs 7.1–7.3 landed before the inspector can offer those capabilities.
- 7.6 depends on 7.5: the canvas can only draw the page windows once the engine accepts content-band
  placements beyond one page and reports the window height and count in its projection.
- 7.8 was added mid-epic when 7.4's plan gate routed the justified-table problem out of it: 7.4 stops
  the designer *offering* a justified table, 7.8 makes the format *refuse* one. It can ship
  independently of everything after 7.4, and must invert (not delete) the shipped tests that pin
  today's accepting behaviour while keeping a text element's justify accepted.
- Wider line extents from 7.2 and taller content from 7.5 feed the existing pagination unchanged;
  do not modify the layout package's break rules or window model to accommodate them.
