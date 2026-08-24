---
baseline_commit: 17f5f7a
---

# Story 2.5a: Align the first baseline with the leading model

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-5a-align-first-baseline-with-the-leading-model`
**Status:** `done`
**Charter:** **DW-15** in `deferred-work.md`. **This story does not come from `epics.md`.** It has no
FR of its own: it makes an already-shipped FR (FR5/FR6 text placement) internally consistent.
**Primary invariant:** **AD-21 / AD-22** (every feature ships its golden; a hash change is an
intended, versioned event) — this story exists *because* five goldens move at once and that movement
must be attributable to exactly one cause.
**Adjacent invariants:** AD-24 (boxes are absolute, and nothing negotiates; bands are placed by
`internal/layout` alone) · AD-5 (the page model knows nothing about PDF) · AD-2 (one fixed-point
unit; font scaling is one exported function) · AD-13 (derived geometry, one function) · AD-23 (no
`float32`/`float64`) · AD-3 (no number reaches an output byte by any other route)
**Governing rulings:** **D-2.4.2** (THE leading rule) · **D-2.3.5** (frozen with sign-off pending;
the machine-checkable half is never deferrable) · **D-000.22** (semantic acceptance at first
recording) · **D-000.26 (refined)** (a sign-off binds to the artifact expressing the property
judged) · **D-000.41** (request a scarce human sign-off only when no scheduled work is known to move
the artifact) · **D-000.30** (capture the red-proof before the fix) · **D-000.39 (sharpened)** ·
**D-000.40 (sharpened)** · **D-000.42** · **D-000.33** · **D-000.34** · **D-000.35** · D-000.9 ·
D-000.21 (sharpened) · D-000.23 · D-000.24 · D-000.26 · D-000.28 · D-000.29 · D-000.32 · D-000.36 ·
D-000.37 · D-000.38 · D-000.4 · D-1.8.10 · D-2.5.1
**Deferred work touched:** **DW-15 is discharged here** (measured live, not vacuous — see
*Measured findings* 1). **DW-16 is NOT touched** (its direction is ruled; its owner is a later
story). **Two new DW candidates are raised, not fixed** — see *DECISIONS NEEDED* DN-3 and DN-4.

---

## Baseline, measured at creation

HEAD is **`17f5f7a`** — *"Story 2.5: Compose a page from three bands (finisher)"* — on `main`.
Every number below was **re-measured at this commit**, not inherited from the brief.

**The working tree is NOT clean at this baseline, and that is reported rather than corrected.**
`git status --porcelain` reports two modified files:

```
 M _bmad-output/implementation-artifacts/folio-mvp-decision-log.md
 M _bmad-output/implementation-artifacts/sprint-status.yaml
```

The decision-log change is a **172-line append** carrying the rulings made today — including
**D-000.41**, which names *this story by key* — and the `sprint-status.yaml` change is the 5-line
insertion of this story's own key and its sequencing comment. **This story neither reverts nor
commits either**: they are the artifacts that authorised it. **No `folio-go/` or `lint/` source file
is dirty**, so every code measurement below is against `17f5f7a`'s code exactly.

Every number is stated with its scope, its invocation and its counting rule (D-000.26):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` occurrence, subtests included | **487 PASS · 1 FAIL** |
| `folio-go/` | the same invocation, counting only **top-level** results (`^--- PASS` at column 0) | **305 PASS · 1 FAIL** |
| `lint/` | the same invocation, all-occurrences | **85 PASS · 0 FAIL** |
| `lint/` | the same invocation, top-level | **47 PASS · 0 FAIL** |
| `folio-go/` | `GOWORK=off go list -m all` | exactly **two** modules: `github.com/panitw/folio/folio-go`, `github.com/boxesandglue/textshape v0.0.15` |
| `folio-go/` | `go vet ./...` | clean |
| both modules | `gofmt -l` on every file this story touches | clean |

**The single failure is `internal/text`'s `TestCorpusMeetsP6ExerciseFloors`** — *"P6g (opaque names)
floor not met: got 7, need >=20"*, with `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115
P6g:7}`. **Story 2.4's AC5 requires it to stay red and byte-identical.** Do not fix it, do not skip
it, do not add corpus items to fill it. **A second failure is a regression** — with the one, named
exception in AC10 (the golden fixture tests are expected to go red *during* this story and green
again at its close; they must not be left red).

> **A note on the measurement instrument.** `rtk` intercepts `go test` and rewrites its output; a
> filtered run reports *"text (47 passed, 1 failed)"* and no `--- PASS` lines at all, so
> `grep -c -- '--- PASS'` returns **0** and looks like a catastrophic regression. Every count above
> was taken through `rtk proxy "go test …"` — the raw stream. **Take yours the same way, or your
> gate table will be fiction** (D-000.40's portable-tooling hazard, one instrument further out).

---

## In plain terms (read this first if you just want the gist)

When Folio draws a paragraph, two distances matter: how far below the top of its box the first line
sits, and how far apart consecutive lines sit. The engine answered them from different sources: spacing
from the tallest metrics any declared font could demand, the first line from nothing but the point
size. They agreed only by accident.

They now come from **one rule with three maxima**, each taken independently across the whole
declared font list: the tallest ascent, the deepest descent, and the largest built-in line gap. The
first line is placed by the first; the gap between lines is the sum of all three; the third is
computed and written down for a later story that will need it.

Maximising each axis separately, rather than trusting the worst single font, also corrected the line
spacing — but **only where more than one font is declared**, because with a single font the old and
new arithmetic are the same expression term for term. Latin-only documents kept their spacing
exactly; only their first line moved, and it moved *up*, opposite to everything else.

Five reference documents moved and were re-recorded; two more were re-hashed and proved unchanged.
Each of the five gained a check that reads the new position back out of the finished PDF, written
before those files existed.

Four things look wrong and are not. One test fails on purpose — an earlier spike's deliberately
unmet corpus floor, untouched here. The pending Thai reading request was deliberately not asked,
because this story moved the very file it would judge; the condition for asking it is recorded in
the epic's boundary document. The heaviest cross-platform checks were compiled but not run, by the agreed
once-per-epic cadence. And the epic's status stays open: flipping it belongs to the gate.

Review found no defect in the behaviour, only in the record: a comment naming a test that does not
exist, a message stating its two directions backwards, a mistyped fingerprint in a document a human
will act on, and counts that did not reproduce. All corrected, and the fingerprint now has a check.

---

## Story

As a template author,
I want the first line of an element placed by the same rule that spaces every line below it,
So that a paragraph's first line is not silently offset from the rest of its own text.

---

## Do not re-open — settled rulings this story inherits

Each of these is **binding and ratified**. Do not relitigate, do not propose an alternative, and do
not write a comment that states the opposite.

1. **D-2.4.2 — the leading rule.** Inter-baseline distance is the **maximum, over the faces of the
   DECLARED chain that are present in the supplied `FontSet`, of that face's `hhea`
   (ascent − descent + lineGap), scaled to the font size**. It is a function of the chain and the
   size and of **nothing that is drawn**. `hhea`, not OS/2 typo metrics, not a multiple of the size;
   the maximum, not the first face. All of that was measured, is recorded, and is out of scope here.
   The symbol is `lineAdvance` in `folio-go/wrap.go:323`.
2. **D-2.4.2 constraint 2 — never inherit a substituted default.** Line metrics are read from the
   `hhea` **table**, never through `(*ot.Face).Ascender/Descender/LineGap`, which substitute
   `800/-200/0` for an absent table. The symbol that does this correctly is
   `(*fontset.Font).LineMetrics` (`folio-go/internal/fontset/fontset.go:973`), and
   `requireReadableTables` makes an absent `hhea` a load error. **This story's new arithmetic reads
   `LineMetrics` and nothing else.** In particular it must NOT read `(*fontset.Font).Metrics`'s
   `Ascent` field (`fontset.go:637`), which is populated through the substituting vendor accessor
   and exists for the PDF `/FontDescriptor`, not for placement.
3. **AD-24 / D-1.8.10 — exactly one inverter.** Every content-stream Y coordinate the PDF module
   emits is derived from `flipY` (`folio-go/internal/pdf/flip.go:28`) and from nowhere else, and the
   guard for that is **positive** (`TestContentStreamYCoordinatesRouteThroughFlipY`,
   `folio-go/internal/pdf/flip_test.go`) — assert that all emission routes through the one function,
   never "nobody else writes a minus sign". **This story does not change `flipY`'s signature,
   arity or body.** It changes only what is passed as its fourth argument for text.
4. **AD-5 / AD-24 — placement is decided before a renderer sees it.** `pagemodel.TextRun.X/Y` are
   **PAGE-ABSOLUTE** offsets from the printable top-left, Y increasing downward. This is
   **PERMANENT, not provisional** (`pagemodel.go:40`). The first-baseline offset is a *layout*
   quantity, so it is computed in `folio-go/` and carried on the page model — **not** derived inside
   `internal/pdf` from `faces[run.Face]`. Deriving it per-run from the resolved face would also make
   it **content-dependent**, which D-2.4.2 rules out for exactly the reason it rules it out for
   leading: adding one CJK character would reflow the element.
5. **D-2.3.5 — the machine-checkable half of a semantic acceptance step is NEVER deferrable.** Only
   *"does this Thai read correctly to someone who reads Thai"* may be pending.
6. **D-000.22 — a golden needs a semantic acceptance step, read off the artifact.** A re-record is a
   recording; it needs the step. The property must be one that would be wrong if the recording were
   wrong, and it must be read off the **produced artifact**, never off the inputs.
7. **D-000.26 (refined) — a sign-off binds to the artifact expressing the property judged.** The
   Thai **reading** judgment binds to `fixtures/shaped-text/expected.pdf`'s digest. The Thai
   **break** judgment binds to the break-opportunity vector, which a baseline shift does not touch.
   **Measured and confirmed in this story** — see *Measured findings* 4.
8. **D-000.41 — do not request a scarce human sign-off while scheduled work is known to move the
   artifact.** This story **must not create** `fixtures/shaped-text/thai-signoff.json`, and must not
   ask anyone to read anything. See *Scope fence* and AC13.
9. **D-000.29 — an audit row ends settled or fixed, never carried.** DW-15 ends **fixed** here. It
   does not end "partially addressed".
10. **D-2.5.1 — the Epic 2 gate obligation set is a declared list, not a count in a test name.**
    `declaredEpic2GateObligations` (`folio-go/byte_neutrality_test.go:161`) is that list, and
    `TestEpic2GateObligationsMatchTheDeclaredSet` asserts the observed set equals it exactly.

---

## Corrections to inherited claims — verified at `17f5f7a`

Both were checked against the source before this story was written, and both are wrong as stated.

### C1 — DW-15 says **four** goldens move. **Five** do. Measured.

DW-15's load-bearing sentence is *"correcting it re-records four goldens"*. That figure is wrong,
and the fifth is `fixtures/font-text/` — the oldest text golden in the repository, the only one that
does not use the shipped Noto chain, and **the one whose baseline moves in the opposite direction**.
Measured by applying a candidate fix in a throwaway worktree at `17f5f7a` and running the full suite
(*Measured findings* 2). **Do not act on the number four.**

This is Story 2.4's lesson repeated exactly: *a stated list can be wrong*, and D-000.41 itself says
this measurement is *"answerable early and cheaply"*.

### C2 — `render.go`'s `splitByFace` is described as the placement path. It has **zero call sites**.

`folio-go/render.go:463` declares `splitByFace`, and **ten** separate comments in the module describe
it as *the* thing that positions face segments — **seven production comments** (`render.go:71`,
`:73`, `:425`, `:477`, `:593`, `:597`, and `internal/fontset/fontset.go:576`), **`wrap.go`'s two**
(`:88`, `:111`) and **`segment_origin_test.go:42`** in the tests.

*(Enumeration corrected at finish under Finding 7(a), which flagged that this list omitted
`render.go:593`, `:597` and `wrap.go:111`. Re-measured by construction rather than re-read:
`git grep -c splitByFace 17f5f7a -- folio-go` reports `internal/fontset/fontset.go:1`,
`render.go:7`, `segment_origin_test.go:1`, `wrap.go:2` — **11 lines**, of which one is the `func`
keyword of the declaration itself, leaving **10 references**. The reviewer's own resolution proposed
aligning this list with DN-4's "nine"; **DN-4's nine is off by one too** — it counts
`render.go`'s six as "six production comments" and drops `internal/fontset/fontset.go:576`. Both
places now say ten. Counting rule: matching **lines** under `folio-go/` at `17f5f7a`, declaration
excluded.)*
Measured at `17f5f7a`: **no production code and no test calls it.** Story 2.4's `packLines` /
`positionSegments` path replaced it and the function was left behind. It computes placement on the
**old** model, so a mechanical "update every placement site" sweep will either miss it or silently
half-update it. See DN-4.

### C3 — DW-15 and the brief both cite `internal/pdf/textdoc.go:730`. The defect is at **`:689`**.

Measured at `17f5f7a`: `internal/pdf/textdoc.go` is 917 lines; the sole `flipY` call in it is at
**line 689**, inside `buildTextContentStream`, which begins at line 677. Line 730 is inside
`appendShapedRun`, a different function that does not place anything. The comment that *describes*
the defect is at `:663`–`:672`; the formula is restated at `:667`.

Minor on its own, and recorded because a developer told to change line 730 will change the wrong
function and a reviewer told to check line 730 will find nothing wrong.

---

## Scope fence — what this story is NOT

- It does **not** re-architect text layout. `packLines`, `shapeSegments`, `positionSegments`,
  `lineAdvance`'s rule and the band model are untouched in behaviour.
- It does **not** touch `pagemodel.ShapedGlyph.CID`. **DW-16's direction is ruled** (glyph ids in
  `pagemodel`, CID allocation entirely inside `internal/pdf`) **but its owner is a later story**, and
  that relocation moves bytes on its own — bundling it here destroys the attributability this story
  exists to buy.
- It does **not** add format surface. `docs/` was grepped at `17f5f7a` for `baseline`, `leading` and
  `ascent`: **zero hits.** No `.folio` key, no schema change, no new document-level knob. A
  template author cannot set a first-baseline offset after this story any more than before it.
- It does **not** change `flipY`, its signature, or the "exactly one inverter" guard.
- It does **not** create `fixtures/shaped-text/thai-signoff.json`, does not fill in a reader or a
  date, and does not send a sign-off request. `TestShapedTextThaiSemanticSignOffIsRecorded` stays
  outstanding and the gate stays unable to pass. (D-000.28: *a claim written before the event it
  asserts is false from birth, and reads identically to a true one*.)
- It does **not** add a fifth Epic 2 gate obligation. See AC12.
- It does **not** fix `TestCorpusMeetsP6ExerciseFloors`.
- It does **not** flip `epic-2: backlog` in `sprint-status.yaml`. That key is the gate's. See
  *Flagged, not fixed*.

---

## Measured findings — read all of these before writing code

Every number below was produced at `17f5f7a`. Measurements 1 and 4 are **pre-fix captures under
D-000.30**: the window in which they are constructible closes permanently when the fix lands, and
they are recorded here as **one-time evidence that the defect was real — not as standing
red-proofs.**

### 1. The discrepancy, per shipped face, per shipped chain, per shipped size

Read from the `hhea` table via `(*fontset.Font).LineMetrics()`, scaled to the 1000-unit em by
`geom.ScaleRound`. Subject cited for every number (D-000.26).

**Per face**, in units of the 1000-em:

| face | file | `unitsPerEm` | `hhea` ascent | `hhea` descent | `hhea` lineGap | a − d + g |
|---|---|---|---|---|---|---|
| Noto Sans | `fonts/notosans/NotoSans-Regular.ttf` | 1000 | **1069** | −293 | 0 | 1362 |
| Noto Sans Thai | `fonts/notosansthai/NotoSansThai-Regular.ttf` | 1000 | **1061** | −450 | 0 | 1511 |
| Noto Sans SC | `fonts/notosanssc/NotoSansSC-Regular.ttf` | 1000 | **1160** | −288 | 0 | 1448 |
| Roboto-Regular (test face) | `folio-go/testdata/fonts/Roboto-Regular.ttf` | 2048 | **928** | −244 | 0 | 1172 |

**The face that wins the ASCENT maximisation is not the face that wins the ADVANCE maximisation.**
Over the shipped chain, `max(ascent)` is **Noto Sans SC's 1160**; `max(a − d + g)` is **Noto Sans
Thai's 1511**. Whose ascent the first baseline uses is therefore a real choice with a real
consequence, not a mechanical port of D-2.4.2 — see **DN-1**.

**Per shipped chain and size**, in millipoints. `now` is `run.FontSize` (`textdoc.go:689`'s current
fourth argument to `flipY`); `aligned` is `ScaleRound(max(ascent), fontSize, 1000)`:

| document · element | declared chain | size | first-baseline offset **now** | **aligned** | **Δ** | inter-baseline advance |
|---|---|---|---|---|---|---|
| `font-text` e1 | `["Roboto-Regular"]` | 14 pt | 14 000 | 12 992 | **−1 008** (baseline moves **UP** 1.008 pt) | 16 408 |
| `font-text` e2 (footer) | `["Roboto-Regular"]` | 9 pt | 9 000 | 8 352 | **−648** | 10 548 |
| `multi-script-fallback` e1 | Noto ×3 | 14 pt | 14 000 | 16 240 | **+2 240** (DOWN 2.24 pt) | 21 154 |
| `shaped-text` e1…e7 | Noto ×3 | 16 pt | 16 000 | 18 560 | **+2 560** (DOWN 2.56 pt) | 24 176 |
| `three-band-page` e1, e2 | `["Noto Sans"]` | 12 pt | 12 000 | 12 828 | **+828** | 16 344 |
| `three-band-page` e4 (header) | `["Noto Sans"]` | 9 pt | 9 000 | 9 621 | **+621** | 12 258 |
| `three-band-page` e3 (footer) | `["Noto Sans"]` | 8 pt | 8 000 | 8 552 | **+552** | 10 896 |
| `wrapped-text` e1…e4 | Noto ×3 | 11 pt | 11 000 | 12 760 | **+1 760** | 16 621 |

**The scale-dependence DW-15 predicted is confirmed, and it is linear.**
Δ = fontSize × (max(ascent) − 1000) / 1000, so the per-em coefficient is a property of the **chain**
alone:

| chain | coefficient | reading |
|---|---|---|
| `["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]` | **+0.160 em** | the shipped default; the *worst* case, driven by Noto Sans SC |
| `["Noto Sans"]` | **+0.069 em** | Latin-only, less than half the divergence |
| `["Roboto-Regular"]` | **−0.072 em** | **opposite sign** — a face whose `hhea` ascent is below its em |

So DW-15's "a Thai or CJK chain diverges more than a Latin one" is right in magnitude and
**incomplete in sign**: the error is not a consistent downward drift, it flips direction with the
face. **A guard asserting the baseline "sits lower than the font size implies" would be false on
`font-text`.** Assert the value, not the direction.

### 2. Exactly which goldens move — enumerated by measurement, not by inspection

A candidate fix was applied in a **throwaway `git worktree` at `17f5f7a`** and
`CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` was run over the whole module. Result:
**481 PASS · 7 FAIL** all-occurrences, **299 · 7** top-level, against the baseline's 487 · 1 / 305 · 1.

**The seven failures, complete:**

| test | fixture | verdict |
|---|---|---|
| `TestRenderMatchesFontTextGoldenFixture` | `fixtures/font-text/` | **MOVES** — the fifth golden DW-15 does not name |
| `TestMultiScriptFallbackGoldenFixture` | `fixtures/multi-script-fallback/` | **MOVES** |
| `TestShapedTextGoldenFixture` | `fixtures/shaped-text/` | **MOVES** — the sign-off-binding one |
| `TestThreeBandPageGoldenFixture` | `fixtures/three-band-page/` | **MOVES** |
| `TestWrappedTextGoldenFixture` | `fixtures/wrapped-text/` | **MOVES** |
| `TestThreeBandPageSemanticAcceptance` | `fixtures/three-band-page/` | the **only** semantic check with teeth on this defect — see finding 3 |
| `TestCorpusMeetsP6ExerciseFloors` | — | the pre-existing, required failure; byte-identical message and stats |

**And what does NOT move, also by measurement:**

- `fixtures/minimal-rect/` and `fixtures/image-embed/` — no text runs. Their tests stayed green.
- `fixtures/expected-breaks/expected_breaks.json` and `fixtures/thai-break-corpus/` — the break
  vector and the corpus. **No break-related test failed.** This is the measured confirmation of
  D-000.26 (refined)'s claim that the break sign-off is unaffected; do not carry it as an assumption
  (finding 4).
- `lint/` — **85 · 0 and 47 · 0, unchanged.** Every `lint` rule is blind to this change. **Claim no
  lint coverage for it** (D-000.42).
- `go vet ./...`, `gofmt -l`, and `go list -m all` (still exactly two modules) — clean under the
  candidate fix, including the added `pagemodel` field.
- **Every architecture guard stayed green** — `arch_test.go`, `arch_blindspot_test.go`,
  `bandcomposition_arch_test.go`, `flip_test.go`'s positive routing guard, `render_arch_test.go`.
  Adding a field to `pagemodel.TextRun` trips nothing. That is worth knowing *and* worth being
  uneasy about: no guard enumerates the page model's fields.

**Probe digests, for orientation only — NOT normative and NOT to be pasted into a fixture.** They
were produced by a deliberately rough candidate patch and will differ if DN-1 is ruled differently:

| fixture | committed at `17f5f7a` | probe produced |
|---|---|---|
| `font-text` | `5b60997e…3ebf338e` | `a69a6653…28612181` |
| `multi-script-fallback` | `20f3388a…eb15dbae` | `4699c8d7…790480b0` |
| `shaped-text` | `5964aad0…c92e00f` | `6c040ef7…c6c85370` |
| `three-band-page` | `2315855a…ada6d04f` | `746efcbc…885372bf` |
| `wrapped-text` | `3845da37…a712288e` | `3d41f462…89c3c22a` |

### 3. The population of places a golden digest is recorded is **four**, not one

Grepping each committed digest across the repository at `17f5f7a` (excluding `.git` and the
git-ignored `folio-go/.matrix-build/`):

| site | which fixtures | what it is |
|---|---|---|
| `fixtures/<f>/expected.pdf` | all 7 | the artifact |
| `fixtures/<f>/expected.json` | all 7 | the **normative** hash; also the file `matrixDocuments` reads per document |
| `folio-go/byte_neutrality_test.go:52` `fixtureDigests` | `minimal-rect`, **`font-text`**, `image-embed`, **`multi-script-fallback`**, **`shaped-text`** | a **SECOND, INDEPENDENT literal**, deliberately not read from the file it checks (D-2.3.4) |
| `fixtures/<f>/README.md` prose | `three-band-page`, `wrapped-text` | the digest quoted in the fixture's own documentation |

**`fixtureDigests` is the trap.** It is checked by `TestStory23aMovedNoGoldenDigest`, which compares
`expected.json`'s digest to the literal and **never re-renders** — so it stayed **green** in the
probe run (which moved no `expected.json`) and will go **red the moment the goldens are
re-recorded**. Its failure message reads, verbatim: *"Do not update this literal to make the test
pass — that is the move AD-21 and D-000.22 exist to prevent."*

That message is correct and this story is the exception it anticipates: the comment above the list
says the second literal exists so that *"if a golden is ever legitimately re-recorded, that is
exactly the conversation the second literal forces."* **This story is that conversation.** Three of
its five entries move; two do not. Its message is now stale in a way D-000.37 rules worse than a
stale comment — it names only Story 2.3a's premise, and a reader in 2026 hitting it after some
*future* accidental move would be told the wrong remedy. See AC7.

### 4. Only **one** of the five moving goldens has a semantic check that can see this defect

The module has exactly two semantic-acceptance tests, found by grepping `SemanticAcceptance` across
`folio-go/`:

| test | fixture | sees a first-baseline shift? |
|---|---|---|
| `TestThreeBandPageSemanticAcceptance` (`three_band_page_fixture_test.go:255`) | `three-band-page` | **YES** — went red in the probe, naming all four runs |
| `TestWrappedTextSemanticAcceptance` (`wrapped_text_fixture_test.go:75`) | `wrapped-text` | **NO** — stayed green |
| — | `font-text` | **no semantic acceptance test exists** |
| — | `multi-script-fallback` | **no semantic acceptance test exists** |
| — | `shaped-text` | **no semantic acceptance test exists** |

**Why `TestThreeBandPageSemanticAcceptance` has teeth**: it pins every run's whole `Tm` operator
against a hand-computed literal in `tbExpectedTm` (`three_band_page_fixture_test.go:238`), and
`tbExpectedTm`'s own doc comment (lines 213–232) hand-derives each number *from
`pageHeight − marginTop − run.Y − run.FontSize`*. Under the probe it reported exactly the deltas
finding 1 predicts:

```
run 0 (e4, pageHeader band): placement is "1 0 0 1 36 798.269 Tm", want the hand-computed "1 0 0 1 36 798.89 Tm"   (−0.621)
run 1 (e1, content   band): placement is "1 0 0 1 36 781.062 Tm", want the hand-computed "1 0 0 1 36 781.89 Tm"   (−0.828)
run 2 (e2, content   band): placement is "1 0 0 1 36 661.062 Tm", want the hand-computed "1 0 0 1 36 661.89 Tm"   (−0.828)
run 3 (e3, pageFooter band): placement is "1 0 0 1 36 51.448 Tm", want the hand-computed "1 0 0 1 36 52 Tm"       (−0.552)
```

Note the sign: `pdfY` is bottom-up, so a baseline moving **down the page** by 0.828 pt is a `pdfY`
**decrease** of 828 millipoints. The four values reconcile exactly with finding 1's table.

**Why `TestWrappedTextSemanticAcceptance` is blind, and why that is worse than "it has no
assertion".** It *does* have a leading assertion — and the assertion does not read the artifact:

```go
const wantAdvance = geom.Length(16621)
got, err := lineAdvance([]string{"Noto Sans", "Noto Sans Thai", "Noto Sans SC"}, geom.Length(11000), …)
if got != wantAdvance { … }
```

It calls the **production function** and compares it to a literal. Its own comment above says *"The
baselines are evenly spaced WITHIN an element by the ruled leading"* — a claim about the **artifact**
that nothing in the test checks against the artifact. The test computes `ys` (the distinct
`OriginYMilli` values, `wrapped_text_fixture_test.go:56`) and **discards them**, using their count
only. This is D-000.21 (sharpened) exactly — asserting on the thing you asked for rather than the
thing produced — and under D-000.42 the `lineAdvance` comparison is **redundant**, not proven: it
has no teeth `TestWrappedTextLayoutProperties` does not already have. **Do not count it as
coverage.** AC5 gives it real teeth by asserting the *observed* spacing between the `ys` it already
computes.

### 5. `internal/pdf` uses `run.FontSize` for two different jobs; only one of them is wrong

At `textdoc.go:689` and `:699`, `run.FontSize` is used **twice**: once as `flipY`'s `drawnHeight` (the
placement — **this is the defect**), and once as the `Tf` operand (the actual font size in the
content stream — **this is correct and stays**). The doc comment at `textdoc.go:640–676` already
states the defect verbatim, names DW-15, and — importantly — says it is *"not a provisional stand-in
and not this story's to change"*. **That sentence becomes false when this story lands**, and under
D-000.37 a stale remedy in a comment is worse than a stale comment. See AC8.

