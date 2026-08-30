package folio

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/panitw/folio/folio-go/internal/layout"
	"github.com/panitw/folio/folio-go/internal/template"
)

// --- Story 7.7's discriminating fixture (FR51) ----------------------------
//
// Every assertion below is about the DIFFERENCE between two documents that
// differ in exactly one respect — three `keepTogether` tags — because
// "the block is on page 2" is a fact any number of unrelated
// implementations produce, and "the block is on page 2 HERE and split
// THERE" is not.

// keepTogetherPages renders one of the pair and returns its page model,
// insisting the document builds cleanly.
func keepTogetherPages(t *testing.T, templateJSON string) []pageProbe {
	t.Helper()
	pages, diags := keepTogetherPagesWithDiagnostics(t, templateJSON)
	if len(diags) != 0 {
		t.Fatalf("the keep-together pair must build with NO diagnostics; got %+v", diags)
	}
	return pages
}

// keepTogetherPagesWithDiagnostics is the same probe for a document that
// is SUPPOSED to say something — the over-tall cases, which return a
// Warning beside their bytes and would otherwise be unprobeable here.
func keepTogetherPagesWithDiagnostics(t *testing.T, templateJSON string) ([]pageProbe, []Diagnostic) {
	t.Helper()
	tpl, err := ParseTemplate([]byte(templateJSON))
	if err != nil {
		t.Fatalf("parse keep-together template: %v", err)
	}
	pages, _, _, diags, err := buildPageModel(tpl, mustDecodeData(t, keepTogetherDataJSON), mustDecodeParams(t), testShippedFontSet())
	if err != nil {
		t.Fatalf("build keep-together page model: %v", err)
	}
	out := make([]pageProbe, 0, len(pages))
	for _, p := range pages {
		probe := pageProbe{rects: len(p.Rects), images: len(p.Images)}
		for _, r := range p.Runs {
			probe.text += r.SourceText
		}
		out = append(out, probe)
	}
	return out, diags
}

// pageProbe is one page reduced to what these assertions are about: the
// text it carries, how many rectangles it draws and how many images it
// places. The signature block's members are a text, a rect, an image and
// a second text, so all three fields are load-bearing — a check that
// looked only at runs could see neither the ruled line nor a signature
// image move.
type pageProbe struct {
	text   string
	rects  int
	images int
}

// The three signature members, named by what a reader would look for on
// the page. e3, the ruled line, has no text at all and is counted
// through the page's rectangles instead.
const (
	keepTogetherSignatureText = "Signed for the Company"
	keepTogetherDateText      = "Date: 31 August 2026"
)

// TestKeepTogetherTwinDiffersOnlyByTheTags is the precondition every
// other assertion in this file rests on. If the pair ever drifted into
// differing for a second reason, "the two renders differ" would stop
// being evidence about grouping.
func TestKeepTogetherTwinDiffersOnlyByTheTags(t *testing.T) {
	const tag = `"keepTogether": "signature", `
	if n := strings.Count(keepTogetherTemplateJSON, tag); n != 3 {
		t.Fatalf("the grouped template must carry exactly three tags, found %d — the fixture is a three-element signature block", n)
	}
	if strings.Contains(keepTogetherUngroupedTemplateJSON, "keepTogether") {
		t.Fatal("the ungrouped twin must carry no keepTogether tag at all")
	}
	if stripped := strings.ReplaceAll(keepTogetherTemplateJSON, tag, ""); stripped != keepTogetherUngroupedTemplateJSON {
		t.Fatalf("the twin is not the grouped template minus its tags — the pair differs for some SECOND reason, so it no longer discriminates:\n%s", stripped)
	}
}

// TestKeepTogetherMovesTheWholeBlock is AC1 and AC5 together: grouped,
// every member is on page 2 and page 1 carries none of them; ungrouped,
// the SAME document splits. Removing the tags must make the grouped
// assertion fail — which is exactly what the second half measures.
func TestKeepTogetherMovesTheWholeBlock(t *testing.T) {
	grouped := keepTogetherPages(t, keepTogetherTemplateJSON)
	if len(grouped) != 2 {
		t.Fatalf("the grouped document must paginate to 2 pages, got %d", len(grouped))
	}
	if strings.Contains(grouped[0].text, keepTogetherSignatureText) ||
		strings.Contains(grouped[0].text, keepTogetherDateText) ||
		grouped[0].rects != 0 {
		t.Fatalf("page 1 must carry NO member of the signature block, got text containing the block and %d rect(s)", grouped[0].rects)
	}
	if !strings.Contains(grouped[1].text, keepTogetherSignatureText) {
		t.Error("page 2 must carry the signature line")
	}
	if !strings.Contains(grouped[1].text, keepTogetherDateText) {
		t.Error("page 2 must carry the date line")
	}
	if grouped[1].rects != 1 {
		t.Errorf("page 2 must carry the ruled line's rectangle, got %d rect(s)", grouped[1].rects)
	}
	// The body is UNTOUCHED: no sibling moved (AC1). Page 1 still ends
	// with the body's own last words.
	if !strings.Contains(grouped[0].text, "and the same instrument.") {
		t.Error("the body text must stay on page 1 — honouring a group moves the group, never a sibling")
	}

	ungrouped := keepTogetherPages(t, keepTogetherUngroupedTemplateJSON)
	if len(ungrouped) != 2 {
		t.Fatalf("the ungrouped twin must paginate to 2 pages, got %d", len(ungrouped))
	}
	if !strings.Contains(ungrouped[0].text, keepTogetherSignatureText) {
		t.Fatal("WITHOUT the tags the signature line must be STRANDED on page 1 — if it is not, this fixture no longer discriminates and every other assertion here is vacuous")
	}
	if strings.Contains(ungrouped[0].text, keepTogetherDateText) || ungrouped[0].rects != 0 {
		t.Fatal("without the tags the rule and the date must fall to page 2 — the split is what the grouped render is being compared against")
	}
	if !strings.Contains(ungrouped[1].text, keepTogetherDateText) || ungrouped[1].rects != 1 {
		t.Fatal("without the tags page 2 must carry the rule and the date")
	}
}

