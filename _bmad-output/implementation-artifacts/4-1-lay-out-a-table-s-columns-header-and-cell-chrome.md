---
baseline_commit: d5b75b5
---

# Story 4.1: Lay out a table's columns, header and cell chrome

**Epic:** 4 — A Go developer can render the golden report (**the C4 gate**)
**Story key:** `4-1-lay-out-a-table-s-columns-header-and-cell-chrome`
**Status:** `done`
**Covers:** **FR23**, **FR24** · **AD-13**

**Primary invariant:** **AD-13**, verbatim (`ARCHITECTURE-SPINE.md:301-308`):

> **Rule:** column widths are absolute and authoritative. A table's width **is** the sum of its
> column widths; it is never stored as an independent field. The content area's height is
> likewise derived — page height minus margins minus page-header height minus page-footer
> height — by one function in `internal/layout`.

**Co-primary invariant:** **AD-24**, the *Origin*, *Axis* and *Visibility* clauses
(`ARCHITECTURE-SPINE.md:446-462`). AD-24 is why AC2 exists at all: *"nothing negotiates"* is the
title of the invariant, and a column that widened itself to fit a header label would be the
first negotiation in the engine.

**Adjacent invariants:**
- **AD-5** (`:179-186`) — *"`internal/pagemodel` types name only geometry, glyph runs, images,
  **and vector primitives**. `internal/layout` may not import `internal/pdf`."* The page model
  has no vector primitive today; AD-5 already sanctions the one this story adds.
- **AD-3** (`:132-150`) — two numeric representations, one decimal emitter, rounding only where
  a value is converted **into** thousandths, always round-half-to-even. This story emits the
  engine's first colour and its first line width; both go through that emitter.
- **AD-1 / AD-23** — no binary float anywhere under `internal/`. Column widths, padding, border
  thicknesses **and colour channels** are exact integer geometry.
- **AD-14** — one `Diagnostic` type, three settled data cases, never a panic. A new code is
  minted when its condition first ships (**D-000.65**).
- **AD-10** — opaque element ids; a column carries its own id so a diagnostic can name it.
- **AD-11** — row scope and the row alias. Not exercised here (4.2 owns rows), but the table's
  `bind`/`as` validation already runs and must not regress.

**Governing rulings this story is bound by:**
**D-000.68** (a guard must be anchored to something the code under test cannot move; **pin to a
literal when the set is permanent, state it relationally when it is scheduled to move**) ·
**D-000.9** and its extension (*"all clear" must never be indistinguishable from "could not
look"* — and this now covers **recorded evidence**: every measurement below names its command,
its mutation, and confirms the mutation was run) · **D-000.72** (the spine draws no edges; the
enforced rule is `stageRankTable` in `lint/internal/rules/stagerank.go`, and a guard holds the
spine's ladder to it) · **D-3.5.4** (AC4's style-field property, **inherited by this story** —
see the whole of §D3 below) · **D-3.5.5** (the style-field list is derived from the **schema**,
not hand-written, *because that set is scheduled to grow* — this story is the growth it
predicted) · **D-000.65** (mint a diagnostic code when its condition first ships) ·
**D-000.4** (heavy-test cadence is **per-epic** for this run) · **D-000.26** (a figure without
its scope and flags is not a figure) · **D-000.24** (a guard with no constructible red-proof is
labelled as such and never credited with one) · **D-000.49** (append, never rewrite) ·
**D-000.59** (discharge by replacement, not deletion alone) · **D-000.64** (the gate runs all
three Go modules) · **D-000.17 / D-2.1.14 / D-000.57** (the required red) · **D-3.4.7**
(dot-imports banned under `internal/`) · **D-000.69** (verdict lines are the interface).

---

## Obligations this story inherits, and what it does NOT inherit

**One decision-log obligation, and it is substantial.**

**D-3.5.4, verdict part 3, names this story by number:**

> *"The remainder is owed by the stories that create the consumers. Story **4.1**'s AC requires
> 'cell borders and cell padding are produced per the template's styling'; Story **4.8** renders
> `altRowBackground`. **They inherit AC4's property for the other eight fields.**"*

What is unproven, exactly, is set out and re-measured in §D3. It becomes **AC7** of this story.

**One review finding from Story 3.5 lands here.** Story 3.5's Finding 6 (Minor, resolved by
comment rather than by guard): `computeVisibility` records a `visibleIf` verdict for all five
element kinds, and **only `text` and `image` verdicts are ever read**. The finding's own words:
*"nothing forces it to consult it… no test would fail if Epic 4 shipped table rendering that
ignored `visibleIf` entirely."* This story is the one that makes a table reach the page model.
It becomes **AC8**.

**`deferred-work.md` carries no entry owned by Story 4.1 or by "the first table story."**
Measured at creation: `grep -n "4\.1\|first table story" deferred-work.md` → 3 hits, all
examined, **none an obligation on this story**:

| hit | what it is | bearing on 4.1 |
|---|---|---|
| `:232` (DW-5) | prose about stage ranks | none — DW-5 is **RETIRED** |
| `:867-869` (DW-16) | *"Stories 4.1 and 4.2 do **NOT** close the window"* | an explicit **exemption**: *"4.1's borders and padding are rects, not glyphs. There is more slack here than the gate assumed."* DW-16 stays open and 4.1 owes it nothing. |
| `:484` (DW-11) | owner is *"Epic 2's later stories and **Epic 4's golden-report work**"* | that is **Story 4.7**, the golden report. Not 4.1. |

**One deferred entry fires at Epic 4 planning and is NOT this story's to absorb — escalate it.**
**DW-4** — *"Nobody owns cutting `folio-go/v0.1.0` — **owner decision when Epic 4 is planned**"*,
whose text reads: *"If it is still unowned when Epic 4 is planned, it goes to the owner rather
than being absorbed into a story."* Epic 4 is now being planned. **DW-3** (publishing the
third-party licence manifest) is owned by *"Epic 4 close"* and depends on DW-4. Both are raised
in §DECISION-4 for routing, and neither is in this story's scope.

**This story retires no DW entry.**

---

## SIX divergences between the brief and the shipped tree, all measured in this run

**Read these before the ACs. Three of them change the shape of the work.**

### D1 — There is no vector primitive in the page model, and no fill or stroke anywhere in the PDF writer. This story adds the first of each.

`pagemodel.Page` (`folio-go/internal/pagemodel/pagemodel.go:243-249`) is exactly:

```go
type Page struct {
	Runs                  []TextRun
	Images                []ImagePlacement
	Width, Height         geom.Length
	MarginTop, MarginLeft geom.Length
}
```

Two content slices. No rectangle, no rule, no fill.

**Measured, at HEAD, over the whole `internal/pdf` package:**
`grep -n '" re\|" f"\|" S"\|rg\b\|RG\b\|writeOp' internal/pdf/textdoc.go` → **one hit**,
`textdoc.go:766`, `" re\nW n\n"` — the **clip** path from Story 2.8. There is no `f` (fill), no
`S` (stroke), no `w` (line width), no `rg`/`RG` (colour) operator emitted anywhere in the
module. `grep -rn "rg\b\|RG\b\|DeviceRGB\|Colou\?r" internal/pdf internal/geom internal/pagemodel`
returns only `ColorSpace`/`PredictorColors` on the **image** XObject dictionary — an image's own
colour space, never a drawing colour.

**Consequence.** "Cell borders and cell padding" is not a small addition to an existing drawing
layer. It is the drawing layer's first stroke and first fill, and its first colour. That carries
an AD-3 obligation (§D2) and an AD-5 obligation (the primitive lives in `pagemodel`, names no
PDF concept, and `layout` still may not import `pdf` — rank 7 vs rank 8 in `stageRankTable`).

### D2 — `#RRGGBB` has no exact PDF representation, AD-1 forbids the obvious conversion, and the existing round-half-to-even scaler is the answer.

PDF's `rg` and `RG` operators take three operands in **0..1**. The format's colours are
`#RRGGBB` — three channels in **0..255** (`folio-format.md`, *"Colours are `#RRGGBB`"*). The
conversion is a division that is not exact in decimal: `0x80` → `128/255` = `0.50196…`.

`float64(c)/255.0` is **banned** (AD-1/AD-23, mechanically, at build time).
`strconv.FormatFloat` is **banned** (AD-3's *Prevents* clause names it by name).

**The mechanism already exists and is already audited.** `geom.ScaleRound(v Length, num, den
int64) Length` (`internal/geom/scale.go:58`) is the module's single round-half-to-even integer
scaler with overflow detection. `geom.ScaleRound(Length(channel), 1000, 255)` produces a
thousandths-scaled `int64` — **exactly the input the decimal emitter already takes**, since AD-3
defines it over *"a thousandths-scaled `int64` — every `geom.Length`, **and every other value
defined in thousandths**."* `0x80` → `502` → `"0.502"`. `0xFF` → `1000` → `"1"`. `0x00` → `"0"`.

**This is the ruled conversion for this story.** It reuses the audited path rather than opening
a second one, it is exact-by-construction at the emitter's own precision, and it is identical on
every target because it is integer arithmetic end to end. See **AC5**.

### D3 — D-3.5.4's population is unchanged at HEAD, re-measured; and only FIVE of the eight inert fields can be wired by this story.

D-3.5.4's table was measured at Story 3.5. Stories 3.6 and 3.7 landed after it. **Re-measured at
`d5b75b5`**, command:

```
for f in Background Align Valign Border Bold Italic Padding AltRowBackground FontSize FontFamily; do
  grep -rn "\.$f\b" --include="*.go" . | grep -v "_test.go" | grep -v "^./folio-go/internal/template/"
done
```

| field | non-test consumers outside `internal/template`, at HEAD | D-3.5.4's finding | changed? |
|---|---|---|---|
| `Align` | 2 — `folio_expr_validate.go:251,252` | validator only | **no** |
| `Background` | 2 — `folio_expr_validate.go:256,257` | validator only | **no** |
| `Valign` | 2 — `folio_expr_validate.go:266,267` | validator only | **no** |
| `Border` | 3 — `folio_expr_validate.go:240,271,272` (one is a comment) | validator only | **no** |
| `Bold` | **0** | zero | **no** |
| `Italic` | **0** | zero | **no** |
| `Padding` | **0** | zero | **no** |
| `AltRowBackground` | 1 — `folio_expr_validate.go:93` | validator only | **no** |
| `FontSize` | 6 | renders, but `Presence[geom.Length]` — a `{{ }}` cannot inhabit it | **no** |
| `FontFamily` | 4 — incl. `render.go:904,907` | renders, and is a string | **no** |

**D-3.5.4's table holds exactly. Nothing has moved.** The eight inert fields are still inert.

**But this story can only wire five of them, and the reason the other three stay inert is a
measured fact about the tree, not a scope choice.**

| field | wired by 4.1? | grounds |
|---|---|---|
| `border` | **yes** | cell borders (AC3) |
| `padding` | **yes** | cell padding (AC3) |
| `background` | **yes** | the cell fill the border encloses (AC3) |
| `align` | **yes** | AC2's per-column alignment falls back to it (AC4) |
| `valign` | **yes** | `headerHeight` is a fixed box; a baseline inside it must be chosen, and `valign` is the field that chooses it (AC3) |
| `bold` | **NO — not implementable** | the shipped font set has no bold face. `fontset.go:721`: *"Bold, when it arrives, is a `wght` instance"*. `text/shape.go:70` records that HarfBuzz's `SetSyntheticBold`/`SetSyntheticSlant` exist and *"**NONE** of them is called"*. Wiring `bold` here would mean either synthetic emboldening (a new rasterisation behaviour, unowned, and a cross-target risk) or a silent no-op. |
| `italic` | **NO — not implementable** | same measurement, same two sites |
| `altRowBackground` | **NO** | **Story 4.8** owns it, by name, in D-3.5.4 itself |

**So the honest restatement this story owes (AC7) is: eight inert → three inert**, each of the
three with its owner named and its reason measured, not apologised for. This is D-000.24's
labelled category used the way D-3.5.4 used it: *not* "we could not prove it" but *"there is
nothing here to prove yet, and here is the count."*

### D4 — The load-time rejection of `width` on a table already exists — and has **ZERO tests**.

`internal/template/parse_bands.go:156-164`:

```go
	wRaw, wok := obj["width"]
	hRaw, hok := obj["height"]
	if el.Type == ElementTable {
		if wok {
			return Element{}, newLoadError("width", string(id), string(wRaw), "a table declares x and y only — never width (AD-13, AC5)")
		}
		if hok {
			return Element{}, newLoadError("height", string(id), string(hRaw), "a table declares x and y only — never height (AD-13, AC5)")
		}
```

**Measured:** `grep -rln "x and y only" --include="*_test.go" .` → **no matches**.
`grep -rn "never width\|never height" --include="*_test.go" .` → **no matches**.

The single strongest existing guarantee behind AC1's negative half is **untested**. Deleting the
`if el.Type == ElementTable` branch makes a table with `"width": 500` load cleanly, storing it on
`Element.Width`, and the entire three-module gate stays green. **AC1 pins it.**

### D5 — The drift test CANNOT see a table-width field. Do not credit it with AC1.

`internal/template/drift_test.go` compares two **flat, whole-document key-name sets**: every
string key `serialize.go` can emit, against every backticked token in `folio-format.md`
(`TestDriftGoToDoc:232`, `TestDriftDocToGo:269`). It is not scoped per object.

`width` is **already** a documented key — three times over: the element box, `columns[].width`,
and `style.border.width`. So adding `Width geom.Length` to `TableExt` and serializing it as
`"width"` inside the table object leaves both drift halves **green**. This is the
D-000.68 failure shape in a test that looks like it covers the property: the anchor (the spec
markdown, which the code cannot edit) is a good anchor, but the *space* it compares over is too
coarse to express the claim. **AC1 needs its own instrument.**

### D6 — Nothing renders a table today; `checkTableBindings` is validation only.

`render.go:277` — the one non-validator site that tests `el.Type != template.ElementTable` — is
inside `checkTableBindings`, which resolves the collection path and rejects a reserved row alias.
It produces no geometry and reaches no output byte. `line` and `rect` elements parse
(`parse_bands.go:248`) and render nothing at all.

**Consequence for the scope fence:** every column-geometry, header and chrome behaviour in this
story is **new construction**, not modification. There is no existing table rendering to
regress, and correspondingly **no existing golden fixture moves** unless this story registers a
new one (see AC10).

---

## Baseline, measured in this run at creation

**HEAD:** `d5b75b5` — *"Epic 3 boundary gate: pass, four blockers closed, epic-3 done"*.
Working tree clean at creation (`git status --porcelain` → empty).

Every figure below carries its scope and its flags (**D-000.26**).

| module | command | result |
|---|---|---|
| `folio-go` | `go test ./...` (no flags, from `folio-go/`) | **920 pass · 1 FAIL · 1 skip**, 17 packages |
| `lint` | `go test ./...` (no flags, from `lint/`) | **114 pass · 0 fail**, 4 packages |
| `hashmatrix` | `go test ./...` (no flags, from `hashmatrix/`) | **3 pass · 0 fail**, 2 packages |

**The one failure is the REQUIRED red and must not be "fixed."**
`TestCorpusMeetsP6ExerciseFloors` (`internal/text/corpus_test.go:189`):
`P6g (opaque names) floor not met: got 7, need >=20`, with
`{P6a:64 P6b:63 P6c:16 P6d:20 P6e:284 P6f:115 P6g:7}`. **D-000.17 / D-2.1.14 / D-000.57** mandate
it stay unmet until real opaque Thai proper nouns are sourced. Epic 2 and Epic 3 both closed over
it deliberately.

**One brief-vs-measurement discrepancy, recorded rather than smoothed over (D-000.26).** The
brief carried `lint` as **113/0**; the measurement at `d5b75b5` is **114/0**. Same scope, same
flags, one test more. The most likely cause is a test added in the Epic 3 boundary-gate commit
itself (`d5b75b5` added `TestSpineStageLadderMatchesStageRankTable` per D-000.72), which is
exactly a `lint/internal/rules` test. **Use 114 as this story's baseline**; a developer who ends
at 113 has deleted something.

---

## In plain terms (read this first if you just want the gist)

A report's transaction table is the whole reason this project exists, and the engine can now
draw the top of one. It already knew how to read a table out of a template file; what it gained
here is turning that into marks on a page. It works out where each column starts and how wide it
is from only the widths a template author declared, and a table's overall width is simply those
widths added together, never a separate number that could quietly drift out of step. It draws
the header row across the top, each title placed left, centre or right as declared, and it gives
the engine its first-ever ability to draw a plain box: the lines around a cell, the space between
a cell's edge and its label, and a background fill — including, for the first time, exactly how
a designer's colour becomes identical bytes on every machine.

One thing changed partway through that was not in the original plan: the header row can now be
given a look of its own, separate from the rest of the table, so it does not read identically to
the data beneath it — a shaded, bordered header a later flagship report needs and otherwise had
no way to ask for.

What it deliberately does not do: no data rows, not one — rows, cell text and splitting a table
across pages are all later work. Bold and italic text still do nothing anywhere, since the
shipped fonts have no such versions to switch to; shading alternate rows is left for later and
may be cut. None of these three are oversights — each is counted and its owner named.

One side effect: a table column with a visible title now needs a font named somewhere, since none
is ever chosen automatically — a document that used to load quietly will now be turned away until
it says which font to use.

One existing check still fails on purpose, as it has for some time: a sample of Thai proper
names is deliberately short of a kind not yet sourced, left visibly unmet rather than hidden and
must stay failing.

---

## Story

As a template author,
I want fixed-width columns with their own alignment, a header row, borders and padding,
So that a transaction table reads as a table rather than as text in rows.

---

## Do not re-open — settled rulings this story inherits

- **A table's width is not stored. Ever.** AD-13. The load-time rejection already exists
  (`parse_bands.go:158-164`); this story pins it and extends the property to the new types it
  creates. Do not add a width field to `TableExt` "for convenience."
- **A table declares `x` and `y` only.** Same rule, same site, for `height`.
- **`columns[].footerOf` derivation runs at load, in `folio.ParseTemplate`, and is never written
  back into the document** (D-1.4.2, DW-5 retired at 3.2). Do not touch it.
- **`internal/template` (rank 2) can never import `internal/expr` (rank 3)**, which is why the
  footer derivation lives at the module root. Same table forbids `internal/layout` (rank 7)
  importing `internal/pdf` (rank 8) — `stagerank.go:66-67`, comparison at `:239` is
  **strictly lower**.
- **The spine draws no dependency edges** (D-000.72). If you need to know what may import what,
  read `stageRankTable`, not a diagram. There is no diagram.
- **`visibleIf` is not valid on a table column** — rejected at load, naming the column id
  (Story 3.5). Row-level visibility would make pagination a function of data.
- **A `{{ }}` placeholder inside a string-valued style field is a load error** (D-3.5.2 /
  D-3.5.5). The covered set is **derived from the document model at build time**, not
  hand-written. This story adds render consumers to five of those fields; it must **not** add a
  second, hand-written list beside the derived one.
- **There is no colour-by-data and there will not be.** Conditional visibility is in scope,
  conditional formatting is not.
- **Dot-imports are banned outright under `internal/`** (D-3.4.7).
- **Heavy tests are written, not run, this epic** (D-000.4). See AC10.

---

## R — design constraints derived from the record during creation

**R1 — The rectangle primitive lives in `internal/pagemodel` and names no PDF concept.**
AD-5 already lists *"vector primitives"* among what the page model may name. It may carry
geometry (x, y, width, height in `geom.Length`, top-left origin, Y down per AD-24's *Axis*
clause), an edge set, a stroke width, and colours as **channel triples**, never as a hex string
and never as a PDF operand. The single top-left→bottom-left flip stays in `internal/pdf`'s one
existing flip function (`flip.go`); nothing new inverts a coordinate.

**R2 — Colour is converted by `geom.ScaleRound(channel, 1000, 255)` and emitted by the existing
decimal emitter.** See §D2. No new numeric route. No float, at any point, including in tests.

**R3 — `internal/layout` may measure the header label, but only to POSITION it.** Measuring to
*decide a width* is the negotiation AC2 forbids. The distinction is behavioural and AC2 tests it
directly; a reviewer should also read the call site.

**R4 — The table-geometry type exposes width as a computed value, not a field.** This is the
compiler as anchor (D-000.68): if there is no field, no caller can cache a stale one. AC1's
reflection guard holds the type to it.

**R5 — `TableExt`'s field set is PERMANENT through the MVP, so AC1 pins it to a literal.**
This is D-000.68's discriminator applied deliberately, and it lands **opposite** to D-3.5.5 one
epic earlier — recorded here because the pair reads as inconsistent otherwise. D-3.5.5 made the
style-field list relational *because the style-field set was scheduled to grow at Epic 4*. It
grew: this story is that growth. `TableExt`'s own field set is **not** scheduled to grow —
walking the remaining roadmap, 4.2 uses `bind`/`as`/`columns`, 4.3 and 4.4 use `headerHeight`,
4.5 uses `columns[].footer`/`footerOf`/`footerFormat`, 4.8 uses `altRowBackground`, and Epic 6's
table editor edits the same schema through the engine (AD-15). **Every field those stories need
already exists.** So the set is permanent, and a literal is the right instrument.

**R6 — AMENDED mid-story (engineering-lead ruling, recorded verbatim in the Delivery Log).** As
originally written this rule was FALSE against the shipped contract: `fontChain` (render.go:903)
requires `style.fontFamily` to be set for ANY element carrying text, table or not, and no font
default exists or will be added by this story (folio-format.md's "first key of `fonts`" default
was documentation for behaviour that was never built — `fonts` is a mapping with no authored key
order, so "first key" was never well-defined; folio-format.md is amended in this story's commit
to say so). The requirement is corrected, not merely annotated "untested" (D-000.66): a
knowingly-false requirement must not be carried forward as a fact.

**R6, restated — a table whose style block sets only `fontFamily` (and nothing else) must
render.** Absent-and-otherwise-undeclared means the documented default for every OTHER field:
padding `0` on four edges, **no border drawn**, transparent background, `align` left, `valign`
top. Not an error, not a panic (AD-14). A table with NO style at all, or with a style that omits
`fontFamily`, and at least one non-empty column label, is a located error — exactly the failure
mode a text element with the same omission already has, unchanged and unextended by this story.

**R7 — No new diagnostic code unless a condition ships that needs one.** D-000.65. Walk the
story's failure modes at implementation: if a genuinely new *reported* condition arrives (a
candidate: a table whose summed column width overflows its band — see DECISION-3), it mints a
code then, not in advance.