### 6. `lineAdvance` is called only when `len(lines) > 1`; the new offset is needed **always**

`render.go:335–341` computes `advance` only for multi-line elements, and its `present == 0` /
`maxUnits <= 0` error paths therefore never fire for a single-line element today. The first-baseline
offset is needed for **every** element with at least one line, so an unconditional call **widens the
set of inputs that can reach those two error paths**. Traced at `17f5f7a`: `shapeSegments` already
fails first for a chain with no present face, so `present == 0` looks unreachable through
`folio.Render` — but *"looks unreachable"* is precisely the claim D-000.9 says must be measured
rather than asserted. AC4 requires it to be measured, and requires the answer to be recorded either
way.

---

## DECISIONS NEEDED — escalate before development starts

### DN-1 (BLOCKING) — which maximisation does the first-baseline offset use?

**This decision changes the bytes of the artifact a pending human sign-off will bind to.** It must
be ruled before any golden is re-recorded, not after.

Three candidates, all defensible, all producing different documents. Measured on `shaped-text`
(Noto ×3 at 16 pt):

| candidate | offset | `shaped-text` first baseline |
|---|---|---|
| **A. `max(ascent)` over the present declared chain** | Noto Sans SC's 1160 → **18 560 mp** | 2.560 pt below the element top |
| **B. the ascent of the face that won the ADVANCE maximisation** | Noto Sans Thai's 1061 → 16 976 mp | 0.976 pt |
| **C. `max(a − d + g)` − `max(−descent)`** | 1511 − 450 = 1061 → 16 976 mp | 0.976 pt |

**Recommendation: A**, on D-2.4.2's own argument transplanted without modification — *"leading must
accommodate what MAY appear, not what DOES appear"*. The tallest ascent that may appear in an element
declaring the shipped chain is Noto Sans SC's 1160. Under B or C, an ideograph in the first line
overshoots the element's declared top by **99 units of the em** (0.99 pt at 10 pt) — a silent
overlap into whatever sits above, which is the same class of failure the chain-maximum leading rule
exists to prevent, and D-000.32's *"derive from the artifact; never fit to the sample"* points the
same way. B and C also make the offset depend on which face won a *different* maximisation, which
is a coincidence of the shipped set, not a rule.

**Recorded so it is not re-inferred:** B and C are numerically identical **on the shipped set only**
— `max(a − d + g)` and `max(−descent)` both land on Noto Sans Thai. They are different rules that a
fourth face would separate.

### DN-2 (non-blocking, byte-neutral today) — does `lineGap` belong in the first-baseline offset?

**Measured: `hhea.lineGap` is 0 on all three shipped faces and on the Roboto test face.** So
including it, excluding it, or splitting it half-above/half-below are **byte-identical on everything
this repository ships**, and no golden can distinguish them.

**This is D-000.39 (sharpened) territory and must be handled as such.** A mutation flipping the
`lineGap` term produces a **byte-identical artifact**. That is evidence of *equivalence on the
shipped set*, not evidence of a gap — but declaring equivalence requires **both** byte-identity
**and** a neighbouring observable mutation **at the same site** (e.g. perturbing the `ascent` term in
the same expression, which does move bytes). **Do not strengthen any golden assertion to try to
catch the `lineGap` term.** That manufactures a guard against a difference that does not exist on
this input set.

**Recommendation: exclude `lineGap` from the first-baseline offset** (it is extra space *between*
lines, which is where D-2.4.2 already spends it; spending it a second time above the first line
double-counts it), and give the decision teeth the only way it can have them — a **direct unit test
over fabricated `fontset.LineMetrics` values with a non-zero `LineGap`**, which requires the
arithmetic to be reachable without a `*fontset.Font`. See AC3.

### DN-3 (escalation only — DO NOT FIX HERE) — D-2.4.2's leading can under-space a mixed-face pair

Falls out of finding 1 and is **independent of this story's defect**. Over the shipped chain, the
advance is Noto Sans Thai's **1511**. But a Thai line's descender is **450** and the *next* line's
ideograph ascender is **1160**: 450 + 1160 = **1610 > 1511**, a potential **99-unit ink overlap**
between consecutive lines when consecutive lines resolve to different faces. This exists at
`17f5f7a`, is unchanged by this story (the fix shifts all baselines of an element by a constant and
leaves every inter-baseline gap exactly as ruled), and is out of the scope fence. **Raise as a new
DW entry for the owner; do not fix, and do not "improve" the leading rule while here.**

### DN-4 (needs a ruling before the sweep) — what happens to the dead `splitByFace`?

Per correction **C2**, `folio-go/render.go:463`'s `splitByFace` has zero callers. It positions face
segments on the old model. Three options:

- **(i) Delete it** and correct the five comments that describe it as live. Cleanest, and it removes
  a placement site that can never again disagree with the real one. But deleting a function is a
  change with no golden consequence, and bundling it here weakens *"one cause"* by exactly one item.
- **(ii) Update it consistently** with the new model, keeping it dead. Costs a parameter and leaves
  dead code that now needs maintaining.
- **(iii) Leave it alone** and say so. **Rejected**: that leaves a placement function computing
  baselines on a model the story just retired, with six live comments pointing readers at it — the
  half-updated hazard in its purest form.

**Recommendation: (i)**, taken as its own commit inside this story with its own line in the delivery
log, so the golden movement stays attributable to one cause and the deletion is separately
reviewable. **Lead's call.**

### DN-5 (needs a ruling) — heavy tests: is a D-000.4 per-story matrix override warranted?

**The criterion, stated first so the argument is measured against it and not around it:** D-000.4's
override is warranted when a story introduces **a new source of cross-target divergence** — float
arithmetic, a vendor call, a compressor, a new dependency — **not merely because it moves a golden.**
Story 2.5 declined on exactly this criterion for `three-band-page`.

**The honest assessment against the criterion: DECLINE.**

- The new arithmetic is `geom.ScaleRound(maxAscent, fontSize, 1000)` — the **same integer function**,
  in the **same package**, that `lineAdvance` has used since Story 2.4. No float enters; AD-23's
  type-aware `no-float-typed-value` rule reports zero on the candidate patch.
- No new vendor call. `(*fontset.Font).LineMetrics()` reads the `hhea` table that was already read
  and already scaled at `17f5f7a`; the vendor entry points this story touches are a **subset** of
  those `lineAdvance` already touches.
- No compressor, no new dependency; `go list -m all` stays at exactly two modules.

**The counter-argument, stated rather than buried, because it is not weak.** This story is a stronger
case than 2.5's integer band arithmetic in one respect: 2.5 moved **one** golden that no human was
waiting on, and this story moves **five**, one of which is the artifact D-2.3.5's pending reading
sign-off binds to. If a cross-target divergence *did* exist, it would surface at the Epic 2 gate —
**after** a Thai reader had been asked to certify the darwin-recorded bytes, and the human would be
spent twice for the exact reason D-000.41 was written.

But that is a **sequencing** argument, not a divergence-source argument, and it has a cheaper remedy
than a per-story matrix run: **the reading sign-off must not be requested until the gate has run the
four legs on the new `shaped-text` digest.** That obligation costs nothing and is recorded in AC13.

**So: decline the blanket override; propose no in-story matrix run.** If the lead prefers belt and
braces, the narrowest useful form is *the four legs for `shaped-text` alone*, on sign-off grounds
only — not the whole document set, and not on divergence grounds. **Lead's call.**

---

## Acceptance Criteria

> **AC1–AC4** are the fix. **AC5–AC9** are the golden movement and its acceptance. **AC10–AC15** are
> the gates, the guards and the honesty obligations.

### AC1 — The first baseline and the inter-baseline advance are derived from one chain walk

The offset the first baseline is placed at and the distance between consecutive baselines are
computed from the **same traversal of the declared chain**, in `folio-go/`, reading
`(*fontset.Font).LineMetrics()` and nothing else. There is exactly **one** place in the module that
decides which chain members are present, tolerates an absent member, and errors when none is present
— `lineAdvance`'s current loop body is that place, and the new arithmetic shares it rather than
duplicating it.

**Why this is an AC and not a style note**: `render.go:463`'s comment states the module's own rule
verbatim — *"the structural fix is not 'sum the right numbers' but 'have only one number' … there is
no second derivation left to drift."* Two independent chain walks answering two halves of one
geometric question is the defect this story is fixing, recreated one level up.

**Not satisfied by**: two functions that happen to loop over the same slice.

### AC2 — `internal/pdf` no longer derives a baseline from a font size

`buildTextContentStream` (`internal/pdf/textdoc.go`) passes a **carried layout quantity** as
`flipY`'s fourth argument for text, not `run.FontSize`. `run.FontSize` remains the `Tf` operand and
that use is unchanged (finding 5).

- `flipY`'s signature, arity, body and package are unchanged.
- `TestContentStreamYCoordinatesRouteThroughFlipY` still passes **and still counts a non-zero number
  of direct `flipY(...)` calls in `textdoc.go`** — its own vacuity guard
  (`flip_test.go:188`) must report the same or greater witness, and the reviewer must state the
  number observed.
- The quantity is computed in `folio-go/` and carried on the page model. It is **not** derived inside
  `internal/pdf` from `faces[run.Face]`, and **not** from `fontset.Metrics.Ascent` (inherited
  ruling 2).
- The value is identical for every run of one element, including runs on different faces and
  different lines — it is a function of the **declared chain and the size**, not of what was drawn.
  Assert this off the artifact: within one element, every run's `Tm` `ty` differs from the first
  line's by an exact multiple of the ruled advance and by nothing else.

### AC3 — The `lineGap` decision has teeth that no golden can give it

Per DN-2, the `lineGap` term is byte-neutral on every face this repository ships. Therefore:

- The pure arithmetic (given a set of `fontset.LineMetrics` and a font size, produce the offset) is
  reachable from a test **without constructing a `*fontset.Font`**, and a direct unit test drives it
  with **fabricated metrics carrying a non-zero `LineGap`** and asserts the ruled answer.
- **No golden assertion is strengthened in an attempt to observe `lineGap`.** Under D-000.39
  (sharpened), a `lineGap` mutation is expected to leave every artifact **byte-identical**, and that
  is evidence of *equivalence on this input set*.
- To be entitled to call it equivalence rather than a gap, the story must **also** demonstrate a
  **neighbouring observable mutation at the SAME SITE** — perturb the `ascent` term in the same
  expression and show the artifact moves. Both halves are reported; neither alone suffices.
- The `lineGap` unit test is labelled **proven** (its red-proof fires on fabricated metrics). It is
  **not** claimed as golden coverage.

### AC4 — The widened error path is measured, not assumed

Per finding 6, the first-baseline offset is needed for every element with at least one line, so its
chain walk runs for inputs that `lineAdvance` never saw. The story **measures** whether the
`present == 0` and `maxUnits <= 0` paths are reachable through `folio.Render` after the change:

- If reachable, a test drives each one through the **public entry point** and asserts the located
  error text.
- If unreachable, that is stated **with the reason and the symbol that fails first** (traced at
  creation: `shapeSegments`), and each unreachable branch is labelled a **forward guard with no
  available red-proof** (D-000.24) — never quietly credited with coverage.
- Either way, **no input that rendered successfully at `17f5f7a` returns an error afterwards.** State
  how that was checked.

### AC5 — Exactly five goldens are re-recorded, and the enumeration is by measurement

`fixtures/font-text/`, `fixtures/multi-script-fallback/`, `fixtures/shaped-text/`,
`fixtures/three-band-page/` and `fixtures/wrapped-text/` are re-recorded — **`expected.pdf` and
`expected.json` together**, per fixture.

`fixtures/minimal-rect/` and `fixtures/image-embed/` are **byte-unchanged**, and that is **asserted
by measurement, not by inspection** — the delivery log states their digests before and after.

**DW-15's "four" is wrong and `font-text` is the fifth** (correction C1). The delivery log states the
five as an enumeration produced by a full-suite run, with the before/after digest of each.

### AC6 — Every re-record carries a D-000.22 semantic acceptance step whose machine-checkable half is present

**Per fixture**, the story names the semantic property, says where it is read **off the artifact**,
and asserts it. Finding 4 measured that only `three-band-page` has such a check today.

| fixture | required |
|---|---|
| `three-band-page` | `tbExpectedTm`'s literals **and its hand-derivation comment** are recomputed on the new model. The comment currently derives every number from `pageHeight − marginTop − run.Y − run.FontSize`; leaving that text while changing the literals produces a derivation that does not reproduce its own answers. |
| `wrapped-text` | `TestWrappedTextSemanticAcceptance` gains an assertion **read off the artifact**: the first baseline's absolute `OriginYMilli`, and the observed spacing between the `ys` it already computes and currently discards. The existing `lineAdvance`-vs-literal comparison is **relabelled redundant** (D-000.42) or replaced; it is not counted as coverage either way. |
| `shaped-text` | a new machine-checkable acceptance step asserting the first baseline's absolute placement, read from the emitted `Tm` operators via the existing `readEmittedRuns` / `emittedRun.OriginYMilli` (`shaped_fixture_test.go:382`, `:325`). |
| `multi-script-fallback` | the same, using the same existing symbol. |
| `font-text` | the same — and this one is the **sign discriminator**: its baseline moves **UP**, so it is the fixture that falsifies a guard written as *"the baseline sits lower than the font size implies"* (finding 1). |

Every asserted number is a **hand-computed literal derived in a comment from the ruled formula**, in
the style `tbExpectedTm` already uses — not a value copied out of a failing test's output.

**D-000.33 applies**: an assertion that baselines are "evenly spaced" is satisfied trivially by a
single-line element. Any spacing assertion states and checks its **non-degeneracy precondition**
(at least two distinct baselines) before it means anything.

### AC7 — `fixtureDigests`' second literals are updated deliberately, and its message is corrected

Per finding 3, `byte_neutrality_test.go`'s `fixtureDigests` pins five digests as an independent
second copy, and `TestStory23aMovedNoGoldenDigest` will go red on three of them.

- The **three moved entries** (`font-text`, `multi-script-fallback`, `shaped-text`) are updated;
  `minimal-rect` and `image-embed` are **not touched**, and the delivery log says so explicitly.
- The mechanism stays: two literals, two files, so a future re-record still forces the conversation.
- **The failure message is rewritten.** It currently states only Story 2.3a's premise and instructs
  the reader *"Do not update this literal to make the test pass"*. After this story that instruction
  is executable-but-wrong for anyone who arrives at it — D-000.37: *a tripwire's failure message is
  executable by a human; a stale remedy is worse than a stale comment.* The new message names both
  authorised movements (2.3a's byte-neutrality premise, 2.5a's intended re-record) and keeps the
  prohibition on updating it to make a test pass.

### AC8 — Every comment describing the old model is corrected in the same commit

Measured at `17f5f7a`, these state the defect as current and become false when the fix lands:

- `internal/pdf/textdoc.go:640–676` — the `buildTextContentStream` doc comment (the formula at
  `:667`, the DW-15 sentence at `:672`), which spells out
  `pdfY = flipY(…, run.FontSize)`, calls it *"a known, RECORDED defect (DW-15)"*, and says it is
  *"not this story's to change"*.
- `internal/pagemodel/pagemodel.go` — the `TextRun` doc comment, if the carried quantity lands there.
- `three_band_page_fixture_test.go:213–232` — `tbExpectedTm`'s derivation (also AC6).
- `folio-go/three_band_page_fixture_test.go:216` — the same formula restated.
- The six production comments describing `splitByFace` as the live placement path (correction C2),
  per DN-4's ruling.

The reviewer greps for `run.FontSize` and for `DW-15` across `folio-go/` and reports every surviving
occurrence with a one-line justification.

### AC9 — DW-15's entry ends **fixed**, not carried

`deferred-work.md`'s DW-15 moves to the **Done** section with: the pre-fix measurement (finding 1's
tables, which cannot be reconstructed afterwards — D-000.30), the corrected golden count (**five**,
with `font-text` named), the ruling taken on DN-1, and the commit. **No open question is carried
forward under DW-15's number** (D-000.29). DN-3's overlap finding gets its **own new DW entry**; it
is not appended to DW-15.

### AC10 — Gates, re-measured, with the one permitted failure named

At the finished tree, from each module's directory, through the raw stream:

```
CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

- `folio-go/`: **exactly one** failure, `TestCorpusMeetsP6ExerciseFloors`, with **byte-identical**
  message and stats `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}` (Story 2.4 AC5).
- `lint/`: **zero** failures. Baseline 85 · 0 and 47 · 0.
- `go vet ./...` and `go vet -tags matrix ./...` clean in both modules; `gofmt -l` clean.
- `GOWORK=off go list -m all` reports **exactly two** modules.
- No `SOURCE_DATE_EPOCH` literal in any `.go` under `folio-go/` outside `testdata/` — **including
  inside error strings and comments**, which is where `absence-source-date-epoch` catches people.
- No compressor import; no `fmt` in `internal/pdf`; no `float32`/`float64` under `internal/` or the
  module root, checked by **both** guards (the syntactic one and `no-float-typed-value`).
- All-occurrences and top-level counts stated separately with the counting rule (D-000.26). Both are
  expected to **rise** by the number of tests added; a *fall* in either is a finding.

### AC11 — A mutation harness that proves it applied, before any test result is read

Every mutation this story reports follows **D-000.40 (sharpened)**: it **asserts it applied by
observing the artifact** — a non-empty diff or a reported count of occurrences changed — **before**
any test outcome is interpreted. An exit code is the tool's claim about itself.

- Use a **Python mutator that reports what it changed**. `sed 's/\bx\b/y/'` on BSD *succeeds at doing
  nothing*, and the suite then reports exactly what a sound guard would report. The named hazards are
  **`\b`, `-i`, `sort`, `grep -P`**.
- The same discipline applies to the count table itself: `rtk`'s filtered `go test` output contains
  no `--- PASS` lines, so any grep-based count taken through it is a fabrication (see the note under
  *Baseline*).

### AC12 — The Epic 2 gate owes the same **four** things, and each one's fate is stated

`declaredEpic2GateObligations` (`byte_neutrality_test.go:161`) is **unchanged** and
`TestEpic2GateObligationsMatchTheDeclaredSet` passes with a witness count no lower than 7-of-7. **No
fifth obligation is added** — the new semantic-acceptance tests are ordinary tests, not
`//go:build matrix` files, and no new document is registered in `matrixDocuments`.

The delivery log states, per obligation, in these words or better:

| obligation | what this story does to it |
|---|---|
| **the four-target matrix legs** | count unchanged; the **content** changes — 5 of the 7 registered `matrixDocuments` now compare against new `expected.json` digests, which the harness picks up automatically via `fixtureRelPath`. Nothing to edit. |
| **D-2.3.5's Thai READING sign-off** | **still outstanding, now bound to different bytes.** `thai-signoff.json` is still absent and this story must not create it. The obligation is not discharged, not weakened, and not re-scoped. |
| **D-2.4.3's Thai BREAK sign-off** | **untouched, and measured so** — no break-related test moved under the candidate fix (finding 2). D-000.26 (refined)'s claim is now confirmed rather than assumed. |
| **`three-band-page`'s deferred matrix legs** | **still deferred to the gate**, still declined under D-000.4's criterion (DN-5). They will compare the **new** digest. |

### AC13 — The sign-off request is held, and the condition for releasing it is written down

**D-000.41, binding.** This story moves the artifact the reading sign-off binds to, so:

- It does **not** request the sign-off and does **not** create the record.
- It appends to `epic-2-boundary-gate.md`: *the reading sign-off is requested only once no scheduled
  work is known to move `fixtures/shaped-text/expected.pdf` again*, and — per DN-5 — **only after the
  gate's four legs have agreed on the new digest**, so a reader is never asked to certify bytes a
  later leg proves target-dependent.
- **Story 2.6's creator carries D-000.41's own obligation**: measure whether 2.6 moves any
  sign-off-pending golden **before** any request is sent. This story states that hand-off explicitly
  so it is not lost between two story files.

### AC14 — D-000.34: check whether any test's discriminating power depended on the defect

*A test built on a bug dies with the bug — silently, because it goes on passing.* Enumerate, do not
sample. At minimum:

- `TestThreeBandPageSemanticAcceptance`'s *"exactly four distinct placements"* and
  `requireThreeBandPageUsesAllThreeBands`' *"four distinct baselines"* — traced at creation: the three
  sizes (12/9/8 pt) map to three distinct offsets (12 828 / 9 621 / 8 552 mp) under the new model, so
  distinctness survives. **Confirm it rather than inheriting this sentence.**
- Any test whose expectation is numerically equal to a font size because the old model made the
  baseline offset equal the font size. Grep for the sizes 8, 9, 11, 12, 14 and 16 in test literals
  near placement assertions.
