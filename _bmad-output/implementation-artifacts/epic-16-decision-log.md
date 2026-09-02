# Epic 16 — decision log

The run-scoped record for the Epic 16 delivery run started 2026-09-02. One entry per decision, newest
appended. `sprint-status.yaml` carries status values only; the per-story narrative lives in each spec's
`## Delivery Log`; build-auto's own findings live in each spec's `## Auto Run Result` and
`## Review Triage Log`.

**The decisions that CREATED this epic are not here.** D-16.1 through D-16.6 were taken before the run
began, while the epic was being planned, and they live at the end of
[`epic-8-15-decision-log.md`](./epic-8-15-decision-log.md) with the Epic 8 and 15 decisions they amend —
D-16.1 reverses a Non-goal that D-8.4d.1 had amended hours earlier, and separating the two halves of that
argument would make neither readable. They are load-bearing for every story here:

| ID | What it settles |
|---|---|
| D-16.1 | **Owner.** The catalogue stops being the only source; fonts are fetched from the web. Reverses *"No live font service"*. |
| D-16.2 | **Owner.** `AVAILABLE LOCALLY` is a machine-local IndexedDB store of fetched faces, never the OS font list. |
| D-16.3 | **Measured.** The index cannot be fetched (no CORS) and is a build-time snapshot; only faces, `OFL.txt` and `METADATA.pb` are live. |
| D-16.4 | The forbidden-host scan is amended and grows a second half, never deleted. |
| D-16.5 | **Owner.** Variable-only families are refused; the ones worth having are derived ahead of the build; browser-side instancing is refused. |
| D-16.6 | **Measured.** The embed command accepts a variable face the renderer refuses — a pick can write an unrenderable `.folio`. Placed in Story 16.0. |

## Lead Grounding

*Filed by the orchestrator from the engineering lead's grounding report. Pending — the lead is grounding
as this log is written; this section is filled in before the first plan dispatch.*

## Standing decisions

### D-16.R.1 — The run's standing decisions, set by the owner at the terminal

**Owner decisions**, taken with `AskUserQuestion` at invocation of `run-dev-cycle Epic 16` on
2026-09-02. Recorded together because they shape every story in the run rather than any one of them.

**Verdict.**

| Decision | Setting | What it means in practice |
|---|---|---|
| Answer channel | **Terminal** | Mid-run questions are asked inline with `AskUserQuestion`, not over Telegram, even though `~/.claude/telegram.env` exists and the Telegram lane is available. The owner is staying at the keyboard. |
| Checkpoints | **After the epic** | The run goes 16.0 → 16.4 without scheduled pauses, reporting per story, and stops at the end or on a genuine escalation. |
| Heavy tests | **End of run** | One catch-up run of the heavy suites after 16.4 lands. Unit + lint + build run every story regardless. |
| Plan gate | **Orchestrator + lead** | Specs are gated silently; the owner sees the plain-terms opener in each per-story report rather than the spec itself. |
| Build order | **Numeric** | Not asked, because for this epic dependency order and numeric order are the same: 16.0 is the precondition, 16.3 needs 16.0–16.2, 16.4 needs 16.2–16.3. |

**Situation.** The run had four settings to fix before dispatching anything, and three of them trade
feedback latency against wall-clock. The one that matters most here is the heavy-test cadence, because
of an unusual property of this repository: **its Playwright e2e specs are compile-checked and never
executed** (D-000.4), and the file heading `e2e/component-properties.spec.ts` says so in its own words —
*"THIS FILE IS COMPILED, NEVER EXECUTED … what follows is a typed witness that the roles, names and
flows below EXIST in the source — not evidence that any of them work in a browser."*

**In simple terms.** The project has a drawer full of tests that have never been run. They prove the
buttons exist by name; they prove nothing about whether pressing one works. That is exactly how the
defect that opened this epic reached the owner: a pick that fails at the WASM boundary is invisible to
every gate the project currently runs, so it arrived as a screenshot instead of as a red test.

**Consequences, and one override the owner does not need to approve each time.** `end-of-run` moves the
heavy suites to a single catch-up run after 16.4. **Two stories override it**, per this skill's standing
rule that a story whose own correctness is integration-shaped carries its integration commands
regardless of cadence:

