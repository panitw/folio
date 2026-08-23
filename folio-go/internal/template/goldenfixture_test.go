package template

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestWorkedExampleMatchesGoldenFixture is AC16 step 3: a test asserts
// the document's fenced worked-example block is byte-identical to the
// shipped golden fixture, closing the self-ratification loop
// permanently. It extracts the fence from the live document — never
// embeds a copy (D-1.4.7's forbidden third-artifact shape).
func TestWorkedExampleMatchesGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	docBytes := mustReadFile(t, filepath.Join(root, "_bmad-output", "specs", "spec-folio", "folio-format.md"))
	golden := mustReadFile(t, filepath.Join(root, "folio-go", "testdata", "template", "golden", "worked-example.json"))

	fence := extractWorkedExampleFence(t, string(docBytes))
	if fence != string(golden) {
		t.Fatalf("folio-format.md's worked example fence is no longer byte-identical to the golden fixture:\n--- doc ---\n%s\n--- golden ---\n%s", fence, golden)
	}
}

// extractWorkedExampleFence returns the content of the LAST ```json
// fence in the document (the worked example is the final fenced block,
// under "## Worked example").
func extractWorkedExampleFence(t *testing.T, doc string) string {
	t.Helper()
	idx := strings.Index(doc, "## Worked example")
	if idx < 0 {
		t.Fatal("could not find \"## Worked example\" heading in folio-format.md")
	}
	rest := doc[idx:]
	start := strings.Index(rest, "```json\n")
	if start < 0 {
		t.Fatal("could not find a ```json fence under the worked example heading")
	}
	rest = rest[start+len("```json\n"):]
	end := strings.Index(rest, "```")
	if end < 0 {
		t.Fatal("worked example fence was never closed")
	}
	return rest[:end]
}
