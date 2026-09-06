---
title: 'Story 11.5: A bold document is a pinned golden'
type: 'feature'
created: '2026-09-06'
status: 'done'
baseline_commit: '4ba9e579676b6df1695d2106c1292ccf76f916d9'
review_loop_iteration: 0
context: []
---

## In plain terms (read this first if you just want the gist)

*Non-normative: a plain-language summary, rewritten after delivery to describe what shipped. The
frozen Intent below governs implementation.*

Until this story, not one document in the pinned test corpus asked for bold or italic type. That was
worse than a coverage gap: the engine could have quietly begun drawing bold text in the ordinary face
and nothing here would have objected — no warning, no failing test, no changed output. The machinery
was guarded; the result was not.

This story pins the result. It adds one page that asks for bold and italic from a typeface genuinely
offering those cuts, renders it identically on all four supported targets, and freezes the resulting
bytes as a reference. To prove the page is a real witness rather than decoration, the resolver was
deliberately broken; of the many tests named for reference documents, exactly one noticed — this new
one.

Then the story stopped and asked a person, because no automated check can confirm a page looks right.
The owner read it and confirmed something narrower than "it rendered": that the bold line is a true
bold typeface, not a thickened ordinary one, and that the sloped lines are drawn italics, not tilted
uprights. That reading is now recorded, and the gate that had been failing on purpose while the story
waited went green when it landed.

Two things not to mistake for defects. The suite still reports two permanent failures — a mandated
floor nobody has met, never to be "fixed". And two weaknesses it found in how attestations are
enforced were left unrepaired on purpose, and registered for an owner rather than quietly fixed.

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** No document in the pinned corpus declares bold or italic, so the outcome of Story 11.2's
face resolution is pinned by nothing. A resolver that silently switched face — the exact failure
DW-237 registers — would fire no Warning, red no test and move no golden. Story 11.2's DW-233
tripwire covers the *mechanism* behaviourally; nothing covers the *outcome* in bytes.

**Approach:** Add one new fixture, `fixtures/declared-variants/`, whose chain entry declares all
three cuts in the object form Story 11.2 introduced, and whose elements exercise all four styles.
Register it as a matrix document and a golden digest, render it byte-identically on all four
targets, and red-prove it by pointing the resolver at the base face. Ship the candidate
`expected.pdf` together with a **red attestation gate**, and **HALT** for the owner rather than
self-attesting (D-11.5.1, arm [A]).

**The attestation obligation is doctrinal, not architectural.** Its descent is **D-000.22 → D-2.3.5**.
AD-21 says nothing about human attestation, and D-4.7.1 is scoped to the statement family — see
D-11.5.2, which corrects a citation that had propagated from `11-2-…:613`. This is not pedantry: an
architectural invariant cannot be deferred, so had the obligation lived in AD-21 the red-gate arm
would have been forced rather than chosen.

## Boundaries & Constraints

**Always:**
- **This story ADDS a document; it alters none (AD-21).** Every pre-existing `expected.pdf` must
  hash exactly as it does at `8d7015a`. The baseline manifest is 23 lines whose own sha256 is
  `892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9`; after this story the diff
  against it must be **exactly one added line**.
- **The full `-tags=matrix` suite runs UNFILTERED, in-story, with `FOLIO_FONTGEN_PYTHON` set
  (D-11.1.9).** A `-run` filter here produces a green from tests that assert nothing (D-000.28).
- **Every new or changed guard is mutation-proved BY DELETION** and restored with `cp`; report each
  mutation, what it redded, and the restore digest.
- **A guard expressed as a pattern is RUN against the defect it forbids AND against the nearest
  legitimate spelling (D-11.3.7).** Reading a regex is not testing a regex.
- **Search with `grep -a`, always (DW-256 / D-11.4.4).** The shell `grep` wraps ugrep with `-I` and
  silently skips NUL-bearing files — exit 1, no output, no warning. Confirm every empty result with
  a second mechanism and report the positive control.
- The new `input.folio` is the serializer's own canonical output, not hand-formatted, and declares
  `"version": "2.0"` (an object-form chain entry raises the save version — `fontsRequireMajor`).
- The new `expected.json` records `goToolchain: "go1.26.0"`, byte-equal to `minimal-rect`'s, or
  `assertFixturesShareToolchain` fatals.

**Ask First:**
- Any change to a file outside the Execution list.
- Any edit that would move a pre-existing golden's bytes.
- Minting a new diagnostic code, or changing the resolver other than as a reverted red proof.

**Never:**
- **Never write `reader`, `date` or `examined` in a sign-off record.** That is a fabricated
  attestation and it is the single failure the whole mechanism exists to prevent.
- **Never edit `fixtures/.../worked-example.json` or `folio-go/testdata/template/golden/worked-example.json`.**
  It is byte-compared against the `## Worked example` fence in `folio-format.md`; the document and
  the doc example move together. Ship a NEW fixture.
- Never add to `baselineAcceptanceFixtures` — it is hard-pinned to exactly five and fatals otherwise.
  `beyondBaselineAcceptance` is the sanctioned route.
- Never bolt this case onto a `statement-*` fixture or any fixture that already carries a sign-off.
- Never re-record, majority-vote or auto-select a golden to make a comparison pass.
- No Thai and no CJK in this fixture. A bolded CJK run earns a Warning (no cuts exist), and Thai
  would drag in the mark-placement sign-off precedent, muddying what the owner is being asked to judge.