- **Story 16.0** — its acceptance criterion *is* a browser run. Its whole subject is a boundary that no
  Go test can observe, and its own spec says the claim is unprovable without one. Deferring that would
  make the story unverifiable in the only dimension that matters.
- **Story 16.3** — the font browser's specimens, focus trap and staged-confirm flow are DOM behaviour
  with no non-browser witness at all.

16.1, 16.2 and 16.4 follow the `end-of-run` cadence: unit + lint + build in-story, plus a compile-only
check of the e2e specs so a package that fails to compile under its build tag cannot silently skip.
**The debt is stated per story**, in each Delivery Log, naming the unrun suites — never reported as
"green" on unit tests alone.

**How we'd know the cadence was wrong.** The end-of-run catch-up going red across more than one story,
which would mean the deferral hid a break that per-story running would have attributed immediately.

### D-16.R.2 — The variable-font refusal lands on the head of the distribution, not the tail. A bounded derive-ahead batch is added, and non-addable families are hidden rather than refused

**Owner decision**, taken at the terminal on 2026-09-02, on a correction to the framing the owner had
been given a few hours earlier. **This does not reverse D-16.5** — it prices it properly and adds the
work D-16.5 left unbounded.

**Verdict.** Epic 16 gains a story that runs a bounded batch of variable-only families through
`tools/fontgen` ahead of the build and commits the static instances, joining them to the existing
derived tier. Families that still cannot be added are **filtered out of the font browser**, not listed
and refused.

**The situation, and the part of it that was my error.** D-16.5 was put to the owner with the cost
quoted as **"558 of 1,946 families — 28.7%"**. That is arithmetically true and materially incomplete:
it reads as a quarter of a long tail, which is how a reader prices it. The engineering lead re-measured
the same index sorted by Google's own `popularity` field at grounding, and I re-verified it
independently rather than relaying it:

- **37 of the 50 most popular families are variable-only**, and therefore refused. Among them:
  **Roboto, Open Sans, Inter, Montserrat, Raleway, Nunito, Oswald, Playfair Display, DM Sans, Roboto
  Mono, Heebo, Nunito Sans, Google Sans**.
- The hardest counterexample was checked directly rather than inferred: `ofl/roboto` contains only
  `Roboto[wdth,wght].ttf` and `Roboto-Italic[wdth,wght].ttf`. There is no static Regular to fetch.

**In simple terms.** The refusal was priced as *"a quarter of the catalogue you have never heard of"*.
It is actually *"the first thing anyone types"*. An author opens a browser whose header promises the
library the web publishes, types **Roboto**, and is refused on day one. D-16.5's own stated falsifier —
*"authors repeatedly wanting a specific VF-only family, often enough that the derive-ahead batch
becomes a queue nobody services"* — does not fire gradually under that distribution. It fires
immediately, which means the batch is load-bearing rather than a courtesy and cannot be left unowned.

**Options considered.** (a) **Derive a bounded batch and hide the rest — chosen.** (b) *Hide
non-addable families, derive nothing* — the browser's count becomes honest and nobody is ever refused
for being variable, but Roboto and Open Sans are simply absent with no explanation and the promise
shrinks quietly, which is a worse failure than a stated one. (c) *List them and refuse with a reason* —
what the specs said until now; the most honest and the most visible, and it means the most common first
action in the product fails. (d) *Reopen browser-side instancing* — refused again, on D-16.5(c)'s
unchanged grounds: it makes the embedded face a function of the author's browser and library version,
so the same template stops rendering the same PDF everywhere. Neither the lead nor the orchestrator
recommended it, and the owner did not take it.

**Why this one wins, and what it costs.** It is the only option that keeps the determinism guarantee
**and** lets an author have Roboto. The accepted costs are stated rather than discovered: one extra
story; committed font bytes in the repository; a NOTICE and licence record per derived face; and a
**standing curation job** — the batch is a snapshot of popularity and popularity moves, so somebody
owns re-running it. That last one is the real price and it is the thing D-16.5 was trying to escape
when it reached for the live library in the first place.

**One constraint this does NOT hit, which is why it is affordable.** D-8.4d.1 already moved catalogue
faces to **fetched from the release's own origin on first pick** rather than precached at first load.
So derived static faces do **not** inflate the first-load payload that Story 15.2 is about to put a
gate on. The binding constraints are committed repository bytes and the per-face licence work, not
`brotli.totalBytes`.

