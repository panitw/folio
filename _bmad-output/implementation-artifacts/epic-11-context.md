# Epic 11 Context: Bold and italic mean what they say

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Bold and italic are the most prominent inert controls in the product: the toggles light up, the document
records them, and nothing on the render path reads them, so two documents differing only by bold and italic
render to identical bytes. This epic realizes them — the shipped families gain real weighted and sloped
faces, the engine resolves a face from the declared weight and slope and shapes and measures from that
face's own metrics, and the canvas paints the resolved face instead of letting the browser fake a weight
over Regular metrics. Page, canvas and panel finally agree on what bold means; where a family genuinely has
no such face, the product says so rather than pretending.

## Stories

- Story 11.1: The shipped families gain a weighted and a sloped face
- Story 11.2: The engine resolves a face from the declared weight and slope
- Story 11.3: The canvas paints the weight the engine resolved

## Requirements & Constraints

- Bold and italic become **real faces**, resolved per rune through the declared fallback chain. Synthetic
  emboldening and obliquing are forbidden everywhere — at emit time and on the canvas. A weight is a face or
  it does not exist.
- **Scope is three families, nine new faces**: Noto Sans, Noto Sans Thai and Roboto each gain bold, italic
  and bold-italic. The CJK family stays Regular-only — three instances of it would take the offline payload
  from roughly 11 MB to roughly 45 MB, where the Latin-and-Thai set costs about 3 MB. So "this family has no
  face at this weight" is a permanent shipped condition to design for, not a rare edge case.
- **Byte identity holds.** A document declaring neither bold nor italic hashes identically on every target,
  before and after. No golden-corpus document declares either, so the corpus is a real witness and must not
  be regenerated; a moved hash is a defect until proven intended.
- **A shipped face is committed output, never generated at build time**, from the same replayable
  derivation, reproducing byte-for-byte under verification.
- **Licence provenance is CI-enforced**: each face travels with its own notice (upstream release URL, source
  sha256) and OFL text, accounted for by the guard that fails the build on an unaccounted asset.
- **Payload cost is stated**: the release manifest carries the measured per-face byte cost and the load
  screen itemises the new faces as it already itemises the large CJK face.
- **A shortfall is stated, never silent**: a diagnostic names the element, the rune and the face, from the
  closed additive registry on the single diagnostic channel.
- A shipped user-facing string tells authors the families are Regular-only with no bold or italic.
  Realizing bold makes that false; it is corrected inside this epic.

## Technical Decisions

- **Weight and slope are declared on the font chain entry and nowhere else.** An entry gains optional
  style-variant siblings via the existing `Presence` idiom, absent by default — a face-name variant on the
  shipped arm, an asset-key variant on the embedded arm. The font-set value keeps its shape and **the public
  API does not change.** A chain entry already serialises as either a bare string or an object, so a
  document with no variants serialises exactly as today: corpus byte identity follows by construction.
- **Per-rune coverage decides the entry; style resolves within that entry only.** A covering entry with no
  face at the requested weight renders that rune in its own Regular with the stated diagnostic — never
  falling further down the chain to another entry's bold, which would change the typeface to keep the
  weight and make one script silently change family because another asked for bold. Absence is a
  first-class result the caller handles, not a nil defaulting to Regular, and the mixed-script proof belongs
  in the main test body, not an edge-case file.
- **Foreclosed, so no spec reopens them**: encoding weight into font-set key names; any shape change to the
  font-set type; synthetic bold or oblique in any form.
- Glyphs are shaped and measured from the resolved face's own metrics, so a bold run may break differently
  from the same words unbolded — and canvas and PDF break identically, both reading one engine measurement.
- **11.1 has two derivation mechanisms, not one.** The existing instancing script covers only the Noto
  families; Roboto is not derivable through it. The one-cut-per-name discipline guarding Roboto extends to
  the new cuts rather than being scoped to Regular.
- **The canvas is told the resolved outcome, not the requested flag** — a projection field, not a format
  field. The browser cannot discover a missing bold face without measuring or holding a second model of the
  font set, both barred.
- **The no-synthetic canvas contract test is an allowlist of permitted paint properties, not a denylist.** A
  denylist naming only weight and style cannot see `font-synthesis`, `-webkit-text-stroke`, `text-shadow`,
  `paint-order`, or a `transform: skewX()` — and a skew is how an oblique gets faked.
- **The B/I controls are three UI states over two document states.** Clearing writes absence, not an
  explicit false, so "off" and "never set" are the same bytes; the third state derives from the projection
  and is never stored. No document is normalised or migrated on load or selection.
- Table cells take weight and slope through the same cascade every other cell property uses, with the header
  style winning for the header row alone.

## UX & Interaction Patterns

- The load screen's manifest rows are where payload growth becomes visible: each new face is a named row
  with its measured size, not a silently larger download.
- The inspector's B/I pair must never show a state the document cannot reach. Where the resolved family has
  no bold face, the control states that reason rather than appearing to be on — anything disabled carries a
  stated reason.
- The canvas-approximate / preview-exact asymmetry must stay legible, and this epic narrows the gap:
  painting the real resolved face stops the preview holding a surprise the canvas could have shown.

## Cross-Story Dependencies

- Strictly ordered. Nothing for 11.2 to resolve until 11.1's faces ship; nothing for 11.3 to paint until
  11.2 resolves a face and projects the outcome, which 11.3 consumes rather than re-deriving in the browser.
- 11.2 rests on an assumption to verify at its plan gate: that the designer can write chain style variants
  through the existing font-chain command family without a new command kind. If not, 11.2 and 11.3 both
  acquire a command-surface change and a dependency on the command-layer work.
- Inherited constraints from shipped work: the current shipped face set and its machine-checked
  byte-identity rule; the toggle semantics where pressing a pressed control clears rather than writing
  false; and the repo-wide scan forbidding browser measurement, whose narrow named exceptions must not widen.
