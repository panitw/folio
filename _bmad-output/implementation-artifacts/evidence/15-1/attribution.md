# Story 15.1 — attribution evidence for the moved statement goldens

**Measured 2026-08-30**, at `main` HEAD `f912463`, on `darwin/arm64`, `go1.26.0`.
Every number below was observed; the command that produced it is quoted beside it.
Nothing here was taken from a commit message.

---

## 1. The move, measured

`cd folio-go && go test -run TestStatementGoldenFixtures ./...`

| fixture | pages | recorded bytes | produced bytes | delta | delta / page |
|---|---|---|---|---|---|
| `statement-1`  |  1 |  76,740 |  76,744 |   +4 | +4 |
| `statement-5`  |  5 | 127,343 | 127,363 |  +20 | +4 |
| `statement-20` | 20 | 269,804 | 269,884 |  +80 | +4 |
| `statement-50` | 50 | 555,629 | 555,829 | +200 | +4 |

Digests (`shasum -a 256`):

| fixture | recorded sha256 (at `f912463`) | produced sha256 (HEAD render) |
|---|---|---|
| `statement-1`  | `ef58bbf6dac1c3d4a5d679a77f9907a8d45f02ccd3f886c4d4e7cbdf9e86611d` | `114df1d6508981d4eb162c585ff6f01eedf2a75393a5a2a9b649809e8ac968db` |
| `statement-5`  | `7f67b317c0a1925a404f8435bd4736b85e831a213f5a69fc2a2934a742ff950f` | `70dce051495cf68daa71fe8185aa2467acfd82d10fb195439a4d71bcf41944d0` |
| `statement-20` | `be6f5e27af94e62e7c15a1814633cc48a2a91c5ee8686f5b76de5dc12e3cd4ed` | `56bfbbd9a7d20a2a9404fc931dfbe70da9d25979eec17cc8027c0f1063f84b9e` |
| `statement-50` | `9c5be7ba7b4f31c7d488c114a377058ec30cec5ffca082d9c76ee26f304c754c` | `5d090b0f01ddb5072636caded9feec2cad24cb16297a1afbba301b2a4802f171` |

The produced renders are committed beside this file as `statement-{1,5,20,50}.pdf`, so the owner's
visual pass reads the same bytes this measurement describes.

**A whole-file `cmp` is useless here and that is worth stating**, because it is the tool a reader
reaches for first: `cmp -l fixtures/statement-1/expected.pdf statement-1.pdf | wc -l` reports 24,700
differing bytes, and 428,241 for `statement-50`. Inserting four bytes early in the file shifts every
later `xref` offset and the `startxref`, so the raw diff is an avalanche with no information in it.
The content-stream diff below is the instrument that reads through it.

---

## 2. The content-stream diff (AC1's deliverable)

Method: render each fixture live from its committed `input.folio` / `data.json` / `params.json`
through the same `renderStatement` seam the golden test uses, resolve both the produced and the
recorded PDF with the EXISTING `splitPageContentStreams` helper
(`folio-go/multi_page_fixture_test.go:377` — no third copy was written), and compare the resolved
per-page streams line by line. Content streams are uncompressed by construction (D-2.0.4, enforced by
the `no-compressor-import` lint rule), so no decode step exists.

Result, on **every page of all four fixtures**:

- page count identical, page tree resolves (`AssertPDFPageTreeResolves`),
- each page's stream grows by exactly **4 bytes**,
- **exactly one line differs**, out of 230 lines on a `statement-1` page and 485–505 on the others,
- the differing line is the same one everywhere:

```
statement-1, statement-5   recorded: 1 0 0 1 436 53.88 Tm
                           produced: 1 0 0 1 522.474 53.88 Tm

statement-20, statement-50 recorded: 1 0 0 1 436 53.88 Tm
                           produced: 1 0 0 1 514.466 53.88 Tm
```

The `Tj` operand on the following line — the glyph string — is byte-identical. The same glyphs, at a
different position.

Moved pages: 1/1, 5/5, 20/20, 50/50. The move is uniform; the `move is not uniform` halt does not
fire.

---

## 3. Candidate discrimination — the three classes, separated before one is named

The three emission classes that could have moved bytes were separated mechanically, not by reading
the diff and telling a story about it.

**Operator census over the joined page streams of each fixture**, counting `re`, `f`, `S`, `rg`,
`m`, `l`, `Tm`, `Tj`, `TJ`, `q`, `Q`, `w`:

> **IDENTICAL, recorded vs produced, for all four fixtures.**

That result eliminates two of the three classes outright:

