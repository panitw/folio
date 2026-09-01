---
title: 'Story 8.4i: The classifier stops lying to the gate'
type: 'bugfix'
created: '2026-09-02'
status: 'done'
baseline_revision: '582a01aea6bf22aef945a02a4e5d46966be1fe26'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: ['oversized', 'multiple-goals']
deferred:
  - summary: >-
      The excluded-path font tripwire keys on four sfnt magics and not on wOFF or wOF2, so a web font
      committed under an excluded path is still redistributed with no manifest row and nothing said.
    evidence: |-
      TestNoRealFontHidesUnderAnExcludedPath documents the gap in its own comment and calls
      looksLikeSfnt (lint/internal/rules/fontsassets.go:144), which checks 00 01 00 00, "true",
      "ttcf" and "OTTO" only. The designer's TypeScript list covers six magics including the two web
      formats. Every other gap found in this story received a DW number; this one is recorded only in
      a test comment, which is the pattern D-8.0.1 exists to stop.
    location: >-
      lint/internal/rules/excludedpathfonts_test.go
    severity: medium
  - summary: >-
      DW-127 (a text naming the 2-clause BSD licence classifies as BSD-3-Clause) is routed to Epic 15
      with no test pinning today's behaviour, so Epic 15 has no red to work against.
    evidence: |-
      The story preserves the behaviour exactly and registers it, correctly declining to fix it under
      D-8.4i.6's bound. But with clauses now demoted behind names, the BSD 2-CLAUSE name entry is the
      only thing that fires for such a text, so the mislabel is no longer masked by branch order. A
      later accidental change to it would be invisible.
    location: >-
      lint/internal/licence/licencesignals.go (licenceNames)
    severity: medium
  - summary: >-
      assetWalkStructuralExclusion derives the asset walk's exclusion by matching manifest.go's source
      with a regexp, so a formatting-neutral refactor reds the tripwire with a misleading message.
    evidence: |-
      The test requires the literal text of the SkipDir condition. Swapping the conjunction order,
      extracting a helper predicate, or splitting the condition across lines makes it Fatalf with
      "the gate's walk has changed shape" when the rule has not changed at all. Deriving it via
      go/ast, or exporting a test-only accessor from manifest, would couple it to the rule rather
      than to its spelling. Deriving rather than hand-copying was the right call (DW-119); the
      derivation mechanism is what is fragile.
    location: >-
      lint/internal/rules/excludedpathfonts_test.go
    severity: low
---

# Story 8.4i: The classifier stops lying to the gate

**Epic:** 8 — A template author can choose a font, and the file carries it
**Story key:** `8-4i-the-classifier-stops-lying-to-the-gate`
**Status:** `done` — delivered 2026-09-02 on `main`; see the Delivery Log.
**Covers:** no FR/NFR. **This story does not come from `epics.md`.** **It was created at Story 8.4h's
close on 2026-09-02 by the engineering lead's rulings D-8.4i.1 … D-8.4i.6, on the `8.4x` insertion
precedent (2.3a / 2.5a / 2.6a / 3.1a / 8.4f / 8.4h).** **`epics.md` is NOT amended by this story and
must not be** — there is no clause in it this story implements, weakens or corrects. The clause it
enforces is already written there; the defect is that the classifier does not honour it.
**Primary invariant:** **AD-26** (the licence boundary), read with **D-8.5.3**'s owner decision
extending its fail-the-build posture to redistributed font assets.
**Adjacent invariants:** AD-21 (byte identity), AD-14 (located diagnostics), AD-22 (pinned toolchain).
**Governing rulings:** **D-8.4i.1 … D-8.4i.6 (the charter)** · **D-8.4h.1** (the discarded family
verdict) · **D-8.4h.5** (the `gofmt` standing red) · **D-8.5.13** (the gate must be red-proved, not
observed passing) · **D-8.5.3** (the owner's four-id allowlist) · **D-8.5.10** (never move the bar to
fit the instrument) · **D-8.5.8c** (a guard owned by its consumer is a guard the consumer can move) ·
**D-8.5.2** (not knowing must not read as fine; extension-keyed is the blind spot) · **D-2.1.3** (a
loud-miss list is a fail-safe; a silent-miss list is a rotting list) · **D-1.3.4 / D-1.3.8** (an
unidentifiable licence fails the build) · **D-8.0.1** (a comment asserting a branch is safe is why
nobody looked) · **D-8.4.33** ("could not look", never omitted) · **D-000.4** (per-epic cadence).
**Bound:** per **D-8.4i.6**, this is the **last** licence-gate insertion Epic 8 gets. Anything further
found in this area is registered and routed to Epic 15's release gate, not built as `8.4j` — the one
exception being a demonstrated live bypass of a gate the epic has declared fail-closed, which returns
to the lead to be ruled *against* the bound rather than around it.

## In plain terms (read this first if you just want the gist)

*Non-normative. This section is a plain-language orientation for a human reader; the intent contract
below is what governs the build.*

The check that decides which licence a document is under could be fooled two ways, both now closed. A
file carrying two labels was judged by whichever came first, so a document the project must refuse
could be waved through by one stray line saying otherwise. And a licence whose name was spelled
slightly differently was not merely unrecognised — it was quietly credited to a different licence, and
that wrong name reached published paperwork. Both now count as not knowing, and not knowing fails.

The notes in the code claiming these paths were safe were false from the start. One is why an earlier
review dismissed the problem. They are corrected in place, original wording kept beside the
correction, because the note is what stopped anyone looking.

The most useful lesson is the part that went wrong. Before anything was made to refuse, a tool was
built to run the new rule over every licence the project relies on and prove nothing changed. Once the
new rule shipped, that tool was comparing the rule against itself. It kept reporting agreement, and
would have done so if every licence here had been misread. It was caught only by deliberately breaking
the classifier to see if the tool noticed — and it did not.

Two gaps are left deliberately, both registered with owners: a web-font format the new tripwire cannot
yet see, and a known mislabel kept on purpose. A third, found while closing, is sent back for a ruling.

<intent-contract>

## Intent

**Problem:** Story 8.4h made the asset licence gate fail closed, but `ClassifyLicenceText` — the
function the gate trusts — mis-reports in two ways. It returns on the **first**
`SPDX-License-Identifier:` line anywhere in the text, so a full GNU GPL v3 text carrying a stray
`SPDX-License-Identifier: MIT` line classifies `(permissive, MIT)` and passes the fail-closed gate
(DW-125), defeating 8.4h's central claim that copyleft is refused by name. Separately, the MIT case is
a **greedy catch-all** keyed on `"PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"` — a clause MIT, OFL
and UFL all share — so no near-miss on a licence *name* ever reaches `FamilyUnknown`; it lands in MIT
and is published under the wrong label in `lint/MANIFEST.md`, a release artifact under AD-26 (DW-124).
Two further gaps ride the same surface: the owner's four-id allowlist is pinned by nothing (DW-120),
and the gate's `*/testdata/lint` exclusion is defended only by the fact that nobody has yet committed a
real font there (DW-123).

**Approach:** Make `ClassifyLicenceText` **collect every signal in the text and resolve them as one**,
adopting the semantics `ClassifySPDXExpression` already implements twenty lines above it — any copyleft
signal refuses as copyleft naming the identifier; two or more distinct permissive identifiers, or a
half-recognised licence name, is `FamilyUnknown` and therefore refused. Correct the code comments that
assert the old behaviour was safe, preserving the original wording verbatim. Pin the allowlist against
a test-owned literal, and add the excluded-path tripwire keyed on sfnt magic bytes. **The very first
task runs the new rule over the entire existing population in report-only mode and records the output;
only then does anything become fatal.**

## Boundaries & Constraints

**Always:**
- **Task 1 is the report-only population census and it lands FIRST.** Run the conflict/anchor rule
  over **every** dependency `LICENSE*` (the Go module graph) and **every** asset `LICENSE*`, record
  the verdicts, and change no behaviour in that step. **The gate must not become fatal in the same
  commit that first measures the population** (D-8.4i.1's Block If, D-8.4i.6's hard constraint).
- **Copyleft detection is NOT replaced by conflict detection, and it outranks it.** Resolution order
  is fixed and is not stylistic: (1) any copyleft signal → refused **as copyleft**, naming the
  identifier; (2) no copyleft but ≥2 distinct permissive identifiers → `FamilyUnknown`; (3) exactly
  one identifier → that identifier. A maintainer reading *"conflicting identifiers"* adds an SPDX
  line; one reading *"GPL detected"* removes the dependency. **Hazard indicators fail toward the
  loudest, never the most precise.**
- The name-anchor rule is **general, not a spelling patch**: a text carrying a licence **name** signal
  that does not resolve to a known `(name, version)` pair returns `FamilyUnknown`. The bare grant
  clause classifies as MIT **only** when no name signal is present at all.
- Every false comment is **corrected in place with the original preserved verbatim** — the form this
  run requires of every correction. The defect is the behaviour **and** the false claim (D-8.0.1's
  shape at instance three); fixing only the first leaves the next reviewer the same trap.
- Multi-signal resolution must be **deterministic and documented**, with a test pinning the chosen
  identifier when several copyleft signals are present. A gate whose message varies by map iteration
  order is not a gate.
