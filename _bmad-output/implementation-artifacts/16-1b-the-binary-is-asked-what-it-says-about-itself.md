---
title: 'Story 16.1b: The binary is asked what it says about itself'
type: 'feature'
created: '2026-09-03'
status: 'done'
review_loop_iteration: 0
followup_review_recommended: true
baseline_commit: '384c8ac'
baseline_revision: 'e250dc78c779d4a725c5584ec0bc898967ef8146'
context:
  - '{project-root}/_bmad-output/specs/spec-fonts/SPEC.md'
  - '{project-root}/_bmad-output/implementation-artifacts/epic-16-decision-log.md'
  - '{project-root}/_bmad-output/implementation-artifacts/8-6-picking-a-family-puts-it-in-the-file.md'
warnings: ['oversized']
deferred:
  - summary: >-
      The licence tie has ONE door where its `fvar` sibling has two, so a hand-authored `.folio`
      embedding a face whose own bytes name a contradicted or copyleft licence loads, renders and
      exports unchecked.
    evidence: |-
      `RefuseVariableFace` is called both from `embedFontFamily` and from `fontset.New` at render,
      expressly because "a hand-written `.folio` bypasses this command entirely"
      (component_commands.go:2404-2408). `RefuseContradictedLicence` has no caller outside
      `embedFontFamily`, and `requireEmbeddedFaceLicence` at load checks only that the three licence
      fields are non-empty and bounded, never that they agree with the bytes. Deferred rather than
      fixed because the intent contract names `embedFontFamily` as the siting; widening to the load
      path is a decision, not a patch.
    location: >-
      folio-go/component_commands.go:2443
    severity: medium
  - summary: >-
      The mirror-contract test compares SPDX ids and never the patterns behind them, so the two
      tables can disagree about what a licence sentence looks like with nothing red.
    evidence: |-
      `TestGoLicenceTableSubsumesTheDesignerTable` extracts the TS keys and asserts each appears in
      the Go id set; it never reads the TS `RegExp` values. Narrowing the TS row to
      /Open Font License/i while Go keeps /SIL Open Font License/i leaves the test green, and a face
      whose record 13 omits "SIL" would pass the build gate and be NO EVIDENCE at runtime. The test's
      own comment names this as the failure it means to prevent. In contract as written — the intent
      asks only that the Go table subsume the TS table's POPULATION — so recorded rather than patched.
    location: >-
      folio-go/internal/fontset/licencesignature_test.go:423
    severity: medium
  - summary: >-
      The refuse-signature patterns fire on any mention of a copyleft licence, including a bare URL
      or a compatibility note, and the record-0 widening makes copyright prose their likely home.
    evidence: |-
      `(?i)\b(?:A|L)?GPL(?:[-\s]?v?\d[.\d]*)?\b` matches "gnu.org/licenses/gpl.html",
      "compatible with the GPL" and "not licensed under the GPL". Measured as not firing on the real
      corpus — zero refuse-signature hits across the 99 upstream faces sampled this dispatch — so it
      is a latent over-broadness rather than an observed one, and narrowing it is a decision about
      the table reserved to the engineering lead.
    location: >-
      folio-go/internal/fontset/licencesignature.go:111
    severity: low
  - summary: >-
      `apache/yellowtail` states Apache terms in record 0 in wording the minted Apache regex cannot
      reach, so it is admitted as NO EVIDENCE — the nearest thing to a table gap on the sample.
    evidence: |-
      Its record 0 reads "...Available under the Apache 2.0 licence. http://www.apache.org/licenses/L",
      which is semantically Apache and invisible to /Apache License,?\s+Version 2\.0/i. It ADMITS,
      so it is not a refusal and not a blocking finding under D-16.R.7's ship criterion; widening the
      regex is a change to the table, i.e. a decision, not a fix. Recorded so the lead sees it.
    location: >-
      folio-go/internal/fontset/licencesignature.go:96
    severity: low
---

## In plain terms (read this first if you just want the gist)

*Owner summary; the intent contract below governs implementation.*

A typeface put into a document carries a statement about whose it is and on what terms it may be passed
on. Until now that statement was checked when the product was built, against twenty-one typefaces
somebody had looked at. Once typefaces start arriving from the web, the same statement has to be checked
on the author's machine, against files nobody has reviewed.

This story adds that check where it cannot be bypassed: in the engine, on the bytes themselves, at the
moment a typeface is put into a document. It asks the file what *it* says about its own licence, and
compares that with the licence being claimed on its behalf.

The important part is what it does when the file says nothing. It admits it. The problem this guards
against is a typeface travelling under **someone else's** terms — a false statement, not a missing one —
and that really happened here once, to seventeen of twenty-one typefaces, unnoticed until review.
Refusing files that are merely silent would have blocked about a sixth of the library while catching
none of that. So: if the file contradicts the claim, it is refused and told why. If it agrees, or says
nothing readable, it is accepted.

It also gains something the old check never had: a file whose own bytes name a share-alike or GPL-family
licence is refused outright, whatever anyone claims on its behalf.

Done looks like: no document can be written carrying a typeface whose own bytes contradict the terms
recorded beside it.

<intent-contract>

## Intent

**Problem:** Epic 16 moves the licence admission that D-8.5.2/D-8.5.3 made a **build gate** into a
**runtime** decision over unreviewed bytes. The build gate's teeth are the **nameID 13 tie** at
`folio-designer/src/font-catalogue.test.ts:355-366`, which binds a face's declared SPDX id to the
binary's own licence description — *"the one statement of a face's licence that cannot be edited from
outside the binary."* Nothing carries that tie to runtime, so without this story Epic 16 replaces a gate
with something **strictly weaker**, on exactly D-8.6.5's axis: 17 of 21 catalogue faces once shipped
under another project's licence, green, until a review caught it.

**Approach:** One byte-taking door in `internal/fontset`, beside `RefuseVariableFace`, called from
`embedFontFamily`. It **re-reads the name table from the bytes in hand** and returns one of three
outcomes (D-16.R.7): **contradiction refuses, confirmation admits, no evidence admits.** It also carries
a **refuse-signature** half checked against every face regardless of what it declares, closing an AD-26
hole the build gate never had.

**Sequence: this story runs BEFORE Story 16.1** (D-16.R.8). It needs no browser and no fetched face —
it tests against the 21 committed faces and its own fixtures — while 16.1 is the first story that puts a
**fetched** face into a document. Building the gate after the population it polices arrives is how
D-8.6.5 shipped green.

