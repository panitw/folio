# Story 2.3a: Audit the vendor boundary

**Epic:** 2 — Text, shaping, breaking and page composition
**Story key:** `2-3a-audit-the-vendor-boundary`
**Status:** `done`
**Covers:** no FR/NFR. **This story does not come from `epics.md`.** It was ruled into existence by
**D-000.25** and sequenced between 2.3 and 2.4. `epics.md` is **not** amended by this story and must
not be — there is no clause in it this story implements, weakens or corrects.
**Primary invariant:** **AD-23** (integer-only under `internal/`).
**Adjacent invariants:** AD-1, AD-2, AD-8, AD-21, AD-22, AD-26.
**Governing rulings:** **D-000.25 (the charter)** · **D-000.9** · **D-000.21 (sharpened)** ·
**D-000.22** · **D-000.23** · **D-000.24** · **D-000.26** · **D-000.28** ·
D-000.14 · D-000.15 · D-000.17 · D-000.18 · D-1.3.5 · D-1.3.6 · D-1.3.11 · D-1.5.2 · D-1.5.10 ·
D-1.8.1 · D-2.1.5 · D-2.2.4 · D-2.2.6 (amended) · D-2.3.5 · D-2.3.6
**Retires no DW item. Adds none** unless a `DECISION NEEDED` below is answered by deferral.

**Baseline measured in this run, at creation.** HEAD is **`431a6a5`** — *"Story 2.3: Shape Latin,
Thai and CJK text (finisher)"* — on branch `main`, **working tree clean** (`git status --porcelain`
empty, verified before every measurement and again after the scratch modules were removed, per
D-000.12 corrected).

Test state at baseline, stated with its scope and flags (D-000.26):

| scope | invocation (verbatim) | result |
|---|---|---|
| `folio-go/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counting **every** `--- PASS` / `--- FAIL` occurrence including subtests | **377 PASS · 2 FAIL** |
| `folio-go/` | the same invocation, counting only **top-level** results (`^--- PASS` at column 0) | **241 PASS · 2 FAIL** |
| `lint/` | `CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, every occurrence | **75 PASS · 0 FAIL** |
| `lint/` | the same invocation, top-level only | **37 PASS · 0 FAIL** |

The two failures are `internal/text`'s `TestCorpusMeetsP6ExerciseFloors` (P6g opaque-names floor:
got 7, need ≥ 20) and `TestP2IndependentDPCrossCheck` (26 violations across 17 items). They are
**Story 2.1's deliberate, pre-stated shortfalls (D-2.1.14) — the baseline, not a regression.** Do not
fix them, do not tune the corpus, and do not count them when reporting this story's results.

> **CORRECTED BY THE FINISHER (Story 2.3a Finding 3, Minor).** This paragraph used to read: *"The
> brief that commissioned this story said '377 pass / 2 fail'. The correct split is 375 PASS + 2 FAIL
> = 377 tests."* **That correction was itself wrong, and the brief was right.** Re-measured at
> `431a6a5` in an isolated worktree under the invocation now printed in the table above:
> **377 `--- PASS` occurrences and 2 `--- FAIL` occurrences**, i.e. 379 test results, not 377. The
> `375` was arrived at by assuming `377` was a total that already contained the two failures, and
> subtracting them; nothing was counted.
>
> The irony is the lesson, so it is left standing rather than tidied away: this paragraph invoked
> **D-000.18** — *"confirm against the artifact, not against a table summarising it"* — and then
> restated a figure without confirming it against the artifact. Every absolute in this file is now
> recomputed from a cited invocation, and the counting rule is stated beside each number, because
> `377` is a true statement under one rule and a false one under another.

Everything asserted below was measured against `431a6a5` with scratch Go modules **outside the
repo** (`textshape v0.0.15` and `golang.org/x/tools v0.49.0` from the module cache; the committed
`.ttf` files read by absolute path; `folio-go` reached through a `replace` directive). One temporary
fixture directory was created under `folio-go/testdata/lint/` to prove a mechanism and was removed;
the clean tree was re-verified at `431a6a5` before this file was written.

---

## In plain terms (read this first if you just want the gist)

The engine leans on an outside font library for everything it knows about a typeface, and that
library has a habit worth taking seriously: when a piece of a font is missing, it does not say so —
it hands back a believable substitute and carries on. Ask it the font's name and it says "Unknown"; ask how tall
the capitals are and it may answer with the height of the tallest letter. A document could come out
looking well-formed while resting on fiction.

That is now closed. The engine checks, the moment a font arrives, that every part it actually reads
is really there, and refuses the font by name if not — instead of rendering something plausible.
Two faults surfaced during the audit and both were fixed here rather than filed away: one crashed
the process outright on a damaged font, and one silently put a made-up name into the finished file,
which now says plainly that it substituted.

A second gap is closed too. A standing rule forbids fractional arithmetic, because fractions are
what make the same document come out differently on different machines, but the check enforcing it
only looked for the *word* for a fraction in the source — so a fraction arriving as an answer from
outside was invisible. There is now a check that understands types rather than spelling, and the
engine no longer asks the font library anything fractional. Every glyph of every shipped typeface
was compared before and after: not one measurement moved, and no finished document changed by a
single byte.

Three things will look wrong to a later reader and are not. Two tests fail on purpose — they belong
to an earlier story and measure a shortfall that was deliberately left visible. The Thai sign-off is
deliberately unfinished, so its gate stays red until a human approves the shaping. And a set of
cross-machine checks was compiled but not run: those run once per epic, not once per story.

---
## Story

**As a** maintainer of an engine whose determinism rests on integer arithmetic and on knowing what
its dependencies actually told it,
**I want** the float ban enforced by type rather than by spelling, and every vendor accessor's
absent-data behaviour written down and asserted,
**So that** the next story to add a vendor call site is written against a known map, and no
substituted default can reach an output byte while every guard reports green.

---

## Do not re-open — settled rulings this story inherits

Reproduced with their rationale so the developer does not re-litigate them.

1. [x] **The shipped faces are STATIC, Regular-only.** No `fvar`, no `gvar`. A caller-supplied variable
   face is rejected at ingestion by `fontset.New` (D-2.2.4). Nothing in this story instances,
   pins, or selects an axis, and no `ot.Shaper` variation setter may be called.
2. [x] **`ot.Face.Upem()` and `ot.Face.PostscriptName()` are already declined**, and the code that
   declines them (`readUnitsPerEm`, `readPostScriptName` in `internal/fontset/fontset.go`) is the
   **pattern this story extends** — not code to revisit. D-000.25's table lists `Upem()` as
   "population path unaudited"; **it has been audited in this run and the ruling's premise no longer
   holds** — see F4. Do not re-derive it.
3. [x] **`fontset.AdvanceForRune` has no caller and is retained deliberately** (D-2.3.6, item 1): it is
   one of the two sites this story owns. **It is in scope for this story's fix.** Whether it survives
   the fix as a caller-less function is AC5's question, not a licence to delete it on sight.
4. [x] **Do not reintroduce a second shaping call in `renderDocument`** (D-2.3.6, item 1). Advances for
   *positioning* come from `GlyphPos.XAdvance` via `internal/text.Shaper.Shape` and nowhere else.
   The two sites this story touches are **not** positioning sites — one is the PDF `/W` width table,
   the other is a caller-less raw-`hmtx` helper. Keep that distinction.
5. [x] **No compressor and no image decoder may be imported under `folio-go/`** (`no-compressor-import`,
   `no-image-decoder-import`, D-1.8.1).
6. [x] **`SOURCE_DATE_EPOCH` as a literal string in any `.go` file under `folio-go/` trips
   `absence-source-date-epoch`** (D-2.1.5) — *including inside error-message strings and comments
   that quote it*. Keep it out of Go sources entirely. It is safe in this `.md` file and in the
   Makefile.
7. [x] **`go list -m all` must stay exactly two modules** for `folio-go`:
   `github.com/panitw/folio/folio-go` and `github.com/boxesandglue/textshape`. The type-aware
   checker lands in **`lint`**, whose own graph already carries `golang.org/x/tools` (D-1.3.6,
   D-1.3.11). **`folio-go/go.mod` is not touched by this story.**
8. [x] **The Epic 2 gate already owes two things and must not be given a third.** (a) The four-target
   matrix legs; (b) **D-2.3.5's Thai sign-off** — `TestShapedTextThaiSemanticSignOffIsRecorded` is
   `//go:build matrix` and is **red as committed** until `fixtures/shaped-text/thai-signoff.json`
   names a reader, a date, what they examined and the digest `5964aad0…92e00f`. **Do not disturb
   either.** AC7 states explicitly that this story adds no third gate obligation, and AC6 is the
   measurement that makes that statement true rather than hopeful.

---

## Scope fence — what this story is NOT

- **It does not change a single output byte.** No fixture is re-recorded. No golden is touched. F6
  is the measurement that makes this a fact rather than an intention; AC6 is the assertion.
- **It does not add a build-tagged test, a matrix leg, or a gate obligation.** See AC7.
- **It does not add a dependency to `folio-go`.** See AC10.
- **It does not fix the two `DECISION NEEDED` items below.** They are live-defect candidates found by
  the audit. Per this story's charter they are **escalated, not designed around.** If the lead rules
  "fix here", the ACs gain a clause; until then, nothing in the ACs depends on them.
- **It does not touch line breaking, segmentation, or anything in Story 2.4's territory.**
- **It does not audit the vendor's *shaping* correctness.** D-2.3.1's HarfBuzz oracle already covers
  that, and it is a different question from "what does this accessor return when its table is
  absent".
- **It does not attempt reachability or dataflow analysis** in the new lint rule. AD-23 is a
  hazard statement, not a mechanism (the same posture `ScanMapRange` takes, D-1.3.5).
- **It does not delete `ot.Face` from the codebase.** `internal/text.NewShaper` legitimately takes
  one; D-1.5.10 fences the *exported* seam, and that fence already holds.

---

## Measured findings — read all of these before writing code

Every row below names its subject: the file, the accessor, the exact input, and the invocation
(D-000.26).

### F1 — The AD-23 guard is syntactic, and a type-aware scan finds exactly two live sites

`folio-go/internal/arch_test.go:50-60` (`findFloatOccurrences`) walks the AST for `*ast.Ident` named
`float32`/`float64` and `*ast.BasicLit` of kind `token.FLOAT`. **A vendor function returning a float
is invisible to it**: the type is inferred, so neither banned identifier ever appears.

**Measurement.** A scratch checker built on `golang.org/x/tools/go/packages v0.49.0`
(`Mode: NeedName|NeedFiles|NeedSyntax|NeedTypes|NeedTypesInfo|NeedImports|NeedDeps`, `Dir` =
`/Users/panitw/Projects/folio/folio-go`, pattern `./...`), reporting every `ast.Expr` whose
`TypesInfo.Types[e]` satisfies `tv.IsValue()` and whose type's underlying `*types.Basic` carries
`types.IsFloat`:

| invocation | float-typed value expressions | typed expressions seen | packages |
|---|---|---|---|
| `Tests: false` | **4**, in **1 file** | 15,883 | 12 |
| `Tests: true` | **10** raw / **6 distinct (file, line)**, in **3 files** | 62,700 | 31 |

The `Tests: true` duplication is the ordinary package/test-variant double-load, not six real sites.
The distinct sites are:

| site | expression | reachable from a render? |
|---|---|---|
| `internal/fontset/fontset.go:328` | `f.face.HorizontalAdvance(gid)` (`*ast.CallExpr`) | no — `AdvanceForRune` has no caller (D-2.3.6) |
| `internal/fontset/fontset.go:329` | `adv` (`*ast.Ident`, the read of that value) | as above |
| `internal/fontset/fontset.go:565` | `f.face.HorizontalAdvance(oldGID)` (`*ast.CallExpr`) | **yes** — the `/W` width table, every render with text |
| `internal/fontset/fontset.go:566` | `adv` (`*ast.Ident`) | **yes** |
| `shaping_expectations_test.go:341` | `face.HorizontalAdvance(gid)` inside `int16(...)` | test only, deliberate (see F7) |
| `internal/fontset/fontset_test.go:515` | the untyped constant **`700`** (`*ast.BasicLit`, kind `INT`) | test only (see F3) |

**D-000.25 cited `fontset.go:289` and `:476`. At `431a6a5` they are `:328` and `:565`** — the
finisher's changes moved them. The count is still two call sites; the type-aware scan reports four
expressions because it also sees each `adv` read. **There is no third production site.** The blast
radius is bounded and closed.

`shaping_expectations_test.go:341` and `fontset_test.go:515` are **both green under the syntactic
guard today** — `TestNoFloat64UnderModule` passes at `431a6a5`, and it *does* walk `_test.go` files.

### F2 — Both sites are provably lossless today, and that is what makes them a forward hazard

`ot.Face.HorizontalAdvance` (`ot/metrics.go`) is:

```go
func (f *Face) HorizontalAdvance(glyph GlyphID) float32 {
	if f.hmtx != nil {
		return float32(f.hmtx.GetAdvanceWidth(glyph))   // uint16
	}
	return float32(f.upem)                              // uint16
}
```

Both branches convert a `uint16`. `float32` has a 24-bit mantissa, so every `uint16` is **exactly**
representable and `int64(float32(x)) == x` for every possible input, on both branches. **The guard is
green for a reason that has nothing to do with the guard** (D-000.25, verbatim).

**Consequence for the ACs, stated precisely, because these are two different claims:**

- The **new lint rule** *has* a red-proof: it reports both sites at `431a6a5`, and the syntactic
  guard reports zero on the same code. AC3 constructs it as a fixture pair as well.
- The **claim that the current sites are lossy** has **no** red-proof and cannot have one — the
  conversion is exact for every input the vendor can produce. **Do not manufacture one, and do not
  write an assertion that depends on a truncation that cannot happen** (D-000.24). The disposition of
  the two sites is not "prove they are wrong"; it is AC5's "stop asking a float question that has an
  integer answer".

### F3 — A second, independent syntactic blind spot: an untyped integer constant in a float parameter

`internal/fontset/fontset_test.go:515` reads:

