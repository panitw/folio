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

*Filed by the orchestrator from the engineering lead's grounding report. The original placeholder said
this section would be "filled in before the first plan dispatch"; it never was. That is itself worth
recording rather than quietly overwriting — the run's first lead grounded and ruled through thirty-one
decisions with its report never filed, so its reasoning survives only as the rulings it produced. The
refresh below is the FIRST content this section has ever carried.*

### 2026-09-03 — refresh, on resuming the run with 16.2/16.3/16.4 outstanding

The run's first lead did not survive its session. A replacement was grounded from THIS LOG and the three
remaining specs rather than re-derived from the ADRs, per the resumption rule: re-ground from the record,
verify only what has plausibly changed. Verified at `8a9e297`, tree clean.

**Confirmed unchanged since the log was written:** the generated catalogue is 31 families (21 + 10);
`familyIndex` is 1,811 rows with 537 `variable: true`; `font-licence.ts`, `font-name-table.ts`,
`font-source.ts` and `font-index.ts` are all present; SPEC-fonts carries both required amendments; the
Non-goal clauses *"No host fonts"* and *"no variable-font axes"* survive untouched.

**Confirmed about CI, which narrows the local-only debt precisely.** CI runs the designer's
test/typecheck/lint/build, `verify:offline:red`, `verify:offline:wasm` and `test:e2e:compile`, and every
Go invocation with `-count=1`. It does **not** run `playwright test`, and it only `go build`/`go vet`s the
`matrix`-tagged suites. So the browser and matrix legs are genuinely local-only debt, and D-16.R.31's
correction to DW-168 (CI *is* watching `-count=1`) is confirmed a second time by a second reader.

**The orchestrator's own verification of the report, before acting on it.** Load-bearing numbers were
re-measured rather than relayed: the six `role="presentation"` children (1456, 1462, 1463, 1469, 1470,
1474); `IndexFamily` = `{family, category, scripts, variable, popularity}`; `maximumCacheAssets = 64`;
`FontFamilyProperty` at 1385; `pickCatalogueFamily` at 660; the decline sentence at 1474 with
`familyIndexDisclosure()` at 1469; 1,811 rows / 537 variable. All hold.

