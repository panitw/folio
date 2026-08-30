# Epic 7 Context: A template author can lay out a multi-page legal document

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A contract is mostly body copy — numbered clauses running for paragraphs, justified, set at a
firm's house line spacing — followed by a schedule and a signature block that sit on later pages.
Today a text element holds one unbroken run and the authoring canvas is one page tall, so neither
half can be authored. This epic adds both: **body text** (a text element holds a paragraph, with
mandatory breaks, author-controlled line spacing and justified edges) and **a multi-page canvas**
(the authoring surface runs as long as the content column, drawing every page the engine will
produce with header/footer chrome repeated and breaks marked where the engine actually takes
them). It also adds the one legal-specific layout primitive neither pillar supplies: a
keep-together group so a signature block is never severed by a page boundary. The pillars are
separable and can ship independently.

## Stories

- Story 7.1: Break a line where the author typed a break
- Story 7.2: Set the space between a paragraph's lines
- Story 7.3: Justify a paragraph's edges
- Story 7.4: Author body text in the designer
- Story 7.5: The content column runs past the first page
- Story 7.6: The canvas draws every page the document will produce
- Story 7.7: Keep a signature block together across a page break
- Story 7.8: Refuse a justified table at load, in the author's own terms
- Story 7.9: The canvas tells the truth about keep-together groups
- Story 7.10: An over-tall element is refused whether or not it is grouped

## Requirements & Constraints

- A mandatory line break authored inside a text value survives into the PDF, so one component can
  hold more than one paragraph. `\r\n` is one break; two consecutive breaks yield an empty line.
- Justified alignment sets a text element's edges to its declared width; the last line of a
  paragraph — and any line ended by a mandatory break — is not justified. Justification is a
  **text-element** capability only; a table must refuse it at load.
- Line spacing is author-controlled as an integer ratio applied to the ruled inter-baseline
  advance. Absent means today's font-derived leading.
- The designer must let an author write and edit multi-paragraph body text, its line spacing and
  its alignment.
- The canvas must extend across every page the content column currently occupies, repeating
  page-header and page-footer chrome per page and drawing breaks where the engine will take them.
- A group of content-band elements can be **declared** to paginate together. Designer-side
  authoring of groups is deliberately out of scope for this epic — file-only declaration
  satisfies the requirement.
- **Every field this epic adds is optional and absent-by-default: the existing golden corpus must
  hash identically afterwards.** This is an acceptance criterion, not an aspiration, and it
  applies per story.
- Overflow behaviour is unchanged: content exceeding its declared bounds is clipped at the
  component boundary with a located diagnostic, never reflowed and never silently dropped.

## Technical Decisions

- **Pagination is an input to this epic, not a target.** The layout package already treats the
  content band as a window onto one unbounded column and already renders fifty pages from a
  one-page template. This epic lets the *designer see* that column; it must not introduce
  per-page layouts, which would mean content that never reflows and silently clipped clause text.
- **No binary floats anywhere.** All geometry is integer millipoints; line spacing is an integer
  ratio in thousandths. Float arithmetic would break byte-identity across the four CI targets
  (darwin/arm64, linux/amd64, linux/arm64, js/wasm), which are compared by hash. Where slack or
  spacing does not divide evenly, the remainder must be placed by one stated, ordered rule in
  integer arithmetic.
- **The browser never measures text.** The canvas paints pre-broken lines and takes every metric,
  line break and page-boundary position from the engine's projection — including justification
  and window origins. Nothing is computed in the browser.
- **The engine owns the document.** There is no TypeScript model of the file. Every committed
  edit travels as an opaque command to the engine and the canvas re-projects from the engine's
  answer. Transient interaction state stays in the UI.
- **Boxes are absolute and nothing negotiates.** An element's coordinates are relative to its
  band's top-left, Y increasing downward. `pageHeader` and `pageFooter` bands are one page tall by
  definition and repeat — a beyond-one-page coordinate is legal only in the content band.
- **Errors and diagnostics are one type on one channel.** Severity `Error` aborts a render;
  `Warning` accompanies a successful one. A load refusal must be *located* — naming element and
  field — and needs a code from the closed registry to survive the designer's error path;
  uncoded load errors are flattened into a generic message and never reach the author in words.
  Whether to mint a further per-field style diagnostic code is an open lead decision that a
  story must not settle unattended.
- **Over-tall discriminator:** an individually over-tall element is a located fatal overflow
  error, tagged or not; only a group that exceeds a window *in aggregate* takes clip-and-warn. A
  group tag does not launder authorship, and a group of one must be indistinguishable from no
  group.
- **Reuse the existing grouping machinery** (the one that serves table rows) rather than adding a
  second grouping model to layout.
- **Grouping is a pure template property** — it needs no data, params or fonts — so the canvas has
  every input it needs to be exactly right about it. A canvas discrepancy caused by grouping is a
  defect to fix, never a new "floor" cause on the window-count estimate.
- Every feature ships its golden fixture; a hash change is investigated as a defect until proven
  to be an intended, versioned behaviour change.

## UX & Interaction Patterns

- The canvas is an **approximation** and the preview is **exact**; the interface must keep that
  asymmetry legible and never let the canvas read as authoritative.
- A component on a later page must be presented as a **column position**, not a pin to a page
  number — the interface should state plainly that its page is a consequence of the content above
  it and can change when the data does.
- Diagnostics that affect output are non-blocking, dismissible, and locate the offending element
  back on the canvas; fatal errors block within the preview surface, name element and path, and
  offer a route back.
- Line spacing and alignment (including justify) belong in the inspector alongside the existing
  typography controls. Pasted word-processor text keeps its paragraph breaks and discards other
  formatting without error.
- Single-page templates must look and behave exactly as they do today.

## Cross-Story Dependencies

- 7.1 → 7.3: justification must treat a mandatory break as a paragraph end, so breaks land first.
- 7.1–7.3 → 7.4: the designer can only offer what the engine and format accept.
- 7.4 → 7.8: 7.4 discharged the product half (the inspector stops offering justify for tables);
  7.8 owns the format half (the loader refuses it).
- 7.5 → 7.6: the engine must accept beyond-page-one content coordinates and report the page-height
  window and window count before the canvas can draw a run of sheets.
- 7.7 → 7.9: 7.7 taught the render path to keep groups whole; 7.9 teaches the canvas projection
  the same thing so its window count and origins match the render exactly. **7.9 gates epic
  completion.**
- 7.7 → 7.10: 7.10 resolves a contradiction inside 7.7's own contract matrix and amends 7.7's
  over-tall clause to read "taller than one window in aggregate". 7.10 does not gate epic
  completion but **does gate the v0.1.0 tag**, because it narrows what is accepted — free only
  while nothing is released.
- 7.7 is the story to cut first if the epic needs trimming (7.9 and 7.10 fall with it).
- 7.8 belongs to the body-text pillar and can ship independently of everything after 7.4.
