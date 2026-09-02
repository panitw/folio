---
title: 'Story 16.1b: The binary is asked what it says about itself'
type: 'feature'
created: '2026-09-03'
status: 'ready-for-dev'
review_loop_iteration: 0
followup_review_recommended: false
baseline_commit: '384c8ac'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: []
deferred: []
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

A typeface put into a document carries a statement about whose it is and on what terms it may be passed
on. Until now that statement was checked when the product was built, against twenty-one typefaces
somebody had looked at. Once typefaces start arriving from the web, the same statement has to be checked
on the author's machine, against files nobody has reviewed.

This story adds that check where it cannot be bypassed: in the engine, on the bytes themselves, at the
moment a typeface is put into a document. It asks the file what *it* says about its own licence, and
compares that with the licence being claimed on its behalf.

The important part is what it does when the file says nothing. It admits it. The problem this guards
against is a typeface travelling under **someone else's** terms — a false statement, not a missing one —
and that really happened here once, to seventeen of twenty-one typefaces, unnoticed until review.
Refusing files that are merely silent would have blocked about a sixth of the library while catching
none of that. So: if the file contradicts the claim, it is refused and told why. If it agrees, or says
nothing readable, it is accepted.

It also gains something the old check never had: a file whose own bytes name a share-alike or GPL-family
licence is refused outright, whatever anyone claims on its behalf.

Done looks like: no document can be written carrying a typeface whose own bytes contradict the terms
recorded beside it.

<intent-contract>

## Intent

**Problem:** Epic 16 moves the licence admission that D-8.5.2/D-8.5.3 made a **build gate** into a
**runtime** decision over unreviewed bytes. The build gate's teeth are the **nameID 13 tie** at
`folio-designer/src/font-catalogue.test.ts:355-366`, which binds a face's declared SPDX id to the
binary's own licence description — *"the one statement of a face's licence that cannot be edited from
outside the binary."* Nothing carries that tie to runtime, so without this story Epic 16 replaces a gate
with something **strictly weaker**, on exactly D-8.6.5's axis: 17 of 21 catalogue faces once shipped
under another project's licence, green, until a review caught it.

**Approach:** One byte-taking door in `internal/fontset`, beside `RefuseVariableFace`, called from
`embedFontFamily`. It **re-reads the name table from the bytes in hand** and returns one of three
outcomes (D-16.R.7): **contradiction refuses, confirmation admits, no evidence admits.** It also carries
a **refuse-signature** half checked against every face regardless of what it declares, closing an AD-26
hole the build gate never had.

**Sequence: this story runs BEFORE Story 16.1** (D-16.R.8). It needs no browser and no fetched face —
it tests against the 21 committed faces and its own fixtures — while 16.1 is the first story that puts a
**fetched** face into a document. Building the gate after the population it polices arrives is how
D-8.6.5 shipped green.

## Boundaries & Constraints

**Always:**
- **Three outcomes, never two** (D-16.R.7), mirroring D-16.R.4's token table one level up:
  - **CONTRADICTION** — the name table matches a **refuse-signature**, or matches a **different**
    admitted licence's signature than the one declared → **REFUSE**, located, naming both what was
    declared and what the bytes say.
  - **CONFIRMATION** — matches the declared licence's signature → admit.
  - **NO EVIDENCE** — no signature matches, or the name table is absent or unparseable → **admit.**
- **Silence admits, and the reasoning is recorded in the code**, because it will look like a hole to the
  next reader: the threat is a face travelling under **another project's** terms, which is a *false*
  statement, not a missing one. Refusing silence catches none of that and costs ~17% of the library —
  **measured**: 50 of 100 sampled upstream faces refused under the original contract, and all three
  static `ufl/` families carry **no nameID 13 at all**, stating their terms in nameID 0.
- **The table is TWO-SIDED.** Admit-signatures keyed by SPDX id: `OFL-1.1` → `/SIL Open Font License/i`;
  `Ubuntu-font-1.0` → `/Ubuntu Font Licence/i`; `Apache-2.0` → `/Apache License,?\s+Version 2\.0/i`.
  **Refuse-signatures, checked against every face whatever it declares:** GPL, LGPL, AGPL, SSPL, and
  CC BY-SA / ShareAlike. Ground: **AD-26 Binds: all**, and its Prevents is copyleft arriving through a
  plausible-looking package. **A face whose own bytes name a copyleft licence is refused even when the
  declared licence is `OFL-1.1`.**
