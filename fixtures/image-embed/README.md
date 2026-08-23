# Fixture: image-embed

Story 1.8's golden record — the first PDF `folio-go` produces with a real, embedded, passthrough
image: an A4 page whose content band draws one `image` element (a 100pt x 60pt box) referencing a
real, supported, non-square 3x2 8-bit RGB PNG asset embedded inside the document itself (FR33,
AD-9).

It sits **beside** `fixtures/minimal-rect/` and `fixtures/font-text/`, replacing neither (AC22,
D-1.8.6): `minimal-rect/` remains the fontless emission baseline and `font-text/` remains the
subsetting coverage document; this is the first fixture covering the **image-embedding** path
through the public `Render` API — the PNG `/FlateDecode` + `/DecodeParms /Predictor 15` passthrough
route (D-1.8.1).

## Contents

- `input.folio` — the `.folio` document rendered to produce this fixture, byte-identical to the
  `imageTestTemplateJSON` constant in `folio-go/render_test.go` (AC25a, the same discipline
  `fixtures/font-text/input.folio` carries).
- `expected.json` — the normative record: SHA-256 of the rendered bytes, `folioGoVersion`, and the
  exact Go toolchain version that produced the hash (AC16, D-1.2.2: `sha256` is always a JSON
  string of exactly 64 lower-case hex characters, never a per-target map). `goToolchain` matches
  the other two fixtures' recorded value exactly (the shared-toolchain assertion, AC25).
- `expected.pdf` — the recorded bytes, kept for human diffing only. **The hash in `expected.json`
  is normative, not this file.**

## Asset

The embedded PNG (81 bytes decoded, key
`5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078`) was generated with Go's
`image/png` encoder and independently cross-checked with Python's `hashlib`/`base64` during story
creation: signature, IHDR (`3x2`, 8-bit, colour type 2 — truecolor, no alpha, non-interlaced),
IDAT/IEND all verified by the second tool before being committed here. It is deliberately
**non-square** (AC27a) so this fixture exercises AC13's binding-axis choice (a 3:2 image inside a
100:60 = 5:3 box — width binds).

## Passthrough route asserted

`TestRenderMatchesImageEmbedGoldenFixture` (`folio-go/fixture_test.go`) asserts the rendered bytes
contain an image XObject (`/Subtype /Image`) BEFORE comparing any hash (AC23's vacuity guard, the
same shape `font-text/`'s `FontFile2` guard uses) — a render that silently dropped the image would
otherwise match nothing and prove nothing.

## If this fixture's hash test goes red

Same rule as the other two fixtures (AD-21, AD-22): a hash change here is a defect until proven to
be an intended, versioned change. Never regenerate this fixture just to make a failing test pass;
investigate first, and re-record deliberately as part of a documented breaking change.