**One correction the orchestrator made to the report.** The rot table rendered the designer anchors as
`src/App.tsx`, which reads as a repo-root path that does not exist (the file is at
`folio-designer/src/App.tsx`). The specs' Code Maps group anchors under module headings — **`Designer
(`folio-designer/`)`** — so `src/App.tsx` is CORRECT as the specs write it and must not be rewritten.
Only the line numbers are stale. Recorded because D-16.R.28 corrected a genuine path error in 16.2
(`internal/template/` → `folio-go/`), and that precedent must not license rewriting a prefix that is right.
The near-miss is the point: two readers looked at the same anchor and one of them saw a defect in a
notation.

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

### D-16.R.19 — CORRECTION (orchestrator error): the batch was computed from the LIVE index, not the committed snapshot

**Orchestrator error, caught by the build dispatch, which refused to build on it.** Recorded in full
because the refusal was correct and the mistake is instructive.

**What happened.** D-16.R.16 recorded an eleven-family goal-set — *Open Sans, Montserrat, Lora, Roboto
Condensed, Arimo, Roboto Mono, Oswald, DM Sans, Nunito, Raleway, Fraunces*. I computed it by fetching
`fonts.google.com/metadata/fonts` **live**. Story 16.1a's TASK 1 requires the survey to be
**reproducible offline from committed artifacts**, and the committed snapshot is
`folio-designer/font-index.json` — 1,811 families, `snapshotDate` 2026-09-03. **The two orderings
differ**, so my set could not be reproduced, and the dispatch **halted rather than reconciling it**.

**The dispatch's refusal was exactly right, and its reasoning was sharper than a mismatch report.** It
observed that `popularity` is monotone, so any cut admitting **Raleway (17)**, **Nunito (18)** and
**Fraunces (19)** must also admit three strictly more popular refused families that survive every stated
exclusion — **Roboto Slab (11), Plus Jakarta Sans (14), Jost (16)** — which appeared **neither in the
set nor in the itemised exclusions**. It also noted the nearest rule producing exactly eleven
(`popularity <= 20`) yields a *different* membership, so **the count matching was a coincidence, not a
confirmation**.

**Two of my stated exclusions had no referent in the committed data**, which the dispatch also caught:
- **`Noto Sans JP` / CJK.** The snapshot **already excludes CJK at build time** — `excludedCjkFamilies:
  135`, and `1946 − 135 = 1811`. The CJK subtraction removes nothing and cannot be why a slot came free.
- Four never-refused statics (`Black Ops One`, `Archivo Black`, `Share Tech`, `Barlow`) were unitemised.

**THE CORRECTED SET, reproducible from the committed snapshot.** Top-20 positions by `popularity`
(name-ascending tie-break), refused = `axes` non-empty, minus already-local, minus `shippedFamilies`:

> **Open Sans · Roboto Mono · DM Sans · Montserrat · Arimo · Roboto Slab · Lora · Roboto Condensed ·
> Oswald · Plus Jakarta Sans · Jost**

That is **twelve candidates, minus `Google Sans` on evidence, leaving ELEVEN.** Coincidentally the same
count as the wrong set and a **different membership**: `Raleway`, `Nunito` and `Fraunces` are out;
`Roboto Slab`, `Plus Jakarta Sans` and `Jost` are in. **11 into 20 free slots — the goal-set still fits,
so D-16.R.16's verdict and criterion are unaffected. Only its membership list was wrong.**

**Exclusions within the top 20, itemised against the committed data:** `Roboto` (pop 2) and `Inter` (4)
already local · `Lato` (6), `Black Ops One` (7), `Poppins` (7), `Archivo Black` (12) static upstream and
never refused · `Noto Sans` (8) and `IBM Plex Sans` (15) `shippedFamilies` collisions · `Google Sans`
(3) unobtainable.

**The lesson, and it is the one this run keeps re-learning.** A measurement taken against a live source
cannot gate a story whose acceptance requires offline reproducibility. The two agreed closely enough to
look right — the same count, most of the same names — which is precisely why it survived my own review.
**A build dispatch that halts rather than reconciling a mismatch it did not create is the control that
caught it.**

### D-16.R.20 — MEASURED (per the lead's ruling): the index-vs-repository gap is 0.4%, so 16.3's header keeps the snapshot count

**Engineering lead ruling** — measure the gap **inside this epic**, because it has a consumer three
stories away: Story 16.3's header prints *"the snapshot's count"*, and D-16.3 already put this epic on
record that describing the product in untrue terms is a defect rather than a wording preference. If a
material share of the families cannot be obtained, that header overclaims a second time.

**Bounded to a measurement**, at the lead's cheap shape: enumerate the four licence directories **once**
each via the git-trees API and intersect slug sets — ~5 requests, not the ~7,800 a per-family probe
would cost. The lead explicitly reopened a door it had shut: **refusing `api.github.com` at D-16.R.13
was about a PRODUCT RUNTIME DEPENDENCY, not about what an engineer may measure** — and
`forbidden-font-hosts.mjs`'s own header already draws that line, bounding the *scanned source
population* rather than anyone's measurements.

**Result** (wd `/Users/panitw/Projects/folio`, tree clean at `d86d547`, 2026-09-03): repository slug
population **2,049** across `ofl` 1,999, `apache` 44, `ufl` 5, `cc-by-sa` 1. Index families **1,811**.
**Not present in `google/fonts`: 8 — 0.4%.** Within the top 50 by popularity: **one**, `Google Sans`.

**And the eight are two different things, which matters more than the number.** Six are `Edu …`
families — and they are **not a slug-rule hole**. Checked directly: upstream **renamed** them
(`Edu QLD Hand` → `eduqldbeginner`, `Edu SA Hand` → `edusabeginner`, `Edu NSW ACT Cursive` →
`edunswactfoundation`), while the index still carries the older display names. **D-16.R.6's slug rule is
sound**; what drifted is the index against the repository. The other two are Google brand fonts
(`Google Sans`, `Google Sans Flex`) genuinely absent from the repo.

**Consequence, per the ruling's own disposition: the number is SMALL, so 16.3's header keeps the
snapshot count and nothing is built.** No build-time pre-flight probing, no snapshot enrichment with an
obtainability flag, no new filter. The residual post-pick failure class stays registered as deferred
work, with `Google Sans` named as instance one so the count starts rather than being re-derived. **A
mechanism would come back to the lead as its own decision; the measurement says none is warranted.**

### D-16.R.21 — Story 16.1a records its payload delta and hands it to the budget gate

**Engineering lead guardrail**, applied — surfaced by the measurement in D-16.R.16 rather than
anticipated.

**The fact.** Story 15.0 is `backlog` and unstarted, so **D-8.4d.1 is a policy with no implementation**.
16.1a therefore adds 11 faces to a tier **whose precaching is scheduled for removal but has not been
removed** — a live **first-load payload increase**, and payload is owned by Story 8.4d/15.2's budget
gate, not by 16.1a.

**So 16.1a measures its payload delta and hands it to that gate rather than adding it silently.** This
is D-000.15's obligation shape: the story that spends records what it spent, and the story that owns the
total collects it. Without it, **15.2 meets a budget that moved for reasons its own record does not
contain** — the same defect as a golden re-recorded without a reason.

**Two consequences noted rather than acted on.** If 15.0 lands after 16.1a the margin grows and the
batch could have been larger — **which is the criterion working**: a goal-bounded batch is insulated
from a constraint scheduled to disappear, while the margin-bounded one the lead declined to supply would
have sized itself to a ceiling with a removal date on it. **That is the cheapest available evidence that
replacing the criterion was right rather than merely fussy.** And sequencing 15.0 before 16.1a would
make the payload question vanish entirely — the orchestrator's call, raised because the dependency now
runs both ways and nothing in the sequence records it.

### D-16.R.22 — Two number corrections, and the pattern they share

**Recorded together because they are the same failure at two removes: a figure quoted forward without
being re-derived, surviving because it looked plausible.**

**Correction 1 — "37 of the top 50" was 38, and it reached the owner.** The engineering lead's grounding
report gave **37 of the top 50 families are variable-only**, and that figure went into the escalation the
orchestrator put to the owner for D-16.R.2. Re-derived on the same file the lead measured from (1,946
families), the answer is **38**. **Story 16.1a's plan-gate table said 38** and sat in front of the
orchestrator for two dispatches without being reconciled. The lead self-reported it.

**The substance is unchanged and marginally stronger** — the refusal is concentrated at the head of the
distribution, and it is 38 of the top 50 rather than 37. **D-16.R.2 does not move.** The record carries
38, and it records that the correction came from the gate table rather than from anyone re-deriving it.

**Correction 2 — the corrected eleven were queried, and the query came from the live index again.** The
lead computed the top 20 from the **live** `fonts.google.com` index and found `Raleway` (17) and
`Nunito` (18) inside it while `Jost` sat at `popularity` 42 — concluding the batch should be twelve and
that the monotonicity argument that killed the first list applied to the second. **It flagged its own
caveat: it was reading live, not committed.** That caveat was the whole answer.

**Measured against `folio-designer/font-index.json`:**

| family | `popularity` | sort position |
|---|---|---|
| **Jost** | **16** | **20 — the last family inside the cut** |
| Raleway | 17 | 21 |
| Share Tech | 17 | 22 |
| Bebas Neue | 18 | 23 |
| Nunito | 18 | 24 |

**The eleven reproduce exactly from the committed file.** The two sources disagree materially — the live
index puts `Jost` at 42 where the committed snapshot has 16 — consistent with the snapshot being
regenerated on 2026-09-03 with 135 CJK families removed. Nothing more popular than `Jost` is silently
admitted, so the monotonicity objection does not bite.

**The guardrail objection stood even though the membership did not**, and it is being acted on.
D-16.R.16 requires the goal-set stated *"with its exclusions itemised, so a later reader can see what was
left out and why"*. `Raleway` and `Nunito` appeared in **neither the batch nor the exclusions** — they
vanished silently, which is exactly what that guardrail exists to prevent. The build dispatch is
itemising **Raleway (21), Share Tech (22), Bebas Neue (23), Nunito (24)** as *"outside the top-20 cut, at
position N"*, and **stating Jost's position beside its inclusion** so the boundary is visible rather than
implied. The spec was not edited by the orchestrator while the dispatch held the file — editing
underneath a running dispatch is how this run lost unstaged work once already.

**The pattern, stated because it has now produced three instances in one epic.** A number is derived
once, quoted forward, and survives review because it is plausible and roughly right: **37 vs 38** in an
owner escalation; **the eleven computed from the live index** when the story required offline
reproducibility; and **the count coincidence** — the wrong list and the right list both containing
eleven — which nearly confirmed the wrong one. In each case the correction came from something
**re-deriving the number against the committed artifact**, never from anyone re-reading the prose.
**Standing consequence: a figure that gates a decision is re-derived at the gate, from the artifact the
story's acceptance names, not carried from the message that first reported it.**

### D-16.R.23 — The batch is pinned to the snapshot's commit, because the ranking field churns a quarter of the batch in a day

**Gate outcome: 16.1a's twenty are AGREED as proposed** — Open Sans, Google Sans, Roboto Mono, DM Sans,
Montserrat, Arimo, Roboto Slab, Lora, Roboto Condensed, Oswald, Plus Jakarta Sans, Jost, Raleway, Nunito,
Fraunces, Heebo, Nunito Sans, Playfair Display, Libre Baskerville, Instrument Sans — with the membership
rule and its explicit family-name tie-break as written.

**The lead asked for one guardrail: pin the batch to the snapshot *commit*, not the file.** Its argument
was that "the boundary is one position wide — Jost in at 20, Raleway out at 21 — and the values
demonstrably move day to day," so `reproducible from font-index.json` is not a reproduction contract
without a sha. Adopted. TASK 1 now reads **reproducible from `folio-designer/font-index.json` at commit
`d6d51f16988cddf20d1a28697cd556b3d0a63f62`**, and 16.1a records that sha beside the membership.

**Measured, because the justification turned out to be much larger than the argument for it.** Applying
16.1a's own membership rule to a live `fonts.google.com/metadata/fonts` fetch taken on the snapshot's own
nominal date:

| Measurement | Value |
|---|---|
| Families whose `popularity` differs, live vs. committed snapshot | **946 of 1,811 — 52%** |
| Members of the top-20 that differ | **5 of 20** |
| Out, against live | Jost, Plus Jakarta Sans, Heebo, Libre Baskerville, Instrument Sans |
| In, against live | Rubik, Dancing Script, Manrope, Merriweather, Public Sans |

The boundary is not one position wide. It is **a quarter of the batch** wide. This does not weaken the
proposal — it vindicates its choice of authority. The committed snapshot is the correct input *because*
it is pinned and the live field is not: **AD-22** (determinism) and **AD-19** (offline is a build
artifact) selecting the reproducible source over the current one. The snapshot is not a stale
convenience; it is the only version of this list that is reproducible at all.

**Two superseded explanations, both recorded because both were stated with confidence.**

1. **"A dense re-rank after the CJK removal"** (offered in D-16.R.22's entry above as *"consistent with
   the snapshot being regenerated with 135 CJK families removed"*) is **false**.
   `scripts/build-font-index.mjs:75` copies `entry.popularity` verbatim; the CJK filter removes rows and
   never renumbers them. No re-rank was ever on the table for either file.
2. **"Only three families moved"** (the lead's) is **a sample claim stated as a corpus claim**. The three
   were the three that mattered to the list under discussion. The corpus figure is 946.

**And `popularity` is not a rank at all** — 1,811 families carry 1,100 distinct values, min 2, max 1373.
It is a sparse, tied score. That is *why* ties straddle the cut, and why the name tie-break in the
membership rule is load-bearing rather than pedantic.

### D-16.R.24 — Two standing practice rules, adopted for the rest of the run

Both proposed by the lead in the same message that carried its own instance of the first. Adopted:

1. **A figure restated downstream carries a pointer to the measurement that produced it.** D-8.4j.8
   governs *taking* a measurement; nothing governed *quoting* one, which is the gap all four instances in
   this epic fell through. One clause closes it.
2. **Agreement on a count is not agreement on a set — verify membership, not size.** The first eleven
   matched in cardinality and differed in membership, and the match read as confirmation. Cardinality is
   the weakest available witness and the one most likely to be checked.

These sit alongside D-16.R.22's standing consequence (re-derive a gating figure at the gate) rather than
replacing it: that rule says *re-derive*, these say *what to carry* and *what to compare*.

### D-16.R.25 — A third rule, added because the first two would not have caught the error that prompted them

The lead verified D-16.R.23's structural claims against the artifacts rather than accepting them
(`n = 1811`, 1,100 distinct values, min 2, max 1373, `sorted(set(p)) != range(1, n+1)`; and
`build-font-index.mjs:75` reading `popularity: entry.popularity ?? 0`), then diagnosed its own error more
precisely than D-16.R.24 does.

**Its error was not a figure travelling.** It computed the number itself, correctly, in the message where
it stated it — so D-16.R.24's rule 1, which governs quoting a figure computed elsewhere, does not reach
it. The number was **true of what was measured and false of what was claimed**: nine families, selected
*because* they sat near the membership boundary, with the movement rate then reported as a property of two
1,811-family files. That is **sampling on the dependent variable**. Hence:

3. **Before any sentence of the form "only N of them…", ask what the denominator is and whether it is the
   population being described.** A subset chosen for relevance is never a neutral sample of the corpus it
   was drawn from.

**The worse half, recorded because the outcome concealed it.** A *mechanism* was inferred from that biased
sample — *"not a dense re-rank; Raleway would have shifted by one and did not"* — where one grep of
`build-font-index.mjs:75` settles it. **The conclusion happened to be correct** (upstream did recompute),
which is exactly why it survived review: *a right answer reached by bad reasoning is never caught by
checking the answer.* The governing rule — read the code that implements the mechanism before reasoning
about the mechanism — was already on the books. This was a failure to apply, not a missing rule, and three
of the four rules now in this log would have caught it.

**The positive exemplar, worth naming because this run has now mishandled the same shape four times.** The
comment above `build-font-index.mjs:73` states its heuristic *with its own sample size and its own limit*:
*"`axes != []` is the only signal available here, it is a good heuristic (verified on Roboto and six
others), and the authority stays Go."* A shipped artifact carrying its confidence level and naming where
final authority lives is the correct treatment of precisely the kind of claim these four rules exist to
discipline. Prefer that form.

### D-16.R.26 — CORRECTION to D-16.R.23: I gated a superseded table, and the churn figure was measured on the wrong cut

**What went wrong.** I read `16-1a-batch-proposal.md` as far as its candidate table, ruled the twenty
families in it agreed, pinned them, and logged the ruling as D-16.R.23. I did not read as far as the
document's own heading **"Consequences for the table above, stated so this document is not read as still
governing"**, which records D-16.R.16 replacing the sizing criterion. The governing criterion is *the
refused families within the **top 20 by `popularity`***, not the top-20-admissible-of-the-top-50 the table
enumerates. Under it the goal-set is twelve, `Google Sans` is out on evidence, `Jost` dropped for cause at
TASK 2, and **ten landed**. The implementation was right; my gate was against a question that had already
been retired.

**Withdrawn from D-16.R.23:** the twenty-family membership, the `Instrument Sans`/`Rubik` tie-break at
N = 20 (that cut does not arise, so no tie decides membership), the "32 candidates want 20 slots" framing,
and the 5-of-20 churn figure as stated — it was computed over the superseded cut.

**Retained and re-derived.** The 946-of-1,811 `popularity` drift is criterion-independent and stands.
Recomputing membership churn over the cut that governs — index positions 1–20, snapshot vs. a live fetch:

| | families |
|---|---|
| Leave the cut under live ordering | `Jost`, `Plus Jakarta Sans`, `Roboto Slab`, `Archivo Black`, `Black Ops One` |
| Enter it | `Raleway`, `Nunito`, `Fraunces`, `Nunito Sans`, `Bebas Neue` |

**This is a much stronger argument for the pin than the one I withdrew, because of *which* families move.**
`Jost`, `Plus Jakarta Sans` and `Roboto Slab` are exactly the three D-16.R.19 *added* when it recomputed
from the committed snapshot; `Raleway`, `Nunito` and `Fraunces` are exactly the three it *removed*. The
entire D-16.R.19 correction is the snapshot-vs-live difference and nothing else. Choose the live index and
the batch inverts on precisely the axis this epic already litigated once. `d6d51f1` is a reproduction
contract, not provenance decoration.

**And the coincidence is itself an instance of D-16.R.24 rule 2.** Both the withdrawn measurement and the
correct one came out at **5 of 20**. The count agreed; the sets are disjoint. Had I checked the recomputation
against the old figure by its cardinality — the natural move — the wrong number would have confirmed itself.
*Agreement on a count is not agreement on a set*, demonstrated in the same run that adopted the rule.

**The failure mode was mine and it is the one this log has been cataloguing.** D-16.R.25 names reasoning
about a mechanism without reading the code that implements it. This is its sibling: **ruling on an artifact
without reading to the end of it**, where the part left unread was the part that said the part I read no
longer applied. Four instances now share the shape — a claim true in the scope where it was produced,
applied in a wider or later scope where it is not. **Standing consequence: a gate ruling states where in the
artifact the governing criterion was found, so that "I read enough of it" becomes a checkable claim rather
than an assumed one.**

### D-16.R.27 — The four rules are one rule: scope travels with the claim

The lead confirmed the landed batch against the tree rather than the report (`font-catalogue.json` is
**31 entries — 21 + 10** at `4aca77f`/`efd0ed1`) and noted that `Jost` falling out at TASK 2 was
incidental to its earlier objection about `Jost`'s rank, not a vindication of it: it fell for an unrelated
cause.

**The generalisation, adopted.** D-16.R.24's two rules, D-16.R.25's third and D-16.R.26's standing
consequence are not four rules. Every instance in this epic was **a claim true in the scope where it was
produced and applied in a scope where it was not**:

| instance | true of | applied to |
|---|---|---|
| "37 of the top 50" | an earlier session's measurement | an owner escalation |
| "only three families moved" | nine families chosen for sitting near the boundary | two 1,811-family files |
| the first eleven | the live index | a contract requiring offline reproducibility |
| the twenty-family gate ruling | a candidate table | a document whose later heading retired that table |

So the rule is **scope travels with the claim**, and *carry a pointer to the measurement*, *state your
denominator*, and *name where you found the governing criterion* are three instances of it. Logged this
way deliberately: the next occurrence will be a fifth artifact type, and a list of four would not cover it.

**The specific consequences are kept, not replaced.** A rule naming a checkable artifact location is worth
more than a general exhortation, so D-16.R.26's standing consequence — *a gate ruling states where in the
artifact it found the governing criterion* — stays scoped to gate rulings and is not broadened. The
general form governs; the specific forms are what anyone can actually check.

**And the confirming coincidence stays on the record** as the sharpest available illustration: 5 of 20
twice, over disjoint sets. Verifying the recomputation by its cardinality would have confirmed the wrong
answer with the right number.

### D-16.R.28 — 16.2's plan gate: the contract passes, the Code Map had rotted

**Gate outcome: 16.2 passes.** `warnings: []` (no `multiple-goals`, not `oversized`), the
`<intent-contract>` matches D-16.2 and D-16.R.14 as written, the plain-terms opener is present and
honest about what *"on this machine"* means, and `## Verification` carries the right commands for the
run's end-of-run heavy cadence. **Per D-16.R.26's standing consequence, the governing criterion was read
at lines 15–180 — the whole file, contract through Verification — not sampled.** The contract is
unedited: 5,567 bytes, md5 `18028ef0e2d186459cb474ed3c125ac8`.