- **nameID 0 is consulted ONLY when nameID 13 is ABSENT.** If 13 is present it alone decides —
  otherwise a face whose 13 says GPL could be laundered by a permissive-sounding 0, defeating the
  contradiction check with the very thing it exists to catch.
- **Go re-reads the bytes. It may NEVER compare the wire `copyright` against the wire `licence`.** Both
  arrive from the browser, so a check over them proves nothing: both sides would move together. The tie
  reads the name table of the bytes in hand.
- **Signature matching is order-deterministic** — an ordered slice, never a map range (AD-1).
- **The Go and TS tables are a MIRROR CONTRACT, enforced by a test and not by a comment.** The Go table
  is the authority for documents and strictly subsumes the TS build-time table's population; something
  must fail when the TS table admits an id the Go table refuses.
- **Untyped `fmt.Errorf`, per `variableface.go`.** No new error type in package `folio` root —
  `TestFolioMethodNamesAreInjective` forbids it.
- **Both guards kept.** The build-time tie in `font-catalogue.test.ts` is not deleted or weakened
  because Go now checks it. **Both, or halt.**
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **The guard would ship without being observed to fire.** A guard that never refuses anything on the
  corpus has shipped vacuous. It requires a **positive control**: a fixture face under
  `folio-go/testdata/` whose name table contradicts its declared id, asserted refused.
- **Any refusal on the sample is a SILENCE rather than a CONTRADICTION.** That is the ship criterion,
  and it is semantic rather than a number, because a number here is a licence to tune. A refusal
  traceable to *"no signature matched"* is a **table gap** and returns to the engineering lead as a
  finding; one traceable to *"the bytes say something else"* is the guard working.
- **The check would be sited where it can be bypassed.** The browser is not the only door;
  `embedFontFamily` is reachable from `wasm.Engine.Apply` and from a hand-authored command.
- **`SupportedMajor` would move**, or any of the 23 golden digests moves.

**Never:** a check over the wire fields · a map range over signatures · softening a refusal to a
warning · deleting the build-time tie · a new error type in package `folio` root.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Declared `OFL-1.1`, nameID 13 carries the SIL sentence | a shipped catalogue face | **CONFIRMATION** → admit | — |
| Declared `OFL-1.1`, nameID 13 carries the Apache sentence | mislabelled face | **CONTRADICTION** → refuse, naming both | Located refusal |
| Declared anything, nameID 13 names GPL/AGPL/SSPL or ShareAlike | copyleft face | **REFUSE**, regardless of the declared id | Located refusal |
| No nameID 13, nameID 0 carries the Ubuntu sentence | the three static `ufl/` families | nameID 0 consulted; **CONFIRMATION** → admit | — |
| nameID 13 present and mismatched, nameID 0 permissive | laundering attempt | **REFUSE** — 0 is not consulted when 13 is present | Located refusal |
| No name table at all, or unparseable | face the parser cannot read | **NO EVIDENCE** → admit; `checkSfnt`/`DecodeFontForRender` refuses genuinely bad bytes one step later | — |
| Declared id with no signature entry | an admitted SPDX id the table does not cover | **NO EVIDENCE** → admit, and the gap is reported as a finding | Finding, not a refusal |
| Sentence in a non-Latin script | `ofl/wdxllubrifonttc` states OFL 1.1 in Traditional Chinese | **NO EVIDENCE** → admit. No ASCII regex reaches under this floor | — |
| Local-tier face | one of the 21 committed | Passes unchanged; the build-time tie still covers it | — |

</intent-contract>

## Code Map

- `folio-go/internal/fontset/variableface.go:69` — `RefuseVariableFace`, **the sibling to copy**: a
  byte-taking door at rank 6, untyped `fmt.Errorf`, called from `component_commands.go:2414`. Its
  `nil`-on-unparsable behaviour is **correct and is now the precedent this guard follows** (D-16.R.7
  dissolved DW-150 on exactly that reading).
- `folio-go/component_commands.go:2414` — where `RefuseVariableFace` is called from `embedFontFamily`.
  **The new call goes beside it**, keeping `component_commands.go` a caller rather than a checker.
- `folio-go/component_commands.go:2360-2466` — `embedFontFamily` in full, with `componentFields(raw, 12)`
  and `embeddedFontRecord`. **Arity is frozen**: `copyright` is one of the twelve wire fields, which is
  why the nameID 0 *reader* lives in the browser (Story 16.1) and this guard re-reads the bytes for its
  own, different question.
