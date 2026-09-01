---
title: 'Story 8.4i: The classifier stops lying to the gate'
type: 'bugfix'
created: '2026-09-02'
status: 'done'
baseline_revision: '582a01aea6bf22aef945a02a4e5d46966be1fe26'
review_loop_iteration: 0
followup_review_recommended: false
context: []
warnings: ['oversized', 'multiple-goals']
deferred: []
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

The previous story made this check refuse things. It did not make it *see* straight. The part that
works out which licence a document is under reads the first label it finds, then stops looking.
Hand it a document carrying two labels — common in redistributed typeface bundles — and it judges
the whole thing by whichever label comes first, even when the body says something the project must
refuse.

It can be fooled a second way. Licences are recognised by name, and a name spelled even slightly
differently is not. Rather than admit it does not know, the check drops it into a catch-all and
credits it to a different licence — and that wrong name reaches paperwork the project publishes.

This story makes "this file says two different things" count as not knowing, and not knowing now
fails. A half-recognised name counts as not knowing too.

The uncomfortable part: the safety these paths were *believed* to have was written down in the code,
as notes explaining why each was fine. Those notes were wrong. They are corrected alongside
the behaviour, original wording kept, because a wrong note is what stopped anyone looking, three
times running.

The risky step is the first. Before anything becomes fatal, the new rule runs over every licence
document the project already relies on, purely to report what it would say. If that turns up
something legitimate, the answer is to stop and ask — never to soften the rule until it goes quiet.

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

Status: ready-for-dev
Blocking condition: none

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
| *(this commit)* | Close DW-117/120/124/125, amend DW-123, register DW-126/127, and record the census (tasks 8–9) |

### AC1 — the report-only census, task 1, committed at `2e9365e` before any refusal became fatal

`TestLicenceSignalCensus` (`lint/internal/licence/licencecensus_test.go`) runs BOTH classifiers side by
side. Population: **26 committed `LICENSE*`/`COPYING` files** — a deliberate superset of the 12 asset
files and the 8 lint fixtures, since a census that enumerated only what it expected to find would not
be a census — plus the **9 dependency licences** the three Go module graphs resolve to. **35 texts, 0
changed verdict.**

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
582a01a..HEAD` lists **six files, all under `lint/internal/`** — no fixture, no golden, no `.folio`, no
engine file, nothing under `folio-go/fonts/`. `lint/MANIFEST.md` regenerated **twice** via
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

**Block Ifs: none tripped.** The census reddened nothing; no golden moved; `lint/MANIFEST.md`
regenerates with no diff; no asset directory holds more than one `LICENSE*` (DW-118 still not live); no
demonstrated live bypass was found beyond the two already chartered, so **D-8.4i.6's bound holds and
nothing here buys an `8.4j`**.
