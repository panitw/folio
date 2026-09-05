# Epic 15 Context: Folio can be released

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

This epic ships no product feature. It closes the process and integrity defects that stand between a
repository and a shippable Go library: a golden report digest that moved without explanation, a CI
signal that is red by design and therefore unreadable, a component-command encoding that can be made
to name one thing and change another, and the absence of any cut release. Byte identity across the
four build targets still holds — the targets agree with each other — so the determinism property
itself is intact; what has failed is the discipline that makes that property verifiable, the
enforcement that makes a command's guarantee real, and the process that would let a third party
depend on either. Until these close, every other epic's work is unreleasable.

## Stories

- Story 15.1: The golden report's moved hash is explained
- Story 15.2: CI's red means something
- Story 15.2a: A component command means exactly what it names
- Story 15.3: `folio-go/v0.1.0` is cut

## Requirements & Constraints

- **Byte-reproducible rendering is the signature requirement.** Identical inputs must produce
  identical PDF bytes, on every target, at every page count. Golden-hash equality is the primary
  verification mechanism for the whole product — not a convenience.
- **A moved golden hash is a defect until proven otherwise.** Regenerating hashes to make CI green,
  rather than investigating why they moved, is an explicit failure mode the project tracks: it means
  the regression suite has stopped being a test. Attribution must come from diffing the produced
  output against the recorded output, never from reading commit history alone.
- **The acceptance fixture is verified at 1, 5, 20 and 50 pages** against recorded reference renders,
  not merely "produces a file". The heavier page counts are not exercised by CI, so they must be run
  deliberately and their digests recorded.
- **A command that names one component and one property must be incapable of changing another.**
  This is a guarantee of the engine, not a property of any one encoder, so it has to hold when the
  engine is handed hostile bytes directly.
- **Versioning is a commitment, not a convention.** Any change to layout, subsetting, emission, the
  locale table, or the pinned toolchain is a breaking change for downstream test suites, because
  third parties pin these hashes. Forward/backward compatibility rules between template format
  versions and library versions remain a known-undefined item — needed before v1, not before MVP,
  but now a commitment to third parties rather than an internal convention.
- **Time-to-first-PDF for a new integrator is measured in minutes**, following the README alone. For
  this epic that check must be re-run against the released artifact, not the working tree.
- **The licence boundary is MIT and nothing copyleft may enter** at any depth of the module graph or
  the designer's lockfile. The third-party licence manifest is a release artifact, published with
  the release — not a README paragraph.

## Technical Decisions

- **The CI matrix is four targets** — `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` under
  Node — with hashes compared across all four. The arm64 arm is mandatory because contraction-class
  divergence is invisible in the output and appears only in the hash.
- **Golden hashes are recorded alongside both the library version and the exact Go toolchain
  version.** The toolchain is pinned exactly; bumping it is a release event and a re-measurement
  exercise, never a routine upgrade. No change lands without a fixture covering it.
- **Release tags carry the module subdirectory prefix** — `folio-go/v0.1.0`, not `v0.1.0` — because
  the module lives in a monorepo subdirectory. A future major becomes a `/v2` import path.
- **The public API surface freezes at the tag.** It was censused at an earlier epic boundary and that
  census is stale; it must be re-measured against the surface the later epics actually left.
  Reserved public contract items freeze at the same moment.
- **The release checklist already carries three obligations** to discharge and record at the tag:
  publish the licence manifest as a release artifact, review the exported surface of the public
  package as a whole, and either make the call-graph walker type-precise or confirm its precondition
  still holds.
- **The release procedure is deliberately incomplete and is completed by whoever cuts the release.**
  Version stamping, changelog policy, the tag command itself, and how a cross-target matrix result is
  recorded against a release are unwritten and owned by this epic, not deferred further.
- **CI's red must be readable per job.** A job that is red by design — keeping an unmet floor
  visible — currently shares one workflow conclusion with the real guardrails, so two conditions
  share one signal and neither can be read alone. The remedy is structural separation into its own
  workflow, with the deliberately-red job keeping its purpose, its redness, and its stated meaning.
- **A local gate that measures less than CI cannot certify what CI reports.** CI runs formatting
  checks across all three Go modules; the boundary-gate procedure must do the same, and must record
  the per-job conclusions it read rather than the workflow badge.
- **An invariant duplicated across the Go/TS boundary moves in one commit**, with a test that reads
  both sides. Command integrity has two enforcement points — the designer must not be able to encode
  an ambiguous command, and the engine must refuse one rather than resolve it silently by last-wins
  — and shipping only one half is the pattern that produced the project's existing untied invariants.
- **Text is encoded by UTF-16 code unit, not by code point.** Escaping a code point from its first
  unit emits a lone surrogate that the Go side replaces with U+FFFD, so a bind path or asset key can
  silently bind somewhere the author never named. Any assertion of this must use non-BMP text; it is
  invisible to BMP-only tests.
- **The designer runs fully offline**, with neither template nor data leaving the machine. Anything
  in this epic that changes network behaviour changes a release-facing property.

## Cross-Story Dependencies

- **15.1 before everything.** A release cannot be cut, and the corpus-hashes-identically guard other
  epics rely on cannot be evaluated, while the primary acceptance fixture's digest is unexplained.
  15.1 has run: the move is attributed to Epic 9 alone — text alignment wired into the emitter,
  moving one text-matrix operand on a per-page footer element, +4 bytes per page across all four
  statement goldens — ruled intended, goldens re-recorded, and the human sign-off re-attested rather
  than patched. The matrix's remaining red is that re-attestation, not an unknown hash.
- **15.2 before 15.3.** The release story's evidence rests on reading CI truthfully; while the
  guardrails workflow is permanently red, "CI is clean" cannot be asserted at all.
- **15.2a before 15.3, and that placement is the deadline.** Its engine-side half narrows an exported
  API, so it must land before the surface freezes. It also shares a file with earlier designer work
  that made a minimal fix to the same quoting helper — re-read that file rather than assuming its
  earlier shape — and it owns the shared-encoder consolidation that the earlier story was forbidden
  from growing into.
- **15.3 is last by construction** and discharges the long-open release-ownership item. It must
  explicitly re-affirm or amend the owner's earlier "cut after Epic 6" decision, naming which epics
  are inside v0.1.0 and which are not: later epics all postdate that decision without amending it,
  and several have already added public surface and format fields.
- **15.2 does not resolve the unmet floor** the deliberately-red job exists to advertise; that item
  stays open and stays visible after the restructuring.
- **Two further items are sequenced into this epic by the sprint plan rather than by the epic text:**
  a catalogue-face fetch tier (placed here because it changes offline behaviour, a release-facing
  property, and because format freedom expires at the tag) and the asset size-budget gate that
  follows it. Both run ahead of 15.2 and 15.3.
