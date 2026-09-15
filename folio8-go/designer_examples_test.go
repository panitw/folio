package folio8

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The folio8 Designer ships example templates (startup templates, CAP-4) from
// folio8-designer/public/templates/examples/: each <id>.folio beside its
// <id>.sample.json. The designer build renders them with the CLI in strict mode
// and fails on any diagnostic; this test holds the same two properties in
// `go test`, so an engine change that breaks an example reds here first:
//
//   - the committed bytes are canonical, so a no-op load/save round trip in the
//     designer is byte-identical;
//   - rendering against the sample data raises zero diagnostics.
func TestDesignerExamplesAreCanonicalAndRenderClean(t *testing.T) {
	dir := filepath.Join(repoRootFromTest(t), "folio8-designer", "public", "templates", "examples")
	templates, err := filepath.Glob(filepath.Join(dir, "*.folio"))
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) == 0 {
		t.Fatalf("no example templates found in %s", dir)
	}
	for _, path := range templates {
		id := strings.TrimSuffix(filepath.Base(path), ".folio")
		t.Run(id, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tpl, err := ParseTemplate(source)
			if err != nil {
				t.Fatalf("ParseTemplate(%s): %v", id, err)
			}
			canonical, err := SerializeTemplate(tpl)
			if err != nil {
				t.Fatalf("SerializeTemplate(%s): %v", id, err)
			}
			if !bytes.Equal(source, canonical) {
				t.Errorf("%s.folio is not in canonical form; SerializeTemplate writes:\n%s", id, canonical)
			}
			sample, err := os.ReadFile(filepath.Join(dir, id+".sample.json"))
			if err != nil {
				t.Fatalf("%s has no sample data: %v", id, err)
			}
			res, err := Render(tpl, Data(sample), nil, testShippedFontSet())
			if err != nil {
				t.Fatalf("Render(%s): %v", id, err)
			}
			requireNoDiagnostics(t, id, res.Diagnostics)
			if !bytes.HasPrefix(res.Bytes, []byte("%PDF-")) {
				t.Fatalf("%s: Render did not return a PDF", id)
			}
		})
	}
}
