# Epics 11–14 — decision log

The audit trail for the run that delivers Epics 11, 12, 13 and 14. One entry per decision that outlives
the story that raised it. Written for someone who was not in the run.

**`## Lead Grounding` is appended below once the engineering lead reports.**

---

## Standing decisions (settled by the owner at the terminal, 2026-09-05)

### D-000.1 — The run's five standing parameters
**Owner decision.** Answered at the terminal at invocation, unprompted by the setup questions.

**Verdict.**
1. **Answer channel: this terminal.** Not Telegram, and this holds for every question for the whole run.
2. **Run continuously to the end.** No checkpoint pause after a story, and none at an epic boundary.
3. **Heavy tests at the end of each epic.** Integration and e2e execution moves to the epic boundary;
   unit tests, typecheck, lint and build still run every story.
4. **Spec approval is delegated** to the orchestrator and the engineering lead. The owner reads the
   plain-terms opener in the per-story report.
5. **Build order: to be settled below** — see D-000.3.

**Consequences.** Under (3) every story's `## Verification` carries unit, typecheck, lint and build, plus
a **compile-only** check of any new integration or e2e package under its build tag. That compile check is
not optional: a package that fails to compile under its tag **silently skips its tests**, so without it
the epic-boundary catch-up opens with a compile error from five stories ago. Each story's Delivery Log
must name the unrun suites explicitly rather than reporting a bare "green".

Under (2), a genuine fork still stops the run — continuous means no *scheduled* pauses, not that a
decision the direction cannot settle gets guessed at.

**How we'd know it was wrong.** An epic-boundary catch-up run that goes red against three or more stories
at once: that is the deferral's cost arriving, and it would argue for moving to `every-story` for the
remaining epics.

---

### D-000.2 — Epics 11–14 have no prior rulings, despite a log that appears to cover them
**Orchestrator decision** (a record-keeping finding, not a design choice).

**Verdict.** This run starts a NEW decision log. `epic-8-15-decision-log.md` does not cover epics 11–15.

**Situation.** Four decision logs already exist. The one named `epic-8-15-decision-log.md` reads as though
it spans Epic 8 to Epic 15, which would make this run a continuation with rulings to inherit. Measured
instead: of its **92** `### D-` entries, the prefixes are **73 × `D-8.`**, **13 × `D-000.`** and
**6 × `D-16.`** — and `D-11.` through `D-14.` occur **zero** times across both that log and
`folio-mvp-decision-log.md`. The count of 92 is the positive control that the query works.

**Why it matters.** Had the name been trusted, the lead would have been told to re-ground from prior
rulings that do not exist, and would have reported finding none — or worse, inferred Epic 8's rulings
applied here.

**Consequences.** The lead grounds from scratch for this program. The misleading log name is left alone:
renaming a 265 KB committed artifact mid-run is churn, and this entry is the pointer.

---

### D-000.3 — The build order must be established by measurement, not by number
**Orchestrator decision**, pending the lead's grounding report.

**Verdict.** Before any spec is written, establish **what is already true** of epics 11–14 in the shipped
codebase. Stories found already satisfied are closed as such with evidence; stories whose premise has been
invalidated are re-scoped at the plan gate.

**Situation.** This repository built epics **15, 16 and 17 before 11–14**. All 23 stories in 11–14 are
`backlog`, and their specs were written against a codebase that has since moved a long way. Three
collisions are visible without looking hard:

- **Epic 11 is font weight and slope** — "the shipped families gain a weighted and a sloped face", "the
  engine resolves a face from the declared weight and slope". **Epic 16 spent itself on the shipped face
  set** and added Roboto as a fourth face with a machine-checked byte-identity rule. 11.1's premise has to
  be re-measured against `Shipped()` as it stands.
- **11.3, "the canvas paints the weight the engine resolved"**, touches the Bold/Italic controls that
  changed on **2026-09-04**: pressing a pressed toggle now clears rather than writing `bold: false`.
- **Epic 14 is "the designer's controls read as one product"** — ten stories about the property panel, the
  data tab and the table editor, written before Epics 16 and 17 reworked the property rows, the typography
  panel, the font browser, and the number-field and prose-field behaviour.

**In simple terms.** These four epics are a set of instructions for a building that has since had three
more floors added. Some instructions now describe work already done; others describe work that would
knock out a wall something new is resting on. Reading the instructions first, and the building second, is
how a wall gets knocked out.

**Why this wins.** The alternative — writing specs in numeric order and letting each build discover the
drift — is exactly the failure this run has already paid for twice at single-story scale: Story 17.5's
Code Map went stale in under a day, and its Design Notes described a hazard that had been fixed hours
earlier. At four epics' scale that cost is multiplied by 23.

**Consequences.** The first work of this run is a survey, not a spec. Its findings are logged per story.

**How we'd know it was wrong.** If the survey finds nothing stale in any of the 23 stories, it was
over-cautious — a cheap outcome, and still worth knowing.
