# Epic 15 Context: Folio can be released

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

This epic ships no product feature. It closes the process and integrity defects between a repository
and a shippable product: a golden digest that moved unexplained, a CI signal red by design and
therefore unreadable, a guard that reads only part of the tree, a command encoding that can name one
thing and change another, a first-load payload carrying thirty-one typefaces most users never pick,
and the absence of any cut release. Byte identity across the four targets still holds, so the
determinism property is intact; what has failed is the discipline that makes it verifiable, the
enforcement that makes a command's guarantee real, and the process that lets a third party depend on
either.

## Stories

- Story 15.0: A catalogue face arrives when it is picked
- Story 15.1: The golden report's moved hash is explained
- Story 15.2: CI's red means something
- Story 15.2b: The font-host scan sees the whole tree
- Story 15.2a: A component command means exactly what it names
- Story 15.3: `folio-go/v0.1.0` is cut

## Requirements & Constraints

- **Byte-reproducible rendering is the signature requirement.** Identical inputs produce identical
  PDF bytes on every target at every page count; golden-hash equality is the primary verification
  mechanism for the whole product.
- **A moved golden hash is a defect until proven an intended, versioned change.** Regenerating rather
  than investigating means the regression suite has stopped being a test. Attribution comes from
  diffing produced against recorded output, never from the commit log alone.
- **The acceptance fixture is verified at 1, 5, 20 and 50 pages.** The heavier counts are not run by
  CI, so they must be run deliberately and their digests recorded.
- **A command naming one component and one property must be incapable of changing another.** That is
  a guarantee of the engine, so it must hold when the engine is handed hostile bytes directly.
- **Versioning is a commitment, not a convention.** Any change to layout, subsetting, emission, the
  locale table, or the toolchain is a breaking change for downstream suites that pin these hashes.
  Format/library compatibility rules remain a known-undefined item, needed before v1.
- **Time-to-first-PDF for a new integrator is minutes, from the README alone** — re-verified against
  the released artifact, not the working tree.
- **The licence boundary is MIT; nothing copyleft enters** at any depth of the module graph or the
  designer lockfile. Every redistributed asset travels with its terms, and the third-party licence
  manifest is a published release artifact.
- **The designer is fully offline after first load** — no network at render or preview time. Changing
  what is fetched, and when, changes a release-facing property.
- **A document must never fail to render for want of a font.** Shipped coverage faces, plus the fact
  that an embedded face travels inside the `.folio`, are what keep an unfetched family harmless.

## Technical Decisions

- **The matrix is four targets** — `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` under Node
  — compared across all four. The arm64 arm is mandatory: contraction divergence shows only in hash.
- **Golden hashes record the library version and the exact pinned Go toolchain.** Bumping the
  toolchain is a release event and a re-measurement, never a routine upgrade. No change lands
  without a fixture covering it.
- **Release tags carry the module subdirectory prefix** — `folio-go/v0.1.0` — and a future major
  becomes a `/v2` import path.
- **The public API surface freezes at the tag.** The existing census is stale and must be re-measured
  against what the later epics actually left. Reserved contract items freeze at the same moment.
- **Three standing release obligations discharge and are recorded at the tag:** publish the licence
  manifest as an artifact, review the exported surface as a whole, and make the call-graph walker
  type-precise or confirm its precondition.
- **The release procedure is deliberately incomplete and completed by whoever cuts the release** —
  version stamping, changelog policy, the tag command, and how a matrix result is recorded against a
  release are owned here, not deferred.
- **CI's red must be readable per job.** A by-design-red job sharing a workflow conclusion with the
  real guardrails makes two conditions one signal. Separate it structurally while keeping its
  purpose, its redness, and its stated meaning.
- **A local gate measuring less than CI cannot certify what CI reports** — same modules, and record
  the per-job conclusions read, not the workflow badge.
- **A guard must not be able to confuse "found nothing" with "could not look".** A scan populated
  from tracked files alone reports clean over files a story just wrote. Widening buys that collapse,
  not a larger count, so it is proved by a red arm, never by a green count.
- **A mutating prover must not share a tree with another story's gates** — it plants what the guard
  catches, so it runs against a throwaway repository in a temp directory.
- **A decision resting on something not having happened carries an executable assertion of that
  absence**, keyed on the capability rather than a filename or tracker proxy, whose failure message
  names what it invalidates; when the capability lands the tripwire is replaced, never deleted.
- **Generated artifacts are ignored one file at a time, with one tracked on purpose.** The build
  knows what it emits and the ignore list knows what it ignores, so assert the subset relation —
  prose beside the ignore block discharges only if someone finds it.
- **An invariant duplicated across the Go/TS boundary moves in one commit**, with a test reading both
  sides. Command integrity has two enforcement points — the encoder must not produce an ambiguous
  command, the engine must refuse one rather than resolve it by last-wins — and half is what produced
  the project's existing untied invariants.
- **Text is escaped by UTF-16 code unit, not code point.** Escaping a code point from its first unit
  emits a lone surrogate that Go replaces with U+FFFD, so a bind path or asset key binds where the
  author never named. Assert with non-BMP text; BMP-only tests cannot see it.
- **Offline is a build artifact, not a hope.** A service worker precaches the shell and engine assets
  under content-hashed, immutable, brotli-served URLs. Moving a payload out of the precache means
  moving it onto the already-shipped fetch-and-persist path — one transport, but not one provenance:
  curated licence terms and build-time allowlists stay with the tier that owns them.
- **Payload claims are measured before-and-after, never estimated**, and a stated budget margin is
  real only while the thing that would consume it has not landed.

## UX & Interaction Patterns

- **First load is stated, not hidden**: it names what is loading, says it happens once, and shows
  real progress. What that install covers is what the screen can honestly promise.
- **No new surface.** Where a fetch tier replaces bundled bytes the family control is already drawn
  and shipped; only the origin of the bytes changes, and an unavailable family says plainly that it
  cannot be picked right now.

## Cross-Story Dependencies

- **15.2b dispatches early, alongside other epics' in-flight work** — every story dispatched before
  it carries the hole it closes, so it must not wait for the tail. It also stands down the
  time-boxed per-story manual check standing in for the widened scan.
- **15.1 before everything**, and it has run: the move is attributed to Epic 9's text alignment
  wiring, +4 bytes per page on a per-page footer element across all four statement goldens, ruled
  intended, goldens re-recorded, sign-off re-attested rather than patched. The matrix's remaining red
  is that re-attestation, not an unknown hash.
- **15.0 depends on Epic 16's shipped fetch tier** — arrival with terms, persistence across sessions,
  the family control naming its sources, installing without embedding. It moves the catalogue onto
  that machinery rather than adding a second mechanism, and must also remove the build-emitted
  stylesheet rules that would otherwise lazily fetch bytes no longer present. It precedes the asset
  size-budget work, which measures its threshold against what 15.0 leaves behind.
- **15.2 before 15.3** — while the guardrails workflow is permanently red, "CI is clean" cannot be
  asserted at all.
- **15.2a before 15.3, and that placement is the deadline.** Its engine half narrows an exported API,
  so it must land before the surface freezes. It shares a file with earlier designer work that
  minimally fixed the same quoting helper — re-read that file rather than assume its shape — and it
  owns the shared-encoder consolidation that story was forbidden from growing into.
- **15.3 is last by construction.** It must re-affirm or amend the owner's earlier "cut after Epic 6"
  decision, naming which epics are inside v0.1.0: later epics postdate that decision without
  amending it, and several added public surface and format fields.
- **15.2 does not resolve the unmet floor** the by-design-red job advertises; that stays open and
  stays visible after the restructuring.
