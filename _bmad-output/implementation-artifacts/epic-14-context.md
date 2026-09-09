# Epic 14 Context: The designer's controls read as one product

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

The inspector, document bar, data panel and table editor were each built story by story, and the
seams show: competing button treatments mixed inside single rows, a Line whose thickness is called
"H" and whose colour is called "Background", a binding section offered on components the engine will
always refuse, a placed component that isn't selected, a data tree showing type strings instead of
the author's own values, a table drawn on the canvas as a box containing the word "Table", and a
table editor that is an eleven-column form where the design specifies a six-column matrix. This epic
makes those surfaces read as one designed product. It changes nothing about the document model, the
command surface, the file format, or a single rendered byte — only what the panels offer and how they
are spelled. No new functional requirements land; FR1, FR4, FR5, FR7 and FR10 are completed.

## Stories

- Story 14.1: One button vocabulary
- Story 14.2: A Line is a thickness and a colour; a Rectangle is a fill and a border
- Story 14.3: A placed component is the selected component
- Story 14.4: The panel offers no control the engine will refuse
- Story 14.5: The product wears its own mark
- Story 14.6: The DATA tab is the binding panel the design drew
- Story 14.7: The table editor is the matrix the design drew
- Story 14.8: The table editor carries the header, cell and border sections
- Story 14.9: The canvas draws the table it will print
- Story 14.10: A table column is bound from the main window

## Requirements & Constraints

- **No format change, no version increment, no new engine surface.** Every story is presentation.
  Where a panel needs data it lacks, the value already exists in an engine projection or the
  sample-data model. The one exception is 14.9, which adds a **canvas projection** field (column
  label, width, alignment, binding) — a projection field, never a format field.
- **Never invent a control the format cannot carry.** Any control not traceable to a field the engine
  consumes is absent rather than disabled-and-mysterious. This is named as how the epic's own defects
  were created.
- **Never normalise a document on selection.** A panel that stops offering a field must still
  preserve that field's stored value untouched on save, including for documents hand-edited into
  shapes the panel's vocabulary does not describe. Hiding a control that still paints costs
  discoverability, so the hidden-but-present value is disclosed with a note rather than vanished.
- **The panel must not offer what the engine will refuse.** Scalar binding is text-only; `params` is
  never a bindable root; a collection is not bindable by a text element. State the reason before the
  attempt, not after the engine rejects it.
- **The padding control is dead everywhere.** An owner ruling forbids the panel authoring padding on
  any element kind, including a table — no gate in the project would catch a violation, because the
  command it sends is legal, so it must not be built.
- **14.8 is a restyle only.** An owner decision reduced it to grouping the controls Epic 12 already
  ships into the drawn HEADER / CELLS / BORDERS sections. "Show header row" and the three-way borders
  preset have no format field and are not coming; the mockup is to be corrected, not implemented.
- **Accessibility is a hard floor:** every interactive element keyboard-reachable and operable,
  visible focus using the select token, an accessible name on every icon-only control, canvas hit
  targets larger than their visual footprint, and the table editor behaving as a data grid under
  keyboard navigation. A sweep replacing words with glyphs must never lose an accessible name.

## Technical Decisions

- **The engine owns the document.** No TypeScript model of a template exists. The UI holds an
  immutable snapshot for painting and sends every committed mutation as a command; undo/redo is
  engine-side history over those commands, and transient interaction state never enters the document.
  This is why a modal Cancel/Apply in the table editor is not a labelling choice — it implies a local
  uncommitted buffer, a second document model, which this rule exists to forbid.
- **The browser never measures text**, including on the canvas: metrics and line breaks come from the
  engine's measure API, and the canvas paints pre-broken lines in DOM/SVG. A richer canvas table must
  paint from the engine's projection, never from a browser-side model of the table.
- **Table geometry is derived, not stored twice.** Column widths are absolute and authoritative; a
  table's width *is* their sum and is never stored separately. The editor's width budget reads out
  that rule.
- **Millipoints are the stored unit and stay millipoints.** The **displayed** unit is **points,
  product-wide** — settled during 14.2. The mockups' millimetres are mockup fidelity only; a future
  story wanting millimetres is proposing a product-wide change, not a label.
