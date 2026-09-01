---
title: 'Story 8.4j: A compound licence line is read whole'
type: 'bugfix'
created: '2026-09-02'
status: 'done'
baseline_revision: 'b4dabd96d7a19fe068f5272b173d0050eef2cb36'
baseline_revision_plan_gate: 'dbd16991148d658e4bcbd65c035eb031715bc173'
review_loop_iteration: 0
followup_review_recommended: false
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
---

## In plain terms (read this first if you just want the gist)

*Non-normative. The contract below governs.*

The previous story taught our licence check to read every label on a file rather than stopping at the
first one it found. That closed a real hole. But it left a smaller one open, and the smaller one has
the same shape. When a single label lists two licences at once — a file offered under either of two
terms, which is an entirely ordinary way for a typeface to be published — the check still reads only
the first word of that label and throws the rest away. So a file offered under an acceptable licence
*or* an unacceptable one is waved through on the strength of the acceptable half, and is then recorded
in our published list as though the unacceptable half were never written down.

The tell that this is reading rather than judging: swap the two licences around and the answer
changes. Put the unacceptable one first and the file is correctly refused. That asymmetry is not a
policy; it is an artefact of where the reading stops.

The tool that reads such a label correctly already exists in this project, has existed for a long
time, and is simply never asked. This story asks it.

Done looks like this: the order stops mattering, in both directions. And a file offering itself under
two *acceptable* licences is still accepted — because the fix has to remain a classifier, not a
blanket refusal of anything that lists more than one name.

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

## Auto Run Result

Status: done
Blocking condition: none

The reverted implementation at `1af9854` was **re-applied from
`8-4j-attempted-implementation.patch`** — not re-written — and the two build-gate rulings were folded
in on top of it, exactly as the previous Auto Run Result asked.

**What the rulings changed on top of the patch:**

- **D-8.4j.9 — SITE B is per-term too, against its OWN list.** `resolveWordlistAssetRow`'s arm 3 was a
  bare `licence.IsPermissiveSPDX(wordlistSPDX)` lookup; it now consults the same term enumeration
  SITE A does, with `licence.IsPermissiveSPDX` passed as the predicate. The shared piece is the
  MECHANISM (`firstTermNotOn`, one function, list supplied by the caller); the POLICIES stay apart —
  SITE A passes the owner's four ids, SITE B passes `permissiveSPDX` — and
  `TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist` still reds if anyone collapses them.
  `manifest.go`'s SITE A helper was renamed from `firstDisallowedFontLicenceTerm` accordingly.
- **D-8.4j.10 — the seen-marking is scoped to the NAME/BODY signal space and marks TERMS, not
  fields.** `markExpressionFieldsSeen(seen, fields)` became `markDeclaredTerms(declaredTerms, terms)`:
  its own map, consulted by steps (2) and (3) only, never by step (1)'s SPDX-line dedup, and fed from
  `ClassifySPDXExpressionTerms` so operators are never marked as ids.
- **The two `patch` findings from the previous review.** `failingTermPhrase` names the failing term
  only when it differs from the whole label, so a single-identifier refusal is byte-identical at both
  gates to the message it carried before this story; and `ClassifySPDXExpressionTerms`' doc comment
  now states its consumer count correctly (three: `collectLicenceSignals`, SITE A, SITE B) and the
  wrapper's own caller count (two: `rules.ScanLicenceGraph`, `npm.go`).

**New red-proofs added on top of the patch's four:** `TestResolveWordlistAssetRowAdmitsACompound`
`PermissiveDeclaration` (RP5, at SITE B) and
`TestASecondSPDXLineIsNeverSwallowedByAnEarlierExpressionsTerms` (RP6, both orders asserted).

**Verification performed** (each command with its working directory):

- `lint` (`/Users/panitw/Projects/folio/lint`): `go test -count=1 ./...` → 4 packages `ok`;
  `go vet ./...` clean.
- census (`lint`): `go test -count=1 -run TestLicenceSignalCensus -v ./internal/licence/` →
  `CENSUS: 35 licence texts measured (26 committed files + 9 dependency licences), all matching their
  pinned verdicts`, PASS. **0 verdicts moved**, agreeing with the spec's plan-gate record. Seven of
  the nine dependency licences still classify `(permissive, "BSD-3-Clause")` through the clause path
  (D-8.4i.9 intact).
- `folio-go`: `go vet ./...` clean; `go test ./...` → **1815 pass / 2 fail / 5 skip**, the two fails
  being exactly `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)` (the mandated standing red).
- goldens (`folio-go`): `TestGoldenDigestAgreesAtEveryDeclaredSite` → *23 artifacts re-hashed; 70
  declared recording sites agree*, PASS.
- `genmanifest` (`lint`, run twice): `lint/MANIFEST.md` md5 stable at
  `239718814444b945463ad776deba643b` across runs; the only diff against `b4dabd9` is the
  label/gate-divergence note Tasks mandates for the header.
- `gofmt -l folio-go lint` from the repo root → exactly `lint/internal/rules/licencegraph_test.go`
  (DW-116, pre-existing).
- `md5 -q README.md` from the repo root → `078d7d80d518d54af2fc04fb270d46b8`.
- designer: `npm run test` → **40 files / 411 tests** passed; `npm run typecheck`, `npm run build`,
  `npm run test:e2e:compile` all exit 0; `npm run lint` → exactly **4** `only-export-components`
  warnings. `maximumCacheAssets = 64` (`src/release-payload.ts:33`) confirmed by
  `scripts/offline-release-contract.test.mjs` (12 tests) inside that run.

**Mutation proofs, each by DELETION, run as `go test -count=1 ./...` from `lint/`. No pre-existing
test reds in ANY of them:**

| mutation | reds |
|---|---|
| restore the single-token capture in `spdxLineRE` | all eight 8.4j tests, each on its own message |
| delete `markDeclaredTerms`' call sites | **exactly one** subtest — *a compound line over the body of one of its own terms is not a conflict* |
| mark terms into the SHARED `seen` map (D-8.4j.10's own defect) | **exactly one** — RP6 |
| SITE A back to the exact-map lookup (half 1 alone) | RP3's gate test, on the predicted false refusal of `OFL-1.1 OR Apache-2.0`, and RP4 |
| delete only the failing-term naming | **exactly one** — RP4 |
| SITE B back to the exact-map lookup | **exactly one** — RP5 |
| delete SITE A's per-term arm entirely | RP4 plus the pre-existing `TestResolveAssetsRefusesAPermissiveButOffAllowlistFontLicence`, which shares the arm |

**Residual risks:** `deferred[1]` (trailing content after an identifier) and `deferred[2]` (an
identifier on the following line) ship as specified — fail-closed, loud, and unreachable by the
present population. `deferred[0]` is discharged: both sites are fixed, against different lists.