- `wrapped_text_fixture_test.go`'s `linesByOrigin` consumers and `layout_probe_test.go`.

Report the sweep as an enumeration with a count of sites examined, not a sample.

### AC15 — Redundant checks are labelled, never counted

Any check this story finds or adds that reddens nothing because it is subsumed by another assertion
is labelled **redundant** (D-000.42) — *not* "unproven", *not* "forward guard". Keep it if it is
cheap belt-and-braces; **never count it as coverage.** The known instance to start from is
`TestWrappedTextSemanticAcceptance`'s `lineAdvance`-vs-literal comparison (finding 4). The story
reports, for every check it claims as coverage, which of the three categories it is in: **proven**
(red-proof fires) · **forward** (D-000.24, not constructible) · **redundant** (D-000.42).

---

## Task breakdown

1. **[x]** **Get DN-1 ruled.** Nothing downstream is worth doing until the offset formula is fixed — every
   re-recorded digest depends on it. Raise DN-2, DN-3, DN-4 and DN-5 in the same pass.
2. **[x]** **Re-take the baseline** at `17f5f7a` through the raw stream, including the P6 stats line, and
   record the table. Confirm 487 · 1 / 305 · 1 and 85 · 0 / 47 · 0 yourself.
3. **[x]** **Capture the pre-fix measurement** (D-000.30) — the per-face metrics and the per-chain,
   per-size delta table — into the story file **before touching production code**. This window
   closes permanently.
4. **[x]** **Implement AC1–AC4** in a worktree: the shared chain walk, the carried layout quantity, the
   `internal/pdf` change, the directly-testable pure arithmetic. Do not re-record anything yet.
5. **[x]** **Run the full suite and enumerate the movement** (AC5). Confirm five, confirm `minimal-rect` and
   `image-embed` are byte-unchanged, confirm the break vector is untouched.
6. **[x]** **Write the semantic acceptance steps first, re-record second** (AC6). A step written after the
   new bytes exist is a step that agrees with them by construction — D-000.22's whole point is that
   the recording moment is the only moment the question is answerable.