## Boundaries & Constraints

**Always:**
- **Three outcomes, never two** (D-16.R.7), mirroring D-16.R.4's token table one level up:
  - **CONTRADICTION** — the name table matches a **refuse-signature**, or matches a **different**
    admitted licence's signature than the one declared → **REFUSE**, located, naming both what was
    declared and what the bytes say.
  - **CONFIRMATION** — matches the declared licence's signature → admit.
  - **NO EVIDENCE** — no signature matches, or the name table is absent or unparseable → **admit.**
- **Silence admits, and the reasoning is recorded in the code**, because it will look like a hole to the
  next reader: the threat is a face travelling under **another project's** terms, which is a *false*
  statement, not a missing one. Refusing silence catches none of that and costs ~17% of the library —
  **measured**: 50 of 100 sampled upstream faces refused under the original contract, and all three
  static `ufl/` families carry **no nameID 13 at all**, stating their terms in nameID 0.
- **The table is TWO-SIDED.** Admit-signatures keyed by SPDX id: `OFL-1.1` → `/SIL Open Font License/i`;
  `Ubuntu-font-1.0` → `/Ubuntu Font Licence/i`; `Apache-2.0` → `/Apache License,?\s+Version 2\.0/i`.
  **Refuse-signatures, checked against every face whatever it declares:** GPL, LGPL, AGPL, SSPL, and
  CC BY-SA / ShareAlike. Ground: **AD-26 Binds: all**, and its Prevents is copyleft arriving through a
  plausible-looking package. **A face whose own bytes name a copyleft licence is refused even when the
  declared licence is `OFL-1.1`.**
- **nameID 0 is consulted ONLY when nameID 13 is ABSENT.** If 13 is present it alone decides —
  otherwise a face whose 13 says GPL could be laundered by a permissive-sounding 0, defeating the
  contradiction check with the very thing it exists to catch.
- **Go re-reads the bytes. It may NEVER compare the wire `copyright` against the wire `licence`.** Both
  arrive from the browser, so a check over them proves nothing: both sides would move together. The tie
  reads the name table of the bytes in hand.
- **Signature matching is order-deterministic** — an ordered slice, never a map range (AD-1).
- **The Go and TS tables are a MIRROR CONTRACT, enforced by a test and not by a comment.** The Go table
  is the authority for documents and strictly subsumes the TS build-time table's population; something
  must fail when the TS table admits an id the Go table refuses.
- **Untyped `fmt.Errorf`, per `variableface.go`.** No new error type in package `folio` root —
  `TestFolioMethodNamesAreInjective` forbids it.
- **Both guards kept.** The build-time tie in `font-catalogue.test.ts` is not deleted or weakened
  because Go now checks it. **Both, or halt.**
- Every recorded measurement carries **command, commit, tree state and working directory** (D-8.4j.8).
- Commit only on `main`. Never push, never branch, never `git add -A`.

**Block If:**
- **The guard would ship without being observed to fire.** A guard that never refuses anything on the
  corpus has shipped vacuous. It requires a **positive control**: a fixture face under
  `folio-go/testdata/` whose name table contradicts its declared id, asserted refused.
- **Any refusal on the sample is a SILENCE rather than a CONTRADICTION.** That is the ship criterion,
  and it is semantic rather than a number, because a number here is a licence to tune. A refusal
  traceable to *"no signature matched"* is a **table gap** and returns to the engineering lead as a
  finding; one traceable to *"the bytes say something else"* is the guard working.
- **The check would be sited where it can be bypassed.** The browser is not the only door;
  `embedFontFamily` is reachable from `wasm.Engine.Apply` and from a hand-authored command.
- **`SupportedMajor` would move**, or any of the 23 golden digests moves.

**Never:** a check over the wire fields · a map range over signatures · softening a refusal to a
warning · deleting the build-time tie · a new error type in package `folio` root.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Declared `OFL-1.1`, nameID 13 carries the SIL sentence | a shipped catalogue face | **CONFIRMATION** → admit | — |
| Declared `OFL-1.1`, nameID 13 carries the Apache sentence | mislabelled face | **CONTRADICTION** → refuse, naming both | Located refusal |
| Declared anything, nameID 13 names GPL/AGPL/SSPL or ShareAlike | copyleft face | **REFUSE**, regardless of the declared id | Located refusal |
| No nameID 13, nameID 0 carries the Ubuntu sentence | the three static `ufl/` families | nameID 0 consulted; **CONFIRMATION** → admit | — |
| nameID 13 present and mismatched, nameID 0 permissive | laundering attempt | **REFUSE** — 0 is not consulted when 13 is present | Located refusal |
| No name table at all, or unparseable | face the parser cannot read | **NO EVIDENCE** → admit; `checkSfnt`/`DecodeFontForRender` refuses genuinely bad bytes one step later | — |
| Declared id with no signature entry | an admitted SPDX id the table does not cover | **NO EVIDENCE** → admit, and the gap is reported as a finding | Finding, not a refusal |
| Sentence in a non-Latin script | `ofl/wdxllubrifonttc` states OFL 1.1 in Traditional Chinese | **NO EVIDENCE** → admit. No ASCII regex reaches under this floor | — |
| Local-tier face | one of the 21 committed | Passes unchanged; the build-time tie still covers it | — |

</intent-contract>

## Code Map

**Every anchor below was re-measured at `ceb5213`, the dispatch HEAD.** The spec's `baseline_commit`
is `384c8ac`; the single commit between them (`ceb5213`) touches **only `_bmad-output/**` documents —
no file under `folio-go/`, `folio-designer/` or `lint/` moved.** Verified with
`git diff --stat 384c8ac..ceb5213`. All five anchors carried in the previous draft are therefore
**correct as written**; the entries below add what the plan-gate investigation found beside them.

### The siting (confirmed available)

- `folio-go/internal/fontset/variableface.go:69` — `func RefuseVariableFace(name string, data []byte) error`.
  **Anchor correct.** The sibling to copy: byte-taking, untyped `fmt.Errorf`, `nil` on unparsable.
- `folio-go/component_commands.go:2414` — `if verr := fontset.RefuseVariableFace(name, decoded); verr != nil`.
  **Anchor correct.** The new call goes beside it. `component_commands.go` (package `folio`) **already
  imports `internal/fontset`**, so the import graph needs no change.
