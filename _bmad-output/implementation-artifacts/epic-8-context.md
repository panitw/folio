# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Today a document can name only the three faces the library ships, and nothing in the designer edits
the `fonts` map, so every new document offers exactly one family from the starter file. This epic
makes fonts an authored choice the file carries: the author edits font chains in the designer, picks
a family from a curated offline catalogue, and the chosen face travels inside the `.folio`. An
integrating developer then renders that file on a build box with no fonts installed and no network
and gets the previewed PDF hash for hash — because the `.folio` is the whole contract between the
two of them, and a font nobody can install is not a choice the author can make.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.4a: The canvas paints with the face the engine measured
- Story 8.4b: The canvas can name the face the engine measured
- Story 8.4c: The designer ships the typeface it specifies
- Story 8.4d: The size budget is a number something checks
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- **Authorable chains.** Create, rename, delete the document's font chains and reorder the entries
  within one, so `fontFamily` names a family the author chose. "Reorder" is entry-level only — a
  chain is an ordered list; the `fonts` map has no authored key order and must not gain one.
- **Faces embedded in the file**, keyed by content hash like every other asset, referenced from a
  chain entry — and rendered from those bytes alone: no network, no host-installed font, no install
  step, in the engine or the designer.
- **Located failure.** A chain naming a font that is neither a shipped face nor a present, decodable
  asset errors with the chain, the entry index and the key — never a substituted face, never a
  silent drop.
- **Curated offline catalogue** ships with the designer; the whole search-and-pick flow completes
  with no request leaving the machine.
- **Byte identity is the acceptance bar.** Every field added is optional and absent-by-default, so
  the existing golden corpus must hash identically after each story on all four CI targets. A new
  emission path is entered only under its own trigger condition, and that is asserted, not assumed.
- **Licence provenance.** Every redistributed font binary travels with its licence text and NOTICE,
  is covered by the CI licence check, and the release licence manifest must be true of what ships.
- **Size budget.** Added weight is measured against the first-load budget, and by the epic's end
  that budget is a declared figure a gate reads and fails on.
- **Out of scope.** Bold and italic gain no new meaning (no weighted face ships); CJK is excluded
  from the catalogue, with the shipped SC face as coverage fallback; a container format is a
  non-goal — `.folio` stays one JSON text file so a hand-written template still renders.

## Technical Decisions

- **Fonts are an explicit value, never a host query.** The engine takes a `FontSet`; resolution is a
  pure lookup, and the chain is part of the render's identity. **The asset key decides** which face
  is used, even where an embedded and a shipped face share a family name; a face record's
  family/style are display identity only. This extends into the browser: an embedded face's CSS
  family name derives from its asset key, never from its declared family.
- **One canonical byte form.** Font bytes ride the existing content-addressed `assets` map — the
  image mechanism, no second storage shape, no new serialization rule — and adding a font must not
  move an existing asset. Edit-and-edit-back reproduces the original bytes.
- **Media types are never a closed set.** An unrecognised media type is *preserved at load* and
  errors at *render*; a recognised type whose bytes do not decode is a *load* error. `Validate` must
  predict what `Render` would do — that third half is the one with no visible symptom when missing.
- **Format version.** The new chain-entry shape (an object where a 1.x reader expects a string) is
  MAJOR and joins the existing `2.0` rank — no new constant. Written into the format doc before code.
- **The engine owns the document.** No TypeScript model of a `.folio`. Every chain mutation is one
  opaque command with one history entry; the designer re-projects from the engine's answer. A rename
  updates every element naming the old chain in that same command, so one undo restores both.
- **`fontFamily` has exactly two attachment points** — an element's style and a table's header style
  — and both are live. Every walk, rename, orphan count and refusal reads on both.
- **The browser never measures text.** The canvas takes every metric and line break from the
  engine's measure API; this epic changes rasterization only, and the shared measurement path stays
  asserted.
- **Subsetting** happens once per render inside the PDF producer, never at save time; the subset tag
  hashes the emitted subset program bytes, not the glyph-id set.
- **Numbers** reaching output bytes go through the single emission path in fixed-point millipoints,
  no `float64` in the render path.
- **Fixtures.** No change lands without a golden fixture carrying a recorded digest, and a fixture
  must exercise what it claims — Latin text over a carried Thai face observes nothing.
- **Guards are widened, never weakened.** Several existing tests were written to forbid exactly the
  shapes this epic introduces; each is re-authored to a narrower still-meaningful rule (its scope is
  load-bearing), never deleted or softened to pass. A deliberate intermediate state is pinned by an
  assertion naming its successor story, not defended by a comment. A guard for a new class of
  artifact lands in the same commit as the first artifact of that class.
- **Name what a test asserts, never a line number** — every line anchor carried across a story
  boundary in this run has rotted.

## UX & Interaction Patterns

- The chain editor opens from the family control on the typography panel — no separate mode, no
  dialog stack: choosing a font and defining what a font *is* are one tool.
- A chain entry displays as the projected entry, unmodified — never an asset key, file name, or
  anything parsed out of one. Once the embedded shape exists it reads as family and style.
- Refusals (orphaning delete, emptied chain, duplicate name) state the concrete reason in text at
  the control that caused them, following existing property-panel error, focus and accessible-name
  conventions. Voice is terse and technical: state the fact, name the location, offer no comfort.
- In the catalogue, entries the document already declares and those it does not are visibly
  distinct; picking is what moves an entry between the two groups.
- A pick is one command and one undo — it embeds the face and declares the chain together. Picking
  an already-embedded family stores no second copy and offers the existing chain.
- Accessibility floor applies: keyboard reach, visible focus, accessible names on icon-only
  controls, errors distinguished by shape before colour.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** The renderer currently refuses any glyph carrying
  a vertical offset, so arbitrary embedded faces could newly stop documents rendering; it must land
  before embedding ships. A characterization test already exists as its red before-state.
- **8.1 before 8.2** — the panel reports what the engine answers, so every refusal it shows must
  exist in the engine first.
- **8.2 → 8.3 forward obligation.** 8.2 delivers only the negative half of entry display; 8.3
  introduces the embedded-entry shape and must extend the display and move the projection's
  entry-shape validator in the same commit.
- **8.3 before 8.4** — the engine cannot render a face the format cannot carry.
- **8.4 → 8.4a → 8.4b → 8.4c → 8.4d, in that order.** 8.4 ships the engine half and discloses the
  canvas gap as a test naming 8.4a; 8.4a gives carried faces a browser family; 8.4b makes the canvas
  name the engine's shipped faces; 8.4c changes which files sit behind those names and updates
  8.4b's pin; 8.4d can only declare its figure after 8.4c's weight lands.
- **8.5 before 8.6** — the catalogue must exist before picking from it is a command. 8.6 also owns
  the cleanup of font assets no chain names.
- **Story 7.10** (Epic 7's over-tall-element repair) is sequenced between 8.0 and 8.1 by owner call:
  it does not gate Epic 7, but must land before the `folio-go/v0.1.0` tag.
- 8.5 is the story to trim rather than cut — the catalogue may ship with one family and grow.
