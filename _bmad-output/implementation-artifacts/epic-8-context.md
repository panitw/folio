# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Today the engine ships three Regular-only faces embedded at build time, and a document's `fonts`
map declares named chains over those face names. Nothing in the designer edits that map, and the
starter document declares exactly one chain — so every template made in the designer offers a font
list of length one. This epic makes the faces a document uses **declared in it, chosen in the
designer, and carried in the file**: the author picks a family from a curated, bundled catalogue,
the face's bytes land in the `.folio`, and an integrating Go developer renders that file on a clean
box with no fonts installed and no network and gets byte-identical output. The `.folio` is the whole
contract between those two people, so a font nobody can install is not a choice the author can make.
The epic opens with a precondition fix: a glyph carrying a vertical offset currently refuses to
render at all, which arbitrary embedded faces would hit far more often than the shipped set does.

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
  exists so a person or an agent can hand-edit a template and so a hand-written template renders.
- **Font bytes reuse the existing `assets` map**: content-addressed by lowercase hex SHA-256 of the
  raw bytes, `data` base64 hard-wrapped at 76 columns, deduplicated by key, emitted in stable key
  order. No second storage shape, no new canonical-serialization rule, and adding a font must never
  move an image in emission order.
- **A chain entry is either a shipped face name (JSON string) or an embedded face
  (`{"asset": "<key>"}`).** Any other shape, an absent key, an out-of-set media type, or bytes that
  do not decode as their declared media type is a **located load error** naming the chain, index and
  key — never a substituted face, never a silent drop. Chain semantics are otherwise unchanged:
  ordered, resolved per rune for coverage, first covering entry wins, a chain may mix both shapes,
  and an empty chain is not something `fontFamily` may name.
- **Nothing is fetched or read at render time** — no network, no host-installed font, no path on
  disk. A document renders from its own bytes plus the supplied shipped set. Font resolution is a
  pure lookup by asset key; family names never resolve or substitute a face.
- **The authoring path is offline too.** The catalogue and its face bytes ship in the offline
  release bundle behind the same verified asset URLs as every other release asset, covered by the
  offline verification job. No `fonts.google.com`, no `fonts.gstatic.com`, no first-use download.
- **Embedded faces are stored whole.** No save-time subsetting — the data a template renders changes
  between saves, so a subset chosen at save time drops glyphs a later render needs. Subsetting stays
  once per font per document, inside the PDF producer, at render time, with subset tags a
  deterministic hash of the glyph set and no wall-clock timestamp in the output.
- **Every embedded face carries a licence record** (family, style, licence identifier, source
  naming the upstream release and instance), so a file that travels alone states what it
  redistributes.
- **Catalogue faces are static, single-instance, OFL-licensed**, derived ahead of the build by the
  same replayable derivation the shipped set uses, with committed outputs and a NOTICE per face
  recording upstream release and hashes. Never generated at build time, which would make the PDF a
  function of the build environment. CJK families are excluded from the embeddable catalogue in this
  scope (a full SC face is 10.6 MB against 646 KB Latin / 47 KB Thai); the shipped SC face remains
  the coverage fallback. Added bundle weight is measured against the size budget and each entry's
  own weight is visible to whoever curates.
- **Byte identity is unchanged.** All four CI targets still produce identical bytes, a no-op round
  trip still rewrites the file byte-for-byte, and the existing golden corpus — none of which embeds
  a font — keeps its recorded digests.
- **The `.folio` version bump** (MINOR vs MAJOR for the additive font fields) must be settled and
  written into the format contract before any Story 8.3 code lands; the version rule makes a higher
  MAJOR a load error rather than a best-effort render.
- **Explicit non-goals:** no container format, no live font service or arbitrary URL, no host-font
  enumeration, no synthetic bold or oblique, no variable-font axes, no save-time subsetting.

## Technical Decisions

- **No package under `internal/` embeds font data.** `Render` takes a `FontSet` value; font
  binaries live in exactly one directory inside the Go module, wrapped with `go:embed` for native
  callers and copied from the same place by the designer build. An embedded face and a shipped face
  reach the producer identically — a `FontSet` entry is a `FontSet` entry. A rune covered by no
  entry in the chain is a diagnostic carrying the element id and the offending rune, never a blank
  box.
