package folio

import (
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/template"
)

// This file pins `chainFaceNames` (render.go) — THE one boundary between the
// document's chain and the render path's face-name list, and the line at which
// Story 8.3 stops.
//
// IT LIVES HERE, IN `package folio`, BECAUSE NOWHERE ELSE CAN. The function's
// consequences are consequences at Render and Validate, and both are exported
// from this package; internal/template is a different package that cannot call
// either, so its tests can pin the FORMAT's behaviour and nothing about the
// render path's. Story 8.3's first draft of the comment on chainFaceNames
// claimed these properties were "pinned by test (fonts_embedded_test.go)" —
// they were not, and could not have been. Deleting the `Embedded()` filter
// outright left the whole Go suite green.
//
// WHAT IS PINNED, and each arm is red-proved by mutation:
//   - embedded entries are DROPPED, and the surviving order is the document's;
//   - a chain left with no usable entry produces the located error the
//     empty-metrics path already produced, at Render AND at Validate.
//
// TWO MUTATIONS, because they are caught in different places and knowing which
// is which is the difference between coverage and the feeling of it:
//   - DELETING the `Embedded()` filter appends an empty face name. The unit
//     half below catches it on length and order; the end-to-end halves do NOT,
//     because an empty name is absent from the FontSet and is skipped exactly
//     as the entry was.
//   - Emitting `entry.AssetKey` INSTEAD of skipping — the defect that actually
//     matters, a 64-character digest reaching the render path as a face name —
//     is caught by both halves, including the located error's own text.

// embeddedChainDoc is a one-text-element document whose `body` chain is
// `chain` verbatim, carrying the same font asset fixtures/embedded-font/ does
// so an `{"asset": …}` entry has something real to name.
func embeddedChainDoc(t *testing.T, chain string) string {
	t.Helper()
	source := embeddedFontTemplateJSON()
	// Spliced between the two keys that BRACKET the fonts block in canonical
	// (sorted) order — `bands` before it, `locale` after it — rather than by
	// counting braces. Brace counting over a document carrying ~47 KB of
	// base64 is a second JSON parser written badly; the sort order is a
	// property AD-9 guarantees.
	const open, next = "  \"fonts\": {", "  \"locale\":"
	start := strings.Index(source, open)
	end := strings.Index(source, next)
	if start < 0 || end < 0 || end < start {
		t.Fatalf("fixture assumption violated: the generated document's fonts block is not bracketed by %q and %q", open, next)
	}
	return source[:start] + "  \"fonts\": {\"body\": " + chain + "},\n" + source[end:]
}

// TestChainFaceNamesDropsEmbeddedEntriesAndKeepsOrder is the unit-level half:
// the function itself, over every mixture, asserting BOTH that embedded
// entries go and that nothing else moves.
//
// The order assertion is not decoration. A filter written as "collect the
// named faces" is trivially order-preserving; one written as "partition, then
// concatenate" is not, and the two are indistinguishable from a chain whose
// embedded entry happens to sit last. Every case here puts one somewhere else.
func TestChainFaceNamesDropsEmbeddedEntriesAndKeepsOrder(t *testing.T) {
	face := template.FaceEntry
	asset := template.AssetEntry
	for _, tc := range []struct {
		name  string
		chain []template.FontChainEntry
		want  []string
	}{
		{"no entries", nil, []string{}},
		{"named faces only", []template.FontChainEntry{face("A"), face("B"), face("C")}, []string{"A", "B", "C"}},
		{"embedded first", []template.FontChainEntry{asset("k1"), face("A"), face("B")}, []string{"A", "B"}},
		{"embedded in the middle", []template.FontChainEntry{face("A"), asset("k1"), face("B")}, []string{"A", "B"}},
		{"embedded last", []template.FontChainEntry{face("A"), face("B"), asset("k1")}, []string{"A", "B"}},
		{"two embedded, interleaved", []template.FontChainEntry{asset("k1"), face("A"), asset("k2"), face("B")}, []string{"A", "B"}},
		{"embedded only", []template.FontChainEntry{asset("k1"), asset("k2")}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := chainFaceNames(tc.chain)
			if len(got) != len(tc.want) {
				t.Fatalf("chainFaceNames(%v) = %v, want %v", tc.chain, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("chainFaceNames(%v) = %v, want %v — the document's authored order is not preserved", tc.chain, got, tc.want)
				}
			}
			// The asset key must NEVER appear as a face name. If it did, the
			// render path would look up a 64-character digest in the FontSet,
			// find nothing, and warn on every rune rather than skipping.
			for _, name := range got {
				if name == "k1" || name == "k2" {
					t.Fatalf("an embedded entry's asset key reached the render path as a face name: %v", got)
				}
			}
		})
	}
}

