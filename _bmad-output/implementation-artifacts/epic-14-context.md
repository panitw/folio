# Epic 14 Context: The designer's controls read as one product

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

The inspector, document bar, data panel and table editor were each built story by story, and the
seams show: five competing button treatments mixed inside single rows, a Line whose thickness is
called "H" and whose colour is called "Background" (because a line is implemented as a filled bar),
border and edge controls offered on a shape that has no edges, a binding section offered on
components the engine will always refuse, a placed component that is not selected, a data tree
showing concatenated type strings where the design shows the author's own values, a table painted on
the canvas as a box containing the word "Table", and a table editor that is an eleven-column form
where the design specifies a six-column matrix. The product also never wore the mark its design gives
it. This epic makes those surfaces read as one designed product. It changes nothing about the
document model, the command surface, the file format, or a single rendered byte — only what the
panels offer and how they are spelled. No new functional requirements land; the designer-side
requirements for canvas layout, the five-component palette, component properties, scalar binding from
discovered JSON paths, sample data, and table-structure editing are completed here.

## Stories

- Story 14.1: One button vocabulary
- Story 14.2: A Line is a thickness and a colour; a Rectangle is a fill and a border
- Story 14.3: A placed component is the selected component
- Story 14.4: The panel offers no control the engine will refuse
- Story 14.5: The product wears its own mark
- Story 14.6: The DATA tab is the binding panel the design drew
- Story 14.7: The table editor is the matrix the design drew
- Story 14.7b: The table editor's Cancel discards what it counted
- Story 14.8: The table editor carries the header, cell and border sections
- Story 14.9: The canvas draws the table it will print
- Story 14.10: A table column is bound from the main window

**Status at compile time:** 14.1 through 14.9 (including 14.7b, split out of 14.7 on 2026-09-10) are
delivered. **14.10 is the only story left**, followed by the Epic 14 boundary gate, which is not yet
booked. The rules 14.1 "wrote down first" and the rulings settled during 14.2–14.9 are inherited
constraints now, not open questions.

## Requirements & Constraints

- **No format change, no version increment, no new engine surface.** Every story is presentation.
  Where a panel needs data it lacks, the value already exists in an engine projection or the
  sample-data model. Two named exceptions were admitted deliberately: 14.9 added a **canvas
  projection** field carrying each column's label, width, alignment and binding (a projection field,
  never a format field); and 14.8's BORDERS section authors `headerStyle.border` as three flat
  members of the closed command field set — capability rather than presentation, ruled in by the
  owner against a recommendation to split it out, because the engine's own source deferred that
  field to this story. Both are closed. 14.10 needs neither: the column id is already on the wire.
- **Never invent a control the format cannot carry.** Any control not traceable to a field the engine
  actually consumes is absent rather than disabled-and-mysterious. A mockup that draws a capability
  the product does not have is a drawing to correct, not a specification to implement — this is named
  as how the epic's own defects were made.
- **Never normalise a document on selection.** A panel that stops offering a field must still present
  the engine's committed values untouched and preserve them on save, including for documents
  hand-edited into shapes the panel's vocabulary does not describe.
- **The panel must not offer what the engine will refuse, and must say why before the attempt.**
  Scalar binding is text-only; runtime `params` are never a bindable root; a collection is not
  bindable by a text element; a path outside a table's row scope is not bindable to its column. The
  reason is stated where the pick was made, before the engine rejects it.
- **The padding control is dead everywhere**, on a table as on any other element kind. Nothing in the
  project would catch a violation, because the command it sends is legal — so it must not be built.
- **Accessibility is a hard floor:** every interactive element keyboard-reachable and operable, visible
  focus using the select token, an accessible name on every icon-only control, canvas hit targets
  larger than their visual footprint, and the table editor behaving as a data grid under keyboard
  navigation. A respelling must never narrow or lose an accessible name, and an unpickable row must
  report that it is unpickable rather than silently failing to respond.
- **An absence claim needs a test that ADDS the forbidden thing.** Reverting the implementation cannot
  falsify "is not offered" or "is preserved"; it makes such a test pass more easily. Add the mutation
  at every position the forbidden thing could occupy, and read preserved values back rather than
  counting dispatches. This is the mechanism behind this epic's false greens.

## Technical Decisions

- **The engine owns the document (AD-15).** There is no TypeScript model of a `.folio`. The UI holds
  an immutable snapshot for painting and sends every committed mutation as a command; undo/redo is
  engine-side history. Transient interaction state — a drag in flight, an uncommitted keystroke —
  lives in the UI and never enters the document. Note that AD-15 forbids a **second schema**, not a
  pending-edit buffer: the epic's earlier claim that a modal Cancel implies a forbidden buffer was
  measured false, and the table editor's model is settled as **commit-on-blur with Cancel issuing a
  counted compensating sequence of undos** (Cancel / Done, not Cancel / Apply). The epic's opening
  prose still carries the withdrawn premise; the story criteria carry the ruling.
