# Story 2.3: Shape Latin, Thai and CJK text

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-3-shape-latin-thai-and-cjk-text`
**Status:** `done`
**Covers:** NFR3 (`epics.md:740`) · **AD-2, AD-3, AD-7, AD-8**
**Adjacent invariants:** AD-1, AD-14, AD-21, AD-22, AD-23, AD-24, AD-25, AD-26
**Governing rulings:** **D-000.21 (sharpened) · D-000.22 · D-000.23 · D-000.9 (+ probe extension)** ·
**D-2.2.4 · D-2.2.5 · D-2.2.6 (amended) · D-2.2.1 · D-2.1.6 · D-2.0.1 · D-2.0.4** ·
D-000.4 · D-000.5 · D-000.6 · D-000.12 (corrected) · D-000.13 · D-000.15 · D-000.16 · D-000.17 ·
D-1.1.a · D-1.3.5 · D-1.5.1 · D-1.5.6 · D-1.5.7 · D-1.5.8 · D-1.8.1 · D-1.8.6 · D-2.1.5
**Does not retire any DW item.** Adds one (DW-2.3-a, AC13's cross-validation) if the lead defers it.

**Baseline measured in this run, at creation.** HEAD is **`3373dac`** — *"Story 2.2: The shipped font
set and its fallback chain (finisher)"* — on branch `main`, **working tree clean**
(`git status --porcelain` empty, verified before and after every measurement below, per D-000.12
corrected). Test state at baseline: **331 pass, 2 fail in 12 packages**. The two failures are
`internal/text`'s `TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck` — **Story
2.1's deliberate, pre-stated shortfalls (D-2.1.14). They are the baseline, not a regression. Do not
fix them, do not tune the corpus to make them pass, and do not count them when reporting this
story's test results.**

Everything asserted below was measured against `3373dac` with a scratch module outside the repo
(`textshape v0.0.15` from the module cache, reading the three committed `.ttf` files by absolute
path). The scratch module was removed and the clean tree re-verified before this file was written.

---

## In plain terms (read this first if you just want the gist)

Letters are not laid out one at a time. In Thai, a vowel or tone mark is a separate character that
has to be drawn on top of the consonant before it, nudged sideways, and sometimes swapped for a
different, lower-sitting shape when the consonant it lands on is unusually tall. Before this story
the engine looked up each character, drew it, and moved right. For a name with a tone mark on a tall
consonant, the mark collided with the letter it belonged to.

That is now fixed. The typeface's own layout rules decide which shapes to draw, in what order, with
what nudges, and the document is built from that answer. English gains proper spacing between
awkward letter pairs and the joined shapes a good typeface provides. Chinese, checked and stated
plainly, is unchanged.

Review found one live fault the story had missed and it is fixed here: text was being drawn with the
tighter spacing but the *next* stretch of a mixed-language line was still positioned as though the
tightening had not happened, so a Thai word after an English one sat visibly too far right. A new
line was added to the reference document specifically so a picture can catch that in future; it
could not before, by luck of ordering rather than by design.

Three things will look wrong and are not. Two tests inherited from the previous story still fail on
purpose. The reference document's hash was frozen before a Thai reader confirmed the Thai actually
reads correctly, so a check has been deliberately left failing at the end-of-epic gate until someone
who reads Thai looks and records what they saw — and that record is tied to the exact frozen bytes,
so re-recording invalidates it. One spec limit in the text-extraction table, inherited and not
reachable today, is written up as future work rather than fixed.

---

## Story

**As a** template author,
**I want** Thai vowels and tone marks to sit correctly on their base characters,
**So that** my customers' names are not merely present but correct.

---

## Do not re-open — settled rulings this story inherits

Reproduced with their rationale so the developer does not re-litigate them.

1. **The shipped faces are STATIC, Regular-only** (`NotoSans-Regular.ttf`,
   `NotoSansThai-Regular.ttf`, `NotoSansSC-Regular.ttf`), generated ahead of the build by
   `tools/fontgen/instance_faces.py`. No `fvar`, no `gvar` anywhere. **Bold and italic do not
   exist** and are fenced out (D-2.2.1, D-2.2.4). Nothing in this story selects a weight, an
   instance, or an axis. `ot.Shaper` exposes `SetVariations`, `SetVariation`, `SetNamedInstance`,
   `SetSyntheticBold` and `SetSyntheticSlant`; **call none of them.** Four of the five take
   `float32`, which `internal/arch_test.go:54` bans under `internal/` and the module root (AD-23),
   and the fifth would fabricate a weight the project deliberately does not ship.
2. **A caller-supplied face carrying `fvar` is rejected at ingestion** by `fontset.New()` with a
   located diagnostic naming the remedy. There is no axis-pinning code path and there must not be.
3. **`/BaseFont` derives from the shipped face's PostScript name** (D-2.2.6 amended), giving
   `TAG+NotoSansSC-Regular`. This story changes the subset *input*, which changes the subset tag.
   **The tag changing is expected; the derivation must not change.**
4. **No compressor may be imported under `folio-go/`** (`no-compressor-import`, D-1.8.1). Font
   programs and content streams ship uncompressed, deliberately. A `TJ` array is longer than a `Tj`
   string; that is accepted, not a reason to reach for a filter.
5. **`SOURCE_DATE_EPOCH` as a literal string in any `.go` file under `folio-go/` trips
   `absence-source-date-epoch`** (D-2.1.5) — *including inside error-message strings*, which caught
   Story 2.2's finisher. Keep it out of Go sources entirely.
6. **Thai word segmentation is declarative, and it is not this story's.** D-2.1.6 (OWNER DECISION):
   *a template declares which fields are unbreakable; the engine never guesses at Thai word
   boundaries.* That work lands in **Stories 2.4 and 2.5**. **Do not invent a segmentation
   algorithm, a heuristic, or a surname list here.** Shaping and breaking are different questions
   and this story answers only the first.
7. **The `ja` glyph-form limitation is stated, not fixed** (D-2.0.1, Epic 2 boundary gate item 1).
   Noto Sans SC is Pan-CJK; Japanese renders with Simplified-Chinese shapes for a subset of shared
   ideographs. Shaping does not change that and must not be described as fixing it.
8. **`ScanMapRange` bans ranging a map anywhere under `internal/`** (D-1.3.5), with no
   site-specific exception. Every new ordered output in this story must be built by ranging a
   **slice**. `fontset.Subset` already models the pattern correctly — copy its shape, read its
   comment.
9. **`fontset.Subset` must not sort its input** (D-1.5.7, AC8a). Permutation-invariance is proved by
   permuting the input; a defensive sort upstream of that check deletes the check.

---

## Scope fence — what this story is NOT

- **It does not break or measure lines.** UAX #14 for Latin, the embedded trie's Thai breaks,
  CJK per-character breaking, the frozen S4 expected-break fixture and D-2.1.6's unbreakable-field
  format key are all **Story 2.4** (`epics.md:759`), which carries its own D-000.4 matrix override.
  This story decides what glyphs to draw and where to put them *within* a run; it decides nothing
  about where a run may end.
- **It does not compose the page.** Bands, `internal/layout`, AD-24's band-relative origin and
  AD-5's arrow are **Story 2.5** (D-2.0.4). `buildTextContentStream`'s provisional origin comment
  and `TestProvisionalBandOriginIsPinned` stay exactly as they are.
- **It does not change the fallback chain or the missing-glyph diagnostic.** Story 2.2's
  coverage-based, per-rune, chain-ordered face resolution (`Font.HasGlyph`) is the input to this
  story, unchanged. Shaping runs **inside** a face-segment the chain already chose (AC8 below).
- **It does not mint a diagnostic code.** `internal/diag` does not exist; missing-glyph and every
  other stable code is **Story 3.6** (`epics.md:1050`). New failures here take the located-error
  shape `render.go` already uses.
- **It does not add a weight, an instance, or a synthetic bold.** See ruling 1 above.
- **It does not enable discretionary typography.** Only the shaper's default feature set runs
  (`Shape(buf, nil)`). No `smcp`, no `onum`, no `ss01`, no manual `Feature` list.
- **It does not change `fixtures/font-text/`, `fixtures/minimal-rect/` or
  `fixtures/image-embed/`.** Measured below: none of their text is shape-observable, so their
  hashes must not move. If one moves, that is a defect in this story, not an intended re-recording.

---

## Measured findings — read all of these before writing code

Every number below was produced at `3373dac` against the three committed faces. Where a finding
contradicts `epics.md`, the contradiction is called out; `epics.md` was written before Stories 2.1
and 2.2 landed.

### F1 — Shaping is observable, and the current path is provably wrong for Thai

Rune → `cmap` lookup (what `fontset.Subset` and `appendHexCIDString` do today) versus
`ot.Shaper.Shape` on the same shipped face:

| Text | Meaning | Unshaped glyph ids (today) | Shaped glyph ids | What changed |
|---|---|---|---|---|
| `ปั` | tall consonant + MAI HAN AKAT | `[80 45]` | `[80 46]` | **GSUB** picks the *lowered* mark form 46; today's 45 collides with `ป`'s ascender |
| `ฟั` | tall consonant + MAI HAN AKAT | `[16 45]` | `[16 46]` | same substitution, plus **GPOS XOffset +21** |
| `ที่` | consonant + SARA II + MAI EK | `[117 94 42]` | `[117 94 44]` | **GSUB** lowered tone mark 44; **XOffset −3** on both marks |
| `น้ำ` | consonant + MAI THO + SARA AM | `[71 47 90]` | `[71 59 49 86]` | **3 runes → 4 glyphs**: SARA AM decomposes and reorders; mark 47→59; **XOffset −29** |
| `ป้ำ` | tall consonant + MAI THO + SARA AM | `[80 47 90]` | `[80 60 49 86]` | 3 → 4 glyphs, 47→60, **XOffset −204** on the NIKHAHIT |
| `office` | Latin | `[82 73 73 76 70 72]` | `[82 1656 70 72]` | **6 runes → 4 glyphs**, `ffi` ligature 1656 |
| `fi` | Latin | `[73 76]` | `[1654]` | **2 runes → 1 glyph** |
| `AV` | Latin kern pair | `[36 57]` | `[36 57]` | glyphs same, **XAdvance 599 vs 639 hmtx — GPOS kerning, −40 font units** |
| `Wo. To,` | Latin | unchanged | unchanged | **two** advance deltas |
| `กรุงเทพ` | plain Thai | `[…]` | identical | **nothing changes** — see F3 |
| `结算单` | CJK | `[21201 20375 10118]` | identical | **nothing changes**, all advances exactly 1000 |
| `结算单，共３页` | CJK + fullwidth punctuation | identical | identical | **nothing changes** |

**This is the story in one table.** Thai is wrong today in a way a Thai reader sees immediately.
Latin gains ligatures and kerning. **CJK changes not at all.**

### F2 — `YOffset` is zero everywhere in the shipped faces; the observable is `XOffset` and `GlyphID`

`epics.md:751` says *"marks are positioned by `GPOS` mark-to-base positioning."* That is true, but
**the observable consequence in these three faces is a horizontal offset plus a substituted glyph,
never a vertical one.** Across every Thai sample above, `GlyphPos.YOffset == 0` without exception.
The mark glyphs are drawn at the correct height by design; mark-to-base contributes the sideways
nudge and `GSUB` contributes the lowered form.

**Consequence, and it is the D-000.21 trap for this story.** An acceptance criterion phrased as
*"assert some mark has a non-zero `YOffset`"* is **vacuously false** and would block development on
a correct implementation. An acceptance criterion phrased as *"assert some mark has a non-zero
`XOffset`"* is true but is D-000.23's error one level down — it guards **the field that happens to
move in these three faces**, and a fourth face that positions vertically sails past it.

**The remedy is cardinality, per D-000.23's consequent obligation.** The meaning "this glyph run is
correctly shaped" is carried by **five independent fields**: `GlyphID`, `XOffset`, `YOffset`,
`XAdvance`, `Cluster`. Assert **all five, for every glyph, against a declarative frozen table** —
not a hand-picked property of a hand-picked sample. See AC2/AC3.

### F3 — Every existing golden fixture is blind to this story

Measured, on the exact text each fixture renders:

| Fixture | Text rendered | Glyph ids identical? | Any position delta? | Verdict |
|---|---|---|---|---|
| `fixtures/font-text/` | `"Hello"` | yes | none | **shaping INVISIBLE** |
| `fixtures/font-text/` | `"Page footer 0123456789"` | yes | none | **shaping INVISIBLE** |
| `fixtures/multi-script-fallback/` | `"Ada ก 汉"` (Latin seg `"Ada "`) | yes | none | **shaping INVISIBLE** |
| `fixtures/multi-script-fallback/` | Thai segment `"ก"` | yes | none | **shaping INVISIBLE** |
| `fixtures/multi-script-fallback/` | CJK segment `"汉"` | yes | none | **shaping INVISIBLE** |

Coverage witness: **5 of 5** rendered text segments across **3 of 3** text-bearing fixtures
evaluated; 0 shape-observable.

**Two consequences, both binding.**

- **Their hashes must not move.** A correct implementation of this story leaves all three golden
  hashes byte-identical. That is a genuine, checkable assertion — and it is the only thing those
  fixtures can say about this story.
- **They cannot confirm the story landed.** If the shaping seam were written and then never called,
  every existing hash test stays green. This is exactly D-000.9's shape: *the "all clear" and the
  "I could not look" are the same value.* **A new fixture is mandatory** (AC10), and its text must
  be chosen so a wrong answer is a different hash (AC11).

**No acceptance criterion in this story may be satisfied by asserting on CJK text, on `"Hello"`, on
`"Ada ก 汉"`, or on plain Thai such as `กรุงเทพ`.** Each of those shapes to itself; each would pass
identically against an implementation that does no shaping at all.

### F4 — The subsetter already retains the shaped glyphs; only the *mapping* is missing

Measured with the rune-derived input `fontset.Subset` builds today:

| Face | Input glyph ids (rune-derived) | Plan output glyphs | Program bytes | Shaped glyph ids **not** retained |
|---|---|---|---|---|
| Noto Sans, `"office fi Acme"` | 14 | 24 | 50,832 | **none** |
| Noto Sans Thai, `"ปั ที่ ป้ำ น้ำ"` | 14 | 19 | 1,600 | **none** |
| Noto Sans SC, `"结算单"` | 3 | 4 | 1,252 | **none** |

`subset.CreatePlan` performs GSUB closure, so ligature 1656 and Thai's substituted forms 46/44/60/59
are already **in the embedded program**. What does not exist is any way to *address* them:
`Subset.GlyphForRune` is keyed by rune, and no rune maps to a ligature. So today a shaped run cannot
be emitted at all — `appendHexCIDString` would fail, or silently emit the wrong CID.

**This makes the change surgical rather than structural.** The seam
(`input.AddGlyphs(gids...)` → `plan.MapGlyph(oldGID)`) is already glyph-keyed. What changes is what
is fed in and what is handed back.

### F5 — Shaped positions are pure integers; only one vendor accessor is `float32`

`ot.GlyphPos` (`shaper.go:227`) is `XAdvance`, `YAdvance`, `XOffset`, `YOffset`, all `int16`. The
shaper works in font units and does no scaling. **The shaped output is exact integer data and needs
no float anywhere**, which is what makes AD-2/AD-23 satisfiable here.

**One live hazard, and it predates this story.** `ot.Face.HorizontalAdvance` (`ot/metrics.go:403`)
returns **`float32`**. `fontset.AdvanceForRune` already converts it with `int64(adv)` — a
float→int truncation on the render path that `arch_test.go`'s guard **cannot see**, because the
guard flags the *identifier* `float32`/`float64` and a bare `int64(x)` names neither. It is safe
today (every `hmtx` advance is exactly representable in `float32`), but it is a blind spot, and this
story would multiply it if advances were taken from that accessor.

**Therefore: take advances from `GlyphPos.XAdvance` (`int16`), never from `Face.HorizontalAdvance`,
on the shaping path.** This is strictly better than avoiding the guard — the shaped advance also
*includes GPOS kerning*, which `HorizontalAdvance` does not. See AC12 for the guard extension.

### F6 — Shaping is deterministic across processes, and builds for every target

- Five separate process invocations, same inputs: output SHA-256
  `b80a4e3bbd7a0f826a1bbacdb55c4b335a8a8314da85694a8c5d0c3c8d798188` **five times out of five.**
- `GOOS=js GOARCH=wasm CGO_ENABLED=0 go build` — **succeeds.**
- `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build` — **succeeds.**
- `textshape`'s own `go.mod` declares **zero requirements**; `folio-go/go.sum` contains exactly one
  module. No cgo is reachable.

`epics.md:747` (*"no cgo is used and the build succeeds for `js/wasm`"*) is **already true at
baseline** and stays true. Note what that means: **an AC asserting only "wasm builds" is vacuous
for this story** — it passes at `3373dac` with no shaping code written. It is kept in AC9 as a
*regression* guard and is explicitly labelled as such, never as evidence the story landed.

### F7 — Cluster values are byte offsets, and merged Thai clusters are the `/ToUnicode` trap

`GlyphInfo.Cluster` is the byte index into the input string. Measured:

- `"office"` → clusters `0, 1, 4, 5`. Ligature 1656 sits at cluster 1 and spans bytes `[1,4)` =
  `"ffi"`. Clean many-to-one.
- `"ณัฐวุฒิ"` → clusters `0, 0, 2, 3, 3, 5, 5`. Proper per-syllable clusters.
- `"น้ำ"` → clusters `0, 0, 0, 0`. **One cluster, four glyphs**, because the shaper merges the SARA
  AM decomposition.

`buildToUnicodeCMap` today inverts `GlyphForRune`. Under shaping that map has **no entry for a
ligature or for a substituted mark form**, so those CIDs would get **no `/ToUnicode` entry at all**
and the text would stop being extractable or copyable. This is a silent correctness loss of exactly
the kind D-000.23 exists to catch: nobody would notice until someone tried to select the text.

The naive repair — map every glyph in a cluster to that cluster's full text — is **also wrong**: it
would make `"น้ำ"` extract as `"น้ำน้ำน้ำน้ำ"`. See **DECISION NEEDED D-2.3-Q1**.

### F8 — `epics.md:753`'s cross-validation collides with a shipped guard

`epics.md:753` requires shaped output be *"cross-validated against `go-text/typesetting` for the
same input."* Measured: `gomod_test.go:59` declares

```
var wantModuleGraph = []string{
	"github.com/panitw/folio/folio-go",
	"github.com/boxesandglue/textshape",
}
```

and `TestModuleGraphAllowlist` asserts `go list -m all` **equals that set exactly**.
`go list -m all` includes test-only dependencies. **Adding `github.com/go-text/typesetting`, even
for a single test, fails that guard** — the primary assertion D-1.5.1 required, which replaced an
"exactly one module" count precisely so that *anything new must be a conscious act*.

See **DECISION NEEDED D-2.3-Q2**. It is a conscious act, and it is not this story's to make alone.

---

## DECISIONS NEEDED — escalate before development starts

### D-2.3-Q1 — How does `/ToUnicode` map a merged cluster?

**The fork.** Shaping breaks the one-rune-one-glyph assumption `buildToUnicodeCMap` is built on
(F7). Three mechanisms, all implementable:

- **(a) Cluster text on the first glyph, empty on the rest.** CID of the cluster's first glyph maps
  to the cluster's full rune sequence; every other glyph in the cluster maps to the empty UTF-16BE
  string `<>`. `"office"` extracts as `"office"`; `"น้ำ"` extracts as `"น้ำ"`. Standard producer
  behaviour, and what the story recommends.
- **(b) Cluster text on every glyph.** Wrong — `"น้ำ"` extracts four times over.
- **(c) Disable `liga`** so Latin stays one-rune-one-glyph, and handle only Thai. Rejected on its
  face: Thai *still* merges, so it does not remove the problem; and turning off a default-on
  OpenType feature is a typography decision, not a plumbing one.

**Recommendation: (a).** It is the only option that round-trips, and the round-trip is testable
(AC7). **The story is written assuming (a)** so development is not blocked; if the lead rules
otherwise, AC7's mechanism changes and its round-trip assertion does not.

### D-2.3-Q2 — Is `epics.md:753`'s `go-text/typesetting` cross-validation in scope?

**The collision is real** (F8): the dependency cannot be added without amending
`wantModuleGraph`, and that guard exists to make exactly this a deliberate act.

**Three dispositions, framed, not chosen:**

- **(i) Amend the allowlist and do it.** Cost: a second ~1 MB module in the graph and in every
  developer's build, for a test-only benefit. Also worth weighing before treating the result as
  strong evidence: `textshape`'s own README credits `benoitkugler/textlayout` as its inspiration,
  and `go-text/typesetting` descends from the same code. Two ports with shared lineage agreeing is
  **weaker independent evidence than it looks** — the vacuous-citation shape: if both
  implementations would make the same mistake, the agreement predicts the same observation either
  way and you have learned nothing.
- **(ii) Substitute a genuinely independent oracle.** `textshape` ships a **HarfBuzz compatibility
  suite** — 84 `.tests` files in `hb-shape` format, run against HarfBuzz's own reference expectations
  — and HarfBuzz is the actual reference implementation, not a sibling port. Vendoring the handful
  of expectation lines that cover our three faces' scripts into a folio-side fixture gives a
  **stronger** cross-check with **no new module**. This is the story's recommendation.
- **(iii) Defer to the Epic 2 boundary gate** as `DW-2.3-a`, with the frozen expectation table
  (AC2/AC3) standing as this story's correctness evidence in the meantime.

**AC13 is written as (ii) with (iii) as its stated fallback**, and is the one AC whose mechanism the
developer must not choose unilaterally.

### D-2.3-Q3 — `epics.md:756`'s "byte-identical in two separate processes" — which artifact?

`epics.md:756` asks that *"glyph ids and positions"* be byte-identical across two processes. Two
readings:

- **In-process determinism of the shaper.** Already measured true, 5/5 (F6), and it is a property of
  the vendor, not of folio. As an AC it is near-vacuous — it passes at `3373dac`.
- **The rendered PDF's byte-identity**, which is what the project actually cares about and what
  D-000.4's four-target matrix already enforces.

**AC9 implements the second and states the first as a recorded measurement**, because per D-000.21
the artifact that carries the property the product sells is the PDF, not the intermediate buffer.
Flagging it rather than silently reinterpreting.

---

## Acceptance Criteria

Numbering note: assertion groups below follow **D-000.21 (sharpened)** — each group **first asserts
that the field or table it reads exists**, and fails loudly if it does not. A group that reads a
missing field and reports "no discrepancy" is the defect this rule was written for.

### AC1 — A shaping seam exists in `internal/text`, and it is integer-only

**Given** a text string and one resolved `fontset.Font`
**When** it is shaped
**Then** a new exported function in **`internal/text`** returns an ordered slice of shaped glyphs,
each carrying **glyph id, cluster, x-advance, x-offset and y-offset**, obtained from
`ot.Shaper.Shape` with the default feature set (`Shape(buf, nil)`) after `buf.AddString(s)` and
`buf.GuessSegmentProperties()`
**And** every numeric field is an exact integer type; **no `float32`, no `float64`, no untyped
floating-point literal appears in the new code** (`internal/arch_test.go:54`, AD-23)
**And** the shaped glyph slice is built by ranging a **slice**, never a map (D-1.3.5)
**And** one `ot.Shaper` is constructed per `fontset.Font` and reused, per the vendor's documented
contract (*"A `Shaper` is created once per font and reused across shaping calls"*), never one per
call.

**Red-proof.** Replace the shaper call with the rune→`cmap` loop it supersedes: AC2's `ปั` row and
AC3's `AV` row both redden. Measured pre-fix at `3373dac`: the rune→`cmap` loop returns `[80 45]`
for `ปั`, which is the expected-table's *wrong* answer.

**Do not put this in `internal/fontset`.** `internal/text` is the package AD-25 binds and already
owns text-shaped concerns (`break.go`, `cluster.go`, `trie.go`). `internal/fontset` owns the face
and the subset.

### AC2 — GSUB observably fires, asserted from a declarative frozen table

**Given** the frozen expectation table
**When** each row's text is shaped through the named shipped face
**Then** the returned **glyph id sequence** equals the row's expected sequence exactly
**And** the table is a **declarative spec** — one record per case: face name, input text, expected
glyph ids, expected clusters, expected x-advances, expected x-offsets, expected y-offsets — and the
test asserts **spec equals artifact**, per D-000.23's consequent obligation, so that a case added
later inherits every assertion automatically
**And** the table **must contain, at minimum**, these rows, whose expected values are recorded here
and are the values a correct implementation produces:

| Face | Text | Expected glyph ids | Expected clusters |
|---|---|---|---|
| Noto Sans Thai | `ปั` | `80, 46` | `0, 0` |
| Noto Sans Thai | `ฟั` | `16, 46` | `0, 0` |
| Noto Sans Thai | `ที่` | `117, 94, 44` | `0, 0, 0` |
| Noto Sans Thai | `น้ำ` | `71, 59, 49, 86` | `0, 0, 0, 0` |
| Noto Sans Thai | `ป้ำ` | `80, 60, 49, 86` | `0, 0, 0, 0` |
| Noto Sans Thai | `ณัฐวุฒิ` | `70, 45, 118, 134, 97, 116, 92` | `0, 0, 2, 3, 3, 5, 5` |
| Noto Sans Thai | `เกิด` | `91, 29, 92, 12` | `0, 1, 1, 3` |
| Noto Sans | `office` | `82, 1656, 70, 72` | `0, 1, 4, 5` |
| Noto Sans | `fi` | `1654` | `0` |
| Noto Sans | `AV` | `36, 57` | `0, 1` |
| Noto Sans SC | `结算单` | `21201, 20375, 10118` | `0, 1, 2` |

**And** the table must record, for each row, whether the row is **shape-observable** — i.e. whether
its expected glyph ids or positions differ from the naive rune→`cmap` answer — and the test must
assert **at least one observable row per script that has one**, with a printed coverage witness
(`N of N rows evaluated, M observable`)
**And** the test must **fail loudly if the observable count for Thai or Latin is zero** (D-000.9
vacuity guard). For CJK the observable count is **legitimately zero** and the table must say so in
so many words, so that a later reader does not "fix" it.

**Vacuity note, binding.** `结算单`, `กรุงเทพ`, `Hello` and `Ada ก 汉` all shape to themselves (F1,
F3). **No AC in this story may be satisfied by those alone.** They belong in the table as
*negative* controls — proof that shaping does not gratuitously change text that has no rules — and
must be labelled as such.

**Red-proof.** Change the `ปั` row's expected `46` to `45` (the naive answer): the test reddens
naming that row. Independently: neuter the shaper call as in AC1 and every observable row reddens
while every negative-control row stays green — which is itself the proof that the negative controls
are negative controls.

### AC3 — GPOS observably fires, on advances *and* offsets, over the same table

**Given** the same frozen table
**When** each row is shaped
**Then** the returned **x-advance, x-offset and y-offset** for every glyph equal the row's expected
values exactly, including:

| Face | Text | Notable expected position values |
|---|---|---|
| Noto Sans | `AV` | `XAdvance` of glyph 36 is **599**, not its `hmtx` 639 — GPOS kerning, −40 |
| Noto Sans | `Wo. To,` | **two** glyphs whose `XAdvance` differs from `hmtx` |
| Noto Sans Thai | `ฟั` | mark 46 has `XAdvance 0`, **`XOffset +21`** |
| Noto Sans Thai | `ที่` | marks 94 and 44 each have **`XOffset −3`** |
| Noto Sans Thai | `น้ำ` | glyph 49 has **`XOffset −29`**; glyph 86 has `XAdvance 406` |
| Noto Sans Thai | `ป้ำ` | glyph 49 has **`XOffset −204`** |
| Noto Sans SC | `结算单` | every `XAdvance` exactly **1000**, every offset **0** |

**And** `YOffset` is asserted for **every glyph of every row** — expected **0** throughout in these
three faces — so that a face which positions vertically changes the table rather than slipping
past it
**And** the assertion set covers **all five** fields (`GlyphID`, `Cluster`, `XAdvance`, `XOffset`,
`YOffset`) for **every** glyph, not a selected property of a selected sample (D-000.23).

**Why every field, spelled out.** D-000.23: *the field that just burned you is the one most likely
already fixed.* The field that is about to burn this project is `GlyphID` (the Thai lowered forms).
Guarding only that would leave the four position fields uncovered. Guarding only `XOffset` — the
field that visibly moves — would leave `YOffset` uncovered, and `YOffset` is the field a fourth face
would use. **Count the places the meaning lives, then assert all of them.**

**Red-proof.** Zero out `XOffset` after shaping: the `ป้ำ` and `น้ำ` rows redden. Substitute
`Face.HorizontalAdvance` for `GlyphPos.XAdvance`: the `AV` and `Wo. To,` rows redden (this is
the F5 hazard made into a test). Zero out `YOffset`: **nothing reddens** — which is the correct and
expected result at these faces, and the reason the `YOffset` assertion is documented as a
*forward* guard rather than claimed as red-proved. **Say this in the test comment; do not claim a
red-proof you do not have** (D-000.13, D-000.9 extended to probes).

### AC4 — The naive rune→glyph path no longer produces a rendered glyph

**Given** the render path from `.folio` template to content stream
**When** the glyphs that reach the PDF are traced
**Then** **every** CID emitted for a text run originates in a shaped glyph run, and **no** CID
originates in a direct `cmap.Lookup` of a rune
**And** this is asserted **positively** — by reading the emitted CIDs off the produced content
stream and matching them against the shaped run — never by a denylist, a grep, or an assertion that
some function is not called (D-000.23: *denylists are never coverage*).

**Why this AC exists.** Without it, a developer can add a fully correct shaping package that nothing
calls, and every hash test in the repo stays green (F3). This is the AC that proves the seam is
*wired*, and it is the one whose absence D-000.9 predicts.

**`Font.HasGlyph` and `Font.AdvanceForRune` stay.** `HasGlyph` is Story 2.2's coverage test and
drives the missing-glyph diagnostic — it is a *coverage* question, correctly answered per rune,
before shaping. `AdvanceForRune` positions the *second and later face-segments* of a multi-face run
and is not on the per-glyph path. Neither is a rune→CID route into the content stream.

**Red-proof.** Route one run's CIDs through the old `GlyphForRune` map: the assertion names the
diverging CID and the run.

### AC5 — Subsetting is keyed on the glyphs actually drawn

**Given** a document's shaped glyph runs
**When** a face is subsetted
**Then** the subset input is the set of **glyph ids the shaped runs actually contain**, not the set
of runes the document contains
**And** the returned mapping is keyed by **source glyph id → subset glyph id**, obtained from
`plan.MapGlyph`, and covers **every** glyph id in every shaped run for that face
**And** a shaped glyph id the plan did not retain is a **located error naming the face and the
glyph id**, never a silent substitution and never `.notdef`
**And** `Subset` still does no instancing, no sorting of its input, and no ranging of a map
(D-2.2.4, D-1.5.7, D-1.3.5), and its permutation-invariance test is updated to permute **glyph
ids** rather than runes, preserving AC8a's discriminating power.

**This is D-1.5.8's own rule applied one level up.** That ruling said the subset tag must not name
glyphs that were never embedded — *assert on the produced thing, never on the thing you asked for*.
The same logic says the subset must be built from the glyphs the renderer draws, not from the runes
the author typed. They are measurably different sets (F1: `office` draws glyph 1656, which no rune
maps to).

**Expected, not a defect: the subset tag changes, and so do the font-program bytes.** The tag is
derived from the returned program bytes (D-2.2.2 superseded, AC6 of Story 2.2), so a different
input glyph set yields a different tag. `/BaseFont` must still read `TAG+NotoSansSC-Regular` etc.
(D-2.2.6 amended) — **only the tag moves, never the derivation.** Assert the shape of `/BaseFont`,
not a literal tag value.

**The developer must record the measured program-size delta** for each of the three faces
(rune-derived input versus shaped-glyph input) in the completion notes, against the baseline in F4
(50,832 / 1,600 / 1,252 bytes). Do not predict it; measure it. If any face's program grows, say by
how much and why — it bears on NFR7's payload budget, which the Epic 2 boundary gate is tracking.

**Red-proof.** Feed the rune-derived glyph set while drawing shaped glyphs: the `office` case fails
with "shaped glyph id 1656 was not retained" — measured at baseline to be reachable, since
`GlyphForRune` has no entry for it.

### AC6 — The content stream expresses offsets and kerning, in integers, or fails closed

**Given** a shaped glyph run
**When** it is written to the content stream
**Then** it is emitted as a **`TJ` array** — CID hex strings interleaved with integer adjustments —
rather than a bare `Tj` string, whenever any glyph carries a non-zero `XOffset` or an `XAdvance`
differing from its `/W` width
**And** each adjustment is computed as: **before** glyph *i*, `−XOffset_i`; **after** glyph *i*,
`XOffset_i + (W_i − XAdvance_i)`, where `W_i` is the glyph's `/W` array width and every font-unit
value is scaled to the 1000-unit em by **`geom.ScaleRound`** — this module's one scaling function,
one documented rounding mode (AD-2, AD-3)
**And** the adjustment numbers reach the output through **`internal/pdf/numbers.go`'s existing
integer writer**, never through a new emitter and never through any float formatting (AD-3: *"no
number reaches an output byte by any other route"*)
**And** a glyph carrying a **non-zero `YOffset`** is a **located error naming the face, the glyph id
and the offset** — because `TJ` cannot express a vertical offset and the alternative (splitting the
run and emitting a fresh text matrix) is not built here
**And** a run in which every adjustment computes to zero emits **exactly the bytes it emits today**,
so the three existing fixtures' hashes do not move (F3).

**On the fail-closed `YOffset` branch.** Measured: `YOffset` is zero for every glyph of every sample
across all three shipped faces (F2), so **this branch is unreachable with the shipped set and cannot
be red-proved through the render path.** Test it directly, by handing the emitter a synthetic shaped
run with a non-zero `YOffset`, and say plainly in the test comment that the production trigger does
not exist today. **A guard that fails closed on a case you cannot reach is the correct answer here;
silently dropping the offset is not** — that is the "absence reads as success" shape, where the
healthy output and the broken output are the same bytes.

**Red-proof.** Drop the post-glyph adjustment term: the `AV` kern case emits a `TJ` array whose
recorded bytes differ, and AC10's fixture hash moves. Emit `Tj` unconditionally: the `ป้ำ` case
loses its `−204` and the fixture hash moves.

### AC7 — `/ToUnicode` stays complete and round-trips

**Given** a rendered document containing a ligature and a merged Thai cluster
**When** its `/ToUnicode` CMap is read back
**Then** **every CID that appears in the content stream has an entry** — asserted by first
collecting the CID set off the *produced* content stream, then checking the CMap covers it, and
**failing loudly if the CMap object is absent** (D-000.21 sharpened: assert the table exists before
asserting about its contents)
**And** concatenating each run's CIDs through the CMap reproduces the **original source text
exactly**, for `office`, for `น้ำ`, and for `ป้ำ`
**And** the mechanism is D-2.3-Q1(a) unless the lead rules otherwise: the cluster's runes on the
cluster's **first** glyph, empty UTF-16BE on the rest.

**Why "round-trip" and not "has entries".** An assertion that every CID has *an* entry is satisfied
by mapping all four glyphs of `น้ำ` to `"น้ำ"`. The round-trip is what distinguishes correct from
plausible, and it is the assertion the naive repair fails.

**Red-proof.** Map every glyph of a cluster to the cluster's full text (option (b)): `น้ำ`
round-trips as `น้ำน้ำน้ำน้ำ` and the assertion reddens naming the surplus. Drop the ligature's
entry entirely: `office` round-trips as `"oce"` and reddens.

### AC8 — Shaping runs inside a face-segment, and the chain still decides the face

**Given** a text run whose characters resolve to more than one face in the declared fallback chain
**When** it is rendered
**Then** the run is first segmented by **Story 2.2's per-rune, chain-ordered coverage resolution,
unchanged**, and shaping runs **once per contiguous face-segment**, never across a face boundary
**And** `buf.GuessSegmentProperties()` is called per segment, so each segment's script and direction
are derived from that segment's own text
**And** the missing-glyph diagnostic still fires on a rune covered by no face in the chain, naming
the element id and the offending rune, **before** shaping is attempted — asserted by
`ac4_coverage_test.go`'s existing tests continuing to pass unmodified.

**Why this is stated rather than assumed.** `GuessSegmentProperties` derives the script from the
buffer's contents. Handing it `"Ada ก 汉"` in one call would let one script's rules govern another's
characters. The chain already produces exactly the segmentation shaping needs; use it.

**Red-proof.** Shape `"Ada ก 汉"` as a single buffer against a single face: the Thai and CJK
characters resolve to `.notdef` in Noto Sans and the run's glyph ids diverge from AC2's expectation.

### AC9 — Determinism and target coverage

**Given** the same template, data and parameters
**When** the document is rendered on each of the four D-000.4 matrix targets
**Then** the output bytes are identical across all four, including the new fixture from AC10
**And** `CGO_ENABLED=0` builds succeed for `js/wasm` and for the native targets — **a regression
guard, not evidence this story landed**: measured true at `3373dac` before any shaping code exists
(F6), and labelled as such in the test comment
**And** the developer records, in the completion notes, that shaping is process-stable — five
invocations of the same shaping over the same inputs produced one distinct output digest at
baseline (F6).

**Matrix cadence.** This story is **not** a D-000.4 per-story override. D-000.4's per-story overrides
are 1.2, 1.5, 1.8, 2.4 and 4.7; 2.3 is not among them, and the four-target matrix therefore runs at
the **Epic 2 boundary gate**. **However**, AC10 records a golden fixture, and AD-21 binds every
feature to ship its golden. The developer must **write** the matrix legs for the new fixture and
**register** it in the harness, then state in the completion notes that they run at the epic
boundary. Do not run Docker legs in this story; do not silently skip registering them either.

### AC10 — A new golden fixture, with a semantic acceptance step (D-000.22)

**Given** this story's shaping behaviour
**When** a golden fixture is recorded
**Then** a **new** fixture directory `fixtures/shaped-text/` is added — **beside** the three
existing fixtures, replacing none of them (the D-1.8.6 / D-2.2-D5 precedent: add a document, never
switch one)
**And** it contains `input.folio`, `expected.json` (the normative hash, plus the recorded toolchain,
matching the other fixtures' value — the shared-toolchain assertion), `expected.pdf` (for human
diffing only) and a `README.md` carrying the same "if this hash goes red" rule the other three carry
**And**, because this is a **first recording**, the following **semantic properties are asserted off
the produced PDF**, separately from and additionally to the hash (D-000.22, D-000.21 sharpened —
each assertion first proves the object it reads exists):
  1. the content stream contains at least one **`TJ`** array with a **non-zero** adjustment (proves
     AC6's mechanism is live, not merely written);
  2. the CID set emitted for the Latin face **contains the ligature CID** and does **not** contain
     the CIDs of its unshaped components in that position (proves GSUB reached the page);
  3. the CID set emitted for the Thai face contains the **lowered** mark form and not the unlowered
     one (proves the defect this story exists to fix is fixed);
  4. `/ToUnicode` round-trips the document's full source text (AC7);
  5. `/BaseFont` reads `TAG+<PostScriptName>` for each embedded face, `TAG` six uppercase letters
     (D-2.2.6 amended);
  6. the embedded programs carry **no `fvar` and no `gvar`** (D-2.2.4, carried forward from Story
     2.2's semantic set — a face regressing to variable must redden here too);
  7. the page count and declared page size are what the template declares.
**And** — the part the machine cannot do — **a human looks at the rendered Thai before the fixture
is frozen** and confirms the marks sit on their bases, exactly as Story 2.2 required a human to look
at the rendered Chinese. **Record in the completion notes that this happened and what was seen.**

**Why the human step is not ceremony.** D-000.22, verbatim: *a hash guards against change; it has
nothing whatever to say about whether what we recorded was right, and it cannot acquire that ability
later.* The moment this fixture is committed it becomes an **input**, and every downstream guard
agrees with it forever. Story 2.2 shipped Thin Chinese past three correct guards because none was
asked whether the value *meant* what its name implied — **and the face that broke was the one whose
script nobody on this project reads.** Thai is that face this time.

### AC11 — The fixture's text is shape-observable, and that is asserted, not assumed

**Given** `fixtures/shaped-text/input.folio`
**When** its text is examined
**Then** a test asserts that **for each of the three shipped faces, the fixture's text contains at
least one shape-observable segment** — one whose shaped glyph ids or positions differ from the naive
rune→`cmap` answer — **except CJK, for which the expected observable count is zero and the test
asserts exactly that**, with the reason stated inline
**And** the assertion prints a **coverage witness**: `N of N fixture text segments evaluated, M
shape-observable`
**And** the fixture's text must include, at minimum, a Thai tall-consonant-plus-mark case, a Thai
SARA AM case, and a Latin ligature or kern pair.

**Why this AC exists, and it is the most important one in the story.** F3 measured that **every
existing golden is blind to shaping**. Without AC11, the new fixture can be recorded over text that
is equally blind, and the project acquires a fourth fixture that cannot fail when this story
regresses — while looking, in every report, like coverage. That is D-000.9's exact shape: the "all
clear" and the "I could not look" produced by the same code path. **AC11 is the coverage witness for
AC10.**

**Red-proof.** Point the assertion at `fixtures/multi-script-fallback/input.folio` (`"Ada ก 汉"`):
it must redden with `0 shape-observable` for Latin and Thai. This red-proof is **available today**,
against a committed file, and the developer should run it first.

### AC12 — The float boundary is named and held

**Given** the new shaping code
**When** `internal/arch_test.go`'s float guard runs
**Then** it passes — no `float32`, no `float64`, no untyped float literal anywhere under
`internal/` or the module root (AD-23)
**And** the new code takes advances from **`ot.GlyphPos.XAdvance`** (`int16`) and never from
`ot.Face.HorizontalAdvance` (`float32`), with a comment recording **why**: the accessor returns a
float the arch guard cannot see through an `int64(x)` conversion (F5), *and* it omits GPOS kerning,
so it is wrong on both counts for shaped text
**And** the developer reports, in the completion notes, whether any **other** call site under
`internal/` converts a vendor `float32` through an `int64(...)` conversion the guard cannot see —
`internal/fontset/fontset.go`'s `AdvanceForRune` and `Metrics` are the known ones at baseline.

**Not in scope: widening the guard.** Whether `arch_test.go` should learn to flag float-typed vendor
*expressions* is a real question and a larger one; it touches every existing call site and it is not
this story's. **Report it; do not fix it.** If the count is non-trivial, it belongs in the Epic 2
boundary gate beside the payload figures.

### AC13 — Cross-validation against an independent oracle *(mechanism: pending D-2.3-Q2)*

**Given** the frozen expectation table (AC2/AC3)
**When** its expected values are cross-checked
**Then** **either** (ii) a folio-side fixture carrying the relevant HarfBuzz reference expectations
— from `textshape`'s own `harfbuzz-tests/` corpus, for the scripts and faces folio ships — confirms
the table, **with no new module in the graph** and `TestModuleGraphAllowlist` untouched
**Or** (iii) the cross-validation is deferred to the Epic 2 boundary gate as **`DW-2.3-a`**, with
the reason recorded in `deferred-work.md` and the expectation table standing as this story's
correctness evidence.

**`go-text/typesetting` (option (i)) may not be added without an explicit ruling** amending
`gomod_test.go:59`'s `wantModuleGraph`. `epics.md:753` names it; `TestModuleGraphAllowlist` forbids
it; the guard is newer and load-bearing (D-1.5.1). See F8 and D-2.3-Q2. **`epics.md` predates the
guard.**

**If (iii) is chosen, say so explicitly in the story's completion notes and in `deferred-work.md`,
and do not describe the story as having cross-validated anything.**

### AC14 — Standing guards stay green

**Given** the full guard suite
**When** it runs
**Then** `no-compressor-import` passes — no compressor is imported under `folio-go/` (D-1.8.1), and
the `TJ` array's extra bytes are accepted uncompressed
**And** `absence-source-date-epoch` passes — the literal string does not appear in any `.go` file
under `folio-go/`, **including inside error-message strings** (D-2.1.5)
**And** `embed-font` passes — no package under `internal/` embeds font data (AD-8)
**And** `TestModuleGraphAllowlist` and `TestModuleGraphDenylistsKnownPDFWriters` both pass, with
`wantModuleGraph` **unchanged** unless D-2.3-Q2 rules otherwise
**And** `internal/text`'s two inherited failures remain exactly two, unchanged, with unchanged
messages — **not fixed, not tuned, not skipped** (D-2.1.14)
**And** `fixtures/minimal-rect/`, `fixtures/image-embed/` and `fixtures/multi-script-fallback/`
hashes are **byte-identical to baseline** — these are the three fixtures genuinely blind to this
story, re-measured in development: `minimal-rect` is fontless, `image-embed` bears no text, and
`multi-script-fallback`'s `"Ada ก 汉"` shapes to itself on all three shipped faces. If one of
these moves, this story has a defect.

**AMENDED IN DEVELOPMENT (D-2.3.4).** This clause originally named `fixtures/font-text/` in place
of `fixtures/multi-script-fallback/`, on the strength of F3. **F3 was wrong about `font-text`, and
that fixture legitimately MOVED.** F3 measured the string `"Hello"` against a shipped Noto face;
`font-text` actually renders `"Hello, World!"` and `"Page footer 0123456789"` through
`folio-go/testdata/fonts/Roboto-Regular.ttf` at `unitsPerEm` 2048 — a different string *and* a
different face. Roboto's `GPOS` kerns `W` (gid 59: hmtx 1817, shaped 1786, −31) and `P` (gid 52:
hmtx 1292, shaped 1281, −11), so the pre-2.3 bytes recorded **unkerned** output: two `Tj`
operators, no `TJ`, 22,299 bytes. Cross-checked against `hb-shape` (HarfBuzz) 14.2.0 before
re-recording, which reports the same kerned advances.

**`fixtures/font-text/` has therefore moved from the BLIND list to the OBSERVING list**, and
carries a D-000.22 semantic acceptance step recording why: its guard asserts the operator is `TJ`
(never a bare `Tj`) and asserts the adjustments **`15` and `6` by value, not by hash**, so that a
future loss of kerning cannot be answered by re-recording a digest without deleting a visible
assertion. Both value assertions were red-proved independently (drop the advance-correction term;
then perturb it by one unit) and restored green. Provenance — tool version and exact invocations,
verbatim — is recorded in `fixtures/font-text/README.md`.

---

## Task breakdown

1. [x] **Restore and run the baseline evidence.** Reproduce F1/F3's measurements — the naive rune→`cmap`
   answer versus the shaped answer, for the AC2 table's rows and for the three existing fixtures'
   text. Confirm the numbers in F1 and F3 reproduce at your HEAD. If they do not, the baseline has
   moved and the evidence needs re-taking before anything else.
2. [x] **Run AC11's red-proof first**, against `fixtures/multi-script-fallback/input.folio`. It must
   report `0 shape-observable` for Latin and Thai. This is available today and costs nothing.
3. [x] Write the shaping seam in `internal/text` (AC1) and the declarative expectation table (AC2, AC3).
4. [x] Re-key `fontset.Subset` on glyph ids and update its permutation-invariance test (AC5). Record the
   program-size deltas.
5. [x] Wire shaping into the render path per face-segment (AC8), and prove it is wired positively (AC4).
6. [x] Emit `TJ` with integer adjustments; fail closed on `YOffset` (AC6).
7. [x] Rebuild `/ToUnicode` from clusters and assert the round-trip (AC7).
8. [x] Author `fixtures/shaped-text/input.folio` so AC11's observability assertion passes, then record
   the fixture, run the semantic acceptance step, **look at the rendered Thai**, and only then freeze
   the hash (AC10).
9. [x] Register the new fixture's four matrix legs; do **not** run the Docker legs (AC9).
10. [x] Confirm AC14's guards and the three unchanged fixture hashes.
11. [x] Escalate D-2.3-Q1 and D-2.3-Q2 if they are still open when you reach steps 7 and 13.
12. [x] Write the completion notes: program-size deltas, the vendor-float call-site count, the human
    check on rendered Thai, which suites were deferred and to which gate.
13. [x] Set the story status to `review`.

**There is no commit task.** This story ends at `review`; the finisher commits.

---

## Heavy-test cadence — what is deferred, and to which gate

Epic 2's cadence is **per epic** (D-000.4), and 2.3 is **not** one of D-000.4's per-story overrides.

**Deferred to the Epic 2 boundary gate:**

- The **four-target byte-identity matrix** (`darwin/arm64`, `linux/amd64`, `linux/arm64` under
  Docker, `js/wasm`) over all four fixtures including `fixtures/shaped-text/`. This story **writes
  and registers** the legs; it does not run the Docker ones.
- `fontgen_matrix_test.go`'s `//go:build matrix` reproduction of the shipped faces — untouched by
  this story and unaffected by it.