```go
in.PinAxisLocation(ot.MakeTag('w', 'g', 'h', 't'), 700)
```

`PinAxisLocation`'s second parameter is `float32`, so go/types gives the literal `700` the type
`float32`. It is an `*ast.BasicLit` of kind **`token.INT`**, and the syntactic guard only flags kind
`token.FLOAT` — **so `700` is invisible to it.**

This **falsifies a claim standing in two production comments.**
`internal/fontset/fontset.go` (in `New`'s D-2.2.4 block) says:

> *"Pinning an explicitly chosen weight is unreachable: the vendor's `PinAxisLocation` requires the
> identifier `float32`, which `internal/arch_test.go:54` bans under `internal/` AND the module
> root (AD-23)."*

and `internal/text/shape.go` (the `Shaper` doc comment) says the same of the four `float32`-taking
variation setters:

> *"…four take `float32`, which `internal/arch_test.go` bans under `internal/` and the module root
> (AD-23)."*

**`PinAxisLocation` does not require the identifier**, and the test at `:515` calls it right now with
the guard green. The *conclusion* (do not pin) is still correct and still ruled by D-2.2.4 for
independent reasons — the vendor never rewrites `OS/2.usWeightClass`. But the **stated mechanism is
false**, and a reader who trusts it will believe AD-23 fences a door that is open. AC9 corrects both
comments; the new rule is what makes the corrected statement true.

### F4 — `Upem()`'s population path, traced and settled: D-000.25's premise no longer holds

D-000.25's table lists `Face.Upem()` as *"plain field read — but its population path is unaudited"*,
and warns that if construction substitutes a default for a missing `head`, **D-1.5.2's 16–16384
validation would pass on fiction.**

**Traced.** `ot.NewFace` (`ot/metrics.go`) does exactly the feared thing:

```go
if data, err := font.TableData(TagHead); err == nil { f.head, _ = ParseHead(data) }
if f.head != nil { f.upem = f.head.UnitsPerEm }
if f.upem == 0 { f.upem = 1000 }   // "Default for CFF"
```

**Measured** (`folio-go/testdata/fonts/Roboto-Regular.ttf`, upem 2048, `head` record removed from the
table directory): `face.Upem()` returns **1000**.

**But folio never reads it.** `fontset.New` calls `readUnitsPerEm(parsed)` → `parseHead` →
`ot.ParseHead(font.TableData(ot.TagHead))`, which returns `ErrTableNotFound` and is reported as a
located load error. **Measured end to end at the public entry point:** `folio.Render` with a
`FontSet` whose bytes lack `head` returns
`folio: Render: element e1: fontset: font "Roboto-Regular": read head table: table not found`.

**Settled: the ruling's hazard is real in the vendor and closed in folio, by Story 2.2's
`readUnitsPerEm`. D-1.5.2's validation does not pass on fiction.** This is recorded so the question
is not re-opened; AC8's enumeration carries the row with this disposition.

### F5 — The full enumeration: every vendor accessor folio calls, and what it returns when its data is absent

**Subject:** `folio-go/testdata/fonts/Roboto-Regular.ttf` (upem 2048, 1,294 glyphs), textshape
`v0.0.15`. Each row measured by removing exactly one 16-byte record from the font's table directory
(decrementing `numTables`, leaving the table's bytes in place but unreferenced) and re-reading
through `ot.ParseFont` → `ot.NewFace`.

**Vendor behaviour, measured:**

| accessor | intact | with its table absent | honest or substituted? |
|---|---|---|---|
| `ot.ParseFont` | ok | — | honest (returns `error`) |
| `ot.NewFace` | `err=nil` | **`err=nil` in every case measured** | **it has no failure mode** — every table parse error is discarded with `_`. See F8. |
| `Font.HasTable(tag)` | true | false | honest |
| `Font.TableData(tag)` | bytes | `ErrTableNotFound` | honest |
| `Font.NumGlyphs()` | 1294 | **`0`** on missing/short `maxp` | **substituted** — a caller looping to it silently does nothing |
| `Face.Upem()` | 2048 | **`1000`** (no `head`) | **substituted** — declined by folio (F4) |
| `Face.PostscriptName()` | `"Roboto-Regular"` | **`"Unknown"`** (no `name`) | **substituted** — declined by folio |
| `Name.FamilyName()` | `"Roboto"` | `"Unknown"` / unguarded `entries[1]` | **substituted** — **folio never calls it** (grep: zero call sites at `431a6a5`) |
| `Face.Ascender()` | 1900 | **`800`** (no `hhea`) | **substituted** |
| `Face.Descender()` | −500 | **`−200`** (no `hhea`) | **substituted** |
| `Face.CapHeight()` | 1456 | **`1900` = `Ascender()`** (no `OS/2`, **or `sCapHeight == 0`**) | **substituted** |
| `Face.BBox()` | (−1509,−555,2352,2163) | **`(0, −200, 1000, 800)`** (no `head`) | **substituted** |
| `Face.HorizontalAdvance(g)` | 1839 for gid 36 | **`float32(upem)` = 2048** (no `hhea` **or** no `hmtx`) | **substituted** |
| `Face.HorizontalAdvance(65535)` | **`506`** on a 1,294-glyph face | — | **fabricated** — `Hmtx.GetAdvanceWidth` returns `lastAdvanceWidth` for any out-of-range gid |
| `Face.Cmap()` | non-nil | **`nil`** | **honest** — and folio nil-checks it at `fontset.go:290` and `:320` |
| `Cmap.Lookup(cp)` | `(gid, true)` | `(_, false)` | honest |
| `subset.CreatePlan` / `plan.Execute` | ok | `subset: required table missing` | honest |
| `plan.MapGlyph` / `plan.OldGlyph` | `(gid, true)` | `(_, false)` | honest |
| `ot.NewShaperFromFace` | ok | returns `error` | honest |
| `Shaper.Shape` → `buf.Info` / `buf.Pos` | parallel slices | — | **unchecked contract**, already checked by folio (`internal/text/shape.go`) |

**Now the half that matters — what folio does with each, measured at the public entry point.**
`folio.Render(tmpl, folio.Data("{}"), folio.Params("{}"), folio.FontSet{"Roboto-Regular": <bytes>})`
against `fixtures/font-text/input.folio`:

| table removed | `folio.Render` outcome | does folio distinguish the substitution? |
|---|---|---|
| *(none)* | 22,310 bytes, `err=nil` | — |
| `head` | **located error**, `read head table: table not found` | **yes** — `readUnitsPerEm` (F4) |
| `hhea` | **located error**, `execute subset: subset: required table missing` | **yes**, but by the subsetter, not by folio |
| `hmtx` | **located error**, `execute subset: subset: required table missing` | **yes**, same |
| `name` | **22,310 bytes, `err=nil`, `/BaseFont /HXRYNT+Roboto-Regular`** | **no** — see `DECISION NEEDED Q2` |
| `OS/2` | **22,198 bytes, `err=nil`, `/CapHeight 928`** (intact: `/CapHeight 711`) | **NO — a substituted default reaches the produced PDF** |
| `maxp` | **PANIC**: `runtime error: makeslice: len out of range` | **no** — see `DECISION NEEDED Q1` |

**The `OS/2` row is the enumeration's answer in one line.** With `OS/2` removed, folio renders a
document that reports success and declares `/CapHeight 928` — which is not a cap height, it is the
ascender. Every guard in the repo stays green. This is D-000.25's Finding 2 arriving at an output
byte, and it is the reason AC8 exists.

**It does not affect anything folio ships.** All three shipped faces carry `OS/2` **version 4** with a
real `sCapHeight` (NotoSans 714, NotoSansSC 733, NotoSansThai 714) — measured. The exposure is
**caller-supplied faces**, which is exactly the surface `folio.FontSet` is.

**`hhea`/`hmtx` absence is unreachable through `Render` today**, because the vendor subsetter refuses
first. That makes the `Ascender()=800` / `Descender()=−200` / `HorizontalAdvance()=upem` rows
**forward hazards with no available red-proof through the public path** (D-000.24) — and AC8 must
label them that way rather than crediting them with a proof.

### F6 — The fix is byte-neutral, and this is measured, not assumed

The vendor exposes the same number in integer form: `ot.ParseHmtxFromFont(font) (*ot.Hmtx, error)`
and `(*ot.Hmtx).GetAdvanceWidth(glyph GlyphID) uint16`. This is precisely the shape Story 2.2 used
for `readUnitsPerEm` and `readPostScriptName` — read the table directly, take the integer, decline
the accessor.

**Measured**, over every glyph of every face folio ships or tests with:

| face | glyphs | `int64(face.HorizontalAdvance(g))` vs `int64(hmtx.GetAdvanceWidth(g))` |
|---|---|---|
| `folio-go/testdata/fonts/Roboto-Regular.ttf` | 1,294 | **0 mismatches** |
| `folio-go/fonts/notosans/NotoSans-Regular.ttf` | 4,515 | **0 mismatches** |
| `folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf` | 467 | **0 mismatches** |
| `folio-go/fonts/notosanssc/NotoSansSC-Regular.ttf` | 31,036 | **0 mismatches** |
| **total** | **37,312** | **0** |

**So the swap changes no width, no `/W` array, no content stream and no golden.** That is what makes
this story landable without a third gate obligation (AC6, AC7).

`ParseHmtxFromFont` is also **strictly more honest** than the accessor: it returns `ErrInvalidTable`
when `NumGlyphs()` is 0 and propagates `TableData`'s error for a missing `hhea` or `hmtx`, where
`HorizontalAdvance` substitutes `upem`. The fix closes a float hole and a substitution hole with one
change.

### F7 — The two test-file sites are legitimate, and the honest disposition is an inventory, not an exemption

`shaping_expectations_test.go:341` reads `int16(face.HorizontalAdvance(gid))` **deliberately** — its
own comment says so: *"It reads `ot.Face.HorizontalAdvance` deliberately — that IS the accessor the
old path used, float32 return and all, and reproducing it exactly is the point."* It is the negative
control for `TestShapedExpectationsObservability`. Deleting it would delete the control.

`fontset_test.go:515`'s `700` is D-2.2.4's sanctioned V5 tag-discrimination test, which
`fontset.go`'s own comment blesses: *"This package's own tag-discrimination test (V5) exercises a
second, non-default instance by calling the vendor subsetter directly, entirely from the test file,
never through this method."*

The existing syntactic guard **does** scan `_test.go` files (`walkGoFiles` skips only `testdata` and
dot-directories), so narrowing the type-aware rule to non-test files would make it **strictly weaker
in file scope than the guard it strengthens** — a silent regression dressed as a fix.

**Disposition (AC2):** run the rule at **two scopes**, mirroring the existing two-caller structure —
a **production scope** asserted to **zero**, and a **test scope** asserted to **exactly the named
(file, rule) pairs**, by file and rule, using the `assertExactFindings` pattern already in both
modules. **This is an inventory, not a denylist.** A new float-typed expression anywhere in a
`_test.go` file fails the build; the two known ones are legible on the page instead of hidden behind
a named exemption. D-2.1.3 and D-000.15's rotting-list objection does not attach, because nothing is
excused by name — every site is enumerated and any addition is a failure.

### F8 — `ot.NewFace` cannot fail, so folio's error branch at `fontset.go:70` is unreachable

`ot.NewFace` ends `return f, nil` on every path; every table parse error inside it is discarded into
`_`. Measured: `err=nil` for all seven table-removal cases in F5, including the one that panics.

`internal/fontset/fontset.go:70-73` therefore contains an **unreachable error branch**. Under
D-000.24 the honest treatment is to **keep it and label it** — it is the assertion that would catch
the vendor gaining a failure mode — exactly as `Subset`'s "not retained" branch is already labelled
(D-2.3.6 / Story 2.3 Finding 8). AC9 adds the label. **Do not delete the branch and do not count it
as coverage.**

---

## DECISIONS NEEDED — escalate before development starts

Both are **live-defect candidates found by the audit**, not guard gaps. Per this story's charter they
are reported, not designed around. **No AC below depends on either being answered**; if the lead
rules "fix in 2.3a", the answer becomes an additional AC.

### Q1 — `folio.Render` **panics** on a caller-supplied font whose `maxp` is missing or short

**Measured, at the public entry point.**
`folio.Render(tmpl, folio.Data("{}"), folio.Params("{}"), folio.FontSet{"Roboto-Regular": <Roboto with the `maxp` record removed>})` against `fixtures/font-text/input.folio`:

```
panic: runtime error: makeslice: len out of range
  ot.ParseHmtx        (textshape@v0.0.15/ot/hmtx.go:38)
  ot.NewFace          (textshape@v0.0.15/ot/metrics.go:342)
  internal/fontset.New (folio-go/internal/fontset/fontset.go:70)
```

**Mechanism.** `NewFace` calls `ParseHmtx(data, int(hhea.NumberOfHMetrics), font.NumGlyphs())`.
`NumGlyphs()` returns **0** when `maxp` is absent or shorter than 6 bytes; `NumberOfHMetrics` is
1294; `ParseHmtx` then evaluates `make([]int16, numGlyphs-numberOfHMetrics)` = `make([]int16, -1294)`.
`ParseHmtxFromFont` guards this (`if numGlyphs == 0 { return nil, ErrInvalidTable }`); **`NewFace`
does not.**

**Why it is a defect and not a curiosity.** `folio.FontSet` is `map[string][]byte` of
**caller-supplied bytes** — a public input, on the same footing as a template. The engine's own
precedent is explicit: `TestParseTemplateRejectsPNGTruncatedInHeaderWithoutPanicking` exists because
malformed caller input must produce a located error, never a panic. Font bytes have no such test, and
they crash the process.

**The remedy is small and folio-side** (validate `maxp` before `ot.NewFace`, in the same block that
already reads `head` directly), which is precisely why it should be a ruling rather than a quiet
inclusion: it adds a load error that did not exist, and that is a behaviour change to a public entry
point.

**Question for the lead:** fix in 2.3a (one guard in `fontset.New`, one located-error test,
red-proved by the panic above), or file as a `DW` item?

### Q2 — folio computes "observably absent" and then substitutes a plausible name one layer up

`readPostScriptName` returns `""` when the face declares no `name` record — Story 2.2's correct
disposition, and D-000.25's stated pattern for the rest. `internal/pdf/textdoc.go:334-336` then does:

```go
psName := face.PostScriptName
if psName == "" {
	psName = name          // the FontSet key
}
```

**Measured:** rendering Roboto with its `name` record removed produces
`/BaseFont /HXRYNT+Roboto-Regular` — byte-identical in length to the intact render (22,310 bytes).
The PDF declares a CIDFontName the embedded program does not carry.

**It is documented and deliberate** — the comment says *"Falls back to the FontSet key — the pre-2.2
behaviour — only when the supplied program declares no name record at all."* So this is **not** an
undisclosed defect, and it is raised at the lower severity of the two.

**But it is the exact shape D-000.25 rules against**, one layer above where 2.2 fixed it: an
observable absence converted back into a plausible value, under ISO 32000-1 Table 117, which requires
`/BaseFont` to be *"the value of the CIDFontName entry in the CIDFont program"*. A program with no
name has no such value, and the FontSet key is the caller's filing label — the very thing D-2.2.6
moved away from.

**Question for the lead:** (i) leave as ruled and merely pin it with a test naming it deliberate;
(ii) emit the tag alone (`/HXRYNT`) when the program declares no name; or (iii) make a nameless
program a load error. **The audit's job is to surface it; the choice is not the audit's.**

---

## Acceptance Criteria

### AC1 — A type-aware float checker exists in `lint`, as a pure function

`lint/internal/rules/` gains a checker with the shape every rule in that package already has (AC1 of
Story 1.3, D-1.3.6): **a pure function over a target directory returning `(findings, stats, error)`,
with no `*testing.T` parameter, no hard-coded root and no repo-root discovery inside it**, so the
same function can be pointed at the real tree and at a fixture tree.

- It resolves the target subtree as one coherent package graph via `golang.org/x/tools/go/packages`,
  with the same `Mode` `ScanMapRange` uses, for the reason D-1.3.11 gives.
- **Detection is by TYPE, not by spelling, and covers the class rather than the instance** (D-000.23):
  the predicate is *"an `ast.Expr` whose `TypesInfo.Types[e]` satisfies `tv.IsValue()` and whose
  type's underlying `*types.Basic` has `types.IsFloat` set"*. **It must not name
  `HorizontalAdvance`, `textshape`, or any accessor, package or symbol** — a list of known
  float-returning functions is the rotting-list pattern (D-2.1.3, D-000.15) and would miss the next
  one.
- **A package that fails to load or type-check is a hard error, never zero findings** (D-1.3.11):
  `packages.Load`'s nil top-level error is not sufficient; every package's own `Errors` field is
  inspected and any entry fails the scan loudly. An expression whose type does not resolve is a hard
  error, not a silent skip.
- It carries a stable rule id declared as a `const` beside the checker, matching this package's
  existing convention (`RuleMapRange = "map-range"`, `RuleNoCompressor = "no-compressor-import"`, …).
- Findings carry `Path` relative to the scanned root, the rule id, the line, and a message that
  **names the resolved type and the expression's position** — not "a float was found".

**Red-proof: AC3.**

### AC2 — The rule's scope matches AD-23's, and the test-file surface is an inventory, not an exemption

Two production callers, one checker (the structure `TestNoFloat64UnderInternal` /
`TestNoFloat64UnderModule` and `TestMapRangeProductionScan` / `TestMapRangeUnderModule` already use):

1. [x] **Production scope** — `folio-go/`, `Tests: false`. **Asserted to report zero findings**, after
   AC5's fix. Before AC5's fix it reports the four expressions of F1; that transition is AC5's
   red-proof and must be recorded with both numbers.
2. [x] **Test scope** — `folio-go/`, `Tests: true`, findings **deduplicated by (path, line)** because
   go/packages loads a package and its test variant separately (F1: 10 raw → 6 distinct).
   **Asserted by `assertExactFindingSites` against exactly the named (file, rule, line) triples** —
   at `431a6a5` the files are `shaping_expectations_test.go` and
   `internal/fontset/fontset_test.go`; after this story they are those two plus
   `internal/fontset/vendorboundary_test.go`, carrying **five sites across three files**.

**The test-scope assertion is an inventory. Nothing is excused by name.** Adding a float-typed
expression to any `_test.go` file fails the build; removing one fails it too. The assertion compares
**by site — (file, rule, line) — never by count** (D-1.3.3 amended): a scan finding the right
*number* of wrong things must still fail.

> **CORRECTED BY THE FINISHER (Story 2.3a Finding 2, Major).** As shipped for review, this AC said
> the assertion compared *"by file and rule"* while the paragraph beneath it claimed *"adding a
> float-typed expression to any `_test.go` file fails the build"*. **Those two sentences contradict
> each other, and the guard implemented the weaker one.** Keyed by `(file, rule)`, every additional
> site inside an already-enumerated file was invisible, and so was removing one of two — a per-file
> allowlist wearing an inventory's name, which is the false-credit class D-000.24 and D-000.28 name.
>
> The gap was **live, not hypothetical**. When the finisher re-keyed the assertion and measured, the
> tree held **four** sanctioned sites, not three: `internal/fontset/vendorboundary_test.go` already
> carried **two** (the `hhea` and `hmtx` rows), and the file-keyed `want` list could represent only
> one of them. A real site was standing unlisted while the test reported `ok`. The review had to
> *inject* a second site to expose the gap; the gap was already occupied.
>
> The assertion is now site-exact and the comment matches it. Both directions are red-proved — see
> the Delivery Log. The line numbers are load-bearing and will churn when unrelated edits shift a
> site; that friction is deliberate, because this list is the register of sanctioned AD-23
> violations and it should be impossible to change what is in it without a human reading the diff.

**Both callers assert their vacuity guard from the checker's OWN returned stats**, never from an
independently re-derived second walk (Major 5 of Story 1.3's QA review; the same reasoning
`noFloat64Stats` and `MapRangeStats` carry): injecting `if true { return nil, stats{}, nil }` as the
checker's first statement must zero the stats too. At minimum the stats report packages visited **by
name** (`"internal/fontset"` and the module-root package must both appear) and a non-zero count of
**typed expressions examined** — the statistic that would make "a checker that resolved nothing"
visible.

**Scope consequence, stated explicitly so nobody is surprised:** this rule runs under
`cd lint && go test ./...`, **not** under `folio-go`'s own suite. CI already runs both
(`.github/workflows/ci.yml`, the `folio-go` and `lint` jobs). The existing syntactic guard in
`folio-go/internal/arch_test.go` **stays exactly as it is** — it is not deleted, not weakened, and
not merged into the new one. Two guards, two mechanisms, one invariant.

### AC3 — Red-proof: a retained fixture pair, and the syntactic guard proven blind to it

A retained fixture tree at **`folio-go/testdata/lint/<rule-id>/`** (never under `folio-go/internal/`,
F-10; the location every other lint fixture uses), containing at minimum:

- a **violating** file that truncates a vendor float to an integer with **no float identifier and no
  float literal anywhere in the file** — e.g. `int64(f.HorizontalAdvance(gid))` where `f` is an
  `*ot.Face`. **The word `float32`/`float64` must not appear even in a comment in that file**, so the
  fixture's two readings cannot be confused.
- a **compliant** file reading an integer-typed vendor accessor (e.g. `int64(f.Ascender())`), so the
  rule is shown not to fire on every vendor call.

**Both halves of the red-proof are asserted:**

1. [x] The new checker, pointed at the fixture root, reports **exactly** the violating file (by file and
   rule, `assertExactFindings`) and **not** the compliant one.
2. [x] **`folio-go/internal/arch_test.go`'s `findFloatOccurrences`, run over the same violating fixture
   file, reports ZERO.** This is the assertion that states the gap D-000.25 named, rather than
   describing it in prose. It lives in `folio-go`'s suite (that is where the syntactic scanner is),
   points at the same fixture path, and its failure message says what it means: *the syntactic guard
   cannot see an inferred float.*

**Measured precondition (verified in this run at `431a6a5`):** a `packages.Load`-based scan pointed
directly at a `folio-go/testdata/lint/…` directory **does** type-check files importing
`github.com/boxesandglue/textshape/ot` — 1 finding in the violating file, 0 in the compliant one,
18 typed expressions, 1 package — **and the module-scope scan is unchanged by the fixture's presence
(4 findings, identical to without it)**, because `testdata` is outside package matching. The fixture
directory was removed and the clean tree re-verified.

### AC4 — The enumeration of the vendor boundary is a committed artifact

A committed document enumerating **every vendor accessor `folio-go` calls** — the `ot` and `subset`
packages of `textshape v0.0.15` — with, per accessor: **what it returns when its data is absent**,
and **whether folio currently distinguishes that from a real value**. F5 is the measured content;
this AC is that it is committed, cited and maintainable.

Requirements:

- **Every row cites its subject** (D-000.26): the file, the accessor, the exact input (which table
  was removed), and the invocation that produced the observation. A row reading "returns a default"
  without naming what was measured does not satisfy this AC.
- **Rows are classified into three states, not two**: *honest* (reports absence), *substituted*
  (returns a plausible value), *fabricated* (returns a value for input outside its domain — e.g.
  `HorizontalAdvance(65535)` on a 1,294-glyph face returning 506).
- **Rows whose hazard is unreachable through the public path today are labelled
  "forward hazard, no available red-proof"** (D-000.24) — specifically the `hhea`/`hmtx` rows, which
  the vendor subsetter refuses before folio can observe them. **They are not credited with a proof
  they lack, and the assertions around them are not weakened to match.**
- It records the **version it was measured against** (`textshape v0.0.15`) prominently, because the
  document is a claim about a specific dependency version and nothing else.
- It states the **method** precisely enough to be replayed: one 16-byte record removed from the table
  directory, `numTables` decremented, table bytes left in place.

Location is the developer's call between `folio-go/internal/fontset/` (beside the code that declines
the accessors) and `docs/`; whichever is chosen, AC8's test must cite it by path.

### AC5 — Both live float sites are FIXED by declining the accessor, not exempted

`internal/fontset/fontset.go` stops calling `ot.Face.HorizontalAdvance` at **both** sites (`:328` and
`:565` at `431a6a5`). Advances are read through the **integer** path — `ot.ParseHmtxFromFont` /
`(*ot.Hmtx).GetAdvanceWidth`, returning `uint16` — parsed once per `Font` at construction and held on
the struct, in the same shape as `psName`, `unitsPerEm`, `created` and `modified`.

- **This is Story 2.2's `readPostScriptName` pattern, applied a third time**, and the code comment
  must say so and say why: the accessor is declined because it is float-typed *and* because it
  substitutes `upem` for a missing `hmtx`, where `ParseHmtxFromFont` reports the absence.
- **`AdvanceForRune` is fixed in place, not deleted.** D-2.3.6 retained it deliberately as one of
  this story's two subjects. Its doc comment is updated to record that this story converted it to the
  integer path; whether it survives without a caller is a separate question this story does not
  answer.
- **Every glyph id handed to `GetAdvanceWidth` is bounds-checked against the face's own glyph count
  first**, because `GetAdvanceWidth` **fabricates** for an out-of-range id (returns
  `lastAdvanceWidth`; measured: 506 for gid 65535 on 1,294-glyph Roboto). `Subset` already validates
  its input ids against `NumGlyphs()` (Story 2.3 Finding 8) — the new read must be inside that
  guarantee, or carry its own.
- If parsing `hmtx` fails, it is a **located load error naming the font**, never a substitution and
  never a panic — matching `readUnitsPerEm`'s and `readPostScriptName`'s error shapes.

**Red-proof:** the production-scope scan of AC2 reports **4 findings in 1 file before this change and
0 after**, and both numbers are recorded in the completion notes with the invocation that produced
them. **Do not assert that a truncation was lossy** — F2 proves it cannot be. The assertion is that
the float-typed expression is gone, which is exactly what the rule measures.

### AC6 — The change is byte-neutral, and that is asserted rather than assumed

**No fixture is re-recorded and no golden hash changes.** Specifically, after AC5:

- `fixtures/minimal-rect/`, `fixtures/font-text/`, `fixtures/image-embed/`,
  `fixtures/multi-script-fallback/` and `fixtures/shaped-text/` all still match, with their committed
  `expected.json` digests **unchanged**. In particular **`fixtures/shaped-text/`'s
  `5964aad0…92e00f` is untouched**, which is what keeps D-2.3.5's sign-off record valid — that record
  is bound to those exact bytes, and re-recording would invalidate a human sign-off that has not even
  been given yet.
- The assertion group **first asserts each fixture's `expected.json` exists and carries a `sha256`
  field**, then asserts its value (D-000.21 sharpened: assert on the artifact that carries the
  property, and prove it carries it). A test that silently passes over a missing fixture proves
  nothing.
- The completion notes record the **per-face equivalence measurement** — the F6 table, re-run by the
  developer against the four committed faces, with the total glyph count and the mismatch count. If
  any face reports a non-zero mismatch, **stop and escalate**: the premise of this AC is false and
  the story's shape changes.

### AC7 — No third Epic 2 gate obligation

Stated as an assertion, not as an intention:

- **No new `//go:build matrix` test is added by this story.** `go build -tags=matrix ./...` and
  `go vet -tags=matrix ./...` still succeed, and the set of matrix-tagged test functions in
  `folio-go/` is **unchanged** from `431a6a5`.
- `TestShapedTextThaiSemanticSignOffIsRecorded` is **not modified**, and
  `fixtures/shaped-text/thai-signoff.json` is **not created** by this story. It remains the Epic 2
  gate's outstanding red, owned by D-2.3.5.
- `.github/workflows/matrix.yml` is **not modified**, and
  `TestMatrixDocumentSlugsAreRegisteredInCI` still passes.
- The completion notes state, in one line and against the measurement above, that the Epic 2 gate
  owes **exactly the two things it owed at `431a6a5`** — the four-target matrix legs and the Thai
  sign-off — **and nothing added here.**

**Per the heavy-test cadence (D-000.4), no matrix leg is written and none is run. There is nothing to
register.** See the cadence section below.

### AC8 — The enumeration's load-bearing rows are executable, not narrated

A test that **derives** the enumeration's claims from the vendor rather than restating them, so the
document cannot drift (D-000.14: a total narrated beside the artifact it summarises is a second
source of truth; D-000.28: a claim written before the fact it asserts is false from birth).

At minimum, and each **naming its subject in its failure message**:

1. [x] **Substitution is real and detectable.** For each of `head`, `name`, `OS/2`, `hhea`, `hmtx`, using
   a test-local table-stripping helper over `folio-go/testdata/fonts/Roboto-Regular.ttf`: assert the
   vendor accessor returns the **exact substituted value** the enumeration names (`Upem()==1000`,
   `PostscriptName()=="Unknown"`, `CapHeight()==Ascender()`, `Ascender()==800`,
   `Descender()==-200`, `BBox()==(0,-200,1000,800)`, `HorizontalAdvance(g)==upem`). If a future
   `textshape` changes a default, this fails and the document is corrected from the artifact.
   **Assert the stripped font still parses first** — a strip that produced an unparseable font would
   make every row pass vacuously for the wrong reason.
2. [x] **Folio's declines hold.** Loading each stripped font through `fontset.New` (or, where the seam is
   unexported, through `folio.Render` at the public entry) yields the outcome F5's second table
   records: a **located error** for `head`; a **located error** for `hhea`/`hmtx`; **success** for
   `name` and `OS/2`. Assert the error's *text* names the font and the table, not merely that an
   error occurred.

   > **STALE AS WRITTEN — corrected by the finisher (Story 2.3a Finding 8, Nit).** **D-2.3a.1 moved
   > `OS/2` from *success* to *refused at ingestion***, so this AC's expected outcome changed
   > mid-story and the sentence above was never updated. The shipped behaviour is the ruled one:
   > `OS/2` sits in the decline loop beside `head`, `maxp`, `hhea` and `hmtx`, and only `name` and
   > `cmap` are tolerated. Read this row as: *a located error for `head`, `maxp`, `hhea`, `hmtx` and
   > `OS/2`; success for `name` and `cmap`.*
   >
   > Recorded rather than silently edited because the story's own standard is that mid-story rulings
   > are **stated, not absorbed** — and the "Two departures" section disclosed the AC8.3 consequence
   > of this same ruling in detail while leaving AC8.2's consequence undisclosed. One ruling moved
   > two ACs; only one of them was declared. A reader checking AC8.2 against the test would otherwise
   > find them in apparent conflict with nothing reconciling them.
