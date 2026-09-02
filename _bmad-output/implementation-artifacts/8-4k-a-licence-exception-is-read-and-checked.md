---
title: 'Story 8.4k — A licence exception is read and checked'
type: 'feature'
created: '2026-09-02'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: [oversized]
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The contract below governs.*

Some licences are not written as a single name. They are written as "this licence, plus a named
carve-out" — a standard, ordinary way to publish a typeface. Today the checker that decides whether a
font may ship with this project cannot read that shape at all. It sees the joining word, decides the
sentence is not something it understands, and refuses the file — even when both halves are things the
project is perfectly happy with. The refusal it prints says the licence could not be identified, which
is misleading: the licence is fine, the reading is not.

This story teaches the checker that shape, and requires **both** halves to be approved: the licence
must be one of the four the owner has accepted, and the carve-out must be one somebody has read and
approved as well. Neither half on its own is enough — a carve-out can take permissions away as easily
as grant them.

Done looks like this: a file offering an approved licence with an approved carve-out is accepted and
recorded as exactly what it says; anything else is refused with a message naming **which half** was
the problem.

And the first thing this story does is check whether any real typeface it is about to ship actually
uses this form — because it is possible that the rule, as written, accepts none of them, and that is
worth knowing before the typefaces are chosen rather than after.

<intent-contract>

## Intent

**Problem:** The licence classifier refuses every SPDX `<id> WITH <exception>` declaration as
unreadable syntax (measured at `771d82f`: `Apache-2.0 WITH LLVM-exception` →
`unsupported SPDX operator "WITH"`, and the text classifies `(unknown, "")`, so both asset gates refuse
it on their *"could not be classified"* arm). That check was untouched by Story 8.4j and is only newly
reachable. Story 8.5 procures 20+ typefaces next, so a face turned away on this technicality would
produce a refusal saying nothing about its licence being acceptable.

**Approach:** **This story does NOT come from `epics.md`, and `epics.md` MUST NOT be amended.** It was
created 2026-09-02 by ruling **D-8.4k.1**, implementing the owner's decision **D-8.5.14**. **It is NOT
an exception to Epic 8's insertion bound: that bound governs work the epic discovers about ITSELF
(endogenous); this is work the OWNER added (exogenous), which is unbounded but PLACED — before the
story that consumes it, never in the same commit as that story's population.** Teach the **single
existing enumerator** the `WITH` form, and admit a declaration **iff its base licence is one of the
owner's four ids AND its exception is on a named, owner-approved exception list. Neither half alone
admits.**

## Boundaries & Constraints

**Always:**
- **TASK 1 IS GATING AND RUNS BEFORE ANY PARSER WORK (D-8.4k.3).** Measure the domain first; record
  every `WITH` declaration found, its base id, its exception id, and whether D-8.5.14 admits it.
- `WITH` is enumerated by `licence.ClassifySPDXExpressionTerms`, the **sole** SPDX enumerator.
  **NEVER a second parser** — no `strings.Split` on an expression at a gate, in a marker, or in a test
  helper. This shape has been found five times in this area.
- Admission is **per part**: base ∈ `fontAssetLicenceAllowlist` **AND** exception ∈ the owner-approved
  exception list. The refusal **names the part that failed** — the exception when the exception failed,
  the base id when the base failed.
- The exception list is an **OWNER-DECIDED list, guarded exactly as the four ids are**: a test-owned
  literal citing D-8.5.14, so appending an entry reds **on its own message and nothing else**. Story
  8.4j's `TestFontAssetAllowlistIsTheOwnersFourIDs` shape is the template — **do not invent a second
  shape.**
- **Byte identity is absolute.** No `.folio` file, no engine change, nothing under `folio-go/fonts/`.
  **All 23 golden digests byte-identical — a moved golden is a HALT.** `maximumCacheAssets` stays `64`.
  `lint/MANIFEST.md` regenerates with **no diff**; a legitimately changed label is a **finding to
  route, not a re-record**.
- Existing resolution semantics are **behaviourally unchanged** for every expression that does not
  contain `WITH`: same family, same error, same label, at every caller.
