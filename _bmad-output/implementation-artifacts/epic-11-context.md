# Epic 11 Context: Bold and italic mean what they say

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Bold and italic are the most prominent inert controls in the product: the toggles light up, the
document records them, the canvas fakes a thicker stroke, and nothing on the render path reads
them — two documents differing only by bold and italic render to identical bytes. The epic's
governing ruling was to **realize** them rather than retire them. So the shipped families gain real
weighted and sloped faces; the engine resolves a face from the declared weight and slope and shapes
and measures from that face's own metrics; the canvas paints the face the engine resolved instead of
letting the browser embolden Regular; the family control writes the cuts a family has into the
document, so bold is reachable in a file the author did not hand-edit; and a bold document joins the
pinned corpus, so a face that silently changes is caught by a golden rather than by a reader. Where
a family genuinely has no such cut, the product states the absence rather than showing a state the
document cannot reach.

## Stories

- Story 11.1: The shipped families gain a weighted and a sloped face
- Story 11.2: The engine resolves a face from the declared weight and slope
- Story 11.3: The canvas paints the weight the engine resolved
- Story 11.4: Picking a family declares the cuts it has
- Story 11.5: A bold document is a pinned golden

## Requirements & Constraints

- **A weight is a face or it does not exist.** Synthetic bold and synthetic oblique are forbidden at
  emit time *and* in the browser, as are variable-font axes. Faking a weight would put a fabricated
  outline inside the byte-identity regime. Story 11.3 must *remove* the browser's faking, not tidy it.
- **A shipped face is committed output, never a function of the build environment.** Each new face is
  derived by the same replayable derivation, committed rather than generated at build time, and the
  verification target reproduces every committed face byte-for-byte.
- **Licence provenance is a build gate, not a warning.** Every face travels with its own notice
  recording the upstream release URL and source sha256, plus its OFL text, and is accounted for by
  the font-asset guard; an unaccounted or unlicensed asset fails the build.
- **Payload growth is stated where the author waits for it.** The offline release manifest carries
  the measured per-face byte cost and the load screen itemises each face as a named row, the way the
  large CJK face is already itemised. First-load weight is a considered price, not a silent one.
- **Byte identity is a constraint on this epic, not a target of it.** A document declaring neither
  bold nor italic must hash identically across all four targets, before and after every story. The
  inherited corpus declares bold or italic nowhere, so it is a genuine witness that nothing moved; a
  moved hash is investigated as a defect until proven an intended, versioned change.
- **The epic must leave a bold document pinned.** No inherited fixture declares bold, so a resolver
  that silently switched face would fire no diagnostic, red no test and move no golden. The hole is
  closed by a *new* fixture declaring bold and italic on a chain that declares its variants, rendered
  identically on all four targets and pinned by a JSON and a PDF golden registered in the golden
  digest record and CI's matrix slug list. The one existing in-tree document that declares bold is
  byte-compared against the format documentation's worked example, so the two move together and it
  cannot be adapted into a fixture. A pinned PDF is a human-attested artifact: attestation is an owner
  action and the work halts for it, never self-attesting and never shipping the fixture goldenless.
- **A shortfall is stated, never silent.** A covered rune with no face at the requested weight renders
  in the nearest available face and raises a diagnostic naming the element, the rune and the face —
  one diagnostic type on one channel, from the closed, additive code registry.
- **Not every family has every cut.** An absent cut is a permanent shipped condition to design for,
  declared absent rather than declared and empty.

## Technical Decisions

- **Fonts are an explicit value.** The engine takes a font set; nothing under the internal packages
  embeds font data, the binaries live in exactly one directory inside the Go module, and resolution
  during a render is a pure lookup against the supplied set — never a host font query, never a fetch,
  never a path on disk.
- **Chain semantics are unchanged.** A template's `fontFamily` names an ordered chain, resolved per
  rune for coverage, first covering entry wins; an entry may name a shipped face or an embedded asset
  and a chain may mix both. The chain is part of the font set's identity, so a different chain is a
  different render, never a silent substitution.
- **The engine resolves a weight only to a face the document explicitly names, and never infers one.**
  This is the ruling that keeps font resolution an explicit value, and it is precisely why Story 11.4
  exists: without the family control writing the available variants into the chain entry, only the
  starter document could ever bold.
- **11.4 adds no new command kind.** The family control writes variants through the same command
  family it already uses; the engine infers nothing from what it is handed.
- **No migration, ever.** Documents whose chain entries predate 11.4 are not rewritten, repaired or
  normalised on load or on selection; they acquire variants only when their author re-picks the
  family. A panel narrowing or widening its vocabulary is never a document migration.
- **The browser never measures.** The canvas paints pre-broken lines with browser wrapping disabled
  and takes every metric and break from the engine's measure API. Because a bold run is measured from
  the bold face's own metrics, it may wrap differently from the same words unbolded — and canvas and
  PDF must show that difference identically, since both read one engine measurement.
- **The canvas is given the resolved outcome through the engine's own text paint projection**, not the
  requested flag; the no-synthetic rule is enforced by a contract test that names forbidden CSS the
  way the canvas-authority contract already names the banned measurement APIs.
- **The engine owns the document.** There is no TypeScript model of a `.folio`; the UI holds an
  immutable snapshot for painting and sends every committed mutation as a command.
- Table cells take weight and slope through the same cascade every other cell property uses, with the
  header style winning for the header row alone.
- Subsetting and embedding are untouched: one subset per font per document, at render time, inside the
  PDF producer; a shipped face and an embedded face reach that path identically. Layout, subsetting,
  emission and toolchain changes are breaking changes for downstream test suites.

## UX & Interaction Patterns

- The load screen's payload manifest is where new bytes become visible: each face is a named row with
  its measured size, never an unexplained larger download.
- The inspector's TYPOGRAPHY B / I pair must never appear to be on in a state the document cannot
  reach. Where the resolved family has no bold or italic cut, the control states that fact — the
  standing rule is that anything declined or disabled carries its reason beside it, in the product's
  terse, technical, located voice.
- The family control uses the same words for an absent cut that the B / I controls use, so one absence
  is described one way wherever the author meets it.
- The canvas-approximate / preview-exact asymmetry must stay legible without a tutorial; this epic
  narrows the gap by making the canvas paint the real face, so the preview holds no surprise the
  canvas could have shown.

## Cross-Story Dependencies

- The stories are ordered, but they are not one chain. 11.2 has nothing to resolve until 11.1's faces
  ship; 11.3 has nothing real to paint until 11.2 resolves a face and projects the outcome, which the
  canvas consumes rather than re-deriving; 11.4 depends on 11.3's wording for a stated absence and on
  11.2's resolution being explicit-only. 11.5 depends on 11.2 alone — it pins rendered bytes and does
  not touch the canvas.
- 11.5 was split out of 11.3 so the four-target render matrix falls on the one story that moves a
  rendered byte: 11.3 is otherwise entirely designer-side and runs no matrix, while 11.5 runs the
  matrix suite unfiltered and carries the owner attestation halt.
- 11.4 was added by owner decision after 11.2's no-inference ruling; treat it as the story that makes
  bold reachable in existing documents, not as optional polish.
- The epic builds on the shipped-font and font-chain authoring work already in the product: the chain
  entry shape, the designer's chain-editing commands, the offline release payload and its manifest,
  and the licence-accounting lint. Those are inherited constraints to satisfy, not surfaces to redesign.
