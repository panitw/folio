# `fixtures/multi-page-flow/` — a section break on every page, and Page Break off

The golden for SPEC-multi-pages **CAP-6** and **CAP-9**: a document with three designed pages, written
in the `pages` shape, where each page has its own section break or follows the page before it.

| File | What it is |
|---|---|
| `input.folio` | A canonical three-page document on A4. Page 1 holds a synthetic statement: two heading lines, a `transactions[]` table and, below an **unanchored** section break at 600pt, a code legend. Page 2 has `"pageBreak": false` and holds a `charges[]` table and, below an **anchored** section break at 400pt, a note. Page 3 has `"pageBreak": false` and holds a signature block. The shared page header and page footer print `Page {{page}} of {{pages}}`. |
| `data.json` | Sixty **synthetic** transactions and forty synthetic service charges for a fictitious account holder — no real customer data |
| `expected.pdf` | The render, sha256 `787d4707423f67975f1c26cee7d29002561b2ae5494e1f6f4acb50a12299cb07` |

Page 1's rows run onto output page 2, and its unanchored legend follows them there. Page 2 has Page
Break off, but its charges cross its own break and need two output pages, so it **does not fit** after
page 1: it starts at the top of output page 3, and its anchored note lands on output page 4 at its
declared position. Page 3 has Page Break off and **fits** under page 2's note on output page 4, moved
down as one block with its internal offsets kept. The PDF has four pages, every page carries the page
header and page footer, and the footers read `Page 1 of 4` to `Page 4 of 4`.

## What the tests prove

- `TestMultiPageFlowGoldenFixture` pins `input.folio` and `data.json` to the Go constants in
  `multi_page_flow_template.go`, checks that `input.folio` saves back byte-for-byte, and pins the
  render's sha256 to `expected.json` and `expected.pdf`.
- `TestMultiPageFlowSemanticAcceptance` checks the page model against oracles rendered from each
  page alone as a one-page document: pages 1 and 2 draw exactly what they draw alone, and page 3's
  block is drawn on output page 4 moved down by exactly where page 2's content ends.
- `TestMultiPageFlowRendersIdenticallyInAFreshProcess` renders it again in a fresh process.
- The four-target hash matrix renders it on darwin/arm64, linux/amd64, linux/arm64 and js/wasm.

## Manual check (not automated)

Open `expected.pdf`. It must have four pages: the legend directly under the last transaction on page 2,
the charges starting at the top of page 3, and the signature block directly under the charges note on
page 4. The footers must read 1–4 of 4.