**Consequences.**
- A new story is inserted: **16.1a**, after 16.1 (which establishes the snapshot and the join) and
  before 16.3 (whose family count and empty states must be honest about what is addable).
- **16.1 and 16.3 change from show-and-refuse to filter-out.** A family that cannot be added is not
  offered. The refusal path stays in the engine for anything that slips through — a hidden row is a
  presentation choice, never a guard.
- The browser's family count is **the addable count**, not 1,946, and it says which it is.
- The batch's size, membership rule and owner are 16.1a's to fix and must be written down there, not
  left implicit. A batch nobody owns is the failure mode this decision exists to prevent.

**How we'd know it was wrong.** Authors asking for families outside the batch at a rate that makes the
curation job a queue — the same falsifier D-16.5 named, now measured against a batch that exists rather
than against a policy that had none.

### D-16.R.2a — CORRECTION to D-16.R.2: the derivation premise was false, and the remedy is cheaper than the option the owner was offered

**Orchestrator correction**, appended rather than rewritten (D-16.R.2 stands above as taken). Raised by
the engineering lead while reading for Ruling 1, and **independently verified here before being acted
on** — the orchestrator's standing rule is that a claim a decision hinges on is re-measured, not
relayed.

**What was false.** D-16.5 said the families worth having are *"derived ahead of the build by
`tools/fontgen`, which is how two of the six affected Thai families are already available"*, and
D-16.R.2 inherited that and priced the remedy as **funding a derivation queue**. Measured:

- **`tools/fontgen/instance_faces.py` drives a hardcoded three-entry `UPSTREAM` list** — `Noto Sans`,
  `Noto Sans Thai`, `Noto Sans SC`. All three are the **engine's** faces under `folio-go/fonts/`.
- **None of the 21 designer catalogue faces is derived.** Every `NOTICE.md` under
  `folio-designer/public/fonts/` states it in terms: *"NO DERIVATION APPLIES. This file is the upstream
  file itself, byte for byte. The upstream project publishes static TTF instances"*, and adds that
  *"this repository cannot derive a face at all"*.
- So of the two Thai families D-16.5 cited as already done, **only `Noto Sans Thai` is a derivation**.
  `Noto Serif Thai` is an upstream static from `github.com/notofonts/thai`.

**The fact that changes the economics.** This repository **already ships Roboto and Inter** — two of the
families my escalation named as refused — as byte-for-byte upstream statics, from
`github.com/googlefonts/roboto-classic` and `github.com/rsms/inter` v4.1. They appear among the 37
refused only because the **`google/fonts` mirror** carries VF-only builds of them.

**In simple terms.** "Roboto is variable-only" is not a fact about Roboto. It is a fact about *the shelf
we were looking at*. The Roboto project publishes a static Regular; Google's mirror just doesn't stock
it. We already have that file, sitting in this repository, with its licence and its NOTICE.

**What this changes, and what it does not.** The owner's verdict — **bounded batch, hide the rest** —
**stands and is not re-opened**: the price is still the head of the distribution rather than the tail,
which was the substance of the escalation. What changes is the **mechanism and its cost**. The batch is
not a derivation queue needing a `fontTools` pipeline per family; it is *"add an upstream static to the
local face tier"*, **the mechanism this repository has already executed 21 times**, with a proven
provenance regime behind it (committed `LICENSE*`, `NOTICE.md` recording both upstream and committed
digests, and the build-time nameID 13 tie).

**Consequences.**
- **Story 16.1a is not a fontgen story.** Its unit of work is a catalogue addition per family, from that
  family's **own project release**, not from the `google/fonts` mirror. Derivation stays available for a
  family that genuinely publishes no static, and that is now the exception rather than the plan.
- The batch's membership rule gains a prior question 16.1a must answer per family: *does this family's
  own upstream publish a static Regular?* Only a "no" needs derivation.
- **The record must not be read later as "the owner chose a derivation queue".** They chose a bounded
  batch; the queue was my framing and it was wrong.

**How we'd know this correction was itself wrong.** A family in the batch whose own upstream turns out
to publish no static either — at which point derivation is the remedy for that family, exactly as
D-16.5 imagined, and the exception proves to be commoner than this correction assumes.

### D-16.R.3 — RULING: the bundled catalogue survives as the LOCAL FACE TIER