- `folio-go/component_commands.go:668` — the 6,288,384-byte face cap.
- `folio-designer/src/font-catalogue.test.ts:197-200` — **the build-time signature table, two rows**, and
  `:355-366` the tie itself with its stated rationale. This is the mirror the Go table must subsume, and
  it is not weakened here.
- `folio-go/internal/template/fontasset.go` — `DecodeFontForRender`/`checkSfnt`, the step that refuses
  genuinely unreadable bytes, which is why NO EVIDENCE is safe.
- `folio-go/testdata/` — where the positive-control fixture lands.

## Tasks & Acceptance

**Execution:**
- `folio-go/internal/fontset/` — a new byte-taking door beside `RefuseVariableFace`: parses the name
  table, applies the refuse-signatures to every face, then the admit-signature for the declared id, and
  returns contradiction / confirmation / no-evidence. Ordered slice, untyped error.
- `folio-go/component_commands.go` — call it from `embedFontFamily` beside the `fvar` refusal, before
  any byte reaches `t.doc.Assets`.
- `folio-go/testdata/` — the **positive control**: a fixture face whose name table contradicts its
  declared id, asserted refused. Plus the disequality control the build-time tie already uses.
- Tests: each of the three outcomes; the copyleft refuse-signature firing against a face declaring
  `OFL-1.1`; nameID 0 consulted only on absence; a face with a correct description still embedding
  (so the guard cannot be over-broad); **red-prove by deleting the guard**, not by falsifying a
  condition.
- A test enforcing the **mirror contract** between the Go and TS tables.
- **Re-run the 100-face sample** and report it: every refusal must be a contradiction and none a
  silence.
- `deferred-work.md` — **close DW-150 as reconciled**, recording that the lead's contract was what was
  out of step, not `RefuseVariableFace`.
- Record the narrowing on **D-000.15's running list** for Story 15.3 — this narrows an
  exported-API-reachable command and is in the **before-the-tag set**.

**Acceptance Criteria:**
- Given a face whose name table contradicts its declared licence, when it is embedded, then the command
  refuses it, located, naming both what was declared and what the bytes say — and no byte reaches
  `Assets`.
- Given a face whose own bytes name a GPL-family or ShareAlike licence, when it is embedded, then it is
  refused **whatever** licence is declared for it.
- Given a face with no readable licence statement, when it is embedded, then it is **admitted**, and the
  reasoning is stated in the code so a later reader does not read it as a hole.
- Given a face with no nameID 13 but a licence statement in nameID 0, when it is embedded, then nameID 0
  is consulted; and given one with nameID 13 present, then nameID 0 is not consulted at all.
- Given the guard, when it is proposed for merge, then a positive-control fixture proves it fires, and
  the re-run sample shows every refusal is a contradiction and none a silence.
- Given the TS build-time table, when it admits an id the Go table refuses, then a test fails.
- Given the build-time tie, when this story lands, then it is unchanged and still green.

## Design Notes

**Why silence admits, written here because it is the part that will look wrong.** The guard exists for
D-8.6.5 — a face travelling under another project's terms. That is a **false** statement. A face that
says nothing has made no statement to be false, and the decode path refuses it a step later if the bytes
are actually bad. The original contract refused silence and was measured at **50 of 100 refused**, of
which the great majority were silences: a guard that loud is one somebody eventually turns off.

**Why the refuse-signatures apply to every face.** AD-26's Binds is *all*, and its Prevents is copyleft
arriving through a plausible-looking package. A declared-id-only check would let a GPL face in under an
`OFL` token. This half is new — the build gate never had it.

**Why the nameID 0 widening cannot be a general fallback.** If 0 were consulted whenever 13 failed to
match, a face whose 13 says GPL could be laundered by a permissive 0. Absence is a different condition
from disagreement, and only absence opens the second door.

## Verification

- `cd folio-go && go test ./... && go vet ./... && gofmt -l folio-go`
- `cd folio-designer && npm run test && npm run build`
- The **positive control** reds when the guard is deleted; the guard admits a correct face.
- The **re-run 100-face sample**, reported with command, commit, tree state and working directory.
- The 23 golden digests unmoved; `SupportedMajor` still 2.
- Heavy suites run at the end-of-run catch-up per D-16.R.1; e2e specs compile.