3. [x] **The `OS/2` substitution reaches an output byte.** Render `fixtures/font-text/input.folio` with
   an `OS/2`-stripped Roboto and assert the produced PDF contains **`/CapHeight 928`** while the
   intact render contains **`/CapHeight 711`** — asserting the **emitted bytes**, not that a field
   was read. This is the enumeration's single strongest row and the one most worth a live guard.
   Assert the intact render contains a `/CapHeight` entry at all before comparing the two.
4. [x] **`FamilyName` has no call site.** An assertion that `folio-go`'s sources contain zero calls to it
   — so the day someone adds one, the enumeration's "not called" row stops being true and the build
   says so. **This is a live guard, not a comment.**
5. [x] **The enumeration document exists and is cited by path**, and the test names it. A test asserting
   properties of a document that has been renamed or deleted must fail, not pass.

Where a row cannot be red-proved through the public path — the `hhea`/`hmtx` rows of F5 — the
assertion **says so in its own text** and is not counted as coverage (D-000.24).

### AC9 — The three false or unlabelled statements found by the audit are corrected at their source

Each is a claim standing in committed code that the audit measured to be wrong or unlabelled. All
three are corrected in this story, at the file that carries them:

1. [x] **`internal/fontset/fontset.go`, `New`'s D-2.2.4 block** — *"`PinAxisLocation` requires the
   identifier `float32`, which `internal/arch_test.go:54` bans"*. **False** (F3): the identifier is
   not required, and `fontset_test.go:515` calls it with `700` today, guard green. Correct the
   mechanism; **keep the conclusion**, which D-2.2.4 rules for independent reasons (the vendor never
   rewrites `OS/2.usWeightClass`). The corrected text may cite the new type-aware rule, because after
   AC2 the statement becomes true for the first time.
