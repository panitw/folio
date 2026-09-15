# `fixtures/section-break-statement/` — a legend moved below a growing table

The golden for the content band's **`sectionBreak`** (spec-section-break CAP-2). It is the first
committed document declaring **`"version": "4.1"`**, because it is the first carrying a section break.

| File | What it is |
|---|---|
| `input.folio` | A bilingual (English/Thai) statement on A4: a customer block, a `transactions[]` table at y 38, and a transaction-code legend at y 410–508 below a `sectionBreak` at 400 pt |
| `data.json` | Forty **synthetic** transactions for a fictitious account holder — no real customer data |
| `expected.pdf` | The render, sha256 `4f10e8b62fc4bbdb1abc807c9ec7c8d9600b1d04571f5bc904a7bb3cc86f0dae` |

Without the break, the forty rows would run down page 1 and draw over the legend (AD-24: nothing moves
because a neighbour grew). With it, the rows end below the line on page 1, so everything declared at or
below the line moves to an added page 2. There the legend sits at exactly its declared position, with
nothing drawn above it but the page header, and both pages print `Page X of 2`.

## What the tests prove

- `TestSectionBreakStatementGoldenFixture` pins `input.folio` and `data.json` to the Go constants in
  `section_break_statement_template.go`, and the render's sha256 to `expected.json` and `expected.pdf`.
- `TestSectionBreakStatementSemanticAcceptance` checks the page model: two pages, all forty rows on
  page 1, the legend on page 2 only, and its baseline equal to its page-1 baseline when the same
  document renders five rows. It also decodes the PDF's text per font resource to check that the
  legend is drawn once, on the last page, and that `Page X of 2` is correct on both pages.
- `TestSectionBreakStatementRendersIdenticallyInAFreshProcess` renders it again in a fresh process and
  compares bytes.
- The four-target hash matrix renders it on darwin/arm64, linux/amd64, linux/arm64 and js/wasm.

## Manual check (not automated)

Open `expected.pdf`. The legend must not be overdrawn, and it must appear once, at its designed
position, on the last page.
