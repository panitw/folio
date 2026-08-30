# Epic 7 Context: A template author can lay out a multi-page legal document

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Folio today can author a statement: one page tall on the canvas, one run of text per component.
A contract is a different document — numbered clauses that run for paragraphs, justified to the
house style's line spacing, followed by a schedule and a signature block that sit on later pages.
Epic 7 makes both halves authorable. It has two separable pillars: **body text** (7.1–7.4) makes a
text element hold a paragraph rather than a line, and **the multi-page canvas** (7.5–7.6) makes the
authoring surface as long as the document actually is. 7.7 adds the one legal-specific layout
primitive neither pillar supplies, and is the story to cut first if the epic must be trimmed. The
epic changes the authoring surface and the typography, and deliberately changes **nothing** about
the engine's pagination model.

## Stories

- Story 7.1: Break a line where the author typed a break
- Story 7.2: Set the space between a paragraph's lines
- Story 7.3: Justify a paragraph's edges
- Story 7.4: Author body text in the designer
- Story 7.5: The content column runs past the first page
- Story 7.6: The canvas draws every page the document will produce
- Story 7.7: Keep a signature block together across a page break

## Requirements & Constraints

- **Mandatory line breaks** authored inside a text value survive into the PDF, so one component can
  hold more than one paragraph. A break is mandatory, not an opportunity the line packer may decline.
- **Justified alignment** joins the closed align set: interior lines fill the declared width; the
  last line of a paragraph, and any line ended by a mandatory break, are set at the natural start edge.
- **Author-controlled line spacing**, expressed as an integer ratio applied to the ruled
  baseline-to-baseline advance. Absent means today's font-derived leading, unchanged.
- **Body-text authoring in the designer**: multi-line entry and paste, plus line spacing and
  alignment (including justify) as inspector controls.
- **A multi-page authoring canvas** that extends across every page the content column currently
  occupies, with page-header and page-footer chrome repeated per page and breaks drawn where the
  engine will take them.
- **Keep-together groups** of content-band elements that paginate as a unit, so a signature block is
  never split across a page boundary.
- **Every field this epic adds is optional and absent-by-default.** The existing golden corpus must
  hash identically after the epic lands, on all four CI targets. This is an acceptance criterion, not
  an aspiration.
- Byte-reproducibility across `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm` is the
  product's signature requirement; every feature ships its golden fixture, and a hash change is a
  defect until proven an intended, versioned change.
- Overflow behaviour is unchanged: content exceeding its declared bounds is clipped and reported as a
  Warning alongside the PDF bytes — never silently reflowed, never fatal.

## Technical Decisions

- **The pagination model is an input, not a target.** The content band is already a window onto one
  unbounded column, and the engine already renders fifty pages from a template nobody designed fifty
  pages for. Epic 7 lets the *designer* see that column. Do not introduce per-page layouts — that would
  mean content that never reflows and, for a contract, silently clipped clause text.
- **Boxes are absolute; nothing negotiates.** An element's x/y is relative to its band's top-left,
  never to the page; bands are placed on the page by the layout package alone. The page model is
  top-left origin, Y down, and the flip to PDF user space happens in exactly one function.
- **No binary floats anywhere under `internal/`**, for geometry or for data. Line spacing arithmetic
  must be an integer ratio in thousandths applied in millipoints. Any slack distribution for
  justification needs one stated, ordered remainder rule in integer millipoints so all four targets
  produce identical bytes.
- **Only the content band grows.** The page-header and page-footer bands are one page tall by
  definition and repeat; coordinates beyond one page stay refused there. Negative coordinates and
  coordinates beyond the JavaScript-safe geometry bound stay refused everywhere, with unchanged messages.
- **Reuse the existing item-group machinery** for keep-together rather than adding a second grouping
  model to the layout package. A group taller than one window follows the existing over-tall exception:
  a page of its own, clipped, recorded as a Warning beside the PDF bytes.
- **Errors and diagnostics are one type on one channel**, carrying a severity, a stable code from a
  closed registry, and the offending element's stable id. A bad line-spacing value is a located load
  error naming the element, never a silent clamp.
- **The engine owns the document in the designer.** There is no TypeScript model of the file: the wasm
  engine parses, mutates, validates and serializes it; the UI holds an immutable snapshot for painting
  and sends every committed mutation as an opaque command over one channel. Transient interaction state
  (a drag in flight, an uncommitted keystroke) never enters the document.
- **The browser never measures text, including on the canvas.** Every metric, line break, page-break
  position and window count comes from the engine's projection. The canvas paints pre-broken lines with
  browser wrapping and justification disabled; it must never compute a break itself.
- **The file has one canonical byte form.** Any new style field must round-trip through the canonical
  parser/serializer with sorted keys and stable output.

## UX & Interaction Patterns

- The workspace is a single dark-chrome / light-canvas theme — the white page is the only bright
  surface, and contrast must be met on dark chrome's own terms with no light-mode fallback.
- The canvas shows the page at true proportions with visible boundaries, three bands, and a snapping
  grid; a multi-page document draws one sheet per window in order, with header and footer chrome
  repeated because the engine repeats it.
- A component's page is a *consequence* of the content above it, not a pin to a page number. The
  interface must say so plainly, because the page can change when the data does.
- New properties join the existing property-field and canvas-component patterns rather than inventing
  a surface; the declared per-surface states (loading, empty, populated, rendering, diagnostic, error,
  unsaved, stale preview) apply to the multi-line editor too.
- A clause pasted from a word processor keeps its paragraph breaks and quietly loses everything else —
  no error, no partial formatting.
- A single-page template must look and behave exactly as it does today.

## Cross-Story Dependencies

- 7.1 (mandatory breaks) precedes 7.3 (justification), which must know not to justify a line ended by
  a break, and precedes 7.4, which is the authoring surface for them.
- 7.2 (line spacing) must feed pagination and the canvas text paint plan alike; the canvas must consume
  the same advance the renderer does, so the canvas/PDF agreement invariant established in Epic 5 holds.
- 7.4 depends on the designer's inspector and command channel from Epics 5 and 6.
- 7.6 depends on 7.5: the canvas can only draw multiple sheets once the engine accepts and projects a
  content column longer than one page, including the window height and window count.
- 7.7 depends on the table-row grouping machinery and the over-tall-row exception delivered in Epic 4.
- Epic 7 consumes the pagination window model from Epic 2 and the clip-and-warn overflow contract from
  the public-API epic without modifying either.