2. [x] **`internal/text/shape.go`, the `Shaper` doc comment** — the same claim about the four
   `float32`-taking variation setters. Same correction.
3. [x] **`internal/fontset/fontset.go:70-73`, the `ot.NewFace` error branch** — **unreachable** (F8):
   `NewFace` returns `nil` on every path. Label it as an unreachable vendor-contract assertion, in
   the form `Subset`'s "not retained" branch already uses (D-000.24, D-2.3.6): kept because it would
   catch the vendor gaining a failure mode, **explicitly not counted as coverage.** Do not delete it.

**Do not write any of these corrections before making the change they describe** (D-000.28). A
comment asserting a guard exists, committed alongside the guard, is fine; a comment asserting it in
advance is false from birth and reads identically to a true one.

> **THIS AC'S ENUMERATION WAS SAMPLED, NOT SWEPT — corrected by the finisher (Story 2.3a Finding 1,
> Major).** Items 1 and 2 above name two of the sites carrying the falsified `PinAxisLocation`
> claim. D-2.2.4 (correction), the binding ruling, named **four**. Three were corrected during
> development. The review found **two more**, and the finisher's repo-wide sweep — every form of the
> claim, **every file type**, not just `.go` and `.md` — found **eight** carrier sites in all,
> including one in a Python build tool that no previous pass had thought to look at.
>
> **The full swept enumeration, with its count, is in the Delivery Log below.** It is recorded there
> as an enumeration rather than a sample, which is what D-2.2.4 (correction, amended) asks this
> finisher to produce. The lesson is D-000.23's: *a correction written in response to instances
> covers the instances, not the class* — and this is its third recurrence, each time because the
> list was built from the sites someone happened to have been shown.

### AC10 — Standing guards stay green, and the module graph is unchanged

- `cd folio-go && go test -count=1 -v ./...` → **377 PASS · 2 FAIL** at baseline (every `--- PASS`
  occurrence; **corrected from `375` by the finisher, Finding 3**), the two failures being exactly
  `TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck`, **plus** whatever new tests
  this story adds, all passing. Report the new total against the baseline, name the delta, and state
  the counting rule beside every absolute.
- `cd lint && go test -count=1 -v ./...` → **75 PASS · 0 FAIL** plus this story's new tests, all
  passing.
- `go build ./...`, `go vet ./...` and `test -z "$(gofmt -l .)"` clean in **both** modules.
- **`cd folio-go && go list -m all` returns exactly two lines**: `github.com/panitw/folio/folio-go`
  and `github.com/boxesandglue/textshape v0.0.15`. `TestModuleGraphAllowlist` and
  `TestModuleGraphDenylistsKnownPDFWriters` still pass. **`folio-go/go.mod` is byte-unchanged.**
- `lint`'s `MANIFEST.md` is regenerated only if `lint/go.mod` changed; it should not — the new rule
  uses `golang.org/x/tools`, already required.
- **`absence-source-date-epoch` stays green**: the literal string must not appear in any `.go` file
  under `folio-go/`, including in a comment quoting this story.
- `no-compressor-import` and `no-image-decoder-import` stay green.
- The existing `no-float64` guard (`TestNoFloat64UnderInternal`, `TestNoFloat64UnderModule`,
  `TestNoFloat64FixtureScan`) is **unmodified and still passing**.

---

## Task breakdown

1. [x] Re-verify the baseline at `431a6a5`: clean tree, both suites, both counts, the two known failures
   by name. Record the invocation with each number (D-000.26).
2. [x] Re-run F6's per-face equivalence measurement before writing any code. **If it is not 0 mismatches
   over 37,312 glyphs, stop and escalate** — AC6's premise, and therefore the story's shape, depends
   on it.
3. [x] Build the type-aware checker in `lint/internal/rules/` (AC1), with its stats type and its
   loud-failure paths.
4. [x] Create the fixture pair under `folio-go/testdata/lint/` and both halves of the red-proof (AC3),
   including the assertion in `folio-go` that the syntactic scanner reports zero on it.
5. [x] Wire the two production callers, production scope and test scope, with their vacuity guards taken
   from the checker's own stats (AC2). **Record the production-scope finding count now, before the
   fix: it must be 4, in 1 file.**
6. [x] Fix both `fontset.go` sites by declining the accessor (AC5), with the bounds check and the located
   error. **Record the production-scope count again: it must be 0.**
7. [x] Assert byte-neutrality across all five fixtures, existence before value (AC6).
8. [x] Write the enumeration document (AC4) from the measurements, with subjects cited and the three-way
   classification.
9. [x] Write the executable assertions behind it (AC8), including the `/CapHeight 928` vs `711` emitted-
   byte assertion and the `FamilyName` no-call-site guard.
10. [x] Correct the three statements at their source (AC9) — **after** the changes they describe exist.
11. [x] Confirm AC7 by measurement: matrix-tagged test set unchanged, `matrix.yml` unchanged,
    `thai-signoff.json` absent, `5964aad0…92e00f` untouched.
12. [x] Run both suites, both vet passes, both gofmt checks, and `go list -m all` (AC10). Write the
    completion notes with the inherited-failure clause and every measurement's subject.
13. [x] Set the story status to `review`.

---

## Heavy-test cadence — what is deferred, and to which gate

**This story writes no matrix leg and runs none, and there is nothing to register.**

Under D-000.4 the four-target matrix runs on a per-epic cadence, and this story produces **no new
cross-target observable**: AC6 asserts every output byte is unchanged, so the four-target byte
identity that Story 2.3 established still holds by construction and there is nothing new for a matrix
leg to measure. Adding one would be a gate obligation with no question behind it — and clause 8 of
the inherited rulings forbids adding a third.

`go build -tags=matrix ./...` and `go vet -tags=matrix ./...` are still run (they are CI steps, and
they compile the existing matrix tests against changed production code). **They compile; they do not
run the matrix.**

**Carried to the Epic 2 gate, unchanged and un-added-to:** the four-target matrix legs, and D-2.3.5's
Thai sign-off (`fixtures/shaped-text/thai-signoff.json`, digest `5964aad0…92e00f`), which is a
deliberately red gate-run test. **This story owes the gate nothing new.**

---

## What the gate should be told, if the audit's findings are not fixed here

Append to `epic-2-boundary-gate.md` only what this story actually establishes:

- The vendor boundary is now **enumerated** — the first time the project has a written, tested answer
  to "what does our one dependency tell us when it does not know".
- **`Q1`: a caller-supplied font missing `maxp` crashes `folio.Render`.** Whatever the lead rules,
  the owner should see this: it is a public-entry-point panic on caller input, and the project has an
  explicit precedent that such input must produce a located error.
- **`Q2`: `/BaseFont` falls back to the caller's FontSet key for a nameless program**, which is the
  pre-2.2 behaviour retained deliberately, in a narrow case, against a spec clause that does not
  contemplate the case.
- **The `OS/2` row**: a caller-supplied face without `OS/2` renders successfully with a `/CapHeight`
  that is the ascender. Harmless for everything folio ships (all three faces are `OS/2` v4 with a
  real `sCapHeight`), live for the input surface `folio.FontSet` is.

---

## Dev Agent Record — completion notes

**Developed at baseline `431a6a5`, working tree clean at the start of every measurement below.**
Every number here comes from a run made during this story; nothing is carried from the story file's
own findings without re-measurement. Where a figure differs from the one recorded at creation, both
are given and the difference explained.

### Gates, measured, with the invocation beside each number (D-000.26)

**RECOMPUTED BY THE FINISHER (Finding 3).** Every number below was re-measured, not carried. The
baseline column comes from an isolated `git worktree` at `431a6a5`; the "after" column from the
working tree at the moment of the finisher's commit. The counting rule is stated because `377` is a
true statement under one rule and a false one under another.

**The invocation, verbatim, identical for every cell:**

