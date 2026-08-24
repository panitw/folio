package folio_test

// Story 2.2, AC4: the missing-glyph diagnostic is COVERAGE-based,
// evaluated per rune against each face's cmap in chain order — never a
// proxy such as locale or "preferred face for the script" (D-2.2-D4).
// These tests exercise the REAL public Render path against the REAL
// shipped faces (fonts.Shipped()), not a synthetic fixture, so a
// regression here is a regression an integrator would actually hit.

import (
	"strings"
	"testing"

	folio "github.com/panitw/folio/folio-go"
	"github.com/panitw/folio/folio-go/fonts"
)

const multiScriptTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 500, "height": 20, "value": "{{name}}", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {
    "body": ["Noto Sans", "Noto Sans Thai", "Noto Sans SC"]
  },
  "locale": "en",
  "nextId": 2,
  "page": {
    "margin": {"bottom": 36, "left": 36, "right": 36, "top": 36},
    "orientation": "portrait",
    "size": "A4"
  },
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// TestCoverageBasedFallbackSpansAllThreeShippedFaces is AC7/AC8's own
// premise made concrete: a SINGLE text element mixing Latin, Thai and
// CJK runes must actually embed all three shipped faces, each covering
// only the runes it has glyphs for — never "the whole element uses
// whichever chain member came first" (the pre-Story-2.2 behaviour).
func TestCoverageBasedFallbackSpansAllThreeShippedFaces(t *testing.T) {
	tpl, err := folio.ParseTemplate([]byte(multiScriptTemplateJSON))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	// "Ada" (Latin, Noto Sans) + "ก" (Thai, Noto Sans Thai) + "汉" (a
	// Han ideograph, Noto Sans SC) — one line genuinely mixing scripts,
	// exactly AD-8's "a template names a family plus an ordered
	// fallback chain, tried left to right per glyph" scenario.
	data := folio.Data(`{"name": "Ada ก 汉"}`)
	pdfBytes, err := folio.Render(tpl, data, folio.Params(`{}`), fonts.Shipped())
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	// V6's vacuity guard, applied to all three faces at once: a render
	// that silently dropped to one face (or produced tofu) would still
	// "succeed" — so assert the resource NAMES for all three faces
	// actually appear as Font resources, not just that /FontFile2
	// appears at all (which one face alone would also satisfy).
	// Resource/BaseFont names have spaces stripped (measured: "Noto
	// Sans" -> "NotoSans" in the Resources dict and /BaseFont), so
	// match on the sanitized form.
	s := string(pdfBytes)
	for _, want := range []string{"NotoSans", "NotoSansThai", "NotoSansSC"} {
		if !strings.Contains(s, want) {
			t.Errorf("rendered PDF does not reference face %q as a resource/BaseFont name — the fallback chain did not actually reach it", want)
		}
	}
	if n := strings.Count(s, "/FontFile2"); n < 3 {
		t.Errorf("expected at least 3 distinct /FontFile2 embeddings (one per shipped face actually used), got %d", n)
	}
}

// TestMissingGlyphDiagnosticFiresOnUncoveredRune is V8: the fixture
// must contain a rune genuinely covered by NO face in the chain, and
// the failure must name both the element id and the offending rune —
// never a silently emitted blank box.
func TestMissingGlyphDiagnosticFiresOnUncoveredRune(t *testing.T) {
	const singleFaceTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 500, "height": 20, "value": "{{name}}", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	tpl, err := folio.ParseTemplate([]byte(singleFaceTemplateJSON))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	// "ก" (Thai) has no glyph in Noto Sans, and the chain has no other
	// member to fall back to — this MUST fail, never silently render a
	// blank box in its place.
	data := folio.Data(`{"name": "ก"}`)
	_, err = folio.Render(tpl, data, folio.Params(`{}`), fonts.Shipped())
	if err == nil {
		t.Fatal("expected an error for a rune with no coverage in any chain member; render succeeded instead")
	}
	if !strings.Contains(err.Error(), "e1") {
		t.Errorf("error does not name the element id (e1): %v", err)
	}
	if !strings.Contains(err.Error(), `U+0E01`) {
		t.Errorf("error does not name the offending rune (U+0E01, ก): %v", err)
	}
}

// TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic is V9's other
// direction (D-2.2-D4): Noto Sans SC is Pan-CJK, so Japanese text HAS
// coverage (kana + the shared ideograph set), and the diagnostic must
// NOT fire for it — the `ja` gap is glyph FORM quality (AC10), never a
// coverage failure, and conflating the two makes a diagnostic fire on a
// correct render.
func TestJapaneseTextThroughPanCJKFaceDoesNotFireDiagnostic(t *testing.T) {
	const jaTemplateJSON = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 500, "height": 20, "value": "{{greeting}}", "style": {"fontFamily": "body", "fontSize": 14}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans SC"]},
  "locale": "ja",
  "nextId": 2,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`
	tpl, err := folio.ParseTemplate([]byte(jaTemplateJSON))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	// Hiragana "こんにちは" ("hello") — genuinely Japanese, covered by
	// Noto Sans SC's shared kana + ideograph set, never routed through
	// any "is this face preferred for ja" check (D-2.2-D4 forbids that
	// proxy outright).
	data := folio.Data(`{"greeting": "こんにちは"}`)
	if _, err := folio.Render(tpl, data, folio.Params(`{}`), fonts.Shipped()); err != nil {
		t.Fatalf("Japanese text through the Pan-CJK Noto Sans SC face must render without a missing-glyph diagnostic (D-2.2-D4): %v", err)
	}
}