// TestEmbeddedEntryIsSkippedAtRender is the end-to-end half of the same claim:
// the document renders, and it renders EXACTLY as the same document without the
// embedded entry does. Byte identity is the assertion, because "it rendered" is
// satisfied by a render that quietly did something else with the carried face.
func TestEmbeddedEntryIsSkippedAtRender(t *testing.T) {
	withEmbedded := embeddedChainDoc(t, `["Noto Sans", {"asset": "`+embeddedFontAssetKey()+`"}]`)
	withoutEmbedded := embeddedChainDoc(t, `["Noto Sans"]`)

	render := func(source string) []byte {
		t.Helper()
		tpl, err := ParseTemplate([]byte(source))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if len(res.Diagnostics) != 0 {
			t.Fatalf("render produced diagnostics: %+v", res.Diagnostics)
		}
		return res.Bytes
	}

	a, b := render(withEmbedded), render(withoutEmbedded)
	if string(a) != string(b) {
		t.Fatalf("a chain carrying an embedded entry rendered %d bytes and the same chain without it rendered %d — an embedded entry must contribute nothing until Story 8.4", len(a), len(b))
	}
}

// TestEmbeddedEntryBeforeTheShippedFaceStillResolves is the ORDER half at the
// render surface. An embedded entry sitting FIRST must not shadow the named
// face behind it: the chain still resolves, and to the same bytes.
//
// It is a separate case because a filter that dropped everything AFTER the
// first embedded entry — a plausible off-by-one — would pass the test above,
// where the embedded entry is last.
func TestEmbeddedEntryBeforeTheShippedFaceStillResolves(t *testing.T) {
	leading := embeddedChainDoc(t, `[{"asset": "`+embeddedFontAssetKey()+`"}, "Noto Sans"]`)
	plain := embeddedChainDoc(t, `["Noto Sans"]`)

	render := func(source string) []byte {
		t.Helper()
		tpl, err := ParseTemplate([]byte(source))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		res, err := Render(tpl, Data(`{}`), nil, testShippedFontSet())
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		return res.Bytes
	}
	if string(render(leading)) != string(render(plain)) {
		t.Fatal("an embedded entry ahead of a named face changed the render — it must be skipped, not allowed to shadow what follows it")
	}
}

// TestAllEmbeddedChainProducesTheExistingLocatedError is the OTHER consequence,
// and it is the one the interim state actually turns on: after Story 8.3 a
// chain may legally hold an entry nothing can draw, and a chain of NOTHING BUT
// such entries therefore has no usable face at all.
//
// THE BEHAVIOUR IS THE PRE-EXISTING ONE, DELIBERATELY NOT WIDENED. It is the
// same located error the empty-metrics path already produced for a chain whose
// every named face was absent from the FontSet — element id, and the chain it
// searched. Nothing new was minted for this, and the assertion is here so that
// if a later story does widen it, the change is deliberate.
//
// Validate is asserted alongside Render, because Validate PREDICTS Render
// (D-1.8.1 amended) and the two returning different answers about the same
// document is exactly the second rule system that ruling forbids.
func TestAllEmbeddedChainProducesTheExistingLocatedError(t *testing.T) {
	source := embeddedChainDoc(t, `[{"asset": "`+embeddedFontAssetKey()+`"}]`)

	// The document LOADS. It is a valid `.folio`; only drawing it fails.
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("a chain of only embedded entries must LOAD — the format can express it: %v", err)
	}

	_, rerr := Render(tpl, Data(`{}`), nil, testShippedFontSet())
	if rerr == nil {
		t.Fatal("a chain with no usable entry must fail at Render — nothing can draw the text")
	}
	for _, want := range []string{"element e1", "none of the fallback chain's faces"} {
		if !strings.Contains(rerr.Error(), want) {
			t.Errorf("the render error does not locate itself (%q missing): %v", want, rerr)
		}
	}
	// The asset key must NOT appear in the message. If it did, the embedded
	// entry would have reached the chain as a face name — the defect the
	// filter exists to prevent, reported as a diagnostic that reads as though
	// the author mistyped a font name.
	if strings.Contains(rerr.Error(), embeddedFontAssetKey()) {
		t.Errorf("the render error names the asset key as a searched face: %v", rerr)
	}

	diags, verr := Validate([]byte(source), Data(`{}`), nil, testShippedFontSet())
	if verr == nil {
		t.Fatal("Validate must PREDICT Render (D-1.8.1 amended): Render refuses this document, so Validate must too")
	}
	if verr.Error() != rerr.Error() {
		t.Errorf("Validate and Render disagree about the same document:\n\tValidate: %v\n\tRender:   %v", verr, rerr)
	}
	if len(diags) != 0 {
		t.Errorf("Validate returned %d diagnostic(s) alongside a hard error: %+v", len(diags), diags)
	}
}
