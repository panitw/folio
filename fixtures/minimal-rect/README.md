# Fixture: minimal-rect

This fixture is the golden record for Story 1.1 — the first PDF `folio-go` ever produced: a
single A4 page containing one filled rectangle, rendered with no compression, no `/Info`
dictionary, and no `/CreationDate` or `/ModDate` (AD-7).

## Contents

- `expected.json` — the normative record. It carries the SHA-256 of the rendered bytes, the
  `folio-go` version (`folioGoVersion`) and the exact Go toolchain version (`goToolchain`) that
  produced the hash. The hash is what tests compare against; the toolchain field exists so a test
  can detect and refuse a toolchain drift (see below) rather than silently re-measuring against it.
- `expected.pdf` — the recorded bytes, kept for human diffing only. **The hash in `expected.json`
  is normative, not this file** — do not "fix" a mismatch by eyeballing the PDF.

## If this fixture's hash test goes red

Under AD-21 and AD-22, a hash change here is **a defect until proven to be an intended, versioned
change**. Do not regenerate this fixture to make a failing test pass (vacuity guard 4 in Story
1.1). Investigate what changed in the rendering pipeline first. Only re-record `expected.json` /
`expected.pdf` deliberately, as part of a change that is explicitly documented as a byte-identity
breaking event (e.g. a Go toolchain bump under AD-22, or a deliberate format change), never as a
routine "make CI green" step.

If `runtime.Version()` at test time differs from the recorded `goToolchain`, the test that reads
this fixture fails on that mismatch specifically (before ever comparing hashes), with a message
explaining that a toolchain bump is a versioned breaking change under AD-22 and must be
re-measured deliberately — this is counter-metric C6.