- `folio-go/component_commands.go:2360-2466` — `embedFontFamily` in full (`func` at 2360, closing brace
  at 2466). **Anchor correct.** `componentFields(raw, 12)` confirmed inside it — arity frozen at 12.
- `folio-go/component_commands.go:668` — `const maxComponentAssetBytes = (engineProtocolMaxPayloadBytes
  - maxComponentAssetPayloadOverheadBytes) * 3 / 4`. **Anchor correct**, and the value is
  **6,288,384** — verified arithmetically from `:638` (`8 * 1024 * 1024`) and `:649` (`4 * 1024`).
  Note `component_commands_test.go:2207` records that the literal `6288384` **appears nowhere in code**
  and must not be introduced.
- **No architecture test forbids the siting.** `TestFolioMethodNamesAreInjective`
  (`folio-go/render_arch_test.go:461`) constrains package `folio` **root**; the new door and its error
  are born in package `fontset`, so the constraint is satisfied by siting rather than by care. The
  untyped-`fmt.Errorf` requirement stands anyway, per `variableface.go`.

### The name-table reader — it does NOT need to be written from scratch

This was an open question at dispatch. Answer: **a generic nameID reader already exists in the vendored
package, and `internal/fontset` already has the in-package precedent for calling it.**

- `github.com/boxesandglue/textshape@v0.0.15/ot/metrics.go:280` — **`func (n *Name) Get(nameID uint16) string`**.
  A generic, arbitrary-nameID accessor over `Name.entries map[uint16]string`. **This is the reader.**
  nameID 13 and nameID 0 are both reachable through it; nothing new must be parsed by hand.
- `…/ot/metrics.go:213` — `func ParseName(data []byte) (*Name, error)`.
- `folio-go/internal/fontset/fontset.go:502` — `readPostScriptName(font *ot.Font) (string, error)`: the
  **in-package precedent**, and the shape to copy verbatim —
  `font.HasTable(ot.TagName)` → `font.TableData(ot.TagName)` → `ot.ParseName(data)`. Its guard order is
  what makes "absent" distinguishable from "unparseable".
- **The vendor boundary permits this.** `folio-go/internal/fontset/vendorboundary_test.go:639`
  (`TestFamilyNameHasNoCallSite`) is an **AST guard over one named accessor only** — `FamilyName`. It
  does not close a census over all vendor accessors, so **`Get` has no call-site prohibition.**
  `vendor-boundary.md` contains no row for `Get` (grep: no match).
  **Do not call `(*ot.Name).FamilyName()` anywhere**, including in tests: the guard walks the whole
  module, skips `testdata` by category, and matches selectors in call position.
- `folio-go/shipped_faces_test.go:196` — `sfntNameRecords(...) map[uint16]string`, a direct
  name-table parse (prefers platform 3) already in the repo, available as a **test-side** cross-check
  independent of the vendor accessor.
- `folio-designer/src/font-catalogue.test.ts:126` — `nameTableString(view, tables, 13)`, the TS side's
  equivalent generic reader. Useful as a semantic reference for what "the same record" means.

### The controls — both already exist as committed bytes (measured at this gate)

The contract requires a positive control "under `folio-go/testdata/`". **It is already there, and no new
font binary need be committed.** Measured at the plan gate by parsing the committed files directly
(pure Python, read-only, nothing written to the repository; command, commit `ceb5213`, clean tree, cwd
`/Users/panitw/Projects/folio`):

| File | nameID 13 | nameID 0 |
|---|---|---|
| `folio-go/testdata/fonts/Roboto-Regular.ttf` | `Licensed under the Apache License, Version 2.0` | `Copyright 2011 Google Inc. All Rights Reserved.` |
| `folio-go/testdata/fonts/notosansthai-variable-testonly/NotoSansThai-VF.ttf` | `This Font Software is licensed under the SIL Open Font License, Version 1.1…` | `Copyright 2022 The Noto Project Authors…` |

- **POSITIVE CONTROL (contradiction), free:** `Roboto-Regular.ttf` is recorded as **`Apache-2.0`**
  (`lint/internal/licence/licencecensus_test.go:97`) and its nameID 13 carries the **Apache** sentence.
  Embedding it while **declaring `OFL-1.1`** is a genuine CONTRADICTION — the bytes match a *different
  admitted licence's* signature than the one declared — and must be REFUSED. This is exactly the
  fixture the ruling asks for, with no new binary, no new `LICENSE*`/`NOTICE*`, and no new row in the
  licence census.
- **CONFIRMATION control, free:** `NotoSansThai-VF.ttf` is recorded as `OFL-1.1`
  (`licencecensus_test.go:98`) and its nameID 13 carries the SIL sentence → admit. It is also the
  face `RefuseVariableFace` rejects, so **order matters**: assert the licence outcome through the new
  door directly, not through `embedFontFamily`, where the `fvar` refusal fires first.
- **Both minted regexes verified against these real strings at this gate:**
  `/Apache License,?\s+Version 2\.0/i` matches Roboto's record; `/SIL Open Font License/i` matches
  Noto's. Neither is speculative.
- **If a synthetic fixture is nonetheless wanted**, the in-repo precedent is
  `folio-go/internal/fontset/fontset_test.go:36` `patchUnitsPerEm` — copies the sfnt bytes and walks the
  16-byte table directory to overwrite a field in place. **Caveat, and the reason the free control is
  preferred:** `name` records are variable-length behind a string-storage pool, so only a *same-length*
  substitution is as safe as `patchUnitsPerEm`; and any **new committed binary** is swept by
  `lint/internal/manifest/manifest.go:259` (which special-cases `folio-go/testdata/fonts`) and by
  `folio-designer/src/font-binary-identity.test.ts`, and would need its own licence-census row.

### The mirror contract — constructible today, with precedent, and no shared artifact needed

This was flagged at dispatch as possibly unimplementable. **It is implementable, and the repo has a
line-for-line template.**

- `folio-designer/src/font-catalogue.test.ts:197-200` — `const licenceSignatures: Readonly<Record<string, RegExp>>`
  with exactly two rows, `'OFL-1.1'` and `'Ubuntu-font-1.0'`. **Anchor correct.** The file exports
  **nothing**, so the table cannot be imported — but it does not need to be.
