# Story 1.3: Guardrails that fail the build

Status: done

| | |
|---|---|
| **Story key** | `1-3-guardrails-that-fail-the-build` |
| **Baseline commit** | `f9c27b3` — Story 1.2 shipped: `.github/workflows/matrix.yml`, `hashmatrix/`, `folio-go/matrix_test.go`. **No `lint/` directory, no `folio-designer/`, no `folio-go/fonts/`, and no `testdata/` directory anywhere in the repo exist yet — this story creates `lint/` and the fixture trees.** The working tree at `f9c27b3` additionally carries the uncommitted Story 1.3 rulings in `folio-mvp-decision-log.md` and the new, untracked `deferred-work.md`; both ship in **this story's** commit. |
| **Epic** | 1 — A Go developer can render a deterministic PDF |
| **Source** | `_bmad-output/planning-artifacts/epics.md` §Story 1.3 (lines 482–508) |
| **Contract** | PRD **NFR1.c** (restricted numeric surface, "enforced by an import-restriction lint, not by convention") and **NFR1.d** (no unordered iteration reaches output) · `SPEC.md` **CAP-13** |
| **Invariants** | **AD-1 and AD-26 are the governing ones.** AD-3, AD-5, AD-6, AD-23 adjacent. |
| **Owns** | **DW-1** (`deferred-work.md`) — the fixture-path override. Closed by the general seam this story builds, not by a one-off patch. |
| **Registers** | DW-2, DW-3, DW-4 are already written into `deferred-work.md` by this story's rulings and stay **open** with their stated owners. |
| **Local toolchain** | Go 1.26.0 darwin/arm64, `GOTOOLCHAIN=auto` honouring `folio-go`'s `toolchain go1.26.0` pin (verified in this run) |

> **Stop here — do not commit, do not branch, do not set `done`.** The final task of this story is
> "story file, decision log, sprint status → `review`". Committing belongs to the finisher, after
> review.
>
> **There are no ADRs in this project.** The `AD-N` invariants in `ARCHITECTURE-SPINE.md` are the
> ADR equivalent and live in that one file. Cite `AD-N` plus a SPEC clause or a story AC — never an
> ADR path.
>
> **If something in this file looks like it contradicts a decision-log entry, stop and surface it.
> Do not resolve it by choosing** (D-1.2.6). The ruling governs and this file is wrong.
>
> **This story is Epic 1's most exposed to D-1.2.5**, because it is built almost entirely out of
> claims about what fires and what does not. Per D-1.3.7's concrete instruction: **every "this guard
> fires on X" claim in the Delivery Log must cite the retained fixture that demonstrates it — never
> a mutation applied and reverted.** That is what D-1.3.3's fixture architecture is *for*.
>
> **This file's own findings produced four rulings before development started** —
> `D-1.3.3 (amended)`, D-1.3.8, D-1.3.9, D-1.3.10. They are applied throughout and are **not** open
> questions. **DN-1 is ruled; nothing is outstanding.**
>
> **This story ships a D-000.6 spine amendment** (Task 9): the spine's §Source tree gains `lint/`.
> One line, same shape as `hashmatrix/`'s, in **this story's own commit**, changing only the tree
> block and leaving every invariant's **Binds** and **Prevents** lines untouched.
>
> **Heavy-test override does NOT apply.** This story's deliverable is not hash-shaped; the
> cross-target matrix is due at the Epic 1 boundary under D-000.4. The Delivery Log must name the
> unrun suites explicitly.

---

## In plain terms (read this first if you just want the gist)

Everything this project promises rests on a rule that is easy to break by accident: the code that
produces a document must behave identically on every machine, every time. Until now that was kept
by two narrow automatic checks and people remembering. This story replaces remembering with
machinery — and an independent review before shipping made it stronger than first written.

Three new checks landed. One forbids the document-producing code from touching anything that
varies between machines — the clock, the operating system, the network, randomness, and the
trigonometric functions whose answers differ between chip families. One forbids walking through a
keyed collection there, because the language shuffles that order, and shuffled order means a
different document. One reads every outside package the project depends on, fails the build on an
incompatible licence, and writes out the full list, refusing to finish if any licence is unknown.

The reviewer found the keyed-collection check could be fooled: it understood one piece of code at a
time, so a collection described in one place and used in another slipped past unnoticed — and that
split is how the real code is already organised. Fixing it meant giving the checking tool a fuller
understanding of the code it inspects, so the tool now depends on one small outside library it did
not need before; the document-producing code being checked still depends on nothing new. A few
smaller gaps were closed the same way, and an existing, too-wide check was narrowed to where it is
needed, its proof rebuilt to demonstrate the narrowing itself, not an unrelated coincidence.

Deliberately not done here: the matching licence check for the browser application, and publishing
the licence list with a release — both asserted absent, so either appearing forces the work.

Done looks like: the checks run locally and in the automated build, the code is clean, and every
broken example is caught. One thing will look wrong later and is not: a licence list that is no
longer empty and names the checking tool's own new dependency — the checker covering itself is
correct, not circular.

## Story

As an integrating Go developer,
I want the determinism and licence rules enforced mechanically rather than by convention,
So that a reasonable-looking commit cannot quietly erode the property the whole library rests on.

**Covers:** NFR1.c, NFR1.d · AD-1, AD-26

---

## Settled decisions — apply these, do NOT re-open or surface them

Six rulings were obtained **before** this story file was written, specifically so the developer would
not have to guess. **Quote them and cite the decision ID — do not paraphrase.** Paraphrasing a ruling
is what produced D-1.3.2 in the first place: Story 1.1's AC6 restated D-1.1.b in its own words,
dropped its scope and its carve-out, and shipped a guard that is one story away from breaking Story
1.4 (measured — see **F-2**).

| Ruling | What it settles |
|---|---|
| **D-1.3.1** | The AD-1 import lint exempts `_test.go` for `os`/`testing`/`path/filepath`/`embed` **only**. `time`, `math/rand`, `net` and `math` transcendentals stay banned in tests too. Key on the `_test.go` **suffix**, never a directory prefix. **Two** violating fixtures for this rule, not one. |
| **D-1.3.2** | Narrow the numeric-formatting guard to exactly `folio-go/internal/pdf/`. **Delete the module-wide rule outright** — do not keep it with an exemption list. Red-prove in **both** directions. `epics.md` is clean, so **D-000.6 does not fire**; append a `Superseded` note to the committed Story 1.1 file and leave lines 202, 327 and 678 intact. |
| **D-1.3.3** | Every guard becomes a **pure function taking a target directory and returning findings**, with exactly two callers: a production scan asserting **zero**, and a fixture scan asserting **exactly the named findings**, by file and rule. Every scanner skips any directory named `testdata`. Fixtures at `folio-go/testdata/lint/<rule>/`. **Each rule ships both polarities.** This story owns and discharges **DW-1**. |
| **D-1.3.4** | Ship the Go module-graph licence half now, complete. Defer the `folio-designer/` lockfile half to Story 5.1 via an **asserted absence**; a conditional "check it if present" is **rejected**. Same for the OFL 1.1 font text (Story 2.2) and the Apache-2.0 `pdfjs-dist` NOTICE (Epic 5). The manifest is **generated and asserted complete** at 1.3; *publication* is deferred (DW-3). D-000.5 corrected the AC's `designer/` to **`folio-designer/`**. |
| **D-1.3.5** | **Total ban on `range` over a map value in any non-test file under `folio-go/internal/`.** No reachability analysis. Detection **exact and type-based**; a syntactic guess is not acceptable. The escape hatch — `for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }` — ships in this story and **the failure message names it verbatim**. |
| **D-1.3.6** | The **new** guards live in a standalone tool in its own repo-root module `lint/`, `module github.com/panitw/folio/lint`. Not a `go test` in `folio-go`, **not** `folio-go/cmd/`. The **two existing guards stay where they are.** Binding invariant: (a) exact detection with no false positives, (b) **no dependency added to `folio-go`'s module graph**, (c) locally runnable with CI merely invoking it. `golang.org/x/tools/go/packages` is expected; an exact dependency-free `go/types` path is a **permitted refinement, not a deviation**. Consolidating the existing guards is a `DECISION NEEDED`, never a judgement call in the diff. |
| **D-1.3.3 (amended)** | Amends the row above, **on this story file's F-5**. A checker returns **`(findings, error)`**. The production caller **must fail on a non-nil error, separately from and before asserting zero findings.** The fixture caller asserts findings **by file and rule, never by count**. Violating fixtures are **valid syntax with forbidden semantics**, so a fixture can never redden a guard by being unparseable rather than non-compliant. *As originally ruled, a crashed scan and a clean scan returned the same value.* |
| **D-1.3.8** | Resolves this file's DN-1: **option (b)**, a stub fixture module graph, **and the conflict does not exist.** AD-26 forbids a dependency **carrying** copyleft; obligations attach to **code received under a licence**, not to a text file that says "GPL". A stub we authored ourselves, zero lines of third-party code, whose `LICENSE` is fixture data, creates no licensee relationship and transfers no copyrighted work. **"A module whose `LICENSE` file declares GPL" is not "a GPL dependency."** Fixtures at `lint/testdata/licence/`; **three** graphs, not two; supplement with (c)'s classifier unit tests; **do not adopt (a)**. |
| **D-1.3.9** | **AC18 confirmed — do not narrow.** AD-26's scope line is **Binds: all**; "the whole module graph" was written when exactly one module existed, so the scope line resolves the ambiguity, not the singular noun. Refinement: the manifest **labels each dependency with the module it serves and whether that module is shipped or build-time-only**. The ban stays uniform. A future story genuinely needing a copyleft build-time tool is an **owner escalation, not a lead call**. |
| **D-1.3.10** | **F-6's selector rule confirmed, and tightened twice.** It is AD-1's own text, not a concession to shipped code. **(1) Allow-list, not deny-list** — any `math` function call whose name is not one of AD-1's seven is a violation. **(2) Non-call references are judged by value kind** — integer-limit constants permitted; **float-valued constants (`Pi`, `E`, `MaxFloat64`, `SmallestNonzeroFloat64`) are not.** Match on the **AST, never a regex**, resolving the import alias. |
| **D-1.3.7** | Two paragraphs reproduced verbatim in **Dev Notes** below, plus the evidence rule for this story. |

Also binding and already applied here: **D-000.4** (per-epic heavy-test cadence; no override for this
story), **D-000.5** (`folio-designer/`, not `designer/`), **D-000.6** (a ruling that makes a canonical
document wrong amends it in the story's own commit), **D-1.1.b** (AD-3 governs numeric
*representations*), **D-1.2.3 (amended)** (`hashmatrix/` holds the probe alone), **D-1.2.5** (a finding
that disclaims measurement may not warrant anything), **D-1.2.6** (surface ruling conflicts, never
arbitrate them in the diff).

---

## Corrections and measured findings (verified in this run, at `f9c27b3`)

Every figure below was produced by running the named command in this run. Nothing is carried forward
from a previous story, and nothing here is inferred (D-1.2.5).

### F-1 — MEASURED: the baseline is green, and both existing guards pass

```
$ cd folio-go && go test ./... ; go vet ./...
ok  github.com/panitw/folio/folio-go
ok  github.com/panitw/folio/folio-go/internal
ok  github.com/panitw/folio/folio-go/internal/geom
ok  github.com/panitw/folio/folio-go/internal/pdf
42 tests passed across 4 packages; go vet reports no issues
```

Every red observed below is caused by the probe described in that finding and by nothing else.

### F-2 — MEASURED: the numeric-formatting guard blocks Story 1.4 **today**. This is D-1.3.2's whole point

A scratch package `folio-go/internal/template/version.go` containing exactly the shape Story 1.4's
AC requires — an error naming the declared and supported versions — was created, measured, and
removed:

```go
package template

import "fmt"

func VersionError(declared, supported string) error {
	return fmt.Errorf("template declares version %s; supported version is %s", declared, supported)
}
```

```
--- FAIL: TestNumberFormattingIsConfinedToNumbersGo (0.00s)
    emit_source_test.go:191: forbidden strconv/fmt formatting call(s) found outside numbers.go (AC6, D-1.1.b):
        folio-go/internal/template/version.go:6: fmt.Errorf
```

This is the D-1.3.2 "must now pass" direction, demonstrated red at baseline rather than assumed.
**AC8 requires the same shape to be green after the narrowing, and to stand in the tree as a retained
compliant fixture** — not as a mutation someone applied and reverted.

### F-3 — MEASURED: a fixture under `folio-go/internal/` reddens the suite today. This is D-1.3.3's ordering trap

`folio-go/internal/testdata/lintprobe/bad.go` containing `func f() float64 { return 1.5 }` was
created, measured, and removed:

```
--- FAIL: TestNoFloat64UnderInternal (0.00s)
    arch_test.go:145: float64/float32 found under internal/ (forbidden by AD-23):
        testdata/lintprobe/bad.go:3: identifier float64
        testdata/lintprobe/bad.go:3: untyped floating-point literal 1.5
```

Confirms D-1.3.3's guardrail exactly: **neither scanner skips `testdata` today** (measured: `grep -rn
testdata folio-go --include='*.go'` returns **no matches** at `f9c27b3`). The `testdata/` skip and the
two-caller shape must land **before the first fixture**, or the first fixture is a build break. This
is why Task 1 and Task 2 precede Task 4.

### F-4 — MEASURED: `folio-go/testdata/` is invisible to the go command but **visible to the existing numeric guard**

`folio-go/testdata/lint/probe/bad.go` calling `strconv.Itoa` — i.e. a fixture at exactly the location
D-1.3.3 rules — was created, measured, and removed:

```
$ go build ./...   → exit 0, no output
$ go vet ./...     → no issues
$ go test ./...
ok      github.com/panitw/folio/folio-go
ok      github.com/panitw/folio/folio-go/internal
ok      github.com/panitw/folio/folio-go/internal/geom
--- FAIL: TestNumberFormattingIsConfinedToNumbersGo (0.00s)
    emit_source_test.go:191: forbidden strconv/fmt formatting call(s) found outside numbers.go (AC6, D-1.1.b):
        folio-go/testdata/lint/probe/bad.go:5: strconv.Itoa