- **Box paint** (`element_box.go` → `table_render.go` → `internal/pdf/rectdoc.go`) would have added
  `re`/`f`/`S`/`m`/`l` operators. None appeared. `rectdoc.go`'s `appendEdge` prefix→postfix flip is
  the seductive candidate — +1 byte per call, 4 calls per stroked rect, multiplying to exactly 4 —
  but its gate is `r.HasStroke`, which is set only for a declared `style.border`, and no statement
  template declares one. The census confirms the path is never reached rather than assuming it.
- **Text ink** (`internal/pdf/textdoc.go:829-834`) emits a `q … rg` bracket only when `HasColor` is
  true, and emits zero bytes otherwise. No `rg` count changed. This is why Epic 10 (`304442f`) moved
  nothing.
- **Geometry** is what is left, and it is what the diff shows: a variable-width `Tm` x-operand. A
  coordinate change adds bytes **without adding an operator**, which is exactly the signature the
  census reports.

---

## 4. The byte arithmetic, which is the real test of the attribution

`internal/pdf/numbers.go:39` `appendLength` spells a `geom.Length` as sign, integer part, and up to
three trailing-zero-trimmed fractional digits. So:

```
"436"      →  3 characters
"522.474"  →  7 characters      +4 bytes
"514.466"  →  7 characters      +4 bytes
```

Four bytes, once.

**And +4 is contingent on the decimal spelling, not an invariant of alignment** — worth stating so a
future reader does not carry it forward as a law. `appendLength` emits up to three fractional digits
with trailing zeros trimmed, so the delta is +4 only because both new x-values happen to spell as
seven characters against `436`'s three. A value spelling `522.47` would give +3 and `1022.474` would
give +5; a value landing on a whole point would give +1 or less. The uniformity checked in §2 — and
therefore the `move is not uniform` halt not firing — is a property **of these four documents**,
which share one footer element, one page size and one margin. It does not generalise to a corpus
whose aligned elements sit at different coordinates, and a future alignment change should re-measure
rather than expect 4.

The element that carries it is the page footer `e4`:

```json
{"id": "e4", "type": "text", "x": 400, "y": 8, "width": 123, "height": 12,
 "value": "Page {{page}} of {{pages}}",
 "style": {"fontFamily": "body", "fontSize": 7, "align": "right"}}
```

It sits in the `pageFooter` band, so it is drawn **once per page**. One site × 4 bytes × N pages
gives +4 / +20 / +80 / +200 for N = 1 / 5 / 20 / 50 — which is precisely the file-level deltas in
§1, with no residue. The arithmetic closes.

## 5. Why the new position is the correct one

Element coordinates are band-relative; the absolute position adds the 36pt page margin.

```
box left  edge = 36 + 400       = 436.000
box right edge = 36 + 400 + 123 = 559.000
```

**559 is the BOX's right edge, and it is NOT the page's content right edge** — a distinction this
story's intent contract blurs and this file corrects. The content right edge is
`595.276 − 36 = 559.276` (A4 media-box width from `folio-go/internal/pdf/document.go:22`, less the
36pt right margin), so the declared box stops **0.276pt inside** it. Alignment pins text to the box
it is given and knows nothing about the margin; the 0.276pt is the template author's rounding. The
arithmetic below is stated against 559.000, the box edge, which is the number the rule actually
produces.

- The **recorded** x is `436.000` on every page of every fixture — the box's **left** edge, exactly.
  `style.align: "right"` was parsed, validated, round-tripped and displayed, and then ignored at
  emission.
- The **produced** x is the box's **right** edge minus the packed line width:
  `522.474 + 36.526 = 559.000` and `514.466 + 44.534 = 559.000`. Both land on the box's right edge
  exactly.

The two distinct x values are the two distinct footer widths, not two behaviours: `statement-1` and
`statement-5` resolve an 11-glyph footer (`Page 1 of 1` / `Page 5 of 5`), `statement-20` and
`statement-50` a 13-glyph one (`Page 20 of 20` / `Page 50 of 50`). The 8.008pt difference is exactly
two digit advances at 7pt (4.004pt each), confirming tabular digits. Measured with
`statementPageRuns`, which decodes each run through its own `/ToUnicode` CMap:

```
statement-1  p1  "Page 1 of 1"    recorded x=436.000   produced x=522.474
statement-5  p1..p5               recorded x=436.000   produced x=522.474
statement-20 p1..p20              recorded x=436.000   produced x=514.466
statement-50 p1..p50              recorded x=436.000   produced x=514.466
```

The offset is constant within a document because the alignment slack is computed once from the
resolved footer form, which is why the per-page delta is a uniform 4 and not a function of the page
number's digit count.

## 6. The causal proof

Attribution by diff names a site. This step proves the site is the *cause*, and it is the strongest
evidence in this file:

> `folio-go/render.go`'s two alignment offsets were temporarily multiplied by zero —
> `lineX := el.X + geom.Length(0)*textAlignOffset(...)` and the matching `textValignOffset(...)` —
> and `go test -run TestStatementGoldenFixtures .` was re-run.
>
> **All four fixtures went green, byte for byte, against their recorded goldens.** The edit was then
> reverted (`git diff --stat render.go` → empty).

Nothing else in the tree contributes a single byte to the move. `textAlignOffset` is the whole delta.

## 7. The commit, and the population that closes the red set

`git log --diff-filter=A -- folio-go/text_alignment.go` → **`791ed00`** ("Make a component's box
print, on the page and on the canvas"). The file was created in that commit — with no story file, no
AC and no FR — and wired into `render.go`'s band text collection in the same commit. `791ed00` also
re-recorded no fixture, which is why the divergence shows up as render-vs-golden rather than as a
recorded move.

`grep -o '"align"[^,}]*' fixtures/*/input.folio`:

- Exactly four templates in the whole corpus declare a non-`left` alignment, and they are exactly the
  four that went red: `fixtures/statement-{1,5,20,50}/input.folio`. No other fixture could have moved.
- Each declares `"align": "right"` twice: once on the table column `ed` (Amount) — a *column*
  alignment, a different mechanism that predates `791ed00` and did not move — and once on the
  element-level `style.align` of `e4`, which is the corpus's only element-level non-left alignment.
  Only the `e4` line differs in the diff, which agrees.
- `grep -rn 'valign' fixtures/` returns **nothing**. The vertical half of `791ed00` —
  `textValignOffset` — shipped with **zero corpus coverage**. Measured the same way as §6:
  multiplying `textValignOffset` alone by zero and running `go test ./...` reddens
  `TestAlignedTextElementsMoveInsideTheirDeclaredBox` and
  `TestCanvasPaintMatchesTheShippingRunPathUnderAlignment` — the behaviour suite does cover it — and
  **no golden anywhere in the repository**.
- Nor does any fixture declare `align: "center"`. The full census is
  `grep -oh '"align"[^,}]*' fixtures/*/input.folio` → **16 `"left"`, 8 `"right"`, no `"center"`**.
  `center` and `middle` are the only two branches that round — both are
  `geom.ScaleRound(slack, 1, 2)` at `text_alignment.go:56` and `:74`, the file's only rounding — so
  the one construct where a cross-target divergence could plausibly appear is declared by no document
  the matrix renders. Filed together as **DW-24**, owner Story 7.1, trigger: any story touching the
  vertical model, `text_alignment.go`, or the rounding rule.

## 8. Heavy-fixture verification (AC5)

CI never runs the 5-, 20- and 50-page legs; they were run here rather than assumed from the one-page
case. Each was rendered in full, its page tree resolved through `AssertPDFPageTreeResolves`, its
resolved page count compared against `statementFixtures`' declared count, and its sha256 computed:

| fixture | declared pages | pages the page tree resolved | rendered bytes | rendered sha256 |
|---|---|---|---|---|
| `statement-1`  |  1 |  1 |  76,744 | `114df1d6508981d4eb162c585ff6f01eedf2a75393a5a2a9b649809e8ac968db` |
| `statement-5`  |  5 |  5 | 127,363 | `70dce051495cf68daa71fe8185aa2467acfd82d10fb195439a4d71bcf41944d0` |
| `statement-20` | 20 | 20 | 269,884 | `56bfbbd9a7d20a2a9404fc931dfbe70da9d25979eec17cc8027c0f1063f84b9e` |
| `statement-50` | 50 | 50 | 555,829 | `5d090b0f01ddb5072636caded9feec2cad24cb16297a1afbba301b2a4802f171` |

No render failed and no page tree was malformed. All four page counts are verified, not assumed.

These four digests are the ones re-recorded into `fixtures/*/expected.json`, the four `README.md`
digest lines, and `folio-go/byte_neutrality_test.go`'s `goldenDigestRecord` — and they are the four
the owner's re-attestation must name.

## 9. Reproducing this

```
cd folio-go
go test -run TestStatementGoldenFixtures ./...        # the byte counts in §1
git log --diff-filter=A -- folio-go/text_alignment.go # the commit in §7
grep -o '"align"[^,}]*' ../fixtures/*/input.folio     # the population in §7
grep -rn 'valign' ../fixtures/                        # empty — the coverage gap in §7
```

The diff harness of §2–§3 and the causal experiment of §6 were throwaway: a temporary
`folio-go/*_test.go` file using `t.TempDir()` semantics and the existing `splitPageContentStreams`,
and a temporary two-token edit to `render.go`. Neither is committed; both are reproducible from the
descriptions above in a few minutes.
