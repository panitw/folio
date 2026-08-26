package folio

import (
	"fmt"

	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/template"
)

// This file is Story 3.6's D-3.6.3, RULED ARM A: AD-14 requires "every
// failure mode named in FR41 has a code" and a code is a FIELD of
// Diagnostic — a bare fmt.Errorf has no fields. So FR41's four Error
// modes (malformed template, unresolvable binding, invalid expression,
// unlayoutable content) must produce a Diagnostic, and an aborting one
// must carry SeverityError. Arm A reconciles that with Go's ordinary
// error-return convention rather than departing from either: it wraps
// each existing, pre-3.6 error type, so no existing error type,
// message or errors.As target changes (AC8).
//
// This is FORCED, not preferred (D-3.6.3's own grounds): arm B (leave
// SeverityError unexercised) needed both a scope cut and an AD-14
// amendment with no reason on the table to amend it, and arm C
// (construct SeverityError but keep it caller-unreachable) is
// D-000.9's rejected shape at the type level.

// RenderError is the first publicly constructible SeverityError value
// in this module's history (AC8). It carries the AD-14 Diagnostic AC4's
// four Error modes require, WITHOUT replacing the underlying error a
// caller may already be matching on with errors.As:
//
//   - Error() returns the wrapped error's OWN message, byte-for-byte —
//     printing a RenderError looks exactly like printing what it wraps,
//     because AC8 requires "no existing error type, message or
//     errors.As target changes";
//   - Unwrap() returns the wrapped error, so errors.As/errors.Is walks
//     straight through to it — a *layout.OverflowError,
//     *template.LoadError or expr.KernelOverflowError inside a
//     RenderError is exactly as recoverable as it always was.
//
// A caller wanting the STABLE CODE (AC9: "match on the code, never on
// message text") uses errors.As(err, &renderErr) and reads
// renderErr.Diagnostic.Code — never Message, and never the wrapped
// error's own type, which stays an implementation detail this type
// does not require a caller to know about.
type RenderError struct {
	// Diagnostic is AD-14's SeverityError-carrying value: Code names
	// which of FR41's four Error modes this is; ElementID and DataPath
	// are populated where the failure mode has them (AD-10); Message
	// duplicates Err.Error() — present so a Diagnostic value taken in
	// isolation (e.g. logged, or read off Result.Diagnostics-shaped
	// tooling) is still self-describing, never because it is parsed
	// (Diagnostic.Message's own doc comment).
	Diagnostic Diagnostic

	// Err is the underlying, PRE-EXISTING error this Diagnostic
	// describes, unchanged from what this module returned before Story
	// 3.6.
	Err error
}

// Error returns Err's own message, unchanged — printing a RenderError
// is indistinguishable from printing what it wraps (AC8).
func (e *RenderError) Error() string { return e.Err.Error() }

// Unwrap exposes Err to errors.As/errors.Is, so every pre-existing
// target (*layout.OverflowError, *template.LoadError,
// expr.KernelOverflowError) keeps resolving through this wrapper.
func (e *RenderError) Unwrap() error { return e.Err }

// newRenderError builds a RenderError: a SeverityError Diagnostic
// carrying code/elementID/dataPath, wrapping err. err must be non-nil.
func newRenderError(code, elementID, dataPath string, err error) *RenderError {
	return &RenderError{
		Diagnostic: Diagnostic{
			Severity:  SeverityError,
			Code:      code,
			ElementID: elementID,
			DataPath:  dataPath,
			Message:   err.Error(),
		},
		Err: err,
	}
}

// wrapTemplateError is ParseTemplate's boundary for FR41's "malformed
// template" mode (AC4/AC8) AND for R8's TABLE_FOOTER_SOURCE_FORBIDDEN
// (DW-6): both originate inside internal/template, which may not
// import the module root (AD-1) and so cannot construct a Diagnostic
// itself. A *template.LoadError with an explicit Code (set only at
// parse_bands.go's three footer-source sites — two DW-6 FORBIDDEN
// sites, plus the out-of-collection UNRESOLVED site swept in at Story
// 4.5 per D-000.67 part 2 — see internal/template/errors.go's own doc
// comment) keeps that code; every other LoadError — the
// overwhelming majority of load-time failures — becomes
// DiagCodeTemplateMalformed. A non-LoadError (should not occur; kept
// as a fallback rather than a panic, AD-14) is wrapped the same way.
func wrapTemplateError(err error) error {
	if le, ok := err.(*template.LoadError); ok {
		code := DiagCodeTemplateMalformed
		if le.Code != "" {
			code = string(le.Code)
		}
		return newRenderError(code, le.ElementID, "", err)
	}
	return newRenderError(DiagCodeTemplateMalformed, "", "", err)
}

// wrapOverflowError is layout.Paginate's boundary for FR41's
// "unlayoutable content" mode (AC4/AC8, R9): an element (a text line
// or an image) taller than the content window it must fit inside
// (internal/layout.OverflowError, FR44/D-2.6.1).
//
// AC8's "no existing error type, message or errors.As target changes"
// applies to the MESSAGE too: both of this function's call sites
// previously returned fmt.Errorf("folio: Render: %w", err) directly, so
// wrapped is built the SAME way here before being handed to
// newRenderError — Error() on the result is byte-identical to what
// Render returned before this story, and errors.As still resolves
// *layout.OverflowError through the extra Unwrap hop (RenderError ->
// the fmt-wrapped error -> the OverflowError).
func wrapOverflowError(err error) error {
	wrapped := fmt.Errorf("folio: Render: %w", err)
	elementID := ""
	if oe, ok := err.(*layout.OverflowError); ok {
		elementID = oe.ElementID
	}
	return newRenderError(DiagCodeContentUnlayoutable, elementID, "", wrapped)
}