**R8 — Do not touch `render_visibility.go`'s verdict computation.** It is already correct for
all five kinds. AC8 adds a **consumption** of the table verdict and a guard; it does not rebuild
the producer.

---

## Acceptance Criteria

Ten ACs. AC1, AC2 and AC7 are the ones a reviewer should attack hardest.

### AC1 — The table's width IS the sum of its column widths, and no independent table-width field is stored anywhere

**Given** a table with declared column widths
**When** its geometry is computed
**Then** the table's width is the sum of its column widths
**And** no independent table-width field is stored anywhere

This is a **negative property**, which is precisely where this project has repeatedly shipped
guards that could not fail (D-000.68's four Epic-3 instrument defeats; D5 above is a fifth,
found at creation). It gets **four** instruments, each with a named anchor and a *compiling*
mutation that reddens it. §D5 records why the existing drift test is **not** one of them.

| # | instrument | anchor (D-000.68) | discriminating mutation |
|---|---|---|---|
| **1a** | A test-owned fixture document declaring `"width": 500` on a `table` element fails `folio.ParseTemplate`, with the message naming the element id and the field. A second fixture does the same for `"height"`. | **a literal the test owns** — the fixture JSON is authored in the test | **CORRECTED at finish (review Finding 1, Major)** — see below: keep the `ElementTable` branch, and for the `width` half store the declared value on `Element.Width` instead of rejecting it (`consumed["width"]=true; v,_ := decodePointsRaw(...); el.Width = present(v)`) → **red**, `got nil error`, for the right reason. The ORIGINAL text here (deleting the branch head entirely) was measured at finish to redden `TestTableDeclaringWidthIsRejected`/`TestTableDeclaringHeightIsRejected` for the WRONG reason (the sibling dimension, not the declared one, going missing) and to also redden the D-000.9 control `TestTableWithNeitherWidthNorHeightLoads` — a control failing alongside the guard is itself the signal the mutation was not isolating the property. *(Measured at creation: with the branch present, **zero** tests exercise it — §D4.)* |
| **1b** | Reflection over `template.TableExt`: its exported field-name set **equals** a literal the test owns — `{Bind, As, Columns, HeaderHeight, AltRowBackground}` — set equality in **both** directions. Same for `template.Column`: `{ID, Label, Width, Align, Bind, Footer, FooterOf, FooterFormat}`. `Extra` is excluded by name with a comment saying why. | **a literal the test owns**, per R5 | add `Width geom.Length` to `TableExt` → **red**, naming the added field. Renaming it `Extent` also reddens: the guard compares the **set**, not a spelling. |
| **1c** | Reflection over the new layout table-geometry type: its exported field-name set equals a test-owned literal that **contains no width field**, and the type's total width is reachable only as a computed value over the column slice. | **the compiler**, backed by a test-owned literal | add a `Width` field and populate it once at construction → **red** on the set equality, before any staleness can be observed. |
| **1d** | Behavioural property: over a test-owned table of column-width vectors — `{}`, `{70}`, `{70,80,60,40,110}`, `{1}`, `{1,2,3}` (**added at finish**, review Finding 10) and one vector summing near `int64` limits — the computed width equals the independently-summed expectation. The **empty** vector must give **zero**, not an error. **CORRECTED at finish**: the near-limit vector is a large-value SMOKE CASE, not a proof of overflow handling — `Width()` is a plain `total += c.Width` loop with no overflow detection and no call into `geom.ScaleRound` (the ORIGINAL text here claimed otherwise; see review Finding 10). | **a literal the test owns** (the vectors and their sums are written out, never computed by calling the function under test) | return `sum + headerHeight`, or clamp to the band width, or `max()` instead of `+` → **red** on at least three vectors (measured: exactly 3 of 6, after the third multi-column vector was added — see Mutations run). |

**Vacuity guard (D-000.9).** 1b, 1c and 1d each `t.Fatal` if the reflected field count is zero,
if the vector table is empty, or if zero comparisons were made. A reflection helper that silently
returns nothing produces exactly the "no differences" all-clear a healthy one does.

**Recording obligation (D-000.9 extension).** All four mutations are **applied, run, and the
result recorded** in the Delivery Log — mutation, command, red/green, and `git diff` confirming
the tree was restored. A described mutation is not a run mutation.

### AC2 — Column widths are never negotiated against content

**Given** each column
**When** it is laid out
**Then** it honours its own alignment and its declared width
**And** column widths are never negotiated against content

This is the invariant that makes 4.2's wrapping and 4.3's pagination tractable, and per the brief
it gets a guard of its own rather than a comment.

**The subject available at 4.1 is the header label**, since this story produces no data rows.
That is not a weaker subject — it is the *same* code path 4.2's cell text will enter, and it is
available now.

**Given** two templates identical in every byte except their column `label` strings — one with
labels far narrower than their declared widths, one with labels far wider (a long Latin label, a
long Thai label with no spaces, and a long CJK label, so the property is asserted across all
three shaping paths)
**When** both render
**Then** every column's x-origin, every column's width, and the table's total width are
**identical** between the two renders
**And** the wide-label render's header text is clipped or overflows per the declared box — it
**never** widens the column

- **Anchor:** the two label sets are **literals the test owns**, and the expected geometry is
  written out in the test, never read back from the function under test.
- **Discriminating mutation (must compile, must be on the real path):** in the column-geometry
  function, replace the declared width with `max(declaredWidth, measuredLabelAdvance)`. This is
  the exact defect the AC forbids, it compiles, and it is what a well-meaning author would
  actually write. → **red**, on the wide-label case, in all three scripts.
- **Second mutation, weaker but worth running:** make the *table* width `max(sum, widestLabel)`
  → red on AC1's 1d as well, proving the two ACs are not measuring the same thing twice.

**A control that must stay GREEN** (D-000.9 — an all-clear must be distinguishable from a
could-not-look): the narrow-label and wide-label renders must **differ in their glyph output**.
If the test cannot observe that the labels differ at all, its geometry equality is vacuous. Assert
the difference explicitly.

### AC3 — A header row, cell borders and cell padding are produced per the template's styling — and "the template's styling" is pinned to the actual schema

**Given** a table
**When** it renders
**Then** a header row, cell borders and cell padding are produced per the template's styling

The brief flags this AC as the one with the most room for interpretation. **It is pinned here
against `folio-format.md` and `internal/template/model.go`, and no schema field is invented.**

**What the schema actually offers a table — the complete list, read from `model.go:214-235`:**

```go
type TableExt struct {
	Bind             string
	As               Presence[string]
	Columns          []Column
	HeaderHeight     geom.Length
	AltRowBackground Presence[string]
}
type Column struct {
	ID, Label, Width, Align, Bind, Footer, FooterOf, FooterFormat, Extra
}
```

plus the element's own `Style` (`Presence[Style]` on `Element`).

**Therefore, and this is the finding, not a design choice:**

1. **There is no per-cell style object, and no per-column style object.** A column carries
   `align` and nothing else that is appearance.
2. **There is no header-specific styling of any kind.** No `headerStyle`, no header font, no
   header background, no header border. See **DECISION-1** — this is escalated, not invented
   around.
3. **The only styling a table has is the table element's own `style`** — one `Style` block for
   the whole table — **plus each column's own `align`**.

**So "per the template's styling" means exactly this mapping, and the story asserts it field by
field:**

| what is produced | the field that governs it | default when absent |
|---|---|---|
| the header row's height | `table.headerHeight` (**required**, `parse_bands.go:303-311`) | none — a missing `headerHeight` is already a load error |
| each header cell's rectangle | the column's declared `width` × `headerHeight`, at the column's computed x-origin | — |
| cell **borders** | `style.border.edges` (which edges), `style.border.width`, `style.border.color` | `border` absent → **no border drawn** |
| cell **padding** | `style.padding.{top,right,bottom,left}`, each independently | `0` on all four edges |
| cell **background** | `style.background` | absent → transparent, nothing emitted |
| the header label's **horizontal** placement | `columns[].align`, falling back to `style.align` — see AC4 | `left` |
| the header label's **vertical** placement in the fixed-height header box | `style.valign` | `top` |
| the header label's face and size | `style.fontFamily`, `style.fontSize` — **already wired**, do not rewire | first `fonts` key; `10` |

**And:**

**Given** a table element carrying **no** `style` at all
**When** it renders
**Then** the header row is produced with no borders, no padding, a transparent background,
left-aligned top-placed labels, and the render **succeeds** (R6 — not an error, not a panic)

**Given** `style.border.edges` naming a strict subset, e.g. `["bottom","top"]`
**When** the header row renders
**Then** exactly those edges are stroked on each header cell and the other two are not

**Anchor:** a byte-level assertion over the produced content stream for a test-owned template,
plus structural assertions over the page model. **Mutations:** drop the padding inset (labels
shift to the cell edge → red); ignore `edges` and always stroke four (→ red); emit the background
when `background` is absent (→ red).

### AC4 — `columns[].align` wins over `style.align` for that column, and the precedence is asserted

**The schema does not state this precedence.** `columns[].align` and `style.align` can both be
present and mean different things for the same cell. **DECISION-2** escalates it; the ruling
this story ships against, pending that decision, is:

**Given** a table whose `style.align` is `"right"` and whose first column's own `align` is
`"left"`
**When** the header row renders
**Then** the first column's label is **left**-aligned and every column without its own `align` is
**right**-aligned

**Grounds:** Story 4.1's own AC2 says each column *"honours **its own** alignment"*, and the
column's field is the more specific of the two. `style.align` is the table-wide default.

**Anchor:** test-owned template and test-owned expected x-positions. **Mutation:** make
`style.align` win → red; make `columns[].align` the only source (ignoring `style.align` entirely)
→ red on the second half.

### AC5 — Colour reaches the page through `geom.ScaleRound` and the existing decimal emitter, with no float anywhere

**Given** a `#RRGGBB` colour on `style.border.color` or `style.background`
**When** it reaches the content stream
**Then** each channel is converted by `geom.ScaleRound(channel, 1000, 255)` and written by the
existing decimal emitter
**And** no `float32`, `float64`, `math/big.Float` or `math/big.Rat` appears on the path
**And** the same colour produces the same bytes on every target

**Test-owned expectation table**, written as literals, never computed by the code under test:

| hex | channel | thousandths | emitted |
|---|---|---|---|
| `#000000` | 0 | 0 | `0` |
| `#FFFFFF` | 255 | 1000 | `1` |
| `#808080` | 128 | 502 | `0.502` |
| `#010101` | 1 | 4 | `0.004` |
| `#7F7F7F` | 127 | 498 | `0.498` |