- Every new refusal arm is **red-proved by deletion, not only by substitution** — falsifying a
  condition proves arm *order*; deleting it proves the arm is *reached* (D-8.4h.4's own lesson).
- **Byte identity:** the **23** golden digests stay byte-identical, `maximumCacheAssets` stays **64**,
  `README.md` md5 stays `078d7d80d518d54af2fc04fb270d46b8`, and `lint/MANIFEST.md` regenerates with
  **no diff**.
- Commit only on `main`.

**Block If:**
- **The report-only census reds anything legitimate.** That is a **HALT and a finding to route to the
  engineering lead**. It must **NOT** be resolved by weakening the rule, narrowing its population, or
  exempting a file — moving the bar to fit the instrument is the failure this whole thread exists to
  stop (D-8.5.10). *(Taken at this plan gate and clean — see Design Note 1. Re-run it at build; if the
  build's own run disagrees with the recorded one, that disagreement is itself the halt.)*
- A golden digest moves, or `lint/MANIFEST.md` regenerates with a diff. A changed manifest **label** is
  a finding to route, **not** a re-record.
- The census turns up a **demonstrated live bypass** of a gate Epic 8 has declared fail-closed. Per
  D-8.4i.6 that class — and only that class — may buy a further story, and it returns to the lead to
  be ruled against the bound rather than around it.
- An asset directory is found to hold more than one `LICENSE*` file (DW-118 becoming live).

**Never:**
- No `.folio` change, no engine change, no change under `folio-go/fonts/`.
- No edit to `epics.md`, `ARCHITECTURE-SPINE.md`, or this `<intent-contract>`.
- **Do not implement the rejected alternative** — *"a copyleft marker in the body outranks a permissive
  SPDX line."* It is an exception list: it fixes GPL-text-plus-MIT-line while leaving MIT+BSD,
  OFL+Apache and every future pair mislabelled (D-8.4i.1, option (a), rejected).
- **Do not "accept both spellings"** of the Ubuntu Font Licence. That fixes one instance of a defect
  that has now produced three (D-8.4i.2, rejected).
- **Do not reimplement a second sfnt reader** to satisfy the tripwire (D-8.4i.4).
- Do not add `CC0-1.0` to the four-id font allowlist, and do not collapse the two asset sites onto one
  list — they are independent in **scope**, not merely in mechanism (D-8.5.13).
- Do not reformat `lint/internal/rules/licencegraph_test.go` (DW-116, D-8.4h.5).
- Not in scope: catalogue faces, `@font-face` rules, picking behaviour, byte thresholds (8.4d owns),
  bold/italic (Epic 11, D-000.7). Never push, never branch, never `git add -A`.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Single SPDX id | `SPDX-License-Identifier: MIT` | `(permissive, "MIT")` — unchanged | No error expected |
| DW-125 bypass | GNU GPL v3 text + a stray `SPDX-License-Identifier: MIT` line | `(copyleft, "GPL-3.0")` — refused **as copyleft**, naming the id | Gate fails naming the directory and the copyleft id |
| Copyleft SPDX after permissive SPDX | `MIT` line then `GPL-3.0-only` line | `(copyleft, "GPL-3.0-only")` — copyleft outranks conflict | Gate fails as copyleft |
| Two permissive ids | `MIT` line and `BSD-3-Clause` line, either order | `(unknown, "")` — order-independent | Gate fails as unclassifiable |
| DW-124, American spelling | `UBUNTU FONT LICENSE Version 1.0` + the shared grant clause | `(unknown, "")` — name signal present, does not resolve | Gate fails as unclassifiable |
| OFL 1.0 (the predicted fourth instance) | `SIL OPEN FONT LICENSE Version 1.0` + the shared grant clause | `(unknown, "")` | Gate fails as unclassifiable |
| Committed OFL 1.1 text | any of the 10 committed `LICENSE-OFL.txt` | `(permissive, "OFL-1.1")` — unchanged | No error expected |
| Bare grant clause, no name | MIT body with no recognisable licence name | `(permissive, "MIT")` — the fallback survives | No error expected |
| Go-style BSD text | `REDISTRIBUTION AND USE IN SOURCE`, no name signal, no SPDX line | `(permissive, "BSD-3-Clause")` — unchanged (7 of 9 dependencies rely on this) | No error expected |
| Allowlist widened | `"GPL-3.0"` appended to `fontAssetLicenceAllowlist` | The **named** allowlist test reds on its own message | Exactly one test reds; no other test reds |
| Excluded-path font | a real sfnt program committed under `*/testdata/lint` | The tripwire reds, keyed on magic bytes | Fails naming the file |

</intent-contract>

## Code Map

- `lint/internal/licence/classify.go` (287 lines) — **the whole subject.** `ClassifyLicenceText` at
  `:165-275`; the first-match-and-return bypass is `:166-169`. Marker branches: AGPL `:173`, SSPL
  `:175`, LGPL `:177`, GPL `:179`, **OFL `:181` (requires `"VERSION 1.1"`)**, **UFL `:207` (requires
  `"VERSION 1.0"`, British `LICENCE`)**, **the greedy MIT catch-all `:251`**, Apache `:253`, BSD `:255`
  (note its third disjunct `"REDISTRIBUTION AND USE IN SOURCE"` is a *clause*, not a name), CC0 `:257`.
  `classifyBySPDX` `:277-287`; `spdxLineRE` `:157`; `permissiveSPDX` `:96-134`; `IsPermissiveSPDX`
  `:150`; `copyleftSPDXPrefixes` `:152`.
  **The template to copy is in the same file:** `ClassifySPDXExpression` `:14-40` — starts
  `FamilyPermissive`, sets `FamilyCopyleft` if any term is copyleft, returns `FamilyUnknown` if any
  term is unrecognised. That is D-8.4i.1's ruling, already written, twenty lines above.
- **The two false comments to correct in place:** `:241-249` (UFL — *"classifies FamilyUnknown … a
  LOUD miss, and therefore fail-safe (D-2.1.3)"*, *"for no measured need"*) and `:201-205` (OFL —
  *"makes a 1.0 file classify as FamilyUnknown"*). **Both are now measured false** (Design Note 2).
  The collision they describe is recorded at `:184-187` (OFL) and `:220-224` (UFL) — those two are
  **true** and are the evidence the catch-all is the root defect.
- `lint/internal/licence/classify_test.go` (285 lines) — `TestClassifyLicenceText` `:12` (18 cases,
  **family only**, `gotFamily, _ :=` at `:45`), `TestClassifyOFL` `:62`, `TestCommittedOFLTextClassifiesAsOFL11`
  `:128` (reads the real artifact), `TestClassifyUbuntuFontLicence` `:180`,
  `TestUbuntuFontLicenceSPDXLineIsPermissive` `:247`, `TestIsPermissiveSPDXReadsTheSameList` `:271`.
  ⚠ **Two shipped cases pin the behaviour this story reverses** — `TestClassifyOFL`'s bundled-notice
  case (`:~100`, MIT file bundling an OFL notice → expects `MIT`) and `TestClassifyUbuntuFontLicence`'s
  equivalent. They must be **re-pointed to `FamilyUnknown` loudly, with the reversal recorded**, not
  silently edited. This is a guard being **replaced by a stricter one**, which the epic permits; a
  weakening would not be.
  ⚠ **`ClassifySPDXExpression` has NO unit test at all** — its operator/parity/`()` logic at `:16-38`
  is exercised only end-to-end through the npm lockfile, whose 248 packages are all bare identifiers.
  Adopting its semantics is a good moment to give it one; not required by the charter.
- `lint/internal/manifest/manifest.go` — **SITE A, the font path**, `:357-365`: `family, spdx :=
  ClassifyLicenceText(...)` then a three-arm switch, `spdx == ""` (`:359`) / `family == FamilyCopyleft`
  (`:361`) / `!fontAssetLicenceAllowed[spdx]` (`:363`). **The arm order is load-bearing** — copyleft
  sits above the allowlist arm, which is exactly why appending `"GPL-3.0"` reddens nothing.
  `fontAssetLicenceAllowlist` `:140`; `fontAssetLicenceAllowed` `:142-148`. The walk's exclusions are
  `:211-219` (`.git` at `:212`; a directory literally named `lint` whose immediate parent is literally
  named `testdata` at `:215` — **structural, not glob**), documented at `:196-202`. **SITE B**, the
  wordlist, uses `licence.IsPermissiveSPDX` — do not route the font path through it.
- `lint/internal/manifest/manifest_test.go` — every allowlist reference is a read **into**
  `fontAssetLicenceAllowed[<literal>]` or a printf arg; **nothing pins membership**. Tests that touch
  it: `:642` `TestTheTwoAssetSitesDoNotShareAPolicy`, `:676` `TestCommittedAssetPopulationClassifiesCleanly`,
  `:416/:446/:478/:504` the four refusal/acceptance tests, `:542` the wordlist policy test.
- `lint/internal/rules/fontsassets.go:140-159` — **`looksLikeSfnt(data []byte) bool`**, unexported,
  `package rules`, four magics (`00 01 00 00`, `true`, `ttcf`, `OTTO`; **no `wOFF`/`wOF2`**). Sole call
  site `:253`. **This is the sfnt reader the tripwire uses** — see Design Note 4.
- `lint/internal/rules/licencegraph.go:28-65` — `ScanLicenceGraph`, the **dependency** consumer of
  `ClassifyLicenceText`. It **switches on `Family` only**; `spdx` is used solely as a `%s` in the
  copyleft message (`:49`) and is discarded on the permissive and unknown paths. Population: the Go
  module graph via `licence.ResolveGraph` (`graph.go:37`, `GOPROXY=off`), licence file names closed to
  `{LICENSE, LICENSE.txt, LICENSE.md, COPYING}` (`graph.go:73`).
  ⚠ `ScanNPMGraph` (`:71-84`) uses **`ClassifySPDXExpression` over the lockfile's declared `license`
  string**, never `ClassifyLicenceText` — see Design Note 3.
- `lint/testdata/licence/{copyleft,permissive,unknown}/` — three fixture modules, every `LICENSE` a
  **one-line SPDX marker** (AC30), never full legal text. `permissive/example.test/ufl-lib/` was added
  at 8.4h.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-116 (`:5900`), DW-117 (`:5924`),
  DW-118 (`:5961`), DW-119 (`:5982`), **DW-120 (`:6003`)**, DW-121 (`:6028`), DW-122 (`:6045`),
  **DW-123 (`:6064`)**, **DW-124 (`:6083`)**, **DW-125 (`:6127`)**. **DW-117 is the same marker branch
  and folds into this story's Task 3** — its own entry says so.
- `_bmad-output/implementation-artifacts/epic-8-15-decision-log.md` — D-8.4i.1 `:1008`, D-8.4i.2
  `:1063`, D-8.4i.3 `:1115`, D-8.4i.4 `:1135`, D-8.4i.5 `:1155`, D-8.4i.6 `:1175`.

## Tasks & Acceptance

**Execution:**

1. `lint/internal/licence/` — **REPORT-ONLY CENSUS, FIRST, ALONE, NO BEHAVIOUR CHANGE.** Implement the
   signal-collection rule behind a report-only path and run it over the full population: **9** dependency
   `LICENSE` files (the Go module graph across `folio-go`, `lint`, `hashmatrix` — `hashmatrix` resolves
   **zero** non-main modules) plus **12** asset `LICENSE*` files (11 font directories + the wordlist)
   plus the **8** committed lint fixtures. Record every verdict in the Delivery Log. **Commit this
   census before any refusal becomes fatal.** If it reds anything legitimate — HALT and route.
2. `lint/internal/licence/classify.go` — replace the first-match-and-return at `:166-169` with
   **collect-all-signals**, resolved in the fixed order copyleft → conflict → single. Adopt
   `ClassifySPDXExpression`'s semantics; document the tie-break when several copyleft signals appear.
3. `lint/internal/licence/classify.go` — **anchor the name signal** (D-8.4i.2): a name signal that does
   not resolve to a known `(name, version)` pair returns `FamilyUnknown`; the bare grant clause reaches
   MIT only when no name signal is present. **Verify the BSD branch's `"REDISTRIBUTION AND USE IN
   SOURCE"` disjunct keeps behaving as a clause, not a name** — 7 of the 9 dependency licences depend
   on it (Design Note 1). This discharges **DW-117** as well as DW-124; amend both entries.
4. `lint/internal/licence/classify.go` — **correct the two false comments in place, original preserved
   verbatim**: `:241-249` (UFL) and `:201-205` (OFL). Each correction states the measurement, its date,
   and that the claim it replaces was the reason the review rejected DW-124's finding. Audit the file's
   other comments for the same class of claim while you are in it.
5. `lint/internal/licence/classify_test.go` — cases for every I/O Matrix row, plus the two **reversed**
   expectations (bundled-notice OFL and UFL → `FamilyUnknown`), each carrying a comment naming
   D-8.4i.1 and recording the old expectation. **Red-proof each new arm by deletion**, and assert
   order-independence for the two-permissive-ids case in both orders.
6. `lint/internal/manifest/manifest_test.go` — **DW-120**: pin `fontAssetLicenceAllowlist` against an
   **exact test-owned `[]string` literal naming D-8.5.3** as the authority. **Not** a derivation from
   the constant — that passes any edit. Red-proof: append a copyleft id, confirm **the named test reds
   and no other test reds**; if four packages red, the guard is not the thing catching it.
7. `lint/internal/rules/` — **DW-123's tripwire**: assert that **no file under an excluded path is a
   real font program**, keyed on **sfnt magic bytes via the existing `looksLikeSfnt`**, never on the
   extension. Derive the excluded paths from the gate's own rule rather than hand-copying it (DW-119's
   lesson). See Design Note 4 — the module boundary is real but is **not** blocking.
8. `lint/MANIFEST.md` — regenerate via `cd lint && go run ./cmd/genmanifest`, run **twice** for
   idempotency, and confirm `git diff lint/MANIFEST.md` is empty both times.
9. `_bmad-output/implementation-artifacts/deferred-work.md` — close DW-117, DW-120, DW-124, DW-125;
   amend DW-123 with the tripwire's outcome. **Register the four unenumerated rejections (Design Note
   5) as a finding in its own right**, per D-8.4i.5.

