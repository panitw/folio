# Fixture: font-text

This fixture is Story 1.5's golden record — the first PDF `folio-go` produces with real,
embedded, subsetted text: an A4 page whose content band renders `"Hello, World!"` and whose page
footer renders `"Page footer 0123456789"`, both in a subset of the committed Latin test face
(`folio-go/testdata/fonts/Roboto-Regular.ttf`, AC26), embedded as a `Type0`/`Identity-H` composite
font with a `FontFile2` stream and a `ToUnicode` CMap (AC5).

It sits **beside** `fixtures/minimal-rect/`, not in place of it (AC14a, D-1.5.9):
`minimal-rect/` is retained unchanged and reclassified as pinning `internal/pdf`'s **fontless**
emission path; this fixture is the first covering the **font-embedding** path through the public
`Render` API.

## Contents

- `input.folio` — the `.folio` document rendered to produce this fixture. `folio-go/fixture_test.go`
  renders the `fontTestTemplateJSON` constant directly, not this file — exactly as `minimal-rect/`'s
  fixture test renders `internal/pdf.Serialize()` directly rather than reading a file — but
  `TestRenderMatchesFontTextGoldenFixture` now asserts this file is byte-identical to that constant
  before rendering (Finding 14, this story's QA review), so the two can no longer drift apart
  silently.
- `expected.json` — the normative record: SHA-256 of the rendered bytes, `folioGoVersion`, and the
  exact Go toolchain version that produced the hash (Story 1.2 AC16, D-1.2.2: `sha256` is always a
  JSON string of exactly 64 lower-case hex characters, never a per-target map).
- `expected.pdf` — the recorded bytes, kept for human diffing only. **The hash in `expected.json`
  is normative, not this file.**

## Font input

Rendered with `FontSet{"Roboto-Regular": <bytes of folio-go/testdata/fonts/Roboto-Regular.ttf>}`.
The font bytes are NOT copied into this directory a second time — AD-26 governs one committed copy
per redistributed asset, and `folio-go/testdata/fonts/` (with its own `LICENSE-Roboto.txt` and
`NOTICE.md`) is that copy (AC25).

## Measured, at record time

- `head.created` (source face) = 3304067374, `head.modified` = 3573633780 (F-5; embedded subset
  copies these verbatim — AC11a).
- Independent-reader validation: `qpdf --check` reports no syntax/stream errors; PyMuPDF's
  `page.get_text()` on the rendered PDF returns exactly
  `"Hello, World!\nPage footer 0123456789\n"` — confirming the `ToUnicode` CMap and glyph mapping
  are correct, not merely well-formed (AC13).

## If this fixture's hash test goes red

Same rule as `minimal-rect/` (AD-21, AD-22): a hash change here is a defect until proven to be an
intended, versioned change (e.g. a `textshape` version bump, which AD-22 already classes as a
subsetting-affecting change with a moving subset tag as its correct signal — D-1.5.8). Never
regenerate this fixture just to make a failing test pass; investigate first, and re-record
deliberately as part of a documented breaking change.