- **Mechanism (the established idiom): a Go test reads the TypeScript source as text and regex-extracts
  the keys.** Template: `folio-go/canvas_projection_wire_test.go:353-400`, which `os.ReadFile`s
  `folio-designer/src/engine-protocol.ts` and pulls key lists out with package-level
  `regexp.MustCompile` extractors **anchored on the TS declaration's own identifier**, then compares
  sets. Two rules to copy exactly: **anchor the regex on `const licenceSignatures`**, and **`t.Fatal`
  when the file or the anchor does not match — never skip** (`:376-379`).
- Doctrine to cite: `folio-designer/src/engine-bounds-mirror.test.ts:6-40` states **D-7.4.5 / DW-25** —
  any invariant duplicated across the Go/TS boundary moves in ONE commit, with a test that reads both
  sides. Its `sites` regexes are the anti-vacuity trick: assert the mirrored key is *consumed*, not
  merely declared.
- **A shared generated artifact is NOT required and must not be introduced.** The only generated Go/TS
  bridge, `folio-designer/src/generated/font-catalogue.ts`, is **gitignored** (`.gitignore:71`) and
  TS-only. Reading source text needs no export and no build step.
- **Subsumption holds today, so the test will pass on landing and is not vacuous:** TS declares
  {`OFL-1.1`, `Ubuntu-font-1.0`}; Go will declare {`OFL-1.1`, `Ubuntu-font-1.0`, `Apache-2.0`}. TS ⊂ Go,
  strictly. Red-prove it by deleting the `Apache-2.0` row's counterpart, or by adding a TS row with no
  Go entry.

### Read-only evidence and counts (confirmed, do not move)

- `folio-designer/font-catalogue.json` — **21 faces**; distinct `licence` values are exactly two:
  `OFL-1.1` (19) and `Ubuntu-font-1.0` (2). **Not read by Go at all** (`grep font-catalogue` over
  `*.go`: no match) — so this story must not make Go depend on it.
- `folio-designer/public/fonts/` holds **27** directories, not 21: the extra six are the hand-declared
  chrome/engine families at `font-catalogue.test.ts:59`. **Never derive the guard's population by
  walking that directory.**
- `folio-go/byte_neutrality_test.go:100` — `goldenDigestRecord`, **23 entries**, with the length guard
  at `:634` and `TestGoldenDigestAgreesAtEveryDeclaredSite` at `:649`. The "23 golden digests" the
  contract requires unmoved.
- `folio-go/internal/template/version.go:77` — `SupportedMajor = 2`.
- `lint/internal/manifest/manifest.go:140` — `fontAssetLicenceAllowlist = []string{"OFL-1.1",
  "Apache-2.0", "MIT", "Ubuntu-font-1.0"}` (D-8.5.3). The owner's four; the Go signature table covers
  three of them by design (see Design Notes).
- `folio-go/internal/template/fontasset.go` — `DecodeFontForRender`/`checkSfnt`, the step that refuses
  genuinely unreadable bytes one later, which is what makes NO EVIDENCE safe.

## Tasks & Acceptance

**Execution:**

1. `folio-go/internal/fontset/` — a new byte-taking door beside `RefuseVariableFace` (new file, e.g.
   `licencesignature.go`). It parses the name table exactly as `readPostScriptName` does
   (`HasTable` → `TableData` → `ot.ParseName`), reads nameID 13 via `(*ot.Name).Get(13)`, falls back to
   `Get(0)` **only when 13 is absent**, applies the **refuse-signatures to every face first**, then the
   **admit-signature for the declared id**, and returns contradiction / confirmation / no-evidence.
   **Ordered slice, never a map range** (AD-1). Untyped `fmt.Errorf`. Return `nil` (admit) on every
   parse failure, per `RefuseVariableFace`'s precedent.
2. **Define "present" explicitly in code**, because the whole nameID-0 door hangs on it: a record is
   PRESENT when the parsed name table yields a **non-empty** string for it. Record the known fidelity
   limit — see Design Notes.
3. `folio-go/component_commands.go` — call it from `embedFontFamily` **beside the `fvar` refusal at
   `:2414`**, before any byte reaches `t.doc.Assets`. Keep `component_commands.go` a caller, not a
   checker.
4. Tests in `internal/fontset`, using the **committed controls** rather than a new binary:
   - CONTRADICTION: `testdata/fonts/Roboto-Regular.ttf` declared as `OFL-1.1` → refused, and the
     message names **both** what was declared and what the bytes say.
   - CONFIRMATION: `Roboto-Regular.ttf` declared as `Apache-2.0` → admitted (the guard is not
     over-broad); `NotoSansThai-VF.ttf` declared as `OFL-1.1` → admitted.
   - COPYLEFT: a face whose record names GPL/AGPL/SSPL/ShareAlike, declared `OFL-1.1` → refused.
     Build it by same-length substitution per `patchUnitsPerEm`'s directory-walking method, in-test.
   - NO EVIDENCE: name table stripped → admitted; record present but matching nothing → admitted.
   - nameID 0 consulted on absence of 13; **not** consulted when 13 is present and mismatched.
   - **Red-prove by DELETING the guard**, never by falsifying a condition — deletion is what proves the
     call site is reached (a falsified condition only proves arm order).
5. A Go test enforcing the **mirror contract** by reading `font-catalogue.test.ts` as source text,
   anchored on `const licenceSignatures`, `t.Fatal` on a miss. Include the vacuity guard: fail if the
   extracted key set is empty.
6. **Re-run the 100-face sample** and report it with command, commit, tree state and working directory
   (D-8.4j.8). Every refusal must be a CONTRADICTION and none a silence.
7. `deferred-work.md` — **close DW-150 as reconciled**, recording that the lead's contract was what was
   out of step, not `RefuseVariableFace`.
8. Record the narrowing on **D-000.15's running list** for Story 15.3 (before-the-tag set).
9. **Do not** add a row to `vendor-boundary.md` for `Get` unless the implementation finds a
   substitution to declare; `Get` returns the zero value, which is observably absent, so there is
   nothing to substitute. If a row is added, `TestVendorBoundaryDocumentExistsAndIsCited` must stay green.

**Acceptance Criteria:**

- Given a face whose name table contradicts its declared licence, when it is embedded, then the command
  refuses it, located, naming both what was declared and what the bytes say — and no byte reaches
  `Assets`.
- Given a face whose own bytes name a GPL-family or ShareAlike licence, when it is embedded, then it is
  refused **whatever** licence is declared for it.
- Given a face with no readable licence statement, when it is embedded, then it is **admitted**, and the
  reasoning is stated in the code so a later reader does not read it as a hole.
