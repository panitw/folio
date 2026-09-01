# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks a typeface from a catalogue inside the designer — offline, no account, nothing fetched — lays out the document, and saves. An integrating Go developer renders that same `.folio` on a build box with no fonts installed and no network, and gets the PDF she previewed, hash for hash. Today both halves are missing: nothing in the designer edits the document's `fonts` map (so a designer-created document offers exactly one family, `body`, from the starter file), and nothing but the three build-time faces can be named at all. This epic makes chains authorable, makes a face something the file itself carries, makes the canvas draw the face the engine actually measured, and turns choosing a font into a search rather than a hand-edit — because the `.folio` file is the whole contract between the author and the developer, and a font nobody can install is not a choice she can make.

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
- Story 8.4f: A bound nobody can cross silently
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file
- Story 8.4d: The size budget is a number something checks

## Requirements & Constraints

- **Authorable chains.** Create, rename, delete the document's font chains and reorder the *entries* within a chain from the designer, so `fontFamily` names a family the author chose. A document created blank in the designer must end up with more than one usable chain, with no hand-editing.
- **Embedded faces.** A face is embedded in the template, keyed by content hash like every other asset, and referenced from a chain entry. Two elements naming the same face store one copy; the same document saved twice is byte-identical.
- **Render from the file alone.** No network, no host-installed font, no path on disk, no install step on the rendering machine — at render time *or* on the designer's authoring path. The catalogue and its faces ship in the offline bundle.
- **Curated catalogue.** A freely-licensed, redistribution-permitting family set ships with the designer and is searchable with the browser offline; no request leaves the machine at any point in the pick flow.
- **Located failure.** A chain entry naming neither a shipped face nor a present, decodable font asset fails where it is written, naming the chain and the entry index — never silently substituted or dropped.
- **Byte identity is a constraint, not a target.** Every field this epic adds is optional and absent-by-default; the existing golden corpus must hash identically after it, across all four targets, and a no-op round trip must still rewrite the file byte-for-byte. This is an acceptance criterion.
- **Licence provenance.** Every embedded face carries a licence record so a file that travels alone states what it redistributes. Catalogue faces are restricted to redistribution-permitting licences, and an unidentifiable licence fails the build rather than warning.
- **First-load size budget** applies to every catalogue face added to the bundle, and to the weight each embedded face adds to every document using it.

## Technical Decisions

- **`.folio` stays a single JSON text file.** No zip, no directory package, no sidecar — the format exists so a person or an agent can edit a template without the designer and so a hand-written template renders. A container format is an explicit non-goal; revisit only if embedded document weight becomes measured, with the CJK weight case as the sole trigger.
- **Font bytes ride the existing content-addressed `assets` map** — the same mechanism images use: lowercase hex SHA-256 keys, base64 hard-wrapped at 76 columns, deduplicated, emitted in stable key order. No second storage mechanism and no new canonical-serialization rule. A font asset differs from an image asset only by its media type and its licence/identity record; the two are told apart by media type, never by inspecting bytes.
- **Chain entry shapes.** A JSON string is a shipped face name; a one-key object referencing an asset key is an embedded face. Any other shape is a load error. A chain may freely mix both — an embedded Latin face over shipped Thai/CJK fallback is the expected case. Ordering semantics are unchanged: first entry covering the rune wins. `fontFamily` still names a chain, never a face.
- **Format version: MAJOR.** A document carrying an embedded-face chain entry declares `2.0`, joining the `2.0` already opened by Epic 7 rather than opening a `3.0`; the library's supported-major ceiling does not move. The trigger is the *entry shape*, not the presence of a font asset — a font asset no chain references still loads and renders correctly on an older reader.
- **Media types are an open set.** A recognised font media type whose bytes are not that format is a load error; an unrecognised one loads clean, is preserved verbatim, and errors only at render, and only when a render actually needs that face.
- **Faces are stored whole — no save-time subsetting.** The data a template renders changes between saves, so a subset chosen at save time drops glyphs a later render needs. Subsetting stays exactly where it is: once per font per document, inside the PDF producer, byte-stably. An embedded face and a shipped face reach that path identically.
- **The engine owns the document.** Chain and entry edits travel as opaque engine commands, one command per author action, undoable as one history entry; the designer never invents a family name and holds no model of the document.
- **Catalogue faces are static, single-instance, prepared ahead of the build** — never generated at build time, which would make the PDF a function of the build environment. Two acceptable procurement routes: replayable derivation with committed outputs and both hashes for a face this project derives, or pinned version plus NOTICE for a vendored static face.
- **Out of scope by decision:** bold and italic (stored and projected today, consumed by no producer, and no weighted face ships — giving them meaning is a separate face-inventory decision); synthetic emboldening or obliquing; variable-font axes; live font services or arbitrary URLs; enumerating host-installed fonts; CJK families in the embeddable catalogue.