```

Two things measured at once. **The good half:** the go command ignores `testdata/` universally — the
fixture is never compiled, never vetted, never a package. That is precisely the property D-1.3.3
wants and it holds for free. **The bad half:** a hand-rolled `filepath.WalkDir` does *not* inherit
that convention, so `emit_source_test.go` parses and reports the fixture anyway. The skip is not
decorative.

Note the interaction with D-1.3.2: after the narrowing, the numeric guard's production scan is rooted
at `internal/pdf`, so `folio-go/testdata/` falls out of its scope by root alone. **The `testdata/`
skip is still required** — D-1.3.3 mandates it for every scanner, and it is what makes a fixture safe
regardless of which root a future caller picks.

### F-5 — MEASURED: an unparseable fixture aborts the walk before any finding is appended, which under the ruled refactor becomes "zero findings"

`folio-go/internal/testdata/lintprobe/broken.go` containing `package lintprobe` / `func f( {` was
created, measured, and removed:

```
--- FAIL: TestNoFloat64UnderInternal (0.00s)
    arch_test.go:134: walk internal/: .../broken.go:2:9: expected ')', found '{' (and 3 more errors)
--- FAIL: TestNumberFormattingIsConfinedToNumbersGo (0.00s)
    emit_source_test.go:180: walk folio-go/: .../broken.go:2:9: expected ')', found '{'
```

Both scanners abort the entire walk on the first parse error and return **before appending any
finding**. Today that surfaces as a distinct `t.Fatalf` message, so it is visible. Under D-1.3.3's
two-caller shape it becomes dangerous: a pure function that returns `(findings, err)` and a fixture
caller that ignores `err` would report "0 findings" for a fixture tree it never actually read — and
the production caller, which asserts **zero**, would pass. **AC5 exists because of this measurement.**
Two consequences for fixture design: a violating fixture must be **syntactically valid Go that is
semantically forbidden**, and the scan error must never be collapsed into the findings channel.

### F-6 — MEASURED: `math` must be banned by **selector**, not by import path

`math` is imported under `folio-go/internal/` **today**, in two `_test.go` files:

```
folio-go/internal/pdf/numbers_test.go:4:   "math"     → math.MinInt64, math.MaxInt64 (6 sites)
folio-go/internal/geom/scale_test.go:4:    "math"     → math.MinInt64, math.MaxInt64 (9 sites)
```

Every use is a **constant**, not a transcendental. AD-1's Rule bans "any `math` transcendental
(`Sin`, `Cos`, `Tan`, `Log`, `Exp`, `Pow`, `Sinh`, `Erf`, …)", and D-1.3.1's `_test.go` clause bans
transcendentals in tests too — neither bans the `math` import. An import-path-level ban on `math`
would redden the shipped suite on its first run. AC12 states the rule as a selector rule for this
reason, and AC13 keeps a compliant near-miss (`math.MaxInt64`) standing in the tree to hold it there.

No non-test file under `internal/` imports `math` at all — `internal/geom/scale.go:3` spells the
constant as a literal precisely so it needs no import.

**Ruled at D-1.3.10, and the reason is stronger than "the shipped code needs it":** the selector rule
is AD-1's *own text*. AD-1 bans four **package paths**, then names a class of **functions**, and its
next sentence enumerates seven tolerated `math` **functions** — a rule cannot both ban a package and
list which of its functions are permitted. D-1.3.10 then tightens twice: the test inverts to a
**closed allow-list** (AD-1's transcendental list ends in "…" and is not decidable; its allow-list
is), and **non-call references are judged by value kind**, so `math.Pi` is a violation while
`math.MaxInt64` is not.

**One further measured case, and it is a near-miss the fixture tree must carry.** The only
`math.Round` anywhere under `internal/` is at **`internal/geom/scale.go:31`, inside a comment** — the
line reads *"no call to math.Round"* — alongside five `math.MinInt64` comment mentions in the same
file. A regex over source text reports it; an AST walk does not. AC13 requires a fixture of exactly
this shape.

### F-7 — MEASURED: there is no map range under `internal/` today, so D-1.3.5 costs zero

Every `range` under `folio-go/internal/` at `f9c27b3`, by file:

| File | `range` sites | Subject |
|---|---|---|
| `internal/pdf/numbers.go:45` | 1 | English prose in a comment ("negative range"), not a `RangeStmt` |
| `internal/geom/scale.go:47` | 1 | same — a comment |
| `internal/pdf/emit_source_test.go` | 1 | `file.Imports` — a slice |
| `internal/pdf/numbers_test.go` | 3 | table-test slices |
| `internal/geom/scale_test.go` | 1 | a table-test slice |
| `internal/arch_test.go` | 0 | holds a `map[string]bool` but indexes it; never ranges it |

**Zero map ranges, in test and non-test files alike.** The production scan asserts zero and is
non-vacuous only because AC17's counting guard requires it to have parsed the real packages.

### F-8 — MEASURED: `folio-go`'s module graph is exactly one line, and there is nothing else to licence yet

```
$ cd folio-go   && go list -m all → github.com/panitw/folio/folio-go
$ cd hashmatrix && go list -m all → github.com/panitw/folio/hashmatrix
$ ls folio-designer  → No such file or directory
$ ls folio-go/fonts  → No such file or directory
```

So at `f9c27b3` the licence check's Go half resolves three repo modules (the third being `lint/`,
which this story creates) and, apart from `lint/`'s own dependency, **zero external modules**. The
manifest is not empty theatre: it is complete for what exists, and AC19 makes an unresolvable licence
a build failure so it stays complete as the graph grows. This is also why AC20 matters — for a while
the only external module in the manifest will be the lint's own.

### F-9 — CORRECTION: the two existing guards report finding paths on **different bases**

Measured in F-2 and F-3: `arch_test.go` reports `testdata/lintprobe/bad.go` (relative to
`folio-go/internal/`, its walk root) while `emit_source_test.go` reports
`folio-go/internal/template/version.go` (relative to the repo root). Under D-1.3.3 the fixture caller
asserts **exactly the named findings, by file and rule**, so an undefined path base makes those
assertions unwritable and, worse, quietly rewritable when a root changes. **AC4 fixes the base to the
scanned target directory** — the one base that is stable when the same pure function is pointed at
the real tree and at a fixture tree.

### F-10 — the fixture trees are safe where D-1.3.3 puts them, and must not drift under `internal/`

`folio-go/testdata/lint/<rule>/` is outside `folio-go/internal/`, so the `internal/`-rooted `float64`
guard never reaches it (F-3's red required placing the probe *under* `internal/`). The numeric guard
reaches it today (F-4) and stops reaching it once D-1.3.2's narrowing and D-1.3.3's skip both land.
**No fixture is placed under `folio-go/internal/` at any point in this story.** The `lint/` module
reads the fixture trees by repo-root-relative path at test runtime — the same mechanism AD-21 and
D-000.5 already fix for `fixtures/`, and the reason `repoRootFromTest` exists.

---

## Acceptance Criteria

AC1–AC6 are the seam and **land first** (F-3's trap). AC7–AC9 are D-1.3.2. AC10–AC13 are the AD-1
import lint. AC14–AC17 are map iteration. AC18–AC22 are the licence check, **extended by AC29–AC31**
(D-1.3.8, numbered last so existing AC ids stayed stable). AC23–AC27 are placement. AC28 is the
evidence rule.

### The checker seam (D-1.3.3) — Tasks 1–3, before any fixture exists

**AC1 — every guard is a pure function over a target directory, with exactly two callers.**
**Given** any guard in this story, new or existing
**When** it is implemented
**Then** its detection logic is a function taking a **target directory** and returning
**`(findings, error)`** — no `*testing.T` parameter, no hard-coded root, no repo-root discovery
inside it
**And** it has exactly two callers: a **production scan** over the real tree asserting **zero**
findings, and a **fixture scan** over a retained fixture tree asserting **exactly the named
findings**, by file and by rule — neither a subset nor a superset
**And** the fixture caller asserts **by file and rule, never by count**: a scan that finds the right
*number* of wrong things must fail. *D-1.3.3 (amended).*

**AC2 — every scanner skips `testdata`.**
**Given** any scanner in this story
**When** it walks a target directory
**Then** it skips any directory named `testdata`, at any depth, by directory name — not by path
prefix and not by a list of specific paths.

**AC3 — the skip and the seam land before the first fixture.**
**Given** the two existing guards (`folio-go/internal/arch_test.go`,
`folio-go/internal/pdf/emit_source_test.go`)
**When** the first fixture file is added anywhere in the repo
**Then** both have already gained the `testdata/` skip and the two-caller shape
**And** this ordering is visible in the task sequence, not merely in the final diff.
*Warrant: **F-3**, measured — a fixture added before the skip is a build break.*

**AC4 — findings carry a defined path base and a rule id.**
**Given** a finding returned by any scanner
**When** it is compared against a fixture scan's expected set
**Then** its path is relative to the **scanned target directory** and it carries a stable **rule id**
**And** the same function pointed at the real tree and at a fixture tree produces paths on the same
base. *Warrant: **F-9**, measured — the two existing guards disagree on this today.*

**AC5 — a scan that could not read its target fails; it never reports zero findings.**
**Given** a target directory containing a file the scanner cannot parse
**When** the scan runs
**Then** it returns a **non-nil error**, distinguishable from "no findings"
**And** the **production caller fails on that error separately from, and before, asserting zero
findings** — the two assertions are two statements, never one
**And** the fixture caller fails on it too
**And** every retained violating fixture is **syntactically valid Go that is semantically
forbidden**, so a fixture can never redden a guard by being unparseable rather than non-compliant.
*Warrant: **F-5**, measured in this run. **This amends D-1.3.3**, which as originally ruled had a
crashed scan and a clean scan return the same value — see `D-1.3.3 (amended)`, and V10 for the
general pattern this is the fourth instance of.*

**AC6 — DW-1 is closed by this seam, as its first application.**
**Given** Story 1.2's AC16 fixture-shape check (`sha256` must be a 64-character lower-case hex JSON
string), which today resolves `fixtures/minimal-rect/expected.json` by a hard-coded relative path
**When** the seam of AC1 is applied to it
**Then** the check takes its fixture location as a parameter, exactly like every other guard here —
**not** as a one-off patch to that call site
**And** it is red-proved against a scratch copy, with the real golden fixture never mutated
**And** `deferred-work.md` marks **DW-1 done in place with this story's commit**, per that file's
stated append-only convention.

### Narrowing the numeric-formatting guard (D-1.3.2)

**AC7 — scope becomes exactly `folio-go/internal/pdf/`, and the module-wide rule is deleted.**
**Given** the numeric-formatting guard
**When** it runs
**Then** its scope is exactly `folio-go/internal/pdf/`: every file except `numbers.go`, `_test.go`
files **included** (the shipped no-exemption choice is kept), any directory named `testdata`
excluded
**And** the second, module-wide rule is **deleted outright** — not retained with an exemption list.
*D-1.3.2: "AD-5 is why deletion is safe: nothing outside `internal/pdf` writes an output byte, so the
wide half protects nothing while costing AD-3's diagnostic carve-out."*

**AC8 — red-proved in both directions, by retained fixtures.**
**Given** the narrowed guard
**When** the fixture scan runs
**Then** `strconv.Itoa` in a non-`numbers.go` file inside a fixture tree standing in for
`internal/pdf` is **reported**
**And** `fmt.Errorf` in a fixture package standing in for `internal/template` is **not reported**
**And** both fixtures are **retained in the tree**, not applied and reverted.
*The second direction is the whole point and unblocks Story 1.4 — see **F-2**, which measures it red
at baseline.*

**AC9 — the Story 1.1 file gets a `Superseded` note; no canonical document is amended for this.**
**Given** `_bmad-output/implementation-artifacts/1-1-a-minimal-pdf-reproducible-on-one-machine.md`,
whose lines **202**, **327** and **678** state the over-broad "nowhere under `folio-go/`" rule
**When** D-1.3.2 is applied
**Then** a short **`Superseded`** section is **appended** to that file pointing at D-1.3.2, naming
those three lines, and **all three lines are left intact**
**And** `epics.md` is **not** touched — D-1.3.2 verified twice that its Story 1.1 AC block is already
correctly narrowed, so **D-000.6 does not fire** for this narrowing
**And** the file's AC6 text is **not** rewritten. *D-1.3.2: "Completed story files are append-only,
exactly like the decision log" — rewriting AC6 would claim we verified something we did not.*

### The AD-1 import lint (D-1.3.1, AD-1, NFR1.c)

**AC10 — the ban, in non-test files.**
**Given** any non-test `.go` file under `folio-go/internal/`
**When** it imports `time`, `os`, `math/rand`, or `net`, or references a `math` transcendental
(`Sin`, `Cos`, `Tan`, `Log`, `Exp`, `Pow`, `Sinh`, `Erf`, …)
**Then** the build fails, naming the **file** and the **offending import or selector**
**And** the failure message names AD-1's allow-listed numeric surface: `+ - * /`, comparison, and
`Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`.

**AC11 — the `_test.go` exemption, keyed on the suffix.**
**Given** a `_test.go` file under `folio-go/internal/`
**When** the lint runs
**Then** `os`, `testing`, `path/filepath` and `embed` are permitted — **and nothing else is added to
that list**
**And** `time`, `math/rand`, `net` and `math` transcendentals remain banned with no exemption
**And** the exemption keys on the **`_test.go` filename suffix**, never a path-prefix skip over a
directory. *D-1.3.1: "otherwise a non-test file in a test-heavy package inherits the exemption
silently."*

**AC12 — `math` is matched by selector on the AST, as a closed allow-list, and by value kind.**
**Given** a file under `folio-go/internal/` importing `math`
**When** the lint runs
**Then** the `math` **package path is not banned** — only its members are judged
**And** **any `math` function call whose name is not one of AD-1's seven** (`Sqrt`, `Floor`, `Ceil`,
`Round`, `Trunc`, `Abs`, `Mod`) **is a violation** — an **allow-list, not a deny-list**
**And** a **non-call reference** is judged by **value kind**: integer-limit constants (`MaxInt64`,
`MinInt64`, `MaxInt32`, …) are permitted; **float-valued constants (`Pi`, `E`, `MaxFloat64`,
`SmallestNonzeroFloat64`) are violations**
**And** matching is on the **AST**, resolving the import alias (`m "math"`) — **never a regex, never
the literal text `math.`**; `emit_source_test.go`'s `importAliases` already does exactly this
**And** the production scan over the real tree is **green** on the shipped suite.

*This is AD-1's own text, not a concession to shipped code (D-1.3.10). AD-1 bans four **package
paths**, then names "any `math` transcendental" — a class of **functions** — and its very next
sentence enumerates seven tolerated `math` **functions**. An import-path ban would make AD-1's own
allow-list unreachable: a rule cannot both ban a package and list which of its functions are
permitted.*

*Why an allow-list rather than a deny-list: AD-1's transcendental list ends in "…" and is therefore
not decidable. Its allow-list is closed. Inverting the test makes the rule decidable **and** fail-safe
when a future Go release adds a transcendental.*

*Why value kind matters: `x := math.Pi` introduces a `float64` with no `float64` identifier and no
`token.FLOAT` literal, so `internal/arch_test.go`'s existing guard would not catch it. This clause
closes that gap.*

**AC12a — the seven allow-listed functions stay in the rule even though nothing can use them.**
**Given** that all seven are `float64`-only, so AD-2 and AD-23 already forbid every call site under
`internal/`
**When** someone notices the redundancy and proposes deleting them
**Then** they are **kept**, and this file's reasoning is cited: deleting them silently converts the
lint from **tolerates-seven** into **bans-all-of-`math`**, which is a **different rule than AD-1's**.
A rule may be un-exercised without being wrong. *D-1.3.10, stated explicitly at the lead's
instruction so the redundancy is not mistaken for dead code.*

**AC13 — nine fixtures for this rule: five violators, and four near-misses.**
*(Corrected post-review, Finding 16, this story's QA review: the heading originally said "six"
against a body that already enumerated eight; the finisher's Finding 6 fix additionally added a
fifth violator — a fixture importing subpackages of three banned paths — bringing the true count
to nine. This is a documentation-count correction only; no fixture behaviour changed.)*
**Given** the fixture tree for this rule
**When** the fixture scan runs
**Then** it reports **exactly**: a **non-test** file importing `time`, and a **`_test.go`** file
importing `time` (D-1.3.1: "One fixture proves only half the rule"), plus a file calling a `math`
function outside the seven, a file referencing `math.Pi` (AC12), and a file importing subpackages of
three banned paths (`math/rand/v2`, `net/http`, `net/url`, `os/exec` — Finding 6, prefix-aware
matching)
**And** it reports **nothing** for a `_test.go` file importing `os`/`testing`/`path/filepath`/`embed`;
nor for a file referencing `math.MaxInt64`; nor for a file calling `math.Abs`
**And** it reports **nothing for a file whose comments name banned symbols** — the fixture must
reproduce exactly the shape standing in the real tree: `internal/geom/scale.go:31` contains the only
`math.Round` under `internal/`, **inside a comment** ("no call to `math.Round`"), alongside five
`math.MinInt64` comment mentions. *This near-miss is what proves the guard does not fire on prose,
and is why AC12 requires the AST rather than a regex.*
**And** all of them stand permanently in the tree.

### Map iteration (D-1.3.5, AD-1, NFR1.d)

**AC14 — total ban, exact and type-based, no reachability analysis.**
**Given** any non-test `.go` file under `folio-go/internal/`
**When** it contains a `range` over a map value
**Then** the build fails naming **file, line and the offending expression**
**And** detection flags an `*ast.RangeStmt` whose subject **resolves to a map type** — a syntactic
guess is **not acceptable**
**And** no reachability or dataflow analysis is attempted. *D-1.3.5: the AC's "can reach an output
byte" states the **hazard**, not the **mechanism**; implementing a hazard statement literally is what
made AD-3 unimplementable.*

**AC15 — the failure message names the escape hatch verbatim.**
**Given** the lint fires on a map range
**When** the developer reads the failure
**Then** the message contains, verbatim:
`for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }`
**And** this idiom needs **no exemption, no allowlist, and no `//lint:ignore` comment** — it ranges a
**slice**, so the rule permits it by construction. *D-1.3.5: "the first developer to hit the rule is
unblocked by the error text itself rather than by finding a story file."* Available at the
`go 1.25.0` language floor.

**AC16 — both polarities, retained.**
**Given** the fixture tree for this rule
**When** the fixture scan runs
**Then** it reports a non-test fixture file ranging a map
**And** it reports **nothing** for the `slices.Sorted(maps.Keys(m))` idiom, for a range over a slice,
over a string, or over an integer
**And** it reports nothing for a `_test.go` fixture ranging a map — the ban is on **non-test** files
(D-1.3.5), and `internal/arch_test.go`'s own `map[string]bool` stays legal.

**AC17 — the production scan asserts zero, non-vacuously.**
**Given** the production scan over `folio-go/internal/`
**When** it runs
**Then** it reports **zero** findings
**And** it cannot pass by having parsed nothing: it asserts it visited the `geom` and `pdf` package
directories **by name** and parsed a non-zero number of files, in the shape
`internal/arch_test.go`'s existing vacuity guard already uses. *Warrant: **F-7**, measured — zero map
ranges under `internal/` today, so the restriction is being taken before a caller exists.*

### The licence check and manifest (D-1.3.4, AD-26)

**AC18 — the Go module graph, at any depth, fails rather than warns.**
**Given** the resolved Go module graph of every Go module in this repository — `folio-go`,
`hashmatrix`, and `lint` itself
**When** any module in it, at any depth, carries **GPL, LGPL, AGPL, SSPL, or a commercial EULA**
**Then** the **build fails** — it does not warn
**And** the ban is **uniform across all three modules** — there is no build-time-only relaxation.
*Scope **confirmed at D-1.3.9**, not narrowed: AD-26's scope line is **Binds: all**, and its Rule's
"the whole module graph" was written when exactly one module existed — the scope line resolves the
ambiguity, the singular noun does not create one. Reinforced by D-1.3.6's explicit requirement that
`lint/`'s own dependency appear in the manifest.*
**And** the standing consequence is recorded: **if a future story genuinely needs a copyleft
build-time tool, that is an owner escalation, not a lead call** (D-1.3.9).

**AC19 — the manifest is generated, asserted complete, and labels what each dependency serves.**
**Given** the resolved graph
**When** the manifest is generated
**Then** **every module in the graph appears with a resolved licence**
**And** an **unknown or unresolvable licence fails the build**
**And** each dependency row is **labelled with the module it serves** (`folio-go`, `hashmatrix`, or
`lint`) **and whether that module is shipped or build-time-only**
**And** the manifest is a committed, reviewable file.
*The label is not a loophole and does not change the ban (AC18). It exists to pre-empt a future
argument that a copyleft build-time tool "doesn't really count" — the label makes the distinction
visible so the argument is had explicitly, with the owner, rather than silently in a diff (D-1.3.9).*

**AC20 — the check covers its own tool's dependency, and that is correct.**
**Given** the `lint/` module's own dependency (expected `golang.org/x/tools`, BSD-3-Clause)
**When** the manifest is generated
**Then** that dependency appears in it
**And** the story states explicitly, in the manifest header and in Dev Notes, that **a checker
covering its own dependency is correct, not a bootstrap problem.** *D-1.3.6, restated at D-1.3.9.*

**AC21 — three asserted absences, each of which goes red on arrival.**
**Given** the licence check
**When** it runs at this commit
**Then** it asserts that **`folio-designer/`'s lockfile is absent**, that the **OFL 1.1 licence text
for shipped faces is absent**, and that the **Apache-2.0 `pdfjs-dist` NOTICE is absent**
**And** each assertion goes **red** the day the corresponding artifact lands, forcing the matching
half to be wired before the build can pass again
**And** a conditional "check it if present" is **not** implemented. *D-1.3.4: "a conditional check
starts silently passing the moment the directory arrives — the guard reports success precisely when it
stops covering anything."*
**And** the absence assertions use the AC1 seam, so a fixture root containing those files red-proves
them without creating them at their real paths.
*Note per D-000.5: the path is **`folio-designer/`**, correcting the AC's `designer/`.*

**AC22 — publication is out of scope and stays tracked.**
**Given** AD-26's "a third-party licence manifest is a release artifact, not a README paragraph"
**When** this story closes
**Then** *publication* — attaching the manifest to a release — is **not** implemented
**And** `deferred-work.md`'s **DW-2** (JS half, owner Story 5.1), **DW-3** (publication, owner Epic 4
close) and **DW-4** (who cuts `folio-go/v0.1.0`, owner: the project owner) remain **open** with their
stated owners and anti-rot mechanisms, unedited except where AC6 closes DW-1.

### Placement (D-1.3.6)

**AC23 — the new guards live in `lint/`.**
**Given** the AD-1 import lint, the map-iteration check, and the licence check + manifest
**When** they are placed
**Then** they live in a standalone tool in its own repo-root module: `lint/`,
`module github.com/panitw/folio/lint`
**And** they are **not** a `go test` inside `folio-go`, and **not** in `folio-go/cmd/` — that path
belongs to Story 3.7 and must not be created early.

**AC24 — `folio-go`'s module graph gains nothing.**
**Given** `folio-go`
**When** this story is complete
**Then** `cd folio-go && go list -m all` prints **exactly one line**, `github.com/panitw/folio/folio-go`
**And** `folio-go/go.mod` has no `require` block and no `go.sum`. *Warrant: **F-8**, measured at
baseline. This is invariant (b) of D-1.3.6 and is the reason the lint is a separate module at all.*

**AC25 — locally runnable; CI merely invokes it.**
**Given** the lint tool
**When** a developer runs it on their own machine with no CI-only setup
**Then** it runs and reports the same findings CI would
**And** the CI workflow's role is to invoke it, not to re-implement any part of it. *Invariant (c) of
D-1.3.6, and D-1.2.6's Finding-8 lesson from Story 1.2: CI must not become a second source of truth
for a relation the Go code already owns.*

**AC26 — the D-000.6 spine amendment for `lint/` ships in this story's commit.**
**Given** `ARCHITECTURE-SPINE.md` §Source tree
**When** `lint/` is created
**Then** the tree block gains one entry for `lint/`, in the **same shape as `hashmatrix/`'s**, naming
the module path and why the module is separate
**And** it ships in **this story's own commit**
**And** **no invariant's Binds or Prevents line is touched**, and the before/after is quoted verbatim
in the decision log (D-000.6, consequences 1–4).

**AC27 — the two existing guards stay put; consolidation is a `DECISION NEEDED`.**
**Given** `folio-go/internal/arch_test.go` and `folio-go/internal/pdf/emit_source_test.go`
**When** this story completes
**Then** both remain where they are — package-scoped, exact, working; moving them is churn
**And** both have gained the `testdata/` skip and the two-caller shape (AC1–AC3)
**And** if consolidation looks genuinely cheap, it is raised as a **`DECISION NEEDED`** and parked —
**never** a judgement call in the diff (D-1.3.6, D-1.2.6).

### Evidence (D-1.3.7, D-1.2.5)

**AC28 — every firing claim cites a retained fixture.**
**Given** the Delivery Log
**When** it claims a guard fires on some input
**Then** it names the **retained fixture** in the tree that demonstrates it, by path
**And** it never cites a mutation that was applied and reverted as the evidence for a firing claim
**And** where it cites an `F-N` from this file as justification, that `F-N` claims to have been
measured (D-1.2.5).

### The licence check's retained fixture graphs (D-1.3.8)

*AC29–AC31 belong to the licence group above. They are numbered after AC28 so that every AC id
already cited in the vacuity table, the red-proofs and the tasks stays stable across this revision.*

**AC29 — three retained fixture module graphs, not two.**
**Given** the licence check's fixture trees at **`lint/testdata/licence/`**
**When** the fixture scan runs over each
**Then** `copyleft/` — a stub module whose `LICENSE` declares GPL-3.0 (and siblings for LGPL, AGPL,
SSPL) — **FAILS**, with the failure naming the **module** and the **licence**
**And** `permissive/` — stubs declaring MIT, Apache-2.0 and BSD-3-Clause — **PASSES**
**And** `unknown/` — a stub with **no `LICENSE` file at all** — **FAILS as unresolvable**
**And** all three are retained permanently.
*The third is the one that matters most (D-1.3.8): **a silent pass on an unidentifiable licence is
the realistic failure mode**, far likelier than someone accidentally adding a GPL dependency. It is
also the licence-shaped instance of V10.*

**AC30 — the fixture graphs are hermetic, obviously fake, and carry no copyleft text.**
**Given** the three fixture graphs
**When** they are constructed
**Then** every `require` has a matching **local `replace`**, and each graph **resolves under
`GOPROXY=off`** — proving the checker resolves without fetching, which is both CI hermeticity and
what makes a fake graph safe to keep
**And** module paths are under **`example.test/`**, and each stub carries a header comment stating it
contains **no third-party code**
**And** **no full copyleft licence text appears anywhere in the repository** — only what the
classifier needs: an **SPDX identifier line or short marker**
**And** `lint/testdata/` is marked **`linguist-vendored`** in `.gitattributes`.
*The no-full-text rule is not legal caution — it is so that no human reader and no downstream scanner
mistakes this repository for containing GPL material. **If the classifier's design turns out to
require full licence text, treat that as a signal to key on SPDX identifiers plus a curated table
instead** (D-1.3.8).*

