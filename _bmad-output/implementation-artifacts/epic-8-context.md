# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks the firm's typeface from a catalogue inside the designer — offline, no account, nothing fetched — lays out the document, and saves. An integrating Go developer renders that same file on a build box with no fonts installed and no network, and gets the PDF she previewed, hash for hash. Today both halves are missing: nothing edits the document's `fonts` map, so every document offers exactly the one family the starter file declares, and nothing but the three build-time faces can be named at all. This epic makes font chains authorable, makes a face something the file itself carries and the engine renders from, and turns choosing one into a search rather than a hand-edit — so that the `.folio` file remains the whole contract between author and integrator.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.4a: The canvas paints with the face the engine measured
- Story 8.4b: The canvas can name the face the engine measured
- Story 8.4c: The designer ships the typeface it specifies
- Story 8.4e: A shipped face carries its identity to the fragment
- Story 8.4g: The bundle is a function of the source, not of the tree
- Story 8.4f: A bound nobody can cross silently
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file
- Story 8.4d: The size budget is a number something checks

## Requirements & Constraints

- Font chains and their entries are authorable from the designer: create, rename, delete a chain; add, move, remove entries within one. A chain is an ordered list; the chain map itself has no authored key order and must not acquire one.
- A face is embeddable in the template, keyed by content hash like every other asset, and referenced from a chain entry.
- A template renders from its embedded faces alone — no network, no host-installed font, no install step on the rendering machine.
- A family is choosable from a curated, freely-licensed catalogue that ships with the designer and works offline.
- A chain naming a font that is neither a shipped face nor a present, decodable asset fails with a located error.
- Byte-reproducibility is absolute: same template + data + params + engine version yields identical bytes across darwin/arm64, linux/amd64, linux/arm64 and js/wasm. Every field this epic adds is optional and absent-by-default, so the existing golden corpus must hash identically after the epic — an acceptance criterion, not an aspiration.
- Subsetting stays byte-stable and happens once per render inside the PDF producer, never at save time. The subset tag derives from the emitted subset program's own bytes, not from the glyph-id set.
- Licence obligations are enforced, not described: catalogue faces must carry an acceptable permissive licence and travel with licence text and copyright lines; an unidentifiable licence fails the build rather than warning.
- Bundle weight is a governed number: the first-load budget and the release's cache-asset ceiling must be enforced by gates that fail, with figures declared in one place and read by a check rather than by a human.
- Bold and italic are explicitly out of scope; they are Epic 11's subject.

## Technical Decisions

- **The file format stays a single JSON text file.** Font bytes ride the existing content-addressed `assets` map — the same mechanism images use — with no second storage shape and no new canonical-serialization rule. A container format is a recorded non-goal because it would end hand-editability and put entry order, timestamps and compression inside the byte-identity regime.
- **Canonical serialization is unchanged:** sorted keys, two-space indent, LF, trailing newline; assets keyed by lowercase hex SHA-256 with base64 hard-wrapped at 76 columns. Adding a font must not move an image in `assets` emission order, and an edited-then-reverted document must round-trip to the original bytes.
- **The entry shape change is a format-version event** and joins the existing major rank rather than adding a new one; the trigger is the entry shape (an object where a string was expected), not the mere presence of a font asset. The format document is updated before any code lands.
- **Resolution is by asset key alone.** Face family and style are display identity only, never used to resolve or substitute — including where an embedded face and a shipped face share a family name. This extends one layer down into the browser: an embedded face's CSS family name derives from its asset key, never from its declared family.
- **The engine owns the document.** There is no TypeScript model of the `fonts` map. Every chain edit travels as one opaque command producing one history entry; the designer re-projects from the engine's answer. A rename updates the map and every element naming the old chain inside one command, so one undo restores both.
- **`fontFamily` has exactly two attachment points** — an element's style and a table's header style. Both are live and both fail identically at render, so every orphan check, rename walk and refusal must read on both.
- **Refusals belong to the engine, not the panel:** orphaning deletes, empty chains and duplicate names are refused by the engine with located errors that the panel then reports.
- **The browser never measures text.** Metrics and line breaks come from the engine's measure API on the canvas too; the browser contributes rasterization only. The canvas must therefore be able to *name* the face the engine measured with, and must attribute faces per fragment rather than relying on a fixed stylesheet stack.
- **Media types are never a closed set.** An unrecognised font media type is preserved at load and errors at render; a recognised type whose bytes do not decode is a load error. A chain entry naming a non-font asset must be accepted at load, error at render, and be predicted by validation.
- **No number reaches an output byte except through the normative emitters**, in integer fixed-point, with no floating-point under the engine's internals.
- **Every feature ships a golden fixture** with a recorded digest, and the fixture's document must actually exercise the feature it guards. A moved hash is a defect until proven an intended, versioned change.
- **Guards are widened or replaced, never weakened**, with the old state red-proved; and build output must be a function of the source rather than the working tree, so recorded byte figures mean something.

## UX & Interaction Patterns

- The chain editor lives on the typography panel, reached from the family control on the same panel — no separate mode and no dialog stack. Choosing a font and defining what a font *is* are one tool.
- Catalogue entries the document already declares and entries it does not are visibly distinct; picking is the act that moves an entry from the second group to the first.
- Errors state the concrete reason in text at the control that caused them, in the product's terse, technical voice — the fact and the location, no comfort.
- The accessibility floor applies: keyboard reach and operability, visible focus, accessible names on icon-only controls, and diagnostics distinguished by shape before colour.
- Every declared surface state must be implemented, including the loading, populated, diagnostic and error states of the font-picking flow.
- Undo covers every mutation the author commits, and a font pick is one undoable step.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4**, not a preface: the emission path currently refuses any glyph with a vertical offset, and arbitrary embedded faces reach that branch far more often than the shipped set does.
- **8.1 → 8.2:** the engine-side refusals and command shape must exist before the panel can report them.
- **8.2's positive half depends on 8.3:** there is no font-record projection to display until the embedded-face entry shape exists; 8.3 must move the projection's entry-shape validator in the same commit that changes the shape.
- **8.3 → 8.4:** faces must travel in the file before the engine can render from them.
- **8.4a follows 8.4 immediately** (canvas rasterization split from the engine half). **8.4b precedes 8.4c**, which precedes **8.4d** (the enforceable budget figure cannot be written until the typeface work has landed). **8.4e** extends 8.4a's per-fragment attribution to shipped faces and closes the canvas-fidelity thread.
- **8.4g lands before 8.4f, 8.5 and 8.4d**, so every byte figure those stories record is trustworthy.
- **8.4f lands before 8.5**, so the cache-asset ceiling can never again be crossed silently by a catalogue.
- **8.6 depends on 8.3 and 8.5**: a pick embeds a face and declares a chain in one command, and must drop font assets no chain names any longer.
- **Epic 11 owns bold and italic** (realize or retire); nothing in this epic gives them meaning.
- 8.5 is the story to trim rather than cut — the catalogue may ship with one family and grow by release. CJK families are excluded from the catalogue in this epic; the shipped CJK face remains the coverage fallback.
