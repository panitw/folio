# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks a typeface from a catalogue inside the designer, lays out the document, and
saves; the integrating Go developer renders that file on a box with no fonts installed and no
network, and gets the same PDF byte for byte. Today the engine ships three Regular-only faces,
nothing in the designer edits the document's `fonts` map, and the starter file declares one chain —
so the author's font list has length one and any other answer means hand-editing the file. This
epic makes the faces a document uses declared in it, chosen in the designer, and carried in the
file, because the `.folio` is the whole contract between the two users and a font nobody can install
is not a choice the author can make. It also removes a fail-closed emission branch that would
otherwise turn arbitrary embedded faces into documents that stop rendering.

## Stories

- Story 8.0: A stacked Thai mark reaches the page
- Story 8.1: The document's font chains become editable
- Story 8.2: The chain editor sits where fonts are chosen
- Story 8.3: A font travels inside the template
- Story 8.4: The engine renders from an embedded face
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file

## Requirements & Constraints

- **Chains are authorable.** Create, rename, delete chains and reorder the entries within one, from
  the designer. Chain-*key* order is out of scope — `fonts` is a mapping with no authored key order.
- **Faces travel in the file.** A face is embedded keyed by content hash like every other asset and
  referenced from a chain entry; a template renders from its own bytes plus the shipped set alone —
  no network, no host-installed font, no path on disk, no install step.
- **Broken references fail where they are written.** An entry naming neither a shipped face nor a
  present, decodable font asset is a located load error naming the chain, index and key — never a
  substituted face, never a silent drop.
- **The authoring path is offline too.** Catalogue and faces ship inside the offline release bundle
  behind the same verified asset URLs as every other release asset; no request leaves the machine
  during search-and-pick. Added bundle weight is measured and recorded against the size budget.
- **Byte identity is the standing gate.** Every existing golden fixture renders byte-identically on
  all four targets after every story here, and a no-op round trip still rewrites the file byte for
  byte. Each new capability ships its own fixture with a recorded digest.
- **Licensing is a hard boundary.** Only redistribution-permitting (OFL) families enter the
  catalogue; every embedded face carries a licence record so a file travelling alone states what it
  redistributes. Nothing copyleft enters the dependency graph.
- **Out of scope:** any container format, a live font service or download-on-first-use, host-font
  enumeration, synthetic bold/oblique or variable axes, save-time subsetting, and CJK families in
  the embeddable catalogue (CJK stays on the shipped-face path).

## Technical Decisions

- **`.folio` stays a single JSON text file.** Font bytes reuse the existing `assets` map:
  content-addressed by lowercase hex SHA-256, `data` base64 hard-wrapped at 76 columns, deduplicated
  by key, emitted in stable key order. No second storage shape, no new canonical-serialization rule,
  and adding a font never moves an image.
- **Chain entry shapes.** A JSON string is a shipped face name; a one-key object
  `{"asset": "<key>"}` is an embedded face; any other shape is a load error naming chain and index.
  Chains may mix both. Per-rune coverage resolution and first-covering-entry-wins are unchanged, an
  empty chain is still not nameable, and `style.fontFamily` still names a chain, never a face.
- **Font asset record.** `mediaType` from a closed set (`font/ttf` at minimum; outside the set, or
  bytes that do not decode as the declared type, is a load error), plus a `font` record carrying
  family, style, licence and source. Family and style are display identity only — resolution is by
  asset key alone, so an embedded and a shipped face sharing a family name never substitute. Font
  and image assets are told apart by `mediaType`, never by inspecting bytes.
- **Version bump.** The additions are additive, but the version rule is a load-time error path
  (higher MAJOR is a load error, not a best-effort render). MINOR-vs-MAJOR must be settled and
  written into the format doc before any embedding code lands.
- **Faces are stored whole.** No save-time subsetting — the data a template renders changes between
  saves. Subsetting stays once per render inside the PDF producer, subset tags a deterministic hash
  of the glyph set.
- **The engine owns the document.** No TypeScript model of the `fonts` map. Every chain and entry
  edit is one opaque command with one history entry; the designer re-projects from the engine's
  answer. `fontFamily` has exactly two attachment points — an element's style and a table's header
  style — and every rename, delete and orphan check must read on both.
- **Fonts are an explicit value.** No package under `internal/` embeds font data; `Render` takes a
  `FontSet` and resolution is a pure lookup against it. Catalogue faces are derived ahead of the
  build by the same replayable derivation the shipped set uses, with committed output, licence text
  and a NOTICE recording upstream release and hashes — never generated at build time.
- **Determinism.** Every number reaching an output byte goes through the single number emitter, in
  fixed-point millipoints, no `float64` under `internal/`, no unordered iteration into emission.
- **One measurement authority.** The canvas measures embedded faces through the same engine path the
  PDF uses; the browser never measures text.

## UX & Interaction Patterns

- The chain editor lives on the typography panel, reached from the family control by an affordance
  on that same panel — no separate mode, no dialog stack.
- A chain entry displays what the engine projects, never an asset key or a parsed file name.
- Refusals (orphaning delete, empty chain, duplicate name) are stated as concrete text at the
  control that caused them, following existing property-panel error, focus and accessible-name
  conventions, in the product's terse, location-naming voice.
- Picking a catalogue family is a single act: one command embeds the face and declares the chain,
  one undo removes both. Families already in the document are visibly distinct from those not yet
  embedded, and a picked face that does not cover every script gets an editable shipped-face tail.
- Accessibility floor applies: keyboard reach and operation, visible focus, accessible names on
  icon-only controls, errors distinguished by shape before colour.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** It removes the fail-closed refusal of glyphs
  carrying a non-zero vertical offset. Embedding arbitrary faces widens that branch's reach, so
  8.4's promise is at risk until 8.0 lands. Its characterization test already exists and becomes the
  before/after arms — re-pointed, never deleted.
- **Story 7.10 (Epic 7's over-tall-element repair) is sequenced between 8.0 and 8.1** by owner call.
  It does not gate Epic 7 being done, but must land before the `folio-go/v0.1.0` tag.
- **8.2 depends on 8.3 for half of its entry-display criterion.** No font-record projection exists
  yet, so 8.2 delivers only the negative half (entry shown unmodified); 8.3 must extend it to read
  as family and style and move the projection's entry-shape validator in the same commit.
- **8.1 before 8.2:** the panel can only report refusals the engine actually makes, so the
  duplicate-name, orphaning-delete and empty-chain refusals belong to the engine story.
- **8.3 → 8.4 → 8.6:** the format must carry a face before the engine can render from one, and both
  before a catalogue pick can put one in the file. 8.5 supplies the catalogue 8.6 picks from.
- 8.6 also depends on save-time dropping of font assets no chain names any longer.
