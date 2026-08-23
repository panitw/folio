// Package main is AC19b's residual-gap fixture (illustrative,
// D-1.7.5): an EXPLICIT conversion between the two defined types still
// compiles. The guardrail proves the ACCIDENTAL swap is a compile
// error; it does not and cannot stop a deliberate cast — this fixture
// exists so that fact is measured, not merely asserted in prose.
package main

import "github.com/panitw/folio/folio-go"

func main() {
	var t *folio.Template
	var d folio.Data
	var p folio.Params
	var f folio.FontSet
	_, _ = folio.Render(t, folio.Data(p), folio.Params(d), f)
}
