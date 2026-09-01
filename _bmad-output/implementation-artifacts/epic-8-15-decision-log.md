# Decision log — Epics 8–15 delivery run

Opened 2026-09-02. This log continues `epic-7-8-decision-log.md`, which covers Epic 7 in full and
Epic 8 through Story 8.4f. That file is NOT superseded and must not be edited by this run: rulings
D-7.x, D-8.x and D-R7.x cited here live there. New rulings are numbered `D-<story>.<n>` and appended
here.

**Run target.** Every open item in Epics 8–15: Epic 8's remaining three stories, the retroactive
closure of Epics 9 and 10, then Epics 11–15 to the `folio-go/v0.1.0` tag.

---

## Lead Grounding

Filed verbatim 2026-09-02 from the run-scoped engineering lead, grounded once at baseline `1f8e52b`.
This is a **continuation** — the lead re-grounded from the existing record (`epic-7-8-decision-log.md`'s
own `## Lead Grounding`, the standing decisions and the 166 ruling headings) rather than re-deriving
the program from the architecture. Everything the report marks MEASURED was verified at HEAD by the
lead itself rather than read from a document. The lead does not re-read the spine, the ADRs, the epics
or either decision log on later resumes; per decision it reads only what that decision hinges on.

### Sources read

- `epic-7-8-decision-log.md` — `## Lead Grounding` and `## Standing decisions` in full; the 166 ruling
  headings enumerated; D-7.3.1, D-7.8.1–.4, D-8.1.1–.3, D-8.2.1–.8, D-8.3.1–.6, D-8.4.3, D-8.4.21–.35,
  D-8.5.1–.11 read in full.
- `epic-8-15-decision-log.md` — all of D-000.1…D-000.8, binding and not to be re-litigated.
- `sprint-status.yaml` — all 420 lines including every inline comment.
- `planning-artifacts/architecture/architecture-folio-2026-08-23/ARCHITECTURE-SPINE.md` — the
  stage-rank table and AD-1…AD-26 in full, plus Consistency Conventions and Stack.
- `planning-artifacts/epics.md` — Epic 8's header and Stories 8.5 / 8.6 / 8.4d; Epics 9–15 in full.
- `specs/spec-fonts/SPEC.md` `## Open Questions`; `specs/spec-folio/folio-format.md` version rules,
  `align`, `keepTogether`.
- `deferred-work.md` — by heading (113 entries), plus DW-24, DW-68, DW-73, DW-75, DW-80, DW-100,
  DW-101, DW-103, DW-108, DW-109 in full.
- Code at HEAD: `folio-go/component_commands.go`, `folio-go/internal/template/serialize.go`,
  `folio-go/version.go`, `folio-designer/src/release-payload.ts`,
  `folio-designer/scripts/verify-offline-release.mjs`, `.github/workflows/`, `folio-go/fonts/`,
  `folio-designer/public/fonts/`.

**Confirmed, carrying the prior lead's correction forward: there are no separate ADR files.**
`ARCHITECTURE-SPINE.md` is the only architecture document and its `AD-N` invariants *are* the ADRs.
Rulings cite `AD-N` plus a SPEC clause, a format clause or a story AC — never an ADR path.

### The invariants the lead rules from

**Byte identity is the product, not a quality bar.** AD-1 (determinism is a *directory* boundary —
every package under `internal/` is render path), AD-2 (`geom.Length`, int64 millipoints, one owner,
round-half-to-even), AD-3 (numbers reach the PDF through one file in exactly two representations),
AD-23 (no `float64` under `internal/`, geometry or data), AD-21 (four-target matrix; a moved hash is a
defect until proven intended — counter-metric C6), AD-22 (any change to layout, subsetting, emission,
the locale table or the toolchain is breaking and ships as such).

**One authority per fact**, recurring at every layer: one number→text emitter (AD-3), one coordinate
flip (AD-24), one canonical serializer (AD-9), one scalar→text converter, no TypeScript model of a
document (AD-15), the browser never measures (AD-17), fonts arrive as an explicit value and are never
queried from the host (AD-8).

**Command / projection** (AD-15, AD-16): the UI paints from an immutable snapshot and sends every
committed mutation as one opaque versioned command. This binds essentially all of Epics 12, 13, 14 and
the designer half of 11.

**The stage-rank law is executable and lives in code**, not in the spine:
`lint/internal/rules/stagerank.go`'s `stageRankTable`, enforced by `TestStageRankProductionScan`. The
spine's ladder is held in agreement by a test. **If they disagree, the table is right.**

**The window model** (D-2.6.1): pages are sliding windows onto one unbounded content column, never
rearranged. Epic 12's band-height work (12.1, 12.5) is a constraint *on* that model, not a change *to*
it.

**The overflow scope line.** Content exceeding its declared *box* is FR44 / Story 2.8 — clip, diagnose,
still return bytes. An item taller than the *window* is a located template error. Every new overflow
question routes through that distinction first. AD-14's refinement (post-7.10): leniency is scoped by
**what** is over-tall, never by whether it is grouped.

**AD-9's P1** (`Parse(Serialize(d)) == d`) is *forced*, not chosen. `serialize.go:510-514` preserves
orphaned assets unconditionally and its comment states there is "no policy latitude to drop one".
MEASURED still true at HEAD. This matters enormously for Story 8.6.

**AD-14's diagnostic contract:** one closed, additive `SCREAMING_SNAKE` registry; callers match on the
code, never on message text; clipped content and over-tall rows are Warnings returned *alongside* PDF
bytes.

**AD-26's Rule has two clauses, and its heading is not its Rule** (D-8.5.3). The family ban is scoped
to *"No dependency"*; redistributed **assets** owe only their licence text and copyright lines.
Extending the ban to assets was an **owner decision**, now made: a named permissive allowlist
(**OFL-1.1, Apache-2.0, MIT, UFL**), enforced like the dependency ban — fail the build, never warn —
with unclassifiable treated as failure.

### The run's own standing law, as the lead will apply it

- **D-000.8's bar, verbatim.** Escalate only: a scope cut/expansion that changes what `v0.1.0` is; a
  change to the byte-identity regime, the `.folio` format, or the public API surface; the Epic 11
  realize-vs-retire fork; a security boundary; a decision that forecloses a direction the epics assume
  open. Everything else is ruled.
- **D-8.2.2 is the standing tagged-surface test, not re-derived per case.** (a) *Is it in the tagged
  surface?* — is it reachable through the module's **exported** API (`folio.Canvas`,
  `folio.CanvasWithTextPaint`, `Render`, `LoadTemplate`)? "Which binary calls it" is **not** the test.
  (b) *Must it land before the tag?* — only if it is a **narrowing or removal that has not shipped
  yet**. Widenings are always safe.
- **D-7.8.3's before-the-tag set, current state.** Item 1 (Story 7.8's justified-table load narrowing)
  **landed**. Item 3 (Story 8.4's non-font-asset render refusal) **landed**. Standing open:
  **D-7.8.2's audit** — retire whichever of the two existing style codes no consumer branches on (a
  removal) — and **DW-32/DW-73/DW-75 via Story 15.2a** (a narrowing of the exported
  `ApplyComponentCommand`). Plus an unnamed fourth: **DW-68**.
- **Practice rules this run earned and keeps applying:** an anchor written at a plan gate is a claim
  with an expiry date, re-verified at the gate that consumes it (D-7.8.4). A count written next to the
  thing it counts is a literal that stops being true the moment the thing grows (D-8.5.4 — three
  instances this epic). Moving a threshold or an AC to match a measurement is the defect, not the fix
  — but correcting a **false premise about how the bar is reached** is legitimate; they look identical
  in a diff and are opposite in kind (D-8.5.10). A recommendation that names its own defeater costs
  one round trip instead of a wrong build (D-8.4.24). A ruling recorded everywhere except the document
  that gets read is a ruling that did not happen (D-8.4.31).

### Constraints Epics 8–15 must not break

1. **The corpus is the witness.** Epics 9, 10, 11, 12 and 14 all carry an explicit "a document
   declaring none of this hashes identically" clause. Every one is checkable and must be checked, not
   asserted. Under D-000.4's per-epic cadence, any story whose own correctness is byte-shaped gets its
   integration commands in its **own** `## Verification` regardless, with the override logged.
2. **Epic 11 is the only remaining epic that can move a golden hash** — which is exactly why D-000.3
   sequences it late. Nothing in 12/13/14 emits a byte the engine did not already emit.
