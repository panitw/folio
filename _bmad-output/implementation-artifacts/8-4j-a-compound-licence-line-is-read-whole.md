---
title: 'Story 8.4j: A compound licence line is read whole'
type: 'bugfix'
created: '2026-09-02'
status: 'done'
baseline_revision: 'b4dabd96d7a19fe068f5272b173d0050eef2cb36'
baseline_revision_plan_gate: 'dbd16991148d658e4bcbd65c035eb031715bc173'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized']
deferred:
  - summary: >-
      A legitimately dual-licensed permissive FONT is now refused at the asset gate's font site. The
      classifier's honest label for a compound declaration is the whole expression ("MIT OR
      Apache-2.0"), and fontAssetLicenceAllowed is an EXACT-ID map of the owner's four ids, so no
      compound expression can ever be a member. Whether a disjunction of two allowlisted ids is
      itself an acceptable font licence is D-8.5.3's owner decision, not a builder's.
    evidence: |-
      Measured at 45985ef by probe. manifest.go:363's arm is `!fontAssetLicenceAllowed[spdx]` over
      {OFL-1.1, Apache-2.0, MIT, Ubuntu-font-1.0}. With this story's fix a font LICENSE declaring
      "SPDX-License-Identifier: OFL-1.1 OR Apache-2.0" classifies (permissive, "OFL-1.1 OR
      Apache-2.0") and fails that arm. The refusal is LOUD and fail-closed, which is why it is
      registered rather than fixed here: D-8.4j.1 bounds this story at one defect, one fix, and
      D-8.4j.2 forbids adding a second authority for a property this fix owns. NOT reachable today
      (0 of 35 population texts carry a compound line); reachable at Story 8.5, which lands ~20 font
      LICENSE files and where dual licensing is an ordinary shape.
    location: >-
      lint/internal/manifest/manifest.go:363 (SITE A) and :477 (SITE B, same shape against
      IsPermissiveSPDX)
    severity: high
    status: >-
      DISCHARGED 2026-09-02 at the build gate. This entry registered the TWO-SITE shape; D-8.4j.6
      pulled SITE A inside the story and was silent on SITE B, and D-8.4j.9 pulled SITE B inside on
      the same criterion. Both sites are fixed in this story, against DIFFERENT lists (the owner's
      four ids at SITE A, permissiveSPDX at SITE B), so D-8.5.13's two-populations-two-policies
      ruling is preserved exactly. Kept verbatim rather than deleted: the entry is the record that
      the second site was seen and registered before it was ruled on, which is why the halt was a
      good halt rather than a rediscovery.
  - summary: >-
      An SPDX line carrying trailing content after the identifier — a comment terminator, a
      parenthetical note — classifies FamilyUnknown where it previously classified on the first
      token alone.
    evidence: |-
      Measured at 45985ef: "SPDX-License-Identifier: MIT */" is (permissive, "MIT") today and
      (unknown, "") under the fix; likewise "MIT (see LICENSE)". This is the direct, intended cost
      of reading the whole line, and its direction is fail-closed and loud (D-1.3.4). ZERO of the 35
      population texts carry such a line, so nothing changes today. Registered so Story 8.5's
      arriving faces are watched for it rather than having it rediscovered as a build red.
    location: >-
      lint/internal/licence/classify.go (spdxLineRE)
    severity: medium
  - summary: >-
      "SPDX-License-Identifier:" with the identifier on the FOLLOWING line stops being a signal.
    evidence: |-
      The current pattern separates marker and capture with `\s*`, which matches newlines, so the
      capture can cross a line break. Widening the capture to the rest of the line makes that
      unsafe — it would swallow an arbitrary following line (a copyright line, say) as an
      expression — so the separator narrows to `[ \t]*`. Consequence, measured: the split-line form
      classified (permissive, "MIT") and now classifies (unknown, ""). It is not a valid SPDX
      declaration, no population text uses it, and the direction is fail-closed.
    location: >-
      lint/internal/licence/classify.go (spdxLineRE)
    severity: low
  - summary: >-
      An SPDX "WITH" exception expression — "Apache-2.0 WITH LLVM-exception", an ordinary way for a
      font or library to be published — stops being a signal and classifies FamilyUnknown.
    evidence: |-
      Measured at 66d445b by in-package probe (deleted; tree verified clean afterwards):
      ClassifyLicenceText("SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception\n") is
      (unknown, "") under this story and was (permissive, "Apache-2.0") at dbd1699, because the
      single-token capture took "Apache-2.0" and never saw the rest of the line.
      ClassifySPDXExpressionTerms rejects WITH at its operator check — `unsupported SPDX operator
      "WITH"` — and that check is UNMODIFIED by this story; what changed is that the rest-of-line
      capture now routes such a declaration to it. Same family as deferred[1] and deferred[2] (the
      priced cost of reading the line whole) but a DISTINCT shape covered by neither, and a more
      likely one: WITH-exception licensing is common where OFL/Apache faces are redistributed.
      Direction is fail-closed and loud, and the expression is left UNNAMED, so it reaches the asset
      gates' "could not be classified" arm rather than a message asserting a false classification.
      ZERO of the 35 population texts carry a WITH expression, so nothing changes today. Registered
      so Story 8.5's ~20 arriving font LICENSE files are watched for it rather than having it
      rediscovered as a build red. Whether to TEACH the parser WITH is an owner/lead question about
      SPDX grammar coverage, not a build decision, and D-8.4j.2 forbids a second authority for it.
    location: >-
      lint/internal/licence/classify.go (ClassifySPDXExpressionTerms' operator check, reached via
      spdxLineRE's rest-of-line capture)
    severity: medium
    status: >-
      REGISTERED 2026-09-02 at the close as DW-132 in deferred-work.md, owner OWNER. Re-verified there
      by driving both asset gates end to end: refused through the "could not be classified" arm at each
      site, left UNNAMED. Two further residues of the same mechanism that reached no register — an
      unsupported operator SPELLING ("MIT or Apache-2.0", ADMITTED at b4dabd9 and refused now) and the
      dependency gate's exposure to deferred[1]/[2]/[3]'s shapes via rules.ScanLicenceGraph — are
      registered as DW-133, owner ENGINEERING LEAD.
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The contract below governs.*

Our licence check reads the label a file puts on itself. When that label named two licences at once —
a file offered under either of two sets of terms, an ordinary way to publish a typeface — the check
read only the first name and threw the rest away. So a file offering itself under an acceptable licence
*or* an unacceptable one was waved through on the strength of the acceptable half, then recorded in our
published list as though the other half had never been written down. Reversing the order changed the
answer: put the unacceptable name first and the same file was correctly refused. That asymmetry was
never a policy, only an artefact of where the reading stopped.

The tool that reads such a declaration properly already existed and was simply never asked. Now it
is. Both of the places that judge a redistributed file weigh every option the file offers, each against
its own separate list of what it will accept: the list for typefaces is not the list for the dictionary,
deliberately.

This story was stopped once, partway through. The first fix closed the hole but would have started
refusing files that are perfectly acceptable — the mirror-image mistake — and that was caught, ruled on
and repaired before it shipped rather than after. A few unusual ways of writing a licence label are now
refused rather than half-read; each is written down, each refuses loudly, and none appears anywhere in
what we ship today.

<intent-contract>

## Intent

**This story does NOT come from `epics.md`, and `epics.md` must NOT be amended.** **It was created on
2026-09-02 at Story 8.4i's close by the engineering lead's ruling D-8.4j.1, on the `8.4x` insertion
precedent (2.3a / 2.5a / 2.6a / 3.1a / 8.4f / 8.4h / 8.4i).** It is Epic 8's **fifth and final**
licence-gate insertion; D-8.4j.1 restates D-8.4i.6's bound so it cannot erode again, and this defect
meets all three of its conjuncts: a bypass of a gate this epic declared fail-closed, demonstrated by
probe, reachable by the population Story 8.5 adds.

**Problem:** `spdxLineRE` captures a **single token**, so Story 8.4i's collect-all rule works across
lines but not **within** one — a licence text declaring `SPDX-License-Identifier: OFL-1.1 OR
GPL-3.0-only` classifies `(permissive, "OFL-1.1")`, sits on the owner's four-id font allowlist,
**passes the fail-closed asset gate and publishes under a permissive label in `MANIFEST.md`**; the
reversed order refuses correctly, which is DW-125's order-dependence surviving in the one place
collect-all does not reach. `ClassifySPDXExpression` already resolves such expressions correctly and
`ClassifyLicenceText` never calls it — so **D-8.4i.1, whose entire basis was "making
`ClassifyLicenceText` agree with the function it sits beside", is half-applied until this lands.**

**Approach:** Widen the SPDX-line capture to the rest of the line and route the captured expression
through the **existing** `ClassifySPDXExpression`; and make asset **admission per-term at BOTH gates**, so a
compound is admitted iff **every** term is on the governing list — the owner's four ids at the font
site, `permissiveSPDX` at the wordlist site.

**AMENDED 2026-09-02 at the build gate (D-8.4j.9); the original read "make font-asset admission
per-term, so a compound is admitted iff every term is on the owner's four-id list."** It said
*font-asset* because D-8.4j.6 was answering a question reported about the font gate; **the criterion
given with it (D-8.4j.7) was never site-scoped.** `manifest.go:477`'s arm is the identical mechanism —
a bare exact-match lookup on a label that may now be compound — and after half 1 it refuses
`CC0-1.0 OR MIT`, which `dbd1699` **admitted**. Worse, `:471` binds `wordlistFamily` and
`wordlistSPDX` from **one** `ClassifyLicenceText` call, so arm 3 would refuse with *"…does not
recognise as a permissive licence"* while `wordlistFamily` holds `FamilyPermissive` **in scope from
the same call** — the refusal contradicting its own function's other return value inside a single
switch. **D-8.5.13 separated the two sites' POLICIES (which list governs which population), not their
MECHANISM (how a label is tested against a list). Fixing a shared mechanism at both sites does not
merge two policies; it keeps them and stops both from being wrong in the same way.** **Reuse, do not rebuild:** no
second expression parser, no re-implementation of resolution inside `ClassifyLicenceText`, and no
string-splitting inside the allowlist check.

**Why admission is inside this story and is not a widening of D-8.4j.1 (ruled at the plan gate).**
`fontAssetLicenceAllowed` is an **exact-id map of four ids**, so once the label becomes the whole
expression **no compound can ever be a member**. Run the case both ways:

| declaration | today | after the label fix alone |
|---|---|---|
| `OFL-1.1 OR GPL-3.0-only` | `(permissive,"OFL-1.1")` → **admitted** (the bypass) | copyleft → **refused** ✔ |
| `OFL-1.1 OR Apache-2.0` | `(permissive,"OFL-1.1")` → **admitted, correctly** | `(permissive,"OFL-1.1 OR Apache-2.0")` → **REFUSED** ✘ |

**So the label fix alone is not an incomplete fix — it is a fix that introduces a false refusal which
does not exist in the tree today**, trading a bypass for a regression on a legitimate asset. The
governing test is: **work is inside a repair story iff excluding it leaves that story's own fix
incorrect.** Admission is inside. DW-128's wOFF/wOF2 gap is **outside and stays outside** — this fix
is correct without it (D-8.4j.4).

## Boundaries & Constraints

**Always:**
- **One defect, one fix, one red-proof (D-8.4j.1). Nothing else rides this story.** Another
  licence-gate defect found while working is **registered in `deferred:` and routed** — never added.
- `ClassifySPDXExpression` is the **sole** authority for what a compound expression resolves to.
  Family semantics (start permissive; any copyleft term → copyleft; any unrecognised term → unknown)
  are read from it, never restated.
- Everything Story 8.4i established is preserved: collect-all **across** lines; **copyleft refused AS
  copyleft, naming the identifier**; two or more distinct permissive identifiers → `FamilyUnknown`;
  resolution order **1-before-2** (a maintainer reading *"conflicting identifiers"* adds an SPDX line,
  one reading *"GPL detected"* removes the dependency — hazard indicators fail toward the loudest).
- **The BSD clause path is untouched (D-8.4i.9).** Seven of nine dependency licences carry no name
  signal and no SPDX line and classify through `"REDISTRIBUTION AND USE IN SOURCE"` — a **clause, not
  a name**. This story changes only the SPDX-line step.
- **Byte identity:** no `.folio`, no engine, nothing under `folio-go/fonts/`. **23 golden digests
  byte-identical.** `maximumCacheAssets` = 64. README md5 `078d7d80d518d54af2fc04fb270d46b8`.
- **`MANIFEST.md` regenerates with NO diff.** Measured precondition: 0 of 35 population verdicts move.
- Commit only on `main`. Never push, never branch, never `git add -A`.
- **D-8.4j.3 applies to this story's own reasoning and to its review:** a rejection resting on a true
  fact about code must name **which caller, path or population** it verified, and that must be the one
  the finding is about. Story 8.4i's R4 was rejected on a premise true for the **npm** path and false
  for the path under test — that mis-attribution is the defect this story exists to fix.

- **Admission is per-term, and no more permissive than the classification it consumes.** A compound
  is admitted iff **every** term is on `{OFL-1.1, Apache-2.0, MIT, Ubuntu-font-1.0}`; otherwise it is
  refused **naming the term that failed**, not merely the expression. This is existing direction, not
  a new call: `ClassifySPDXExpression` has resolved conservatively across terms since Story 1.3 —
  copyleft if **any** term is copyleft, unknown if **any** term is unrecognised — so it already
  declines to elect the favourable term. **If admission were more permissive than classification, the
  gate would re-open what the classifier just closed.**
- **ONE parser, ONE term enumeration, TWO consumers.** The resolved signal carries its **term set**
  alongside its label; `ClassifySPDXExpression` (or a sibling that returns terms instead of discarding
  them) stays the sole enumerator; admission tests the set. **Never `strings.Split(spdx, " OR ")`
  inside the allowlist check** — that is a second SPDX expression parser, a shape this run has now
  found four times.
- **The label and the gate key on different things, and that is correct — say so in TWO places.** The
  label states **what the file says**; the gate states **whether every option the file offers is
  acceptable**. A later reader seeing `OFL-1.1 OR Apache-2.0` in `MANIFEST.md` beside a four-id
  allowlist will file that as a bug, so state it at the admission site **and** in `MANIFEST.md`'s own
  header. **A divergence that is correct and undocumented gets "fixed" by the next person.**

**Block If:**
- The build's own population re-measurement **disagrees** with this spec's record of **35 texts, 0
  verdict changes**. The disagreement is itself the halt (D-8.4i.9's shape) — never a reason to weaken
  the rule, narrow the population or exempt a file (D-8.5.10).
- A golden digest moves, or `genmanifest` produces any diff. A legitimately changed label is a
  **finding to route, not a re-record**.
- Making the red-proofs pass requires changing `ClassifySPDXExpression`'s **resolution semantics**,
  `licenceNames`, `licenceClauses`, or `resolveLicenceSignals`'s arm order. Any of those means the
  composition rule below is wrong, and that is a lead question, not a build decision.
- **Either gate's LIST changes, or the SET OF LICENCES either gate admits changes.**
  **AMENDED 2026-09-02 at the build gate (D-8.4j.9); the original said "or either asset gate's
  policy", which is too broad and halts on the fix itself.** Rebound to **policy, not mechanism**:
  this keeps every tooth the guard was given — it still stops a silent widening of D-8.5.3 — while
  permitting a mechanism repair that provably changes no admission decision. **Relative to
  `dbd1699`, the SITE B fix changes admission only for expressions containing a non-permissive term
  — the bypass class, which is this story's subject. Every all-permissive expression is admitted
  exactly as before. No licence's admission status changes.** If a build finds otherwise, that is
  the halt.

**Never:**
- **No compound-expression BAN anywhere (D-8.4j.2).** A guard asserting "no `LICENSE*` carries a
  compound expression" is a **proxy** for the property this fix owns, would refuse a legitimately
  dual-licensed permissive face, and is two authorities for one fact.
- **Do not build Story 8.5's census guard here (D-8.4j.2).** 8.5's census must later pin at least one
  compound case; that is 8.5's task, recorded and not started here.
- **DW-128 (the wOFF/wOF2 tripwire gap) does not ride this story (D-8.4j.4).** Nor does DW-127/DW-129,
  DW-130, catalogue faces, `@font-face` rules, picking behaviour, byte thresholds (8.4d), or
  bold/italic (Epic 11).
- No edit to `epics.md`, `ARCHITECTURE-SPINE.md`, or this `<intent-contract>`.
- No new file in the licence-census population (no new `LICENSE*`/`COPYING*`), so `pinnedCensus` stays
  a 35-entry hand-written literal the code under test cannot derive (D-8.4i.11).

## I/O & Edge-Case Matrix

All rows measured at `45985ef` by throwaway in-package probe, deleted before this spec was written.

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| RP1 — compound copyleft | `SPDX-License-Identifier: OFL-1.1 OR GPL-3.0-only` | `(copyleft, "OFL-1.1 OR GPL-3.0-only")`. Today: `(permissive, "OFL-1.1")` | Gate refuses as copyleft, naming the declaration |
| RP2 — reversed order | `SPDX-License-Identifier: GPL-3.0-only OR OFL-1.1` | `(copyleft, "GPL-3.0-only OR OFL-1.1")` — same verdict as RP1 | Order-independent refusal |
| RP3 — permissive OR permissive | `SPDX-License-Identifier: MIT OR Apache-2.0` | `(permissive, "MIT OR Apache-2.0")` — **RESOLVES, is not refused** | No error expected |
| RP3′ — reversed | `SPDX-License-Identifier: Apache-2.0 OR MIT` | `(permissive, "Apache-2.0 OR MIT")` — resolves | No error expected |
| Conjunction | `SPDX-License-Identifier: MIT AND GPL-3.0-only` | `(copyleft, "MIT AND GPL-3.0-only")`. Today: `(permissive, "MIT")` | Refused as copyleft |
| Compound line over a matching body | compound permissive header above the committed OFL 1.1 text | `(permissive, "OFL-1.1 OR Apache-2.0")` — the body's OFL name signal must NOT count as a second identifier | No error expected |
| Compound line over an unrelated body | `MIT OR Apache-2.0` header above the OFL text | `(unknown, "")` — a genuine conflict, preserved | Refused as unclassifiable |
| Compound line over a GPL body | `MIT OR Apache-2.0` header above the GPL text | `(copyleft, "GPL-3.0-only")` — DW-125 with roles swapped, still caught | Refused as copyleft |
| Two distinct compound lines | `MIT OR Apache-2.0` and `BSD-3-Clause OR ISC` | `(unknown, "")` — arm 3 across lines, unchanged | Refused as unclassifiable |
| Same compound line twice | the identical declaration on two lines (any spacing) | `(permissive, "MIT OR Apache-2.0")` — one signal, not two | No error expected |
| Parenthesised (DW-131's pin) | `SPDX-License-Identifier: (MIT OR Apache-2.0)` | `(unknown, "")` — **byte-identical to today**, fail-closed | "could not be classified" arm |
| Single identifier (control) | `SPDX-License-Identifier: MIT` | `(permissive, "MIT")` — unchanged | No error expected |
| Single unknown identifier (control) | `SPDX-License-Identifier: CC-BY-SA-4.0` | `(unknown, "CC-BY-SA-4.0")` — **still names the id it could not place** | Reaches the "not a permissive licence" arm, not the neighbouring one |
| Compound with an unknown term | `SPDX-License-Identifier: MIT OR CC-BY-SA-4.0` | `(unknown, "")`. Today: `(permissive, "MIT")` | "could not be classified" arm |
| Trailing content | `SPDX-License-Identifier: MIT */` | `(unknown, "")`. Today: `(permissive, "MIT")` — registered as deferred | Fail-closed, loud |

</intent-contract>

## Code Map

- `lint/internal/licence/classify.go:157` — `spdxLineRE`, `(?i)SPDX-License-Identifier:\s*([A-Za-z0-9.\-+]+)`.
  **The single-token capture is the defect.** Note the separator is `\s*`, which matches newlines.
- `lint/internal/licence/classify.go:14-40` — `ClassifySPDXExpression`. **Reuse target.** Trims, rejects
  `()` and even field counts, requires `AND`/`OR` at odd indices, starts `FamilyPermissive`, any
  copyleft term → `FamilyCopyleft`, any unrecognised term → error + `FamilyUnknown`. Returns
  **`(Family, error)` — no identifier**, which is why the composition rule below must decide the label.
- `lint/internal/licence/classify.go:300` — `classifyBySPDX`, exact map lookup + copyleft prefixes.
  Unchanged.
- `lint/internal/licence/licencesignals.go:243-312` — `collectLicenceSignals`. **Only step (1)
  (`:270-275`, the `FindAllStringSubmatch` loop) changes.** Step (2) names and step (3) clauses are
  untouched; `namedAny` and the clause-only-when-unnamed rule are untouched.
- `lint/internal/licence/licencesignals.go:249-268` — the `addID` closure and its `seen` dedup map. The
  new expression path must participate in `seen`, and must NOT route a whole expression through
  `classifyBySPDX` (an exact lookup would call every compound expression unrecognised).
- `lint/internal/licence/licencesignals.go:341-354` — `resolveLicenceSignals`, D-8.4i.1's five arms.
  **Untouched.** Composition across lines is entirely this function's job.
- `lint/internal/licence/classify_test.go` — `TestClassifyCollectsEverySignal:432`,
  `TestCopyleftTieBreakIsDeterministic:639` are the shape to follow for the new named tests.
- `lint/internal/licence/licencecensus_test.go:47-84` — `pinnedCensus`, 35 hand-written entries
  (D-8.4i.11). **Read-only for this story**; helpers `committedLicenceFiles`, `repoRootForCensus`,
  `ResolveGraph`, `ReadLicenceText`.
- `lint/internal/manifest/manifest.go:357-365` — **SITE A**, the font gate: `spdx == ""` /
  `family == FamilyCopyleft` / `!fontAssetLicenceAllowed[spdx]`. Three disjoint arms; the fix must not
  move a case from one arm to another except where the matrix says so.
- `lint/internal/manifest/manifest.go:471-479` — **SITE B**, the wordlist, same shape against
  `IsPermissiveSPDX`. **Different policy on purpose — do not collapse.**
- `lint/internal/manifest/manifest.go:81` — the dependency-row label path (`_, spdx :=`); publishes the
  identifier verbatim into `MANIFEST.md`.
- `lint/internal/rules/licencegraph.go:44` — the dependency scan; switches on **family only**, so a
  compound permissive dependency passes and is labelled honestly.
- `lint/internal/manifest/manifest_test.go:511` — `TestResolveAssetsRefusesTheDW125BypassAtTheGate`,
  the scratch-repo recipe (`t.TempDir` → `initGitRepo` → fake `.ttf` → `gitAdd`, **add only, no
  commit**) to copy for the gate-surface red-proof.
- `_bmad-output/implementation-artifacts/deferred-work.md:6468` — DW-131, this story's charter entry.

## Tasks & Acceptance

**Execution:**

- `lint/internal/licence/classify.go` — widen `spdxLineRE` to capture the rest of the line:
  `(?i)SPDX-License-Identifier:[ \t]*(\S[^\n\r]*)`. **Three deliberate details, each with a reason to
  record in the comment:** `[ \t]*` replaces `\s*` because a newline-crossing separator plus a
  rest-of-line capture would swallow an arbitrary following line as an expression; `\S` first keeps a
  bare `SPDX-License-Identifier:` with nothing after it a non-signal, exactly as today; `[^\n\r]*`
  stops at the line end and tolerates CRLF.
- `lint/internal/licence/licencesignals.go` — replace step (1)'s `addID(m[1])` with the expression
  path. **Normalize** the capture to single-spaced fields (so `MIT  OR   Apache-2.0` and
  `MIT OR Apache-2.0` are one signal), dedup it through the existing `seen` map, then call
  `ClassifySPDXExpression` and translate its answer, per **the composition rule in Design Notes**.
  **Do not** compute a family from the terms yourself. Rewrite the step's doc comment to state the
  composition rule and why one line is one signal.
- `lint/internal/manifest/manifest.go` — make SITE A's admission **per-term**. `:363`'s
  `case !fontAssetLicenceAllowed[spdx]` becomes a check over the resolved signal's **term set**:
  refused unless every term is in `fontAssetLicenceAllowlist`, and the refusal message names **the
  failing term**. **Do not split the string here** — consume the term set the classifier already
  produced. Add the comment stating why label and gate key on different things, and add the matching
  note to `MANIFEST.md`'s header.
- `lint/internal/licence/` — expose the resolved expression's **term set** alongside its family,
  either from `ClassifySPDXExpression` or from a sibling that returns terms instead of discarding
  them. It currently enumerates via `strings.Fields` and returns `(Family, error)`, exposing none.
  **This is the only sanctioned modification to that function**; its resolution semantics must not
  change.
- `lint/internal/licence/classify_test.go` — add the **four** red-proof cases as **separately named
  tests with distinct failure messages**: the compound-copyleft refusal, its reversed twin
  (order-independence), the permissive-OR-permissive case that **resolves and is admitted**, and
  **`OFL-1.1 OR CC0-1.0`** — permissive and classifiable but **not** on the four-id list — which must
  red **naming `CC0-1.0` specifically**. **The fourth is what proves admission is per-term rather
  than all-or-nothing**, and the third is what proves the fix is a classifier and not a ban. Add the matrix's
  controls in the same file: single identifier, single unknown identifier, parenthesised form,
  compound-with-unknown-term, the two same-text/different-body cases, the two-distinct-lines case,
  and the same-line-twice dedup.
- `lint/internal/manifest/manifest_test.go` — add a **gate-surface** red-proof for RP1 and RP2 on the
  scratch-repo recipe: a tracked font directory whose `LICENSE` declares `OFL-1.1 OR GPL-3.0-only`
  (and, in a second case, the reverse) must fail `ResolveAssets` **through the copyleft arm**, with
  the message naming the declaration. **RP3 is proved at the classifier only** — see Design Notes on
  why it cannot be proved at the font gate, and the `deferred:` entry that owns the reason.
- `lint/internal/licence/licencecensus_test.go` — **no edit.** Run it and record that all 35 pinned
  verdicts still hold. Do **not** add rows, and do **not** cite the census as evidence that this story
  *did* anything — see Design Notes.

**Acceptance Criteria:**

- Given a font directory whose `LICENSE` declares `SPDX-License-Identifier: OFL-1.1 OR GPL-3.0-only`,
  when `ResolveAssets` runs over it, then the build **fails through the copyleft arm** with a message
  naming the declaration — and given the two identifiers are **reversed**, then the verdict and the
  arm are the same.
- Given a licence text declaring `SPDX-License-Identifier: MIT OR Apache-2.0`, when
  `ClassifyLicenceText` classifies it, then it returns `FamilyPermissive` — the fix **resolves** a
  wholly-permissive compound expression rather than refusing it — and the same holds with the two
  identifiers reversed.
- Given a font directory whose `LICENSE` declares `OFL-1.1 OR Apache-2.0` — **both on the four-id
  list** — when `ResolveAssets` runs, then it is **admitted**, and the manifest row carries the whole
  expression as its label. **This is the regression guard: the label fix alone would refuse it.**
- Given a font directory whose `LICENSE` declares `OFL-1.1 OR CC0-1.0` — permissive and classifiable,
  but `CC0-1.0` is **not** one of the owner's four — when `ResolveAssets` runs, then it is **refused
  with a message naming `CC0-1.0`**, the failing term, rather than naming only the expression.
- Given the admission check, when it is inspected, then it consumes a **term set produced by the
  single existing enumerator** and performs **no string splitting of its own** — one parser, one term
  enumeration, two consumers.
- Given a wordlist `LICENSE` declaring `CC0-1.0 OR MIT` — **both permissive**, and `CC0-1.0` is
  `permissiveSPDX`'s deliberate member since Story 2.1 — when `resolveWordlistAssetRow` runs, then it
  is **admitted**. **`dbd1699` admitted it; half 1 alone refuses it, with a message contradicting
  `wordlistFamily` in scope from the same call.** (RP5)
- Given the seen-marking, when a text declares `MIT` on one line and `MIT OR Apache-2.0` on another,
  then it resolves to **the same verdict in both orders** — the conflict arm fires either way.
  **Both orders must be asserted in the test: order-independence is the property, and a single-order
  proof passes by luck.** (RP6)
- Given a `MIT OR Apache-2.0` header over a **GPL body**, when it is classified, then it still refuses
  **as copyleft** — the masking falsifier, **re-measured after the seen-marking moved**, never carried
  forward on inspection.
- Given each of the **six** red-proofs in turn, when the guard it exercises is **deleted** (not
  falsified) and the suite re-run, then **exactly that one named test reds, on its own message**, and
  no other test reds — proved for each of the six independently. **All four pre-existing proofs are
  RE-RUN after the rulings are folded in: they are proofs over the old patch and are not proofs over
  the new one.**
- Given the whole committed-plus-dependency licence population, when the census runs, then **35 texts
  classify exactly as pinned and no verdict moves**, and `genmanifest` run twice leaves
  `git diff --exit-code` clean from the repo root.
- Given the fix is in place, when `licenceNames`, `licenceClauses` and `resolveLicenceSignals` are
  inspected, then **none of them has been modified**. `ClassifySPDXExpression` may be modified **only**
  to expose its term set (or gain a sibling that does); **its resolution semantics — copyleft if any
  term is copyleft, unknown if any term is unrecognised — must be byte-for-byte unchanged in
  behaviour**, and the census is the witness that they are.
- Given the 23 golden digests and the README, when the witness runs, then every digest is
  byte-identical and the README md5 is `078d7d80d518d54af2fc04fb270d46b8`.

## Design Notes

### The known cost of conservative admission, named now with its escape hatch

The per-term rule **refuses `OFL-1.1 OR <proprietary>`** — a real font-business shape — even though the
OFL term is takeable on its own. That is accepted deliberately: the refusal is **loud**, naming the
directory and the failing term, so a real case surfaces at build time rather than shipping mislabelled.

**If Story 8.5 hits a face whose only availability is permissive-OR-proprietary, that is an OWNER
question about electing a term. It is not resolved in the build, and it is not resolved by widening
the allowlist.** Registered here so it arrives as a scheduled question rather than a surprise.

**Admitting by the first permissive term was considered and REJECTED on a sharper ground than
"it answers an owner decision silently".** DW-131 *is* first-term-wins in the classifier; admitting by
first term is first-term-wins **in the gate**. Fixing one while implementing the other in the next
function along is the worst available outcome — because the story's own record would then say the
defect was closed.


### The seen-marking scopes to the NAME SIGNAL space — CORRECTED at the build gate (D-8.4j.10)

**The original rule read: "mark every whitespace-separated field of a resolved expression as seen."
That is preserved here verbatim, and it is wrong in two ways the build measured.**

**Defect.** It writes term ids into the **shared** `seen` map, so it swallows a later SPDX **line**,
not just a body name signal. Measured at `1af9854`: `"MIT"` then `"MIT OR Apache-2.0"` →
`(unknown, "")`; **reversed** → `(permissive, "MIT OR Apache-2.0")`; at `dbd1699` **both** orders gave
`(permissive, "MIT")`. **A new cross-line order dependence, introduced by the story whose subject is
order dependence.** Fail-closed rather than a bypass, and the copyleft pair is order-independent
(measured both ways) — but a new order dependence from this story must not ship, reachable or not.

**Correction 1 — scope the marking to the NAME/BODY signal space, not the shared map.** The sub-rule's
justification does not extend to a second SPDX line. A dual-licensed font ships **one** licence's
body, so suppressing a duplicate **body name signal** is correct. **Two SPDX lines are two explicit
declarations**, and a file declaring `MIT` on one line and `MIT OR Apache-2.0` on another **genuinely
says two things** — D-8.4i.1's conflict arm should fire, and scoped this way it fires **in both
orders**, which is the property this story exists to establish. Stricter, consistent, fail-closed.

**Correction 2 — mark TERMS, not whitespace-separated FIELDS.** As written the rule marks the
**operators** `OR` and `AND` as seen ids. Harmless today only because no licence is named "OR". Take
the term list from the **single enumerator**, never by re-splitting — the same one-parser guardrail as
the admission fix, and **the fifth place in this area where a second splitter would have been the easy
thing to write.**

**Correction 3 — the masking falsifier is RE-RUN, not carried.** `MIT OR Apache-2.0` header over a
GPL body must still refuse **as copyleft**. It survives on inspection (GPL is a name signal, not a term
of the expression), **and inspection is not the standard this run holds**: the property was established
by measurement and must be re-established after the mechanism moves.

### THE COMPOSITION RULE (state it, do not infer it)

D-8.4j.1 asks for the interaction between a compound expression and 8.4i's multi-line collect-all to
be settled here. It is:

> **One SPDX line contributes exactly ONE signal, whatever its arity. Resolution WITHIN a line is
> `ClassifySPDXExpression`'s job; composition ACROSS lines and against name signals is
> `resolveLicenceSignals`'s job. Neither reaches into the other.**

Concretely, per line, after normalizing the capture to single-spaced fields:

- **resolves copyleft** → append **the whole expression** to `copyleftIDs`;
- **resolves permissive** → append **the whole expression** to `permissiveIDs` as one identifier;
- **does not resolve** (error, or `FamilyUnknown`) → set `unresolved`, and set `unresolvedID` **only
  when the expression is a single term** — today's behaviour, unchanged. A compound expression that
  fails to resolve names nothing, exactly as an unresolved licence NAME names nothing. *This detail is
  load-bearing:* naming a malformed expression would move a font from SITE A's *"could not be
  classified"* arm to its *"not one of the permitted licences"* arm, whose message would then assert
  the text *classifies as* something it demonstrably does not. It would also break the arm-disjointness
  `TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist` pins, and would move DW-131's own
  requested pin for the parenthesised form away from `(unknown, "")`.
- **additionally**, when a line resolves, mark **every whitespace-separated field of it** as seen. No
  grammar knowledge is needed or wanted — operators are marked too and are inert, since no name-table
  or clause id is `OR`/`AND`. **Why it is required:** a real dual-licensed font ships the full text of
  one of its licences. Without this, `OFL-1.1 OR Apache-2.0` above the committed OFL body yields the
  expression *and* the body's `OFL-1.1` name signal — two permissive identifiers — and arm 3 refuses a
  file that says one thing. Measured: `(unknown, "")` without the marking, `(permissive, "OFL-1.1 OR
  Apache-2.0")` with it. It cannot mask a *different* licence's signal, because only the terms the
  expression itself declared are marked — a `MIT OR Apache-2.0` header over a GPL body still refuses as
  copyleft, measured.

Everything else composes for free and is **unchanged code**: two lines each resolving permissive give
two distinct permissive identifiers and arm 3 refuses (measured); any line resolving copyleft gives
arm 1 (measured); a single-term line behaves byte-identically to today, because
`ClassifySPDXExpression` on one term *is* `classifyBySPDX` on that term.

### Why the identifier is the WHOLE expression and not its first term

The alternative — label a permissive compound with its first term, so `MIT OR Apache-2.0` reports
`MIT` — would let RP3 pass the font gate end to end and would raise no deferred finding. It was
rejected, and the ground is recorded rather than assumed:

1. D-8.4i.1 and the whole 8.4h/8.4i thread hold that **the label is the defect**: *"the family verdict
   is permissive either way; the LABEL was wrong, and the label is what lands in `lint/MANIFEST.md`"*,
   which AD-26 names a release artifact. Attributing a dual-licensed file to one of its two licences is
   a partial label produced by reading part of the line — this story's own subject.
2. **Direction of error.** D-2.1.3, binding: *"an allowlist whose miss is LOUD is a fail-safe; an
   allowlist whose miss is SILENT is a rotting list."* Whole-expression labelling fails **loudly** (the
   gate refuses and names the declaration); first-term labelling fails **silently** (a wrong label on a
   green build).
3. **Whose decision it is.** Whether a disjunction of two allowlisted ids is itself an acceptable font
   licence is a question about D-8.5.3's four-id **owner** decision. First-term labelling answers it
   silently, in the affirmative. Whole-expression labelling leaves it to the owner, loudly, and it is
   registered as this spec's first `deferred:` entry — routed to the lead **before Story 8.5 lands the
   ~20 faces that make it reachable**. That is D-8.4j.1's *register and route* instruction, not an
   evasion of it.

### What the census can and cannot witness (D-8.4i.11)

The census is cited here for exactly one claim: **this story moves no existing verdict** (35 texts, 0
changes, measured). It is **not** evidence that the story does anything — no population text carries a
compound line, so corrupting the new expression branch leaves the census green. Saying so is the whole
point of D-8.4i.11: a differential whose two sides can converge is not a witness, and neither is a
population that cannot express the input. The witnesses that the fix *works* are the three named
red-proofs and their independent deletion probes. Do not report the census as if it covered this
change.

### Measurement record — taken at the plan gate, 2026-09-02, at `45985ef`

By throwaway in-package probe (deleted; tree verified clean before this spec was written). Every
matrix row above was measured under the exact rule specified here, plus: **35 population texts, 8 of
which carry an SPDX line, 0 capture differences between the old and new patterns, 0 verdict changes.**
Non-vacuity control: the same sweep reports `MIT OR GPL-3.0-only` moving from `(permissive, "MIT")` to
`(copyleft, "MIT OR GPL-3.0-only")`, so a sweep that could not report a change was ruled out. **If the
build's own run disagrees with any figure here, the disagreement is itself the halt** — this record is
a second, independent check, not a substitute for running it.

## Verification

Cadence **PER-EPIC** (D-000.4). **Exclude** the four `FOLIO_MATRIX_TARGET` legs,
`TestCrossTargetByteIdentity` and Playwright — Epic 8's boundary gate.

**Commands:**
- `cd lint && go test -count=1 ./...` — expected: all packages pass. **`-count=1` is mandatory**: this
  module's rules packages walk directories, and Go's test cache does not track `ReadDir`, so a cached
  `ok` here is not a measurement.
- `cd lint && go test -count=1 -run 'TestLicenceSignalCensus' -v ./internal/licence/` — expected: 35
  texts measured, every verdict matching its pin.
- `cd lint && go vet ./...` and `cd folio-go && go vet ./...` — expected: clean.
- `cd folio-go && go test ./...` — expected **1815 pass / 2 fail / 5 skip**. The two fails are the
  mandated standing red `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`. **Never "fix" it.**
- `cd lint && go run ./cmd/genmanifest` **twice** (it must be run from inside `lint/` — RELEASING.md
  says so and there is no root `go.mod`), then **from the repo root** `git diff --exit-code` —
  expected: no diff. A legitimately changed label is a **finding to route, not a re-record**.
- **From the repo root**: `gofmt -l folio-go lint` — expected: **exactly**
  `lint/internal/rules/licencegraph_test.go` (DW-116, pre-existing — do NOT reformat). Run from inside
  `lint/` it prints `lstat` errors that read as clean; Story 8.4i's close caught this live.
- `cd folio-designer && npm run test && npm run typecheck && npm run lint && npm run build && npm run test:e2e:compile`
  — expected: **40 files / 411 tests** pass; lint prints **exactly 4** `only-export-components`
  warnings; build and e2e compile succeed.
- The **23-digest golden witness** is `TestGoldenDigestAgreesAtEveryDeclaredSite`
  (`folio-go/byte_neutrality_test.go:631`), which rides the `folio-go` run above — expected: all 23
  entries agree at every declared site, no digest moved.
- `md5 -q README.md` from the repo root — expected **exactly**
  `078d7d80d518d54af2fc04fb270d46b8` (re-measured at `45985ef`; this is the repo-root `README.md`).
- `maximumCacheAssets` = 64, enforced by `folio-designer/scripts/offline-release-contract.test.mjs`,
  which rides `npm run test` above. Confirm it appears in that run rather than asserting it by eye.
- **Mutation proof, three independent runs:** delete (do not falsify) the guard behind each red-proof
  in turn and re-run `cd lint && go test -count=1 ./...` — expected each time: **exactly one named
  test reds, on its own message**, and no other.

## Auto Run Result — plan gate (2026-09-02)

Status: ready-for-dev
Blocking condition: none — the dispatch directed `Halt after planning.`

Planning only. **No production or test code was written**, and no commit was created. The working
tree carries this spec and nothing else. The plan-gate probe used to take the measurements recorded
above was a throwaway in-package `_test.go`, deleted before the spec was written;
`git status --porcelain` was verified empty afterwards.

## Review Triage Log

### 2026-09-02 — Review pass

- intent_gap: 1: (high 1, medium 0, low 0)
- bad_spec: 2: (high 0, medium 1, low 1)
- patch: 2: (high 0, medium 0, low 2)
- defer: 4: (high 0, medium 2, low 2)
- reject: 12
- addressed_findings:
  - none

**Attempted implementation saved at** `8-4j-attempted-implementation.patch` (this directory). Code
changes reverted; the implementation commit `1af9854` was reversed by `git revert`, not rewritten.

#### intent_gap (1) — the halt

**[high] SITE B (the wordlist gate) is left on exact-id admission, and this story's own half-1 change
makes it refuse a wholly-permissive compound with a message asserting the opposite of the
classifier's verdict.** `resolveWordlistAssetRow`'s third arm is
`case !licence.IsPermissiveSPDX(wordlistSPDX):` — a bare map lookup, so once the label may be a whole
expression **no compound can ever be a member**, which is verbatim the argument D-8.4j.6 used to rule
per-term admission *inside* this story at SITE A.

Measured at `1af9854`, in-package probe (deleted; tree verified clean afterwards):

- `ClassifyLicenceText("SPDX-License-Identifier: CC0-1.0 OR MIT\n")` → `(permissive, "CC0-1.0 OR MIT")`
- `licence.IsPermissiveSPDX("CC0-1.0 OR MIT")` → `false`
- so SITE B refuses with `wordlist licence text classifies as "CC0-1.0 OR MIT", which this project
  does not recognise as a permissive licence` — a refusal whose message contradicts the
  classification it just consumed.
- At `dbd1699` the same text labelled `(permissive, "CC0-1.0")` and was **admitted**. The false
  refusal is **new**, and is created by this story.

**Why this cannot be resolved from the spec.** The contract points both ways and never settles SITE B:

- *Inside:* the Boundaries rule **"Admission is per-term, and no more permissive than the
  classification it consumes… If admission were more permissive than classification, the gate would
  re-open what the classifier just closed"**; and D-8.4j.7's criterion — half 1 is consumed by SITE B,
  and at SITE B half 1 alone is exactly the "fix that introduces a false refusal which does not exist
  in the tree today" that D-8.4j.6 declared *not an incomplete fix but a wrong one*.
- *Outside:* the Approach scopes the second half to **"font-asset admission"**; the Code Map says
  SITE B is **"Different policy on purpose — do not collapse"**; `deferred[0]` already registers the
  same shape at `:477`; and the Block If names **"either asset gate's policy"** as *"a lead question,
  not a build decision"*.

D-8.4j.6 pulled one half of a **two-site** registered finding inside the story and was silent on the
other half. That silence is the gap. **Question for the lead:** does SITE B become per-term against
its own permissive set in this story, or does 8.4j ship the new SITE B false refusal with the residue
re-registered and routed? The answer changes what code exists, so it is asked before the code is
re-derived rather than after.

#### bad_spec (2) — moot under the cascade, recorded so the amendment is not lost

- **[medium] `markExpressionFieldsSeen` writes term ids into the SHARED `seen` map, so a resolved
  expression silently swallows a LATER SPDX LINE, not merely a body name signal — and the verdict
  becomes ORDER-DEPENDENT ACROSS LINES.** Measured at `1af9854`:
  `"SPDX-License-Identifier: MIT\nSPDX-License-Identifier: MIT OR Apache-2.0\n"` → `(unknown, "")`,
  while the same two lines **reversed** → `(permissive, "MIT OR Apache-2.0")`. At `dbd1699` both
  orders gave `(permissive, "MIT")` — order-**in**dependent. The copyleft pair is unaffected (both
  orders `(copyleft, "MIT OR GPL-3.0-only")`, measured), so the direction is fail-closed and loud, not
  a bypass. Root cause is the **Design Notes** composition rule ("mark **every whitespace-separated
  field** of it as seen"), which is outside `<intent-contract>` and can be amended: the marking needs
  its own set, consulted by steps (2) and (3) only, not the map that dedups SPDX lines. Order
  dependence is the very defect this story exists to remove; introducing a new one across lines while
  closing one within a line is the outcome D-8.4j.6 called *the worst available*.
- **[low] `deferred[0]` is stale against the ruled reading.** It states the compound-font-admission
  problem at `manifest.go:363 (SITE A)` is *"registered rather than fixed here"* — which D-8.4j.6
  reversed. Its SITE B half is still live and is now the intent_gap above. The entry should be split.

#### patch (2) — moot under the cascade, recorded for re-derivation

- **[low]** SITE A's arm-3 message duplicates itself for a **single-term** expression:
  `classifies as "CC-BY-SA-4.0", whose term "CC-BY-SA-4.0" is not one of…`. Name the term only when it
  differs from the expression.
- **[low]** `ClassifySPDXExpressionTerms`' doc comment says it has *"three callers"*, two of which are
  `rules.ScanLicenceGraph` and `npm.go` — those two call the **wrapper**, not the sibling. In a file
  whose whole style is precision about counts, the count is wrong.

#### defer (4) — moot under the cascade; NOT written to `deferred:` this pass

- **[medium]** Fail-open if `fontAssetLicenceAllowlist ⊄ permissiveSPDX`: an allowlisted-but-
  unrecognised id would yield `(unknown, id)`, clear arms 1 and 2, and be **admitted**. Containment
  holds today (measured: all four of OFL-1.1, Apache-2.0, MIT, Ubuntu-font-1.0 are permissive), and
  the arm shape is **unchanged from `dbd1699`**, so this is pre-existing — but nothing pins it, and
  `classify.go`'s own `Ubuntu-font-1.0` comment records that the containment was absent once before.
- **[medium]** Trailing content after an identifier poisons the **whole text** through arm 2 even when
  the body is unambiguous, and that reaches `rules.ScanLicenceGraph` — a build-failing
  "licence unresolvable" for a future dependency whose LICENSE is ordinary. `deferred[1]` registers
  the cost against the classifier and the font gate, not against the **dependency** gate.
- **[low]** `manifest.Generate` writes the same possibly-compound label into the **module-dependency**
  table, but the label/gate divergence note sits under the **assets** heading only.
- **[low]** No pin for arity ≥ 3 expressions, lowercase operators (`MIT or Apache-2.0`), unknown
  operators, or `WITH` exception expressions — all of which the rest-of-line capture feeds to the
  parser for the first time.

#### reject (12) — enumerated individually, each naming the path actually verified (D-8.4j.3)

1. *"Partially enumerated expression admits when its enumerated prefix is all allowlisted."* Verified
   at the path the finding is about — `ResolveAssets`' SITE A, not the classifier in isolation. A
   non-resolving **compound** sets `unresolvedID = ""`, so `spdx == ""` and **arm 1 fires before arm
   3**; a non-resolving **single** term yields `terms = [that id]`, which is not allowlisted
   (containment measured above). No input reaching SITE A carries a non-empty label with a
   partially-enumerated allowlisted prefix.
2. *"`WITH` exception operator misclassified."* Verified at `ClassifySPDXExpressionTerms`' operator
   check, which is **unmodified** by this diff — `WITH` was already refused. The only change is that
   the rest-of-line capture now reaches it, which is the same class as `deferred[1]`; recorded in the
   defer list above rather than as a distinct defect.
3. *"Two lines declaring the same disjunction in opposite term order should not conflict."* Verified
   against the matrix row that rules it: *"Two distinct compound lines → `(unknown, "")` — arm 3
   across lines, unchanged."* `MIT OR Apache-2.0` and `Apache-2.0 OR MIT` are distinct expressions;
   sorting terms to equate them is a second opinion about SPDX grammar the contract forbids.
4. *"Mark only even-index terms, not every whitespace field."* Verified at `markExpressionFieldsSeen`
   against `licenceNames` and `licenceClauses`: neither table has an entry with id `OR` or `AND`, so
   the operators are inert. (The **shared-map over-reach onto later SPDX lines** is a different claim
   and is NOT rejected — it is the bad_spec finding above.)
5. *"A family that is neither copyleft nor permissive falls through the switch."* Verified at
   `ResolveAssets`' switch: the only third family is `FamilyUnknown`, which arrives with `spdx == ""`
   except for the single-unrecognised-id case, caught by arm 3 through containment. Same ground as
   reject 1; the residual is the deferred containment pin.
6. *"Callers relying on `addID` to name an unresolved SPDX identifier."* Verified at
   `collectLicenceSignals`: after the change step (1) is the **only** writer of `unresolvedID`, and
   `addID` has no SPDX caller. The concern is about a hypothetical future path, not a present one.
7. *"Comment-to-code ratio is too high."* Verified against the **Tasks & Acceptance** section, which
   mandates each of these comments by name (the three regex details, the composition rule, the
   label/gate divergence in two places). Style objection against an explicit instruction.
8. *"`ScanLicenceGraph` now prints a whole expression where it printed one id."* Verified at
   `rules/licencegraph.go:44` and against the census: that honest label **is** the story's subject,
   and all 9 dependency verdicts are pinned unchanged (35/35 hold).
9. *"`firstDisallowedFontLicenceTerm`'s signature fights its name."* Style. Its one real consequence —
   the whole input landing in the term slot — is the patch finding above, not a separate defect.
10. *"The zero-terms branch is untested / may be unreachable."* Verified at `ResolveAssets`: it **is**
    unreachable, because arm 1 catches `spdx == ""` first, and the code's own comment says so.
11. *"The term set does not travel with the signal; the expression is enumerated twice."* Verified
    against the contract's wording, which requires one **enumerator** and explicitly offers *"a
    sibling that returns terms"*. Calling the sole enumerator twice is not a second parser.
12. *"`MANIFEST.md` was hand-edited, breaking the Block If."* Verified by running `genmanifest` twice
    from **inside `lint/`** and `git diff --exit-code` from the **repo root**: clean. The Block If is
    about `genmanifest` producing a diff, which it does not; the header note is required by Tasks.

### 2026-09-02 — Review pass (re-dispatch after the ruled halt)

- intent_gap: 0
- bad_spec: 0
- patch: 5: (high 0, medium 1, low 4)
- defer: 1: (high 0, medium 1, low 0)
- reject: 14
- addressed_findings:
  - `[medium]` `[patch]` **`unresolvedID` was set from a FIELD count, not a TERM count, so a
    malformed expression got NAMED.** Step (1) used `len(fields) == 1`, but the Design Notes rule is
    *"only when the expression is a single term"*. Measured at `66d445b`:
    `SPDX-License-Identifier: (MIT)` → `(unknown, "(MIT)")` and `MIT()` → `(unknown, "MIT()")`. A
    named label makes `spdx` non-empty, so BOTH gates skip arm 1 and fall to arm 3, whose message
    asserts the text *"classifies as `(MIT)`"* — false, and precisely the wrong-ground refusal the
    Design Notes call load-bearing to avoid and D-8.4j.9 ruled a defect. Fixed by deriving the
    decision from the sole enumerator. **The fix as I specified it (`len(terms) == 1`) was
    INSUFFICIENT and the patch agent proved it:** `ClassifySPDXExpressionTerms("MIT XOR Apache-2.0")`
    returns exactly ONE term with a non-nil operator error, so a bare term count newly named
    `MIT XOR Apache-2.0` — the same defect one input over. The shipped predicate is
    `len(terms) == 1 && terms[0] == expression`. New gate-level coverage at BOTH sites pins that an
    unenumerable expression is refused as *unclassifiable*, not as off-allowlist, and yields no row;
    SITE A's off-allowlist test now pins its WHOLE message as one literal (before this, forcing
    `failingTermPhrase` to always stutter reddened only the wordlist subtest).
  - `[low]` `[patch]` `firstTermNotOn` discarded the enumerator's error, so a partially-enumerated
    expression whose enumerated prefix was all-allowlisted would have been ADMITTED on a partial read
    — this story's own failure class, one function along. Unreachable today (measured:
    `MIT XOR Apache-2.0` classifies `(unknown, "")` and arm 1 catches it), but the helper is written
    and documented as a general per-term primitive. It now fails closed on `err != nil` as well as on
    zero terms, and the doc states that precondition as the contract. The single-unrecognised-id
    refusal message is byte-identical at both gates across the change (the failing term equals the
    whole label there, so `failingTermPhrase` collapses to *"which"*).
  - `[low]` `[patch]` The consumer count disagreed with itself — `classify.go` said THREE consumers,
    `classify_test.go` still said *"two consumers"*. THREE is correct since D-8.4j.9 pulled SITE B
    inside. In a file whose whole style is precision about counts, the stale count is the defect.
  - `[low]` `[patch]` `RenderAssets` emitted the new label/gate-divergence paragraph BEFORE the
    `len(rows) == 0` branch, so an empty corpus rendered a paragraph explaining a Licence column above
    *"No redistributed non-code assets are committed at this commit."* Moved inside the populated
    path; `lint/MANIFEST.md` byte-identical (the real corpus is non-empty).
  - `[low]` `[patch]` Two stated `spdxLineRE` design details were unpinned: the TAB separator that
    motivated narrowing `\s*` → `[ \t]*`, and CRLF tolerance. Only the NEGATIVE case was tested, so
    the narrowing was proved to EXCLUDE and never proved to still INCLUDE. Both rows added. **The
    patch agent reported that the CRLF half is doubly defended and no single mutation reds it**
    (`[^\n\r]*` → `.*` alone stays green because `strings.Fields` drops the stray `\r`; using raw
    `m[1]` alone stays green because the capture already excluded it) — the row is kept as a pin on
    the observable promise, and its comment records the measurement instead of repeating the claim.

#### defer (1) — written to `deferred:` this pass

- **[medium] An SPDX `WITH` exception expression stops being a signal.** `Apache-2.0 WITH
  LLVM-exception` classified `(permissive, "Apache-2.0")` at `dbd1699` — the single-token capture took
  the first term and never saw the rest — and classifies `(unknown, "")` now. The operator check that
  refuses `WITH` is UNMODIFIED by this story; what changed is that the rest-of-line capture routes
  such a declaration to it. Same family as `deferred[1]`/`deferred[2]` (the priced cost of reading the
  line whole), but a distinct shape covered by neither, and a likelier one where OFL/Apache faces are
  redistributed. Fail-closed, loud, left UNNAMED so it reaches the *"could not be classified"* arm;
  0 of 35 population texts carry it. Registered as `deferred[3]` and routed ahead of Story 8.5.

#### reject (14) — enumerated individually, each naming the caller, path or population actually
verified, and that being the one the finding is about (D-8.4j.3)

1. *"Two lines declaring the same disjunction in opposite term order should not conflict."* Verified
   at `resolveLicenceSignals`' arm 3, which this story does not modify, and against the matrix row
   that rules it (*"Two distinct compound lines → `(unknown, "")` — arm 3 across lines, unchanged"*).
   `MIT OR Apache-2.0` and `Apache-2.0 OR MIT` are distinct expressions; sorting terms to equate them
   is a second opinion about SPDX grammar the contract forbids.
2. *"`declaredTerms` is keyed on SPDX term spelling while name signals use `licenceNames`' canonical
   spelling, so a dual-licensed file whose body spells the id differently hits arm 3."* Verified at
   the path the finding is about — `ClassifyLicenceText` over a real header-plus-body pair, not the
   name table in isolation. Measured: `Apache-2.0 OR MIT` above the Apache 2.0 body →
   `(permissive, "Apache-2.0 OR MIT")`, no conflict. The one spelling divergence that exists
   (`GPL-3.0-only` vs the table's `GPL-3.0`) is on the COPYLEFT path, where a conflict is not the
   failure mode: `GPL-3.0-only OR MIT` above a GPL body → `(copyleft, "GPL-3.0-only OR MIT")`.
3. *"The wrapper-agreement test does not compare error TEXT."* Verified at the four `return`
   statements of `ClassifySPDXExpressionTerms`: every `fmt.Errorf` format string is character-identical
   to `dbd1699`'s. The Block If constrains resolution SEMANTICS, and the census is its named witness
   (35/35 unmoved).
4. *"The cited decision records and the deferred finding are dangling — nothing in the diff produces
   them."* Verified against the repository, which is the population the claim is about, rather than
   against the diff the reviewer was handed (`lint/` only): D-8.4j.1/.9/.10 are in
   `_bmad-output/implementation-artifacts/epic-8-15-decision-log.md` and the costs are registered in
   this spec's own `deferred:` frontmatter.
5. *"Blast radius: something downstream may parse the Licence column."* Verified at every non-test
   consumer of the label, not at the column in the abstract: `rules/licencegraph.go:44` switches on
   FAMILY only; `manifest.go:81` publishes the string verbatim into prose; no machine consumer of
   `MANIFEST.md` exists in the repo. All 9 dependency verdicts pinned unchanged by the census.
6. *"`firstTermNotOn`'s name and return polarity disagree (a double negative at the call sites)."*
   Style. Verified at both call sites: `case !everyTermAllowed` / `case !everyTermPermissive` are
   correct and sit in the arm order the matrix requires.
7. *"'four-identifier allowlist' is hardcoded prose in three places and nothing fails if the list
   changes size."* Verified at `manifest_test.go`'s `TestFontAssetLicenceAllowlistIsTheOwnersFourIds`,
   which pins the population the finding is about — the font allowlist — to exactly the owner's four
   ids, so the count cannot drift silently.
8. *"The manifest tests hardcode allowlist-membership assumptions."* Same verified pin as 7, plus
   `TestTheTwoAssetSitesDoNotShareAPolicy`. A list revision reds those first and by name.
9. *"`terms` is `nil` on some error paths and empty-non-nil on others."* Verified at the only
   consumer, `firstTermNotOn`, which tests `len(terms) == 0` exclusively — the distinction is
   unobservable.
10. *"Subtest names include an empty string and punctuation-mangled forms."* Cosmetic; `t.Run("")`
    yields `#00`, which is greppable and affects no assertion.
11. *"`ClassifySPDXExpression`'s doc names a file, not a caller, and 'both callers' is unenforced."*
    Cosmetic. Verified: the wrapper's callers are `rules.ScanLicenceGraph` and the npm scan, which is
    what the comment describes; the count claim it shares with the sibling is corrected by this pass.
12. *"`MANIFEST.md`'s bytes moved, so 'regenerates with NO diff' is broken."* Verified by running the
    Block If's ACTUAL condition rather than reading the byte count: `go run ./cmd/genmanifest` twice
    from inside `lint/`, then `git diff --exit-code -- lint/MANIFEST.md` from the repo root — exit 0
    both times. The Block If names `genmanifest` PRODUCING a diff; the header note itself is mandated
    by Tasks & Acceptance.
13. *"The Block If's admission-delta argument does not reach the separator narrowing `\s*` →
    `[ \t]*`."* Verified against `deferred[2]`, which registered exactly that consequence at the plan
    gate with the same measurement, and against the census: 0 of 35 texts carry the split-line form,
    so no licence's admission status changes.
14. *"Lowercase `spdx-license-identifier: mit` is asymmetric and untested."* Verified at `spdxLineRE`'s
    `(?i)` flag and at `classifyBySPDX`'s exact map, BOTH unmodified by this story: a lowercase
    identifier was unrecognised at `dbd1699` and is unrecognised now. Pre-existing, unchanged, 0
    population texts affected.

## Auto Run Result

Status: done
Blocking condition: none

Story 8.4j is implemented, reviewed, patched and committed on `main`. This dispatch resumed the
spec after D-8.4j.9…D-8.4j.13 discharged the `intent gap` halt, re-applied
`8-4j-attempted-implementation.patch` rather than re-deriving it, and folded the rulings in.

### Summary of the change

`spdxLineRE`'s capture widens from a single token to the rest of the line and is routed through the
classifier that has resolved compound expressions correctly since Story 1.3, and asset admission
becomes **per-term at BOTH gates**: a compound is admitted iff every term is on the list governing
that gate's population — the owner's four ids at the font site, `permissiveSPDX` at the wordlist
site. The gates share the MECHANISM and keep their POLICIES (D-8.4j.9); D-8.5.13's
two-populations-two-policies ruling is preserved and the test that pins it still reds if the lists
are collapsed.

### Files changed

- `lint/internal/licence/classify.go` — rest-of-line capture with its three mandated regex details;
  `ClassifySPDXExpression` becomes a thin wrapper over a new `ClassifySPDXExpressionTerms`, the SOLE
  term enumerator. Resolution semantics unchanged.
- `lint/internal/licence/licencesignals.go` — step (1) routes each line's normalized capture through
  the sibling; one line is one signal whatever its arity; a resolved expression's terms are recorded
  in `declaredTerms`, a set consulted by the NAME and CLAUSE steps only (D-8.4j.10), never the shared
  `seen` map that dedups SPDX lines.
- `lint/internal/manifest/manifest.go` — `firstTermNotOn` is the shared mechanism with the list as a
  PARAMETER; SITE A passes the four-id allowlist, SITE B passes `licence.IsPermissiveSPDX`.
  `failingTermPhrase` names the failing term only when it differs from the label.
  `RenderAssets` carries the label/gate-divergence note.
- `lint/MANIFEST.md` — the same note in the assets header (generated, not hand-edited: `genmanifest`
  run twice leaves `git diff --exit-code` clean).
- `lint/internal/licence/classify_test.go`, `lint/internal/manifest/manifest_test.go` — the six
  red-proofs, the matrix controls, and the review pass's added coverage.

### Review findings breakdown

5 patched (1 medium, 4 low), 1 deferred (medium, `deferred[3]`), 14 rejected. No intent_gap, no
bad_spec, `review_loop_iteration` unchanged at 0.

**Follow-up review recommended: TRUE.** Patched severities: high 0, medium 1, low 4 →
`3 × 1 + 1 × 4 = 7`, which is ≥ 5.

### Verification — every invocation with its WORKING DIRECTORY (D-8.4j.8), re-run AFTER the patches

| working directory | command | result |
|---|---|---|
| `lint` | `go test -count=1 ./...` | exit 0 — 4 packages `ok` |
| `lint` | `go vet ./...` | exit 0 |
| `lint` | `go test -count=1 -run TestLicenceSignalCensus -v ./internal/licence/` | PASS — *35 licence texts measured (26 committed files + 9 dependency licences), all matching their pinned verdicts* |
| `folio-go` | `go vet ./...` | exit 0 |
| `folio-go` | `go test -count=1 ./...` | **1815 pass / 2 fail / 5 skip** — the two are `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`, the mandated standing red |
| `folio-go` | `TestGoldenDigestAgreesAtEveryDeclaredSite` (rides the run above) | PASS — *23 artifacts re-hashed; 70 declared recording sites agree* |
| `lint` | `go run ./cmd/genmanifest` ×2 | exit 0, exit 0 |
| **repo root** | `git diff --exit-code -- lint/MANIFEST.md` and `git diff --exit-code` | exit 0 — `genmanifest` idempotent |
| **repo root** | `gofmt -l folio-go lint` | exactly `lint/internal/rules/licencegraph_test.go` (DW-116, pre-existing, NOT reformatted) |
| **repo root** | `md5 -q README.md` | `078d7d80d518d54af2fc04fb270d46b8` |
| `folio-designer` | `npm run test` | **40 files / 411 tests** passed; `maximumCacheAssets = 64` confirmed inside the run by `scripts/offline-release-contract.test.mjs` |
| `folio-designer` | `npm run typecheck` / `build` / `test:e2e:compile` | exit 0 / 0 / 0 |
| `folio-designer` | `npm run lint` | exactly **4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331`) |

Cadence PER-EPIC (D-000.4): the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and
Playwright are excluded by design — Epic 8's boundary gate.

### The six red-proofs, each mutation-proved by DELETION, RE-RUN after the patches

Run as `go test -count=1 ./...` from `lint/`. **The four pre-existing proofs are proofs over the OLD
patch; they were re-run after the rulings were folded in, and again after the review patches, because
the patches touched the mutation targets (D-8.4j.11 Correction 3).**

| deletion | reds |
|---|---|
| half 1 — restore the single-token capture | ALL ELEVEN 8.4j tests, each on its own message; **zero pre-existing** |
| the compound-copyleft path (RP1/RP2) | RP1, RP2 and the gate's `copyleft_term_first`/`copyleft_term_second` |
| SITE A's per-term admission — restore the exact-map lookup (half 1 alone) | RP3's gate test, on the predicted FALSE REFUSAL of `OFL-1.1 OR Apache-2.0`, plus RP4; zero pre-existing |
| the failing-term naming in the message | **exactly one** — RP4 |
| SITE B's per-term admission — restore the bare `IsPermissiveSPDX` lookup | **exactly one** — RP5, both subtests |
| `markDeclaredTerms` into the SHARED `seen` map (D-8.4j.10's own defect) | **exactly one** — RP6 |
| `markDeclaredTerms`' call sites | **exactly one** subtest — *a compound line over the body of one of its own terms is not a conflict* |

**RP1 and RP2 do not have a guard separable from half 1.** Deleting the copyleft arm they share reds
pre-existing single-identifier copyleft tests too, because that arm serves single ids as well. Their
isolating deletion is half 1 itself, which reds zero pre-existing tests — which is also
D-8.4j.13's second record: a defect no existing test could express was unobserved, not absent.

**The masking falsifier was RE-MEASURED, not carried (D-8.4j.11 Correction 2), twice** — after the
fold and again after the patches: a `MIT OR Apache-2.0` header over a GPL body still refuses
`(copyleft, "GPL-3.0")`.

### Admission parity — the measurable claim D-8.4j.9 made

**No licence's admission status changes at either gate.** Verified three ways: the census reports 35
of 35 verdicts unmoved; `genmanifest` run twice produces no diff, so every published row is unchanged;
and a probe confirms every single-identifier declaration classifies exactly as it did at `dbd1699`
(the failing term equals the whole label, so `failingTermPhrase` collapses to *"which"* and the
refusal message is byte-identical). Admission moves only for expressions carrying a term outside the
governing list — the bypass class that is this story's subject.

### What the census can and cannot witness (D-8.4i.11)

`pinnedCensus` stays a 35-entry hand-written literal derived from nothing under test, and no row was
added. **Its 0-of-35 result is evidence that this story BREAKS NOTHING, and is not evidence that it
DOES anything** — no population text carries a compound line, so corrupting the new expression branch
leaves the census green. The witnesses that the fix works are the six red-proofs and their
independent deletion probes.

### Residual risks

- `deferred[1]` (trailing content after an identifier), `deferred[2]` (an identifier on the following
  line) and the newly registered `deferred[3]` (`WITH` exception expressions) ship as specified —
  fail-closed, loud, left unnamed so they reach the *"could not be classified"* arm, and unreachable
  by the present 35-text population. All three become reachable at Story 8.5.
- `deferred[0]` is discharged: both sites are fixed, against different lists.
- `followup_review_recommended` is `true` and is not discharged here.
- The implementation subagent created commit `66d445b` itself — **instance six** of D-8.5.9. Audited
  under D-8.4j.12's standing close step: `main`, both trailers exact and present, author and committer
  consistent, only this story's files, nothing pushed (`origin/main` is still `c985b9c`).

## Delivery Log

### 2026-09-02 — done

**Story 8.4j — delivered on `main`, baseline `b4dabd9`. Two code commits, one halt, one revert.**

| commit | subject |
|---|---|
| `1af9854` | Read a compound licence line whole, and admit a font term by term — **HALTED, then reverted** |
| `d21684a` | Revert 8.4j's implementation: the same false refusal, one gate over |
| `66d445b` | Read a compound licence line whole, and admit a term at a time at both gates |
| `24048c6` | Name a malformed SPDX expression nothing, and fail closed on a partial read |
| `ea1f4f9` | Close 8.4j's gates, and record that my own patch instruction was wrong (decision log only) |
| *(this commit)* | Close Story 8.4j: the record, the register, and the proof that RP1/RP2 do separate |

**The halt is part of the delivery, not a detour.** The first dispatch shipped half 1 (read the line
whole) plus per-term admission at the FONT site only, and stopped `blocked` / `intent gap`: half 1 makes
the label the whole expression, so at the WORDLIST site — left on a bare exact-id lookup — a
wholly-permissive `CC0-1.0 OR MIT` was newly **refused**, with a message asserting the project does not
recognise it as permissive while the same call held `FamilyPermissive` in scope. A fix that trades a
bypass for a false refusal one gate over is not an incomplete fix; it is a wrong one. The build did not
guess. It halted, the lead ruled (D-8.4j.9: policy and mechanism are different things — the two sites
keep their DIFFERENT lists and share ONE way of testing a label against a list), and the work was
**reverted to zero net code change** by `git revert`, not by rewriting history, with the attempted
implementation preserved as `8-4j-attempted-implementation.patch`. The re-dispatch re-applied that patch
with the rulings folded in rather than re-deriving it. **All four pre-existing mutation proofs were
re-run afterwards** — a proof over the old patch is not a proof over the new one — and again after the
review patches, because those patches touched the mutation targets.

**What shipped.** The SPDX-line capture widens from one token to the rest of the line and routes through
`ClassifySPDXExpressionTerms`, the sole term enumerator, whose resolution semantics are unchanged since
Story 1.3. One line is one signal whatever its arity; composition across lines stays
`resolveLicenceSignals`' job. Both asset gates admit **per term**, each against its own list — the
owner's four ids at the font site, `permissiveSPDX` at the wordlist site — through one shared helper
that takes the list as a parameter. **Two lists, one mechanism, one enumerator.** Grepped at the close:
no `strings.Split` or re-split of an expression anywhere in admission, seen-marking or either gate.

**Decisions applied by ID:** D-8.4j.1 (the story exists and is bounded at one defect), .2 (no
compound-expression ban, no 8.5 census guard here), .3 (every rejection names the path it verified), .4
(DW-128 stays out), .5, .6 (font admission inside), .7 (the bound's criterion), .8 (every measurement
carries its working directory), .9 (SITE B inside; policy ≠ mechanism), .10 (Approach and Block If
amended; seen-marking scoped to the name/body space), .11 (the masking falsifier re-measured, not
carried), .12 (the provenance audit is a listed step), .13, .14 (an orchestrator's implementation sketch
is a hypothesis — `len(terms) == 1` alone was insufficient), .15 (instance six), .16 (follow-up review
discharged into this close), .17 (two volunteered disclosures). Also D-8.4i.9 (the BSD clause path is
untouched), D-8.4i.11 (what the census can witness), D-8.5.13 (two populations, two policies),
D-000.4 (per-epic cadence).

**Triage: 5 patched / 1 deferred / 14 rejected**, over the re-dispatch's pass (the halted pass's own
1 intent_gap / 2 bad_spec / 2 patch / 4 defer / 12 reject are recorded above and were superseded by the
ruling). Patched: the `unresolvedID` field-count-not-term-count defect that NAMED `(MIT)` and pushed both
gates off their "could not be classified" arm onto one asserting a false classification (medium,
corroborated three independent ways); `firstTermNotOn` discarding the enumerator's error, which would
have admitted on a partially-enumerated prefix — this story's own failure class one function along; the
consumer count disagreeing with itself (THREE, not two, since SITE B came inside); the label/gate
divergence paragraph rendering above an empty-corpus manifest; and two unpinned `spdxLineRE` details
(TAB separator, CRLF). Deferred: `deferred[3]`, `WITH` exception expressions. Rejected: 14, each naming
the caller, path or population it verified.

**What the review caught that the build would not have.** The naming defect is the one that matters: the
layers found it and triage KEPT it. Across three stories the weak link has been triage, not generation
(D-8.4j.16) — at 8.4i the same layers found the compound-line defect and triage discarded it on a
wrong-caller premise, which is why this story exists at all.

**`followup_review_recommended: true` is DISCHARGED without a second review dispatch, per D-8.4j.16.**
The profile — zero high, zero `intent_gap`, zero `bad_spec`, all 14 rejections enumerated, the one
medium corroborated three ways before it was patched — is not a review that needs repeating. The
scrutiny was spent on the close instead, as a directed adversarial pass.

**Gates, measured at the close, each with its working directory (D-8.4j.8).**

| working directory | command | measured |
|---|---|---|
| `lint` | `go test -count=1 ./...` | exit 0 — `cmd/genmanifest`, `internal/licence`, `internal/manifest`, `internal/rules` all `ok` |
| `lint` | `go vet ./...` | exit 0 |
| `lint` | `go test -count=1 -run TestLicenceSignalCensus -v ./internal/licence/` | PASS — *35 licence texts measured (26 committed files + 9 dependency licences), all matching their pinned verdicts* |
| `folio-go` | `go test -count=1 ./...` | **1815 pass / 2 fail / 5 skip**, counted from `-json`. The two are `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` — the standing red, by identity |
| `folio-go` | `go vet ./...` | exit 0 |
| `folio-go` | `TestGoldenDigestAgreesAtEveryDeclaredSite` (rides the run above) | PASS |
| `lint` | `go run ./cmd/genmanifest` ×2 | exit 0, exit 0 |
| **repo root** | `git diff --exit-code -- lint/MANIFEST.md`, then `git diff --exit-code` | exit 0 after each run — idempotent, whole tree clean |
| **repo root** | `gofmt -l folio-go lint` | exactly `lint/internal/rules/licencegraph_test.go` (DW-116, pre-existing, NOT reformatted) |
| **repo root** | `md5 -q README.md` | `078d7d80d518d54af2fc04fb270d46b8` |
| **repo root**, vs a worktree at `b4dabd9` | `shasum -a 256 fixtures/*/expected.pdf` | 23 digests, `diff` exit 0 |
| `folio-designer` | `npm run test` | **40 files / 411 tests** passed; `maximumCacheAssets = 64` confirmed inside the run by `scripts/offline-release-contract.test.mjs` |
| `folio-designer` | `npm run typecheck` / `build` / `test:e2e:compile` | exit 0 / 0 / 0 |
| `folio-designer` | `npm run lint` | exactly **4** `only-export-components` warnings (`preview/pdf-viewer.tsx:16,17`; `App.tsx:1324,1331`) |

**"No licence's admission status changed" — re-derived independently at the close, and not from the
build's three measurements.** A throwaway in-package probe drove the REAL `ResolveAssets` and
`resolveWordlistAssetRow` — a scratch git repo per input, not a replication of their arm logic — over all
**26 committed population licence texts at BOTH gates**, in this tree and in a detached worktree at
`b4dabd9`. **78 output lines each, byte-identical, `diff` exit 0.** Non-vacuity control, run in the same
harness: the red-proof inputs DO diverge across the two revisions — `OFL-1.1 OR GPL-3.0-only` is
`ADMITTED label="OFL-1.1"` at `b4dabd9` and refused as copyleft here, which is the bypass itself — so the
harness demonstrably discriminates and the zero is a measurement, not a no-op. The 9 dependency licences
do not reach either asset gate (`licencegraph.go` switches on FAMILY only) and are pinned by the census.

**The census is reported as breaks-nothing, NOT as does-something (D-8.4i.11).** No population text
carries a compound line, so corrupting the new expression branch leaves the census green. The witnesses
that the fix WORKS are the red-proofs and their mutation screen.

**Mutation screen, re-run at the close over the whole `lint` module — these are the closer's own
integers, not the build's.** `go test -count=1 -json ./...` from `lint/` after each mutation, reverted
between runs, `git status --porcelain` empty at the end.

| mutation | top-level tests red | pre-existing red |
|---|---|---|
| restore the single-token capture (half 1) | **11**, 33 leaf failures | **0** |
| condition the copyleft arm on `len(terms) == 1` | **5** — RP1, RP2, order-independence, composition, and the gate's `copyleft_term_first`/`copyleft_term_second` | **0** |
| restore SITE A's exact-map lookup | **1** — RP3's gate test, on the predicted FALSE REFUSAL (3 subtests) | 0 |
| restore SITE B's bare `IsPermissiveSPDX` lookup | **1** — RP5 (2 subtests) | 0 |
| make `failingTermPhrase` always say "which" | **1** — RP4 | 0 |
| write `markDeclaredTerms` into the SHARED `seen` map | **1** — the second-SPDX-line test (D-8.4j.10's own defect) | 0 |
| drop `markDeclaredTerms`' permissive call site | **1** — the compound-over-its-own-body case | 0 |
| drop the `terms[0] == expression` conjunct | **3** — the D-8.4j.14 patch has teeth | 0 |
| let `firstTermNotOn` ignore the enumerator's error | **1** — `TestFirstTermNotOnFailsClosedOnAnIncompleteEnumeration` | 0 |

**D-8.4j.17's first disclosure is over-modest, and the proof CAN be tightened.** The build reported that
RP1 and RP2 have no guard separable from half 1, because deleting the copyleft arm they share also reds
pre-existing single-identifier copyleft tests. That is true **of a deletion**, because step (1) now
handles single and compound SPDX lines through the same arm. It is not a structural limit: conditioning
that arm on `len(terms) == 1` — compound copyleft treated as permissive, single-id copyleft untouched —
isolates RP1/RP2 exactly. Measured: five tests red, all five introduced by this story (none exists at
`b4dabd9`), **zero pre-existing**. The honest disclosure was worth more than an overstated claim, and it
is now replaced by a measurement rather than by an argument.

**Deferrals, with owners.** DW-131 is CLOSED by this story. `deferred[3]` — `Apache-2.0 WITH
LLVM-exception` — is registered as **DW-132, owner OWNER**: the `WITH` operator refusal is untouched by
this story and only newly reachable, and teaching the parser SPDX exceptions is a grammar-coverage
question D-8.4j.2 forbids answering in a build. The closer additionally registered **DW-133, owner
ENGINEERING LEAD**: two residues of reading the line whole that the review saw and named but that reached
no register — an unsupported operator SPELLING (`MIT or Apache-2.0` was ADMITTED at `b4dabd9`, measured,
and is refused now) and the fact that `deferred[1]`/`[2]`/`[3]`'s shapes also reach the **dependency**
gate, where they become a build-failing "licence unresolvable" for a future dependency whose LICENSE is
ordinary. Both are fail-closed and loud, both are zero-of-35 today, and neither is this story's to fix.
`deferred[0]` is confirmed DISCHARGED: it registered the TWO-SITE shape, and both sites are fixed against
different lists — verified at the close by driving both gates, not by reading the entry.

**One deferral the build registered is in fact CLOSED, and the record should say so.** The halted pass
deferred a medium: an id on the font allowlist that the classifier cannot place would clear arms 1 and 2
and be ADMITTED. The re-dispatch's `firstTermNotOn` error patch closed it. Probed at the close by adding
a synthetic unrecognised id to the allowlist: the declaration is REFUSED at both gates. That fail-open no
longer exists and needs no register entry.

**Heavy tests: WRITTEN AND COMPILING, NOT RUN.** Cadence PER-EPIC (D-000.4). The four
`FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and Playwright are **due at Epic 8's boundary
gate**. `npm run test:e2e:compile` passes, so the Playwright suite compiles; the matrix legs sit behind
their env-var/build-tag gates and were not exercised. This story changes only `lint/`, touches no
`.folio`, no engine and nothing under `folio-go/fonts/`, and the 23 golden digests are byte-identical to
`b4dabd9`, so the deferral is low-risk — but it is a deferral, not a pass.

**Provenance audit — the standing close step (D-8.4j.12), reported by SHA.** `1af9854`, `d21684a`,
`66d445b`, `24048c6` and `ea1f4f9`: every one on `main`, only this story's files, author and committer
both `Panit Wechasil <panitw@hotmail.com>` and consistent with the surrounding history, both required
trailers present and exact. **Nothing is pushed — `origin/main` is still `c985b9c`, 33 commits behind
`HEAD`.** **`66d445b` was created by the step-03 implementation subagent: that is instance SIX of
D-8.5.9** (D-8.4j.15 corrected the count from five — the count is of occurrences, not of survivors, and
`1af9854` was instance five even though it was reverted). Six instances, six catches by the audit.