**What the gate caught.** Every line-anchored Code Map reference was stale except one. 16.2 was planned
at `baseline_commit: a40c34d`; 16.0, 16.1, 16.1b and 16.1a have landed since.

| cited | actually | now |
|---|---|---|
| `src/App.tsx:186-224` carried-face registration | a block of `useRef` declarations | `216-243` |
| `src/App.tsx:608-627` `pickCatalogueFamily` | `applyFontChain` | `660` |
| `internal/template/component_commands.go:2359-2410` | **path does not exist** | `folio-go/component_commands.go:2360-2410` |
| `src/release-payload.ts:33` `maximumCacheAssets = 64` | correct | unchanged |

Corrected in place, all four **outside** `</intent-contract>`, with a note in the spec telling the
implementer to re-verify before relying on them.

**Why this is worth a decision entry rather than a silent fix.** A Code Map is the part of a spec that
**rots silently**, because a wrong line number still reads as a right one — nothing in a stale anchor
announces itself, and an implementer who follows `608-627` lands in real code that is simply the wrong
real code. This is the same shape as the four instances in D-16.R.27: a reference true in the scope where
it was written (`a40c34d`) and applied in a scope where it is not (`HEAD`). **`baseline_commit` is stale
by construction for any story planned before its predecessors landed**, so every later story in this epic
gets its anchors re-verified at its plan gate, not trusted.

### D-16.R.29 — The orchestrator staged a running dispatch's working tree into its own commit

**Reported by the implementation agent, against me, and confirmed.** Commit `0e3b576` —
*"Re-verify 16.2's anchors at its plan gate"* — carries five files. Two are mine (the 16.2 spec, this
log). **Three are Story 16.1a's snapshot-pin records**, which the agent had in its working tree,
unfinished-from-my-point-of-view, when I committed. Its own commit `0bc36b1` was left holding 2 of the
4 files it changed.

**Cause: `git add -A _bmad-output`.** Path-scoped is not scope. `-A` under a path stages *everything
dirty under that path*, including edits another agent is mid-way through writing, and the pipeline's own
instruction — *never `git add -A` your way out of this* — exists for exactly this. I had already
recorded, in this run, that editing a spec underneath a running dispatch is how unstaged work was lost
once; I then did the commit-side version of the same thing.

**Outcome: nothing lost, record misattributed.** The pin text is present in `HEAD` in all three records
(8 occurrences), and the tree is clean. What is wrong is the *history*: 16.1a's pin is attributed to a
16.2 commit. **Not rewriting it** — the branch is local and nothing is missing, the implementation agent
declined to amend across `757a4f0` for the same reason, and a rewrite to fix a message risks real work to
repair a bookkeeping fault. This entry is the correction: **`0e3b576` contains Story 16.1a's snapshot-pin
records, and `0bc36b1` is the other half of that same change.**

**Standing consequence, effective now: stage explicit paths, never `-A`, not even under a path.** A
commit names the files it means. This is the same rule as D-16.R.27's in a different medium — a staging
command written for the scope *"my edits"* and executed in the scope *"everything dirty"*.

**And the near-miss is the part that matters.** The agent's edits happened to be complete when I swept
them, so the commit is merely mislabelled. Had I committed ninety seconds earlier I would have captured a
half-written spec and the story's records would have been internally inconsistent with nothing announcing
it — which is precisely how commit `b8d286d` earlier in this run came to describe edits that never
applied.

### D-16.R.30 — Two records disagreed about 16.1a, and the one claiming more was mine

**Reported by the implementation agent, which correctly refused to decide it.** After the review pass, the
story file read `status: 'in-review'` while `sprint-status.yaml` read `done`. The agent left both alone —
it had been told not to touch sprint-status, and it declined to settle the story's own state by fiat.
That is the right instinct: a disagreement between two records is not resolved by whichever agent notices
it first.

**How it arose.** 16.1a's first build pass reported `done`, and sprint-status was set accordingly. The
story was then **reopened for a review pass**, which landed twelve patches (`5d5409f`). The spec tracked
the reopening; sprint-status did not, because nothing updates it on a reopen — it is written on the way
in and never on the way back.

**Reconciled downward, not upward.** `sprint-status.yaml` is the orchestrator's file and the spec is
build-auto's, so this was mine to fix, and the direction is not arbitrary: **a status record must never
claim more than the artifact it summarises.** `done` beside an `in-review` spec asserts verified gates
that no one has verified since the patches landed. Set to `in-progress`, with the reason in the file
rather than only here. It returns to `done` when the closer verifies the gates **itself**, which is the
one thing the closer exists for.

**The general fault, since this epic is now keeping a list.** A derived record was updated on the
transition it was designed for and not on the transition that reverses it. Same family as D-16.R.27 —
a claim (`done`) true in the scope where it was written (after the first pass) and left standing in a
scope where it is not (after the reopen). **Consequence: a reopened story's sprint-status entry moves
with it, and the closer is the only writer of `done`.**

**Also worth recording from the same pass, because the agent volunteered it against itself:** patch P2
was the same defect as P1, and the agent's words were *"I broke D-16.R.18's rule in the act of citing
it."* Four sites now say four. And P3 was worse than its brief described — `split(' — ')` returns the
**whole string** when the separator is absent, so a separator-free `source` passed a non-empty check
while naming neither project nor path.

### D-16.R.31 — Story 16.1a closes with its follow-up review deliberately OUTSTANDING

**The decision, recorded here because the closer correctly refused to give it a D-number itself.** It
wrote the call into 16.1a's Delivery Log as instructed and then flagged that it had written nothing into
this log — the number is the orchestrator's to assign, and an unnumbered decision is one nobody can cite.

**Decision:** `followup_review_recommended: true` stays **true** on 16.1a. No second review pass was run.
The story is `done`; the obligation is **open**, inherited by whoever resumes Epic 16.

**Reasons:** the flag fired mechanically on *"two or more mediums"*, not on evidence of a defect — the
triage was 0 high / 4 medium / 8 low, all twelve patched with red-proofs, and the nine deferrals are
registered rather than dropped. The owner stopped the run at this story, and a fresh review pass is new
work beyond it. **Reversible by construction:** the flag is left set, so resuming the epic surfaces it.

**What the closer found by not trusting the report — the reason this phase exists.**

1. **The teeth are seven, not four.** Re-proving the batch's tripwire by removing the `lora` row and
   running the *whole module* gives **8 failed / 508 passed across 5 files**: four population floors, the
   variable-upstream list, the offered-once loop, and the delta assertion. The build's "four" was a
   **narrower command scope**, not a smaller guard. Same shape as D-16.R.27 — a count true of the scope it
   was measured in, reported as a property of the guard.
2. **DW-168 overstated its own class, and the correction cuts the other way.** The register claimed
   `-count=1` is *"written into no script, Makefile target or CI config"* and the class *"fully live"*.
   **False for CI:** `.github/workflows/ci.yml` already runs every Go invocation with `-count=1` across all
   four modules, on every push and PR, attributed to D-000.11 and carrying this story's own rediscovered
   sentence verbatim. **CI would have caught the census failure on first push.** DW-168 is narrowed to its
   real residue — the local by-hand path — rather than deleted.

   The closer also reported *how* it nearly missed this: its first grep ended in `head -20` and truncated
   the evidence away. **A search whose output is truncated is a search whose negative result means
   nothing**, and "I grepped and found none" is the most quietly load-bearing claim in this whole epic.

