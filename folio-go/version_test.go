package folio

import (
	"testing"

	"github.com/panitw/folio/folio-go/internal/expr"
)

// TestLocaleTableVersionSurfacedAtLibraryLevel is Finding 9's fix,
// asserted (this story's QA review): AC6/AD-22 requires the locale
// table's version to be surfaced wherever the library version is
// surfaced. internal/expr.LocaleTableVersion is unexported-adjacent
// (it lives under internal/, unreachable from outside the module);
// this pins that package folio's own exported LocaleTableVersion is
// defined as exactly that value, so a future edit to either constant
// in isolation reddens this test rather than silently drifting.
func TestLocaleTableVersionSurfacedAtLibraryLevel(t *testing.T) {
	if LocaleTableVersion != expr.LocaleTableVersion {
		t.Fatalf("folio.LocaleTableVersion = %d, internal/expr.LocaleTableVersion = %d: must be defined as exactly the same value (AC6/AD-22)", LocaleTableVersion, expr.LocaleTableVersion)
	}
	if LocaleTableVersion < 1 {
		t.Fatalf("LocaleTableVersion = %d, want >= 1", LocaleTableVersion)
	}
}
