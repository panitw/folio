# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Today a document can only name the three faces the library compiles in, and nothing in the designer
edits the `fonts` map — so every document created there offers exactly one family, inherited from the
starter file. This epic closes both halves: an author edits the document's font chains from the
typography panel, picks a family from a curated catalogue that ships inside the designer and works
offline, and the face itself rides inside the `.folio` file. An integrator then renders that file on
a build box with no fonts installed and no network and gets the same PDF, hash for hash. The `.folio`
file is the whole contract between author and integrator, so a font nobody can install is not a
choice the author can actually make.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.4a: The canvas paints with the face the engine measured
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- Font chains are authorable from the designer: create, rename, delete a chain, and reorder the
  entries inside one. "Reorder" has its referent in entries only — the chain map has no authored key
  order and must not acquire one.
- A face can be embedded in the template, content-hash keyed like every other asset and named from a
  chain entry; a template renders from its embedded faces alone — no network, no host-installed font,
  no install step on the rendering machine.
- A chain naming a font that is neither a shipped face nor a present, decodable asset fails with a
  located error — never a silent substitution, never a dropped entry.
- The catalogue works fully offline: no font-service request, no first-use download.
- **Byte identity is the hard gate.** Every field this epic adds is optional and absent-by-default,
  so the whole existing golden corpus must hash identically after each story, on all four targets.
- Bold and italic are out of scope: no weighted or sloped face ships, and giving those flags meaning
  is a face-inventory decision, not a consequence of embedding.
- CJK families are excluded from the catalogue this epic (a full SC face is ~10.6 MB against 646 KB
  and 47 KB for the shipped Latin and Thai faces); the shipped SC face stays the coverage fallback.
  Catalogue weight is measured against the bundle budget, and each face travels with its licence text
  and a NOTICE recording upstream release and hashes. Nothing copyleft enters, at any depth.

## Technical Decisions

- **The author picks a face; the document names a chain** — an ordered list resolved *per rune* for
  coverage. Resolution is a pure lookup against the `FontSet` passed to `Render`, never a host font
  query, and the chain is part of that render's identity.
- **Resolution is by asset key alone.** An embedded face's family/style are display identity only,
  never used to resolve or substitute — including where an embedded and a shipped face share a family
  name. This extends into the browser: a canvas CSS family name for a carried face is derived from the
  asset key, or an embedded "Inter" collides with a shipped one in the font registry.
- **`.folio` stays a single canonical JSON text file.** Font bytes ride the existing
  content-addressed `assets` map — the mechanism images already use — with no second storage shape,
  no new canonical-serialization rule and no container format. Dedup is by content hash, and adding a
  font must not move an image in emission order.
- **`mediaType` is never a closed set.** An unrecognised font media type is *preserved* at load and
  errors at *render*; a recognised type whose bytes do not decode is a *load* error. `Validate` must
  predict what `Render` would do.
- **Format version:** the embedded-entry shape is a MAJOR change joining the existing `2.0` rank — no
  new constant, no renumbering — written into the format document before code lands.
- **The engine owns the document.** Every chain edit travels as one opaque command with one history
  entry; the designer re-projects from the engine's answer and holds no TypeScript model of the
  `fonts` map. One undo restores both a rename and every element it rewrote.
- **The browser never measures text**, canvas included: metrics and line breaks come from the engine's
  measure API; the browser rasterizes only. The canvas fragment rule follows the engine's chain order,
  not a CSS token — otherwise Latin text gets the Thai face's Latin glyphs.
- **Determinism rules that bind here:** no number reaches an output byte except through the single
  emitter file, in millipoints, with no `float64` under `internal/`; subsetting happens once per
  render inside the PDF producer, never at save time; the subset tag hashes the *emitted subset
  program bytes*, not the glyph set. Catalogue faces are derived ahead of the build by the same
  replayable derivation the shipped set uses, with committed output — generating at build time would
  make the PDF a function of the build environment.
- **Every feature ships its golden fixture** with a recorded digest, and a font fixture must draw text
  the embedded face actually covers or the digest observes nothing. Diagnostics are one type on one
  channel with a stable code; a font-chain error names the chain and the entry index.

## UX & Interaction Patterns

- The chain editor lives on the typography panel, reachable from the family control — no separate mode,
  no dialog stack, because choosing a font and defining what a font *is* should not be two tools.
- Refusals surface as concrete text at the originating control, following existing property-panel
  error, focus and accessible-name conventions, in terse technical voice that names the location.
  Keyboard reach, visible focus and accessible names on icon-only controls are obligations.
- In the catalogue, families the document already declares and those it does not are visibly distinct
  — picking is what moves an entry between the groups — and a pick proposes a chain whose tail is the
  shipped faces for uncovered scripts, still editable.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** The emitter currently refuses any glyph with a
  non-zero vertical offset and returns zero bytes; arbitrary embedded faces reach that branch far more
  often than the shipped set does, so shipping 8.4 over it would newly stop documents rendering. 8.0
  narrows the fail-closed branch rather than deleting it.
- **8.1 before 8.2** — the panel reports what the engine answers, so every refusal (orphaning delete,
  empty chain, duplicate name) must exist in the engine first. `fontFamily` has exactly two attachment
  points, element style and table header style; both are live and every rule reads on both.
- **8.2 → 8.3 forward obligation.** 8.2 displays a chain entry unmodified because no font-record
  projection exists yet; 8.3 introduces the embedded-entry shape and must extend the display and move
  the projection's entry-shape validator in the same commit.
- **8.3 before 8.4**, which renders what 8.3 stores. **8.4a is sequenced immediately after 8.4** and
  owns the canvas paint half (8.4 pins the measurement half only, disclosing the canvas limitation as
  a test). 8.4a must widen, never weaken, the two designer guards that currently forbid dynamic font
  registration.
- **8.5 before 8.6**, which picks from the catalogue 8.5 ships. 8.6 also makes font assets visible to
  asset-reference counting so unreferenced faces are dropped on save. 8.5 is the story to trim rather
  than cut — the catalogue can ship with one family and grow by release.
- **Story 7.10** (Epic 7's over-tall-element repair) is sequenced immediately after 8.0 and before
  8.1 by owner call; it must land before the `folio-go/v0.1.0` tag.