**Outstanding at the stop, stated so nothing reads as finished that is not:** the follow-up review pass on
`4aca77f..99ac74c`; the heavy-test catch-up (matrix corpora, four AD-21 legs, `TestCrossTargetByteIdentity`,
browser specs) due before Epic 16 closes; DW-168's narrowed discharge; DW-162, whose margin is now **halved
to 10** and watched by nothing; and DW-166's tripwire, named but not built.

### D-16.R.32 — The five outstanding obligations get a disposition each, and only one of them is a story

**Orchestrator decision**, on the engineering lead's recommendation, taken on resuming the run on
2026-09-03. D-16.R.31 closed 16.1a with five obligations named and none of them scheduled. Naming a debt
and scheduling it are different acts, and the gap between them is where obligations go to die.

**Verdict — one line each, and each has a named home.**

| Obligation | Disposition | Home |
|---|---|---|
| 16.1a follow-up review (`4aca77f..99ac74c`) | **Before 16.2 is dispatched**, as its own read-only pass | its own pass, not folded into a story |
| Heavy-test catch-up | **Split.** Browser half in 16.3; Go half (matrix corpora, four AD-21 legs, `TestCrossTargetByteIdentity`) at end of run | 16.3 + end-of-run |
| DW-168 (`-count=1` on local Go runs) | **Discharged during**, as a Verification edit to all three remaining specs | 16.2's plan gate, applied to 16.2/16.3/16.4 at once |
| DW-162 (margin halved to 10 of 64) | **Assert-only.** Mechanism stays deferred | 16.2's Verification |
| DW-166 (batch tripwire) | **Stays standing.** No completion state; the lead holds it | none — not a story |

**Situation.** Each of these was registered honestly and left unowned. The temptation on resuming is to
fund them all as work, which would put roughly a story's worth of effort in front of 16.2 and price two
*permanent* obligations (DW-162's margin watch, DW-166's tripwire) as one-off tasks — the inverse of the
criterion error D-16.R.16 already corrected once in this epic.

**In simple terms.** Five items on a list are not five jobs. One is a job you should do now because it
gets more expensive every day you wait. Two are things you should simply *write down in the right place*
so a number gets checked automatically. One is a standing watch with no finish line, which belongs to a
person, not a sprint. And one is a job whose right moment is later, when a browser is already open for
another reason and the marginal cost is near zero.

**Why the follow-up review goes first, and this is the only one that changes the dispatch order.** Its
subject — the fetch path, the provenance record, the licence verdicts — is the code that 16.2, 16.3 and
16.4 all build on top of. Reviewed now, a finding lands on a two-commit-old diff. Reviewed after three
more stories, the same finding has to be separated from three stories of change layered over it: a scoped
review becomes archaeology. The expected yield is low (the flag fired mechanically on *"two or more
mediums"*, over twelve findings already patched with red-proofs) but the expected cost is an hour, and
the asymmetry is entirely one-sided.

**Why it is NOT folded into 16.2.** Putting a 16.1a review finding inside a 16.2 dispatch makes one
commit answer for two stories, which is the shape D-16.R.11 spent this run's own money learning to
refuse. It runs as its own pass, read-only, producing findings for triage rather than patches.

**Why the browser half of the catch-up lands in 16.3 specifically.** 16.3 already carries a browser run
under D-16.R.1's own override — its specimens, focus trap and staged-confirm flow have no non-browser
witness — so the container is already being paid for. Scoping that run to also discharge **DW-161**'s
three cases (a pick with the network up, with it down, and with a licence outside the allowlist, against
the real hosts) costs almost nothing and converts the largest evidence gap in the epic from a note into a
measurement. This is not a cadence exception; it is the override D-16.R.1 already granted, used well.

**The accepted cost.** The Go half of the catch-up still runs once, at the end, across three stories'
worth of change — so a red there is still expensive to bisect. That is the `end-of-run` bargain the owner
struck in D-16.R.1 and this decision does not reopen it. What makes it tolerable is that 16.2, 16.3 and
16.4 touch **no Go code**, so the Go suites have no story-attributable reason to move; if they go red, the
cause is upstream of this epic, not inside it.

**Consequences.** 16.2 is not dispatched until the follow-up review returns and its findings are triaged.
All three remaining specs get `-count=1` written into their Go Verification commands at 16.2's gate.
16.2's Verification gains a reported measurement of `s1.assetCount` after build, so a store implementation
that accidentally lands anything in the release manifest is caught by a number rather than by a release
failing. DW-166 is not built and not closed; the lead carries it, and a regeneration of `font-index.json`
by any of the three remaining stories is trigger 1 and returns it to the lead at that moment.

**How we'd know this was wrong.** The follow-up review returning a high-severity finding in shipped 16.1
or 16.1a code — that would mean "expected yield is low" was a guess dressed as an estimate, and that the
same reasoning should not be used to defer the next such flag.

### D-16.R.33 — Four gate rulings for the remaining stories, and the one place the lead's own method gave the opposite answer

**Lead rulings**, requested by the orchestrator before 16.2's dispatch and applied. All four were RULINGs;
none ran out of direction. Recorded together because they were taken as one set against one warm
measurement of the tree at `8a9e297`.

**R1 — 16.2 owns the store and the read-before-fetch; 16.4 owns the group in the control.** 16.2 needs no
new dropdown group at all: its "picked again / no network" criteria are satisfied through **today's**
control by putting a store read in front of the existing fetch. Only AC 4 reached into the control, and it
is restated against the store's listing plus a red-provable no-host-fonts assertion. **The seam is the
ruling's real content:** the store's listing joins the shipped `FamilySource` union as a third `'stored'`
arm, so 16.4 adds headings rather than plumbing, and every exhaustive switch over the union reds until
16.4 handles the arm. That is the mechanism enforcing the hand-off instead of a sentence in a spec.

**R2 — no keyboard shortcut in this epic, and no hint glyph.** `Add fonts…` is 16.3's entry point (a story
cannot verify a modal it cannot open); the footer row is 16.4's. `⌘G` is declined: it is the browser's
Find Next, and this application's own convention puts conventional document actions on Command and
**app-specific actions on Option** (`⌥P`, `⌥S` — verified in `src/shortcuts.ts`, an 8-key closed record).
The glyph is not rendered because **a `⌘G` label beside a key that does nothing is a false UI string** —
the class this epic has already ruled against three times (D-16.3 on "live", D-16.R.2 on the family count,
D-16.R.20 on the header). Both or neither, and so neither. Deferred with `⌥F` named as its shape.
16.3's spec pre-authorised exactly this: *"or the omission is ruled and recorded"*.

**R3 — both of 16.3's data-less arms are dropped, and the decline is PRICED rather than asserted.**
*Most styles* sorts on a property the product collapses: `font-source.ts:197` selects only
`style === 'normal' && weight === 400` and `:314` refuses a family publishing no upright Regular, so
**exactly one face is embedded per family** — eighteen styles and one style deliver the author the
identical thing. `designer` is not a field addition but a snapshot regeneration, which breaks the
`d6d51f1` pin, fires DW-166 trigger 1, and per D-16.R.26 inverts 16.1a's landed batch on the axis this
epic litigated twice. **The lead measured the payload first, on purpose:** `+1,326` brotli bytes
(~0.008% of a 15,729,262-byte first load) — **affordable, and declined on the criterion anyway.** It also
established that DW-162's halved margin does **not** bind here: DW-162 counts cache slots and the
snapshot is source bundled into JS. Recording the affordability is the point — otherwise the next reader
reverses this the moment 15.0 frees room, believing it was a budget call.

**R4 — the fetch timeout's BASIS, not its value.** Sited in the injected `Fetcher` default so one place
covers all six round-trips. **The signal must reach `fetch()` so the abort propagates to the body
stream** — the bytes are read by `response.arrayBuffer()` *after* the fetcher returns, so a header-only
timeout leaves the worst real stall completely uncovered, and the test must stall during `arrayBuffer()`
rather than during the request. The number comes from timed fetches against the real host over a **large**
face (shipped faces reach 646 KB; a budget fitted to the 90 KB measurement refuses the tail it exists to
tolerate), taking the **maximum** not the median — it is a ceiling on patience, not an estimate of typical
latency — recorded per D-8.4j.8, and **explicitly not measured in jsdom against our own stub**, which
would be D-16.R.12's vacuity. Worst case is 6 x T with the control disabled and must be stated. Discharges
DW-165.

**THE CORRECTION, and it is the entry's most useful part.** R1 established that 16.2's AC 4 sits at line
158 while `</intent-contract>` closes at 105 — **outside the lock**, so a Tasks/AC edit, not an amendment.
The orchestrator applied that same method to 16.3 and 16.4 and it returned the **opposite** answer:
16.3's contract spans **27-91**, and R3's Sort row (`:82`) and the three `~1,946` statements (`:31`,
`:50`, `:52`) are all **inside** it — on a spec already at `ready-for-dev`, i.e. approved at a previous
session's CHECKPOINT 1. **So R3's edits are contract amendments and R1's were not.** The lead did not
claim otherwise; it located the lock for the spec it was asked about and did not generalise. The lesson is
that "locate the lock" is a **per-spec** check, and a ruling that was correctly not-an-amendment in one
spec is not evidence about another. Verified after editing by the honest method the lead prescribed —
every diff hunk at a line number beyond the contract's close (16.2: all six at >= 144; 16.3/16.4: all at
>= 157) — rather than by md5, because the lead could not reproduce D-16.R.28's recorded digest and
correctly refused to quote either figure forward.

**Consequence: 16.3's and 16.4's ruled changes are RECORDED but NOT APPLIED.** Both specs gain a
non-normative `## PENDING GATE AMENDMENTS` section listing every ruled change with its reasoning, plus the
anchors known stale at `8a9e297`, explicitly labelled do-not-trust. They are applied at each story's own
plan gate. **Reason: Story 16.2 lands before either of them, so anchors re-verified today would be rotted
again by the time they are read** — work done twice and trusted once. The contract amendments are then
reopened deliberately, at a gate, and recorded in a `## Spec Change Log`, rather than edited in silently
three stories early.

**How we'd know R2 was wrong.** The owner reading the mockup's `⌘G` as a requirement rather than a hint
label. The lead put its own confidence at medium on this one alone, and the reversal is one line — a ninth
`ShortcutHints` entry plus one arm in the existing handler. It goes to the owner at 16.3's gate rather
than being presented as settled taste.

### D-16.R.34 — The follow-up review was not clean, and the estimate that deferred it was a guess

