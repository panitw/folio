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
