package folio

// Story 2.6's AC8: {{page}} and {{pages}} are NOT implemented, and this test
// makes implementing them fail.
//
// THE HAZARD IS NOT THE EPIC TEXT. epics.md §2.6 is silent on page numbers,
// checked line by line, and folio-format.md says normatively that Story 2.7
// owns them (D-1.6.5, AC18). The hazard is that a developer building
// pagination HAS THE PAGE INDEX IN HAND — layout.Pagination literally
// enumerates the pages — and substituting it into a footer string is ONE
// LINE. This test exists solely to make that one line fail.
//
// D-000.34 (extended) is why this is worth writing DOWN rather than trusting:
// the reservation is currently held only by internal/bind, which pagination
// does not touch, so nothing today would notice the one line being added in
// package folio instead. A negative control dies invisibly, because a control
// passing is what a control does.
//
// WHY IT ASSERTS OFF THE PRODUCED CONTENT STREAM AND NOT OFF THE INPUT.
// D-000.21 sharpened: assert on the artifact that carries the property. The
// property is what a READER of the PDF sees, so the literal characters are
// recovered from the drawn glyphs through the document's own /ToUnicode CMap.
// Asserting that the input still contains "{{page}}" would pass on a renderer
// that substituted it on the way out.

import (
	"strings"
	"testing"
)