**Measured**, by the read-only follow-up pass D-16.R.32 scheduled on `4aca77f..99ac74c`.

**D-16.R.32's stated falsifier has partially fired, and this entry is that admission.** That decision
deferred nothing but justified its ordering with *"the expected yield is low"* — the flag having fired
mechanically on "two or more mediums" over twelve already-patched findings. Seven findings came back, one
of them medium-high and sitting in a file Story 16.2 will touch. **The reasoning was sound and the
estimate inside it was a guess dressed as an estimate**, which is the same species as every other error
this epic has catalogued: a claim asserted at a confidence its evidence did not carry. The *decision* to
run it first was right; the *prediction* attached to it should not have been stated so flatly.

**F1 — the tripwire's whole prohibition half is asserted by nothing (medium-high, MEASURED).**
`folio-designer/src/test/provenance-shape.ts`. Replacing `hostShaped`, `schemeShaped`, `digestShaped` and
`branchShaped` with `/(?!)/` and `projectShaped` with `/[\s\S]*/` — five of seven assertions rendered
inert — and running the helper's **complete** call scope (exactly two importers) gives **33 passed, 0
failed.** Every call site feeds known-good values; nothing anywhere feeds the predicate a violating
string. **Patch P6's widening (`develop`, `trunk`, `Main` "cannot pass") could not have had a red-proof,
because no input reaches those branches.** This is the "test that passes whether or not the feature works"
shape one level up: the *guard* is untested, not the feature.

**In simple terms.** The story built a tripwire and proved it by walking around it, never over it. Every
test steps carefully on the safe side of the wire and reports that nothing went off. Cutting the wire
changes nothing anyone can observe.

**F2 (medium, measured).** Two false positives in those same regexes, in the direction no deferral covers:
`hostShaped` matches `notofonts/notofonts.github.io@v2.0` because `github.io` hits the `io` TLD — a
plausible next-batch upstream that would red the committed-tier loop for the whole catalogue — and
`branchShaped` trips on any path segment in `dev|head|latest|default|trunk|main|master|develop`, e.g.
`ofl/dev/Dev-Regular.ttf`. Deferral #2 registers the **under**-inclusive direction only.

**F3 (medium, measured).** `lint/internal/licence/licencecensus_test.go:23-24` calls `pinnedCensus` the
whole population at "**48 committed** files plus 9 dependency licences". Actual: 67 rows, 9 `dep`,
**58 committed** — and the story's own Delivery Log says 58 + 9 = 67. The commit added ten rows and an
"AND BY 10 AT STORY 16.1a" paragraph **three lines below** the stale total and left the total at 48. The
vacuity floor at `:229` is still `< 20` against 58, its message still naming "the 12 asset files".
**Story 16.1a raised four population floors in `folio-designer` for precisely this reason and walked past
the fifth one, in the module it was editing.**

**F4-F7 (medium-low to low).** `build-wasm.mjs` corrected "44 -> 54" at `:272` and left "21 texts" at
`:269` and "all 21 committed faces" at `:236`, with `:238-240` describing SPDX-identifier keying that
`:257` says was **replaced** — the paragraph contradicts the one below it. Eight stale population claims
in `folio-go`, including "the OFL row covers 19 of the 21 catalogue faces" (actual 28 of 31) and a
justification for having no build-time Apache row that reads *"no committed catalogue face is
Apache-licensed"* — `robotoslab` now is, though the behaviour was verified correct and only the
justification is false. `assertProvenanceShape`'s comment promises a path assertion the code never makes.
The "one writer per tier" halves are asymmetric: fetched exhaustive, committed a bare `toContain`.

**What the pass confirmed, stated so it is falsifiable.** The "four population floors" claim is exact at
repo scope. P3's `split(' — ')` fix is complete — repo-wide, `provenance-shape.ts` is the only thing that
splits on that separator, and the engine treats `source` as an opaque string it never parses.
`provenance-shape.ts` weakens nothing it centralised; it is strictly stronger than the inline original.
Three of the twelve red-proofs were independently re-run and all three genuinely red.

**One attribution corrected before it entered the record.** The reviewer saw `epic-16-decision-log.md`
dirty in its baseline and attributed it to D-16.R.31. It was the orchestrator's — the grounding refresh
and D-16.R.32, written while the review ran. Recorded because an unexamined "pre-existing" label is
exactly the claim this pipeline is supposed to verify rather than accept.

### D-16.R.35 — Story 16.1a is reopened for a bounded remediation pass, and stale comments get a rule

**Lead ruling**, requested by the orchestrator after D-16.R.34's findings and applied in full. Six findings
in, one registered, nothing folded into 16.2.

**Verdict.** Reopen 16.1a for **one** remediation pass over F1, F2, F3, F4, F6 and F7, bounded to four
files. Register F5. No new story. `status: in-review`, sprint-status `in-progress` — moved **with** the
reopen per D-16.R.30, and returning to `done` only when the closer verifies the gates itself.

**The discriminator, and it settled all seven at once: causation.** For each finding the lead asked *did
16.1a's own change make this false or leave it unproven?* — and every one answers yes. It wrote the helper
D-16.R.13 demanded (F1, F6, F7), wrote those regexes (F2), took the committed licence population from 48
to 58 (F3), corrected 44 to 54 and left the siblings (F4), and added the first Apache catalogue face (F5).
**16.1a is the toucher, so this is not going looking.** D-16.R.11 cuts *for* the reopen rather than
against it: what that ruling forbids is dragging a story's findings into a **later** story's dispatch,
which is precisely what is being refused for 16.2.

**Why F1 is not discretionary.** D-16.R.13's guardrail said it in terms: *"A test asserts `source` contains
no scheme and no host on either tier — **that is the tripwire, and the convention alone will not hold**."*
A tripwire that reports 33 passed / 0 failed with five of its seven assertions replaced by never-match is
not the tripwire that ruling required. **An unmet ruling guardrail is the strongest category the lead
has.** It is also the second time this one file has looked asserted while asserting nothing — D-16.R.30
recorded 16.1a's own P3 as *"worse than its brief described"*.

**F2 lands BEFORE F1, and the ordering is load-bearing rather than stylistic.** If the negative-case table
is written first, it encodes `google/fonts — ofl/dev/Dev-Regular.ttf` as *must throw* and **pins the false
positive as intended behaviour**. The next batch carrying a `github.io` upstream then reds the
committed-tier loop for the whole catalogue, and the cheapest-looking repair at that point is to weaken
the guard. Narrow first; write the table against the narrowed predicates.

**The anti-softening control, because "narrowing a guard" is exactly what a regression looks like.**
Narrowing to the correct operand is a correctness fix: a repository whose *name* contains a TLD is not a
host reference, and a directory literally named `dev` is not a moving ref. What makes that checkable
rather than assertable is that the original defect must still trip both prohibitions afterwards —
`raw.githubusercontent.com/google/fonts/main/ofl/kanit/Kanit-Regular.ttf`, confirmed today to trip both.
The property is pinned and the mechanism is left to the implementer: **a host or ref prohibition must be
evaluated where a host or ref would actually appear, not as a substring scan over the whole field.**

**THE RULE ON STALE COMMENTS, which is the entry's reusable half.** This epic has hit stale-comment
findings four times and been deciding each by hand.

> **A stale comment is a FIX when acting on it fails SILENTLY, and a DEFERRAL when acting on it fails
> LOUDLY.** Ask: if the next engineer believes this sentence and acts on it, does a test go red, or does
> the system quietly accept the wrong thing?
>
> **Override: a comment that contradicts another comment in the same file is always a fix**, whichever
> way it fails. Once two paragraphs disagree about the mechanism, a reader cannot tell which half is
> current, so **neither half is evidence any more.** The fault is not the wrong sentence; it is that the
> file's whole comment layer has stopped being usable.

**And the rule's first draft was wrong, which is worth more than the rule.** The lead's initial version was
*"stale reason vs stale fact"* — and **F5 refuted it within the hour.** F5's sentence is grammatically a
justification (*"it never needed one — no committed catalogue face is Apache-licensed"*), so the draft rule
would have called it a fix. But the lead verified rather than classified: both the runtime and the
build-time Apache rows **exist and are correct**, and the strict `.toBeDefined()` assertion means the worst
action the false sentence licenses — deleting the TS Apache row — reds immediately. **Grammar was the wrong
axis; failure mode is the right one.** A rule that had been adopted from its own plausibility would have
sent a loud-failing comment to the fix pile and, worse, would have been carried forward as settled.

**F3's floor is DERIVED, not raised.** `if len(committed) < 20` becomes a floor computed from
`pinnedCensus`'s own non-`dep` row count. **A literal is the thing that just rotted** — 20 was written
against a population that has since nearly tripled, so it passes while the walk finds a third of the tree,
and replacing it with 58 re-arms the identical failure on a longer fuse. The two sides are independently
produced (one walks the tree, one is typed by a person), so this is not circular, and the floor becomes
self-maintaining — which is what D-16.R.12's guardrail asked for when it said *"a floor left at 21 while
the tier grows to 30 is a floor that stops measuring the thing it was built to measure."* Its job is
**vacuity detection**, and the message must say so: fail early and legibly when the walk itself breaks,
rather than emit 58 per-row errors.

**DW-168 collides with this on purpose.** `TestLicenceSignalCensus` is that entry's **worked instance** —
the filesystem-walking test served a cached PASS while ten new licence texts sat unrecorded. So the pass
must run `cd lint && go test -count=1 ./...` and **record the invocation, not merely the verdict**; by
DW-168's own text a green taken without `-count=1` on this test is not evidence. This discharges the
entry's narrowed residue on the path where it actually bit.

**What 16.2 owes, and what it does not.** F1 stays 16.1a's: **16.2 becomes a third *carrier* of `source`,
not a third *writer*** — the store persists a string the fetched or committed tier minted, and asserting
the provenance shape of a value you copied verbatim proves nothing about the copier. 16.2 gains one line
instead: a round-trip assertion at the **retrieval** side, which makes its existing contract Boundary
(*"the store never becomes an authority on a document"*) checkable rather than adding scope — and gives
the shared helper a **third real call site**, the absence of which is what let F1 hide in the first place.
The coupling is a reason to order the pass first, not to merge them: if the pass slips, 16.2's line still
lands and is simply weaker until F1 arrives. It never blocks.

**Scope, stated so nobody re-derives it.** Four files, no product behaviour change — every fix is a test
predicate or a comment — so the before-the-tag set is untouched and **D-000.15 is not engaged**. The 23
goldens are not consulted because nothing rendered changes.

