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

**Fonts half retired (Story 2.2, AC5).** Story 2.2 shipped the three production faces at
`folio-go/fonts/` (Noto Sans, Noto Sans Thai, Noto Sans SC — AC1, AC9), each with its OFL-1.1
`LICENSE-OFL.txt` and `NOTICE.md`, so the `absence-fonts-dir` tripwire's job is done: it fired the
day the directory landed, exactly as designed. It has been **removed** from
`lint/internal/rules/absences.go`'s `absenceChecks` and replaced by a fail-closed guard with the
opposite polarity — `ScanFontsAssets` (`lint/internal/rules/fontsassets.go`, rule ids
`fonts-asset-unaccounted` / `fonts-asset-missing`), which now REQUIRES `folio-go/fonts/` to exist
and to hold only recognised shapes (a font file, a `LICENSE*`/`NOTICE*` pair, or a `.go` source
file), red-proved both ways (a stray file at the real location, and — via
`lint/internal/manifest.ResolveAssets`, AC25 — a missing `NOTICE*` file). **The JS/lockfile half of
this entry is unaffected and remains open, owned by Story 5.1** — only the fonts half is retired
here; DW-2 as a whole is not closed.

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

### DW-9 — Re-test AC4's "nothing ceremonial" claim once a shipped font set exists — **RETIRED at Story 2.2**
- **Deferred by:** Story 1.7 (ruling D-1.7.1)
- **Owner:** **Story 2.2** — "The shipped font set and its fallback chain"
- **Retired:** the re-test ran. `folio-go/example_test.go` carries a compiled, EXECUTED
  `func Example()` (`go test -run Example` passes with its `// Output:` comment), rewritten against
  the shipped set: `folio.LoadTemplate`, then `fonts.Shipped()` — the `FontSet` obtained in ONE
  expression taking NO arguments — then `folio.Render`. `folio-go/README.md`'s "Your first PDF"
  section now shows this Example verbatim — and "verbatim" is mechanically checked, not asserted:
  `TestREADMEExampleBlockMatchesSource` byte-compares the fenced block against `example_test.go`,
  after the two were found to have drifted (the README kept the dead `err == nil &&` conjunct that
  the .go file had genuinely dropped). The "this step is ceremony Story 2.2 REMOVES" comment is
  removed (it is no longer true), and a new subsection explains why `folio`/`folio/fonts` stay two
  separate imports: a root re-export would embed **~11.3 MB raw** into every caller's binary,
  including the wasm build. *(That figure was `~9 MB` here and in the README until Story 2.2's
  finisher — `go:embed` stores RAW bytes, so the binary-size argument takes the uncompressed
  measurement; ~9 MB is NFR7's COMPRESSED download budget, against which the shipped faces measure
  5.07 MB at `brotli -q 11`.) **Verdict: the claim held — the re-test did not surface a
  `DECISION NEEDED`.**

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
which check kind that entry is, and `TestAbsencesChecksIncludeAllFourEntries` still pins the rule ids
by name (now including `absence-source-date-epoch`), so a silently shrunk list still fails loudly
either way. *(Renamed and reduced from five to four by Story 2.2, which retired `absence-fonts-dir`
when it shipped the faces that tripwire existed to force — see DW-2.)*

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

#### Story 2.4's discharge — AC18. **No new items were added. The load-bearing count remains 2.**

Story 2.4 falls inside DW-11's stated owner window ("Epic 2's later stories"), so it owes an answer in
writing rather than silence. The answer is: **none were found, and none were invented.**

The dev agent had no access to a sourced, independently-attested register of Thai personal names, and
D-000.17 — reinforced by D-2.1.15 Major 5's own precedent, where "a plausible surname" was demoted to
`synthetic_probe` rather than counted — forbids manufacturing attestation to reach a number. Adding a
name that *looks* opaque would move the figure from 2 to 3 while moving the evidence not at all, which
is the failure this entry exists to keep visible. **`ดอเลาะ` and `แนแซ` remain the only two.**
`corpus.json` and `cmd/gencorpus/main.go` are byte-unchanged by Story 2.4.

**What did change, and why it is not a discharge.** Story 2.4 closed the P2 defect these two items were
the load-bearing witnesses *for* (26 violations across 17 corpus items to 0, D-2.1.9). That removes the
**symptom** DW-11 was tracking the risk of, and it does not remove the **thinness**: coverage of the
hard path still rests on two attested items, and a future regression in that path would still be
detected by only those two. Story 2.4's new fixture `fixtures/expected-breaks/` exercises the same path
from a second direction — `ดอเลาะ`, `แนแซ`, `ชัยวัฒน์` and `ฉั่วสมบูรณ์` all appear there as
zero-break labels — but that is a **conformance fixture, not the corpus**, it contributes nothing to
P6g, and it adds no attested name. **DW-11 stays open, at 2.**