- **AC13's cross-validation**, if D-2.3-Q2 resolves to (iii).

**Run in this story** (its correctness is not integration-shaped): the whole `go test ./...` suite
natively, the lint rule suite, and the native `js/wasm` build. Report the result as *"N pass, 2 fail
(the two inherited from Story 2.1)"* — never as *"2 failures"* without that clause.

---

## Cross-read of the ACs — universally-quantified requirements against scoped implementations

Per the standing habit of checking new ACs for the collision pattern *a universally-quantified
requirement meeting a deliberately-scoped implementation*. Four found; all four resolve in the file
above rather than at development time.

1. **AC3's "`YOffset` asserted for every glyph" vs. "`YOffset` is 0 everywhere."** Collision is real
   and the resolution is stated: the assertion is a **forward** guard with **no available
   red-proof**, and the test comment must say so rather than claim one.
2. **AC2's "at least one observable row per script" vs. CJK, which has none.** Resolved by carving
   CJK out explicitly and asserting its observable count is **exactly zero**, with the reason
   inline — an assertion, not an omission.
3. **AC6's "emit `TJ`" vs. AC14's "the three existing hashes must not move."** Resolved by the
   conditional: a run whose adjustments all compute to zero emits today's bytes exactly. Both hold
   because F3 measured that those fixtures produce no non-zero adjustments.
