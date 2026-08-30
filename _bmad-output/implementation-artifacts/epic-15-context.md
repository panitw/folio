# Epic 15 Context: Folio can be released

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

This epic ships no product feature. It closes the three process defects that stand between a
repository and a shippable Go library: a golden report digest that moved without an explanation, a
CI signal that is red by design and therefore unreadable, and the absence of any cut release. Byte
identity across the four build targets still holds — the four targets agree with each other — so the
determinism property itself is intact; what has failed is the discipline that makes that property
verifiable and the process that would let a third party depend on it. Until these are closed, every
other epic's work is unreleasable.

## Stories

- Story 15.1: The golden report's moved hash is explained
- Story 15.2: CI's red means something
- Story 15.3: `folio-go/v0.1.0` is cut

## Requirements & Constraints

- **Byte-reproducible rendering is the signature requirement.** Identical inputs must produce
  identical PDF bytes, on every target, at every page count. Golden-hash equality is the primary
  verification mechanism for the whole product — not a convenience.
- **A moved golden hash is a defect until proven otherwise.** Regenerating hashes to make CI green,
  rather than investigating why they moved, is an explicit failure mode the project tracks: it means
  the regression suite has stopped being a test. Attribution must come from diffing the produced
  output against the recorded output, not from reading commit history.
- **The acceptance fixture is verified at 1, 5, 20 and 50 pages** against recorded reference renders,
  not merely "produces a file". The heavier page counts are not exercised by CI, so they must be run
  and their digests recorded deliberately.
- **Versioning is a commitment, not a convention.** Any change to layout, subsetting, emission, the
  locale table, or the pinned toolchain is a breaking change for downstream test suites, because
  third parties pin these hashes. The forward/backward compatibility policy between template format
  versions and library versions is a known-undefined item and is the most expensive open question
  now that the library ships publicly.
- **Time-to-first-PDF for a new integrator is measured in minutes**, following the README alone. For
  this epic that check must be re-run against the released artifact, not the working tree.
- **The licence boundary is MIT and nothing copyleft may enter** at any depth of the module graph.
  The third-party licence manifest is a release artifact, published with the release — not a README
  paragraph.

## Technical Decisions

- **The CI matrix is four targets** — `darwin/arm64`, `linux/amd64`, `linux/arm64`, `js/wasm` under
  Node — with hashes compared across all four. An arm64 arm is mandatory because contraction-class
  divergence is invisible in output and appears only in the hash.
- **Golden hashes are recorded alongside both the library version and the exact Go toolchain
  version.** The toolchain is pinned exactly; bumping it is a release event and a re-measurement
  exercise, never a routine upgrade. No change lands without a fixture covering it.
- **Release tags carry the module subdirectory prefix** — `folio-go/v0.1.0`, not `v0.1.0` — because
  the module lives in a monorepo subdirectory. A future major becomes a `/v2` import path.
- **The public API surface freezes at the tag.** It was censused at an earlier epic boundary (40
  items: 31 top-level declarations plus 9 struct fields, zero internal leakage); that census is stale
  and must be re-measured against the surface the later epics actually left. Reserved public contract
  items — including reserved `params` keys — freeze at the same moment.
- **The release checklist already carries three obligations** that must each be discharged and
  recorded at the tag: publish the licence manifest as a release artifact, review the exported
  surface of the public package as a whole, and either make the call-graph walker type-precise or
  confirm its no-duplicate-method-name precondition still holds.
- **The release procedure is deliberately incomplete and is completed by whoever cuts the release.**
  Version stamping (the version constant currently reads a dev placeholder), changelog policy, the
  tag command itself, and how a cross-target matrix result is recorded against a release are all
  unwritten and owned by this epic, not deferred further.
- **CI's red must be readable per job.** A job that is red by design (keeping an unmet coverage floor
  visible) currently shares one workflow conclusion with the real guardrails, so two conditions share
  one signal and neither can be read alone. The remedy is structural separation into its own
  workflow, with the deliberately-red job keeping its purpose, its redness, and its stated meaning.
- **A local gate that measures less than CI cannot certify what CI reports.** CI runs formatting
  checks across all three Go modules; the boundary-gate procedure must do the same, and must record
  the per-job conclusions it read rather than the workflow badge.

## Cross-Story Dependencies

- **15.1 before 15.3.** A release cannot be cut against a matrix whose primary acceptance fixture is
  unexplained; the tag also has to record a matrix result, which presumes a green one.
- **15.2 before 15.3.** The release story's evidence rests on reading CI truthfully; while the
  workflow is permanently red, "CI clean" cannot be asserted.
- **15.1 depends on Epic 9, and on Epic 9 alone** — settled by measurement rather than suspicion
  (D-15.1.1, D-R7.5). All four golden statements moved, +4 bytes per page each. The mover is not the
  element-box paint but `folio-go/text_alignment.go`, created in the same commit (`791ed00`) and
  wired into the emitter, which finally acted on `style.align`. The text-ink work (Epic 10,
  `304442f`) moved nothing and is not a suspect.
- **15.3 discharges the long-open release-ownership item** and must explicitly re-affirm or amend the
  owner's earlier "cut after Epic 6" decision, naming which epics are inside v0.1.0 and which are
  not — the later epics all postdate that decision without amending it, and some have already added
  public surface and format fields.
- **15.2 does not resolve the coverage floor** that the deliberately-red job exists to advertise; that
  item stays open and stays visible after the restructuring.
