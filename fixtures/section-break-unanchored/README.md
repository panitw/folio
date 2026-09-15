# `fixtures/section-break-unanchored/` — a legend pushed down below a growing table

The golden for the content band's **`sectionBreakAnchor`** (spec-section-break CAP-7): an
**unanchored** section break.

| File | What it is |
|---|---|
| `input.folio8` | section-break-statement's bilingual (English/Thai) statement on A4, with the break at 390 pt — 20 pt above the transaction-code legend at y 410–508 — and `"sectionBreakAnchor": false` |
| `data.json` | Thirty-five **synthetic** transactions for a fictitious account holder — no real customer data |
| `expected.pdf` | The render, sha256 `3ae4e8d50eccbc10f0aa055490a1184a8b9601496e2ac0d58613319b61b5186b` |

The thirty-five rows end past the line on page 1, and the legend still fits in the room left below
them. Because the break is unanchored, the whole section is pushed down on page 1 by exactly the
distance the rows ran past the line, so the legend sits 20 pt below the last row. Anchored, the same
document would move the legend to an added page 2. The document is one page and prints `Page 1 of 1`.

## What the tests prove

- `TestSectionBreakUnanchoredGoldenFixture` pins `input.folio8` and `data.json` to the Go constants in
  `section_break_unanchored_template.go`, and the render's sha256 to `expected.json` and `expected.pdf`.
- `TestSectionBreakUnanchoredSemanticAcceptance` checks the page model: one page, all thirty-five rows
  on it, the legend drawn once, and its baseline equal to its five-row (declared) baseline plus the
  distance the rows end past the line. The anchored document with the same data has two pages.
- `TestSectionBreakUnanchoredRendersIdenticallyInAFreshProcess` renders it again in a fresh process.
- The four-target hash matrix renders it on darwin/arm64, linux/amd64, linux/arm64 and js/wasm, and
  the wasm engine's preview render is compared with the native one.

## Manual check (not automated)

Open `expected.pdf`. The legend must follow the last row, must not be overdrawn, and must appear once.
