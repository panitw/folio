# `fixtures/multi-page-statement/` — a second designed page after a statement that overflows

The golden for SPEC-multi-pages **CAP-5**: a document with two designed pages, written in the
`pages` shape.

| File | What it is |
|---|---|
| `input.folio8` | A canonical two-page document on A4. Page 1 holds a synthetic statement: two heading lines and a `transactions[]` table. Page 2 holds static terms, a signature line and its caption, with `"pageBreak": true`. The shared page header and page footer print `Page {{page}} of {{pages}}`. |
| `data.json` | One hundred and thirty **synthetic** transactions for a fictitious account holder — no real customer data |
| `expected.pdf` | The render, sha256 `4b0367f57e39141e82ef67c7ad701f78c9167c8ebb9a795d3fc8f1f59c52c130` |

Page 1's table runs three output pages. Page 2 starts a new output page after them, so its content is
drawn on output page 4 at its declared positions. The PDF has four pages, every page carries the
page header and page footer, and the footers read `Page 1 of 4` to `Page 4 of 4`.

## What the tests prove

- `TestMultiPageStatementGoldenFixture` pins `input.folio8` and `data.json` to the Go constants in
  `multi_page_statement_template.go`, checks that `input.folio8` saves back byte-for-byte, and pins
  the render's sha256 to `expected.json` and `expected.pdf`.
- `TestMultiPageStatementSemanticAcceptance` checks the page model: four pages, the rows on pages 1
  to 3 only, and page 2's heading on page 4 at exactly the baseline it has when page 2's elements
  are rendered alone as a one-page document.
- `TestMultiPageStatementRendersIdenticallyInAFreshProcess` renders it again in a fresh process.
- The four-target hash matrix renders it on darwin/arm64, linux/amd64, linux/arm64 and js/wasm.

## Manual check (not automated)

Open `expected.pdf`. It must have four pages, page 2's terms must start at the top of page 4, and the
footers must read 1–4 of 4.