## UX & Interaction Patterns

- The chain editor lives **in the typography panel**, where fonts are chosen — choosing a font and defining what a font *is* must not be two different tools.
- The family control is a search-and-select over the engine-projected list of declared chains; the catalogue appears as a second, clearly separated group. Entries the document already declares and entries it does not must be visibly distinct — one is in the file, the other is not yet — and picking is what moves an entry between those groups.
- Picking a family does everything: embeds the face, declares a chain naming it, and proposes a fallback tail of shipped faces for scripts the picked face lacks, which the author can then edit. Picking the same family twice stores no second copy.
- Unreferenced font assets are dropped on save, so a file cannot accumulate megabytes of faces nothing draws with.
- All ten declared per-surface states apply to these new controls; undo covers every mutation they make; voice stays terse and technical, naming the location and offering no comfort; the accessibility floor (keyboard reachability, visible focus, accessible names on icon-only controls) applies.
- **Canvas fidelity is a stated goal of this epic, not a side effect.** The canvas must paint with the face the engine measured — an embedded face, and a shipped face by its real identity — rather than a fallback that happens to be installed. The canvas-approximate / preview-exact asymmetry remains the product concept; this narrows the gap where the face itself is concerned, and the page model stays blind to the emission stage.

## Cross-Story Dependencies

- **Story 8.0 is a precondition of Story 8.4, not a preface.** The PDF emission stage currently refuses any glyph carrying a non-zero vertical offset, so a class of ordinary Thai produces a hard render error and zero bytes. Embedding arbitrary faces means arbitrary mark positioning, so shipping 8.4 over an unfixed fail-closed branch ships a feature that can newly stop documents rendering. 8.0 must land first.
- **8.1 → 8.2:** chains must be editable through engine commands before the typography panel can host that editing.
- **8.3 → 8.4:** the format must carry a face before the engine can render from one; 8.4 is what proves the file-is-the-contract claim on a clean machine.
- **8.4a–8.4f** are the canvas-fidelity follow-ons split out of 8.4 and sequenced immediately after it; 8.4c (the designer shipping the typeface its own design system names) supplies the vendored-static-face precedent that 8.5's procurement rules rest on.
- **8.5 → 8.6:** a catalogue must exist before picking from it can put a face in the file. 8.6 also owns making font assets visible to the asset-reference sweep, without which unreferenced-font cleanup cannot work.
- **8.5 is the story to trim rather than cut** — the catalogue can ship with one family and grow by release. If 8.5 derives even one new face rather than vendoring, the derivation-bootstrap work comes into its scope.
- **Epic 7's over-tall-element repair (Story 7.10) is sequenced between 8.0 and 8.1** by owner decision; it does not gate Epic 7's completion but must land before the library's first tag.
- Constraints inherited unchanged from earlier epics — canonical serialization, the engine's ownership of the document, cross-target byte identity, and licence provenance — bind this epic rather than being targets of it.
