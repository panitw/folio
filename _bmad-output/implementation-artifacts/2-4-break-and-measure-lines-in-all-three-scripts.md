# Story 2.4: Break and measure lines in all three scripts

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-4-break-and-measure-lines-in-all-three-scripts`
**Status:** `done`
**Covers:** NFR3 · AD-2, AD-25 · S4 (`epics.md:815`–`:842`)
**Primary invariant:** **AD-25** (Thai break opportunities fail toward not breaking) — **amended by
this story under D-000.6**, per D-2.1.10.
**Adjacent invariants:** AD-2 (one fixed-point unit), AD-21 (every feature ships its golden),
AD-22 (toolchain pin / versioned change), AD-23 (no `float64`), AD-24 (boxes are absolute),
AD-1 (import lints), AD-14 (located diagnostics).
**Governing rulings:** **D-2.1.6 (the owner's decision — the spine of this story)** ·
**D-2.1.9** · **D-2.1.10** · **D-000.16** · **D-000.4** · **D-000.6** ·
**D-000.21 (sharpened)** · **D-000.22** · **D-000.23** · **D-000.24** · **D-000.26** ·
**D-000.28** · **D-000.29** · **D-000.30** · D-000.9 · D-000.14 · D-000.15 · D-000.17 ·
D-000.18 · D-000.19 · D-000.20 · D-1.1.b · D-1.3.5 · D-1.4.9 · D-1.4.12 · D-1.8.1 ·
D-2.1.1 · D-2.1.4 · D-2.1.14 · D-2.1.15 · D-2.2.1 · D-2.3.1 · D-2.3.2 · D-2.3.5 · D-2.3a.1
**Deferred work touched:** **DW-14** (owned by this story *if its corpora reach the 100-entry
`beginbfchar` cap* — see AC17) · **DW-11** (this story is inside its stated owner window; it is
discharged in writing at AC18, not silently skipped).

---

## Baseline, measured at creation

HEAD is **`0266a86`** — *"Story 2.3a: Audit the vendor boundary (finisher)"* — on `main`,
working tree clean (`git status --porcelain` empty, re-verified after every scratch file created
during this investigation was removed).

Every number below is stated with its scope, its invocation and its counting rule (D-000.26):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` occurrence, subtests included | **400 PASS · 2 FAIL** |
| `folio-go/` | the same invocation, counting only **top-level** results (`^--- PASS` at column 0) | **254 PASS · 2 FAIL** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **81 PASS · 0 FAIL** |
| `lint/` | the same invocation, top-level only | **43 PASS · 0 FAIL** |
| `folio-go/` | `GOWORK=off go list -m all` | exactly **two** modules: `github.com/panitw/folio/folio-go`, `github.com/boxesandglue/textshape v0.0.15` |

The two failures are `internal/text`'s `TestCorpusMeetsP6ExerciseFloors` (P6g opaque-name floor:
got 7, need ≥ 20) and `TestP2IndependentDPCrossCheck` (26 violations across 17 items). They are
**Story 2.1's deliberate, pre-stated shortfalls (D-2.1.14, D-2.1.15) — the baseline, not a
regression.** One of them, `TestP2IndependentDPCrossCheck`, is **this story's to close** (D-2.1.9);
the other, `TestCorpusMeetsP6ExerciseFloors`, is **not** and must be left failing.

---

## In plain terms (read this first if you just want the gist)

The engine now wraps text. Until this story it drew each piece of text as one unbroken line, however
long; it now decides where a line may end, measures each candidate, and lays the words out across as
many lines as the box needs. English wraps at spaces, Chinese between characters. Thai has no spaces
between words, so the engine consults an embedded dictionary — and where the dictionary cannot
recognise a stretch of text it refuses to guess and keeps that stretch whole.

The owner's decision from the Thai spike landed with it. Thai surnames are coined by law out of
everyday words, so a dictionary cannot tell a person's name from the words it was built from.
Guessing harder was rejected: a template may now declare that the values it fills in are never to be
split, and the engine honours that rather than inferring it.

The defect the spike left visible is closed. In some Thai text the promise to keep unrecognised
stretches whole quietly dissolved and the engine broke inside them anyway; twenty-six such places
were known and none remain. That moved the recorded file of computed break positions, on purpose,
and it is the only recorded file that moved. Line spacing comes from the tallest face a template's
declared font list allows, never from the text that happens to appear.

Three things will look wrong later and are not. One Thai test stays red on purpose — a different
story's disclosed shortfall, which closing here would hide. And two human sign-offs are outstanding
by design: someone must confirm the Thai reads correctly, and someone must confirm the break
positions fall correctly. Those are separate judgments, each held open by its own failing check
until a real person records a real date. Nothing here signed either.

Review closed two test weaknesses rather than arguing them away. A check meant to prove that line
measurement adds up could not have caught the error it named, because every place it split its
sample fell on a seam where the rival readings agree; it now splits inside a word too. And the rule
that line spacing must be read from the font file itself, never from a plausible substitute the font
library invents, was unguarded; it now has a check that fails if anyone breaks it.

---

## Story

As a template author,
I want text to wrap where a reader would expect it to,
so that a multi-line address or transaction description is legible rather than merely fitting.

---

## Do not re-open — settled rulings this story inherits

Each of these is **binding**. If something in this file appears to contradict one of them, **stop
and surface it** — do not resolve it by choosing.

### 1. D-2.1.6 (OWNER) — the template declares unbreakable fields. No segmentation heuristic.
A template may declare that a bound value must never be broken. The engine **never** infers it.
The reason is not implementation difficulty — it is that **the information is not in the character
stream**: `ศรีสุข` as a surname and `ศรี สุข` as two ordinary words are byte-identical, both halves
are in the dictionary, and Thailand's Surname Act makes the surname population unbounded by law.
**100 of 120 personal names (83%) in the corpus are built this way — the common case, not the
tail.** Measured over the shipped 62,107-word list.

**Explicitly rejected, with reasons on the record:** a bigger dictionary (both halves are already
in it); a surname list (makes the failure rarer *and less predictable*, plus payload); all-Thai-
atomic (trivially satisfies the rule and breaks the product). **Do not propose any of these, and
do not write a heuristic that approximates one.** If any acceptance criterion below appears to
require one, that is a defect in this file — raise it as `DECISION NEEDED`.

**Accepted cost, disclosed:** a Thai name embedded in **free-form text** ("transfer to Srisuk")
remains breakable. That is the part this mechanism does not reach, and it is stated, not fixed.

