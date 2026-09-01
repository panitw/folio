# Epic 8 Context: A template author can choose a font, and the file carries it

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Today a document can only name the three faces the library ships at build time, and nothing in the
designer edits the font-chain map at all — so every document created there offers exactly one
family inherited from the starter file. This epic makes the typeface a real authoring choice and
makes the `.folio` file carry it: chains become editable in the designer, a face's bytes ride inside
the template as a content-addressed asset, the engine renders from those bytes on a machine with no
fonts installed and no network, the canvas rasterizes with the face the engine measured, and a
curated freely-licensed catalogue makes picking a font a search rather than a hunt for a licence and
a file. The payoff is that the `.folio` stays the whole contract between author and integrator.

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

- **Authorable chains.** Create, rename, delete chains and reorder entries within one, from the
  designer. "Reorder" is entry-level: a chain is an ordered list, while the chain map has no
  authored key order and must not gain one.
- **Faces embedded in the file**, keyed by content hash like every other asset and referenced from a
  chain entry. Stored once when shared; dropped on save when no chain names them.
- **Render from embedded faces alone** — no network, no host-installed font, no path on disk, no
  install step on the rendering machine.
- **Curated offline catalogue** shipping with the designer. Licences restricted to a permissive
  allowlist; the build *fails* (never warns) on anything outside it or on a licence text the
  classifier cannot identify. CJK families excluded this epic on size grounds; the shipped CJK face
  stays the coverage fallback.
- **Located failure for a broken reference** — a chain naming a font that is neither shipped nor a
  present, decodable asset errors naming the chain, entry index and key; never a substituted face,
  never a silent drop.
- **Byte identity is non-negotiable.** Every added field is optional and absent-by-default, so the
  existing golden corpus must hash identically on all four targets. New fixtures carry recorded
  digests and must actually exercise the embedded face they carry.
- **Bounds and gates fail loudly.** First-load weight is a declared figure a build gate checks and
  fails on, set deliberately once against the epic's finished weight — never rewritten to match what
  the build happens to weigh. Same for the asset-count bound and unlicensed font files. An all-clear
  must be distinguishable from a couldn't-look.

## Technical Decisions

- **The format stays one JSON text file.** Font bytes ride the existing content-addressed asset map
  — the mechanism images use — with a font media type and a record carrying family, style, licence
  and source. No second storage shape, no container format, no new canonical-serialization rule.
  Sorted keys, two-space indent, LF, trailing newline and round-trip byte identity are unchanged.
- **Media types are never a closed set.** An unrecognised media type is *preserved* at load and
  errors at *render*; a recognised type whose bytes fail to decode is a *load* error. Two distinct
  outcomes that must not be collapsed. `Validate` must predict what render would do.
- **Version bump**: the entry-shape change (an object where a 1.x reader expects a string) is the
  trigger, and it joins the existing `2.0` MAJOR rank rather than adding a new one. Written into the
  format doc before code lands.
- **Resolution is by asset key, never by name.** An embedded face joins the same per-rune coverage
  resolution as a shipped one, in declared chain order. Family/style are display identity only.
  Consequently an embedded face's browser CSS family name derives from its **asset key**, so an
  embedded "Inter" cannot collide with a shipped "Inter" in the font registry.
- **The engine owns the document.** No TypeScript model of the chain map. Every edit travels as one
  opaque command producing one history entry; the designer re-projects from the engine's answer. A
  rename updates the map and every element naming it inside that one command, so one undo restores
  both.
- **The browser never measures** — every metric and line break comes from the engine's measure API,
  including on the canvas; the browser contributes rasterization only. Which is exactly why the
  canvas must be able to name and register the face the engine measured with.
- **Two font-family attachment points, not one** — an element's style and a table's header style.
  Both are live; rename, delete and orphan checks must read on both.
- **Subsetting stays once per render inside the PDF producer**; no face is subset at save time. The
  subset tag remains a deterministic hash of the emitted subset program's own bytes, not the glyph
  set.
- **Diagnostics are one type on one channel**, stable codes from a closed registry, carrying element
  id and/or data path.
- **Licence provenance.** Redistributed faces travel with licence text and copyright lines and
  appear in the release licence manifest; the licence gate must be able to *see* every font file
  reaching the bundle, whatever its extension or procurement route.
- **The build itself must be deterministic** — a function of the source, not of the working tree —
  or recorded byte figures mean nothing.
- **Guards are widened or replaced, never weakened.** Several existing tests were written to forbid
  exactly the shapes this epic introduces; each must be re-authored to assert the new invariant and
  red-proved. Intermediate states are pinned by an assertion, not defended by a comment.

## UX & Interaction Patterns

- The chain editor lives on the typography panel where a font is chosen — no separate mode, no
  dialog stack. Choosing a font and defining what a font *is* are one tool.
- A refused edit (orphaning delete, empty chain, duplicate name) states its concrete reason in text
  at the control that caused it, following existing property-panel error, focus and accessible-name
  conventions. Voice is terse and technical: state the fact, name the location, offer no comfort.
- In the family control, catalogue entries the document already declares and those it does not are
  visibly distinct; picking moves an entry from the second group to the first.
- Picking a family is one undoable step that both embeds the face and declares the chain. Where the
  picked face does not cover every script the document may render, the proposed chain tail is the
  shipped faces for the uncovered scripts, editable by the author.
- Accessibility floor applies: keyboard reach and operation, visible focus, accessible names on
  icon-only controls, errors distinguished by shape before colour.

## Cross-Story Dependencies

- **8.0 is a precondition of 8.4, not a preface.** Epic 7's Story 7.10 is sequenced immediately
  after 8.0 and before 8.1 by owner call; it does not gate Epic 7 but must land before the
  `folio-go/v0.1.0` tag.
- **8.1 → 8.2**: the engine must make a refusal before the panel can report it. 8.2 delivers only
  the negative half of its entry-display criterion, because the embedded-face entry shape does not
  exist until 8.3 — which must extend that display and move the entry-shape validator in one commit.
- **8.3 → 8.4**: a face must travel in the file before the engine can render from it.
- **8.4 → 8.4a**: 8.4 delivers the measurement half; 8.4a the rasterization half and the
  asset-key-derived CSS family name.
- **8.4a / 8.4b / 8.4e** close canvas fidelity in three parts: the carried-face cause, the
  shipped-face vocabulary, and per-fragment face attribution for shipped faces. 8.4e closes it out.
- **8.4b → 8.4c**: 8.4b creates a deliberate state where two family names resolve to one file, which
  must be asserted; 8.4c makes them diverge by shipping the real specified typeface for the chrome.
- **8.4g before 8.4f, 8.5 and 8.4d**: byte figures are untrustworthy until the build depends on the
  source alone.
- **8.4f before 8.5**: the asset-count bound must fail loudly before the catalogue pushes against
  it; raising that bound is legitimate only afterwards, and only with stated headroom.
- **8.5 → 8.4d**: 8.5 computes per-asset compressed weights that 8.4d consumes rather than
  re-measuring. **8.5 → 8.6**: the catalogue must exist before picking from it can write a chain and
  embed a face; 8.6 also makes font assets visible to the asset-reference walk.
- **8.4d is last in the epic** — the threshold is set once, against the finished weight, after every
  story that adds to the bundle has landed.