// reservedPlaceholderFooterTemplate is the multi-page fixture's geometry and
// content with ONE change: the page footer's value carries both reserved
// placeholders, in the shape a template author would naturally write.
//
// It is a SEPARATE document rather than an edit to fixtures/multi-page/ so
// that the committed golden is not re-recorded when Story 2.7 implements
// these placeholders — at which point this document's rendering legitimately
// changes and this test legitimately becomes 2.7's to rewrite.
const reservedPlaceholderFooterTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 480, "height": 700, "value": "This statement covers the period from the first of January to the thirty-first of March. Every page of it carries the same running header and the same running footer, so a reader handed only the second sheet can still tell what they are holding and who issued it. The content below is one single text element. It is not divided into paragraphs, sections or boxes, and nothing in it has been repositioned in order to make it fit. It simply runs on until it reaches the bottom of the content band, and then it continues at the top of the next page, at exactly the leading it had before. No line is ever split across the boundary between two pages. A line that would otherwise fall half on one sheet and half on the next is placed whole on the next sheet instead, and the small band of whitespace left behind at the foot of the earlier page is correct typesetting rather than a defect. A bank statement cannot ship a half line. That single rule is what makes this document safe to print, safe to file, and safe to read aloud from. Nothing on any page negotiates for space with anything else, and no element ever moves sideways or shuffles upward to close a gap that pagination opened.", "style": {"fontFamily": "body", "fontSize": 24}}
      ]
    },
    "pageFooter": {
      "elements": [
        {"id": "e2", "type": "text", "x": 0, "y": 6, "width": 480, "height": 16, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 8}}
      ],
      "height": 24
    },
    "pageHeader": {
      "elements": [
        {"id": "e3", "type": "text", "x": 0, "y": 4, "width": 480, "height": 16, "value": "HEADER REPEATED ON EVERY PAGE", "style": {"fontFamily": "body", "fontSize": 9}}
      ],
      "height": 18
    }
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 42, "left": 36, "right": 54, "top": 30}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.0"
}
`

// reservedFooterLiteral is what a reader must see, character for character,
// on EVERY page: the placeholders spelled out, unresolved.
const reservedFooterLiteral = "Page {{page}} of {{pages}}"

// TestReservedPagePlaceholdersPassThroughOnEveryPage is AC8.
//
// PRESENCE PRECONDITION: the document must render as N >= 2 pages before the
// assertion means anything. On a one-page document "the literal appears on
// every page" is a claim about a single page and says nothing about the
// hazard, which is specifically that pagination hands you a page INDEX.
func TestReservedPagePlaceholdersPassThroughOnEveryPage(t *testing.T) {
	tpl, err := ParseTemplate([]byte(reservedPlaceholderFooterTemplate))
	if err != nil {
		t.Fatalf("presence precondition: the template does not parse: %v", err)
	}

	// The reserved tokens really are in the INPUT — otherwise the assertion
	// below is about a document that never contained them.
	if !strings.Contains(reservedPlaceholderFooterTemplate, "{{page}}") ||
		!strings.Contains(reservedPlaceholderFooterTemplate, "{{pages}}") {
		t.Fatal("presence precondition: the template carries neither reserved placeholder")
	}

	b, rerr := Render(tpl, Data("{}"), nil, testShippedFontSet())
	if rerr != nil {
		t.Fatalf("Render: %v — a reserved placeholder must PASS THROUGH, never be an error", rerr)
	}

	streams := splitPageContentStreams(t, b)
	if len(streams) < 2 {
		t.Fatalf("presence precondition: the document rendered as %d page(s); this assertion needs at least 2, because the hazard it guards is that PAGINATION puts a page index in the implementer's hand", len(streams))
	}

	cmap := mpParseToUnicode(t, b)
	for p, stream := range streams {
		found := false
		for _, run := range mpExtractRuns(t, stream, cmap) {
			if run.text == reservedFooterLiteral {
				found = true
				break
			}
			// A substituted footer is the exact failure this guards. Name it
			// specifically rather than letting it fall through to "not found",
			// so the message tells the next reader what happened.
			if strings.HasPrefix(run.text, "Page ") && !strings.Contains(run.text, "{{") {
				t.Errorf("page %d draws the footer as %q — the reserved placeholders have been RESOLVED.\n\n"+
					"{{page}} and {{pages}} are owned by STORY 2.7 (D-1.6.5, AC18). They pass through unchanged "+
					"and are never resolved from data. Story 2.6 paginates, which is exactly what puts the page "+
					"index within reach — and implementing the substitution early would quietly foreclose 2.7's "+
					"design, because the choice of one-based vs zero-based, of where the count is computed, and "+
					"of whether a second pass is needed would all have been made by accident.",
					p+1, run.text)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("page %d does not draw the literal %q anywhere in its content stream", p+1, reservedFooterLiteral)
		}
	}
}

// TestReservedPlaceholderSetIsUnchanged pins the reservation itself.
//
// The pass-through test above asserts the BEHAVIOUR on one document. This
// asserts the SET, so that adding a third reserved token — or removing one of
// these two — is a visible, deliberate edit rather than a side effect.
// internal/bind is a different package, so this reaches the property through
// the behaviour that expresses it rather than through the variable.
func TestReservedPlaceholderSetIsUnchanged(t *testing.T) {
	for _, token := range []string{"{{page}}", "{{pages}}"} {
		src := strings.Replace(reservedPlaceholderFooterTemplate,
			"Page {{page}} of {{pages}}", "["+token+"]", 1)
		tpl, err := ParseTemplate([]byte(src))
		if err != nil {
			t.Fatalf("%s: template does not parse: %v", token, err)
		}
		b, rerr := Render(tpl, Data("{}"), nil, testShippedFontSet())
		if rerr != nil {
			t.Errorf("%s: Render failed (%v) — a RESERVED token must pass through, not error", token, rerr)
			continue
		}
		cmap := mpParseToUnicode(t, b)
		streams := splitPageContentStreams(t, b)
		found := false
		for _, stream := range streams {
			for _, run := range mpExtractRuns(t, stream, cmap) {
				if run.text == "["+token+"]" {
					found = true
				}
			}
		}
		if !found {
			t.Errorf("%s: the literal %q is not drawn anywhere — it is no longer passing through unchanged", token, "["+token+"]")
		}
	}

	// The NEGATIVE half: a token that is NOT reserved must still be resolved
	// (or error) rather than passing through. Without this, a change that made
	// EVERY placeholder pass through unchanged would leave the assertions
	// above green while silently disabling all data binding.
	src := strings.Replace(reservedPlaceholderFooterTemplate,
		"Page {{page}} of {{pages}}", "{{notreserved}}", 1)
	tpl, err := ParseTemplate([]byte(src))
	if err != nil {
		t.Fatalf("template does not parse: %v", err)
	}
	if _, rerr := Render(tpl, Data("{}"), nil, testShippedFontSet()); rerr == nil {
		t.Error("a NON-reserved placeholder {{notreserved}} rendered without error against empty data — " +
			"if every placeholder now passes through unchanged, the two assertions above are vacuous and " +
			"data binding is silently disabled")
	}
}
