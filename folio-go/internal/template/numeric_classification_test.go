package template

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Numeric field kinds (AC24, AC29): kind 1 is "points" (geom.Length,
// decoded through decodePoints/appendPoints); kind 2 is "plain integer"
// (decimal, never scaled — only nextId). There is no third kind:
// report-data numbers (AD-23's exact scaled decimals) are Story 1.6's,
// not modelled here (D-1.4.3).
const (
	numericKindPoints   = 1
	numericKindPlainInt = 2
)

// numericFieldRegistry is AC29's enumeration: every numeric key named
// anywhere in folio-format.md (M-3 measured 11 distinct keys, zero
// escapees), classified into kind 1 or kind 2. This is the coverage
// witness AC29 requires — the developer confirms this list is complete
// for the current document rather than assuming it (M-3: "the document
// will grow, and a test that was true once is not a coverage witness").
var numericFieldRegistry = map[string]int{
	"bottom":       numericKindPoints, // page.margin.bottom, style.padding.bottom
	"fontSize":     numericKindPoints,
	"headerHeight": numericKindPoints,
	"height":       numericKindPoints, // element height, page.size.height, band height
	"left":         numericKindPoints, // page.margin.left, style.padding.left
	"nextId":       numericKindPlainInt,
	"right":        numericKindPoints,
	"top":          numericKindPoints,
	"width":        numericKindPoints, // element width, page.size.width, columns[].width, border.width
	"x":            numericKindPoints,
	"y":            numericKindPoints,
}

// classifyNumericFields returns which of the given field names are
// absent from the registry entirely (AC29: "no field escaped
// classification") and how many fields it considered.
func classifyNumericFields(fields []string, registry map[string]int) (unclassified []string, considered int) {
	for _, f := range fields {
		considered++
		if _, ok := registry[f]; !ok {
			unclassified = append(unclassified, f)
		}
	}
	return unclassified, considered
}

// numericFenceKeyValue matches a "key": NUMBER pair inside a fenced JSON
// code block — the same shape M-3 measured by hand ("every JSON key
// given a numeric literal anywhere in folio-format.md — fences and
// prose"). Deliberately narrower than extractDocKeys' generic key
// extraction (drift_test.go): it only registers a key when the value
// immediately following it is itself a JSON number literal, which is
// what actually distinguishes AC29's "numeric field" from every other
// documented key.
var numericFenceKeyValue = regexp.MustCompile(`"([A-Za-z][A-Za-z0-9_]*)"\s*:\s*-?[0-9]`)

// extractNumericDocKeys scans folio-format.md's fenced code blocks (the
// same fence-delimiter convention extractDocKeys uses) and returns every
// key documented with a numeric JSON literal — this story's finisher
// review, Finding 3: the inventory this test checks against must be
// DERIVED from the document, not hand-copied into the test, or a field
// added to folio-format.md after this test was written escapes silently
// (the exact case measured live: adding a twelfth key, "lineSpacing",
// left the old hardcoded-list version green).
func extractNumericDocKeys(doc []byte) (keys map[string]bool, tokensExtracted int) {
	keys = map[string]bool{}
	inFence := false
	for _, line := range strings.Split(string(doc), "\n") {
		if fenceDelim.MatchString(strings.TrimSpace(line)) {
			inFence = !inFence
			continue
		}
		if !inFence {
			continue
		}
		for _, m := range numericFenceKeyValue.FindAllStringSubmatch(line, -1) {
			tokensExtracted++
			keys[m[1]] = true
		}
	}
	return keys, tokensExtracted
}

// TestNumericFieldClassificationInventory is AC29's production
// assertion, over the field names folio-format.md ACTUALLY documents
// with a numeric literal — derived at test time via
// extractNumericDocKeys, never hand-copied (Finding 3; M-3's 11-key
// count is retained only as a sanity floor, not as the source of
// truth). It reports how many fields it considered and fails on zero
// (D-000.9 shape) — a coverage witness that would pass identically
// whether the registry was checked or not is exactly what this guards
// against.
func TestNumericFieldClassificationInventory(t *testing.T) {
	root := repoRootFromTest(t)
	docBytes := mustReadFile(t, filepath.Join(root, "_bmad-output", "specs", "spec-folio", "folio-format.md"))
	documented, tokensExtracted := extractNumericDocKeys(docBytes)
	if tokensExtracted == 0 {
		t.Fatal("coverage witness: zero numeric key/value tokens extracted from folio-format.md")
	}
	if len(documented) < 11 {
		t.Fatalf("M-3 measured at least 11 distinct numeric keys in folio-format.md; extraction found only %d — extractNumericDocKeys may have regressed", len(documented))
	}

	var names []string
	for k := range documented {
		names = append(names, k)
	}
	sort.Strings(names)

	unclassified, considered := classifyNumericFields(names, numericFieldRegistry)
	if considered == 0 {
		t.Fatal("coverage witness: zero numeric fields considered")
	}
	if len(unclassified) > 0 {
		t.Fatalf("fields folio-format.md documents with a numeric literal but numericFieldRegistry does not classify: %v", unclassified)
	}
	for _, f := range names {
		if numericFieldRegistry[f] != numericKindPoints && numericFieldRegistry[f] != numericKindPlainInt {
			t.Fatalf("field %q classified into neither kind 1 nor kind 2", f)
		}
	}
}

// TestNumericFieldClassificationCatchesAnEscapee is AC29's red-proof, by
// construction: a retained fixture list carrying a field the registry
// does not know about (simulating folio-format.md growing a new
// numeric field before this inventory is updated) must be reported as
// unclassified — proving the checker can actually fail, not just pass
// vacuously on every input.
func TestNumericFieldClassificationCatchesAnEscapee(t *testing.T) {
	fixture := []string{"width", "height", "futureNumericFieldNotYetClassified"}
	unclassified, considered := classifyNumericFields(fixture, numericFieldRegistry)
	if considered != len(fixture) {
		t.Fatalf("considered = %d, want %d", considered, len(fixture))
	}
	if len(unclassified) != 1 || unclassified[0] != "futureNumericFieldNotYetClassified" {
		t.Fatalf("expected exactly one escapee to be caught, got %v", unclassified)
	}
}