**Acceptance Criteria:**

- **AC1** — Given the committed tree at baseline, when the report-only census of task 1 runs over all
  9 dependency and 12 asset `LICENSE*` files, then every verdict is recorded in the Delivery Log, and
  the commit containing that census makes **no** refusal fatal.
- **AC2** — Given a licence text carrying both a copyleft signal and one or more permissive
  identifiers, when it is classified, then the result is `FamilyCopyleft` **naming the copyleft
  identifier** — never `FamilyUnknown` and never the permissive id — regardless of which appears first.
- **AC3** — Given a licence text carrying two or more distinct permissive identifiers and no copyleft
  signal, when it is classified, then the result is `FamilyUnknown`, and the result is **identical for
  both orderings** of the same two identifiers.
- **AC4** — Given a text carrying a licence name signal that resolves to no known `(name, version)`
  pair — American-spelled UFL 1.0, or OFL 1.0 — when it is classified, then the result is
  `FamilyUnknown`, **not** `(permissive, "MIT")`.
- **AC5** — Given a text carrying the bare grant clause and **no** licence name signal, when it is
  classified, then it still returns `(permissive, "MIT")`; and given a Go-style BSD text, it still
  returns `(permissive, "BSD-3-Clause")`.
- **AC6** — Given the two code comments at `:241-249` and `:201-205`, when the story lands, then each
  carries its original wording verbatim alongside a correction stating the measured result, and no
  comment in the file still asserts that a near-miss on a licence name reaches `FamilyUnknown` by
  virtue of a required version conjunct.
- **AC7** — Given `"GPL-3.0"` appended to `fontAssetLicenceAllowlist`, when the lint suite runs, then
  **exactly one** named test reds, on a message naming D-8.5.3, and no other test in any package reds.
- **AC8** — Given a real sfnt program placed under a path the asset walk excludes, when the tripwire
  runs, then it reds naming the file; and given the tree as committed, it passes. The tripwire keys on
  magic bytes and **no second sfnt reader is introduced**.
- **AC9** — Given the story's commits, when the goldens are hashed, then the **23** digests are
  byte-identical to baseline, `lint/MANIFEST.md` regenerates with no diff, `maximumCacheAssets` is
  **64**, and `README.md` md5 is `078d7d80d518d54af2fc04fb270d46b8`.

## Spec Change Log

## Review Triage Log

### 2026-09-02 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 9: (high 2, medium 3, low 4)
- defer: 3: (high 0, medium 2, low 1)
- reject: 24: (high 0, medium 5, low 19)
- addressed_findings:
  - `[high]` `[patch]` The report-only census had become a self-comparison: `ClassifyLicenceText`'s
    whole body is `return classifyByAllSignals(text)`, and the census put those two calls in its
    `shipped=` and `collect-all=` columns, so `changed` was 0 by construction while its doc comment
    claimed it "becomes a standing witness that the two agree on the committed population". Measured
    at the gate before patching: setting the OFL entry's `requiredVersion` to `VERSION 9.9`
    reclassifies all 11 committed OFL files to `(unknown, "")` — a total failure of the shipped asset
    gate — and the census still printed `0 change verdict` and PASSED. Patched by pinning all 35
    verdicts against a test-owned table and attributing the differential figure to `2e9365e` in all
    three places that had cited it in the present tense. Re-measured after: the same mutation now
    reds the census on 11 named files.
  - `[high]` `[patch]` The I/O Matrix's "Error Handling" column had no test at the gate surface: every
    bypass row states a classifier tuple AND a gate outcome, but all sixteen new cases asserted only
    `ClassifyLicenceText`'s return value — nothing exercised the surface DW-125 was reported at.
    Patched with `TestResolveAssetsRefusesTheDW125BypassAtTheGate`, asserting the copyleft arm's
    message, the copyleft id named, the stray MIT id absent, the directory located, and no manifest
    row.
  - `[medium]` `[patch]` `readFirstBytes` used a single `f.Read`, so a legal short read on a regular
    file could yield fewer than 4 bytes and make `looksLikeSfnt` say no — the tripwire failing OPEN,
    which is DW-123's own recorded defect shape reintroduced inside the test discharging it. Patched
    to `io.ReadFull` semantics.
  - `[medium]` `[patch]` `collectLicenceSignals`' doc ranked the three signal kinds "in decreasing
    strength" with SPDX lines strongest, which the resolver contradicts — the unresolved tier sits
    above the single-identifier arm. Measured: `SPDX-License-Identifier: MIT` plus prose naming
    `UBUNTU FONT LICENSE` returns `(unknown, "")`. Patched: the doc now states the real order and why
    that direction is the fail-safe one, and the pairing is now a pinned case.
  - `[medium]` `[patch]` `TestUbuntuFontLicenceSPDXLineIsPermissive`'s doc described a short-circuit
    and a marker switch this story deleted, and asserted Story 8.4h's AC5 "two mutations, two distinct
    reds" bound, which this story removed by routing name-table ids through `classifyBySPDX`. Patched:
    comment corrected and the lost bound recorded with its measurement rather than allowed to vanish.
  - `[low]` `[patch]` The BSD clause comment credited "the Go standard library's own licence and six of
    the eight other dependency licences"; the standard library is a module in none of the three graphs,
    and the correct seven are enumerated twelve lines above. Corrected to the measured seven of nine.
  - `[low]` `[patch]` The Delivery Log's AC9 line claimed the diff was "six files, all under
    `lint/internal/`" in a commit that also changed two records. Corrected to the true file set.
  - `[low]` `[patch]` The commit table recorded `*(this commit)*` instead of a SHA, in the story whose
    central finding is that unwritten records cannot be re-checked. Resolved to `fb92156`.
  - `[low]` `[patch]` Added a pointer from the surviving `licenceNames` / `licenceClauses` tables to
    the preserved CORRECTION blocks in `classify.go`, and documented the new coupling that every table
    `id` must also appear in `permissiveSPDX` or `copyleftSPDXPrefixes`.

**Rejections, enumerated (D-8.4i.10 — a bare `reject: N` is not acceptable on this story).** Each was
checked before being dropped; two were checked and found FALSE as claims about the code.

| # | Claim | Why rejected |
|---|---|---|
| R1 | `mitGrantClause`'s doc is false because the constant is the *font* wording, not MIT's | **Claim is FALSE.** The constant is `"PERMISSION IS HEREBY GRANTED, FREE OF CHARGE"` (licencesignals.go:135) — the genuinely shared prefix, not the font sentence. Verified by reading. |
| R2 | The two-permissive-ids refusal names neither identifier, abandoning AD-14 | The intent's own I/O Matrix specifies `(unknown, "")` for that row. Measured: the code returns exactly that. Rejected on the authority of the intent itself. |
| R3 | An unrecognised SPDX id returns a non-empty id, contradicting matrix rows specifying `(unknown, "")` | **Claim is FALSE of every matrix row.** Measured all four: two-permissive-ids, American UFL, OFL 1.0 and self-declared-plus-unresolved-name all return `(unknown, "")`. The non-empty id arises only for an unrecognised SPDX id, which is not a matrix row. |
| R4 | A compound SPDX line (`MIT OR GPL-3.0-only`) drops the copyleft half | Speculative. A compound expression is routed to `ClassifySPDXExpression`, which returns copyleft if any term is copyleft. Not demonstrated. |
| R5 | SPDX ids differing in case (`mit`) are not matched | SPDX identifiers are case-sensitive by specification; no instance in the population; failure direction is fail-safe. |
| R6 | Name matching should be title-anchored so prose like "not the MIT License" cannot signal | Speculative, no instance; failure direction is `FamilyUnknown`, which is loud. |
| R7 | A dangling symlink or permission error under the excluded dir should skip, not fatal | Fatal is the correct fail-closed direction for a compliance tripwire. |
| R8 | The outer walk should `SkipDir` on unreadable directories | Same; fail-closed is correct here. |
| R9 | The census should `t.Skip` when the module graph is unavailable | A compliance census that silently skips is precisely "not knowing must read as not fine" (D-8.5.2). Fatal is correct. |
| R10 | `git ls-files` quoting breaks on non-ASCII paths | No such path in the repository; speculative. |
| R11 | The basename regexp could admit `LICENSING.md` | Prefix matching is deliberately over-inclusive for a census; over-inclusion is the fail-safe direction. |
| R12 | The `< 20` population floor is loose and 26/35 are pinned nowhere | Superseded — P1's pinned verdict table now pins the whole population by name. |
| R13 | `TestCopyleftTieBreakIsDeterministic`'s 32-iteration loop proves nothing, since no map is in the resolution path | The test does pin the chosen identifier, which is what the intent required. The loop is redundant but harmless — churn. |
| R14 | The new tests are non-hermetic and should `t.Skip` rather than fatal | The tripwire's subject *is* `folio-go/testdata/lint`; it cannot skip its own population and remain a tripwire. |
| R15 | The ~100 lines of correction narrative in `classify.go` are attached to no declaration and belong in `doc.go` | The intent requires the corrections "in place… with the original preserved verbatim", and AC6 names those line ranges. Moving them would weaken AC6. Addressed instead by P9's pointer. |
| R16 | DW-123 carries two unrelated defects under one number and should be split | Register hygiene; the story's charter routes the tripwire through that entry's number. No code consequence. |
| R17 | DW-126 is unclosable in a code-deferral register and belongs in a process record | Register hygiene; no code consequence. |
| R18 | Collecting every SPDX line widens the false-refusal surface for future dependency texts | That is the ruled behaviour (D-8.4i.1), and the widened direction is fail-safe. Narrowing it is the rejected alternative. |
| R19 | A dependency shipping no licence file classifies unknown and "fails silently" in the census | The census is report-only by charter; the dependency scan itself fails on `FamilyUnknown`. |
| R20 | Frontmatter `deferred: []` / `context: []` were not updated | Handled by this pass's own defer mechanism, not a code finding. |
| R21 | The red-proofs live in comment prose rather than in the suite | The intent requires the proofs be *performed and recorded*, which they were; AC7 and AC8 were additionally re-performed independently at review. |
| R22 | Byte identity (AC9) is recorded in a table but checked by nothing this diff adds | AC9 is a build-gate obligation discharged by the Verification section, not a guard the story was asked to add. Re-measured independently at review. |
| R23 | "Corrected in place" is satisfied at the old file rather than beside the surviving code | Same disposition as R15; addressed by P9's pointer rather than by moving the verbatim originals. |
| R24 | DW-127 should have been fixed rather than routed to Epic 15 | Permissive→permissive and not live in the population, so not the demonstrated-live-bypass class D-8.4i.6's bound admits. Routed, and additionally deferred here so Epic 15 inherits it. |


