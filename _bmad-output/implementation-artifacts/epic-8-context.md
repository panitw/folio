# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks a typeface inside the designer — offline, no account, nothing fetched — lays out the document, and saves. An integrating developer renders that same `.folio` on a build box with no fonts installed and no network, and gets the same PDF, hash for hash. Two halves are missing today: nothing edits the document's `fonts` map (so a designer-created document offers exactly one chain), and nothing but the three build-time faces can be named. This epic makes chains authorable, makes a face something the file itself carries and the engine renders from, and turns choosing one into a search rather than a hand-edit.

## Stories

- Story 8.0: A stacked Thai mark reaches the page (precondition of 8.4)
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.4a: The canvas paints with the face the engine measured
- Story 8.4b: The canvas can name the face the engine measured
- Story 8.4c: The designer ships the typeface it specifies
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- Chains are authorable from the designer: create, rename, delete a chain, and add/remove/reorder entries within one. "Reorder" has its referent in entries only — the chain map has no authored key order and must not gain one.
- A face can be embedded in the template, keyed by content hash like every other asset, and referenced from a chain entry.
- A template must render from its embedded faces alone: no network, no host-installed font, no install step, no path read on the rendering machine.
- A family can be chosen from a curated, freely-licensed catalogue shipped inside the designer and working fully offline — no request leaves the machine, including at first use.
- A chain naming a font that is neither a shipped face nor a present, decodable asset fails with a located error naming the chain, the entry index and the key — never a substituted face, never a silent drop.
- Every field added is optional and absent-by-default, so the existing golden corpus must hash identically on all four CI targets. That is an acceptance criterion, not an aspiration. Every feature ships its own recorded-digest fixture, and a fixture must actually exercise what it claims to protect.
- Bold and italic are out of scope — they need a face-inventory decision left open elsewhere, not a consequence of embedding.
- CJK families are excluded from the embeddable catalogue (a full SC face is ~10.6 MB against 646 KB / 47 KB for the shipped Latin and Thai faces); the shipped SC face stays the coverage fallback. Added bytes are measured against the offline bundle budget.
- Redistributed faces keep their own terms: OFL text, NOTICE and copyright lines ship with them, the release licence manifest must be true of what actually ships, and the CI licence check covers them.

## Technical Decisions

- **`.folio` stays a single JSON text file.** Font bytes ride the existing content-addressed `assets` map — the mechanism images already use — so a person or an agent can still edit and render a template without the designer. A container format is a named non-goal.
- **Entry shapes.** A string entry is a shipped face name; a one-key object `{"asset": "<key>"}` is an embedded face; any other shape is a located load error. Chains mix both freely; first entry covering the rune wins.
- **Asset record.** A font asset differs from an image asset only in `mediaType` and an optional `font` record (family, style, licence, source). `font.family`/`font.style` are display identity only — **resolution is by asset key alone**, even where an embedded and a shipped face share a family name. Key rule, base64 wrapping, dedup and emission order are unchanged, so adding a font never moves an image.
- **`mediaType` is an OPEN set.** A recognised font media type whose bytes do not decode is a load error; an unrecognised one loads clean, is preserved verbatim, and errors at render only when something actually needs to draw it — with `Validate` predicting what render would do.
- **Format version.** The embedded-face *entry shape* (not the presence of a font asset) is the MAJOR trigger, and it joins the existing `2.0` rather than opening a new rank.
- **Canonical serialization is untouched** — same sorting, wrapping, byte-identical no-op round trip. Subsetting stays once per render inside the PDF producer; no face is subset at save time, and the subset tag stays a deterministic hash of the emitted subset program bytes, not of the glyph set.
- **The engine owns the document.** Every font-chain change travels as one opaque command producing one history entry; the browser holds no model of the `fonts` map and re-projects from the engine's answer. A rename updates the map and every element naming the old chain inside that same command, so one undo restores both.
- **`fontFamily` has exactly two attachment points** — an element's style and a table's header style. Both are live, so every rename walk and orphan check must read on both.
- **Refusals the engine must make:** deleting a chain elements still name, renaming onto an existing key, leaving a chain with no entries.
- **The browser never measures text.** The canvas gets every metric and line break from the engine's measure API through the same path the PDF render uses; this epic changes rasterization only. An embedded face's CSS family name is derived from its **asset key**, never from `font.family`, so an embedded "Inter" cannot collide with a shipped one in the browser's font registry.
- **Catalogue faces are static, single-instance faces derived ahead of the build** by the same replayable derivation the shipped set uses, with committed outputs and per-face NOTICE — never generated at build time, which would make the PDF a function of the build environment. They ship behind the same verified asset URLs as every other release asset.

## UX & Interaction Patterns

- The chain editor lives on the typography panel, opened from the family control — no separate mode, no dialog stack — because choosing a font and defining what a font *is* should not be two tools.
- The family control shows declared chains and catalogue entries as two visibly distinct groups: one is in the file, the other is not yet, and picking is what moves an entry between them.
- A catalogue pick is one command and one undo: it embeds the face, declares a chain naming it, and proposes a fallback tail of shipped faces for scripts the picked face lacks — a tail the author can edit. Picking the same family twice stores one copy. Font assets no chain names any longer are dropped on save.
- Refusals are reported in text at the control that caused them, following existing property-panel error, focus and accessible-name conventions. Voice is terse and technical: state the fact, name the location, offer no comfort. Everything stays keyboard-reachable with visible focus and accessible names.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4**, not a preface: it removes a fail-closed emission branch refusing any glyph with a non-zero vertical offset. Embedding arbitrary faces makes that branch far more reachable, so shipping 8.4 over it would newly stop documents rendering. The narrowed refusal stays pinned by a test rather than deleted.
- **8.1 → 8.2:** the panel can only report refusals the engine actually makes, so duplicate-name, orphan and empty-chain refusals belong to the engine story.
- **8.2 → 8.3:** 8.2 delivers only the negative half of entry display (the projected entry, unmodified — no parsing, no key detection, no extension stripping). 8.3 introduces the embedded-face entry shape and must extend display to family + style **and** move the projection's entry-shape validator in the same commit.
- **8.3 → 8.4:** embedding must exist before the engine can render from it; 8.4's "chain entry names a non-font asset" case requires all three of load-accepts, render-errors, validate-predicts.
- **8.4 → 8.4a:** 8.4 keeps the measurement half (asserted, not assumed); 8.4a owns rasterization — registering the carried face in the browser and re-authoring two guards written to forbid exactly this shape. Each guard is widened, never weakened, and the widened tie is scoped to the carried case, because shipped faces are deliberately disjoint.
- **8.4a → 8.4b → 8.4c:** 8.4a closes only one cause of the canvas/engine font mismatch. 8.4b registers each shipped face additionally under the engine's own face name, so the canvas vocabulary becomes the engine's by identity rather than by a mapping table, touching no chrome token. 8.4c ships real IBM Plex binaries for chrome, leaving chrome and canvas vocabularies separate by design rather than by accident.
- **8.5 → 8.6:** the catalogue must ship before picking from it can put a face in the file. 8.6 also makes font assets visible to the asset-reference walk so unreferenced faces are dropped. 8.5 is the story to trim rather than cut — the catalogue can ship with one family and grow by release.
- Story 7.10 (Epic 7's over-tall-element repair) is sequenced between 8.0 and 8.1 by owner call; it does not gate Epic 7's completion but must land before the `folio-go/v0.1.0` tag.