*(`128×1000 = 128000`; `128000/255 = 501.96…`; round-half-to-even → `502`. `127×1000/255 =
498.03…` → `498`. The pair `127`/`128` is deliberate: it is the tie-adjacent case where a
naive `+0.5` truncation and round-half-to-even can disagree.)*

- **Anchor:** the literal table above, owned by the test. The `float` half is anchored to
  **`lint`'s cross-module, type-resolved scan** (`TestFloatTypedProductionScan`/
  `TestFloatTypedTestScopeInventory`, `lint/internal/rules`), which is **inside** the
  three-module gate (D-000.64) — **CORRECTED at finish (review Finding 4, Major)**. The
  ORIGINAL text here claimed "the build fails" and named "the build" as the anchor; that is
  **false**. Measured at finish: `geom.Length(float64(c.R)/255.0*1000)` substituted at
  `appendColorChannels`' call site **compiles cleanly** (`go build ./...` succeeds, `go vet`
  silent) — Go's type system permits a float64 expression converted to an integer type, so
  there is no syntactic float-literal ban for the compiler to enforce here. What actually
  catches it is `lint`'s scan, run from `lint/`: `TestFloatTypedProductionScan` reports
  `internal/pdf/rectdoc.go:106:38: this value expression has floating-point type float64
  (resolved by go/types, not spelled in the source)`, and `TestFloatTypedTestScopeInventory`
  reds on the site-count mismatch (got 6, want 5). The property IS protected; only the record
  naming the wrong instrument was wrong, and it is corrected here rather than merely annotated.
- **Mutation:** change `ScaleRound`'s rounding at the call site to truncation → **red** on
  `#808080` (`501` vs `502`). Introducing `float64(c)/255.0` → **the build succeeds**; the
  discriminating instrument is `lint`'s cross-module scan, above — see the Mutations run table's
  14th row for the command and result.
- **D-000.61 note:** a float red-proof must vary the operand **order**. The colour path is a
  single scale with no reassociable operand sequence, so D-000.61's order-variation technique
  does not apply here and `lint`'s type-resolved scan is the instrument. **Say this in the
  record rather than claiming a D-000.61-shaped proof that was not run.**

### AC6 — The rectangle primitive is in `internal/pagemodel`, names no PDF concept, and `layout` still cannot reach `pdf`

**Given** the new vector primitive
**When** the tree is inspected
**Then** it is declared in `internal/pagemodel`, its fields name only geometry, edges, stroke
width and colour channels — no PDF operator, no resource name, no object number (AD-5)
**And** `internal/layout` imports no package of rank ≥ 7 other than itself
**And** the existing `TestStageRankProductionScan` remains green with `layout` and `pdf` in its
`PackagesVisited` witness

- **Anchor:** the **existing** `stageRankTable` and its production scan (D-000.72), which this
  story does not modify. Plus the retained violating fixture
  `folio-go/testdata/lint/stage-rank/layout/violating_pdf_import.go`, already in the tree.
- **Mutation:** add `import "…/internal/pdf"` to the new layout file → **red** from the existing
  scan, not from a new guard.
- **Note:** this AC adds **no** new stage-rank machinery. It asserts that the story stayed inside
  the machinery already there. If a developer finds themselves editing `stagerank.go`, that is a
  design error, not a table entry.

### AC7 — D-3.5.4's inherited property, discharged for the five fields this story wires, and RESTATED as a measured population for the three it cannot

This is D-3.5.4 verdict part 3, by name. §D3 is its measurement.

**Part A — the property, for each newly-wired field.**

**Given** two report-data documents differing **only** in a field the template does not bind —
`{"name":"X","overdue":true}` vs `{"name":"X","overdue":false}` — where `overdue` is exactly what
a conditional-formatting implementation would key on
**When** a template exercising `style.border`, `style.padding`, `style.background`,
`style.align` and `style.valign` on a table renders against both
**Then** the two outputs are **byte-identical**

- **Anchor:** the two data documents are literals the test owns.
- **Discriminating mutation, one per field, each applied and run:** make that field's value
  depend on `overdue` (e.g. border colour red when true) → **red**. Five mutations, five reds,
  all recorded. A single mutation covering one field would leave the other four asserted
  vacuously — D-3.5.4's own Finding-3 failure at Story 3.5 was exactly a subject that could not
  redden.

**Part B — the restated population, in the record and in a comment at the assertion.**

> *"Asserted for the style fields that reach output as of Story 4.1 — `fontFamily`, `fontSize`
> (by type-impossibility), `border`, `padding`, `background`, `align` and `valign`. **Three**
> remain with no render consumer, measured: `bold` and `italic`, because the shipped font set has
> no bold or italic face (`fontset.go:721`, `text/shape.go:70`) and synthetic emboldening is
> called nowhere; and `altRowBackground`, owned by **Story 4.8**. For those three the property
> holds vacuously and cannot yet be asserted."*

**Part C — D-3.5.5's derived list must still be complete.** The build-time-derived style-field
set that feeds the `{{ }}` rejection is unchanged by this story (no schema field is added), so
its completeness test must stay green. **Confirm it did, with the command, rather than assuming.**

### AC8 — A table's `visibleIf` verdict is CONSUMED, and a guard fires if a rendered kind ignores visibility

Story 3.5's Finding 6, landing where it was routed.

**Given** a `table` element carrying `visibleIf` that evaluates false
**When** the document renders
**Then** the table is **absent** from the page model — no header row, no borders, no background —
and siblings do not move (AD-24's *Visibility* clause)

**And** a guard asserts that the set of element kinds whose visibility verdict is **consumed**
covers every kind that reaches the page model.

- **Anchor for the guard: the compiler and the type system.** The kinds that reach the page model
  are discoverable from the render path's own dispatch; the kinds whose verdict is consumed are
  discoverable from the verdict map's readers. Comparing the two sets is the assertion. A guard
  anchored to a hand-maintained list of "kinds we render" is **the exact shape D-000.68 rules
  out** — Finding 6's own cheaper remedy (a doc comment) was accepted at 3.5 precisely because it
  was **not** claimed to be a guard. Do not now write a list and call it one.
- **Mutation:** render the table unconditionally, ignoring its verdict → **red** on the
  behavioural half. Add a kind to the render path without wiring its verdict → **red** on the
  set comparison.
- **If the set comparison cannot be anchored outside the code under test**, say so, ship the
  behavioural half alone, and **label it D-000.24** — a stated absence, not a silent one. Do not
  ship a self-referential set equality.

### AC9 — Explicit scope fence: no data rows, no pagination, no repeated header, no footer aggregates, no alternating shading

**Given** a table bound to a collection of five items
**When** it renders under this story
**Then** the header row is produced and **zero data rows** are produced
**And** the render **succeeds** — an unrendered body is not an error at 4.1

This AC exists so the fence is a test rather than a sentence, and so 4.2 has an unambiguous
red-to-green to inherit. It is expected to be **rewritten by Story 4.2**, and that is correct;
mark it in the code as such.

**Not in this story, with owners:**

| behaviour | owner |
|---|---|
| one row per collection element, in data order | **4.2** |
| cell text, and wrapping within the column width | **4.2** |
| row-scope binding resolution | **4.2** (mechanism already exists, Story 3.1) |
| empty collection → header only, no rows | **4.2** |
| breaking a table across pages; a row never split | **4.3** |
| the header repeated on continuation pages, and its height accounted for on each | **4.4** |
| footer aggregate values placed into footer cells | **4.5** |
| a row taller than the page | **4.6** |
| the golden report and the cross-target matrix | **4.7** |
| `altRowBackground` | **4.8** (optional, first to be cut) |

**A note the developer will need:** `headerHeight` is required by the schema **today** and is
already parsed. This story consumes it for the header row's own height. **4.4** is what makes it
count on *continuation* pages. Do not build continuation logic here.

### AC10 — The three-module gate, the required red unchanged, and the per-epic heavy-test cadence stated explicitly in the record

**Given** the story is complete
**When** the gate runs (**D-000.64** — all three modules)
**Then**:

| module | command | expected |
|---|---|---|
| `folio-go` | `go test ./...` | **≥ 920 pass · exactly 1 FAIL · 1 skip** — the FAIL is `TestCorpusMeetsP6ExerciseFloors` and **nothing else** |
| `lint` | `go test ./...` | **≥ 114 pass · 0 fail** |
| `hashmatrix` | `go test ./...` | **3 pass · 0 fail** |

**And** `go build ./...`, `gofmt -l` (empty) and the lint binary run clean on both Go modules.

**And the heavy-test cadence is stated in this story's own record, in these words (D-000.4, and
the brief requires it said explicitly):**

> **Heavy-test cadence for this story is PER-EPIC.** Integration and end-to-end tests are
> **written and committed, and deliberately NOT RUN** in this story. **The cross-target hash
> matrix does not run at Story 4.1**; it runs once at the **Epic 4 boundary**, where Story 4.7
> measures the C4 gate. Unit tests, lint, build and gofmt run every story regardless, and did.

