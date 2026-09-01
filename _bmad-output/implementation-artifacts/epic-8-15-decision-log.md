# Decision log — Epics 8–15 delivery run

Opened 2026-09-02. This log continues `epic-7-8-decision-log.md`, which covers Epic 7 in full and
Epic 8 through Story 8.4f. That file is NOT superseded and must not be edited by this run: rulings
D-7.x, D-8.x and D-R7.x cited here live there. New rulings are numbered `D-<story>.<n>` and appended
here.

**Run target.** Every open item in Epics 8–15: Epic 8's remaining three stories, the retroactive
closure of Epics 9 and 10, then Epics 11–15 to the `folio-go/v0.1.0` tag.

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