- **The browser never measures text, including on the canvas (AD-17).** Metrics and line breaks come
  from the engine's measure API; the canvas paints pre-broken lines in DOM/SVG. A richer canvas table
  paints from the engine's projection, never from a browser-side model.
- **Table geometry is derived, not stored twice (AD-13).** Column widths are absolute and
  authoritative; a table's width *is* their sum and is never stored separately. The editor's width
  budget reads that rule out.
- **Millipoints are the stored unit and stay millipoints.** The **displayed** unit is **points,
  product-wide**; the mockups' millimetres exist nowhere in the product.
- **The control vocabulary is written down and guarded** (from 14.1): a control is a word; a glyph
  only as one member of a segmented control, on the 16px grid; every control in one class or named
  group is spelled the same way; a glyph always carries an accessible name. A contract test over the
  swept population enforces uniformity and accessible names.
- **Design tokens are the single source of styling** — no hard-coded hex anywhere. The brand mark is
  one size-parameterised component drawn in the existing select token, used at exactly two sizes (the
  document bar and the load screen) and nowhere else. Source-text contract tests pin the token file
  and an exact allowlist of CSS declarations, so any restyle reds them until updated deliberately.
- **Every mutation is a single undo step**, including a column binding made from the main window,
  which commits through the existing column-binding command unchanged. A multi-key panel edit goes as
  one command carrying several changes, never as a sequence.

## UX & Interaction Patterns

- **The mockups and this epic's own prose are a reference, not a transcript.** Both have been measured
  to overstate: the main mockup draws a **Text** inspector only, with no Line inspector at all; the
  table-editor mockup draws controls the format cannot carry and is itself a deliverable to correct.
  Verify the control exists in the mockup, and that the defect still exists in the product, before
  building from either.
- **Two-accent grammar, without exception.** Cyan means structure, focus and authority; amber means
  data and only data. Selection handles stay cyan even on a bound element; canvas binding
  placeholders are amber.
- **Voice is terse and technical** — state the fact, name the location, offer no comfort. Anything
  disabled or unpickable carries a stated reason, never a bare grey-out.
- **Density is a feature.** The table editor is the densest surface in the product and must be a
  matrix (columns × attributes), never a repeated single-column form; reorder and remove are row
  affordances, not extra columns.
- **One control per concept product-wide.** Alignment is one segmented control everywhere; a section
  must not mix icons and words at the same size in the same row; same-family actions are spelled
  alike.
- **The binding panel is docked, not a destination.** Its tree shows values beside paths, `{ }` and
  `[]` markers, badges for collections and runtime params, and dimmed unpickable rows, so the author
  recognises their data rather than decoding a type name. A context bar states what is selected and
  what a pick would bind, before the pick.
- **Canvas is approximate, preview is exact**, and that asymmetry must stay legible without a
  tutorial — the canvas table shows structure (one representative row and its column placeholders),
  not the sample's full data.
- **All declared states must be handled**, including empty-table-with-no-columns, an unbound column,
  and no-sample-data; an empty frame that looks like a rendering failure is a defect.
- Square corners throughout, 1px borders carrying separation, mono for anything a machine reads and
  sans for anything a person wrote.

## Cross-Story Dependencies

- **14.9 → 14.10 (the live one).** Binding a column from the main window requires the canvas to draw
  the columns first; 14.10's selection target is the column 14.9 paints, on the projection field 14.9
  added. 14.10 also carries a prerequisite its criteria do not name: today's selection holds
  **element** ids, so holding a column is a selection-model change, and the canvas table drawing is
  presentational, so 14.10 owns hit-testing as well.
- **14.6 → 14.10.** 14.6 dims row-scope fields as unpickable when a text element is selected; 14.10
  makes exactly those fields pickable when a column is selected, in the same tree and context bar.
  The collection itself, and any path outside the row scope, stay unpickable for a column.
- **14.4 ↔ 14.7 ↔ 14.10 — three different bindings, not to be conflated.** 14.10 is the **per-column
  bound field**; 14.4's "state a table's binding where it is editable" is the **table's own
  collection**, a different value and a different command; and 14.7 deliberately **keeps** the
  collection and row alias editable inside the table editor, restyled into a labelled group rather
  than moved. 14.10 does not relocate the collection.
- **Settled, not open, for 14.10:** the table editor's matrix keeps **showing** each column's bound
  field as context, and that column is **display-only** — binding happens in the main window. The
  owner ruled this on 2026-09-07 (D-14.10.1); it is not a question for the story to reopen.
- **14.8 depended on Epic 12's table styling capability** (header height, header text style,
  alternating row colour) and adds no second way to store it.
- Remaining epic-level work after 14.10 is the **boundary gate**, which is unbooked and must confirm
  the deferrals the epic's stories registered rather than assume them closed.