### DW-12 — Every later pinned instance inherits AC7's golden + matrix obligation
- **Deferred by:** Story 2.2 (ruling D-2.2.1)
- **Owner:** **whichever story next adds a pinned instance of a shipped variable face** (on current
  planning, a candidate is Bold — a `wght`-axis instance — should it arrive; not committed to any
  numbered story here)

Story 2.2 ships exactly one pinned instance per shipped face — each face's DEFAULT instance (Bold
and other non-default named instances are explicitly out of this story's scope). D-2.2.1's binding
standing condition: **value-dependence is the whole hazard** — a clean four-target result on one
(face, pinned instance) pair says nothing about a different pair of the SAME face, because
instancing arithmetic (gvar/avar interpolation) is value-dependent (D-2.2.0's own measured limits).

**So:** any later story that adds a NEW pinned instance of a face already shipped (e.g. a Bold
weight) must (a) add its own golden — the embedded, instanced program's own digest, the same shape
`folio-go/fixture_test.go`'s `TestMultiScriptFallbackGoldenFixture` already records for Story 2.2's
three default instances — and (b) re-run the full four-target matrix in that story, not defer it to
an epic boundary. This is a standing obligation this story registers so a later reader does not have
to rediscover why (D-000.4's per-story override list already gets amended each time a story is
identified as one; this is the reasoning that identifies the NEXT one before it is drafted).

**How we'd know it was forgotten.** A story ships a new pinned instance of an existing shipped face
with no new golden recorded for it, and no four-target matrix re-run logged in its Delivery Log.

---

### DW-13 — Size the uncompressed-`FontFile2` payload cost against real CJK content, then put it to the owner

**Owner:** orchestrator to schedule the sizing; **the adoption decision is the project owner's.**
**Raised at:** Story 2.2, while verifying the embedded programs' table sets.

**The observation.** Folio's `FontFile2` streams ship **uncompressed** — `/Length` equals `/Length1`,
no `/Filter` — while the project uses `/FlateDecode` with `/Predictor 15` elsewhere.

**It is deliberate and mechanically enforced, not drift.** `lint/internal/rules/nocompressor.go`
(`no-compressor-import`) forbids any file under `folio-go/` importing `compress/flate`,
`compress/zlib` or `compress/gzip`, with a retained violating fixture at
`testdata/lint/no-compressor/violating-compressor/bad.go`. Per D-1.8.1, *"no compressor is invoked"*
is **the mechanism that keeps R4 closed** — `acceptance.md:83`: *"compressor output is stable by
observation, not by contract."* The image route embeds each file's **own already-compressed bytes**;
it never invokes a compressor. **Compressing font streams is therefore impossible without retiring or
narrowing a guard that has been proved to fire.** Nothing to change in Story 2.2; its report cites the
rule id and D-1.8.1 and stops.

**Why it still deserves sizing.** The fixtures hide the magnitude — the CJK subset is **732 bytes** —
and that is exactly why the question read as invisible. Rough bound: the static CJK face is 10.6 MB
over ~20k glyphs, so roughly 300–800 bytes of `glyf` per character; a Chinese bank statement plausibly
uses 300–800 distinct characters, giving **~150–400 KB embedded per PDF**, against which Flate on
`glyf` typically saves about half. **Honest expectation: tens to low-hundreds of KB per document —
real, bounded, and quite possibly not worth its cost.** Measuring may well return "leave it", which is
a good outcome and cheap to reach.

**Two constraints to settle before anyone proposes a clever middle path:**

1. **The image passthrough precedent does not extend here.** Images work because the file arrives
   already compressed and we embed its own bytes. A font subset is **synthesised at render time**, so
   there are no pre-existing compressed bytes to pass through — and WOFF2 is not available, since
   `FontFile2` requires raw TTF. **Along that axis the choice is binary:** invoke a compressor or do
   not. There is no third way, and someone will propose one.
2. **There is a genuine third option on a different axis** — a **vendored, version-pinned DEFLATE
   implementation** instead of `compress/flate`. That converts R4 from *"we cannot control the
   compressor"* into *"we pin the compressor exactly as we pin the toolchain"*, which is
   philosophically identical to AD-22's move and arguably **more** stable than stdlib, whose flate has
   been tuned across releases while a pinned vendored copy would not be. Costs: a new dependency
   through AD-26's licence check and D-1.5.1's allowlist, plus narrowing the guard to permit exactly
   that one import — which keeps the decision **visible** rather than dissolving it.

**Framing for the owner, not to be over-stated:** adopting compression **widens** R4 rather than
creating a new risk. Golden hashes are *already* toolchain-sensitive — AD-22 makes a toolchain bump a
release event requiring re-measurement. Compression makes **more** of the output sensitive, and
sensitive to a component historically tuned across releases. That is a difference of **degree**, and
the owner should weigh it as one.

**Also do, cheaply and now:** state the **consequence** beside `acceptance.md:83`'s R4 note — *font
programs and content streams ship uncompressed, so Folio's PDFs are deliberately larger than a typical
producer's.* The document already records the **reason**; what is missing is what follows from it,
which is the sentence someone will otherwise re-derive — or "fix".


---

### DW-14 — `/ToUnicode` emits one unbounded `beginbfchar` section; the spec caps a section at 100

**Raised by:** Story 2.3's QA review (Finding 7, Nit). **Deferred by:** Story 2.3's finisher.
**Owner:** the Epic 2 boundary gate, or Story 2.4 if its corpora reach the limit first.

**What.** `internal/pdf.buildToUnicodeCMap` emits the whole CMap as a single
`N beginbfchar … endbfchar` block with `N = len(face.ToUnicode)`. The ToUnicode CMap specification
limits one `beginbfchar`/`endbfchar` section to **100 entries**. A document whose face needs more than
100 distinct CIDs would emit a section a strict validator rejects.

**Why it is deferred rather than fixed in 2.3.** It is **pre-existing, not a 2.3 regression** — the
pre-2.3 derivation had the same unbounded shape, so the defect is inherited and charging it to this
story would be wrong on the record. It is also not reachable by anything folio ships today. Measured
across every text fixture at this commit, every section is well under the cap:

```
font-text              : 25
multi-script-fallback  : 4, 1, 1
shaped-text            : 14, 7, 28
```

**Story 2.4's measurement — AC17. NOT TRIGGERED; this entry stays open.** DW-14 named 2.4 as its owner
*"if its corpora reach the limit first"*, and a wrapped Thai paragraph was the plausible input that
would. It does not. Measured on the produced bytes of the new `fixtures/wrapped-text/` render (four
text elements, three scripts, all three shipped faces embedded):

```
wrapped-text           : 28, 18, 38
```

Largest section **38**, against a cap of 100. The fixture was **not sized to duck the cap** — its box
widths were chosen to force wrapping (the numbers are in the story's AC15 record) and the section sizes
are simply what that produced; the largest is comparable to `shaped-text`'s existing 28.

`folio-go/wrapped_text_fixture_test.go`'s `assertToUnicodeSectionsUnderCap` now measures this on every
run and **fails loudly if any section exceeds 100**, with a message saying stop-and-escalate rather
than fix-in-place — because the remedy (chunking into ≤100-entry sections) moves the golden hash of
every document over the cap, and this entry's own text asks that it *"land with a deliberate re-record
rather than as a drive-by"*. So the trigger is now a test rather than a reader's vigilance.

**Why it is worth a standing entry rather than a passing mention.** Story 2.3 is the story that both
**rewrote this function's entry source** (CIDs are now allocated per (glyph, cluster text), so the
entry count is no longer bounded by the glyph count) **and produced the largest section to date (28)**.
Story 2.4's larger Thai corpora push it further in the same direction. The trend and the rewrite are
both this story's; only the fix is not.

**Fix when taken:** chunk into sections of at most 100 entries. It is a local change to one emitter
with no effect on any document currently under the cap — but it **will move every golden hash of any
document that exceeds it**, so it wants to land with a deliberate re-record rather than as a drive-by.

**Do not** "fix" this by capping the number of CIDs; the CID allocation is D-2.3.2 and is correct.

---

### DW-15 — First-baseline placement and inter-baseline spacing use two different models

**Owner:** engineering lead to schedule; **must land BEFORE the Thai reading sign-off is recorded.**
**Raised at:** Story 2.5 (creator), ruled a defect by the lead at DN-3. **Deliberately not landed in 2.5.**

**The inconsistency.** `internal/pdf/textdoc.go:730` places the first baseline with
`pdfY = flipY(..., run.FontSize)` — i.e. from the **font size**. But [[D-2.4.2]] derives inter-baseline
spacing from **`hhea` ascent − descent (+ lineGap), maximised over the declared chain**. So the *first*
line is positioned on one model and *every subsequent* line spaced on another.

**Why it is a defect and not a tolerable approximation.** The two models agree only by coincidence, and
the symptom is the classic one: *"why does the first line sit differently from the rest?"* — asked
months later, by someone with no reason to suspect two different formulas. It is also
scale-dependent: the discrepancy grows with the gap between font size and the chain's maximum
ascent, so a Thai or CJK chain diverges more than a Latin one.

**Why it was not fixed in Story 2.5** *(this is the load-bearing part)*: correcting it **re-records
four goldens**, and one of them is `fixtures/shaped-text/expected.pdf`, whose digest
`5964aad0…c92e00f` is bound to a **pending human sign-off**. Landing it inside 2.5 would have moved a
golden a human had been asked to examine, silently invalidating a request already in flight.

**Sequencing obligation** *(binding, per [[D-000.26]] refined)*: **this fix lands before the Thai
reading sign-off is recorded**, not after. The reading sign-off binds to the rendered image, so a
baseline shift invalidates it and the human would be asked twice. **The scarce human is the
constraint, not the code.** The Thai *break* sign-off is unaffected — it binds to the break-opportunity
vector, which a uniform baseline shift does not touch.

**On landing it**: re-record all four goldens as one intended versioned change under AD-21/AD-22, each
with its D-000.22 semantic acceptance step — and per [[D-000.30]], capture the pre-fix measurement of
the discrepancy **before** applying the fix, since that window closes permanently.


---

### DW-16 — `pagemodel.ShapedGlyph.CID` is not always a glyph id, and the table needed to interpret it is not in the page model

**Owner:** **engineering lead to rule on the shape**; the fix lands in **the first non-PDF renderer
story**, which is its natural forcing function.
**Raised at:** Story 2.5 (reviewer Finding 2, a Major; finisher DEFER — recorded, not re-architected).
**Nothing regressed.** The allocation is Story 2.3's [[D-2.3.2]], unchanged. Story 2.5's type move
relocated it into `internal/pagemodel` and thereby made it visible.

**The property, precisely.** `pagemodel.ShapedGlyph.CID` carries **two different kinds of value**,
and the field's name and doc comment describe only the first:

- In `buildShapedPDFRuns`' **base block** (`internal/pdf` is not involved; the site is `render.go`'s
  `cid = newGID`, guarded by `state.baseClaimed`), `CID` **is** the subset glyph id. A renderer
  holding the subset font can look the glyph up directly. This is AD-5's "glyph ids", legitimately.
