Deliberately empty of DW-2's remaining open artifact (no folio-designer/
lockfile, no pdfjs-dist NOTICE) — the compliant near-miss for the
absences rule (AC21). Story 2.2 (AC5) retired this fixture's third,
fonts-dir artifact: folio-go/fonts/ now ships faces, and its own
fail-closed guard is ScanFontsAssets (fontsassets.go), not this one.
