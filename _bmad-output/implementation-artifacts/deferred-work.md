# Folio MVP — Deferred Work

Work consciously **not** done, with a named owner and the story that deferred it. This file exists
so a deferral is a tracked decision rather than a silent gap. Append-only; mark items done in place
with the commit that closed them.

Owners: **Story N.M** = the named story picks it up as part of its own scope. **Epic N close** = due
before that epic's key is marked `done`. **Owner** = needs the project owner's call before anyone acts.

---

## Done

### DW-1 — A fixture-path override so shape-check red-proofs never touch the real golden fixture — **DONE**
- **Deferred by:** Story 1.2 (reviewer Finding 7, finisher DEFER)
- **Owner:** **Story 1.3** — it is the guardrails story, and this is a guardrail's testability gap
- **Committed at:** `f9c27b3`
- **Closed by:** Story 1.3's finisher commit (this file's own commit; see that commit's message for
  the SHA — a commit cannot self-reference its own hash in its diff). `folio-go/matrix_test.go`'s
  `loadExpectedFixture` was split into a pure `checkFixtureShape(path string) (expectedFixture,
  error)` taking the fixture's location as a parameter, plus `TestFixtureShapeCheckRedProof`, which
  red-proves AC16's shape check against scratch copies (sha256 widened into a per-target object,
  **and** sha256 of the wrong length or case — Finding 12, this story's QA review) and asserts
  byte-identity of the real `fixtures/minimal-rect/expected.json` before and after (`bytes.Equal`)
  — the real fixture is never mutated.
- **Correction (Blocker 3, this story's QA review):** the developer's first pass placed
  `checkFixtureShape`, `loadExpectedFixture` and `TestFixtureShapeCheckRedProof` in
  `folio-go/matrix_test.go`, which carries the `//go:build matrix` tag — so the red-proof executed
  in **zero** gates (the tagged suite is deferred to the Epic 1 boundary, D-000.4). DW-1 was marked
  DONE on that strength anyway, which was the defect. The finisher moved all three into the
  untagged `folio-go/fixture_test.go` (package `folio`'s ordinary home for fixture-shape helpers,
  alongside `isSHA256HexString`); `go test ./...` now runs `TestFixtureShapeCheckRedProof` on every
  story, measured. DW-1 is marked done here only now that its proof runs in a gate that runs.

**What was deferred.** AC16's fixture-shape check (`sha256` must be a 64-character lower-case hex
JSON *string*) has no way to be red-proved against a scratch copy of `fixtures/minimal-rect/expected.json`,
because the check reads the real fixture by a hard-coded relative path. The finisher re-verified the
property by exercising `isSHA256HexString` directly through a temporary, uncommitted test, and
deliberately never mutated the real golden fixture to produce a red.

**Why it was deferred rather than fixed.** Adding a path override is a change to how every fixture
consumer resolves its path — a seam that AD-21 and D-000.5 both constrain (repo-root `fixtures/`,
read by relative path at test runtime, never `go:embed`ed). Doing it inside Story 1.2, whose
deliverable is the matrix harness, would have widened that story into fixture-resolution design
with no review coverage for it.

**Why it matters.** AC16 exists to stop a developer meeting a red matrix leg from converting `sha256`
into a per-target map (D-1.2.2). A guard that cannot be shown going red is a guard nobody can trust,
and this one protects the project's central reproducibility claim.

**How we'd know it was forgotten.** A future story adding a second fixture-shape assertion and
red-proving it the same indirect way, or not at all.

---

## Open

### DW-2 — The licence check's JS half: `folio-designer/`'s lockfile
- **Deferred by:** Story 1.3 (ruling D-1.3.4)
- **Owner:** **Story 5.1** — the story that creates `folio-designer/`
- **Anti-rot mechanism:** Story 1.3 ships an assertion that the lockfile is **absent**. The day Epic 5
  creates it, that assertion goes **red** and the build stays red until the JS half is wired.

AD-26 requires the licence check to cover the whole Go module graph **and** `folio-designer/`'s
lockfile at any depth. The Go half ships complete at 1.3; the JS half has nothing to check yet.
A conditional "check it if present" was rejected because it starts silently passing the moment the
directory arrives — the guard would report success exactly when it stopped covering anything.

Same treatment, same story, same mechanism: the OFL 1.1 text for shipped faces (fonts arrive at
**Story 2.2**) and the Apache-2.0 NOTICE for `pdfjs-dist` (**Epic 5**) are asserted absent now, so
landing either breaks the build until the manifest covers it.

**Confidence on the owner:** medium. If Epic 5's shape shifts, the owner moves with whichever story
first creates the lockfile.