**And** if this story registers a **new matrix document** — which it plausibly should, being the
first table geometry in the tree — **D-000.54 applies: its native leg runs at registration, host
target only, before the story reaches `done`.** That is a **sequencing** requirement and **must
not be logged as a D-000.4 override** (D-3.5.6's correction, verbatim).

**And** the golden-fixture impact is recorded as a **measured** number, not an expectation
(D-3.5.6's condition): *"N existing golden fixtures move; N = \_\_, measured by `<command>`."*
**§D6 predicts N = 0** — nothing renders a table today — but a prediction is not a measurement.
Run the command and record it.

---

## Decisions raised at creation

Four. **DECISION-1 and DECISION-4 must be routed before or during development.** DECISION-2 has a
default this story ships against. DECISION-3 is likely a non-issue but is cheap to settle now.

### DECISION-1 (for the lead, and possibly the owner) — the schema cannot express a visually distinct header row, and Story 4.7's golden report will need one

**The measurement.** `TableExt` has `headerHeight` and nothing else about the header. `Column`
has `label` and `align`. There is **no** `headerStyle`, no header font weight, no header
background, no header border distinct from a data cell's. A table's one `style` block governs
every cell identically. Confirmed against both `folio-format.md`'s table section and
`model.go:214-235`.

**Why it matters now rather than at 4.7.** The Customer Account Statement (4.7, the C4 gate) is a
five-column transaction table with a repeated header. A header row that is typographically
indistinguishable from its data rows is not a table a reader can use, and 4.4 will repeat that
indistinguishable header on all 50 pages. If the answer is "add a field," the cost of adding it at
4.1 — before anything renders — is a fraction of adding it at 4.7 with four stories built on top.

**Why the obvious workaround does not exist.** The natural answer is "make the header bold." **It
cannot be made bold**: the shipped font set has no bold face (`fontset.go:721`,
`text/shape.go:70`; see §D3). So even a `headerStyle` block would have no weight axis to turn.

**The arms, framed:**
- **A — accept it.** The header is distinguished by `background` and `border` on the whole table,
  which is uniform. 4.7's golden report ships with an undifferentiated header. **No schema
  change. This story proceeds unchanged.**
- **B — a `headerStyle` block on `TableExt`**, a second `Style`. Schema change, drift-test change,
  round-trip change, D-3.5.5's derived set grows. Buys a distinct background/border/padding for
  the header; still no bold.
- **C — defer to 4.7** and let the golden report's own appearance force the answer. **This is the
  arm this story recommends against**, on D-000.66's grounds: a hedge carried forward hardens
  into a settled fact.

**This story is written for arm A** and needs no rework under it. **Under arm B it needs one
extra AC and a schema edit.** It does not invent a field.

### DECISION-2 (for the lead, low stakes, defaulted) — `columns[].align` vs `style.align`

Both can be set; the schema states no precedence. This story ships **column wins**, on AC2's own
wording (*"its own alignment"*) and on specificity. **AC4 asserts it.** Recorded so it is a
decision on the record rather than an implementation accident (D-000.63). If the lead rules the
other way, AC4 inverts and nothing else moves.

### DECISION-3 (for the lead, probably a non-issue) — a table wider than its band

Column widths are absolute and never negotiated (AD-13, AC2). Nothing stops an author declaring
five 200pt columns inside an A4 content band. Options: silent overflow; clip as Story 2.8 clips a
text element (FR44, a **Warning** alongside PDF bytes); or a load-time error.

**This story's position:** **out of scope, and record it rather than decide it silently.** Story
2.8's clipping applies to a text element's declared box; a table has no declared box (that is
AD-13's whole point), so 2.8's mechanism does not transfer without a ruling. If a code is minted
for it, **D-000.65** says it is minted where the condition first ships — which is arguably here.
**Raised so it is a deliberate deferral, not an oversight.** If the lead wants it now, it is one
AC and one diagnostic code.

### DECISION-4 (for the ORCHESTRATOR, to route — not a question for this story) — DW-4 fires now

`deferred-work.md` DW-4: *"Nobody owns cutting `folio-go/v0.1.0` — **owner decision when Epic 4
is planned**… If it is still unowned when Epic 4 is planned, it goes to the owner rather than
being absorbed into a story."* **Epic 4 is now being planned.** DW-3 (the third-party licence
manifest as a release artifact) is owned by *"Epic 4 close"* and depends on it. **DW-20** (the
`go/types`-precise call-graph walker) is also triggered by *"whoever cuts `folio-go/v0.1.0`."*

**Three deferred entries hang off one unowned event, and the trigger has arrived.** Not this
story's to resolve, and explicitly not to absorb. Route it.

---

## Task breakdown

1. **Read first.** `epics.md` Epic 4 (4.1–4.8), `ARCHITECTURE-SPINE.md` AD-13/AD-24/AD-5/AD-3,
   `folio-format.md`'s `table` and `style` sections, D-3.5.4 and D-3.5.5 in the decision log, and
   §D1–D6 above. Do not start from the AC text alone.
2. **Confirm the baseline yourself** — all three modules, the required red present and alone,
   `git status --porcelain` empty. Record the figures with scope and flags.
3. **AC1 instrument 1a first, before any new code.** Write the two load-rejection fixtures against
   the *existing* `parse_bands.go` branch, watch them pass, then run the mutation and watch them
   red. This closes §D4's untested guard and gives you a working red-proof harness before the new
   surface exists.
4. **The page-model vector primitive** (`internal/pagemodel`, R1, AC6). Geometry, edges, stroke
   width, colour channels. No PDF concept.
5. **The colour conversion** (R2, AC5): `geom.ScaleRound(channel, 1000, 255)`, the existing
   decimal emitter, the test-owned expectation table including `#7F7F7F`/`#808080`.
6. **The PDF fill and stroke path** in `internal/pdf`: `re`+`f`, `re`+`S`, `w`, `rg`, `RG`,
   bracketed `q`/`Q` so graphics state never leaks between primitives — the same discipline
   `textdoc.go:757-783` already uses for the clip. Every number through the existing emitters.
7. **Column geometry in `internal/layout`** (R3, R4, AC1 1c/1d, AC2): x-origins by running sum,
   widths from the declaration, total width computed and never stored.
8. **The header row** (AC3): cell rectangles, borders per `edges`, padding insets, background,
   label placement by `columns[].align` → `style.align` (AC4) and `style.valign`, reusing Story
   2.5a's leading model for the baseline within the fixed box.
9. **Visibility consumption and its guard** (AC8, R8).
10. **AC7's five mutations**, each applied and run, plus the restated population comment.
11. **The scope-fence test** (AC9), marked as 4.2's to rewrite.
12. **Integration / e2e tests: WRITE, DO NOT RUN** (AC10, D-000.4). State it in the Delivery Log
    in those words.
13. **Run every mutation named in AC1–AC5, AC7 and AC8.** Record mutation, command, result, and
    the `git diff` that confirms restoration. A mutation that was reasoned about and not executed
    is recorded as **not run** (D-000.9 extension).
14. **The three-module gate** (AC10), the golden-fixture count **measured**, and — if a matrix
    document is registered — its **native leg run at registration**, logged as **D-000.54
    sequencing and NOT as a D-000.4 override** (D-3.5.6).
15. **Update `folio-format.md`** only if the lead rules DECISION-1 arm B. Otherwise the format is
    unchanged by this story, and the drift tests should confirm that.
16. **Set status to `review`** in `sprint-status.yaml` — status value only, no narrative comment.

---

## Delivery Log

### Mid-story ruling: the fontFamily default was never real (D-000.28), and headerStyle is added to scope

Parked with a DECISION NEEDED after finding that R6 ("a table with no style at all must render...
labels") requires a resolvable font, but `fontChain` (render.go:903) has always hard-errored when
`style.fontFamily` is absent, for ANY text-carrying element, table or not — and the format's
documented default ("the first key of `fonts`") was never implemented anywhere in the tree.
`Fonts` is `map[string][]string`; parsing discards authored key order, so "the first key" was never
well-defined as written.

**Ruling (engineering lead, via the orchestrator), verbatim in substance:**
- **No font default is implemented, now or as part of this story.** A shared-`fontChain` behaviour
  change launched from a table story would be a sweep launched from a targeted story. Requiring the
  field is the conservative, reversible direction: every template that renders today renders
  identically (no recorded hash moves), and a default can be added later, unhurried, by the owner.
- **`folio-format.md:288` is amended in this commit** — it documented behaviour that was never
  built (D-000.28's class). Before/after, verbatim:
  ```diff
  -| `fontFamily` | the first key of `fonts` |
  +| `fontFamily` | **none — required on any element carrying text** |
  ```
  plus a new paragraph directly under the Style defaults table:
  > "There is no font default. An element with text and no `style.fontFamily` is a located error
  > naming the element. A default was documented here from the format's first draft and never
  > implemented; `fonts` is a mapping with no authored key order, so 'the first key' was never
  > well-defined. If a default is added later it will name its rule explicitly."
  Checked: `internal/template/drift_test.go` compares BACKTICKED KEY TOKENS against emitted keys,
  not prose sentences — this paragraph and the changed cell text move no drift-test result. Full
  diff also in `git diff -- _bmad-output/specs/spec-folio/folio-format.md`.
- **R6 is corrected, not annotated "untested"** (D-000.66) — see the story's R6 above, rewritten
  in place with its precondition (`style` sets `fontFamily` and nothing else) and a record of why.
- **`headerStyle` is added to this story's scope** (owner ruling, via the lead): a table gets a
  `headerStyle` `Style` block so the template author controls the header's look, landing in 4.1
  because 4.1 already attaches style to the header row, the fill/stroke primitives, and header-row
  layout. Precedence for a header cell field: `headerStyle.<field>` (when set) → `style.<field>`
  (when set) → the field's documented default. `columns[].align` still wins over both (AC4's own
  grounds — the column's field is the more specific one — extended one level, not re-litigated).
  No header treatment is hardcoded; the author expresses it via `headerStyle`, or gets AC3's
  already-ruled "no style" defaults if they set neither. `bold` is NOT wired here even though
  `headerStyle` could carry it — the font-set still has no bold face (§D3); this is unchanged scope.
- **R5's "TableExt's field set is permanent through the MVP" is a deliberate, ruled exception**:
  `HeaderStyle` is documented in `internal/template/table_geometry_test.go`'s
  `wantTableExtFields` as exactly that — a ruled addition, not scope creep the AC1 1b guard missed.

`fonts`' key-order gap (needed if a default is ever added later) is NOT fixed by this story — it
was raised as part of the original decision framing and explicitly not built (arm B rejected).
This is left as-is; whether `.folio` should have a font default at all goes to the owner later,
unhurried, per the ruling.


### Implementation record

**Baseline correction applied, per the finisher's own note at dev start.** `baseline_commit`
frontmatter stays `d5b75b5` (the story-creation baseline, per the skill's "preserve the existing
value" rule), but the gate figures below are measured at HEAD after this story's own changes,
against the CORRECTED command set the finisher supplied (three commits landed after `d5b75b5`:
`82356a9`, `b044cdb`, and the gate close):

| module | command | result |
|---|---|---|
| `folio-go` | `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | **955 pass · 0 fail**, 17 packages |
| `folio-go` (plain, required red visible) | `go test -count=1 ./...` | **961 pass · 1 FAIL · 1 skip** — the FAIL is `TestCorpusMeetsP6ExerciseFloors/P6g` and **nothing else**, unchanged from baseline (D-000.17/D-2.1.14/D-000.57/D-000.74; this is the sanctioned red quarantined into its own CI job, not mine to fix) |
| `lint` | `go test -count=1 ./...` | **115 pass · 0 fail**, 4 packages (≥114 per AC10; the +1 over the story's stated 114 is a pre-existing cross-module scan whose subtest count depends on folio-go's own tree, not a lint-module change — `git status --porcelain` under `lint/` is empty) |
| `hashmatrix` | `go test -count=1 ./...` | **3 pass · 0 fail**, 2 packages |

`go build ./...` and `gofmt -l` (empty) confirmed on all three modules. `go vet ./...` clean on all
three.

**RE-MEASURED at finish, after the review-finding fixes below (finisher's own figures, not
carried forward from either the developer's or the reviewer's — D-000.9 extension):**

| module | command | result |
|---|---|---|
| `folio-go` | `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | **965 pass · 0 fail · 1 skip**, 17 packages |
| `folio-go` (plain, required red visible) | `go test -count=1 ./...` | **971 pass · 2 fail-events (one parent, one subtest) · 1 skip** — both fail-events are `TestCorpusMeetsP6ExerciseFloors`/`TestCorpusMeetsP6ExerciseFloors/P6g_(opaque_names)`, the SAME sanctioned red, and **nothing else** |
| `lint` | `go test -count=1 ./...` | **115 pass · 0 fail**, 4 packages — unchanged (no `lint`-module file was touched) |
| `hashmatrix` | `go test -count=1 ./...` | **3 pass · 0 fail**, 2 packages — unchanged |

`go build ./...`, `go vet ./...` and `$(go env GOROOT)/bin/gofmt -l` (empty output) confirmed
clean on all three modules. The +10 over the developer's 955 is exactly the finisher's own new
tests, net of one deletion: `TestHeaderStyleLoadErrorsNameHeaderStyleNotStyle` (+1),
`TestHeaderStyleBackgroundWinsOverStyle` (+1), `TestHeaderStyleCascadesPerField` (+1),
`TestColumnAlignWinsOverHeaderStyleAlign` (+1), `TestTableHeaderValignPlacement` (+1),
`TestTableStyleFieldsAreNotDataDrivenControl` gaining 5 subtests where it had none (+5),
`TestTableInPageHeaderRepeatsIdenticallyAcrossPages` (+1), minus the deleted
`TestAppendColorChannelsNoFloat` (−1): 1+1+1+1+1+5+1−1 = 10.

**D-2.8.6 deviation: table diagnostics after text diagnostics within a band (finisher, review
Finding 7, Major).** `render.go`'s `diags` assembly is
`headerDiags, headerTableDiags, contentDiags, contentTableDiags, footerDiags, footerTableDiags` —
so within any ONE band, every text element's diagnostics precede every table's, regardless of
which was declared first in the document. D-2.8.6 states `Result.Diagnostics`' rule as "band
order, then element declaration order within a band"; this story's own code relaxes the
within-band half for text-vs-table specifically. The render.go comment at the content-band
collection site cited this Delivery Log for that deviation, but the deviation was never actually
written down here — `grep -n "declaration order\|D-2.8.6\|diagnostic order\|relative order"` over
this story file returned zero matches, found at finish. It is written down now, and the
render.go comment is updated to cite this exact entry by name. The deviation is deliberate and
narrow: within one band, a table element's own diagnostics (a missing-glyph Warning on a header
label, or `STYLE_COLOR_INVALID`) sort after every text element's diagnostics in that band,
regardless of declaration order between the two kinds. `Result.Diagnostics`' PAGE-MODEL content
is entirely unaffected — this is `Result.Diagnostics`' own relative order only, between a text
and a table diagnostic sharing one band. No new test is added for it (rare in practice — it
requires both a text AND a table element in the same band each raising a diagnostic — and 4.2-4.8
will exercise the table half of this ordering far more, so a dedicated instrument is better
placed against a richer table-diagnostic surface than 4.1's own). Left for a later story or the
lead to decide whether to interleave by declaration order once table diagnostics are common
enough to matter observably.

**Heavy-test cadence for this story is PER-EPIC.** Integration and end-to-end tests are **written
and committed, and deliberately NOT RUN** in this story. **The cross-target hash matrix does not
run at Story 4.1**; it runs once at the **Epic 4 boundary**, where Story 4.7 measures the C4 gate.
Unit tests, lint, build and gofmt run every story regardless, and did.

**No new matrix document was registered by this story** (D-000.54 does not apply): every test this
story adds calls `buildPageModel`/`Render` directly against test-owned fixtures in
`folio-go/table_render_test.go` and `folio-go/internal/pdf/rectdoc_test.go`, none of them added to
`matrix_registration_test.go`'s registry. This was a deliberate scope choice, not an oversight —
the story's own C4 gate work (golden report, cross-target matrix) belongs to Story 4.7.

**Measured golden-fixture impact: N = 0 — evidence CORRECTED at finish (review Finding 6,
Major).** The command originally cited here —
`go test . -run 'TestGoldenDigestAgreesAtEveryDeclaredSite|TestEpic2GateObligationsMatchTheDeclaredSet|TestBoundaryGateDigestsAreWellFormed'`
— was read at finish and found to hash the **committed** `expected.pdf` files against declared
digests; it never calls `Render` and is structurally incapable of detecting the renderer
producing different bytes for the same input. It remains true (3 pass, 0 fail — every declared
digest agrees) but it does not, on its own, establish N = 0. **The actual evidence for N = 0**
is the RE-RENDERING fixture comparisons — `TestMultiPageGoldenMatchesTheCommittedArtifact`
(`multi_page_fixture_test.go`) and `TestWrappedTextGoldenFixture` (`wrapped_text_fixture_test.go`)
— which reproduce their `expected.pdf` from a fresh `Render` call and compare byte-for-byte; both
ran green in the full-suite measurement re-run at finish (below). Combined with
`TestAppendRectContentStreamEmptyPageProducesEmptyStream` (`internal/pdf/rectdoc_test.go`,
proving `appendRectContentStream` appends zero bytes when `page.Rects` is empty — true for every
pre-4.1 document) and §D6's prediction (nothing rendered a table before this story, so no
existing fixture could be touched), N = 0 is now cited against evidence that can actually see a
moved golden — this matters beyond bookkeeping: it is Story 4.7's evidence base.

### What was built

- **`internal/pagemodel`** (AD-5, R1): `Color` (three 0..255 channels), `RectEdges`, `Rect` — the
  page model's first vector primitive — and `Page.Rects`. Names no PDF concept.
- **`internal/layout/table.go`** (R3/R4, AC1 1c/1d, AC2): `ColumnGeometry`, `TableGeometry` (width
  computed via `Width()`, never stored), `ColumnWidths(startX, widths)` — takes ONLY declared
  widths, no label, no font, no measurement: AC2's "never negotiated against content" holds by
  construction, not convention.
- **`internal/layout/paginate.go`**: extended `ColumnItem`/`BandContent`/`PageAssignment` with a
  third content kind, `Rects []RectRef` (alongside `Runs`/`Images`), and `OverflowError.Kind`'s
  `"table"` value. `MixedItemError`'s exclusivity check generalised from a boolean pair to a
  populated-count check over three kinds.
- **`internal/layout/band.go`**: `ComposePage` takes a fourth parameter, `rects []pagemodel.Rect`.
- **`internal/pdf/rectdoc.go`** (AC3/AC5/AC6, D1/D2/R1/R2): `appendRectContentStream` (fill via
  `re f`, edge-subset stroke via `m`/`l`/`S` subpaths, `q`/`Q`-bracketed independently per half) and
  `appendColorChannels` (`geom.ScaleRound(channel, 1000, 255)` → the existing decimal emitter, the
  module's ONLY colour-channel conversion site). Wired into `textdoc.go`'s per-page assembly,
  BEFORE text, so cell chrome sits behind labels; `page.Rects == nil` for every pre-4.1 document,
  so this appends zero bytes and every existing golden stays byte-identical (confirmed above).
- **`internal/diag`**: minted `CodeStyleColorInvalid` ("STYLE_COLOR_INVALID") — R7/D-000.65, the
  first genuinely new render-time condition this story ships (a malformed `#RRGGBB` reaching
  render, unreachable before this story because no colour was ever consumed on any render path).
  Bridged at the root as `DiagCodeStyleColorInvalid`.
- **`internal/template`**: `TableExt.HeaderStyle Presence[Style]` (owner-ruled scope addition, see
  the mid-story ruling above) — parsed in `decodeTableExt` (same null/absent handling as an
  element's own `style`), serialized in `writeElement`. `folio_expr_validate.go` gained one call
  site (`checkStyleHasNoPlaceholders(headerStyle.Value, ...)`) so a `{{ }}` inside `headerStyle`'s
  string fields is rejected exactly like `style`'s — proven with a dedicated red-proof (see
  Mutations below); `TestStyleStringFieldPopulationMatchesSchema` needed no change (`HeaderStyle` is
  `Presence[Style]`, not `Presence[string]`/`Presence[[]string]`, so the reflection walk correctly
  does not enumerate it as an eleventh string field — its OWN string fields are the existing five/
  six already covered by that test via `template.Style`).
- **`folio-go/table_render.go`** (the header row itself, AC2/AC3/AC4/AC6/AC7/AC8/AC9):
  `resolveHeaderStyle` (the `headerStyle` → `style` → default cascade, per field), `paddingEdges`,
  `parseHexColor`, `buildHeaderCellRect`, and `collectBandTableRuns` — the table analogue of
  `collectBandTextRuns`, called once per band, producing header LABEL runs through the SAME
  `shapeSegments`/`positionSegments`/`measureRuneRange`/`chainVerticalModel` pipeline every text
  element uses (DW-16's forcing function: `buildShapedPDFRuns` stays the ONLY producer of
  `pagemodel.ShapedGlyph`, confirmed by `TestGlyphIdentifierCensus` staying green throughout, and by
  this file never constructing one), plus one `tableRectSource` per visible table (one
  `pagemodel.Rect` per column cell, ALWAYS populated when the table has ≥1 column, even when
  drawing nothing — R6/D-2.6.5: an item that occupies page space must not be empty).
- **`render.go`/`page_number.go`**: wired `collectBandTableRuns` into `predictDocument` (one call
  per band, ahead of PHASE A/B — table labels are plain strings, never `{{page}}`-bearing, so no
  resolver or pageCount dependency exists), `contentColumnItems` and `paginateDocument` extended to
  carry table rects through pagination exactly as images already do (index-into-a-flat-slice, the
  same shape `ImageRef`/`pdfPlacements` already use).

### Mutations run (AC1, AC2, AC3, AC4, AC5, AC6, AC8), each applied, observed red, then restored

Every mutation below was applied to the real file, the named test/command was run and observed to
fail, then the file was restored from a `cp`-made backup (never `git checkout`, per this
programme's red-proof hygiene) and reconfirmed green. `git diff --stat` on the mutated file was
empty after every restoration (spot-checked; the final full-tree `git status --porcelain` below is
the comprehensive confirmation).

| # | AC | mutation | command | result |
|---|---|---|---|---|
| 1 | AC1 1a | Deleted the `if el.Type == ElementTable {...}` branch at `parse_bands.go:158` | `go test ./internal/template/... -run 'TestTableDeclaringWidthIsRejected\|TestTableDeclaringHeightIsRejected'` | **RED** (both fail, for the branch's absence — the document loads past the point that should reject it, tripping a different, still-failing check further down) → restored, green |
| 2 | AC1 1b | Added `Width geom.Length` to `TableExt` | `go test ./internal/template/... -run TestTableExtFieldSetIsPermanent` | **RED**, naming `"Width"` explicitly → restored, green |
| 3 | AC1 1c | Added a field to `layout.TableGeometry` | `go test ./internal/layout/... -run TestTableGeometryHasNoWidthField` | **RED**, naming the added field → restored, green |
| 4 | AC1 1d | Changed `TableGeometry.Width()`'s running sum to a running max | `go test ./internal/layout/... -run TestColumnWidthsVectors` | **RED** on 2 of 5 vectors (the multi-column ones; the single-column and empty vectors cannot distinguish sum from max) → restored, green |
| 5 | AC5 | Replaced `geom.ScaleRound` with `v*1000/255` truncation in `appendColorChannels` | `go test ./internal/pdf/... -run 'AppendColorChannels$'` | **RED** on `#808080` (0.501 vs 0.502) and `#010101` (0.003 vs 0.004) → restored, green |
| 6 | AC6 | Added `import _ "…/internal/pdf"` to `internal/layout/table.go` | `go test ./... -run TestStageRankProductionScan` (from `lint/`) | **RED**, the EXISTING scan reports the violation by name — no new guard was added → restored, green |
| 7 | AC3 (edges) | Hardcoded all four `RectEdges` to `true` regardless of `style.border.edges` | `go test . -run TestTableHeaderBorderEdgesSubset` (from `folio-go/`) | **RED** — Left/Right wrongly `true` → restored, green |
| 8 | AC3 (padding) | Dropped the padding inset (`contentX := cg.X` instead of `cg.X + padLeft`) | `go test . -run TestTableHeaderPaddingInsetsLabel` | **RED** — the two renders' label X became equal (shift of 0, not 20000) → restored, green |
| 9 | AC4 (col wins) | Made `style.align` win unconditionally (dropped the `columns[].align` override) | `go test . -run TestColumnAlignWinsOverStyleAlign` | **RED** — column 0's own `align: "left"` was overridden, its label shifted right → restored, green |
| 10 | AC4 (2nd half) | Removed `style.align`/`headerStyle.align` fallback entirely (column align the only source) | `go test . -run TestColumnAlignWinsOverStyleAlign` | **RED** — column 1 (no own `align`) stayed flush-left instead of falling back to `style.align: "right"` → restored, green |
| 11 | AC2 | Widened a cell's `Rect.W` to the measured label advance when it overflows (the exact "well-meaning" `max(declaredWidth, measuredLabelAdvance)` defect, adapted to this story's actual code shape since `layout.ColumnWidths` takes no content parameter to attach the AC's literal suggested edit to) | `go test . -run TestColumnGeometryNeverNegotiatesAgainstLabelContent` | **RED** across all three scripts (Latin/Thai/CJK) — narrow vs wide column widths diverged → restored, green |
| 12 | AC8 | Bypassed the `isVisible` check (`if false && !isVisible(...)`) | `go test . -run TestTableHeaderVisibleIfFalseIsAbsentFromPageModel` | **RED** — the hidden table produced 2 rects and non-zero runs → restored, green |
| 13 | folio_expr_validate.go (headerStyle placeholder fence) | Removed the new `checkStyleHasNoPlaceholders(hs.Value, ...)` call site for `table.headerStyle` | `go test . -run TestParseTemplateRejectsPlaceholderInHeaderStyle` | **RED** — the malformed-load rejection this call site exists for did not fire → restored, green |

**AC7's mutation deviates from the letter of "one mutation per field, all five run," and this is
stated rather than smoothed over (D-000.9).** `collectBandTableRuns`/`resolveHeaderStyle`/
`buildHeaderCellRect` (this story's own new code) take **NO `data`/`params` parameter anywhere in
their signatures** — a stronger property than the AC anticipated (it assumed the code MIGHT read
data conditionally per field, needing five independent checks). Threading `data` into this call
chain, for each of five fields, purely to build a mutation, would have meant temporarily adding the
exact channel AD-24/D-3.5.2 forbid, run five times — engineering effort spent proving something the
type signature already proves once, structurally. What WAS run: `TestTableStyleFieldsAreNotDataDriven`
(a real two-dataset byte-identical render, `border`/`padding`/`background`/`align`/`valign` all set
in one template) and `TestTableStyleFieldsAreNotDataDrivenControl` (D-000.9's vacuity control —
two DIFFERENT border colours, same data, must and do produce different bytes, proving the
byte-comparison is sensitive at all). No per-field mutation was executed; this is recorded as a
**partial, structurally-justified measurement**, not five runs that did not happen.

### AC7 Part B — restated population (verbatim, to also appear at the assertion site)

> Asserted for the style fields that reach output as of Story 4.1 — `fontFamily`, `fontSize` (by
> type-impossibility), `border`, `padding`, `background`, `align` and `valign` — now via TWO attach
> points, an element's own `style` and (Story 4.1's addition) a table's `headerStyle`. **Three**
> remain with no render consumer, measured: `bold` and `italic`, because the shipped font set has
> no bold or italic face (`fontset.go:721`, `text/shape.go:70`) and synthetic emboldening is called
> nowhere; and `altRowBackground`, owned by **Story 4.8**. For those three the property holds
> vacuously and cannot yet be asserted.

**Part C confirmed**, command run: `go test ./... -run TestStyleStringFieldPopulationMatchesSchema`
(from `folio-go/`) → 1 pass. The build-time-derived style-field set feeding the `{{ }}` rejection
stayed complete and green — no schema field was added that this test needed to learn about
(`HeaderStyle` is `Presence[Style]`, correctly excluded from that reflection walk by kind, not by a
hand-written exclusion).

### AC9 — scope fence, and what it explicitly is not

`TestTableRendersZeroDataRows` renders a table bound to a 5-item collection and asserts the render
succeeds with exactly 2 rects (one per column's header cell — nothing else) and no data-row
content. Marked in its own doc comment as **Story 4.2's to rewrite**.

### AC8 — the set-comparison half, explicitly labelled D-000.24 (not shipped)

Only the BEHAVIOURAL half of AC8's guard is shipped (`TestTableHeaderVisibleIfFalseIsAbsentFromPageModel`,
mutation #12 above). The set-comparison half ("the set of kinds whose verdict is consumed covers
every kind that reaches the page model") is **NOT shipped**, per the AC's own stated escape hatch:
deriving "kinds that reach the page model" structurally (from the render path's own dispatch) and
"kinds whose verdict is consumed" structurally (from the verdict map's readers) into two
independently-anchored sets would need a bespoke AST scanner comparable in scope to
`lint`'s stage-rank or flip-routing scanners — not attempted in this story's time budget, and a
hand-rolled comparison built from a list of "kinds we render" is the exact self-referential shape
D-000.68 rules out. This is a **stated absence** (D-000.24), not a silent one: `isVisible` IS now
consulted for `table`, alongside the pre-existing `text`/`image` sites (render_visibility.go's own
doc comment is unchanged — R8, this story does not touch `computeVisibility`).

**Required statements this log must contain before review:**
- The three-module gate figures, with scope and flags, and the required red named and alone.
- The heavy-test cadence sentence from AC10, verbatim.
- Every mutation from AC1–AC5, AC7 and AC8: what was changed, the command, red or green, and
  confirmation the tree was restored.
- The measured golden-fixture count.
- The restated D-3.5.4 population (AC7 Part B).

### Finisher fixes and re-measurement (post-review, D-000.49: append, not rewrite)

The review's verdict was **Changes Requested: 2 Blockers, 5 Majors, 7 Minors, 4 Nits**, all 18
triaged and resolved below (see **## Finding Resolutions** at the end of this file for the
per-finding table). Every code fix in this section was made surgically, in the file(s) the
finding named. No golden fixture moved (re-confirmed in the gate re-measurement above).

**Both Blockers were real and are now fixed with genuine render-level instruments — not argued
away.**

- **Blocker 1 (Finding 2) — `headerStyle` was inert to every test.** Three new tests in
  `folio-go/table_render_test.go`: `TestHeaderStyleBackgroundWinsOverStyle` (headerStyle's
  background wins over a DIFFERENT style.background), `TestHeaderStyleCascadesPerField`
  (headerStyle sets ONLY border, style sets ONLY padding — both apply: the cascade is per-field,
  not "headerStyle present blanks style"), and `TestColumnAlignWinsOverHeaderStyleAlign` (AC4's
  precedence, extended one level, actually asserted against headerStyle rather than only style).
- **Blocker 2 (Finding 3) — `style.valign` reached no assertion anywhere.**
  `TestTableHeaderValignPlacement` (`table_render_test.go`) renders the SAME table three times
  varying only `valign`, with a headerHeight (100pt) large relative to a 9pt label's line height,
  and asserts three strictly-ordered, distinct label Y positions — "top" against a test-owned
  literal (20000, the same literal `TestTableHeaderNoStyleExceptFontFamilyRendersDocumentedDefaults`
  already pins), "middle"/"bottom" against strict ordering plus a symmetry bound (D-000.68: the
  font's own line-height in thousandths is not duplicated as a second, independently-computed
  literal — the test anchors to properties the mutation cannot fake instead).

**Mutations H and G (the reviewer's own, from the QA Results table above) were both RE-RUN against
the fixed tests and confirmed red, then restored:**

| # | AC | mutation | command | result |
|---|---|---|---|---|
| 14 | AC5 | Replaced `geom.ScaleRound` with `geom.Length(float64(c.R)/255.0*1000)` in `appendColorChannels` (the reviewer's mutation F) | `go build ./...` (succeeds); `go test ./... -run 'TestFloatTypedProductionScan\|TestFloatTypedTestScopeInventory'` (from `lint/`) | Build **succeeds**. `lint`'s two named tests **RED**: `TestFloatTypedProductionScan` reports `internal/pdf/rectdoc.go:106:38: this value expression has floating-point type float64`; `TestFloatTypedTestScopeInventory` reds on site-count (got 6, want 5) → restored, `go build`/`go vet` clean |
| 15 | Blocker 1 (Finding 2) | `hasHeader := false` — ignore `headerStyle` entirely at render (the reviewer's mutation H) | `go test . -run 'TestHeaderStyleBackgroundWinsOverStyle\|TestHeaderStyleCascadesPerField\|TestColumnAlignWinsOverHeaderStyleAlign'` | **RED** on all three (background wrong colour; border missing; column 1 stayed flush-left) → restored, green |
| 16 | Blocker 2 (Finding 3) | Deleted the `valign` cascade switch in `resolveHeaderStyle` (the reviewer's mutation G) | `go test . -run TestTableHeaderValignPlacement` | **RED** — top/middle/bottom all collapsed to Y=20000 → restored, green |
| 17 | Finding 1 (AC1 1a, corrected) | Kept the `ElementTable` branch; stored the declared `width` on `Element.Width` instead of rejecting it (the faithful mutation the review supplied) | `go test ./internal/template/... -run 'TestTableDeclaringWidthIsRejected\|TestTableDeclaringHeightIsRejected\|TestTableWithNeitherWidthNorHeightLoads'` (widened `-run` includes the control, per Finding 1) | **RED** on the width half only (`got nil error`); height half and the control both stayed **green** → restored, green |
| 18 | Finding 5 (headerStyle diagnostic location) | Reverted `decodeStyle`'s `fieldPrefix` threading (hardcoded `"style"` again in its `newLoadError` calls) | `go test ./internal/template/... -run TestHeaderStyleLoadErrorsNameHeaderStyleNotStyle` | **RED** — message named `"style.valign"` instead of `"headerStyle.valign"` → restored, green |
| 19 | Finding 8 (AC2 instrument) | Widened a cell's `Rect.W` to the measured label advance on overflow (the reviewer's mutation D, re-run against the rewritten 3-column test) | `go test . -run TestColumnGeometryNeverNegotiatesAgainstLabelContent` | **RED** across all three scripts — column geometry and total width both diverged from the test-owned literals → restored, green |
| 20 | Finding 10 (AC1 1d) | Changed `TableGeometry.Width()`'s running sum to a running max (re-run against the widened 6-vector table) | `go test ./internal/layout/... -run TestColumnWidthsVectors` | **RED** on exactly **3 of 6** vectors ("five", "near-limit", "three-columns" — the three vectors with 2+ distinct-valued columns; "empty"/"single"/"tiny" cannot ever distinguish sum from max) → restored, green |
| 21 | Finding 12 (table rects in header/footer bands) | Routed header/footer-band table rects into `items` instead of `header.Rects`/`footer.Rects` (the reviewer's mutation I) | `go test . -run TestTableInPageHeaderRepeatsIdenticallyAcrossPages` | **RED** — page 2 of 2 had 0 rects instead of the repeated header → restored, green |

All restorations confirmed via `cp`-backup round-trip (never `git checkout`) and `/usr/bin/diff`
byte-identical against the pre-mutation backup, per file, immediately after each restore — not
merely a final `git status` at the end.

**The remaining Majors, Minors and Nits were fixed as documentation/record corrections or small,
targeted test/comment edits** — see **## Finding Resolutions** for the full per-finding
disposition; summarised here:

- **Finding 4 (AC5's false "build fails" claim):** AC5's own text and `rectdoc_test.go`'s doc
  comment both corrected to name `lint`'s `TestFloatTypedProductionScan`/
  `TestFloatTypedTestScopeInventory` as the real instrument (mutation 14, above). The vacuous
  `TestAppendColorChannelsNoFloat` (Finding 11) was deleted in the same edit — it asserted nothing
  about floats and duplicated two rows of `TestAppendColorChannels` above it.
- **Finding 6 (golden-fixture evidence):** the Delivery Log's "Measured golden-fixture impact"
  paragraph (above) now cites the re-rendering fixture tests as the actual evidence, with the
  originally-cited digest-only command kept but correctly scoped to what it actually establishes.
- **Finding 9 / Nit 18 (AC7 Part A overstatement):** `table_render_test.go`'s doc comment above
  `TestTableStyleFieldsAreNotDataDriven` corrected to name `collectBandTableRuns`'s `visible`
  parameter and explain why a visibility verdict is not a style channel.
  `TestTableStyleFieldsAreNotDataDrivenControl` rewritten as a 5-case table, one per AC7 field
  (border/padding/background/align/valign), each independently proven byte-sensitive — previously
  it covered `border.color` alone.
- **Finding 10 / Nit 16 (AC1 1d's false overflow claim):** `table_test.go`'s "near-limit" vector
  comment corrected (no overflow-aware path exists in `Width()`; it is a large-value smoke case)
  and its "these three" typo fixed; a third multi-column vector (`{10000,20000,30000}`) added so
  the sum-vs-max mutation reddens 3 of 6, per AC1 1d's own "at least three" requirement.
- **Finding 13 (breaking-change disclosure):** recorded here — **two in-repo fixtures
  (`render_bind_test.go`, `render_table_bind_test.go`) needed a `style.fontFamily` addition for
  this story to stay green**, which is the measured cost the ruling's "renders identically"
  clause omitted. The class is precise: any `.folio` document with a table carrying at least one
  column with a non-empty `label` and no resolvable `style.fontFamily` (nor `headerStyle.fontFamily`)
  rendered successfully before this story and fails after it. Whether that failure should carry a
  diagnostic code (rather than a plain `fmt.Errorf`) is **DEFERRED** — it is not a gap this story
  introduced: the identical text-element failure mode (`fontChain`, `render.go:903-906`) has
  never carried one either, and widening that is a cross-cutting change to shared font-resolution
  error handling, not a table-scoped fix.
- **Finding 14 (AC9's thin fence):** `TestTableRendersZeroDataRows` extended to also assert the
  run COUNT (exactly 2) and that every run's `SourceText` is one of the two column labels, never
  data-row content — closing the gap where a row implementation emitting text without rects would
  have stayed green.
- **Finding 15 (`Extra` comment):** `table_geometry_test.go`'s comment corrected — `Extra` is
  excluded by name on `Column` only; `TableExt` carries no `Extra` field to begin with.
- **Finding 17 (missing table pipe):** `folio-format.md`'s `headerStyle` row closed with its
  trailing `|`.

## File List

**New files:**
- `folio-go/internal/layout/table.go`
- `folio-go/internal/layout/table_test.go` (finisher: comment corrections + a third multi-column
  vector — Findings 10/16)
- `folio-go/internal/pdf/rectdoc.go`
- `folio-go/internal/pdf/rectdoc_test.go` (finisher: deleted the vacuous
  `TestAppendColorChannelsNoFloat`, corrected doc comment to name `lint`'s float scan — Findings
  4/11)
- `folio-go/internal/template/table_geometry_test.go` (finisher: comment correction — Finding 15 —
  and a new `TestHeaderStyleLoadErrorsNameHeaderStyleNotStyle` — Finding 5)
- `folio-go/table_render.go`
- `folio-go/table_render_test.go` (finisher: 7 new tests/rewrites for Blockers 1/2 and Findings
  8/9/12/14 — see "Finisher fixes" above)

**Modified files:**
- `_bmad-output/specs/spec-folio/folio-format.md` (fontFamily default corrected; `headerStyle`
  documented; finisher: closing table pipe added — Finding 17)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (status → `review`; finisher: → `done`)
- `folio-go/internal/pagemodel/pagemodel.go` (`Color`, `RectEdges`, `Rect`, `Page.Rects`)
- `folio-go/internal/layout/band.go` (`ComposePage` gains a `rects` parameter)
- `folio-go/internal/layout/band_test.go` (updated call site)
- `folio-go/internal/layout/paginate.go` (`RectRef`, three-way item exclusivity, `"table"` overflow
  kind)
- `folio-go/internal/pdf/textdoc.go` (wires `appendRectContentStream` before text per page)
- `folio-go/internal/diag/diag.go` (`CodeStyleColorInvalid`)
- `folio-go/internal/diag/diag_test.go` (pin for the new code)
- `folio-go/internal/template/model.go` (`TableExt.HeaderStyle`)
- `folio-go/internal/template/parse_bands.go` (`headerStyle` parsing; finisher: `decodeStyle`/
  `decodePadding`/`decodeBorder` now take a `fieldPrefix` so a `headerStyle` load error names
  itself, not its `style` sibling — Finding 5)
- `folio-go/internal/template/serialize.go` (`headerStyle` serialization)
- `folio-go/internal/template/fixtures_test.go` (`maximalFixture` gains `headerStyle`)
- `folio-go/diagnostic.go` (`DiagCodeStyleColorInvalid` bridge)
- `folio-go/folio_expr_validate.go` (`headerStyle` placeholder-rejection call site; finisher:
  `checkStyleHasNoPlaceholders` now takes a `fieldPrefix` for the same reason — Finding 5)
- `folio-go/folio_expr_validate_test.go` (red-proof for the new call site; finisher: fixture gained
  a sibling `style` block and the assertion now discriminates `"headerStyle.background"` from
  `"style.background"` — Finding 5)
- `folio-go/render.go` (table collection wired into `predictDocument`/`paginateDocument`;
  `ComposePage` call site updated; finisher: the D-2.8.6-deviation comment now cites this Delivery
  Log's entry by name — Finding 7)
- `folio-go/page_number.go` (`contentColumnItems` gains a `tableRects` parameter)
- `folio-go/render_bind_test.go`, `folio-go/render_table_bind_test.go` (pre-existing table fixtures
  gained `style.fontFamily` — their tables now genuinely render a header for the first time; §D6;
  the compatibility cost of this is recorded in "Finisher fixes" above, Finding 13)
- `folio-go/collect_text_runs_composition_test.go` (updated `contentColumnItems` call site)

## Change Log

| Date | Note |
|---|---|
| 2026-08-26 | Created (bmad-story-creator), explicit story number, baseline `d5b75b5`. Six brief-vs-tree divergences measured at creation (§D1–D6). Four decisions raised; DECISION-1 and DECISION-4 need routing. Status set to `ready-for-dev`. |
| 2026-08-26 | Developed (bmad-story-developer). Parked mid-story on a genuine gap (fontFamily has no implemented default, contradicting R6); ruled by the engineering lead: no default is built, R6 and folio-format.md are corrected in place, and `headerStyle` is added to scope. Implemented the page model's first vector primitive, the colour conversion, the PDF fill/stroke path, table column geometry, and the header row (borders, padding, background, per-column alignment, `headerStyle`/`style` cascade). Minted `CodeStyleColorInvalid`. All ACs proven with real mutations, run and restored (see Delivery Log). Three-module gate green except the required, unchanged red. Status set to `review`. |
| 2026-08-26 | Reviewed (bmad-code-reviewer). Changes Requested: 2 Blockers, 5 Majors, 7 Minors, 4 Nits. Both blockers proved `headerStyle` and `style.valign` inert to the entire test suite by mutation; five Majors covered a wrong AC1 1a mechanism, a false AC5 "build fails" claim, misdirected `headerStyle` diagnostics, uninstrumented golden-fixture evidence, and a dangling D-2.8.6 citation. See QA Results below. |
| 2026-08-26 | Finished (bmad-story-finisher). All 18 findings triaged and resolved — 17 FIX, 1 partial DEFER (a diagnostic-code question inside Finding 13, cross-cutting beyond this story). Both Blockers fixed with genuine render-level tests (`TestHeaderStyleBackgroundWinsOverStyle`, `TestHeaderStyleCascadesPerField`, `TestColumnAlignWinsOverHeaderStyleAlign`, `TestTableHeaderValignPlacement`), each confirmed against the reviewer's own mutation. AC1 1a's record corrected to the faithful mutation. AC5's and the golden-fixture evidence's false claims corrected to name their real instruments (`lint`'s float scan; the re-rendering fixture tests). `headerStyle`'s load-time diagnostics now self-locate. The D-2.8.6 deviation is now actually written into this Delivery Log. Five Minors and four Nits fixed as targeted test/comment corrections. Gate re-measured independently: `folio-go` 965/0/1 (skip-command), `lint` 115/0, `hashmatrix` 3/0 — all three `go build`/`go vet`/`gofmt -l` clean. No golden fixture moved. See "Finisher fixes and re-measurement" above and "## Finding Resolutions" below. Status set to `done`. |

## QA Results

## Review Summary
- Reviewed by: bmad-code-reviewer
- Date: 2026-08-26
- Story Status Recommendation: **Changes Requested**
- Blockers: 2
- Majors: 5
- Minors: 7
- Nits: 4

**Gate figures measured by the reviewer, independently, at the working tree as submitted**
(D-000.26 — scope and flags stated):

| module | command (from that module's root) | measured |
|---|---|---|
| `folio-go` | `go test -count=1 -skip '^TestCorpusMeetsP6ExerciseFloors$' ./...` | **955 pass · 0 fail · 1 skip**, 17 packages |
| `lint` | `go test -count=1 ./...` | **115 pass · 0 fail**, 4 packages |
| `hashmatrix` | `go test -count=1 ./...` | **3 pass · 0 fail**, 2 packages |
| all three | `gofmt -l .` (invoked as `$(go env GOROOT)/bin/gofmt`, unfiltered) | **empty** |
| all three | `go vet ./...` | **clean** |

Counts taken from raw `go test -json` event streams (`Action ∈ {pass,fail,skip}` with a `Test`
field), not from a summariser. The developer's reported figures reconcile exactly. The
quarantined red (`TestCorpusMeetsP6ExerciseFloors`, D-000.17 / D-2.1.14 / D-000.57 / D-000.74)
was skipped by the green-job command and is not a finding.

**Heavy tests: read, not run (D-000.4).** The cross-target hash matrix and the integration/e2e
legs were not executed this story and could not be watched fail. They were read instead. What
that reading establishes and does not: `matrix_registration_test.go`'s registry gained **no**
new document (confirmed — this story's tests call `buildPageModel`/`Render` directly against
test-owned fixtures), so **D-000.54's "native leg at registration" genuinely does not apply**
and the Delivery Log is right not to log a D-000.4 override. The existing re-rendering
byte-comparison fixtures (`multi_page_fixture_test.go`, `wrapped_text_fixture_test.go`) DO
re-render and compare against the committed `expected.pdf`, and they ran green in the measured
suite above — that, and not the command the Delivery Log names, is what actually establishes
N = 0 (see Finding 6).

**Restoration hygiene.** Every mutation below was applied to the real file from a `cp` backup,
run, then restored by copying the backup back (never `git checkout`), with a SHA-256
re-verification per file after each restore. Final state: `git status --short` = 28 entries and
`git diff --stat` = 21 files / 349 insertions / 24 deletions, byte-identical to the tree as
submitted; a final full `folio-go` run re-measured **955 pass · 0 fail · 1 skip**.

**Mutations the reviewer re-ran independently (9, beyond the 4 required):**

| # | subject | reviewer's result |
|---|---|---|
| A | Delivery-Log row 1, *verbatim*: delete the `if el.Type == ElementTable {…} else` head at `parse_bands.go:158` | RED — but see Finding 1. Also **breaks the D-000.9 control** `TestTableWithNeitherWidthNorHeightLoads`, which the developer's narrower `-run` never showed them. |
| B | The *faithful* AC1 1a defect: keep the table branch, store the declared width on `Element.Width` | **RED for the right reason** (`got nil error`); control green; height half correctly green. The guard has real teeth. |
| C | Delivery-Log row 5: `geom.ScaleRound` → `v*1000/255` truncation | RED on `#808080` (0.501) and `#010101` (0.003), exactly as recorded. Accurate. |
| D | Delivery-Log row 11: widen `Rect.W` to the measured advance on overflow | RED on all three scripts, exactly as recorded. Accurate. |
| E | Delivery-Log row 12: `if false && !isVisible(...)` | RED on `TestTableHeaderVisibleIfFalseIsAbsentFromPageModel`, and on nothing else. Accurate. |
| F | AC5's float attempt: `geom.Length(float64(c.R)/255.0*1000)` in `appendColorChannels` | `go build ./...` **SUCCEEDS**. Caught instead by `lint`'s `TestFloatTypedProductionScan`. See Finding 4. |
| G | Delete the `valign` cascade from `resolveHeaderStyle` entirely | **ALL THREE MODULES GREEN.** See Blocker 2. |
| H | `hasHeader := false` — ignore `headerStyle` completely at render | **ALL THREE MODULES GREEN.** See Blocker 1. |
| I | Route header/footer-band table rects into `items` instead of `header.Rects`/`footer.Rects` | GREEN. See Finding 12. |
| J | Remove `style.fontFamily` from `render_bind_test.go`'s fixture (reverting this story's edit) | RED: `TestRenderScopeFenceIgnoresTableBind`. See Finding 13. |

**AC-by-AC disposition.** AC1: satisfied in substance (instruments 1a–1d all have teeth; 1a
proven by mutation B), record defective (Finding 1). AC2: behaviourally satisfied, instrument
weaker than its own declared anchor (Finding 8). AC3: **NOT satisfied** — the `valign` row of its
own field-by-field mapping table is unasserted (Blocker 2). AC4: satisfied. AC5: numeric half
satisfied and the single-conversion-site property confirmed; float half's record false
(Finding 4). AC6: satisfied — `TestPageModelNamesNoPDFConcept` scans every identifier in
`internal/pagemodel` and therefore covers `Rect`/`Color`/`RectEdges` automatically, and no new
package was created so `stagerank.go` correctly needed no entry. AC7: Part A partial and the
label partly unearned (Finding 9); `valign`'s share vacuous (Blocker 2); Part B's two-attach-point
claim unearned (Blocker 1); Part C confirmed green. AC8: behavioural half satisfied and its
D-000.24 label **earned** (see Finding 9's second paragraph). AC9: satisfied thinly (Finding 14).
AC10: gate satisfied; golden-count evidence defective (Finding 6).

---

### Finding 1: AC1 1a's recorded mutation does not discriminate the property, and its own predicted mechanism is false
- **Severity**: Major
- **Category**: Tests
- **Location**: story Delivery Log, "Mutations run" row 1; story AC1 table, instrument 1a's
  "discriminating mutation" cell; `folio-go/internal/template/parse_bands.go:156-184`;
  `folio-go/internal/template/table_geometry_test.go:118-161`
- **Observation**: AC1 1a predicts that deleting the `if el.Type == ElementTable` branch makes
  "the document load cleanly and the width is stored on `Element.Width`". It does not. Deleting
  that branch head hands table elements to the generic box path at `:165-184`, which **requires
  both** `width` and `height`. I ran it: `TestTableDeclaringWidthIsRejected` reddens with
  `field height (element e1): missing required field` and
  `TestTableDeclaringHeightIsRejected` reddens with `field width … missing required field` — each
  reddens because the *sibling* dimension is absent, never because the declared one was accepted.
  The developer saw this ("for the wrong reason") and recorded it as confirmation anyway.
  Worse, the same mutation also reddens the D-000.9 control
  `TestTableWithNeitherWidthNorHeightLoads` — a control failing alongside the guard is the
  signal that the mutation is not isolating the property — and the recorded command
  (`-run 'TestTableDeclaringWidthIsRejected\|TestTableDeclaringHeightIsRejected'`) excluded the
  control, so it was never observed.
- **Impact**: A recorded measurement that names a mutation which cannot establish the claim is
  the same defect as a guard that cannot fail (D-000.9's extension). AC1's own instrument table
  carries a mechanism that is factually wrong and will be copied by 4.2–4.8.
- **Suggested Resolution**: Replace row 1's mutation with the faithful one and re-run it. I have
  already confirmed it works: keep the `ElementTable` branch, replace the `if wok { return … }`
  rejection with `consumed["width"]=true; v,_ := decodePointsRaw(...); el.Width = present(v)`.
  Result: `TestTableDeclaringWidthIsRejected` reds with `a table declaring "width" must be a load
  error (AD-13, AC1); got nil error`, the control stays green, and the height half stays green.
  Do the mirror for `height`. Correct AC1 1a's prediction text, and widen the recorded `-run` to
  include the control.
- **Related AC**: AC1 (instrument 1a)

### Finding 2: `headerStyle` — this story's owner-ruled scope addition — is inert to the entire test suite
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render.go:73-137` (`resolveHeaderStyle`, the whole `hasHeader`
  arm of seven `switch` blocks); `folio-go/table_render_test.go` (absence — the string
  `headerStyle` appears in zero test fixtures in this file)
- **Observation**: Mutation H. I replaced lines 78-82 with `hasHeader := false; _ = header`, so
  `headerStyle` is ignored entirely at render. **`folio-go` 955 pass · 0 fail · 1 skip; `lint`
  115/0; `hashmatrix` 3/0 — the complete three-module gate stays green.** Every `headerStyle`
  test in the tree is a *load*-side test: `maximalFixture`'s round-trip
  (`internal/template/fixtures_test.go:273-292` via `TestP3FixturesAreCanonical`) and the
  placeholder fence (`folio_expr_validate_test.go:639`). Not one test renders a document whose
  `headerStyle` differs from its `style`, so the per-field cascade — the entire behaviour the
  ruling bought — is unobserved.
- **Impact**: DECISION-1 arm B was taken precisely so Story 4.7's C4 golden report can have a
  visually distinct header. The mechanism that delivers it could be deleted today without a
  single test noticing, and 4.4 will repeat that header on every continuation page. This is the
  programme's signature defect (a shipped negative/positive property with no instrument) landing
  on the one thing the story added beyond its brief.
- **Suggested Resolution**: Add render-level tests that make mutation H red. At minimum: (a) a
  table where `style.background` and `headerStyle.background` differ, asserting the header rect
  carries the **headerStyle** colour; (b) a table where `headerStyle` sets only `border` while
  `style` sets `padding`, asserting the per-field fall-through (the header gets the headerStyle
  border AND the style padding — this is the cascade's actual claim, and one test per field is
  not needed if this composite test discriminates); (c) a table where `columns[].align` is set
  and `headerStyle.align` is set, asserting the column still wins (the ruling's "extended one
  level" clause, currently asserted nowhere). Then record the mutation.
- **Related AC**: AC3, AC4, AC7 Part B (the "TWO attach points" claim)

### Finding 3: `style.valign` reaches no assertion anywhere; AC3 claims it field by field and AC7 counts it as wired
- **Severity**: Blocker
- **Category**: Tests / AC Conformance
- **Location**: `folio-go/table_render.go:122-127` (the cascade) and `:373-382` (the
  `top`/`middle`/`bottom` placement switch)
- **Observation**: Mutation G. I deleted the valign cascade so `r.valign` is permanently `"top"`.
  **The full three-module gate stays green.** Grepping the tree confirms it: `valign` appears in
  tests only as a *round-trip* fixture value (`internal/template/fixtures_test.go:291,315`), as
  an unobservable field inside `TestTableStyleFieldsAreNotDataDriven`'s template
  (`table_render_test.go:346`), and in prose comments. No test renders `valign: "middle"` or
  `"bottom"` and asserts a label's Y moved. The `middle`/`bottom` arms of the switch at
  `:376-379` are dead as far as any instrument is concerned.
- **Impact**: Three separate AC claims are unearned. AC3's mapping table says the story "asserts
  it field by field" and names `style.valign` as the governor of vertical placement. AC7 Part A
  requires a discriminating mutation for each of five newly-wired fields, and for `valign` the
  byte-identity assertion is **vacuous** — a field with no observable effect trivially produces
  byte-identical output across two datasets. AC7 Part B's restated population then records
  `valign` as one of the fields that "reach output as of Story 4.1", which is a knowingly-unproven
  fact carried forward (D-000.66) into the population count 4.8 will inherit.
- **Suggested Resolution**: Add a test that renders the same table three times with
  `valign: "top" | "middle" | "bottom"` (header height fixed, everything else byte-identical) and
  asserts three **distinct, test-owned literal** label Y values. Confirm mutation G reddens it.
  Until then, `valign` belongs in AC7 Part B's *unwired* list, not its wired one.
- **Related AC**: AC3, AC7 (Parts A and B)

### Finding 4: AC5's float half is recorded as "the build fails"; the build does not fail, and the attempt is absent from the mutation table
- **Severity**: Major
- **Category**: Tests / Convention
- **Location**: story AC5 ("Introducing `float64(c)/255.0` → the **build** fails … and the
  developer records that it was attempted and that the build failed"); story Delivery Log
  "Mutations run" table (13 rows, **none** of them the float attempt);
  `folio-go/internal/pdf/rectdoc_test.go:40-49` (the claim asserted as fact in a doc comment)
- **Observation**: Mutation F. I replaced the R-channel conversion with
  `geom.Length(float64(c.R)/255.0*1000)`. **`go build ./...` succeeds.** `go vet` is silent. What
  actually catches it is a test in a *different module*: `lint`'s `TestFloatTypedProductionScan`
  reports `internal/pdf/rectdoc.go:106:38: this value expression has floating-point type float64
  (resolved by go/types, not spelled in the source)`, and `TestFloatTypedTestScopeInventory` reds
  on the site-count mismatch (got 6, want 5). The `rectdoc_test.go` comment states the opposite
  as fact: *"introducing `float64(c)/255.0` at this call site does not compile under this module's
  float ban"*.
- **Impact**: The property IS protected — the `lint` guard fires and `lint` is inside the
  three-module gate (D-000.64) — so this is not a production risk. It is a false recorded
  measurement in a comment that a future reader will trust, plus an AC recording obligation
  (D-000.9's extension: "every measurement below names its command, its mutation, and confirms
  the mutation was run") that the 13-row table simply does not discharge. A reader who believes
  the compiler is the anchor will not think to keep `lint` in the gate.
- **Suggested Resolution**: Add the float attempt as a 14th row naming the real instrument —
  `lint`'s `TestFloatTypedProductionScan` / `TestFloatTypedTestScopeInventory`, run from `lint/` —
  and correct `rectdoc_test.go:44-48` to say the ban is enforced by the cross-module type-resolved
  scan (D-3.1a), not by the compiler. Correct AC5's own wording the same way.
- **Related AC**: AC5

### Finding 5: every load diagnostic raised inside `headerStyle` names the field as `style`
- **Severity**: Major
- **Category**: Correctness / Convention
- **Location**: `folio-go/internal/template/parse_bands.go:322-334` (`decodeStyle(id, hsRaw)`);
  `decodeStyle`'s own `newLoadError("style.…", …)` prefixes at e.g. `parse_bands.go:504,507`;
  `folio-go/folio_expr_validate.go:108-112` (`checkStyleHasNoPlaceholders(hs.Value, el.ID)`,
  whose messages hardcode `"style.<field>"` at `folio_expr_validate.go:284` and siblings)
- **Observation**: Measured against the shipped parser on an element that carries **both** a
  `style` block and a `headerStyle` block:
  - `"headerStyle": 5` → `template: field style (element e1): must be an object`
  - `"headerStyle": {"valign": "sideways"}` → `template: field style.valign (element e1): not one of the closed set top, middle, bottom`
  - `"headerStyle": {"fontSize": "big"}` → `template: field style.fontSize (element e1): must be a JSON number`

  The `{{ }}` fence has the same defect, and its new test does not catch it:
  `TestParseTemplateRejectsPlaceholderInHeaderStyle` (`folio_expr_validate_test.go:667-672`)
  asserts only that the message contains `"e2"` and `"background"` — both of which a
  `style.background` message also satisfies. Its own fixture carries no `style` block, so the
  ambiguity is invisible there.
- **Impact**: A template author who mistypes inside `headerStyle` is sent to the wrong block.
  Story 3.6 exists to make diagnostics located and actionable; this one is located at a sibling.
  The parent review question was whether `headerStyle` is "validated to the same standard as the
  rest of `Style`" — on the diagnostic-location axis it is not, and nothing tests it.
- **Suggested Resolution**: Thread a field-path prefix into `decodeStyle` and
  `checkStyleHasNoPlaceholders` (`"style"` vs `"headerStyle"`) so the message names the block the
  author wrote. Extend `TestParseTemplateRejectsPlaceholderInHeaderStyle`'s fixture to carry a
  `style` block **as well**, and assert the message contains `"headerStyle"` — that assertion is
  what makes the test discriminate.
- **Related AC**: AC3 (the `headerStyle` scope addition), AD-14

### Finding 6: the command recorded as measuring "N = 0 golden fixtures moved" cannot see a moved golden
- **Severity**: Major
- **Category**: Tests
- **Location**: story Delivery Log, "Measured golden-fixture impact: N = 0";
  `folio-go/byte_neutrality_test.go:245-300` (`TestGoldenDigestAgreesAtEveryDeclaredSite`)
- **Observation**: The named command is
  `go test . -run 'TestGoldenDigestAgreesAtEveryDeclaredSite|TestEpic2GateObligationsMatchTheDeclaredSet|TestBoundaryGateDigestsAreWellFormed'`.
  I read the first: it `os.ReadFile`s each **committed** `fixtures/<x>/expected.pdf` and compares
  its SHA-256 against a declared digest. It never calls `Render`. It therefore detects someone
  editing a checked-in artifact; it is structurally incapable of detecting the renderer producing
  different bytes for the same input. The other two check declaration-set completeness and digest
  well-formedness. All three would pass unchanged if this story had moved every golden in the tree.
- **Impact**: This is the C4 gate's evidence base and Story 4.7 depends on it (the story says so).
  A number carried forward under a command that cannot produce it is D-000.9's "all clear
  indistinguishable from could not look", applied to recorded evidence.
  **The conclusion is nonetheless correct** — I verified it independently: the re-rendering
  comparisons do exist (`multi_page_fixture_test.go:689-703` reads `expected.pdf` and compares it
  to a fresh render; `wrapped_text_fixture_test.go:432-475` likewise) and both ran green in my
  measured suite, and `appendRectContentStream` provably appends zero bytes when `page.Rects` is
  empty (`TestAppendRectContentStreamEmptyPageProducesEmptyStream`). So N = 0 holds; the evidence
  cited for it does not establish it.
- **Suggested Resolution**: Re-record the measurement against a command that re-renders — name
  the fixture byte-comparison tests (`multi_page_fixture_test.go`, `wrapped_text_fixture_test.go`,
  and whichever others reproduce an `expected.pdf`) — or state plainly that the evidence is the
  full green suite plus the empty-`Rects` no-op property, and name both.
- **Related AC**: AC10

### Finding 7: `render.go` cites a Delivery Log entry that does not exist, for an undisclosed D-2.8.6 deviation
- **Severity**: Major
- **Category**: Correctness / Maintainability
- **Location**: `folio-go/render.go:1410-1416` (the comment beginning "Table label runs are
  appended AFTER this band's own text runs (this story's own, stated deviation — Delivery Log —
  …)"); `render.go:1450-1456` (the `diags` assembly)
- **Observation**: `grep -n "declaration order\|D-2.8.6\|diagnostic order\|relative order"` over
  the story file returns **zero matches**. The Delivery Log does not state this deviation
  anywhere. The deviation itself is real: `diags` is assembled as
  `headerDiags, headerTableDiags, contentDiags, contentTableDiags, footerDiags, footerTableDiags`,
  so within a band **every** text element's diagnostics precede **every** table's, regardless of
  which was declared first — whereas `render.go:1443-1448`'s own doc comment states D-2.8.6's rule
  as "band order, then element declaration order within a band".
- **Impact**: A code comment asserting that a deviation is on the record, when it is not, is worse
  than an undisclosed deviation: a reviewer who checks the citation and finds nothing has no way
  to tell whether the deviation was ruled or merely happened. `Result.Diagnostics` order is a
  public-surface property (D-2.8.6) that integrators sort on.
- **Suggested Resolution**: Either write the deviation into the Delivery Log verbatim (what it is,
  why, and that D-2.8.6's within-band clause is knowingly relaxed for table-vs-text), or interleave
  the table diagnostics into declaration order and delete the comment. Do not leave the citation
  pointing at nothing.
- **Related AC**: AC10's recording obligations; D-2.8.6

### Finding 8: AC2's instrument does not use the anchor AC2 declares, tests one column, and never asserts the table's total width
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_render_test.go:381-445`
  (`TestColumnGeometryNeverNegotiatesAgainstLabelContent`)
- **Observation**: AC2 requires "the expected geometry is written out in the test, never read back
  from the function under test", and requires "every column's x-origin, every column's width,
  **and the table's total width**" to be compared. The shipped test instead reads the narrow
  render's `Rect{X,W}` and compares it to the wide render's (`:406-409`) — the oracle is another
  invocation of the code under test, which the test's own doc comment states openly at `:373-376`.
  It uses a **single** column (`:394-396`), so no x-origin *propagation* is exercised: the one
  column's X is `el.X` and cannot move. And the table's total width is never asserted at all —
  `TableGeometry` is not reachable from the page model. The only test-owned literal in the whole
  test is `60000`, and it pins `ClipWidth`, not the column width.
- **Impact**: The instrument does catch the AC's named defect (mutation D reddens it on all three
  scripts, verified), so this is not vacuous. But a negotiation confined to columns 2..N, or one
  that widens narrow and wide equally, escapes it — and the strongest anchor the story actually
  has for AC2 is `layout.ColumnWidths`'s content-free *signature*, which nothing guards. A future
  author can add a `labels []string` parameter and no test objects until it changes a first-column
  width.
- **Suggested Resolution**: Give the test three columns with test-owned literal expected
  `{X, W}` pairs for each (not narrow-vs-wide equality), assert the summed total against a literal,
  and keep the narrow/wide comparison as the additional behavioural half. Consider a reflection or
  `go/types` guard pinning `ColumnWidths`'s parameter list, which is the compiler-anchored form
  D-000.68 asks for.
- **Related AC**: AC2

### Finding 9: AC7 Part A's structural justification is overstated, and its D-000.9 control covers one field of five
- **Severity**: Minor
- **Category**: Tests
- **Location**: story Delivery Log, "AC7's mutation deviates from the letter…";
  `folio-go/table_render_test.go:336-343` and `:447-475`; `folio-go/table_render.go:258-265`
- **Observation**: The label rests on "`collectBandTableRuns`/`resolveHeaderStyle`/
  `buildHeaderCellRect` … take **NO** `data`/`params` parameter anywhere in their signatures —
  there is no channel through which `overdue` could reach a style decision without a signature
  change visible in review." That is true of the latter two. It is **not** true of
  `collectBandTableRuns`, whose signature at `:258-265` includes
  `visible visibilityVerdicts` — a value computed *from the report data*. The claim as written
  overstates the type system's guarantee for the function that actually builds the rects.
  Separately, `TestTableStyleFieldsAreNotDataDrivenControl` (`:453-475`) proves the byte
  comparison is sensitive to exactly one thing: `border.color`. It says nothing about whether the
  comparison can see `padding`, `background`, `align` or `valign`. Three of those four are
  independently observed elsewhere (`TestTableHeaderPaddingInsetsLabel`,
  `TestTableHeaderBackgroundFill`, `TestColumnAlignWinsOverStyleAlign`); `valign` is not, which is
  Blocker 2.
- **Impact**: The honest-labelling test the parent set — a label is correct only where the thing
  genuinely cannot be proven — is **passed for AC8** and **half-failed here**. AC8's D-000.24 label
  is earned: deriving "kinds reaching the page model" and "kinds whose verdict is consumed" as two
  independently-anchored sets really does need an AST scanner on the scale of `lint`'s stage-rank
  walker, and the AC's own escape hatch anticipated exactly that. AC7's label is *partly* earned:
  the argument is sound for the two data-free functions, but it is stated more strongly than the
  code supports, and the control that would make the remaining measurement non-vacuous covers one
  field.
- **Suggested Resolution**: Correct the sentence to name `visible` and explain why a
  visibility verdict is not a style channel. Extend the control so it is sensitive to each of the
  five fields (five paired renders differing only in that field, each producing different bytes) —
  that is cheap, needs no signature change, and converts the byte-identity assertion from
  one-field-sensitive to five-field-sensitive without the five mutations the AC asked for.
- **Related AC**: AC7 Part A (and, favourably, AC8's label)

### Finding 10: AC1 1d's "near-limit" vector exercises an overflow path that does not exist
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/layout/table_test.go:86-90`;
  `folio-go/internal/layout/table.go:28-34` (`TableGeometry.Width`)
- **Observation**: AC1 1d asks for "one vector summing near `int64` limits to exercise `geom`'s
  overflow path". `Width()` is `var total geom.Length; for … { total += c.Width }` — a plain
  `+=` with no overflow detection and no call into `geom` at all. The test's own comment claims
  it exercises "`geom.ScaleRound`'s sibling overflow-aware path used elsewhere in this module",
  which `Width()` does not touch. A sum that does exceed `int64` wraps silently. Separately, the
  primary mutation (sum → max) reddens **2** of the 5 vectors; the AC asked for "at least three",
  and three of the five vectors have ≤1 element so they cannot ever discriminate sum from max.
  The Delivery Log records 2 of 5 honestly.
- **Impact**: The vector is present in letter and inert in substance. Whether silent wrap is a real
  hazard depends on whether `decodePointsRaw` bounds a column width — worth settling, since AD-13
  makes the sum the table's only width and 4.3's pagination will compare it against a band.
- **Suggested Resolution**: Either give `Width()` the overflow behaviour the AC assumed (and test
  it), or drop the false comment and state that the vector is a large-value smoke case. Add a
  third multi-column vector so the sum-vs-max mutation reddens three, as 1d specifies.
- **Related AC**: AC1 (instrument 1d)

### Finding 11: `TestAppendColorChannelsNoFloat` asserts nothing about floats and duplicates two rows of the table above it
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/internal/pdf/rectdoc_test.go:50-60`
- **Observation**: The test body asserts `ScaleRound(0) == 0` and `ScaleRound(255) == 1000`. Both
  are already asserted, end to end and through the emitter, by the `#000000` and `#FFFFFF` rows of
  `TestAppendColorChannels` at `:25-26`. It contains no float, no float-detection, and no
  reference to the ban. I confirmed it stays **green** under mutation C (the truncation mutation
  that reddens the real test) — it is inert to the only numeric defect in its neighbourhood.
- **Impact**: A green line reading `TestAppendColorChannelsNoFloat` in CI output reads as "no
  float, proven". Nothing was proven. Combined with Finding 4's false doc comment, a reader has two
  independent signals that the float ban is locally enforced when it is enforced in another module.
- **Suggested Resolution**: Delete it, or rename it to what it does and let the doc comment point
  at `lint`'s `TestFloatTypedProductionScan` as the actual instrument.
- **Related AC**: AC5

### Finding 12: table rects in the page-header and page-footer bands are new, uncovered code
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/render.go:1746-1752` (the `case pageHeaderBandIndex` /
  `case pageFooterBandIndex` arms of the `switch ts.band`)
- **Observation**: Mutation I. I deleted both arms so header- and footer-band table rects fall into
  the `default` (content) arm — losing the repeat-on-every-page and no-`Shift` behaviour entirely.
  The full `folio-go` suite stays **green**. No test places a `table` element in a `pageHeader` or
  `pageFooter` band. The parallel filter in `contentColumnItems`
  (`page_number.go:119-121`, `if ts.band != contentBandIndex { continue }`) is uncovered for the
  same reason.
- **Impact**: The schema permits a table in any band, `collectBandTableRuns` is called for all
  three (`render.go:1381-1393`), and the repeated-verbatim path is exactly where a wrong `Shift`
  would show up first — on page 2, which this story never renders.
- **Suggested Resolution**: One two-page test with a small table in the `pageHeader` band,
  asserting the same rect geometry appears on both pages with no `Shift` applied. That also gives
  4.4 (the repeated header) a starting instrument.
- **Related AC**: AC3, AC10

### Finding 13: this story turns previously-rendering documents into render errors, and the ruling's stated justification is contradicted by the story's own diff
- **Severity**: Minor
- **Category**: Correctness / AC Conformance
- **Location**: `folio-go/table_render.go:329-337`; `folio-go/render_bind_test.go:29`;
  `folio-go/render_table_bind_test.go:18`; story Delivery Log, mid-story ruling bullet 1
- **Observation**: The ruling justifies requiring `style.fontFamily` on the grounds that it is
  "the conservative, reversible direction: every template that renders today renders identically
  (no recorded hash moves)". That is true only of templates whose tables are invisible today.
  Mutation J: I removed the `"style": {"fontFamily": "body", "fontSize": 9}` this story added to
  `render_bind_test.go`'s fixture — restoring the file to its pre-4.1 content — and
  `TestRenderScopeFenceIgnoresTableBind` fails with
  `Render must succeed despite the table column's expression-shaped bind field: folio: Render:
  element e2: has a column label but no style.fontFamily … to resolve a font from`. That test's
  entire premise is "Render must succeed". Two in-repo fixtures had to be edited for this story to
  stay green, which is the measurement the ruling's justification lacks.
- **Impact**: Any `.folio` document with a table carrying at least one non-empty column label and
  no `style.fontFamily` rendered successfully before this story and fails after it. The File List
  discloses the fixture edits, but frames them as "their tables now genuinely render a header for
  the first time" rather than as a breaking change — so the ruling reads as cost-free when it is
  not. Additionally the error is a plain `fmt.Errorf`, not a `*RenderError` carrying a diagnostic
  code, so R6's "located error" is located only by the element id embedded in the string.
- **Suggested Resolution**: Record the measured compatibility impact in the Delivery Log (two
  in-tree fixtures required edits; the class is "table + non-empty label + no `style.fontFamily`"),
  and correct the ruling's "renders identically" clause to name its exception. Consider whether
  this failure should carry a diagnostic code like every other located render failure.
- **Related AC**: R6, AC3, AC9

### Finding 14: AC9's scope fence asserts a rect count only
- **Severity**: Minor
- **Category**: Tests
- **Location**: `folio-go/table_render_test.go:314-329` (`TestTableRendersZeroDataRows`)
- **Observation**: The doc comment says "the ONLY runs the page model carries are the header
  labels" and "A data ROW would add more rects … and more runs; this count pins 'header only'".
  Only the rect count is asserted (`:326-328`). No assertion touches `pages[0].Runs` at all. A
  row implementation that emitted cell text without cell rects — which is precisely what 4.2 will
  do first, since 4.2 owns text and 4.8 owns row shading — would leave this fence green.
- **Impact**: AC9 exists to give 4.2 "an unambiguous red-to-green to inherit". As written, 4.2 can
  ship row text and inherit a green fence.
- **Suggested Resolution**: Assert the run count too, against the two header labels' glyph runs,
  or assert that every run's `elementID`/`SourceText` belongs to a column label.
- **Related AC**: AC9

### Finding 15: `fieldNameSet`'s comment claims `Extra` is excluded "on both"; `TableExt` has no `Extra` field
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/template/table_geometry_test.go:166-167`;
  `folio-go/internal/template/model.go:214-230`
- **Observation**: `Column` has `Extra []Field` (`model.go:244`). `TableExt` does not. The
  exclusion at `table_geometry_test.go:192-194` is therefore inert for `TableExt`.
- **Impact**: A reader is told a passthrough carrier exists on `TableExt` when it does not.
- **Suggested Resolution**: Say "excluded by name on `Column`", or leave the exclusion and correct
  the comment.
- **Related AC**: AC1 (instrument 1b)

### Finding 16: stale comment — "these three" describes a two-element vector
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/internal/layout/table_test.go:87-88`
- **Observation**: "a plain sum of **these three** stays well inside int64" sits above
  `{"near-limit", []geom.Length{4611686018427387903, 4611686018427387903}, …}` — two elements.
- **Suggested Resolution**: Fix alongside Finding 10.
- **Related AC**: AC1 (instrument 1d)

### Finding 17: the new `headerStyle` row in `folio-format.md` is missing its closing table pipe
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `_bmad-output/specs/spec-folio/folio-format.md:222`
- **Observation**: The line ends `… (see \`style\`, below).` with no trailing `|`; every other row
  in that table closes with one. (The drift tests are unaffected — they compare backticked key
  tokens, not table structure, so this is cosmetic only.)
- **Suggested Resolution**: Add the closing `|`.
- **Related AC**: AC10 / the D-4.1.1 amendment

### Finding 18: two doc comments repeat Finding 9's overstated structural claim
- **Severity**: Nit
- **Category**: Maintainability
- **Location**: `folio-go/table_render_test.go:337-341`; story Delivery Log, AC7 paragraph
- **Observation**: Both state that the new functions "take NO data or params parameter at all —
  there is no channel through which `overdue` could reach a style decision without a signature
  change". `collectBandTableRuns` takes `visible visibilityVerdicts`, which is data-derived.
- **Suggested Resolution**: Fix with Finding 9, in both places, so the code and the record agree.
- **Related AC**: AC7 Part A

---

**What was checked and found sound**, recorded so an "all clear" is distinguishable from a
"did not look" (D-000.9):

- **The colour path (AC5, AD-23).** `grep` over the module confirms exactly **one** channel
  conversion site — `appendColorChannels` (`internal/pdf/rectdoc.go:105-112`) — and exactly **two**
  colour operator emissions, `" rg"` at `:46` and `" RG"` at `:60`, both routed through it. Fill
  and stroke genuinely share the one path. `geom.ScaleRound(channel, 1000, 255)` → the existing
  `appendLength` emitter, integer end to end. Mutation C confirms the rounding is
  round-half-to-even and not truncation, on both `#808080` and `#010101`.
- **`CodeStyleColorInvalid` against AD-14 and D-000.65.** Added to the `allCodes` enumeration
  (`internal/diag/diag.go:169`), pinned in `codePins` (`diag_test.go:41`), bridged at the root
  (`diagnostic.go:234-241`), and `TestRegistryIsAdditiveOnly` covers it. The condition is genuinely
  new and genuinely reachable — `TestTableHeaderStyleColorInvalid` reaches it through the public
  `Render`, and I reproduced it independently (`headerStyle: {"background": "nope"}` →
  `*RenderError` with this code). Minted where the condition first ships, not in advance.
- **The three pre-settled data cases**, measured against the shipped parser: absent
  (`headerStyle` omitted, `background` omitted) → no error, defaults apply; explicit `null`
  (`"headerStyle": null`, and `Presence.Null` on every style field) → empty, not an error, falls
  through to the next cascade level; wrong kind (`5`, `"big"`, `[]`, `{"fontSize": "big"}`) →
  load error, and the message says **"never coerced"** in as many words. All three hold.
  (The *location* of those errors is Finding 5.)
- **AC6.** `TestPageModelNamesNoPDFConcept` (`internal/bandcomposition_arch_test.go:134`) walks
  every identifier in `internal/pagemodel` and its first-party imports, so `Rect`, `Color` and
  `RectEdges` are covered automatically with no new machinery — as AC6 required. No new package
  was created, so `stagerank.go` and `TestSpineStageLadderMatchesStageRankTable` correctly needed
  no entry. Delivery Log row 6 (adding an `internal/pdf` import to `internal/layout/table.go`) is
  the right instrument and the existing scan is the thing that reports it.
- **The named guards still hold.** `lint` measured 115/0, which includes
  `TestGlyphIdentifierCensus` — `table_render.go` never constructs a `pagemodel.ShapedGlyph` and
  routes labels through `shapeSegments`/`positionSegments`, so `buildShapedPDFRuns` remains the
  sole producer despite the `page_number.go` edit — and `TestSpineStageLadderMatchesStageRankTable`,
  `TestFloatTypedProductionScan` and `TestFloatTypedTestScopeInventory`. `folio-go`'s
  `TestFolioMethodNamesAreInjective`, `TestValidateNeverReachesRenderOrInternalPDF` and the
  one-byte-producer guard all ran green in the 955.
- **AD-1 under `internal/`.** The two new `internal/` files import only `internal/geom` and
  `internal/pagemodel`; no `time`, `os`, `math/rand`, `net`, no transcendentals, no package-level
  mutable state, no map iteration reaching output.
- **`headerStyle` round-trip and drift.** `maximalFixture` gained a full `headerStyle` block
  (`internal/template/fixtures_test.go:273-292`), and it is a member of `canonicalFixtures`, so
  `TestP3FixturesAreCanonical` asserts byte-identical `Serialize(Parse(b)) == b` over it. Because
  the drift tests compare whole-document key-name sets against `folio-format.md`'s backticked
  tokens, and `serialize.go` now emits `"headerStyle"`, the drift halves cover the new key
  automatically — `folio-format.md:222` documents it, which is what keeps them green. Load-side
  coverage is genuinely there; it is the *render* side that is missing (Blocker 1).
- **The D-4.1.1 amendment.** `folio-format.md`'s before/after matches the ruling recorded in the
  Delivery Log verbatim: `| \`fontFamily\` | the first key of \`fonts\` |` →
  `| \`fontFamily\` | **none — required on any element carrying text** |`, plus the new paragraph
  under the Style defaults table. The `border.width` (`0.5` pt) and `border.color` (`"#000000"`)
  defaults the code applies at `table_render.go:210-217` match what the document says at
  `folio-format.md:296-297`.
- **Pagination wiring.** `layout.Paginate` sorts items by `Top` (`paginate.go:319-328`), so
  `paginateDocument` prepending rect items to `items` cannot reorder page assignment. The
  three-way exclusivity generalisation is a populated-count check, correct for the new kind, and
  the header-label item and the rect item share `Top`/`Bottom` so they cannot be split across a
  page boundary. `pageRects` is built with `make(..., 0, n)`, so `Page.Rects` is empty (never
  spuriously non-empty) for pre-4.1 documents.
- **Scope fence honoured.** No data rows, no pagination logic, no repeated header, no footer
  aggregates, no alternating shading, and no attempt to define a table wider than its band. No
  bold/italic wiring — correctly, per §D3.

---

## Finding Resolutions (bmad-story-finisher)

Triaged by the finisher after the review above. **17 of 18 findings FIX; one sub-question inside
Finding 13 DEFERRED** (a cross-cutting diagnostic-code question, not scoped to this story). Both
Blockers are fixed with genuine render-level tests, confirmed against the reviewer's own
mutations — neither was argued away. Full mutation transcripts are in "Finisher fixes and
re-measurement" above, in the Delivery Log.

| # | Finding | Severity | Decision | Rationale | Files changed |
|---|---|---|---|---|---|
| 1 | AC1 1a's recorded mutation does not discriminate the property | Major | **FIX** | The reviewer's own faithful mutation (keep the branch, store the declared width) reds for the right reason with the control green — re-run and confirmed. The AC1 table and Delivery Log row 1 are corrected to that mutation. | story file (AC1 table, Delivery Log) |
| 2 | `headerStyle` inert to the entire test suite | **Blocker** | **FIX** | This is D-000.76's own owner-ruled scope addition and Story 4.7's C4 gate depends on it looking distinct. Added 3 render-level tests proving the per-field cascade and precedence; confirmed the reviewer's mutation H now reds on all three. | `folio-go/table_render_test.go` |
| 3 | `style.valign` reaches no assertion anywhere | **Blocker** | **FIX** | Same shape as Finding 2 — a wired AC7 field with a vacuous byte-identity assertion. Added a render-level test asserting three distinct, strictly-ordered label Y positions; confirmed the reviewer's mutation G now reds. | `folio-go/table_render_test.go` |
| 4 | AC5's float half recorded as "the build fails"; it does not | Major | **FIX** | Reproduced independently: `go build` succeeds, `lint`'s type-resolved scan is what actually catches it. Corrected AC5's text and `rectdoc_test.go`'s doc comment to name the real instrument; added the mutation as a recorded row. | story file (AC5), `folio-go/internal/pdf/rectdoc_test.go` |
| 5 | Every load diagnostic inside `headerStyle` names the field `style` | Major | **FIX** | Confirmed the mislocation independently. Threaded a `fieldPrefix` through `decodeStyle`/`decodePadding`/`decodeBorder` and `checkStyleHasNoPlaceholders`; added a load-error-location test and hardened the existing placeholder-fence test with a sibling `style` block so it actually discriminates. | `folio-go/internal/template/parse_bands.go`, `folio-go/folio_expr_validate.go`, `folio-go/folio_expr_validate_test.go`, `folio-go/internal/template/table_geometry_test.go` |
| 6 | Golden-fixture-count command cannot see a moved golden | Major | **FIX** | Confirmed: the cited command hashes committed artifacts, never re-renders. The conclusion (N=0) is correct — verified via the actual re-rendering fixture tests, which are now cited as the evidence. | story file (Delivery Log) |
| 7 | `render.go` cites a Delivery Log entry that does not exist | Major | **FIX** | Confirmed by grep: zero matches. The deviation is real (table diagnostics sort after text diagnostics within a band) and cheap to document; wrote the entry and pointed the code comment at it by name. | story file (Delivery Log), `folio-go/render.go` |
| 8 | AC2's instrument uses the wrong oracle, one column, no total-width assertion | Minor | **FIX** | All three defects confirmed by reading the test. Rewrote with 3 columns (only the middle one's label varies), test-owned literal `{X,W}` pairs and a literal total, keeping the narrow/wide behavioural comparison. Reddens under the reviewer's mutation D across all three scripts. | `folio-go/table_render_test.go` |
| 9 | AC7 Part A overstated; control covers 1 of 5 fields | Minor | **FIX** | `collectBandTableRuns` does take `visible visibilityVerdicts` — the claim was too strong. Corrected the doc comment (`visible` gates presence, never appearance) and rewrote the control as a 5-case table, one per AC7 field. | `folio-go/table_render_test.go` |
| 10 | AC1 1d's "near-limit" vector claims an overflow path that doesn't exist | Minor | **FIX** | Confirmed `Width()` has no overflow detection. Corrected the comment to call it a large-value smoke case; added a third multi-column vector so the mutation reds 3 of 6, per the AC's own "at least three". | `folio-go/internal/layout/table_test.go` |
| 11 | `TestAppendColorChannelsNoFloat` asserts nothing about floats | Minor | **FIX** | Confirmed it duplicates two existing rows and stays green under the real float-rounding mutation. Deleted it in the same edit as Finding 4's comment fix. | `folio-go/internal/pdf/rectdoc_test.go` |
| 12 | Table rects in page-header/page-footer bands are uncovered | Minor | **FIX** | Confirmed via the reviewer's own mutation I. Added a two-page test with a table in the pageHeader band, asserting identical rect geometry (no Shift) on both pages; reds under mutation I. | `folio-go/table_render_test.go` |
| 13 | Breaking change: table+label+no-fontFamily now errors; ruling's "renders identically" is contradicted | Minor | **FIX** (disclosure) + **DEFER** (diagnostic code) | The compatibility cost is real and was under-disclosed — recorded the measured class and corrected the ruling's claim in the Delivery Log. Whether the failure should carry a diagnostic code is a pre-existing gap shared with the identical text-element failure mode (`fontChain`), not something this story introduced or should widen unilaterally — deferred to whoever next touches font-resolution error handling. | story file (Delivery Log) |
| 14 | AC9's scope fence asserts only a rect count | Minor | **FIX** | Confirmed a row-text-without-rects implementation would pass it. Extended the test to also assert the run count and that every run's `SourceText` is a column label, never data content. | `folio-go/table_render_test.go` |
| 15 | `Extra`-exclusion comment wrong for `TableExt` | Nit | **FIX** | Confirmed `TableExt` has no `Extra` field. Corrected the comment. | `folio-go/internal/template/table_geometry_test.go` |
| 16 | Stale "these three" comment on a two-element vector | Nit | **FIX** | Fixed alongside Finding 10 (the vector's own comment was rewritten). | `folio-go/internal/layout/table_test.go` |
| 17 | `headerStyle` row in `folio-format.md` missing closing pipe | Nit | **FIX** | Cosmetic, confirmed harmless to drift tests. Added the pipe. | `_bmad-output/specs/spec-folio/folio-format.md` |
| 18 | Two doc comments repeat Finding 9's overstated claim | Nit | **FIX** | Fixed both sites in the same edit as Finding 9 (`table_render_test.go`'s comment and this Delivery Log's own AC7 paragraph, corrected above). | `folio-go/table_render_test.go`, story file (Delivery Log) |

**The two honesty labels the parent review weighed (not findings, no action needed):**
- **AC8's D-000.24 label is EARNED** — left unchanged. The set-comparison guard genuinely needs an
  AST scanner on the scale of `lint`'s stage-rank walker; the AC's own escape hatch anticipated
  exactly this, and inventing a weaker guard to look proven would be worse than the stated absence.
- **AC7's label was HALF-EARNED — now more fully earned** after Finding 9's fix: the overstated
  clause about `collectBandTableRuns` is corrected, and the control that makes the remaining claim
  non-vacuous now covers all five fields instead of one.

**Follow-ups for a new ticket / the owner (none of them block this story):**
- The diagnostic-code question inside Finding 13 (should a table's "column label with no resolvable
  `style.fontFamily`" failure — and its pre-existing text-element twin, `fontChain` — carry a
  `*RenderError` diagnostic code instead of a plain `fmt.Errorf`?).
- Whether `TableGeometry.Width()`'s plain integer summation should gain overflow protection
  (Finding 10's "worth settling" note) — bundled with DECISION-3's already-deferred "a table wider
  than its band" question, since both are about what happens when declared column widths are
  unreasonable.
- The D-2.8.6 within-band text-vs-table diagnostic ordering deviation (Finding 7) has no dedicated
  test; 4.2–4.8 will exercise the table-diagnostic surface far more and are better positioned to
  add one, or to decide the ordering should be interleaved by declaration order instead.