### 2. D-2.1.10 — the format key and its enforcement land **together**, here, in one commit.
Not earlier and not in 2.5. A format field with no consumer is a schema addition that silently
does nothing (D-1.4.8's presence flags before Story 1.6). Landing both makes it provably
load-bearing from its first commit: **set → no break inside the span; unset → breaks as before.**
The **AD-25 amendment executes in the same commit** (D-000.6). The field is *additive* — a new
optional key, **not** a closed-set extension — so D-1.4.12 permits it as MINOR and D-1.4.9's
passthrough already guarantees an older library loads a file carrying it.

### 3. D-000.16 — `internal/text` must **not** import `internal/bind`.
> *"The signal rides on the value, never through an import. `internal/text` must not import
> `internal/bind`; the BoundTree's text nodes carry the atomic flag and the breaking API accepts
> it as a **parameter**. Stages communicate by what they pass, not by what they import."*

This is a live, binding ruling written for exactly this story. The breaking API takes atomic spans
as an argument. **Do not add an import to satisfy convenience.**

D-000.16 also schedules a **stage-rank import guard** for *"the next story that creates an
`internal/` package"*. **This story creates none** (see the Scope fence) — so the guard does **not**
land here; it lands at 2.5, which creates `internal/layout` / `internal/pagemodel`.

### 4. D-2.1.9 — the P2 engine defect is **this story's to fix**.
> *"The engine behaviour — atomicity dissolving under context — Story 2.4; it is line-breaking
> behaviour, and 2.4 owns line breaking."*

### 5. AD-25's two existing absolutes are unchanged and still absolute.
No break inside a Thai character cluster. A run of Thai the dictionary cannot cover yields no
interior break. This story **adds a third mechanism** beside them; it removes neither.

### 6. D-000.30 — the red-proof window closes when the fix lands.
The pre-fix measurement of the P2 defect is captured in this file (Finding **M2**), taken at
`0266a86` *before* any fix exists. It is **evidence the defect was real. It is not a standing
red-proof** and must never be recorded as one.

### 7. The Epic 2 gate owes exactly two things today, and this story adds a third — stated, not smuggled.
Today: the four-target matrix legs, and the Thai sign-off `fixtures/shaped-text/thai-signoff.json`
bound to digest `5964aad0…c92e00f` (the sha256 of `expected.pdf`, recorded as a field in
`expected.json`). **Neither may be disturbed.** This story adds a **third**: the hand-check of the
new S4 expected-break fixture, which `acceptance.md:65` requires in so many words. See **ESC-2**.

---

## Corrections to inherited claims — verified at `0266a86`

| Claim as received | Status | Evidence |
|---|---|---|
| *"the two failures are Story 2.1's intentional …, expected, do not treat as regressions"* (brief) | **Half right.** `TestCorpusMeetsP6ExerciseFloors` is indeed not this story's. `TestP2IndependentDPCrossCheck` **is** — D-2.1.9 routes the engine fix to 2.4 by name. Leaving it red is a scope failure, not a preserved baseline. | D-2.1.9's two-fixes table; `folio-mvp-decision-log.md:4469`–`:4501` |
| *"`name-101`/`name-102` are 2.4's retained fixture"* (D-2.1.9, and still readable there) | **Superseded.** D-2.1.15 Major 4 re-points it to **`name-116`** (`สุขสันต์ ดอเลาะ`) and **`name-117`** (`ปรีดา แนแซ`). Confirmed by measurement: `name-101` (`ประภา ธรรมบุญ`) and `name-102` (`กาญจนา ธรรมตระกูล`) produce **zero** P2 violations at HEAD; `name-116`/`name-117` produce one each. | M2 below |
| *"P2: 2 violations"* (D-2.1.9, written pre-rebuild) | **Stale by an order of magnitude.** The live figure is **26 violations across 17 items** — the figure D-2.1.15 records as current. Do not carry the `2`. | M2 below |
| *"Heavy tests: per-epic cadence … do not run them"* (brief) | **Conflicts with a live binding ruling.** **D-000.4 names `2.4` explicitly as a per-story matrix override** — *"2.4 (line breaking feeds every measurement)"* — and D-2.1.1's test for the override (*"a story whose own deliverable is hash-shaped"*, *"emits PDF bytes"*) is **satisfied** by this story, unlike 2.1. Escalated as **ESC-1**. This file follows the brief (register, do not run) and does **not** silently retire the ruling. | `folio-mvp-decision-log.md:308`–`:332`; D-2.1.1 at `:4114` |
| *"breaks occur at whitespace per UAX #14"* (`epics.md:827`) | **Not implementable as literally written without a new dependency**, which `go list -m all` forbids. Narrowed explicitly — see **DN-3**. Nothing in this story may claim UAX #14 conformance. | `gomod_test.go:59` (`wantModuleGraph`) |
| *"the fixture is hand-checked once and then never regenerated to make a test pass"* (`epics.md:836`) | **Correct and load-bearing** — but the fixture it names **does not exist yet**. `fixtures/thai-break-corpus/computed_breaks.json` is the engine's own output and its own README says in terms that it is *"a cross-target REGRESSION ANCHOR ONLY, never a correctness oracle"* and *"must not be promoted to the frozen S4 conformance fixture"*. This story creates the real one. | `fixtures/thai-break-corpus/README.md` |
| `sprint-status.yaml` reads `epic-2: backlog` while 2.1–2.3a are `done` | **True, and deliberately not this story's to change** — the tracker advances epic-level keys at the epic's own gate (D-2.1.15, Nit 5, left open on explicit instruction). **Flagged for the Epic 2 gate. Do not edit it here.** | `sprint-status.yaml:44` |

---

## Scope fence — what this story is NOT

- **It does not create an `internal/` package.** Breaking and measurement live in the existing
  `internal/text`; the shaped advances that feed measurement are supplied by the caller as
  parameters. Consequently D-000.16's stage-rank guard does **not** land here.
- **It does not compose bands and does not paginate.** Vertical placement of a band, and a
  document growing onto a second page, are Stories 2.5 and 2.6. This story wraps **within one
  element's declared box**, at the band-relative origin that already exists.
- **It does not implement UAX #14.** See DN-3.
- **It does not implement CJK kinsoku** (prohibition of line-start/line-end punctuation). Named as
  absent, by category, not patched with a denylist of characters (D-000.23).
- **It does not introduce a second advance path.** Line widths come from the **same**
  `[]text.ShapedGlyph` that is drawn. Per-rune `hmtx` advances (`fontset.AdvanceForRune`) may not
  be used for measurement — that is precisely Story 2.3's Blocker 1.
- **It does not re-shape a line's text after breaking.** See AC7 and its reasoning.
- **It does not fix `TestCorpusMeetsP6ExerciseFloors`.** P6g's floor stands unmet (D-2.1.14).
  Do not add corpus items to reach a number (D-000.17).
- **It does not move any existing golden.** One recorded artifact changes, deliberately and with
  its own AC: `fixtures/thai-break-corpus/computed_breaks.json` (AC12). Every `expected.json`
  digest in `fixtures/` stays byte-identical — including `shaped-text`'s `5964aad0…c92e00f`.
- **It does not touch `fixtures/shaped-text/` or the two existing gate obligations.**
- **It does not commit.** See the Task breakdown's last line.

---

## Measured findings — read all of these before writing code

Every measurement below names its subject: the exact invocation, the file, the face and the input
string (D-000.26). Story 2.3 exists partly because a measured table named the wrong font and the
wrong text; a mis-aimed measurement propagates further than a stated assumption, because nobody
re-checks a measurement.

### M1 — The render path today draws one unbroken line per text element, and already carries the shaped answer that measurement needs

`folio-go/render.go`'s `splitByFace` resolves each rune's face across the fallback chain, groups
maximal same-face runs, shapes each one **once** with its own buffer, and carries the resulting
`[]text.ShapedGlyph` on the `textRunSource` it returns. The next segment's origin is computed from
**those same glyphs** — per-glyph advance scaled to the 1000-unit em individually via
`geom.ScaleRound`, then summed, then scaled by font size. There is exactly one advance path and it
must stay that way.

Nothing anywhere consults `Element.Width` for text. `internal/template.Element` already carries
`Width, Height Presence[geom.Length]`, and `folio-format.md:126` already documents them as
*"band-relative position and size, in points"*. **The box exists; nothing reads it.**

### M2 — The P2 defect at `0266a86`: 26 violations across 17 items (pre-fix capture, D-000.30)

Invocation: `CGO_ENABLED=0 GOWORK=off go test -count=1 -run TestP2IndependentDPCrossCheck -v ./internal/text/`
Subject: the committed 243-item corpus `fixtures/thai-break-corpus/corpus.json`, the committed trie
`folio-go/internal/text/data/thai_words.trie`, constrained mode (`unconstrained=false`).

**26 violations across 17 items.** Verbatim, from the test's own output:

| item | text | uncoverable Thai span | breaks proposed inside it (rune index) |
|---|---|---|---|
| `name-021` | `ชัยวัฒน์ วงศ์ไพร` | `ชัยวัฒน์` `[0,8)` | 3, 5, 6 |
| `name-026` | `อรุณี ทองตระกูล` | `อรุณี` `[0,5)` | 3 |
| `name-030` | `วรรณา ทองลักษณ์` | `วรรณา` `[0,5)` | 2 |
| `name-031` | `ศิริพร ทองทิพย์` | `ศิริพร` `[0,6)` | 2, 4 |
| `name-035` | `จันทิมา แก้วจิตร` | `จันทิมา` `[0,7)` | 3, 5 |
| `name-045` | `ไพโรจน์ จันทร์โภคา` | `จันทร์โภคา` `[8,18)` | 14, 16 |
| `name-060` | `ธัญญา พงษ์อนันต์` | `ธัญญา` `[0,5)` | 3 |
| `name-081` | `ชัยวัฒน์ ศักดิ์ฟ้า` | `ชัยวัฒน์` `[0,8)` | 3, 5, 6 |
| `name-086` | `อรุณี รัตน์เจริญ` | `อรุณี` `[0,5)` | 3 |
| `name-090` | `วรรณา รัตน์เมือง` | `วรรณา` `[0,5)` | 2 |
| `name-091` | `ศิริพร รัตน์ลักษณ์` | `ศิริพร` `[0,6)` | 2, 4 |
| `name-095` | `จันทิมา มณีหวาน` | `จันทิมา` `[0,7)` | 3, 5 |
| **`name-116`** | `สุขสันต์ ดอเลาะ` | `ดอเลาะ` `[9,15)` | 11 |
| **`name-117`** | `ปรีดา แนแซ` | `แนแซ` `[6,10)` | 8 |
| `name-120` | `ธัญญา หวั่งหลี` | `ธัญญา` `[0,5)` | 3 |
| `synthetic-039` | `ฉั่วสมบูรณ์` | `ฉั่วสมบูรณ์` `[0,11)` | 4 |
| `txn-006` | `ค่าน้ำประปา กปน.` | `กปน` `[12,15)` | 13 |

**`name-116` and `name-117` are the two retained fixtures** (D-2.1.15 Major 4). They are also the
only two genuinely-uncoverable, independently-attested sourced opaque names in the corpus (DW-11) —
the *hard* path, built from characters appearing in thousands of dictionary words. The 38
`ฅ`/`ฃ` synthetics exercise the *easy* path (`ฅ` appears in **2** of 62,107 words) and are not
substitutes.

**`name-101` and `name-102` produce zero violations.** The brief's inherited reference to them is
the pre-D-2.1.15 text and is wrong.

### M3 — The shape of the fix, measured over the whole corpus at `0266a86`

Measured by a scratch test written into `folio-go/internal/text/`, run, and then **removed**
(clean tree re-verified). Its source is reproduced verbatim below under *The measurement fixture*.

The candidate rule — call it the **both-sides-coverable filter**:

> A break opportunity strictly interior to a Thai-script span survives only if the span text
> **before** it and the span text **after** it are each fully tileable by dictionary words. The
> filter only ever **removes** break opportunities; it never adds one.

Measured effect over all 243 corpus items:

| quantity | at `0266a86` | with the filter |
|---|---|---|
| **P2 violations** (independent DP ground truth) | **26** | **0** |
| total break opportunities across the corpus | 558 | **526** |
| items whose break set changes | — | **23** |
| items that lose *all* their break opportunities | — | **1** (`synthetic-039`, `ฉั่วสมบูรณ์`, genuinely uncoverable — correct) |
| P3 violations (breaks inside a hand-labelled proper-noun span) | 172 | **142** |
| Thai-script spans in the corpus | 383 | — |
| …of which not fully dictionary-tileable | 58 | — |
| longest such span | 11 runes (`ฉั่วสมบูรณ์`) | — |

**Why P2 goes to exactly zero, and why that is a proof rather than a coincidence:** if some
interior position `p` had both sides tileable, concatenating the two tilings would tile the whole
span — contradicting the premise that the span is not tileable. The filter therefore closes P2 **by
construction**, in both directions, over any input. The measurement above confirms the
implementation matches the argument; it is not the argument.

**Why this is not the rejected "never break Thai" option:** the filter costs the corpus 32 of 558
break opportunities (5.7%) and empties exactly one item, which is an item nothing can legally
break. It also *reduces* proper-noun violations by 30. A variant that proposes a break at **every**
doubly-tileable position (rather than filtering the greedy matcher's proposals) was also measured:
749 breaks, 141 items changed — **rejected**, because it invents break positions the dictionary
matcher never proposed and would make P3 worse. **Filter, do not re-propose.**

### M4 — Introducing width-driven wrapping cannot move any existing golden. Measured, not assumed.

Every text-bearing fixture template already declares a `width` on every text element, so
"wrapping ignores absent widths" is **not** what protects the goldens. What protects them is that
every one of them fits, with a wide margin. Measured at `0266a86` with the same arithmetic
`splitByFace` uses (per-glyph `ScaleRound` to the 1000-em, summed, scaled by font size), against
the same font sets the golden tests use (`testFontSet()` for `font-text`, `testShippedFontSet()`
for the other two):

| fixture | element | size | text | measured width | declared width | fits |
|---|---|---|---|---|---|---|
| `fixtures/font-text/input.folio` | `e1` | 14 pt | `Hello, World!` | 78,218 mp | 400,000 mp | yes |
| `fixtures/font-text/input.folio` | `e2` | 9 pt | `Page footer 0123456789` | 99,522 mp | 400,000 mp | yes |
| `fixtures/multi-script-fallback/input.folio` | `e1` | 14 pt | `Ada ก 汉` | 55,090 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e1` | 16 pt | `ปั ฟั ที่ ป้ำ` | 58,992 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e2` | 16 pt | `ณัฐวุฒิ เกิด กรุงเทพ` | 129,696 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e3` | 16 pt | `น้ำ า` | 26,960 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e4` | 16 pt | `office fi AV Wo. To,` | 137,248 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e5` | 16 pt | `结算单，共３页` | 112,000 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e6` | 16 pt | `Ada ปั 结` | 63,040 mp | 500,000 mp | yes |
| `fixtures/shaped-text/input.folio` | `e7` | 16 pt | `AV ก` | 32,944 mp | 500,000 mp | yes |

Widest ratio: `e2`, **26%** of its box. `fixtures/minimal-rect/` and `fixtures/image-embed/` carry
no text elements at all.

**This measurement covers the five golden fixtures only.** It does **not** cover the templates
embedded as Go string literals in `render_test.go`, `render_bind_test.go`, `shaped_fixture_test.go`
and their siblings. If any of those overflows its declared width once wrapping is live, its
assertions will move. **That is a signal, not a nuisance:** AC13 requires the full suite green with
no digest moved, and any such move must be reported, not absorbed.

### M5 — What the leading (line-to-line advance) rule may be built from, with no new table read

`internal/fontset.Font.Metrics()` already returns `Ascent`, `Descent` and `CapHeight` **scaled to
the 1000-unit em** by `geom.ScaleRound`, and `hhea` is already in D-2.3a.1's `requiredTables` list
(`internal/fontset/fontset.go:273`), validated at ingestion with a located error naming face and
table. `hhea.lineGap` is **not** exposed and reading it would widen the vendor surface an audited
story just fenced. **A leading rule built from `Ascent - Descent` needs no new accessor, no new
table, and introduces no float.** The rule itself is not settled — see **DN-2**.

### M6 — `internal/text` imports nothing, and must keep that property

Verified at `0266a86` by reading the import blocks: `internal/text` imports only the standard
library and `github.com/boxesandglue/textshape/ot` (in `shape.go`). D-000.16's verified graph
(`text -> (nothing)` for internal packages) still holds. The atomic-span signal must arrive as a
**parameter**.

### M7 — `/ToUnicode` section sizes at `0266a86`, and why this story is DW-14's trigger

DW-14 records that `internal/pdf.buildToUnicodeCMap` emits one unbounded `beginbfchar` section
while the ToUnicode CMap specification caps a section at **100** entries. Measured section sizes at
this commit: `font-text` 25; `multi-script-fallback` 4, 1, 1; `shaped-text` 14, 7, 28. Story 2.3
both rewrote the entry source (CIDs are now allocated per (glyph, cluster text) — D-2.3.2, so the
count is no longer bounded by the glyph count) and produced the largest section so far.

**DW-14 names this story as its owner *"if its corpora reach the limit first"*.** A wrapped Thai
paragraph is exactly the input that would. AC17 makes that a measured decision rather than an
accident.

---

## DECISIONS NEEDED — escalate before development starts

> Standing instruction: if anything in this file looks like it contradicts a decision-log entry,
> **stop and surface it. Do not resolve it by choosing.**

### DN-1 — The spelling and granularity of the unbreakable-field format key
D-2.1.6 mandates *"one optional format field"*; D-2.1.10 mandates it land here with its enforcement.
Neither names it. Two independent questions:

1. **Granularity.** (a) **Element-level boolean** — every value substituted into this text
   element's `value` is an atomic span; literal text around the placeholders breaks normally.
   (b) **Per-placeholder** — the declaration names which bindings are atomic.
2. **Spelling.** `"unbreakableValues"` vs `"unbreakable"` vs something else.

**Recommendation: 1(a) + `"unbreakableValues"`.** Reasons: (a) matches D-2.1.6's own framing
(*"the way a real bank statement already knows which value is a customer name"*) and needs no new
grammar; the bare spelling `"unbreakable"` reads as *"this whole element is unbreakable"*, which is
the rejected all-atomic option wearing a smaller hat, and a reader who mis-reads it will author a
template that silently over-restricts. **The mechanism is identical under either spelling** — only
the key's text changes — so a different ruling costs one identifier and one field-table row, not a
redesign. The ACs below are written against the recommendation.

### DN-2 — The leading rule, and whether wrapped lines are placed by this story at all
Wrapping produces N lines; N > 1 needs a line-to-line advance, and no planning document defines
one. Two questions, and they are not independent:

1. Does 2.4 **emit** the wrapped lines into the PDF, or only compute the wrap and leave placement
   to 2.5? **D-000.4's own justification for naming 2.4 an override — *"line breaking feeds every
   measurement"* — and the epic's `S4` binding both read as "2.4 changes bytes".** Emitting is the
   reading this file is written against.
2. If it emits: what is the advance? **Recommendation:**
   `lineAdvance = ScaleRound(Ascent - Descent, fontSize, 1000)`, taken from the **first face of the
   element's chain that is present in the supplied FontSet** — not per-run, because a mixed-script
   line would otherwise have two candidate leadings and no rule to pick between them. Computed in
   **one** function, per AD-2's *"font scaling is one exported function … never open-coded at a call
   site"*. Needs no new vendor accessor (M5).

**This decision is permanent in a way most are not:** every golden recorded after it is welded to
it, and changing it later is an AD-21/AD-22 versioned break for every downstream test suite.

### DN-3 — "per UAX #14" cannot be implemented as written; state the narrowing or supply the data
`epics.md:827` says Latin *"breaks occur at whitespace per UAX #14"*. Full UAX #14 needs the
`LineBreak` property for every code point. That is either a new module (forbidden — `go list -m all`
must stay exactly two) or a generated table nobody has budgeted.

**Recommendation: implement the narrow rule and never claim the standard.** Specifically: a break
opportunity exists **after** a maximal run of Unicode `White_Space` characters, and the whitespace
run itself is consumed by the break (drawn on neither line). Nothing else in Latin breaks — no
hyphenation, no break at `-`, no break inside a word. Document the narrowing in the package doc and
in `folio-format.md`, and **amend `epics.md:827` under D-000.6** so the clause stops asserting a
conformance the engine does not have. `epics.md` is in the SPEC's companions list and D-000.6's
mechanism is exactly this case: a canonical document stating something unimplementable.

**If instead the owner wants real UAX #14**, that is a story of its own with a data-generation step,
and this story's Latin AC should be marked pending rather than narrowed.

### ESC-1 — D-000.4 names 2.4 as a per-story matrix override; the brief says per-epic
D-000.4's override clause names **`2.4`** explicitly, with its reason (*"line breaking feeds every
measurement"*), and D-2.1.1's test for the clause — *"a story whose **own deliverable** is
hash-shaped"*, *"2.1 emits no PDF bytes"* — is **satisfied** here: this story emits PDF bytes and
records a new golden. The commissioning brief instructs per-epic cadence (register the legs, do not
run them).

**This file follows the brief.** It does **not** rewrite D-000.4, and it does not pretend the
conflict is absent. D-000.4's own text requires that *"each override is logged with its reason in
the story's Delivery Log"* — a **waiver** of a named override deserves at least the same. Requested:
a logged ruling (a `D-2.4.x` entry) recording that 2.4's named override is waived to the Epic 2
gate, with its reason, so the next reader does not re-derive it or, worse, conclude it was
overlooked.

### ESC-2 — This story adds a THIRD Epic 2 gate obligation, and says so
`acceptance.md:65` requires *"an expected-break fixture for Thai and CJK line breaking (S4),
**hand-checked once** and then frozen"*, and `epics.md:836` repeats it. A hand-check of Thai break
positions is **irreducibly human** — no agent may claim it. D-2.3.5 permits exactly this state, on
three binding conditions, of which the first is that the pending obligation is a **failing test**,
gated as the matrix legs are, never a log entry.

So after this story the Epic 2 gate owes **three** things: the four-target matrix legs; the
existing Thai *rendering* sign-off (`fixtures/shaped-text/thai-signoff.json`, digest
`5964aad0…c92e00f`); and a new Thai *break* sign-off bound to the new fixture's digest.

**Recommendation: a second, separate `//go:build matrix` sign-off test, mirroring
`shaped_signoff_matrix_test.go` exactly, bound to the new fixture's own digest.** Do **not** extend
the existing record's schema or scope — it is welded to a digest and to a question ("does this Thai
*read* correctly"), and this is a different question ("do these breaks fall where a Thai reader
would put them"). Merging them would disturb an artifact this story is instructed not to disturb.

**The machine-checkable half is NOT deferrable** (D-2.3.5, D-000.22): every property in AC10 lands
at recording. Only the human read may be pending.

### DN-4 — Does the wrapped-PDF golden need its own human sign-off?
**Recommendation: no**, and the reasoning should be ruled on rather than assumed. Its Thai *glyph
forms* are the existing `shaped-text` sign-off's question; its *break positions* are ESC-2's
fixture's question; everything else about it (line count, per-line width against the box, no break
inside a cluster, no break inside a declared span) is machine-checkable and lands at recording.
Two human obligations, not three.

---

## Acceptance Criteria

Numbering is this story's own. Every AC states the artifact it asserts on, the precondition proving
that artifact carries the property (D-000.21 sharpened), and its red-proof (or its explicit
D-000.24 label).

### Group A — break opportunities in all three scripts

**AC1 — Latin breaks at whitespace, and the rule is stated as narrow.**
`internal/text` exposes break opportunities for non-Thai, non-CJK text at the position **after**
each maximal run of Unicode `White_Space`, and nowhere else. The whitespace run is consumed by the
break. The package doc and `folio-format.md` state, in words, that this is **not** UAX #14 and name
what is absent (hyphenation, break at `-`, contextual pair rules).
*Presence precondition:* the assertion runs against a returned opportunity set that is non-empty
for at least one input, so a function returning `nil` for everything cannot pass.
*Red-proof:* change the rule to break **before** the whitespace run instead of after; the
`"Page footer 0123456789"` case moves its break index and the assertion fails.
*Subject to DN-3.*

**AC2 — CJK breaks between characters, and kinsoku is named as absent.**
A break opportunity exists between any two adjacent runes that are both in the CJK ideograph or
kana ranges the engine already classifies for face resolution. The package doc states that
line-start/line-end punctuation prohibition (kinsoku) is **not implemented**, **as a category** —
no denylist of characters may appear (D-000.23: a guard written for a defect covers the defect, not
its class; a denylist is never coverage).
*Presence precondition:* asserted on `结算单，共３页` from the existing shaped-text fixture, whose
CJK content is already committed and whose glyph coverage is already proven.
*Red-proof:* restrict the rule to one of the two ranges; the mixed sample loses opportunities and
the count assertion fails.

**AC3 — Thai breaks come from the embedded dictionary, under both existing absolutes.**
Thai break opportunities continue to come from `ComputeBreaks` in constrained mode. No break is
proposed inside a Thai character cluster (AD-25, P1); a Thai run the dictionary cannot cover
proposes no interior break (AD-25, P2 — now genuinely, see AC4).
*Red-proof:* AC4's.

**AC4 — The atomic-unknown-run absolute holds under context. `TestP2IndependentDPCrossCheck` goes GREEN.**
The both-sides-coverable filter of **M3** is implemented: a break opportunity strictly interior to a
Thai-script span survives only if the span text before it and the span text after it are each
fully tileable by dictionary words. It **filters** the existing matcher's proposals and never adds
a proposal.
*Assert on:* `TestP2IndependentDPCrossCheck`, unchanged — its ground truth is an independent DP
that never calls `ComputeBreaks` (D-000.9), which is exactly why it is the right gate.
*Presence precondition:* the test's own vacuity guards already require a non-empty corpus and
non-zero break totals; do not weaken them.
*Expected result, pre-stated:* **0 violations**, from 26 across 17 items.
*Red-proof (standing, constructible after the fix):* delete the filter; the test reports 26
violations across 17 items again, `name-116` and `name-117` among them.
*Pre-fix capture (D-000.30):* Finding **M2** above, taken at `0266a86`. It is evidence the defect
was real; **it is not a standing red-proof and must not be recorded as one.**

**AC5 — `TestCorpusMeetsP6ExerciseFloors` stays RED, and nothing in this story touches the corpus.**
P6g's floor (7 of ≥ 20) is D-2.1.14's disclosed, deliberate shortfall. `corpus.json` and
`cmd/gencorpus` are not edited by this story. Reporting this story's results must state the
remaining failure explicitly rather than counting it as a regression.
*Red-proof:* not applicable — this is a **preservation** criterion; its evidence is a byte-identity
check on `fixtures/thai-break-corpus/corpus.json` and `folio-go/cmd/gencorpus/main.go`.

### Group B — the declared unbreakable span (D-2.1.6, D-2.1.10)

**AC6 — The format key, its consumer and the AD-25 amendment land in ONE commit.**
Three things, together, per D-2.1.10:
1. `internal/template` gains the optional element key (DN-1: `"unbreakableValues"`), parsed,
   serialized, round-tripping as an absent key when absent (the canonical-fixed-point property
   Story 1.4 pins), and documented as a row in `folio-format.md`'s element field table.
2. `internal/bind` reports, alongside the bound string, the **rune spans** it substituted — so the
   caller can mark them atomic. `internal/bind` decides nothing about breaking.
3. The breaking API accepts atomic spans as a **parameter** (D-000.16). `internal/text` gains **no
   new import**; `TestArchImports`-style guards must still pass.
4. `ARCHITECTURE-SPINE.md`'s **AD-25 Rule** gains a third mechanism, in the same commit, under
   D-000.6 — changing only the Rule's mechanism list, leaving **Binds** and **Prevents** untouched,
   and quoted verbatim before/after in the decision log.
*Red-proof:* one template, two renders — key set vs key absent — over a bound value the dictionary
would otherwise break (`ศรีสุข` is the D-2.1.6 worked case, both halves in the dictionary). Set →
the value occupies one line unbroken; unset → it breaks. **A test asserting only the "set" case is
vacuous:** it cannot distinguish an honoured declaration from a value that never had a break
opportunity. Both polarities are required.

**AC7 — A declared span is atomic; a literal is not; and the guard covers the mechanism, not the sample.**
Within one text element, break opportunities strictly interior to any substituted span are removed
when the key is set. Literal text between placeholders is unaffected. Coverage is by
**construction of the span set** — every substituted span, whatever its script — never by a list of
sample strings (D-000.23).
*Presence precondition:* the assertion first proves the un-declared form of the same input **does**
carry an interior break opportunity in that span; otherwise it is asserting the absence of
something that was never there.
*Red-proof:* mark only the first substituted span atomic instead of all of them; a two-placeholder
case fails.

**AC8 — The disclosed limitation is written down where a reader will meet it.**
`folio-go/README.md` and the package doc state plainly: a Thai name in **free-form text** remains
breakable; the mechanism protects **declared values**. This is D-2.1.6's accepted cost and it is
disclosed, not fixed.
*Red-proof:* none available — this is documentation. Labelled as such; not credited with a proof it
does not have (D-000.24).

### Group C — measurement and wrapping

**AC9 — Line width is measured from the shaped advances that are drawn, in exact integer millipoints, with no second advance path.**
A candidate line's width is the sum of the **already-scaled** per-glyph advances of the glyphs on
that line, taken from the same `[]text.ShapedGlyph` the run will be drawn from — scaled per glyph
to the 1000-unit em via `geom.ScaleRound`, summed, then scaled by font size, in that order (the
already-rounded space the viewer's pen consumes). `fontset.AdvanceForRune` may **not** appear on
the measurement path. No `float32`/`float64` anywhere; `lint`'s syntactic **and** type-aware rules
both stay at zero findings under `internal/` and the module root (AD-23, D-000.25).
*Presence precondition:* asserted on a run whose shaped advance is provably ≠ its `hmtx` sum — the
committed `AV` case in Noto Sans (shaped 599 vs `hmtx` 639) is the discriminating input; a run
with no kerning cannot tell the two hypotheses apart.
*Red-proof:* swap the measurement to per-rune `hmtx`; the `AV` case's measured width moves and the
assertion fails.

**AC10 — Lines are cut at cluster boundaries by slicing the ONE shaped run, never by re-shaping.**
A line boundary at rune index `r` splits a face-segment's glyph slice at the first glyph whose
`Cluster >= r`. Break opportunities never fall inside a cluster (AC3), so the slice is exact. The
sub-run is **emitted from the same glyphs that were measured**. Re-shaping the shorter text is
forbidden: it would reintroduce the second derivation Story 2.3's Blocker 1 removed, and it can
legitimately produce different glyphs at the new boundary.
*Presence precondition:* `GlyphInfo.Cluster` is a **rune index, not a byte offset** — pinned by
`TestClusterValuesAreRuneIndices`. Note which inputs can tell the two apart: `"office"` cannot
(ASCII); `"ณัฐวุฒิ"` can (byte offsets would be multiples of 3). Use a Thai case.
*Red-proof:* treat `Cluster` as a byte offset in the slice search; the Thai case slices at the
wrong glyph and the emitted line's width no longer equals the measured line's width.

**AC11 — A text element wraps within its declared width; the box is honoured, and overflow is not silently absorbed.**
When a text element declares a `width`, its bound text is laid out as one or more lines, each of
whose measured widths is ≤ the declared width **except** where a single atomic unit (an unbreakable
declared span, an uncoverable Thai run, an unbroken Latin word) is itself wider than the box. That
case **overflows visibly** — AD-25's own words, *"it overflows visibly under FR44 — clipped, with a
located diagnostic"* — and is **not** re-broken at a guess. Clipping and the diagnostic are Story
2.8's; this story must not silently drop, silently squeeze, or silently re-break.
*Presence precondition:* the assertion is made on an input measured (before the assertion) to be
wider than its box, so a fitting input cannot satisfy it vacuously.
*Red-proof:* allow a break inside an atomic unit when it does not fit; the unbreakable-span case
splits and AC7's assertion fails too — which is the point: the two ACs guard the same invariant
from opposite sides.

### Group D — fixtures, goldens and gates

**AC12 — `computed_breaks.json` is re-derived as an INTENDED, versioned change, with its reason recorded.**
The P2 fix changes computed break positions for **23 of 243** corpus items (M3). `computed_breaks.json`
is regenerated by `go run ./cmd/genbreaks` and committed. `TestAC10ComputedBreaksMatchS4Basis` goes
green against the new file. The story records: the pre-change and post-change totals (**558 → 526**),
the count of items changed (**23**), the one item that loses all breaks (`synthetic-039`), and the
ruling that authorises the move.
**This file is not a correctness oracle and must not be promoted to one** — its own README says so,
and says re-derivation against 2.4's mechanism is exactly what must happen first.
*Red-proof:* leave the file unregenerated; `TestAC10ComputedBreaksMatchS4Basis` reddens on 23 items.
*Constraint:* **this is the only recorded artifact this story moves.** Every `expected.json` digest
under `fixtures/` stays byte-identical, `shaped-text`'s `5964aad0…c92e00f` included. A hash change
anywhere else is investigated as a defect until proven intended (AD-21) — **stop and escalate; do
not re-record.**

**AC13 — Every existing golden is byte-identical, proven, not assumed.**
The full suite is green at the same counting rule as the baseline table, with **exactly one**
remaining failure (`TestCorpusMeetsP6ExerciseFloors`, AC5) where the baseline had two. Every
fixture digest is re-verified. M4 shows why this should hold for the five golden fixtures; it does
**not** cover the templates embedded as Go literals in the test files, and any assertion that moves
there is reported in the Delivery Log with its cause.

**AC14 — The new S4 expected-break fixture is HAND-CHECKED, small enough to be hand-checkable, and never regenerated to make a test pass.**
A new fixture — recommended `fixtures/expected-breaks/` — holds a **small** table (target: 20–30
items; it must be readable by one person in one sitting) of Thai and CJK inputs with their
**expected** break positions, authored as **labels**, not as engine output. `epics.md:836` and
`acceptance.md:65` both bind here: *hand-checked once and then never regenerated to make a test
pass.*
*Ordering is load-bearing* (Trap 1, Story 2.1): author and review the labels **first**, commit them
**second**, assert against them **third**. Never the reverse.
*Presence precondition + vacuity guard:* the fixture must contain at least one item with a non-zero
expected break count and at least one with zero, and the test must fail if the fixture is empty,
shorter than the item list, or all-zero.
*Red-proof:* invert one expected position in a scratch copy; the test reddens naming the item.
*The human half* is ESC-2's pending sign-off. **Machine-checkable properties are asserted at
recording and are not deferrable** (D-2.3.5).

**AC15 — A new rendered golden for wrapped text, with its semantic acceptance step at first recording (D-000.22).**
A new fixture — recommended `fixtures/wrapped-text/` — with `input.folio`, `expected.pdf` and
`expected.json` carrying `folioGoVersion`, `goToolchain` and `sha256`, matching the shape every
existing fixture uses. It must contain a Latin paragraph, a CJK paragraph and a Thai paragraph,
each in a box narrow enough to force **at least two lines** (measured and stated, not assumed), plus
one element exercising a **declared unbreakable span**.
*Semantic properties asserted at recording — all machine-checkable, none deferrable:*
(a) each element's rendered line count equals the expected count; (b) no line's measured width
exceeds its declared width except at a declared atomic unit; (c) no line boundary falls inside a
Thai character cluster; (d) no line boundary falls inside a declared unbreakable span; (e) the
declared page size and the embedded faces are what the template asked for; (f) AC17's
`beginbfchar` section sizes.
*Red-proof for the fixture's own observability:* point the same assertions at
`fixtures/shaped-text/input.folio`, every element of which fits its box (M4) — it reports **one
line per element** and (a) fails. A fixture that cannot fail this way is not exercising wrapping.

**AC16 — The four matrix legs are written and registered for the new document, and NOT run.**
`wrapped-text` (and, if it renders a document, the expected-breaks fixture) is added to
`matrix_test.go`'s `matrixDocuments` **and** to `.github/workflows/matrix.yml`'s `docs="…"` line
**and** to the per-target `upload-artifact` paths for all four targets
(`darwin-arm64`, `linux-amd64`, `linux-arm64`, `js-wasm`). `TestMatrixDocumentSlugsAreRegisteredInCI`
is the untagged guard that catches a document registered in one list and not the other — it must
stay green. `go vet -tags=matrix ./...` must compile the tagged files.
**The legs are not executed in this story.** See ESC-1 — this is a deviation from D-000.4's named
override for 2.4 and is recorded as one, not passed over.
*Red-proof:* add the slug to one list only; the registration guard reddens naming the missing side.

**AC17 — `/ToUnicode` section sizes are measured for the new fixtures, and DW-14 is either not triggered or taken deliberately.**
For every face in every new fixture, the emitted `beginbfchar` section size is measured and
recorded. If every section is **≤ 100**, DW-14 stays open and the measured numbers are appended to
it (baseline for comparison: `font-text` 25; `multi-script-fallback` 4, 1, 1; `shaped-text` 14, 7,
28). If any section **exceeds 100**, DW-14 is triggered on this story's watch: **stop and escalate**
— the fix (chunking into ≤ 100-entry sections) moves every golden hash of every document over the
cap, and DW-14's own text says it *"wants to land with a deliberate re-record rather than as a
drive-by."*
*Do not* shrink the fixture merely to duck the cap without saying so; if the fixture is sized to
stay under it, say that is why.

**AC18 — DW-11 is discharged in writing.**
DW-11's owner window is *"Epic 2's later stories and Epic 4's golden-report work"*, and this story
is inside it. Either add genuinely-opaque, **independently attested** sourced Thai personal names to
the corpus, or **state in writing that none were found and that the load-bearing count remains 2**
(`ดอเลาะ`, `แนแซ`). **Do not invent items to reach a number** (D-000.17, and D-2.1.15 Major 5's
precedent, where "a plausible surname" was demoted rather than counted).

### Group E — the standing invariants this story must not break

**AC19 — Guards and constraints, all re-verified.**
`go list -m all` stays **exactly two modules**. No new dependency; HarfBuzz remains a one-time
offline oracle frozen into a fixture, never a runtime dependency (D-2.3.1, AD-25). No compressor
import under `folio-go/` (`no-compressor-import`, D-1.8.1). No `fmt` in `internal/pdf` (D-1.1.b).
**No literal `SOURCE_DATE_EPOCH` in any `.go` under `folio-go/` — including inside an
error-message string** (`absence-source-date-epoch`, D-2.1.5). No map ranged where its order can
reach an output byte (`map-range`, D-1.3.5). `lint` stays at **81/0** (every occurrence) and
**43/0** (top-level).

---

## The measurement fixture, verbatim

The scratch test below was written into `folio-go/internal/text/zz_scratch_test.go` at `0266a86`,
run, and then **removed**; the clean tree was re-verified with `git status --porcelain`. The dev
agent may restore it byte-for-byte as an early task to confirm the baseline has not moved. If its
numbers differ from M3's, **the baseline moved and the evidence must be re-taken** — do not proceed
on M3's figures.

Invocation:
`CGO_ENABLED=0 GOWORK=off go test -count=1 -run TestZZScratchP2Shape -v ./internal/text/`

Recorded output at `0266a86`:

```
SCRATCH: items=243 thaiSpans=383 nonSegmentableSpans=58 maxNonSegLen=11 ("ฉั่วสมบูรณ์")
SCRATCH: total breaks current=558 newRule=526 itemsChanged=23 itemsLosingAllBreaks=1
SCRATCH: P2 violations under variant C = 0
SCRATCH: P3 violations current=172 newRule=142
SCRATCH: nonseg span length histogram map[3:4 4:27 5:18 6:3 7:2 8:2 10:1 11:1]
```

```go
package text

import (
	"testing"
)

// reachFwd[i] : runes[0:i] is fully tileable by dictionary words.
func reachFwd(dict *BytesTrie, runes []rune) []bool {
	n := len(runes)
	r := make([]bool, n+1)
	r[0] = true
	for i := 0; i < n; i++ {
		if !r[i] {
			continue
		}
		for j := i + 1; j <= n; j++ {
			if dict.Contains(string(runes[i:j])) {
				r[j] = true
			}
		}
	}
	return r
}

// reachBwd[i] : runes[i:n] is fully tileable by dictionary words.
func reachBwd(dict *BytesTrie, runes []rune) []bool {
	n := len(runes)
	r := make([]bool, n+1)
	r[n] = true
	for j := n; j > 0; j-- {
		if !r[j] {
			continue
		}
		for i := 0; i < j; i++ {
			if dict.Contains(string(runes[i:j])) {
				r[i] = true
			}
		}
	}
	return r
}

func TestZZScratchP2Shape(t *testing.T) {
	items := loadCorpus(t)
	dict := Dictionary()

	var spans, nonSeg int
	maxNonSegLen := 0
	var maxNonSegText string
	lenHist := map[int]int{}
	var totalCur, totalNew, p3cur, p3new int
	itemsLosingAll := 0
	itemsChanged := 0
	for _, it := range items {
		runes := []rune(it.Text)
		cur, _ := ComputeBreaks(dict, it.Text, false)
		newBreaks := map[int]bool{}
		ss := scriptSpans(runes)
		for si, sp := range ss {
			if si < len(ss)-1 {
				newBreaks[sp.end] = true
			}
			if !sp.thai {
				continue
			}
			sr := runes[sp.start:sp.end]
			spans++
			f := reachFwd(dict, sr)
			b := reachBwd(dict, sr)
			if !f[len(sr)] {
				nonSeg++
				if len(sr) > maxNonSegLen {
					maxNonSegLen = len(sr)
					maxNonSegText = string(sr)
				}
				lenHist[len(sr)]++
			}
			// VARIANT C: filter the engine's own proposals; never add one.
			for _, p := range cur {
				if p > sp.start && p < sp.end {
					if f[p-sp.start] && b[p-sp.start] {
						newBreaks[p] = true
					}
				}
			}
		}
		for _, pn := range it.ProperNounSpans {
			for _, p := range cur {
				if p > pn[0] && p < pn[1] {
					p3cur++
				}
			}
			for p := range newBreaks {
				if p > pn[0] && p < pn[1] {
					p3new++
				}
			}
		}
		totalCur += len(cur)
		totalNew += len(newBreaks)
		if len(newBreaks) != len(cur) {
			itemsChanged++
		}
		if len(cur) > 0 && len(newBreaks) == 0 {
			itemsLosingAll++
		}
	}

	// P2 under variant C, measured against the SAME independent DP
	// ground truth p2_independent_test.go uses (isFullySegmentable).
	var p2new int
	for _, it := range items {
		runes := []rune(it.Text)
		cur, _ := ComputeBreaks(dict, it.Text, false)
		kept := map[int]bool{}
		for _, sp := range scriptSpans(runes) {
			if !sp.thai {
				continue
			}
			sr := runes[sp.start:sp.end]
			f := reachFwd(dict, sr)
			b := reachBwd(dict, sr)
			for _, p := range cur {
				if p > sp.start && p < sp.end && f[p-sp.start] && b[p-sp.start] {
					kept[p] = true
				}
			}
		}
		for _, sp := range scriptSpans(runes) {
			if !sp.thai {
				continue
			}
			if isFullySegmentable(dict, runes[sp.start:sp.end]) {
				continue
			}
			for p := range kept {
				if p > sp.start && p < sp.end {
					p2new++
				}
			}
		}
	}

	t.Logf("SCRATCH: items=%d thaiSpans=%d nonSegmentableSpans=%d maxNonSegLen=%d (%q)",
		len(items), spans, nonSeg, maxNonSegLen, maxNonSegText)
	t.Logf("SCRATCH: total breaks current=%d newRule=%d itemsChanged=%d itemsLosingAllBreaks=%d",
		totalCur, totalNew, itemsChanged, itemsLosingAll)
	t.Logf("SCRATCH: P2 violations under variant C = %d", p2new)
	t.Logf("SCRATCH: P3 violations current=%d newRule=%d", p3cur, p3new)
	t.Logf("SCRATCH: nonseg span length histogram %v", lenHist)
}
```

**This scratch file must not be committed.** It ranges maps for reporting only and is not written
to this project's standards; it exists to reproduce M3, nothing more.

The width measurement of **M4** was taken by an equivalent scratch test in package `folio` using
`documentBands`, `bind.BindText`, `fontChain` and `splitByFace`, with `testFontSet()` for
`fixtures/font-text/` and `testShippedFontSet()` for the other two. It was likewise removed.

---

## Task breakdown

1. **Restore and run the M3 scratch test** (verbatim above) at HEAD. Confirm the five recorded
   lines reproduce. If they do not, **stop** — the baseline moved and every figure in this file
   needs re-taking. Remove the scratch file afterwards; `git status --porcelain` must be empty
   before task 2.
2. **Surface DN-1, DN-2, DN-3, DN-4, ESC-1 and ESC-2** and wait for rulings. Several of them
   (DN-2 especially) are welded into every golden this story records; recording first and asking
   after is not recoverable.
3. **Close P2** (AC4): implement the both-sides-coverable filter in `internal/text`. Run
   `TestP2IndependentDPCrossCheck` and confirm **0**. Then delete the filter, confirm **26 across
   17**, and restore it — that is the standing red-proof, and it is constructible only now.
4. **Regenerate `computed_breaks.json`** (AC12) via `go run ./cmd/genbreaks`; confirm 23 items
   changed and `TestAC10ComputedBreaksMatchS4Basis` green. Confirm `TestCorpusRegeneratedMatchesCommitted`
   still green and `corpus.json` byte-unchanged (AC5).
5. **Latin and CJK break opportunities** (AC1, AC2), with their red-proofs and their stated
   narrowings in the package doc.
6. **The unbreakable declaration** (AC6, AC7, AC8): the format key in `internal/template`; the
   substituted-span report from `internal/bind`; the atomic-span **parameter** into `internal/text`
   (no new import — re-run the arch/import guards); the `folio-format.md` field-table row; the
   **AD-25 amendment** in `ARCHITECTURE-SPINE.md`, quoted verbatim before/after in the decision log.
   Both polarities of the red-proof.
7. **Measurement and wrapping** (AC9, AC10, AC11), reusing the one shaped run. Confirm `lint`'s
   syntactic and type-aware float rules both report zero under `internal/` and the module root.
8. **Author the S4 expected-break labels** (AC14) — labels first, commit second, assert third —
   and wire ESC-2's `//go:build matrix` sign-off test bound to the fixture's digest.
9. **Record the wrapped-text golden** (AC15) with its full machine-checkable acceptance set, and
   run its observability red-proof against `fixtures/shaped-text/input.folio`.
10. **Measure `/ToUnicode` section sizes** for the new fixtures (AC17). Append to DW-14, or stop
    and escalate if anything exceeds 100.
11. **Register the matrix legs** in `matrix_test.go`, `matrix.yml`'s `docs="…"` and all four
    per-target `upload-artifact` paths (AC16). `go vet -tags=matrix ./...` compiles. **Do not run
    the legs.**
12. **Discharge DW-11 in writing** (AC18).
13. **Re-verify everything** (AC13, AC19): full suite at both counting rules, exactly one remaining
    failure, every fixture digest unchanged except `computed_breaks.json`, `go list -m all` exactly
    two modules, `lint` at 81/0 and 43/0.
14. **Story file, decision log and `sprint-status.yaml` → `review`.**

> **Stop here — do not commit, do not branch, do not set `done`.** Committing belongs to the
> finisher, after review. Do **not** change `epic-2`'s tracker key; that advances at the epic's own
> gate.

**All fourteen tasks are complete.** Task 2's escalations were resolved by D-2.4.1–D-2.4.4 before
development and by D-2.4.6 mid-story; nothing was recorded before its ruling. Task 11's legs were
**registered AND run**, per D-000.4's named override. Nothing is committed and `epic-2` is
untouched.

---

## Heavy-test cadence — what is deferred, and to which gate

**SUPERSEDED AT DEVELOPMENT TIME — read this, not the paragraph ESC-1 was written against.**

The four-target matrix legs for this story's new document were **written, registered AND RUN, in
this story.** D-000.4 names `2.4` as a per-story override explicitly (*"line breaking feeds every
measurement"*), and **there was never a waiver to log — the commissioning brief was simply wrong.**
ESC-1 asked for the deviation to be recorded as a ruling; asking for it is how it was discovered
never to have been sanctioned (D-000.31).

Logged here as D-000.4 requires of an override, with its reason: **line breaking feeds every
measurement**, and this story's own deliverable is hash-shaped. Result: all six registered documents
agree across `darwin/arm64`, `linux/amd64`, `linux/arm64` and `js/wasm`, and each matches its
recorded golden. `wrapped-text` at `3845da37…a712288e`, 72,743 bytes on all four.

After this story the **Epic 2 gate owes three things**, not two:

1. the four-target matrix legs still owed by 2.1, 2.2, 2.3 and 2.3a — **2.4's own are discharged**;
2. the existing Thai *rendering* sign-off, bound to `5964aad0…c92e00f`;
3. the new Thai *break* sign-off, bound to `expected_breaks.json`'s digest, computed at runtime
   (currently `a545e042…99324de`) (**D-2.4.3**).

Verified, not asserted: `fixtures/shaped-text/` is byte-unchanged, `thai-signoff.json` is still
absent, and the two sign-offs are bound to different digests by different tests with different
schemas — **a re-record of either invalidates exactly one.** D-2.4.3 also confirms DN-4: there is
**no third sign-off** for the wrapped-PDF golden, because its human judgment is break placement and
that is already covered one layer up, where it is legible.

Also carried to the gate: **`sprint-status.yaml` still reads `epic-2: backlog` although 2.1, 2.2,
2.3 and 2.3a are all `done`.** Flagged, not fixed.

---

## Dev Agent Record — completion notes

**Status: COMPLETE. Every task done, every AC verified, the four-target matrix RUN and green.**
Nothing below is anticipatory; every line records an event that has happened, with the invocation
that produced it (D-000.28, D-000.26).

**The central result: P2 went 26 → 0, and the standing red-proof was constructed and verified
BEFORE the fix landed** — removing the filter reproduces exactly 26 violations across 17 items,
`name-116` and `name-117` among them. A fix whose absence is provable is the strongest form
available (D-000.30).

**One decision was escalated and ruled mid-story:** D-2.4.2's leading rule resolved to the
**maximum over the declared font chain**, recorded as D-2.4.6 with its measurements. Work resumed
from that point; nothing was recorded before the ruling.

### Task 1 — baseline re-verified, M3 reproduced exactly

`CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...` in `folio-go/` at `0266a86`:
**400 PASS / 2 FAIL** all-occurrences, **254 / 2** top-level. Matches the baseline table.

The M3 scratch test was restored byte-for-byte, run, and removed (`git status --porcelain` clean
before task 2). All five recorded lines reproduced **identically**:

```
SCRATCH: items=243 thaiSpans=383 nonSegmentableSpans=58 maxNonSegLen=11 ("ฉั่วสมบูรณ์")
SCRATCH: total breaks current=558 newRule=526 itemsChanged=23 itemsLosingAllBreaks=1
SCRATCH: P2 violations under variant C = 0
SCRATCH: P3 violations current=172 newRule=142
SCRATCH: nonseg span length histogram map[3:4 4:27 5:18 6:3 7:2 8:2 10:1 11:1]
```

M3's figures are therefore live, not inherited.

### Task 3 — P2 closed (AC4). `TestP2IndependentDPCrossCheck` is GREEN.

The both-sides-coverable filter is implemented in `folio-go/internal/text/tileable.go` and applied
in `ComputeBreaks` **constrained mode only** (unconstrained is P6f/P6g's naive matcher and must
keep reproducing AD-25's Prevents line verbatim).

- After the fix: **0 violations across 0 items** (from 26 across 17).
- **Standing red-proof constructed and verified** (D-000.30's window, used before it closed):
  with the filter call removed, the test reports **26 violations across 17 items**, `name-116`
  (`ดอเลาะ`, break at rune 11) and `name-117` (`แนแซ`, break at rune 8) among them. Restored and
  re-confirmed green.
- `runs` is deliberately NOT filtered — it is the greedy dictionary decomposition the P6 exercise
  floors count. P6 stats are byte-identical to baseline: `{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284
  P6f:115 P6g:7}`. The divergence between `runs` and `breaks` is documented at `ComputeBreaks`.
- P3 moved **172 → 142**, exactly as M3 predicted.

Three new independent cross-checks (`tileable_test.go`), all green:
`tileableForward`/`tileableBackward` agree with the independent DP on all **383** Thai spans
(**325** tileable, **58** untileable); `forEachWordPrefix` agrees with `Contains` on **29,205 of
29,205** prefix comparisons, **2,086** of them positive; the filter is subtractive over the whole
corpus — **259** interior proposals examined, **32** withdrawn, **0** added (32 = 558 − 526).

### Task 4 — `computed_breaks.json` re-derived (AC12)

`go run ./cmd/genbreaks`. Measured against `HEAD`: **558 → 526** total breaks, **23 of 243** items
changed, **`synthetic-039`** the one item losing all breaks. `TestAC10ComputedBreaksMatchS4Basis`
green (526 total positions). `corpus.json` and `cmd/gencorpus/main.go` byte-unchanged (AC5).
**It is the only fixture that moved** — `git diff --stat -- fixtures/` lists it alone.

### Task 5 — Latin and CJK break opportunities (AC1, AC2)

`internal/text/opportunity.go`: `Opportunities(dict, s, atomic)` returning `Opportunity{LineEnd,
NextStart}` — modelling the consumed whitespace run explicitly rather than as a special case at
the measuring site. Both red-proofs **verified by mutation and restored**:

- AC1: emitting `NextStart = i` (break before the run) fails naming `"Page footer 0123456789"`.
- AC2: dropping `unicode.Han` from the classifier makes `结算单，共３页` yield zero opportunities and
  the presence precondition fires.

**A correction to AC2's premise, recorded rather than absorbed.** AC2 says the CJK ranges are ones
"the engine already classifies for face resolution". **No such classification existed** — face
resolution is COVERAGE-based (`resolveRuneFace` asks each face whether it has a glyph) and never
classifies a rune by script. The classifier is therefore new, built from the standard library's own
`unicode.Han` / `Hiragana` / `Katakana` script tables — a category test, not a character list
(D-000.23) — and adds no module.

### Task 6 — the unbreakable declaration (AC6, AC7, AC8)

Per **D-2.4.1**, document-level, a list of bare root-relative dotted data paths:

1. `internal/template`: `Document.UnbreakableValues`, parsed, serialized, **round-tripping as an
   absent key when absent** (asserted directly — emitting `[]` would move every existing golden).
   Every entry validated against D-1.4.1's convention at load; `{{customer.name}}`, `sum(x)`,
   `transactions[].payee`, interior whitespace, leading/trailing dots and duplicates are all load
   errors naming the field. The serializer's key count went 51 → **52** and `TestDriftGoToDoc`
   caught the undocumented key before any test did — the guard working as designed.
2. `internal/bind`: `BindTextSpans` returns the substituted **rune** spans with their paths.
   `BindText` now delegates to it, so there is exactly **one** implementation of the binding
   grammar; `TestBindTextDelegatesToBindTextSpans` asserts they agree on all 9 inputs (5
   successful, 4 errors, both polarities non-zero).
3. The breaking API takes atomic spans as a **parameter**. `internal/text` gained **no new import**
   — it still imports only the standard library and `textshape/ot` (D-000.16 holds).
4. **AD-25 amended in `ARCHITECTURE-SPINE.md`** (D-000.6): the Rule's mechanism list goes from two
   constraints to three. **Binds and Prevents are untouched.** Verbatim before/after below.
5. `folio-format.md`: the `unbreakableValues` field-table row, the document example, and a new
   **Line breaking** section stating the narrowings.
6. `folio-go/README.md`: AC8's disclosed limitation — declared values are protected, a Thai name in
   free-form text is not.

AC7's guard asserts **both polarities on every case**: the undeclared form is first proved to carry
an interior break in the span, then the declared form is proved not to. The two-placeholder case is
the red-proof for "marked only the first span atomic".

### Task 7 (partial) — measurement and slicing (AC9, AC10); wrapping core (AC11)

`folio-go/wrap.go`. `splitByFace` was refactored into `shapeSegments` + `positionSegments` so the
line breaker measures and slices **the same shaped glyphs that are drawn** — no re-shaping, no
second advance path. The refactor is byte-neutral: every existing golden still passes.

- **AC9 red-proof verified by mutation.** Reintroducing the per-rune `hmtx` path makes `"AV "`
  measure **23,984** and `"Wo. "` **33,008** — the naive values — against the correct shaped
  **23,344** / **32,688**. Expected values are literals hand-derived from hb-shape 14.2.0 against
  `fonts/notosans/NotoSans-Regular.ttf` at 16 pt, not recomputed the way production computes them.
  `"Ada "` is the negative control (33,200 under both hypotheses).
- **AC10 red-proof verified by mutation — and the first version of the guard FAILED TO REDDEN.**
  Reading `Cluster` as a byte offset left the additivity assertion green, because every boundary
  past rune 1 collapses to all-glyphs/no-glyphs and `whole + 0 == whole`. **A conservation
  assertion cannot see a degenerate slice.** The guard was strengthened to assert the boundary
  indices themselves as literals hand-derived from `"ณัฐวุฒิ"`'s cluster vector `[0 0 2 3 3 5 5]`
  (`0,2,2,3,5,5,7,7`); the mutation then reddens at rune 1 and at every boundary after it. Recorded
  because a red-proof that does not redden is a finding, not a formality.

`packLines` implements AC11's overflow rule: when not even the first opportunity fits, the line
takes it and **overflows visibly** — not re-broken, not squeezed, not dropped.

### Amendments executed

**`epics.md:827` (D-2.4.4, under D-000.6).**
Before: `**Then** breaks occur at whitespace per UAX #14`
After: `**Then** breaks occur at whitespace, and at script-appropriate opportunities in every other
script`

**`ARCHITECTURE-SPINE.md`, AD-25 Rule (D-2.1.10, under D-000.6).**
Before: `- **Rule:** two constraints sit **under** whatever the dictionary proposes, and both
override it:` followed by *Unknown runs are atomic* and *No break inside a Thai character cluster*.
After: `- **Rule:** three constraints sit **under** whatever the dictionary proposes, and all
override it:` followed by the same two, plus **A declared value is never split** (the
`unbreakableValues` mechanism, its reason, and "reaches the breaking stage as a **parameter**,
never through an import"). **Binds and Prevents unchanged, verbatim.**

### Task 7 (completed) — line emission, on the ruled leading

`lineAdvance` (wrap.go) is the ONE leading function (AD-2): the maximum, over the faces of the
declared chain present in the FontSet, of that face's `hhea` ascent − descent + lineGap, scaled to
the font size. `internal/fontset` gained `LineMetrics()`, which reads the hhea TABLE at
construction rather than through `(*ot.Face).Ascender`/`Descender`/`LineGap` — those substitute
800 / −200 / 0 for an absent table, and D-2.4.2 constraint 2 forbids inheriting that.

Red-proofs, all verified by mutation and restored:

- **first-face instead of maximum** → the shipped chain produces **21,792 mp** (= 1362 × 16), the
  first face's leading, and the test fails naming the hypothesis. The Thai-first ordering is the
  negative control and still passes under the mutation, so the test is not rejecting every answer.
- **routing through the substituting accessors** is asserted against by value: 800 − (−200) = 1000
  units is named in the test, so a face whose hhea went missing cannot pass as plausible.
- **content dependence** is asserted absent directly: the same chain and size yield the same
  advance across five very different contents, and the advance still scales with size.

`collectTextRuns` now shapes each element ONCE, computes its opportunities and atomic spans, packs
lines, and emits each line at `elementY + i × advance`. **The leading is computed once per element,
outside the line loop**, because it is a function of the chain and the size and of nothing on any
individual line.

### Task 8 — the S4 expected-break fixture (AC14) and its separately-bound sign-off (D-2.4.3)

`fixtures/expected-breaks/expected_breaks.json`: **25 items, 9 expecting at least one break and 16
expecting none.** Ordering was Trap 1's: labels authored first, committed second, asserted third.

Each item carries a `words` field — the segmentation — and `expectedBreaks` is only its arithmetic.
`TestS4ExpectedBreaksAreLabelsNotEngineOutput` enforces that the numbers follow the words, so a
number cannot be moved without moving a word boundary a human can see. **There is no generator for
this file**; nothing in the repository can write it.

**Eight labels were WRONG on first authoring, and the correction is on the record rather than
quietly applied.** `thai-003`…`thai-010` were split into their historical morphemes — *school* as
hall + study, *train* as vehicle + fire — and each seam labelled a break. The engine proposed none,
and the engine was right: **all eight are single dictionary HEADWORDS**, measured directly against
the shipped trie, while `ประเทศไทย` and `เก็บเงิน` (which DO break) are not. **A break opportunity is
not an etymology.** The labels were corrected on evidence independent of the engine's output —
headword membership — and the eight now serve as the fixture's negative controls. This is not
"regenerating the fixture to make a test pass": the same correction would be right had the engine
agreed with the original labels.

*Red-proof:* inverting `cjk-001` to `[1 3]` reddens naming the item; restored.

**D-2.4.3's sign-off:** `folio-go/expected_breaks_signoff_matrix_test.go`, `//go:build matrix`,
bound to the sha256 of `expected_breaks.json` itself, **computed at runtime over the fixture's
bytes and compared against the digest the reader wrote into `break-signoff.json`**. The test carries
**no digest literal at all** — corrected at finish: the wording above previously read as a pinned
constant. The binding is the stronger of the two readings, since a runtime digest cannot go stale;
the value it currently computes is `a545e042…99324de` (`shasum -a 256`, re-measured at the finish
commit). It **fails today**, which is D-2.3.5's first condition — the obligation is a red gate,
not a log entry. `fixtures/shaped-text/thai-signoff.json` was not created, read, or extended, and
`5964aad0…c92e00f` is byte-unchanged: **a re-record of either invalidates exactly one.**

### Task 9 — the wrapped-text golden (AC15)

`fixtures/wrapped-text/`, recorded **after** its semantic acceptance step, never before.

| element | script | box | full text measures | lines |
|---|---|---|---|---|
| `e1` | Latin | 150 pt | 331,683 mp | **3** |
| `e2` | Thai | 150 pt | 192,830 mp | **2** |
| `e3` | CJK | 150 pt | 198,000 mp | **2** |
| `e4` | declared span | **20 pt** | 46,585 mp | **2** |

Every box is measured to be narrower than its text **before** any wrapping assertion runs.

**`e4` is the discriminating element, and its box is deliberately narrower (20,000 mp) than the
bound value measures (24,585 mp).** That is what makes the declaration load-bearing:

- **declared:** `ผู้รับ` / `ศรีสุข` — the surname whole, **overflowing its box visibly** (AC11).
- **undeclared**, same template with only the list removed: `ผู้รับ` / `ศรี` / `สุข` — the surname
  **split** at the seam between the two common dictionary words it is spelled with. D-2.1.6's
  worked case, reproduced in the golden.

Both polarities asserted. A width ≥ 24,585 mp would make the two identical and the element would
prove nothing; the README says so, in those terms.

*Observability red-proof:* the same assertions pointed at `fixtures/shaped-text/input.folio` report
**one line for all 7 elements** and (a) fails — so the multi-line counts here are a real signal.

Semantic properties asserted at recording, none deferred: 9 distinct baselines against 4 elements
(with a vacuity guard that fires if it ever reports 4 — an unwrapped document), leading 16,621 mp,
all three shipped faces embedded, no boundary inside a Thai cluster, no boundary inside a declared
span, and every over-box line proved to carry no interior break opportunity.

### Task 10 — `/ToUnicode` section sizes (AC17). **DW-14 NOT triggered.**

Measured on the produced bytes of the wrapped-text render: **28, 18, 38** — largest **38** against
the specification's cap of **100**. Baselines for comparison: `font-text` 25; `multi-script-fallback`
4, 1, 1; `shaped-text` 14, 7, 28.

**The fixture was not sized to duck the cap** — its box widths were chosen to force wrapping, and
the sections are what that produced; the largest is comparable to `shaped-text`'s existing 28. The
numbers are appended to DW-14, which stays open. `assertToUnicodeSectionsUnderCap` now measures this
on every run and fails with a **stop-and-escalate** message if any section exceeds 100, so the
trigger is a test rather than a reader's vigilance.

### Task 11 — matrix registration (AC16), and Task 13's matrix RUN

`wrapped-text` is the **sixth** registered document, in `matrixDocuments` **and** in
`matrix.yml`'s `docs="…"` line **and** in all four per-target `upload-artifact` paths.
`TestMatrixDocumentSlugsAreRegisteredInCI` reports *"6 documents registered in both … across 4
targets"*. It fired first — naming `wrapped-text` as present in one list and not the other — which
is AC16's red-proof, live. `go vet -tags=matrix ./...` compiles.

Its own feature guard, `requireWrappedTextIsWrapped`, runs on **every leg before any byte
comparison** and asserts more distinct baselines than text elements. Without it, four legs of a
document that never wrapped would agree perfectly and certify nothing — the same reasoning
`requireShapedTextIsShaped` rests on.

**THE FOUR-TARGET MATRIX WAS RUN IN THIS STORY** (D-000.4's named override for 2.4). All six
documents pass; `wrapped-text`:

```
darwin/arm64   3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
linux/amd64    3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
linux/arm64    3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
js/wasm        3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e  72743 bytes
```

All four agree with each other **and** with the recorded golden.

### Task 12 — DW-11 discharged in writing (AC18). **No items added; the count remains 2.**

Story 2.4 is inside DW-11's owner window, so it owes an answer rather than silence. The answer:
**none were found, and none were invented.** The dev agent had no access to a sourced, attested
register of Thai personal names, and D-000.17 — reinforced by D-2.1.15 Major 5's precedent, where
"a plausible surname" was demoted rather than counted — forbids manufacturing attestation to reach
a number. **`ดอเลาะ` and `แนแซ` remain the only two.** `corpus.json` and `cmd/gencorpus/main.go`
are byte-unchanged.

Recorded alongside it: closing P2 removes the **symptom** DW-11 tracked, not the **thinness**.
The new expected-breaks fixture exercises the same path from a second direction but is a
conformance fixture, not the corpus, contributes nothing to P6g, and adds no attested name.

### The guard audit D-000.34 now requires — applied to every guard the filter touches

Two findings, both reported rather than absorbed:

1. **`TestComputeBreaksNoBreakInsideCluster` was VACUOUS, and had been since before this story.**
   Its subject `เก็บ` is itself a dictionary entry, so the matcher swallows it whole and proposes
   **no** interior break — the loop asserting "no break at 1 or 2" never executed a single
   iteration. **Measured at `0266a86` with the filter disabled: also `breaks=[]`.** So this is a
   pre-existing weakness the audit found, not one the fix created. Re-pointed to `เก็บเงิน`
   ("save money"), measured to carry a real break at rune 4 and **four** interior positions that are
   not cluster boundaries. The forbidden set is now computed from `ClusterBoundaries` rather than
   hand-written, so it covers every non-boundary position by construction, and both vacuity
   polarities are guarded.
2. **`TestP1NeverBreaksInsideCluster` had no vacuity guard**, and the filter shrank the population
   it sweeps from 558 to 526. P1 is a negative assertion made once per proposed break, so its
   strength *is* that population: withdrawing them all would leave it iterating zero times and
   reporting "zero violations". It now fails if it sweeps fewer than 100, and logs the count
   (**526 across 165 of 243 items**).

3. **`TestP2NeverBreaksInsideUnknownRun` — ADDED AT FINISH, and the sweep's completeness claim
   corrected with it.** The audit above reported two findings and said it had covered *"every guard
   the filter touches"*. It had not: this one consumes constrained `ComputeBreaks` output and the
   filter touches it. It is **self-referential by construction** — its own doc comment says so, in
   terms: it checks the engine's own `RunUnknownThai` spans against the engine's own breaks, so it
   can never find the failure mode that matters, and after the filter its assertion is doubly
   implied. It is **retained deliberately** on that basis, and it is **not** counted as P2's
   measurement; `TestP2IndependentDPCrossCheck` is. What it lacked, and its P1 sibling had just been
   given, was a vacuity floor. It now has one, **derived from a measurement rather than guessed**:
   66 `RunUnknownThai` spans and 85 (run, break) pairs across 243 items at the finish commit, with
   the floor set at 15 — the same fraction-of-measured shape P1's 100-against-526 uses. A first
   draft of that floor carried an invented figure and failed loudly on its first run, which is the
   floor working.

The V11 sample re-pointing recorded at the parking point is the fourth instance, and was the one that
prompted the rule.

### Measured gates at completion

| gate | invocation | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **463 PASS · 1 FAIL** |
| `folio-go/` | the same, top-level only | **292 PASS · 1 FAIL** |
| the one failure | — | `TestCorpusMeetsP6ExerciseFloors` (P6g 7, need ≥ 20) — **AC5's preserved shortfall**. `TestP2IndependentDPCrossCheck` is **GREEN**. |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **81 PASS · 0 FAIL** |
| `lint/` | the same, top-level only | **43 PASS · 0 FAIL** |
| modules | `GOWORK=off go list -m all` | exactly **two** |
| matrix | `go test -tags=matrix -run TestCrossTargetByteIdentity` | **6 documents × 4 targets, all agreeing, all matching their goldens** |
| `go vet -tags=matrix ./...` | — | compiles |
| fixtures moved | `git diff --stat -- fixtures/` | **`computed_breaks.json` only** |
| existing goldens | per-directory `git diff --quiet` | `minimal-rect`, `font-text`, `image-embed`, `multi-script-fallback`, `shaped-text` all **byte-identical** |
| `5964aad0…c92e00f` | `fixtures/shaped-text/expected.json` | **unchanged**; `thai-signoff.json` still **absent** |

The baseline had **two** failures; there is now **one**, and it is the one AC5 requires be left
failing.

**One inherited guard had to be updated, and it fired correctly first.** Story 2.3a's
`TestStory23aAddedNoThirdEpic2GateObligation` holds an inventory of matrix-tagged files and failed
naming `expected_breaks_signoff_matrix_test.go` as *"a THIRD Epic 2 gate obligation… a third may not
be added without saying so explicitly."* **D-2.4.3 says so explicitly**, so the inventory was
extended with each file annotated by the story and the ruling that authorised it, and the guard's
message now catches a **FOURTH** — quoting D-2.4.3's own "do not add a fourth". Part (a) of that
test, which asserts `thai-signoff.json` is absent, is untouched and still passes.

**A lint regression I introduced and fixed, recorded rather than quietly corrected:** the first
version of `Opportunities` ranged a map to collect its results, which took `lint` to **79/2** and
**41/2** (`map-range`, D-1.3.5/NFR1.d, firing on `internal/text/opportunity.go`). The collection is
now an ascending index walk. `lint` is back at 81/0 and 43/0. The rule is syntactic on purpose and
did not care that the range was followed by a sort.

---

## Delivery Log

### D-2.4.2's leading measurement — the numbers, and why exactly one metric survives

Measured directly from the tables (`font.TableData` + `ot.ParseHead`/`ParseHhea`/`ParseOS2`), which
is Story 2.2's `readPostScriptName` pattern applied a fourth time: **read the table, take the
integer, decline the accessor**. The accessors substitute — `(*ot.Face).Ascender()` returns a
hard-coded `800` when `hhea` is nil, and `CapHeight()` falls back to `Ascender()` — and constraint 2
forbids inheriting that class of fiction.

Subjects: the three shipped faces as committed under `folio-go/fonts/`, plus
`testdata/fonts/Roboto-Regular.ttf`. All values scaled to the 1000-unit em.

| face | upem | `hhea` A / D / gap | A−D(+gap) | `OS/2` typo A / D / gap | A−D(+gap) | head bbox ink | `fsSelection` | USE_TYPO_METRICS |
|---|---|---|---|---|---|---|---|---|
| Noto Sans | 1000 | 1069 / −293 / 0 | **1362** | 1069 / −293 / 0 | 1362 | 1575 | 0x00c0 | **true** |
| Noto Sans Thai | 1000 | 1061 / −450 / 0 | **1511** | 1061 / −450 / 0 | 1511 | 1442 | 0x00c0 | **true** |
| Noto Sans SC | 1000 | 1160 / −288 / 0 | **1448** | **880 / −120 / 0** | **1000** | 2856 | 0x0040 | **false** |
| Roboto (testdata) | 2048 | 1900 / −500 / 0 → 928 / −244 / 0 | **1172** | 1536 / −512 / 102 → 750 / −250 / 50 | 1000 (1050) | 1327 | 0x0040 | **false** |

**Candidate B — `OS/2` typo metrics: ELIMINATED on constraint 2, by measurement.** Noto Sans SC
declares `USE_TYPO_METRICS = false`, i.e. the face itself says its `sTypo*` values are **not** its
line metrics — yet they are perfectly plausible numbers. Consuming them yields **1000/em against a
true 1448**, which is *below the face's own declared ascent plus descent*: CJK lines would collide.
That is exactly the `/CapHeight 928 against a true 711` shape D-2.4.2 constraint 2 names. Two
further strikes: the vendor exposes **no accessor at all** for the typo fields (there is no
`TypoAscender()` on `ot.Face`), so reading them needs a fresh raw parse *plus* an `fsSelection`
bit-7 test to be honest — widening a surface Story 2.3a has just audited and fenced.

**Candidate C — a size multiple: ELIMINATED on measurement.** 1.2 em = 1200 is **below all three**
shipped faces (1362, 1511, 1448). Even 1.5 em = 1500 is below Noto Sans Thai's 1511. Clearing the
shipped set needs ≥ 1.511 em — a constant fitted to the faces that happen to ship today, which
silently breaks on the first caller-supplied face. It also discards the "declared font chain" half
of constraint 1's stated input.

**Candidate A — `hhea` ascent − descent (+ lineGap): SURVIVES.** `hhea` is in
`requiredTables` (`internal/fontset/fontset.go:273`), validated at ingestion with a located error
naming face and table, so no substituted default is reachable. `lineGap` is **0 on all three
shipped faces**, so including it is byte-neutral for the shipped set and honest for a face that
declares one. It is a function of (chain, size) and of nothing on the line. No new table, no new
accessor, no float.

**Exactly one METRIC survives, as D-2.4.2 hoped. But a second question is left open by it, and the
measurement contradicts the story's own recommendation — so it is escalated, not chosen.** See
below.

### D-000.4's matrix override — NOT waived

ESC-1 asked for a logged waiver of D-000.4's named per-story override for 2.4. **There was never a
waiver to log — the commissioning brief was simply wrong**, and D-000.4 names `2.4` explicitly
(*"line breaking feeds every measurement"*). The full four-target matrix therefore runs **in this
story**, per the correcting brief. Logged here as D-000.4 requires of an override, with its reason:
**line breaking feeds every measurement**, and this story's own deliverable is hash-shaped.

> **Amended at finish (2026-08-24).** The sentence that stood here — *"It has not run yet — it runs
> once the wrapped-text golden exists, which is behind the open decision"* — was true when it was
> written and is not the outcome. **The override is DISCHARGED IN-STORY.** The golden exists, the
> matrix ran here (Task 13), and the reviewer deleted `.matrix-build/` entirely and re-ran it from
> scratch: **6 documents × 4 targets, all agreeing, all matching their goldens**, with `wrapped-text`
> at `3845da37…a712288e`, 72,743 bytes, on darwin/arm64, linux/amd64, linux/arm64 and js/wasm. It is
> left as an amendment rather than a rewrite so the reader can see that the pending state was real
> and was closed, not that it never existed (D-000.28).

### Golden immunity — the scope of the claim, stated

M4's "no existing golden can move" was measured over the **five golden fixtures** only (ten text
elements, all fitting, the widest at 26% of its box). It did **NOT** cover templates embedded as Go
string literals in `render_test.go`, `render_bind_test.go`, `shaped_fixture_test.go` and their
siblings, which a fixture sweep cannot see. Measured outcome at the parking point: **no golden and
no in-test literal template moved** — the whole suite is green but for AC5's disclosed floor.

**One in-test Go literal DID have to change, and it is not a golden.**
`TestUnconstrainedVsConstrainedSwitchActuallyToggles` (V11) used `"ดอเลาะ"` as its discriminating
sample, chosen because the **constrained** engine proposed a spurious interior break there that the
unconstrained one did not. That spurious break **was the P2 defect**, and this story's filter
withdrew it — so both modes now return `[]` and the sample discriminates nothing. This is D-000.30's
phenomenon exactly: closing a defect destroys evidence that depended on it. Measured replacement:
across the corpus, **23 of 243** items still discriminate, but **no single-span (space-free) corpus
item does**. `"ชัยวัฒน์"` — the uncoverable `[0,8)` span of `name-021`/`name-081` — measures
unconstrained `[3 7]` against constrained `[]`, is single-span, and is a *stronger* discriminator
than the one it replaces. The test now documents why the sample moved, and why that is not a
regression.

### Finish (2026-08-24) — what the finisher changed, and the gates it measured

**Scope of the finish commit: `_test.go` files and documents only.** No production file was
modified. `wrap.go` and `internal/fontset/fontset.go` were each mutated for a red-proof and restored
by hand from a pre-mutation copy, then verified byte-identical with `/usr/bin/diff` — never
`git checkout`. Because no production byte moved, **the four-target matrix result cannot have moved
either**, and it was not re-run at finish; it was run in-story (Task 13) and re-run from scratch by
the reviewer. Re-measured at the finish commit as the standing check on that claim:
`fixtures/wrapped-text/expected.pdf` is still `3845da37…a712288e`.

**Gates, each with its invocation and its counting rule (D-000.26).** Run from
`/Users/panitw/Projects/folio/folio-go` and `/Users/panitw/Projects/folio/lint`:

| gate | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` occurrence | **464 PASS · 1 FAIL** |
| `folio-go/` | the same, counting only top-level (`^--- PASS` at column 0) | **293 PASS · 1 FAIL** |
| the one failure | — | `TestCorpusMeetsP6ExerciseFloors` (P6g 7, need ≥ 20) — **AC5's preserved shortfall**. It is the only failure. |
| `TestP2IndependentDPCrossCheck` | the same run | **GREEN** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **81 PASS · 0 FAIL** |
| `lint/` | the same, top-level only | **43 PASS · 0 FAIL** |
| modules | `GOWORK=off go list -m all` | exactly **two** |
| `go vet ./...` / `go vet -tags=matrix ./...` | — | both exit 0 |
| `gofmt -l folio-go/` | — | empty |
| goldens | `shasum -a 256` | `shaped-text/expected.pdf` **`5964aad0…c92e00f`** (unchanged); `expected-breaks/expected_breaks.json` **`a545e042…99324de`**; `wrapped-text/expected.pdf` **`3845da37…a712288e`** |
| `thai-signoff.json` | `find fixtures -name '*signoff*'` | **absent**, as it must be |
| fixtures moved | `git status --porcelain fixtures/` | **`computed_breaks.json` only**, plus this story's two NEW fixture directories |

**The count moved by exactly +1 PASS against the review's 463/292**, and that one is
`TestLineMetricsReadsTheHheaTableNotTheVendorAccessors` — the new guard below. Nothing else changed
count, and no second failure was introduced.

**Both Thai sign-offs remain PENDING at the Epic 2 gate**, deliberately, each as its own red
`//go:build matrix` test bound to its own digest: `TestShapedTextThaiSemanticSignOffIsRecorded`
(reading, against `5964aad0…c92e00f`) and `TestExpectedBreaksHumanSignOffIsRecorded` (breaks,
against `expected_breaks.json`'s runtime-computed digest). **The gate owes three things and was not
given a fourth**; the inventory guard in `byte_neutrality_test.go` still catches a fourth.

### Finish — Minor B's guard, and the red-proof it now carries

D-2.4.2 constraint 2 was held **by code reading, not by the build**: routing `LineMetrics` through
`ot/metrics.go:434,442,500` (the substituting accessors, 800 / −200 / 0 on a nil `hhea`) left the
**entire suite green**. Reproduced at finish before fixing, and the reason is now stated in the code
rather than left to be rediscovered: `requireReadableTables` makes an absent `hhea` a load error, so
no `*Font` whose accessors substitute can exist — and for a face whose `hhea` **is** present the
accessors return the table's own numbers. The two readings are indistinguishable on every
constructible face, which is why every existing assertion survived.

`internal/fontset.TestLineMetricsReadsTheHheaTableNotTheVendorAccessors` asserts the property where
the readings *can* be told apart, in two legs:

- **leg 1, anchored to the artifact** — the three numbers are read out of the font file's own `hhea`
  table bytes (ascender at offset 4, descender at 6, lineGap at 8, located through the sfnt table
  directory) and `LineMetrics()` must reproduce them scaled. Measured on
  `testdata/fonts/Roboto-Regular.ttf`: `{Ascent:928 Descent:-244 LineGap:0}`. Leg 1b records, as a
  measurement rather than an assumption, that the accessor reading **agrees** here — the fact that
  makes leg 2 necessary.
- **leg 2, the discriminator** — the table-derived fields are overwritten with sentinels
  (1234 / −321 / 77, asserted distinct from both the face's real values and the substituted
  800 / −200 / 0) and `LineMetrics()` must follow them.

**Red-proof, run:** with `LineMetrics` rerouted to `scale(f.face.Ascender())` /
`Descender()` / `LineGap()`, the full suite goes to **463 PASS · 2 FAIL** — the AC5 shortfall plus
this guard, which fails naming the hypothesis (*"If it now routes through
(\*ot.Face).Ascender/Descender/LineGap, it has inherited the vendor's substituting accessors and
D-2.4.2 constraint 2 is broken"*). `fontset.go` restored by hand and verified identical.

**A blanket source rule would have been wrong, and was declined for a measured reason.** The
reviewer offered *"an AST/source assertion that Ascender()/Descender()/LineGap() appear nowhere
under `folio-go/`"* as one option. `fontset.go:637-638` uses `f.face.Ascender()` and
`f.face.Descender()` legitimately, on the PDF `FontDescriptor` path, where the FontDescriptor's own
semantics are what the accessor returns. The rule is about **leading**, not about the accessors as
such, so it is asserted behaviourally at `LineMetrics` and not by grepping the module.

**The self-referential precondition is also gone.** `TestLineAdvanceDeclinesTheSubstitutingAccessors`
looped over the same hard-coded `perFace` map that supplied `want`, so `differing == 0` was
impossible by construction. It now inspects the values `lineAdvance` actually returned (measured:
`Noto Sans 21792`, `Noto Sans Thai 24176`, `Noto Sans SC 23168` mp at 16 pt; **3 of 3**
distinguishable from the substituted 16,000). That test's `substitutedUnits` branch is now
**labelled an unreachable vendor-contract forward guard** in the wording `fontset.go` already uses
for `ot.NewFace`, per D-000.24 — it is not counted as constraint 2's coverage, and the doc comment
names where the reachable guard lives.

### Finish — the conservation sweep (Minor A), enumerated

D-000.23's rule, applied: *a guard written in response to a defect covers the defect, not its class.*
D-000.33 was closed for AC10's guard and not for its sibling. The sweep below is an **enumeration,
not a sample** (the D-2.2.4-correction precedent).

**Method.** Two mechanical greps over **all 31 `.go` files this story touches** (the full
`git status --porcelain` set) — one for sum-vs-whole comparisons, one for `+=` accumulators later
compared — then each hit read in context. The class swept: *an assertion that a quantity over a
whole equals the combination of that quantity over the parts of a partition*, which a degenerate
partition satisfies trivially.

**Seven sites found. Five lacked a non-degeneracy precondition; all five now have one. Two already
carried an adequate one and were left alone.**

| # | site | shape | guard before | action at finish |
|---|---|---|---|---|
| 1 | `wrap_test.go` `TestMeasureIsAdditiveAcrossASplit` | `left + consumed + right == whole` | **none** | **FIXED** — per-split non-degeneracy on the *measured width*, N-of-N witness across all 5 subjects, plus an interior-split precondition (below) |
| 2 | `wrap_test.go` `TestSliceAtRuneBoundaryUsesRuneIndices`, the retained additivity leg | `left + right == whole` for `r ∈ [0,n]` | **none** (r=0 and r=n are degenerate and swept) | **FIXED** — counts non-degenerate splits, fatals at zero |
| 3 | `wrap_test.go` `TestPackLinesWrapsWithinTheDeclaredWidth`, the line-reconstruction block | `first.from == 0`, `last.to == n`, no overlap | **partial** — `len(lines) >= 2` guards the aggregate, per-line emptiness unguarded | **FIXED** — every line asserted non-empty before the coverage checks |
| 4 | `internal/text/s4_expected_test.go`, `concat(words) == text` | concatenation reconstructs the whole | **none** — a one-word item is `whole == whole` | **FIXED** — at least one item must carry ≥ 2 words (**measured: 9 of 25**) |
| 5 | `internal/text/s4_expected_test.go`, `expectedBreaks == cumulative rune lengths of words[:-1]` | same partition, one loop later | **none** | **FIXED by the same precondition** — a single-word item derives an empty list, so #4's guard is exactly what this leg needed |
| 6 | `wrapped_text_fixture_test.go` (a), `len(ys) == Σ per-element line counts` | sum of parts equals whole | **adequate** — `len(ys) <= len(expectedLines)` fatals, i.e. "nothing wrapped" is caught | **no change**, recorded as already-guarded |
| 7 | `internal/text/tileable_test.go` `TestFilterOnlyRemovesBreakOpportunities` | `kept ⊆ proposals`, i.e. `removed + kept == proposals` | **adequate** — both `checked == 0` and `removed == 0` fatal | **no change**, recorded as already-guarded |

**Swept and excluded, with the reason.** Six `+=` accumulators are production summations with no
assertion attached — `wrap.go:101` (`advance1000`), `wrap.go:126` (`measureRuneRange`),
`render.go:544` (the segment cursor), `internal/bind/text.go:117` (`runesWritten`),
`internal/text/tileable.go:120`, `internal/text/trie.go:57`/`:61` — and
`internal/text/corpus_test.go:273` (`stats.P6e`) is a corpus statistic. None is a conservation
*assertion*, so none is in the class.

**THE REVIEWER'S SUGGESTED RESOLUTION, APPLIED LITERALLY, DOES NOT REDDEN — measured, and this is
the finding's one factual correction.** Finding 1 proposed *"require `left > 0 && right > 0` for at
least one split per subject"*. That was written and run against the byte-offset mutation of
`glyphRangeForRunes`, and `TestMeasureIsAdditiveAcrossASplit` **stayed green, 4 of 4 subjects
witnessed**. Two separate reasons, both measured:

1. **Additivity cannot see a shifted-but-conserving boundary at all.** It is preserved by *any*
   monotone boundary function: a rule that cuts at the wrong glyph but puts both sides' glyphs
   somewhere still sums to the whole. `"结算单，共３页"` under the mutation cuts at glyph 3 instead of
   glyph 1 — wrong, non-degenerate, and additive. Only a **saturating** boundary (one side collapsing
   to no glyphs) is visible to it, and that is exactly the `whole + 0 == whole` hole D-000.33 names.
2. **No split in the original subject set fell inside a face segment.** Probed:
   `"ณัฐวุฒิ เกิด กรุงเทพ"` — the subject the test's own comment called the discriminating one —
   segments as `[Thai 0,7) [Latin 7,8) [Thai 8,12) [Latin 12,13) [Thai 13,20)` and breaks at **7 and
   12**, which are *exactly* the segment boundaries. `GlyphInfo.Cluster` is segment-local, so at a
   segment boundary the rune-index and byte-offset readings both read 0 and coincide. Every split in
   the table was of that kind. **Interior splits: 0 of 4 for that subject.**

**So the fix is the guard plus a subject that gives it somewhere to fire.** `"ณัฐวุฒิเกิดกรุงเทพ"` is one
whitespace-free Thai face segment breaking at runes 7 and 11, strictly interior. A second
precondition now fatals if *no* split in the whole table falls inside a face segment, naming why.

**Red-proof, run and recorded.** Green at head: *"additive across all 10 break opportunities in 5
subjects; 10 splits are NON-DEGENERATE, witnessed in 5 of 5 subjects; 7 splits fall strictly inside a
face segment."* Under the byte-offset mutation of `glyphRangeForRunes` the strengthened test now
**FAILS**, naming the subject and the hypothesis: *"every one of `"ณัฐวุฒิเกิดกรุงเทพ"`'s 4 splits is
DEGENERATE (a side measures 0 millipoints), so for this subject the check reduces to whole+0 ==
whole"*. Whole-suite selection under that mutation: **460 PASS · 5 FAIL** —
`TestMeasureIsAdditiveAcrossASplit` (**new**), `TestSliceAtRuneBoundaryUsesRuneIndices`,
`TestWrappedTextSemanticAcceptance`, `TestWrappedTextGoldenFixture`, and AC5's standing shortfall.
`wrap.go` restored by hand and verified identical with `/usr/bin/diff`.

**D-000.34 applied to the finisher's own fix, and it found one.** `TestSliceAtRuneBoundaryUsesRuneIndices`'s
red-proof paragraph claimed the byte-offset mutation was caught *"by the additivity assertion
below"* — while the paragraph beside its boundary table already recorded that additivity **survived**
that mutation. The two could not both be true; the measurement says the boundary table catches it
and the additivity leg does not. The doc comment now says so, and says which paragraph was wrong.

### Finish — the third minor and the four nits

- **Finding 3 (Minor, stale doc comment).** `byte_neutrality_test.go`'s comment said the gate owes
  *"exactly two things"* while the inventory and failure message beneath it said three/fourth. The
  comment now enumerates all three obligations by ruling, names both `//go:build matrix` sign-off
  tests, and scopes the D-000.4 cadence sentence to **Story 2.3a** — contrasting it with 2.4, which
  registers a leg *and* runs the matrix. The function's own name (`…AddedNoThirdEpic2GateObligation`)
  is 2.3a's and was deliberately left: renaming it would break the citations that point at it, so the
  comment says outright that the name is 2.3a's and the count is not.
- **Nit 4 (`chain` vs `stack` misquote).** Confirmed: `folio-format.md:263` says *"**A stack**
  declares what may appear in an element…"* and D-2.4.2 (resolved) and D-2.4.6 both block-quote it as
  *"**A chain** declares…"*. The decision log is **append-only by its own header** (*"a reversal is
  appended, never a rewrite"*), so the two entries were **not** edited; the correction is recorded as
  **D-2.4.7**, which prints the canonical sentence verbatim. The code/spec terminology split
  (`chain` in Go, `stack` in the spec) is real and is **DEFERRED** — see below.
- **Nit 5 (`epics.md:823` → `:827`).** Verified at `0266a86` and at head: the UAX #14 clause is line
  **827**, and the amendment is exactly one line (`@@ -824,7 +824,7 @@`). Four citations corrected in
  this file; the decision log's two are recorded in D-2.4.7 rather than rewritten.
- **Nit 6 (digest described as a literal).** Reworded. The sign-off computes `sha256.Sum256` over the
  fixture bytes at runtime and carries no digest literal; re-measured with `shasum -a 256` at the
  finish commit, the value is `a545e042…99324de`.
- **Nit 7 (a third guard the sweep did not name).** Both halves of the reviewer's option taken: it is
  now **named in the D-000.34 audit** as a deliberately retained self-referential guard, *and* given
  P1's vacuity floor so it cannot decay to zero iterations unnoticed. Floor derived from a
  measurement, not chosen: **66 `RunUnknownThai` spans, 85 (run, break) pairs, across 243 items**,
  floor **15**.

### Finish — deferred, with the follow-up named

**The `chain` / `stack` terminology split.** The Go code says `chain` (`fontChain`,
`lineAdvance(chain …)`, `atomicSpansFor`'s callers) and `folio-format.md` says `stack`, consistently,
throughout. Unifying them means either renaming a public-facing spec term or renaming identifiers
across `wrap.go`, `render.go` and their tests — neither belongs in a finish commit whose scope is
otherwise `_test.go` and documents, and either would touch code the matrix has already certified.
**Follow-up:** pick one term and align both sides, in a story that can re-run the matrix behind it.
Recorded in D-2.4.7.

---

## QA Results

## Review Summary
- Reviewed by: bmad-code-reviewer
- Date: 2026-08-24
- Story Status Recommendation: **Approved with Minor Notes**
- Blockers: 0
- Majors: 0
- Minors: 3
- Nits: 4

Every load-bearing claim was **re-constructed independently**, not read. The six items flagged as
highest-risk all verified clean, including the two that a reviewer would most expect to fail: the
P2 structural argument and the eight hand-corrected S4 labels. All mutations were reverted by hand
(`python3` string replacement + `diff` against a pre-mutation copy); `git checkout`/`git stash` were
never used, and the final `git status --porcelain` is byte-identical to the tree handed over
(27 modified, 16 untracked).

### Gates re-measured by the reviewer

| gate | reviewer's own measurement | claimed | verdict |
|---|---|---|---|
| `folio-go/` all occurrences | **463 PASS · 1 FAIL** | 463 / 1 | match |
| `folio-go/` top-level | **292 PASS · 1 FAIL** | 292 / 1 | match |
| the one failure | `TestCorpusMeetsP6ExerciseFloors` (P6g 7, need ≥ 20) | AC5's preserved shortfall | match |
| `TestP2IndependentDPCrossCheck` | **GREEN** | GREEN | match |
| `lint/` | **81 / 0** and **43 / 0** | 81/0, 43/0 | exactly baseline |
| `GOWORK=off go list -m all` | exactly **two** modules | two | match |
| `go vet ./...` and `go vet -tags=matrix ./...` | both exit 0 | compiles | match |
| `gofmt -l` | empty | — | clean |

### 1. The P2 fix and its standing red-proof — VERIFIED BY CONSTRUCTION

Disabling `filterBothSidesCoverable` in `ComputeBreaks` reproduced **26 violations across 17 items**,
matching M2's table *item for item and position for position*, including **`name-116` (`ดอเลาะ`) at
rune 11** and **`name-117` (`แนแซ`) at rune 8**. Restored; P2 green again.

**The structural argument was attacked, not just the number.** Its load-bearing premise is that
`tileableForward`/`tileableBackward` are sound at the *interior* offsets the filter consults — the
committed `TestTileabilityAgreesWithIndependentDP` only checks the two whole-span endpoints
(`fwd[n]`, `bwd[0]`), so a function wrong in the middle would pass it. Probed directly against
`isFullySegmentable` at **every** interior offset of every Thai span: **3,361 offsets, 0 mismatches**.
The engine and `TestP2IndependentDPCrossCheck` also share the *same* `scriptSpans` decomposition, so
the span the filter operates on is the span P2 governs — closing the one gap that would have broken
the concatenation argument.

Fuzzed beyond the corpus for the "over ANY input" claim: **200,000 random Thai strings**, of which
**198,324** contained an untileable span; the engine proposed **233** breaks overall and **0** interior
to an untileable span. The argument holds; no counterexample exists.

### 2. The leading rule (D-2.4.2) — VERIFIED, BOTH HALVES

`LineMetrics()` reads `ot.TagHhea` via `TableData` + `ParseHhea` at construction. The vendor claim was
checked against textshape v0.0.15 itself: `ot/metrics.go:434,442,500` return hard-coded **800 / −200 / 0**
when `f.hhea == nil` — exactly the substitution class 2.3a audited. `lineAdvance` is the maximum over
the declared chain, per D-2.4.2 (resolved).

Both red-proof halves reproduced by mutation:
- first-face instead of maximum → **21,792 mp (= 1362 × 16)**, failing and *naming the hypothesis*;
  the CJK-first case additionally reddens at 23,168 (1448 × 16).
- the **Thai-first negative control PASSES** under the same mutation — the test is not rejecting
  every answer. See Finding 2 for the one gap here.

### 3. The four-target matrix — RE-RUN FROM SCRATCH

`folio-go/.matrix-build/` (gitignored) was **deleted entirely** and the matrix re-run, so nothing could
be reused. Result: **6 documents × 4 targets, all agreeing, all matching their goldens**;
`wrapped-text` at **`3845da37ae198beae3d3ef98211678b02a397a87336cea025e2e8286a712288e`, 72,743 bytes**
on darwin/arm64, linux/amd64, linux/arm64 and js/wasm.

The prior-story instrument failure was specifically checked for and is **absent**: the binaries were
rebuilt during this run (timestamps seconds before the run) and `file` reports genuine
**ELF 64-bit x86-64**, **ELF 64-bit ARM aarch64**, Mach-O arm64 and WebAssembly. `runOnTarget`'s
docker/wasm paths are `t.Fatalf` on every failure — never a skip — and `matrix.yml` uses
`if-no-files-found: error`, so a silently-lost leg is not reachable. Registration is complete on both
sides (all four `upload-artifact` paths **and** the `docs="…"` line).

*Note: `.matrix-build/diverged/` held mid-story dumps timestamped 09:54–09:55 against 17:50 binaries —
stale intermediates, not evidence of a divergence in the final run. The directory is gitignored and
was cleared by this review.*

### 4. The eight corrected S4 labels — VERIFIED INDEPENDENTLY OF THE ENGINE

Each of the eight was measured against the shipped trie with `BytesTrie.Contains` — a dictionary
lookup that shares no code path with break computation:

| id | text | `Contains` (headword) | expected |
|---|---|---|---|
| `thai-003`…`thai-010` | `นักเรียน` `โรงเรียน` `โรงพยาบาล` `รถไฟ` `หนังสือพิมพ์` `วันเกิด` `ที่อยู่` `ใบเสร็จ` | **true (all eight)** | `[]` |
| `thai-001` | `ประเทศไทย` | **false** | `[6]` |
| `thai-002` | `เก็บเงิน` | **false** | `[4]` |

The contrast is real and the correction rests on **headword membership, not engine output** — so this
is not "the fixture changed and then the test passed". All 25 items' `expectedBreaks` were
independently re-derived from their `words` and agree exactly (9 items expecting breaks, 16 expecting
none). The correction is recorded in both `fixtures/expected-breaks/README.md:31-40` and the fixture's
own `_README` field. `TestS4ExpectedBreaksMatchTheEngine` asserts against `Opportunities` (the unified
API), so the CJK half is genuinely exercised, and carries both-polarity vacuity guards.

### 5. Vacuity audit (D-000.34) — the pre-dating claim CONFIRMED, and swept further

With the filter **disabled** (i.e. the pre-fix engine), `ComputeBreaks(dict, "เก็บ", false)` returns
**`[]`**, and `เก็บ` is itself a dictionary headword (`Contains` = true). The old
`TestComputeBreaksNoBreakInsideCluster` loop therefore executed **zero iterations at `0266a86` as
well** — the vacuity genuinely pre-dates the fix and was not created by it. Both re-pointed guards now
carry both-polarity vacuity guards, and P1's floor (100) sits well below its measured 526 without
duplicating AC10's exact pin.

Swept every test consuming constrained `ComputeBreaks` output (14 call sites). All carry adequate
guards except one pre-existing case — see Nit 4.

### 6. D-000.33 on the new assertions — AC10 CONFIRMED, one sibling still holed

Mutating `glyphRangeForRunes` to read `Cluster` as a byte offset makes
`TestSliceAtRuneBoundaryUsesRuneIndices` **redden at rune 1 and at every boundary after it** (5
failures), exactly as recorded. Under that *same* mutation the sibling conservation test stays green —
see Finding 1.

### Verified, not treated as findings

- **DW-11 discharged with nothing fabricated.** The count remains 2 (`ดอเลาะ`, `แนแซ`). `corpus.json`
  and `cmd/gencorpus/main.go` are byte-unchanged. No new Thai personal name appears anywhere in the
  diff; the four names in the new S4 fixture all already existed at `0266a86`.
- **AC17.** Measured **28, 18, 38** against the cap of 100. The fixture was **not** sized to duck the
  cap: it carries genuine 63/41/18-character paragraphs in 150 pt boxes, and its largest section (38)
  *exceeds* every pre-existing fixture's (`shaped-text`'s 28). `assertToUnicodeSectionsUnderCap`
  stops-and-escalates above 100.
- **The third gate obligation.** `fixtures/shaped-text/` is byte-unchanged; `thai-signoff.json` does
  not exist anywhere in the tree; `5964aad0…c92e00f` is unchanged **and** was independently confirmed
  to be the actual sha256 of `expected.pdf`. The new sign-off is `//go:build matrix`, binds the
  fixture's own digest (`a545e042…99324de`, confirmed by `shasum`), and never reads, creates or
  extends `thai-signoff.json`. 2.3a's inventory guard now carries per-file attribution and catches a
  **fourth**; its part (a) is untouched.
- **Only `computed_breaks.json` moved.** 558 → 526 (−32), **23** items changed, `synthetic-039`
  `[4]` → `[]`. Every other `expected.json` under `fixtures/` is byte-identical; no deletions.
- **No `float32`/`float64` on any production path**; no kinsoku denylist anywhere — CJK classification
  is `unicode.Is(unicode.Han/Hiragana/Katakana, r)`, a category test (D-000.23 honoured).
- **Canonical amendments.** AD-25's **Binds** and **Prevents** are absent from the diff entirely;
  `epics.md`'s amendment is exactly one line.

---

### Finding 1: A conservation assertion with no non-degeneracy guard — D-000.33 applied to AC10 but not to its sibling
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/wrap_test.go:111`–`:149` (`TestMeasureIsAdditiveAcrossASplit`)
- **Observation**: The story records (Task 7) that AC10's first guard failed to redden because
  `whole + 0 == whole` survives a degenerate slice, and strengthened *that* guard with literal
  boundary indices. The lesson was not carried back to the sibling test, which asserts only
  `left + consumed + right == whole` and never asserts that the partition is non-degenerate.
  Demonstrated: under the byte-offset mutation of `glyphRangeForRunes`,
  `TestSliceAtRuneBoundaryUsesRuneIndices` fails 5 times while
  `TestMeasureIsAdditiveAcrossASplit` **passes**, logging *"measurement is additive across all 8
  break opportunities in 4 subjects"* — including the Thai subject `"ณัฐวุฒิ เกิด กรุงเทพ"` that is
  supposed to be the discriminating one.
- **Impact**: Low today — AC10's strengthened guard covers the same defect. But this test's doc
  comment advertises it as catching *"if slicing lost or double-counted a glyph"*, which it cannot do
  once a boundary collapses to all-glyphs/no-glyphs. If AC10's guard were ever narrowed, the
  remaining "additivity" coverage would be worth nothing, silently. This is D-000.33's exact shape,
  closed for the instance but not for the class.
- **Suggested Resolution**: Add a non-degeneracy assertion inside the split loop — e.g. require
  `left > 0 && right > 0` for at least one split per subject, or count non-degenerate splits and fail
  if the count is zero — and note in the doc comment that additivity is necessary, not sufficient
  (the wording AC10's guard already carries).
- **Related AC**: AC9, AC10

### Finding 2: D-2.4.2 constraint 2 has no standing guard — the accessor-decline is unfalsifiable
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/wrap_test.go:555`–`:599`
  (`TestLineAdvanceDeclinesTheSubstitutingAccessors`); `folio-go/internal/fontset/fontset.go:973`–`:983`
- **Observation**: Mutating `LineMetrics()` to route through the substituting vendor accessors —
  `scale(f.face.Ascender())`, `scale(f.face.Descender())`, `scale(f.face.LineGap())` — leaves the
  **entire suite green** (only AC5's disclosed failure remains). The guard's discriminating branch
  compares against `substitutedUnits = 1000`, which is only reachable for a face whose `hhea` is
  absent, and `requireReadableTables` makes such a face a load error — so the branch cannot fire.
  Its accompanying vacuity precondition loops over the same hard-coded `perFace` map that supplies
  `want`, so `differing == 0` is impossible by construction and the precondition can never fail.
- **Impact**: The implementation is correct and `fontset.go`'s own comment is admirably honest about
  the branch being unreachable. But the Delivery Log presents this test as constraint 2's guard
  (*"asserted against by value"*), and a future refactor swapping to the accessors would pass every
  gate. The property is held by code reading, not by the build. Under D-000.24 an unreachable
  assertion should be labelled as such rather than counted as coverage.
- **Suggested Resolution**: Either label the `substitutedUnits` branch explicitly as an unreachable
  vendor-contract assertion (the wording `fontset.go` already uses for `ot.NewFace`), or give
  constraint 2 a reachable guard — e.g. a synthetic face built without `hhea` driven through
  `LineMetrics` directly, or an AST/source assertion that `Ascender()`/`Descender()`/`LineGap()`
  appear nowhere under `folio-go/`. Replace the self-satisfying `differing` precondition with one
  that inspects measured values rather than the literal map.
- **Related AC**: AC9 (leading), D-2.4.2 constraint 2

### Finding 3: Stale doc comment contradicts the guard it documents
- **Severity**: Minor
- **Category**: Maintainability
- **Location**: `folio-go/byte_neutrality_test.go:128`–`:143`
- **Observation**: The story extended `TestStory23aAddedNoThirdEpic2GateObligation`'s inventory and
  failure message to sanction a third obligation and catch a fourth, but the doc comment above it was
  not updated. It still reads *"The Epic 2 gate owes exactly two things and must not be given a
  third: the four-target matrix legs, and D-2.3.5's Thai sign-off"* — directly contradicting the
  inline inventory comments and the failure message beneath it, both of which now say three/fourth.
  The same comment also states *"Under the per-epic heavy-test cadence (D-000.4) this story writes no
  matrix leg and runs none"*, which is now misleading next to a registered-and-run 2.4 leg.
- **Impact**: The assertions are correct; only the prose is wrong. But this is the exact comment a
  future reader consults to learn how many obligations the gate carries, and it now states a number
  the code beneath it refutes.
- **Suggested Resolution**: Update the doc comment to name three obligations and D-2.4.3, mirroring
  the failure message. Scope the cadence sentence to Story 2.3a explicitly.
- **Related AC**: AC16, ESC-2

### Finding 4: Two rulings block-quote `folio-format.md` non-verbatim (`chain` vs `stack`)
- **Severity**: Nit
- **Category**: Convention
- **Location**: `_bmad-output/implementation-artifacts/folio-mvp-decision-log.md:6248` (D-2.4.2
  resolved) and `:6422` (D-2.4.6); actual text at
  `_bmad-output/specs/spec-folio/folio-format.md:263`
- **Observation**: Both rulings present the binding formulation as a block quote reading
  *"**A chain declares what may appear in an element.** Leading must accommodate what may appear — not
  what does appear."* The file actually says *"**A stack declares what may appear in an element…**"*.
  `folio-format.md` uses "stack" consistently throughout that section (`declared font stack`,
  `a Latin-only stack`), so the file is internally consistent and the rulings are the inaccurate side.
  D-2.4.6's own heading claims *"`folio-format.md` says why"*. There is also a terminology split
  between the code (`chain`, `fontChain`, `lineAdvance(chain …)`) and the spec (`stack`).
- **Impact**: Cosmetic. But D-2.4.5 exists precisely to quote canonical documents verbatim, and a
  ruling that misquotes the document it points at weakens the mechanism for the next reader who
  greps for the sentence.
- **Suggested Resolution**: Correct the two block quotes to `stack`, or align `folio-format.md` to
  `chain` to match the code. Pick one term and use it in both places.
- **Related AC**: AC6, D-2.4.6

### Finding 5: `epics.md:823` cited throughout; the amended line is 827
- **Severity**: Nit
- **Category**: Convention
- **Location**: story file (DN-3, AC1, the Corrections table, Amendments Executed); decision log
  D-2.4.4 / D-2.4.5; actual line `_bmad-output/planning-artifacts/epics.md:827`
- **Observation**: The UAX #14 clause is at line **827** at both `0266a86` and head, not 823. The
  amendment itself is correct and is exactly one line; only the citation is off by four.
- **Impact**: Minimal — D-000.14/D-000.26 ask locations to be precise so a reader can land on them.
- **Suggested Resolution**: Correct the line references, or cite the clause by its text rather than
  its line number.
- **Related AC**: AC1, D-2.4.4

### Finding 6: The new sign-off's digest is described as a literal; it is computed at runtime
- **Severity**: Nit
- **Category**: Convention
- **Location**: story file (Task 8, Heavy-test cadence section);
  `folio-go/expected_breaks_signoff_matrix_test.go`
- **Observation**: The story says the sign-off is *"bound to the sha256 of `expected_breaks.json`
  itself (**`a545e042…99324de`**)"*, reading as a pinned literal. The test contains **no digest
  literal at all** — it computes `sha256.Sum256` over the fixture bytes at runtime and compares that
  to the reader-written `break-signoff.json`. Independently confirmed: `shasum -a 256` of the fixture
  is `a545e042…99324de`, so the binding is genuine and arguably stronger than a literal (it cannot
  go stale).
- **Impact**: None functionally; the description just overstates the mechanism.
- **Suggested Resolution**: Reword to "binds the fixture's digest, computed at runtime (currently
  `a545e042…99324de`)".
- **Related AC**: ESC-2, D-2.4.3

### Finding 7: The D-000.34 sweep did not name a third guard the filter touches
- **Severity**: Nit
- **Category**: Tests
- **Location**: `folio-go/internal/text/corpus_test.go:350` (`TestP2NeverBreaksInsideUnknownRun`)
- **Observation**: The sweep reports two findings and states it covered *"every guard the filter
  touches"*. `TestP2NeverBreaksInsideUnknownRun` consumes constrained `ComputeBreaks` output and is a
  guard the filter touches, but was not named. It is self-referential by construction — its own doc
  comment says it *"can never find a violation by construction"* because it checks the engine's own
  `RunUnknownThai` spans against the engine's own breaks — and unlike its sibling `TestP1…`, which
  gained a vacuity floor in this story, it received none and still reports "zero violations" via
  `t.Log`. After the filter its assertion is doubly implied.
- **Impact**: None new — this is a pre-existing weakness, honestly labelled in place, and the real
  measurement (`TestP2IndependentDPCrossCheck`) is the one that matters and is now green. Raised only
  because the sweep claims completeness.
- **Suggested Resolution**: Either name it in the audit as a known self-referential guard retained
  deliberately, or give it the same vacuity floor P1 received so it cannot decay to zero iterations
  unnoticed.
- **Related AC**: AC4, D-000.34


---

## Finding Resolutions (finisher, 2026-08-24)

**7 findings triaged: 7 FIX · 0 DISMISS · 0 DEFER.** Every finding was reproduced before it was
acted on; one was found to be right in its conclusion and wrong in its proposed remedy, and that
correction is recorded rather than absorbed. One follow-up was *raised* by the triage (the
`chain`/`stack` terminology split) and is deferred as its own item below — it is not a deferred
finding.

| # | severity | decision | resolution |
|---|---|---|---|
| 1 | Minor | **FIX** | Conservation sweep applied to the **class**, not the instance: 7 conservation assertions enumerated across all 31 story-touched `.go` files, 5 fixed, 2 already adequately guarded. The reviewer's suggested remedy was measured **not** to redden and was corrected — see below. Red-proved. |
| 2 | Minor | **FIX** | New reachable guard `internal/fontset.TestLineMetricsReadsTheHheaTableNotTheVendorAccessors`, red-proved against the accessor reroute (suite goes 463·1 → 463·2). The self-referential `differing` precondition replaced with one over measured values. The `substitutedUnits` branch labelled an unreachable forward guard (D-000.24). |
| 3 | Minor | **FIX** | `byte_neutrality_test.go`'s doc comment now enumerates three gate obligations by ruling, names both sign-off tests, and scopes the D-000.4 cadence sentence to Story 2.3a. |
| 4 | Nit | **FIX** | Misquote confirmed against `folio-format.md:263`. Corrected by **appending D-2.4.7**, since the decision log's own header makes it append-only. Terminology split deferred separately. |
| 5 | Nit | **FIX** | Line **827** confirmed at both `0266a86` and head. Four citations corrected in this file; the log's two recorded in D-2.4.7. |
| 6 | Nit | **FIX** | Reworded to "binds the fixture's digest, computed at runtime (currently `a545e042…99324de`)"; digest re-measured with `shasum -a 256`. |
| 7 | Nit | **FIX** | `TestP2NeverBreaksInsideUnknownRun` named in the D-000.34 audit as a deliberately retained self-referential guard **and** given a measured vacuity floor (66 spans / 85 pairs measured, floor 15). |

### Finding 1 — right conclusion, wrong remedy, corrected

The finding is **correct**: `TestMeasureIsAdditiveAcrossASplit` stayed green under the byte-offset
mutation that reddened its AC10 sibling, and it had no non-degeneracy guard. Its **suggested
resolution does not fix that**, and this was established by running it, not by reading it: with
*"require `left > 0 && right > 0` for at least one split per subject"* implemented exactly as
written, the test **still passed the mutation, 4 of 4 subjects witnessed**.

Two measured reasons, both now stated in the test's own doc comment: additivity is preserved by any
*monotone* boundary function and can only see a **saturating** one; and every split in the original
subject table fell on a **face-segment boundary**, where `GlyphInfo.Cluster` — which is
segment-local — reads 0 under both the rune-index and byte-offset interpretations. The subject the
finding called *"supposed to be the discriminating one"*, `"ณัฐวุฒิ เกิด กรุงเทพ"`, breaks at runes 7
and 12 and its segments are `[0,7) [7,8) [8,12) [12,13) [13,20)` — **0 of its splits are interior**.

The fix is therefore the non-degeneracy guard **plus** a whitespace-free Thai subject whose breaks
fall strictly inside one face segment, **plus** a precondition that fatals if the table ever loses
its interior splits again. With all three, the mutation reddens. The full swept enumeration, its
method, and the red-proof transcript are in the Delivery Log above.

### Not reopened

The six items the review verified by construction were **verified, not re-litigated**: the P2
red-proof and its 3,361-offset/200,000-string probes, the eight corrected S4 labels
(dictionary-headword membership, not engine output), `TestComputeBreaksNoBreakInsideCluster`'s
pre-dating vacuity, the four-target matrix, and DW-11's discharge with the count still at **2**. The
finisher independently re-confirmed only the things a finish commit can move: the three golden
digests, `thai-signoff.json`'s absence, the two-module graph, and the fixture change set.

### Follow-up raised (not a finding)

**`chain` (code) vs `stack` (spec).** Pick one term and align both sides. Deferred out of a finish
commit whose scope is otherwise `_test.go` files and documents; it should land in a story that can
re-run the four-target matrix behind the rename. Recorded in **D-2.4.7**.