**A negative result, recorded because it is worth as much as a finding.** The lead went looking for an
eighth defect: whether the build-time Apache row 16.1a added at `font-catalogue.test.ts:203` is vulnerable
to D-16.R.12's standing rule 1 (a test that admits because nothing matched). It is **not** — that table
asserts `.toMatch(signature)` against real bytes and `robotoslab`'s nameID 13 genuinely reads the Apache
sentence, so "admits because it matched" and "admits because nothing matched" are distinguishable and a
broken regex reds. There is no eighth finding.

**How we'd know this was wrong.** The remediation pass finding that a fix cannot be red-proved — that would
mean the defect is in the shape of the helper rather than in its predicates, and the repair is a redesign
rather than a table.

### D-16.R.36 — ADDENDUM to D-16.R.35's stale-comment rule: a correct fact can launder a wrong one

**Reported by an implementation subagent during the remediation pass, unprompted, as an aside to a
question it had been asked about something else.** It is a real refinement to the rule recorded one entry
earlier, so it gets a number rather than living in a Delivery Log.

**The rule as recorded** sorts a stale comment by whether acting on it fails loudly or silently. The
subagent was asked which side a particular sentence fell on — one naming both a filename and a count — and
answered: **both, in the worst possible mix.**

**The mechanism.** The stale *filename* fails loudly: grep it, find nothing, look again. The stale *count*
fails silently, because a number is not greppable — a reader simply carries it away. What makes the pair
worse than either alone is that they **share one sentence**, and one of the two filenames it names is
real. So a reader who spot-checks the sentence finds one name valid and one missing, corrects the missing
one, and treats the number sitting beside the **verified** name as checked-by-association.

**In simple terms.** A sentence says *"see `manifest.go` and `licencegraph.go` — there are 12 of them."*
You check the first name: it exists. You check the second: it doesn't, so you fix it. You never check the
12, because the sentence has just demonstrated that it can be verified and you did verify it. The half
that survived your check vouches for the half you never checked. **The correct filename launders the wrong
number.**

**Consequence, added to D-16.R.35's rule rather than replacing it:** *when one sentence carries both a
loud-failing claim and a silent-failing one, it is a **fix**, and the loud half is the reason — not the
mitigation.* A partially-verifiable sentence is more dangerous than a wholly stale one, because a wholly
stale one fails every spot-check it meets. This is the same family as everything else this epic has
catalogued, in yet another medium: **confidence earned in one scope, spent in another.**

**How we'd know it was wrong.** If the mixed sentences we leave standing under this rule turn out never to
mislead anyone — i.e. readers do re-check the numeric half — then the laundering is theoretical and the
rule can go back to sorting by failure mode alone.

### D-16.R.37 — The remediation pass's two checkpoints: a guard whose reach a formatter decides, and a rule confirmed by independent convergence

**Orchestrator decisions**, both answered without the lead because both were settled by rulings already on
this log. Recorded because one of them extends a fix beyond its literal brief and that should never be
silent.

**Q1 — F7's fetched side is exhaustive only by FORMATTING. Fixed, not deferred.**

The dispatched fix was *"make the committed side exhaustive"*. Done — and the independent red-prover then
attacked the **fetched** side of the same test and found it scrapes `font-source.ts` with a **line-anchored,
comma-terminated** pattern. Measured, with two spellings of the same second writer: `source:` starting its
own line **reds**; the identical code reflowed onto one line **stays green, 46 passed / 0 failed**. The
evading writer emitted the exact original defect — a bare `main` branch URL.

**In simple terms.** The alarm is wired to the door frame rather than the door. Come through the same
doorway with your shoulder turned and it never sounds. Nothing about the intrusion changed; only its
formatting did.

**Why this is completing F7 rather than opening F8, which is the whole question the builder raised against
itself.** The builder argued fairly that a literal reading of its brief puts this out of scope, and asked
rather than deciding — correctly. But the ruling's own words were that the two halves *"are **asymmetric**"*
and its reason was that the helper exists *"because the defect this tripwire guards is the two tiers
drifting apart."* **An assertion permissive on one tier and exhaustive on the other IS that drift, inside
the guard** — so swapping which tier is permissive relocates the defect instead of discharging it. "Make the
committed side exhaustive" named the measured instance, not the boundary of the fix.

**Three things settled it.** (1) It is a **false negative on the property the test's own name claims**, and
the frontmatter deferral that already exists covers the **opposite** direction — it registered the regex
going *red* on a reflow; this is the reflow making the guard *silently admit*. A deferral pointing one way
is not coverage for the other. (2) It is **F1's shape recurring inside F1's own remediation**: F1 was a
guard that looked asserted and asserted nothing, and a guard whose reach a formatter decides is the same
defect in a different mechanism — closing the pass that fixed an inert tripwire while shipping an evadable
one. (3) **D-16.R.35's override applies literally**: the file's newly written comment condemns what would
have been left standing, so two paragraphs in one file would disagree about whether the guard is symmetric,
and neither would be evidence any more.

**Red-proof required in all three spellings**, not the first: the second writer reds on its own line, reds
reflowed onto one line, and a `source:` inside a `//` comment stays green as the stated exclusion requires.

**Q2 — the third stale claim in `licencecensus_test.go`. Fixed, counts DROPPED rather than retyped.**

*"ClassifyLicenceText is shared between the ASSET path (`manifest.go`, 12 files) and the DEPENDENCY path
(`licencegraph.go`, 9 files)."* Measured: `manifest.go` exists, `licencegraph.go` does not (the package has
`graph.go`), the `9` is right and the `12` is the same rotted figure F3 was told to stop repeating.

**This is D-16.R.36 arriving a second time, from a different agent, in the same sentence — and that
convergence is the entry's real content.** D-16.R.36 was recorded roughly ninety minutes earlier from an
unrelated subagent's unprompted aside. The remediation pass's red-prover reached *"the correct name
launders the wrong number"* independently, without having seen it. **Two independent readers converging on
the same mechanism is the strongest evidence available that a rule is real rather than a plausible
story** — which matters because D-16.R.35's first draft was plausible and wrong. The rule is now applied
rather than proposed: a sentence carrying both a loud-failing and a silent-failing claim is a **fix**, and
the loud half is the reason, not the mitigation.

**The counts are dropped, not refreshed.** Retyping is how this rotted twice, and `~43` describes a
population **no longer defined in that package**, so a fresh number there would be born stale. Where scale
matters, name where the population is defined instead of how big it is — the same principle as F3's derived
floor: **stop writing down numbers that nothing maintains.**

**Three disposals, recorded so they are not re-derived.**

- **`build-wasm.mjs` ~109, "six slots" of an object holding nine — NOT fixed, registered as a deferral
  stating the AMBIGUITY as the finding.** It is defensible as meaning the six font slots and nothing
  mechanical depends on either reading. **An ambiguous comment is not a stale one and D-16.R.35's rule does
  not cover it**; picking the likelier reading would be inventing an answer, which is the one thing this
  pipeline forbids everywhere else. It needs whoever next edits that object to say which it intends.
- **The fifth stale site swept at ~118** ("Twenty-seven hardcoded keys" → no count at all) was not on the
  F4 list. Kept, and **said so in the Implementation Notes** — a sweep that silently exceeds its brief is
  indistinguishable from scope creep unless it declares itself.
- **S6's three bounds on the derived floor** (catches under-collection only; counts entries not distinct
  files; the `dep ` discriminator is sound by observation of today's paths rather than by construction) are
  **stated in comments rather than guarded**. That is the correct application of this file's own subject: a
  comment claiming a safety it does not have is exactly what was being fixed. No guards built for them here.

**And the pass broke a guard with its own remediation, which is worth recording as a success of the gates
rather than a defect.** The first drafts spelled real font hosts in test values and comments;
`npm run scan:font-hosts` went red with **9 occurrences**, two of them `fonts.gstatic.com`, forbidden
outright by D-16.3. The repair composes the host rows from the exported `fontsRepositoryHost` constant
rather than spelling them — **strictly better than a literal because it is rename-proof** — and refuses the
obvious alternative of substituting a fictional host, on the grounds that *the tripwire is stated against
the declared hosts by identity, and a fictional host would make it prove less.* Same reasoning as F2's
operand narrowing: repair at the right level, never retreat to something weaker that looks equivalent. The
original-defect row's trip set was re-confirmed unchanged at `['host','ref','date','separator','project',
'path']`.

**How we'd know Q1 was wrong.** The unanchored scrape producing a false **positive** — reddening on a
legitimate `source:` occurrence the anchors were excluding on purpose. The stated exclusion (a `source:`
inside a `//` comment stays green) is the specific case that would show it, which is why it is red-proved
rather than assumed.

### D-16.R.38 — CORRECTION to D-16.R.36 and D-16.R.37: the filename was never wrong, and "two independent readers" were one method failing twice

**Orchestrator error, caught by a subagent that refused the instruction rather than executing it.** The
originals stay above, unedited, per this log's standing practice — never rewrite history, append the
reversal.

**What I ruled, and what was true.** In D-16.R.37 (Q2) I instructed: *correct `licencegraph.go` →
`graph.go`, and drop both counts.* The implementation subagent did the counts half, **refused the rename**,
and returned the measurement. Verified independently by me before accepting it:

- `lint/internal/rules/licencegraph.go` **exists**, is git-tracked since `5dddbea`, and line 44 calls
  `licence.ClassifyLicenceText` — it **is** the dependency path.
- `lint/internal/licence/graph.go` holds `ResolveGraph` and `ReadLicenceText` and **never touches the
  classifier**.
- The two non-test callers of `ClassifyLicenceText` are exactly `internal/manifest/manifest.go` (asset
  path) and `internal/rules/licencegraph.go` (dependency path).

**Both filenames in the sentence were correct, and precisely chosen — one per path.** My rename would have
put a false statement into the file whose named subject (D-8.0.1) is *comments claiming things that are not
so*. That is the worst available outcome for this particular fix, and the pass would have shipped it on my
authority.

**How the premise failed, and this is the part with teeth.** Grepping for `licencegraph` **inside**
`internal/licence` returns nothing, because the caller is in a **sibling package**. The follow-up reviewer
did that search, the remediation builder's red-prover did that search, and I accepted the result without
running a wider one. In D-16.R.37 I then wrote that *"two independent readers converging on the same
mechanism is the strongest evidence available that a rule is real."*

**They were not independent.** The subagent's words, which are the correction:

> *Two agents reaching the same wrong answer via the same too-narrow search isn't corroboration — it's one
> method failing twice.*

**Convergence is only evidence when the methods differ.** Two readers running the same scoped grep are one
measurement reported twice, and counting it as two is how a shared blind spot acquires the authority of
replication. I made exactly the error this log has catalogued a dozen times, in its most embarrassing
medium: **a claim true in the scope where it was produced** — `licencegraph.go` is absent *from
`internal/licence`* — **applied in a scope where it is false** — absent *from the repository*. And I made it
while writing the entry that named that pattern.

**The mechanism survives and is STRENGTHENED, which is why D-16.R.36 is corrected rather than withdrawn.**
D-16.R.36 said a correct filename launders a wrong number beside it. With **both** filenames correct, the
two rotted counts were laundered by **two** verifiable names rather than one. The rule stands; only the
worked example's details were wrong. What must be struck from D-16.R.37 is the *independent-convergence*
argument I used to promote the rule from proposed to applied — the rule is well-founded on its mechanism,
not on a replication that never happened.

**The real defect, which survives the bad ruling.** The paths were **unqualified**. A bare `manifest.go` /
`licencegraph.go`, read from inside `internal/licence`, is unfindable — which is precisely why it read as
rot to three separate readers. The fix applied: both names written as **full repository paths** so the next
reader's grep succeeds; **both counts dropped** (the correct `9` as well as the rotted `12`, since a bare
number beside a now-findable path is still a number nothing maintains); one line recording that an earlier
draft named them bare, **so the next agent who greps locally and finds nothing reads why before concluding
rot**. That last line is the durable part: it inoculates the next reader against the specific error three
readers just made.

**Consequences, standing from here.**
1. **A "file does not exist" claim is repo-scoped or it is not a claim.** Verify absence with a
   repo-rooted search (`git ls-files`, or a `find` from the root) before acting on it. A package-scoped
   grep establishes absence *from that package* and nothing more.
2. **Convergence counts only across differing methods.** Two agents, one method, is one datum. When citing
   agreement as evidence, state the methods; if they are the same, say so and discount it.
3. **Comment references to code in another package carry their full path.** A bare filename is only
   findable from the directory that holds it, and a reference that cannot be resolved from where it is read
   will eventually be *corrected* into a falsehood by someone acting in good faith — which is exactly what
   nearly happened here.

**The part worth more than the correction.** The subagent was given a direct instruction by the
orchestrator, through the builder, and **declined it on measurement** rather than complying and letting a
false statement land under someone else's authority. Then it reported the refusal with its evidence and
flagged that an entry in this log needed amending. **That is the behaviour this pipeline's every checkpoint
exists to produce**, and it worked at the lowest level of the stack, against the highest. Recorded so it is
visibly rewarded rather than merely absorbed.

**How we'd know THIS is wrong.** `internal/rules/licencegraph.go` ceasing to call `ClassifyLicenceText` —
at which point the sentence names a file that no longer belongs to the dependency path, and the right fix is
to re-derive both callers rather than edit one name.

### D-16.R.39 — The collision of D-16.R.29 ran in reverse, and "stage explicit paths" turns out not to be the whole rule

**Observed, not decided.** Commit `efa9b75` carries eight files: the four in the remediation pass's scope,
plus the story spec, `deferred-work.md`, `sprint-status.yaml` and **299 lines of this decision log** —
D-16.R.35 through D-16.R.38, all four written by the orchestrator while the pass was running.

**This is D-16.R.29 exactly, with the parties swapped.** There, *I* swept the implementation agent's
in-flight working tree into my commit with `git add -A` under a path. Here the builder swept **my** in-flight
writing into its commit. It staged **explicitly**, named all eight files, declined to leave a 99-line ruling
out of history, and said in the commit message which parts are mine — so it did everything D-16.R.29's
standing consequence asks for, and the collision happened anyway.

**So the rule was incomplete.** D-16.R.29 concluded *"stage explicit paths, never `-A`, not even under a
path."* That prevents sweeping files you never looked at. It does not prevent staging a file you looked at,
that is genuinely dirty, that you did not write. **The missing half: name the paths you OWN, not the paths
that are dirty.** In a tree with concurrent writers, "explicit" and "mine" are different sets, and only the
second one is safe.

**Outcome is benign, again, and for the same reason as last time: nothing was lost, only misattributed.**
The tree is clean, every line is in `HEAD`, and this log's own text records who wrote what. Not rewriting
history — the branch is local, nothing is missing, and a rewrite to repair a bookkeeping fault risks real
work. **`efa9b75` contains four orchestrator decision-log entries that are not part of Story 16.1a's
remediation.**

**The near-miss is the same one, too.** D-16.R.29 noted that had the commit landed ninety seconds earlier it
would have captured a half-written spec. Here, D-16.R.38 — the entry correcting the builder's own false
premise — was finished minutes before the commit. Had it not been, the commit would have preserved
D-16.R.36 and D-16.R.37 stating a falsehood, with the correction landing in a later commit that reads as a
reversal rather than as what it was: an error caught before it shipped.

**Standing consequence, amending D-16.R.29's:** *while a dispatch is running, the orchestrator does not
commit and the builder stages only files it wrote.* If the orchestrator's records must be committed, that
happens **after** the dispatch reports, in its own commit. The cheap version of this rule, which would have
prevented both instances: **whoever commits, commits only what they touched, and checks `git status` for
files that appeared without them.**

### D-16.R.40 — What this pass demonstrated about the author/prover separation, recorded as evidence rather than as principle

**The remediation pass introduced TWO regressions and caught BOTH before commit — each by someone other
than whoever introduced it.**

1. **Nine forbidden-host occurrences**, two of them a host forbidden outright by D-16.3, introduced by the
   fix-writing agents spelling real hosts in test values and prose. Caught by the gate, repaired by
   composing from the exported constant.
2. **The `licencegraph.go` false premise**, introduced by the reviewer, propagated by the builder,
   **ratified by me in a ruling**, and refused at the last possible moment by the implementing agent — the
   lowest level of the stack declining a direct instruction from the highest, on evidence.

**Neither was caught by its author.** In both cases the person who introduced the error had already checked
their own work and been satisfied. The F1 finding that started this whole pass has the same shape one level
up: an assertion that its author believed was proving something, which proved nothing, and which no amount
of that author re-reading it would have revealed — it took a different agent deleting the assertions one at
a time to establish that not one of them could fail.

**This is why the separation is mandatory rather than advisory**, and it is worth stating as measured
evidence because the cost is real: every red-proof in this pass cost a second agent, and the pass ran to a
hundred tool calls. The return on that spend, this pass, was two regressions caught and one inert guard
armed. **A pass that had trusted its authors would have shipped all three.**

**The generalisation, since this epic is keeping a list.** Every error catalogued in this run — the
37-vs-38, the live-index eleven, "only three families moved", "four floors" that were three files, the gate
against a superseded table, the narrower-scope tripwire count, and now a package-scoped grep read as a
repository-scoped absence — was **produced by someone with every reason to believe they had checked.** The
discipline that catches them is not more care by the author. It is a second party with a different method,
and the second half of that sentence is the load-bearing one: **a second party using the SAME method is not
a second party** (D-16.R.38).

### D-16.R.41 — A ruled deferral vanished and the COUNT reconciled, because a different finding filled its slot

**Found by the closer**, by checking a ruling's disposition against the artifact rather than against the
build's own tally. Story 16.1a's remediation pass is closed `done` at `efa9b75` + `5a3745f`; this is the
one thing that did not survive it intact.

**What happened.** D-16.R.35 ruled seven findings: six FIX, and **F5 REGISTER** — the eight stale
population claims in `folio-go`, with a named trigger and owner. The pass reported `defer: 1`. That one
deferral is **DW-169**, the `build-wasm.mjs` "six slots" ambiguity — a finding the pass discovered *itself*,
during its F4 sweep, and which I disposed of separately in D-16.R.37. **F5 appears nowhere**: not in
`deferred-work.md`, not in the spec's frontmatter `deferred:` list, under no number.

**In simple terms.** The order was "fix six things and file a report about a seventh." Six things were
fixed, and *a* report was filed — about something else discovered along the way. The tally read "6 fixed, 1
filed", matched the order exactly, and the thing actually ordered to be filed was never written down.
Counting the outputs verified the arithmetic and not the assignment.

**Why the count is the trap, and this is the fourth distinct instance in this epic.** `defer: 1` was
*true*. Every enumerated finding was correctly recorded. The failure is only visible if you ask **"is the
specific thing that was ruled present?"** rather than **"does the number of deferrals match?"** — and a
reconciling total is precisely what stops anyone asking. Same family as D-16.R.27's scope-travels-with-the-
claim: a true statement (`defer: 1`) standing in for a different one (`F5 is registered`).

**Consequence, and it is checkable rather than exhortative:** **a ruling's disposition is verified by
locating the artifact it names, never by matching a count.** For a REGISTER disposition that means grepping
`deferred-work.md` for the finding's own subject, not counting its entries. The closer did exactly this and
it is why the loss was caught before the story closed.

**Resolution.** Filed as **DW-170**, with two of the eight claims verified fresh by the closer rather than
carried over from the review:
- `folio-go/internal/fontset/licencesignature.go:80-81` still justifies the absent build-time Apache row
  with *"no committed catalogue face is Apache-licensed"* — `robotoslab` ships `LICENSE-APACHE.txt` and the
  census classifies it `Apache-2.0`.
- `licencesignature_test.go:266`/`:292` still say *"the OFL row covers 19 of the 21 catalogue faces"*
  against a measured **28 OFL / 1 Apache / 2 UFL over 31**.

The closer verified two and **deliberately did not enumerate the other six**, correctly declining to widen
the lead's four-file bound or to restate a count it had not measured. **Behaviour, not defect** — and the
right call under D-16.R.27, whose whole subject is claims asserted beyond their measured scope.

**Three smaller findings from the same close, recorded so they are not rediscovered.**

1. **The spec's frontmatter already read `status: 'done'` at `efa9b75`**, written by the pass, while
   `sprint-status.yaml` correctly sat at `review`. D-16.R.30 established that the closer is the only writer
   of `done` — that reservation held for the tracker and **not** for the spec's own frontmatter, which the
   build writes itself at step-05. The two records are written by different hands and only one of them is
   under D-16.R.30's rule. The gates now verify it, so it is left standing; noted so the next reader does
   not read the agreement as corroboration.