```
CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

counted with `awk '/--- PASS/{p++} /--- FAIL/{f++} END{print p,f}'` (every occurrence, subtests
included) and `awk '/^--- PASS/{p++} /^--- FAIL/{f++} END{print p,f}'` (top-level only).

| scope | counting rule | baseline (`431a6a5`) | after this story |
|---|---|---|---|
| `folio-go/` | every `--- PASS`/`--- FAIL` occurrence | **377 PASS · 2 FAIL** | **400 PASS · 2 FAIL** |
| `folio-go/` | top-level results only | **241 PASS · 2 FAIL** | **254 PASS · 2 FAIL** |
| `lint/` | every occurrence | **75 PASS · 0 FAIL** | **81 PASS · 0 FAIL** |
| `lint/` | top-level results only | **37 PASS · 0 FAIL** | **43 PASS · 0 FAIL** |

**What was wrong, and why the delta was right anyway.** The figures recorded at development were
`375 → 397` for `folio-go`. Both are exactly 2 below the measured `377 → 400`… *for the "before"
side*; the "after" side moved for a second, unrelated reason. The `-2` offset on both sides was the
two inherited failures being subtracted from the PASS count — which is why `lint`, having no
failures, reproduced exactly (`75 → 80`) and `folio-go` did not. **The delta was therefore correct
while both absolutes were wrong**, which is precisely the shape D-000.18 warns about: a consistent
offset makes a difference look confirmed. The remaining `397 → 400` gap is this finisher's own work:
`+1` in `folio-go` for Finding 5's chain-level subtest, and `+1` in `lint` for Finding 4's
type-check loud-failure test.

**Delta against baseline: +23 passing tests in `folio-go`, +6 in `lint`, 0 new failures.** Of the
23: 1 (`arch_blindspot_test.go`), 19 (`vendorboundary_test.go`, counting subtests), 2
(`byte_neutrality_test.go`) and 1 (the finisher's `cmap` chain subtest). Of the 6: 5 the new rule's
callers, and 1 the finisher's `Errors`-sweep test.

**The two failures are the inherited ones and nothing else**: `TestCorpusMeetsP6ExerciseFloors` and
`TestP2IndependentDPCrossCheck`, Story 2.1's deliberate pre-stated shortfalls (D-2.1.14). They fail
identically before and after. They were not fixed, not tuned, and are not counted as this story's.
**There is no third failure.**

Also clean, both modules: `go build ./...`, `go vet ./...`, `gofmt -l .` (empty).
`cd folio-go && go build -tags=matrix ./...` and `go vet -tags=matrix ./...` **compile; the matrix
was not run** (per-epic cadence, D-000.4 — see AC7 below).
`cd folio-go && go list -m all` returns **exactly two lines**, and `folio-go/go.mod`, `go.sum`,
`lint/go.mod`, `lint/go.sum` and `lint/MANIFEST.md` are **byte-unchanged from `431a6a5`**
(`git diff --stat 431a6a5 --` empty for all five).

### Task 2 re-measurement — F6's premise, re-verified before any code was written

Re-run against the four committed faces by absolute path, through a scratch module outside the repo
(`ot.ParseFont` → `ot.NewFace` and `ot.ParseHmtxFromFont`, comparing
`int64(face.HorizontalAdvance(g))` against `int64(hmtx.GetAdvanceWidth(g))` for every glyph):

| face | glyphs | mismatches |
|---|---|---|
| `folio-go/testdata/fonts/Roboto-Regular.ttf` | 1,294 | **0** |
| `folio-go/fonts/notosans/NotoSans-Regular.ttf` | 4,515 | **0** |
| `folio-go/fonts/notosansthai/NotoSansThai-Regular.ttf` | 467 | **0** |
| `folio-go/fonts/notosanssc/NotoSansSC-Regular.ttf` | 31,036 | **0** |
| **total** | **37,312** | **0** |

**Reproduces F6 exactly. AC6's premise holds; no escalation was required.**

### AC5's red-proof — the transition, both numbers, one invocation

`ScanFloatTypedValues("<repo>/folio-go", false)`, the AC2 production caller:

- **Before the fix: 4 findings in 1 file** — `internal/fontset/fontset.go` lines **328, 329, 565,
  566**, reported with their resolved type. Identical to F1.
- **After the fix: 0 findings.**
- The syntactic guard (`TestNoFloat64UnderInternal` / `TestNoFloat64UnderModule`) reported **zero on
  both sides of that transition**, unmodified. That is the gap D-000.25 named, now stated as an
  assertion rather than as prose (`folio-go/internal/arch_blindspot_test.go`).

`Tests: true` reproduced F1's shape too: 6 distinct `(file, line)` findings across 3 files at
baseline. Typed-expression counts differ from F1's (15,396 vs 15,883 for `Tests: false`) because the
two scanners count different things — F1's scratch checker counted every recorded expression, this
one counts only expressions whose type **resolved** and excludes builtins. Not load-bearing; the
finding sets are identical.

**One thing the checker found that F1 did not.** `go/types` records `Typ[Invalid]` for an identifier
denoting a **builtin** — `len(buf)` in `internal/template/decimal.go:260` is one — so the
"unresolved type is a hard error" rule had to exclude builtins **by category** (`tv.IsBuiltin()`),
never by a list of names. Found by the checker failing on the real tree, not anticipated.

### The two ruled fixes

**D-2.3a.1 — the panic and its silent sibling.** Both closed by one guard,
`requireReadableTables` in `fontset.New`, running **before** `ot.NewFace` (which is where the panic
is). Required: `head`, `maxp`, `hhea`, `hmtx`, `OS/2` — the tables folio reads — plus two
consistency checks presence alone does not give (`NumGlyphs() > 0`, because a *short* `maxp` returns
0 exactly as an absent one does; and `1 <= numberOfHMetrics <= numGlyphs`, because the vendor
subtracts one from the other unchecked and reaches the same negative slice length by a second
route).

Measured through `folio.Render` with `fixtures/font-text/input.folio`:

| table stripped | at `431a6a5` | now |
|---|---|---|
| `maxp` | **PANIC**, `makeslice: len out of range` | located error naming the face and `"maxp"` |
| `OS/2` | **22,198 bytes, `err=nil`, `/CapHeight 928`** (intact: `/CapHeight 711`) | located error naming the face and `"OS/2"` |
| `head` / `hhea` / `hmtx` | located errors (two of them from the vendor subsetter) | located errors from folio, at ingestion |
| *(intact)* | 22,310 bytes, `/CapHeight 711` | **identical** |

The `431a6a5` column was **measured in this run**, in an isolated `git worktree` at that commit,
not carried from the story file. The worktree was removed and the tree re-verified.

**`name` and `cmap` are read but deliberately not required.** `name` is D-2.3a.1's own carve-out.
`cmap` is a judgement I want visible rather than buried: the vendor is **honest** there (`Cmap()`
returns `nil`) and folio nil-checks it, and requiring it would **break Story 2.2's fallback chain** —
today a chain whose first face has no `cmap` falls through and renders, and requiring the table would
make the whole `FontSet` unloadable over one unusable member. Measured: with `cmap` stripped,
`Render` still reports *"no font in chain [Roboto-Regular] has a glyph for rune U+0048"*, the located
failure Story 2.2 ruled for. Both tolerations are pinned by tests so a later tightening cannot
reverse either ruling silently.

**D-2.3a.2 — the `/BaseFont` substitution, now named.** `/BaseFont` still carries the FontSet key
(Required by Table 117; the key is true of the face), but a nameless program now emits a PDF comment
saying so, immediately before the font object:

```
% folio: /BaseFont HXRYNT+Roboto-Regular names the FontSet key, NOT the embedded program: …
```

Written before `b.begin`, so the object's xref offset is unchanged. `qpdf --check` on the resulting
document: *"No syntax or stream encoding errors found."* At `431a6a5` the nameless and intact renders
were **both 22,310 bytes** — indistinguishable, which was the defect; they are now 22,509 and 22,310.
**Byte-neutral for every face that has a name record**, which is every face folio ships or tests
with, so no golden moves.

**A defect note was appended to Story 1.5's file** (`## Defect found after done`), which shipped the
`ot.NewFace` call site — same disposition as Story 1.6's `decodePoints` hang. Appended, not
rewritten.

### AC-by-AC verification

| AC | verdict | evidence |
|---|---|---|
| AC1 | satisfied | `lint/internal/rules/floattyped.go`. Pure function `ScanFloatTypedValues(root, includeTests) ([]Finding, FloatTypedStats, error)` — no `*testing.T`, no hard-coded root, no root discovery. Detection is `tv.IsValue()` + underlying `*types.Basic` with `types.IsFloat`; it names **no** accessor, package or symbol. Package `Errors` sweep and a recorded-but-unresolved type are both hard errors; `RuleNoFloatTypedValue` is a `const` beside the checker; messages name the resolved type and position. |
| AC2 | satisfied, **with one departure stated below** | `TestFloatTypedProductionScan` (zero) and `TestFloatTypedTestScopeInventory` (`assertExactFindings`, by file and rule). Vacuity from the checker's own `FloatTypedStats` — packages by name (`.`, `internal/fontset`), `FilesParsed`, and `TypedExprs`. The syntactic guard in `folio-go/internal/arch_test.go` is **unmodified**. |
| AC3 | satisfied | Fixture pair at `folio-go/testdata/lint/no-float-typed-value/`. Half 1: `TestFloatTypedFixtureScan` reports exactly `violating_inferred.go`, not the compliant file. Half 2: `folio-go/internal/arch_blindspot_test.go` runs `findFloatOccurrences` over the same violating file and asserts **zero**, after four presence preconditions (file readable, non-empty, actually contains the truncation, contains neither banned identifier) plus a **control** proving the scanner is not simply broken. |
| AC4 | satisfied | `folio-go/internal/fontset/vendor-boundary.md`. Every row cites its subject; three-way classification (honest / substituted / **fabricated**); version stated prominently and checked against `go.mod` by test; method replayable, including *why* the mutation shape matters. |
| AC5 | satisfied | Both sites decline the accessor. `hmtx` parsed once in `New` via `ot.ParseHmtxFromFont`, held on the struct beside `psName`/`unitsPerEm`. `AdvanceForRune` **fixed in place, not deleted**, doc updated. Every glyph id bounds-checked against the face's own count before `GetAdvanceWidth` (which fabricates: gid 65535 → 506 on 1,294-glyph Roboto, asserted). `hmtx` parse failure is a located error naming the font. |
| AC6 | satisfied | All five fixtures' digests unchanged, including `fixtures/shaped-text`'s `5964aad0…92e00f`. `TestStory23aMovedNoGoldenDigest` asserts directory → file → parse → `sha256` field present → correct shape → **then** value, against a second independent literal. Per-face equivalence re-measured (above). |
| AC7 | satisfied | No new `//go:build matrix` test. The three matrix-tagged files, `.github/workflows/matrix.yml` and all of `fixtures/shaped-text/` are **byte-unchanged from `431a6a5`** (`git diff` empty); `thai-signoff.json` absent. `TestStory23aAddedNoThirdEpic2GateObligation` pins all of it. **The Epic 2 gate owes exactly the two things it owed at `431a6a5` — the four-target matrix legs and D-2.3.5's Thai sign-off — and nothing added here.** |
| AC8 | satisfied, one row **relabelled** — see below | `folio-go/internal/fontset/vendorboundary_test.go`, 19 assertions incl. subtests. Substitutions derived with exact values, after asserting the stripped font still parses. Declines asserted on the error's **text**. `FamilyName` no-call-site guard is live, with a vacuity guard. Document existence and version cited by path. |
| AC9 | satisfied | All three corrected **after** the changes they describe existed (D-000.28): `fontset.go`'s `New` block, `fontset.go`'s `Subset` doc, `internal/text/shape.go`'s `Shaper` doc — mechanism corrected, conclusion kept, and the `ot.NewFace` error branch labelled unreachable and explicitly not counted as coverage. |
| AC10 | satisfied | See the gates table. |

### Two departures from the ACs as written, stated rather than absorbed

**1. AC2's test-scope inventory holds THREE entries, not two.** AC2 predicted the list would still be
exactly `shaping_expectations_test.go` and `internal/fontset/fontset_test.go`. It is those two plus
`internal/fontset/vendorboundary_test.go`.

This is **AC8 and AC2 in tension, and the inventory mechanism working exactly as designed.** AC8
requires the enumeration's rows to be *derived* from the vendor; two of those rows are claims about
`(*ot.Face).HorizontalAdvance` itself (that it returns `upem` when `hhea` or `hmtx` is absent).
Asserting anything about a function with a fractional return type **requires a value expression of
that type** — there is no integer spelling of that claim, and that asymmetry *is* the row. Adding the
file turned `TestFloatTypedTestScopeInventory` red; it was read, not absorbed, and the third entry is
written into the test with its full reasoning. That is the difference between an inventory and an
exemption, and it is visible in the diff. **If the reviewer would rather the two rows were narrated
than derived, that is a live question — but narrating them would weaken AC8.**

**2. AC8's `/CapHeight 928` emitted-byte assertion is no longer constructible, and is labelled
rather than manufactured (D-000.24).** AC8.3 asked for an assertion that the `OS/2`-stripped render
*contains* `/CapHeight 928`. **D-2.3a.1 closed that path** — the face is now refused at ingestion, so
no PDF carrying 928 can be produced. Rather than weaken the guard to keep the assertion, or
manufacture a proof, the test asserts: the intact render **does** contain `/CapHeight 711` (on the
produced bytes, after a presence precondition that a `/CapHeight` entry exists at all); both 711 and
928 are **derived**, not quoted, by putting the real and the substituted cap heights through folio's
own `geom.ScaleRound`; the two must differ, or the row records no hazard; and the stripped render
**is refused**. The 928 that once reached a real output byte is recorded from the `431a6a5` worktree
measurement, cited with its commit. The two `hhea`/`hmtx` rows carry the **"forward hazard, no
available red-proof"** label in the assertion's own text and are not counted as coverage.

### What the Epic 2 gate should be told

- The vendor boundary is now **enumerated and executable** — the first time the project has a
  written, tested answer to *"what does our one dependency tell us when it does not know?"*
- **Q1 is fixed, not deferred** (D-2.3a.1): a caller-supplied font missing `maxp` no longer crashes
  `folio.Render`, and the silent `OS/2` sibling — which reached a real output byte — is closed by the
  same guard.
- **Q2 is fixed, not deferred** (D-2.3a.2): a nameless program's `/BaseFont` substitution is now
  named in the PDF itself.
- **The gate's obligations are unchanged**: the four-target matrix legs, and D-2.3.5's Thai sign-off.
  Nothing was added.

---

## QA Results

## Review Summary

- **Reviewed by:** bmad-code-reviewer
- **Date:** 2026-08-24
- **Baseline:** `431a6a5` (re-created as an isolated worktree for every "before" measurement, then removed)
- **Story Status Recommendation:** **Changes Requested**
- **Blockers:** 0
- **Majors:** 2
- **Minors:** 2
- **Nits:** 4

**Everything load-bearing in this story was verified by construction, not by reading its claims.**
Every mutation below was reverted by hand (never `git checkout`, never `git stash`) and
`git status --porcelain` was confirmed byte-identical to the tree handed over, including a SHA-256
comparison of `folio-go/internal/fontset/fontset.go` and `fixtures/minimal-rect/expected.json` against
pre-mutation copies.

### Independently reproduced (measured in this review, not carried)