- Given a face with no nameID 13 but a licence statement in nameID 0, when it is embedded, then nameID 0
  is consulted; and given one with nameID 13 present, then nameID 0 is not consulted at all.
- Given the guard, when it is proposed for merge, then a positive-control fixture proves it fires, and
  the re-run sample shows every refusal is a contradiction and none a silence.
- Given the TS build-time table, when it admits an id the Go table refuses, then a test fails.
- Given the build-time tie, when this story lands, then it is unchanged and still green.

## Design Notes

**Why silence admits, written here because it is the part that will look wrong.** The guard exists for
D-8.6.5 — a face travelling under another project's terms. That is a **false** statement. A face that
says nothing has made no statement to be false, and the decode path refuses it a step later if the bytes
are actually bad. The original contract refused silence and was measured at **50 of 100 refused**, of
which the great majority were silences: a guard that loud is one somebody eventually turns off.

**Why the refuse-signatures apply to every face.** AD-26's Binds is *all*, and its Prevents is copyleft
arriving through a plausible-looking package. A declared-id-only check would let a GPL face in under an
`OFL` token. This half is new — the build gate never had it.

**Why the nameID 0 widening cannot be a general fallback.** If 0 were consulted whenever 13 failed to
match, a face whose 13 says GPL could be laundered by a permissive 0. Absence is a different condition
from disagreement, and only absence opens the second door.

**A declared id with NO signature entry admits, and D-16.R.4 selects that reading rather than leaving it
open.** D-16.R.5 had said "an SPDX id with no signature entry is a refusal, not a skip"; D-16.R.7's
three-outcome table supersedes it — *no signature matches* is NO EVIDENCE, which admits. The
reachable case decides it: D-8.5.3 admits **`MIT`**, and D-16.R.4 rules that **"MIT stays admissible and
gets no table entry"** because `google/fonts` publishes no MIT token — *"absence, not narrowing."*
Refusing an id with no entry would therefore refuse a licence the owner has explicitly admitted. The
Go table covers three of the owner's four ids **by design**, and the fourth's absence is the ruling.

**Known fidelity limit of the vendor reader, recorded rather than discovered in review.**
`ot.ParseName` (`metrics.go:213`) keeps a record only when it decodes to a non-empty string, and it
decodes only platform 0/3 (UTF-16BE) and platform 1 encoding 0 (Mac Roman); anything else is skipped,
and `Get` returns `""` for both "no such record" and "record skipped". So a face stating its terms
*only* under an exotic platform/encoding reads as **absent**, which opens the nameID 0 door one step
early. This can only make the check **quieter**, never louder — it cannot produce a false refusal — and
D-16.R.7's own "how we'd know it was wrong" accepts exactly that cost: *"a check quieter than intended,
never a document publishing false terms."* Do not "fix" this by hand-parsing the name table; state it.

**Why the controls are the committed faces and not a new binary.** A new committed `.ttf` under
`folio-go/testdata/fonts` is swept by the manifest guard (`manifest.go:259`), the binary-identity guard,
and the licence census, and would need its own `LICENSE*`/`NOTICE*` — for a file whose entire purpose is
to carry a licence statement that is a lie about itself. The already-committed Apache-licensed Roboto
supplies a *true* contradiction against an `OFL-1.1` declaration with none of that cost.

**Order at the call site.** `RefuseVariableFace` runs first at `:2414`. `NotoSansThai-VF.ttf` is
variable, so a licence assertion routed through `embedFontFamily` with that face would be masked by the
`fvar` refusal. Assert the three outcomes against the new door directly; assert *reachability* from
`embedFontFamily` with a static face.

## Verification

Run from the repository root unless stated.

- `cd folio-go && go test ./... && go vet ./... && gofmt -l .`
  **Note the `gofmt` path is `.`, not `folio-go`** — the previous draft's `gofmt -l folio-go` names a
  directory that does not exist from inside `folio-go` and would silently list nothing.
- **Baseline reds, pre-declared so they are not attributed to this story.** Both were verified red at
  baseline by Story 16.0's close; if either changes shape, that IS this story's problem:
  - `folio-go/internal/text` — the two failures of the **mandated P6g exercise floor**
    (`TestCorpusMeetsP6ExerciseFloors`). Permanent; never "fixed".
  - `folio-designer` — `npm test` is red at `canvas-authority-contract.test.ts:190` over
    `e2e/e9-5-border-no-ink.spec.ts`, **registered as DW-152** and owned by the Epic 9/10 lane.
- `cd folio-designer && npm run test && npm run build` — expect exactly the DW-152 red above.
- **The positive control reds when the guard is deleted** (deletion, not falsification), and the guard
  admits `Roboto-Regular.ttf` declared `Apache-2.0`.
- **The mirror-contract test** fails when a row is added to the TS table with no Go counterpart.
- **The re-run 100-face sample**, reported with command, commit, tree state and working directory.
- **The 23 golden digests unmoved** — `goldenDigestRecord` at `byte_neutrality_test.go:100`, its length
  guard at `:634`, and `TestGoldenDigestAgreesAtEveryDeclaredSite`.
- `SupportedMajor` still `2` at `folio-go/internal/template/version.go:77`.
- `font-catalogue.test.ts:197-200` and `:355-366` unchanged — **both guards kept, or halt.**
- Heavy suites at the **end-of-run catch-up** per D-16.R.1. **No in-story browser run**: this story has
  no browser surface. e2e specs are **compile-only**.

## Spec Change Log

**2026-09-03 — plan-gate re-plan at `ceb5213` (halt after planning).**

- The intent-contract block was **preserved byte-identically** (7,019 bytes, md5
  `f1be40fb96e326f3c567909f1f66def6`, verified before and after). No scope amendment was supplied, so
  step-02's preservation rule applies in full. Everything changed is outside the block.
- Spec `status` was `ready-for-dev`, which step-01 routes straight to IMPLEMENT. Set to `draft` for the
  dispatch and restored to `ready-for-dev` at this gate, so the dispatch could re-plan rather than build.
- **All five Code Map anchors re-measured at the dispatch HEAD and found CORRECT.** The only commit
  between `baseline_commit` `384c8ac` and `ceb5213` touches `_bmad-output/**` only.
- Answered the three questions the plan gate was asked to settle, and recorded the evidence:
  **the reader exists** (`(*ot.Name).Get`, plus the `readPostScriptName` precedent); **the siting is
  available** (import already present, no arch test forbids it); **the mirror contract is
  constructible** with an existing template and no shared generated artifact.
