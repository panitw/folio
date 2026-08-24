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

## Re-recorded at Story 2.2 (AD-22 versioned change)

Story 2.2 (`D-2.2.2 (superseded)`, AC6/AC6a) re-derives the six-letter subset tag: it now hashes
the embedded font program's own returned bytes (`subset.Subset()`'s output, in full) instead of a
hash of the sorted glyph-id set alone — the previous form collided across two pinned instances of
one variable face (B6), which this story's shipped Noto faces make a live hazard for the first
time. `(*Font).Subset` is the one shared function every embedded face goes through, so this
fixture's Roboto embedding moves too, even though nothing about Roboto or this document changed.

**The delta was measured, not assumed, before re-recording**: both `expected.pdf` revisions are
**22299 bytes**; exactly **77 bytes differ, in 7 contiguous runs**; `startxref 21957` is identical.
Three of the runs are the six-letter subset tag at its three appearances (`/BaseFont` ×2 and
`/FontDescriptor /FontName`) — `OETEKT` → `HXRYNT`. The remaining four runs are the two duplicated
`/ID` array entries, which move as a *consequence*: `computeID` derives `/ID` as a SHA-256 prefix
over the serialized body up to that point (AD-7), and the tag is upstream of it. **The `FontFile2`
stream itself — the embedded, subsetted Roboto program — is byte-identical**; nothing about the
outlines, widths, or glyph mapping changed. Old hash: `dcd453a1f593c1135c44b52d732683d19419c24743344d3843826fccbafacbdf`.
New hash: recorded in `expected.json` after Story 2.2's in-story four-target matrix run confirmed
all four targets agree (D-2.2-D5's first reason for not switching `font-text`'s face is exactly
this four-target verification — re-recording from one machine would have silently retired it).

`fixture_test.go` additionally now pins the embedded `FontFile2` stream's own SHA-256 as a
permanent constant, independent of the whole-document hash — the assertion AC8's now-corrected
"bytes unchanged" clause needed to become, so a future change that moves the *program* stays
distinguishable from one that only moves the *tag*.

The "Measured, at record time" facts above (`head.created`/`head.modified`, the PyMuPDF text
extraction) did not need re-measurement: the program bytes did not move, only its tag and the
consequently-derived `/ID`.