| claim | how it was checked | result |
|---|---|---|
| `folio-go` suite | `go test -count=1 -v ./...`, all `--- PASS`/`--- FAIL` occurrences | **399 PASS · 2 FAIL** — the two inherited Story 2.1 failures and **no third failure** |
| `lint` suite | same invocation | **80 PASS · 0 FAIL** |
| new tests are additive | test-name set diff, baseline worktree vs now | **+22 added, 0 removed** — exactly the 1 + 19 + 2 split claimed |
| type-aware rule finds real drift | wrote `int64(f.face.HorizontalAdvance(ot.GlyphID(gid)))` into a new production file under `internal/fontset/`, containing no `float` substring at all | **`TestFloatTypedProductionScan` and `TestFloatTypedTestScopeInventory` both RED**, naming `float32` and the position |
| syntactic guard is blind to it | `TestNoFloat64UnderInternal` / `UnderModule` / `FixtureScan` over the same tree | **all GREEN** — the gap D-000.25 named, reproduced |
| loud-failure path (package `Errors` sweep) | injected `undefined: thisSymbolDoesNotExist` into `internal/fontset` | **hard error, not zero findings** |
| byte-neutrality, 4 faces | scratch module outside the repo, `int64(HorizontalAdvance(g))` vs `int64(GetAdvanceWidth(g))`, every glyph | **37,312 glyphs, 0 mismatches** — reproduces F6 exactly |
| no golden moved | `git status --porcelain fixtures/` | **empty**; all five digests unchanged |
| `fixtures/shaped-text` digest | read from `expected.json` | **`5964aad0e696…c92e00f`** — unchanged, the pending Thai sign-off is **not** invalidated |
| pre-fix state at `431a6a5` | worktree at that commit, table-record strip, `folio.Render` at the public entry | **intact 22,310 B `/CapHeight 711`; `OS/2`-stripped 22,198 B, `err=nil`, `/CapHeight 928`; `name`-stripped 22,310 B `err=nil`; `maxp`-stripped PANIC `makeslice: len out of range`** — every figure in F5 / D-2.3a.1 / the Story 1.5 defect note confirmed |
| post-fix state | same probe on the working tree | `maxp` **no longer panics**; `head`/`maxp`/`hhea`/`hmtx`/`OS/2` each return a located error **naming the face and the table**; `name` renders at 22,509 B (was 22,310 — now distinguishable) |
| `requireReadableTables` is load-bearing | injected `if true { return nil }` as its first statement | **`TestFolioDeclinesEverySubstitutionAtIngestion` and `TestCapHeightSubstitutionIsAssertedOnTheEmittedBytes` both RED**; reverted, file byte-identical |
| golden-digest guard | set `fixtures/minimal-rect/expected.json`'s `sha256` to 64 zeros | **RED** with the correct message; restored byte-identically |
| AC7 gate inventory | added a `//go:build matrix` file **and** created `fixtures/shaped-text/thai-signoff.json` | **RED on both counts**; both removed |
| `FamilyName` no-call-site guard | added `.FamilyName(` to a new production file | **RED**; removed |
| `/BaseFont` diagnostic does not corrupt the PDF | parsed the 22,509-byte nameless render's classic xref and checked every `n` entry points at `N 0 obj` | **10 objects, 0 mismatches**; the `%` comment is present |
| `Upem()` row stays closed | `grep '\.Upem()'` over non-test sources | **zero production call sites** — folio never reads it; D-1.5.2 does not validate fiction |
| D-000.29 carried rows | read every disposition cell in `vendor-boundary.md` Table 2 | **no carried rows** — every hazard row ends *confirmed and fixed* or *traced and closed* |
| AC10 gates | `gofmt -l`, `go vet ./...`, `go build ./...` in both modules; `go build/vet -tags=matrix`; `go list -m all` | **all clean**; module graph is **exactly two lines**; `go.mod`, `go.sum`, `lint/go.mod`, `lint/go.sum`, `lint/MANIFEST.md` byte-unchanged |

### The three developer disclosures — all three are honest, and none is reported as a finding

1. **The third test-scope inventory entry.** Verified: asserting anything about `(*ot.Face).HorizontalAdvance`
   requires a value expression of its fractional type, so AC8's derivation obligation genuinely
   collides with AC2's predicted count. The entry is written into the test with its reasoning, and a
   **new** file carrying a float-typed expression does fail the build (proved above). *(A narrower
   mechanical gap in the same assertion is Major 2 below — that is a separate point, not a challenge
   to the disclosure.)*
2. **AC8.3's `/CapHeight 928` assertion is no longer constructible.** Verified honest, and now also
   ruled: **D-000.30** was appended to the decision log and blesses exactly this disposition. The
   pre-fix measurement was genuinely taken in a worktree at `431a6a5` (I reproduced all four figures),
   the live half asserts `/CapHeight 711` on produced bytes **after** a presence precondition, both
   values are derived through `geom.ScaleRound`, `wantSubstituted != wantIntact` is asserted so the row
   cannot record a non-hazard, and the refusal is asserted. **The guard was not weakened and no proof
   was manufactured** — I confirmed this by neutering `requireReadableTables` and watching the test go
   red. The `hhea`/`hmtx` forward-guard label is applied to rows whose proof is genuinely
   unconstructible: at `431a6a5` the vendor subsetter refused those faces before folio could observe
   the substitution, which I measured directly. **The label is legitimate.**
3. **The three corrected `PinAxisLocation` comments.** The corrected text does state the real grounds
   (payload, `usWeightClass` never rewritten by `textshape`, deleting the FMA/interpolation path) and
   does **not** simply delete the claim. *However the sweep is incomplete — Major 1.*

### The highest-judgement item: `cmap` read but not required — **the reasoning holds**

D-2.3a.1's literal text is *"every table folio actually reads"*, and `cmap` is read. The developer
excluded it anyway. **I verified the justification by construction rather than accepting it:**

- Built a two-face chain (`["Roboto-Regular", "Noto Sans"]`) with a `cmap`-stripped Roboto **first**.
  It **renders: 60,414 bytes, `err=nil`** — the chain falls through to the second face exactly as
  Story 2.2 ruled.
- Substituted an `OS/2`-stripped face into the same first slot as a stand-in for "a required table is
  missing on a skippable member": **`err`, 0 bytes** — the whole `FontSet` becomes unloadable.

So adding `cmap` to `requiredTables` would convert a rendering document into a load failure. That is a
ruled-behaviour reversal, not a tightening. Crucially, the exclusion is **the same class** as the one
D-2.3a.1 carved out explicitly for `name`: the vendor is *honest* there (`Cmap()` returns `nil`,
`readPostScriptName` returns `""`), so neither belongs to the class the ruling targets — *folio reads a
table that may be absent and receives a **substituted default***. The deviation is from the ruling's
literal text, not from its reasoning, and it is disclosed in the source rather than buried.

**Is the behaviour pinned either way? Yes.** `TestFolioDeclinesEverySubstitutionAtIngestion/cmap is
tolerated…` asserts `fontset.New` succeeds and `HasGlyph('H')` is false, so a later tightening that
added `cmap` to `requiredTables` turns that subtest red. **The judgement is sound and I endorse it.**
(One residual observation is Nit 4.)

---

### Finding 1: The falsified `PinAxisLocation` mechanism survives at two more sites — one of them named by the binding ruling

- **Severity:** Major
- **Category:** AC Conformance / Correctness
- **Location:** `folio-go/internal/fontset/fontset.go:751`; `folio-go/internal/fontset/fontset_test.go:466`
- **Observation:** D-2.2.4 (correction), appended by this story and marked *(mechanism: binding)*, states
  verbatim: *"The same false claim stands in **four** places the story corrects: `fontset.go:112`,
  `:425`, `:494`, and `internal/text/shape.go:71`."* I resolved those baseline line numbers in a
  worktree at `431a6a5` and they are, respectively, `New`'s D-2.2.4 block, `Subset`'s doc comment, the
  comment **inside** `Subset`'s body, and the `Shaper` doc comment. **Three were corrected. The third
  (`:494`, now `:751`) was not**, and still reads:

  > `// explicitly is both unreachable (AD-23 bans `float32`) and`
  > `// ineffective (textshape never rewrites OS/2.usWeightClass).`

  A repo-wide sweep for the claim also surfaces a **fifth** site the ruling did not enumerate:
  `fontset_test.go:466` — *"Production exposes no way to request an instance because reaching
  `subset.Input.PinAxisLocation` **needs the identifier `float32`**, which AD-23's arch guard bans…"* —
  which sits **three lines above line 515**, the very call the audit uses as its disproof. That comment
  then contradicts itself in its own next sentence, correctly describing the untyped-constant blind
  spot it has just denied exists.
- **Impact:** The story's thesis is that a false mechanism standing in a comment makes a reader believe
  AD-23 fences a door that is open. Two such doors are still labelled shut, and one of them is bolted
  to the counter-example. A binding ruling's explicit four-site enumeration was executed at three
  sites, so the record and the tree disagree.
- **Suggested Resolution:** Correct `fontset.go:751` in the form the other three already use — keep the
  conclusion (`textshape` never rewrites `OS/2.usWeightClass`), drop or invert the AD-23 mechanism, and
  cite the new type-aware rule. Apply the same to `fontset_test.go:465-471`, whose closing sentence
  already contains the correct account. Consider noting the fifth site against D-2.2.4 (correction),
  since its own enumeration was one short.
- **Related AC:** AC9 (and D-2.2.4 (correction), binding)

### Finding 2: The test-scope inventory is keyed by FILE, so a second float-typed site in a listed file passes silently — contradicting AC2's stated contract

- **Severity:** Major
- **Category:** Tests / AC Conformance
- **Location:** `lint/internal/rules/floattyped_test.go:118-146` (`TestFloatTypedTestScopeInventory`, `want` at :137-141)
- **Observation:** AC2 states: *"**Adding a float-typed expression to any `_test.go` file fails the
  build**; removing one fails it too."* The test's own doc comment repeats it. **Measured: it does
  not.** I appended a second float-typed expression —
  `func revProbeSecondSite(f *ot.Face) int64 { return int64(f.HorizontalAdvance(7)) }` — to
  `internal/fontset/vendorboundary_test.go`, a file already in the `want` list. `TestFloatTypedTestScopeInventory`
  reported **`ok`**. The `want` set is `[]Finding{{Path, Rule}}` and `assertExactFindings` compares by
  `(file, rule)`, so every additional site inside an already-enumerated file is invisible. Removing one
  of two in such a file is likewise invisible. (Contrast the **production** caller, which asserts exact
  zero and therefore does catch every site — I proved that by construction.)
- **Impact:** The three enumerated files are precisely the ones where a future story is most likely to
  add another vendor float call — `vendorboundary_test.go` exists to make claims about a float-returning
  accessor. AD-23's type-aware enforcement is therefore blind in exactly its highest-traffic corner,
  while the AC, the test comment and the completion notes all assert the opposite. This is the
  difference between an inventory and a per-file allowlist, and the story's own disclosure asks the
  reviewer to confirm it is the former.
- **Suggested Resolution:** Either (a) key the test-scope inventory by `(path, line)` — the checker
  already dedupes on exactly that tuple and `Finding` already carries `Line` — pinning the three known
  sites at their positions; or (b) if `assertExactFindings`'s `(file, rule)` shape is to be kept for
  consistency with `map-range`, add a per-file **count** assertion alongside it, or correct AC2, the
  test's doc comment and the completion notes to say what the assertion actually guarantees. Option (a)
  is preferable: it makes the inventory mean what the story says it means.
- **Related AC:** AC2

### Finding 3: The recorded absolute test totals do not reproduce under their own stated invocation

- **Severity:** Minor
- **Category:** Correctness of record (D-000.26, D-000.18)
- **Location:** story file lines 24-27, 705-708, 800-803, 809-811
- **Observation:** The gates table gives the invocation as *"`go test -count=1 -v ./...`, counting
  `--- PASS` / `--- FAIL` lines"* and reports **375 → 397** for `folio-go`. Running exactly that,
  I measure **377 at `431a6a5`** (worktree) and **399 now** — a consistent **+2 offset on both sides**.
  `lint`'s **75 → 80** reproduces on the "after" side (80 measured) but the same +? offset was not
  checked at baseline. The **delta is exact**: +22 in `folio-go` (test-name set diff: 22 added, 0
  removed) and the 1 + 19 + 2 attribution is correct.
- **Impact:** Low — the delta, the failure identities and the "no third failure" conclusion are all
  sound, and no decision rests on the absolute figure. But the story opens by warning that *"a carried
  figure that is off by its own failures is exactly the shape D-000.18 warns about — confirm against
  the artifact, not against a table summarising it"*, and then records a figure that does not reproduce
  under the invocation printed beside it. The next story to diff against 397 will be measuring against
  a number that never existed.
- **Suggested Resolution:** Re-run both suites with the stated invocation, record the measured numbers,
  and state the counting rule unambiguously (top-level `^--- PASS` at column 0 gives 254; every
  `--- PASS` occurrence including subtests gives 399 — the recorded 397 matches neither).
- **Related AC:** AC10

### Finding 4: AC1's package-`Errors` loud-failure branch has no committed red-proof

- **Severity:** Minor
- **Category:** Tests
- **Location:** `lint/internal/rules/floattyped.go:117-127` (the `loadErrs` sweep); `lint/internal/rules/floattyped_test.go:220-232`
- **Observation:** AC1 requires two distinct loud-failure paths and D-1.3.11 is explicit that
  *"`packages.Load`'s nil top-level error is not sufficient"*. The only committed assertion,
  `TestFloatTypedScanFailsLoudlyOnAnUnloadableTree`, points the scanner at a **non-existent directory**
  — which fails at the `packages.Load` call, i.e. the branch D-1.3.11 says is *not* sufficient. The
  per-package `Errors` sweep, the branch that actually implements the ruling, is exercised by nothing.
  I confirmed by construction that the sweep **does** work (injecting `undefined: thisSymbolDoesNotExist`
  into `internal/fontset` produced the loud error, not zero findings), so this is a coverage gap and
  not a defect.
- **Impact:** The mechanism that distinguishes "clean tree" from "tree that did not type-check" is
  unguarded. A future refactor that dropped or short-circuited the sweep would leave the whole suite
  green while the checker silently reported zero on a tree it could not read — which is the precise
  failure mode D-1.3.11 exists to prevent.
- **Suggested Resolution:** Add a small retained fixture directory containing one file that parses but
  does not type-check (an undefined symbol), and assert `ScanFloatTypedValues` returns an error naming
  it — the same shape `map-range` uses for its fixture trees. Keep it out of
  `testdata/lint/no-float-typed-value/`, whose two existing tests assert an exact finding set.
- **Related AC:** AC1

