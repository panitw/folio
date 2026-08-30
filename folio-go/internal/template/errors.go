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
// Code is Story 3.6's one addition (R8, DW-6), extended at Story 4.5 and
// again at Story 7.2. It is set at the coded call sites only
// (newLoadErrorCoded, below). Three of those name a footer SOURCE
// condition — D-1.4.2's
// TABLE_FOOTER_SOURCE_FORBIDDEN parenthetical (footerOf paired with
// footer: "count", and a footer field present with no footer at all),
// plus D-1.4.1's TABLE_FOOTER_SOURCE_UNRESOLVED "out-of-collection
// source" arm (the footerOf prefix check), which Story 3.6 left
// uncoded and Story 4.5 swept in under D-000.67 part 2. Story 7.2 added
// a FOURTH, non-footer coded site: STYLE_LINE_SPACING_INVALID, in
// parse_bands.go, for an out-of-range or inexact `style.lineSpacing`.
// So "coded" no longer means "about a footer source" — it means the
// condition has a stable code of its own, whatever it is about.
// Every other call site leaves it at its zero value (""), which
// folio.ParseTemplate's boundary (this package may never import the
// module root, AD-1) reads as "no stable code — mint TEMPLATE_MALFORMED
// instead." Getting a NEW coded condition wrong is not cosmetic: the
// WASM host replaces a TEMPLATE_MALFORMED message wholesale, so an
// uncoded style rejection reaches the author as "The template could not
// be processed" and names neither the element nor the range. That is the
// CORRECT reading for
// the four remaining uncoded footer sites (a `footer`/`footerOf`/
// `footerFormat` that is not a string, and a `footer` outside the
// closed set sum/count/avg): those are malformed-template/type
// failures, not statements about a footer's numeric SOURCE.
type LoadError struct {
	Field     string
	ElementID string // empty when the error is not scoped to one element/column
	Value     string
	Reason    string
	Code      diag.Code // "" unless this is one of the three coded footer-source conditions
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
// called ONLY at parse_bands.go's three footer-source sites: the two
// TABLE_FOOTER_SOURCE_FORBIDDEN ones (one code, two sites, because the
// code names the CONDITION, not the line) and the single
// TABLE_FOOTER_SOURCE_UNRESOLVED one (the out-of-collection footerOf
// prefix check, swept in at Story 4.5 under D-000.67 part 2 — Story
// 3.6 coded the two FORBIDDEN checks beside it and left this one
// plain, which was an absorption gap, not a decision).
func newLoadErrorCoded(field, elementID, value, reason string, code diag.Code) error {
	return &LoadError{Field: field, ElementID: elementID, Value: value, Reason: reason, Code: code}
}