**Review triage, 2026-09-02. Nine findings, `patch: 9`, `defer: 0`, `reject: 0`.** Every finding is
named with its disposition and the probe that shows the fix changed an outcome — DW-126, this story's
own finding, is that a bare count cannot be re-checked.

| # | Finding | Disposition | Probe |
|---|---|---|---|
| **P1** | The census was a self-comparison; its doc claimed it was a standing witness | **PATCHED** — census now pins all 35 verdicts against a test-owned table; the differential figure is attributed to `2e9365e` in all three places that cited it | `VERSION 9.9` on the OFL entry reclassifies 11 committed OFL files: census **passed** before, **reds on 11 named files** after |
| **P2** | No gate-surface test for DW-125's reported shape | **PATCHED** — `TestResolveAssetsRefusesTheDW125BypassAtTheGate` | deleting the copyleft arm reproduces DW-125's exact symptom: a clean row `{synthetic-fonts/Synthetic.ttf MIT …}` on a green build |
| **P3** | `readFirstBytes` failed open on a short read — DW-123's own shape in the test discharging it | **PATCHED** — `io.ReadFull` semantics; control asserts a full read | single `Read` over a one-byte-at-a-time source → 1 byte, `looksLikeSfnt=false`; `io.ReadFull` → 12 bytes, `true` |
| **P4** | `collectLicenceSignals`' "decreasing strength" doc contradicted the resolver | **PATCHED** — doc states the real order and why unresolved-name outranks a declaration; the pairing is now tested | moving the unresolved tier below the single-id arm makes the new case return `(permissive, "MIT")` |
| **P5** | A doc described deleted machinery, and 8.4h's AC5 two-distinct-reds bound was lost silently | **PATCHED** — comment corrected; the lost independence recorded with its measurement | deleting `permissiveSPDX["Ubuntu-font-1.0"]` reds **5** tests, not 1; deleting the `licenceNames` entry reds 2 and **not** the map-entry test — the independence is now one-way |
| **P6** | BSD clause comment credited the Go standard library, which is in no graph | **PATCHED** — corrected to the measured seven of nine, named | the census enumerates the nine; none is the standard library |
| **P7** | AC9 claimed "six files, all under `lint/internal/`" in a commit that also changed two records | **PATCHED** — true file set stated | `git diff --stat 582a01a..HEAD` |
| **P8** | The commit table said `*(this commit)*` instead of a SHA | **PATCHED** — resolved to `fb92156`, with the two triage commits added | `git log` |
| **P9** | No pointer from the surviving tables to the preserved CORRECTION blocks; the new `classifyBySPDX` coupling undocumented | **PATCHED** — both added | renaming MIT's table id to `"MIT-Expat"` makes the text classify **unknown**, exactly as the note now warns |

**Two findings worth carrying forward as facts rather than as fixes.** P2's red-proof disagrees with
itself in a useful way: narrowing SPDX collection back to the first match does **not** red the new gate
test, because the GPL **name** signal reaches the copyleft arm on its own path — so the gate test and
the collect-all rows measure different mechanisms and are not substitutes. And P5's bound is genuinely
**gone**, not restored: routing name-table ids through `classifyBySPDX` is what makes every id provably
a recognised identifier, and it costs mutation resolution. Both are written into the code rather than
only here.


## Design Notes

**1. THE REPORT-ONLY CENSUS WAS TAKEN AT THIS PLAN GATE, AND IT IS CLEAN. Measured at `7e4b2c4`, not
inferred.** D-8.4i.1's Block If reserves this for the story's first task; taking it at the gate is
strictly better, because a red here halts before any design is committed to. The proposed rule was
implemented in a throwaway probe and run over the whole population. **Nothing changes verdict:**

- **9 dependency `LICENSE` files** (Go module graph; `hashmatrix` resolves **zero** non-main modules —
  its scan is vacuous). **None contains an SPDX line at all.** `textshape` and `goldmark` classify
  `MIT` by name signal; the other **seven** (`go-cmp`, `x/mod`, `x/net`, `x/sync`, `x/sys`,
  `x/telemetry`, `x/tools`) classify `BSD-3-Clause` with **no name signal and no SPDX id** — they reach
  the BSD branch through its `"REDISTRIBUTION AND USE IN SOURCE"` **clause** disjunct.
  ⚠ **This is the load-bearing constraint on task 3:** if the name-anchor rule treats that clause as a
  *name*, seven of the nine dependency licences become `FamilyUnknown` and the dependency scan reds on
  the Go standard library's own licence. Treat it as a clause, as MIT's grant clause is treated.
- **12 asset `LICENSE*` files** across 11 font directories + the wordlist: 10 × `OFL-1.1`, 1 ×
  `Apache-2.0` (`folio-go/testdata/fonts/LICENSE-Roboto.txt`), 1 × `CC0-1.0`. Every directory carries
  **exactly one** `LICENSE*` — DW-118 is not live.
- **8 lint fixtures**: each a single SPDX marker, no conflicts.
- **Zero** files in the entire population carry two distinct identifiers, an unresolved name signal, or
  a copyleft/permissive mix.

Task 1 still runs and still commits its census — the requirement is a *recorded measurement in the
commit history*, not merely a known answer. **If the build's own run disagrees with this record, the
disagreement is the halt.**

**2. THE PREDICTED FOURTH INSTANCE HOLDS. Measured two ways, not taken on authority.** D-8.4i.2
predicted that the OFL comment's claim — a 1.0 file *"classifies FamilyUnknown … a LOUD build failure"*
— cannot hold if OFL 1.0's text carries the same grant clause. It does.

*Source:* the SPDX-published OFL-1.0 text carries the sentence *"Permission is hereby granted, free of
charge, to any person obtaining a copy of the Font Software, to use, study, copy, merge, embed, modify,
redistribute, and sell modified and unmodified copies of the Font Software, subject to the following
conditions:"* — the exact substring `:251` matches on — under the title line *"SIL OPEN FONT LICENSE
Version 1.0 - 22 November 2005"*, and contains **`"Version 1.1"` nowhere**.

*Probe, at `7e4b2c4`:*

| input | measured result | the comment claims |
|---|---|---|
| OFL 1.0 title line + grant sentence | **`(permissive, "MIT")`** | `FamilyUnknown`, a loud failure |
| UFL, American `LICENSE` + grant clause | **`(permissive, "MIT")`** | `FamilyUnknown`, fail-safe |
| UFL, American, bare synthetic (no grant clause) | `(unknown, "")` | — (this is the only input the claim is true of) |
| GPL v3 body + stray `SPDX-License-Identifier: MIT` | **`(permissive, "MIT")`** | — (DW-125 confirmed) |
| the same GPL body with the SPDX line removed | `(copyleft, "GPL-3.0")` | — (control: the marker works) |
| `MIT` line then `BSD-3-Clause` line | `(permissive, "MIT")` | — (first wins) |
| `BSD-3-Clause` line then `MIT` line | `(permissive, "BSD-3-Clause")` | — (order-dependence proved) |
| **`MIT` line then `GPL-3.0-only` line** | **`(permissive, "MIT")`** | — **a copyleft SPDX line is bypassed too, not only a copyleft marker** |

Controls, non-vacuous: the committed OFL 1.1 text → `(permissive, "OFL-1.1")`; the wordlist CC0 →
`(permissive, "CC0-1.0")`; Roboto → `(permissive, "Apache-2.0")`. And the committed OFL 1.1 text does
carry the grant clause, confirming the collision `:184-187` describes.

**So the "required conjunct makes a near-miss loud" justification is false for BOTH branches, and has
been since Story 2.2.** The last row is new and is not in any register: **DW-125's bypass works with a
copyleft *SPDX line* as well as with copyleft *marker text*.** The ruled fix covers it — collecting all
signals catches both — but a fix scoped to "marker text outranks an SPDX line" would not have, which is
a second, independent reason the rejected alternative is wrong.

**3. A PREMISE IN D-8.4i.1's GUARDRAIL IS FALSE, AND THE VERDICT SURVIVES IT.** The Block If reasons
that *"the Go module graph and the npm lockfile are not [visible], and a bundled multi-licence
`LICENSE` is common in npm."* **Measured: the npm path never reaches `ClassifyLicenceText`.**
`ScanNPMGraph` (`licencegraph.go:71-84`) resolves `folio-designer/package-lock.json` and classifies the
lockfile's **declared `license` string** through `ClassifySPDXExpression`. No npm `LICENSE` **file** is
ever read, so no bundled multi-licence npm file can reach this change at all. The population is
therefore the Go module graph (9 files) plus the assets (12), not the 248 npm packages.

Per D-8.5.10's own distinction, this is a **correction to a premise, not a movement of the bar**: the
verdict, the scope and the acceptance are unchanged, and the census requirement stands exactly as
ruled — it is simply smaller and better-characterised than the ruling feared. Recorded here rather than
silently repaired, so the next reader is not re-deriving it and does not read the spec as disagreeing
with its own charter. *(Noted for completeness: all 248 lockfile licences are bare identifiers — no
conjunction, disjunction or parenthesised expression appears in the production population, so
`ClassifySPDXExpression`'s operator logic is unexercised by data as well as by unit tests.)*

**4. THE TRIPWIRE IS BUILDABLE, AND D-8.4i.4's ESCAPE HATCH DOES NOT NEED TO BE TAKEN.** The ruling
says to key on `checkSfnt` *"(already exists, D-8.4.12)"* and to defer with the tripwire as the stated
discharge **if it is not reachable from the `lint` module**, while forbidding a second sfnt reader.

**Measured, and the premise is half wrong in a way that helps.** `checkSfnt` is at
`folio-go/internal/template/fontasset.go:249` — **unexported, in the `folio-go` module.** Two
independent barriers: capitalisation, and a module boundary (`lint`, `folio-go` and `hashmatrix` are
three separate modules; there is no `go.work` and no root `go.mod`). So `checkSfnt` itself is
**genuinely unreachable**, exactly as the ruling suspected.

**But the `lint` module already owns an sfnt reader**: `looksLikeSfnt` at
`lint/internal/rules/fontsassets.go:144`, unexported in `package rules`, checking four magics. A test
file in that same directory declaring `package rules` calls it **as-is**, with no production change and
**no second reader written** — which satisfies the prohibition rather than evading it. The ruling's
fallback is conditioned on an obstacle that does not arise; its primary verdict (*"it lands in 8.4i,
beside code already being edited"*, *"a register entry keyed on nobody adding a real font is not a
defence"*) is served. **Ruled here rather than deferred, and flagged for the plan gate** — if the lead
reads the naming of `checkSfnt` as binding on the *identifier* rather than on the *capability*, the
deferral is cheap to reinstate now and pointless later.

Two implementation notes. `looksLikeSfnt` covers **four** magics and **not** `wOFF`/`wOF2` — narrower
than the designer's six-magic TypeScript list; state the coverage rather than implying completeness. And
the excluded paths must be **derived** from the gate's own structural rule (a directory named `lint`
whose parent is named `testdata`, `manifest.go:215`), not restated in a literal — restating it is
DW-119's defect committed a second time. Exporting `LooksLikeSfnt` so the test can live beside
`manifest.go` is an acceptable alternative; it is still one reader.

*Baseline for AC8:* `git ls-files` lists **66** tracked files under `folio-go/testdata/lint/`; exactly
three carry a font extension (`embed-font/{allembed,dirembed,globembed}/fonts/shipped-face.ttf`) and all
three begin `6e 6f 74 20` — ASCII `"not a real font; extension-matching fixt…"`. Reading the first four
bytes of **all 66** yields zero font magics. So the tripwire passes today and is not vacuous: the same
probe over the committed `public/fonts/*-Regular.ttf` returns `00010000`.

**5. STORY 8.4h's FOUR UNENUMERATED REJECTIONS — RECOVERED AS FAR AS THE RECORD ALLOWS (D-8.4i.5).**
This is reading, not reviewing. 8.4h's record states `reject: 7` in three places (`:553`, `:824`,
`:937`) and **describes only three**. The recovery:

| # | Claim | Disposition |
|---|---|---|
| 1 | The duplicated case in `TestClassifyUbuntuFontLicence` | **Rejected as churn** — "redundant but correct". **Stands.** Re-checked: the duplication is real and harmless. |
| 2 | The absence of a test for the American spelling `"UBUNTU FONT LICENSE"`, rejected because *"its miss is loud — `FamilyUnknown` is a build failure at the gate — and D-2.1.3 makes a loud miss fail-safe by design, which the comment records"* | **OVERTURNED — the rejection is FALSE.** Measured: `(permissive, "MIT")`, a silent pass under the wrong label (Design Note 2). This is DW-124, and this story fixes it. The rejection was made **by quoting the code comment**, which is D-8.0.1's shape. |
| 3 | The reading that AC7 belongs in the Go lint suite rather than the designer's | **Rejected; stands** — the spec's Code Map named the designer artifact, settling it. |
| 4 | — | **COULD NOT LOOK.** The record does not say. |
| 5 | — | **COULD NOT LOOK.** The record does not say. |
| 6 | — | **COULD NOT LOOK.** The record does not say. |
| 7 | — | **COULD NOT LOOK.** The record does not say. |

**So: three of seven recovered, one of those three overturned; four recorded as "could not look",
never omitted (D-8.4.33).** The only thing recoverable *about* the four is a negative: none of them is
one of the three named above. 8.4h's Review Triage Log has **no `rejected_findings` block**, no
per-finding rejection list elsewhere, and **no reviewer-layer attribution for any finding in that
story at all** — patched, deferred or rejected. `deferred-work.md` was searched too; DW-117…DW-125
account for the seven deferrals plus the two closer findings and describe none of the four.

**The audit gap is itself a finding, per D-8.4i.5,** and its price is now on the record rather than
argued: **of the three rejections that were written down, one has been measured false.** Four
unexamined rejections at an observed 1-in-3 falsification rate, on the story whose subject is *"does
the licence gate hold"*, on a compliance boundary. Task 9 registers this.

**6. WHY TWO SHIPPED TESTS REVERSE, AND WHY THAT IS A STRENGTHENING.** `TestClassifyOFL`'s
bundled-notice case and `TestClassifyUbuntuFontLicence`'s equivalent both assert that a file bundling a
second licence's notice classifies as `MIT`. Under D-8.4i.1 that is precisely the input now ruled
unclassifiable. The epic's standing rule is *"guards are widened or replaced, never weakened, with the
old state red-proved"* — this is a **replacement by a stricter guard**: the set of texts that pass
shrinks, and the two reversed cases move from "passes under a possibly-wrong label" to "refused". Both
must carry the old expectation in a comment naming D-8.4i.1, so the reversal is legible as deliberate
rather than as a test edited to make a change go green.

## Verification

