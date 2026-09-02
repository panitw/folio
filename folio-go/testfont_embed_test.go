package folio

import _ "embed"

// testRobotoFontBytes is Story 1.5's one small Latin test face (AC26),
// embedded directly into the TEST BINARY via go:embed — not via a
// filesystem read at runtime. This is required for the subprocess and
// cross-target matrix harnesses (matrix_test.go, //go:build matrix):
// both build this package's test binary with `go test -c` and then run
// ONLY that single binary — Story 1.2's Docker runner bind-mounts just
// the prebuilt binaries directory, and js/wasm has no filesystem access
// to testdata/ at all — so a font file living only on disk would be
// unreachable from a cross-compiled child process. go:embed in a
// _test.go file at the module root is unaffected by AC3's guard, which
// scans only folio-go/internal/ for a go:embed directive naming a font
// file (AD-8's Rule is about internal/ packages never embedding font
// data as part of the module's PUBLIC surface — this is a test fixture,
// not shipped font data, and does not name a directory this story's
// scope fence (AC26) reserves for Story 2.2).
//
//go:embed testdata/fonts/Roboto-Regular.ttf
var testRobotoFontBytes []byte

// testNotoSansThaiVariableFontBytes is Story 2.2's variable-font
// (fvar/gvar/avar, glyf outlines) test fixture — the SAME reasoning as
// testRobotoFontBytes above applies (cross-target matrix harness needs
// it embedded in the test binary, not read from disk at runtime). Used
// by this package's own AC4 coverage/fallback tests to exercise a
// SECOND, Thai-covering face alongside Roboto's Latin-only coverage —
// not the shipped copy (folio-go/fonts/notosansthai/).
//
//go:embed testdata/fonts/notosansthai-variable-testonly/NotoSansThai-VF.ttf
var testNotoSansThaiVariableFontBytes []byte

// testShippedNotoSans/Thai/SC are Story 2.2's actual SHIPPED faces
// (folio-go/fonts/...), embedded a SECOND time directly here rather
// than obtained via `github.com/panitw/folio/folio-go/fonts`.Shipped().
// This package's OWN test files (render_test.go, matrix_test.go) are
// `package folio` (white-box tests): `go vet` correctly refuses that
// import as a cycle — `fonts` imports `folio`, and a white-box test
// file is compiled as part of package folio itself, unlike an external
// `package folio_test` file (ac4_coverage_test.go), which imports
// `fonts` safely. Duplicating the embed here is the same shape
// testRobotoFontBytes above already uses, for the same underlying
// reason (the cross-target/wasm matrix harness needs the bytes IN the
// test binary, not resolved through another package's own go:embed at
// a path this binary might not carry — though here it's a build
// requirement, not a runtime one).
//
//go:embed fonts/notosans/NotoSans-Regular.ttf
var testShippedNotoSans []byte

//go:embed fonts/notosansthai/NotoSansThai-Regular.ttf
var testShippedNotoSansThai []byte

//go:embed fonts/notosanssc/NotoSansSC-Regular.ttf
var testShippedNotoSansSC []byte

// testShippedNotoSansThaiLicence is Story 8.6's addition, and it is here
// for the SAME reason the three faces above are: fixtures/embedded-font/
// now has to state the face's terms, embeddedFontTemplateJSON() builds
// that document rather than transcribing it, and the cross-target matrix
// runs a prebuilt binary with no filesystem under it. A licence read
// from disk at runtime would be unreachable on js/wasm — and a licence
// hand-copied into a Go string constant would be a SECOND authority on
// what the terms are, which is precisely the thing this story exists to
// stop happening inside a `.folio`.
//
// Its FIRST LINE is also where the fixture's `copyright` comes from
// (embeddedFontCopyright), so the record and the terms it belongs to
// have one source between them.
//
//go:embed fonts/notosansthai/LICENSE-OFL.txt
var testShippedNotoSansThaiLicence string

// testShippedFontSet returns the same (name -> bytes) pairs
// `fonts.Shipped()` returns, built from the test-local embeds above so
// this package's own white-box test files never import `fonts` (see
// the comment above).
//
// This is a HAND-COPY of another package's map, and nothing in this
// file can detect it drifting: a face added to fonts.Shipped() would
// silently never reach the matrix fixture, which is exactly the
// mechanism by which the golden set comes to cover fewer (face,
// instance) pairs than actually ship (D-2.2.1's V7 hazard). The parity
// assertion that closes that lives in shipped_faces_test.go, in
// `package folio_test`, which CAN import `fonts` without a cycle:
// TestTestShippedFontSetMatchesFontsShipped compares both key sets AND
// the bytes behind them, in both directions.
func testShippedFontSet() FontSet {
	return FontSet{
		"Noto Sans":      testShippedNotoSans,
		"Noto Sans Thai": testShippedNotoSansThai,
		"Noto Sans SC":   testShippedNotoSansSC,
	}
}
