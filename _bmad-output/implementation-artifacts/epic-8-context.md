# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks the firm's typeface from a catalogue inside the designer — offline, no
account, nothing fetched — lays out the document, and saves. An integrating Go developer renders
that same `.folio` on a build box with no fonts installed and no network, and gets the PDF the
designer previewed, hash for hash. Both halves are missing today: nothing in the designer edits the
document's `fonts` map (so a template created from blank offers exactly one chain, `body`), and
nothing but the three build-time faces can be named at all. This epic makes chains authorable,
makes a face something the file carries and the engine renders from, and makes choosing one a
search rather than a hand-edit — because the `.folio` file is the whole contract between the two
users, and a font nobody can install is not a choice an author can make.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- **`.folio` stays a single JSON text file.** No zip, no directory package, no sidecar — the format
  exists so a person or an agent can edit a template without the designer and so a hand-written
  template renders. A container ends both and drags entry order, timestamps and compression into the
  byte-identity regime.
- **Font bytes ride the existing `assets` map** — the mechanism images already use: content-addressed
  by lowercase hex SHA-256 of raw bytes, base64 hard-wrapped at 76 columns, deduplicated by key,
  emitted in stable key order. No second storage shape, no new serialization rule, and adding a font
  must never move an image.
- **Nothing is fetched or read at render time**: no network, no host-installed font, no path on disk.
  The authoring path is offline too — catalogue and faces ship in the offline release bundle behind
  the service worker's verified asset URLs, with no call to any font service.
- **Embedded faces are stored whole.** No save-time subsetting — what a template renders changes
  between saves, so a save-time subset drops glyphs a later render needs. Subsetting stays once per
  render inside the PDF producer, its tag a deterministic hash of the font program's bytes.
- **Every embedded face carries a licence record** (identifier plus source/instance provenance), so a
  file that travels alone states what it redistributes. The catalogue is restricted to
  redistribution-permitting licences.
- **Byte identity is an acceptance criterion, not an aspiration.** Every field this epic adds is
  optional and absent-by-default, so the existing golden corpus must hash identically after every
  story, on all four CI targets. A no-op round trip still rewrites the file byte-for-byte.
- **A broken font reference fails where it is written.** An absent asset key, a media type outside the
  closed set, or bytes that do not decode as their declared type is a located load error naming the
  chain, entry index and key — never a substituted face, never a silent drop.
- **Out of scope:** bold and italic (stored and projected today, consumed by no producer, no weighted
  face ships — a face-inventory decision left open, not a consequence of embedding); synthetic
  bold/oblique; variable-font axes; host fonts; any live font service; CJK in the embeddable
  catalogue (10.6 MB for a full SC face against 646 KB and 47 KB for the shipped Latin and Thai
  faces — CJK stays on the shipped-face path as coverage fallback).

## Technical Decisions

- **Chain semantics unchanged.** A chain is an ordered list resolved per rune for coverage, first
  covering entry wins. `style.fontFamily` names a *chain*, never a face. An entry is a JSON string (a
  shipped face name) or a one-key object `{"asset": "<key>"}` (an embedded face); any other shape is
  a load error. Chains may mix both. **Resolution is by asset key alone** — the stored family/style
  are display identity only and never substitute. An empty chain is still not nameable.
- **The engine takes fonts as an explicit value.** Resolution inside a render is a pure lookup against
  the supplied `FontSet`; an embedded face and a shipped face must reach the PDF producer
  identically. No package under `internal/` embeds font data.
- **`fontFamily` has exactly two attachment points** — an element's own `style` and a table's
  `headerStyle`. Both are live and fail at render identically, so every rename, orphan check and
  reference count must walk both.
- **`fonts` has no authored key order, and that absence is load-bearing** — the format reasons from it
  to rule out a font default ("the first key was never well-defined"). Chain-key order is
  inapplicable, not deferred; reorder has a referent only in a chain's *entries*, and panel display
  order is designer UI state needing no format change.
- **In the designer, the engine owns the document.** No TypeScript model of a `.folio`. Every chain
  and entry edit travels as one opaque command producing one history entry (so one undo restores the
  map and the elements it touched together), and the UI re-projects from the engine's answer rather
  than writing the `fonts` map itself.
- **Version rule:** `version` describes the document, not the writer — a file declares the lowest
  version its content requires; saving raises it, never lowers it. A higher MAJOR is a load error,
  never a best-effort render. The additive font fields point at MINOR, but the choice must be settled
  and written into the format contract before code lands.
- **Catalogue faces are derived ahead of the build**, by the same replayable derivation the shipped
  set uses, with committed outputs, licence text and a NOTICE per face recording upstream release and
  hashes. A face generated at build time makes the PDF a function of the build environment.
- **Determinism plumbing that binds here:** every number reaching an output byte goes through the PDF
  package's two emitters and no other route; no `float64` under `internal/`; every feature ships a
  golden fixture with a digest recorded across all four targets. The canvas measures through the same
  engine path as the PDF — the browser never measures text — so an embedded face must reach the
  preview measurement path too.

## UX & Interaction Patterns

- The family control is already a search-and-select over the engine-projected list of declared
  chains. It gains an affordance opening the chain editor **on the same panel** (no separate mode, no
  dialog stack) and the catalogue as a second, clearly separated group.
- **Declared chains and not-yet-embedded catalogue entries must be visibly distinct** — one is in the
  file, the other is not yet — and picking is what moves an entry between the groups.
- An entry naming an embedded face reads as that face's family and style from the projection, never
  as an asset key or a file name.
- Refusals (orphaning delete, empty chain, duplicate name) state the concrete reason in text at the
  control that caused it, following existing property-panel error, focus and accessible-name
  conventions. Voice is terse and technical: state the fact, name the location, no apology, no
  exclamation marks. Keyboard reach, visible focus and accessible names apply throughout.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** The PDF emitter currently refuses any glyph with a
  non-zero vertical offset and produces zero bytes. This epic lets an author embed arbitrary faces
  whose mark positioning is arbitrary, so 8.4's criterion — a template with embedded faces renders on
  a machine that has never seen them — is at risk from the first embedded face whose glyphs carry
  one. Shipping 8.4 over that fail-closed branch ships a feature that can newly stop documents
  rendering.
- **Story 7.10** (Epic 7's over-tall-element repair) is sequenced immediately after 8.0 and before
  8.1 by the owner's call. It does not gate Epic 7's completion but must land before the
  `folio-go/v0.1.0` tag.
- **8.1 → 8.2:** the panel reports what the engine answers, so every refusal 8.2 surfaces must exist
  as an engine-side refusal from 8.1.
- **8.3 → 8.4:** the engine can only render from an embedded face once the file carries one. 8.3 also
  owns settling and writing the version-bump decision before any of its code lands.
- **8.5 → 8.6:** picking requires the bundled catalogue. 8.6 also depends on 8.1's chain-editing
  commands (to propose and let the author edit the fallback tail) and on 8.3's content-addressed
  storage (so picking the same family twice stores one copy).
- **8.5 is the story to trim rather than cut** — the catalogue can ship with one family and grow by
  release.