- Replaced the assumed new `testdata` fixture with **two already-committed controls**, after measuring
  their name tables at this gate. Removes a committed-binary change and its licence-census cost.
- Corrected the `gofmt` invocation, which named a non-existent directory.
- Pre-declared the two baseline reds (P6g floor; DW-152) so the implementer does not chase them.
- Recorded the vendor reader's fidelity limit and the no-signature-entry resolution, both of which a
  reviewer would otherwise raise as findings.
- `warnings:` gained **`oversized`**. Honest accounting: the spec **grew** this dispatch, from ~3.7k to
  ~7.6k tokens, because the investigation was drained into the Code Map as step-02 requires. It was
  already over the 1,600-token threshold before that and could not have cleared it. For scale, Story
  16.1 was ~14.5k tokens before the D-16.R.8 cut, so this remains roughly half its size. **Do not thin
  the acceptance criteria to chase the flag.** If it must shrink, the lever is moving the Code Map's
  three evidence subsections — the controls table, the mirror-contract mechanism, and the read-only
  counts — into a companion artifact cited from `context:`.

## The falsifier, re-run

**D-16.R.7's ship criterion is semantic, not a number: every refusal must be a CONTRADICTION and no
refusal a silence.** It is met.

**Measurement provenance (D-8.4j.8).** Command: a pure-Python script in the session scratchpad
(`sample.py`), no `fontTools`, fetching `METADATA.pb` and each family's `style:"normal" weight:400`
TTF from `raw.githubusercontent.com/google/fonts/main` into memory and parsing the sfnt `name` table
by hand; family directory listings taken with `gh api repos/google/fonts/contents/{apache,ofl,ufl}`
because the unauthenticated contents API rate-limits immediately. Its classifier is a line-for-line
transcription of `licencesignature.go`'s two tables and its three-outcome order. Run 2026-09-03 at
commit `e250dc7` with the working tree carrying this story's changes only, working directory
`/Users/panitw/Projects/folio`. **Nothing was written to the repository**; the corpus lives in the
scratchpad and is not committed.

**Population.** 135 families attempted — all 44 `apache/`, all 5 `ufl/`, and an A–Z spread of 86
`ofl/` families taken evenly across the 1,000 the contents API will enumerate. **99 had a usable
static Regular and parsed**; 33 were variable-only and skipped; 3 could not be fetched; **zero sfnt
parse errors**. Declared ids: `OFL-1.1` 55, `Apache-2.0` 41, `Ubuntu-font-1.0` 3.

| Outcome | n | |
|---|---|---|
| **CONFIRMATION** — admitted | 90 | |
| **NO EVIDENCE** — admitted | 8 | all "no signature matched"; none a refusal |
| **CONTRADICTION** — REFUSED | **1** | `apache/mountainsofchristmas` |

**The one refusal is the guard working, and it is the same face Story 16.1's gate found.**
`apache/mountainsofchristmas` declares `APACHE2` in `METADATA.pb` while its own record 13 states *"This
Font Software is licensed under the SIL Open Font License, Version 1.1"*. Under the original contract
this face was caught only **incidentally**, as a signature miss lost among 49 others; under D-16.R.7 it
is **diagnosed** — the bytes match a different admitted licence's signature than the one declared — and
the refusal names both sides. **Refusal rate 1.0%, against 50.0% under the contract D-16.R.7 replaced.**

**No refusal is a silence.** All 8 NO EVIDENCE faces are ADMITTED, so none of them is a table gap
reportable as a refusal; they are reported here as observations rather than findings:

- Six `apache/jsmath*` faces and `apache/yellowtail` carry **no record 13**, so record 0 was consulted
  and matched nothing: the jsMath faces' copyright reads *"Generated from MetaFont bitmap by
  mftrace…"*, which names no licence at all. `ofl/arvo`'s record 0 is a bare copyright line.
- **`apache/yellowtail` is the nearest thing to a table gap and it is worth naming**: its record 0
  reads *"…Available under the Apache 2.0 licence. http://www.apache.org/licenses/L…"* — semantically
  Apache, and invisible to `/Apache License,?\s+Version 2\.0/i`. It **admits**, which is D-16.R.7's
  accepted cost ("a check quieter than intended, never a document publishing false terms"), and
  widening the regex to reach it is a change to the table, i.e. a decision, not a fix made here.

**Two rows that were dead against the real corpus are now alive.**

- **`Apache-2.0`**: 41 of 99 faces declare it and 40 of them CONFIRM against the minted signature.
  Under the pre-D-16.R.7 table — which had no Apache row — every one of them was refused.
- **`Ubuntu-font-1.0` through record 0**: all three static `ufl/` families (`ubuntu`,
  `ubuntucondensed`, `ubuntumono`) carry **no record 13** and state their terms in record 0, and all
  three CONFIRM. The record-0 widening is what makes that row fire positively at all; without it every
  genuine Ubuntu family was refused, and the row had never once been observed to confirm anything.
  `ofl/lohittamil` also confirms through record 0.

## Review Triage Log

### 2026-09-03 — Review pass