**Correction (Finding 8, this story's QA review, Major).** The three absence checks originally keyed
on exact guessed filenames (`folio-designer/package-lock.json`, `folio-go/fonts/OFL.txt`,
`folio-designer/third-party-notices/pdfjs-dist/NOTICE`) — a `pnpm-lock.yaml` or an `OFL-1.1.txt`
would have satisfied none of them and gone build-green, D-1.3.4's own rejected hazard arriving
through a side door. The finisher re-keyed both checks on the **directory** instead
(`folio-designer/` absent; `folio-go/fonts/` absent) — Story 5.1 cannot create the project without
creating `folio-designer/` first, and Story 2.2 cannot ship a face without creating
`folio-go/fonts/` first, so neither check depends on guessing a filename. The directory-level
`folio-designer/` check is strictly broader than the two checks it replaces (package-lock.json's
and the pdfjs-dist NOTICE's): both artifacts live inside `folio-designer/`, so any artifact landing
there — regardless of name — now trips the same finding. This still fully implements this
paragraph's "same treatment, same story, same mechanism" for the OFL text and the pdfjs-dist
NOTICE; it does not narrow anything DW-2 promises.

### DW-3 — Publishing the third-party licence manifest as a release artifact
- **Deferred by:** Story 1.3 (ruling D-1.3.4)
- **Owner:** **Epic 4 close** — the `folio-go/v0.1.0` tag

Story 1.3 ships the manifest **generated and asserted complete**: every module in the resolved graph
appears with a resolved licence, and an unknown or unresolvable licence fails the build. That is the
substance of AD-26. What is deferred is *publication* — attaching it to a release — because no
release process exists yet.

### DW-4 — Nobody owns cutting `folio-go/v0.1.0` — **owner decision when Epic 4 is planned**
- **Raised by:** the engineering lead during Story 1.3 rulings
- **Owner:** **the project owner**, at Epic 4 planning

No story in Epics 1–6 owns the tagging event itself. "Who cuts `folio-go/v0.1.0`, and what ships with
it" is a release-process decision with licence-compliance consequences (DW-3 depends on it), and
D-1.1.c fixes the public API at that tag — so it is also the moment the medium-confidence
argument-packaging question becomes irreversible. Not due now; **must not evaporate.** If it is still
unowned when Epic 4 is planned, it goes to the owner rather than being absorbed into a story.

### DW-5 — Derivation validation of `columns[].footerOf` from `bind`
- **Deferred by:** Story 1.4 (ruling D-1.4.2, AC43/AC44)
- **Owner:** **Story 3.2**, backstop **Story 3.7** (`folio.Validate` must include it)
- **Anti-rot mechanism:** a test asserts `folio-go/internal/expr` is **absent**. The day Story 3.2
  (or any story) creates that package, the assertion goes red and forces the derivation logic to be
  wired before the build can pass again.

D-1.1's derivation rule (a bare row-scoped path, or a single `formatNumber(...)` call over one) needs
the parsed expression tree to check `bind`'s shape — machinery Story 1.4 deliberately does not build
(no `internal/expr` package exists yet). Until Story 3.2 lands it, a `footer` with no `footerOf`
simply loads (AC44's known, fixture-pinned gap) rather than being derived or rejected.

### DW-6 — The two footer diagnostic codes: `TABLE_FOOTER_SOURCE_UNRESOLVED` / `TABLE_FOOTER_SOURCE_FORBIDDEN`
- **Deferred by:** Story 1.4 (ruling D-1.4.2)
- **Owner:** **Story 3.6** — the story that mints `internal/diag`'s stable diagnostic codes
- **Anti-rot mechanism:** a test asserts `folio-go/internal/diag` is **absent**. The day Story 3.6
  creates it, the assertion goes red and forces these two codes to be wired before the build can
  pass again.

Story 1.4's load failures are plain Go errors (D-1.4.2: *"1.4 must not mint them early"*).
`internal/diag` does not exist yet; AD-14 lands with Story 1.6 and the codes with Story 3.6.

### DW-7 — Footer evaluation sameness with `{{sum(...)}}` / `{{avg(...)}}` / `{{count(...)}}`
- **Deferred by:** Story 1.4 (ruling D-1.4.2)
- **Owner:** **Story 4.5**, by name
- **Anti-rot mechanism:** **none possible.** D-1.4.2's own honesty: *"no package tripwire exists, so
  `deferred-work.md` is the only trigger — flagged as the weakest of the three."* Nothing renders a
  table (and so nothing evaluates an aggregate) until Story 4.5, so there is no absent-package
  structural seam to key a red-proof on the way DW-5 and DW-6 have. This entry itself is the only
  thing keeping the requirement visible until then.

`columns[].footer`'s `sum`/`count`/`avg` must eventually use the *same* aggregate evaluation as the
`{{sum(...)}}` family of expression functions — a single implementation, not two that can drift.
Story 1.4 builds neither; nothing renders a table until Story 4.5.

### DW-8 — `Decimal` moves to `internal/expr` (or a leaf) and 1.6's path matcher is deleted
- **Deferred by:** Story 1.6 (rulings D-1.6.1, D-1.6.5)
- **Owner:** **Story 3.2** — the expression-language story
- **Forcing function:** DW-5's existing `internal/expr`-absent tripwire reddens the moment that
  package is created, which is what makes someone re-read this entry.

**Two obligations, one owner.**

1. **The `Decimal` type moves; it is never duplicated.** AD-23 **Binds** both `internal/bind` and
   `internal/expr`, and Epic 3's `sum`/`avg`/comparison need exact decimal arithmetic. But
   `internal/bind` imports `internal/expr`, so `expr → bind` would be an **import cycle** — a hard
   compile error. **The dangerous resolution is not the cycle (Go stops that) but duplicating the
   type to break it.** Pre-committed: `internal/expr` may never import `internal/bind`; when 3.2
   needs `Decimal`, the type **moves**.
2. **Story 3.2's parser replaces 1.6's path matcher — deleted, not kept alongside.** 1.6 accepts only
   a bare dotted path and rejects everything expression-shaped precisely so that two parsers never
   coexist. If 3.2 leaves the matcher in place, the wrong one eventually wins.

**How we'd know it was forgotten.** A second `Decimal` type anywhere in the module, or a dotted-path
matcher surviving alongside the expression parser after 3.2.
