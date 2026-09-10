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
are spelled. No new functional requirements land; the designer-side requirements for component
properties, scalar binding from discovered JSON paths, and table-structure editing are completed.

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

⚠ **Refreshed 2026-09-10.** **14.1–14.7 and 14.7b are delivered**; **14.8 is next**, then 14.9 and
14.10, and then the Epic 14 boundary gate — which is not yet booked and must confirm DW-332's
document-bar fit at 1024px, DW-339's layout claims, DW-356's accessibility floor, and the
`assetCount` margin. (This paragraph previously read "14.6 is the next in the intended order"; it was
stale by four stories.) Story **14.7b** was split out of 14.7 on 2026-09-10 and is not in the list
above. The rules below, "written down first" by 14.1 and settled during 14.2–14.5, are now inherited
constraints rather than open questions.

## Requirements & Constraints

- **No format change, no version increment, no new engine surface.** Every story is presentation.
  Where a panel needs data it lacks, the value already exists in an engine projection or the
  sample-data model. The exceptions are named ones: 14.9 adds a **canvas projection** field (column
  label, width, alignment, binding), and 14.7 does **NOT** need one (corrected
  2026-09-10; this previously read "14.7 needs a projection field for a table's sample item count").
  Measured: `SampleNode.count` and `components[].band` are already in `App.tsx` scope at the
  `<TableEditor>` call site, and the width budget is `available = band.width - table.x` with a table's
  `component.width` already the sum of its column widths. Going through Go for data already in hand
  would cost a six-file `hasExactKeys` wire dance whose failure mode is a silently terminated worker.
  Projection fields, never format fields.
- **Never invent a control the format cannot carry.** Any control not traceable to a field the engine
  consumes is absent rather than disabled-and-mysterious. This is named as how the epic's own defects
  were created.
- **Never normalise a document on selection.** A panel that stops offering a field must still
  preserve that field's stored value untouched on save, including for documents hand-edited into
  shapes the panel's vocabulary does not describe. Hiding a control that still paints costs a working
  capability, so the hidden-but-present value is disclosed with a note rather than vanished.
- **The panel must not offer what the engine will refuse.** Scalar binding is text-only; runtime
  `params` are never a bindable root; a collection is not bindable by a text element. State the
  reason before the attempt, not after the engine rejects it. One live case runs the other way and
  needs a check of its own: the panel can offer an **empty collection** that the engine *accepts* and
  the render then fails — agreement between panel and command layer cannot catch that one.
- **The padding control is dead everywhere.** An owner ruling forbids the panel authoring padding on
  any element kind, including a table — no gate in the project would catch a violation, because the
  command it sends is legal, so it must not be built.
- **14.8 is a restyle PLUS exactly one new capability.** ⚠ **Corrected 2026-09-10** — this bullet said
  "a restyle only", which was true of the owner's 2026-09-05 decision and is **no longer true**. The
  bulk of the story is still grouping the controls Epic 12 already ships into the drawn HEADER and
  CELLS sections, and "Show header row" and the three-way None/Horizontal/All borders preset still
  have **no format field and are not coming**; the mockup is to be corrected, not implemented.
  **But the owner amended that decision on 2026-09-10:** the BORDERS section authors
  `headerStyle.border` — the header row's own border, distinct from the table's — extending the
  closed command field set at `component_commands.go` from nine fields to ten. `epics.md` calls this
  **"capability rather than presentation"** in the criterion itself. It was ruled in knowingly,
  against a recommendation to split it into its own story, because the engine's own source comment
  says `border` *"waits on Story 14.8's BORDERS section"* — so emptying BORDERS would have stranded a
  deferral pointed at this story. **Dispatching 14.8 from the old bullet would hand the builder a
  scope smaller than the owner ruled.** Read `epics.md` §14.8 directly; it is the authority, and it
  also carries the story's principal correctness risk (the header border cascade is block-granular,
  so setting one attribute replaces the whole block).
- **Accessibility is a hard floor:** every interactive element keyboard-reachable and operable,
  visible focus using the select token, an accessible name on every icon-only control, canvas hit
  targets larger than their visual footprint, and the table editor behaving as a data grid under
  keyboard navigation. A sweep replacing words with glyphs must never lose an accessible name.
- **An "is not offered / is preserved" claim needs a test that ADDS the forbidden thing.** Reverting
  the implementation cannot falsify an absence claim; it makes it pass more easily. The add-mutation
  must be applied at every position the forbidden thing could occupy — on the subject, on its
  container, and on any wrapper between — and a preservation test must read the value back rather
  than count dispatches. Four false greens in this epic were produced by the revert-shaped proof.

