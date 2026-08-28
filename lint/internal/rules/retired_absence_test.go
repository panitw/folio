package rules

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDesignerAbsenceTripwireIsRetiredByPositiveCoverage prevents DW-2 from
// being "fixed" by restoring an empty absence scanner. The positive npm
// lockfile scan has its own prohibited and unknown red proofs in this package.
func TestDesignerAbsenceTripwireIsRetiredByPositiveCoverage(t *testing.T) {
	root := repoRootFromTest(t)
	var files int
	err := filepath.WalkDir(filepath.Join(root, "lint"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || filepath.Base(path) == "retired_absence_test.go" || strings.Contains(path, string(filepath.Separator)+"testdata"+string(filepath.Separator)) {
			return nil
		}
		files++
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, retired := range []string{"ScanAbsences", "AbsencesStats", "absenceChecks", "absence-designer-project"} {
			if strings.Contains(string(data), retired) {
				return fmt.Errorf("retired absence symbol %q returned in %s", retired, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("retirement guard scanned zero live Go files")
	}
}