7. **[x]** **Re-record the five fixtures**, `expected.pdf` and `expected.json` together, and update
   `fixtureDigests` and both fixture READMEs (AC7, and finding 3's fourth population site).
8. **[x]** **Run the mutation set** under AC11's discipline: the observable `ascent` mutation at the offset
   site (must redden), the inert `lineGap` mutation at the same site (expected byte-identical), and
   the neighbouring-observable pair D-000.39 (sharpened) requires.
9. **[x]** **Sweep the comments** (AC8) and **sweep for defect-dependent discriminators** (AC14).
10. **[x]** **Close DW-15 as fixed** (AC9), open the DN-3 entry, append to `epic-2-boundary-gate.md`
    (AC12, AC13).
11. **[x]** **Re-measure every gate** (AC10), fill the delivery log, and set the story status to `review`.

**This story's breakdown ends at "status → review".** Committing is not one of its tasks.

---

## Heavy-test cadence — what is deferred, and to which gate

Per **DN-5**: the D-000.4 per-story matrix override is **DECLINED** on the stated criterion — this
story introduces **no new source of cross-target divergence**. The four-target legs for all seven
registered documents, `three-band-page`'s included, remain the **Epic 2 boundary gate's**, and they
will run against the re-recorded digests.

**The one obligation this adds to the gate is a sequencing one, not a new test** (AC13): the reading
sign-off is requested only after those legs agree on `shaped-text`'s new digest.

---

## Flagged, not fixed

`sprint-status.yaml` still reads **`epic-2: backlog`** while `2-1`…`2-5` are all `done`. **The epic
key is the gate's to flip**, and Story 2.5 flagged it in the same words. This story edits only its
own key. Carried forward so it is not lost a third time.

---

## Dev Agent Record — completion notes

### Scope widening received before development, and its effect on this file

Five rulings were made **after this story file was written** and they change what it builds. They are
applied in preference to the file's own text wherever the two disagree, and each disagreement is named
here rather than silently resolved:

| ruling | effect on this story as written |
|---|---|
| **D-2.4.2 (amended)** | **DN-1 is no longer a judgment** — it falls out of the corrected model as `max(ascent)`, candidate **A**. And the leading rule itself changes: `max(A−D+gap)` over the chain was **the wrong quantity**. |
| **D-2.5a.1** | **DN-3 is IN SCOPE, not an escalation.** The story's *"DO NOT FIX HERE"* is superseded. DW-15 and DN-3 are two symptoms of one defect. |
| **D-000.43** | DN-5's decline is **ratified**; AC13's sequencing substitute is ruled strictly better than an override. |
| **D-000.44** | AC6 is **binding for all five**, not aspirational. |
| **D-000.45** | AC6's sign-discriminator language is **hardened**: assert the computed value from a declarative table, never a direction. |
| **D-000.46** | **DN-4 is narrowed to (i) or (ii)**; (iii) is forbidden. Taken as **(i) delete** — see *Delivery Log*. |
| **D-000.47** | AC7 is **widened**: the digest lives at **four** sites and the check must read a declared list, not be a hand-maintained literal. |

**The corrected vertical model implemented here — one rule, three maxima, each maximised
INDEPENDENTLY over the declared chain:**

| span | value |
|---|---|
| top → first baseline | `max(A)` |
| baseline → baseline | `max(A) + max(D) + max(gap)` |
| last baseline → bottom | `max(D)` |

### Baseline, RE-MEASURED at `17f5f7a` (not inherited)

Taken through the raw stream (`rtk proxy "env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./..."`),
counted by a Python counter rather than a shell pipe (the `rtk`/`rg` hazard under AC11 bites both the
`go test` stream **and** `grep` itself — a bare `grep -c -- '--- PASS'` in this environment is
rewritten to `rg` and returns *"unrecognized flag"*, i.e. **0**, which reads exactly like a
catastrophic regression):

| scope | all-occurrences | top-level |
|---|---|---|
| `folio-go/` | **487 PASS · 1 FAIL** | **305 PASS · 1 FAIL** |
| `lint/` | **85 PASS · 0 FAIL** | **47 PASS · 0 FAIL** |

The single failure is `TestCorpusMeetsP6ExerciseFloors`, byte-identical to the required text:

```
corpus_test.go:189: P6g (opaque names) floor not met: got 7, need >=20
corpus_test.go:192: P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}
```

**Every figure the story file asserts about the baseline reproduces exactly.**

### PRE-FIX CAPTURE (D-000.30) — taken BEFORE any production file was edited

Produced by a throwaway white-box probe (`zz_prefix_probe_test.go`, deleted after capture) that reads
`(*fontset.Font).LineMetrics()` and **cross-checks itself against the production `lineAdvance`** at
every size before reporting — so these are the shipped function's own numbers, not a re-derivation
that might disagree with it. Subject cited for every row (D-000.26).

**Per face**, units of the 1000-em. **Reproduces the story's finding-1 table exactly, all four faces:**

| face | A | D | gap | A − D + gap |
|---|---|---|---|---|
| Noto Sans | 1069 | −293 | 0 | 1362 |
| Noto Sans Thai | 1061 | **−450** | 0 | **1511** |
| Noto Sans SC | **1160** | −288 | 0 | 1448 |
| Roboto-Regular | **928** | −244 | 0 | 1172 |

**Per chain — and this is the row the amendment turns on:**

| chain | max(A) | max(\|D\|) | max(gap) | OLD advance `max(A−D+g)` | NEW advance `maxA+maxD+maxg` | Δ |
|---|---|---|---|---|---|---|
| Noto ×3 | 1160 (SC) | 450 (Thai) | 0 | **1511** | **1610** | **+99** |
| `["Noto Sans"]` | 1069 | 293 | 0 | 1362 | 1362 | **0** |
| `["Roboto-Regular"]` | 928 | 244 | 0 | 1172 | 1172 | **0** |

**A finding that shapes the whole story, and which neither the story file nor the widening brief
states: the ADVANCE changes for the multi-face chain ONLY.** For a single-face chain
`max(A) + max(|D|) + max(g)` is *identically* `A − D + g` — one face cannot fail to supply both axes.
So DN-3's blast radius is confined to the Noto ×3 chain, and **within that, to elements that actually
have more than one line.** That is measured, not assumed, in the Delivery Log.

**Per chain and size** — `first OLD` is `run.FontSize` (the defect); `first NEW` is
`ScaleRound(max(A), size, 1000)`:

| chain | size | first OLD | first NEW | Δ first | adv OLD | adv NEW | Δ adv |
|---|---|---|---|---|---|---|---|
| Noto ×3 | 11 000 | 11 000 | 12 760 | **+1 760** | 16 621 | **17 710** | **+1 089** |
| Noto ×3 | 14 000 | 14 000 | 16 240 | **+2 240** | 21 154 | **22 540** | **+1 386** |
| Noto ×3 | 16 000 | 16 000 | 18 560 | **+2 560** | 24 176 | **25 760** | **+1 584** |
| `["Noto Sans"]` | 8 000 | 8 000 | 8 552 | +552 | 10 896 | 10 896 | 0 |
| `["Noto Sans"]` | 9 000 | 9 000 | 9 621 | +621 | 12 258 | 12 258 | 0 |
| `["Noto Sans"]` | 12 000 | 12 000 | 12 828 | +828 | 16 344 | 16 344 | 0 |
| `["Roboto-Regular"]` | 9 000 | 9 000 | 8 352 | **−648** | 10 548 | 10 548 | 0 |
| `["Roboto-Regular"]` | 14 000 | 14 000 | **12 992** | **−1 008** | 16 408 | 16 408 | 0 |

**The sign flips, and D-000.45's subject is real**: Roboto's `hhea` ascent is **928 < 1000**, so
`font-text`'s baselines move **UP** while every Noto-chain baseline moves **DOWN**. A guard phrased
*"the baseline sits lower than the font size implies"* is **false on `font-text`**, which ships. Every
assertion this story adds is therefore a **computed value from a declarative per-subject table**.

**The story file's own finding-1 delta column is confirmed correct in every row.** Nothing in it had
to be revised — only extended with the advance columns the amendment introduces.

## Delivery Log

### Gates, re-measured at the finished tree (AC10)

Raw stream (`rtk proxy "env CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./..."`), counted with a
Python counter — **never a shell pipe**, because `grep` is itself rewritten to `rg` in this
environment and `grep -c -- '--- PASS'` returns *"unrecognized flag"* → **0**, which reads exactly
like a catastrophic regression. That is D-000.40's portable-tooling hazard **one instrument further
out than the story file warned about**: the story flagged `rtk`'s interception of `go test`; the
counting tool is intercepted too.

| scope | counting rule | baseline `17f5f7a` | final | Δ |
|---|---|---|---|---|
| `folio-go/` | all occurrences | 487 · 1 | **503 · 1** | **+16 PASS**, failures unchanged |
| `folio-go/` | top-level (`^--- PASS`) | 305 · 1 | **311 · 1** | **+6 PASS**, failures unchanged |
| `lint/` | all occurrences | 85 · 0 | **85 · 0** | unchanged |
| `lint/` | top-level | 47 · 0 | **47 · 0** | unchanged |

**Both counts ROSE, as AC10 requires; neither fell.** +6 top-level is exactly the six tests added
(`TestVerticalModelArithmeticOverFabricatedMetrics`,
`TestVerticalModelRefusesAChainWithNoPresentFace`,
`TestVerticalModelRefusesANonPositiveLineHeight`,
`TestVerticalModelErrorPathsAreUnreachableThroughRender`,
`TestChainVerticalModelIsOneWalkFeedingBothSpans`,
`TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden`);
`TestStory23aMovedNoGoldenDigest` was **replaced** by `TestGoldenDigestAgreesAtEveryDeclaredSite`,
net zero.

**The one permitted failure, byte-identical to baseline** — compared line-for-line, not eyeballed:

```
corpus_test.go:189: P6g (opaque names) floor not met: got 7, need >=20
corpus_test.go:192: P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}
```

| other gate | result |
|---|---|
| `go vet ./...` — both modules | clean |
| `go vet -tags matrix ./...` — both modules | clean (so the un-run matrix legs still **compile** against the new model) |
| `gofmt -l` — both modules | clean |
| `GOWORK=off go list -m all` | **exactly two** modules |
| `SOURCE_DATE_EPOCH` literal in any `.go` under `folio-go/` outside `testdata/` | **0**, comments and error strings included |
| compressor imports | **0** under `folio-go/` outside `testdata/`, re-measured at finish |
| `fmt` in `internal/pdf` | **0** |
| `float32`/`float64` under `internal/` or module root | both guards green: `TestNoFloat64UnderInternal`, `TestNoFloat64UnderModule`, and the type-aware `no-float-typed-value` |

### What was built (AC1–AC4)

**One chain walk, one arithmetic, three spans.** `folio-go/wrap.go` now carries:

- `chainLineMetrics` — **the** place that walks the declared chain, decides presence, tolerates an
  absent member, and reads `(*fontset.Font).LineMetrics()` and nothing else.
- `verticalModel(chain, []fontset.LineMetrics, fontSize)` — the **pure** arithmetic. It takes
  metrics **as a value** and never touches a `*fontset.Font`. That signature is load-bearing, not
  tidy: it is the only way the `lineGap` term can ever be observed (AC3).
- `chainVerticalModel` — the production entry point: one walk feeding the one arithmetic.
- `lineAdvance` — kept under its existing name as a **projection** of `chainVerticalModel`, so every
  existing call site and test keeps working and there is no second derivation.

`verticalMetrics` carries all three spans. `LastDescent` (`max(D)`) has **no production consumer
today** and says so in its own doc comment; it is computed here so the ruled model is stated once
rather than re-derived by whoever first needs it (Story 2.8's clipping/overflow). It **is** asserted,
by `TestVerticalModelArithmeticOverFabricatedMetrics`.

**AC2 — `internal/pdf` no longer derives a baseline from a font size.** `textdoc.go:689` now passes
`run.BaselineOffset`, a carried layout quantity on `pagemodel.TextRun`, resolved in `folio-go/` from
the declared chain. `flipY`'s signature, arity, body and package are unchanged; only its fourth
argument changed. `run.FontSize` remains the `Tf` operand at `:716` and that use is untouched.
`TestContentStreamYCoordinatesRouteThroughFlipY` passes with its vacuity witness intact.

**AC4 — the widened error path, measured not assumed.** The model is now computed for **every**
element with at least one line, where the superseded code computed the advance only when
`len(lines) > 1`. Traced and asserted by
`TestVerticalModelErrorPathsAreUnreachableThroughRender`:

- `present == 0` is **UNREACHABLE** through `folio.Render`. The call sits after `shapeSegments` has
  already succeeded on non-empty text, and `resolveRuneFace` is a located error when no present
  chain member covers a rune — **`resolveRuneFace` is the symbol that fails first**, and the test
  asserts *which* error comes back, not merely that one does. An element with **empty** text never
  reaches either path: `collectTextRuns` short-circuits on `boundText == ""` one branch earlier, and
  does so **after** `fontChain` has validated the chain, so the widening cannot turn a
  previously-rendering empty element into an error.
- `maxUnits <= 0` is **UNREACHABLE** with any committed face (all four declare a positive ascent and
  a negative descent; `requireReadableTables` makes an absent `hhea` a load error).

**D-000.24's label would be WRONG here, and the distinction is the point of the seam.** Both paths
**are** red-proved — one level down, over fabricated metrics — because the arithmetic is reachable
without a font. They are **proven at the seam and unreachable through the public entry point**, and
neither half alone is the honest statement.

*No input that rendered at `17f5f7a` errors now*: asserted directly, by re-rendering all five
fixtures through the public entry point inside that test, and confirmed by the suite's 300+ other
`Render` callers staying green.

### The blast radius of the widened scope (D-2.5a.1), measured rather than assumed

D-2.5a.1 warned that DN-3 changes **line height**, so *"fewer lines fit a band and wrapping or
overflow may shift"*, and required the fixtures be checked explicitly. Checked, three ways:

1. **Structurally, wrapping cannot shift.** `packLines(segs, ops, totalRunes, fontSize, maxWidth)`
   takes **no vertical quantity**, and it is called at `render.go:345` — *before* the vertical model
   is computed at `:360`. The advance is not an input to line breaking and cannot become one without
   a signature change.
2. **Measured.** Every line-range and line-count assertion stayed green:
   `TestWrappedTextLayoutProperties`, `layout_probe_test.go`'s independent re-derivation of the
   horizontal half, and the per-element line counts in `wrappedTextExpectedLines` (3/2/2/2, total 9
   distinct baselines — unchanged).
3. **No vertical fitting exists to disturb.** `internal/layout/band.go` states that content taller
   than its band overflows visibly and that clipping/overflow diagnostics are Story 2.8's. Nothing
   in the repository counts lines against a band height.

**And a finding neither the story file nor the widening brief states, which bounds the whole thing:
the ADVANCE changes for a MULTI-FACE chain only.** For a chain resolving to one present face,
`max(A) + max(|D|) + max(gap)` is *identically* `A − D + gap` — one face cannot fail to supply both
axes. Measured: Noto ×3 moves 1511 → 1610 units; `["Noto Sans"]` and `["Roboto-Regular"]` do not
move at all. Of the three fixtures declaring the Noto ×3 chain, **only `wrapped-text` has a
multi-line element**, so `wrapped-text` is the **only artifact in the repository** on which the
amendment is observable — confirmed by mutation C below.

### AC5 — exactly five goldens re-recorded, enumerated by measurement

Full-suite run after the fix, before any re-recording: **476 · 12** all-occurrences (297 · 9
top-level). The failures were the five golden hash comparisons, `TestThreeBandPageSemanticAcceptance`,
`TestWrappedTextSemanticAcceptance`, `TestLineAdvanceIsTheMaxOverTheDeclaredChain` (+3 subtests) and
the required P6 failure. **Nothing else moved.**

| fixture | before | after | bytes | verdict |
|---|---|---|---|---|
| `minimal-rect` | `0f925e1b…6cb4f7c` | `0f925e1b…6cb4f7c` | 547 → 547 | **BYTE-UNCHANGED** |
| `font-text` | `5b60997e…3ebf338e` | `a69a6653…28612181` | 22 310 → 22 315 | MOVED |
| `image-embed` | `e5778eb8…abe689fc` | `e5778eb8…abe689fc` | 995 → 995 | **BYTE-UNCHANGED** |
| `multi-script-fallback` | `20f3388a…eb15dbae` | `4699c8d7…790480b0` | 55 086 → 55 086 | MOVED |
| `shaped-text` | `5964aad0…3c92e00f` | `6c040ef7…c6c85370` | 91 059 → 91 059 | MOVED |
| `three-band-page` | `2315855a…ada6d04f` | `746efcbc…885372bf` | 54 445 → 54 452 | MOVED |
| `wrapped-text` | `3845da37…a712288e` | `277bc5c0…056b0ad5` | 72 743 → 72 738 | MOVED |

`minimal-rect` and `image-embed` are **asserted byte-unchanged by re-hashing**, not by inspection —
their digests above are the sha256 of the committed files taken before and after.

**DW-15's "four" was wrong and `font-text` is the fifth** — confirmed live, not inherited.

**An independent confirmation of the multi-face finding, from the digests themselves.** The story
file recorded probe digests from a DW-15-*only* candidate patch. Four of the five landed **exactly**
on those probe values (`a69a6653…`, `4699c8d7…`, `6c040ef7…`, `746efcbc…`). The fifth,
`wrapped-text`, did **not** (probe `3d41f462…89c3c22a`, actual
`277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5`) — because it is the only fixture
the DN-3 half also touches. The arithmetic predicted that before the bytes existed.

**Corrected at finish (Finding 4).** The `wrapped-text` "after" cell above previously read
`277bc5c0…89c3c22a` — head from the real digest, tail from the *probe* digest, spliced. The
conclusion it supports was and is correct: the two values differ entirely. Re-measured at finish,
subject cited: `shasum -a 256 fixtures/*/expected.pdf` from the repo root gives
`277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5  wrapped-text/expected.pdf`, and
`goldenDigestRecord` plus `fixtures/wrapped-text/expected.json` both already carried that value, so
no code was affected. Recomputed, not transcribed (D-000.18).

**Break vector untouched, measured:** `fixtures/expected-breaks/` and `fixtures/thai-break-corpus/`
are byte-unchanged and no break test moved at any point. D-000.26 (refined)'s claim is now
**confirmed rather than assumed** (AC12).

### AC6 / D-000.44 — semantic acceptance for all five, written BEFORE re-recording

**Ordering matters and was honoured**: every acceptance step was written and made green *against the
un-recorded tree*, then the goldens were recorded. A step written after the new bytes exist agrees
with them by construction.

`folio-go/first_baseline_acceptance_test.go` drives all five from one declarative table.
**All 23 baselines, across 18 elements of the five fixtures, matched their hand-derived literals on
the FIRST run** — they were derived from the ruled formula in the comment beside each, not copied
from output. (Counting rule, corrected at finish under Finding 7(b): the table declares **18
elements** carrying **23 baselines**; `wrapped-text` alone contributes 9 baselines from 4 elements.
Counted from the table itself, not restated.)

| fixture | had a semantic step at `17f5f7a`? | now |
|---|---|---|
| `three-band-page` | yes, with teeth | `tbExpectedTm` **and its hand-derivation comment** recomputed on the new model; the comment now derives from `run.BaselineOffset`, not `run.FontSize` |
| `wrapped-text` | yes, but **blind** | **repaired** — see below |
| `font-text` | **none** | added; **the sign discriminator** |
| `multi-script-fallback` | **none** | added; also proves the offset is per-**element**, not per-run (three face segments, one baseline) |
| `shaped-text` | **none** | added; seven elements |

**The blind one, repaired.** `TestWrappedTextSemanticAcceptance` called `lineAdvance`, compared it to
a literal, and — having already computed `ys`, the distinct baselines of the produced document —
**used only their count and threw the values away.** It now asserts the **observed** spacing between
those `ys`, with a stated non-degeneracy precondition (D-000.33) and a reported count of intervals
actually compared (**5**). **Counting rule** (corrected at finish under Finding 3, which recorded 8):
**intra-element intervals only** — the fixture's four elements carry 3/2/2/2 lines, giving
2+1+1+1 = 5. Eight would be every consecutive pair across the flat 9-baseline list, which folds in
three **cross-element** gaps that are not the ruled advance at all, so 8 would describe an assertion
that would be wrong to make. Re-measured at finish, subject cited:
`CGO_ENABLED=0 GOWORK=off go test -count=1 -v -run TestWrappedTextSemanticAcceptance ./...` prints,
at `wrapped_text_fixture_test.go:194`, *"9 text runs across 9 baselines […]; 5 OBSERVED
inter-baseline intervals all equal the ruled 17710 mp (superseded rule would give 16621)"*. The code
was correct; the number in this log was wrong. The `lineAdvance`-vs-literal comparison is **kept and explicitly labelled
REDUNDANT** (D-000.42) in both its comment and its failure message, and is **not counted as
coverage**.

**D-000.45 honoured throughout.** Every expectation is a computed value from a declarative
per-subject table. Two vacuity guards enforce it: the acceptance table fails unless **both** an
element whose baseline sits lower than its point size **and** one that sits higher are present
(measured: 16 lower, 2 higher — the 2 are `font-text`'s, Roboto `max(A) = 928 < 1000`). A guard
phrased as a direction would be **false on `font-text`, which ships.**

### AC3 / D-000.39 (sharpened) — the `lineGap` term, and the mutation pair

`TestVerticalModelArithmeticOverFabricatedMetrics` drives `verticalModel` over five fabricated
subjects, asserting all three spans against hand-derived literals, plus three vacuity guards: at
least one subject must discriminate the amended rule from the superseded one (3 do), at least one
must carry a **non-zero** `max(gap)` (1 does — the only place in the repository where that term is
observable), and both first-baseline directions must occur.

**No golden assertion was strengthened to try to observe `lineGap`.**

### AC11 — the mutation set, each asserting it applied before any test result was read

Python mutator (`sed 's/\b…/…/'` on BSD *succeeds at doing nothing*; `\b`, `-i`, `sort` and `grep -P`
all avoided). Each mutation reported the file sha256 before/after **and the diff hunk** and refused
to interpret a test result unless the artifact had demonstrably changed. Restored by `cp` from a
byte-for-byte backup — never `git checkout`, which would have discarded this story's real work in
the same file — with the restored sha256 compared against the pristine one.

| # | mutation, all at the same site | goldens | fabricated-metrics unit test | verdict |
|---|---|---|---|---|
| **A** | perturb the `max(ascent)` term of the first-baseline span | **RED** | **RED** | as predicted |
| **B** | drop the `max(lineGap)` term from the advance span | **byte-IDENTICAL** | **RED** | as predicted |
| **C** | reinstate the superseded `max(A − D + gap)` advance | **RED** | **RED** | as predicted |

**B alone would prove nothing.** Byte-identity is evidence of *equivalence on this input set* only
alongside **A** — a **neighbouring observable mutation in the same expression**, proving the site is
reachable and the goldens are sensitive to it. Both halves are reported; neither alone suffices.

**C, enumerated rather than summarised**, because *which* goldens it reddens is the load-bearing
claim: **`TestWrappedTextGoldenFixture`, `TestWrappedTextSemanticAcceptance`, and the `wrapped-text`
leg of the acceptance table — and nothing else.** `font-text`, `multi-script-fallback`,
`shaped-text` and `three-band-page` stayed **green**. That is the amendment's blast radius, measured.

### AC7 / D-000.47 — the digest now lives at four DECLARED sites, and the check reads the list

`fixtureDigests` + `TestStory23aMovedNoGoldenDigest` are replaced by `goldenDigestRecord` +
`TestGoldenDigestAgreesAtEveryDeclaredSite`, which asserts three things — the third being one the
old shape **could not**:

1. the **artifact** re-hashes to the declared digest (the old test never re-hashed anything: a
   re-record moving the PDF *and* its JSON together would have left it green);
2. every **declared** site records that digest;
3. **no undeclared site does** — the completeness half, which scans the declared search scope for
   each digest *value*.

**The completeness half caught a real gap on its first run**: with all seven fixtures now in the
declaration, `three-band-page` and `wrapped-text` acquired a second literal that I had not declared
as a site. **Owner-visible consequence:** the independent-second-literal discipline now covers **all
seven** fixtures where it covered **five** — those two had *no* second copy of their digest at all.

**Search scope is declared, and `_bmad-output/` is deliberately excluded** with the reason stated in
the code: story files record digests as **past-tense measurements**, and rewriting them would destroy
the record of what produced which bytes.

**The three moved entries** (`font-text`, `multi-script-fallback`, `shaped-text`) were updated;
**`minimal-rect` and `image-embed` were not touched.**

**AC7's message rewrite (D-000.37) — a TRUE guard carrying a remedy that FORBADE the correct
action.** The old text said *"Do not update this literal to make the test pass"*; this story had to
update three of them. The new `goldenDigestRemedy` names both authorised movements (2.3a's
byte-neutrality premise, 2.5a's re-record), keeps the prohibition, and restates it as the **rule**
rather than one story's premise: *updating a digest is legal only as a deliberate, attributable
re-recording, never as the way to make a red test go green.*

**The fourth site class also carried a fifth stale thing**: both READMEs quote the **byte count**
alongside the digest. Updated (54 445 → 54 452; 72 743 → 72 738, on all four matrix lines).

### AC8 — every comment describing the old model, corrected in the same commit

| site | done |
|---|---|
| `internal/pdf/textdoc.go:640–676` | rewritten. States the formula with `run.BaselineOffset`, why it must never become `run.FontSize` again, and why it may not be re-derived from `faces[run.Face]` (AD-5 + AD-24). |
| `internal/pagemodel/pagemodel.go` | `BaselineOffset` documented: why carried, why not derivable from `Face`, why not `FontSize` — with the measured per-face numbers and the explicit note that the old proxy **had no consistent direction**. |
| `three_band_page_fixture_test.go:213–232` | derivation **and** literals recomputed together — leaving the prose while changing the literals would have produced a derivation that does not reproduce its own answers. |
| `wrap.go` | `lineAdvance`'s doc comment replaced by the full amended-model exposition. |
| the `splitByFace` comments | see DN-4 below. |
| `fixtures/three-band-page/README.md`, `fixtures/wrapped-text/README.md` | both described the superseded rule in prose (`pdfY = … - fontSize`; *"maximum `ascent - descent + lineGap`… 1511/1000 em, 16,621 mp"*). Both corrected. **Neither is a `.go` file, and neither would have been found by grepping code.** |

**Sweep result**: every surviving `run.FontSize` in `folio-go/` is either the `Tf` operand
(`textdoc.go:716`, correct and unchanged), template style parsing, or a comment *about* the fix.
Every surviving `DW-15` mention is past-tense.

### DN-4 / D-000.46 — `splitByFace` deleted

Measured at `17f5f7a`: **zero call sites**, in production or test, while **ten** references
described it as the live placement path — **seven production comments** (`render.go`'s six plus
`internal/fontset/fontset.go:576`), `wrap.go`'s two and `segment_origin_test.go`'s one. It computed
placement on the **old** model.

*(Corrected at finish: this said **nine**, and C2 said six. Both were wrong, in opposite directions,
and neither matched the code. `git grep -c splitByFace 17f5f7a -- folio-go` gives 11 matching lines
across four files; one is the declaration, so ten are references. Finding 7(a) named C2's omissions
but endorsed DN-4's nine as correct — it is not, and the reviewer's own tally is not authority over a
re-measurement.)*

D-000.46 forbids option (iii) and leaves (i) delete or (ii) update-and-keep-dead. **Taken as (i)**,
per DN-4's own recommendation and because (ii) means maintaining dead code that D-000.46 exists to
condemn. The function's entire body was `shapeSegments` + `positionSegments`, both live. All ten
references were re-pointed at the functions that actually execute. **Zero occurrences of
`splitByFace` remain under `folio-go/`** — 11 → 0, measured by walking the whole module.

*(Scoped at finish under Finding 6, which read the previous wording — "anywhere in the repository" —
as false. It was.

**The durable fact, and the only load-bearing one, is `folio-go/`: 11 → 0.** Every surviving
occurrence is `_bmad-output/` markdown — this story file, the decision log, `deferred-work.md`, and
the 2.2 / 2.3 / 2.4 story files: **six files**, and nothing else in the repository. Those are
**past-tense narrative and must survive**, by exactly the reasoning `goldenDigestSearchScope` uses to
exclude `_bmad-output/`: a record of what the code once was is not a stale pointer.

**The TOTAL across those six files is deliberately not asserted here**, and that is the finding's
real lesson. The reviewer counted 20; a count taken while writing this resolution gave 25; one taken
after the decision-log entries were appended gave 35. **The number rises every time a document
narrates the deletion — including this sentence.** A record is not a consumer, so the counting rule
(*occurrences under `folio-go/`*, which is 0, versus *occurrences in `_bmad-output/`*, which are
history) is what survives re-measurement; the sum of the second is not a fact about the code.)*

*Flagged for the reviewer:* this is the one item in the change set with **no golden consequence**,
and it is separately reviewable — the deletion and the comment re-pointing are the whole of it.

### AC14 / D-000.34 — the sweep, enumerated with a count

**140 `.go` files examined** (`testdata/` excluded), searched for every value whose meaning depended
on the superseded model: the old `Tm` literals (`798.89`, `781.89`, `661.89`, `52`), the old `pdfY`
values for all five fixtures, and the old advances (`16621`, `21154`, `24176`, `1511`).

**30 occurrences found. Every one is deliberate**: a reference to the superseded rule as the
hypothesis being discriminated against, a still-true per-face fact (`1511` *is* Noto Sans Thai's
`A − D + gap`, and single-face chains still answer `1511`), or a coincidence (`769890` is the
printable height). **No live expectation under the superseded model survives anywhere.**

**The named items, confirmed rather than inherited:**

- **`TestThreeBandPageSemanticAcceptance`'s "exactly four distinct placements"** — survives. The
  three sizes (12/9/8 pt) map to three distinct offsets (12 828 / 9 621 / 8 552 mp), so the four
  placements 798.269 / 781.062 / 661.062 / 51.448 remain distinct. Green.
- **`requireThreeBandPageUsesAllThreeBands`' "four distinct baselines"** — survives. Its band
  thresholds are `y >= 793890` (header) and `y <= 66000` (footer); the new baselines classify
  798 269 → header, 781 062 and 661 062 → content, 51 448 → footer. **Compile-checked under
  `go vet -tags matrix`; not executed** (heavy).
- **`wrapped_text_fixture_test.go`'s `linesByOrigin` consumers** — the function is now *load-bearing*
  rather than count-only, which is the repair itself.
- **`layout_probe_test.go`** — purely horizontal (widths, `packLines`); no vertical quantity, no
  baseline. Unaffected, and its staying green is independent evidence that line breaking did not
  move.

**AND ONE THE STORY FILE DID NOT ANTICIPATE — the clearest D-000.34 instance in the change set.**
`TestLineAdvanceIsTheMaxOverTheDeclaredChain`'s **negative control** was the row labelled *"Thai
first — the one order where (a) and (b) agree"*: under the superseded rule a Thai-first chain gave
1511 by both hypotheses. **The amendment destroys that property** — no single-face-first ordering can
produce 1610, because 1610 requires two faces — so that row silently became a *discriminating* case
and the table would have been left **with its negative control gone and nobody told**. The table now
carries **two** rival hypotheses (`firstFace` and `superseded`), the negative control moved
explicitly to the single-face rows where all three agree **by construction**, and three vacuity
guards fail if any of the three roles is unfilled. It reports 3 / 5 discriminating vs first-face,
3 / 5 vs the superseded rule, and 2 negative controls.

### AC15 — every check this story claims, categorised

| check | category |
|---|---|
| `TestVerticalModelArithmeticOverFabricatedMetrics` | **PROVEN** (mutations A, B, C all redden it) |
| `TestFirstBaselineSemanticAcceptanceAcrossEveryReRecordedGolden` | **PROVEN** (A and C redden it) |
| the five golden hash comparisons | **PROVEN** (A and C redden them) |
| `TestWrappedTextSemanticAcceptance`'s **observed-spacing** assertion | **PROVEN** (C reddens it) |
| `TestWrappedTextSemanticAcceptance`'s `lineAdvance`-vs-literal comparison | **REDUNDANT** (D-000.42) — kept as belt-and-braces, **not counted** |
| `TestVerticalModelRefusesAChainWithNoPresentFace` / `…RefusesANonPositiveLineHeight` | **PROVEN at the seam**, over fabricated metrics — and the corresponding paths are **unreachable through `folio.Render`**. Both halves stated; neither alone. |
| `TestGoldenDigestAgreesAtEveryDeclaredSite` completeness half | **PROVEN** — it reddened on a real undeclared site during development |
| `lint/`'s 85 checks | **NO COVERAGE CLAIMED.** 85 · 0 / 47 · 0 before and after; every `lint` rule is blind to this change (D-000.42). |

### AC12 / AC13 — the gate, and the human

Recorded in full in `epic-2-boundary-gate.md`. In brief: **the gate owes the same four things and no
fifth** (`declaredEpic2GateObligations` is byte-unchanged;
`TestEpic2GateObligationsMatchTheDeclaredSet` green with a 7-of-7 witness). **`thai-signoff.json` was
NOT created**, no sign-off was requested, and the release condition is written down — *the reading
sign-off is requested only once no scheduled work is known to move the artifact **and** the gate's
four legs have agreed on the new digest*. Story 2.6's creator carries D-000.41's own obligation, and
that hand-off is stated in the gate file so it is not lost between two story files.

**Heavy tests: the override was DECLINED and the legs were NOT run.** Stated plainly rather than left
to inference. Same integer `geom.ScaleRound`, same package, a **subset** of the vendor entry points
`lineAdvance` already touched, no float, no compressor, no new dependency, still exactly two modules.

### Flagged, not fixed

- **`sprint-status.yaml` still reads `epic-2: backlog`** while `2-1`…`2-5` are `done`. **That key is
  the gate's to flip.** This story edited only its own key (`ready-for-dev` → `review`). Flagged for
  the third consecutive story.
- **`TestCorpusMeetsP6ExerciseFloors` remains red**, byte-identically, per Story 2.4 AC5. Not fixed,
  not skipped, no corpus item added.
- **The pre-existing dirty files** — `folio-mvp-decision-log.md` and the original `sprint-status.yaml`
  hunk — were neither reverted nor committed. They are the artifacts that authorised this story.

### Deviations from the story file as written, each with its authority

| the story file says | what was done | authority |
|---|---|---|
| DN-1 is **BLOCKING**, needs a ruling | Not escalated. It **dissolved**: `max(ascent)` falls out of the corrected model. | **D-2.4.2 (amended)** |
| DN-3 is *"escalation only — DO NOT FIX HERE"* and *"raise as a new DW entry"* | **Fixed here**, and **no new DW entry opened.** | **D-2.5a.1** |
| AC9: *"DN-3's overlap finding gets its own new DW entry"* | Superseded — DW-15's Done entry records DN-3 as fixed, not carried. | **D-2.5a.1** + D-000.29 |
| AC7: *"the mechanism stays: two literals, two files"* | Kept, and **widened** to all seven fixtures rather than five, as a consequence of declaring the list once. | **D-000.47** |
| DN-4 *"Lead's call"* | Taken as **(i) delete**. | **D-000.46** narrows to (i)/(ii); DN-4 recommends (i) |
| `folio-format.md` unchanged (*"does not add format surface"*) | **The leading section was rewritten** to state the corrected three-maxima model. This is the one permitted format change: correcting existing wording, not adding surface. No new key; the section now ends *"There is no `lineHeight` key and no first-baseline key."* | the widening brief, explicitly |

**No decision was guessed.** Every fork this story faced was either ruled in the decision log before
development began, or determined by an existing convention, and each is named above with its
authority.


## QA Results

**Reviewed by:** bmad-code-reviewer · **Date:** 2026-08-24 · **Baseline:** `17f5f7a`, work uncommitted
**Method:** every claim below was verified by construction — the suites were re-run in both modules,
five mutations were applied with a Python mutator that asserted the artifact changed (sha256
before/after + diff hunk) before any test result was read, and every mutation was reverted by hand
(`cp` from a byte-exact backup, never `git checkout`/`git stash`). Final `git status --porcelain`
reports the same 30 entries handed to the reviewer, with no `.bak` and no probe residue.

### Instrument note (confirms the story's own warning, one step further)

The story warns that `rtk` rewrites `go test`. Confirmed, and **`rtk proxy` additionally cannot
parse an env-var prefix** — `rtk proxy "CGO_ENABLED=0 … go test …"` fails with
*"Failed to execute command: CGO_ENABLED=0: No such file or directory"* and produces an empty
stream, which counts as 0 PASS / 0 FAIL and reads as a total collapse. Also confirmed: bare `grep`
is rewritten to `rg`, and a pattern beginning `--` returns *"unrecognized flag"*. Every count below
was taken by running a **shell script** through `rtk proxy` (env set inside the script, `cd` inside
the script) and counted with **Python**, never a shell pipe. The story's counting discipline is
correct and necessary.

### Gates, re-measured independently

| scope | counting rule | expected | **observed** | verdict |
|---|---|---|---|---|
| `folio-go/` | all occurrences | 503 · 1 | **503 · 1** | reproduces |
| `folio-go/` | top-level | 311 · 1 | **311 · 1** | reproduces |
| `lint/` | all occurrences | 85 · 0 | **85 · 0** | exactly baseline |
| `lint/` | top-level | 47 · 0 | **47 · 0** | exactly baseline |

The single failure is `TestCorpusMeetsP6ExerciseFloors`, **byte-identical** to baseline:
`P6g … got 7, need >=20` / `P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`.
**No second failure.** `gofmt -l` clean, `go vet ./...` and `go vet -tags matrix ./...` clean in
**both** modules, `GOWORK=off go list -m all` reports **exactly two** modules for `folio-go`.
`SOURCE_DATE_EPOCH` under `folio-go/` outside `testdata/`: **0**. `fmt` in `internal/pdf`: **0**.
`fixtureDigests` and `TestStory23aMovedNoGoldenDigest`: **0 occurrences** — cleanly replaced.

### The seven attack points — each verified by construction

**1 — The central narrowing claim: VERIFIED, algebra and both reproductions.**
The algebra holds: in `verticalModel` the maxima are seeded at zero and taken independently, so for a
chain with one present face declaring `A>0, D<0, gap>=0`, `max(A)+max(|D|)+max(gap)` reduces term for
term to `A − D + gap`. The advance therefore moves for a multi-face chain only.
**Mutation C** (reinstate `max(A−D+gap)`, applied at the same site, artifact change observed) reddened
exactly: `TestWrappedTextGoldenFixture`, `TestWrappedTextSemanticAcceptance`, the **`wrapped-text` leg**
of the acceptance table — and, as the story's own mutation table predicts, the arithmetic unit tests
(`TestVerticalModelArithmeticOverFabricatedMetrics` ×3 subtests, `TestChainVerticalModelIsOneWalkFeedingBothSpans`,
`TestLineAdvanceIsTheMaxOverTheDeclaredChain` ×3). **`font-text`, `multi-script-fallback`, `shaped-text`
and `three-band-page` stayed green.** The blast radius claim is true as scoped ("which *goldens*").
Independently confirmed from the artifacts: four of the five committed digests (`a69a6653…`,
`4699c8d7…`, `6c040ef7…`, `746efcbc…`) land **exactly** on the story's DW-15-only probe values, and
`wrapped-text` alone does not — see Finding 4 for a transcription defect in how that fifth value is
recorded, which does not affect the conclusion.

**2 — D-000.45: VERIFIED, guard red-proved.**
Every expectation is a computed value from a declarative table. The table checks its own premises
(assertion 1: `wantFirstBaselineMP == ScaleRound(maxAscent1000, fontSize, 1000)`), then production
(assertion 2), then **reads the baselines off the produced PDF's `Tm` `ty` operands** via
`readEmittedRuns`/`linesByOrigin` (assertion 3). Direction is never asserted.
**Red-proof:** replacing the sign discriminator with a same-direction duplicate (keeping the table at
five, so the `len != 5` guard could not mask it) made the direction guard the **sole** failure:
*"vacuity: 20 elements place the first baseline further down … and 0 place it further UP"*. Reverted.
Pristine counts confirmed: 16 down / 2 up, and the 2 are `font-text`'s (Roboto `max(A)=928`).
See Finding 2 — the success log states these two directions backwards.

**3 — D-000.44: VERIFIED for all five.**
`first_baseline_acceptance_test.go` drives all five fixtures from one table, asserting computed
baselines read off the artifact. **Mutation A** (perturb the `max(ascent)` term, same site, artifact
change observed) reddened **all five golden hash comparisons and all five acceptance legs**. The
repaired `wrapped-text` step genuinely asserts the *observed* spacing between the `ys` it previously
discarded, with a stated non-degeneracy precondition and a reported interval count, discriminating
explicitly against the superseded 16621; it reddens under mutation C. The `lineAdvance`-vs-literal
comparison is labelled REDUNDANT in both comment and failure message and is not counted.

**4 — AC6 ordering: CORROBORATED by independent evidence, with a caveat.**
Filesystem mtimes are consistent with the claim and inconsistent with back-filling:
`vertical_model_test.go` 20:35:06 → `first_baseline_acceptance_test.go` 20:36:59 →
`three_band_page_fixture_test.go` 20:37:35 → `wrapped_text_fixture_test.go` 20:38:02 → **all five
`expected.pdf` re-recorded 20:39:28** → `expected.json` 20:39:47. Every acceptance step was last
written **before** the goldens existed and **none was touched afterwards** — had the literals been
read off the artifact and written down, those files would carry mtimes after 20:39:28. Independently,
I re-derived the `three-band-page` literals from the ruled formula (`811890 − 4000 − 9621 = 798269`;
`− 18000 − 12828 = 781062`; `− 138000 − 12828 = 661062`; `− 751890 − 8552 = 51448`) and the `font-text`
and `multi-script` ones, and all reproduce.
**Caveat, not a defect in the work:** mtimes are the only ordering evidence available, because the
whole change set sits in **one uncommitted working tree**. DN-4 recommended the `splitByFace` deletion
be *"taken as its own commit inside this story with its own line in the delivery log"*; there are no
commits at all, so the intra-story attributability the story designed for itself is not yet realised.
That is the finisher's to land, and is noted rather than charged.

**5 — D-000.47: VERIFIED, completeness half red-proved on a constructed site.**
I created `fixtures/zz_reviewer_probe.txt` containing `wrapped-text`'s digest (presence of the digest
in the file asserted before running). `TestGoldenDigestAgreesAtEveryDeclaredSite` failed with
*"fixtures/zz_reviewer_probe.txt records fixtures/wrapped-text's digest … but goldenDigestRecord does
not declare it as a site"*. File removed; tree confirmed clean. The check is genuinely a set-equality
over a filesystem walk in both directions, not a hand-maintained literal, and assertion (1) re-hashes
the artifact — teeth the superseded shape did not have. Discipline now covers **7** fixtures / **16**
declared sites; `three-band-page` and `wrapped-text` genuinely had no second literal before.

**6 — D-000.34: VERIFIED, and the sweep extended.**
`TestLineAdvanceIsTheMaxOverTheDeclaredChain` is rebuilt with two named rival hypotheses (`firstFace`,
`superseded`), the negative control moved to the two single-face rows where all three agree *by
construction*, and **three** vacuity guards (one per role). Reports 3 / 3 / 2, matching the story.
**I checked the rest of the touched set for the same silent-control-loss and found none.** In
particular I verified — by hand, not by inheritance — the matrix-only `requireThreeBandPageUsesAllThreeBands`,
which is compile-checked but **never executed**: its thresholds are `y >= 793890` (header) and
`y <= 66000` (footer); the new baselines classify 798269 → header (margin 4379 mp), 781062 and 661062
→ content, 51448 → footer. The control survives with room. `TestThreeBandPageSemanticAcceptance`'s
"four distinct placements" survives on three distinct offsets (9621 / 12828 / 8552).

**7 — D-000.39 sharpened: VERIFIED, the pair is properly formed.**
`hhea lineGap` is 0 on all four committed faces. **Mutation B** (drop the `max(lineGap)` term) left
**every golden green** — 501 PASS, the only non-P6 failure being the fabricated non-zero-`lineGap`
subtest. **Mutation A** is a neighbouring **observable** mutation in the **same expression and site**
and moves the artifacts. Both are reported together and neither is claimed alone. **No golden
assertion was strengthened to chase `lineGap`** — the teeth come solely from
`TestVerticalModelArithmeticOverFabricatedMetrics` over fabricated `fontset.LineMetrics`, which the
`verticalModel(chain, []fontset.LineMetrics, fontSize)` signature exists to make reachable.

### The "verify, do not treat as findings" set

- **`splitByFace` deletion — the right call, and correctly executed.** Verified at `17f5f7a`: the
  declaration at `render.go:463` had **zero** invocations in production or test (the only non-comment
  occurrence was the `func` keyword itself, so no caller was silently removed), and its body was a
  pure pass-through — `shapeSegments` forwarded verbatim, then `positionSegments` over the full rune
  range, **no arithmetic, no parameter not forwarded, no ordering of its own**. The live per-line path
  reproduces it as `positionSegments(segs, ln.from, ln.to, …)`. All nine references were re-pointed and
  **none now attributes to `positionSegments`/`shapeSegments` anything they do not do**. The only prose
  deleted rather than moved (the 640/320 mp divergence measurement) survives at
  `internal/fontset/fontset.go:570-576`. Option (i) over (ii) is correct: (ii) would have left dead
  code that must now track the amended model.
- **Wrapping/overflow — structural claim VERIFIED.** `packLines(segs, ops, totalRunes, fontSize, boxWidth)`
  (`render.go:345`) takes **no vertical quantity** and is called **before** `chainVerticalModel`
  (`render.go:360`). The advance is not an input to line breaking and cannot become one without a
  signature change. Corroborated: `layout_probe_test.go` and `TestWrappedTextLayoutProperties` stayed
  green under every mutation, and the acceptance table fails on a differing *count* of baselines
  before it ever compares a position.
- **Thai BREAK sign-off unaffected — MEASURED, not assumed.** `fixtures/expected-breaks/` and
  `fixtures/thai-break-corpus/` are **byte-identical** to `17f5f7a` (all six files hash-compared
  individually, not just `git diff --stat`). No break-related source or test file appears in the change
  set. `thai-signoff.json` does **not** exist anywhere in the repo, and the release condition is
  written into `epic-2-boundary-gate.md` with **both** halves plus the Story 2.6 hand-off.
- **`present == 0` / `maxUnits <= 0` — the judgment is correct, and better than the alternative.**
  Declining the D-000.24 "forward guard" label is right: a forward guard is one with *no available
  red-proof*, and both paths **are** red-proved one level down over fabricated metrics precisely
  because the `verticalModel` seam is reachable without a `*fontset.Font`.
  `TestVerticalModelErrorPathsAreUnreachableThroughRender` asserts **which** error returns
  (`"has a glyph for rune"`, and explicitly *not* `"no line height can be derived from it"`), which is
  the measurement; asserting merely that an error came back would pass either way. Both halves are
  stated. I concur.
- **Heavy tests declined; legs compile-checked only.** Confirmed: `go vet -tags matrix ./...` clean in
  both modules, and the four `//go:build matrix` files are unchanged. Stated plainly in the log.
- **The gate owes exactly four things — no fifth.** `declaredEpic2GateObligations` is **byte-identical**
  to baseline; `TestEpic2GateObligationsMatchTheDeclaredSet` passes with a **7-of-7** witness and is
  **not circular** — the observed set is built from a filesystem walk (140 Go files) plus a Go-source
  parse of `matrix_test.go`, never from the literal. `matrixDocuments` is untouched at 7 entries and
  reads each digest from the fixture file at runtime.
- **AC2's mandated witness, stated as required:** `internal/pdf/textdoc.go` contributes **exactly 1**
  direct `flipY(...)` call — **unchanged from baseline's 1** (`:689` → `:706`), with only the fourth
  argument changed from `run.FontSize` to `run.BaselineOffset`. `internal/pdf/flip.go` is
  **byte-identical** to `17f5f7a`. `TestContentStreamYCoordinatesRouteThroughFlipY` passes and its
  per-file vacuity guard (both `textdoc.go` and `imagedoc.go` non-zero) is satisfied.

---

### Finding 1: `render.go` cites a test that does not exist

- **Severity**: Minor
- **Category**: Convention (D-000.37 — a stale pointer a reader will act on)
- **Location**: `folio-go/render.go:359`
- **Observation**: The comment introduced by this story reads *"That widening is measured rather than
  assumed: see TestVerticalModelErrorPathsThroughThePublicEntryPoint."* A repo-wide search finds
  **exactly one** occurrence of that name — this comment. The test is actually
  `TestVerticalModelErrorPathsAreUnreachableThroughRender` (`folio-go/vertical_model_test.go:404`).
- **Impact**: A reader following the module's own pointer to the AC4 evidence finds nothing, and a
  grep-based audit of "is this widening covered?" returns empty. AC8 is precisely the comment-
  correctness sweep, and this comment was *added* by the sweep rather than missed by it.
- **Suggested Resolution**: Rename the reference to `TestVerticalModelErrorPathsAreUnreachableThroughRender`.
- **Related AC**: AC4, AC8

### Finding 2: The D-000.45 acceptance test's success log states both directions backwards

- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/first_baseline_acceptance_test.go:386` (counters at `:365-372`, guard at `:383`)
- **Observation**: `belowEm` counts `wantFirstBaselineMP < fontSizeMP` — a **smaller** offset from the
  element top, i.e. a baseline sitting **higher** on the page; that is `font-text`'s 2 elements. The
  `Fatalf` at `:383` gets this right (`aboveEm` → *"further down the page"*, `belowEm` → *"further UP"*).
  The `Logf` at `:386` passes `belowEm, aboveEm` into *"%d elements sit lower than the point size
  implies, %d sit higher (font-text, Roboto max(A)=928 < 1000)"* and so emits, verbatim on the passing
  run: *"2 elements sit lower … 16 sit higher (font-text, Roboto max(A)=928 < 1000)"*. Both halves are
  inverted, and the parenthetical attributes `font-text` to the 16 by adjacency when it is the 2. The
  story's own Delivery Log states it correctly (*"16 lower, 2 higher — the 2 are font-text's"*), so the
  code and the prose disagree.
- **Impact**: No assertion is weakened — the guard is symmetric and I red-proved it fires. But this is
  the one diagnostic a future reader consults to understand *which* fixture is the sign discriminator,
  inside the test whose entire purpose is D-000.45's "never assert a direction". A reader trusting it
  would conclude Roboto moves the baseline down.
- **Suggested Resolution**: Swap the two arguments in the `Logf`, or reword to match the `Fatalf`'s
  phrasing, and attach the `font-text` parenthetical to the correct count.
- **Related AC**: AC6, AC15

### Finding 3: The Delivery Log misreports the wrapped-text interval count (8 vs the 5 the test prints)

- **Severity**: Minor
- **Category**: AC Conformance (reported measurement does not reproduce)
- **Location**: story file *AC6 / D-000.44* section; code at `folio-go/wrapped_text_fixture_test.go:138-159,194`
- **Observation**: The log states the repaired step reports *"a reported count of intervals actually
  compared (8)"*. Running it, the test prints: *"9 text runs across 9 baselines …; **5** OBSERVED
  inter-baseline intervals all equal the ruled 17710 mp"*. **The code is correct and the number in the
  story is wrong**: the fixture's elements carry 3/2/2/2 lines, giving 2+1+1+1 = **5** *intra-element*
  intervals. 8 would be all consecutive pairs across the flat 9-baseline list, which would fold in
  three **cross-element** gaps that are not the ruled advance at all — i.e. the story's number
  describes an assertion that would be wrong to make.
- **Impact**: A reported coverage figure that does not reproduce, in the same section that repairs a
  test for reporting a figure that meant nothing. Low risk, but D-000.26 asks every number to state a
  counting rule that survives re-measurement.
- **Suggested Resolution**: Correct the figure to 5 and state the counting rule (intra-element
  intervals only, which is what makes the assertion meaningful).
- **Related AC**: AC6, AC15

### Finding 4: The Delivery Log's wrapped-text "after" digest is a head/tail splice of two different values

- **Severity**: Minor
- **Category**: Documentation / measurement integrity (AD-21)
- **Location**: story file, *AC5* table row `wrapped-text`; cross-check `fixtures/wrapped-text/expected.pdf`
- **Observation**: The AC5 table records the after-digest as **`277bc5c0…89c3c22a`**. The committed
  artifact hashes to **`277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5`** — head
  `277bc5c0`, tail **`056b0ad5`**. The recorded tail `89c3c22a` is the tail of the *probe* digest
  `3d41f462…89c3c22a` from *Measured findings* 2. Two distinct sha256 values sharing a last-8 is ~1 in
  4·10⁹, so this is a splice, not a coincidence. `goldenDigestRecord` and `expected.json` both carry the
  correct value, so **no code is affected**.
- **Impact**: This is the single row carrying the story's load-bearing claim that *"the fifth did not
  land on the probe value, because it is the only fixture the DN-3 half also touches."* As transcribed,
  the recorded before/after pair shares a tail with the probe and reads as if it partially matched. The
  conclusion is nonetheless **correct** — I verified the real digest differs from the probe entirely.
- **Suggested Resolution**: Correct the tail to `…056b0ad5` in the AC5 table.
- **Related AC**: AC5

### Finding 5: `epic-2-boundary-gate.md` records a 65-character digest for `three-band-page`

- **Severity**: Minor
- **Category**: Correctness (human-facing instruction; no guard covers it)
- **Location**: `_bmad-output/implementation-artifacts/epic-2-boundary-gate.md`, Story 2.5a section,
  `three-band-page` row
- **Observation**: The appended text states the new digest as `746efcbcfb5be30a06caaaefae25e3eaba1962c3fa47a74da10af6d0885372bff`
  — **65 hex characters**, one trailing `f` too many. The artifact's actual digest is the 64-character
  `…885372bf`. The other digests quoted in the same section (`5964aad0…c92e00f`, `6c040ef7…c6c85370`)
  are correct.
- **Impact**: This is the digest a human is told the **deferred** `three-band-page` matrix legs will
  compare at the Epic 2 boundary — the one row of that table whose legs were never run in-story. The
  new `goldenDigestRecord` completeness check cannot catch it: `goldenDigestSearchScope` is
  `{"fixtures", "folio-go/byte_neutrality_test.go"}`, and `_bmad-output/` is deliberately excluded — and
  even in scope, a `strings.Contains` scan would match the correct 64 chars *inside* the 65-char string
  and report a declared site rather than a defect.
- **Suggested Resolution**: Drop the trailing `f`. Optionally, since the exclusion of `_bmad-output/` is
  correct and should stay, consider whether digests quoted in the **gate** file (as forward-looking
  instructions rather than past-tense measurements) warrant a length/shape check.
- **Related AC**: AC12, AC13

### Finding 6: "Zero occurrences of `splitByFace` remain anywhere in the repository" is false as written

- **Severity**: Nit
- **Category**: Documentation
- **Location**: story file, *DN-4 / D-000.46* section
- **Observation**: A full working-tree walk finds **20** occurrences across 6 files, all in
  `_bmad-output/` markdown (including the claim sentence itself, and the historical story files for
  2.2/2.3/2.4). Zero remain in **`folio-go/`** (11 → 0), which is what was actually measured and what
  matters.
- **Impact**: Cosmetic. The surviving occurrences are past-tense narrative and *should* survive, by the
  same reasoning `goldenDigestSearchScope` uses to exclude `_bmad-output/`.
- **Suggested Resolution**: Scope the sentence to `folio-go/`.
- **Related AC**: AC8

### Finding 7: Two internal count inconsistencies in the story's own prose

- **Severity**: Nit
- **Category**: Documentation
- **Location**: story file, correction **C2**; and *AC6 / D-000.44*
- **Observation**: (a) C2 enumerates the `splitByFace` references as *"`render.go:71`, `:73`, `:425`,
  `:477`, `internal/fontset/fontset.go:576`, `wrap.go:88`, plus `segment_origin_test.go:42`"* — that list
  omits `render.go:593`, `render.go:597` and `wrap.go:111`, and labels six what its own DN-4 later counts
  correctly as *"six production comments plus `wrap.go`'s two and `segment_origin_test.go`'s one"* = nine.
  DN-4's count is the right one and matches the code. (b) *"All **18 baselines** across the five fixtures
  matched their hand-derived literals"* — the table declares **18 elements** carrying **23 baselines**
  (`wrapped-text` alone contributes 9 from 4 elements). The 18 is the element count.
- **Impact**: Cosmetic; neither figure is load-bearing and both are corrected elsewhere in the same file.
- **Suggested Resolution**: Align C2's enumeration with DN-4's, and say "18 elements / 23 baselines".
- **Related AC**: AC6, AC8

### Observation (not a finding): the reading sign-off's test runs in zero automated gates

`TestShapedTextThaiSemanticSignOffIsRecorded` correctly `t.Fatalf`s when `thai-signoff.json` is absent —
but the file carries `//go:build matrix`, so `go test ./...` reports *"no tests to run"* for it, and CI
only **compiles** it (`go build -tags=matrix`, `go vet -tags=matrix`); `matrix.yml` runs only
`-run TestTargetRenderHash` / `TestTargetProbeHex`. This is **pre-existing** and by design — the story's
phrase *"the gate stays unable to pass"* is accurate about the deliberate `-tags=matrix` boundary run,
and this story neither created nor weakened the arrangement. Recorded only so that "outstanding" is not
later read as "an automated gate is currently red on it". The anti-rot binding is real and now bites:
`assertSignOffMatchesFrozenHash` compares the record against `expected.json`'s digest, which this story
moved, so any sign-off written today must name the new bytes.

### Self-verification

Every AC was explicitly considered. AC1 satisfied (one `chainLineMetrics` walk feeding one
`verticalModel`; `lineAdvance` is a projection, not a second derivation). AC2 satisfied, witness stated.
AC3 satisfied. AC4 satisfied. AC5 satisfied (five moved, `minimal-rect`/`image-embed` re-hashed
byte-unchanged, all 7 digests re-derived independently). AC6 satisfied, with Findings 2 and 3. AC7
satisfied and red-proved. AC8 satisfied, with Finding 1. AC9 satisfied. AC10 satisfied, reproduced.
AC11 satisfied — and applied by this review to its own five mutations. AC12 satisfied, with Finding 5.
AC13 satisfied. AC14 satisfied and extended. AC15 satisfied. No production code, test, fixture or
document other than this story file was modified by the review; all five mutations and one probe file
were reverted by hand and the final tree matches the 30 entries handed over.

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-24
- **Story Status Recommendation:** **Approved with Minor Notes**
- **Blockers:** 0
- **Majors:** 0
- **Minors:** 5
- **Nits:** 2

**Rationale.** Every load-bearing claim was verified by construction rather than read: both gate tables
reproduce exactly (503 · 1 / 311 · 1 and 85 · 0 / 47 · 0, single permitted failure byte-identical); the
narrowing algebra is sound and its blast radius reproduces under mutation; the D-000.39 mutation pair is
properly formed at one site with the inert half never reported alone; the D-000.45 direction guard, the
D-000.47 completeness half and the D-000.34 control rebuild were each **red-proved by a constructed
mutation** and each fired for the right reason. The `splitByFace` deletion is the correct option and was
executed without losing behaviour or leaving an inaccurate re-pointed comment. The AC6 ordering claim —
the one that separates prediction from transcription — is corroborated by independent filesystem
evidence, not merely asserted.

The five Minor findings are all defects of *record* rather than of behaviour: a comment citing a
non-existent test (F1), a success log that states its two directions backwards (F2), two reported
measurements that do not reproduce (F3, F4), and a mistyped 65-character digest in a human-facing gate
instruction that no guard can catch (F5). None weakens an assertion; F1, F2 and F5 are the ones worth
fixing before commit, because each is a text a human will execute. The one structural note for the
finisher is that DN-4's "its own commit inside this story" has not happened — the whole change set is a
single uncommitted tree, which is also why mtimes were the only available ordering evidence for AC6.

---

## Finding Resolutions (finisher)

**Triage:** 7 reviewer findings — **7 FIX, 0 DISMISS, 0 DEFER** — plus **2 defects found by the
finisher's own sweeps**, both FIXed, listed as F1b and F8. Nothing the review endorsed was reopened.

| # | Sev | Decision | Rationale (and what changed) |
|---|---|---|---|
| **F1** | Minor | **FIX** | `render.go`'s AC4 pointer named `TestVerticalModelErrorPathsThroughThePublicEntryPoint`, which resolves nowhere. Renamed to `TestVerticalModelErrorPathsAreUnreachableThroughRender`, the real declaration. Not fixed as a single line — see *the citation sweep* below, which is the D-000.48 obligation this finding created. |
| **F1b** | Minor | **FIX** *(finisher-found, by the sweep F1 mandated)* | The same AC8 sweep also wrote `present == 0` and `maxUnits <= 0` into `vertical_model_test.go`'s doc comment and its success log as if they named symbols. Neither exists: `verticalModel`'s error paths are `len(metrics) == 0` and `units <= 0`. Both now name the production expression **and** keep AC4's shorthand in quotes, so the AC tie is not lost. Exactly F1's class, missed because the reviewer had already concurred with the surrounding *judgment*. |
| **F2** | Minor | **FIX** | The D-000.45 acceptance log passed `belowEm, aboveEm` into *"%d sit lower … %d sit higher"* and attached the `font-text` parenthetical to the wrong count, contradicting its own `Fatalf`. Fixed **structurally**, not by swapping two arguments: the per-fixture attribution is now **collected in the same walk that increments the counters** (`aboveEmBy` / `belowEmBy`, rendered by `contributorsOf`), so no fixture name survives as a literal and narration and guard cannot disagree. Red-proved — see below. |
| **F3** | Minor | **FIX** | The AC6 log claimed *"a reported count of intervals actually compared (8)"*; the test prints **5**. Corrected to 5 **with its counting rule** (intra-element intervals only: 3/2/2/2 lines → 2+1+1+1), and the reviewer's point that 8 would describe an assertion that would be wrong to make is recorded. Recomputed from a cited invocation, not transcribed (D-000.18). |
| **F4** | Minor | **FIX** | The AC5 table's `wrapped-text` after-digest `277bc5c0…89c3c22a` spliced the real head onto the *probe*'s tail. Corrected to `…056b0ad5` from a cited `shasum -a 256` run, and the full 64 characters are now written out once beside the probe value so the two cannot be confused again. The claim the row supports was already correct. |
| **F5** | Minor | **FIX** | `epic-2-boundary-gate.md` recorded a **65-character** digest for `three-band-page`. Trailing `f` dropped; the value now matches the artifact exactly. And, per the finding's own suggestion, a **mechanical check was added**: `TestBoundaryGateDigestsAreWellFormed` asserts every delimited hex run of ≥32 characters in a boundary-gate document is exactly 64. `_bmad-output/` stays out of `goldenDigestSearchScope` — this is a **shape** check, not a membership one, so the exclusion the reviewer endorsed is untouched. |
| **F6** | Nit | **FIX** | *"Zero occurrences of `splitByFace` remain anywhere in the repository"* was false. Scoped to **`folio-go/`** (11 → 0, measured by walking the module), with the surviving `_bmad-output/` occurrences counted, named, and justified as past-tense narrative that must survive. |
| **F7** | Nit | **FIX**, and **the reviewer's own resolution corrected** | (a) C2's enumeration was incomplete, as flagged. But the finding told the finisher to align C2 with DN-4's *"nine"* — **DN-4 is wrong too**: `git grep -c splitByFace 17f5f7a -- folio-go` gives 11 lines across four files; one is the declaration, so there are **ten** references. DN-4's "six production comments" drops `internal/fontset/fontset.go:576`. Both sites now say ten. (b) *"18 baselines"* → **18 elements carrying 23 baselines**, counted from the table itself. |
| **F8** | Minor | **FIX** *(finisher-found)* | The Delivery Log's gate table recorded *"compressor imports: 1, `internal/template/image.go`'s pre-existing `compress/zlib`"*. That file imports **only `fmt`**; its sole mention of `compress/zlib` is a comment saying PNG bytes are passed through **without** it (AC10 / D-1.8.1). The true count under `folio-go/` outside `testdata/` is **0**. Left standing, the row would have read as carried risk R4 being open when the module's whole design is that it is closed. |

**Nothing was DISMISSed or DEFERred.** Every finding had a concrete, in-scope, cheaply-verifiable
resolution and none required widening the story. The reviewer's endorsed set — the `splitByFace`
option, the `packLines` structural claim, the declined D-000.24 label, the D-000.39 mutation pair,
the four gate obligations — was verified as still true after the edits and **not** reopened.

### The citation sweep F1 demanded (D-000.48, D-000.35, D-2.2.4's *enumerate, never sample*)

Mechanical, not by reading. Population: **1 799 added lines across 12 Go files** (the tracked diff
against `17f5f7a` plus the two wholly-new test files, every line of which counts as added), and
**2 276 added lines across 8 documents**. Resolution was checked against **368 real `Test…`
declarations** parsed out of every `.go` file in the repository.

| swept population | citations found | resolve | do not resolve |
|---|---|---|---|
| `Test…` names in added **Go** lines | **11** | 10 | **1** — F1 |
| `Test…` names in added **document** lines | **26** | 24 | **2** |
| camelCase/PascalCase **symbols** in added Go comments | **77** | 75 | **2** |
| **total distinct citations swept** | **114** | **109** | **5** |

The five that do not resolve, each dispositioned:

1. `TestVerticalModelErrorPathsThroughThePublicEntryPoint` in `render.go` — **F1, fixed.**
2. The same name in this story file — the QA finding **quoting** the defect. Correct as written.
3. The same name in `folio-mvp-decision-log.md` (D-000.48) — the ruling **naming the instance**. The
   log is append-only; quoting a wrong name in order to condemn it is the entry's purpose.
4. `TestStory23aMovedNoGoldenDigest` in this story file — **intentional**. Every occurrence is
   past-tense (*"was replaced by `TestGoldenDigestAgreesAtEveryDeclaredSite`"*, *"0 occurrences —
   cleanly replaced"*) or a pre-fix measurement taken at `17f5f7a` when it existed. A citation of a
   deliberately deleted symbol is a record, not a pointer — the same reasoning as F6.
5. `maxUnits` (and `present`) in `vertical_model_test.go` — **F1b, fixed.**

Of the 77 comment symbols, 16 were flagged by the mixed-case heuristic and hand-adjudicated: 14 are
genuine and resolve — OpenType/PDF **format** names (`FontDescriptor`, `fsSelection`,
`sTypoAscender`, `lineGap`), template **JSON keys** (`pageHeader`, `pageFooter`, `footerHeight`),
`flipY`'s **parameters** (`pageHeight`, `marginTop`, `pdfY`), `PageGeometry.MarginBottom`,
`PlaceInBand`'s `bandOrigin` parameter, `firstFace` (a struct field in `wrap_test.go`), `boundText`
(a local in `render.go`), and `printableHeight` (defined inline on the next comment line). The
remaining two are F1b.

### The two non-reproducing measurements, recomputed from cited invocations

| claim | recorded | **measured at finish** | invocation |
|---|---|---|---|
| `wrapped-text` intervals compared | 8 | **5** | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v -run TestWrappedTextSemanticAcceptance ./...` → `wrapped_text_fixture_test.go:194`: *"9 text runs across 9 baselines …; **5** OBSERVED inter-baseline intervals all equal the ruled 17710 mp"* |
| `wrapped-text` after-digest | `277bc5c0…89c3c22a` | **`277bc5c023475b77fbcaebf0421c982e1456ccec292b4c92d88efa89056b0ad5`** | `shasum -a 256 fixtures/*/expected.pdf` from the repo root |

Both were **recomputed, never transcribed** (D-000.18). Neither changes a conclusion: the 5 is what
makes the interval assertion meaningful, and the real `wrapped-text` digest differs from the probe
`3d41f462…89c3c22a` in every position that matters, which is what the AC5 row claims.

### Mutations applied at finish — each asserting it applied before any test result was read (AC11)

**M1 — the new gate-digest shape check, red-proved on the very defect it was written for.**
Re-inserted the trailing `f` into `epic-2-boundary-gate.md`. Precondition asserted first (exactly
one occurrence of the correct 64-char value); artifact change observed
(`sha256 43a90746…` → `d917a5df…`, 15 190 → 15 191 bytes) **before** the suite was run.
`TestBoundaryGateDigestsAreWellFormed` became the **sole** failure:

> `_bmad-output/implementation-artifacts/epic-2-boundary-gate.md records "746efcbc…885372bff" as a
> digest: that is 65 hex characters, and a sha256 is 64.`

and its own witness line dropped from 3 well-formed digests to 2 — so the check saw the change, it
did not merely fail. Reverted by hand from a byte-exact backup (`cp`, never `git checkout`),
confirmed with `/usr/bin/diff` (identical) and by re-hashing to `43a90746…`; backup deleted and its
absence verified.

**M2 — the repaired narration, proved to be derived rather than restated.**
A `Logf` cannot be red-proved by failing, so the property asserted is *"the sentence tracks the
table"*. Mutated `font-text/e1`'s `maxAscent1000` 928 → 1160 (precondition: exactly one occurrence;
artifact change observed, `sha256 1860d74a…` → `5b7a9104…`). The message followed the data with no
edit of its own — **17 LOWER `[font-text:1 multi-script-fallback:1 shaped-text:7 three-band-page:4
wrapped-text:4]` and 1 HIGHER `[font-text:1]`** — while the artifact assertions correctly reddened.
The superseded line would still have printed its hard-coded *"(font-text, Roboto max(A)=928 < 1000)"*
beside the wrong number. Reverted by hand, `/usr/bin/diff` clean, backup deleted.

The pre-fix and post-fix narrations, from live runs of the same tree:

- before — *"**2** elements sit lower than the point size implies, **16** sit higher (font-text …)"*
- after — *"**16** elements sit LOWER on the page than the point size implies (contributed by
  [multi-script-fallback:1 shaped-text:7 three-band-page:4 wrapped-text:4]) and **2** sit HIGHER
  (contributed by [font-text:2] …)"*

## Delivery Log — finisher

### Gates, re-measured on the finished tree

Raw stream, `rtk proxy` into a shell script with the environment set **inside** the script, counted
with Python — never a shell pipe, for the reason the developer's own log gives (`grep` is rewritten
to `rg` here and `grep -c -- '--- PASS'` returns *"unrecognized flag"* → 0). Every count below states
its file, its byte size and its line count before it is believed.

| scope | counting rule | baseline `17f5f7a` | handed to finisher | **at commit** | Δ vs handover |
|---|---|---|---|---|---|
| `folio-go/` | all occurrences | 487 · 1 | 503 · 1 | **504 · 1** | **+1 PASS**, failures unchanged |
| `folio-go/` | top-level (`^--- PASS`) | 305 · 1 | 311 · 1 | **312 · 1** | **+1 PASS**, failures unchanged |
| `lint/` | all occurrences | 85 · 0 | 85 · 0 | **85 · 0** | unchanged |
| `lint/` | top-level | 47 · 0 | 47 · 0 | **47 · 0** | unchanged |

**The handover figures reproduced exactly before any edit** (503 · 1 / 311 · 1 / 85 · 0 / 47 · 0),
so the +1 is attributable. It is **not** asserted from the totals: the top-level test-name sets of
the two runs were differenced, and the difference is **added: `TestBoundaryGateDigestsAreWellFormed`;
removed: none**. Exactly the one check F5 asked for, and nothing else moved.

**The single permitted failure is `TestCorpusMeetsP6ExerciseFloors`, byte-identical to baseline** —
compared line for line, not eyeballed:

```
corpus_test.go:189: P6g (opaque names) floor not met: got 7, need >=20
corpus_test.go:192: P6 stats: {P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}
```

**There is no second failure.**

| other gate | result at commit |
|---|---|
| `gofmt -l` — both modules | clean (absolute paths for both arguments) |
| `go vet ./...` — both modules | clean |
| `go vet -tags matrix ./...` — both modules | clean |
| `GOWORK=off go list -m all` | **exactly two**: `github.com/panitw/folio/folio-go`, `github.com/boxesandglue/textshape v0.0.15` |
| `SOURCE_DATE_EPOCH` literal in any `.go` under `folio-go/` outside `testdata/` | **0**, comments and error strings included |
| compressor imports under `folio-go/` outside `testdata/` | **0** — see F8; the only occurrence in the module is the deliberate lint fixture at `testdata/lint/no-compressor/violating-compressor/bad.go` |
| `fmt` imported in `internal/pdf` | **0** |
| float guards | both green (`TestNoFloat64UnderInternal`, `TestNoFloat64UnderModule`, `no-float-typed-value`) |

**Heavy tests: the D-000.4 override stays DECLINED, and the matrix legs were COMPILE-CHECKED BUT NOT
RUN.** Stated plainly because the distinction is the whole point: `go vet -tags matrix ./...` is
clean in both modules, so every `//go:build matrix` leg still **compiles** against the amended model
and against the re-recorded digests — but **no matrix leg was executed** at finish, by the
once-per-epic cadence. In particular `three-band-page`'s deferred legs and
`TestShapedTextThaiSemanticSignOffIsRecorded` did not run here and are the Epic 2 gate's to run.

### Constraints re-verified at finish, by measurement

- **Exactly five goldens moved, and no sixth.** `git diff --name-only 17f5f7a -- fixtures` lists
  **five** `expected.pdf` (`font-text`, `multi-script-fallback`, `shaped-text`, `three-band-page`,
  `wrapped-text`) with their five `expected.json` and two `README.md`. `minimal-rect/expected.pdf`
  and `image-embed/expected.pdf` **do not appear**, and were independently re-hashed to
  `0f925e1b…6cb4f7c` and `e5778eb8…abe689fc`. `fixtures/shaped-text/expected.pdf` is
  `6c040ef7a82a3604912fb3793324da72dcf421527db753ae59e5813ac6c85370`, as intended. **Nothing was
  reverted.**
- **The Thai BREAK sign-off is unaffected — verified, not assumed.**
  `git diff --name-only 17f5f7a -- fixtures/expected-breaks fixtures/thai-break-corpus` lists
  **zero** files. The break-opportunity vector is byte-identical to baseline.
- **`fixtures/shaped-text/thai-signoff.json` was NOT created.** A whole-repository `find` for that
  name returns **0** results. D-000.41 names this story by key; the release condition lives in
  `epic-2-boundary-gate.md` and only there.
- **Every digest site is updated, and D-000.47's declared list is complete.**
  `TestGoldenDigestAgreesAtEveryDeclaredSite` passes with **7 artifacts re-hashed and 16 declared
  recording sites agreeing**, and its completeness half reports **no undeclared occurrence** in the
  search scope. The finisher's only digest edit was in `epic-2-boundary-gate.md`, which is outside
  that scope by design — which is precisely why it needed `TestBoundaryGateDigestsAreWellFormed`.
- **`sprint-status.yaml`: `epic-2: backlog` is FLAGGED, NOT CHANGED.** Flipping the epic is the
  boundary gate's act, not this story's. Only `2-5a-…: review → done` was touched.

### AC15 — what the finisher added, categorised honestly

`TestBoundaryGateDigestsAreWellFormed` is **counted as coverage**, not labelled redundant: it has
independent teeth, demonstrated by M1 above, and **nothing else in the repository looks at
`_bmad-output/` at all**. Its vacuity guard fails if no boundary-gate document carries a digest, so
it cannot pass by looking nowhere; its threshold is derived from a measurement (647 delimited hex
runs across both gate documents, longest sub-threshold run **8** characters) rather than picked.

The `contributorsOf` narration helper adds **no** assertion and is **not** counted as coverage — it
makes an existing report derive from an existing walk. Recorded so the check count is not inflated
by a reporting fix.

### DN-4 — the deletion is now its own commit, as DN-4 asked

DN-4 required the `splitByFace` deletion to be *"taken as its own commit inside this story with its
own line in the delivery log"*. The review found there were **no commits at all**, which is also why
filesystem mtimes were the only ordering evidence available for AC6. This story therefore lands as
**two commits**, the deletion **first** and separately revertible:

1. **`splitByFace` deleted** — the declaration removed from `render.go`, and all **ten** references
   re-pointed at `shapeSegments` / `positionSegments`, the functions that actually execute. Four
   files, no behaviour: the function had zero call sites and a pure pass-through body. For
   `internal/fontset/fontset.go` and `segment_origin_test.go` this commit is their **entire** change
   in this story; for `render.go` and `wrap.go` it is a strict subset, built from the `17f5f7a` blobs
   by applying only the re-pointings and the deletion, each asserted to match exactly once.
2. **The vertical model itself**, its tests, the five re-recorded goldens, the documents, and every
   finding resolution above.

**Verified rather than assumed, and the number is the proof.** Commit 1 was checked out into a
detached worktree and measured there on its own, under `CGO_ENABLED=0 GOWORK=off`: `go build ./...`,
`go vet ./...`, `go vet -tags matrix ./...` and `gofmt -l` all clean, **zero files under `folio-go/`
still mentioning `splitByFace`**, and the suite reporting **487 PASS · 1 FAIL all-occurrences,
305 · 1 top-level** — *exactly* the `17f5f7a` baseline this story recorded at creation, with the same
single `TestCorpusMeetsP6ExerciseFloors` failure and byte-identical failure text. A deletion that
leaves the baseline's own counts untouched is the strongest available statement that nothing was
lost, and it settles that commit 1 is not a broken intermediate.