## Technical Decisions

- **The engine owns the document.** No TypeScript model of a template exists. The UI holds an
  immutable snapshot for painting and sends every committed mutation as a command; undo/redo is
  engine-side history over those commands, and transient interaction state never enters the document.
  This is why a modal Cancel/Apply in the table editor is not a labelling choice — it implies a local
  uncommitted buffer, a second document model, which this rule exists to forbid. The shipped editor
  already commits every cell on blur, so this is settled in practice, not open.
- **The browser never measures text**, including on the canvas: metrics and line breaks come from the
  engine's measure API, and the canvas paints pre-broken lines in DOM/SVG. A richer canvas table must
  paint from the engine's projection, never from a browser-side model of the table.
- **Table geometry is derived, not stored twice.** Column widths are absolute and authoritative; a
  table's width *is* their sum and is never stored separately. The editor's width budget reads out
  that rule.
- **Millipoints are the stored unit and stay millipoints.** The **displayed** unit is **points,
  product-wide.** The mockups' millimetres exist nowhere in the product, so that is a
  mockup-fidelity decision, not a live two-unit conflict; a future story wanting millimetres is
  proposing a product-wide change.
- **The control vocabulary is written down and guarded** (delivered by 14.1): a control is a word; a
  glyph only as one member of a segmented control, drawn as a stroked SVG on the 16px grid; every
  control in one class or one named group is spelled the same way; a glyph always carries an
  accessible name that a respelling never narrows. Uniformity and accessible-name clauses are
  enforced by a contract test over the whole swept population, so a control added later is checked;
  the word-versus-glyph judgement on the remaining icon-only controls is deferred to the story that
  owns each surface.
- **Design tokens are the single source of styling** — no hard-coded hex anywhere; the brand mark's
  colour comes from the existing select token, whose value is already the design's cyan, and the mark
  is one size-parameterised component used at exactly two sizes (document bar and load screen) and
  nowhere else. Source-text contract tests pin the token file, an exact allowlist of CSS
  declarations, and several verbatim lines, so **any restyle in this epic reds them** until they are
  updated deliberately.
- Every mutation stays a single undo step, including a column binding made from the main window,
  which commits through the existing column-binding command unchanged. A multi-key panel edit (for
  example an orientation toggle) goes as one command carrying several changes, never as a sequence.

## UX & Interaction Patterns

- **The mockups and this epic's own prose are a reference, not a transcript.** Both have been
  measured to overstate: the main mockup draws a **Text** inspector only, with no Line inspector at
  all; the table-editor mockup draws controls the format cannot carry; and several acceptance
  criteria describe as missing what is already built or already in the data. Verify a control exists
  in the mockup, and verify the defect still exists in the product, before building from either.
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
  what a pick would bind, before the pick — and the panel's standing prose must stop soliciting a
  pick it is about to refuse.
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
- **14.4 → 14.6.** 14.4 closed the *command* path but left the DATA tab's prose still inviting the
  refused pick, left `params.*` candidates offered undimmed, and left a generic refusal for a Table —
  the one kind that legally does take a (collection) binding. All of these are deliberately held for
  14.6, which replaces that surface's presentation entirely; patching them earlier writes wording
  14.6 deletes.
- **14.1 → 14.2, 14.3, 14.4, 14.7.** The vocabulary rule was written down first; the per-panel sweeps
  are its consequence, each owning one of the surfaces 14.1 deliberately deferred, and 14.7's
  alignment control must be the one 14.1 settled.
- **14.4 ↔ 14.10 — these are two different bindings, and must not be conflated.** 14.10 is the
  **per-column** bound field; 14.4's "state a table's binding where it is editable" is the **table's
  own collection**, a different value and a different command, and 14.7 explicitly *keeps* the
  collection and row alias editable inside the table editor. 14.10 does not settle where the
  collection is edited.
- **14.8 depends on Epic 12's table styling capability** (header height, header style, alternating row
  colour) and adds no second way to store it.
- **14.7's rebuild from eleven columns to six rewrites the grid's focus, move and Home/End keyboard
  handling** — the largest mechanical risk in the epic — and that navigation must survive intact.
- **One ruling is still owed before its story is built:** whether a picked path in the DATA tab
  commits immediately (undoable) or keeps an explicit commit control (14.6). The other two this epic
  once owed are settled: the display unit is points product-wide, and the table editor's transaction
  model is commit-on-blur plus undo.