Cadence is **per-epic** (D-000.4). This story is `lint`-module-shaped and touches no rendering path,
so the four `FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and Playwright are **excluded** —
they come due at Epic 8's boundary gate. `test:e2e:compile` still runs (the compile-only obligation).

**Commands:**
- `cd lint && go test -count=1 ./...` — expected: **four `ok`, zero FAIL**. ⚠ **`-count=1` is
  mandatory**: the `rules` package scans by walking a directory, and Go's test cache does not track
  `ReadDir`, so a cached `ok` here is **no measurement at all**. *(Measured at `7e4b2c4`: four `ok`.)*
- `cd lint && go vet ./...` — expected: no output. *(Measured clean at `7e4b2c4`.)*
- `cd lint && go run ./cmd/genmanifest && git diff --exit-code lint/MANIFEST.md` — expected: exit 0.
  Run **twice** for idempotency. A changed label is a **finding to route**, not a re-record.
- `cd folio-go && go test -count=1 ./...` — expected: **1815 pass / 2 fail / 5 skip**. Standing red
  **#1**, by identity: `TestCorpusMeetsP6ExerciseFloors` + its `P6g_(opaque_names)` subtest
  (`got 7, need >=20`). Re-measure; the pass count moves between stories.
- `cd folio-go && go vet ./...` — expected: no output.
- `gofmt -l folio-go lint` **run from the REPO ROOT** — expected: **exactly one line**,
  `lint/internal/rules/licencegraph_test.go`. Standing red **#2**, **proven pre-existing** (DW-116,
  D-8.4h.5). **Do NOT reformat it** — that would put an unrelated file in this story's diff. A
  **second** file appearing is a real failure. ⚠ Running this from inside `folio-go/` prints an
  `lstat` error that reads like "clean"; treat any `lstat` line as a non-measurement.
  *(Measured at `7e4b2c4`: exactly the one expected line.)*
- `cd folio-designer && npm test && npm run typecheck && npm run lint && npm run build && npm run test:e2e:compile`
  — expected: tests pass from a baseline of **40 files / 411 tests**; typecheck exit 0; **lint prints
  exactly 4 `only-export-components` warnings** — standing red **#3**, pre-existing, not a regression;
  build and e2e-compile exit 0.
- `shasum -a 256 fixtures/*/expected.pdf` **from the repo root** — expected: **23** digests,
  byte-identical to `7e4b2c4`. A moved golden is a **HALT**. ⚠ A `cd lint &&` earlier in the same
  compound command breaks this relative path and prints `no matches found`, which reads like a missing
  corpus rather than a wrong cwd. *(Measured at `7e4b2c4`: 23.)*
- `md5 -q README.md` — expected: `078d7d80d518d54af2fc04fb270d46b8`. *(Measured, unchanged.)*
- `awk '/maximumCacheAssets/' folio-designer/src/release-payload.ts` — expected: `= 64`.
  *(Measured, unchanged.)*

**Manual checks:**
- The task-1 census output is present in the Delivery Log, and the commit carrying it makes **no**
  refusal fatal — verify by `git show --stat` on that commit and by reading its diff for any change to
  a `switch` arm or a returned `Family`.
- Each new refusal arm was red-proved **by deletion**, and the red sets are **disjoint** — each arm
  reddens its own named test and nothing else. Record which test each mutation reddened.
- The allowlist red-proof (AC7) reddened **exactly one** test. If four packages red, the guard is not
  what caught it — report that rather than accepting the red.
- The two corrected comments contain their original wording verbatim.

## Auto Run Result

Status: done
Blocking condition: none

### Build dispatch, 2026-09-02 — baseline `582a01a` on `main`, tree clean

The spec's frontmatter `baseline_revision` was `7e4b2c4`; the single commit between it and `582a01a`
touches only `_bmad-output/`, so it is not drift and the Code Map's anchors still resolve. Baseline
re-recorded as `582a01a`.

**Implemented change.** `ClassifyLicenceText` no longer returns on the first signal it finds. It now
collects every SPDX line, every licence name and — only when the text names no licence at all — the
grant clauses, and resolves them in one fixed order: any copyleft signal refuses as copyleft naming
the identifier; an unresolved name is `FamilyUnknown`; two or more distinct permissive identifiers is
`FamilyUnknown`; exactly one identifier is that identifier. This closes DW-125's gate bypass and
DW-124's greedy catch-all, and discharges DW-117. The two comments that asserted those paths were
fail-safe are corrected in place with their originals preserved verbatim. Task 1's report-only census
was committed alone and first, before any refusal became fatal.

**Files changed.**
- `lint/internal/licence/licencesignals.go` — new: signal collection, the name and clause tables, and
  the fixed resolution order.
- `lint/internal/licence/classify.go` — `ClassifyLicenceText` delegates to the new rule; the two false
  comments corrected in place with originals verbatim.
- `lint/internal/licence/classify_test.go` — a case per I/O Matrix row, the two shipped expectations
  reversed to `FamilyUnknown` with their old expectations recorded, and the determinism pin.
- `lint/internal/licence/licencecensus_test.go` — new: the population census, now pinning all 35
  verdicts against a test-owned table.
- `lint/internal/manifest/manifest_test.go` — the allowlist pinned to a test-owned literal naming
  D-8.5.3, and the gate-surface test for DW-125's reported shape.
- `lint/internal/rules/excludedpathfonts_test.go` — new: the excluded-path sfnt tripwire.
- `_bmad-output/implementation-artifacts/deferred-work.md` — DW-117/120/124/125 closed, DW-123
  amended, DW-126 and DW-127 registered.

**Review findings breakdown.** 9 patched (2 high, 3 medium, 4 low), 3 deferred (2 medium, 1 low),
24 rejected — every rejection enumerated with its reason in the Review Triage Log above, per
D-8.4i.10. Two of the rejected findings were checked and found FALSE as claims about the code, and
are recorded as such rather than dropped silently. No intent gaps and no bad_spec findings: the
`<intent-contract>` is byte-identical to baseline (verified by digest), and every matrix row was
measured to return exactly the value the matrix specifies.

**Follow-up review recommendation: `true`.** Patched by severity: high 2, medium 3, low 4. The rule
triggers twice over — a high-severity patched finding is present, and the score
`3 x 3 medium + 1 x 4 low = 13` is at or above 5.

**Verification performed** (independently re-measured at review, not taken from the implementation's
report):
- `cd lint && go test -count=1 ./...` — four `ok`, zero FAIL.
- `cd lint && go vet ./...` and `cd folio-go && go vet ./...` — no output, exit 0.
- `cd lint && go run ./cmd/genmanifest && git diff --exit-code -- lint/MANIFEST.md` — exit 0 on both
  runs. Note the cwd trap: run from inside `lint/` the `git diff` path is ambiguous and exits 128,
  which does not read as a failure of the manifest.
- `cd folio-go && go test -count=1 ./...` — **1815 pass / 2 fail / 5 skip**, counted from `-json`
  test-level events. The two failures are standing red #1 by identity:
  `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` subtest.
- `gofmt -l folio-go lint` from the repo root — exactly one line,
  `lint/internal/rules/licencegraph_test.go` (standing red #2, DW-116). Not reformatted.
- designer: **40 files / 411 tests** passed; typecheck, build and `test:e2e:compile` exit 0; lint
  prints exactly **4** `only-export-components` warnings (standing red #3).
- `shasum -a 256 fixtures/*/expected.pdf` — **23** digests; no fixture appears in
  `git diff 582a01a..HEAD`, so the goldens are byte-identical to baseline.
- `md5 -q README.md` — `078d7d80d518d54af2fc04fb270d46b8`. `maximumCacheAssets = 64`.

**Manual checks.**
- **The build-time census agrees with the plan gate's.** 35 texts measured (26 committed files + 9
  dependency licences), 0 changing verdict. The nine dependencies are `textshape` and `goldmark`
  (MIT by name signal) and `go-cmp`, `x/mod`, `x/net`, `x/sync`, `x/sys`, `x/telemetry`, `x/tools`
  (BSD-3-Clause through the clause disjunct, no name signal and no SPDX line) — the seven Design
  Note 1 names, confirming the BSD clause is still treated as a clause and the dependency scan does
  not red.
- The census commit `2e9365e` is additive only — two new files, no change to `classify.go`, no
  switch arm and no returned `Family` — so no refusal became fatal in the commit that first measured
  the population.
- **AC7 red-proved independently at review:** appending `"GPL-3.0"` to `fontAssetLicenceAllowlist`
  reds exactly one test, `TestFontAssetLicenceAllowlistIsTheOwnersFourIds`, in one package; the other
  three packages stayed `ok`.
- **AC8 red-proved independently at review:** a real sfnt program planted under
  `folio-go/testdata/lint/embed-font/allembed/fonts/` reds the tripwire naming the file, while the
  `.ttf`-named non-font stub beside it still passes — the key is the magic bytes, not the extension.
- The three bypass shapes are each refused by a named subtest that ran and passed, and the corrected
  comments carry their original wording verbatim.

**Residual risks.**
- P2's own red-proof disagrees with itself usefully: narrowing SPDX collection back to the first match
  does **not** red the new gate test, because the GPL *name* signal reaches the copyleft arm on its own
  path. The gate test and the collect-all classifier rows measure different mechanisms and are not
  substitutes; this is recorded in the test's header.
- Story 8.4h's AC5 bound — two independent mutations producing two distinct reds — is genuinely gone,
  not merely weakened, because name-table ids now route through `classifyBySPDX`. Measured: deleting
  `permissiveSPDX["Ubuntu-font-1.0"]` reds five tests; deleting the `licenceNames` entry reds two and
  not the map-entry test. The independence is now one-way, and that is written into the code.
- Three deferrals are carried in frontmatter, the sharpest being that the tripwire keys on four sfnt
  magics and not on wOFF/wOF2, so a web font under an excluded path still slips past.
- Design Note 4's figure of 66 tracked files under `folio-go/testdata/lint/` is 64 at `582a01a`; the
  drift is in the baseline, not in this story.

### Plan-only dispatch, 2026-09-02 — baseline `7e4b2c4`

Plan-only dispatch (`Halt after planning.`), 2026-09-02, baseline `7e4b2c4` on `main`, tree clean.
No code written. The only tree change is this spec file (untracked). No commits, nothing pushed.

Measurements taken at this plan gate, with their probes deleted and the tree re-verified clean:

- **D-8.4i.1's report-only census — TAKEN AND CLEAN.** 9 dependency + 12 asset + 8 fixture
  `LICENSE*` files; zero carry two distinct identifiers, an unresolved name signal, or a
  copyleft/permissive mix. Nothing changes verdict under the proposed rule. See Design Note 1.
- **The predicted OFL-1.0 fourth instance HOLDS.** OFL 1.0 classifies `(permissive, "MIT")`, not
  `FamilyUnknown`. Both false comments confirmed false. See Design Note 2.
- **A finding not in any register:** DW-125's bypass works with a copyleft **SPDX line** as well as
  with copyleft **marker text** (`MIT` line then `GPL-3.0-only` line → `(permissive, "MIT")`). The
  ruled fix covers it; the rejected alternative would not have.
- **Two premise corrections, neither changing the verdict** (D-8.5.10's distinction): the npm path
  never reaches `ClassifyLicenceText` (Design Note 3), and `checkSfnt` is genuinely unreachable but
  the `lint` module owns its own sfnt reader, so D-8.4i.4's tripwire is buildable without a second
  reader (Design Note 4). Both flagged for the plan gate.
- **D-8.4i.5 recovery: 3 of 7 rejections recovered, 1 of those 3 overturned, 4 recorded as
  "could not look".** See Design Note 5.

Baselines re-measured at `7e4b2c4`: `lint` four `ok` with `-count=1`; `go vet ./...` clean;
`gofmt -l folio-go lint` from the repo root prints exactly `lint/internal/rules/licencegraph_test.go`
(DW-116); 23 golden digests; `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`;
`maximumCacheAssets = 64`.

## Delivery Log

**Story 8.4i — delivered 2026-09-02 on `main`, five commits, baseline `582a01a`.**

| commit | subject |
|---|---|
| `2e9365e` | Measure every licence file in the repo before making any refusal fatal (task 1) |
| `0c6e3b4` | Make the classifier collect every signal, and correct the comments that said it was safe (tasks 2–5) |
| `7cb8148` | Pin the owner's four-id font allowlist to a literal the code cannot move (task 6) |
| `90f0820` | Trip if a real font ever hides under the path the asset gate skips (task 7) |
| `fb92156` | Close DW-117/120/124/125, amend DW-123, register DW-126/127, and record the census (tasks 8–9) |
| `7c6e56a` | Review triage: make the census observe something, and test the bypass where it was reported (P1–P6, P9) |
| *(this commit)* | Review triage: correct the records that cited a differential which had stopped differing (P1b, P7, P8) |

### AC1 — the report-only census, task 1, committed at `2e9365e` before any refusal became fatal

**AT `2e9365e`, AND ONLY THERE, THE CENSUS WAS A DIFFERENTIAL.** That commit left
`ClassifyLicenceText` holding the old first-match switch and called the new rule only from the test, so
the two columns were **genuinely different code**. Population: **26 committed `LICENSE*`/`COPYING`
files** — a deliberate superset of the 12 asset files and the 8 lint fixtures, since a census that
enumerated only what it expected to find would not be a census — plus the **9 dependency licences** the
three Go module graphs resolve to. **35 texts, 0 changed verdict.** That measurement is what licensed
the next commit to make refusals fatal, and it is a fact about `2e9365e`.

⚠ **IT STOPPED BEING A MEASUREMENT AT `0c6e3b4`, AND THE TEST WENT ON CLAIMING OTHERWISE** — review
finding P1, and this story's own subject arriving at instance four (D-8.0.1) inside the test written to
catch instance three. Once `ClassifyLicenceText`'s body became `return classifyByAllSignals(text)` the
census compared a function to **itself**: `differs` was false by construction and "0 change verdict"
could not have printed anything else. Probed at triage: setting the OFL entry's `requiredVersion` to
`"VERSION 9.9"` reclassifies all eleven committed OFL files to `(unknown, "")` — a total failure of the
shipped asset gate — and the differential census still printed `0 change verdict` and **passed**.

**Repaired at `7c6e56a`:** the census now **pins** every one of the 35 verdicts against a test-owned
table (D-8.4i.3's reasoning — a derived expectation passes any edit), reds on a population member
missing from the table and on a table entry whose file has vanished, and keeps the vacuity floors.
Re-probed with the same `VERSION 9.9` mutation: **it now reds, on 11 named files.** The verdicts
recorded below are the pinned table's contents.

```
   LICENSE                                                                  shipped=(permissive,"MIT") collect-all=(permissive,"MIT")
   folio-designer/public/fonts/ibmplexmono/LICENSE-OFL.txt                  shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/public/fonts/ibmplexsans/LICENSE-OFL.txt                  shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/public/fonts/ibmplexsansthai/LICENSE-OFL.txt              shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/public/fonts/notosans/LICENSE-OFL.txt                     shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/public/fonts/notosanssc/LICENSE-OFL.txt                   shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/public/fonts/notosansthai/LICENSE-OFL.txt                 shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-designer/third-party-notices/pdfjs-dist/LICENSE-APACHE-2.0         shipped=(permissive,"Apache-2.0") collect-all=(permissive,"Apache-2.0")
   folio-designer/third-party-notices/pdfjs-dist/LICENSE-CMAPS              shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   folio-designer/third-party-notices/pdfjs-dist/LICENSE-LIBERATION         shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-go/fonts/notosans/LICENSE-OFL.txt                                  shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-go/fonts/notosanssc/LICENSE-OFL.txt                                shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-go/fonts/notosansthai/LICENSE-OFL.txt                              shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt                      shipped=(permissive,"CC0-1.0") collect-all=(permissive,"CC0-1.0")
   folio-go/testdata/fonts/LICENSE-Roboto.txt                               shipped=(permissive,"Apache-2.0") collect-all=(permissive,"Apache-2.0")
   folio-go/testdata/fonts/notosansthai-variable-testonly/LICENSE-OFL.txt   shipped=(permissive,"OFL-1.1") collect-all=(permissive,"OFL-1.1")
   folio-go/testdata/lint/wordlist-assets/compliant/folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt shipped=(unknown,"") collect-all=(unknown,"")
   folio-go/testdata/lint/wordlist-assets/violating/folio-go/internal/text/wordlist/LICENSE-CC0-1.0.txt shipped=(unknown,"") collect-all=(unknown,"")
   lint/testdata/licence/copyleft/example.test/agpl-lib/LICENSE             shipped=(copyleft,"AGPL-3.0-only") collect-all=(copyleft,"AGPL-3.0-only")
   lint/testdata/licence/copyleft/example.test/gpl-lib/LICENSE              shipped=(copyleft,"GPL-3.0-only") collect-all=(copyleft,"GPL-3.0-only")
   lint/testdata/licence/copyleft/example.test/lgpl-lib/LICENSE             shipped=(copyleft,"LGPL-3.0-only") collect-all=(copyleft,"LGPL-3.0-only")
   lint/testdata/licence/copyleft/example.test/sspl-lib/LICENSE             shipped=(copyleft,"SSPL-1.0") collect-all=(copyleft,"SSPL-1.0")
   lint/testdata/licence/permissive/example.test/apache-lib/LICENSE         shipped=(permissive,"Apache-2.0") collect-all=(permissive,"Apache-2.0")
   lint/testdata/licence/permissive/example.test/bsd-lib/LICENSE            shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   lint/testdata/licence/permissive/example.test/mit-lib/LICENSE            shipped=(permissive,"MIT") collect-all=(permissive,"MIT")
   lint/testdata/licence/permissive/example.test/ufl-lib/LICENSE            shipped=(permissive,"Ubuntu-font-1.0") collect-all=(permissive,"Ubuntu-font-1.0")
   dep folio-go -> github.com/boxesandglue/textshape                        shipped=(permissive,"MIT") collect-all=(permissive,"MIT")
   dep lint -> github.com/google/go-cmp                                     shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> github.com/yuin/goldmark                                     shipped=(permissive,"MIT") collect-all=(permissive,"MIT")
   dep lint -> golang.org/x/mod                                             shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> golang.org/x/net                                             shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> golang.org/x/sync                                            shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> golang.org/x/sys                                             shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> golang.org/x/telemetry                                       shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
   dep lint -> golang.org/x/tools                                           shipped=(permissive,"BSD-3-Clause") collect-all=(permissive,"BSD-3-Clause")
