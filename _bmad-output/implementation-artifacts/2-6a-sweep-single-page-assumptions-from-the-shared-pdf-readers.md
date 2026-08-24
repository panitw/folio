---
baseline_commit: 804eea1
---

# Story 2.6a — Sweep single-page assumptions from the shared PDF readers

**Epic**: 2 — Text, shaping and page composition
**Sprint status key**: `2-6a-sweep-single-page-assumptions-from-the-shared-pdf-readers`
**Status**: `done`
**Inserted by**: D-2.6.9 (between Story 2.6 and Story 2.7)
**Baseline commit**: `d1af588` — *"Insert Story 2.6a per D-2.6.9"*
**Predecessor**: Story 2.6, committed at `4001354`

---

## FIRST — the sign-off measurement D-000.41 requires, performed at creation

**Two of the three pinned digests have an owner sign-off pending against them. Moving either silently
invalidates a review the owner has already been asked to perform, so this is stated before anything
else.**

Re-computed at `d1af588` via `rtk proxy shasum -a 256 … > file` (D-000.12 as corrected — never through
the wrapper's pipes), not inherited from the brief:

| artifact | digest at `d1af588` | sign-off state |
|---|---|---|
| `fixtures/shaped-text/expected.pdf` | `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370` | **Thai READING sign-off pending** (D-2.3.5) |
| `fixtures/expected-breaks/expected_breaks.json` | `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de` | **Thai BREAK sign-off pending** (D-2.4.3) |
| `fixtures/multi-page/expected.pdf` | `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b` | recorded by Story 2.6 |

All three match the brief exactly.

**Why the answer here is structural rather than probabilistic.** This story edits **test files only**.
Every artifact above is produced by the renderer, and the renderer is not in this story's scope at all —
see the scope fence below, which makes the claim machine-checkable rather than aspirational. A helper
that reads a PDF back cannot change the PDF it reads. **The correct expectation is not "these three do
not move"; it is that no emitted byte anywhere in the repository moves**, and that is AC1.

**Both pending sign-offs may stand.** Nothing this story is permitted to touch can reach either artifact.

---

## Baseline, measured at creation

HEAD is **`d1af588`** on `main`. Working tree **clean** (`git status`: *nothing to commit, working tree
clean*). Every number below was measured at this commit; none is inherited.

**Full ordinary suite** — `rtk proxy go test -count=1 -v ./...`, output captured to a file:

| measure | value |
|---|---|
| all-occurrences PASS (`^ *--- PASS`) | **557** |
| all-occurrences FAIL (`^ *--- FAIL`) | **1** |
| top-level PASS (`^--- PASS`) | **338** |
| top-level FAIL (`^--- FAIL`) | **1** |

**Exactly one expected failure**, and it is not this story's: `TestCorpusMeetsP6ExerciseFloors`, the
intentional Story 2.1 red, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. These figures
reproduce the brief's measurement at `4001354` exactly, which is expected — `d1af588` adds one line to
`sprint-status.yaml` and no Go.

**Scope note on the numbers (D-000.14, extended).** These are scoped figures: *whole module, no build
tags, `-count=1`, all-occurrences and top-level reported separately*. Any figure this story reports
later must carry the same three-part scope, or it is not comparable.

---

## In plain terms (read this first if you just want the gist)

The engine only recently learned to produce a document with more than one page. Everything before that
produced exactly one, and for years that was not a limitation anyone had to think about — it was simply
what a document was.

The problem is that the checks we wrote to verify our own output were written during that period. Many
of them quietly assume there is one of whatever they are looking at. One reader, for example, used to
find the part of a document that holds the text and stop as soon as it found one — which was a perfectly
correct thing to do when there could only be one. Point it at a two-page document and it would examine
one of the two pages and report what it found as though it were the whole document, with no complaint
and no failure. That silence was the danger this story exists to remove, and this is the story that found
and fixed the worst instances of it, plus a second, unrelated well-formedness check that read only part
of a document's internal index and stopped, in a place a reader would never see.

That is the intended outcome, and it landed — but it did not land cleanly on the first pass, and a
reviewer caught two serious problems before this work was allowed to close, which is worth stating
plainly rather than quietly folding away. First, the fix to the main reader shipped without any check
that could catch someone accidentally undoing it later — the only evidence it had ever worked was a
one-off measurement that was written down and then thrown away. Second, and worse: in the course of
tightening one of the other checks, its fix accidentally introduced a new way for that same check to pass
when it should not have — the exact class of silent failure this whole effort was meant to eliminate,
reappearing inside the effort itself. Both were caught, both are now fixed, and both now have a permanent
check that would catch either mistake if it ever happened again — built from small hand-written examples
rather than real documents, since the real documents this engine currently produces are not varied enough
to expose either problem on their own.

One thing will look unfinished, deliberately, and it is worth saying so rather than letting a later
reader stumble on it and wonder: one very small, cosmetic risk noted early on was deliberately left alone,
because fixing it was optional and record-keeping about it matters more than the fix itself. It is written
down by name rather than fixed, which was always the allowed outcome for it.

---

## Story

**As** the folio engine's maintainer,
**I want** every shared PDF-reading test helper to behave correctly on a document with more than one of
anything the engine can now emit more than one of,
**so that** the checks that guard Story 2.7's `Page X of Y` — which adds a text run to **every page** —
are measuring the whole document rather than an arbitrary page of it.

---

## Do not re-open — settled rulings this story inherits

Apply these. Do not re-litigate them.

| ruling | what it binds here |
|---|---|
| **D-2.6.9** | This story exists, sits between 2.6 and 2.7, and its predicate is **behavioural, never a grep**. Enumeration-and-count is its **first** task. |
| **D-000.53** | A golden is not accepted until a reader we did not write resolves it into the objects it claims to contain. The three defects this story generalises are its evidence. |
| **D-000.50** | Before writing a guard, ask whether **any subject can express the defect**. If none can, it is a forward guard and is **not** coverage. |
| **D-000.24** | *"Forward guard with no available red-proof"* is a named, permitted category. Use the name; do not disguise a forward guard as a proven one. |
| **D-000.14** | When a count is load-bearing, count by AST; a text or filtered count **measures the filter**. This story's own inherited inventory is an instance — see the correction below. |
| **D-000.17** | A floor that is not met is **reported as unmet. It is never filled.** The enumeration's remainder is named instance by instance. |
| **D-000.34** (and its extension) | Repairing a helper may **kill a detector**, including a negative control. Re-point such a site rather than re-run it. |
| **D-000.42** | *"Redundant"* is a third guard category beside proven and forward. Count what the sweep **fixes**, not what it inspects. |
| **D-000.48** | A correction sweep can introduce a fresh instance of the class it corrects. This one corrects **silent truncation**, so every repaired helper is verified against the multi-page subject **mechanically**. |
| **D-2.6.6** | The Epic 2 gate owes **exactly five**. This story adds no sixth. |
| **D-000.4** | Heavy-test cadence is per-epic. Unit tests always. |
| **D-000.54** | *Only if* a new `matrixDocuments` entry is registered: run that document's native leg once before `review`. Not expected here. |
| **D-000.55** | `go vet -tags matrix` is a **compile gate, not an honesty check**. The phrase *"written, compiled and vetted, deliberately not run"* is **banned**. |
| **D-000.12** (corrected) | Verify bytes and hashes with `rtk proxy <cmd>` redirected to a file, never through the wrapper's pipes — and measure the mitigation itself. |

---

## Corrections to the inherited inventory — measured at `d1af588`

D-2.6.9 supplied its inventory explicitly labelled *"the sweep's starting inventory, not its answer"*
and *"text-matched"*. It was right to. **The inventory is wrong, and it is wrong in exactly the way
D-000.14 predicts.**

### The inherited figure of 12 `readEmittedRuns` call sites is 11. `matrix_test.go` has 3, not 4.

A substring search for `readEmittedRuns` also matches **`readEmittedRunsAllPages`** — the multi-page
sibling Story 2.6's finisher added. Its one invocation, `matrix_test.go:689`, was counted as a
`readEmittedRuns` call site. It is not one; it is the *fix*.

Counted by a means that distinguishes the two identifiers (`[^A-Za-z]readEmittedRuns\(`, excluding the
declaration):

| file | invocations | lines |
|---|---|---|
| `shaped_fixture_test.go` | **5** | 534, 776, 854, 905, 1073 |
| `matrix_test.go` | **3** | 782, 829, 874 |
| `wrapped_text_fixture_test.go` | **1** | 79 |
| `segment_origin_test.go` | **1** | 95 |
| `first_baseline_acceptance_test.go` | **1** | 291 |
| **total** | **11** | across **5** files |

**The story leads with this deliberately.** The sweep's own opening measurement is an instance of the
class of error the sweep exists to correct — a count that measured its filter. It is the cheapest
possible demonstration that AC2 is not ceremony.

### The in-code comment already disagrees with itself

`shaped_fixture_test.go:379–395` — the handoff comment Story 2.6's finisher wrote — says both *"sweep
this function's 12 call sites"* **and** *"the eleven callers that do not need it"*, in one comment block.
The second figure is the correct one. That comment is a work item for this story, not a citation.

### `matrix_test.go` is behind `//go:build matrix`

**The only multi-page-aware reader in the module is unreachable from the ordinary suite.** Measured: an
untagged test file calling `readEmittedRunsAllPages` fails to compile —

```
./zz_probe.go:28:9: undefined: readEmittedRunsAllPages
FAIL	github.com/panitw/folio/folio-go [build failed]
```

So a repair that consists of *"call the sibling instead"* is not available to 11 of the 11 untagged call
sites. This is scope the inherited framing does not mention, and it shapes AC5.

---

## Measured findings — read all of these before writing code

Every finding below was produced by **running** the helper against a subject, in a throwaway test file
that was deleted afterwards. The working tree is clean at `d1af588`. None was produced by inspection —
which is the whole point (D-2.6.9).

### 1. THE RED-PROOF — `readEmittedRuns` does not truncate to the first page. It truncates to an **arbitrary** page.

Running the untouched helper against `fixtures/multi-page/expected.pdf`, **20 times inside one process**:

```
readEmittedRuns over 20 calls, run-count histogram = map[9:4 24:16]
```

The document carries **33** runs across its two content streams (24 + 9). The helper returns **9 on some
calls and 24 on others, within a single process run** — because `pdfObjects` returns a `map[int][]byte`
and Go randomises map iteration order, so *"the first text-bearing stream"* is not a defined thing.

This is worse than the inherited framing (*"reads the first stream and breaks"*) in two ways:

- it is **nondeterministic**, so a test built on it can pass on one run and fail on the next with no
  code change at all; and
- it is **silent** — no fatal, no warning, a plausible non-empty slice of runs every time.

The helper's own comment predicts truncation and calls it *"NOT SAFE"*. The measurement upgrades that to
*non-deterministically unsafe*. **This is the story's red-proof and it is captured at baseline before any
fix exists (D-000.30).**

### 2. A second instance, same class, in a different helper — `resourceType0Objects`

`shaped_fixture_test.go:120` walks the object map, takes **the first object containing `/Type /Page `
that has a `/Font <<`**, and `break`s. On a two-page document it reads **one page's** font resource
dictionary and discards the other, silently, and again over a randomised map.

Every `/ToUnicode` and `/CIDToGIDMap` assertion in the module reaches the document through this
function. If a future document used a face on page 2 that page 1 does not — precisely what the fallback
chain exists to do — those assertions would become vacuous or fatal at random.

**But no committed subject can express it.** Measured on the multi-page golden:

```
page object  8: << /Type /Page … /Resources << /Font << /NotoSans 3 0 R >> >> /Contents  9 0 R >>
page object 10: << /Type /Page … /Resources << /Font << /NotoSans 3 0 R >> >> /Contents 11 0 R >>
resourceType0Objects over 20 calls: map[NotoSans,:20]
```

Both pages carry a **byte-identical** `/Font` dictionary, so the truncated answer and the correct answer
are the same value. **D-000.50 applies in full**: the repair is real and cheap, and its guard is a
**forward guard (D-000.24)** that does **not** count as coverage. Building a subject that could express
it would mean recording new emitted bytes, which the scope fence forbids — so the guard stays forward,
by ruling, not by preference.

It also matches on `"/Type /Page "` **with a trailing space** — the exact hazard `countPageObjects` was
repaired for (Nit 26). Note it; do not necessarily fix it here.

### 3. A third instance, in the well-formedness checker itself — `assertXrefEntriesPointAtTheirObjects`

`render_test.go:844` parses **one** xref subsection header, requires `start == 0`, and walks `count`
entries. A second subsection is **never examined, and its absence produces no failure** — the loop simply
ends. Same class: *"one of something" encoded as an invariant.*

**No subject can express it**: every committed golden has exactly **1** xref subsection, and the
serializer emits exactly one by construction. **Forward guard, or redundant** (D-000.42) if the
already-present offset check is judged to carry the property — the story states which, with the reason.

### 4. The subject population, measured — which quantities have a subject with N > 1 at all

This is D-000.50's pre-flight question answered with numbers rather than taste. Measured across every
committed golden:

| fixture | pages | streams | text streams | embedded faces | image XObjects | xref subsections |
|---|---|---|---|---|---|---|
| `minimal-rect` | 1 | 1 | 0 | 0 | 0 | 1 |
| `font-text` | 1 | 3 | 1 | 1 | 0 | 1 |
| `image-embed` | 1 | 2 | 0 | 0 | **1** | 1 |
| `three-band-page` | 1 | 3 | 1 | 1 | 0 | 1 |
| `multi-script-fallback` | 1 | 7 | 1 | **3** | 0 | 1 |
| `shaped-text` | 1 | 8 | 1 | **3** | 0 | 1 |
| `wrapped-text` | 1 | 8 | 1 | **3** | 0 | 1 |
| **`multi-page`** | **2** | 4 | **2** | 1 | 0 | 1 |

Read off it directly:

- **pages** and **text content streams**: only `multi-page` has N > 1. It is the sole subject, and every
  behavioural probe in this story runs against it.
- **embedded faces**: three fixtures already carry 3. Any face-singularity assumption would **already**
  be failing today, so this axis is exercised and is not where the defects are.
- **image XObjects**: the maximum anywhere is **1**. **No subject can express** a multi-image assumption.
- **xref subsections**: **1** everywhere, by construction. **No subject can express** it.
- **page-header runs / page-footer runs**: N = 2 each only in `multi-page`, and they are reached through
  the same content-stream reader as everything else — so they are covered by finding 1's repair rather
  than being a separate axis.

### 5. Helpers that behave correctly on the multi-page subject — recorded so they are not re-swept

Measured against `fixtures/multi-page/expected.pdf`:

| helper | observed | verdict |
|---|---|---|
| `countPageObjects` | returns **2** | correct |
| `pdfObjects` | returns **11** objects | correct |
| `streamBody` | **4** of 11 objects carry a stream — all four found | correct |
| `assertWellFormedPDF` | parameterised on page count by Story 2.6; called at `multi_page_fixture_test.go:683,693` with `mpTotalPages` | already repaired |
| `parseContentStreamRuns` | parses **one** stream by contract, correctly | correct **by design** — see AC6 |

`parseContentStreamRuns` is deliberately per-stream. **It is not a defect and must not be "fixed".** It
is the seam the repair is built from, and it is the site a re-pointed detector should target.

### 6. The D-000.34 candidate, named in advance

`TestReadEmittedRunsScansEveryTJAdjustment` (`shaped_fixture_test.go:532`) cross-checks the parser's TJ
adjustments against an **independent recount taken over the whole file's bytes** (`strings.Split(string(b),
"BT\n")`). On a single-page subject *"one stream"* and *"the whole file"* coincide, which is why it
passes.

That coincidence is what the repair removes. The developer must decide, and **record the reason either
way**, whether this guard's discriminating power depends on the parser side being **one stream**:

- if it does, **re-point it at `parseContentStreamRuns` over one explicitly-selected stream** — D-000.34's
  *"re-point rather than re-run"*; or
- if it does not, state the argument for why the whole-document recount and a whole-document parser side
  remain **independent** re-derivations rather than the same scan twice.

**A test that silently stops discriminating is worse than one that fails.** This is the one place in the
story where that is a live risk.

---

## Scope fence — what this story is NOT

- **It emits no bytes.** No file under `fixtures/` changes. No `.go` file outside `*_test.go` changes.
  `render.go`, `render_entry.go`, `internal/pdf`, `internal/layout` and the templates are **out of
  scope**. AC1 makes this machine-checkable rather than a promise.
- **It records no new golden and registers no new `matrixDocuments` entry.** Consequently D-000.54's
  registration leg is not owed. If the developer finds a reason to register one, that is an escalation,
  not a decision to take in-story.
- **It adds no sixth Epic 2 gate obligation** (D-2.6.6). `declaredEpic2GateObligations`
  (`byte_neutrality_test.go:466`) stays **byte-unchanged** and
  `TestEpic2GateObligationsMatchTheDeclaredSet` stays green untouched.
- **`epic-2: backlog` stays** in `sprint-status.yaml`. Flipping it is the gate's to do, not this story's.
- **It does not implement `Page X of Y`.** That is Story 2.7. This story exists so 2.7's new call sites
  are written against a known map.
- **It does not build a new fixture to make a forward guard provable.** That would mean new emitted
  bytes. Forward guards stay forward and are labelled (D-000.24).
- **It does not re-open the pagination model** (D-2.6.1 as amended), the two-oracle split (D-2.6.7) or the
  three-layer structure (D-2.6.8).

---

## Acceptance Criteria

### AC1 — No emitted byte moves, and it is proven, not asserted

Every committed golden's digest is unchanged after this story. The three sign-off-relevant artifacts are
pinned explicitly in the Delivery Log with digests **re-computed at the finishing commit** via
`rtk proxy shasum -a 256 … > file` (D-000.12 corrected), and compared against the table at the top of
this file. Additionally: the diff for this story contains **no** non-test Go file and **no** file under
`fixtures/`.

*Carried by*: the existing digest tests plus an explicit re-measurement. **Proven** — the subject is the
committed corpus, which exists.

### AC2 — The inventory is established by a count that does not measure its filter (D-000.14) — FIRST TASK

Before any repair, the story produces the real inventory: for each shared PDF-reading helper, its
definition site and its **invocation** count, counted by a means that distinguishes `readEmittedRuns(`
from `readEmittedRunsAllPages(`. The record states the corrected number, states that the inherited
figure of **12 / matrix-4** was a substring artifact, and states the corrected **11 / matrix-3**.

*Red-proof*: the naive substring count returns **12**; the corrected count returns **11**. Both are run
and both figures recorded. **Proven.**

### AC3 — The enumeration is behavioural and is recorded as a table

For each quantity the engine can now emit N of — **pages, text content streams, page-header runs,
page-footer runs, page `/Font` resource dictionaries, embedded faces, image XObjects, xref subsections**
— the story records: the subject with N > 1 (or **"none exists"**), the helpers exercised against it, and
the **observed** outcome for each, drawn from a closed vocabulary: `correct` / `fatals` / `mis-counts` /
**`silently truncates`** / `no subject`.

Every row is produced by **running** the helper. No row may be justified by reading the helper's source.

*Red-proof*: finding 1's histogram is a row of this table produced before any fix. **Proven.**

### AC4 — `readEmittedRuns`' silent, nondeterministic truncation is repaired and verified mechanically (D-000.48)

After the repair, reading `fixtures/multi-page/expected.pdf` returns:

- the **same** run count on every call within a process (the baseline defect is nondeterminism, so
  determinism is part of the assertion, not a bonus), and
- a count equal to the sum of the per-stream parses taken in **ascending object-number order** — **33** at
  the time of writing, against the baseline's 9-or-24.

Verified by a test that **runs** the repaired helper against the multi-page subject, not by reading it.
D-000.48 is explicit that a sweep correcting silent truncation is the sweep most likely to introduce one.

*Red-proof*: the baseline histogram `map[9:4 24:16]` over 20 calls, captured at `d1af588`. **Proven.**

### AC5 — A multi-page-aware reader is reachable from the ordinary, untagged suite

At baseline the module's only multi-page-aware reader, `readEmittedRunsAllPages`, sits in
`matrix_test.go` behind `//go:build matrix`, so **no untagged test can call it**. After this story an
untagged test can. The matrix leg's existing use continues to work and `go vet -tags matrix` compiles
(D-000.55: that is a **compile gate**, and the Delivery Log says so in those terms).

*Red-proof*: at baseline, an untagged file calling it fails to build with `undefined:
readEmittedRunsAllPages`. Captured. **Proven.**

### AC6 — D-000.34: no detector is silently disarmed, and `parseContentStreamRuns` stays per-stream

Each of the **11** call sites is classified in the record: *does its discriminating power depend on
reading exactly one content stream?* Any site for which the answer is yes is **re-pointed** at the
per-stream parser rather than re-run against the whole document.

`TestReadEmittedRunsScansEveryTJAdjustment` (`shaped_fixture_test.go:532`) is named as the candidate and
its disposition is recorded **with the argument**, whichever way it goes.
`parseContentStreamRuns` retains its per-stream contract; its comment says so explicitly so a later
reader does not "fix" it.

*Red-proof*: for any re-pointed site, a mutation that reverts the parser's per-stream behaviour must make
that site fail. **Proven** where a site is re-pointed; if none is, the record says *no site depended on
it* and shows the classification table that establishes it.

### AC7 — `resourceType0Objects` is repaired; its guard is labelled a forward guard (D-000.50, D-000.24)

The helper resolves font resources across **every** page object rather than an arbitrary one. Verified
byte-for-byte equivalent on all eight committed goldens — the repair changes **no** test outcome today,
which is the honest statement of its value.

The record states, in D-000.50's terms, that **no committed subject can express the defect** (both
`multi-page` page objects carry a byte-identical `/Font` dictionary — measurement in finding 2), and
therefore labels the accompanying guard a **forward guard**. It is **not** counted as coverage.

*Red-proof*: **none available.** Declared as such under D-000.24. **Forward.**

### AC8 — Quantities with no expressible subject are named as forward or redundant, never as coverage

`image XObjects` (max 1 across the corpus) and `xref subsections` (exactly 1 by construction, and
`assertXrefEntriesPointAtTheirObjects` reads the first and stops **without fatalling**) are each recorded
as **forward** (D-000.24) or **redundant** (D-000.42) with the reason. Neither is counted toward the
sweep's fix total.

*Red-proof*: none available for either. Declared. **Forward / redundant.**

### AC9 — D-000.17: the enumeration's remainder is reported unmet, instance by instance

If AC2/AC3's enumeration exceeds what this story carries, **every** unrepaired instance is listed by
helper, site, quantity and observed behaviour, in the Delivery Log **and** appended to
`deferred-work.md`. The remainder is **never** quietly dropped and **never** triaged by preference or by
apparent importance. If the enumeration is fully discharged, the record says so explicitly and shows the
count that establishes it.

*Carried by*: the record itself, cross-checked against AC3's table — every row not marked repaired
appears in the deferred list. **Proven.**

### AC10 — D-000.42: the record counts what the sweep FIXES

The Delivery Log reports **four separate numbers**: helpers **repaired**, helpers **already correct**,
guards **forward**, instances **deferred**. *"Helpers inspected"* and *"call sites reviewed"* are **not**
among them and must not appear as a headline figure.

### AC11 — The baseline is preserved, and any delta is itemised

After this story, with the same scope (whole module, no build tags, `-count=1`): **557 PASS / 1 FAIL**
all-occurrences and **338 / 1** top-level, sole failure `TestCorpusMeetsP6ExerciseFloors` with stats
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` unchanged — **or**, if the sweep adds or re-points
tests, the new totals are stated with the delta itemised **test by test**, not as a net number.

### AC12 — D-000.55: the Delivery Log states what was executed, on which target by name

The Delivery Log separately states (a) what was **executed** and on **which target, named**, and (b) each
**unrun leg, named individually**. The banned compound phrase *"written, compiled and vetted, deliberately
not run"* does **not** appear. `go vet -tags matrix` is described as a **compile gate**, never as evidence
that anything ran.

---

## Heavy-test cadence — proposed DECLINED, stated so it can be refused

**D-000.4 is per-epic. This story proposes no override.** The four matrix legs stay deferred to the Epic 2
boundary gate.

**D-000.4's override criterion (D-000.4, override criterion) is not met**: an override needs *a new source
of cross-target divergence*, and this story introduces none — it emits no bytes, registers no document
and touches no serializer. A test-helper change cannot diverge by target.

**D-000.54 is not triggered.** It applies at the registration of a **new `matrixDocuments` entry**, and
this story registers none. If the developer concludes one is needed, that is an **escalation**, not a
decision to take in-story; if it is nonetheless ruled in, the obligation is exactly:

```
FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .
```

**`go vet -tags matrix` is owed** as a **compile gate** — `matrix_test.go` is edited by AC5 and must
still build. Per D-000.55 it is recorded as a compile gate and nothing more.

---

## Task breakdown

- [x] 1. **Establish the real inventory (AC2). This is the first task and the scope decision depends on its
   number.** Count invocations of every shared PDF-reading helper by a means that distinguishes
   `readEmittedRuns(` from `readEmittedRunsAllPages(`. Record both the naive figure and the corrected
   one. Do not begin any repair before this table exists.
- [x] 2. **Run the enumeration (AC3).** For each quantity in AC3's list, point the existing helpers at
   `fixtures/multi-page/expected.pdf` (or record *"no subject"*) and write down what happened. Use a
   throwaway probe; delete it. Findings 1–5 of this story are the seed, not the answer — re-measure them.
- [x] 3. **Decide the scope on the number (D-000.17).** With the table in hand, state what this story carries
   and what it defers. Defer by **enumeration**, never by preference.
- [x] 4. **Repair `readEmittedRuns`' truncation and make a multi-page-aware reader reachable untagged
   (AC4, AC5).** Adopt `readEmittedRunsAllPages`' deterministic ascending-object-number ordering. Verify
   against the multi-page subject **mechanically** (D-000.48).
- [x] 5. **Classify all 11 call sites for D-000.34 (AC6).** Re-point any site whose discriminating power came
   from reading one stream. Record the disposition of
   `TestReadEmittedRunsScansEveryTJAdjustment` with its argument.
- [x] 6. **Repair `resourceType0Objects` and label its guard forward (AC7).** Prove byte-for-byte equivalence
   on all eight goldens.
- [x] 7. **Record the no-subject quantities as forward or redundant (AC8).**
- [x] 8. **Update the stale in-code comments.** `shaped_fixture_test.go:379–395` states *"12 call sites"* and
   *"eleven callers"* in one block, and describes a contract this story changes. Leave no comment citing
   a figure this story has corrected.
- [x] 9. **Write the record (AC9, AC10, AC11, AC12).** Four fix-counts, the deferred list by name, the
   re-measured suite figures with scope, the digest re-measurement, and D-000.55's execution statement.
- [x] 10. **Set the story's sprint status to `review`.**

---

## Flagged, not fixed

- **`resourceType0Objects` matches `"/Type /Page "` with a trailing space** — the exact substring hazard
  `countPageObjects` was repaired for under Nit 26. Note it in the record; repairing it is optional and
  must not be bundled silently.
- **`nonStreamRegions` is a second, independent copy of `assertStreamLengthsAreExact`'s walk** with an
  undeclared lockstep requirement (Story 2.6 finisher, Finding 8). It is not a single-page assumption and
  is **out of scope**; do not absorb it.

---

## DECISIONS NEEDED — escalate before development starts

**None blocking.** Every fork this story presents is settled by a ruling already in force:

- *Repair `readEmittedRuns` in place, or add a second helper?* — settled by **D-000.34**: keep
  `parseContentStreamRuns` per-stream as the seam, and re-point rather than re-run any site that
  depended on single-stream reading. Either shape satisfies AC4/AC6 provided that holds.
- *Build a subject that can express `resourceType0Objects`' defect?* — settled by the **scope fence**
  (no new emitted bytes) plus **D-000.50/D-000.24**: the guard stays **forward** and is labelled.
- *Does the sweep earn a gate obligation?* — settled by **D-2.6.6**: no. The gate owes exactly five.

If the developer concludes a new `matrixDocuments` entry is required, **stop and escalate** — that would
engage D-000.54 and change this story's cadence answer.

---

## Dev Agent Record

### Delivery Log

**AC2 — Real inventory, re-confirmed at this commit.** Counted with
`grep -rn '[^A-Za-z]readEmittedRuns(' . | grep -v 'func readEmittedRuns('`, which distinguishes
`readEmittedRuns(` from `readEmittedRunsAllPages(` (the latter's open paren is preceded by `AllPages`,
not by a non-identifier character, so it never matches this pattern):

| file | invocations |
|---|---|
| `shaped_fixture_test.go` | 5 (lines 534, 776, 854, 905, 1073) |
| `matrix_test.go` | 3 (lines 782, 829, 874) |
| `wrapped_text_fixture_test.go` | 1 (line 79) |
| `segment_origin_test.go` | 1 (line 95) |
| `first_baseline_acceptance_test.go` | 1 (line 291) |
| **total** | **11** |

Confirms the correction already recorded in commit `804eea1` (11, not the inherited 12) at this story's
own start-of-work commit. No repair began before this table existed (Task 1).

**Correction (finisher, review Finding 5 — Major).** The table above is right about `804eea1`; it is
stale for the tree it claims to describe — it names no commit and every line number in it moved. Re-taken
at **`99f8f4a` (this story's own finishing commit — the single local commit this finisher produced;
working tree confirmed clean of all other changes at the time of this count)**, with the same
discriminating pattern:

| file | invocations |
|---|---|
| `shaped_fixture_test.go` | 7 (lines 562, 610, 732, 974, 1052, 1103, 1271) |
| `matrix_test.go` | 4 (lines 701, 778, 825, 870) |
| `wrapped_text_fixture_test.go` | 1 (line 79) |
| `segment_origin_test.go` | 1 (line 95) |
| `first_baseline_acceptance_test.go` | 1 (line 291) |
| **total** | **14** |

The delta from 11 (at `804eea1`) to 14 (at `99f8f4a`) is itemised, not netted:

- **+1**: `matrix_test.go:701`, `requireMultiPageIsGenuinelyMultiPage`. This story's own AC5 work
  re-pointed it from the retired `readEmittedRunsAllPages` onto `readEmittedRuns` directly — the
  "twelfth" site the review's Finding 5 first identified.
- **+2**: `shaped_fixture_test.go:562` and `:610`, `TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument`
  and `TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument`, committed by this
  finisher closing review Blocker 1 (Finding 1): the story's central repair had shipped with no test in
  the ordinary suite that could see it reverted.

Of the 14, **3 render the multi-page (two-text-stream) fixture** — the sites above — and **11 render a
single-text-stream fixture**, the same 11 named in the `804eea1` table. AC6's classification below is
extended to cover all 14.

**AC3 — Behavioural enumeration, run against `fixtures/multi-page/expected.pdf`** (the only committed
subject with N > 1 on any axis except embedded faces), via a throwaway probe deleted immediately after
each measurement:

| quantity | subject (N>1) | helper(s) | baseline observed | repaired? | post-repair observed |
|---|---|---|---|---|---|
| pages | multi-page (2) | `countPageObjects`, `assertWellFormedPDF` | correct (2) | already repaired (Story 2.6) | correct |
| text content streams | multi-page (2 of 4 streams) | `readEmittedRuns` | **silently truncates** — 20-call histogram `map[9:4 24:16]` against 33 actual, nondeterministic (Finding 1, the red-proof) | **yes, this story** | correct — 200/200 calls return 33, deterministic |
| page-header runs | multi-page (2 occurrences, Y=798269) | `readEmittedRuns` | **silently truncates** — sees at most one page's header run, and which one is nondeterministic (subsumed by the streams defect above) | **yes, this story** | correct — 2 header runs found |
| page-footer runs | multi-page (2 occurrences, Y=51448) | `readEmittedRuns` | **silently truncates** (same mechanism) | **yes, this story** | correct — 2 footer runs found |
| page-header/footer runs (correction, finisher, review Finding 3) | multi-page (2 pages) | `mpExtractRuns` + `splitPageContentStreams` (`multi_page_fixture_test.go`, untagged, pre-existing at baseline `804eea1`) | **already correct** — asserts per-page, not merely per-document, via `TestMultiPageHeaderAndFooterAppearOnEveryPage`; this row was OMITTED from the original enumeration entirely (AC5's false "only multi-page-aware reader" premise), not merely mis-classified | not needed — already correct, pre-existing | correct |
| page `/Font` resource dictionaries | multi-page (2 page objects) | `resourceType0Objects` | **silently truncates** — reads only the first `/Type /Page` object's `/Font` dict and `break`s; observably indistinguishable from correct on this subject only because both pages' dicts are byte-identical (Finding 2) | **yes, this story** | correct — verified with a throwaway synthetic two-page subject carrying two DIFFERENT font resource names; both now present |
| embedded faces | `multi-script-fallback`, `shaped-text`, `wrapped-text` (3 each) | `extractAllFontFile2Programs` | correct — already exercised at N=3 today | not needed | correct |
| image XObjects | **no subject** (max 1 across the corpus) | `containsImageXObject` (presence/OR scan, not a take-first walk) | no subject | not applicable | no subject |
| xref subsections | **no subject** (exactly 1 by construction) | `assertXrefEntriesPointAtTheirObjects` | no subject on any committed golden, but the mechanism read one subsection and returned without ever checking whether a second, unexamined subsection followed (Finding 3) | **yes, this story** (mechanism repaired even without a provable subject, per D-000.50) | no subject (repair verified with a throwaway synthetic two-subsection buffer: a bad second-subsection entry now fatals, a good one now passes) |

Every row above was produced by **running** the helper against the subject, never by reading source.

**AC4 — `readEmittedRuns` repaired, verified mechanically.** `readEmittedRuns`
(`shaped_fixture_test.go`) now walks every `pdfObjects` object number in ascending order, parses every
Tf-bearing stream via the unchanged per-stream `parseContentStreamRuns`, and concatenates the results.
Verified against `fixtures/multi-page/expected.pdf` with 200 calls in one process: **histogram
`map[33:200]`** — every call returns exactly 33 runs, against baseline's nondeterministic 9-or-24. The
probe was a throwaway test file, deleted after the run.

**Correction (finisher, review Blocker 1 — the story's central repair shipped with zero coverage in the
ordinary suite).** The measurement above is real, but it was never committed as an assertion — only the
throwaway probe that produced it, then deleted, per AC11's own original statement that no test was added.
The reviewer red-proofed this directly: reverting `readEmittedRuns` to `break` after the first Tf-bearing
stream (the pre-repair defect, restored exactly) left `go test -count=1 ./...` green, because the only
call site that renders a multi-stream document was `matrix_test.go`'s `requireMultiPageIsGenuinelyMultiPage`,
reachable only under `-tags=matrix`. The correction to "reads one stream and stops" shipped alongside no
guard that could see a regression back to it in the ordinary suite — the story's own defect class, one
level out (D-000.48).

**Fixed.** Two tests are now committed in `shaped_fixture_test.go`, both untagged, both rendering the
multi-page fixture directly via `renderMultiPage` (`multi_page_fixture_test.go`, no new fixture, no new
emitted bytes):

- `TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument` — cross-checks
  `readEmittedRuns`'s total run count against an INDEPENDENTLY derived total (`splitPageContentStreams`,
  which follows the page tree's `/Kids` and each page's own `/Contents` reference — a different traversal
  than `readEmittedRuns`'s `pdfObjects` + `" Tf\n"` scan — re-parsed per page by the same
  `parseContentStreamRuns`). Two independently reached traversals agreeing is stronger than one hard-coded
  expected count (D-2.6.8). Red-proofed against both the delivered tree's exact defect (`break` after the
  first stream: fails, 24 vs. 33) and the true baseline defect (unsorted map iteration + `break`: fails,
  9-or-24 vs. 33).
- `TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument` — the determinism half
  D-2.6.8 requires separately from construction: 50 calls in one process must all agree. 50 is chosen to
  comfortably exceed the 20 calls that empirically produced both branches under the true pre-repair code
  (Measured findings 1); the chance of 50 independent calls coincidentally agreeing were this run against
  the unrepaired helper is on the order of 2^-49. Red-proofed against the true baseline defect (histogram
  `map[9:13 24:37]` over 50 calls when reverted); the `break`-only mutation alone does not redden this
  test, because it preserves the deterministic sorted iteration order this story's `sort.Ints` call
  already provides — the finding is specifically about UNSORTED map iteration, which is what the true
  baseline defect requires to reproduce.

**AC5 — Multi-page-aware reader reachable from the untagged suite.** `readEmittedRuns` itself is now a
multi-page-aware reader, and it always lived in `shaped_fixture_test.go` (package `folio`, no build tag).
`readEmittedRunsAllPages` (`matrix_test.go`, behind `//go:build matrix`) is now redundant with it: its one
call site (`requireMultiPageIsGenuinelyMultiPage`) is re-pointed at `readEmittedRuns` directly, and the
sibling function is deleted rather than kept as an unused alias. `go build ./...` and `go vet ./...`
(untagged) and `go build -tags matrix ./...` and `go vet -tags matrix ./...` all pass — the tagged file
still compiles with the sibling gone.

**Correction (finisher, review Finding 3 — Major): AC5's premise as originally written was false.** It
claimed `readEmittedRunsAllPages` was *"the module's only multi-page-aware reader"* at baseline. It was
not: `multi_page_fixture_test.go` (untagged, `package folio`, present at baseline `804eea1`) already held
`mpParseToUnicode`, `mpExtractRuns` and `splitPageContentStreams`, driving
`TestMultiPageHeaderAndFooterAppearOnEveryPage` — which resolves the page tree's `/Kids` array, follows
each page's own `/Contents` reference, and asserts *per page* that exactly one header and one footer run
appear at the declared `Tm`. That is strictly more multi-page-aware than `readEmittedRuns` even after this
story's repair: `readEmittedRuns` returns a flat `[]emittedRun` with no page attribution, while the `mp*`
family already carried real per-page assertion before this story began.

AC5's own red-proof (`undefined: readEmittedRunsAllPages` on an untagged probe) only established that
*that one identifier* was tagged; it never established the "only" claim, which was the load-bearing half
and was never measured. This is the same shape of unmeasured "X is the only Y" claim Story 2.6's finisher
introduced with `/Kids` and D-2.6.9 exists to generalise away — reproduced here inside the very story
generalising it.

**Consequence for AC3 and AC9, corrected below**: the `page-header runs` / `page-footer runs` axis rows
named only `readEmittedRuns` as the reaching helper and marked the axis *"covered by finding 1's
repair"*. The `mp*` family — the helper that actually carries a genuine per-page assertion for that
property — was never named by the enumeration at all, so AC9's "enumeration fully discharged" was a claim
about an axis list that had already missed an entire pre-existing, already-correct family. Nothing needs
repairing here (the `mp*` family was correct before this story and remains correct), but the enumeration's
completeness claim needed the correction recorded rather than silently left as delivered.

**AC6 — D-000.34 classification of the 11 single-stream call sites at the story's own baseline.** Every
one of the 11 sites was traced to the fixture it renders:

| site | fixture rendered | text streams in that fixture | depends on single-stream reading? |
|---|---|---|---|
| `shaped_fixture_test.go:534,776,854,905,1073` (5 sites) | `renderShapedTextFixture` (shaped-text) | 1 | no |
| `matrix_test.go:782` (`requireThreeBandPageUsesAllThreeBands`) | three-band-page | 1 | no |
| `matrix_test.go:829` (`requireWrappedTextIsWrapped`) | wrapped-text | 1 | no |
| `matrix_test.go:874` (`requireShapedTextIsShaped`) | shaped-text | 1 | no |
| `wrapped_text_fixture_test.go:79` | `renderWrappedText` | 1 | no |
| `segment_origin_test.go:95` | synthetic one-element template, rendered directly | 1 | no |
| `first_baseline_acceptance_test.go:291` | `baselineAcceptanceFixtures` (font-text, etc.) | 1 | no |

**No site's discriminating power depends on single-stream reading** — every fixture any of these 11 sites
renders carries exactly one text-bearing content stream, so "read the one stream" and "read every stream
and concatenate" are the same value for all of them. **Zero of these 11 sites are re-pointed.**

**Correction (finisher, review Finding 5): the classification above was never extended to the twelfth
site this story's own AC5 work created, and this finisher adds a further two.** All 3 are the sites named
in AC2's corrected inventory above that DO render the multi-page (two-text-stream) fixture, and — unlike
the 11 above — DO depend on reading more than one stream; that dependency is exactly why they exist:

| site | fixture rendered | text streams | depends on multi-stream reading? |
|---|---|---|---|
| `matrix_test.go:701` (`requireMultiPageIsGenuinelyMultiPage`) | multi-page | 2 | **yes** — this is the site `readEmittedRunsAllPages` existed for; AC5 re-pointed it onto the repaired `readEmittedRuns` |
| `shaped_fixture_test.go:562` (`TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument`) | multi-page | 2 | **yes** — committed by this finisher closing Blocker 1 |
| `shaped_fixture_test.go:610` (`TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument`) | multi-page | 2 | **yes** — committed by this finisher closing Blocker 1 |

None of these 3 is "re-pointed" in D-000.34's sense (that language applies to a site whose *guard* would
be disarmed by reading more than one stream) — they are simply the 3 sites for which every-stream reading
is the entire point, and are unaffected by that concern.

`TestReadEmittedRunsScansEveryTJAdjustment` (the named D-000.34 candidate,
`shaped_fixture_test.go:730`) is disposed of the same way, with its own argument: its independent
recount already scans `strings.Split(string(b), "BT\n")` over the **whole file's raw bytes**, not over
one stream's bytes — it was never a per-stream re-derivation. Because the shaped-text fixture it renders
carries exactly one text-bearing content stream, the whole-file BT scan and "every Tf-bearing stream,
concatenated" enumerate the identical set of `BT`…`ET` blocks both before and after the repair. Its
discriminating power comes from re-deriving TJ adjustments straight off the bytes independently of the
parser, not from any assumption about stream count, so it is **not re-pointed** and remains a genuine
independent cross-check.

**AC7 — `resourceType0Objects` repaired; guard PROVEN (corrected below).** It now collects the `/Font`
resource dictionary from **every** `/Type /Page` object (in deterministic ascending object-number order)
and merges their entries, instead of stopping at the first.

**Correction (finisher, review Finding 4 — Major): the original "byte-for-byte-equivalent on all eight
committed goldens" claim was false as evidence.** Tracing every path into `resourceType0Objects`: its only
callers are `cidToGlyphForResources` and `toUnicodeForResources`, and every call site of those two renders
the **shaped-text** fixture only. `resourceType0Objects` is never reached by any committed test, tagged or
untagged, against the multi-page golden — the one subject with more than one page object. An unchanged
full-suite total is evidence about one single-page fixture, not eight goldens, and specifically says
nothing about the only subject with N > 1 page objects. The conclusion (equivalent behaviour today) is
correct; the stated proof for it was a count that measured its filter, in the story whose AC2 exists to
name that exact class of error.

**Correction (finisher, review Finding 7 — Minor: the stronger, correct D-000.50 justification).** The
right reason no document can express the defect is not merely that `fixtures/multi-page/`'s two page
objects happen to carry a byte-identical `/Font` dict — it is that no document **this engine can emit**
ever could: `internal/pdf/textdoc.go` collects the face set document-globally
(`faceNames := slices.Sorted(maps.Keys(faces))`, `textdoc.go:130`), rejects an escaped-name collision
across the **whole document** up front (`textdoc.go:140-143`), and then writes that identical
`/Font << … >>` dictionary, in the same sorted order, into **every** page object (`textdoc.go:217`). Two
pages with genuinely different fonts is not a fixture this repo happens to lack; it is a document shape
the serializer cannot produce until it writes per-page font dictionaries. This sharpens AC7's honest
statement of value from *"changes no test outcome today"* to *"cannot change any test outcome until the
emitter writes per-page font dictionaries"* — the fact a future page-aware repair needs. This reasoning
now lives in `resourceType0Objects`'s own docblock (`shaped_fixture_test.go`), not only here.

**Correction (finisher, review Finding 6 — Major: the discarded red-proof, rebuilt and committed).** The
Delivery Log as originally written stated discriminating power was *"proven with a throwaway synthetic
two-page subject carrying two different font resource names (`/FaceA`, `/FaceB`) ... The probe was
deleted after the run"* — declaring the guard **forward** (D-000.24) while simultaneously describing an
available red-proof that had been built and thrown away. D-000.50 asks whether **any** subject can express
the defect, not whether a committed golden can; an in-memory literal is neither an emitted byte nor a
fixture, so there was never a reason to discard it. `TestResourceType0ObjectsResolvesFontsFromEveryPageObject`
(`shaped_fixture_test.go`) is that exact synthetic subject, committed: it builds a two-page buffer with
`/FaceA` and `/FaceB` naming two different Type0 objects and asserts both resolve. Red-proofed by
reverting the repair to `break` after the first page: the test fails, resolving only one face. **This
guard is PROVEN (D-000.50), not forward — the AC10 count below is corrected accordingly.**

**Shape note (Finding 7), not a live defect**: the repaired helper is page-**complete** (it visits every
page) but not page-**aware** — it merges every page's dict into one `map[string][]byte`, so two pages
mapping the same resource *name* to two *different* objects would still be last-writer-wins, silently.
Cannot fire today for the same serializer reason above; recorded for whoever changes that assumption next.

**AC8 — No-subject quantities named as forward or redundant.**

- **Image XObjects** — no committed subject has more than one, and the module's one image-related helper,
  `containsImageXObject`, is a presence/OR scan over all objects, not a take-first walk — it already
  reports correctly regardless of how many image XObjects exist. There is no truncating mechanism for this
  axis to guard against, so this is recorded as **redundant** (D-000.42): a dedicated forward guard would
  duplicate a property the existing implementation already has by construction. Not counted as coverage.
- **Xref subsections** — no committed subject has more than one (the serializer emits exactly one by
  construction), but `assertXrefEntriesPointAtTheirObjects` (`render_test.go`) was repaired anyway
  (D-000.50 permits repairing a real defect even without a provable subject): it now loops over
  subsections until it reaches the `trailer` keyword, validating each one's entries and requiring each
  subsequent subsection's `start` be no less than the previous one's next object number, and fatals if
  what follows the entries is neither `trailer` nor a well-formed further subsection header. Previously it
  examined exactly one subsection and returned unconditionally.

  **Correction (finisher, review Blocker 2): the repair as originally delivered strictly WEAKENED this
  check.** The `trailer` test was evaluated **before** any subsection header was ever consumed, so a
  buffer declaring **zero** subsections (`"xref\ntrailer\n..."`) returned cleanly, with no assertion made
  at all — the pre-repair function correctly rejected that input (`"xref subsection header \"trailer\"
  does not have two fields"`); the repair accepted it. This is the exact defect class this repair exists
  to remove — silent truncation to nothing — reintroduced by the repair itself, inside the story
  generalising that class (D-000.48). **Fixed**: the loop now fatals if `trailer` is reached with zero
  subsections examined. Red-proofed by reverting to the original shape: the new committed test
  (`TestAssertXrefEntriesPointAtTheirObjectsRejectsZeroSubsections`, `render_test.go`) fails against the
  reverted code and passes against the fix.

  **Correction (finisher, review Finding 8 — Minor): the original repair's continuity check was too
  strict.** It required each subsequent subsection's `start` to equal the running next-object-number
  exactly, which rejects the legal **non-contiguous** form ISO 32000-1 §7.5.4 explicitly permits (e.g. a
  table describing objects `0-1` then `5-7`, skipping the range between). **Fixed**: the check now only
  rejects a subsection that *overlaps* an object number a previous subsection already claimed
  (`start < nextObjNum`), and accepts a gap. Proven both directions by two new committed tests
  (`TestAssertXrefEntriesPointAtTheirObjectsAcceptsNonContiguousSubsections`,
  `TestAssertXrefEntriesPointAtTheirObjectsRejectsOverlappingSubsections`).

  **Correction (finisher, review Finding 6 — Major: the discarded red-proofs, rebuilt and committed).**
  The Delivery Log as originally written declared this a **FORWARD guard (D-000.24)** while its own text
  described discriminating power *"verified with a throwaway synthetic two-subsection buffer ... a bad
  second-subsection entry offset now fatals ... and a well-formed second subsection now passes"* — an
  available red-proof, built and then discarded. Per D-000.50, an in-memory literal that expresses the
  defect is an available red-proof whether or not any committed golden also expresses it. Four tests are
  now committed in `render_test.go`, all re-executing the negative cases in a re-exec'd subprocess of the
  test binary (the same technique this file's `subprocessEnvVar` family already uses, needed here because
  a failing `t.Run` subtest unconditionally fails its parent test, with no supported in-process way to
  observe and swallow that failure): `...RejectsZeroSubsections` (Blocker 2's red-proof),
  `...AcceptsNonContiguousSubsections` and `...RejectsOverlappingSubsections` (Finding 8's red-proof, both
  directions), and `...RejectsBadOffsetInSecondSubsection` (the review's own probe, committed, proving the
  pre-existing per-entry offset check applies uniformly to every subsection, not only whichever one the
  loop happens to reach first). **This guard is PROVEN (D-000.50), not forward — the AC10 count below is
  corrected accordingly.**

- **Flagged-not-fixed hazard (finisher, review Finding 9 — Minor): recorded here, not fixed.** The story's
  "Flagged, not fixed" section requires noting that `resourceType0Objects` matches `"/Type /Page "` with a
  trailing space — the exact substring hazard `countPageObjects` was repaired for under Nit 26 — and that
  repairing it here is optional. The originally delivered record never discharged that requirement; this
  is the discharge. It was seen, deliberately not repaired (repairing it was always optional and is
  outside this story's charter — it is a well-formedness hazard in a test helper, not a single-page
  assumption), and is left exactly as flagged for whichever future story next touches
  `resourceType0Objects`.

**AC9 — Enumeration remainder.** All 8 axes in AC3's table are accounted for: 2 repaired this story
(text/header/footer runs via the single `readEmittedRuns` repair, and page `/Font` resource dictionaries
via `resourceType0Objects`), 1 repaired with a now-PROVEN guard (xref subsections — corrected under AC8
above), 1 already correct (embedded faces), 1 already repaired by a prior story (pages), and 1 with no
subject and no defect-shaped helper to guard (image XObjects, redundant). **The enumeration is fully
discharged — zero instances are deferred.** `deferred-work.md` receives no new entries from this story.

**Correction (finisher, review Finding 3).** The axis list itself was incomplete, not merely its
classifications: the `mp*` reader family (`mpExtractRuns`, `splitPageContentStreams`,
`mpParseToUnicode`) already carried a genuine per-page guard for the header/footer-runs axis at baseline,
and the original enumeration never named it (see the AC3 table row added above and the AC5 correction).
Nothing on that axis needed repair — the family was already correct — but "enumeration fully discharged"
is a claim about the axis list's completeness, and the axis list needed the addition recorded here to
support that claim honestly. With the addition made, the enumeration is discharged for real: the `mp*`
family needs no fix, and zero instances remain deferred.

**AC10 — Fix-counts (D-000.42: counts what the sweep FIXES, not what it inspected).**

| category | count | items |
|---|---|---|
| helpers repaired | **3** | `readEmittedRuns` (truncation + nondeterminism), `resourceType0Objects` (single-page assumption), `assertXrefEntriesPointAtTheirObjects` (single-subsection assumption, plus the finisher's Blocker 2 / Finding 8 corrections to that same repair) |
| helpers already correct | **7** | `pdfObjects`, `streamBody`, `countPageObjects`, `assertWellFormedPDF` (repaired by Story 2.6), `parseContentStreamRuns` (correct by design — stays per-stream, AC6), `extractAllFontFile2Programs`, and (finisher, review Finding 3) the `mp*` per-page reader family (`mpExtractRuns`, `splitPageContentStreams`, `mpParseToUnicode`) — pre-existing, omitted from the original enumeration |
| guards PROVEN | **3** | `readEmittedRuns` (2 committed tests, finisher), `resourceType0Objects` (1 committed test, finisher), `assertXrefEntriesPointAtTheirObjects` (4 committed tests, finisher) — see below |
| guards forward | **0** | — corrected from the original **2** (finisher, review Finding 6): both were declared forward while describing a discarded, working synthetic red-proof; both are rebuilt and committed |
| guards redundant | **1** | image XObjects axis (AC8) |
| repairs shipped unguarded | **0** | — corrected from the original delivered state, in which all 3 repairs shipped with no committed assertion at all (review Blocker 1, Finding 6) |
| instances deferred | **0** | — enumeration fully discharged (AC9, corrected) |

**Correction (finisher, review Findings 1 and 6): the story as originally delivered added no test and no
assertion at all — only helper bodies, imports and comments.** That is why the whole diff could be
reverted with the default suite staying green (review Blocker 1's exact demonstration), and why AC10's
original "forward: 2" line was counting two guards that did not exist as committed code. The finisher adds
**7 committed tests** across `render_test.go` and `shaped_fixture_test.go`, red-proofed individually by
reverting each corresponding repair and confirming the specific new test reddens, then restoring the fix
and confirming green (D-000.30, D-000.40 — verified as a non-empty diff, never an exit code, for every
mutation): `TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument`,
`TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument`,
`TestResourceType0ObjectsResolvesFontsFromEveryPageObject`,
`TestAssertXrefEntriesPointAtTheirObjectsRejectsZeroSubsections`,
`TestAssertXrefEntriesPointAtTheirObjectsAcceptsNonContiguousSubsections`,
`TestAssertXrefEntriesPointAtTheirObjectsRejectsOverlappingSubsections`,
`TestAssertXrefEntriesPointAtTheirObjectsRejectsBadOffsetInSecondSubsection`.

**AC11 — Baseline preserved, delta itemised (finisher added tests, so the fallback clause applies).**
Same scope as the baseline (whole module, no build tags, `-count=1`), measured with
`go test -count=1 -v ./... > file` and counted by pattern over the file: **all-occurrences PASS 564 / FAIL
1** (was 557/1), **top-level PASS 345 / FAIL 1** (was 338/1). The sole failure is unchanged:
`TestCorpusMeetsP6ExerciseFloors`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`, byte-identical
to the pre-story baseline. The +7 top-level delta is itemised, test by test, not netted — it is exactly
the 7 tests AC10 lists above, all added by the finisher closing review Blockers 1 and 2 and Finding 6.
Zero tests were removed or re-pointed.

**AC12 — Execution statement (D-000.55).**

*Executed by the finisher, this session, on darwin/arm64 (native):*
- `go build ./...` — passes.
- `go vet ./...` — passes.
- `go test -count=1 -v ./...` (whole module, untagged) — 564/1 all-occurrences, 345/1 top-level (AC11).
- `go build -tags matrix ./...` — passes (compiles).
- `go vet -tags matrix ./...` — passes. **This is a compile gate confirming `matrix_test.go` still
  compiles. It is not evidence that any matrix leg ran.**
- `FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .` — **ran**,
  passes (`ok`). This is the one matrix leg the story's own "Heavy-test cadence" section owed as a
  non-deferred check; the finisher re-ran it after the review's fixes landed.
- `shasum -a 256 fixtures/shaped-text/expected.pdf fixtures/expected-breaks/expected_breaks.json
  fixtures/multi-page/expected.pdf` — all three digests match the values pinned at the top of this file
  exactly (AC1).

*Unrun legs, named individually:*
- The three remaining matrix native legs (`FOLIO_MATRIX_TARGET={linux/amd64,linux/arm64,js/wasm} go test
  -tags=matrix -count=1 -run TestTargetRenderHash .`) did not run. Deferred to the Epic 2 boundary gate
  per D-000.4 (heavy-test cadence is per-epic); no override criterion is met — this story emits no bytes,
  registers no document, and touches no serializer, so it cannot introduce cross-target divergence (see
  "Heavy-test cadence — proposed DECLINED" above).

**AC1 — No emitted byte moved.** `git status` after this story's changes shows exactly three modified
`*_test.go` files (`matrix_test.go`, `render_test.go`, `shaped_fixture_test.go`) plus the story file and
`sprint-status.yaml`. No file under `fixtures/` and no non-test `.go` file appears in the diff. The three
pinned digests are re-verified unchanged under AC12 above.

### File List

- `folio-go/matrix_test.go` — modified: retired `readEmittedRunsAllPages` (AC5), re-pointed its one call
  site at the repaired `readEmittedRuns`, updated stale comments, removed the now-unused `sort` import.
  **Finisher**: corrected the "other eleven call sites" narrative to name the corrected baseline/finishing
  counts (Finding 5); converted the orphan tombstone comment for the deleted
  `readEmittedRunsAllPages` into an explicit `// NOTE (Story 2.6a, tombstone ...)` so it cannot read as
  documenting a live declaration (Finding 10), and added the Finding 5 correction note to it.
- `folio-go/render_test.go` — modified: repaired `assertXrefEntriesPointAtTheirObjects` to loop over every
  xref subsection instead of silently stopping after the first (AC8). **Finisher**: fixed Blocker 2 (a
  zero-subsection table passed silently), fixed Finding 8 (an overly strict contiguity check rejected the
  legal non-contiguous form), added the `fmt` import, and committed four new tests plus the
  `xrefEntryBytes`/`xrefNegativeCaseEnvVar`/`TestXrefEntriesRejectsMalformedSubprocess`/
  `runXrefNegativeCaseSubprocess` supporting plumbing (Finding 6). Rewrote the function's docblock to
  record both corrections and the PROVEN (not forward) guard status.
- `folio-go/shaped_fixture_test.go` — modified: repaired `readEmittedRuns` to walk every text-bearing
  stream deterministically in ascending object-number order (AC4/AC5/AC6), repaired `resourceType0Objects`
  to read every page object's font resources (AC7), added the `sort` import, rewrote both functions'
  docblocks to remove the stale "12 call sites" / "eleven callers" contradiction (Task 8). **Finisher**:
  fixed Blocker 1 by committing
  `TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument` and
  `TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument`; fixed Finding 6 by
  committing `TestResourceType0ObjectsResolvesFontsFromEveryPageObject`; rewrote `readEmittedRuns`'s and
  `resourceType0Objects`'s docblocks to correct the call-site count (Finding 5), adopt the stronger
  serializer-level D-000.50 justification (Finding 7), and record the PROVEN guard status (Finding 6).
- `_bmad-output/implementation-artifacts/2-6a-sweep-single-page-assumptions-from-the-shared-pdf-readers.md`
  — this story file: added `baseline_commit` frontmatter, checked off all tasks, added this Delivery Log /
  Dev Agent Record. **Finisher**: added the Findings Triage table, corrected AC2/AC3/AC5/AC6/AC7/AC8/AC9/
  AC10/AC11/AC12 in place, rewrote "In plain terms", set Status to `done`.
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — story status updated to `review`, then
  (finisher) to `done`.

### Change Log

- Repaired `readEmittedRuns`'s silent, nondeterministic single-stream truncation; it now reads every
  text-bearing content stream deterministically (AC4).
- Retired the now-redundant `readEmittedRunsAllPages` (`matrix_test.go`, behind `//go:build matrix`),
  making a multi-page-aware reader reachable from the ordinary untagged suite for the first time (AC5).
- Classified the 11 baseline `readEmittedRuns` call sites for D-000.34; none required re-pointing,
  including the named candidate `TestReadEmittedRunsScansEveryTJAdjustment` (AC6).
- Repaired `resourceType0Objects` to resolve font resources across every page object rather than the
  first (AC7).
- Repaired `assertXrefEntriesPointAtTheirObjects` to examine every xref subsection instead of silently
  stopping after the first. Labelled the image-XObjects axis redundant (AC8).
- Updated stale in-code comments that contradicted themselves on the call-site count (Task 8).
- Verified no emitted byte moved: all three sign-off-pinned digests unchanged (AC1).

**Finisher (review response):**

- Fixed Blocker 1: committed two tests (`TestReadEmittedRunsMatchesThePerPageStreamSplitOnAMultiPageDocument`,
  `TestReadEmittedRunsIsDeterministicAcrossRepeatedCallsOnAMultiPageDocument`) closing the story's central
  repair having shipped with no guard reachable from the ordinary suite.
- Fixed Blocker 2: `assertXrefEntriesPointAtTheirObjects`'s own repair had introduced a NEW silent-pass
  defect (a zero-subsection table); fixed and red-proofed with a committed test.
- Fixed Finding 6 (Major): rebuilt and committed the two in-memory synthetic red-proofs the delivered
  story had built, run, and discarded (`resourceType0Objects`'s two-different-faces subject; the xref
  well-formed/malformed two-subsection subjects), converting both repairs' guards from FORWARD to PROVEN;
  recounted AC10 accordingly.
- Fixed Finding 8 (Minor): relaxed the xref contiguity check from exact continuation to
  overlap-rejection-only, matching what ISO 32000-1 §7.5.4 actually permits; proved both directions with
  committed tests.
- Corrected Finding 3 (Major): AC5's "only multi-page-aware reader" premise was false; named the
  pre-existing `mp*` per-page reader family the original enumeration missed, in AC3/AC5/AC9/AC10.
- Corrected Finding 4 (Major): restated AC7's evidence — the full-suite-unchanged claim covers shaped-text
  only, never the multi-page golden; the multi-page equivalence claim now rests on the serializer's
  construction (Finding 7) instead.
- Corrected Finding 7 (Minor): adopted the stronger, serializer-level D-000.50 justification for why no
  document this engine can emit can express `resourceType0Objects`' pre-repair defect; recorded the
  page-complete-not-page-aware shape note.
- Recorded Finding 9 (Minor): the "Flagged, not fixed" trailing-space hazard discharge was missing from
  the Delivery Log; added.
- Fixed Nit 10: converted the orphan tombstone doc comment into an explicit non-doc `NOTE` and corrected
  its stale count.
- Corrected Finding 5 (Major): re-took the call-site inventory at the finishing commit (14, not 12),
  named the commit, extended AC6's classification to all 14 sites, fixed the two in-code comments still
  citing "eleven" as if it were the current total.
- Re-measured AC11/AC12 with the finisher's 7 new tests: 564/1 all-occurrences, 345/1 top-level (was
  557/1, 338/1); re-ran the darwin/arm64 matrix native leg; re-verified all three pinned digests unchanged.

### Completion Notes

All 10 tasks complete. All 12 acceptance criteria verified and recorded above. No new
`matrixDocuments` entry was registered (the one tripwire condition never triggered) and no decision needed
escalation — every fork this story presented was already settled by a standing ruling, as anticipated in
"DECISIONS NEEDED" above. Every change is confined to `*_test.go` files; no fixture and no non-test `.go`
file changed. Epic 2's gate obligations remain at exactly five (`byte_neutrality_test.go` untouched);
`epic-2: backlog` is left as-is in `sprint-status.yaml`.

**Finisher's note.** The code-reviewer's `## QA Results` below found 2 Blockers, 4 Majors, 3 Minors and 1
Nit against the record above — all real, none dismissed. Every finding is FIXED; none is deferred. The
"Findings Triage" section immediately below the QA Results records the decision and rationale for each.
Both Blockers concerned test coverage the story's own AC4/AC8 required and did not deliver — one shipped
repair (`readEmittedRuns`) with no guard at all reachable from the ordinary suite, and a second repair
(`assertXrefEntriesPointAtTheirObjects`) introduced a fresh silent-pass defect in the course of fixing a
different one, inside the very story chartered to remove that class (D-000.48, twice over). Both are now
guarded by committed, red-proofed tests. The 4 Majors were record-accuracy corrections (a false "only"
premise, an evidence claim that measured the wrong population, two discarded red-proofs, and a stale
call-site count) rather than code defects; all are corrected in place above. Full details, decisions and
rationale are in the "Finisher — Findings Triage" section following the QA Results.

---

## QA Results

## Review Summary

- **Reviewed by**: bmad-code-reviewer
- **Date**: 2026-08-25
- **Tree reviewed**: working tree over baseline `804eea1` (`M matrix_test.go, render_test.go, shaped_fixture_test.go, sprint-status.yaml` + `?? ` this story file). Tree restored byte-identical after every mutation; final `git diff --stat 804eea1` re-measured as `190 insertions(+), 145 deletions(-)` across 4 files — unchanged from the pre-review measurement.
- **Story Status Recommendation**: **Changes Requested**
- **Blockers**: 2
- **Majors**: 4
- **Minors**: 3
- **Nits**: 1

### Suites measured vs. suites read

**Measured (executed on darwin/arm64, native, this session):**

| command | result |
|---|---|
| `rtk proxy go test -count=1 -v ./...` | **557 PASS / 1 FAIL** all-occurrences, **338 / 1** top-level. Sole failure `TestCorpusMeetsP6ExerciseFloors`, stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` — identical to the story's figures and to the pre-story baseline. **Not a break.** |
| `rtk proxy go vet -tags matrix ./...` | exit 0 (compile gate only — not evidence any matrix leg ran) |
| `FOLIO_MATRIX_TARGET=darwin/arm64 rtk proxy go test -tags=matrix -count=1 -run TestTargetRenderHash .` | **pass** (`ok … 0.777s`) |
| `rtk proxy shasum -a 256` on the three pinned artifacts | **all three match exactly** |
| 3 red-proof mutations + 1 throwaway probe (2×2 against a `804eea1` worktree) | see Findings 1, 2, and the AC6 confirmation below |

**Read only, not run**: the three non-native matrix legs (`linux/amd64`, `linux/arm64`, `js/wasm`) for `TestTargetRenderHash`. They are correctly deferred to the Epic 2 boundary gate. I read `matrix_test.go` for teeth and found them (Finding 1 turns on that fact).

### Constraints — all confirmed

- **AC1 / no emitted bytes**: confirmed. Diff contains no file under `fixtures/` and no non-test `.go` file. `git status --porcelain -- folio-go/fixtures` is empty.
- **Three pinned digests unmoved**: confirmed by independent re-measurement.
  - `fixtures/shaped-text/expected.pdf` = `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`
  - `fixtures/expected-breaks/expected_breaks.json` = `a545e04259033429d2cf8d1bba07f3137f6c0a106d635e918d31eabd599324de`
  - `fixtures/multi-page/expected.pdf` = `66ce0ee477fa1ce5e42d51bcc87d859bcddafb3d2bb2ca6ade3e35d3f895869b`
  - **Both pending owner sign-offs may stand.**
- **Epic 2 gate owes exactly five (D-2.6.6)**: confirmed. `byte_neutrality_test.go` is absent from the diff; `declaredEpic2GateObligations` is byte-unchanged; no sixth obligation. `epic-2: backlog` stays (`sprint-status.yaml:49`).
- **D-000.55 / AC12**: satisfied. The banned compound phrase does not appear as a claim — its only occurrence (line 124) is the ruling table quoting it as banned. Executed legs are named with their target; unrun legs are named individually; `go vet -tags matrix` is described as a compile gate and explicitly disclaimed as evidence of a run.
- **`readEmittedRunsAllPages` retirement**: confirmed independently. No `func readEmittedRunsAllPages` anywhere in the tree; exactly three surviving references, all comments (`matrix_test.go:678`, `matrix_test.go:720`, `shaped_fixture_test.go:425`).

### AC roll-call

AC1 satisfied · AC2 **see Finding 5** · AC3 partially — **see Finding 3** · AC4 **see Finding 1** · AC5 **see Findings 1, 3** · AC6 conclusion sound (red-proved below), record wrong — **see Finding 5** · AC7 **see Findings 4, 7** · AC8 **see Findings 2, 8** · AC9 satisfied given AC3's axis list, but the axis list is incomplete (Finding 3) · AC10 **see Finding 6** · AC11 satisfied, independently re-measured · AC12 satisfied.

**AC6's named candidate is confirmed sound and I am recording the confirmation, not a finding.** I red-proved `TestReadEmittedRunsScansEveryTJAdjustment` (`shaped_fixture_test.go:565`) by mutating `readEmittedRuns` to drop one `Adjustments` entry per stream: the test **fails**. Its independent side (`strings.Split(string(b), "BT\n")`) was always a whole-file scan, never a per-stream re-derivation, so the repair did not collapse the two sides into the same scan. The developer's argument is correct and the disposition (not re-pointed) is right.

**On the "20 calls vs 200 calls" discrepancy: these are two distinct measurements, not one restated.** The in-code comment (`shaped_fixture_test.go`, `readEmittedRuns` docblock) records the **pre-fix** baseline — 20 calls, 9-or-24 against 33 actually present. The Delivery Log's AC4 entry records the **post-fix** verification — 200 calls, `map[33:200]`. Both are honest. Neither is committed as an assertion, which is Finding 1.

---

### Finding 1: The story's central repair ships with zero coverage in the ordinary suite — a full regression is invisible

- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/shaped_fixture_test.go:427-455` (`readEmittedRuns`); only multi-stream call site `folio-go/matrix_test.go:692`, file gated by `//go:build matrix` at line 1
- **Observation**: I red-proved this rather than inferring it. I reinserted a `break` after the first Tf-bearing stream — i.e. restored **exactly** the pre-repair defect — and ran the ordinary suite:

  ```
  rtk proxy go test -count=1 ./...
  ok  	github.com/panitw/folio/folio-go	4.936s      <-- GREEN with the defect restored
  --- FAIL: TestCorpusMeetsP6ExerciseFloors           <-- pre-existing, unrelated (internal/text)
  ```

  The same mutation is caught **only** under the matrix tag:

  ```
  FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .
  --- FAIL: TestTargetRenderHash (0.93s)
      matrix_test.go:1349: darwin/arm64: the multi-page leg carries 1 page-header run(s) and
      1 page-footer run(s) at their declared Y; a two-page document must carry exactly 2 of
      each — one per page
  ```

  Cause: of the 12 invocation sites in the delivered tree (Finding 5), exactly one renders a document with more than one text stream — `matrix_test.go:692`, `requireMultiPageIsGenuinelyMultiPage`. The other 11 render single-stream fixtures, on which "first stream found" and "every stream concatenated" are the same value. The story committed **no test at all**: AC11 states this outright — *"No test was added, removed, or re-pointed in the committed diff (all probes used during development were throwaway and deleted)."*

  AC4 requires the repair be *"Verified by a test that **runs** the repaired helper against the multi-page subject, not by reading it."* What was delivered is a **measurement** (200 calls, `map[33:200]`) in a probe that was deleted. AC5's reachability claim is true of the *file*; it is false of *exercise*.

  D-2.6.8's holding — that construction and guard are different properties and both belong — is also unmet on the determinism half: determinism now follows from `sort.Ints`, but nothing asserts that repeated calls agree.
- **Impact**: The correction to the module's most widely-shared PDF reader can regress in full — back to silent, nondeterministic single-stream truncation — and `go test ./...` stays green. Its only guard sits behind a build tag in a leg this story itself declares unrun and defers to the Epic 2 gate. This is the story's own defect class one level out (D-000.48): a sweep against silent truncation shipping a correction whose guard cannot see truncation.
- **Suggested Resolution**: Commit an untagged test in the `folio` package that renders (or loads) the multi-page subject, calls `readEmittedRuns`, and asserts (a) a run count equal to the sum of the per-stream parses — 33 at the time of writing — and (b) that N repeated calls in one process all agree. `multi_page_fixture_test.go` already renders that subject untagged (`renderMultiPage`), so this costs no new emitted bytes and no new fixture. Then re-run the story's own red-proof (reinsert the `break`) and confirm the new test reddens.
- **Related AC**: AC4, AC5, AC11

---

### Finding 2: The repair to `assertXrefEntriesPointAtTheirObjects` strictly *weakened* it — a zero-subsection xref table now passes silently

- **Severity**: Blocker
- **Category**: Correctness / Tests
- **Location**: `folio-go/render_test.go:871-874` (the `if bytes.HasPrefix(xrefBody, []byte("trailer")) { return }` at the top of the loop body)
- **Observation**: The trailer check is evaluated **before** `firstSubsection` is ever consumed, so an xref table that declares no subsections at all returns clean without making a single assertion. Proved as a 2×2 with a throwaway probe (deleted after the run) against a `804eea1` worktree and the delivered tree, using the buffer `"%PDF-1.7\nxref\ntrailer\n<< /Size 1 >>\nstartxref\n9\n%%EOF\n"`:

  | code | result |
  |---|---|
  | baseline `804eea1` | **FAIL** — `probe: xref subsection header "trailer" does not have two fields` |
  | delivered tree | **PASS** — `PROBE: zero-subsection xref table PASSED assertXrefEntriesPointAtTheirObjects` |

  The pre-repair function rejected this input; the repaired one accepts it.
- **Impact**: This story corrected *"reads one subsection and stops"* and, in the same function, introduced *"reads zero subsections and reports success"* — the identical class, in a well-formedness checker, in the artifact of the sweep that exists to remove that class. D-000.48 names exactly this outcome. Reachability, stated honestly: no committed golden and no document this serializer can emit reaches it — `internal/pdf/builder.go:105` and `internal/pdf/document.go:272` both hard-code `"xref\n0 "`, so no live assertion is vacuous today. But AC8's entire deliverable on this axis **is** the forward guard, and the forward guard has a false-pass path on its own degenerate input.
- **Suggested Resolution**: Fatal instead of returning when the trailer keyword is reached with `firstSubsection` still true — an xref table with no subsections is malformed and must be rejected as loudly as the pre-repair code rejected it. Re-run the probe above as the 2×2 red-proof; consider committing it, since (unlike the font-resource case) a synthetic in-memory buffer needs no fixture and no emitted bytes.
- **Related AC**: AC8

---

### Finding 3: AC5's premise is false — `readEmittedRunsAllPages` was not the module's only multi-page-aware reader

- **Severity**: Major
- **Category**: AC Conformance / Correctness of record
- **Location**: story AC5 (lines 376-384) and Delivery Log AC5; refuted by `folio-go/multi_page_fixture_test.go:162,220,332,502` at baseline `804eea1`
- **Observation**: AC5 asserts *"At baseline the module's only multi-page-aware reader, `readEmittedRunsAllPages`, sits in `matrix_test.go` behind `//go:build matrix`, so **no untagged test can call it**."* Verified against a `804eea1` worktree: `multi_page_fixture_test.go` begins `package folio` with **no build tag**, and already contained `mpParseToUnicode` (:162), `mpExtractRuns` (:220) and `splitPageContentStreams` (:332), driving `TestMultiPageHeaderAndFooterAppearOnEveryPage` (:502) — which splits the document into per-page content streams and asserts *per page* that exactly one header and one footer run appear at the declared `Tm`. That is **strictly more** multi-page-aware than the repaired `readEmittedRuns`, which returns a flat `[]emittedRun` carrying no page attribution.

  AC5's red-proof (`undefined: readEmittedRunsAllPages`) establishes only that *that identifier* was tagged; it does not establish the "only" claim, which is the load-bearing half.

  Consequence for AC3: the enumeration's `page-header runs` / `page-footer runs` rows name `readEmittedRuns` as the reaching helper and declare them *"covered by finding 1's repair"*. The actual committed per-page guard for that property is `mpExtractRuns` + `splitPageContentStreams`, which the enumeration never names and the sweep never examined.
- **Impact**: This is the same shape as Story 2.6's false *"`/Kids` is the only ref array"* claim that this story was created to generalise — an unmeasured "X is the only Y" hardening into an AC. It matters concretely: because the story believed the untagged suite had no multi-page reader, it did not look for the untagged multi-page reader stack that exists, and so left an entire family of shared-in-practice readers outside the sweep's enumeration. AC9's "enumeration fully discharged — zero instances deferred" is therefore a claim about an incomplete axis list.
- **Suggested Resolution**: Correct AC5's premise in the record, name the `mp*` reader family explicitly, and either sweep it behaviourally under this story's own predicate or defer it by name under D-000.17 (never silently). Re-derive AC3's header/footer rows against the helper that actually carries the property.
- **Related AC**: AC3, AC5, AC9

---

### Finding 4: AC7's byte-equivalence evidence measures a population of one golden, not eight

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: Delivery Log AC7; reach-points at `folio-go/shaped_fixture_test.go:213,268` → `:811,889,992,1127` and `folio-go/matrix_test.go:868`
- **Observation**: AC7's record states *"Full-suite run after the repair is unchanged (557/1 all-occurrences, 338/1 top-level — see AC11), confirming byte-for-byte-equivalent behaviour on all eight committed goldens."* I traced every path into `resourceType0Objects`: its only callers are `cidToGlyphForResources` (`:211`) and `toUnicodeForResources` (`:266`), and every call site of those two — `shaped_fixture_test.go:811,889,992,1127` and `matrix_test.go:868` — renders the **shaped-text** fixture. **The multi-page golden never reaches `resourceType0Objects` in any committed test, tagged or untagged.**
- **Impact**: An unchanged suite is evidence about one single-page fixture, not about eight goldens and specifically not about the only subject with more than one page object. The equivalence claim on the N>1 subject rests entirely on the deleted throwaway probe. The conclusion happens to be true (see Finding 7), but the stated proof does not establish it — a count that measured its filter, in the story whose AC2 exists to stop that.
- **Suggested Resolution**: Restate AC7's evidence to say what was actually measured — the helper is exercised on shaped-text only — and derive the multi-page equivalence from the serializer's construction (Finding 7) rather than from the unchanged suite total.
- **Related AC**: AC7, AC10

---

### Finding 5: The delivered tree has 12 `readEmittedRuns` call sites, not 11; AC2's table names no commit and every line number in it is stale; AC6 classifies a population that excludes the site this story created

- **Severity**: Major
- **Category**: AC Conformance / Convention (D-2.6.9 correction, D-000.14)
- **Location**: story AC2 Delivery Log table (lines 537-552) and AC6 table (lines 586-601); `folio-go/matrix_test.go:692`; in-code at `folio-go/shaped_fixture_test.go` (`readEmittedRuns` docblock) and `folio-go/matrix_test.go:722`
- **Observation**: Measured with the story's own discriminating pattern:

  | tree | invocations (excluding the declaration) |
  |---|---|
  | `804eea1` (baseline) | **11** — `shaped_fixture_test.go:534,776,854,905,1073`; `matrix_test.go:782,829,874`; `wrapped_text_fixture_test.go:79`; `segment_origin_test.go:95`; `first_baseline_acceptance_test.go:291` |
  | delivered working tree | **12** — `shaped_fixture_test.go:567,809,887,938,1106`; `matrix_test.go:692,759,806,851`; `wrapped_text_fixture_test.go:79`; `segment_origin_test.go:95`; `first_baseline_acceptance_test.go:291` |

  The story's own AC5 work created the twelfth: retiring `readEmittedRunsAllPages` re-pointed `requireMultiPageIsGenuinelyMultiPage` back onto `readEmittedRuns` at `matrix_test.go:692`.

  Three separate problems follow.

  1. **Citation (D-2.6.9's appended correction).** The Delivery Log presents its table as *"Real inventory, re-confirmed at this commit"* but names no commit, and every one of the eight `shaped_fixture_test.go` / `matrix_test.go` line numbers it lists is a `804eea1` line number that has moved in the delivered tree. The correction requires an enumeration to name the commit it was taken at and to have been taken against a tree with no uncommitted changes in the counted files — or to say explicitly that it was not. It does neither. The arithmetic was right *at the baseline*; the citation is not.
  2. **AC6's universal claim is false of the delivered tree.** The classification table concludes *"every fixture any of the 11 sites renders carries exactly one text-bearing content stream."* The twelfth site renders `fixtures/multi-page/`, which carries **two**. The *conclusion* survives — that site wanted all-stream reading and now gets it — but the record's population and its universal statement do not.
  3. **Task 8 is not fully discharged.** Two in-code comments still cite 11 in the delivered tree: `shaped_fixture_test.go` (*"every fixture all eleven of this function's call sites use, at the time of this repair"*) and `matrix_test.go:722` (*"its other eleven call sites"*). Task 8 required leaving no comment citing a figure this story corrected.
- **Impact**: The story's headline demonstration is that its own opening count measured its filter. Delivering a record whose count is stale for the artifact it describes — and stale specifically because of the story's own edit — reproduces the pattern at the next level. It also makes the record unusable as the "known map" Story 2.7 is supposed to write its new call sites against, since 2.7 will read "11 sites, all single-stream" and find twelve, one of them multi-stream.
- **Suggested Resolution**: Re-take the enumeration at the finishing commit, name that commit, state the tree's cleanliness, list both the 11-at-baseline and 12-at-delivery figures with the reason for the delta, extend AC6's classification to the twelfth site, and correct the two in-code comments.
- **Related AC**: AC2, AC6, Task 8

---

### Finding 6: AC10 reports "guards forward: 2" and "guards redundant: 1" for guards that do not exist — and the record declares "no red-proof available" for two red-proofs the developer built and then deleted

- **Severity**: Major
- **Category**: AC Conformance / Tests
- **Location**: Delivery Log AC10 table (lines 648-656); AC7 (line 619-621) and AC8 (line 635-639)
- **Observation**: The diff adds **no test and no assertion** — only helper bodies, imports and comments. D-000.24's permitted category is *"forward **guard** with no available red-proof"*: a guard is a committed assertion. Two of AC10's four headline numbers therefore count repairs that ship with nothing guarding them.

  The record refutes its own "no red-proof available" declaration:

  - AC7: *"Discriminating power was proven with a throwaway synthetic two-page subject carrying two different font resource names (`/FaceA`, `/FaceB`); before the repair only one would have survived, after the repair both do. The probe was deleted after the run."*
  - AC8: *"Verified with a throwaway synthetic two-subsection buffer: a bad second-subsection entry offset now fatals … and a well-formed second subsection now passes."*

  In both cases a subject that **expresses the defect** was constructed and observed to discriminate — and then thrown away. The scope fence forbids new **emitted bytes**, new fixtures and new `matrixDocuments` entries. An in-memory synthetic buffer inside a `_test.go` file is none of those. I demonstrated this myself in Finding 2: the xref red-proof takes one 50-byte literal and no fixture.
- **Impact**: D-000.50 asks whether *any* subject can express the defect; the answer recorded is "no committed golden can", but the operative answer for guard-building is "yes, and I built one". Declaring a guard forward under D-000.24 when an available red-proof was in hand and discarded inverts the ruling's purpose, and it converts three demonstrated behavioural properties into three prose claims. Together with Finding 1, the net position is that all three of this story's repairs ship with no committed assertion whatsoever — which is why the whole diff can be reverted with the default suite staying green.
- **Suggested Resolution**: Commit the two synthetic probes as ordinary untagged tests (they need no fixture and move no bytes), then re-derive AC10's four numbers honestly. If a repair genuinely ships without a guard, AC10 should say so with a fifth count — *repairs shipped unguarded* — rather than reporting them under a guard heading.
- **Related AC**: AC7, AC8, AC10

---

### Finding 7: `resourceType0Objects`' repair is inert across the engine's entire output space, not merely "no committed subject" — and the repaired helper is page-complete but still not page-aware

- **Severity**: Minor
- **Category**: Correctness / Convention (D-000.50)
- **Location**: `folio-go/shaped_fixture_test.go:134-193`; serializer evidence at `folio-go/internal/pdf/textdoc.go:120-146` and `:216-222`
- **Observation**: The record grounds its D-000.50 answer in a property of one fixture: *"both of `fixtures/multi-page/`'s page objects carry a byte-identical `/Font << /NotoSans 3 0 R >>`."* The stronger and correct reason is a property of the **serializer**: `textdoc.go` collects `faceNames` document-globally (`slices.Sorted(maps.Keys(faces))`), rejects escaped-name collisions **across the whole document** (`resourceNameCollisionError`, `:143`), and then writes the **identical** `/Font << … >>` dictionary into every page object in the same sorted order (`:216-222`). Every page's font dict is byte-identical for **any** document this engine can produce.

  Separately: the repaired helper merges every page's dict into a single `map[string][]byte` (`out[name] = type0`), so two pages mapping the same resource name to different objects would still be last-writer-wins, silently — the helper became page-*complete*, not page-*aware*. That cannot fire today for the same serializer reason (names are globally assigned and globally collision-checked), so I am recording it as a shape note, not a live defect.
- **Impact**: Understating the constraint misleads the next reader. Someone reading "no committed subject can express this" will reasonably conclude that authoring a fixture arms the guard; it does not — the serializer would have to change first. This also sharpens AC7's honest statement of value from "changes no test outcome today" to "cannot change any test outcome until the emitter writes per-page font dictionaries", which is the fact a future story needs.
- **Suggested Resolution**: Restate AC7's D-000.50 reason in terms of `textdoc.go`'s global face set and per-page-identical dictionary, and note the `map[string][]byte` collapse so a later page-aware repair knows the return type is the next thing to change.
- **Related AC**: AC7

---

### Finding 8: The repaired xref loop encodes a contiguity invariant the PDF format does not require

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/render_test.go:896-898`
- **Observation**: `if !firstSubsection && start != nextObjNum { t.Fatalf("%s: xref subsection header %q does not continue from object %d", …) }`. Cross-reference subsections exist precisely to describe **non-contiguous** object ranges (ISO 32000-1 §7.5.4); a table reading `0 1` followed by `5 3` is well-formed. The repair rejects it.
- **Impact**: The forward guard AC8 delivers would fatal on the legal form of the very thing it was built to examine, with a message that misattributes the cause. Inert today (this serializer emits one contiguous subsection), but it is a forward guard, so its only value is on input that does not exist yet — and it is wrong about a share of that input.
- **Suggested Resolution**: Require each subsequent subsection's `start` to be **greater than or equal to** the previous subsection's end (monotonic, non-overlapping) rather than exactly equal, or drop the continuity check and keep the per-entry offset validation, which is the property the function actually exists to assert.
- **Related AC**: AC8

---

### Finding 9: The "Flagged, not fixed" instruction was not discharged in the record

- **Severity**: Minor
- **Category**: AC Conformance / Maintainability
- **Location**: story lines 508-510; unaddressed in the Delivery Log; code at `folio-go/shaped_fixture_test.go:148`
- **Observation**: The story requires that `resourceType0Objects`' `"/Type /Page "` trailing-space substring match — *"the exact substring hazard `countPageObjects` was repaired for under Nit 26"* — be **noted in the record**, with repair optional. Grepping the story file, the string appears only in the pre-development *Measured findings 2* (line 230) and *Flagged, not fixed* (line 508) sections. No Delivery Log entry mentions it. The repair edited the surrounding loop and left `bytes.Contains(obj, []byte("/Type /Page "))` untouched without comment.
- **Impact**: Small in itself, but this is the mechanism by which a known hazard stops being tracked: the story that was told to carry it forward did not, and the next story has no record that it was ever considered. Not fixing it was permitted; not recording it was not.
- **Suggested Resolution**: Add one Delivery Log line noting the hazard was seen, deliberately not repaired, and why — or append it to `deferred-work.md` by name under D-000.17.
- **Related AC**: AC9

---

### Finding 10: Orphan doc comment for a deleted function

- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/matrix_test.go:720-734`
- **Observation**: A 15-line doc comment opening `// readEmittedRunsAllPages was this file's multi-page-aware sibling of …` is attached to no declaration — the function it documents was deleted. The record calls this deliberate (a tombstone so a future reader does not re-add the sibling), which is a defensible choice.
- **Impact**: A godoc-shaped comment with no declaration reads as a truncated file to a reader landing mid-file, and this particular block also carries the now-stale "eleven call sites" figure flagged in Finding 5.
- **Suggested Resolution**: If keeping the tombstone, prefix it so it does not read as a doc comment (e.g. a `// NOTE (Story 2.6a):` opener) and correct the call-site figure.

---

## Finisher — Findings Triage

Every finding below is **FIX**. None is dismissed; none is deferred — every one is either a real code
defect this story's own charter covers (Blockers) or a record-accuracy defect in this story's own Delivery
Log (Majors, Minors, the Nit). Rationale is stated per finding; where a fix touches code, the file and
function are named. Full technical detail for each correction lives inline at the AC it corrects (linked
below) rather than repeated here.

| # | Severity | Decision | Rationale |
|---|---|---|---|
| 1 | Blocker | **FIX** | AC4 required the repair be verified by a **committed** test running the repaired helper against the multi-page subject; only a deleted probe existed. In scope by the story's own acceptance criterion. Fixed: two tests committed in `shaped_fixture_test.go` (correctness cross-check + determinism), both red-proofed against the delivered defect and the true baseline defect. See AC4. |
| 2 | Blocker | **FIX** | The xref repair's own first draft introduced a genuine new silent-pass defect (zero-subsection tables) — D-000.48 (a correction sweep can introduce a fresh instance of the class it corrects) names this exact failure mode, and it occurred inside the story chartered to eliminate it. Not fixing it would ship a known-worse regression in a well-formedness checker. Fixed in `render_test.go`'s `assertXrefEntriesPointAtTheirObjects`, red-proofed by reverting and confirming the new committed test reddens. See AC8. |
| 3 | Major | **FIX** | AC5's "only multi-page-aware reader" claim is factually false and was never measured before being written — the exact unmeasured-"only" pattern D-2.6.9 exists to generalise away, reproduced inside the story doing the generalising. Correcting the record (not the code — the `mp*` family needs no repair) is squarely this story's own charter: AC2/AC3's enumeration is the story's product. See AC5, AC3, AC9. |
| 4 | Major | **FIX** | AC7's "confirmed on all eight goldens" claim measured the wrong population (shaped-text only; the multi-page golden is unreachable from `resourceType0Objects`). A false evidentiary claim in the Delivery Log this story itself produces is squarely in scope to correct. See AC7. |
| 5 | Major | **FIX** | The call-site inventory is stale for the tree it claims to describe, and the story's own AC5 edit is what moved it (12→14 once the finisher's own new tests are counted). AC2/AC6 are the story's own record of its own scope; leaving them stale defeats their purpose as "the known map" Story 2.7 will read. Re-taken at the finishing commit, named, itemised. See AC2, AC6. |
| 6 | Major | **FIX** | Two working, discriminating in-memory red-proofs were built during development and discarded, then the guards they would have proven were declared "forward" under D-000.24 — which inverts that ruling's purpose (forward guard = no available red-proof; one was available and built). Committing them is cheap (no fixture, no emitted byte) and directly required by D-000.50's own test ("can ANY subject express the defect" — yes). See AC7, AC8, AC10. |
| 7 | Minor | **FIX** | The reviewer's `textdoc.go`-level justification is strictly stronger and correct where the original per-fixture justification was true-but-under-derived; adopting it costs nothing and gives the next reader the fact that actually matters (the serializer, not the fixture set, is what makes the defect unreachable). The page-complete-vs-page-aware shape note is a one-paragraph addition with no code change (not a live defect today). See AC7 (docblock in `shaped_fixture_test.go`, and the Delivery Log). |
| 8 | Minor | **FIX** | The repaired xref contiguity check rejected an ISO-32000-1-legal input its own review demonstrated it shouldn't. It shares a function with Blocker 2 and was being edited regardless; leaving a known-wrong check in place while fixing its sibling in the same loop would be inconsistent. Fixed in `render_test.go`, proven both directions (accepts a gap, still rejects an overlap) with committed tests. See AC8. |
| 9 | Minor | **FIX** | The story's own "Flagged, not fixed" section required a Delivery Log line; none existed. One-line, zero-risk discharge of an instruction already in the story. See AC8 (new bullet). |
| 10 | Nit | **FIX** | Cheap, directly suggested by the reviewer, and the tombstone comment's stale count is the same defect class as Finding 5. Fixed in `matrix_test.go`: the comment now opens `// NOTE (Story 2.6a, tombstone — deliberately not a doc comment ...)` and its count is corrected. |

No finding was **DISMISS**ed or **DEFER**red: every Blocker was a real gap inside this story's own stated
acceptance criteria (AC4, AC8), and every Major/Minor/Nit was a correction to a claim this story's own
Delivery Log makes about itself — none reached into `render.go`, `internal/pdf`, `internal/layout`, or any
other file the scope fence places out of bounds, and none required a new fixture, a new `matrixDocuments`
entry, or a sixth Epic 2 gate obligation. The scope fence held throughout: the finisher's diff, like the
developer's, touches only `*_test.go` files, the story file, and `sprint-status.yaml`.
