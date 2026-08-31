# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Today a document can name only the three faces the library happens to ship, and nothing in the
designer edits the document's `fonts` map — so every document created in the designer offers
exactly one family, inherited from the starter file. This epic makes typeface choice real end to
end: the author picks a family from a curated, freely-licensed catalogue inside the designer,
entirely offline; the chosen face travels inside the `.folio` file as a content-addressed asset;
and the engine renders that file on a machine with no fonts installed and no network, producing the
same bytes the author previewed. Two units matter throughout — the **face** is what the author
picks, the **chain** is what the document names (an ordered list resolved per rune for coverage).
The epic exists because the `.folio` file is the entire contract between the template author and
the integrating developer, and a font that only one of them can install is not a choice either can
rely on. It is post-MVP and adds no capability the format's existing asset mechanism cannot carry.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- **Authorable chains.** Chains and their entries can be created, renamed, reordered and deleted
  from the designer, so `fontFamily` names a family the author chose rather than a starter-file
  leftover.
- **Faces ride in the file.** A face is embedded in the template, keyed by content hash exactly as
  images already are, and referenced from a chain entry. One copy per distinct face; a face no
  chain names is dropped on save.
- **Render from the file alone.** A template renders from its embedded faces with no network, no
  host-installed font, and no install step on the rendering machine — the same prohibition that
  already governs images.
- **Located failure, never substitution.** A chain naming a font that is neither a shipped face nor
  a present, decodable asset is a load error naming the chain, the entry index and the key. Same
  for an asset whose bytes do not decode as its declared media type, or whose media type is outside
  the closed font set. Never a silent substitute, never a silent drop.
- **Offline catalogue.** The family catalogue ships with the designer and search-and-pick completes
  with no request leaving the machine, consistent with the designer's standing offline guarantee.
- **Byte identity is the acceptance bar.** Every field this epic adds is optional and
  absent-by-default, so the entire existing golden corpus must hash identically after it, on all
  four CI targets. This is a criterion, not an aspiration.
- **Licence provenance.** Every redistributed face carries its licence text and NOTICE recording
  the upstream release and hashes; nothing copyleft may enter at any depth.
- **Bundle weight is measured, not assumed.** The catalogue's added weight is recorded against the
  bundle budget and each entry's own weight is visible to whoever curates the list. CJK families are
  out of scope for the catalogue this epic (an order of magnitude heavier than the Latin and Thai
  faces); the shipped SC face remains the coverage fallback.

## Technical Decisions

- **`.folio` stays a single JSON text file.** No container format — that would end
  hand-editability and put entry order, timestamps and compression inside the byte-identity
  regime. Font bytes use the existing content-addressed `assets` map: key is the lowercase hex
  SHA-256 of the raw bytes, `data` base64 hard-wrapped at 76 columns, plus a font record carrying
  family, style, licence and source. No second storage shape, no new canonical-serialization rule,
  and `assets` emission order must not shift (adding a font never moves an image).
- **The engine owns the document.** There is no browser-side model of the `fonts` map. Every
  font-chain edit is one opaque command with one history entry; the designer re-projects from the
  engine's answer. Rename updates the naming elements inside the same command so one undo restores
  both. A delete that would orphan an element, or leave a chain empty, is refused with a located
  error.
- **Fonts are an explicit value, resolved by lookup.** Font resolution inside a render is a pure
  lookup against the supplied font set — never a host query. The chain is part of that set's
  identity, so the same template with a different chain is a different render, not a silent
  substitution. An embedded face joins the same per-rune coverage resolution as a shipped one, in
  declared order, keyed by asset — never by name, even where an embedded and a shipped face share
  a family name.
- **Subsetting stays where it is.** Once per render inside the PDF producer, with the subset tag a
  deterministic hash of the embedded program's bytes. No face is subset at save time, and no
  wall-clock value enters subset output.
- **One measurement authority.** The canvas gets every text metric from the engine, so a component
  using an embedded face previews through the same path that will produce the PDF.
- **Numbers and determinism.** Any number reaching an output byte goes through the PDF module's two
  emitters; geometry is integer millipoints and no `float64` appears under `internal/`. No
  unordered iteration may reach output.
- **Catalogue faces are built ahead of time.** Each is a static, single-instance face produced by
  the same replayable derivation the shipped set uses, with its output committed. Generating at
  build time would make the PDF a function of the build environment.
- **Bold and italic stay out.** They are stored and projected but consumed by nothing, and no
  weighted face ships. Giving them meaning is a separate face-inventory decision.
- **Version policy.** The rule that a higher MAJOR is a load error rather than a best-effort render
  means the format-version question must be settled and written into the format doc before any
  embedding code lands.

## UX & Interaction Patterns

- The chain editor lives on the typography panel, reachable from the family control — no separate
  mode and no dialog stack. Choosing a font and defining what a font is are one tool.
- Chain entries naming an embedded face read as family and style from the engine's projection,
  never as an asset key or a file name.
- Catalogue entries the document already declares and entries it does not are visibly distinct;
  picking is what moves an entry between the two groups.
- A refused edit states its concrete reason in text at the control that caused it, following the
  existing property-panel error, focus and accessible-name conventions. Voice is terse and
  technical: state the fact, name the location, offer no comfort.
- Every control is keyboard-reachable with visible focus and an accessible name; errors are
  distinguished by shape before colour.
- Undo covers every one of these mutations, and a catalogue pick is a single undoable step.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** The PDF emitter currently refuses any glyph with
  a non-zero vertical offset, which makes ordinary stacked-mark Thai fail to render at all. Epic 8
  admits arbitrary faces with arbitrary mark positioning, so shipping embedding over an unfixed
  fail-closed branch ships a feature that can newly stop documents rendering. 8.0 narrows that
  branch (it does not remove it) and must land first. A characterization test for it already exists.
- **8.1 → 8.2:** the chain editor UI consumes the engine commands and projection 8.1 introduces.
- **8.3 → 8.4:** the engine renders from embedded faces only once the format carries them.
- **8.5 → 8.6:** picking from the catalogue requires the catalogue to be bundled and offline-served.
- **8.6 depends on 8.1/8.3:** one pick must embed the face and declare a chain in a single command,
  which needs both the chain-command surface and the asset storage shape.
- **External sequencing:** Epic 7's over-tall-element repair (Story 7.10) is scheduled immediately
  after 8.0 and before 8.1 by owner call. It does not gate Epic 7's completion but must land before
  the `folio-go/v0.1.0` tag. 8.0 itself is deliberately *not* in the before-the-tag set: it only
  widens what renders, so it can never move an existing golden.
- **8.5 is the trim target** if the epic must be reduced — the catalogue can ship with one family
  and grow by release.