- intent_gap: 0
- bad_spec: 0
- patch: 11: (high 0, medium 3, low 8)
- defer: 4: (high 0, medium 2, low 2)
- reject: 12
- addressed_findings:
  - `[medium]` `[patch]` The OFL admit row — the row 19 of the 21 catalogue faces and the OFL
    majority of upstream travel under — had no CONTRADICTION control: every assertion that a refusal
    HAPPENS was driven by the Apache or Ubuntu row, and "admits" is also what a row matching nothing
    returns. Confirmed by demonstration: narrowing the OFL pattern to an anchored prefix left the
    entire `internal/fontset` suite green while breaking real faces such as `cascadiacode`, whose
    record 13 opens "Microsoft supplied font…" and carries the OFL sentence further in. Added
    `TestRefuseContradictedLicenceRefusesWhatTheOFLRowContradicts` with two arms — the committed OFL
    fixture declared `Apache-2.0`, and the same statement behind a prefix. Re-verified independently:
    the narrowing now reds.
  - `[medium]` `[patch]` The refuse half's over-broadness control asserted almost nothing:
    `err != nil && strings.Contains(err.Error(), "copyleft")` passes when the error is a contradiction
    and goes fully vacuous if the refusal prose ever drops the word "copyleft". Rewritten so each
    permissive sentence is declared the id its own bytes name and must return `nil`, consulting no
    substring of the prose. Probed by widening a refuse row to match "License": reds on all three.
  - `[medium]` `[patch]` Refusal messages could exceed the host's own cut and lose their most
    informative clause. `maxComponentFailureMessageBytes` is 512 and the host truncates there;
    measured, the contradiction message was 378 bytes and the copyleft message 420 with only a
    46-byte statement, so a face whose record 13 carries the full licence body (common) had the
    trailing `The face says: %q` clause silently cut off. Added a 72-byte rune-boundary excerpt with
    a marked elision, and a test that builds ~5.9 kB statements and asserts the message stays within
    512 with both sides still named. Probed three ways: excerpt removed (6211/6302 bytes), rune walk
    removed (a split UTF-8 sequence), host constant moved (the source-read tie reds).
  - `[low]` `[patch]` The mirror-contract extraction could truncate silently without tripping its own
    vacuity guard — the non-greedy regex stops at the first column-0 `}`, so a nested object drops
    every later id while still returning a non-empty list. Now asserts the extracted set contains the
    ids the TS table declares today, with a message saying a miss means the extraction rotted rather
    than that the table shrank. Probed with a nested literal; TS file restored byte-exact.
  - `[low]` `[patch]` `matched[0].label` reintroduced the order-dependence the comment above it
    rejects — the scan visits every admit row so table order does not become policy, then named only
    the first match. Now names every matched label.
  - `[low]` `[patch]` The I/O matrix's non-Latin row (`ofl/wdxllubrifonttc`, which states OFL 1.1 in
    Traditional Chinese) was unexercised. Added that case with a genuine Traditional Chinese sentence,
    round-tripped through the name table and admitted under two different declarations.
  - `[low]` `[patch]` Two comment overclaims in `licencesignature.go`: (a) `ReadLicenceStatement`
    promised a distinction — "they admit for different reasons and a later reader is owed the
    difference" — that its `("", false)` return cannot express; now states the three failures are
    deliberately collapsed. (b) The fidelity note framed the vendor reader's record-dropping as only
    ever making the check quieter; it now names the case it does not cover — `ot.Name.entries` is
    keyed by nameID alone and each decodable record overwrites the last, so the FINAL record 13 in
    table order decides. Recorded as a known limit; no second parser, per the spec's prohibition.
  - `[low]` `[patch]` `buildNameTable` cast lengths to `uint16` with no overflow check, so a record
    or storage pool over 65535 bytes wrapped into a corrupt table that the assertions above it would
    have passed over. Now fatals on all three overflow shapes; probed with an 80,000-byte record.
  - `[low]` `[patch]` The refuse-beats-confirm precedence had no test over a statement that names
    BOTH, which is the shape real dual-licensed faces ship (OFL plus GPL-with-font-exception). Added
    it; probed by moving the admit half ahead of the refuse half.
  - `[low]` `[patch]` `TestLicenceGuardPreconditions` did not pin the three sentence constants to the
    admit rows they stand for, so a drift between a constant and its pattern could quiet the record-0
    tests. Now pinned by each row's own pattern.
  - `[low]` `[patch]` The `## Auto Run Result` called the falsifier a "100-face re-run" while the
    section it summarises reports 135 attempted and 99 parsed, and the NO EVIDENCE prose left it
    ambiguous whether `ofl/arvo` lacked record 13 or merely failed to match. Corrected below.

**Rejected (12), with the authority each was tested against.** Case- and whitespace-normalising
`declared` (three layers proposed it): a lowercase `ofl-1.1` is not an admitted SPDX id at all, the
resulting refusal names both sides accurately, and the browser path is closed by D-16.R.4's token
table — normalising would mint a canonicalisation policy reserved to that table. An empty `declared`
is unreachable (`requireEmbeddedFaceLicence` requires non-empty). Committing the sample script:
not required, and the spec asks for provenance rather than an artifact. Sample selection bias: already
stated in the spec's own falsifier section. DW-150's status "not updated": factually wrong — it reads
`Status: CLOSED 2026-09-03 — reconciled`. `ot.ParseFont(data, 0)` ignoring later faces of a collection:
the spec directs the parse to copy `readPostScriptName` exactly, which does the same. Brittle
line-number citations: the needle form is already used where it is load-bearing. `removeNameTable`
leaving `searchRange`/`entrySelector` stale: test-only, and the vendor reader does not read them.
The remaining rejects are the two frontmatter-baseline observations, the "one implementation also
covers the local tier" phrasing, and DW-153 lacking a back-reference planted in Story 15.3.

**One reviewer observation is accepted as true and recorded rather than actioned:** the 99-face
corpus run falsifies the ADMIT half only. It produced one refusal and that refusal was a
contradiction; zero faces hit a refuse-signature, so the copyleft floor's only evidence is the
synthesised control D-16.R.9 requires. That is what the ruling anticipated when it ordered a control
for that half specifically, but the sample must not be read as evidence for both halves.

## Auto Run Result

Status: done
Blocking condition: none

**Dispatch:** build of Story 16.1b at HEAD `e250dc7`, 2026-09-03, working directory
`/Users/panitw/Projects/folio`, branch `main`. No branch created, nothing pushed, no `git add -A`.
The `<intent-contract>` block is byte-identical to the version at dispatch (7,019 bytes, md5
`f1be40fb96e326f3c567909f1f66def6`), verified before implementation and again at close.

**What landed.**

- `folio-go/internal/fontset/licencesignature.go` (new) — `ReadLicenceStatement` and
  `RefuseContradictedLicence`, the byte-taking door beside `RefuseVariableFace`: two ORDERED slices
  (AD-1, never a map range), refuse-signatures applied to every face first, then the declared id's
  admit-signature, then the three outcomes. Untyped `fmt.Errorf`, `nil` on every parse failure.
  Record 0 consulted **only** when record 13 is absent; "present" defined in code as a non-empty
  parsed string; the vendor reader's fidelity limit — including which record wins when a nameID
  appears twice — recorded rather than left to be discovered. Refusal messages carry a bounded
  rune-safe excerpt of what the bytes said, so the host's 512-byte cut cannot swallow the clause
  the message exists for.
- `folio-go/component_commands.go` — one call from `embedFontFamily`, beside the `fvar` refusal and
  before anything reaches `t.doc.Assets`. The file stays a caller, not a checker.
