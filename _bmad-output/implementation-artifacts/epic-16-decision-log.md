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

### D-16.R.7 — RULING (replaces part of D-16.R.5): the nameID 13 tie becomes three-valued and two-sided. Silence admits; contradiction refuses

**Engineering lead ruling**, applied. Confidence: high on the shape, **medium on the Apache regex
specifically**. **This replaces D-16.R.5's "absent or unparseable ⇒ REFUSE" clause**, which the lead
identified as its own error — a guard that conflates *"caught a lie"* with *"could not look"*, which is
the exact failure it had ruled against one level up in the token table (D-16.R.4's three states).

**How it was caught, and this is the process working rather than a mishap.** D-16.R.5 required a
≥50-family sample **before** the guard ships. Story 16.1's plan dispatch ran it at the gate — 100
upstream faces, name tables parsed in pure Python, nothing written to the repository — and measured
**50 of 100 refused**, 41 of them from a single gap. The falsifier fired exactly where it was aimed.

**The two measurements that forced the change.**

1. **`Apache-2.0` was admitted by the token table and structurally refused by the tie.** D-16.R.4 admits
   `APACHE2` → `Apache-2.0`; the build-time signature table the tie was told to mirror
   (`font-catalogue.test.ts:197-200`) has **exactly two rows**, `OFL-1.1` and `Ubuntu-font-1.0`. Every
   Apache family passed the licence gate and was refused one step later.
2. **All three static `ufl/` families carry no nameID 13 at all** — their terms live in **nameID 0**.
   Invisible to the build-time tie because the two Ubuntu faces committed here are not the upstream
   files. Under D-16.R.5 as written, every genuine UFL family would have been refused.

**Verdict — three outcomes, mirroring D-16.R.4's own three states one level up.**

| Outcome | Condition | Result |
|---|---|---|
| **CONTRADICTION** | nameID 13 matches a **refuse-signature**, or matches a **different admitted licence's** signature than the one declared | **REFUSE**, located, naming both what was declared and what the bytes say |
| **CONFIRMATION** | matches the declared licence's signature | Admit |
| **NO EVIDENCE** | no signature matches, or the name table is absent or unparseable | **Admit** |

**In simple terms.** The tie exists to catch a face travelling under **another project's** terms —
D-8.6.5, where 17 of 21 faces carried a licence that was not theirs. That is a **contradiction** between
the label and the bytes. Refusing a face whose bytes say nothing we can read catches none of that and
costs 17% of the library: it is noise, not safety. *"The bytes made no claim; they did not make a false
one."*

**The table is now two-sided, and that is a net strengthening.** Admit-signatures keyed by SPDX id:
`OFL-1.1` → `/SIL Open Font License/i`; `Ubuntu-font-1.0` → `/Ubuntu Font Licence/i`; **`Apache-2.0` →
`/Apache License,?\s+Version 2\.0/i`** (minted here; 33/33 exact on the sample, and owed because
D-16.R.4 admits the token). **Refuse-signatures, checked against every face regardless of what it
declares:** GPL / LGPL / AGPL / SSPL, and CC BY-SA / ShareAlike. Ground: **AD-26 Binds: all**, and its
Prevents is *copyleft arriving through a plausible-looking package*. **A face whose own bytes name a
copyleft licence is now refused even when `METADATA.pb` says `OFL`** — a hole the guard does not have
today and did not have under the original contract either.

**nameID 0 is a widening of where the binary speaks, not a fallback that weakens the tie.** The property
worth anything is *"the one statement of a face's licence that cannot be edited from outside the
binary"*, and that belongs to the **name table**, not to record 13 specifically. **Consulted ONLY when
nameID 13 is ABSENT. If 13 is present it alone decides** — otherwise a face whose 13 says GPL could be
laundered by a permissive-sounding 0, defeating the contradiction check with the very thing it exists to
catch.

**The ship criterion is semantic, not a number, because a number here is a licence to tune.** The guard
may ship when, on the sample, **every refusal is a CONTRADICTION and no refusal is a silence.** A
refusal traceable to *"no signature matched"* is a table gap and returns as a finding; one traceable to
*"the bytes say something else"* is the guard working.

**And it must be proved to fire, or it ships vacuous.** If the sample yields zero contradictions the
guard has never been observed to work. It therefore requires a **positive control** — a fixture face
under `folio-go/testdata/` whose nameID 13 contradicts its declared id, asserted refused — plus the
disequality control the build-time tie already uses. Red-prove by deleting the guard, never by
falsifying a condition.

**Guardrails.**
- **Go must re-read the bytes, and may never compare the wire `copyright` against the wire `licence`.**
  Both arrive from the browser, so a check over them proves nothing — both sides move together.
- **Signature matching must be order-deterministic** — an ordered slice, never a map range (AD-1).
- **Two tables in two languages is a mirror contract.** The Go table is the authority for documents and
  strictly subsumes the TS build-time table's population; **a test, not a comment**, enforces that the
  TS table admits no id the Go table refuses.
- Untyped `fmt.Errorf` per `variableface.go`; no new error type in package `folio` root.
- The build-time tie is not deleted or weakened. **Both, or halt.**

**What can and cannot be filtered before a pick**, stated so nobody builds a filter that cannot be fed:
**(a) variable-only — yes**, from the snapshot's `axes`; **(b) licence-token refusals (`cc-by-sa`,
unmapped) — no**, the index carries no licence field, so it is unknowable until `METADATA.pb` is
fetched; **(c) contradictions — no**, by construction. The browser filters (a) and states (b) and (c)
at pick time with their reasons.