**Engineering lead ruling**, applied. Confidence: high. *(Full text in the run transcript; the operative
content is recorded here because it governs 16.1, 16.3 and 16.4.)*

**Verdict.** The 21 committed faces and their whole provenance regime survive Epic 16 unchanged.
`pickCatalogueFamily` **gains a source; it does not swap one.** The term in specs and code is the
**local face tier** — *not* "derived-static tier", which would carry D-16.5's error (none of the 21 is
derived; see D-16.R.2a).

- **Join key: exact `family` string equality.** No case-folding, no whitespace normalisation, no fuzzy
  match. Measured ground: `Inter Display` and `Source Serif 4 Display` have **no index row at all**, so
  2 of 21 are unjoinable under any normalisation — while `Geist` / `Geist Mono` / `Geist Pixel` is
  precisely the neighbourhood a loose matcher gets wrong. **A local face with no index row is
  local-tier-only, and that is correct behaviour, not a defect.**
- **Precedence: local wins, with no fetch at all** — no `METADATA.pb`, no `OFL.txt`. Ground: the
  committed bytes carry a **stronger** record than any fetch can produce. Preferring a fetch would
  replace a verified record with an unverified one.
- **Divergence is deliberately not reconciled.** Under AD-8 and D-16.2 a face is identified by the
  SHA-256 of its bytes, so "upstream released a newer version" is a **different face**, not a newer one,
  and silently preferring it is the silent substitution both forbid. **No staleness check, no update
  prompt, no version compare in this epic** — registered to `deferred-work.md` with its trigger.
- **Disclosure rides 16.4's existing grammar; there is no fourth group.** `AVAILABLE LOCALLY` reads as
  *"faces this machine holds without a third-party fetch"*, covering both a previously-fetched web face
  and a local-tier face. In the browser, a local-tier row states that no download is needed.

**Guardrails.** `font-catalogue.json` is not deleted or emptied; `font-catalogue.test.ts` stays green
with its 21-face population floor; `maximumCacheAssets` stays 64 and the snapshot module consumes no
slot; no path by which a local-tier pick issues a third-party request.

### D-16.R.4 — RULING: one licence vocabulary, a closed token table, and three states rather than two

**Engineering lead ruling**, applied. Confidence: high. **This is the ruling that stops Epic 16 shipping
broken**: applied literally, 16.1's original AC would have refused every family except by the accident
that `UFL` matches — a partial pass that looks like it works.

**Verdict.** A new module `folio-designer/src/font-licence.ts` owns the mapping and **names no host**
(so D-16.4's declared-host module stays a small subject for the scan's second half).

- **Closed table, `METADATA.pb` token → SPDX id:** `OFL` → `OFL-1.1`; `APACHE2` → `Apache-2.0`; `UFL` →
  `Ubuntu-font-1.0`; `CC-BY-SA` → **present and refused**, with its reason stated (ShareAlike is
  copyleft and outside the allowlist D-8.5.3 admits).
- **Three states, never two** — *mapped-and-admitted*, *mapped-and-refused* (named token, stated
  reason), *unmapped* (refused, and the message says **the token was not recognised**, not that it is
  forbidden). A fifth upstream directory must read as *"we have not classified this"* rather than as a
  policy decision nobody made. `cc-by-sa` is in the table precisely so that absent and refused stop
  looking the same.
- **An unmapped token is a refusal and never a default.** No fall-through, no permissive default, no
  warn-and-continue.
- **`font.licence` in the `.folio` carries the SPDX id, never the upstream token.** Precedent, not
  preference: `font-catalogue.json` already writes `Ubuntu-font-1.0` rather than `UFL`, and
  `licenceSignatures` is keyed on SPDX. Two vocabularies in one field make a document unsortable by its
  own terms.
- **MIT stays admissible and gets no table entry**, and the module says why: D-8.5.3's four identifiers
  are **owner policy** about acceptable licences and are not amended here; `google/fonts` publishes no
  MIT token, so MIT has nothing to map. That is **absence, not narrowing.**
- **A test asserts the table's admitted value set is a subset of D-8.5.3's four**, so the table can
  never outrun the owner's allowlist.

### D-16.R.5 — RULING: the nameID 13 tie transfers to runtime, and it lives in Go

**Engineering lead ruling**, applied. Confidence: high. This closes the gap that made the runtime check
**strictly weaker** than the build gate it replaces — the defect class of D-8.6.5.

