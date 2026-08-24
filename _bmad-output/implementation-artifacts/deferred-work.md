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

### DW-9 — Re-test AC4's "nothing ceremonial" claim once a shipped font set exists
- **Deferred by:** Story 1.7 (ruling D-1.7.1)
- **Owner:** **Story 2.2** — "The shipped font set and its fallback chain"

Story 1.7's AC4 requires that producing a first PDF takes *"a load call and a render call, and nothing
ceremonial"*, and D-1.1.c named the README as the test of whether five positional arguments read as
ceremony. **That test cannot be run fairly at 1.7.** Verified: there is no README in the repo (1.7
writes the first), and `folio-go/fonts/` does not exist — the shipped faces arrive at 2.2. So 1.7's
first-PDF example must show a caller assembling a `FontSet` from their own bytes, **which is ceremony
that Story 2.2 removes, not ceremony the signature causes.**

**The packaging decision itself is closed** (D-1.7.1, on AD-8: an options struct would make `FontSet`
omittable at compile time, turning an AD-8 violation from a compile error into a runtime one). What is
deferred is only the **ceremony judgement** on the README example.

**At 2.2:** rewrite the first-PDF example using the shipped default font set and re-read it. If it
still reads as ceremony, that is a `DECISION NEEDED` — and it must be raised **before the
`folio-go/v0.1.0` tag** (Epic 4 close, DW-3/DW-4), after which the signature is fixed.

**How we'd know it was forgotten.** A `v0.1.0` tag cut without anyone re-reading the first-PDF example
against a shipped font set.

### DW-10 — `/CreationDate` and `/ModDate` wiring — **ALREADY OWNED, not newly deferred**
- **Raised by:** Story 1.7's creator (the clause became reachable when `params` first existed)
- **Owner:** **Story 3.7** — "Validate a template and render it from the command line"
- **Status:** already scheduled with acceptance criteria written; this entry exists only so the state
  *"`params` exists and nothing reads it for `/Info`"* is not silently forgotten.

AD-7 (`ARCHITECTURE-SPINE.md:201`) says `/CreationDate` and `/ModDate` are *"omitted **unless a date
arrives through `params`**"*. Until Story 1.7 `params` did not exist, so the condition was
**unreachable**. Story 1.7 creates the trigger and deliberately does not wire it.

**Story 3.7 already owns it** (`epics.md:1074–1100`): its user story says *"pin document dates
reproducibly"*, its **Covers** line names **AD-7**, and it carries criteria for `SOURCE_DATE_EPOCH`
being read by the CLI and passed in as a parameter — with the library core still reading no
environment variable — plus the negative case (*"no date supplied by any route → omitted"*).

**Forcing function — RE-KEYED at Story 2.1 (D-2.1.5).** Story 1.7 originally added an absence
tripwire keyed on the PATH `folio-go/cmd/` existing at all, as a proxy for "the CLI that reads
`SOURCE_DATE_EPOCH` has arrived." That key was **broader than the purpose**: `cmd/` has more than one
legitimate tenant, and Story 2.1's own build-time tooling (`cmd/gentrie`, `cmd/gencorpus`,
`cmd/genbreaks`) tripped it despite having nothing to do with AD-7 or params-date wiring — a measured
false positive, confirmed independently (`TestAbsencesProductionScan` failing, naming
`absence-cmd-dir`, before this re-key).

**The row is now keyed on its trigger, not on a path:** `SOURCE_DATE_EPOCH` must not appear in any Go
source under `folio-go/` until this is settled. Implemented as a new check KIND
(`absenceKindContent`) in `lint/internal/rules/absences.go`, rule id **`absence-source-date-epoch`**
(was `absence-cmd-dir`) — it scans `.go` files under `folio-go/` (excluding `testdata/`) for the
literal string `SOURCE_DATE_EPOCH`, rather than checking whether a directory exists. Red-proofed by
injection at the real repo location (a scratch reference added under `folio-go/`, observed
`TestAbsencesProductionScan` fail naming the new rule, then removed) and by a permanent fixture
(`folio-go/testdata/lint/absences/violating/folio-go/internal/paramsdate/placeholder.go`, replacing
the old `.../folio-go/cmd/placeholder.go`). The coverage witness (`AbsencesStats.ChecksEvaluated`)
was verified to still count this row: it increments once per entry in `absenceChecks` regardless of
which check kind that entry is, and `TestAbsencesChecksIncludeAllFiveEntries` still pins five rule ids
by name (now including `absence-source-date-epoch`), so a silently shrunk list still fails loudly
either way.