4. **AC5's "subset the glyphs actually drawn" vs. D-2.2.6's `/BaseFont`.** Resolved by separating
   the tag (which **will** change, expectedly) from the derivation (which must not). Assert the
   *shape* of `/BaseFont`, never a literal tag.

---

## Where `epics.md` and post-2.2 reality disagree

| `epics.md` | Reality at `3373dac` | Disposition |
|---|---|---|
| `:753` cross-validate against `go-text/typesetting` | `TestModuleGraphAllowlist` (`gomod_test.go:59`) asserts the module graph equals exactly two modules; a test-only dep fails it | **DECISION NEEDED D-2.3-Q2**; AC13 written as (ii)/(iii) |
| `:751` marks positioned by GPOS mark-to-base | True, but the observable in these faces is `XOffset` + a **GSUB** substitution; `YOffset` is 0 throughout | AC2 + AC3 cover both mechanisms and all five fields |
| `:747` no cgo, `js/wasm` builds | **Already true at baseline**, with no shaping code | AC9 keeps it as a labelled *regression* guard, never as evidence |
| `:756` glyph ids and positions byte-identical across two processes | True of the vendor at baseline (5/5); the artifact the product sells is the PDF | **D-2.3-Q3**; AC9 asserts the PDF, records the shaper measurement |
| `:718` (Story 2.2) CJK face is *variable* | Amended to **static** under D-000.6 (D-2.2.5) | Already amended; AC10's semantic set carries the `fvar`/`gvar` absence forward |
| Epic 2 preamble: *"Thai words break by dictionary"* | D-2.1.6 (owner) made unbreakability **declarative**; the work is 2.4/2.5 | Fenced out of this story entirely |

