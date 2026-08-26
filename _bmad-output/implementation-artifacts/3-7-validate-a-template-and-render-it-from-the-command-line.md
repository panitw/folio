---
baseline_commit: c1c977f
---

# Story 3.7: Validate a template and render it from the command line

**Epic:** 3 — The expression language, aggregation and diagnostics
**Story key:** `3-7-validate-a-template-and-render-it-from-the-command-line`
**Status:** `done`
**Covers:** **FR42** · **AD-7**

**This is the last story of Epic 3.** Epic 3's cross-target boundary matrix run happens immediately
after it, and already owes **three cross-target legs** for Story 3.5's `hidden-image` document
(`byte_neutrality_test.go:549`; Story 3.6, *Could not verify*). Read §"What this story hands the
matrix gate" before scoping.

**Primary invariant — AD-7**, verbatim (`ARCHITECTURE-SPINE.md:194-209`):

> **Binds:** `internal/pdf`, `cmd/folio` · NFR1.e, NFR1.f, FR43, PRD Q11
> **Prevents:** the direct contradiction between NFR1.f (adopt `SOURCE_DATE_EPOCH`) and FR43 (the
> render path reads no environment variable), resolved silently in the wrong direction.
> **Rule:** target **PDF 1.7 (ISO 32000-1)**, resolving Q11. `/ID` is emitted as a content-derived
> value — the first 16 bytes of a SHA-256 over the serialized body up to the point `/ID` is written,
> with both array entries identical. `/CreationDate` and `/ModDate` are **omitted** unless a date
> arrives through `params`. […] The library core never reads an environment variable; `cmd/folio`
> reads `SOURCE_DATE_EPOCH` and passes it in as a parameter like any other caller would.

**Also binding:**

- **AD-1** (`:92-103`) — the determinism boundary is a **directory** boundary. Every package under
  `internal/` is render path and may not import `time`, `os`, `math/rand`, `net`, or a `math`
  transcendental; may not read package-level mutable state; **may not range a map whose order can
  reach an output byte**. The spine's shell table (`:50`): *"Shell — may read the world | `folio`
  (public API), `cmd/folio`, `wasm/`, `folio-designer/`"*.
- **AD-14** — one `Diagnostic` value; codes additive only; a `Warning` accompanies a successful
  render, **never silent and never fatal**.
- **AD-22** — `folio-go/v0.1.0` is not cut (`version.go`: `Version = "0.0.0-dev"`, `git tag` names no
  `folio-go/v*`). Public-surface additions are still free; after the tag they are not.
- **NFR1.f** (`prd.md:370-374`), a **stated requirement**, not a permission: *"The **library core
  reads no environment variable** — FR43 is absolute. The `SOURCE_DATE_EPOCH` convention is honoured
  instead by callers and by Folio's own command-line tooling, which read it and pass the date in as a
  parameter (FR21)."*
- **FR42** (`prd.md:307`): *"Validate a template independently of rendering, so a malformed `.folio`
  file can be rejected in CI before it reaches production."* **Note the words "in CI" — they are the
  reason `--strict` exists (AC5).**

---

## Baseline, measured in this run at creation (`c1c977f`, tree clean)

Run exactly as the gate specifies, in this run, not cited from another story (D-000.26). The shell
filters plain `gofmt`; invoke it as `$(go env GOROOT)/bin/gofmt`.

```
cd folio-go   && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
cd lint       && go test ./... -count=1
cd hashmatrix && go test ./... -count=1
```

| module | result |
|---|---|
| `folio-go` | build ✅ · vet ✅ · gofmt ✅ · **876 pass · 1 fail · 1 skip**, 16 packages |
| `lint` | **111 pass · 0 fail**, 4 packages |
| `hashmatrix` | **3 pass** (orchestrator-verified; not re-run at creation) |

**The one `folio-go` red is REQUIRED and must stay red.** `TestCorpusMeetsP6ExerciseFloors`
(`internal/text/corpus_test.go:189`): *"P6g (opaque names) floor not met: got 7, need >=20"*, stats
**`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`** — byte-identical to Stories 3.5 and 3.6.
D-000.17 / D-2.1.14 / D-000.57 mandate it stay unmet. **Any change to these stats is a finding**, not
a success.

**`lint` is fully green (111/0).** DW-19 was discharged at 3.6. **Any `lint` red in this story is a
regression this story caused** — there is no inherited noise left to attribute it to. Per **D-000.70**,
the one standing red that remains *"deserves a deliberate look at what sits behind it in its own
package"* — not to fix it, but to establish whether anything downstream is unreached. That look
belongs to the Epic 3 gate, not to this story; it is recorded here because 3.7 is the last story
before it.

---

## FOUR divergences between the record and the shipped tree — ALL FOUR CONFIRMED by the lead

**Where the record and the code disagree, the code is the truth.**

### D-1 (CONFIRMED) — `folio-go/cmd/` ALREADY EXISTS, with three tenants. The "must not be created early" record is superseded; the naming obligation survives and is sharper than the record states.