2. **The triage header's severity split contradicts its own enumeration** — it reads `patch: 8 (high 1,
   medium 5, low 2)` while the eight enumerated findings are high 1, **medium 6, low 1**, because F2a and
   F7a were appended mid-pass and the header was not re-derived. Total right, per-finding record right,
   summary stale. The closer **reported rather than silently rewrote it**, which is correct: a summary
   corrected after the fact by someone who did not do the triage is a fabricated agreement.
3. **There is no `epic-16-context.md` cache**, where epics 6, 7, 8 and 15 all have one. Nothing stale to
   flag — but if `bmad-build` expects one, this epic has been running without it, and that is worth knowing
   before three more dispatches rather than after.

**How we'd know this consequence is wrong.** If verifying dispositions by artifact turns up nothing over the
remaining stories while costing a grep per ruling, it is over-engineered for a one-off. One instance is not
a pattern — but it is the *fourth* count-that-reconciled in this epic, and that is a pattern.

### D-16.R.42 — Story 16.2's gate: a timeout overruled by its own arithmetic, and a dependency admitted by a procedure rather than a preference

**Lead rulings**, on two forks the builder correctly refused to guess. Both verified independently by the
orchestrator before relay — including the one that overrules the builder, and the one where the lead ruled
rather than escalating.

**Q1(1a) — T = 30 s, not the recommended 20 s. Overruled on the recommendation's OWN arithmetic.**

The builder measured against the real host — max observed **2,097 ms** for a 24,271,604-byte face — and
proposed `AbortSignal.timeout(20_000)`, justified as *"max × 10 … keeps the worst legitimate fetch inside
budget on a connection ten times slower."* **That sentence is false at 20 s:**

```
2,097 × 10 = 20,970  >  20,000
```

**The single largest offerable face falls outside the budget the factor was chosen to cover — by 5%, on
the one case the factor exists for.** Everything below it is comfortable (p99 1,715,888 B at 365 ms →
3,650 ms), so the failure sits exactly at the tail the measurement was extended to reach. Per D-16.R.25's
positive exemplar a shipped constant must carry a reason that is **true**; at 20 s it would ship carrying
arithmetic that contradicts it.

**In simple terms.** Someone measured the heaviest thing they might have to carry, said "I'll size the bag
at ten times that," and then bought a bag slightly smaller than ten times that. Everything except the
heaviest thing fits, which is why the mistake survives casual checking — the only item that exposes it is
the one the rule was written for.

**Direction of correction settled by cost asymmetry.** Too short **wrongly refuses a legitimate pick of a
real font** — loudly, repeatably, the author retries and fails again. Too long only lengthens an
already-bounded hold on a genuine stall. T exists to prevent "dead for the session", which 30 s serves
identically. **When a bound's failure modes are "wrongly refuse a real thing" and "release a hang slightly
later", the second is always the cheaper side to err on.** The constant carries the arithmetic rather than
the prose, including the sample's own limit — one connection, one day, five reps — which is the honest
statement of why the factor is ×10 and not ×2.

**Q1(1c) — the 646 KB measurement target was a denominator error, and it was the LEAD'S OWN sentence.**
*"SPEC-fonts records shipped faces up to 646 KB"* is true of the **31 committed faces** and false of the
**1,273 fetchable families** the budget actually serves, whose tail is ~37× larger: median 107,440 B, p90
420,092 B, p99 1,715,888 B, **max 24,271,604 B**. A budget fitted to 646 KB would time out real, offerable
picks. Verified independently: `Noto Color Emoji` sits in the snapshot as `{variable: false, scripts: [],
popularity: 51}`, so it is in `webFamilies` and offerable **today**. Same class as R3's, in the lead's own
writing, caught by the builder and **reported rather than quietly fixed** — the behaviour D-16.R.22 asked
for.

**A pleasant property, recorded so nobody writes a revisit clause for it:** that face is one of DW-163's 66
script-less families. If the owner filters them at 16.3 the maximum drops to 7,881,488 B / 933 ms and T
becomes very slack. **A later filter can only ADD margin here, never remove it**, so sizing for today's
population is safe under either outcome.

**Q1(1b) — no chain deadline, and no deferral either. The property is ASSERTED instead.**

The builder's own caveat was the real risk: it argued from *current* code shape, and a future change making
a catch `continue` rather than `return refuse(...)` would make 6 × T reachable with nothing watching. **But
a deferral with a trigger is the wrong instrument for a property you can simply assert.** Register it and
you get a note that ages; assert it and the change that would break it reds. So 16.2 adds a table-driven
test that the chain terminates on first abort — for *N* across the probe loop, the licence read and the
byte read, with no further fetcher calls after — **red-proved by changing one catch to `continue`.**

**And the 6 × T premise itself was wrong — in text the orchestrator wrote into the spec from R4.** Verified
separately by the orchestrator and the lead: the catches at `:281-283`, `:339-341` and `:381` each end the
chain, and the loop's only `continue` is on a **404**, which an abort never produces. **Worst-case hold is
T plus the requests already completed — about T + 3 s, not 6 × T.** R4 demanded the arithmetic be shown;
showing it is what revealed the premise was false.

**Q2 — `fake-indexeddb` as a devDependency, RULED rather than escalated, because the constraint everyone
was deferring to does not exist.**

jsdom 28.1.0 ships no IndexedDB and the package is absent, so every store test the spec requires currently
has no backing store. The builder recommended adding it but flagged that it could not make the call, citing
a standing decision on dependency count. **The orchestrator searched for that decision and could not find
it, and told the lead not to rule around it** — if it crossed an owner decision it was the owner's. **The
lead then looked specifically for it and established it does not exist.** What exists is three other
things, all verified:

1. **A hard no-new-dependency rule scoped to the WRONG MODULE** — D-1.3.6 invariant (b), measured and
   asserted across three stories, about **`folio-go`**. `folio-designer/` is not in it.
2. **Story 5.5's project-structure note**, which is what everyone was sensing: *"No new dependency is
   justified for pickers, JSON schema, storage, sync, or downloads … audit any unavoidable package through
   AD-26 before adding it."* Its subject is a **runtime** package doing a job a browser API already does,
   and **its own disposition clause is a procedure, not a prohibition**.
3. **AD-26**, verified by the orchestrator to be a **licence boundary with CI automation** — not a count
   cap — which names `pdfjs-dist` as a shipped **Apache-2.0** dependency, making Apache-2.0 in the designer
   **precedented rather than novel**.

**Direction and licence are governed; COUNT is not.** The repo's dependency discipline is real, but it is a
culture of specific reasoned refusals whose stated procedure returns *admit* here. **Three agents reasoned
from that culture as though it were a written invariant** — which is its own finding, noted below.

**(a) over (b) for the reason the builder gave AGAINST ITSELF**, and it is the rule this run holds hardest:
*a fake written by the adapter's own author is shaped by that author's reading of the plumbing, so it
agrees with the code by construction and proves the plumbing not at all.* Both-sides-move-together vacuity,
the construction D-16.R.12 named. **An independent implementation can disagree with you, and that capacity
is the entire value of the test.** Not (c): it is (b) plus a smaller claim, and it would leave the adapter
with **no witness of any kind** until the epic catch-up — stacking a second unwitnessed adapter on
DW-161's already-unwitnessed web tier, which is D-16.R.1's own stated falsifier for the cadence.

**Four conditions, the fourth a halt.** devDependency only, asserted by a test that no shipping module
imports it; the AD-26 audit runs over the **transitive** tree because AD-26 says *"at any depth"*, reporting
every package added with its licence; `lint/MANIFEST.md` regenerated and `go test -count=1` green (DW-168's
class, load-bearing for the second time in three stories); and **if any transitive package carries a licence
outside AD-26's admitted set, STOP and return it — no substituting, no vendoring, no silent fallback.**

**One verification that removed a worry rather than adding one:** `committedLicenceFiles` shells out to
`git ls-files` (`licencecensus_test.go:381-385`), so it enumerates **tracked** files. `node_modules` is
untracked — confirmed — so the devDependency does **not** move `pinnedCensus` and does **not** move the
derived floor from D-16.R.35's ruling C. Only `MANIFEST.md` moves.

**The browser-witness residual goes to 16.3, NOT the epic catch-up.** Under every option real browser
IndexedDB is proven only by a browser. 16.3's run is already scoped to discharge DW-161's three cases and
already reloads to exercise the offline path, so a fourth case — *a stored face survives a reload and is
offered with the network disabled* — costs nothing and **retires the residual one story after it is
created** rather than at the epic's end. The store and the web tier then get their first real-browser
exercise together, which is how they are actually used.

**A silently dead gate, found at this checkpoint.** 16.2's Verification read `npm run test && npm run
build`. `npm test` exits 1 at baseline on the DW-152 red, so **the `&&` short-circuits and the build never
runs** — that gate has been measuring nothing for as long as that baseline red has existed. Split into
separate lines, both reported. Worth more than its size: a *conjunction* is a gate that silently drops
everything after its first known-failing term.

### D-16.R.43 — Two standing rules this checkpoint earned

**1. A ruled deferral is a NAMED OBLIGATION, not a slot.** D-16.R.41 recorded that F5 was ruled REGISTER
and displaced by a finding the pass discovered itself, while `defer: 1` stayed true. The lead identifies
this as the **fifth** instance of cardinality agreeing while membership differs — after the eleven families,
the 5-of-20 churn, and two others. **Consequence: a pass reporting `defer: N` NAMES which N. A ruled
deferral absent from that list is a miss regardless of what N says.** Cardinality is not membership, and a
count that reconciles is the specific thing that stops anyone checking the set.

**2. Concurrence reached by one method is one measurement, not several** — and it now has two independent
confirmations in two days. D-16.R.38 recorded three readers "independently" agreeing `licencegraph.go` was
missing, all via the same package-scoped grep. The lead adds, unprompted and against itself, that it **ran
a package-scoped grep of its own three messages ago and got lucky**. For accuracy: **the lead never ruled
on `licencegraph.go`** — that ruling was the orchestrator's, in D-16.R.37, and D-16.R.38 records it as such.
**Consequence: when citing agreement as evidence, state the METHOD. If the methods match, it is one datum
and must be discounted to one.**

**The control that actually worked, in both instances, was the same one:** a junior agent refusing on
evidence rather than deferring to concurring seniors. In the `licencegraph.go` case it refused an
orchestrator ruling. In this checkpoint's case the builder declined to add a dependency on a constraint it
believed existed, and asking rather than complying is what caused three people to look and discover the
constraint was never written down.