---

## Completion notes

*(filled by the developer; every number below is from a real run at this HEAD, none carried forward)*

### Test result, stated with the inherited-failure clause

**369 pass, 2 fail — the two inherited from Story 2.1** (`internal/text`'s
`TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck`, D-2.1.14). Baseline at
`3373dac` was 331 pass, 2 fail. **+38 tests, no third failure.** Not fixed, not tuned, not skipped.

Other gates, all run in this story:

| gate | result |
|---|---|
| `go vet ./...` | clean |
| `go vet -tags=matrix ./...` | clean — the registered matrix legs compile |
| `gofmt -l` (`folio-go` and `lint`) | empty |
| lint rule suite, `GOPROXY=off go test -count=1 ./...` | 4 packages, all pass |
| `GOOS=js GOARCH=wasm CGO_ENABLED=0 go build ./...` | succeeds |
| `GOOS=linux GOARCH=amd64/arm64 CGO_ENABLED=0 go build ./...` | both succeed |
| `go list -m all` | **exactly two modules** — `wantModuleGraph` untouched |

Witnesses printed by the new tests:

- shaping table: `Thai: 9 of 9 rows evaluated, 7 observable; Latin: 5 of 5, 4 observable; CJK: 2 of 2, 0 observable`
- fixture observability: `21 of 21 fixture text segments evaluated, 9 shape-observable (Thai 8/10, Latin 1/9, CJK 0/2)`
- AC11 red-proof: `fixtures/multi-script-fallback/: 4 of 4 segments evaluated, 0 shape-observable`
- AC4: `3 of 21 emitted runs have a glyph count differing from their rune count`
- AC13: `16 of 16 expectation rows checked against hb-shape (HarfBuzz) 14.2.0, all five fields per glyph`
- matrix registration: `5 documents registered in both matrixDocuments and matrix.yml, across 4 targets`
- semantic acceptance: `9 of 21 emitted runs use TJ, 9 carry a non-zero adjustment`

### Program-size delta per face (AC5)

Measured on F4's own inputs. **F4's baseline reproduces exactly** (50,832 / 1,600 / 1,252):

| face | rune-derived input | shaped-glyph input | program delta |
|---|---|---|---|
| Noto Sans, `"office fi Acme"` | 14 gids in → 24 out, **50,832 B** | 11 gids in → 24 out, **50,832 B** | **+0** |
| Noto Sans Thai, `"ปั ที่ ป้ำ น้ำ"` | 14 gids in → 19 out, **1,600 B** | 16 gids in → 13 out, **1,240 B** | **−360** |
| Noto Sans SC, `"结算单"` | 3 gids in → 4 out, **1,252 B** | 3 gids in → 4 out, **1,252 B** | **+0** |

**No face grew; Thai SHRANK by 360 bytes (−22.5%).** That is worth understanding rather than
banking, because the direction is counter-intuitive: the shaped input has *more* entries (16 vs 14)
yet yields *fewer* output glyphs (13 vs 19). The rune-derived input asks for the *unshaped* mark
glyphs (45, 47) whose `GSUB` closure then drags in every substitution target they can reach; the
shaped input asks for the glyphs actually drawn (46, 59, 60, 49), whose closure is smaller. Asking
for what you draw is cheaper than asking for what you typed. For NFR7's payload budget this is
neutral-to-favourable, and the Epic 2 boundary gate should treat it as such rather than as a
saving to rely on — it is a property of these three faces' `GSUB` tables, not a general law.

`fixtures/shaped-text/` itself embeds 51,572 + 29,324 + 3,364 = **84,260 B** of font program in a
**90,931 B** document. The Thai figure (29,324) is large relative to the table above because this
fixture's text covers far more of the face.

### Vendor-`float32` call sites reached through an `int64(...)` conversion (AC12 — report only)

**Two, both in `internal/fontset/fontset.go`, both `ot.Face.HorizontalAdvance`:**

| site | line | what it feeds |
|---|---|---|
| `(*Font).AdvanceForRune` | `internal/fontset/fontset.go:316` | positions the second and later face-segments of a multi-face run |
| `(*Font).Subset`'s `/W` width loop | `internal/fontset/fontset.go:511` | the PDF `/W` array's per-glyph width |

Both spell `int64(adv)` where `adv` is `float32`, which `internal/arch_test.go`'s guard **cannot
see**: it flags the *identifiers* `float32`/`float64` and a bare `int64(x)` names neither. Both are
safe today — every `hmtx` advance is exactly representable in `float32` — and both are *correct in
their own right*: `/W` is defined as the `hmtx` advance, and `AdvanceForRune` is a segment-level
cursor, not a per-glyph position.

> **CORRECTED BY THE FINISHER (Blocker 1). The second half of that last sentence is false, and the
> table's "what it feeds" column for the first row is now stale.** `AdvanceForRune` was *not* correct
> as a segment-level cursor: it reports the **unkerned** `hmtx` advance, while runs are drawn
> **kerned**, so summing it over a segment's runes placed the following face-segment as though the
> kerning had not happened — measured at **640 millipoints (0.64 pt)** for `"AV ก"` at 16 pt in the
> shipped chain. **The float half of the note stands; the shaping half does not.** The segment cursor
> now derives from the shaped advances and `AdvanceForRune` has no caller on the render path. The
> **count of vendor-`float32` sites is still two** and both are still
> `2-3a-audit-the-vendor-boundary`'s (D-000.25) — the accessor was deliberately **retained**, not
> deleted, so that story's subject does not silently shrink. See § Finding Resolutions, Finding 1.

**One correction to AC12's own text:** it names *"`AdvanceForRune` and `Metrics`"* as the known
sites. **`Metrics` is not one.** Its four accessors (`Ascender`, `Descender`, `CapHeight`, `BBox`)
all return `int16` in `textshape` v0.0.15 — verified in `ot/metrics.go`. The count is two, not
three.

**Not fixed, not widened, as instructed.** This belongs to the already-scheduled
`2-3a-audit-the-vendor-boundary` (D-000.25: vendor-boundary audit + type-aware AD-23 guard), which
sprint-status already sequences before 2.4 precisely so it lands before more vendor call sites
arrive. The new shaping path adds **zero** such sites: it takes advances from
`ot.GlyphPos.XAdvance` (`int16`), which is both integer-typed *and* the only one that includes GPOS
kerning — for `AV` in Noto Sans, 599 against an `hmtx` 639. The accessor is wrong about the number
as well as the type.

### Visual check on the rendered Thai (AC10) — DEVELOPER CHECK DONE, HUMAN SIGN-OFF OUTSTANDING

**Stated plainly because the distinction is the entire point of D-000.22: I am the developer agent,
not a human, so AC10's human step is NOT satisfied by this note.** What I can report is a
developer-side visual check, done properly, with the artifacts left in place for a human to confirm
in about thirty seconds.

