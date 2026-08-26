Deliberately empty of DW-2's remaining open artifact (no folio-designer/
lockfile, no pdfjs-dist NOTICE) — the compliant near-miss for the
absences rule (AC21). Story 2.2 (AC5) retired this fixture's third,
fonts-dir artifact: folio-go/fonts/ now ships faces, and its own
fail-closed guard is ScanFontsAssets (fontsassets.go), not this one.
Story 3.7 (AC13, D-000.67 part 1) removed this fixture's fourth
artifact, a folio-go/internal/clean/clean.go stand-in for the
content-check mechanism: that mechanism (and both fixtures it existed
to exercise) was removed together with the check kind it witnessed —
see absences.go's own doc comment.
