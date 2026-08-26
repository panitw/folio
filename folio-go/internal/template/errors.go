package template

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/diag"
)

// LoadError is every load-time failure this package produces (AC41,
// D-1.4.2): a plain Go error for the overwhelming majority of sites —
// internal/diag did not exist before Story 3.6 (D-1.4.2's sequencing
// ruling), and R7's criterion (D-3.6.6) does not license minting a code
// for every one of them now that it does: "an internal-invariant
// violation stays a plain error." Field, ElementID (where applicable)
// and Value are always populated so the message can name what a
// person needs to fix, without minting a stable code for it.
//
// Code is Story 3.6's one addition (R8, DW-6): set ONLY at the two
// call sites (newLoadErrorCoded, below) matching D-1.4.2's
// TABLE_FOOTER_SOURCE_FORBIDDEN parenthetical — footerOf paired with
// footer: "count", and a footer field present with no footer at all.
// Every other call site leaves it at its zero value (""), which
// folio.ParseTemplate's boundary (this package may never import the
// module root, AD-1) reads as "not one of DW-6's two conditions — mint
// TEMPLATE_MALFORMED instead."
type LoadError struct {
	Field     string
	ElementID string // empty when the error is not scoped to one element/column
	Value     string
	Reason    string
	Code      diag.Code // "" unless this is one of DW-6's two coded conditions
}

func (e *LoadError) Error() string {
	if e.ElementID != "" {
		return fmt.Sprintf("template: field %s (element %s): %s (value: %s)", e.Field, e.ElementID, e.Reason, e.Value)
	}
	return fmt.Sprintf("template: field %s: %s (value: %s)", e.Field, e.Reason, e.Value)
}

// newLoadError is the single constructor every UNCODED load-error site
// in this package calls (AC41's enumeration test walks call sites of
// this function).
func newLoadError(field, elementID, value, reason string) error {
	return &LoadError{Field: field, ElementID: elementID, Value: value, Reason: reason}
}

// newLoadErrorCoded is Story 3.6's second constructor (R8/DW-6),
// called ONLY at parse_bands.go's two TABLE_FOOTER_SOURCE_FORBIDDEN
// sites — one code, two sites, because the code names the CONDITION,
// not the line.
func newLoadErrorCoded(field, elementID, value, reason string, code diag.Code) error {
	return &LoadError{Field: field, ElementID: elementID, Value: value, Reason: reason, Code: code}
}