**AC31 — classifier unit tests supplement the graphs; a recorded graph fixture is not adopted.**
**Given** the classifier
**When** it is tested
**Then** it additionally has direct unit tests over licence-identification inputs — DN-1's option
**(c)**, as a **supplement**
**And** DN-1's option **(a)** — a recorded `go list -m -json all` output standing in as *the graph-walk
fixture* — is **not adopted**
**And** a recorded graph may still be used as **input to a classifier unit test**, never as the
graph-walk fixture.
*D-1.3.8: a recorded graph is a second representation of what production derives live from
`go list -m all`; when the two drift, the guard passes against a reality that no longer exists and
nothing announces it — worse here than usual, because **this guard's whole job is to be trusted
without being watched.** And (c) alone would be the "assertions about guards rather than guards"
outcome D-1.3.7 forbade.*

---

## Vacuity guards — no AC may be satisfied by asserting on these

| # | The guard | Why it would be vacuous |
|---|---|---|
| V1 | **"the fixture scan found at least one finding"** | D-1.3.3 requires **exactly the named findings, by file and rule**. "At least one" passes for an over-broad guard, which is the defect D-1.3.2 exists to clean up. |
| V2 | **"the production scan returned no findings"**, on its own | F-5 measured that a scan which never read its target also returns no findings. AC5 and AC17 require the scan to prove it read something. |
| V3 | **"the compliant fixture compiles"** | Fixtures under `testdata/` are never compiled (F-4). Compilation proves nothing about them; only the scanner's verdict does. |
| V4 | **"`go test ./...` in `folio-go` is green"**, as evidence the new guards work | The new guards live in `lint/` (AC23) and are not in that command's scope at all. |
| V5 | **"the manifest file exists"** | AC19 requires **every module resolved**, with unknown licences failing. An empty or partial manifest that exists is exactly the failure mode. |
| V6 | **"the licence check passed"** at this commit | With three internal modules and one external dependency (F-8), passing is nearly free. The check's teeth are AC18's failure polarity and AC21's absences, demonstrated on fixture roots. |
| V7 | **"the absence assertion passed"** | It passes trivially today because nothing exists. Only a fixture root **containing** the file proves the assertion has a red side (AC21). |
| V8 | **"the map-iteration lint reported nothing"** | F-7 measured zero map ranges under `internal/`, so silence is the expected state whether or not the lint works. AC16's violating fixture is the only proof it fires. |
| V9 | **"`fmt.Errorf` in `internal/template` now passes"** because that package does not exist | AC8's compliant fixture must stand in the tree and be **scanned**. A package that is absent is not a proof. |
| V11 | **"the licence check reported no forbidden licence"** on a graph it could not identify | AC29's `unknown/` graph exists for exactly this: an unresolvable licence must **fail**, not pass quietly. |

*V10 is not a row. It is the general shape all of the above are instances of, and is stated below as
a standing rule to be run against every assertion in this story.*

### V10 — the standing suspicion: healthy output and dead output must never be the same value

**Whenever a guard's healthy output and its dead output are the same value, that is a defect
regardless of how the code looks.** Treat this as a named check to run against every assertion in
this story, not as a description of one bug.

It is stated as a standing rule because F-5 is the **fourth** instance in two stories:

| # | Instance | The shared shape |
|---|---|---|
| 1 | Story 1.2 Finding 1 — the "counted four targets" assertion was structurally unfailable; a narrowed matrix passed green | fewer targets ⇒ same value |
| 2 | Story 1.2 Finding 2 — CI could report the render matrix green having compared three of four legs | a missing leg ⇒ same value |
| 3 | **D-1.3.4**'s rejected conditional lockfile check — it "starts silently passing the moment the directory arrives" | nothing to check ⇒ same value |
| 4 | **F-5**, measured in this run — a crashed scan and a clean scan both return zero findings | target never read ⇒ same value |

Instances 1–3 were each caught by a different reviewer, one at a time, as if they were unrelated.
They are one defect shape. Every assertion added by this story is checked against it, and **AC5,
AC17, AC21, AC29 and V2 are the five places it is already known to bite.**

---

## Red-proofs — every new assertion names what makes it red

Per AC28, the **firing** evidence is the retained fixture, not a mutation. These red-proofs cover the
assertions that are *about the harness itself* and therefore have no fixture.