- In the **`default` block**, a **second, synthetic identifier is minted** for a glyph that carries a
  *different source text* — `cid = uint16(sub.NumGlyphs + len(state.extras))`. That value is **not**
  a glyph id in any font. It is an index into `state.extras`, and its ceiling is stated by the code
  itself: *"exceeds Identity-H's two-byte CID ceiling of 65535."*

That branch exists **only** because PDF's `/ToUnicode` CMap maps one CID to one text. So the field's
value range and its second meaning are both **defined by a PDF encoding**.

**What a non-PDF consumer cannot do today.** The `CID → GID` map that resolves the synthetic values —
`state.extras`, and the `pdf.CIDText` entries built from it — is deliberately kept **out** of the page
model, correctly, because it is a PDF construct. The consequence is that a PNG/SVG/HTML renderer
handed a `pagemodel.TextRun` **cannot resolve `Glyphs` back to glyphs**: for any glyph whose text
differs from its first-seen text, `CID` indexes a table the renderer was not given. It cannot detect
which case it is holding either — the two kinds of value are indistinguishable at the type.

**Why this matters more than the field.** AD-5's stated purpose for keeping PDF out of the page model
is *"that absence is what keeps PNG/SVG/HTML renderers possible later."* The page model now carries
a field that partially defeats that, and AC1's substring guard **cannot see it**: `"cid"` is not in
`pdfConceptSubstrings`, and adding it would be the **wrong fix**, because the base-block value is a
perfectly legitimate subset glyph id. There is no lint that closes this; only a ruling does.

**What would fix it** — the ruling needed, not a decision this entry may take:

1. **Renderer-neutral.** `ShapedGlyph` carries a true `GlyphID` plus a separate text association, and
   the PDF writer performs the CID allocation on its own side of the boundary, from those two.
   `internal/pdf` already owns `CIDText`; this moves the *allocation* to sit beside it. Strictly more
   faithful to AD-5, and strictly more work.
2. **Amend AD-5.** Admit an encoding-scoped identifier explicitly, rename the field to say so, and
   record that a non-PDF renderer needs an accompanying table which the page model does not carry.
   Cheaper, and honest, but it narrows what the page model promises.

**Why it is worth an entry rather than a passing mention.** The window in which this is cheap to see
closes as soon as **more producers write the field**. Today there is exactly one
(`buildShapedPDFRuns`), so option 1 is a local change. Every additional producer makes both options
more expensive, and makes the field's dual meaning harder to establish from the code.

**Do not** "fix" this by adding `cid` to `pdfConceptSubstrings`, and do not re-architect
`buildShapedPDFRuns` opportunistically — neither Story 2.5's AC1 nor its Task 4 asked for it, and the
developer was right not to attempt it.