**Verdict.** Required, and sited in **Go, inside `embedFontFamily`, beside the `fvar` check Story 16.0
adds.**

- **What is compared:** the command payload's `licence` (an SPDX id, per D-16.R.4) against **nameID 13**
  read from the bytes about to be embedded, through a closed signature table mirroring the build-time
  one — `OFL-1.1` → /SIL Open Font License/i, `Ubuntu-font-1.0` → /Ubuntu Font Licence/i. **Substring,
  not prefix or equality**: measured at build time, `cascadiacode`'s description opens *"Microsoft
  supplied font…"* and carries the OFL sentence further in. An SPDX id with **no signature entry is a
  refusal**, not a skip.
- **Why Go and not the browser:** the browser is not the only door. `embedFontFamily` is reachable from
  `wasm.Engine.Apply` and from a hand-authored command, so a browser-side check leaves exactly the gap
  `fontset.go:228` exists to cover for `fvar` — D-16.6's argument, unchanged. One implementation also
  covers the local tier and the machine store instead of three.
- **Absent or unparseable nameID 13: REFUSE, and say which of the two.** D-8.6.1 already makes the
  licence record required of a chain-named face; a face that cannot state its own terms is the case that
  requirement exists for.
- **The falsifier is owed BEFORE the guard ships.** The build-time tie is measured over **21** faces,
  not 1,946. Story 16.1 must sample **≥50 families across `ofl`/`apache`/`ufl`** and report the rate at
  which the guard would refuse them. A rate materially above zero means the signature table is too
  narrow for the wider population and comes back to the lead as a finding — **never fixed by softening
  the guard to a warning.**
- **Both guards kept.** The build-time tie at `font-catalogue.test.ts:355-366` is not deleted or
  weakened because Go now checks it.

**Guardrail with a deadline.** This narrows `embedFontFamily`, which is reachable through the exported
API, so under D-7.8.3/D-8.2.2 it joins the **before-the-tag set**. `folio-go/v0.1.0` is verified **not
cut**, so it is free now and expensive after. **It lands in this epic or it does not land**, and it goes
on D-000.15's running list of what the format freedom was spent on, for Story 15.3.

### D-16.R.6 — RULING: the family directory is derived, then CONFIRMED, and never read as a licence

**Engineering lead ruling**, applied. Confidence: high.

**Verdict.** **Slug rule, exact: lowercase the family name, then delete every character outside
`[a-z0-9]`.** Verified 8 of 8 against deliberately awkward families — `Press Start 2P` →
`pressstart2p`, `Baloo Bhai 2` → `baloobhai2`, `Alegreya SC` → `alegreyasc`, `Source Serif 4` →
`sourceserif4`, `DM Sans` → `dmsans`, `Ma Shan Zheng` → `mashanzheng`, `Playpen Sans Thai` →
`playpensansthai`, `Noto Sans Thai Looped` → `notosansthailooped`.

- **This is a derivation closed by verification, which is why it is not the guess the story forbids one
  level down.** The resolved directory is accepted **only if `METADATA.pb`'s `name` string-equals the
  index family the author picked.** A mismatch is a **refusal**, never a fallback to the next directory.
  The directory gets the same discipline the filename already has: proposed from the family name,
  confirmed by the resource itself.
- **`METADATA.pb` always wins on licence, and the probe result is never evidence of terms.** Probe order
  `ofl`, `apache`, `ufl`, `cc-by-sa`. Measured ground: **upstream moves families between directories —
  `apache/roboto` now 404s and Roboto lives in `ofl/`** — so reading layout as a licence assertion would
  let a family that moved silently change the terms a document publishes. That is AD-26's Prevents
  exactly.
- **Layout disagreement is an observation, not a refusal.** Resolved at `ofl/x` while the token says
  `APACHE2` → `METADATA.pb` wins, `Apache-2.0` is admitted, and the divergence is recorded because
  systematically it means the probe order is costing round-trips.
- **Probing is once per pick**, never at index render and never on a keystroke; 4 probes × 1,946 must
  not become a browsing cost. The build step **may** carry the resolved directory in the snapshot as a
  **path** (a path is not a claim about terms), re-confirmed by the `name`-equality check at pick time —
  and the snapshot **must not** carry the licence token, which would be a second licence authority
  ageing on its own schedule.