- `licenceNames`, `licenceClauses` and `resolveLicenceSignals` are **unmodified**.
- The seven clause-classified BSD dependency licences stay `(permissive, "BSD-3-Clause")` (D-8.4i.9).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **THE DOMAIN IS EMPTY under D-8.5.14** — i.e. Task 1 finds no `WITH` declaration among the real
  candidate families whose base is one of the owner's four ids. **HALT with blocking condition
  `empty reachable domain under D-8.5.14` and report the measurement in full.** Frame it as *"the rule
  is implemented and admits nothing real; here are the declarations that exist and here is why each is
  refused."* **NEVER as a reinterpretation. NEVER resolved in the build. NEVER by widening the four
  ids.** (See Design Notes, *"The domain, measured at the plan gate"* — measured EMPTY there; if the
  build's own run disagrees with that record, **the disagreement is itself the halt**.)
- **The owner-approved exception list has no owner-named seed.** It starts **empty** (D-8.5.14: *"starts
  empty or near-empty and grows by owner decision"*). Seeding it is an owner decision and **must not be
  taken in the build**. This is the same escalation as the one above and travels with it.
- Any golden digest moves, `lint/MANIFEST.md` regenerates with a diff, or `maximumCacheAssets` ≠ 64.
- Either asset gate's **list** changes, or the **set of licences either gate admits** changes for any
  expression not containing `WITH` (D-8.4j.9's rebound guard).

**Never:**
- Never widen `fontAssetLicenceAllowlist`, and never point either gate at the other's list.
- Never accept on the base licence alone. The owner rejected that **explicitly**, because an exception
  can remove permissions as well as grant them. Do not re-litigate, soften, or reinterpret it.
- **`WITH` composed with `OR`/`AND` is OUT OF SCOPE and stays refused.** Measured at `771d82f`:
  `MIT OR (GPL-2.0-only WITH Font-exception-2.0)` → `unsupported SPDX expression`, **zero terms**;
  `OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` → `unsupported SPDX operator "WITH"`. Both stay
  refused after this story. The `WITH` form is admitted **only as the whole expression** (exactly one
  base, one exception, no other operator).
- **DW-133(b) does NOT ride this story** — `licencegraph.go:44` discarding the parse error is
  registered MEDIUM with the boundary-gate checklist as its owner.
- No new SPDX operator, no parenthesis support, no `+` suffix handling, no exception aliasing.
- Nothing else rides this story. One capability, one fix (D-8.4k.2, D-8.4j.7).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| RP1 — admit path | Font dir, `SPDX-License-Identifier: <owner id> WITH <approved exception>`, exception list containing that exception | Classified `(permissive, "<owner id> WITH <approved exception>")` — **labelled with the whole expression** — and **admitted** at SITE A | No error expected |
| RP2 — exception fails | `Apache-2.0 WITH LLVM-exception`, exception not on the list | Refused at SITE A **naming the exception** (`LLVM-exception`), not the base and not the whole label | Build fails, names the directory |
| RP3 — base fails | `ISC WITH <approved exception>` | Refused at SITE A **naming the base** (`ISC`), listing the owner's four ids | Build fails, names the directory |
| RP4 — neither half alone | Mutation making the base test OR the exception test unconditionally true | Exactly one named test reds per mutation (RP3 for the base half, RP2 for the exception half) | Mutation-proved by **deletion** of the conjunct |
| RP5 — list guard | Append any entry to the owner-approved exception list | The exception-list guard reds **on its own message**; no other test reds | n/a |
| Copyleft base | `GPL-2.0-only WITH Font-exception-2.0` | Refused **as copyleft**, on the copyleft arm, naming the expression | Build fails |
| Composed with `OR` | `OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` | Refused, `(unknown, "")`, **unchanged from `771d82f`** | Build fails on the *"could not be classified"* arm |
| Degenerate | `Apache-2.0 WITH`, `WITH LLVM-exception`, `A WITH B WITH C` | Refused, zero-or-partial terms, **not admissible**; `(unknown, "")` | Fail closed, unchanged |
| Wordlist gate (SITE B) | Any `WITH` declaration at `resolveWordlistAssetRow` | **Refused** — SITE B's exception list is empty by policy; its admitted set is byte-identical to `771d82f` | Build fails naming the part |
| Dependency / npm path | `MIT WITH Anything` reaching `ClassifySPDXExpression` | Refused, `FamilyUnknown` + error, **byte-identical to `771d82f`** | Build fails, unchanged |

</intent-contract>

## Code Map

- `lint/internal/licence/classify.go:51-79` — `ClassifySPDXExpressionTerms`, **the sole enumerator**.
  `parts := strings.Fields`; odd indices must be `AND`/`OR`, even indices are terms. `WITH` at an odd
  index returns `FamilyUnknown, terms-so-far, unsupported SPDX operator "WITH"`. **This is the one
  function that learns `WITH`.**
- `lint/internal/licence/classify.go:20-23` — `ClassifySPDXExpression`, the family-only wrapper. Its
  doc comment already states *why* it exists: **its two callers do not gate per term.** See Design
  Notes 2.
- `lint/internal/rules/licencegraph.go:78` and `lint/internal/licence/npm.go:64` — the wrapper's only
  non-test callers, both on the dependency/npm path. Neither can evaluate an exception.
- `lint/internal/licence/classify.go:53` — the parenthesis refusal; `:57` — the even-field-count
  refusal. **Both are the composition boundary and both stay.**
- `lint/internal/licence/classify.go:140-173` — `permissiveSPDX`; `:378-388` — `classifyBySPDX`, an
  **exact map lookup** (so a whole `WITH` expression must never be routed through it).
- `lint/internal/licence/licencesignals.go:313-365` — step (1), the SPDX-line loop. One line = one
  signal whatever its arity; dedup key is the whole expression; `unresolvedID` is set only when
  `len(terms) == 1 && terms[0] == expression`. **A `WITH` expression must keep naming nothing when it
  fails to resolve** — naming it would push both gates off the *"could not be classified"* arm onto an
  arm asserting the text classifies as something it does not (D-8.4j.14).
- `lint/internal/licence/licencesignals.go:437-441` — `markDeclaredTerms`. Must mark the **base id** of
  a `WITH` term (so an Apache body name signal does not double-count), never the exception.
- `lint/internal/licence/licencesignals.go:100-135` / `:159-174` / `:470-483` — `licenceNames`,
  `licenceClauses`, `resolveLicenceSignals`. **READ-ONLY. Do not touch.**
- `lint/internal/manifest/manifest.go:140` — `fontAssetLicenceAllowlist`, the owner's four ids
  (D-8.5.3). **READ-ONLY.** `:142-148` — `fontAssetLicenceAllowed`, the derived map.
- `lint/internal/manifest/manifest.go:215-230` — `firstTermNotOn`, the **shared mechanism** of both
  gates' per-term admission. Policy stays at the call site; the list is a parameter. Fails closed on
  `err != nil || len(terms) == 0`.
- `lint/internal/manifest/manifest.go:239-244` — `failingTermPhrase`, which renders *which* part a
  refusal objected to.
- `lint/internal/manifest/manifest.go:455-470` — **SITE A**, the font gate (four ids).
  `:586-602` — **SITE B**, the Thai wordlist gate (`licence.IsPermissiveSPDX`). **Two lists, one
  mechanism** (D-8.4j.9, D-8.5.13). Never collapse them.
- `lint/internal/manifest/manifest_test.go:875-965` — **THE ALLOWLIST GUARD TEMPLATE.** A test-owned
  literal of the owner's four ids citing D-8.5.3, with a length check, an ordered element check, and a
  map/slice-agreement check. Copy this shape for the exception list, citing **D-8.5.14**.
- `lint/internal/manifest/manifest_test.go:646-650` — precedent for exercising `firstTermNotOn` with an
  **injected** predicate. This is how RP1/RP3 are proved while the production exception list is empty.
- `lint/internal/licence/licencesignals_test.go` / `classify_test.go` — `TestLicenceSignalCensus`, the
  **35-entry hand-written pin** (D-8.4i.11). See Design Notes 5.
- `lint/testdata/licence/permissive/example.test/*/LICENSE` — the eight one-line SPDX fixtures. **All
  eight are bare single identifiers; none carries `WITH`.**
- `lint/MANIFEST.md` + its assets header — states the label/gate divergence for a reader who has the
  artifact but not the code. If the header gains a sentence about exceptions, it must regenerate with
  no diff afterwards.

## Tasks & Acceptance

**Execution:**

1. **`lint/internal/licence/` (probe, GATING, no production edit) — TASK 1: MEASURE THE DOMAIN.**
   Enumerate, from the real candidate families Story 8.5 will draw on, every `WITH` declaration that
   occurs; record base id, exception id, and whether D-8.5.14 admits it. Record the **current**
   verdict beside the **proposed** verdict for each. Commit the measurement as a record. **Run this
   before any parser work.** If the domain is empty → **HALT** per `Block If`. The plan gate already
   took this measurement (Design Notes 1); **if the build's own run disagrees with that record, the
   disagreement is itself the halt.**
2. `lint/internal/licence/classify.go` — teach `ClassifySPDXExpressionTerms` the `WITH` form. The
   property: an expression of exactly `<base> WITH <exception>` enumerates as **one term carrying two
   parts**; the family is the **base's** family; any other appearance of `WITH` (composed with
   `AND`/`OR`, repeated, missing either side) keeps today's refusal byte-for-byte. **The gate must
   receive base and exception as separate, enumerator-derived values** — a struct term is the obvious
   sketch, but *the property is what binds*: no consumer may split a term string. Non-`WITH` inputs
   must return the same family and the same error text as at `771d82f`.
3. `lint/internal/licence/classify.go` — `ClassifySPDXExpression` (family-only wrapper) **refuses any
   expression containing a `WITH` term**, returning exactly today's `FamilyUnknown` + error. Rationale
   in Design Notes 2. This keeps the dependency and npm paths byte-identical.
4. `lint/internal/licence/licencesignals.go` — `markDeclaredTerms` marks a `WITH` term's **base id**
   only. Re-run the masking falsifier (`MIT OR Apache-2.0` over a GPL body still refuses as copyleft)
   **by measurement, not inspection** (D-8.4j.11's correction 2). Do not touch `licenceNames`,
   `licenceClauses` or `resolveLicenceSignals`.
5. `lint/internal/manifest/manifest.go` — add `fontAssetLicenceExceptionAllowlist`, **EMPTY**, with a
   doc comment naming D-8.5.14 as its authority and stating that it grows **only** by owner decision.
   Extend `firstTermNotOn`'s predicate to receive the whole term (base + exception) so **policy stays
   at the call site**. SITE A: base on the four ids **AND** exception approved. **SITE B: exception
   list empty by policy** — a `WITH` term is refused there, keeping SITE B's admitted set identical to
   `771d82f`. Extend `failingTermPhrase` so the refusal says *which* part failed (an exception failure
   must not be reported as a "term").
6. `lint/internal/manifest/manifest_test.go` — the **exception-list guard**, on
   `TestFontAssetAllowlistIsTheOwnersFourIDs`' shape: a test-owned literal citing **D-8.5.14**,
   asserting the production list's exact contents and that the derived map is a view onto it.
   Appending an entry must red **this test and nothing else**.
7. `lint/internal/manifest/manifest_test.go` + `lint/internal/licence/classify_test.go` +
   `lint/testdata/` — the five red-proofs (I/O matrix RP1–RP5), each **mutation-proved by deletion**
   → exactly one named test reds on its own message. RP1 and RP3 exercise the mechanism with an
   **injected** exception list (`manifest_test.go:646` precedent), because the production list is
   empty; state that limitation in the test's own doc comment rather than claiming end-to-end proof.
   Add the degenerate and `OR`-composed cases as pins that must **not** move.
8. `lint/internal/licence/licencesignals_test.go` — re-run `TestLicenceSignalCensus` and state
   explicitly, in the report and in a comment, that **its 0-of-35 cannot witness this change** (Design
   Notes 5).

**Acceptance Criteria:**
- Given Task 1's measurement is complete and the domain is non-empty, when the build proceeds, then
  the measurement is committed as a record naming every declaration, its base, its exception and its
  verdict; and given the domain is empty, when Task 1 completes, then the build **HALTS** with
  `empty reachable domain under D-8.5.14` and the measurement in hand.
- Given a font directory whose licence declares an allowlisted base with an approved exception, when
  `ResolveAssets` runs, then the asset is admitted and its manifest label is the **whole expression**.
- Given a font directory whose licence declares an allowlisted base with an **unapproved** exception,
  when `ResolveAssets` runs, then the build fails with a message naming **the exception**.
- Given a font directory whose licence declares a **non-allowlisted** base with an approved exception,
  when `ResolveAssets` runs, then the build fails with a message naming **the base** and listing the
  owner's four ids.
- Given either admission conjunct is mutated to be unconditionally true, when the suite runs, then
  exactly one named test reds, on its own message.
- Given an entry is appended to the owner-approved exception list, when the suite runs, then the
  exception-list guard reds on its own message and no other test reds.
- Given any expression not containing `WITH`, when it is classified at any of the four callers, then
  its family, error text, label and admission verdict are **byte-identical to `771d82f`**.
- Given the full verification set runs, then all 23 golden digests are byte-identical,
  `lint/MANIFEST.md` regenerates with no diff (twice), `maximumCacheAssets` = 64, and the README md5 is
  `078d7d80d518d54af2fc04fb270d46b8`.

## Design Notes

**1. THE DOMAIN, MEASURED AT THE PLAN GATE (2026-09-02, baseline `771d82f`, tree clean, probes run
from `/Users/panitw/Projects/folio/lint` as untracked in-package `_test.go` files, deleted afterwards,
`git status --porcelain` empty).** Taken here as well as in Task 1 because a red here halts before a
design is committed to.

*Population.* **There is NO recorded candidate-family list anywhere in the repository.** Story 8.5's
AC1 is *route*-shaped, not family-shaped; its twenty-row procurement table is named only as
hypothetical future work. So the "real candidate families" can only be characterised by 8.5's own
recorded filters: **20+ families**, admitted by the four-id allowlist; **vendored static `.ttf`/`.otf`**
from upstream release archives; **Regular upright only**; **Latin and Thai**, no CJK. That is the
libre text-face population — SIL, Google Fonts, IBM, Adobe and equivalents.

*Declarations found in that population.* **Zero.** No file under `folio-designer/public/fonts/`,
`folio-go/fonts/`, `folio-go/testdata/`, `folio-designer/third-party-notices/` or `lint/testdata/`
contains a `WITH` expression; the eleven committed font directories carry OFL-1.1 (ten) and Apache-2.0
(one) as **prose markers**, and the only eight SPDX lines in the repo are bare single identifiers.
Permissive font licences do not carry SPDX exceptions: OFL 1.1 has no exception convention (its
Reserved Font Name mechanism is in-licence), and neither Apache-2.0, MIT nor Ubuntu-font-1.0 has one in
font use.

*The `WITH` declarations that DO occur in the font world, each with the verdict D-8.5.14 gives it:*

| Declaration | Base | Exception | Base ∈ four ids? | D-8.5.14 |
|---|---|---|---|---|
| `GPL-2.0-only WITH Font-exception-2.0` | `GPL-2.0-only` | `Font-exception-2.0` | **No** — copyleft | **REFUSED** (copyleft arm) |
| `GPL-2.0-or-later WITH Font-exception-2.0` | `GPL-2.0-or-later` | `Font-exception-2.0` | **No** — copyleft | **REFUSED** (copyleft arm) |
| `GPL-3.0-or-later WITH Font-exception-2.0` | `GPL-3.0-or-later` | `Font-exception-2.0` | **No** — copyleft | **REFUSED** (copyleft arm) |
| `OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` (Libertine-family dual form) | — | — | — | **REFUSED** — composed with `OR`, out of scope, unchanged from `771d82f` |
| `GPL-2.0-only WITH Liberation-font-exception` (Liberation v1; not a registered SPDX exception id) | `GPL-2.0-only` | — | **No** — copyleft | **REFUSED** (copyleft arm) |
| `Apache-2.0 WITH LLVM-exception` | `Apache-2.0` | `LLVM-exception` | **Yes** | **REFUSED** — exception not on the (empty) owner list; and this is an **LLVM-project** shape, not a typeface shape |

*The decisive fact.* **SPDX registers exactly one font-related exception, `Font-exception-2.0`, and it
is by construction an exception to the GNU GPL v2.** Its base can therefore never be one of the owner's
four ids. **So the reachable domain of D-8.5.14 for fonts is EMPTY.**

*The asymmetry, stated as D-8.4k.3 states it.* The owner rejected base-alone because *"an exception can
remove permissions as well as grant them."* True — **and the font exception is the canonical case of
the opposite.** It exists to let a **copyleft** font be embedded in a document **without copyleft
propagating to that document**, which is precisely Folio's mechanism and precisely what AD-26's stated
Prevents is about. **This is a report, not a reinterpretation.** The rule is implemented exactly as
ruled; it admits nothing real; the owner is shown that before twenty faces are procured around it.

*Current verdicts, measured (`ClassifySPDXExpressionTerms` → family / terms / error):*
`Apache-2.0 WITH LLVM-exception` → unknown / `["Apache-2.0"]` / `unsupported SPDX operator "WITH"`;
`GPL-2.0-only WITH Font-exception-2.0` → unknown / `["GPL-2.0-only"]` / same;
`OFL-1.1 OR GPL-2.0-or-later WITH Font-exception-2.0` → unknown / `["OFL-1.1" "GPL-2.0-or-later"]` / same;
`MIT OR (GPL-2.0-only WITH Font-exception-2.0)` → unknown / `[]` / `unsupported SPDX expression`;
`Apache-2.0 WITH` and `WITH LLVM-exception` → unknown / `[]` / `malformed SPDX expression`.
Every one classifies `(unknown, "")` through `ClassifyLicenceText`, i.e. lands on both gates' arm 1.

**2. WHY THE FAMILY-ONLY WRAPPER REFUSES `WITH`.** `ClassifySPDXExpression`'s own doc comment records
that it exists *because both of its callers do not gate per term*. `WITH` is the first form whose
admissibility is **not** determined by the family verdict alone: a consumer that cannot test the
exception half must not be handed a permissive verdict for one, or `MIT WITH <unevaluated carve-out>`
would newly pass the dependency and npm paths — **precisely the "carve-outs nobody has evaluated" the
owner rejected**, on a population that grows by `go get` and version bumps. Refusing at the wrapper
keeps those two paths **byte-identical to `771d82f`** and puts the exception policy in exactly one
place. It is not a second parser and not a second policy; it is the wrapper declining to answer a
question it cannot answer.

**3. THE EXCEPTION LIST STARTS EMPTY, AND THAT MAKES TWO RED-PROOFS MECHANISM-LEVEL.** D-8.5.14: the
list *"starts empty or near-empty and grows by owner decision."* Nothing in the record names a seed, and
seeding it in the build would be taking an owner decision. **Consequence, stated rather than
discovered:** with an empty right half, every `WITH` declaration at SITE A fails on the exception,
so the base half is **unobservable end-to-end**. RP1 (admit) and RP3 (base fails) are therefore proved
against the mechanism with an **injected** exception list — the `manifest_test.go:646` precedent — and
each test says so in its own doc comment. **A red-proof whose isolation is imperfect and says so is
worth more than one that overstates** (D-8.4j.17/18). RP2, RP4's exception half and RP5 are end-to-end.
This is the second half of the owner escalation and travels with the first.

**4. PRECEDENT FOR A LIST MEMBER WITH NO POPULATION.** `Ubuntu-font-1.0` was added at Story 8.4h with
nothing in the repository licensed under it, proved *"by the classifier table, by the SPDX-line fixture
module … and by nothing else"*. DW-115 is still open on it, owned by Story 8.5. The exception list is
the same shape one step further: a rule with no population at all today.

**5. THE CENSUS CANNOT WITNESS THIS CHANGE.** `TestLicenceSignalCensus`' 35-entry pin is a
hand-written literal derived from nothing under test (D-8.4i.11). **No population text carries a `WITH`
declaration** (measured above), so corrupting the new branch leaves the census green. **Its 0-of-35 is
evidence the story breaks nothing, never that it does anything.** Report it in those terms. **The
witnesses are the five red-proofs.**

**6. MUTATION PROOF IS BY DELETION, NOT SUBSTITUTION.** Deleting the base conjunct must red RP3 alone;
deleting the exception conjunct must red RP2 alone. Falsifying a condition proves arm order; deleting
the conjunct proves it is reached. Where a deletion cannot isolate, **condition** rather than delete,
and record the mutation used (D-8.4j.18).

**7. THE LABEL/GATE DIVERGENCE EXTENDS.** The manifest label states **what the file says**; the gate
states **whether every option the file offers is acceptable**. A row reading `Apache-2.0 WITH
LLVM-exception` beside a four-id allowlist is not a bug and must not be "fixed" by widening the list or
shortening the label. If `MANIFEST.md`'s assets header gains a sentence saying so, regenerate and
confirm no diff.

## Verification

**Commands** (each with its **working directory** — D-8.4j.8; an invocation without one is a claim,
not a measurement). Cadence is **per-epic** (D-000.4), so the four `FOLIO_MATRIX_TARGET` legs,
`TestCrossTargetByteIdentity` and Playwright are **excluded** and owed to Epic 8's boundary gate.

- **from `/Users/panitw/Projects/folio`**: `gofmt -l folio-go lint` — expected: **exactly one** path,
  `lint/internal/rules/licencegraph_test.go` (DW-116, **do not reformat**). Run from the repo root:
  from `lint/` it prints `lstat` errors that read as clean.
- **from `/Users/panitw/Projects/folio/lint`**: `go test ./... -count=1` — `-count=1` is mandatory; a
  cached `ok` hides red tests. Expected: all packages pass; report pass/fail counts per package.
- **from `/Users/panitw/Projects/folio/lint`**: `go vet ./...` — clean.
- **from `/Users/panitw/Projects/folio/lint`**: `go run ./cmd/genmanifest` (per `RELEASING.md`; there
  is no root `go.mod`, so a repo-root invocation does not execute at all), then **from
  `/Users/panitw/Projects/folio`**: `git diff --exit-code -- lint/MANIFEST.md`. **Run the pair twice**
  — expected: exit 0 both times.
- **from `/Users/panitw/Projects/folio/folio-go`**: `go test ./...` — expected **1815 pass / 2 fail /
  5 skip**, the two failures being exactly `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)`
  subtest (mandated permanent red — never "fix" it).
- **from `/Users/panitw/Projects/folio/folio-go`**: `go vet ./...` and `go vet -tags=matrix ./...` — clean.
- **from `/Users/panitw/Projects/folio/folio-designer`**: `npm run typecheck`, `npm run lint`
  (expected **exactly 4** `only-export-components` warnings), `npm test` (expected **40 files / 411
  tests**), `npm run build`, `npm run test:e2e:compile`.
- **from `/Users/panitw/Projects/folio`**: the **23-digest golden witness** — expected byte-identical;
  README md5 `078d7d80d518d54af2fc04fb270d46b8`; `maximumCacheAssets` = 64.

**Manual checks:**
- Re-derive Task 1's domain measurement independently of the plan-gate record above and diff the two.
  A disagreement is a halt, not a reconciliation.
- Confirm `licenceNames`, `licenceClauses` and `resolveLicenceSignals` are untouched (`git diff`).
- Confirm the seven clause-classified BSD dependency licences still classify
  `(permissive, "BSD-3-Clause")`.

## Auto Run Result

Status: ready-for-dev
Blocking condition: none (halt-after-planning directive honoured; the spec's `Block If` carries the
designed halt `empty reachable domain under D-8.5.14`, which the plan-gate measurement in Design
Notes 1 already reports as TRIGGERED — route to the owner before dispatching the build)
Baseline commit: 771d82f0d31608d8e036c0d6cd08651baac67e12
Commits this dispatch: none (plan only; no code written)