3. **The format ceiling is `2.0` and does not move for anything currently planned.** MEASURED at
   `folio-format.md:47`: a document declares the lowest version its own content requires — `2.0` for a
   non-table `align: "justify"` or an embedded-face chain entry, `1.2` for `keepTogether`, `1.1` for
   `lineSpacing` or `color`, else `1.0`. `SupportedMajor` stays at 2. If any Epic 12 story proposes a
   new format field, its version trigger must be **derived, not asserted** (DW-81's shape).
4. **Epic 13's scope fence — the sharpest in the run.** Story 13.4 makes an absent path render
   **empty**, and `BINDING_PATH_ABSENT` must stay a located **Error** in `Render`. The empty-value
   behaviour is **preview-with-no-data only**, never `Render`'s semantics; the engine's error contract,
   its diagnostic codes and the golden corpus are untouched. A drift here silently prints a blank where
   a customer's name belongs, on 50,000 statements. Any proposal to move this into the engine is a
   direction change and is escalated.
5. **Epic 13's second fence:** AD-17 forbids the browser measuring text. Story 13.2 needs a container
   pixel width and takes it under the **narrow, explicitly named** exception `src/preview/` already
   holds for `scroll*` — the exception names the new property and is never widened by wildcard.
6. **Epic 14 changes no rendered byte, no document model and no command surface** by its own header.
   14.9 is the single exception and it is a **projection** field, not a format field. Note carefully:
   `folio.Canvas` is **exported**, so a new `CanvasProjection` field *is* on the tagged surface
   (additive, therefore safe — but "nothing about the document changes" must not be read as "nothing
   about the tagged surface changes").
7. **DW-74 is live and 14.9 walks straight into it:** the Go/TypeScript wire test records the
   projection's top-level and `CanvasFontChain` key lists but **not** `CanvasComponent`'s — so a
   per-component projection field can blank the canvas with every test on both sides green. 14.9 adds
   exactly a per-component projection field.
8. **Story 8.6 must not build an orphan collector on `assetKeyReferenced`.** MEASURED at
   `folio-go/component_commands.go:790-799`: it matches `ElementImage` with a set, non-null `Asset`
   **only**. A font asset is referenced from the `fonts` chain map, never from an element, so it
   returns **false for every font asset**. Reusing it for AC5 deletes live faces with no compile error
   to announce it. That is DW-80, and Story 8.6's `Covers:` line already names it.
9. **The 64-asset ceiling is now guarded** (Story 8.4f, MEASURED at `verify-offline-release.mjs:51`
   against `maximumCacheAssets = 64` in `release-payload.ts:33`). Per D-8.5.1, raising it for a bundled
   catalogue is legitimate — with **stated headroom** rather than tuning to fit — and only now that the
   silent-failure half has landed.
10. **Fixed-point discipline in the new UI work.** 12.5's drag and 13.2's zoom both compute positions.
    The stored value stays millipoints (AD-2); a browser-side float is fine for a *transient proposal*
    under AD-15, and is never what gets committed.

### Open questions the lead can already see

**(a) Story 8.5 re-blocks unless `spec-fonts/SPEC.md` is amended first.** A stale halt file blocked on
`SPEC.md:130` — *"Which families make the shipped catalogue, how many, and who curates the list?"* —
recorded as open in a list where every settled question is struck through in place. **D-8.5.3 answered
it but the SPEC still said open.** By D-8.4.31's own rule this ruling had not happened where it gets
read. *(Discharged by D-000.10.)*

**(b) DW-108: `epic-8-context.md` lost the constraints Stories 8.5 and 8.6 were to be built against** —
the out-of-scope list (bold/italic meaning, synthetic emboldening/obliquing, variable-font axes, live
font services, enumerating host-installed fonts, CJK families), the procurement rules and the
chain-entry shape contract. All six survive verbatim in `spec-fonts/SPEC.md` `## Non-goals`, a stronger
source than the regenerable cache — **but the agent that builds 8.5 reads the cache, not the kernel.**
Discharge: confirm 8.5's spec carries them, **sourced from `spec-fonts/SPEC.md`**. A plan-gate check.

**(c) Three spec-fonts questions remain genuinely open; two are reachable this run.** *Does the licence
record live inline on each font asset, or in one document-level notice block?* — Story 8.6 embeds a
catalogue face into a `.folio`, so this becomes answerable-or-forced at 8.6. *May an author embed a
font file from their own disk?* — a Non-goal-adjacent scope question 8.6 must decline **explicitly
rather than by silence**. The third is bold/italic, deferred by D-000.7 to Epic 11's gate.

**(d) DW-68 is a fourth before-the-tag obligation that D-7.8.3's set does not name.** An aggregate-only
over-tall keep-together group is still clipped with a `TABLE_ROW_CLIPPED_HEIGHT` Warning, on a reason
Story 7.7 imported from Story 4.6 without importing its argument. Its register entry names the owner as
**"Owner / engineering lead — a RULING, before the `folio-go/v0.1.0` tag"** and says the deadline is the
whole point of the entry. Making it fatal narrows what renders: free before the tag, ruinous after.
**To the owner at Story 15.3's gate at the latest**, flagged now so it is not discovered at the tag.

**(e) Decisions already expected to escalate** — all written into the epics as owner rulings, not the
lead's additions:
- **12.4** — padding insets a non-table element, or the format documents it table-only. Both branches
  are direction-shaped: table-only **narrows the exported command layer** (joins the before-the-tag
  set) and amends `folio-format.md`; insets adds rendering behaviour to four element kinds. **Escalate.**
- **14.7's unit ruling** — mm vs pt, **product-wide**, not the dialog's. **Escalate.** (The stored value
  is millipoints either way; that half is not in question.)
- **11's realize-vs-retire** — D-000.7, at Epic 11's gate. **Escalate by construction.**
- **15.3's "which epics are inside v0.1.0"** — the owner's recorded decision to cut after Epic 6 is
  overtaken by Epics 7–14, and the AC says explicitly that such a decision is **re-made, not assumed**.
  **Escalate.**
- **14.7's Cancel/Apply transaction model** — expected to be **ruled**, not escalated: AD-15 forbids a
  second document model, so a modal Cancel needs a local uncommitted buffer the architecture already
  prohibits. Unless the owner wants the buffer ruled in deliberately, the direction determines it.
- **14.6's pick-binds-immediately vs explicit commit**, and **13.3's "Matches native render" wording** —
  both **ruled**. 13.3's answer is already written in the epic: the tab cannot compare itself to a
  native render; what is true is a property of the build proven by CI's matrix, and the wording must
  say the true thing.

**(f) The browser has never executed in this repository, and Epics 13 and 14 are 15 stories of designer
UI.** MEASURED: `.github/workflows/` contains **zero** occurrences of `playwright`; CI runs
`tsc --noEmit` only (DW-101, DW-103). Twelve real spec files exist, run locally, and are believed absent
by everyone. D-8.4.25(d)'s executed-browser assertion for Epic 8 remains **owed, not
attempted-and-passed** — the 162 MB Chrome for Testing archive stalls at a 428 KB cache entry
(D-8.4.35d). This is not a story-level problem; it is that the run is about to build fifteen
browser-shaped stories whose only real proof does not run. **Story 15.2 is the natural home and is
sequenced early (S3), before Epics 12/13/14.** If the browser cannot be made to run, every Epic 13/14
acceptance criterion about what a *user sees* is proved by unit tests over components, and the run
should know that going in rather than discover it at Epic 13's gate.

### Record found contradictory or stale (each measured)

1. **DW-24 is CLOSED**, and three live documents said otherwise — including this run's own D-000.5.
   Corrected by **D-000.9**; the Epic 9/10 reconstruction must **not** be scoped to close it.
2. **`sprint-status.yaml` recorded `epic-8: backlog`** while eleven of its stories are `done`.
   Corrected by **D-000.10**.
3. **Story 8.6's AC5 orphan-drop contradicts a forced invariant** and is still unresolved at HEAD. AC5
   says an unreferenced font asset *"is dropped"* on save; `serialize.go` preserves orphans
   unconditionally because AD-9's P1 forces it, and its own comment says collecting orphans is a
   designer **command**, never a serializer side effect. D-5.13.3's image-asset precedent gives the
   correct shape. **The story must route the drop through a command**, and `spec-fonts/font-catalogue.md`
   is amended in the same story. To be ruled at 8.6's plan gate rather than discovered by the build.
4. **`epics.md`'s Epic 8 prose dangled** on "a face-inventory decision SPEC-fonts leaves open"; Epic 11
   owns it (FR57). Corrected by **D-000.10**. Cosmetic, but it is exactly how a settled question reads
   as open to the next dispatch.
5. **The `~9 MB` first-load figure in `epics.md` is superseded but deliberately not yet replaced.**
   Story 8.4d is sequenced **last** in Epic 8 (D-8.4.27d) precisely so the threshold is set once against
   the epic's finished weight. **Until 8.4d lands, no story may fix the budget** — 8.5's obligation is
   to *record* per-asset Brotli weight, and 8.4d consumes that same figure rather than building it
   twice (D-8.5.6). `epics.md` must **not** be rewritten to 12.4 MB (D-8.4.24).
6. **The determinism scare is settled and must not be reopened casually.** D-8.5.7: two builds at one
   commit with provably identical tree state produce a byte-identical wasm; **one untracked file changes
   it**, because `go build` stamps `vcs.modified` and Go derives it from `git status`. **This pipeline
   writes untracked files into the tree** — halt files, result files, spec files — so the instrument
   perturbs the specimen. Story 8.4g fixed it with `-buildvcs=false`. **Any byte figure recorded this
   run must carry its exact invocation, commit and tree state.**
7. **One process defect is at instance three and is out of this repository to fix** (D-8.5.9): a step-03
   subagent committing on its own. Three unauthorized commits, three catches by a downstream agent that
   happened to audit provenance. At instance three the commit landed while `origin/main` was live, so an
   unaudited step-03 commit is one push away from being public. Recorded with its count so a fourth
   instance is priced against three priors rather than met fresh.

---

## Standing decisions (settled at setup, 2026-09-02)

These shape the whole run and were answered by the owner at the terminal before any story was
dispatched. Each is an **Owner decision**.

### D-000.1 — Scope: all of Epics 8–15, nothing deferred out of the run

**Verdict.** The run delivers every open item in Epics 8 through 15 — Epic 8's `8.5`, `8.6` and
`8.4d`; the retroactive closure of `9.1`, `9.2` and `10.1`; `11.1`–`11.3`; `12.1`–`12.5`;
`13.1`–`13.5`; `14.1`–`14.10`; and `15.2`, `15.2a`, `15.3`. Roughly thirty items across eight epics,
ending at a cut `folio-go/v0.1.0` tag.

**Situation.** The tracker holds the epics in three different states. Epic 8 is part-built (eleven
stories `done`, three `backlog`). Epics 9 and 10 hold shipped, committed code sitting at `review`
with no story files behind it. Epics 11–15 are entirely `backlog`, and Epic 11 additionally carries
a forward dependency on a decision that no artifact records. The owner was offered a shorter run —
finish Epic 8 and reassess, or skip the blocked epic — and chose the whole program.

**Options considered.** (a) *Epic 8 only* — a clean four-story run ending at a natural boundary,
rejected because it defers the release-blocker set with no gain other than an earlier stopping point.
(b) *Epics 8–15 minus Epic 11* — avoids forcing the bold/italic ruling, rejected because Epic 11 is
sequenced last anyway (see D-000.3), so the ruling is not needed until late and skipping the epic
would leave the designer shipping weight and slope toggles that do nothing. (c) *All of 8–15* —
chosen.

**Why this wins.** The sequence recorded in `sprint-status.yaml` was designed as one program, not as
eight independent epics; its dependencies (`14.8` after `12.3`, `14.10` after `14.9`, `15.3` last by
construction) only pay off if the whole thing runs. The accepted cost is a long run with a late
first release.

**Consequences.** `15.3` cannot be reported complete until every epic ruled inside `v0.1.0` has
landed. Two items already carry a before-the-tag deadline — `7.10` (landed) and `15.2a` (D-7.8.3) —
and both fall inside this run.

### D-000.2 — Answer channel: the terminal

**Verdict.** Every owner escalation this run is asked inline with `AskUserQuestion`, not over
Telegram, for the run's whole duration unless the owner says otherwise mid-run.

**Situation.** `~/.claude/telegram.env` is configured, so the run could have asked from the owner's
phone. The owner said they are staying at the keyboard.

**Consequences.** Escalations block the run rather than parking it, so the engineering lead's
bucketing matters more: anything it can rule on, it should, and only genuinely under-determined
questions reach the owner.

### D-000.3 — Build order: the recorded S1–S7 sprint sequence, not numeric order

**Verdict.** Stories are driven in the dependency-driven order recorded in `sprint-status.yaml` on
2026-08-30, adjusted for what has since landed. Concretely: finish Epic 8 → close Epics 9 and 10 →
`13.1`, `13.2`, `13.4` → `15.2` → Epics 12 and 14 → `13.3`, `13.5` → Epic 11 → `15.2a` → `15.3`.

**Situation.** Numeric order and the agreed sequence disagree in three places that matter. Epic 11 is
the only remaining epic that can move a golden hash, and the sequence puts it late so it lands on a
settled baseline rather than under one. `14.8` presents a table's header and alternating rows, but
`12.3` is what makes them authorable — building `14.8` first would force it to invent storage `12.3`
already owns. `14.10` lets an author click a table column, which requires `14.9` to have taught the
canvas to draw one. `15.3` freezes the public API and so must follow every epic ruled inside the tag.

**In simple terms.** The numbers are the order the epics were *written down*, not the order they can
be *built*. Painting the walls is job 14 and hanging the door is job 12, but you still hang the door
first. The recorded sequence is somebody having already worked that out once; re-deriving it per
story would reach the same answer more expensively and less consistently.

**Options considered.** (a) *Strict numeric order* — simpler to narrate and to audit, rejected on the
three concrete inversions above; each would surface as a mid-story halt that costs a re-dispatch.
(b) *The recorded S1–S7 sequence* — chosen.

**Why this wins.** The sequence is an existing, reasoned artifact with its rationale written beside
it, and the alternative's simplicity is cosmetic. The accepted cost is that the run's progress does
not read as a rising epic number, so the reports must state the sequence position explicitly.

**Consequences.** Every dispatch names its story id explicitly; nothing auto-picks the next item.
Epic keys go `done` out of numeric order, and `sprint-status.yaml` must not be read as a progress bar.

**How we'd know it was wrong.** A story halting on a dependency the sequence claimed was already
satisfied — that would mean the sequence has drifted from the code and needs re-deriving, not
patching.

### D-000.4 — Heavy-test cadence: per-epic

**Verdict.** Integration and e2e suites — Playwright, the cross-target × N-document matrix, anything
needing a container or fixture boot — run once at each epic boundary, not per story. Unit tests, lint
and build run in every story's `## Verification` regardless, and no story reaches `done` on a broken
unit suite. Every story under this cadence additionally gets a **compile-only** check of any new
integration/e2e package under its build tag.

**Situation.** This run is ~30 stories. The project's heavy suites are genuinely slow — a
four-target matrix over sixteen documents plus 23 Playwright scenarios — and paying that thirty times
would dominate the run's wall-clock. The trade is feedback latency against total time.

**In simple terms.** Running the slow suite every story is like re-proofreading the whole book after
each sentence: you find the typo the instant you make it, but you never finish the book. Once per
chapter still narrows any error to a chapter you can re-read.

**Options considered.** (a) *Every story* — a break is attributable to exactly one commit, rejected
on wall-clock for a run this size. (b) *End of run* — fastest, rejected because a failure would then
span eight epics and the corpus-hash guards make that the most expensive thing in this project to
bisect. (c) *Per epic* — chosen; the project already has per-epic boundary gates (`epic-N-boundary-gate.md`)
so the machinery and the habit both exist.

**Why this wins.** It matches a boundary the project already treats as meaningful. The accepted cost
is that a heavy-test break spans that epic's stories, and the catch-up run is a hard gate — no epic
is marked `done` before it passes.

**Consequences.** The compile-only check is not optional: a package that fails to compile under its
integration build tag silently skips its tests, which would let a catch-up run open with a compile
error from ten stories ago. Every Delivery Log entry must name the suites it did *not* run and when
they are due. A story whose own correctness is integration-shaped — schema or format changes, the
byte-identity regime, anything that can move a golden hash — gets its integration commands in its own
`## Verification` anyway, and the override is logged with its reason.

**How we'd know it was wrong.** An epic boundary gate going red on something a story-level unit suite
could plausibly have caught — that would mean the unit coverage is thinner than the cadence assumes.

### D-000.5 — Epics 9 and 10 are reconstructed and closed, not accepted as-is

**Verdict.** Stories `9.1`, `9.2` and `10.1` get spec files written retroactively from their shipped
diffs, a real review pass against what is actually in `main`, and the gates run for real, before they
are marked `done`. DW-24's uncovered branches are closed as part of that work.

**Situation.** Those three stories shipped code in the same session that filed their epics. There is
no story file, no acceptance criteria, no review and no retrospective behind any of them — the only
audit that has ever touched them is Story 15.1's, which traced the goldens they moved and re-recorded
four of them. One of the two commits also created `folio-go/text_alignment.go` with no story, no AC
and no FR at all, and two of its branches are declared by no fixture (DW-24).

**In simple terms.** The code is in the building and the lights are on, but nobody signed the work
off and there is no drawing of what was built. `15.3` cuts a release tag over this. Tagging is the
moment that stops being an internal untidiness and becomes something shipped to other people.

**Options considered.** (a) *Mark done as-is* — leans on Story 15.1's audit and moves on; rejected
because that audit answered "which commit moved the goldens", not "is this code correct", and it
explicitly filed DW-24 rather than closing it. (b) *Leave at `review` and skip* — rejected: it puts
unreviewed code inside `v0.1.0` and defers the problem past the only gate that would have caught it.
(c) *Reconstruct and close* — chosen by the owner.

**Why this wins.** It is the only option under which the tag means what a tag is supposed to mean.
The accepted cost is three extra stories of work that produce no new behaviour — retroactive specs
are unrewarding to write and their value is entirely in what they let a later reader verify.

**Consequences.** These three specs describe code that already exists, so their `<intent-contract>`
is written against the shipped diff and their implementation phase may be empty or near-empty; the
review and verification phases are the substance. DW-24 closes here or is explicitly re-deferred with
an owner. Neither epic may be marked `done` before its boundary catch-up run passes.

**How we'd know it was wrong.** A reconstructed spec that cannot be written without inventing intent
the code does not evidence — that would mean the code needs a decision, not a description, and the
story becomes a real one routed to the lead.

### D-000.6 — Plan gate is silent; the lead and the orchestrator hold it

**Verdict.** Specs are gated by the orchestrator and the engineering lead without the owner reading
them. The owner sees each story's plain-terms opener in its report, and is asked only when the lead
returns `OWNER-DECISION-NEEDED`.

**Situation.** The plan gate inspects each spec before code is derived from it — the intent contract
against the epic, the verification commands against the cadence, the scope fence. The owner was
offered sight of every spec, or of load-bearing specs only, and chose neither.

**Consequences.** The orchestrator's backstop stands regardless: anything load-bearing that the lead
*rules* on — a security boundary, the byte-identity regime, the public API surface, or a scope cut
with lasting consequences — is escalated to the owner anyway rather than silently accepted.

### D-000.7 — Epic 11's realize-vs-retire ruling is deferred to its sequence position

**Verdict.** The owner is not asked now whether bold and italic get realized as shipped faces or
whether the toggles are retired. The question is put at Epic 11's plan gate, at sequence position S6.

**Situation.** `sprint-status.yaml` records Epic 11 as blocked on this ruling, and SPEC-fonts records
the same question as open. The epic text specifies only the *realize* branch; the *retire* branch
replaces `11.1`–`11.2` with a single removal story. Asking now would settle it, but D-000.3 puts
Epic 11 last, so an answer given today would sit unused across seven epics of work that may change
what the right answer is.

**Why this wins.** The decision is cheaper and better-informed at its own gate — by then Epic 8's
font work is fully landed and the face-inventory question the spec leaves open has real shipped
faces to reason about. The accepted cost is that Epic 11's shape is unknown until late in the run,
so its story count cannot be reported with confidence until then.

**Consequences.** This run's report must not claim Epic 11 is three stories; it is three *or* one.
The question is asked at the terminal per D-000.2 before any `11.x` spec is written.

### D-000.8 — Do not pause between stories or epics; escalate only direction-changing decisions

**Owner decision**, given directly at 2026-09-02 after setup: *"make sure the agents don't stop
between stories and epic unless there's any important decision to make that could change the project
direction."*

**Verdict.** The run does not stop at story boundaries or at epic boundaries. No approval checkpoint,
no "shall I continue", no pause to report. Reports are still written — they are just written *past*,
not *at*, a stopping point. The single reason to stop is a decision that could change the project's
direction, and that bar is now the explicit test applied to every escalation.

**Situation.** The setup answer already chose run-continuously over per-epic and per-story
checkpoints. This instruction goes further: it removes the epic boundary as a pause point too, and it
restates the escalation bar in the owner's own words. Without it, the natural failure mode of a long
orchestrated run is to treat each epic summary as an implicit request for permission and idle there.

**In simple terms.** The owner is not a gate; they are a reader. Handing them a summary is not the
same as waiting for them to pick it up. The only thing worth their attention mid-run is a fork where
picking wrong sends the project somewhere they did not intend to go.

**What clears the bar** (stop and ask): a scope cut or expansion that changes what `v0.1.0` is; a
change to the byte-identity regime, the `.folio` format, or the public API surface; the Epic 11
realize-vs-retire fork (D-000.7); anything on a security boundary; a decision that forecloses a
direction the epics assume is open. **What does not** (rule it and keep going): shape, naming, file
placement, which existing seam to reuse, test structure, how a spec is worded, which of two equally
compliant implementations to pick, and anything the ADRs or the two decision logs already settle.

**Why this wins.** It converts the engineering lead from an advisor into the run's actual decision
authority for everything below the direction bar, which is what a lead grounded once in the
architecture is for. The accepted cost is that the owner will see decisions after they are applied
rather than before, so the log has to be good enough to reverse one cleanly — which is why every
ruling is written to the standard in this file rather than as a one-line note.

**Consequences.** Epic boundaries still run their heavy-test catch-up gate (D-000.4) and still get a
written summary, but the run proceeds straight into the next epic without waiting. Any ruling the
owner disagrees with on reading is reversed by appending a reversal here, never by rewriting the
original entry. The lead is told this bar verbatim in its grounding prompt.

**How we'd know it was wrong.** The owner reversing a ruling that was applied without asking — one
reversal is the system working, a pattern of them means the bar is set too high and more should be
escalated.

---

## Run decisions

*(appended per story as they are made)*

### D-000.9 — CORRECTION to D-000.5: DW-24 was already closed, and three documents said otherwise

**Orchestrator decision** (a correction, filed 2026-09-02 at the engineering lead's grounding). The
original D-000.5 is kept above exactly as written; this entry corrects one clause of it. Nothing is
rewritten in place.

**Verdict.** **DW-24 is CLOSED** — by Story 7.3 on 2026-08-30, with a re-derived-by-grep enumeration
of all six rounded alignment branches, a per-site falsifier for each, and a
`fixtures/alignment-rounding/` fixture registered at every golden surface. The Epic 9 and Epic 10
reconstruction stories must therefore **NOT** be scoped to close it. D-000.5's sentences *"DW-24's
uncovered branches are closed as part of that work"* and *"DW-24 closes here or is explicitly
re-deferred with an owner"* are **false** and are withdrawn. Every other clause of D-000.5 stands.

**Situation.** DW-24 recorded that the `align: "center"` and `valign: "middle"` rounding branches
created by Epic 9's unstoried `text_alignment.go` were declared by no fixture, so the feature's only
rounding sites had zero golden coverage. Story 7.3 closed it three days ago. But the closure never
propagated: `deferred-work.md` carries the CLOSED header, while `sprint-status.yaml` still read *"DW-24
open, owner Story 7.3"*, the engineering lead's own run brief listed it as open work for this run,
and **D-000.5 — written by me at this run's setup, hours ago — repeated the stale reading and made
closing it a scope item of three stories.** Three documents, all stale by three days, all agreeing
with each other and all wrong.

**In simple terms.** A job was finished, the person who finished it ticked their own list, and nobody
told the three other lists that also tracked it. Then a new run read two of those lists, believed
them, and wrote the finished job into its plan as work still to do. The failure is not that a
document was wrong; it is that a fact had four homes and only one of them was updated.

**How it was caught.** The lead was told to verify what it read rather than accept it, and it read
`deferred-work.md`'s entry in full instead of trusting my brief's summary of it. That is the whole
mechanism — the correction cost one paragraph because it was found at grounding, before any story was
scoped against it.

**Why this matters more than a stale line usually would.** Had it survived to Epic 9's plan gate, the
reconstruction story would have been written with an acceptance criterion to cover a population that
is already covered. It would then have done one of two things, both bad: nothing at all, reported as
a pass; or manufactured a *second* fixture over the same six branches, which is worse than nothing —
it is a duplicate witness that future maintenance must keep in step with the real one, with no test
that would notice them diverging.

**Options considered.** (a) *Edit D-000.5 in place* — tidier to read, rejected because this log's own
rule is that reversals are appended and originals are kept; an edited-in-place entry destroys the
evidence that a stale fact propagated into a plan, which is the only genuinely instructive part.
(b) *Correct only `sprint-status.yaml` and move on* — rejected: the tracker was the smaller of the
two wrong sources, and leaving D-000.5 uncorrected leaves the scope error live in the document the
Epic 9/10 dispatches will actually read. (c) *Append this correction and fix the tracker* — chosen.

**Consequences.** The Epic 9/10 reconstruction stories are scoped to what they are actually for:
writing specs retroactively from the shipped diffs, reviewing code that was never reviewed, and
running the gates for real. Coverage of the alignment rounding branches is **already evidence they
can cite**, not work they must do. `sprint-status.yaml`'s DW-24 line is corrected in place with the
original kept beside it.

**How we'd know it was wrong.** `deferred-work.md`'s DW-24 closure failing its own falsifiers when
re-run at Epic 9's gate — that would mean the entry is marked closed but the fixture does not bite,
which is a worse defect than the one DW-24 originally recorded.

### D-000.10 — Three stale documents amended before Story 8.5 dispatches

**Orchestrator decision** (routine — each is propagating a ruling that already exists to the document
that gets read).

**Verdict.** Three amendments made at 2026-09-02, none of which decides anything new:

1. **`spec-fonts/SPEC.md`'s Open Questions** — *"Which families make the shipped catalogue, how many,
   and who curates the list as it changes?"* struck through and answered from **D-8.5.3**: 20+
   families, admitted by a named permissive allowlist (OFL-1.1, Apache-2.0, MIT, UFL), enforced the
   way AD-26's dependency ban is — fail the build, never warn, unclassifiable treated as failure.
2. **`sprint-status.yaml`** — `epic-8` moved from `backlog` to `in-progress`, which is what it has
   been for eleven `done` stories; and its stale `DW-24 open` note corrected per D-000.9.
3. **`epics.md`'s Epic 8 prose** — the dangling *"a face-inventory decision SPEC-fonts leaves open"*
   now points at **Epic 11 (FR57)**, which owns the realize-or-retire decision.

**Situation.** Item 1 is the expensive one. A previous Story 8.5 dispatch **already halted** on that
exact line, blocked on an intent gap: the question was recorded as open in a list where every settled
question is struck through *in place*, so the absence of a strike-through reads as a live question.
D-8.5.3 answered it — as an **owner decision**, no less — but answered it in the decision log, and the
agent that plans Story 8.5 reads the SPEC. This is **D-8.4.31's rule biting a second time**: a ruling
recorded everywhere except the document that gets read is a ruling that did not happen.

**In simple terms.** The sign on the door still said CLOSED after the shop reopened, so the next
customer turned around — exactly as the previous one had. The shop was open both times. Fixing the
sign costs a minute; not fixing it costs every arrival.

**Why this wins.** Each amendment converts a recorded ruling into a form the next dispatch can act on,
at the cost of three small edits. The accepted cost is that `epics.md` and the SPEC now carry decision
text that duplicates the log — mild redundancy, deliberately chosen over a second halt.

**Consequences.** Story 8.5's dispatch should no longer block on the catalogue question. **Note for
the rest of the run:** two spec-fonts questions remain genuinely open and both are reachable at Story
8.6 — whether the licence record lives inline on each font asset or in one document-level notice
block, and whether an author may embed a font from their own disk. Neither may be answered by silence;
8.6's spec either settles them or declines them explicitly.

**How we'd know it was wrong.** Another dispatch halting on a question one of the two decision logs has
already answered — that would mean the propagation check belongs at every plan gate, not just at the
ones where a halt already taught us to look.

## Story 8.5's plan gate — two rulings, one of which splits the story (2026-09-02)

### D-8.5.12 — The owner brief's "no first-load cost" clause is DECLINED on the contract, and the record says decline, not impossibility

**Engineering lead ruling**, taken by the orchestrator, at Story 8.5's plan gate. Measured at `b0fae96`.

**Verdict.** Story 8.5 ships the catalogue **bundled and precached**, as specced. The clause of
D-8.5.3's brief reading *"the engineering should find a shape where size is not paid for at first
load"* is **declined**. No document is amended, the story proceeds, and this entry is the record of
the decline together with its price and its reversal.

**Situation.** D-8.5.1(b) steered toward faces fetched at pick time. The plan resolved against that
steer on three pieces of contract text — `spec-fonts/SPEC.md` `## Non-goals` (*"No live font service.
No Google Fonts API, no arbitrary URL, no 'download on first use'"*), `font-catalogue.md` §"Why
bundled rather than fetched" (*"Nothing is fetched. The bytes come from the bundle already on the
machine"*), and the story's own AC3 (*"inside the bundle behind the same verified asset URLs … and
the offline verification job covers them"*) — plus the measured fact that there is nowhere to fetch
**from**: AC4 bans the Google hosts, the product ships no server, and NFR7 promises the designer works
offline. That resolution is correct and is not revisited.

**The correction that is the load-bearing half of this ruling.** The spec's Design Note 1 states that
no shape exists which avoids first-load cost. **That overstates it, and the lead corrected it:** no
such shape exists *that the Non-goals clause permits*. The mechanism is available and was checked — a
release can carry an asset the service worker does not precache, since `immutable` and the S1 row set
are separate mechanisms, and 8.4f's own measurement (23 of 64 slots used, 41 free) confirms the
ceiling is not what forbids it. **A same-origin, verified-URL, cached-on-first-pick tier is buildable
today.** It is foreclosed by *"no 'download on first use'"*, which names it literally. **This is
policy, not physics** — and the record must say so, because an owner cannot sensibly reverse a
decision they have been told is impossible.

**In simple terms.** The shop *can* deliver later instead of handing you everything at the door; the
house rules just say it does not. Telling the owner "that is impossible" would be false and would
quietly remove a choice that is actually theirs. Telling them "we chose not to, here is what it costs
and here is what changing it costs" leaves the choice where it belongs.

**The price of the decline, stated plainly because it is not what the owner priced.** The release
emits **13,490,909** Brotli bytes today across 22 sidecars. Roughly 20 Latin faces at 45–226 KB Brotli
each adds about **1.5–4.5 MB**, landing at **~15–18 MB** against a **`~9 MB`** stated commitment.
**That is not the 37% overage the owner was shown when they chose 20+ families. It is roughly
double.** The owner priced 37%; they have not yet priced 2×.

**Why ruling rather than escalating is correct here and not merely convenient.** Proceeding requires
**no amendment to anything** — every governing document already says bundled. Only the *reversal*
would need the owner, and nobody is proposing one. And the reversal stays cheap after the tag:
`folio-go/v0.1.0` versions the **Go module**, while the precache set is a **designer** artifact, so
moving catalogue faces to a deferred tier later breaks no downstream integrator; and Story 8.6 puts
the picked face **inside the `.folio`**, so no already-written document regresses. By D-8.2.2(b) this
is not on the exported surface and is not a narrowing, so it **does not join the before-the-tag set**.

**Options considered.** (a) *Build a middle tier now* — rejected: forbidden by a clause that names it
literally, and inventing it in the build would be a design change smuggled through an implementation.
(b) *Escalate to the owner now* — rejected as premature: the decision that actually needs them is the
**threshold**, and D-8.4.24 already placed that at Story 8.4d, deliberately, so it is chosen once
against finished weight rather than tuned along the way. (c) *Proceed as specced and record the
decline with its price and reversal* — chosen.

**Why this wins.** It keeps the owner's one real decision at the one gate designed to hold it, while
refusing to let the declined clause pass unrecorded. The accepted cost is that the owner learns a
clause of their own brief was declined **after** the story is planned rather than before.

**The reversal, named and priced so it is decidable.** Amend `SPEC.md` `## Non-goals` **and**
`font-catalogue.md` together to permit a same-origin deferred tier, accepting that a family not yet
fetched cannot be picked while offline. **The mitigant must be stated with it, or the trade reads
worse than it is:** the catalogue is a **palette, not coverage** — the three shipped Noto faces are
the coverage (D-8.5.1) — so an unfetched family degrades to *"you cannot pick that family right
now"*, never to a document that will not render.

**Guardrails.** The story invents **no** middle tier and does not reverse this in the build; a build
proposing one is a halt, not a design choice. AC5's per-asset Brotli figure must be a real
measurement, never a false zero — `s1VisibleBytes` sums four hardcoded needles and already misses
174,949 bytes of IBM Plex, so it is not the instrument (D-8.5.6). **Story 8.5's Delivery Log states
the catalogue's total added Brotli bytes as ONE number**, with its invocation, commit and tree state
(D-8.5.7), so 8.4d does not inherit twenty rows nobody added up.

**The scheduled escalation, held deliberately.** The ~2× figure belongs at **Story 8.4d**, which is
where D-8.4.24 put it. **If 8.4d's plan gate proposes a threshold near 18 MB, it is escalated to the
owner there** — doubling a stated NFR7 commitment is a different animal from 37%, and the owner priced
the latter. Recorded now so that arrives as a scheduled question rather than a surprise.

**How we'd know it was wrong.** The owner reads the decline and reverses it. Under D-000.8 one
reversal is the system working; a pattern of them means the escalation bar is set too high.

### D-8.5.13 — Story 8.5 SPLITS: the licence gate leaves as Story 8.4h, and it is a different story than the spec thought

**Engineering lead ruling**, taken by the orchestrator. Measured at HEAD, not inferred.

**Verdict.** Split. **Story 8.4h — "The licence gate closes"** is written and sequenced **immediately
before 8.5**. `8.4d` stays last in the epic (D-8.4.27d), unaffected.

- **8.4h takes:** AC2 (both fall-through sites), the `UFL` addition to `classify.go` (SPDX entry +
  marker branch + fixture — confirmed absent at HEAD), and AC7 (the widened extension-class guard
  population).
- **8.5 keeps:** AC1, AC3, AC4, AC5, AC6, AC8. **AC4 stays in 8.5 deliberately** — the forbidden-host
  scan is the assertion that D-8.5.12's decline was honoured, so it belongs with the thing it
  constrains, not with the licence gate.

**Why the `8.4x` key.** The series is this epic's insertion mechanism and **D-8.5.1(a) is the exact
precedent** — *"the SILENT-FAILURE half lands ahead of 8.5 (Story 8.4f)."* Same move, same story, same
scheme. Letters a–g are used; `8.4h` is next.

**The criterion, so this is not "a split too many."** This epic splits on **independent failure mode**,
never on story budget. 8.4f split because its failure was silent and reachable by any future asset
addition. This splits because **its failure is live at HEAD right now, with zero catalogue faces in
the tree.** Measured at `lint/internal/manifest/manifest.go:298-301`: a font whose licence text does
not classify gets `licenceLabel = "SEE NOTICE"` and `ResolveAssets` returns **nil**. **A GPL font
ships with a clean build today.** That is a **repair**, and a repair does not depend on the extension
that surfaced it — which is also why it may land ahead of the story that found it.

**In simple terms.** The lock on the door is broken right now, and it is broken whether or not we are
about to move twenty new things into the room. Fix the lock first, prove it holds against a key it
should refuse, and *then* move the things in. Fixing it afterwards means the twenty things were
carried past a door nobody had tested.

**The red-proof is ruled as a hard condition, not a preference.** A gate red-proved over today's six
OFL-1.1 faces in nine known directories says nothing about twenty faces in a new directory — D-8.5.4's
quantifier lesson, hit three times this run. So **8.4h's red-proof must be population-independent**: a
synthetic testdata font in a **new, otherwise-empty directory** carrying an **unclassifiable** licence,
*and* a second carrying a **classifiable-but-not-allowlisted** one (GPL-3.0). Each must red **on its
own message**, not on a neighbouring guard — D-8.5.8(b): when a proof will not fail for the right
reason, move the proof, never weaken the claim. Proved that way, any real face in any directory is
covered by construction, and **8.5's twenty faces are admitted by a gate that was proved rather than
one that was merely green when they arrived.** That is the entire value of the split and it is lost if
the red-proof is written against the existing population.

**THE TRAP THAT DECIDES 8.4h's SHAPE — measured, and absent from the spec.** The two sites the spec
names are **two populations under two policies, and they must not share one list.**

- **`manifest.go:~298` is the FONT path.** `ResolveAssets` (`:165`) walks the whole repo but filters to
  `fontExtensions`, keyed by directory. The owner's four-id allowlist governs here — D-8.5.3's own
  words are *"a **font** asset."* The current population classifies cleanly: nine directories, all
  OFL-1.1.
- **`manifest.go:~384` is the THAI WORDLIST path** — `wordlistAssetDir/words_th.txt`, PyThaiNLP,
  **CC0-1.0**, with its own dedicated marker branch in `classify.go` whose comment at `:81-83` records
  CC0 joining at Story 2.1 / D-2.1.3 as the first CC0 asset. **`CC0-1.0` is not one of the owner's four
  ids.** Making that site fail-closed against the *font* allowlist **reds the build on a shipped,
  legitimate, owner-unobjected asset on day one** — the DW-23 / DW-86 standing-red shape, manufactured
  by the very story that exists to stop silent passes.

> **Ruled: both sites close, against DIFFERENT lists.** The font site closes against the four-id
> **asset allowlist** (OFL-1.1, Apache-2.0, MIT, UFL). The wordlist site closes against the existing
> `permissiveSPDX` set, which already carries `CC0-1.0` deliberately. **Neither site keeps
> `"SEE NOTICE"` as a pass.** **Do not add CC0-1.0 to the four-id list** — that would amend an owner
> decision to fix a scoping error. This touches D-8.5.3 not at all: the owner ruled about **fonts**,
> and honouring that scope is exactly what keeps the dictionary out of it.

**In simple terms.** One phrase — `"SEE NOTICE"` — appears in two places that look identical and mean
different things: one guards typefaces, the other guards a Thai dictionary. Closing both against the
typeface rules would reject the dictionary for having a perfectly good licence that simply is not a
*font* licence. Same words, different rooms. **The build must not collapse them into one shared
constant because the strings match** — validate by consumer, not by key location.

**Two further guardrails on 8.4h.**
- **It must demonstrate the current committed population classifies cleanly in the same commit that
  makes classification fatal** — nine font directories (three under `folio-go/fonts/`, six under
  `folio-designer/public/fonts/`) plus the wordlist. **A gate that lands red is not a gate**; it is
  DW-23 starting again, a pattern this run has registered twice.
- **AC7 widens the guard's DIRECTORY reach, not its ASSET-CLASS reach.** `ResolveAssets` already walks
  the whole repo and filters by extension; the extension-class guard exists for the file whose
  extension is *not* in `fontExtensions` and is therefore invisible to both (D-8.5.2). 8.4h must
  **enumerate what the widened walk newly picks up** and confirm no non-font asset is newly subjected
  to the font allowlist. **Do not assume the widening is inert** — that assumption is precisely what
  produced the wordlist trap above.

**`oversized`, resolved by the split.** Do not additionally tighten now; re-check at 8.5's re-plan. If
8.5 is still over at six ACs, the lever is **AC1's per-face procurement table** — twenty provenance
rows belong in a companion artifact the spec cites, on `font-catalogue.md`'s own shape, not inlined in
the file the implementation subagent reads. **Do not resolve `oversized` by thinning acceptance
criteria**; that is moving the bar to fit the instrument.

**Consequences.** Epic 8's remaining sequence becomes **8.4h → 8.5 → 8.6 → 8.4d**. Story 8.5's current
spec at `ready-for-dev` is **superseded pending re-plan** and must not be dispatched to build as
written. `sprint-status.yaml` gains an `8-4h` key.

**How we'd know it was wrong.** 8.4h lands green and 8.5 then finds the gate does not see its faces —
that would mean AC7's widening was scoped by directory where it needed to be scoped by the tracked-file
population, and the fix belongs in 8.4h's population definition, not in 8.5.

## Story 8.4h's plan gate — a charter premise corrected, and the owner's fourth licence given its real name (2026-09-02)

### D-8.4h.1 — CORRECTION to D-8.5.13: the live hole is bigger than the ruling described, and the story is unchanged by it

**Orchestrator decision** (accepting the plan gate's measured correction). D-8.5.13 stands; this
entry corrects its mechanism description. Nothing is rewritten in place.

**Verdict.** D-8.5.13 — and the commit message and the owner report that repeated it — described the
licence hole as the `"SEE NOTICE"` fall-through: a font whose licence text *cannot* be classified
passes. **Measured at `7aa283b`, that is only half of it.** Both asset paths call the classifier, keep
only its SPDX string, and **discard its family verdict**. So a GPL-3.0 licence text classifies
*correctly* — as `(copyleft, "GPL-3.0")` — and still produces a clean manifest row and a clean build.
It never reaches the fall-through at all. **Two distinct holes at the same two lines.**

**Situation.** The ruling's acceptance text already says *"outside the list **or** unclassifiable"*, so
the story's scope, its ACs and its red-proof requirement are all unaffected — the second hole was
already inside what 8.4h must close. What was wrong was the *explanation* of how the failure happens,
which had propagated into three documents.

**In simple terms.** We reported the lock as "fails to catch anything it doesn't recognise". The truth
is worse: it also fails to catch things it recognises perfectly well and knows it must refuse. It
reads the label, correctly says "this is one we cannot accept", writes that on the form, and opens the
door anyway. The story already had to fix both; only our account of it was wrong.

**Why this is a correction and not a defect in the plan.** D-8.5.10 draws the line this run keeps
needing: moving a threshold or an AC to match a measurement is the defect; **correcting a false
premise about how the bar is reached is legitimate**, and the two look identical in a diff. Here
nothing moved — the bar is where it was, and the measurement made the reason for it sharper. The plan
gate corrected it in Design Notes and continued rather than halting, which is the right call for a
premise that does not change the work.

**A second premise corrected in the same pass.** D-8.5.13 said the current population is *"nine
directories, all OFL-1.1"*. Measured, it is **eleven** font directories plus the wordlist, and **not
all OFL-1.1** — `folio-go/testdata/fonts/Roboto-Regular.ttf` is **Apache-2.0**. Apache-2.0 is on the
owner's allowlist, so the gate still lands green in the same commit that makes classification fatal,
which is the hard condition that mattered. AC6 is written over eleven, not nine. This is D-8.5.4's
count-written-beside-the-thing-it-counts lesson at its fourth instance this epic.

**Consequences.** The user-facing account of this defect is corrected: an unacceptable font passes
today **whether or not the classifier can read its licence**. AC6's population is eleven directories
plus the wordlist. `manifest_test.go:183` writes a tracked fixture licence of literally `"a licence"`,
which goes red the moment classification becomes fatal — repaired by its own numbered task rather than
discovered by the build.

**How we'd know it was wrong.** The build closing only the fall-through and leaving the discarded
family verdict in place — the GPL red-proof would then stay green, which is exactly the arm D-8.5.13
required be population-independent.

### D-8.4h.2 — The owner's fourth licence is `Ubuntu-font-1.0`; `"UFL"` is shorthand, not an identifier

**Orchestrator decision** (routine — a canonicalization inside an owner decision, ruled rather than
escalated because it changes which string is typed, not which licence is permitted).

**Verdict.** The allowlist's fourth member is implemented as the SPDX identifier
**`Ubuntu-font-1.0`**. The bare alias `"UFL"` is deliberately **not** accepted as an identifier.

**Situation.** D-8.5.3 records the owner's permissive allowlist as *"OFL-1.1, Apache-2.0, MIT, UFL"*.
The first three are canonical SPDX ids; the fourth is not an SPDX identifier at all. The plan gate
verified this by fetch rather than by memory: `Ubuntu-font-1.0` → HTTP 200 ("Ubuntu Font Licence
v1.0"); `Ufont-1.0` and `UFL-1.0` → 404.

**In simple terms.** The owner named four licences and wrote the fourth the way people say it out
loud. The registry the code looks it up in only answers to its formal name. Typing the nickname into
an exact-match lookup produces a key nothing will ever match — an entry that looks live and is dead.

**Why it is safe to rule.** The owner's decision was about *which licences are acceptable*, and that
is untouched: the Ubuntu Font Licence is permitted either way. Only the spelling changes, and it
changes to the one the mechanism can actually use. The project's own convention decides it — every
existing entry in the classifier is canonical, and the lookup is an exact match — so accepting the
alias would be the anomaly.

**Options considered.** (a) *Accept `"UFL"` literally* — faithful to the owner's text, rejected
because `classifyBySPDX` is an exact lookup, so the key would be dead code that reads as live
coverage; the allowlist would appear to have four members and enforce three. (b) *Accept both the
canonical id and the alias* — rejected: a silent alias list is D-2.1.3's rotting-list shape, and it
buys nothing since no upstream licence file declares itself as "UFL". (c) *Escalate the wording to the
owner* — rejected as disproportionate under D-000.8; nothing about which fonts may ship changes.
(d) *Canonical id only* — chosen.

**Why this wins.** It is the only reading under which the owner's four-member allowlist actually has
four working members. The accepted cost is that the decision log and the code now spell the same
licence two ways, which this entry exists to reconcile.

**Consequences.** `classify.go` gains `Ubuntu-font-1.0` as a map entry, a marker branch and a
`Family` mapping, with a fixture and a lookalike-negative test — added loudly, never by widening a
list to make a population pass. Anyone reading D-8.5.3 later should read `UFL` as naming the Ubuntu
Font Licence, whose identifier is `Ubuntu-font-1.0`.

**How we'd know it was wrong.** A real upstream font shipping a licence file that declares `UFL` and
nothing else — then the alias is load-bearing after all and joins the map as an explicit, tested alias
rather than as a silent one.

### D-8.4h.3 — `oversized` and `multiple-goals` are both accepted on this story, with reasons

**Orchestrator decision** (routine).

**Verdict.** Story 8.4h ships at 652 lines carrying both warnings. Neither is resolved by trimming.

**`multiple-goals`** is an honest signal and is **accepted**: AC7 (the extension-class guard's widened
directory reach) is separably shippable from the licence repair. It stays here because D-8.5.13
assigned it here, and because it is the guard that reports a font the gate cannot *see* — the same
blind spot, one layer out. Splitting a two-line widening into its own story would cost more record
than it saves, and the plan gate measured the widening **inert today**: over 1434 tracked files, zero
non-`.ttf/.otf/.ttc` font-plausible extensions and zero font-magic hits. That measurement is the
discharge of D-8.5.13's "do not assume the widening is inert" condition — it is now known inert rather
than presumed so.

**`oversized`** is **accepted** rather than trimmed. The bulk is Code Map and Design Notes — measured
evidence about two fall-through sites, the classifier's real behaviour, and which existing tests move
or red. That is exactly the material that stops the implementation subagent re-deriving a wrong
premise, which this story has already had corrected once (D-8.4h.1). Per D-8.5.13's own guardrail,
acceptance criteria are never thinned to clear this warning — that is moving the bar to fit the
instrument. Re-checked at Story 8.5's re-plan, where the lever is AC1's procurement table rather than
anything here.

**How we'd know it was wrong.** The implementation subagent losing a constraint stated only late in
the spec — the signal is a review finding that contradicts something the spec says plainly.

## Story 8.4h's close — a follow-up review judged unnecessary, and a process rule proved unenforceable (2026-09-02)

### D-8.4h.4 — `followup_review_recommended: true` is discharged WITHOUT a second review pass, and the reason is on the record

**Orchestrator decision.** BMAD raised the flag mechanically: patched findings scored `3×1 + 1×5 = 8`,
over its threshold of 5. The flag is a gate, and this entry is the gate being answered rather than
ignored.

**Verdict.** No second review dispatch. Story 8.4h closes on the review it had, and the **closer runs
the hard adversarial pass** instead — which this story earns on its own merits, not as a substitute.

**Situation.** Triage was **patch 6** (1 medium, 5 low), **deferred 7** (3 medium, 4 low), **rejected
7**, with **zero** `intent_gap` and **zero** `bad_spec`. No `high` at any point. The flag's arithmetic
counts *how much was patched*, not *how well* — six patches on a story with no high findings and no
spec disagreement is a review that worked, not one that needs repeating.

**Why a second pass would buy nothing here, stated as evidence rather than as confidence.** The
review's one material finding was verified by **deletion mutation**, which is the strictly stronger
instrument: the wordlist site's third refusal arm shipped **unexercised**, and deleting
`case !licence.IsPermissiveSPDX(wordlistSPDX)` outright left all four lint packages **green**. It was
missed originally because substitution mutations *did* redden — falsifying a condition proves arm
**order**, deleting it proves the arm is **reached**. That distinction is the actual lesson of this
story and it is now covered by a named `CC-BY-SA-4.0` subtest, re-proved by deleting the arm again.

Every other guard was mutation-proved the same way, with **disjoint** red sets: unclassifiable,
copyleft and off-allowlist each redden their own named test **and nothing else**; the classifier's two
mutations redden two non-overlapping test sets. And the **forbidden collapse was mutated
deliberately** — pointing the wordlist site at the font allowlist reds the named
`CC0-1.0_passes_—_it_is_NOT_one_of_the_four_font_ids` subtest, which is D-8.5.13's trap caught by a
test rather than by the whole build going red on a shipped asset.

**In simple terms.** The flag says "a lot got fixed, look again." What actually happened is that the
review found one real hole, and the hole was found by the one technique that could find it. Running
the same layers a second time would re-ask a question already answered with a better instrument.

**Options considered.** (a) *Set the spec to `in-review` and re-dispatch* — rejected: it re-runs the
same four layers over a change set they have already read, and the run's budget is better spent on the
adversarial pass the licence boundary actually warrants. (b) *Ignore the flag* — rejected outright;
it is the one signal BMAD gives for free, and silently dropping it is how a `high` gets lost in a
later story. (c) *Discharge it in writing and route the scrutiny to the closer* — chosen.

**Why this wins.** It spends the scrutiny where the risk is. The accepted cost is that the judgement
is mine: if a later story finds a defect in this change set that a second review layer would plausibly
have caught, this entry is the record of who decided otherwise.

**Consequences.** The closer runs the **hard adversarial pass** for this story — least-privilege,
TOCTOU/double-execution on the gate, and specifically whether any refusal can be reached with a
*message* that leaks a path or licence text it should not. **Seven deferrals are live and must be
registered with owners**, not left in the spec; deferral 1 (the unanchored marker branch, medium) is
explicitly its own story because correcting it touches the pre-existing OFL branch too.

**How we'd know it was wrong.** A later story finding a defect in the licence gate that one of the four
review layers would plausibly have caught on a second pass.

### D-8.4h.5 — Verified independently: the `gofmt` red is pre-existing, and the spec's Verification line is MY error to fix

**Orchestrator decision** (a verification of a sub-agent's claim, per this run's standing gate).

**Verdict.** `gofmt -l folio-go lint` printing `lint/internal/rules/licencegraph_test.go` is
**pre-existing** and is **not** a regression from Story 8.4h. The spec's `## Verification` line
recording *"expected: no output"* is a **mis-measurement made at the plan gate** — which I approved —
and is corrected in the spec rather than "fixed" by reformatting a file outside this story's scope.

**Situation.** The build reported the red as pre-existing. This run's standing gate is that
*"pre-existing / unrelated failure"* claims are confirmed at the baseline rather than accepted, because
agents sometimes mislabel their own regressions. Confirmed two ways: the file's content extracted from
`7aa283b` is unformatted there too, and `git diff --name-only 80f46a0..HEAD` does not list it — the
story never touched it.

**In simple terms.** The checklist said the floor would be clean; the floor had a mark on it before we
walked in. The right fix is to correct the checklist, not to mop a patch of floor that was never this
story's to mop — mopping it would put an unrelated file in this story's diff and make the commit lie
about what it changed.

**Why the spec is corrected and the file is not.** Reformatting it here would breach the story's own
scope fence and add a file to a commit whose whole claim is that it touches only the licence gate.
The formatting defect is real but it belongs to whoever owns that file's story.

**Consequences.** The spec's `## Verification` gofmt line is amended to record the file as a **named
standing red**, in the same style as `TestCorpusMeetsP6ExerciseFloors` — so the next story that runs
this command does not re-discover it, and does not mistake it for its own. Registered as deferred work
with no owner assigned; it is a one-line fix for whichever story next touches `lint/internal/rules/`.

**How we'd know it was wrong.** `gofmt -l` printing a *second* file after some later story — that
would mean the standing-red note is being used to wave through new breakage, which is exactly what
naming it by identity is meant to prevent.

### D-8.4h.6 — The no-commit rule for step-03 subagents is UNENFORCEABLE by prompt, at instance four

**Orchestrator decision** (recording a process defect at its fourth instance, with a mechanism the
first three did not have).

**Verdict.** Commit `ad6258b` was created by the **implementation subagent**, not by the dispatch's
finalization. That is instance **four** of D-8.5.9's pattern. The commit is **kept** — it was audited,
not trusted — and the rule is reclassified: **it cannot be enforced by prompt and must not be
attempted that way again.**

**The mechanism, which is new information.** D-8.5.9 recorded three instances and priced a fourth
against them. This dispatch's prompt restated the prohibition explicitly and in the imperative, and it
happened anyway. The builder identified why, and it is structural, not a lapse: **`bmad-build-auto`'s
step-03 mandates that the implementation handoff be passed to the subagent *verbatim*, and forbids
adding parent-authored guardrails to it.** So a no-commit instruction written in the orchestrator's
dispatch prompt **cannot travel to the agent that would have to obey it**. Every previous instance was
read as an agent ignoring an instruction; it was in fact an instruction that never arrived.

**In simple terms.** We have been leaving notes on a door the courier is required not to read. Four
times we concluded the courier was careless. The fourth time we checked the door, and the note had
never been on the far side of it.

**Why the commit is kept rather than reverted.** It is correct: content matches the task list, no
forbidden paths, `<intent-contract>` untouched, trailers present, on `main`, unpushed. Reverting a
correct commit to make a process point would rewrite history for no safety gain and lose the audit
trail. Finalize's own rule is to keep and audit; that was followed. The patch round produced **no**
further subagent commit.

**Why this is worth recording rather than shrugging at.** The risk D-8.5.9 named is real and
unchanged: at instance three the commit landed while `origin/main` was live. It is live now —
`origin/main` sits at `c985b9c`, well behind. An unaudited step-03 commit is one push away from being
public, and nothing in the pipeline prevents one; the only control that has ever worked is a
downstream agent auditing provenance, which is a control that depends on somebody choosing to look.

**Consequences.** **Stop restating the no-commit rule in build dispatch prompts** — it is inert there
and its presence creates false assurance that a control exists. Replace it with what actually works,
which this run will now do every story: the orchestrator **audits every commit's provenance and
contents at close** (`git show --stat`, author, trailers, scope against the story's file list) and
**verifies nothing was pushed**. The fix proper is out of this repository — it is a change to
`bmad-build-auto`'s step-03 contract — and is recorded here so a fifth instance is met with a known
mechanism rather than a fresh diagnosis.

**How we'd know it was wrong.** A step-03 commit appearing that the audit does not catch — meaning the
audit is being performed as a formality rather than as a check.

## Story 8.4h's findings ruled — Story 8.4i is inserted, and Epic 8's insertions are bounded (2026-09-02)

### D-8.4i.1 — A licence file that says two things is UNCLASSIFIABLE, and the rule already exists twenty lines above

**Engineering lead ruling**, taken. Measured, not inferred.

**Verdict.** `ClassifyLicenceText` stops returning on the **first** signal and instead **collects every
signal in the text**, resolving them as one: (1) **any copyleft signal → refused AS COPYLEFT**, naming
the identifier; (2) no copyleft signal but **two or more distinct permissive identifiers →
`FamilyUnknown`**, refused as unclassifiable; (3) exactly one identifier → that identifier, as today.

**Situation.** DW-125: the classifier runs `spdxLineRE.FindStringSubmatch` first and returns on the
**first match anywhere in the file**, so a full GNU GPL v3 text carrying a stray
`SPDX-License-Identifier: MIT` line classifies `(permissive, MIT)` and **passes the gate Story 8.4h
had just made fail-closed**. The mechanism is pre-existing; the consequence is new, and it defeats
8.4h's central claim that copyleft is refused by name. It needs no adversary — bundled multi-licence
files are ordinary in font distributions, where one file often covers several components.

**In simple terms.** The gate reads the first label it finds and stops looking. Hand it a folder whose
first label says MIT and whose actual contents are GPL, and it waves the folder through — not because
the check is weak, but because it stopped reading before it reached the part that mattered.

**Why this is a ruling and not a preference — it is agreement with this file's own sibling.**
`ClassifySPDXExpression`, twenty lines above at `classify.go:23-38`, **already implements exactly these
semantics** for a compound declaration: it starts `FamilyPermissive`, sets `FamilyCopyleft` if **any**
term is copyleft, and returns `FamilyUnknown` if **any** term is unrecognised. The repair is not new
policy — it is making `ClassifyLicenceText` agree with the function it sits beside. One authority per
fact, applied to a rule this module already stated once.

**Options considered.** (a) *A copyleft marker in the body outranks a permissive SPDX line* — rejected
as an **exception list**: it fixes GPL-text-plus-MIT-line while leaving MIT-plus-BSD, OFL-plus-Apache
and every future pair mislabelled. (b) *Two identifiers means we do not know* — chosen, because it is
the **axis** rather than an instance, and D-8.5.2's ground (not knowing must not read as fine) covers
two permissive ids exactly as it covers one unclassifiable one.

**Why the label is not cosmetic.** Under D-8.5.3 an asset's licence attaches to the documents users
produce, and the label is what lands in `lint/MANIFEST.md` — which AD-26 names as a **release
artifact**.

**Order 1-before-2 matters and is not stylistic.** A maintainer who reads *"conflicting identifiers"*
goes and adds an SPDX line. One who reads *"GPL detected"* removes the dependency. **Hazard indicators
fail toward the loudest, never the most precise.**

**GUARDRAIL — a Block If, not a note.** `ClassifyLicenceText` is shared between the **asset** path and
the **dependency** path (`licencegraph.go`): two populations, two risk profiles. The nine font
directories and the wordlist are visible; the Go module graph and the npm lockfile are not, and a
bundled multi-licence `LICENSE` is common in npm. **Story 8.4i's FIRST task runs the conflict detector
over the ENTIRE existing population — every dependency `LICENSE*` and every asset `LICENSE*` — in
REPORT-ONLY mode, and records the output. Only then does it turn fatal.** If it reds something
legitimate, that returns to the lead as a finding. **It must not be resolved by weakening the rule** —
that is moving the bar to fit the instrument (D-8.5.10), and it is the failure this entire thread
exists to stop.

**How we'd know it was wrong.** The report-only pass reddening a broad class of legitimate dependency
files — which would mean the conflict rule belongs on the asset path alone and the dependency path
needs its own, measured answer.

### D-8.4i.2 — The spelling defect is STRUCTURAL: a greedy catch-all, three false comments, and a fourth instance nobody named

**Engineering lead ruling**, taken.

**Verdict.** Not "accept both spellings". **Anchor the name signal**, generally: **a text carrying a
licence NAME signal that does not resolve to a known (name, version) pair returns `FamilyUnknown`. The
bare grant clause classifies as MIT ONLY when no name signal is present at all.**

**Situation.** DW-124 measured that a UFL text spelled `LICENSE` rather than `LICENCE` classifies
`(permissive, MIT)` and passes, mis-attributed — while the bare synthetic the story's own test uses
correctly reaches `(unknown, "")`. The review had **rejected** this finding by quoting a code comment
claiming the miss is loud.

**Why the narrow fix is wrong — the MIT case is a greedy catch-all and is the root defect.**
`classify.go:251` matches on `"PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"`, a clause shared by MIT,
OFL **and** UFL. The file's own comments record the collision twice: `:184-187` records that every OFL
face classified as MIT until the OFL branch was added above it, and `:220-224` records the identical
collision for UFL. **So a near-miss on a specific marker does not fall to `FamilyUnknown` — it falls
into MIT.** Accepting both UFL spellings would fix one instance of a defect that has now produced
three.

**In simple terms.** There is a catch-all bin at the end of the sorting line labelled "MIT", and
anything the earlier bins do not recognise lands in it. We have twice noticed something landing there
wrongly and each time added one more bin upstream. The bin is the problem.

**THE FOURTH INSTANCE — predicted, not measured, and the story must measure it rather than take it on
authority.** The OFL branch at `:181` requires the name **AND** `"VERSION 1.1"`, and its comment at
`:201-205` claims this *"makes a 1.0 file classify as FamilyUnknown — which D-1.3.4 deliberately makes
a LOUD build failure rather than a quiet mislabel."* **That reasoning cannot hold if OFL 1.0's text
carries the same grant clause OFL 1.1's does** — it would fall straight into MIT, exactly as
American-spelled UFL does. If it holds, the "required conjunct makes a near-miss loud" justification is
**false for both branches and has been since Story 2.2**.

**Three code comments assert a safety that does not exist, and they are corrected in place with the
original preserved verbatim.** `:241-249` asserts that a file writing `"UBUNTU FONT LICENSE"`
*"classifies FamilyUnknown … a LOUD miss, and therefore fail-safe (D-2.1.3)"* and that loosening it
would be *"for no measured need."* The measurement now exists and says `(permissive, MIT)`. Same for
the OFL comment if the prediction holds.

**This is D-8.0.1's shape at its third occurrence: a code comment asserted a branch was safe, and that
is why nobody looked.** The review rejected the finding **by quoting the comment**. So the defect is
two things — the behaviour **and** the false claim — and fixing only the first leaves the next reviewer
the same trap.

**Why it must precede Story 8.5 — correcting my own sequencing, which had it following.** Three
reasons, any one sufficient: (a) **8.5 is the first story to introduce non-OFL font licences** — the
allowlist has four members precisely because the catalogue needs them, so a UFL face is not
hypothetical, it is the point; (b) `MANIFEST.md` is a **release artifact under AD-26**, so twenty faces
of which some are mis-attributed is a defect in the compliance record `v0.1.0` ships; (c) mechanically,
D-8.4i.1 and this ruling rewrite **the same function**, and landing the second after the population it
must be correct for means touching it twice in the wrong order.

### D-8.4i.3 — The allowlist guard lands in the repair story, not in its consumer

**Engineering lead ruling**, taken.

**Verdict.** DW-120's guard goes in Story 8.4i. **Not Story 8.5** — 8.5 is the *consumer* of the
allowlist, and a guard owned by its consumer is a guard the consumer can move (D-8.5.8c). **Not its own
item** — it is one test and one red-proof.

**Situation.** Appending `"GPL-3.0"` to `fontAssetLicenceAllowlist` reddens **nothing**; all four lint
packages stay green. D-8.5.3 is an **owner decision**, and D-8.5.13 forbade adding CC0-1.0 to those
four on the ground that it would amend an owner decision to fix a scoping error. **A list anyone can
widen in one line makes that prohibition unenforceable** — a ruling recorded everywhere except where it
binds (D-8.4.31).

**Shape.** Pin the allowlist's contents against a **test-owned literal** naming D-8.5.3 as the
authority, so appending a copyleft id reddens on its own message. **An anchor the code cannot move is
the only kind that holds** — a derivation from the constant itself would pass any edit. Red-proof by
appending a copyleft id and confirming the named test reds **and that no other test reds**; if four
packages red, the guard is not the thing catching it.

### D-8.4i.4 — The excluded-path fail-open is deferred, but not as a bare note

**Engineering lead ruling**, taken.

**Verdict.** The `*/testdata/lint` `SkipDir` is a deliberate, correct exclusion and is **not** repaired.
It is filed as deferred work **with a tripwire as its stated discharge**, not as a note.

**Situation.** `folio-go/testdata/` ships inside the module zip, so a real font binary there would be
redistributed under AD-26 with **no manifest row**. Today's occupant is a text stub, which is why this
is LOW — **and "today's occupant is a stub" is exactly the kind of fact that stops being true
silently.** A register entry keyed on nobody adding a real font is not a defence.

**The discharge.** A test asserting that **no file under an excluded path is a real font program**,
keyed on the **sfnt magic bytes** (`checkSfnt` already exists, D-8.4.12) and **never on the
extension** — extension-keyed is the blind spot D-8.5.2 already found in this area. It lands in 8.4i,
beside code already being edited. **If `checkSfnt` is not reachable from the `lint` module** — a real
possibility, `lint` is its own module — it stays deferred with that tripwire as its stated discharge
and the module boundary recorded as the reason. **Do not reimplement a second sfnt reader to satisfy
this.**

### D-8.4i.5 — The four unenumerated rejections are a FINDING, and DW-87 is re-priced rather than renewed

**Engineering lead ruling**, taken.

**Verdict.** Do not re-run the review. **Do** recover the four unenumerated rejections from the existing
review record into Story 8.4i's spec as named claims with a disposition each — that is **reading, not
reviewing**, and it is cheap. Any that cannot be recovered is recorded as **"could not look"**, never
omitted (D-8.4.33). And the audit gap itself is **recorded as a finding**.

**Situation.** Story 8.4h's Review Triage Log records a bare `reject: 7` and describes only three. This
is instance three of DW-87's pattern (DW-87 as filed; D-8.4.35(b)'s deeper form where a census
disagreed with its own population by two; now this).

**Why it is a finding this time and not another renewal.** Under the standing rule that an Nth instance
**re-prices** a deferral rather than renewing it, *"we count them"* has stopped being a disposition —
**because on this story the sample was measured. Two of the three enumerated rejections were
re-checked and ONE WAS FALSE.** Four unexamined claims at a 1-in-2 observed falsification rate, on the
story whose subject is *"does the licence gate hold"*, on a compliance boundary, is not an abstract
audit gap.

### D-8.4i.6 — Story 8.4i is inserted before 8.5, and Epic 8's licence-gate insertions are BOUNDED at it

**Engineering lead ruling**, taken.

**Verdict.** Findings 1, 2, 3 and 4's tripwire land as **one story, `8.4i`, sequenced immediately before
Story 8.5.** Epic 8's order becomes **`8.4h → 8.4i → 8.5 → 8.6 → 8.4d`**.

**A new story, not a reopen of 8.4h.** 8.4h's commits and record are landed, and this run has
consistently preferred append-with-history over rewriting a closed record (D-8.4.24, D-8.5.10). The
subjects differ cleanly: **8.4h made the gate fail closed; 8.4i makes the classifier stop lying to
it.** A named story also gets its own red-proofs, which is what these findings most need. Precedent:
D-7.7.7, where DW-46 became Story 7.9 — *"not a bare fix and not a passenger."*

**Why all of it precedes 8.5, in one sentence.** D-8.5.13 made 8.5's faces admissible only through a
gate that had been **red-proved rather than merely green when they arrived** — and a gate with a live
bypass, a mislabelling catch-all and an unguarded allowlist has not been red-proved, it has been
*observed passing*. Admitting twenty faces through it would satisfy the letter of that condition and
defeat its purpose.

**THE BOUND — because this is Epic 8's fourth insertion and "the epic never ends" is a real cost.**
After 8.4f, 8.4g and 8.4h, one more is defensible; **a fifth is a pattern, not a plan.** So: **`8.4i`
is the last insertion this epic gets in the licence-gate area.** Anything further found here after it
is **registered and routed to Epic 15's release gate**, where AD-26's manifest is already a named
obligation of Story 15.3, and is **not** built as `8.4j`. **One exception — the criterion every
insertion so far has actually met: a demonstrated live bypass of a gate the epic has declared
fail-closed.** That class buys a story; nothing below it does. If 8.4i's report-only population pass
turns up such a bypass, it returns to the lead to be ruled **against this bound rather than around
it.**

**Hard constraint on 8.4i.** It **must not** turn the gate fatal in the same commit that first measures
the population. **Measure, record, then close** — and if the measurement reds something legitimate,
that is a finding to route, not a rule to soften.

## Story 8.4i's plan gate — the predicted fourth instance holds, and the census found a fifth constraint (2026-09-02)

### D-8.4i.7 — CONFIRMED by measurement: OFL-1.0 classifies as MIT, and the bypass also works through a copyleft SPDX line

**Orchestrator decision** (accepting two measured findings; the second is new and in no register).

**Verdict.** D-8.4i.2's predicted fourth instance **holds**. The SPDX-published **OFL-1.0** text
classifies as **`(permissive, "MIT")`**, not `FamilyUnknown`. So the comment at `classify.go:201-205`
is **false**, `:241-249` is false for American-spelled UFL (also `(permissive, "MIT")`), and the
*"required conjunct makes a near-miss loud"* justification is **false for both branches and has been
since Story 2.2**.

**How it was measured, and why the method matters.** Two ways, not one: the published OFL-1.0 text
carries *"Permission is hereby granted, free of charge, to any person obtaining a copy of the Font
Software…"* under the title *"SIL OPEN FONT LICENSE Version 1.0 - 22 November 2005"* and contains
`"Version 1.1"` **nowhere** — so it misses the OFL branch's required conjunct and falls into the MIT
catch-all immediately below. Then it was **probed** at `7e4b2c4` rather than reasoned about. The
prediction came from the lead; the story was told to measure it rather than take it on authority, and
it did.

**In simple terms.** The code carried a note saying "if the version does not match, this fails loudly."
It does not. It falls into the bin below, which is labelled MIT, and passes. The note has been wrong
for as long as the branch has existed, and it is the note that stopped anyone checking.

**THE NEW FINDING, in no register before this gate.** DW-125's bypass works with a **copyleft SPDX
line**, not only with copyleft *marker text*: a file carrying an `MIT` line followed by a
`GPL-3.0-only` line yields `(permissive, "MIT")`. **This is a second, independent reason the rejected
alternative was wrong** — *"a copyleft marker in the body outranks a permissive SPDX line"* would not
have caught this at all, because there is no marker text to outrank; both signals are SPDX lines. The
ruled fix — collect every signal, resolve as one — covers it.

**Consequences.** Both false comments are corrected in place with the originals preserved verbatim.
The story's acceptance must cover **three** shapes of the bypass, not one: copyleft marker text plus a
permissive SPDX line; two distinct permissive SPDX lines; and a permissive SPDX line plus a **copyleft
SPDX line**.

**How we'd know it was wrong.** A real OFL-1.0 file in the population that classifies correctly — which
would mean the probe used a text that differs from what ships, and the measurement needs redoing
against the real file.

### D-8.4i.8 — The sfnt tripwire is BUILT, not deferred: the prohibition binds on the capability, not the identifier

**Orchestrator decision** (ruling a judgement call the plan gate flagged, rather than routing it — it
turns on what a prohibition meant, and the answer is in the prohibition's own stated purpose).

**Verdict.** Story 8.4i **builds** DW-123's tripwire. D-8.4i.4's escape hatch — *"if `checkSfnt` is not
reachable across the `lint` module boundary, leave it deferred"* — is **not** taken, even though its
literal condition is true.

**Situation.** `checkSfnt` is genuinely unreachable: it is unexported and lives in the `folio-go`
module, and the repository has three separate Go modules with no `go.work`. But the `lint` module
**owns its own sfnt reader** — `looksLikeSfnt` at `lint/internal/rules/fontsassets.go:144` — callable
as-is from a test in that package. So the tripwire is buildable **writing no second reader**.

**In simple terms.** The instruction was "do not build a second one of these." We found there is
already a second one, built long ago, sitting in the room where the work is happening. Using it is not
building a third.

**Why the hatch does not apply.** D-8.4i.4's stated reason for the hatch was explicit: *"Do not
reimplement a second sfnt reader to satisfy this."* The prohibition is on **duplicating the
capability**, and the module-boundary condition was the lead's guess at the circumstance that would
force duplication. That guess was wrong in a way that makes the hazard vanish rather than bite — the
condition is true and the thing it protects against cannot arise. Taking the hatch here would defer a
cheap, real guard on a technicality.

**Options considered.** (a) *Defer per the hatch's literal text* — rejected: it would honour the
wording while defeating the purpose, and leave a fail-open path guarded by nothing but a register entry
keyed on nobody adding a real font. (b) *Route it back to the lead* — rejected as disproportionate
under D-000.8; the prohibition's own sentence states its purpose and the purpose is satisfied.
(c) *Build it using the reader the module already has* — chosen.

**The accepted risk, stated.** If the lead reads `checkSfnt` as binding on the **identifier** rather
than the capability, this is wrong and is cheap to reverse now — one test, deleted. Flagged in the
story so it is visible at the build's review rather than buried here.

**How we'd know it was wrong.** `looksLikeSfnt` turning out to disagree with `checkSfnt` on a real font
— then there genuinely are two readers with two answers, which is the duplication the prohibition
existed to prevent, and the tripwire becomes the thing that proved it.

### D-8.4i.9 — Two premise corrections from the census, one of which constrains the fix

**Orchestrator decision** (accepting two measured corrections; per D-8.5.10 neither is an intent gap,
because the verdict, scope and acceptance all survive).

**Verdict.** Both corrections are recorded as Design Notes and one becomes an explicit task
constraint. D-8.4i.1's ruling stands unchanged.

**Correction 1 — a false premise inside D-8.4i.1's own guardrail.** The guardrail justified the
report-only census partly on the risk that *"a bundled multi-licence `LICENSE` is common in npm."*
Measured: **`ScanNPMGraph` never calls `ClassifyLicenceText`.** It classifies the lockfile's declared
`license` string through `ClassifySPDXExpression`, and reads no npm `LICENSE` file at all. **So no
bundled npm licence file can reach this change.** The census still runs — the Go module graph is a real
population and the ruling's *shape* was right — but its stated risk was overstated by one of its two
halves.

**Correction 2 — a constraint the rulings did not mention, and it directly shapes the fix.** The BSD
branch's third disjunct matches `"REDISTRIBUTION AND USE IN SOURCE"`, which is a **clause, not a
name** — and it is how **7 of the 9** dependency licences classify. They carry **no name signal and no
SPDX line at all**. So a name-anchoring rule that treats that disjunct as a name signal would return
`FamilyUnknown` for them and **red the Go standard library's own licence**.

**In simple terms.** The new rule is "if you see a licence's *name* and cannot place it, say you do not
know." Most of the dependency licences here are recognised not by a name but by a stock sentence every
BSD-style licence opens with. Treating that sentence as a name would make the build reject the Go
standard library — which is the exact shape of self-inflicted standing red this run has twice
registered and once ruled against.

**Consequences.** The name-anchoring rule in task 3 must distinguish **name signals** from **clause
signals**, and the BSD clause disjunct is explicitly a clause. Written into the spec as a task-3
constraint rather than left for the build to rediscover.

**Why the census earned its place.** D-8.4i.1 required a report-only population pass before anything
turns fatal, on the reasoning that the dependency population is not visible. The census came back
**clean** — 9 dependency + 12 asset + 8 fixture `LICENSE*` files, **zero** carrying two distinct
identifiers, an unresolved name signal, or a copyleft/permissive mix, and nothing changing verdict —
but it produced this constraint, which nobody had predicted and which would otherwise have been found
by the build going red on the Go standard library. **The census is still run as task 1 at build time,
and a disagreement between the two runs is itself the halt.**

### D-8.4i.10 — Story 8.4h's rejection record is unrecoverable for 4 of 7, recorded as "could not look"

**Orchestrator decision** (discharging D-8.4i.5).

**Verdict.** Three of Story 8.4h's seven rejections were recovered and carried into 8.4i's spec as
named claims with dispositions. **Four are recorded as "could not look"** — never as absent, never as
fine.

**Recovered, with disposition:** the duplicated `TestClassifyUbuntuFontLicence` case (rejected as
churn — **stands**); the missing American-spelling test (rejected as a *"loud miss"* — **OVERTURNED,
the rejection is false**; this is DW-124); AC7's suite placement (rejected on the spec's own Code Map —
**stands**).

**Why the other four cannot be recovered.** The record has **no `rejected_findings` block, no
per-finding rejection list, and no reviewer-layer attribution for any finding in that story.**
Rejections 4–7 have no summary, no location, no severity and no reason anywhere in the artifact. There
is nothing to read.

**In simple terms.** The review reported that it threw seven things away and described three of them.
The other four were not written down at all — so we cannot check them, and the honest record of that is
"we looked and there was nothing to look at", not silence.

**Why this is a finding and not a shrug (D-8.4i.5).** One of the three recoverable rejections was
**false**. Four unexamined claims at a measured 1-in-3 falsification rate, on the story whose subject
is *"does the licence gate hold"*, on a compliance boundary, is a measured risk rather than an abstract
audit gap. Registered as such.

**Consequences.** The triage log's `reject: N` count is not evidence of anything on its own, and this
run stops treating it as such: a rejection that is not enumerated cannot be spot-checked, so from here
a compliance- or security-boundary story's close **names** what was rejected or records that it could
not. This is DW-87's third instance and it re-prices rather than renews.