// TestKeepTogetherChangesTheRenderedBytes is the difference at the
// PUBLIC boundary, which is where AC5 actually lives: the same
// document with and without its tags produces different documents.
func TestKeepTogetherChangesTheRenderedBytes(t *testing.T) {
	grouped := renderKeepTogether(t)
	tpl, err := ParseTemplate([]byte(keepTogetherUngroupedTemplateJSON))
	if err != nil {
		t.Fatalf("parse ungrouped twin: %v", err)
	}
	res, err := Render(tpl, Data(keepTogetherDataJSON), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("render ungrouped twin: %v", err)
	}
	if len(grouped) == 0 || len(res.Bytes) == 0 {
		t.Fatal("both renders must produce bytes")
	}
	if sha256Hex(grouped) == sha256Hex(res.Bytes) {
		t.Fatal("the grouped and ungrouped renders hash identically — the tags changed nothing, so this fixture proves nothing")
	}
}

// keepTogetherNonContiguousTemplate is D-7.7.1's no-contiguity case,
// which is the property that lets the shipped machinery carry this
// feature at all: paginate.go records R7's contiguity premise as
// MEASURED FALSE and removed, so two tagged elements separated in column
// order by an intervening UNTAGGED one are still one group.
//
// It is built so the intervening element does the OPPOSITE of riding
// along, which is what makes it discriminating rather than merely
// present:
//
//	e2  tagged, y 690  -> 690.000 .. 704.982
//	e3  UNTAGGED rect, y 700, height 40 -> 700.000 .. 740.000   OUT
//	e4  tagged, y 710  -> 710.000 .. 724.982
//
// The group's union extent is 690.000 .. 724.982, which FITS window one
// (729.890), so the group's page is decided as page 1 at e2. e3 then
// fails the fit test on its own account and advances the window to
// page 2. e4 is visited LAST, with the window already on page 2 — and it
// lands on page 1 anyway, because its group's page was resolved at e2
// and every later-visited member rides along without asking the window a
// second question.
//
// So: an intervening item advances the window FOR ITSELF and can never
// split the group. Without the tags, e4 follows the window to page 2.
const keepTogetherNonContiguousTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 523, "height": 40, "value": "Body", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 690, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e3", "type": "rect", "x": 300, "y": 700, "width": 200, "height": 40, "style": {"background": "#000000"}},
        {"id": "e4", "type": "text", "x": 0, "y": 710, "width": 240, "height": 16, "keepTogether": "signature", "value": "Date: 31 August 2026", "style": {"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 5,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherMembersNeedNotBeContiguous asserts both halves of the
// property, and the second half against the same document with its tags
// removed, so neither half can pass by accident.
func TestKeepTogetherMembersNeedNotBeContiguous(t *testing.T) {
	pages := keepTogetherPages(t, keepTogetherNonContiguousTemplate)
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
	if !strings.Contains(pages[0].text, keepTogetherSignatureText) {
		t.Error("the group fits window one, so its first member belongs on page 1")
	}
	if !strings.Contains(pages[0].text, keepTogetherDateText) {
		t.Error("the LAST-visited member must ride along to its group's page even though the window has already advanced past it — that ride-along IS the no-contiguity property")
	}
	if pages[0].rects != 0 || pages[1].rects != 1 {
		t.Errorf("the intervening UNTAGGED rect must advance the window for itself and land on page 2; got %d rect(s) on page 1 and %d on page 2", pages[0].rects, pages[1].rects)
	}

	// The control, without which the assertions above would hold for a
	// build that simply never advanced the window here.
	ungrouped := strings.ReplaceAll(keepTogetherNonContiguousTemplate, `"keepTogether": "signature", `, "")
	loose := keepTogetherPages(t, ungrouped)
	if len(loose) != 2 {
		t.Fatalf("the untagged twin must also paginate to 2 pages, got %d", len(loose))
	}
	if strings.Contains(loose[0].text, keepTogetherDateText) {
		t.Fatal("WITHOUT the tags the last element must follow the window to page 2 — if it does not, this case no longer discriminates")
	}
	if !strings.Contains(loose[1].text, keepTogetherDateText) {
		t.Fatal("without the tags the last element belongs on page 2")
	}
}

// keepTogetherSingleMemberTemplate is the degenerate case: ONE element
// carries a tag, so its "group" has exactly one member and its union
// extent is its own extent. Placement must be identical to the same
// element untagged — a group of one constrains nothing.
//
// The element is deliberately SINGLE-LINE. A multi-line text element
// tagged alone would be a real change (every line of it becomes one
// atomic unit), and that is the feature working, not a violation of this
// row; the row is about an item whose group adds no member it did not
// already have.
const keepTogetherSingleMemberTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 523, "height": 40, "value": "Body", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 720, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherSingleMemberChangesNothing asserts the identity at the
// bytes, which is the strongest form available: the two documents render
// to the same document.
func TestKeepTogetherSingleMemberChangesNothing(t *testing.T) {
	tagged := keepTogetherRenderOf(t, keepTogetherSingleMemberTemplate)
	untagged := keepTogetherRenderOf(t, strings.ReplaceAll(keepTogetherSingleMemberTemplate, `"keepTogether": "signature", `, ""))
	if sha256Hex(tagged) != sha256Hex(untagged) {
		t.Fatal("a one-member group must place its element exactly where the same element untagged is placed — the two renders differ")
	}
	// And the placement is the interesting one: the element does not fit
	// window one, so both documents put it on page 2. Without this the
	// identity above would hold vacuously for an element that never went
	// near a boundary.
	pages := keepTogetherPages(t, keepTogetherSingleMemberTemplate)
	if len(pages) != 2 || strings.Contains(pages[0].text, keepTogetherSignatureText) || !strings.Contains(pages[1].text, keepTogetherSignatureText) {
		t.Fatalf("the one-member group's element must fall to page 2 exactly as an untagged element would; got %d page(s)", len(pages))
	}
}

// keepTogetherTwoGroupsTemplate is the independence case: TWO tags in one
// content band.
//
//	group "a": e2 (690.000 .. 704.982) and e3 (700.000 .. 714.982)
//	           union 690.000 .. 714.982 — FITS window one (729.890)
//	group "b": e4 (720.000 .. 734.982) and e5 (760.000 .. 774.982)
//	           union 720.000 .. 774.982 — does NOT fit
//
// The discriminator is that group "a" STAYS. If the two tags were folded
// into one group the union would be 690.000 .. 774.982 and all four
// elements would move together, which is exactly the failure a test that
// only checked "b" moved would miss.
const keepTogetherTwoGroupsTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 523, "height": 40, "value": "Body", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 690, "width": 240, "height": 16, "keepTogether": "first-signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e3", "type": "text", "x": 0, "y": 700, "width": 240, "height": 16, "keepTogether": "first-signature", "value": "Director", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e4", "type": "text", "x": 300, "y": 720, "width": 240, "height": 16, "keepTogether": "second-signature", "value": "Countersigned by the Auditor", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e5", "type": "text", "x": 300, "y": 760, "width": 240, "height": 16, "keepTogether": "second-signature", "value": "Date: 31 August 2026", "style": {"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherTwoGroupsAreIndependent is the I/O matrix's
// "two distinct tags" row: each group moves whole on its own, and
// neither constrains the other.
func TestKeepTogetherTwoGroupsAreIndependent(t *testing.T) {
	pages := keepTogetherPages(t, keepTogetherTwoGroupsTemplate)
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
	if !strings.Contains(pages[0].text, keepTogetherSignatureText) || !strings.Contains(pages[0].text, "Director") {
		t.Error("the first group fits window one and must stay whole on page 1 — a second group's extent must not drag it anywhere")
	}
	if strings.Contains(pages[0].text, "Countersigned by the Auditor") || strings.Contains(pages[0].text, keepTogetherDateText) {
		t.Error("the second group does not fit window one and must move whole to page 2")
	}
	if !strings.Contains(pages[1].text, "Countersigned by the Auditor") || !strings.Contains(pages[1].text, keepTogetherDateText) {
		t.Error("both members of the second group belong on page 2")
	}
}

// keepTogetherRenderOf renders an inline template through the public
// entry point and returns its bytes.
func keepTogetherRenderOf(t *testing.T, templateJSON string) []byte {
	t.Helper()
	tpl, err := ParseTemplate([]byte(templateJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data(keepTogetherDataJSON), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return res.Bytes
}

// keepTogetherOverTallTemplate is AC3's population: a declared group
// whose union extent exceeds a whole content window (729.890 pt), so no
// window of any position contains it.
//
// Its two members span 0 .. 900+, which is taller than any page this
// document has. Under D-4.6.2 as amended on 2026-08-31, that is CLIPPED
// with a Warning beside the bytes — never a fatal error, and never
// silent.
const keepTogetherOverTallTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 900, "width": 240, "height": 16, "keepTogether": "signature", "value": "Date: 31 August 2026", "style": {"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherOverTallGroupIsClippedWithAWarning is AC3.
//
// It asserts the disposition (bytes plus exactly one Warning, never an
// error), the code (TABLE_ROW_CLIPPED_HEIGHT, reused with a fourth role
// arm rather than minted — D-4.6.2 as amended), the location (the
// group's first member) and, most importantly, the PROSE: without the
// fourth arm a signature block is announced to its author as "row 0 of
// the bound collection" with a remedy about cell padding.
func TestKeepTogetherOverTallGroupIsClippedWithAWarning(t *testing.T) {
	tpl, err := ParseTemplate([]byte(keepTogetherOverTallTemplate))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, err := Render(tpl, Data(keepTogetherDataJSON), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("an over-tall keep-together group must never be fatal: %v", err)
	}
	if len(res.Bytes) == 0 {
		t.Fatal("an over-tall keep-together group must still return a document")
	}
	var clips []Diagnostic
	for _, d := range res.Diagnostics {
		if d.Code == DiagCodeTableRowClippedHeight {
			clips = append(clips, d)
		}
	}
	if len(clips) != 1 {
		t.Fatalf("want exactly one TABLE_ROW_CLIPPED_HEIGHT diagnostic, got %d: %+v", len(clips), res.Diagnostics)
	}
	c := clips[0]
	if c.Severity != SeverityWarning {
		t.Errorf("severity = %v, want Warning — AD-14 makes clipped content a Warning returned alongside the bytes", c.Severity)
	}
	if c.ElementID != "e1" {
		t.Errorf("ElementID = %q, want the group's FIRST member e1", c.ElementID)
	}
	if strings.Contains(c.Message, "of the bound collection") {
		t.Errorf("the message announces a keep-together group as a table row — the fourth role arm is missing:\n%s", c.Message)
	}
	if !strings.Contains(c.Message, `keep-together group "signature"`) {
		t.Errorf("the message must name the keep-together group the author declared:\n%s", c.Message)
	}
	if !strings.Contains(c.Message, "stop declaring the group") {
		t.Errorf("the message must name an action its author can actually take on a group:\n%s", c.Message)
	}
	if strings.Contains(c.Message, "cell padding") {
		t.Errorf("the remedy for an over-tall group is not a table row's remedy:\n%s", c.Message)
	}
}

// TestKeepTogetherGroupKeyIsNotAValidElementID verifies D-7.7 Ruling C
// rather than asserting it in prose.
//
// The keep-together key's ElementID lives in a namespace no element id
// can occupy, and THAT — not a comment — is what makes every FR26
// header-repeat, row-displacement, header-suppression and footer-orphan
// path unreachable for such a group. The day someone "tidies" the
// prefix into a plain identifier, every one of those paths reopens
// silently, and this is the test that says so.
func TestKeepTogetherGroupKeyIsNotAValidElementID(t *testing.T) {
	key := keepTogetherIndex{"e1": "signature"}.keepTogetherGroup("e1").Key
	if key.IsHeader {
		t.Error("a keep-together group must never be a header group — IsHeader gates the FR26 repeat paths")
	}
	if key.Index == footerGroupIndex {
		t.Errorf("a keep-together group's Index must not collide with the footer sentinel %d", footerGroupIndex)
	}
	// The grammar itself, through the public loader: a document whose
	// element id IS the group key must be REFUSED at load, which is what
	// makes the two namespaces provably disjoint.
	doc := strings.Replace(keepTogetherOverTallTemplate, `"id": "e1"`, `"id": "`+key.ElementID+`"`, 1)
	if _, err := template.ParseDocument([]byte(doc)); err == nil {
		t.Fatalf("element id %q was ACCEPTED — the keep-together namespace is not disjoint from the element-id grammar, and every table-shaped path in internal/layout is reachable by a colliding key", key.ElementID)
	}
	// And the control, so the refusal above is about the id and not
	// about the document: the same document with a legal id loads.
	if _, err := template.ParseDocument([]byte(keepTogetherOverTallTemplate)); err != nil {
		t.Fatalf("the control document must load: %v", err)
	}
}

// TestKeepTogetherReachesNoTablePath is the behavioural half of D-7.7.1's
// co-extensiveness audit.
//
// The prefix argument says a keep-together key cannot reach any of the
// four sites in internal/layout that read an ItemGroupKey as a TABLE
// (enumerated in keepTogetherKeyPrefix's own doc comment). This measures
// the consequence instead of restating the argument: a document made
// ENTIRELY of keep-together groups, including one over-tall enough to
// take the Story 4.6 clip branch — the branch that contains two of the
// four sites — must produce no header repeat, no row displacement and no
// header suppression at all.
func TestKeepTogetherReachesNoTablePath(t *testing.T) {
	geometry := layout.PageGeometry{
		Width: 595_276, Height: 841_890,
		MarginTop: 36_000, MarginBottom: 36_000,
		MarginLeft: 36_000, MarginRight: 36_000,
		PageHeaderHeight: 20_000, PageFooterHeight: 20_000,
	}
	idx := keepTogetherIndex{"e1": "signature", "e2": "signature"}
	g := idx.keepTogetherGroup("e1")
	items := []layout.ColumnItem{
		{ElementID: "e1", Top: 20_000, Bottom: 34_443, Runs: []layout.TextRunRef{0}, Group: g},
		// Far enough below to make the UNION taller than a whole content
		// window (729.890 pt), which is what routes this group through
		// the clip branch rather than the ordinary slide.
		{ElementID: "e2", Top: 940_000, Bottom: 954_443, Runs: []layout.TextRunRef{1}, Group: g},
	}
	plan, err := layout.Paginate(geometry, items)
	if err != nil {
		t.Fatalf("an over-tall keep-together group must be clipped, never fatal: %v", err)
	}
	if len(plan.Clipped) != 1 {
		t.Fatalf("want exactly one clip record, got %d", len(plan.Clipped))
	}
	if got := plan.Clipped[0].Key; got != g.Key {
		t.Errorf("the clip record must carry the group's own key, got %+v", got)
	}
	for i, p := range plan.Pages {
		if len(p.HeaderRepeats) != 0 {
			t.Errorf("page %d produced %d TableHeaderRepeat(s) for a group that is not a table — Gate B let the key through", i, len(p.HeaderRepeats))
		}
		if len(p.RowDisplacement) != 0 {
			t.Errorf("page %d produced %d RowDisplacement(s) for a group that is not a table", i, len(p.RowDisplacement))
		}
	}
	if len(plan.Suppressed) != 0 {
		t.Errorf("a keep-together group produced %d TableHeaderSuppressed record(s) — and TableHeaderSuppressed carries no Key, so a spurious one is indistinguishable at the caller", len(plan.Suppressed))
	}
	// The footer-orphan machinery must build no target for it either.
	if targets := footerOrphanTargetsFrom(items); len(targets) != 0 {
		t.Errorf("footerOrphanTargetsFrom built %d target(s) for a keep-together group: %+v", len(targets), targets)
	}
}

// TestKeepTogetherLeavesEveryTableRowUntouched is AC4's mechanism,
// asserted at the derivation rather than only through the goldens: the
// substitution fills ONLY the ungrouped case, so a table row's own key
// survives it and an untagged element gets the zero value it always had.
func TestKeepTogetherLeavesEveryTableRowUntouched(t *testing.T) {
	idx := keepTogetherIndex{"e9": "signature"}
	row := layout.ItemGroup{Present: true, Key: layout.ItemGroupKey{ElementID: "e9", Index: 3}}
	if got := idx.orKeepTogether(row, "e9"); got != row {
		t.Fatalf("a table row's own group must survive the substitution even when its element carries a tag, got %+v", got)
	}
	if got := idx.orKeepTogether(layout.ItemGroup{}, "e7"); got.Present {
		t.Fatalf("an UNTAGGED element must keep the zero ItemGroup, got %+v", got)
	}
	if got := idx.orKeepTogether(layout.ItemGroup{}, "e9"); !got.Present {
		t.Fatal("a tagged, otherwise-ungrouped element must acquire the keep-together group")
	}
}

// --- PHASE A must see the grouping too (Story 7.7) -----------------------
//
// contentColumnItems (page_number.go) is PHASE A: the page-count-only
// pass whose len(Pages) is the Y that {{pages}} prints and that
// D-2.7.2's {{page}} reservation is sized from. paginateDocument
// (render.go) is PHASE B, the pass that actually places content. Both
// carry the keep-together substitution, and BOTH comments say why: a
// grouping seen by only one of them makes the page COUNT disagree with
// the render.
//
// Nothing measured that until this test. Every other keep-together
// document here declares an EMPTY pageHeader and pageFooter, so nothing
// in them ever reads PHASE A's count at all, and all three PHASE A
// substitutions could be mutated away with the whole suite still green.
// This document is the one that reads it.

// keepTogetherPageCountTemplate is authored so the GROUPING CHANGES THE
// PAGE COUNT, not merely where the block sits — which is the only shape
// of document in which PHASE A's answer is observable at all.
//
//	e1  y 0     -> 0.000 .. 14.982
//	e2  y 100   -> 100.000 .. 114.982    tagged
//	e3  y 720   -> 720.000 .. 734.982    tagged
//	e4  y 1000  -> 1000.000 .. 1014.982
//
// UNGROUPED: window one is [0, 729.890], so e1 and e2 sit on page 1 and
// e3 opens page 2 with the window sliding to e3's own top, 720 — window
// two is [720, 1449.890], which also holds e4. TWO pages.
//
// GROUPED: the group's union extent is 100.000 .. 734.982 (634.982 pt,
// comfortably shorter than a whole window, so it is not over-tall) and
// it does not fit window one, so the window slides to the group's
// EARLIEST top, 100 — window two is [100, 829.890]. e2 and e3 ride there
// together, and e4 at 1000 no longer fits, so it opens a THIRD page.
// THREE pages.
//
// The page footer prints "Page {{page}} of {{pages}}". {{pages}} is not
// a late-bound slot: it is resolved once, from PHASE A's count, and
// printed literally (D-2.7.1/D-2.7.2). So if PHASE A does not see the
// grouping, every footer in a three-page document reads "of 2".
const keepTogetherPageCountTemplate = `{
  "assets": {},
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 240, "height": 16, "value": "Body", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 100, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e3", "type": "text", "x": 0, "y": 720, "width": 240, "height": 16, "keepTogether": "signature", "value": "Date: 31 August 2026", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e4", "type": "text", "x": 0, "y": 1000, "width": 240, "height": 16, "value": "Continued overleaf", "style": {"fontFamily": "body", "fontSize": 11}}
      ]
    },
    "pageFooter": {"elements": [{"id": "e5", "type": "text", "x": 0, "y": 0, "width": 240, "height": 16, "value": "Page {{page}} of {{pages}}", "style": {"fontFamily": "body", "fontSize": 11}}], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 6,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherPageCountReachesTheRenderedFooter is the PHASE A
// assertion, made at the only place PHASE A's answer is visible to a
// reader of the document: the printed total in the page footer.
//
// It fails — with the footers reading "of 2" over three pages — if any
// of contentColumnItems' three keep-together substitutions is removed,
// because the two passes then disagree about how many pages the
// document has and the count the footers print is PHASE A's.
func TestKeepTogetherPageCountReachesTheRenderedFooter(t *testing.T) {
	grouped := keepTogetherPages(t, keepTogetherPageCountTemplate)
	if len(grouped) != 3 {
		t.Fatalf("the grouped document must paginate to 3 pages, got %d — this fixture only measures PHASE A while the grouping CHANGES the count", len(grouped))
	}
	for i, p := range grouped {
		if !strings.Contains(p.text, "of 3") {
			t.Errorf("page %d's footer must state the GROUPED total, 3 — PHASE A (contentColumnItems) and PHASE B (paginateDocument) disagree about the page count, and {{pages}} prints PHASE A's answer. Page text: %q", i+1, p.text)
		}
	}
	if !strings.Contains(grouped[1].text, keepTogetherSignatureText) || !strings.Contains(grouped[1].text, keepTogetherDateText) {
		t.Error("both members of the group belong on page 2")
	}
	if !strings.Contains(grouped[2].text, "Continued overleaf") {
		t.Error("the element after the group opens page 3 — that displacement IS the page-count change")
	}

	// The control, so "of 3" is evidence about the grouping rather than
	// about this document: the SAME document without its tags is two
	// pages, and says so.
	ungrouped := keepTogetherPages(t, strings.ReplaceAll(keepTogetherPageCountTemplate, `"keepTogether": "signature", `, ""))
	if len(ungrouped) != 2 {
		t.Fatalf("without the tags the document must paginate to 2 pages, got %d — if it does not, this case no longer discriminates", len(ungrouped))
	}
	for i, p := range ungrouped {
		if !strings.Contains(p.text, "of 2") {
			t.Errorf("untagged page %d's footer must state 2: %q", i+1, p.text)
		}
	}
}

// TestBothPaginationPassesAgreeWithAKeepTogetherGroup is
// TestBothPaginationPassesAgreeOnRowPartition's sibling for this story's
// population (table_pagination_test.go).
//
// It asserts the stronger fact the footer test observes only the
// consequence of: with a group declared, the two builders — which append
// their items in DIFFERENT orders (PHASE A text-then-images, PHASE B
// rects-then-text-then-images) — produce the same page count AND the
// same per-element partition. PHASE A here is the production function
// itself, contentColumnItems, so removing its substitution reddens this
// too.
func TestBothPaginationPassesAgreeWithAKeepTogetherGroup(t *testing.T) {
	tpl, err := ParseTemplate([]byte(keepTogetherImageTemplate))
	if err != nil {
		t.Fatalf("ParseTemplate: %v", err)
	}
	bands, err := documentBands(tpl)
	if err != nil {
		t.Fatalf("documentBands: %v", err)
	}
	data := mustDecodeData(t, keepTogetherDataJSON)
	params := mustDecodeParams(t)
	geometry, err := pageGeometryOf(tpl)
	if err != nil {
		t.Fatalf("pageGeometryOf: %v", err)
	}
	visible, err := computeVisibility(bands, data, params, testFormatContext())
	if err != nil {
		t.Fatalf("computeVisibility: %v", err)
	}
	imageRuns, err := collectImageRuns(tpl)
	if err != nil {
		t.Fatalf("collectImageRuns: %v", err)
	}
	contentRuns, _, _, err := collectBandTextRuns(tpl, bands, contentBandIndex, data, params, testShippedFontSet(), newFontCache(), contentBandResolver, visible)
	if err != nil {
		t.Fatalf("collectBandTextRuns: %v", err)
	}
	keepTogether := keepTogetherTags(tpl)
	if len(keepTogether) != 2 {
		t.Fatalf("presence precondition: the fixture must declare a two-member group, got %d tagged element(s)", len(keepTogether))
	}

	// PHASE A: the real production function.
	phaseA, err := layout.Paginate(geometry, contentColumnItems(contentRuns, imageRuns, nil, visible, keepTogether))
	if err != nil {
		t.Fatalf("PHASE A Paginate: %v", err)
	}

	// PHASE B: paginateDocument's OWN item order, over the SAME slices,
	// so any divergence is about order and grouping alone.
	var phaseBItems []layout.ColumnItem
	for i := 0; i < len(contentRuns); i++ {
		j := i
		item := layout.ColumnItem{
			ElementID: contentRuns[i].elementID,
			Top:       contentRuns[i].itemTop,
			Bottom:    contentRuns[i].itemBottom,
			Group:     keepTogether.orKeepTogether(contentRuns[i].lineRowGroup(), contentRuns[i].elementID),
		}
		for j < len(contentRuns) &&
			contentRuns[j].elementID == contentRuns[i].elementID &&
			contentRuns[j].lineIndex == contentRuns[i].lineIndex {
			item.Runs = append(item.Runs, layout.TextRunRef(j))
			j++
		}
		phaseBItems = append(phaseBItems, item)
		i = j - 1
	}
	for i, r := range imageRuns {
		if r.band != contentBandIndex || !isVisible(visible, template.ElementID(r.elementID)) {
			continue
		}
		phaseBItems = append(phaseBItems, layout.ColumnItem{
			ElementID: r.elementID,
			Top:       r.y,
			Bottom:    r.y + r.boxH,
			Images:    []layout.ImageRef{layout.ImageRef(i)},
			Group:     keepTogether.keepTogetherGroup(r.elementID),
		})
	}
	phaseB, err := layout.Paginate(geometry, phaseBItems)
	if err != nil {
		t.Fatalf("PHASE B Paginate: %v", err)
	}

	if len(phaseA.Pages) != len(phaseB.Pages) {
		t.Fatalf("PHASE A produced %d pages, PHASE B produced %d — {{page}}/{{pages}} (D-2.7.2) would print the WRONG total", len(phaseA.Pages), len(phaseB.Pages))
	}
	if len(phaseA.Pages) < 2 {
		t.Fatalf("presence precondition: the fixture must paginate to >= 2 pages, got %d", len(phaseA.Pages))
	}
	pagesPerElement := func(plan layout.Pagination) map[string]map[int]bool {
		out := map[string]map[int]bool{}
		add := func(id string, p int) {
			if out[id] == nil {
				out[id] = map[int]bool{}
			}
			out[id][p] = true
		}
		for p, pg := range plan.Pages {
			for _, ref := range pg.ContentRuns {
				add(contentRuns[ref].elementID, p)
			}
			for _, ref := range pg.ContentImages {
				add(imageRuns[ref].elementID, p)
			}
		}
		return out
	}
	a, b := pagesPerElement(phaseA), pagesPerElement(phaseB)
	if len(a) == 0 {
		t.Fatal("coverage witness: PHASE A placed no element at all")
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("the two passes partitioned the SAME elements differently:\nPHASE A: %v\nPHASE B: %v", a, b)
	}
}

// --- a tagged IMAGE (Story 7.7) ------------------------------------------
//
// Both item builders tag images as well as text and rects, and D-4.6.2's
// amendment makes an image-specific promise: an image inside an
// over-tall group is REMOVED, not moved. Every other keep-together
// document here is text and rect only, so both image substitutions could
// be deleted with the suite still green. These two documents are what
// exercise them.

// keepTogetherImageAsset is the 3x2 8-bit RGB PNG maximalFixture ships,
// under its own SHA-256 key. Its CONTENT is irrelevant here — what
// matters is that the element is a real, resolvable image, so it reaches
// the page model as an ImagePlacement and its travel is observable.
const keepTogetherImageAsset = `"assets": {
    "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078": {
      "data": ["iVBORw0KGgoAAAANSUhEUgAAAAMAAAACCAIAAAASFvFNAAAAGElEQVR42mL6z8DAAMZMEOo/AwMg", "AAD//zwUBf/NjsW5AAAAAElFTkSuQmCC"],
      "mediaType": "image/png"
    }
  }`

// keepTogetherImageTemplate is a signature block whose second member is
// a signature IMAGE.
//
//	e2  text,  y 700 -> 700.000 .. 714.982   tagged — FITS window one
//	e3  image, y 718 -> 718.000 .. 758.000   tagged — does NOT fit
//
// Untagged, that severs the block exactly as the text-only fixture does,
// with the name at the foot of page 1 and the image at the head of
// page 2. Tagged, the union 700.000 .. 758.000 slides whole to page 2.
const keepTogetherImageTemplate = `{
  ` + keepTogetherImageAsset + `,
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 240, "height": 16, "value": "Body", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "text", "x": 0, "y": 700, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e3", "type": "image", "x": 0, "y": 718, "width": 60, "height": 40, "keepTogether": "signature", "asset": "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078"}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 4,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherCarriesATaggedImage is the image member's own travel,
// asserted as the difference between the tagged document and the same
// document untagged — the same discriminating shape every other case
// here uses.
func TestKeepTogetherCarriesATaggedImage(t *testing.T) {
	grouped := keepTogetherPages(t, keepTogetherImageTemplate)
	if len(grouped) != 2 {
		t.Fatalf("want 2 pages, got %d", len(grouped))
	}
	if strings.Contains(grouped[0].text, keepTogetherSignatureText) || grouped[0].images != 0 {
		t.Errorf("page 1 must carry NO member of the block: got %d image(s) and text %q", grouped[0].images, grouped[0].text)
	}
	if !strings.Contains(grouped[1].text, keepTogetherSignatureText) {
		t.Error("page 2 must carry the signature line")
	}
	if grouped[1].images != 1 {
		t.Errorf("page 2 must carry the signature IMAGE, got %d image(s) — an image is a group member like any other, and both item builders must tag it", grouped[1].images)
	}

	// The control: untagged, the text is stranded on page 1 and only the
	// image falls to page 2. Without this the assertions above would
	// hold for a build that never tagged the image at all.
	loose := keepTogetherPages(t, strings.ReplaceAll(keepTogetherImageTemplate, `"keepTogether": "signature", `, ""))
	if len(loose) != 2 {
		t.Fatalf("the untagged twin must also paginate to 2 pages, got %d", len(loose))
	}
	if !strings.Contains(loose[0].text, keepTogetherSignatureText) || loose[0].images != 0 {
		t.Fatalf("WITHOUT the tags the signature line must be STRANDED on page 1 and the image must not be there — this case no longer discriminates; got %d image(s) and text %q", loose[0].images, loose[0].text)
	}
	if loose[1].images != 1 {
		t.Fatalf("without the tags the image alone falls to page 2, got %d image(s)", loose[1].images)
	}
}

// keepTogetherOverTallImageTemplate is D-4.6.2's consciously accepted
// consequence, made measurable: a tagged image inside a group too tall
// for any window.
//
//	e1  text,  y 0   -> 0.000 .. 14.982     tagged
//	e2  image, y 900 -> 900.000 .. 940.000  tagged
//
// The union is 940.000 pt against a 729.890 pt window, so the group
// takes the Story 4.6 clip branch: a page of its own, clipped at that
// page's content bottom, one Warning beside the bytes. The clip drops
// whole members' Runs AND Images, so the image is REMOVED from the
// document rather than moved to a page of its own.
const keepTogetherOverTallImageTemplate = `{
  ` + keepTogetherImageAsset + `,
  "bands": {
    "content": {
      "elements": [
        {"id": "e1", "type": "text", "x": 0, "y": 0, "width": 240, "height": 16, "keepTogether": "signature", "value": "Signed for the Company", "style": {"fontFamily": "body", "fontSize": 11}},
        {"id": "e2", "type": "image", "x": 0, "y": 900, "width": 60, "height": 40, "keepTogether": "signature", "asset": "5a05ad01e89c143b7061b0c93450566568d38a23da9b9c5c9dfe449016433078"}
      ]
    },
    "pageFooter": {"elements": [], "height": 20},
    "pageHeader": {"elements": [], "height": 20}
  },
  "fonts": {"body": ["Noto Sans"]},
  "locale": "en",
  "nextId": 3,
  "page": {"margin": {"bottom": 36, "left": 36, "right": 36, "top": 36}, "orientation": "portrait", "size": "A4"},
  "utcOffset": "+00:00",
  "version": "1.2"
}
`

// TestKeepTogetherOverTallGroupDropsATaggedImage holds folio-format.md's
// image-specific promise — "an image inside an over-tall group is
// removed, not moved" — to a measurement rather than to prose.
func TestKeepTogetherOverTallGroupDropsATaggedImage(t *testing.T) {
	pages, diags := keepTogetherPagesWithDiagnostics(t, keepTogetherOverTallImageTemplate)
	if len(pages) != 1 {
		t.Fatalf("an over-tall group gets a page of its OWN and is clipped there, so this document is one page; got %d", len(pages))
	}
	if pages[0].images != 0 {
		t.Errorf("the tagged image is past the clipped page's content bottom, so it must be ABSENT from the document, not moved to a page of its own; got %d image(s)", pages[0].images)
	}
	if !strings.Contains(pages[0].text, keepTogetherSignatureText) {
		t.Error("the member that DOES fit the clipped page must still print — the clip drops what is past the bottom, not the whole group")
	}
	var clips int
	for _, d := range diags {
		if d.Code == DiagCodeTableRowClippedHeight {
			clips++
		}
	}
	if clips != 1 {
		t.Errorf("want exactly one TABLE_ROW_CLIPPED_HEIGHT Warning beside the bytes, got %d: %+v", clips, diags)
	}

	// The control: untagged, the very same image is MOVED to page 2 and
	// nothing is clipped or warned about. So the absence above is the
	// grouping's doing, not the image's own placement.
	loose, looseDiags := keepTogetherPagesWithDiagnostics(t, strings.ReplaceAll(keepTogetherOverTallImageTemplate, `"keepTogether": "signature", `, ""))
	if len(loose) != 2 || loose[1].images != 1 {
		t.Fatalf("WITHOUT the tags the image is simply moved to page 2 and nothing is dropped — this case no longer discriminates; got %d page(s)", len(loose))
	}
	if len(looseDiags) != 0 {
		t.Fatalf("the untagged twin must be clean; got %+v", looseDiags)
	}
}

// renderKeepTogether renders fixtures/keep-together/ through the public
// entry point.
func renderKeepTogether(t *testing.T) []byte {
	t.Helper()
	tpl, err := ParseTemplate([]byte(keepTogetherTemplateJSON))
	if err != nil {
		t.Fatalf("parse keep-together template: %v", err)
	}
	res, err := Render(tpl, Data(keepTogetherDataJSON), nil, testShippedFontSet())
	if err != nil {
		t.Fatalf("render keep-together: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Fatalf("the keep-together fixture must render with NO diagnostics; got %+v", res.Diagnostics)
	}
	return res.Bytes
}

// TestKeepTogetherDeclaresItsOwnVersion: the fixture carries the key, so
// it must declare 1.2 and must keep declaring it on re-save — never the
// library's ceiling (D-7.2.1, D-7.7.2).
func TestKeepTogetherDeclaresItsOwnVersion(t *testing.T) {
	if !strings.Contains(keepTogetherTemplateJSON, `"version": "1.2"`) {
		t.Fatal("the keep-together fixture declares the 1.2 key, so it must declare 1.2")
	}
	d, err := template.ParseDocument([]byte(keepTogetherTemplateJSON))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out, err := template.SerializeDocument(d)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	if !strings.Contains(string(out), `"version": "1.2"`) {
		t.Fatalf("re-serializing the fixture must keep 1.2, never raise it to the library's ceiling:\n%s", out)
	}
	if !strings.Contains(string(out), `"keepTogether": "signature"`) {
		t.Fatalf("the tag must round-trip:\n%s", out)
	}
}

// TestKeepTogetherGoldenFixture is the byte-identity half. It runs AFTER
// the semantic assertions above in file order and in intent (D-000.22).
func TestKeepTogetherGoldenFixture(t *testing.T) {
	root := repoRootFromTest(t)
	dir := filepath.Join(root, "fixtures", "keep-together")

	inputPath := filepath.Join(dir, "input.folio")
	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read %s: %v", inputPath, err)
	}
	if string(inputBytes) != keepTogetherTemplateJSON {
		t.Fatalf(
			"%s has drifted from folio-go/keepTogetherTemplateJSON (keep_together_template.go) — the two are "+
				"supposed to be byte-identical (kept in sync by hand, per alignment-rounding's precedent)",
			inputPath,
		)
	}

	fixture := loadExpectedFixture(t, filepath.Join(dir, "expected.json"))
	if fixture.FolioGoVersion == "" {
		t.Fatal("fixture is missing folioGoVersion")
	}
	if fixture.GoToolchain == "" {
		t.Fatal("fixture is missing goToolchain")
	}
	if !isSHA256HexString(fixture.SHA256) {
		t.Fatalf("fixture sha256 %q is not 64 lower-case hex characters", fixture.SHA256)
	}

	b := renderKeepTogether(t)
	got := sha256Hex(b)

	expectedPDF, err := os.ReadFile(filepath.Join(dir, "expected.pdf"))
	if err != nil {
		t.Fatalf("read expected.pdf: %v", err)
	}
	if onDisk := sha256Hex(expectedPDF); onDisk != fixture.SHA256 {
		t.Fatalf(
			"fixtures/keep-together/expected.pdf's own sha256 (%s) does not match expected.json's recorded sha256 (%s) — the fixture's two halves have drifted apart",
			onDisk, fixture.SHA256,
		)
	}
	if got != fixture.SHA256 {
		t.Fatalf(
			"golden fixture mismatch: got sha256 %s, want %s (fixtures/keep-together). Under AD-21/AD-22 this is a defect until proven to be an intended, versioned change. Do not regenerate the fixture to make this pass.",
			got, fixture.SHA256,
		)
	}
}
