# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

A template author picks a typeface from a catalogue inside the designer, lays out the document, and
saves. An integrating Go developer renders that same file on a machine with no fonts installed and
no network, and gets the identical PDF, byte for byte. This epic makes the faces a document uses
declarable in the file, choosable in the designer, and carried inside the `.folio` — because the
file is the whole contract between the two of them, and a font nobody can install is not a choice
an author can make. It then closes the fidelity debts that creates: the canvas must rasterize with
the face the engine measured with, the shipped chrome typeface must be the one the design system
actually specifies, and the bundle-size budget must become a number something checks.

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
- Story 8.5: A curated catalogue ships with the designer
- Story 8.6: Picking a family puts it in the file
- Story 8.4d: The size budget is a number something checks (sequenced last)

## Requirements & Constraints

- **Chain authoring.** Create, rename, delete the document's font chains and reorder entries within
  a chain, so `fontFamily` names a family the author chose. Chain *key* order deliberately does not
  exist — `fonts` is a mapping with no authored key order, and that absence is load-bearing.
- **Embedding.** A face lives inside the template, keyed by content hash like every other asset, and
  is referenced from a chain entry. A template renders from its embedded faces alone: no network, no
  host-installed font, no install step on the rendering machine.
- **Catalogue.** A curated, freely-licensed catalogue ships with the designer and works fully
  offline — no request leaves the machine during the pick flow. CJK is excluded from the catalogue
  this epic (a full SC face dwarfs the Latin and Thai faces); the shipped SC face stays the fallback.
- **Failure behaviour.** A chain naming a font that is neither a shipped face nor a present,
  decodable asset fails with a located error. An unrecognised media type is *preserved* at load and
  errors at render; a recognised type whose bytes do not decode is a load error. `Validate` predicts
  what `Render` would do. There is no closed set of font media types, and there must not be.
- **Determinism is the product.** Existing golden fixtures keep their recorded digests on all four
  targets unless a story explicitly owns re-recording one. Subsetting stays byte-stable and happens
  once per render inside the PDF producer, never at save time. Every new capability ships a fixture
  whose document actually exercises it.
- **Licensing and weight.** Redistributed binaries keep their licence text and notices, and the
  release licence manifest must be *true*. The licence gate must see any font file reaching the
  bundle, whatever its extension. Added weight is measured and recorded; the budget ends the epic as
  a declared, enforced figure chosen deliberately — never rewritten to match the current build.

## Technical Decisions

- **Fonts are an explicit value to the engine.** Nothing under `internal/` embeds font data;
  resolution inside a render is a pure lookup, never a host query and never name-based
  substitution. **The asset key decides which face is used**, even when an embedded and a shipped
  face share a family name; declared family/style is display identity only. A rune covered by no
  face in the chain is a located diagnostic, never a blank box.
- **One canonical byte form.** Embedded faces reuse the existing `assets` mechanism — lowercase hex
  SHA-256 key, base64 hard-wrapped at 76 columns, a media type — with no second storage shape and no
  new serialization rule. Faces deduplicate, and adding a font must not move an image's emission
  position. The version bump is triggered by the entry *shape* (an object where a string was legal)
  and joins the existing MAJOR rank; the format doc is updated before code lands.
- **The engine owns the document in the designer.** No TypeScript model. Every mutation is one
  opaque command over one channel, and the UI re-projects from the engine's answer. A rename updates
  the map and every element naming it in one command and one history entry. `fontFamily` has
  **two** attachment points — element style and table header style — and every rule reads on both.
- **The browser never measures text**, on canvas as in the PDF path; it contributes rasterization
  only. The corollary this epic must satisfy: the canvas cannot rasterize faithfully unless it can
  *name* the face the engine measured with, so face identity travels **per fragment** rather than
  through a fixed stylesheet stack. An embedded face's CSS family name derives from its **asset
  key**, never from its declared family, or an embedded and a shipped face collide in the browser's
  font registry.
- **Numbers and profile.** Every number reaching an output byte goes through the single
  number-emission path in fixed-point units, with no `float64` under `internal/`. Subset tags hash
  the emitted subset program's own bytes, not the glyph-id set.
- **Errors are one type on one channel** — severity, stable code, optional element id and data path,
  message. Chain load errors must locate to the chain name *and entry index*.
- **Guards get widened, never weakened.** Several existing tests were written to forbid exactly the
  shapes this epic introduces; each is re-authored to assert the new rule, and the old state it
  recorded should visibly go red. Where an intermediate state is deliberate, assert it and name the
  successor story rather than defending it in a comment.

## UX & Interaction Patterns

- The chain editor lives on the typography property panel, opened from the family control — no
  separate mode, no dialog stack, so choosing a font and defining what a font *is* are one tool.
- A refused edit states its concrete reason in text at the control that caused it, following the
  existing property-panel error, focus and accessible-name conventions. Voice is terse and
  technical: state the fact, name the location, offer no comfort.
- In the family control, entries the document already declares and entries it does not are visibly
  distinct; picking is what moves an entry between the groups, as one undoable step.
- Controls stay keyboard-reachable with visible focus and accessible names; errors read by shape
  before colour. Each catalogue entry's weight is visible to whoever curates the list.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4**, not a preface: it removes a fail-closed branch refusing any
  glyph with a vertical offset, which arbitrary embedded faces would reach far more often. Story
  **7.10** (Epic 7's over-tall-element repair) is sequenced immediately after 8.0, before 8.1.
- **8.1 before 8.2** — the panel reports what the engine answers, so refusals (duplicate name,
  orphaning delete, empty chain) must exist in the engine first.
- **8.3 completes 8.2.** 8.2 delivers only the negative display rule (an entry renders as the
  projected string, unmodified); 8.3 introduces the embedded-entry shape and must make entries read
  as family and style, moving the projection's entry-shape validator in the same commit.
- **8.3 before 8.4** — storage and load precede rendering from a carried face.
- **8.4 → 8.4a → 8.4b → 8.4c → 8.4e.** 8.4 ships the engine half and discloses the canvas gap as a
  test; 8.4a closes it for carried faces; 8.4b gives the canvas the engine's face vocabulary; 8.4c
  ships real chrome binaries, with the licence-gate extension and mono-face fix landing in its first
  commit before any binary arrives; 8.4e extends per-fragment attribution to shipped faces and
  closes the canvas-fidelity debt.
- **8.5 before 8.6** — the catalogue must exist and be offline-reachable before picking can write a
  chain. 8.6 also owns dropping font assets no chain names, closing a gap where font assets were
  invisible to the asset-reference sweep.
- **8.4d is last.** Its declared figure is the epic's finished weight, so it cannot be written until
  8.4c and 8.5 have landed. Its gating first task is reconciling conflicting first-load
  measurements; if the build proves nondeterministic, that finding outranks the budget gate.