- **The control vocabulary is written down and guarded** (delivered by 14.1): a control is a word; a
  glyph only as one member of a segmented control, drawn as a stroked SVG on the 16px grid; every
  control in one class or one named group is spelled the same way; a glyph always carries an
  accessible name that a respelling never narrows. Uniformity and accessible-name clauses are
  enforced by a contract test over the whole swept population, so a control added later is checked;
  the word-versus-glyph judgement on the remaining icon-only controls is deliberately deferred to
  the story that owns each surface.
- **Design tokens are the single source of styling** — no hard-coded hex anywhere; the brand mark's
  colour comes from the existing select token, whose value is already the design's cyan. Source-text
  contract tests pin the token file and several exact CSS declarations, so a restyle will red them
  until they are updated deliberately.
- Every mutation stays a single undo step, including a column binding made from the main window,
  which commits through the existing column-binding command unchanged. A multi-key panel edit (for
  example an orientation toggle) goes as one command carrying several changes, never as a sequence.

## UX & Interaction Patterns

- **The mockups are a reference, not a transcript.** Each has been measured to draw less than its
  citation implies — the main mockup draws a **Text** inspector only, with no Line inspector at all,
  and the table-editor mockup draws controls the format cannot carry. Build from the acceptance
  criteria and the vocabulary rule; verify a mockup actually contains a control before treating it as
  a specification for one.
- **Two-accent grammar, without exception.** Cyan means structure, focus and authority; amber means
  data and only data. Selection handles stay cyan even on a bound element; canvas binding
  placeholders are amber.
- **Voice is terse and technical** — state the fact, name the location, offer no comfort. Anything
  disabled or unpickable carries a stated reason, never a bare grey-out.
- **Density is a feature.** The table editor is the densest surface in the product and must be a
  matrix (columns × attributes), never a repeated single-column form; reorder and remove are row
  affordances, not extra columns.
- **One control per concept product-wide.** Alignment is one segmented control everywhere; a section
  must not mix icons and words at the same size in the same row; same-family actions (Open / Save /
  Save As) are spelled alike, and the design draws them as text.
- **The binding panel is docked, not a destination.** Its tree shows values beside paths, `{ }` and
  `[]` markers, badges for collections and runtime params, and dimmed unpickable rows, so the author
  recognises their data rather than decoding a type name. A context bar states what is selected and
  what a pick would bind, before the pick.
- **Canvas is approximate, preview is exact**, and that asymmetry must stay legible without a
  tutorial — the canvas table shows structure (one representative row), not the sample's full data.
- **All declared states must be handled**, including empty-table-with-no-columns and no-sample-data;
  an empty frame that looks like a rendering failure is a defect.
- Square corners throughout, 1px borders carrying separation, mono for anything a machine reads and
  sans for anything a person wrote.

## Cross-Story Dependencies

- **14.9 → 14.10.** Binding a column from the main window requires the canvas to draw the columns
  first; 14.10's selection target is the column 14.9 paints, on the projection field 14.9 adds.
  14.10 also carries an unnamed prerequisite: today's selection holds **element** ids only, so
  holding a column is a selection-model change.
- **14.6 → 14.10.** 14.6 dims row-scope fields as unpickable for a text element; 14.10 makes exactly
  those fields pickable when a column is selected, in the same tree and context bar.
- **14.1 → 14.2, 14.3, 14.4, 14.7.** The vocabulary rule is written down first; the per-panel sweeps
  are its consequence, each owning one of the surfaces 14.1 deliberately deferred, and 14.7's
  alignment control must be the one 14.1 settles.
- **14.4 ↔ 14.10.** 14.4 requires a table's binding to be stated once where it is editable; 14.10
  rules that "where" is the main window, leaving the editor's bound-field column as display.
- **14.8 depends on Epic 12's table styling capability** (header height, header style, alternating row
  colour) and adds no second way to store it.
- **14.7 needs a projection field it does not name** — the sample's item count for a table's
  read-only scope line is carried by nothing today — and its rebuild from eleven columns to six
  rewrites the grid's focus, move, and Home/End keyboard handling, which must survive intact.
- **Two rulings must still be recorded before their story is built:** whether a picked path in the
  DATA tab commits immediately or keeps an explicit commit control (14.6), and the table editor's
  transaction model — Close-plus-undo versus a deliberately ruled-in buffer (14.7). The third ruling
  this epic once owed, the display unit, is settled: points, product-wide.