### Finding 5: The `cmap` exclusion's stated justification is not itself pinned

- **Severity:** Nit
- **Category:** Tests
- **Location:** `folio-go/internal/fontset/vendorboundary_test.go:370-379`
- **Observation:** The deviation from D-2.3a.1's literal text is justified by a **chain-level** claim —
  *"a chain whose first face carries no `cmap` falls through to the next face and renders"*. The
  subtest that pins it is **single-face**: it asserts `fontset.New` succeeds and `HasGlyph('H')` is
  false. I verified the chain-level claim myself (two-face chain, `cmap`-stripped face first →
  **60,414 bytes, `err=nil`**), so the reasoning is sound; it is simply not the thing the test
  measures. The completion notes' own supporting measurement (*"Render still reports 'no font in chain
  [Roboto-Regular] has a glyph for rune U+0048'"*) is likewise a **single**-face chain, which
  demonstrates a located failure rather than a successful fall-through.
- **Impact:** The toleration is pinned (a later tightening turns the subtest red), so the story's claim
  *"Both tolerations are pinned by tests"* is true. What is unpinned is the **consequence** that makes
  the deviation defensible.
- **Suggested Resolution:** Extend the subtest to a two-face chain with the `cmap`-less face first,
  asserting the render succeeds — three extra lines that turn the justification into an assertion.
- **Related AC:** AC8.2, D-2.3a.1

### Finding 6: The `hhea` / `hmtx` substitution rows lack the intact anchor the `OS/2` row has