- Never run `npm run build`. Never push, never branch, never open a PR.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| The four cuts | Chain entry declaring `bold`/`italic`/`boldItalic`; elements declaring each combination | Four **distinct** font resources selected by four `Tf` operators; four distinct `FontFile2` programs embedded | N/A |
| Clean render | The fixture as authored | **Zero diagnostics** | Any diagnostic is a defect in the fixture, not a thing to exempt |
| Metrics witness | Same string, same box, `align: center`, one regular one bold | The two lines start at **different** x (bold measures wider) | Equal x means bold metrics never reached layout — a defect |
| Red proof | Resolver returns the base face instead of the declared variant | The new golden moves and this story's test fails; **no pre-existing golden test reds** | Restore by `cp`; report the restore digest |
| Byte identity | The fixture rendered on darwin/arm64, linux/amd64, linux/arm64, js/wasm | All four legs agree, and agree with `expected.pdf` | Pairwise disagreement falsifies NFR1; all-four-agree-but-differ is an AD-22 versioned change for a human |
| Existing corpus | Every pre-existing fixture | Bytes unchanged (AD-21) | A moved digest is a HALT, never a re-record |
| Attestation gate, unattested | No record file (the shipped state) | The matrix gate is **RED**, naming the story and saying "This is not a broken test" | Transient by construction — it clears when the owner attests |
| Attestation gate, attested | A record file carrying reader/date/examined/sha256 | The gate goes **GREEN** | A gate that stays red with a valid record is worse than no gate: it lies |

</frozen-after-approval>

## Code Map

All anchors re-derived at HEAD `4ba9e57` (D-000.4: cite by symbol, re-derive the line). **If HEAD has
moved again when you implement, re-derive them — the plan gate's anchors rot before the build runs.**