- `folio-go/internal/fontset/licencesignature_test.go` (new) — both positive controls D-16.R.9
  requires with no new committed binary: contradiction over committed `Roboto-Regular.ttf` declared
  `OFL-1.1`, and the copyleft half over a name table synthesised in-process across ten
  GPL/LGPL/AGPL/SSPL/ShareAlike statements. Plus the OFL row's own contradiction control, both
  over-broadness controls, the silence arms (no name table, non-font bytes, unrecognised statement,
  non-Latin statement, an id with no signature row), both halves of the record-0 rule, the
  refuse-beats-confirm precedence, the AD-1 ordering guard, the message-bound test, the
  **mirror-contract** test reading `font-catalogue.test.ts` as source text, and a test asserting the
  build-time tie is still present ("both, or halt").
- `folio-go/component_commands_test.go` — `embedCommandDeclaring` (the declared id is load-bearing
  now), the reachability-and-writes-nothing test at the command, and
  `TestStaticFaceIsStillEmbeddedAtBothDoors` corrected to declare Roboto's actual `Apache-2.0` —
  without which it would have been measuring a refusal from a different guard.
- `deferred-work.md` — DW-150 confirmed reconciled at landing; **DW-153** added as the
  before-the-tag ledger entry for Story 15.3 (D-000.15).

**Review findings breakdown.** 11 patches applied (medium 3, low 8), 4 items deferred, 12 rejected;
no `intent_gap` and no `bad_spec`, so no repair loopback ran and `review_loop_iteration` stays 0.
Follow-up review recommended: **true** — 3 × medium + 8 × low scores 17 against a threshold of 5.
Details and the authority each rejection was tested against are in the Review Triage Log above.

**Red-proved by DELETION, never by falsifying a condition, and every one re-run independently of the
implementing agent** (2026-09-03, wd `folio-go`):

| Deletion | Reds |
|---|---|
| the `RefuseContradictedLicence` call in `embedFontFamily` | `TestEmbedFontFamilyRefusesAFaceWhoseOwnBytesContradictTheDeclaredLicence` |
| the guard's body (`return nil` before the tables) | the contradiction, copyleft and record-0 tests |
| a `'BSD-3-Clause'` row in the TS table with no Go counterpart | `TestGoLicenceTableSubsumesTheDesignerTable` |
| renaming `const licenceSignatures` in the TS file | the same test, with `t.Fatal` and not a skip |
| narrowing the OFL row to an anchored prefix | `TestRefuseContradictedLicenceRefusesWhatTheOFLRowContradicts` (added this pass; the suite was green under this mutation before it) |

**Verification, measured this dispatch and re-measured after the review patches.**

- `cd folio-go && go test -count=1 ./...` → **1910 pass / 2 fail / 5 skip**. Both failures are the
  pre-declared mandated P6g exercise floor, `TestCorpusMeetsP6ExerciseFloors` and its
  `P6g_(opaque_names)` subtest (`got 7, need >=20`) — identical in shape and count to the baseline
  measured at `e250dc7` before any change.
- `go vet ./...` clean; `gofmt -l .` printed nothing.
- `cd folio-designer && npm run test` → **1 failed | 436 passed (437)**, files 1 failed | 42 passed.
  The single failure is the pre-declared DW-152 red at `canvas-authority-contract.test.ts:190` over
  `e2e/e9-5-border-no-ink.spec.ts`, owned by the Epic 9/10 lane.
- `npm run build`, including `build:offline` and `verify:offline`, exit 0.
- `cd lint && go test -count=1 ./...` → four `ok` packages, no failures (run with `-count=1`
  deliberately: the rules package walks a directory and Go's test cache does not track `ReadDir`).
- `go vet -tags=matrix ./...` clean; `npm run typecheck` clean; `npm run lint` shows the four
  pre-existing `only-export-components` warnings and nothing new.
- **23 golden digests unmoved** (`goldenDigestRecord` still 23 entries;
  `TestGoldenDigestAgreesAtEveryDeclaredSite` green). `SupportedMajor` still `2`.
  `folio-designer/` has **no diff at all** against the dispatch baseline, so
  `font-catalogue.test.ts:197-200` and `:355-366` are untouched — both guards kept.
- **The one refusal on the sample was verified by hand, independently of the implementing agent.**
  `apache/mountainsofchristmas` was fetched fresh from `google/fonts`, its `METADATA.pb` read
  (`license: "APACHE2"`) and its name table parsed directly: record 13 states *"This Font Software is
  licensed under the SIL Open Font License, Version 1.1"*. Put through the shipped door it refuses
  under `Apache-2.0`, naming both sides, and admits under `OFL-1.1`. A genuine contradiction, not a
  silence.

**The falsifier's result, in the terms D-16.R.7 set.** 135 families attempted, **99 parsed** with zero
sfnt parse errors; 90 CONFIRMATION, 8 NO EVIDENCE (all admitted), **1 CONTRADICTION, refused**.
**Every refusal is a contradiction and none is a silence — the ship criterion is met.** Refusal rate
1.0% against 50.0% under the contract D-16.R.7 replaced. The full breakdown, provenance and the eight
NO EVIDENCE faces are recorded under "The falsifier, re-run" above; note it is a 135-attempt run, not
the "100-face" sample the earlier draft of this section called it.

**Residual risks, stated rather than discovered later.**

- **The corpus run falsifies the ADMIT half only.** No sampled face hit a refuse-signature, so the
  copyleft floor's sole evidence is the synthesised control D-16.R.9 required for exactly that reason.
  The 1-refusal result must not be read as evidence that both halves fire.
- **One door, where the `fvar` sibling has two.** A hand-authored `.folio` embedding a mislabelled
  face is not seen by this guard; the load path checks only that the licence fields are non-empty.
  Deferred, with evidence, in the frontmatter.
- **The mirror contract is a population contract, not a behaviour one** — the two tables' patterns can
  drift with nothing red. Also deferred.
- The refuse patterns would fire on a face that merely mentions a copyleft licence; zero instances
  across the 99 sampled, so latent rather than observed.
- **No browser run**, per D-16.R.8: this story has no browser surface, and its e2e specs are
  compile-only.

**Not attempted, deliberately:** no `.folio` format change, no new committed font binary, no row added
to `vendor-boundary.md` (`Get` substitutes nothing — it returns the zero value, which is observably
absent), no change to the build-time tie, and no softening of any refusal to a warning.