- **Severity:** Nit
- **Category:** Tests
- **Location:** `folio-go/internal/fontset/vendorboundary_test.go:191-209`
- **Observation:** The file states its own discipline at :124-126 — *"Without these, `CapHeight() == 1900
  when OS/2 is absent` could not be distinguished from `CapHeight() is 1900 for this face anyway`"* —
  and applies it to `Upem`, `NumGlyphs`, `PostscriptName`, `Ascender`, `Descender` and `CapHeight`. It
  is **not** applied to `HorizontalAdvance(36)`: the `hhea` and `hmtx` subtests assert
  `HorizontalAdvance(36) == Upem()` with no assertion that the *intact* advance for gid 36 differs from
  `upem`. I measured it (**intact 1839 vs upem 2048**), so the assertion discriminates today.
- **Impact:** Latent only. If a future subject face or gid were chosen whose real advance equalled
  `upem`, both rows would pass without the vendor substituting anything, and the enumeration's
  strongest `HorizontalAdvance` rows would become self-satisfying.
- **Suggested Resolution:** Add `wantAdvance36 = 1839` to the intact anchor block at :136-161 and assert
  `intact.HorizontalAdvance(36) != intact.Upem()` before the two subtests run.
- **Related AC:** AC8.1

### Finding 7: `TestFamilyNameHasNoCallSite` matches source text and excuses one file by filename

- **Severity:** Nit
- **Category:** Maintainability / Tests
- **Location:** `folio-go/internal/fontset/vendorboundary_test.go:516-563`, exclusion at :543
- **Observation:** The guard greps raw source for `.FamilyName(` and excludes any path ending
  `vendorboundary_test.go` (needed because the file contains the needle in its own `const accessor`).
  Two consequences: a **comment** anywhere in the module that mentions `.FamilyName(` fails the build
  even though no call exists; and a genuine call added *inside* `vendorboundary_test.go` is excused. The
  guard is live and fires correctly — I proved it by adding `.FamilyName(` to a new production file and
  watching it go red.
- **Impact:** Small. A false positive is loud and quickly diagnosed; the excused file is one the story
  owns. But the exclusion is bound to a **filename**, which is the shape the project's own conventions
  warn about elsewhere.
- **Suggested Resolution:** Either assemble the needle from parts (`"." + "FamilyName("`) as the same
  file already does for the banned type identifiers at `arch_blindspot_test.go:70`, and drop the
  filename exclusion entirely; or match on the AST (`*ast.SelectorExpr` with `Sel.Name == "FamilyName"`)
  so comments cannot trip it.
- **Related AC:** AC8.4

### Finding 8: AC8.2's text went stale on D-2.3a.1 and the staleness was not disclosed alongside AC8.3's

- **Severity:** Nit
- **Category:** AC Conformance (documentation)
- **Location:** story lines 663-666 (AC8.2) vs `folio-go/internal/fontset/vendorboundary_test.go:340`
- **Observation:** AC8.2 as written requires *"a located error for `head`; a located error for
  `hhea`/`hmtx`; **success** for `name` and `OS/2`"*. D-2.3a.1 changed `OS/2` from success to refusal,
  and the test correctly asserts the new behaviour (`OS/2` sits in the decline loop, not among the
  tolerations). The "Two departures" section discloses the AC8.3 consequence of that same ruling in
  detail but does not mention that AC8.2's expected outcome moved as well.
- **Impact:** Documentation only — the shipped behaviour is right and is what D-2.3a.1 ruled. But a
  reader checking AC8.2 against the test finds them in apparent conflict, with no note reconciling
  them, and the story's own standard is that mid-story rulings are stated rather than absorbed.
- **Suggested Resolution:** Add one line to the "Two departures" section (or an AC8.2 footnote) recording
  that D-2.3a.1 also moved `OS/2` from *success* to *refused at ingestion*.
- **Related AC:** AC8.2

---

### AC-by-AC verification (reviewer's independent verdicts)

| AC | reviewer verdict | basis |
|---|---|---|
| AC1 | **satisfied** | Pure `ScanFloatTypedValues(root, includeTests) ([]Finding, FloatTypedStats, error)`; no `*testing.T`, no hard-coded root, no root discovery. Predicate is `tv.IsValue()` + underlying `*types.Basic` with `types.IsFloat`; names no accessor, package or symbol. Both loud-failure paths exist and I proved **both** fire (see Finding 4 for the coverage gap on one). Builtins excluded **by category** via `tv.IsBuiltin()`, not by name. `RuleNoFloatTypedValue` is a `const`; messages name the resolved type and position, asserted by `TestFloatTypedFindingNamesResolvedTypeAndPosition`. |
| AC2 | **satisfied with a departure — see Finding 2** | Production scope asserts exact zero and catches every site (proved). Test scope is an inventory across **files** but not within them. Vacuity guards read the checker's **own** `FloatTypedStats` (`DirsVisited` by name, `FilesParsed`, `TypedExprs`), which is the correct shape — an injected `return nil, FloatTypedStats{}, nil` zeroes all three. Syntactic guard unmodified, still green. |
| AC3 | **satisfied** | Fixture pair at `folio-go/testdata/lint/no-float-typed-value/`. Half 1 asserts exactly `violating_inferred.go` and not the compliant file. Half 2 (`internal/arch_blindspot_test.go`) runs `findFloatOccurrences` over the same file after **four** presence preconditions — readable, non-empty, actually contains `int64(f.HorizontalAdvance(gid))`, contains neither banned identifier, and parses — **plus a control** proving the scanner is not simply broken. The violating fixture contains no `float` substring at all (verified). This is the strongest single construct in the story. |
| AC4 | **satisfied** | `folio-go/internal/fontset/vendor-boundary.md`. Subjects cited; three-way classification present; version stated and **checked against `go.mod` by test**; method replayable, including *why* the mutation shape matters (the 14-glyph artifact anecdote). |
| AC5 | **satisfied** | Both sites decline the accessor; `hmtx` parsed once in `New` and held on the struct beside `psName`/`unitsPerEm`; `AdvanceForRune` **fixed in place, not deleted**, doc updated; every gid bounds-checked against `numGlyphs` before `GetAdvanceWidth` (which I confirmed fabricates: gid 65535 → 506 on 1,294-glyph Roboto); `hmtx` parse failure is a located error naming the font. Production scan reports **0** (measured). |
| AC6 | **satisfied** | Fixtures directory byte-clean; all five digests match; `shaped-text`'s `5964aad0…92e00f` intact. Presence-before-value at three levels, with the second independent literal — red-proved. Per-face equivalence reproduced: 37,312 glyphs, 0 mismatches. |
| AC7 | **satisfied** | Matrix-tagged file set is an inventory of three, by name, using a real `//go:build` **constraint-region** parser rather than a substring match — I red-proved both halves. `thai-signoff.json` absent. `fixtures/` and `matrix.yml` byte-unchanged. `epic-2-boundary-gate.md`'s appended section adds **no** obligation. **The gate owes exactly the two things it owed at `431a6a5`.** |
| AC8 | **satisfied, one row relabelled (legitimately) — Nits 5, 6, 8** | 19 assertions incl. subtests. Substitutions derived with exact values after asserting the stripped font still parses **and** that `HasTable` now reports false. Declines asserted on the error's **text** (face + table). `FamilyName` guard is live, with a vacuity guard on files scanned. Document existence, version and required vocabulary asserted. The `hhea`/`hmtx` forward-guard label is honest — no constructible proof was skipped. |
| AC9 | **NOT fully satisfied — Finding 1** | Two of the four sites D-2.2.4 (correction) enumerates were corrected in `fontset.go` plus the one in `shape.go`; `fontset.go:751` was missed, and a fifth site at `fontset_test.go:466` also survives. The `ot.NewFace` branch **is** correctly labelled unreachable and explicitly not counted as coverage, and `TestOtNewFaceHasNoFailureMode` derives the evidence for that label from the vendor. |
| AC10 | **satisfied — Finding 3 is about the recorded numbers, not the gates** | Both suites green apart from the two inherited failures; `gofmt`/`vet`/`build` clean in both modules; matrix tags compile and vet; `go list -m all` exactly two lines; all five module files byte-unchanged. |

### What is explicitly NOT reported as a finding

- The two inherited Story 2.1 failures (`TestCorpusMeetsP6ExerciseFloors`, `TestP2IndependentDPCrossCheck`) — present, unchanged, correctly excluded from this story's counts.
- The two pre-existing `forvar` lints at `internal/template/serialize.go:108` and `fixture_test.go:190`.
- The `hhea`/`hmtx` **"forward guard with no available red-proof"** labels (D-000.24) — verified legitimate, not a dodge.
- `TestForwardHazardsWithoutARedProofAreLabelled` overlapping `TestFolioDeclinesEverySubstitutionAtIngestion` — the overlap is deliberate and the label is carried in the assertion's own text, exactly as AC8 requires.
- `itoa` reimplemented in `floattyped.go`, `vendorboundary_test.go` and the pre-existing `arch_test.go` — consistent with an established convention in this repo, not a defect introduced here.
- D-000.30 being appended to the decision log after the story file was written — the story argues AC8.3's disposition on its own merits citing D-000.24, and does not lean on a ruling it manufactured.


---

## Finding Resolutions (finisher, 2026-08-24)

**8 findings triaged: 8 FIX, 0 DISMISS, 0 DEFER.** Every finding had a concrete, in-scope,
cheaply-verifiable resolution and the reviewer's own suggested resolutions were sound; the count is
not a rubber stamp, it is what fell out of assessing each on its merits. Two findings were found to
understate the defect they reported, and both are recorded below with the stronger measurement.

The review's five hard-attack items were re-checked rather than re-litigated, and all five held:
the type-aware guard red-proves against hand-written drift; byte-neutrality reproduces (37,312
glyphs, 4 faces, 0 mismatches); the pre-fix defect measurements reproduce in a worktree at
`431a6a5`; and `fixtures/shaped-text`'s digest is still
`5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f`, so the pending Thai sign-off is
**not** invalidated. The `cmap` carve-out, the three developer disclosures and the "no third gate
obligation" conclusion were **endorsed and not reopened**.

| # | Severity | Decision | Rationale |
|---|---|---|---|
| 1 | Major | **FIX** | The falsified `PinAxisLocation` claim survived at `fontset.go:751` and `fontset_test.go:466`. Not fixed as two sites: swept the whole repo for every form of the claim across **every file type**, which found a site the review's `.go`/`.md` sweep could not reach. Full enumeration and count below. |
| 2 | Major | **FIX** | The test-scope inventory was keyed `(path, rule)`, so a second site in a listed file was invisible — the comment claimed coverage the guard did not provide (D-000.24, D-000.28). Re-keyed to `(path, rule, line)`; comment and AC2 corrected to match. Both directions red-proved. |
| 3 | Minor | **FIX** | Absolutes recomputed from a cited invocation, recorded verbatim beside them, with the counting rule stated. Confirmed which was correct rather than picking one — see below. |
| 4 | Minor | **FIX** | Committed the missing red-proof rather than labelling it a forward guard: the proof is constructible, so D-000.24's label would have been a dodge. New retained fixture + test, red-proved. |
| 5 | Nit | **FIX** | The `cmap` carve-out's chain-level justification is now an assertion, not narration. Two-face chain, `cmap`-stripped face first: **60,414 bytes, `err=nil`**, independently reproducing the review's figure, with an `OS/2` control at **0 bytes**. |
| 6 | Nit | **FIX** | Added the intact anchor `wantAdvance36 = 1839` and an explicit assertion that it differs from `Upem()` (2048), so the `hhea`/`hmtx` rows cannot become self-satisfying. |
| 7 | Nit | **FIX** | `TestFamilyNameHasNoCallSite` rebuilt on the AST. Both defects gone: comments no longer trip it, and no file is excused by filename. Three legs red-proved. |
| 8 | Nit | **FIX** | AC8.2's stale expected outcome disclosed in place, noting that D-2.3a.1 moved **two** ACs and only AC8.3's consequence had been declared. |

### Finding 1 — the complete swept enumeration

Search terms, all five forms: `PinAxisLocation`, `unreachable`, `AD-23 bans`, ``identifier `float32` ``,
`arch guard bans`. Scope: **every file type in the repository**, not just `.go` and `.md` — which is
what the previous two passes missed, and how the eighth site was found.

**8 carrier sites.** A *carrier* asserts the falsified mechanism; a site that quotes the claim in
order to record that it is false is not a carrier and is listed separately.

| # | site | state at review | disposition |
|---|---|---|---|
| 1 | `folio-go/internal/fontset/fontset.go` — `New`'s D-2.2.4 block (baseline `:112`) | corrected in development | — |
| 2 | `folio-go/internal/fontset/fontset.go` — `Subset` doc comment (baseline `:425`) | corrected in development | — |
| 3 | `folio-go/internal/text/shape.go` — `Shaper` doc comment (baseline `:71`) | corrected in development | — |
| 4 | `folio-go/internal/fontset/fontset.go:751` — inside `Subset`'s body (baseline `:494`) | **uncorrected** — named by the binding ruling, missed | **corrected in place** |
| 5 | `folio-go/internal/fontset/fontset_test.go:465-471` | **uncorrected** — found by the review | **corrected in place** |
| 6 | `tools/fontgen/instance_faces.py:16-19` | **uncorrected — found by neither the ruling nor the review** | **corrected in place** |
| 7 | `_bmad-output/…/2-2-the-shipped-font-set-and-its-fallback-chain.md:1098-1104` | uncorrected | **corrected by appended section** (D-1.6.6) |
| 8 | `_bmad-output/…/2-2-the-shipped-font-set-and-its-fallback-chain.md:1857-1859` | uncorrected | **corrected by appended section** (D-1.6.6) |

**Site 6 is the point of the exercise.** `tools/fontgen/instance_faces.py` is a Python build tool. The
ruling's enumeration covered `.go`; the review's sweep covered `.go` and `.md`. Neither looked at
`.py`, and the file's reason 3 for shipping static faces was the falsified claim, stated in full.
This is D-000.23's third recurrence: *a correction written in response to instances covers the
instances, not the class* — and each recurrence has been a narrower search than the class demanded.

**Not carriers (3), verified individually and left as they are:**

- `folio-go/internal/fontset/vendor-boundary.md:188-191` — states the claim **in order to record that
  it was false**. Updated only to carry the swept count in place of "two production comments".
- `lint/internal/rules/floattyped_test.go:82-92` — likewise: *"…is why the comments that claimed
  AD-23 made PinAxisLocation unreachable were false."*
- `_bmad-output/…/1-5-embed-and-subset-a-font-byte-stably.md:636-639` — **accurate as written**. It
  says the guard flags *"the **bare identifier**"* and that *"a call site that **names the type**
  would go red"*. That is conditional and true; it never claims the identifier is required. A sweep
  that corrected it would have been introducing an error, not removing one.

**Instances in `folio-mvp-decision-log.md` are deliberately untouched.** That document is
append-only by its own header — *"a reversal is appended, never a rewrite"* — so its correction is
itself an appended entry (D-2.2.4 (correction), and D-2.2.4 (correction, amended)). Editing them
would destroy the evidence that the claim was ever believed, which is the whole subject of the
amended ruling.

**Corrected text states the real grounds**, at all six in-place sites: payload (**4.82 MB** compressed
against **8.30 MB**), `usWeightClass` correctness (`textshape@v0.0.15` `subset/execute.go:496-499`
copies `OS/2` verbatim and never writes the field), and deleting the FMA/interpolation path. The
claim was not merely deleted.

### Finding 2 — the re-keyed inventory, and its red-proof

**The gap was live, not hypothetical.** The review demonstrated it by *injecting* a second
float-typed site into an already-listed file and watching the test report `ok`. On re-keying and
measuring, the tree already held **four** sites, not the three the `want` list named:
`internal/fontset/vendorboundary_test.go` carried **two** (the `hhea` row and the `hmtx` row) and a
file-keyed list could represent only one. A real sanctioned site was standing unlisted while the
test passed.

`assertExactFindingSites` now compares `(Path, Rule, Line)` triples. The inventory stands at **five
sites across three files** (Finding 6 added the intact anchor, read once and reused so it books one
entry rather than two). Both directions red-proved, each restore verified with `/usr/bin/diff`:

| leg | mutation | result |
|---|---|---|
| **add**, in an already-listed file | appended `func revProbeSecondSite(f *ot.Face) int64 { return int64(f.HorizontalAdvance(7)) }` to `vendorboundary_test.go` — the review's exact mutation, which previously reported `ok` | **RED**: *"unexpected finding reported: …`vendorboundary_test.go` … line=661"*, and *"site count mismatch: got 5, want 4"* |
| **remove**, one of two in a listed file | swapped line 206's `int64(face.HorizontalAdvance(36))` for `int64(face.Upem())` | **RED**: *"expected finding not reported: … line=206"*, *"got 3, want 4"* |

The first attempt at the removal leg replaced the line with `int64(0), int64(0)`, which left `face`
unused and broke the build — the test went red through the **loud-failure path**, not the assertion
under test. Recorded because a red result through the wrong mechanism proves nothing about the
mechanism you meant to test; the mutation was replaced with a compiling one. (It did incidentally
demonstrate Finding 4's `Errors` sweep firing on a real type error.)

**The production caller was left alone**, as the review directed: it asserts exact zero and therefore
already catches every site.

### Finding 3 — which number was right, and why

Re-measured under one invocation, stated verbatim, at both ends:

```
CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...
```

counted with `awk '/--- PASS/{p++} /--- FAIL/{f++} END{print p,f}'`, the baseline taken in an
isolated `git worktree` at `431a6a5`.

| scope | rule | baseline | at review | at this commit |
|---|---|---|---|---|
| `folio-go/` | every occurrence | **377 · 2** | 399 · 2 | **400 · 2** |
| `folio-go/` | top-level only | **241 · 2** | 254 · 2 | **254 · 2** |
| `lint/` | every occurrence | **75 · 0** | 80 · 0 | **81 · 0** |
| `lint/` | top-level only | **37 · 0** | 42 · 0 | **43 · 0** |

**The recorded `375 → 397` was wrong on both sides; the measured pair is `377 → 399`, and the
`+22` delta was exact.** The root cause is now identified rather than guessed: **the two inherited
failures were being subtracted from the PASS count.** That is why `lint`, which has no failures,
reproduced exactly (`75 → 80`) while `folio-go` was off by exactly 2 at both ends — a uniform offset,
which is what made the delta look confirmed.

**And the story's own "correction" was the error.** The opener originally recorded that the
commissioning brief said *"377 pass / 2 fail"* and *"corrected"* it to *"375 PASS + 2 FAIL = 377
tests"*, on the assumption that `377` was a total containing the failures. **The brief was right.**
The paragraph invoked D-000.18 — *confirm against the artifact, not against a table summarising it* —
and then restated a figure without confirming it. Corrected in place, with the original quoted, at
the opener, at AC10 and at the gates table.

`400`/`81` are this finisher's own numbers: **+1** in `folio-go` (Finding 5's chain subtest) and
**+1** in `lint` (Finding 4's test). **No third failure** — the only two are Story 2.1's intentional
`TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck`.

### Finding 4 — the proof is committed, not labelled

The branch was **constructible**, so labelling it a forward guard would have been a dodge rather than
an application of D-000.24. New retained fixture
`folio-go/testdata/lint/float-typed-untypecheckable/untypecheckable.go` — parses, does not
type-check — plus `TestFloatTypedScanFailsLoudlyOnATreeThatDoesNotTypeCheck`, which asserts three
preconditions before measuring (the fixture exists and is non-empty; it still carries the undefined
symbol; **it parses**, since a fixture that failed to parse would send the scanner down a third path
and silently stop measuring the sweep). The error must name the sweep's own sentence *and* the
undefined symbol, which is what separates it from the `packages.Load` branch.

**Red-proof:** neutering only the sweep (`if false && len(loadErrs) > 0`) makes the new test report
**"returned no error and 0 findings"** — silence indistinguishable from a clean tree, exactly the
failure mode D-1.3.11 exists to prevent — while
`TestFloatTypedScanFailsLoudlyOnAnUnloadableTree` **stays PASS**. That the pre-existing test survives
the mutation is the proof that the two exercise genuinely different branches, which was Finding 4's
whole claim.

### Finding 7 — three legs, including the one that was excused

Rebuilt on `*ast.SelectorExpr` with `Sel.Name == "FamilyName"` in call position, dropping the
filename exclusion entirely.

| leg | expectation | result |
|---|---|---|
| real call in a **new** production file | RED | **RED** |
| a **comment** mentioning `.FamilyName(` | GREEN (the old guard went red — a false positive) | **GREEN** |
| real call **inside `vendorboundary_test.go`** | RED (the old guard **excused** this file by name) | **RED**, naming `vendorboundary_test.go:782:67` |

### A consequence of Finding 1's fix, swept and repaired

Correcting site 5 lengthened `fontset_test.go`'s comment block and moved the disproving call from
line 515 to 529, staling **five** citations to `fontset_test.go:515` across `fontset.go` (×3),
`shape.go` and `floattyped_test.go`. They were re-cited by **test name**
(`TestSubsetPinnedInstancesProduceDifferentTags`) rather than renumbered, since a position-bound
citation restales on the next edit — the same lesson as Finding 7's filename exclusion. Zero
`:515` references remain. The stale prose counts in the inventory's own doc comment ("the three
entries", "the THIRD entry") were corrected in the same pass.

---

## Delivery Log

| Date | Actor | Entry |
|---|---|---|
| 2026-08-24 | story finisher | **All 8 QA findings resolved: 8 FIX, 0 DISMISS, 0 DEFER.** **Major 1 — the falsified `PinAxisLocation` claim was SWEPT, not sampled**: five search forms across **every file type**, **8 carrier sites** found (3 corrected in development, 2 named by the review, **1 found only by widening past `.go`/`.md` — `tools/fontgen/instance_faces.py:16-19`, a Python build tool**, and 2 in the closed Story 2.2 record). Six corrected in place; the two in the closed story corrected by an **appended** section under D-1.6.6, never by editing the original lines; decision-log instances deliberately untouched (append-only by its own header). Three further term-matches verified **not** carriers and left alone, including Story 1.5's, which is accurate as written. Corrected text states the real grounds — payload **4.82 vs 8.30 MB** compressed, `usWeightClass` correctness, deleting the FMA path — not merely deleting the claim. **Major 2 — the test-scope inventory is now site-exact**: re-keyed from `(path, rule)` to `(path, rule, line)` via a new `assertExactFindingSites`; the comment and AC2 corrected to match the guard (D-000.24/D-000.28 false-credit class). **The gap was live, not hypothetical** — on measuring, the tree held **four** sanctioned sites while the `want` list named three, because `vendorboundary_test.go` already carried two. **Red-proved both directions**: adding the review's exact second site → RED (`line=661`, `got 5, want 4`); removing one of two → RED (`line=206`, `got 3, want 4`). The first removal mutation broke the build and reddened through the loud-failure path instead — discarded and replaced with a compiling one, recorded because a red result through the wrong mechanism proves nothing. **Minor 3 — absolutes recomputed from a cited invocation** (`CGO_ENABLED=0 GOWORK=off go test -count=1 -v ./...`, counted with `awk '/--- PASS/{p++} /--- FAIL/{f++}'`, baseline in an isolated worktree at `431a6a5`): recorded `375 → 397` is wrong on **both** sides; measured **377 → 399**, delta `+22` exact. **Root cause identified, not guessed: the two inherited failures were being subtracted from the PASS count**, which is why `lint` (no failures) reproduced exactly at `75 → 80`. Also corrected the story's own opener, which had "corrected" the commissioning brief's *accurate* "377 pass / 2 fail" into "375 + 2 = 377" while citing D-000.18. **Minor 4 — the `Errors`-sweep red-proof is COMMITTED, not labelled**: the proof was constructible, so a D-000.24 forward-guard label would have been a dodge. New retained fixture `testdata/lint/float-typed-untypecheckable/` (parses, does not type-check) + `TestFloatTypedScanFailsLoudlyOnATreeThatDoesNotTypeCheck`, asserting three preconditions incl. that it **parses**. Red-proved by neutering only the sweep: new test reports **"no error and 0 findings"** while the pre-existing `OnAnUnloadableTree` test **stays PASS** — proving the two exercise different branches. **Nits 5-8**: the `cmap` carve-out's chain-level justification is now an assertion — two-face chain, `cmap`-stripped face first, **60,414 bytes `err=nil`**, independently reproducing the review's figure, with an `OS/2` control at **0 bytes**; intact anchor `wantAdvance36 = 1839` added with an explicit `!= Upem()` (2048) assertion; `TestFamilyNameHasNoCallSite` rebuilt on the **AST**, red-proved on three legs incl. the file the old filename exclusion **excused**; AC8.2's stale text disclosed, noting D-2.3a.1 moved **two** ACs and only AC8.3's was declared. **Consequence of Major 1 swept and repaired**: correcting site 5 moved the disproving call 515→529, staling **five** citations across four files; re-cited by **test name**, not renumbered, since position-bound citations restale. **Gates, each with its invocation**: `folio-go` **400 PASS · 2 FAIL** (every `--- PASS` occurrence; 254 top-level), `lint` **81 PASS · 0 FAIL** (43 top-level) — the two failures are Story 2.1's intentional `TestCorpusMeetsP6ExerciseFloors` and `TestP2IndependentDPCrossCheck`, unchanged, **no third failure**. `gofmt -l` empty, `go vet ./...` and `go build ./...` clean in **both** modules; `go build/vet -tags=matrix` clean. **`go list -m all` returns exactly two lines**, unchanged. **NO GOLDEN MOVED**: `git status --porcelain fixtures/` empty, all five digests unchanged, and **`fixtures/shaped-text` is still `5964aad0e696010c6e3f34a48d0775af6ae527a6cbe2f5c6319158f43c92e00f`** — the pending Thai sign-off is **not** invalidated. No `SOURCE_DATE_EPOCH` literal in any `.go` under `folio-go/`. **The Epic 2 gate still owes exactly two things and no more**: the four-target matrix legs (written, registered, **compiled and vetted but deliberately NOT RUN — per-epic cadence, D-000.4**) and the Thai sign-off (**deliberately red**; `fixtures/shaped-text/thai-signoff.json` confirmed absent). **No third obligation was added.** Endorsed and not reopened, per the review: the `cmap` carve-out, all three developer disclosures (incl. the `hhea`/`hmtx` forward-guard label, which D-000.30 now rules exactly), and the no-third-gate conclusion. Status → `done`. |
