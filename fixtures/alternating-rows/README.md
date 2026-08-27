# Alternating rows golden fixture

This fixture records Story 4.8's odd-zero-based-index convention for
`table.altRowBackground`. The one-column table has five data rows, no ordinary body
background, and a distinct header background. The produced PDF therefore contains exactly two
alternate-colour fills: collection indexes 1 and 3. Indexes 0, 2, and 4 remain unfilled.

The committed PDF was rendered on 2026-08-27 with Go `go1.26.0` on `darwin/arm64`:

```text
CGO_ENABLED=0 GOWORK=off go run ./cmd/folio render -data ../fixtures/alternating-rows/data.json -o ../fixtures/alternating-rows/expected.pdf ../fixtures/alternating-rows/input.folio
```

Its independently recorded SHA-256 digest is
`e491d628ecd1dae9ad2d396341c014fb9dc5ce1e55c535a2f88bdae15b0e8bbd`.

Independent structural validation used qpdf 12.4.0 on 2026-08-27:

```text
$ qpdf --check fixtures/alternating-rows/expected.pdf
checking fixtures/alternating-rows/expected.pdf
PDF Version: 1.7
File is not encrypted
File is not linearized
No syntax or stream encoding errors found; the file may still contain
errors that qpdf cannot detect
```

The document has one page. Its content stream carries alternate-colour fills at the two
test-owned vertical positions `741.374` and `716.858` points, in that order. The untagged fixture
test re-renders through the public API, performs the byte comparison, validates the PDF page tree,
and reads the produced PDF bytes to verify those two semantic fill operations.

Matrix registration covers slug `alternating-rows`, `input.folio`, `data.json`, `expected.pdf`, and
`expected.json`. Story 4.8 runs the native `darwin/arm64` leg when registering this document. That
single leg proves the document executes and hashes on its native host; cross-target equality remains
deferred to the Epic 4 boundary gate.
