package folio

import (
	"testing"

	"github.com/panitw/folio/folio-go/internal/geom"
)

// TestCanvasMeasuresWithTheEmbeddedFace is Story 8.4's AC5, stated as VERIFIED
// rather than deleted (D-8.4.1a).
//
// WHY THIS IS A TEST AND NOT A FEATURE. addCanvasTextPaint (page_setup.go)
// already calls the render path's own fontChain, shapeSegments,
// chainVerticalModel and positionSegments — AD-17's "the canvas consumes the
// IDENTICAL advance the renderer does … the browser never measures text". So
// the moment the render path can resolve an embedded entry, the canvas can
// too, and the measurement half of AC5 arrives with the seam rather than with
// any canvas code of its own.
//
// THAT IS EXACTLY WHY IT NEEDS A PIN. "Shared today" and "shared tomorrow" are
// different claims, and the difference is one refactor wide. There is ONE
// thing the two paths do not share — each builds its own fontCache
// (predictDocument at render.go, addCanvasTextPaint at page_setup.go) — and if
// the canvas's were built without the document, every assertion below would
// fail while the PDF stayed perfect. That is the mutation this test is
// red-proved against: replace newDocumentFontCache(t) with newFontCache() in
// page_setup.go and this reddens on its own.
//
// WHAT IS NOT CLAIMED HERE. The canvas MEASURES with the embedded face; the
// browser cannot PAINT with it, because the designer has no CSS family for a
// carried face at all and falls through to `sans-serif`. That is Story 8.4a
// (DW-35), and the disclosure is carried by
// folio-designer/src/canvas-font-stack.test.ts rather than by a comment.
func TestCanvasMeasuresWithTheEmbeddedFace(t *testing.T) {
	tpl, err := ParseTemplate([]byte(embeddedFontTemplateJSON()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	fs := testShippedFontSet()

	projection, err := CanvasWithTextPaint(tpl, fs)
	if err != nil {
		t.Fatalf("canvas projection: %v", err)
	}
	// A DEGRADED CHAIN IS HOW THIS FAILS SILENTLY, and it has to be excluded
	// before anything below means anything. addCanvasTextPaint disposes of an
	// element it cannot shape — an empty paint, and on carries — so a canvas
	// that could not see the carried face returns a perfectly well-formed
	// projection with nothing in it, and a comparison over zero fragments
	// passes. Two independent witnesses, because they fail differently:
	// ContentWindowCountIsExact is the ENGINE's own report that no chain
	// degraded (page_setup.go's `exact` term; this document has no bound table
	// and no conditional visibility, so a degraded chain is the only cause
	// that can be present), and the fragment count below is the fact itself.
	if !projection.ContentWindowCountIsExact {
		t.Fatal("the engine reports an inexact window count for a document with no table and no conditional visibility — its font chain degraded, so the embedded entry did not reach the canvas's own fontCache")
	}

	data := emptyBindValue(t)
	runs, err := collectTextRuns(tpl, data, data, fs, newDocumentFontCache(tpl))
	if err != nil {
		t.Fatalf("shipping run collection: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("presence precondition: the PDF path produced no text runs, so nothing below is compared")
	}

	// THE WITNESS THAT THIS DOCUMENT EXERCISES THE SUBJECT AT ALL. Every run
	// must be drawn with the face the document CARRIES — not the shipped Latin
	// face its chain names first, which covers none of this Thai. Without this
	// the whole test would keep passing over a document whose embedded entry
	// was never consulted, which is the vacuous version of it.
	wantFace := embeddedFaceName(embeddedFontAssetKey())
	for _, run := range runs {
		if run.face != wantFace {
			t.Fatalf("a run is drawn with face %q, want the carried face %q — this document is not exercising the embedded path", run.face, wantFace)
		}
	}

	byElement := make(map[string][]textRunSource)
	for _, run := range runs {
		byElement[run.elementID] = append(byElement[run.elementID], run)
	}
	bandY := make(map[string]int64)
	for _, band := range projection.Bands {
		bandY[band.Name] = band.Y
	}

	// The comparison is canvas_text_paint_test.go's own — term for term, not a
	// looser restatement of it — so the two tests cannot drift into asserting
	// different strengths of the same claim.
	cache := newDocumentFontCache(tpl)
	compared := 0
	for _, component := range projection.Components {
		if component.Type != "text" || component.TextPaint == nil {
			continue
		}
		wantRuns := byElement[component.ID]
		if len(wantRuns) == 0 {
			t.Fatalf("component %s has paint but no PDF-path runs", component.ID)
		}
		for lineIndex, line := range component.TextPaint.Lines {
			var lineRuns []textRunSource
			for _, run := range wantRuns {
				if run.lineIndex == lineIndex {
					lineRuns = append(lineRuns, run)
				}
			}
			if len(lineRuns) != len(line.Fragments) {
				t.Fatalf("component %s line %d has %d canvas fragments and %d PDF runs", component.ID, lineIndex, len(line.Fragments), len(lineRuns))
			}
			var width geom.Length
			for fragmentIndex, run := range lineRuns {
				runWidth, werr := shippingRunWidth(run, fs, cache)
				if werr != nil {
					t.Fatal(werr)
				}
				width += runWidth
				bandOrigin := bandY[component.Band] - projection.MarginTop
				if line.Top != int64(run.y)-bandOrigin || line.Baseline != int64(run.y+run.baselineOffset)-bandOrigin {
					t.Errorf("component %s line %d origin/baseline diverges from the PDF path: %#v versus y=%d baseline=%d", component.ID, lineIndex, line, run.y, run.y+run.baselineOffset)
				}
				fragment := line.Fragments[fragmentIndex]
				if fragment.Text != run.text || fragment.X != int64(run.x) {
					t.Errorf("component %s line %d fragment %d = %#v, want the PDF path's text=%q x=%d", component.ID, lineIndex, fragmentIndex, fragment, run.text, run.x)
				}
				compared++
			}
			// The ADVANCE half. A fragment origin that matched while the line
			// advance did not would still stack the second line wrongly, and
			// the advance is what the embedded face's own hhea metrics feed.
			if line.Width != int64(width) {
				t.Errorf("component %s line %d width = %d, want the PDF path's shaped width %d", component.ID, lineIndex, line.Width, width)
			}
			if line.Advance <= 0 {
				t.Errorf("component %s line %d advance = %d — the vertical model derived nothing from the carried face", component.ID, lineIndex, line.Advance)
			}
		}
	}
	if compared == 0 {
		t.Fatal("vacuity guard: no fragment was compared, so this test asserted nothing")
	}
	t.Logf("canvas/PDF measurement agreement witness — %d fragments compared, all drawn with the carried face %q", compared, wantFace)
}

// TestCanvasVerticalModelUsesTheEmbeddedFacesOwnMetrics is the half the
// fragment comparison above cannot see on a one-line document: that the line
// advance the canvas reports is derived from the CARRIED face's hhea metrics
// and not from the shipped face the chain names first.
//
// The two faces have different vertical metrics, so a canvas that resolved
// only the shipped entry would report a different advance — and on a
// single-line element nothing else would move, which is exactly how this would
// otherwise ship unnoticed.
func TestCanvasVerticalModelUsesTheEmbeddedFacesOwnMetrics(t *testing.T) {
	tpl, err := ParseTemplate([]byte(embeddedFontTemplateJSON()))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	fs := testShippedFontSet()
	chain := []string{"Noto Sans", embeddedFaceName(embeddedFontAssetKey())}
	const fontSize = geom.Length(12_000)

	withCarried, err := chainVerticalModel(chain, fontSize, defaultLineSpacing, fs, newDocumentFontCache(tpl))
	if err != nil {
		t.Fatalf("vertical model over the carried chain: %v", err)
	}
	// The same chain against a cache that cannot see the document: the
	// embedded entry supplies nothing and only "Noto Sans" constrains the
	// model. This is the pre-Story-8.4 answer, and the two must differ or the
	// assertion above is measuring the shipped face under another name.
	shippedOnly, err := chainVerticalModel(chain, fontSize, defaultLineSpacing, fs, newFontCache())
	if err != nil {
		t.Fatalf("vertical model over the shipped entry alone: %v", err)
	}
	if withCarried == shippedOnly {
		t.Fatalf("the carried face contributed nothing to the vertical model: %#v", withCarried)
	}

	projection, err := CanvasWithTextPaint(tpl, fs)
	if err != nil {
		t.Fatalf("canvas projection: %v", err)
	}
	found := false
	for _, component := range projection.Components {
		if component.TextPaint == nil {
			continue
		}
		for _, line := range component.TextPaint.Lines {
			found = true
			if line.Advance != int64(withCarried.Advance) {
				t.Errorf("canvas line advance = %d, want %d — the canvas's vertical model does not include the carried face (it reports %d without it)",
					line.Advance, withCarried.Advance, shippedOnly.Advance)
			}
		}
	}
	if !found {
		t.Fatal("vacuity guard: the projection carries no painted line, so nothing was compared")
	}
}

// TestCanvasDegradesRatherThanAbortingOnANonFontChainEntry pins D-7.4.2 on the
// element arm Story 8.4 made reachable.
//
// THE CONDITION IS NEW AND IT COMES FROM DOCUMENT CONTENT. Before this story a
// chain entry naming a non-font asset resolved to nothing anywhere; since it,
// coverage resolution refuses it, located — which is DW-83's whole point on
// the render path, and exactly the wrong answer on the canvas. The designer is
// the ONE surface on which an author can repair such an entry, and a
// projection that returns an error opens no document at all.
//
// So the capability refusal is tolerated HERE, per element, in the shape
// addCanvasTextPaint's own fontChain arm a few lines above already uses: an
// empty paint, the column flagged as degraded, and on to the next element. It
// is scoped to the capability error (template.UnsupportedFontMediaTypeError)
// so a genuine internal shaping fault still aborts the projection as it did.
//
// RED-PROOF: restore the unconditional `return` in addCanvasTextPaint's
// shapeSegments arm and this test fails on the projection error.
func TestCanvasDegradesRatherThanAbortingOnANonFontChainEntry(t *testing.T) {
	source := nonFontChainDoc(t, `["Noto Sans", {"asset": "`+nonFontAssetKey+`"}]`, "สัญญา")
	tpl, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatalf("the document must LOAD: %v", err)
	}
	fs := testShippedFontSet()

	// The geometry-only projection never shaped anything and always worked;
	// asserting it here is what makes the paint half's failure specific.
	if _, gerr := Canvas(tpl); gerr != nil {
		t.Fatalf("the geometry projection must succeed: %v", gerr)
	}

	projection, err := CanvasWithTextPaint(tpl, fs)
	if err != nil {
		t.Fatalf("a document the FORMAT calls valid must still open in the designer — the surface on which its chain can be repaired: %v", err)
	}
	found := false
	for _, component := range projection.Components {
		if component.ID != "e1" {
			continue
		}
		found = true
		if component.TextPaint == nil {
			t.Fatal("the element carries no paint at all — it must carry an EMPTY paint, which is what the canvas draws as nothing")
		}
		if len(component.TextPaint.Lines) != 0 {
			t.Errorf("the element painted %d line(s) from a chain entry the engine refuses to draw with", len(component.TextPaint.Lines))
		}
	}
	if !found {
		t.Fatal("vacuity guard: the projection carries no component e1")
	}
	// AND IT SAYS SO. A silently-degraded element and a well-measured one are
	// indistinguishable downstream, which is why the honesty flag exists.
	if projection.ContentWindowCountIsExact {
		t.Error("the projection reports an EXACT window count for a document whose only text element could not be shaped — the degradation is not disclosed")
	}

	// The render path is unchanged: this document is still refused there,
	// located. The canvas tolerance is a canvas decision and must not have
	// leaked into Render.
	if _, rerr := Render(tpl, Data(`{}`), nil, fs); rerr == nil {
		t.Error("Render must still refuse this document (DW-83) — the canvas's tolerance leaked onto the render path")
	}
}
