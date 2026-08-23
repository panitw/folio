package template

import "fmt"

// LoadError is every load-time failure this package produces (AC41,
// D-1.4.2): a plain Go error, never a minted diag.Diagnostic code —
// internal/diag does not exist yet (D-1.4.2's sequencing ruling), and
// AD-14 lands with Story 1.6, codes with Story 3.6. Field, ElementID
// (where applicable) and Value are always populated so the message can
// name what a person needs to fix, without minting a stable code for
// it.
type LoadError struct {
	Field     string
	ElementID string // empty when the error is not scoped to one element/column
	Value     string
	Reason    string
}

func (e *LoadError) Error() string {
	if e.ElementID != "" {
		return fmt.Sprintf("template: field %s (element %s): %s (value: %s)", e.Field, e.ElementID, e.Reason, e.Value)
	}
	return fmt.Sprintf("template: field %s: %s (value: %s)", e.Field, e.Reason, e.Value)
}

// newLoadError is the single constructor every load-error site in this
// package calls (AC41's enumeration test walks call sites of this
// function).
func newLoadError(field, elementID, value, reason string) error {
	return &LoadError{Field: field, ElementID: elementID, Value: value, Reason: reason}
}
