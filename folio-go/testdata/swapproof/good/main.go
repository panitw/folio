// Package main is AC19's control fixture (D-1.7.5, D-000.13's "control
// mutation must be valid syntax with forbidden semantics" pattern
// applied in reverse — this is the UNMUTATED shape, and it must
// compile): the exact same variables as bad/main.go, in the correct
// argument order. If this ever failed to build, the "bad" fixture's
// failure would prove nothing about the swap specifically.
package main

import "github.com/panitw/folio/folio-go"

func main() {
	var t *folio.Template
	var d folio.Data
	var p folio.Params
	var f folio.FontSet
	_, _ = folio.Render(t, d, p, f)
}
