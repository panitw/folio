# Epic 7 Context: A template author can lay out a multi-page legal document

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A contract is mostly body copy — numbered clauses running for paragraphs, justified, set at a
firm's house line spacing — followed by a schedule and a signature block on later pages. Today a
text element holds one unbroken run and the authoring canvas is one page tall, so neither half can
be authored. This epic adds both pillars, and they are separable. **Body text** makes a text
element hold a paragraph rather than a line: mandatory breaks the author typed, author-set line
spacing, justified edges, all editable in the designer. **The multi-page canvas** makes the
authoring surface as long as the content column, drawing every page the engine will produce with
header and footer chrome repeated and breaks marked where the engine actually takes them. Above
both sits the one legal-specific layout primitive neither supplies: a keep-together group, so a
signature block is never severed by a page boundary.

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

- A line break the author typed survives into the PDF, so one component holds more than one
  paragraph. `\r\n` is one break; two consecutive breaks give an empty line of one full advance. A
  typed break is *mandatory* — the packer may never decline it — and, unlike an inferred
  opportunity, it survives inside a declared unbreakable value.
- Justified alignment flushes a text element's edges to its declared width. Three conditions leave
  a line ragged at its own start edge: it is the paragraph's last line, it was ended by a mandatory
  break, or it has no interior break opportunity. Justify is a **text-element** capability only — a
  table must refuse it at load, naming element and field.
- Line spacing scales the baseline-to-baseline advance and nothing else; the first baseline does
  not move, so no sibling appears to shift. Absent means today's ruled value; an out-of-range or
  non-integral value is a located load error, never a silent clamp.
- The designer must let an author write, paste and edit multi-paragraph body text and set its line
  spacing and alignment; pasted word-processor text keeps paragraph breaks and drops the rest
  without error.
- Content-band coordinates beyond one page become legal; `pageHeader` and `pageFooter` stay one
  page tall and keep refusing them, as do negative and out-of-safe-range coordinates.
- A group of content-band elements can be **declared** to paginate as one indivisible unit: the
  whole set stays in the window it started in or moves to the next, each member at its own declared
  position, no sibling moved, no gap invented. Members need not be adjacent; a tag is
  document-scoped, invalid on a table and outside the content band. Designer-side *authoring* of
  groups is deliberately out of scope — file-only declaration satisfies the requirement.
- **Every field this epic adds is optional and absent-by-default: the existing golden corpus must
  hash identically afterwards.** A per-story acceptance criterion, not an aspiration. Every new
  feature ships its own golden fixture, and a hash change is a defect until proven an intended,
  versioned change.
- Overflow behaviour is unchanged: content past its declared bounds is clipped with a located
  diagnostic, never reflowed, never silently dropped.

## Technical Decisions

- **Pagination is an input to this epic, not a target.** Layout already treats the content band as
  a window onto one unbounded column and already renders fifty pages from a one-page template. This
  epic lets the designer *see* that column; it must not introduce per-page layouts, which would mean
  content that never reflows and silently clipped clause text.
- **No binary floats.** Geometry is integer millipoints; line spacing is an integer ratio in
  thousandths. Floats would break the byte-identity the four CI targets are compared on. Where
  slack does not divide evenly, one stated ordered rule places the remainder: every gap takes the
  quotient, the first *remainder* gaps in reading order take one more, so the sum is exact.
- **The browser never measures text.** The canvas paints pre-broken lines and takes every metric,
  break, justification result and window origin from the engine's projection — origins are
  projected from the pagination result, never recomputed in the browser.
- **The engine owns the document.** No TypeScript model of the file: every committed edit travels
  as an opaque command and the canvas re-projects from the engine's answer.
- **Errors and diagnostics are one type on one channel.** `Error` aborts a render; `Warning`
  accompanies a successful one. A load refusal must be *located* and needs a code from the closed
  registry to reach the author — uncoded load errors are flattened to a generic message. Whether to
  mint a further per-field style code is an open lead decision a story must not settle unattended.
- **Over-tall discriminator:** an individually over-tall element is a located fatal overflow error,
  tagged or not; only a group over-tall *in aggregate* takes clip-and-warn. A tag does not launder
  authorship, and a group of one must be indistinguishable from no group.
- **Reuse the existing grouping machinery** (the one serving table rows) rather than adding a
  second grouping model to layout.
- **Grouping is a pure template property** — no data, params or fonts — so the canvas has every
  input it needs to be exactly right. A canvas discrepancy caused by grouping is a defect to fix,
  never a new "floor" cause on the window-count estimate.
- **Partition a closed set by the code that consumes the value, not by where it sits in the
  document.** Splitting the align set by JSON key path let a justified table in through the other
  door.
- **Version bumps are real costs.** Justify raises a document to a MAJOR version earlier readers
  cannot open; a keep-together group raises a MINOR one. A setting that raises the version and then
  renders as if absent charges the author for nothing, and narrowing what the loader accepts is
  free exactly once — before the first release tag.

## UX & Interaction Patterns

- The canvas is an **approximation**, the preview **exact**; keep that asymmetry legible and never
  let the canvas read as authoritative.
- A component on a later page is a **column position**, not a pin to a page number — say plainly
  that its page follows from the content above it and changes when the data does.
- Voice is terse and technical: state the fact, name the element and path, never apologise, and
  keep errors (render stopped) visually distinct from diagnostics (render proceeded with a
  caveat).
- Line spacing and alignment, justify included, belong in the inspector beside the existing
  typography controls — and the inspector must not offer justify for a table or a mixed selection.
- Single-page templates must look and behave exactly as they do today.

## Cross-Story Dependencies

- 7.1 → 7.3: justification treats a mandatory break as a paragraph end, so breaks land first.
- 7.1–7.3 → 7.4: the designer can only offer what the engine and format accept.
- 7.4 → 7.8: 7.4 discharged the product half (the inspector stops offering justify for tables);
  7.8 owns the format half, inverting or widening the shipped tests that pinned today's acceptance
  rather than deleting them.
- 7.5 → 7.6: the engine must accept beyond-page-one coordinates and report the window height and
  count before the canvas can draw a run of sheets.
- 7.7 → 7.9: 7.7 taught the render path to keep groups whole; 7.9 teaches the canvas projection the
  same, so its window count and origins match a real render. **7.9 gates epic completion.**
- 7.7 → 7.10: 7.10 resolves a contradiction inside 7.7's contract matrix and amends 7.7's over-tall
  clause to "taller than one window in aggregate". It does not gate the epic but **does gate the
  v0.1.0 tag**, because it narrows what is accepted.
- 7.7 is the story to cut first if the epic needs trimming (7.9 and 7.10 fall with it).
- 7.8 belongs to the body-text pillar and can ship independently of everything after 7.4.