- **In the designer there is no TypeScript model of the document.** The wasm engine parses, holds,
  mutates, validates and serializes it; the UI holds an immutable snapshot for painting and sends
  every committed mutation as one opaque command over one channel. Every chain, entry and family
  value the panel shows comes back from the engine's answer. Undo/redo is engine-side history over
  committed commands — a multi-part edit (a rename that touches elements, a catalogue pick that both
  embeds and declares) is **one command and one history entry**.
- **The browser never measures text**, including on the canvas: every metric and line break comes
  from the engine's measure API, so an embedded face is measured through the same path for the
  canvas preview and the PDF — one measurement authority.
- **The page model knows nothing about PDF**, which is why the canvas can draw text the emitter
  refuses. Emission-stage gaps are invisible to the canvas by design; that is an invariant, not a
  defect to route around.
- **Numbers reach output bytes through exactly one file** in the PDF package, in exactly two
  representations (a thousandths-scaled decimal emitter and a plain integer writer). Anything new
  written into a content stream — including a text rise — goes through those emitters, in
  `geom.Length` millipoints, with `float64` appearing nowhere under `internal/`.
- **The PDF profile is pinned** and text-state operators are inside it; the profile's exclusion list
  covers encryption, annotations, forms, transparency groups, shading, ICC and tagging.
- **Errors and diagnostics are one type on one channel**: a stable code from a closed registry, an
  optional element id and data path, and a message. Callers match on the code, never on message
  text. Codes are additive only.
- **Every feature ships its golden fixture**, and hashes are compared across `darwin/arm64`,
  `linux/amd64`, `linux/arm64` and `js/wasm` under Node. A hash change is investigated as a defect
  until proven an intended, versioned change. Any change to layout, subsetting, emission or the
  toolchain is a breaking change for downstream test suites and is released as one; the toolchain is
  pinned and bumping it is a release event.
- **Licence boundary is MIT and nothing copyleft enters**, enforced by a CI licence check over the
  whole module graph and the designer lockfile. Redistributed non-code assets keep their own terms
  and travel with their licence text; a third-party licence manifest is a release artifact.
- **Conventions that bind here:** `lowerCamelCase` JSON keys; asset keys lowercase hex SHA-256;
  diagnostic codes `SCREAMING_SNAKE`; nothing under `internal/` panics on malformed input —
  untrusted font and template bytes return diagnostics.

## UX & Interaction Patterns

- The chain editor lives **on the typography panel**, reached from the family control — no separate
  mode, no dialog stack. Choosing a font and defining what a font *is* are one tool.
- The family control is a search-and-select over the engine-projected chains, with the catalogue as
  a second, **visibly distinct** group: a chain is in the document, a catalogue entry is not yet, and
  picking is what moves an entry from one group to the other.
- A chain entry naming an embedded face reads as its family and style from the projection, never as
  an asset key or a file name.
- A refused edit states the concrete reason **in text at the control that caused it**, following the
  existing property-panel error, focus and accessible-name conventions.
- Diagnostic voice is terse and technical: state the fact, name the location, offer no comfort.
- Accessibility floor binds: every control keyboard-reachable and operable, visible focus on
  everything focusable, accessible names on icon-only controls, diagnostics distinguished by shape
  before colour.
- Undo covers every property mutation; a catalogue pick and a chain rename must each undo whole.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** Arbitrary embedded faces have arbitrary mark
  positioning, so 8.4's acceptance — a template with embedded faces renders on a machine that has
  never seen them — is at risk while the emitter fail-closes on any glyph with a vertical offset.
  8.0 must land first. Its characterization test already exists and is mutation-proved; it becomes
  8.0's before/after and is re-pointed, never deleted.
- **8.1 before 8.2:** the engine must own chain edits as commands before the panel can project and
  command them.
- **8.3 before 8.4 and 8.6:** the storage shape, the load-error surface and the settled version bump
  must exist before rendering from an embedded face or before a catalogue pick can write one.
- **8.5 before 8.6:** the bundled catalogue and its faces must ship before picking a family can
  embed bytes already on the machine.
- **8.6 depends on 8.1's chain editor** for the editable fallback tail it proposes.
- **Story 7.10** (Epic 7's over-tall-element repair) is sequenced immediately after 8.0 and before
  8.1 by owner decision. It does not gate Epic 7's completion, but it must land before the
  `folio-go/v0.1.0` tag.
- **8.0 is deliberately outside the before-the-tag set:** emitting for glyphs that currently refuse
  changes no existing golden by construction, since those documents produce zero bytes today. It
  widens what renders rather than narrowing it.
