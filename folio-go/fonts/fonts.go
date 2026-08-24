// Package fonts holds Story 2.2's three shipped production faces —
// Noto Sans, Noto Sans Thai and Noto Sans SC — each embedded from its
// own subdirectory under folio-go/fonts/, alongside its OFL-1.1 licence
// text and NOTICE (AD-26, AC25). It is declared at the module root, NOT
// under internal/: AD-8's "no package under internal/ embeds font data"
// binds internal/ packages to never carry shipped font DATA — it says
// nothing about a leaf, non-internal package whose entire purpose is to
// hold that data (spine's own scope fence, AC26).
//
// fonts imports package folio (also module root, folio-go/*.go) for its
// FontSet type — a one-directional import (fonts -> folio) that package
// folio never reverses (folio -> fonts): confirmed by inspection of
// folio-go/*.go, none of which imports this package. AC9's blockquote
// makes that direction load-bearing: "folio must NOT import
// folio/fonts... two-package import is load-bearing" — reversing it
// would force every consumer of package folio to pull in ~11.3 MB of
// embedded font data whether or not it uses the shipped set — go:embed
// stores RAW bytes, so that is the figure that lands in every binary,
// not the compressed download budget NFR7 is written in.
package fonts

import (
	_ "embed"

	folio "github.com/panitw/folio/folio-go"
)

// The three shipped faces are STATIC, Regular-only instances derived
// from their upstream variable builds ahead of the build, by
// tools/fontgen/instance_faces.py (D-2.2.4). The .ttf files are
// COMMITTED, not generated at build time, on purpose: generating them
// here would make the shipped font a function of the build environment
// — a different fontTools produces a different font, which produces a
// different PDF — which is AD-22's drift class reintroduced at the
// asset layer. Each face's NOTICE.md records the exact invocation, both
// sha256s and the pinned toolchain, so the derivation can be replayed;
// folio-go/fontgen_matrix_test.go (//go:build matrix) proves it still
// reproduces.

//go:embed notosans/NotoSans-Regular.ttf
var notoSans []byte

//go:embed notosansthai/NotoSansThai-Regular.ttf
var notoSansThai []byte

//go:embed notosanssc/NotoSansSC-Regular.ttf
var notoSansSC []byte

// Shipped returns folio-go's shipped font set — Noto Sans, Noto Sans
// Thai, and Noto Sans SC — keyed by the exact face names a `.folio`
// document's `fonts` fallback chains reference (folio-format.md's own
// `"fonts": {"body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]}`
// example uses these three names verbatim). One expression, no
// arguments (AC9) — callers wire the shipped set into a render with
// `fonts.Shipped()`, never a package-level variable a caller could
// mutate out from under another caller.
func Shipped() folio.FontSet {
	return folio.FontSet{
		"Noto Sans":      notoSans,
		"Noto Sans Thai": notoSansThai,
		"Noto Sans SC":   notoSansSC,
	}
}
