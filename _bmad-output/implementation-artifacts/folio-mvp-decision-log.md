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
