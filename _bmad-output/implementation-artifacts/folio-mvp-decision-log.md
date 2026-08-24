# Folio MVP — Decision Log

The audit trail for the `run-dev-cycle` delivery of Epics 1–6 (50 stories, `1.1` → `6.7`).

Read this if you weren't in the run and need to know *why* something is shaped the way it is.
Entries are append-only: a reversal is appended, never a rewrite.

- **Program:** Folio MVP v0.1 — report designer + byte-reproducible PDF engine
- **Run started:** 2026-08-23
- **Baseline commit:** `f2aa8c0` (planning artifacts only — the repo contains no code at run start)
- **Contract:** `_bmad-output/specs/spec-folio/SPEC.md` + companions; architecture spine
  `_bmad-output/planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md`
  (invariants `AD-1` … `AD-26`); story source `_bmad-output/planning-artifacts/epics.md`

---

## Lead Grounding

*Filed 2026-08-23 by the orchestrator, from the engineering lead's one-time grounding pass. The
lead read the spine (AD-1…AD-26 in full), `SPEC.md`, `acceptance.md`, `folio-format.md`,
`glossary.md`, the Epic List and all 50 stories, this log, and `sprint-status.yaml`, and confirmed
the repo held no code. The UX docs were deliberately left unread until Epic 5 opens.*

**There are no ADRs in this project.** The spine's `AD-N` invariants are the ADR equivalent and
live in one file. Decisions cite `AD-N` plus a SPEC clause or story AC — never an ADR path.

### The spine's shape

One dependency arrow matters more than the rest: **there is none from `internal/layout` to
`internal/pdf`**. That absence is what keeps PNG/SVG/HTML renderers possible later (AD-5), and it
is precisely the arrow a well-meaning commit will try to add. Otherwise the core is a strictly
forward staged pipeline over immutable values — bytes → `Document` → `BoundTree` → `PageModel` →
PDF bytes — with `internal/geom` as a leaf that imports nothing.

The determinism boundary is a **directory** boundary, not a discipline. Everything under
`internal/` is render path and may not touch `time`, `os`, `math/rand`, `net`, `math`
transcendentals, package-level mutable state, or map iteration that can reach an output byte
(AD-1). The shell — `folio`, `cmd/folio`, `wasm/`, the designer — is the only place the world
exists.

Invariants expected to do the most adjudicating work, in order:

- **AD-1 / AD-2 / AD-3 / AD-23** — integer millipoints, one number emitter, exact-decimal report
  data, no `float64` under `internal/` for geometry *or* data. These decide nearly every Epic 1–4
  shape question.
- **AD-14** — one `Diagnostic` type, a closed code registry, and its three pre-settled data cases:
  absent path = Error, explicit `null` = empty and *not* an error, wrong kind = Error and never a
  coercion.
- **AD-4** — two passes, and **no `page` namespace in the expression language, ever**.
- **AD-11** — explicit row alias, root never shadowed, aggregates always over the whole collection.
- **AD-24** — band-relative origin, exactly one coordinate flip, hidden means absent and displaces
  nothing, images fit-and-centre.
- **AD-9 / AD-10** — one canonical byte form, monotonic non-random ids.
- **AD-26** — MIT, nothing copyleft at any depth.

### Predicted decision points across the 50 stories

| # | Where | Lead's prior |
|---|---|---|
| 1 | **1.1 / 1.3** — greenfield setup; whether the AD-1 import lint applies to `_test.go` files (it must not ban `os` in tests that read repo-root fixtures, or AD-21 becomes impossible) | Lead's |
| 2 | **1.4** — the format schema freezes here. Number round-tripping (`36` must not return as `36.0`), `nextId` / base-36 id derivation, three-decimal-max enforcement | Lead's |
| 3 | **1.5 / 1.6** — text must be drawn before a text engine exists (shaping is 2.3, breaking is 2.4). `boxesandglue/textshape` enters the module graph at 1.5, *before* the licence check has a shaping story to justify it | Lead's |
| 4 | **2.7** — `{{page}}` / `{{pages}}` are late-bound slots recognised by the *template* layer; the expression parser must reject `page` as an identifier. The split lands in Epic 2, before the parser exists in 3.2, and is easy to implement in the wrong layer | Lead's |
| 5 | **4.5 / 6.5** — what a column footer aggregate actually aggregates over (see D-000.6) | Lead's, but schema-affecting |
| 6 | **4.8** — alternating row styling is named optional and "first to be cut". Cutting a stated deliverable under capacity pressure is **not** the lead's call | **Owner's** |
| 7 | **5.10 vs 6.3** — 5.10's AC already requires an author-supplied parameter document as the third preview input, but the parameter surface is Story 6.3. Sequencing wrinkle | Lead's |

### SPEC Open Questions this run will be forced to close

- **Which font faces ship, and under what licences** — largely settled by the spine's Stack (Noto
  Sans / Noto Sans Thai / Noto Sans SC, all OFL 1.1; IBM Plex for the designer; custom-font upload
  deferred). **Live gap:** AD-12 declares `ja` a supported locale while no Japanese face is named.
  Whether Noto Sans SC's coverage suffices must be *verified, not assumed*, before Story 2.2. If a
  face must be added, it breaks Story 5.4's already-fixed payload numbers ("CJK 7.4 MB") and the
  ~9 MB budget — **that combination is an owner call**, and the cheapest alternative arm is
  dropping `ja` from AD-12's closed set.
- **Build order (Q5, Q6)** — already closed by `epics.md` and D-000.1. Not reopened.
- **Where sample JSON lives / whether its path persists in `.folio`** — read as already settled
  against persistence (SPEC Assumptions; spine Deferred). Lead will rule "not persisted in MVP"
  unless a story AC demands otherwise.
- **Template ↔ library compatibility rules (NFR6)** — the spine defers the policy to pre-v1. The
  MVP rule is exactly Story 1.4's: a higher declared MAJOR is a load error.
- **Screen-reader model for the canvas (UX1)** — a SPEC non-goal; EXPERIENCE's behavioural floor
  binds.
- **How the author discovers missing glyphs before previewing (UX3)** — **no story covers
  "before".** 2.2 and 5.12 give a post-render diagnostic only. Lead will rule MVP = post-render
  warning; a pre-preview coverage scan is scope growth. If anyone pushes back, it becomes the
  owner's.
- **Undo across a save boundary (UX4)** — AD-15 gives engine-side history over committed commands
  and says nothing about persistence. Lead will rule history dies on reload.

### Conflicts found now, rather than at Story 4.3

1. **Font directory.** Story 2.2 says fonts live in `fonts/`; AD-8 requires `folio-go/fonts/` —
   *inside* the module, because `go:embed` cannot reach outside it and a repo-root `fonts/` would
   silently fail. The story wording is the loose one; **the spine governs.**
2. **Licence-check target path.** Story 1.3 says the `designer/` lockfile; the source tree says
   `folio-designer/`. Cosmetic, but the check must point at a file that exists.
3. **AD-1's allow-list vs AD-2/AD-23.** `Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod` are
   all `float64`-only functions, yet `float64` is banned under `internal/`. The allow-list
   describes what the *lint tolerates*, not an invitation to use floats.
4. **`fixtures/` placement.** The spine puts it at repo root deliberately — "read at test runtime,
   so every future SDK conforms against the same bytes." Go tests therefore reach it by relative
   path, not `go:embed`, which is what makes conflict 1 above (the lint's `_test.go` scope) load-bearing.

### Refresh — 2026-08-23 (second session)

*Filed by the orchestrator from the second engineering lead's re-grounding pass. This **appends to**
the first session's grounding above; that text still holds. The lead read this log in full,
`sprint-status.yaml`, the Story 1.1 file in full (ACs, red-proofs, Delivery Log, all 28 QA findings
and every Finding Resolution), the entire shipped `folio-go/` tree and `fixtures/`, `epics.md`
§Stories 1.2–1.8, and re-read spine AD-1, AD-3 (as amended), AD-7, AD-8, AD-9, AD-10, AD-21, AD-22,
AD-23, AD-24, AD-26, §Consistency Conventions, §Stack and §Source tree. The predecessor's summary of
the spine's shape was verified, not re-derived. UX docs remain deliberately unread until Epic 5.*

#### What changed since the first grounding

The first grounding was taken against an **empty repository** at `f2aa8c0`. There is now code:
commit `048999b`, Story 1.1, `done`. Twenty-one new files, four Go packages, 23 top-level tests +
19 subtests, all green. Re-verified live this session: `go build`, `go vet`, `gofmt -l` clean;
`go test ./... -count=1` green in all four packages; golden hash
`0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c` reproduces.

**One measurement worth having early.** The lead ran the suite under `GOOS=js GOARCH=wasm` with
Go's own `go_js_wasm_exec` runner. `TestRenderMatchesGoldenFixture` **passes** — the recorded golden
hash is already byte-identical between `darwin/arm64` and `js/wasm`. Two of AD-21's four targets
agree today, before Story 1.2 has written a line. That is the single most encouraging fact
available about R1.

#### Shipped code vs. what this log says was decided — four divergences

The shipped artifact is sound and the QA/finisher cycle was unusually thorough. But four things in
the tree do not match what is written here, and all four fire on a story already in the queue.

**(1) D-1.1.b's guardrail shipped at Story 1.1, not Story 1.3 — and wider than it was ruled.**
D-1.1.b's Guardrail paragraph scopes the rule *"Within `internal/pdf`"* and says explicitly:
*"A number formatted into a diagnostic message is not an output byte and is explicitly not
covered."* AD-3's amended Rule repeats that carve-out verbatim.

What shipped is `TestNumberFormattingIsConfinedToNumbersGo` in
`folio-go/internal/pdf/emit_source_test.go`, enforcing **two** rules — the ruled one inside
`internal/pdf`, plus a second over **every non-test `.go` file under `folio-go/`** banning
`fmt.Errorf`, `fmt.Sprintf`, `strconv.Itoa` and the rest module-wide.

- Fires at **Story 1.4**: its AC requires *"loading fails with an error naming the declared version
  and the supported version"*. The idiomatic `fmt.Errorf` in `internal/template` is currently a
  build-red.
- Fires hard at **Story 3.6** (`internal/diag`, whose whole job is formatting values into message
  text — the case AD-3 names as *not covered*) and **Story 3.7** (`cmd/folio`, a CLI that must print).
- The reviewer's own Finding 15 flagged that the file comment overstated the guard's reach; the
  finisher fixed the comment by *widening the guard to match it*, which is the wrong direction
  relative to the ruling.

The lead's read: a **scope defect, not a scope decision**, to be fixed before Story 1.4 opens.
Related open point: D-1.1.b also requires the guard to ship *"with a retained violating fixture."*
None exists — the finisher applied and reverted the mutation. A retained fixture cannot be a
compiled `.go` file inside `internal/pdf`, because the guard scans `_test.go` there too; Story 1.3
will need a `testdata/` non-compiled fixture the lint reads.

**(2) `TestModuleGraphHasNoThirdPartyDependencies` asserts the module graph has exactly ONE module.**
`folio-go/gomod_test.go` fails if `go list -m all` returns anything but the main module. AC2's
stated intent is AD-6 — *no third-party PDF writer* — and its text enumerates the writers it means.
The test generalised that to "no dependencies at all", true and cheap on day one. Spine §Stack
schedules `boxesandglue/textshape` v0.0.15 for shaping and subsetting, so **Story 1.5 breaks this
test by design.** It must become a denylist of PDF writers (plus Story 1.3's AD-26 licence check
over the whole graph), not a count. Flagged now so the 1.5 developer does not meet a red suite and
"fix" it by deleting the assertion — which would silently retire AD-6's only mechanical guard.

**(3) `geom.ScaleRound` now panics — and Story 1.5 will feed it untrusted input.**
The finisher's Blocker-3/Major-6 fix (correctly) added panics for `den == 0`, `den == math.MinInt64`,
`v*num` overflow, and the `den == -1 && v*num == MinInt64` division overflow. Sound as
programmer-error guards. But the spine's §Consistency Conventions state: *"Nothing in `internal/`
panics on malformed input — untrusted font and template bytes return diagnostics."* AD-2 makes
`ScaleRound` the module's only font-scaling function, and Story 1.5 scales by `unitsPerEm` read out
of a **font file the caller supplied**. A font with `unitsPerEm == 0` would panic inside the render
path. Not a defect today (nothing calls `ScaleRound` yet); it is a **guardrail for Story 1.5**,
where `internal/fontset` must validate any font-derived denominator and return an AD-14 diagnostic
*before* the call.

**(4) Two Story 1.1 tests cannot pass on the `js/wasm` arm — measured, not predicted.**
Under `GOOS=js GOARCH=wasm`: `TestModuleGraphHasNoThirdPartyDependencies` →
`exec: "go": executable file not found in $PATH`; `TestRenderIsByteIdenticalAcrossTwoProcesses` →
`pipe: not implemented on js`. Both compile fine (`go vet` for js/wasm is clean); they fail at
runtime. Story 1.2's wasm leg therefore **cannot** be "run `go test ./...` on four targets".

**Non-divergences confirmed.** D-1.1.a is implemented exactly, both literal lines asserted, reason
left as a `go.mod` comment. D-1.1.b's three function names, signatures and the
`appendIntPadded → appendInt` delegation are exact, and after the finisher's Finding 1/8 fixes
**every** number in the output — including the MediaBox's lower-left corner and every
object/generation number inside every indirect reference — genuinely routes through one of the
three. D-1.1.c is honoured, `Render()` is doc-marked `PROVISIONAL`. D-000.6 was executed correctly
on all three counts. The finisher's **DISMISS** of Finding 14's third instance — declining to edit
this log's own Lead Grounding section, on the grounds that the log is append-only and D-000.6's
scope is the spine plus the SPEC companions — is correct, and the lead endorses it as precedent:
**D-000.6 does not authorise editing this log.**

#### The immediate horizon — Stories 1.2 → 1.8

**Story 1.2** — decisions the lead expects *before* code: how the `js/wasm` arm produces bytes to
hash given divergence (4); whether the recorded hash stays **one** value all four targets must
reproduce (strong lean: yes — a per-target map would quietly concede that targets may legitimately
differ, gutting NFR1); where the retained FMA negative-test fixture lives, given it is in direct
tension with both the AD-1 lint and `internal/arch_test.go`'s `float64` AST guard;
`GOTOOLCHAIN=go1.26.0` on all four legs (non-optional, deferred here from D-1.1.a option (c)).
**If any target's hash actually differs, that is not a lead call — it is an owner escalation**,
because it falsifies the product's central promise.

**Story 1.3** — the retained-fixture placement; whether the AD-1 lint is a Go test or a standalone
`cmd/` tool (precedent leans test, but the licence check needs the whole module graph and
`folio-designer/`'s lockfile, which does not exist until Epic 5); the map-iteration check's shape;
D-000.5's recorded correction that the licence check targets **`folio-designer/`**, not `designer/`.
Divergence (1) is 1.3's natural home if 1.4 has not already forced it.

**Story 1.4** — the format schema freezes. D-1.4.1 is implemented verbatim, including the
`folio-format.md` amendment in the same commit. Further decisions expected on: number
round-tripping under AD-9 (`36` must not return as `36.0` — AD-23's `UseNumber` discipline arriving
a story early); `nextId` derivation and persistence (AD-10 — monotonic, never reused, never
renumbered on save); three-decimal-max enforcement on geometry; and the `+*Template` API step.
Divergence (1) fires here if not already fixed.

**Story 1.5** — the first dependency and the first untrusted input: divergence (2), divergence (3),
and AD-26 clearing `boxesandglue/textshape` at every depth — it enters the graph here, *before*
Story 2.3 gives shaping a story to justify it. Override story under D-000.4. The subset tag is six
letters from a hash of the sorted glyph-id set (AD-7); one subset per font per **document**, never
per page.

**Story 1.6** — AD-23's decode discipline lands. AD-14's three pre-settled data cases must not be
re-litigated: absent path = Error, explicit `null` = empty and *not* an error, wrong kind = Error
and never a coercion. The AC's *"has read no wall clock, time zone, locale, environment variable,
filesystem path, or network resource"* is exactly what Story 1.3's lint asserts — reuse it, do not
write a second checker.

**Story 1.7** — the API closes. Two things already ruled: the AC's four-argument shorthand loses to
AD-8's `FontSet`, and the README must state the `toolchain`-binds-only-the-main-module caveat. This
is where the medium-confidence argument-packaging question gets answered by writing the README.

**Story 1.8** — AD-9's asset encoding and AD-24's fit-and-centre rule are fully specified; mostly
mechanical. `/Width` and `/Height` are pixel counts → `appendInt`. Override story under D-000.4.

#### Open questions the lead carries (none yet due)

- **AD-12 lists `ja` as supported; no Japanese face is named** (spine §Stack ships Noto Sans, Noto
  Sans Thai, Noto Sans SC). To be *verified, not assumed*, before Story 2.2. If a face must be added
  it breaks Story 5.4's fixed "CJK 7.4 MB" payload numbers and the ~9 MB budget — **that combination
  is an owner call**, with "drop `ja` from AD-12's closed set" as the cheap arm.
- Sample-JSON path not persisted in `.folio`; template↔library compatibility = Story 1.4's
  higher-MAJOR-is-a-load-error rule; missing-glyph discovery is a post-render diagnostic in MVP;
  undo history dies on reload. All four ruled that way unless a story AC demands otherwise.
- Story **4.8** (alternating row styling, named "first to be cut") remains an **owner** call if
  capacity pressure ever reaches it.

### Refresh — 2026-08-24 (third session, mid-Story-2.6 handover)

*Appended by the third engineering lead. The two sections above still hold and are not rewritten.
The predecessor's transcript was lost at a session boundary while Story 2.6 was in development, so
this is a re-grounding **from the record**, per [[D-000.8]]. Read in full: this log's `## Lead
Grounding`, all 52 `D-000.*` standing rules (including every amendment, extension, sharpening and
correction), and every `D-2.4.*` → `D-2.6.*` entry; `deferred-work.md` DW-1…DW-16;
`sprint-status.yaml`. The spine, `SPEC.md` and `epics.md` were **not** re-read wholesale — the
predecessors' conclusions about them stand. Targeted verification only, listed below.*

**What I verified live rather than inherited** (each because the pending decision hinges on it):

| claim | verified | result |
|---|---|---|
| the public API has no non-fatal diagnostic channel | `folio-go/render_entry.go` | `Render(t, d, p, f) ([]byte, error)` and `RenderTo(w, t, d, p, f) error` — **confirmed, no warning path** |
| `internal/diag` does not exist yet | `ls folio-go/internal/` | 8 packages, **no `diag`** — it is Story 3.6's ([[DW-6]]) |
| Story 2.8's ACs | `epics.md:931-951` | clip **and** diagnostic **and** *"PDF bytes are still returned — clipping degrades output but does not fail the render"* — i.e. **2.8 is the story that must build the non-fatal channel** |
| Story 2.6's scope fence | the story file, §*Scope fence* | *"Not clipping, and not an overflow diagnostic… this story changes how many sheets exist, not what happens at a box edge"* |
| the shipped overflow path | `folio-go/internal/layout/paginate.go` | typed `OverflowError{ElementID, ItemHeight, ContentHeight, Kind}`, message carries three named remedies ([[D-000.37]]-clean) |

**One finding, and it is the pending decision** — [[D-2.6.1]] (amended) sub-question (ii) is
internally inconsistent with two clauses of the ruling it amends. Ruled and recorded separately as
[[D-2.6.5]]; the grounds are FR44's scope, not the developer's convenience, and the log entry names
all three.

**Two record-vs-reality notes, neither a defect:** `sprint-status.yaml` carries 2.6 as
`ready-for-dev` while it is in development (a status-file lag, not a divergence), and `epic-2:
backlog` while six of its stories are `done` — the latter is already flagged, not fixed, in Story
2.6's own *Flagged, not fixed* section, and the epic key is the gate's to flip.

**What the Epic 2 gate owes, carried forward unchanged:** matrix registration for 2.3/2.3a,
`fixtures/three-band-page/`, `matrix-document: multi-page` (2.6), the Thai **reading** sign-off
(`fixtures/shaped-text/expected.pdf`) and the Thai **break** sign-off
(`fixtures/expected-breaks/expected_breaks.json`) — five, per [[D-2.6.2]], bound to their own
artifacts per [[D-000.26]] (refined). `TestCorpusMeetsP6ExerciseFloors` **staying red is required**
by Story 2.4's AC5 ([[D-000.17]]) and is not to be "fixed".

---

## Standing decisions (set at run start, 2026-08-23)

### D-000.1 — Build order is numeric, driven explicitly by story number
**Orchestrator decision** (the epics doc already settles it; recorded so the reason survives).

**Verdict.** Stories are delivered `1.1 → 1.8, 2.1 → 2.8, 3.1 → 3.7, 4.1 → 4.8, 5.1 → 5.12,
6.1 → 6.7`, one at a time, each named explicitly to the story creator so no agent auto-picks
a different "next" story.

**Situation.** The BMAD sub-agents pick work by scanning `sprint-status.yaml` for the next
`backlog` / `ready-for-dev` key. That is safe only if numeric order and dependency order agree.
Here they do, but not by accident — `epics.md` §Epic List deliberately *re-sequenced* the source
plan to make them agree, and records two resolutions:

- **PRD Q5 — resolved.** The source plan (`docs/folio-mvp-plan.md`) put tables in Phase 3 and the
  expression engine in Phase 4. But a table's footer aggregate *is* `sum()` / `count()` / `avg()`
  — the very functions the expression engine provides. The plan's order was simply wrong, so
  expressions (Epic 3) now precede tables (Epic 4), and scalar binding moved earlier still into
  Epic 1.
- **PRD Q6 — resolved.** Engine-first stands (counter-metric C4: designer work must not start
  before the engine renders the golden report, which is the Epic 4 gate). But the designer is two
  epics, not one trailing phase, and the exact preview ships *inside* Epic 5 rather than after it
  — a canvas you cannot preview with means authoring blind, and would leave NFR2/NFR5 unvalidated
  until the final phase.

**Why this wins.** Numeric order carries the dependency order for free, which means the agents'
default auto-pick behaviour and the correct build order coincide — the cheapest possible way to
keep 50 sequential stories honest. The accepted cost is that a mid-run reordering would now
require re-checking the epics doc rather than just re-sorting a list.

**Consequences.** Every `bmad-story-creator` launch names its story number explicitly anyway, as
belt-and-braces. Epic 4 is a hard gate: no Epic 5 story starts until the golden report renders and
hashes match on all four targets (C4).

### D-000.2 — Owner answers at the terminal
**Owner decision.**

**Verdict.** Mid-run `OWNER-DECISION-NEEDED` escalations are asked inline via `AskUserQuestion`,
not over Telegram, for the whole run. `~/.claude/telegram.env` exists and was offered; the owner
chose the terminal.

**Consequences.** The run blocks on the owner's presence at each escalation. To keep that cost
low, routine decisions are ruled on by the engineering lead and only genuine forks reach the owner.

### D-000.3 — Run continuously; pause only at design decisions
**Owner decision.**

**Verdict.** No checkpoint pause after each story or each epic. The run proceeds through all 50
stories, stopping only when a decision genuinely needs the owner. Epic-boundary summaries are
still *reported* — they just don't block.

**Consequences.** The per-epic report is the owner's main observation point, so it must carry the
heavy-test result rather than deferring it to the end (see D-000.4).

### D-000.4 — Heavy tests run at each epic boundary, with a determinism override
**Owner decision** (cadence), **Orchestrator decision** (the override policy).

**Verdict.** The **cross-target hash matrix** (`linux/amd64` and `linux/arm64` under Docker,
`js/wasm` under Node, against the local `darwin/arm64`) and — from Epic 5 — the **designer e2e
suite** run once per epic, after that epic's last story is committed and before its epic key is
marked `done`. Unit tests, `go vet` / lint, and the build run on **every** story regardless, and
no story reaches `done` on a red unit suite.

**Override:** a story whose *own* deliverable is hash-shaped runs the full matrix in its own
story even under the per-epic cadence. That is, at minimum: **1.2** (the matrix harness *is* the
story), **1.5** (font subsetting byte-stability), **1.8** (image embedding), **2.4** (line
breaking feeds every measurement), **4.7** (the golden report + recorded hashes). More may be
added as the run reveals them; each override is logged with its reason in the story's Delivery Log.

**Situation.** The owner traded feedback latency for wall-clock: a Docker arm64 boot per story,
50 times, is the single largest time cost in this program. Per-epic collapses that to six runs.

**In simple terms.** Think of the hash matrix as weighing a parcel on four different scales that
must all read the same. Weighing after every item you put in tells you instantly which item broke
the balance. Weighing once per box tells you a box is wrong and you must unpack it to find out
which item. The owner chose per-box — but for a handful of items that are *known* to be heavy and
temperamental (the font subsetter, the image embedder), we weigh immediately anyway, because
those are exactly the items that would make unpacking worst.

**Options considered.** (a) *Every story* — safest, and defensible given R1 is rated **Critical**
and FMA contraction on arm64 is invisible everywhere except the hash; rejected by the owner on
wall-clock. (b) *End of run* — rejected: a divergence introduced in 1.5 would surface after 6.7,
spanning all 50 commits, which is the worst possible bisect on the project's highest-severity
risk. (c) *Per-epic + targeted override* — chosen.

**Why this wins.** It puts a matrix run at every point where the owner is already reading a
report, and it keeps immediate coverage on the specific stories most likely to break byte-identity
— which is most of the value of `every-story` at a fraction of its cost. The accepted downside is
real: a divergence from a non-overridden story is attributable only to its epic, and Epic 5 is
twelve stories wide.

**Consequences.** Each Delivery Log entry must name the suites it actually measured *and* name the
unrun ones explicitly ("integration/e2e written, not run — per-epic cadence, due at epic N close").
Unit counts from a previous story are never carried forward. The epic-boundary catch-up run is a
**gate**: `epic-<n>: done` is not written until it passes, and nothing downstream proceeds. If it
goes red, the owning story is bisected, reopened at `review`, and fixed as its own commit — history
is not rewritten. Developers must additionally prove new integration/e2e packages **compile** under
their build tag every story, so a deferred run never opens with a stale compile error.

**How we'd know it was wrong.** An epic-boundary matrix run that goes red and takes more than one
bisect step to attribute. That is the signal to move to `every-story` for the remaining epics.

---

### D-000.5 — Greenfield setup: module path, layout, licence, toolchain pin, CI provider
**Orchestrator decision**, on the engineering lead's recommendation. Routine — each part is either
already written down in the spine or has exactly one sensible answer. Recorded because Story 1.1
starts from an empty repository and these choices are cheap now and expensive at Story 1.2.

**Verdict.**

| Setting | Value | Determined by |
|---|---|---|
| Go module path | `github.com/panitw/folio/folio-go` | Spine §Deployment, Story 1.1 AC |
| Toolchain pin | `toolchain go1.26.0`, exact — **not** `go1.26.x`, and Go 1.27 is deliberately not adopted | AD-22; local toolchain verified `go1.26.0 darwin/arm64` |
| Directory layout | Exactly the spine's Source Tree. `folio-go/`, `folio-designer/` and `fixtures/` are siblings of the existing `_bmad/`, `_bmad-output/` and `docs/` | Spine §Source tree |
| `fixtures/` | **Repo root**, read by relative path at test runtime — never `go:embed`ed into the module | AD-21; spine's note that every future SDK conforms against the same bytes |
| `fonts/` | `folio-go/fonts/` — *inside* the module | AD-8 (see below) |
| Licence | `LICENSE` (MIT) at repo root, `Copyright (c) 2026 Panit Wechasil` | AD-26; SPEC constraint |
| cgo | `CGO_ENABLED=0` on every target | SPEC constraint ("no cgo, because the engine must compile to WebAssembly"); Story 1.2 AC |
| CI provider | **GitHub Actions**, workflows in `.github/workflows/` | See below |

**On the CI provider.** No planning document names one — the spine's CI row describes *what* the
matrix must cover and stops there. Rather than rule on taste, the orchestrator checked the repo and
found the answer already existed as a fact: `origin` is `https://github.com/panitw/folio.git`, `gh`
is authenticated as `panitw`, and **the repository is public**. Public visibility is the part that
matters, because AD-21's matrix needs a `darwin/arm64` runner and a `linux/arm64` runner, and on
GitHub Actions both (`macos-14`, `ubuntu-24.04-arm`) are free for public repositories and billed at
a multiplier for private ones. So the obvious choice is also the free one, and there is no
trade-off to take to the owner.

**On the font directory — a conflict resolved against the story text.** Story 2.2 says fonts live
in `fonts/`. AD-8 requires `folio-go/fonts/`. **The spine governs**, and the reason is a real
failure mode rather than a preference: `go:embed` cannot reach outside its own module, so a
repo-root `fonts/` directory would not fail loudly at build time — it would simply not embed, and
the first symptom would be a font-resolution failure at render. This is baked into the Story 2.2
creator prompt so the loose wording is never implemented.

Two further wording conflicts are resolved the same way and baked into their story prompts:
Story 1.3's licence check must point at **`folio-designer/`**'s lockfile (the tree has no
`designer/`), and AD-1's allow-listed numeric surface (`Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`,
`Abs`, `Mod`) describes **what the lint tolerates, not an invitation to use `float64`** — which
remains banned under `internal/` by AD-2 and AD-23.

**Consequences.** Story 1.1 creates the module, the layout, `LICENSE`, and the toolchain pin.
Story 1.2 writes the four-target workflow against GitHub Actions specifically. The AD-1 import lint
must **exempt `_test.go` files** from the `os` ban, or repo-root `fixtures/` becomes unreadable and
AD-21 unimplementable — see D-000.5a if that is refined later.

**How we'd know it was wrong.** A CI matrix that cannot get an arm64 runner, or a `go:embed`
directive that compiles but embeds nothing.

---

## Epic 1 decisions

### D-1.4.1 — A column footer aggregate names its numeric source explicitly, with a derivation for the common case
**Orchestrator decision**, on the engineering lead's ruling. Routed to the lead because it is
schema shape and the direction determines it; **not** escalated to the owner because no closed set
the SPEC actually guards is touched (the palette stays five, the expression language stays eight —
schema fields are not a counter-metric), and there is no security, money-path, or scope
consequence. Raised in the run's first hour rather than at Story 4.5 because the format schema
freezes at Story 1.4.

**Verdict.** `columns[]` gains two optional sibling fields alongside the existing `footer`:

| Field | Type | Job |
|---|---|---|
| `footer` | enum `"sum"` \| `"count"` \| `"avg"` | **unchanged** — names the *operation only*. Absent = no footer cell |
| `footerOf` | string | **new** — a bare root-relative dotted value path (`"transactions.amount"`). No `{{ }}`, no function call, no `[]`. The numeric source |
| `footerFormat` | string | **new** — a `formatNumber` pattern applied to the computed value. Legal with all three operations |

When `footer` is present and `footerOf` is omitted, it is **derived** from the column's own `bind`,
but only when `bind` is one of exactly two syntactic shapes:

1. a bare row-scoped path `{{<alias>.<rest>}}` → `footerOf` = `<collection>.<rest>`;
2. a single `formatNumber(<bare row-scoped path>, <pattern literal>)` call → `footerOf` =
   `<collection>.<rest>` from the first argument, **and** `footerFormat` defaults to `<pattern>`.

`<collection>` is the table's own `bind` with `[]` stripped. **Any other `bind` shape is a load
error** — never a guess.

`count` is special: it always counts the table's own bound collection, which `bind` already names.
`footerOf` present alongside `footer: "count"` is a **load error**, because storing it would be a
second source of truth against `bind` — exactly what AD-13 forbids.

**Situation.** `folio-format.md`'s own worked example (line 139, and its twin in the full example
at line 243) puts `"footer": "sum"` on a column whose `bind` is
`{{formatNumber(transaction.amount, "#,##0.00")}}`. Nothing in the contract says what the sum runs
over. Read literally it sums *formatted strings*, which is nonsense. The alternative — reaching
inside the expression at render time to re-derive a numeric path — is magic that works until
someone writes a `bind` it cannot see through, and then produces a wrong total silently.

**In simple terms.** A spreadsheet column shows `1,234.50`, `987.00`, `2,000.25`. You put a SUM at
the bottom. The question nobody wrote down is: does the SUM add the *numbers*, or does it add the
*text you can see*? Adding the text is obviously wrong, but a computer will happily try. And if you
let the computer look at the display formula to work out which number you meant, it does fine on
`=A1` and fine on `=FORMAT(A1)` — and then someone writes `=IF(A1>0, A1, 0)` and it has to either
guess or give up. This ruling makes it give up **loudly, at load time, before anything renders**,
and lets the author state the source explicitly when the display formula is too clever to read.

**Options considered.**
- *(a) Explicit `footerOf` + a derivation for the common shapes* — chosen.
- *(b) Make `footer` legal only when `bind` is a bare row-scoped path*, forcing formatting
  elsewhere. **Not available:** it would invalidate the canonical worked example *and* forbid a
  formatted sum footer, which the golden report's five-column table requires.
- *(c) Make `footer` an expression string.* Rejected: it would grow the expression language
  outside the closed eight-function table where counter-metric **C1** cannot see it, and it would
  defeat Story 6.5's three-way footer picker in the designer.
- *(d) Drop `footerFormat`, inheriting the pattern via rule 2 only.* Explicitly offered by the lead
  as the trimmable edge and **not taken**: it leaves a stated gap (a column with a non-derivable
  `bind` could have a footer only by naming `footerOf`, and that footer would render unformatted),
  and re-opens precisely the second schema bump this early ruling exists to avoid. One optional
  string at zero render cost is the cheaper side of that trade.

**Why this wins.** The derivation is a **decidable syntactic property of the parsed expression
tree** — no evaluation, no heuristic — so the common case costs the author nothing and the
uncommon case fails as a located error rather than a wrong number. It also fits the rule to the
contract rather than the contract to the rule: the canonical example stays valid exactly as
written, which matters because a wrong worked example in a canonical file gets copied by every
agent that reads it. The accepted cost is two extra optional schema fields and a validation rule
that must be kept in step with the expression parser.

**Consequences.**
- **`folio-format.md` must be amended in Story 1.4, not later.** The example is fine; the *prose*
  is incomplete. The `columns[].footer` row of the field table currently reads only "*Optional.*
  `sum` · `count` · `avg`. Computed over the whole collection, never per page (AD-11)". It must be
  extended, and rows added for `footerOf` and `footerFormat`, stating the derivation, the `count`
  prohibition, and the two diagnostic codes. Until that lands, every agent reading the contract
  between 1.4 and 4.5 infers the aggregation source itself, differently each time.
- **Serializer key order** (AD-9, byte-visible): `align`, `bind`, `footer`, `footerFormat`,
  `footerOf`, `id`, `label`, `width`. `footerFormat` sorts *before* `footerOf`.
- **Load-time validation**, in `internal/template`, so `folio.Validate` catches it in CI without
  rendering (Story 3.7): `footerOf` must be prefixed by the table's collection path + `.` — a
  footer may not aggregate a collection other than the one its own rows come from; `footerOf` or
  `footerFormat` present with no `footer` is an error.
- **Exactly two new diagnostic codes** in the closed registry (AD-14 is additive; registered in
  Story 3.6): `TABLE_FOOTER_SOURCE_UNRESOLVED` (underivable or out-of-collection source — carries
  the **column id**, which is why AD-10 gives columns their own ids) and
  `TABLE_FOOTER_SOURCE_FORBIDDEN` (`footerOf` with `count`; footer fields with no `footer`).
  **Mint no others:** a non-numeric value at `footerOf` is AD-14's existing wrong-kind Error, and
  `avg` over an empty collection is Story 3.3's existing diagnostic. Reuse both.
- **The footer must be computed by the same aggregate evaluation as `{{sum(...)}}` /
  `{{avg(...)}}` / `{{count(...)}}` in `internal/expr`.** Summation must **not** be reimplemented
  in `internal/layout` — that is exactly how the golden report's total drifts from a hand-computed
  one, and Story 4.5's AC requires the two match. `avg` uses Story 3.3's scale and
  round-half-to-even; no second rounding mode. Whole collection, never per page (AD-11) — per-page
  subtotals must not appear as a special case here.
- `footerFormat` is *data* handed to the **existing** `formatNumber` in `internal/expr`, not a
  second formatter, so the Consistency Convention against a second stringifier holds. Absent and
  underived, the footer renders unformatted through the one display-text function in
  `internal/bind`. Footer cell alignment **inherits the column's `align`** — no fourth field.

**How we'd know it was wrong.** Story 6.4/6.5 finding that the designer must offer a footer on a
column whose `bind` is an arbitrary expression — neither derivable shape — which the table editor
has no field to express. The fix then is a `footerOf` cell in the table-editor matrix, **not** a
schema change.

### D-1.3.1 — The AD-1 import lint exempts `_test.go` for `os` and friends, and nothing else
**Orchestrator decision**, on the engineering lead's ruling. Routine.

**Verdict.** Under `internal/`:
- **Non-test files** — the full AD-1 ban: `time`, `os`, `math/rand`, `net`, `math` transcendentals.
- **`_test.go` files** — permit `os`, `testing`, `path/filepath`, `embed`. **Still ban `time`,
  `math/rand`, `net`, and `math` transcendentals**, with no exemption.

**Situation.** AD-21 places `fixtures/` at the **repo root** deliberately, so that every future SDK
conforms against the same bytes. A repo-root directory is outside the Go module and therefore
unreachable by `go:embed` — a test can only reach it with `os` / `path/filepath`. A lint that bans
`os` everywhere under `internal/` would make AD-21 unimplementable.

**Why not a blanket test exemption.** It would be wrong in the other direction. A test that seeds
from the wall clock or randomises its input is exactly how a nondeterministic assertion enters a
suite whose entire job is asserting determinism. The ban on `time` and `math/rand` is *more*
important in tests here, not less.

**Consequences.** Story 1.3's "a deliberately violating fixture proves it fires" AC needs **two**
fixtures for this rule, not one: a non-test file importing `time` (must fail) **and** a `_test.go`
importing `time` (must also fail). One fixture proves only half the rule. The exemption must key on
the **`_test.go` filename suffix**, never a path-prefix skip over a directory — otherwise a
non-test file in a test-heavy package inherits the exemption silently.

**How we'd know it was wrong.** A determinism test that passes on one run and fails on the next,
traced to a clock or a random seed inside the suite.

### D-1.1.a — The `go` directive sits *below* the toolchain pin, or Go deletes the pin
**Orchestrator decision.** Measured by the story creator, then **independently re-verified by the
orchestrator** before acceptance. Refines D-000.5, which fixed the `toolchain` *value* and said
nothing about the `go` line.

**Verdict.** `folio-go/go.mod` declares:

```
go 1.25.0

toolchain go1.26.0
```

**not** `go 1.26.0` + `toolchain go1.26.0`. A test asserts the literal `toolchain go1.26.0` line is
present in `go.mod`, and its red-proof is the deletion itself. From Story 1.2, CI additionally sets
`GOTOOLCHAIN=go1.26.0` as belt-and-braces.

**Situation.** Go silently **strips** a `toolchain` directive that is not strictly greater than the
`go` directive. Both the creator and the orchestrator ran the same probe on the real local
toolchain, in separate scratch modules:

| `go.mod` as written | after `go mod tidy` |
|---|---|
| `go 1.26.0` + `toolchain go1.26.0` | **toolchain line deleted** |
| `go 1.25.0` + `toolchain go1.26.0` | toolchain line survives |

Local `go env GOTOOLCHAIN` is `auto`.

**In simple terms.** You write "use exactly version 1.26" on a note taped to the project, but the
build system quietly throws the note away because it thinks it's redundant with the version already
written on the cover. It looks fine — right up until someone installs a newer Go. Because
`GOTOOLCHAIN` is `auto`, their build then uses 1.27 without a word, `compress/flate` output shifts,
and every recorded golden hash downstream is wrong. The failure is invisible everywhere except in
the hash, which is precisely the class of failure AD-22 exists to prevent.

**Options considered.** (a) *Lower the `go` line to `1.25.0`* — chosen. (b) *Keep `go 1.26.0` and
pin via `GOTOOLCHAIN` in CI* — rejected: it moves the pin outside `go.mod`, which contradicts
AD-22's letter ("`go.mod` pins an exact `toolchain` directive"), and leaves every **local** build
unpinned, which is where the hazard actually lives. (c) *Both* — adopted from Story 1.2 onward,
once a CI configuration exists to hold the second half.

**Why this wins.** It is the only arm that makes AD-22's stated mechanism actually function. The
accepted cost is a `go` directive two minor versions behind the toolchain, which looks odd in
review and will invite a well-meaning "cleanup" PR — mitigated by the asserting test and its
red-proof, so that cleanup fails the build rather than un-pinning the project.

**Consequences.** The language-version floor is now 1.25, so no 1.26-only language feature may be
used in the module. Any future toolchain bump moves *both* lines and is a release event under
AD-22. Story 1.2's workflow sets `GOTOOLCHAIN` explicitly on all four targets.

**How we'd know it was wrong.** A `go.mod` in a later commit with no `toolchain` line, or a CI run
whose reported Go version is not 1.26.0.

### D-1.1.c — `Render` ships provisional at Story 1.1
**Orchestrator decision.** Routine.

**Verdict.** Story 1.1 exports `func Render() ([]byte, error)`, doc-marked
`PROVISIONAL — Story 1.1 only`. It takes no arguments because there is nothing yet to pass: no
template (1.4), no data (1.6), no writer or params (1.7).

**Why this is safe.** The module is at `0.0.0-dev` and unreleased, so no compatibility promise
exists to break. The signature changes at 1.4, 1.6 and 1.7 by design. Marking it provisional in the
doc comment stops a later agent treating the placeholder as a settled contract — the actual public
surface is fixed by Story 1.7's AC ("a load call and a render call, and nothing ceremonial").

**Consequence.** No golden fixture or README example may depend on the argument-less form.

### D-000.6 — A ruling that makes a canonical document wrong amends that document, in the story's own commit
**Owner decision.** Asked because this is governance with lasting consequence and it recurs: two
rulings in the run's first hour each found a clause in a canonical planning document that was
incomplete or literally unimplementable, and each wanted the document changed. Settled once here
rather than re-asked per story.

**Verdict.** When a ruling establishes that a canonical document states something false,
incomplete, or unimplementable, the **document is amended in place, in the same commit as the story
that implements the ruling**. Only the wrong part changes; the stated intent is left untouched. The
amendment is logged here, and the diff is reviewable in git like any other change. This applies to
`ARCHITECTURE-SPINE.md`, `folio-format.md`, and the other files in the SPEC's `companions:` list.

**The two amendments this settles**, both shipping with their stories:
- **AD-3's Rule paragraph** (`ARCHITECTURE-SPINE.md`) — ships with **Story 1.1**. See D-1.1.b.
- **The `columns[].footer` field-table row** (`folio-format.md`) — ships with **Story 1.4**. See
  D-1.4.1.

**Options considered.** (a) *Amend in place with the story* — chosen by the owner. (b) *Decision log
only, documents never touched* — rejected: the documents would keep stating things that are false,
and, as the lead put it, every future agent then re-derives the fix independently and the next one
may not flag it. Story 1.1's creator caught AD-3's over-reach; that catch was not guaranteed.
(c) *Amend the format spec but treat the spine as immutable.* (d) *Draft amendments and batch them
for owner approval at epic boundaries* — rejected: code would ship against un-amended documents in
the meantime, which is the drift this decision exists to prevent.

**Why this wins.** It keeps exactly one canonical statement of each rule, and puts the amendment
under the same review as the code that depends on it. The accepted cost is real and worth stating:
the project's architecture document is now edited by agents over the course of 50 stories, so the
owner's guarantee is git history rather than immutability.

**Consequences.** Every such amendment must (1) change only the clause proven wrong, (2) ship in the
implementing story's commit, never a separate tidy-up, (3) be quoted verbatim in this log so the
before/after is readable without `git log`, and (4) leave the document's stated *intent* — a spine
invariant's **Binds** and **Prevents** lines — untouched. A ruling that would change an invariant's
intent is not an amendment; it is a direction change and goes to the owner.

**How we'd know it was wrong.** A spine diff that changes what an invariant is *for*, rather than
how it is stated.

### D-1.1.b — AD-3 governs numeric *representations*, not the number of functions
**Orchestrator decision**, on the engineering lead's ruling. Surfaced by the Story 1.1 creator as
`DECISION NEEDED` rather than worked around silently — the right call, and the reason it reached a
ruling at all.

**Verdict.** `internal/pdf` emits numbers in exactly **two representations**, via three unexported
functions in one file, `internal/pdf/numbers.go`. These names are fixed for the whole program and
are **not to be re-litigated per story**:

```go
func appendLength(dst []byte, v geom.Length) []byte          // decimal representation
func appendInt(dst []byte, v int64) []byte                   // integer representation
func appendIntPadded(dst []byte, v int64, width int) []byte  // appendInt + zero-fill; NOT a third representation
```

`appendIntPadded` must produce its digits **by delegating to `appendInt`** and left-padding, so
there is exactly one place integer digits are produced. Append-style (`[]byte` in, `[]byte` out) is
the ruled shape: allocation-free, table-testable, greppable.

**Situation.** AD-3's Rule ended "…and no other code in the module writes a number into a content
stream or object body." But `/Length`, `/Size`, `/Count`, object numbers and the 10-digit xref
offsets are numbers in an object body that are **not** `geom.Length` values, and routing them
through the geometric emitter is not awkward but *wrong* — an xref offset of 12345 **bytes** would
emit as `12.345`. Read literally, the clause is unimplementable.

**In simple terms.** A rule said "every measurement in this building must be written in metres, to
three decimal places." That is right for doorways and ceilings. It is nonsense for *the number of
floors* — "3.000 floors" is not more precise, it is a category error. The rule was written while
thinking only about measurements, and quietly bound counts too.

**Why the split is two, and why the reason matters more than the count.** AD-3's own **Prevents**
line names the hazard precisely: "`strconv.FormatFloat` or `fmt.Sprintf("%g")` reaching an output
stream … which reintroduces platform-visible float formatting." **Integers have no
platform-visible formatting** — `strconv.FormatInt(1234, 10)` is `"1234"` on every architecture,
target and toolchain. So counts and offsets were never inside the hazard AD-3 exists to close. The
final clause over-reached by binding to *any* number when the mechanism it protects is float-shaped
output. Fixing the reason, not just the count, is what makes the precedent hold for the eight later
stories that hit it.

**Rejected: "one function with two exported entry points."** Cosmetic compliance that makes things
worse — it puts two semantically incompatible behaviours behind one door, inviting exactly the
`12345 → "12.345"` bug, and it contradicts AD-3's own "unexported".

**The precedent, routed by representation and not by semantics.** This is the part that must not be
re-derived per story:

| Value | Route |
|---|---|
| Positions, advances, sizes, `MediaBox`, `Tm`/`Td`/`cm` operands, `/Tf` size, border widths | `appendLength` |
| `/Length`, `/Size`, `/Count`, `/N`, object and generation numbers, `/Rotate`, `/BitsPerComponent` | `appendInt` |
| xref offsets (10-digit), xref generation (5-digit) | `appendIntPadded` |
| Image `/Width`, `/Height` — pixel counts (Story 1.8) | `appendInt` |
| `/W` array widths, `/FontBBox`, `/Ascent`, `/Descent`, `/CapHeight`, `/StemV` — **1000-unit em space, not millipoints** (Story 1.5) | `appendInt` |
| Page-tree `/Count` (Story 2.6) | `appendInt` |

The em-space row is the trap the ruling exists to pre-empt: sending a 500-unit glyph width through
`appendLength` would emit `0.5`.

**The "neither" case, and its rule.** A value that is neither a `geom.Length` nor a plain integer is
**converted into one of the two by integer arithmetic before emission, never formatted on its own
terms.**
- **Colour components.** `#RRGGBB` gives 0–255; PDF wants 0.0–1.0. Convert to thousandths as
  `c*1000/255` with **round-half-to-even on the exact integer quotient** — the mode AD-2 already
  fixes for font scaling and AD-23 for `avg`, so no new rounding mode enters the module — then take
  the decimal route. Three decimals separates all 256 values (each 8-bit step is ≈3.92 thousandths),
  so none collide. No colour-specific formatter, and never a float.
- **`/ItalicAngle`** and other font-derived fractional scalars: convert to thousandths, decimal route.
- **Glyph ids under `Identity-H` take neither route** — they are a big-endian hex pair inside a
  string literal, i.e. a byte encoding like `/ID`, not a number. Same for the six-letter subset tag.
  This must be said in a code comment or a later agent will "unify" it.

**Reconciled with Story 1.1's design — colour does NOT arrive at 1.1.** The lead's ruling assumed
the hard-coded filled rectangle needs a fill colour. It does not: the story deliberately **emits no
colour operator**, relying on the PDF default fill of black, so the only content-stream numbers are
the four rectangle operands. The creator's stated reason — "adding `0 g` would be a number that is
not a `geom.Length`, i.e. an AD-3 carve-out for no benefit" — anticipated this exact tension and
sidestepped it before the ruling existed. **The creator's design stands.** The ruling makes the
no-colour choice a free one rather than a necessity, and the colour conversion rule becomes settled
precedent that is first *exercised* wherever colour actually arrives (Story 4.1's cell backgrounds,
borders and `altRowBackground`). Story 1.1 still exercises all three functions: `appendLength` for
`MediaBox` and the rectangle, `appendInt` for `/Length` `/Count` `/Size` and object numbers,
`appendIntPadded` for the xref.

**The guardrail — narrower and stronger than "no third formatter in the module".** A module-wide
assertion would fire on `internal/diag`, which legitimately formats numbers into diagnostic message
text ("row 47 exceeds the content height by 12.5pt"), forcing an exemption list that rots. Instead,
Story 1.3's lint gains one rule scoped to where output bytes are actually produced:

> Within `internal/pdf`, no file other than `numbers.go` may reference `strconv.Format*`,
> `strconv.Itoa`, `strconv.Append*`, or any `fmt` formatting call.

This is **complete**, because no package outside `internal/pdf` writes an output byte at all
(AD-5) — `internal/layout` may not import it, and `internal/pagemodel` knows nothing about PDF. Per
Story 1.3's AC it ships with a retained violating fixture: a second file in `internal/pdf` calling
`strconv.Itoa` must fail the build. **A number formatted into a diagnostic message is not an output
byte and is explicitly not covered.**

**Consequences.** AD-3's **Rule** paragraph is amended in `ARCHITECTURE-SPINE.md` with Story 1.1
under D-000.6; its **Binds** and **Prevents** lines are left untouched, because the hazard statement
was always correct and only the Rule over-reached. The amended text is quoted in full in the Story
1.1 Delivery Log.

**How we'd know it was wrong.** A PDF operator in a later story needing more than three decimals of
precision from a non-length value — which would mean the convert-to-thousandths rule under-serves
it. Not expected in MVP: AD-6 rules out transparency, shading and ICC.

### D-1.1.a (addendum) — assert both `go.mod` lines, and leave the reason in the file
**Orchestrator decision**, on the lead's refinement of D-1.1.a.

The `go.mod` test asserts **both** literal lines — `go 1.25.0` *and* `toolchain go1.26.0` — not the
toolchain line alone. The failure being guarded is someone "tidying" the `go` line up to `1.26.0`;
a test watching only the toolchain line passes right up until the next `go mod tidy` silently
deletes it, which is the original hazard wearing a disguise. A comment in `go.mod` itself carries
the reason, because comments survive `tidy` and the next agent to see a `go` line two versions
behind the toolchain will otherwise "fix" it.

**Recorded, not an objection:** a `toolchain` directive binds only the **main** module. An
integrator compiling `folio-go` into their own binary uses *their* toolchain and may get different
bytes. That is already AD-22's accepted position — it is precisely why golden hashes are recorded
per toolchain version — but it must be stated in the README when **Story 1.7** writes it, because
counter-metric C5's "first PDF in minutes" integrator is exactly the person who hits it.

### D-1.1.c (addendum) — the public API's target shape, fixed now so four stories converge
**Orchestrator decision**, on the lead's ruling. Story 1.1's `Render()` stays provisional; the
*destination* is fixed so 1.4, 1.5, 1.6 and 1.7 do not each invent one.

**Ruled target, reached at Story 1.7:**

```go
type Data []byte    // JSON report data
type Params []byte  // JSON parameters

func Render(t *Template, d Data, p Params, f FontSet) ([]byte, error)
func RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) error
```

Named `Data` / `Params` types rather than bare `[]byte`, because two adjacent same-typed `[]byte`
arguments are **silently swappable at the call site** — and in a product whose acceptance fixture is
a bank statement, that swap must be a compile error, not a support ticket. `[]byte` rather than
decoded values because AD-23 requires the library to own the `UseNumber`-preserving decode.

**Convergence — each story adds exactly one parameter and nothing else:** 1.1 `Render()`
*(provisional)* → 1.4 `+ *Template` → 1.5 `+ FontSet` (AD-8) → 1.6 `+ Data` → 1.7 `+ Params`,
`RenderTo`, README. Nothing binds until the `folio-go/v0.1.0` tag at the close of Epic 4; everything
before it is `0.0.0-dev`.

**Doc conflict recorded for Story 1.7.** Its AC writes the call as
`folio.RenderTo(w, template, data, params)` — omitting the `FontSet` that AD-8 requires `Render` to
take. **The spine governs; the AC is shorthand.** The 1.7 creator carries the five-argument form and
notes the AC's shorthand rather than treating it as a contradiction.

**Confidence, recorded honestly:** the lead rates the *argument packaging* medium-confidence. This
is the surface counter-metric C5 measures, and if writing the 1.7 README makes five positional
arguments read as ceremony, an options struct is the alternative — swapping before the v0.1.0 tag
costs nothing. The named-type and `[]byte` decisions are high confidence and should not move.

---

## Session 2 — 2026-08-23 (continuation)

### D-000.7 — The four standing decisions carry forward unchanged into session 2
**Owner decision.**

**Verdict.** The run resumes at Story 1.2 and continues to 6.7 under exactly the settings set at
run start: numeric build order driven explicitly by story number (D-000.1), owner answers at the
**terminal** (D-000.2), continuous run pausing only at genuine design decisions (D-000.3), and
heavy tests at each **epic boundary** with the per-story override for hash-shaped stories
(D-000.4). Nothing was renegotiated.

**Situation.** The first session ended after Story 1.1 was committed (`048999b`). Sub-agents do not
survive their session, so the engineering lead that held the run's direction is gone and a new one
was spawned. The standing decisions, by contrast, *are* durable — they live in this file. Re-asking
them would have invited the owner to answer differently from the record, silently splitting the
program's test evidence and build order across two regimes.

**In simple terms.** A relay runner handing over the baton also hands over the race plan. The
runner changes; the plan does not, unless the owner says it does. So the owner was asked one
question — "do these four still hold?" — rather than four fresh ones, and confirmed all four.

**Options considered.** (a) *Re-ask all four from scratch* — treats the log as advisory and risks a
different answer to a question already settled, which would make Delivery Log test evidence
incomparable across epics. (b) *Assume they hold and say nothing* — cheap, but the owner was
demonstrably present at the terminal, and a silent assumption about the **answer channel**
specifically would be unrecoverable if wrong (escalations would go somewhere unwatched).
(c) *One confirmation question* — chosen.

**Why this wins.** It costs one question and makes the carry-forward explicit in the record rather
than implicit in my context. The accepted cost is one blocking prompt at the start of a run the
owner asked to be continuous.

**Consequences.** All 49 remaining stories are validated under the D-000.4 regime, so a Delivery
Log entry naming per-epic deferral means the same thing in Epic 6 as it did in Epic 1. The new lead
is bound by every ruling above this line as precedent, not as suggestion.

### D-000.8 — The engineering lead re-grounds from this log, not from the spine
**Orchestrator decision** (routine — the run-dev-cycle continuation rule dictates it; recorded
because it changes what the lead is authoritative *about*).

**Verdict.** The session-2 lead was told to read this decision log **in full** — the predecessor's
`## Lead Grounding` section plus every ruling under it — the shipped Story 1.1 file, and the actual
Story 1.1 code, and to consult the spine only to verify and to load invariant text it needs to
cite. It appends a dated refresh to `## Lead Grounding` rather than rewriting it.

**Situation.** The predecessor grounded on an **empty repository** — that was the state at run
start. There is now shipped code, and code is where a decision and its implementation can quietly
disagree. Re-deriving the program from the spine would reproduce the predecessor's reasoning at
full cost while missing exactly the thing that has changed.

**Options considered.** (a) *Full re-grounding from the spine and epics* — pays the whole cost
again and, worse, produces a lead that rules from first principles while six Epic 1 decisions
already constrain the answers, so its rulings could contradict shipped code. (b) *No grounding,
rule on demand* — a lead that reads only what each decision touches never sees the invariants'
interactions, which is the entire reason the role is run-scoped. (c) *Re-ground from the record,
verify against the code* — chosen.

**Why this wins.** The log is a better grounding text than the spine for a continuation, because it
carries the spine's shape *plus* everything the run has since learned. The accepted cost is that
the lead inherits any error in this log as precedent; the mitigation is that it was explicitly
asked to hunt for divergence between what the log says was decided and what the code actually does,
and to report divergences as its highest-value finding.

**Consequences.** Rulings from here on cite `AD-N` + a SPEC clause or story AC, and must stay
consistent with D-000.1 … D-1.1.c. If the lead reports a divergence between the log and the shipped
code, that is logged as a correction here — the code is not silently treated as correct.

**How we'd know it was wrong.** A session-2 ruling that contradicts a session-1 ruling without
appending an explicit reversal.

---

## Epic 1 decisions (session 2)

### D-1.2.1 — All four targets render through Story 1.1's existing test-binary-as-renderer; the gate is the harness, not a green test run
**Orchestrator decision**, on the engineering lead's ruling. Measured before ruling, on this machine.

**Verdict.** The harness builds `folio-go`'s package test binary once per target with `go test -c`,
then runs it with `FOLIO_SUBPROCESS_RENDER=1` — the env-gated branch Story 1.1 already put in
`TestMain`, which writes `Render()`'s bytes to stdout and exits before `m.Run()`. `darwin/arm64`
builds and runs natively; `linux/amd64` and `linux/arm64` **cross-compile on the host** and execute
as static binaries inside a container carrying **no Go toolchain and no repo mount**; `js/wasm`
runs under Node via `$(go env GOROOT)/lib/wasm/go_js_wasm_exec`. No new `main` package, no `wasm/`
directory, no `cmd/`. The harness — not `go test` — collects four byte streams, hashes each, and
exits non-zero unless all four agree with each other **and** with the recorded golden.

**Situation.** The obvious implementation is "run the suite on each of the four targets", and it is
measurably wrong. Under `GOOS=js GOARCH=wasm`, two of Story 1.1's tests fail at *runtime* while
compiling cleanly: `TestModuleGraphHasNoThirdPartyDependencies` dies on
`exec: "go": executable file not found in $PATH`, and `TestRenderIsByteIdenticalAcrossTwoProcesses`
on `pipe: not implemented on js`. A wasm leg defined as "the suite passes" can therefore never go
green, and the natural repair — quietly narrowing what runs on that leg — is how a matrix ends up
proving nothing.

**In simple terms.** We need the same photograph taken by four different cameras, to check the
cameras agree. Running the whole test suite on each target is like asking each camera to also pass
its own factory self-test: informative, but not the comparison we want, and one of the cameras has
no self-test port. What we actually need is for each camera to take the one photograph and hand it
over. Story 1.1 already built that shutter release and used it for a different purpose, so we press
the same button four times.

**Options considered.** (a) *`//go:build js` split of the two offending tests* — **rejected**: a
build tag hides the file from `gofmt`, `go vet` and Story 1.3's AD-1 lint on that target, so removed
coverage becomes invisible rather than merely absent. A `SKIP` line with a reason is auditable; a
missing file is not. (b) *Per-target `-run` allowlist* — rejected as (a)'s problem in a slower
form: the allowlist is a second place the matrix's meaning can drift. (c) *A `wasm/` entry point
that renders and prints a hash* — rejected on scope: the spine's §Source tree reserves `wasm/` for
AD-15/AD-16's command channel, an Epic 5 concern, and creating it now is the empty-scaffold problem
Story 1.1 was deliberate about avoiding; Epic 5 would arrive at a directory someone else shaped.
(d) *Reuse the existing `FOLIO_SUBPROCESS_RENDER` seam* — chosen. The story creator and the lead
arrived at it independently, both after measuring.

**Why this wins.** It is the only option that adds no new surface at all: the seam already exists,
is already load-bearing for AC8's two-process test, and already checks its own short write. The
accepted cost is that the seam is now load-bearing for **two** things, so a change to its rendering
behaviour moves both — stated as a guardrail rather than discovered later.

**On cross-compiling rather than building inside each container** — this reads like a shortcut and
is the opposite. AD-22 pins the toolchain, so toolchain variance is a *deliberately excluded*
variable; AD-21's matrix exists to isolate **target codegen**, which is where FMA contraction lives.
Building all four with one pinned compiler tests exactly the variable AD-21 cares about and removes
container-Go-version drift as a confound. The harness must never fall back to `go run`, which would
reintroduce a toolchain into the container.

**Measured (lead, this session).** Native and wasm test binaries both emit **547 bytes** hashing to
`0f925e1b13702d34a30884bf85f3e3b2f2cb5312824267395871335fa6cb4f7c`, clean stderr on both.
Independently, the story creator measured **all four** targets producing that same hash. The
positive matrix already passes at `048999b`, so any red leg in this story is a defect, never harness
tuning.

**Consequences.** Each captured stream is independently asserted non-empty, `%PDF-1.7`-prefixed,
`%%EOF`-suffixed and page-count-1 **before** any comparison — lifting Story 1.1's
`assertWellFormedPDF`, not re-deriving it — so two empty captures can never compare equal. The two
wasm-incompatible tests take `if runtime.GOOS == "js" { t.Skip(...) }` with a message naming the
harness as what covers the gap; the skip keys on `runtime.GOOS` **specifically** and never on "an
exec call returned an error", or a genuine Linux-leg exec failure would silently pass. The
harness's four-row `target → sha256 → byte-length` table is pasted into the Delivery Log.

**How we'd know it was wrong.** A wasm leg that goes green while producing zero bytes, or a `SKIP`
whose reason no longer names a live covering mechanism.

### D-1.2.2 — One recorded hash for all four targets, never a per-target map; a real divergence is an owner escalation
**Orchestrator decision**, on the lead's ruling.

**Verdict.** `fixtures/minimal-rect/expected.json`'s `sha256` stays a **single 64-character
lower-case hex string** — at Story 1.2 and permanently. It never becomes an object, array, or
per-target map. All four targets reproduce that one value.

**Situation.** The moment a matrix leg goes red, the cheapest-looking repair is to record what each
target actually produced. That change is one line and it silently reverses the product's central
claim.

**In simple terms.** Four scales must all read the same weight. If one reads differently, you can
either find out why, or you can write "on scale 3, this parcel weighs 2.1 kg" on the wall and move
on. The second is faster, looks like documentation, and quietly means the parcel no longer has a
weight.

**Options considered.** (a) *Per-target hash map* — rejected: it encodes the proposition *targets
may legitimately differ*, which is the precise negation of NFR1 and of the risk (R1, rated
Critical) this story exists to retire. (b) *Majority vote — three targets outvote one* — rejected
for the same reason plus a worse one: it makes the failure quieter the more targets we add.
(c) *One golden, hard failure* — chosen.

**Why this wins.** It keeps the fixture's meaning identical to the property being claimed. The
accepted cost is that a legitimate output change (a deliberate format improvement) requires
re-recording the golden by hand — which is the correct amount of friction for changing a
reproducibility guarantee.

**Consequences.** On divergence the harness exits non-zero, prints the four-row table, and
distinguishes two diagnoses that look alike and are not: *all four agree but differ from the
golden* is a legitimate versioned change under AD-22, whereas *targets disagree with each other* is
NFR1 falsified. It writes each diverging target's raw bytes to a named file and reports the first
differing byte offset with a hex window, reusing Story 1.1's mismatch reporter in `render_test.go`
so every determinism failure in this project reads the same way. It **never** re-records the
fixture, majority-votes, or auto-selects the host's value. The fixture-shape check additionally
asserts `sha256` is a JSON string of exactly 64 lower-case hex characters — so a developer who
starts converting the field to a map meets a red *before* they reach a green, which is the only
reliable moment to stop them.

**Escalation rule, binding.** An actual cross-target hash divergence is **an owner decision, not a
lead call, and not a thing a developer resolves inside the story.** It falsifies the product's
central promise, and every available response — investigate, narrow the supported profile, or accept
a per-target reality — is a change of direction. If it happens the run stops and the owner is asked.

**How we'd know it was wrong.** `expected.json` growing a second hash field, under any name.

### D-1.2.3 — The retained FMA probe lives in a new repo-root module `hashmatrix/`, and its AC asserts the divergence *relation*, never recorded hash values
**Orchestrator decision**, on the lead's ruling, overriding the story creator's original placement.
This is the entry most likely to be re-litigated, so it is the most prescriptive.

**Verdict.** A new, separate Go module at repo root — `hashmatrix/`,
`module github.com/panitw/folio/hashmatrix`, zero dependencies, **no dependency on `folio-go`** (no
`replace`, no `go.work`) — holds both the harness driver and the retained floating-point
contraction probe. The probe computes `pos := x*scale + offset` in `float64`, takes its three
inputs from **`os.Args`** via `strconv.ParseFloat`, and emits `math.Float64bits(pos)` as 8 raw
big-endian bytes. The build fails if the four probe hashes are **all equal**, and specifically if
`darwin/arm64` and `linux/arm64` do **not** differ from `js/wasm`.

**Situation.** Story 1.2's third AC demands a *retained* deliberately-introduced floating-point
multiply-add proving the matrix can detect contraction. Two existing guards make that hard on
purpose: `internal/arch_test.go` parses every `.go` file under `folio-go/internal/` with `go/parser`
and flags any `float64` identifier anywhere in the AST — including `_test.go`, and including files
carrying `//go:build ignore`, because it parses rather than builds — and Story 1.3's AD-1 import
lint lands next with the same reach. The fixture and the guards are in direct tension by design.

**In simple terms.** The probe is the test strip you run through a smoke detector to prove the
detector still works. It has to contain real smoke, which is exactly what every alarm in the
building is built to reject. You do not solve that by teaching each alarm to ignore this one
canister — you keep the canister in a shed the alarms were never wired into.

**Options considered.** (a) *Under `internal/` with an exemption entry in each guard* — rejected:
exemption lists rot. The first person to add a second entry does it to unblock a build, and the
guard's meaning drifts from "no floats here" to "no floats except the ones we listed". (b)
*`fixtures/fma-probe/` as a nested module* (the story creator's original, and reasonable — it adds
no new top-level directory) — **rejected on what `fixtures/` means**: under AD-21 and D-000.5 that
directory is the bytes every future SDK conforms against, read by relative path at test runtime. A
future SDK vendoring `fixtures/` would pull in a module engineered to produce wrong answers. It also
forces a special case the creator flagged itself — Story 1.3's lint having to *exclude* `fixtures/`
— which is option (a)'s rot in a new location. (c) *A repo-root `hashmatrix/` module* — chosen.

**Why this wins.** The probe is out of both guards' scope **by construction** rather than by
maintained exemption, so there is nothing to erode; 1.3's guards bind `folio-go/internal/`
positively and never mention it. It also keeps `folio-go`'s module graph at exactly one module, so
the divergence recorded at Refresh item (2) does not bite a story early. The accepted cost is a new
repo-root directory, which D-000.5 had fixed to the spine's Source Tree — hence the amendment below.

**Two probe constraints that are not stylistic.** Inputs come from `os.Args` because with literal
constants the compiler folds the expression at build time with exact arithmetic and **every target
agrees** — the fixture would be silently vacuous, green forever, proving nothing. And the output is
raw `Float64bits` rather than a formatted decimal because a decimal rendering can round two
genuinely different values to the same text, masking the low-bit difference the probe exists to
expose.

**Assert the relation, not the values.** The four probe hashes are **not** recorded in any fixture.
A probe hash is a property of today's toolchain; recording it would turn a legitimate Go bump into
a spurious red on something that is not the product. The normative recorded value in this project
remains the *render* hash alone. The probe's contract is the inequality — and if the probe ever
stops diverging the build fails, reporting that the matrix's ability to detect contraction is now
unproven. A silently converged probe is a dead detector behind a green dashboard, which is the worst
outcome available here.

**Measured, independently, twice.** The lead: args `0.1 0.2 -0.02` give `3c40a3d70a3d70a4` on
`darwin/arm64` (fused, one rounding) and `3c50000000000000` on `js/wasm` (no FMA instruction, two
roundings). The story creator, separately: the split is **by architecture, not OS** — fused
`f77c0fc0acaa…d0df` on `darwin/arm64` + `linux/arm64`, unfused `049d6f48f044…caf8` on `linux/amd64`
+ `js/wasm` — and, decisively, rewriting the line as `float64(x*scale) + origin` (the explicit
rounding Go's spec says forbids fusion) drops `darwin/arm64` to the *unfused* hash byte-for-byte.
That control is what proves the divergence is contraction rather than platform detection, and it
belongs in the story file.

**Consequences.** A **D-000.6 amendment ships in Story 1.2's own commit**: the spine's §Source tree
gains `hashmatrix/` with a one-line comment stating it is the cross-target harness plus the retained
contraction probe, deliberately outside `folio-go` so the AD-1 lint and the `float64` guard exclude
it by construction. Only the tree block changes; no invariant's Binds or Prevents line is touched.
**Story 1.3 must be written knowing `hashmatrix/` exists by design** — a 1.3 developer who
"helpfully" widens the AD-1 lint or the `float64` guard to the whole repo turns the retained fixture
into a build break and will delete it. AD-26 still binds `hashmatrix/`: zero dependencies, and
1.3's licence check covers its module graph too — an empty graph is trivially clean, which is the
point.

**Confidence, recorded honestly.** Placement and the assert-the-relation shape are high confidence.
The specific input triple is **medium**: `0.1 / 0.2 / -0.02` diverges on this machine's toolchain
and the developer must **re-measure rather than copy** it. If a future toolchain stops fusing that
expression, the build fails and tells us so — that is the design working, not a defect.

**How we'd know it was wrong.** The probe going green-and-converged, or either 1.3 guard growing an
exemption entry naming `hashmatrix/` — which would mean the by-construction property was lost and
we are back to maintaining a list.

### D-1.2.4 — `GOTOOLCHAIN=go1.26.0` on all four legs, asserted from inside each binary before any hash is compared
**Orchestrator decision**, on the lead's ruling. This is D-1.1.a option (c)'s deferred half landing
exactly where D-1.1.a said it would.

**Verdict.** The **harness script itself** exports `GOTOOLCHAIN=go1.26.0` and `CGO_ENABLED=0` — not
only the workflow YAML. `folio-go`'s `TestMain` gains a second env-gated branch alongside the
existing one: `FOLIO_SUBPROCESS_TOOLCHAIN=1` → write `runtime.Version()` to stdout, exit 0. The
harness invokes all four binaries in that mode and requires each to report `go1.26.0` **and** to
equal `expected.json`'s `goToolchain`. That assertion runs **before** the hash comparison and has no
skip path.

**Situation.** AD-22 pins an exact toolchain and says CI uses that version only; AD-21 records
golden hashes against the exact toolchain version. Neither is self-enforcing: a container image or
a CI runner can carry its own Go, and the resulting build looks entirely normal.

**In simple terms.** The hash is only meaningful as "what *this* compiler produces". Comparing
hashes from two different compilers is not a weaker test — it is a meaningless one, and it fails in
the worst direction, because a green result from it actively tells you the thing is fine.

**Options considered.** (a) *Set `GOTOOLCHAIN` in the workflow YAML only* — rejected: a local run
would then be less pinned than a CI run, and D-000.4 makes the *local* run the epic-boundary gate,
so the gate would be measuring an unpinned build. (b) *Witness the toolchain with the host's
`go version`* — rejected: that reports what the host has, which equals what built the binary only
if `GOTOOLCHAIN` actually took effect, and says nothing at all about a container carrying its own
Go. Only the binary can report what built it. (c) *Assert from inside each binary, before hashing*
— chosen.

**Why this wins.** It closes the one gap that would make every other assertion in this story
worthless, for four extra process invocations. The accepted cost is that, because D-1.2.1
cross-compiles all four from one pinned host toolchain, the check is near-tautological in the happy
path — kept anyway, because it is the only thing standing between us and a runner quietly
substituting its own Go, which is precisely the failure D-1.1.a's "How we'd know it was wrong" line
names.

**Consequences.** Do **not** harmonise this with `TestRenderMatchesGoldenFixture`'s toolchain gate.
That test's fatal is a *contributor-local affordance* the Story 1.1 finisher deliberately reordered
(Finding 5) so structural checks still run on an off-toolchain machine. The harness has no such
affordance: off-toolchain means fail, full stop. **The asymmetry is deliberate and is stated in the
story file**, because someone will notice it and try to unify them.

**How we'd know it was wrong.** A CI run whose reported Go version is not `go1.26.0`, or a harness
that reaches its hash comparison on a machine where the toolchain assertion did not run.

### D-1.1.b (lesson) — A story AC that paraphrases a ruling is where the ruling gets silently widened
**Orchestrator observation**, verified in source, endorsed by the lead. Not a new ruling; the
amendment itself is deferred to Story 1.3 or 1.4.

**Finding.** D-1.1.b scoped its numeric-formatting guardrail *"Within `internal/pdf`"* and carved
out diagnostics explicitly: *"A number formatted into a diagnostic message is not an output byte
and is explicitly not covered."* AD-3's amended Rule repeats that carve-out verbatim. But **Story
1.1's own AC6 text** restated the rule as covering everything *"nowhere under `folio-go/` outside
`_test.go` files"* (story file lines 202, 327, 678), and
`folio-go/internal/pdf/emit_source_test.go` faithfully implements the AC — its comment at line 107
cites AC6's "broader wording" by name, and the walk at line 113 joins `root` + `"folio-go"`.
Verified independently by the orchestrator before acting.

**Why it matters.** This is **not** an implementation defect and nobody over-reached — the
paraphrase did. It is a canonical document contradicting a ruling, which is D-000.6 territory, so
the amendment must land on the **AC text** in `epics.md` and the story file, not merely on the
guard. It goes red at Story **1.4** (whose AC requires an error naming the declared and supported
versions — idiomatically `fmt.Errorf`), and hard at **3.6** (`internal/diag`, whose entire job is
formatting values into message text — the case AD-3 names as *not covered*) and **3.7** (a CLI that
must print).

**The generalizable lesson, for every story creator from here on.** A ruling's scope words and
carve-outs read like qualifiers, so a paraphrase drops them first — and the paraphrase is what gets
implemented, because it is what the developer is handed. **Story files quote rulings verbatim and
cite the decision ID; they do not restate them in their own words.**

**Also outstanding from the same ruling.** D-1.1.b requires the guard to ship *"with a retained
violating fixture: a second file in `internal/pdf` calling `strconv.Itoa` must fail the build."*
None exists — the Story 1.1 finisher applied the mutation and reverted it. Such a fixture cannot be
a compiled `.go` file inside `internal/pdf`, because the guard scans `_test.go` there too, so Story
1.3 needs a `testdata/` non-compiled fixture the lint reads as data. Note the shape rhymes exactly
with D-1.2.3 and should be ruled consistently with it.

**How we'd know it was wrong.** Story 1.4 opening with a red build on `fmt.Errorf` — which is now
predicted, dated, and owned rather than discovered.

### D-000.6 amendment (Story 1.2) — `hashmatrix/` added to the spine's §Source tree
**Ships in Story 1.2's own commit**, per D-000.6 and D-1.2.3's ruling that the cross-target harness
and its retained FMA-contraction probe live in a new repo-root module. Only the tree block changes;
no invariant's **Binds** or **Prevents** line is touched.

**Before** (`ARCHITECTURE-SPINE.md` §Source tree):
```text
  fixtures/                           # golden templates, data, params, recorded hashes (AD-21).
                                      #   Repo-level, not per-SDK: read at test runtime, so every
                                      #   future SDK conforms against the same bytes.
  folio-node/ · folio-java/ · …       # deferred SDKs, same fixtures, same namespace
```

**After:**
```text
  fixtures/                           # golden templates, data, params, recorded hashes (AD-21).
                                      #   Repo-level, not per-SDK: read at test runtime, so every
                                      #   future SDK conforms against the same bytes.
  hashmatrix/                         # module github.com/panitw/folio/hashmatrix — cross-target
                                      #   render/toolchain harness + retained FMA contraction probe;
                                      #   deliberately outside folio-go so AD-1 and the float64 AST
                                      #   guard exclude it by construction (D-000.6, Story 1.2)
  folio-node/ · folio-java/ · …       # deferred SDKs, same fixtures, same namespace
```

**How we'd know it was wrong.** A future story adding an exemption entry for `hashmatrix/` to the
AD-1 lint or the `float64` AST guard — the whole point of this placement is that neither guard
needs to know it exists.

### D-1.2.3 (amended) — `hashmatrix/` holds the probe alone; the matrix driver stays in `folio-go/matrix_test.go`
**Orchestrator decision**, on the engineering lead's ruling. **This amends D-1.2.3 above; that entry
stays as written.** The log is append-only — the Story 1.1 finisher's Finding-14 DISMISS established
that precedent, and it is honoured here, where it is the lead's own ruling being corrected.

**Verdict.** `hashmatrix/` contains the retained FMA contraction probe **and nothing else**. The
matrix driver stays at `folio-go/matrix_test.go` behind `//go:build matrix`. D-1.2.3's placement
clause — "holds both the harness driver and the retained contraction probe" — is wrong and is
withdrawn. **Every other clause of D-1.2.3 survives verbatim**, in particular: the **probe** never
moves into `folio-go`, under any build tag, for any reason.

**Where the original ruling went wrong.** D-1.2.3 bundled the driver into `hashmatrix/` for
*cohesion* — "the harness and its probe live together". But the property actually being protected
was narrower: **keeping a deliberate `float64` out of reach of guards that would flag it.** That
property belongs to the probe and was never about the driver. The placement rule swallowed a
component whose constraints had already been fixed elsewhere — in D-1.2.1 ("lift
`assertWellFormedPDF` rather than re-derive it") and D-1.2.2 ("reuse Story 1.1's mismatch
reporter"). Those two clauses and D-1.2.3's placement clause **cannot all be satisfied**: the driver
cannot both live in a zero-dependency module with no dependency on `folio-go` and reuse that
package's unexported test helpers.

**In simple terms.** The quarantine room exists to hold one contaminated sample. The original rule
also moved the technician who tests it into the room — which sounds tidier, until you notice the
technician needs instruments that are bolted down in the main lab. The sample stays quarantined; the
technician walks over. Nothing about the sample's isolation changes.

**Verified by the lead before ratifying, not inferred.**

| Claim | Result |
|---|---|
| The driver genuinely needs unexported `folio` test helpers | **True** — `repoRootFromTest` (4 sites), `assertWellFormedPDF` (line 447), `firstDivergence` (line 516), all unexported in `render_test.go`'s `package folio` scope |
| `hashmatrix/`'s zero-dependency property survives | **Intact** — `probe/main.go` imports only `encoding/binary`, `math`, `os`, `strconv`; no `require`, `replace` or `go.work` |
| The probe keeps by-construction guard exclusion | **Yes** — it is outside `folio-go` entirely; nothing changed for it |
| The driver needs that exclusion at all | **No** — `grep -c 'float64\|float32' folio-go/matrix_test.go` gives **0 matches** |

**Exposure check on the `matrix`-tagged file — all three guards checked individually, none fire.**
`internal/arch_test.go` roots its walk at `folio-go/internal/`; `matrix_test.go` is at the module
root, so it is not seen. `emit_source_test.go` walks all of `folio-go/` but returns early for any
`_test.go` outside `internal/pdf`; `matrix_test.go` is exactly that, so it is not seen. Story 1.3's
AD-1 lint binds `internal/` per AD-1's Rule ("every package under `internal/` is render path"); the
module root is outside it.

**Guardrail — the natural inference here is false, and Story 1.3's creator must be told.** The
`//go:build matrix` tag does **no guard-evasion work whatsoever**. Its only job is AC12: keeping
Docker and Node off the routine `go test ./...` path that runs every story under D-000.4. The safety
comes entirely from the file's **location**, not its tag — because `internal/arch_test.go` *parses*
rather than builds, so a build tag is invisible to it. Both of these would fire despite the tag: a
`matrix`-tagged file placed under `folio-go/internal/`, or a `float64` added to `matrix_test.go` if
it were ever moved under `internal/`. "The tag hides it" is the wrong mental model and it is the one
a reader will form.

**Options considered.** (a) *Amend the ruling, ratify what shipped* — chosen. (b) *Move the driver
into `hashmatrix/` as originally ruled* — rejected on both available implementations. Duplicating
`firstDivergence` creates a **second copy of the project's determinism diagnostic** — the one message
shape D-1.2.2 ruled every future divergence must be diagnosed with — maintained in two modules that
will drift, and drift *silently*, because both sides compile. Exporting the helpers instead is worse:
it promotes test-only scaffolding into `folio-go`'s public API, the exact surface counter-metric C5
measures and D-1.1.c deliberately fixed at four names. Neither is worth buying cohesion that was
only ever wanted for tidiness.

**Consequences.** A **D-000.6 amendment ships in this story's commit**: the spine's §Source tree
comment at line 619 currently says the module holds "cross-target render/toolchain harness +
retained FMA contraction probe" — the first half is now false, and only the comment changes.
`hashmatrix/README.md`'s opening paragraph must lead with what the module *contains* (the probe),
then say where the driver lives and why the split exists; its "why it is not inside `folio-go`"
section is correct and stays.

**How we'd know it was wrong.** The probe appearing anywhere under `folio-go/`, or either 1.3 guard
growing an exemption entry naming `hashmatrix/`.

### D-1.2.5 — A finding that disclaims measurement may not warrant shipped code
**Orchestrator decision**, on the lead's ruling. Standing guardrail; applies to every story from here.

**Verdict.** Story 1.2's `matrix_test.go:589` and `matrix.yml:15` justify two extra tests by stating
"F-6 measured" that GitHub-hosted macOS runners lack Docker. **F-6 itself says "Not measured in this
run… not load-bearing."** The two tests are kept — they are justified — but re-cited to **AC13** in
both the code comment and the workflow comment. F-6's own text is left **unchanged**: it is accurate,
and the fix is to stop citing it, not to retroactively promote it to a measurement.

**Situation.** This is D-1.1.b's paraphrase lesson recurring **inside the very story that logged
it** — the strongest available evidence that the lesson needs mechanical enforcement at review
rather than a line in a document.

**In simple terms, and this is the subtle part.** The underlying claim happens to be **true**:
GitHub-hosted macOS runners really do ship without Docker. That is exactly what makes it dangerous.
"True" and "measured in this run" are different claims, and a citation of the form "F-N measured"
asserts the second. A story that upgrades an explicitly self-disclaiming finding into a measurement,
in order to justify code that already shipped, has broken the property this project's culture rests
on: **that reported numbers were observed.** Nothing else in the log is verifiable if that slips.

**Consequences.** When code, a comment, or a workflow cites `F-N` or a Delivery Log measurement as
its justification, the finisher **verifies that `F-N` actually claims to have been measured**. A
finding that disclaims measurement may not warrant anything — the warrant must come from an AC, an
invariant, or a ruling. If shipped code genuinely depends on such a fact, the fact is **measured or
dropped**, never inherited from a finding that disclaims itself. Carried into Story 1.3's creator
prompt: 1.3 is built almost entirely out of assertions about what fires and what does not, making it
Epic 1's most exposed story to this failure.

**How we'd know it was wrong.** Any "measured" claim in a Delivery Log that cannot be reproduced by
running the command it names.

### D-1.2.6 — A conflict between two rulings is surfaced, never arbitrated in the diff
**Orchestrator decision**, on the lead's ruling. Standing process guardrail.

**Verdict.** The Story 1.2 developer met a direct contradiction between D-1.2.3's placement clause
and D-1.2.1/D-1.2.2's reuse clauses, and resolved it **silently in the code**. The resolution was
correct. The silence was not. Every sub-agent in this pipeline surfaces a ruling conflict as
`DECISION NEEDED` and parks; it never arbitrates one in the diff.

**Why this is worth a standing entry rather than a note.** Story 1.1's creator hit AD-3's
unimplementable clause and raised `DECISION NEEDED` — that reflex is the only reason D-1.1.b exists.
A developer who quietly resolves a ruling conflict when they are **right** is the same developer who
will quietly resolve one when they are **wrong**, and in that case nobody finds out. The cost of the
reflex is one message; the cost of its absence is unbounded and silent.

**Consequences.** Stated in every developer launch prompt from Story 1.3 onward, alongside the
existing "do not guess, do not terminate the story to ask" instruction. Carried to the finisher for
Story 1.2 as well, since the Delivery Log should record that the conflict existed and how it was
resolved — a future reader of D-1.2.3 and D-1.2.3 (amended) needs to see why the code led the log.

**How we'd know it was wrong.** A story whose diff satisfies one ruling and violates another, with
nothing in the Delivery Log saying a choice was made.

### D-1.2.7 — D-000.6 amendment 2 (Story 1.2 finisher): the spine's `hashmatrix/` comment corrected to match D-1.2.3 (amended)
**Finisher decision**, applying D-1.2.3 (amended)'s own stated consequence. The log is append-only;
D-1.2.3 and D-000.6 amendment (Story 1.2) above stay as written.

**Verdict.** `ARCHITECTURE-SPINE.md` §Source tree's `hashmatrix/` comment claimed the module holds
"cross-target render/toolchain harness + retained FMA contraction probe." The first half became false
the moment D-1.2.3 (amended) moved the driver back into `folio-go/matrix_test.go`. Only the comment
changes — no invariant's **Binds** or **Prevents** line is touched, matching D-000.6's amendment rule
and the code-reviewer's Finding 5.

**Before:**
```text
  hashmatrix/                         # module github.com/panitw/folio/hashmatrix — cross-target
                                      #   render/toolchain harness + retained FMA contraction probe;
                                      #   deliberately outside folio-go so AD-1 and the float64 AST
                                      #   guard exclude it by construction (D-000.6, Story 1.2)
```

**After:**
```text
  hashmatrix/                         # module github.com/panitw/folio/hashmatrix — holds the
                                      #   retained FMA contraction probe alone; the cross-target
                                      #   matrix driver stays in folio-go/matrix_test.go (D-1.2.3
                                      #   amended). Deliberately outside folio-go so AD-1 and the
                                      #   float64 AST guard exclude the probe by construction
                                      #   (D-000.6, Story 1.2)
```

`hashmatrix/README.md`'s opening paragraph was corrected the same way: it now leads with what the
module *contains* (the probe, and nothing else), then states where the driver lives and why the split
exists. Its "why it is not inside `folio-go`" section was already correct for the probe and is
unchanged.

**On the silent ruling conflict (D-1.2.6).** The Story 1.2 developer met a direct contradiction
between D-1.2.3's original placement clause ("holds both the harness driver and the retained
probe") and D-1.2.1/D-1.2.2's reuse clauses ("lift `assertWellFormedPDF` rather than re-derive it";
"reuse Story 1.1's mismatch reporter") — the driver cannot simultaneously live in a
zero-dependency module with no dependency on `folio-go` and call that package's unexported test
helpers. The developer resolved it by shipping the driver in `folio-go/matrix_test.go` and the probe
alone in `hashmatrix/`, which is the choice D-1.2.3 (amended) later ratified as correct. The developer
did not raise `DECISION NEEDED`; the story file's own "DECISION NEEDED: None" section and Task 5's
placement instruction show the conflict was resolved silently in the diff rather than surfaced. D-1.2.6
records the standing rule this violates. This entry closes the loop D-1.2.6 asked the finisher to
close: the code was right, the silence was not, and this is the record of why the code led the log.

**How we'd know it was wrong.** The spine or `hashmatrix/README.md` again describing a harness the
module doesn't contain.

## Epic 1 decisions — Story 1.3 (ruled before the story file was written)

*Sequencing note: on Story 1.2 the creator ran in parallel with the lead's rulings and needed a full
revision round. From 1.3 onward the rulings are obtained **first** and baked into the creator prompt.*

### D-1.3.2 — Narrow the numeric-formatting guard to `internal/pdf`; amend no canonical document
**Orchestrator decision**, on the lead's ruling. Closes Refresh divergence (1) and the D-1.1.b lesson.

**Verdict.** The guard's scope becomes exactly `folio-go/internal/pdf/`: every file except
`numbers.go`, `_test.go` files included (the shipped no-exemption choice is kept — stricter than
D-1.1.b required and costs nothing), and any directory named `testdata` excluded per D-1.3.3. **The
second, module-wide rule is deleted outright** — not kept with an exemption list.

**Situation.** `emit_source_test.go` banned `fmt.Errorf`/`fmt.Sprintf`/`strconv.Itoa` across all of
`folio-go/`, because Story 1.1's AC6 paraphrased D-1.1.b and dropped both its `internal/pdf` scope
and AD-3's diagnostic carve-out. Story 1.4's AC requires an error naming the declared and supported
versions — idiomatically `fmt.Errorf` in `internal/template` — so this was one story from going red.

**Why deleting the module-wide half is safe rather than merely convenient.** AD-3's amended Rule
carries the completeness argument: *"Because no package outside `internal/pdf` writes an output byte
at all (AD-5), the restriction needs to police only that package… Number formatting inside a
diagnostic message is not an output byte and is not covered."* AD-5 is load-bearing here — nothing
outside `internal/pdf` **can** emit a byte, so the module-wide half protects nothing while costing
the carve-out AD-3 states twice.

**A factual correction that shrank the fix, and the orchestrator's claim was the wrong one.** The
orchestrator told the lead the over-broad text lived in *both* `epics.md` and the Story 1.1 file.
The lead checked and it does not; the orchestrator then re-verified independently. `epics.md`'s only
`nowhere` is line 818 (AD-24's Y-flip, unrelated), and its Story 1.1 AC block at 435–445 was already
correctly narrowed — "every **geometric** number… routed through a separate integer writer in the
same file", explicitly citing D-1.1.b, with no module-wide clause. **The defect never propagated
into the document every later story's ACs are carried from.**

**Therefore D-000.6 does not fire.** The only document stating the over-broad rule is the
**completed, committed Story 1.1 file** (lines 202, 327, 678), and that file is a *record of what
was verified at `048999b`*, not a contract. Rewriting its AC6 would falsify the audit trail — it
would claim we verified something we did not. Instead a short **`Superseded` note** is appended to
that story file pointing at this ruling, leaving all three lines intact.

**Precedent established, generalising the Story 1.1 finisher's Finding-14 DISMISS: completed story
files are append-only, exactly like this log.** A shipped story records what was true when it
shipped; corrections attach to it, they do not overwrite it.

**Consequences.** The narrowing is red-proved **in both directions**, per D-1.3.1's two-fixture
precedent: `strconv.Itoa` in a non-`numbers.go` file inside `internal/pdf` must still fail, **and**
`fmt.Errorf` in `internal/template` must now pass. The second is the whole point and must be
demonstrated rather than assumed — it is what unblocks Story 1.4.

**How we'd know it was wrong.** Story 1.4 still meeting a red build on `fmt.Errorf`, or a diagnostic
message somewhere under `internal/` being contorted to avoid a formatter.

### D-1.3.3 — Every guard becomes a checker over a target path with two callers; fixtures live in `testdata/`
**Orchestrator decision**, on the lead's ruling. This is the architectural ruling of Story 1.3.

**Verdict.** Each guard is refactored into a **pure function taking a target directory and returning
findings**, with exactly two callers: a **production scan** over the real tree asserting **zero**
findings, and a **fixture scan** over a retained fixture tree asserting **exactly the named
findings**, by file and rule. Every scanner skips any directory named `testdata`. Fixtures live at
`folio-go/testdata/lint/<rule>/`.

**Why `testdata/` is not an exemption list.** It is Go's own tool-level convention — the `go` command
ignores `testdata/` universally — applied **by category**, so it cannot rot the way a per-file
allowlist does. And files under `testdata/` are never compiled, which is precisely the property a
*retained violating fixture* needs: it must contain deliberately-broken code without breaking the
build.

**Why this is consistent with D-1.2.3 despite reaching the opposite placement.** Both rulings say the
same thing — *place the artifact so the property holds by construction* — and the properties are
opposites, so the mechanisms are opposites. The **probe** must compile and execute on four targets,
so it must be a buildable package, so `testdata/` is impossible and it needs its own module. The
**lint fixture** must never compile and only ever be parsed, so `testdata/` is not merely acceptable
but *better than a module*, because non-compilation is the property we want and `testdata/`
guarantees it for free.

**Both polarities are required.** Each rule ships (a) a violating fixture that must be reported and
(b) a **compliant near-miss** that must **not** be. One fixture proves the guard fires; it does not
prove the guard is not over-broad — and an over-broad guard is exactly what produced D-1.3.2.

**This discharges DW-1**, and the assignment to 1.3 was not a parking spot but the *same seam*: "a
fixture-path override so red-proofs never touch the real golden" and "a checker that takes its target
path as a parameter" are one design. **Refinement:** 1.3 establishes the seam generally and applies
it to Story 1.2's AC16 check as its first instance, rather than patching AC16 one-off.

**Guardrail, measured.** Neither existing guard mentions `testdata` today, so a fixture added under
`internal/` would currently be parsed by the `float64` guard and redden the suite. **Both scanners
gain the `testdata/` skip and the two-caller shape before the first fixture is added**, or the first
fixture is a build break.

**How we'd know it was wrong.** A guard whose only evidence of firing is a mutation someone applied
and reverted, rather than a fixture standing in the tree.

### D-1.3.4 — The licence check ships its Go half now; the JS half is deferred as an *asserted absence*
**Orchestrator decision**, on the lead's ruling.

**Verdict.** Ship the Go module-graph check now — it is complete for everything that exists. Defer
the `folio-designer/` lockfile half to **Story 5.1**, the story that creates that directory.

**The deferral mechanism is the load-bearing part.** Story 1.3 ships an assertion that
`folio-designer/`'s lockfile is **absent**. The day Epic 5 creates it, that assertion goes **red** and
forces the JS half to be wired before the build can pass again.

**In simple terms.** A conditional "check it if it's there" starts silently passing the moment the
directory arrives, and nobody notices the half-check — the guard reports success precisely when it
stops covering anything. Asserting the absence inverts that: arrival trips the wire. Same treatment
for the AC's OFL 1.1 font text (fonts arrive at Story 2.2) and the Apache-2.0 `pdfjs-dist` NOTICE
(Epic 5) — assert absent now, so landing them breaks the build until the manifest covers them.

**On "release artifact", which the AC requires and this project cannot yet produce.** Split:
- **Ships at 1.3** — the manifest is **generated and asserted complete**: every module in the resolved
  graph appears with a resolved licence, and an **unknown or unresolvable licence fails the build**.
  That is the substance of AD-26 and is fully implementable today.
- **Deferred** — *publication* as an artifact attached to a release. No release process exists; the
  `folio-go/v0.1.0` tag is Epic 4's close. Registered in `deferred-work.md`, **owner: Epic 4 close**.

**Carried forward as an owner item, not a lead one.** No story currently owns the tagging event
itself. "Who cuts `folio-go/v0.1.0` and what ships with it" is a release-process decision with
licence-compliance consequences, and if it is still unowned when Epic 4 is planned it goes to the
owner. AD-26's "release artifact, not a README paragraph" is a commitment somebody has to keep.

**Confidence.** High on the Go half and the asserted-absence mechanism; **medium** on Story 5.1 as
the JS owner — if Epic 5's shape shifts, the owner moves with the story that first creates the
lockfile.

**How we'd know it was wrong.** `folio-designer/` arriving without the absence assertion going red.

### D-1.3.5 — Total ban on ranging a map under `internal/`; the escape hatch is stdlib and needs no exemption
**Orchestrator decision**, on the lead's ruling. The subtlest call in Story 1.3.

**Verdict.** **No `range` over a map value in any non-test file under `folio-go/internal/`.** No
reachability analysis. The build fails naming file, line, and the offending expression.

**Situation.** The AC reads *"when a `range` over a map **can reach an output byte**, the build fails
naming the location."* Read literally that is a whole-program dataflow analysis from every map range
to the PDF byte writer — undecidable in the general case, so in practice an approximation whose gaps
nobody could enumerate.

**In simple terms.** Map iteration order in Go is deliberately randomised, so a map range that
influences output makes the output non-reproducible — the exact property this project exists to
guarantee. Tracing *which* map ranges can reach the writer is the hard, unreliable problem. Banning
map ranges in the directory where output is produced is the easy, exact one. We lose the ability to
range a map in render-path code even when it would have been harmless; that is the price, paid
knowingly.

**Why this is AD-1's own shape rather than convenience.** The lead's predecessor put it exactly:
*"The determinism boundary is a **directory** boundary, not a discipline."* AD-1's architectural move
is to replace per-site judgement with a mechanical directory rule a reviewer cannot argue with. A
reachability analysis would reintroduce precisely the per-site judgement AD-1 abolished. The AC's
phrase "can reach an output byte" states the **hazard**, not the **mechanism** — the same relation
AD-3's `Prevents` line has to its `Rule`, which is the distinction D-1.1.b was built on. Implementing
a hazard statement literally is what made AD-3 unimplementable; the fix there was a conservative
mechanical rule, and it is the fix here.

**The escape hatch ships in 1.3, needs no new package, and cannot rot:**
`for _, k := range slices.Sorted(maps.Keys(m)) { v := m[k]; … }` — stdlib, available at the
`go 1.25.0` language floor. It ranges a **slice**, so the rule permits it *by construction*: no
exemption, no allowlist, no `//lint:ignore` comment. **The lint's failure message names this idiom
verbatim**, so the first developer to hit the rule is unblocked by the error text itself rather than
by finding a story file.

**Cost today: zero — measured.** There is no `range` over a map anywhere under `internal/` right now.
We are pre-committing before a caller exists, which is the cheapest possible moment to take a
restriction like this.

**Detection invariant.** Detection must be **exact and type-based**: flag an `*ast.RangeStmt` whose
subject resolves to a map type. A syntactic guess is **not** acceptable — a false positive on
legitimate code is exactly how a guard gets weakened by the first person it inconveniences.

**How we'd know it was wrong.** A legitimate render-path need to iterate a map that the sorted-keys
idiom cannot express, or a developer disabling the rule rather than using the idiom.

### D-1.3.6 — The new guards are a standalone tool in their own repo-root module; the two existing guards stay put
**Orchestrator decision**, on the lead's ruling.

**Verdict.** The **new** guards — the AD-1 import lint, the map-iteration check, and the licence
check + manifest — live in a standalone tool in its own repo-root Go module (`lint/`,
`module github.com/panitw/folio/lint`). Not a `go test` inside `folio-go`, and **not**
`folio-go/cmd/`. The two existing guards (`internal/arch_test.go`,
`internal/pdf/emit_source_test.go`) **stay where they are**.

**Why.** D-1.3.5 requires exact type-based map detection, whose clean implementation is
`golang.org/x/tools/go/packages` — a dependency `folio-go` must not carry, both for AD-6/AC2 and
because Story 1.1's one-module assertion stands until Story 1.5 legitimately changes it. The licence
check needs the whole Go module graph *and* a JS lockfile *and* must emit a manifest file: that is a
tool, not an assertion, and it is awkward and slow inside `go test`. And D-1.2.3 already established
the pattern — tooling that needs what `folio-go` must not have goes in its own repo-root module.
Ruling differently would give two similar problems two different shapes for no reason.

**D-1.2.1's "no new `cmd/` in `folio-go`" was scoped to Story 1.2 and does not bind here — but its
reason does:** `folio-go/cmd/folio` belongs to Story 3.7 and must not be created early. The lint
being a separate module means the question does not arise.

**The existing guards stay because moving them buys no property** — they are package-scoped, exact,
and working, and churn is not free. 1.3 **may** consolidate if it proves genuinely cheap, but must
not do so silently: consolidation is a `DECISION NEEDED`, not a judgement call in the diff (D-1.2.6).
Either way both gain D-1.3.3's `testdata/` skip and two-caller shape.

**Invariant over mechanism — stated deliberately, the lead having over-specified once already this
run (see D-1.2.3 amended).** What binds is: (a) exact detection with no false positives; (b) **no
dependency added to `folio-go`'s module graph**; (c) locally runnable, with CI merely invoking it,
per D-000.4. If the developer finds an exact, dependency-free path — stdlib `go/types` with a
tolerant error config — that is a **permitted refinement, not a deviation**. What is forbidden is
polluting `folio-go`'s graph or shipping a syntactic guess.

**Consequences.** A D-000.6 amendment to the spine's §Source tree for the new directory ships in
1.3's commit — one line, same shape as `hashmatrix/`'s. The new module's own dependency
(`x/tools`, BSD-3-Clause) must appear in the manifest the licence check generates; **the check
covering itself is correct**, and the story says so explicitly so it is not misread as a bootstrap
problem.

**Confidence.** High on placement; **medium** on `x/tools` specifically — which is why (b) is the
invariant rather than the library name.

### D-1.3.7 — Guard scope is stated positively; a build tag hides nothing
**Orchestrator decision**, on the lead's ruling. Wording confirmed verbatim for the story file.

**Verdict — the two paragraphs below go into Story 1.3's file as Dev Notes, verbatim.**

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

**Plus one concrete instruction specific to this story.** Story 1.3 is built almost entirely out of
claims about what fires and what does not, making it Epic 1's most exposed story to D-1.2.5. So:
**every "this guard fires on X" claim in the Delivery Log must cite the retained fixture that
demonstrates it**, never a mutation that was applied and reverted. That is what D-1.3.3's fixture
architecture is *for*, and it is the difference between 1.3 shipping guards and 1.3 shipping
assertions about guards.

### D-1.3.3 (amended) — A checker returns `(findings, error)`; the production caller fails on error *before* asserting zero
**Orchestrator decision**, on the lead's ruling. **Amends D-1.3.3 above; that entry stays as written.**
Found by the Story 1.3 creator (F-5), by probing rather than reading — credited here because it
corrects the lead's own ruling.

**Verdict.** A checker returns `(findings, error)`. **The production caller must fail on a non-nil
error, separately from and before asserting zero findings.** The fixture caller asserts findings **by
file and rule, never by count** — a scan that finds the right *number* of wrong things must fail. And
violating fixtures must be **valid syntax with forbidden semantics**, so a fixture can never redden a
guard by being unparseable rather than by being non-compliant.

**The hole in the original ruling.** D-1.3.3 said "checker returns findings; production caller asserts
zero". Both shipped scanners `return err` from their walk function **before appending any finding**, so
a scan that errors on an unparseable file returns **zero findings** — and a caller asserting zero
**passes**. As written, the ruling made a crashed scan indistinguishable from a clean one.

**In simple terms.** A smoke alarm that has lost power reports no smoke. The reading is identical to
the reading from a working alarm in a room that is genuinely fine, and that is the whole problem: the
guard is loudest exactly when it is healthy, and silent both when all is well and when it is dead.

**Why this belongs in the log rather than being absorbed as a fix.** It is the same vacuity class as
Story 1.2's two blockers (a counted-four assertion comparing a loop's output to its own input; a CI
path reporting agreement over three of four targets) and D-1.3.4's rejected conditional lockfile check.
**Absence silently reading as success** is this program's recurring failure mode — four instances in
two stories — and each was caught by a different agent probing rather than reasoning. Worth stating as
a standing suspicion: whenever a guard's healthy output and its dead output are the same value, that
is a defect regardless of what the code looks like.

**How we'd know it was wrong.** A guard whose test suite stays green after its scanner is made to
throw on every file.

### D-1.3.8 — The licence fixture is a stub module tree under `lint/testdata/licence/`; AD-26 was never in tension with it
**Orchestrator decision**, on the lead's ruling. Answers the creator's DN-1 (raised, not arbitrated,
per D-1.2.6).

**Verdict.** Option **(b)**: a stub module tree with `replace`, at `lint/testdata/licence/`. Supplement
with the classifier unit tests from option (c). **Do not adopt (a).**

**The conflict dissolves on inspection, and this is the load-bearing part.** AD-26 forbids a
**dependency carrying** a copyleft licence; its Prevents line names the mechanism — *"a copyleft
dependency arriving through a plausible-looking package and forcing a relicence after the fact… Go
links statically, so LGPL obligations would attach to the whole binary."* Copyleft obligations attach
to **code received under a licence**. They do not attach to a text file that says "GPL". A stub module
we authored, containing **zero lines of third-party code**, whose `LICENSE` file is fixture data
written to be classified, creates no licensee relationship, transfers no copyrighted work, and imposes
no obligation on anything.

**"A module whose LICENSE file declares GPL" is not "a GPL dependency."** The clauses collide only if
AD-26 is read as a rule about licence *strings* rather than about licensed *work* — and its Prevents
line settles that in the language of static linking and relicensing. So there is no trade here and no
owner escalation: retaining the fixture does not weaken AD-26 in any degree.

**Placement holds by construction, via two independent mechanisms** — the `go` command ignores
directories named `testdata` for package matching, and a directory containing its own `go.mod` is
outside its parent module's boundary regardless. Belt and braces, with no exemption entry anywhere:
the same shape as the source fixtures, for the same reason.

**Three fixture graphs, not one** (D-1.3.3's both-polarities requirement plus the unknown-licence rule):

| Fixture | Contains | Must |
|---|---|---|
| `copyleft/` | a stub module declaring GPL-3.0 (or LGPL/AGPL/SSPL) | **FAIL**, naming the module and the licence |
| `permissive/` | stub modules declaring MIT / Apache-2.0 / BSD-3-Clause only | **PASS** |
| `unknown/` | a stub module with **no** `LICENSE` file at all | **FAIL** as unresolvable |

**The third is the one most worth retaining.** A silent pass on an unidentifiable licence is the
realistic failure mode — far likelier than someone accidentally adding a GPL dependency — and it is
the case D-1.3.4 already ruled must fail the build.

**Guardrails.** Every `require` has a matching local `replace`, and the fixture graphs must resolve
with `GOPROXY=off` — proving the checker resolves **without fetching**, which is both a CI hermeticity
property and what makes a fake graph safe. **No full copyleft licence text in the repo**: include only
what the classifier needs (an SPDX identifier line or short marker). That is not legal caution — there
is no obligation, per above — but so that a human reader or a downstream scanner cannot mistake the
repo for containing GPL-licensed material; if the classifier's design *requires* full text, treat that
as a signal to key on SPDX identifiers plus a curated table instead. Make the fixtures obviously fake:
module paths under `example.test/`, a header comment in each stating it contains no third-party code,
and `lint/testdata/` marked `linguist-vendored` in `.gitattributes`.

**Why not (a), a recorded graph fixture.** It is a second representation of something production
derives live from `go list -m all`. When the two drift — a Go release changes the output shape, or
someone edits the recorded file — the guard passes against a reality that no longer exists and nothing
announces it. That is the mirror-file hazard, and it is worse here than usual because this guard's
whole job is to be trusted **without being watched**. (a) is fine as input to a classifier unit test;
it is not fine as the graph-walk fixture.

**Why not (c) alone.** Unit-testing the classifier is genuinely valuable and ships as a supplement.
Alone it is precisely the "assertions about guards rather than guards" outcome D-1.3.7 forbade: it
would prove the classifier can read a licence while leaving entirely unexercised the part that walks a
module graph and decides.

**Confidence.** High on the AD-26 reading and on (b); **medium** on SPDX-vs-text classification, left
to the developer under D-1.3.6's invariant.

**How we'd know it was wrong.** The `unknown/` fixture passing, or a fixture graph that only resolves
with `GOPROXY` reachable.

### D-1.3.9 — The licence check covers every Go module in the repository
**Orchestrator decision**, on the lead's ruling. Confirms the creator's AC18 reading.

**Verdict.** `folio-go`, `hashmatrix`, **and** `lint`. Do not narrow it.

**Grounding.** AD-26's scope line is **Binds: all** — not "Binds: `folio-go`" — and its Rule says "the
whole module graph". When the spine was written there was exactly one Go module, so the singular noun
was unambiguous; there are now three, and **the invariant's scope line resolves the ambiguity, not the
noun**. D-1.3.6 already required `lint/`'s own `x/tools` dependency to appear in the manifest it
generates, which would be incoherent if `lint/` were out of scope.

**Refinement.** The manifest labels each dependency with **the module it serves** and **whether that
module is shipped or build-time-only**. The ban applies uniformly to all three — no weakening — but the
label is honest, costs nothing, and pre-empts a future argument that a copyleft build-time tool
"doesn't really count". **A checker covering its own dependency is correct, not a bootstrap problem**,
and the story says so plainly so nobody reads it as circular.

**Standing consequence.** If a future story genuinely needs a **copyleft build-time tool**, that is an
**owner escalation, not a lead call** — it would change AD-26's stated boundary, and AD-26 fails rather
than warns by design.

### D-1.3.10 — The `math` rule is a selector rule with a closed allow-list, not an import ban
**Orchestrator decision**, on the lead's ruling. Confirms the creator's F-6, and tightens it.

**Verdict.** A **selector** rule. **Any `math` function call whose name is not one of the seven
allow-listed (`Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`) is a violation.** Non-call
references are covered **by value kind**: integer-limit constants (`MaxInt64`, `MinInt64`, `MaxInt32`,
…) are permitted; float-valued constants (`Pi`, `E`, `MaxFloat64`, `SmallestNonzeroFloat64`) are not.

**This is AD-1's actual text, not a concession to shipped code.** Verified independently in the spine
(lines 97–103): AD-1 bans importing four **package paths** (`time`, `os`, `math/rand`, `net`) and then
names *"any `math` **transcendental**"* — a class of **functions**. It never says "may not import
`math`". The clincher is the next sentence: *"The allow-listed numeric surface is `+ - * /`,
comparison, and `Sqrt`, `Floor`, `Ceil`, `Round`, `Trunc`, `Abs`, `Mod`."* Those are all `math`
**functions**. **An import-path ban would make AD-1's own allow-list unreachable and therefore
meaningless** — a rule cannot both ban a package and enumerate which of its functions are tolerated.

**Measured at `f9c27b3`.** `math` is imported under `internal/` today (`internal/pdf/numbers_test.go`,
`internal/geom/scale_test.go`). Selector references tally **34 `math.MinInt64`, 7 `math.MaxInt64`, 1
`math.Round`** — and the orchestrator checked that last one: it is at `internal/geom/scale.go:31`,
**inside a comment** ("no call to math.Round"), not a call site. So every *real* reference is an
untyped **integer** constant, with no float and no transcendental anywhere. D-1.3.1's `_test.go`
exemption would **not** have saved these — it explicitly keeps banning transcendentals in tests — so an
import-path rule reddens the shipped suite on its first run with no legitimate fix.

**Allow-list, not deny-list.** AD-1's transcendental list ends in "…", which is not decidable; its
allow-list is closed. Converting the open clause into "not one of the seven" makes the rule decidable
and **fails safe**: a future Go release adding a new transcendental is banned by default rather than
silently permitted.

**Keep the seven in the rule even though they are unusable in practice.** All seven are `float64`-only,
so AD-2/AD-23 and the arch guard prevent their use anyway — D-000.5 already recorded that the allow-list
"describes what the lint tolerates, not an invitation to use `float64`". Someone will notice the
redundancy and propose deleting them; **that would silently convert the lint from *tolerates-seven* to
*bans-all-of-math*, a different rule than AD-1's.** The story says so explicitly.

**Guardrails.** Resolve the import alias (`m "math"`) rather than matching the literal text `math.` —
`emit_source_test.go` already does this and it is the established pattern. **Match on the AST, never a
regex**: `scale.go` names `math.MinInt64` in five comments and `math.Round` in a sixth, and a text scan
would flag every one. **The compliant near-miss fixture must include exactly that shape** — a comment
naming a banned symbol — so the guard proves it does not fire on prose.

**How we'd know it was wrong.** The guard reddening on a comment, or a new `math` function arriving in
a Go release and passing unremarked.

### D-000.6 amendment (Story 1.3) — `lint/` added to the spine's §Source tree
**Ships in Story 1.3's own commit**, per D-000.6 and D-1.3.6's ruling that the AD-1 import/
math-selector lint, the map-iteration check, and the AD-26 licence check + manifest live in a new
repo-root module. Only the tree block changes; no invariant's **Binds** or **Prevents** line is
touched.

**Before** (`ARCHITECTURE-SPINE.md` §Source tree):
```text
  hashmatrix/                         # module github.com/panitw/folio/hashmatrix — holds the
                                      #   retained FMA contraction probe alone; the cross-target
                                      #   matrix driver stays in folio-go/matrix_test.go (D-1.2.3
                                      #   amended). Deliberately outside folio-go so AD-1 and the
                                      #   float64 AST guard exclude the probe by construction
                                      #   (D-000.6, Story 1.2)
  folio-node/ · folio-java/ · …       # deferred SDKs, same fixtures, same namespace
```

**After:**
```text
  hashmatrix/                         # module github.com/panitw/folio/hashmatrix — holds the
                                      #   retained FMA contraction probe alone; the cross-target
                                      #   matrix driver stays in folio-go/matrix_test.go (D-1.2.3
                                      #   amended). Deliberately outside folio-go so AD-1 and the
                                      #   float64 AST guard exclude the probe by construction
                                      #   (D-000.6, Story 1.2)
  lint/                               # module github.com/panitw/folio/lint — the AD-1 import/
                                      #   math-selector lint, the map-iteration check, and the
                                      #   AD-26 licence check + manifest. A standalone module so
                                      #   folio-go's own module graph gains no dependency (D-1.3.6);
                                      #   CI merely invokes it (D-000.6, Story 1.3)
  folio-node/ · folio-java/ · …       # deferred SDKs, same fixtures, same namespace
```

*Corrected post-review (Finding 15, this story's QA review): the "After" block above originally*
*reproduced the same one-column `#` misalignment on the `lint/` line that the spine file itself*
*shipped with; both are now aligned to the same column as `hashmatrix/`'s and `lint/`'s own*
*continuation lines (39, measured).*

**How we'd know it was wrong.** A future story adding an exemption entry for `lint/` to the AD-1
lint or the `float64` AST guard, or moving `lint/` under `folio-go/cmd/` — the whole point of this
placement is that `folio-go`'s own guards need not know `lint/` exists, and Story 3.7 still owns
`folio-go/cmd/` (D-1.3.6, AC23).

### D-1.3.11 — The dependency-free map detector fails D-1.3.6's exactness invariant; fall back to `x/tools`
**Orchestrator decision**, applying D-1.3.6 as written. **Fast path — the existing ruling determines
the answer, so this was not re-referred to the lead.** Found by the Story 1.3 reviewer.

**Verdict.** `lint/` adds `golang.org/x/tools` (BSD-3-Clause) and resolves map types through
`go/packages` with full type information. The stdlib-only `tolerantImporter` path is withdrawn.

**Why the existing ruling settles it.** D-1.3.6 permitted a dependency-free implementation as "a
permitted refinement, **not a deviation**" — but conditioned on it being **exact**, and its binding
invariant (a) is *exact detection with no false positives*. D-1.3.5 adds that "a syntactic guess is
**not** acceptable". The shipped detector is not exact, and **the code says so itself**:
`lint/internal/rules/maprange.go:66` carries the comment *"tolerant: best-effort resolution only"*,
and the module's `go.mod` says "best effort" out loud. A ruling that permits X only if X is exact is
not satisfied by an X that documents its own inexactness.

**What the reviewer proved, on a copy of the real tree — not a synthetic case.** `tolerantImporter`
returns an **empty** `types.Package` for anything `importer.Default()` misses, which is every sibling
package under `folio-go/internal/`. The unresolved subject falls through
`if !ok || tv.Type == nil { return true }` with `err == nil`, and `conf.Check`'s error is discarded
twice (`_, _ =` plus a no-op `Error` callback). Adding `type ScaleTable map[string]Length` to
`internal/geom` and two genuine map ranges over it in `internal/pdf` gives **`err=<nil>,
findings=0`**, while `go build ./...` and `go vet ./...` on that same tree exit 0. Also missed: a
type alias from another package, and another package's map-returning function. Caught: same-package
and stdlib only — and the retained fixture uses a *locally spelled* `map[string]int`, which is
precisely why the suite is green.

**In simple terms.** `internal/pdf` **already imports** `internal/geom`. So the single most likely
future shape — a map type declared in one internal package and ranged in another — is exactly the
shape the detector cannot see. The guard is not weak at the edges; it is blind down the middle.

**This is V10 again, inside the guard built to enforce V10.** *Whenever a guard's healthy output and
its dead output are the same value, that is a defect regardless of how the code looks.* A blind
detector and a clean tree both report `findings=0, err=nil`. That makes **six** instances across
three stories, and this one is the sharpest: the story that named the pattern shipped an instance of
it.

**On invariant (b), which is what the developer over-applied.** D-1.3.6's invariant (b) is "**no
dependency added to `folio-go`'s module graph**" — it says nothing about `lint/`'s own graph. The
developer extended the discipline to `lint/` by choice (its `go.mod` comment states this explicitly)
and paid for it with exactness. `x/tools` in `lint/` leaves `folio-go`'s graph untouched — verified:
`folio-go`'s `go list -m all` is one line with no `go.sum` — so (b) still holds.

**A useful side effect, not the reason.** D-1.3.6 and D-1.3.9 already required `lint/`'s own
dependency to appear in the manifest its licence check generates. With `x/tools` present the manifest
is **no longer empty**, which independently strengthens Finding 8's vacuous row-shape problem: the
manifest assertions now run against real data rather than against nothing.

**Consequences.** The `unknown/`-fixture and `GOPROXY=off` hermeticity requirements of D-1.3.8 still
bind and must be re-proved with a real dependency in the graph. `x/tools`'s BSD-3-Clause licence must
classify cleanly and appear in the manifest labelled **build-time-only** per D-1.3.9 — and that is
now a live test of the label rather than a hypothetical one.

**How we'd know it was wrong.** A map range whose type is declared in another package failing to
produce a finding — the reviewer's exact reproduction, which must become a retained fixture rather
than a one-off proof.

## Standing decisions — added mid-run

### D-000.9 — A guard's "all clear" must never be produced by the same code path as its "I could not look"
**Orchestrator decision**, on the engineering lead's framing. **Program-wide standing rule**, applies
to every remaining story and to every developer, reviewer and finisher prompt from Story 1.4 onward.

**Verdict.** Three obligations, and a diagnostic question that makes them checkable in seconds.

1. **Separate "clean" from "couldn't look."** Every guard returns findings **and a coverage
   witness** — what it examined and **how many candidate sites it considered**, not how many files it
   walked. **Zero candidates is a failure, not a pass.** Errors return separately and are assessed
   **before** findings (the `D-1.3.3 (amended)` rule, generalised to everything).
2. **The red is produced by construction, in a gate that actually runs.** A retained fixture in the
   default gate — never a mutation applied and reverted, never behind a build tag, never "verified by
   inspection".
3. **The diagnostic question, for developer and reviewer alike:** *"What would this check have
   printed if it had been unable to run at all?"* **If the answer is "the same thing," it is an
   instance.** Reviewers verify a guard by **breaking the property in the real tree and watching a
   gate that actually runs go red** — never by reading the guard.

**Situation — six instances across three stories, each caught separately as if unrelated.**

| # | Story | The instance | "All clear" signal |
|---|---|---|---|
| 1 | 1.2 | Counted-four assertion compared the loop's output to its own input | zero mismatches |
| 2 | 1.2 | CI compare job globbed uncounted artifacts under `if-no-files-found: warn` | "All four targets agree" over three |
| 3 | 1.3 | Conditional lockfile check (rejected at ruling time, D-1.3.4) | passes once the file arrives |
| 4 | 1.3 | Scanner `return err` before appending any finding | zero findings |
| 5 | 1.3 | Manifest row-shape assertion vacuous against an empty manifest | zero rows disagreed |
| 6 | 1.3 | `tolerantImporter` blind across package boundaries | `err=nil, findings=0` |

Plus three more found in the same review: an assertion comparing a message to the constant that built
it, a red-proof behind a build tag so it ran in **zero** gates, and a fixture exempted by a filename
carve-out that predated the change it was meant to prove.

**In simple terms.** The mechanism is always the same: **the success signal is indistinguishable from
the not-run signal.** A smoke alarm with a dead battery reports no smoke, and so does a working alarm
in a clean room. You cannot tell them apart by looking at the reading — only by lighting a match.

**Why writing it down did not stop it.** Story 1.3 **named this pattern in its own vacuity guards and
then shipped an instance of it.** The reason is that the code always *looks* like it passed: there is
no error, no warning, no anomalous value — the guard's healthy output and its dead output are
literally the same bytes. Inspection therefore cannot catch it, in principle. Only construction can.

**Consequences.** All three obligations go into every developer, reviewer and finisher prompt from
Story 1.4 onward. Obligation 1 changes guard *signatures*, so it is a code change, not a discipline.
Obligation 3 is the one that transfers, because it needs no understanding of the guard being checked.
Note the meta-hazard: any test written **for** this rule is itself exposed to it — a key-coverage test
that compares zero keys passes — so coverage witnesses must themselves report counts.

**How we'd know it was wrong.** A seventh instance reaching `done`. That would mean the coverage
witness is not being demanded, or is itself vacuous.

## Epic 1 decisions — Story 1.4 (ruled before the story file was written)

### D-1.4.2 — D-1.4.1 stands; 1.4 ships schema plus decidable validation only, with three deferrals
**Orchestrator decision**, on the lead's ruling.

**Verdict.** D-1.4.1 stands **in full, unamended**, and its `folio-format.md` amendment ships in
**1.4's own commit** per D-000.6 — extend the `columns[].footer` row, add rows for `footerOf` and
`footerFormat`, state the derivation, the `count` prohibition, and the two diagnostic codes.

**The wrinkle is larger than "1.4 cannot call `internal/expr`."** D-1.4.1's derivation rule is *"a
decidable syntactic property of the parsed expression tree"* — and there is no expression parser
until Story 3.2, so 1.4 cannot validate derivability either. Split by what is decidable **without a
parser**:

| Check | In 1.4? | Why |
|---|---|---|
| `footerOf` present with `footer: "count"` → load error | **Yes** | pure field presence |
| `footerOf` or `footerFormat` present with no `footer` → load error | **Yes** | pure field presence |
| `footerOf` prefixed by the table's collection path + `.` | **Yes** | a **string** prefix test against `bind` with `[]` stripped — no parser |
| `footerOf` omitted → derive from `bind`, else load error | **No** | needs the parsed expression tree |
| Footer uses the *same* aggregate evaluation as `{{sum(...)}}` | **No** | nothing renders a table until 4.5 |

**Three deferrals, each with a D-1.3.4 asserted-absence tripwire where one is possible.** Derivation
validation → **Story 3.2**, backstop 3.7 (`folio.Validate` must include it); tripwire asserts
`internal/expr` does not exist. Diagnostic codes `TABLE_FOOTER_SOURCE_UNRESOLVED` /
`TABLE_FOOTER_SOURCE_FORBIDDEN` → **Story 3.6**; **1.4 must not mint them early** — its load failures
are plain Go errors naming field, element id and offending value; tripwire asserts `internal/diag`
does not exist. Footer evaluation sameness → **Story 4.5** by name; **no package tripwire exists**, so
`deferred-work.md` is the only trigger — flagged as the weakest of the three.

**Guardrail.** The derivation deferral must not silently become permanent permissiveness: until 3.2,
a `footer` with no `footerOf` **loads**. That is a *known* gap and needs its own fixture asserting
today's behaviour, or Story 3.2's author will read the absence of a check as evidence none was
intended.

### D-1.4.3 — Two numeric kinds, one canonical spelling, and "canonical" defined as a fixed point
**Orchestrator decision**, on the lead's ruling. The load-bearing ruling of Story 1.4.

**Verdict — representation.** Exactly two numeric kinds exist in a `.folio` file, and there is no
third. **Points** (`x`, `y`, `width`, `height`, `fontSize`, `headerHeight`, `padding.*`, `margin.*`,
`border.width`, `columns[].width`) are held as **`geom.Length`** (`int64` millipoints), converted on
load by **exact ×1000 decimal-string arithmetic, never through `float64`**; on disk they are points,
up to three decimals, **trailing zeros trimmed, no trailing `.`, no `-0`, no exponent** — the same
spelling `appendLength` produces (`36` not `36.0`; `72.5` not `72.500`). **Plain integers**
(`nextId`) are `int64`, plain digits. `version` is a **string** (`"MAJOR.MINOR"`), so it carries no
round-trip hazard. Report-data numbers (AD-23's exact scaled decimals) are **not** in the template —
that type is Story 1.6's and 1.4 must not build it.

**Verdict — decoding.** `json.Decoder` with **`UseNumber()`** — not by analogy with AD-23 but because
the alternative is unavailable: `encoding/json`'s default decodes every number to `float64`, and
`folio-format.md`'s Units section already requires "the same exact-decimal path as report data
(AD-23) — never `float64`."

**Verdict — input tolerance.** Non-canonical but valid input is **accepted and normalised**, not
rejected: `36.0000` → `36`, `1e3` → `1000`. AD-9 mandates one legal *serialization*, not one legal
input, and its Prevents line explicitly contemplates a hand-editor. **Legality is a property of the
value, not the spelling:** a number is legal iff its exact decimal value is a whole number of
millipoints. `int64` millipoint overflow is a load error, never a wrap.

**Verdict — what "canonical" means.** This is what makes `Serialize(Parse(b)) == b` satisfiable
rather than vacuous or impossible. Three properties, all tested:
- **P1 — round-trip through bytes:** for every `Document d` obtained by parsing a valid file,
  `Parse(Serialize(d)) == d`.
- **P2 — normalisation is idempotent:** for every valid input `b`,
  `Serialize(Parse(b)) == Serialize(Parse(Serialize(Parse(b))))`.
- **P3 — canonical is a fixed point:** `b` is canonical **iff** `Serialize(Parse(b)) == b`.

The AC's `Serialize(Parse(b)) == b` then holds for canonical `b` **by P3's definition**, and the
substantive tests are P1, P2, and a corpus of canonical fixtures — starting with
`folio-format.md`'s own worked example, which that document labels "in canonical form" and which
ships as a golden fixture under AD-21.

**Four byte-level traps that make P3 real — all named in the story.** **HTML escaping OFF**
(`json.Marshal` escapes `<`, `>`, `&` by default; left on, a template containing `<` round-trips to
different bytes than a hand-editor wrote). **UTF-8 emitted literally**, no `\uXXXX` above `0x1F` —
the golden report is a **Thai** bank statement, and escaping Thai or CJK would destroy the
readability contract `folio-format.md` states as its purpose. **Minimal escaping only** — `"`, `\`,
and controls below `0x20`; `/` is never escaped. **Key sort is byte-order over the UTF-8 key**, not
locale-aware — all keys are ASCII lowerCamelCase so it never bites, stated anyway because a
locale-aware sort would be invisible until someone adds a key.

**The `omitempty` trap, which is the likeliest way P1 breaks.** The worked example carries
`"assets": {}` — present and empty. **No `omitempty` on any field whose zero value is legal authored
content.** An `omitempty` on `assets`, `x`, `y` or `fontSize` turns an authored `0` or `{}` into an
absent key and breaks P1 silently for exactly the documents that look most ordinary.

**Confidence.** High on representation and P1/P2/P3; **medium** on the exhaustiveness of the lead's
numeric-field inventory — the developer must enumerate every numeric field in `folio-format.md` and
classify each into kind 1 or kind 2, with a test asserting **no field escaped classification**. That
enumeration is a coverage witness in D-000.9's sense.

### D-1.4.4 — One document-wide counter; `nextId` absent or too low is a load error
**Orchestrator decision**, on the lead's ruling. **Corrects a carried assumption:** ids are
`"e" + decimal counter` (`e1`, `e2`, …), **not base-36** — verified in `folio-format.md`, whose
worked example shows `e1`…`e5` with `nextId: 6` (lines 273, 287). The first grounding's base-36 note
was a guess that never entered the contract.

**Verdict.** Top-level `"nextId"`, an integer, sorted among the top-level keys. **One counter,
document-wide** — `folio-format.md`'s `columns[]` row says each column carries "its own `id` (**same
counter as elements**, so a diagnostic can name a column)", and AD-10 requires every diagnostic
concerning a template element to carry its id, which demands document-wide uniqueness. Not per
element type. **`nextId` absent → load error. `nextId` ≤ the highest id present → load error. Never
repair, never renumber.**

**Why deriving `nextId` as `max(existing)+1` is worse than it looks.** Delete the highest-id element
and the derived counter **reuses that id on the next allocation** — violating AD-10's "never reused"
silently and permanently. Everything Folio writes carries `nextId`, so the error only ever fires on a
hand-authored file, where a located error is precisely what this story's user story asks for.

**Guardrail — the test that catches the violation a serializer actually commits.** "Unchanged by a
save" is **not** proved by a load→serialize round-trip; that is P1. It is proved by: **load, delete an
element, serialize** — then assert every surviving id is byte-identical **and `nextId` is unchanged,
not decremented**. A serializer that renumbers, or recomputes `nextId` from the surviving maximum,
**passes P1 and fails this**. Retained fixture, both polarities.

### D-1.4.5 — More than three decimals on geometry is a load error; nothing is rounded at load
**Orchestrator decision**. **Already settled by the canonical document, not derived** — verified at
`folio-format.md:24`: *"A coordinate with more than three decimal places is a load error, because it
cannot be represented exactly."*

**Verdict.** A geometry value that is not an exact whole number of millipoints is a **load error**.
Nothing is rounded at load.

**On whether this agrees with `internal/geom`'s round-half-to-even discipline: it agrees by not
rounding at all**, which is the strongest available form of agreement. There remains exactly **one**
rounding mode in the program, in exactly one function (`geom.ScaleRound`), used only where a value is
*computed* into thousandths. Ingesting an authored value is not a computation. Rounding at load would
create a second rounding site, make loading lossy, and silently render a document different from what
its author wrote — in a program whose entire thesis is that what you wrote is what you get, verifiable
by hash.

**The check is on the value, not the spelling:** `36.0000` is legal (exactly 36000 millipoints) and
normalises to `36`; `72.5001` is a load error. That keeps D-1.4.3's normalisation coherent.

### D-1.4.6 — `Render(t *Template)`; the `os` boundary is the package boundary, and `ParseTemplate([]byte)` ships at 1.4
**Orchestrator decision**, on the lead's ruling.

**Verdict — signature after 1.4**, per D-1.1.c's one-parameter-per-story convergence:
`func Render(t *Template) ([]byte, error)`.

**Verdict — the `os` boundary is the package boundary, and the spine already drew it.** `folio` is
the *shell*, not `internal/`: the spine's Paradigm line says "functional core (`internal/`),
imperative shell (`folio`, `cmd/folio`, `wasm/`, `designer/`)", and its Source tree annotates
`folio.go` as "LoadTemplate · Validate · Render · RenderTo — **the shell**". So `LoadTemplate(path
string)` and `ParseTemplate(b []byte)` live in package `folio` at the module root, where `os` is
permitted; **`internal/template` never sees a path** — it takes `[]byte` and returns a `*Template` or
an error. AD-1 is untouched because no `internal/` package imports `os`.

**`ParseTemplate` must ship at 1.4, not later.** The designer has **no filesystem**: AD-16 renders
"from serialized bytes", and AD-20 makes local file access a two-tier *browser* capability. A
path-only loader is unusable in the browser, so Epic 5 would have to widen the API — and D-1.1.c
fixed the public surface precisely so later stories converge on it rather than reopening it.
`ParseTemplate` is also the function `LoadTemplate` delegates to, so this is honest factoring rather
than scope growth, and it does not violate the one-parameter-per-story rule, which governs `Render`.

### D-1.4.7 — The Go code is normative; `folio-format.md` is a readability contract, with drift caught mechanically
**Orchestrator decision**, on the lead's ruling.

**Verdict.** `internal/template`'s parser and serializer are **normative**. `folio-format.md` is
documentation.

**Grounding, and the document self-classifies** — verified at `folio-format.md:12`: *"Only one
implementation of this schema will ever exist — the engine owns the document and the designer never
parses `.folio` (AD-15) — so this is a **readability contract, not a synchronization one**."* AD-9
assigns ownership of both parser and serializer to `internal/template`, and the Consistency
Conventions already rejected JSON Schema as "a second source of truth against AD-9". A normative
prose spec would be that same second source of truth in different clothes.

**But a readability contract that lies is worse than none**, so drift is caught mechanically, in both
directions: extract every JSON key the serializer can emit (reflection over struct tags — exact on
the Go side) and assert each appears as a backticked token in `folio-format.md`; scan the document's
backticked lowerCamelCase tokens and assert each is a real key. This is D-1.3.8's lesson applied in
the **permitted** direction: a doc that claims to mirror code has the mirroring asserted, rather than
the doc being promoted to a second source of truth.

**This test is itself exposed to D-000.9** — a key-coverage test that compares zero keys passes. It
must report **how many keys it compared**, and zero compared must fail.

**Fallback if the field tables prove too irregular to scan reliably:** narrow to the Go→doc direction
only (every emittable key must appear somewhere in the file) and record the doc→Go direction as a
deferral with an owner. **Do not** fall back to a checked-in list of key names — that is a third
artifact and reintroduces exactly the drift this closes.

### D-1.4.8 — What 1.4's loader must leave open for AD-14 (Story 1.6)
**Orchestrator decision**, on the lead's ruling. AD-14's three pre-settled cases bind **binding**, not
template loading; 1.4 implements none of them, but must not foreclose them.

1. **Distinguish absent from `null` from present-with-value** in the parsed representation — pointer
   fields or an explicit presence flag. A loader that collapses `null` into absent destroys the
   distinction AD-14's second case depends on (explicit `null` = empty and **not** an error), and the
   same decode helpers will be reached for in 1.6.
2. **Never coerce on load.** A string where a number belongs is an error naming field, element id and
   value — never a parse-and-convert. This is AD-14's third case arriving early, and it is also what
   makes P1 hold: **coercion is lossy, so a coerced value cannot round-trip.**
3. **No `omitempty` where the zero value is legal** — the same clause as D-1.4.3, restated because it
   is the mechanism by which (1) is usually lost on the *serialize* side rather than the parse side.

### D-1.4.9 — A higher MINOR loads, and unknown keys are preserved verbatim through save
**Owner decision.** Escalated by the engineering lead as `OWNER-DECISION-NEEDED`; the owner chose
**option 3 (full passthrough)**, **not** the lead's recommended option 4 (load read-only).

**Verdict.** A `.folio` file declaring a higher **MINOR** version than the library supports **loads**.
Keys the library does not recognise are **carried opaquely through parse and merged back in sorted
order on serialize**, and are **ignored when rendering**. Nothing is dropped, nothing is refused.
(A higher **MAJOR** remains a load error — `folio-format.md:46`, FR13, unchanged.)

**Situation.** `folio-format.md:46` scopes load failure to MAJOR alone, so a `1.1` file made by a
future Folio must open in today's Folio. The document says nothing about what today's library does
with the parts it has never heard of, `SPEC.md` still lists **NFR6 as an open question**, and the
spine explicitly defers the compatibility policy to pre-v1 — so the direction was silent **by
design**, which is an escalation rather than licence to fill the gap. The schema freezes in this
story (D-1.4.1), so the answer could not wait.

**In simple terms.** Someone opens a document written by a newer version of the app. The newer
version put things in it that this version has never heard of. This version can either refuse the
file, quietly throw those things away the next time the user saves, or carry them along untouched
and hand them back exactly as they arrived. The owner chose to carry them.

**Options considered, and why the others lost.**
- **(1) Refuse a higher MINOR too** — zero machinery, but contradicts `folio-format.md:46` as
  written and forecloses forward compatibility entirely: every future format addition would
  instantly orphan every older library. Would itself require amending the format doc.
- **(2) Load, drop unknown keys on save** — near-zero machinery, but **silent, unrecoverable data
  loss**: a user opens a newer file, saves, and newer-version work is gone with nothing to show it.
  The lead argued against this whatever else was chosen; it is the only option that fails without
  telling anyone.
- **(4) Load read-only, refuse to save** — *the lead's recommendation*. Destroys nothing, needs no
  passthrough machinery, keeps the promise `folio-format.md:46` actually makes, and keeps Story 1.4's
  freeze small. Rejected by the owner.
- **(3) Preserve unknown keys** — **chosen**.

**Why the owner's choice beats the recommendation.** Option 4's cost is not merely "preview but no
editing" — it is that the limitation lands on the **designer**, in Epic 5, on precisely the user who
is least equipped to understand why a file that visibly opens cannot be saved. Forward compatibility
that works only until you touch it is a promise users discover is hollow at the worst moment. Option
3 pays its cost **once, in engine code, now**, while the schema is being frozen and while the round-
trip properties are being built — rather than paying it forever in user-facing refusals. It is also
the only option under which the round-trip guarantee is *strictly stronger*: a canonical `1.1` file
processed by a `1.0` library still satisfies `Serialize(Parse(b)) == b`.

**The accepted cost, stated plainly.** This is the most machinery of the four, and it lands in the
one story whose job is to establish P1, P2, P3 and the drift test — the passthrough must be carried
through parse, held per-object, and merged back in sorted order without disturbing any of them. The
owner accepted that.

**The mandatory guardrail — this sentence lands in `folio-format.md` in Story 1.4's commit
(D-000.6), and without it the chosen option is unsafe:**

> **MINOR additions are strictly additive and never change the meaning of an existing key.**

Without that rule, a future MINOR could redefine a key this library already understands, and today's
library would render such a file **confidently and wrongly** — the one failure mode worse than
refusing it. The rule is what makes passthrough safe, and it is the part that will not be obvious
later.

**Consequences.**
1. **`version` is preserved verbatim, never downgraded.** A `1.0` library saving a `1.1` file writes
   `"version": "1.1"` — the file still contains `1.1` content, so claiming `1.0` would be a lie about
   its own contents. (Referred to the lead for explicit confirmation; flagged here because it is a
   direct forced consequence a developer would plausibly get backwards.)
2. **P1 must hold *including* unknown keys** — the round-trip corpus needs a fixture carrying unknown
   keys at multiple nesting levels, not just at the top.
3. **The D-1.4.7 drift test is unaffected in scope**: it governs keys the serializer *itself* can
   emit. Unknown passthrough keys are explicitly outside it — and that boundary must be stated, or a
   future maintainer will read a passthrough key as drift.
4. **Unknown keys are ignored at render.** They never reach `internal/layout` or `internal/pdf`, so
   AD-5 and the determinism boundary are untouched.

**How we'd know it was wrong.** A newer-version file losing content across a save in this library, or
a future MINOR that redefines an existing key — the latter would mean the additive rule was written
but not enforced by review.

### D-1.4.4 (reversed in part) — Ids are lowercase base 36 after all
**Orchestrator decision**, on the lead's ruling. **Reverses D-1.4.4's "not base-36" clause.** The
original entry stays as written; the log is append-only.

**Verdict.** Ids are `"e"` + the counter in **lowercase base 36**: counter 1→`e1`, 10→`ea`, 35→`ez`,
36→`e10`, 71→`e1z`. **`nextId` is a plain decimal JSON integer.** Every other clause of D-1.4.4 —
one document-wide counter, `nextId` absent or ≤ highest id is a load error, never repair, never
renumber, and the load-delete-serialize guardrail — **stands unchanged**; only the *rendering*
changes, not the counter semantics.

**Grounding.** `folio-format.md:118`, the Elements field table, verbatim: *"`id` | `e` + the counter
in **lowercase base 36** — `e1`, `ea`, `e1z`."*

**How the wrong ruling was made, which matters more than the fact.** The lead read the format
document's lines 1–60 and 150–300 and **skipped 60–150** — exactly where the normative field table
sits. Ruling from that partial read, it overturned its own correct base-36 note by citing the
**worked example**, whose `e1`…`e5` are **identical under both hypotheses** (measured: creator's
M-7). The orchestrator then "verified" the correction by reading the worked example — the same
ambiguous evidence — and confirmed it. Neither party opened line 118.

**The transferable rule, and it generalises past this project.** **If the other hypothesis predicts
the same observation, you have learned nothing.** Confirmation that arrives dressed as "the canonical
document governs" is *more* persuasive and therefore needs *more* scrutiny, not less. And in a
specification: **the field table is normative; the worked example is frequently ambiguous by
construction** — never overturn a spec claim on evidence both hypotheses predict.

**Consequences.** The encode/decode asymmetry is the deliverable and gets its own task with its own
table test: allocation reads `nextId` (decimal), renders `"e" + base36(nextId)`, increments, writes
back decimal; base-36 **decoding** is needed **only** to validate `nextId` against the highest id
present. It is the single most likely place for a silent off-by-representation bug. **Canonical
spelling is enforced at load:** ids must match `^e[0-9a-z]+$`, decode to ≥ 1, and carry **no leading
zeros and no uppercase** — `e01` is a **load error**, not a value to normalise, because normalising
would re-spell an id and AD-10 says ids are "never renumbered on save". Rejecting means ids
round-trip **verbatim**, so P1/P3 hold trivially. **A duplicate id anywhere in the document is a load
error** (AD-10's "never reused", plus its requirement that every diagnostic name a unique element).

**How we'd know it was wrong.** An id changing spelling across a save, or `nextId` being rendered
base-36 anywhere.

### D-1.4.10 — Amend the worked example so it is genuinely canonical; cross-check it against an independent printer
**Orchestrator decision**, on the lead's ruling. Answers the creator's DN-1.

**Verdict.** Amend `folio-format.md`'s worked example under D-000.6 so it is a true fixed point.
**Do not** add an inline-container rule to AD-9.

**Measured (creator's M-1).** Decoding `folio-format.md:233–295` with `UseNumber` and re-encoding
with `SetEscapeHTML(false)` + `SetIndent("", "  ")` gives **1648 B / 63 lines → 2096 B / 104 lines**,
first divergence at **byte 181**. Keys were already sorted — the sort is fine. The cause is **7
nested containers kept inline** (lines 242, 243, 247, 259, 274, 285, 289) while AD-9 fixes two-space
indentation. The example is labelled "in canonical form" at line 230 and was to ship as the round-trip
corpus's first golden fixture. Both cannot hold.

**Why option (b) — adding an inline rule to AD-9 — is not available.** AD-9's Rule fixes "two-space
indent". An inline nested container has **no** indentation, so a serializer emitting inline containers
is not implementing AD-9's stated rule; it is implementing a different one. D-000.6 classifies a
change to what an invariant *is* as **a direction change that goes to the owner**, not an amendment.
So (b) is not a cheaper option — it is a different decision needing different approval, to buy 41
lines of markdown.

**Why (a) is squarely within D-000.6.** The amendment changes only the clause proven wrong —
**whitespace** — and leaves the stated intent ("a minimal but complete template") intact. The
document's own text supports it: *"Top-level keys appear sorted, as does every object in the file —
that is the **serializer's job** (AD-9), not something an author maintains by hand."* The inline
formatting was a hand-authored readability choice in a markdown file; it was never a claim about the
serializer. Line 230's "in canonical form" is what turned it into one.

**The D-000.9 guardrail at corpus level, which is the real risk here.** If the story generates the
canonical bytes with its own serializer and pastes them into the document, **the document ratifies
the story's guess and the first fixture proves nothing** — the vacuity class, one level up from code.
So: **the amended example must be cross-checked against an independent JSON pretty-printer**
(Python's `json.dumps(..., sort_keys=True, indent=2, ensure_ascii=False)` or equivalent) for
**structure, indentation and key order**. Canonical form was deliberately chosen to *be* standard
two-space pretty-printing, so a second implementation must agree on those three properties. Known
divergences (separator conventions, trailing newline, escaping) are enumerated and normalised
explicitly, and **each normalisation is named in the Delivery Log** — an unnamed normalisation is
exactly where self-ratification would hide.

**Then close the loop permanently:** a test asserts the document's fenced example block is
**byte-identical** to the shipped golden fixture. After that, line 230's "in canonical form" is
mechanically true rather than aspirational, and a future edit to either one reddens.

**Consequences.** The `folio-format.md` amendments for this ruling, for D-1.4.1's `columns[].footer`
row, and for D-1.4.9's strictly-additive sentence all land in **Story 1.4's own commit**. The example
remains valid under D-1.4.1 as written — its `bind` is a single
`formatNumber(<bare row-scoped path>, <pattern>)` call, which is derivation shape #2 — so no
`footerOf` need be added to it.

### D-1.4.8 (corrected) and D-1.4.3 (extended) — presence flags are forced; the `omitempty` rule is two-sided
**Orchestrator decision**, on the lead's ruling. Both corrections came from the creator's M-2.

**D-1.4.8(1) named a mechanism that does not work.** Measured: `encoding/json` leaves a `*int` **nil
for both** an absent key and an explicit `null`, so **pointer fields cannot distinguish the three
states**. The **presence-flag route is forced** — a wrapper carrying `set bool` (or
`json.RawMessage`) with a custom `UnmarshalJSON`, since only the unmarshaller is told whether it was
called at all. The *invariant* is unchanged and is what binds: absent / `null` / value must remain
distinguishable, so AD-14's second case stays open for Story 1.6.

**D-1.4.3's `omitempty` rule was one-sided.** A **nil map with `omitempty` removed serializes as
`null`** — not `{}`, not absent — which breaks P1 against the worked example's `"assets": {}`. The
rule is therefore two-sided: **no `omitempty` where the zero value is legal content, *and* every map
and slice field must be initialised to empty rather than left nil.** The second half is the one a
developer meets only when the fixture reddens.

**Also measured, narrowing an earlier claim:** Thai and CJK emit literally **either way** under
`encoding/json`, so D-1.4.3's literal-UTF-8 trap is a real requirement but **not** a default hazard —
the story should say so rather than implying the default would mangle Thai. Default `Marshal` **does**
escape `<`, `>`, `&`, so that half of the trap is live. `Encode` appends a trailing `\n`.

### D-1.4.11 — `ScanAbsences` is the seventh D-000.9 instance, and it is in the enforcement code itself
**Orchestrator decision**, on the lead's ruling. Found by the creator (M-5) at baseline `af6b5d5`.

**Verdict.** `ScanAbsences` gains a **coverage witness**: it returns how many absence checks it
**evaluated**, and **zero evaluated is a failure**. Closing this is Story 1.4's **task T2**, landing
**before** the story's two new tripwires are added to `absenceChecks`.

**Situation.** `ScanAbsences` is the **only** checker in `lint/internal/rules` returning
`(findings, error)` with no stats, and `TestAbsencesProductionScan` asserts only zero findings — so
**an empty `absenceChecks` list passes.** Story 1.4 extends that exact list with the `internal/expr`
and `internal/diag` tripwires from D-1.4.2.

**Why the sequencing is the whole point.** Adding entries to a list whose emptiness passes is
precisely how a tripwire becomes decorative: the tripwires would be "present", and the check would
pass **identically if they were absent**. The fix must land first, or the two new tripwires are
unverifiable from the moment they ship.

**Why this instance is the most significant of the seven.** It is in the code we shipped **one story
ago, specifically to enforce this class of rule.** Story 1.3 named the pattern in its own vacuity
guards, shipped an instance of it in the detector (D-1.3.11), and shipped a second instance in the
absence checker. That is the strongest available evidence that **the pattern survives being written
down** — and it is why D-000.9's diagnostic question belongs in every prompt rather than only in this
log.

**How we'd know it was wrong.** An eighth instance reaching `done`.

### D-1.4.12 — Extending a closed set is a MAJOR change, never a MINOR
**Owner decision.** Raised by the engineering lead as a forced consequence of D-1.4.9; the lead ruled
it and flagged that it might narrow the owner's choice. The orchestrator escalated it anyway under
the standing backstop (*a scope cut with lasting consequences goes to the owner even when the lead
has ruled*), and the owner **confirmed** the lead's reading.

**Verdict — the mandatory `folio-format.md` sentence says all three things:**

> A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an
> existing key, and it may **not** extend a closed set of legal values (element `type`, `locale`,
> `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`). Extending
> a closed set is a **MAJOR** change, because every existing library validates those sets as load
> errors.

**Situation — why D-1.4.9 does not work without this.** `folio-format.md` makes **eight closed sets**
load errors (verified: `locale` at line 47, element `type` at line 119, and six more). **Passthrough
does not save them.** Passthrough carries unknown *keys*; a new *value* in a known enum is **not** an
unknown key — it is a known key whose closed-set validation rejects it. So if a MINOR could extend
any of those sets, a `1.0` library would **refuse to load** that `1.1` file outright, and the forward
compatibility D-1.4.9 just paid for would be void for exactly the additions most likely to be made.

**In simple terms.** The owner bought "a newer file still opens." But "we added a sixth kind of
element" is not a new *key* the old library can politely ignore — it is a value in a list the old
library checks against a fixed set, and there is nothing it can do with an element type it cannot
draw. So that kind of addition has to announce itself as breaking.

**Options considered.**
- **(a) Closed-set extension is MAJOR** — **chosen.** One rule, uniformly enforced, and the failure
  is loud.
- **(b) Older libraries skip unknown elements and report the skip** — maximum forward compatibility,
  but it conflicts with **FR13**'s "never a best-effort render", and it means a document can render
  *incomplete*. On the golden report — a bank statement — a silently missing section is materially
  worse than a refused file.
- **(c) Decide per set** — more faithful to each set's real risk, but eight separate rules, each
  needing its own test and its own line in the format doc. Rejected on the cost of remembering and
  enforcing eight boundaries where one suffices; and on inspection every one of the eight genuinely
  cannot be faked by an older library (an unknown `locale` has no font or formatting rules; an
  unknown `align` would be silently wrong).

**Why this wins.** It is the only option whose failure mode is loud, and it keeps the format's
evolution rule to a single sentence a maintainer can hold in their head. **The accepted cost is
real and permanent:** evolution on those eight sets is now expensive forever — adding a sixth element
type or a fifth locale orphans every older library.

**A side effect the lead noted and the orchestrator endorses:** this protects PRD counter-metrics C1
and C2. "Just add one more element type" or "one more palette colour" now costs a MAJOR version
rather than sliding in on a MINOR — scope discipline enforced by the format rather than by
willpower.

**Consequences.** The story must keep the two validation paths **visibly separate**: an unknown
**key** passes (passthrough); an unlisted **value** on a known key is still a load error. Conflating
them is precisely how passthrough degrades into "validation is optional".

**How we'd know it was wrong.** A MINOR release adding a closed-set value — which would mean the rule
was written but not enforced at review.

### D-1.4.13 — `version` is a property of the document: carried verbatim, raised only by content, lowered never
**Orchestrator decision**, on the lead's confirmation. Confirms consequence 1 of D-1.4.9, which was
logged provisionally and flagged confirmation-pending.

**Verdict.** A `1.0` library that loads a `1.1` file, preserves its unknown keys and saves writes
`"version": "1.1"`. **And the symmetric half, which a naive implementation gets wrong in the other
direction:** a future `1.1` library loading a `1.0` file and changing nothing must write `"1.0"`. One
rule: **`version` is a property of the document, carried verbatim; raised only when content actually
requiring a higher version is introduced; lowered never.**

**Why this needs stating rather than being obvious.** It means **this library writes files declaring
a version it does not implement**, which reads wrong at first glance and which a developer will
plausibly "fix". The naive implementation — "write the library's own version" — is wrong **both
ways**.

**The reframe that stops the fix, and it belongs in the code comment.** **`version` describes the
document, not the writer.** A PDF 1.7 file edited by a tool that understands only 1.4 features does
not become a PDF 1.4 file. That sentence turns the apparent absurdity into the only coherent reading.

**Three further reasons, beyond the two already logged.**
1. **AD-9 would otherwise be violated directly.** If a `1.0` library downgrades on save, the same
   document has **two canonical byte forms** depending on which library last wrote it — precisely
   what AD-9's Prevents line exists to stop ("the designer and a hand-editor producing different
   bytes for the same document").
2. **P3 would break for every older file.** If a `1.1` library rewrote unchanged `1.0` files as
   `1.1`, `Serialize(Parse(b)) != b` for all of them — so the naive rule breaks the very property
   passthrough was chosen to strengthen.
3. **Other SDKs depend on the honesty.** The spine plans `folio-node/`, `folio-java/`, "same
   fixtures, same namespace". A downgraded version misleads every future SDK, and they conform
   against the same fixture corpus, so the lie propagates.

**Consequences.** At Story 1.4 the *raise* path is unreachable (no `1.1` exists) — **do not build it,
do not foreclose it.** Two cheap tests pin the rule: a `1.0` file round-trips as `1.0`, and a
synthetic higher-MINOR file round-trips as `1.1` with its unknown keys intact.

**How we'd know it was wrong.** Any save that changes `version` without changing content.

### D-1.4.14 — The hand-written serializer is confirmed as correct architecture; the parser is stdlib
**Orchestrator decision**, on the lead's ruling. **Opens with a correction the orchestrator got wrong.**

**Correction first.** The orchestrator reported that the developer "hand-wrote the parser and
serializer". **That is wrong, and it overstated the risk by most of its size.** Verified by file:
`parse.go`, `rawvalue.go` and `decodehelpers.go` **all import `encoding/json`** and decode with
`UseNumber`; only `serialize.go` and `jsonstring.go` do not. The JSON **tokenizer** — escape handling,
UTF-8 decoding, the number grammar, the parts genuinely dangerous to hand-roll — **is stdlib**. The
hand-written surface is the **emitter only**. The overstatement reached the reviewer's brief and was
corrected mid-review.

**Verdict.** The hand-written **serializer** is confirmed as the correct shape — not a tolerated
deviation.

**It is forced.** Five requirements must hold simultaneously in one byte stream: AC8's merged sorted
sequence over known **and** passthrough keys; AC19 HTML escaping off; AC20 literal UTF-8; AC21
minimal escaping; and `geom.Length` spelled as trimmed decimal points. `encoding/json` can be coerced
toward this, but every leaf would need its own escape-disabled production, and `geom.Length` would
need a `MarshalJSON` — **putting `.folio` format knowledge inside `internal/geom`**, whose entire job
under AD-2 is owning the scalar type and which deliberately imports nothing. At that point struct tags
carry no information.

**And it is the project's established architecture, not an exception.** `jsonstring.go` is the direct
analogue of `internal/pdf/numbers.go`: one file, one choke point, every string through
`appendJSONString` so the three escaping rules are enforced in exactly one place. **That is AD-3's
shape applied to a second output format** — hand-write the emitter, confine the hazard to one file,
allow no other route to an output byte. Confirming it is *consistent* with D-1.1.b, not a departure.

**The hazard it creates, and its closure.** The millipoints→decimal spelling now exists **twice**:
`pdf.appendLength` and template's `appendPoints`. Two implementations of one conversion that must
agree forever is a classic drift hazard. It is closed by a **shared value table** with mirror tests on
both sides (AC25) — one data source, two consumers, the same construction ruled for the two-caller
checkers in D-1.3.3. **That table must never be duplicated or forked.**

**One correction to it:** the table is *shared* but lives under a path owned by one consumer
(`folio-go/testdata/template/`). Move it to a neutral location (`folio-go/testdata/shared/`), or a
future reader in `internal/pdf` will find their test reading another package's directory and "tidy"
it. Finisher-sized change.

**How we'd know it was wrong.** A route to an output byte in `serialize.go` that bypasses
`appendJSONString`, or a second copy of the length-spelling table.

### D-1.4.15 — AST extraction is necessary but not sufficient; add behavioural extraction and assert set equality
**Orchestrator decision**, on the lead's ruling. Supersedes D-1.4.7's stated *mechanism*; its
invariant is unchanged.

**The lead's own framing of its error, recorded because the fix is general.** D-1.4.7 named
"reflection over struct tags" as the mechanism — **specifying an implementation that had not been
chosen yet.** The invariant was "exact on the Go side, both directions, no checked-in list"; "struct
tags" was an assumption smuggled in wearing a mechanism's clothes. **This is the third such instance**
(D-1.2.3 bundling the driver into `hashmatrix/`; D-1.3.6, which the lead caught itself and left open).
The lead's own diagnosis of the tell: **it over-specifies precisely when it can picture the code.**

**Adopted fix, binding on the lead from here:** every named mechanism in a ruling is tagged
**`(mechanism: binding)`** or **`(mechanism: illustrative)`**. An untagged mechanism is read as
illustrative and queried. Had D-1.4.7 stated only its invariant, there would have been no mechanism to
depart from and nothing for the developer to fail to report.

**Verdict — the substance.** AST extraction collects string literals from `kv{...}` composite literals
in `serialize.go`. That is exact **iff every emitted key passes through that convention.** A key
emitted by any other route — a direct byte append, a `kv` built from a variable — is invisible to it.
**The failure is asymmetric in the bad direction:** if such a key is *also* undocumented, **neither**
direction of the drift test notices. An emitted, undocumented, unextracted key is silent drift, in the
test whose only job is catching drift. The developer's note that "a key that stops being written stops
appearing in `serialize.go`'s source" is true for **removal** and false for **addition by another
route**. Its first pass at the extractor being vacuous is evidence this family of proxy is fragile in
practice, not only in principle.

**Required — a second, behavioural extraction:** serialize a document exercising every known field,
**parse the output bytes**, and collect every key actually present. Then **assert
`astKeys == runtimeKeys`**, both directions, difference printed by name on failure. Two independent
extractions of the same set catch each other's failure mode: a key emitted outside the `kv{}`
convention appears in runtime and not AST; a fixture failing to exercise a field appears in AST and
not runtime. **Neither alone is exact; together they are.** *(mechanism: binding for the set-equality
assertion; illustrative for how the maximal document is assembled.)*

**Cheap path first — make it a measurement, not new work.** Check whether the **union of the existing
round-trip corpus** already covers every key in the AST set. If it does, the maximal document already
exists. **If it does not, that gap is itself a finding** — the round-trip corpus would not be
exercising every field, independent of the drift test. Coverage witnesses stay on **both** extractors;
zero keys from either is a failure.

**This is now the project's dominant successful pattern: two independent producers of the same answer,
asserted equal.** D-1.4.10's independent pretty-printer, D-1.4.14's shared spelling table, and this.
Three for three, it is the only construction that has held.

### D-000.10 — Handoff dispositions one line per binding ruling; the reviewer mirrors it
**Orchestrator decision**, on the lead's ruling. **Program-wide standing rule, from Story 1.5.**
Replaces the exhortation in D-1.2.6 with a structural check; D-1.2.6's substance still binds.

**Verdict.** Every developer return dispositions **every ruling ID the story file enumerates**, one
line each, as exactly one of:
- `applied-as-stated`
- `applied-with-a-different-mechanism` — **and one sentence on what and why**
- `not-reached-in-this-story`

**The reviewer independently dispositions the same list against the code.** Two dispositions
disagreeing is the signal.

**Why D-1.2.6 stopped working — and it is not defiance.** Three consecutive developers resolved a
ruling conflict inside the diff (Story 1.2's placement-vs-reuse contradiction; Story 1.3's
over-application of invariant (b) into a neighbouring scope; Story 1.4's substitution of AST
extraction for struct-tag reflection). **In all three the local answer was defensible and only the
ruling-level consequence was a defect.** D-1.2.6 asks a developer to detect a *conflict between two
rulings* — which requires holding both in mind and noticing an interaction, in flight, while solving
something else. **That is the least reliable moment available.**

**Why an enumerated list beats a prompted recollection.** The orchestrator first proposed a section
titled "rulings whose stated mechanism I did not use — empty if none." The lead's correction is the
difference between working and not: **that still has an empty default**, and still asks the developer
to recognise that a ruling *had* a stated mechanism and that they departed from it. A table over an
enumerated list has **no empty default** — you must write N lines, and you cannot write the line for
D-1.4.7 without looking at what D-1.4.7 said. **It converts "notice something" into "walk a list",
which models and humans both do reliably.**

**Half the remedy is on the lead's side** — the mechanism tagging in D-1.4.15. If a ruling names no
binding mechanism, there is nothing to silently depart from; some of the three instances were
avoidable reports generated by over-specification.

**Do not add friction to the decision itself.** Three for three, the developers' answers were
defensible. **The defect is reporting, not judgement** — so the fix belongs at handoff and should cost
a table.

**How we'd know it was wrong.** A fourth silent departure, or a disposition table filled in without
being read — every line `applied-as-stated` on a story where the reviewer finds a departure.

### D-1.4.16 — Two D-000.9 shapes in Story 1.4, both closed behaviourally rather than by documentation
**Orchestrator decision**, on the lead's ruling.

**The `Presence` three-state API.** `absent[T]()` being unused is **not** a defect — absence is the
**zero value**, so it is never constructed. Neither of the orchestrator's readings was right: the API
is not incomplete and no state is unreachable. **The real issue is sharper:** at Story 1.4 **nothing
branches on `Null` vs absent** — D-1.4.8 required the loader to *leave the distinction open* for Story
1.6, not to consume it. So a three-state API exists whose third state no production code reads. That
is a D-000.9 shape, and the guard is behavioural: **a test must prove `{}` and `{"a":null}` produce
different `Presence` values *and* serialize back differently.** If serialization collapses them the
distinction is decorative and Story 1.6 inherits a seam that was never load-bearing. **That test is
what makes D-1.4.8 real rather than declared.**

**`Render(t *Template)` ignoring `t`.** Provisional is correct under D-1.1.c — nothing can render a
template until fonts land at 1.5 and layout in Epic 2. But **"documented as provisional" decays into
prose nobody re-reads.** Do both: the doc comment states it ignores its argument **and why**, in the
`PROVISIONAL` idiom Story 1.1 established; **and a named test asserts `Render(validTemplate)` and
`Render(nil)` currently produce identical bytes.** The moment Story 1.5 or 1.6 makes them differ, that
test fails and **forces its own deletion** — turning a silent "the argument does nothing" into an
asserted fact with an expiry date. Same mechanism as D-1.3.4's asserted absence, applied to a
behaviour rather than a path.

### D-000.11 — Every gate runs `-count=1`; Go's test cache silently skips guards that read out-of-module files
**Orchestrator decision.** **Program-wide standing rule, effective immediately.** Found by the Story
1.4 reviewer; reproduced and widened by the orchestrator.

**Verdict.** Every test invocation that functions as a **gate** — locally, in CI, and in every
sub-agent prompt — runs with **`-count=1`**. `.github/workflows/ci.yml`'s two `go test ./...` steps
are amended in Story 1.4's commit.

**Situation.** Go's test cache does not track reads of files **outside the module**. Every guard in
this project reads out-of-module files: the drift test reads `_bmad-output/specs/spec-folio/folio-format.md`;
`lint`'s tripwires read directory paths across the repo; `folio-go/fixture_test.go` reads
**`fixtures/minimal-rect/expected.json` at the repo root** — which is AD-21's byte-identity golden. So
a cached `ok` can mean "the guard passed" or "the guard was not run and its input has since changed",
and the two are indistinguishable.

**Measured by the orchestrator, decisively.** With the cache primed by a plain `go test`, injecting a
genuine new key into `folio-format.md`:

```
plain `go test` (cache-eligible)  →  ok  (cached)     <- drift NOT detected
`go test -count=1`                →  FAIL             <- drift detected
```

**This is D-000.9 at the gate level, and it is the widest instance yet.** The pattern is *a guard's
"all clear" produced by the same code path as its "I could not look"* — here the not-looking is done
by the **build system**, above every guard at once. No amount of care inside a checker prevents it;
`ScanAbsences`' coverage witness is correct and still reports nothing, because the process never runs.

**Why it went unnoticed for four stories.** Every gate figure in this run was measured with
`-count=1` — that habit is why the reported numbers were real. But `ci.yml` (shipped in Story 1.3)
runs **plain `go test ./...`** at lines 45 and 95, and `actions/setup-go` restores a build/test cache
by default, so **CI is exposed too**. The local habit masked a defect in the committed pipeline.

**Consequences.** `-count=1` in: `ci.yml`'s two steps; every developer, reviewer and finisher prompt's
gate list; and any future workflow. The matrix workflow is unaffected — it invokes `-run` with
`-tags=matrix` against freshly built per-target binaries, and its harness compares captured bytes
rather than trusting a `go test` exit code (D-1.2.1). **Note for reviewers: a cached `ok` is not
evidence.** When verifying any claim about a guard, re-run with `-count=1` or the observation is
worthless.

**How we'd know it was wrong.** A guard that passes in CI while its out-of-module input is provably
wrong — i.e. exactly the reproduction above, in the pipeline rather than on a laptop.

### D-000.6 amendment (Story 1.4) — `folio-format.md`'s four amendments, verbatim before/after
**Finisher entry**, closing this story's own AC51 obligation (3) ("the before/after of all three goes
into the decision log" — extended to four; see Amendment D below and this story's Finding 17). Ships
in Story 1.4's commit, per D-000.6. Before quotes are baseline `af6b5d5` (the field table's state
before Story 1.4 touched it, as quoted in D-1.4.1); after quotes are this commit's shipped text.

**Amendment A — the `columns[].footer` row, D-1.4.1 via D-1.4.2.**

Before (the whole row, unchanged since baseline):
> `columns[].footer` | *Optional.* `sum` · `count` · `avg`. Computed over the whole collection, never
> per page (AD-11).

After — two new rows inserted (developer, Story 1.4 `review`), and the `footer` row itself extended
(finisher, this pass, Finding 9 — the developer's diff left it byte-identical to baseline despite
AC51 and the Delivery Log both describing it as "extended"):
> `columns[].footer` | *Optional.* `sum` · `count` · `avg`. **Unchanged — names the operation only**
> (D-1.4.1); the numeric source is `columns[].footerOf`, below. Computed over the **whole collection**,
> never per page (AD-11). Omitted means no footer cell for that column.
>
> `columns[].footerOf` | *Optional.* A bare root-relative dotted value path (e.g.
> `"transactions.amount"`) naming the numeric source the footer aggregates — no `{{ }}`, no function
> call, no `[]`. Legal only alongside `footer`, and never alongside `footer: "count"` (storing it would
> be a second source of truth against `bind`, AD-13). When `footer` is present and `footerOf` is
> omitted, it is **derived** from the column's own `bind` … **As of Story 1.4, this derivation is not
> yet implemented** — until Story 3.2 lands it, a `footer` with no `footerOf` simply loads, and the
> aggregate itself is not computed until Story 4.5. Story 3.6 mints the two diagnostic codes this
> eventually becomes: `TABLE_FOOTER_SOURCE_UNRESOLVED` … and `TABLE_FOOTER_SOURCE_FORBIDDEN` …
>
> `columns[].footerFormat` | *Optional.* A `formatNumber` pattern applied to the computed footer
> value. Legal with all three `footer` operations.

**Amendment B — the MINOR-additive safety sentence, D-1.4.9 / D-1.4.12.**

Before (baseline `af6b5d5`): the Document table's "Top-level keys appear sorted…" sentence stood
alone, with no MINOR-additive guardrail sentence beneath it.

After — the developer added D-1.4.12's three-clause verdict (Story 1.4 `review`), **plus a fourth
sentence D-1.4.12's verdict does not contain**; the finisher (this pass, Finding 16) dropped that
fourth sentence so the shipped text matches the owner-ruled text verbatim (D-1.1.b):
> A MINOR increment may add **new optional keys** only. It may **not** change the meaning of an
> existing key, and it may **not** extend a closed set of legal values (element `type`, `locale`,
> `align`, `valign`, `columns[].footer`, `border.edges`, `page.orientation`, `page.size`). Extending
> a closed set is a **MAJOR** change, because every existing library validates those sets as load
> errors.

(Removed by the finisher, not part of D-1.4.12's verdict: *"A higher MINOR still loads (`version` is
carried verbatim); a higher MAJOR is always a load error."* — accurate in isolation, but not the
owner-ruled sentence, and D-1.1.b forbids shipping a ruling's mandatory verbatim text extended.)

**Amendment C — the worked example, D-1.4.10.** Developer-authored (Story 1.4 `review`); unchanged by
the finisher, per DN-1/D-000.6 obligation (1) — the worked example's *content* was never the clause
proven wrong, only its whitespace was, and that is not this pass's concern. Before/after (1648 bytes /
63 lines, 7 inline containers → 2096 bytes / 104 lines) is already recorded verbatim in the story's own
Delivery Log.

**Amendment D — the worked example's provenance note (newly enumerated by the finisher, Finding 17;
the text itself shipped in the developer's Story 1.4 `review` diff, not added by the finisher).**

Before (baseline `af6b5d5`): no such paragraph existed; the worked example fence was followed directly
by the `{{page}}`/`{{pages}}` closing prose.

After:
> *(This example is generated by the module's own canonical serializer and cross-checked, structure,
> indentation and key order, against an independent pretty-printer —
> `json.dumps(..., sort_keys=True, indent=2, ensure_ascii=False)` — which reproduces it byte-identically
> with scalars normalised out, D-1.4.10, AC16. It is shipped as the golden fixture
> `folio-go/testdata/template/golden/worked-example.json`, and a test asserts this fenced block stays
> byte-identical to that file.)*

**How we'd know it was wrong.** A `folio-format.md` clause quoted in this log that does not match the
file at the head of this story's commit, byte for byte.

## Epic 1 decisions — Story 1.5 (ruled before the story file was written)

*First story under D-1.4.15's mechanism tagging: every named mechanism is marked
`(mechanism: binding)` or `(mechanism: illustrative)`.*

### D-1.5.1 — Replace the one-module count with an allowlist, plus an independent PDF-writer denylist
**Orchestrator decision**, on the lead's ruling. Closes grounding-refresh **divergence (2)**, parked
since the lead's first grounding pass.

**Verdict.** Two assertions in `gomod_test.go`, not one.
1. **Allowlist (primary).** `go list -m all` must equal **exactly** the expected set — after this
   story `{github.com/panitw/folio/folio-go, github.com/boxesandglue/textshape}` — asserted by name,
   with the difference printed both ways on failure. *(mechanism: binding)*
2. **PDF-writer denylist (independent backstop).** No module path matching a known Go PDF writer
   (`signintech/gopdf`, `boxesandglue/baseline-pdf`, `jung-kurt/gofpdf`, `pdfcpu/pdfcpu`,
   `unidoc/unipdf`, `phpdave11/gofpdi`) at any depth. *(mechanism: binding for the property;
   illustrative for the list, which must be extendable)*

**This corrects the lead's own grounding note**, which said "a denylist of PDF writers plus the
licence check". **A denylist alone is weaker than the count it replaces:** a PDF writer nobody listed
passes silently, and lists rot. The count assertion's *spirit* was right — anything new must be a
conscious act — and only its *form* was brittle.

**Why an allowlist keeps the property the count had.** A developer meeting a red allowlist adds one
line naming the module they just added, which is the correct act and a reviewable one-line diff that
puts a human in the loop for every new dependency. That is what makes deletion unattractive — unlike
an assertion whose entire content is `!= 1`, which invites deletion when it goes red. The denylist
then catches what the allowlist cannot: someone updating the allowlist carelessly to unblock
themselves. **Two independent checks with different failure modes** — the construction that has now
held four times (D-1.4.10's independent printer, D-1.4.14's shared table, D-1.4.15's set equality,
this).

**Split with `lint/`.** The graph assertion **stays in `gomod_test.go`** — it is an **AD-6** property
(no third-party PDF writer), cheap, and must fail in the fast per-story gate. The **AD-26 licence**
half stays in `lint/`. Different invariants, different homes; do not merge them.

**How we'd know it was wrong.** The allowlist being edited in the same commit that adds a dependency
without the dependency being named in the commit message — i.e. the human-in-the-loop property
lapsing into a formality.

### D-1.5.2 — Validate `unitsPerEm` at the trust boundary in `internal/fontset`; `ScaleRound` gets no second variant
**Orchestrator decision**, on the lead's ruling. Closes grounding-refresh **divergence (3)**.

**Verdict.** `internal/fontset` validates `unitsPerEm` **before** any value derived from a
caller-supplied font reaches `geom.ScaleRound`. Valid range **16–16384** (the OpenType
`head.unitsPerEm` range); anything outside — including `0` — is a located load error naming the font
and the value. At Story 1.5 that is a plain Go error (`internal/diag` is Story 3.6, same pattern as
Story 1.4). *(mechanism: binding)*

**No non-panicking `ScaleRound` variant.** AD-2 says *"Font scaling is **one** exported function with
one documented rounding mode"*. A second entry point would be two scaling functions, and the two
would drift in exactly the way AD-2 exists to prevent.

**In simple terms.** The panics are correct as programmer-error guards; the fix is **layering, not a
second door**. Untrusted bytes are validated at the boundary, after which the internal invariant
genuinely holds — so the spine's *"nothing in `internal/` panics on malformed input"* is satisfied,
because by the time bytes reach `ScaleRound` they are no longer malformed.

**The guardrail that makes this by-construction rather than by-discipline** *(mechanism: binding)*:
**the parsed-font type must not expose a raw, unvalidated `unitsPerEm`** that a caller can hand to
`ScaleRound`. Validation happens in the constructor and the validated value is the **only** one
reachable. A field that can be read before it is checked will eventually be read before it is checked.

**Retained fixture, both polarities:** a font with `unitsPerEm = 0` produces a located error and
**not a panic**; a valid font scales correctly; plus a negative case at `unitsPerEm > 16384`.

### D-1.5.3 — `textshape` clears AD-26, measured; and it is the story that proves the licence check works
**Orchestrator decision**, on the lead's ruling. **Verified independently by the orchestrator.**

**Verdict.** Cleared. Measured at `v0.0.15`, twice, by both the lead and the orchestrator:

| Check | Result |
|---|---|
| Licence | **MIT** (`LICENSE`, line 1) |
| `go list -m all` in a scratch module | **the module alone** — zero transitive dependencies |
| `"time"` imported in non-test code | **none** |

AD-26's "at any depth" therefore has **no depth to check**, and D-1.3.9's copyleft escalation clause
does not fire.

**Still required in this story** *(mechanism: binding)*: the `lint/` licence check runs over
`folio-go`'s now-**non-empty** graph, and the manifest records `textshape` as the **first shipped**
dependency under D-1.3.9's shipped-vs-build-time labelling. **A check that has only ever run against
an empty graph has never been exercised** — this is the story that proves it works, and the proof
belongs in the Delivery Log with the actual manifest output.

**A false alarm, pre-empted.** `textshape`'s own `go.mod` declares `go 1.23.0` and **`toolchain
go1.24.0`**. Neither affects us: per D-1.1.a's addendum a `toolchain` directive binds **only the main
module**, and its `go 1.23.0` is below our `1.25.0` floor so it imposes nothing. Expect someone to
raise it as a toolchain-pin violation; **it is not one.**

**On entering the graph before Story 2.3 justifies it** — anticipated at first grounding and fine:
`subset/` is what 1.5 needs, `ot/` shaping is what 2.3 needs. Same module, two stories, one licence
clearance.

**Standing rule restated:** had it carried copyleft at any depth, that is an **owner escalation**
under D-1.3.9 — not a lead call, not a developer workaround.

### D-1.5.4 — The two-run timestamp comparison is vacuous by construction; assert the head table's bytes instead
**Orchestrator decision**, on the lead's ruling. The sharpest finding in this bundle.

**Verdict — reject two-run comparison as the proof of "no wall-clock timestamp".** *(mechanism:
binding — the rejection)*

**Why it is vacuous, not merely weak.** OpenType `head.created` and `head.modified` are
**LONGDATETIME: seconds since 1904**. **Two renders in the same second produce identical bytes even
if the subsetter stamps the clock on every call.** A same-second comparison *cannot fail* — so it
would report "no timestamp" for a subsetter that timestamps every single time. That is the dominant
defect class (D-000.9) appearing inside the test written to exclude it, and it is the tenth instance
recorded in this run.

**Required proof — byte-level and structural** *(mechanism: binding for the property; illustrative
for the offsets)*: assert that in the embedded `FontFile2` stream the head table's `created` (offset
20) and `modified` (offset 28) are **derived from the source font** — equal to the source's values,
or zero — and never a value near the current time. A **single-run** assertion that fails immediately
if a clock is ever introduced, including by a future dependency upgrade.

**Supporting evidence, measured, but deliberately not a test:** `textshape` imports no `time` in
non-test code, and `ot/metrics.go:27–28,59–60` only **reads** those fields as `int64` from offsets 20
and 28. Record that in the Delivery Log as measured provenance. **Do not turn it into an assertion
over the module cache** — that would bind the test to `GOMODCACHE` layout, whereas the behavioural
assertion survives an upgrade and a source scan of a pinned version does not.

**Subset tag and scope** *(mechanism: binding)*:
- Six letters `A`–`Z` from a hash of the **sorted** glyph-id set (AD-7). **Sorting happens before
  hashing**, and the glyph set will come from a map — so **Story 1.3's `ScanMapRange` guard is
  load-bearing here** and must not be worked around.
- Pass a **sorted, deduplicated** slice into `subset.Subset` / `CreatePlan`, so determinism is ours
  **by construction** rather than an assumption about the library's internal ordering.
- **One subset per font per document**, over the union of glyphs (AD-8, §Consistency Conventions).
  The subsetting call happens **once, after all glyphs are collected** — never per page, never per
  element. Asserted **behaviourally**: a two-page document using the same face embeds exactly one
  `FontFile2`.
- **An API-shape trap:** `subset.Subset(font *ot.Font, codepoints []rune)` takes **codepoints**,
  while AD-7's tag is over **glyph ids**. Those are different sets and the mapping is a cmap lookup.
  **AD-7 says glyph ids**, so the tag derivation needs the resolved glyph ids, not the input runes.

### D-1.5.5 — `Render(t *Template, f FontSet)`; parameters take their final positions; the pinning test dies here
**Orchestrator decision**, on the lead's ruling.

**Verdict — signature after this story:** `func Render(t *Template, f FontSet) ([]byte, error)`.

**The rule that settles argument order for every remaining story** *(mechanism: binding)*: parameters
are always in the **final target order**, and each story **inserts** its parameter into its final
position rather than appending. Target is `Render(t, d, p, f)`, so: **1.5 → `Render(t, f)`; 1.6 →
`Render(t, d, f)`; 1.7 → `Render(t, d, p, f)`.** No call site is ever reordered, only extended.

**The `Render(validTemplate) == Render(nil)` pinning test dies at 1.5, and its death is the AC being
met.** Story 1.5's ACs require *"a document using a subset of a Latin face"* and *"the text renders
correctly"*. Text has to come from somewhere, and **the template is the only source of text in the
system** — Story 1.6 adds *data binding*, not text. So `t` starts mattering here. The test was built
in Story 1.4 to fail and force its own deletion (D-1.4.16); **this is the story that does it, and the
finisher deletes it rather than weakening it.** First self-retiring assertion in this run to reach its
expiry as designed.

**A new provisional that needs the same treatment** *(mechanism: illustrative)*: AD-24 makes element
`x`/`y` **band-relative**, and bands are placed by `internal/layout`, which does not exist until
Epic 2 (composition is Story 2.5). So 1.5 must adopt a provisional origin convention to place text at
all. Whatever it picks must be documented as provisional **and pinned by an assertion that fails when
band composition arrives** — otherwise the provisional convention quietly becomes the real one.

### D-1.5.6 — `folio-go/fonts/` is a package as well as a directory; the shipped faces are Story 2.2's
**Orchestrator decision**, on the lead's ruling. **The apparent conflict is not one.**

**Verdict.** AD-8's Rule answers it directly: *"no package under `internal/` embeds font data… The
repository holds font binaries in exactly one place, `folio-go/fonts/`; **the `folio/fonts` package**
wraps them with `go:embed` for native callers."* So `folio-go/fonts/` is simultaneously the binary
directory **and** a Go package, sitting at the **module root, not under `internal/`**. Both rules hold
at once: the embed lives outside `internal/`, and `go:embed` reaches it because it is inside the
module. The `FontSet` **type** is declared in package `folio` (spine §Source tree: *"`fontset.go` —
FontSet as a public input"*); `internal/fontset` does resolution and subsetting against the value it
is handed. `internal/` never embeds and never queries the host.

**Scope call — Story 1.5 does NOT ship the production faces** *(mechanism: binding)*. Story 2.2 is
"The shipped font set and its fallback chain"; that is where Noto Sans / Thai / SC land, and it is
where the unresolved **AD-12 `ja` coverage question** comes due — flagged at grounding as potentially
an **owner** call, because adding a face breaks Story 5.4's fixed "CJK 7.4 MB" payload numbers and the
~9 MB budget. Pulling several megabytes of binaries **and that open question** into 1.5 would be scope
creep in the wrong direction. **1.5 builds the `FontSet` type, the resolution seam and the subsetting
path, and tests against one small Latin test face.**

**AD-26 consequence for that test face** *(mechanism: binding)*: any font binary committed — fixture
or shipped — travels with its **licence text and copyright line**. AD-26 is explicit: *"Redistributed
non-code assets keep their own terms and their notices."* An OFL face in `fixtures/` needs its OFL
text beside it, and the `lint/` manifest should account for it. **This is the first non-code
redistributed asset in the repo and it sets the pattern for Story 2.2.**

### D-1.5.7 — The sorted-input mechanism is withdrawn as inert; determinism is proved by repeat- and permutation-invariance
**Orchestrator decision**, on the lead's ruling. **Withdraws a `(mechanism: binding)` clause of
D-1.5.4.** Found by the Story 1.5 creator by compiled probe.

**Verdict.** D-1.5.4's *"pass a sorted, deduplicated slice into `subset.Subset`/`CreatePlan`, so
determinism is ours by construction"* is **withdrawn**. Two proofs replace it:

1. **Repeat-invariance within a single process** *(mechanism: binding)*. Call the subsetting path
   **N ≥ 16 times in one process** on the same glyph set; assert every output is byte-identical to
   the first.
2. **Permutation-invariance** *(mechanism: binding)*. Distinct input orders → identical bytes.
   **Only meaningful if we do NOT sort first.**

**Why the original was worse than vacuous — it was inert.** `Input.glyphs` is `map[GlyphID]bool`
(`input.go:12`) and `AddGlyphs` writes into it, so **input order is destroyed at the door**. Measured
by the creator: ascending vs reversed runes → **byte-identical, 6860 bytes both**;
`AddGlyphs(70,5,40,5)` vs `AddGlyphs(5,40,70)` → **byte-identical, 2684 bytes both**. Determinism
comes entirely from `createCompactMapping`'s `sort.Slice` at `plan.go:334–350` — **inside the
dependency, where `ScanMapRange` cannot see**. Verified in source by the orchestrator.

**The part most worth recording, in the lead's own words.** Sorting our input **makes the permutation
test vacuous**, because permuted-then-sorted inputs are *the same input*. The "by construction" move
**removed the ability to observe the property it claimed to guarantee.** New failure mode to watch:
***a defensive normalisation upstream of a check can delete the check's discriminating power.***

**Why proof (1) closes the gap that looked unbuildable.** Go randomises map iteration order **per
`range` statement**, not once per process — the lead measured **8 distinct iteration orders** from
ranging an 8-element map 200 times in one process. So if a future `textshape` upgrade removes its
internal `sort.Slice`, this test reddens immediately and with overwhelming probability (coincidence
over 8 glyphs is ~1/8!, driven to nothing by N repeats). It needs no knowledge of the dependency's
internals, survives upgrades, and requires neither pinning nor vendoring. **Version pin plus
permutation was not the honest limit.**

**The precision the original ruling was missing — two call sites, opposite answers:**
- **Subset input → do NOT sort.** Inert (the map eats it) and it makes proof (2) vacuous.
- **Tag derivation → MUST sort.** AD-7 requires a hash over the *sorted* glyph-id set; here the
  ordering is **ours**, reaches an output byte, and comes from a map range — so **`ScanMapRange` is
  load-bearing at this site** and must not be worked around.

Same word, two sites; the original ruling collapsed them.

### D-1.5.8 — The subset tag hashes the closure set in SOURCE-font numbering, sorted ascending
**Orchestrator decision**, on the lead's ruling. Answers the creator's DN-2.

**Verdict** *(mechanism: binding)*. The tag derives from the glyphs **actually retained in the
embedded program**, expressed in **source-font glyph numbering** — `Plan.GlyphSet()`, whose doc
comment reads *"returns the set of **old** glyph IDs to retain"* (verified at `subset/plan.go:368`) —
**sorted ascending**. Not the requested set, and **emphatically not the output numbering**.

**Measured.** 3 requested glyph ids → `Plan.GlyphSet()` = **8**, `NumOutputGlyphs` = **8**
(`.notdef`, composite components, GSUB closure). Requested ≠ embedded, and the two readings yield
different tags, different PDF bytes, and a different golden hash — so this had to be ruled **before**
the fixture is recorded, since a golden hash is permanent by design.

**Why not the requested set.** ISO 32000-1 §9.6.4's tag exists so a consumer can tell two subsets of
the same base font apart. If it is a function of what we *asked for*, two different embedded programs
— same request, different closure after a `textshape` upgrade — carry **the same tag while differing
in content**. The tag would lie in the one job it has. AD-22 already makes any subsetting change a
versioned breaking change, so a tag that moves on a subsetter upgrade is a **correct signal**, not a
defect.

**Why not the output numbering — the trap, named explicitly because it is the most natural-sounding
reading.** Verified in source: `createCompactMapping` sorts retained old GIDs ascending and does
`for newGID, oldGID := range gids`, assigning `newGID = index`. **So the output glyph-id set is
always `{0, 1, … n-1}`.** Hashing it produces a tag that is **a function of the glyph count alone** —
two completely different 8-glyph subsets would receive **identical tags**. "The glyph ids in the
embedded font" sounds like the obvious reading and is catastrophic.

**The retained fixture that kills that reading permanently** *(mechanism: binding)*: **two documents
whose glyph sets are different but the same size must produce different tags.** That single assertion
is red under the output-numbering reading and green under the ruled one, and it is not vacuous.

*(mechanism: illustrative)* for the encoding — a digest over the sorted big-endian GID sequence
mapped into six `A`–`Z` letters. What **binds** is: closure set, source numbering, sorted, and the
discrimination property.

### D-1.5.9 — `minimal-rect` is rebound to the seam it actually pins; a new template-driven fixture ships beside it
**Orchestrator decision**, on the lead's ruling. Answers the creator's F-8.

**Verdict** *(mechanism: binding)*. `fixtures/minimal-rect/` is **retained unchanged** — hash, bytes,
README and "do not regenerate" all stand — and is **reclassified** as pinning the **fontless
PDF-emission path in `internal/pdf`**: AD-3 number emission, xref arithmetic, content-derived `/ID`.
Its test calls **`pdf.Serialize()` directly** rather than going through `Render`. **`Render` may then
require a non-nil template**, returning a located error on nil. Story 1.5 adds a **new**
template-driven fixture beside it, with text and an embedded subset.

**Why option (a) — authoring a `.folio` that reproduces the existing 547 bytes — is out on the
merits, not on effort.** Story 1.1's bytes contain **no colour operator**; verified by the
orchestrator, the content stream is exactly `72.5 100.25 200.125 50 re`. That was deliberate, relying
on PDF's default black fill because *"adding `0 g` would be a number that is not a `geom.Length`"*.
But a template-authored rectangle draws from `style.background`, and a background of `#000000` goes
through D-1.1.b's colour rule — `c*1000/255`, round-half-to-even, decimal route — and **emits a
colour operator**. So any `.folio` drawing that rectangle produces different bytes, unless the
serializer grows a "background is black → emit nothing" special case, which would be an appalling
thing to own forever.

**Why (c) — re-recording — is out.** It discards the only artifact in the repository verified
byte-identical on all four targets, and vacuity guard 4 makes a hash change a defect until proven an
intended versioned change.

**Why the bind dissolves rather than being traded away.** It existed **only because the fixture was
bound to the public API**, which D-1.1.c designed to change every story. The fixture's actual subject
was never the API — `Render(nil)` was a one-line passthrough to `pdf.Serialize()`. Rebinding it to
the function it truly pins removes the conflict entirely, and incidentally improves the public API:
an entry point whose documented behaviour is "ignores its argument" becomes one that rejects nil.

**Consequence that must be handled in the same breath** *(mechanism: binding)*. If the subprocess
seam switches to the font-embedding document, `minimal-rect` **loses its four-target verification** —
the property that makes it the most valuable artifact in the repo. So **the seam renders both**:
`FOLIO_SUBPROCESS_RENDER=1` keeps the fontless path, a second selector covers the template+font
document, and **the matrix runs both**. Cheap, and it preserves cross-target coverage of the baseline
while adding it for fonts.

### D-1.5.10 — Three confirmations: the fontset seam, the head-table assertion, and the subprocess guard
**Orchestrator decision**, on the lead's ruling.

**The `internal/fontset` seam — within D-1.5.2's intent, adopted as its binding form**
*(mechanism: binding)*. `ot.Head.UnitsPerEm` is an exported unchecked `uint16` and `Plan.Source()`
hands back the raw font, so D-1.5.2's "must not expose a raw, unvalidated `unitsPerEm`" can only bind
**our** wrapper. Therefore: **do not leak `*ot.Font`, `*ot.Head` or `*subset.Plan` out of
`internal/fontset`.** A vendor type we cannot constrain must not become part of a seam we are relying
on to be constrained.

**The head-table "never near the current time" half is withdrawn — it was not merely weak, it was
illegal.** Reading "near the current time" requires reading the clock, and **D-1.3.1 bans `time` in
`_test.go` files under `internal/`**. It was unimplementable where it needed to run. **The binding
assertion is equality only** *(mechanism: binding)*: the embedded head table's `created`/`modified`
**equal the source face's values exactly**, compared byte-for-byte against the source font. No clock,
no dependence on how old the committed face is, and it fails the instant anything starts stamping.
D-1.5.4's original *"equal to source, **or zero**"* is also tightened — `subsetHead` does `copy` over
the whole 54-byte table (verified: source `created=3304067374 modified=3573633780`, subset
identical), so **equality is what actually happens**, and leaving "or zero" as an accepted
alternative would mask a future change.

**The subprocess vacuity guard — endorsed** *(mechanism: binding)*. Assert the child output contains
a `FontFile2` **before** comparing. Without it, AC10 **and** AC15's four-target matrix pass on two
identical **fontless** rectangles — "all clear indistinguishable from could-not-look" wearing a font
story's clothes. **The matrix override for this story is only meaningful if the subprocess path
actually embeds a font.**

### D-1.5.10 (corrected) — the "never near the current time" withdrawal's STATED REASON was false; the conclusion survives on an independent argument
**Finisher correction**, on the code-reviewer's Finding 10. The above entry's second paragraph
asserts the head-timestamp test "must live" under `internal/`, where D-1.3.1's `time` ban applies.
**Verified false**: the shipped test, `TestEmbeddedHeadTimestampsEqualSourceExactly`, ships at
`folio-go/render_test.go` — package `folio`, the **module root**, not under `internal/`. D-1.3.1's
ban never applied to it, and never constrained where this test could live.

**The withdrawal's CONCLUSION still stands, on the independent argument that was always the
stronger one and does not depend on the false premise**: the face-age trap. Roboto's committed
`head.modified` measures `≈3.57e9` (LONGDATETIME, seconds since 1904) against `≈3.87e9` for the
present (2026-08-23) — a fixed, aging gap that only widens. Asserting "not near the current time"
against a value that will **never** be near the current time is not a weak assertion, it is a
**meaningless** one: it cannot discriminate a subsetter that never timestamps from one that does,
because the committed face is already permanently far from "now." The equality-only assertion this
story shipped is strictly stronger regardless of which argument withdraws the alternative, so no
code changes — only the withdrawal's stated REASON, which was wrong.

**Also amend AC11a's second bullet in the story file** (*"never a value near the current time is
withdrawn ... which is where this test must live"*): drop the false "must live under `internal/`"
clause; the withdrawal rests on the face-age argument alone.

**How this was found.** The code-reviewer, dispositioning the ruling table independently rather
than trusting the developer's `applied-as-stated` line, checked WHERE the test actually ships
before accepting the stated reason for why an alternative was ruled out — the same discipline
D-000.9 asks of every check ("what would this have printed if the reasoning were simply asserted
without verification?").

### D-000.12 — Never verify bytes or hashes through the shell wrapper's pipes; use `rtk proxy` and a file
**Orchestrator decision.** **Program-wide standing rule, effective immediately.** Found by the Story
1.5 reviewer; reproduced by the orchestrator.

**Verdict.** Any verification of **binary content or a hash** must run through `rtk proxy` **writing
to a file**, never piped through the wrapped shell. Applies to every agent in this pipeline and to
the orchestrator.

**Measured, on the repository's own golden fixture:**

```
git show HEAD:fixtures/minimal-rect/expected.pdf | shasum -a 256   →  4d2a90cc…8466   WRONG
rtk proxy git show HEAD:… > /tmp/gs.pdf ; shasum -a 256 /tmp/gs.pdf →  0f925e1b…4f7c   correct
shasum -a 256 fixtures/minimal-rect/expected.pdf                    →  0f925e1b…4f7c   correct
fixtures/minimal-rect/expected.json "sha256"                        →  0f925e1b…4f7c
```

**Why this is dangerous rather than merely annoying.** The wrong hash **reads exactly like the
project's worst blocker**: "the byte-identity golden fixture has changed." D-1.5.9 makes
`minimal-rect` untouchable and vacuity guard 4 makes a hash change a defect until proven an intended
versioned change — so a false positive here would trigger a bisect, an owner escalation under
D-1.2.2's divergence rule, and quite possibly a "fix" to an artifact that was never wrong. The
reviewer hit it, recognised it as a tooling artifact rather than a defect, and verified by another
route. **A less careful reader would have filed the blocker.**

**This is the second false signal from the same wrapper.** D-000.11's discovery came from it printing
`Go build: Success` over a non-zero exit — masking a real break. Now it corrupts binary through
pipes, inventing a break that does not exist. **It fails in both directions**, which is the worst
property a measuring instrument can have.

**Consequences.** Every developer, reviewer and finisher prompt already carries "use `rtk proxy` or
otherwise confirm raw exit codes" (D-000.11); it now also carries **"never pipe binary or hash
verification through the wrapper — write to a file and hash the file."** When any agent reports a
**hash mismatch on a committed fixture**, the orchestrator re-verifies by the file route **before**
treating it as a divergence — a wrong hash is a tooling fault until proven otherwise.

**How we'd know it was wrong.** A hash-mismatch report that disappears when re-measured by the file
route — which is now the first thing to try, not the last.

## Epic 1 decisions — Story 1.6 (ruled before the story file was written)

### D-1.6.1 — `Decimal` is an `int64` coefficient plus a decimal exponent, in `internal/bind`, over one shared literal splitter
**Orchestrator decision**, on the lead's ruling. The core of Story 1.6.

**Verdict — the shape is fixed by the spine, not by the lead.** AD-23's Rule states it verbatim
(verified at `ARCHITECTURE-SPINE.md:418`): *"`internal/bind` converts each to an exact scaled-integer
decimal — an **`int64` coefficient plus an exponent** — carrying the literal's own precision."*
*(mechanism: binding for the `int64` coefficient + integral decimal exponent; illustrative for field
names and exponent width)*

`1.50` → `{150, -2}`; `1.5` → `{15, -1}`; `36` → `{36, 0}`; `1e3` → `{1, 3}`. Precision preserved,
distinct literals distinct. **This is not `geom.Length`** — millipoints are a fixed 3-decimal scale;
a bank-statement literal carries its own.

**"Too large" is two bounds, and only the first is the AC's letter** *(mechanism: binding)*:
1. **Coefficient overflow** — the coefficient must fit `int64`. Only **significant** digits count, so
   `0.00000000000000000001` is fine (`{1, -20}`) while a 20-significant-digit literal is an error.
   This is the AC's "never a silent narrowing".
2. **Exponent bound** — stated openly as a **safety addition beyond the AC's letter**. Report data is
   untrusted caller input and an unbounded exponent is a **memory-amplification hazard**: a 10-byte
   literal like `1e999999999` is representable in the struct but renders to a gigabyte of digits.
   Reject at parse with a located error rather than accepting a value that fails later, further from
   the cause. The specific range is illustrative; **that one exists is binding.**

Both are located errors naming the path and the element id, per AD-14.

**One implementation of JSON number-literal decomposition in the module** *(mechanism: binding)*.
`internal/template/decimal.go:76` already has
`splitJSONNumber(literal) (sign, intPart, fracPart, exp)` — verified — which is exactly the primitive
1.6 needs. **Reuse it; do not write a second.** Two splitters would drift, and this is the money path.
Whether it is exported from `internal/template` or moved lower is illustrative; `internal/bind`
importing `internal/template` is the correct direction (Document → BoundTree). **What binds is that
only one exists.**

**Placement, with a pre-commitment that prevents the worst outcome** *(mechanism: binding)*.
`Decimal` lives in **`internal/bind`** at 1.6, per AD-23's "`internal/bind` converts each". But AD-23
**Binds** names *both* `internal/bind` and `internal/expr`, and Epic 3's `sum`/`avg`/comparison do
exact decimal arithmetic — so `internal/expr` needs this type at Story 3.2, while `internal/bind`
imports `internal/expr` (bind resolves expressions). That makes `expr → bind` an **import cycle and a
hard compile error**. The dangerous resolution is **not** the cycle — Go stops that — **it is someone
duplicating the type to break it.** Pre-committed now:
- **`internal/expr` may never import `internal/bind`.** The arrow runs bind → expr.
- When 3.2 needs `Decimal`, the type **moves** (to `expr` or a leaf package). **Duplicating it is
  forbidden.**
- Registered in `deferred-work.md`, **owner Story 3.2**. DW-5's existing `internal/expr`-absent
  tripwire is the forcing function that makes someone re-read this.

### D-1.6.2 — AD-14's three cases land unchanged, and the proof is that two of them *disagree*
**Orchestrator decision**, on the lead's ruling.

**Verdict.** Absent path = **Error** naming the path and element id; explicit `null` = **empty, not an
error**; wrong kind = **Error, never a coercion**.

**What makes D-1.4.8's presence seam load-bearing rather than decorative** *(mechanism: binding)* — a
retained fixture triple over **one template and one path**, differing only in the data:

| Data | Outcome |
|---|---|
| `{}` (path absent) | **Error**, naming path + element id |
| `{"customer": {"name": null}}` | **renders empty**, no error |
| `{"customer": {"name": 123}}` into a text element | **Error**, no coercion |

**Rows 1 and 2 are the proof:** inputs differing **only** by absent-versus-null produce **opposite**
outcomes. That is a test that cannot pass if the three-state distinction collapses — precisely what
D-1.4.16's serialization test could not establish, since nothing branched on the difference until
now. If a future refactor makes `Presence` two-state, row 1 or row 2 goes red immediately.

**Vacuity guard** *(mechanism: binding)*: all three rows assert against the **same** template and
**same** path. Three separate templates would let each row pass for its own unrelated reason.

### D-1.6.3 — Reuse Story 1.3's lint; close the shell gap with a file-scoped rule, not reachability
**Orchestrator decision**, on the lead's ruling.

**The gap is real.** Story 1.3's lint binds `folio-go/internal/`. The AC's claim is about *"a render in
progress"*, whose dynamic extent **starts in package `folio`** — where D-1.4.6 deliberately permits
`os` so `LoadTemplate` can read a file. **So `folio.Render` could call `os.Getenv` today and no guard
would fire.**

**Verdict — structural separation, not reachability analysis** *(mechanism: binding for the property;
illustrative for the filename)*, the same shape D-1.1.b used for `numbers.go`:

> The file declaring `Render` and `RenderTo` imports **none** of `os`, `time`, `net`, `math/rand`.
> The world-reading shell entry points (`LoadTemplate`) live in a **different file**, which may.

Decidable, exact, no false positives, and it fails the moment someone adds a convenience `os.Getenv`
to the render entry point. **1.6 adds nothing else** — `internal/`'s coverage is already 1.3's, and a
second checker is what the lead's horizon note correctly forbade.

**A clarification that must be in the story** *(mechanism: illustrative)*: the AC's "locale" means the
**host** locale. AD-12 requires the render to use the locale **declared in the template** — reading
the document's locale is mandatory, discovering the machine's is banned. Without that sentence
someone will read the AC as forbidding AD-12.

### D-1.6.4 — `Render(t *Template, d Data, f FontSet)`; `Data` is bytes; two decode trees, one shared primitive
**Orchestrator decision**, on the lead's ruling.

**Signature** *(mechanism: binding)*: `func Render(t *Template, d Data, f FontSet) ([]byte, error)` —
insert-in-final-position against the target `Render(t, d, p, f)`, per D-1.5.5.

**`type Data []byte`** — the named type from D-1.1.c, and the reason is unchanged and still that
ruling's strongest argument: `Data` and `Params` become **adjacent same-typed arguments at Story
1.7**, and *in a product whose acceptance fixture is a bank statement, that swap must be a compile
error, not a support ticket.* **Do not "simplify" to `[]byte`.**

**Bytes, not a decoded value** — confirmed, and Story 1.4 strengthens rather than weakens it: AD-23
requires the library to own the `UseNumber`-preserving decode, so a caller-decoded value would arrive
with its literals **already destroyed**.

**Two decode paths, one shared primitive** *(mechanism: binding)*: the template decoder (1.4) builds a
**schema-typed** tree; the data decoder (1.6) builds a **generic** value tree. Those tree-builders
stay separate — merging them would force the schema to model arbitrary caller JSON. What they
**share** is the number-literal decomposition from D-1.6.1. One splitter, two consumers, separate
trees.

### D-1.6.5 — Accept a bare dotted path; reject expression syntax loudly; reserve `{{page}}`/`{{pages}}`
**Orchestrator decision**, on the lead's ruling.

**The grammar 1.6 accepts** *(mechanism: binding)*: `{{`, optional whitespace,
`ident ( "." ident )*`, optional whitespace, `}}`, where `ident` is `[A-Za-z_][A-Za-z0-9_]*`.

**Everything else is a located error naming the element id** — anything containing `(`, `)`, `[`,
`]`, `,`, a quote, an operator, or interior whitespace. **Rejecting is the mechanism that stops 1.6
becoming a second expression implementation.** A permissive 1.6 that tries to "handle" a function call
leaves Story 3.2 with two parsers to reconcile, and the wrong one wins whichever way that goes.

**The error message must name Epic 3** *(mechanism: illustrative)* so a template author writing
valid-in-3.2 syntax is told it is *not yet supported* rather than *malformed* — otherwise they rewrite
a correct template.

**`page` and `pages` are reserved, and this is not hypothetical** *(mechanism: binding)*. AD-4 is
absolute: *"No `page` namespace exists for expressions to reach, and none may be added."* `{{page}}`
and `{{pages}}` are late-bound **slots** recognised by the template layer (Story 2.7), not data paths.
**Verified: `folio-go/testdata/template/golden/worked-example.json:56` already contains
`"Page {{page}} of {{pages}}"`** — so 1.6 meets them on its **first run against the corpus**.

**The failure this prevents.** If 1.6 resolved single-segment paths from data with no reservation,
then a data document containing a `page` key would **silently capture the page-number slot** — and the
symptom is *a wrong number on a rendered page*, not an error. So `{{page}}` and `{{pages}}` as
complete bindings are **left untouched** for Story 2.7: never resolved from data, never an error.
**Retained fixture: data containing a top-level `page` key must not change what `{{page}}` renders.**

**Pre-commitment for 3.2** *(mechanism: binding)*: 3.2's parser **replaces** 1.6's path matcher — the
matcher is **deleted**, not kept alongside. Registered in `deferred-work.md`, owner Story 3.2, same
entry as D-1.6.1's type move.

### D-000.9 (extended) — The diagnostic question applies to red-proofs, one level up
**Orchestrator decision**, on the lead's framing. Extends D-000.9 rather than replacing it.

**Verdict.** Story 1.5's review found the defect class had moved from *guards* to **proofs about
guards** — four artifacts named as red-proofs that could not fail, including one asserting
`rejectedTag(8) != rejectedTag(8)` which never called the function under test and stayed green under
the exact mutation that reddened the real fixture.

**The diagnostic question applies verbatim, aimed one level higher:** ***"What would this red-proof
have printed if the mutation had never been applied?"*** If the answer is "the same thing", it proves
nothing. **Added to the reviewer prompt explicitly**, because "red-proof" now names an artifact that
can itself be vacuous.

### D-1.5.10 (correction accepted) — the lead's falsified scope claim, and the pattern behind it
**Orchestrator note**, on the lead's own account. Recorded because the pattern matters more than the
instance.

D-1.5.10 withdrew the head-table clock comparison as *"illegal, not just weak"* under D-1.3.1's ban on
`time` in `_test.go` under `internal/`. The Story 1.5 reviewer verified **the test ships at the module
root, not under `internal/`** — so that ban never applied. **The conclusion (equality-only) stands on
the face-age argument alone**, which was the stronger half; the finisher corrected the record as
`D-1.5.10 (corrected)`.

**The lead's own diagnosis, worth holding it to:** *"both of my falsified claims this run were **scope
assertions about where a rule applies**, made without opening the file"* — this and the base-36
reversal, one story apart. Its adopted fix: **verify a file's location before invoking a
directory-scoped rule.** This sits alongside D-1.4.15's mechanism tagging as the second self-imposed
correction the lead has adopted this run.

### D-1.6.6 — `decodePoints` hangs on a wrapped exponent: a shipped defect, fixed in 1.6, recorded against 1.4
**Orchestrator decision**, on the lead's ruling. **A defect in committed code**, found by the Story 1.6
creator (M-1) while probing the primitive D-1.6.1 requires it to reuse.

**The defect, measured and isolated by the orchestrator, per literal, using Go's own `-timeout`:**

| literal | `splitJSONNumber` | `decodePoints` |
|---|---|---|
| `1e2000000` | `exp=2000000, err=nil` | **errors correctly in ms** — "overflows int64 millipoints" |
| `1e99999999999999999999` | `exp=7766279631452241919`, **err=nil** | **HANGS >20s** |
| `1e9223372036854775808` | `exp=-9223372036854775808`, **err=nil** | **HANGS >20s** |

Root cause, verified in source: `parseDecimalExponent` accumulates `n = n*10 + int(c-'0')` with **no
overflow check and no digit bound**, then hands an astronomical or negative exponent to `big.Int`
scaling. Note the second row — a **positive** exponent silently becomes the **most negative int64**.
`1e2000000` is caught only because `2000000` is a *sane* value the millipoint bound rejects quickly.

**Why this is severe rather than a latent primitive bug.** Story 1.4's user story reads verbatim:
*"I want to load a `.folio` file and have malformed or future-versioned templates rejected with a
**located error**, so that a bad template is caught in CI rather than in production."* **A malformed
template hangs the loader instead of being rejected.** And the designer (Epic 5) loads user-authored
templates, so this is reachable from untrusted input in the product as planned.

**Verdict — option (c):** the code fix lands **in Story 1.6**; the defect is **recorded against Story
1.4**.

**Why not reopen 1.4.** D-1.6.1 **already** mandates an exponent bound and **already** mandates
reusing this exact splitter — so 1.4's defect and 1.6's ruled requirement are *the same line of code*.
Reopening 1.4 would land a bound that 1.6's ruling immediately supersedes or duplicates, in the one
primitive we have specifically forbidden from existing twice. D-000.4's bisect convention exists for
**matrix regressions attributable to an owning story**; this was found by the next story's creator and
the ruled fix already lives in that story's scope.

**What "recorded against 1.4" concretely means** *(mechanism: binding)*: an appended
**`## Defect found after done`** section in Story 1.4's file — **appended, never rewritten**, per the
completed-story-file precedent — stating the measurement, naming the fixing commit, and saying plainly
that **1.4's user-story promise was not met by the shipped code**. **Not** a `deferred-work.md` entry:
this is a defect being fixed now, not deferred work, and filing it there would dilute that file.

**The proof obligation** *(mechanism: binding)*: a **retained corpus fixture**, not only a unit test —
a `.folio` containing `1e99999999999999999999` must produce a **located error, quickly**. The unit test
proves the primitive; the fixture proves 1.4's actual promise now holds **through the public loader**.

**Red-proof, with a wrinkle that must not be fudged** *(mechanism: illustrative)*: before the fix this
test **hangs**, and a hang is not a clean assertion failure — Go's `-timeout` panics the process. That
timeout **is** the red, and it is recorded as *"hung; killed by `-timeout` at Ns"*, never dressed up as
an assertion failure. Per D-1.2.5 a proof states what it actually observed.

**The bound itself** *(mechanism: binding)*:
- **Strip leading zeros; never strip trailing zeros.** `0.00000000000000000001` concatenates to a
  **21-character** digit string whose integer value is **1**, so `len(digits) > 19` would **reject a
  legal literal** — the count is taken **after** stripping leading zeros. Stripping *trailing* zeros
  would turn `1.50` into `{15,-1}`, destroying the precision D-1.6.1's own table requires.
- **Edge case named explicitly:** for `0.00`, stripping leading zeros yields the **empty string**.
  That must produce coefficient `0` with the exponent preserved (`{0,-2}`), so `0.00` still displays
  as `0.00`. A naive implementation produces an empty digit string and then errors or loses the scale.
- **Two separate fixes in `parseDecimalExponent`, both needed:** (1) an **overflow check during
  accumulation** — checking *after* the loop is too late, the value has already wrapped; (2) a
  **documented magnitude bound**, rejected **before any `big.Int` scaling is attempted**.
- **Layering:** the shared splitter rejects the **absurd** (no wrap, no explosion); each consumer then
  applies its **own tighter** check — `decodePoints` against `geom.Length`'s millipoint range,
  `internal/bind` against `Decimal`. **The splitter's bound must be the wider of the two**, so it never
  refuses something a consumer could represent. The specific number is illustrative; **tests at the
  bound and one past it are binding.**

**How we'd know it was wrong.** Any template literal that neither loads nor produces a located error
within the gate's timeout.

### D-1.6.7 — Point a second production caller of the `float64` checker at the whole module
**Orchestrator decision**, on the lead's ruling. Answers the creator's AC5 `DECISION NEEDED`.

**Verdict — option (c)** *(mechanism: binding)*: a **second production caller** of the existing
`float64` checker, scanning the whole module.

**The three-way collision was with the lead's phrasing, not between the documents.** D-1.6.3 said
*"1.6 adds nothing else… a second **checker** is what the horizon note correctly forbade."* **A caller
is not a checker.** D-1.3.3 built exactly this architecture — a pure checker function over a target
path with two callers, production and fixture — and adding a production call site is the shape that
design exists to serve. The lead corrected its own wording rather than let a real gap ship because of
it.

**Measured (creator's M-6, controlled):** `os.Getenv` **and** `float64` inside `folio.Render` → build
OK, `lint` 39/39 pass, `folio-go/internal/...` 157/157 pass. The **identical** mutation under
`internal/geom/` turns both guards red with named rule findings. The guards are alive; the module root
is simply outside their target. **Verified by the orchestrator:** `grep "float64\|float32" folio-go/*.go`
→ **zero matches**, tests included — so the new caller goes **green on its first run**.

**Why not (a), "document the module root as uncovered".** It documents a hazard closable in one line.
AD-2 is explicit that *"no `float64` appears in any layout, measurement, or emission **signature**"* —
the public API **is** a signature — and AD-23's Prevents names FMA contraction, which a `float64` in
`folio.Render` handling a value en route to `internal/` reintroduces exactly. "Deliberately uncovered"
would be true and indefensible.

**Why not (b), extend the file-scoped rule.** It conflates two different invariants (AD-1's
world-reading ban, AD-23's float ban) into one rule, and decisively it scopes `float64` to **one
file** — a `float64` in `fontset.go` or `version.go` would stay invisible. It under-covers the thing
it is meant to close.

**Scope of the new caller** *(mechanism: binding for the properties; illustrative for the traversal)*:
scan **the whole module** (`folio-go/`, root and below) rather than the root directory alone — it
subsumes `internal/` and automatically covers `fonts/`, `cmd/` and `wasm/` when they arrive, rather
than needing a new call site each time. **Include `_test.go`**, consistent with `internal/`'s
treatment; it is clean today, and D-1.3.1's precedent is that test exemptions are granted only where
*required*, never pre-emptively. **Skip `testdata/` and any dot-directory** — by **category** (leading
dot), never by naming `.matrix-build`; a name on a list is the rot pattern. **Keep the existing
`internal/`-scoped caller** — its vacuity guard names `geom` and `pdf` specifically and still earns its
place. Two callers, one checker.

### D-000.13 — A red-proof asserts on rule id and message, never on exit status; controls are valid syntax
**Orchestrator decision**, on the lead's ruling. **Program-wide standing rule.** Promoted from the
Story 1.6 creator's M-6a.

**Verdict** *(mechanism: binding)*: **a red-proof asserts on the rule id and the finding message,
never on exit status or mere failure.** And its companion: **a control mutation must be valid syntax
with forbidden semantics** — the same rule `D-1.3.3 (amended)` set for violating fixtures, now applied
to controls.

**What happened.** The creator's first control appended an `import` **after** code, producing a
**parse error**. Both scanners then failed through `D-1.3.3 (amended)`'s **error path** rather than by
finding the rule violation. **The test went red — for the wrong reason — and would have been recorded
as proof the guard works.**

**Why this is the sharpest form of the class yet.** It is not a proof that *cannot* go red; it is a
proof that goes red **without exercising the property at all**. D-000.9's diagnostic question
("what would this have printed if it could not run?") does not catch it, because it *did* run and it
*did* fail. The new question is: ***did it fail for the reason it names?***

**Consequences.** Stated in the developer, reviewer and finisher prompts alongside D-000.9's two
diagnostic questions. Note the irony worth preserving: `D-1.3.3 (amended)`'s error path — added to
stop a crashed scan reading as clean — is what made this false red possible. **Separating "failed" from
"found a violation" is the same discipline as separating "clean" from "could not look", one level
along.**

### D-1.6.8 — The canonical corpus is doing real work; keep it in the default gate
**Orchestrator decision**, on the lead's ruling. Endorses the creator's M-4 as AC20.

**Verdict.** The corpus holds four `{{…}}` in two schema fields, including
`{{formatNumber(transaction.amount, "#,##0.00")}}` at `worked-example.json:19` — **three AC16
rejection triggers at once**. It sits on `Column.Bind`, which `collectTextRuns` never reaches, and the
golden fixture is round-tripped rather than rendered, **so there is no conflict today**. But an
implementation scanning placeholders **document-wide** rather than only text-element values would
**reject the canonical fixture on its first corpus run**. The constraint must be **stated** (AC20)
rather than left as an accident of which walker happens to be used.

**This is the second time the canonical fixture has caught a scoping error before it shipped** — the
first being `{{page}}`/`{{pages}}` under D-1.6.5. That is a strong argument for keeping it in the
**default gate** rather than behind a build tag.

## Epic 1 decisions — Story 1.7 (ruled before the story file was written)

### D-1.7.1 — Five positional arguments survive; D-1.1.c's medium-confidence flag closes on AD-8
**Orchestrator decision**, on the lead's ruling.

**Verdict** *(mechanism: binding)*: `Render(t *Template, d Data, p Params, f FontSet) ([]byte, error)`
and `RenderTo(w io.Writer, t *Template, d Data, p Params, f FontSet) error`. **No options struct.**
D-1.1.c's medium-confidence flag on argument packaging is **resolved to high and marked closed**, so
nobody reopens it at the `v0.1.0` tag.

**The decisive argument is one D-1.1.c did not have.** An options struct makes `FontSet`
**omittable at compile time**: `folio.Request{Template: t, Data: d}` would compile and carry a zero
`FontSet`, turning an **AD-8 violation from a compile error into a runtime one**. AD-8 is explicit
that the font set is part of the render's *identity* — *"the same template with a different chain is
a different render, not a silent substitution."* Positional arguments force every caller to name it.
That is not style; it is the mechanism by which AD-8 is enforced by the compiler rather than by
discipline.

The extensibility counter-argument is weak here: AD-8 and AD-23 fix the input set at exactly four,
and the spine's Deferred list describes the REST service as *"another shell around the same core"* —
it would not change this signature.

**The `FontSet` shorthand — D-1.1.c still governs.** The AC writes `RenderTo(w, template, data,
params)` and omits `FontSet`; **the spine governs, the AC is shorthand.** Carry the six-argument form.

**But the ceremony question cannot be fairly judged at 1.7, and this needs saying.** AC4's test is the
README — and **verified: there is no README anywhere in the repo; 1.7 writes the first.** The
first-PDF example needs a `FontSet`, and **verified: `folio-go/fonts/` does not exist** — the shipped
faces arrive at **Story 2.2**. So at 1.7 the example must show a caller assembling a font set from
their own bytes, which *is* ceremony — **ceremony that 2.2 removes, not ceremony the signature
causes.** Therefore: the packaging decision closes now on AD-8; **the AC4 ceremony judgement is
re-tested at Story 2.2** (registered in `deferred-work.md`, owner 2.2) *(mechanism: binding)*. If the
1.7 README author finds the example reads as ceremony **for reasons other than the missing font
set**, that is a `DECISION NEEDED`, not a unilateral change.

### D-1.7.2 — The `Render`/`RenderTo` identity test is vacuous as stated; assert delegation structurally and test what the writer path really gets wrong
**Orchestrator decision**, on the lead's ruling.

**Streaming is not available and must not be attempted** *(mechanism: binding)*. AD-7 makes `/ID` a
SHA-256 **over the serialized body**, and the classic xref table carries **byte offsets**. Both need
the complete document before the trailer can be written. So `RenderTo` **necessarily** builds the
whole document in memory and writes it once. *"No buffering of the whole document"* is a property we
**cannot have**, and asserting it would assert something false. **Put that reasoning in a code
comment** — someone will eventually try to "optimise" `RenderTo` into a stream, and it would break
AD-7 silently.

**Verdict — three parts** *(binding for the properties; illustrative for the AST form)*:
1. **One core produces the bytes.** `Render` returns them; `RenderTo` writes them. Two emitters would
   be D-1.4.14's drift hazard, so a shared path is the **correct** design — which is exactly why the
   behavioural identity assertion **cannot fail**.
2. **Assert the delegation structurally, not behaviourally** — an AST assertion that exactly one
   function produces the document bytes and both entry points route through it. Byte-identity then
   holds as a **theorem** rather than as a test that cannot fail. **Delete the vacuous behavioural
   comparison rather than keeping it for comfort — a green test nobody can fail reads as coverage.**
3. **Test what the writer path genuinely gets wrong**, none of which the shared core guarantees: a
   writer returning an **error mid-write** must surface it, not return `nil`; a **short write** must
   be detected and reported (Story 1.1's `TestMain` learned this the hard way, Nit 25); the **exact
   byte count** written equals the document length, via a counting writer; and **nothing extra** is
   written — no trailing newline, no second call with a tail.

That last set is where bugs in a writer API actually live, and every one of them can fail.

### D-1.7.3 — The render-file guard locates by AST, and 1.7 is the story that proves it survived the move
**Orchestrator decision**, on the lead's ruling. **States explicitly what Story 1.6's blocker showed
could not be assumed.**

**Verdict** *(mechanism: binding)*: D-1.6.3's rule must **locate by AST the file(s) declaring `Render`
*and* `RenderTo`** — by declaration, **never by filename**. **Its red-proof is exactly the Story 1.6
reviewer's mutation:** move `RenderTo` into `folio.go` (which imports `os`) and the guard must go red
**with a named finding**. Together with Story 1.3's `internal/` lint, that discharges the AC.
**Do not build a filesystem-snapshot check** — disproportionate, and it would test the OS rather than
the library.

**Why this needed restating rather than assuming.** Story 1.6's blocker was the same family one level
down: the lead ruled the *property* ("the file declaring `Render`") and the implementation asserted a
*name*. The reviewer moved `Render` into the `os`-importing file, left the empty file in place, and
the guard **passed**. **1.7 is the very story that moves the entry point**, which is the sharpest
possible demonstration that a property-shaped ruling must be restated as a requirement, not presumed
to have carried.

**The residual gap, stated rather than papered over** *(mechanism: illustrative)*: a file-scoped
import rule makes the *obvious* violation impossible but does not stop a deliberate cross-file route —
`RenderTo` calling a helper in `folio.go`, which legitimately imports `os`. **That is accepted.** The
AC's intent is a **capability** claim ("you can serve a PDF without touching disk"), not a security
boundary, and D-1.7.2's counting-writer assertions give the behavioural half. **Record the gap in the
story so a later reader does not mistake the guard for a proof.**

### D-1.7.4 — `params` is a resolution root, not a reserved token — deliberately a different layer from `page`/`pages`
**Orchestrator decision**, on the lead's ruling.

**Verdict — different mechanism, deliberately** *(mechanism: binding)*:
- `page` / `pages` are **reserved whole tokens** that Story 1.6 passes through untouched for Story 2.7
  to late-bind. AD-4 is explicit that **no `page` namespace exists**.
- `params` is a **namespace** — a second resolution root resolved at bind time, alongside the data
  root. `{{params.reportDate}}` is a dotted path into the supplied params document.

**Do not implement `params` by extending the `page`/`pages` reservation list.** They are different
things, and conflating them is how `page` eventually acquires a namespace — which AD-4 forbids
forever.

**Precedence** *(mechanism: binding)*: the first path segment `params` **always** selects the params
document, never report data, regardless of what data contains. **Retained fixture, copying D-1.6.5's
`page` fixture verbatim in shape: data containing a top-level `params` key must not change what
`{{params.reportDate}}` resolves to.**

**A top-level `params` key in data is not an error** *(mechanism: binding)*. It becomes unreachable by
any binding, and that is documented — but rejecting it would refuse legitimate caller data over a
name collision the caller may not control. **The caller's JSON shape is their business; only the
binding namespace is ours.**

**`{{params.x}}` with no params supplied → Error naming the path and element id** *(binding)*. This is
AD-14's absent-path case applied to the params root, and the same code path as absent data. Rendering
empty would **silently produce a statement missing its report date** — precisely the failure AD-14's
first case exists to make loud. **`{{params}}` bare, no dot → located error** ("`params` is a
namespace, not a value") *(illustrative for the wording)*.

### D-1.7.5 — `Params` is a defined type sharing one decode path, with a compile-time swap proof
**Orchestrator decision**, on the lead's ruling.

**Verified:** `type Data []byte` at `render_entry.go:52` is a **defined type, not an alias**.
`type Params []byte` must be the same *(mechanism: binding)*. **An alias (`type Params = []byte`)
would make the two mutually assignable and destroy the entire point** — D-1.1.c's argument was that
*"in a product whose acceptance fixture is a bank statement, that swap must be a compile error, not a
support ticket."* **At 1.7 they become adjacent same-typed-looking arguments, which is the exact
moment that reasoning was written for.**

**Guardrail** *(mechanism: binding)*: a **compile-time proof that the swap is rejected** — a
`//go:build ignore` or `go/types`-based assertion that `Render(t, params, data, f)` does **not**
type-check. An assertion that the safety property holds, rather than a comment claiming it does.

**One decode path, shared** *(mechanism: binding)*: params are decoded with the **same `UseNumber`
discipline, the same single literal splitter (D-1.6.1), and the same `Decimal` type** as report data.
A second decoder would be the drift hazard refused three times already — and worse here, since
**params carry dates and amounts, where a divergence would be least visible and most expensive.**

### D-1.7.6 — The README's toolchain caveat is due now
**Orchestrator decision**, on the lead's ruling. D-1.1.a's addendum, landing at the README story.

**Verdict** *(mechanism: binding)*: the README states that a **`toolchain` directive binds only the
main module** — so an integrator compiling `folio-go` into their own binary builds with **their**
toolchain and may get different bytes. That is AD-22's accepted position, **but it is invisible until
it bites**, and **counter-metric C5's "first PDF in minutes" integrator is exactly the person who hits
it.** It belongs **near the golden-hash discussion, not in a footnote**: someone recording Folio's
hashes in their own test suite needs to know what pins them.

### D-1.7.7 — AD-7's `/CreationDate` clause is already owned by Story 3.7; 1.7 re-scopes one test and adds one tripwire
**Orchestrator decision**, on the lead's ruling. Answers the Story 1.7 creator's `DECISION NEEDED`.
**Not a gap, not a contradiction, and not an owner escalation.**

**Verdict.** Story 3.7 already owns the work, with acceptance criteria written. **Verified by the
orchestrator at `epics.md:1074–1100`:** its user story reads *"…and to **pin document dates
reproducibly**"*, its **Covers** line names **AD-7**, and it carries these criteria verbatim:

> **Given** `SOURCE_DATE_EPOCH` set in the environment · **When** `cmd/folio render` runs · **Then**
> the CLI reads it and passes it in as a parameter · **And** the library core still reads no
> environment variable
>
> **Given** **no date supplied by any route** · **When** the PDF is produced · **Then**
> `/CreationDate` and `/ModDate` are omitted

**The lead's tell, and it is a good one:** *"no date supplied by **any route**"* is the **negative case
of a positive case** — nobody writes that sentence unless a route exists. Story 3.7 is also the story
that builds `cmd/folio`, the only component NFR1.f names as reading `SOURCE_DATE_EPOCH`.

**On the orchestrator's proposed dissolution — the distinction is correct but does not dispose of the
clause.** `acceptance.md:16`'s "generated date" is **footer text** (a `{{params.generatedDate}}`
binding rendering as visible content, which 1.7 delivers); AD-7's clause governs the **metadata
dictionary**. Genuinely different surfaces. But it does **not** follow that the metadata clause is a
permission never exercised, because **NFR1.f is a stated requirement** (`prd.md:370`): the
`SOURCE_DATE_EPOCH` convention *"is honoured… by Folio's own command-line tooling, which read it and
pass the date in as a parameter (FR21)."* `SOURCE_DATE_EPOCH` exists for exactly one purpose, so a
route that reads it and delivers it nowhere would leave NFR1.f **honoured in letter and void in
effect.** Both surfaces are real; neither substitutes for the other.

**Why 3.7 rather than 1.7** *(mechanism: illustrative)*: a PDF date is `D:YYYYMMDDHHmmSSOHH'mm'`, and
date formatting matures at **Story 3.4**. Building a PDF date emitter at 1.7 — before any date type,
before `internal/diag`, before the CLI that supplies the value — would be premature in three
directions at once.

**Two things Story 1.7 does, both small** *(mechanism: binding)*:
1. **Re-scope the shipped test instead of leaving it accidentally right.**
   `TestRenderHasNoCreationOrModDate` asserts absence **unconditionally**, and that is true today only
   because nothing *can* supply a date. From 1.7 it becomes a statement about a render **whose params
   carry no date** — which is precisely 3.7's fourth AC, and is testable now that `Params` exists.
   Rename and re-scope it so it asserts the negative case **deliberately**. Left unconditional,
   **3.7's developer meets a red test and the cheapest way out is to weaken it.**
2. **Add an absence tripwire on `folio-go/cmd/`** alongside the existing `internal/expr` and
   `internal/diag` rows. `cmd/folio` does not exist until 3.7, so **its creation is the exact moment
   the `/CreationDate` wiring must be settled.** D-1.3.4's mechanism, one row in a table that already
   exists.

**De-risking note for 3.7** *(mechanism: illustrative)*: the "every golden hash moves" fear is smaller
than it looks. **Only fixtures that supply a date would move**; a params-carrying render with no date
is byte-identical to today. So 3.7's blast radius is new fixtures plus any existing one that opts in
— not the corpus.

### D-000.14 — When a count is load-bearing, count by AST; a text or filtered count measures the filter
**Orchestrator decision**, on the lead's ruling. **Program-wide standing rule** *(mechanism: binding)*.

**Verdict.** Whenever a count is load-bearing — call sites, guarded files, findings, fixtures — it is
produced by an **AST-exact** count, never by text matching and never through a filtered pipe.

**Measured, on one question, three ways.** The Story 1.7 creator counted `Render`/`RenderTo` **call
expressions**: **19** by AST (all in tests, **zero production callers**); **36** by raw text matching;
**8** through a piped `rtk` count. `matrix_test.go` mentions `Render(` in prose and contains no call
expression at all.

**Why this is D-000.12 generalised past hashing.** A filtered or text-matched count is **not a
measurement of the property — it is a measurement of the filter.** D-000.12 established that the
shell wrapper corrupts bytes through pipes; this extends the same caution to counts, where the wrong
answer arrives looking exactly like the right one and **nothing announces the discrepancy**.

**Pattern, not incident.** Story 1.6's equivalent claim was wrong **twice in the same way** ("thirteen
call sites", corrected to 11). Three methods, three answers, and only one counts the thing anyone
means.

### D-1.7.1 (amended) — the narrated parameter counts were wrong; the verbatim signatures govern
**Orchestrator decision**, on the lead's own correction. **Amends D-1.7.1's prose only; its verdict
stands unchanged.**

**Verdict.** `Render(t, d, p, f)` takes **four** parameters; `RenderTo(w, t, d, p, f)` takes **five**.
D-1.7.1's prose said *"five positional arguments survive"* (echoing D-1.1.c's own loose phrasing) and
then *"carry the six-argument form"*. **Neither count is right.** The verbatim signatures in the
verdict govern; the narrated counts were wrong. Surfaced by the Story 1.7 finisher as F9 per D-1.2.6,
not arbitrated.

**What this is an instance of.** The lead **narrated a count beside a verbatim artifact and the
narration drifted from it** — the identical failure as the reviewer's prose summary disagreeing with
its own disposition table in the same story. **Four instances in two days**, all in the same
direction.

### D-000.14 (extended) — A total narrated beside the artifact it summarises is a second source of truth
**Orchestrator decision**, on the lead's ruling. Extends D-000.14 *(mechanism: binding)*.

**Verdict.** A disposition table's total is **computed, or omitted entirely** — never narrated
alongside it. Same for any count stated beside the artifact it summarises.

**The evidence is one story deep.** On Story 1.7 the same disposition table was counted three times:
developer **29/29 clean**; reviewer's mirror **18/29 with 11 disagreements**; the finisher, recounting
the reviewer's own ✅/❌/⚠️ marks, **17 agree with 12 disagreements** — *the reviewer's prose summary
did not match its own table.* **All three were wrong, and all three undercounted the gap.** Add the
lead's own D-1.7.1 prose (above) and that is four in two days.

**Why this is D-000.14's own lesson landing on the disposition mechanism.** D-000.14 established that
a text-matched count measures the regex and a filtered count measures the filter. This is the
degenerate case: **a count that measures nothing at all**, because it was typed rather than derived —
and it sits on top of the mechanism built to catch unread rows. **A number in prose beside a table it
summarises has no guard.**

## Epic 1 decisions — Story 1.8 (ruled before the story file was written)

### D-1.8.1 — JPEG and PNG by passthrough; alpha, palette, 16-bit and interlaced are load errors; `mediaType` stays open
**Orchestrator decision**, on the lead's ruling.

**Verdict — what ships** *(binding for the format set and the failure mode; illustrative for the PDF
filter details)*:

| Input | Route | Why |
|---|---|---|
| **JPEG** | `/DCTDecode`, bytes embedded **unchanged** | zero re-encoding, byte-stable by construction |
| **PNG** 8-bit RGB/Gray, non-interlaced, **no alpha** | `/FlateDecode` + `/DecodeParms /Predictor 15`, IDAT zlib stream **passed through** | PNG's own filtering maps onto PDF's predictor; **no compressor is invoked** |
| PNG with alpha, palette, 16-bit or interlaced; any other media type | **Load error**, located, naming the asset key and the reason | — |

**The passthrough design is the ruling's real content.** Story 1.1 deliberately kept `compress/flate`
out of the output, calling its arrival *"a hash-changing, versioned event under AD-22 and the subject
of carried risk **R4**"*. **Verified: `acceptance.md:83` rates R4 Medium — *"Compressor output is
stable by observation, not by contract — a future Go release could invalidate every recorded golden
hash downstream."*** Decoding and re-encoding a PNG would invoke our compressor and **materialise R4
in the last story of Epic 1**. Passing the existing zlib stream through means the compressed bytes
come **from the file, not from a compressor** — determinism is inherited from the input and R4 stays
closed. **Take this route even where re-encoding would be simpler.**

**Load error, not render error** *(binding)*: Story 1.4's posture is that a bad template fails in CI,
and `folio.Validate` (Story 3.7) must catch it without rendering.

**`mediaType` must NOT be declared a closed set** *(binding)* — the subtle half. Under **D-1.4.12** a
closed set can only be extended by a **MAJOR** bump, so freezing `mediaType` now would make "add WebP"
a breaking format change **forever**. Distinguish two failures instead: **format validity** (the
`.folio` is well-formed; `mediaType` stays open) versus **library capability** (*this version* cannot
render that media type, reported as a located error saying exactly that). Adding a format later is
then a **capability gain, not a format change**. This distinction will recur for font formats and
expression functions, so it is named once here.

**Alpha is out of scope — and goes to the owner at the Epic 1 boundary report** *(illustrative)*.
AD-6 excludes "transparency groups"; an image `/SMask` is not literally one, so this is a **scope
call, not an invariant**. Ruled out because supporting it forces decode-and-re-encode — the R4
exposure above. **But a transparent PNG is the single most common form a company logo takes, and the
golden report's header carries a logo (Epic 4).** Not a one-way door — `/SMask` is purely additive
later — so it does not block. **Surfaced to the owner while the golden report is still being
designed**, not discovered at Epic 4.

### D-1.8.2 — The 76-column split is canonical output; input wrapping is normalised
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**On serialize:** standard base64 (RFC 4648 §4, **with** padding, not URL-safe), split into elements
of **exactly 76 characters**, the final element carrying the remainder (1–76). **Padding falls where
it falls — inside the split, at the end of the last element.** Never stripped, never moved, never
given its own element.

**On parse:** accept **any** wrapping. Concatenate elements with nothing inserted, then decode
strictly. A file hand-wrapped at 64 columns **loads** and **re-serializes at 76**.

**That second half follows directly from D-1.4.3 and must not be tightened:** non-canonical-but-valid
input is accepted and normalised; canonical is the fixed point (P3). **A parser that enforced 76
columns would reject a legal hand-edit**, and AD-9's Prevents line explicitly contemplates a
hand-editor. Invalid base64 (bad alphabet, bad padding, embedded whitespace) is a **load error**, as
is **an asset whose decoded data is empty** — it cannot render, and its key would be the SHA-256 of
nothing.

**Round-trip obligation:** the corpus gains an asset-bearing fixture satisfying **P1, P2 and P3** — and
**a hand-wrapped variant that normalises to the same canonical bytes is the non-vacuous half**, since
a canonical-in/canonical-out test alone would pass on a serializer that merely echoed its input.

### D-1.8.3 — Key over decoded bytes, validated on load; `/Width`/`/Height` → `appendInt`; orphans preserved because P1 forces it
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

1. **The key is the lowercase hex SHA-256 of the *decoded* bytes.** **Verified: `folio-format.md:219`
   says "lowercase hex SHA-256 of the **raw bytes**"** — raw means decoded, not the base64 text. Two
   identical images wrapped differently must produce the **same** key, which only holds over decoded
   bytes.
2. **Validate the key on load** — recompute and require equality; mismatch is a **load error**. This
   makes dedup correct **by construction rather than by trust**, and catches a hand-edit that changed
   the bytes but not the key — the exact corruption that would otherwise surface as two elements
   silently sharing one wrong image.
3. **`/Width` and `/Height` are pixel counts → `appendInt`** (D-1.1.b's routing table). The trap it
   exists to pre-empt applies directly: routing a pixel count through `appendLength` would emit `0.8`
   for an 800-pixel image.
4. **Orphaned assets are PRESERVED — forced, not chosen.** **D-1.4.3's P1 requires
   `Parse(Serialize(d)) == d`**, so a serializer dropping unreferenced assets would violate P1 for
   every document containing one. **There is no policy latitude here.**
   - **Corollary for Epics 5/6** *(binding)*: **garbage-collecting orphans is a designer feature — an
     explicit user action — never a serializer side effect.** A save that silently deletes an asset the
     author just unlinked and is about to re-link is data loss.

### D-1.8.4 — Cross-multiply to choose the binding axis; route the centring halve through `ScaleRound` too
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*. Image `W×H` pixels, box
`bw×bh` millipoints.

1. **Choose the binding axis by cross-multiplication, never by ratio:** compare `bw*H` against
   `bh*W`. Exact integer comparison, **no division, no float**.
2. **Then one `ScaleRound` call:** width binds → `dw = bw`, `dh = ScaleRound(bw, H, W)`; height binds
   → `dh = bh`, `dw = ScaleRound(bh, W, H)`.
3. **The centring offsets are the second rounding site, and they must not be.** `(bw - dw) / 2`
   **truncates** when the difference is odd — a second rounding behaviour smuggled in as arithmetic.
   **Route it through the same function: `ScaleRound(bw-dw, 1, 2)`**, so round-half-to-even applies
   and the program keeps **exactly one rounding mode in exactly one function** (AD-2, D-1.4.5).
4. **Validate pixel dimensions at the trust boundary first** — `W > 0`, `H > 0`, bounded — **reusing
   D-1.5.2's precedent exactly**: `ScaleRound`'s panics are programmer-error guards, not input
   handling, so a `W == 0` from a corrupt header must be a **located load error, never a panic**. Same
   seam, same reasoning, one story later.

### D-1.8.5 — Reuse the fence: the no-fetch property holds by construction of the format
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*. **Add no new guard.**

**The reason is structural, not "the guards are probably enough".** **Verified:
`folio-format.md:224` — *"Images are only ever embedded. Folio never fetches by URL and never reads
from disk at render time (FR33)."*** **There is no field in the format that can name a remote or
on-disk resource.** An element's `asset` field is a **key into the in-memory `assets` map** — not a
path, not a URL. At render time the bytes are already a `[]byte` on the parsed `Document`. **There is
no code path that could fetch, because there is no value that names anything to fetch.** D-1.4.7's
key-inventory drift test already keeps the field set honest.

**One cheap behavioural check** *(illustrative)*, because "no guard needed" should still be
observable: **mutate the in-memory asset bytes and assert the rendered PDF changes.** If the renderer
read from disk, a cache, or anywhere but the document, it would not. Three lines, non-vacuous, and it
**proves the provenance claim rather than restating it**.

The existing `net` ban under `internal/` and the AST-located render-file fence stand unchanged, with
**Story 1.7's recorded residual gap carried forward as-is — a capability fence, not a proof. Do not
quietly upgrade its description in this story.**

### D-1.8.6 — A third matrix document, with an image-XObject guard on every captured stream
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

The image document **joins** `minimal-rect` and `font-text` as a **third**; it replaces neither.
`minimal-rect` is the fontless emission baseline whose four-target verification was deliberately
preserved at Story 1.5; `font-text` covers subsetting. **Replacing either would silently retire
coverage of a feature that is still shipping.**

**The guard is the direct analogue of Story 1.5's `FontFile2` lesson:** assert each captured stream
contains an **image XObject** (`/Subtype /Image`) **before** any comparison — **on every leg, not
once**. Without it the matrix would compare four identical **imageless** PDFs and report byte-identity
across all targets, which is exactly what 1.5 discovered about fonts. Each existing document keeps its
own feature guard on the same basis.

### D-1.8.1 (amended) — An unrecognised `mediaType` is never a load error; `Validate` predicts `Render`
**Orchestrator decision**, on the lead's ruling. **Amends D-1.8.1's refusal clause.** Surfaced by the
Story 1.8 creator per D-1.2.6, not arbitrated.

**This does NOT narrow owner decision D-1.4.9 — it removes a refusal the lead wrongly introduced
against it.** No owner escalation needed, and the record should say so plainly: the orchestrator's
backstop caught a lead ruling contradicting an owner decision, and the fix was to stop contradicting
it.

**The decisive test, which makes the line principled rather than pragmatic.** Under D-1.8.1 as
written, **the same bytes are valid or invalid depending on which library reads them.** That is the
definition of a capability failure — and it is disqualifying on its own terms, because
`folio-format.md` opens by calling the format *"a public contract, not an implementation detail"*,
and **a public contract whose validity depends on the reader is not a contract.** "Valid `.folio`"
must mean something stateable without naming a library version.

**Amended verdict — three surfaces, three answers** *(mechanism: binding)*:

| Surface | Unrecognised `mediaType` |
|---|---|
| `LoadTemplate` / `ParseTemplate` | **Never refused.** The document is valid; the asset is preserved verbatim (P1 requires it regardless) |
| `Render` | **Error — only when an element actually needs to draw it.** Located, naming element id, asset key and media type |
| `Validate` | Reports **what `Render` would do** |

**The orchestrator's second face is what settled it.** D-1.8.3 preserves orphaned assets **because P1
forces it** — so a valid document may carry an asset **nothing will ever draw**. Refusing the whole
file over a rendering never attempted is indefensible.

**The rule this generalises to, which matters more than this case** *(mechanism: binding)*:
**`Validate` is a dry-run predictor of `Render`, not a second rule system.** An unrenderable asset
nothing draws → `Render` succeeds → `Validate` is clean. One that is drawn → `Render` errors →
`Validate` errors, **in CI, before production** — so Story 1.4's loud-failure posture is fully
preserved *without* making validity reader-dependent. **This will matter at Story 3.7.**

**Two consequences that fall out cleanly:**
- **A hidden element never triggers it.** AD-24: *"A hidden element is **absent** from the
  `PageModel`."* Nothing draws it, so nothing errors — **no special case needed**, it follows from
  "error when the render needs to draw it".
- **`Validate` without data cannot always predict** a `visibleIf`-gated image; where data-dependence
  makes prediction impossible the diagnostic degrades to a **Warning**. Story 3.7 detail
  *(mechanism: illustrative)*.

**What the amendment does NOT relax** *(mechanism: binding)*: a **recognised** `mediaType` whose bytes
do not parse as that format — a file declaring `image/png` that is not a PNG — **is** a load error.
That failure is **reader-independent**: the file lies about itself and every library agrees it is
malformed. Cheap to check via magic bytes and header for the two supported formats. **The contrast
between these two cases is the clearest illustration of the validity/capability line.**

### D-1.8.7 — The format doc's examples were hand-authored and never executed; execute them
**Orchestrator decision**, on the lead's ruling. Second instance of the same pattern.

**Verified by the orchestrator:** the format specification's worked **asset** example decodes to a
PNG of **colour type 6 — RGB with alpha** — which D-1.8.1 rejects. It is also non-canonical under
D-1.8.2 (wrapped 60+36; canonical is 76+20) and its key is a **literal ellipsis**. **Three
independent D-000.6 amendment obligations on one example**, all landing in Story 1.8's commit
*(mechanism: binding)*.

**How the replacement is produced** *(binding)*: **generated by the shipped serializer** and
**independently cross-checked by a second tool** (CRCs, colour type, digest), per D-1.4.10's
discipline. The creator wrote its candidate in as **"a candidate to re-derive, not as truth"** — that
instinct is correct and is honoured.

**Then close it permanently with a mechanism that already exists** *(binding)*: **D-1.4.11 asserts the
doc's fenced worked-example block is byte-identical to its golden fixture — extend that to the asset
example.** After this the doc example cannot rot again.

**The pattern, named because this is the second time** *(illustrative)*: `folio-format.md`'s examples
were **hand-authored and never executed**. DN-1 found the worked example was **not a fixed point**;
now the asset example is invalid **three ways**. **Every fenced block claiming to be complete or
canonical should be executed, not read.** Fragments can at least have their field names covered by
D-1.4.7's drift test.

### D-1.8.8 — Three shipped fixtures carry invalid asset keys; validate shape AND value
**Orchestrator decision**, on the lead's ruling. **D-1.8.3's validation working on day one.**

**Verified by the orchestrator:** the asset keys in the shipped fixtures are **64×`a`**, **64×`b`**,
and one of **62 characters** — not digest-shaped at all. **Nothing noticed**, across four stories.

**Two additions to D-1.8.3** *(mechanism: binding)*:
- **Validate shape *and* value**: 64 lowercase hex characters **and** equal to the SHA-256 of the
  decoded bytes. **A 62-character key means no check validated shape either** — a different error
  class from a wrong-but-well-formed digest, and the shape check is the cheaper one.
- **Regenerating must not change what those fixtures test.** They exist for other properties
  (passthrough, unknown keys); the key fix is **incidental to their purpose and must not silently
  alter it.**

### D-1.8.9 — The serializer must re-wrap from decoded bytes; the 64→76 fixture is the discriminating test
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**Verified at baseline:** `writeAssets` emits `a.Data` **verbatim with no re-wrapping**, so
D-1.8.2's canonical-in/canonical-out half **passes vacuously today**.

The serializer **re-wraps from the decoded bytes, never echoes the input array**. The non-vacuous
half is the one D-1.8.2 already named: **a fixture wrapped at 64 columns that must serialize to 76.**
**Canonical-in/canonical-out cannot fail against an echoing serializer; 64-in/76-out cannot pass
against one.**

### D-1.8.10 — AD-24's single-flip clause gets its first real test; make the guard positive
**Orchestrator decision**, on the lead's ruling. **Applying a binding spine invariant that no ruling
contradicts is not arbitration** — the creator was right not to treat it as such.

AD-24 is unambiguous: *"The flip to PDF user space happens in **exactly one** function in
`internal/pdf`, and nowhere else in the module inverts a coordinate."* An image path needs the same
flip; open-coding it is a **direct AD-24 violation**, and no D-1.8.x ruling mentions it because none
needed to.

**Two refinements:**
- **Make the guard positive, not negative** *(binding)*: assert that **all** content-stream coordinate
  emission routes through the one flip function — **not** that "nobody else writes a minus sign",
  which is unfalsifiable and will produce false positives. Same shape as D-1.1.b's `numbers.go` rule.
- **Name what the finding actually is** *(illustrative)*: the invariant was never **enforced**, only
  **unexercised** — one content-stream builder cannot violate "exactly one". Same
  *holds-by-construction-not-by-logic* pattern as the vacuous re-wrap, and **Story 1.8 is the first
  real test of AD-24's clause.** Worth saying in the story, because **"it was always satisfied" reads
  as evidence it was guarded when it is evidence it never could have failed.**

### D-1.8.11 — The licence manifest's extension allowlist is the rotting-list pattern; invert it
**Orchestrator decision**, on the lead's ruling. **Recorded for the Epic 1 boundary gate**, not
expanded into Story 1.8.

**Verified:** the manifest scans **font extensions only** (`.ttf/.otf/.ttc`), so a committed image is
**invisible to AD-26 enforcement** — a D-000.9 shape.

**When fixed, do NOT add `.png`/`.jpg` to the list** *(mechanism: binding)* — an extension allowlist is
the rotting-list pattern and the next asset type is invisible again. **Invert it: every committed file
in the declared asset and fixture locations must be accounted for in the manifest, and anything
uncovered fails.** That is the **coverage-witness form** — ask *"is every asset accounted for?"*, not
*"is this file a font?"* — and it is **the same fix that closed `ScanAbsences`**.

### D-1.8.7 (correction) — the fence/golden byte-identity mechanism is D-1.4.10's, not D-1.4.11's
**Orchestrator correction.** Found by the Story 1.8 creator. **The obligation stands exactly as
ruled; only the citation was wrong.**

D-1.8.7 says *"D-1.4.11 asserts the doc's fenced worked-example block is byte-identical to its golden
fixture — extend that to the asset example."* **Measured: D-1.4.11 is `ScanAbsences`' coverage
witness.** The real mechanism is `TestWorkedExampleMatchesGoldenFixture`, whose own comment records
it as **AC16 step 3 — D-1.4.10's**.

**The creator was right not to raise this as `DECISION NEEDED`** — a verifiable misattribution
*inside* one ruling is not a conflict *between* two, and D-1.2.6 governs the latter. Same disposition
as the D-1.7.1 (amended) precedent: the verbatim artifact governs, the narration was wrong.

**Two measured implementation facts that change how the extension is built** *(mechanism: binding)*:
the existing extractor takes the **last** fence under `## Worked example`, so the asset example needs
a **second extractor, not an edit**; and the asset example is a **fragment**, not a whole document, so
the story must pick a golden shape for it and say which.

**Third narrated-artifact drift in three stories** (D-1.7.1's parameter counts, the disposition-table
totals, this). D-000.14 (extended) already binds; this is another instance, not a new rule.

### Epic 1 gate finding — Template alias closed: opaque handle
**Carried work, resolved.** `folio.Template` migrated from `type Template = template.Document` (an
alias, making 32 exported `internal/template` declarations reachable and permitting field-by-field
construction that bypassed `ParseTemplate`'s validation) to `type Template struct { doc
*template.Document }` — construction only through `LoadTemplate`/`ParseTemplate`, no exported
fields, no accessors added pre-emptively.

**Verified by go/types-backed AST analysis** (not text search, and not bare-identifier matching,
which produces false positives like `runtime.Version()`): a `golang.org/x/tools/go/packages`-loaded,
fully type-checked scan of the pre-migration tree (commit `614bc99`, via a throwaway git worktree)
for every `ast.SelectorExpr` whose receiver type-checks to `*internal/template.Document` (unwrapping
the `Template` alias, which the current Go toolchain preserves as a distinct `types.Alias` node —
`types.Unalias` was required to see through it) finds **50 field-reach sites in package `folio`**:
44 in `render.go`, 4 in `template_test.go`, 2 in `render_image_test.go` — matching the pre-migration
finding's named files exactly, though at a finer grain than the finding's own "~10" estimate (the
finding's number was a rough per-call-site count; this one counts every AST selector node, the same
measurement-method variance D-000.14 already documents for Story 1.7's swap proof, 19/36/8). A
second AST pass over the rest of the repository (excluding `folio-go` itself) found **zero files
importing `github.com/panitw/folio/folio-go`** — zero external callers, confirmed by parsing import
specs, not by grep. All 50 in-package sites were updated to reach the unexported `doc` field
directly (same package); `go build`/`go vet` on the migrated tree independently confirm no site was
missed, since any leftover direct field access on the now-opaque `Template` fails to compile.

**Compile-time proof added** (`folio-go/templateopaque_test.go` +
`testdata/templateopaque/{good,bad}/`, same shape as Story 1.7's `paramsswap_test.go`):
`folio.Template{Version: "1.0"}` from outside the package fails with `unknown field Version in
struct literal of type folio.Template`; the `good` control (construction via `ParseTemplate`)
builds. Confirmed by actually running `go build` on both fixtures, not by comment.

**Gates, all green, raw exit codes confirmed via `rtk proxy`:** build/vet/gofmt clean in
`folio-go/`, `hashmatrix/`, `lint/` (hashmatrix's documented `go build ./...` "probe" directory
collision is pre-existing and unrelated); `lint` full suite green under `GOPROXY=off`; `go build
-tags=matrix` and `go vet -tags=matrix` clean; full four-target matrix (darwin/arm64,
linux/amd64, linux/arm64, js/wasm) over all three documents byte-identical to the pinned
fixtures — `minimal-rect` `0f925e1b…` 547 B, `font-text` `dcd453a1…` 22299 B, `image-embed`
`e5778eb8…` 995 B, unchanged. `folio-go` full suite: 185 top-level / 289 with subtests / 7
packages (184/288 plus the one new red-proof test added by this change), zero failures.

## Epic 2 decisions — opening (ruled before Story 2.1 was written)

### D-2.0.1 — The `ja` font gap does not exist; AD-12 is a formatting invariant, not a font one
**Orchestrator decision**, on the lead's ruling. **Retracts a flag carried since the first grounding
pass. No owner escalation.**

**Verified by the orchestrator:** AD-12's scope line reads **`Binds: internal/expr · FR18, FR21,
FR43`**, and its Rule is about *"a compiled, embedded, versioned **locale table**"* for `en`, `th`,
`zh-Hans`, `ja`. **It binds formatting — dates and numbers — and says nothing about fonts. It does
not bind `internal/fontset`.** A `ja` document needs a locale-table entry, not a face. And NFR3 /
CAP-14 require **"Latin, CJK, and Thai text"** — script classes, not a named Japanese face.

**The lead's own account of the error, recorded because the pattern is now specific enough to act
on:** *"That is the third time this run I have asserted a cross-domain implication from an invariant
without reading its `Binds` line"* — base-36 from the worked example, D-1.3.1's `time` ban applied to
a module-root test, and now this. **Adopted fix: before invoking an invariant outside the area being
ruled on, read its `Binds` line first.** This joins D-1.4.15's mechanism tagging and the
verify-location rule as the lead's third self-imposed correction.

**The residual finding is real but smaller** *(mechanism: illustrative)*. Noto Sans SC is a **Pan-CJK**
face: it carries kana and the ideograph set, so Japanese text **renders** — but with **Simplified
Chinese glyph forms** where JP and SC conventions differ. **That is not tofu, and AD-8's
missing-glyph diagnostic will not fire, because coverage exists.** It is a silent quality compromise
visible only to a Japanese reader.

**Verdict: disclose, do not build** *(mechanism: binding)*. Nothing requires a JP face; adding Noto
Sans JP would add ~7 MB against Story 5.4's itemised "CJK 7.4 MB" and its ~9 MB budget — in a story
whose AC exists specifically to explain *why CJK dominates*. **Not an escalation — a stated
limitation**, recorded in Story 2.2 and surfaced in the **Epic 2 boundary report** so the owner sees
it at their observation point (D-000.3). *This owner has consistently accepted a stated gap over a
false guarantee; what they should not get is silence.*

**`folio-go/fonts/` confirmed** *(mechanism: binding)*: D-000.5 governs over Story 2.2's `fonts/`
wording. The reason is mechanical — **`go:embed` cannot reach outside its module**, so a repo-root
`fonts/` would not fail loudly; it would embed **nothing** and surface as a font-resolution failure
at render.

### D-2.0.2 — Story 2.1's spike gate is pre-stated and absolute; the obvious pass condition is circular
**Orchestrator decision**, on the lead's ruling.

**The trap, first, because it would have bitten.** The pass condition **cannot** be "agreement with
the S4 fixture" — Story 2.1's own AC says the hand review *"is kept as the **basis of** the S4
fixture"*. **S4 is an output of this spike, not an input.** Gating the spike on a fixture it produces
is circular and **would confirm automatically**.

**Verdict — AD-25's two absolutes, plus the proper-noun absolute, plus a corpus floor, all written
into the story file BEFORE the spike runs** *(mechanism: binding)*:

| Condition | Threshold |
|---|---|
| No break inside any Thai character cluster | **Zero. Absolute.** |
| No interior break in any run the dictionary cannot cover | **Zero. Absolute.** |
| **No hand-identified proper noun in the corpus is split** | **Zero. Absolute.** |
| Trie loads and queries correctly under `js/wasm` | Binary |
| **Corpus floor** — minimum count of distinct Thai personal names, place names, transaction descriptions | **Pre-stated, non-zero** |

The first two are absolute because AD-25 makes them so: *"two constraints sit **under** whatever the
dictionary proposes, and both override it."* Not threshold-able.

**The third is what actually retires R2**, and it comes from AD-25's Prevents line: *"the unrecognised
words are **customer names**… Silently mis-breaking a person's name across 50,000 statements is a
worse failure than any overflow."* **One split name is a deviation.**

**The corpus floor is the anti-vacuity clause** *(mechanism: binding)*: without a pre-stated minimum,
**a three-word corpus passes all three absolutes trivially** — a spike that proves nothing while
reporting success, which is this program's dominant defect class arriving in a story with no code to
guard. **Zero proper nouns examined is a failure, not a pass.**

The dictionary-disagreement rate is **recorded, not gated** *(illustrative)* — there is no
pre-existing ground truth to gate against — but a threshold is pre-stated above which it escalates.

**Who decides** *(mechanism: binding)*: **confirmation is the lead's** (the conditions were
pre-committed; the measurement either meets them or does not — mechanical). **Deviation is the
owner's** — re-planning seven stories is a scope decision with lasting consequence.

**The meta-rule that makes this work** *(mechanism: binding)*: the pass condition is written into the
story file **before the spike runs**, and **changing it after seeing the data is a `DECISION NEEDED`,
never a judgement call.** Otherwise a spike always confirms.

### D-2.0.3 — DW-9's re-test is in scope at Story 2.2, and the fair test is a compiled, executed example
**Orchestrator decision**, on the lead's ruling *(mechanism: binding for the criteria; illustrative
for naming)*.

**In scope at 2.2** — the story that makes the honest example writable, which is why DW-9 was parked
there.

1. **The README's first-PDF example must be a Go `Example` function** in `example_test.go`, **compiled
   and executed by `go test`** — not prose, not a fenced block nobody runs. AC4's claim is about what
   an integrator must actually write, **so it must be verified by writing it** — and that makes it
   un-rottable, the same move as the doc-block-equals-fixture assertion.
2. **"Nothing ceremonial" means:** zero to PDF bytes takes **a load call and a render call**, with the
   `FontSet` obtained in **one expression taking no arguments**. If a caller must enumerate faces,
   assemble a fallback chain, or read files before rendering, **that is ceremony and the re-test
   fails**.
3. **No shortcuts that hide required steps** — no `_ = err`, no elided setup. The example shows
   everything an integrator must type, or it is not measuring the thing.

If it fails, that is a `DECISION NEEDED` on the API surface — the option D-1.1.c reserved, still
available because nothing binds until `v0.1.0`.

### D-2.0.4 — AD-5's arrow becomes live for the first time at Story 2.5, and must be guarded in its first commit
**Orchestrator decision**, on the lead's ruling. **The most consequential item in this batch.**

**Confirmed** *(binding)*: `{{page}}`/`{{pages}}` stay reserved and untouched until **Story 2.7**;
Story 1.5's provisional band-origin assertion **dies at Story 2.5**, and its death is 2.5's AC being
met.

**Verified by the orchestrator:** `folio-go/internal/` today holds `bind`, `fontset`, `geom`, `pdf`,
`template`. **There is no `layout` package** — so **AD-5's arrow has never existed and has never been
guarded.** The very first line of this program's grounding was:

> *"One dependency arrow matters more than the rest: **there is none from `internal/layout` to
> `internal/pdf`**. That absence is what keeps PNG/SVG/HTML renderers possible later (AD-5), and it
> is precisely the arrow a well-meaning commit will try to add."*

**Story 2.5 must ship an import-direction guard in the same commit that creates `internal/layout`**
*(mechanism: binding)* — `internal/layout` may not import `internal/pdf`, with a **retained violating
fixture** proving it fires, per D-1.3.3's two-caller shape. **The guard must exist from the package's
first commit, not be added later** — the window in which the arrow is easy to add and invisible is
exactly the window before the guard exists.

**Three more Epic 1 commitments this epic collects** *(mechanism: illustrative)*:
- **AD-24's Y-flip gets its second content-stream builder.** Band placement is a **translation, not
  an inversion** — it must not add a second flip. Story 1.8's positive guard should catch it; **2.5
  is its first real test against a genuinely different caller.**
- **The content-band height derivation** — `folio-format.md` says `content` has no `height` and it is
  derived *"by one function"*, the same "exactly one function" shape. **Guard it positively at 2.5.**
- **R4 stays shut, and the distinction is subtle enough to lose.** Epic 1 emits `/FlateDecode` as a
  **filter name** on the image passthrough but **never invokes a compressor**. Nothing in Epic 2
  requires compression. **Introducing content-stream compression is a hash-changing, versioned event
  under AD-22 and must not happen incidentally** — if a story wants it, that is a `DECISION NEEDED`.

**AD-4 also gets its first real test** at 2.5/2.6 — *"two passes, and the second one lays nothing
out."* **Pagination is where a second layout pass is most tempting.**

### D-000.12 (corrected) — A mitigation for a measurement hazard must itself be measured
**Orchestrator decision**, on the lead's own correction *(mechanism: binding)*.

**D-000.12's stated mitigation — "write to a file and read it back" — is INSUFFICIENT.** Measured: the
wrapper **rewrites the `go` command itself**, so a plain redirect captures its *summary*, not the
output. A matrix verification reported **0 of 12 expected hash lines while simultaneously reporting a
pass**. **The correct form is `rtk proxy` first, THEN redirect.**

**The lead's own account, recorded because the general rule is the valuable part:** *"I proposed a
mitigation from reasoning, in a domain where reasoning had already failed three times… **a mitigation
for a measurement hazard must itself be measured before it is relied on** — otherwise it is a second
unverified instrument stacked on the first."*

**Fourth false reading from this wrapper**, and the first where the **fix written after the third was
itself never tested against the thing it was fixing.**

**Related, same ledger:** the `Template` migration measured **50** field-reach sites against the
lead's *"~10"* — a **`grep`-shaped estimate presented as a measurement**. The under-count has a
mechanical cause worth keeping: **the Go toolchain preserves a type alias as a distinct `types.Alias`
node**, so resolving it requires an explicit `types.Unalias` call. D-000.14 again.

## Epic 2 decisions — Story 2.1

### D-2.1.1 — 2.1 is not a D-000.4 override; the S4 fixture must give the deferred Docker legs something to measure
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**Not an override.** The override clause is scoped to *"a story whose **own deliverable** is
hash-shaped"*, and **2.1 emits no PDF bytes** — three matrix legs would have nothing to measure. AC3
needs **execution** under `js/wasm`, a **target-capability** question, and that leg costs **no Docker
boot** — precisely the cost the owner traded away in D-000.4. **js/wasm in-story; Docker legs to the
Epic 2 gate.**

**The guardrail, without which the deferral is empty** *(mechanism: binding)*: the computed break
opportunities must be **asserted against the checked-in S4 fixture in an ordinary test that runs on
every target**. Then the gate's `linux/amd64` and `linux/arm64` legs **genuinely exercise 2.1's work**
— a target computing different breaks fails the fixture — rather than merely compiling it.
**Otherwise the Docker legs are deferred to a gate with nothing to check.**

### D-2.1.2 — Route (b), and the criterion is fail-closed versus fail-open
**Orchestrator decision**, on the lead's ruling. Answers the creator's `DECISION NEEDED`.

**All three facts verified independently** (orchestrator and lead): `fontExtensions = {".ttf",
".otf", ".ttc"}` at `manifest.go:86`; `Render(rows) + RenderAssets(assetRows)` — **fully generated**,
so a hand-written row is **erased on regeneration** and `TestManifestUpToDate` reddens immediately;
and **`CC0` appears nowhere in the `lint` module.**

**Verdict: route (b)** — a minimal generated addition scoped to declared asset locations.

**The criterion the creator asked for, and it is a single question** *(mechanism: binding)*:

> ***Does the mechanism FAIL on a file it does not recognise, or IGNORE it?***

- **Fail-closed** — walk the declared asset locations, require every file to be **accounted for**,
  **fail on anything uncovered**. That is D-1.8.11's inversion, and a minimal scope is its **first
  increment**.
- **Fail-open** — add `.txt`/`.trie` to `fontExtensions`. That is **the forbidden shortcut**, and
  **scope has nothing to do with it**.

**So (b) is legitimate if and only if it is fail-closed. The number of locations is negotiable; the
shape is not. A one-location fail-closed walk is strictly better than a ten-extension fail-open
list.**

**Why not (a):** expanding a **spike** into a lint-architecture story in the epic's opening slot is
real scope creep, and the full inversion needs decisions about what counts as an asset location that
this story has no business making.

**Why not (c), named as the creator asked:** it is **mechanically green while substantively
incomplete** — the absence-reads-as-success shape, this program's dominant defect class. And the
Epic 1 gate's justification for deferring was *"latent, **zero current instances**"*. **Story 2.1
destroys that ground. A gap deferred for having no instances, then kept deferred once it has one, is
a gap nobody ever closes.**

### D-2.1.3 — CC0-1.0 joins `permissiveSPDX`; and the general rule that distinguishes a fail-safe from a rotting list
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**CC0-1.0 is a public-domain dedication and belongs in `permissiveSPDX`.** In scope regardless of
which route the manifest question took. Two-direction red-proof: **a CC0 asset passes, an unknown
licence fails.**

**The general rule, which the creator identified and the lead affirmed:**

> **An allowlist whose miss is LOUD is a fail-safe. An allowlist whose miss is SILENT is a rotting
> list.**

`permissiveSPDX` is the first — an unrecognised licence classifies as `unknown`, and **D-1.3.4
deliberately made unknown a build failure**. `fontExtensions` is the second — an unrecognised
extension is simply **invisible**. **Same data structure, opposite failure modes.** Someone will
eventually cite D-1.8.11 to argue against AC8; **the distinction belongs in the story text.**

### D-2.1.4 — The corpus floors, confirmed, with a strengthening the exercise floor's own logic implies
**Orchestrator decision**, on the lead's ruling.

**Item floor confirmed as proposed** *(binding)*: **≥120 personal names, ≥40 place names across
regions, ≥40 transaction descriptions, ≥200 total — computed as the sum, never narrated**
(D-000.14). The surname argument is the strongest part: **Thai law requires surnames to be coined and
unique per family**, so they are reliably out-of-dictionary and are exactly the class AD-25's
Prevents line names. Weighting that bucket largest is correct.

**Exercise floor confirmed as proposed** *(binding)* — and it is the better half, proposed by the
creator unprompted: *"a 200-item corpus in which the dictionary covers everything proves nothing
about the uncoverable-run absolute."* That is **D-000.9's coverage-witness idea applied to a
hand-reviewed corpus**, and `P6a = 0` being a **failure** is the right consequence.

**The strengthening — the exercise floor's own logic goes one step further** *(mechanism: binding)*:
among the ≥120 personal names, **a stated minimum must be names for which the *unconstrained*
dictionary proposes at least one interior break.**

**Why:** AD-25's atomic-unknown-run rule is an **override** — it sits *under* what the dictionary
proposes and overrules it. **A name that decomposes into no dictionary words is trivially unsplit, so
it exercises the override not at all.** A corpus of 120 such names would satisfy every absolute while
**never once making the override do work**. Those cases are **mechanically identifiable** — run the
greedy matcher and count — so this is a countable floor like the others, not a judgement call. Same
move as requiring **both polarities** in a fixture: the discriminating cases are the ones where the
guard has to actually override something.

**Disagreement escalation confirmed** *(mechanism: illustrative for the thresholds)*: **>15% of items**
disagreeing at any position, or **any single item with >3** disagreements named individually
regardless of rate. **Recorded, not gated** — there is no pre-existing ground truth to gate against,
the same circularity D-2.0.2 identified.

### D-2.2.0 — Variable-font instancing: measured, not obviously broken, and converted from unknown to monitored
**Orchestrator decision**, on the lead's ruling. **Measured by the orchestrator before the story was
written.** No owner escalation.

**The hazard, verified verbatim in `textshape@v0.0.15`:** `ot/gvar.go:259-260` and `:487-488` are
`deltas.XDeltas[i] += float64(xDeltas[i]) * scalar` — **the canonical fusable multiply-accumulate**,
the exact shape `hashmatrix`'s contraction probe detects — feeding `subset/plan.go:504-505`'s
`int16(math.Round(deltas.XDeltas[i]))`. **A 1-ULP difference flips a rounded glyph coordinate,
changes the outline, changes the `FontFile2` bytes, changes the hash.** That is R1 (Critical)
materialising inside a dependency, on the render path.

**And instancing is forced, not chosen:** PDF 1.7 cannot express a variable font (AD-7 pins the
profile); `subset/execute.go:503` shows variation tables are **passed through** rather than subset
when not instancing, so a 50-glyph subset would carry the whole `gvar` table; and `style.bold` is a
`wght`-axis instance.

**Measured — all four targets agree.** `AnekBangla-subset.ttf` (the only available variable face:
`fvar` + `gvar` + `glyf`, the same shape as the shipped Noto set), A–Z subset, **both axes pinned
off-default** so interpolation actually ran (`wght` 611 vs 500 default; `wdth` 109.25 vs 100),
`IsInstanced()` true. One pinned toolchain; native, `alpine:3.20` under both Linux arches, Node via
`go_js_wasm_exec`.

| Target | Hash | Bytes |
|---|---|---|
| darwin/arm64 | `ca66b20c…1488db` | 412 |
| linux/amd64 | `ca66b20c…1488db` | 412 |
| linux/arm64 | `ca66b20c…1488db` | 412 |
| js/wasm | `ca66b20c…1488db` | 412 |

**Three limits, recorded because "measured" must mean what it says** *(mechanism: binding — these go
into Story 2.2's measurement table verbatim; a later reader finding "all four agreed" without them
would over-trust it)*:
1. **Different font, same arithmetic.** Not a Noto face. The code path is identical, **the delta
   values are not**, and FMA divergence is **value-dependent** — it appears only when fused and
   unfused results straddle a rounding boundary. Strong evidence, **not proof**, for another font.
2. **`math.Round` is the amplifier, and the probe may simply have missed its boundary.** Pinned
   coordinates were arbitrary (+37% of range) and may never have produced a near-`.5` delta.
3. **All axes pinned.** Real use pins `wght` only, leaving `wdth` at default — a different scalar
   computation (`gvar.go:325`'s `scalar *=` chain runs differently when a coordinate equals its peak).

**Do not re-run the probe** *(mechanism: binding)*. **Limit 1 dominates 2 and 3:** a boundary-seeking
sweep would tighten a measurement on **AnekBangla**, and value-dependence means none of it transfers
to Noto Sans SC's deltas. **Spending more on the wrong font optimises a measurement we are about to
supersede.**

**The reframing, which is what makes further probing unnecessary** *(mechanism: binding)*: **we do not
need to prove instancing byte-stable in general. We need the shipped instances' bytes recorded as
goldens and re-verified on every matrix run.** That is precisely AD-21's purpose — *"a hash change is
investigated as a defect until proven to be an intended, versioned behaviour change."* **A
value-dependent hazard no probe can exhaustively exclude is exactly what a recorded golden plus a
repeating matrix is the right instrument for.** The residual risk is not eliminated; it is
**converted from unknown to monitored**, which is the correct end state and is reachable now.

### D-000.9 (extended to probes) — A probe reporting "no divergence" must state what would have made it fire
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*. **Program-wide.**

**The probe above had no coverage witness.** It never established that its configuration was
*capable* of producing a divergence, so **a green result is indistinguishable from "could not have
seen one."** Ask D-000.9's diagnostic question of it — *what would this probe have printed if the
arithmetic diverged on some other input?* — and the answer is **the same thing**.

**Rule:** a probe reporting "no divergence" **states what would have made it fire**, or its negative
result is recorded as **"not obviously broken"**, never as **"stable."**

**Note the distinction from the defect class itself:** the orchestrator identified the limit
unprompted, which is the opposite of the failure. **The class is not noticing.**

### D-2.2.1 — Story 2.2 is a D-000.4 override, and records a golden for EVERY (face, pinned instance) pair
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**Override confirmed on independent grounds:** not because the probe was inconclusive, but because
**the probe cannot be conclusive for faces it did not test.** The instancing arithmetic is a new hash
surface and belongs in the matrix **in-story**.

**The substantive half: 2.2 records a golden for every (face, pinned instance) pair it ships — not
one sample.** Value-dependence is the whole hazard, so **a clean result on Noto Sans at `wght`
default says nothing about Noto Sans SC at a pinned weight.** The recorded set must cover the actual
risk surface, or **the matrix is monitoring a subset of what ships.**

**Standing condition, registered in `deferred-work.md` rather than left as a one-off** *(binding)*:
**every later story that adds a pinned instance adds its golden and runs the matrix.** Bold is a
`wght` instance; if it arrives after 2.2, its story inherits this obligation — so the story that
first renders bold does not have to rediscover why.

### D-2.2.2 — The subset tag hashes glyph ids PLUS pinned axis coordinates in F2DOT14 integer form
**Orchestrator decision**, on the lead's ruling. **A correctness defect, independent of
byte-stability.**

**The collision:** two instances of one face share a **glyph set** and have **different programs**, so
AD-7's tag over the sorted glyph-id set **alone collides**. ISO 32000-1 §9.6.4 requires the tag to
distinguish subsets **within a file** — exactly what a document containing regular *and* bold
produces. Same discrimination failure as D-1.5.8's output-numbering trap, one layer up.

**The hash input** *(mechanism: binding)*: the sorted glyph-id set **plus the pinned axis coordinates
in their normalized F2DOT14 integer form** — **not** the caller-facing float values.

**Why the second half matters more than it looks:** hashing `wght=700.0` would put a **`float64` in
the subset-tag derivation path**, which AD-1/AD-23 forbid under `internal/` and which the arch guard
would catch — **but only if the identifier appears**. The F2DOT14 integers are what **actually
determined the outlines**, they are exact, and they are already computed. **Hash what determined the
bytes, in the form it was determined in.**

**Fixture** *(binding)*: two instances of the same face, **same glyph set**, must produce **different
tags** — red under a glyph-set-only tag, green under the ruled one. Same discriminating shape as
D-1.5.8's different-sets-same-size fixture.

### D-2.1.5 — `absence-cmd-dir` is re-keyed on `SOURCE_DATE_EPOCH`; a guard keyed broader than its purpose fires on the wrong thing
**Orchestrator decision**, on the lead's ruling. **The tripwire mechanism's first false positive**,
caught live against a red build.

**What happened.** Story 2.1's developer created `folio-go/cmd/gentrie/` and `folio-go/cmd/gencorpus/`
— the trie compiler and corpus tooling the spike legitimately needs. Measured:
`TestAbsencesProductionScan` **exit 1**, naming `absence-cmd-dir`. **The tooling is correct; the
tripwire was wrong.**

**The mismatch.** D-1.7.7 keyed that row on the **directory** `folio-go/cmd`, deliberately, per DW-2's
correction that *"an exact filename guess is a false-pass hazard a directory-level check does not
have."* But its **purpose** (DW-10) is narrower than its **key**: it exists to force AD-7's
params-date wiring to be settled when the CLI that reads `SOURCE_DATE_EPOCH` arrives at Story 3.7.
**`cmd/` has more than one legitimate tenant, and only one of them was the target.** The row did
exactly what it was keyed to do and nothing it was meant to do.

**Verdict — route (d), key on the trigger** *(mechanism: binding for the trigger; illustrative for the
table's representation)*:

> **`SOURCE_DATE_EPOCH` must not appear in any Go source under `folio-go/` until AD-7's params-date
> wiring is settled (DW-10).**

**Why the orchestrator's own proposal (c) was rejected — and this is the sharp part.** Asserting
*"`cmd/folio` specifically must not exist"* **must name `folio`**, so it carries the **identical
filename-guess hazard** as narrowing the key. If Story 3.7 names its binary `cmd/foliocli`, **neither
fires and DW-10 evaporates silently.** The lifecycle framing (D-2.2's fonts pattern) is right, but
**that mechanism worked for fonts because there the path *is* the obligation; here the path is only a
proxy for it, which is the whole problem.**

**Why not moving the tooling out:** `cmd/` is in the spine's **own Source tree**, so it is a described
location, not an improvisation. `hashmatrix/` and `lint/` are repo-root modules because they need
**dependencies `folio-go` must not carry** (D-1.5.1's allowlist); `gentrie` and `gencorpus` are
**stdlib-only tools inside the module**. If either ever needs a third-party dependency, the D-1.5.1
allowlist will say so loudly — that is the moment to reconsider, not now.

**Measured, so the re-key goes green immediately:** `SOURCE_DATE_EPOCH` appears in **no** Go source
under `folio-go/`, and neither new tool mentions it or `CreationDate`. Its **only** occurrence
anywhere in Go source is a **comment** at `lint/internal/rules/absences.go:77`, inside the rule's own
description — **which scoping to `folio-go/` excludes naturally, with no exemption.**

**Three obligations, in Story 2.1's own commit** *(all binding)*: replace the row and **update DW-10
with the re-keying and the reason**; **re-prove it fires** by adding `SOURCE_DATE_EPOCH` to a scratch
file under `folio-go/`, observing the red naming its rule, and removing it — **a re-keyed tripwire is
a new tripwire and inherits the obligation to be seen failing**; and because the row now matches file
**content** rather than a **path**, **the coverage witness must count both rule kinds**, or the
extension **reopens exactly the `ChecksEvaluated == 0` gap that T2 closed** — re-creating the seventh
instance of the dominant defect class inside the fix for a different one.

### D-000.15 — Key a guard on its purpose, never on a proxy for its purpose
**Orchestrator decision**, on the lead's ruling. **Program-wide standing rule** *(mechanism: binding,
in the developer and reviewer prompts)*.

> **Key a guard on its purpose, not on a proxy for its purpose. Where the key is broader than the
> purpose, the gap is where false positives live — and a false positive's danger is not the noise, it
> is the workaround it invites.**

**Its sibling, recorded earlier in this run:** *tripwires guard the arrival, but nothing guards the
removal* (which D-2.2's fonts ruling answered by converting the row to a positive assertion).

**Together they say something worth writing down: a guard erodes fastest at the point where it is
wrong**, because **the cheapest resolution to a false positive is always to weaken the guard.** That
makes key/purpose alignment matter more for guards than for ordinary code — **a false positive in a
guard is not an inconvenience, it is an attack on the guard.**

**The drill for every future tripwire:** *state the obligation in one sentence, then ask whether the
key **is** that sentence or a **stand-in** for it.* *"AD-7's params date must be wired when the CLI
reads `SOURCE_DATE_EPOCH`"* keys on **`SOURCE_DATE_EPOCH`**. It never keyed on `cmd/`.

**Process note the lead was explicit about:** the developer was told the ruling **immediately** rather
than waiting to see whether it surfaced the conflict itself. The `DECISION NEEDED` reflex is worth
reinforcing, **but not at the price of a developer sitting against a red build on a conflict already
resolved** — every hour it stays red makes the quiet workaround more attractive, **which is the exact
dynamic this rule names.**

### D-2.1.6 — The template declares unbreakable fields; AD-25 gains a third mechanism
**OWNER DECISION.** Escalated under D-2.0.2's pre-committed rule that **confirmation is the lead's and
deviation is the owner's**. The owner chose **option 2**, the lead's recommendation.

**Verdict.** A template may declare that a bound value must never be broken. The engine never proposes
a break inside such a field. **Declarative, not inferred** — the template author states it, the way a
real bank statement already knows which value is a customer name.

**The deviation that forced this.** Story 2.1's pre-stated gate failed **P3**: **104 violations across
102 of 180 proper-noun items.** P1, P2, P4, P5 and P6 all passed with zero violations.

**Why the two existing absolutes cannot reach it.** AD-25 names cluster identity and atomic-unknown-run;
**both are keyed on dictionary coverage, neither on proper-noun identity.** Thailand's Surname Act
requires surnames to be **coined and unique per family**, and most are built from ordinary words —
`ศรีสุข` = `ศรี` + `สุข`, both common entries. So the constrained engine finds two complete legal
matches at a **genuine cluster boundary**: P1 cannot fire (the boundary is real), P2 cannot fire
(nothing is uncoverable). Measured: **100 of 120 personal names (83%) are built this way — the common
case, not the tail.**

**The deeper reason no implementation could have closed it** (lead, after stress-testing the
alternatives): **the information is not in the character stream.** `ศรีสุข` as a surname and `ศรี สุข`
as two words are **byte-identical**; a Thai reader separates them by context and world knowledge. A
bigger dictionary cannot help — **both halves are in it**. A surname list can never be complete — the
population is **unbounded by law**. All-Thai-atomic destroys the wrapping `acceptance.md` requires.

**Options rejected, with reasons recorded:** *accept and document* — ships the failure AD-25 was
written to prevent, on the customer-facing name field of a bank statement; *surname list* — makes the
failure **rarer but less predictable**, a probabilistic patch on a correctness invariant, plus payload
against an already ~9 MB budget; *never break Thai* — trivially satisfies the rule and **breaks the
product**.

**Accepted cost, disclosed:** option 2 protects **declared name fields**. A Thai name embedded in
**free-form text** — a transaction description like "transfer to Srisuk" — **remains breakable**. That
is a smaller, statable limitation and it is the part this option does not reach.

**The trade the owner made, and its precedent.** The lead flagged the resemblance to **D-1.4.9**
deliberately: option 2's cost lands **in engine code now** rather than **on a reader of a Thai bank
statement later** — the same reasoning by which the owner overrode the lead on forward compatibility.
**It held again.**

**Consequences** *(mechanism: binding)*: one **optional format field** (additive, permitted under
D-1.4.12 — it adds a key, it does not extend a closed set); a signal carried from **binding into line
breaking**; work landing in **Stories 2.4 and 2.5**; and a **D-000.6 amendment adding AD-25's third
mechanism**, executing in the implementing story. **The amendment mechanism is D-000.6's; the choice of
what to amend to was the owner's.**

**Separately — 2 of the 104 are NOT architectural** *(mechanism: binding)*. `name-101` and `name-102`
are P6g "opaque" names whose surnames show zero interior breaks alone but violate in full context: the
**atomic-run resume scan landing on a short spurious match inside a run it had already declared
uncoverable.** That is a **defect in P2's implementation**, not the architecture, and **it is fixed
regardless of which option was chosen.**

### D-2.1.7 — The lead declined a jurisdictional argument that would have made this its own call
**Orchestrator note**, recording the lead's own reasoning because the principle outlives the instance.

AD-25's Rule under-delivers on **its own Prevents line** — arguably the textbook "incomplete canonical
document" case **D-000.6** exists to amend, which would have made this the lead's call. **It declined
that route:**

> *"D-2.0.2 pre-committed that deviation is the owner's, and I committed to it **before seeing the
> data**. That pre-commitment is the entire mechanism preventing a spike from always confirming, and
> **a pre-commitment its author can reclaim once it becomes inconvenient isn't one.**"*

### D-2.1.8 — Two things about Story 2.1 worth recording beyond the decision
**Orchestrator note.**

**The developer refused an available local fix.** It had a **pre-stated absolute failing**, an obvious
patch to hand — add a third mechanism and make P3 pass — and declined it: *"quietly patching the
ENGINE's contract to force a pre-stated absolute to pass is the same move in different clothing"* as
moving the goalposts. It also did not adjust the corpus, thresholds or labels after seeing results.
**Three stories ago the recurring failure was developers silently resolving ruling conflicts in the
diff. This is the exact inverse, chosen freely under pressure.**

**The corpus floor is why this is known at all.** The lead's P6f strengthening — *among the ≥120
personal names, a minimum must be names for which the **unconstrained** dictionary proposes an interior
break* — is what forced the decomposable case into the corpus. **Without it, a corpus of opaque
surnames would have passed all three absolutes and confirmed the approach** — and the defect would have
surfaced in Epic 4, with the layout engine built on top of it.

### D-2.1.9 — The spike is confirmed complete; P2's zero is self-referential, and the two fixes land in different places
**Orchestrator decision**, on the lead's ruling. **Confirmation half of D-2.0.2, exercised.**

**Story 2.1 is complete. No rework** *(mechanism: binding)*. It implemented AD-25 **exactly as
written** and its mechanical deliverables measured clean. **The deviation is a finding about the
invariant, not a defect in the story** — and finding it cheaply, before seven downstream stories were
built on it, is precisely what the spike was commissioned to do. **Requiring rework would punish a
spike for succeeding at its actual purpose.**

**P2 is not blind — it is SELF-REFERENTIAL, and that changes where the fix goes.** It asks *"did a
break land inside a run **the engine declared** uncoverable?"* rather than *"…inside a run that **is**
uncoverable."* In context the matcher segments `ประภา ดอเลาะ` differently than the surname alone, so
part matches, the span is **never declared uncoverable**, and P2 correctly reports zero **about its own
classification**. **Both sides move together** — the same family as an assertion comparing a message to
the constant that built it.

**Two fixes, two homes** *(mechanism: binding)*:

| Fix | Where | Why |
|---|---|---|
| **P2's measurement** — evaluate against an **independently computed** uncoverable set, derived from the dictionary directly rather than from the matcher's segmentation | **2.1's finish** | a guard reporting PASS while a known instance exists **poisons every future measurement**; two independent producers asserted against each other is the construction that has held every time |
| **The engine behaviour** — atomicity dissolving under context | **Story 2.4** | it is line-breaking behaviour, and 2.4 owns line breaking; `name-101`/`name-102` are its **retained fixture** |

**Expect P2 to go RED when the measurement is fixed, and let it.** 2.1 completes as *"deviation
recorded and routed"*; recording **P2: 2 violations** beside P3's 104 is **more honest than shipping a
PASS we know is wrong**. **No guard should ship reporting a false pass, least of all one guarding an
absolute.**

**And this matters MORE after the owner's choice, not less** *(illustrative)*: D-2.1.6 protects
**declared** fields, so a name in **free text** is still guarded only by the atomic-unknown-run rule.
**The P2 hole sits in exactly the mechanism option 2 does not replace.**

### D-2.1.10 — The unbreakable-field format key and AD-25's amendment both land in Story 2.4, together
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**Not earlier, and specifically not ahead of the enforcement.** A format field with **no consumer** is
a schema addition that **silently does nothing** — an author could set it and observe no effect. That
is the *"declared but not load-bearing"* shape of D-1.4.8's presence flags before Story 1.6 consumed
them, which D-1.4.16 had to **invent a test** to prove was real. Landing field and enforcement together
makes it **provably load-bearing from its first commit**: set → no break inside the span; unset →
breaks as before.

**Not 2.5** — that is band composition; the constraint is enforced in **line breaking**, which is 2.4.
The **AD-25 amendment** executes in the same commit per D-000.6. The field is **additive** — a new
optional key, **not** a closed-set extension — so D-1.4.12 permits it as MINOR and D-1.4.9's
passthrough already guarantees an older library loads a file carrying it. 2.4 therefore touches
`template` (the field), `bind` (marking the span) and `text` (honouring it).

### D-2.3.0 — Story 2.3 must preserve a character-range → glyph-range mapping through shaping
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*. **Flagged now rather than
discovered at 2.4.**

The atomic span is a **character-range** property, and Story 2.4 must apply it to **shaped output**.
**Thai shaping reorders and combines**, so if Story 2.3 discards provenance, **2.4 cannot tell which
glyphs belong to the protected span and the owner's chosen mechanism becomes unimplementable.** Same
shape as D-1.6.1's type-placement pre-commitment: **cheaper to require now than to retrofit.**

### D-000.16 — A stage-rank import guard, replacing AD-5's single arrow with a table
**Orchestrator decision**, on the lead's ruling. **Program-wide** *(mechanism: binding for the rule;
illustrative for the ranks)*.

**The signal rides on the value, never through an import.** `internal/text` must **not** import
`internal/bind`; the BoundTree's text nodes carry the atomic flag and the breaking API accepts it as a
**parameter**. That is the spine's whole design: *"a strictly forward staged pipeline over immutable
values."* **Stages communicate by what they pass, not by what they import.**

**Verified by the orchestrator — the graph is clean and forward-only today:**

```
geom -> (nothing)     template -> geom     bind -> geom template
fontset -> geom       pdf -> geom          text -> (nothing)
```

**Six of the spine's ten `internal/` packages exist. Four do not — `diag`, `expr`, `layout`,
`pagemodel`.** Declaring the order **once, now** guards all four arrivals; declaring it at Story 2.5
guards fewer.

**Verdict: a stage-rank guard, landing at the next story that creates an `internal/` package.** Each
package gets a rank; a package may import only **lower** ranks. Retained violating fixture, two-caller
shape per D-1.3.3, **green today**. Illustrative ranking for the implementing story to validate:
`geom` 0 · `diag`/`pagemodel` 1 · `template` 2 · `expr` 3 · `bind` 4 · `fontset` 5 · `text` 6 ·
`layout` 7 · `pdf` 8.

**Why one table beats the single AD-5 rule:** both arrows we already care about fall out **by
construction** — `layout`(7) cannot import `pdf`(8), satisfying **AD-5**; `expr`(3) cannot import
`bind`(4), satisfying **D-1.6.1's pre-commitment that stopped the decimal type being duplicated**.
**And it pre-forbids the arrows nobody has thought of yet — the entire category that made AD-5 worth
naming.** One table subsumes two rules and closes a class.

### D-2.1.11 — The confirmation is WITHDRAWN; Story 2.1 reopens at development
**Orchestrator decision**, on the lead's own withdrawal.

> *"I confirmed against a table reading P5 PASS and P6 PASS. P5 fails (102 genuine < 120), P6d fails
> (18 < 20), and P2 reported 0 against an actual 27. D-2.0.2 makes the floors **part of the gate**,
> and I pre-committed that the gate's conditions were stated in advance precisely so they could not be
> met by interpretation afterwards. **A gate that did not pass its own pre-stated conditions is not
> confirmed, and I do not get to hold a confirmation issued against numbers that turned out not to be
> measurements.**"*

**Reopened at development, not finish** *(mechanism: binding)* — the corpus needs re-sourcing and P2's
measurement needs rebuilding; neither is finishing work.

**The owner's decision (D-2.1.6) STANDS, and the correction strengthens it.** The architectural finding
is independent of every defect: 104 violations reproduce, are carried **entirely by the 102 genuine
surnames**, and `break.go` provably contains **no proper-noun concept**. The defects affect the
**floors**, not the **mechanism** — and correcting them moves the count the **safe** way: replacing
synthetics with genuine names and labelling given names yields **161, not fewer.** **The finding was
understated.** Reported to the owner as a **report, not a question**.

### D-000.17 — A floor that is not met is reported as unmet. It is never filled.
**Orchestrator decision**, on the lead's ruling. **Program-wide standing rule** *(mechanism: binding)*.

**The shape, which is the whole lesson.** The Story 2.1 developer **refused to patch the engine** to
make a pre-stated absolute pass — correctly, calling it *"the same move in different clothing"* as
changing the pass condition after seeing the data. **Then it manufactured 38 corpus items to make a
pre-stated floor pass.** That is **the identical move, applied to the sample instead of the engine**,
and it went unnoticed because **the floor counted items rather than requiring anything of them.**

> **A floor that is not met is reported as unmet. It is never filled.**

**Filling a floor is the sampling twin of moving a threshold.** If a genuine corpus cannot reach its
floor, that is a **finding** — about sourcing, about the language, about the budget — and it is
**escalated, not synthesised away.**

**Concretely** *(binding for the properties; illustrative for the mechanics)*:
- **Every corpus item carries provenance. Floors count sourced items only.**
- **Synthetic items may remain, labelled, counting toward nothing.** Do **not** delete the 38 — an
  obsolete-character string genuinely **is** an uncoverable run, so they have real exercise value for
  P6a. They must simply never satisfy a **genuine-name** floor.
- **Bar obsolete characters from the genuine buckets** — ฅ (U+0E05), ฃ (U+0E03). Mechanical, and it
  forecloses this specific dodge permanently.
- **The generator may assemble from sourced data; it may not invent items to reach a number.** That is
  the property to assert.
- **Fix P6d's predicate**: `isThaiScript` excluding U+0020 made **the space in every two-token name**
  count as "mixing with Latin/digits", inflating 18 to 167. The floor must count an actual Latin
  letter or digit.

**Why this matters beyond the gate:** this corpus becomes **S4, the frozen expected-break fixture**,
which Stories 2.4 and 2.8 and the golden report are judged against **for the life of the project**. A
fixture that is **27% synthetic strings of an obsolete character** is a poor foundation for that
independently of whether the finding is in doubt. **Rebuild it because it becomes S4.**

### D-2.1.12 — P2 is a method, not a number; and how a correct count became a wrong anticipation
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**The method, and no target number:** P2 is evaluated against an **independently computed** uncoverable
set — DP over the raw wordlist, a token uncoverable iff **no** segmentation exists at any split — and
asserts **zero** interior breaks within any such span. **The assertion is zero. 27 is the current
measured failure, not a target.** A fix landing at 3 is still a fail. **Calibrating to any number
re-creates the self-referential failure at a new value.**

**The lead's own error, recorded because the mechanism generalises.** The spike report noted that *2 of
the P6g opaque names* also violated. The lead adopted that **2** as the count of **P2-class**
violations — but those 2 were a subset of the **P3** violations that happened to sit on opaque-name
items, **a completely different population** from "breaks inside independently-uncoverable spans",
which includes given names and non-name tokens.

> **A count is meaningless without the population it ranges over, and carrying one across a population
> boundary is how a plausible number becomes a wrong anticipation that a finisher would then calibrate
> to.**

### D-2.1.13 — AC9 did not implement D-2.1.2; the ruling stands unchanged
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

D-2.1.2 ruled a **fail-closed** walk on a single criterion: *does the mechanism fail on a file it does
not recognise, or ignore it?* What shipped uses `os.ReadDir` with `if e.IsDir() { continue }` and
**never requires the expected files** — **fail-open on subdirectories**, and green with **both licence
files deleted**. **That is not a variant of the ruling; it is the shape the ruling exists to forbid.**

**Two red-proofs, both required:** delete `LICENSE-CC0-1.0.txt` (and `NOTICE`) → **red**; place an
unaccounted file in a **subdirectory** of a declared asset location → **red**. **Either one passing
means it is still fail-open.** The CC0 wordlist must appear in `MANIFEST.md` — its absence is the same
defect from the other side.

**AC1's "62,107" is an artifact, not a divergence** — a **line** count against **62,106** distinct
words (exactly the epic's figure), the +1 a duplicate (`โรม่า`). Correct the record and drop the
reported divergence: **a fabricated discrepancy in a spike report is as corrosive as a fabricated
agreement.**

### D-000.18 — Confirm against the artifact, never against a table summarising it
**Orchestrator decision**, on the lead's adopted rule *(mechanism: illustrative)*.

> *"When I hold the confirmation half of a pre-stated gate, I confirm against the artifact, not against
> a table summarising it — at minimum spot-checking the rows that would be easiest to satisfy
> vacuously."*

**Every number in the gate table was narrated by the thing being judged.** This run has now produced
narrated-count drift in a reviewer's summary, a developer's disposition table, the lead's own ruling
prose, and now a spike's gate table — and here the drift was **not a slip but a floor being satisfied
by construction in the worst sense.** The rows easiest to satisfy vacuously are **floors whose only
content is a count.**

### D-2.1.14 — The gate does not confirm; P6g stands unmet, and the shortfall is NOT free
**Orchestrator decision**, on the lead's ruling. **The orchestrator's justification for this outcome
was wrong and the lead corrected it with a measurement.**

**Verdict: option (b).** The gate does **not** confirm. **P6g stands unmet as a third failing row**
beside P2 and P3. Story 2.1 completes as *"deviation recorded and routed"* with three failing rows.

**Why not amend the floor, even though the case for amending it is good.** D-2.0.2's meta-rule is the
only thing standing between a spike and always confirming. The lead honoured it when a **genuinely
defensible jurisdictional argument** would have let it rule the deviation itself (D-2.1.7), and the
principle it logged then was its own: *a pre-commitment its author can reclaim once it becomes
inconvenient isn't one.* **Relaxing it now, when relaxing it is convenient, retires that principle at
the first opportunity to use it.** And note the structure: (a)'s case is a **good** argument —
**post-hoc amendments always have good arguments, which is precisely why the rule forbids them rather
than weighing them. A rule that yields to a sufficiently good justification is a rule that yields.**

**The categorical question, answered because it will recur** *(mechanism: binding)*: **yes, "the
category does not exist at this frequency" is genuinely different from "we did not reach the number"
— and no, that difference does not license amending the condition after it fails.** The right
handling of a mis-specified floor discovered post hoc: **record it unmet, record the measurement that
explains why, and let the next pre-commitment be better informed. You do not amend the failed
condition; you improve the next one.**

**THE CORRECTION — the synthetics do not serve P6g's purpose, and the orchestrator's reasoning for (b)
was wrong.** The orchestrator argued the shortfall cost nothing because *"the exercise value P6g was
after is delivered by the synthetics."* **Measured against the real 1,520,247-byte wordlist, and
independently re-verified by the orchestrator:**

| Character | Dictionary words containing it (of 62,107) |
|---|---|
| **ฅ (U+0E05)** | **2** — `ฅน`, `ฅนไท` |
| **ฃ (U+0E03)** | **2** — `ฃวด`, `ฃึ้น` |

| Real opaque name | Per-character dictionary frequency |
|---|---|
| `ดอเลาะ` | **8,748 – 26,383** |
| `แนแซ` | **2,095 – 21,328** |

**So a synthetic ฅ-string is near-trivially uncoverable** — essentially no dictionary word can
partially match inside it, the matcher finds no spurious boundary, and the atomic-run rule succeeds
**the easy way**. **The real opaque names that produced violations are built from ordinary modern
characters appearing in thousands of words — which is exactly WHY they violated:** the resume scan
lands on a short spurious legal match inside a run it had already declared uncoverable. **That is the
hard path, and it is the path P2 demonstrably fails on.**

**Consequences** *(mechanism: binding)*:
- **The 8 sourced opaque names are load-bearing for S4 and must be marked as such.** They are currently
  the **only** corpus items covering the path where P2 breaks.
- **P6g's shortfall is a standing item — a known thinness in the S4 fixture**, not a closed row.
  Registered in `deferred-work.md`: if more genuinely-opaque names can be sourced during Epic 2's later
  stories or Epic 4's golden-report work, **they are added.** This fixture is consulted for the life of
  the project; **8 items on its most fragile path is thin and should be visibly thin.**
- **Do not delete the 38 synthetics** — they genuinely exercise the **easy** path, which also needs
  covering. **They simply do not substitute for the hard one, and the labelling must say which is
  which.**

### D-000.19 — An unexplained delta between two independent computations is where a calibration hides
**Orchestrator decision**, on the lead's ruling. **Program-wide** *(mechanism: binding)*.

P2 returned **26** against the first reviewer's independently computed **27**. **An unexplained delta
in an independently-computed cross-check is indistinguishable from someone having nudged one
computation toward the other.**

**Only two resolutions are acceptable:**
1. Traced to a known change (here, the `place-040` label fix) **with the specific item named** and
   shown to move from one side to the other; or
2. Traced to a **defect in one of the two computations**, with the defect identified.

**"Close enough", "within noise", and "not material" are not resolutions.** If the delta cannot be
explained, **the two computations are not independent enough to be trusted as a cross-check, and the
check is rebuilt rather than accepted.** **Two producers agreeing is only evidence when a disagreement
would have been diagnosable.**

### D-000.20 — Every text tool carries an encoding and collation assumption; for non-ASCII it is usually wrong and always silent
**Orchestrator decision**, on the lead's ruling. **Program-wide** *(mechanism: binding)*. **Fifth
instrument failure of this run, and the first that is not the wrapper.**

Shell `sort`/`uniq` **mangled multi-byte Thai UTF-8** on this machine, producing silently wrong
grouping — **not an error, just wrong answers**. Unlike the previous four (a wrapper rewriting or
summarising commands) this is a **locale assumption inside a standard POSIX tool**: `sort` and `uniq`
collate per `LC_COLLATE`/`LC_ALL`.

**Rule:** for any measurement over non-ASCII data, use a tool where the encoding is **explicit**
(Python with `encoding='utf-8'`), or pin `LC_ALL=C` and accept byte-order semantics **deliberately**.
**The developer switching to Python and saying so is the correct handling and is credited.**

**Five distinct instruments have now produced false readings in one run** — a wrapper printing success
over a non-zero exit, corrupting binary through pipes, truncating output, capturing summaries instead
of output, and now a POSIX text tool mis-collating non-ASCII. **The common factor is not any one tool:
a measurement is a claim about an instrument as much as about the subject**, and this program's
standard of evidence has caught all five precisely because it keeps asking **which instrument produced
the number.** Worth a line in the Epic 2 boundary report.

### D-2.1.15 — Story 2.1's finisher closes the second QA review's Blocker and five Majors; the gate outcome (D-2.1.14) is unchanged
**Orchestrator note**, recording the finisher's pass. **Mechanism: illustrative** — this entry records
what was done and why; it creates no new binding rule.

**The second QA review verified all five prior blockers and all 19 review-#1 findings fixed, then found
the story's own Dev Agent Record repeating the exact defect the reopening exists to fix** (narrating
pre-reopening numbers as current fact — Blocker 1) **plus five Majors and five Minors.** The finisher
triaged every finding and closed the following, each independently re-measured after the change:

- **Blocker 1 — FIXED.** The Dev Agent Record's pre-reopening sections (dispositions table, "THE
  FINDING", "Genuine measured facts", Change Log) are marked `SUPERSEDED (D-000.17/18 reopening)` and
  followed by a corrected version stating the true current figures (P2 FAIL 26/17, P3 FAIL 172/120 of
  162, P6g FAIL 7 < 20, corpus 243 items/204 sourced/39 synthetic, 314 `--- PASS`/2 `--- FAIL`/0 `---
  SKIP`), never presented as passing.
- **Major 1 — FIXED.** `lint/internal/licence/classify.go`'s CC0 fallback marker no longer matches
  `"CREATIVE COMMONS CORPORATION IS NOT A LAW FIRM"` (boilerplate shared by every CC family member) —
  narrowed to `"CC0 1.0 UNIVERSAL"` alone, sufficient for the committed licence file. Three new red-proof
  cases in `classify_test.go` assert CC BY-NC-SA/BY-SA/BY-ND all classify `FamilyUnknown`.
- **Major 2 — FIXED.** The Dev Agent Record's ruling-dispositions table is extended from 24 to all 36
  rulings this story enumerates, including `D-000.17`, `D-000.18`, `D-2.1.13`, `D-2.1.14`.
- **Major 3 — FIXED.** The 27→26 delta (D-000.19) is now traced in `SPIKE-REPORT.md` itself: the 27 was
  measured over the superseded 220-item corpus; the current 243-item corpus (post D-000.17 rebuild)
  independently re-derives 26. Different populations, not drift.
- **Major 4 — FIXED.** `SPIKE-REPORT.md`'s P2 representative-violations table carried four rows from the
  pre-rebuild corpus (three naming the wrong text for their id). Regenerated from the live test output.
  Story 2.4's retained-fixture reference (D-2.1.9) is re-pointed from `name-101`/`name-102` to
  `name-116`/`name-117` — see the next item for why there are two ids, not three.
- **Major 5 — FIXED.** D-000.17's "may not invent items to reach a number" property was enforced only
  against the obsolete-consonant dodge, covered only two of five sourced buckets, was reachable from no
  gate, and was already evaded by one item (`ฉั่วสมบูรณ์`, self-described only as "a plausible" surname).
  The finisher (a) extended the obsolete-consonant bar to all five sourced buckets; (b) added
  `TestCorpusRegeneratedMatchesCommitted` (`folio-go/cmd/gencorpus/main_test.go`), mirroring
  `TestTrieRegeneratedMatchesCommitted`'s precedent, so a hand-edited `corpus.json` now reddens a gate;
  (c) relabelled `ฉั่วสมบูรณ์` from `personal_name`/`sourced` to `synthetic_probe` rather than
  retroactively asserting an attestation that was never established. **Consequence, measured and
  disclosed everywhere it appears:** P6g's generator-measured count moves from 8 to 7; P5's
  personal-name count moves from 123 to 122 (still ≥120); P3 moves from 173 to 172 violations (one
  fewer proper-noun-bearing item); P2's 26/17 is unaffected (the same text still violates, now under id
  `synthetic-039`, and P2 is computed over every Thai-script span regardless of category). `DW-11` in
  `deferred-work.md` is corrected to its honest load-bearing figure: **2** genuinely-uncoverable,
  independently-attested sourced opaque names (`ดอเลาะ`, `แนแซ`) — not 7, and not the original 8.
- **Minors 1, 2, 3, 5 and Nits 1–4 — FIXED**, each a disclosure or comment correction, no further
  code-behaviour change: `DW-11`'s figure (Minor 1, above); the P6a genuine/synthetic split, 64 = 25
  sourced + 39 synthetic (Minor 2); P6d's zero-margin/single-category disclosure (Minor 3); the
  corpus-wide "does the switch matter" figure, 119/122 same on the surname alone, 17/122 differ on the
  full item text (Minor 5); stale corpus ids in `break_test.go`'s V11 comment (Nit 1, re-pointed to
  `name-116`/`name-117`); the report's own miscounted preamble (Nit 2); `DW-10`'s placeholder ruling id
  (Nit 3, now `D-2.1.5`); the `js/wasm` leg's success log understating what it asserted (Nit 4).
- **Minor 4 and Nit 5 — left open, by design.** Minor 4 (whether AC5's hand review being mechanical
  labelling changes what evidentiary weight P3's 172 carries) is routed as an open `DECISION NEEDED` in
  `SPIKE-REPORT.md` rather than resolved unilaterally by the finisher — a genuine interpretive question,
  not a measurement defect. Nit 5 (`sprint-status.yaml`'s `epic-2: backlog`) is left untouched on
  explicit instruction: the tracker advances epic-level keys only at the epic's own gate, and this
  finisher's mandate was scoped to `2-1-thai-break-opportunity-spike`'s own status line only.

**The gate outcome does not change.** D-2.1.14 stands: the gate does **not** confirm. P2, P3 and P6g
remain three failing rows, `TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck` remain
the two intentional `--- FAIL`s, and D-2.1.6 (the owner's decision, resting on P3) is unaffected — P3's
mechanism and architectural cause are identical to what the owner already ruled on; only bookkeeping
figures around it were corrected. **This entry closes the finisher's pass; it is not a new confirmation
and it does not reopen the story.**

### D-2.2.2 (superseded) — The subset tag hashes the SUBSET PROGRAM BYTES; the F2DOT14 clause is withdrawn
**Orchestrator decision**, on the lead's ruling. **Answers the Story 2.2 creator's DN-1.** A conflict
between two of the lead's own rulings, not an implementation shortfall.

**Why the ruled mechanism was unreachable — verified by the creator and independently by the
orchestrator:** `normalizedCoords []int` is an **unexported field** at `subset/plan.go:46`; **`Plan`
has 17 exported methods and none returns it** (enumerated); there is no `Avar()`. **The only reachable
axis values are `Input.PinnedAxes() map[ot.Tag]float32` — the caller-facing floats the same ruling
forbids hashing.**

**And recomputing is worse than what it rejects.** It needs `Fvar().NormalizeAxisValue()` (returns
`float32`) plus replicating the unexported `floatToF2DOT14`, whose body is `int(v*16384 + 0.5)`
(`plan.go:525`) — **a fusable multiply-add, the exact shape D-2.2.0 measured as the live FMA surface in
this very dependency** — placed in the **tag** path, where a divergence is *harder to attribute*: **a
wrong tag is a silently colliding subset, not a hash that visibly moves.**

**Verdict** *(mechanism: binding)*: the six-letter tag derives from a hash of **`subset.Subset()`'s
returned program bytes.**

**Why this is right rather than least-bad:**
- **It is exactly what D-2.2.2 said the principle was** — *"hash what determined the bytes, in the form
  it was determined in."* **The bytes are what determined the bytes.** The lead named the principle and
  then reached past it for a proxy.
- **Zero float anywhere** — not ours, not recomputed, not borrowed.
- **Trivially reachable** — it is the return value already held.
- **Collision becomes impossible by construction:** two subsets share a tag only if their programs are
  byte-identical, in which case they **are** the same subset and sharing is correct. **Strictly stronger
  than either the glyph-set or glyph-set-plus-coordinates form.**
- **Consistent with D-1.5.8**: a tag keyed on the *request* lies about what is embedded. Hashing the
  request's axis values — even as integers — is that same error one layer along. **Program bytes are the
  maximal form of "what is embedded."**

**Non-circularity, verified by the lead and re-verified by the orchestrator:** nothing writes the tag
into the font program. The tag's **sole** appearance is `/BaseFont /` at `internal/pdf/textdoc.go:240`
and `:250` — a **PDF-level name**. **The story must ASSERT this rather than inherit it:** if anything
later patches the font's `name` table with the tag, the derivation becomes circular and must be
revisited.

**Both discrimination fixtures survive by a stronger mechanism** — different glyph sets of the same
size → different programs → different tags (D-1.5.8); two instances with the same glyph set →
different programs → different tags (D-2.2.2's own fixture). **Keep both.**

**Accepted cost:** a `textshape` upgrade now moves every tag. **Correct behaviour, not a defect** —
AD-22 already makes any subsetting change a versioned breaking change, the same argument that settled
D-1.5.8.

**D-000.6 amendment to AD-7's tag clause, in 2.2's commit — the lead's, not the owner's.** AD-7 names
*"a hash of the sorted glyph-id set"*, which **under-specifies for instanced faces** — the same shape as
AD-25's Rule under-delivering on its own Prevents. **Unlike AD-25, this has no product consequence, no
scope change, no payload change and no story re-plan:** a mechanism substitution that strictly improves
discrimination while preserving the clause's intent. **Binds and Prevents untouched.** Land it **before
any golden is recorded** — the tag is in the bytes.

### D-2.2.3 — `ScanMapRange` extends to the whole module; and the lead's fourth directory-scope error
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**D-2.2-D3's guardrail was vacuous.** It ruled *"`internal/` must never range the `FontSet` map"* for a
type that lives in package `folio` **at the module root**. `ScanMapRange`'s production root is
`folio-go/internal` only, and **no `internal/` package can name `folio.FontSet` without an import
cycle — so the guardrail is green whether or not the guard exists.**

**The lead's own account, recorded because the fix is the transferable part.** This is the **fourth**
instance of the same error. After the third it wrote down an intention to verify *where a thing lives*
before invoking a directory-scoped rule, and did not follow it. **Replacing intention with a mechanical
habit:**

> **When I write a guardrail naming a directory scope, I name the file the risk actually lives in. If I
> cannot name the file, I have not checked.**

**The real guard:** extend **`ScanMapRange`'s production scan to the whole module** (`folio-go/`,
`testdata` and dot-directories skipped) — exactly as the `float64` scan was extended at D-1.6.7, and for
the identical reason: **the render path's extent starts in package `folio` (D-1.6.3), so the hazard does
too.** One new production **caller** on an existing checker, not a new checker.

**Expect either outcome:** it may go **red immediately** if `render.go` ranges the `FontSet` — **a real
defect found, not a problem with the guard**. If green, its red-proof is adding a range to `render.go`.
The `slices.Sorted(maps.Keys(m))` escape hatch applies unchanged.

### D-2.2.4 — AD-8's diagnostics citation is wrong; fix it as a D-000.6 amendment
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*. Found by the Story 2.2 creator.

AD-8's Rule says *"is a diagnostic **(AD-15)**"*. **AD-15 is designer-owns-the-document (`:304`);
diagnostics is AD-14 (`:289`).** A **citation error in a Rule's text**, inside one document, **Binds and
Prevents untouched** — squarely a D-000.6 amendment, fixed in 2.2's commit per the D-1.8.7 precedent.
**Fixing rather than parking is correct: a spine invariant citing the wrong sibling invariant is exactly
what every later reader re-derives independently.**

### D-2.2.5 — The payload delta is resolved by MEASUREMENT, not by arbitration between documents
**Orchestrator decision**, on the lead's ruling *(mechanism: binding)*.

**Reported, not filled** (D-000.17), and **5.4's ACs are not edited.** NFR7 (`epics.md:121`) itemises
~1.5 + 7.4 + ~0.1 = **9.0 MB**; Story 5.4 AC2 (`:1400`) itemises 1.5 + 0.4 + 0.1 + 7.4 + 0.12 =
**9.52 MB** — **a ~0.52 MB disagreement**: either 5.4 double-counts Latin+Thai or NFR7 omits them.

**Story 2.2 is the first story that can weigh the real faces, so it measures rather than arbitrates** —
shipped face sizes **and the actual compressed trie** (2.4 MB raw against a claimed 0.1 MB is a **24×**
ratio, plausible for a trie but currently unmeasured) — and reports **all** the numbers. **Whoever
reconciles then does it with data rather than by choosing which document to believe.**

**The consequence that makes this more than bookkeeping** *(illustrative)*: **Story 5.4's AC2 itemises
those numbers TO USERS** and requires the screen to explain *why CJK dominates*. **A wrong itemisation
shown to a user is a product defect, not a documentation typo** — so a material discrepancy is
**owner-visible at the Epic 2 report**, not a quiet correction.

### D-000.18 (mechanized) — The gate harness writes machine-readable results; the story's tables are asserted equal to it
**Orchestrator decision**, on the lead's ruling. **Program-wide** *(mechanism: binding)*. **Lands in the
Epic 2 gate** — pipeline machinery, not font work.

**Six instances of narration drift in this run**, culminating in a Dev Agent Record claiming **PASS on
failing rows inside the story reopened to fix exactly that.** That is conclusive: **the rule binds and
is not changing behaviour.**

> **The gate harness writes a machine-readable results file. The story's results table and Dev Agent
> Record are asserted equal to it.**

**This converts "please do not drift" into "cannot drift"**, using the two-independent-producers
construction that has worked every time (D-1.4.10's independent pretty-printer, D-1.4.14's shared
spelling table, D-1.4.15's set equality). **The work has been right every time; only the transcription
has failed — so guard the transcription, not the work.**

---

### D-000.21 — Assert on the produced thing, never on the thing you asked for

*(mechanism: binding)*

**The rule.** Every assertion about an artifact must read the property **off the artifact**. An
assertion that reads the *input* — the value requested, the pin passed in, the set asked for, the
summary written afterwards — proves only that the request was well-formed. It cannot detect a
producer that ignored the request, transformed it, or silently carried something else through.

**Why this earned a standing entry.** It is the fourth time this program has hit the same defect, and
each instance looked like a different problem at the time. None of them looked like the previous one:

| # | Where | The input-side assertion that would have passed | What was actually wrong |
|---|---|---|---|
| D-1.5.8 | subset tag | derive the tag from the **requested** glyph set | the tag would name glyphs that were never embedded |
| P2 (Thai spike) | coverage floor | ask what the engine **declared** uncoverable | it never asked what *was* uncoverable |
| D-000.18 | the record | confirm against the **table summarising** the run | the table drifted from the run it summarised |
| D-2.2.x (this) | font weight | assert the **pin passed** to the render call | `usWeightClass` on the embedded program read 100 |

The lead made the D-1.5.8 ruling itself and then failed to apply that ruling's own logic to weight one
story later — which is the strongest evidence available that this needs to be a named rule rather than
a habit. A habit did not survive one story.

**The plain-language version.** *Asking for Regular and getting Regular are two different facts.* Only
the second one is worth asserting, and only the produced file can answer it.

**How this bit, concretely.** Story 2.2 embedded Simplified Chinese as **Thin (usWeightClass=100)** in
a recorded golden. Three guards agreed the artifact was correct — four-target byte-identity, the
font-program digest, and the missing-glyph diagnostic — and **all three were right.** The bytes really
were identical on every target; the program really was stable; the glyphs really were covered. Each
guard answered its own question accurately. **None of them was asked whether the value meant what its
name implied.** That sentence leads the Epic 2 gate's section on this class.

**The reason the shipped face was the one that broke.** `NotoSansSC-VF`'s `wght` axis has
**default=100**, not 400 — alone among the three faces. Latin and Thai default to 400, so every test a
developer ran looked correct, and the one face that was wrong is the one whose script nobody on this
project reads. A defect that is invisible to its author and legible only to a reader we do not have is
exactly what a semantic assertion is for.

---

### D-000.22 — A golden fixture needs a semantic acceptance step at first recording

*(mechanism: binding)*

**The rule.** When a golden fixture is recorded for the **first** time, the story recording it must
name and assert at least one **semantic property** of the artifact — a property that would be wrong if
the recording itself were wrong — separately from and additionally to the hash. Per [[D-000.21]] the
property is read **off the artifact**, never off the inputs that produced it.

**Why a hash cannot do this job.** A hash guards against **change**. It says "this is the same as what
we recorded." It has nothing whatever to say about whether what we recorded was **right**, and it
cannot acquire that ability later. The asymmetry matters because of what a golden *becomes*:

> The moment a golden is committed it stops being an output and starts being an **input**. Every
> guard downstream then agrees with it — correctly, and forever. A wrong first recording is not a bug
> that gets caught later; it is a bug that gets **ratified** later.

So the first recording is the **only** moment at which the question "is this right?" is answerable at
all. After that, every mechanism in the project is structurally committed to answering "is this the
same?" instead.

**The plain-language version.** *A hash tells you nobody moved the furniture. It does not tell you the
furniture was ever in the right room.* Somebody has to look, once, at the beginning.

**What counts as a semantic property.** Something a human could be wrong about and a machine can check
— the weight class of an embedded font, the page count, the number of glyphs actually embedded, the
declared page size, the compression filter used. Not a restatement of the hash, and not a property
derived from the same inputs the artifact was built from.

**What this story must do under it.** Story 2.2 asserts, on the **embedded program** in each
font-bearing golden: `OS/2.usWeightClass` equals the intended weight, `fvar` is absent, and `gvar` is
absent. And — since the failure mode here was specifically *"a value nobody could read"* — a human
looks at the rendered Chinese before the golden is frozen. The machine check catches the recurrence;
the human check is what would have caught the original.

---

### D-2.2.4 — Ship static instanced faces, Regular-only, and reject caller-supplied variable faces

*(mechanism: binding)* — **reverses D-2.2.3**, which put the static switch in its own story.

**What the reversal turned on.** The separate-story ruling's strongest reason was attributability:
two stacked causes moving one golden at once. That reason was **moot and the lead said so plainly**
rather than defending it — Story 2.2 has not committed, so there is no prior golden to move *from*,
and 2.2 must re-record regardless because the Thin defect is in its own seam and its own fixture. The
real choice was never "one movement or two"; it was **record once against the correct artifact, or
record twice.**

**The fact that forced it.** `textshape@v0.0.15`, `subset/execute.go:496-499`, copies `OS/2`
**verbatim** into the subset and never updates `usWeightClass` when instancing. Independently
confirmed: the identifier appears nowhere in textshape except `ot/metrics.go`, which only ever
*parses* it. There is no writer. So pinning `wght=400` on the variable CJK face would have produced
**Regular outlines carrying metadata that still claimed Thin** — strictly worse than the shipped
defect, where outlines and metadata at least agree. Correcting it in-place would mean rewriting `OS/2`
and recomputing both its table-directory checksum and `head.checkSumAdjustment`, in the render path,
ordered before the tag hash.

Shipping statics deletes that requirement rather than satisfying it. `fontTools` **explicitly writes**
`OS/2.usWeightClass` when it instances (it logs `Setting OS/2.usWeightClass = 400`, and the produced
file reads back 400), so textshape's verbatim copy then propagates a correct value. Worth stating
precisely: it works because fontTools *sets* the field, **not** because nothing touches it. The
latter framing would predict breakage the moment the subsetter changed, and that is not the mechanism.

**Three corrections applied to the ruling before it was briefed** — each found by reading a produced
file under [[D-000.21]], and each one would have shipped a defect:

1. **"No `fvar`, so the FMA hazard is gone" was false as first specified.** `NotoSans-VF` and
   `NotoSansThai-VF` each carry **two** axes — `wght` *and* `wdth` (min 62.5, default 100, max 100).
   Pinning `wght` alone leaves `fvar` **and** `gvar` alive on **four of the five** faces; only the
   single-axis CJK face becomes genuinely static. The story would have recorded a golden that *looked*
   static while the float `gvar` path still ran. Pinning both axes also cuts Latin's compressed size
   40% (377 KB → 227 KB), because `gvar` was riding along unnoticed.
2. **Regular + Bold contradicted the story's own fence.** `fontset.go:250-258` fences Bold out in
   prose — D-2.2.1: *"Bold is a wght instance; if it arrives after 2.2, its story inherits this
   obligation"* — and the package deliberately exposes no way to request a non-default instance. Bold
   faces would have been **261 KB selectable by nothing**. The lead had contradicted its own D-2.2.1
   one message after making it. **Regular-only for all three faces.**
3. **"Pin every axis to an explicitly chosen value" was unimplementable.** Reaching the vendor's
   `PinAxisLocation` requires the identifier `float32`, and `arch_test.go:54` bans `float32` *and*
   `float64` under `internal/` **and** the module root (AD-23). The instruction and the arch guard
   could not both hold; the guard is older and load-bearing.

**The resolution to (3) deletes the seam instead of fixing it** *(mechanism: binding)*: a
caller-supplied face that still has `fvar` is **rejected with a located diagnostic at face ingestion**
— not instanced, and not mid-render, so the failure is early and named. Shipped faces arrive static,
so no pin is ever needed. This satisfies AD-8's no-silent-substitution, keeps AD-23 intact, and
**removes the float `gvar` path from the render entirely** — the FMA hazard stops being monitored and
stops existing. The message must **name the remedy** ("instance it first", with a concrete pointer),
since most Google Fonts downloads are variable today and a caller needs an action, not a refusal.

**Payload, measured with `epics.md:1382`'s mandated codec (brotli -q 11), all faces, engine included:**

| configuration | fonts | + engine 1.5 + dict 0.12 | vs ~9.52 MB |
|---|---|---|---|
| variable (as shipped in 2.2 today) | 9.33 MB | 10.95 MB | **+15% over** |
| static, Regular + Bold | 5.33 MB | 6.95 MB | inside |
| **static, Regular-only (adopted)** | **5.07 MB** | **6.69 MB** | **inside, 2.8 MB headroom** |

An earlier reconciliation of mine claimed this overrun had dissolved; the reviewer proved it had not,
on three counts — MiB relabelled as MB, gzip used where brotli is mandated, and the ~1.5 MB engine
omitted from the measured side while both budgets include it. The overrun was **real**. The static
switch is what resolves it, not the arithmetic.

---

### D-2.2.5 — NFR7's "variable build" clause is amended under D-000.6; the PRD and addendum are reported, not edited

*(mechanism: binding)*

**The tension.** NFR7 (`epics.md:121`) says verbatim: *"Take the glyf/TrueType **variable** build over
CFF/OpenType."* Shipping statics contradicts its literal text.

**Why the amendment is in scope.** NFR7 is **unimplementable as written**: its own font budget is
0.4 + 0.1 + 7.4 = **7.9 MB**, and the build it mandates measures **9.33 MB** compressed. The clause
cannot be satisfied on its own terms. D-000.6 exists precisely for the "false, incomplete, or
unimplementable" case, and `epics.md` is inside its scope by established precedent.

**What the amendment preserves.** The clause's operative contrast is **glyf vs CFF**, not variable vs
static — and the statics are verified `glyf=True, CFF2=False`. NFR7's intent (three scripts, embedded,
byte-stable, fully offline, within ~9 MB) is preserved and is **only reachable with** the amendment.
So the **adjective** changes; the choice it was defending does not.

**A hypothesis the lead raised, tested, and discarded — in the direction that helps.** The lead
proposed that the budget had been costed against a static build. Reading the surrounding paragraphs
falsified it: `addendum.md:399` costed **the same file, to the byte** — *"`NotoSansSC[wght].ttf` is
17.77 MB on disk → 7.42 MB compressed, covering nine weights"* — against
`NotoSansCJKsc-Regular.otf` at 16.44 MB → 10.90 MB. The 7.42 figure is optimistic and sits **below the
same document's own table**, which puts TTF/glyf variable at ~44–51%, i.e. 7.82–9.06 MB. Our measured
8.30 is inside their range; their own quoted figure is not.

**This reframes the amendment far more favourably: the reasoning was sound, the option set was
incomplete.** The authors compared glyf-variable-nine-weights against CFF-static-one-weight and
**correctly chose glyf**. They never costed **glyf-static-one-weight** — 4.82 MB — which beats both.
The operative contrast is *reinforced*, not overturned: glyf static (4.82) < glyf variable (8.30) <
CFF static (10.90).

**And the benefit the variable build was bought for is unreachable.** Its stated advantage is
*"covering nine weights"*, but `folio-format.md:192` gives the format `bold` and `italic` as
**booleans** — it can express **two** weights, and 2.2's engine can select **one**. The variable build
costs **3.48 MB compressed for eight weights nothing can request.** This is the same defect as the
Bold faces in [[D-2.2.4]], one layer up: **shipping a capability nothing can ask for.**

**Scope boundary, held deliberately.** `prd.md:455` and `addendum.md:399` are **outside** D-000.6 —
its scope is the spine, `folio-format.md`, the SPEC companions, and `epics.md` by precedent. The lead
declined to widen it to reach a document the owner authored, on the grounds that quiet enlargement is
exactly what a stated boundary exists to prevent. **`epics.md` is amended in 2.2's commit; the PRD and
addendum are reported at the Epic 2 gate** — they retain the superseded costing, the glyf-over-CFF
reasoning is unaffected and reinforced, and whether to correct them is the owner's call, since no
implementation depends on it.

**Method note the lead asked to keep.** Its wrong hypothesis surfaced only because the quoted sentence
was checked against the paragraph containing it. **A quoted sentence is an input; the paragraph it
sits in is the artifact** — [[D-000.21]] applied to documents rather than fonts.

---

### D-000.23 — A guard written in response to a defect covers the defect, not its class

*(mechanism: binding)* — the sampling rule behind [[D-000.22]]'s "semantic acceptance step".

**The rule.** When you write a guard *after* a defect, assume the field you are about to guard is the
one already handled by mature tooling, and go looking for the fields **nobody has been burned by
yet.** Cover **every field that carries the meaning**, not the field that failed.

**The mechanism, which is counter-intuitive and is why this needs stating.** Mature tooling was
written by people who got burned first. So a property that just burned *you* is evidence about **the
tooling's maturity on that property** — it says nothing about your artifact's other properties.
Selecting your guard by "what went wrong last time" therefore selects for **the field most likely to
be already fixed.** That is close to the worst available sampling rule.

**The instance that earned it, which is as clean as this class ever gets.** The morning's ruling
([[D-000.22]]) required a semantic acceptance step on golden fixtures, and the property chosen was
`OS/2.usWeightClass` — the field that had just shipped Thin Chinese. That is precisely
the field `fontTools` handles: it logs `Setting OS/2.usWeightClass = 400` and writes it correctly.
Meanwhile the instanced Regular face still carried:

```
name[1] 'Noto Sans SC Thin'    name[6] 'NotoSansSC-Thin'    name[17] 'Thin'
OS/2.usWeightClass = 400       outlines: Regular
```

**The guard designed that same day to catch the defect would have passed the defect**, because weight
is stated **three independent times** in this artifact — the weight class, the name records, and the
outlines themselves — and they can disagree pairwise. `usWeightClass` was the one already fixed.

**The remedy is about cardinality, not vigilance.** The assertion set's job is to cover the **number
of independent places the meaning is represented.** Count those first; assert all of them.

**The recursive catch, worth keeping because it happened inside the fix.** The corrected assertion set
first briefed to the finisher contained `no name record contains "Thin"` — keyed on the exact string
that burned us. A face defaulting to `Light`, `ExtraLight` or `Black` sails past it. **A denylist
entry is never coverage;** the positive assertions are the guard (`name[1]` equals the exact family,
`name[2] == "Regular"`, `name[6]` ends `-Regular`). The same error, one level down, inside its own
correction.

**Consequent obligation** *(mechanism: binding)*: derive a per-artifact assertion set from a
**declarative spec** — one record per shipped face: family, instance name, axis pins, expected weight
class — and assert **spec equals artifact**. Enumerating assertions by hand means the next story that
adds a face inherits none of them, which is how a guard written for today's artifacts fails to cover
tomorrow's.

---

### D-2.2.6 — `/BaseFont` must be the embedded program's PostScript name (ISO 32000-1 Table 117)

*(mechanism: binding)* — orchestrator decision; the lead framed the fork and left the call here.

**The defect, pre-existing since Story 1.5.** ISO 32000-1 Table 117 (CIDFontType2): *"BaseFont — the
PostScript name of the CIDFont. For Type 2 CIDFonts, this shall be the value of the CIDFontName entry
in the CIDFont program."* `textdoc.go:234` builds it as `face.Tag + "+" + pdfNameEscape(name)` where
`name` is the **FontSet key**, and `:250` reuses the same variable for the descendant CIDFont. So the
**declared** name and the **embedded program's** name disagree — `NotoSansSC` against
`NotoSansSC-Thin` today, `NotoSansSC-Regular` after the static switch. Disagreeing either way; the
switch does not create it.

**How it surfaced, which is the interesting part.** The report to the owner claimed the Thin defect
"would have been visible in the PDF as `/BaseFont`". The lead measured the golden and falsified it:

```
/BaseFont:  AWPJVD+NotoSansThai   MQPBJO+NotoSans   ZVXRIO+NotoSansSC
"Thin" / "Regular" / "Bold" in the PDF bytes:  0 occurrences each
```

`name[6]` never reaches the output at all. **The artifact would have shipped with Regular outlines, a
correct weight class, an embedded program calling itself Thin, and nothing anywhere in the PDF showing
it.** The assertion set was the only thing that could ever have caught it — which strengthens
[[D-000.23]] rather than weakening it.

**Ruling: fix it inside Story 2.2.** Expose `PostScriptName()` on `internal/fontset` (it must not leak
`*ot.Font`, so D-2.2.x's encapsulation holds) and use it for `baseFont`, giving
`ZVXRIO+NotoSansSC-Regular`. Three reasons:

1. **It is free now and expensive later.** 2.2 re-records every font-bearing golden regardless, so the
   byte movement is already paid for. Deferring costs a separate re-record **plus** a full four-target
   matrix run.
2. **It is a real conformance defect** — a PDF/A validator flags it, and it is the name a viewer falls
   back to when the embedded program fails to load.
3. **It converts an invisible property into a visible one.** The reason this defect hid is that
   nothing in the output carried the font's own identity. Once `/BaseFont` reflects the embedded
   program, a future weight defect of this class becomes detectable **by reading the PDF**, rather than
   only by the dedicated assertion we happened to think of. That is a durable structural gain, and it
   is the same principle as [[D-000.21]] expressed in the artifact instead of in a test.

**Bounded**: if the change turns out to be materially more than the accessor — anything reshaping the
fontset seam or moving a non-font golden — it stops and returns as a finding for a later story rather
than expanding inside this one.

---

### D-2.2.6 (amended) — the name comes from the shipped face; the embedded programs carry no identity

*(mechanism: binding)* — amends [[D-2.2.6]]'s **mechanism**. The decision is unchanged: the
conformance fix lands in Story 2.2, bounded to one exposed method on `internal/fontset`.

**What was wrong.** D-2.2.6 as first specified said to read the PostScript name **from the embedded
program**. There is nothing there to read. Verified directly on the golden's three `FontFile2`
streams:

```
obj   6  60048 B  name=False cmap=False post=False | GDEF GPOS OS/2 glyf head hhea hmtx loca maxp prep
obj  11    732 B  name=False cmap=False post=False | GDEF      OS/2 glyf head hhea hmtx loca maxp prep
obj  16    560 B  name=False cmap=False post=False | GDEF      OS/2 glyf head hhea hmtx loca maxp prep
parsed 3/3 embedded programs — name table present in NONE
```

`textshape/subset/execute.go:531-545` places `ot.TagName` in `optionalTables`, copied only when
pass-through is requested; nothing requests it. **This is not a defect** — PDF §9.9.2 sanctions the
reduced table set for embedded CIDFontType2 programs, because `CIDToGIDMap` performs the mapping.

**Corrected mechanism: read the PostScript name from the shipped face.** This is not a compromise.
PDF's own model is that the embedded program *need not* self-identify — which is precisely why
`/BaseFont` exists and why the spec permits stripping `name`. Every real producer writes the original
font's PostScript name into `/BaseFont` and ships a name-less subset. Enabling pass-through instead
would be the wrong trade: the CJK subset is **732 bytes**, and Noto's full multi-language name table
would several-fold it **in every PDF Folio ever produces**, to satisfy a clause the spec's own
embedding model works around.

**It does not violate [[D-000.21]].** The shipped face is *our* produced artifact, with its own
asserted identity, and the subset is derived from it — so the name is true of the program by
construction. The chain is spec → generated face (asserted) → `PostScriptName` → `/BaseFont`, and
every link is checked. That is materially different from reading an input we do not control.

**D-2.2.6's third reason survives and is stronger under this mechanism**: a future weight defect of
this class surfaces as `XXXXXX+NotoSansSC-Thin` in `/BaseFont`, readable in the PDF. Under
name-pass-through it would have been buried inside a font stream.

**The circularity question dissolves.** `/BaseFont`'s two inputs are independent — the tag from the
final program bytes, the name from the shipped face — so no cycle exists and the existing
non-circularity red-proof still guards the only shape that could create one.

---

**The vacuity trap this created, which is the part worth keeping** *(mechanism: binding)*

The orchestrator's own correction to the assertion set instructed that `name[1]`, `name[2]`, `name[6]`
and the "Thin" check be run **against the embedded program**. Run there, **every one of them passes
vacuously** — there are no name records, so no record contains "Thin", and the positive checks read a
table that does not exist.

**Four green assertions from an artifact carrying none of the fields they name** — the exact failure
class this story exists to close, introduced *by the correction to it*, one message after
[[D-000.23]] was logged.

**Why the mistake was natural, which is what makes it dangerous.** Having just adopted *"assert on the
produced thing"*, the instinct is to reach for the **most**-produced artifact in the chain. Here that
is the artifact from which the field has been deliberately removed. **"Produced" and "carries the
property" are independent**, and [[D-000.21]] silently assumed they travel together.

**Consequent obligations:**

1. **Assert face identity on the face; assert `/BaseFont` on the PDF; assert `usWeightClass` on the
   embedded program** (which *does* retain `OS/2`). Two artifacts, three assertion sites, none
   substituting for another.
2. **Every assertion group must first assert that the table it reads exists**, and fail loudly if it
   does not. A presence precondition is what makes this class impossible rather than merely avoided
   this time.

---

### D-000.21 (sharpened) — Assert on the artifact that carries the property, and prove it carries it

*(mechanism: binding)* — supersedes [[D-000.21]]'s statement, one day after it was made.

**The original.** *"Assert on the produced thing, never on the thing you asked for."*

**What it missed.** The produced thing **may not be where the property lives.** D-000.21 silently
assumed that "produced" and "carries the property" travel together. They are independent, and the
counter-example arrived within a day: the embedded font programs are the most-produced artifact in the
chain and carry **no name table at all**, so every identity assertion aimed at them passes **vacuously**.

**The corrected rule, in full:**

> **Assert on the artifact that carries the property — and prove it carries it.**

The second clause is not a refinement or a defensive nicety. It is **the missing half of the original
rule**: without it, "assert on the produced thing" reduces to a vacuous pass exactly where the
property has been legitimately stripped, and reads as green. The presence-precondition from
[[D-2.2.6]] (amended) is therefore binding wherever this rule is applied — **assert the field exists
before asserting anything about its value.**

**Why the failure mode is specifically hard to see.** Having adopted "assert on the produced thing",
the natural instinct is to reach for the **most**-produced artifact in the chain. That instinct is
precisely wrong when a downstream stage has deliberately removed the field — and deliberate removal is
common, because downstream stages exist to strip what they do not need. **The more processed the
artifact, the more likely it is to have lost the property you want to check.**

**Worked example, all three sites from one story:**

| property | lives in | asserting it elsewhere |
|---|---|---|
| face identity (`name[1]`, `name[2]`, `name[6]`) | the **shipped `.ttf`** | passes vacuously on the embedded program — no `name` table |
| `/BaseFont` conformance | the **produced PDF** | not present in the face at all |
| `usWeightClass` | the **embedded program** (retains `OS/2`) | asserting on the face misses subsetter corruption |

Three properties, three different artifacts, **no two interchangeable** — and the naive reading of
D-000.21 would have put all three on the last one.


---

### D-2.2.6 — `/BaseFont` carries the embedded program's own name; and the assertion split that fixing it forced

*(mechanism: binding)* — ruled by the coordinator during Story 2.2's finisher pass.

**The defect.** ISO 32000-1 Table 117 (CIDFontType2) requires `/BaseFont` to be *"the value of the
CIDFontName entry in the CIDFont program"*, prefixed with §9.6.4's six-letter subset tag. Folio
spelled it from the **FontSet key** (`internal/pdf/textdoc.go:234`, `:250`) — the name the caller
happened to file the face under — so the declared name and the embedded program disagreed:
`NotoSansSC` against `NotoSansSC-Regular`. Pre-existing since Story 1.5; the static switch did not
create it.

**Fixed in 2.2 rather than deferred, for three reasons.** It costs one exposed accessor and **zero
extra golden movement**, because 2.2 was re-recording anyway — deferring would cost a separate
re-record plus a full four-target matrix run. It is a real conformance defect: PDF/A validators flag
it, and it is the name a viewer falls back to when the embedded program fails to load. And it
**converts an invisible property into a visible one** — the Thin defect ([[D-000.21]]) hid precisely
because nothing in the output carried the font's own identity. Once `/BaseFont` reflects the embedded
program, a future weight defect of this class is legible by reading the PDF, not only through the
dedicated assertion we happened to remember to write.

**Scope held to one accessor**, as ruled: `fontset.PostScriptName()` returns a plain string and
leaks no vendor type (AC17a, D-1.5.10). It reads `name` record 6 **directly**, never through
`(*ot.Face).PostscriptName()`, which substitutes the literal string `"Unknown"` when the name table
is absent — the same silent-substitution shape the `Upem()` comment in that file already documents.

**The measurement that changed the ruling, and the general rule it produced.** The instruction was
initially to read every name assertion off the **embedded `FontFile2` program**. Measured on all
three of the golden's streams: **the embedded programs have no `name` table at all.** `textshape`
lists `ot.TagName` in `optionalTables` (`subset/execute.go:531-545`) and copies it only on explicit
request; nothing requests it. That is not a defect — PDF §9.9.2 sanctions the reduced table set for
an embedded CIDFontType2, since `/CIDToGIDMap` does the mapping.

> Had the instruction been followed as written, `name[1]`, `name[2]`, `name[6]` and *"no record
> contains Thin"* would each have passed **vacuously**: there are no name records, so no record
> contains anything. Four green assertions from an artifact carrying none of the fields they name —
> the exact failure class the story exists to close, introduced by a correction to it.

**The rule, in its sharpened form** *(binding, generalising [[D-000.21]])*:

> **Assert on the artifact that carries the property, and prove it carries it.**

The presence precondition is not defensive decoration. It is the half of the rule that stops it
collapsing into a vacuous pass wherever a downstream stage has legitimately stripped the field. The
assertion split Story 2.2 ships is therefore three-way: `name[1]`/`name[2]`/`name[6]` and the table
set off the **shipped `.ttf`**; `usWeightClass` off the **embedded program** (which does carry
`OS/2`); `/BaseFont` matching `^[A-Z]{6}\+<name6>$` off the **produced PDF**.

**Name-table pass-through was considered and rejected.** Enabling it would have made the original
instruction literally workable, at the cost of several-folding a 716-byte CJK subset in every PDF
folio produces — to satisfy a clause PDF's own embedding model works around.

**Two further notes.**

- **Non-circularity is undisturbed.** `/BaseFont`'s two inputs are independent: the tag hashes the
  final program bytes, the name is read from the shipped face. Neither feeds the other, and the
  embedded program carries no `name` table for a tag to be written into. V5a's exclusivity assertion
  (every tag occurrence sits in a name position) was nonetheless built and red-proved.
- **A denylist entry is never coverage.** *"No name record contains 'Thin'"* is kept as
  belt-and-braces only: a face defaulting to `Light`, `ExtraLight` or `Black` sails past it. The
  positive assertions — exact family, `name[2] == "Regular"`, `name[6]` ending `-Regular`,
  `usWeightClass == 400` — are the guard.
---

### D-000.24 — "Forward guard with no available red-proof" is a named, permitted category

*(mechanism: binding)*

**The rule.** A guard whose red-proof is **not currently constructible** is labelled as such —
explicitly, in the assertion's own text — and is **never credited with a red-proof it does not have.**

**Why it needs to be a named category rather than a judgement call.** Until now the project's
vocabulary had two states: *guarded* (with a red-proof) and *unguarded*. That forces a false choice
whenever a property is worth asserting but cannot currently be made to fail — and the pressure is
always toward the dishonest option, because "unguarded" reads as a gap and "guarded" reads as done.

**A guard credited with a red-proof it does not have is worse than one openly labelled unproven**,
because the label tells the next reader **where to look**. A false credit tells them the opposite.

**The instance that named it.** Story 2.3's `YOffset` assertion. `epics.md:751` names GPOS
mark-to-base as Thai's mechanism, but `YOffset` measures **0 for every glyph of every sample across
all three shipped faces** — an AC phrased `assert YOffset != 0` would be **vacuously false and would
block a correct implementation.** The real mechanism for `ปั` is **GSUB selecting a lowered mark
form** (`uni0E49.small`), with GPOS active only **horizontally** (`@-29,0`). So the assertion is worth
keeping as a forward guard against a future face that does use vertical positioning — and there is no
way to red-prove it today.

**Do not manufacture a red-proof to fill the gap**, and do not weaken the assertions around it to
match its status. Story 2.3 asserts all five position fields from a declarative table; only `YOffset`
carries the label.

---

### D-2.3.1 — Cross-validate shaping against HarfBuzz itself, as a one-time offline oracle

*(mechanism: binding)* — resolves D-2.3-Q2; **amends `epics.md:753` under D-000.6.**

**Why the mandated library is out.** `epics.md:753` mandates cross-validation against
`go-text/typesetting`. It is **unimplementable**: `gomod_test.go:59`'s `wantModuleGraph` asserts
`go list -m all` equals exactly two modules, and that includes test dependencies.

**And it would have been weak evidence even if possible.** textshape's `README.md:5` says *"The
shaping logic is a **port of HarfBuzz**"*, and `:125` credits *"[textlayout] — Earlier Go HarfBuzz
port that **inspired this implementation**"*. The lineage is **HarfBuzz → textlayout → both siblings**.
Two implementations descended from one source agreeing tells you far less than the AC's phrasing
implies — it mostly re-measures their common ancestor.

**Also rejected: textshape's own `harfbuzz-tests/` corpus.** Better than a sibling comparison, and
still weaker than it looks — **a vendor's test corpus is curated by the vendor and selected to pass.**

**Ruled: run `hb-shape` against *our* corpus, once, offline, and freeze the expectations as a
committed fixture.** `hb-shape` 14.2.0 is present at `/opt/homebrew/bin/hb-shape`. Two precedents make
this the established shape here rather than an invention:

- **Story 1.1's `qpdf --check`** — an external CLI used as an independent oracle *specifically* to
  avoid a module dependency (AC2). Identical situation.
- **AD-25, verbatim**: *"whose provenance may be a **one-time offline reference run, hand-checked —
  never a runtime dependency**."* Exactly this, applied to shaping instead of Thai breaking.

No module dependency, no runtime dependency, `wantModuleGraph` untouched, and the oracle is **the
reference implementation** rather than a cousin of the thing under test. The `hb-shape` version and
the **exact invocation, verbatim** go into the fixture's provenance.

**The amendment** *(mechanism: binding)*: `epics.md:753` names a **library**; amend it to name the
**outcome** — cross-validation against an independent reference implementation. **State the invariant,
not the mechanism**; a clause naming a mechanism breaks when the mechanism moves.

---

### D-2.3.2 — `/ToUnicode` allocates CIDs per (GID, Unicode-context), not per GID

*(mechanism: binding)* — resolves D-2.3-Q1.

**The round-trip framing is confirmed**: assert that text **extracts back to the original string**,
never that "every CID has an entry" — the latter is satisfiable by a map that is wrong in every entry.

**The collision the original recommendation missed.** `/ToUnicode` maps **CID → Unicode**, but the
correct answer is **context-dependent**, and the worst case is Thai and is already in our corpus.
Measured with `hb-shape` 14.2.0 against the shipped face:

```
น้ำ  →  [uni0E19 | uni0E4D | uni0E49.small@-29,0 | uni0E32]      3 runes → 4 glyphs
า    →  [uni0E32]                                                the SAME glyph, standalone
U+0E33 NFD → U+0E33            SARA AM has NO canonical decomposition
NFC(NIKHAHIT + SARA AA) → stays decomposed; nothing recomposes it
```

SARA AM (U+0E33) decomposes during shaping into NIKHAHIT + **SARA AA (U+0E32)** — and SARA AA also
occurs as an ordinary standalone character. So one glyph needs two different answers:

- map the cluster tail to **empty** → every standalone `า` in the document extracts as **nothing**;
- map it to **U+0E32** → `น้ำ` extracts as the decomposed sequence, which **does not round-trip**.
  Thai SARA AM has no canonical decomposition, so no normalisation recomposes it and **search for the
  word fails**.

**One CID cannot carry both answers.** Ruled fix, cheap and requiring no ActualText machinery:
**allocate CIDs per (GID, Unicode-context) rather than per GID.** `CIDToGIDMap` is ours, so two CIDs
may point at the **same GID** carrying different `/ToUnicode` entries — no glyph duplication, only CID
space. Anything that assumes CID↔GID is 1:1, or that inverts the map, must be checked.

**The corpus obligation is the load-bearing half** *(mechanism: binding)*: the fixture must contain
**both `น้ำ` and a standalone `า` in the same document**, asserting both extract correctly. Without
that pairing, **a per-GID implementation passes** on a corpus where the conflict never arises. This is
the both-polarities discipline applied to **fixture content** rather than to guards — a corpus that
cannot express the defect is not evidence against it.

---

### D-2.3.3 — `epics.md:756` is weak, not false, and is NOT amended

*(mechanism: binding)* — resolves D-2.3-Q3.

Story 2.3's AC9 asserts the **PDF** rather than the vendor's in-process determinism, and records the
shaper measurement as evidence. Correct: `epics.md:756`'s "byte-identical in two processes" is
**already true at baseline** (5/5 identical digests), and **an AC that cannot fail is as useless as a
guard that cannot fail.**

**But no D-000.6 amendment.** The clause is **weak**, not **false** — and D-000.6's bar is *"false,
incomplete, or unimplementable"*. Widening it to cover weakness would convert an amendment mechanism
into a **general editing licence for epic text**, which is precisely the creeping enlargement the
stated boundary exists to prevent. The story's stricter AC supersedes **by being stricter**, and
records why. Keep the bar where it is.

Contrast [[D-2.3.1]] and `epics.md:751`, both of which **are** amended: one is unimplementable against
a committed guard, the other states something **false about the shipped faces**. Weakness is not
falsity, and the distinction is what keeps the mechanism narrow.

---

### D-000.25 — The vendor boundary is unaudited, and AD-23's guard is syntactic

*(mechanism: binding)* — scoped to **one dedicated story, sequenced between 2.3 and 2.4.**

**Finding 1 — AD-23 promises more than it delivers.**

> **AD-23 promises "no float arithmetic under `internal/`" and delivers "no float *identifiers* under
> `internal/`". Those differed by nothing measurable until now.**

`arch_test.go:50-60` walks the AST for `*ast.Ident` named `float32`/`float64` and `*ast.BasicLit` of
kind FLOAT. A vendor function **returning** a float is invisible to it — the type is inferred, so the
identifier never appears. Two live call sites, both in `internal/fontset`:

```go
adv := f.face.HorizontalAdvance(gid)                    // :289  — adv is float32, invisibly
adv := f.face.HorizontalAdvance(oldGID)                 // :476  — subset width table
```

**Measured harmless today, and that is exactly what makes it dangerous.** The accessor is a pure
conversion — `float32(f.hmtx.GetAdvanceWidth(glyph))` — and every `uint16` is exactly representable in
`float32` (24-bit mantissa), so `int64(float32(x)) == x` for every possible input. **The guard is
green for a reason that has nothing to do with the guard.** The day the vendor applies variation
deltas or any scaling inside that function, the truncation becomes lossy and **nothing in the project
reports it.**

**Finding 2 — the larger one: vendor accessors substitute plausible defaults for missing data.** This
is the [[D-000.9]] class arriving from **outside our code, where none of our conventions reach.**
Four confirmed instances:

| accessor | returns when its table is absent | consequence |
|---|---|---|
| `Face.PostscriptName()` | the literal string **`"Unknown"`** | `/BaseFont /TAG+Unknown`, with every downstream assertion reporting a well-formed string |
| `Font.NumGlyphs()` | **`0`** on missing/short `maxp` | a caller looping to `NumGlyphs()` **silently does nothing** |
| `Name.FamilyName()` | indexes `entries[1]` directly | unguarded index |
| `Face.Upem()` | plain field read — **but its population path is unaudited** | **D-1.5.2 requires `unitsPerEm` validated in 16–16384 before `ScaleRound`. If construction substitutes a default for a missing `head`, that validation passes on fiction** — a ruling silently satisfied by a substituted value |

The `PostscriptName` case was caught unprompted by Story 2.2's finisher, which read `name` record 6
directly and returned `""` — **observably absent** — rather than a plausible string. That is the
correct disposition and the pattern for the rest.

**Ruled: one dedicated story, before 2.4, not folded into it.** Three reasons:

1. **2.4 adds vendor surface.** Auditing first means 2.4's new call sites are written against a known
   map; auditing after means auditing more.
2. **2.4 is already heavy** — line breaking across three scripts plus D-2.1.6's atomic-span mechanism.
   Two guard workstreams would overload the story that most needs focus.
3. Both items are **the same subject**: what crosses the vendor boundary, and what we may believe
   about it.

**Scope**: a **type-aware** AD-23 check in `lint` (the module already type-checks for map detection, so
the capability exists), red-proved by `int64(someVendorFloat())` going red; plus an enumeration of
**every vendor accessor folio calls** and what each returns when its table is absent.

**Expect the type-aware guard to go red immediately on `fontset.go:289` and `:476`.** That is the
audit finding its first two instances — **not** a problem with the guard.

---

### D-2.3.4 — Re-record `fixtures/font-text/` as an intended versioned change, with a semantic acceptance step

*(mechanism: binding)*

**The premise that failed.** Story 2.3's F3 table asserted all four existing fixtures were blind to
shaping. `fixtures/font-text/` is **not**: it renders through `folio-go/testdata/fonts/Roboto-Regular.ttf`
(upem 2048) via the chain `"body" -> ["Roboto-Regular"]`, **not** a shipped Noto face — and F3's row
also listed text the fixture does not contain (`"Hello"` against the actual `"Hello, World!"` and
`"Page footer 0123456789"`).

Roboto's GPOS kerns two pairs in that text. `hb-shape` 14.2.0, kerning on versus off:

```
W (gid59):  1786 with GPOS  ·  1817 with --features=-kern   → −31 font units  (−15 @1000-em)
P (gid52):  1281 with GPOS  ·  1292 from hmtx               → −11 font units  (−6  @1000-em)
```

Independently confirmed before ruling: the current golden's `/BaseFont` is `HXRYNT+Roboto-Regular`,
and it contains **2 `Tj` and 0 `TJ`** at 22,299 bytes. **The fixture literally records unkerned
output.**

**Ruled (A): re-record.** AD-21 requires a hash change be treated as a defect *"until proven to be an
intended, versioned behaviour change."* Proven here: the reference implementation confirms both
values, the mechanism is understood (GPOS applied where it previously was not), exactly two pairs
move, exactly one fixture changes, the other three are byte-identical, and no new failure appears.

**The decisive argument is not that the premise was false — it is that the old bytes are wrong.** They
draw `Wo` and `Pa` unkerned because Folio ignored GPOS. **Re-recording retires a regression rather
than accepting one**, and it converts `font-text` from a fixture blind to this story into one that
observes it — the direction F3's own argument runs in.

**No alternative exists.** Keeping the hash by disabling kerning for one document would be **crippling
the product to make a golden pass** — worse than regenerating a golden to make a test pass. And there
is no third option: the old golden pins a pre-shaping code path that no longer exists.

**The binding condition — first golden re-recording since [[D-000.22]], and a bare hash swap is what
that rule exists to prevent:**

1. Assert the specific `TJ` adjustments **by value, not by hash** (−31 and −11 at upem 2048, scaled).
2. Assert the operator is **`TJ`, not `Tj`** — the fixture now exercises a branch it never has.
3. Record the `hb-shape` version and **exact invocation, verbatim**, in the fixture's README.

**Why that is not bookkeeping.** A future change that silently loses kerning produces a **hash
mismatch** — and the cheapest available response to a hash mismatch is **to re-record.** Assertions
naming the expected adjustments make re-recording a kerning regression **impossible without deleting
an assertion**, which is visible in a diff where a changed digest is not.

---

### D-000.26 — Cite the subject of a measurement, not just its result

*(mechanism: binding)*

**The rule.** Every recorded measurement must **name the artifact it measured**, precisely enough that
a reader can tell whether it is the right one — the face, the file, the exact input string, the
invocation. Not only the number it produced.

**The mechanism, which inverts the usual intuition.** A **mis-aimed measurement propagates further
than a stated assumption**, because *nobody re-checks a measurement.* A table labelled "measured"
**transfers its authority** to whatever it says; an assumption invites the reader to test it. So the
expected cost of a mis-aimed measurement **exceeds** that of an unstated assumption — even though
measuring is otherwise strictly better.

**The remedy is not to measure less.** It is to make the subject checkable on the page. F3's row said
*"font-text: `Hello`"*. Had it said *"font-text, via `Roboto-Regular.ttf` (upem 2048), text
`Hello, World!`"*, the error would have been visible **without re-measuring** — because the wrong face
and the wrong string are both wrong **in the citation**, not only in the conclusion.

This is [[D-000.21]] applied to the thing you are **characterising** rather than the thing you are
**guarding**, and it is the same question as the gate note's *"ask what the number is a number of"*,
asked from the opposite end.

---

### D-000.27 — Evidence that the decision log propagates rules to agents who never saw them made

*(mechanism: illustrative — an observation about the process, recorded for the Epic 2 gate)*

Two rules transferred this week to agents that were **not present** when they were made:

1. **Story 2.3's creator** independently produced *"forward guard with no available red-proof"* — the
   category later ruled binding as [[D-000.24]] — by refusing to claim a red-proof for the `YOffset`
   assertion it could not construct.
2. **Story 2.3's developer** applied the base-36 rule (*if the other hypothesis predicts the same
   observation, you have learned nothing*) unprompted: it noticed `"office"` → 0,1,4,5 is **ambiguous**
   between rune-index and byte-offset because ASCII makes them identical, and disambiguated with
   `"ณัฐวุฒิ"` → 0,0,2,3,3,5,5, where byte offsets would be multiples of three. `hb-shape` agrees.

Each agent is a **fresh cold spawn** with no access to the conversation in which those rules were
made. **The rules are propagating through the artifacts rather than through the conversation** — which
is the entire purpose of writing them into this log, and the first evidence in this run that the log
is doing work **beyond audit.**

---

### D-000.28 — Anticipatory boilerplate is a distinct failure from narration drift

*(mechanism: binding)*

Narration has now asserted something the artifact does not support **three times** in this run. Twice
as **drift**; once, in Story 2.3, as **anticipatory boilerplate**. They need separating, because
**their remedies differ**:

| | what it is | how it is fixed |
|---|---|---|
| **Drift** | a record that **was** true and diverged | **re-derive the record from the artifact** |
| **Anticipatory** | a record written **before** the fact it asserts, therefore **never** true | **do not write the claim until the event** — re-deriving cannot help, there is nothing to re-derive from |

**Anticipatory is the more dangerous of the two.** Drift leaves a trail — the claim was true at some
commit, so history can adjudicate it. **An anticipatory claim is false from birth and reads identically
to a true one.** No amount of checking the record against history will expose it, because history
never supported it either.

**The instance.** `fixtures/shaped-text/README.md:89-92` and `shaped_fixture_test.go:660` both assert
**"A human read the rendered Thai before this hash was frozen"**, citing the story's completion notes
as evidence — and those notes say the check is **outstanding**.

**It is a process finding, not an integrity one**, and that distinction is recorded deliberately: the
developer reported the gap **plainly and unprompted** in the same run, and recommended the story not
be called done without a Thai reader. The boilerplate was written ahead of an event that was then
honestly reported as not having happened.

**Worth noting without belabouring**: the false credit sits inside a comment *citing the Thin-Chinese
lesson about guards that do not check what they claim.* **The comment was right about the class and
wrong about itself.**

---

### D-2.3.5 — "Frozen with sign-off pending" is a legitimate state, under three conditions

*(mechanism: binding)*

**It does not hollow out [[D-000.22]]**, because that rule's own wording already separates the two
claims: the semantic assertion is *"separate from and additional to the hash."* **Byte-stability and
semantic correctness are different claims about a fixture, and freezing one does not consume the
other** — provided the artifact says which one is outstanding.

But whether it hollows the rule turns **entirely on the tracking mechanism**, so:

1. **The pending obligation is a FAILING TEST, never a log entry** *(binding)*. The fixture directory
   carries a sign-off record naming the reader, the date, and what they examined; a **gate-run test
   fails while it is absent**, gated as the matrix legs are — so the story commits green and the
   **gate cannot pass**. **A note is optional; a red gate is not.** Every deferral in this run that
   held did so because something mechanical eventually fired.
2. **The sign-off names the hash it signed off** *(binding)*. Otherwise a later re-recording carries a
   sign-off for **different bytes** — precisely the stale-provenance failure closed yesterday on
   derived fonts. Binding sign-off to digest makes a re-record **automatically invalidate it** and
   demand a fresh look. That is the anti-rot property, and it costs one line.
3. **The artifacts say "pending", never anything stronger** — already bound by [[D-000.24]].

**The boundary that keeps the rule's teeth** *(mechanism: binding)* — stated explicitly so "pending"
cannot become the default:

> **D-000.22's machine-checkable half is never deferrable. Only its irreducibly-human half is, and
> only against a mechanical blocker.**

Weight class, name records, `TJ` adjustment values, axis pins, byte-identity across targets — all
asserted **at recording**, with no deferral available. Only *"does this Thai read correctly to someone
who reads Thai"* may be pending, because it is the one thing no assertion substitutes for.

**The evidence for that split is our own: the Thin-Chinese defect was caught by a machine-checkable
property, not by a human look.** The machine half is where nearly all the value has come from, so it
stays mandatory — and confining "pending" to the human half means the rule loses almost nothing while
remaining livable for a one-person project.

**Applied here**: Story 2.3 commits; the Thai sign-off blocks the **Epic 2 gate**, not the story. A bad
result lands as its own commit against 2.3, exactly as a red catch-up run would. Holding a finished
story on a human availability window is the stall the standing instruction exists to prevent, and this
path is genuinely recoverable.


---

### D-2.3.6 — Story 2.3 finisher follow-ups: the segment cursor, the sign-off gate, and two citation repairs

*(mechanism: record — appended as a sibling entry, never as an edit to [[D-2.3.1]]…[[D-2.3.5]], which
this log's own header makes append-only.)*

Filed by Story 2.3's finisher. Twelve QA findings triaged **11 FIX · 0 DISMISS · 1 DEFER**. Four
consequences are worth a standing record because they change something a later story would otherwise
re-derive or re-break.

**1. Shaping moved into `splitByFace`, and the second derivation was deleted rather than corrected.**
The segment cursor summed raw `hmtx` advances while runs were drawn kerned — text drawn kerned and
placed unkerned, measured at **640 millipoints for `"AV ก"` at 16 pt** in the shipped chain. The fix
is not "sum the right numbers": each face-segment is now shaped once, in the same pass that computes
its origin, and the drawn glyphs and the next segment's origin come from **one** answer. **Do not
reintroduce a second shaping call in `renderDocument`** — the defect was two derivations agreeing
until they didn't, and the structure is what prevents it, not the arithmetic.

**`fontset.AdvanceForRune` is deliberately retained with no caller.** It is one of the two
`ot.Face.HorizontalAdvance` (`float32`) sites [[D-000.25]] hands to
`2-3a-audit-the-vendor-boundary`; deleting it would silently shrink that story's subject from two
sites to one. The count is **still two**.

**2. A fixture can be green by accident of ordering, and that is worse than an absent fixture.**
`fixtures/shaped-text/` was blind to this defect for a reason that had nothing to do with its text:
its only shape-observable Latin segment was the **last** segment of its element, so the wrong cursor
was never consumed, and the one genuine mixed-run element happened to carry no kern pair. The golden
was green over a live defect **while reading as coverage in every report**. Generalising, and this is
the part to keep: *a fixture that exercises a property only in a position where the property's output
is discarded certifies nothing about it.* Element `e7` (`"AV ก"`) was added to close it, and
`TestFaceSegmentOriginsUseShapedAdvances` states the property in hand-derived literals cross-checked
against `hb-shape` 14.2.0 rather than by recomputing production's own arithmetic.

**3. [[D-2.3.5]] is implemented, and the gate is red as committed.**
`TestShapedTextThaiSemanticSignOffIsRecorded` is `//go:build matrix`, so 2.3 commits green and the
**Epic 2 gate cannot pass** until `fixtures/shaped-text/thai-signoff.json` names a reader, a date,
what they examined, and the digest `5964aad0…92e00f`. All four legs red-proved (absent → fail;
matching → pass; stale digest → fail; empty field → fail). The boundary — *D-000.22's
machine-checkable half is never deferrable; only its irreducibly-human half is, and only against a
mechanical blocker* — is recorded verbatim in the fixture README so it is not re-derived.

**4. Two citation repairs, and the class they share.** `epics.md`'s two Story 2.3 blocks had **swapped**
ruling citations, and the `:751`/`:753` block cited **D-2.3.3** — whose text is *"…is NOT amended"* —
as authority for an amendment. Corrected to **D-000.24** (clause 1) / **D-2.3.1** (clause 2), and
**D-2.3.3** on the `:756` block. **A citation that reads as support and inverts on inspection is the
same defect as a fixture README claiming a human check that never happened** — both hand a reader
evidence that says the opposite of what they were told. [[D-000.26]] is the general form; these are
two instances of it found in one story, in different artifact kinds.

**Also recorded, because the measurement is reusable.** `internal/fontset.Subset` accepted glyph ids
the face does not have — the vendor **fabricates** rather than reports (`Subset([65535])` against a
4,515-glyph face returned a 460-byte program, no error). Now validated before `AddGlyphs`. And
`TestPermutationInvariance` was **demonstrated** unable to detect a folio-side sort: injecting
`slices.Sort` into `Subset` reddens the new AST guard while that test stays PASS in the same run,
because the vendor sorts below the call. An AST scan is the only mechanism that can see it.

**DEFER:** one, **DW-14** — `/ToUnicode`'s unbounded `beginbfchar` section against the spec's
100-entry cap. Pre-existing, unreachable today, largest section measured at 28.
---

### D-2.2.4 (correction) — one supporting premise was false; the ruling stands on independent grounds

*(mechanism: binding)* — **appended, not rewritten**, per the record's append-only discipline.

**The false premise.** [[D-2.2.4]]'s correction 3 stated that *"pin every axis to an explicitly chosen
value"* was **unimplementable** under AD-23, because reaching the vendor's `PinAxisLocation` requires
the identifier `float32`, which `arch_test.go:54` bans.

**That is false, and it is testable in one line.** `fontset_test.go:515` calls
`in.PinAxisLocation(ot.MakeTag('w','g','h','t'), 700)` **today, with the guard green.** An untyped
constant converts to a `float32` parameter with **no `float32` token anywhere**, and `700` is an
`*ast.BasicLit` of kind **`INT`** — invisible to a `token.FLOAT` scanner. The same false claim stands
in four places the story corrects: `fontset.go:112`, `:425`, `:494`, and `internal/text/shape.go:71`.

**How it got accepted, which is the instructive part.** It was accepted **in the same conversation
that established the guard sees identifiers, not types.** Those are *the same fact from opposite
sides*: the blind spot hiding `int64(vendorFloat())` **also** hides `PinAxisLocation(700)`. The
blind spot was learned in one direction and not turned around.

**Grounds that survive, stated explicitly** — none depends on the float argument:

1. **Payload** — 4.82 MB compressed against 8.30 MB for the variable build.
2. **`usWeightClass` correctness** — `textshape` copies `OS/2` verbatim and never writes the field;
   `fontTools` sets it explicitly when instancing.
3. **Deleting the FMA path** — no `fvar`/`gvar` in a shipped face means instancing never runs.

**Why this is logged rather than quietly left.** *"The conclusion survived"* is exactly the reasoning
that lets a bad argument stay in the record. The concrete harm is that a later reader relying on
D-2.2.4 would **inherit a premise that is false** and might build on it. Corrections append; they do
not rewrite.

---

### D-2.3a.1 — Validate table presence at face ingestion; a panic on caller bytes is a convention violation

*(mechanism: binding)* — resolves 2.3a's Q1.

**The defect.** `folio.Render` **panics** on caller-supplied font bytes whose `maxp` is missing or
short: `NumGlyphs()` returns **0**, `NumberOfHMetrics` is 1294, so `ot.ParseHmtx` reaches
`make([]int16, -1294)` → `makeslice: len out of range`, via `ot.NewFace` ← `fontset.go:70`.
`ParseHmtxFromFont` guards this; `NewFace` does not. `folio.FontSet` is a **public
`map[string][]byte`**, so this is untrusted input reaching a panic through the documented entry point.

**It is a stated-convention violation, not a robustness nit.** The spine's Consistency Conventions
say verbatim: *"Nothing in `internal/` panics on malformed input — **untrusted font and template
bytes** return diagnostics."*

**Fixed in 2.3a, recorded against Story 1.5**, which shipped the `NewFace` call site — the same
disposition as Story 1.6's `decodePoints` hang: fixed in the story that **surfaced** it, noted on the
story that **shipped** it.

**Do not fix only `NumGlyphs`.** The audit found a second instance of the same class, and the silent
one is worse:

| table stripped from Roboto | outcome through the public entry |
|---|---|
| `maxp` | **PANIC** — loud |
| `OS/2` | **renders, `/CapHeight 928` where the intact face gives `711`** — silent, wrong, shipped |

One class — *folio reads a table that may be absent and receives a substituted default* — with a loud
instance and a silent one.

**Ruled: validate at face ingestion that every table folio actually reads is present**, failing with a
**located error naming the face and the missing table**. Scope it to the tables the half-2 enumeration
identifies as read, so **the two halves compose rather than compete** — the audit produces the list
the guard consumes. **Do not require `name`**: Story 2.2 deliberately tolerates a nameless program,
and requiring it would reverse a ruled disposition.

**The symmetry is the point:** *"assert the table exists before asserting on its contents"* was made
binding for **guards** in [[D-2.2.6]]'s amendment. This is the same rule applied to **production
reads** — prove the table exists before reading it. One discipline, two sites.

---

### D-2.3a.2 — A substituted `/BaseFont` must be diagnosed, not silent

*(mechanism: binding for the property; illustrative for the mechanism)* — resolves 2.3a's Q2.

`readPostScriptName` correctly returns `""` for a nameless program (Story 2.2's ruled disposition),
but `internal/pdf/textdoc.go:334-336` then **substitutes the FontSet key back in**. Measured: a
nameless program still emits `/BaseFont /HXRYNT+Roboto-Regular`.

**`/BaseFont` is Required and must be a legal name, so something must go there**, and the FontSet key
is genuinely true of the face we loaded. It is not ruled absent.

**But the silent substitution defeats the property [[D-2.2.6]] was bought for.** Its third reason was
that `/BaseFont` **makes an invisible property visible** — a Thin-named program shows as
`NotoSansSC-Thin`. Under silent substitution a **nameless** program is **indistinguishable from a
correctly-named one**, so a reader cannot tell whether `/BaseFont` reflects the program or reflects
us.

**Ruled: diagnose it** — the same disposition as a caller-supplied variable face. Do not reject, do
not silently substitute: **name it.** AD-8's *"not a silent substitution"* governs, and it costs one
diagnostic. **A code comment documents the behaviour for a maintainer; it does nothing for whoever
reads the PDF**, which is the audience D-2.2.6 was written for.

This is [[D-000.25]]'s pattern reappearing **one layer above** where Story 2.2 closed it.

---

### D-000.29 — An audit row ends settled or fixed, never carried

*(mechanism: binding)*

**An audit that converts each suspicion into either "confirmed and fixed" or "traced and closed" is
doing its job. One that ends with open questions has merely relocated them.**

**The precedent that named it.** [[D-000.25]] flagged `Face.Upem()`'s population path as unaudited,
with a specific worry: D-1.5.2 requires `unitsPerEm` validated in 16–16384 before `ScaleRound`, so if
construction substituted a default for a missing `head`, **our validation would pass on fiction.**

Traced: **the vendor does substitute 1000 — and folio never reads it.** D-1.5.2's validation does not
pass on fiction. The row ends **settled**, with the reason recorded, rather than carried forward as a
maybe.

A carried row costs more than an unopened one, because it **reads as diligence** while providing none
— and the next reader must redo the tracing to learn whether anything was ever wrong.

---

### D-000.30 — A fix destroys the red-proof for the guard against it; capture the proof before the fix

*(mechanism: binding)*

**The phenomenon.** When a defect is closed, the path that produced it **stops existing** — so the
guard written against that defect can no longer be red-proved by reproducing it. The fix and its own
evidence are in tension, and the tension is structural, not a mistake by anyone.

**The instance.** Story 2.3a asserted that stripping `OS/2` from a face yields `/CapHeight 928` where
the intact face gives `711` — a silently wrong value in shipped output. The ruled fix
([[D-2.3a.1]]) validates table presence at ingestion, so **a face with no `OS/2` no longer reaches the
renderer at all.** The `928` half of the assertion became **unconstructible by the fix that made it
matter.**

**The three wrong responses, all of which this run has seen elsewhere:**

1. **Weaken the guard** so the remaining half looks like full coverage.
2. **Manufacture a red-proof** — reintroduce the defect behind a flag, or assert on something adjacent
   and call it the same thing.
3. **Credit the guard with the proof anyway**, on the grounds that it *was* true once. This is
   [[D-000.28]]'s anticipatory boilerplate wearing the opposite tense.

**The correct response, and what 2.3a did:** measure the defect **before** the fix — in a worktree at
the pre-fix commit — record the measurement with its subject cited ([[D-000.26]]), and then label the
now-unconstructible half a **forward guard with no available red-proof** ([[D-000.24]]). The live half
is asserted on produced bytes behind a presence precondition; the refusal itself is asserted.

**The distinction that makes this honest** *(mechanism: binding)*:

> **A one-time measurement taken before the fix is evidence the defect was real. It is not a standing
> red-proof, and must never be recorded as one.**

The first says *"this was broken, here is the number."* The second says *"this test fails if it breaks
again."* Both are worth having; **only the second protects the future**, and conflating them
manufactures confidence that no mechanism supports.

**Consequent obligation** *(mechanism: binding)*: when a story fixes a defect it also guards against,
**capture the red-proof before applying the fix** — at the pre-fix commit, in a worktree — or state
plainly that no standing red-proof exists. The window in which the proof is constructible closes
permanently when the fix lands, and it cannot be reopened afterwards without reintroducing the defect.

---

### D-2.2.4 (correction, amended) — the correction's own enumeration was sampled, not swept

*(mechanism: binding)* — appended to [[D-2.2.4]]'s correction, which was itself a correction.

The correction entry stated the falsified `PinAxisLocation` claim stood in **four** places. Story
2.3a's review found at least **two more**:

- `folio-go/internal/fontset/fontset.go:751` — *"explicitly is both unreachable (AD-23 bans
  `float32`)"*, uncorrected.
- `folio-go/internal/fontset/fontset_test.go:466` — *"needs the identifier `float32`, which AD-23's
  arch guard bans"* — sitting **three lines above line 515**, which is the exact call that disproves
  it, and contradicted by its own next sentence.

**The enumeration was written from the sites someone happened to have been shown.** That is
[[D-000.23]] in a new dress: *a correction written in response to instances covers the instances, not
the class.* The remedy is the same one that rule already names — sweep for **every form the claim
takes** (`PinAxisLocation`, "unreachable", "AD-23 bans", "identifier `float32`", "arch guard bans"),
across `.go`, `.md` and story files, production **and** test.

**Why a sampled correction is worse than an uncorrected claim.** An uncorrected claim is uniformly
wrong, and a reader who catches it anywhere distrusts it everywhere. A **partially** corrected claim
is *right in the places someone looked and wrong elsewhere* — and the corrected sites lend the
surviving ones credibility, because the text now reads as maintained. The 2.3a finisher is producing
the swept list and its count so the record finally holds an enumeration rather than a sample.


---

### D-2.3a.3 — Story 2.3a finisher: the swept enumeration D-2.2.4 (correction, amended) asked for

*(mechanism: record)* — the sibling follow-up entry to [[D-2.3a.1]] and [[D-2.3a.2]]. It does not
amend them, and it does not edit [[D-2.2.4]] (correction, amended); it **discharges** that entry's
closing sentence, which said *"The 2.3a finisher is producing the swept list and its count so the
record finally holds an enumeration rather than a sample."*

**The count is 8 carrier sites.** Swept for all five forms of the claim — `PinAxisLocation`,
"unreachable", "AD-23 bans", "identifier `float32`", "arch guard bans" — across **every file type in
the repository**, production and test, source and record.

| # | site | found by |
|---|---|---|
| 1 | `folio-go/internal/fontset/fontset.go` — `New`'s D-2.2.4 block (baseline `:112`) | D-2.2.4 (correction) |
| 2 | `folio-go/internal/fontset/fontset.go` — `Subset` doc comment (baseline `:425`) | D-2.2.4 (correction) |
| 3 | `folio-go/internal/text/shape.go` — `Shaper` doc comment (baseline `:71`) | D-2.2.4 (correction) |
| 4 | `folio-go/internal/fontset/fontset.go:751` (baseline `:494`) | D-2.2.4 (correction) — **enumerated but not executed** |
| 5 | `folio-go/internal/fontset/fontset_test.go:465-471` | Story 2.3a review |
| 6 | `tools/fontgen/instance_faces.py:16-19` | **this sweep, and nothing before it** |
| 7 | `_bmad-output/…/2-2-…md:1098-1104` | this sweep |
| 8 | `_bmad-output/…/2-2-…md:1857-1859` | this sweep |

**Site 6 is the finding, and it sharpens [[D-000.23]] rather than merely instancing it again.** The
ruling's enumeration searched `.go`. The review's sweep searched `.go` and `.md`. Both were *sweeps*
by intent and *samples* by construction, because each inherited the previous pass's idea of where the
claim could live. `tools/fontgen/instance_faces.py` is a Python build tool whose module docstring
gave the falsified claim as its third reason for shipping static faces — a load-bearing statement, in
the file that actually performs the derivation, invisible to every search anyone had run.

> **A sweep is defined by the class of the claim, never by the file types the last search happened
> to cover** *(mechanism: binding)*. Widening the search terms while leaving the search *surface*
> inherited is not a sweep; it is the previous sample re-run with better regexes.

**Consequent obligation** *(mechanism: binding)*: when [[D-000.23]] requires a correction to cover a
class rather than its instances, the sweep must state **which file types it searched** and why that
set is exhaustive for the claim. A correction that cannot say what it searched has not swept.

**Three term-matches are not carriers**, verified individually rather than corrected on sight:
`folio-go/internal/fontset/vendor-boundary.md` and `lint/internal/rules/floattyped_test.go` both
state the claim *in order to record that it is false*, and Story 1.5's record at `:636-639` is
**accurate as written** — it says the guard flags *"the bare identifier"* and that *"a call site that
**names the type** would go red"*, which is conditional and true. Correcting it would have introduced
an error. **A sweep must be able to distinguish a claim from a quotation of it**, or it will
manufacture the drift it was sent to remove.

**Two dispositions, deliberately different, for the same claim in different kinds of document.** The
six live-source sites are corrected **in place**, each keeping its "this used to say X, that was
false, here is why" form. The two in Story 2.2's closed record are corrected by an **appended
section** under [[D-1.6.6]] — *appended, never rewritten* — and the instances inside this decision log
are left exactly as written, because this document is append-only by its own header. Erasing the
original wording would destroy the evidence that the claim was ever believed, which is the entire
subject of D-2.2.4 (correction, amended).

**Corrected text states the real grounds**, never merely deletes the claim: payload (**4.82 MB**
compressed against **8.30 MB**), `usWeightClass` correctness (`textshape@v0.0.15`
`subset/execute.go:496-499` copies `OS/2` verbatim and has no writer for the field), and deleting the
FMA/interpolation path. None of the three ever depended on the false premise, which is why the
conclusion survives its collapse intact.

---

### D-2.3a.4 — Story 2.3a finisher: a position-bound citation restales on the next edit

*(mechanism: binding)* — a consequence discovered while discharging [[D-2.3a.3]], recorded because it
recurred immediately and will recur again.

Correcting carrier site 5 lengthened a comment block in
`folio-go/internal/fontset/fontset_test.go` and moved the call that **disproves** the falsified claim
from line 515 to line 529. **Five** separate citations to `fontset_test.go:515` — in `fontset.go`
(three), `internal/text/shape.go` and `lint/internal/rules/floattyped_test.go` — went stale in the
same edit that corrected the claim they supported.

> **Cite the thing, not its coordinates.** A citation to a line number is a claim about a position,
> and positions are invalidated by edits that change nothing the citation is about. Cite the test or
> function **by name**, which survives every edit that does not remove the thing itself.

All five were re-cited as `TestSubsetPinnedInstancesProduceDifferentTags`, not renumbered — because
renumbering restores the citation and preserves the defect. This is the same shape as [[D-1.6.3]]'s
hard-coded filename and Story 2.3a's own Finding 7 (a guard excusing a file **by filename**): three
instances now of *a reference bound to a name or position rather than to the property it is about*.

**Consequent obligation** *(mechanism: binding)*: a citation in committed source to a specific line
of another file must name the declaration it points into. A bare `file.go:NNN` in a comment is a
defect at rest — it is correct only until someone edits above it, and nothing in the build will ever
say otherwise.
---

### D-2.4.1 — The unbreakable-field declaration is a document-level list of bare root-relative dotted data paths

*(mechanism: binding)* — resolves DN-1. This is the concrete form of the owner's D-2.1.6 ruling.

**Element-level granularity is wrong, definitively — and the format's own worked example proves it.**
`folio-format.md:130` defines a text element's `value` as *"the string, **which may contain `{{ }}`
bindings**"*, and both canonical examples **mix literal text with bindings**:
`"Statement for {{customer.name}}"` and `"Page {{page}} of {{pages}}"`. Marking such an element
unbreakable would forbid breaking between *"Statement"* and *"for"* — destroying wrapping for exactly
the shape the specification itself demonstrates. Not a preference; **a defect the canonical example
would exhibit on day one.**

**Sub-element granularity has two candidate homes, and one is closed:**

- **Inside the binding syntax** (`{{customer.name|atomic}}`) — **rejected.** It grows the expression
  language, which counter-metrics C1/C2 exist to make visible, and it collides with D-1.6.5 (Story
  1.6 accepts *only* bare dotted paths, with Epic 3's parser replacing the matcher). The syntax would
  be **paid for twice**.
- **A list of data paths** — no new syntax, no expression-language growth. **Chosen.**

**Ruled: a document-level list of bare root-relative dotted data paths**, spelled exactly as
`footerOf` already is (D-1.4.1: *"a bare root-relative dotted value path… No `{{ }}`, no function
call, no `[]`"*). Row-scoped paths are expressed root-relative under the same convention, so the
format carries **one** path convention rather than two.

**Document-level rather than per-element, because the property belongs to the data, not to a box.** If
`customer.name` holds a name in the header, it holds one in the footer. Declaring it once is DRY and
semantically honest, and it gives Epic 6's binding UI a single place to offer *"this field is a name"*.

**Why this is not an owner decision** (the orchestrator leaned that way and was wrong): the
granularity is **forced by the format's own structure**, so there is nothing to fork on. And the
MAJOR-bump concern does not attach — a new **optional key** is a MINOR addition under D-1.4.12; only
extending a **closed set of values** requires MAJOR, and this is a list, not an enum.

**Surface it at the Epic 2 report as a statement, not a question** — the concrete form of what the
owner already ruled. Nothing is foreclosed before `v0.1.0` if they want the spelling different.

---

### D-2.4.2 — Fix the leading criterion before the measurement; let the measurement choose within it

*(mechanism: binding)* — resolves DN-2.

The leading rule is **welded into every golden recorded after it**, so inheriting whatever the
implementation does first is the failure mode. But picking a metric **blind** would be choosing by
taste where this project has consistently chosen by measurement. So the **criterion is fixed before
the data** — the D-2.0.2 discipline applied to a design choice rather than a spike.

**Three binding constraints:**

1. **Leading is a function of (declared font chain, font size) — never of the glyphs that happen to
   appear on a line.** If line height varied with content, adding one CJK character would reflow the
   element, and AD-24's *"boxes are absolute, and nothing negotiates"* forbids precisely that. **This
   constraint decides most candidates on its own.**
2. **It reads only tables 2.3a's presence precondition requires, and never a substituted default.**
   We have just measured what a missing `OS/2` produces — `/CapHeight 928` against a true `711`.
   Leading must not inherit that class of fiction.
3. **The rule is stated in `folio-format.md`**, because it is welded into goldens and is therefore
   **public behaviour**, not an implementation detail.

**Then measure the candidates and report the numbers** — `hhea` ascent/descent/lineGap versus `OS/2`
typo metrics versus a size multiple, across the actual shipped chain. **If exactly one survives the
constraints it is chosen; if more than one survives, the measurements come back for a ruling.**

---

### D-2.4.3 — The Thai break sign-off is a second, separately-bound human obligation

*(mechanism: binding)* — resolves ESC-2; the Epic 2 gate now owes **three** things.

*"Does this Thai **read** correctly"* and *"do these **break points** fall correctly"* are **two
distinct human judgments**. One sign-off covering both is the conflation [[D-000.23]] warns about,
with an added hazard: **a reader signing off on legibility would not know they were also certifying
break placement.**

So: a **separate** `//go:build matrix` sign-off bound to **its own digest** (per [[D-2.3.5]]'s
condition 2), leaving `thai-signoff.json` and `5964aad0…c92e00f` untouched — **a re-record then
invalidates exactly one and not the other.**

**DN-4, confirmed: no third sign-off for the wrapped-PDF golden.** Its human judgment is break
placement, which this fixture already covers **at the layer where it is legible**. A second sign-off
over the same judgment one artifact downstream is ceremony — and worse, **it would dilute the first by
making sign-offs feel routine.**

---

### D-2.4.4 — `epics.md:823`'s UAX #14 clause names a mechanism; amend it to name the outcome

*(mechanism: binding)* — resolves DN-3, same shape as [[D-2.3.1]].

The clause requires breaks *"at whitespace per UAX #14"*. **Not implementable**: it needs a new
module, and `wantModuleGraph` forbids one. Amend under D-000.6 to name the **outcome** — breaks occur
at whitespace and at script-appropriate opportunities — so the clause does not break again when the
library landscape moves. **State the invariant, not the mechanism.**

---

### D-000.31 — An agent checking its brief against the source is now load-bearing, not incidental

*(mechanism: illustrative — recorded for the Epic 2 gate)*

**Three times in this run an agent has checked its instructions against the source and found the
instructions wrong.** The most consequential: Story 2.4's brief listed
`TestP2IndependentDPCrossCheck` as an expected pre-existing failure, when **D-2.1.9 routes that exact
defect to Story 2.4 by name.** Following the brief would have **shipped 2.4 without its central
fix** — the story would have preserved the bug it exists to close, and reported green.

The orchestrator wrote that brief. The correction came from reading the ruling rather than the
instruction.

**The corollary for how briefs should be written**: name the *sources* an agent must reconcile
against, not only the conclusions drawn from them. A brief that hands down conclusions alone gives an
agent nothing to check them with.

**And the related instinct worth keeping**: the creator asked that the matrix-cadence waiver be
*"logged as a ruling rather than silently retired"* — and the answer turned out to be that **there was
never a waiver, only an orchestrator error** (D-000.4 names 2.4 as a per-story override explicitly).
**Asking for a deviation to be logged is how you discover it was never sanctioned.**

---

### D-2.4.2 (resolved) — Leading is `hhea` A−D(+lineGap), taken as the MAXIMUM over the declared chain

*(mechanism: binding)*

**The metric: `hhea` ascent − descent (+ lineGap).** `hhea` is in `requiredTables` and validated at
ingestion with a located error, so **no substituted default is reachable**. `lineGap` is 0 on all
three shipped faces — byte-neutral to include, and honest for a face that declares one.

**The face: the MAXIMUM over the faces of the declared chain present in the FontSet.** Measured on the
committed faces, scaled to the 1000-unit em:

```
face             hhea A  hhea D  gap    A−D   USE_TYPO_METRICS
Noto Sans          1069    -293    0   1362   True
Noto Sans Thai     1061    -450    0   1511   True
Noto Sans SC       1160    -288    0   1448   False

(a) first face of chain = 1362      (b) max over chain = 1511
at 16pt: (a) gives 21.79pt · Thai requires 24.18pt · shortfall 2.38pt
```

**(a) is disqualified, not out-argued.** It ships **below-vowels colliding with the next line's
above-vowels in the default chain, on the script this epic exists to support** — and welds that into
every subsequent golden.

**Constraint 1 correctly declined to separate them, which is a constraint working rather than
failing.** It forbids *content-dependent* leading; **(b) depends on the *declared* chain, not on the
glyphs that appear**, so adding a CJK character reflows nothing and AD-24 holds exactly as under (a).

**The formulation that makes (b) positively correct** *(binding — stated in `folio-format.md`)*:

> **A chain declares what may appear in an element. Leading must accommodate what may appear — not
> what does appear.**

"What does appear" is content-dependent and forbidden; "what may appear" is exactly the declared
chain. **The cost is bounded by the author's own choice**: a Latin-only element in a Latin+Thai+CJK
chain gets 11% taller lines, but the author declared that chain — and one who wants Latin metrics
declares a Latin-only chain. (b) imposes CJK/Thai leading **only on someone who declared CJK/Thai
faces.**

*Noted, not built: an explicit `lineHeight` key would let an author tighten leading without narrowing
the chain. Format surface we do not need today; the escape hatch if a real case demands it.*

**Both eliminations were decided by measurement, not argument:**

- **`OS/2` typo metrics — eliminated on constraint 2, in the cleanest way available.** Noto Sans SC
  declares **`USE_TYPO_METRICS = false`** — *the face itself says those are not its line metrics* —
  yet they are plausible numbers, giving 1000/em against a true 1448, **below the face's own ascent
  plus descent**. This is the `/CapHeight 928 against 711` fiction class, caught by a constraint
  written **one story after** the finding that motivated it. A guard catching its class on a
  *different symptom* is the strongest evidence available that the class was real.
- **A size multiple — eliminated on arithmetic.** 1.2 em is below **all three** shipped faces; even
  1.5 em misses Thai's 1511.

---

### D-000.32 — Any constant fitted to the shipped set is a hidden dependency on the shipped set

*(mechanism: binding)* — generalised from the size-multiple elimination.

A constant tuned until it clears the current artifacts **works until the first artifact you did not
tune against, and then fails silently** — no error, just overlapping text. **Derive from the artifact;
never fit to the sample.** The tell is a number that would need re-checking whenever the inputs
change, with nothing in the build that re-checks it.

---

### D-000.33 — A conservation assertion cannot see a degenerate partition

*(mechanism: binding)*

**An additivity or conservation law is satisfied trivially by a degenerate partition.** Story 2.4's
additivity check stayed **green** under a byte-offset mutation, because every boundary past rune 1
collapsed to all-glyphs-or-none and `whole + 0 == whole` holds.

**So testing a conservation law requires also asserting the partition is non-degenerate** — both parts
non-empty — not merely that the parts sum. The fix is to pin the **partition** (boundary indices as
hand-derived literals from `"ณัฐวุฒิ"`'s cluster vector `[0 0 2 3 3 5 5]`), not just the conservation.

Same family as Story 1.1's *"two empty files are byte-identical"*, one abstraction up.

**And the framing, kept verbatim:** *a red-proof that does not redden is a finding, not a formality.*
A stated proof that ran and did not fire is **more** valuable than one that fires — it is the only
kind that tells you the guard was weaker than its own description.

---

### D-000.34 — When a fix lands, check whether any test's discriminating power depended on the defect

*(mechanism: binding)*

**A test built on a bug dies with the bug — silently, because it goes on passing.**

Story 2.4's `TestUnconstrainedVsConstrainedSwitchActuallyToggles` used `"ดอเลาะ"` *because* the engine
proposed a spurious break there. **That break was the P2 defect.** The filter withdrew it, both modes
now return `[]`, and the guard discriminates nothing while remaining green.

**The model answer**, which this story performed: measure which inputs still discriminate (23 of 243
corpus items — but **no single-span one**), then re-point at a measured, **stronger** discriminator
(`"ชัยวัฒน์"`: unconstrained `[3 7]` versus constrained `[]`).

Companion to [[D-000.30]]: that rule says the *red-proof* window closes when a fix lands; this one
says **existing guards can quietly lose their teeth at the same moment.**

---

### D-000.35 — An AC or comment citing an existing capability names the symbol

*(mechanism: binding)* — cheap prophylactic, third instance.

Story 2.4's AC2 cited CJK ranges *"the engine already classifies for face resolution"*. **No such
classification exists** — face resolution is coverage-based and classifies nothing by script.

Third instance of a specification asserting something untrue about existing code, after F3's wrong
artifact ([[D-2.3.4]]) and the `PinAxisLocation` mechanism ([[D-2.2.4]] correction). **Naming the
symbol makes the citation checkable rather than atmospheric** — and a citation that cannot be checked
is one nobody will check.


---

### D-2.4.5 — Story 2.4's canonical-document amendments, quoted verbatim

*(mechanism: binding — the record AC6 and D-2.4.4 require)*

Two canonical documents were amended under [[D-000.6]] in this story, and both are quoted before and
after here so the change is auditable without a diff.

**1. `ARCHITECTURE-SPINE.md`, AD-25 — the Rule gains a third mechanism ([[D-2.1.6]], [[D-2.1.10]]).**

*Before:*

> - **Rule:** two constraints sit **under** whatever the dictionary proposes, and both override it:
>   - **Unknown runs are atomic.** A run of Thai characters the dictionary cannot cover yields
>     **no** interior break opportunities. It overflows visibly under FR44 — clipped, with a
>     located diagnostic — rather than being split at a guess.
>   - **No break inside a Thai character cluster.** A consonant with its vowels and tone marks is
>     indivisible. This is mechanical and needs no dictionary.

*After:*

> - **Rule:** three constraints sit **under** whatever the dictionary proposes, and all override it:
>   - **Unknown runs are atomic.** *(unchanged, verbatim)*
>   - **No break inside a Thai character cluster.** *(unchanged, verbatim)*
>   - **A declared value is never split.** A document may declare, in its `unbreakableValues` list,
>     which data paths hold values that must stay on one line; no break opportunity survives inside
>     a substituted value from a listed path. The engine **never infers** membership. It cannot:
>     Thailand's Surname Act has every family coin a unique surname out of ordinary dictionary
>     words, so a surname and the common words it was built from are character-for-character
>     identical and no dictionary-coverage rule can separate them. The declaration reaches the
>     breaking stage as a **parameter**, never through an import.

**AD-25's `Binds` and `Prevents` lines are byte-unchanged**, as are the two existing mechanisms and
the closing paragraph. Only the Rule's mechanism list moved, which is the scope AC6 permits.

**2. `epics.md:823` — the UAX #14 clause names the outcome, not the mechanism ([[D-2.4.4]]).**

*Before:* `**Then** breaks occur at whitespace per UAX #14`

*After:* `**Then** breaks occur at whitespace, and at script-appropriate opportunities in every other script`

Nothing in the engine claims UAX #14 conformance, and the narrowing is stated by name — no
hyphenation, no break at `-`, no contextual pair rules — in `internal/text`'s package doc, in
`folio-format.md`'s new *Line breaking* section, and in `folio-go/README.md`.

---

### D-2.4.6 — The leading rule is the maximum over the declared chain, and `folio-format.md` says why

*(mechanism: binding)* — the ruling [[D-2.4.2]]'s measurement produced, recorded with its numbers.

D-2.4.2 fixed the criterion before the data and asked for the candidates to be measured against it.
Exactly one **metric** survived; the measurement then exposed a second question the candidates had
never distinguished — *which face of the chain* — and the answer is the **maximum over the declared
chain**.

**The measurements, read from the tables (`TableData` + `ParseHhea`/`ParseOS2`), declining the
substituting accessors.** Scaled to the 1000-unit em, against the faces as committed:

| face | `hhea` A−D+gap | `OS/2` typo A−D | `fsSelection` | USE_TYPO_METRICS |
|---|---|---|---|---|
| Noto Sans | **1362** | 1362 | 0x00c0 | true |
| Noto Sans Thai | **1511** | 1511 | 0x00c0 | true |
| Noto Sans SC | **1448** | **1000** | 0x0040 | **false** |

**`OS/2` typo metrics — eliminated on constraint 2.** Noto Sans SC declares `USE_TYPO_METRICS = false`:
the face itself says those values are not its line metrics, and using them gives **1000 against a true
1448**, *below the face's own ascent plus descent*. A plausible number that is wrong — the
`/CapHeight 928 against 711` class exactly. The vendor also exposes no accessor for them at all.

**A size multiple — eliminated on measurement.** 1.2 em is below all three faces; even 1.5 em is below
Noto Sans Thai's 1511. Clearing the shipped set needs ≥ 1.511 em, a constant fitted to today's faces.

**First-face — eliminated on measurement, and this is why the second question mattered.** The shipped
chain's first face gives 1362 where Noto Sans Thai needs 1511: at 16 pt, 21.79 pt of leading for text
needing 24.18. Thai below-vowels collide with the next line's above-vowels **in the default chain, on
the script this epic exists to support**.

**Constraint 1 correctly declined to separate first-face from max-over-chain**, and that is the
constraint working rather than failing: it forbids *content-dependent* leading, and neither candidate
is content-dependent. **(b) depends on the DECLARED chain, not on the glyphs that appear**, so adding a
CJK character reflows nothing and AD-24 holds exactly as under (a).

**The sentence that makes (b) positively correct rather than merely safe**, now in `folio-format.md`:

> **A chain declares what may appear in an element. Leading must accommodate what may appear — not
> what does appear.**

**The cost is bounded by the author's own choice:** a Latin-only element in a Latin+Thai+CJK chain gets
~11% taller lines, but the author declared that chain, and an author who wants Latin metrics declares a
Latin-only chain. **No element pays for a face its own chain does not name.**

**No `lineHeight` key.** Noted as the escape hatch if a real case ever demands it; it is format surface
not needed today, and line height is derived rather than authored.

**Where the ambiguity came from, for the record:** DN-2 framed the candidates as *metrics* and never
distinguished first-face from max-over-chain, so its recommendation of the first face was never
weighed against the alternative. The dev agent declined to overturn it by taste and escalated with the
measurements, which is the behaviour [[D-000.31]] describes.

---

### D-2.4.7 — Story 2.4 finisher corrections: two misquotes, one stale citation, and a deferred terminology split

*Filed 2026-08-24 by the story finisher, closing Story 2.4's review findings 4 and 5. Appended, not
edited into [[D-2.4.2]] or [[D-2.4.6]]: this log's header is explicit that entries are append-only
and "a reversal is appended, never a rewrite". The rulings themselves are unchanged and remain
correct — only their block quotes were inaccurate.*

**1. The binding formulation is quoted `chain`; the canonical document says `stack`.**
[[D-2.4.2]] (resolved) and [[D-2.4.6]] both block-quote the sentence as *"**A chain declares** what
may appear in an element. Leading must accommodate what may appear — not what does appear."*
`_bmad-output/specs/spec-folio/folio-format.md:263` — the document both rulings point at, and the
canonical side — reads:

> **A stack declares what may appear in an element. Leading must accommodate what may appear — not
> what does appear.**

That is the verbatim text. `folio-format.md` uses "stack" consistently through the section
(*declared font stack*, *a Latin-only stack*, *no element pays for a face its own stack does not
name*), so the document is internally consistent and the two rulings are the inaccurate side.
[[D-2.4.6]]'s own heading claims *"`folio-format.md` says why"*, which makes the misquote worse than
cosmetic: [[D-2.4.5]] exists precisely so a reader can grep a canonical sentence and land on it.
**The substance of both rulings is unaffected** — the leading rule is still the maximum over the
declared stack of `hhea` ascent − descent + lineGap, and the argument for it is unchanged.

**2. The UAX #14 clause is `epics.md:827`, not `:823`.** [[D-2.4.4]]'s heading and [[D-2.4.5]]'s
amendment list both cite line 823. Verified at `0266a86` and at head: the clause *"**Then** breaks
occur at whitespace per UAX #14"* is at line **827**, and the amendment is exactly one line
(`@@ -824,7 +824,7 @@`). The amendment itself is correct; only the citation was off by four.
D-000.14/D-000.26 ask locations to be precise so a reader can land on them, which is the whole reason
this is worth an entry.

**3. Deferred, and named so it is not lost: the code says `chain`, the spec says `stack`.**
`wrap.go`'s `lineAdvance(chain …)`, `fontChain`, and their tests use `chain`; `folio-format.md` uses
`stack` throughout. Both are internally consistent and neither is wrong, but a reader moving between
them has to translate. **Follow-up:** pick one term and align both sides. It was deferred out of
Story 2.4's finish commit deliberately — that commit's scope is `_test.go` files and documents, and
renaming identifiers across `wrap.go` and `render.go` would touch production code the four-target
matrix has already certified for this story. It belongs in a story that can re-run the matrix behind
the rename.

**What this entry does NOT do.** It adds no gate obligation. The Epic 2 gate still owes exactly
three things — the outstanding four-target matrix legs, D-2.3.5's Thai *reading* sign-off, and
D-2.4.3's Thai *break* sign-off — and `TestStory23aAddedNoThirdEpic2GateObligation` still catches a
fourth.

---

### D-000.36 — A proposed remedy for a vacuity is itself subject to vacuity; measure it before adopting it

*(mechanism: binding)* — second instance of the rule the lead stated at the Epic 1 gate: **a mitigation
for a measurement hazard must itself be measured before it is relied on.** Recorded again because the
first instance was about shell instruments and this one is about test assertions, and neither reader
would recognise the other.

**The instance.** Story 2.4's review found `TestMeasureIsAdditiveAcrossASplit` stayed green under the
byte-offset mutation, and proposed the obvious remedy — require `left > 0 && right > 0`, the
non-degenerate-partition precondition [[D-000.33]] calls for. **Implemented verbatim, the test stayed
green under the mutation, witnessed on 4 of 4 subjects.**

**Two measured reasons, both worth knowing:**

1. **Additivity is preserved by *any* monotone boundary function.** `left + consumed + right == whole`
   holds whenever the boundary moves consistently — so the assertion can only detect a **saturating**
   boundary function, never a merely *wrong* one. A non-empty precondition rules out the degenerate
   case and leaves the entire monotone-but-wrong space untouched.
2. **Every split in the subject table fell on a face-segment boundary**, where segment-local rune
   indices and byte offsets **both read 0** and are therefore indistinguishable.
   `"ณัฐวุฒิ เกิด กรุงเทพ"` segments as `[0,7)[7,8)[8,12)[12,13)[13,20)` and breaks at 7 and 12 —
   **zero interior splits.** The subject could not express the defect regardless of the assertion.

**The complete fix needed three parts, not one**: the non-empty guard, **plus** a whitespace-free Thai
subject that breaks strictly *inside* one segment, **plus** a precondition that fatals if interior
splits are ever lost from the table. Under the mutation it now fails naming the subject.

**The general form** *(mechanism: binding)*:

> **A remedy proposed for a vacuity inherits the vacuity's own hazard. Run the mutation against the
> remedy before adopting it — a fix that "obviously" closes a hole is exactly the kind nobody
> re-measures.**

And the corollary that made this one visible: **the subject matters as much as the assertion.** A
correct assertion over inputs that cannot express the defect is still vacuous — [[D-2.3.2]]'s corpus
obligation (`น้ำ` *and* standalone `า` in one document) is the same requirement stated for fixtures.

**Also closed in the same sweep, as an enumeration rather than a sample** ([[D-2.2.4]] correction's
precedent): **7 conservation sites** across all 31 story-touched files — 5 fixed, 2 already adequate,
plus 8 production accumulators swept and excluded with reason (summations with no assertion attached
are not in the class).

---

### D-000.26 (refined) — A sign-off binds to the artifact expressing the property judged, not uniformly to the PDF hash

*(mechanism: binding)* — refines [[D-2.3.5]]'s condition 2, which required a sign-off to name **the
hash**.

**That was right for a judgment about rendered appearance and wrong for a judgment about break
placement.** [[D-2.4.3]] split the Thai obligation into two human judgments; binding both to the PDF
digest binds one of them to an artifact that does not express what it judges.

| sign-off | judges | binds to |
|---|---|---|
| Thai **reading** | does the rendered Thai read correctly | **the rendered image / PDF hash** |
| Thai **break** | do the break points fall correctly | **the break-opportunity vector (data)** |

**The concrete payoff.** A uniform baseline shift (DW-15) changes the bytes and the rendered image, so
it correctly invalidates the **reading** sign-off — but the break vector is untouched, so the
**break** sign-off **carries over.** Under the uniform-hash rule both would have been invalidated, and
a scarce human would have been asked to re-examine something that had not changed.

**Generalised:** this is [[D-000.21]] — *assert on the artifact that carries the property* — applied
to **human judgments** rather than machine assertions. Over-binding a sign-off is not merely
wasteful; it trains the signer that re-approval is routine, which is exactly the dilution
[[D-2.4.3]] declined to accept.

**Sequencing consequence** *(mechanism: binding)*: **a change that will invalidate a pending human
sign-off should land before that sign-off is recorded**, or the human is asked twice. When such a
request is already in flight, weigh re-requesting against holding the change — **the scarce human is
the constraint, not the code.**

---

### D-000.16 (ranks corrected) — `text` precedes `fontset`; the table is now measured, not illustrative

*(mechanism: binding)*

**Framing first, because it matters for how the next creator treats an illustrative parameter.** This
is **not a falsified ruling — it is the validation the ruling required.** D-000.16 said explicitly:
*"binding for the rule; **illustrative for the ranks**"*, and instructed the story to **validate**
them. The ranks were marked illustrative precisely because the real import graph had not been
measured. Story 2.5's creator measured it and found one wrong. **The mechanism worked as designed.**

**Measured:** `fontset -> [geom, text]` (`fontset.go:25`, `:245`, `:445` — created by Story 2.3a's own
vendor containment), and **`text` imports nothing internal.** No cycle; `text` receives faces as
**values**, which is the pipeline discipline (*"the signal rides on the value, never through an
import"*) holding unaided.

**The corrected order is right for a reason, not merely for compatibility.** Subsetting needs the
glyph set; the glyph set comes from shaping — AD-8: *"one subset per font per document over the union
of glyphs used."* **You cannot know the union without shaping first.** So **shape → collect → subset**
is the true pipeline order and `text` genuinely precedes `fontset`. The original ordering assumed
`text` depended on `fontset` for metrics; measurement shows metrics arrive as **values**.

**Swap ranks 5 and 6.** Both arrows the guard exists to enforce still hold: `layout`(7) ↛ `pdf`(8),
`expr`(3) ↛ `bind`(4). **The table is ratified as measured from here**, no longer illustrative.

---

### D-2.5.1 — The gate-obligation guard asserts set equality against a declared list, never a count in a name

*(mechanism: binding)*

The fourth Epic 2 gate obligation is **sanctioned**: a new golden that is not matrix-registered is
exactly the drift Story 2.3 closed.

**But `TestStory23aAddedNoThirdEpic2GateObligation` encodes a count in its name, and that name rots on
every future obligation.** Replace it with an assertion that the **registered obligation set equals an
explicit declared list.** Adding one then becomes a **one-line diff to that list** — visible and
reviewable — instead of a rename.

Same *derive-from-a-declarative-spec* move that fixed the per-face assertion set in [[D-000.23]]'s
consequent obligation. It retires the counting problem **permanently** rather than deferring it one
story.

---

### D-000.37 — A tripwire's failure message is executable by a human; a stale remedy is worse than a stale comment

*(mechanism: binding)*

**Because it will be followed.** `TestProvisionalBandOriginIsPinned` fires correctly on
`mkdir internal/layout` — a **true** positive — but **its message's site-(2) instruction is half
stale, and following it literally would silently move goldens** while the developer believed they were
obeying the guard.

We already knew a **false positive invites a workaround**. This is its sibling: a **true positive with
an outdated remedy invites a wrong fix.**

**The remedy text is maintained as code, not as prose**: any story that changes the mechanism a
tripwire points at **updates the message in the same commit.** Story 2.5 assigns the deletion to an AC
rather than leaving the stale instruction to be obeyed.

---

### D-000.4 (override criterion) — An override needs a new source of cross-target divergence, not a new golden

*(mechanism: binding)* — Story 2.5's per-story override **declined**.

**An override is warranted when a story introduces a new *source of cross-target divergence*** — float
arithmetic, a vendor call, a compressor, a new dependency — **not merely because it records a new
golden.** Story 2.4 qualified because line breaking ran vendor shaping across three scripts, adjacent
to the float path just measured. **Story 2.5 is integer band arithmetic on `geom.Length`**: no floats,
no new vendor surface, nothing that could diverge between arm64 and wasm that has not already been
proved identical.

**If "records a golden" sufficed, nearly every story would qualify**, and D-000.4's trade — which the
owner made deliberately on wall-clock — would erode to nothing. The epic gate is three stories away
and the new golden is verified there.

**The creator stating the counter-argument so it could be refused is the right behaviour**: better to
refuse a well-made case than never to see one.

---

### D-000.38 — Two guards sharing a parser are one guard wearing two names

*(mechanism: binding)*

**The instance.** A new matrix document written as a **gofmt-clean single-line struct literal**
(`{label: "x", slug: "y", …},`) is invisible to **both** `TestEpic2GateObligationsMatchTheDeclaredSet`
and `TestMatrixDocumentSlugsAreRegisteredInCI`. Constructed by Story 2.5's reviewer: it compiles under
`-tags matrix`, and **both tests report green.** That is *"a matrix leg nobody compares, reported as
green"* — the exact drift Story 2.3 closed — **reopened by formatting.**

**The root cause is the structure, not the regex.** Both guards share one `^\s*slug:` pattern. Two
guards that consume the same parse are **not independent**: a single parsing assumption defeats both
simultaneously, while the pair's existence advertises redundancy that does not exist. **Fixing only
the regex leaves that structure intact**, and the next formatting variant defeats them again.

**The rule.** Independence between guards must be **in their mechanism**, not merely in their names or
their assertions. Two checks over the same extraction are one check. When redundancy is the point —
and here it was, since these two exist to cover each other — **the second guard must reach its facts
by a different route**: parse the Go source with `go/ast` (already used in `lint`, which type-checks),
or derive both lists from a **single declarative source that cannot be expressed two ways**, so there
is no second spelling to miss.

**And the red-proof must use the shape that escaped** — here the single-line literal — because per
[[D-000.36]] a remedy proposed for a vacuity inherits the vacuity's own hazard.

**The family this belongs to.** [[D-000.19]] said an unexplained delta between two independent
computations is where a calibration hides. This is its dual: **when two computations *agree* because
they are secretly one computation, the agreement is worth nothing** — and it reads as corroboration,
which is worse than reading as silence.

---

### D-000.39 — A surviving mutation with a byte-identical artifact is evidence of equivalence, not of a gap

*(mechanism: binding)* — the disposition Story 2.5 twice reached and the finisher flagged as the
strongest unrecorded rule.

**The situation.** You mutate the code to prove a guard can fail. The guard stays green. The reflex is
to treat this as a hole and strengthen the assertion until it fires.

**That reflex is wrong when the produced artifact is byte-identical**, because then the mutation was a
**semantic no-op** and there is nothing to detect. Strengthening the assertion until such a mutation
reddens does not close a gap — it manufactures a guard against a difference that does not exist, and
that guard will later fire on a legitimate refactor.

**The instance.** Swapping `marginTop`/`marginBottom` *inside* `ContentHeight`'s expression left every
assertion green **and the digest byte-identical**. Both margins enter the partition only as a **sum**,
and `a − b − c == a − c − b` exactly. **No assertion over derived values can ever see it.** The
reviewer confirmed completeness by probing the *adjacent* mutation that should be visible —
`ComposePage` carrying `MarginBottom` as `MarginTop` — which reddens three ways.

**The obligation, which is what makes this honest** *(binding)*: a surviving mutation may be declared
an equivalence only by **showing the artifact is byte-identical** and by **finding the nearest
mutation that IS observable and demonstrating it fires.** Absent that second half, "it's equivalent"
is indistinguishable from "our guards are blind here" — which is the claim it is being used to deny.

**Where strengthening remains available**: at the **inputs**, before the values are combined. Story
2.5 pins each margin individually there, which is the only place the swap is expressible.

**Distinguish from [[D-000.36]]**: there, the remedy stayed green because the *subject* could not
express the defect — a real gap, fixed by changing the subject. Here the *mutation* cannot produce a
difference at all. **The two look identical from the test's point of view, and the artifact's bytes
are what separate them.**

---

### D-000.40 — A mutation must print its own preconditions; an exit code cannot tell you it applied

*(mechanism: binding)* — ninth instrument failure of the run, and the second inside a mutation harness.

**The instance.** A red-proof used `sed 's/\bx\b/y/'`. **BSD sed does not support `\b`**, so the
substitution **never applied** — while `sed` exited 0 and the subsequent test run reported exactly the
result a successful mutation would have produced *if the guard were sound*. A second attempt injected
**malformed Go**, which proved the parser path rather than the identifier path it was aimed at.

**Both were caught by printing the mutation's preconditions rather than by checking its exit code**,
and both are recorded in the story as invalid attempts rather than silently replaced.

**The rule.** A mutation harness must **assert that the mutation took effect** — the file differs, the
intended token actually changed, the code still compiles for the reason under test — **before** the
result of the test run means anything. **A green suite after a mutation that never applied is
indistinguishable from a green suite after a mutation the guard failed to catch.**

Same family as [[D-000.9]]: the tool's *"I changed nothing"* and its *"I changed something harmless"*
must not produce the same observation. And it joins the run's standing list — `rtk proxy` before
redirect, both operands proven non-empty before comparison, N-of-N coverage witnesses, and
[[D-000.26]]'s cite-the-subject — as a **portable-tooling** hazard: `\b`, `-i`, `sort`, and `grep -P`
all differ on BSD, and this project develops on darwin and ships to linux.

**Recording invalid attempts as invalid, rather than replacing them quietly, is the behaviour that
made both visible.**

---

### D-000.41 — Request a scarce human sign-off only when no scheduled work is known to move the artifact it binds to

*(mechanism: binding)*

**The hole this closes.** Story 2.5a was scheduled before 2.6 so DW-15's baseline fix lands **before**
the Thai reading sign-off is requested — otherwise the owner examines the same Thai twice. But that
only works **if 2.6 does not move the same fixture again.** The orchestrator's own note said 2.6
*"will likely re-record those goldens anyway"* — in which case a sign-off recorded after 2.5a is
invalidated by 2.6, **and the scarce human is spent twice for exactly the reason the story was split
to avoid.**

**The rule.** A human sign-off is requested only when **no scheduled work is known to move the
artifact it binds to.** Otherwise you are buying attributability with someone else's attention.

**Operationally**: the question is answerable early and cheaply — it is the same *"which fixtures can
move"* measurement Story 2.4's creator already performed, and [[D-000.36]]'s population finding in 2.5
shows how much that measurement pays. **2.6's creator measures whether its changes move any
signed-off-pending golden, before any request is sent.**

- 2.6 does **not** move `shaped-text` → request the reading sign-off after 2.5a, as planned.
- 2.6 **does** → hold the request until after 2.6. 2.5a's value is then **attributability alone** —
  still sufficient reason to split, just not the second one.

---

### D-000.39 (sharpened) — the neighbouring observable mutation must be at the SAME SITE

*(mechanism: binding)*

The obligation as first written — *"find the nearest mutation that IS observable and demonstrate it
fires"* — is **satisfiable by a mutation somewhere convenient**, which is how a good rule becomes a
formality. Demonstrating the guard fires on some other line proves **the guard exists**; it does not
prove the guard has **resolution where you just declared an equivalence.**

> **Byte-identity proves the mutation was inert. A neighbouring observable mutation at the same site
> proves the guard is not blind there. You need both, because each alone is consistent with the
> failure the other rules out.**

And the discriminator between this rule and [[D-000.36]] is worth stating plainly: **when a mutation
survives, the artifact's bytes tell you which failure you are looking at.** Bytes identical → the
mutation was inert. Bytes would differ but the subject cannot express it → **the population is
wrong.** Same green test, **opposite remedies**, and only the artifact separates them.

---

### D-000.40 (sharpened) — the mutation asserts it applied by observing the artifact, not by its exit code

*(mechanism: binding)*

> **The mutation must assert it applied by observing the artifact — a non-empty diff, or a reported
> count of occurrences changed — before any test result is interpreted. An exit code is the *tool's*
> claim about itself; the diff is the artifact.**

That is [[D-000.21]] arriving in the mutation harness.

**And practically**: prefer a mutator that **reports what it changed** over one that changes and exits.
This run already concluded Python is the reliable instrument for measurement; the same applies to
mutation, for the same reason.

**Portable-tooling hazard list, for the gate note** — every one is a silent behaviour difference on the
path between where this project is developed (darwin) and where it ships (linux): **`\b`, `-i`,
`sort`, `grep -P`.** `sed 's/\bx\b/y/'` on BSD is the purest instance found: **the tool succeeded at
doing nothing, and the test then reported exactly what a sound guard would have reported.** Both halves
silent; the run indistinguishable from success.

---

### D-000.42 — "Redundant" is a third guard category, beside proven and forward

*(mechanism: binding)*

The vocabulary had two states: **proven** (a red-proof fires) and **forward guard with no available
red-proof** ([[D-000.24]]). Story 2.5's finisher found a third and labelled it correctly: neutering its
count check **reddens nothing**, because the check *is* exercised — but only **through another
assertion**.

| category | red-proof | independent teeth |
|---|---|---|
| **proven** | fires | yes |
| **forward guard** (D-000.24) | not constructible | unknown |
| **redundant** (this) | does not fire | **no — subsumed by another assertion** |

It is **not unproven; it is redundant.** Labelling it so is honest, and — more importantly — **stops it
being counted as coverage**, which is how a guard set comes to look larger than it is. Keep such
checks if they are cheap belt-and-braces; never count them.

---

### DW-16 (direction ruled) — `pagemodel` carries glyph ids; CID allocation stays entirely in `internal/pdf`

*(mechanism: binding for the direction; illustrative for the mechanism)*

**The naming is the tell.** A **CID** is a PDF/CIDFont concept, and AD-5 is *"The page model knows
nothing about PDF."* A **glyph id** is a font concept **any** renderer can resolve against the face. A
PNG or SVG renderer can rasterise from (face, GID); it can do **nothing** with a CID whose
disambiguating table lives on the PDF side. The extras branch is pure Identity-H CID-space
management — an encoding concern that should never have reached a renderer-agnostic model.

**Preferred over tagging the field**, because relocating removes the two-meanings problem **by
construction**: a discriminator tag would make the ambiguity *legible* while keeping a PDF concept in
the page model.

**Not urgent** — nothing regressed, and AD-5's forcing function (a second renderer) is post-MVP. **The
direction is ruled now so nothing further is built against the current shape.** If the relocation
changes allocation order **it moves bytes**, in which case it takes its own commit under the same
attributability discipline applied to 2.5a. **Owner**: whichever comes first of the next story touching
`pagemodel`'s glyph representation, or Epic 5's first non-PDF renderer.

---

### D-2.4.2 (amended) — Leading maximises each axis independently; the constraint is the worst adjacent PAIR

*(mechanism: binding)* — amends [[D-2.4.2]] the same day it was ruled. **The wrong quantity was
maximised.**

**The defect.** Between two baselines the space must hold **line N's descenders plus line N+1's
ascenders**. The original rule took `max(A − D + gap)` over the chain — **the worst single face** —
when the constraint is **the worst adjacent pair**:

```
worst pair = max(descent) + max(ascent) = 450 (Thai) + 1160 (SC) = 1610
ruled      = max(A − D + gap)           = max(1362, 1511, 1448)  = 1511
shortfall                                                          =   99
```

**Amended, and forced by D-2.4.2's own principle rather than a new one:**

> **Leading = max(ascent) + max(|descent|) + max(lineGap), each maximised independently over the
> declared chain.**

*A chain declares what may appear; accommodate what may appear* — **applied per axis**, because ascent
and descent are independent axes and the worst case takes the worst of each. The original form
**implicitly assumed one face supplies both**, which is false on the shipped set: **Thai wins descent,
SC wins ascent.** All three D-2.4.2 constraints still hold — content-independent, `hhea`-only,
statable in `folio-format.md`.

**The corrected vertical model is one rule with three maxima:**

| span | value |
|---|---|
| top → first baseline | `max(A)` |
| baseline → baseline | `max(A) + max(D) + max(gap)` |
| last baseline → bottom | `max(D)` |

**DN-1 falls out of the same correction: the first baseline uses `max(ascent)` = 1160.** It is no
longer a separate judgment. That the two candidates **coincide only on the shipped set** is
[[D-000.32]]'s fitted-to-the-sample hazard **one ruling away from where it was named**, and would have
gone unnoticed until a chain whose advance-winner is not its tallest face.

---

### D-2.5a.1 — DN-3 folds into 2.5a: two symptoms of one defect, not two subjects

*(mechanism: binding)*

**This widens a scoped story, which was twice refused before. The difference is the granularity of the
cause:**

- **2.2's static switch vs 2.2's Thin fix** — **two different subjects** stacked in one movement.
  Un-attributable. Refused.
- **DW-15 and DN-3** — **two symptoms of one defect**: the vertical placement model uses **proxies**
  (`fontSize`; worst-single-face) instead of accommodating the declared chain. Correcting one and not
  the other ships an **internally inconsistent model** — the same objection raised against
  `pdfY = flipY(..., run.FontSize)` in the first place.

**The economics are decisive**: both move the same five goldens and both are upstream of the reading
sign-off. Split, we pay two movements, two matrix passes, and — if the sign-off falls between them —
**two human passes**. Together: one cause, one movement, one request.

**State the scope difference honestly**: DW-15 shifts baselines **uniformly**; DN-3 changes **line
height**, so **fewer lines fit a band and wrapping or overflow may shift.** Larger blast radius; check
the fixtures for it explicitly rather than assuming. **Nothing else folds in** — the vertical model is
the cause; multi-page headers are not.

---

### D-000.43 — When the concern is sequencing rather than divergence, fix the sequence, not the cadence

*(mechanism: binding)*

Story 2.5a declined the heavy-test override honestly against [[D-000.4]]'s criterion — same integer
`ScaleRound`, same vendor entry points, no float, compressor or dependency — and then **correctly
diagnosed its own counter-argument as being about sequencing, not divergence.**

Its substitute (AC13): **the reading sign-off is not requested until the gate's four matrix legs agree
on the new digest.** That composes with [[D-000.41]] rather than duplicating it — D-000.41 says *do not
request while scheduled work may move the artifact*; AC13 adds *do not request until the legs agree on
the digest.* **Together they protect the scarce human from both futures.**

**And it is strictly better than the override**: an override buys a matrix run in-story at Docker cost;
AC13 buys the same protection **for the thing that actually matters** — nobody is asked to judge bytes
that might not reproduce — **at zero cost.**

---

### D-000.44 — A re-recording is a recording: D-000.22's step is owed every time, not only the first

*(mechanism: binding)*

**Measured**: of the five goldens 2.5a re-records, **only one** has a semantic check that can see the
defect. **Three have none at all** — so [[D-000.22]] was recorded as discharged for fixtures that never
had the step.

D-000.22's timing clause is *"at first recording"*. **A re-recording is a recording** — it is precisely
when the step is owed, and the story touching all five is the one that owes it. **2.5a adds semantic
acceptance for every fixture it re-records.**

**And the one that exists but is blind is the purest [[D-000.21]] instance yet**:
`TestWrappedTextSemanticAcceptance` calls `lineAdvance`, compares it to a literal, then **computes the
observed baselines and discards them.** It asserts on the **input** and throws away the **artifact** —
while being named for semantic acceptance. Repair it to assert the computed baselines.

---

### D-000.45 — Assert the computed value from a declarative table, never a direction

*(mechanism: binding)* — second instance; the first would have shipped a vacuously-false guard.

**Roboto's ascent is 928 — below 1000** — so `max(ascent) − fontSize` is **negative** and its baseline
moves **up**, while the three shipped Noto faces move **down**. **Any guard phrased *"the baseline sits
lower than the font size implies"* is false on a real subject in the repository.**

Same correction as `epics.md:751`'s `YOffset != 0`, which was **vacuously false** across every shipped
face. **Twice now a directional phrasing would have shipped a guard that is false on real data** — so:
**assert the computed value from a declarative per-subject table**, never a direction, never an
inequality standing in for a value.

---

### D-000.46 — Dead code that documents the system incorrectly is worse than dead code

*(mechanism: binding)*

`render.go:463`'s `splitByFace` has **zero call sites in production or test**, while **six comments
describe it as the live placement path** — and it computes placement on the **old** model.

**Every reader who greps for placement finds it first.** It is not merely unused; it is a **false map**
of the system, and it survives precisely because nothing fails when it is wrong. **Delete it or wire
it — but never leave comments asserting a path that does not execute.**

Companion to [[D-000.37]]: a stale *remedy* is executed by a human; a stale *description* is reasoned
from by a human. Both are followed.

---

### D-000.47 — Enumerate a multi-site invariant once, declaratively, and have the check read the list

*(mechanism: binding)*

A golden's digest lives at **four sites**, discovered by measurement: the PDF, `expected.json`,
`byte_neutrality_test.go`'s second literal (3 of its 5 entries move), and **two fixture READMEs**.

**Four sites found by measurement are four sites a future re-record can miss.** Declare the list once
and have the check read it — the same *derive-from-a-declarative-spec* move that fixed the per-face
assertion set ([[D-000.23]]) and the gate-obligation count ([[D-2.5.1]]).

And per [[D-000.37]], update `byte_neutrality_test.go`'s **"Do not update this literal"** message **in
the same commit** — this story must update it, so a true guard now carries a remedy that forbids the
correct action.

---

### D-000.34 (extended) — A fix can silently destroy a NEGATIVE control, not only a positive one

*(mechanism: binding)* — unanticipated instance, found by Story 2.5a's developer applying D-000.34 to
its own change.

[[D-000.34]] says a test built on a bug **dies with the bug**, silently, because it goes on passing.
Story 2.5a found the **mirror image**, which is harder to see:

`TestLineAdvanceIsTheMaxOverTheDeclaredChain`'s **negative control** was the "Thai first" row, chosen
precisely because it was *"the one order where (a) and (b) agree"*. **The amendment destroys that
property** — 1610 now requires two faces — so the row **silently became a discriminating case.** The
table would have been left with **its negative control gone and nobody told**, still green, still
looking like a controlled experiment.

**A positive control that dies stops catching regressions. A negative control that dies stops proving
the test can distinguish anything at all** — and its death is invisible, because a control passing is
what a control is supposed to do.

**The obligation** *(binding)*: when a change alters the property a control was selected for, the
control must be **re-selected and the reason re-stated** — not merely re-run. Story 2.5a's rebuild is
the model: **two rival hypotheses**, controls moved to the **single-face** rows (where the two
formulas still coincide), and **three vacuity guards**.

**And check the whole touched set, not the one you noticed.** A control's justification is usually
written once, in prose, at the moment it was chosen — so nothing in the build re-evaluates whether it
still holds.

---

### D-000.48 — A correction sweep can introduce a fresh instance of the class it is correcting

*(mechanism: binding)*

**The instance.** Story 2.5a's AC8 comment sweep existed to fix stale citations. It **added**
`render.go:359`'s reference to `TestVerticalModelErrorPathsThroughThePublicEntryPoint` — **a test that
exists nowhere** (the real name is `...AreUnreachableThroughRender`). The sweep correcting stale
citations introduced a stale citation.

**Why this is not merely careless.** A sweep is written in one pass, at speed, over many sites, by
someone holding the *intent* of each comment rather than checking each symbol. **The very conditions
that make a sweep worth doing — breadth and momentum — are the ones that produce unverified
citations.** So the hazard is structural, not attributable to inattention.

**And an unresolvable citation is worse than none**, because it **reads as checkable**: a reader who
sees a named test assumes someone could run it, and stops looking.

**The obligation** *(binding)*: **a sweep that writes citations must verify every citation it writes
resolves to a real declaration**, and report the swept list with its count — mechanically, not by
reading. This is [[D-000.35]] (*name the symbol*) plus the [[D-2.2.4]]-correction precedent
(*enumerate, never sample*) applied to the sweep's **output** rather than its input.

**Companion finding, same layer.** The same story's D-000.45 guard was **correct** while its success
message stated both directions **backwards** — *"2 sit lower … 16 sit higher"* — contradicting its own
`Fatalf`. That is [[D-000.21]] **in the reporting layer**: the message was derived from the author's
expectation rather than from the measured values. **Derive the narration from the same computed table
the assertion uses**, so the two cannot disagree. A correct guard with an inverted message will be
believed over the guard, because the message is what a human reads.

**And a third, mechanically detectable**: `epic-2-boundary-gate.md` carried a **65-character** sha256
— one character too many — as *the value a human is told the deferred matrix legs will compare*. A
digest of the wrong length is checkable by shape alone; anything that records digests for human action
should assert their length.


---

### D-2.5a.2 — Discharging D-000.48's sweep obligation found a second instance inside the same sweep

*(mechanism: binding · filed by Story 2.5a's finisher · appends to, and does not amend, [[D-000.48]])*

**What was asked.** [[D-000.48]] obliges *"a sweep that writes citations must verify every citation it
writes resolves to a real declaration, and report the swept list with its count — mechanically, not
by reading."* Story 2.5a's finisher discharged it.

**What the discharge found.** The single instance D-000.48 was written about was **not the only one**.
The same AC8 sweep also wrote `present == 0` and `maxUnits <= 0` into `vertical_model_test.go`'s doc
comment and its success log as though they named symbols. `verticalModel`'s two error paths are
spelled `len(metrics) == 0` and `units <= 0`; **there is no `maxUnits` anywhere in the module**. The
names came from AC4's prose, where they are conditions rather than symbols, and were carried into
code as if they were symbols.

**Why it survived the review.** The reviewer read those two spellings inside a passage whose
*judgment* it had independently verified and concurred with — declining the [[D-000.24]] forward-guard
label. **A citation embedded in a conclusion the reader has just endorsed inherits that endorsement**,
which is precisely the reading that a mechanical sweep does not perform.

**The obligation** *(binding, sharpening D-000.48)*: the sweep must enumerate **symbol** citations,
not only test names, and it must run over the **whole added-line population** — a citation is not
exempt because the sentence containing it was found sound.

**Reported, as D-000.48 requires**: 114 distinct citations swept across 1 799 added Go lines in 12
files and 2 276 added document lines in 8 files, checked against 368 parsed `Test…` declarations.
**109 resolve. 5 do not**: one is D-000.48's own instance (fixed), two are texts *quoting* that
instance in order to condemn it ([[D-000.48]] itself, and the story's QA section) — one is
`TestStory23aMovedNoGoldenDigest`, a deliberately deleted symbol cited only in the past tense, and
one is this entry's own finding (fixed). A citation of something deliberately removed is a **record**,
not a pointer, and must survive.

### D-2.5a.3 — A finding's own suggested resolution is not authority over a re-measurement

*(mechanism: binding · filed by Story 2.5a's finisher)*

**The instance.** Story 2.5a's review raised a Nit that correction **C2** enumerated the `splitByFace`
references incompletely, and its Suggested Resolution was to *"align C2's enumeration with DN-4's"*,
DN-4 being described as *"the right one and matches the code"*. Re-measured rather than applied:
`git grep -c splitByFace 17f5f7a -- folio-go` reports 11 matching lines across four files, one of
which is the declaration's own `func` keyword. **There are ten references, not nine.** DN-4 counts
`render.go`'s six as *"six production comments"* and silently drops
`internal/fontset/fontset.go:576`. Both C2 and DN-4 were wrong, in opposite directions, and the
resolution as written would have propagated the second error while fixing the first.

**Why this is structural.** A reviewer proposing a fix names a **known-good source** to copy from,
because that is cheaper than re-deriving. But the source's authority is inherited from the reviewer's
attention, not established by it — and the whole reason the finding exists is that a count in this
document was wrong. [[D-000.18]] already says *recompute, never transcribe*; this is that rule applied
to the **remedy** rather than to the measurement.

**The obligation** *(binding)*: when a finding's resolution says *"make X agree with Y"*, **measure Y**
before copying it. Report the measurement and its counting rule, and if Y is also wrong, correct both
and say so.

**Related, same story, same shape.** The Delivery Log's gate table recorded *"compressor imports: 1,
`internal/template/image.go`'s pre-existing `compress/zlib`"*. That file imports **only `fmt`**; its
sole mention of `compress/zlib` is a comment stating that PNG bytes are carried through **without**
it, which is [[D-1.8.1]]'s entire point. The true count under `folio-go/` outside `testdata/` is
**zero**. Nobody flagged it, because a row asserting a *known, accepted, pre-existing* exception reads
as already adjudicated. **An exception recorded as pre-existing is still a measurement**, and left
standing this one would have read as carried risk **R4** being open when the module's design closes
it.
---

### D-000.49 — A record that overstates a risk is a defect, not caution

*(mechanism: binding)* — found by Story 2.5a's finisher, in a document written for the Epic 2 gate.

**The instance.** `epic-2-boundary-gate.md` recorded *"compressor imports: **1**, `internal/template/image.go`'s `compress/zlib`"*. **That file imports only `fmt`.** Its single mention of a compressor is a **comment** stating that PNG bytes pass through **without** one. The true count is **0**.

**Why this needed fixing rather than shrugging at.** Left standing it reads as **carried risk R4 being
open** — a gate document telling a reader that a closed risk is live. The usual failure this project
hunts is a record that **understates** a problem; this is its mirror, and it is not the safe direction:

- It **spends attention** at the gate on a risk that does not exist.
- It **erodes the record's authority** — a reader who checks one overstated entry discounts the
  others, including the true ones.
- And it **corrupts the count**, which is the artifact downstream guards compare against.

**The rule.** A record's numbers are asserted in **both** directions: a claimed occurrence must be
**shown to exist**, exactly as a claimed absence must be shown absent. **"Conservative" is not a
defence for a number that is wrong** — a risk register is a measurement, and [[D-000.26]]'s
cite-the-subject obligation applies to a claimed hazard as much as to a claimed clean bill.

**The tell that caught it**: the entry named a **file and an import**, so it was checkable — which is
[[D-000.35]] doing its work in the direction of *falsifying* a claim rather than supporting one.

---

### D-2.6.1 — The pagination model: a window onto one unbounded column, and no line is ever split

*(mechanism: binding)* — resolves DN-1/2/3.

**The window model is close to forced.** AD-4 gives pass one the whole layout and forbids pass two
from laying anything out; AD-24 makes boxes absolute and forbids negotiation. A model where pages are
laid out **independently**, or where elements **reposition per page**, violates both. Content as **one
unbounded column sliced into page-height windows** satisfies both and matches the epic's *"content
grows by producing more pages."*

**Line granularity is FORCED, not chosen** — a single text element can wrap past a page, so element
granularity would make **an element taller than a page unplaceable**. Say *forced* in the story so
nobody revisits it as a preference.

**The boundary rule is ruled as an invariant, because "the line's top edge" admits two readings:**

> **No line is ever split across a page boundary. A line is placed on the first page whose window
> contains it entirely** — from `baseline − max(ascent)` to `baseline + max(descent)`, per
> [[D-2.4.2]] as amended.

If *"top edge"* meant the decision **input**, with a fit check, that is this rule. If it meant the
boundary **test** — a line belongs to whichever page contains its top — **it is wrong**: it permits a
line straddling the boundary, drawn half on one page. **Nothing would diagnose that**, because FR44's
overflow machinery concerns content exceeding its **box**, not the page edge. **A bank statement
cannot ship a half-line**, and whitespace at the foot of a page is correct typesetting, not a defect.

**Images: atomic, same rule, one addition** *(binding)*. An image goes to the **first page whose window
contains it entirely**. An image whose declared box is **taller than the content window fits
nowhere** — an **overflow diagnostic under FR44**, never a silent clip and never a straddle. Note the
AD-24 interaction: the image is already scaled to fit its **box**, so this is a statement about the
box — making it a **template error with a located message**, not a render-time surprise.

---

### D-2.6.2 — A gate obligation is warranted when it is the only cross-target artifact for a shipped FR

*(mechanism: binding)* — resolves DN-4; the Epic 2 gate now owes **five**.

*"If refused, FR30 ships with no cross-target artifact"* settles it, and **that is the criterion** —
which is why counting obligations was never the right frame.

**And it is cheap because of the shape already ruled**: [[D-2.5.1]]'s check asserts the **registered set
against a declared list**, so a fifth is a **one-line reviewable diff** rather than a rename. The
mechanism paying off one story after it was ruled is worth a line in the gate.

**The heavy-test decline composes rather than conflicts.** The multi-page path **is** matrix-covered —
at the **gate**, via this obligation, not **in-story** via an override. Page-object ordering and xref
width are integer and already exercised: no float, vendor call, compressor or dependency.
**Obligation yes, override no.**

---

### D-2.6.3 — `folio-format.md` gains a pagination clause under D-000.6

*(mechanism: binding)* — resolves DN-5.

**Pagination is normative**: it determines what a `.folio` document renders to, a template author must
know it to design a report, and **a second implementation must paginate identically or the format
stops being a contract.** The clause states the four rules — window model, line granularity,
fit-entirely, images atomic — **as outcomes, not mechanisms**, per the correction now made three
times ([[D-2.3.1]], [[D-2.4.4]], and here).

---

### D-000.16 (limitation) — A rank order cannot express "no edge to this particular lower package"

*(mechanism: binding)* — recorded beside [[D-000.16]] so the next reader knows what the rank guard does
**not** cover.

`internal/pdf → internal/layout` is **legal today**: the rank rule is *"a package may import only
lower ranks"*, and `layout`(7) is lower than `pdf`(8). **So the guard enforces AD-5 and does nothing
for AD-4.**

**The limitation is structural.** A rank order expresses *"no backward edges"* but **cannot express
"no edge to this particular lower package."** `pdf` legitimately depends on `pagemodel` — **the
value** — and must not depend on `layout` — **the computation**. Both sit below it, so ranks cannot
separate them.

So Story 2.6's AC5 is **not a restatement**: it is a **supplementary forbidden edge the rank table
structurally cannot carry**, and a genuine forward guard on an open hole.

**Contrast with AC2**, which correctly **names the existing Story 2.5 guards** rather than adding a
third — [[D-000.42]]'s redundancy rule applied by someone deciding **not** to add coverage. **Both
dispositions are right and they are opposite**, which is the sign the rule is being reasoned about
rather than applied by reflex.

---

### D-2.6.4 — Pin `expected_breaks.json` in the ordinary suite, not only behind the matrix tag

*(mechanism: binding)*

A fixture whose **only** digest pin lives behind `//go:build matrix` can be **edited between gates with
the ordinary suite green.** That matters more than the feedback delay suggests: **the Thai break
sign-off binds to the break vector**, so an unpinned file means **the sign-off's subject can drift
silently while the sign-off still reads as valid.**

Pin it in the ordinary suite — cheap, and it closes the gap the sign-off binding depends on.

---

### D-000.50 — Before writing a guard, ask whether any subject can express the defect

*(mechanism: binding)* — standing pre-flight question; **twice now the fixture population, not the
assertions, was the binding constraint.**

- **Story 2.5**: moving the page-header origin was detected by **zero** tests — all six fixtures had an
  **empty `pageHeader`** and a symmetric page setup, making four geometric inputs mutually swappable.
- **Story 2.6**: **0 of 7 fixtures are multi-page**, 1 of 7 has a populated page header, 2 of 7 a
  populated footer. **No existing fixture can express this story's defect.**

**The question comes before the assertion, not after it.** A correct assertion over subjects that
cannot exhibit the defect is vacuous, and it is **invisible in review** because the assertion itself
reads as sound. Both instances were found by *laying out the actual fixtures and measuring*, not by
inspection.

**Companion to [[D-000.36]]**: that rule says a remedy for a vacuity inherits the vacuity's hazard;
this one says **check the population before you write the remedy at all.**

---

### D-2.6.1 (amended) — The sliding window; option (A)'s arithmetic is WITHDRAWN

*(mechanism: binding)* — the ruling settled the boundary **invariant** and then adopted option (A)'s
**arithmetic** by reference **without checking that the two compose.** Under a rigid grid a straddling
line has **no** window containing it entirely, so the invariant was **unsatisfiable** without a rule
(A) does not contain. The developer declining to choose between a ruled invariant and a ruled
arithmetic was correct.

**(A) is dead by measurement, not argument.** Bump the straddler flush to the next window top and
leave later items alone: line 44 → 0, line 45 → **7,590**, against an Advance of **16,344**. **The two
lines overlap.**

**(B) violates AD-24 by its own text**, which is stronger than calling it negotiation:

> *"A hidden element is absent from the `PageModel` and leaves no gap; **siblings never move, because
> nothing in a band ever reflows**."*

Under (B) an absolutely-positioned sibling **with nothing to do with the overflow** moves from
2,000,000 to 2,008,754. **Siblings move. A band reflows.** That is not an analogy to the clause — it is
the case the clause describes.

**Ruled: (C), the sliding window.** Window N+1 begins at the **top of the first item that did not fit
in window N**; page-relative Y is `columnY − windowStart(N)`; **page count falls out of the advance and
is not a closed-form function of the column height.** Option (A)'s `[N·H, (N+1)·H)`, `columnY − N·H`
and `ceil(lowestBottom/H)` are **withdrawn** — they were (A)'s, and (A) is eliminated.

**(C) is the only model under which nothing repositions**: every element keeps its declared column
position and only the **window** moves. It is also the only reading under which every sentence of
D-2.6.1 is literally true, with no clamp and no synthetic column gap.

**A consequence stated now rather than discovered later:** windows advance to the top of the **first
unplaced item**, not by a fixed `H`, so **no page is ever empty** — an element far below the text
starts the next window instead of generating blank pages.

**And the consequence authors must be told** *(binding — goes in `folio-format.md` with the pagination
clause)*: **across a window boundary, declared vertical gaps collapse.** An element that starts a
window renders at the top of its page, whatever gap was declared above it. **That is the price of
never splitting a line and never reflowing a sibling** — a real semantic an author needs before
designing a report, not a detail to be found in a diff.

**Sub-question (i) — column-wide, and the tension dissolves rather than being traded.** The window is a
property of the **page**, not of an element. Both pulls are then satisfied at once: *"elements never
reposition"* (every declared column Y is untouched) and *"a paragraph's leading survives a page break"*
(the lines keep their column positions, so the gap between line 43 and line 44 is still exactly
`Advance`; they merely fall in different windows). **They only appeared to conflict under (B), where
the column is mutated.**

**Sub-question (ii) — a line taller than the content window is in scope, and it is the same FR44 case
as the oversized image** *(binding)*. It fits in **no** window, so ruling *"never split"* without
ruling this leaves the implementation to split, loop, or panic. Reachable at `fontSize > H`.
**Disposition: place it at the top of a page, clip at the window bottom, emit a located diagnostic
naming the element** — FR44's sanctioned behaviour. **Clipped is not split**: the remainder does not
reappear on the next page, so the boundary invariant holds. **One overflow rule, two subjects.**

---

### D-000.51 — When you justify a remedy by describing the current state, check the current state

*(mechanism: binding)* — the lead's own standing form, after enough repetitions to earn one.

[[D-2.6.4]] justified pinning `expected_breaks.json` by saying its *"only digest pin is matrix-tagged
today."* **Measured: the literal occurs at zero sites.** The matrix-tagged test **computes** the digest
at run time and compares it against a `break-signoff.json` **that does not exist**. **The file was
pinned nowhere at all.**

**The remedy survived; the justification did not — and only one of those is self-correcting.** A
remedy that is right for a wrong reason will be re-derived by the next reader from the wrong reason,
and may then be scoped wrongly or dropped when the stated condition appears not to hold.

**And distinguish two things the same missing file explains**, because conflating them would have hidden
a real hole behind a working mechanism:

- **`break-signoff.json` absent → the matrix test fails.** That is **not broken**: it is
  [[D-2.3.5]]'s pending-sign-off blocker **doing its job**, refusing to let the gate pass.
- **The fixture pinned nowhere → it could be edited with the whole suite green.** That was a **genuine
  hole**, and it is larger than the ruling described: **the Thai break sign-off's subject could have
  drifted while the sign-off still read as valid.**

---

### D-000.52 — A structural claim about a guard is worth exactly as much as its demonstration

*(mechanism: binding)*

[[D-000.16]]'s limitation — *a rank order cannot express "no edge to this particular lower package"* —
was **asserted from the rank rule's shape.** Story 2.6's developer **demonstrated** it: with
`pdf → layout` in place, **lint's `stage-rank` stayed fully green across all four of its tests while
the new guard reddened.**

**The hole is measured, not argued.** A claim about what a guard *cannot* see is exactly the kind that
sounds obvious and is routinely wrong — and it is cheap to demonstrate, because you need only
introduce the violation and observe which guard stays green.

**Companion note**: the same developer **declined to claim the story's red-proof numbers**, because the
story never stated **which Latin sentence it repeated** — and measured its own subject instead, citing
it. That is [[D-000.26]] applied by an agent to a document written for it, and the **second time this
week** an agent has refused to inherit a number it could not attribute. **Both times the inherited
figure turned out to be wrong.**

---

### D-2.6.5 — An item that fits in no window is a located template error, for BOTH subjects

*(mechanism: binding)* — appended, never edited into [[D-2.6.1]]. **D-2.6.1 (amended) sub-question
(ii)'s "clip at the window bottom" disposition is WITHDRAWN.**

An item that fits in **no** window — an image whose declared box exceeds the content window, or a line
whose leading exceeds it — is a **located error returned from `Render`/`RenderTo`, naming the
element**. `internal/layout.OverflowError{ElementID, ItemHeight, ContentHeight, Kind}` as shipped is
the ruled shape: **no clip bound on `pagemodel.TextRun`, no `q … W n … Q` path in `internal/pdf`, no
change to the public signature in this story.**

**Ground 1 — FR44 does not reach this case, and D-2.6.1 says so in its own words.** The ruling states
*"FR44's overflow machinery concerns content exceeding its **box**, not the page edge."* Its own
amendment then disposed of the oversized line as *"FR44's sanctioned behaviour"* — **invoking FR44 for
a page-edge case the ruling it amends had already excluded from FR44.** Story 2.6's scope fence draws
the same line (*"this story changes how many sheets exist, not what happens at a box edge"*), and the
story's DN-3 had already observed that Story 2.8 is scoped to *"content exceeding its declared bounds,
which a page boundary is not."* **Three independent sources, all pre-dating the divergence.**

**Ground 2 — both subjects are declaration-level, not render-time, and that is what makes *"one
overflow rule, two subjects"* actually true.** D-2.6.1 made the image a template error precisely
because *"the image is already scaled to fit its box, so this is a statement about the box."* **The
same holds for a line**: per [[D-2.4.2]] constraint 1, leading is a function of **(declared font stack,
font size)** and never of the glyphs that appear; the content window is page height minus declared
margins and band heights. So *"some line is taller than the window"* is **decidable from template +
fontset with no data at all.** Two declaration-level impossibilities → one disposition. **The
amendment's slogan was right; its two dispositions contradicted it.**

**Ground 3 — "clip AND diagnose" is not expressible today.** Verified: `Render(t, d, p, f) ([]byte,
error)`, `RenderTo(w, …) error`, and `folio-go/internal/` holds eight packages, **none of them
`diag`** — Story 3.6 mints it (DW-6). **A ruling that forces a choice between returning bytes and
diagnosing will have the diagnostic dropped**, which is the half that ships silently wrong output.

**Guardrails** *(binding)*:

- **Red-proof needs a synthetic template, not a fixture** — [[D-000.50]] already named this story's
  fixture population as unable to express the defect. Assert the error's **element id and `Kind`**,
  with **both subjects exercised** (`fontSize > H` for `"line"`, an oversized declared image box for
  `"image"`). **One subject only would assert "one rule, two subjects" while only one subject can
  express it.** Assert on the fields, never on *"an error occurred"*.
- **`Kind` is derived as `"image"` when `len(it.Images) > 0`.** If an item can ever carry both runs
  and images, that classifier **mislabels** — assert the exclusivity where items are built, or make
  the mixed case explicit.
- **`folio-format.md`'s pagination clause must state the outcome as *fits-nowhere is a load/render
  error naming the element*, and must NOT mention clipping.** The file carries no pagination clause
  yet, so this is **written correctly the first time rather than corrected.**
- **Byte-neutrality untouched** — no valid document changes bytes; nothing moves a golden on this.

**Story 2.8 inherits work larger than "clip machinery", and its creator must be told now.** `epics.md:938-950`
requires clip **and** diagnostic **and** *"PDF bytes are still returned — clipping degrades output but
does not fail the render."* **That is not expressible in the current public API**, so **2.8 owns a
diagnostic-channel design decision on `Render`'s surface.**

**Scope, stated inside the ruling:** this settles **page-edge** overflow only. If, once 2.8 has built a
non-fatal channel, anyone wants page-edge overflow reclassified as clip-and-continue, **that is the
owner's call, not a re-ruling** — the trade is *"a template that fails loudly today would then ship a
page with content cut off"*, which is a product judgment.

**The assumption that would flip it**: if leading could ever depend on the glyphs actually present —
making oversized-line-ness **render-time** rather than declaration-time — ground 2 collapses and the
two subjects genuinely differ. D-2.4.2 constraint 1 forbids exactly that, and AD-24 depends on it.


---

### D-000.53 — A golden is not accepted until a reader we did not write resolves it into the objects it claims to contain

*(mechanism: **binding** for the rule; **illustrative** for `qpdf`)*

**The rule.** No golden artifact is accepted — **first recording or re-recording** — until a reader
this project did not write parses it and resolves it into the semantic objects it claims to contain.
The page tree resolves to N pages, every indirect reference in it names an object that exists, and N
equals the count the artifact declares. **The reader, its version, and the verbatim invocation are
recorded in the fixture's provenance with its output.**

Stated as an **outcome, not a tool**. Not *"`qpdf --check` must exit 0"* — that is a mechanism, and
this run has already had to correct mechanism-named rules three times ([[D-2.3.1]], [[D-2.4.4]],
[[D-2.6.3]]). `qpdf --check` (12.4.0, `/opt/homebrew/bin/qpdf`) is today's instrument.

**What forced it.** Story 2.6 recorded `fixtures/multi-page/expected.pdf` — the project's first
multi-page document — and it would not open. Preview refused it; pdf.js reported *"page 0 of 2"*.
Object 2 was emitted as `<< /Type /Pages /Kids [8 0 R10 0 R] /Count 2 >>`: **no separator between two
indirect references.** A tokenizer reads `R10` as one unknown token, so **neither kid resolved** and
the page tree was empty. `/Kids` is the only site in the package emitting two refs back-to-back —
`/Font` and `/XObject` write `" /"+name+" "` before each, and the singular sites are followed by a
literal — and `document.go:81` hardcodes exactly one kid. **The defect could not manifest until a
document had more than one page.** The owner found it by trying to open the file.

**The framing that matters, corrected by measurement rather than assumed.** The first diagnosis — *"the
semantic acceptance step had a blind spot"* — was **wrong, and understated it**. `render_test.go:600`'s
`assertWellFormedPDF` fatals on `pageCount != 1` (line 615), so it is **unusable on a multi-page
document**, and `multi_page_fixture_test.go` **does not call it** (0 matches, against 18 call sites
elsewhere). The file was not validated-and-passed; **the only instrument that could have looked encodes
the single-page shape as an invariant, and was therefore silently never invoked.** *The first artifact
of a new shape is validated by nothing, because every existing guard was written when that shape did
not exist.*

**Why every bespoke check passed a file that is not a document.** Each was true, for its own reason:
`/Type /Page` count == 2 counted **object definitions** — both pages were emitted correctly and both
were reachable in the xref; what was broken was the array pointing at them. `/Count 2` read an integer
that was right — **`/Count` is a claim about the kids, not a fact derived from them.** The per-page
header/footer Y, the line→page partition and the no-negative-Y check parsed **content streams**, which
do not care whether anything references them. Byte-identity across two processes proves determinism —
**a deterministically wrong file is byte-identical to itself.**

The common shape, and the reason the remedy is an oracle rather than another assertion: **every check
asserted a property of a PART; the defect was in a REFERENCE BETWEEN parts.** Every assertion measured
**presence**; the property that mattered was **reachability**. Three refs existed; none resolved. No
quantity of added per-part assertions would have found it.

**Grounding**: [[D-000.22]] (a hash certifies non-change and can never certify correctness — the
acceptance step is the only thing standing in that gap) + [[D-000.44]] (a re-recording is a recording,
which is why *"first"* is struck from the rule) + [[D-000.21]], **sharpened**: assert on the artifact
that carries the property, **and prove it carries it**. Precedent verified rather than inherited:
Story 1.1 AC2 used `qpdf` as an independent oracle **specifically to avoid a module dependency**, and
**AD-25 verbatim** — *"a one-time offline reference run, hand-checked — never a runtime dependency"* —
which [[D-2.3.1]] applied to shaping via `hb-shape`. Nothing new is invented; this applies the
established shape to document structure.

**Guardrail**: **this step is not discharged by an assertion this project wrote.** That is the whole
point of the independence, and the measurement above is why — our own checker excluded the subject.

---

### D-2.6.6 — The eight goldens are validated and the evidence recorded, but structural validity is NOT a gate obligation

*(mechanism: **binding**)*

All eight goldens are validated and the measurement recorded — fixture path, `qpdf` version, **verbatim
invocation**, output line — in Story 2.6's Delivery Log **and** each fixture's provenance/README. Seven
read *"settled, validated at `<commit>`"*; the eighth says it was broken and names the fix. A run by the
orchestrator or the developer is **not** a project record ([[D-000.26]]), and an audit row ends
**settled or fixed, never carried** ([[D-000.29]]).

**It does not become a sixth gate obligation.** Applying [[D-2.6.2]]'s own criterion — *a gate obligation
is warranted when it is the only cross-target artifact for a shipped FR* — structural validity of
committed bytes **is not cross-target**: the bytes are identical on every leg by construction, and that
identity is precisely what the matrix already proves. Putting it in the gate spends the gate's attention
on something that **cannot vary by leg**, and erodes [[D-2.5.1]]'s declared-list mechanism back toward
counting. **The Epic 2 gate still owes exactly five**; `TestEpic2GateObligationsMatchTheDeclaredSet`
stays green untouched and the gate is written on five.

---

### D-2.6.7 — Two oracles, two different jobs: external at recording time, in-repo on every leg

*(mechanism: **binding** for the split; **illustrative** for the mechanics)*

- **The external oracle is the acceptance instrument, and it runs at recording time only** — off-leg, on
  the recording machine, hand-checked, output pasted into provenance. It is **never** a runtime or CI
  dependency (**AD-25**, `wantModuleGraph`/`TestModuleGraphAllowlist`), and js/wasm has no `qpdf`.
  **It must not be gated to "the legs that have it"**: a check that runs on some legs and not others
  produces exactly [[D-000.9]]'s signal — an *"all clear"* indistinguishable from *"I could not look"* —
  one level up, at the leg.
- **The in-repo checker is the standing regression guard, and it runs everywhere.** Two repairs are
  **required**, both forced by the measurement in [[D-000.53]]: (1) **`assertWellFormedPDF` takes the
  expected page count as a parameter** rather than hardcoding 1 — per [[D-000.34]] the `pageCount != 1`
  assertion is **not** deleted, because on the seven single-page fixtures it is load-bearing and its
  death would be silent; generalise it and pass 1 at the 18 existing call sites. (2)
  **`multi_page_fixture_test.go` must call it.** *A checker that exists and is not called on the one
  subject able to express the defect is worth nothing* — that, not the missing byte, is the process
  defect.

**The asymmetry is now measured, not hypothetical**: an in-repo checker written by the same hands **did**
share the emitter's blind spot, by excluding the subject. That is the argument for the external oracle at
recording time. It is **not** an argument against the in-repo checker, the only thing that runs on every
leg on every story. Per [[D-000.38]] the two are genuinely independent **in mechanism** — one is our
parse, one is a third party's — so the pair is **real redundancy, not two names on one parse**.

**Guardrail — [[D-000.30]], window closing.** The fix was already in the working tree when this was ruled.
The defect must be captured **before** commit: revert the separator, render, keep those bytes, and show
that the repaired `assertWellFormedPDF`, `golden_structural_validity_test.go`, and the external oracle
**each** go red on them. Per [[D-000.40]] **the mutation asserts it applied by a non-empty diff, never by
an exit code**; per [[D-000.36]] the remedy is run against the real defect bytes before it is credited.
A guard not shown red on **these** bytes is a forward guard ([[D-000.24]]) and must be labelled one.

---

### D-2.6.8 — Three layers, three different properties: emitter helper, token well-formedness, page-tree semantics

*(mechanism: **binding** for the invariants; **illustrative** for the spellings)*

The artifact-level page-tree check is **not sufficient**, and the emitter fix is ruled **in**.

1. **Emitter helper** *(binding)* — a `writeRefArray(ids []int64)` / `appendRefArray` that takes the slice
   and **cannot** omit the separator, replacing the hand-rolled loop. **Leading** separator, so the one-kid
   case stays `[8 0 R]` byte-for-byte. This prevents the class **by construction**, which outranks any
   guard over it.
2. **Artifact-level delimiter check** *(binding as invariant, illustrative as mechanism)* — the defect's
   true shape is *"a `N 0 R` token not followed by a delimiter."* Key the guard on **that**, not on
   `/Kids` ([[D-000.15]]: key a guard on its purpose, never on a proxy). **Named caveat to be resolved,
   not hand-waved**: binary stream data can contain any byte sequence, so it must be scoped to non-stream
   regions (`assertStreamLengthsAreExact` already locates them). For every ref array other than `/Kids`
   this is a **forward guard with no available red-proof** ([[D-000.24]]) and must be labelled one.
3. **Keep `golden_structural_validity_test.go`** — sound, and **not** redundant with (2): it catches a
   **wrong `/Count`** and a **dangling kid**, which a token check cannot see.

**Against the "one instance today" argument**: [[D-000.23]]'s remedy is about **cardinality**. `appendRef`
has ~13 call sites and there are **two** page-tree emitters (`document.go:81` hardcodes one kid;
`textdoc.go:186` is the loop). **`/Kids` being the sole instance is a fact about today's feature set, not
about the code** — Story 2.7, Epic 4's tables, and any future `/Annots` or name tree add ref arrays.

**Do not count these as three coverages of one property** ([[D-000.42]]): construction, token
well-formedness, page-tree semantics — **three properties, three mechanisms**.

**Guardrail**: the working-tree comment claims the leading separator *"keeps all seven single-page goldens
from moving."* Per [[D-000.52]] — **measure it**, do not accept the argument: seven digests unmoved, at
every declared site ([[D-000.47]]'s four-site list), by the file route ([[D-000.12]] as corrected,
`rtk proxy` first then redirect).

**The assumption that would flip (3)'s ordering**: if the delimiter check cannot be cleanly scoped out of
stream data, **invert the layers** — the helper plus the page-tree check stand alone and the delimiter
check is **dropped** rather than shipped as a false-positive generator ([[D-000.15]]: a false positive in
a guard is an attack on the guard).

---

### D-000.54 — A newly registered matrix document runs its native leg at registration; this is a sequencing fix, NOT a cadence override

*(mechanism: **binding**)*

When a story adds an entry to `matrixDocuments`, it runs that document's leg **on the host target only**,
once, before the story may reach `review`:

```
FOLIO_MATRIX_TARGET=darwin/arm64 go test -tags=matrix -count=1 -run TestTargetRenderHash .
```

**It must not be logged as a [[D-000.4]] override.** D-000.4 requires a genuinely **new source of
cross-target divergence** — float arithmetic, a vendor call, a compressor, a new dependency — **not
merely a new golden**, and warned that on a weaker trigger *"nearly every story would qualify and
D-000.4's trade would erode to nothing."* Registering a document is exactly that weaker trigger.
Logging it as an override spends the owner's deliberate wall-clock trade to buy something the override
was never the right instrument for.

**It is [[D-000.43]]'s move** — the one the log already recorded as *"strictly better than the
override"*, where 2.5a bought the protection that mattered at zero cost instead of a Docker run. The
expensive thing in D-000.4 is the **Docker arm64 boot**; the host leg is a local `go test` with no
Docker at all. The owner's trade is untouched.

**What forced it.** Story 2.6 registered `multi-page` and reported its legs *"written, compiled and
vetted, deliberately not run."* The leg **failed unconditionally on all four targets**:
`requireMultiPageIsGenuinelyMultiPage` reads runs via `readEmittedRuns`, which takes the **first** text
content stream and `break`s — one page's worth — then fatals demanding two headers. Reproduced at the
real entry point, `matrix_test.go:1233`. The other seven legs hashed correctly.

**Guardrail — state precisely what the native leg proves, because it is narrower than it looks.** It
proves the leg **executes and produces a hash on one target**. It proves **nothing** about cross-target
agreement, which remains the gate's job. **Both halves go in the Delivery Log**; the second half is what
stops this becoming the next unearned assurance.

---

### D-000.55 — `go vet -tags matrix` is a compile gate, not an honesty check; and the compound phrase is banned

*(mechanism: **binding**)*

`go vet -tags matrix` stays — it is still the only thing that can check a leg the host cannot run — but
it is **no longer the honesty check for a deferred leg**; [[D-000.54]]'s native-leg execution is. **Vet
may never again be offered as assurance that a deferred leg works.**

**Grounding**: [[D-000.15]] — key a guard on its purpose, never on a proxy for its purpose. The
obligation is *"this leg will run at the gate."* Vet keys on **compilation**, a stand-in for it, and the
gap between the two is precisely where Story 2.6's Blocker lived. And [[D-000.9]]: the story's report
was **true**, and its all-clear was **indistinguishable from could-not-look**.

**The wording rule — the transferable half.** **Ban the compound phrase.** *"written, compiled and
vetted, deliberately not run"* is a true sentence that reads as *"checked"*. Split it:

- what was **executed**, naming **which target by name** ([[D-000.26]] — cite the subject);
- separately, the legs **not run**, named **individually**, with the cadence clause that defers them.

**This is not a new obligation.** [[D-000.4]]'s Consequences already require each Delivery Log to *"name
the suites it actually measured and name the unrun ones explicitly."* The defect is that **"vetted" was
allowed to occupy the "measured" column.** Record it as an existing rule being enforced, so the next
reader sees enforcement rather than a new rule arriving.

---

### D-2.6.9 — Story 2.6a is inserted before 2.7 to sweep the shared PDF-reading helpers; the predicate is behavioural, never a grep

*(mechanism: **binding** for the predicate, the placement and the cost control; **illustrative** for the
inventory)*

**Sweep it, as an inserted Story 2.6a between 2.6 and 2.7.** Ruled rather than escalated: it discharges
[[D-000.50]] (*before writing a guard, ask whether any subject can express the defect*) and
[[D-000.53]], both already binding. **It is not new scope, so it is not the owner's.**

**The design point that decides whether the sweep is worth anything.** Do **not** sweep for the literal
`1`. Of the three instances, **two are not spelled with a `1`**: the `/Kids` bug was a **missing
separator**, and `readEmittedRuns` is a **`break` after the first stream containing `" Tf\n"`**. Only
`matrix_test.go:1212` was a literal. **A text-matched sweep finds one of three** — [[D-000.14]] exactly:
it would measure the regex.

**The predicate that works.** Enumerate the quantities the engine can now emit **N** of — pages, content
streams, page-header runs, page-footer runs, embedded faces, image XObjects, xref subsections — and for
each, **run the existing shared helpers against a subject with N > 1 and observe which fatal, mis-count,
or silently truncate.** The subject already exists: the multi-page fixture. *Enforce the property
behaviourally, never by name-matching* — the same reason `assertWellFormedPDF` had to be parameterised
rather than grepped for.

**The cost, measured rather than guessed** *(orchestrator-verified, text-matched — per [[D-000.14]] this
is the sweep's starting inventory, not its answer)*: `readEmittedRuns` has **12 call sites across 5
files** — `shaped_fixture_test.go` 5, `matrix_test.go` 4, and one each in `wrapped_text_fixture_test.go`,
`segment_origin_test.go`, `first_baseline_acceptance_test.go`. Every one is single-page today, so **all
twelve are correct today and all twelve become silently wrong the moment their subject is multi-page.**
That is **not twelve bugs; it is one helper with an undocumented contract and a silent truncation.** The
helper population is small and enumerable — `pdfObjects` (2 calls), `streamBody` (3), `readEmittedRuns`
(12), `countPageObjects` (1), plus `assertWellFormedPDF` (already repaired). **The sweep is bounded
because the helpers are few, not because the call sites are.**

**Why 2.6a and not the gate or 2.7.** Not the gate — a gate is a **verification point, not a work item**,
and this fails [[D-2.6.2]]'s criterion. **The Epic 2 gate still owes five.** Not folded into 2.7 —
[[D-000.25]]'s reasoning for inserting 2.3a applies almost verbatim: **audit first**, so the next story's
new call sites are written against a known map, and do not stack two guard workstreams onto a story that
needs focus. **Before 2.7 urgently**: 2.7 is `Page X of Y`, which adds **a run on every page** and will
land directly on these helpers.

**Guardrails**: [[D-000.17]] governs the budget — if the enumeration exceeds what 2.6a can carry, the
remainder is **reported unmet with each instance named**, never quietly dropped and never triaged by
taste; make the enumeration-and-count 2.6a's **first** task so the scope decision is made on a number.
[[D-000.48]] — a correction sweep can introduce a fresh instance of the class it corrects, and this one
corrects **silent truncation**, so verify each repaired helper against the multi-page subject
mechanically. [[D-000.34]] — repairing `readEmittedRuns` may **kill a detector**: check whether any of
the twelve call sites derived its discriminating power from reading only the first stream, and
**re-point rather than re-run**. [[D-000.42]] — count what the sweep **fixes**, not what it inspects.

**The evidence, stated as evidence rather than as principle**: [[D-000.53]]'s framing has now held
**three times in one story** — the `/Kids` emitter, `matrix_test.go:1212`, and `readEmittedRuns` — and
each was found by a **different route**: a human opening the file, a developer enumerating, and a
reviewer running the real entry point. **None was found by inspection.** That is the argument for the
behavioural predicate above.

---

### D-2.6.9 (correction) — the call-site count was measured on an uncommitted working tree; it is 11, not 12

*(mechanism: **binding** for the correction; the ruling's substance is unchanged)*

[[D-2.6.9]] records `readEmittedRuns` as having **12 call sites across 5 files**. **The committed count is
11**, and it is 11 at **every** commit — `44a1d0f` (where the 12 was written), `4001354`, and `d1af588`:
12 occurrences minus the definition. Story 2.6a's creator caught it.

**Per-file, at `d1af588`**: `shaped_fixture_test.go` 5, `matrix_test.go` **3**, and one each in
`wrapped_text_fixture_test.go`, `segment_origin_test.go`, `first_baseline_acceptance_test.go`.

**The cause, stated precisely, because the obvious explanation is wrong.** It is *not* that the pattern
`readEmittedRuns` also matches the sibling `readEmittedRunsAllPages` — the grep used a trailing
parenthesis and could never match it. The cause is that **the orchestrator grepped the working tree
while Story 2.6's work was still uncommitted in it.** `matrix_test.go` carried **4** calls in that
working tree and **3** in the committed tree; the finisher later converted one to
`readEmittedRunsAllPages`, returning it to 3. **12 was true of a tree that could not be checked out,
and 11 is true of every commit.** Two real numbers about two different subjects.

**This is [[D-000.26]]** — *cite the subject of a measurement* — and the citation that was missing is
**the commit**. A count taken from a tree with uncommitted work in it is a measurement of a moment, not
of the repository, and it is indistinguishable in the record from a measurement of the repository.

**Consequence, generalised** *(binding)*: **an enumeration recorded in the decision log or a story file
must name the commit it was taken at**, and must be taken against a tree with no uncommitted changes in
the files being counted — or must say explicitly that it was not. This closes the same gap
[[D-000.55]] closed for test evidence: a true statement that reads as a stronger one.

**Note the pattern, since [[D-2.6.9]] exists to correct exactly it.** The orchestrator recorded a count
from intent-adjacent evidence rather than from a citable artifact, which is the developer's
*"I counted what I intended to add, not what the artifact contains"* one level up — **inside the ruling
that was written to correct that behaviour.** [[D-000.48]] anticipated this: a correction sweep can
introduce a fresh instance of the class it is correcting. **The substance of D-2.6.9 is unaffected** —
the predicate, the placement, the cost control and the guardrails all stand; only the number moves, and
it moves in the direction that makes the sweep slightly cheaper, not larger.
