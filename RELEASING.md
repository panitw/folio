# Releasing Folio

This document is the procedure for cutting a Folio release. It exists because
**a release with no written procedure is not a release** — the same rule
D-000.58 applies to gate procedures, one level up.

It is deliberately **incomplete**. It was written at Epic 4 planning to give
three deferred obligations somewhere real to live, and it grows a line each
time one is ruled. An incomplete procedure that names an obligation is
strictly better than a complete-looking backlog nobody opens at the moment of
release.

## When `folio-go/v0.1.0` is cut

**After Epic 6, at the end of the MVP** (owner decision, D-000.78). The public
Go API stays unfrozen through the designer work and the data-binding work, so
anything the browser side needs can still be added without a breaking change.

Cutting that tag is **irreversible in one specific way**: D-1.1.c fixes the
public API at it. Everything below is what must be true first.

## Release checklist

### 1. The third-party licence manifest ships with the release

**Obligation:** the committed manifest at **`lint/MANIFEST.md`** is attached to
the release as an artifact.

AD-26's substance already ships and is guarded continuously: every module in
the resolved graph carries a resolved licence, an unresolvable one fails the
build, and `TestManifestUpToDate` (`lint/internal/manifest/manifest_test.go`)
fails if the committed file drifts from what the generator produces. What this
line adds is **publication** — the part that had no home until this document
existed.

Regenerate with `cd lint && go run ./cmd/genmanifest` (from inside `lint/`,
which has its own `go.mod` — not from the repo root).

*Discharges DW-3, retired at Epic 4 planning.*

### 2. The public API surface is deliberate

**Obligation:** the exported surface of package `folio` is reviewed as a whole
before it freezes, because D-1.1.c fixes it at the tag.

Measured at the Epic 3 boundary gate: **40 items** — 31 top-level declarations
plus 9 struct fields — with zero internal leakage, verified by two independent
instruments.

*Backstop for DW-4. The live trigger is the pinned surface census, which
reddens when anything is added or removed; this line is what the tagging story
reads, not what fires.*

### 3. The call-graph walker is precise, or its precondition still holds

**Obligation:** `buildFolioCallGraph` (`folio-go/render_arch_test.go`) resolves
methods by name alone. Before the tag, either replace it with a `go/types`
version in `lint`, or confirm its precondition — that no two receiver types in
package `folio` declare the same method name — still holds.

*Backstop for DW-20. The live trigger is the pinned injectivity assertion
beside the walker itself; it fires at the commit that creates the collision,
which is years earlier than anyone reads this file.*

## What is NOT yet written here

Version stamping (`version.go` currently reads `Version = "0.0.0-dev"`),
changelog policy, the tag command itself, and how the cross-target hash matrix
result is recorded against a release. **None of these is deferred work with an
owner** — they are simply unwritten, and whoever cuts the first release writes
them here as they go.