**DW-150 dissolves, and the lead's own contract was the thing out of step.** `RefuseVariableFace`
returning `nil` for an unparsable face is **correct**: it asks *"does this have `fvar`"*, an unparsable
face is not proven variable, and `DecodeFontForRender`/`checkSfnt` refuses genuinely bad bytes one step
later. Under the corrected contract the nameID guard does the same thing for the same reason. **Close
DW-150 as reconciled**, with that as the recorded reason.

**How we'd know it was wrong.** An Apache family stating its terms in wording the minted regex misses —
which now lands in NO EVIDENCE and **admits**, so the cost of being wrong is a check quieter than
intended, never a document publishing false terms.

**FOR THE OWNER, flagged by the lead itself under D-000.6's backstop, as applied-and-reversible rather
than a blocking question.** This ruling **loosens a guard the lead had said must not be softened**, on
the axis the owner personally owns (D-8.5.3). What is untouched: the owner's policy on *which* licences
are acceptable, and the primary provenance guard (licence text fetched from the same family directory as
the bytes). What changed is the mechanism of a cross-check invented three messages earlier, and the net
**adds** a copyleft refusal that does not exist today. **If the owner wants silence to refuse, it is a
one-line change.**

### D-16.R.8 — RULING: the split is confirmed. The Go guard is Story 16.1b, and it lands FIRST

**Engineering lead ruling**, applied. Confidence: high.

**Verdict.** Story 16.1 carried `multiple-goals` **and** `oversized` (~14,500 tokens against a 1,600
threshold). The seam is the **Go guard**, which becomes **Story 16.1b**; the browser half keeps `16.1`.
Three-part test met: separably shippable, the only part in the **before-the-tag set**, and it carries a
decision reserved to the lead. The lead adds a fourth reason: it is the only half whose correctness is
**Go-testable against committed bytes**, so it is provable without a browser, which the other half is
not.

**16.1b runs BEFORE 16.1, and that ordering is the point.** It has no dependency on 16.1 — it tests
against the 21 committed faces and its own fixtures — while **16.1 is the first story that puts a
fetched face into a document**. *Building the gate after the population it polices arrives is how
D-8.6.5 shipped green.* Numeric order is not build order here (D-000.3; Story 16.0 is already the
precedent), so the position is stated explicitly rather than inferred.

**Pre-authorised second seam, so a further round trip is not needed:** if 16.1 is still `oversized`
after the cut, the **D-16.4 forbidden-host-scan amendment** becomes its own story — a guard rather than
a feature, separably shippable, its acceptance entirely *"the new half reds when the declaring module is
removed"*, sharing no state with the fetch path. **Do NOT cut the licence-token mapping out of 16.1**:
that is the fetch path's own admission decision, and separating it would put the fetch and its gate in
two stories — the split D-16.6 refused, for the same reason.