*HEAD moved mid-plan, from `8d7015a` to `4ba9e57` (the coordinator's own rulings commit). Measured:
`4ba9e57` touches only `epics.md` and `epic-11-14-decision-log.md` — nothing under `folio-go/`, `lint/`,
`folio-designer/`, `fixtures/` or `.github/` — so every baseline and every anchor below, taken at
`8d7015a`, still holds. Recorded rather than assumed.*

**The resolver — the subject, and the red-proof site**
- `folio-go/render.go:1353` -- `chainFaceNames(chain, want) (base, styled []string)`. The one boundary
  between the document's chain and the render path. **Line 1367, `variant := entry.Variant(want)`, is
  the red-proof mutation site.** Its doc comment (1258–1352) states the rule the fixture exists to
  pin: *"A VARIANT IS READ, NEVER CONSTRUCTED."*
- `folio-go/internal/template/model.go:284` -- `FontChainEntry.Variant(s FontStyle) string`. `""` is a
  first-class result meaning "declares none".
- `folio-go/internal/template/model.go:186` -- `FontChainEntry{Face, AssetKey, Bold, Italic, BoldItalic}`.
- `folio-go/render.go:2026` -- the `if variant == ""` arm in `shapeSegments`; raises
  `DiagCodeTextStyleFaceUndeclared` at `:2039`. **A correctly-authored fixture never reaches it.**
- `folio-go/internal/template/version.go:382` -- `fontsRequireMajor`, keyed on
  `FontChainEntry.SerialisesAsObject()` (`model.go:272`) — the shared predicate that forces `"2.0"`.

**Reference documents to copy from**
- `folio-designer/public/templates/starter.folio:10` -- the only committed object-form shipped-face
  entry: `{"bold": "Roboto Bold", "boldItalic": "Roboto Bold Italic", "face": "Roboto", "italic": "Roboto Italic"}`.
  The new fixture pins exactly the chain shape the product now writes for an author.
- `folio-go/fonts/fonts.go:159` -- `fonts.Shipped()`, 11 face names. `folio-go/testfont_embed_test.go:138`
  -- `testShippedFontSet()`, the same 11, kept in parity. **Measured: `go run ./cmd/folio render`
  reproduces `fixtures/three-band-page/expected.pdf` byte-for-byte, so the CLI recording recipe and
  the in-test render agree.**
- `folio-go/thai_stacked_marks_fixture_test.go:224` -- `TestThaiStackedMarksGoldenFixture`, the newest
  golden test; copy its shape.
- `folio-go/embedded_font_fixture_test.go:251` -- `TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero`,
  the canonicality + `"2.0"` + zero-diagnostics precedent; copy it wholesale.
- `folio-go/matrix_test.go:698` / `:719` -- `captureThaiStackedMarksRender` / `requireThaiStackedMarksCarriesTheRise`.
  Note the discipline: the `extraGuard` calls the **same** helper the untagged fixture test uses,
  declared once (D-7.4.5).

**Registration surface — TWELVE edit sites across eight files, empirically derived from `26e3ba1`
(thai-stacked-marks) and `ed485eb` (keep-together)**
- `folio-go/matrix_test.go:1551` -- `matrixDocuments`, **24 entries**, closes at `:2015`. Insert
  before the `alignment-rounding` entry (the precedent both 8.0 and 7.7 followed).
- `folio-go/render_test.go:659` -- `TestMain`; arms run 664–930. Const block 933–1037. Convention:
  `subprocess<PascalSlug>EnvVar = "FOLIO_SUBPROCESS_RENDER_<UPPERSLUG>"`.
- `.github/workflows/matrix.yml` -- **FIVE sites, verified**: `docs=` at `:267` (24 slugs), plus one
  path line in each upload block ending at `:76`, `:120`, `:164`, `:208`, every one
  `if-no-files-found: error`.
- `folio-go/byte_neutrality_test.go:100` -- `goldenDigestRecord`, 23 entries. Site kinds documented at
  `:83`. `folio-go/byte_neutrality_test.go:949` -- `declaredEpic2GateObligations` (7 `matrix-file:` +
  24 `matrix-document:`); its enforcer `TestEpic2GateObligationsMatchTheDeclaredSet` at `:1004` is
  **UNTAGGED** and says *"An obligation may not be added without a ruling that says so explicitly."*
- `folio-go/missing_glyph_corpus_test.go:41` -- corpus table (13 entries) **and** `:254`
  `beyondBaselineAcceptance` (8 entries). **Both**, or the identity at `:300`
  (`len(fixtures) == len(baselineAcceptanceFixtures) + len(beyondBaselineAcceptance)`, currently
  `13 == 5 + 8`) fatals. `baselineAcceptanceFixtures` (`first_baseline_acceptance_test.go:100`) is
  frozen at five — do not touch it.

**Guards that pick the fixture up with no edit — check, do not edit**
- `folio-go/matrix_registration_test.go:40` -- `TestMatrixDocumentSlugsAreRegisteredInCI`, **untagged**,
  parses `matrix_test.go` with `go/parser` and cross-checks matrix.yml both directions.
- `folio-go/embedded_font_fixture_test.go:445` -- `TestTheFontRecordCostsAnExistingDocumentNothing`
  ReadDirs `fixtures/`, requires each `input.folio` be a serializer fixed point, and hard-codes
  `withAssets != 7`. **Our fixture's `assets` is empty, so 7 stands — confirm, do not bump.**
- `folio-go/tounicode_corpus_test.go:137` -- ReadDir-driven; a per-section cap of 100 bfchars. A second
  font resource pushes per-section counts down, not up.
- `folio-go/byte_neutrality_test.go:1512` -- `TestNoPreStory80GoldenCarriesATextRise`. Latin-only, no
  marks, so no ` Ts`. Do **not** add an exemption.
- `folio-go/golden_structural_validity_test.go:90` -- page-tree resolution over every registered golden.

**The attestation surface**
- Five records exist; **none was ever committed empty**. `fixtures/shaped-text/thai-signoff.json`,
  `fixtures/expected-breaks/break-signoff.json`, `fixtures/statement-signoff.json`,
  `fixtures/thai-stacked-marks/signoff.json`, `fixtures/embedded-font/signoff.json` (a *transferred*
  reading carrying no reader/date/examined by design).
- `folio-go/shaped_signoff_matrix_test.go:78`, `expected_breaks_signoff_matrix_test.go:84`,
  `statement_signoff_matrix_test.go:103`, `thai_stacked_marks_signoff_matrix_test.go:45` -- the gates,
  all `//go:build matrix`. The absence message shape is at
  `thai_stacked_marks_signoff_matrix_test.go:53`: *"…does not exist, so Story 8.0's human semantic
  acceptance step is still OUTSTANDING.\n\nThis is not a broken test."*
- `folio-go/byte_neutrality_test.go:1190` -- `assertSignOffIsRealAndStillBinding`, called at `:1049`,
  `:1053`, `:1061` for **three** records only. A fourth fixture with no record reds nothing here.
- `folio-go/statement_golden_fixture_test.go:163` -- the record schema, whose comment at `:170-173`
  says a thin `examined` *is a defect* even though the non-empty check passes.

## Tasks & Acceptance

**Execution:**

- [x] `fixtures/declared-variants/input.folio` (NEW) -- author the document; six text elements on one
      A4 portrait page, one chain entry in object form naming `Roboto` and its three cuts, `"version": "2.0"`,
      empty `assets`. **Generate the bytes through `SerializeTemplate` so the file is the serializer's own
      canonical output** (build a throwaway module under the scratchpad with a `replace` onto `folio-go`,
      call `ParseTemplate`→`SerializeTemplate`, write the file, delete the module). Element design is in
      Design Notes and is load-bearing — do not simplify it.
- [x] `folio-go/declared_variants_template.go` (NEW, **non-test**, `package folio`) -- `const declaredVariantsTemplateJSON`,
      kept byte-identical to `input.folio` by hand. No data const: this document binds nothing.
- [x] `folio-go/render_test.go` -- add `subprocessDeclaredVariantsEnvVar = "FOLIO_SUBPROCESS_RENDER_DECLAREDVARIANTS"`
      to the const block (933–1037) with a doc comment naming this story, and its `TestMain` arm
      (parse the const, `Render(tpl, nil, nil, testShippedFontSet())`, `writeToStdoutOrDie`).
      Rationale: there is no in-process capture path; the four legs render from the const.
- [x] `fixtures/declared-variants/expected.pdf` (NEW) -- record with
      `CGO_ENABLED=0 GOWORK=off go run ./cmd/folio render -o ../fixtures/declared-variants/expected.pdf ../fixtures/declared-variants/input.folio`
      from `folio-go/`. **Confirm `SOURCE_DATE_EPOCH` is unset first** — exported, the CLI injects an
      `/Info` dict the untagged test can never reproduce. This artifact is a **CANDIDATE** until the owner attests.
- [x] `fixtures/declared-variants/expected.json` (NEW) -- `folioGoVersion: "0.0.0-dev"`,
      `goToolchain: "go1.26.0"`, `sha256` of the recorded PDF, 64 lower-case hex, 2-space indent, trailing newline.
- [x] `fixtures/declared-variants/README.md` (NEW) -- what the fixture red-proves, the exact recording
      command, **and what a human is being asked to judge** (Design Notes has the wording). If it quotes
      the digest, the `goldenDigestRecord` entry must declare a `readme` site or the completeness half fails.
- [x] `folio-go/declared_variants_fixture_test.go` (NEW) -- the untagged golden test, modelled on
      `TestThaiStackedMarksGoldenFixture` and `TestEmbeddedFontFixtureIsCanonicalAndDeclaresTwoPointZero`:
      `input.folio` ≡ the const; the document is a serializer fixed point and declares `"2.0"`;
      `expected.json` shape; committed PDF's live hash ≡ recorded `sha256`; **live render ≡ recorded `sha256`**;
      **zero diagnostics**. Plus the shared semantic-acceptance helper `declaredVariantsAssertFourCuts`,
      declared ONCE here and consumed by the matrix `extraGuard` (D-7.4.5).
- [x] `folio-go/matrix_test.go` -- `captureDeclaredVariantsRender`; `requireDeclaredVariantsUsesFourCuts`
      (calls the shared helper, `t.Fatalf` per leg); a `matrixDocuments` entry
      (`slug: "declared-variants"`, `requireFontFile2: true`, `extraGuard`, `wantPages: 1`,
      `fixtureRelPath` → `expected.json`), inserted before the `alignment-rounding` entry.
- [x] `.github/workflows/matrix.yml` -- **five edits**: append the slug to `docs=` (`:267`) and one
      `hash.<target>.declared-variants.txt` path to each of the four upload blocks.
- [x] `folio-go/byte_neutrality_test.go` -- a `goldenDigestRecord` entry (sites: `expected.json`,
      `second-literal`, and `readme` iff the README quotes the digest). Then TWO
      `declaredEpic2GateObligations` entries: `"matrix-document: declared-variants"` **citing this
      story's own acceptance criterion**, and `"matrix-file: declared_variants_signoff_matrix_test.go"`
      **citing D-11.5.1**, which says explicitly that the gate ships red pending owner attestation —
      that citation is what `TestEpic2GateObligationsMatchTheDeclaredSet` demands. Do **not** add a
      `signoff` site to `goldenDigestRecord`: no record file exists yet, and the site kind requires one.
      That site lands with the record.
- [x] `folio-go/declared_variants_signoff_matrix_test.go` (NEW, **`//go:build matrix`**) -- the
      attestation gate, ruled by D-11.5.1 arm [A]. Model it on
      `thai_stacked_marks_signoff_matrix_test.go:45`: absent record → fail naming **this story** and
      ending *"This is not a broken test."*; present record → require `reader`, `date`, `examined`
      and `sha256` all non-empty after `TrimSpace`, `date` parsing as `2006-01-02`, and `sha256`
      equal to the fixture's live digest (a moved golden must invalidate the reading). **Ship it RED.**
      **No agent writes `reader`, `date` or `examined`, and no record file is committed.**
- [x] `folio-go/missing_glyph_corpus_test.go` -- a corpus-table entry (parse the const, `testShippedFontSet()`)
      **and** a `beyondBaselineAcceptance` entry with a non-empty story-naming reason. The identity becomes
      `14 == 5 + 9`.
- [x] **The red proof.** Mutate `render.go:1367` **by deletion** (`variant := entry.Variant(want)` → the
      variant is never read), run the untagged suite, and record: the new golden test reds; **no
      pre-existing golden test reds** — the measurement that demonstrates DW-237 directly. Restore with
      `cp` from a pre-mutation snapshot and report the restored file's sha256.
- [x] **Mutation-prove the attestation gate IN THE CLEARING DIRECTION (D-11.5.1's condition).** This is
      the opposite of every other guard in this story, and it is not optional. Red is the gate's resting
      state, so proving it can red proves nothing; the property in question is that the owner's
      attestation will actually **discharge** it.
      1. Write a **synthetic** record file into the fixture directory — plainly fake reader/date/examined
         plus the real digest. Run the gate. Confirm it goes **GREEN**.
      2. Delete it. Run the gate. Confirm it returns to **RED**.
      3. Report both outcomes, the synthetic file's sha256, and — after deletion — that
         `git status --porcelain` shows no record file. **The synthetic record must not survive the
         story**; it is a probe, never a commit, and it is the one artifact here that would be a
         fabricated attestation if it leaked.
- [x] **THE HALT.** Stop with the candidate `expected.pdf` and report its path. Do not self-attest.

**Acceptance Criteria:**

- Given a chain entry declaring `bold`, `italic` and `boldItalic`, when the fixture is rendered, then the
  page embeds four distinct font programs and selects four distinct font resources — and the render
  raises **zero** diagnostics.
- Given the centred pair (same string, same box, one regular one bold), when the content stream is read,
  then their line-start x-coordinates **differ** — pinning that bold layout came from the bold face's
  metrics. (Measured on the candidate: 132.548 vs 130.724 at width 400.)
- Given the fixture rendered on all four targets, when `TestCrossTargetByteIdentity` runs, then all four
  legs agree with each other and with `expected.pdf`.
- Given the resolver pointed at the base face instead of the declared variant, when the suite runs, then
  the new golden moves and this story's own test fails, **and no pre-existing golden test fails** — the
  red proof that this fixture is a witness rather than a decoration.
- Given every pre-existing fixture, when the corpus is re-hashed, then the manifest differs from the
  `8d7015a` baseline by **exactly one added line**.
- Given the new `expected.pdf`, when the story ends, then it is a **candidate** carrying no
  self-attestation, no record file is committed, and the story has HALTED for the owner.
- Given a synthetic record file, when the attestation gate runs, then it goes **GREEN**; and given that
  file removed, then it returns to **RED**. A gate that cannot be cleared is a decoration, not a halt.

## Rulings applied at CHECKPOINT 1

All three questions were ruled by the coordinator and logged as **D-11.5.1** and **D-11.5.2** in
`epic-11-14-decision-log.md`. They are settled — apply them, do not re-open them at review.

- **Token gate: [K] KEEP** (D-11.5.1). No seam: the fixture, its goldens, its twelve registration sites
  and its red proof are one artifact, and splitting would *multiply* the matrix gate across both halves
  rather than remove it from either — D-11.1.18's situation, and the inverse of D-11.2.12's.
- **Q1: arm [A], ship the red gate** (D-11.5.1). Arm [B] leaves a tree that says "done" and a register
  that says "not really" — a state indistinguishable from its opposite by looking. Arm [C] has no
  precedent and trips the empty-field errors.
- **The third-failure rule survives, with the distinction made explicit** (D-11.5.1, quoted):

  > `TestCorpusMeetsP6ExerciseFloors` and its `P6g` subtest are **permanent** mandated reds — a floor
  > nobody has met. The attestation gate is a **transient** red that clears the moment the owner
  > attests. A permanent red teaches everyone to ignore a number; a transient one is a countdown. Both
  > must be enumerated BY NAME in every baseline, and a failure whose name is not on that list is still
  > a hard stop.

- **Q2: citations.** `matrix-file: declared_variants_signoff_matrix_test.go` cites **D-11.5.1**;
  `matrix-document: declared-variants` cites this story's own acceptance criterion.
- **Q3: agreed** — slug `declared-variants`, family **Roboto**.
- **D-11.5.2: the citation correction is ratified and `epics.md` is already corrected** (committed in
  `4ba9e57`). The AC now reads *"under D-000.22 → D-2.3.5 (doctrinal, NOT architectural…)"*. Story 11.2's
  line is deliberately left uncorrected as the historical record of where the error came from — **do not
  "fix" it.**

## Spec Change Log

*Append-only. Empty until the first `bad_spec` loopback — the CHECKPOINT 1 rulings above were applied
before approval and are not loopback entries.*

**2026-09-06 — known gap, closed OUTSIDE the frozen block (not a loopback).** The frozen
`## Boundaries & Constraints` says "never push, never branch, never open a PR" but does **not** say
"never commit / never `git add`". That is an omission from planning: the orchestrator's prohibition on
agent commits is absolute for this run, and this repo has a recorded history of implementation
subagents committing unasked at step-03. The hole is closed **operationally** — the implementation
dispatch carries an explicit git-writes prohibition — rather than by reopening the frozen block. The
coordinator ratified the deviation from step-03's "add nothing beyond the spec" and **declined to
reopen the block**, on the ground that restating an already-enforced rule buys nothing and costs the
block's authority. The boundary template is to be fixed at the epic boundary so the next story
inherits it. **Do not "fix" the frozen block to match this note.**

**2026-09-06 — step-04 patch round (no loopback).** Triaged every finding as `patch` rather than
`bad_spec`, deliberately. The discriminator was local-vs-pervasive — each fix is uniquely determined
and touches no rendered byte — but a second reason applies here that does not generalise from token
cost: **the golden had already been attested by a human.** A `bad_spec` loopback reverts and re-derives
code, and re-deriving a fixture risks moving bytes a person has read and signed. That makes a revert
expensive in a way that is not about tokens or time: it would spend a second reading of the same page
by the same person. **An attested artifact raises the bar for reverting anything that could move it**,
and that is worth carrying beyond this story.

Applied: the `signoff` site declared in `goldenDigestRecord`; the sign-off completeness set **derived**
from `{kind:"signoff"}` sites rather than hand-listed (D-11.5.1's stronger form — a declared site with
no completeness check is now inexpressible, with a reasoned exemption map for the two records whose
schema differs); an index-keyed per-run resource assertion; a second absolute clause for the metrics
witness; `assertBaseFontNames` routed through the injected `fatalf`; `~1.4%` corrected to **`~1.76%`**;
six now-false "CANDIDATE / red by construction" prose sites past-tensed; the README's attestation
instructions corrected to hash the PDF actually opened.

**KEEP on any re-derivation.** (1) The fixture bytes — `input.folio`, `expected.pdf` and the owner's
`signoff.json` — are frozen; `2405d005…` is attested. (2) The centred pair, not a wrap demonstration:
the wrap version was measured vacuous across 14 box widths. (3) The derived sign-off set, and its three
proofs — existing records named, a declared-site-without-record redding, and the blanked-record
mutation redding. (4) The per-run expectation must stay a **literal**: a derived one was measured
passing a genuinely wrong page.

## Design Notes

**The fixture, validated by rendering it before it was specified.** Six text elements, A4 portrait,
36pt margins, one chain: `{"bold": "Roboto Bold", "boldItalic": "Roboto Bold Italic", "face": "Roboto", "italic": "Roboto Italic"}`.

```
e1  24pt              "Handgloves — Roboto Regular"
e2  24pt bold         "Handgloves — Roboto Bold"
e3  24pt italic       "Handgloves — Roboto Italic"
e4  24pt bold+italic  "Handgloves — Roboto Bold Italic"
e5  24pt align:center "Handgloves quickly"          (regular)
e6  24pt align:center "Handgloves quickly"          (bold)
```

Measured on the candidate: zero diagnostics; four subset faces
(`Roboto-Regular`, `Roboto-Bold`, `Roboto-Italic`, `Roboto-BoldItalic`); four `Tf` operators selecting
four distinct resources; 60,594 bytes.

**Why each half is load-bearing.**
- **e1–e4 are the human's test, and each line makes its claim in the face it claims.** A line that
  *says* "Roboto Bold" while *looking* regular is a swap a person sees instantly — the page witnesses
  itself. "Handgloves" is the type-tester's word because it carries ascender, descender, round bowl and
  tight counters in ten letters.
- **e5/e6 are the machine's test, and they exist because the obvious version was vacuous.** The first
  design was a wrap demonstration: the same paragraph in the same box, regular vs bold, expecting bold
  to take an extra line. **It never did** — swept across 14 box widths from 150 to 340pt, the line
  counts were equal every single time. Rather than tune until it looked right, I measured why: centred,
  the same string starts at x=132.548 regular and x=130.724 bold, so the bold line *is* 3.648pt wider
  and bold metrics *do* reach layout — Roboto's bold is simply only ~1.4% wider than its regular, so a
  break almost never moves. The centred pair is the sensitive form of the same assertion: any width
  delta shows, and D-11.2.8 is satisfied by measurement rather than by hope — the two sides are pinned
  unequal, real output against real output.

  **Someone will propose the wrap test again**, because it is the obvious way to show bold metrics
  reaching layout. It is recorded here so the next person inherits the measurement instead of repeating
  the sweep: bold *does* reach layout, Roboto's bold is simply ~1.4% wider, and a line break almost
  never moves. A wrap assertion on this family would be a test that passes for the wrong reason.

**What the owner is being asked to judge, and how to judge it by eye.** Open `expected.pdf`. Four lines,
each naming the cut it is set in. The question is only: *is line 2 a real bold face, or a regular face
thickened?* The tells, in order of reliability:
1. **Counters** — the enclosed white inside `a`, `e`, `o`, `g`. A real bold keeps them open and shaped;
   a smeared regular chokes them toward slits. Line 2's "Handgloves" against line 1's is the direct comparison.
2. **Stem-to-round contrast** — a real bold thickens vertical stems more than the thin parts of curves.
   A synthetic bold thickens everything uniformly and looks inflated rather than drawn.
3. **Width** — a real Roboto Bold is slightly wider; line 2 should end marginally right of line 1's
   comparable point, not sit exactly on top of it.
4. **Lines 3 and 4** — a real italic is a *drawn* italic, not a slanted regular: check `a`, `f`, `e` for
   different letterform construction rather than the same shapes leaning over.

**Why this cannot be a variation on an existing fixture.** `worked-example.json` is the one in-tree
document that declares bold, and `goldenfixture_test.go` byte-compares it against the `## Worked example`
fence in `folio-format.md` — document and doc example move together. The `statement-*` family already
carries a human sign-off, and widening one makes every future change cost a person re-reading four
documents.

**Known inherited weakness, NOT this story's to fix (D-R7.7, owed to Story 15.2).** Every sign-off gate
checks only that `reader`/`date`/`examined` are non-empty — never that the record is *recent* or *about
this photograph*. If the owner wants this fixture's gate to be the first to assert recency, that is a
scope decision for them, not a thing to build unasked.

## Verification

Run each module's commands in **its own** invocation — a `cd` persists through a compound command and
makes later relative paths resolve elsewhere, printing `lstat …` lines that read like passes. Use
`-count=1` on every Go run. **Never `&&`-chain these**; a conjunction drops everything after its first
failing term, and `$?` after a pipe is the tail command's. Count from `go test -json` `Action` events
carrying a `Test` field — a plain run prints no totals.

**Commands:**

- `cd folio-go && go test -count=1 ./...` -- baseline **measured at `8d7015a`: 2237 pass / 2 fail / 5 skip**,
  the only failures `TestCorpusMeetsP6ExerciseFloors` and its `P6g_(opaque_names)` child, a mandated
  permanent red never to be "fixed". **A third distinct failure is a HARD STOP**, checked by enumerated
  name, never by exit code. Expect the pass count to rise by this story's new tests.
- `cd folio-go && FOLIO_FONTGEN_PYTHON=/private/tmp/claude-501/-Users-panitw-Projects-folio/c910e6ef-835c-4b93-9710-f11720b44f3a/scratchpad/fontgen-venv/bin/python go test -count=1 -tags=matrix ./...`
  -- **UNFILTERED, all four targets, in-story (D-11.1.9)** — this story ships goldens, so it is
  non-negotiable. Baseline measured at `8d7015a`: **2249 / 2 / 5**, with
  `fontgen: derived and compared 7 of 7 faces` and `TestCrossTargetByteIdentity` genuinely executing
  four legs (docker + node confirmed present). ⚠ A `-run` filter produces a green from tests that
  assert nothing (D-000.28).

  **This story ends this suite at 2248 / 3 / 5, and all THREE failures are named (D-11.5.1):**
  1. `TestCorpusMeetsP6ExerciseFloors` — **PERMANENT** mandated red, a floor nobody has met. Never "fix" it.
  2. `TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)` — **PERMANENT**, the subtest of the same.
  3. `TestDeclaredVariantsSemanticSignOffIsRecorded` — **TRANSIENT**. It is this story's halt, and it is
     discharged by exactly one thing: the owner attesting and a record file landing. A permanent red
     teaches everyone to ignore a number; a transient one is a countdown.

  **A failure whose name is not on that list of three is still a hard stop.** The pass count also rises
  with this story's new tests; diff test NAME sets, not totals.
- `cd lint && go test -count=1 ./...` -- **227 pass**, exit 0. `-count=1` is mandatory: the rules package
  walks `folio-go` by ReadDir and the Go test cache does not track that, so a cached `ok` is no measurement.
- `gofmt -l folio-go lint` (from the repo root) -- **empty output**, exit 0.
- `cd folio-designer && npx tsc -b --force` -- exit 0. `--force` is required.
- `cd folio-designer && npm test` -- baseline **65 files / 976 tests**. This story touches no designer
  source; the expectation is *unchanged*.
- `cd folio-designer && npx oxlint` -- exit 0, **exactly 4** pre-existing `only-export-components` warnings.
- `cd folio-designer && npm run test:e2e:compile` -- exit 0. A green here is a **TYPECHECK**, not coverage:
  the suite is exercised at epic boundaries (D-000.30).
- **`npm run build` is Ask First. Do not run it.**

**Manual checks:**

- **`shasum -a 256 fixtures/*/expected.pdf`, diffed against the `8d7015a` baseline manifest
  (23 lines, manifest sha256 `892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9`).
  Exactly ONE added line, zero changed lines.** State plainly that no pre-existing golden moved.
- **Diff test NAME sets, not totals**, against the baseline JSON runs. A count that matches can still
  hide a swap.
- **`qpdf --check`** on the new `expected.pdf`.
- **Mutation-prove every new guard BY DELETION**, restore with `cp`, and report each mutation, what it
  redded, and the restore digest. Specifically: delete the four-cut assertion and confirm it reds on a
  single-face render; delete the centred-offset assertion and confirm it reds when the two x values are equal.
- **Mutation-prove the attestation gate IN THE CLEARING DIRECTION (D-11.5.1's condition)** — synthetic
  record in, gate GREEN; record removed, gate RED. Report both outcomes and the synthetic file's sha256,
  and confirm by `git status --porcelain` that it did not survive. This is the one proof this story
  cannot substitute a red-direction proof for: red is the gate's resting state.
- **Run the red proof and report both halves**: the new golden test reds, AND no pre-existing golden
  test reds. The second half is the measurement that demonstrates DW-237 rather than restating it.
- **Confirm `SOURCE_DATE_EPOCH` is unset** in the recording shell, and say how you confirmed it.

## Suggested Review Order

**Start here — the artifact the story exists to introduce, and the one a human has read**

- The attested page itself: six elements, four cuts, each line naming the cut it is set in.
  [`input.folio`](../../fixtures/declared-variants/input.folio)

- What the owner was asked to judge, and how to tell a real bold from a thickened regular.
  [`README.md`](../../fixtures/declared-variants/README.md)

**The semantic guard — why this fixture is a witness and not a digest**

- Per-run expectation held as a LITERAL; a derived one measurably passed the defect.
  [`declared_variants_fixture_test.go:145`](../../folio-go/declared_variants_fixture_test.go#L145)

- The four-cut and metrics clauses, declared once and shared with the matrix leg.
  [`declared_variants_fixture_test.go:303`](../../folio-go/declared_variants_fixture_test.go#L303)

- The metrics witness' second, absolute clause: both lines moving together no longer passes.
  [`declared_variants_fixture_test.go:95`](../../folio-go/declared_variants_fixture_test.go#L95)

**The attestation mechanism — derived, so it cannot be forgotten again**

- The sign-off completeness set is derived from declared sites, not hand-listed.
  [`byte_neutrality_test.go:1356`](../../folio-go/byte_neutrality_test.go#L1356)

- The two reasoned exemptions, and the residual statement-record gap said plainly.
  [`byte_neutrality_test.go:1341`](../../folio-go/byte_neutrality_test.go#L1341)

- The gate that shipped red and was discharged inside the story.
  [`declared_variants_signoff_matrix_test.go:95`](../../folio-go/declared_variants_signoff_matrix_test.go#L95)

**Registration — twelve sites; a miss here makes the fixture invisible rather than red**

- The digest record, now including the signoff site the record required.
  [`byte_neutrality_test.go:630`](../../folio-go/byte_neutrality_test.go#L630)

- Four-target registration and the per-leg guard.
  [`matrix_test.go:1976`](../../folio-go/matrix_test.go#L1976)

- CI slug list plus one upload path per target.
  [`matrix.yml:77`](../../.github/workflows/matrix.yml#L77)

**Peripherals**

- The template const, kept byte-identical to `input.folio` by hand.
  [`declared_variants_template.go:79`](../../folio-go/declared_variants_template.go#L79)

- The subprocess selector the four legs render through.
  [`render_test.go:1055`](../../folio-go/render_test.go#L1055)

- Corpus row and its `beyondBaselineAcceptance` reason; the identity is now 14 == 5 + 9.
  [`missing_glyph_corpus_test.go:228`](../../folio-go/missing_glyph_corpus_test.go#L228)

## Delivery Log

### 2026-09-06 — done

Baseline `4ba9e57`. Shipped as one commit, `6d26a80` — 15 files, +2123/−15, all of them this story's.
`fixtures/declared-variants/` is now a pinned golden: one A4 page, four Roboto cuts, rendered
byte-identically on all four targets, plus twelve registration sites, a semantic guard and a human
attestation. Rulings applied by ID: **D-11.5.1** (token gate KEEP; arm [A] ship the red gate; the
third-failure rule survives with permanent-vs-transient made explicit), **D-11.5.2** (the
`AD-21 / D-4.7.1` citation is wrong — the obligation descends **D-000.22 → D-2.3.5**; 11.2's line is
deliberately left as the record of where the error travelled), **D-11.1.9** (matrix suite unfiltered,
in-story), **D-000.28**, **D-11.3.7**, **D-11.4.4 / DW-256**.

**The attestation is real, and its provenance is in the record.** The owner confirmed the **narrow**
judgment on a direct question — that line 2 is a real bold font program and lines 3–4 are drawn
italics — not merely that the page rendered. The question was put precisely because "rendered
correctly" could have meant "displayed", and attesting the wrong property would have looked settled
forever. `signoff.json` records that verbatim, including that the orchestrator transcribed it at the
owner's instruction, and binds to the LIVE hash of `expected.pdf` so a re-record invalidates the
reading by construction.

**The attestation was enforced by nothing automated, and review found it.** Blanking `reader` and
`examined` and setting `date: "tomorrow-ish"` left the untagged suite byte-identical to baseline;
zeroing the digest turned the check FAIL→ok — the gradient pointing the wrong way. The fix DERIVES the
checked set from `goldenDigestRecord`'s declared `{kind:"signoff"}` sites rather than adding a fourth
hardcoded call. That choice matters: the adjacent comment had already warned about exactly this and did
not prevent the fourth, so it would not have prevented the fifth. A declared site with no completeness
check is now inexpressible, with a reasoned exemption map for the two records whose schema differs.

**The index-to-element tie shipped vacuous, and was caught by mutating rather than reading.** It
derived expected faces via `entry.Variant(style)` while arguing that read the document rather than the
resolver — false, since that call IS the resolver's lookup. **The mutation SITE decided it:**
swapping bold↔italic perturbs the centred pair and looked caught; swapping italic↔boldItalic touches
nothing else and exposed the hole. Replaced with a literal table plus a staleness tie reading the
struct fields directly.

**Triage: every finding `patch`, none `bad_spec`, none rejected.** No numeric tally was recorded; the
"Applied" list enumerates **8** distinct patches, which is a floor rather than a census. The
discriminator was local-vs-pervasive, but a second reason applies that does not generalise from token
cost: **an attested artifact makes reverting expensive in a way that is not about tokens** — a
`bad_spec` loopback re-derives the fixture and risks moving bytes a person has read, spending a second
reading by the same person. That generalises beyond this story and is why triage chose `patch`.

**Instruments produced three false signals this session** — an empty `git diff`, a PATH-less loop
reporting 14 phantom drifts, and a `cp` restore silently failing behind a bad `cd`. Re-measuring with a
second mechanism caught all three. A warning you wrote yourself is not immunity: this story's own spec
mandates `grep -a` (DW-256), and the instrument still had to be double-checked each time.

**Gates, re-measured independently at `6d26a80` at close — not carried from the build report.** Counted
from `go test -json` `Action` events, since a plain run prints no totals. Untagged **2242 pass / 2 fail
/ 5 skip**; unfiltered `-tags=matrix` with `FOLIO_FONTGEN_PYTHON` set **2255 / 2 / 5**, reporting
`fontgen: derived and compared 7 of 7 faces`. **Both fail only `TestCorpusMeetsP6ExerciseFloors` and
its `P6g_(opaque_names)` child, checked by enumerated name and not by exit code.** The third failure
this spec predicted — `TestDeclaredVariantsSemanticSignOffIsRecorded`, the transient attestation
halt — **now PASSES**: the owner's record landed in this same commit and discharged it, so the suite
ends at 2 failures rather than the 3 the Verification section forecasts. `lint` **227 pass / 0 fail**.
`gofmt -l` over absolute `folio-go` and `lint` paths: **empty**. `tsc -b --force` **0**. `npm test`
**65 files / 976 tests**, all passing. `oxlint` **exactly 4** pre-existing `only-export-components`
warnings. `test:e2e:compile` **0** — a typecheck, not coverage; the e2e suite itself is an
epic-boundary gate (D-000.30) and is **unrun** here, due at the Epic 11 boundary. `npm run build` was
**not run** (Ask First).

**Golden manifest, re-derived rather than relayed.** The `8d7015a` baseline rebuilt from `git ls-tree`
is 23 lines whose own sha256 is `892a1505e5e7fff0184310d5f70eb7bfcfa10d18cda9af4e2aecf262a0630ce9` —
matching the figure the spec declares, which validates the reconstruction method. Diffed against
`6d26a80`: **exactly one added line, zero removed, zero changed.** **No pre-existing golden moved.**
The three frozen artifacts verify unchanged at close: `expected.pdf` `2405d005…`, `input.folio`
`abc8a997…`, `signoff.json` `737a672d…`.

**Deferred, with owners.** Three entries minted at this close from weaknesses the builder named but
nobody numbered: **DW-267** (statement sign-off FIELDS checked only under `-tags=matrix` while its
digests are checked untagged — HIGH, pre-existing, unassigned), **DW-268** (no workflow runs the
unfiltered `-tags=matrix` suite — HIGH, **OWNER-DECISION-PENDING** at the Epic 11 boundary gate; note
`ci.yml:83`'s existing `-skip "$KNOWN_RED_TEST"` mechanism as a candidate), and **DW-269** (two
committed measurements of the fixture-directory count disagree, and the surviving comment's premise was
discharged by this very story — LOW). **D-R7.7** (sign-off records are never checked for recency) is
cross-referenced from DW-267 and remains owed to **Story 15.2** — not duplicated here.

**Epic 11 has no remaining stories, but the epic is deliberately NOT closed** — its boundary gate and
DW-268's owner decision are both outstanding.