**Method.** Rasterised `fixtures/shaped-text/expected.pdf` at 2828×4000, then rendered a **second,
deliberately broken** document: `internal/text.Shaper.Shape` temporarily replaced by the
rune→`cmap` loop it supersedes (AC1's red-proof), rendered, rasterised identically, and the two
compared line by line. The mutation was reverted from a `cp` backup and the restored file
`diff`-verified byte-identical.

**What was seen, on the first line (`ปั ฟั ที่ ป้ำ`):**

| | naive (broken) | shaped (shipped) |
|---|---|---|
| `ปั` | MAI HAN AKAT sits **on** ป's ascender stem; strokes touch | mark sits clear to the **left** of the ascender and lower — the `GSUB` lowered form |
| `ฟั` | same collision against ฟ's ascender | mark clear of the ascender |
| `ที่` | tone mark jammed into SARA II, strokes overlapping | two marks cleanly stacked, both legible |
| `ป้ำ` | MAI THO and NIKHAHIT pile onto the ascender into a blob | MAI THO above-left, NIKHAHIT clear to the left, `า` following — four glyphs, all legible |

This is exactly the defect the story's plain-terms section describes ("the mark collides with the
letter it belongs to"), visible and fixed. Latin `office`/`fi` show the ligatures joined in the
shaped render and separate glyphs in the naive one. CJK is pixel-identical between the two, as
predicted.

**Artifacts for the human check** (scratch, not committed):

- shaped, full page: `/private/tmp/claude-501/-Users-panitw-Projects-folio/8923b36a-7c8d-446e-90c8-35679685fef4/scratchpad/hi/expected.pdf.png`
- naive, full page: `/private/tmp/claude-501/-Users-panitw-Projects-folio/8923b36a-7c8d-446e-90c8-35679685fef4/scratchpad/hi/naive.pdf.png`
- line-1 crops: `.../hi/expected-l1z.png` and `.../hi/naive-l1z.png`

**The fixture hash was frozen before this sign-off exists.** That is a deliberate, stated risk, not
an oversight: freezing was needed to run the semantic assertions at all. If the human check
disagrees with what I saw, the fixture must be re-recorded, and that is cheaper than the reverse.
**A Thai reader should confirm before this story is called done.**

### D-2.3-Q1 disposition — RULED, implemented as ruled (and extended)

Ruled: cluster runes on the cluster's **first** glyph, empty UTF-16BE on the rest, asserting the
**round-trip**. Implemented in `internal/text.ClusterTexts`.

Development found one case the mechanism alone does not cover, and it is reachable rather than
theoretical: **the `/ToUnicode` CMap is per-face, but the mechanism's answer is per-occurrence.**
Shaping `"น้ำ"` yields four glyphs whose last is the **same glyph** a standalone `"า"` yields. That
glyph must map to `""` inside the merged cluster and to `U+0E32` when it stands alone, and one CID
cannot carry both. Neither collapse is acceptable: mapping it to `""` loses every standalone `า`,
and mapping it to `U+0E32` makes `น้ำ` extract as NIKHAHIT + SARA AA — which `U+0E33` has **no
canonical decomposition to** and which NFC never recomposes, so a reader searching the PDF for
`น้ำ` finds nothing.

Resolved per the ruling on CID allocation: **CIDs are allocated per (subset glyph, extracted
text)**. CIDs `0..NumGlyphs-1` remain the identity base block and extras append above it, so
`/CIDToGIDMap` still emits `/Identity` and byte-identical output for every document needing no
extras — which is why the three blind fixtures did not move. In `fixtures/shaped-text/` only the
Thai face needs the extension (one extra CID, `0x0024 → U+0E32`). Both `น้ำ` and standalone `า` are
in the fixture **on purpose**, so this case is exercised rather than reasoned about.

### D-2.3-Q2 disposition — RULED (ii), implemented; **not deferred, no `DW-2.3-a`**

The oracle is **HarfBuzz itself** — `hb-shape (HarfBuzz) 14.2.0` — run against **folio's** corpus
as a one-time offline reference run, hand-checked and frozen at
`fixtures/shaped-text/harfbuzz-oracle.json`. **No new module; `go list -m all` is still exactly two
entries and `TestModuleGraphAllowlist` is untouched** (AD-25; Story 1.1's `qpdf --check`
precedent).

16 rows, all five fields per glyph, agreeing with `textshape` **value for value**. Each row records
its exact `argv` verbatim and the SHA-256 of the font file it was run against; `assertOracleSubject`
re-checks that digest against the committed face on every run, so an agreement measured against an
artifact folio no longer ships fails rather than reassures.

Neither rejected option was implemented, not even speculatively. `go-text/typesetting` would have
broken `wantModuleGraph` *and* been weak evidence (shared `benoitkugler/textlayout` lineage);
`textshape`'s own `harfbuzz-tests/` corpus is curated by the vendor and selected to pass.

### D-2.3-Q3 disposition — as written

AC9 asserts the **PDF**, and records the shaper measurement. Both re-measured here:

- **Cross-process, on the artifact that matters:** the built test binary run **5 separate times**
  through the `shaped-text` subprocess selector produced
  `0aeb930a0cc29e7adff2ea757d61f87c2eea033d6dca3dbd6382ef154207b52a` **5 times out of 5** — equal to
  the recorded golden.
- **In-process shaper stability:** 5 runs over the whole 16-row table produced **1 distinct
  digest**.

The `js/wasm` + `CGO_ENABLED=0` clause is kept and **labelled a regression guard**: it was already
true at `3373dac` with no shaping code, so it is never evidence this story landed.

### Suites deferred, and to which gate

**Deferred to the Epic 2 boundary gate**, as the per-epic cadence requires (2.3 is not one of
D-000.4's per-story overrides — those are 1.2, 1.5, 1.8, 2.4 and 4.7):

- The **four-target byte-identity matrix** (`darwin/arm64`, `linux/amd64`, `linux/arm64` under
  Docker, `js/wasm`) over all **five** documents including `fixtures/shaped-text/`. The legs are
  **written and registered** in this story and **not run**: no Docker leg was executed.
  `go vet -tags=matrix ./...` is clean, so they compile.
- `fontgen_matrix_test.go`'s `//go:build matrix` face reproduction — untouched and unaffected.

**Nothing else is deferred.** AC13 is *done*, not deferred, so **no `DW-2.3-a` entry was added to
`deferred-work.md`** and this story **does** claim to have cross-validated.

### Two findings for the gate, reported not fixed

1. **`.github/workflows/matrix.yml` was silently one document short — since Story 2.2.** That
   workflow carries a hand-written second copy of `matrixDocuments` (the `docs="..."` line and the
   per-target `upload-artifact` paths). Story 2.2 registered `multi-script-fallback` in
   `matrix_test.go` and **never in the workflow**, so CI has been publishing four documents' worth
   of legs and comparing three — Story 1.2's own Finding 8 recurring one story later, which is what
   a hand-maintained duplicate list does. Both `multi-script-fallback` and `shaped-text` are now in
   the workflow, and **`TestMatrixDocumentSlugsAreRegisteredInCI`** (untagged, so it runs every
   story, not only at a gate) asserts the two lists agree and that every document has an artifact
   path on every target. It reads source text, which is the weaker mechanism — running in every
   story is what makes it worth having.
2. **`GlyphInfo.Cluster` is a RUNE index, not a byte offset** — F7 says byte offset. `"office"` →
   `0,1,4,5` is ambiguous because it is ASCII; `"ณัฐวุฒิ"` → `0,0,2,3,3,5,5` is not, since byte
   offsets over seven three-byte runes would be multiples of 3. `hb-shape` agrees.
   `internal/text.ClusterTexts` indexes a `[]rune` accordingly and
   **`TestClusterValuesAreRuneIndices`** pins it, so a vendor switch to byte offsets reddens
   loudly instead of silently corrupting `/ToUnicode`.

### Amendments made in development, all recorded at their source

| what | where |
|---|---|
| AC14's byte-identity clause re-aimed; `font-text` moved to the **observing** list | this file, AC14 |
| `epics.md:751` — `GPOS` mark-to-base is true but not the observable; assert all five fields | `epics.md`, amended in place |
| `epics.md:753` — `go-text/typesetting` replaced by HarfBuzz as the oracle | `epics.md`, amended in place |
| `epics.md:756` — clarified to the PDF, not the intermediate buffer | `epics.md`, amended in place |
| `font-text`'s re-record, its `TJ` provenance and its `hb-shape` invocations | `fixtures/font-text/README.md` |
| the new fixture's rationale, semantic set and oracle provenance | `fixtures/shaped-text/README.md` |

---

## QA Results

### Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-24
- **Baseline:** HEAD `3373dac`, story work uncommitted. Working tree captured before the first
  command and re-diffed after the last; `git status --porcelain` is byte-for-byte the list I was
  handed (22 entries, identical). Every mutation below was reverted **by hand from a `cp` backup and
  re-verified by SHA-256** — no `git checkout`, no `git stash`. Two scratch `_test.go` files were
  created and deleted; the tree is clean of them.
- **Suite state:** **369 pass, 2 fail — the two inherited from Story 2.1** (`TestCorpusMeetsP6ExerciseFloors`,
  `TestP2IndependentDPCrossCheck`, D-2.1.14). Counted from `go test ./... -count=1 -v`.
  **No third failure.** The claimed count is exact.
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 2 · **Majors:** 1 · **Minors:** 7 · **Nits:** 2

**This is a strong story.** The four things I was asked to attack hardest all hold up under
construction, and several of them are the best-built guards in this repo so far. The two Blockers
are *not* in the shaping engine — they are one live positioning defect at the seam the story did not
look at, and one false claim in a committed artifact. Details below, verification first.

---

### What I verified by construction (all four high-value attacks clear)

#### 1. D-2.3.4 — the `font-text` re-record's semantic acceptance step: **genuinely red-proved, all three conditions**

I ran the red-proofs myself rather than reading the claim. `fixture_test.go:509-546` asserts on the
**rendered** bytes, and its assertions run **before** the hash comparison, so they are not shadowed
by it.

| mutation (in `internal/pdf/textdoc.go:709`) | result |
|---|---|
| perturb the advance-correction term by **+1** (`width - g.XAdvance + 1`) | **both value assertions redden by name** — `">15<"` and `">6<"` each fail with their full explanatory message, at `fixture_test.go:538`, *before* the hash mismatch is reported |
| **drop** the advance-correction term entirely | `fixture_test.go:510` fires: *"emits no TJ array at all… Tj-only output is the unkerned pre-2.3 behaviour"* |
| drop **only the smaller** kern (`corr` under ±10 → 0), leaving the larger | `fixture_test.go:517` fires: *"still emits a bare Tj operator… a Tj here means one run lost its adjustment"* — this is the discriminating case condition 2 exists for, and it fires |

The value assertions are also **discriminating in the artifact**: `>15<` and `>6<` each occur
**exactly once** in the 22,310-byte PDF (measured), and in a `TJ` array `>` can only be followed by
`<` across an adjustment, so neither pattern can be satisfied by hex-string bytes.

**Provenance replays byte-identical.** `fixtures/font-text/README.md:113-125` records
`hb-shape (HarfBuzz) 14.2.0` and two verbatim invocations. I ran both against
`/opt/homebrew/bin/hb-shape` (confirmed 14.2.0) and the output matched the README **character for
character**, including `59=7+1786` and `52=0+1281`. The scaling checks out:
`ScaleRound(1817,1000,2048)=887`, `ScaleRound(1786,1000,2048)=872`, `887−872=15`; `631−625=6`.

One subtlety worth recording because it looks like a bug and is not: `appendShapedRun` subtracts two
**separately-scaled** values (`W_i` scaled in `fontset.go:512`, `XAdvance` scaled in `render.go:709`)
rather than scaling their difference. That is **correct** — the viewer's pen consumes the scaled
`/W`, so the correction must be expressed in the same already-rounded space. Scaling the difference
would be the defect.

#### 2. D-2.3.2 — `/ToUnicode` CIDs per (GID, Unicode-context): **the corpus can express the defect, and does**

`fixtures/shaped-text/input.folio` element `e3` is `"น้ำ า"` — the merged SARA AM cluster **and** a
standalone SARA AA **in the same document**, which is the corpus obligation D-2.3.2 calls "the
load-bearing half". Confirmed present.

I collapsed the allocation to **per-GID** (`render.go:700-701`, replacing the extra-CID allocation
with `cid = newGID`) — i.e. exactly the implementation D-2.3.2 rules out — and the corpus caught it
precisely, not incidentally:

```
run 12 ("น้ำ"): /ToUnicode round-trips as "น้ำา", want the source text "น้ำ"
run 12 ("น้ำ"): CID 18 at position 3 extracts as "า", but that glyph's shaped cluster text is ""
```

That is the exact failure mode D-2.3.2 names, reproduced from the fixture. `TestShapedTextFixtureToUnicodeRoundTrips`
also guards its own premise (`shaped_fixture_test.go:570-574`: the fixture must still contain
`office`, `น้ำ`, `ป้ำ`, `า`) — and `า` (U+0E32) appears in **no other** fixture string, so that
clause has teeth.

**Nothing else assumes CID↔GID is 1:1.** I checked the three places it could break and all three
handle the 1:N case correctly:
- `/W` (`textdoc.go:361-370`) emits base widths then repeats each extra CID's glyph width;
- `/CIDToGIDMap` (`textdoc.go:562-580`) emits the identity block then the extras, and the dictionary
  correctly switches from `/Identity` to a stream object only when `len(ExtraCIDs) > 0`
  (`textdoc.go:372-379`) — which is why the three blind fixtures stayed byte-identical;
- `glyphForCID` (`textdoc.go:760-769`) resolves extras back to their glyph and falls through to the
  `WidthForGlyph` presence check, a located error rather than a silently wrong width.

`WidthForGlyph` covers `0..NumGlyphs-1` densely (`fontset.go:505-514`), so extras always resolve.

#### 3. D-2.3.1 — the HarfBuzz oracle: **real, and it replays**

I re-ran **every one of the 16 rows'** own `argv` verbatim against `/opt/homebrew/bin/hb-shape`
(`hb-shape (HarfBuzz) 14.2.0`) and compared all six emitted fields per glyph:

```
16 rows, 16 AGREE, 0 disagree — all five asserted fields (g, cl, ax, dx, dy) value-for-value
3 font files, all 3 SHA-256 digests match the committed faces
```

The oracle is a **frozen fixture, not a runtime dependency**: `go list -m all` returns exactly
`github.com/panitw/folio/folio-go` and `github.com/boxesandglue/textshape v0.0.15`.
`TestModuleGraphAllowlist` is untouched and passes.

`shaping_oracle_test.go` is well-built: it checks **both directions** (row-count equality at :124,
plus a per-row lookup at :133), calls `assertOracleSubject` per row so an agreement measured against
a face folio no longer ships fails rather than reassures, carries a `checked == 0` vacuity guard, and
fatals on a missing `toolVersion`/`method`/empty rows.

The table itself is load-bearing in **two independent directions** — I changed the `ปั` row's
expected glyph `46` to the naive `45` and both reddened by name:

```
shaping_expectations_test.go:441: glyph 1: GlyphID = 46, want 45
shaping_oracle_test.go:154: "Noto Sans Thai"/"ปั" glyph 1: HarfBuzz glyph id 46, table expects 45
```

So the chain is genuinely `hb-shape ↔ frozen table ↔ textshape`, with the table pinned from both
ends.

#### 4. Vacuity — checked every new guard; they fire

- **`shaped_fixture_test.go:716-718`** — the guard that fires when no `TJ` array carries a non-zero
  adjustment. I constructed exactly its trigger (forced `anyAdjustment = true` **and** zeroed every
  adjustment, so `TJ` is emitted with no numbers) and it fired verbatim:
  *"the content stream contains no TJ array with a NON-ZERO adjustment — every offset and kern
  computed to zero, so this fixture certifies nothing about positioning."* **Confirmed live.**
- **`TestShapedExpectationsObservability`** does not trust the table's `Observable` flag — it
  **measures** it against a reconstructed naive rune→`cmap` answer (`shaping_expectations_test.go:495-499`)
  and errors if a frozen flag disagrees with the measurement. That is the right shape.
- **`TestShapedTextFixtureObservabilityRedProof`** (`:841`) is a genuine *mirror*: it runs the same
  `rowIsObservable` machinery against `multi-script-fallback` and requires **0**. The pair is
  mutually constraining — a detector stuck on `false` reddens the first test, one stuck on `true`
  reddens the second. Neither can be satisfied by a broken detector.
- **`TestShapedTextFixtureCIDsOriginateInShapedRuns`** carries its own `countDiffered == 0` guard
  (`:519-525`), so the CID-count assertion cannot pass without discriminating.
- **`YOffset` is labelled honestly and is NOT credited with a proof.** Both sites say so in the
  assertion's own text: `shaping_expectations_test.go:32-40` ("**a guard credited with a red-proof it
  does not have is worse than one openly labelled unproven**… Do not manufacture a red-proof for it")
  and `:452-457` at the assertion itself. This is D-000.24 applied correctly and **is not a finding.**
  The related `textdoc.go:687-703` fail-closed branch is a *different* thing and is properly tested:
  `TestShapedRunFailsClosedOnYOffset` (`textdoc_test.go:130-145`) carries a 2×2 — the same run with
  `YOffset` zeroed must succeed — so it cannot pass for an unrelated reason.
- **`AC4`'s positive assertion is structural**, not a denylist: `GlyphForRune` no longer exists
  anywhere in source (only in a comment at `render.go:624`), so there is no rune→CID route to deny.

#### The two developer claims

**`matrix.yml` was one document short since Story 2.2 — CONFIRMED, and the new guard catches it.**
At `3373dac`, `.github/workflows/matrix.yml:166` reads `docs="minimal-rect font-text image-embed"`
while `matrix_test.go:615` declares **four** `matrixDocuments` including `multi-script-fallback`, and
no `hash.*.multi-script-fallback.txt` upload path exists on any target. The drift is exactly as
described. I red-proved the new `TestMatrixDocumentSlugsAreRegisteredInCI` in **both** halves:

| constructed drift | result |
|---|---|
| remove `shaped-text` from the workflow's `docs=` line | fires at `:63` printing both lists side by side |
| remove one `hash.linux-arm64.shaped-text.txt` upload path | fires at `:78` — *"that leg's hash never reaches the comparison job"* |

The guard is untagged (runs every story), fails closed if the slug regex finds nothing (`:44-46`),
and its `^\s*slug:` anchor correctly cannot match a commented-out line. The workflow's own comparison
job is also sound: `if-no-files-found: error` on every upload, explicit filenames rather than a glob,
a `count -ne 4` check and a hex-shape check per file.

**Program sizes — CONFIRMED exactly, and nothing was dropped.** I re-measured F4's own inputs:

```
Noto Sans      rune-derived: 24 out, 50832 B | shaped: 24 out, 50832 B | delta   +0
Noto Sans Thai rune-derived: 19 out,  1600 B | shaped: 13 out,  1240 B | delta -360
Noto Sans SC   rune-derived:  4 out,  1252 B | shaped:  4 out,  1252 B | delta   +0
```

F4's baseline reproduces to the byte, and the −360 (−22.5%) figure is right. The direction is
explained correctly: the shaped input asks for the glyphs actually drawn, whose GSUB closure is
smaller than the closure of the unshaped mark glyphs. **Nothing is silently dropped** — I asserted
that every shaped glyph id is present in `GlyphForSource` of the shaped subset, and it is; the
"not retained" path (`render.go:677-685`) is a located error naming face and glyph id, never a
`.notdef`.

#### AC-by-AC

AC1 ✅ (`internal/text/shape.go`; one `Shaper` per `Font`, ranges a slice, all `int16`) ·
AC2 ✅ · AC3 ✅ (all five fields, every glyph, every row) · AC4 ✅ (structural; `GlyphForRune` gone) ·
AC5 ✅ *substance* — the render-side "not retained" guard is live and red-proved, and `Subset`
neither sorts nor ranges a map — but **see Findings 8 and 9** for a dead second guard and a coverage
regression · AC6 ✅ · AC7 ✅ · **AC8 ⚠ — see Finding 1** (segmentation
is correct; segment *positioning* is not) · AC9 ✅ (legs written and registered, Docker not run, as
required) · AC10 ⚠ — mechanism ✅, human step openly outstanding (**not** counted as a finding, per
the story's own disclosure) but **see Finding 2** · AC11 ✅ · AC12 ✅ — **and its self-correction is
right**: I verified `ot/metrics.go` in `textshape@v0.0.15` — `Ascender`, `Descender`, `CapHeight`
and `BBox` all return `int16`; only `HorizontalAdvance` (`:403`) returns `float32`. **The count is
two, not three.** · AC13 ✅ · AC14 ✅

Other gates re-run independently: `go vet ./...` clean · `go vet -tags=matrix ./...` clean ·
`gofmt -l` empty for `folio-go` and `lint` · lint rule suite 4/4 packages pass under `GOPROXY=off` ·
`js/wasm`, `linux/amd64` and `linux/arm64` all build with `CGO_ENABLED=0`.

---

### Finding 1: Face-segment origins still use unshaped per-rune advances — a mixed-script element is drawn kerned but positioned unkerned

- **Severity**: **Blocker**
- **Category**: Correctness
- **Location**: `folio-go/render.go:400-408` (`splitByFace`), consuming `internal/fontset/fontset.go:307` (`AdvanceForRune`)
- **Related AC**: AC8, AC4, AC6 (and the completion notes' AC12 claim about `AdvanceForRune`)

**Observation.** `splitByFace` advances the cursor between face-segments by summing **per-rune
`hmtx` advances**:

```go
var advance1000 int64
for _, r := range seg.runes {
    a, ok := f.AdvanceForRune(r)      // hmtx advance: no GPOS, no ligature collapsing
    advance1000 += a
}
cursor += geom.ScaleRound(geom.Length(advance1000), int64(fontSize), 1000)
```

Every glyph *within* a run is now positioned from the shaper's answer — `appendShapedRun` goes to
real trouble to emit a trailing adjustment so the run's total advance is exactly the shaped advance
(`textdoc.go:681-684, 748-750`). But the origin of the **next** segment is still computed from the
pre-2.3 rune sum. Before this story the two agreed, because nothing was kerned. They no longer do.

I measured the divergence across the new fixture's own segments:

```
SEG 15  Noto Sans  "office fi AV Wo. To,"   cursor(naive)=8708  shaped=8578  DELTA=-130 (1000-em)
1 of 21 segments diverge
```

and then constructed the case the fixture cannot produce — a shape-observable segment **followed by
another face's segment**, at the fixture's own 16pt:

| element text | next segment's actual x | correct x from the shaped advance | error |
|---|---|---|---|
| `"AV ก"` | 23984 millipoints | 23344 | **+640 millipoints = 0.64 pt** |
| `"Wo. ก"` | 33008 | 32688 | **+320 millipoints = 0.32 pt** |
| `"office ก"` | 45680 | 45680 | 0 (this ligature's advance happens to equal its components') |
| `"Ada ก"` | 33200 | 33200 | 0 (no kern pair) |

**Impact.** A single element mixing scripts — which is precisely what Stories 2.2 and 2.3 exist to
serve, e.g. a Thai name inside a Latin sentence — draws its first segment kerned (narrower) and then
places the following segment as though it were not, leaving a visible gap. The error is
proportional to the kerning in the preceding segment and scales with font size.

**Why the fixture cannot see it, and why that matters.** `fixtures/shaped-text/`'s only
shape-observable Latin segment (`e4`, `"office fi AV Wo. To,"`) is the **last** segment of its
element, so its wrong cursor is never consumed. `e6` (`"Ada ปั 结"`) *is* the mixed-run case but
`"Ada "` carries no kern, so it happens to agree. The golden hash is therefore green over a live
defect. This is AC11's own hazard one level up: the fixture is shape-observable for *glyph choice*
and blind for *segment placement*.

**It also falsifies a completion-note claim.** The notes say `AdvanceForRune` is *"correct in its own
right… a segment-level cursor, not a per-glyph position."* The float half of that is fine; the
shaping half is now false — as a segment-level cursor it is measurably wrong wherever the preceding
segment kerns.

**Suggested resolution** (do not implement from this review): carry each segment's shaped total
advance out of `buildShapedPDFRuns` (it already computes every `XAdvance`) and use **that** for the
cursor, instead of re-deriving an unshaped sum in `splitByFace`; then add a fixture element in which
a kerning Latin segment is followed by a different face (`"AV ก"` reproduces it), so the golden can
see the property. If the lead rules this belongs to Story 2.5's layout work rather than here, that
must be an **explicit deferral with a `deferred-work.md` entry** — at present nothing in the repo
records it, and the story reports the seam as correct.

---

### Finding 2: Two committed artifacts state that a human looked at the rendered Thai; the completion notes state the sign-off is outstanding

- **Severity**: **Blocker**
- **Category**: AC Conformance / Maintainability
- **Location**: `fixtures/shaped-text/README.md:89-92` and `folio-go/shaped_fixture_test.go:660-661`
- **Related AC**: AC10 (D-000.22, D-000.24)

**Observation.** The fixture README carries a section headed **"### And a human looked"** whose first
sentence is:

> **A human read the rendered Thai before this hash was frozen**, exactly as Story 2.2 required a
> human to look at the rendered Chinese. What was checked and what was seen is recorded in the
> story's completion notes.

`shaped_fixture_test.go:660-661` repeats it in the golden test's own doc comment:

> …which is why a human looked at the rendered Thai before this hash was frozen (recorded in the
> story's completion notes and in this fixture's README).

The completion notes those two artifacts cite as their evidence say the opposite, in the developer's
own words: *"**DEVELOPER CHECK DONE, HUMAN SIGN-OFF OUTSTANDING**… I am the developer agent, not a
human, so AC10's human step is **NOT** satisfied by this note"* and *"**The fixture hash was frozen
before this sign-off exists.**"*

**To be explicit about scope: the missing human check is NOT the finding.** The developer discloses
it plainly and it is with the owner; that disclosure is exactly right. The finding is the **false
credit in the committed artifacts**, and it is the "other exposure" that freezing the hash early
created.

**Impact.** D-000.24 is on the point, verbatim: *"A guard credited with a red-proof it does not have
is worse than one openly labelled unproven, because the label tells the next reader **where to
look**. A false credit tells them the opposite."* The README is the artifact a future re-recorder
consults — its own closing section tells them the human check is "**not** optional extras to be
skipped" — and it will tell them the check is already done. D-000.22's warning applies with full
force: the moment this fixture is committed it becomes an **input**, and this claim becomes part of
it. If the outstanding sign-off never happens, **nothing in the repository will ever say so**, while
two files will say it did. Both citations are also self-referentially broken: they point at notes
that contradict them.

**Suggested resolution**: change both sites to state that a **developer-side** visual check was done
(the notes describe it well, with a naive-vs-shaped rasterised comparison, and it is genuinely
useful) and that the **human sign-off is outstanding**, naming it as the condition on calling the
story done. Update both to the affirmative form only once a human has actually confirmed. Nothing
else about the fixture needs to change.

---

### Finding 3: `epics.md`'s two amendment blocks cite the wrong rulings — one cites a ruling that explicitly refuses the amendment it is offered as authority for

- **Severity**: **Major**
- **Category**: Convention (D-000.6 / D-000.26)
- **Location**: `_bmad-output/planning-artifacts/epics.md`, the two block-quoted notes added under Story 2.3 (`:751`/`:753` block and `:756` block)
- **Related AC**: AC13, AC9

**Observation.** The block amending `epics.md:751` (GPOS mark-to-base) and `:753`
(`go-text/typesetting`) is headed:

> **AMENDED IN DEVELOPMENT, Story 2.3 (D-2.3.2, D-2.3.3).**

Neither citation is the governing ruling:
- **D-2.3.1** is the one that amends `:753` — its own header says so: *"resolves D-2.3-Q2; **amends
  `epics.md:753` under D-000.6**."* It is not cited.
- **D-2.3.2** is *"`/ToUnicode` allocates CIDs per (GID, Unicode-context)"* — unrelated to either
  clause.
- **D-2.3.3** is titled *"`epics.md:756` is weak, not false, and is **NOT** amended."* Citing it as
  the authority for an amendment inverts its content.

The very next block is headed **"CLARIFIED IN DEVELOPMENT, Story 2.3 (D-2.3.1)"** and applies to
`:756` — but D-2.3.1 is the HarfBuzz oracle ruling. The ruling that actually governs `:756` is
**D-2.3.3**. The two citations are effectively swapped.

**Impact.** `epics.md` is a governing artifact and D-000.6 requires an amendment to name its
authorising ruling. A reader following the cited D-2.3.3 from the `:751`/`:753` amendment lands on a
ruling whose entire content is a refusal to amend, and will reasonably conclude the amendment is
unlicensed. This is D-000.26's own class — the subject of a citation is checkable on the page, and
here it is wrong. (The *substance* is fine: `:756`'s clause text is untouched, only a clarifying note
was appended, which is what D-2.3.3 permits. It is the attribution that is broken.)

**Suggested resolution**: cite **D-2.3.1** (and **D-000.24** for the `YOffset`/mark-to-base half) on
the `:751`/`:753` block, and **D-2.3.3** on the `:756` block.

---

### Finding 4: `matrix.yml` names a test that does not exist, in the wrong file, and implies it runs only at the gate

- **Severity**: **Minor**
- **Category**: Maintainability
- **Location**: `.github/workflows/matrix.yml:179`

**Observation.** The comment introducing the `docs=` list reads:

> `TestMatrixDocumentSlugsAreInCI` (`folio-go/matrix_test.go`) now asserts the two agree.

`grep -rn TestMatrixDocumentSlugsAreInCI .` matches **only this comment**. The real test is
`TestMatrixDocumentSlugsAreRegisteredInCI`, in `folio-go/matrix_registration_test.go`.

**Impact.** Both halves mislead, and the file half is the worse one: `matrix_test.go` carries
`//go:build matrix`, so pointing a reader there implies the guard runs **only at the epic gate** —
the exact opposite of the property the story claims for it and the reason it was put in a separate
untagged file (*"running in every story is what makes it worth having"*). A maintainer hunting for
the guard that prevents this specific drift recurring will not find it.

**Suggested resolution**: correct the name and the filename in that comment.

---

### Finding 5: `readEmittedRuns` never scans a `TJ` array's trailing adjustment

- **Severity**: **Minor**
- **Category**: Tests
- **Location**: `folio-go/shaped_fixture_test.go:272-303`

**Observation.** The scan loop breaks as soon as no `<` remains in the remaining body, so the number
emitted **after the final `>` and before `] TJ`** is never examined. That number is real output:
`textdoc.go:748-750` emits `adjustments[len(run.Glyphs)]` there, and it is the term that makes a
run's total advance equal the shaped advance rather than the sum of `/W` widths.

Latent today, not live: the witness prints `9 of 21 emitted runs use TJ, 9 carry a non-zero
adjustment` — equal counts, so no current run has a trailing-only adjustment.

**Impact.** `NonZeroAdjust` under-reports, and it is consumed by two guards that matter: the
semantic acceptance step (`shaped_fixture_test.go:716`) and the matrix leg guard
`requireShapedTextIsShaped` (`matrix_test.go`). The failure direction is **fail-closed** — a
trailing-only run would make `adjustedRuns` under-count and could trip the `== 0` guard spuriously
— so this is not a vacuity. But it means both guards are blind to a class of adjustment the emitter
can produce, and it is the kind of parser gap that becomes a vacuity the moment someone "fixes" the
spurious failure by loosening the guard instead of the parser.

**Suggested resolution**: after the loop, scan the remaining `body` for integers as well.

---

### Finding 6: `supplementaryPlaneError` prints the offending character rather than its code point

- **Severity**: **Nit**
- **Category**: Maintainability
- **Location**: `folio-go/internal/pdf/textdoc.go:576-578`

**Observation.** `string(rune(e.r))` embeds the character itself in the message
(*"cannot yet express rune <char> above the Basic Multilingual Plane"*). A located error whose whole
subject is a rune the pipeline cannot encode should name it as `U+XXXX`; the character may well be
unrenderable in the terminal or log reading the error. Every other located error in this codebase
names its subject numerically.

**Suggested resolution**: format as `U+%04X`.

---

### Finding 7: `buildToUnicodeCMap` emits one unbounded `beginbfchar` section (pre-existing, forward note)

- **Severity**: **Nit**
- **Category**: Convention (spec conformance)
- **Location**: `folio-go/internal/pdf/textdoc.go:533-534`

**Observation.** The CMap is emitted as a single `N beginbfchar … endbfchar` block with
`N = len(face.ToUnicode)`. The `ToUnicode` CMap specification limits a `beginbfchar`/`endbfchar`
section to **100 entries**. Measured across all text fixtures today, every section is well under:

```
font-text              : 25
multi-script-fallback  : 4, 1, 1
shaped-text            : 14, 7, 28
```

**This is not a Story 2.3 regression** — the pre-2.3 derivation had the same unbounded shape, so it
is inherited, and I am recording it rather than charging it to this story. It is worth a forward
note because 2.3 is the story that both rewrote this function's entry source and produced the largest
section to date (28), and because Story 2.4's larger Thai corpora will push it further. A document
whose face needs >100 distinct CIDs would emit a section a strict validator rejects.

**Suggested resolution**: chunk into sections of at most 100 at whatever point the Epic 2 gate judges
appropriate; not this story's to fix.

---

### Finding 8: `fontset.Subset`'s "not retained" guard is unreachable dead code, and `Subset` silently accepts glyph ids the face does not have

- **Severity**: **Minor**
- **Category**: Tests / Correctness
- **Location**: `folio-go/internal/fontset/fontset.go:493-500`
- **Related AC**: AC5

**Observation.** AC5 requires that *"a shaped glyph id the plan did not retain is a located error naming
the face and the glyph id, never a silent substitution and never `.notdef`."* Two guards implement it.

**The render-side one is live and I red-proved it.** Dropping the `ffi` ligature from the subset
union (`render.go:466-469`) produced exactly the message AC5's red-proof predicts:

```
folio: Render: face "Noto Sans": shaped glyph id 1656 was not retained by the subset plan
```

**The `fontset` one cannot fire.** `plan.MapGlyph` reads `p.glyphMap` (`subset/plan.go:357-360`),
and `createCompactMapping` (`subset/plan.go:334-347`) inserts **every** member of `glyphSet` — which
is the input set plus its closure. So a glyph passed to `AddGlyphs` is always mapped. Coverage
confirms it: block `fontset.go:495.10,500.4` has count **0**, and `grep -rn "was not retained"
--include='*_test.go'` matches nothing.

**The corollary is the more interesting half, and it is new to this story.** Because `Subset` is now
**glyph-keyed**, a caller can hand it an id the face does not have — and the vendor fabricates one
rather than reporting it absent. Measured against `fonts/notosans/NotoSans-Regular.ttf` (≈4,000
glyphs):

```
input [36 37]             -> NumGlyphs=6, program 49340 B, unmapped inputs: none
input [36 65000]          -> NumGlyphs=5, program 49100 B, unmapped inputs: none
input [65535]             -> NumGlyphs=2, program   460 B, unmapped inputs: none
input [36 37 60000 65535] -> NumGlyphs=8, program 49356 B, unmapped inputs: none
```

Glyph **65535** does not exist in that face, and `Subset` returns a mapping for it with no error.

**Impact.** The guard reads as coverage in the diff and cannot fire; a maintainer trusting it will
believe the vendor boundary is checked here when it is not. And the direction that *is* reachable —
a bogus glyph id becoming a fabricated glyph in the embedded program — has no guard at all. This is
D-000.25's Finding 2 class ("vendor accessors substitute plausible defaults for missing data")
arriving at a call site **this story created**. AC5's substance is nonetheless satisfied by the
render-side guard, which is why this is Minor rather than higher.

**Suggested resolution**: either label `fontset.go:495-500` explicitly as an unreachable
vendor-contract assertion (the honest form, per D-000.24), or make it a real guard by validating
each source id against the face's glyph count **before** `AddGlyphs`. Flag the second half for
`2-3a-audit-the-vendor-boundary`, which already owns this class.

---

### Finding 9: `TestPermutationInvariance` cannot detect a folio-side sort, and the dedup branch lost its only coverage

- **Severity**: **Minor**
- **Category**: Tests
- **Location**: `folio-go/internal/fontset/fontset_test.go:142-150` (`shapedGlyphIDs`), `:242-259`
- **Related AC**: AC5 (D-1.5.7 / AC8a)

**Observation, part one.** The re-key is correct — the test now permutes **glyph ids**, and it proves
set-equality before comparing (`:253-259`), so it cannot compare different sets. And `Subset` genuinely
does not sort: `fontset.go:436-449` dedups through a map read only by index and appends in
first-occurrence caller order.

But the invariance the test observes is **produced by the vendor, two frames below the call**:
`Plan.createCompactMapping` sorts the glyph set itself (`subset/plan.go:341`,
`sort.Slice(gids, …)`). So a defensive sort added inside `fontset.Subset` would **not** redden this
test — the output is identical either way. The test is a vendor forward-guard, not a proof that
folio does not sort. AC5's clause *"preserving AC8a's discriminating power"* is therefore overstated:
the discriminating power against a folio-side sort was never there. (This property is inherited, not
introduced — but the story's explicit claim about it is what makes it worth recording.)

**Observation, part two — a real coverage regression this story did introduce.** `shapedGlyphIDs`
**deduplicates before calling `Subset`**, so no test in the package now passes a slice containing
duplicates. Coverage confirms: `fontset.go:440.14,441.12` (the dedup `continue`) has count **0**. The
pre-change rune-based tests did exercise it — `"Hello, World! 0123456789"` repeats `l`, `o` and space.
The production path *does* pass duplicates (`render.go:466-468` appends **every** shaped glyph id),
so invariance-under-multiplicity is now a live property with no unit coverage.

**Suggested resolution**: add a `fontset` test that passes a duplicate-bearing slice and asserts the
program is byte-identical to the deduplicated one; and either soften AC5's "discriminating power"
wording or add a source-scanning guard forbidding `sort`/`slices.Sort` inside `Subset`, which is the
only mechanism that could actually catch its reintroduction.

---

### Finding 10: `buildTextContentStream`'s doc comment still asserts the invariant this story broke

- **Severity**: **Minor**
- **Category**: Maintainability
- **Location**: `folio-go/internal/pdf/textdoc.go:598`

**Observation.** The comment reads: *"…show its text as a hex string of 2-byte CIDs (Identity-H:
**CID == subset glyph id**)."* That is precisely the assumption `ExtraCIDs` now falsifies, and it sits
in the doc comment of the function that emits the CIDs — three lines from `glyphForCID`, which exists
because it is no longer true.

**Impact.** D-2.3.2 closes with *"Anything that assumes CID↔GID is 1:1, or that inverts the map, must
be checked."* The code was checked; this comment was not. The same file quotes D-1.1.b's warning that
*"this must be said in a code comment or a later agent will 'unify' it"* — a comment asserting the
**false** version of the invariant is the same hazard pointed the other way.

**Suggested resolution**: restate it as "CID == subset glyph id for the base block; extras resolve
through `glyphForCID`", and cross-check `EmbeddedFace.NumGlyphs`'s doc (`textdoc.go:20`), which
carries the same overloading ("subset glyph count; also the size of the BASE CID block") without
anything asserting the two stay equal.

---

### Finding 11: The story's headline Thai assertion compares a subset glyph id against an emitted CID

- **Severity**: **Minor**
- **Category**: Tests
- **Location**: `folio-go/shaped_fixture_test.go:619-632`
- **Related AC**: AC10 (semantic property 3)

**Observation.** `TestShapedTextFixtureThaiDrawsTheLoweredMarkForm` — the PDF-level statement of the
defect this story exists to fix — does:

```go
loweredCID, ok := sub.GlyphForSource[loweredMarkGlyph]   // this is a SUBSET GLYPH ID
...
if run.CIDs[1] != loweredCID { ... }                     // compared against an emitted CID
```

The variable is *named* `loweredCID` but holds a subset glyph id. The comparison is only valid while
that glyph's base CID is unclaimed — i.e. while nothing gives glyph 46 a second `(glyph, cluster
text)` pair. Today that holds (glyph 46 only ever occurs as a cluster tail with text `""`), which is
why it passes.

**Impact.** Two directions, both bad. If glyph 46 ever needs an extra CID, the test reports *"the
`ปั` run draws the UNLOWERED mark form"* — a **false red on the story's headline defect**, pointing
the reader at a typography regression that did not happen. And in the other direction it silently
stops testing what its name says: it would be checking CID identity rather than glyph identity. This
is the one assertion in the story whose message a future reader is most likely to trust without
re-deriving.

**Suggested resolution**: resolve `run.CIDs[1]` back through the CID→glyph indirection (the same
mapping `pdf.glyphForCID` performs) before comparing, or assert via the face's `/ToUnicode` and
`/CIDToGIDMap` as the surrounding tests do. Rename the variable either way.

---

### Finding 12: Two fail-open spots on the new CID path, in a file whose posture is fail-closed everywhere else

- **Severity**: **Minor**
- **Category**: Correctness
- **Location**: `folio-go/render.go:700` and `folio-go/internal/pdf/textdoc.go:365`

**Observation.**

1. `render.go:700` allocates an extra CID as `cid = uint16(sub.NumGlyphs + len(state.extras))` with no
   bound check. Identity-H's CID is two bytes, so past 65535 this **wraps silently and collides with
   the base block**, producing both a wrong glyph and a wrong `/ToUnicode` entry. Far off in practice
   — it needs a subset near the 65535-glyph ceiling plus context-distinct CIDs — but it is a silent
   wrap where every other limit in this codebase is a located error.
2. `textdoc.go:365` emits the `/W` base block as `b.writeInt(face.WidthForGlyph[uint16(cid)])`
   **without** the `, ok` presence check that `appendShapedRun` uses ten lines earlier
   (`textdoc.go:704-707`, which returns `missingGlyphError`). A missing width would be written as
   `0` — a zero-width glyph — rather than failing. Complete today, because `WidthForGlyph` is dense
   over `0..NumGlyphs-1` (`fontset.go:505-514`); fragile by shape.

**Impact.** Neither is reachable today. Both are the exact "healthy output and broken output are the
same bytes" shape that `appendShapedRun`'s own `YOffset` branch is written to avoid, sitting on the
CID path this story introduced.

**Suggested resolution**: make the CID-space exhaustion a located error naming the face and the CID
count; add the `, ok` check to the `/W` loop and reuse `missingGlyphError`.

---

### Not findings (checked and cleared, recorded so they are not re-raised)

- The two inherited Story 2.1 failures — present, unchanged, correctly excluded from the count.
- The two pre-existing `forvar` lints at `internal/template/serialize.go:108` and `fixture_test.go:190`.
- The two vendor-`float32` sites at `internal/fontset/fontset.go:316` and `:511` — out of scope,
  owned by `2-3a-audit-the-vendor-boundary` (D-000.25), correctly sequenced in `sprint-status.yaml`
  before 2.4. The count of **two** is verified correct, and AC12's own text naming `Metrics` as a
  third site is properly corrected in the completion notes.
- AC10's outstanding human visual check as a *concealed* gap — it is disclosed, not concealed.
  (Finding 2 is about the contradictory claims in the committed artifacts, which is a different
  thing.)
- The `YOffset` forward guard — correctly labelled at both sites, not credited with a proof it
  lacks. D-000.24 applied exactly as intended.
- The new fixture's `expected.pdf` / `expected.json` two-halves check, `/BaseFont` shape assertion,
  `fvar`/`gvar` absence, page count and MediaBox — all present and all assert the object exists
  before asserting about it.

---

### Verification hygiene

Every mutation I made and reverted, with the SHA-256 confirming restoration:

| file | mutations | restored to |
|---|---|---|
| `folio-go/internal/pdf/textdoc.go` | 4 (adjustment +1; drop correction; drop small kern only; force zero-adjustment TJ) | `295353c4…f762f5c3` ✅ |
| `folio-go/render.go` | 3 (per-GID cidKey; collapse extra CID onto base; drop the `ffi` glyph from the subset union) | `5861689c…2945d6d` ✅ |
| `.github/workflows/matrix.yml` | 2 (drop a doc from `docs=`; drop one upload path) | `a58c86f9…8ab7bbc` ✅ |
| `folio-go/shaping_expectations_test.go` | 1 (`ปั` row 46→45) | `58632bb7…3bd7e468` ✅ |
| scratch `_test.go` × 3 | created for measurement (program sizes; segment-cursor divergence; out-of-range glyph-id reachability) | deleted ✅ |

Suite re-run after the final revert: **369 pass, 2 fail** — the same two inherited failures, same
messages. `go vet`, `gofmt -l` and the lint suite re-confirmed clean.

Final `git status --porcelain` diffed against the list captured before the first command: **identical**.

---

## Finding Resolutions

*(Story 2.3 finisher. Every finding in § QA Results triaged; every fix independently reproduced
before AND after, with the subject of each measurement named per D-000.26.)*

**Triage: 11 FIX · 0 DISMISS · 1 DEFER**, over 2 Blockers, 1 Major, 7 Minors and 2 Nits.

The review was strong and its four named attacks all cleared under construction. Every finding was
reproduced at this HEAD before being acted on; two of them turned out to carry a factual detail worth
correcting, recorded inline below.

| # | Severity | Decision | One line |
|---|---|---|---|
| 1 | Blocker | **FIX** | Segment cursor now derives from the shaped advances; a fixture element and a literal-valued guard added, both red-proved. |
| 2 | Blocker | **FIX** | Both false-credit sites rewritten to "pending"; the obligation is now a `matrix`-gated FAILING test bound to the fixture digest. |
| 3 | Major | **FIX** | Both `epics.md` blocks re-cited (D-000.24 / D-2.3.1 and D-2.3.3), each with the correction stated in place. |
| 4 | Minor | **FIX** | `matrix.yml` comment corrected to the real test name and file, plus why that file is untagged. |
| 5 | Minor | **FIX** | Parser now scans the trailing `TJ` term; a cross-check against an independent recount guards it. |
| 6 | Nit | **FIX** | `U+XXXX`, hand-rolled — `fmt` is banned in `internal/pdf` by D-1.1.b's guard. |
| 7 | Nit | **DEFER** | Pre-existing, unreachable today; **DW-14** in `deferred-work.md`. |
| 8 | Minor | **FIX** | Out-of-range glyph ids now rejected before the vendor can fabricate one; the dead guard is labelled, not counted. |
| 9 | Minor | **FIX** | Source-scanning sort guard added (the only mechanism that can see it) + a duplicate-bearing test. |
| 10 | Minor | **FIX** | Both stale `CID == subset glyph id` comments restated, including `NumGlyphs`' overloading. |
| 11 | Minor | **FIX** | The emitted CID is now resolved through the document's own `/CIDToGIDMap` before comparison. |
| 12 | Minor | **FIX** | CID-space exhaustion is a located error; the `/W` loops take the `, ok` check. |

---

### Finding 1 (Blocker) — FIX. Face-segment origins used unshaped advances

**Reproduced first.** Against the shipped chain `["Noto Sans","Noto Sans Thai","Noto Sans SC"]` at
16 pt, before any change:

```
"AV ก"                   seg 1 actual_x=23984  correct_x=23344  error=+640 millipoints
"Wo. ก"                  seg 1 actual_x=33008  correct_x=32688  error=+320
"office fi AV Wo. To, ก" seg 1 actual_x=143488 correct_x=141408 error=+2080
"office ก" / "Ada ก"     error=0 (no kern pair in the leading segment)
COVERAGE WITNESS: 10 of 10 segments evaluated across 5 inputs
```

Exactly the review's numbers, on the review's own inputs.

**The fix is structural, not arithmetic.** Rather than teaching `splitByFace` to sum different
numbers, shaping was **moved into it**: each face-segment is shaped once, the shaped glyphs are
carried on the `textRunSource` the renderer later draws, and the cursor is the sum of those same
glyphs' advances. `renderDocument` no longer re-shapes. There is now **one** shaping answer per
segment, so the drawn run and the next segment's origin cannot disagree again — the second derivation
that drifted no longer exists. Per-glyph advances are scaled to the 1000-em individually and then
summed, never summed then scaled, because that is the already-rounded space the viewer's pen consumes
(the same reasoning the run's own advance-correction term rests on).

**`fontset.AdvanceForRune` was deliberately RETAINED, not deleted.** It has no caller on the render
path now, and deleting it was tempting — but it is one of the two `ot.Face.HorizontalAdvance`
(`float32`) sites that `2-3a-audit-the-vendor-boundary` (D-000.25) owns, which this story is fenced
out of. Removing it would have silently shrunk that story's subject from two sites to one. Its doc
comment, which claimed it positioned second-and-later face-segments, is rewritten to say plainly that
it does not and why.

**Why no fixture caught it — the more useful half of the finding, recorded as asked.**
`fixtures/shaped-text/`'s only shape-observable Latin segment (`e4`, `"office fi AV Wo. To,"`) is the
**last segment of its element**, so its wrong cursor was never consumed by anything. `e6`
(`"Ada ปั 结"`) *is* the mixed-run case, but `"Ada "` carries no kern pair, so the unkerned and shaped
cursors happened to agree. **The golden was green by accident of segment ordering, over a live
defect.** That is worse than an absent fixture, because it reads as coverage in every report: the
next change to element ordering would have surfaced this as a mystery hash movement months later,
with nothing in the repository connecting it to kerning. It is AC11's own hazard one level up — the
fixture was shape-observable for *glyph choice* and blind for *segment placement*, and nothing said
so.

**Closed in both directions.** A new fixture element `e7` (`"AV ก"`) puts a kerning Latin segment in
front of another face's segment, so a regression is now a different hash. And
`TestFaceSegmentOriginsUseShapedAdvances` states the property directly in **hand-derived literals**
rather than by recomputing production's own arithmetic (which would agree with any arithmetic
production chose, including the wrong one). Its expected values come from `hb-shape (HarfBuzz)
14.2.0` against `folio-go/fonts/notosans/NotoSans-Regular.ttf` (unitsPerEm **1000**, so a font-unit
advance *is* a 1000-em advance):

```
"AV "   ax 599+600+260      = 1459 -> 23344 milli at 16pt   (hmtx 639: A/V kerns -40)
"Wo. "  ax 910+605+268+260  = 2043 -> 32688                 (hmtx 930: W/o kerns -20)
"Ada "  ax 639+615+561+260  = 2075 -> 33200                 NEGATIVE CONTROL, no kern
```

**RED-PROOF (live).** Reverting the cursor to the `AdvanceForRune` sum in the working-tree file:

```
--- FAIL: TestFaceSegmentOriginsUseShapedAdvances/AV_ก
    "AV ก": segment 2 sits 23984 millipoints after segment 1, want 23344
      this is EXACTLY the unkerned `hmtx` sum (23984 millipoints): the segment cursor has
      regressed to fontset.AdvanceForRune, so "AV " is drawn kerned and placed unkerned
--- FAIL: TestFaceSegmentOriginsUseShapedAdvances/Wo._ก   (33008, want 32688)
--- PASS: TestFaceSegmentOriginsUseShapedAdvances/Ada_ก   <- control stays green
```

The negative control staying green is the half that matters: a test that merely rejected the old
numbers would have had to fail here too. Restored from a `cp` backup and verified with `/usr/bin/diff`.

**Consequences.** `fixtures/shaped-text/` is re-recorded (`e7` added), from
`0aeb930a…07b52a` to `5964aad0…92e00f`, 91,059 bytes. The other four fixtures are **byte-identical to
baseline** — `git diff --name-only fixtures/` lists only `shaped-text` and the `font-text` files the
developer had already re-recorded under D-2.3.4. `font-text` is single-face per element, so it has no
cross-face cursor to move; `minimal-rect`, `image-embed` and `multi-script-fallback` are unaffected
for the reasons AC14 states.

---

### Finding 2 (Blocker) — FIX. False credit in committed artifacts, replaced by a red gate

**D-2.3.5's three binding conditions, all implemented.**

**(1) The pending obligation is a FAILING TEST, not a log entry.**
`TestShapedTextThaiSemanticSignOffIsRecorded` lives in a **`//go:build matrix`** file, gated exactly
as the cross-target matrix legs are. The routine per-story `go test ./...` does not compile it, so
2.3 commits green; the Epic 2 boundary gate is the run that carries `-tags=matrix`, and it **cannot
pass** until `fixtures/shaped-text/thai-signoff.json` exists. `go vet -tags=matrix ./...` compiles it
every story, so the deferral cannot rot unnoticed.

**(2) The sign-off names the hash it signed off.** The record's `sha256` must equal
`expected.json`'s. A re-record moves that digest and **automatically invalidates the sign-off**,
demanding a fresh look rather than silently carrying a human's word forward onto bytes nobody saw.

**(3) The artifacts say "pending", never anything stronger.** Both sites rewritten. The fixture
README's section is now headed **"The human step: PENDING"** and opens *"A human has NOT yet read the
rendered Thai. This hash was frozen with that sign-off outstanding."* The golden test's doc comment
says the same, in the negative, and points at the gate test by name. Both record that the **false
credit was there** and what it said, because deleting the claim silently would leave the next reader
no way to know the artifact had ever misreported itself.

**RED-PROOF, all four legs (live, `-tags=matrix`).**

| leg | state | result |
|---|---|---|
| record **absent** — the committed state | no `thai-signoff.json` | **FAIL**, with the resolution recipe in the message |
| record present, digest **matches** | `sha256` = `5964aad0…92e00f` | **PASS** |
| record present, digest **stale** | `sha256` = the pre-`e7` `0aeb930a…07b52a` | **FAIL** — *"the fixture has been RE-RECORDED since it was signed off"* |
| record present, `reader` **empty** | `"reader": ""` | **FAIL** — *"indistinguishable from no sign-off at all"* |

The placeholder was removed afterwards; the gate is **red as committed**, and the untagged run
reports `[no tests to run]` for that name. Both properties confirmed after removal.

**The boundary that keeps the rule's teeth**, recorded in `fixtures/shaped-text/README.md` under its
own heading so it is never re-derived:

> **D-000.22's machine-checkable half is never deferrable. Only its irreducibly-human half is, and
> only against a mechanical blocker.**

Weight class, name records, `TJ` adjustment values, axis pins, byte-identity across targets — all
asserted **at recording**, no deferral available, and all seven of this fixture's semantic properties
are asserted today. Only *"does this Thai read correctly to someone who reads Thai"* may be pending,
because nothing substitutes for it. **The evidence for that split is ours: the Thin-Chinese defect was
caught by a machine-checkable property, not by a human look.**

---

### Finding 3 (Major) — FIX. Swapped ruling citations in `epics.md`

Verified against the decision log before editing. `D-2.3.1`'s own header reads *"amends `epics.md:753`
under D-000.6"*; `D-2.3.2` is the `/ToUnicode` CID-allocation ruling and governs neither clause;
`D-2.3.3` is titled *"`epics.md:756` is weak, not false, and is **NOT** amended"*. `D-000.24` is the
ruling that names the `epics.md:751` / `YOffset` case explicitly and permits the forward-guard label.

- The `:751`/`:753` block now cites **clause 1: D-000.24 · clause 2: D-2.3.1**.
- The `:756` block now cites **D-2.3.3**, and states what that ruling actually licenses: a clarifying
  note *beside* the clause, **not** a D-000.6 amendment, because `:756` is weak rather than false.

Both corrections are stated in place rather than made silently, so a reader who remembers the old
citation can see what changed and why.

**Worth its own line, as asked:** a citation that reads as support and **inverts on inspection** is
the citation-integrity analogue of Blocker 2. Both point a reader at evidence that says the opposite
of what they were told it says; one did it with a fixture README, the other with a ruling id. D-000.26
is the general form — the subject of a citation is checkable on the page, so check it.

---

### Finding 4 (Minor) — FIX

Corrected to `TestMatrixDocumentSlugsAreRegisteredInCI` in `folio-go/matrix_registration_test.go`,
and the comment now says **why** that file carries no `//go:build matrix` line: the guard runs in
every story's plain `go test ./...`, not only at the epic gate, and running every story is what makes
it worth having. Pointing a reader at `matrix_test.go` implied the opposite. `actionlint` clean.

### Finding 5 (Minor) — FIX

`readEmittedRuns` now scans the tail after the final `>`. Measured on the re-recorded fixture:
**5 of 23 emitted runs carry a trailing adjustment** the old parser could never reach. It was latent
rather than live (no run is trailing-*only*, so `NonZeroAdjust` did not actually under-report today),
which is exactly why it needed a permanent guard rather than a one-off correction.
`TestReadEmittedRunsScansEveryTJAdjustment` re-derives every bare integer in every `TJ` array
straight off the content-stream bytes and requires the parser to have seen the same multiset, with a
vacuity leg that fails if the fixture ever stops containing a trailing case.
**RED-PROOF:** removing the tail scan gives *"readEmittedRuns saw 18 TJ adjustments but the content
stream carries 23"* — exactly the five trailing terms. Restored and re-verified.

### Finding 6 (Nit) — FIX, but not the suggested way

Formatted as `U+XXXX`. **`fmt.Sprintf` could not be used**: `internal/pdf`'s own
`scanNumericFormatting` guard (D-1.1.b) forbids every `fmt` formatting call and the
`strconv` `Format*`/`Itoa`/`Append*` family anywhere in the package except `numbers.go`, with no
diagnostic exemption — D-1.3.2 kept that no-exemption choice deliberately. Reaching for `fmt` would
have traded a nit for a guard violation. Done with byte arithmetic, on the same basis as the adjacent
`appendHex4`, and checked against seven codepoints spanning `U+0000`, `U+FFFF`, `U+10000` and
`U+10FFFF`. The package's own guard suite passes.

### Finding 7 (Nit) — **DEFER**, as **DW-14**

The only DEFER, and the reviewer proposed it as one. Genuinely pre-existing — the pre-2.3 derivation
had the same unbounded shape — and not reachable by anything folio ships. Re-measured at this commit
rather than carried forward from the review: `font-text` 25; `multi-script-fallback` 4, 1, 1;
`shaped-text` 14, 7, 28 (unchanged by `e7`, whose glyphs were all already in their faces' CID sets).
Deferred rather than dismissed because 2.3 both rewrote this function's entry source and produced the
largest section to date, and 2.4's larger Thai corpora push the same way — so the trend is this
story's even though the fix is not. `deferred-work.md` § DW-14 records the fix, the reason it wants a
deliberate re-record, and the wrong fix to avoid.

### Finding 8 (Minor) — FIX. Both halves

Reproduced first, subject named: `folio-go/fonts/notosans/NotoSans-Regular.ttf`, **4,515 glyphs**.
`Subset([]uint16{65535})` returned a 460-byte program and a complete mapping **with no error** — the
vendor fabricates rather than reports. `Subset` now validates every source id against the face's own
glyph count *before* `AddGlyphs`. **RED-PROOF:** removing the check makes all four out-of-range inputs
succeed again, each reported by name. A negative leg asserts the face's **highest valid** id is still
accepted, so the guard cannot be satisfied by an off-by-one that rejects a legitimate glyph.

The unreachable `MapGlyph` guard is **labelled** rather than deleted or credited (D-000.24): the
comment now states that `createCompactMapping` inserts every member of the glyph set so the branch
cannot fire, that its coverage count is 0, and that AC5's *live* guard is the render-side one. Kept
because it is the assertion that would catch the vendor changing that contract.

### Finding 9 (Minor) — FIX, and part one is now demonstrated rather than argued

**Part one.** `TestSubsetDoesNotSortItsInput` scans `Subset`'s own AST for `sort`/`slices` sort calls,
resolving import aliases by path, with a vacuity guard that fails if the method is ever renamed (this
project's recurring hard-coded-name trap). **RED-PROOF, and it proves the finding as well as the
fix:** injecting `slices.Sort(gids)` into `Subset` reddens the new guard by file and line **while
`TestPermutationInvariance` stays PASS** in the same run — the vendor's own
`createCompactMapping` sorts below the call, so the output is identical either way. AC5's claim to
have preserved AC8a's discriminating power against a folio-side sort was overstated, exactly as the
review said, and that is now visible in one test run rather than by reading vendor source.

**Part two — correction to the finding's framing, made after measuring.**
`TestSubsetIsInvariantUnderDuplicateGlyphIDs` adds the duplicate-bearing input the package lost. But
its **red-proof came back GREEN**: disabling folio's own dedup does not change the output. Measured
directly against the vendor (same face, 10 shaped ids expanded to 40): `AddGlyphs` with duplicates
yields `NumOutputGlyphs` 22 and a 7,672-byte program, **byte-identical** to the deduplicated call.
The vendor deduplicates too. So invariance under multiplicity is enforced **twice**, the new test
asserts the property the render path actually relies on, and it is labelled a forward guard on the
vendor contract rather than credited with a red-proof of folio's branch. Reported this way round
because a non-reddening mutation is a result, not a failure to report.

### Finding 10 (Minor) — FIX, both sites

`buildTextContentStream`'s comment now says `CID == subset glyph id` **for the base block only**, with
extras resolving through `glyphForCID`. The cross-check the review asked for found the same
overloading in `EmbeddedFace.NumGlyphs`' doc, which is also corrected: the two readings coincide by
construction, not by law, and the total CID count is `NumGlyphs + len(ExtraCIDs)`. A comment asserting
the **false** version of an invariant is D-1.1.b's own hazard pointed the other way, which is why both
were worth the words.

### Finding 11 (Minor) — FIX

`TestShapedTextFixtureThaiDrawsTheLoweredMarkForm` now resolves the emitted CID through the
document's **own `/CIDToGIDMap`** before comparing it to a subset glyph id, and the variables are
renamed to say which space they are in. The mapping is read off the produced PDF — the artifact that
carries the property and the one a viewer obeys — via a helper factored out of the existing
`/ToUnicode` reader so both walk the same path to the same object. It is keyed by the run's own
resource name rather than a hard-coded face name.
**RED-PROOF:** perturbing `buildCIDToGIDMap` so the base block is no longer the identity reddens it by
name (*"mark CID 8 resolves through /CIDToGIDMap to subset glyph 7, want 8"*). Under the old
comparison that mutation was invisible.

### Finding 12 (Minor) — FIX, both spots

CID-space exhaustion past Identity-H's two-byte ceiling is now a located error naming the face, the
subset size and the extra count, instead of a silent `uint16` wrap into the base block. The `/W`
emitter takes the `, ok` presence check `appendShapedRun` already takes — in **both** loops, base and
extras; the review named the base one. Neither is reachable today (`WidthForGlyph` is dense over
`0..NumGlyphs-1`), and both are written so the file has one posture rather than two: a missing width
read off a Go map is `0`, and a zero-width glyph is output that looks healthy.

---

### Not acted on, by instruction

- The two vendor-`float32` sites at `internal/fontset/fontset.go` — `2-3a-audit-the-vendor-boundary`'s
  (D-000.25). **Still two**, deliberately: see Finding 1 on why `AdvanceForRune` was retained.
- The two pre-existing `forvar` lints at `internal/template/serialize.go:108` and `fixture_test.go:190`.
- Story 2.1's two inherited failures — present, unchanged, same messages.

---

## Delivery Log

**Finisher:** Story 2.3 close-out. **Baseline:** HEAD `3373dac`, working tree carrying the
developer's uncommitted work. **Branch:** `main`.

**Suite — measured at the end, on the committed tree:**

| gate | result |
|---|---|
| `go test ./... -count=1 -v` (`folio-go`) | **377 pass, 2 fail** |
| the 2 failures | `TestCorpusMeetsP6ExerciseFloors`, `TestP2IndependentDPCrossCheck` — **Story 2.1's inherited, deliberate shortfalls (D-2.1.14)**, unchanged messages. **No third failure.** |
| delta from the review's count | 369 → 377 pass; **+8**, all added by this finisher's guards |
| `go vet ./...` | clean |
| `go vet -tags=matrix ./...` | clean — the matrix legs and the new sign-off gate compile |
| `gofmt -l` (`folio-go`, `lint`) | empty |
| lint rule suite (`GOPROXY=off go test ./...`) | 4 packages, all pass — incl. `TestAbsencesProductionScan`, `TestNoCompressorProductionScan`, `TestEmbedFontProductionScan` |
| `go list -m all` | **exactly two modules**; `TestModuleGraphAllowlist` untouched and passing |
| `CGO_ENABLED=0` builds | `js/wasm`, `linux/amd64`, `linux/arm64`, `darwin/arm64` — all four OK |
| `actionlint .github/workflows/matrix.yml` | clean |

**Guards added by this finisher (the +8):** `TestFaceSegmentOriginsUseShapedAdvances` (3 subtests),
`TestReadEmittedRunsScansEveryTJAdjustment`, `TestSubsetRejectsGlyphIDsTheFaceDoesNotHave`,
`TestSubsetIsInvariantUnderDuplicateGlyphIDs`, `TestSubsetDoesNotSortItsInput`, and — compiled only
under `-tags=matrix`, so outside this count — `TestShapedTextThaiSemanticSignOffIsRecorded`.

**Fixtures.** `fixtures/shaped-text/` **re-recorded** (element `e7`, `"AV ก"`, added so the golden can
see segment placement): `0aeb930a0cc29e7adff2ea757d61f87c2eea033d6dca3dbd6382ef154207b52a` →
`5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f`, 91,059 bytes, `go1.26.0`.
`input.folio` and `folio-go`'s `shapedTextTemplateJSON` re-verified **byte-identical** (1,654 bytes
each). `fixtures/minimal-rect/`, `fixtures/image-embed/` and `fixtures/multi-script-fallback/` are
**unmodified against baseline** — `git diff --name-only fixtures/` names none of them.
`fixtures/font-text/` carries only the developer's D-2.3.4 re-record; this finisher did not move it.

**Cross-target hash matrix — NOT RUN, per the per-epic cadence (D-000.4).** 2.3 is not one of
D-000.4's per-story overrides (those are 1.2, 1.5, 1.8, 2.4 and 4.7). All five documents' legs
including `shaped-text` are **written and registered and unrun**; `go vet -tags=matrix ./...` proves
they compile. **No Docker leg was executed.** They belong to the **Epic 2 boundary gate**.

**Owed at the Epic 2 gate, and mechanically enforced:**

1. **The four-target matrix legs** above — written but unrun.
2. **The Thai semantic sign-off — PENDING.** `fixtures/shaped-text/thai-signoff.json` does not exist,
   and `TestShapedTextThaiSemanticSignOffIsRecorded` is **RED under `-tags=matrix` as committed**.
   This is deliberate and is D-2.3.5's first binding condition: the story commits green, the gate
   cannot pass. The record must name the reader, the date, what they examined, and the digest
   `5964aad0…92e00f`.

**Deferred work opened:** **DW-14** (`/ToUnicode`'s unbounded `beginbfchar` section) — the only DEFER.

**Documents amended.** `_bmad-output/planning-artifacts/epics.md`: the two Story 2.3 blocks' ruling
citations corrected in place, each with the correction stated rather than made silently (Finding 3).
`fixtures/shaped-text/README.md`: the false human-check credit replaced by a PENDING section, the
D-2.3.5 boundary recorded verbatim, and `e7` documented. `.github/workflows/matrix.yml`: the guard's
name, file and untagged status corrected in the `docs=` comment.

**Not amended, and why:** `folio-mvp-decision-log.md` is append-only by its own header — the finisher's
follow-ups are appended as their own sibling entry, never as edits to D-2.3.1–D-2.3.5.

**Measurement hygiene.** Every mutation was applied to the working-tree file, run, and restored from a
`cp` backup verified with **`/usr/bin/diff`** (a bare `diff` has been observed lying in this
environment). All scratch `_test.go` files were deleted and their absence re-verified by absolute
path. Every comparison asserts both operands exist and are non-empty before comparing, and prints an
`N of N` coverage witness.