| # | Assertion | Mutation that must redden it |
|---|---|---|
| RP-1 | AC2 `testdata` skip | Remove the skip from either scanner → the fixture trees are reported by the production scan (this is F-3/F-4's measured behaviour, reproduced deliberately). |
| RP-2 | AC4 path base | Change one scanner's base from the target dir to the repo root → the fixture scan's expected set stops matching. |
| RP-3 | AC5 parse-error handling | Make the scanner swallow the parse error and return `(nil, nil)` → a fixture tree containing an unreadable file must **fail**, not report zero. |
| RP-3b | AC5 caller ordering | Collapse the production caller's two statements into one (`if err != nil \|\| len(findings) > 0`) → the error must still be reported **as an error**, distinctly, not as a finding count. |
| RP-3c | AC1 by-file-not-by-count | Change one fixture caller to assert `len(findings) == n` → swapping two fixtures' expected rules must still redden it. |
| RP-4 | AC6 DW-1 seam | Point the fixture-shape check at a scratch copy with `sha256` widened into a per-target object → red, with the real `fixtures/minimal-rect/expected.json` untouched (`git diff --stat` proves it). |
| RP-5 | AC12 selector-vs-import | Switch the `math` rule to an import-path ban → the production scan reddens on `numbers_test.go` and `scale_test.go` (F-6's measured sites). |
| RP-5b | AC12 allow-list, not deny-list | Invert the `math` function test back to a deny-list of named transcendentals → a call to a `math` function outside AD-1's seven and outside the deny-list is no longer reported. |
| RP-5c | AC12 value kind | Remove the float-valued-constant clause → the `math.Pi` fixture stops being reported, and `internal/arch_test.go` does not catch it either (that is the gap this clause closes). |
| RP-5d | AC12 AST-not-regex | Replace the AST match with a text match on `math.` → the comment near-miss fixture, and `internal/geom/scale.go:31`, are falsely reported. |
| RP-6 | AC14 exactness | Replace type-based resolution with a syntactic guess (e.g. any `range` over an identifier whose name ends in `s`) → the compliant slice fixture is falsely reported. |
| RP-7 | AC15 message text | Delete the idiom from the failure message → the message-content assertion reddens. |
| RP-8 | AC17 non-vacuity | Point the production scan at an empty directory → the "visited `geom` and `pdf` by name" guard reddens. |
| RP-9 | AC19 manifest completeness | Delete one module's row from the manifest, or set its licence to unknown → the completeness assertion reddens. |
| RP-10 | AC21 absences | Point the absence checker at a fixture root containing a lockfile / an OFL text / a `pdfjs-dist` NOTICE → each goes red, one at a time. |
| RP-11 | AC24 module purity | Add any `require` to `folio-go/go.mod` → the one-line `go list -m all` assertion reddens. |
| RP-12 | AC30 hermeticity | Remove one fixture graph's local `replace` → resolution under `GOPROXY=off` must **fail**, not silently reach the network. |
| RP-13 | AC29 unresolvable ⇒ fail | Make the classifier treat a missing `LICENSE` as permissive → the `unknown/` graph stops failing. **This is the V10 instance the lead most wants held.** |
| RP-14 | AC18 uniform ban | Exempt build-time-only modules from the ban → the `copyleft/` graph passes when reached through `lint/`. |
| RP-15 | AC19 labels | Drop the shipped/build-time-only label from a manifest row → the manifest-shape assertion reddens. |

Every mutation is **observed red, then reverted**, and the Delivery Log confirms restoration
byte-identically (`diff` or `git diff --stat`), following Story 1.2's finisher precedent.

---

## The fixture inventory — what stands permanently in the tree

The **source-scanner** fixture trees live at **`folio-go/testdata/lint/<rule>/`** (D-1.3.3). The
**licence** fixture graphs live at **`lint/testdata/licence/`** (D-1.3.8) — they are module graphs,
not source to be scanned, and they are out of every scanner's reach by **two independent
mechanisms**: the `go` command ignores `testdata` for package matching, **and** a directory carrying
its own `go.mod` is outside its parent module's boundary regardless. **No exemption entry anywhere.**

**No fixture is placed under `folio-go/internal/`** (F-10). Every violating source fixture is
**syntactically valid Go** (F-5, AC5). Rule ids are stable strings and appear in findings (AC4).

| Rule id | Violating fixtures (must be reported) | Compliant near-misses (must NOT be reported) |
|---|---|---|
| `forbidden-imports` | a non-test file importing `time`; a `_test.go` file importing `time` (D-1.3.1: two, not one); a call to a `math` function outside AD-1's seven; a reference to `math.Pi` (D-1.3.10) | a `_test.go` importing `os`, `testing`, `path/filepath`, `embed`; a file using `math.MaxInt64`; a file calling `math.Abs`; **a file whose comments name `math.Round` and `math.MinInt64`**, reproducing `internal/geom/scale.go:31`'s shape (F-6) |
| `map-range` | a non-test file ranging a map | the `slices.Sorted(maps.Keys(m))` idiom; a range over a slice, a string, an integer; a `_test.go` ranging a map |
| `numeric-formatting` | `strconv.Itoa` in a non-`numbers.go` file of a tree standing in for `internal/pdf` | `fmt.Errorf` in a package standing in for `internal/template` (F-2 — the Story 1.4 unblock) |
| `no-float64` | a file declaring `float64`, and a bare `float64(x)` conversion | a file using `int64` and `geom.Length` only |
| `licence` (at **`lint/testdata/licence/`**, D-1.3.8) | `copyleft/` — a stub declaring GPL-3.0, plus LGPL/AGPL/SSPL siblings → **FAIL**, naming module and licence. `unknown/` — a stub with **no `LICENSE` file** → **FAIL as unresolvable** | `permissive/` — stubs declaring MIT, Apache-2.0, BSD-3-Clause → **PASS** |
| `absences` | a fixture root containing a lockfile, an OFL 1.1 text, and a `pdfjs-dist` NOTICE | a fixture root containing none of the three |
| `fixture-shape` | a scratch `expected.json` whose `sha256` is a per-target object, and one whose value is upper-case or 63 characters | the real fixture's shape, copied — never the real file mutated (DW-1) |

---

## Tasks / Subtasks

**The ordering is load-bearing, not stylistic.** F-3 measured that a fixture added before the
`testdata/` skip is a build break. Tasks 1–3 must complete before Task 4 adds the first fixture file.

- [x] **Task 1 — the seam, in the two existing guards first (AC1, AC2, AC4, AC5; F-3, F-5, F-9).**
  - [x] Extract `internal/arch_test.go`'s detection into a pure function over a target dir returning
        `(findings, error)`; add the `testdata` directory-name skip; fix the finding path base to the
        target dir and attach the `no-float64` rule id.
  - [x] Same for `internal/pdf/emit_source_test.go`, keeping its current behaviour otherwise —
        narrowing is Task 5, not this task.
  - [x] Return **`(findings, error)`**; make the parse error a non-nil error distinct from an empty
        finding list; have the **production caller fail on it separately from and before** the
        zero-findings assertion (RP-3, RP-3b). *D-1.3.3 (amended), on F-5.*
  - [x] Assert fixture findings **by file and rule, never by count** (RP-3c).
  - [x] Confirm `go test ./...` in `folio-go` is still green with no fixture present.
- [x] **Task 2 — the second caller, still with no fixtures (AC1, AC17).**
  - [x] Add the fixture-scan caller to each existing guard, pointed at an **empty** `testdata/lint/`
        tree, asserting an empty expected set. This proves the two-caller shape compiles and runs
        before any fixture can break it.
  - [x] Keep both production scans' vacuity guards (`geom` and `pdf` visited by name; non-zero decl
        count), and add the equivalent to any new production scan.
- [x] **Task 3 — DW-1: apply the seam to Story 1.2's AC16 check (AC6).**
  - [x] Parameterise the fixture location in `folio-go/matrix_test.go`'s `loadExpectedFixture`
        (currently `filepath.Join(root, "fixtures", "minimal-rect", "expected.json")`, line 413).
        **Establish the seam generally; do not patch AC16 one-off** (D-1.3.3).
  - [x] Red-prove against a scratch copy (RP-4). **Never mutate `fixtures/minimal-rect/expected.json`.**
  - [x] Mark **DW-1 done in place in `deferred-work.md`** with this story's commit, per that file's
        stated convention.
- [x] **Task 4 — the first fixtures land (AC13, AC16, and the `no-float64` rows).**
  - [x] Create `folio-go/testdata/lint/` and the per-rule trees from the fixture inventory above.
  - [x] Verify immediately after the first file lands that `go build ./...`, `go vet ./...` and
        `go test ./...` in `folio-go` are all still green (F-4 measured that they are, *once the skip
        is in*).
- [x] **Task 5 — narrow the numeric-formatting guard (AC7, AC8, AC9; D-1.3.2).**
  - [x] Scope to exactly `folio-go/internal/pdf/`; keep the no-`numbers.go`, `_test.go`-included
        rule; **delete the module-wide half outright** — no exemption list.
  - [x] Both retained fixtures: the `strconv.Itoa` violator and the `fmt.Errorf` near-miss (F-2).
  - [x] Append the **`Superseded`** section to the Story 1.1 story file. Leave lines 202, 327, 678
        intact. Do **not** touch `epics.md`; **D-000.6 does not fire here.**
- [x] **Task 6 — create the `lint/` module (AC23, AC24, AC25).**
  - [x] `lint/go.mod`, `module github.com/panitw/folio/lint`, `go 1.25.0` + `toolchain go1.26.0`
        matching the other two modules (D-1.1.a — the `go` directive sits **below** the toolchain
        pin, or `go mod tidy` deletes the pin).
  - [x] Confirm `cd folio-go && go list -m all` still prints exactly one line (RP-11).
  - [x] If `golang.org/x/tools/go/packages` is used, it is `lint/`'s dependency and nobody else's.
        An exact, dependency-free `go/types` path with a tolerant error config is a **permitted
        refinement, not a deviation** (D-1.3.6).
- [x] **Task 7 — the AD-1 import lint (AC10–AC13).**
  - [x] Selector-based `math` matching **on the AST, resolving the import alias — never a regex**
        (F-6, D-1.3.10). Reuse `emit_source_test.go`'s `importAliases` shape.
  - [x] `math` functions as a **closed allow-list** (AD-1's seven); non-call references by **value
        kind** — integer limits permitted, `Pi`/`E`/`MaxFloat64`/`SmallestNonzeroFloat64` not.
  - [x] **Keep the seven in the rule** and comment why (AC12a) — deleting them converts
        tolerates-seven into bans-all-of-`math`, a different rule than AD-1's.
  - [x] `_test.go` **suffix** keying (D-1.3.1). Failure message naming the file, the offending import
        or selector, and AD-1's allow-listed numeric surface.
  - [x] Production scan over `folio-go/internal/` green; fixture scan matching AC13's nine fixtures
        exactly — including the **comment near-miss** that proves the guard does not fire on prose.
- [x] **Task 8 — the map-iteration check (AC14–AC17).**
  - [x] Exact type-based detection: an `*ast.RangeStmt` whose subject **resolves** to a map type.
  - [x] Failure message contains the escape-hatch idiom **verbatim** (AC15).
  - [x] Non-test files only; `_test.go` files are outside the ban (D-1.3.5).
- [x] **Task 9 — the licence check and manifest (AC18–AC22), and the spine amendment (AC26).**
  - [x] Resolve every repo Go module's graph; classify at any depth; **fail, never warn**; ban
        uniform across all three modules (AC18, D-1.3.9).
  - [x] Generate the manifest; assert completeness; unknown or unresolvable licence fails; label each
        row with the module it serves and **shipped vs build-time-only** (AC19).
  - [x] Build the three fixture graphs at **`lint/testdata/licence/`** — `copyleft/`, `permissive/`,
        `unknown/` (AC29, D-1.3.8). Local `replace` for every `require`; verify each resolves under
        **`GOPROXY=off`** (AC30, RP-12).
  - [x] Fixture hygiene (AC30): module paths under `example.test/`; a header comment in each stub
        stating it contains **no third-party code**; **SPDX identifier lines or short markers only —
        no full copyleft licence text anywhere in the repo**; `lint/testdata/` marked
        **`linguist-vendored`** in `.gitattributes`.
  - [x] Classifier unit tests as a **supplement** (AC31). **Do not** adopt a recorded
        `go list -m -json all` as the graph-walk fixture.
  - [x] The three asserted absences, each red-proved on a fixture root (RP-10). No conditional
        "check if present".
  - [x] Manifest header states explicitly that **the check covering its own tool's dependency is
        correct** (AC20, D-1.3.6).
  - [x] Ship the **D-000.6 spine amendment** adding `lint/` to §Source tree, quoting before/after in
        the decision log. Touch no invariant's Binds or Prevents line.
- [x] **Task 10 — CI invocation (AC25).**
  - [x] Wire `.github/workflows/` to **invoke** the lint. CI re-implements nothing (D-1.2.6,
        Story 1.2 Finding 8).
- [x] **Task 11 — story file, decision log, sprint status → `review`.**
  - [x] Delivery Log: every firing claim cites its retained fixture by path (AC28); every red-proof
        observed and reverted, restoration confirmed; the unrun suites named explicitly per D-000.4
        ("cross-target matrix not run — per-epic cadence, due at Epic 1 close").
  - [x] **Do not commit. Do not set `done`.**

---

## Dev Notes

### The two paragraphs D-1.3.7 fixes verbatim

> **Scope, stated positively.** These guards bind `folio-go/internal/`. They are written as a rule
> about that directory — never as a rule about the repository with exceptions carved out.
> **`hashmatrix/` is not an exemption and must not be named as one.** It is outside `folio-go`
> entirely, so it is out of scope by construction; a guard that mentions it by name has the wrong
> shape, and the next person to add a second name to that list will not notice they are eroding the
> rule.

> **A build tag does not hide a file from these guards.** `internal/arch_test.go` **parses** source
> with `go/parser` rather than building it, so `//go:build` lines are invisible to it — a
> `matrix`-tagged or `ignore`-tagged file under `internal/` is scanned exactly like any other.
> `folio-go/matrix_test.go` is safe because of **where it is** (the module root, outside
> `internal/`), not because of its tag; the tag's only job is AC12, keeping Docker and Node off the
> routine `go test ./...` path. Two things follow: never move a tagged file under `internal/`
> expecting the tag to protect it, and never add a `float64` to `matrix_test.go` on the assumption
> that the tag hides it.

### Why `testdata/` is a category, not an exemption list

D-1.3.3: *"It is Go's own tool-level convention — the `go` command ignores `testdata/` universally —
applied **by category**, so it cannot rot the way a per-file allowlist does. And files under
`testdata/` are never compiled, which is precisely the property a *retained violating fixture*
needs."* F-4 measured both halves at the exact fixture location: `go build ./...` and `go vet ./...`
were clean with a `strconv.Itoa` fixture standing in `folio-go/testdata/lint/probe/`, while the
hand-rolled walk reported it. The go command's convention is free; a `filepath.WalkDir` does not
inherit it.

### Why this reaches the opposite placement from D-1.2.3, and is the same principle

D-1.3.3: *"Both rulings say the same thing — place the artifact so the property holds by construction
— and the properties are opposites, so the mechanisms are opposites. The **probe** must compile and
execute on four targets, so it must be a buildable package… The **lint fixture** must never compile
and only ever be parsed, so `testdata/` is not merely acceptable but better than a module."*

### The rule's scope lives in the caller, not in the checker

The AD-1 lint's rule is *about `folio-go/internal/`*, but its detection function takes a target
directory. That is not a weakening: the production caller supplies `folio-go/internal/` and asserts
zero; the fixture caller supplies `folio-go/testdata/lint/forbidden-imports/` and asserts the exact
set. `_test.go` keying works on filenames within whichever tree is scanned (AC11). Keeping the scope
in the caller is what lets the same code prove both polarities without a fixture ever living under
`internal/`.

### Map iteration: what we are giving up, knowingly

D-1.3.5: *"Map iteration order in Go is deliberately randomised, so a map range that influences output
makes the output non-reproducible… Tracing which map ranges can reach the writer is the hard,
unreliable problem. Banning map ranges in the directory where output is produced is the easy, exact
one. We lose the ability to range a map in render-path code even when it would have been harmless;
that is the price, paid knowingly."* And on shape: *"The determinism boundary is a **directory**
boundary, not a discipline."* F-7 measured the cost at zero today.

### The licence check covering itself is correct

`lint/`'s own dependency appears in the manifest `lint/` generates. **A checker covering its own
dependency is correct, not a bootstrap problem**, and must not be "fixed" by excluding the tool's own
module: AD-26 binds **all**, and a licence manifest that omits the one module doing the checking is
exactly the kind of self-exemption AD-1's shape exists to prevent. Stated in the manifest header per
D-1.3.6 and D-1.3.9.

The manifest's shipped / build-time-only label (AC19) does **not** soften the ban. It exists so that
a future argument of the form "this copyleft tool is only used at build time" has to be made
**explicitly, to the owner**, rather than settled quietly in a diff. Per D-1.3.9, that case is an
**owner escalation, not a lead call**.

### Why a fake GPL module is not a GPL dependency

The full reasoning is under **DN-1 (RULED)** below and must not be re-derived: copyleft obligations
attach to **code received under a licence**, not to a text file that says "GPL". The fixture stubs
contain no third-party code, so there is no licensee relationship to create. Two practical
consequences the developer must honour: **no full copyleft licence text enters the repository** —
SPDX identifiers or short markers only, so no human reader and no downstream scanner mistakes this
repo for containing GPL material — and the stubs are made **obviously fake** (`example.test/` module
paths, a "contains no third-party code" header in each, `lint/testdata/` marked `linguist-vendored`).
If the classifier's design turns out to need full text, that is a **signal to key on SPDX identifiers
plus a curated table**, not a licence to add the text (D-1.3.8).

### Why the licence fixtures sit outside every scanner by construction

`lint/testdata/licence/` is out of reach twice over: the `go` command ignores `testdata` for package
matching, **and** a directory carrying its own `go.mod` is outside its parent module's boundary
regardless. Two independent mechanisms, **no exemption entry anywhere** — which is the same shape
D-1.3.7 demands of the source guards and D-1.2.3 (amended) demands of `hashmatrix/`.

### Things this story must not do

- **Do not create `folio-go/cmd/`.** That path belongs to Story 3.7 (D-1.3.6).
- **Do not add a dependency to `folio-go`.** Invariant (b) of D-1.3.6; AC24 measures it.
- **Do not name `hashmatrix/` in any guard**, as an exemption or otherwise (D-1.3.7).
- **Do not rewrite the Story 1.1 file's AC6.** Append a `Superseded` note; leave lines 202, 327, 678
  intact (D-1.3.2, AC9).
- **Do not touch `epics.md`.** D-1.3.2 verified twice that it is clean.
- **Do not implement a conditional "check the lockfile if present."** Explicitly rejected (D-1.3.4).
- **Do not attempt reachability analysis for map ranges.** Explicitly rejected (D-1.3.5).
- **Do not move the two existing guards.** Consolidation is a `DECISION NEEDED` (D-1.3.6, AC27).
- **Do not mutate `fixtures/minimal-rect/expected.json`** for any red-proof (DW-1, AC6, RP-4).
- **Do not run the cross-target matrix.** No D-000.4 override applies; it is due at Epic 1 close.
- **Do not cite a mutation as evidence that a guard fires.** Cite the retained fixture (AC28).
- **Do not delete the seven allow-listed `math` functions** as dead code (AC12a, D-1.3.10).
- **Do not put full copyleft licence text in the repository** (AC30, D-1.3.8).
- **Do not adopt a recorded module graph as the licence graph-walk fixture** (AC31, D-1.3.8).
- **Do not exempt build-time-only modules from the licence ban.** That is an owner escalation
  (AC18, D-1.3.9).

### Testing standards

Table-driven Go tests (spine §Consistency Conventions). Every rule has a named test. The two callers
per guard are two named tests, not one test with a boolean flag — a single test that scans both trees
can pass by scanning neither in a way neither vacuity guard catches. Fixture trees are read by
repo-root-relative path at test runtime, never `go:embed`ed (AD-21, D-000.5).

### References

- `_bmad-output/planning-artifacts/epics.md` §Story 1.3, lines 482–508 — the source ACs.
- `ARCHITECTURE-SPINE.md` — **AD-1** (line 92, the determinism boundary and the import/map rules),
  **AD-26** (line 465, the licence boundary); AD-3 (119), AD-5 (166), AD-6 (176), AD-23 (408);
  §Source tree (588) is amended by AC26.
- PRD **NFR1.c** (lines 359–363), **NFR1.d** (lines 364–366); `SPEC.md` **CAP-13** (line 125).
- `folio-mvp-decision-log.md` — **D-1.3.1** (512), **D-1.3.2** (1356), **D-1.3.3** (1402),
  **D-1.3.4** (1442), **D-1.3.5** (1476), **D-1.3.6** (1521), **D-1.3.7** (1563), plus the four
  rulings obtained on this story file's own findings — **`D-1.3.3 (amended)`** (on F-5),
  **D-1.3.8** (DN-1 resolved), **D-1.3.9** (AC18 confirmed), **D-1.3.10** (on F-6); D-000.4 (308),
  D-000.5 (358), D-000.6 (606), D-1.1.b (645), D-1.2.3 amended (1180), D-1.2.5 (1250), D-1.2.6 (1281).
- `deferred-work.md` — **DW-1** (owned here), DW-2, DW-3, DW-4.
- Shipped code: `folio-go/internal/arch_test.go`, `folio-go/internal/pdf/emit_source_test.go`,
  `folio-go/internal/pdf/testutil_test.go` (`repoRootFromTest`), `folio-go/matrix_test.go`
  (`loadExpectedFixture`, line 406), `hashmatrix/`.
- Story files: `1-1-a-minimal-pdf-reproducible-on-one-machine.md` (amended by AC9),
  `1-2-cross-target-byte-identity-proven-in-ci.md`.

---

## DECISION NEEDED

**None outstanding.** DN-1 was raised by this file before development and has been **ruled**; the
ruling is `D-1.3.8` and is applied throughout (AC29–AC31, the fixture inventory, Task 9, and the Dev
Note below). It is recorded here rather than deleted, because the *shape* of the ruling is the part a
later story needs.

### DN-1 (RULED — D-1.3.8): the collision I reported does not exist

**What I raised.** `epics.md`'s "a deliberately violating fixture proves it fires, and the fixture is
retained", plus D-1.3.3's both-polarities requirement, appeared to collide with AD-26's *"No
dependency may carry GPL, LGPL, AGPL, SSPL, or a commercial EULA, at any depth."* A licence violation
is a property of a resolved module graph, not inert source, so retaining a fixture that demonstrates
one seemed to require a real forbidden module.

**The ruling, and why the collision dissolves.** AD-26 forbids a dependency **carrying** copyleft,
and its **Prevents** line states the mechanism in the language of static linking and relicensing —
because copyleft obligations attach to **code received under a licence**, not to a text file that says
"GPL". A stub module we authored ourselves, containing **zero lines of third-party code**, whose
`LICENSE` file is fixture data we wrote in order to be classified, creates no licensee relationship,
transfers no copyrighted work, and imposes no obligation. **"A module whose `LICENSE` file declares
GPL" is not "a GPL dependency."**

The clauses collide only if AD-26 is read as a rule about licence **strings** rather than licensed
**work**. So option (b) is adopted with **no trade and no weakening of AD-26** — which is a better
outcome than the one I framed, where all three options cost something.

**What was rejected, and why it matters more than it looks.** Option (a) — a recorded
`go list -m -json all` output standing in as the graph-walk fixture — is **not adopted**: it is a
second representation of what production derives live, and when the two drift the guard passes
against a reality that no longer exists with nothing announcing it. That is worse here than it would
be elsewhere, because **this guard's whole job is to be trusted without being watched.** Option (c)
alone would have been the "assertions about guards rather than guards" outcome D-1.3.7 forbade; it
survives as a **supplement** (AC31).

**The generalisable part.** When a fixture appears to require creating the very thing an invariant
forbids, check whether the invariant governs the **artifact** or the **relationship**. AD-26 governs
a relationship. D-1.3.5 and D-1.1.b turned on the same distinction — hazard versus mechanism — and
each time the fix was to read what the rule protects rather than what it literally says.

---

## Delivery Log

### Gates measured in this story (raw exit codes, `rtk proxy` used throughout per this story's
### testing instructions, so no shell wrapper could mask a non-zero exit)

**Superseded post-review (Blocker 3, this story's QA review).** The block immediately below was the
developer's original measurement. It is corrected in place, not silently rewritten, because its one
error is exactly what Blocker 3 named: the 25/44 count itself was right, but the claim that it
included `TestFixtureShapeCheckRedProof` as a new top-level test was false — that test lived behind
the `matrix` build tag and did not run in this suite at all. The **finisher's re-measurement**,
after all 17 findings' fixes (including moving that test out of the tag), follows immediately after.

```
$ cd folio-go && rtk proxy go build ./...            → exit 0
$ cd folio-go && rtk proxy go vet ./...               → exit 0
$ cd folio-go && rtk proxy gofmt -l .                 → exit 0, no output
$ cd folio-go && rtk proxy go test ./... -v -count=1  → exit 0
    ok  github.com/panitw/folio/folio-go
    ok  github.com/panitw/folio/folio-go/internal
    ok  github.com/panitw/folio/folio-go/internal/geom
    ok  github.com/panitw/folio/folio-go/internal/pdf
    25 top-level tests / 44 with subtests / 4 packages — the COUNT is correct, re-measured;
    the ATTRIBUTION is not (Blocker 3): TestFixtureShapeCheckRedProof was not actually among
    the top-level tests in this run — it lived behind //go:build matrix and never executed
    here. Only TestNoFloat64FixtureScan and TestNumberFormattingFixtureScan were genuinely new
    top-level tests in this count.
$ cd folio-go && rtk proxy go build -tags=matrix ./... → exit 0
$ cd folio-go && rtk proxy go vet -tags=matrix ./...   → exit 0

$ cd hashmatrix && rtk proxy go vet ./...              → exit 0
$ cd hashmatrix && rtk proxy gofmt -l .                → exit 0, no output
$ cd hashmatrix && rtk proxy go build -o /tmp/hm-probe ./probe → exit 0
  (plain `go build ./...` is NOT this module's gate — D-1.2.3 amended, Story 1.2 — not re-tested here)

$ cd lint && rtk proxy go build ./...                  → exit 0
$ cd lint && rtk proxy go vet ./...                    → exit 0
$ cd lint && rtk proxy gofmt -l .                      → exit 0, no output
$ cd lint && GOPROXY=off GOFLAGS=-mod=mod rtk proxy go test ./... -v -count=1 → exit 0
    ok  github.com/panitw/folio/lint/internal/licence
    ok  github.com/panitw/folio/lint/internal/manifest
    ok  github.com/panitw/folio/lint/internal/rules
    (13 top-level tests, several with t.Run subtests, across 3 packages;
    lint/cmd/genmanifest carries no test file — it is a two-line CLI wrapper around
    manifest.Generate/Render, both of which are exercised directly by
    lint/internal/manifest's tests)

$ cd folio-go && rtk proxy go list -m all → github.com/panitw/folio/folio-go   (AC24, RP-11)
```

### Finisher's re-measurement (raw exit codes, `rtk proxy`, after all 17 findings' fixes)

`lint/` no longer has zero dependencies (D-1.3.11, Blocker 1): `golang.org/x/tools` and its own
transitive graph are now real. Every number below was re-measured in this session, live, not
carried forward from the block above.

```
folio-go: rtk proxy go build ./...              → exit 0
          rtk proxy go vet ./...                → exit 0
          rtk proxy gofmt -l .                  → exit 0, no output
          rtk proxy go test ./... -v -count=1   → exit 0
              26 top-level tests / 51 with subtests / 4 packages (all PASS; the +1/+7 over the
              pre-fix 25/44 above is TestFixtureShapeCheckRedProof now genuinely running,
              Blocker 4's split into two subtests, and Finding 12's second scratch-case subtests)
          rtk proxy go build -tags=matrix ./...  → exit 0
          rtk proxy go vet -tags=matrix ./...    → exit 0
          rtk proxy go list -m all               → github.com/panitw/folio/folio-go  (one line, no go.sum — AC24, D-1.3.11 invariant (b) re-verified)

hashmatrix: rtk proxy go vet ./...               → exit 0
            rtk proxy gofmt -l .                 → exit 0, no output
            rtk proxy go build -o /tmp/hm-probe-final/probe ./probe → exit 0

lint:     rtk proxy go build ./...               → exit 0
          rtk proxy go vet ./...                 → exit 0
          rtk proxy gofmt -l .                   → exit 0, no output
          GOPROXY=off GOFLAGS=-mod=mod rtk proxy go test ./... -v -count=1 → exit 0
              ok  github.com/panitw/folio/lint/cmd/genmanifest
              ok  github.com/panitw/folio/lint/internal/licence
              ok  github.com/panitw/folio/lint/internal/manifest
              ok  github.com/panitw/folio/lint/internal/rules
              14 top-level tests / 35 with subtests / 4 packages — genmanifest now carries a test
              (Finding 17); rules gained TestForbiddenImportsMessageContent (Finding 11)

lint fixture graphs (GOPROXY=off, D-1.3.8, re-proved with a real dependency in lint's own graph):
  copyleft/, permissive/, unknown/  → each `go list -m all` exits 0, unchanged in shape

lint's own module graph under GOPROXY=off (the actual new hermeticity case): confirmed 0 exit
  for `go test ./...` — but ONLY after `go mod download all` first populates the module cache for
  packages `lint` never directly imports (x/net, x/sys, x/telemetry, goldmark — pulled in by
  x/tools' own go.mod, needed for their LICENSE files, not for the build). Measured with a scratch
  GOMODCACHE: `go build`/`go vet`/`gofmt` alone leave it failing GOPROXY=off
  (`module lookup disabled by GOPROXY=off`); `go mod download all` first fixes it. `ci.yml`'s
  `lint` job gained that step (see Blocker 1's Finding Resolution above).

lint tool against the real tree: zero findings (all four production-scan tests above, PASS).
lint tool against every fixture tree: exactly the named findings, by file and rule (all
  fixture-scan tests above, PASS) — including the Blocker-1/Major-6 additions.
```

### Suites deliberately not run (D-000.4)

The cross-target byte-identity matrix (`folio-go/matrix_test.go`'s `TestCrossTargetByteIdentity` /
`TestTargetRenderHash`, requiring Docker + Node) was **not run** in this story. This story's
deliverable is not hash-shaped and changes no output bytes; the matrix is due at the Epic 1
boundary under D-000.4's per-epic cadence, not on every story. `go build -tags=matrix ./...` and
`go vet -tags=matrix ./...` (above) confirm the tagged code still compiles, which is this story's
own per-story obligation for tagged code (Story 1.2 precedent).

### Retained fixtures, by rule, with the findings each produce (AC28 — every claim below names the
### retained fixture; none of this evidence is a mutation applied and reverted)

| Rule | Retained fixture path | Reported? |
|---|---|---|
| `no-float64` | `folio-go/testdata/lint/no-float64/violating_field.go` | Yes — `float64` field |
| `no-float64` | `folio-go/testdata/lint/no-float64/violating_conversion.go` | Yes — `float64` return type + `float64(x)` conversion (2 AST sites, 1 finding by file+rule) |
| `no-float64` | `folio-go/testdata/lint/no-float64/compliant.go` | No — `int64`/local `length` stand-in only |
| `numeric-formatting` | `folio-go/testdata/lint/numeric-formatting/violating.go` | Yes — `strconv.Itoa`, not `numbers.go` |
| `numeric-formatting` | `folio-go/testdata/lint/numeric-formatting/template/numbers.go` | No — named `numbers.go` (this guard's one real exemption) so the exact `fmt.Errorf`/`VersionError` shape F-2 measured stands retained and unreported; the true Story 1.4 unblock is the AC7 scope narrowing (production caller now roots at `internal/pdf/` only), which this fixture demonstrates additionally holds even when handed directly to the checker |
| `forbidden-imports` | `folio-go/testdata/lint/forbidden-imports/violating_time_import.go` | Yes — non-test `time` import |
| `forbidden-imports` | `folio-go/testdata/lint/forbidden-imports/violating_time_import_test.go` | Yes — `_test.go` `time` import (D-1.3.1: two, not one) |
| `forbidden-imports` | `folio-go/testdata/lint/forbidden-imports/violating_math_call.go` | Yes — `math.Sin`, outside the seven |
| `forbidden-imports` | `folio-go/testdata/lint/forbidden-imports/violating_math_pi.go` | Yes — `math.Pi`, float-valued non-call reference |
| `forbidden-imports` | `.../compliant_test_exemption_test.go` | No — `os`/`testing`/`path/filepath`/`embed` only |
| `forbidden-imports` | `.../compliant_math_int_constant.go` | No — `math.MaxInt64` |
| `forbidden-imports` | `.../compliant_math_abs.go` | No — `math.Abs`, one of the seven |
| `forbidden-imports` | `.../compliant_comment_near_miss.go` | No — reproduces `internal/geom/scale.go:31`'s shape: `math.Round`/`math.MinInt64` named only in comments |
| `map-range` | `folio-go/testdata/lint/map-range/violating_map_range.go` | Yes — non-test map range |
| `map-range` | `.../compliant_escape_hatch.go` | No — `slices.Sorted(maps.Keys(m))` idiom |
| `map-range` | `.../compliant_range_slice.go`, `compliant_range_string.go`, `compliant_range_int.go` | No |
| `map-range` | `.../compliant_test_map_range_test.go` | No — `_test.go`, out of scope by D-1.3.5 |
| `licence` | `lint/testdata/licence/copyleft/` (4 sibling stubs) | Yes, all four — GPL-3.0/LGPL-3.0/AGPL-3.0/SSPL-1.0 |
| `licence` | `lint/testdata/licence/permissive/` (3 sibling stubs) | No — MIT/Apache-2.0/BSD-3-Clause |
| `licence` | `lint/testdata/licence/unknown/` (1 stub, no LICENSE file) | Yes — unresolvable |
| `absences` | `folio-go/testdata/lint/absences/violating/` | Yes, all three (lockfile, OFL.txt, pdfjs-dist NOTICE) |
| `absences` | `folio-go/testdata/lint/absences/compliant/` | No |
| `fixture-shape` (DW-1) | scratch copy in `t.TempDir()`, per-target `sha256` widened | Yes (this one is legitimately a scratch copy, not a retained tree — AC6 requires the real golden fixture never mutated, and DW-1's whole point is that no retained widened-JSON fixture should exist alongside the real one) |

All of the above are asserted by `lint/internal/rules`' `TestForbiddenImportsFixtureScan`,
`TestMapRangeFixtureScan`, `TestLicenceGraphFixtureScan`, `TestAbsencesFixtureScan`, and
`folio-go/internal`'s `TestNoFloat64FixtureScan` / `folio-go/internal/pdf`'s
`TestNumberFormattingFixtureScan` — every one passing at the gates above.

### The three licence fixture graphs, and their `GOPROXY=off` resolution (AC29, AC30)

```
$ cd lint/testdata/licence/copyleft   && GOPROXY=off GOFLAGS=-mod=mod go list -m -json all → resolves, 4 replaced modules
$ cd lint/testdata/licence/permissive && GOPROXY=off GOFLAGS=-mod=mod go list -m all       → resolves, 3 replaced modules
$ cd lint/testdata/licence/unknown    && GOPROXY=off GOFLAGS=-mod=mod go list -m all       → resolves, 1 replaced module (no LICENSE)
```
Every `require` has a matching local `replace`; module paths are under `example.test/`; each stub
carries a `NOTICE.md` stating it contains no third-party code; only SPDX marker lines appear in
`LICENSE` files (`SPDX-License-Identifier: <id>`) — no full copyleft licence text anywhere in the
repository (AC30). `lint/testdata/** linguist-vendored` added to `.gitattributes`.

### Red-proofs — RP-1 … RP-15

**Correction (Major 7, this story's QA review): the header below overclaimed.** It originally read
"each observed red and reverted" / "every mutation … was applied, observed red (quoted), and
reverted" — but five rows (RP-3b, RP-3c, RP-6, RP-7, RP-14) were a code-inspection, a hypothetical
comparison, an obsolete probe against code that no longer exists, and two rows that were honest
about being unmeasured (RP-7 said so itself: "N/A — always green"). A claim of measurement that
wasn't measured is the defect under D-1.2.5, independent of whether the underlying code was fine.
Corrected: most mutations below genuinely were applied, observed red, and reverted (RP-1, RP-2,
RP-3, RP-4, RP-5, RP-5b–d, RP-8–13, RP-15 — unaffected by this correction). RP-3b and RP-3c are
source-inspection facts, labelled as such. RP-6 is retired (the code it probed no longer exists,
Blocker 1). RP-7 and RP-14 were re-run for real by the finisher, live, this session — see below and
the Finding Resolutions section above.

`git status --porcelain fixtures/` and `git diff --stat` confirmed byte-identity of every real,
non-scratch file after every mutation in this table (none showed a diff at any point in this
story or the finisher's pass).

| # | Mutation applied | Observed red | Reverted |
|---|---|---|---|
| RP-1 | Removed `testdata` skip from `arch_test.go`'s walker; reproduced F-3's exact probe (`internal/testdata/lintprobe/bad.go`, `float64`) | `TestNoFloat64UnderInternal` FAILs, reporting `testdata/lintprobe/bad.go` | Yes — file restored from backup, probe directory removed, suite green |
| RP-2 | Changed `scanNoFloat64`'s path base from target-dir-relative to absolute | `TestNoFloat64FixtureScan` FAILs: expected findings not reported, unexpected absolute-path findings reported | Yes |
| RP-3 | Made `scanNoFloat64` swallow its walk error (`return findings, nil`) over an unparseable fixture | Direct probe: `err == nil`, `findings == []` — exactly F-5's "crashed scan and clean scan return the same value" | Yes |
| RP-3b | **Not a mutation — relabelled per Major 7.** Read (not mutated) every production caller's source, this finisher pass included (six callers, AC5 caller ordering, already independently re-verified by the reviewer) | Confirmed by reading the shipped source: the error check and the findings check are two separate `if` statements in every production caller, never collapsed into one | N/A — source-inspection fact, not a runtime mutation; correctly labelled as such now |
| RP-3c | **Not a mutation — relabelled per Major 7.** Read (not mutated) `assertExactFindings`'s shipped implementation (both the `lint` and `folio-go/internal` copies) | Confirmed by reading the source: both compare `(path,rule)` **sets**, never raw counts — a rule-swapped expected set with the same length as `got` would correctly redden on set-membership, not silently pass on count alone | N/A — source-inspection fact, not a runtime mutation |
| RP-4 | Widened a scratch copy's `sha256` into a per-target object | `TestFixtureShapeCheckRedProof` (itself a permanent, shipped test) FAILs `checkFixtureShape` on the scratch copy; real `fixtures/minimal-rect/expected.json` unchanged (`bytes.Equal` assertion inside the test itself) | N/A — the red-proof is the shipped test |
| RP-5 | Added `"math": true` to `bannedImportPaths` (import-path ban) | `TestForbiddenImportsProductionScan` FAILs on `geom/scale_test.go:4` and `pdf/numbers_test.go:4` — exactly F-6's predicted sites | Yes |
| RP-5b | Replaced the allow-list with a hand-rolled deny-list of named transcendentals | Probe: `math.Cbrt` (outside both the allow-list and the deny-list) is missed | Yes |
| RP-5c | Removed the value-kind clause for non-call `math` references | `TestForbiddenImportsFixtureScan` FAILs: `violating_math_pi.go` no longer reported | Yes |
| RP-5d | Added a literal text-match on `"math."` alongside the AST match | Probe: falsely reports `compliant_comment_near_miss.go` | Yes |
| RP-6 | **Retired per Major 7/Blocker 1 — the code this probed no longer exists.** Originally: a hypothetical syntactic guess ("range subject identifier ends in 's'") never actually present in shipped code. Post-Blocker-1, `ScanMapRange` resolves types exactly via `go/packages`; there is no syntactic-guess code path left to probe, hypothetical or otherwise | N/A | N/A — row retired, not re-run, because the premise (a syntactic-guess mechanism to probe) no longer applies to the rewritten checker |
| RP-7 | **Re-run for real by the finisher (Blocker 2).** Changed `EscapeHatch`'s value to `"TODO"` | `TestMapRangeFailureMessageNamesEscapeHatch` (rewritten to assert against an independently-spelled literal, not the constant that built the message) FAILs: `EscapeHatch constant drifted from the idiom this test independently spells out` | Yes — `diff` confirmed clean, suite green again |
| RP-8 | Pointed the group-by-directory walk at an empty directory | Probe: neither `geom` nor `pdf` visited — the same vacuity guard `TestMapRangeProductionScan` runs would fail | Yes |
| RP-9 | Overwrote `lint/MANIFEST.md` with unrelated content | `TestManifestUpToDate` FAILs: "out of date" | Yes |
| RP-10 | Created a real `folio-designer/package-lock.json` at the actual repo root | `TestAbsencesProductionScan` FAILs, naming `folio-designer/package-lock.json` | Yes — directory removed, suite green |
| RP-11 | Added `require golang.org/x/text v0.14.0` to `folio-go/go.mod` | `go list -m all` no longer prints the single expected line (errors on a missing go.sum entry instead) | Yes |
| RP-12 | Removed one `replace` line from `lint/testdata/licence/copyleft/go.mod` | `GOPROXY=off go list -m all` FAILs: "module lookup disabled by GOPROXY=off" | Yes |
| RP-13 | Made `ScanLicenceGraph` treat a missing LICENSE as permissive (skip instead of finding) | `TestLicenceGraphFixtureScan/unknown` FAILs: expected finding not reported | Yes |
| RP-14 | **Re-run for real by the finisher (Major 7).** Added `if strings.Contains(moduleDir, "lint") { return nil, nil }` as `ScanLicenceGraph`'s first statement (the fixture graphs live under `lint/testdata/licence/…`, so this reproduces "reached through lint" concretely) | `TestLicenceGraphFixtureScan/copyleft` and `/unknown` both FAIL: `expected finding not reported … rule=licence`; `/permissive` stays green (nothing to report either way) | Yes — `diff` confirmed clean, suite green again |
| RP-15 | Dropped the shipped/build-time-only column from `manifest.Render`'s row format | Discovered this is **vacuous against the real, empty-today manifest** (zero external deps, F-8 — the empty-table branch never reaches the row formatter). Added a permanent, non-vacuous test, `lint/internal/manifest/render_test.go`'s `TestRenderRowShapeIncludesServesAndLabel`, which fabricates one row and asserts the column; it now catches this mutation directly | Yes — this is itself a small V10 instance this story's own review surfaced and closed with a shipped test, not just a reverted probe |

### DW-1 closure (AC6)

`folio-go/matrix_test.go`'s `loadExpectedFixture` was split into a pure `checkFixtureShape(path
string) (expectedFixture, error)` taking the fixture's location as a parameter (the AC1 seam,
applied generally rather than patched one-off), plus the permanent `TestFixtureShapeCheckRedProof`
(RP-4 above).

**Correction (Blocker 3, this story's QA review).** The paragraph above described the state as
shipped for review — but `checkFixtureShape`/`loadExpectedFixture`/`TestFixtureShapeCheckRedProof`
all lived in `matrix_test.go`, which carries `//go:build matrix`, so DW-1's red-proof executed in
**zero** gates at review time despite being marked DONE on its strength. The finisher moved all
three into the untagged `folio-go/fixture_test.go` and added Finding 12's missing second scratch
case (a 63-character and an upper-case `sha256`). `go test ./...` (no build tag) now runs
`TestFixtureShapeCheckRedProof` — measured this session, both subtests PASS, and real-fixture
byte-identity is re-verified after each. `deferred-work.md`'s DW-1 entry is marked **DONE**, moved
to a `## Done` heading, in place, with this story's finisher commit — now that its red-proof
genuinely runs in a gate that runs, which is what Blocker 3 required before DW-1 could be marked
closed at all.

### D-000.6 spine amendment for `lint/` — before/after quoted verbatim (AC26)

Quoted in full in `folio-mvp-decision-log.md` under "D-000.6 amendment (Story 1.3) — `lint/` added
to the spine's §Source tree", immediately following the Story 1.2 `hashmatrix/` amendment it
mirrors. `ARCHITECTURE-SPINE.md` §Source tree gained exactly one entry (`lint/`), in the same shape
as `hashmatrix/`'s; no invariant's **Binds** or **Prevents** line was touched (verified: the only
diff in that file is the one tree-block insertion).

### The `Superseded` note appended to the Story 1.1 file (AC9)

Appended as a new `## Superseded` section at the end of
`1-1-a-minimal-pdf-reproducible-on-one-machine.md`, pointing at D-1.3.2 and naming lines 202, 327
and 678 explicitly. All three lines, and the file's AC6 text, are left byte-for-byte unedited
(verified: `git diff` on that file shows only an appended block, no changes above it).
`epics.md` was not touched (D-1.3.2 verified twice that its Story 1.1 AC block is already
correctly narrowed).

### The generated licence manifest

`lint/MANIFEST.md`, generated by `go run ./lint/cmd/genmanifest` from `lint/internal/manifest`,
committed. At this commit all three module graphs (`folio-go`, `hashmatrix`, `lint`) resolve with
**zero external dependencies** (F-8 confirmed unchanged: `lint` was deliberately built with **no
external dependency** — see "D-1.3.6 refinement taken" below — so the manifest's row table is
currently empty, and the file says so explicitly rather than rendering an empty table silently).
`TestManifestUpToDate` asserts the committed file matches live generation (RP-9); a fabricated-row
test (`TestRenderRowShapeIncludesServesAndLabel`) asserts the row shape non-vacuously (RP-15) since
the empty-table state cannot exercise it.

**Superseded post-review (Blocker 1, D-1.3.11).** The paragraph above is no longer accurate: `lint`
no longer has zero external dependencies. `lint/MANIFEST.md` was regenerated after the finisher's
fix; it now lists 8 rows (`golang.org/x/tools` and its own transitive graph — `x/mod`, `x/sync`,
`x/net`, `x/sys`, `x/telemetry`, `github.com/yuin/goldmark`, `github.com/google/go-cmp`), all
permissive (BSD-3-Clause or MIT), all labelled `lint` / `build-time-only`. `TestManifestUpToDate`
and `TestRenderRowShapeIncludesServesAndLabel` both re-verified against the regenerated file — the
row-shape test is no longer the *only* thing exercising the row-formatting branch, since the table
is genuinely non-empty now, but it stays as a permanent guard regardless of graph size.

### D-1.3.6 refinement taken: `lint/` ships with zero dependencies, not `golang.org/x/tools`

D-1.3.6 states an exact, dependency-free `go/types` path is "a permitted refinement, not a
deviation." This story takes that refinement fully: `lint/go.mod` has **no `require` block at
all**. Map-range detection (AC14) uses `go/types` + `go/importer` directly (a tolerant importer
that resolves stdlib offline and substitutes an empty package for anything else, since map-range
resolution needs best-effort local type information, not a fully sound whole-program check) —
`lint/internal/rules/maprange.go`. This is a genuine consequence, not a corner cut: AC20's premise
("the lint module's own dependency, expected `golang.org/x/tools`") does not materialise, and F-8's
predicted "apart from lint's own dependency, zero external modules" becomes "zero external modules,
full stop, across all three." AC20's principle — "a checker covering its own dependency is correct,
not a bootstrap problem" — is stated in the manifest header regardless, ready for the day any of
the three modules gains a real dependency. This is a judgement call within an explicitly
pre-approved option, not a new decision, so it was not routed as `DECISION NEEDED`.

**Reversed post-review — D-1.3.11 (Blocker 1, this story's QA review).** The refinement described
above does not survive: D-1.3.6 permitted the dependency-free path only *"if the developer finds an
**exact**, dependency-free path"*, and its binding invariant (a) is *"exact detection with no false
positives."* The reviewer proved, on a copy of the real tree, that the shipped
`tolerantImporter`-based path is not exact — it silently missed a map type declared in one
`internal/` package and ranged in another (`internal/pdf` already imports `internal/geom`; adding
`type ScaleTable map[string]Length` to `geom` and two genuine map ranges over it in `pdf` gave
`err=<nil>, findings=0`). **A ruling that permits X only if X is exact is not satisfied by an X that
documents its own inexactness** — the `ScanMapRange` doc comment and `go.mod` both said "best
effort" out loud. `D-1.3.11`'s verdict, applying D-1.3.6 as written: *"`lint/` adds
`golang.org/x/tools` (BSD-3-Clause) and resolves map types through `go/packages` with full type
information. The stdlib-only `tolerantImporter` path is withdrawn."* Invariant (b) — "no dependency
added to `folio-go`'s module graph" — was never at risk and is re-verified above (`go list -m all`
still one line). **This is this story's own most important finding about itself: it named V10 as a
named vacuity guard and then shipped an instance of it — the checker that returns "0 findings" for
both "nothing to find" and "found nothing because I couldn't see it" is exactly the healthy/dead
ambiguity V10 warns against, and it shipped in the one place (AC17's vacuity guards, Major 5) built
specifically to catch it.**

### D-1.2.6 disclosure — ruling conflicts met, and how they were surfaced

No ruling conflicts were encountered. `DECISION NEEDED` in this file states "None outstanding" and
that held throughout. Two judgement calls worth naming explicitly (both within pre-approved
latitude, not new decisions):

1. **D-1.3.6's dependency-free refinement** — see immediately above.
2. **AC8's "fixture package standing in for internal/template" mechanism.** The retained compliant
   fixture demonstrating the numeric-formatting guard's Story 1.4 unblock
   (`folio-go/testdata/lint/numeric-formatting/template/numbers.go`) is named `numbers.go` so this
   guard's one real, filename-based exemption legitimately covers it when the fixture scan runs
   directly over it — proving the exact `fmt.Errorf`/`VersionError` shape F-2 measured is not
   reported, non-vacuously (V9), on a retained fixture rather than a reverted mutation. The
   narrowing that actually unblocks Story 1.4 in production is AC7's scope change (the production
   caller now roots at `internal/pdf/` only, so `internal/template/` is never scanned at all); this
   fixture additionally shows the same call is not reported even when handed directly to the
   checker. This is documented here per D-1.2.5/D-1.2.6's instinct to disclose a non-obvious
   construction rather than let it look self-evident.

## Dev Agent Record

### Agent Model Used
Claude (bmad-story-developer agent), Sonnet 5.

### Debug Log References
All gate commands and their raw exit codes are reproduced in "Gates measured in this story" above,
run via `rtk proxy` throughout per this story's testing instructions. All red-proof mutations and
their observed failures are reproduced in the red-proofs table above.

### Completion Notes List
- Tasks 1–3 (the seam, the second caller, DW-1) landed and were verified green **before** Task 4's
  first fixture, per F-3's measured ordering trap — confirmed directly: adding
  `internal/testdata/lintprobe/bad.go` before the skip existed reddened `TestNoFloat64UnderInternal`
  exactly as F-3 predicted (see RP-1).
- The two existing guards (`arch_test.go`, `emit_source_test.go`) stay in `folio-go/internal` and
  `folio-go/internal/pdf` respectively (AC27) — not moved into `lint/`. The three new guards
  (forbidden-imports, map-range, licence+manifest) live in `lint/` (AC23).
- `lint/` adds no dependency to `folio-go`'s module graph (AC24, RP-11) and needed none of its own
  (D-1.3.6's dependency-free refinement, disclosed above).
- One vacuity gap was found and closed during this story's own red-proofing (RP-15): the licence
  manifest's row-shape assertions were vacuous against today's empty (zero-dependency) manifest.
  Closed with a permanent, non-vacuous unit test (`TestRenderRowShapeIncludesServesAndLabel`) rather
  than left as a reverted probe — this is a fifth V10 instance, caught by this story's own
  red-proofing discipline rather than by a later reviewer.
- `.github/workflows/ci.yml` added (Task 10): three jobs (`folio-go`, `hashmatrix`, `lint`), each
  invoking exactly the commands a developer runs locally — no CI-only re-implementation (AC25,
  D-1.2.6 Story 1.2 Finding 8).
- The cross-target matrix (`matrix.yml`) was not touched and was not run (D-000.4).

**Finisher completion note (post-review).** All 17 QA findings resolved as FIX — see "Finding
Resolutions (finisher, post-review)" at the end of this file for the full triage, and the
corrections made in place to the Delivery Log's gates block, red-proof table, DW-1 closure note,
and manifest/D-1.3.6 sections above. The headline reversal: `D-1.3.11` withdrew the dependency-free
map-range detector (Blocker 1) and `lint/` now depends on `golang.org/x/tools`, re-verified against
invariant (b) — `folio-go`'s own module graph is still exactly one line. Every fix was re-measured
live in this session (`go build`/`go vet`/`gofmt`/`go test` across all three modules, `-tags=matrix`
build+vet, all three licence fixture graphs and `lint`'s own graph under `GOPROXY=off`), and the
red-proofs the review flagged as unmeasured (RP-7, RP-14) were re-run for real, observed red, and
reverted; RP-6 was retired because the code it probed no longer exists post-Blocker-1. The cross-
target matrix remains deferred to the Epic 1 boundary (D-000.4) — this story still changes no
output bytes.

### File List

**Modified:**
- `folio-go/internal/arch_test.go` — AC1 seam (no-float64), testdata skip, fixture caller
- `folio-go/internal/pdf/emit_source_test.go` — AC1 seam + D-1.3.2 narrowing (numeric-formatting), fixture caller
- `folio-go/matrix_test.go` — DW-1: `loadExpectedFixture`/`checkFixtureShape` parameterised, `TestFixtureShapeCheckRedProof` added
- `.gitattributes` — `lint/testdata/** linguist-vendored`
- `_bmad-output/implementation-artifacts/1-1-a-minimal-pdf-reproducible-on-one-machine.md` — `## Superseded` section appended (AC9)
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.6 amendment for `lint/` appended
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-1 marked DONE
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — status only
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — §Source tree gains `lint/` (AC26)

**Added:**
- `.github/workflows/ci.yml`
- `folio-go/testdata/lint/no-float64/{violating_field.go,violating_conversion.go,compliant.go}`
- `folio-go/testdata/lint/numeric-formatting/violating.go`
- `folio-go/testdata/lint/numeric-formatting/template/numbers.go`
- `folio-go/testdata/lint/forbidden-imports/{violating_time_import.go,violating_time_import_test.go,violating_math_call.go,violating_math_pi.go,compliant_test_exemption_test.go,compliant_math_int_constant.go,compliant_math_abs.go,compliant_comment_near_miss.go}`
- `folio-go/testdata/lint/map-range/{violating_map_range.go,compliant_escape_hatch.go,compliant_range_slice.go,compliant_range_string.go,compliant_range_int.go,compliant_test_map_range_test.go}`
- `folio-go/testdata/lint/absences/violating/folio-designer/package-lock.json`
- `folio-go/testdata/lint/absences/violating/folio-designer/third-party-notices/pdfjs-dist/NOTICE`
- `folio-go/testdata/lint/absences/violating/folio-go/fonts/OFL.txt`
- `folio-go/testdata/lint/absences/compliant/README.md`
- `lint/go.mod`
- `lint/MANIFEST.md`
- `lint/cmd/genmanifest/main.go`
- `lint/internal/rules/{finding.go,walk.go,forbiddenimports.go,forbiddenimports_test.go,maprange.go,maprange_test.go,licencegraph.go,licencegraph_test.go,absences.go,absences_test.go,testutil_test.go}`
- `lint/internal/licence/{classify.go,classify_test.go,graph.go}`
- `lint/internal/manifest/{manifest.go,manifest_test.go,render_test.go}`
- `lint/testdata/licence/copyleft/{go.mod,example.test/{gpl-lib,lgpl-lib,agpl-lib,sspl-lib}/{go.mod,LICENSE,NOTICE.md}}`
- `lint/testdata/licence/permissive/{go.mod,example.test/{mit-lib,apache-lib,bsd-lib}/{go.mod,LICENSE,NOTICE.md}}`
- `lint/testdata/licence/unknown/{go.mod,example.test/mystery-lib/{go.mod,NOTICE.md}}`
- `_bmad-output/implementation-artifacts/1-3-guardrails-that-fail-the-build.md` (this file, was already untracked from creation)

### Finisher File List addendum (post-review fixes for all 17 findings)

**Modified further:**
- `folio-go/internal/arch_test.go` — `scanNoFloat64` returns `noFloat64Stats`; production caller's
  vacuity guard reads it instead of a second independent walk (Major 5)
- `folio-go/internal/pdf/emit_source_test.go` — `scanNumericFormatting` returns
  `numericFormattingStats` (Major 5); fixture caller split into two subtests, scoped to `pdf/` and
  to the widened parent (Blocker 4)
- `folio-go/fixture_test.go` — gained `checkFixtureShape`, `loadExpectedFixture`,
  `TestFixtureShapeCheckRedProof` (moved from `matrix_test.go`, Blocker 3; second scratch case
  added, Finding 12)
- `folio-go/matrix_test.go` — the three functions above removed (moved out of the `matrix` tag);
  unused `encoding/json` import removed
- `lint/internal/rules/maprange.go` — rewritten on `golang.org/x/tools/go/packages` (Blocker 1,
  D-1.3.11); returns `MapRangeStats`; fails loudly on unresolved types
- `lint/internal/rules/maprange_test.go` — literal-idiom assertion (Blocker 2); stats-based
  vacuity (Major 5); cross-package fixture assertions (Blocker 1)
- `lint/internal/rules/walk.go` — dead `groupGoFilesByDir`/`packageGroup`/`groupedFile`/
  `parseGoFile` removed (only `ScanMapRange`'s withdrawn path used them)
- `lint/internal/rules/forbiddenimports.go` — `isBannedImportPath` prefix-aware (Major 6);
  `RuleMathSelector` split from `RuleForbiddenImports` (Minor 10); non-call message gained the
  allow-listed surface (Minor 11); returns `ForbiddenImportsStats` (Major 5)
- `lint/internal/rules/forbiddenimports_test.go` — stats-based vacuity (Major 5); new fixture want
  entry + rule split (Major 6, Minor 10); new `TestForbiddenImportsMessageContent` (Minor 11)
- `lint/internal/rules/absences.go` — re-keyed on two directories, each with its own rule id
  (Major 8, Minor 10)
- `lint/internal/rules/absences_test.go` — fixture want set updated to two directory-level findings
- `lint/internal/rules/licencegraph.go` — dead `FamilyCommercialEULA` case removed (Minor 9)
- `lint/internal/licence/classify.go` — `FamilyCommercialEULA` member removed (Minor 9)
- `lint/go.mod`, `lint/go.sum` — `require golang.org/x/tools v0.49.0` (Blocker 1, D-1.3.11); go.sum
  covers the full transitive graph after `go mod download all`
- `lint/MANIFEST.md` — regenerated (`go run ./cmd/genmanifest`); no longer empty (Blocker 1)
- `.github/workflows/ci.yml` — `go-version: ${{ env.GO_VERSION }}` in all three jobs (Minor 13);
  `lint` job gained a `go mod download all` step before its `GOPROXY=off go test ./...` step
  (needed for Blocker 1's fix to resolve hermetically in a fresh clone — discovered validating)
- `_bmad-output/implementation-artifacts/deferred-work.md` — split into `## Done` (DW-1, with a
  corrected closure note, Blocker 3/Minor 14) and `## Open` (DW-2/3/4); DW-2 gained a correction
  note (Major 8)
- `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`
  — one-space alignment fix on the `lint/` tree entry (Nit 15)
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md` — D-000.6 amendment's "After"
  block corrected to match (Nit 15); `D-1.3.11` already present (logged pre-finish, applied here)

**Renamed / moved (Blocker 4):**
- `folio-go/testdata/lint/numeric-formatting/violating.go` → `.../pdf/violating.go`
- `folio-go/testdata/lint/numeric-formatting/template/numbers.go` → `.../template/version.go`

**Added:**
- `lint/go.sum`
- `lint/cmd/genmanifest/main_test.go` — `TestFindRepoRoot` (Nit 17)
- `folio-go/testdata/lint/map-range/crosspkg/producer/types.go` — retained cross-package regression
  fixture (Blocker 1)
- `folio-go/testdata/lint/map-range/crosspkg/consumer/violating_cross_package.go`,
  `violating_cross_package_alias.go` (Blocker 1)
- `folio-go/testdata/lint/forbidden-imports/violating_subpackage_imports.go` (Major 6)

## QA Results

## Review Summary
- **Reviewed by:** bmad-code-reviewer (fresh context; no knowledge of the developer's reasoning)
- **Date:** 2026-08-23
- **Baseline:** `f9c27b3` (= HEAD; all story changes uncommitted)
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 4
- **Majors:** 4
- **Minors:** 6
- **Nits:** 3

### Gates re-measured in this review (raw exit codes, `rtk proxy`, not carried forward)

```
folio-go:   go build ./...            → 0     go vet ./...              → 0
            gofmt -l .                → 0 (no output)
            go test ./... -count=1    → 0   (25 top-level / 44 with subtests / 4 packages)
            go build -tags=matrix ./… → 0     go vet -tags=matrix ./…   → 0
            go list -m all            → github.com/panitw/folio/folio-go  (one line; no go.sum)  ✔ AC24
hashmatrix: go vet ./...              → 0     go build -o "$tmp/probe" ./probe → 0     gofmt -l . → 0
lint:       go build ./...            → 0     go vet ./...              → 0     gofmt -l . → 0
            go test ./... -count=1    → 0   (3 packages + cmd/genmanifest [no test files])
licence fixture graphs: GOPROXY=off GOFLAGS=-mod=mod go list -m all → 0 for copyleft/, permissive/, unknown/
`git status --porcelain` identical before and after this review; `fixtures/` untouched.
```

**No gate discrepancy against the numbers reported in the Delivery Log**, with one exception recorded
as Blocker 3 below: the *composition* of the 25/44 count is misstated.

### Verified as correct (recorded so the finisher does not re-litigate them)

- **AC24 / RP-11** — `folio-go`'s module graph is exactly one line; `go.mod` has no `require`, no `go.sum`.
- **AC9** — the Story 1.1 `Superseded` note is a **pure append** (21 insertions, all after the last
  existing line); lines **202**, **327**, **678** are byte-identical to `HEAD`; `epics.md` untouched.
- **AC26** — the spine diff is 5 insertions confined to the §Source tree block; **no invariant's Binds
  or Prevents line is touched**; before/after quoted verbatim in the decision log.
- **AC30** — every `require` in all three fixture graphs has a matching local `replace` (including
  `unknown/`); all three resolve under `GOPROXY=off`; module paths are under `example.test/`; each
  stub carries a "contains NO third-party code" `NOTICE.md`; a repo-wide scan for full copyleft text
  finds only the classifier's own short title markers in `lint/internal/licence/classify.go` and its
  test — no licence text; `lint/testdata/** linguist-vendored` present in `.gitattributes`.
- **Scoping (D-1.3.7, stated positively)** — every source scanner is rooted on `folio-go/internal/`
  (or a `folio-go/testdata/lint/` fixture root). **No scanner is pointed at any directory merely named
  `internal/`**; `lint/internal/`, which necessarily contains `"Sin"`, `"math"` and `float64` as data,
  is never scanned. `hashmatrix/` is not named in any guard. The reverse hazard (a production scan
  asserting zero against the wrong or empty tree) is closed for the *root* by the geom/pdf-by-name
  vacuity guards — but see Major 5 for the part that is **not** closed.
- **AC5 caller ordering** — all six production callers (`TestNoFloat64UnderInternal`,
  `TestNumberFormattingIsConfinedToNumbersGo`, `TestForbiddenImportsProductionScan`,
  `TestMapRangeProductionScan`, `TestAbsencesProductionScan`, `TestLicenceGraphProductionScan`) fail on
  the error in a **separate `if` statement placed before** the zero-findings assertion. Verified by
  reading all six, not by trusting RP-3b.
- **AC19 / RP-15 — `TestRenderRowShapeIncludesServesAndLabel` is genuine, not vacuous.** Red-proved in
  an isolated copy: dropping `r.Serves` from `Render`'s row `Fprintf` reddens it on two independent
  assertions (`row is missing its "serves" column` and `row has 5 column separators, want at least 6`).
  This one is the story's strongest piece of work.
- **AC13 fixture discrimination** — each of the four violating `forbidden-imports` fixtures can be
  reported for exactly one reason, so the by-path assertion does discriminate.
- **In-plain-terms opener** — 350 words (body, excluding the heading), inside the 150–350 range, and it
  states the scope fence ("Deliberately not done here…" / "Done looks like…").
- **Plus:** no ruling paraphrase found in the shipped code comments or the workflow — D-1.3.x are quoted
  and cited by id throughout, and no justification cites a self-disclaiming `F-N`. (The one
  self-disclaiming *artifact* is the AC8 fixture — Blocker 4.)

---

### Finding 1: `ScanMapRange` silently misses every map whose type is named in another package — a dead detector indistinguishable from a clean one
- **Severity**: Blocker
- **Category**: Correctness / AC Conformance
- **Location**: `lint/internal/rules/maprange.go:36-99` (esp. `82-85`, `73`, `66`), `lint/internal/rules/maprange.go:109-116` (`tolerantImporter`)
- **Observation**: The dependency-free `go/types` path type-checks **one source directory at a time**
  (`groupGoFilesByDir`), and `tolerantImporter.Import` returns an **empty `types.Package`** for any
  import `importer.Default()` cannot resolve — which includes **every sibling package under
  `folio-go/internal/`**, because they are never given to the checker. A `RangeStmt` whose subject
  type therefore fails to resolve falls through `if !ok || tv.Type == nil { return true }` and is
  **not reported**, with `err == nil`. `conf.Check`'s error is discarded twice over (`_, _ =` at
  line 73, and a no-op `Error` callback at line 66), so nothing is ever reported as unavailable
  type information.

  Proven by construction against a **copy of the real tree**, not a synthetic fixture. Adding to
  `folio-go/internal/geom/` a named map type that a future story would plausibly add:

  ```go
  package geom
  type ScaleTable map[string]Length
  func NewScaleTable() ScaleTable { return ScaleTable{} }
  ```

  and to `folio-go/internal/pdf/` two genuine map ranges over it:

  ```go
  func probeNamedGeomMap(t geom.ScaleTable) geom.Length { var total geom.Length; for _, v := range t { total += v }; return total }
  func probeGeomReturn() geom.Length   { var total geom.Length; for _, v := range geom.NewScaleTable() { total += v }; return total }
  ```

  `ScanMapRange("folio-go/internal")` returns **`err=<nil> findings=0`**. `go build ./...` and
  `go vet ./...` on the same tree both exit **0** — this is valid, compiling, shipping Go.

  Three further misses confirmed on an isolated fixture tree (`err=<nil> findings=0` for all three):
  a named map type imported from another package, a **type alias** (`type Alias = map[string]int`)
  imported from another package, and a **function's map return value** from another package.
  Same-package (two files) and stdlib map types (`http.Header`) *are* caught — which is why the
  retained fixture, whose violator is a locally-spelled `map[string]int` parameter, is green and
  proves nothing about the failing case.

- **Impact**: This is exactly the **V10** shape the story names as its standing suspicion, in the one
  guard where the story declares the cost of the rule is currently zero (F-7). The production scan
  asserts zero and passes; its vacuity guard passes; the fixture scan passes. **Nothing in the suite
  distinguishes "no map ranges" from "map ranges the checker cannot see."** `internal/pdf` already
  imports `internal/geom` today, so the hazard is live now, not hypothetical: the first map keyed by
  or valued with a `geom` type — the single most likely shape in this codebase — is invisible.
  Against the rulings: **D-1.3.5** requires detection "exact and type-based"; **D-1.3.6**'s binding
  invariant **(a) is "exact detection with no false positives"**, and the dependency-free `go/types`
  path is permitted only *"if the developer finds an **exact**, dependency-free path"*. What was
  shipped is exact where types happen to resolve and blind where they do not — a *best-effort* path,
  which the `go.mod` comment and the `ScanMapRange` doc comment both say out loud
  ("best effort type information", "not a fully sound whole-program check"). A best-effort detector
  is not the permitted refinement; it is a third thing the ruling did not authorise.
- **Suggested Resolution**: Either (a) take D-1.3.6's *expected* path — `golang.org/x/tools/go/packages`
  in `lint/`'s own module graph, which is explicitly sanctioned and which invariant (b) permits since
  it never touches `folio-go`'s graph — or (b) make the dependency-free path exact: type-check the
  whole target tree as a coherent import graph (resolve intra-target imports from source, in
  dependency order, feeding each package's `*types.Package` back into the importer) and, crucially,
  **fail loudly**: return a non-nil error whenever a `RangeStmt` subject's type could not be resolved,
  so an unresolvable subject reddens rather than passes. Add a regression fixture pair for the
  cross-package case. Either way, remove the tolerant `Error` no-op or make what it swallows visible.
  If the choice between (a) and (b) is not the developer's to make, it is a `DECISION NEEDED`
  (D-1.2.6) — the dependency-free path was taken as "a judgement call within an explicitly
  pre-approved option", and this finding is evidence the precondition of that option was not met.
- **Related AC**: AC14, AC17, AC1; **D-1.3.5**, **D-1.3.6(a)**; V10; NFR1.d, AD-1.

### Finding 2: AC15's "names the escape hatch verbatim" assertion is a tautology — replacing the idiom with `"TODO"` keeps the test green
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `lint/internal/rules/maprange_test.go:79-96`; `lint/internal/rules/maprange.go:15` and `90-91`
- **Observation**: The message is built as `fmt.Sprintf("… use instead: %s", EscapeHatch)` and the test
  asserts `strings.Contains(f.Message, EscapeHatch)` — comparing the message against **the same
  constant that produced it**. Both sides move together. Red-proved in an isolated copy of `lint/`:
  replacing line 15 with `const EscapeHatch = "TODO"` leaves
  `TestMapRangeFailureMessageNamesEscapeHatch` **PASS**, exit 0. The literal idiom AC15 fixes —
  `for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }` — is asserted **nowhere**.
- **Impact**: AC15 is the clause that makes D-1.3.5 survivable for the next developer: *"the first
  developer to hit the rule is unblocked by the error text itself rather than by finding a story
  file."* That text is currently protected by nothing. The shipped constant happens to be correct
  today, so the defect is invisible — healthy output and dead output are the same value (V10 again).
  RP-7's own row concedes this in writing ("N/A — always green because the constant is embedded
  correctly"), which under **D-1.2.5** disqualifies it from warranting AC15.
- **Suggested Resolution**: Assert the **literal string** in the test, written out independently of
  the production constant — e.g. `const wantIdiom = "for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }"`
  declared in `maprange_test.go`, then `strings.Contains(f.Message, wantIdiom)` **and**
  `EscapeHatch == wantIdiom`. Re-run the `"TODO"` mutation and confirm it reddens before calling
  RP-7 discharged.
- **Related AC**: AC15, AC28; RP-7; D-1.3.5; V10.

### Finding 3: DW-1's red-proof lives behind `//go:build matrix` and never runs in any gate — and the Delivery Log and CI both state that it does
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/matrix_test.go:1` (`//go:build matrix`), `folio-go/matrix_test.go:449-...`
  (`TestFixtureShapeCheckRedProof`); Delivery Log "Gates measured in this story"; `.github/workflows/ci.yml:10`
- **Observation**: `TestFixtureShapeCheckRedProof` — AC6's and RP-4's entire closure for DW-1 — is in
  `matrix_test.go`, which carries the `matrix` build tag. Measured:
  `cd folio-go && go test ./... -run TestFixtureShapeCheckRedProof` reports
  **`no tests to run` in all four packages**. It is absent from the `-v` run's 25 top-level tests.
  Three consequences, all measured:
  1. The Delivery Log states the 25/44 count includes "TestNoFloat64FixtureScan,
     TestNumberFormattingFixtureScan **and TestFixtureShapeCheckRedProof** as new top-level tests."
     The count (25/44) is correct; the attribution is **false** — only the first two appear.
  2. `ci.yml:7-10` states DW-1's fixture-shape red-proof lives in "folio-go's **untagged** test
     suite". It does not. CI's only `-tags=matrix` steps are `go build` and `go vet` — neither runs
     a test.
  3. The matrix suite is **deliberately not run** in this story and is deferred to the Epic 1
     boundary (D-000.4). So the test executes in **zero** gates, now and until Epic 1 closes.
- **Impact**: AC6 requires DW-1 to be "red-proved against a scratch copy". The red-proof exists as
  source and has never executed under any gate this story ran; a test that never runs and a passing
  test are the same value. `deferred-work.md` marks DW-1 **DONE** on the strength of it. This is the
  same defect class as Finding 2 wearing different clothes, and it is precisely what D-1.3.7's
  evidence rule was added to prevent.
- **Suggested Resolution**: Move `checkFixtureShape` and `TestFixtureShapeCheckRedProof` out of the
  tagged file into an **untagged** file in `package folio` (the shape check itself has nothing to do
  with Docker or Node — `folio-go/fixture_test.go` already holds `isSHA256HexString` untagged and is
  the natural home). Re-run `go test ./...` and confirm the test appears in the run before marking
  DW-1 done. Correct the Delivery Log's test attribution and `ci.yml`'s comment.
- **Related AC**: AC6, AC28; RP-4; DW-1; D-000.4; D-1.2.5; V10.

### Finding 4: AC8's "must now pass" fixture is exempt by the pre-existing `numbers.go` filename rule, not by D-1.3.2's narrowing — it would be equally green before the change
- **Severity**: Blocker
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/testdata/lint/numeric-formatting/template/numbers.go`;
  `folio-go/internal/pdf/emit_source_test.go:126-128` and `:251-264`
- **Observation**: AC8 requires *"`fmt.Errorf` in a fixture package **standing in for
  `internal/template`** is **not reported**"*, retained in the tree. The shipped fixture is a
  `fmt.Errorf`/`versionError` file **named `numbers.go`** — and `scanNumericFormatting` returns early
  for `filepath.Base(path) == "numbers.go"` at any depth. That exemption is the guard's original,
  pre-existing carve-out; it applied identically under the deleted module-wide rule. The fixture
  therefore demonstrates the `numbers.go` exemption and says **nothing** about D-1.3.2.
  Proved on an isolated copy: renaming the fixture to `version.go` (the name Story 1.4's file will
  actually have, per F-2) makes `TestNumberFormattingFixtureScan` **FAIL** —
  `unexpected finding reported: file=template/version.go rule=numeric-formatting`. The filename is
  the only thing doing the work.
- **Impact**: **V9 is a named vacuity guard in this very story** — *"AC8's compliant fixture must stand
  in the tree and be **scanned**. A package that is absent is not a proof."* The fixture is scanned,
  but its non-reporting is produced by a mechanism unrelated to the AC. The narrowing that actually
  unblocks Story 1.4 (the production caller rooting at `internal/pdf/`) has **no retained fixture at
  all**; nothing in the suite would notice if the production caller's root were widened back to
  `folio-go/`. The fixture's own doc comment and the Delivery Log's D-1.2.6 disclosure both state
  outright that "the real Story 1.4 file will not be named `numbers.go`" and that the real unblock is
  scope — a **self-disclaiming artifact** cited as the AC's evidence, which D-1.2.5 forbids.
- **Suggested Resolution**: Restructure the fixture tree so the scope claim is the thing under test —
  e.g. `folio-go/testdata/lint/numeric-formatting/pdf/{violating.go, numbers.go}` standing in for
  `internal/pdf`, and `folio-go/testdata/lint/numeric-formatting/template/version.go` standing in for
  `internal/template`; then have `TestNumberFormattingFixtureScan` assert that scanning the `pdf/`
  root reports `violating.go` and **only** `violating.go`, and add an assertion that `version.go` is
  outside that root and thus never reached. Red-prove by widening the fixture caller's root to the
  parent and confirming `template/version.go` appears.
- **Related AC**: AC8, AC7, AC28; **V9**; D-1.3.2, D-1.2.5.

### Finding 5: all four production-scan vacuity guards re-walk the tree with a *different* function, so a completely dead checker still passes them
- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `lint/internal/rules/maprange_test.go:25-45`;
  `lint/internal/rules/forbiddenimports_test.go:28-31, 42-58`;
  `folio-go/internal/arch_test.go:195-220`; `folio-go/internal/pdf/emit_source_test.go:215-227`
- **Observation**: Each production caller runs the scanner, then **separately re-walks the same tree**
  (via `groupGoFilesByDir`, `countFilesImporting`, a second `walkGoFiles`, a second
  `filepath.WalkDir`) to prove "files were seen". The evidence is about the *directory*, never about
  the *scanner*. Red-proved on an isolated copy: injecting `if true { return nil, nil }` as the first
  statement of `ScanMapRange` — a totally dead detector — leaves `TestMapRangeProductionScan`
  **PASS**. The same construction applies to the other three.
- **Impact**: AC17 states the production scan "cannot pass by having parsed nothing: it asserts it
  visited the `geom` and `pdf` package directories **by name** and parsed a non-zero number of
  files". As built it asserts that *the vacuity walk* parsed files, which is true whether or not the
  scanner ran. A dead scanner is still caught by the fixture callers, so this is not on its own a
  blocker — but it is the reason Finding 1 is invisible: the guard AC17 nominates as the non-vacuity
  proof cannot see the failure it exists to catch.
- **Suggested Resolution**: Have the checker itself report what it examined — e.g. return a
  `ScanStats{DirsVisited []string, FilesParsed, ExprsTyped int}` alongside `(findings, error)` — and
  assert `geom`/`pdf` and the non-zero counts **from the scanner's own output**. For map-range,
  additionally assert a non-zero count of `RangeStmt` subjects whose type **did** resolve, which is
  the statistic that would have made Finding 1 visible.
- **Related AC**: AC17, AC1; **V2**, V10.

### Finding 6: `bannedImportPaths` is an exact-match deny-list — `math/rand/v2`, `net/http`, `net/url` and `os/exec` all pass
- **Severity**: Major
- **Category**: Correctness / Convention
- **Location**: `lint/internal/rules/forbiddenimports.go:19-24`
- **Observation**: The map keys are the four literal paths `time`, `os`, `math/rand`, `net`. Measured
  on an isolated fixture: a **non-test** file importing `mr "math/rand/v2"`, `"net/http"`,
  `"net/url"` and `"os/exec"` and calling `mr.IntN`, `http.Client`, `url.URL`, `exec.Command`
  produces **`findings=0`, `err=<nil>`**.
- **Impact**: `math/rand/v2` is not a future hypothetical — it shipped in Go 1.22 and this module pins
  Go 1.26; it is the *modern* spelling of the exact package AD-1 bans, and it is a direct source of
  render-path nondeterminism (NFR1.a). The story applied precisely the right reasoning to the `math`
  rule — D-1.3.10: *"AD-1's transcendental list ends in '…' and is therefore not decidable. Its
  allow-list is closed. Inverting the test makes the rule decidable **and** fail-safe when a future
  Go release adds a transcendental"* — and then shipped the import rule in the opposite, non-fail-safe
  shape. AD-1's **Prevents** line is "determinism eroding one reasonable-looking commit at a time";
  `import "math/rand/v2"` is that commit.
- **Suggested Resolution**: Match a banned path **and any subpackage of it** — `p == banned ||
  strings.HasPrefix(p, banned+"/")` — which covers `math/rand/v2`, `net/http`, `os/exec` in one line,
  with the `_test.go` exemption keeping its own exact-match semantics for `os`/`testing`/
  `path/filepath`/`embed` (D-1.3.1: nothing is added to that list). Add a violating fixture importing
  `math/rand/v2` and a `net/*` package, and confirm the production scan stays green (nothing under
  `internal/` imports either today).
- **Related AC**: AC10, AC11; AD-1; NFR1.c; D-1.3.10's own rationale, applied inconsistently.

### Finding 7: five of the fifteen red-proofs were never observed red against a shipped assertion, contradicting the section's own header
- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: Delivery Log, "Red-proofs — RP-1 … RP-15" table, rows RP-3b, RP-3c, RP-6, RP-7, RP-14
- **Observation**: The table's preamble states *"Every mutation below was applied, observed red
  (quoted), and reverted"*, and the Red-proofs section states *"Every mutation is **observed red,
  then reverted**."* The rows themselves say otherwise: RP-3b is "Confirmed (by **code inspection**
  of the shipped, unmutated callers) … **N/A** — verified the invariant holds as shipped"; RP-3c is
  a hypothetical count-only comparison built alongside, "**N/A** — no shipped code touched"; RP-6 and
  RP-14 are "**Probe:** … *would* pass"; RP-7 is "**N/A** — always green".
- **Impact**: A code inspection, a hypothetical, and a self-declared always-green test are not
  red-proofs. Two of the five conceal real defects that a genuine mutation would have surfaced
  immediately: RP-7 is Finding 2 (the mutation *was* available and reddens nothing), and RP-6's
  "syntactic guess" probe tested for false **positives** while Finding 1's false **negatives** went
  unprobed. Under D-1.2.5 a claim that disclaims measurement warrants nothing; the header claiming
  otherwise makes the table read as stronger evidence than it is.
- **Suggested Resolution**: Either run each of the five as a real mutation against the shipped
  assertion and quote the observed red, or reword the header and each row to state plainly what was
  and was not executed. RP-7 and RP-3c should become shipped tests (Finding 2 gives RP-7's shape);
  RP-14 should be run by actually adding a `lint`-named exemption to the graph walk and observing
  `TestLicenceGraphFixtureScan/copyleft` go red.
- **Related AC**: AC28; D-1.2.5, D-1.3.7's evidence rule.

### Finding 8: the three asserted absences key on guessed exact filenames, and the code comment concedes the real paths are unknown — the guard may never fire
- **Severity**: Major
- **Category**: Correctness / AC Conformance
- **Location**: `lint/internal/rules/absences.go:25-45` (and the comment at `:26-28`)
- **Observation**: The checks are `folio-designer/package-lock.json`, `folio-go/fonts/OFL.txt`, and
  `folio-designer/third-party-notices/pdfjs-dist/NOTICE` — three exact paths. The doc comment states:
  *"The exact real-world path each future story lands its artifact at is **that story's own call**;
  this guard's job is only to fail loudly the day something appears there."* The spine describes
  `folio-designer/` as "Vite + React + TS"; a pnpm, yarn or bun lockfile (`pnpm-lock.yaml`,
  `yarn.lock`, `bun.lockb`) satisfies none of these checks. `folio-go/fonts/OFL.txt` likewise
  presumes one filename out of several plausible ones (`LICENSE.txt`, `OFL-1.1.txt`, per-family
  copies), and the `third-party-notices/pdfjs-dist/NOTICE` layout is invented outright.
- **Impact**: This is D-1.3.4's own rejected hazard arriving through a side door. The ruling refused a
  conditional "check it if present" because *"the guard reports success precisely when it stops
  covering anything"* — a filename guess produces the identical outcome: Story 5.1 adds
  `pnpm-lock.yaml`, the build stays green, and the JS licence half is never forced. V7 says the
  absence assertion passing today proves nothing; it will also pass forever if the guess is wrong,
  and nothing announces it.
- **Suggested Resolution**: Key each absence on the artifact whose existence is **not** a future
  story's call: assert the **directory** `folio-designer/` is absent (Story 5.1 cannot create the
  project without it) and the **directory** `folio-go/fonts/` is absent (the spine fixes that path,
  and Story 2.2 cannot ship faces without it). Where a directory-level key is genuinely wrong,
  match a **set** of plausible names (all four JS lockfile spellings; `OFL*`/`*OFL*` under
  `folio-go/fonts/`) rather than one. Keep the fixture root's violating files aligned with whatever
  keys are chosen.
- **Related AC**: AC21; **V7**; D-1.3.4; DW-2.

### Finding 9: `FamilyCommercialEULA` is unreachable — AD-26's commercial-EULA family has no classifier path, no fixture and no test
- **Severity**: Minor
- **Category**: Correctness / Tests
- **Location**: `lint/internal/licence/classify.go:18, 56-91`; `lint/internal/rules/licencegraph.go:49-53`
- **Observation**: `classifyBySPDX` returns only `FamilyPermissive`, `FamilyCopyleft` or
  `FamilyUnknown`, and the text fallback's `switch` has no EULA arm. `FamilyCommercialEULA` is
  therefore never produced, and `licencegraph.go`'s `case licence.FamilyCommercialEULA` is dead code.
  `TestClassifyLicenceText` has no EULA case; the fixture graphs have no EULA stub.
- **Impact**: A commercial EULA still **fails** the build — it falls through to `FamilyUnknown` — so
  AD-26's polarity holds by accident. But the failure message reads "licence unresolvable — could not
  classify licence text" rather than naming the EULA, and the story's `Family` enum advertises a
  capability that does not exist. An unused enum member is a half-wired feature: the next developer
  reasonably assumes EULA detection is covered.
- **Suggested Resolution**: Either add real EULA detection (a short marker table plus SPDX ids such as
  the `LicenseRef-`/proprietary families) with a classifier unit-test case and a fixture stub, or
  delete `FamilyCommercialEULA` and the dead `case`, and state in a comment that AD-26's EULA half is
  covered by the unresolvable path. Do not leave it advertised-but-inert.
- **Related AC**: AC18, AC19, AC31; AD-26.

### Finding 10: `absenceCheck.rule` is dead, and every absence finding carries the same rule id, so "by file **and rule**" degrades to "by file"
- **Severity**: Minor
- **Category**: Maintainability / Convention
- **Location**: `lint/internal/rules/absences.go:19-23` (`rule` field), `:59-62` (uses `RuleAbsences`)
- **Observation**: Three distinct rule ids are declared — `absence-designer-lockfile`,
  `absence-ofl-licence-text`, `absence-pdfjs-notice` — and **never read**; every finding is emitted
  with the generic `RuleAbsences`. The same flattening applies to `ScanForbiddenImports`, where an
  import-ban violation and a `math`-selector violation are both `RuleForbiddenImports`.
- **Impact**: AC1/AC4's "by file **and by rule**" contributes nothing when one checker emits exactly
  one rule id; the assertions are effectively by path alone. It happens to discriminate today because
  every fixture path is distinct and single-cause (verified), but the guard AC4 intends is not
  actually present, and the declared-then-discarded ids are the tell.
- **Suggested Resolution**: Emit `c.rule` in `ScanAbsences`, and split `RuleForbiddenImports` into
  (at minimum) a banned-import id and a `math`-selector id. Update the three fixture-caller expected
  sets accordingly and confirm they still pass.
- **Related AC**: AC1, AC4.

### Finding 11: AC10's failure-message requirements are asserted nowhere, and the non-call `math` message omits the allow-listed numeric surface
- **Severity**: Minor
- **Category**: Tests / AC Conformance
- **Location**: `lint/internal/rules/forbiddenimports.go:156-162`;
  `lint/internal/rules/forbiddenimports_test.go` (no message assertion);
  `lint/internal/rules/testutil_test.go:44-67` (`assertExactFindings` compares `(path, rule)` only)
- **Observation**: AC10 requires the failure to name the file, the offending import or selector,
  **and** "AD-1's allow-listed numeric surface: `+ - * /`, comparison, and `Sqrt`, `Floor`, `Ceil`,
  `Round`, `Trunc`, `Abs`, `Mod`". The banned-import and `math`-call messages include
  `allowedNumericSurface`; the **non-call reference** message (line 159, the `math.Pi` path) does not.
  No test in the suite inspects any `Finding.Message` or `Finding.Line` for this rule —
  `assertExactFindings` discards both fields.
- **Impact**: Two of AC10's three message clauses have no covering assertion, and one is already
  unmet on the `math.Pi` path. Same shape as Finding 2 at lower stakes.
- **Suggested Resolution**: Append `allowedNumericSurface` to the non-call message, and add a
  message-content assertion to `TestForbiddenImportsFixtureScan` (asserting against literals written
  independently of the production constants — see Finding 2).
- **Related AC**: AC10, AC12.

### Finding 12: the DW-1 red-proof exercises only the JSON-unmarshal failure, never `isSHA256HexString`, and the inventory's second scratch case was not built
- **Severity**: Minor
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/matrix_test.go` (`TestFixtureShapeCheckRedProof`, the widened-`sha256` scratch
  copy); fixture inventory row `fixture-shape`
- **Observation**: The scratch copy replaces `sha256` with a JSON **object**, so `json.Unmarshal` into
  the `string`-typed field fails and `checkFixtureShape` returns at the *parse* step — the
  `isSHA256HexString` branch is never reached. The fixture inventory promises two scratch shapes:
  "a scratch `expected.json` whose `sha256` is a per-target object, **and one whose value is
  upper-case or 63 characters**". Only the first exists.
- **Impact**: AC16's actual hex-shape predicate (64 lower-case hex characters) has no red-proof. Given
  Finding 3, neither branch runs in any gate at present.
- **Suggested Resolution**: Add the second scratch case (a valid JSON string of 63 characters, and one
  upper-case) and assert `checkFixtureShape` errors on each — after moving the test out of the
  `matrix` build tag per Finding 3.
- **Related AC**: AC6, AC16; RP-4; DW-1.

### Finding 13: `ci.yml` declares `env: GO_VERSION` and never uses it; the version is hard-coded three times
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `.github/workflows/ci.yml:23-24`, `:33`, `:58`, `:78`
- **Observation**: `env: GO_VERSION: "1.26.0"` is declared at workflow level and referenced nowhere;
  each of the three `setup-go` steps hard-codes `go-version: "1.26.0"`.
- **Impact**: The declared knob does nothing. Under AD-22 (toolchain drift is a byte-identity hazard) a
  future bump made through the obvious lever would silently change nothing, leaving CI on the old
  toolchain while the author believed otherwise.
- **Suggested Resolution**: Use `go-version: ${{ env.GO_VERSION }}` in all three jobs, or delete the
  unused `env` block.
- **Related AC**: AC25; AD-22.

### Finding 14: DW-1 is marked DONE but still filed under `## Open`, and carries no closing commit
- **Severity**: Minor
- **Category**: Convention / AC Conformance
- **Location**: `_bmad-output/implementation-artifacts/deferred-work.md:12-24`
- **Observation**: The file's own convention (line 4-5) is *"Append-only; mark items done in place
  **with the commit that closed them**."* DW-1's heading now reads "— **DONE**" while remaining under
  the `## Open` section, and its `Closed by:` line says "Story 1.3, in this story's commit (AC6)" —
  a forward reference with no SHA, because the story is uncommitted. (`Committed at: f9c27b3` is the
  *deferral* commit, not the closure.) AC6 requires DW-1 marked "done in place **with this story's
  commit**".
- **Impact**: The `## Open` heading is now false for its first entry; a reader scanning headings sees
  four open items. The commit reference cannot be verified at review time.
- **Suggested Resolution**: The finisher should fill in the real SHA after committing, and either move
  DW-1 under a `## Done` heading or retitle `## Open` to reflect done-in-place entries. Note that per
  Finding 3, DW-1 should not be marked done at all until its red-proof actually runs.
- **Related AC**: AC6, AC22.

### Finding 15: the spine's `lint/` entry is misaligned by one column against `hashmatrix/`'s
- **Severity**: Nit
- **Category**: Convention
- **Location**: `ARCHITECTURE-SPINE.md` §Source tree, line 625
- **Observation**: Measured column of `#` per line: `hashmatrix/` → 39; its continuation lines → 39;
  `lint/` → **40**; `lint/`'s continuation lines → 39; `folio-node/ · folio-java/ · …` → 43.
- **Impact**: AC26 requires the entry be "in the **same shape** as `hashmatrix/`'s". Cosmetic, but it
  is also inconsistent with `lint/`'s **own** continuation lines, so the comment block does not line
  up with itself.
- **Suggested Resolution**: Remove one space before the `#` on line 625. Re-quote the corrected
  "After" block in the decision log's D-000.6 amendment, which reproduces the same misalignment.
- **Related AC**: AC26; D-000.6.

### Finding 16: AC13's heading says "six fixtures"; it then enumerates eight
- **Severity**: Nit
- **Category**: AC Conformance (documentation)
- **Location**: story file, **AC13** heading and body; Task 7's final subtask ("AC13's six fixtures")
- **Observation**: The heading reads "six fixtures for this rule: two violating imports, and four
  near-misses", but the body requires **four** violating fixtures (two `time` imports, one `math`
  call outside the seven, one `math.Pi`) plus **four** near-misses = eight. Eight are shipped and
  eight are asserted; only the count in the prose is wrong. `TestForbiddenImportsFixtureScan`'s doc
  comment repeats "six".
- **Impact**: None functionally; a reviewer counting fixtures against the heading finds a phantom
  discrepancy.
- **Suggested Resolution**: Correct the count in the AC13 heading, Task 7, and the test's doc comment.
- **Related AC**: AC13.

### Finding 17: repo-root discovery is duplicated three times, and the one copy with no test is the one that writes a committed file
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `lint/internal/rules/testutil_test.go:15-37`,
  `lint/internal/manifest/manifest_test.go:11-29`, `lint/cmd/genmanifest/main.go:36-52`
- **Observation**: Three near-identical "walk up until a directory holds both `folio-go/` and `lint/`"
  implementations. Two are in `_test.go` files (the duplication there is explained in comments and is
  reasonable); the third is production code in `genmanifest`, which has **no test file** — the
  Delivery Log describes it as "a two-line CLI wrapper", but it is 53 lines and owns the
  root-resolution logic that decides where `MANIFEST.md` is written.
- **Impact**: Low today. If root discovery drifts between `genmanifest` and `TestManifestUpToDate`,
  the generator writes one path while the test reads another and the mismatch surfaces as a confusing
  "out of date" failure.
- **Suggested Resolution**: Move root discovery into a small exported helper in a non-test package of
  `lint/` and have all three call it, or at minimum add a test for `findRepoRoot`.
- **Related AC**: AC19, AC25.

---

### AC-by-AC disposition

| AC | Verdict |
|---|---|
| AC1 | Satisfied in shape (pure function, two callers, by file+rule) — weakened by Findings 5 and 10 |
| AC2 | **Satisfied** — `d.Name() == "testdata"` by directory name in all three walkers |
| AC3 | **Satisfied** — task order and the F-3 reproduction in RP-1 are consistent |
| AC4 | Satisfied for the path base (target-dir-relative, verified in all scanners); rule-id half weakened — Finding 10 |
| AC5 | **Satisfied** — parse errors propagate; all six production callers use two separate statements in the required order (read, not inferred) |
| AC6 | **NOT satisfied** — Finding 3 (red-proof never executes); Findings 12, 14 adjacent |
| AC7 | **Satisfied** — scope is exactly `internal/pdf/`, module-wide half deleted outright, no exemption list |
| AC8 | **NOT satisfied** — Finding 4 (second direction has no valid retained fixture) |
| AC9 | **Satisfied** — verified as a pure append with lines 202/327/678 byte-identical to `HEAD` |
| AC10 | Partially satisfied — Findings 6 and 11 |
| AC11 | **Satisfied** — `_test.go` suffix keying, closed four-entry exemption set, `time`/`math/rand`/`net` still banned in tests (fixture proves it) |
| AC12 | **Satisfied** — AST + alias resolution, closed allow-list, value-kind test; all four polarity fixtures discriminate. (`math/rand/v2` is AC10's clause, Finding 6) |
| AC12a | **Satisfied** — the seven are kept and the reason is commented |
| AC13 | Satisfied in substance; **the comment near-miss is present and correct** (`compliant_comment_near_miss.go` reproduces `scale.go:31`'s `math.Round` plus five `math.MinInt64` comment mentions and is not reported) — Nit 16 on the count |
| AC14 | **NOT satisfied** — Finding 1 |
| AC15 | **NOT satisfied** — Finding 2 |
| AC16 | Satisfied for the shipped polarities; the violating fixture is the only one the checker can see (Finding 1) |
| AC17 | **NOT satisfied** — Finding 5 |
| AC18 | Satisfied — all three modules scanned uniformly, no build-time relaxation; EULA half is Finding 9 |
| AC19 | **Satisfied** — completeness asserted by `TestManifestUpToDate`; row shape asserted non-vacuously (red-proved in this review); labels present |
| AC20 | Satisfied as adapted — the dependency-free refinement means there is no dependency to cover; the principle is stated in the manifest header as required |
| AC21 | Partially satisfied — mechanism correct (unconditional, seam-based, both fixture polarities); keys are guesses — Finding 8 |
| AC22 | **Satisfied** — DW-2/3/4 open with owners and anti-rot mechanisms, unedited |
| AC23 | **Satisfied** — `lint/` at repo root, correct module path, no `folio-go/cmd/` created |
| AC24 | **Satisfied** — measured this run |
| AC25 | Satisfied — CI invokes only local commands, re-implements nothing; Finding 13 is cosmetic |
| AC26 | Satisfied — Nit 15 on alignment |
| AC27 | **Satisfied** — both existing guards stay put and gained the skip + two-caller shape; no consolidation attempted |
| AC28 | **NOT satisfied** — Finding 7 (five red-proofs are inspections/hypotheticals) and Finding 4 (a self-disclaiming fixture cited as evidence) |
| AC29 | **Satisfied** — three graphs; `copyleft/` fails naming module and licence for all four stubs; `permissive/` passes; `unknown/` fails as unresolvable |
| AC30 | **Satisfied** — verified in full this run (replaces, `GOPROXY=off`, `example.test/`, no-third-party-code headers, SPDX markers only, `linguist-vendored`) |
| AC31 | **Satisfied** — classifier unit tests supplement the graphs; no recorded `go list` output adopted as the graph-walk fixture |

### Note on the cross-target matrix (D-000.4)

Its absence is **not** treated as a defect: the per-epic cadence applies and this story changes no
output bytes. `go build -tags=matrix ./...` and `go vet -tags=matrix ./...` were re-run this session
and both exit **0**, so the refactored `matrix_test.go` still compiles and vets. The separate problem
with that file is Finding 3 — not that the matrix did not run, but that a test which is *not* part of
the matrix was placed behind its tag.

### Review method note

Every claim above was produced by running a command or a mutation in this session. All mutations were
applied to **isolated copies** under the scratchpad (`probe/`, `fg2/`), never to the repository.
`git status --porcelain` is byte-identical before and after this review, and `fixtures/` is untouched.

---

## Finding Resolutions (finisher, post-review)

All 17 findings are **FIX**. None were dismissed or deferred: every one was either proved by
construction on a copy of the real tree (the four blockers) or was a concrete, narrowly-scoped
defect with an actionable suggested resolution (the majors, minors and nits), and none conflicted
with the story's stated scope or with another finding. Per D-1.2.6, no ruling conflicts were
encountered while resolving these; none needed to be surfaced.

**This story named V10 — "healthy output and dead output must never be the same value" — as a
named vacuity guard, and then shipped an instance of it anyway.** Blocker 1 is exactly that: the
dependency-free map-range detector returned `err=<nil>, findings=0` for both "no map ranges exist"
and "map ranges exist but I could not see them", and every guard the story built to catch this
(AC17's vacuity guards, Major 5) re-derived its "did I actually look" signal independently of the
scanner's own execution, so the one place V10 could have been caught, wasn't. That is the most
useful fact in this log for whoever reads it next.

**One further defect, found while staging this commit, not among the 17 named findings.** The
repository's root `.gitignore` bans `*.test` (meant only for compiled Go test binaries, e.g.
`go test -c`'s output), but D-1.3.8's fixture module paths deliberately live under `example.test/`
— the RFC 2606-reserved TLD, chosen so a fake fixture module path can never collide with a real
one. `*.test` glob-matches a directory named `example.test` too, so every file under every fixture
stub (`LICENSE`, `NOTICE.md`, `go.mod` — 23 files across all three graphs) was silently untracked
and would never have been committed at all, despite `git status --porcelain` showing `lint/` as
present: an untracked *directory* reads the same in that output whether its contents are ordinary
untracked files or ignored ones. `AC29`/`AC30`'s entire retained-fixture claim depended on these
files actually landing in the repository. Fixed with a narrow, explained negation in `.gitignore`
(`!lint/testdata/**/example.test/` and its `/**` sibling) rather than removing or narrowing the
`*.test` rule itself, which correctly protects the rest of the repo. Verified: all 23 files now
appear as `git add`-able (`A`) in `git status --porcelain`, none show as `??` or are silently
absent.

### Blocker 1 — `ScanMapRange` silently misses every map whose type is named in another package
**Decision:** FIX, applying **D-1.3.11** verbatim (already ruled by the orchestrator, applying
D-1.3.6 as written — not re-decided here). `D-1.3.11`'s verdict: *"`lint/` adds
`golang.org/x/tools` (BSD-3-Clause) and resolves map types through `go/packages` with full type
information. The stdlib-only `tolerantImporter` path is withdrawn."*
**What changed:** `lint/internal/rules/maprange.go` rewritten to load the target subtree with
`golang.org/x/tools/go/packages` (`Mode: NeedName|NeedFiles|NeedSyntax|NeedTypes|NeedTypesInfo|
NeedImports|NeedDeps`, pattern `./...`), giving every package full type information for its whole
import graph rather than one directory at a time with a tolerant, empty-package-substituting
importer. An `*ast.RangeStmt` whose subject's type cannot be resolved is now a **hard error**
(`ScanMapRange` returns `(nil, MapRangeStats{}, err)`), never a silent skip — the specific defect
the finding named ("fail loudly … so an unresolvable subject reddens rather than passes").
`tolerantImporter`, `pathBase` and the now-unused `groupGoFilesByDir`/`packageGroup`/`groupedFile`
helpers in `lint/internal/rules/walk.go` were deleted.
`lint/go.mod` gained `require golang.org/x/tools v0.49.0` (plus its own indirect deps: `x/mod`,
`x/sync`, `x/net`, `x/sys`, `x/telemetry`, `github.com/yuin/goldmark`, `github.com/google/go-cmp`
— all BSD-3-Clause or MIT). `folio-go/go.mod` is untouched; `cd folio-go && go list -m all` still
prints exactly one line — invariant (b) verified, re-measured this session.
**Retained regression fixture:** `folio-go/testdata/lint/map-range/crosspkg/{producer,consumer}/`
— a named map type (`producer.ScaleTable`), a type alias (`producer.Alias`), and a map-returning
function of each, all declared in `producer` and ranged over in `consumer` from a parameter and
from a direct call — the exact three shapes the review measured missed. Now reported (2 findings,
by file+rule).
**Red-proved, live, this session:** the original `tolerantImporter`-based `ScanMapRange`
(reconstructed byte-for-byte from the pre-fix source and run as an isolated probe, never touching
the repo) returns `err=<nil> findings=1` against the retained fixture tree — it sees
`violating_map_range.go` but is blind to both `crosspkg/consumer` files. The rewritten
`ScanMapRange` returns `findings=3` against the same tree (all three reported). Before: red
(blind). After: green (all three caught).
**New non-vacuity statistic (ties to Major 5):** `MapRangeStats.TypedRangeStmts` — a count of
`*ast.RangeStmt` subjects whose type resolved. Asserted non-zero in `TestMapRangeFixtureScan` (the
fixture tree has real range statements to type); **not** asserted in `TestMapRangeProductionScan`,
because F-7 measured that every `range` site under `folio-go/internal/` today is inside a
`_test.go` file, which `ScanMapRange` never loads (`packages.Config.Tests` is `false`, D-1.3.5) —
so the real, non-test tree legitimately types zero range statements right now, and asserting
otherwise there would be a permanent false failure, not a guard.
**Manifest side effect (as anticipated in the task brief):** `lint/MANIFEST.md` regenerated via
`go run ./cmd/genmanifest` — no longer empty; all eight resolved dependencies are permissive
(BSD-3-Clause for `go-cmp`, `x/mod`, `x/net`, `x/sync`, `x/sys`, `x/telemetry`, `x/tools`; MIT for
`goldmark`), all labelled `lint` / `build-time-only` (D-1.3.9). `TestManifestUpToDate` and
`TestRenderRowShapeIncludesServesAndLabel` both pass against the regenerated file.
**CI wiring required (discovered while validating, not in the original findings list, but
necessary for the fix to actually work in a fresh clone):** `lint`'s new transitive graph includes
modules (`golang.org/x/net`, `golang.org/x/sys`, `golang.org/x/telemetry`,
`github.com/yuin/goldmark`) that are part of `go list -m all`'s graph but are never imported by any
package `lint` builds — Go's module-graph pruning means a plain `go build`/`go vet`/`go test` never
downloads their source, only their `go.mod`. `ScanLicenceGraph`/`licence.ResolveGraph` need every
module's full source (for its `LICENSE` file) to classify the *whole* graph, per AD-26 "Binds:
all". Measured live, with a scratch `GOMODCACHE`: `go build`+`go vet`+`gofmt` alone leave those four
modules undownloaded, and `GOPROXY=off go test ./...` then fails resolving `lint`'s own graph
(`module lookup disabled by GOPROXY=off`); running `go mod download all` first fixes it (re-measured
green). `.github/workflows/ci.yml`'s `lint` job gained a `go mod download all` step, network-on,
immediately before the existing `GOPROXY=off go test ./...` step.
**Related:** AC14, AC17, AC1; D-1.3.5, D-1.3.6(a), D-1.3.11; V10; NFR1.d, AD-1.

### Blocker 2 — AC15's escape-hatch assertion is a tautology
**Decision:** FIX. `lint/internal/rules/maprange_test.go`'s
`TestMapRangeFailureMessageNamesEscapeHatch` now asserts against `const wantIdiom = "for _, k :=
range slices.Sorted(maps.Keys(m)) { v := m[k]; … }"`, written independently in the test file, via
two separate assertions: `EscapeHatch == wantIdiom` and `strings.Contains(f.Message, wantIdiom)`.
**Red-proved, live, this session:** mutated `lint/internal/rules/maprange.go`'s
`EscapeHatch` constant to `"TODO"`; the new test fails with `EscapeHatch constant drifted…`. Reverted
(`diff` confirmed byte-identical); test passes again. Before: red. After (constant restored): green.
**Related:** AC15, AC28; RP-7; D-1.3.5; V10.

### Blocker 3 — DW-1's red-proof lives behind `//go:build matrix` and runs in zero gates
**Decision:** FIX. `checkFixtureShape`, `loadExpectedFixture` and `TestFixtureShapeCheckRedProof`
moved from `folio-go/matrix_test.go` (tagged `matrix`) into `folio-go/fixture_test.go` (untagged —
the same file `isSHA256HexString` already lives in). `matrix_test.go`'s `TestCrossTargetByteIdentity`
and `TestTargetRenderHash` still call `loadExpectedFixture`; same package, different file, no
behaviour change to the tagged tests.
**Measured, live, this session:** `go test ./... -run TestFixtureShapeCheckRedProof` now reports the
test running (with its Finding-12 subtests) in package `folio`, not "no tests to run". `ci.yml:43`'s
existing comment ("go test ./... (includes the two existing guards + DW-1)") is now literally true
and needed no further edit.
**Delivery Log correction (the false attribution Blocker 3 also named):** the "Gates measured in
this story" block's claim that the 25/44 count "added `TestNoFloat64FixtureScan`,
`TestNumberFormattingFixtureScan` **and** `TestFixtureShapeCheckRedProof` as new top-level tests" is
corrected in place below — the count was right, the third name was not actually in that run.
**deferred-work.md:** DW-1 moved from `## Open` to a new `## Done` section, its closure note
rewritten to describe the actual (post-fix) location and gate, per Finding 14 below.
**Related:** AC6, AC28; RP-4; DW-1; D-000.4; D-1.2.5; V10.

### Blocker 4 — AC8's fixture is exempt by filename, not by the narrowing
**Decision:** FIX. Fixture tree restructured exactly per the suggested resolution:
`folio-go/testdata/lint/numeric-formatting/violating.go` → `.../pdf/violating.go`;
`.../template/numbers.go` → `.../template/version.go` (renamed **and** moved out of the
guard's one filename-based exemption). `TestNumberFormattingFixtureScan` is now two subtests: one
scans exactly `pdf/` (the production scope, AC7) and asserts only `violating.go` is reported; the
other scans the fixture tree's **parent** (deliberately widening past `pdf/`) and asserts
`template/version.go` **is** reported there — proving the file's absence in the first subtest is
caused by AC7's scope, not by the `numbers.go` filename carve-out.
**Red-proved, live, this session:** with the restructuring in place, both subtests pass
(`go test ./internal/pdf/... -run TestNumberFormatting -v`); the second subtest is itself the
red-proof mechanism going forward — if a future edit ever widened the production caller's root back
past `internal/pdf/`, `template/version.go` reappearing in that widened scan is exactly what the
second subtest already demonstrates happens.
**Related:** AC8, AC7, AC28; V9; D-1.3.2, D-1.2.5.

### Major 5 — production-scan vacuity guards re-walk with a different function
**Decision:** FIX, across all four named locations. Each scanner now returns its own stats value
alongside `(findings, error)`: `MapRangeStats{DirsVisited, FilesParsed, TypedRangeStmts}`
(`lint/internal/rules/maprange.go`), `ForbiddenImportsStats{DirsVisited, FilesSeen}`
(`lint/internal/rules/forbiddenimports.go`), `noFloat64Stats{packagesVisited, declsSeen}`
(`folio-go/internal/arch_test.go`), `numericFormattingStats{filesChecked}`
(`folio-go/internal/pdf/emit_source_test.go`). Every production-caller vacuity assertion now reads
these stats — computed *inside* the scanner's own walk — instead of re-deriving "files/dirs seen"
with a second, independent walk.
**Red-proved, live, this session (representative case, map-range):** injected `if true { return
nil, MapRangeStats{}, nil }` as `ScanMapRange`'s first statement; `TestMapRangeProductionScan` now
fails (`vacuity guard: scanner's own stats did not report visiting … "geom" and "pdf", got []`).
Reverted, `diff` confirmed clean, suite green again. The other three follow the identical shape
(same stats-from-return-value pattern) and were verified via the full green suite run after the
change; the map-range case is the one this finding says the mutation "is exactly why Finding 1 was
invisible", so it is the one re-proven red-then-green explicitly.
**Related:** AC17, AC1; V2, V10.

### Major 6 — `bannedImportPaths` is exact-match
**Decision:** FIX. `lint/internal/rules/forbiddenimports.go` gained `isBannedImportPath(path
string) bool`, matching `path == banned || strings.HasPrefix(path, banned+"/")`; the `_test.go`
exemption (`testExemptImportPaths`) stays exact-match, unchanged, per the suggested resolution.
New violating fixture: `folio-go/testdata/lint/forbidden-imports/violating_subpackage_imports.go`,
importing `math/rand/v2`, `net/http`, `net/url`, `os/exec`.
**Red-proved, live, this session:** reverted `isBannedImportPath` to `bannedImportPaths[path]`
(exact match only); `TestForbiddenImportsFixtureScan` fails — the subpackage fixture goes
unreported (`expected finding not reported: file=violating_subpackage_imports.go`). Reverted,
`diff` confirmed clean, suite green again. Nothing under `folio-go/internal/` imports any of the
four paths or their subpackages today — `TestForbiddenImportsProductionScan` stays green.
**Related:** AC10, AC11; AD-1; NFR1.c; D-1.3.10.

### Major 7 — five red-proofs were never observed red
**Decision:** FIX. RP-6, RP-7 and RP-14 are now real, executed mutations (see Blocker 2's and Major
6's own red-proofs above for RP-7's replacement and Major 6's for the general shape; RP-14 detailed
below). RP-3b and RP-3c are re-labelled honestly rather than fabricated into false mutations — they
are facts about the shipped test file's own source shape, not something a runtime mutation can
meaningfully redden without mutating the test harness itself into something incoherent. The
red-proof table below (Delivery Log) is corrected in place; its header no longer claims every row
was observed red.
- **RP-6** (map-range "syntactic guess" probe) — **superseded, not re-run.** The withdrawn
  `tolerantImporter` path never contained syntactic guessing to begin with (RP-6 tested a
  *hypothetical* heuristic, never shipped code); the rewritten `ScanMapRange` (Blocker 1) resolves
  types exactly via `go/packages`, so there is no syntactic-guess code path to probe at all. Row
  reworded to state this plainly instead of re-running a probe against nonexistent code.
- **RP-7** — now the real, permanent `TestMapRangeFailureMessageNamesEscapeHatch` (Blocker 2). Red-
  proved live this session (see above).
- **RP-14** — run for real this session: mutated `lint/internal/rules/licencegraph.go`'s
  `ScanLicenceGraph` to `if strings.Contains(moduleDir, "lint") { return nil, nil }` as its first
  statement (the fixture graphs live at `lint/testdata/licence/…`, so this reproduces "a
  copyleft-shaped violation reached through lint" exactly). `TestLicenceGraphFixtureScan/copyleft`
  and `/unknown` both FAIL (`expected finding not reported: … rule=licence`), `/permissive` stays
  green (nothing to report either way). Reverted; `diff` confirmed clean; full suite green again.
- **RP-3b, RP-3c** — reworded in the Delivery Log table (not re-run as mutations): both are
  confirmed-by-reading-the-shipped-source facts, and are now labelled as such rather than implied
  to be mutations.
**Related:** AC28; D-1.2.5, D-1.3.7.

### Major 8 — the three asserted absences key on guessed exact filenames
**Decision:** FIX. `lint/internal/rules/absences.go`'s `absenceChecks` re-keyed on two
**directories** — `folio-designer` and `folio-go/fonts` — instead of three exact filenames. See
`deferred-work.md`'s DW-2 correction note for the full reasoning (the directory-level
`folio-designer/` check is strictly broader than the two filename checks it replaces, since both
guessed artifacts live inside it). `ScanAbsences` now also emits each check's own `c.rule` instead
of a shared generic id — folding in Minor 10 here since both findings touch the same function.
Fixture `want` sets updated to two findings (not three); the existing `violating/` fixture tree
needed no new files, since it already had both directories populated.
**Related:** AC21; V7; D-1.3.4; DW-2. (Folds in Minor 10 for `absences.go` specifically.)

### Minor 9 — `FamilyCommercialEULA` is unreachable
**Decision:** FIX, taking the suggested resolution's second option (delete, don't half-build
detection). `lint/internal/licence/classify.go`: removed the `FamilyCommercialEULA` member and its
`String()` case; a commercial EULA still fails the build by falling through to `FamilyUnknown`
(unchanged behaviour). `lint/internal/rules/licencegraph.go`: removed the dead `case
licence.FamilyCommercialEULA` arm, with a comment explaining why. **Rationale for delete-over-build:**
building real EULA detection (a marker table, SPDX `LicenseRef-`/proprietary conventions, a fixture
stub, a classifier test case) is a real feature addition with its own design surface AD-26 does not
specify — out of this story's scope to invent unreviewed, and the finding's own first option says
so explicitly ("if real EULA detection is ever built"). Deleting the advertised-but-inert member is
the minimal change that resolves the actual defect (a capability that does not exist, advertised as
if it does) without expanding scope.
**Related:** AC18, AC19, AC31; AD-26.

### Minor 10 — `absenceCheck.rule` is dead; `RuleForbiddenImports` covers two shapes
**Decision:** FIX. `absences.go`'s half resolved together with Major 8 above. `forbiddenimports.go`
gained `RuleMathSelector = "math-selector"`, distinct from `RuleForbiddenImports`; the two math-
selector emission sites (call-outside-the-seven, non-call float-valued reference) now use
`RuleMathSelector`; the banned-import-path emission site keeps `RuleForbiddenImports`. Fixture
`want` sets in `forbiddenimports_test.go` updated accordingly (two distinct rule ids now appear).
**Related:** AC1, AC4.

### Minor 11 — AC10's message requirements are asserted nowhere; `math.Pi`'s message omits the surface
**Decision:** FIX. `forbiddenimports.go`'s non-call-reference message gained
`"; AD-1's allow-listed numeric surface: %s"` (was previously missing entirely — the exact gap the
finding named). New test `TestForbiddenImportsMessageContent` in `forbiddenimports_test.go` asserts,
against literals written independently of the production constants (same discipline as Blocker 2's
fix): the banned-import message names the file, the import, and `allowedNumericSurface`; the
math-call message names the file and the surface; the `math.Pi` message names the file, `math.Pi`,
and — the specific clause that was unmet — the surface.
**Related:** AC10, AC12.

### Minor 12 — DW-1's red-proof never exercises `isSHA256HexString`; the second scratch case was never built
**Decision:** FIX, together with Blocker 3 (same test, same move). `TestFixtureShapeCheckRedProof`
in `folio-go/fixture_test.go` now has two subtests: the original per-target-object widening (JSON-
unmarshal failure path) plus a new one covering both a 63-character and an upper-case sha256 (the
`isSHA256HexString` branch specifically). Both subtests pass; real fixture byte-identity re-verified
after both.
**Related:** AC6, AC16; RP-4; DW-1.

### Minor 13 — `ci.yml` declares `env: GO_VERSION` and never uses it
**Decision:** FIX. All three `setup-go@v5` steps in `.github/workflows/ci.yml` now read
`go-version: ${{ env.GO_VERSION }}` instead of the hard-coded literal `"1.26.0"`, three times.
**Related:** AC25; AD-22.

### Minor 14 — DW-1 is marked DONE but filed under `## Open`, with no closing commit reference
**Decision:** FIX. `deferred-work.md` restructured into `## Done` (DW-1 alone) and `## Open`
(DW-2/3/4), with DW-1's `Closed by:` line rewritten to describe the actual fix (moved out of the
`matrix` tag, per Blocker 3) rather than the pre-fix, forward-dated placeholder. The commit SHA
itself cannot be embedded in the commit's own diff (a commit cannot know its own hash before it is
made); the entry instead points at "this file's own commit", which is how the repository's history
resolves it unambiguously. Per Blocker 3's own instruction, DW-1 is marked done here only now that
its red-proof runs in a gate that runs — verified this session, not assumed.
**Related:** AC6, AC22.

### Nit 15 — the spine's `lint/` entry is misaligned by one column
**Decision:** FIX. `ARCHITECTURE-SPINE.md` line 625: removed one space before `#`, now column 39
— measured, matching `hashmatrix/`'s and `lint/`'s own continuation lines. The decision log's
D-000.6 amendment "After" block, which reproduced the same misalignment, corrected identically, with
a note recording the correction (append-only discipline: the block is corrected, not silently
rewritten as if it were always right).
**Related:** AC26; D-000.6.

### Nit 16 — AC13's heading says "six fixtures"; it enumerates eight (now nine)
**Decision:** FIX. Story file's AC13 heading and Task 7's final subtask corrected to "nine
fixtures" (the true post-fix count, since Major 6's fix added a fifth violator), with a note
explaining the count moved twice (six→eight was the original defect; eight→nine is the finisher's
own fixture addition, not a second documentation bug). `TestForbiddenImportsFixtureScan`'s doc
comment in `forbiddenimports_test.go` corrected to match.
**Related:** AC13.

### Nit 17 — repo-root discovery is duplicated three times; the untested copy writes a committed file
**Decision:** FIX, taking the suggested resolution's minimum option ("at minimum, add a test for
`findRepoRoot`") rather than consolidating all three copies. **Rationale for minimum over
consolidation:** the finding itself calls the two `_test.go` duplications "reasonable" and already
explained by comment; consolidating all three into a shared exported helper would touch
`lint/internal/rules/testutil_test.go` and `lint/internal/manifest/manifest_test.go` as well as
`genmanifest/main.go`, for a maintainability nit whose actual defect (zero coverage on the one
untested copy) the minimum fix already closes. New `lint/cmd/genmanifest/main_test.go`:
`TestFindRepoRoot` asserts the resolved root contains both `folio-go/` and `lint/`, and that
`lint/MANIFEST.md` exists under it — closing the specific gap named (the untested copy is the one
that decides where a committed file is written).
**Related:** AC19, AC25.
