package manifest

import (
	"strings"
	"testing"
)

// TestRenderRowShapeIncludesServesAndLabel guards against a defect
// TestManifestUpToDate cannot catch on its own today: the live graph has
// zero external dependencies (F-8), so the "|---|" table branch of
// Render is never exercised by that test, and a column silently dropped
// from row formatting would go unnoticed until a real dependency
// arrived (V10: healthy output and dead output must never be the same
// value — an empty manifest table is what today's healthy AND a broken
// row-formatter both produce). This test exercises Render directly with
// a fabricated row so the row shape (AC19: labelled with the module it
// serves and shipped/build-time-only) is asserted right now (RP-15).
func TestRenderRowShapeIncludesServesAndLabel(t *testing.T) {
	rows := []Row{
		{Module: "example.test/dep", Version: "v1.0.0", Licence: "MIT", Serves: "lint", ShippedBy: "build-time-only"},
	}
	out := Render(rows)

	var rowLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "example.test/dep") {
			rowLine = line
		}
	}
	if rowLine == "" {
		t.Fatalf("row was not rendered at all:\n%s", out)
	}
	if !strings.Contains(rowLine, "lint") {
		t.Errorf("row is missing its \"serves\" column: %q", rowLine)
	}
	if !strings.Contains(rowLine, "build-time-only") {
		t.Errorf("row is missing its shipped/build-time-only label: %q", rowLine)
	}
	if got := strings.Count(rowLine, "|"); got < 6 {
		t.Errorf("row has %d column separators, want at least 6 (Module|Version|Licence|Serves|Label|): %q", got, rowLine)
	}
}