**The general rule this produced** (recorded in Story 2.1's Dev Notes too): key a guard on its
purpose, not on a proxy for its purpose. Where the key is broader than the purpose, the gap is where
false positives live — and a false positive in a guard invites exactly the workaround (weakening the
guard) that erodes it fastest. `cmd/folio` — the CLI itself — will still trip this the moment Story
3.7 writes `os.Getenv("SOURCE_DATE_EPOCH")` anywhere under `folio-go/`, regardless of what its `cmd/`
subpackage is named.

**Trade-off, named explicitly (this story's code review, Finding 14) so 3.7's reviewer checks for
it**: purpose-keying traded an UNEVADABLE predicate for an EVADABLE one. The old path key
(`folio-go/cmd/` existing) could not be dodged by spelling. The new content key
(`strings.Contains(source, "SOURCE_DATE_EPOCH")`) does not fire on `"SOURCE_DATE_" + "EPOCH"`, on a
constant defined elsewhere and referenced by name, or on a value read from a variable — Story 3.7's
developer meeting a red build now has a cheaper workaround than before existed. This is an accepted
trade (a guard keyed on a real proxy that can occasionally be evaded beats one keyed on the wrong
thing that never fires falsely but also never fires correctly for a legitimate second tenant of the
path), but it is a trade, not a strict improvement, and 3.7's reviewer should specifically check for
`os.Getenv` calls under `folio-go/` (already banned outside `_test.go` by AD-1) rather than trusting
this content match alone.

Story 1.7 also re-scopes `TestRenderHasNoCreationOrModDate` from an unconditional assertion to
*"params carrying no date"*, so 3.7's developer does not meet a red test whose cheapest resolution is
to weaken it — this part is unchanged by the Story 2.1 re-key.

**Blast radius, measured smaller than feared:** only fixtures that **supply a date** would move. A
params-carrying render with **no** date is byte-identical to today, so 3.7's impact is new fixtures
plus any existing fixture that opts in — **not the corpus**.

**How we'd know it was forgotten.** `cmd/folio` existing while `/CreationDate` is still emitted
unconditionally-absent with no params date path.

### DW-11 — S4's opaque-name coverage is thin: 2 genuinely-uncoverable sourced items on its most fragile path
- **Raised by:** Story 2.1's re-measurement under D-000.17 (a floor reported unmet, not filled)
- **Corrected by:** Story 2.1's finisher, per its second QA review (Minor 1 and Major 5) — the load-bearing
  count this entry originally carried (8) conflated two different properties. See *Corrected count*
  below.
- **Owner:** **Epic 2's later stories and Epic 4's golden-report work** — add genuinely-opaque sourced
  Thai personal names as they are found
- **Status:** open, and deliberately visible

P6g's floor asked for **≥20** genuinely opaque (zero-interior-break) sourced Thai personal names.
The generator's own P6g count (its literal criterion — "the unconstrained matcher proposes no interior
break") is **7** (was reported as 8 before this correction; see below). Per D-000.17 this was
**reported unmet, not filled.**

**Corrected count — the honest load-bearing figure is 2, not 7 and not the original 8.** The second QA
review (Minor 1) measured that P6g's criterion is satisfied by two structurally different populations
that must not be conflated:

| Surname | Whole dictionary entry? | Genuinely uncoverable (the hard path P2 fails on)? |
|---|---|---|
| `ดอเลาะ` | no | ✅ yes — independently attested (Thai-Malay/Muslim regional surname) |
| `แนแซ` | no | ✅ yes — independently attested (Thai-Malay/Muslim regional surname) |
| `ชินวัตร`, `จิราธิวัฒน์`, `หวั่งหลี`, `ประยูรวงศ์`, `ทวีสิน` | **yes — all five** | ❌ no — these exercise the OPPOSITE path (whole-word match, nothing to override) |

An eighth item, `ฉั่วสมบูรณ์` ("a plausible Sino-Thai family name" per its own original comment), was
**removed from this bucket entirely** by the finisher (D-000.17's "may not invent items to reach a
number" applies to attestation, not just to obsolete-character padding): "plausible" is not sourced,
so rather than retroactively claim it was attested, `cmd/gencorpus/main.go` now labels it
`synthetic_probe` and excludes it from every genuine floor, exactly like the 38 obsolete-consonant
probes. This is why the generator's own P6g figure is 7, not 8.

**So the true load-bearing count for S4's most fragile path (the one P2 demonstrably fails on) is 2,
not 7 and not 8.** The five whole-dictionary-entry names satisfy P6g's literal wording but exercise the
*other* polarity P6g exists to guarantee (nothing proposed, nothing to override) — they are real and
correctly counted under P6g's criterion, but they carry none of the risk this deferred item is about.

**Why the shortfall is not free — measured.** `ฅ (U+0E05)` appears in **2 of 62,107** dictionary words,
so the 38 synthetic obsolete-character strings are **near-trivially uncoverable**: nothing can
partially match inside them, and the atomic-run rule succeeds **the easy way**. The real opaque names
that produced violations (`ดอเลาะ`, `แนแซ`) are built from ordinary characters appearing in **thousands**
of words — which is **why** they violated, via the resume scan landing on a spurious short match inside
a run already declared uncoverable. **That is the hard path, and the path P2 demonstrably fails on.**

**So the 2 genuinely-uncoverable, independently-attested sourced opaque names are load-bearing for
S4** — currently the **only** items covering the path where P2 breaks. **The 38 synthetics (plus the
one reclassified name) cover the easy path and must not be counted as substitutes.**

**Context that explains the shortfall rather than excusing it:** 115 of 122 sourced names (**94%**)
decompose into recognisable morphemes, because Thai naming convention favours composing from meaningful
words. **Genuinely opaque real names appear to be a real minority of the language, not a sourcing-effort
gap.** That measured fact is now available to whoever specifies S4's adequacy criteria.

**How we'd know it was forgotten.** S4 still carrying only 2 genuinely-uncoverable, independently-attested
opaque items when Epic 4's golden report ships.
