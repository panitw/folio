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
