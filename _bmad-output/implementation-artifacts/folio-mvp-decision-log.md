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
