package folio

import "testing"

// TestDiagCodeBridgePreservesExactStrings is AC3/D-3.6.4: the two
// codes Story 2.8 and Story 3.3 shipped bridge to internal/diag with
// their EXACT string values, not merely a compiling redefinition. AD-14
// makes changing a code's meaning a breaking change, and a bridge that
// alters one byte is that change wearing a refactor — invisible in a
// diff that looks like a tidy-up. The expected strings below are
// literals this test owns, independent of both diagnostic.go and
// internal/diag/diag.go: if either side drifts, this test catches it
// even though the two source files would still compile against each
// other.
func TestDiagCodeBridgePreservesExactStrings(t *testing.T) {
	if DiagCodeTextClippedWidth != "TEXT_CLIPPED_WIDTH" {
		t.Errorf("DiagCodeTextClippedWidth = %q, want %q", DiagCodeTextClippedWidth, "TEXT_CLIPPED_WIDTH")
	}
	if DiagCodeEmptyAverage != "AGGREGATE_EMPTY_AVERAGE" {
		t.Errorf("DiagCodeEmptyAverage = %q, want %q", DiagCodeEmptyAverage, "AGGREGATE_EMPTY_AVERAGE")
	}
}