**Guardrails.** 16.1b carries the before-the-tag record for Story 15.3 (D-000.15's running list); its
`## Verification` includes the goldens unmoved **and** the positive control; and 16.1's `multiple-goals`
warning is **re-evaluated after the cut rather than carried forward unread**.

### D-16.R.9 — The positive control uses committed fixtures, and the refuse-signature half gets a control of its own

**Orchestrator decision** at Story 16.1b's plan gate. Routine on its face, recorded because it departs
from D-16.R.7's literal wording.

**Verdict.** D-16.R.7 requires *"a fixture face under `folio-go/testdata/` whose nameID 13 contradicts
its declared id"*. The builder measured that **no new binary is needed**:
`testdata/fonts/Roboto-Regular.ttf` already carries nameID 13 = *"Licensed under the Apache License,
Version 2.0"* and is recorded `Apache-2.0`. Declaring it `OFL-1.1` in a test is a **true** contradiction
→ refuse; declaring it `Apache-2.0` is the not-over-broad confirmation; `NotoSansThai-VF.ttf` gives the
OFL confirmation. **Accepted.**

**Why this is better than the literal reading, not a dilution of it.** What the ruling wants proved is
that the guard fires on a real contradiction. Committing a **new** font file whose sole purpose is to
lie about its own licence would trip the manifest guard, the binary-identity walk and the licence
census — three gates that exist to keep exactly that kind of file out of the repository — and would put
a deliberately mislabelled binary into a product whose whole subject is licence honesty. The existing
fixtures prove the same property with real bytes.

**The addition the gate makes, because the builder's proposal leaves half the guard unproven.** The
committed fixtures can demonstrate **contradiction-by-mismatched-admitted-signature**. They cannot
demonstrate the **refuse-signature half** (GPL / LGPL / AGPL / SSPL / ShareAlike) — no committed face
carries a copyleft statement, and none should. That half is the part D-16.R.7 added as *new*, and a new
guard with no control is the vacuity the ruling exists to prevent. **It gets its own control by
synthesising a name table in the test rather than committing a binary** — bytes built in-process, never
a file in the tree.

**Consequence.** Story 16.1b's acceptance now requires **two** positive controls: one contradiction by
mismatched admitted signature, over a committed fixture; one refusal by refuse-signature, over
synthesised bytes. Either missing and the guard has shipped half-observed.

**Call-site ordering, noted at the same gate.** `RefuseVariableFace` fires first at
`component_commands.go:2414`, and `NotoSansThai-VF.ttf` is variable — so a licence assertion routed
through `embedFontFamily` with that face would be **masked** by the earlier refusal. Outcomes are
asserted against the new door **directly**, and reachability from `embedFontFamily` is proved separately
with a **static** face.

### D-16.R.10 — A declared SPDX id with no signature entry is NO EVIDENCE, and admits

**Orchestrator decision**, applied on the fast path rather than routed, because D-16.R.4 settles it
unambiguously.

**The contradiction.** D-16.R.5 said *"an SPDX id with no signature entry is a refusal, not a skip."*
D-16.R.7's three-outcome table makes it **NO EVIDENCE → admit**, and does not name that clause as
replaced. It is reachable: D-8.5.3 admits **MIT**, for which the Go table has no row.

**Verdict: NO EVIDENCE, admits.** D-16.R.4 selects the reading in its own words — *"MIT stays
admissible and gets no table entry… absence, not narrowing"* — so refusing would reject a licence **the
owner has admitted**, which is the one thing a mechanism ruling may not do to a policy ruling.

**Why this is safe rather than a hole, and the reasoning belongs in the record.** The reachable set of
declared ids is already closed by D-16.R.4's token table: only `OFL-1.1`, `Apache-2.0` and
`Ubuntu-font-1.0` can arrive from the mapping, `CC-BY-SA` is refused and an unmapped token is refused.
**Every id that can actually reach the guard today has a signature row**, so the admit-on-absence arm is
currently unreachable in production and exists for a future token. And the **refuse-signatures apply to
every face regardless of what it declares**, so the copyleft floor does not depend on this arm at all.

**Consequence.** If a future token maps to MIT, the guard admits MIT faces with no tie until a signature
row is minted. That is a known, bounded gap rather than a surprise, and minting the row is the fix.

**How we'd know it was wrong.** A token mapping to an SPDX id whose faces then travel with no tie at
all — visible the moment the table gains a fifth admitted value without a matching signature row, which
the mirror-contract test should be extended to catch.

### D-16.R.11 — RULING: the licence tie's second door is DEFERRED PAST THE TAG, and no migration is owed

**Engineering lead ruling**, applied. Confidence: high on the deferral, **medium on "no migration
owed"**. Routed as a possible before-the-tag item and returned as a deferral **with a condition**, which
is the shape D-000.15 asks a deferral to carry.

**Verdict.** Do not add the licence tie to the load or ingestion path in Epic 16. Defer it explicitly
past `folio-go/v0.1.0`, and record that **no migration is owed**.

**Why this is not the `fvar` case, despite looking exactly like it.** D-16.6 ruled *"both, or halt"* for
`fvar` because that guard already **had** two doors, and its defect was a document that **saves cleanly
and fails at render** — a correctness break, where the bytes cannot be honoured at all. The licence tie
is a different class: the bytes **can** be honoured, the render is **correct**, and what is wrong is a
**label the hand-author wrote themselves**. And the asymmetry is not a door being removed — it is a
second door never having existed. *"Both, or halt" governs removing a guard, not declining to duplicate
a new one.*

**Why it is off-goal, on a test this run has already paid to learn.** D-000.15's operative rule is
*"already-touching, not going-looking."* Epic 16 has not touched `internal/template/parse.go`'s licence
rule — Story 8.6 did. The counter-argument, that Epic 16 **manufactured** the one-door condition, is
real; it is also **exactly the 8.4x shape**, a licence-gate finding surfaced inside a licence-gate story
and argued in on *"we are already here"*, which cost this programme 34 commits and 3,567 lines with zero
lines advancing the epic's stated goal. Epic 16's goal is *"a font arrives from the web, and stays on
this machine."* A guard on hand-authored files advances none of it, and the epic is already at seven
stories.

**And here is what actually decides it: the deadline does not bite.** The whole force of "land it now"
is D-000.14's asymmetry — free before the tag, needs a migration after. **Price that migration
honestly.** Under the three-valued contract (D-16.R.7), the only documents this narrowing could ever
break are documents whose font bytes **contradict** their declared licence. **Measured: 1.0%** — and the
product's own writer can no longer produce one. So the break-population is **definitionally the
mislabelled set**, and the "migration" is a located diagnostic saying *"this document declares terms its
own font bytes contradict."* That is **a refusal telling the truth, not a compatibility break.**
D-000.15's membership test exists because *"after the tag a narrowing needs a migration this project has
never written — that is a CATEGORY change."* **No category change happens here.**

**In simple terms.** The rule we would add later only ever rejects files that are already lying about
themselves. Rejecting a lie is not breaking someone's document; it is the document being caught. There
is nothing to migrate.

**The lead's own correction is what bought the deferral.** D-16.R.5's original *absent-or-unparseable
refuses* contract would have made this genuinely expensive — the break-population would have been every
silent face, ~17% of the library. Correcting the guard to admit silence is what made deferring it cheap.

**The trigger, and it is a CONDITION rather than a date.** Widening the table on the **admit** side stays
free forever. **Adding a refuse-signature after the tag is itself the narrowing**, because a face that
admits today as NO EVIDENCE becomes a CONTRADICTION tomorrow. The register entry therefore reads: *close
the second door before any refuse-signature is added after the tag, or accept that the addition is
itself the narrowing and price it then.*

**Guardrails.**
- Registered as **deferred past `folio-go/v0.1.0` with its trigger stated** — not ordinary backlog, and
  not "the next story touching the load path". Owner: whoever adds a refuse-signature, or Story 15.3 if
  it elects to close it inside the tag.
- **`internal/template/parse.go` carries a comment naming the asymmetry deliberately**, so the next
  reader does not close it casually as a tidy-up and thereby ship the narrowing unpriced.
- **The one door must be provably present.** Story 16.1b's test that a contradiction is refused at the
  command may not be deleted on the reasoning that the load path covers it — it does not.
- If Story 15.3 elects to close it inside the tag, that is a **scope decision for the owner**, not a
  lead ruling: it adds a story to the release gate.

**How we'd know it was wrong.** The refuse-signature set growing after the tag — which is precisely why
that is the registered trigger rather than a footnote.

### D-16.R.12 — Two standing rules the 16.1b build earned, and one declined regex widening

**Engineering lead**, unprompted, at the close of 16.1b. Recorded because both outlive the story.

**Standing rule 1 — every admit-signature row owes a CONTRADICTION control, and the row carrying the
largest population owes it most.** The 16.1b review found that the **OFL** row — the row 19 of 21
catalogue faces and the upstream OFL majority travel under — had no test that could distinguish *"admits
because it matched"* from *"admits because nothing matched"*. **Those two states are indistinguishable
at the return value**, which is the same both-sides-move-together vacuity the lead aimed at one level up
in the token table and missed one level down. The largest row is the one a narrowing breaks most
silently. **The anchored-prefix mutation — which left the whole `internal/fontset` suite green while
breaking real faces like `cascadiacode`, whose record 13 opens *"Microsoft supplied font…"* — is
retained as the red-proof rather than described.**

**Standing rule 2 — the copyleft floor is proven to fire, but NOT proven to fire on the right strings.**
No face in the 99-face sample hit a refuse-signature, so the refuse half's sole evidence is the
synthesised control D-16.R.9 required — which is exactly why it was required, and it did its job. But
the refuse-signatures' **wording** has never met a real copyleft font name table. Not a defect; a stated
limit. **Standing instruction: the first real copyleft-declaring face encountered anywhere has its
record 13 captured as a fixture.**

**Declined — do not widen the Apache regex for `apache/yellowtail`.** It states Apache terms in record 0
in wording the minted regex misses, so it **admits** (NO EVIDENCE) and costs nothing; the ship criterion
is unaffected. Two reasons to leave it: a regex fitted to one face's wording is fitted to a sample of
one; and **widening an admit-signature is not risk-free** — a *different* face declaring `OFL-1.1` whose
record 13 matches a widened Apache pattern becomes a **CONTRADICTION**, so widening the admit side can
create refusals elsewhere, over a population nobody has measured. **Recorded as observation one.
Trigger for revisiting: the same wording seen a third time, or a table-gap REFUSAL (not an admit)
tracing to it.**

**Guardrail carried to Story 16.1a**, unprompted and adopted: every face the batch adds owes the **full
committed regime** — its own `LICENSE*`, a `NOTICE.md` recording upstream release, archive digest,
committed digest, byte size and fetch date, and the digest tie — and **`font-catalogue.test.ts`'s
population floor must rise with the batch.** *"A floor left at 21 while the tier grows to 30 is a floor
that stops measuring the thing it was built to measure."*

### D-16.R.13 — RULING (DW-160): `source` stops being a URL, on both tiers

**Engineering lead ruling**, applied. Confidence: high. **Not a new provenance requirement — one tier
failing a contract the other tier already honours.**

**The measurement that decides it.** Both tiers write `source`, and they disagree:

| Tier | Writes | Shape |
|---|---|---|
| Committed (`build-wasm.mjs:302`) | `folio-designer/public/fonts/<dir>/<file> — see that directory's NOTICE.md for the pinned upstream release and digest` | a path plus a pointer to a fixed-point record |
| Fetched (`font-source.ts:338`) | `<host>/google/fonts/main/<dir>/<slug>/<file>` | **a bare mutable branch URL** |

And `font-catalogue.md` already defines the field as *"upstream release and the derivation that produced
the shipped instance"* — **a release, not an address.** The fetched tier does not meet its own field's
definition.

**Verdict.** `source` carries the **upstream project**, the **path within it**, and the **fetch date**.
Three things it must not carry:
- **Not a resolvable-looking URL.** A string shaped like an address is a promise of fetchability, and a
  promise that decays is worse than none — a dead link in a year reads as *"this provenance is broken"*
  when the provenance is in fact intact.
- **Not the SHA-256.** It is already the asset key, and the key travels with the record when an asset is
  lifted into another document (D-8.6.1's own scenario). Duplicating it creates **two authorities on one
  fact that can disagree**.
- **Not `main`, or any branch name.** That is the defect.

**Option 1 (pin the commit SHA) is REFUSED, on a measurement.** `raw.githubusercontent.com` returns **no
commit ref** — its `etag` is a content SHA-256, not a commit. Obtaining one requires `api.github.com`, a
**fourth host**, against D-16.1's *"three hosts, and no others"* — and **the lead will not amend an
owner's host allowlist by ruling.** The trade is poor regardless: a pinned commit still decays (history
rewrite, repo removal) and buys no licence audit unless `OFL.txt` and `METADATA.pb` are pinned at the
same commit too, which one string cannot express. Recorded as considered-and-rejected so it is not
re-derived.

**The cost, stated rather than hidden.** A recipient cannot refetch the exact bytes. Acceptable because
**they do not need to**: CAP-2 puts the face inside the file, the asset key pins its identity, and
`licenceText` and `copyright` travel with it. **`source` names provenance, not a retrieval path**, and
should read as one.

**Story 16.1a carries both halves, because it is already touching the field.** The committed tier's
current string points at a `NOTICE.md` **the recipient of a `.folio` does not have** — honest but
incomplete — and a reader cannot tell which tier a face came from, so if the two conventions differ in
*kind* the field is uninterpretable. 16.1a **inlines the pinned upstream release and the committed
digest** instead of pointing at a file that does not travel.

**Guardrails.** `source` stays a `string` in the twelve wire fields — **no arity change**. **Verify, do
not assume, that no golden moves**: check whether any `.folio` fixture feeding a golden embeds a face
rather than reasoning it through. **A test asserts `source` contains no scheme and no host on either
tier** — that is the tripwire, and the convention alone will not hold. Record on D-000.15's running list
for Story 15.3.

### D-16.R.14 — RULING (DW-165): the fetch timeout is pulled into Story 16.2, and the contrast with D-16.R.11 is deliberate

**Engineering lead ruling**, applied. Confidence: high.

**Verdict.** Pull it in. **And the contrast with the immediately preceding deferral is put on the record
so the line reads as principled rather than convenient.** D-16.R.11 deferred the load-path licence tie
because it guards **hand-authored files, which Epic 16 does not create** — off-goal. This is the
opposite: **16.2's stated subject is what happens when a fetch does not give you bytes**, and its own
I/O matrix already carries **network-down** and **store-failure** rows. **A stall is the one member of
that class it does not cover, and it is the worst** — a rejection degrades with a stated message; a
stall leaves the control **dead for the session with no message at all**. A degradation matrix missing
its silent arm is not a scope addition to fix; it is **the matrix being incomplete**.

**Guardrails.**
- **A timeout is a number, and a number needs a basis.** Derived from something stated — a measured
  fetch time against the real host, or an explicit *"chosen because"* — never a magic constant, and the
  measurement carries command, commit, tree state and working directory (D-8.4j.8).
- **The hold must be proven to release**, by a test that reds when the release is removed.
- The timeout releases the hold and states the degradation. **It must not silently retry** — a retry
  over a deterministic stall hides it (Story 16.0's `Never:` clause, same reasoning).

### D-16.R.15 — PATTERN, named because it has now fired three times in two stories: a fix for a state-lifetime bug is the likeliest site of the next one

**Engineering lead**, unprompted, recorded as a standing expectation rather than a finding.

Story 16.1's review patched a state-lifetime defect — the pick's re-entry guard released **before** the
new work rather than around it, so two overlapping picks could both commit an embed. **The patch then
introduced the same class of defect**: the guard was backed by a ref, but `setCurrentSnapshot`'s
document-reset path cleared only the **state** copy, so replacing the document mid-pick left the hold
stuck `true` for the session — the control looking enabled while **every later pick silently did
nothing**. The story's closer caught it and fixed it in one line with a test that reds without it.

**Why it recurs:** *"the author is holding exactly one lifetime in mind while adding a second carrier
for the same state."* Third instance in two stories.

**Standing consequence.** Wherever a hold, a busy flag or a generation guard is added or repaired, its
**release path is red-proved rather than asserted** — a test that fails when the release is removed. That
is the reason D-16.R.14's guardrail demands it explicitly rather than trusting the implementation.

### D-16.R.16 — RULING + MEASUREMENT: the batch is sized by GOAL, the squeeze dissolves, and the criterion with no budget term is replaced

**Engineering lead ruling**, applied, with both demanded measurements taken by the orchestrator.
Confidence: high on the criterion; part 1's outcome measured rather than assumed.

**The lead declined to supply N and replaced the criterion instead.** *"The most popular N that fit"* is
**a criterion with no budget term**: it says yes until the resource runs out, and will consume whatever
margin exists on any run, in any epic, forever. That is the shape that cost this run the 8.4x series —
34 commits and 3,567 lines against a criterion admitting every member of its class.

**The criterion.** The batch is **the refused families within the top 20 by `popularity`**, minus CJK
(standing non-goal), minus families already in the local tier, minus `shippedFamilies` collisions, minus
anything with no obtainable static from its own project upstream. **Whatever that yields is the batch.**
If it fits the margin, the reserve is a **residue and no allocation policy is needed**. **If it
overflows: halt and return the overflow — do NOT truncate to fit**, because truncation-by-resource
silently converts a goal-bounded set back into a margin-bounded one.

**Measurement 1 — is the precache slot actually the binding cap? It is a SEQUENCING fact, not an
engineering unknown.** The lead asked whether a face made fetch-on-first-pick by D-8.4d.1 still needs a
**precache** slot. Measured: **`15-0-a-catalogue-face-arrives-when-it-is-picked` is `backlog`.** D-8.4d.1
is a **policy** decision whose implementing story has not shipped, so catalogue faces **are** precached
today, the 20-slot margin **is** real today, and it dissolves when 15.0 lands. Also confirmed:
`maximumCacheAssets = 64` is **self-imposed** — `release-payload.ts:26-33` states it exists so *"a
release over the bound can no longer be emitted in silence"*, and `static-host-contract.json` sets **no
asset-count limit**. A device for forcing a decision, not a physical ceiling — which is why the owner
may move it and the lead may not.

**Measurement 2 — the goal-set is ELEVEN families, against twenty free slots.**

> Open Sans · Montserrat · Lora · Roboto Condensed · Arimo · Roboto Mono · Oswald · DM Sans · Nunito ·
> Raleway · Fraunces

**Excluded, itemised so a later reader can see what was left out and why:** `Roboto` and `Inter` already
in the local tier · `Poppins`, `Lato`, `Bebas Neue` static upstream and never refused · `Noto Sans JP`
CJK · `IBM Plex Sans` the chrome collision · `Noto Sans` takes the predicate fix at zero slots ·
`Google Sans` dropped on evidence (below).

**11 into 20. The reserve is a residue and there is nothing to allocate.** No tie-break is needed
either: the N = 20 cut that made `Instrument Sans` and `Rubik` straddle a boundary at `popularity` 22
does not arise. **No owner question about raising the cap arises.**

**`Google Sans` dropped on evidence, and the class it exposes is registered.** `isBrandFont: true`, and
it **404s in all three licence directories** of `google/fonts` (`ofl/`, `apache/`, `ufl/`). No obtainable
static, no fetchable licence text — it drops by D-16.R.2a's own mechanism. **The class matters more than
the family:** the index lists **1,946 families and `google/fonts` does not carry all of them**, and
nothing measures the difference. **`isBrandFont` is not the signal** — Roboto and every Noto carry it and
are all present. So there is an **unmeasured population** the browser lists, filters as addable, and then
fails at pick time when all four directory probes miss. Story 16.1's slug-and-confirm refuses that
correctly, so it is not a correctness hole; it is registered as deferred work rather than sized now.

**Correction carried, at the lead's own request.** D-16.R.2's *affordability* argument — that the
payload is not the binding constraint because D-8.4d.1 moved faces to fetch-on-first-pick — is **true of
`brotli.totalBytes` and false of slots**. Recorded as a correction to the argument, **not to the
verdict**: the owner's choice to derive a bounded batch stands; what was wrong was the account of what
bounds it.

**Guardrails.** The ranking stays a ranking so popularity movement cannot silently change membership; the
goal-set is stated in the story **with its exclusions itemised**; `maximumCacheAssets` is **not edited in
this epic under any circumstance**.

### D-16.R.17 — RULING: `Noto Sans` and `IBM Plex Sans` are two different problems sharing a symptom

**Engineering lead ruling**, applied. Confidence: high. **`shippedFamilies` conflates two populations,
which is why these looked like one bug.** Measured: `folio-go/fonts/fonts.go:59-61` — the engine's
shipped set is **exactly** `Noto Sans`, `Noto Sans Thai`, `Noto Sans SC`. **IBM Plex is not in it**; it is
the designer's own chrome typeface.

**`Noto Sans` — a DEFECT against Story 16.1's own shipped acceptance criterion.** It is an **engine
render face**: every document can already name it in a chain, and it needs no embedding because it ships
with the renderer. It is not *"present and unofferable"* — it is **present and always available**. 16.1's
AC reads: *"a row is refused because no static face is obtainable, never because upstream happens to be
variable."* The browser is refusing `Noto Sans` on the `axes` field alone, which is exactly what that AC
forbids.
- **Fix the predicate, not the batch.** The availability check must ask *"is a static face obtainable for
  this family"* — from the **engine's shipped set**, the local tier, or upstream — not *"does the index
  say it has axes"*. Same authority ordering as D-16.R.1: **`axes` predicts, obtainability decides.**
- **Zero slots**, and it recovers a top-20 family.
- **No family name is special-cased**, and a test asserts a shipped-set family with upstream axes is
  offered.

**`IBM Plex Sans` — register it, do not build it.** It is designer **chrome**, absent from the engine's
`FontSet`, so a document genuinely cannot use it. But *"the product has this font"* overstates it: the
product has it **for its own interface**, which is a different relationship. Making it embeddable means
promoting a chrome asset to a document asset — a new relationship carrying its own
`licenceText`/`copyright` obligations, and a decision about whether the designer's UI typeface should
travel inside users' documents at all. **A story, and off Epic 16's goal.** Same discipline as the
second-door deferral: repair what is the goal, register the rest.
- **Trigger:** an author asking for IBM Plex specifically, **or** Epic 11 touching `shippedFamilies` —
  whichever comes first. Owner: whoever next changes that list.
- **The register entry states the broader finding:** `shippedFamilies` **conflates engine render faces
  with designer chrome faces, and the two have opposite availability semantics.** The two populations are
  not merged or split in this epic, only documented.

### D-16.R.18 — Two corrections the lead made to its own earlier rulings, recorded rather than absorbed

**"One place" was wrong, and the lesson generalises.** D-16.R.12 said *"a floor left at 21 while the tier
grows to 30"*. Measured: **there is no 21 anywhere** — the floor is `>= 20` in **three** files
(`font-catalogue.test.ts:292`, `font-index.test.ts:107`, `font-name-table.test.ts:25`), two of which
Story 16.1 added **after** that ruling was written. The substance held; the wording asserted a fact
nobody had checked. All three are pinned by task. The lead's own generalisation is worth keeping:
**"a floor that exists in three files is three floors, and a ruling that says *the floor* has already
lost track of one of them."**

**The `axes` trap, recorded because a wrong reading would have made this epic look unnecessary.**
`font-index.json` stores **`axes`, not a synthesised `variable` flag** — 0 of 1,811 rows carry a
`variable` key. Reading it the wrong way yields a confident *"0 of the top 50 are variable-only"*, which
is the **dangerous** kind of wrong: it does not look like an error, it looks like the work is
unnecessary. Written into Story 16.1a's Design Notes.

**A tripwire nobody would expect:** a fourth `UPSTREAM` entry reds
`folio-go/fontgen_matrix_test.go:117`, whose witness string hardcodes *"derived and compared 3 of 3
faces"*. Under D-16.R.2a the batch takes upstream statics rather than deriving, so it **should not**
bite — noted in the story anyway, because *"should not bite"* is how tripwires get discovered the
expensive way.