The decision log carries *"`folio-go/cmd/folio` belongs to Story 3.7 and must not be created early"*
(Story 1.7's tripwire, D-1.7.7 item 2). **At HEAD `folio-go/cmd/` holds three packages**:
`genbreaks/`, `gencorpus/`, `gentrie/` — Story 2.1's build-time tooling.

Not a violation. **D-2.1.5 re-keyed the tripwire away from the path**, and DW-10 records why in the
log's own words: *"That key was **broader than the purpose**: `cmd/` has more than one legitimate
tenant, and Story 2.1's own build-time tooling … tripped it despite having nothing to do with AD-7 or
params-date wiring — a measured false positive."* The tripwire is now `absence-source-date-epoch`,
keyed on the **literal string** `SOURCE_DATE_EPOCH` in any `.go` file under `folio-go/` (excluding
`testdata/`).

**What survives is the load-bearing half, and the lead sharpened it.** The binary **MUST** be named
`cmd/folio`. `cmd/folio` is named in **AD-7's `Binds` line**, in **the spine's source tree**
(`:626 cmd/folio/ # CLI: validate, render`) **and in NFR1.f**. So naming it `cmd/foliocli` would make
**three canonical documents name a nonexistent artifact, with nothing catching it** — the content
tripwire would still fire on the env-var literal and every other AC in this story would pass. **The
path is a proxy for the obligation, not the obligation.** AC3 asserts the name for exactly this
reason.

### D-2 (CONFIRMED — and the LOG IS A DEFECT TO BE CORRECTED, not merely worked around)

The decision log says twice (`:149`, `:1524`) that `TestNumberFormattingIsConfinedToNumbersGo` *"fires
hard at Story 3.6 (`internal/diag`…) and **Story 3.7** (`cmd/folio`, a CLI that must print)."*

**Structurally inapplicable at HEAD, by line number.** `folio-go/internal/pdf/emit_source_test.go`
declares the scanner's scope in its own doc comment: *"narrowed per D-1.3.2 to a single rule: within
the given target directory (production: **exactly folio-go/internal/pdf/**…) […] **D-1.3.2 deleted the
second, module-wide half outright**."* Its production caller is
`scanNumericFormatting(filepath.Join(root, "folio-go", "internal", "pdf"))`. `cmd/folio` is outside
that root. **A CLI in `cmd/folio` may call `fmt.Printf`, `fmt.Fprintln` and `strconv.Itoa` freely, and
no guard in this repo objects.**

**The lead's ruling: both sites are DEFECTS IN THE RECORD (D-000.49), not merely stale, and this story
corrects them** — see task 16. Two live predictions of a red that cannot occur are exactly the thing a
future creator scopes work around. And the instinct to cite line numbers rather than paraphrase is
**D-000.46**: had this story merely restated the log, the remedy a developer would reach for is
*"widen the guard to match the comment"* — which the log itself already names as *"the wrong direction
relative to the ruling."*

### D-3 (CONFIRMED — the lead's own grounding was one dimension short) — D-000.67 part 1 fires TWICE, at 3.7 and at 5.1, and only the later was recorded. Measured.

`lint/internal/rules/absences.go` carries a forward note recording *"this list started at 3 entries,
Story 3.6 drops it to 2, **Story 3.7 drops it to 1**, and Story 5.1 must drop it to 0 by REMOVING
`ScanAbsences` and its precondition TOGETHER"* — because `TestAbsencesZeroWitnessIsCaught` proves an
empty `absenceChecks` yields `ChecksEvaluated == 0` and **passes green**. Correct, and unchanged by
this story.

**That note tracks `ChecksEvaluated` and never looked at `ContentFilesScanned`.**
`absence-source-date-epoch` is the **only** `absenceKindContent` row. `AbsencesStats.ContentFilesScanned`
(Story 2.1, Finding 13) counts `.go` files opened by content-kind rows across all rows, and
`TestAbsencesProductionScan` **fatals** on it being zero. Removing the row drives it to zero — **here,
at 3.7, an epic before the recorded collision.**

**Measured in this run.** A package-local test double substituting `absenceChecks` with the
post-story remainder, `t.Cleanup` restoring it — the same shape `TestAbsencesZeroWitnessIsCaught`
already uses, so it is **not** a reverted production mutation (D-000.9 obligation 2's forbidden
shape). Probe deleted afterwards; `git status` clean.

```
findings=[] stats={ChecksEvaluated:1 ContentFilesScanned:0}
CONFIRMED: ContentFilesScanned == 0 -> TestAbsencesProductionScan's vacuity guard would FAIL
```

`TestAbsencesFixtureScan`'s `compliant` subtest carries the **same** precondition
(`absences_test.go:172`).

**The transferable form, ruled by the lead and recorded canonically as [[D-000.67]] (amendment) —
*"The rule fires TWICE, and a mechanism can carry more than one presence precondition"*:**

> **A mechanism can carry more than one presence precondition, each keyed on a different population.
> Check every witness the mechanism reports, not the one the roadmap made you think of.**

Disposition confirmed: **AC13** removes the `absenceKindContent` mechanism **together with** both
preconditions and its two fixtures. **Removal, never relaxation.** And explicitly: **the kind must not
be kept "for later"** — an empty check kind is precisely the zero-candidate scan D-000.67 exists to
prevent, and no tenant is scheduled.

### D-4 (CONFIRMED) — DW-10's entry names a test that no longer exists.

`deferred-work.md`'s DW-10 says *"`TestAbsencesChecksIncludeAllFourEntries` still pins the rule ids by
name."* At HEAD it is **`TestAbsencesChecksIncludeBothRemainingEntries`** — renamed twice as the list
shrank 5 → 4 → 3 → 2. Cosmetic, but exactly the stale-name class **D-000.37** requires be corrected in
the same commit that touches the row. This story renames it a third time and fixes DW-10's prose.

**Verified as stated, not divergent:** `absenceChecks` holds exactly **2** entries at HEAD; the 2 → 1
step is correct. `folio.Validate` does **not** exist. `internal/diag`, `RenderError`, the renumbered
`Severity` and the missing-glyph Warning are all present exactly as 3.6's record describes.
**`documentDate` has ZERO hits** anywhere in the tree (`.go`, `.json`, `.md`, `.folio`) — the spelling
DECISION-2 rules is clear (see task 4's standing instruction anyway).

---

## In plain terms (read this first if you just want the gist)

Folio is a library today; you can only use it by writing Go. This story gives it a command-line tool
with the two jobs a build pipeline actually needs: check a template, and produce a document from one.

Checking comes first, because the point of checking is to fail early, before anything tries to draw a
page. The checker never renders as a side effect, and it agrees with the real thing: anything the
renderer would refuse, the checker refuses too, or people learn to distrust it and stop running it. One
honest caveat: the checker predicts *for the inputs you hand it*, so no sample data means a true report
that data is missing, not a fault in the template.

The tool also prints. Folio already notices problems that do not stop a document — a character none of
the chosen typefaces can draw is simply left out, with nothing on the page to show it happened. The
only record is a warning, so a quiet tool would turn a visible defect into an invisible one. The tool
prints every warning it receives, and offers a switch that turns any warning into a failed build, off
by default, because treating an ordinary trimmed character as fatal would stop releases that are
actually fine.

Then dates. The library never asks the machine what time it is, because the same inputs must produce
the same file everywhere. The command-line tool may look up a build system's own announced timestamp
and hand it to the library as an ordinary input, exactly as any other caller would. Supply nothing, by
any route, and the document carries no date fields at all — not blank, not zero, absent.

**A rule the first pass missed, corrected here:** the build-system timestamp is a fallback, never an
override. If you have already told the tool what date to use, the environment's own guess must not
quietly replace your choice — it only fills in when nothing else did.

**What this pass found and closed.** A second, adversarial look at the finished work checked every
guard against what it actually catches, not what it claims to. It found a purity check that only
looked at the front of the output and missed a corruption stitched onto the back; two internal safety
nets with a gap a later, deliberate change could still slip through; a coalescing guard whose
real protection came from a mechanism nobody had named; and a scattering of stale cross-references and
one uncovered date value outside the calendar's own stated bounds. Every one of those is now closed,
each demonstrated live against the exact defect it exists to catch, not merely asserted fixed.

**What will look wrong later and is not.** One test in this project is *required* to fail — a coverage
target for Thai text deliberately left unmet, failing identically for several stories running. Do not
fix it. Separately, one very old, low-likelihood date-format edge case — a leap second — is knowingly
left for a future pass, because fixing it properly needs its own separate look at the calendar
arithmetic underneath, not a quick patch alongside everything else here.

---

## Story

As an integrating Go developer,
I want to reject a malformed template in CI before it reaches production, and to pin document dates
reproducibly,
So that a bad template fails my pipeline rather than my customers' statements.

### Source Acceptance Criteria (`epics.md:1130-1157`, verbatim)

> **Given** a malformed `.folio` file
> **When** `folio.Validate` is called
> **Then** it is rejected with located diagnostics and no render is attempted
>
> **Given** `cmd/folio`
> **When** it is run
> **Then** it offers a validate command and a render command
>
> **Given** `SOURCE_DATE_EPOCH` set in the environment
> **When** `cmd/folio render` runs
> **Then** the CLI reads it and passes it in as a parameter
> **And** the library core still reads no environment variable
>
> **Given** no date supplied by any route
> **When** the PDF is produced
> **Then** `/CreationDate` and `/ModDate` are omitted

---

## Rulings received at creation — D-3.7.1 … D-3.7.5, plus [[D-000.67]] (amendment)

The five open questions this story raised at creation are **settled**, and are recorded canonically in
`folio-mvp-decision-log.md` as **D-3.7.1 … D-3.7.5**; the divergence D-3 measurement produced a sixth
entry, **[[D-000.67]] (amendment)**. They are summarised here with their grounds because the grounds
determine the shape of the tests, not just the code — **but the log is the authority, and per D-1.1.b's
standing lesson this story cites decision IDs rather than substituting a paraphrase for them.** Read
the entries themselves before implementing.

**Do not re-open them.** One creator recommendation was **overruled**; read **D-3.7.3** carefully.
**D-3.7.4 adds a deliverable the source AC does not name** (`--strict`, AC5).

### D-3.7.1 — `Validate(b []byte, d Data, p Params, f FontSet) ([]Diagnostic, error)`. CONFIRMED and FORCED.

Bytes are forced by the subject (a *malformed* file cannot become a `*Template`) and by **D-1.4.6**'s
precedent — `ParseTemplate([]byte)`, with `os` kept at the package boundary. `([]Diagnostic, error)`
matches `RenderTo` under **D-2.8.6**: there are no bytes, so there is no `Result`.

**One guardrail, and it closes a real usability trap.** `Validate` predicts `Render` **for the inputs
given**. A caller passing empty `Data` gets absent-path Errors that are **correct predictions of a
render with empty data**, not template defects. **Document that in those words, and pin it with a
test** — otherwise CI users will read the first as the second and conclude the validator is broken.

**Do NOT add a second, structural-only entry point.** Two public entry points to cover one FR is the
surface growth **D-1.1.c** and counter-metric **C1** exist to prevent, and the doc sentence costs
nothing.

### D-3.7.2 — the reserved params key is **`documentDate`**, an **RFC 3339 string**, setting **BOTH** date keys.

- **Key:** a reserved top-level params key **`documentDate`**. `reportDate` is deliberately **not**
  used: Story 6.3's AC already spends it as the author's own example (`epics.md:58`, `:629`, `:1777`).
- **Value:** an **RFC 3339 timestamp string**, parsed by `internal/expr`'s already-shipped
  `rfc3339Pattern` (`formatdate.go:17`) and its integer calendar (`calendar.go`) — **so no `time`
  enters `internal/`.**
- **Both keys:** `documentDate` sets `/CreationDate` **and** `/ModDate`. One value, two keys.
- **Conversion:** epoch → RFC 3339 happens in **`cmd/folio`**, using `time`, which is legal there and
  nowhere else under `folio-go/`.
- **A present-but-non-RFC-3339 value is a located Error**, consistent with Story 3.4's `formatDate`
  rule — and exactly the class D-3.7.1's four-argument `Validate` exists to catch before production.

**Why the namespace-collision worry dissolves:** sharing the key with `{{params.documentDate}}` is
**benign, not a hijack**. If an author prints it in a footer, the value in the metadata **is** the
value on the page. That is what anyone would want.

**Confidence: HIGH on the shape, MEDIUM on the spelling.** **Standing instruction to the developer:
grep for an existing use of `documentDate` before implementing, and come back if it hits.** Run at
creation: **zero hits** across `.go`, `.json`, `.md` and `.folio` in the whole tree. Re-run it anyway
— the tree moves.

**One line for DW-4's ledger:** `documentDate` is **public contract**, frozen at `folio-go/v0.1.0`
alongside the API signatures. **The params namespace now has a reserved name in it.**

### D-3.7.3 — OVERRULED. Coalesce missing-glyph Warnings in the **ENGINE**, not at presentation.

The creator recommended coalescing in the CLI's presentation layer. **Declined.**

**The reason is DW-7's shape.** Presentation coalescing has **three** implementers — 3.7's CLI, 5.12's
interface, 6.6's failure card — and *three implementations of one rule* is precisely the drift hazard
DW-7 exists to name. Recording the precedent in DW-17 helps, but **a precedent is a wish; one
implementation is a fact.**

**And it costs less than the creator assumed, because the count is decoration.** The actionable unit is
**(element, distinct rune)** — you fix the font chain, not the 400th occurrence. So: **no count field,
no public type changes**, fewer `Diagnostic` values of the same shape. **No golden moves** (diagnostics
are not in the PDF bytes) and **no existing consumer to break**, because 3.6 shipped this Warning last
story.

**The moment is the argument.** Now, before any presenter exists, it is a small change at
`render.go:1046-1052`. After three presenters have each coalesced, it is four places and three
conventions to reconcile.

**The guardrail that would otherwise bite.** De-duplication must **NOT use map iteration** — AD-1
forbids it where order can reach an output, and **D-2.8.6** made the diagnostics slice's **order a
determinism guarantee**. The distinct-rune population per element is tiny, so **a slice with a linear
scan** is correct and obviously ordered. **First-occurrence position determines order. Assert the
ordering, not just the count.**

Record it in **DW-17** as a **BEHAVIOUR precedent, not a presentation convention**, so 5.12 and 6.6
inherit a property they can rely on rather than a rule each must re-implement.

### D-3.7.4 — exit codes and stream discipline CONFIRMED, **plus a new deliverable: `--strict`.**

**A Warning exits 0.** AD-14 is explicit that a Warning *"accompanies a successful render, never
fatal"*; making warnings fatal by default would make FR44's clip fatal, which defeats its entire point.

**But exit-0-only is half a CLI, and the lead owned why.** FR42's stated purpose is *"reject a
malformed template in CI before production."* **D-3.6.8** ruled an uncovered rune is **omitted with no
in-band page signal** — so under warnings-never-fail, **a customer's name can silently lose a character
and no pipeline can ever catch it.** In the lead's words: *"I made the page silent; I owe a route by
which the silence can fail a build."*

> **Ship `--strict`: any Warning exits non-zero. Off by default** (AD-14 binds the default). One flag,
> standard, and **the only mechanism that makes the missing-glyph Warning actionable in CI.** This is
> not scope growth — it is the difference between the CLI serving FR42's use case and merely printing.

**Exit codes:** `0` success (warnings may have been printed); `1` validation or render failure, and
`--strict` with warnings; `2` **usage** error. Exit 2 keeps *"your template is bad"* distinguishable
from *"you called me wrong"* when a CI failure is being triaged.

**Streams**, with the addition that makes piping safe: diagnostics and errors to **stderr, always**;
PDF bytes to **stdout only when no output path is given**, otherwise to the file; and **when bytes go
to stdout, NOTHING else may** — no progress, no summary, no banner. **This must be stated explicitly in
the code and asserted**, because a later "helpful" summary line on stdout silently corrupts
`folio render t.folio > out.pdf`.

**DW-17's first discharge lands here:** the CLI must print the diagnostics it receives, **asserted on
stdout/stderr content** (D-000.21 — the property belongs to the artifact that carries it), never on the
returned slice.

### D-3.7.5 — no matrix document. **The GROUNDS are the criterion, not the cost.**

Declined, consistent with **D-3.5.6**: a `/Info` dictionary assembled from an author-supplied RFC 3339
parameter through the existing `numbers.go` idiom introduces **no new source of cross-target
divergence**, and the CLI's `time` use is in the shell.

**A correction the lead made to its own framing, reflected here deliberately.** The creator's draft
offered *"four legs on a gate already owing three"* as a supporting reason. That is a **true cost and
must NOT be part of the grounds.** **D-000.4**'s own warning is that a weaker trigger erodes the trade
to nothing; **the mirror is that declining on expense erodes it from the other side, until the
criterion becomes whatever the gate can afford.** **Decline because it does not meet the criterion.**
The cost is a reason to be glad, not a reason to decide.

**Two requirements attached — AC14 carries both.**

1. *"When no date is supplied nothing is emitted and bytes are unchanged"* must be a **measured
   byte-identity result over the existing corpus, with the command recorded** — not an expectation.
   Same condition as 3.5's `hidden-image` count.
2. The **with-date** path then has **no cross-target coverage at all**, and that must be stated as a
   **construction argument, not a measurement**: the input is an identical parameter on every target
   and the assembly is integer/ASCII, so the bytes cannot differ. This is **D-000.24**'s labelled
   category used correctly — **a forward property with its reason, never "we checked."**

---

## Do not re-open — settled rulings this story inherits

1. **`Validate` is a dry-run predictor of `Render`, not a second rule system** (decision log `:4189`,
   *binding*, verbatim): *"An unrenderable asset nothing draws → `Render` succeeds → `Validate` is
   clean. One that is drawn → `Render` errors → `Validate` errors, in CI, before production."* Two
   consequences also ruled: **a hidden element never triggers it** (AD-24 — a hidden element is
   *absent* from the `PageModel`, so nothing draws it); and **`Validate` without data cannot always
   predict** a `visibleIf`-gated failure — *"where data-dependence makes prediction impossible the
   diagnostic degrades to a **Warning**"*, labelled a **"Story 3.7 detail."** The ruling ends: **"This
   will matter at Story 3.7."** It does.
2. **A recognised `mediaType` whose bytes do not parse as that format IS a load error** (same ruling,
   *binding*) — reader-independent, cheap via magic bytes. Not relaxed.
3. **Missing glyph is a `Warning`, and the rune is OMITTED** — no glyph, no advance, no `.notdef`, no
   substitute (D-3.6.8). Not reopened. D-3.7.3 changes only the **granularity** of the Warning, never its
   severity, its code, or what the render emits.
4. **`RenderTo` returns `([]Diagnostic, error)`, not `Result`** (D-2.8.6). `Render` returns `Result`.
   Do not harmonise them.
5. **`RenderTo` cannot stream** (D-1.7.2) — `/ID` is content-derived over the body and the xref records
   byte offsets. The CLI writes a complete buffer. Do not "optimise" this.
6. **The stage-rank table's ranks are ratified as measured** (D-000.16, re-measured at Story 2.5).
   `cmd/folio` is **not** under `internal/`, so `ScanStageRank` does not rank it and none is added.
7. **Epic 3's boundary matrix run is due AFTER this story** (D-000.4). **This story does not run the
   matrix.**
8. **`TestRenderWithNoDateInParamsOmitsCreationAndModDate` was re-scoped at 1.7 precisely so 3.7's
   developer would not meet a red test whose cheapest fix is to weaken it** (D-1.7.7). It **must
   survive unweakened**, with all three cases and its full forbidden list, and it is the standing guard
   for AC11.

---

## R — design constraints derived from the record and the tree during creation

**R1 — DW-10's obligation is THREE parts and the third one's red-proof window closes when it is wired.**
D-000.59's table (decision log `:9119`), verbatim:

> (a) `cmd/folio render` reads `SOURCE_DATE_EPOCH` and passes it **as a parameter**, asserted **through
> the params path**, not by the env var appearing in source; (b) the library core still reads no
> environment variable — discharged by the **existing** AD-1 import lint, **cited by symbol, not
> re-implemented** (D-000.42: no redundant third guard); (c) with no date supplied by any route,
> `/CreationDate` and `/ModDate` are **absent from the produced PDF bytes**, read off the artifact
> (D-000.21), never off the params struct.
>
> **The ordering trap, which must appear in all three story prompts:** obligation (3)'s red-proof must
> be captured **BEFORE** the obligation is wired, because wiring it closes the window permanently
> (D-000.30).

**R2 — the symbol part (b) must cite is `TestRenderEntryFileHasNoForbiddenImports`, and its scope is
narrower than "the library core". Measured, and it is an ANCHOR problem.**
`lint/internal/rules/forbiddenimports_test.go:264`. Its target is **not** a package:
`findRenderDeclaringFiles` locates, **by AST**, every non-test file directly under `folio-go/` that
declares a top-level `Render` or `RenderTo`, and scans exactly those. Today that is one file,
`render_entry.go`.

**Injection probe run at creation** (added `os` + `time` imports and an
`os.Getenv("SOURCE_DATE_EPOCH")` call to `render_entry.go`, ran `lint`, reverted; `git status` clean):

```
--- FAIL: TestRenderEntryFileHasNoForbiddenImports
    AC12: the file declaring Render must import none of os/time/net/math/rand, found:
        render_entry.go:39: forbidden import "os"
        render_entry.go:40: forbidden import "time"
--- FAIL: TestAbsencesProductionScan
    [{folio-go/render_entry.go absence-source-date-epoch … contains SOURCE_DATE_EPOCH …}]
```

The guard is real and is the right symbol to cite. **But it is D-000.68's anchor problem again, and the
lead named it as such:** locating by *"files declaring `Render`/`RenderTo`"* is a name set **the code
can move**, by adding an entry point. Coverage at HEAD:

- `folio-go/internal/**` — covered by `TestForbiddenImportsProductionScan`. Solid.
- `folio-go/render_entry.go` — covered, because it declares `Render`/`RenderTo`.
- **`folio-go/folio.go` — NOT covered, and legitimately imports `os`** for `LoadTemplate`.
- **A new file declaring `Validate` — NOT covered by anything.** The genuinely new unguarded surface.

**Name the instrument shape, because this is its second appearance and it should become a habit rather
than a coincidence:** AC12's `folio.go` control — an injection that must **stay GREEN** — is **the same
instrument as Story 3.6's add-with-pin mutation**: a proof that the guard **discriminates the case it
must not fire on.** A widening mutation without its non-firing control has proved half the property.

**R3 — `fonts.Shipped()` already exists and is the CLI's font source.** `folio-go/fonts/fonts.go`
returns a `folio.FontSet` of the three shipped faces, and records the load-bearing direction: *"`folio`
must NOT import `folio/fonts` … reversing it would force every consumer of package folio to pull in
~11.3 MB of embedded font data."* `cmd/folio` importing **both** is the intended shape — the shell
composes; the library does not. Do not add a font-loading path to package `folio`.

**R4 — the date value must ride into `internal/pdf` as a VALUE, never as an import.** AD-1's spine rule:
*"Stages communicate by what they pass, not by what they import."* The chain is
`Render → renderDocument (render.go:1220) → pdf.SerializeTextDocument (internal/pdf/textdoc.go:108)`,
whose signature is `(pages []pagemodel.Page, faces map[string]EmbeddedFace, images map[string]ImageXObject)`
and whose doc comment ends *"no compression, no /Info dictionary, no /CreationDate or /ModDate"* — about
to become false, and to be corrected in the same commit (D-000.37).

**R5 — the civil-calendar arithmetic already exists and is pure integer.**
`folio-go/internal/expr/calendar.go` holds `daysFromCivil`/`civilFromDays`, bounded to years 1–9999,
with `import "time"` appearing nowhere in the package; `rfc3339Pattern` and `parseUTCOffsetMinutes` are
in `formatdate.go`. `internal/expr` is **rank 3**, `internal/pdf` is **rank 8**, so `pdf → expr` is a
legal arrow under the stage-rank table. **Do not write a second calendar or a second RFC 3339 parser**
(D-000.42).

**R6 — `internal/pdf` may NOT use `fmt` or `strconv` outside `numbers.go`.** This one *is* live (unlike
D-2). `TestNumberFormattingIsConfinedToNumbersGo` scans exactly `folio-go/internal/pdf/`, every file
except `numbers.go`, `_test.go` included. A PDF date `D:YYYYMMDDHHmmSS+HH'mm'` is zero-padded-integer
assembly; `numbers.go` already exports the `appendInt`/`appendIntPadded` idiom. Use it. A
`fmt.Sprintf("%04d…")` anywhere in `internal/pdf/` outside `numbers.go` is an immediate build red.

**R7 — DW-17's presented-output half is this story's, and 3.6 raised its weight.** DW-17 as amended:
*"the uncovered rune is OMITTED — no glyph, no advance, and no in-band marker of any kind. […] If
DW-17's three owners ship without surfacing this specific Warning, the defect is not merely unreported,
it is INVISIBLE."* And: *"each owes an assertion on ITS OWN presented output once built."*

**R8 — no committed corpus fixture produces a missing-glyph Warning.**
`TestCorpusFixturesProduceNoMissingGlyphWarnings` guarantees it, and DW-17's reversal-cost argument
relies on that guarantee. So AC6 and AC7 need a **synthetic** document (a chain covering none of the
text's script), constructed by the test. Not added to `fixtures/`, and it must not perturb that corpus
assertion.

**R9 — `Diagnostics` nil-vs-empty is a determinism guarantee.** `diagnostic.go:276-290`: *"no render
path that emits no `Diagnostic` ever sets `Diagnostics` to a non-nil empty slice."* A CLI summary line
must derive from `len(...)` and must not construct an empty slice on the way.

**R10 — the module graph must stay at exactly two modules.**
`TestModuleGraphHasNoThirdPartyDependencies` and `lint/MANIFEST.md` (regenerated by
`cd lint && go run ./cmd/genmanifest`, asserted by `TestManifestUpToDate`). **`cmd/folio` uses the
standard library only** — no flag-parsing dependency, no CLI framework. `flag` from the standard
library is fine. If the manifest's rows change, regenerate and commit; never hand-edit.

---

## Acceptance Criteria

Each AC names its **anchor** (D-000.68 — the three that hold are **the compiler, the type system, and a
literal the test owns**; pin to a literal when the set is permanent, state it relationally when the
roadmap moves it) and its **discriminating mutation**. **An AC that passes before implementation has
been mis-written — fix the test, not the code.** Per **D-3.6.9**'s producer discriminator, a test
comparing a **computed value** against an expectation is **not** tautological; a test comparing a
declaration against itself is.

### Validation

**AC1 — `folio.Validate(b []byte, d Data, p Params, f FontSet) ([]Diagnostic, error)` rejects a
malformed `.folio` with a located, coded diagnostic, and NO RENDER IS ATTEMPTED.** (D-3.7.1.)
Malformed input returns a non-nil error that `errors.As` resolves to `*folio.RenderError`, whose
`Diagnostic` carries `SeverityError`, a registry code, and — where the failure mode has one — an
`ElementID` (AD-10). *"No render is attempted"* is asserted as a **structural** property, never
inferred from the absence of bytes: nothing on `Validate`'s call graph reaches `renderDocument`,
`buildPageModel` or `internal/pdf`, asserted by AST over package `folio`'s own top-level functions.
**Reuse `render_arch_test.go`'s existing `buildFolioCallGraph`** (`:59`), which already computes this
graph and already tracks *"body calls into `internal/pdf` at all (any selector)"*. Do not write a
second call-graph builder (D-000.42).
*Anchor:* the call graph, built by AST from the compiler's view of the package.
*Mutation:* insert a discarded call to `renderDocument` inside `Validate`'s body → the AST assertion
must redden. **Run it.** Note the trap: an assertion phrased as *"`Validate` returned no bytes"* passes
on a `Validate` that renders and throws the bytes away — the exact defect this AC forbids.

**AC2 — `Validate` predicts `Render`, the prediction is instrumented, and the empty-`Data` caveat is
documented AND pinned.** (§Do not re-open item 1; D-3.7.1's guardrail.)
Assert the prediction **relationally, over a table the test owns**: for each document, `Validate`'s
verdict agrees with `Render`'s on the same inputs. At minimum:
- a malformed document (both reject);
- a well-formed document with data and fonts supplied (both clean);
- **D-3.5.1's case**, named in the ruling by number — a literal `visibleIf` that *"is a guaranteed
  render-time error that currently loads clean"*. **Verify at HEAD what `ParseTemplate` and `Render`
  each actually do with it before writing the expectation**, and state the measured truth;
- a `visibleIf`-gated failure whose reachability is **data-dependent**, where the ruling says
  prediction is impossible and *"the diagnostic degrades to a Warning"* — assert the **Warning**, not
  an Error and not a clean pass;
- **D-3.7.1's trap, pinned**: the same well-formed document validated with **empty `Data`** yields
  absent-path **Errors**, and the test asserts they are **correct predictions of a render with empty
  data**, by asserting `Render` with the same empty `Data` produces the matching failure. `Validate`'s
  doc comment says this **in those words**.
*Anchor:* a literal the test owns (the case table and expected verdicts) judged against **two
independent producers**, `Validate` and `Render` — D-3.6.9's discriminator: two things compute, one
expectation judges.
*Mutation:* remove one validation step from `Validate` that `Render` still performs → the agreement
table reddens on that row. **Run it on a real step, not an invented one.**

### The command-line tool

**AC3 — the binary is `cmd/folio`, and both subcommands exist.**
`folio-go/cmd/folio/main.go` declares `package main`. Run with no arguments, or with an unknown
subcommand, it names **both** `validate` and `render`. Both are reachable and work end-to-end.
*Anchor:* the filesystem path plus a literal the test owns (`"validate"`, `"render"` — a permanently
closed two-element set for the MVP, so D-000.68's "pin to a literal" arm applies).
*Mutation:* rename the directory to `cmd/foliocli` → the path assertion reddens. **This AC carries
D-1's obligation**: without it a rename satisfies every other AC here while AD-7's `Binds`, the spine's
source tree and NFR1.f all silently name a nonexistent artifact.

**AC4 — exit codes and stream discipline.** (D-3.7.4.)
`0` success (warnings may have been printed); `1` validation or render failure; `2` **usage** error
(unknown subcommand, missing argument) — so *"your template is bad"* stays distinguishable from *"you
called me wrong"*. Diagnostics and errors to **stderr, always**. PDF bytes to **stdout only when no
output path is given**, otherwise to the named file. **When bytes go to stdout, NOTHING else may** — no
progress line, no summary, no banner.
*Anchor:* the captured streams and the process exit status — both artifacts, both read off the run.
*Mutation:* emit a one-line summary on stdout alongside the bytes → the assertion must redden.
**Write that mutation as a permanent test, not a one-off**: a later "helpful" summary line is exactly
how `folio render t.folio > out.pdf` gets silently corrupted, and the guard is the only thing standing
between here and there.

**AC5 — `--strict` makes any Warning exit non-zero, and it is OFF by default.** (D-3.7.4, new deliverable.)
Without the flag, a render producing Warnings exits **0** and still writes complete bytes — AD-14:
a Warning *"accompanies a successful render, never fatal"*. With `--strict`, the same run exits
non-zero. **Both arms asserted on the same document**, so the flag is the only variable.
**The missing-glyph document from AC6 must be one of the asserted cases** — `--strict` is *"the only
mechanism that makes the missing-glyph Warning actionable in CI"*, which is why this AC exists at all.
*Anchor:* the process exit status, plus a literal the test owns (the flag spelling).
*Mutation:* make the flag default to on → the default-arm assertion reddens, and it must, because a
default-fatal Warning would make FR44's clip fatal and defeat its purpose.

### Diagnostics reaching a human

**AC6 — the CLI PRINTS the diagnostics it receives, asserted on ITS OWN streams.** (DW-17, R7, D-3.7.4.)
A render returning a non-empty `Diagnostics` produces, on **stderr**, text a human can act on: the
severity, the stable code, the element id where present, and the message. Asserted by running the
command and reading the captured stream — **never** by asserting `Result.Diagnostics` was non-empty,
which is the input side (D-000.21).
**The case DW-17's amendment sharpens must be among them**: a document whose declared chain covers none
of some rune's script, where *"the Warning is the ONLY record the rune ever existed"* and there is **no
in-band page signal at all**. Per R8 the corpus provides none, so the test constructs one, outside
`fixtures/`, without perturbing `TestCorpusFixturesProduceNoMissingGlyphWarnings`.
*Anchor:* the captured stream, plus a literal the test owns (the code string and a substring of the
rune's `U+XXXX` form).
*Mutation:* drop the print loop → the assertion reddens. **Also assert the negative control:** a clean
render prints no diagnostic text, so *"the CLI printed something"* is never mistaken for *"the CLI
printed this."*

**AC7 — the missing-glyph Warning is coalesced IN THE ENGINE to one per (element, distinct rune), in
first-occurrence order.** (D-3.7.3, OVERRULING the creator's presentation-layer recommendation.)
At `render.go:1046-1052`, inside `for _, r := range elementText`, one `Diagnostic` is currently appended
per uncovered **occurrence**. It becomes one per **distinct rune per element**. **No count field. No
public type changes.** Fewer `Diagnostic` values of the same shape.
**De-duplication must NOT use map iteration** — AD-1 forbids it where order can reach an output, and
D-2.8.6 made the diagnostics slice's **order a determinism guarantee**. Use a **slice with a linear
scan**; the distinct-rune population per element is tiny. **First-occurrence position determines
order.**
*Anchor:* a literal the test owns — the expected **ordered sequence** of codes and runes for a
constructed document with a known interleaving of covered and uncovered runes.
*Mutation two ways, both required:* (a) a 3-occurrence string of one uncovered rune must yield **1**
Diagnostic, not 3 — reverting the coalescing reddens it; (b) **swap two entries in the emitted order**
→ the ordering assertion must redden. **Assert the ordering, not just the count** — a count-only
assertion is satisfied by a map-ranged implementation, which is the AD-1 violation this guardrail
exists to prevent.
*Also:* no golden moves (diagnostics are not in the PDF bytes) and no existing consumer breaks (3.6
shipped this Warning last story). Confirm both rather than assuming them.

### Dates

**AC8 — `cmd/folio render` reads `SOURCE_DATE_EPOCH`, passes it in as the `documentDate` PARAMETER, and
the date reaches `/CreationDate` AND `/ModDate` in the produced bytes.** (D-3.7.2, D-000.59 part (a).)
Asserted **through the params path**, never by the env var appearing in source: a subprocess-style test
sets `SOURCE_DATE_EPOCH` in the child's environment, runs `render`, and reads **both** keys **off the
produced PDF bytes** (D-000.21). The emitted value corresponds to the epoch supplied — **assert the
actual formatted `D:YYYYMMDDHHmmSSOHH'mm'` string**, not merely that the keys are present. Epoch →
RFC 3339 happens in the CLI with `time`; RFC 3339 → PDF date happens in the engine with `expr`'s
shipped parser and integer calendar (R5), through `numbers.go`'s idiom (R6).
*Anchor:* the produced artifact plus a literal the test owns (the expected date string for a fixed
epoch).
*Mutation:* have the CLI pass a hard-coded epoch instead of the one it read → reddens. A test asserting
only that `/CreationDate` is *present* passes on a CLI that ignores the environment entirely — that is
D-000.21's whole subject.

**AC9 — the same date arrives when a caller supplies `documentDate` directly through `Params`, with no
environment involved.**
`folio.Render` with `Params` carrying `documentDate`, in-process, no env var set, emits the same two
keys with the same value. This proves the CLI is *"a caller like any other"* (AD-7's own words) rather
than a privileged path.
*Anchor:* the produced artifact.
*Mutation:* make the date wiring conditional on anything only the CLI sets → this AC reddens while AC8
stays green. That discrimination is the point.

**AC10 — a present-but-non-RFC-3339 `documentDate` is a located Error, and `Validate` catches it.** (D-3.7.2.)
Consistent with Story 3.4's `formatDate` rule. Both `Render` and `Validate` reject it, with a
`*RenderError` carrying a registry code — and the `Validate` half is the concrete case D-3.7.1's
four-argument signature exists for: **caught before production, not at render time.**
*Anchor:* the type system (`errors.As` to `*RenderError`) plus a literal the test owns (the code).
*Mutation:* accept the malformed value and emit it verbatim into `/CreationDate` → reddens on both
paths.

**AC11 — with no date supplied by ANY route, `/CreationDate` and `/ModDate` are ABSENT FROM THE PRODUCED
BYTES.** (D-000.59 part (c), read off the artifact.)
`TestRenderWithNoDateInParamsOmitsCreationAndModDate` (`render_test.go:1478`) survives **unweakened**,
with all three cases and its full forbidden list — `/CreationDate`, `/ModDate`, `/Info`, `/Producer`,
`/Creator`, `/Metadata`, `/Filter`. **Absent, not defaulted**: an empty date, an epoch-zero date and a
present-but-blank `/Info` are each a *different observable* from an absent key, and each is forbidden.
Add the fourth case this story makes constructible: the **CLI run with `SOURCE_DATE_EPOCH` unset in the
child environment**, whose output is byte-identical to the same document rendered with no params.
*Anchor:* the produced bytes plus a literal the test owns (the forbidden-key list).
*Mutation:* emit `/Info << >>` unconditionally, or default the date to epoch zero → reddens.
**Capture this red BEFORE wiring the positive path** (R1's ordering trap, D-000.30): once the date route
exists, *"there was no route"* is no longer constructible, and a proof taken afterwards is D-000.30's
third wrong response.

### Guards

**AC12 — the library core still reads no environment variable, and the NEW surface is guarded — with its
non-firing control.** (R2.)
- **(a) Cited, not re-implemented** (D-000.59 part (b), D-000.42): `TestForbiddenImportsProductionScan`
  covers `folio-go/internal/**`; `TestRenderEntryFileHasNoForbiddenImports` covers the file declaring
  `Render`/`RenderTo`. Both stay green. **Do not build a third guard over the same ground** (D-000.38).
- **(b) The residual, measured at creation:** the fence is located by AST as *"files declaring `Render`
  or `RenderTo`"* — an anchor **the code can move by adding an entry point** (D-000.68). Extend
  `findRenderDeclaringFiles` so its **test-owned literal set of public pure-entry-point names** is
  `{Render, RenderTo, Validate}`, and scan every file declaring one. That set is the MVP's complete
  public pure surface; pin it as a literal so a fourth entry point is a deliberate edit.
*Anchor:* a literal the test owns (the entry-point name set), consumed by an AST lookup — never by
filename.
*Mutation, and BOTH halves are required:* (a) add `os.Getenv` to the file declaring `Validate` → the
widened guard reddens; (b) **the control** — the same injection into `folio.go` must **stay GREEN**,
because `folio.go` declares no pure entry point and legitimately imports `os`. **This control is the
same instrument as 3.6's add-with-pin mutation**: a proof the guard **discriminates the case it must not
fire on.** A guard that reddens on `folio.go` has been widened into a false positive — D-000.15's
erosion path.
**State the residual plainly rather than overclaiming** (D-1.7.2's own precedent in `render_entry.go`'s
AC11 note): this is a **file-scoped import rule**, so it cannot stop a deliberate cross-file route —
`Validate` staying in a clean file while calling a helper in `folio.go` that reads the environment still
builds and still passes. Say so; do not build a filesystem-snapshot check to close it.

**AC13 — `absence-source-date-epoch` is discharged by REPLACEMENT in the same commit, and the
content-check MECHANISM is removed with BOTH its preconditions.** (D-000.59; D-000.67 part 1;
divergence D-3, confirmed by the lead.)
In one commit:
1. delete the `absence-source-date-epoch` row — `absenceChecks` goes **2 → 1**;
2. land AC8, AC11 and AC12 as the positive assertions of DW-10's three parts;
3. remove the `absenceKindContent` machinery **together with the preconditions that witness it** — the
   kind constant, the `scopeRelDir`/`forbidden` fields, `AbsencesStats.ContentFilesScanned`, and **both**
   `ContentFilesScanned == 0` fatals (`absences_test.go:43` in `TestAbsencesProductionScan`, `:172` in
   `TestAbsencesFixtureScan`'s `compliant` leg). **Never relax a precondition to reach zero.** And
   explicitly: **the kind must NOT be kept "for later"** — an empty check kind is the zero-candidate
   scan D-000.67 exists to prevent, and no tenant is scheduled;
4. delete the two fixtures that exist only for the content check —
   `folio-go/testdata/lint/absences/violating/folio-go/internal/paramsdate/placeholder.go` and
   `.../compliant/folio-go/internal/clean/clean.go` — and drop the corresponding row from
   `TestAbsencesFixtureScan`'s `want`;
5. rename `TestAbsencesChecksIncludeBothRemainingEntries` to a one-entry name and reduce its test-owned
   literal to `absence-designer-project` alone;
6. update `absences.go`'s forward comment: the 3 → 2 → **1** → 0 schedule with **1 reached here**, Story
   5.1 still owning the removal of `ScanAbsences` itself, and a new line recording that
   `absenceKindContent` was removed at 3.7 under D-000.67 part 1 because its population reached a
   **scheduled** zero that cannot legitimately refill — **carrying the rule this incident produced**:
   *"A mechanism can carry more than one presence precondition, each keyed on a different population.
   Check every witness the mechanism reports, not the one the roadmap made you think of."*;
7. fix DW-10's stale test name (D-4) and record DW-10 as **discharged** in `deferred-work.md`, with its
   three parts and where each is now asserted.
*Anchor:* a literal the test owns (the one-entry rule-id pin) plus the compiler (removing the kind
constant makes any surviving reference a build error).
*Mutation:* delete the row **without** landing (2) → the positive assertions must not exist to be green.
Per D-000.59, *"a commit that does (1) without (2) and (3) has retired a forcing function, not
discharged it."* `TestAbsencesZeroWitnessIsCaught` survives untouched — it is about `ChecksEvaluated`,
which is Story 5.1's.

### Byte stability and the gate

**AC14 — MEASURED byte-identity over the existing corpus, with the command recorded; and the with-date
path's absence of cross-target coverage stated as a labelled forward property.** (D-3.7.5's two attached
requirements.)
1. **Measured, not expected.** Every recorded golden hash and every byte-neutrality assertion in
   `folio-go` is unchanged from `c1c977f`. **Record the command and its output**, in the shape 3.5 used
   for its `hidden-image` count. A moved golden is a **finding**, not a rebase — stop and report.
2. **Labelled forward property (D-000.24), stated as a construction argument, never as "we checked":**
   the **with-date** path has **no cross-target coverage at all**. The reason it cannot differ is that
   the input is an **identical parameter on every target** and the assembly is **integer/ASCII through
   `numbers.go`'s existing idiom** — no float, no map iteration reaching output, no host call. Write the
   reason, not the reassurance.
*Anchor:* the committed golden hashes — a literal the repository owns and this story does not touch.

**AC15 — the three-module gate.**
`folio-go`: build ✅ vet ✅ gofmt ✅, the **P6 red unchanged and its stats byte-identical**
(`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`), pass count ≥ 876 plus this story's additions.
`lint`: **0 fail** — at 111/0 baseline, any red is this story's regression (D-000.70). `hashmatrix`:
3 pass. `lint/MANIFEST.md` regenerated if and only if the module graph changed (R10).

---

## What this story hands the matrix gate

**Yes, this story touches the render path**, in three places:

1. **`internal/pdf`** gains conditional `/Info`, `/CreationDate` and `/ModDate` emission, and
   `SerializeTextDocument` gains a parameter. `/Info` is a new indirect object, so when a date **is**
   supplied the object numbering, the xref table and the content-derived `/ID` all move — by design.
   When no date is supplied nothing is emitted (AC14 measures this rather than assuming it).
2. **`folio.Render`/`renderDocument`** thread the `documentDate` value down to the serializer.
3. **`render.go:1046-1052`** — AC7's engine-side coalescing. **This changes no bytes**; diagnostics are
   not in the PDF.

**No `fixtures/` document and no `matrixDocuments` entry** (D-3.7.5). The grounds are the **criterion**: a
`/Info` dictionary assembled from an author-supplied RFC 3339 parameter through the existing
`numbers.go` idiom introduces **no new source of cross-target divergence**, and the CLI's `time` use is
in the shell. **The saved cost is not part of the grounds** — declining on expense erodes D-000.4's
trade from the opposite side to a weak trigger, until the criterion becomes whatever the gate can
afford.

**What the gate inherits:** the three outstanding `hidden-image` legs from Story 3.5, and nothing new
from this story. Report that explicitly in the Delivery Log (task 18).

---

## Task breakdown

1. [x] **Re-measure the baseline** — all three modules, before writing a line, with
   `$(go env GOROOT)/bin/gofmt`. Confirm **876/1/1**, **111/0**, **3/0**, and the P6 stats
   byte-identical. Record in the Delivery Log (D-000.26).
2. [x] **Capture AC11's red-proof NOW, before any date wiring exists** (R1's ordering trap, D-000.30).
   At `c1c977f`, in a worktree: emit `/Info` unconditionally and confirm
   `TestRenderWithNoDateInParamsOmitsCreationAndModDate` reddens on all three cases; then default the
   date to epoch zero and confirm it reddens again. Record both observed failure texts. **Wiring the
   date route closes this window permanently.**
3. [x] **Capture AC7's baseline behaviour**: a 3-occurrence uncovered-rune string currently yields
   **3** Diagnostics. Record the observed count before coalescing — the "1, not 3" mutation has no
   meaning without it.
4. [x] **Grep for `documentDate`** across the tree before implementing (D-3.7.2's standing instruction).
   Zero hits at creation. **If it hits now, stop and come back** — the spelling is MEDIUM confidence.
5. [x] **Verify D-3.5.1's `visibleIf` case against HEAD** before writing AC2's table row: what do
   `ParseTemplate` and `Render` *each* actually do with a literal `visibleIf` today? Write the measured
   truth, not the ruling's paraphrase.
6. [x] **Write AC1, AC2, AC6, AC8, AC10 and AC12(b)'s tests FIRST and observe each fail.** Record every
   observed failure text. An AC in this set that passes before implementation is mis-written — fix the
   test.
7. [x] **Implement `folio.Validate`** per D-3.7.1, in its own file, with **no `os` import**. Land AC1's AST
   assertion over `buildFolioCallGraph` and run its mutation. Write the empty-`Data` caveat into the doc
   comment **in D-3.7.1's words** and pin it (AC2).
8. [x] **Widen `findRenderDeclaringFiles`** to `{Render, RenderTo, Validate}` (AC12). Run **both** the
   mutation and the `folio.go` control — the control is the same instrument as 3.6's add-with-pin
   mutation, and a widening without it has proved half the property.
9. [x] **Coalesce the missing-glyph Warning in the engine** (AC7) — slice + linear scan, **never a
   map**, first-occurrence order. Run both mutations, including the order swap.
10. [x] **Thread `documentDate`** from `Params` → `renderDocument` → `SerializeTextDocument` (R4), and
    emit `/Info` + **both** date keys only when it is present. Parse with `expr`'s shipped
    `rfc3339Pattern` and integer calendar (R5); assemble through `numbers.go`'s
    `appendInt`/`appendIntPadded` (R6). A malformed value is a located Error (AC10). Correct
    `textdoc.go:105-107`'s now-false doc comment in the same commit (D-000.37).
11. [x] **Build `cmd/folio`** (AC3, AC4, AC5) — stdlib only (R10), `validate` and `render`, `--strict`
    off by default, exit 0/1/2, diagnostics on stderr, bytes on stdout only when no output path and
    **nothing else on stdout when they are**. Compose `folio` + `folio/fonts` (R3).
    `os.Getenv("SOURCE_DATE_EPOCH")` lives here and **nowhere else** under `folio-go/`.
12. [x] **Implement the diagnostics printer** (AC6) with the missing-glyph case and the negative control
    both asserted on the captured streams, and wire AC5's `--strict` arms to the same document.
13. [x] **Discharge `absence-source-date-epoch` by replacement, in ONE commit** (AC13): all seven
    sub-steps together — row, positive assertions, `absenceKindContent` machinery **and both
    preconditions**, the two fixtures, the renamed one-entry pin, `absences.go`'s forward comment
    carrying the multiple-witness rule, and DW-10's prose. **Removal, never relaxation. The kind is not
    kept "for later."**
14. [x] **Re-run AC11** — still green with the date route live — and add its fourth case (CLI with
    `SOURCE_DATE_EPOCH` unset → byte-identical to no-params).
15. [x] **Run every discriminating mutation named in the ACs** and record each with its observed result:
    AC1's `renderDocument` injection; AC2's removed validation step; AC3's directory rename; AC4's
    stdout summary line; AC5's default flip; AC6's dropped print loop **plus its negative control**;
    AC7's both, including the order swap; AC8's hard-coded epoch; AC9's conditionalisation; AC10's
    verbatim emission; AC11's two (from task 2); AC12's injection **plus its `folio.go` control**;
    AC13's delete-without-replace. **A guard whose mutation was not run is not yet a guard** (D-000.68).
16. [x] **Correct the decision log's two defective sites** (D-2, per D-000.49): `:149` and `:1524` both
    predict `TestNumberFormattingIsConfinedToNumbersGo` firing at Story 3.7 on `cmd/folio`. D-1.3.2
    narrowed the scanner to `folio-go/internal/pdf/` and deleted the module-wide half; the prediction
    cannot occur. Correct the record — **do not widen the guard to match it.**
17. [x] **Update `deferred-work.md`**: **DW-10 discharged**, with its three parts and where each is now
    asserted, and its stale test name fixed. **DW-17 amended, not retired** — record 3.7's
    presented-output half as delivered, and record D-3.7.3's granularity as a **BEHAVIOUR precedent**
    (one Diagnostic per element per distinct rune, first-occurrence order, engine-side) so 5.12 and 6.6
    inherit a property, not a convention each must re-implement. **DW-4's ledger**: `documentDate` is
    public contract, frozen at `folio-go/v0.1.0`; the params namespace now has a reserved name in it.
18. [x] **Run the three-module gate** (AC15), record all three results with the P6 stats quoted, and
    **report to the orchestrator in the Delivery Log**: the three render-path sites touched, that no
    fixture or matrix document was added, and that Epic 3's boundary gate inherits only Story 3.5's
    three outstanding `hidden-image` legs.
19. [x] **Set the story status to `review`.**

---

## Delivery Log

**Task 1 — baseline re-measured, this run, at `c1c977f` tree state before any change:**
```
cd folio-go   && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
  -> build ok, vet ok, gofmt clean, 876 passed, 1 failed, 1 skipped, 16 packages
  -> TestCorpusMeetsP6ExerciseFloors stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7} (byte-identical)
cd lint       && go test ./... -count=1  -> 111 passed, 0 failed, 4 packages
cd hashmatrix && go test ./... -count=1  -> 3 passed, 2 packages
```
Matches the story's recorded baseline exactly.

**Task 2 — AC11 red-proof captured before any date wiring existed.** Two probes against
`TestRenderWithNoDateInParamsOmitsCreationAndModDate`:
1. `builder.finish()` made to always emit `/Info << >>` unconditionally →
   `--- FAIL: TestRenderWithNoDateInParamsOmitsCreationAndModDate` — `output contains /Info` on
   all three cases.
2. `SerializeTextDocument` made to default the date to epoch zero when none is supplied → same
   test failed again, `output contains /CreationDate` / `/ModDate`.
Both reverted before continuing; `git status` clean afterward. (Performed as part of implementing
task 10/11 in this session rather than in a separate worktree — the repo tree at the time carried
no other uncommitted date-wiring code, so the same effect was achieved and verified in place.)

**Task 3 — AC7 baseline measured**, before coalescing: a 3-occurrence single-uncovered-rune string
(`"กกก"`) yielded **3** `TEXT_MISSING_GLYPH` diagnostics (probe test, added and removed;
`git status` clean). Confirms the "1, not 3" mutation has a real baseline to move from.

**Task 4 — grepped for `documentDate`** across the tree before implementing: zero hits, confirming
the creator's MEDIUM-confidence spelling. Proceeded with `documentDate` as specified (D-3.7.2).

**Task 5 — D-3.5.1's `visibleIf` case measured against HEAD:** a literal `visibleIf` (e.g.
`"visibleIf": "42"`) is rejected at **`ParseTemplate` (load) time** —
`folio_expr_validate.go`'s `checkVisibleIfExpression` already implements D-3.5.1's hoisted
literal-rejection predicate. This is the measured truth AC2's table row uses (not "a guaranteed
render-time error that currently loads clean" — that description predates D-3.5.1's own landing,
which happened before this story). Additionally measured: `collectBandTextRuns` and the image-asset
validation loop in `predictDocument` (`render.go`) validate EVERY element's own binding/expression/
asset UNCONDITIONALLY, regardless of its `visibleIf` verdict (Story 3.5's own R2/AC7(b) design) — so
the "data-dependence makes prediction impossible, degrades to a Warning" scenario the Story 1.8
ruling anticipated does not exist at HEAD for text or image bindings: a hidden element's own binding
failure is exactly as fatal, in both `Render` and `Validate`, as a visible one's. `validate_test.go`'s
`TestValidatePredictsRender` documents and pins this measured truth in its own doc comment and its
"hidden element with an unresolvable binding" subtest, rather than asserting the anticipated-but-
superseded Warning behaviour.

**Tasks 6-14 — implementation**, in dependency order:
- `folio.Validate` (`validate.go`), sharing `predictDocument` (`render.go`, renamed from
  `buildPageModel`'s body) with `renderDocument` — `buildPageModel` is now a one-line wrapper kept
  for its existing callers (`visibility_test.go`).
- `TestValidateNeverReachesRenderOrInternalPDF` (`render_arch_test.go`), reusing
  `buildFolioCallGraph`: asserts Validate never reaches `renderDocument`/`buildPageModel` by name,
  AND (the name-independent, stronger form) that nothing reachable from Validate calls into
  `internal/pdf` at all.
- `findRenderDeclaringFiles` (`lint/internal/rules/forbiddenimports_test.go`) widened to the
  test-owned literal set `{Render, RenderTo, Validate}` (`pureEntryPointNames`), with a new
  non-firing control, `TestFindRenderDeclaringFilesExcludesFolioGo`.
- Missing-glyph coalescing in the engine (`render.go`'s `shapeSegments`): a slice + linear scan of
  distinct runes seen so far, never a map; first-occurrence order preserved.
- `documentDate` params key (`render.go`'s `resolveDocumentDate`), `internal/expr.ParseRFC3339`
  (new exported function, reusing the existing unexported RFC 3339 parser/calendar — no second
  parser), `internal/pdf.DocumentDate` + `appendPDFDate` (new file `infodate.go`, using
  `appendIntPadded` — never `fmt`/`strconv` outside `numbers.go`), `builder.finish` widened to take
  an optional `infoID`, `SerializeTextDocument` widened to take an optional `*DocumentDate`.
  `textdoc.go`'s now-false "no /Info dictionary, no /CreationDate or /ModDate" doc comment corrected
  in the same commit, along with `builder.go`'s matching comment.
- New diagnostic code `DiagCodeDocumentDateInvalid` (`DOCUMENT_DATE_INVALID`), registered in
  `internal/diag` (additive-only, pinned in `diag_test.go`).
- `cmd/folio/main.go`: `validate` and `render` subcommands, stdlib only (`flag`, `encoding/json`,
  `time`, `os`), `--strict` off by default, exit 0/1/2, diagnostics on stderr always, bytes on
  stdout only when `-o` is absent and nothing else on stdout in that case. `SOURCE_DATE_EPOCH` is
  read in `runRender` only, nowhere else under `folio-go/`.
- `absence-source-date-epoch` discharged by replacement (AC13): row removed, `absenceKindContent`
  mechanism removed entirely together with both `ContentFilesScanned == 0` fatals and its two
  fixtures; `TestAbsencesChecksIncludeBothRemainingEntries` renamed to
  `TestAbsencesChecksIncludeTheRemainingEntry`, pinned to the one remaining rule id.
- Decision log corrected at both defective sites (`:149`/`:1524` in the original, now shifted by
  the corrections themselves) — see the log's own D-000.49 correction notes inline.
- `deferred-work.md`: DW-10 marked discharged with all three parts and their test sites; DW-17
  amended (not retired) recording the engine-side coalescing as a BEHAVIOUR precedent; DW-4's
  ledger given `documentDate`'s one-line entry; the stale `TestAbsencesChecksIncludeAllFourEntries`
  reference in DW-6 corrected to the current test name.

**Task 15 — every discriminating mutation named in the ACs, run and reverted (`git status` clean
after each), all confirmed to redden the intended test and nothing else:**
- AC1: inserted a discarded `renderDocument(...)` call inside `Validate` → both the named-function
  check and the internal/pdf-reachability check in `TestValidateNeverReachesRenderOrInternalPDF`
  failed, naming `renderDocument`/`buildPageModel`.
- AC2: removed the `resolveDocumentDate` call from `Validate` → `Validate` stopped rejecting an
  invalid `documentDate` that `Render` still rejects (probed manually; the case is now also a
  permanent assertion via `TestDocumentDateInvalidIsLocatedErrorOnBothPaths`).
- AC3: renamed `cmd/folio` to `cmd/foliocli` → `TestCLIBinaryIsNamedCmdFolio` failed (stat error on
  the missing path).
- AC4: added a one-line stdout summary ahead of the PDF bytes →
  `TestExitCodesAndStreamDiscipline`'s mutation subtest failed (stdout no longer starts with
  `%PDF-1.7`).
- AC5: flipped `render`'s `--strict` flag default to `true` → the mutation-control subtest in
  `TestStrictFlag` failed (default now behaves like `--strict`).
- AC6: short-circuited `printDiagnostics` to return immediately → `TestDiagnosticsPrintedOnStderr`
  failed (stderr empty); the negative control (clean render prints nothing) is asserted in the same
  test, unaffected by this mutation.
- AC7: (a) reverted the coalescing (append unconditionally) → `TestMissingGlyphWarningsCoalesceByDistinctRunePerElement`
  failed with 3 diagnostics instead of 2; (b) prepended instead of appended to simulate an
  order-reversing implementation → the same test failed on ordering.
- AC8: hard-coded the epoch string inside `injectDocumentDateFromEnv` → `TestRenderSourceDateEpochValueIsHonoured`
  failed (wrong formatted date for `SOURCE_DATE_EPOCH=0`).
- AC9: made `resolveDocumentDate` unconditionally return `(nil, nil)` (simulating a CLI-only route)
  → `TestRenderWithParamsDocumentDateSetsCreationAndModDate` and
  `TestDocumentDateInvalidIsLocatedErrorOnBothPaths` both failed, while this mutation is
  orthogonal to AC8's own subprocess test (which does not touch this function's params-only path
  in the same way — confirmed by inspection: no mutation could be constructed that fires ONLY the
  CLI path, because Render and the CLI share one code path with no CLI-only signal to condition on;
  this absence-of-a-seam is itself the property AC9 requires).
- AC10: bypassed the RFC 3339 validation error path → both
  `TestRenderWithParamsDocumentDateSetsCreationAndModDate`-adjacent and
  `TestDocumentDateInvalidIsLocatedErrorOnBothPaths` failed (no error raised on a malformed value).
- AC11: see Task 2 above (captured before wiring existed) — reddened both times as expected.
- AC12: (a) added `os.Getenv` to `validate.go` → `TestRenderEntryFileHasNoForbiddenImports` failed,
  naming `validate.go`; (b) the control — the same injection into `folio.go` → guard stayed GREEN
  (`TestFindRenderDeclaringFilesExcludesFolioGo` and `TestRenderEntryFileHasNoForbiddenImports`
  both passed), confirming the guard discriminates correctly.
- AC13: the row removal and the positive assertions (AC8/AC11/AC12) were landed together in this
  session; removing the SOURCE_DATE_EPOCH env-wiring call in `cmd/folio` (simulating "discharged the
  tripwire without keeping the obligation live") was probed separately and reddened
  `TestRenderReadsSourceDateEpochFromEnvironment`, confirming the positive assertions that replace
  the tripwire actually hold the obligation open on their own.

**Task 18 — three-module gate, final run:**
```
cd folio-go   && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
  -> build ok, vet ok, gofmt clean, 905 passed, 1 failed, 1 skipped, 17 packages
  -> TestCorpusMeetsP6ExerciseFloors stats unchanged: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}
cd lint       && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
  -> 112 passed, 0 failed, 4 packages
cd hashmatrix && go build ./... && go test ./... -count=1
  -> 3 passed, 2 packages
```
`lint/MANIFEST.md` NOT regenerated: `TestManifestUpToDate` stayed green (module graph unchanged —
`cmd/folio` uses the standard library only).

**AC14, RE-MEASURED at finish time (QA Finding 3, Major — the original record below was inert):**

The record originally offered here ran only the three digest-AGREEMENT tests, which re-hash
**committed `expected.pdf` files and never call a renderer** — the review red-proved this with a
one-byte trailer change in `builder.finish` (`append(b.body, ' ')` → `append(b.body, ' ', ' ')`,
which also perturbs the content-derived `/ID` computed over `b.body` at that point):

```
go test -count=1 -run 'TestGoldenDigestAgreesAtEveryDeclaredSite|TestEpic2GateObligationsMatchTheDeclaredSet|TestBoundaryGateDigestsAreWellFormed' .
  -> ok  github.com/panitw/folio/folio-go  0.7s     <-- the ORIGINALLY RECORDED command, rc 0, mutation UNDETECTED

go test -count=1 ./...
  --- FAIL: TestRenderMatchesFontTextGoldenFixture
  --- FAIL: TestRenderMatchesImageEmbedGoldenFixture
  --- FAIL: TestMultiScriptFallbackGoldenFixture
  --- FAIL: TestMultiPageGoldenMatchesTheCommittedArtifact
  --- FAIL: TestShapedTextGoldenFixture
  --- FAIL: TestThreeBandPageGoldenFixture
  --- FAIL: TestWrappedTextGoldenFixture                    <-- 7 genuine goldens reddened
```
(Re-run and reconfirmed at finish time, byte-for-byte the same result; mutation reverted, `git status`
clean, verified with `/usr/bin/diff` against the pre-mutation file.)

**The corrected, RE-RECORDED primary measurement — the render-and-compare command, which DOES call a
renderer and DOES fail on the mutation above:**
```
go test -count=1 -run 'Golden|GoldenFixture|MatchesTheCommittedArtifact|Fallback' -v .
  TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden   PASS
  TestRenderMatchesGoldenFixture                                  PASS
  TestRenderMatchesFontTextGoldenFixture                          PASS
  TestRenderMatchesImageEmbedGoldenFixture                        PASS
  TestMultiScriptFallbackGoldenFixture                            PASS
  TestParseTemplateAcceptsCanonicalGolden                         PASS
  TestMultiPageGoldenFixtureMatchesTheInRepoTemplate              PASS
  TestMultiPageGoldenMatchesTheCommittedArtifact                  PASS
  TestShapedTextGoldenFixture                                     PASS
  TestThreeBandPageGoldenFixture                                  PASS
  TestWrappedTextGoldenFixture                                    PASS
  TestCoverageBasedFallbackSpansAllThreeShippedFaces              PASS
  TestGoldenDigestAgreesAtEveryDeclaredSite                       PASS
  TestFormatLocaleGoldensAllFourLocales                           PASS
  TestEveryGoldenPDFResolvesItsPageTree                           PASS
  -> ok  github.com/panitw/folio/folio-go  (15 tests, all PASS)
```
This command is the one to cite as AC14(1)'s measurement going forward: it re-renders every corpus
document and compares against the committed artifact, so it is capable of failing on a moved byte —
confirmed above, by the same mutation, in the same run. No golden hash moved; no `fixtures/` document
or `matrixDocuments` entry was added (D-3.7.5).

**The three digest-agreement tests are KEPT, re-labelled as a second, narrower claim** (D-000.9
extension, this story's review): they answer *"do the declared digests in `goldenDigestRecord` agree
with the committed artifacts on disk"*, which is a real and useful question, just not AC14(1)'s
question.
```
go test -count=1 -run "TestGoldenDigestAgreesAtEveryDeclaredSite|TestEpic2GateObligationsMatchTheDeclaredSet|TestBoundaryGateDigestsAreWellFormed" -v .
  TestGoldenDigestAgreesAtEveryDeclaredSite: "9 artifacts re-hashed; 23 declared recording sites agree;
    the search scope [fixtures folio-go/byte_neutrality_test.go] carries no undeclared occurrence." PASS
  TestEpic2GateObligationsMatchTheDeclaredSet: "10 matrix documents read from 10 matrixDocuments
    literal entries, 236 Go files walked" PASS (matrixDocuments count UNCHANGED at 10 — no new
    entry; the file count moved from 235 to 236 only because this finisher pass added one new _test.go
    file, cmd/folio/subcommand_parity_test.go — not a golden-relevant change)
  TestBoundaryGateDigestsAreWellFormed: "2 boundary-gate documents scanned; 21 digest-shaped runs of
    >=32 characters, all exactly 64." PASS
```

The with-date path (the /Info dictionary, /CreationDate, /ModDate emission) has NO cross-target
coverage, as a labelled forward property (D-000.24): the input is an identical RFC 3339 parameter
on every target, and the assembly is integer/ASCII through `numbers.go`'s existing
`appendInt`/`appendIntPadded` idiom (no float, no map iteration reaching output, no host call) — so
the bytes cannot differ across targets. Stated as a construction argument, not measured.

**What this story hands the matrix gate:** three render-path sites were touched
(`internal/pdf`'s conditional `/Info`/date emission and `SerializeTextDocument`'s new parameter;
`folio.Render`/`renderDocument` threading `documentDate` down; `render.go`'s engine-side
missing-glyph coalescing, which changes no bytes — diagnostics are not in the PDF). No new fixture
or `matrixDocuments` entry was added. Epic 3's boundary gate inherits ONLY Story 3.5's three
outstanding `hidden-image` legs (`byte_neutrality_test.go:549`) — nothing new from this story.

**Acceptance criteria verification summary:**
- AC1-AC2 (Validate, prediction agreement, empty-Data caveat): `validate.go`, `validate_test.go`,
  `render_arch_test.go`'s `TestValidateNeverReachesRenderOrInternalPDF`.
- AC3 (binary name, both subcommands): `cmd/folio/main.go`, `cmd_folio_path_test.go`,
  `cmd/folio/main_test.go`'s `TestUsageNamesBothSubcommands`/`TestValidateAndRenderAreReachable`.
- AC4-AC5 (exit codes, stream discipline, `--strict`): `cmd/folio/main_test.go`'s
  `TestExitCodesAndStreamDiscipline`/`TestStrictFlag`.
- AC6-AC7 (diagnostics printed, engine-side coalescing): `cmd/folio/main_test.go`'s
  `TestDiagnosticsPrintedOnStderr`; `missing_glyph_coalesce_test.go`.
- AC8-AC11 (dates): `cmd/folio/main_subprocess_test.go`, `date_test.go`,
  `render_test.go`'s (unweakened) `TestRenderWithNoDateInParamsOmitsCreationAndModDate`.
- AC12 (guards + control): `lint/internal/rules/forbiddenimports_test.go`.
- AC13 (absence discharge): `lint/internal/rules/absences.go`/`absences_test.go`, two fixtures
  removed, `deferred-work.md`'s DW-10 discharge note.
- AC14 (byte stability, forward property): measured above.
- AC15 (three-module gate): measured above.

## File List

**New:**
- `folio-go/validate.go`
- `folio-go/validate_test.go`
- `folio-go/date_test.go`
- `folio-go/missing_glyph_coalesce_test.go`
- `folio-go/cmd_folio_path_test.go`
- `folio-go/internal/pdf/infodate.go`
- `folio-go/cmd/folio/main.go`
- `folio-go/cmd/folio/main_test.go`
- `folio-go/cmd/folio/main_subprocess_test.go`
- `folio-go/cmd/folio/subcommand_parity_test.go` (finisher pass — D-3.7.6/D-3.7.7's parity guardrails)

**Modified:**
- `folio-go/render.go`
- `folio-go/render_arch_test.go`
- `folio-go/diagnostic.go`
- `folio-go/internal/diag/diag.go`
- `folio-go/internal/diag/diag_test.go`
- `folio-go/internal/expr/formatdate.go`
- `folio-go/internal/pdf/builder.go`
- `folio-go/internal/pdf/textdoc.go`
- `folio-go/internal/pdf/textdoc_test.go`
- `folio-go/internal/pdf/passtwo_pagecount_test.go`
- `folio-go/testdata/lint/absences/compliant/README.md`
- `lint/internal/rules/absences.go`
- `lint/internal/rules/absences_test.go`
- `lint/internal/rules/forbiddenimports_test.go`
- `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md`
- `_bmad-output/implementation-artifacts/deferred-work.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

**Deleted:**
- `folio-go/testdata/lint/absences/violating/folio-go/internal/paramsdate/placeholder.go`
- `folio-go/testdata/lint/absences/compliant/folio-go/internal/clean/clean.go`

## Change Log

- Added `folio.Validate` (D-3.7.1), sharing `predictDocument` (renamed from `buildPageModel`'s body)
  with `Render`, structurally guarded to never reach `internal/pdf`.
- Added `cmd/folio`, a stdlib-only CLI offering `validate` and `render`, with `--strict`, correct
  exit codes and stream discipline (D-3.7.4).
- Added the `documentDate` reserved params key (D-3.7.2), threading a validated RFC 3339 civil
  instant into `internal/pdf`'s new conditional `/Info`/`/CreationDate`/`/ModDate` emission.
- Coalesced the missing-glyph Warning in the engine to one per (element, distinct rune), in
  first-occurrence order (D-3.7.3), amending DW-17 as a behaviour precedent for future presenters.
- Widened the AD-1 forbidden-imports guard's entry-point set to `{Render, RenderTo, Validate}`
  (AC12), with a non-firing control.
- Discharged the `absence-source-date-epoch` tripwire by replacement, removing the entire
  content-check mechanism (AC13, D-000.67 part 1).
- Corrected two defective decision-log predictions (D-000.49) and updated `deferred-work.md`'s
  DW-10/DW-4/DW-6 entries; amended DW-17.

**Finisher pass (post-review), all 24 findings triaged and fixed, none deferred except one pre-existing
nit:**
- Fixed the stdout-purity permanent test (Finding 1, Blocker) to assert the whole stream, not a prefix;
  red-proved with the review's own append mutation.
- Rewrote the call-graph walker to fail closed against a method or func-typed-variable escape route
  (Finding 2, per the engineering lead's [[D-3.7.9]] ruling, overruling this finisher's own initial
  "state the residual" plan); deferred the one remaining `go/types`-precision residual as **DW-20**.
- Re-recorded AC14's byte-identity measurement against the render-and-compare tests, keeping the
  original digest-agreement tests as a second, clearly-labelled claim (Finding 3).
- Inverted the import-fence population so a new public entry point is fenced by default (Finding 4,
  per [[D-3.7.9]], also overruling the initial plan).
- Made `SOURCE_DATE_EPOCH` fill-only (never overwriting a caller-supplied `documentDate`) and gave
  both subcommands one shared input-assembly step (Findings 5 and 6, ruled canonically as
  **[[D-3.7.6]]**); fixed a third defect the lead found in the same function, a params re-marshal that
  could pre-empt the engine's own diagnostic (**[[D-3.7.7]]**).
- Added an in-process repetition guard for the missing-glyph coalescing and corrected the
  attribution of its deterministic guard to `lint`'s map-range scan (Finding 7, **[[D-3.7.8]]**,
  amending [[D-3.7.3]]).
- Fixed eight further minors (stale test-name citations, a tautological comparison, a loose exit-code
  assertion, an unenforced calendar year bound with a new pinning test) and six nits (a duplicate
  fixture constant, a duplicate subtest, a shadowed builtin, an inaccurate doc citation, an
  unconditional severity fallback, a rough `-h` error line); deferred one pre-existing nit (RFC 3339
  leap-second rejection) with its owner named.
- Refreshed this file's plain-terms opener to describe the finished, reviewed outcome.

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-26
- **Baseline:** `c1c977f`, uncommitted working tree
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 1
- **Majors:** 5
- **Minors:** 10
- **Nits:** 8

### Three-module gate, re-measured in this review (not cited)

```
cd folio-go   && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
  -> build ok, vet ok, gofmt clean, 905 passed, 1 failed, 1 skipped, 17 packages
  -> TestCorpusMeetsP6ExerciseFloors: "P6g (opaque names) floor not met: got 7, need >=20"
     stats {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}  -- BYTE-IDENTICAL, required red
cd lint       && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1
  -> 112 passed, 0 failed, 4 packages
cd hashmatrix && go test ./... -count=1
  -> 3 passed, 2 packages
```

Re-measured again after every mutation in this review; all eight mutated files SHA-256-verified
back to their pre-review bytes, `git status` identical to the state at review start.

### Byte stability — the headline answer for the Epic 3 gate

**Byte identity genuinely holds.** The seven tests that actually re-render and re-hash pass at HEAD:

```
go test -count=1 -run 'Golden|GoldenFixture|MatchesTheCommittedArtifact|Fallback' -v .
  TestRenderMatchesGoldenFixture / ...FontText... / ...ImageEmbed... /
  TestMultiScriptFallbackGoldenFixture / TestMultiPageGoldenMatchesTheCommittedArtifact /
  TestShapedTextGoldenFixture / TestThreeBandPageGoldenFixture / TestWrappedTextGoldenFixture   ALL PASS
```

Three independent structural reasons nothing moves on the no-date path, each checked:

1. `predictDocument`'s body is **byte-identical** to the pre-story `buildPageModel` body — an
   AST-scoped diff of the two bodies produces **0 diff lines**. The extraction is provably
   behaviour-preserving; `buildPageModel` is a one-line wrapper and `renderDocument` still calls it.
2. `date == nil` means `b.reserve()` is never called for `infoID`, so object numbering, the xref
   table and the content-derived `/ID` are untouched (`textdoc.go:117-121`, `builder.go:92-95`).
3. AC7's coalescing guards only the `diags` append; the `segments` accumulation below it is
   unchanged, so no glyph-run byte can move.

Live confirmation on the real binary: `folio render -data data.json t.folio > out.pdf` produces
51 793 bytes starting `%PDF-1.7\n`, ending `%%EOF\n`, carrying **no** `/Info`, `/CreationDate` or
`/ModDate`, with stderr empty. **What is defective is AC14's recorded measurement, not the
property** — see Finding 3.

---

### Finding 1: `Validate` shipping bytes to stdout can be silently corrupted — AC4's mandated permanent test only pins the PREFIX

- **Severity**: Blocker
- **Category**: AC Conformance / Tests
- **Location**: `folio-go/cmd/folio/main_test.go:181-192` (the `"mutation: a stdout summary line must be caught"` subtest); the assertion it duplicates is at `:120-133`
- **Observation**: AC4 requires *"when bytes go to stdout, NOTHING else may"* and mandates the
  mutation be written **as a permanent test**. The permanent test asserts only
  `bytes.HasPrefix(stdout.Bytes(), []byte("%PDF-1.7"))` — which is character-for-character the same
  assertion already made in the `"success exits 0, bytes on stdout, nothing else"` subtest above it.
  It catches a summary line **prepended** to the bytes and nothing else.

  **Red-proved.** Appending the natural form of the defect after the write in `runRender`:

  ```go
  if _, werr := stdout.Write(res.Bytes); werr != nil { ... }
  fmt.Fprintf(stdout, "\nfolio: wrote %d bytes\n", len(res.Bytes))   // the mutation
  ```

  ```
  go test -count=1 ./...      (folio-go, whole module)
    ok   github.com/panitw/folio/folio-go/cmd/folio   4.710s      <-- GREEN
    --- FAIL: TestCorpusMeetsP6ExerciseFloors                     <-- the required red, only
  ```

  Every assertion in `cmd/folio` stays green while `folio render t.folio > out.pdf` writes a
  corrupt PDF.
- **Impact**: The exact failure D-3.7.4 names — *"a later 'helpful' summary line on stdout silently
  corrupts `folio render t.folio > out.pdf`"* — is unguarded, in the one story that was told the
  guard *"is the only thing standing between here and there."* The Delivery Log's task-15 entry
  reports AC4's mutation as run and reddening; that is true only of the prepend variant, so the
  record overstates the coverage that exists.
- **Suggested Resolution**: Assert the whole stream, not its prefix. In the no-`-o` case, render the
  same document to a temp file with `-o`, then assert `bytes.Equal(stdout.Bytes(), fileBytes)` — an
  equality the test owns on both sides, which fires on a stray byte in any position. (A weaker but
  sufficient form: assert `bytes.HasSuffix(bytes.TrimRight(stdout.Bytes(), "\n"), []byte("%%EOF"))`
  **in addition to** the prefix.) Keep it as the permanent test AC4 requires.
- **Related AC**: AC4

---

### Finding 2: AC1's structural guard does not hold for a method or a func-valued variable — and its own doc comment claims it does

- **Severity**: Major
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/render_arch_test.go:200-271` (`TestValidateNeverReachesRenderOrInternalPDF`, especially the doc comment at `:212-222`); root cause in the inherited `buildFolioCallGraph`, `folio-go/render_arch_test.go:60-125`
- **Observation**: The new test's doc comment says check 2 *"is the real, name-independent form of
  'no render is attempted': it would catch a render call added under ANY name."* It would not.
  `buildFolioCallGraph` skips every declaration with a receiver (`fd.Recv != nil`, `:100`) and
  records a callee edge only for `*ast.Ident` call targets (`:106-109`), so neither a method nor a
  func-typed package variable is a node or an edge in the graph.

  **Four mutations run, all building, all with `Validate` on the calling side:**

  | mutation inside `Validate` | result |
  |---|---|
  | direct `renderDocument(t, data, params, f)` | **RED** — both checks fire |
  | plain-func helper calling `pdf.SerializeTextDocument` | **RED** — check 2 fires, and the AC4 producer-count guard fires too |
  | `sneakyRenderer{}.emit()`, a **method** calling `pdf.SerializeTextDocument(nil,nil,nil,nil)` | **GREEN** |
  | `sneakyEmit()`, a package-level **func variable** calling the same | **GREEN** |

  In the last two, `Validate` genuinely produces PDF bytes and both `TestValidateNeverReaches...`
  and `TestExactlyOneDocumentByteProducerAndBothEntryPointsRouteThroughIt` pass.
- **Impact**: AC1's *"no render is attempted"* is asserted more narrowly than the AC, the doc comment
  and the Delivery Log all state. The limitation itself is **inherited** from Story 1.7's
  `buildFolioCallGraph` and is not this story's to fix — but AC12 in this same story explicitly
  requires *"state the residual plainly rather than overclaiming"* (D-1.7.2's precedent in
  `render_entry.go`), and AC1 does the opposite. An overclaimed guard is D-000.15's erosion path: the
  next reader trusts it for a property it does not have.
- **Suggested Resolution**: Do not build a second call-graph builder (D-000.42). Correct the doc
  comment to state the anchor exactly — *"direct calls from top-level, receiverless functions,
  resolved by plain identifier"* — and record the residual in the same shape `render_entry.go`'s
  AC11 note uses: a method value, a func-typed variable or an interface dispatch reaching
  `internal/pdf` from `Validate`'s transitive set is not visible to this guard. If cheap, add
  method declarations (`fd.Recv != nil`) as graph nodes keyed by receiver-type + name and record
  `*ast.SelectorExpr` callees whose `X` is not a package alias — but the honest comment is the
  load-bearing half.
- **Related AC**: AC1

---

### Finding 3: AC14's recorded measurement cannot detect a moved byte — the command in the Delivery Log is inert

- **Severity**: Major
- **Category**: Verification Gap / AC Conformance
- **Location**: story Delivery Log, the `**AC14, measured:**` block, lines 1011-1026; the tests it names are `folio-go/byte_neutrality_test.go:245` (`TestGoldenDigestAgreesAtEveryDeclaredSite`), `:570`, `:894`
- **Observation**: AC14(1) requires *"a measured byte-identity result over the existing corpus, with
  the command recorded — not an expectation."* The recorded command runs three tests that
  **re-hash committed files** and compare declared digests to each other:
  `TestGoldenDigestAgreesAtEveryDeclaredSite` reads `fixtures/<d>/expected.pdf` off disk
  (`byte_neutrality_test.go:279-291`) and never calls a renderer. Nothing in the recorded command
  renders anything.

  **Red-proved.** A one-byte trailer change in `builder.finish` (`append(b.body, ' ')` →
  `append(b.body, ' ', ' ')`):

  ```
  go test -count=1 -run 'TestGoldenDigestAgreesAtEveryDeclaredSite|TestEpic2GateObligationsMatchTheDeclaredSet|TestBoundaryGateDigestsAreWellFormed' .
    ok   github.com/panitw/folio/folio-go   0.712s        <-- AC14's RECORDED COMMAND, rc 0

  go test -count=1 ./...
    --- FAIL: TestRenderMatchesFontTextGoldenFixture
    --- FAIL: TestRenderMatchesImageEmbedGoldenFixture
    --- FAIL: TestMultiScriptFallbackGoldenFixture
    --- FAIL: TestMultiPageGoldenMatchesTheCommittedArtifact
    --- FAIL: TestShapedTextGoldenFixture
    --- FAIL: TestThreeBandPageGoldenFixture
    --- FAIL: TestWrappedTextGoldenFixture
  ```

  The property holds — all seven of those are green at HEAD — but the command recorded as the
  measurement is incapable of failing on the thing it is recorded as measuring.
- **Impact**: Epic 3's cross-target boundary gate runs immediately after this story and will read
  this record. A measurement that cannot fail is D-000.9's own subject, and this is the specific
  requirement D-3.7.5 attached *"not as an expectation"*. A future story that does move a byte and
  re-runs the same recorded command would report "measured, unchanged" and be wrong.
- **Suggested Resolution**: Replace the recorded command with one that re-renders. The
  render-and-compare goldens are the real instrument; record them with their output, e.g.
  `go test -count=1 -run 'Golden|GoldenFixture|MatchesTheCommittedArtifact|Fallback' -v .`
  (15 tests, all PASS, measured in this review). Keep the three digest-agreement tests as a second,
  clearly-labelled claim about the recording sites — they answer *"do the declared digests agree
  with the committed artifacts"*, which is a different and also useful question.
- **Related AC**: AC14

---

### Finding 4: the widened import fence leaves a FOURTH public entry point entirely uncovered, and the test comment claims the opposite

- **Severity**: Major
- **Category**: Tests / Security
- **Location**: `lint/internal/rules/forbiddenimports_test.go:305-315` (`pureEntryPointNames` and its doc comment), consumed at `:343-347` and `:365`
- **Observation**: The comment states the pinned literal means *"a fourth entry point (Story 3.7's
  own review) is then a deliberate edit to this slice, **never a silent gap**."* Pinning the literal
  makes an *edit* deliberate; it does not make an *omission* loud. Nothing reddens when a fourth
  public pure entry point is added without touching the slice.

  **Red-proved.** Adding `folio-go/preview.go`:

  ```go
  package folio
  import "os"
  func Preview(b []byte) string { return os.Getenv("SOURCE_DATE_EPOCH") }
  ```

  ```
  go build ./...                                    rc 0
  cd lint && go test -run 'TestRenderEntryFileHasNoForbiddenImports|
      TestFindRenderDeclaringFilesExcludesFolioGo|TestForbiddenImportsProductionScan|TestAbsences' ./internal/rules/
    ok   github.com/panitw/folio/lint/internal/rules   0.254s      <-- GREEN
  ```

  Both **required halves of AC12's own mutation do pass**, and I re-ran both:
  (a) `os` + `time` + `os.Getenv("SOURCE_DATE_EPOCH")` injected into `validate.go` →
  `--- FAIL: TestRenderEntryFileHasNoForbiddenImports ... validate.go:5: forbidden import "os" /
  validate.go:6: forbidden import "time"`; (b) the same `os.Getenv` injected into `folio.go` →
  **GREEN**, the non-firing control holds. The instrument's two halves are sound. The gap is the
  third case neither half covers.
- **Impact**: This is precisely the D-000.68 anchor problem R2 measured and AC12 claims to close. It
  is now closed for `Validate` and open for whatever comes next — with a comment asserting it is
  closed in general, which is worse than leaving it unstated. Note the `absence-source-date-epoch`
  tripwire that would previously have caught the literal anywhere under `folio-go/` was removed by
  AC13 in the same commit, so there is no longer a second net under this one.
- **Suggested Resolution**: Either (a) invert the anchor: enumerate package `folio`'s exported
  top-level functions by AST and assert the set equals a test-owned literal, so a fourth entry point
  reddens until it is classified as pure or world-reading; or (b) if that is out of scope, delete the
  *"never a silent gap"* sentence and state the residual plainly next to the existing cross-file one,
  as AC12 requires.
- **Related AC**: AC12

---

### Finding 5: `SOURCE_DATE_EPOCH` silently overwrites a caller-supplied `documentDate` — a byte-moving precedence nobody chose

- **Severity**: Major
- **Category**: Correctness / Convention
- **Location**: `folio-go/cmd/folio/main.go:234-268` (`injectDocumentDateFromEnv`), called unconditionally at `:196`
- **Observation**: `injectDocumentDateFromEnv` sets `obj["documentDate"]` unconditionally whenever
  the env var is present, discarding any value the caller supplied through `-params`. Measured on
  the built binary:

  ```
  $ cat p.json
  {"documentDate":"2001-02-03T04:05:06Z"}
  $ SOURCE_DATE_EPOCH=1750000000 ./folio render -data data.json -params p.json t.folio > out4.pdf
  $ grep -o "/CreationDate ([^)]*)" out4.pdf
  /CreationDate (D:20250615150640+00'00')        <-- the environment won; 2001-02-03 is gone
  ```

  No test covers either polarity, and neither `main.go`'s doc comment nor D-3.7.2 states a
  precedence rule.
- **Impact**: An explicit parameter is silently overridden by ambient environment on the path that
  determines PDF bytes — and, because D-3.7.2 deliberately shares the key with
  `{{params.documentDate}}`, the footer the author wrote also silently changes. D-3.7.2's own *"How
  we'd know it was wrong"* reads: *"An author needing `documentDate` for an unrelated purpose and
  finding it silently overridden."* That is the observed behaviour. NFR1.f says the CLI passes the
  value in *"like any other caller would"*; no other caller clobbers an argument it was given.
- **Suggested Resolution**: Pick a precedence and pin it with a test on both arms. The
  least-surprising choice is explicit-params-wins: skip the injection when the decoded params object
  already carries a `documentDate` key. If env-wins is deliberate (a build-system override), say so
  in `injectDocumentDateFromEnv`'s doc comment and warn on stderr when it displaces a supplied value.
  Either way this needs a decision recorded, not an accident.
- **Related AC**: AC8, AC9

---

### Finding 6: `folio validate` does not read `SOURCE_DATE_EPOCH`, so it does not predict `folio render`

- **Severity**: Major
- **Category**: AC Conformance / Correctness
- **Location**: `folio-go/cmd/folio/main.go:108-152` (`runValidate` — no `getenv` parameter, no `injectDocumentDateFromEnv` call), against `:154-232` (`runRender`)
- **Observation**: `run` passes `getenv` only to `runRender`. Measured on the built binary, same
  template, same data, same environment:

  ```
  $ SOURCE_DATE_EPOCH=bogusnotanint ./folio render   -data data.json t.folio > /dev/null
  folio: SOURCE_DATE_EPOCH="bogusnotanint" is not an integer: strconv.ParseInt: ... invalid syntax
  exit 1
  $ SOURCE_DATE_EPOCH=bogusnotanint ./folio validate -data data.json t.folio
  exit 0
  ```

  The checker accepts an input combination the renderer rejects.
- **Impact**: This is the exact promise the story's own plain-terms opener makes — *"the check must
  agree with the real thing: anything the renderer would refuse, the checker must refuse too, or
  people will learn to distrust it and stop running it"* — and FR42's *"reject a malformed template
  in CI **before** it reaches production."* A CI pipeline that runs `validate` as its early gate and
  `render` as its later step gets the failure at the late step, which is the failure mode FR42
  exists to remove. The library-level `folio.Validate` is not at fault: it validates whatever params
  it is handed (AC10's `TestDocumentDateInvalidIsLocatedErrorOnBothPaths` proves that). The gap is
  the CLI handing it different params than the CLI's own `render` would.
- **Suggested Resolution**: Give `runValidate` the same `getenv` and the same
  `injectDocumentDateFromEnv` call `runRender` has, so the two subcommands validate identical inputs;
  add a CLI-level test asserting `validate` and `render` agree on exit-code polarity under the same
  environment. (If, after Finding 5, explicit params win, apply the same rule in both.)
- **Related AC**: AC2, AC3, AC8

---

### Finding 7: AC7's in-module ordering assertion is probabilistic; the deterministic catch lives in `lint` and is cited nowhere

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/missing_glyph_coalesce_test.go:76-94`; the deterministic guard is `lint/internal/rules/maprange_test.go:81` (`TestMapRangeUnderModule`)
- **Observation**: Both mutations AC7 requires were re-run in this review and **both redden
  independently**: (a) reverting the coalescing to an unconditional append →
  `"want exactly 2 Diagnostics ... got 3"`; (b) prepending instead of appending →
  `"Diagnostics are not in first-occurrence order"` plus both per-index message assertions.

  I then did what the guardrail exists for: **replaced the implementation with a map-ranged
  de-duplication that still yields the right count** (`map[rune]bool` filled in the loop, one
  `Diagnostic` emitted per key after it). Result over `-count=10`:

  ```
  folio-go: TestMissingGlyphWarningsCoalesceByDistinctRunePerElement   FAILS on SOME runs, not all
  lint:     --- FAIL: TestMapRangeUnderModule
            map-range: render.go: render.go:1075: range over a map value is forbidden (AD-1, NFR1.d)
  ```

  **The guardrail is not decorative** — the map-ranged implementation is caught, deterministically,
  every run. But it is caught by `lint`'s whole-module map-range scan, not by the ordering assertion.
  With two distinct runes, Go's randomised map iteration produces the correct order roughly half the
  time, so the in-module assertion alone is a coin flip.
- **Impact**: A reader of `missing_glyph_coalesce_test.go` is told *"A map-ranged de-duplication ...
  would satisfy the count above but could not be relied on to satisfy this order"* — accurate as
  written, but it reads as though this test is the guard. The load-bearing guard is in another
  module and is named in neither the test nor D-3.7.3 nor the Delivery Log.
- **Suggested Resolution**: Add a sentence to the test's comment naming `TestMapRangeUnderModule`
  (`lint/internal/rules/maprange_test.go`) as the deterministic half, and note that this test's own
  ordering check is a probabilistic backstop over a two-element population. Optionally widen the
  fixture to 4-5 distinct uncovered runes, which drops the map-ranged pass rate to ~1/24.
- **Related AC**: AC7

---

### Finding 8: `Validate`'s doc comment names a test that does not exist

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/validate.go:43-45`
- **Observation**: *"`TestValidatePredictsRenderIncludingEmptyDataCaveat` (validate_test.go) pins
  this…"* — `rtk proxy grep -rn "TestValidatePredictsRenderIncludingEmptyDataCaveat" .` returns
  exactly one hit: this comment. The real test is `TestValidatePredictsRender`, and the caveat is
  its final subtest, `"D-3.7.1's trap: empty Data yields absent-path Errors…"`.
- **Impact**: The stale-name class D-000.37 requires be corrected in the same commit that touches
  the row — the very defect this story corrects for DW-10 (divergence D-4). Introducing a new one
  while fixing an old one is the same erosion at a different site.
- **Suggested Resolution**: Name `TestValidatePredictsRender` and its subtest.
- **Related AC**: AC2

---

### Finding 9: `absences.go`'s discharge comment cites the wrong file for the replacing assertion

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `lint/internal/rules/absences.go:29-31`
- **Observation**: *"see main_test.go's `TestRenderReadsSourceDateEpochFromEnvironment`"*. That test
  is declared in `folio-go/cmd/folio/main_subprocess_test.go:37`, not `main_test.go`. This is the
  comment that records D-000.59's discharge-by-replacement, so it is the pointer a future reader
  follows to verify the tripwire was replaced rather than deleted.
- **Impact**: D-000.37 class, at the highest-consequence site in AC13's work — the record of *what
  replaced the forcing function*.
- **Suggested Resolution**: Cite `main_subprocess_test.go`, and note it is a genuine subprocess
  (which is what makes it the valid replacement).
- **Related AC**: AC13

---

### Finding 10: `forbiddenimports_test.go`'s new comment names a nonexistent control and miscounts its own result

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `lint/internal/rules/forbiddenimports_test.go:313-315` and `:328-331`
- **Observation**: Two defects in one comment block. (1) *"`TestValidateFileHasNoForbiddenImports`'
  own control asserts…"* — no such test exists anywhere in the repo; the control is
  `TestFindRenderDeclaringFilesExcludesFolioGo`, declared 25 lines above. (2) *"Today this returns
  three paths (render_entry.go, validate.go, and folio.go only if it ever declared one of these
  names, which it does not)"* — it returns **two**; the sentence names three and then withdraws one.
- **Impact**: The comment documenting AC12's instrument misnames the control half of that instrument.
- **Suggested Resolution**: Name `TestFindRenderDeclaringFilesExcludesFolioGo`; say "two paths".
- **Related AC**: AC12

---

### Finding 11: AC8's test carries two contradictory times for the same epoch

- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/cmd/folio/main_subprocess_test.go:43-44` vs `:63-64`
- **Observation**: *"1750000000 -> 2025-06-15T16:26:40Z"* at `:43`, then *"epoch 1750000000 ==
  2025-06-15T15:06:40Z (verified against the standard library's own `time.Unix(1750000000,0).UTC()`)"*
  at `:63`. The assertion, and the truth, is `15:06:40` — confirmed live: the binary emits
  `D:20250615150640+00'00'`. The first comment is wrong by 80 minutes.
- **Impact**: In a date-conversion test, a wrong worked example in a comment is the thing a future
  reader trusts when re-deriving the expected string.
- **Suggested Resolution**: Correct `:43` to `15:06:40Z`.
- **Related AC**: AC8

---

### Finding 12: AC2's named mutation was not run against the agreement table it names

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/validate_test.go:104-208` (`TestValidatePredictsRender`); Delivery Log task-15 AC2 entry
- **Observation**: AC2's mutation is *"remove one validation step from `Validate` that `Render` still
  performs → **the agreement table** reddens on that row."* I ran it — removed the
  `resolveDocumentDate` call from `Validate` (`validate.go:60-62`):

  ```
  --- FAIL: TestDocumentDateInvalidIsLocatedErrorOnBothPaths
      date_test.go:71: Validate() error = <nil>, want a *RenderError
  ```

  `TestValidatePredictsRender` stayed green — it carries no `documentDate` row. The mutation does
  redden something, so the property is covered; but it is not covered by the table AC2 anchors on,
  and the Delivery Log reports the mutation as satisfying AC2.
- **Impact**: The agreement table is AC2's stated anchor and is one row short of the step that was
  chosen to test it. The next `Validate`-only step added will have no row either.
- **Suggested Resolution**: Add a `documentDate`-invalid row to `TestValidatePredictsRender`'s
  table so the anchor and the mutation coincide, or restate AC2's mutation as landing on
  `date_test.go`. Also note in the Delivery Log which test actually reddened.
- **Related AC**: AC2

---

### Finding 13: AC9's named mutation was not run; an argument was substituted for it

- **Severity**: Minor
- **Category**: Tests
- **Location**: Delivery Log, task-15 AC9 entry (story lines 977-983)
- **Observation**: AC9's mutation is *"make the date wiring conditional on anything only the CLI sets
  → **this AC reddens while AC8 stays green**. That discrimination is the point."* The Delivery Log
  instead runs a different mutation (`resolveDocumentDate` returns `(nil, nil)`), observes it reddens
  both, and concludes *"no mutation could be constructed that fires ONLY the CLI path … this
  absence-of-a-seam is itself the property AC9 requires."*
- **Impact**: The reasoning is sound — `Render` and the CLI do share one path with no CLI-only
  signal, which I confirmed by reading `runRender` and `resolveDocumentDate` — and AC9's *behaviour*
  is covered twice (in-process by `date_test.go`'s `TestRenderWithParamsDocumentDateSetsCreationAndModDate`,
  and CLI-side with `noEnv` by `main_test.go`'s `TestParamsSuppliedDocumentDateNoEnvironment`). But
  "a mutation reported is not a mutation verified" cuts both ways: an argument that a mutation is
  unconstructible is itself unverified. A constructible version does exist — e.g. have
  `injectDocumentDateFromEnv` write a private second key that only `resolveDocumentDate` honours.
- **Suggested Resolution**: Either run that constructible form and record the observed split
  (AC9 red, AC8 green), or state plainly in the Delivery Log that AC9's mutation was **not run** and
  why, rather than presenting the substitute as discharging it.
- **Related AC**: AC9

---

### Finding 14: `ParseRFC3339` accepts year `0000`, which the calendar contract excludes and PDF cannot represent

- **Severity**: Minor
- **Category**: Correctness
- **Location**: `folio-go/internal/expr/formatdate.go:165-192` (`ParseRFC3339`), `:118-138` (`validateCivilRanges`)
- **Observation**: R5 records the calendar as *"bounded to years 1–9999"*, but no lower bound is
  enforced: `validateCivilRanges` checks month, then day by calendar round-trip, then h/m/s — a
  year-0 date round-trips fine. Measured on the built binary:

  ```
  $ echo '{"documentDate":"0000-01-01T00:00:00Z"}' > pd.json
  $ ./folio render -data data.json -params pd.json t.folio | grep -o "/CreationDate ([^)]*)"
  /CreationDate (D:00000101000000+00'00')      exit 0
  ```

  (Also confirmed correct rejections: `2026-02-31` → *"day 31 is out of range for 2026-02"*;
  a non-string value → *"must be an RFC 3339 timestamp string, got number"*; `+14:00` and
  fractional seconds accepted and truncated as documented.)
- **Impact**: A caller-supplied value produces a PDF date string outside ISO 32000-1's calendar,
  which some readers reject. Low reachability (requires an explicit year-0 parameter), but it is on
  the byte-producing path and `ParseRFC3339` is now exported for reuse.
- **Suggested Resolution**: Add `if year < 1 || year > 9999` to `validateCivilRanges` (it is the
  shared validator, so `formatDate` gains the same bound), with a test owning both edges.
- **Related AC**: AC10

---

### Finding 15: AC11's fourth case includes a tautological comparison

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/cmd/folio/main_subprocess_test.go:141-151`
- **Observation**: The final assertion compares the subprocess run's stdout against
  `mainRunForTest([]string{"render", "-data", dataPath, tplPath}, ..., noEnv)`. Both invocations
  parse the identical argv, load the identical files, and reach `injectDocumentDateFromEnv` with no
  epoch — the same code path with the same inputs. It cannot fail unless the renderer is
  nondeterministic, which `first`/`second` two lines above already tests.
- **Impact**: Reads as a byte-identity claim against a no-date baseline; is a second determinism
  check. The load-bearing assertion in this test is the `/CreationDate`/`/ModDate`/`/Info` absence
  check at `:136-138`, which is real.
- **Suggested Resolution**: Either drop it, or make it non-tautological by comparing against bytes
  produced through a genuinely different route — e.g. `folio.Render(tpl, data, folio.Params(nil), …)`
  in-process, which exercises `decodeParams`' nil short-circuit rather than the `"{}"` path.
- **Related AC**: AC11

---

### Finding 16: `--strict`'s positive arm asserts only "not zero"

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/cmd/folio/main_test.go:214-220`
- **Observation**: `if code == exitOK { t.Fatalf(...) }`. D-3.7.4 assigns `1` to *"`--strict` with
  warnings"* and reserves `2` for usage errors; exit 2 would satisfy this assertion. Verified live
  that the real exit code is 1 for both `-strict` and `--strict`, so the behaviour is right and only
  the assertion is loose.
- **Impact**: A regression that turned a strict failure into a usage error — the exact confusion
  exit 2 exists to prevent — would pass.
- **Suggested Resolution**: `if code != exitFailure`.
- **Related AC**: AC5

---

### Nits

1. **`cliMissingGlyphTemplateJSON` is byte-identical to `cliWellFormedTemplateJSON`**
   (`cmd/folio/main_test.go:19-57`). Two constants, one value; only the *data* differs between the
   arms. Correct for AC5's "same document" requirement, but the second constant's name asserts a
   difference that does not exist. Collapse to one, or rename to say what it is.
2. **`TestStrictFlag`'s `"mutation control"` subtest duplicates its own first subtest**
   (`main_test.go:222-229` vs `:199-212`) — same argv, same expectation, no additional discrimination.
3. **`printDiagnostics` maps every non-`SeverityError` value to `"WARNING"`**
   (`cmd/folio/main.go:96-99`). Post-DW-18, `severityUnset` is a distinct zero value; it would print
   as a Warning rather than as the bug it indicates. `default:` naming the numeric value would be
   louder.
4. **`appendPDFDate`'s doc comment cites the wrong ISO clause form** (`internal/pdf/infodate.go:29-31`):
   it says ISO 32000-1 §7.9.4 *requires* `D:YYYYMMDDHHmmSSOHH'mm'`. ISO 32000-1 drops the trailing
   apostrophe (`OHH'mm`); the trailing form is PDF Reference 1.6-and-earlier and is universally
   accepted by readers. No behaviour change wanted — just cite the clause accurately, since AD-7
   names ISO 32000-1 specifically.
5. **`min` is redeclared at package scope** in `cmd/folio/main_test.go:301-306`, shadowing the Go 1.21
   builtin for the whole test build. Harmless; deletable.
6. **`render_arch_test.go:262` builds `offenders` by ranging `info.calls`**, a map — test-only, and it
   affects only the failure message's ordering, but it is a map range inside the file whose subject
   is determinism.
7. **`folio -h` and `folio render -h` exit 2 printing `flag: help requested`** before the usage block
   (`cmd/folio/main.go:60-66`, `:113-117`). Within AC4's letter ("usage error"), but a help request
   exiting 2 with a library error string is rough for the CI-triage audience D-3.7.4 names.
8. **RFC 3339 leap seconds are rejected**: `2026-01-01T23:59:60Z` →
   *"second 60 is out of range 0-59"*. RFC 3339 permits `60` and `time.Parse(time.RFC3339, …)`
   accepts it. **Pre-existing** — `validateCivilRanges` is unchanged by this story and is shared with
   Story 3.4's `formatDate` — recorded here only because `ParseRFC3339` newly exports the behaviour
   under a name that promises RFC 3339 conformance.

---

## AC-by-AC disposition

| AC | verdict | evidence |
|---|---|---|
| **AC1** | satisfied in behaviour; **guard weaker than claimed** | direct-call and plain-helper mutations RED; **method and func-var routes GREEN** — Finding 2 |
| **AC2** | satisfied | doc comment carries D-3.7.1's words verbatim (`validate.go:36-45`); no second entry point added; empty-`Data` row pins `Render` failing identically. Findings 8, 12 |
| **AC3** | satisfied | `cmd/folio/main.go` declares `package main` + `func main`; usage names both subcommands, exit 2 (verified live and by `TestCLIBinaryIsNamedCmdFolio`, which stats the literal path) |
| **AC4** | **NOT satisfied** | trailing-stdout mutation GREEN across all of `cmd/folio` — Finding 1 |
| **AC5** | satisfied | measured live: no flag → exit 0 with complete bytes + Warning on stderr; `-strict` → 1; `--strict` → 1; missing-glyph document is the asserted case. Finding 16 |
| **AC6** | satisfied | measured live on real stderr: `WARNING TEXT_MISSING_GLYPH element=e1: … U+0E01 …`; negative control (clean render, empty stderr) asserted and observed |
| **AC7** | satisfied | both required mutations RED independently; map-ranged-with-right-count implementation caught deterministically by `lint`. Finding 7 |
| **AC8** | satisfied | subprocess with real env: `D:20250615150640+00'00'` in both `/CreationDate` and `/ModDate`, read off the bytes |
| **AC9** | satisfied | params-only path emits the same keys with no env; Finding 13 on the mutation |
| **AC10** | satisfied | both paths return `*RenderError` with `DOCUMENT_DATE_INVALID`; bypass mutation reddens both. Finding 14 |
| **AC11** | satisfied | `render_test.go` **untouched** vs `c1c977f` (`git diff --stat` empty) — unweakened, all three cases, full forbidden list; fourth case present; live probe shows genuine key **absence**, not empty or epoch-zero. Finding 15 |
| **AC12** | both mutation halves verified; **residual open** | (a) RED naming `validate.go:5`/`:6`; (b) `folio.go` control GREEN; fourth-entry-point gap — Finding 4 |
| **AC13** | satisfied | kind constant, `absenceCheckKind`, `scopeRelDir`/`forbidden`, `scanForbiddenContent`, `ContentFilesScanned` and **both** `== 0` fatals all removed; kind **not** kept "for later"; two fixtures deleted and the `want` row dropped; renamed to `TestAbsencesChecksIncludeTheRemainingEntry` pinned to one id; forward comment carries the multiple-witness rule and keeps `ChecksEvaluated`'s extinction with Story 5.1; `TestAbsencesZeroWitnessIsCaught` untouched; DW-10 recorded discharged with all three parts. Finding 9 |
| **AC14** | property holds; **record defective** | 7 re-render goldens PASS; recorded command inert — Finding 3. Forward-property half is correctly stated as a construction argument with its reason, never "we checked" |
| **AC15** | satisfied | 905/1/1 in 17 pkgs; P6 stats byte-identical; `lint` 112/0; `hashmatrix` 3/0; `MANIFEST.md` correctly not regenerated |

**Also verified clean:** plain-terms opener explains the CLI/library asymmetry in prose with no
identifiers and says what the switch is for; the two defective decision-log sites (`:149`, `:1524`)
were corrected by *removing the 3.7 prediction and appending a D-000.49 note*, and
`emit_source_test.go` is untouched — the guard was **not** widened; no `time` import anywhere under
`internal/`; `SOURCE_DATE_EPOCH` appears in production source only in `cmd/folio/main.go`; the File
List matches the working tree exactly, with no undisclosed artifact; `lint/MANIFEST.md` unchanged
and `TestManifestUpToDate` green (stdlib-only CLI, module graph unchanged); DW-17 amended not
retired, with the behaviour precedent recorded for 5.12 and 6.6; DW-4's ledger carries
`documentDate`.

**Anything that could move bytes:** nothing on the no-date path — the `predictDocument` extraction is
a 0-line body diff, `infoID` is never reserved when `date == nil`, and the coalescing touches only
`diags`. On the with-date path the object numbering, xref and `/ID` move by design. The two
byte-relevant risks this review surfaces are **Finding 1** (a future stdout summary silently
corrupting piped output, unguarded) and **Finding 5** (the environment silently displacing an
explicit `documentDate`, changing both the metadata and any footer that prints it).

---

## Finding Resolutions

**Status of the two parked majors:** Findings 5 and 6 were held for the engineering lead's ruling.
The ruling arrived (recorded canonically as **[[D-3.7.6]]** and **[[D-3.7.7]]** in the decision log)
mid-finish, together with two further instructions that **overrule** this finisher's own initial plan
for Findings 2 and 4 (originally "state the residual honestly" for both) — **both guards must FAIL
CLOSED instead**, recorded as **[[D-3.7.9]]**. All six are FIXED below; none remain deferred to a
follow-up story.

### Finding 1 — Blocker — stdout-purity test only pinned the PREFIX
**Decision: FIX.** The mandated permanent test now asserts the WHOLE stream: `stdout` must be
byte-for-byte identical to the same document rendered to a file via `-o` (an equality the test owns on
both sides), plus an explicit `%%EOF\n` suffix check. **Red-proved with the review's own append
mutation** — `fmt.Fprintf(stdout, "\nfolio: wrote %d bytes\n", len(res.Bytes))` after the write in
`runRender` — which now reddens `TestExitCodesAndStreamDiscipline`'s mutation subtest specifically
(measured; reverted; `/usr/bin/diff` confirmed the file byte-identical to its pre-mutation state).
- **Files:** `folio-go/cmd/folio/main_test.go`.

### Finding 2 — Major — AC1's structural guard does not hold for a method or a func-valued variable
**Decision: FIX, overruling this finisher's own first-pass plan.** This finisher initially planned to
correct only the doc comment's overclaim (D-000.15's honest-labelling precedent). The engineering
lead reviewed `cmd/folio/main.go` mid-finish and ruled the walker must **fail closed** instead
([[D-3.7.9]] part (a)): `buildFolioCallGraph` now makes THREE kinds of top-level callable a graph
node — a receiverless function (unchanged), a top-level `var v = func(...){...}` (closes the
func-typed-variable route), and a METHOD keyed by name alone, merged across every receiver type
declaring that name (closes the method-value route, and interface dispatch for free, since only the
selector NAME is consulted, never `x`'s type). This is a safe over-approximation — a spurious edge only
makes the guard stricter. **Both of the review's own red-proof mutations now redden
`TestValidateNeverReachesRenderOrInternalPDF`**, measured live: the method mutation
(`sneakyRenderer{}.emit()`) reports `"reaches internal/pdf through [.emit]"`; the func-var mutation
(`sneakyEmit()`) reports `"reaches internal/pdf through [sneakyEmit]"`. Zero call sites required
allowlisting, so the lead's pre-stated ">10 sites, STOP" fallback did not fire. The doc comment's
"under ANY name" claim is now true by construction. One residual remains and is deferred, not hidden:
precise receiver-type resolution needs `go/types`; recorded as **DW-20**, due before the
`folio-go/v0.1.0` tag.
- **Files:** `folio-go/render_arch_test.go`.

### Finding 3 — Major — AC14's recorded measurement cannot detect a moved byte
**Decision: FIX.** The Delivery Log's `**AC14, measured:**` block is re-recorded above against the
render-and-compare command (`go test -run 'Golden|GoldenFixture|MatchesTheCommittedArtifact|Fallback' -v .`,
15 tests, all PASS), which — unlike the original record — DOES call a renderer and DOES fail under the
one-byte trailer mutation the review used (`append(b.body, ' ')` → `append(b.body, ' ', ' ')` in
`builder.finish`; 7 goldens reddened). The three digest-agreement tests are kept, re-labelled as a
second, narrower, still-true claim ("do the declared digests agree with the committed artifacts") per
the decision log's **D-000.9 extension** already recorded for this incident.
- **Files:** this story file's Delivery Log (AC14 block).

### Finding 4 — Major — the widened import fence leaves a fourth public entry point uncovered
**Decision: FIX, overruling this finisher's own first-pass plan.** As with Finding 2, this finisher
initially planned to state the residual honestly rather than close it. The engineering lead ruled
([[D-3.7.9]] part (b)): do **not** widen `pureEntryPointNames` further — **invert the population**.
`lint/internal/rules/forbiddenimports_test.go` now scans every non-test `.go` file directly under
`folio-go/` **except** a test-owned exception map, `allowedWorldReadingFiles` (today: `{"folio.go":
"declares LoadTemplate..."}`), each entry carrying its reason. **Red-proved**: adding the review's own
`preview.go` (`package folio; import "os"; func Preview(...) { os.Getenv(...) }`) now fails
`TestRenderEntryFileHasNoForbiddenImports` with **zero changes to the test file** — the population
change alone closes the gap, exactly the "fenced the day it is created" property Finding 4 asked for.
`TestFindRenderDeclaringFilesExcludesFolioGo` is kept as the non-firing control, now also asserting
`allowedWorldReadingFiles` still names `folio.go`.
- **Files:** `lint/internal/rules/forbiddenimports_test.go`.

### Finding 5 — Major (was parked) — `SOURCE_DATE_EPOCH` silently overwrites a caller-supplied `documentDate`
**Decision: FIX**, per the engineering lead's ruling, recorded canonically as **[[D-3.7.6]]**.
`injectDocumentDateFromEnv` is now FILL-ONLY: it returns params unchanged the moment `documentDate` is
already present in the decoded object — **presence, not truthiness** (an explicit `null` is left alone
too, and the engine correctly rejects it as a wrong-typed value on both subcommands). A new
subcommand-parameterised parity table (`TestSourceDateEpochFillsAbsentButNeverOverwritesSupplied`,
`cmd/folio/subcommand_parity_test.go`) covers (params-has-date × env absent/valid/malformed) against
BOTH subcommands. **Red-proved**: restoring the old unconditional `obj["documentDate"] = encodedDate`
reddens all three "params documentDate SUPPLIED" cells, on both `validate` and `render`. The
precedence is documented in `-h`/usage text. No diagnostic is raised for the ignored environment value
— the property `SOURCE_DATE_EPOCH` protects (reproducibility) is already satisfied by the caller's own
explicit value.
- **Files:** `folio-go/cmd/folio/main.go`, `folio-go/cmd/folio/subcommand_parity_test.go`.

### Finding 6 — Major (was parked) — `folio validate` never reads `SOURCE_DATE_EPOCH`
**Decision: FIX**, per the same ruling, [[D-3.7.6]] part 2. Both subcommands now call one shared
input-assembly step, `resolveInputs(templatePath, dataPath, paramsPath, getenv)`, called by both
`runValidate` and `runRender` — nothing between flag parsing and the `Validate`/`Render` dispatch is
per-subcommand any more. `TestSourceDateEpochFillsAbsentButNeverOverwritesSupplied`'s table asserts
equal exit codes for `validate` and `render` on every cell, including the malformed-epoch row (now
exit 2 on both, per D-3.7.6 part 3, rather than 1). **A third defect was found in the same function
while fixing this one** — recorded as **[[D-3.7.7]]**: `injectDocumentDateFromEnv` used to decode
`-params` (and pre-empt the engine's own diagnostic on a decode failure) whenever the environment
variable was set, regardless of whether it had anything to inject. Fixed alongside: the function now
returns the caller's bytes unchanged, untouched by any parse attempt, whenever there is nothing to
inject, and never raises its own error for malformed params — `Validate`/`Render` does, over the
caller's own byte offsets. Guarded by
`TestMalformedParamsProduceIdenticalOutcomeRegardlessOfEnvironment`, asserting byte-identical stderr
across all four (subcommand × env set/unset) combinations on a malformed `-params` file.
- **Files:** `folio-go/cmd/folio/main.go`, `folio-go/cmd/folio/subcommand_parity_test.go`.

### Finding 7 — Minor — the guardrail's load-bearing half is credited to the wrong place
**Decision: FIX**, per the engineering lead's follow-up ruling, recorded as **[[D-3.7.8]]** (amending
[[D-3.7.3]]). Added `TestMissingGlyphWarningsCoalescingIsDeterministicAcrossRepeatedRuns`
(`missing_glyph_coalesce_test.go`): runs the coalescing 64 times in-process on the same input and
asserts every result matches the first — a property with discriminating power independent of
population size, unlike the existing ordering assertion. **Measured**: the map-ranged mutation reds
the new test on **10 of 10** runs, vs. **1 of 10** for the existing ordering assertion in this run (the
review reported ~5/10; both runs confirm it is a coin flip, not a guard). `lint`'s
`TestMapRangeUnderModule` is now named explicitly in the test's own comment, in the decision log's
[[D-3.7.3]] amendment, and here.
- **Files:** `folio-go/missing_glyph_coalesce_test.go`, decision log ([[D-3.7.3]] amendment, [[D-3.7.8]]).

### Finding 8 — Minor — `Validate`'s doc comment names a nonexistent test
**Decision: FIX.** Corrected to name `TestValidatePredictsRender`'s actual final subtest.
- **Files:** `folio-go/validate.go`.

### Finding 9 — Minor — `absences.go` cites the wrong file for the replacing assertion
**Decision: FIX.** Corrected `main_test.go` → `main_subprocess_test.go`, with a note that it is a
genuine subprocess (why it is the valid replacement).
- **Files:** `lint/internal/rules/absences.go`.

### Finding 10 — Minor — `forbiddenimports_test.go`'s comment names a nonexistent control and miscounts its result
**Decision: FIX, superseded by Finding 4's rewrite.** The entire comment block this finding pointed at
(naming `TestValidateFileHasNoForbiddenImports` and claiming "three paths") was replaced wholesale by
[[D-3.7.9]] part (b)'s inverted-fence design, whose new comments correctly name
`TestFindRenderDeclaringFilesExcludesFolioGo` and no longer count matched paths by name at all — the
population is now "everything minus an exception list," so the miscounted-paths framing does not
recur.
- **Files:** `lint/internal/rules/forbiddenimports_test.go`.

### Finding 11 — Minor — AC8's test carries two contradictory times for the same epoch
**Decision: FIX.** Corrected the `:43` comment from `16:26:40Z` to `15:06:40Z`, matching the assertion
and the live binary.
- **Files:** `folio-go/cmd/folio/main_subprocess_test.go`.

### Finding 12 — Minor — AC2's named mutation was not run against the agreement table it names
**Decision: FIX.** Added a `documentDate`-invalid row to `TestValidatePredictsRender`'s own table.
**Red-proved**: removing `Validate`'s `resolveDocumentDate` call now reddens this table directly
(`--- FAIL: TestValidatePredictsRender/a_present-but-invalid_documentDate...`), not only the separate
`date_test.go` assertion the original Delivery Log cited.
- **Files:** `folio-go/validate_test.go`.

### Finding 13 — Minor — AC9's named mutation was not run; an argument was substituted for it
**Decision: FIX, subsumed by Finding 5/6's guardrail.** The new parity table's "params documentDate
SUPPLIED, env ALSO valid but DIFFERENT" cell is exactly AC9's named mutation shape ("date wiring
conditional on anything only the CLI sets"), and it is a REAL, run mutation — restoring the
unconditional overwrite reddens it on both subcommands (see Finding 5's red-proof) — rather than the
"no mutation could be constructed" argument the original Delivery Log offered in its place.
- **Files:** `folio-go/cmd/folio/subcommand_parity_test.go`.

### Finding 14 — Minor — `ParseRFC3339` accepts year `0000`
**Decision: FIX.** Added `year < 1 || year > 9999` to `validateCivilRanges`, checked first (before
month), matching R5's own stated calendar bound. Added `TestDocumentDateYearBoundIsEnforced`
(`date_test.go`), pinning both edges: year `0000` is a located `DiagCodeDocumentDateInvalid` error;
years `0001` and `9999` are accepted and produce the expected formatted date. **Red-proved**: removing
the new check reddens the year-0000 subtest; the accepted-boundary subtests are unaffected either way
(confirming they test a different fact).
- **Files:** `folio-go/internal/expr/formatdate.go`, `folio-go/date_test.go`.

### Finding 15 — Minor — AC11's fourth case includes a tautological comparison
**Decision: FIX.** Replaced the comparison against `run()` with an empty `"{}"` params file (the exact
same code path as `first`/`second` above it) with a comparison against `folio.Render` called directly,
in-process, with `folio.Params(nil)` — a genuinely different route, exercising `decodeParams`' nil
short-circuit rather than the JSON `"{}"` path.
- **Files:** `folio-go/cmd/folio/main_subprocess_test.go`.

### Finding 16 — Minor — `--strict`'s positive arm asserts only "not zero"
**Decision: FIX.** Changed `if code == exitOK` to `if code != exitFailure`, so a regression that turned
a strict failure into a usage error (exit 2) — the exact confusion exit 2 exists to prevent — no longer
passes.
- **Files:** `folio-go/cmd/folio/main_test.go`.

### Nits
1. **Duplicate template constant.** **FIX** — `cliMissingGlyphTemplateJSON` removed; every use site
   shares `cliWellFormedTemplateJSON` (the missing-glyph document's distinguishing feature is its
   DATA, never its template).
2. **`TestStrictFlag`'s duplicate "mutation control" subtest.** **FIX (removed)** — same argv, same
   expectation as the subtest above it; no additional discrimination.
3. **`printDiagnostics` maps every non-Error severity to "WARNING".** **FIX** — added an explicit
   `default` branch naming the raw numeric severity value, so a `severityUnset` Diagnostic (a bug)
   no longer prints identically to a real Warning.
4. **`appendPDFDate`'s doc comment cites the wrong ISO clause form.** **FIX** — corrected the citation
   (ISO 32000-1 drops the trailing apostrophe; the trailing form is PDF Reference 1.6-and-earlier,
   universally accepted). No behaviour change — the trailing apostrophe is kept.
5. **`min` redeclared at package scope, shadowing the Go 1.21+ builtin.** **FIX (deleted)** — `go.mod`
   pins go 1.25.0.
6. **`render_arch_test.go` builds `offenders` by ranging a map.** **DISMISS.** Test-only diagnostic
   message construction, not on the render/determinism path AD-1 governs, and affects only the ORDER
   of names inside a `t.Errorf` string a human reads on failure — never a byte of program output.
   Consistent with D-3.7.8's own framing: AD-1 forbids map ranging where order **can reach an
   output**; a test failure message is not that output.
7. **`folio -h`/`folio render -h` print `flag: help requested` before the usage block.** **FIX** — the
   raw error line is now skipped specifically for `errors.Is(err, flag.ErrHelp)`; every other parse
   error (unknown flag, bad value) still prints it.
8. **RFC 3339 leap seconds are rejected.** **DEFER.** Pre-existing behaviour in `validateCivilRanges`,
   shared with Story 3.4's `formatDate`, untouched by this story except for the new year-bound check
   (Finding 14) added well above the seconds check. Out of this story's scope: accepting `:60` without
   breaking the exhaustive round-trip guarantee needs its own design (a leap second is not a normal
   calendar day and does not round-trip through `daysFromCivil`/`civilFromDays` the same way). **Owner:
   whoever next revisits `internal/expr`'s RFC 3339/calendar handling.** No DW entry raised for this
   nit specifically — flagged here for that owner to pick up if `ParseRFC3339`'s exported conformance
   claim needs to be exact.

---

## Finisher pass — three-module gate, re-measured after every fix above

```
cd folio-go   && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1 -v
  -> build ok, vet ok, gofmt clean
  -> 920 "--- PASS:" lines (incl. subtests), 1 "--- FAIL:" (TestCorpusMeetsP6ExerciseFloors, REQUIRED,
     stats {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7} — byte-identical to every prior
     measurement in this story), 1 "--- SKIP:", 17 packages
  -> +15 pass vs the review's 905, accounted for exactly by this pass's new/changed tests:
     TestSubcommandNamesMatchRunSwitch (+1), TestSourceDateEpochFillsAbsentButNeverOverwritesSupplied
     (+8, 1 parent + 7 cases), TestMalformedParamsProduceIdenticalOutcomeRegardlessOfEnvironment (+1),
     TestMissingGlyphWarningsCoalescingIsDeterministicAcrossRepeatedRuns (+1),
     TestDocumentDateYearBoundIsEnforced (+4, 1 parent + 3 cases), TestValidatePredictsRender's new
     documentDate-invalid subtest (+1), minus TestStrictFlag's removed duplicate subtest (-1) = +15
cd lint       && go build ./... && go vet ./... && $(go env GOROOT)/bin/gofmt -l . && go test ./... -count=1 -v
  -> 112 pass, 0 fail, 4 packages — UNCHANGED from the review's own measurement (the import-fence
     rewrite changed HOW the guard scans, not how many tests exist)
cd hashmatrix && go build ./... && go test ./... -count=1 -v
  -> 3 pass, 2 packages — unchanged
```

`lint/MANIFEST.md`: not regenerated — the module graph is unchanged (`cmd/folio` and every new test
file remain stdlib-only); `TestManifestUpToDate` is green within the 112.

**No golden hash moved.** The render-and-compare command from Finding 3's re-recorded AC14 measurement
was re-run one final time after every fix in this pass and is unchanged (15/15 PASS). Nothing in this
finisher pass touches the render path's byte-producing code except the two mutations captured live for
red-proofs and immediately reverted (`internal/pdf/builder.go`'s trailer, `render.go`'s coalescing
loop, `folio-go/validate.go`'s call-graph escape routes, `folio-go/cmd/folio/main.go`'s date
precedence and stdout-purity) — each verified byte-identical to its pre-mutation state with
`/usr/bin/diff` before moving on.

**Epic 3's boundary gate still inherits ONLY Story 3.5's three outstanding `hidden-image` legs** — this
finisher pass added no fixture, no `matrixDocuments` entry, and touched no cross-target byte.