CENSUS: 35 licence texts measured (26 committed files + 9 dependency licences); 0 change verdict
```

Notes on the record: `hashmatrix` resolves **zero** non-main modules, so its scan is vacuous — recorded
rather than assumed. The two `folio-go/testdata/lint/wordlist-assets/…/LICENSE-CC0-1.0.txt` fixtures
read `(unknown, "")` under **both** classifiers; they are deliberately truncated fixtures under the
gate's excluded path, and are unchanged by this story.

**`2e9365e` makes no refusal fatal.** `git show --stat 2e9365e` lists two files, both new; it touches
no `switch` arm and no returned `Family` in `classify.go`. That is why the rule landed in its own file.

### AC2–AC5 — the classifier, tasks 2–5

Every I/O Matrix row is a named subtest of `TestClassifyCollectsEverySignal`. The MIT and BSD fallbacks
**survive** (AC5): a bare grant clause with no name is still `MIT`, and a Go-style BSD text is still
`BSD-3-Clause` — the constraint Design Note 1 flagged, since 7 of the 9 dependency licences reach BSD
through the `"REDISTRIBUTION AND USE IN SOURCE"` **clause**, which the implementation keeps as a clause
and not a name.

**Red-proof by DELETION, four mutations applied alone, red sets MEASURED and DISJOINT:**

| mutation | what reds |
|---|---|
| delete the copyleft arm | 7 copyleft rows of `TestClassifyLicenceText`; 4 copyleft rows here; all 3 of `TestCopyleftTieBreakIsDeterministic`; manifest's `TestResolveAssetsRefusesACopyleftFontLicence` and the wordlist copyleft arm |
| delete the unresolved arm | the two REVERSED bundled-notice cases; the unrecognised-SPDX-id row; the wordlist non-permissive arm |
| weaken `len(permissiveIDs) != 1` to `== 0` | exactly the three two-identifier rows — **nothing else in any package**, which proves conflict detection is a NEW refusal |
| delete the `!namedAny` guard on the clause table | `TestCommittedOFLTextClassifiesAsOFL11` (the real artifact), both grant-clause cases, and five manifest/rules tests including `TestCommittedAssetPopulationClassifiesCleanly` — the blast radius that IS the reason a clause is weaker than a name |

**A regression caught by an existing guard, and fixed rather than accommodated.** The first
implementation returned `(FamilyUnknown, "")` for an SPDX identifier on neither list. That reddened
`TestWordlistSiteEnforcesThePermissiveSetNotTheFontAllowlist`'s "not a permissive licence" arm, which
collapsed onto the neighbouring "could not be classified" one. An unrecognised SPDX id now still NAMES
the identifier it could not place (AD-14); an unresolved licence **name** still returns `""`, because
there is no identifier to give. `TestClassifyCollectsEverySignal`'s corresponding row was updated to
the better diagnostic, with the reason recorded in the case.

### AC6 — the corrected comments

Both live in `classify.go`, where the branches were, each carrying its **original wording verbatim** in
an indented block followed by the correction, the measurement, its date, and the note that the claim it
replaces is why 8.4h's review rejected DW-124. Three claims of the banned class were present, not two —
the OFL version-conjunct claim, the UFL version-conjunct claim, and the UFL British-spelling
loud-miss claim — and all three are corrected. **No comment in the file now asserts that a near-miss on
a licence name reaches `FamilyUnknown` by virtue of a required version conjunct.** The rest of the file
was audited: the CC0 narrow-marker comment, `permissiveSPDX`'s D-2.1.3 fail-safe note, the
`Ubuntu-font-1.0` entry note and the `Family` commercial-EULA note are all still true, and the last two
are more true after this change than before.

### AC7 — the allowlist pin, task 6

`TestFontAssetLicenceAllowlistIsTheOwnersFourIds`. With `"GPL-3.0"` appended, **exactly one test reds,
in one package**, on its own message naming D-8.5.3; `genmanifest`, `licence` and `rules` stay green.

### AC8 — the excluded-path tripwire, task 7

`TestNoRealFontHidesUnderAnExcludedPath`. **64 files scanned across 1 excluded directory, zero font
magics.** Non-vacuous: the same detector over `folio-go/fonts/notosans/NotoSans-Regular.ttf` returns
true. Red-proved by copying a real sfnt program under the excluded path — the tripwire reds naming the
file; probe removed. **No second sfnt reader**: the test declares `package rules` and calls the
existing `looksLikeSfnt` as-is, so D-8.4i.4's escape hatch was not needed. The excluded paths are
**derived from `manifest.go`'s own source**, not restated (DW-119's lesson).

*(Design Note 4 recorded 66 tracked files under that path at `7e4b2c4`; the tree carries 64 at
`582a01a`, and `git ls-files` and the on-disk walk agree exactly. The figure moved with the baseline,
not with this story.)*

### AC9 — byte identity

`shasum -a 256 fixtures/*/expected.pdf` from the repo root: **23** digests. `git diff --stat
582a01a..HEAD` lists **six code files, every one under `lint/internal/`**, plus the two records this
story is obliged to keep — `_bmad-output/implementation-artifacts/deferred-work.md` and this spec.
*(Corrected at review triage, finding P7: the original wording said "six files" full stop, in the very
commit that also changed those two records.)* No fixture, no golden, no `.folio`, no engine file,
nothing under `folio-go/fonts/`. `lint/MANIFEST.md` regenerated **twice** via
`cd lint && go run ./cmd/genmanifest`, `git diff --exit-code lint/MANIFEST.md` exit 0 both times.
`README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`. `maximumCacheAssets = 64`.

### Task 9 — the register

Closed **DW-117**, **DW-120**, **DW-124**, **DW-125**, each with a closing note recording that the fix
delivered was the RULED one and not the one the entry proposed. **DW-123** amended with the Go-side
tripwire's outcome, kept OPEN for its own unrelated `readSync` half. Two new entries:

- **DW-126** — D-8.4i.5's finding: four of 8.4h's seven rejections were never written down, and of the
  three that were, one has been measured false. Recorded as a **priced finding**, with the four
  recorded as "could not look" rather than omitted (D-8.4.33).
- **DW-127** — found while rebuilding the name table: a text naming the **2-clause** BSD licence is
  labelled `BSD-3-Clause`, and has been since Story 1.3. LOW, attribution only, not live in the
  population. **Behaviour preserved exactly** and routed to Epic 15's release gate per **D-8.4i.6's
  bound** — it is not a live bypass of a fail-closed gate, so it does not buy an `8.4j`.

### Verification, measured at delivery

| command | result |
|---|---|
| `cd lint && go test -count=1 ./...` | **four `ok`, zero FAIL** |
| `cd lint && go vet ./...` | no output |
| `cd lint && go run ./cmd/genmanifest && git diff --exit-code lint/MANIFEST.md` | exit 0, twice |
| `cd folio-go && go test -count=1 ./...` | standing red **#1** by identity: `TestCorpusMeetsP6ExerciseFloors` in `internal/text`; every other package `ok` |
| `cd folio-go && go vet ./...` | no output |
| `gofmt -l folio-go lint` **from the repo root** | **exactly one line**, `lint/internal/rules/licencegraph_test.go` — standing red **#2** (DW-116), untouched |
| `cd folio-designer && npm test` | **40 files / 411 tests**, all pass |
| `npm run typecheck` / `npm run build` / `npm run test:e2e:compile` | exit 0 |
| `npm run lint` | **exactly 4** `only-export-components` warnings — standing red **#3** |
| `shasum -a 256 fixtures/*/expected.pdf` | **23** digests |
| `md5 -q README.md` | `078d7d80d518d54af2fc04fb270d46b8` |
| `awk '/maximumCacheAssets/' folio-designer/src/release-payload.ts` | `= 64` |

**Re-run in full after review triage (2026-09-02).** Unchanged: four `ok` in `lint` with `-count=1`;
`go vet` clean in both modules; `lint/MANIFEST.md` regenerates with no diff, twice; **23** golden
digests; `README.md` md5 `078d7d80d518d54af2fc04fb270d46b8`; `maximumCacheAssets = 64`; designer 40
files / 411 tests with typecheck, build and e2e-compile at exit 0. All three standing reds reproduce
**by identity and by figure**: `TestCorpusMeetsP6ExerciseFloors` / `P6g_(opaque_names)` at
`got 7, need >=20`; exactly one `gofmt -l` line, `lint/internal/rules/licencegraph_test.go`, still
unformatted and still untouched; exactly **4** `only-export-components` warnings.

**Block Ifs: none tripped.** The census reddened nothing; no golden moved; `lint/MANIFEST.md`
regenerates with no diff; no asset directory holds more than one `LICENSE*` (DW-118 still not live); no
demonstrated live bypass was found beyond the two already chartered, so **D-8.4i.6's bound holds and
nothing here buys an `8.4j`**.

### 2026-09-02 — done (close)

Baseline `582a01a` on `main`. **Eight story commits plus one record commit**, all local, **nothing
pushed** — `origin/main` is still `c985b9c`, twenty-five commits behind. Closed by this entry's own
commit, the tip of `main` at close (a commit cannot carry its own hash in its diff).

| commit | what it did |
|---|---|
| `2e9365e` | Task 1's report-only population census. **Additive only — two new files, 458 insertions, zero deletions, `classify.go` untouched**, so the new rule was dead code called by the census test alone. This is D-8.4i.6's hard constraint discharged: nothing became fatal in the commit that first measured the population. |
| `0c6e3b4` | Tasks 2–5. `ClassifyLicenceText`'s body becomes the collect-all rule; the two false comments corrected in place with originals verbatim; the classifier's test cases, including the two shipped expectations reversed to `FamilyUnknown`. |
| `7cb8148` | Task 6, DW-120: the owner's four-id allowlist pinned to a test-owned literal naming D-8.5.3. |
| `90f0820` | Task 7, DW-123: the excluded-path sfnt tripwire, keyed on magic bytes via the `lint` module's own reader. |
| `fb92156` | Task 9: DW-117/120/124/125 closed, DW-123 amended, DW-126 and DW-127 registered. |
| `7c6e56a` | Review triage, code half: the census re-pinned against a test-owned table, the gate-surface bypass test added, `readFirstBytes` given `io.ReadFull` semantics, two docs corrected. |
| `66e2db7` | Review triage, record half: the `0 of 35` figure re-attributed to `2e9365e` in the three places that had stated it as a standing property. |
| `ec40fc2` | The Review Triage Log, with all 24 rejections enumerated. |
| `98e98f1` | D-8.4i.11 / D-8.4i.12 / D-8.4i.13 written into the decision log. |

**Provenance, the standing gate (D-8.4h.6).** All nine `git show --stat` clean: only this story's files
in every one, no unrelated churn, no generated artifact. `lint/MANIFEST.md` appears in no commit —
correct, since the labels did not move. Author and committer are `Panit Wechasil <panitw@hotmail.com>`
throughout; **both required trailers exact on all nine**. Branch `main`, no branch created. **No commit
in this story was created by a step-03 subagent** — there is no instance five.

**Decisions applied.** D-8.4i.1 (collect every signal, fixed resolution order) · D-8.4i.2 (anchor the
name, not the spelling) · D-8.4i.3 (pin the allowlist to a literal the code cannot move) · D-8.4i.4 /
D-8.4i.8 (the tripwire is built, using the reader the module already had) · D-8.4i.5 / D-8.4i.10 (the
rejection record) · D-8.4i.6 (the bound) · D-8.4i.7 (OFL-1.0's fourth instance; the copyleft-SPDX-line
bypass) · D-8.4i.9 (the BSD clause is a clause) · D-8.4i.11 (the census that stopped measuring) ·
D-8.4i.12 (the follow-up flag) · D-8.4i.13 (two claims kept as facts). Also D-8.5.3, D-8.5.10,
D-8.5.13, D-8.4h.5, D-8.0.1, D-2.1.3, D-1.3.4.

**Triage: patch 9 (2 high / 3 medium / 4 low) · defer 3 (2 medium / 1 low) · reject 24 (5 medium / 19
low).** All 24 rejections are individually enumerated with a reason in the Review Triage Log — that is
**D-8.4i.10's requirement discharged**, and it was bought by Story 8.4h's record being **unrecoverable
for 4 of its 7 rejections**, with 1 of the 3 readable ones measured false. The enumeration earned its
cost immediately: it is what made **R4 checkable**, and R4 turned out to be false (below).

**`followup_review_recommended: true` was discharged WITHOUT a second review pass, per D-8.4i.12** —
the flag fired on 2 highs, score 13. The reason is measured, not asserted: at Story 8.4h the same four
layers produced seven rejections, four unwritten and one false, and **both** genuine defects were found
by the adversarial close rather than by the layers. This close carried the scrutiny instead, directed at
three named targets. **It found one HIGH the layers had seen and the triage had wrongly dropped.**

**Gates measured at the close, not relayed.**

| gate | measured |
|---|---|
| `lint` `go test -count=1 ./...` | four `ok`, zero FAIL |
| `lint` / `folio-go` `go vet ./...` | no output, both |
| `genmanifest` twice, then `git diff --exit-code -- lint/MANIFEST.md` **from the repo root** | exit 0 both runs |
| `folio-go` `go test -count=1 ./...` | **1815 pass / 2 fail / 5 skip**, counted from `-json` test events |
| `gofmt -l folio-go lint` from the repo root | exactly one line, `lint/internal/rules/licencegraph_test.go` |
| designer `npm test` | **40 files / 411 tests**, all pass |
| designer `typecheck` / `build` / `test:e2e:compile` | exit 0 |
| designer `npm run lint` | exactly **4** `only-export-components` |
| `shasum -a 256 fixtures/*/expected.pdf`, **diffed against a worktree at `582a01a`** | **23** digests, `diff` exit 0 — byte-identical |
| `md5 -q README.md` | `078d7d80d518d54af2fc04fb270d46b8` |
| `maximumCacheAssets` | `= 64` |

All three standing reds reproduce **by identity**: `TestCorpusMeetsP6ExerciseFloors` + `P6g_(opaque_names)`
in `folio-go/internal/text` (#1); the one `gofmt` line, left unformatted and untouched (#2, DW-116); the
four designer lint warnings (#3). None is this story's.

**Heavy tests: NOT RUN, and not waived.** Cadence is **per-epic** (D-000.4). The four
`FOLIO_MATRIX_TARGET` legs, `TestCrossTargetByteIdentity` and the Playwright suite are **written and
compiling but unrun**, and come due at **Epic 8's boundary gate**. Compilation was verified rather than
assumed: `go vet -tags matrix ./...` exits 0 and every `folio-go` package links its `matrix`-tagged test
binary, so the eventual catch-up run will not open on a compile error banked from this story; the
Playwright specs compile via `test:e2e:compile`.

**The census's `0 of 35`.** Cited only as *a measurement taken at commit `2e9365e`*, where the two
classifiers were genuinely different code. It is **not** a standing property and is not repeatable: from
`0c6e3b4` the differential compared a function with itself. What stands today is the **pinned table** of
35 verdicts, which is a different and stronger claim.

**Adversarial pass — six directed checks, all performed.**

1. **The census pin is genuine.** `pinnedCensus` is a hand-written 35-entry literal; nothing in it is
   derived from the classifier, the name table, the clause table or `permissiveSPDX`, and the only
   derived part is the *population* (`git ls-files`), which is the right thing to derive. Mutation-proved
   twice, not once: the `VERSION 9.9` mutation reds it on **11 named files**, and a second, independent
   mutation — promoting the BSD clause to a non-resolving **name** signal, D-8.4i.9's exact named hazard —
   reds it on **8 named files**, naming all seven Go-style BSD dependencies plus `LICENSE-CMAPS`. A first
   attempt at that mutation (promoting the clause to a *resolving* name) left the suite green and is
   recorded as degenerate: it changed the mechanism without reinstating the defect.
2. **No fail-open path in the new classifier.** Every early return, empty collection and default was
   probed with 27 synthetic texts. The only outcome that passes a gate is `FamilyPermissive` with exactly
   one identifier `classifyBySPDX` recognises; everything else — empty text, whitespace, an unrecognised
   SPDX id, a copyleft name under any ordering, an unresolved name beside a valid declaration, MPL-2.0
   prose, EULA prose, a lower-cased identifier — lands on `FamilyUnknown` or `FamilyCopyleft`, both
   refusals. One latent coupling worth naming: the gate's arms would admit a `(FamilyUnknown, id)` pair if
   `id` were ever on the font allowlist, and that cannot happen only because `unresolvedID` is by
   construction an identifier `classifyBySPDX` rejects while the allowlist is a subset of what it accepts.
   That invariant is unstated in the code but is now guarded by `7cb8148`'s allowlist pin.
3. **The three bypass shapes are refused, both orders each.** GPL text + stray `MIT` line → `(copyleft,
   "GPL-3.0")` either order. Two permissive ids → `(unknown, "")`, identical both orders. `MIT` +
   `GPL-3.0-only` → `(copyleft, "GPL-3.0-only")` either order. Shapes (a) and (c) name the copyleft
   identifier, as D-8.4i.1 requires.
4. **The BSD clause disjunct still behaves as a clause.** A Go-style BSD text with no name signal and no
   SPDX line returns `(permissive, "BSD-3-Clause")`, and the census shows all seven dependencies reaching
   it that way. Check 1's second mutation is the proof that the guard against it becoming a name is live
   and names the seven.
5. **All three re-checked items held, one with a correction.** *(a)* The **wOFF/wOF2 gap is real and the
   deferral is honestly scoped** — files beginning `wOF2` and `wOFF` planted under an excluded path left
   the tripwire at `ok`. Registered as **DW-128**; it had been recorded only in prose. *(b)* The **gate
   test genuinely does not measure the collect-all rule** — narrowing SPDX collection to first-match
   leaves `TestResolveAssetsRefusesTheDW125BypassAtTheGate` green and the whole `manifest` package `ok`,
   while four subtests of `TestClassifyCollectsEverySignal` red. D-8.4i.13's first claim is true as
   written. *(c)* The **one-way independence conclusion holds but its figures were understated** — see
   the correction below.
6. **Both teeth-proofs pass.** Appending `"GPL-3.0"` to the allowlist reds **exactly one** test,
   `TestFontAssetLicenceAllowlistIsTheOwnersFourIds`, in one package; the other three packages stayed
   `ok`. A real sfnt planted under `folio-go/testdata/lint/…/fonts/` reds the tripwire naming the file —
   **and reds it again when the same font is renamed `.README`**, while the `.ttf`-named text stub beside
   it still passes. Keyed on magic bytes, not on the extension, proved in both directions. All mutations
   reverted; `git status` clean after each.

**ONE HIGH FOUND AT THE CLOSE, registered as DW-131 and NOT patched.** `spdxLineRE` captures a **single
token**, and `ClassifyLicenceText` never routes to `ClassifySPDXExpression` — that function's only
non-test callers are on the npm lockfile path. So a **compound** SPDX expression in a licence text
contributes only its first term. Measured: `SPDX-License-Identifier: MIT OR GPL-3.0-only` →
`(permissive, "MIT")`; a font `LICENSE` declaring `OFL-1.1 OR GPL-3.0-only` → `(permissive, "OFL-1.1")`,
which is on the owner's four-id allowlist and therefore **passes the fail-closed asset gate and lands in
`lint/MANIFEST.md` under a permissive label**. Reversing the terms refuses correctly, so it is
**order-dependent — DW-125's exact defect shape, surviving in the case collect-all does not reach**. The
controls are decisive: `ClassifySPDXExpression("MIT OR GPL-3.0-only")` returns **copyleft**, so the two
functions D-8.4i.1 ruled must agree still disagree on the same string.

**This overturns rejection R4**, which was dropped as *"Speculative… routed to `ClassifySPDXExpression`…
Not demonstrated"* — a premise true of the npm path and false of the path under test. The reviewer was
right; the triage was not. It is **not live**: no file in the pinned 35 carries a compound expression,
and the table would red if one appeared. Per **D-8.4i.6** a demonstrated bypass of a gate the epic has
declared fail-closed is the **one class that returns to the engineering lead** to be ruled *against* the
bound — so it is registered and escalated, deliberately **not** fixed here and deliberately **not**
routed to Epic 15 on the closer's own authority. **This supersedes the Auto Run Result's line that "no
demonstrated live bypass was found beyond the two already chartered."**

**One correction applied at the close.** D-8.4i.13's second claim and its code comment state that
deleting `permissiveSPDX["Ubuntu-font-1.0"]` reds **five** tests and deleting the `licenceNames` UFL
entry reds **two**. Re-measured over the whole `lint` module: **seven** and **three** — the extra reds
are `manifest.TestResolveAssetsAcceptsEveryAllowlistedFontLicence` in both cases and
`rules.TestLicenceGraphFixtureScan` in the first. The figures were taken with `go test
./internal/licence/` and written as though they were the suite's whole mutation resolution, which is the
property they are cited for. **The conclusion is unchanged and was re-verified** — the second mutation
still does not red the map-entry test while the first does, so the independence is genuinely one-way.
Only the figures were understated. Corrected in place in `classify_test.go`, in the same
correction-beside-original form the story requires, because *a measurement quoted without the scope it
was taken at* is the defect class this story exists to close.

**Deferrals registered with owners.** **DW-128** — the tripwire is blind to `wOFF`/`wOF2`; owner **Epic
15's release gate**; medium, and it is the one that fails open. **DW-129** — DW-127 is routed to Epic 15
with nothing pinning today's behaviour, so Epic 15 inherits a defect with no red to work against; owner
**Epic 15's release gate**, alongside DW-127. **DW-130** — the tripwire derives the gate's walk exclusion
by regexping source text, so a formatting-neutral refactor reds it with a misleading message; owner
**whichever story next changes the asset walk**; low, and a false positive only. Plus **DW-131** above,
owner **engineering lead**.

**For the orchestrator.** DW-131 needs a ruling before Epic 8 closes: it is D-8.4i.6's stated exception
class, and only the lead can say whether it buys an `8.4j` or is routed to Epic 15 with the rest. Nothing
else blocks. `epic-8` stays `in-progress` — 8.5, 8.6 and 8.4d remain.
